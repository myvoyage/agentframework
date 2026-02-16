// Agent Framework - CLI Entry Point
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmdcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"AgentFramework/core"
	"AgentFramework/agent"
)

var (
	// Global flags
	configFile string
	modelName  string
	outputFormat string
	verbose bool

	// Global application instance
	app *core.Application
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentframework",
	Short: "AgentFramework - 高性能企业级 AI 代理框架",
	Long: `AgentFramework 是一个高性能、企业级的 AI 代理框架，为构建智能应用提供强大的基础设施。

支持多种 Agent 类型、工作流编排、技能系统和安全沙箱。`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Persistent flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "host.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "指定模型名称")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "输出格式 (table/json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	// Add subcommands
	addWorkflowCommands()
	addSkillCommands()
	addEnhancedSkillCommands()
	addConfigCommands()
	addFileCommands()
	addAgentCommands()

	// Execute command
	return rootCmd.Execute()
}

// initApplication initializes the core application
func initApplication() error {
	// Load config
	configPath := configFile
	if configPath == "" {
		configPath = "host.yaml"
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	hostCfg, err := agent.LoadHostConfigFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create model factory
	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

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
}

// rootContext returns the root context for operations
func rootContext() context.Context {
	if app != nil {
		return app.GetContext()
	}
	return context.Background()
}

// init sets up the CLI before execution
func init() {
	// Bind flags to viper
	viper.SetEnvPrefix("AGENTFRAMEWORK")
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Set global hooks
	rootCmd.PersistentPreRunE = preRunE
	rootCmd.PersistentPostRun = postRun
}

func init() {
	// Add completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "生成 shell 自动补全脚本",
		Long: `生成 shell 自动补全脚本。

要加载补全，请根据你使用的 shell 执行相应的命令：

Bash:
  $ source <(agentframework completion bash)

  # 为了永久生效，将补全脚本添加到 bash_completion:
  $ agentframework completion bash > /etc/bash_completion.d/agentframework

Zsh:
  # 如果补全功能还没启用，需要在 ~/.zshrc 中启用：
  autoload -U compinit; compinit
  $ source <(agentframework completion zsh)

  # 为了永久生效：
  $ agentframework completion zsh > "${fpath[1]}/_agentframework"

fish:
  $ agentframework completion fish | source

  # 为了永久生效：
  $ agentframework completion fish > ~/.config/fish/completions/agentframework.fish

PowerShell:
  PS> agentframework completion powershell | Out-String | Invoke-Expression

  # 为了永久生效，将输出添加到 PowerShell profile
`,
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
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("Compiler: %s\n", runtime.Compiler)
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	})
}
