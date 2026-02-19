// Agent Framework - Agent Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"time"

	beadscontext "AgentFramework/pkg/beads/context"

	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/framework/workflow"
)

// MemoryOptions defines options for memory management
type MemoryOptions struct {
	MaxMessages    int           `json:"max_messages"`
	MaxTokens      int           `json:"max_tokens"`
	RetentionTime  time.Duration `json:"retention_time"`
	EnableSummary  bool          `json:"enable_summary"`
	SummaryThreshold int         `json:"summary_threshold"`
}

// MemoryManager manages agent memory and context
type MemoryManager interface {
	AddMessage(ctx context.Context, message *Message) error
	GetMessages(ctx context.Context, limit int) ([]*Message, error)
	Clear(ctx context.Context) error
	Summarize(ctx context.Context) (string, error)
	GetStats(ctx context.Context) (*MemoryStats, error)
}

// Message represents a message in agent memory
type Message struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // user, assistant, system
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	TotalMessages int       `json:"total_messages"`
	TotalTokens   int       `json:"total_tokens"`
	LastUpdated   time.Time `json:"last_updated"`
}

// StateMachine manages agent state transitions
type StateMachine interface {
	CurrentState() string
	TransitionTo(state string) error
	CanTransitionTo(state string) bool
	AddTransition(from, to string) error
	GetAllowedTransitions() []string
}

// ChartGenerator generates charts for data visualization
type ChartGenerator interface {
	GenerateChart(data interface{}, chartType string) ([]byte, error)
	GenerateBarChart(data map[string]interface{}) ([]byte, error)
	GenerateLineChart(data []interface{}) ([]byte, error)
	GeneratePieChart(data map[string]float64) ([]byte, error)
}

// Skill represents an agent skill
type Skill struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Handler     func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
	Enabled     bool                   `json:"enabled"`
}

// CheckpointStore defines the interface for storing workflow checkpoints
type CheckpointStore interface {
	SaveCheckpoint(checkpoint *Checkpoint) error
	GetCheckpoint(workflowID, checkpointID string) (*Checkpoint, error)
	ListCheckpoints(workflowID string) ([]*Checkpoint, error)
	DeleteCheckpoint(workflowID, checkpointID string) error
}

// Checkpoint represents a workflow execution checkpoint
type Checkpoint struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflow_id"`
	State      map[string]interface{} `json:"state"`
	Timestamp  int64                  `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// AgentState represents the current state of an agent
type AgentState string

const (
	AgentStateIdle     AgentState = "idle"
	AgentStateRunning  AgentState = "running"
	AgentStatePaused   AgentState = "paused"
	AgentStateError    AgentState = "error"
	AgentStateComplete AgentState = "complete"
)

// StateTransition represents a state transition event
type StateTransition struct {
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Timestamp int64                  `json:"timestamp"`
	Reason    string                 `json:"reason"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SkillMetadata represents metadata about a skill
type SkillMetadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Tags        []string               `json:"tags"`
	Capabilities []string              `json:"capabilities"`
	Parameters  map[string]Parameter   `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Parameter represents a skill parameter
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// RealTimeStats is an alias for context.RealTimeStats
type RealTimeStats = beadscontext.RealTimeStats

// Query is an alias for context.Query
type Query = beadscontext.Query

// QueryResult is an alias for context.QueryResult
type QueryResult = beadscontext.QueryResult

// SearchResult is an alias for context.SearchResult
type SearchResult = beadscontext.SearchResult

// WorkflowState is an alias for memory.WorkflowState
type WorkflowState = memory.WorkflowState

// SkillLibrary is an alias for workflow.SkillLibrary
type SkillLibrary = workflow.SkillLibrary

// ModelFactory is an alias for workflow.ModelFactory
type ModelFactory = workflow.ModelFactory
