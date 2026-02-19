// Agent Framework - Process Manager Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ProcessManagerModule 进程管理模块
type ProcessManagerModule struct {
	config ProcessConfig
	mu     sync.RWMutex
	stats  *ProcessStats
}

// ProcessConfig 进程管理配置
type ProcessConfig struct {
	MaxProcesses       int      `json:"max_processes"`        // 最大监控进程数
	EnableAutoCleanup  bool     `json:"enable_auto_cleanup"`  // 启用自动清理僵尸进程
	CleanupInterval    int      `json:"cleanup_interval"`     // 清理间隔（秒）
	AllowedProcesses   []string `json:"allowed_processes"`    // 允许操作的进程列表
	BlockedProcesses  []string `json:"blocked_processes"`   // 禁止操作的进程列表
	EnableMonitoring  bool     `json:"enable_monitoring"`   // 启用进程监控
}

// ProcessStats 进程统计信息
type ProcessStats struct {
	TotalListings    int64     `json:"total_listings"`
	TotalStarts      int64     `json:"total_starts"`
	TotalStops       int64     `json:"total_stops"`
	TotalKills       int64     `json:"total_kills"`
	MonitoredCount   int       `json:"monitored_count"`
	mu               sync.RWMutex `json:"-"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID         int      `json:"pid"`
	Name        string   `json:"name"`
	CmdLine     string   `json:"cmd_line"`
	Status      string   `json:"status"`
	CPUPercent  float64  `json:"cpu_percent"`
	MemoryMB    float64  `json:"memory_mb"`
	Username    string   `json:"username"`
	CreateTime  int64    `json:"create_time"`
}

// NewProcessManagerModule 创建进程管理模块实例
func NewProcessManagerModule(config ProcessConfig) (*ProcessManagerModule, error) {
	if config.MaxProcesses <= 0 {
		config.MaxProcesses = 1000 // 默认最多监控1000个进程
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 300 // 默认5分钟清理一次
	}

	stats := &ProcessStats{}

	module := &ProcessManagerModule{
		config: config,
		stats:  stats,
	}

	// 启动自动清理协程
	if config.EnableAutoCleanup {
		go module.autoCleanup()
	}

	return module, nil
}

// GetTools 返回进程管理模块的 MCP 工具列表
func (m *ProcessManagerModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 进程列表工具
		&processListTool{module: m},
		// 进程详情工具
		&processInfoTool{module: m},
		// 启动进程工具
		&processStartTool{module: m},
		// 停止进程工具
		&processStopTool{module: m},
		// 终止进程工具
		&processKillTool{module: m},
		// 系统信息工具
		&systemInfoTool{module: m},
	}

	return tools, nil
}

// 进程列表工具
type processListTool struct {
	module *ProcessManagerModule
}

func (t *processListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_list",
		Desc: "List all running processes on the system with optional filtering",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filter": {
				Type: "string",
				Desc: "Optional filter string (e.g., process name pattern)",
			},
			"limit": {
				Type: "integer",
				Desc: "Maximum number of processes to return (default: 100)",
			},
		}),
	}, nil
}

func (t *processListTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Filter string `json:"filter"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Limit <= 0 {
		args.Limit = 100
	}

	result, err := t.module.listProcesses(args.Filter, args.Limit)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 进程详情工具
type processInfoTool struct {
	module *ProcessManagerModule
}

func (t *processInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_info",
		Desc: "Get detailed information about a specific process",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pid": {
				Type:     "integer",
				Desc:     "Process ID",
				Required:  true,
			},
		}),
	}, nil
}

func (t *processInfoTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.getProcessInfo(args.PID)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 启动进程工具
type processStartTool struct {
	module *ProcessManagerModule
}

func (t *processStartTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_start",
		Desc: "Start a new process with the given command",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Type:     "string",
				Desc:     "Command to execute",
				Required:  true,
			},
			"args": {
				Type: "array",
				Desc: "Command arguments",
			},
			"detached": {
				Type: "boolean",
				Desc: "Run process in background (detached mode)",
			},
			"timeout": {
				Type: "integer",
				Desc: "Timeout in milliseconds for detached process wait",
			},
		}),
	}, nil
}

func (t *processStartTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Command   string   `json:"command"`
		Args      []string `json:"args"`
		Detached  bool     `json:"detached"`
		Timeout   int      `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.startProcess(args.Command, args.Args, args.Detached, args.Timeout)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 停止进程工具
type processStopTool struct {
	module *ProcessManagerModule
}

func (t *processStopTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_stop",
		Desc: "Gracefully stop a process by PID",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pid": {
				Type:     "integer",
				Desc:     "Process ID to stop",
				Required:  true,
			},
			"timeout": {
				Type: "integer",
				Desc:     "Timeout in milliseconds (default: 5000)",
			},
		}),
	}, nil
}

func (t *processStopTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		PID     int `json:"pid"`
		Timeout int `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Timeout <= 0 {
		args.Timeout = 5000
	}

	result, err := t.module.stopProcess(args.PID, args.Timeout)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 终止进程工具
type processKillTool struct {
	module *ProcessManagerModule
}

func (t *processKillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_kill",
		Desc: "Forcefully terminate a process by PID",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pid": {
				Type:     "integer",
				Desc:     "Process ID to kill",
				Required:  true,
			},
			"signal": {
				Type: "integer",
				Desc: "Signal number (default: 9 for SIGKILL)",
			},
		}),
	}, nil
}

func (t *processKillTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		PID    int `json:"pid"`
		Signal int `json:"signal"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Signal <= 0 {
		args.Signal = 9 // 默认 SIGKILL
	}

	result, err := t.module.killProcess(args.PID, args.Signal)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 系统信息工具
type systemInfoTool struct {
	module *ProcessManagerModule
}

func (t *systemInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "system_info",
		Desc:        "Get system information including OS, architecture, and resource usage",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *systemInfoTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.getSystemInfo()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭进程管理模块
func (m *ProcessManagerModule) Close() error {
	return nil
}

// GetStats 获取统计信息
func (m *ProcessManagerModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_listings": m.stats.TotalListings,
		"total_starts":    m.stats.TotalStarts,
		"total_stops":     m.stats.TotalStops,
		"total_kills":     m.stats.TotalKills,
	}
}

// ==================== 核心功能实现 ====================

// listProcesses 列出系统进程
func (m *ProcessManagerModule) listProcesses(filter string, limit int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalListings++
	m.stats.mu.Unlock()

	// 使用 ps 命令获取进程列表
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/fo", "csv", "/nh")
	} else {
		cmd = exec.Command("ps", "aux")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to list processes: %v", err),
		}, nil
	}

	processes, err := m.parseProcessList(string(output), filter, limit)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to parse process list: %v", err),
		}, nil
	}

	return map[string]any{
		"success":   true,
		"processes": processes,
		"count":     len(processes),
	}, nil
}

// parseProcessList 解析进程列表输出
func (m *ProcessManagerModule) parseProcessList(output string, filter string, limit int) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		if limit > 0 && len(processes) >= limit {
			break
		}

		info, err := m.parseProcessLine(line)
		if err != nil {
			continue
		}

		// 应用过滤器
		if filter != "" && !strings.Contains(strings.ToLower(info.Name), strings.ToLower(filter)) {
			continue
		}

		// 检查是否在允许列表中
		if len(m.config.AllowedProcesses) > 0 {
			allowed := false
			for _, allowedProc := range m.config.AllowedProcesses {
				if strings.Contains(strings.ToLower(info.Name), strings.ToLower(allowedProc)) {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// 检查是否在阻止列表中
		for _, blockedProc := range m.config.BlockedProcesses {
			if strings.Contains(strings.ToLower(info.Name), strings.ToLower(blockedProc)) {
				continue
			}
		}

		processes = append(processes, info)
	}

	return processes, nil
}

// parseProcessLine 解析单行进程信息
func (m *ProcessManagerModule) parseProcessLine(line string) (ProcessInfo, error) {
	var info ProcessInfo

	if runtime.GOOS == "windows" {
		// Windows tasklist CSV 格式
		fields := strings.Split(line, "\",\"")
		if len(fields) >= 2 {
			// 格式: "name","pid","session","mem"
			name := strings.Trim(fields[0], "\"")
			pidStr := strings.Trim(fields[1], "\"")
			pid, _ := strconv.Atoi(pidStr)

			info.Name = name
			info.PID = pid
			info.Status = "running"
		}
	} else {
		// Unix ps aux 格式
		fields := strings.Fields(line)
		if len(fields) >= 11 {
			// 格式: USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
			pid, _ := strconv.Atoi(fields[1])
			cpu, _ := strconv.ParseFloat(fields[2], 64)
			mem, _ := strconv.ParseFloat(fields[3], 64)

			info.Username = fields[0]
			info.PID = pid
			info.CPUPercent = cpu
			info.MemoryMB = mem
			info.Status = fields[7]
			info.CmdLine = strings.Join(fields[10:], " ")
			if len(fields[0]) > 15 {
				info.Name = fields[0][:15]
			} else {
				info.Name = fields[0]
			}
		}
	}

	return info, nil
}

// getProcessInfo 获取进程详细信息
func (m *ProcessManagerModule) getProcessInfo(pid int) (map[string]any, error) {
	// 使用 ps 命令获取详细信息
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %d", pid), "/fo", "list", "/v")
	} else {
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,user,%cpu,%mem,etime,comm")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Process not found: %d", pid),
		}, nil
	}

	info := ProcessInfo{
		PID:    pid,
		Name:    strings.TrimSpace(string(output)),
		Status:  "running",
		CmdLine: strings.TrimSpace(string(output)),
	}

	return map[string]any{
		"success": true,
		"process": info,
	}, nil
}

// startProcess 启动新进程
func (m *ProcessManagerModule) startProcess(command string, args []string, detached bool, timeout int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalStarts++
	m.stats.mu.Unlock()

	// 检查命令是否允许
	if !m.isCommandAllowed(command) {
		return map[string]any{
			"success": false,
			"error":   "Command not allowed by policy",
		}, nil
	}

	cmd := exec.Command(command, args...)
	if !detached {
		// 同步执行
		if timeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
			defer cancel()
			cmd = exec.CommandContext(ctx, command, args...)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
				"stdout":  string(output),
			}, nil
		}

		return map[string]any{
			"success": true,
			"stdout":  string(output),
			"detached": false,
		}, nil
	}

	// 后台执行
	if err := cmd.Start(); err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success":  true,
		"pid":      cmd.Process.Pid,
		"detached": true,
	}, nil
}

// stopProcess 停止进程
func (m *ProcessManagerModule) stopProcess(pid int, timeout int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalStops++
	m.stats.mu.Unlock()

	// 检查进程是否允许操作
	if !m.isProcessAllowed(pid) {
		return map[string]any{
			"success": false,
			"error":   "Process operation not allowed by policy",
		}, nil
	}

	// 查找进程
	process, err := os.FindProcess(pid)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Process not found: %d", pid),
		}, nil
	}

	// 发送中断信号
	if err := process.Signal(os.Interrupt); err != nil {
		// 如果中断失败，尝试终止
		if err := process.Kill(); err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to stop process: %v", err),
			}, nil
		}
	}

	// 等待进程结束
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-done:
		return map[string]any{
			"success": true,
			"pid":      pid,
			"action":   "stopped",
		}, nil
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		// 超时，强制终止
		process.Kill()
		return map[string]any{
			"success": true,
			"pid":     pid,
			"action":  "killed",
			"reason":  "timeout",
		}, nil
	}
}

// killProcess 终止进程
func (m *ProcessManagerModule) killProcess(pid int, signal int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalKills++
	m.stats.mu.Unlock()

	// 检查进程是否允许操作
	if !m.isProcessAllowed(pid) {
		return map[string]any{
			"success": false,
			"error":   "Process operation not allowed by policy",
		}, nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Process not found: %d", pid),
		}, nil
	}

	if err := process.Kill(); err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to kill process: %v", err),
		}, nil
	}

	return map[string]any{
		"success": true,
		"pid":      pid,
		"action":   "killed",
		"signal":   signal,
	}, nil
}

// getSystemInfo 获取系统信息
func (m *ProcessManagerModule) getSystemInfo() (map[string]any, error) {
	return map[string]any{
		"success": true,
		"system": map[string]any{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"hostname":     m.getHostname(),
			"cpu_count":    runtime.NumCPU(),
			"go_version":   runtime.Version(),
		},
		"resources": map[string]any{
			"memory_mb":        m.getMemoryUsage(),
			"goroutines":       runtime.NumGoroutine(),
		},
	}, nil
}

// getHostname 获取主机名
func (m *ProcessManagerModule) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getMemoryUsage 获取内存使用情况
func (m *ProcessManagerModule) getMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Sys / 1024 / 1024 // 转换为 MB
}

// isCommandAllowed 检查命令是否允许执行
func (m *ProcessManagerModule) isCommandAllowed(command string) bool {
	// 检查阻止列表
	for _, blockedProc := range m.config.BlockedProcesses {
		if strings.Contains(strings.ToLower(command), strings.ToLower(blockedProc)) {
			return false
		}
	}

	// 如果允许列表为空，默认允许
	if len(m.config.AllowedProcesses) == 0 {
		return true
	}

	// 检查允许列表
	for _, allowedProc := range m.config.AllowedProcesses {
		if strings.Contains(strings.ToLower(command), strings.ToLower(allowedProc)) {
			return true
		}
	}

	return false
}

// isProcessAllowed 检查进程操作是否允许
func (m *ProcessManagerModule) isProcessAllowed(pid int) bool {
	// 关键系统进程保护
	if pid < 10 {
		return false
	}

	// 获取进程名并检查
	info, err := m.getProcessInfo(pid)
	if err != nil {
		return false
	}

	if procData, ok := info["process"].(ProcessInfo); ok {
		return m.isCommandAllowed(procData.Name)
	}

	return true
}

// autoCleanup 自动清理僵尸进程
func (m *ProcessManagerModule) autoCleanup() {
	ticker := time.NewTicker(time.Duration(m.config.CleanupInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 这里可以实现僵尸进程检测和清理逻辑
		// 由于跨平台差异，这里仅作为框架预留
	}
}
