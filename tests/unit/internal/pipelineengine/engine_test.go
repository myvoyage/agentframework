// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


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
