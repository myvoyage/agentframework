// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package registry

import (
	"testing"
)

type MockRegistryTool struct{}

func (m *MockRegistryTool) Name() string    { return "mock3" }
func (m *MockRegistryTool) Version() string { return "0.1" }
func (m *MockRegistryTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func TestInMemoryToolRegistry_RegisterAndGet(t *testing.T) {
	reg := NewInMemoryToolRegistry()
	reg.RegisterTool(&MockRegistryTool{})
	tool, err := reg.GetTool("mock3", "")
	if err != nil {
		t.Fatalf("failed to get tool: %v", err)
	}
	if tool.Name() != "mock3" {
		t.Fatalf("expected mock3, got %s", tool.Name())
	}
}
