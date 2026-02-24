// Agent Framework - Agent Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	beadscontext "AgentFramework/pkg/beads/context"

	"github.com/cloudwego/eino/schema"

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
	EnableTrimming bool          `json:"enable_trimming"`
}

// MemoryManager manages agent memory and context
type MemoryManager interface {
	AddMessage(ctx context.Context, message *Message) error
	GetMessages(ctx context.Context, limit int) ([]*Message, error)
	Clear(ctx context.Context) error
	Summarize(ctx context.Context) (string, error)
	GetStats(ctx context.Context) (*MemoryStats, error)
	SetOptions(opts MemoryOptions)
	GetOptions() MemoryOptions
	ClearHistory() []*schema.Message
	ProcessMessage(msg *schema.Message) *schema.Message
	LimitHistory(messages []*schema.Message) []*schema.Message
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
	Transition(ctx context.Context, state string, reason string, metadata map[string]any) error
	// Additional methods for enhanced state machine functionality
	Current() AgentState
	History() []StateTransition
	IsTerminal() bool
	IsActive() bool
	AddHook(state AgentState, hook StateHook)
	OnTransition(callback func(transition StateTransition))
	Reset() error
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

// SkillInfo contains information about a skill (alias for SkillMetadata)
type SkillInfo = SkillMetadata

// Info returns skill information (for SkillAgent compatibility)
func (s *Skill) Info(ctx context.Context) (*SkillInfo, error) {
	return &SkillInfo{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Version:     s.Version,
		Author:      "",
		Tags:        []string{},
		Capabilities: []string{},
		Parameters:  map[string]Parameter{},
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Now(),
	}, nil
}

// IsEnabled checks if the skill is enabled
func (s *Skill) IsEnabled(ctx context.Context) bool {
	return s.Enabled
}

// Invoke executes the skill with the given input
func (s *Skill) Invoke(ctx context.Context, input string) (map[string]interface{}, error) {
	if s.Handler == nil {
		return nil, fmt.Errorf("skill %s has no handler", s.Name)
	}

	// Convert string input to map input
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(input), &inputMap); err != nil {
		// If JSON parsing fails, use the input as a simple string
		inputMap = map[string]interface{}{
			"input": input,
		}
	}

	return s.Handler(ctx, inputMap)
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

// State constants (aliases for AgentState)
const (
	StateIdle     AgentState = AgentStateIdle
	StateRunning  AgentState = AgentStateRunning
	StatePaused   AgentState = AgentStatePaused
	StateError    AgentState = AgentStateError
	StateComplete AgentState = AgentStateComplete
	StateFinished AgentState = AgentStateComplete
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

// DefaultMemoryManager provides a simple in-memory implementation of MemoryManager
type DefaultMemoryManager struct {
	messages    []*Message
	maxMessages int
	maxTokens   int
}

// NewMemoryManager creates a new DefaultMemoryManager with the given options
func NewMemoryManager(opts MemoryOptions) MemoryManager {
	return &DefaultMemoryManager{
		messages:    make([]*Message, 0),
		maxMessages: opts.MaxMessages,
		maxTokens:   opts.MaxTokens,
	}
}

// AddMessage adds a message to memory
func (m *DefaultMemoryManager) AddMessage(ctx context.Context, message *Message) error {
	m.messages = append(m.messages, message)
	// Trim messages if exceeding max
	if m.maxMessages > 0 && len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}
	return nil
}

// GetMessages retrieves messages from memory
func (m *DefaultMemoryManager) GetMessages(ctx context.Context, limit int) ([]*Message, error) {
	if limit > 0 && len(m.messages) > limit {
		return m.messages[len(m.messages)-limit:], nil
	}
	return m.messages, nil
}

// Clear clears all messages from memory
func (m *DefaultMemoryManager) Clear(ctx context.Context) error {
	m.messages = make([]*Message, 0)
	return nil
}

// Summarize creates a summary of the messages in memory
func (m *DefaultMemoryManager) Summarize(ctx context.Context) (string, error) {
	// Simple implementation: return count of messages
	return fmt.Sprintf("Memory contains %d messages", len(m.messages)), nil
}

// GetStats returns statistics about memory usage
func (m *DefaultMemoryManager) GetStats(ctx context.Context) (*MemoryStats, error) {
	return &MemoryStats{
		TotalMessages: len(m.messages),
		TotalTokens:   0, // Would need to calculate actual tokens
		LastUpdated:   time.Now(),
	}, nil
}

// SetOptions updates the memory options (for DefaultMemoryManager, this is a no-op)
func (m *DefaultMemoryManager) SetOptions(opts MemoryOptions) {
	m.maxMessages = opts.MaxMessages
	m.maxTokens = opts.MaxTokens
}

// GetOptions returns the current memory options
func (m *DefaultMemoryManager) GetOptions() MemoryOptions {
	return MemoryOptions{
		MaxMessages: m.maxMessages,
		MaxTokens:   m.maxTokens,
	}
}

// ClearHistory clears and returns cleared messages (for compatibility with ChatAgent)
func (m *DefaultMemoryManager) ClearHistory() []*schema.Message {
	// Convert to schema.Message format
	result := make([]*schema.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		result = append(result, &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content,
		})
	}
	m.messages = make([]*Message, 0)
	return result
}

// ProcessMessage processes a message (for now, just returns it as-is)
func (m *DefaultMemoryManager) ProcessMessage(msg *schema.Message) *schema.Message {
	// Simple implementation - just return the message
	// A more sophisticated implementation would handle trimming, summarization, etc.
	return msg
}

// LimitHistory limits the message history based on memory options
func (m *DefaultMemoryManager) LimitHistory(messages []*schema.Message) []*schema.Message {
	if m.maxMessages > 0 && len(messages) > m.maxMessages {
		return messages[len(messages)-m.maxMessages:]
	}
	return messages
}

// DefaultStateMachine provides a simple implementation of StateMachine
type DefaultStateMachine struct {
	currentState          string
	allowedTransitions    map[string][]string
	transitions          []StateTransition
	hooks                 map[AgentState][]StateHook
	transitionCallbacks   []func(transition StateTransition)
}

// NewStateMachineWithDefaults creates a new DefaultStateMachine with default agent states
func NewStateMachineWithDefaults() StateMachine {
	return &DefaultStateMachine{
		currentState:       string(AgentStateIdle),
		allowedTransitions: map[string][]string{
			string(AgentStateIdle):     {string(AgentStateRunning), string(AgentStateError)},
			string(AgentStateRunning):  {string(AgentStateComplete), string(AgentStateError), string(AgentStatePaused)},
			string(AgentStatePaused):   {string(AgentStateRunning), string(AgentStateComplete), string(AgentStateError)},
			string(AgentStateError):    {string(AgentStateIdle)},
			string(AgentStateComplete): {string(AgentStateIdle)},
		},
	}
}

// CurrentState returns the current state
func (sm *DefaultStateMachine) CurrentState() string {
	return sm.currentState
}

// TransitionTo transitions to a new state
func (sm *DefaultStateMachine) TransitionTo(state string) error {
	if !sm.CanTransitionTo(state) {
		return fmt.Errorf("cannot transition from %s to %s", sm.currentState, state)
	}
	sm.currentState = state
	return nil
}

// CanTransitionTo checks if a transition to the given state is allowed
func (sm *DefaultStateMachine) CanTransitionTo(state string) bool {
	allowed, ok := sm.allowedTransitions[sm.currentState]
	if !ok {
		return false
	}
	for _, allowedState := range allowed {
		if allowedState == state {
			return true
		}
	}
	return false
}

// AddTransition adds a allowed transition between two states
func (sm *DefaultStateMachine) AddTransition(from, to string) error {
	if sm.allowedTransitions[from] == nil {
		sm.allowedTransitions[from] = []string{}
	}
	sm.allowedTransitions[from] = append(sm.allowedTransitions[from], to)
	return nil
}

// GetAllowedTransitions returns the list of states that can be transitioned to from the current state
func (sm *DefaultStateMachine) GetAllowedTransitions() []string {
	allowed, ok := sm.allowedTransitions[sm.currentState]
	if !ok {
		return []string{}
	}
	return allowed
}

// Transition is an alias for TransitionTo (for compatibility with existing code)
func (sm *DefaultStateMachine) Transition(ctx context.Context, state string, reason string, metadata map[string]any) error {
	fromState := sm.currentState
	if err := sm.TransitionTo(state); err != nil {
		return err
	}

	// Create state transition record
	transition := StateTransition{
		From:      fromState,
		To:        state,
		Timestamp: time.Now().Unix(),
		Reason:    reason,
		Metadata:  metadata,
	}
	sm.transitions = append(sm.transitions, transition)

	// Call hooks
	if hooks, ok := sm.hooks[AgentState(fromState)]; ok {
		for _, hook := range hooks {
			hook(fromState, state, transition)
		}
	}

	// Call transition callbacks
	for _, callback := range sm.transitionCallbacks {
		callback(transition)
	}

	return nil
}

// Current returns the current state as AgentState
func (sm *DefaultStateMachine) Current() AgentState {
	return AgentState(sm.currentState)
}

// History returns the state transition history
func (sm *DefaultStateMachine) History() []StateTransition {
	return sm.transitions
}

// IsTerminal checks if the current state is a terminal state (complete or error)
func (sm *DefaultStateMachine) IsTerminal() bool {
	return sm.currentState == string(AgentStateComplete) || sm.currentState == string(AgentStateError)
}

// IsActive checks if the current state is an active state (running or paused)
func (sm *DefaultStateMachine) IsActive() bool {
	return sm.currentState == string(AgentStateRunning) || sm.currentState == string(AgentStatePaused)
}

// AddHook adds a hook that will be called when transitioning from a specific state
func (sm *DefaultStateMachine) AddHook(state AgentState, hook StateHook) {
	if sm.hooks == nil {
		sm.hooks = make(map[AgentState][]StateHook)
	}
	sm.hooks[state] = append(sm.hooks[state], hook)
}

// OnTransition registers a callback that will be called on any state transition
func (sm *DefaultStateMachine) OnTransition(callback func(transition StateTransition)) {
	sm.transitionCallbacks = append(sm.transitionCallbacks, callback)
}

// Reset resets the state machine to the initial idle state
func (sm *DefaultStateMachine) Reset() error {
	sm.currentState = string(AgentStateIdle)
	sm.transitions = make([]StateTransition, 0)
	return nil
}

// StateHook is a function that gets called during state transitions
type StateHook func(from, to string, transition StateTransition)

// AddStateHook adds a hook that will be called during state transitions (placeholder)
func AddStateHook(hook StateHook) {
	// Placeholder for future implementation - this is a global hook registry
}
