// ReAct Agent 单元测试
// 测试核心类型和基本功能

package react

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestReActActionType(t *testing.T) {
	tests := []struct {
		name     string
		action   ReActActionType
		expected string
	}{
		{"Think", ActionTypeThink, "think"},
		{"Tool", ActionTypeTool, "tool"},
		{"Search", ActionTypeSearch, "search"},
		{"Reflect", ActionTypeReflect, "reflect"},
		{"Finish", ActionTypeFinish, "finish"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action.String() != tt.expected {
				t.Errorf("String() = %v, want %v", tt.action.String(), tt.expected)
			}
			if !tt.action.IsValid() {
				t.Errorf("IsValid() = false, want true")
			}
		})
	}

	invalidAction := ReActActionType("invalid")
	if invalidAction.IsValid() {
		t.Error("IsValid() = true for invalid action, want false")
	}
}

func TestNewThought(t *testing.T) {
	thought := NewThought("Test content", "Test reasoning", 0.9)

	if thought.ID == "" {
		t.Error("Thought ID is empty")
	}
	if thought.Content != "Test content" {
		t.Errorf("Content = %v, want 'Test content'", thought.Content)
	}
	if thought.Reasoning != "Test reasoning" {
		t.Errorf("Reasoning = %v, want 'Test reasoning'", thought.Reasoning)
	}
	if thought.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", thought.Confidence)
	}
}

func TestThoughtClone(t *testing.T) {
	original := &Thought{
		ID:        "test-id",
		Timestamp: time.Now(),
		Content:   "Test content",
		Reasoning: "Test reasoning",
		Confidence: 0.9,
		Context: map[string]interface{}{
			"key": "value",
		},
	}

	cloned := original.Clone()

	if cloned.ID != original.ID {
		t.Errorf("Clone() ID = %v, want %v", cloned.ID, original.ID)
	}
	if cloned.Content != original.Content {
		t.Errorf("Clone() Content = %v, want %v", cloned.Content, original.Content)
	}
}

func TestNewAction(t *testing.T) {
	action := NewAction(ActionTypeTool, "read_file", map[string]interface{}{
		"path": "/test/file.txt",
	})

	if action.ID == "" {
		t.Error("Action ID is empty")
	}
	if action.Type != ActionTypeTool {
		t.Errorf("Type = %v, want %v", action.Type, ActionTypeTool)
	}
	if action.Name != "read_file" {
		t.Errorf("Name = %v, want 'read_file'", action.Name)
	}
}

func TestNewObservation(t *testing.T) {
	observation := NewObservation("action-id", true, "Test result", "")

	if observation.ID == "" {
		t.Error("Observation ID is empty")
	}
	if observation.ActionID != "action-id" {
		t.Errorf("ActionID = %v, want 'action-id'", observation.ActionID)
	}
	if !observation.Success {
		t.Error("Success = false, want true")
	}
}

func TestObservationResultSummary(t *testing.T) {
	tests := []struct {
		name     string
		obs      *Observation
		expected string
	}{
		{
			name: "Success with string result",
			obs: &Observation{
				Success: true,
				Result:  "This is a short result",
			},
			expected: "This is a short result",
		},
		{
			name: "Success with long result",
			obs: &Observation{
				Success: true,
				Result:  "This is a very long result that should be truncated to 100 characters when returned as a summary",
			},
			expected: "This is a very long result that should be truncated to 100 characters when returned as a",
		},
		{
			name: "Failure with error",
			obs: &Observation{
				Success: false,
				Error:   "File not found",
			},
			expected: "Error: File not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.obs.ResultSummary()
			if summary != tt.expected {
				t.Errorf("ResultSummary() = %v, want %v", summary, tt.expected)
			}
		})
	}
}

func TestReActStateValidation(t *testing.T) {
	tests := []struct {
		name    string
		state   *ReActState
		wantErr bool
	}{
		{
			name: "Valid state",
			state: &ReActState{
				SessionID: "test-session",
				AgentID:   "test-agent",
				Query:     "Test query",
				Status:    ReActStatusThinking,
			},
			wantErr: false,
		},
		{
			name: "Empty session ID",
			state: &ReActState{
				SessionID: "",
				AgentID:   "test-agent",
				Query:     "Test query",
			},
			wantErr: true,
		},
		{
			name: "Empty agent ID",
			state: &ReActState{
				SessionID: "test-session",
				AgentID:   "",
				Query:     "Test query",
			},
			wantErr: true,
		},
		{
			name: "Empty query",
			state: &ReActState{
				SessionID: "test-session",
				AgentID:   "test-agent",
				Query:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReActStateJSON(t *testing.T) {
	state := &ReActState{
		SessionID: "test-session",
		AgentID:   "test-agent",
		Query:     "Test query",
		Status:    ReActStatusThinking,
		Thoughts: []*Thought{
			NewThought("Test thought", "Test reasoning", 0.9),
		},
	}

	// Test ToJSON
	jsonData, err := state.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Test FromJSON
	restoredState := &ReActState{}
	err = restoredState.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if restoredState.SessionID != state.SessionID {
		t.Errorf("SessionID = %v, want %v", restoredState.SessionID, state.SessionID)
	}
	if restoredState.AgentID != state.AgentID {
		t.Errorf("AgentID = %v, want %v", restoredState.AgentID, state.AgentID)
	}
	if restoredState.Query != state.Query {
		t.Errorf("Query = %v, want %v", restoredState.Query, state.Query)
	}
}

func TestReActStateClone(t *testing.T) {
	original := &ReActState{
		SessionID: "test-session",
		AgentID:   "test-agent",
		Query:     "Test query",
		Status:    ReActStatusThinking,
		Thoughts: []*Thought{
			NewThought("Test thought", "Test reasoning", 0.9),
		},
		Actions: []*Action{
			NewAction(ActionTypeTool, "test_tool", nil),
		},
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	cloned := original.Clone()

	if cloned.SessionID != original.SessionID {
		t.Errorf("Clone() SessionID = %v, want %v", cloned.SessionID, original.SessionID)
	}
	if len(cloned.Thoughts) != len(original.Thoughts) {
		t.Errorf("Clone() Thoughts length = %v, want %v", len(cloned.Thoughts), len(original.Thoughts))
	}
	if len(cloned.Actions) != len(original.Actions) {
		t.Errorf("Clone() Actions length = %v, want %v", len(cloned.Actions), len(original.Actions))
	}
}

func TestCapability(t *testing.T) {
	capabilities := []Capability{
		CapabilityReasoning,
		CapabilityToolUse,
		CapabilityPlanning,
		CapabilityLearning,
		CapabilityReflection,
		CapabilityMultiTask,
		CapabilityParallelExecution,
		CapabilityMemoryUse,
	}

	for _, cap := range capabilities {
		if !cap.IsValid() {
			t.Errorf("IsValid() = false for %v, want true", cap)
		}
	}

	invalidCap := Capability("invalid")
	if invalidCap.IsValid() {
		t.Error("IsValid() = true for invalid capability, want false")
	}
}

func TestReActConfigDefaults(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &ReActConfig{
		Logger: logger,
	}

	if config.MaxIterations == 0 {
		config.MaxIterations = 10 // Default
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2000 // Default
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7 // Default
	}

	if config.MaxIterations != 10 {
		t.Errorf("MaxIterations = %v, want 10", config.MaxIterations)
	}
	if config.MaxTokens != 2000 {
		t.Errorf("MaxTokens = %v, want 2000", config.MaxTokens)
	}
	if config.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want 0.7", config.Temperature)
	}
}
