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

	"AgentFramework/agent"
	"AgentFramework/agent/tools"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/tool"
)

func main() {
	ctx := context.Background()

	searchTool, err := tools.NewWebSearchTool()
	if err != nil {
		log.Fatal(err)
	}

	toolRegistry := map[string]tool.BaseTool{
		"web_search": searchTool,
	}

	config := &agent.HostConfig{
		Name:    "support_demo",
		Version: "1.0",
		Agents: []agent.AgentSpec{
			{
				Name:         "support_bot",
				Kind:         "react",
				Model:        "llama3",
				Instructions: "You are a helpful customer support assistant. Answer in Chinese and use web_search when you need external information.",
				Tools:        []string{"web_search"},
			},
			{
				Name: "approver",
				Kind: "human",
			},
		},
		Workflows: []agent.WorkflowSpec{
			{
				Name: "support_with_approval",
				Kind: "dag",
				Nodes: []agent.NodeSpec{
					{ID: "support", AgentName: "support_bot"},
					{ID: "approver", AgentName: "approver"},
				},
				Edges: map[string]string{
					"support": "approver",
				},
			},
		},
	}

	modelFactory := func(ctx context.Context, modelName string) (agent.ChatModel, error) {
		return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			Model: "llama3:latest",
		})
	}

	host, err := agent.NewHost(ctx, config, modelFactory, toolRegistry)
	if err != nil {
		log.Fatalf("failed to create host: %v", err)
	}

	server := agent.NewAgentRuntimeServer(host, ":8082", "")
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
