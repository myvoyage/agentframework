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
	errutil "AgentFramework/internal/errors"
	reg "AgentFramework/internal/registry"
	hello "AgentFramework/internal/tools/hello"
	"testing"
)

// Test that invalid input parameters trigger schema validation error
func TestInputSchemaValidationFailure(t *testing.T) {
	reg := reg.NewInMemoryToolRegistry()
	// Register hello tool instance
	reg.RegisterTool(hello.NewHelloTool())
	// Bind a simple input schema requiring a string 'name'
	spec := helloToolSpec()
	reg.SetToolSpec("hello", spec)

	eng := NewPipelineEngine(reg)
	yamlData := []byte(`version: "0.1"
id: test
memory_context_size: 100
steps:
  - id: s1
    type: task
    tool: hello
    params:
      name: 123
    next: end
  - id: end
    type: end`)
	p, err := eng.LoadPipeline(yamlData)
	if err != nil {
		t.Fatalf("LoadPipeline error: %v", err)
	}
	_, err = eng.RunPipeline(context.Background(), p)
	if err == nil {
		t.Fatalf("expected schema validation error, got nil")
	}
	// Expect an AppError with schema validation code
	if se, ok := err.(*errutil.AppError); ok {
		if se.Code != errutil.ErrCodeSchemaValidation {
			t.Fatalf("expected schema validation error code, got %s", se.Code)
		}
	}
}

func helloToolSpec() ToolSpec {
	// minimal spec with input name as string
	var spec ToolSpec
	spec.Name = "hello"
	spec.Version = "0.1"
	spec.InputsSchema = map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
		"required":   []string{"name"},
	}
	return spec
}
