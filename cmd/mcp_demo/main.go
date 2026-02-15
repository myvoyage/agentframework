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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"

	"AgentFramework/agent"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ctx := context.Background()
	fmt.Println("Starting MCP + HITL Demo...")

	// ==========================================
	// 1. MCP Integration (using eino-ext + mark3labs)
	// ==========================================

	// 1.1 Create an In-Process MCP Server
	s := server.NewMCPServer("demo-server", "1.0.0")

	// Define a tool
	weatherTool := mcp.NewTool("get_weather",
		mcp.WithDescription("Get weather for a location"),
		mcp.WithString("location", mcp.Required(), mcp.Description("City name")),
	)

	// Add tool handler
	s.AddTool(weatherTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		loc := req.GetString("location", "Unknown")
		return mcp.NewToolResultText(fmt.Sprintf("Weather in %s is Sunny, 25C", loc)), nil
	})

	// 1.2 Create an In-Process Client
	// Note: NewInProcessClient returns (*client.Client, error)
	mcpClient, err := client.NewInProcessClient(s)
	if err != nil {
		log.Fatalf("Failed to create in-process client: %v", err)
	}
	defer mcpClient.Close()

	// 1.2.1 Initialize Client
	fmt.Println("Initializing Client...")
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "demo-client",
		Version: "1.0.0",
	}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	_, err = mcpClient.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// 1.3 Adapt to Eino Tools
	fmt.Println("Discovering MCP Tools...")
	tools, err := agent.NewMCPTools(ctx, mcpClient)
	if err != nil {
		log.Fatalf("Failed to get MCP tools: %v", err)
	}

	for _, t := range tools {
		info, _ := t.Info(ctx)
		fmt.Printf("- Discovered Tool: %s (%s)\n", info.Name, info.Desc)
	}

	// 1.4 Invoke Tool (Verify Eino Integration)
	if len(tools) > 0 {
		fmt.Println("Invoking Tool 'get_weather' via Eino interface...")
		invokable, ok := tools[0].(tool.InvokableTool)
		if !ok {
			log.Fatalf("Tool is not InvokableTool")
		}

		// Eino passes JSON string as argument
		res, err := invokable.InvokableRun(ctx, `{"location": "Shanghai"}`)
		if err != nil {
			log.Fatalf("Tool execution failed: %v", err)
		}
		fmt.Printf("Tool Output: %s\n", res)
	}

	// ==========================================
	// 2. Workflow Checkpointing & HITL
	// ==========================================
	fmt.Println("\nVerifying HITL (Human-in-the-Loop)...")

	dag := agent.NewDAGWorkflow("hitl_demo")
	dag.AddNode("AgentA", &MockAgent{NameVal: "AgentA"})
	dag.AddNode("HumanReview", agent.NewHumanNode("HumanReview", "Please approve the output."))
	dag.AddNode("AgentB", &MockAgent{NameVal: "AgentB"})

	dag.AddEdge("AgentA", "HumanReview")
	dag.AddEdge("HumanReview", "AgentB")

	// Run 1: Should suspend at HumanReview
	fmt.Println(">> Run 1: Starting Workflow...")
	resp, state, err := dag.RunResumable(ctx, "Start Task", nil)

	if err == agent.ErrSuspended {
		fmt.Println(">> Workflow Suspended! (Expected)")
		fmt.Printf("Current State: %v\n", state.NodeStates)

		// Simulate User Input "Approved"
		fmt.Println(">> Simulating User Approval...")
		state.NodeStates["HumanReview"] = "APPROVED by User"

		// Run 2: Resume
		fmt.Println(">> Run 2: Resuming Workflow...")
		resp, state, err = dag.RunResumable(ctx, "Start Task", state)
		if err != nil {
			log.Fatalf("Resume failed: %v", err)
		}
		fmt.Printf(">> Workflow Completed. Final Output: %s\n", resp.Content)

	} else if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	} else {
		log.Fatalf("Workflow finished without suspension? Output: %s", resp.Content)
	}
}

// MockAgent implementation
type MockAgent struct {
	NameVal string
}

func (m *MockAgent) Name() string {
	return m.NameVal
}

func (m *MockAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "Output from " + m.NameVal,
	}, nil
}
