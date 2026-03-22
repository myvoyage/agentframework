// Agent Framework - Agent Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "管理和运行 AI 代理",
	Long:  `管理和运行各种类型的 AI 代理，包括 ChatAgent、ReActAgent、WorkerAgent 等。`,
}

// addAgentCommands adds agent-related commands to root command
func addAgentCommands() {
	// ── list ──────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有可用的 agents",
		Long:  `列出系统中所有已注册的 AI 代理及其基本信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			agents := app.GetHost().ListAgents()

			if len(agents) == 0 {
				fmt.Println("No agents available")
				return nil
			}

			if outputFormat == "json" {
				type agentJSON struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
				}
				list := make([]agentJSON, 0, len(agents))
				for _, agentID := range agents {
					a, err := app.GetHost().GetAgent(agentID)
					if err == nil && a != nil {
						list = append(list, agentJSON{ID: agentID, Name: a.Name(), Type: fmt.Sprintf("%T", a)})
					} else {
						list = append(list, agentJSON{ID: agentID, Name: agentID, Type: "Unknown"})
					}
				}
				b, _ := json.MarshalIndent(list, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Available Agents:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, agentID := range agents {
				a, err := app.GetHost().GetAgent(agentID)
				if err == nil && a != nil {
					fmt.Printf("  %-20s  %s\n", agentID, a.Name())
				} else {
					fmt.Printf("  %-20s  (unknown)\n", agentID)
				}
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Total: %d agent(s)\n", len(agents))
			return nil
		},
	}
	agentCmd.AddCommand(listCmd)

	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info [agent-id]",
		Short: "查看 agent 详细信息",
		Long:  `显示指定 agent 的详细信息，包括类型、名称和状态。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			agentID := args[0]

			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("agent '%s' not found: %w", agentID, err)
			}

			if outputFormat == "json" {
				info := map[string]interface{}{
					"id":   agentID,
					"name": a.Name(),
					"type": fmt.Sprintf("%T", a),
				}
				_ = ctx
				b, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Agent Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("ID:   %s\n", agentID)
			fmt.Printf("Name: %s\n", a.Name())
			fmt.Printf("Type: %T\n", a)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	agentCmd.AddCommand(infoCmd)

	// ── chat ─────────────────────────────────────────────────────────────────
	chatCmd := &cobra.Command{
		Use:   "chat [message]",
		Short: "与默认 agent 对话",
		Long:  `使用第一个可用的 AI 代理进行对话。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			input := args[0]
			if len(args) > 1 {
				// 拼接多个参数
				for _, a := range args[1:] {
					input += " " + a
				}
			}

			agents := app.GetHost().ListAgents()
			if len(agents) == 0 {
				return fmt.Errorf("no agents available")
			}

			a, err := app.GetHost().GetAgent(agents[0])
			if err != nil {
				return fmt.Errorf("failed to get agent: %w", err)
			}

			response, err := a.Run(ctx, input)
			if err != nil {
				return fmt.Errorf("chat failed: %w", err)
			}

			fmt.Println(response.Content)
			return nil
		},
	}
	agentCmd.AddCommand(chatCmd)

	// ── run ──────────────────────────────────────────────────────────────────
	runCmd := &cobra.Command{
		Use:   "run [agent-id] [task...]",
		Short: "运行指定 agent 执行任务",
		Long:  `使用指定的 AI 代理执行特定任务，支持将剩余参数拼接为完整任务描述。`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			agentID := args[0]

			// 拼接任务内容
			task := args[1]
			for _, a := range args[2:] {
				task += " " + a
			}

			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent '%s': %w", agentID, err)
			}

			result, err := a.Run(ctx, task)
			if err != nil {
				return fmt.Errorf("agent run failed: %w", err)
			}

			if outputFormat == "json" {
				out := map[string]string{"agent_id": agentID, "result": result.Content}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Printf("Agent Result:\n%s\n", result.Content)
			return nil
		},
	}
	agentCmd.AddCommand(runCmd)

	// ── describe ─────────────────────────────────────────────────────────────
	describeCmd := &cobra.Command{
		Use:     "describe [agent-id]",
		Aliases: []string{"desc"},
		Short:   "描述 agent 的配置和能力",
		Long:    `显示指定 agent 的详细配置，包括模型、工具、中间件等信息。`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("agent '%s' not found: %w", agentID, err)
			}

			cfg := app.GetHost().Config()

			fmt.Printf("Agent: %s\n", agentID)
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Name:     %s\n", a.Name())
			fmt.Printf("Type:     %T\n", a)

			// 从配置中查找 AgentSpec
			for _, spec := range cfg.Agents {
				if spec.Name == agentID {
					fmt.Printf("Kind:     %s\n", spec.Kind)
					fmt.Printf("Model:    %s\n", spec.Model)
					if spec.Instructions != "" {
						fmt.Printf("Instructions: %s\n", spec.Instructions)
					}
					if len(spec.Tools) > 0 {
						fmt.Printf("Tools:    %v\n", spec.Tools)
					}
					if len(spec.Middlewares) > 0 {
						fmt.Printf("Middlewares: %v\n", spec.Middlewares)
					}
					break
				}
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	agentCmd.AddCommand(describeCmd)

	// ── workflows ────────────────────────────────────────────────────────────
	agentWorkflowsCmd := &cobra.Command{
		Use:   "workflows [agent-id]",
		Short: "列出与指定 agent 关联的工作流",
		Long:  `列出系统中所有工作流，以及哪些工作流使用了指定的 agent。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := app.GetHost()
			workflows := host.ListWorkflows()

			filterAgent := ""
			if len(args) > 0 {
				filterAgent = args[0]
			}

			if len(workflows) == 0 {
				fmt.Println("No workflows found")
				return nil
			}

			fmt.Println("Workflows:")
			fmt.Println("────────────────────────────────────────────────────────────")
			cfg := host.Config()
			for _, wfID := range workflows {
				for _, spec := range cfg.Workflows {
					if spec.Name == wfID {
						if filterAgent == "" {
							fmt.Printf("  %-25s  kind=%-15s  steps=%v\n", wfID, spec.Kind, spec.Steps)
						} else {
							// 检查工作流是否包含该 agent
							contains := false
							for _, step := range spec.Steps {
								if step == filterAgent {
									contains = true
									break
								}
							}
							for _, a := range spec.Agents {
								if a == filterAgent {
									contains = true
									break
								}
							}
							if contains {
								fmt.Printf("  %-25s  kind=%-15s\n", wfID, spec.Kind)
							}
						}
						break
					}
				}
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	agentCmd.AddCommand(agentWorkflowsCmd)

	rootCmd.AddCommand(agentCmd)
}
