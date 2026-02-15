// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"testing"
)

// TestCreateReActAgentFromNodeDefinition tests the creation of ReActAgent from node definition
func TestCreateReActAgentFromNodeDefinition(t *testing.T) {

	// Create a mock model factory that returns a tool calling model
	modelFactory := func(ctx context.Context, modelName string) (ChatModel, error) {
		return NewMockToolCallingChatModel(modelName), nil
	}

	// Test case 1: Basic ReActAgent creation with default config
	t.Run("BasicReActAgent", func(t *testing.T) {
		nodeDef := NodeDefinition{
			Type: "agent",
			Name: "test_react_agent",
			Config: map[string]interface{}{
				"kind":         "react",
				"model":        "default",
				"instructions": "You are a helpful assistant.",
			},
		}

		agent, err := createReActAgentFromNodeDefinition("test_react", nodeDef, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create ReActAgent: %v", err)
		}

		if agent == nil {
			t.Fatal("Expected agent to be non-nil")
		}

		if agent.Name() != "test_react" {
			t.Errorf("Expected agent name 'test_react', got '%s'", agent.Name())
		}
	})

	// Test case 2: ReActAgent with custom max_iterations
	t.Run("ReActAgentWithMaxIterations", func(t *testing.T) {
		nodeDef := NodeDefinition{
			Type: "agent",
			Name: "test_react_agent_iterations",
			Config: map[string]interface{}{
				"kind":           "react",
				"model":          "default",
				"instructions":   "You are a helpful assistant.",
				"max_iterations": 15,
			},
		}

		agent, err := createReActAgentFromNodeDefinition("test_react_iter", nodeDef, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create ReActAgent with max_iterations: %v", err)
		}

		if agent == nil {
			t.Fatal("Expected agent to be non-nil")
		}
	})

	// Test case 3: ReActAgent with custom memory options
	t.Run("ReActAgentWithMemoryOptions", func(t *testing.T) {
		nodeDef := NodeDefinition{
			Type: "agent",
			Name: "test_react_agent_memory",
			Config: map[string]interface{}{
				"kind":         "react",
				"model":        "default",
				"instructions": "You are a helpful assistant.",
				"memory": map[string]interface{}{
					"enable_trimming": true,
					"max_messages":    30,
					"max_tokens":      5000,
					"preserve_recent": 10,
				},
			},
		}

		agent, err := createReActAgentFromNodeDefinition("test_react_mem", nodeDef, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create ReActAgent with memory options: %v", err)
		}

		if agent == nil {
			t.Fatal("Expected agent to be non-nil")
		}

		// Check if agent is a ReActAgent
		reactAgent, ok := agent.(*ReActAgent)
		if !ok {
			t.Fatal("Expected agent to be a ReActAgent")
		}

		// Verify memory options
		memOpts := reactAgent.GetMemoryOptions()
		if !memOpts.EnableTrimming {
			t.Error("Expected EnableTrimming to be true")
		}
		if memOpts.MaxMessages != 30 {
			t.Errorf("Expected MaxMessages to be 30, got %d", memOpts.MaxMessages)
		}
		if memOpts.TrimRatio != 0.7 {
			t.Errorf("Expected TrimRatio to be 0.7, got %f", memOpts.TrimRatio)
		}
	})

	// Test case 4: ReActAgent with default instructions
	t.Run("ReActAgentWithDefaultInstructions", func(t *testing.T) {
		nodeDef := NodeDefinition{
			Type: "agent",
			Name: "test_react_agent_default",
			Config: map[string]interface{}{
				"kind":  "react",
				"model": "default",
			},
		}

		agent, err := createReActAgentFromNodeDefinition("test_react_default", nodeDef, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create ReActAgent with default instructions: %v", err)
		}

		if agent == nil {
			t.Fatal("Expected agent to be non-nil")
		}
	})
}

// TestCreateReActAgentInWorkflow tests ReActAgent creation within a workflow definition
func TestCreateReActAgentInWorkflow(t *testing.T) {

	// Create a mock model factory that returns a tool calling model
	modelFactory := func(ctx context.Context, modelName string) (ChatModel, error) {
		return NewMockToolCallingChatModel(modelName), nil
	}

	// Create a mock skill library
	skillLibrary := NewSkillLibrary()

	// Test case 1: Sequential workflow with ReActAgent
	t.Run("SequentialWorkflowWithReActAgent", func(t *testing.T) {
		workflowDef := &WorkflowDefinition{
			Type: "sequential",
			Name: "test_sequential_react",
			Nodes: map[string]NodeDefinition{
				"react_node": {
					Type: "agent",
					Name: "react_agent",
					Config: map[string]interface{}{
						"kind":         "react",
						"model":        "default",
						"instructions": "You are a helpful assistant that solves problems step by step.",
					},
				},
			},
		}

		workflow, err := CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create workflow with ReActAgent: %v", err)
		}

		if workflow == nil {
			t.Fatal("Expected workflow to be non-nil")
		}

		if workflow.Name() != "test_sequential_react" {
			t.Errorf("Expected workflow name 'test_sequential_react', got '%s'", workflow.Name())
		}
	})

	// Test case 2: DAG workflow with ReActAgent
	t.Run("DAGWorkflowWithReActAgent", func(t *testing.T) {
		workflowDef := &WorkflowDefinition{
			Type: "dag",
			Name: "test_dag_react",
			Nodes: map[string]NodeDefinition{
				"start": {
					Type: "agent",
					Name: "start_agent",
					Config: map[string]interface{}{
						"kind":         "chat",
						"model":        "default",
						"instructions": "You are a starting agent.",
					},
				},
				"react_node": {
					Type: "agent",
					Name: "react_agent",
					Config: map[string]interface{}{
						"kind":           "react",
						"model":          "default",
						"instructions":   "You are a ReAct agent that processes information.",
						"max_iterations": 12,
					},
				},
			},
			Edges: []EdgeDefinition{
				{From: "start", To: "react_node"},
			},
		}

		workflow, err := CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
		if err != nil {
			t.Fatalf("Failed to create DAG workflow with ReActAgent: %v", err)
		}

		if workflow == nil {
			t.Fatal("Expected workflow to be non-nil")
		}
	})
}

// TestReActAgentFromJSONDefinition tests creating ReActAgent from JSON workflow definition
func TestReActAgentFromJSONDefinition(t *testing.T) {

	// Create a mock model factory that returns a tool calling model
	modelFactory := func(ctx context.Context, modelName string) (ChatModel, error) {
		return NewMockToolCallingChatModel(modelName), nil
	}

	// Create a mock skill library
	skillLibrary := NewSkillLibrary()

	jsonDef := `{
		"type": "sequential",
		"name": "test_react_json",
		"nodes": {
			"react_agent": {
				"type": "agent",
				"name": "ReAct Agent",
				"config": {
					"kind": "react",
					"model": "default",
					"instructions": "You are a helpful AI assistant.",
					"max_iterations": 10,
					"memory": {
						"enable_trimming": true,
						"max_messages": 25,
						"max_tokens": 4500
					}
				}
			}
		}
	}`

	// Parse workflow definition
	workflowDef, err := ParseWorkflowDefinition(jsonDef)
	if err != nil {
		t.Fatalf("Failed to parse workflow definition: %v", err)
	}

	// Create workflow from definition
	workflow, err := CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
	if err != nil {
		t.Fatalf("Failed to create workflow from JSON definition: %v", err)
	}

	if workflow == nil {
		t.Fatal("Expected workflow to be non-nil")
	}

	if workflow.Name() != "test_react_json" {
		t.Errorf("Expected workflow name 'test_react_json', got '%s'", workflow.Name())
	}
}

// TestReActAgentFromYAMLDefinition tests creating ReActAgent from YAML workflow definition
func TestReActAgentFromYAMLDefinition(t *testing.T) {

	// Create a mock model factory that returns a tool calling model
	modelFactory := func(ctx context.Context, modelName string) (ChatModel, error) {
		return NewMockToolCallingChatModel(modelName), nil
	}

	// Create a mock skill library
	skillLibrary := NewSkillLibrary()

	yamlDef := `
type: sequential
name: test_react_yaml
nodes:
  react_agent:
    type: agent
    name: ReAct Agent
    config:
      kind: react
      model: default
      instructions: You are a helpful AI assistant.
      max_iterations: 10
      memory:
        enable_trimming: true
        max_messages: 25
        max_tokens: 4500
`

	// Parse workflow definition
	workflowDef, err := ParseWorkflowDefinition(yamlDef)
	if err != nil {
		t.Fatalf("Failed to parse workflow definition: %v", err)
	}

	// Create workflow from definition
	workflow, err := CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
	if err != nil {
		t.Fatalf("Failed to create workflow from YAML definition: %v", err)
	}

	if workflow == nil {
		t.Fatal("Expected workflow to be non-nil")
	}

	if workflow.Name() != "test_react_yaml" {
		t.Errorf("Expected workflow name 'test_react_yaml', got '%s'", workflow.Name())
	}
}

// Note: MockChatModel is defined in test_utils.go
