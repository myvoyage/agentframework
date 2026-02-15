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

// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// CodeScanner 代码扫描器，用于检测危险操作
type CodeScanner struct {
	dangerousPatterns map[string][]string
}

// NewCodeScanner 创建代码扫描器
func NewCodeScanner() *CodeScanner {
	return &CodeScanner{
		dangerousPatterns: map[string][]string{
			"python": {
				"os.system",
				"subprocess.call",
				"subprocess.run",
				"subprocess.Popen",
				"eval(",
				"exec(",
				"__import__",
				"open(",
				"file(",
				"input(",
				"raw_input(",
			},
			"javascript": {
				"eval(",
				"Function(",
				"require(",
				"import(",
				"process.exit",
				"child_process",
				"fs.readFile",
				"fs.writeFile",
			},
			"bash": {
				"rm ",
				"rmdir",
				"shutdown",
				"reboot",
				"mkfs",
				"dd ",
				"format",
				"> /dev/",
			},
			"go": {
				"os.Exit",
				"os.Remove",
				"os.RemoveAll",
				"exec.Command",
				"syscall",
				"unsafe",
			},
		},
	}
}

// Scan 扫描代码中的危险操作
func (s *CodeScanner) Scan(language, code string) ([]string, error) {
	language = strings.ToLower(language)
	patterns, exists := s.dangerousPatterns[language]
	if !exists {
		return nil, nil // 不支持的语言，跳过扫描
	}

	var warnings []string
	for _, pattern := range patterns {
		if strings.Contains(code, pattern) {
			warnings = append(warnings, fmt.Sprintf("Detected potentially dangerous operation: %s", pattern))
		}
	}

	return warnings, nil
}

// ResourceLimiter 资源限制器
type ResourceLimiter struct {
	maxMemoryMB      int
	maxCPUCores      int
	maxExecutionTime time.Duration
	activeExecutions int
	maxConcurrent    int
	mu               sync.RWMutex
}

// NewResourceLimiter 创建资源限制器
func NewResourceLimiter(maxMemoryMB, maxCPUCores int, maxExecutionTime time.Duration) *ResourceLimiter {
	if maxConcurrent := runtime.NumCPU(); maxCPUCores > maxConcurrent {
		maxCPUCores = maxConcurrent
	}

	return &ResourceLimiter{
		maxMemoryMB:      maxMemoryMB,
		maxCPUCores:      maxCPUCores,
		maxExecutionTime: maxExecutionTime,
		maxConcurrent:    maxCPUCores * 2, // 允许 CPU 核心数的 2 倍并发
	}
}

// AcquireExecution 获取执行权限
func (r *ResourceLimiter) AcquireExecution(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查并发限制
	if r.activeExecutions >= r.maxConcurrent {
		return fmt.Errorf("max concurrent executions reached: %d", r.maxConcurrent)
	}

	r.activeExecutions++
	return nil
}

// ReleaseExecution 释放执行权限
func (r *ResourceLimiter) ReleaseExecution() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeExecutions > 0 {
		r.activeExecutions--
	}
}

// GetActiveExecutions 获取当前活跃执行数
func (r *ResourceLimiter) GetActiveExecutions() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeExecutions
}

// GetMaxConcurrent 获取最大并发数
func (r *ResourceLimiter) GetMaxConcurrent() int {
	return r.maxConcurrent
}

// GetMemoryLimit 获取内存限制（MB）
func (r *ResourceLimiter) GetMemoryLimit() int {
	return r.maxMemoryMB
}

// GetCPULimit 获取 CPU 限制（核心数）
func (r *ResourceLimiter) GetCPULimit() int {
	return r.maxCPUCores
}

// GetExecutionTimeLimit 获取执行时间限制
func (r *ResourceLimiter) GetExecutionTimeLimit() time.Duration {
	return r.maxExecutionTime
}

// GetResourceStats 获取资源统计信息
func (r *ResourceLimiter) GetResourceStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"active_executions": r.activeExecutions,
		"max_concurrent":    r.maxConcurrent,
		"max_memory_mb":     r.maxMemoryMB,
		"max_cpu_cores":     r.maxCPUCores,
		"current_memory_mb": int(m.Alloc / 1024 / 1024),
		"num_goroutines":    runtime.NumGoroutine(),
		"num_cpu":           runtime.NumCPU(),
	}
}

// ValidateResourceLimits 验证资源限制配置
func ValidateResourceLimits(memoryLimitMB, cpuLimit int) error {
	if memoryLimitMB < 0 {
		return fmt.Errorf("memory limit must be non-negative")
	}
	if cpuLimit < 0 {
		return fmt.Errorf("CPU limit must be non-negative")
	}
	if memoryLimitMB > 0 && memoryLimitMB < 64 {
		return fmt.Errorf("memory limit too low, minimum 64MB recommended")
	}
	return nil
}
