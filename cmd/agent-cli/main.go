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
	"flag"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	
	"AgentFramework/agent"
)

// Simple tool registry for CLI
var toolRegistry = map[string]tool.BaseTool{
	// Add common tools here if needed, or implement a dynamic loader
	// For now, empty or basic tools
}

func main() {
	configPath := flag.String("config", "host.yaml", "Path to host configuration file")
	port := flag.String("port", ":8080", "Server listening address")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("Please provide a config file path using -config")
	}

	ctx := context.Background()

	// 1. Load Config
	log.Printf("Loading config from %s...", *configPath)
	hostCfg, err := agent.LoadHostConfigFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Setup Model Factory (Generic OpenAI compatible)
	// Expects env vars OPENAI_API_KEY and OPENAI_BASE_URL to be set if not using default
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	openaiModel := os.Getenv("OPENAI_MODEL")

	if apiKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set, using placeholder")
		apiKey = "sk-placeholder"
	}
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4/"
	}
	if openaiModel == "" {
		openaiModel = hostCfg.DefaultModel
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   openaiModel, // Use default model from config	
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	modelFactory := func(ctx context.Context, name string) (agent.ChatModel, error) {
		// In a real CLI, we might support multiple models based on name
		// Here we just return the default one
		return chatModel, nil
	}

	// 3. Initialize Host
	host, err := agent.NewHost(ctx, hostCfg, modelFactory, toolRegistry)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	// 4. Start Server
	server := agent.NewAgentRuntimeServer(host, *port, "")
	log.Printf("Starting Agent Runtime Server on %s", *port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
