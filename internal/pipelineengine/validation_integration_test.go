package pipelineengine

import (
	"context"
	"testing"

	errutil "AgentFramework/internal/errors"
	reg "AgentFramework/internal/registry"
)

// Simple mock tool with input/output schemas
type MockInputsTool struct{}

func (m *MockInputsTool) Name() string    { return "mockin" }
func (m *MockInputsTool) Version() string { return "0.1" }
func (m *MockInputsTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	name, _ := inputs["name"].(string)
	if name == "" {
		name = "world"
	}
	return map[string]interface{}{"greeting": "Hi " + name}, nil
}

func (m *MockInputsTool) _assertImplemented() {}

var _ Tool = (*MockInputsTool)(nil)

func TestValidationIntegration_InputSchema_OK(t *testing.T) {
	regSvc := reg.NewInMemoryToolRegistry()
	regSvc.RegisterTool(&MockInputsTool{})
	spec := ToolSpec{
		Name:    "mockin",
		Version: "0.1",
		InputsSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
			"required":   []string{"name"},
		},
		OutputsSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"greeting": map[string]interface{}{"type": "string"}},
			"required":   []string{"greeting"},
		},
	}
	regSvc.SetToolSpec("mockin", spec)
	eng := NewPipelineEngine(regSvc)
	yamlData := []byte(`version: "0.1"
id: test
memory_context_size: 100
steps:
  - id: s1
    type: task
    tool: mockin
    params: {"name": "Alice"}
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
	if exec.Outputs["s1"] == nil {
		t.Fatalf("expected outputs, got none: %#v", exec.Outputs)
	}
}

func TestValidationIntegration_InputSchema_Failure(t *testing.T) {
	regSvc := reg.NewInMemoryToolRegistry()
	regSvc.RegisterTool(&MockInputsTool{})
	spec := ToolSpec{
		Name:    "mockin",
		Version: "0.1",
		InputsSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
			"required":   []string{"name"},
		},
		OutputsSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"greeting": map[string]interface{}{"type": "string"}},
			"required":   []string{"greeting"},
		},
	}
	regSvc.SetToolSpec("mockin", spec)
	eng := NewPipelineEngine(regSvc)
	yamlData := []byte(`version: "0.1"
id: test
memory_context_size: 100
steps:
  - id: s1
    type: task
    tool: mockin
    params: {"name": 123}
    next: end
  - id: end
    type: end`)
	p, err := eng.LoadPipeline(yamlData)
	if err != nil {
		t.Fatalf("LoadPipeline error: %v", err)
	}
	ctx := context.Background()
	_, err = eng.RunPipeline(ctx, p)
	if err == nil {
		t.Fatalf("expected schema validation error for input, got nil")
	}
	if se, ok := err.(*errutil.AppError); ok {
		if se.Code != errutil.ErrCodeSchemaValidation {
			t.Fatalf("expected schema validation error code, got %s", se.Code)
		}
	}
}
