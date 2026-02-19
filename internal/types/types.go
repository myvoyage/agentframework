// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package types

// Tool interface and registry abstraction for MVP
type Tool interface {
	Name() string
	Version() string
	Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error)
}

type ToolSpec struct {
	Name          string                 `yaml:"name"`
	Version       string                 `yaml:"version"`
	Description   string                 `yaml:"description"`
	InputsSchema  map[string]interface{} `yaml:"inputs_schema"`
	OutputsSchema map[string]interface{} `yaml:"outputs_schema"`
	HandlerRef    string                 `yaml:"handler_ref"`
}

type ToolRegistry interface {
	RegisterTool(t Tool) error
	GetTool(name string, version string) (Tool, error)
	ListTools() []ToolSpec
	GetToolSpec(name string) (ToolSpec, error)
}