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
// Agent Framework - TUI Messages
// Copyright (C) 2025 Agent Framework Contributors
//
// 消息系统定义 - 借鉴 Memoh 的模块化消息设计

package tui

import (
	"time"
)

// ========== 视图类型 ==========

// View 视图类型
type View int

const (
	ViewDashboard View = iota
	ViewAgents
	ViewChat
	ViewWorkflows
	ViewSkills
	ViewSettings
	ViewLogs
)

func (v View) String() string {
	return [...]string{"Dashboard", "Agents", "Chat", "Workflows", "Skills", "Settings", "Logs"}[v]
}

// ========== 基础消息 ==========

// TickMsg 定时器消�?type TickMsg time.Time

// ========== 数据项类�?==========

// AgentItem Agent 数据�?type AgentItem struct {
	ID   string
	Name string
	Type string
}

func (a AgentItem) Title() string       { return a.Name }
func (a AgentItem) Description() string { return a.Type }
func (a AgentItem) FilterValue() string { return a.Name }

// WorkflowItem 工作流数据项
type WorkflowItem struct {
	ID     string
	Name   string
	Desc   string
	Status string
}

func (w WorkflowItem) Title() string       { return w.Name }
func (w WorkflowItem) Description() string { return w.Desc }
func (w WorkflowItem) FilterValue() string { return w.Name }

// SkillItem 技能数据项
type SkillItem struct {
	ID      string
	Name    string
	Desc    string
	Version string
	Enabled bool
}

func (s SkillItem) Title() string       { return s.Name }
func (s SkillItem) Description() string { return s.Desc }
func (s SkillItem) FilterValue() string { return s.Name }

// ========== 数据加载消息 ==========

// AgentListLoadedMsg Agent 列表加载完成
type AgentListLoadedMsg struct {
	Agents []AgentItem
	Error  error
}

// WorkflowListLoadedMsg 工作流列表加载完�?type WorkflowListLoadedMsg struct {
	Workflows []WorkflowItem
	Error     error
}

// SkillListLoadedMsg 技能列表加载完�?type SkillListLoadedMsg struct {
	Skills []SkillItem
	Error  error
}

// ConfigLoadedMsg 配置加载完成
type ConfigLoadedMsg struct {
	Config map[string]interface{}
	Error  error
}

// ========== 聊天消息 ==========

// ChatSendMsg 发送聊天消�?type ChatSendMsg struct {
	Content string
	AgentID string
}

// ChatResponseMsg 聊天响应消息（流式）
type ChatResponseMsg struct {
	ContentChunk string // 内容片段，用于流式显�?	Done         bool   // 是否完成
	Error        error
}

// ChatHistoryLoadedMsg 聊天历史加载完成
type ChatHistoryLoadedMsg struct {
	Messages []ChatMessageItem
	Error    error
}

// ========== 状态变更消�?==========

// ViewChangeMsg 视图切换消息
type ViewChangeMsg struct {
	From View
	To   View
}

// SelectionChangeMsg 选择项变更消�?type SelectionChangeMsg struct {
	ItemType string // "agent", "workflow", "skill", etc.
	ItemID   string
}

// StatusUpdateMsg 状态更新消�?type StatusUpdateMsg struct {
	Message string
	Type    StatusType
}

// StatusType 状态类�?type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
	StatusLoading
)

// ========== 操作消息 ==========

// AgentActionMsg Agent 操作消息
type AgentActionMsg struct {
	Action string // "start", "stop", "restart"
	AgentID string
}

// WorkflowActionMsg 工作流操作消�?type WorkflowActionMsg struct {
	Action    string // "execute", "stop", "delete"
	WorkflowID string
	Input      string
}

// SkillActionMsg 技能操作消�?type SkillActionMsg struct {
	Action  string // "enable", "disable", "run"
	SkillID string
	Input   string
}

// RefreshMsg 刷新消息
type RefreshMsg struct {
	Target string // "current", "agents", "workflows", "skills", "all"
}

// ========== 聊天数据�?==========

// ChatMessageItem 聊天消息�?type ChatMessageItem struct {
	ID        string
	Role      string // "user", "assistant", "system"
	Content   string
	Timestamp time.Time
	AgentID   string
	Streaming bool   // 是否正在流式输出
}

// IsStreaming 检查消息是否正在流式输�?func (m ChatMessageItem) IsStreaming() bool {
	return m.Streaming
}
