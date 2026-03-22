# AgentFramework TUI

基于 **Memoh** 架构设计的终端用户界面（TUI），为 AgentFramework 提供强大的交互式操作体验。

![Version](https://img.shields.io/badge/version-2.1.0-blue.svg)
![License](https://img.shields.io/badge/license-AGPL--3.0--or--later-green.svg)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)

---

## ✨ 特性

- 🎨 **美观的界面** - 基于 lipgloss 的统一样式系统
- 🚀 **高性能** - 异步命令处理，批量数据加载
- 💾 **会话持久化** - 自动保存聊天历史
- ⚡ **流式输出** - 实时响应显示（框架支持）
- 🔧 **易于扩展** - 模块化架构，插件化组件
- 📊 **多视图管理** - 7个专业视图（Dashboard/Agents/Chat/Workflows/Skills/Settings/Logs）

---

## 🚀 快速开始

### 安装

```bash
go build -o af.exe
```

### 启动

```bash
./af.exe -tui
```

### 基本使用

```bash
# 列出 Agents
agent list

# 选择 Agent
agent select <id>

# 发送消息
chat 你好

# 查看工作流
workflow list

# 管理技能
skill list
skill enable <id>
```

---

## 📖 文档

- **[TUI_USER_GUIDE.md](docs/TUI_USER_GUIDE.md)** - 完整使用指南
- **[TUI_FINAL_SUMMARY.md](docs/TUI_FINAL_SUMMARY.md)** - 项目总结
- **[TUI_INTEGRATION_COMPLETE.md](docs/TUI_INTEGRATION_COMPLETE.md)** - 实现报告

---

## ⌨️ 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Tab` | 切换视图 |
| `Ctrl+R` | 刷新数据 |
| `Enter` | 执行命令 |
| `Q` | 退出 |

---

## 🏗️ 架构

```
┌─────────────────────────────────┐
│   Model (Bubble Tea)            │
│   - 事件处理                    │
│   - 视图渲染                    │
│   - 状态管理                    │
└─────────────────────────────────┘
              ↓
┌─────────────────────────────────┐
│   IntegrationLayer              │
│   - API 调用                    │
│   - 数据转换                    │
└─────────────────────────────────┘
              ↓
┌─────────────────────────────────┐
│   core.Application             │
│   - Host                        │
│   - SkillLibrary                │
│   - WorkflowManager             │
└─────────────────────────────────┘
```

---

## 📁 文件结构

```
cmd/tui/
├── main.go              # 包文档
├── messages.go          # 消息系统
├── styles.go            # 样式配置
├── config.go            # 配置管理
├── stream.go            # 流式处理
├── model.go             # 主模型
├── integration.go       # 集成层
├── views.go             # 视图渲染
├── session.go           # 会话持久化
└── run.go               # 入口函数
```

---

## 🎨 视图

1. **Dashboard** - 系统概览和统计
2. **Agents** - Agent 管理
3. **Chat** - 对话界面
4. **Workflows** - 工作流管理
5. **Skills** - 技能管理
6. **Settings** - 配置和帮助
7. **Logs** - 日志查看

---

## 💡 使用示例

### 对话流程

```bash
# 1. 启动 TUI
./af.exe -tui

# 2. 查看 Agents
agent list

# 3. 选择 Agent
agent select chat-agent-001

# 4. 开始对话
chat 你好
chat 帮我写一个函数
```

### 会话管理

```bash
# 创建新会话
session new

# 列出会话
session list

# 加载会话
session load <id>

# 导出会话
session export <id>
```

---

## 🔄 借鉴 Memoh

本项目借鉴了 [Memoh](https://github.com/memohai/Memoh) 的优秀架构设计：

- 模块化命令注册系统
- 统一的配置管理
- 流式响应处理
- 交互式用户体验

同时进行了创新改进：
- Go 实现（从 TypeScript）
- Bubble Tea 框架
- 增强的会话管理
- 更多的视图支持

---

## 📄 许可证

AGPL-3.0-or-later

---

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](../CONTRIBUTING.md)

---

**AgentFramework Team** © 2025
