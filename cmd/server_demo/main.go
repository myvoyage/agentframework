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
	"log"
	"os"
	"strings"

	"AgentFramework/agent"
	"AgentFramework/agent/tools"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer() func(context.Context) error {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("agent-server-demo"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func main() {
	ctx := context.Background()

	// 1. Initialize OpenTelemetry
	shutdown := initTracer()
	defer shutdown(ctx)

	// 2. Create Host with Configuration
	// Initialize Tools
	// 2.1 Native Mock Tools
	searchTool, err := tools.NewWebSearchTool()
	if err != nil {
		log.Fatal(err)
	}

	// 2.2 MCP Tools (Filesystem)
	cwd, _ := os.Getwd()
	cmd := "npx"
	if os.PathSeparator == '\\' {
		cmd = "npx.cmd"
	}
	args := []string{"-y", "@modelcontextprotocol/server-filesystem", cwd}
	
	mcpClient, err := client.NewStdioMCPClient(cmd, nil, args...)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer mcpClient.Close()

	// Initialize MCP Client
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "agent-server", Version: "1.0.0"}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}
	
	_, err = mcpClient.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("Failed to initialize MCP client: %v", err)
	}

	// Convert MCP tools to Eino tools
	mcpTools, err := agent.NewMCPTools(ctx, mcpClient)
	if err != nil {
		log.Fatalf("Failed to get MCP tools: %v", err)
	}

	// Register all tools
	toolRegistry := map[string]tool.BaseTool{
		"web_search": searchTool,
	}

	var mcpToolNames []string
	for _, t := range mcpTools {
		info, _ := t.Info(ctx)
		toolRegistry[info.Name] = t
		mcpToolNames = append(mcpToolNames, info.Name)
		log.Printf("Registered MCP Tool: %s", info.Name)
	}

	// We will create a declarative DAG via config
	config := &agent.HostConfig{
		Name:    "server_demo",
		Version: "1.0",
		Agents: []agent.AgentSpec{
			{
				Name:         "writer",
				Kind:         "chat",
				Model:        "llama3", 
				Instructions: "You are a creative writer.",
			},
			{
				Name:         "reviewer",
				Kind:         "chat",
				Model:        "llama3",
				Instructions: "You are a critical reviewer.",
			},
			{
				Name:         "researcher",
				Kind:         "react",
				Model:        "llama3",
				Instructions: "You are a researcher. Use web_search to find information. Use filesystem tools (list_directory, read_file) to check local files.",
				Tools:        append([]string{"web_search"}, mcpToolNames...),
			},
			{
				Name: "approver",
				Kind: "human",
			},
			{
				Name:         "analyst",
				Kind:         "react",
				Model:        "llama3",
					Instructions: "You are a data analyst. Analyze the data provided.",
					Tools:        mcpToolNames,
			},
		},
		Workflows: []agent.WorkflowSpec{
			{
				Name: "write_and_review",
				Kind: "dag",
				Nodes: []agent.NodeSpec{
					{ID: "writer", AgentName: "writer"},
					{ID: "reviewer", AgentName: "reviewer"},
				},
				Edges: map[string]string{
					"writer": "reviewer",
				},
			},
			{
				Name: "write_review_approve",
				Kind: "dag",
				Nodes: []agent.NodeSpec{
					{ID: "writer", AgentName: "writer"},
					{ID: "approver", AgentName: "approver"},
					{ID: "publisher", AgentName: "writer"}, // Reuse writer as publisher for demo
				},
				Edges: map[string]string{
					"writer":   "approver",
					"approver": "publisher",
				},
			},
		},
	}

	// Create a model factory that supports multiple model types
	modelFactory := agent.NewDefaultModelFactory(agent.DefaultModelFactoryConfig{
		Models: map[string]agent.ModelConfig{
			"llama3": {
				Type:  "ollama",
				Model: "llama3:latest",
			},
			"mistral": {
				Type:  "ollama",
				Model: "mistral:latest",
			},
			"vllm-llama": {
				Type:    "vllm",
				Model:   "llama3:latest",
				BaseURL: "http://localhost:8000",
			},
		},
	})

	host, err := agent.NewHost(ctx, config, modelFactory, toolRegistry)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	// 2.3 Construct and Register Cyclic Graph Workflow (Critique-Refine)
	cyclicWf := agent.NewGraphWorkflow("critique_refine_loop")

	writer, _ := host.Agent("writer")
	reviewer, _ := host.Agent("reviewer")

	// Wrap agents as sequential workflows for use in graph workflow
	writerWf := agent.NewSequentialWorkflow("writer_node", writer)
	reviewerWf := agent.NewSequentialWorkflow("reviewer_node", reviewer)

	cyclicWf.AddNode("writer", writerWf)
	cyclicWf.AddNode("reviewer", reviewerWf)
	cyclicWf.SetStartNode("writer")
	
	cyclicWf.AddEdge("writer", "reviewer")
	
	cyclicWf.AddConditionalEdge("reviewer", func(ctx context.Context, input string) (string, error) {
		if strings.Contains(strings.ToUpper(input), "APPROVED") {
			return agent.NodeEnd, nil
		}
		return "writer", nil
	})
	
	host.AddWorkflow(cyclicWf)
	log.Printf("Registered Cyclic Workflow: %s", cyclicWf.Name())

	// TODO: Semantic Router not implemented yet
	// 2.4 Construct and Register Semantic Router
	// embedder := router.NewOllamaEmbedding("http://localhost:11434", "nomic-embed-text")
	// semRouter := router.NewSemanticRouter(embedder)

	// analystAgent, _ := host.Agent("analyst")
	// host.AddWorkflow(agent.NewSequentialWorkflow("analyst_workflow", analystAgent))
	//
	// researcherAgent, _ := host.Agent("researcher")
	// host.AddWorkflow(agent.NewSequentialWorkflow("researcher_workflow", researcherAgent))

	// Add routes. Note: AddRoute makes a network call to embed, so it might fail if Ollama is not up.
	// We wrap in a simple check or ignore error for demo resilience.
	// if err := semRouter.AddRoute(ctx, "write a blog post about AI", "write_and_review"); err != nil {
	// 	log.Printf("Warning: Failed to add route: %v", err)
	// }
	// _ = semRouter.AddRoute(ctx, "research quantum computing", "researcher_workflow")
	// _ = semRouter.AddRoute(ctx, "analyze data and plot chart", "analyst_workflow")
	// _ = semRouter.AddRoute(ctx, "find information on the web", "researcher_workflow")
	// _ = semRouter.AddRoute(ctx, "critique and refine text", "critique_refine_loop")

	// semRoutingWf := &SemanticRoutingWorkflow{
	// 	NameStr: "semantic_router",
	// 	Router:  semRouter,
	// 	Host:    host,
	// }
	// host.AddWorkflow(semRoutingWf)
	// log.Printf("Registered Semantic Router Workflow: %s", semRoutingWf.Name())
	_ = ctx // suppress unused warning

	// 3. Start Server
	server := agent.NewAgentRuntimeServer(host, ":8081", "")
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// TODO: Semantic routing workflow requires router package implementation
/*
type SemanticRoutingWorkflow struct {
	NameStr string
	Router  *router.SemanticRouter
	Host    *agent.Host
}

func (w *SemanticRoutingWorkflow) Name() string {
	return w.NameStr
}

func (w *SemanticRoutingWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	targetName, err := w.Router.Route(ctx, input)
	if err != nil {
		return nil, err
	}

	log.Printf("[SemanticRouter] Routing input %q to %q", input, targetName)

	wf, ok := w.Host.Workflow(targetName)
	if !ok {
		ag, ok := w.Host.Agent(targetName)
		if ok {
			return ag.Run(ctx, input, opts...)
		}
		return nil, fmt.Errorf("target %q not found", targetName)
	}

	return wf.Run(ctx, input, opts...)
}
*/
// TODO: Resume method for semantic routing workflow (requires router implementation)
/*
func (w *SemanticRoutingWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for semantic routing workflow")
}
*/
