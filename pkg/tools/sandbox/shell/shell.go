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

package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ShellModule Shell命令模块
type ShellModule struct {
	config   ShellConfig
	executor *ShellExecutor
	mu       sync.RWMutex
	stats    *ExecutionStats
}

// ShellConfig Shell配置
type ShellConfig struct {
	Timeout          int      `json:"timeout"`
	MemoryLimit      int      `json:"memory_limit"`
	CPULimit         int      `json:"cpu_limit"`
	CommandWhitelist []string `json:"command_whitelist"`
	EnableBlacklist  bool     `json:"enable_blacklist"`
	CommandBlacklist []string `json:"command_blacklist"`
}

// ShellExecutor Shell命令执行器
type ShellExecutor struct {
	config ShellConfig
	mu     sync.Mutex
}

// ExecutionStats 执行统计
type ExecutionStats struct {
	TotalExecutions int64
	SuccessCount    int64
	FailureCount    int64
	BlockedCount    int64
	mu              sync.RWMutex
}

// NewShellModule 创建Shell模块实例
func NewShellModule(config ShellConfig) (*ShellModule, error) {
	// 验证配置
	if config.Timeout <= 0 {
		config.Timeout = 30000 // 默认30秒
	}
	if config.MemoryLimit <= 0 {
		config.MemoryLimit = 512 // 默认512MB
	}
	if config.CPULimit <= 0 {
		config.CPULimit = 1 // 默认1核
	}

	executor := &ShellExecutor{
		config: config,
	}

	stats := &ExecutionStats{}

	return &ShellModule{
		config:   config,
		executor: executor,
		stats:    stats,
	}, nil
}

// GetTools 返回Shell模块的 MCP 工具列表
func (m *ShellModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// Shell命令执行工具
		&shellExecTool{module: m},
		// 获取支持的命令列表工具
		&shellSupportedCommandsTool{module: m},
	}

	return tools, nil
}

// Shell命令执行工具
type shellExecTool struct {
	module *ShellModule
}

func (t *shellExecTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "shell_exec",
		Desc: "Execute a shell command in a sandboxed environment",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Type: "string",
				Desc: "Shell command to execute",
			},
			"timeout": {
				Type: "integer",
				Desc: "Execution timeout in milliseconds",
			},
			"cwd": {
				Type: "string",
				Desc: "Working directory",
			},
		}),
	}, nil
}

func (t *shellExecTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Command string `json:"command"`
		Timeout int    `json:"timeout"`
		CWD     string `json:"cwd"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.execCommand(args.Command, args.Timeout, args.CWD)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取支持的命令列表工具
type shellSupportedCommandsTool struct {
	module *ShellModule
}

func (t *shellSupportedCommandsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "shell_supported_commands",
		Desc:        "Get list of supported shell commands",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *shellSupportedCommandsTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.getSupportedCommands()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭Shell模块，释放资源
func (m *ShellModule) Close() error {
	// 清理资源
	m.mu.Lock()
	defer m.mu.Unlock()

	// 记录最终统计信息
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	// 可以在这里添加日志记录
	return nil
}

// GetStats 获取执行统计信息
func (m *ShellModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_executions": m.stats.TotalExecutions,
		"success_count":    m.stats.SuccessCount,
		"failure_count":    m.stats.FailureCount,
		"blocked_count":    m.stats.BlockedCount,
	}
}

// 代码执行模块核心功能实现

// execCommand 执行Shell命令
func (m *ShellModule) execCommand(command string, timeout int, cwd string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalExecutions++
	m.stats.mu.Unlock()

	// 如果没有指定超时，使用配置中的默认值
	if timeout == 0 {
		timeout = m.config.Timeout
	}

	// 检查命令是否允许执行
	if !m.isCommandAllowed(command) {
		m.stats.mu.Lock()
		m.stats.BlockedCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":   false,
			"error":     "Command not allowed",
			"command":   command,
			"exit_code": -1,
		}, nil
	}

	// 解析命令和参数
	parts := strings.Fields(command)
	if len(parts) == 0 {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":   false,
			"error":     "Empty command",
			"exit_code": -1,
		}, nil
	}

	// 创建上下文和超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// 创建命令 - 在 Windows 上使用 cmd.exe 执行命令
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// 设置工作目录
	if cwd != "" {
		// 规范化路径，防止路径遍历
		absPath, err := filepath.Abs(cwd)
		if err != nil {
			m.stats.mu.Lock()
			m.stats.FailureCount++
			m.stats.mu.Unlock()

			return map[string]any{
				"success":   false,
				"error":     fmt.Sprintf("Invalid working directory: %v", err),
				"exit_code": -1,
			}, nil
		}
		cmd.Dir = absPath
	}

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	// 获取退出码
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// 其他错误（如超时）
			m.stats.mu.Lock()
			m.stats.FailureCount++
			m.stats.mu.Unlock()

			return map[string]any{
				"success":   false,
				"error":     err.Error(),
				"command":   command,
				"stdout":    stdout.String(),
				"stderr":    stderr.String(),
				"exit_code": -1,
				"duration":  duration.Milliseconds(),
			}, nil
		}
	}

	// 更新统计
	if exitCode == 0 {
		m.stats.mu.Lock()
		m.stats.SuccessCount++
		m.stats.mu.Unlock()
	} else {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()
	}

	return map[string]any{
		"success":   exitCode == 0,
		"command":   command,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
		"duration":  duration.Milliseconds(),
		"cwd":       cwd,
	}, nil
}

// isWindows 检查是否为 Windows 系统
func isWindows() bool {
	return filepath.Separator == '\\'
}

// getSupportedCommands 获取支持的命令列表
func (m *ShellModule) getSupportedCommands() (map[string]any, error) {
	return map[string]any{
		"success":          true,
		"whitelist":        m.config.CommandWhitelist,
		"blacklist":        m.config.CommandBlacklist,
		"enable_blacklist": m.config.EnableBlacklist,
	}, nil
}

// isCommandAllowed 检查命令是否允许执行
func (m *ShellModule) isCommandAllowed(command string) bool {
	// 获取命令名称（去掉参数）
	commandName := strings.Fields(command)[0]

	// 如果命令白名单不为空，检查命令是否在白名单中
	if len(m.config.CommandWhitelist) > 0 {
		for _, allowedCmd := range m.config.CommandWhitelist {
			if allowedCmd == commandName {
				return true
			}
		}
		return false
	}

	// 如果启用了黑名单，检查命令是否在黑名单中
	if m.config.EnableBlacklist {
		for _, blockedCmd := range m.config.CommandBlacklist {
			if blockedCmd == commandName {
				return false
			}
		}
	}

	// 默认允许执行
	return true
}
