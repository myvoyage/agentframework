// Agent Framework - CLI Entry Point
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// 重构后的 CLI - 缩短命令名，优化结构，添加 TUI 支持

package cmdcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"AgentFramework/core"
	"AgentFramework/agent"
	"AgentFramework/pkg/local"
)

var (
	// Global flags
	configFile   string
	modelName    string
	outputFormat string
	verbose      bool
	_TIMEOUT     time.Duration
	_WATCH       bool

	// Global application instance
	app   *core.Application
	store local.Store
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "af",
	Short: "AgentFramework - 高性能企业级 AI 代理框架",
	Long: `AgentFramework 是一个高性能、企业级的 AI 代理框架，为构建智能应用提供强大的基础设施。

支持多种 Agent 类型、工作流编排、技能系统和安全沙箱。

快速开始:
  af tui              启动交互式 TUI 界面
  af agent list       列出可用 agents
  af agent chat       与 agent 对话
  af workflow list    管理工作流
  af skill list       管理技能`,
	Version:      "1.2.0",
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Persistent flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "指定模型名称")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "输出格式 (table/json/yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().DurationVar(&_TIMEOUT, "timeout", 30*time.Second, "操作超时时间")
	rootCmd.PersistentFlags().BoolVar(&_WATCH, "watch", false, "监视模式（持续更新）")

	// Add subcommands
	addWorkflowCommands()
	addSkillCommands()
	addConfigCommands()
	addFileCommands()
	addAgentCommands()
	addTUICommands()

	// Execute command
	return rootCmd.Execute()
}

// initApplication initializes the core application
func initApplication() error {
	// Load config from file or use defaults
	var hostCfg *agent.HostConfig
	var err error

	if configFile != "" {
		hostCfg, err = agent.LoadHostConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Try to load from default locations
		// 1. Local store
		// 2. ~/.agentframework/config.yaml
		// 3. Use defaults
		store, err = local.GetDefaultStore()
		if err == nil {
			var savedCfg agent.HostConfig
			if err := store.GetConfig("host", &savedCfg); err == nil {
				hostCfg = &savedCfg
			}
		}
	}

	// Use defaults if no config found
	if hostCfg == nil {
		hostCfg = &agent.HostConfig{
			Models: map[string]agent.ModelConfig{
				"default": {
					Type:  "ollama",
					Model: "llama3",
				},
			},
			SkillSystemDir: ".skills",
		}
	}

	// Override model if specified
	if modelName != "" {
		if hostCfg.Models == nil {
			hostCfg.Models = make(map[string]agent.ModelConfig)
		}
		hostCfg.Models["default"] = agent.ModelConfig{
			Type:  "ollama",
			Model: modelName,
		}
	}

	// Create model factory
	modelFactory := agent.NewModelFactoryWithConfig(hostCfg.Models["default"])

	// Create application
	ctx := rootContext()
	app, err = core.NewApplication(ctx, hostCfg, modelFactory, nil)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	// Initialize application
	if err := app.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// Initialize managers
	app.GetWorkflowManager().Init(ctx)
	app.GetFileExplorer().Init(ctx)

	return nil
}

// PreRunE is a persistent pre-run hook for all commands
func preRunE(cmd *cobra.Command, args []string) error {
	// Initialize application on first use
	if app == nil {
		if err := initApplication(); err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
	}
	return nil
}

// PostRun is a persistent post-run hook for all commands
func postRun(cmd *cobra.Command, args []string) {
	// Cleanup application if needed
	if app != nil {
		app.Shutdown(rootContext())
	}
	if store != nil {
		store.Close()
	}
}

// rootContext returns the root context for operations
func rootContext() context.Context {
	if app != nil {
		return app.GetContext()
	}
	return context.Background()
}

// getStore returns the local store, initializing if necessary
func getStore() (local.Store, error) {
	if store == nil {
		var err error
		store, err = local.GetDefaultStore()
		if err != nil {
			return nil, fmt.Errorf("failed to get store: %w", err)
		}
	}
	return store, nil
}

// init sets up the CLI before execution
func init() {
	// Bind flags to viper
	viper.SetEnvPrefix("AF")
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.AutomaticEnv()

	// Set global hooks
	rootCmd.PersistentPreRunE = preRunE
	rootCmd.PersistentPostRun = postRun
}

func init() {
	// Add TUI command (high priority)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "启动交互式 TUI 界面",
		Long:  `启动基于 Bubble Tea 的交互式终端用户界面，提供可视化的 Agent、工作流和技能管理。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Import and run TUI
			return runTUI()
		},
	})

	// Add completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "生成 shell 自动补全脚本",
		Long: `生成 shell 自动补全脚本。

要加载补全，请根据你使用的 shell 执行相应的命令：

Bash:
  $ source <(af completion bash)
  $ af completion bash > /etc/bash_completion.d/af

Zsh:
  $ af completion zsh > "${fpath[1]}/_af"

Fish:
  $ af completion fish | source
  $ af completion fish > ~/.config/fish/completions/af.fish

PowerShell:
  PS> af completion powershell | Out-String | Invoke-Expression`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			switch args[0] {
			case "bash":
				err = cmd.GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				err = cmd.GenZshCompletion(os.Stdout)
			case "fish":
				err = cmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type: %s", args[0])
			}
			return err
		},
	})

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long:  `显示 AgentFramework 的版本信息`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("AgentFramework CLI v%s\n", rootCmd.Version)
			fmt.Printf("Build: %s\n", getBuildInfo())
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

			if verbose {
				fmt.Println("\n组件信息:")
				if app != nil {
					fmt.Printf("  Agents: %d\n", len(app.GetHost().ListAgents()))
				}
			}
		},
	})

	// Add config init command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "初始化配置",
		Long:  `初始化 AgentFramework 配置，创建默认配置文件和目录结构。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	})
}

// getBuildInfo returns build information
func getBuildInfo() string {
	// Can be replaced with actual build info using ldflags
	return "devel"
}

// runTUI 启动 TUI 界面
func runTUI() error {
	// This will be implemented by importing the TUI package
	// For now, we'll return an error indicating the feature
	return fmt.Errorf("TUI feature requires compiling with the tui build tag\n" +
		"Run: go build -tags=tui ./cmd/tui\n" +
		"Or use: af tui in standalone mode")
}

// initConfig 初始化配置
func initConfig() error {
	// Create default data directory
	dataDir, err := getDefaultDataDir()
	if err != nil {
		return err
	}

	// Create directories
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "conversations"),
		filepath.Join(dataDir, "workflows"),
		filepath.Join(dataDir, "skills"),
		filepath.Join(dataDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default config
	configPath := filepath.Join(dataDir, "config.yaml")
	defaultConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
	}

	if err := agent.SaveHostConfigFile(configPath, defaultConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ 配置初始化完成\n")
	fmt.Printf("  配置文件: %s\n", configPath)
	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("\n使用方法:\n")
	fmt.Printf("  af config edit      # 编辑配置\n")
	fmt.Printf("  af agent list       # 列出 agents\n")
	fmt.Printf("  af tui              # 启动 TUI 界面\n")

	return nil
}

// getDefaultDataDir 获取默认数据目录
func getDefaultDataDir() (string, error) {
	// Check environment variable
	if dir := os.Getenv("AGENT_FRAMEWORK_DATA_DIR"); dir != "" {
		return dir, nil
	}

	// Use user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".agentframework"), nil
}

// addTUICommands adds TUI-related commands
func addTUICommands() {
	// TUI commands can be added here in the future
	// For example: tui-agent, tui-workflow, etc.
}
