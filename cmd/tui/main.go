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
// Agent Framework - TUI (Terminal User Interface)
// Copyright (C) 2025 Agent Framework Contributors
//
// 本包提供基于 Bubble Tea 的终端用户界面实�?//
// 架构概览:
//   - messages.go: 消息系统定义
//   - styles.go: 统一样式配置
//   - config.go: 配置和会话管�?//   - stream.go: 流式聊天处理
//   - model.go: �?TUI 模型
//   - run.go: TUI 入口函数
//
// 快速开�?
//
//	import cmdtui "AgentFramework/cmd/tui"
//
//	// 启动 TUI
//	err := cmdtui.Run(nil, nil)
//
// 使用方式:
//
//	AgentFramework -tui
//
package tui
