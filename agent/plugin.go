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
	"fmt"
	"sync"
)

// Plugin 定义了插件接口，所有插件都需要实现此接口
// 插件是扩展AgentFramework功能的一种方式，可以动态加载和卸载
// 插件可以提供新的技能、工作流、模型适配器等功能
// 插件应该是线程安全的，可以同时被多个Agent调用
// 插件应该是可测试的，即可以独立测试，不需要依赖其他组件
// 插件应该是可扩展的，即可以通过配置调整行为
// 插件应该是可监控的，即可以记录执行时间、调用次数等
// 插件应该是可审计的，即可以记录调用者、参数、结果等
// 插件应该是安全的，即不允许越权操作，不允许执行恶意代码
// 插件应该是高效的，即执行时间短，资源消耗低
// 插件应该是可靠的，即不会崩溃，不会产生不可预期的结果
// 插件应该是文档化的，即提供清晰的使用说明

type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Version 返回插件版本
	Version() string

	// Description 返回插件描述
	Description() string

	// Initialize 初始化插件，在插件加载时调用
	Initialize(ctx context.Context, host *Host) error

	// Shutdown 关闭插件，在插件卸载时调用
	Shutdown(ctx context.Context) error

	// IsEnabled 检查插件是否启用
	IsEnabled() bool

	// Enable 启用插件
	Enable() error

	// Disable 禁用插件
	Disable() error
}

// ClaudeCodePlugin 定义 Claude Code 插件接口
type ClaudeCodePlugin interface {
	Plugin

	// 获取 MCP 服务器配置
	GetMCPConfig() *MCPConfig

	// 启动 MCP 服务器
	StartMCPServer(ctx context.Context) error

	// 停止 MCP 服务器
	StopMCPServer(ctx context.Context) error

	// 检查 MCP 服务器状态
	IsMCPServerRunning() bool

	// 获取 MCP 服务器地址
	GetMCPServerAddress() string
}

// MCPConfig MCP 服务器配置
type MCPConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Disabled bool             `json:"disabled"`
}

// MCPServerInfo MCP 服务器信息
type MCPServerInfo struct {
	PluginName string
	Address    string
	Port       int
	PID        int
	Status     string
	Metrics    *MCPServerMetrics
}

// MCPServerMetrics MCP 服务器指标
type MCPServerMetrics struct {
	Connections int
	Requests    int
	Errors      int
	Uptime      string
}

// PluginManager 插件管理器接口，用于管理插件的加载、卸载、启用、禁用等操作
// 插件管理器负责插件的生命周期管理
// 插件管理器可以从本地文件系统、远程服务器、数据库等加载插件
// 插件管理器可以支持插件的热加载，即无需重启系统即可更新插件
// 插件管理器可以支持插件的版本管理，即同一插件可以有多个版本
// 插件管理器可以支持插件的依赖管理，即自动加载依赖的插件
// 插件管理器可以支持插件的权限管理，即控制哪些插件可以被加载
// 插件管理器可以支持插件的监控，即记录插件的调用情况
// 插件管理器可以支持插件的测试，即提供插件测试框架
// 插件管理器可以支持插件的文档生成，即自动生成插件文档

type PluginManager interface {
	// LoadPlugin 加载插件
	LoadPlugin(ctx context.Context, pluginPath string) error

	// UnloadPlugin 卸载插件
	UnloadPlugin(ctx context.Context, pluginName string) error

	// GetPlugin 获取插件
	GetPlugin(pluginName string) (Plugin, bool)

	// GetAllPlugins 获取所有插件
	GetAllPlugins() []Plugin

	// EnablePlugin 启用插件
	EnablePlugin(ctx context.Context, pluginName string) error

	// DisablePlugin 禁用插件
	DisablePlugin(ctx context.Context, pluginName string) error

	// IsPluginEnabled 检查插件是否启用
	IsPluginEnabled(pluginName string) bool

	// LoadPluginsFromDirectory 从目录加载所有插件
	LoadPluginsFromDirectory(ctx context.Context, directory string) error

	// ReloadPlugin 重新加载插件
	ReloadPlugin(ctx context.Context, pluginName string) error

	// InitializeAllPlugins 初始化所有插件
	InitializeAllPlugins(ctx context.Context, host *Host) error

	// ShutdownAllPlugins 关闭所有插件
	ShutdownAllPlugins(ctx context.Context) error
}

// ClaudeCodePluginManager 管理 Claude Code 插件
type ClaudeCodePluginManager interface {
	PluginManager

	// 从 .claude-plugin 目录加载插件
	LoadClaudeCodePlugin(ctx context.Context, pluginPath string) (ClaudeCodePlugin, error)

	// 获取 MCP 服务器信息
	GetMCPServerInfo(pluginName string) (*MCPServerInfo, bool)

	// 列出所有运行中的 MCP 服务器
	ListRunningMCPServers() []*MCPServerInfo
}

// PluginInfo 插件信息，包含插件的基本信息

type PluginInfo struct {
	Name        string `json:"name"`        // 插件名称
	Version     string `json:"version"`     // 插件版本
	Description string `json:"description"` // 插件描述
	Enabled     bool   `json:"enabled"`     // 是否启用
	Path        string `json:"path"`        // 插件路径
}

// DefaultPluginManager 插件管理器的默认实现
// 负责管理插件的加载、卸载、启用、禁用等操作

type DefaultPluginManager struct {
	plugins    map[string]Plugin      // 插件映射，键为插件名称，值为插件实例
	pluginInfo map[string]*PluginInfo // 插件信息映射
	host       *Host                  // 主机实例
	mu         sync.RWMutex           // 读写锁，保证线程安全
}

// NewPluginManager 创建一个新的插件管理器实例

func NewPluginManager() PluginManager {
	return &DefaultPluginManager{
		plugins:    make(map[string]Plugin),
		pluginInfo: make(map[string]*PluginInfo),
	}
}

// LoadPlugin 加载插件
// 目前只支持内置插件，后续可以扩展支持动态加载外部插件

func (pm *DefaultPluginManager) LoadPlugin(ctx context.Context, pluginPath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 目前只支持内置插件，后续可以扩展支持动态加载外部插件
	// 这里可以实现从文件系统加载插件的逻辑

	return fmt.Errorf("dynamic plugin loading not supported yet")
}

// UnloadPlugin 卸载插件

func (pm *DefaultPluginManager) UnloadPlugin(ctx context.Context, pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	// 关闭插件
	if err := plugin.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown plugin %s: %w", pluginName, err)
	}

	// 从映射中删除插件
	delete(pm.plugins, pluginName)
	delete(pm.pluginInfo, pluginName)

	return nil
}

// GetPlugin 获取插件

func (pm *DefaultPluginManager) GetPlugin(pluginName string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginName]
	return plugin, exists
}

// GetAllPlugins 获取所有插件

func (pm *DefaultPluginManager) GetAllPlugins() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// EnablePlugin 启用插件

func (pm *DefaultPluginManager) EnablePlugin(ctx context.Context, pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	if err := plugin.Enable(); err != nil {
		return fmt.Errorf("failed to enable plugin %s: %w", pluginName, err)
	}

	// 更新插件信息
	if info, exists := pm.pluginInfo[pluginName]; exists {
		info.Enabled = true
	}

	return nil
}

// DisablePlugin 禁用插件

func (pm *DefaultPluginManager) DisablePlugin(ctx context.Context, pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	if err := plugin.Disable(); err != nil {
		return fmt.Errorf("failed to disable plugin %s: %w", pluginName, err)
	}

	// 更新插件信息
	if info, exists := pm.pluginInfo[pluginName]; exists {
		info.Enabled = false
	}

	return nil
}

// IsPluginEnabled 检查插件是否启用

func (pm *DefaultPluginManager) IsPluginEnabled(pluginName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return false
	}

	return plugin.IsEnabled()
}

// LoadPluginsFromDirectory 从目录加载所有插件

func (pm *DefaultPluginManager) LoadPluginsFromDirectory(ctx context.Context, directory string) error {
	// 目前只支持内置插件，后续可以扩展支持动态加载外部插件
	return fmt.Errorf("dynamic plugin loading from directory not supported yet")
}

// ReloadPlugin 重新加载插件

func (pm *DefaultPluginManager) ReloadPlugin(ctx context.Context, pluginName string) error {
	// 目前只支持内置插件，后续可以扩展支持动态加载外部插件
	return fmt.Errorf("plugin reloading not supported yet")
}

// InitializeAllPlugins 初始化所有插件

func (pm *DefaultPluginManager) InitializeAllPlugins(ctx context.Context, host *Host) error {
	pm.mu.Lock()
	pm.host = host
	pm.mu.Unlock()

	for _, plugin := range pm.plugins {
		if err := plugin.Initialize(ctx, host); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
		}
	}

	return nil
}

// ShutdownAllPlugins 关闭所有插件

func (pm *DefaultPluginManager) ShutdownAllPlugins(ctx context.Context) error {
	for _, plugin := range pm.plugins {
		if err := plugin.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown plugin %s: %w", plugin.Name(), err)
		}
	}

	return nil
}

// RegisterPlugin 注册插件
// 用于注册内置插件

func (pm *DefaultPluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := plugin.Name()

	// 检查插件是否已存在
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// 注册插件
	pm.plugins[name] = plugin

	// 保存插件信息
	pm.pluginInfo[name] = &PluginInfo{
		Name:        name,
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Enabled:     plugin.IsEnabled(),
		Path:        "built-in",
	}

	return nil
}

// DefaultPlugin 插件的默认实现，提供基本的插件功能
// 其他插件可以嵌入DefaultPlugin来简化实现

type DefaultPlugin struct {
	name        string
	version     string
	description string
	enabled     bool
	mu          sync.RWMutex
}

// NewDefaultPlugin 创建一个新的默认插件实例

func NewDefaultPlugin(name, version, description string) *DefaultPlugin {
	return &DefaultPlugin{
		name:        name,
		version:     version,
		description: description,
		enabled:     true,
	}
}

// Name 返回插件名称

func (dp *DefaultPlugin) Name() string {
	return dp.name
}

// Version 返回插件版本

func (dp *DefaultPlugin) Version() string {
	return dp.version
}

// Description 返回插件描述

func (dp *DefaultPlugin) Description() string {
	return dp.description
}

// Initialize 初始化插件

func (dp *DefaultPlugin) Initialize(ctx context.Context, host *Host) error {
	return nil
}

// Shutdown 关闭插件

func (dp *DefaultPlugin) Shutdown(ctx context.Context) error {
	return nil
}

// IsEnabled 检查插件是否启用

func (dp *DefaultPlugin) IsEnabled() bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	return dp.enabled
}

// Enable 启用插件

func (dp *DefaultPlugin) Enable() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.enabled = true
	return nil
}

// Disable 禁用插件

func (dp *DefaultPlugin) Disable() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.enabled = false
	return nil
}
