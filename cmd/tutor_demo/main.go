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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"AgentFramework/agent"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
)

func main() {
	ctx := context.Background()

	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		Model: "llama3:latest",
	})
	if err != nil {
		log.Fatalf("failed to create model: %v", err)
	}

	baseAgent, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "tutor",
		Instructions: "You are a patient tutor. Explain concepts in simple Chinese and use the provided context when available.",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("failed to create tutor agent: %v", err)
	}

	client := agent.NewMockGraphlitClient()
	ragAgent := agent.NewRAGMiddleware(client)(baseAgent)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Tutor demo started. Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("read error: %v", err)
		}

		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if text == "exit" {
			break
		}

		resp, err := ragAgent.Run(ctx, text, model.WithTemperature(0.2))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		fmt.Println(resp.Content)
	}
}

