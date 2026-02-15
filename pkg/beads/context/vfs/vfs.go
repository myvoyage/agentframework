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

package vfs

import (
	stdcontext "context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"AgentFramework/pkg/beads/context"
)

// VirtualFileSystem 虚拟文件系统实现
// 使用 viking:// URI scheme 提供类似文件系统的操作
type VirtualFileSystem struct {
	store    context.ContextStore
	mu       sync.RWMutex
	cache    map[string]*cachedContent
	cacheTTL time.Duration
	started  bool
}

// cachedContent 缓存的内容
type cachedContent struct {
	content   []byte
	timestamp time.Time
	layer     context.LayerType
}

// NewVirtualFileSystem 创建新的虚拟文件系统
func NewVirtualFileSystem(store context.ContextStore) *VirtualFileSystem {
	return &VirtualFileSystem{
		store:    store,
		cache:    make(map[string]*cachedContent),
		cacheTTL: 5 * time.Minute,
		started:  false,
	}
}

// Start 启动 VFS
func (vfs *VirtualFileSystem) Start(ctx stdcontext.Context) error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	if vfs.started {
		return fmt.Errorf("VFS already started")
	}

	vfs.started = true

	// 启动缓存清理协程
	go vfs.cacheCleanupLoop()

	return nil
}

// Stop 停止 VFS
func (vfs *VirtualFileSystem) Stop(ctx stdcontext.Context) error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	if !vfs.started {
		return nil
	}

	vfs.cache = make(map[string]*cachedContent)
	vfs.started = false

	return nil
}

// ===== URI 操作 =====

// ParseURI 解析 URI
func (vfs *VirtualFileSystem) ParseURI(uri string) (*context.VFSPath, error) {
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return nil, err
	}

	return &context.VFSPath{
		Scheme:    vuri.Scheme(),
		Workspace: vuri.Workspace(),
		Path:      vuri.Path(),
		Layer:     vuri.Layer(),
		Query:     vuri.Query(),
	}, nil
}

// BuildURI 构建 URI
func (vfs *VirtualFileSystem) BuildURI(scheme, path string, opts ...URIOption) (string, error) {
	vuri := &VikingURI{
		scheme: scheme,
		path:   path,
		query:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(vuri)
	}

	return vuri.String(), nil
}

// ===== 文件操作 =====

// ReadFile 读取文件内容
func (vfs *VirtualFileSystem) ReadFile(ctx stdcontext.Context, uri string, layer context.LayerType) ([]byte, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	if !vfs.started {
		return nil, fmt.Errorf("VFS not started")
	}

	// 检查缓存
	cacheKey := uri + ":" + string(layer)
	if cached, ok := vfs.cache[cacheKey]; ok && time.Since(cached.timestamp) < vfs.cacheTTL {
		return cached.content, nil
	}

	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return nil, fmt.Errorf("parse URI: %w", err)
	}

	// 确定要读取的层级
	readLayer := layer
	if layer == context.LayerAuto {
		readLayer = vuri.Layer()
	}

	// 从上下文存储读取
	contextID := vfs.pathToContextID(vuri.FullPath())
	contentLayer, err := vfs.store.GetLayer(ctx, contextID, readLayer)
	if err != nil {
		// 尝试直接作为上下文 ID 读取
		contentLayer, err = vfs.store.GetLayer(ctx, uri, readLayer)
		if err != nil {
			return nil, fmt.Errorf("read content: %w", err)
		}
	}

	// 转换为字节数组
	var content []byte
	switch c := contentLayer.(type) {
	case *context.LayerSummary:
		content = []byte(c.Content)
	case *context.LayerOverview:
		content = []byte(c.Content)
	case *context.LayerDetails:
		content = []byte(c.Content)
	case string:
		content = []byte(c)
	case []byte:
		content = c
	default:
		return nil, fmt.Errorf("unsupported layer type: %T", contentLayer)
	}

	// 更新缓存
	vfs.cache[cacheKey] = &cachedContent{
		content:   content,
		timestamp: time.Now(),
		layer:     readLayer,
	}

	return content, nil
}

// WriteFile 写入文件内容
func (vfs *VirtualFileSystem) WriteFile(ctx stdcontext.Context, uri string, data []byte) error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	if !vfs.started {
		return fmt.Errorf("VFS not started")
	}

	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return fmt.Errorf("parse URI: %w", err)
	}

	// 确定目标层级
	layer := vuri.Layer()
	if layer == context.LayerAuto {
		layer = context.LayerTypeL2 // 默认写入 L2
	}

	// 创建或更新上下文
	contextID := vfs.pathToContextID(vuri.FullPath())

	// 检查上下文是否存在
	ctxt, err := vfs.store.GetContext(ctx, contextID)
	if err != nil {
		// 创建新上下文
		ctxt = context.NewContext(context.ContextTypeFile, vuri.path)
		ctxt.ID = contextID
		ctxt.Workspace = vuri.workspace
		ctxt.URI = uri
		ctxt.Layers.L2 = &context.LayerDetails{
			Content:     string(data),
			Tokens:      len(data) / 4, // 简单估算
			Format:      "raw",
			Source:      "vfs",
			GeneratedAt: time.Now(),
		}
		_, err = vfs.store.CreateContext(ctx, ctxt)
	} else {
		// 更新现有上下文
		if layer == context.LayerTypeL2 || layer == context.LayerAuto {
			update := context.ContextUpdate{
				Layers: &context.ContextLayersUpdate{
					L2: &context.LayerDetailsUpdate{
						Content: func(s string) *string { return &s }(string(data)),
						Tokens:  func(i int) *int { return &i }(len(data) / 4),
						Format:  func(s string) *string { return &s }("plain"),
					},
				},
			}
			err = vfs.store.UpdateContext(ctx, contextID, update)
		} else {
			err = vfs.store.SetLayer(ctx, contextID, layer, &context.LayerDetails{
				Content:     string(data),
				Tokens:      len(data) / 4,
				Source:      "vfs",
				GeneratedAt: time.Now(),
			})
		}
	}

	if err != nil {
		return fmt.Errorf("write context: %w", err)
	}

	// 使缓存失效
	vfs.invalidateCache(uri)

	return nil
}

// DeleteFile 删除文件
func (vfs *VirtualFileSystem) DeleteFile(ctx stdcontext.Context, uri string) error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	if !vfs.started {
		return fmt.Errorf("VFS not started")
	}

	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return fmt.Errorf("parse URI: %w", err)
	}

	contextID := vfs.pathToContextID(vuri.FullPath())

	// 删除上下文
	if err := vfs.store.DeleteContext(ctx, contextID); err != nil {
		return fmt.Errorf("delete context: %w", err)
	}

	// 使缓存失效
	vfs.invalidateCache(uri)

	return nil
}

// ListFiles 列出目录中的文件
func (vfs *VirtualFileSystem) ListFiles(ctx stdcontext.Context, uri string) ([]*context.VFSFileInfo, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	if !vfs.started {
		return nil, fmt.Errorf("VFS not started")
	}

	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return nil, fmt.Errorf("parse URI: %w", err)
	}

	// 获取工作区下的所有上下文
	// 这里简化实现，实际需要更复杂的查询

	// 模拟列出文件
	var files []*context.VFSFileInfo

	// 如果是根目录，列出工作区
	if vuri.IsRoot() {
		files = append(files, &context.VFSFileInfo{
			URI:      "viking://" + vuri.Workspace() + "/",
			Name:     vuri.Workspace(),
			Type:     "dir",
			ModTime:  time.Now(),
			Layers:   context.LayerAvailability{},
		})
	} else {
		// 列出子目录和文件
		// 这里需要从存储中查询
		// 简化实现：返回占位符
		files = append(files, &context.VFSFileInfo{
			URI:     uri + "/example.txt",
			Name:    "example.txt",
			Type:    "file",
			Size:    1024,
			ModTime: time.Now(),
			Layers: context.LayerAvailability{
				L0: true,
				L1: true,
				L2: true,
			},
		})
	}

	return files, nil
}

// ===== 目录操作 =====

// Mkdir 创建目录
func (vfs *VirtualFileSystem) Mkdir(ctx stdcontext.Context, uri string) error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	if !vfs.started {
		return fmt.Errorf("VFS not started")
	}

	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return fmt.Errorf("parse URI: %w", err)
	}

	// 创建目录上下文
	contextID := vfs.pathToContextID(vuri.FullPath())
	ctxt := context.NewContext(context.ContextTypeProject, filepath.Base(vuri.path))
	ctxt.ID = contextID
	ctxt.Workspace = vuri.workspace
	ctxt.URI = uri
	ctxt.Metadata = map[string]string{"type": "directory"}

	_, err = vfs.store.CreateContext(ctx, ctxt)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return nil
}

// MkdirAll 递归创建目录
func (vfs *VirtualFileSystem) MkdirAll(ctx stdcontext.Context, uri string) error {
	// 解析 URI
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return fmt.Errorf("parse URI: %w", err)
	}

	// 分割路径
	parts := strings.Split(vuri.Path(), "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		currentPath = path.Join(currentPath, part)
		currentURI := fmt.Sprintf("viking://%s/%s", vuri.Workspace(), currentPath)

		// 尝试创建目录（如果不存在）
		vfs.Mkdir(ctx, currentURI)
	}

	return nil
}

// ===== 查询操作 =====

// Glob 模式匹配文件
func (vfs *VirtualFileSystem) Glob(ctx stdcontext.Context, pattern string) ([]string, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	if !vfs.started {
		return nil, fmt.Errorf("VFS not started")
	}

	// 简化实现：返回空列表
	return []string{}, nil
}

// Search 搜索文件
func (vfs *VirtualFileSystem) Search(ctx stdcontext.Context, query string, opts ...context.SearchOption) ([]*context.VFSSearchResult, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	if !vfs.started {
		return nil, fmt.Errorf("VFS not started")
	}

	// 解析搜索选项
	options := &context.SearchOptions{
		Layer:      context.LayerAuto,
		MaxResults: 10,
		MinScore:   0.5,
	}

	for _, opt := range opts {
		opt(options)
	}

	// 使用存储的搜索功能
	return vfs.store.SearchFiles(ctx, query, opts...)
}

// ===== 移动和重命名 =====

// Move 移动文件或目录
func (vfs *VirtualFileSystem) Move(ctx stdcontext.Context, oldPath, newPath string) error {
	// 读取旧文件
	data, err := vfs.ReadFile(ctx, oldPath, context.LayerTypeL2)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}

	// 写入新文件
	if err := vfs.WriteFile(ctx, newPath, data); err != nil {
		return fmt.Errorf("write destination: %w", err)
	}

	// 删除旧文件
	if err := vfs.DeleteFile(ctx, oldPath); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	return nil
}

// Rename 重命名文件或目录
func (vfs *VirtualFileSystem) Rename(ctx stdcontext.Context, oldPath, newPath string) error {
	return vfs.Move(ctx, oldPath, newPath)
}

// ===== 缓存管理 =====

// invalidateCache 使缓存失效
func (vfs *VirtualFileSystem) invalidateCache(uri string) {
	// 删除所有匹配的缓存项
	for key := range vfs.cache {
		if strings.HasPrefix(key, uri) {
			delete(vfs.cache, key)
		}
	}
}

// cacheCleanupLoop 缓存清理循环
func (vfs *VirtualFileSystem) cacheCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		vfs.mu.Lock()
		now := time.Now()
		for key, cached := range vfs.cache {
			if now.Sub(cached.timestamp) > vfs.cacheTTL {
				delete(vfs.cache, key)
			}
		}
		vfs.mu.Unlock()
	}
}

// ===== 辅助方法 =====

// pathToContextID 将路径转换为上下文 ID
func (vfs *VirtualFileSystem) pathToContextID(path string) string {
	// 简单的路径到 ID 转换
	// 实际实现可能需要更复杂的逻辑
	return strings.ReplaceAll(path, "/", "_")
}

// ClearCache 清空缓存
func (vfs *VirtualFileSystem) ClearCache() {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()
	vfs.cache = make(map[string]*cachedContent)
}

// GetCacheSize 获取缓存大小
func (vfs *VirtualFileSystem) GetCacheSize() int {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()
	return len(vfs.cache)
}

// IsStarted 检查是否已启动
func (vfs *VirtualFileSystem) IsStarted() bool {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()
	return vfs.started
}
