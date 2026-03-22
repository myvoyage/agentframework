// Agent Framework - Host Commands (Host Instance Management)
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// hostCmd represents the host command
var hostCmd = &cobra.Command{
	Use:     "host",
	Aliases: []string{"instance", "app"},
	Short:   "Host 实例信息与管理",
	Long: `查看和管理 AgentFramework Host 实例的信息和状态。
Host 是框架的核心容器，管理所有 Agents、Workflows、Plugins 和各种子系统。

示例:
  af host info            # 查看 Host 实例概览
  af host config          # 查看完整配置
  af host models          # 列出所有配置的模型
  af host summary         # 查看系统能力摘要
  af host threads         # 查看线程存储状态`,
}

// addHostCommands adds host instance commands to root command
func addHostCommands() {
	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "查看 Host 实例概览",
		Long:  `显示 Host 实例的概览信息，包括 agents、workflows、插件和子系统状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			h := app.GetHost()
			cfg := h.Config()

			agentList := h.ListAgents()
			workflowList := h.ListWorkflows()
			pluginList := h.PluginManager().GetAllPlugins()

			if outputFormat == "json" {
				info := map[string]interface{}{
					"name":           cfg.Name,
					"version":        cfg.Version,
					"default_model":  cfg.DefaultModel,
					"agents":         len(agentList),
					"workflows":      len(workflowList),
					"plugins":        len(pluginList),
					"scheduler":      h.Scheduler() != nil,
					"heartbeat":      h.Heartbeat() != nil,
					"task_manager":   h.TaskManager() != nil,
					"token_compress": h.TokenCompressor() != nil,
					"channel_mgr":    h.ChannelManager() != nil,
				}
				b, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Host Instance Overview:")
			fmt.Println("════════════════════════════════════════════════════════════")
			fmt.Printf("Name:             %s\n", orDefault(cfg.Name, "(unnamed)"))
			fmt.Printf("Version:          %s\n", orDefault(cfg.Version, "(unset)"))
			fmt.Printf("Default Model:    %s\n", orDefault(cfg.DefaultModel, "(unset)"))
			fmt.Println()

			// Agents & Workflows
			fmt.Printf("Agents:           %d registered\n", len(agentList))
			for _, id := range agentList {
				a, err := h.GetAgent(id)
				if err == nil {
					fmt.Printf("  ├─ %-18s  (%T)\n", id, a)
				} else {
					fmt.Printf("  ├─ %s\n", id)
				}
			}

			fmt.Printf("Workflows:        %d registered\n", len(workflowList))
			for _, name := range workflowList {
				fmt.Printf("  ├─ %s\n", name)
			}

			fmt.Printf("Plugins:          %d loaded\n", len(pluginList))
			for _, p := range pluginList {
				status := "disabled"
				if p.IsEnabled() {
					status = "enabled"
				}
				fmt.Printf("  ├─ %-18s  v%s  (%s)\n", p.Name(), p.Version(), status)
			}
			fmt.Println()

			// Subsystems
			fmt.Println("Subsystems:")
			printSubsystem("Scheduler    ", h.Scheduler() != nil)
			printSubsystem("Heartbeat    ", h.Heartbeat() != nil)
			printSubsystem("Task Manager ", h.TaskManager() != nil)
			printSubsystem("Token Compress", h.TokenCompressor() != nil)
			printSubsystem("Channel Mgr  ", h.ChannelManager() != nil)
			fmt.Println()

			// Runtime
			fmt.Println("Runtime:")
			fmt.Printf("  Go Version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())
			fmt.Println("════════════════════════════════════════════════════════════")
			return nil
		},
	}
	hostCmd.AddCommand(infoCmd)

	// ── config ───────────────────────────────────────────────────────────────
	cfgCmd := &cobra.Command{
		Use:   "config",
		Short: "查看 Host 完整配置",
		Long:  `以 JSON 格式输出当前 Host 的完整配置（敏感信息已脱敏）。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(cfg, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Host Configuration:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Name:           %s\n", orDefault(cfg.Name, "(unnamed)"))
			fmt.Printf("Version:        %s\n", orDefault(cfg.Version, "(unset)"))
			fmt.Printf("Default Model:  %s\n", orDefault(cfg.DefaultModel, "(unset)"))
			fmt.Printf("Skill Dir:      %s\n", orDefault(cfg.SkillSystemDir, "(unset)"))
			fmt.Printf("Agents:         %d defined\n", len(cfg.Agents))
			fmt.Printf("Workflows:      %d defined\n", len(cfg.Workflows))
			fmt.Printf("Models:         %d configured\n", len(cfg.Models))
			fmt.Println()
			fmt.Println("Optional Subsystems:")
			fmt.Printf("  Scheduler:          %v\n", cfg.Scheduler != nil && cfg.Scheduler.Enabled)
			fmt.Printf("  Heartbeat:          %v\n", cfg.Heartbeat != nil && cfg.Heartbeat.Enabled)
			fmt.Printf("  AsyncTask:          %v\n", cfg.AsyncTask != nil && cfg.AsyncTask.Enabled)
			fmt.Printf("  TokenCompression:   %v\n", cfg.TokenCompression != nil && cfg.TokenCompression.Enabled)
			fmt.Printf("  Messaging:          %v\n", cfg.Messaging != nil && cfg.Messaging.Enabled)
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println()
			fmt.Println("To view full JSON config, use: af host config -o json")
			return nil
		},
	}
	hostCmd.AddCommand(cfgCmd)

	// ── models ───────────────────────────────────────────────────────────────
	modelsCmd := &cobra.Command{
		Use:   "models",
		Short: "列出所有配置的模型",
		Long:  `显示 Host 配置中所有已注册的模型及其参数。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			if len(cfg.Models) == 0 {
				fmt.Println("No models configured")
				return nil
			}

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(cfg.Models, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Configured Models:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for name, m := range cfg.Models {
				defaultMark := ""
				if name == cfg.DefaultModel || (cfg.DefaultModel == "" && name == "default") {
					defaultMark = " [default]"
				}
				fmt.Printf("  %-20s  type=%-12s  model=%s%s\n",
					name, m.Type, m.Model, defaultMark)
			}
			fmt.Printf("Total: %d model(s)\n", len(cfg.Models))
			return nil
		},
	}
	hostCmd.AddCommand(modelsCmd)

	// ── summary ──────────────────────────────────────────────────────────────
	summaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "查看系统能力摘要",
		Long:  `以简洁格式显示系统整体能力状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			h := app.GetHost()
			cfg := h.Config()

			agentList := h.ListAgents()
			workflowList := h.ListWorkflows()
			pluginList := h.PluginManager().GetAllPlugins()

			fmt.Println("AgentFramework System Summary")
			fmt.Println("════════════════════════════════════════════════════════════")

			// Counts
			fmt.Printf("  Agents:          %d\n", len(agentList))
			fmt.Printf("  Workflows:       %d\n", len(workflowList))
			fmt.Printf("  Models:          %d\n", len(cfg.Models))
			fmt.Printf("  Plugins:         %d\n", len(pluginList))

			// Subsystems
			subsystemActive := 0
			for _, active := range []bool{
				h.Scheduler() != nil,
				h.Heartbeat() != nil,
				h.TaskManager() != nil,
				h.TokenCompressor() != nil,
				h.ChannelManager() != nil,
			} {
				if active {
					subsystemActive++
				}
			}
			fmt.Printf("  Subsystems:      %d/5 active\n", subsystemActive)

			// Skill system
			skillInfo := "(disabled)"
			if cfg.SkillSystemDir != "" {
				skillInfo = cfg.SkillSystemDir
			}
			fmt.Printf("  Skill System:    %s\n", skillInfo)
			fmt.Println("════════════════════════════════════════════════════════════")
			return nil
		},
	}
	hostCmd.AddCommand(summaryCmd)

	// ── threads ──────────────────────────────────────────────────────────────
	threadsCmd := &cobra.Command{
		Use:   "threads",
		Short: "查看线程存储状态",
		Long:  `显示线程存储（ThreadStore）的基本配置和状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			h := app.GetHost()
			cfg := h.Config()
			ts := h.ThreadStore()

			fmt.Println("Thread Store Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Type:           %s\n", orDefault(string(cfg.ThreadStore.Type), "memory"))
			if ts == nil {
				fmt.Println("Status:         Not initialized")
			} else {
				fmt.Printf("Status:         Active (%T)\n", ts)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	hostCmd.AddCommand(threadsCmd)

	rootCmd.AddCommand(hostCmd)
}

// orDefault returns val if non-empty, otherwise returns def
func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// printSubsystem prints a subsystem status line
func printSubsystem(name string, active bool) {
	status := "✗ not configured"
	if active {
		status = "✓ active"
	}
	fmt.Printf("  %-16s  %s\n", name, status)
}
