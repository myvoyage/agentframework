package agent

import (
	"context"
	"testing"
)

// TestMultiTurnChat tests that ChatAgent can handle multiple conversation turns
func TestMultiTurnChat(t *testing.T) {
	ctx := context.Background()

	// Create a mock model (using existing MockChatModel from test_utils.go)
	mockModel := &MockChatModel{NameValue: "test-model"}

	// Create ChatAgent
	chatAgent, err := NewChatAgent(ctx, ChatAgentConfig{
		Name:         "test-agent",
		Instructions: "You are a helpful assistant.",
		Model:        mockModel,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// First turn
	t.Log("=== First turn ===")
	t.Logf("State before: %s", chatAgent.State())

	_, err = chatAgent.Run(ctx, "你好")
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}
	t.Logf("State after first turn: %s", chatAgent.State())

	// Verify state is FINISHED
	if chatAgent.State() != StateFinished {
		t.Errorf("Expected state FINISHED, got %s", chatAgent.State())
	}

	// Second turn (previously would fail with: state transition failed: invalid state transition from FINISHED to RUNNING)
	t.Log("=== Second turn ===")
	t.Logf("State before: %s", chatAgent.State())

	_, err = chatAgent.Run(ctx, "我的名字是什么？")
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}
	t.Logf("State after second turn: %s", chatAgent.State())

	// Third turn
	t.Log("=== Third turn ===")
	t.Logf("State before: %s", chatAgent.State())

	_, err = chatAgent.Run(ctx, "再见")
	if err != nil {
		t.Fatalf("Third turn failed: %v", err)
	}
	t.Logf("State after third turn: %s", chatAgent.State())

	t.Log("✅ Multi-turn chat test passed!")
}
