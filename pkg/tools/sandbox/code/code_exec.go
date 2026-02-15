// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// CodeExecutorModule 代码执行模块
type CodeExecutorModule struct {
	config     CodeExecutorConfig
	fullConfig *FullConfig
	runner     *CodeRunner
}

// CodeExecutorConfig 代码执行配置
type CodeExecutorConfig struct {
	Timeout            int             `json:"timeout" yaml:"timeout"`
	MemoryLimit        int             `json:"memory_limit" yaml:"memory_limit"`
	CPULimit           int             `json:"cpu_limit" yaml:"cpu_limit"`
	SupportedLanguages []string        `json:"supported_languages" yaml:"supported_languages"`
	ContainerConfig    ContainerConfig `json:"container_config" yaml:"container_config"`
	ExecutionMode      string          `json:"execution_mode" yaml:"execution_mode"` // "local", "container", "auto"
}

// CodeRunner 代码运行器
type CodeRunner struct {
	config            CodeExecutorConfig
	runners           map[string]LanguageRunner
	mu                sync.RWMutex
	stats             *ExecutionStats
	tempDir           string
	resourceLimiter   *ResourceLimiter
	codeAnalyzer      *CodeAnalyzer
	containerExecutor *ContainerExecutor
}

// LanguageRunner 语言运行器接口
type LanguageRunner interface {
	// Run 执行代码
	Run(ctx context.Context, code string, input string) (*ExecutionResult, error)

	// Format 格式化代码
	Format(code string) (string, error)

	// Validate 验证代码语法
	Validate(code string) error

	// GetLanguage 获取语言名称
	GetLanguage() string
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success  bool          `json:"success"`
	Output   string        `json:"output"`
	Error    string        `json:"error"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
	MemoryMB int           `json:"memory_mb"`
}

// ExecutionStats 执行统计
type ExecutionStats struct {
	TotalExecutions int64
	SuccessCount    int64
	FailureCount    int64
	BlockedCount    int64
	mu              sync.RWMutex
}

// NewCodeExecutorModule 创建代码执行模块实例（向后兼容）
func NewCodeExecutorModule(config CodeExecutorConfig) (*CodeExecutorModule, error) {
	// 使用默认的完整配置
	fullConfig := DefaultFullConfig()
	fullConfig.Executor = config

	// 向后兼容：如果 ExecutionMode 为空，设置默认值
	if fullConfig.Executor.ExecutionMode == "" {
		fullConfig.Executor.ExecutionMode = "local"
	}

	return NewCodeExecutorModuleWithFullConfig(&fullConfig)
}

// NewCodeExecutorModuleWithFullConfig 使用完整配置创建代码执行模块实例
func NewCodeExecutorModuleWithFullConfig(fullConfig *FullConfig) (*CodeExecutorModule, error) {
	// 验证配置
	if err := ValidateConfig(fullConfig); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	config := fullConfig.Executor

	// 应用默认值
	if config.Timeout <= 0 {
		config.Timeout = 60000 // 默认60秒
	}
	if config.MemoryLimit <= 0 {
		config.MemoryLimit = 512 // 默认512MB
	}
	if config.CPULimit <= 0 {
		config.CPULimit = 2 // 默认2核
	}
	if len(config.SupportedLanguages) == 0 {
		config.SupportedLanguages = []string{"go", "python", "javascript", "bash"}
	}
	if config.ExecutionMode == "" {
		config.ExecutionMode = "local" // 默认本地执行
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "code_exec_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 创建容器执行器（使用完整配置）
	containerExecutor, err := NewContainerExecutor(fullConfig.Container)
	if err != nil {
		// 容器执行器创建失败不影响模块初始化
		containerExecutor = nil
	}

	// 创建代码分析器（使用完整配置）
	codeAnalyzer := NewCodeAnalyzer()
	// 应用分析器配置
	if fullConfig.Analyzer.CustomRulesFile != "" {
		if err := codeAnalyzer.LoadCustomRules(fullConfig.Analyzer.CustomRulesFile); err != nil {
			// 自定义规则加载失败不影响初始化，但记录错误
			fmt.Printf("Warning: failed to load custom rules: %v\n", err)
		}
	}

	runner := &CodeRunner{
		config:            config,
		runners:           make(map[string]LanguageRunner),
		stats:             &ExecutionStats{},
		tempDir:           tempDir,
		resourceLimiter:   NewResourceLimiter(config.MemoryLimit, config.CPULimit, time.Duration(config.Timeout)*time.Millisecond),
		codeAnalyzer:      codeAnalyzer,
		containerExecutor: containerExecutor,
	}

	// 初始化语言运行器
	for _, lang := range config.SupportedLanguages {
		switch strings.ToLower(lang) {
		case "python":
			runner.runners["python"] = NewPythonRunner(config, tempDir)
		case "javascript", "js":
			runner.runners["javascript"] = NewJavaScriptRunner(config)
		case "bash", "sh":
			runner.runners["bash"] = NewBashRunner(config)
		case "go":
			// 使用完整配置创建 Go 运行器
			runner.runners["go"] = NewGoRunnerWithConfig(config, tempDir, fullConfig.Yaegi)
		}
	}

	module := &CodeExecutorModule{
		config: config,
		runner: runner,
	}

	// 存储完整配置的引用（用于后续更新）
	module.fullConfig = fullConfig

	return module, nil
}

// NewCodeExecutorModuleFromFile 从配置文件创建代码执行模块实例
func NewCodeExecutorModuleFromFile(configFile string) (*CodeExecutorModule, error) {
	fullConfig, err := LoadConfigFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	return NewCodeExecutorModuleWithFullConfig(fullConfig)
}

// GetTools 返回代码执行模块的 MCP 工具列表
func (m *CodeExecutorModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 代码执行工具
		&codeExecRunTool{module: m},
		// 获取支持的语言列表工具
		&codeExecSupportedLanguagesTool{module: m},
		// 代码格式化工具
		&codeExecFormatTool{module: m},
		// 代码分析工具
		&codeExecAnalyzeTool{module: m},
		// 设置执行模式工具
		&codeExecSetModeTool{module: m},
		// 容器状态查询工具
		&codeExecContainerStatusTool{module: m},
	}

	return tools, nil
}

// 代码执行工具
type codeExecRunTool struct {
	module *CodeExecutorModule
}

func (t *codeExecRunTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "code_exec_run",
		Desc: "Execute code in a specified language with optional input",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"language": {
				Type: "string",
				Desc: "Programming language (python, javascript, bash, go)",
			},
			"code": {
				Type: "string",
				Desc: "Code to execute",
			},
			"input": {
				Type: "string",
				Desc: "Optional input to pass to the code (stdin)",
			},
			"timeout": {
				Type: "integer",
				Desc: "Execution timeout in milliseconds (optional, defaults to config value)",
			},
		}),
	}, nil
}

func (t *codeExecRunTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Language string `json:"language"`
		Code     string `json:"code"`
		Input    string `json:"input"`
		Timeout  int    `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.runCode(args.Language, args.Code, args.Input, args.Timeout)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取支持的语言列表工具
type codeExecSupportedLanguagesTool struct {
	module *CodeExecutorModule
}

func (t *codeExecSupportedLanguagesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "code_exec_supported_languages",
		Desc:        "Get list of supported programming languages",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *codeExecSupportedLanguagesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.getSupportedLanguages()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 代码格式化工具
type codeExecFormatTool struct {
	module *CodeExecutorModule
}

func (t *codeExecFormatTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "code_exec_format",
		Desc: "Format code in a specified language",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"language": {
				Type: "string",
				Desc: "Programming language",
			},
			"code": {
				Type: "string",
				Desc: "Code to format",
			},
		}),
	}, nil
}

func (t *codeExecFormatTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.formatCode(args.Language, args.Code)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭代码执行模块，释放资源
func (m *CodeExecutorModule) Close() error {
	// 清理临时目录
	if m.runner.tempDir != "" {
		os.RemoveAll(m.runner.tempDir)
	}

	// 关闭容器执行器
	if m.runner.containerExecutor != nil {
		if err := m.runner.containerExecutor.Close(); err != nil {
			return fmt.Errorf("failed to close container executor: %w", err)
		}
	}

	return nil
}

// GetStats 获取执行统计信息
func (m *CodeExecutorModule) GetStats() map[string]int64 {
	m.runner.stats.mu.RLock()
	defer m.runner.stats.mu.RUnlock()

	stats := map[string]int64{
		"total_executions": m.runner.stats.TotalExecutions,
		"success_count":    m.runner.stats.SuccessCount,
		"failure_count":    m.runner.stats.FailureCount,
		"blocked_count":    m.runner.stats.BlockedCount,
	}

	// 添加资源统计
	resourceStats := m.runner.resourceLimiter.GetResourceStats()
	stats["active_executions"] = int64(resourceStats["active_executions"].(int))
	stats["max_concurrent"] = int64(resourceStats["max_concurrent"].(int))
	stats["current_memory_mb"] = int64(resourceStats["current_memory_mb"].(int))

	return stats
}

// AnalyzeCode 分析代码
func (m *CodeExecutorModule) AnalyzeCode(ctx context.Context, language, code string) (*AnalysisResult, error) {
	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		return nil, fmt.Errorf("language not supported")
	}

	// 分析代码
	result := m.runner.codeAnalyzer.Analyze(language, code)

	return result, nil
}

// FormatCode 格式化代码
func (m *CodeExecutorModule) FormatCode(ctx context.Context, language, code string) (string, error) {
	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		return "", fmt.Errorf("language not supported")
	}

	// 获取语言运行器
	m.runner.mu.RLock()
	runner, exists := m.runner.runners[strings.ToLower(language)]
	m.runner.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("language runner not initialized")
	}

	// 格式化代码
	formattedCode, err := runner.Format(code)
	if err != nil {
		return "", fmt.Errorf("failed to format code: %w", err)
	}

	return formattedCode, nil
}

// ExecuteCode 执行代码
func (m *CodeExecutorModule) ExecuteCode(ctx context.Context, language, code, input string, timeout int) (*ExecutionResult, error) {
	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		return nil, fmt.Errorf("language not supported")
	}

	// 如果没有指定超时，使用配置中的默认值
	if timeout == 0 {
		timeout = m.config.Timeout
	}

	// 获取语言运行器
	m.runner.mu.RLock()
	runner, exists := m.runner.runners[strings.ToLower(language)]
	m.runner.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("language runner not initialized")
	}

	// 创建上下文和超时控制
	execCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// 执行代码
	result, err := runner.Run(execCtx, code, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	return result, nil
}

// 代码执行模块核心功能实现

// runCode 执行代码
func (m *CodeExecutorModule) runCode(language, code, input string, timeout int) (map[string]any, error) {
	// 更新统计
	m.runner.stats.mu.Lock()
	m.runner.stats.TotalExecutions++
	m.runner.stats.mu.Unlock()

	// 如果没有指定超时，使用配置中的默认值
	if timeout == 0 {
		timeout = m.config.Timeout
	}

	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		m.runner.stats.mu.Lock()
		m.runner.stats.BlockedCount++
		m.runner.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   "Language not supported",
		}, nil
	}

	// 代码安全分析
	analysisResult := m.runner.codeAnalyzer.Analyze(language, code)
	if !analysisResult.Safe {
		m.runner.stats.mu.Lock()
		m.runner.stats.BlockedCount++
		m.runner.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    "Code contains security issues",
			"analysis": analysisResult,
			"blocked":  true,
		}, nil
	}

	// 获取执行权限（资源限制）
	ctx := context.Background()
	if err := m.runner.resourceLimiter.AcquireExecution(ctx); err != nil {
		m.runner.stats.mu.Lock()
		m.runner.stats.BlockedCount++
		m.runner.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"blocked": true,
		}, nil
	}
	defer m.runner.resourceLimiter.ReleaseExecution()

	// 决定执行模式
	executionMode := m.selectExecutionMode(language)

	// 如果使用容器模式
	if executionMode == "container" && m.runner.containerExecutor != nil && m.runner.containerExecutor.IsEnabled() {
		execCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()

		result, err := m.runner.containerExecutor.Execute(execCtx, language, code)
		if err != nil {
			// 容器执行失败，回退到本地执行
			return m.executeLocally(language, code, input, timeout)
		}

		// 更新统计
		if result.Success {
			m.runner.stats.mu.Lock()
			m.runner.stats.SuccessCount++
			m.runner.stats.mu.Unlock()
		} else {
			m.runner.stats.mu.Lock()
			m.runner.stats.FailureCount++
			m.runner.stats.mu.Unlock()
		}

		return map[string]any{
			"success":        result.Success,
			"language":       language,
			"output":         result.Output,
			"error":          result.Error,
			"exit_code":      result.ExitCode,
			"duration":       result.Duration.Milliseconds(),
			"memory_mb":      result.MemoryMB,
			"analysis":       analysisResult,
			"execution_mode": "container",
		}, nil
	}

	// 本地执行
	return m.executeLocally(language, code, input, timeout)
}

// getSupportedLanguages 获取支持的语言列表
func (m *CodeExecutorModule) getSupportedLanguages() (map[string]any, error) {
	return map[string]any{
		"success":   true,
		"languages": m.config.SupportedLanguages,
	}, nil
}

// formatCode 格式化代码
func (m *CodeExecutorModule) formatCode(language, code string) (map[string]any, error) {
	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		return map[string]any{
			"success": false,
			"error":   "Language not supported",
		}, nil
	}

	// 获取语言运行器
	m.runner.mu.RLock()
	runner, exists := m.runner.runners[strings.ToLower(language)]
	m.runner.mu.RUnlock()

	if !exists {
		return map[string]any{
			"success": false,
			"error":   "Language runner not initialized",
		}, nil
	}

	// 格式化代码
	formatted, err := runner.Format(code)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success":        true,
		"language":       language,
		"formatted_code": formatted,
	}, nil
}

// isLanguageSupported 检查语言是否支持
func (m *CodeExecutorModule) isLanguageSupported(language string) bool {
	language = strings.ToLower(language)
	for _, lang := range m.config.SupportedLanguages {
		if strings.ToLower(lang) == language {
			return true
		}
	}
	return false
}

// selectExecutionMode 选择执行模式
func (m *CodeExecutorModule) selectExecutionMode(language string) string {
	mode := m.config.ExecutionMode

	// 如果是 auto 模式，根据语言和配置自动选择
	if mode == "auto" {
		// 如果容器执行器可用，优先使用容器
		if m.runner.containerExecutor != nil && m.runner.containerExecutor.IsEnabled() {
			return "container"
		}
		return "local"
	}

	return mode
}

// executeLocally 本地执行代码
func (m *CodeExecutorModule) executeLocally(language, code, input string, timeout int) (map[string]any, error) {
	// 获取语言运行器
	m.runner.mu.RLock()
	runner, exists := m.runner.runners[strings.ToLower(language)]
	m.runner.mu.RUnlock()

	if !exists {
		m.runner.stats.mu.Lock()
		m.runner.stats.BlockedCount++
		m.runner.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   "Language runner not initialized",
		}, nil
	}

	// 创建上下文和超时控制
	execCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// 执行代码
	result, err := runner.Run(execCtx, code, input)
	if err != nil {
		m.runner.stats.mu.Lock()
		m.runner.stats.FailureCount++
		m.runner.stats.mu.Unlock()

		return map[string]any{
			"success":   false,
			"error":     err.Error(),
			"language":  language,
			"exit_code": -1,
		}, nil
	}

	// 更新统计
	if result.Success {
		m.runner.stats.mu.Lock()
		m.runner.stats.SuccessCount++
		m.runner.stats.mu.Unlock()
	} else {
		m.runner.stats.mu.Lock()
		m.runner.stats.FailureCount++
		m.runner.stats.mu.Unlock()
	}

	// 获取分析结果
	analysisResult := m.runner.codeAnalyzer.Analyze(language, code)

	return map[string]any{
		"success":        result.Success,
		"language":       language,
		"output":         result.Output,
		"error":          result.Error,
		"exit_code":      result.ExitCode,
		"duration":       result.Duration.Milliseconds(),
		"memory_mb":      result.MemoryMB,
		"analysis":       analysisResult,
		"execution_mode": "local",
	}, nil
}

// ============================================================================
// Python Runner
// ============================================================================

// PythonRunner Python 代码运行器
type PythonRunner struct {
	config  CodeExecutorConfig
	tempDir string
}

// NewPythonRunner 创建 Python 运行器
func NewPythonRunner(config CodeExecutorConfig, tempDir string) *PythonRunner {
	return &PythonRunner{
		config:  config,
		tempDir: tempDir,
	}
}

// Run 执行 Python 代码
func (r *PythonRunner) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp(r.tempDir, "python_*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		return nil, fmt.Errorf("failed to write code: %w", err)
	}
	tmpFile.Close()

	// 执行代码
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "python", "-I", tmpFile.Name())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(startTime)

	// 获取退出码
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return &ExecutionResult{
				Success:  false,
				Error:    err.Error(),
				ExitCode: -1,
				Duration: duration,
			}, nil
		}
	}

	return &ExecutionResult{
		Success:  exitCode == 0,
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// Format 格式化 Python 代码
func (r *PythonRunner) Format(code string) (string, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp(r.tempDir, "python_format_*.py")
	if err != nil {
		return code, nil // 如果创建文件失败，返回原代码
	}
	defer os.Remove(tmpFile.Name())

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		return code, nil
	}
	tmpFile.Close()

	// 尝试使用 black 格式化
	cmd := exec.Command("black", "--quiet", tmpFile.Name())
	if err := cmd.Run(); err == nil {
		// black 成功，读取格式化后的代码
		formatted, err := os.ReadFile(tmpFile.Name())
		if err == nil {
			return string(formatted), nil
		}
	}

	// 如果 black 不可用，尝试使用 autopep8
	cmd = exec.Command("autopep8", "--in-place", tmpFile.Name())
	if err := cmd.Run(); err == nil {
		// autopep8 成功，读取格式化后的代码
		formatted, err := os.ReadFile(tmpFile.Name())
		if err == nil {
			return string(formatted), nil
		}
	}

	// 如果都不可用，返回原代码
	return code, nil
}

// Validate 验证 Python 代码语法
func (r *PythonRunner) Validate(code string) error {
	// 使用 python -m py_compile 验证语法
	tmpFile, err := os.CreateTemp(r.tempDir, "python_validate_*.py")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("python", "-m", "py_compile", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}

	return nil
}

// GetLanguage 获取语言名称
func (r *PythonRunner) GetLanguage() string {
	return "python"
}

// ============================================================================
// JavaScript Runner
// ============================================================================

// JavaScriptRunner JavaScript 代码运行器
type JavaScriptRunner struct {
	config CodeExecutorConfig
}

// NewJavaScriptRunner 创建 JavaScript 运行器
func NewJavaScriptRunner(config CodeExecutorConfig) *JavaScriptRunner {
	return &JavaScriptRunner{
		config: config,
	}
}

// Run 执行 JavaScript 代码
func (r *JavaScriptRunner) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	// 使用 Node.js 执行代码
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "node", "-e", code)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	// 获取退出码
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return &ExecutionResult{
				Success:  false,
				Error:    err.Error(),
				ExitCode: -1,
				Duration: duration,
			}, nil
		}
	}

	return &ExecutionResult{
		Success:  exitCode == 0,
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// Format 格式化 JavaScript 代码
func (r *JavaScriptRunner) Format(code string) (string, error) {
	// 尝试使用 prettier 格式化
	cmd := exec.Command("prettier", "--parser", "babel", "--stdin-filepath", "code.js")
	cmd.Stdin = strings.NewReader(code)

	output, err := cmd.Output()
	if err != nil {
		// 如果 prettier 不可用或失败，返回原代码
		return code, nil
	}

	return string(output), nil
}

// Validate 验证 JavaScript 代码语法
func (r *JavaScriptRunner) Validate(code string) error {
	// 使用 node --check 验证语法
	cmd := exec.Command("node", "--check", "-e", code)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}
	return nil
}

// GetLanguage 获取语言名称
func (r *JavaScriptRunner) GetLanguage() string {
	return "javascript"
}

// ============================================================================
// Bash Runner
// ============================================================================

// BashRunner Bash 代码运行器
type BashRunner struct {
	config CodeExecutorConfig
}

// NewBashRunner 创建 Bash 运行器
func NewBashRunner(config CodeExecutorConfig) *BashRunner {
	return &BashRunner{
		config: config,
	}
}

// Run 执行 Bash 代码
func (r *BashRunner) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	startTime := time.Now()

	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/C", code)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", code)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	// 获取退出码
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return &ExecutionResult{
				Success:  false,
				Error:    err.Error(),
				ExitCode: -1,
				Duration: duration,
			}, nil
		}
	}

	return &ExecutionResult{
		Success:  exitCode == 0,
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// Format 格式化 Bash 代码
func (r *BashRunner) Format(code string) (string, error) {
	// 尝试使用 shfmt 格式化
	cmd := exec.Command("shfmt", "-")
	cmd.Stdin = strings.NewReader(code)

	output, err := cmd.Output()
	if err != nil {
		// 如果 shfmt 不可用或失败，返回原代码
		return code, nil
	}

	return string(output), nil
}

// Validate 验证 Bash 代码语法
func (r *BashRunner) Validate(code string) error {
	// 使用 bash -n 验证语法
	var cmd *exec.Cmd
	if isWindows() {
		// Windows 上不支持语法检查
		return nil
	} else {
		cmd = exec.Command("bash", "-n", "-c", code)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}
	return nil
}

// GetLanguage 获取语言名称
func (r *BashRunner) GetLanguage() string {
	return "bash"
}

// isWindows 检查是否为 Windows 系统
func isWindows() bool {
	return filepath.Separator == '\\'
}

// ============================================================================
// Go Runner
// ============================================================================

// GoRunner Go 代码运行器
type GoRunner struct {
	config           CodeExecutorConfig
	tempDir          string
	yaegiInterpreter *YaegiInterpreter
	executionMode    ExecutionMode
	stats            *GoRunnerStats
	mu               sync.RWMutex
}

// GoRunnerStats Go 运行器统计
type GoRunnerStats struct {
	YaegiExecutions int64
	GoRunExecutions int64
	YaegiFailures   int64
	Fallbacks       int64
	mu              sync.RWMutex
}

// NewGoRunner 创建 Go 运行器
func NewGoRunner(config CodeExecutorConfig, tempDir string) *GoRunner {
	// 使用默认 yaegi 配置
	return NewGoRunnerWithConfig(config, tempDir, DefaultYaegiConfig())
}

// NewGoRunnerWithConfig 使用指定的 yaegi 配置创建 Go 运行器
func NewGoRunnerWithConfig(config CodeExecutorConfig, tempDir string, yaegiConfig YaegiConfig) *GoRunner {
	// 创建 yaegi 解释器
	yaegi, err := NewYaegiInterpreter(yaegiConfig)
	if err != nil {
		// 如果 yaegi 初始化失败，只使用 go run
		return &GoRunner{
			config:        config,
			tempDir:       tempDir,
			executionMode: ModeGoRun,
			stats:         &GoRunnerStats{},
		}
	}

	return &GoRunner{
		config:           config,
		tempDir:          tempDir,
		yaegiInterpreter: yaegi,
		executionMode:    ModeAuto, // 默认使用自动模式
		stats:            &GoRunnerStats{},
	}
}

// Run 执行 Go 代码
func (r *GoRunner) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	r.mu.RLock()
	mode := r.executionMode
	r.mu.RUnlock()

	// 根据执行模式选择执行方式
	switch mode {
	case ModeYaegi:
		return r.runWithYaegi(ctx, code, input)
	case ModeGoRun:
		return r.runWithGoRun(ctx, code, input)
	case ModeAuto:
		return r.runWithAuto(ctx, code, input)
	default:
		return r.runWithAuto(ctx, code, input)
	}
}

// runWithYaegi 使用 yaegi 执行代码
func (r *GoRunner) runWithYaegi(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	r.stats.mu.Lock()
	r.stats.YaegiExecutions++
	r.stats.mu.Unlock()

	if r.yaegiInterpreter == nil || !r.yaegiInterpreter.IsAvailable() {
		r.stats.mu.Lock()
		r.stats.YaegiFailures++
		r.stats.mu.Unlock()
		return nil, fmt.Errorf("yaegi interpreter not available")
	}

	result, err := r.yaegiInterpreter.Run(ctx, code, input)
	if err != nil {
		r.stats.mu.Lock()
		r.stats.YaegiFailures++
		r.stats.mu.Unlock()
	}

	return result, err
}

// runWithGoRun 使用 go run 执行代码
func (r *GoRunner) runWithGoRun(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	r.stats.mu.Lock()
	r.stats.GoRunExecutions++
	r.stats.mu.Unlock()

	// 创建临时文件
	tmpFile, err := os.CreateTemp(r.tempDir, "go_*.go")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// 如果代码没有 package main，添加它
	if !strings.Contains(code, "package main") {
		code = "package main\n\n" + code
	}

	// 如果代码没有 main 函数，包装它
	if !strings.Contains(code, "func main()") {
		code = code + "\n\nfunc main() {}\n"
	}

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		return nil, fmt.Errorf("failed to write code: %w", err)
	}
	tmpFile.Close()

	// 执行代码
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "go", "run", tmpFile.Name())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(startTime)

	// 获取退出码
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return &ExecutionResult{
				Success:  false,
				Error:    err.Error(),
				ExitCode: -1,
				Duration: duration,
			}, nil
		}
	}

	return &ExecutionResult{
		Success:  exitCode == 0,
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// runWithAuto 自动选择执行模式
func (r *GoRunner) runWithAuto(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	// 首先尝试使用 yaegi（更快）
	if r.yaegiInterpreter != nil && r.yaegiInterpreter.IsAvailable() {
		result, err := r.runWithYaegi(ctx, code, input)

		// 如果 yaegi 执行成功，直接返回
		if err == nil && result.Success {
			return result, nil
		}

		// 如果 yaegi 失败，回退到 go run
		r.stats.mu.Lock()
		r.stats.Fallbacks++
		r.stats.mu.Unlock()
	}

	// 回退到 go run
	return r.runWithGoRun(ctx, code, input)
}

// SetExecutionMode 设置执行模式
func (r *GoRunner) SetExecutionMode(mode ExecutionMode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executionMode = mode
}

// GetExecutionMode 获取当前执行模式
func (r *GoRunner) GetExecutionMode() ExecutionMode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.executionMode
}

// GetStats 获取运行器统计信息
func (r *GoRunner) GetStats() map[string]int64 {
	r.stats.mu.RLock()
	defer r.stats.mu.RUnlock()

	return map[string]int64{
		"yaegi_executions":  r.stats.YaegiExecutions,
		"go_run_executions": r.stats.GoRunExecutions,
		"yaegi_failures":    r.stats.YaegiFailures,
		"fallbacks":         r.stats.Fallbacks,
	}
}

// Format 格式化 Go 代码
func (r *GoRunner) Format(code string) (string, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp(r.tempDir, "go_format_*.go")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	// 使用 gofmt 格式化
	cmd := exec.Command("gofmt", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return code, nil // 如果格式化失败，返回原代码
	}

	return string(output), nil
}

// Validate 验证 Go 代码语法
func (r *GoRunner) Validate(code string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp(r.tempDir, "go_validate_*.go")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		return err
	}
	tmpFile.Close()

	// 使用 go build 验证语法
	cmd := exec.Command("go", "build", "-o", "/dev/null", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}

	return nil
}

// GetLanguage 获取语言名称
func (r *GoRunner) GetLanguage() string {
	return "go"
}

// ============================================================================
// Code Analysis Tool
// ============================================================================

// 代码分析工具
type codeExecAnalyzeTool struct {
	module *CodeExecutorModule
}

func (t *codeExecAnalyzeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "code_exec_analyze",
		Desc: "Analyze code for security issues, complexity, and best practices",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"language": {
				Type: "string",
				Desc: "Programming language",
			},
			"code": {
				Type: "string",
				Desc: "Code to analyze",
			},
		}),
	}, nil
}

func (t *codeExecAnalyzeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.analyzeCode(args.Language, args.Code)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// analyzeCode 分析代码
func (m *CodeExecutorModule) analyzeCode(language, code string) (map[string]any, error) {
	// 检查语言是否支持
	if !m.isLanguageSupported(language) {
		return map[string]any{
			"success": false,
			"error":   "Language not supported",
		}, nil
	}

	// 分析代码
	result := m.runner.codeAnalyzer.Analyze(language, code)

	return map[string]any{
		"success":           true,
		"language":          language,
		"safe":              result.Safe,
		"issues":            result.Issues,
		"complexity":        result.Complexity,
		"line_count":        result.LineCount,
		"char_count":        result.CharCount,
		"has_dangerous_ops": result.HasDangerousOps,
		"suggestions":       result.Suggestions,
		"network_ops":       result.NetworkOps,
		"filesystem_ops":    result.FileSystemOps,
		"process_ops":       result.ProcessOps,
		"crypto_issues":     result.CryptoIssues,
		"database_ops":      result.DatabaseOps,
		"quality_issues":    result.QualityIssues,
		"score":             result.Score,
	}, nil
}
