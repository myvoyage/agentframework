// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Agent Framework - 技能系统代码验证
// 独立验证程序，不依赖项目其他部分

package main

import (
	"fmt"
	"os"
	"strings"
)

func CheckFiles() {
	fmt.Println("=== AgentFramework 技能系统代码验证 ===\n")

	passed := 0
	failed := 0

	// 检查核心文件是否存在
	files := map[string]string{
		"agent/skills/registry.go":           "技能注册表",
		"agent/skills/definition.go":         "定义管理器",
		"agent/skills/examples.go":           "示例模板库",
		"agent/skills/loader.go":             "渐进式加载器",
		"agent/skills/enhanced_executor.go": "增强执行器",
		"agent/skills_integration.go":       "系统集成层",
		"agent/skills_cache.go":             "多级缓存",
		"agent/skills_pool.go":              "连接池",
	}

	fmt.Println("检查核心文件:")
	for file, desc := range files {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("✅ %-40s 存在\n", desc)
			passed++
		} else {
			fmt.Printf("❌ %-40s 缺失: %v\n", desc, err)
			failed++
		}
	}

	// 检查技能定义文件
	fmt.Println("\n检查技能定义:")
	skillFiles := map[string]string{
		"agent/skills/definitions/http_request/SKILL.yaml":  "HTTP请求技能",
		"agent/skills/definitions/file_operation/SKILL.yaml": "文件操作技能",
		"agent/skills/definitions/api_call/SKILL.yaml":      "API调用技能",
	}

	for file, desc := range skillFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("✅ %-40s 存在\n", desc)
			passed++
		} else {
			fmt.Printf("❌ %-40s 缺失: %v\n", desc, err)
			failed++
		}
	}

	// 检查演示程序
	fmt.Println("\n检查演示程序:")
	demoFiles := map[string]string{
		"examples/skills_demo/main.go":                    "基础使用示例",
		"examples/enhanced_executor_demo/main.go":          "执行器演示",
		"examples/integration_demo/main.go":               "集成演示",
		"examples/validate_skills/main.go":                "系统验证",
	}

	for file, desc := range demoFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("✅ %-40s 存在\n", desc)
			passed++
		} else {
			fmt.Printf("❌ %-40s 缺失: %v\n", desc, err)
			failed++
		}
	}

	// 检查文档
	fmt.Println("\n检查文档:")
	docs := map[string]string{
		"SKILL_USER_GUIDE.md":              "用户指南",
		"SKILL_COMPARISON_ANALYSIS.md":     "深度分析",
		"SKILL_ENHANCEMENT_GUIDE.md":       "实施指南",
		"SKILLS_API_QUICK_REFERENCE.md":    "API参考",
		"PROJECT_COMPLETION_REPORT.md":    "项目报告",
		"FINAL_COMPLETION_REPORT.md":      "最终报告",
	}

	for file, desc := range docs {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("✅ %-40s 存在\n", desc)
			passed++
		} else {
			fmt.Printf("❌ %-40s 缺失: %v\n", desc, err)
			failed++
		}
	}

	// 统计代码行数
	fmt.Println("\n代码统计:")
	totalLines := 0
	coreFiles := []string{
		"agent/skills/registry.go",
		"agent/skills/definition.go",
		"agent/skills/examples.go",
		"agent/skills/loader.go",
		"agent/skills/enhanced_executor.go",
		"agent/skills_integration.go",
		"agent/skills_cache.go",
		"agent/skills_pool.go",
	}

	for _, file := range coreFiles {
		lines, err := countLines(file)
		if err == nil {
			fmt.Printf("  %-40s %5d 行\n", file, lines)
			totalLines += lines
		}
	}

	fmt.Printf("\n  总计:                                   %5d 行\n", totalLines)

	// 总结
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("验证结果: %d 通过, %d 失败\n", passed, failed)
	if failed == 0 {
		fmt.Println("🎉 所有检查通过！技能系统文件完整。")
		fmt.Println("\n注意：由于项目依赖问题，需要先解决 go.mod 才能编译运行。")
		fmt.Println("建议步骤：")
		fmt.Println("  1. 删除 go.mod 和 go.sum")
		fmt.Println("  2. 运行 go mod init")
		fmt.Println("  3. 运行 go mod tidy")
	} else {
		fmt.Printf("⚠️  有 %d 个文件缺失，请检查。\n", failed)
	}
	fmt.Println(strings.Repeat("=", 60))
}

func countLines(filename string) (int, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	return lines, nil
}
