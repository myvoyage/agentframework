// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"encoding/json"
	"testing"
)

// TestCodeExecSetModeTool 测试设置执行模式工具
func TestCodeExecSetModeTool(t *testing.T) {
	// 创建模块
	config := CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "local",
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 创建工具
	tool := &codeExecSetModeTool{module: module}

	// 测试 Info 方法
	ctx := context.Background()
	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}

	if info.Name != "code_exec_set_mode" {
		t.Errorf("Expected name 'code_exec_set_mode', got '%s'", info.Name)
	}

	// 测试设置为 container 模式
	input := `{"mode": "container"}`
	output, err := tool.InvokableRun(ctx, input)
	if err != nil {
		t.Fatalf("InvokableRun() failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if result["mode"].(string) != "container" {
		t.Errorf("Expected mode 'container', got '%s'", result["mode"])
	}

	// 验证模式已更改
	if module.config.ExecutionMode != "container" {
		t.Errorf("Expected execution mode 'container', got '%s'", module.config.ExecutionMode)
	}

	// 测试设置为 auto 模式
	input = `{"mode": "auto"}`
	output, err = tool.InvokableRun(ctx, input)
	if err != nil {
		t.Fatalf("InvokableRun() failed: %v", err)
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["mode"].(string) != "auto" {
		t.Errorf("Expected mode 'auto', got '%s'", result["mode"])
	}

	// 测试无效模式
	input = `{"mode": "invalid"}`
	_, err = tool.InvokableRun(ctx, input)
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
}

// TestCodeExecContainerStatusTool 测试容器状态查询工具
func TestCodeExecContainerStatusTool(t *testing.T) {
	// 创建模块（不启用容器）
	config := CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "local",
		ContainerConfig: ContainerConfig{
			Enabled: false,
		},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 创建工具
	tool := &codeExecContainerStatusTool{module: module}

	// 测试 Info 方法
	ctx := context.Background()
	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}

	if info.Name != "code_exec_container_status" {
		t.Errorf("Expected name 'code_exec_container_status', got '%s'", info.Name)
	}

	// 测试查询状态（容器未启用）
	output, err := tool.InvokableRun(ctx, "{}")
	if err != nil {
		t.Fatalf("InvokableRun() failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if result["enabled"].(bool) {
		t.Error("Expected enabled to be false")
	}
}

// TestCodeExecContainerStatusToolWithDocker 测试容器状态查询工具（Docker 可用时）
func TestCodeExecContainerStatusToolWithDocker(t *testing.T) {
	// 创建模块（启用容器）
	config := CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "container",
		ContainerConfig: ContainerConfig{
			Enabled: true,
		},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 如果容器执行器未初始化，跳过测试
	if module.runner.containerExecutor == nil || !module.runner.containerExecutor.IsEnabled() {
		t.Skip("Docker not available, skipping test")
	}

	// 创建工具
	tool := &codeExecContainerStatusTool{module: module}

	// 测试查询状态（容器已启用）
	ctx := context.Background()
	output, err := tool.InvokableRun(ctx, "{}")
	if err != nil {
		t.Fatalf("InvokableRun() failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if !result["enabled"].(bool) {
		t.Error("Expected enabled to be true")
	}

	// 验证统计信息存在
	if _, ok := result["stats"]; !ok {
		t.Error("Expected stats to be present")
	}

	// 验证容器列表存在
	if _, ok := result["containers"]; !ok {
		t.Error("Expected containers to be present")
	}
}

// TestEnhancedCodeExecAnalyzeTool 测试增强的代码分析工具
func TestEnhancedCodeExecAnalyzeTool(t *testing.T) {
	// 创建模块
	config := CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "local",
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 创建工具
	tool := &codeExecAnalyzeTool{module: module}

	// 测试分析 Python 代码
	ctx := context.Background()
	input := `{
		"language": "python",
		"code": "import requests\nresponse = requests.get('http://example.com')\nprint(response.text)"
	}`

	output, err := tool.InvokableRun(ctx, input)
	if err != nil {
		t.Fatalf("InvokableRun() failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// 验证新增的字段
	if _, ok := result["quality_issues"]; !ok {
		t.Error("Expected quality_issues to be present")
	}

	if _, ok := result["score"]; !ok {
		t.Error("Expected score to be present")
	}

	// 验证网络操作被检测到
	if _, ok := result["network_ops"]; !ok {
		t.Error("Expected network_ops to be present")
	}
}
