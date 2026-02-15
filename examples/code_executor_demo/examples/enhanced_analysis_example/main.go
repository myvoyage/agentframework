// Agent Framework - Enhanced Code Analysis Example
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"AgentFramework/pkg/tools/sandbox/code"
)

// 示例 1: 基础代码分析
func basicAnalysisExample() {
	fmt.Println("=== 示例 1: 基础代码分析 ===\n")

	// 创建代码执行模块
	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "go"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 要分析的 Python 代码
	pythonCode := `
import requests

def fetch_data(url):
    response = requests.get(url)
    return response.text

data = fetch_data('http://example.com')
print(data)
`

	// 分析代码
	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", pythonCode)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	// 打印结果
	fmt.Printf("代码安全: %v\n", result.Safe)
	fmt.Printf("代码评分: %d/100\n", result.Score)
	fmt.Printf("发现问题数: %d\n", len(result.Issues))

	if len(result.Issues) > 0 {
		fmt.Println("\n发现的问题:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n改进建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println()
}

// 示例 2: 网络操作检测
func networkDetectionExample() {
	fmt.Println("=== 示例 2: 网络操作检测 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 包含多种网络操作的代码
	code := `
import socket
import requests
from urllib.request import urlopen

# HTTP 请求
response = requests.get('http://api.example.com/data')

# Socket 连接
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('example.com', 80))

# URL 打开
data = urlopen('https://example.com')
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("检测到网络操作: %d 个\n", len(result.NetworkOps))

	if len(result.NetworkOps) > 0 {
		fmt.Println("\n网络操作详情:")
		for i, op := range result.NetworkOps {
			fmt.Printf("\n  操作 %d:\n", i+1)
			fmt.Printf("    类型: %s\n", op.Type)
			fmt.Printf("    位置: 第 %d 行\n", op.Line)
			fmt.Printf("    代码: %s\n", op.Code)
			fmt.Printf("    说明: %s\n", op.Description)
		}
	}

	fmt.Println()
}

// 示例 3: 文件系统操作检测
func filesystemDetectionExample() {
	fmt.Println("=== 示例 3: 文件系统操作检测 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 包含文件系统操作的代码
	code := `
import os
import shutil

# 读取文件
with open('/etc/passwd', 'r') as f:
    data = f.read()

# 删除文件
os.remove('/tmp/test.txt')

# 修改权限
os.chmod('/tmp/file.txt', 0o777)

# 删除目录
shutil.rmtree('/tmp/data')
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("检测到文件系统操作: %d 个\n", len(result.FileSystemOps))

	if len(result.FileSystemOps) > 0 {
		fmt.Println("\n文件系统操作详情:")
		for i, op := range result.FileSystemOps {
			fmt.Printf("\n  操作 %d:\n", i+1)
			fmt.Printf("    类型: %s\n", op.Type)
			fmt.Printf("    位置: 第 %d 行\n", op.Line)
			fmt.Printf("    代码: %s\n", op.Code)
			fmt.Printf("    说明: %s\n", op.Description)
			if op.Path != "" {
				fmt.Printf("    路径: %s\n", op.Path)
			}
		}
	}

	fmt.Println()
}

// 示例 4: 代码质量检查
func qualityCheckExample() {
	fmt.Println("=== 示例 4: 代码质量检查 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 代码质量不佳的示例
	code := `
def f(x):
  if x>0:
   return x*2
  else:
   return x*3

def VeryLongFunctionNameThatExceedsReasonableLength(a,b,c):
    result=a+b+c
    return result

x=10
y=20
z=x+y
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("代码质量评分: %d/100\n", result.Score)
	fmt.Printf("检测到质量问题: %d 个\n", len(result.QualityIssues))

	if len(result.QualityIssues) > 0 {
		fmt.Println("\n质量问题详情:")

		// 按类型分组
		issuesByType := make(map[string][]code.QualityIssue)
		for _, issue := range result.QualityIssues {
			issuesByType[issue.Category] = append(issuesByType[issue.Category], issue)
		}

		for issueType, issues := range issuesByType {
			fmt.Printf("\n  %s 问题 (%d 个):\n", issueType, len(issues))
			for _, issue := range issues {
				fmt.Printf("    - 第 %d 行: %s\n", issue.Line, issue.Description)
			}
		}
	}

	fmt.Println()
}

// 示例 5: 自定义规则检测
func customRulesExample() {
	fmt.Println("=== 示例 5: 自定义规则检测 ===\n")

	// 创建完整配置，包含自定义规则
	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = "custom_rules.yaml"
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 包含自定义规则要检测的代码
	code := `
import pickle

# 使用 eval（不安全）
result = eval(user_input)

# 使用 pickle（可能不安全）
data = pickle.loads(untrusted_data)

# 使用 exec（不安全）
exec(code_from_user)
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("代码安全: %v\n", result.Safe)
	fmt.Printf("发现问题: %d 个\n", len(result.Issues))

	if len(result.Issues) > 0 {
		fmt.Println("\n检测到的问题:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	fmt.Println()
}

// 示例 6: 完整的代码审查流程
func completeReviewExample() {
	fmt.Println("=== 示例 6: 完整的代码审查流程 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 要审查的代码
	code := `
import requests
import json

def fetch_user_data(user_id):
    url = f"http://api.example.com/users/{user_id}"
    response = requests.get(url)
    
    if response.status_code == 200:
        return json.loads(response.text)
    else:
        return None

def process_data(data):
    result = []
    for item in data:
        if item['active']:
            result.append(item)
    return result

user_data = fetch_user_data(123)
if user_data:
    active_users = process_data(user_data)
    print(f"Found {len(active_users)} active users")
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	// 生成审查报告
	fmt.Println("代码审查报告")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Printf("\n总体评分: %d/100\n", result.Score)
	fmt.Printf("安全状态: %v\n", result.Safe)

	// 安全问题
	if len(result.Issues) > 0 {
		fmt.Println("\n🔴 安全问题:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	// 网络操作
	if len(result.NetworkOps) > 0 {
		fmt.Println("\n🌐 网络操作:")
		for i, op := range result.NetworkOps {
			fmt.Printf("  %d. [%s] 第 %d 行: %s\n", i+1, op.Type, op.Line, op.Description)
		}
	}

	// 质量问题
	if len(result.QualityIssues) > 0 {
		fmt.Println("\n⚠️  质量问题:")
		for i, issue := range result.QualityIssues {
			fmt.Printf("  %d. [%s] 第 %d 行: %s\n", i+1, issue.Category, issue.Line, issue.Description)
		}
	}

	// 改进建议
	if len(result.Suggestions) > 0 {
		fmt.Println("\n💡 改进建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	// 总结
	fmt.Println("\n" + string(make([]byte, 50)))
	if result.Safe && result.Score >= 80 {
		fmt.Println("✅ 代码质量良好，可以使用")
	} else if result.Safe && result.Score >= 60 {
		fmt.Println("⚠️  代码可用，但建议改进")
	} else {
		fmt.Println("❌ 代码存在问题，需要修复")
	}

	fmt.Println()
}

// 示例 7: JSON 格式输出
func jsonOutputExample() {
	fmt.Println("=== 示例 7: JSON 格式输出 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	code := `
import requests
response = requests.get('http://example.com')
print(response.text)
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	// 转换为 JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
	}

	fmt.Println("分析结果 (JSON 格式):")
	fmt.Println(string(jsonData))
	fmt.Println()
}

func main() {
	fmt.Println("代码执行模块 - 增强代码分析示例")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// 运行所有示例
	basicAnalysisExample()
	networkDetectionExample()
	filesystemDetectionExample()
	qualityCheckExample()
	customRulesExample()
	completeReviewExample()
	jsonOutputExample()

	fmt.Println("所有示例运行完成！")
}
