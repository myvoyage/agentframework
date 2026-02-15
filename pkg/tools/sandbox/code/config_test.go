// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"os"
	"testing"
	"time"
)

// TestDefaultConfigs 测试默认配置
func TestDefaultConfigs(t *testing.T) {
	// 测试默认分析器配置
	analyzerConfig := DefaultAnalyzerConfig()
	if !analyzerConfig.EnableNetworkDetection {
		t.Error("Expected EnableNetworkDetection to be true")
	}
	if !analyzerConfig.EnableQualityCheck {
		t.Error("Expected EnableQualityCheck to be true")
	}

	// 测试默认 yaegi 配置
	yaegiConfig := DefaultYaegiConfig()
	if !yaegiConfig.PreloadStdlib {
		t.Error("Expected PreloadStdlib to be true")
	}
	if !yaegiConfig.EnableCache {
		t.Error("Expected EnableCache to be true")
	}
	if yaegiConfig.CacheCapacity != 100 {
		t.Errorf("Expected CacheCapacity 100, got %d", yaegiConfig.CacheCapacity)
	}

	// 测试默认容器配置
	containerConfig := DefaultContainerConfig()
	if containerConfig.Enabled {
		t.Error("Expected Enabled to be false by default")
	}
	if containerConfig.NetworkMode != "none" {
		t.Errorf("Expected NetworkMode 'none', got '%s'", containerConfig.NetworkMode)
	}

	// 测试默认执行器配置
	executorConfig := DefaultCodeExecutorConfig()
	if executorConfig.Timeout != 60000 {
		t.Errorf("Expected Timeout 60000, got %d", executorConfig.Timeout)
	}
	if executorConfig.ExecutionMode != "local" {
		t.Errorf("Expected ExecutionMode 'local', got '%s'", executorConfig.ExecutionMode)
	}

	// 测试默认完整配置
	fullConfig := DefaultFullConfig()
	if fullConfig.Executor.Timeout != 60000 {
		t.Error("Expected Executor.Timeout to be 60000")
	}
	if !fullConfig.Analyzer.EnableNetworkDetection {
		t.Error("Expected Analyzer.EnableNetworkDetection to be true")
	}
}

// TestValidateConfig 测试配置验证
func TestValidateConfig(t *testing.T) {
	// 测试有效配置
	validConfig := DefaultFullConfig()
	if err := ValidateConfig(&validConfig); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// 测试无效超时
	invalidTimeout := DefaultFullConfig()
	invalidTimeout.Executor.Timeout = 0
	if err := ValidateConfig(&invalidTimeout); err == nil {
		t.Error("Expected error for invalid timeout")
	}

	// 测试无效内存限制
	invalidMemory := DefaultFullConfig()
	invalidMemory.Executor.MemoryLimit = -1
	if err := ValidateConfig(&invalidMemory); err == nil {
		t.Error("Expected error for invalid memory limit")
	}

	// 测试无效执行模式
	invalidMode := DefaultFullConfig()
	invalidMode.Executor.ExecutionMode = "invalid"
	if err := ValidateConfig(&invalidMode); err == nil {
		t.Error("Expected error for invalid execution mode")
	}

	// 测试空语言列表
	emptyLanguages := DefaultFullConfig()
	emptyLanguages.Executor.SupportedLanguages = []string{}
	if err := ValidateConfig(&emptyLanguages); err == nil {
		t.Error("Expected error for empty supported languages")
	}

	// 测试无效缓存容量
	invalidCache := DefaultFullConfig()
	invalidCache.Yaegi.CacheCapacity = -1
	if err := ValidateConfig(&invalidCache); err == nil {
		t.Error("Expected error for invalid cache capacity")
	}

	// 测试无效容器池大小
	invalidPool := DefaultFullConfig()
	invalidPool.Container.Enabled = true
	invalidPool.Container.PoolMaxSize = 1
	invalidPool.Container.PoolMinSize = 5
	if err := ValidateConfig(&invalidPool); err == nil {
		t.Error("Expected error for invalid pool size")
	}
}

// TestLoadSaveConfig 测试配置加载和保存
func TestLoadSaveConfig(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 保存默认配置
	defaultConfig := DefaultFullConfig()
	if err := SaveConfigToFile(&defaultConfig, tmpFile.Name()); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 加载配置
	loadedConfig, err := LoadConfigFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证加载的配置
	if loadedConfig.Executor.Timeout != defaultConfig.Executor.Timeout {
		t.Errorf("Expected Timeout %d, got %d", defaultConfig.Executor.Timeout, loadedConfig.Executor.Timeout)
	}
	if loadedConfig.Executor.ExecutionMode != defaultConfig.Executor.ExecutionMode {
		t.Errorf("Expected ExecutionMode '%s', got '%s'", defaultConfig.Executor.ExecutionMode, loadedConfig.Executor.ExecutionMode)
	}
	if loadedConfig.Yaegi.EnableCache != defaultConfig.Yaegi.EnableCache {
		t.Error("Expected Yaegi.EnableCache to match")
	}
}

// TestMergeConfigs 测试配置合并
func TestMergeConfigs(t *testing.T) {
	// 创建基础配置
	base := DefaultFullConfig()
	base.Executor.Timeout = 30000
	base.Executor.ExecutionMode = "local"

	// 创建覆盖配置
	override := DefaultFullConfig()
	override.Executor.Timeout = 60000
	override.Executor.ExecutionMode = "container"
	override.Analyzer.StrictMode = true

	// 合并配置
	merged := MergeConfigs(&base, &override)

	// 验证合并结果
	if merged.Executor.Timeout != 60000 {
		t.Errorf("Expected Timeout 60000, got %d", merged.Executor.Timeout)
	}
	if merged.Executor.ExecutionMode != "container" {
		t.Errorf("Expected ExecutionMode 'container', got '%s'", merged.Executor.ExecutionMode)
	}
	if !merged.Analyzer.StrictMode {
		t.Error("Expected StrictMode to be true")
	}
}

// TestApplyConfigToModule 测试将配置应用到模块
func TestApplyConfigToModule(t *testing.T) {
	// 创建模块
	initialConfig := DefaultCodeExecutorConfig()
	initialConfig.Timeout = 30000
	initialConfig.ExecutionMode = "local"

	module, err := NewCodeExecutorModule(initialConfig)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 创建新配置
	newConfig := DefaultFullConfig()
	newConfig.Executor.Timeout = 60000
	newConfig.Executor.ExecutionMode = "container"

	// 应用配置
	if err := ApplyConfigToModule(module, &newConfig); err != nil {
		t.Fatalf("Failed to apply config: %v", err)
	}

	// 验证配置已应用
	if module.config.Timeout != 60000 {
		t.Errorf("Expected Timeout 60000, got %d", module.config.Timeout)
	}
	if module.config.ExecutionMode != "container" {
		t.Errorf("Expected ExecutionMode 'container', got '%s'", module.config.ExecutionMode)
	}
}

// TestConfigWithContainerPool 测试容器池配置
func TestConfigWithContainerPool(t *testing.T) {
	config := DefaultFullConfig()
	config.Container.Enabled = true
	config.Container.EnablePool = true
	config.Container.PoolMinSize = 2
	config.Container.PoolMaxSize = 10
	config.Container.Timeout = 30 * time.Second

	// 验证配置
	if err := ValidateConfig(&config); err != nil {
		t.Errorf("Valid container pool config should pass validation: %v", err)
	}

	// 测试无效的池大小配置
	invalidConfig := config
	invalidConfig.Container.PoolMinSize = 10
	invalidConfig.Container.PoolMaxSize = 2
	if err := ValidateConfig(&invalidConfig); err == nil {
		t.Error("Expected error for invalid pool size configuration")
	}
}

// TestConfigYAMLFormat 测试 YAML 格式
func TestConfigYAMLFormat(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "config_yaml_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 写入 YAML 配置
	yamlContent := `
executor:
  timeout: 45000
  memory_limit: 1024
  cpu_limit: 4
  supported_languages:
    - python
    - go
  execution_mode: auto

analyzer:
  enable_network_detection: true
  enable_quality_check: false
  strict_mode: true

yaegi:
  enable_cache: true
  cache_capacity: 200

container:
  enabled: true
  cpu_limit: "1.0"
  memory_limit: "1g"
  timeout: 30s
  enable_pool: true
  pool_min_size: 3
  pool_max_size: 15
`
	if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// 加载配置
	config, err := LoadConfigFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证解析结果
	if config.Executor.Timeout != 45000 {
		t.Errorf("Expected Timeout 45000, got %d", config.Executor.Timeout)
	}
	if config.Executor.MemoryLimit != 1024 {
		t.Errorf("Expected MemoryLimit 1024, got %d", config.Executor.MemoryLimit)
	}
	if config.Executor.ExecutionMode != "auto" {
		t.Errorf("Expected ExecutionMode 'auto', got '%s'", config.Executor.ExecutionMode)
	}
	if config.Analyzer.EnableQualityCheck {
		t.Error("Expected EnableQualityCheck to be false")
	}
	if !config.Analyzer.StrictMode {
		t.Error("Expected StrictMode to be true")
	}
	if config.Yaegi.CacheCapacity != 200 {
		t.Errorf("Expected CacheCapacity 200, got %d", config.Yaegi.CacheCapacity)
	}
	if !config.Container.Enabled {
		t.Error("Expected Container.Enabled to be true")
	}
	if config.Container.PoolMinSize != 3 {
		t.Errorf("Expected PoolMinSize 3, got %d", config.Container.PoolMinSize)
	}
}
