// Agent Framework - Setup & Onboard Commands
// Based on OpenClaw Architecture: https://docs.openclaw.ai/zh-CN/cli
//
// Setup and Onboard provide interactive and non-interactive initialization:
//   - Setup: Interactive wizard for first-time setup
//   - Onboard: Fast non-interactive onboarding
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"AgentFramework/agent"
)

var (
	_wizardMode     bool
	_nonInteractive bool
	_setupMode      string
	_profileName    string
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "交互式初始化设置",
	Long: `通过交互式向导完成 AgentFramework 的首次设置。

支持的功能：
  • 创建配置文件
  • 初始化数据目录
  • 配置模型连接
  • 设置工作空间
  • 配置消息渠道

示例：
  af setup                # 启动交互式向导
  af setup --wizard       # 启动完整向导模式
  af setup --quick       # 快速设置（跳过可选步骤）`,
	RunE: runSetup,
}

// onboardCmd represents the onboard command
var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "快速配置向导（非交互式）",
	Long: `非交互式快速配置 AgentFramework，适合自动化部署和脚本使用。

支持的模式：
  • local: 本地模式（默认）
  • remote: 远程模式
  • hybrid: 混合模式

示例：
  af onboard --mode local --non-interactive    # 本地模式
  af onboard --mode remote --profile production  # 生产环境`,
	RunE: runOnboard,
}

func init() {
	// Setup flags
	setupCmd.Flags().BoolVar(&_wizardMode, "wizard", true, "启用完整向导模式")
	setupCmd.Flags().BoolVar(&_nonInteractive, "non-interactive", false, "非交互式模式")

	// Onboard flags
	onboardCmd.Flags().StringVar(&_setupMode, "mode", "local", "配置模式 (local/remote/hybrid)")
	onboardCmd.Flags().BoolVar(&_nonInteractive, "non-interactive", true, "非交互式模式")
	onboardCmd.Flags().StringVar(&_profileName, "profile", "default", "配置文件名")
}

func runSetup(cmd *cobra.Command, args []string) error {
	if _nonInteractive {
		return runOnboard(cmd, args)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     AgentFramework - 初始化向导                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Data directory
	fmt.Println("步骤 1/5: 数据目录配置")
	fmt.Println("────────────────────────────────────────────────────────────")
	defaultDataDir, _ := getDefaultDataDir()
	dataDir := promptInput(reader, "数据目录路径", defaultDataDir)
	fmt.Printf("  ✓ 数据目录: %s\n\n", dataDir)

	// Step 2: Model configuration
	fmt.Println("步骤 2/5: 模型配置")
	fmt.Println("────────────────────────────────────────────────────────────")
	modelType := promptInput(reader, "模型类型 (ollama/openai/anthropic)", "ollama")
	modelName := promptInput(reader, "模型名称", "llama3")
	fmt.Printf("  ✓ 模型类型: %s\n", modelType)
	fmt.Printf("  ✓ 模型名称: %s\n\n", modelName)

	// Step 3: Workspace
	fmt.Println("步骤 3/5: 工作空间配置")
	fmt.Println("────────────────────────────────────────────────────────────")
	workspace := promptInput(reader, "工作空间路径 (可选)", "")
	if workspace == "" {
		workspace = filepath.Join(dataDir, "workspace")
	}
	fmt.Printf("  ✓ 工作空间: %s\n\n", workspace)

	// Step 4: Channels (optional)
	fmt.Println("步骤 4/5: 消息渠道配置（可选）")
	fmt.Println("────────────────────────────────────────────────────────────")
	if _wizardMode {
		enableChannels := promptYesNo(reader, "是否配置消息渠道?", false)
		if enableChannels {
			channelType := promptInput(reader, "渠道类型 (telegram/lark/discord)", "")
			if channelType != "" {
				fmt.Printf("  ✓ 渠道类型: %s\n", channelType)
			}
		} else {
			fmt.Println("  ⊘ 跳过渠道配置")
		}
	}
	fmt.Println()

	// Step 5: Confirm
	fmt.Println("步骤 5/5: 确认配置")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("  模型: %s/%s\n", modelType, modelName)
	fmt.Printf("  工作空间: %s\n", workspace)
	fmt.Println()

	if !promptYesNo(reader, "确认创建配置?", true) {
		fmt.Println("已取消设置")
		return nil
	}

	// Create configuration
	configPath := filepath.Join(dataDir, "config.yaml")
	hostCfg := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  modelType,
				Model: modelName,
			},
		},
		SkillSystemDir: ".skills",
	}

	if err := agent.SaveHostConfigFile(configPath, hostCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create directories
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "conversations"),
		filepath.Join(dataDir, "workflows"),
		filepath.Join(dataDir, "skills"),
		filepath.Join(dataDir, "logs"),
		workspace,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     ✓ 设置完成！                                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n配置文件: %s\n", configPath)
	fmt.Printf("数据目录: %s\n", dataDir)
	fmt.Printf("\n快速开始:\n")
	fmt.Printf("  af config edit      # 编辑配置\n")
	fmt.Printf("  af agent list       # 列出 agents\n")
	fmt.Printf("  af --tui           # 启动 TUI 界面\n")

	return nil
}

func runOnboard(cmd *cobra.Command, args []string) error {
	fmt.Println("[Onboard] 快速配置模式")
	fmt.Println()

	dataDir, _ := getDefaultDataDir()

	// Create configuration
	configPath := filepath.Join(dataDir, "config.yaml")
	hostCfg := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
	}

	if err := agent.SaveHostConfigFile(configPath, hostCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
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

	fmt.Printf("✓ 配置创建完成: %s\n", configPath)
	fmt.Printf("✓ 数据目录: %s\n", dataDir)

	if _setupMode != "" && _setupMode != "local" {
		fmt.Printf("✓ 配置模式: %s\n", _setupMode)
	}

	return nil
}

// addSetupCommands adds setup and onboard commands
func addSetupCommands() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(onboardCmd)
}

// Helper functions

func promptInput(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptYesNo(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "Y/n"
	if !defaultValue {
		defaultStr = "y/N"
	}
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}
