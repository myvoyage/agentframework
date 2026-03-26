// Workflow Integration - Integrates existing Workflow system with Runtime
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

// WorkflowAgent wraps a workflow to be used as a RunnableAgent
type WorkflowAgent struct {
	name       string
	workflow   interface{} // *agent.Workflow or similar
	execute    func(ctx context.Context, input string) (string, error)
}

// NewWorkflowAgent creates a new workflow agent wrapper
func NewWorkflowAgent(name string, executeFunc func(ctx context.Context, input string) (string, error)) *WorkflowAgent {
	return &WorkflowAgent{
		name:    name,
		execute: executeFunc,
	}
}

// Name returns the workflow name
func (wa *WorkflowAgent) Name() string {
	return wa.name
}

// Run executes the workflow
func (wa *WorkflowAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	if wa.execute == nil {
		return nil, fmt.Errorf("workflow execute function not set")
	}

	output, err := wa.execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: output,
	}, nil
}

// Stop stops the workflow (no-op for workflows)
func (wa *WorkflowAgent) Stop() error {
	return nil
}

// WorkflowRegistry manages workflow agents
type WorkflowRegistry struct {
	workflows map[string]*WorkflowAgent
	mu        sync.RWMutex
}

// NewWorkflowRegistry creates a new workflow registry
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]*WorkflowAgent),
	}
}

// Register registers a workflow agent
func (r *WorkflowRegistry) Register(name string, workflow *WorkflowAgent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workflows[name] = workflow
	log.Printf("[WorkflowRegistry] Registered workflow: %s", name)
}

// Get gets a workflow agent by name
func (r *WorkflowRegistry) Get(name string) (*WorkflowAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	workflow, ok := r.workflows[name]
	return workflow, ok
}

// List lists all registered workflows
func (r *WorkflowRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.workflows))
	for name := range r.workflows {
		names = append(names, name)
	}
	return names
}

// WorkflowRuntime integrates workflows with the agent runtime
type WorkflowRuntime struct {
	registry      *WorkflowRegistry
	agentRuntime  *AgentRuntime
	mu            sync.RWMutex
}

// NewWorkflowRuntime creates a new workflow runtime
func NewWorkflowRuntime(agentRuntime *AgentRuntime) *WorkflowRuntime {
	return &WorkflowRuntime{
		registry:     NewWorkflowRegistry(),
		agentRuntime: agentRuntime,
	}
}

// RegisterWorkflow registers a workflow as an agent
func (wr *WorkflowRuntime) RegisterWorkflow(name string, executeFunc func(ctx context.Context, input string) (string, error)) error {
	workflow := NewWorkflowAgent(name, executeFunc)
	wr.registry.Register(name, workflow)

	// Also register as agent factory
	if wr.agentRuntime != nil {
		wr.agentRuntime.RegisterAgent(name, func(ctx context.Context, config map[string]interface{}) (RunnableAgent, error) {
			return NewWorkflowAgent(name, executeFunc), nil
		})
	}

	return nil
}

// ExecuteWorkflow executes a workflow by name
func (wr *WorkflowRuntime) ExecuteWorkflow(ctx context.Context, name string, input string) (string, error) {
	workflow, ok := wr.registry.Get(name)
	if !ok {
		return "", fmt.Errorf("workflow not found: %s", name)
	}

	msg, err := workflow.Run(ctx, input)
	if err != nil {
		return "", err
	}

	return msg.Content, nil
}

// ListWorkflows lists all registered workflows
func (wr *WorkflowRuntime) ListWorkflows() []string {
	return wr.registry.List()
}

// RuntimeWorkflowExecutor provides a unified interface for executing agents or workflows
type RuntimeWorkflowExecutor struct {
	agentRuntime    *AgentRuntime
	workflowRuntime *WorkflowRuntime
}

// NewRuntimeWorkflowExecutor creates a new unified executor
func NewRuntimeWorkflowExecutor(agentRuntime *AgentRuntime, workflowRuntime *WorkflowRuntime) *RuntimeWorkflowExecutor {
	return &RuntimeWorkflowExecutor{
		agentRuntime:    agentRuntime,
		workflowRuntime: workflowRuntime,
	}
}

// Execute executes an agent or workflow by name
func (e *RuntimeWorkflowExecutor) Execute(ctx context.Context, name string, input string) (string, error) {
	// Try workflow first
	if e.workflowRuntime != nil {
		if _, ok := e.workflowRuntime.registry.Get(name); ok {
			return e.workflowRuntime.ExecuteWorkflow(ctx, name, input)
		}
	}

	// Try agent
	if e.agentRuntime != nil {
		// Get any idle agent with matching name
		agents := e.agentRuntime.ListAgents()
		for _, agent := range agents {
			if agent.agent.Name() == name {
				msg, err := agent.Run(ctx, input)
				if err != nil {
					return "", err
				}
				return msg.Content, nil
			}
		}
	}

	return "", fmt.Errorf("agent or workflow not found: %s", name)
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	ID        string
	Name      string
	Type      string // "agent" or "workflow"
	Input     string
	Output    string
	Error     string
	Duration  time.Duration
	Timestamp time.Time
}

// ExecutionHistory tracks execution history
type ExecutionHistory struct {
	results map[string]*ExecutionResult
	mu      sync.RWMutex
}

// NewExecutionHistory creates a new execution history
func NewExecutionHistory() *ExecutionHistory {
	return &ExecutionHistory{
		results: make(map[string]*ExecutionResult),
	}
}

// Add adds an execution result
func (h *ExecutionHistory) Add(result *ExecutionResult) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.results[result.ID] = result
}

// Get gets an execution result by ID
func (h *ExecutionHistory) Get(id string) (*ExecutionResult, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result, ok := h.results[id]
	return result, ok
}

// List lists all execution results
func (h *ExecutionHistory) List(limit int) []*ExecutionResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make([]*ExecutionResult, 0, len(h.results))
	for _, result := range h.results {
		results = append(results, result)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

// TrackedExecutor wraps an executor with history tracking
type TrackedExecutor struct {
	executor *RuntimeWorkflowExecutor
	history  *ExecutionHistory
}

// NewTrackedExecutor creates a new tracked executor
func NewTrackedExecutor(executor *RuntimeWorkflowExecutor) *TrackedExecutor {
	return &TrackedExecutor{
		executor: executor,
		history:  NewExecutionHistory(),
	}
}

// Execute executes and tracks the result
func (te *TrackedExecutor) Execute(ctx context.Context, name string, input string) (*ExecutionResult, error) {
	startTime := time.Now()

	result := &ExecutionResult{
		ID:        generateID(),
		Name:      name,
		Input:     input,
		Timestamp: startTime,
	}

	output, err := te.executor.Execute(ctx, name, input)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Output = output
	}

	te.history.Add(result)
	return result, err
}

// GetHistory returns execution history
func (te *TrackedExecutor) GetHistory(limit int) []*ExecutionResult {
	return te.history.List(limit)
}
