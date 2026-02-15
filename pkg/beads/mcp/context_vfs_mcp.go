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

package mcp

import (
	"context"
	"encoding/base64"
	"fmt"

	"AgentFramework/pkg/beads"
	beadscontext "AgentFramework/pkg/beads/context"
)

// VFSMCP VFS (虚拟文件系统) MCP 工具
// 提供 Agent 与 VFS 交互的 MCP 接口
type VFSMCP struct {
	tracker beads.TaskTracker
}

// NewVFSMCP 创建新的 VFS MCP 工具
func NewVFSMCP(tracker beads.TaskTracker) *VFSMCP {
	return &VFSMCP{
		tracker: tracker,
	}
}

// ===== Input/Output Structures =====

// ReadFileInput 读取文件的输入
type ReadFileInput struct {
	URI   string `json:"uri"`             // viking:// URI
	Layer string `json:"layer,omitempty"` // 层级 (l0/l1/l2/auto)
}

// ReadFileOutput 读取文件的输出
type ReadFileOutput struct {
	Success  bool   `json:"success"`
	Content  string `json:"content,omitempty"`  // Base64 编码的内容
	Size     int64  `json:"size,omitempty"`     // 文件大小
	Encoding string `json:"encoding,omitempty"` // 编码方式
	Error    string `json:"error,omitempty"`
}

// WriteFileInput 写入文件的输入
type WriteFileInput struct {
	URI     string `json:"uri"`              // viking:// URI
	Content string `json:"content"`          // Base64 编码的内容
	Layer   string `json:"layer,omitempty"`  // 目标层级
}

// WriteFileOutput 写入文件的输出
type WriteFileOutput struct {
	Success bool   `json:"success"`
	Size    int64  `json:"size,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListFilesInput 列出文件的输入
type ListFilesInput struct {
	URI string `json:"uri"` // viking:// URI (目录)
}

// ListFilesOutput 列出文件的输出
type ListFilesOutput struct {
	Success bool                      `json:"success"`
	Files   []*beadscontext.VFSFileInfo    `json:"files,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// DeleteFileInput 删除文件的输入
type DeleteFileInput struct {
	URI string `json:"uri"` // viking:// URI
}

// DeleteFileOutput 删除文件的输出
type DeleteFileOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MkdirInput 创建目录的输入
type MkdirInput struct {
	URI     string `json:"uri"`              // viking:// URI
	Parents bool   `json:"parents,omitempty"` // 递归创建父目录
}

// MkdirOutput 创建目录的输出
type MkdirOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MoveInput 移动文件的输入
type MoveInput struct {
	OldPath string `json:"old_path"` // 源 URI
	NewPath string `json:"new_path"` // 目标 URI
}

// MoveOutput 移动文件的输出
type MoveOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// VFSSearchInput VFS 搜索文件的输入
type VFSSearchInput struct {
	Query      string  `json:"query"`                // 搜索查询
	Layer      string  `json:"layer,omitempty"`      // 搜索层级
	MaxResults int     `json:"max_results,omitempty"` // 最大结果数
	MinScore   float64 `json:"min_score,omitempty"`  // 最小相似度分数
}

// VFSSearchOutput VFS 搜索的输出
type VFSSearchOutput struct {
	Success bool                        `json:"success"`
	Results []*beadscontext.VFSSearchResult  `json:"results,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// ParseURIInput 解析 URI 的输入
type ParseURIInput struct {
	URI string `json:"uri"` // viking:// URI
}

// ParseURIOutput 解析 URI 的输出
type ParseURIOutput struct {
	Success bool               `json:"success"`
	Path    *beadscontext.VFSPath   `json:"path,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// ===== MCP Tool Implementations =====

// ReadFile MCP 工具：从 VFS 读取文件内容
func (vm *VFSMCP) ReadFile(
	ctx context.Context,
	input *ReadFileInput,
) (*ReadFileOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &ReadFileOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	// 设置默认层级
	if input.Layer == "" {
		input.Layer = "auto"
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &ReadFileOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 读取文件
	layer := beadscontext.LayerType(input.Layer)
	data, err := vfs.ReadFile(ctx, input.URI, layer)
	if err != nil {
		return &ReadFileOutput{
			Success: false,
			Error:   fmt.Sprintf("read file failed: %v", err),
		}, nil
	}

	// Base64 编码内容
	encoded := base64.StdEncoding.EncodeToString(data)

	return &ReadFileOutput{
		Success:  true,
		Content:  encoded,
		Size:     int64(len(data)),
		Encoding: "base64",
	}, nil
}

// WriteFile MCP 工具：向 VFS 写入文件内容
func (vm *VFSMCP) WriteFile(
	ctx context.Context,
	input *WriteFileInput,
) (*WriteFileOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &WriteFileOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	if input.Content == "" {
		return &WriteFileOutput{
			Success: false,
			Error:   "content is required",
		}, nil
	}

	// 解码 Base64 内容
	data, err := base64.StdEncoding.DecodeString(input.Content)
	if err != nil {
		return &WriteFileOutput{
			Success: false,
			Error:   fmt.Sprintf("invalid base64 content: %v", err),
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &WriteFileOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 写入文件
	if err := vfs.WriteFile(ctx, input.URI, data); err != nil {
		return &WriteFileOutput{
			Success: false,
			Error:   fmt.Sprintf("write file failed: %v", err),
		}, nil
	}

	return &WriteFileOutput{
		Success: true,
		Size:    int64(len(data)),
	}, nil
}

// ListFiles MCP 工具：列出目录中的文件
func (vm *VFSMCP) ListFiles(
	ctx context.Context,
	input *ListFilesInput,
) (*ListFilesOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &ListFilesOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &ListFilesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 列出文件
	files, err := vfs.ListFiles(ctx, input.URI)
	if err != nil {
		return &ListFilesOutput{
			Success: false,
			Error:   fmt.Sprintf("list files failed: %v", err),
		}, nil
	}

	return &ListFilesOutput{
		Success: true,
		Files:   files,
	}, nil
}

// DeleteFile MCP 工具：删除文件
func (vm *VFSMCP) DeleteFile(
	ctx context.Context,
	input *DeleteFileInput,
) (*DeleteFileOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &DeleteFileOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &DeleteFileOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 删除文件
	if err := vfs.DeleteFile(ctx, input.URI); err != nil {
		return &DeleteFileOutput{
			Success: false,
			Error:   fmt.Sprintf("delete file failed: %v", err),
		}, nil
	}

	return &DeleteFileOutput{
		Success: true,
	}, nil
}

// Mkdir MCP 工具：创建目录
func (vm *VFSMCP) Mkdir(
	ctx context.Context,
	input *MkdirInput,
) (*MkdirOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &MkdirOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &MkdirOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 创建目录
	if input.Parents {
		err = vfs.MkdirAll(ctx, input.URI)
	} else {
		err = vfs.Mkdir(ctx, input.URI)
	}

	if err != nil {
		return &MkdirOutput{
			Success: false,
			Error:   fmt.Sprintf("mkdir failed: %v", err),
		}, nil
	}

	return &MkdirOutput{
		Success: true,
	}, nil
}

// Move MCP 工具：移动或重命名文件
func (vm *VFSMCP) Move(
	ctx context.Context,
	input *MoveInput,
) (*MoveOutput, error) {
	// 验证输入
	if input.OldPath == "" {
		return &MoveOutput{
			Success: false,
			Error:   "old_path is required",
		}, nil
	}

	if input.NewPath == "" {
		return &MoveOutput{
			Success: false,
			Error:   "new_path is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &MoveOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 移动文件
	if err := vfs.Move(ctx, input.OldPath, input.NewPath); err != nil {
		return &MoveOutput{
			Success: false,
			Error:   fmt.Sprintf("move failed: %v", err),
		}, nil
	}

	return &MoveOutput{
		Success: true,
	}, nil
}

// Search MCP 工具：搜索文件
func (vm *VFSMCP) Search(
	ctx context.Context,
	input *VFSSearchInput,
) (*VFSSearchOutput, error) {
	// 验证输入
	if input.Query == "" {
		return &VFSSearchOutput{
			Success: false,
			Error:   "query is required",
		}, nil
	}

	// 设置默认值
	if input.Layer == "" {
		input.Layer = "auto"
	}
	if input.MaxResults <= 0 {
		input.MaxResults = 10
	}
	if input.MinScore <= 0 {
		input.MinScore = 0.5
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &VFSSearchOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 构建搜索选项
	opts := []beadscontext.SearchOption{
		beadscontext.WithSearchLayer(beadscontext.LayerType(input.Layer)),
		beadscontext.WithMaxResults(input.MaxResults),
		beadscontext.WithMinScore(input.MinScore),
	}

	// 执行搜索
	results, err := vfs.Search(ctx, input.Query, opts...)
	if err != nil {
		return &VFSSearchOutput{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, nil
	}

	return &VFSSearchOutput{
		Success: true,
		Results: results,
	}, nil
}

// ParseURI MCP 工具：解析 viking:// URI
func (vm *VFSMCP) ParseURI(
	ctx context.Context,
	input *ParseURIInput,
) (*ParseURIOutput, error) {
	// 验证输入
	if input.URI == "" {
		return &ParseURIOutput{
			Success: false,
			Error:   "uri is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &ParseURIOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 解析 URI
	path, err := vfs.ParseURI(input.URI)
	if err != nil {
		return &ParseURIOutput{
			Success: false,
			Error:   fmt.Sprintf("parse URI failed: %v", err),
		}, nil
	}

	return &ParseURIOutput{
		Success: true,
		Path:    path,
	}, nil
}

// BuildURI MCP 工具：构建 viking:// URI
type BuildURIInput struct {
	Scheme    string            `json:"scheme"`               // URI scheme (通常为 "viking")
	Path      string            `json:"path"`                 // 路径
	Workspace string            `json:"workspace,omitempty"`  // 工作区
	Layer     string            `json:"layer,omitempty"`      // 层级
	Query     map[string]string `json:"query,omitempty"`      // 查询参数
}

type BuildURIOutput struct {
	Success bool   `json:"success"`
	URI     string `json:"uri,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (vm *VFSMCP) BuildURI(
	ctx context.Context,
	input *BuildURIInput,
) (*BuildURIOutput, error) {
	// 验证输入
	if input.Scheme == "" {
		input.Scheme = "viking"
	}
	if input.Path == "" {
		return &BuildURIOutput{
			Success: false,
			Error:   "path is required",
		}, nil
	}

	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &BuildURIOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 简化实现：不使用选项构建基础 URI
	uri, err := vfs.BuildURI(input.Scheme, input.Path)
	if err != nil {
		return &BuildURIOutput{
			Success: false,
			Error:   fmt.Sprintf("build URI failed: %v", err),
		}, nil
	}

	// TODO: 添加对 workspace、layer 和 query 参数的支持

	return &BuildURIOutput{
		Success: true,
		URI:     uri,
	}, nil
}

// GetVFSInfo MCP 工具：获取 VFS 信息
type GetVFSInfoOutput struct {
	Success  bool   `json:"success"`
	Started  bool   `json:"started,omitempty"`
	CacheSize int  `json:"cache_size,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (vm *VFSMCP) GetVFSInfo(
	ctx context.Context,
) (*GetVFSInfoOutput, error) {
	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &GetVFSInfoOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取 VFS 信息
	type vfsInformer interface {
		IsStarted() bool
		GetCacheSize() int
	}

	if infoVFS, ok := vfs.(vfsInformer); ok {
		return &GetVFSInfoOutput{
			Success:   true,
			Started:   infoVFS.IsStarted(),
			CacheSize: infoVFS.GetCacheSize(),
		}, nil
	}

	return &GetVFSInfoOutput{
		Success: true,
		Started: true,
	}, nil
}

// ClearCache MCP 工具：清空 VFS 缓存
type ClearCacheOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (vm *VFSMCP) ClearCache(
	ctx context.Context,
) (*ClearCacheOutput, error) {
	// 获取 VFS
	vfs, err := vm.getVFS(ctx)
	if err != nil {
		return &ClearCacheOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 清空缓存
	type cacheCleaner interface {
		ClearCache()
	}

	if cleanerVFS, ok := vfs.(cacheCleaner); ok {
		cleanerVFS.ClearCache()
		return &ClearCacheOutput{
			Success: true,
		}, nil
	}

	return &ClearCacheOutput{
		Success: false,
		Error:   "cache clearing not supported",
	}, nil
}

// ===== 辅助方法 =====

// getVFS 获取 VFS 实例
func (vm *VFSMCP) getVFS(ctx context.Context) (beadscontext.VFS, error) {
	type vfsTracker interface {
		IsContextEnabled() bool
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := vm.tracker.(vfsTracker)
	if !ok {
		return nil, fmt.Errorf("VFS operations not supported")
	}

	if !tracker.IsContextEnabled() {
		return nil, fmt.Errorf("context system is disabled")
	}

	store := tracker.GetContextStore()
	if store == nil {
		return nil, fmt.Errorf("context store is not available")
	}

	// 尝试从存储获取 VFS
	type vfsStore interface {
		GetVFS(scheme string) (beadscontext.VFS, error)
	}

	if vfsStore, ok := store.(vfsStore); ok {
		vfs, err := vfsStore.GetVFS("viking")
		if err != nil {
			return nil, fmt.Errorf("get VFS failed: %w", err)
		}
		return vfs, nil
	}

	return nil, fmt.Errorf("VFS not available in context store")
}
