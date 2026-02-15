// Agent Framework - Dynamic Container Pool Implementation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"AgentFramework/agent/errors"
)

// DynamicPoolConfig contains configuration for dynamic container pool
type DynamicPoolConfig struct {
	InitialSize           int           `json:"initial_size"`             // Initial pool size per language
	MinSize               int           `json:"min_size"`                // Minimum pool size
	MaxSize               int           `json:"max_size"`                // Maximum pool size (0 = unlimited)
	ResizeThreshold      float64       `json:"resize_threshold"`       // Threshold (0.0-1.0) to trigger resize
	ResizeMultiplier    float64       `json:"resize_multiplier"`     // Multiplier for resizing
	IdleTimeout         time.Duration `json:"idle_timeout"`           // Idle timeout before destruction
	HealthCheckInterval  time.Duration `json:"health_check_interval"`  // Health check interval
	EnableMonitoring     bool          `json:"enable_monitoring"`      // Enable performance monitoring
}

// DefaultDynamicPoolConfig returns default configuration for dynamic container pool
func DefaultDynamicPoolConfig() DynamicPoolConfig {
	return DynamicPoolConfig{
		InitialSize:          3,
		MinSize:              2,
		MaxSize:              50,
		ResizeThreshold:     0.7,   // Resize when 70% full/used
		ResizeMultiplier:   2.0,  // Double the size
		IdleTimeout:        5 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		EnableMonitoring:    true,
	}
}

// DynamicLanguagePool represents a pool for a specific language with dynamic sizing
type DynamicLanguagePool struct {
	language   string
	containers chan *PooledContainer
	active     map[string]*PooledContainer
	mu         sync.RWMutex

	// Dynamic sizing
	currentSize atomic.Int32
	peakSize    atomic.Int32
	targetSize  atomic.Int32

	// Statistics
	stats      DynamicPoolStats
	muStats    sync.RWMutex
}

// DynamicPoolStats represents statistics for dynamic container pool
type DynamicPoolStats struct {
	TotalCreated    atomic.Int64
	TotalDestroyed  atomic.Int64
	TotalReused     atomic.Int64
	CurrentSize     int32
	ActiveCount      int32
	IdleCount        int32
	ResizeCount      atomic.Int64
	AvgWaitTime      atomic.Int64 // nanoseconds
	PeakConcurrency  atomic.Int32
}

// DynamicContainerPool is an enhanced container pool with dynamic sizing
type DynamicContainerPool struct {
	cfg      DynamicPoolConfig
	executor *ContainerExecutor
	pools    map[string]*DynamicLanguagePool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// Global monitoring
	globalStats struct {
		totalAcquires atomic.Int64
		totalReleases atomic.Int64
		totalWaits    atomic.Int64
	}
}

// NewDynamicContainerPool creates a new dynamic container pool
func NewDynamicContainerPool(executor *ContainerExecutor, cfg DynamicPoolConfig) *DynamicContainerPool {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.InitialSize <= 0 {
		cfg.InitialSize = 3
	}
	if cfg.MinSize <= 0 {
		cfg.MinSize = 2
	}
	if cfg.ResizeThreshold <= 0 || cfg.ResizeThreshold > 1.0 {
		cfg.ResizeThreshold = 0.7
	}
	if cfg.ResizeMultiplier < 1.0 {
		cfg.ResizeMultiplier = 2.0
	}

	pool := &DynamicContainerPool{
		cfg:      cfg,
		executor: executor,
		pools:    make(map[string]*DynamicLanguagePool),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start background goroutines
	go pool.monitorAndResize()

	return pool
}

// getOrCreatePool retrieves or creates a language pool
func (p *DynamicContainerPool) getOrCreatePool(language string) *DynamicLanguagePool {
	p.mu.RLock()
	langPool, exists := p.pools[language]
	p.mu.RUnlock()

	if exists {
		return langPool
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if langPool, exists := p.pools[language]; exists {
		return langPool
	}

	// Create new pool
	langPool = &DynamicLanguagePool{
		language:   language,
		containers: make(chan *PooledContainer, p.cfg.InitialSize),
		active:     make(map[string]*PooledContainer),
	}
	langPool.currentSize.Store(int32(p.cfg.InitialSize))
	langPool.targetSize.Store(int32(p.cfg.InitialSize))
	langPool.peakSize.Store(int32(p.cfg.InitialSize))

	p.pools[language] = langPool

	// Initialize pool with containers
	go langPool.initialize(p, p.cfg.InitialSize)

	return langPool
}

// initialize warms up the pool with initial containers
func (lp *DynamicLanguagePool) initialize(pool *DynamicContainerPool, count int) {
	for i := 0; i < count; i++ {
		ctx, cancel := context.WithTimeout(pool.ctx, 30*time.Second)
		container, err := pool.createContainer(ctx, lp.language)
		cancel()

		if err != nil {
			// Log error but continue
			continue
		}

		lp.muStats.Lock()
		lp.stats.TotalCreated.Add(1)
		lp.muStats.Unlock()

		// Add to pool
		select {
		case lp.containers <- container:
		case <-pool.ctx.Done():
			return
		}
	}
}

// Acquire retrieves a container from the pool
func (p *DynamicContainerPool) Acquire(ctx context.Context, language string) (*PooledContainer, error) {
	p.globalStats.totalAcquires.Add(1)

	langPool := p.getOrCreatePool(language)

	startTime := time.Now()
	waited := false

	// Try to get from pool
	select {
	case container := <-langPool.containers:
		waitTime := time.Since(startTime)
		if waitTime > 100*time.Millisecond {
			waited = true
			p.globalStats.totalWaits.Add(1)
		}

		langPool.muStats.Lock()
		langPool.stats.TotalReused.Add(1)
		langPool.stats.AvgWaitTime.Add(int64(waitTime))
		langPool.muStats.Unlock()

		// Check container health
		if !container.Healthy {
			// Destroy unhealthy container
			p.destroyContainer(container)
			// Try again recursively
			return p.Acquire(ctx, language)
		}

		// Update usage
		container.LastUsedAt = time.Now()
		container.UseCount++

		langPool.mu.Lock()
		langPool.active[container.ID] = container
		langPool.mu.Unlock()

		// Trigger resize check if waited
		if waited {
			go p.checkAndResize(langPool)
		}

		return container, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-time.After(100 * time.Millisecond):
		// Pool appears empty, trigger resize and retry
		go p.checkAndResize(langPool)

		// Fall through to create new container
	}

	// Create new container if pool is exhausted or empty
	return p.createContainer(ctx, language)
}

// Release returns a container to the pool
func (p *DynamicContainerPool) Release(container *PooledContainer) error {
	p.globalStats.totalReleases.Add(1)

	langPool, ok := p.pools[container.Language]
	if !ok {
		return fmt.Errorf("unknown language: %s", container.Language)
	}

	langPool.mu.Lock()
	delete(langPool.active, container.ID)
	langPool.mu.Unlock()

	// Check container health
	if !container.Healthy {
		return p.destroyContainer(container)
	}

	// Check idle timeout
	if time.Since(container.LastUsedAt) > p.cfg.IdleTimeout {
		return p.destroyContainer(container)
	}

	// Try to return to pool
	select {
	case langPool.containers <- container:
		return nil
	default:
		// Pool is full, destroy container
		return p.destroyContainer(container)
	}
}

// createContainer creates a new container
func (p *DynamicContainerPool) createContainer(ctx context.Context, language string) (*PooledContainer, error) {
	image := p.executor.getImage(language)
	containerID, err := p.executor.createContainer(ctx, image, language, "")
	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to create container for language %s", language)
	}

	container := &PooledContainer{
		ID:         containerID,
		Language:   language,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UseCount:   0,
		Healthy:    true,
	}

	langPool := p.getOrCreatePool(language)
	langPool.muStats.Lock()
	langPool.stats.TotalCreated.Add(1)
	langPool.mu.Lock()
	langPool.active[containerID] = container
	langPool.mu.Unlock()
	langPool.muStats.Unlock()

	return container, nil
}

// destroyContainer destroys a container
func (p *DynamicContainerPool) destroyContainer(container *PooledContainer) error {
	err := p.executor.removeContainer(context.Background(), container.ID)

	langPool, ok := p.pools[container.Language]
	if ok {
		langPool.muStats.Lock()
		langPool.stats.TotalDestroyed.Add(1)
		langPool.mu.Lock()
		delete(langPool.active, container.ID)
		langPool.mu.Unlock()
		langPool.muStats.Unlock()
	}

	return err
}

// checkAndResize checks if pool needs resizing and performs resize if needed
func (p *DynamicContainerPool) checkAndResize(langPool *DynamicLanguagePool) bool {
	langPool.mu.RLock()
	currentActive := len(langPool.active)
	currentIdle := len(langPool.containers)
	currentTotal := currentActive + currentIdle
	currentSize := int(langPool.currentSize.Load())
	langPool.mu.RUnlock()

	// Calculate usage ratio
	usageRatio := float64(currentActive) / float64(currentTotal)

	// Update peak size
	if currentTotal > int(langPool.peakSize.Load()) {
		langPool.peakSize.Store(int32(currentTotal))
	}

	// Check if expansion is needed (high usage)
	if usageRatio > p.cfg.ResizeThreshold {
		// Calculate new size
		newSize := int(float64(currentSize) * p.cfg.ResizeMultiplier)

		// Check max size constraint
		if p.cfg.MaxSize > 0 && newSize > p.cfg.MaxSize {
			newSize = p.cfg.MaxSize
		}

		// Only expand if growth is possible
		if newSize > currentSize {
			return p.resizePool(langPool, newSize)
		}
	}

	// Check if shrinkage is needed (low usage)
	if usageRatio < (1.0-p.cfg.ResizeThreshold) && currentTotal > p.cfg.MinSize {
		// Calculate new size (shrink)
		newSize := int(float64(currentSize) / p.cfg.ResizeMultiplier)

		// Check min size constraint
		if newSize < p.cfg.MinSize {
			newSize = p.cfg.MinSize
		}

		// Only shrink if reduction is possible
		if newSize < currentSize {
			return p.resizePool(langPool, newSize)
		}
	}

	return false
}

// resizePool resizes the pool to a new size
func (p *DynamicContainerPool) resizePool(langPool *DynamicLanguagePool, newSize int) bool {
	langPool.mu.Lock()
	defer langPool.mu.Unlock()

	oldSize := cap(langPool.containers)
	if newSize == oldSize {
		return false
	}

	// Create new channel
	newContainers := make(chan *PooledContainer, newSize)

	// Transfer existing containers
	close(langPool.containers)
	for container := range langPool.containers {
		select {
		case newContainers <- container:
		default:
			// New pool is smaller, destroy excess containers
			p.destroyContainer(container)
		}
	}

	langPool.containers = newContainers
	langPool.currentSize.Store(int32(newSize))
	langPool.muStats.Lock()
	langPool.stats.ResizeCount.Add(1)
	langPool.stats.CurrentSize = int32(newSize)
	langPool.muStats.Unlock()

	return true
}

// monitorAndResize periodically monitors and adjusts pool sizes
func (p *DynamicContainerPool) monitorAndResize() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.RLock()
			pools := make([]*DynamicLanguagePool, 0, len(p.pools))
			for _, langPool := range p.pools {
				pools = append(pools, langPool)
			}
			p.mu.RUnlock()

			for _, langPool := range pools {
				p.checkAndResize(langPool)
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// GetStats returns statistics for all language pools
func (p *DynamicContainerPool) GetStats() map[string]DynamicPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]DynamicPoolStats)

	for lang, langPool := range p.pools {
		langPool.muStats.Lock()
		poolStats := langPool.stats

		langPool.mu.RLock()
		poolStats.CurrentSize = langPool.currentSize.Load()
		poolStats.ActiveCount = int32(len(langPool.active))
		poolStats.IdleCount = int32(len(langPool.containers))
		langPool.mu.RUnlock()

		stats[lang] = poolStats
		langPool.muStats.Unlock()
	}

	return stats
}

// Close closes the dynamic container pool
func (p *DynamicContainerPool) Close() error {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Destroy all containers in all pools
	for _, langPool := range p.pools {
		// Close containers channel
		close(langPool.containers)

		// Destroy idle containers
		for container := range langPool.containers {
			p.destroyContainer(container)
		}

		// Destroy active containers
		langPool.mu.Lock()
		for _, container := range langPool.active {
			p.destroyContainer(container)
		}
		langPool.mu.Unlock()
	}

	return nil
}

// Helper function to create dynamic container pool with default config
func NewDefaultDynamicContainerPool(executor *ContainerExecutor) *DynamicContainerPool {
	return NewDynamicContainerPool(executor, DefaultDynamicPoolConfig())
}
