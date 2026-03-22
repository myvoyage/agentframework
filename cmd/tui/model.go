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
// Agent Framework - TUI Main Model (Refactored)
// Copyright (C) 2025 Agent Framework Contributors
//
// 重构后的主模�?- 借鉴 Memoh 的模块化架构

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"AgentFramework/core"
)

// ========== 主模�?==========

// Model �?TUI 模型
type Model struct {
	// 核心服务
	ctx            context.Context
	core           *core.Application
	config         *ConfigManager
	session        *SessionManager
	integration    *IntegrationLayer // 新增：集成层

	// 样式
	styles *StyleManager

	// 界面状�?	width, height int
	currentView   View
	quitting      bool
	ready         bool

	// 侧边栏（导航�?	sidebar     list.Model
	sidebarKeys *SidebarKeys

	// 主内容区
	viewport viewport.Model

	// 输入组件
	cmdInput textinput.Model

	// 数据状�?	agents       []AgentItem
	workflows    []WorkflowItem
	skills       []SkillItem
	selectedItem interface{}

	// 聊天状�?	chatMessages []ChatMessageItem
	currentAgent string
	streamingSession *StreamingChatSession

	// 状态栏
	statusMessage string
	statusType    StatusType
	statusTime    time.Time

	// 加载状�?	loadingAgents    bool
	loadingWorkflows bool
	loadingSkills    bool
}

// NewTUIModel 创建新的 TUI 模型
func NewTUIModel(ctx context.Context, coreApp *core.Application) *Model {
	// 初始化样�?	styles := NewStyleManager()

	// 创建输入�?	ti := textinput.New()
	ti.Placeholder = "输入命令或消�?.."
	ti.Focus()

	// 创建侧边�?	sidebar := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sidebar.SetShowStatusBar(false)
	sidebar.SetFilteringEnabled(false)
	sidebar.Styles.Title = styles.Header
	sidebar.Styles.FilterPrompt = styles.Muted

	// 初始化所有视图项
	sidebar.SetItems(getAllViewItems(styles))

	// 创建模型
	m := &Model{
		ctx:       ctx,
		core:      coreApp,
		styles:    styles,
		cmdInput:  ti,
		sidebar:   sidebar,
		sidebarKeys: &SidebarKeys{
			ForwardKeys:  KeyMap{"tab", "ctrl+right"},
			BackwardKeys: KeyMap{"shift+tab", "ctrl+left"},
		},
		currentView: ViewDashboard,
		chatMessages: make([]ChatMessageItem, 0),
	}

	// 创建集成�?	m.integration = NewIntegrationLayer(ctx, coreApp)

	// 初始化配置和会话管理
	if configPath, err := GetConfigPath(); err == nil {
		if cfg, err := NewConfigManager(configPath); err == nil {
			m.config = cfg
		}
	}

	if sessionsPath, err := GetSessionsPath(); err == nil {
		if sess, err := NewSessionManager(sessionsPath); err == nil {
			m.session = sess
		}
	}

	return m
}

// ========== Bubble Tea 接口实现 ==========

// Init 初始化模�?func (m *Model) Init() tea.Cmd {
	// 启动定时�?	ticker := func() tea.Msg {
		return TickMsg(time.Now())
	}

	// 加载初始数据
	cmds := []tea.Cmd{
		tea.Batch(ticker, m.loadData()),
	}

	return tea.Batch(cmds...)
}

// Update 更新模型状�?func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, nil

	case TickMsg:
		// 定时更新状态栏
		if time.Since(m.statusTime) > 5*time.Second {
			m.statusMessage = ""
		}
		cmds = append(cmds, func() tea.Msg {
			return TickMsg(time.Now())
		})

	case AgentListLoadedMsg:
		m.loadingAgents = false
		if msg.Error != nil {
			m.setStatus(fmt.Sprintf("加载 Agent 失败: %v", msg.Error), StatusError)
		} else {
			m.agents = msg.Agents
			m.setStatus(fmt.Sprintf("已加�?%d �?Agent", len(msg.Agents)), StatusSuccess)
		}

	case WorkflowListLoadedMsg:
		m.loadingWorkflows = false
		if msg.Error != nil {
			m.setStatus(fmt.Sprintf("加载工作流失�? %v", msg.Error), StatusError)
		} else {
			m.workflows = msg.Workflows
			m.setStatus(fmt.Sprintf("已加�?%d 个工作流", len(msg.Workflows)), StatusSuccess)
		}

	case SkillListLoadedMsg:
		m.loadingSkills = false
		if msg.Error != nil {
			m.setStatus(fmt.Sprintf("加载技能失�? %v", msg.Error), StatusError)
		} else {
			m.skills = msg.Skills
			m.setStatus(fmt.Sprintf("已加�?%d 个技�?, len(msg.Skills)), StatusSuccess)
		}

	case ChatSendMsg:
		return m.handleChatSend(msg)

	case ChatResponseMsg:
		return m.handleChatResponse(msg)

	case ViewChangeMsg:
		m.currentView = msg.To
		cmds = append(cmds, m.renderView())

	case RefreshMsg:
		cmds = append(cmds, m.loadData())

	case StatusUpdateMsg:
		m.setStatus(msg.Message, msg.Type)
	}

	// 更新子组�?	m.updateSubviews(msg)

	return m, tea.Batch(cmds...)
}

// View 渲染视图
func (m *Model) View() string {
	if !m.ready {
		return "\n  正在初始�?AgentFramework TUI..."
	}

	// 计算布局
	sidebarWidth := 20
	contentWidth := m.width - sidebarWidth - 2

	// 渲染侧边�?	sidebar := m.styles.Card.
		Width(sidebarWidth).
		Height(m.height).
		Render(m.renderSidebar())

	// 渲染主内容区
	content := m.renderContentArea(contentWidth, m.height)

	// 组合布局
	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	return layout
}

// ========== 事件处理 ==========

// handleKeyPress 处理键盘事件
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 全局快捷�?	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "ctrl+r":
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return RefreshMsg{Target: "all"}
		})

	case "tab":
		// 切换到下一个视�?		nextView := View((int(m.currentView) + 1) % 7)
		return m, func() tea.Msg {
			return ViewChangeMsg{From: m.currentView, To: nextView}
		}

	case "shift+tab":
		// 切换到上一个视�?		prevView := View((int(m.currentView) - 1 + 7) % 7)
		return m, func() tea.Msg {
			return ViewChangeMsg{From: m.currentView, To: prevView}
		}

	case "enter":
		// 处理输入
		if m.cmdInput.Value() != "" {
			return m.handleCommand(m.cmdInput.Value())
		}
	}

	// 视图特定的键盘处�?	return m.handleViewKeyPress(msg)
}

// handleResize 处理窗口大小变化
func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height

	// 更新侧边栏大�?	m.sidebar.SetWidth(20)
	m.sidebar.SetHeight(msg.Height)

	// 更新视口大小
	m.viewport.Width = m.width - 22
	m.viewport.Height = m.height - 5

	m.ready = true
}

// handleCommand 处理命令输入
func (m *Model) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	m.cmdInput.Reset()

	// 解析命令
	parts := parseCommand(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	// ── 对话 ──────────────────────────────────────────────────────────────
	case "chat":
		if len(parts) > 1 {
			msg := strings.Join(parts[1:], " ")
			return m.handleChatSend(ChatSendMsg{Content: msg, AgentID: m.currentAgent})
		}

	// ── 核心子系�?────────────────────────────────────────────────────────
	case "agent", "a":
		return m.handleAgentCommand(parts)

	case "workflow", "wf", "w":
		return m.handleWorkflowCommand(parts)

	case "skill", "sk":
		return m.handleSkillCommand(parts)

	case "config", "cfg":
		return m.handleConfigCommand(parts)

	case "file", "f":
		return m.handleFileCommand(parts)

	// ── 高级子系�?────────────────────────────────────────────────────────
	case "task", "t":
		return m.handleTaskCommand(parts)

	case "schedule", "sched":
		return m.handleScheduleCommand(parts)

	case "channel", "ch", "msg":
		return m.handleChannelCommand(parts)

	case "plugin", "pl":
		return m.handlePluginCommand(parts)

	case "monitor", "mon":
		return m.handleMonitorCommand(parts)

	case "token":
		return m.handleTokenCommand(parts)

	case "host", "instance":
		return m.handleHostCommand(parts)

	// ── 会话管理 ──────────────────────────────────────────────────────────
	case "session":
		return m.handleSessionCommand(parts)

	// ── 视图快捷切换 ──────────────────────────────────────────────────────
	case "view", "v":
		if len(parts) >= 2 {
			return m.handleViewSwitch(parts[1])
		}
		m.setStatus("用法: view <dashboard|agents|chat|workflows|skills|settings|logs>", StatusWarning)

	// ── 系统命令 ──────────────────────────────────────────────────────────
	case "refresh", "r":
		return m, func() tea.Msg { return RefreshMsg{Target: "all"} }

	case "clear":
		m.chatMessages = make([]ChatMessageItem, 0)
		m.setStatus("对话历史已清�?, StatusSuccess)

	case "quit", "exit", "bye":
		m.quitting = true
		return m, tea.Quit

	case "help", "?", "h":
		m.currentView = ViewSettings
		return m, nil

	default:
		m.setStatus(fmt.Sprintf("未知命令: %s (输入 'help' 查看帮助)", parts[0]), StatusWarning)
	}

	return m, nil
}

// handleViewSwitch 处理视图切换命令
func (m *Model) handleViewSwitch(viewName string) (tea.Model, tea.Cmd) {
	viewMap := map[string]View{
		"dashboard": ViewDashboard,
		"dash":      ViewDashboard,
		"agents":    ViewAgents,
		"agent":     ViewAgents,
		"chat":      ViewChat,
		"workflows": ViewWorkflows,
		"workflow":  ViewWorkflows,
		"wf":        ViewWorkflows,
		"skills":    ViewSkills,
		"skill":     ViewSkills,
		"settings":  ViewSettings,
		"config":    ViewSettings,
		"logs":      ViewLogs,
		"log":       ViewLogs,
	}

	if v, ok := viewMap[strings.ToLower(viewName)]; ok {
		oldView := m.currentView
		m.currentView = v
		return m, func() tea.Msg {
			return ViewChangeMsg{From: oldView, To: v}
		}
	}

	m.setStatus(fmt.Sprintf("未知视图: %s", viewName), StatusWarning)
	return m, nil
}

// ========== 数据加载 ==========

// loadData 加载数据
func (m *Model) loadData() tea.Cmd {
	m.loadingAgents = true
	m.loadingWorkflows = true
	m.loadingSkills = true

	return tea.Batch(
		m.integration.BatchLoadAgentsCmd(),
		m.integration.BatchLoadWorkflowsCmd(),
		m.integration.BatchLoadSkillsCmd(),
	)
}

// ========== 状态管�?==========

// setStatus 设置状态消�?func (m *Model) setStatus(message string, statusType StatusType) {
	m.statusMessage = message
	m.statusType = statusType
	m.statusTime = time.Now()
}

// ========== 辅助类型 ==========

// KeyMap 键盘映射
type KeyMap []string

// SidebarKeys 侧边栏快捷键
type SidebarKeys struct {
	ForwardKeys  KeyMap
	BackwardKeys KeyMap
}

// ========== 视图项创�?==========

// ViewItem 视图�?type ViewItem struct {
	name  string
	view  View
	desc  string
}

func (v ViewItem) Title() string       { return v.name }
func (v ViewItem) Description() string { return v.desc }
func (v ViewItem) FilterValue() string { return v.name }

// getAllViewItems 获取所有视图项
func getAllViewItems(styles *StyleManager) []list.Item {
	views := []struct {
		name string
		view View
		desc string
	}{
		{"📊 Dashboard", ViewDashboard, "系统概览和统�?},
		{"💬 Agents", ViewAgents, "Agent 管理"},
		{"💭 Chat", ViewChat, "对话界面"},
		{"🔄 Workflows", ViewWorkflows, "工作流管�?},
		{"�?Skills", ViewSkills, "技能管�?},
		{"⚙️ Settings", ViewSettings, "配置设置"},
		{"📋 Logs", ViewLogs, "日志查看"},
	}

	items := make([]list.Item, len(views))
	for i, v := range views {
		items[i] = ViewItem{name: v.name, view: v.view, desc: v.desc}
	}

	return items
}

// ========== 命令解析 ==========

// parseCommand 解析命令
func parseCommand(cmd string) []string {
	// 简单的命令解析
	// TODO: 改进为支持引号和转义的解析器
	return strings.Fields(cmd)
}

// ========== 子组件更�?==========

// updateSubviews 更新子组件状�?func (m *Model) updateSubviews(msg tea.Msg) {
	// 更新侧边�?	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	_ = cmd // 忽略侧边栏的命令

	// 更新输入�?	m.cmdInput, cmd = m.cmdInput.Update(msg)
	_ = cmd
}

// ========== 视图渲染辅助方法 ==========

// renderSidebar 渲染侧边�?func (m *Model) renderSidebar() string {
	return m.sidebar.View()
}

// renderContentArea 渲染主内容区
func (m *Model) renderContentArea(width, height int) string {
	switch m.currentView {
	case ViewDashboard:
		return m.renderDashboard(width, height)
	case ViewAgents:
		return m.renderAgents(width, height)
	case ViewChat:
		return m.renderChat(width, height)
	case ViewWorkflows:
		return m.renderWorkflows(width, height)
	case ViewSkills:
		return m.renderSkills(width, height)
	case ViewSettings:
		return m.renderSettings(width, height)
	case ViewLogs:
		return m.renderLogs(width, height)
	default:
		return "未知视图"
	}
}

// renderView 渲染当前视图（返回命令）
func (m *Model) renderView() tea.Cmd {
	// 这里可以根据需要发送消息来更新视图
	return nil
}

// ========== 视图特定处理 ==========

// handleViewKeyPress 处理视图特定的键盘事�?func (m *Model) handleViewKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 根据当前视图处理特定按键
	switch m.currentView {
	case ViewChat:
		return m.handleChatKeyPress(msg)
	default:
		return m, nil
	}
}

// handleChatKeyPress 处理聊天视图的键盘事�?func (m *Model) handleChatKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 聊天视图的特定按键处�?	return m, nil
}

// handleChatSend 处理聊天消息发�?func (m *Model) handleChatSend(msg ChatSendMsg) (tea.Model, tea.Cmd) {
	// 添加用户消息
	userMsg := ChatMessageItem{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   msg.Content,
		Timestamp: time.Now(),
		AgentID:   msg.AgentID,
	}
	m.chatMessages = append(m.chatMessages, userMsg)

	// 发送消息并获取响应（使用集成层�?	cmd := m.integration.StreamChatCmd(msg.AgentID, msg.Content, "")

	// 自动保存会话
	if m.config != nil && m.config.Get().AutoSaveSession {
		// 使用 tea.Batch 组合多个命令
		return m, tea.Batch(cmd, m.AutoSaveSession())
	}

	return m, cmd
}

// handleChatResponse 处理聊天响应
func (m *Model) handleChatResponse(msg ChatResponseMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.setStatus(fmt.Sprintf("聊天失败: %v", msg.Error), StatusError)
		return m, nil
	}

	// 添加响应消息
	if len(m.chatMessages) > 0 {
		lastIdx := len(m.chatMessages) - 1
		if m.chatMessages[lastIdx].Role == "assistant" && m.chatMessages[lastIdx].Streaming {
			// 更新正在流式输出的消�?			m.chatMessages[lastIdx].Content += msg.ContentChunk
			if msg.Done {
				m.chatMessages[lastIdx].Streaming = false

				// 流式输出完成，保存会�?				if m.config != nil && m.config.Get().AutoSaveSession {
					return m, m.AutoSaveSession()
				}
			}
		} else {
			// 创建新的响应消息
			assistantMsg := ChatMessageItem{
				ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
				Role:      "assistant",
				Content:   msg.ContentChunk,
				Timestamp: time.Now(),
				AgentID:   m.currentAgent,
				Streaming: !msg.Done,
			}
			m.chatMessages = append(m.chatMessages, assistantMsg)
		}
	}

	return m, nil
}

// ========== 命令处理 ==========

// handleAgentCommand 处理 Agent 相关命令
func (m *Model) handleAgentCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.currentView = ViewAgents
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.currentView = ViewAgents
		return m, func() tea.Msg { return RefreshMsg{Target: "agents"} }
	case "select", "use":
		if len(parts) >= 3 {
			m.currentAgent = parts[2]
			m.setStatus(fmt.Sprintf("已选择 Agent: %s", parts[2]), StatusSuccess)
		} else {
			m.setStatus("用法: agent select <agent-id>", StatusWarning)
		}
	case "chat":
		if len(parts) >= 3 {
			m.currentAgent = parts[2]
		}
		m.currentView = ViewChat
	case "run":
		if len(parts) >= 4 {
			agentID := parts[2]
			task := strings.Join(parts[3:], " ")
			m.currentAgent = agentID
			m.currentView = ViewChat
			return m.handleChatSend(ChatSendMsg{Content: task, AgentID: agentID})
		}
		m.setStatus("用法: agent run <agent-id> <task>", StatusWarning)
	case "info", "describe", "desc":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在获取 Agent '%s' 信息...", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: agent info <agent-id>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? agent %s (list|select|chat|run|info)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleWorkflowCommand 处理工作流相关命�?func (m *Model) handleWorkflowCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.currentView = ViewWorkflows
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.currentView = ViewWorkflows
		return m, func() tea.Msg { return RefreshMsg{Target: "workflows"} }
	case "execute", "run", "exec":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在执行工作�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: workflow run <name>", StatusWarning)
		}
	case "info", "describe", "desc":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("工作�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: workflow info <name>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? workflow %s (list|run|info)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleSkillCommand 处理技能相关命�?func (m *Model) handleSkillCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.currentView = ViewSkills
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.currentView = ViewSkills
		return m, func() tea.Msg { return RefreshMsg{Target: "skills"} }
	case "enable":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在启用技�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: skill enable <name>", StatusWarning)
		}
	case "disable":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在禁用技�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: skill disable <name>", StatusWarning)
		}
	case "info":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("技�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: skill info <name>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? skill %s (list|enable|disable|info)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleConfigCommand 处理配置相关命令
func (m *Model) handleConfigCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.currentView = ViewSettings
		return m, nil
	}

	switch parts[1] {
	case "show", "list", "ls":
		m.currentView = ViewSettings
	case "get":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("配置�? %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: config get <key>", StatusWarning)
		}
	case "set":
		if len(parts) >= 4 {
			m.setStatus(fmt.Sprintf("设置 %s = %s", parts[2], parts[3]), StatusInfo)
		} else {
			m.setStatus("用法: config set <key> <value>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? config %s (show|get|set)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleFileCommand 处理文件相关命令
func (m *Model) handleFileCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? file list|read|write|delete|mkdir|info", StatusInfo)
		return m, nil
	}

	m.setStatus(fmt.Sprintf("文件命令: %s (请使�?CLI 模式执行文件操作)", strings.Join(parts, " ")), StatusWarning)
	return m, nil
}

// handleTaskCommand 处理异步任务相关命令
func (m *Model) handleTaskCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? task list|get|cancel|wait|stats", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.setStatus("正在加载任务列表...", StatusInfo)
	case "stats":
		m.setStatus("正在获取任务统计...", StatusInfo)
	case "cancel":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在取消任务: %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: task cancel <task-id>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? task %s (list|get|cancel|wait|stats)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleScheduleCommand 处理调度相关命令
func (m *Model) handleScheduleCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? schedule status|info|heartbeat|heartbeat-info", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "status":
		m.setStatus("正在获取调度器状�?..", StatusInfo)
	case "info":
		m.setStatus("正在获取调度器配�?..", StatusInfo)
	case "heartbeat":
		m.setStatus("正在发送心跳信�?..", StatusInfo)
	case "heartbeat-info":
		m.setStatus("正在获取心跳配置...", StatusInfo)
	default:
		m.setStatus(fmt.Sprintf("未知子命�? schedule %s", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleChannelCommand 处理消息通道相关命令
func (m *Model) handleChannelCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? channel list|info|status|config", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.setStatus("正在获取通道列表...", StatusInfo)
	case "status":
		m.setStatus("正在获取通道状�?..", StatusInfo)
	case "info":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在获取通道 '%s' 信息...", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: channel info <channel-id>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? channel %s (list|info|status|config)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handlePluginCommand 处理插件相关命令
func (m *Model) handlePluginCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? plugin list|info|enable|disable|load|unload|reload", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "list", "ls":
		m.setStatus("正在获取插件列表...", StatusInfo)
	case "enable":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在启用插件: %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: plugin enable <name>", StatusWarning)
		}
	case "disable":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在禁用插件: %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: plugin disable <name>", StatusWarning)
		}
	case "info":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("插件: %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: plugin info <name>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? plugin %s (list|info|enable|disable|load|unload)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleMonitorCommand 处理监控相关命令
func (m *Model) handleMonitorCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? monitor status|list|metrics|stats|start|stop|alerts", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "status":
		m.setStatus("正在获取监控系统状�?..", StatusInfo)
	case "list", "ls":
		m.setStatus("正在获取监控器列�?..", StatusInfo)
	case "metrics":
		m.setStatus("正在获取监控指标...", StatusInfo)
	case "start":
		m.setStatus("正在启动监控�?..", StatusInfo)
	case "stop":
		m.setStatus("正在停止监控�?..", StatusInfo)
	case "alerts":
		m.setStatus("正在获取告警规则...", StatusInfo)
	default:
		m.setStatus(fmt.Sprintf("未知子命�? monitor %s", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleTokenCommand 处理 Token 压缩相关命令
func (m *Model) handleTokenCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.setStatus("可用子命�? token stats|config|strategy|count|compress-text", StatusInfo)
		return m, nil
	}

	switch parts[1] {
	case "stats":
		m.setStatus("正在获取 Token 压缩统计...", StatusInfo)
	case "config":
		m.setStatus("正在获取 Token 压缩配置...", StatusInfo)
	case "strategy":
		if len(parts) >= 3 {
			m.setStatus(fmt.Sprintf("正在设置压缩策略: %s", parts[2]), StatusInfo)
		} else {
			m.setStatus("用法: token strategy <truncate|summarize|hybrid|sliding_window>", StatusWarning)
		}
	case "count":
		if len(parts) >= 3 {
			text := strings.Join(parts[2:], " ")
			m.setStatus(fmt.Sprintf("统计 Token: %q", text), StatusInfo)
		} else {
			m.setStatus("用法: token count <text>", StatusWarning)
		}
	default:
		m.setStatus(fmt.Sprintf("未知子命�? token %s (stats|config|strategy|count)", parts[1]), StatusWarning)
	}

	return m, nil
}

// handleHostCommand 处理 Host 相关命令
func (m *Model) handleHostCommand(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.currentView = ViewDashboard
		return m, nil
	}

	switch parts[1] {
	case "info":
		m.currentView = ViewDashboard
	case "summary":
		m.currentView = ViewDashboard
	case "config":
		m.currentView = ViewSettings
	case "models":
		m.setStatus("正在获取模型配置...", StatusInfo)
	default:
		m.setStatus(fmt.Sprintf("未知子命�? host %s (info|summary|config|models)", parts[1]), StatusWarning)
	}

	return m, nil
}

// ========== 渲染方法（已移至 views.go�?=========

// 所有渲染方法已移至 views.go 文件�?// 这里保留是为了兼容性，实际使用 views.go 中的增强版本
