// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ============================================================================
// 配置结构定义
// ============================================================================

// AnalyzerConfig 代码分析器配置
type AnalyzerConfig struct {
	// EnableNetworkDetection 启用网络操作检测
	EnableNetworkDetection bool `yaml:"enable_network_detection" json:"enable_network_detection"`
	// EnableFileSystemDetection 启用文件系统操作检测
	EnableFileSystemDetection bool `yaml:"enable_filesystem_detection" json:"enable_filesystem_detection"`
	// EnableProcessDetection 启用进程操作检测
	EnableProcessDetection bool `yaml:"enable_process_detection" json:"enable_process_detection"`
	// EnableCryptoDetection 启用加密操作检测
	EnableCryptoDetection bool `yaml:"enable_crypto_detection" json:"enable_crypto_detection"`
	// EnableDatabaseDetection 启用数据库操作检测
	EnableDatabaseDetection bool `yaml:"enable_database_detection" json:"enable_database_detection"`
	// EnableQualityCheck 启用代码质量检查
	EnableQualityCheck bool `yaml:"enable_quality_check" json:"enable_quality_check"`
	// CustomRulesFile 自定义规则文件路径
	CustomRulesFile string `yaml:"custom_rules_file" json:"custom_rules_file"`
	// StrictMode 严格模式（发现任何问题都标记为不安全）
	StrictMode bool `yaml:"strict_mode" json:"strict_mode"`
}

// YaegiConfig yaegi 解释器配置（已在 yaegi_interpreter.go 中定义）
// 这里只是为了文档完整性而重新声明

// ContainerConfig Docker 容器配置（已在 container_executor.go 中定义）
// 这里只是为了文档完整性而重新声明

// FullConfig 完整配置结构
type FullConfig struct {
	// Executor 执行器配置
	Executor CodeExecutorConfig `yaml:"executor" json:"executor"`
	// Analyzer 分析器配置
	Analyzer AnalyzerConfig `yaml:"analyzer" json:"analyzer"`
	// Yaegi yaegi 解释器配置
	Yaegi YaegiConfig `yaml:"yaegi" json:"yaegi"`
	// Container 容器配置
	Container ContainerConfig `yaml:"container" json:"container"`
}

// ============================================================================
// 默认配置
// ============================================================================

// DefaultAnalyzerConfig 返回默认的分析器配置
func DefaultAnalyzerConfig() AnalyzerConfig {
	return AnalyzerConfig{
		EnableNetworkDetection:    true,
		EnableFileSystemDetection: true,
		EnableProcessDetection:    true,
		EnableCryptoDetection:     true,
		EnableDatabaseDetection:   true,
		EnableQualityCheck:        true,
		CustomRulesFile:           "",
		StrictMode:                false,
	}
}

// DefaultYaegiConfig 返回默认的 yaegi 配置
func DefaultYaegiConfig() YaegiConfig {
	return YaegiConfig{
		PreloadStdlib:   true,
		PreloadPackages: []string{"fmt", "strings", "time", "math"},
		EnableCache:     true,
		CacheCapacity:   100,
	}
}

// DefaultContainerConfig 返回默认的容器配置
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		Enabled: false,
		DefaultImages: map[string]string{
			"python":     "python:3.11-alpine",
			"javascript": "node:18-alpine",
			"go":         "golang:1.21-alpine",
			"bash":       "alpine:latest",
		},
		CPULimit:    "0.5",
		MemoryLimit: "512m",
		NetworkMode: "none",
		Timeout:     30 * time.Second,
		AutoCleanup: true,
		EnablePool:  false,
		PoolMinSize: 2,
		PoolMaxSize: 10,
	}
}

// DefaultCodeExecutorConfig 返回默认的执行器配置
func DefaultCodeExecutorConfig() CodeExecutorConfig {
	return CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"go", "python", "javascript", "bash"},
		ExecutionMode:      "local",
		ContainerConfig:    DefaultContainerConfig(),
	}
}

// DefaultFullConfig 返回默认的完整配置
func DefaultFullConfig() FullConfig {
	return FullConfig{
		Executor:  DefaultCodeExecutorConfig(),
		Analyzer:  DefaultAnalyzerConfig(),
		Yaegi:     DefaultYaegiConfig(),
		Container: DefaultContainerConfig(),
	}
}

// ============================================================================
// 配置加载和验证
// ============================================================================

// LoadConfigFromFile 从 YAML 文件加载配置
func LoadConfigFromFile(filename string) (*FullConfig, error) {
	// 读取文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 YAML
	var config FullConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// SaveConfigToFile 将配置保存到 YAML 文件
func SaveConfigToFile(config *FullConfig, filename string) error {
	// 验证配置
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig 验证配置
func ValidateConfig(config *FullConfig) error {
	// 验证执行器配置
	if config.Executor.Timeout <= 0 {
		return fmt.Errorf("executor timeout must be positive")
	}
	if config.Executor.MemoryLimit <= 0 {
		return fmt.Errorf("executor memory limit must be positive")
	}
	if config.Executor.CPULimit <= 0 {
		return fmt.Errorf("executor CPU limit must be positive")
	}
	if len(config.Executor.SupportedLanguages) == 0 {
		return fmt.Errorf("executor must support at least one language")
	}

	// 验证执行模式
	validModes := map[string]bool{
		"local":     true,
		"container": true,
		"auto":      true,
	}
	if !validModes[config.Executor.ExecutionMode] {
		return fmt.Errorf("invalid execution mode: %s", config.Executor.ExecutionMode)
	}

	// 验证 yaegi 配置
	if config.Yaegi.CacheCapacity < 0 {
		return fmt.Errorf("yaegi cache capacity must be non-negative")
	}

	// 验证容器配置
	if config.Container.Enabled {
		if config.Container.Timeout <= 0 {
			return fmt.Errorf("container timeout must be positive")
		}
		if config.Container.PoolMinSize < 0 {
			return fmt.Errorf("container pool min size must be non-negative")
		}
		if config.Container.PoolMaxSize < config.Container.PoolMinSize {
			return fmt.Errorf("container pool max size must be >= min size")
		}
	}

	// 验证分析器配置
	if config.Analyzer.CustomRulesFile != "" {
		// 检查文件是否存在
		if _, err := os.Stat(config.Analyzer.CustomRulesFile); os.IsNotExist(err) {
			return fmt.Errorf("custom rules file not found: %s", config.Analyzer.CustomRulesFile)
		}
	}

	return nil
}

// MergeConfigs 合并两个配置（第二个配置覆盖第一个）
func MergeConfigs(base, override *FullConfig) *FullConfig {
	result := *base

	// 合并执行器配置
	if override.Executor.Timeout > 0 {
		result.Executor.Timeout = override.Executor.Timeout
	}
	if override.Executor.MemoryLimit > 0 {
		result.Executor.MemoryLimit = override.Executor.MemoryLimit
	}
	if override.Executor.CPULimit > 0 {
		result.Executor.CPULimit = override.Executor.CPULimit
	}
	if len(override.Executor.SupportedLanguages) > 0 {
		result.Executor.SupportedLanguages = override.Executor.SupportedLanguages
	}
	if override.Executor.ExecutionMode != "" {
		result.Executor.ExecutionMode = override.Executor.ExecutionMode
	}

	// 合并分析器配置
	result.Analyzer = override.Analyzer

	// 合并 yaegi 配置
	result.Yaegi = override.Yaegi

	// 合并容器配置
	result.Container = override.Container

	return &result
}

// ApplyConfigToModule 将配置应用到模块
func ApplyConfigToModule(module *CodeExecutorModule, config *FullConfig) error {
	// 更新执行器配置
	module.config = config.Executor

	// 更新容器配置
	if module.runner.containerExecutor != nil {
		module.runner.containerExecutor.config = config.Container
	}

	// 更新 yaegi 配置
	for _, runner := range module.runner.runners {
		if goRunner, ok := runner.(*GoRunner); ok {
			if goRunner.yaegiInterpreter != nil {
				goRunner.yaegiInterpreter.config = config.Yaegi
			}
		}
	}

	return nil
}
