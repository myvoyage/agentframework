// Agent Framework - Agent Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmdcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"AgentFramework/core"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "运行 AI 代理",
	Long:  `运行各种类型的 AI 代理，包括 ChatAgent、ReActAgent 等。`,
}

// addAgentCommands adds agent-related commands to root command
func addAgentCommands() {
	// Chat agent
	chatCmd := &cobra.Command{
		Use:   "chat [message]",
		Short: "运行对话代理",
		Long:  `运行一个对话式 AI 代理。可以指定消息直接进行交互。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewAgentRunnerService(app)

			input := ""
			if len(args) > 0 {
				input = args[0]
			}

			if input == "" {
				return fmt.Errorf("please provide a message to chat")
			}

			response, err := svc.Chat(ctx, modelName, input)
			if err != nil {
				return fmt.Errorf("chat failed: %w", err)
			}

			fmt.Println(response)
			return nil
		},
	}
	agentCmd.AddCommand(chatCmd)

	rootCmd.AddCommand(agentCmd)
}
