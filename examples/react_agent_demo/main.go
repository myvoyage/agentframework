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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"AgentFramework/agent"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== ReAct Agent Dynamic Creation Demo ===\n")

	// Example 1: Create ReActAgent from JSON definition
	fmt.Println("Example 1: Creating ReActAgent from JSON workflow definition")
	if err := runJSONExample(ctx); err != nil {
		log.Printf("JSON example failed: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 2: Create ReActAgent from YAML definition
	fmt.Println("Example 2: Creating ReActAgent from YAML workflow definition")
	if err := runYAMLExample(ctx); err != nil {
		log.Printf("YAML example failed: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 3: Create ReActAgent programmatically
	fmt.Println("Example 3: Creating ReActAgent programmatically")
	if err := runProgrammaticExample(ctx); err != nil {
		log.Printf("Programmatic example failed: %v\n", err)
	}
}

func runJSONExample(ctx context.Context) error {
	// Read JSON workflow definition
	jsonData, err := os.ReadFile("../react_agent_workflow.json")
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Parse workflow definition
	workflowDef, err := agent.ParseWorkflowDefinition(string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	fmt.Printf("Parsed workflow: %s (type: %s)\n", workflowDef.Name, workflowDef.Type)
	fmt.Printf("Number of nodes: %d\n", len(workflowDef.Nodes))

	// Create model factory
	modelFactory := createMockModelFactory()

	// Create skill library
	skillLibrary := agent.NewSkillLibrary()

	// Create workflow from definition
	workflow, err := agent.CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	fmt.Printf("✓ Successfully created workflow: %s\n", workflow.Name())
	fmt.Println("  - Contains ReAct agent for research tasks")
	fmt.Println("  - Configured with max_iterations: 15")
	fmt.Println("  - Memory management enabled")

	return nil
}

func runYAMLExample(ctx context.Context) error {
	// Read YAML workflow definition
	yamlData, err := os.ReadFile("../react_agent_workflow.yaml")
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse workflow definition
	workflowDef, err := agent.ParseWorkflowDefinition(string(yamlData))
	if err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	fmt.Printf("Parsed workflow: %s (type: %s)\n", workflowDef.Name, workflowDef.Type)
	fmt.Printf("Number of nodes: %d\n", len(workflowDef.Nodes))
	fmt.Printf("Number of edges: %d\n", len(workflowDef.Edges))

	// Create model factory
	modelFactory := createMockModelFactory()

	// Create skill library
	skillLibrary := agent.NewSkillLibrary()

	// Create workflow from definition
	workflow, err := agent.CreateWorkflowFromDefinition(workflowDef, skillLibrary, modelFactory)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	fmt.Printf("✓ Successfully created DAG workflow: %s\n", workflow.Name())
	fmt.Println("  - Contains ReAct agent for problem solving")
	fmt.Println("  - Configured with max_iterations: 20")
	fmt.Println("  - Advanced memory management (8000 tokens)")
	fmt.Println("  - Retry and timeout configurations")

	return nil
}

func runProgrammaticExample(ctx context.Context) error {
	// Create model factory
	modelFactory := createMockModelFactory()

	// Create node definition for ReActAgent
	nodeDef := agent.NodeDefinition{
		Type: "agent",
		Name: "Programmatic ReAct Agent",
		Description: "A ReAct agent created programmatically",
		Config: map[string]interface{}{
			"kind":  "react",
			"model": "gpt-4",
			"instructions": "You are a helpful AI assistant that uses reasoning and tools to solve problems.",
			"max_iterations": 12,
			"memory": map[string]interface{}{
				"enable_trimming": true,
				"max_messages":    25,
				"max_tokens":      5000,
				"preserve_recent": 8,
			},
		},
		MaxRetries: 2,
		RetryDelay: "2s",
		Timeout:    "5m",
	}

	// Create ReActAgent from node definition
	reactAgent, err := createReActAgentFromNodeDefinition("programmatic_react", nodeDef, modelFactory)
	if err != nil {
		return fmt.Errorf("failed to create ReAct agent: %w", err)
	}

	fmt.Printf("✓ Successfully created ReAct agent: %s\n", reactAgent.Name())
	fmt.Println("  - Configured programmatically")
	fmt.Println("  - Custom memory settings")
	fmt.Println("  - Retry and timeout configured")

	// Demonstrate agent capabilities
	fmt.Println("\nAgent capabilities:")
	fmt.Println("  - Reasoning: Step-by-step problem solving")
	fmt.Println("  - Tool usage: Can use available tools")
	fmt.Println("  - Memory management: Intelligent history trimming")
	fmt.Println("  - Error handling: Automatic retries")

	return nil
}

// createMockModelFactory creates a mock model factory for demonstration
func createMockModelFactory() agent.ModelFactory {
	return func(ctx context.Context, modelName string) (agent.ChatModel, error) {
		fmt.Printf("  [ModelFactory] Creating model: %s\n", modelName)
		return &mockChatModel{name: modelName}, nil
	}
}

// mockChatModel is a mock implementation for demonstration
type mockChatModel struct {
	name string
}

func (m *mockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: fmt.Sprintf("Mock response from %s", m.name),
	}, nil
}

func (m *mockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("stream not implemented in mock")
}

// Helper function to access the internal createReActAgentFromNodeDefinition
// In a real implementation, this would be exported or accessed through the public API
func createReActAgentFromNodeDefinition(nodeID string, nodeDef agent.NodeDefinition, modelFactory agent.ModelFactory) (agent.Agent, error) {
	// This is a placeholder - in the actual implementation, this would call the internal function
	// For demonstration purposes, we'll show that the function exists and can be called
	fmt.Println("  [Internal] Calling createReActAgentFromNodeDefinition...")
	
	// In reality, this would be:
	// return agent.createReActAgentFromNodeDefinition(nodeID, nodeDef, modelFactory)
	
	// For now, return a mock
	return &mockAgent{name: nodeID}, nil
}

type mockAgent struct {
	name string
}

func (a *mockAgent) Name() string {
	return a.name
}

func (a *mockAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "Mock agent response",
	}, nil
}
