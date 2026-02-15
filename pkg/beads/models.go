package beads

import (
	"time"
)

// TaskType represents the type of a task
type TaskType string

const (
	TaskTypeEpic       TaskType = "epic"
	TaskTypeTask       TaskType = "task"
	TaskTypeBug        TaskType = "bug"
	TaskTypeFeature    TaskType = "feature"
	TaskTypeResearch   TaskType = "research"
	TaskTypeCheckpoint TaskType = "checkpoint"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	StatusOpen       TaskStatus = "open"
	StatusInProgress TaskStatus = "in_progress"
	StatusBlocked    TaskStatus = "blocked"
	StatusCompleted  TaskStatus = "completed"
	StatusCancelled  TaskStatus = "cancelled"
)

// Task represents a unit of work with metadata, status, dependencies, and Git-tracked history
type Task struct {
	ID          string            `json:"id"`
	Type        TaskType          `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      TaskStatus        `json:"status"`
	Assignee    string            `json:"assignee,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DependencyType represents the type of relationship between tasks
type DependencyType string

const (
	DependencyTypeBlocks         DependencyType = "blocks"
	DependencyTypeParentChild    DependencyType = "parent-child"
	DependencyTypeRelated        DependencyType = "related"
	DependencyTypeDiscoveredFrom DependencyType = "discovered-from"
)

// Dependency represents a relationship between two tasks
type Dependency struct {
	FromTaskID string         `json:"from_task_id"`
	ToTaskID   string         `json:"to_task_id"`
	Type       DependencyType `json:"type"`
	CreatedAt  time.Time      `json:"created_at"`
}

// EventType represents the type of task event
type EventType string

const (
	EventTaskCreated       EventType = "task_created"
	EventTaskUpdated       EventType = "task_updated"
	EventTaskClosed        EventType = "task_closed"
	EventDependencyAdded   EventType = "dependency_added"
	EventDependencyRemoved EventType = "dependency_removed"
)

// Event represents an immutable task event in the event sourcing system
type Event struct {
	Type       EventType              `json:"event_type"`
	TaskID     string                 `json:"task_id,omitempty"`
	FromTaskID string                 `json:"from_task_id,omitempty"`
	ToTaskID   string                 `json:"to_task_id,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`
}

// Config represents the configuration for the task tracker
type Config struct {
	StoragePath  string        `json:"storage_path"`
	GitEnabled   bool          `json:"git_enabled"`
	SyncInterval time.Duration `json:"sync_interval"`
	MaxTasks     int           `json:"max_tasks"`
	DBPath       string        `json:"db_path"`
	JSONLPath    string        `json:"jsonl_path"`
}

// TaskUpdate represents fields that can be updated on a task
type TaskUpdate struct {
	Title       *string            `json:"title,omitempty"`
	Description *string            `json:"description,omitempty"`
	Status      *TaskStatus        `json:"status,omitempty"`
	Assignee    *string            `json:"assignee,omitempty"`
	Tags        *[]string          `json:"tags,omitempty"`
	Metadata    *map[string]string `json:"metadata,omitempty"`
}

// LogicalOp represents logical operations for tag queries
type LogicalOp string

const (
	LogicalOpAND LogicalOp = "AND"
	LogicalOpOR  LogicalOp = "OR"
)

// Direction represents the direction for dependency queries
type Direction string

const (
	DirectionIncoming Direction = "incoming"
	DirectionOutgoing Direction = "outgoing"
)

// Query represents a flexible query for tasks
type Query struct {
	Status   *TaskStatus
	Assignee *string
	Tags     []string
	TagOp    LogicalOp
}

// SyncStatus represents the current status of the sync daemon
type SyncStatus struct {
	Running      bool      `json:"running"`
	LastSyncTime time.Time `json:"last_sync_time"`
	ErrorCount   int       `json:"error_count"`
	LastError    string    `json:"last_error,omitempty"`
}

// Error represents a structured error response
type Error struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retryable bool                   `json:"retryable"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}
