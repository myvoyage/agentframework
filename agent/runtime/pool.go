// Instance Pool - Pool management for agent instances
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"log"
	"sync"
	"time"
)

// InstancePool manages a pool of reusable agent instances
type InstancePool struct {
	maxSize  int
	ttl      time.Duration
	
	// Pool storage
	available map[string]*AgentInstance // instance ID -> instance
	mu        sync.RWMutex
	
	// Statistics
	hitCount  int64
	missCount int64
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewInstancePool creates a new instance pool
func NewInstancePool(maxSize int, ttl time.Duration) *InstancePool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &InstancePool{
		maxSize:   maxSize,
		ttl:       ttl,
		available: make(map[string]*AgentInstance),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the pool's background processes
func (p *InstancePool) Start(ctx context.Context) {
	// Start cleanup goroutine
	p.wg.Add(1)
	go p.cleanupLoop()
	
	log.Printf("[Pool] Instance pool started (max: %d, ttl: %v)", p.maxSize, p.ttl)
}

// Stop stops the pool
func (p *InstancePool) Stop() {
	p.cancel()
	p.wg.Wait()
	
	p.mu.Lock()
	p.available = make(map[string]*AgentInstance)
	p.mu.Unlock()
	
	log.Printf("[Pool] Instance pool stopped")
}

// Get gets an instance from the pool or returns nil if none available
func (p *InstancePool) Get(agentName, agentType string) *AgentInstance {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Find matching instance
	for id, inst := range p.available {
		if inst.AgentName == agentName && (agentType == "" || inst.AgentType == agentType) {
			// Remove from pool
			delete(p.available, id)
			p.hitCount++
			log.Printf("[Pool] Reusing instance %s from pool (hit rate: %.2f%%)", 
				id, p.hitRate())
			return inst
		}
	}
	
	p.missCount++
	return nil
}

// Put puts an instance back into the pool
func (p *InstancePool) Put(inst *AgentInstance) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Check if pool is full
	if len(p.available) >= p.maxSize {
		return false
	}
	
	// Check instance status
	if inst.Status != StatusRunning && inst.Status != StatusIdle {
		return false
	}
	
	// Add to pool with timestamp
	inst.Status = StatusIdle
	inst.LastActive = time.Now()
	p.available[inst.ID] = inst
	
	log.Printf("[Pool] Instance %s added to pool (size: %d/%d)", 
		inst.ID, len(p.available), p.maxSize)
	return true
}

// Remove removes an instance from the pool
func (p *InstancePool) Remove(instanceID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if _, ok := p.available[instanceID]; ok {
		delete(p.available, instanceID)
		return true
	}
	return false
}

// cleanupLoop periodically cleans up idle instances
func (p *InstancePool) cleanupLoop() {
	defer p.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

// cleanup removes expired instances from the pool
func (p *InstancePool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	expired := []string{}
	
	for id, inst := range p.available {
		if now.Sub(inst.LastActive) > p.ttl {
			expired = append(expired, id)
		}
	}
	
	for _, id := range expired {
		delete(p.available, id)
		log.Printf("[Pool] Removed expired instance %s from pool", id)
	}
	
	if len(expired) > 0 {
		log.Printf("[Pool] Cleaned up %d expired instances", len(expired))
	}
}

// GetStats returns pool statistics
func (p *InstancePool) GetStats() *PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return &PoolStats{
		Size:        len(p.available),
		IdleCount:   len(p.available),
		ActiveCount: 0, // Not tracked in pool
		HitRate:     p.hitRate(),
	}
}

// hitRate calculates the pool hit rate
func (p *InstancePool) hitRate() float64 {
	total := p.hitCount + p.missCount
	if total == 0 {
		return 0
	}
	return float64(p.hitCount) / float64(total) * 100
}

// Clear clears all instances from the pool
func (p *InstancePool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	count := len(p.available)
	p.available = make(map[string]*AgentInstance)
	
	log.Printf("[Pool] Cleared %d instances from pool", count)
}
