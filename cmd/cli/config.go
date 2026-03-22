// Agent Framework - Config Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"AgentFramework/agent"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置",
	Long:  `管理 AgentFramework 的配置，包括查看、设置、验证和导出配置。`,
}

// addConfigCommands adds config-related commands to root command
func addConfigCommands() {
	// ── get ──────────────────────────────────────────────────────────────────
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "获取配置值",
		Long: `获取指定的配置值。如果不提供 key，则显示所有配置。

支持的 key:
  model / model.default      默认模型配置
  model.<name>               指定模型配置
  skill-dir / skillSystemDir 技能目录
  agents                     Agent 列表
  workflows                  工作流列表
  scheduler                  调度器配置
  heartbeat                  心跳配置
  async-task                 异步任务配置
  token-compression          Token 压缩配置`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			if len(args) == 0 {
				return printConfig(cfg)
			}

			key := args[0]
			return printConfigKey(cfg, key)
		},
	}
	configCmd.AddCommand(getCmd)

	// ── set ──────────────────────────────────────────────────────────────────
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "设置配置值",
		Long: `设置指定的配置值。支持动态修改配置。

示例:
  af config set model llama3
  af config set model.default ollama/llama3
  af config set skill-dir ./.skills`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg := app.GetHost().Config()

			switch key {
			case "model", "model.default":
				cfg.Models["default"] = agent.ModelConfig{
					Type:  "ollama",
					Model: value,
				}
				fmt.Printf("Model set to: %s\n", value)
			case "skill-dir", "skillSystemDir":
				cfg.SkillSystemDir = value
				fmt.Printf("Skill directory set to: %s\n", value)
			case "default-model":
				cfg.DefaultModel = value
				fmt.Printf("Default model set to: %s\n", value)
			default:
				return fmt.Errorf("unknown config key: %s\nUse 'af config list-keys' to see all available keys", key)
			}

			if configFile != "" {
				if err := agent.SaveHostConfigFile(configFile, cfg); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Println("Configuration saved.")
			} else {
				fmt.Println("Note: Configuration not persisted (no --config file specified)")
			}

			return nil
		},
	}
	configCmd.AddCommand(setCmd)

	// ── validate ─────────────────────────────────────────────────────────────
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "验证配置",
		Long:  `验证当前配置的完整性和有效性。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			fmt.Println("Validating configuration...")
			fmt.Println()

			hasErrors := false

			// 检查模型
			if len(cfg.Models) == 0 {
				fmt.Println("✗ No models configured")
				hasErrors = true
			} else {
				for name, model := range cfg.Models {
					if model.Model == "" {
						fmt.Printf("✗ Model '%s' has no model name specified\n", name)
						hasErrors = true
					} else {
						fmt.Printf("✓ Model '%s': %s/%s\n", name, model.Type, model.Model)
					}
				}
			}

			// 检查技能目录
			if cfg.SkillSystemDir == "" {
				fmt.Println("⚠ No skill directory configured")
			} else {
				fmt.Printf("✓ Skill directory: %s\n", cfg.SkillSystemDir)
			}

			// 检查 Agents
			if len(cfg.Agents) == 0 {
				fmt.Println("⚠ No agents defined in configuration")
			} else {
				fmt.Printf("✓ Agents defined: %d\n", len(cfg.Agents))
				for _, spec := range cfg.Agents {
					if spec.Name == "" {
						fmt.Println("  ✗ Agent has no name")
						hasErrors = true
					} else if spec.Kind == "" {
						fmt.Printf("  ⚠ Agent '%s' has no kind specified\n", spec.Name)
					} else {
						fmt.Printf("  ✓ Agent: %s (kind=%s, model=%s)\n", spec.Name, spec.Kind, spec.Model)
					}
				}
			}

			// 检查工作流
			if len(cfg.Workflows) > 0 {
				fmt.Printf("✓ Workflows defined: %d\n", len(cfg.Workflows))
				for _, spec := range cfg.Workflows {
					fmt.Printf("  ✓ Workflow: %s (kind=%s)\n", spec.Name, spec.Kind)
				}
			}

			// 检查可选特性
			if cfg.Scheduler != nil && cfg.Scheduler.Enabled {
				fmt.Printf("✓ Scheduler: enabled (maxJobs=%d)\n", cfg.Scheduler.MaxJobs)
			}
			if cfg.AsyncTask != nil && cfg.AsyncTask.Enabled {
				fmt.Printf("✓ AsyncTask: enabled (maxTasks=%d)\n", cfg.AsyncTask.MaxTasks)
			}
			if cfg.TokenCompression != nil && cfg.TokenCompression.Enabled {
				fmt.Printf("✓ TokenCompression: enabled (strategy=%s)\n", cfg.TokenCompression.Strategy)
			}
			if cfg.Heartbeat != nil && cfg.Heartbeat.Enabled {
				fmt.Printf("✓ Heartbeat: enabled (interval=%ds)\n", cfg.Heartbeat.Interval)
			}

			fmt.Println()
			if hasErrors {
				return fmt.Errorf("configuration validation failed with errors")
			}
			fmt.Println("✓ Configuration is valid")
			return nil
		},
	}
	configCmd.AddCommand(validateCmd)

	// ── show ─────────────────────────────────────────────────────────────────
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "显示完整配置（YAML 格式）",
		Long:  `以 YAML 格式输出完整的当前配置。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Print(string(data))
			return nil
		},
	}
	configCmd.AddCommand(showCmd)

	// ── list-keys ────────────────────────────────────────────────────────────
	listKeysCmd := &cobra.Command{
		Use:   "list-keys",
		Short: "列出所有可用的配置键",
		Long:  `列出所有可通过 'af config get/set' 操作的配置键及其说明。`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available Configuration Keys:")
			fmt.Println("────────────────────────────────────────────────────────────")
			keys := []struct {
				key  string
				desc string
			}{
				{"model", "默认模型配置 (格式: model-name)"},
				{"model.default", "默认模型配置 (同 model)"},
				{"model.<name>", "指定名称的模型配置"},
				{"default-model", "默认模型键名"},
				{"skill-dir", "技能系统目录路径"},
				{"skillSystemDir", "技能系统目录路径 (同 skill-dir)"},
				{"agents", "Agent 配置列表 (只读)"},
				{"workflows", "工作流配置列表 (只读)"},
				{"scheduler", "调度器配置 (只读)"},
				{"heartbeat", "心跳服务配置 (只读)"},
				{"async-task", "异步任务配置 (只读)"},
				{"token-compression", "Token 压缩配置 (只读)"},
			}
			for _, k := range keys {
				fmt.Printf("  %-25s  %s\n", k.key, k.desc)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println("Use 'af config get <key>' to read a specific key")
			fmt.Println("Use 'af config set <key> <value>' to set a writable key")
		},
	}
	configCmd.AddCommand(listKeysCmd)

	// ── export ───────────────────────────────────────────────────────────────
	exportCmd := &cobra.Command{
		Use:   "export [output-file]",
		Short: "导出配置到文件",
		Long:  `将当前配置导出到 YAML 文件。如果不指定输出文件，则输出到 stdout。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			if len(args) == 0 {
				data, err := yaml.Marshal(cfg)
				if err != nil {
					return fmt.Errorf("failed to marshal config: %w", err)
				}
				fmt.Print(string(data))
				return nil
			}

			outputFile := args[0]
			if err := agent.SaveHostConfigFile(outputFile, cfg); err != nil {
				return fmt.Errorf("failed to export config: %w", err)
			}

			fmt.Printf("Configuration exported to: %s\n", outputFile)
			return nil
		},
	}
	configCmd.AddCommand(exportCmd)

	// ── import ───────────────────────────────────────────────────────────────
	importCmd := &cobra.Command{
		Use:   "import [config-file]",
		Short: "从文件导入配置",
		Long:  `从指定的 YAML 文件加载配置（需要重启才能完全生效）。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile := args[0]
			cfg, err := agent.LoadHostConfigFile(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config from '%s': %w", cfgFile, err)
			}

			// 验证基本结构
			fmt.Printf("Configuration loaded from: %s\n", cfgFile)
			fmt.Printf("  Models: %d\n", len(cfg.Models))
			fmt.Printf("  Agents: %d\n", len(cfg.Agents))
			fmt.Printf("  Workflows: %d\n", len(cfg.Workflows))
			fmt.Println()
			fmt.Println("Note: Restart required for configuration changes to take full effect.")
			return nil
		},
	}
	configCmd.AddCommand(importCmd)

	rootCmd.AddCommand(configCmd)
}

// printConfig 打印完整配置
func printConfig(cfg *agent.HostConfig) error {
	switch outputFormat {
	case "json":
		b, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "yaml":
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		fmt.Println("Current Configuration:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("Name:            %s\n", cfg.Name)
		fmt.Printf("Version:         %s\n", cfg.Version)
		fmt.Printf("Default Model:   %s\n", cfg.DefaultModel)
		fmt.Printf("Skill Dir:       %s\n", cfg.SkillSystemDir)
		fmt.Println()

		if len(cfg.Models) > 0 {
			fmt.Println("Models:")
			for name, model := range cfg.Models {
				fmt.Printf("  %-15s  type=%-10s  model=%s\n", name, model.Type, model.Model)
			}
			fmt.Println()
		}

		if len(cfg.Agents) > 0 {
			fmt.Printf("Agents: %d defined\n", len(cfg.Agents))
			for _, spec := range cfg.Agents {
				fmt.Printf("  %-20s  kind=%-10s  model=%s\n", spec.Name, spec.Kind, spec.Model)
			}
			fmt.Println()
		}

		if len(cfg.Workflows) > 0 {
			fmt.Printf("Workflows: %d defined\n", len(cfg.Workflows))
			for _, spec := range cfg.Workflows {
				fmt.Printf("  %-20s  kind=%s\n", spec.Name, spec.Kind)
			}
			fmt.Println()
		}

		if cfg.Scheduler != nil {
			fmt.Printf("Scheduler:       enabled=%v, maxJobs=%d\n", cfg.Scheduler.Enabled, cfg.Scheduler.MaxJobs)
		}
		if cfg.AsyncTask != nil {
			fmt.Printf("AsyncTask:       enabled=%v, maxTasks=%d\n", cfg.AsyncTask.Enabled, cfg.AsyncTask.MaxTasks)
		}
		if cfg.TokenCompression != nil {
			fmt.Printf("TokenCompress:   enabled=%v, strategy=%s\n", cfg.TokenCompression.Enabled, cfg.TokenCompression.Strategy)
		}
		if cfg.Heartbeat != nil {
			fmt.Printf("Heartbeat:       enabled=%v, interval=%ds\n", cfg.Heartbeat.Enabled, cfg.Heartbeat.Interval)
		}
		fmt.Println("────────────────────────────────────────────────────────────")
	}
	return nil
}

// printConfigKey 打印特定配置键
func printConfigKey(cfg *agent.HostConfig, key string) error {
	switch key {
	case "model", "model.default":
		if model, ok := cfg.Models["default"]; ok {
			fmt.Printf("%s/%s\n", model.Type, model.Model)
		} else {
			fmt.Println("(not configured)")
		}
	case "skill-dir", "skillSystemDir":
		fmt.Println(cfg.SkillSystemDir)
	case "default-model":
		fmt.Println(cfg.DefaultModel)
	case "agents":
		if len(cfg.Agents) == 0 {
			fmt.Println("(no agents defined)")
		}
		for _, spec := range cfg.Agents {
			fmt.Printf("%s (kind=%s, model=%s)\n", spec.Name, spec.Kind, spec.Model)
		}
	case "workflows":
		if len(cfg.Workflows) == 0 {
			fmt.Println("(no workflows defined)")
		}
		for _, spec := range cfg.Workflows {
			fmt.Printf("%s (kind=%s)\n", spec.Name, spec.Kind)
		}
	case "scheduler":
		if cfg.Scheduler == nil {
			fmt.Println("(not configured)")
		} else {
			fmt.Printf("enabled=%v, maxJobs=%d, jobTimeout=%ds\n",
				cfg.Scheduler.Enabled, cfg.Scheduler.MaxJobs, cfg.Scheduler.JobTimeout)
		}
	case "heartbeat":
		if cfg.Heartbeat == nil {
			fmt.Println("(not configured)")
		} else {
			fmt.Printf("enabled=%v, interval=%ds, timeout=%ds\n",
				cfg.Heartbeat.Enabled, cfg.Heartbeat.Interval, cfg.Heartbeat.Timeout)
		}
	case "async-task", "asyncTask":
		if cfg.AsyncTask == nil {
			fmt.Println("(not configured)")
		} else {
			fmt.Printf("enabled=%v, maxTasks=%d, taskTimeout=%ds\n",
				cfg.AsyncTask.Enabled, cfg.AsyncTask.MaxTasks, cfg.AsyncTask.TaskTimeout)
		}
	case "token-compression", "tokenCompression":
		if cfg.TokenCompression == nil {
			fmt.Println("(not configured)")
		} else {
			fmt.Printf("enabled=%v, strategy=%s, targetTokens=%d\n",
				cfg.TokenCompression.Enabled,
				cfg.TokenCompression.Strategy,
				cfg.TokenCompression.TargetTokens)
		}
	default:
		// 尝试从 models 中查找
		if len(key) > 6 && key[:6] == "model." {
			modelName := key[6:]
			if model, ok := cfg.Models[modelName]; ok {
				fmt.Printf("%s/%s\n", model.Type, model.Model)
				return nil
			}
		}
		return fmt.Errorf("unknown config key: %s\nUse 'af config list-keys' to see available keys", key)
	}
	return nil
}
