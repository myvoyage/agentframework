// AgentFramework - TUI (Terminal User Interface)
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"

	"AgentFramework/core"
	"AgentFramework/agent"
)

// ========== 视图类型 ==========

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

// ========== 消息类型 ==========

// AgentListMsg 加载 Agent 列表完成
type AgentListMsg struct {
	Agents []AgentItem
}

// WorkflowListMsg 加载工作流列表完成
type WorkflowListMsg struct {
	Workflows []WorkflowItem
}

// SkillListMsg 加载技能列表完成
type SkillListMsg struct {
	Skills []SkillItem
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// ChatResponseMsg Agent 响应消息
type ChatResponseMsg struct {
	Response string
	Error    error
}

// ========== TUI Model ==========

type TUIModel struct {
	// 核心服务
	ctx  context.Context
	core *core.Application

	// 界面状态
	width, height int
	currentView   View
	quitting      bool

	// 侧边栏
	sidebar list.Model

	// 主内容区
	viewport viewport.Model

	// 输入组件
	chatInput textarea.Model
	filterInput textinput.Model

	// 数据
	agents    []AgentItem
	workflows []WorkflowItem
	skills    []SkillItem
	messages  []ChatMessage
	selectedAgent string

	// 样式
	styles *Styles
}

// ========== 数据项类型 ==========

type AgentItem struct {
	ID   string
	Name string
	Type string
}

func (a AgentItem) Title() string       { return a.Name }
func (a AgentItem) Description() string { return fmt.Sprintf("Type: %s", a.Type) }
func (a AgentItem) FilterValue() string { return a.Name }

type WorkflowItem struct {
	ID          string
	Name        string
	Desc        string
	Status      string
}

func (w WorkflowItem) Title() string       { return w.Name }
func (w WorkflowItem) Description() string { return w.Desc }
func (w WorkflowItem) FilterValue() string { return w.Name }

type SkillItem struct {
	ID          string
	Name        string
	Desc        string
	Version     string
	Enabled     bool
}

func (s SkillItem) Title() string       { return s.Name }
func (s SkillItem) Description() string { return s.Desc }
func (s SkillItem) FilterValue() string { return s.Name }

// ========== 样式配置 ==========

type Styles struct {
	// 颜色
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color

	// 样式函数
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Border     lipgloss.Style
	Sidebar    lipgloss.Style
	ActiveItem lipgloss.Style
	NormalItem lipgloss.Style

	// 布局样式
	SidebarStyle    lipgloss.Style
	ContentStyle    lipgloss.Style
	InputStyle      lipgloss.Style
	MessageUser     lipgloss.Style
	MessageAssistant lipgloss.Style
}

func NewStyles() *Styles {
	s := &Styles{
		// 定义颜色方案
		Primary:   lipgloss.Color("#7D56F4"),
		Secondary: lipgloss.Color("#FA7970"),
		Success:   lipgloss.Color("#3E8E41"),
		Warning:   lipgloss.Color("#F5D647"),
		Error:     lipgloss.Color("#FF5F5F"),
		Muted:     lipgloss.Color("#626262"),
	}

	// 创建样式
	s.Title = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(s.Muted).
		MarginBottom(1)

	s.Border = lipgloss.NewStyle().
		Foreground(lipgloss.Color("236")).
		Bold(true)

	s.Sidebar = lipgloss.NewStyle().
		Width(25).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(s.Muted)

	s.ActiveItem = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true).
		Padding(0, 1)

	s.NormalItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)

	// 布局样式
	s.SidebarStyle = lipgloss.NewStyle().
		Width(25).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("#3C3C3C"))

	s.ContentStyle = lipgloss.NewStyle().
		Padding(1, 2)

	s.InputStyle = lipgloss.NewStyle().
		Width(60).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(s.Primary)

	s.MessageUser = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	s.MessageAssistant = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3E8E41")).
		Padding(0, 1)

	return s
}

// ========== Model 初始化 ==========

func NewTUIModel(ctx context.Context, coreApp *core.Application) TUIModel {
	styles := NewStyles()

	// 初始化侧边栏
	sidebarItems := []list.Item{
		MainItem{TitleStr: "Dashboard", Desc: "Overview and stats"},
		MainItem{TitleStr: "Agents", Desc: "Manage AI agents"},
		MainItem{TitleStr: "Chat", Desc: "Interact with agents"},
		MainItem{TitleStr: "Workflows", Desc: "Manage workflows"},
		MainItem{TitleStr: "Skills", Desc: "Manage skills"},
		MainItem{TitleStr: "Settings", Desc: "Configuration"},
		MainItem{TitleStr: "Logs", Desc: "View logs"},
	}

	sidebar := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	sidebar.SetItems(sidebarItems)
	sidebar.SetShowTitle(false)
	sidebar.SetShowStatusBar(false)
	sidebar.SetFilteringEnabled(false)
	sidebar.SetShowHelp(false)

	// 初始化输入框
	chatInput := textarea.New()
	chatInput.Placeholder = "Type your message..."
	chatInput.Focus()

	filterInput := textinput.New()
	filterInput.Placeholder = "Filter..."

	// 初始化视口
	viewport := viewport.New(0, 0)

	return TUIModel{
		ctx:       ctx,
		core:      coreApp,
		currentView: ViewDashboard,
		styles:    styles,
		sidebar:   sidebar,
		chatInput: chatInput,
		filterInput: filterInput,
		viewport:  viewport,
		messages:  []ChatMessage{},
	}
}

type MainItem struct {
	TitleStr string
	Desc     string
}

func (m MainItem) Title() string       { return m.TitleStr }
func (m MainItem) Description() string { return m.Desc }
func (m MainItem) FilterValue() string { return m.TitleStr }

// ========== Model 方法 ==========

// Init 初始化 Model
func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		m.loadAgents(),
		m.loadWorkflows(),
		m.loadSkills(),
		textarea.Blink,
	)
}

// Update 更新 Model 状态
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			m.cycleView()

		case "enter":
			if m.currentView == ViewChat && m.chatInput.Focused() {
				content := strings.TrimSpace(m.chatInput.Value())
				if content != "" && m.selectedAgent != "" {
					m.messages = append(m.messages, ChatMessage{
						Role:    "user",
						Content: content,
					})
					m.chatInput.SetValue("")
					cmds = append(cmds, m.sendChatMessage(content, m.selectedAgent))
				}
			}

		case "ctrl+r":
			// 刷新数据
			cmds = append(cmds,
				m.loadAgents(),
				m.loadWorkflows(),
				m.loadSkills(),
			)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case AgentListMsg:
		m.agents = msg.Agents

	case WorkflowListMsg:
		m.workflows = msg.Workflows

	case SkillListMsg:
		m.skills = msg.Skills

	case ChatResponseMsg:
		if msg.Error != nil {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Error: %v", msg.Error),
			})
		} else {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: msg.Response,
			})
		}
		// 更新视口内容
		m.viewport.SetContent(m.renderChatMessages())
		m.viewport.GotoBottom()
	}

	// 更新子组件
	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	cmds = append(cmds, cmd)

	if m.currentView == ViewChat {
		m.chatInput, cmd = m.chatInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.currentView == ViewAgents || m.currentView == ViewSkills || m.currentView == ViewWorkflows {
		m.filterInput, cmd = m.filterInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染界面
func (m TUIModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// 渲染侧边栏
	sidebar := m.renderSidebar()

	// 渲染主内容区
	content := m.renderContent()

	// 组合布局
	return lipgloss.JoinHorizontal(lipgloss.Left, sidebar, content)
}

// ========== 渲染方法 ==========

func (m TUIModel) renderSidebar() string {
	// 更新选中状态
	items := []list.Item{
		MainItem{TitleStr: "Dashboard", Desc: "Overview"},
		MainItem{TitleStr: "Agents", Desc: "AI Agents"},
		MainItem{TitleStr: "Chat", Desc: "Interact"},
		MainItem{TitleStr: "Workflows", Desc: "Automation"},
		MainItem{TitleStr: "Skills", Desc: "Capabilities"},
		MainItem{TitleStr: "Settings", Desc: "Config"},
		MainItem{TitleStr: "Logs", Desc: "Activity"},
	}
	m.sidebar.SetItems(items)
	m.sidebar.Select(int(m.currentView))

	return m.styles.SidebarStyle.Render(m.sidebar.View())
}

func (m TUIModel) renderContent() string {
	switch m.currentView {
	case ViewDashboard:
		return m.renderDashboard()
	case ViewAgents:
		return m.renderAgents()
	case ViewChat:
		return m.renderChat()
	case ViewWorkflows:
		return m.renderWorkflows()
	case ViewSkills:
		return m.renderSkills()
	case ViewSettings:
		return m.renderSettings()
	case ViewLogs:
		return m.renderLogs()
	default:
		return "Unknown view"
	}
}

func (m TUIModel) renderDashboard() string {
	title := m.styles.Title.Render("📊 Dashboard")

	stats := []string{
		fmt.Sprintf("Agents: %d", len(m.agents)),
		fmt.Sprintf("Workflows: %d", len(m.workflows)),
		fmt.Sprintf("Skills: %d", len(m.skills)),
		fmt.Sprintf("Messages: %d", len(m.messages)),
	}

	content := strings.Join(stats, "\n\n")

	// 添加操作提示
	help := "\n\n" + m.styles.Subtitle.Render("Controls:") +
		"\n  • Tab - Switch views\n" +
		"  • Ctrl+R - Refresh data\n" +
		"  • Q/Ctrl+C - Quit"

	return m.styles.ContentStyle.Render(title + "\n\n" + content + help)
}

func (m TUIModel) renderAgents() string {
	title := m.styles.Title.Render("🤖 Agents")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if len(m.agents) == 0 {
		content.WriteString("No agents available. Press Ctrl+R to refresh.")
	} else {
		for _, agent := range m.agents {
			line := fmt.Sprintf("  • %s (%s)", agent.Name, agent.Type)
			if agent.ID == m.selectedAgent {
				line = m.styles.ActiveItem.Render("▶ " + line)
			}
			content.WriteString(line + "\n")
		}
	}

	help := "\n\n" + m.styles.Subtitle.Render("Press Enter to select agent for chat")

	return m.styles.ContentStyle.Render(content.String() + help)
}

func (m TUIModel) renderChat() string {
	title := m.styles.Title.Render("💬 Chat")

	// 渲染消息
	messages := m.renderChatMessages()
	m.viewport.SetContent(messages)
	m.viewport.GotoBottom()

	// 渲染输入框
	input := m.styles.InputStyle.Render(m.chatInput.View())

	// 状态信息
	status := ""
	if m.selectedAgent != "" {
		status = fmt.Sprintf("Chatting with: %s", m.selectedAgent)
	} else {
		status = "No agent selected (go to Agents view first)"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.styles.ContentStyle.Render(title),
		m.viewport.View(),
		"\n",
		m.styles.Subtitle.Render(status),
		input,
	)
}

func (m TUIModel) renderChatMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		if msg.Role == "user" {
			b.WriteString(m.styles.MessageUser.Render("You: "+msg.Content))
		} else {
			b.WriteString(m.styles.MessageAssistant.Render("Assistant: "+msg.Content))
		}
		b.WriteString("\n\n")
	}

	return b.String()
}

func (m TUIModel) renderWorkflows() string {
	title := m.styles.Title.Render("⚙️ Workflows")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if len(m.workflows) == 0 {
		content.WriteString("No workflows available. Press Ctrl+R to refresh.")
	} else {
		for _, wf := range m.workflows {
			status := "●"
			if wf.Status == "running" {
				status = "🔄"
			} else if wf.Status == "completed" {
				status = "✓"
			} else if wf.Status == "failed" {
				status = "✗"
			}
			line := fmt.Sprintf("  %s %s - %s", status, wf.Name, wf.Desc)
			content.WriteString(line + "\n")
		}
	}

	help := "\n\n" + m.styles.Subtitle.Render("Press Enter to execute workflow")

	return m.styles.ContentStyle.Render(content.String() + help)
}

func (m TUIModel) renderSkills() string {
	title := m.styles.Title.Render("🎯 Skills")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	if len(m.skills) == 0 {
		content.WriteString("No skills available. Press Ctrl+R to refresh.")
	} else {
		for _, skill := range m.skills {
			status := "✓"
			if !skill.Enabled {
				status = "○"
			}
			line := fmt.Sprintf("  %s %s (v%s) - %s", status, skill.Name, skill.Version, skill.Desc)
			content.WriteString(line + "\n")
		}
	}

	help := "\n\n" + m.styles.Subtitle.Render("Press Enter to toggle skill, E to enable, D to disable")

	return m.styles.ContentStyle.Render(content.String() + help)
}

func (m TUIModel) renderSettings() string {
	title := m.styles.Title.Render("⚙️ Settings")

	content := title + "\n\n" +
		"  Model Configuration:\n" +
		"    • Default: ollama/llama3\n" +
		"    • Skill System: Enabled\n" +
		"    • Cache: Enabled\n\n" +
		"  Storage:\n" +
		"    • Local Config: ~/.agentframework\n" +
		"    • Conversations: ~/.agentframework/conversations\n\n" +
		"  API:\n" +
		"    • HTTP Server: localhost:8080\n" +
		"    • WebSocket: ws://localhost:8080/ws"

	return m.styles.ContentStyle.Render(content)
}

func (m TUIModel) renderLogs() string {
	title := m.styles.Title.Render("📋 Logs")

	content := title + "\n\n" +
		"Recent activity:\n\n" +
		"  [INFO] TUI started\n" +
		"  [INFO] Loaded " + fmt.Sprintf("%d", len(m.agents)) + " agents\n" +
		"  [INFO] Loaded " + fmt.Sprintf("%d", len(m.workflows)) + " workflows\n" +
		"  [INFO] Loaded " + fmt.Sprintf("%d", len(m.skills)) + " skills\n"

	return m.styles.ContentStyle.Render(content)
}

// ========== 辅助方法 ==========

func (m *TUIModel) updateLayout() {
	// 更新侧边栏高度
	m.sidebar.SetHeight(m.height)

	// 更新视口大小
	viewportHeight := m.height - 10 // 留出空间给标题和输入
	m.viewport.Height = viewportHeight
	m.viewport.Width = m.width - 30 // 减去侧边栏宽度
}

func (m *TUIModel) cycleView() {
	m.currentView = (m.currentView + 1) % 7
	if m.currentView == ViewChat {
		m.chatInput.Focus()
	} else {
		m.chatInput.Blur()
	}
}

// ========== 数据加载命令 ==========

func (m TUIModel) loadAgents() tea.Cmd {
	return func() tea.Msg {
		agents := m.core.GetHost().ListAgents()
		items := make([]AgentItem, 0, len(agents))
		for _, id := range agents {
			agent, err := m.core.GetHost().GetAgent(id)
			if err == nil && agent != nil {
				items = append(items, AgentItem{
					ID:   id,
					Name: agent.Name(),
					Type: fmt.Sprintf("%T", agent),
				})
			}
		}
		return AgentListMsg{Agents: items}
	}
}

func (m TUIModel) loadWorkflows() tea.Cmd {
	return func() tea.Msg {
		workflows, err := m.core.GetWorkflowManager().GetWorkflows(m.ctx)
		items := make([]WorkflowItem, 0)
		if err == nil {
			for _, wf := range workflows {
				items = append(items, WorkflowItem{
					ID:     wf.ID,
					Name:   wf.Name,
					Desc:   wf.Description,
					Status: "pending",
				})
			}
		}
		return WorkflowListMsg{Workflows: items}
	}
}

func (m TUIModel) loadSkills() tea.Cmd {
	return func() tea.Msg {
		skills := m.core.GetSkillLibrary().GetAllSkills(m.ctx)
		items := make([]SkillItem, 0)
		for name, skill := range skills {
			metadata := skill.GetMetadata(m.ctx)
			items = append(items, SkillItem{
				ID:      name,
				Name:    metadata.Name,
				Desc:    metadata.Description,
				Version: metadata.Version,
				Enabled: skill.IsEnabled(m.ctx),
			})
		}
		return SkillListMsg{Skills: items}
	}
}

func (m TUIModel) sendChatMessage(message, agentID string) tea.Cmd {
	return func() tea.Msg {
		agent, err := m.core.GetHost().GetAgent(agentID)
		if err != nil || agent == nil {
			return ChatResponseMsg{Error: fmt.Errorf("agent not found")}
		}

		response, err := agent.Run(m.ctx, message)
		if err != nil {
			return ChatResponseMsg{Error: err}
		}

		return ChatResponseMsg{Response: response.Content}
	}
}

// ========== 主入口 ==========

func main() {
	ctx := context.Background()

	// 初始化核心应用
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
	}

	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	coreApp, err := core.NewApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create core application: %v\n", err)
		os.Exit(1)
	}

	if err := coreApp.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize core application: %v\n", err)
		os.Exit(1)
	}

	// 初始化工作流管理器和文件浏览器
	coreApp.GetWorkflowManager().Init(ctx)
	coreApp.GetFileExplorer().Init(ctx)

	// 创建 TUI 模型
	model := NewTUIModel(ctx, coreApp)

	// 启动 TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
