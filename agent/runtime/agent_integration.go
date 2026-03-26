// Agent Integration - Integrates existing Agent system with Runtime
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// RunnableAgent defines the interface for agents that can be run in the runtime
type RunnableAgent interface {
	// Name returns the agent name
	Name() string

	// Run executes the agent with the given input
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)

	// Stop stops the agent (optional, for cleanup)
	Stop() error
}

// AgentFactory creates agent instances
type AgentFactory func(ctx context.Context, config map[string]interface{}) (RunnableAgent, error)

// AgentRegistry manages agent factories
type AgentRegistry struct {
	factories map[string]AgentFactory
	mu        sync.RWMutex
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		factories: make(map[string]AgentFactory),
	}
}

// Register registers an agent factory
func (r *AgentRegistry) Register(name string, factory AgentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
	log.Printf("[AgentRegistry] Registered agent factory: %s", name)
}

// Get gets an agent factory by name
func (r *AgentRegistry) Get(name string) (AgentFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.factories[name]
	return factory, ok
}

// List lists all registered agent factories
func (r *AgentRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// RuntimeAgent wraps a RunnableAgent for use in the runtime
type RuntimeAgent struct {
	instanceID string
	agent      RunnableAgent
	status     InstanceStatus
	lastActive time.Time
	mu         sync.RWMutex
}

// NewRuntimeAgent creates a new runtime agent wrapper
func NewRuntimeAgent(instanceID string, agent RunnableAgent) *RuntimeAgent {
	return &RuntimeAgent{
		instanceID: instanceID,
		agent:      agent,
		status:     StatusRunning,
		lastActive: time.Now(),
	}
}

// Run runs the agent
func (ra *RuntimeAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	ra.mu.Lock()
	ra.status = StatusRunning
	ra.lastActive = time.Now()
	ra.mu.Unlock()

	defer func() {
		ra.mu.Lock()
		ra.status = StatusIdle
		ra.lastActive = time.Now()
		ra.mu.Unlock()
	}()

	return ra.agent.Run(ctx, input, opts...)
}

// Stop stops the agent
func (ra *RuntimeAgent) Stop() error {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	ra.status = StatusStopping
	if stopper, ok := ra.agent.(interface{ Stop() error }); ok {
		if err := stopper.Stop(); err != nil {
			ra.status = StatusError
			return err
		}
	}
	ra.status = StatusStopped
	return nil
}

// Status returns the current status
func (ra *RuntimeAgent) Status() InstanceStatus {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return ra.status
}

// AgentRuntimeConfig holds configuration for the agent runtime
type AgentRuntimeConfig struct {
	DefaultAgentType string
	MaxAgents        int
	AgentTTL         time.Duration
}

// DefaultAgentRuntimeConfig returns default configuration
func DefaultAgentRuntimeConfig() *AgentRuntimeConfig {
	return &AgentRuntimeConfig{
		DefaultAgentType: "chat",
		MaxAgents:        100,
		AgentTTL:         30 * time.Minute,
	}
}

// AgentRuntime integrates the agent system with the runtime manager
type AgentRuntime struct {
	config    *AgentRuntimeConfig
	registry  *AgentRegistry
	manager   *RuntimeManager
	agents    map[string]*RuntimeAgent
	mu        sync.RWMutex
}

// NewAgentRuntime creates a new agent runtime
func NewAgentRuntime(config *AgentRuntimeConfig, manager *RuntimeManager) *AgentRuntime {
	if config == nil {
		config = DefaultAgentRuntimeConfig()
	}

	return &AgentRuntime{
		config:   config,
		registry: NewAgentRegistry(),
		manager:  manager,
		agents:   make(map[string]*RuntimeAgent),
	}
}

// RegisterAgent registers an agent factory
func (ar *AgentRuntime) RegisterAgent(name string, factory AgentFactory) {
	ar.registry.Register(name, factory)
}

// CreateAgent creates a new agent instance
func (ar *AgentRuntime) CreateAgent(ctx context.Context, name string, config map[string]interface{}) (*RuntimeAgent, error) {
	// Check max agents limit
	ar.mu.RLock()
	if len(ar.agents) >= ar.config.MaxAgents {
		ar.mu.RUnlock()
		return nil, fmt.Errorf("max agents (%d) reached", ar.config.MaxAgents)
	}
	ar.mu.RUnlock()

	// Get factory
	factory, ok := ar.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("agent factory not found: %s", name)
	}

	// Create agent
	agent, err := factory(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Create instance in runtime manager
	inst, err := ar.manager.CreateInstance(ctx, name, "agent")
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Wrap agent
	runtimeAgent := NewRuntimeAgent(inst.ID, agent)

	// Store agent
	ar.mu.Lock()
	ar.agents[inst.ID] = runtimeAgent
	ar.mu.Unlock()

	log.Printf("[AgentRuntime] Created agent instance: %s (%s)", inst.ID, name)

	return runtimeAgent, nil
}

// GetAgent gets an agent by instance ID
func (ar *AgentRuntime) GetAgent(instanceID string) (*RuntimeAgent, bool) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	agent, ok := ar.agents[instanceID]
	return agent, ok
}

// GetOrCreateAgent gets an existing agent or creates a new one
func (ar *AgentRuntime) GetOrCreateAgent(ctx context.Context, name string, config map[string]interface{}) (*RuntimeAgent, error) {
	// Try to get from pool
	ar.mu.RLock()
	for _, agent := range ar.agents {
		if agent.agent.Name() == name && agent.Status() == StatusIdle {
			ar.mu.RUnlock()
			return agent, nil
		}
	}
	ar.mu.RUnlock()

	// Create new agent
	return ar.CreateAgent(ctx, name, config)
}

// RunAgent runs an agent with the given input
func (ar *AgentRuntime) RunAgent(ctx context.Context, instanceID string, input string, opts ...model.Option) (*schema.Message, error) {
	agent, ok := ar.GetAgent(instanceID)
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", instanceID)
	}

	startTime := time.Now()
	msg, err := agent.Run(ctx, input, opts...)
	duration := time.Since(startTime)

	// Record metrics
	if ar.manager.metrics != nil {
		ar.manager.metrics.RecordRequest(err == nil, duration)
	}

	if err != nil {
		return nil, err
	}

	return msg, nil
}

// StopAgent stops an agent
func (ar *AgentRuntime) StopAgent(instanceID string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	agent, ok := ar.agents[instanceID]
	if !ok {
		return fmt.Errorf("agent not found: %s", instanceID)
	}

	if err := agent.Stop(); err != nil {
		return err
	}

	delete(ar.agents, instanceID)

	// Stop instance in runtime manager
	if err := ar.manager.StopInstance(instanceID); err != nil {
		log.Printf("[AgentRuntime] Warning: failed to stop instance: %v", err)
	}

	log.Printf("[AgentRuntime] Stopped agent instance: %s", instanceID)
	return nil
}

// ListAgents lists all agent instances
func (ar *AgentRuntime) ListAgents() []*RuntimeAgent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	agents := make([]*RuntimeAgent, 0, len(ar.agents))
	for _, agent := range ar.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetStats returns runtime statistics
func (ar *AgentRuntime) GetStats() *AgentRuntimeStats {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	running := 0
	idle := 0
	stopped := 0

	for _, agent := range ar.agents {
		switch agent.Status() {
		case StatusRunning:
			running++
		case StatusIdle:
			idle++
		case StatusStopped:
			stopped++
		}
	}

	return &AgentRuntimeStats{
		TotalAgents:     len(ar.agents),
		RunningAgents:   running,
		IdleAgents:      idle,
		StoppedAgents:   stopped,
		RegisteredTypes: len(ar.registry.List()),
	}
}

// AgentRuntimeStats holds agent runtime statistics
type AgentRuntimeStats struct {
	TotalAgents     int
	RunningAgents   int
	IdleAgents      int
	StoppedAgents   int
	RegisteredTypes int
}

// HealthCheck performs a health check on all agents
func (ar *AgentRuntime) HealthCheck() []*AgentHealth {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	healths := make([]*AgentHealth, 0, len(ar.agents))
	for id, agent := range ar.agents {
		health := &AgentHealth{
			InstanceID: id,
			AgentName:  agent.agent.Name(),
			Status:     string(agent.Status()),
			LastActive: agent.lastActive,
		}

		// Check if agent is healthy
		if agent.Status() == StatusError {
			health.Healthy = false
			health.Message = "agent in error state"
		} else if agent.Status() == StatusStopped {
			health.Healthy = false
			health.Message = "agent is stopped"
		} else {
			health.Healthy = true
			health.Message = "agent is healthy"
		}

		healths = append(healths, health)
	}

	return healths
}

// AgentHealth represents the health status of an agent
type AgentHealth struct {
	InstanceID string    `json:"instance_id"`
	AgentName  string    `json:"agent_name"`
	Status     string    `json:"status"`
	Healthy    bool      `json:"healthy"`
	Message    string    `json:"message"`
	LastActive time.Time `json:"last_active"`
}
