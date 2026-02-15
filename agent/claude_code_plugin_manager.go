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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// DefaultClaudeCodePlugin Claude Code 插件的默认实现
type DefaultClaudeCodePlugin struct {
	*DefaultPlugin
	mcpConfig     *MCPConfig
	mcpServerInfo *MCPServerInfo
	running       bool
	processManager ProcessManager
	managedProcess *ManagedProcess
}

// NewDefaultClaudeCodePlugin 创建一个新的 Claude Code 插件实例
func NewDefaultClaudeCodePlugin(name, version, description string, mcpConfig *MCPConfig) *DefaultClaudeCodePlugin {
	return &DefaultClaudeCodePlugin{
		DefaultPlugin: NewDefaultPlugin(name, version, description),
		mcpConfig:     mcpConfig,
		mcpServerInfo: &MCPServerInfo{
			PluginName: name,
			Status:     "stopped",
			Metrics: &MCPServerMetrics{
				Connections: 0,
				Requests:    0,
				Errors:      0,
			},
		},
		running:       false,
		processManager: NewProcessManager(),
	}
}

// GetMCPConfig 获取 MCP 服务器配置
func (p *DefaultClaudeCodePlugin) GetMCPConfig() *MCPConfig {
	return p.mcpConfig
}

// StartMCPServer 启动 MCP 服务器
func (p *DefaultClaudeCodePlugin) StartMCPServer(ctx context.Context) error {
	// 检查是否已经在运行
	if p.running {
		return nil
	}

	// 检查 MCP 配置是否被禁用
	if p.mcpConfig.Disabled {
		return fmt.Errorf("MCP server is disabled by configuration")
	}

	// 创建命令
	cmd := exec.Command(p.mcpConfig.Command, p.mcpConfig.Args...)

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range p.mcpConfig.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 启动进程
	process, err := p.processManager.StartProcess(ctx, cmd)
	if err != nil {
		p.mcpServerInfo.Status = "error"
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// 更新插件状态
	p.running = true
	p.managedProcess = process
	p.mcpServerInfo.Status = "running"
	p.mcpServerInfo.PID = process.PID
	p.mcpServerInfo.Metrics.Uptime = time.Now().Format(time.RFC3339)

	return nil
}

// StopMCPServer 停止 MCP 服务器
func (p *DefaultClaudeCodePlugin) StopMCPServer(ctx context.Context) error {
	// 检查是否已经停止
	if !p.running {
		return nil
	}

	// 停止进程
	if p.managedProcess != nil {
		if err := p.processManager.StopProcess(ctx, p.managedProcess.PID); err != nil {
			p.mcpServerInfo.Status = "error"
			return fmt.Errorf("failed to stop MCP server: %w", err)
		}
	}

	// 更新插件状态
	p.running = false
	p.managedProcess = nil
	p.mcpServerInfo.Status = "stopped"

	return nil
}

// IsMCPServerRunning 检查 MCP 服务器是否正在运行
func (p *DefaultClaudeCodePlugin) IsMCPServerRunning() bool {
	return p.running
}

// GetMCPServerAddress 获取 MCP 服务器地址
func (p *DefaultClaudeCodePlugin) GetMCPServerAddress() string {
	return p.mcpServerInfo.Address
}

// DefaultClaudeCodePluginManager Claude Code 插件管理器的默认实现
type DefaultClaudeCodePluginManager struct {
	*DefaultPluginManager
	mcpServerInfos map[string]*MCPServerInfo
	pluginPaths    map[string]string
	configCache    map[string]*PluginConfig  // 插件配置缓存
	mcpConfigCache map[string]*MCPConfig     // MCP 配置缓存
	loadingPlugins map[string]bool           // 正在加载的插件标记，防止重复加载
}

// NewClaudeCodePluginManager 创建一个新的 Claude Code 插件管理器实例
func NewClaudeCodePluginManager() ClaudeCodePluginManager {
	return &DefaultClaudeCodePluginManager{
		DefaultPluginManager: NewPluginManager().(*DefaultPluginManager),
		mcpServerInfos:       make(map[string]*MCPServerInfo),
		pluginPaths:          make(map[string]string),
		configCache:          make(map[string]*PluginConfig),
		mcpConfigCache:       make(map[string]*MCPConfig),
		loadingPlugins:       make(map[string]bool),
	}
}

// LoadClaudeCodePlugin 从 .claude-plugin 目录加载 Claude Code 插件（优化版本）
func (pm *DefaultClaudeCodePluginManager) LoadClaudeCodePlugin(ctx context.Context, pluginPath string) (ClaudeCodePlugin, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查是否正在加载此插件（防止并发加载同一插件）
	if pm.loadingPlugins[pluginPath] {
		// 等待其他 goroutine 完成加载
		for pm.loadingPlugins[pluginPath] {
			pm.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			pm.mu.Lock()
		}

		// 加载完成后，从插件列表中获取
		name := filepath.Base(pluginPath)
		if plugin, exists := pm.plugins[name]; exists {
			// 进行类型断言，确保是 ClaudeCodePlugin
			if ccp, ok := plugin.(ClaudeCodePlugin); ok {
				return ccp, nil
			}
			return nil, fmt.Errorf("plugin is not a ClaudeCodePlugin")
		}
		return nil, fmt.Errorf("plugin loading failed")
	}

	// 标记为正在加载
	pm.loadingPlugins[pluginPath] = true
	defer func() {
		delete(pm.loadingPlugins, pluginPath)
	}()

	// 检查插件路径是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin path %s does not exist", pluginPath)
	}

	// 尝试从缓存获取插件配置
	var pluginConfig *PluginConfig
	var ok bool
	if pluginConfig, ok = pm.configCache[pluginPath]; !ok {
		// 缓存未命中，解析插件配置文件
		config, err := pm.parsePluginConfig(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse plugin config: %w", err)
		}
		pluginConfig = config
		pm.configCache[pluginPath] = pluginConfig
	}

	// 尝试从缓存获取 MCP 配置
	var mcpConfig *MCPConfig
	if mcpConfig, ok = pm.mcpConfigCache[pluginPath]; !ok {
		// 缓存未命中，解析 MCP 服务器配置
		config, err := pm.parseMCPConfig(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MCP config: %w", err)
		}
		mcpConfig = config
		pm.mcpConfigCache[pluginPath] = mcpConfig
	}

	// 创建插件实例
	plugin := NewDefaultClaudeCodePlugin(
		pluginConfig.Name,
		pluginConfig.Version,
		pluginConfig.Description,
		mcpConfig,
	)

	// 保存插件路径
	pm.pluginPaths[pluginConfig.Name] = pluginPath

	// 保存 MCP 服务器信息
	pm.mcpServerInfos[pluginConfig.Name] = plugin.mcpServerInfo

	// 直接在 LoadClaudeCodePlugin 方法中实现注册插件的逻辑，避免锁冲突
	name := plugin.Name()

	// 检查插件是否已存在
	if _, exists := pm.plugins[name]; exists {
		return nil, fmt.Errorf("plugin %s already registered", name)
	}

	// 注册插件
	pm.plugins[name] = plugin

	// 保存插件信息
	pm.pluginInfo[name] = &PluginInfo{
		Name:        name,
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Enabled:     plugin.IsEnabled(),
		Path:        pluginPath,
	}

	return plugin, nil
}

// GetMCPServerInfo 获取 MCP 服务器信息
func (pm *DefaultClaudeCodePluginManager) GetMCPServerInfo(pluginName string) (*MCPServerInfo, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, exists := pm.mcpServerInfos[pluginName]
	return info, exists
}

// ListRunningMCPServers 列出所有运行中的 MCP 服务器
func (pm *DefaultClaudeCodePluginManager) ListRunningMCPServers() []*MCPServerInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var runningServers []*MCPServerInfo

	for _, info := range pm.mcpServerInfos {
		if info.Status == "running" {
			runningServers = append(runningServers, info)
		}
	}

	return runningServers
}

// parsePluginConfig 解析 plugin.json 配置文件
func (pm *DefaultClaudeCodePluginManager) parsePluginConfig(pluginPath string) (*PluginConfig, error) {
	configPath := filepath.Join(pluginPath, ".claude-plugin", "plugin.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin.json not found in %s", filepath.Join(pluginPath, ".claude-plugin"))
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	var config PluginConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	return &config, nil
}

// parseMCPConfig 解析 .mcp.json 配置文件
func (pm *DefaultClaudeCodePluginManager) parseMCPConfig(pluginPath string) (*MCPConfig, error) {
	configPath := filepath.Join(pluginPath, ".mcp.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".mcp.json not found in %s", pluginPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .mcp.json: %w", err)
	}

	var mcpServerConfig struct {
		MCPServers map[string]*MCPConfig `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &mcpServerConfig); err != nil {
		return nil, fmt.Errorf("failed to parse .mcp.json: %w", err)
	}

	// 假设只有一个 MCP 服务器配置
	for _, config := range mcpServerConfig.MCPServers {
		// 替换环境变量占位符
		config = pm.replaceEnvVariables(config)
		return config, nil
	}

	return nil, fmt.Errorf("no MCP servers configured in %s", configPath)
}

// replaceEnvVariables 替换配置中的环境变量占位符
func (pm *DefaultClaudeCodePluginManager) replaceEnvVariables(config *MCPConfig) *MCPConfig {
	// 替换命令和参数中的环境变量
	config.Command = pm.replaceEnvInString(config.Command)

	for i, arg := range config.Args {
		config.Args[i] = pm.replaceEnvInString(arg)
	}

	// 替换环境变量中的引用
	for key, value := range config.Env {
		config.Env[key] = pm.replaceEnvInString(value)
	}

	return config
}

// replaceEnvInString 替换字符串中的环境变量占位符
func (pm *DefaultClaudeCodePluginManager) replaceEnvInString(s string) string {
	// 简单实现：替换 ${VAR} 格式的占位符
	// 更复杂的情况可以使用 github.com/spf13/viper 等库
	return os.Expand(s, func(key string) string {
		// 支持 ${VAR:default} 格式
		defaultVal := ""
		parts := splitEnvKey(key)
		if len(parts) > 1 {
			key = parts[0]
			defaultVal = parts[1]
		}

		val := os.Getenv(key)
		if val == "" {
			return defaultVal
		}
		return val
	})
}

// splitEnvKey 分割环境变量键和默认值
func splitEnvKey(key string) []string {
	for i, c := range key {
		if c == ':' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}

// PluginConfig 插件配置结构
type PluginConfig struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      *Author  `json:"author"`
	Keywords    []string `json:"keywords"`
}

// Author 作者信息结构
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
