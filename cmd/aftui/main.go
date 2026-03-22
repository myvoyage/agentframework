// Agent Framework - AFTUI Standalone Application
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"AgentFramework/cmd/tui"
	"AgentFramework/agent"
)

func main() {
	// 显示启动横幅
	printBanner()

	// 启动 TUI
	if err := runTUI(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ 启动失败: %v\n", err)
		os.Exit(1)
	}
}

// printBanner 显示启动横幅
func printBanner() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         AgentFramework TUI - 终端用户界面                 ║")
	fmt.Println("║                   Version 2.1.0                            ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  基于 Memoh 架构设计                                        ║")
	fmt.Println("║  支持 Agents、工作流、技能管理                                ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  快捷键: Tab=切换视图  Ctrl+R=刷新  Q=退出                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// runTUI 启动 TUI 应用
func runTUI() error {
	// 创建默认配置
	hostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
	}

	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	// 启动 TUI
	fmt.Println("⏳ 正在初始化...")

	err := tui.Run(hostConfig, modelFactory)
	if err != nil {
		return fmt.Errorf("TUI 启动失败: %w", err)
	}

	return nil
}
