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
// Agent Framework - TUI Integration Layer
// Copyright (C) 2025 Agent Framework Contributors
//
// TUI �?core.Application 的集成层
// 实现 Memoh 风格的数据加载和 Agent 调用

package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"AgentFramework/core"
)

// ========== 集成�?==========

// IntegrationLayer TUI 与核心应用的集成�?type IntegrationLayer struct {
	core *core.Application
	ctx  context.Context
}

// NewIntegrationLayer 创建集成�?func NewIntegrationLayer(ctx context.Context, coreApp *core.Application) *IntegrationLayer {
	return &IntegrationLayer{
		core: coreApp,
		ctx:  ctx,
	}
}

// ========== 数据加载命令实现 ==========

// BatchLoadAgentsCmd 批量加载 Agent
func (il *IntegrationLayer) BatchLoadAgentsCmd() tea.Cmd {
	return func() tea.Msg {
		host := il.core.GetHost()
		agentIDs := host.ListAgents()

		items := make([]AgentItem, 0, len(agentIDs))
		for _, agentID := range agentIDs {
			// 获取 Agent 实例以获取名�?			agent, err := host.GetAgent(agentID)
			if err != nil {
				// 如果获取失败，使�?ID 作为名称
				items = append(items, AgentItem{
					ID:   agentID,
					Name: agentID,
					Type: "Unknown",
				})
				continue
			}

			items = append(items, AgentItem{
				ID:   agentID,
				Name: agent.Name(),
				Type: il.inferAgentType(agent),
			})
		}

		return AgentListLoadedMsg{
			Agents: items,
			Error:  nil,
		}
	}
}

// BatchLoadWorkflowsCmd 批量加载工作�?func (il *IntegrationLayer) BatchLoadWorkflowsCmd() tea.Cmd {
	return func() tea.Msg {
		workflowIDs := il.core.GetHost().ListWorkflows()

		items := make([]WorkflowItem, 0, len(workflowIDs))
		for _, wfID := range workflowIDs {
			// 获取工作流信�?			// 注意：这里需要根据实际的 WorkflowManager API 调整
			items = append(items, WorkflowItem{
				ID:     wfID,
				Name:   wfID,
				Desc:   "工作�?,
				Status: "Ready",
			})
		}

		return WorkflowListLoadedMsg{
			Workflows: items,
			Error:     nil,
		}
	}
}

// BatchLoadSkillsCmd 批量加载技�?func (il *IntegrationLayer) BatchLoadSkillsCmd() tea.Cmd {
	return func() tea.Msg {
		skillLibrary := il.core.GetSkillLibrary()
		skills := skillLibrary.GetAllSkills(il.ctx)

		items := make([]SkillItem, 0, len(skills))
		for skillID, skill := range skills {
			metadata := skill.GetMetadata(il.ctx)

			items = append(items, SkillItem{
				ID:      skillID,
				Name:    metadata.Name,
				Desc:    metadata.Description,
				Version: metadata.Version,
				Enabled: skill.IsEnabled(il.ctx),
			})
		}

		return SkillListLoadedMsg{
			Skills: items,
			Error:  nil,
		}
	}
}

// ========== Agent 调用实现 ==========

// StreamChatCmd 创建流式聊天命令
func (il *IntegrationLayer) StreamChatCmd(agentID, message, sessionID string) tea.Cmd {
	return func() tea.Msg {
		response, err := il.streamChat(agentID, message)
		if err != nil {
			return ChatResponseMsg{
				Error: fmt.Errorf("chat failed: %w", err),
				Done:  true,
			}
		}

		return ChatResponseMsg{
			ContentChunk: response,
			Done:         true,
		}
	}
}

// streamChat 执行聊天（实际实现）
func (il *IntegrationLayer) streamChat(agentID, message string) (string, error) {
	host := il.core.GetHost()

	// 获取 Agent
	agent, err := host.GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to get agent %s: %w", agentID, err)
	}

	// 运行 Agent
	response, err := agent.Run(il.ctx, message)
	if err != nil {
		return "", fmt.Errorf("agent run failed: %w", err)
	}

	return response.Content, nil
}

// ========== 辅助方法 ==========

// inferAgentType 推断 Agent 类型
func (il *IntegrationLayer) inferAgentType(agent interface{}) string {
	// 使用类型断言检�?Agent 类型
	switch agent.(type) {
	default:
		return "ChatAgent"
	}
}

// ========== 工作流操�?==========

// ExecuteWorkflow 执行工作�?func (il *IntegrationLayer) ExecuteWorkflow(workflowID, input string) tea.Cmd {
	return func() tea.Msg {
		wfManager := il.core.GetWorkflowManager()

		// 执行工作�?		result, err := wfManager.ExecuteWorkflow(il.ctx, workflowID, input)
		if err != nil {
			return StatusUpdateMsg{
				Message: fmt.Sprintf("工作流执行失�? %v", err),
				Type:    StatusError,
			}
		}

		return StatusUpdateMsg{
			Message: fmt.Sprintf("工作流执行成�? %s", result),
			Type:    StatusSuccess,
		}
	}
}

// ========== 技能操�?==========

// ToggleSkill 切换技能状�?func (il *IntegrationLayer) ToggleSkill(skillID string, enable bool) tea.Cmd {
	return func() tea.Msg {
		skillLibrary := il.core.GetSkillLibrary()

		var err error
		if enable {
			err = skillLibrary.EnableSkill(il.ctx, skillID)
		} else {
			err = skillLibrary.DisableSkill(il.ctx, skillID)
		}

		if err != nil {
			return StatusUpdateMsg{
				Message: fmt.Sprintf("技能操作失�? %v", err),
				Type:    StatusError,
			}
		}

		action := "禁用"
		if enable {
			action = "启用"
		}

		return StatusUpdateMsg{
			Message: fmt.Sprintf("技能已%s: %s", action, skillID),
			Type:    StatusSuccess,
		}
	}
}
