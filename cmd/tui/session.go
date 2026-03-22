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
// Agent Framework - TUI Session Persistence
// Copyright (C) 2025 Agent Framework Contributors
//
// 会话持久�?- 借鉴 Memoh 的自动保存模�?
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ========== 会话持久化增�?==========

// SaveChatSession 保存聊天会话
func (m *Model) SaveChatSession() error {
	if m.session == nil {
		return fmt.Errorf("session manager not initialized")
	}

	// 获取当前配置
	cfg := m.config.Get()
	sessionID := cfg.SessionID

	// 检查会话是否存�?	_, exists := m.session.Get(sessionID)
	if !exists {
		// 创建新会�?		_, err := m.session.Create(m.currentAgent)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	// 更新会话数据
	err := m.session.Update(sessionID, func(s *Session) {
		s.AgentID = m.currentAgent
		s.Messages = m.chatMessages
		s.UpdatedAt = time.Now()

		// 保存元数�?		if s.Metadata == nil {
			s.Metadata = make(map[string]string)
		}
		s.Metadata["lastView"] = m.currentView.String()
		s.Metadata["agentCount"] = fmt.Sprintf("%d", len(m.agents))
		s.Metadata["workflowCount"] = fmt.Sprintf("%d", len(m.workflows))
		s.Metadata["skillCount"] = fmt.Sprintf("%d", len(m.skills))
	})

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// LoadChatSession 加载聊天会话
func (m *Model) LoadChatSession(sessionID string) error {
	if m.session == nil {
		return fmt.Errorf("session manager not initialized")
	}

	session, exists := m.session.Get(sessionID)
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// 恢复会话数据
	m.currentAgent = session.AgentID
	m.chatMessages = session.Messages

	// 恢复配置（如果需要）
	if m.config != nil {
		m.config.Update(func(cfg *TUIConfig) {
			cfg.SessionID = sessionID
			cfg.LastAgentID = session.AgentID
		})
	}

	return nil
}

// AutoSaveSession 自动保存会话（用于定时保存）
func (m *Model) AutoSaveSession() tea.Cmd {
	return func() tea.Msg {
		if m.config != nil && m.config.Get().AutoSaveSession {
			if err := m.SaveChatSession(); err != nil {
				return StatusUpdateMsg{
					Message: fmt.Sprintf("自动保存失败: %v", err),
					Type:    StatusWarning,
				}
			}
		}
		return nil
	}
}

// ListSessions 列出所有会�?func (m *Model) ListSessions() []*Session {
	if m.session == nil {
		return []*Session{}
	}

	return m.session.List()
}

// DeleteSession 删除指定会话
func (m *Model) DeleteSession(sessionID string) error {
	if m.session == nil {
		return fmt.Errorf("session manager not initialized")
	}

	return m.session.Delete(sessionID)
}

// ExportSession 导出会话为文本格�?func (m *Model) ExportSession(sessionID string) (string, error) {
	if m.session == nil {
		return "", fmt.Errorf("session manager not initialized")
	}

	session, exists := m.session.Get(sessionID)
	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	var content string
	content += fmt.Sprintf("AgentFramework TUI - 会话导出\n")
	content += fmt.Sprintf("会话 ID: %s\n", session.ID)
	content += fmt.Sprintf("Agent ID: %s\n", session.AgentID)
	content += fmt.Sprintf("创建时间: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("更新时间: %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("消息�? %d\n\n", len(session.Messages))
	content += "─────────────────────────────────────────\n\n"

	for _, msg := range session.Messages {
		role := "用户"
		if msg.Role == "assistant" {
			role = "助手"
		}

		content += fmt.Sprintf("[%s] [%s]: %s\n",
			msg.Timestamp.Format("15:04:05"),
			role,
			msg.Content)
	}

	return content, nil
}

// ========== 会话管理命令 ==========

// NewSessionCmd 创建新会话命�?func (m *Model) NewSessionCmd(agentID string) tea.Cmd {
	return func() tea.Msg {
		if m.session == nil {
			return StatusUpdateMsg{
				Message: "会话管理器未初始�?,
				Type:    StatusError,
			}
		}

		session, err := m.session.Create(agentID)
		if err != nil {
			return StatusUpdateMsg{
				Message: fmt.Sprintf("创建会话失败: %v", err),
				Type:    StatusError,
			}
		}

		// 更新配置
		if m.config != nil {
			m.config.Update(func(cfg *TUIConfig) {
				cfg.SessionID = session.ID
				cfg.LastAgentID = agentID
			})
		}

		// 清空当前消息
		m.chatMessages = make([]ChatMessageItem, 0)
		m.currentAgent = agentID

		return StatusUpdateMsg{
			Message: fmt.Sprintf("新会话已创建: %s", session.ID),
			Type:    StatusSuccess,
		}
	}
}

// LoadSessionCmd 加载会话命令
func (m *Model) LoadSessionCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		if err := m.LoadChatSession(sessionID); err != nil {
			return StatusUpdateMsg{
				Message: fmt.Sprintf("加载会话失败: %v", err),
				Type:    StatusError,
			}
		}

		return StatusUpdateMsg{
			Message: fmt.Sprintf("会话已加�? %s", sessionID),
			Type:    StatusSuccess,
		}
	}
}

// DeleteSessionCmd 删除会话命令
func (m *Model) DeleteSessionCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		if err := m.DeleteSession(sessionID); err != nil {
			return StatusUpdateMsg{
				Message: fmt.Sprintf("删除会话失败: %v", err),
				Type:    StatusError,
			}
		}

		return StatusUpdateMsg{
			Message: fmt.Sprintf("会话已删�? %s", sessionID),
			Type:    StatusSuccess,
		}
	}
}

// ========== 会话相关命令处理 ==========

// handleSessionCommand 处理会话相关命令
func (m *Model) handleSessionCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		// 显示会话列表
		sessions := m.ListSessions()

		if len(sessions) == 0 {
			m.setStatus("无可用会�?, StatusInfo)
			return m, nil
		}

		m.setStatus(fmt.Sprintf("共有 %d 个会�?, len(sessions)), StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "new":
		agentID := m.currentAgent
		if len(parts) >= 3 {
			agentID = parts[2]
		}
		return m, m.NewSessionCmd(agentID)

	case "load":
		if len(parts) >= 3 {
			return m, m.LoadSessionCmd(parts[2])
		}
		m.setStatus("用法: session load <id>", StatusWarning)

	case "delete":
		if len(parts) >= 3 {
			return m, m.DeleteSessionCmd(parts[2])
		}
		m.setStatus("用法: session delete <id>", StatusWarning)

	case "list":
		sessions := m.ListSessions()
		m.setStatus(fmt.Sprintf("共有 %d 个会�?, len(sessions)), StatusInfo)

	case "export":
		if len(parts) >= 3 {
			content, err := m.ExportSession(parts[2])
			if err != nil {
				m.setStatus(fmt.Sprintf("导出失败: %v", err), StatusError)
				return m, nil
			}
			// 这里可以将内容写入文�?			m.setStatus(fmt.Sprintf("会话已导�? %d 字符", len(content)), StatusSuccess)
		}
	}

	return m, nil
}
