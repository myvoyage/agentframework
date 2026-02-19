// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Agent Framework - SWEAgent 使用示例
// 演示如何使用软件工程智能体进行代码审查、重构和 Git 操作

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	ctx := context.Background()

	// 示例1: 创建 SWEAgent 并进行代码审查
	fmt.Println("=== SWEAgent 代码审查示例 ===")
	codeReviewExample(ctx)

	// 示例2: 代码重构
	fmt.Println("\n=== SWEAgent 代码重构示例 ===")
	refactoringExample(ctx)

	// 示例3: Git 操作
	fmt.Println("\n=== SWEAgent Git 操作示例 ===")
	gitOperationsExample(ctx)

	// 示例4: 代码库分析
	fmt.Println("\n=== SWEAgent 代码库分析示例 ===")
	codebaseAnalysisExample(ctx)

	// 示例5: 测试生成
	fmt.Println("\n=== SWEAgent 测试生成示例 ===")
	testGenerationExample(ctx)
}

// codeReviewExample 代码审查示例
func codeReviewExample(ctx context.Context) {
	// 注意: 这需要配置实际的 LLM 模型
	// 这里展示 API 使用方法

	// 模拟代码文件路径
	filePath := "examples/swe_demo/sample.go"
	content := `package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func main() {
	result := add(1, 2)
	fmt.Println(result)
}`

	fmt.Printf("审查文件: %s\n", filePath)
	fmt.Printf("代码内容:\n%s\n\n", content)

	// 实际使用需要创建真实的模型实例
	// sweAgent, _ := agent.NewSWEAgent(ctx, agent.SWEAgentConfig{
	//     Name: "code-reviewer",
	//     Model: model, // 需要实际的模型
	//     RepoPath: ".",
	// })
	//
	// result, _ := sweAgent.ReviewCode(ctx, filePath, content)
	// fmt.Printf("审查结果:\n")
	// fmt.Printf("- 质量评分: %d/10\n", result.QualityScore)
	// fmt.Printf("- 问题数量: %d\n", len(result.Issues))
	// for _, issue := range result.Issues {
	//     fmt.Printf("  * [%s] %s: %s\n", issue.Severity, issue.Description, issue.Suggestion)
	// }

	fmt.Println("使用 SWEAgent.ReviewCode() 方法进行代码审查")
	fmt.Println("返回结果包含: 质量评分、问题列表、性能建议、安全检查等")
}

// refactoringExample 代码重构示例
func refactoringExample(ctx context.Context) {
	filePath := "examples/swe_demo/before_refactor.go"
	content := `func processData(items []string) map[string]int {
	result := make(map[string]int)
	for i := 0; i < len(items); i++ {
		item := items[i]
		count := 0
		for j := 0; j < len(items); j++ {
			if items[j] == item {
				count++
			}
		}
		result[item] = count
	}
	return result
}`

	fmt.Printf("重构文件: %s\n", filePath)
	fmt.Printf("原始代码:\n%s\n\n", content)

	fmt.Println("使用 SWEAgent.RefactorCode() 方法进行代码重构")
	fmt.Println("支持的重构类型:")
	fmt.Println("  - performance: 性能优化")
	fmt.Println("  - readability: 可读性改进")
	fmt.Println("  - maintainability: 可维护性提升")
	fmt.Println("  - design-pattern: 应用设计模式")
}

// gitOperationsExample Git 操作示例
func gitOperationsExample(ctx context.Context) {
	fmt.Println("SWEAgent 支持的 Git 操作:")
	fmt.Println()

	operations := []struct {
		name string
		desc string
	}{
		{"status", "查看工作区状态"},
		{"diff", "查看文件差异"},
		{"log", "查看提交历史"},
		{"branch", "列出所有分支"},
		{"commit", "提交更改"},
		{"create_branch", "创建新分支"},
		{"checkout", "切换分支"},
	}

	for _, op := range operations {
		fmt.Printf("  %-15s - %s\n", op.name, op.desc)
	}

	fmt.Println("\n示例用法:")
	fmt.Println("  result, _ := sweAgent.GitOperations(ctx, \"status\", nil)")
	fmt.Println("  result, _ := sweAgent.GitOperations(ctx, \"commit\", map[string]string{")
	fmt.Println("    \"message\": \"feat: add new feature\",")
	fmt.Println("    \"files\": \"*.go\",")
	fmt.Println("  })")
}

// codebaseAnalysisExample 代码库分析示例
func codebaseAnalysisExample(ctx context.Context) {
	// 获取当前目录
	cwd, _ := os.Getwd()

	fmt.Printf("分析代码库: %s\n\n", cwd)

	fmt.Println("使用 SWEAgent.AnalyzeCodebase() 方法进行代码库分析")
	fmt.Println("返回结果包含:")
	fmt.Println("  - 总文件数统计")
	fmt.Println("  - 编程语言分布")
	fmt.Println("  - 项目架构分析")
	fmt.Println("  - 代码组织评估")
	fmt.Println("  - 依赖关系分析")
	fmt.Println("  - 技术栈识别")
	fmt.Println("  - 潜在问题识别")
	fmt.Println("  - 改进建议")
}

// testGenerationExample 测试生成示例
func testGenerationExample(ctx context.Context) {
	filePath := "calculator.go"
	content := `package calculator

func Add(a, b float64) float64 {
	return a + b
}

func Subtract(a, b float64) float64 {
	return a - b
}

func Multiply(a, b float64) float64 {
	return a * b
}

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}`

	fmt.Printf("为文件生成测试: %s\n", filePath)
	fmt.Printf("代码内容:\n%s\n\n", content)

	fmt.Println("使用 SWEAgent.GenerateTests() 方法生成测试")
	fmt.Println("支持多个测试框架:")
	fmt.Println("  - Go: testing 包")
	fmt.Println("  - Python: pytest, unittest")
	fmt.Println("  - JavaScript: Jest, Mocha")
	fmt.Println("  - Java: JUnit")
	fmt.Println("\n生成的测试包含:")
	fmt.Println("  - 单元测试")
	fmt.Println("  - 边界条件测试")
	fmt.Println("  - 错误处理测试")
	fmt.Println("  - Mock 数据")
}

// init 创建示例文件
func init() {
	// 创建示例目录
	dirs := []string{
		"examples/swe_demo",
		"examples/swe_demo/code_samples",
		"examples/swe_demo/tests",
	}

	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(".", dir), 0755)
	}
}
