// Agent Runtime Manager - Core runtime management for agent instances
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RuntimeConfig holds configuration for the agent runtime
type RuntimeConfig struct {
	MaxInstances       int           // Maximum number of agent instances
	InstanceTTL        time.Duration // Time-to-live for idle instances
	HealthCheckInterval time.Duration // Health check frequency
	EnablePool         bool          // Enable instance pooling
	EnableMetrics      bool          // Enable metrics collection
}

// DefaultRuntimeConfig returns default runtime configuration
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		MaxInstances:        100,
		InstanceTTL:         30 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		EnablePool:          true,
		EnableMetrics:       true,
	}
}

// InstanceStatus represents the status of an agent instance
type InstanceStatus string

const (
	StatusStarting  InstanceStatus = "starting"
	StatusRunning   InstanceStatus = "running"
	StatusIdle      InstanceStatus = "idle"
	StatusStopping  InstanceStatus = "stopping"
	StatusStopped   InstanceStatus = "stopped"
	StatusError     InstanceStatus = "error"
)

// AgentInstance represents a running agent instance
type AgentInstance struct {
	ID          string
	AgentName   string
	AgentType   string // "chat", "react", "workflow", etc.
	Status      InstanceStatus
	CreatedAt   time.Time
	LastActive  time.Time
	RequestCount int64
	// Add agent-specific fields based on your Agent interface
	Agent interface{} // agent.Agent or similar
}

// RuntimeManager manages the lifecycle of agent instances
type RuntimeManager struct {
	config *RuntimeConfig
	
	// Instance storage
	instances map[string]*AgentInstance
	byAgent   map[string][]string // agent name -> instance IDs
	mu        sync.RWMutex
	
	// Request routing
	router *Router
	
	// Pool management
	pool *InstancePool
	
	// Health monitoring
	healthChecker *HealthChecker
	
	// Metrics
	metrics *RuntimeMetrics
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Host reference (for creating agents)
	host interface{} // *agent.Host or similar
}

// NewRuntimeManager creates a new runtime manager
func NewRuntimeManager(cfg *RuntimeConfig, host interface{}) *RuntimeManager {
	if cfg == nil {
		cfg = DefaultRuntimeConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	rm := &RuntimeManager{
		config:    cfg,
		instances: make(map[string]*AgentInstance),
		byAgent:   make(map[string][]string),
		ctx:       ctx,
		cancel:    cancel,
		host:      host,
		router:    NewRouter(),
		healthChecker: NewHealthChecker(cfg.HealthCheckInterval),
		metrics:   NewRuntimeMetrics(),
	}
	
	if cfg.EnablePool {
		rm.pool = NewInstancePool(cfg.MaxInstances, cfg.InstanceTTL)
	}
	
	return rm
}

// Start starts the runtime manager
func (rm *RuntimeManager) Start() error {
	log.Printf("[Runtime] Starting runtime manager with max instances: %d", rm.config.MaxInstances)
	
	// Start health checker
	rm.healthChecker.Start(rm.ctx)
	
	// Start instance pool if enabled
	if rm.pool != nil {
		rm.pool.Start(rm.ctx)
	}
	
	// Start metrics collector if enabled
	if rm.config.EnableMetrics {
		rm.metrics.Start(rm.ctx)
	}
	
	log.Printf("[Runtime] Runtime manager started")
	return nil
}

// Stop stops the runtime manager gracefully
func (rm *RuntimeManager) Stop() error {
	log.Printf("[Runtime] Stopping runtime manager")
	
	rm.cancel()
	
	// Wait for all goroutines
	rm.wg.Wait()
	
	// Stop all instances
	rm.mu.Lock()
	for _, inst := range rm.instances {
		rm.stopInstanceLocked(inst)
	}
	rm.instances = make(map[string]*AgentInstance)
	rm.byAgent = make(map[string][]string)
	rm.mu.Unlock()
	
	log.Printf("[Runtime] Runtime manager stopped")
	return nil
}

// CreateInstance creates a new agent instance
func (rm *RuntimeManager) CreateInstance(ctx context.Context, agentName, agentType string) (*AgentInstance, error) {
	// Check max instances limit
	if len(rm.instances) >= rm.config.MaxInstances {
		return nil, fmt.Errorf("max instances (%d) reached", rm.config.MaxInstances)
	}
	
	// Create instance
	inst := &AgentInstance{
		ID:         uuid.New().String(),
		AgentName:  agentName,
		AgentType:  agentType,
		Status:     StatusStarting,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	
	// Initialize agent (placeholder - depends on your Agent interface)
	// This should create the actual agent instance based on host
	// inst.Agent = rm.host.CreateAgent(agentName, agentType)
	
	rm.mu.Lock()
	rm.instances[inst.ID] = inst
	rm.byAgent[agentName] = append(rm.byAgent[agentName], inst.ID)
	rm.mu.Unlock()
	
	// Start instance (async)
	go func() {
		if err := rm.startInstance(ctx, inst); err != nil {
			log.Printf("[Runtime] Failed to start instance %s: %v", inst.ID, err)
			rm.mu.Lock()
			inst.Status = StatusError
			rm.mu.Unlock()
		}
	}()
	
	return inst, nil
}

// startInstance starts an agent instance
func (rm *RuntimeManager) startInstance(ctx context.Context, inst *AgentInstance) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Initialize agent logic here
	// This should call the actual agent initialization
	
	inst.Status = StatusRunning
	inst.LastActive = time.Now()
	log.Printf("[Runtime] Instance %s (%s/%s) started", inst.ID, inst.AgentName, inst.AgentType)
	
	return nil
}

// GetInstance retrieves an instance by ID
func (rm *RuntimeManager) GetInstance(instanceID string) (*AgentInstance, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	inst, ok := rm.instances[instanceID]
	if ok {
		// Update last active time
		inst.LastActive = time.Now()
	}
	return inst, ok
}

// GetOrCreateInstance gets an existing instance or creates a new one
func (rm *RuntimeManager) GetOrCreateInstance(ctx context.Context, agentName, agentType string) (*AgentInstance, error) {
	// Try to get from pool first
	if rm.pool != nil {
		if inst := rm.pool.Get(agentName, agentType); inst != nil {
			rm.mu.Lock()
			inst.LastActive = time.Now()
			inst.Status = StatusRunning
			inst.RequestCount++
			rm.mu.Unlock()
			return inst, nil
		}
	}
	
	// Create new instance
	inst, err := rm.CreateInstance(ctx, agentName, agentType)
	if err != nil {
		return nil, err
	}
	
	// Wait for instance to be ready
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("instance startup timeout")
		case <-ticker.C:
			inst, ok := rm.GetInstance(inst.ID)
			if !ok {
				return nil, fmt.Errorf("instance not found")
			}
			if inst.Status == StatusRunning {
				return inst, nil
			}
			if inst.Status == StatusError {
				return nil, fmt.Errorf("instance failed to start")
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// StopInstance stops an instance
func (rm *RuntimeManager) StopInstance(instanceID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	inst, ok := rm.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	return rm.stopInstanceLocked(inst)
}

// stopInstanceLocked stops an instance (must hold lock)
func (rm *RuntimeManager) stopInstanceLocked(inst *AgentInstance) error {
	inst.Status = StatusStopping
	
	// Cleanup agent
	// if inst.Agent != nil {
	//     inst.Agent.Stop()
	// }
	
	inst.Status = StatusStopped
	
	// Remove from byAgent mapping
	instanceIDs := rm.byAgent[inst.AgentName]
	for i, id := range instanceIDs {
		if id == inst.ID {
			rm.byAgent[inst.AgentName] = append(instanceIDs[:i], instanceIDs[i+1:]...)
			break
		}
	}
	
	delete(rm.instances, inst.ID)
	
	log.Printf("[Runtime] Instance %s stopped", inst.ID)
	return nil
}

// ListInstances lists all instances
func (rm *RuntimeManager) ListInstances() []*AgentInstance {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	instances := make([]*AgentInstance, 0, len(rm.instances))
	for _, inst := range rm.instances {
		instances = append(instances, inst)
	}
	return instances
}

// ListInstancesByAgent lists instances for a specific agent
func (rm *RuntimeManager) ListInstancesByAgent(agentName string) []*AgentInstance {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	instances := make([]*AgentInstance, 0)
	for _, id := range rm.byAgent[agentName] {
		if inst, ok := rm.instances[id]; ok {
			instances = append(instances, inst)
		}
	}
	return instances
}

// GetStats returns runtime statistics
func (rm *RuntimeManager) GetStats() *RuntimeStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	totalRequests := int64(0)
	runningCount := 0
	idleCount := 0
	errorCount := 0
	
	for _, inst := range rm.instances {
		totalRequests += inst.RequestCount
		switch inst.Status {
		case StatusRunning:
			runningCount++
		case StatusIdle:
			idleCount++
		case StatusError:
			errorCount++
		}
	}
	
	poolStats := &PoolStats{}
	if rm.pool != nil {
		poolStats = rm.pool.GetStats()
	}
	
	return &RuntimeStats{
		TotalInstances:    len(rm.instances),
		RunningInstances:  runningCount,
		IdleInstances:     idleCount,
		ErrorInstances:    errorCount,
		TotalRequests:     totalRequests,
		PoolStats:         poolStats,
		Uptime:            time.Since(time.Now()),
	}
}

// RuntimeStats holds runtime statistics
type RuntimeStats struct {
	TotalInstances   int
	RunningInstances int
	IdleInstances    int
	ErrorInstances   int
	TotalRequests   int64
	PoolStats        *PoolStats
	Uptime           time.Duration
}

// PoolStats holds pool statistics
type PoolStats struct {
	Size        int
	IdleCount   int
	ActiveCount int
	HitRate     float64
}
