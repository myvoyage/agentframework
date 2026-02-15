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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"AgentFramework/pkg/tools/sandbox/file"

	"github.com/cloudwego/eino/components/tool"
)

// ResourceQuota 定义沙箱资源配额
// 用于限制沙箱的资源使用，包括文件大小、文件数量、CPU时间、内存使用等
type ResourceQuota struct {
	MaxFileSize     int64 // 单个文件的最大大小，单位为字节
	MaxTotalSize    int64 // 沙箱内所有文件的总大小，单位为字节
	MaxFileCount    int   // 沙箱内允许的最大文件数量
	MaxCPUSeconds   int   // 允许的最大CPU时间，单位为秒
	MaxMemoryBytes  int64 // 允许的最大内存使用，单位为字节
	MaxProcessCount int   // 允许的最大进程数量
}

// SandboxManager 沙箱管理器接口，负责文件访问控制和管理
// 实现与 pkg/tools/sandbox/file.FileModule 的集成，提供路径验证、沙箱目录管理等功能
// 用于确保 AI Agent 只能访问授权的文件路径，防止越权访问
type SandboxManager interface {
	// ValidatePath 验证路径是否合法，返回规范化的绝对路径
	// 如果路径不合法或超出沙箱范围，返回错误
	ValidatePath(ctx context.Context, path string) (string, error)

	// GetSandboxDir 获取沙箱根目录
	GetSandboxDir(ctx context.Context) string

	// WithFileModule 设置文件模块，返回更新后的 SandboxManager
	WithFileModule(fileModule *file.FileModule) SandboxManager

	// CreateFileTools 创建安全的文件操作工具
	// 所有文件操作都会经过路径验证
	CreateFileTools(ctx context.Context) (map[string]tool.BaseTool, error)

	// WithResourceQuota 设置资源配额，返回更新后的 SandboxManager
	WithResourceQuota(quota ResourceQuota) SandboxManager

	// GetResourceQuota 获取当前资源配额
	GetResourceQuota(ctx context.Context) ResourceQuota

	// CheckResourceQuota 检查资源使用是否符合配额
	// 返回剩余可用资源和是否超出配额
	CheckResourceQuota(ctx context.Context) (map[string]interface{}, bool, error)

	// UpdateResourceUsage 更新资源使用情况
	UpdateResourceUsage(ctx context.Context, usage map[string]int64) error

	// Close 关闭沙箱管理器，释放资源
	Close() error
}

// ResourceUsage 定义沙箱资源使用情况
type ResourceUsage struct {
	TotalFileSize int64 // 当前总文件大小，单位为字节
	FileCount     int   // 当前文件数量
	CPUSeconds    int64 // 当前CPU使用时间，单位为秒
	MemoryBytes   int64 // 当前内存使用，单位为字节
	ProcessCount  int   // 当前进程数量
}

// sandboxManagerImpl 沙箱管理器的具体实现
type sandboxManagerImpl struct {
	sandboxDir       string           // 沙箱根目录
	fileModule       *file.FileModule // 关联的文件模块
	resourceQuota    ResourceQuota    // 资源配额
	resourceUsage    ResourceUsage    // 资源使用情况
	hasResourceQuota bool             // 是否设置了资源配额
}

// NewSandboxManager 创建一个新的沙箱管理器实例
// sandboxDir 是沙箱的根目录，所有文件操作都将限制在该目录内
func NewSandboxManager(sandboxDir string) SandboxManager {
	return &sandboxManagerImpl{
		sandboxDir:       filepath.Clean(sandboxDir),
		fileModule:       nil,
		resourceQuota:    ResourceQuota{},
		resourceUsage:    ResourceUsage{},
		hasResourceQuota: false,
	}
}

// WithResourceQuota 设置资源配额，返回更新后的 SandboxManager
func (m *sandboxManagerImpl) WithResourceQuota(quota ResourceQuota) SandboxManager {
	m.resourceQuota = quota
	m.hasResourceQuota = true
	return m
}

// GetResourceQuota 获取当前资源配额
func (m *sandboxManagerImpl) GetResourceQuota(ctx context.Context) ResourceQuota {
	return m.resourceQuota
}

// CheckResourceQuota 检查资源使用是否符合配额
// 返回剩余可用资源和是否超出配额
func (m *sandboxManagerImpl) CheckResourceQuota(ctx context.Context) (map[string]interface{}, bool, error) {
	if !m.hasResourceQuota {
		// 如果没有设置资源配额，默认允许所有资源使用
		return map[string]interface{}{},
			false, nil
	}

	// 计算剩余可用资源
	remaining := map[string]interface{}{
		"maxFileSize":     m.resourceQuota.MaxFileSize - int64(m.resourceUsage.FileCount)*m.resourceQuota.MaxFileSize,
		"maxTotalSize":    m.resourceQuota.MaxTotalSize - m.resourceUsage.TotalFileSize,
		"maxFileCount":    m.resourceQuota.MaxFileCount - m.resourceUsage.FileCount,
		"maxCPUSeconds":   m.resourceQuota.MaxCPUSeconds - int(m.resourceUsage.CPUSeconds),
		"maxMemoryBytes":  m.resourceQuota.MaxMemoryBytes - m.resourceUsage.MemoryBytes,
		"maxProcessCount": m.resourceQuota.MaxProcessCount - m.resourceUsage.ProcessCount,
	}

	// 检查是否超出配额
	exceeded := false
	if m.resourceQuota.MaxTotalSize > 0 && m.resourceUsage.TotalFileSize > m.resourceQuota.MaxTotalSize {
		exceeded = true
	}
	if m.resourceQuota.MaxFileCount > 0 && m.resourceUsage.FileCount > m.resourceQuota.MaxFileCount {
		exceeded = true
	}
	if m.resourceQuota.MaxCPUSeconds > 0 && int(m.resourceUsage.CPUSeconds) > m.resourceQuota.MaxCPUSeconds {
		exceeded = true
	}
	if m.resourceQuota.MaxMemoryBytes > 0 && m.resourceUsage.MemoryBytes > m.resourceQuota.MaxMemoryBytes {
		exceeded = true
	}
	if m.resourceQuota.MaxProcessCount > 0 && m.resourceUsage.ProcessCount > m.resourceQuota.MaxProcessCount {
		exceeded = true
	}

	return remaining, exceeded, nil
}

// UpdateResourceUsage 更新资源使用情况
func (m *sandboxManagerImpl) UpdateResourceUsage(ctx context.Context, usage map[string]int64) error {
	// 遍历更新资源使用情况
	for key, value := range usage {
		switch key {
		case "totalFileSize":
			m.resourceUsage.TotalFileSize += value
		case "fileCount":
			m.resourceUsage.FileCount += int(value)
		case "cpuSeconds":
			m.resourceUsage.CPUSeconds += value
		case "memoryBytes":
			m.resourceUsage.MemoryBytes += value
		case "processCount":
			m.resourceUsage.ProcessCount += int(value)
		default:
			return fmt.Errorf("unknown resource usage key: %s", key)
		}
	}

	// 检查资源使用是否超出配额
	if m.hasResourceQuota {
		_, exceeded, err := m.CheckResourceQuota(ctx)
		if err != nil {
			return err
		}
		if exceeded {
			return errors.New("resource quota exceeded")
		}
	}

	return nil
}

// ValidatePath 验证路径是否合法，返回规范化的绝对路径
func (m *sandboxManagerImpl) ValidatePath(ctx context.Context, path string) (string, error) {
	// 如果路径为空，返回当前沙箱目录
	if path == "" {
		return m.sandboxDir, nil
	}

	// 规范化路径
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(m.sandboxDir, path)
	}
	absPath = filepath.Clean(absPath)

	// 检查路径是否在沙箱目录内
	if !strings.HasPrefix(absPath, m.sandboxDir) {
		return "", errors.New("path is outside sandbox directory")
	}

	// 防止路径遍历攻击
	relPath, err := filepath.Rel(m.sandboxDir, absPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		return "", errors.New("path traversal detected")
	}

	return absPath, nil
}

// GetSandboxDir 获取沙箱根目录
func (m *sandboxManagerImpl) GetSandboxDir(ctx context.Context) string {
	return m.sandboxDir
}

// WithFileModule 设置文件模块，返回更新后的 SandboxManager
func (m *sandboxManagerImpl) WithFileModule(fileModule *file.FileModule) SandboxManager {
	m.fileModule = fileModule
	return m
}

// CreateFileTools 创建安全的文件操作工具
func (m *sandboxManagerImpl) CreateFileTools(ctx context.Context) (map[string]tool.BaseTool, error) {
	if m.fileModule == nil {
		return nil, errors.New("file module not set")
	}

	// 获取文件模块的工具
	tools, err := m.fileModule.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file tools: %w", err)
	}

	// 将工具转换为 map
	toolMap := make(map[string]tool.BaseTool)
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool info: %w", err)
		}
		toolMap[info.Name] = t
	}

	return toolMap, nil
}

// Close 关闭沙箱管理器，释放资源
func (m *sandboxManagerImpl) Close() error {
	if m.fileModule != nil {
		return m.fileModule.Close()
	}
	return nil
}
