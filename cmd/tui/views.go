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
// Agent Framework - TUI Views Enhancement
// Copyright (C) 2025 Agent Framework Contributors
//
// 视图渲染优化 - 借鉴 Memoh 的表格渲染和状态显�?
package tui

import (
	"fmt"
	"strings"
	"time"
)

// ========== 视图渲染增强 ==========

// renderDashboard 渲染仪表板（增强版）
func (m *Model) renderDashboard(width, height int) string {
	var content strings.Builder

	// 标题
	content.WriteString("╔════════════════════════════════════════╗\n")
	content.WriteString("�?    AgentFramework TUI - 仪表�?         ║\n")
	content.WriteString("╚════════════════════════════════════════╝\n\n")

	// 统计信息
	content.WriteString(m.styles.Header.Render("📊 系统统计"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Agents:   %d 个\n", len(m.agents)))
	content.WriteString(fmt.Sprintf("工作�?   %d 个\n", len(m.workflows)))
	content.WriteString(fmt.Sprintf("技�?      %d 个\n", len(m.skills)))
	content.WriteString(fmt.Sprintf("消息:      %d 条\n", len(m.chatMessages)))

	content.WriteString("\n")

	// 当前状�?	content.WriteString(m.styles.Header.Render("�?当前状�?))
	content.WriteString("\n\n")

	if m.currentAgent != "" {
		content.WriteString(fmt.Sprintf("�?当前 Agent: %s\n", m.currentAgent))
	} else {
		content.WriteString("�?未选择 Agent (使用 'agent select <id>' 选择)\n")
	}

	if m.config != nil {
		cfg := m.config.Get()
		content.WriteString(fmt.Sprintf("�?会话 ID: %s\n", cfg.SessionID))
		content.WriteString(fmt.Sprintf("�?流式聊天: %v\n", cfg.StreamChat))
	}

	content.WriteString("\n")

	// 快速操�?	content.WriteString(m.styles.Header.Render("🚀 快速操�?))
	content.WriteString("\n\n")
	content.WriteString("agent list              - 列出所�?Agents\n")
	content.WriteString("agent select <id>        - 选择 Agent\n")
	content.WriteString("chat <message>          - 发送消息\n")
	content.WriteString("workflow list           - 列出工作流\n")
	content.WriteString("skill list              - 列出技能\n")

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderAgents 渲染 Agents 视图（增强版�?func (m *Model) renderAgents(width, height int) string {
	var content strings.Builder

	if m.loadingAgents {
		content.WriteString("�?正在加载 Agents...")
		return m.styles.Card.Render(content.String())
	}

	if len(m.agents) == 0 {
		content.WriteString("�?无可�?Agents\n\n")
		content.WriteString("提示: 确保核心应用已正确初始化")
		return m.styles.Card.Render(content.String())
	}

	// 标题
	content.WriteString(fmt.Sprintf("已加�?%d �?Agent\n\n", len(m.agents)))

	// Agent 列表
	for i, agent := range m.agents {
		prefix := "  "
		if agent.ID == m.currentAgent {
			prefix = "�?"
		}

		content.WriteString(fmt.Sprintf("%s%s\n", prefix, m.styles.Success.Render(agent.Name)))
		content.WriteString(fmt.Sprintf("   ID:    %s\n", agent.ID))
		content.WriteString(fmt.Sprintf("   类型:  %s\n", agent.Type))

		if i < len(m.agents)-1 {
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("使用 'agent select <id>' 选择 Agent"))

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderChat 渲染聊天视图（增强版�?func (m *Model) renderChat(width, height int) string {
	var content strings.Builder

	// 标题和状�?	content.WriteString("💭 对话\n\n")

	if m.currentAgent == "" {
		content.WriteString("�?未选择 Agent\n\n")
		content.WriteString("请先使用 'agent select <id>' 选择一�?Agent\n\n")
		content.WriteString("可用命令:\n")
		content.WriteString("  agent list              - 查看 Agents\n")
		content.WriteString("  agent select <id>        - 选择 Agent")
		return m.styles.Card.Render(content.String())
	}

	content.WriteString(fmt.Sprintf("当前 Agent: %s\n\n", m.styles.Success.Render(m.currentAgent)))

	// 消息历史
	if len(m.chatMessages) == 0 {
		content.WriteString("�?暂无消息\n\n")
		content.WriteString("开始对�?\n")
		content.WriteString("  chat 你好")
	} else {
		content.WriteString(fmt.Sprintf("消息�? %d\n\n", len(m.chatMessages)))

		// 显示最近的 10 条消�?		displayCount := min(10, len(m.chatMessages))
		startIdx := len(m.chatMessages) - displayCount

		for i := startIdx; i < len(m.chatMessages); i++ {
			msg := m.chatMessages[i]

			// 角色标签
			roleLabel := "用户"
			roleStyle := m.styles.Body
			if msg.Role == "assistant" {
				roleLabel = "助手"
				roleStyle = m.styles.Info
			}

			// 时间�?			timeStr := msg.Timestamp.Format("15:04:05")

			// 流式状�?			streamingIndicator := ""
			if msg.Streaming {
				streamingIndicator = " �?
			}

			content.WriteString(fmt.Sprintf("[%s] %s: %s%s\n",
				m.styles.Muted.Render(timeStr),
				roleStyle.Render(roleLabel),
				msg.Content,
				streamingIndicator))

			if i < len(m.chatMessages)-1 {
				content.WriteString("\n")
			}
		}
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("提示: 输入 'chat <message>' 发送消�?))

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderWorkflows 渲染工作流视图（增强版）
func (m *Model) renderWorkflows(width, height int) string {
	var content strings.Builder

	if m.loadingWorkflows {
		content.WriteString("�?正在加载工作�?..")
		return m.styles.Card.Render(content.String())
	}

	if len(m.workflows) == 0 {
		content.WriteString("�?无可用工作流\n\n")
		content.WriteString("提示: 使用 'workflow create' 创建工作�?)
		return m.styles.Card.Render(content.String())
	}

	// 标题
	content.WriteString(fmt.Sprintf("已加�?%d 个工作流\n\n", len(m.workflows)))

	// 工作流列�?	for i, wf := range m.workflows {
		// 状态图�?		statusIcon := "�?
		statusStyle := m.styles.Muted
		if wf.Status == "Ready" || wf.Status == "Active" {
			statusIcon = "�?
			statusStyle = m.styles.Success
		}

		content.WriteString(fmt.Sprintf("%s %s\n", statusIcon, statusStyle.Render(wf.Name)))
		content.WriteString(fmt.Sprintf("   ID:    %s\n", wf.ID))
		content.WriteString(fmt.Sprintf("   状�?  %s\n", wf.Status))
		content.WriteString(fmt.Sprintf("   描述:  %s", wf.Desc))

		if i < len(m.workflows)-1 {
			content.WriteString("\n\n")
		}
	}

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderSkills 渲染技能视图（增强版）
func (m *Model) renderSkills(width, height int) string {
	var content strings.Builder

	if m.loadingSkills {
		content.WriteString("�?正在加载技�?..")
		return m.styles.Card.Render(content.String())
	}

	if len(m.skills) == 0 {
		content.WriteString("�?无可用技能\n\n")
		content.WriteString("提示: 技能将�?.skills 目录自动加载")
		return m.styles.Card.Render(content.String())
	}

	// 标题和统�?	enabledCount := 0
	for _, skill := range m.skills {
		if skill.Enabled {
			enabledCount++
		}
	}

	content.WriteString(fmt.Sprintf("已加�?%d 个技�?(%d 个已启用)\n\n", len(m.skills), enabledCount))

	// 技能列�?	for i, skill := range m.skills {
		// 状态图�?		statusIcon := "�?
		statusStyle := m.styles.Muted
		if skill.Enabled {
			statusIcon = "�?
			statusStyle = m.styles.Success
		}

		content.WriteString(fmt.Sprintf("%s %s (v%s)\n", statusIcon, statusStyle.Render(skill.Name), skill.Version))
		content.WriteString(fmt.Sprintf("   ID:      %s\n", skill.ID))
		content.WriteString(fmt.Sprintf("   描述:    %s", skill.Desc))

		if i < len(m.skills)-1 {
			content.WriteString("\n\n")
		}
	}

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderSettings 渲染设置视图（增强版�?func (m *Model) renderSettings(width, height int) string {
	var content strings.Builder

	content.WriteString("⚙️ 设置与帮助\n\n")

	// 快捷�?	content.WriteString(m.styles.Header.Render("快捷�?))
	content.WriteString("\n\n")
	content.WriteString("  Tab/Shift+Tab    切换视图\n")
	content.WriteString("  Ctrl+R           刷新数据\n")
	content.WriteString("  Enter            执行命令\n")
	content.WriteString("  Q/Ctrl+C         退�?TUI\n")

	content.WriteString("\n")

	// 命令
	content.WriteString(m.styles.Header.Render("命令"))
	content.WriteString("\n\n")
	content.WriteString("  Agent 操作:\n")
	content.WriteString("    agent list              列出所�?agents\n")
	content.WriteString("    agent select <id>        选择 agent\n")
	content.WriteString("\n")
	content.WriteString("  聊天:\n")
	content.WriteString("    chat <message>          发送消息\n")
	content.WriteString("\n")
	content.WriteString("  工作�?\n")
	content.WriteString("    workflow list           列出工作流\n")
	content.WriteString("    workflow execute <id>    执行工作流\n")
	content.WriteString("\n")
	content.WriteString("  技�?\n")
	content.WriteString("    skill list              列出技能\n")
	content.WriteString("    skill enable <id>        启用技能\n")
	content.WriteString("    skill disable <id>       禁用技能\n")

	content.WriteString("\n")

	// 配置信息
	if m.config != nil {
		cfg := m.config.Get()
		content.WriteString(m.styles.Header.Render("当前配置"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("  会话 ID:    %s\n", cfg.SessionID))
		content.WriteString(fmt.Sprintf("  流式聊天:    %v\n", cfg.StreamChat))
		content.WriteString(fmt.Sprintf("  自动滚动:    %v\n", cfg.AutoScroll))
		content.WriteString(fmt.Sprintf("  最大历�?    %d 条\n", cfg.MaxHistory))
	}

	content.WriteString("\n")

	// 系统信息
	content.WriteString(m.styles.Header.Render("系统"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("  版本:        %s\n", "2.1.0"))
	content.WriteString(fmt.Sprintf("  时间:        %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("  架构:        Memoh-inspired"))

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// renderLogs 渲染日志视图（增强版�?func (m *Model) renderLogs(width, height int) string {
	var content strings.Builder

	content.WriteString("📋 系统日志\n\n")

	if m.statusMessage != "" {
		// 显示最近的系统消息
		timeStr := m.statusTime.Format("15:04:05")

		var icon string
		switch m.statusType {
		case StatusSuccess:
			icon = "�?
		case StatusError:
			icon = "�?
		case StatusWarning:
			icon = "�?
		case StatusInfo:
			icon = "�?
		default:
			icon = "�?
		}

		content.WriteString(fmt.Sprintf("[%s] %s %s\n\n",
			m.styles.Muted.Render(timeStr),
			icon,
			m.statusMessage))
	}

	content.WriteString("�?暂无日志记录\n\n")
	content.WriteString("日志功能将在后续版本中实现，包括:\n")
	content.WriteString("  - Agent 执行日志\n")
	content.WriteString("  - 工作流运行日志\n")
	content.WriteString("  - 技能调用日志\n")
	content.WriteString("  - 系统事件日志")

	return m.styles.Card.Width(min(width-4, 80)).Render(content.String())
}

// ========== 辅助函数 ==========

// min 返回两个整数中的较小�?func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
