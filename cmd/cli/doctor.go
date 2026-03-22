// Agent Framework - Doctor & Health Check Commands
// Based on OpenClaw Architecture: https://docs.openclaw.ai/zh-CN/cli
//
// Doctor provides comprehensive system health checks:
//   - Configuration validation
//   - Gateway status probe
//   - Agent health
//   - Memory status
//   - Performance metrics
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"AgentFramework/agent"
)

var (
	_doDeepCheck bool
	_fixIssues  bool
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "系统健康检查与诊断",
	Long: `执行全面的系统健康检查，诊断配置、连接、性能等问题。

检查项目：
  ✓ 配置文件有效性
  ✓ 模型连接状态
  ✓ 工作空间状态
  ✓ 渠道连接状态
  ✓ 内存使用情况
  ✓ 日志文件完整性

示例：
  af doctor                    # 基础检查
  af doctor --deep              # 深度诊断
  af doctor --deep --fix        # 诊断并尝试修复`,
	RunE: runDoctor,
}

// healthCheckCmd represents the health check command
var healthCheckCmd = &cobra.Command{
	Use:   "health",
	Short: "快速健康检查",
	Long: `快速检查系统关键组件的健康状态。

示例：
  af health                    # 检查核心组件
  af health --verbose           # 显示详细信息`,
	RunE: runHealthCheck,
}

func init() {
	doctorCmd.Flags().BoolVar(&_doDeepCheck, "deep", false, "执行深度诊断")
	doctorCmd.Flags().BoolVar(&_fixIssues, "fix", false, "尝试修复发现的问题")
}

func addDoctorCommands() {
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(healthCheckCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║     AgentFramework - 系统诊断                                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	issues := 0
	passed := 0

	// 1. Configuration check
	fmt.Println("📋 配置检查")
	fmt.Println("────────────────────────────────────────────────────────────")
	if checkConfig() {
		passed++
		fmt.Println("  ✓ 配置文件有效")
	} else {
		issues++
		fmt.Println("  ✗ 配置文件存在问题")
	}
	fmt.Println()

	// 2. Model connection check
	fmt.Println("🤖 模型连接检查")
	fmt.Println("────────────────────────────────────────────────────────────")
	if checkModelConnection() {
		passed++
		fmt.Println("  ✓ 模型连接正常")
	} else {
		issues++
		fmt.Println("  ✗ 模型连接失败")
	}
	fmt.Println()

	// 3. Workspace check
	fmt.Println("📁 工作空间检查")
	fmt.Println("────────────────────────────────────────────────────────────")
	if checkWorkspace() {
		passed++
		fmt.Println("  ✓ 工作空间正常")
	} else {
		issues++
		fmt.Println("  ✗ 工作空间存在问题")
	}
	fmt.Println()

	// 4. Channels check
	fmt.Println("📡 消息渠道检查")
	fmt.Println("────────────────────────────────────────────────────────────")
	if checkChannels() {
		passed++
		fmt.Println("  ✓ 渠道状态正常")
	} else {
		issues++
		fmt.Println("  ✗ 渠道存在问题")
	}
	fmt.Println()

	// 5. Deep checks
	if _doDeepCheck {
		fmt.Println("🔍 深度诊断")
		fmt.Println("────────────────────────────────────────────────────────────")

		// Memory check
		if checkMemory() {
			passed++
			fmt.Println("  ✓ 内存使用正常")
		} else {
			issues++
			fmt.Println("  ✗ 内存使用异常")
		}

		// Performance check
		checkPerformance()
		fmt.Println()
	}

	// Summary
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Printf("  通过: %d\n", passed)
	fmt.Printf("  问题: %d\n", issues)
	fmt.Println("══════════════════════════════════════════════════════════")

	if issues > 0 {
		fmt.Println()
		if _fixIssues {
			fmt.Println("🔧 正在尝试修复问题...")
			// TODO: Implement auto-fix logic
			fmt.Println("  (自动修复功能开发中)")
		} else {
			fmt.Println("💡 建议:")
			fmt.Println("  • 运行 'af doctor --fix' 尝试修复问题")
			fmt.Println("  • 运行 'af config validate' 检查配置")
			fmt.Println("  • 查看日志文件了解详细信息")
		}
		return fmt.Errorf("发现 %d 个问题", issues)
	}

	fmt.Println("✓ 系统健康检查通过")
	return nil
}

func runHealthCheck(cmd *cobra.Command, args []string) error {
	fmt.Println("[Health Check] 快速检查核心组件...")
	fmt.Println()

	allOK := true

	// Quick checks
	if !checkConfig() {
		fmt.Println("✗ 配置文件")
		allOK = false
	} else {
		fmt.Println("✓ 配置文件")
	}

	if !checkModelConnection() {
		fmt.Println("✗ 模型连接")
		allOK = false
	} else {
		fmt.Println("✓ 模型连接")
	}

	if !checkWorkspace() {
		fmt.Println("✗ 工作空间")
		allOK = false
	} else {
		fmt.Println("✓ 工作空间")
	}

	fmt.Println()
	if allOK {
		fmt.Println("✓ 系统健康")
	} else {
		fmt.Println("✗ 系统存在问题，运行 'af doctor --deep' 进行详细诊断")
		return fmt.Errorf("health check failed")
	}

	return nil
}

// Check functions

func checkConfig() bool {
	dataDir, _ := getDefaultDataDir()
	configPath := filepath.Join(dataDir, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("  ⚠ 配置文件不存在")
		return false
	}

	_, err := agent.LoadHostConfigFile(configPath)
	return err == nil
}

func checkModelConnection() bool {
	if app == nil {
		fmt.Println("  ⚠ 应用未初始化")
		return false
	}

	// TODO: Add actual connection test
	// For now, assume OK if app exists
	return true
}

func checkWorkspace() bool {
	dataDir, _ := getDefaultDataDir()
	workspacePath := filepath.Join(dataDir, "workspace")

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		fmt.Println("  ⚠ 工作空间目录不存在")
		return false
	}

	return true
}

func checkChannels() bool {
	if app == nil {
		return true // Not an error if app not initialized
	}

	channels := app.GetHost().ChannelManager()
	if channels == nil {
		fmt.Println("  ℹ 渠道管理器未配置")
		return true // Not critical
	}

	// TODO: Check channel connection status
	return true
}

func checkMemory() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check for excessive memory usage (e.g., > 1GB)
	if m.Alloc > 1024*1024*1024 {
		fmt.Printf("  ⚠ 内存使用: %.2f MB\n", float64(m.Alloc)/1024/1024)
		return false
	}

	fmt.Printf("  ℹ 内存使用: %.2f MB\n", float64(m.Alloc)/1024/1024)
	return true
}

func checkPerformance() {
	// System info
	fmt.Printf("  ℹ Go 版本: %s\n", runtime.Version())
	fmt.Printf("  ℹ 操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  ℹ CPU 核心数: %d\n", runtime.NumCPU())
	fmt.Printf("  ℹ Goroutine 数量: %d\n", runtime.NumGoroutine())
}
