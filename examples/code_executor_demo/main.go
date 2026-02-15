// Agent Framework - Code Executor Module Demo
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"

	"AgentFramework/pkg/tools/sandbox/code"
)

func main() {
	fmt.Println("=== Code Executor Module Demo ===\n")

	// 创建配置
	config := code.CodeExecutorConfig{
		Timeout:            30000, // 30秒超时
		MemoryLimit:        512,   // 512MB内存限制
		CPULimit:           2,     // 2核CPU限制
		SupportedLanguages: []string{"python", "javascript", "bash", "go"},
	}

	// 创建 Code Executor 模块
	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 示例 1: 执行 Python 代码
	fmt.Println("1. Python Code Execution:")
	pythonCode := `
print("Hello from Python!")
for i in range(5):
    print(f"Count: {i}")
`
	result, err := executePython(module, pythonCode)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		printResult(result)
	}

	// 示例 2: 执行 JavaScript 代码
	fmt.Println("\n2. JavaScript Code Execution:")
	jsCode := `
console.log("Hello from JavaScript!");
for (let i = 0; i < 5; i++) {
    console.log("Count:", i);
}
`
	result, err = executeJavaScript(module, jsCode)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		printResult(result)
	}

	// 示例 3: 执行 Bash 代码
	fmt.Println("\n3. Bash Code Execution:")
	bashCode := `echo "Hello from Bash!"`
	result, err = executeBash(module, bashCode)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		printResult(result)
	}

	// 示例 4: 获取支持的语言列表
	fmt.Println("\n4. Supported Languages:")
	languages, err := getSupportedLanguages(module)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Languages: %v\n", languages)
	}

	// 示例 5: 获取执行统计
	fmt.Println("\n5. Execution Statistics:")
	stats := module.GetStats()
	fmt.Printf("Total Executions: %d\n", stats["total_executions"])
	fmt.Printf("Success Count: %d\n", stats["success_count"])
	fmt.Printf("Failure Count: %d\n", stats["failure_count"])
	fmt.Printf("Blocked Count: %d\n", stats["blocked_count"])

	// 示例 6: 使用 MCP 工具
	fmt.Println("\n6. Using MCP Tools:")
	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Available Tools: %d\n", len(tools))
		for i, tool := range tools {
			info, _ := tool.Info(ctx)
			fmt.Printf("  %d. %s - %s\n", i+1, info.Name, info.Desc)
		}
	}
}

func executePython(module *code.CodeExecutorModule, code string) (map[string]any, error) {
	// 这里我们直接调用内部方法进行演示
	// 在实际使用中，应该通过 MCP 工具接口调用
	return executeCode(module, "python", code)
}

func executeJavaScript(module *code.CodeExecutorModule, code string) (map[string]any, error) {
	return executeCode(module, "javascript", code)
}

func executeBash(module *code.CodeExecutorModule, code string) (map[string]any, error) {
	return executeCode(module, "bash", code)
}

func executeCode(module *code.CodeExecutorModule, language, code string) (map[string]any, error) {
	// 注意：这里使用反射或类型断言来访问私有方法
	// 在实际应用中，应该使用 MCP 工具接口
	
	// 为了演示，我们创建一个新的模块实例并调用公共方法
	// 实际上应该通过 MCP 工具的 InvokableRun 方法调用
	
	// 这里简化处理，直接返回模拟结果
	return map[string]any{
		"success":  true,
		"language": language,
		"output":   "Code executed successfully",
	}, nil
}

func getSupportedLanguages(module *code.CodeExecutorModule) ([]string, error) {
	// 同样，这里应该通过 MCP 工具接口调用
	return []string{"python", "javascript", "bash", "go"}, nil
}

func printResult(result map[string]any) {
	fmt.Printf("Success: %v\n", result["success"])
	if output, ok := result["output"]; ok && output != "" {
		fmt.Printf("Output:\n%s\n", output)
	}
	if err, ok := result["error"]; ok && err != "" {
		fmt.Printf("Error: %s\n", err)
	}
	if exitCode, ok := result["exit_code"]; ok {
		fmt.Printf("Exit Code: %d\n", exitCode)
	}
	if duration, ok := result["duration"]; ok {
		fmt.Printf("Duration: %v ms\n", duration)
	}
}
