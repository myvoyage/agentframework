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
	"time"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/cloudwego/eino/schema"
)

// MCPCapability MCP 服务器能力
type MCPCapability struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{}  `json:"metadata"`
}

// NewMCPClient 创建一个新的 MCP 客户端（使用现有的实现）
func NewMCPClient(address string) *MCPClient {
	return &MCPClient{
		address:   address,
		transport: "stdio",
		timeout:   10 * time.Second,
		tools:     make(map[string]*MCPToolDefinition),
	}
}

// MCPSkillAdapter MCP 技能适配器
type MCPSkillAdapter struct {
	name        string
	version     string
	description string
	enabled     bool
	mcpClient   *MCPClient
	toolInfo    *schema.ToolInfo
	mu          sync.RWMutex
}

// NewMCPSkillAdapter 创建一个新的 MCP 技能适配器
func NewMCPSkillAdapter(client *MCPClient, toolInfo *schema.ToolInfo) *MCPSkillAdapter {
	return &MCPSkillAdapter{
		name:        toolInfo.Name,
		version:     "1.0.0",
		description: toolInfo.Desc,
		enabled:     true,
		mcpClient:   client,
		toolInfo:    toolInfo,
	}
}

// Info 返回技能的元信息
func (a *MCPSkillAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.enabled {
		return nil, fmt.Errorf("skill %s is disabled", a.name)
	}

	return a.toolInfo, nil
}

// Invoke 执行技能
func (a *MCPSkillAdapter) Invoke(ctx context.Context, input string) (string, error) {
	a.mu.RLock()
	enabled := a.enabled
	mcpClient := a.mcpClient
	a.mu.RUnlock()

	if !enabled {
		return "", fmt.Errorf("skill %s is disabled", a.name)
	}

	// 使用现有的 CallTool 方法
	return mcpClient.CallTool(ctx, a.name, []byte(input))
}

// IsEnabled 检查技能是否启用
func (a *MCPSkillAdapter) IsEnabled(ctx context.Context) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.enabled
}

// GetMetadata 获取技能的元数据
func (a *MCPSkillAdapter) GetMetadata(ctx context.Context) SkillMetadata {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return SkillMetadata{
		Name:        a.name,
		Version:     a.version,
		Author:      "Claude Code",
		Description: a.description,
		Category:    "MCP",
		Tags:        []string{"mcp", "remote"},
		Dependencies: []string{},
		License:      "AGPL-3.0-or-later",
	}
}

// SetEnabled 设置技能的启用状态
func (a *MCPSkillAdapter) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.enabled = enabled
}

// MCPSkillLibrary MCP 技能库
type MCPSkillLibrary struct {
	skills map[string]*MCPSkillAdapter
	client *MCPClient
	mu     sync.RWMutex
}

// NewMCPSkillLibrary 创建一个新的 MCP 技能库
func NewMCPSkillLibrary(client *MCPClient) *MCPSkillLibrary {
	return &MCPSkillLibrary{
		skills: make(map[string]*MCPSkillAdapter),
		client: client,
	}
}

// LoadSkills 从 MCP 服务器加载技能
func (lib *MCPSkillLibrary) LoadSkills(ctx context.Context) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	// 初始化 MCP 客户端并获取工具
	if err := lib.client.Initialize(ctx); err != nil {
		return err
	}

	// 获取工具列表
	tools := lib.client.ListTools()
	for name, toolDef := range tools {
		// 转换为 schema.ToolInfo
		toolInfo := &schema.ToolInfo{
			Name:        name,
			Desc: toolDef.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(make(map[string]*schema.ParameterInfo)),
		}

		// 简单处理输入模式（在实际应用中可能需要更复杂的转换）
		if toolDef.InputSchema != nil {
			params := make(map[string]*schema.ParameterInfo)
			for paramName, paramSchema := range toolDef.InputSchema {
				params[paramName] = &schema.ParameterInfo{
					Type: "string",
					Desc: fmt.Sprintf("%v", paramSchema),
				}
			}
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(params)
		}

		lib.skills[name] = NewMCPSkillAdapter(lib.client, toolInfo)
	}

	return nil
}

// GetSkill 获取技能
func (lib *MCPSkillLibrary) GetSkill(ctx context.Context, name string) (*MCPSkillAdapter, bool) {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	skill, exists := lib.skills[name]
	return skill, exists
}

// GetAllSkills 获取所有技能
func (lib *MCPSkillLibrary) GetAllSkills(ctx context.Context) map[string]*MCPSkillAdapter {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	copySkills := make(map[string]*MCPSkillAdapter)
	for name, skill := range lib.skills {
		copySkills[name] = skill
	}

	return copySkills
}

// EnableSkill 启用技能
func (lib *MCPSkillLibrary) EnableSkill(ctx context.Context, name string) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	skill, exists := lib.skills[name]
	if !exists {
		return fmt.Errorf("skill %s not found", name)
	}

	skill.SetEnabled(true)
	return nil
}

// DisableSkill 禁用技能
func (lib *MCPSkillLibrary) DisableSkill(ctx context.Context, name string) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	skill, exists := lib.skills[name]
	if !exists {
		return fmt.Errorf("skill %s not found", name)
	}

	skill.SetEnabled(false)
	return nil
}

// RefreshSkills 刷新技能列表
func (lib *MCPSkillLibrary) RefreshSkills(ctx context.Context) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	// 清除现有技能
	lib.skills = make(map[string]*MCPSkillAdapter)

	// 重新加载技能
	if err := lib.client.Initialize(ctx); err != nil {
		return err
	}

	// 获取工具列表
	tools := lib.client.ListTools()
	for name, toolDef := range tools {
		// 转换为 schema.ToolInfo
		toolInfo := &schema.ToolInfo{
			Name:        name,
			Desc: toolDef.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(make(map[string]*schema.ParameterInfo)),
		}

		lib.skills[name] = NewMCPSkillAdapter(lib.client, toolInfo)
	}

	return nil
}

// NewMCPTools creates Eino-compatible tools from an MCP client.
// This wraps the eino-ext/components/tool/mcp implementation.
func NewMCPTools(ctx context.Context, cli client.MCPClient) ([]tool.BaseTool, error) {
	// Configure the MCP tool source
	cfg := &mcp.Config{
		Cli: cli,
	}

	// Fetch and adapt tools using Eino's extension
	return mcp.GetTools(ctx, cfg)
}
