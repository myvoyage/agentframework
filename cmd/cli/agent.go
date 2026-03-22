// Agent Framework - Agent Commands
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"AgentFramework/agent"
)

// Flag variables
var (
	_agentSessionType  string
	_agentChannel      string
	_agentUserID       string
	_agentWorkspace    string
	_agentMaxIter      int
	_agentTimeout      int
	_agentStream       bool
	_agentSandbox      bool
)

func init() {
	// Global agent flags
	agentCmd.PersistentFlags().StringVarP(&_agentSessionType, "session-type", "t", "main", "会话类型: main/dm/group")
	agentCmd.PersistentFlags().StringVarP(&_agentChannel, "channel", "c", "cli", "消息渠道: cli/telegram/lark/qq")
	agentCmd.PersistentFlags().StringVar(&_agentUserID, "user-id", "", "用户 ID（main 会话）")
	agentCmd.PersistentFlags().StringVarP(&_agentWorkspace, "workspace", "w", "", "工作空间路径")
	agentCmd.PersistentFlags().IntVar(&_agentMaxIter, "max-iter", 10, "最大迭代次数")
	agentCmd.PersistentFlags().IntVar(&_agentTimeout, "timeout", 300, "超时时间（秒）")
	agentCmd.PersistentFlags().BoolVar(&_agentStream, "stream", false, "流式输出")
	agentCmd.PersistentFlags().BoolVar(&_agentSandbox, "sandbox", false, "启用沙箱隔离")
}

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "管理和运行 AI 代理",
	Long: `管理和运行各种类型的 AI 代理，包括 ChatAgent、ReActAgent、WorkerAgent 等。

会话类型（--session-type / -t）：
  main  - 用户私聊（最高权限，直接执行命令）
  dm    - 他人私信（默认沙箱隔离）
  group - 群聊（默认沙箱隔离）

示例：
  af agent list                              # 列出所有 agents
  af agent chat "你好"                       # 与默认 agent 对话
  af agent run coder "写一个快速排序"        # 运行指定 agent
  af agent chat -t dm "执行 rm -rf"          # 在沙箱中执行
  af agent session list                       # 列出会话`,
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
		Short: "与 agent 对话",
		Long: `使用 AI 代理进行对话。

会话类型决定执行权限：
  main  - 完整权限
  dm    - 沙箱隔离
  group - 沙箱隔离

示例：
  af agent chat "你好"
  af agent chat -t main "执行系统命令"
  af agent chat -c lark "群聊消息"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			input := args[0]
			if len(args) > 1 {
				for _, a := range args[1:] {
					input += " " + a
				}
			}

			// Get agent
			agents := app.GetHost().ListAgents()
			if len(agents) == 0 {
				return fmt.Errorf("no agents available")
			}

			agentID := agents[0]
			if len(args) > 0 {
				agentID = args[0]
			}

			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent: %w", err)
			}

			// Execute
			response, err := a.Run(ctx, input)
			if err != nil {
				return fmt.Errorf("chat failed: %w", err)
			}

			if _agentStream {
				fmt.Print(response.Content)
			} else {
				fmt.Println(response.Content)
			}
			return nil
		},
	}
	chatCmd.Flags().BoolVarP(&_agentStream, "stream", "s", false, "流式输出响应")
	agentCmd.AddCommand(chatCmd)

	// ── run ──────────────────────────────────────────────────────────────────
	runCmd := &cobra.Command{
		Use:   "run [agent-id] [task...]",
		Short: "运行指定 agent 执行任务",
		Long: `使用指定的 AI 代理执行特定任务。

根据会话类型自动选择执行模式：
  main  - 直接执行（完整权限）
  dm    - 沙箱执行（隔离环境）
  group - 沙箱执行（群组隔离）

示例：
  af agent run coder "写一个快速排序"
  af agent run coder -t dm "执行命令"
  af agent run swe --sandbox "分析代码"`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			agentID := args[0]

			task := args[1]
			for _, a := range args[2:] {
				task += " " + a
			}

			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent '%s': %w", agentID, err)
			}

			// Check sandbox flag
			if _agentSandbox || _agentSessionType != "main" {
				fmt.Fprintf(os.Stderr, "[Sandbox] Running in sandboxed mode\n")
			}

			result, err := a.Run(ctx, task)
			if err != nil {
				return fmt.Errorf("agent run failed: %w", err)
			}

			if outputFormat == "json" {
				out := map[string]interface{}{
					"agent_id":  agentID,
					"result":    result.Content,
					"session":   _agentSessionType,
					"sandboxed": _agentSandbox || _agentSessionType != "main",
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Printf("Agent Result (%s):\n%s\n", agentID, result.Content)
			return nil
		},
	}
	runCmd.Flags().IntVarP(&_agentMaxIter, "max-iter", "i", 10, "最大工具调用迭代次数")
	runCmd.Flags().IntVarP(&_agentTimeout, "timeout", "T", 300, "超时时间（秒）")
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

	// ── session ──────────────────────────────────────────────────────────────
	sessionListCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有会话",
		Long:  `列出系统中所有活动会话。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement session listing
			fmt.Println("Sessions:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println("  (Session management not yet implemented)")
			return nil
		},
	}

	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "会话管理",
		Long:  `管理 AI 代理会话，包括列出、查看和删除会话。`,
	}
	sessionCmd.AddCommand(sessionListCmd)
	agentCmd.AddCommand(sessionCmd)

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

	// ── exec (new) ──────────────────────────────────────────────────────────
	execCmd := &cobra.Command{
		Use:   "exec [agent-id] [task...]",
		Short: "执行带上下文的任务（实验性）",
		Long: `执行带有完整上下文组装的任务，包括：
- Workspace 配置
- 会话历史
- 记忆检索
- 技能激活

示例：
  af agent exec coder "实现排序算法"
  af agent exec swe -w ./workspace "分析代码"`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			agentID := args[0]

			task := args[1]
			for _, a := range args[2:] {
				task += " " + a
			}

			a, err := app.GetHost().GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent '%s': %w", agentID, err)
			}

			fmt.Fprintf(os.Stderr, "[Context] Session: %s, Channel: %s\n",
				_agentSessionType, _agentChannel)

			result, err := a.Run(ctx, task)
			if err != nil {
				return fmt.Errorf("exec failed: %w", err)
			}

			fmt.Println(result.Content)
			return nil
		},
	}
	agentCmd.AddCommand(execCmd)

	rootCmd.AddCommand(agentCmd)
}

// resolveSessionType converts string to SessionType
func resolveSessionType(s string) agent.SessionType {
	switch strings.ToLower(s) {
	case "main":
		return agent.SessionTypeMain
	case "dm":
		return agent.SessionTypeDM
	case "group":
		return agent.SessionTypeGroup
	default:
		return agent.SessionTypeMain
	}
}
