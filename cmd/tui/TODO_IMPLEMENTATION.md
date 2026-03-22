# TUI 重构 - 待实现业务逻辑

本文档标记了从 Memoh 借鉴架构后，需要用户根据具体业务逻辑实现的部分。

## 📋 概述

新的 TUI 架构已经建立了模块化的基础框架，包括：
- ✅ 消息系统（`messages.go`）
- ✅ 样式配置（`styles.go`）
- ✅ 配置和会话管理（`config.go`）
- ✅ 流式聊天基础（`stream.go`）
- ✅ 主模型（`model.go`）

以下是需要根据你的具体 Agent 实现来完成的业务逻辑部分。

---

## 🔧 需要实现的核心部分

### 1. Agent 调用逻辑（最重要）

**文件**: `cmd/tui/stream.go`

**位置**: `streamChat()` 函数（第 22-36 行）

**当前状态**: 占位符实现，仅返回模拟数据

**需要实现**:
```go
func streamChat(ctx context.Context, agentID, message, sessionID string) (string, error) {
    // TODO: 实现真实的 Agent 调用逻辑

    // 实现选项：
    //
    // 1. 直接调用 core.Application
    //    - 使用 app.GetHost().GetAgent(agentID)
    //    - 调用 agent.Run(ctx, message)
    //
    // 2. 通过 HTTP API
    //    - 发送 POST 请求到 /chat 端点
    //    - 处理流式响应
    //
    // 3. 直接集成 LLM SDK
    //    - OpenAI/Anthropic/本地模型
    //    - 处理流式响应

    return "", fmt.Errorf("not implemented")
}
```

**实现建议**:

```go
func streamChat(ctx context.Context, agentID, message, sessionID string) (string, error) {
    // 从 core.Application 获取 Agent
    // host := coreApp.GetHost()
    // agent, err := host.GetAgent(agentID)
    // if err != nil {
    //     return "", fmt.Errorf("failed to get agent: %w", err)
    // }
    //
    // response, err := agent.Run(ctx, message)
    // if err != nil {
    //     return "", fmt.Errorf("agent run failed: %w", err)
    // }
    //
    // return response.Content, nil
}
```

**权衡考虑**:
- **直接调用**: 性能更好，但需要访问 core.Application
- **HTTP API**: 更灵活，支持分布式部署，但增加网络开销
- **SDK 集成**: 更精细的控制，但增加代码复杂度

---

### 2. 数据加载命令

**文件**: `cmd/tui/stream.go`（第 156-181 行）

**需要实现**:
- `BatchLoadAgentsCmd` - 从 core.Application 加载 Agent 列表
- `BatchLoadWorkflowsCmd` - 从 core.Application 加载工作流列表
- `BatchLoadSkillsCmd` - 从 core.Application 加载技能列表

**参考实现**:

```go
func BatchLoadAgentsCmd(ctx context.Context, app *core.Application) tea.Cmd {
    return func() tea.Msg {
        agents := app.GetHost().ListAgents()

        items := make([]AgentItem, 0, len(agents))
        for _, agentID := range agents {
            agent, err := app.GetHost().GetAgent(agentID)
            if err == nil && agent != nil {
                items = append(items, AgentItem{
                    ID:   agentID,
                    Name: agent.Name(),
                    Type: "ChatAgent", // 从实际类型获取
                })
            }
        }

        return AgentListLoadedMsg{Agents: items}
    }
}
```

---

### 3. Agent 操作处理

**文件**: `cmd/tui/model.go`（第 370+ 行）

**需要实现**:
- `handleAgentCommand` - Agent 选择、启动、停止等操作
- `handleWorkflowCommand` - 工作流执行、停止等操作
- `handleSkillCommand` - 技能启用、禁用、执行等操作

**参考实现**:

```go
func (m *Model) handleAgentCommand(parts []string) (tea.Model, tea.Cmd) {
    if len(parts) < 2 {
        m.currentView = ViewAgents
        return m, nil
    }

    switch parts[1] {
    case "select":
        if len(parts) >= 3 {
            // 验证 Agent ID 是否有效
            _, err := m.core.GetHost().GetAgent(parts[2])
            if err != nil {
                m.setStatus(fmt.Sprintf("Agent not found: %s", parts[2]), StatusError)
                return m, nil
            }
            m.currentAgent = parts[2]
            m.setStatus(fmt.Sprintf("已选择 Agent: %s", parts[2]), StatusSuccess)
        }
    case "start":
        // 启动 Agent
        // TODO: 实现启动逻辑
    case "stop":
        // 停止 Agent
        // TODO: 实现停止逻辑
    }

    return m, nil
}
```

---

### 4. 视图渲染优化

**文件**: `cmd/tui/model.go`（第 450+ 行）

**需要优化**:
- `renderDashboard` - 添加统计信息、系统状态
- `renderAgents` - 使用 table 组件渲染 Agent 列表
- `renderChat` - 实现滚动消息历史、Markdown 渲染
- `renderWorkflows` - 添加工作流状态可视化
- `renderSkills` - 添加技能切换交互
- `renderLogs` - 集成日志流

**参考 Memoh**:
- 使用 `table` 库渲染表格数据
- 使用 `viewport` 实现可滚动内容
- 使用 `list` 实现可选择的列表

---

### 5. 会话持久化

**文件**: `cmd/tui/model.go`

**需要实现**:
- 聊天历史保存到会话
- 会话恢复功能
- 跨会话的上下文保持

**参考实现**:

```go
func (m *Model) handleChatResponse(msg ChatResponseMsg) (tea.Model, tea.Cmd) {
    // ... 现有代码 ...

    // 保存到会话
    if m.session != nil && m.config.Get().AutoSaveSession {
        sessionID := m.config.Get().SessionID
        err := m.session.Update(sessionID, func(s *Session) {
            s.Messages = m.chatMessages
        })
        if err != nil {
            m.setStatus("保存会话失败", StatusWarning)
        }
    }

    return m, nil
}
```

---

## 🎨 UI/UX 优化建议

基于 Memoh 的设计，以下是可以借鉴的 UX 优化：

1. **加载状态指示器**
   - 使用 `ora` 库（Go 对应: `spinner`）显示加载状态
   - 参考 Memoh 的 `spinner.start()` / `spinner.succeed()` 模式

2. **交互式选择器**
   - 使用 `inquirer` 风格的交互式选择
   - Go 对应: `bubbles/list` 或自定义实现

3. **表格渲染**
   - 使用 `table` 库渲染数据
   - Go 对应: `lipgloss` + 自定义表格

4. **流式输出**
   - 实现逐字或逐行的流式输出
   - 参考 Memoh 的 `streamChat()` 模式

5. **快捷键提示**
   - 在底部显示可用的快捷键
   - 根据当前视图动态更新

---

## 📝 实现优先级

### 高优先级（核心功能）
1. ✅ 基础架构完成
2. ⏳ **Agent 调用逻辑**（最重要）
3. ⏳ 数据加载命令

### 中优先级（用户体验）
4. ⏳ 视图渲染优化
5. ⏳ 会话持久化

### 低优先级（增强功能）
6. ⏳ UI/UX 优化
7. ⏳ 日志查看
8. ⏳ 高级交互

---

## 🚀 快速开始

1. **先实现 Agent 调用**
   - 编辑 `stream.go` 的 `streamChat()` 函数
   - 选择适合你的实现方式（直接调用 / HTTP API / SDK）
   - 测试基本对话功能

2. **实现数据加载**
   - 编辑 `stream.go` 的批量加载命令
   - 确保 Agent、工作流、技能能正确加载

3. **优化视图**
   - 美化各个视图的渲染
   - 添加交互功能

---

## 📚 参考资源

- **Memoh CLI**: [packages/cli/src/cli/index.ts](https://github.com/memohai/Memoh/blob/master/packages/cli/src/cli/index.ts)
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Lipgloss**: https://github.com/charmbracelet/lipgloss
- **Bubbles**: https://github.com/charmbracelet/bubbles

---

**创建日期**: 2026-02-25
**作者**: AgentFramework Team
**状态**: 待实现业务逻辑标记完成
