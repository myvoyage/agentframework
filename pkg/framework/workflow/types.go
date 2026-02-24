// Agent Framework - Workflow Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v3"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // sequential, parallel, dag, routing
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	Variables   map[string]interface{} `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetID returns the workflow ID
func (w *Workflow) GetID() string {
	return w.ID
}

// GetName returns the workflow name
func (w *Workflow) GetName() string {
	return w.Name
}

// GetType returns the workflow type
func (w *Workflow) GetType() string {
	return w.Type
}

// WorkflowInterface defines the workflow execution interface
type WorkflowInterface interface {
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
	RunResumable(ctx context.Context, input string, state interface{}, opts ...model.Option) (*schema.Message, interface{}, error)
	GetID() string
	GetName() string
	GetType() string
}

// NodeExecutionResult represents the result of a node execution
type NodeExecutionResult struct {
	NodeID     string                 `json:"node_id"`
	Success    bool                   `json:"success"`
	Status     string                 `json:"status,omitempty"`     // NodeStatusRunning, NodeStatusCompleted, NodeStatusFailed
	Input      string                 `json:"input,omitempty"`
	Output     string                 `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   int64                  `json:"duration_ms"`
	StartTime  time.Time              `json:"start_time,omitempty"`
	EndTime    time.Time              `json:"end_time,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  int64                  `json:"timestamp"`
}

// Node status constants
const (
	NodeStatusPending   = "pending"
	NodeStatusRunning   = "running"
	NodeStatusCompleted = "completed"
	NodeStatusFailed    = "failed"
	NodeStatusCancelled = "cancelled"
)

// simpleAgentWorkflowAdapter is a simple agent workflow adapter
// Note: This is now defined in workflow_dag.go as a type

// WorkflowNode represents a node in a workflow
type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Position map[string]int         `json:"position"` // x, y coordinates
}

// WorkflowEdge represents an edge between nodes in a workflow
type WorkflowEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // sequential, conditional, parallel
}

// WorkflowExecutionStore stores workflow execution results
type WorkflowExecutionStore interface {
	SaveExecution(ctx context.Context, execution *WorkflowExecution) error
	GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)
	ListExecutions(ctx context.Context, workflowID string) ([]*WorkflowExecution, error)
	DeleteExecution(ctx context.Context, executionID string) error
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"` // pending, running, completed, failed, cancelled
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  int64                  `json:"started_at"`
	CompletedAt int64                 `json:"completed_at,omitempty"`
	NodeStates map[string]*NodeState  `json:"node_states"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NodeState represents the state of a node in workflow execution
type NodeState struct {
	NodeID     string                 `json:"node_id"`
	Status     string                 `json:"status"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  int64                  `json:"started_at"`
	CompletedAt int64                 `json:"completed_at,omitempty"`
}

// WorkflowExecutionResult represents the result of a workflow execution
type WorkflowExecutionResult struct {
	WorkflowID  string                 `json:"workflow_id"`
	ExecutionID string                 `json:"execution_id"`
	Success     bool                   `json:"success"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms"`
	NodeResults map[string]interface{} `json:"node_results"`
}

// InMemoryWorkflowExecutionStore is an in-memory implementation of WorkflowExecutionStore
type InMemoryWorkflowExecutionStore struct {
	executions map[string]*WorkflowExecution
	mu         sync.RWMutex
}

// NewInMemoryWorkflowExecutionStore creates a new in-memory workflow execution store
func NewInMemoryWorkflowExecutionStore() *InMemoryWorkflowExecutionStore {
	return &InMemoryWorkflowExecutionStore{
		executions: make(map[string]*WorkflowExecution),
	}
}

// SaveExecution saves a workflow execution
func (s *InMemoryWorkflowExecutionStore) SaveExecution(ctx context.Context, execution *WorkflowExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[execution.ID] = execution
	return nil
}

// GetExecution gets a workflow execution by ID
func (s *InMemoryWorkflowExecutionStore) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exec, ok := s.executions[executionID]
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}
	return exec, nil
}

// ListExecutions lists all executions for a workflow
func (s *InMemoryWorkflowExecutionStore) ListExecutions(ctx context.Context, workflowID string) ([]*WorkflowExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*WorkflowExecution
	for _, exec := range s.executions {
		if exec.WorkflowID == workflowID {
			result = append(result, exec)
		}
	}
	return result, nil
}

// DeleteExecution deletes a workflow execution
func (s *InMemoryWorkflowExecutionStore) DeleteExecution(ctx context.Context, executionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.executions, executionID)
	return nil
}

// SkillLibrary manages available skills
type SkillLibrary interface {
	GetSkill(name string) (interface{}, bool)
	ListSkills() []string
	RegisterSkill(name string, skill interface{}) error
}

// ModelFactory creates model instances
type ModelFactory interface {
	CreateModel(modelType string) (model.ChatModel, error)
	GetModel(modelType string) (model.ChatModel, error)
	ListModels() []string
}

// Agent represents an agent in the workflow
type Agent interface {
	Run(ctx context.Context, input string) (string, error)
	SetOptions(options ...AgentOption) error
	GetInfo() *AgentInfo
}

// AgentOption configures an agent
type AgentOption func(*AgentConfig)

// AgentConfig contains agent configuration
type AgentConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Model       string                 `json:"model"`
	Instructions string                `json:"instructions"`
	Tools       []string               `json:"tools"`
	Memory      map[string]interface{} `json:"memory"`
	Options     map[string]interface{} `json:"options"`
}

// AgentInfo contains agent information
type AgentInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Capabilities []string              `json:"capabilities"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ChatModel is an alias for model.ChatModel
type ChatModel = model.ChatModel

// CheckpointStore stores workflow checkpoints
type CheckpointStore interface {
	SaveCheckpoint(ctx context.Context, checkpoint *Checkpoint) error
	GetCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error)
	ListCheckpoints(ctx context.Context, workflowID string) ([]*Checkpoint, error)
	DeleteCheckpoint(ctx context.Context, checkpointID string) error
}

// Checkpoint represents a workflow execution checkpoint
type Checkpoint struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	NodeID       string                 `json:"node_id"`
	RunID        string                 `json:"run_id,omitempty"`
	WorkflowName string                 `json:"workflow_name,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Input        string                 `json:"input,omitempty"`
	Output       string                 `json:"output,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Progress     float64                `json:"progress,omitempty"`
	State        map[string]interface{} `json:"state"`
	Timestamp    time.Time              `json:"timestamp"`
	CreatedAt    time.Time              `json:"created_at,omitempty"`
	UpdatedAt    time.Time              `json:"updated_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Workflow status constants
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
	StatusSuspended = "suspended"
)

// Common workflow errors
var (
	ErrSuspended = errors.New("workflow suspended")
	ErrCancelled = errors.New("workflow cancelled")
)

// WorkflowCallbacks defines callbacks for workflow execution events
type WorkflowCallbacks interface {
	OnWorkflowStart(ctx context.Context, workflowID string, input string)
	OnWorkflowEnd(ctx context.Context, workflowID string, output string, status WorkflowExecutionStatus)
	OnNodeStart(ctx context.Context, nodeID string, input string)
	OnNodeEnd(ctx context.Context, nodeID string, output string)
}

// contextKey is a type for context keys to avoid collisions
type contextKey string

const workflowCallbacksKey contextKey = "workflow_callbacks"

// WithWorkflowCallbacks adds workflow callbacks to the context
func WithWorkflowCallbacks(ctx context.Context, callbacks WorkflowCallbacks) context.Context {
	return context.WithValue(ctx, workflowCallbacksKey, callbacks)
}

// GetWorkflowCallbacks retrieves workflow callbacks from the context
func GetWorkflowCallbacks(ctx context.Context) WorkflowCallbacks {
	callbacks, ok := ctx.Value(workflowCallbacksKey).(WorkflowCallbacks)
	if !ok {
		return nil
	}
	return callbacks
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store
func NewMemoryCheckpointStore() CheckpointStore {
	return &memoryCheckpointStore{
		checkpoints: make(map[string]*Checkpoint),
	}
}

// memoryCheckpointStore is an in-memory implementation of CheckpointStore
type memoryCheckpointStore struct {
	checkpoints map[string]*Checkpoint
	mu          sync.RWMutex
}

// SaveCheckpoint saves a checkpoint
func (s *memoryCheckpointStore) SaveCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[checkpoint.ID] = checkpoint
	return nil
}

// GetCheckpoint retrieves a checkpoint
func (s *memoryCheckpointStore) GetCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp, ok := s.checkpoints[checkpointID]
	if !ok {
		return nil, errors.New("checkpoint not found")
	}
	return cp, nil
}

// ListCheckpoints lists all checkpoints for a workflow
func (s *memoryCheckpointStore) ListCheckpoints(ctx context.Context, workflowID string) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Checkpoint
	for _, cp := range s.checkpoints {
		if cp.WorkflowID == workflowID {
			result = append(result, cp)
		}
	}
	return result, nil
}

// DeleteCheckpoint deletes a checkpoint
func (s *memoryCheckpointStore) DeleteCheckpoint(ctx context.Context, checkpointID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.checkpoints, checkpointID)
	return nil
}


// WorkflowExecutionStatus represents workflow execution status
type WorkflowExecutionStatus string

const (
	WorkflowStatusPending   WorkflowExecutionStatus = "pending"
	WorkflowStatusRunning   WorkflowExecutionStatus = "running"
	WorkflowStatusCompleted WorkflowExecutionStatus = "completed"
	WorkflowStatusFailed    WorkflowExecutionStatus = "failed"
	WorkflowStatusCancelled WorkflowExecutionStatus = "cancelled"
)

// Workforce represents a collection of workers for task execution
type Workforce interface {
	AddWorker(worker WorkerRole) error
	RemoveWorker(workerID string) error
	GetWorker(workerID string) (WorkerRole, bool)
	ListWorkers() []WorkerRole
	AssignTask(taskID string, workerID string) error
	GetStatus() WorkforceStatus
}

// WorkforceStatus represents the status of a workforce
type WorkforceStatus struct {
	TotalWorkers int       `json:"total_workers"`
	ActiveWorkers int       `json:"active_workers"`
	IdleWorkers  int       `json:"idle_workers"`
	TasksPending int       `json:"tasks_pending"`
	TasksRunning int       `json:"tasks_running"`
}

// WorkerRole represents a worker in the workforce
type WorkerRole interface {
	ID() string
	Name() string
	Type() string
	Capabilities() []string
	Execute(ctx context.Context, task interface{}) (interface{}, error)
	Status() WorkerStatus
	IsAvailable() bool
}

// WorkerStatus represents the status of a worker
type WorkerStatus struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"` // idle, busy, error, offline
	CurrentTask string   `json:"current_task,omitempty"`
	TasksCompleted int64  `json:"tasks_completed"`
	TasksFailed    int64  `json:"tasks_failed"`
	LastActive time.Time `json:"last_active"`
}

// GraphWorkforce extends Workforce with graph-based coordination
type GraphWorkforce interface {
	Workforce
	AddDependency(fromWorkerID, toWorkerID string) error
	RemoveDependency(fromWorkerID, toWorkerID string) error
	GetDependencyGraph() map[string][]string
	OptimizeAssignment(task interface{}) (string, error)
}

// TaskCoordinator coordinates task distribution across workers
type TaskCoordinator interface {
	AssignTask(ctx context.Context, task interface{}) (string, error)
	ReassignTask(ctx context.Context, taskID string, fromWorkerID, toWorkerID string) error
	GetTaskStatus(taskID string) (*TaskStatus, error)
	ListTasks(workerID string) ([]string, error)
	CancelTask(ctx context.Context, taskID string) error
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	ID          string                 `json:"id"`
	WorkerID    string                 `json:"worker_id"`
	Status      string                 `json:"status"` // pending, running, completed, failed, cancelled
	Input       interface{}            `json:"input"`
	Output      interface{}            `json:"output"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowDefinition defines a workflow for parsing from JSON/YAML
type WorkflowDefinition struct {
	Type        string                    `json:"type" yaml:"type"`
	Name        string                    `json:"name" yaml:"name"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Nodes       map[string]NodeDefinition `json:"nodes" yaml:"nodes"`
	Edges       []WorkflowEdgeDefinition  `json:"edges,omitempty" yaml:"edges,omitempty"`
	Variables   map[string]interface{}    `json:"variables,omitempty" yaml:"variables,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NodeDefinition defines a node in a workflow
type NodeDefinition struct {
	Type   string                 `json:"type" yaml:"type"`
	Name   string                 `json:"name" yaml:"name"`
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// WorkflowEdgeDefinition defines an edge between nodes in a workflow
type WorkflowEdgeDefinition struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// ParseWorkflowDefinition parses a JSON/YAML workflow definition string into a WorkflowDefinition struct
func ParseWorkflowDefinition(definition string) (*WorkflowDefinition, error) {
	var wfDef WorkflowDefinition

	// Try JSON parsing first
	if err := json.Unmarshal([]byte(definition), &wfDef); err != nil {
		// Try YAML parsing if JSON fails
		if err := yaml.Unmarshal([]byte(definition), &wfDef); err != nil {
			return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
		}
	}

	// Validate required fields
	if wfDef.Type == "" {
		return nil, fmt.Errorf("workflow type is required")
	}

	return &wfDef, nil
}

// CreateWorkflowFromDefinition creates a WorkflowInterface instance from a WorkflowDefinition
func CreateWorkflowFromDefinition(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (WorkflowInterface, error) {
	switch def.Type {
	case "sequential":
		return createSequentialWorkflowFromDef(def, skillLibrary, modelFactory)
	case "parallel":
		return createParallelWorkflowFromDef(def, skillLibrary, modelFactory)
	case "dag":
		return createDAGWorkflowFromDef(def, skillLibrary, modelFactory)
	case "graph":
		return createGraphWorkflowFromDef(def, skillLibrary, modelFactory)
	default:
		return nil, fmt.Errorf("unsupported workflow type: %s", def.Type)
	}
}

// createSequentialWorkflowFromDef creates a sequential workflow from definition
func createSequentialWorkflowFromDef(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (WorkflowInterface, error) {
	// For now, create a simple DAG workflow with sequential edges
	dagWf := NewDAGWorkflow(def.Name)

	// Create a list of node IDs in order
	var nodeIDs []string
	for nodeID := range def.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}

	// Add nodes to the workflow (empty for now, as we need agents)
	for i, nodeID := range nodeIDs {
		// Create a placeholder node configuration
		dagWf.AddNode(nodeID, nil) // nil will be replaced with actual workflow/agent later

		// Add edges to create sequential execution
		if i > 0 {
			dagWf.AddEdge(nodeIDs[i-1], nodeID)
		}
	}

	return dagWf, nil
}

// createParallelWorkflowFromDef creates a parallel workflow from definition
func createParallelWorkflowFromDef(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (WorkflowInterface, error) {
	dagWf := NewDAGWorkflow(def.Name)

	// Add all nodes without edges (parallel execution)
	for nodeID := range def.Nodes {
		dagWf.AddNode(nodeID, nil)
	}

	return dagWf, nil
}

// createDAGWorkflowFromDef creates a DAG workflow from definition
func createDAGWorkflowFromDef(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (WorkflowInterface, error) {
	dagWf := NewDAGWorkflow(def.Name)

	// Add all nodes
	for nodeID := range def.Nodes {
		dagWf.AddNode(nodeID, nil)
	}

	// Add edges
	for _, edge := range def.Edges {
		dagWf.AddEdge(edge.From, edge.To)
	}

	return dagWf, nil
}

// createGraphWorkflowFromDef creates a graph workflow from definition
func createGraphWorkflowFromDef(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (WorkflowInterface, error) {
	// Graph workflow is similar to DAG for now
	return createDAGWorkflowFromDef(def, skillLibrary, modelFactory)
}
