// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ClearCache 清理插件缓存
func (pm *DefaultClaudeCodePluginManager) ClearCache(pluginPath string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pluginPath == "" {
		// 清理所有缓存
		pm.configCache = make(map[string]*PluginConfig)
		pm.mcpConfigCache = make(map[string]*MCPConfig)
	} else {
		// 清理指定插件的缓存
		delete(pm.configCache, pluginPath)
		delete(pm.mcpConfigCache, pluginPath)
	}
}

// GetCacheSize 获取缓存大小信息
func (pm *DefaultClaudeCodePluginManager) GetCacheSize() (configCacheSize, mcpConfigCacheSize int) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.configCache), len(pm.mcpConfigCache)
}

// ReloadPlugin 重新加载插件（清除缓存后重新加载）
func (pm *DefaultClaudeCodePluginManager) ReloadPlugin(ctx context.Context, pluginName string) error {
	// 获取插件路径
	pm.mu.RLock()
	pluginPath, exists := pm.pluginPaths[pluginName]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 清除缓存
	pm.ClearCache(pluginPath)

	// 卸载插件
	if err := pm.UnloadPlugin(ctx, pluginName); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// 重新加载插件
	_, err := pm.LoadClaudeCodePlugin(ctx, pluginPath)
	return err
}

// LoadPluginsFromDirectory 批量加载插件（并行优化）
func (pm *DefaultClaudeCodePluginManager) LoadPluginsFromDirectory(ctx context.Context, directory string) error {
	// 读取目录中的所有子目录
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", directory, err)
	}

	// 收集所有插件路径
	var pluginPaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			pluginPath := filepath.Join(directory, entry.Name())
			// 检查是否是 Claude Code 插件
			pluginConfigPath := filepath.Join(pluginPath, ".claude-plugin", "plugin.json")
			if _, err := os.Stat(pluginConfigPath); err == nil {
				pluginPaths = append(pluginPaths, pluginPath)
			}
		}
	}

	// 使用 channel 和 goroutine 并行加载插件
	type pluginResult struct {
		plugin ClaudeCodePlugin
		err    error
		name   string
	}

	resultChan := make(chan pluginResult, len(pluginPaths))

	// 并行加载所有插件
	for _, pluginPath := range pluginPaths {
		go func(path string) {
			plugin, err := pm.LoadClaudeCodePlugin(ctx, path)
			resultChan <- pluginResult{
				plugin: plugin,
				err:    err,
				name:   filepath.Base(path),
			}
		}(pluginPath)
	}

	// 收集结果
	var errors []error
	for i := 0; i < len(pluginPaths); i++ {
		result := <-resultChan
		if result.err != nil {
			errors = append(errors, fmt.Errorf("failed to load plugin %s: %w", result.name, result.err))
		}
	}

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}