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

package agent

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ProcessStatus 进程状态
type ProcessStatus int

const (
	ProcessStatusCreated ProcessStatus = iota
	ProcessStatusStarting
	ProcessStatusRunning
	ProcessStatusStopping
	ProcessStatusStopped
	ProcessStatusError
)

// ProcessStateHandler 进程状态变化处理器
type ProcessStateHandler func(process *ManagedProcess, oldStatus, newStatus ProcessStatus)

// ManagedProcess 管理的进程
type ManagedProcess struct {
	PID         int
	Command     string
	Args        []string
	Env         []string
	Status      ProcessStatus
	ExitCode    int
	CreatedAt   time.Time
	StartedAt   time.Time
	StoppedAt   time.Time
	Stdout      string
	Stderr      string
}

// ProcessManager 进程管理器
type ProcessManager interface {
	// 启动进程
	StartProcess(ctx context.Context, cmd *exec.Cmd) (*ManagedProcess, error)

	// 停止进程
	StopProcess(ctx context.Context, pid int) error

	// 获取进程信息
	GetProcessInfo(pid int) (*ManagedProcess, error)

	// 列出所有进程
	ListProcesses() []*ManagedProcess

	// 监听进程状态
	WatchProcess(pid int, handler ProcessStateHandler) error

	// 移除进程监听
	UnwatchProcess(pid int, handler ProcessStateHandler) error

	// 清理停止的进程
	CleanupStoppedProcesses() error
}

// DefaultProcessManager 进程管理器的默认实现
type DefaultProcessManager struct {
	processes map[int]*ManagedProcess
	watchers  map[int][]ProcessStateHandler
	mu        sync.RWMutex
}

// NewProcessManager 创建一个新的进程管理器实例
func NewProcessManager() ProcessManager {
	return &DefaultProcessManager{
		processes: make(map[int]*ManagedProcess),
		watchers:  make(map[int][]ProcessStateHandler),
	}
}

// StartProcess 启动进程
func (pm *DefaultProcessManager) StartProcess(ctx context.Context, cmd *exec.Cmd) (*ManagedProcess, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 创建进程管理对象
	process := &ManagedProcess{
		Command:   cmd.Path,
		Args:      cmd.Args,
		Env:       cmd.Env,
		Status:    ProcessStatusCreated,
		CreatedAt: time.Now(),
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		process.Status = ProcessStatusError
		return process, err
	}

	// 更新进程信息
	process.PID = cmd.Process.Pid
	process.Status = ProcessStatusRunning
	process.StartedAt = time.Now()

	// 保存到进程映射
	pm.processes[process.PID] = process

	// 启动进程监控 goroutine
	go pm.monitorProcess(cmd.Process, process.PID)

	return process, nil
}

// StopProcess 停止进程
func (pm *DefaultProcessManager) StopProcess(ctx context.Context, pid int) error {
	pm.mu.Lock()
	process, exists := pm.processes[pid]
	if !exists {
		pm.mu.Unlock()
		return nil
	}

	// 检查进程是否已经停止
	if process.Status == ProcessStatusStopped || process.Status == ProcessStatusError {
		pm.mu.Unlock()
		return nil
	}

	// 标记为正在停止
	process.Status = ProcessStatusStopping
	pm.mu.Unlock()

	// 尝试优雅停止
	if err := pm.terminateProcess(pid); err != nil {
		return err
	}

	return nil
}

// GetProcessInfo 获取进程信息
func (pm *DefaultProcessManager) GetProcessInfo(pid int) (*ManagedProcess, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	process, exists := pm.processes[pid]
	if !exists {
		return nil, nil
	}

	// 返回副本以避免外部修改
	copyProcess := *process
	return &copyProcess, nil
}

// ListProcesses 列出所有进程
func (pm *DefaultProcessManager) ListProcesses() []*ManagedProcess {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var processes []*ManagedProcess
	for _, process := range pm.processes {
		copyProcess := *process
		processes = append(processes, &copyProcess)
	}

	return processes
}

// WatchProcess 监听进程状态
func (pm *DefaultProcessManager) WatchProcess(pid int, handler ProcessStateHandler) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processes[pid]; !exists {
		return nil
	}

	pm.watchers[pid] = append(pm.watchers[pid], handler)
	return nil
}

// UnwatchProcess 移除进程监听
func (pm *DefaultProcessManager) UnwatchProcess(pid int, handler ProcessStateHandler) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 移除所有监听器
	delete(pm.watchers, pid)

	return nil
}

// CleanupStoppedProcesses 清理停止的进程
func (pm *DefaultProcessManager) CleanupStoppedProcesses() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var toRemove []int
	for pid, process := range pm.processes {
		if process.Status == ProcessStatusStopped || process.Status == ProcessStatusError {
			// 检查进程是否已经停止超过一定时间
			if time.Since(process.StoppedAt) > time.Minute*5 {
				toRemove = append(toRemove, pid)
			}
		}
	}

	for _, pid := range toRemove {
		delete(pm.processes, pid)
		delete(pm.watchers, pid)
	}

	return nil
}

// monitorProcess 监控进程状态
func (pm *DefaultProcessManager) monitorProcess(osProcess *os.Process, pid int) {
	var err error
	state, err := osProcess.Wait()

	pm.mu.Lock()
	process, exists := pm.processes[pid]
	if exists {
		oldStatus := process.Status

		if err != nil {
			process.Status = ProcessStatusError
		} else {
			process.Status = ProcessStatusStopped
			process.ExitCode = state.ExitCode()
		}
		process.StoppedAt = time.Now()

		// 通知监听器
		if handlers, watcherExists := pm.watchers[pid]; watcherExists {
			for _, handler := range handlers {
				go handler(process, oldStatus, process.Status)
			}
		}
	}
	pm.mu.Unlock()
}

// terminateProcess 终止进程
func (pm *DefaultProcessManager) terminateProcess(pid int) error {
	pm.mu.RLock()
	process, exists := pm.processes[pid]
	pm.mu.RUnlock()

	if !exists || process.Status == ProcessStatusStopped {
		return nil
	}

	// 尝试发送终止信号
	if err := pm.sendTerminationSignal(pid); err != nil {
		return err
	}

	// 等待进程退出
	if err := pm.waitForProcessExit(pid, 30*time.Second); err != nil {
		// 如果优雅停止失败，强制杀死进程
		if err := pm.forceKillProcess(pid); err != nil {
			return err
		}
	}

	return nil
}

// sendTerminationSignal 发送终止信号
func (pm *DefaultProcessManager) sendTerminationSignal(pid int) error {
	// 这里可以根据操作系统发送不同的信号
	// 例如在 Unix 上发送 SIGTERM，Windows 上发送 WM_CLOSE
	return nil
}

// waitForProcessExit 等待进程退出
func (pm *DefaultProcessManager) waitForProcessExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pm.mu.RLock()
		process, exists := pm.processes[pid]
		pm.mu.RUnlock()

		if !exists || process.Status == ProcessStatusStopped {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// forceKillProcess 强制杀死进程
func (pm *DefaultProcessManager) forceKillProcess(pid int) error {
	pm.mu.RLock()
	process, exists := pm.processes[pid]
	pm.mu.RUnlock()

	if !exists || process.Status == ProcessStatusStopped {
		return nil
	}

	// 这里可以根据操作系统强制杀死进程
	// 例如在 Unix 上发送 SIGKILL，Windows 上使用 TerminateProcess
	return nil
}
