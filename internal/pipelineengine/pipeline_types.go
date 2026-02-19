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
	"time"

	"AgentFramework/internal/types"
)

// Core data structures for the MVP Go + Eino MVP pipeline engine
type PipelineSpec struct {
	Version           string         `yaml:"version"`
	Id                string         `yaml:"id"`
	Name              string         `yaml:"name"`
	MemoryContextSize int            `yaml:"memory_context_size"`
	Steps             []PipelineStep `yaml:"steps"`
}

type PipelineStep struct {
	Id        string                 `yaml:"id"`
	Type      string                 `yaml:"type"`
	Tool      string                 `yaml:"tool"`
	Params    map[string]interface{} `yaml:"params"`
	Next      string                 `yaml:"next"`
	Condition string                 `yaml:"condition"`
	Loop      *LoopSpec              `yaml:"loop"`
}

type LoopSpec struct {
	Var string `yaml:"var"`
	Do  string `yaml:"do"`
}

type ExecutionContext struct {
	PipelineID string
	StepIndex  int
	Outputs    map[string]interface{}
	Memory     map[string]interface{}
	Logs       []string
	CreatedAt  time.Time
}

type MCPRequest struct {
	Tool    string                 `json:"tool"`
	Params  map[string]interface{} `json:"params"`
	Context map[string]interface{} `json:"context"`
	Version string                 `json:"version"`
}

type MCPResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error"`
}

// Tool interface and registry abstraction for MVP (using shared types)
type Tool = types.Tool
type ToolSpec = types.ToolSpec
type ToolRegistry = types.ToolRegistry
