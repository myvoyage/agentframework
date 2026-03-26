// Health Checker - Health monitoring for agent instances
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"log"
	"sync"
	"time"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	InstanceID  string
	AgentName   string
	Status      HealthStatus
	Message     string
	Timestamp   time.Time
	Latency     time.Duration
}

// HealthChecker performs periodic health checks on instances
type HealthChecker struct {
	interval time.Duration
	
	// Health results
	healthResults map[string]*HealthCheck // instance ID -> health check
	mu            sync.RWMutex
	
	// Callbacks
	onHealthChange func(*HealthCheck)
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &HealthChecker{
		interval:      interval,
		healthResults: make(map[string]*HealthCheck),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the health checker
func (h *HealthChecker) Start(ctx context.Context) {
	h.wg.Add(1)
	go h.run()
	
	log.Printf("[Health] Health checker started (interval: %v)", h.interval)
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	h.cancel()
	h.wg.Wait()
	
	log.Printf("[Health] Health checker stopped")
}

// SetOnChange sets a callback for health status changes
func (h *HealthChecker) SetOnChange(fn func(*HealthCheck)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onHealthChange = fn
}

// run runs the health check loop
func (h *HealthChecker) run() {
	defer h.wg.Done()
	
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			// Health checks would be triggered by runtime manager
		}
	}
}

// UpdateHealth updates health status for an instance
func (h *HealthChecker) UpdateHealth(check *HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	previous := h.healthResults[check.InstanceID]
	h.healthResults[check.InstanceID] = check
	
	// Trigger callback if status changed
	if h.onHealthChange != nil {
		if previous == nil || previous.Status != check.Status {
			h.onHealthChange(check)
		}
	}
}

// GetHealth returns health status for an instance
func (h *HealthChecker) GetHealth(instanceID string) (*HealthCheck, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	health, ok := h.healthResults[instanceID]
	if ok {
		// Return a copy
		copy := *health
		return &copy, true
	}
	return nil, false
}

// GetAllHealth returns all health statuses
func (h *HealthChecker) GetAllHealth() []*HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	results := make([]*HealthCheck, 0, len(h.healthResults))
	for _, check := range h.healthResults {
		copy := *check
		results = append(results, &copy)
	}
	return results
}

// GetSummary returns health summary
func (h *HealthChecker) GetSummary() *HealthSummary {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	healthy := 0
	unhealthy := 0
	unknown := 0
	
	for _, check := range h.healthResults {
		switch check.Status {
		case HealthStatusHealthy:
			healthy++
		case HealthStatusUnhealthy:
			unhealthy++
		case HealthStatusUnknown:
			unknown++
		}
	}
	
	return &HealthSummary{
		Total:     len(h.healthResults),
		Healthy:   healthy,
		Unhealthy: unhealthy,
		Unknown:   unknown,
	}
}

// HealthSummary holds health summary statistics
type HealthSummary struct {
	Total     int
	Healthy   int
	Unhealthy int
	Unknown   int
}
