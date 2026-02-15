package pipelineengine

import (
	"context"
	reg "AgentFramework/internal/registry"
	"testing"
)

// Mock tool implementation for testing
type MockTool struct{ name string }

func (m *MockTool) Name() string    { return m.name }
func (m *MockTool) Version() string { return "0.1" }
func (m *MockTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestLoadAndRunPipelineBasic(t *testing.T) {
	reg := reg.NewInMemoryToolRegistry()
	reg.RegisterTool(&MockTool{name: "mock1"})
	// Set a simple spec for mock1 with no inputs
	reg.SetToolSpec("mock1", ToolSpec{Name: "mock1", Version: "0.1"})
	eng := NewPipelineEngine(reg)
	yamlData := []byte(`version: "0.1"
id: test
memory_context_size: 100
steps:
  - id: s1
    type: task
    tool: mock1
    params: {}
    next: end
  - id: end
    type: end`)
	p, err := eng.LoadPipeline(yamlData)
	if err != nil {
		t.Fatalf("LoadPipeline error: %v", err)
	}
	ctx := context.Background()
	exec, err := eng.RunPipeline(ctx, p)
	if err != nil {
		t.Fatalf("RunPipeline error: %v", err)
	}
	if len(exec.Outputs) == 0 {
		t.Fatalf("expected outputs, got none")
	}
}
