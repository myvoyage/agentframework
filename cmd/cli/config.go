// Agent Framework - Config Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"AgentFramework/agent"
)

var (
	_configOutput string
	_configForce  bool
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置",
	Long: `管理 AgentFramework 的配置，包括创建、查看、设置、验证等操作。

子命令:
  init        创建默认配置文件
  path        显示配置文件路径
  show        显示完整配置
  get         获取配置值
  set         设置配置值
  edit        编辑配置文件
  validate    验证配置
  models      管理模型配置
  agents      管理 Agent 配置
  export      导出配置
  import      导入配置
  list-keys   列出所有配置键

示例:
  af config init                    # 创建默认配置
  af config init --force            # 强制覆盖现有配置
  af config path                    # 显示配置文件路径
  af config show                    # 显示完整配置
  af config get model               # 获取模型配置
  af config set model llama3        # 设置默认模型
  af config edit                    # 编辑配置文件
  af config validate                # 验证配置
  af config models list             # 列出所有模型
  af config models add my-model ollama:qwen2.5  # 添加模型
`,
}

// addConfigCommands adds config-related commands to root command
func addConfigCommands() {
	// Global flags
	configCmd.PersistentFlags().StringVarP(&_configOutput, "output", "o", "", "输出格式 (yaml|json)")
	configCmd.PersistentFlags().BoolVarP(&_configForce, "force", "f", false, "强制操作（如覆盖现有文件）")

	// Add subcommands
	addConfigInitCmd()
	addConfigPathCmd()
	addConfigShowCmd()
	addConfigGetCmd()
	addConfigSetCmd()
	addConfigEditCmd()
	addConfigValidateCmd()
	addConfigModelsCmd()
	addConfigAgentsCmd()
	addConfigExportCmd()
	addConfigImportCmd()
	addConfigListKeysCmd()

	rootCmd.AddCommand(configCmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config init
// ════════════════════════════════════════════════════════════════════════════

func addConfigInitCmd() {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "创建默认配置文件",
		Long: `创建默认的配置文件。如果配置文件已存在，需要使用 --force 参数覆盖。

示例:
  af config init                    # 在默认位置创建配置
  af config init --force            # 强制覆盖现有配置
  af config init -o my_config.yaml  # 在指定位置创建配置`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			configPath := filepath.Join(dataDir, "config.yaml")

			// Check if config exists
			if _, err := os.Stat(configPath); err == nil && !_configForce {
				return fmt.Errorf("配置文件已存在: %s\n使用 --force 参数覆盖", configPath)
			}

			// Create default config
			defaultCfg := createDefaultConfig()

			// Ensure directory exists
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return fmt.Errorf("创建数据目录失败: %w", err)
			}

			// Save config
			if err := agent.SaveHostConfigFile(configPath, defaultCfg); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			fmt.Println("✓ 配置文件已创建:")
			fmt.Printf("  路径: %s\n", configPath)
			fmt.Println()
			fmt.Println("默认配置:")
			fmt.Println("  模型: ollama/llama3")
			fmt.Println("  技能目录: .skills")
			fmt.Println()
			fmt.Println("快速开始:")
			fmt.Println("  af config edit      # 编辑配置")
			fmt.Println("  af agent list       # 列出 agents")

			return nil
		},
	}
	configCmd.AddCommand(cmd)
}

// createDefaultConfig creates a default HostConfig
func createDefaultConfig() *agent.HostConfig {
	return &agent.HostConfig{
		Name:         "agentframework",
		DefaultModel: "default",
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		Agents: []agent.AgentSpec{
			{
				Name:         "default",
				Kind:         "chat",
				Model:        "default",
				Instructions: "You are a helpful AI assistant.",
			},
			{
				Name:         "coder",
				Kind:         "chat",
				Model:        "default",
				Instructions:  "You are an expert programmer. Help with coding tasks.",
			},
		},
		SkillSystemDir: ".skills",
	}
}

// ════════════════════════════════════════════════════════════════════════════
// config path
// ════════════════════════════════════════════════════════════════════════════

func addConfigPathCmd() {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "显示配置文件路径",
		Long:  `显示当前使用的配置文件路径。`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := getDefaultDataDir()
			defaultPath := filepath.Join(dataDir, "config.yaml")

			fmt.Println("配置文件位置:")
			fmt.Println()

			// Check current config
			if configFile != "" {
				fmt.Printf("  当前配置: %s\n", configFile)
			} else {
				fmt.Println("  当前配置: (使用默认配置)")
			}

			fmt.Printf("  默认位置: %s\n", defaultPath)

			// Check if default config exists
			if _, err := os.Stat(defaultPath); err == nil {
				fmt.Println("  状态: ✓ 配置文件存在")
			} else {
				fmt.Println("  状态: ✗ 配置文件不存在，使用内置默认配置")
			}

			fmt.Println()
			fmt.Println("数据目录:")
			fmt.Printf("  位置: %s\n", dataDir)
			fmt.Printf("  环境变量: AGENT_FRAMEWORK_DATA_DIR\n")
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config show
// ════════════════════════════════════════════════════════════════════════════

func addConfigShowCmd() {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "显示完整配置",
		Long:  `以 YAML 或 JSON 格式输出完整的当前配置。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			return printConfig(cfg, _configOutput)
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config get
// ════════════════════════════════════════════════════════════════════════════

func addConfigGetCmd() {
	cmd := &cobra.Command{
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
				return printConfig(cfg, _configOutput)
			}

			key := args[0]
			return printConfigKey(cfg, key, _configOutput)
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config set
// ════════════════════════════════════════════════════════════════════════════

func addConfigSetCmd() {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置值",
		Long: `设置指定的配置值。支持动态修改配置。

示例:
  af config set model llama3
  af config set model.default ollama/llama3
  af config set skill-dir ./.skills
  af config set default-model my-model`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg := app.GetHost().Config()

			switch key {
			case "model", "model.default":
				// Parse model string: "type:model" or just "model"
				modelType := "ollama"
				modelValue := value

				if parts := strings.SplitN(value, ":", 2); len(parts) == 2 {
					modelType = parts[0]
					modelValue = parts[1]
				}

				cfg.Models["default"] = agent.ModelConfig{
					Type:  modelType,
					Model: modelValue,
				}
				fmt.Printf("✓ 默认模型已设置: %s/%s\n", modelType, modelValue)

			case "skill-dir", "skillSystemDir":
				cfg.SkillSystemDir = value
				fmt.Printf("✓ 技能目录已设置: %s\n", value)

			case "default-model":
				cfg.DefaultModel = value
				fmt.Printf("✓ 默认模型键已设置: %s\n", value)

			default:
				return fmt.Errorf("未知的配置键: %s\n使用 'af config list-keys' 查看所有可用的键", key)
			}

			// Save config
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
				fmt.Printf("配置已保存到: %s\n", configPath)
			} else {
				fmt.Println("注意: 配置未持久化（未指定配置文件）")
			}

			return nil
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config edit
// ════════════════════════════════════════════════════════════════════════════

func addConfigEditCmd() {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "编辑配置文件",
		Long: `使用系统默认编辑器打开配置文件进行编辑。

编辑器选择顺序:
  1. 环境变量 EDITOR
  2. 环境变量 VISUAL
  3. 系统默认编辑器 (notepad/vim/nano)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := getConfigPath()
			if configPath == "" {
				dataDir, _ := getDefaultDataDir()
				configPath = filepath.Join(dataDir, "config.yaml")
			}

			// Check if config exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return fmt.Errorf("配置文件不存在: %s\n使用 'af config init' 创建配置", configPath)
			}

			// Get editor
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = os.Getenv("VISUAL")
			}
			if editor == "" {
				if runtime.GOOS == "windows" {
					editor = "notepad"
				} else {
					editor = "vim"
				}
			}

			fmt.Printf("打开编辑器: %s %s\n", editor, configPath)

			// Run editor
			editCmd := exec.Command(editor, configPath)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			if err := editCmd.Run(); err != nil {
				return fmt.Errorf("编辑器运行失败: %w", err)
			}

			fmt.Println("配置文件已编辑。如需应用更改，请重启服务。")
			return nil
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config validate
// ════════════════════════════════════════════════════════════════════════════

func addConfigValidateCmd() {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "验证配置",
		Long:  `验证当前配置的完整性和有效性。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			return validateConfig(cfg)
		},
	}
	configCmd.AddCommand(cmd)
}

// validateConfig validates the configuration
func validateConfig(cfg *agent.HostConfig) error {
	fmt.Println("验证配置...")
	fmt.Println()

	var errors []string
	var warnings []string

	// Check models
	if len(cfg.Models) == 0 {
		errors = append(errors, "没有配置模型")
	} else {
		for name, model := range cfg.Models {
			if model.Model == "" {
				errors = append(errors, fmt.Sprintf("模型 '%s' 没有指定模型名", name))
			}
			if model.Type == "" {
				warnings = append(warnings, fmt.Sprintf("模型 '%s' 没有指定类型，将使用默认值 'ollama'", name))
			}
		}
	}

	// Check default model
	if cfg.DefaultModel != "" {
		if _, ok := cfg.Models[cfg.DefaultModel]; !ok {
			errors = append(errors, fmt.Sprintf("默认模型 '%s' 不存在于模型配置中", cfg.DefaultModel))
		}
	}

	// Check agents
	if len(cfg.Agents) == 0 {
		warnings = append(warnings, "没有定义 Agent")
	} else {
		for _, spec := range cfg.Agents {
			if spec.Name == "" {
				errors = append(errors, "Agent 缺少名称")
			}
			if spec.Kind == "" {
				warnings = append(warnings, fmt.Sprintf("Agent '%s' 没有指定类型", spec.Name))
			}
			if spec.Model != "" {
				if _, ok := cfg.Models[spec.Model]; !ok {
					warnings = append(warnings, fmt.Sprintf("Agent '%s' 引用了不存在的模型 '%s'", spec.Name, spec.Model))
				}
			}
		}
	}

	// Check skill directory
	if cfg.SkillSystemDir == "" {
		warnings = append(warnings, "没有配置技能目录")
	}

	// Print results
	fmt.Println("模型配置:")
	if len(cfg.Models) == 0 {
		fmt.Println("  ✗ 没有配置模型")
	} else {
		for name, model := range cfg.Models {
			fmt.Printf("  ✓ %s: %s/%s\n", name, model.Type, model.Model)
		}
	}
	fmt.Println()

	fmt.Println("Agent 配置:")
	if len(cfg.Agents) == 0 {
		fmt.Println("  ⚠ 没有定义 Agent")
	} else {
		for _, spec := range cfg.Agents {
			fmt.Printf("  ✓ %s: kind=%s, model=%s\n", spec.Name, spec.Kind, spec.Model)
		}
	}
	fmt.Println()

	if len(cfg.Workflows) > 0 {
		fmt.Println("工作流:")
		for _, spec := range cfg.Workflows {
			fmt.Printf("  ✓ %s: kind=%s\n", spec.Name, spec.Kind)
		}
		fmt.Println()
	}

	// Print optional features
	fmt.Println("可选功能:")
	if cfg.Scheduler != nil && cfg.Scheduler.Enabled {
		fmt.Printf("  ✓ 调度器: maxJobs=%d\n", cfg.Scheduler.MaxJobs)
	}
	if cfg.AsyncTask != nil && cfg.AsyncTask.Enabled {
		fmt.Printf("  ✓ 异步任务: maxTasks=%d\n", cfg.AsyncTask.MaxTasks)
	}
	if cfg.TokenCompression != nil && cfg.TokenCompression.Enabled {
		fmt.Printf("  ✓ Token 压缩: strategy=%s\n", cfg.TokenCompression.Strategy)
	}
	if cfg.Heartbeat != nil && cfg.Heartbeat.Enabled {
		fmt.Printf("  ✓ 心跳: interval=%ds\n", cfg.Heartbeat.Interval)
	}
	fmt.Println()

	// Print errors
	if len(errors) > 0 {
		fmt.Println("错误:")
		for _, err := range errors {
			fmt.Printf("  ✗ %s\n", err)
		}
		fmt.Println()
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Println("警告:")
		for _, warn := range warnings {
			fmt.Printf("  ⚠ %s\n", warn)
		}
		fmt.Println()
	}

	// Summary
	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败，发现 %d 个错误", len(errors))
	}

	fmt.Println("✓ 配置验证通过")
	return nil
}

// ════════════════════════════════════════════════════════════════════════════
// config models
// ════════════════════════════════════════════════════════════════════════════

func addConfigModelsCmd() {
	modelsCmd := &cobra.Command{
		Use:   "models",
		Short: "管理模型配置",
		Long:  `管理模型配置，包括列出、添加、删除模型。`,
	}
	configCmd.AddCommand(modelsCmd)

	// models list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有模型",
		Long:  `列出所有已配置的模型。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			fmt.Println("已配置的模型:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("%-15s %-12s %-20s %s\n", "名称", "类型", "模型", "Base URL")
			fmt.Println("────────────────────────────────────────────────────────────")

			for name, model := range cfg.Models {
				defaultMarker := ""
				if name == cfg.DefaultModel {
					defaultMarker = " (default)"
				}
				baseURL := model.BaseURL
				if baseURL == "" {
					switch model.Type {
					case "ollama":
						baseURL = "http://localhost:11434"
					case "openai":
						baseURL = "https://api.openai.com/v1"
					case "lmstudio":
						baseURL = "http://localhost:1234/v1"
					}
				}
				fmt.Printf("%-15s %-12s %-20s %s%s\n", name, model.Type, model.Model, baseURL, defaultMarker)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	modelsCmd.AddCommand(listCmd)

	// models add
	addCmd := &cobra.Command{
		Use:   "add <name> <type:model>",
		Short: "添加模型配置",
		Long: `添加新的模型配置。

示例:
  af config models add my-ollama ollama:llama3
  af config models add my-gpt openai:gpt-4
  af config models add my-local lmstudio:qwen2.5`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			modelStr := args[1]

			// Parse model string
			parts := strings.SplitN(modelStr, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("模型格式错误，应为 'type:model'，例如 'ollama:llama3'")
			}

			modelType := parts[0]
			modelValue := parts[1]

			cfg := app.GetHost().Config()

			if cfg.Models == nil {
				cfg.Models = make(map[string]agent.ModelConfig)
			}

			cfg.Models[name] = agent.ModelConfig{
				Type:  modelType,
				Model: modelValue,
			}

			// Save
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			}

			fmt.Printf("✓ 模型 '%s' 已添加: %s/%s\n", name, modelType, modelValue)
			return nil
		},
	}
	modelsCmd.AddCommand(addCmd)

	// models remove
	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "删除模型配置",
		Long:  `删除指定的模型配置。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg := app.GetHost().Config()

			if _, ok := cfg.Models[name]; !ok {
				return fmt.Errorf("模型 '%s' 不存在", name)
			}

			// Check if it's the default model
			if name == cfg.DefaultModel {
				return fmt.Errorf("不能删除默认模型 '%s'，请先设置其他默认模型", name)
			}

			delete(cfg.Models, name)

			// Save
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			}

			fmt.Printf("✓ 模型 '%s' 已删除\n", name)
			return nil
		},
	}
	modelsCmd.AddCommand(removeCmd)

	// models set-default
	setDefaultCmd := &cobra.Command{
		Use:   "set-default <name>",
		Short: "设置默认模型",
		Long:  `设置默认使用的模型。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg := app.GetHost().Config()

			if _, ok := cfg.Models[name]; !ok {
				return fmt.Errorf("模型 '%s' 不存在", name)
			}

			cfg.DefaultModel = name

			// Save
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			}

			fmt.Printf("✓ 默认模型已设置为 '%s'\n", name)
			return nil
		},
	}
	modelsCmd.AddCommand(setDefaultCmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config agents
// ════════════════════════════════════════════════════════════════════════════

func addConfigAgentsCmd() {
	agentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "管理 Agent 配置",
		Long:  `管理 Agent 配置，包括列出、添加、删除 Agent。`,
	}
	configCmd.AddCommand(agentsCmd)

	// agents list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有 Agent",
		Long:  `列出所有已配置的 Agent。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			fmt.Println("已配置的 Agent:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("%-20s %-10s %-15s %s\n", "名称", "类型", "模型", "指令预览")
			fmt.Println("────────────────────────────────────────────────────────────")

			for _, spec := range cfg.Agents {
				instructions := spec.Instructions
				if len(instructions) > 30 {
					instructions = instructions[:30] + "..."
				}
				fmt.Printf("%-20s %-10s %-15s %s\n", spec.Name, spec.Kind, spec.Model, instructions)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	agentsCmd.AddCommand(listCmd)

	// agents add
	addCmd := &cobra.Command{
		Use:   "add <name> --kind <kind> --model <model>",
		Short: "添加 Agent 配置",
		Long: `添加新的 Agent 配置。

示例:
  af config agents add my-agent --kind chat --model default
  af config agents add coder --kind chat --model gpt-4 --instructions "You are a coder."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			kind, _ := cmd.Flags().GetString("kind")
			model, _ := cmd.Flags().GetString("model")
			instructions, _ := cmd.Flags().GetString("instructions")

			if kind == "" {
				kind = "chat"
			}
			if model == "" {
				model = "default"
			}
			if instructions == "" {
				instructions = "You are a helpful AI assistant."
			}

			cfg := app.GetHost().Config()

			// Check if model exists
			if _, ok := cfg.Models[model]; !ok {
				return fmt.Errorf("模型 '%s' 不存在", model)
			}

			// Check if agent already exists
			for _, spec := range cfg.Agents {
				if spec.Name == name {
					return fmt.Errorf("Agent '%s' 已存在", name)
				}
			}

			cfg.Agents = append(cfg.Agents, agent.AgentSpec{
				Name:         name,
				Kind:         kind,
				Model:        model,
				Instructions: instructions,
			})

			// Save
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			}

			fmt.Printf("✓ Agent '%s' 已添加\n", name)
			return nil
		},
	}
	addCmd.Flags().String("kind", "chat", "Agent 类型 (chat|react)")
	addCmd.Flags().String("model", "default", "使用的模型")
	addCmd.Flags().String("instructions", "", "Agent 指令")
	agentsCmd.AddCommand(addCmd)

	// agents remove
	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "删除 Agent 配置",
		Long:  `删除指定的 Agent 配置。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg := app.GetHost().Config()

			found := false
			newAgents := make([]agent.AgentSpec, 0, len(cfg.Agents))
			for _, spec := range cfg.Agents {
				if spec.Name == name {
					found = true
					continue
				}
				newAgents = append(newAgents, spec)
			}

			if !found {
				return fmt.Errorf("Agent '%s' 不存在", name)
			}

			cfg.Agents = newAgents

			// Save
			configPath := getConfigPath()
			if configPath != "" {
				if err := agent.SaveHostConfigFile(configPath, cfg); err != nil {
					return fmt.Errorf("保存配置失败: %w", err)
				}
			}

			fmt.Printf("✓ Agent '%s' 已删除\n", name)
			return nil
		},
	}
	agentsCmd.AddCommand(removeCmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config export
// ════════════════════════════════════════════════════════════════════════════

func addConfigExportCmd() {
	cmd := &cobra.Command{
		Use:   "export [output-file]",
		Short: "导出配置到文件",
		Long:  `将当前配置导出到 YAML 文件。如果不指定输出文件，则输出到 stdout。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()

			if len(args) == 0 {
				data, err := yaml.Marshal(cfg)
				if err != nil {
					return fmt.Errorf("序列化配置失败: %w", err)
				}
				fmt.Print(string(data))
				return nil
			}

			outputFile := args[0]
			if err := agent.SaveHostConfigFile(outputFile, cfg); err != nil {
				return fmt.Errorf("导出配置失败: %w", err)
			}

			fmt.Printf("✓ 配置已导出到: %s\n", outputFile)
			return nil
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config import
// ════════════════════════════════════════════════════════════════════════════

func addConfigImportCmd() {
	cmd := &cobra.Command{
		Use:   "import <config-file>",
		Short: "从文件导入配置",
		Long:  `从指定的 YAML 文件加载配置（需要重启才能完全生效）。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile := args[0]
			cfg, err := agent.LoadHostConfigFile(cfgFile)
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}

			// Validate
			fmt.Printf("配置已从 %s 加载:\n", cfgFile)
			fmt.Printf("  模型: %d\n", len(cfg.Models))
			fmt.Printf("  Agent: %d\n", len(cfg.Agents))
			fmt.Printf("  工作流: %d\n", len(cfg.Workflows))

			// Validate
			if err := validateConfig(cfg); err != nil {
				fmt.Println()
				return err
			}

			fmt.Println()
			fmt.Println("注意: 需要重启服务才能使配置生效")
			return nil
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// config list-keys
// ════════════════════════════════════════════════════════════════════════════

func addConfigListKeysCmd() {
	cmd := &cobra.Command{
		Use:   "list-keys",
		Short: "列出所有可用的配置键",
		Long:  `列出所有可通过 'af config get/set' 操作的配置键及其说明。`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("可用的配置键:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("  %-25s  %s\n", "键名", "说明")
			fmt.Println("────────────────────────────────────────────────────────────")

			keys := []struct {
				key  string
				desc string
				rw   string
			}{
				{"model", "默认模型配置", "读写"},
				{"model.default", "默认模型配置 (同 model)", "读写"},
				{"model.<name>", "指定名称的模型配置", "只读"},
				{"default-model", "默认模型键名", "读写"},
				{"skill-dir", "技能系统目录路径", "读写"},
				{"skillSystemDir", "技能系统目录路径", "读写"},
				{"agents", "Agent 配置列表", "只读"},
				{"workflows", "工作流配置列表", "只读"},
				{"scheduler", "调度器配置", "只读"},
				{"heartbeat", "心跳服务配置", "只读"},
				{"async-task", "异步任务配置", "只读"},
				{"token-compression", "Token 压缩配置", "只读"},
			}

			for _, k := range keys {
				fmt.Printf("  %-25s  %-25s [%s]\n", k.key, k.desc, k.rw)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println()
			fmt.Println("使用方法:")
			fmt.Println("  af config get <key>        # 读取配置")
			fmt.Println("  af config set <key> <val>  # 写入配置")
			fmt.Println()
			fmt.Println("模型管理:")
			fmt.Println("  af config models list      # 列出所有模型")
			fmt.Println("  af config models add       # 添加模型")
			fmt.Println("  af config models remove    # 删除模型")
		},
	}
	configCmd.AddCommand(cmd)
}

// ════════════════════════════════════════════════════════════════════════════
// Helper Functions
// ════════════════════════════════════════════════════════════════════════════

// getConfigPath returns the current config file path
func getConfigPath() string {
	if configFile != "" {
		return configFile
	}
	dataDir, _ := getDefaultDataDir()
	configPath := filepath.Join(dataDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}
	return ""
}

// printConfig prints the configuration in the specified format
func printConfig(cfg *agent.HostConfig, format string) error {
	switch format {
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
		fmt.Println("当前配置:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("名称:            %s\n", cfg.Name)
		fmt.Printf("版本:            %s\n", cfg.Version)
		fmt.Printf("默认模型:        %s\n", cfg.DefaultModel)
		fmt.Printf("技能目录:        %s\n", cfg.SkillSystemDir)
		fmt.Println()

		if len(cfg.Models) > 0 {
			fmt.Println("模型:")
			for name, model := range cfg.Models {
				fmt.Printf("  %-15s  type=%-10s  model=%s\n", name, model.Type, model.Model)
			}
			fmt.Println()
		}

		if len(cfg.Agents) > 0 {
			fmt.Printf("Agent: %d 个\n", len(cfg.Agents))
			for _, spec := range cfg.Agents {
				fmt.Printf("  %-20s  kind=%-10s  model=%s\n", spec.Name, spec.Kind, spec.Model)
			}
			fmt.Println()
		}

		if len(cfg.Workflows) > 0 {
			fmt.Printf("工作流: %d 个\n", len(cfg.Workflows))
			for _, spec := range cfg.Workflows {
				fmt.Printf("  %-20s  kind=%s\n", spec.Name, spec.Kind)
			}
			fmt.Println()
		}

		// Optional features
		if cfg.Scheduler != nil && cfg.Scheduler.Enabled {
			fmt.Printf("调度器:          enabled=%v, maxJobs=%d\n", cfg.Scheduler.Enabled, cfg.Scheduler.MaxJobs)
		}
		if cfg.AsyncTask != nil && cfg.AsyncTask.Enabled {
			fmt.Printf("异步任务:        enabled=%v, maxTasks=%d\n", cfg.AsyncTask.Enabled, cfg.AsyncTask.MaxTasks)
		}
		if cfg.TokenCompression != nil && cfg.TokenCompression.Enabled {
			fmt.Printf("Token 压缩:      enabled=%v, strategy=%s\n", cfg.TokenCompression.Enabled, cfg.TokenCompression.Strategy)
		}
		if cfg.Heartbeat != nil && cfg.Heartbeat.Enabled {
			fmt.Printf("心跳:            enabled=%v, interval=%ds\n", cfg.Heartbeat.Enabled, cfg.Heartbeat.Interval)
		}
		fmt.Println("────────────────────────────────────────────────────────────")
	}
	return nil
}

// printConfigKey prints a specific configuration key
func printConfigKey(cfg *agent.HostConfig, key string, format string) error {
	switch key {
	case "model", "model.default":
		if model, ok := cfg.Models["default"]; ok {
			fmt.Printf("%s/%s\n", model.Type, model.Model)
		} else {
			fmt.Println("(未配置)")
		}
	case "skill-dir", "skillSystemDir":
		fmt.Println(cfg.SkillSystemDir)
	case "default-model":
		fmt.Println(cfg.DefaultModel)
	case "agents":
		if len(cfg.Agents) == 0 {
			fmt.Println("(没有定义 Agent)")
			return nil
		}
		for _, spec := range cfg.Agents {
			fmt.Printf("%s (kind=%s, model=%s)\n", spec.Name, spec.Kind, spec.Model)
		}
	case "workflows":
		if len(cfg.Workflows) == 0 {
			fmt.Println("(没有定义工作流)")
			return nil
		}
		for _, spec := range cfg.Workflows {
			fmt.Printf("%s (kind=%s)\n", spec.Name, spec.Kind)
		}
	case "scheduler":
		if cfg.Scheduler == nil {
			fmt.Println("(未配置)")
		} else {
			fmt.Printf("enabled=%v, maxJobs=%d, jobTimeout=%ds\n",
				cfg.Scheduler.Enabled, cfg.Scheduler.MaxJobs, cfg.Scheduler.JobTimeout)
		}
	case "heartbeat":
		if cfg.Heartbeat == nil {
			fmt.Println("(未配置)")
		} else {
			fmt.Printf("enabled=%v, interval=%ds, timeout=%ds\n",
				cfg.Heartbeat.Enabled, cfg.Heartbeat.Interval, cfg.Heartbeat.Timeout)
		}
	case "async-task", "asyncTask":
		if cfg.AsyncTask == nil {
			fmt.Println("(未配置)")
		} else {
			fmt.Printf("enabled=%v, maxTasks=%d, taskTimeout=%ds\n",
				cfg.AsyncTask.Enabled, cfg.AsyncTask.MaxTasks, cfg.AsyncTask.TaskTimeout)
		}
	case "token-compression", "tokenCompression":
		if cfg.TokenCompression == nil {
			fmt.Println("(未配置)")
		} else {
			fmt.Printf("enabled=%v, strategy=%s, targetTokens=%d\n",
				cfg.TokenCompression.Enabled,
				cfg.TokenCompression.Strategy,
				cfg.TokenCompression.TargetTokens)
		}
	default:
		// Check if it's a model name
		if len(key) > 6 && key[:6] == "model." {
			modelName := key[6:]
			if model, ok := cfg.Models[modelName]; ok {
				fmt.Printf("%s/%s\n", model.Type, model.Model)
				return nil
			}
		}
		return fmt.Errorf("未知的配置键: %s\n使用 'af config list-keys' 查看可用的键", key)
	}
	return nil
}
