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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/fsnotify/fsnotify"
)

// ToolSource 工具来源信息
type ToolSource struct {
	// Source 来源地址（文件路径、URL 等）
	Source string
	// Loader 加载器名称
	Loader string
	// Checksum 校验和（用于检测更新）
	Checksum string
	// LastModified 最后修改时间
	LastModified time.Time
	// LoadedAt 加载时间
	LoadedAt time.Time
}

// DynamicToolRegistry 动态工具注册表，支持运行时添加和管理工具
type DynamicToolRegistry struct {
	tools       map[string]tool.BaseTool
	sources     map[string]*ToolSource // 工具名称 -> 来源信息
	loaders     []ToolLoader
	hooks       map[string][]ToolHook
	mu          sync.RWMutex
	enableCache bool
	cache       map[string]*schema.ToolInfo
	initialized bool
	// 热重载相关
	hotReloadEnabled bool
	hotReloadCancel  context.CancelFunc
	watcher          *fsnotify.Watcher
	watchedFiles     map[string]context.CancelFunc
}

// ToolLoader 工具加载器接口
type ToolLoader interface {
	// Name 返回加载器名称
	Name() string
	// CanLoad 检查是否可以加载指定源
	CanLoad(source string) bool
	// Load 从指定源加载工具
	Load(ctx context.Context, source string) (tool.BaseTool, error)
	// Validate 验证工具
	Validate(ctx context.Context, t tool.BaseTool) error
}

// ToolHook 工具钩子函数
type ToolHook func(ctx context.Context, action ToolAction, t tool.BaseTool) error

// ToolAction 工具动作
type ToolAction string

const (
	// ToolActionLoad 工具加载动作
	ToolActionLoad ToolAction = "load"
	// ToolActionUnload 工具卸载动作
	ToolActionUnload ToolAction = "unload"
	// ToolActionBeforeInvoke 工具调用前
	ToolActionBeforeInvoke ToolAction = "before_invoke"
	// ToolActionAfterInvoke 工具调用后
	ToolActionAfterInvoke ToolAction = "after_invoke"
)

// DynamicToolRegistryConfig 动态工具注册表配置
type DynamicToolRegistryConfig struct {
	// EnableCache 是否启用缓存
	EnableCache bool
	// InitialLoaders 初始加载器列表
	InitialLoaders []ToolLoader
	// EnableHotReload 是否启用热重载
	EnableHotReload bool
	// HotReloadInterval 热重载检查间隔
	HotReloadInterval time.Duration
}

// NewDynamicToolRegistry 创建动态工具注册表
func NewDynamicToolRegistry(config DynamicToolRegistryConfig) (*DynamicToolRegistry, error) {
	registry := &DynamicToolRegistry{
		tools:            make(map[string]tool.BaseTool),
		sources:          make(map[string]*ToolSource),
		loaders:          make([]ToolLoader, 0),
		hooks:            make(map[string][]ToolHook),
		enableCache:      config.EnableCache,
		cache:            make(map[string]*schema.ToolInfo),
		initialized:      true,
		hotReloadEnabled: config.EnableHotReload,
		watchedFiles:     make(map[string]context.CancelFunc),
	}

	// 初始化文件监控器
	if config.EnableHotReload {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("failed to create file watcher: %w", err)
		}
		registry.watcher = watcher
	}

	// 添加初始加载器
	for _, loader := range config.InitialLoaders {
		if err := registry.RegisterLoader(loader); err != nil {
			return nil, err
		}
	}

	// 启动热重载（如果启用）
	if config.EnableHotReload && config.HotReloadInterval > 0 {
		go registry.startHotReload(config.HotReloadInterval)
	}

	return registry, nil
}

// Register 注册工具
func (r *DynamicToolRegistry) Register(ctx context.Context, t tool.BaseTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 获取工具信息
	info, err := t.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tool info: %w", err)
	}

	// 检查工具是否已存在
	if _, exists := r.tools[info.Name]; exists {
		return fmt.Errorf("tool %s already registered", info.Name)
	}

	// 执行加载钩子
	if err := r.executeHooks(ctx, ToolActionLoad, t); err != nil {
		return fmt.Errorf("tool hook failed: %w", err)
	}

	// 注册工具
	r.tools[info.Name] = t

	// 更新缓存
	if r.enableCache {
		r.cache[info.Name] = info
	}

	return nil
}

// Unload 卸载工具
func (r *DynamicToolRegistry) Unload(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, exists := r.tools[name]
	if !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	// 执行卸载钩子
	if err := r.executeHooks(ctx, ToolActionUnload, t); err != nil {
		return fmt.Errorf("tool hook failed: %w", err)
	}

	// 移除工具
	delete(r.tools, name)

	// 清除缓存
	if r.enableCache {
		delete(r.cache, name)
	}

	return nil
}

// Get 获取工具
func (r *DynamicToolRegistry) Get(ctx context.Context, name string) (tool.BaseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有工具
func (r *DynamicToolRegistry) List(ctx context.Context) []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]tool.BaseTool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// ListInfo 列出所有工具信息
func (r *DynamicToolRegistry) ListInfo(ctx context.Context) ([]*schema.ToolInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*schema.ToolInfo, 0, len(r.tools))
	for _, t := range r.tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// LoadFromSource 从指定源加载工具
func (r *DynamicToolRegistry) LoadFromSource(ctx context.Context, source string) (tool.BaseTool, error) {
	// 查找合适的加载器
	loader, err := r.findLoader(source)
	if err != nil {
		return nil, err
	}

	// 计算来源校验和（用于热重载检测）
	checksum, modTime, err := r.calculateSourceChecksum(source)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate source checksum: %w", err)
	}

	// 加载工具
	t, err := loader.Load(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to load tool from %s: %w", source, err)
	}

	// 验证工具
	if err := loader.Validate(ctx, t); err != nil {
		return nil, fmt.Errorf("tool validation failed: %w", err)
	}

	// 获取工具信息
	info, err := t.Info(ctx)
	if err != nil {
		return nil, err
	}

	// 存储来源信息
	r.mu.Lock()
	r.sources[info.Name] = &ToolSource{
		Source:       source,
		Loader:       loader.Name(),
		Checksum:     checksum,
		LastModified: modTime,
		LoadedAt:     time.Now(),
	}
	r.mu.Unlock()

	// 注册工具
	if err := r.Register(ctx, t); err != nil {
		return nil, err
	}

	// 如果是文件来源，添加文件监控
	if r.watcher != nil && (strings.HasPrefix(source, "file:") || strings.HasPrefix(source, "http:") || strings.HasPrefix(source, "https:")) {
		filePath := strings.TrimPrefix(source, "file:")
		if err := r.watchFile(filePath, info.Name); err != nil {
			// 监控失败不影响加载
			_ = err
		}
	}

	return t, nil
}

// calculateSourceChecksum 计算来源校验和
func (r *DynamicToolRegistry) calculateSourceChecksum(source string) (string, time.Time, error) {
	var filePath string
	if strings.HasPrefix(source, "file:") {
		filePath = strings.TrimPrefix(source, "file:")
	} else if !strings.Contains(source, ":") || strings.HasPrefix(source, "./") || strings.HasPrefix(source, "/") {
		filePath = source
	}

	if filePath == "" {
		return "", time.Time{}, nil // 非文件来源，无法计算校验和
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", time.Time{}, err
	}

	// 计算 MD5 校验和
	hash := md5.Sum(data)
	checksum := hex.EncodeToString(hash[:])

	// 获取文件修改时间
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return checksum, time.Time{}, nil
	}

	return checksum, fileInfo.ModTime(), nil
}

// LoadFromURL 从URL加载工具
func (r *DynamicToolRegistry) LoadFromURL(ctx context.Context, url string) (tool.BaseTool, error) {
	return r.LoadFromSource(ctx, "url:"+url)
}

// LoadFromFile 从文件加载工具
func (r *DynamicToolRegistry) LoadFromFile(ctx context.Context, path string) (tool.BaseTool, error) {
	return r.LoadFromSource(ctx, "file:"+path)
}

// LoadFromMCP 从MCP服务器加载工具
func (r *DynamicToolRegistry) LoadFromMCP(ctx context.Context, serverURL string) (tool.BaseTool, error) {
	return r.LoadFromSource(ctx, "mcp:"+serverURL)
}

// RegisterLoader 注册工具加载器
func (r *DynamicToolRegistry) RegisterLoader(loader ToolLoader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loaders = append(r.loaders, loader)
	return nil
}

// UnregisterLoader 注销工具加载器
func (r *DynamicToolRegistry) UnregisterLoader(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, loader := range r.loaders {
		if loader.Name() == name {
			r.loaders = append(r.loaders[:i], r.loaders[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("loader %s not found", name)
}

// AddHook 添加工具钩子
func (r *DynamicToolRegistry) AddHook(name string, hook ToolHook) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hooks == nil {
		r.hooks = make(map[string][]ToolHook)
	}

	r.hooks[name] = append(r.hooks[name], hook)
}

// RemoveHook 移除工具钩子
func (r *DynamicToolRegistry) RemoveHook(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hooks, name)
}

// Reload 重新加载工具
func (r *DynamicToolRegistry) Reload(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 获取来源信息
	sourceInfo, exists := r.sources[name]
	if !exists {
		return fmt.Errorf("tool %s source information not found", name)
	}

	t, exists := r.tools[name]
	if !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	// 卸载工具
	_ = r.executeHooks(ctx, ToolActionUnload, t)
	delete(r.tools, name)
	if r.enableCache {
		delete(r.cache, name)
	}

	// 查找原始加载器
	var loader ToolLoader
	for _, l := range r.loaders {
		if l.Name() == sourceInfo.Loader {
			loader = l
			break
		}
	}

	if loader == nil {
		return fmt.Errorf("original loader %s not found", sourceInfo.Loader)
	}

	// 重新加载工具
	newTool, err := loader.Load(ctx, sourceInfo.Source)
	if err != nil {
		return fmt.Errorf("failed to reload tool: %w", err)
	}

	// 验证工具
	if err := loader.Validate(ctx, newTool); err != nil {
		return fmt.Errorf("tool validation failed: %w", err)
	}

	// 获取新工具信息
	info, err := newTool.Info(ctx)
	if err != nil {
		return err
	}

	// 更新来源信息
	newChecksum, modTime, err := r.calculateSourceChecksum(sourceInfo.Source)
	if err == nil {
		sourceInfo.Checksum = newChecksum
		sourceInfo.LastModified = modTime
		sourceInfo.LoadedAt = time.Now()
		r.sources[info.Name] = sourceInfo
	}

	// 执行加载钩子
	if err := r.executeHooks(ctx, ToolActionLoad, newTool); err != nil {
		return fmt.Errorf("tool hook failed: %w", err)
	}

	// 注册新工具
	r.tools[info.Name] = newTool

	return nil
}

// Count 返回工具数量
func (r *DynamicToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Clear 清空所有工具
func (r *DynamicToolRegistry) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 执行所有工具的卸载钩子
	for _, t := range r.tools {
		_ = r.executeHooks(ctx, ToolActionUnload, t)
	}

	r.tools = make(map[string]tool.BaseTool)
	if r.enableCache {
		r.cache = make(map[string]*schema.ToolInfo)
	}

	return nil
}

// findLoader 查找合适的加载器
func (r *DynamicToolRegistry) findLoader(source string) (ToolLoader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, loader := range r.loaders {
		if loader.CanLoad(source) {
			return loader, nil
		}
	}

	return nil, fmt.Errorf("no suitable loader found for source: %s", source)
}

// executeHooks 执行钩子
func (r *DynamicToolRegistry) executeHooks(ctx context.Context, action ToolAction, t tool.BaseTool) error {
	for _, hooks := range r.hooks {
		for _, hook := range hooks {
			if err := hook(ctx, action, t); err != nil {
				return err
			}
		}
	}
	return nil
}

// startHotReload 启动热重载
func (r *DynamicToolRegistry) startHotReload(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.checkForUpdates()
		}
	}
}

// checkForUpdates 检查工具是否有更新
func (r *DynamicToolRegistry) checkForUpdates() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := context.Background()

	for name, sourceInfo := range r.sources {
		// 计算当前校验和
		newChecksum, modTime, err := r.calculateSourceChecksum(sourceInfo.Source)
		if err != nil {
			continue // 跳过无法检查的来源
		}

		// 检查是否有更新（校验和变化或修改时间变化）
		if newChecksum != sourceInfo.Checksum || !modTime.Equal(sourceInfo.LastModified) {
			// 触发重新加载
			_ = r.Reload(ctx, name)
		}
	}
}

// GetCachedInfo 获取缓存的工具信息
func (r *DynamicToolRegistry) GetCachedInfo(name string) (*schema.ToolInfo, bool) {
	if !r.enableCache {
		return nil, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.cache[name]
	return info, ok
}

// UpdateCache 更新缓存
func (r *DynamicToolRegistry) UpdateCache(ctx context.Context) error {
	if !r.enableCache {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for name, t := range r.tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		r.cache[name] = info
	}

	return nil
}

// Search 搜索工具
func (r *DynamicToolRegistry) Search(ctx context.Context, query string) ([]tool.BaseTool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []tool.BaseTool

	for _, t := range r.tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}

		// 简单的名称和描述匹配
		if containsIgnoreCase(info.Name, query) || containsIgnoreCase(info.Desc, query) {
			results = append(results, t)
		}
	}

	return results, nil
}

// containsIgnoreCase 忽略大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			(s[0] == substr[0] || s[0]|0x20 == substr[0]|0x20) &&
			containsIgnoreCase(s[1:], substr[1:]))
}

// GetStats 获取注册表统计信息
func (r *DynamicToolRegistry) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_tools"] = len(r.tools)
	stats["total_loaders"] = len(r.loaders)
	stats["total_hooks"] = len(r.hooks)
	stats["cache_enabled"] = r.enableCache
	stats["cached_tools"] = len(r.cache)

	return stats
}

// watchFile 监控文件变化
func (r *DynamicToolRegistry) watchFile(filePath, toolName string) error {
	if r.watcher == nil {
		return fmt.Errorf("file watcher not initialized")
	}

	// 如果是目录，监控目录
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	var watchPath string
	if fileInfo.IsDir() {
		watchPath = filePath
	} else {
		watchPath = filepath.Dir(filePath)
	}

	// 添加监控
	if err := r.watcher.Add(watchPath); err != nil {
		return err
	}

	// 创建上下文用于取消监控
	ctx, cancel := context.WithCancel(context.Background())
	r.watchedFiles[filePath] = cancel

	// 启动监控协程
	go r.monitorFileChanges(ctx, filePath, toolName)

	return nil
}

// monitorFileChanges 监控文件变化
func (r *DynamicToolRegistry) monitorFileChanges(ctx context.Context, filePath, toolName string) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// 检查是否是目标文件
			if event.Name != filePath {
				continue
			}

			// 检查事件类型
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				// 等待文件写入完成
				time.Sleep(100 * time.Millisecond)

				// 重新加载工具
				_ = r.Reload(context.Background(), toolName)
			}

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			_ = err // 记录错误但不停止监控
		}
	}
}

// Close 关闭注册表
func (r *DynamicToolRegistry) Close(ctx context.Context) error {
	// 停止所有文件监控
	for _, cancel := range r.watchedFiles {
		cancel()
	}

	// 关闭文件监控器
	if r.watcher != nil {
		_ = r.watcher.Close()
	}

	// 清空所有工具
	return r.Clear(ctx)
}
