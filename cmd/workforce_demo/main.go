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
	"AgentFramework/agent/tools"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// Create a model instance for task splitting and aggregation
	model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		Model:   "llama3:latest",
		BaseURL: "http://localhost:11434",
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create some mock tools
	searchTool, err := tools.NewWebSearchTool()
	if err != nil {
		log.Fatalf("Failed to create search tool: %v", err)
	}

	// Demo 1: Basic WorkerAgent and Workforce
	fmt.Println("=== Demo 1: Basic WorkerAgent and Workforce ===")
	
	// Create a workforce
	workforce := agent.NewSimpleWorkforce("demo_workforce")

	// Create different types of WorkerAgents
	researcher, err := agent.NewResearcherAgent(
		"researcher",
		model,
		"You are a researcher. Use web_search to find information.",
		[]tool.BaseTool{searchTool},
	)
	if err != nil {
		log.Fatalf("Failed to create researcher: %v", err)
	}

	writer, err := agent.NewWriterAgent(
		"writer",
		model,
		"You are a creative writer.",
		[]tool.BaseTool{},
	)
	if err != nil {
		log.Fatalf("Failed to create writer: %v", err)
	}

	// Add agents to workforce
	workforce.AddWorker(researcher)
	workforce.AddWorker(writer)

	// Test assigning tasks by role
	fmt.Println("\nAssigning research task to researcher role...")
	researchResult, err := workforce.AssignTaskByRole(ctx, agent.RoleResearcher, "Research the latest developments in AI agents")
	if err != nil {
		log.Fatalf("Failed to assign research task: %v", err)
	}
	fmt.Printf("Research Result: %s\n", researchResult.Content)

	// Demo 2: ParallelWorkforce
	fmt.Println("\n=== Demo 2: ParallelWorkforce ===")
	
	parallelWorkforce := agent.NewParallelWorkforce("parallel_workforce")
	parallelWorkforce.AddWorker(researcher)
	parallelWorkforce.AddWorker(writer)

	// Create parallel tasks
	parallelTasks := []agent.ParallelWorkforceTask{
		{
			Role: agent.RoleResearcher,
			Task: "Research the history of AI",
		},
		{
			Role: agent.RoleWriter,
			Task: "Write a poem about AI",
		},
	}

	// Run parallel tasks
	parallelResults, err := parallelWorkforce.RunParallel(ctx, parallelTasks)
	if err != nil {
		log.Fatalf("Failed to run parallel tasks: %v", err)
	}

	// Print parallel results
	for i, result := range parallelResults {
		if result.Error != "" {
			fmt.Printf("Task %d Error: %s\n", i+1, result.Error)
		} else {
			fmt.Printf("Task %d Result: %s\n", i+1, result.Result.Content)
		}
	}

	// Demo 3: TaskCoordinator
	fmt.Println("\n=== Demo 3: TaskCoordinator ===")
	
	// Create task splitter and aggregator
	taskSplitter := agent.NewSimpleTaskSplitter(model)
	resultAggregator := agent.NewSimpleResultAggregator(model)

	// Create task coordinator
	coordinator := agent.NewTaskCoordinator(
		parallelWorkforce,
		taskSplitter,
		resultAggregator,
		model,
	)

	// Coordinate a complex task
	complexTask := "Write a comprehensive report about AI agents, including their history, latest developments, and future trends. The report should be well-structured and include examples."
	fmt.Printf("\nCoordinating complex task: %s\n", complexTask)
	coordinatedResult, err := coordinator.Coordinate(ctx, complexTask)
	if err != nil {
		log.Fatalf("Failed to coordinate task: %v", err)
	}
	fmt.Printf("Coordinated Result: %s\n", coordinatedResult.Content)

	// Demo 4: GraphWorkforce
	fmt.Println("\n=== Demo 4: GraphWorkforce ===")
	
	// Create a graph workforce
	graphWorkforce := agent.NewGraphWorkforce("graph_workforce")

	// Add workers to graph workforce
	graphWorkforce.AddWorker(researcher)
	graphWorkforce.AddWorker(writer)

	// Create a reviewer agent
	reviewer, err := agent.NewReviewerAgent(
		"reviewer",
		model,
		"You are a critical reviewer. Review the content and provide feedback.",
		[]tool.BaseTool{},
	)
	if err != nil {
		log.Fatalf("Failed to create reviewer: %v", err)
	}
	graphWorkforce.AddWorker(reviewer)

	// Set up the workflow graph
	graphWorkforce.SetStartWorker("researcher")
	graphWorkforce.AddEdge("researcher", "writer")
	graphWorkforce.AddEdge("writer", "reviewer")

	// Add a conditional edge from reviewer
	graphWorkforce.AddConditionalEdge("reviewer", func(ctx context.Context, output *schema.Message) (string, error) {
		if output != nil && len(output.Content) > 0 && containsKeyword(output.Content, "approve") {
			return "", nil
		}
		return "writer", nil
	})

	// Run the graph workforce
	graphResult, err := graphWorkforce.Run(ctx, "Research and write a short article about AI agent frameworks")
	if err != nil {
		log.Fatalf("Failed to run graph workforce: %v", err)
	}
	fmt.Printf("Graph Workforce Result: %s\n", graphResult.Content)

	// Demo 5: WorkforceDAGWorkflow
	fmt.Println("\n=== Demo 5: WorkforceDAGWorkflow ===")
	
	// Create a workforce DAG workflow
	workforceDAG := agent.NewWorkforceDAGWorkflow("workforce_dag")

	// Add workforce nodes to the DAG
	workforceDAG.AddWorkforceNode("research", workforce, agent.RoleResearcher, "")
	workforceDAG.AddWorkforceNode("write", workforce, agent.RoleWriter, "")
	workforceDAG.AddWorkforceNode("review", workforce, agent.RoleReviewer, "")

	// Add edges to create the workflow
	workforceDAG.AddEdge("research", "write")
	workforceDAG.AddEdge("write", "review")

	// Run the workforce DAG
	dagResult, err := workforceDAG.Run(ctx, "Create a comprehensive guide to AI agents")
	if err != nil {
		log.Fatalf("Failed to run workforce DAG: %v", err)
	}
	fmt.Printf("Workforce DAG Result: %s\n", dagResult.Content)

	fmt.Println("\nAll demos completed successfully!")
}

// containsKeyword checks if a string contains a keyword (case-insensitive)
func containsKeyword(s, keyword string) bool {
	return len(s) >= len(keyword) && 
			containsSubstringIgnoreCase(s, keyword)
}

// containsSubstringIgnoreCase checks if a string contains a substring (case-insensitive)
func containsSubstringIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a character to lowercase
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
