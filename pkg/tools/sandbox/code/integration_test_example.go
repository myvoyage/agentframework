// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"encoding/json"
	"fmt"
)

// ExampleUsage demonstrates how to use the code_exec_supported_languages tool
func ExampleUsage() {
	// Create module with configuration
	config := CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "bash", "go"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		fmt.Printf("Failed to create module: %v\n", err)
		return
	}
	defer module.Close()

	// Get all tools
	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		fmt.Printf("Failed to get tools: %v\n", err)
		return
	}

	// Find the code_exec_supported_languages tool
	var supportedLangsTool *codeExecSupportedLanguagesTool
	for _, tool := range tools {
		if t, ok := tool.(*codeExecSupportedLanguagesTool); ok {
			supportedLangsTool = t
			break
		}
	}

	if supportedLangsTool == nil {
		fmt.Println("Tool not found")
		return
	}

	// Get tool info
	info, err := supportedLangsTool.Info(ctx)
	if err != nil {
		fmt.Printf("Failed to get tool info: %v\n", err)
		return
	}

	fmt.Printf("Tool Name: %s\n", info.Name)
	fmt.Printf("Tool Description: %s\n", info.Desc)

	// Call the tool
	result, err := supportedLangsTool.InvokableRun(ctx, "{}")
	if err != nil {
		fmt.Printf("Failed to run tool: %v\n", err)
		return
	}

	// Parse the result
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		fmt.Printf("Failed to parse result: %v\n", err)
		return
	}

	fmt.Printf("Success: %v\n", response["success"])
	fmt.Printf("Supported Languages: %v\n", response["languages"])
}
