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

type MockLoopTool struct{}

func (m *MockLoopTool) Name() string    { return "mock-loop" }
func (m *MockLoopTool) Version() string { return "0.1" }
func (m *MockLoopTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestPipelineWithLoop(t *testing.T) {
	reg := reg.NewInMemoryToolRegistry()
	reg.RegisterTool(&MockLoopTool{})
	eng := NewPipelineEngine(reg)

	yamlData := []byte(`version: "0.1"
id: loop-pipe
memory_context_size: 100
steps:
  - id: s1
    type: task
    tool: mock-loop
    params: {}
    next: s2
  - id: s2
    type: loop
    loop:
      var: x
      do: s3
    next: s3
  - id: s3
    type: task
    tool: mock-loop
    params: {}
    next: end
  - id: end
    type: end`)
	p, err := eng.LoadPipeline(yamlData)
	if err != nil {
		t.Fatalf("LoadPipeline failed: %v", err)
	}
	ctx := context.Background()
	exec, err := eng.RunPipeline(ctx, p)
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}
	if len(exec.Outputs) == 0 {
		t.Fatalf("expected outputs from loop path, got none: %#v", exec.Outputs)
	}
}
