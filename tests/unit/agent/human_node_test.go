package agent

import (
	"context"
	"testing"
)

func TestHumanNodeCreation(t *testing.T) {
	// Test basic HumanNode creation
	node := NewHumanNode("test_human", "Please provide input")
	if node.Name() != "test_human" {
		t.Errorf("Expected node name 'test_human', got '%s'", node.Name())
	}

	// Test HumanNode with options
	nodeWithOptions := NewHumanNode("test_human_with_opts", "Please provide input",
		WithFormSchema(map[string]any{"type": "object", "properties": map[string]any{"input": map[string]any{"type": "string"}}}),
		WithSupportedActions("approve", "reject", "skip"),
		WithTimeout(3600),
	)

	if len(nodeWithOptions.GetSupportedActions()) != 3 {
		t.Errorf("Expected 3 supported actions, got %d", len(nodeWithOptions.GetSupportedActions()))
	}

	if nodeWithOptions.GetTimeout() != 3600 {
		t.Errorf("Expected timeout 3600, got %d", nodeWithOptions.GetTimeout())
	}
}

func TestHumanNodeRun(t *testing.T) {
	ctx := context.Background()
	node := NewHumanNode("test_human", "Please provide input")

	// Test basic run - should return ErrSuspended
	_, err := node.Run(ctx, "test input")
	if err != ErrSuspended {
		t.Errorf("Expected ErrSuspended, got %v", err)
	}

	// Test run with resume input
	resumeCtx := context.WithValue(ctx, ResumeInputKey, "resume input")
	resp, err := node.Run(resumeCtx, "test input")
	if err != nil {
		t.Errorf("Expected no error for resume run, got %v", err)
	}

	if resp == nil || resp.Content != "resume input" {
		t.Errorf("Expected response content 'resume input', got %v", resp)
	}
}

func TestHumanNodeMethods(t *testing.T) {
	node := NewHumanNode("test_human", "Please provide input",
		WithFormSchema(map[string]any{"type": "object"}),
		WithAutoProceed("approve"),
		WithMetadata(map[string]string{"key": "value"}),
	)

	// Test GetFormSchema
	if node.GetFormSchema() == nil {
		t.Errorf("Expected form schema, got nil")
	}

	// Test GetSupportedActions
	if len(node.GetSupportedActions()) != 2 {
		t.Errorf("Expected default 2 supported actions, got %d", len(node.GetSupportedActions()))
	}

	// Test IsAutoProceed
	if !node.IsAutoProceed() {
		t.Errorf("Expected auto-proceed to be true")
	}

	// Test GetDefaultAction
	if node.GetDefaultAction() != "approve" {
		t.Errorf("Expected default action 'approve', got '%s'", node.GetDefaultAction())
	}

	// Test GetMetadata
	if node.GetMetadata()["key"] != "value" {
		t.Errorf("Expected metadata 'key' to be 'value', got '%s'", node.GetMetadata()["key"])
	}

	// Test WithFormSchema method
	node.WithFormSchema(map[string]any{"type": "array"})
	if node.GetFormSchema()["type"] != "array" {
		t.Errorf("Expected form schema type 'array', got '%v'", node.GetFormSchema()["type"])
	}

	// Test WithSupportedActions method
	node.WithSupportedActions("action1", "action2")
	if len(node.GetSupportedActions()) != 2 {
		t.Errorf("Expected 2 supported actions, got %d", len(node.GetSupportedActions()))
	}

	// Test WithTimeout method
	node.WithTimeout(7200)
	if node.GetTimeout() != 7200 {
		t.Errorf("Expected timeout 7200, got %d", node.GetTimeout())
	}
}
