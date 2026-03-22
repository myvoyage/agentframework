// Agent Framework - AFCLI Standalone Application
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

	"AgentFramework/cmd/cli"
)

func main() {
	// 显示启动横幅
	printBanner()

	// 执行 CLI
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ 执行失败: %v\n", err)
		os.Exit(1)
	}
}

// printBanner 显示启动横幅
func printBanner() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        AgentFramework CLI - 命令行界面                     ║")
	fmt.Println("║                    Version 2.1.0                            ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  支持功能:                                                 ║")
	fmt.Println("║    • Agent 管理 - 列出、选择、运行                            ║")
	fmt.Println("║    • 工作流管理 - 列出、创建、执行                            ║")
	fmt.Println("║    • 技能管理 - 列出、启用、禁用、运行                       ║")
	fmt.Println("║    • 配置管理 - 查看、设置、验证                             ║")
	fmt.Println("║    • 文件操作 - 浏览、读取、写入、复制、删除                    ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  使用 'af --help' 查看完整命令帮助                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}
