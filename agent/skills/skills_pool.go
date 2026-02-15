// Agent Framework - Skill System Connection Pool
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// PoolConfig defines connection pool configuration
type PoolConfig struct {
	MinConnections int           // Minimum number of connections
	MaxConnections int           // Maximum number of connections
	IdleTimeout    time.Duration // Idle timeout for connections
	MaxLifetime    time.Duration // Maximum lifetime of connections
	AcquireTimeout time.Duration // Timeout for acquiring connection
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MinConnections: 2,
		MaxConnections: 10,
		IdleTimeout:    5 * time.Minute,
		MaxLifetime:    30 * time.Minute,
		AcquireTimeout: 30 * time.Second,
	}
}

// PooledConnection represents a pooled connection
type PooledConnection struct {
	conn       interface{}
	createdAt  time.Time
	lastUsedAt time.Time
	usageCount int64
	pool       *ConnectionPool
}

// IsExpired checks if the connection is expired
func (pc *PooledConnection) IsExpired(maxLifetime, idleTimeout time.Duration) bool {
	now := time.Now()

	// Check maximum lifetime
	if maxLifetime > 0 && now.Sub(pc.createdAt) > maxLifetime {
		return true
	}

	// Check idle timeout
	if idleTimeout > 0 && now.Sub(pc.lastUsedAt) > idleTimeout {
		return true
	}

	return false
}

// ConnectionPool manages a pool of reusable connections
type ConnectionPool struct {
	config      *PoolConfig
	factory     func() (interface{}, error)
	closeFunc   func(interface{}) error
	idleConns   []*PooledConnection
	activeConns []*PooledConnection
	mu          sync.RWMutex
	closed      bool
	stats       *PoolStats
}

// PoolStats tracks pool statistics
type PoolStats struct {
	Created       int64
	Closed        int64
	Acquired      int64
	Released      int64
	WaitCount     int64
	WaitDuration  int64 // 改为int64类型，存储纳秒数
	MaxActive     int32
	CurrentIdle   int32
	CurrentActive int32
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *PoolConfig, factory func() (interface{}, error), closeFunc func(interface{}) error) (*ConnectionPool, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}

	if factory == nil {
		return nil, errors.New("factory function is required")
	}

	pool := &ConnectionPool{
		config:      config,
		factory:     factory,
		closeFunc:   closeFunc,
		idleConns:   make([]*PooledConnection, 0),
		activeConns: make([]*PooledConnection, 0),
		stats:       &PoolStats{},
	}

	// Initialize minimum connections
	if config.MinConnections > 0 {
		for i := 0; i < config.MinConnections; i++ {
			conn, err := factory()
			if err != nil {
				// Close any created connections
				pool.Close()
				return nil, err
			}

			pc := &PooledConnection{
				conn:       conn,
				createdAt:  time.Now(),
				lastUsedAt: time.Now(),
				pool:       pool,
			}
			pool.idleConns = append(pool.idleConns, pc)
			atomic.AddInt64(&pool.stats.Created, 1)
		}
	}

	return pool, nil
}

// Acquire acquires a connection from the pool
func (p *ConnectionPool) Acquire(ctx context.Context) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, errors.New("pool is closed")
	}

	startTime := time.Now()

	// Try to get an idle connection
	for len(p.idleConns) > 0 {
		pc := p.idleConns[len(p.idleConns)-1]
		p.idleConns = p.idleConns[:len(p.idleConns)-1]

		// Check if connection is expired
		if pc.IsExpired(p.config.MaxLifetime, p.config.IdleTimeout) {
			p.closeConnection(pc)
			continue
		}

		// Move to active
		p.activeConns = append(p.activeConns, pc)
		pc.lastUsedAt = time.Now()
		atomic.AddInt64(&pc.usageCount, 1)
		atomic.AddInt64(&p.stats.Acquired, 1)

		// Update max active
		currentActive := int32(len(p.activeConns))
		for {
			maxActive := atomic.LoadInt32(&p.stats.MaxActive)
			if currentActive <= maxActive {
				break
			}
			if atomic.CompareAndSwapInt32(&p.stats.MaxActive, maxActive, currentActive) {
				break
			}
		}

		return pc.conn, nil
	}

	// Check if we can create a new connection
	if len(p.activeConns) < p.config.MaxConnections {
		conn, err := p.factory()
		if err != nil {
			return nil, err
		}

		pc := &PooledConnection{
			conn:       conn,
			createdAt:  time.Now(),
			lastUsedAt: time.Now(),
			pool:       p,
		}

		p.activeConns = append(p.activeConns, pc)
		atomic.AddInt64(&p.stats.Created, 1)
		atomic.AddInt64(&p.stats.Acquired, 1)

		return pc.conn, nil
	}

	// Wait for a connection to become available
	return p.waitForConnection(ctx, startTime)
}

// waitForConnection waits for a connection to become available
func (p *ConnectionPool) waitForConnection(ctx context.Context, startTime time.Time) (interface{}, error) {
	atomic.AddInt64(&p.stats.WaitCount, 1)

	timeout := time.After(p.config.AcquireTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, errors.New("acquire connection timeout")
		case <-ticker.C:
			p.mu.Lock()
			if len(p.idleConns) > 0 {
				pc := p.idleConns[len(p.idleConns)-1]
				p.idleConns = p.idleConns[:len(p.idleConns)-1]

				if !pc.IsExpired(p.config.MaxLifetime, p.config.IdleTimeout) {
					p.activeConns = append(p.activeConns, pc)
					pc.lastUsedAt = time.Now()
					atomic.AddInt64(&pc.usageCount, 1)
					atomic.AddInt64(&p.stats.Acquired, 1)

					waitDuration := time.Since(startTime)
					atomic.AddInt64(&p.stats.WaitDuration, int64(waitDuration))

					p.mu.Unlock()
					return pc.conn, nil
				}

				p.closeConnection(pc)
			}
			p.mu.Unlock()
		}
	}
}

// Release releases a connection back to the pool
func (p *ConnectionPool) Release(conn interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("pool is closed")
	}

	// Find the connection in active list
	for i, pc := range p.activeConns {
		if pc.conn == conn {
			// Remove from active
			p.activeConns = append(p.activeConns[:i], p.activeConns[i+1:]...)

			// Check if connection is expired
			if pc.IsExpired(p.config.MaxLifetime, p.config.IdleTimeout) {
				p.closeConnection(pc)
				atomic.AddInt64(&p.stats.Released, 1)
				return nil
			}

			// Add back to idle
			pc.lastUsedAt = time.Now()
			p.idleConns = append(p.idleConns, pc)
			atomic.AddInt64(&p.stats.Released, 1)
			return nil
		}
	}

	return errors.New("connection not found in pool")
}

// closeConnection closes a connection
func (p *ConnectionPool) closeConnection(pc *PooledConnection) {
	if p.closeFunc != nil && pc.conn != nil {
		p.closeFunc(pc.conn)
	}
	atomic.AddInt64(&p.stats.Closed, 1)
}

// Close closes the connection pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Close all idle connections
	for _, pc := range p.idleConns {
		p.closeConnection(pc)
	}
	p.idleConns = nil

	// Close all active connections
	for _, pc := range p.activeConns {
		p.closeConnection(pc)
	}
	p.activeConns = nil

	return nil
}

// GetStats returns pool statistics
func (p *ConnectionPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"created":        atomic.LoadInt64(&p.stats.Created),
		"closed":         atomic.LoadInt64(&p.stats.Closed),
		"acquired":       atomic.LoadInt64(&p.stats.Acquired),
		"released":       atomic.LoadInt64(&p.stats.Released),
		"wait_count":     atomic.LoadInt64(&p.stats.WaitCount),
		"wait_duration":  atomic.LoadInt64(&p.stats.WaitDuration),
		"max_active":     atomic.LoadInt32(&p.stats.MaxActive),
		"current_idle":   int32(len(p.idleConns)),
		"current_active": int32(len(p.activeConns)),
		"config": map[string]interface{}{
			"min_connections": p.config.MinConnections,
			"max_connections": p.config.MaxConnections,
			"idle_timeout":    p.config.IdleTimeout.String(),
			"max_lifetime":    p.config.MaxLifetime.String(),
			"acquire_timeout": p.config.AcquireTimeout.String(),
		},
	}
}

// ExecutorPool manages a pool of skill executors
type ExecutorPool struct {
	pool *ConnectionPool
}

// NewExecutorPool creates a new executor pool
func NewExecutorPool(config *PoolConfig) (*ExecutorPool, error) {
	factory := func() (interface{}, error) {
		return NewEnhancedSkillExecutor(&DefaultExecutorConfig), nil
	}

	closeFunc := func(conn interface{}) error {
		// No need to close executor
		return nil
	}

	pool, err := NewConnectionPool(config, factory, closeFunc)
	if err != nil {
		return nil, err
	}

	return &ExecutorPool{pool: pool}, nil
}

// Acquire acquires an executor from the pool
func (ep *ExecutorPool) Acquire(ctx context.Context) (*EnhancedSkillExecutor, error) {
	conn, err := ep.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return conn.(*EnhancedSkillExecutor), nil
}

// Release releases an executor back to the pool
func (ep *ExecutorPool) Release(executor *EnhancedSkillExecutor) error {
	return ep.pool.Release(executor)
}

// Close closes the executor pool
func (ep *ExecutorPool) Close() error {
	return ep.pool.Close()
}

// GetStats returns pool statistics
func (ep *ExecutorPool) GetStats() map[string]interface{} {
	return ep.pool.GetStats()
}
