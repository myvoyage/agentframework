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
// Agent Framework - TUI Styles
// Copyright (C) 2025 Agent Framework Contributors
//
// 统一样式配置 - 借鉴 Memoh 的主题系统和 lipgloss 样式

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// ========== 样式配置 ==========

// StyleManager 统一样式管理�?type StyleManager struct {
	// 颜色定义
	Colors ColorPalette

	// 预定义样�?	Header      lipgloss.Style
	SubHeader   lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Highlight   lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style

	// 组件样式
	SidebarItem       lipgloss.Style
	SidebarActive     lipgloss.Style
	Card              lipgloss.Style
	CardHeader        lipgloss.Style
	CardBorder        lipgloss.Style
	TableCell         lipgloss.Style
	TableHeader       lipgloss.Style
	TableRow          lipgloss.Style
	TableRowAlt       lipgloss.Style
	TableRowSelected  lipgloss.Style

	// 输入框样�?	Input        lipgloss.Style
	InputFocused lipgloss.Style
	InputError   lipgloss.Style

	// 按钮样式
	Button        lipgloss.Style
ButtonActive   lipgloss.Style
	ButtonDisabled lipgloss.Style

	// 状态样�?	StatusDot     lipgloss.Style
	StatusText    lipgloss.Style
}

// ========== 颜色调色�?==========

// ColorPalette 颜色调色�?type ColorPalette struct {
	// 主色
	Primary   lipgloss.Color
	Secondary lipgloss.Color

	// 语义�?	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// 中性色
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Dim        lipgloss.Color

	// 特殊�?	Accent lipgloss.Color
}

// DefaultColorPalette 默认调色�?func DefaultColorPalette() ColorPalette {
	return ColorPalette{
		Primary:     lipgloss.Color("#007AFF"), // 蓝色
		Secondary:   lipgloss.Color("#5856D6"), // 紫色
		Success:     lipgloss.Color("#34C759"), // 绿色
		Warning:     lipgloss.Color("#FF9500"), // 橙色
		Error:       lipgloss.Color("#FF3B30"), // 红色
		Info:        lipgloss.Color("#5AC8FA"), // 青色
		Background:  lipgloss.Color("#1E1E1E"), // 深灰背景
		Foreground:  lipgloss.Color("#FFFFFF"), // 白色前景
		Muted:       lipgloss.Color("#8E8E93"), // 灰色
		Dim:         lipgloss.Color("#48484A"), // 暗灰
		Accent:      lipgloss.Color("#FF2D55"), // 粉色
	}
}

// ========== 样式初始�?==========

// NewStyleManager 创建新的样式管理�?func NewStyleManager() *StyleManager {
	colors := DefaultColorPalette()

	return &StyleManager{
		Colors:   colors,
		Header:   lipgloss.NewStyle().Foreground(colors.Primary).Bold(true),
		SubHeader: lipgloss.NewStyle().Foreground(colors.Secondary).Bold(true),
		Body:     lipgloss.NewStyle().Foreground(colors.Foreground),
		Muted:    lipgloss.NewStyle().Foreground(colors.Muted),
		Highlight: lipgloss.NewStyle().Foreground(colors.Accent).Bold(true),
		Error:    lipgloss.NewStyle().Foreground(colors.Error),
		Success:  lipgloss.NewStyle().Foreground(colors.Success),
		Warning:  lipgloss.NewStyle().Foreground(colors.Warning),
		Info:     lipgloss.NewStyle().Foreground(colors.Info),

		// 组件样式
		SidebarItem: lipgloss.NewStyle().
				Foreground(colors.Foreground).
				Padding(0, 1),
		SidebarActive: lipgloss.NewStyle().
				Foreground(colors.Background).
				Background(colors.Primary).
				Bold(true).
				Padding(0, 1),
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Muted).
			Padding(1),
		CardHeader: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true).
			MarginBottom(1),
		CardBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colors.Muted),

		// 表格样式
		TableCell: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Padding(0, 1),
		TableHeader: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true).
			Padding(0, 1),
		TableRow: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Padding(0, 1),
		TableRowAlt: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Background(lipgloss.Color("#2C2C2E")).
			Padding(0, 1),
		TableRowSelected: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Primary).
			Bold(true).
			Padding(0, 1),

		// 输入框样�?		Input: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Background(lipgloss.Color("#2C2C2E")).
			Padding(0, 1),
		InputFocused: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Background(colors.Primary).
			Padding(0, 1),
		InputError: lipgloss.NewStyle().
			Foreground(colors.Error).
			Background(lipgloss.Color("#2C2C2E")).
			Padding(0, 1),

		// 按钮样式
		Button: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Primary).
			Bold(true).
			Padding(0, 2),
		ButtonActive: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Secondary).
			Bold(true).
			Padding(0, 2),
		ButtonDisabled: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Background(lipgloss.Color("#2C2C2E")).
			Padding(0, 2),

		// 状态样�?		StatusDot: lipgloss.NewStyle().
			Padding(0, 1).
			Width(3),
		StatusText: lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Padding(0, 1),
	}
}

// ========== 辅助函数 ==========

// SuccessDot 成功状态点
func (s *StyleManager) SuccessDot() string {
	return s.StatusDot.Background(s.Colors.Success).Render("�?)
}

// ErrorDot 错误状态点
func (s *StyleManager) ErrorDot() string {
	return s.StatusDot.Background(s.Colors.Error).Render("�?)
}

// WarningDot 警告状态点
func (s *StyleManager) WarningDot() string {
	return s.StatusDot.Background(s.Colors.Warning).Render("�?)
}

// InfoDot 信息状态点
func (s *StyleManager) InfoDot() string {
	return s.StatusDot.Background(s.Colors.Info).Render("�?)
}

// MutedDot 禁用状态点
func (s *StyleManager) MutedDot() string {
	return s.StatusDot.Background(s.Colors.Dim).Render("�?)
}

// RenderCard 渲染卡片
func (s *StyleManager) RenderCard(title, content string) string {
	header := s.CardHeader.Render(title)
	body := s.Body.Render(content)
	return s.Card.Width(60).Render(header + "\n" + body)
}

// RenderStatus 渲染状态消�?func (s *StyleManager) RenderStatus(msg string, statusType StatusType) string {
	switch statusType {
	case StatusSuccess:
		return s.Success.Render("�?" + msg)
	case StatusError:
		return s.Error.Render("�?" + msg)
	case StatusWarning:
		return s.Warning.Render("�?" + msg)
	case StatusInfo:
		return s.Info.Render("�?" + msg)
	case StatusLoading:
		return s.Info.Render("�?" + msg)
	default:
		return s.Body.Render(msg)
	}
}
