// Agent Framework - Workflow Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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
	NodeID    string                 `json:"node_id"`
	Success   bool                   `json:"success"`
	Output    string                 `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration_ms"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// simpleAgentWorkflow is a simple agent workflow for testing
func simpleAgentWorkflow(name string) WorkflowInterface {
	return &DAGWorkflow{
		name: name,
	}
}

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
	ExecutionID string                 `json:"execution_id"`
	Success     bool                   `json:"success"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms"`
	NodeResults map[string]interface{} `json:"node_results"`
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
