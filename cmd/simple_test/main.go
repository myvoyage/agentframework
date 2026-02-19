// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"context"
	"fmt"
	"log"

	"AgentFramework/agent"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel is a mock ChatModel implementation for testing
type MockChatModel struct{}

func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: fmt.Sprintf("Mock response to: %s", input[len(input)-1].Content),
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("stream not supported")
}

func main() {
	ctx := context.Background()

	// Create mock model
	mockModel := &MockChatModel{}

	// Create simple agent
	agentConfig := agent.ChatAgentConfig{
		Name:         "test_agent",
		Instructions: "You are a helpful assistant.",
		Model:        mockModel,
	}

	chatAgent, err := agent.NewChatAgent(ctx, agentConfig)
	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	// Test basic agent functionality
	resp, err := chatAgent.Run(ctx, "Hello, world!")
	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}

	fmt.Println("Basic agent test passed!")
	fmt.Printf("Agent response: %s\n", resp.Content)

	// Test sequential workflow
	seqWf := agent.NewSequentialWorkflow("test_seq_wf", chatAgent, chatAgent)

	seqResp, err := seqWf.Run(ctx, "Hello from sequential workflow!")
	if err != nil {
		log.Fatalf("Failed to run sequential workflow: %v", err)
	}

	fmt.Println("Sequential workflow test passed!")
	fmt.Printf("Sequential workflow response: %s\n", seqResp.Content)

	// Test DAG workflow
	dagWf := agent.NewDAGWorkflow("test_dag_wf")
	dagWf.AddNode("node1", chatAgent)
	dagWf.AddNode("node2", chatAgent)
	dagWf.AddEdge("node1", "node2")

	dagResp, err := dagWf.Run(ctx, "Hello from DAG workflow!")
	if err != nil {
		log.Fatalf("Failed to run DAG workflow: %v", err)
	}

	fmt.Println("DAG workflow test passed!")
	fmt.Printf("DAG workflow response: %s\n", dagResp.Content)

	fmt.Println("All tests passed! AgentFramework is working correctly.")
}