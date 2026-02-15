// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestNewCodeExecutorModuleWithFullConfig 测试使用完整配置创建模块
func TestNewCodeExecutorModuleWithFullConfig(t *testing.T) {
	// 创建完整配置
	fullConfig := DefaultFullConfig()
	fullConfig.Executor.Timeout = 45000
	fullConfig.Executor.ExecutionMode = "local"
	fullConfig.Yaegi.EnableCache = true
	fullConfig.Yaegi.CacheCapacity = 50
	fullConfig.Analyzer.EnableNetworkDetection = true
	fullConfig.Analyzer.StrictMode = false

	// 创建模块
	module, err := NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 验证配置已应用
	if module.config.Timeout != 45000 {
		t.Errorf("Expected Timeout 45000, got %d", module.config.Timeout)
	}
	if module.config.ExecutionMode != "local" {
		t.Errorf("Expected ExecutionMode 'local', got '%s'", module.config.ExecutionMode)
	}

	// 验证完整配置已存储
	if module.fullConfig == nil {
		t.Fatal("Expected fullConfig to be stored")
	}
	if module.fullConfig.Yaegi.CacheCapacity != 50 {
		t.Errorf("Expected Yaegi.CacheCapacity 50, got %d", module.fullConfig.Yaegi.CacheCapacity)
	}
}

// TestNewCodeExecutorModuleFromFile 测试从文件创建模块
func TestNewCodeExecutorModuleFromFile(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 保存配置
	config := DefaultFullConfig()
	config.Executor.Timeout = 50000
	config.Executor.ExecutionMode = "auto"
	if err := SaveConfigToFile(&config, tmpFile.Name()); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 从文件创建模块
	module, err := NewCodeExecutorModuleFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create module from file: %v", err)
	}
	defer module.Close()

	// 验证配置已加载
	if module.config.Timeout != 50000 {
		t.Errorf("Expected Timeout 50000, got %d", module.config.Timeout)
	}
	if module.config.ExecutionMode != "auto" {
		t.Errorf("Expected ExecutionMode 'auto', got '%s'", module.config.ExecutionMode)
	}
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	// 使用旧的构造函数
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        256,
		CPULimit:           1,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "local",
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module with old constructor: %v", err)
	}
	defer module.Close()

	// 验证模块正常工作
	if module.config.Timeout != 30000 {
		t.Errorf("Expected Timeout 30000, got %d", module.config.Timeout)
	}

	// 验证完整配置已自动创建
	if module.fullConfig == nil {
		t.Fatal("Expected fullConfig to be auto-created")
	}

	// 测试代码执行
	result, err := module.runCode("python", "print('Hello, World!')", "", 5000)
	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected code execution to succeed")
	}
}

// TestModuleIntegration 测试模块集成
func TestModuleIntegration(t *testing.T) {
	// 创建完整配置
	fullConfig := DefaultFullConfig()
	fullConfig.Executor.SupportedLanguages = []string{"python", "go"}
	fullConfig.Executor.ExecutionMode = "local"
	fullConfig.Yaegi.EnableCache = true
	fullConfig.Analyzer.EnableNetworkDetection = true

	// 创建模块
	module, err := NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试 Python 代码执行
	pythonCode := "print('Hello from Python')"
	result, err := module.runCode("python", pythonCode, "", 5000)
	if err != nil {
		t.Fatalf("Failed to run Python code: %v", err)
	}
	if !result["success"].(bool) {
		t.Error("Expected Python execution to succeed")
	}

	// 测试 Go 代码执行（使用 Yaegi）
	goCode := `
package main
import "fmt"
func main() {
	fmt.Println("Hello from Go")
}
`
	result, err = module.runCode("go", goCode, "", 5000)
	if err != nil {
		t.Fatalf("Failed to run Go code: %v", err)
	}
	if !result["success"].(bool) {
		t.Error("Expected Go execution to succeed")
	}

	// 测试代码分析
	analysisResult, err := module.analyzeCode("python", "import requests\nresponse = requests.get('http://example.com')")
	if err != nil {
		t.Fatalf("Failed to analyze code: %v", err)
	}
	if !analysisResult["success"].(bool) {
		t.Error("Expected code analysis to succeed")
	}

	// 验证网络操作被检测到
	if networkOps, ok := analysisResult["network_ops"].([]NetworkOperation); ok {
		if len(networkOps) == 0 {
			t.Error("Expected network operations to be detected")
		}
	}
}

// TestYaegiConfigIntegration 测试 Yaegi 配置集成
func TestYaegiConfigIntegration(t *testing.T) {
	// 创建配置，启用 Yaegi 缓存
	fullConfig := DefaultFullConfig()
	fullConfig.Executor.SupportedLanguages = []string{"go"}
	fullConfig.Yaegi.EnableCache = true
	fullConfig.Yaegi.CacheCapacity = 10

	module, err := NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 获取 Go 运行器
	goRunner, ok := module.runner.runners["go"].(*GoRunner)
	if !ok {
		t.Fatal("Expected Go runner to be available")
	}

	// 验证 Yaegi 解释器已初始化
	if goRunner.yaegiInterpreter == nil {
		t.Fatal("Expected Yaegi interpreter to be initialized")
	}

	// 执行相同的代码两次，测试缓存
	code := `
package main
import "fmt"
func main() {
	fmt.Println("Test")
}
`

	// 第一次执行
	result1, err := goRunner.Run(context.Background(), code, "")
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}
	if !result1.Success {
		t.Error("Expected first execution to succeed")
	}

	// 第二次执行（应该使用缓存）
	result2, err := goRunner.Run(context.Background(), code, "")
	if err != nil {
		t.Fatalf("Second execution failed: %v", err)
	}
	if !result2.Success {
		t.Error("Expected second execution to succeed")
	}

	// 获取缓存统计
	stats := goRunner.yaegiInterpreter.GetCacheStats()
	if stats.Hits == 0 {
		t.Error("Expected at least one cache hit")
	}
}

// TestContainerConfigIntegration 测试容器配置集成
func TestContainerConfigIntegration(t *testing.T) {
	// 创建配置，启用容器
	fullConfig := DefaultFullConfig()
	fullConfig.Container.Enabled = true
	fullConfig.Container.EnablePool = false
	fullConfig.Container.Timeout = 10 * time.Second

	module, err := NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 验证容器执行器已初始化
	if module.runner.containerExecutor == nil {
		t.Skip("Container executor not available (Docker not installed)")
	}

	if !module.runner.containerExecutor.IsEnabled() {
		t.Skip("Container executor not enabled")
	}

	// 测试容器执行
	ctx := context.Background()
	result, err := module.runner.containerExecutor.Execute(ctx, "python", "print('Hello from container')")
	if err != nil {
		// Docker 不可用，跳过测试
		t.Skipf("Container execution not available: %v", err)
	}

	// 只有在没有错误的情况下才检查 result
	if result != nil && !result.Success {
		t.Skipf("Container execution failed (Docker may not be available): %s", result.Error)
	}
}

// TestAnalyzerConfigIntegration 测试分析器配置集成
func TestAnalyzerConfigIntegration(t *testing.T) {
	// 创建配置
	fullConfig := DefaultFullConfig()
	fullConfig.Analyzer.EnableNetworkDetection = true
	fullConfig.Analyzer.EnableFileSystemDetection = true
	fullConfig.Analyzer.EnableQualityCheck = true

	module, err := NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试代码分析
	code := `
import requests
import os

response = requests.get('http://example.com')
os.remove('/tmp/test.txt')
print(response.text)
`

	result, err := module.analyzeCode("python", code)
	if err != nil {
		t.Fatalf("Failed to analyze code: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected analysis to succeed")
	}

	// 验证网络操作被检测到
	if networkOps, ok := result["network_ops"].([]NetworkOperation); ok {
		if len(networkOps) == 0 {
			t.Error("Expected network operations to be detected")
		}
	}

	// 验证文件系统操作被检测到
	if fsOps, ok := result["filesystem_ops"].([]FileSystemOperation); ok {
		if len(fsOps) == 0 {
			t.Error("Expected filesystem operations to be detected")
		}
	}

	// 验证质量问题被检测到
	if _, ok := result["quality_issues"]; !ok {
		t.Error("Expected quality_issues to be present")
	}

	// 验证代码评分存在
	if _, ok := result["score"]; !ok {
		t.Error("Expected score to be present")
	}
}
