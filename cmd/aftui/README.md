# AFTUI - AgentFramework TUI 独立程序

## 📖 概述

AFTUI 是 AgentFramework TUI 的独立可执行程序，无需依赖其他组件即可运行。

**文件信息**:
- 文件名: `aftui.exe`
- 大小: ~74 MB
- 版本: 2.1.0
- 基于: Memoh 架构

---

## 🚀 快速开始

### 1. 编译

```bash
# Windows
build_aftui.bat

# Linux/Mac
go build -o aftui ./cmd/aftui/
```

### 2. 运行

```bash
# Windows
build\aftui.exe

# 或从其他目录
.\build\aftui.exe

# 使用便捷脚本
build\tui.bat
```

### 3. 交互

启动后会看到欢迎界面：

```
╔════════════════════════════════════════════════════════════╗
║         AgentFramework TUI - 终端用户界面                 ║
║                   Version 2.1.0                            ║
║                                                            ║
║  基于 Memoh 架构设计                                        ║
║  支持 Agents、工作流、技能管理                                ║
║                                                            ║
║  快捷键: Tab=切换视图  Ctrl+R=刷新  Q=退出                     ║
╚════════════════════════════════════════════════════════════╝

⏳ 正在初始化...
```

---

## ⌨️ 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Tab` | 切换到下一个视图 |
| `Shift+Tab` | 切换到上一个视图 |
| `Ctrl+R` | 刷新当前数据 |
| `Enter` | 执行命令 |
| `Q` | 退出 TUI |

---

## 💬 命令参考

### Agent 管理

```bash
agent list              # 列出所有 Agents
agent select <id>        # 选择 Agent
```

### 聊天

```bash
chat <message>          # 发送消息（需先选择 Agent）
```

### 工作流

```bash
workflow list           # 列出工作流
workflow execute <id>    # 执行工作流
```

### 技能

```bash
skill list              # 列出技能
skill enable <id>        # 启用技能
skill disable <id>       # 禁用技能
```

### 会话

```bash
session new             # 创建新会话
session list            # 列出会话
session load <id>        # 加载会话
session export <id>      # 导出会话
```

---

## 🖥️ 视图说明

### 1. Dashboard（仪表板）

- 系统统计信息
- 当前状态
- 快速操作指南

### 2. Agents（Agent 管理）

- 列出所有可用的 Agents
- 显示当前选中的 Agent
- Agent ID 和类型信息

### 3. Chat（对话界面）

- 显示当前 Agent
- 聊天历史记录
- 时间戳显示
- 流式输出指示

### 4. Workflows（工作流）

- 工作流列表
- 状态指示
- 执行记录

### 5. Skills（技能）

- 技能列表
- 启用/禁用状态
- 版本信息

### 6. Settings（设置）

- 快捷键说明
- 命令参考
- 当前配置
- 系统信息

### 7. Logs（日志）

- 系统消息
- 错误提示
- 事件记录

---

## 📁 文件位置

### Windows

```
可执行文件: build\aftui.exe
配置文件:   %USERPROFILE%\.agentframework\tui\config.json
会话文件:   %USERPROFILE%\.agentframework\tui\sessions\
```

### Linux/Mac

```
可执行文件: ./aftui
配置文件:   ~/.agentframework/tui/config.json
会话文件:   ~/.agentframework/tui/sessions/
```

---

## 🔧 配置

默认配置：

```json
{
  "theme": "default",
  "streamChat": true,
  "autoScroll": true,
  "maxHistory": 100,
  "autoSaveSession": true,
  "refreshInterval": 5000
}
```

---

## 🎯 使用场景

### 场景1: 快速对话

```bash
# 1. 启动 TUI
aftui

# 2. 列出 Agents
agent list

# 3. 选择 Agent
agent select chat-agent-001

# 4. 发送消息
chat 你好，请介绍一下你自己
```

### 场景2: 工作流管理

```bash
# 1. 启动 TUI
aftui

# 2. 按 Tab 切换到 Workflows 视图

# 3. 查看工作流
workflow list

# 4. 执行工作流
workflow execute wf-001 "input data"
```

### 场景3: 会话管理

```bash
# 1. 启动 TUI
aftui

# 2. 创建新会话
session new chat-agent-001

# 3. 进行对话...
chat 任务1
chat 任务2

# 4. 保存会话（自动）

# 5. 导出会话
session export session-id
```

---

## 🆚 故障排查

### 问题1: 启动时黑屏

**解决方案**:
- 确保终端支持 ANSI 颜色
- Windows 10+ 使用 Windows Terminal 或 PowerShell
- 启用虚拟终端模式

### 问题2: 按键无响应

**解决方案**:
- 点击终端窗口确保获得焦点
- 检查是否有其他程序占用键盘
- 尝试 `Ctrl+C` 后重新启动

### 问题3: 无数据显示

**解决方案**:
- 按 `Ctrl+R` 刷新数据
- 检查配置文件是否正确
- 查看 Logs 视图的错误信息

---

## 📊 性能

### 资源占用

- 内存: ~50-100 MB
- CPU: < 5% (空闲时)
- 启动时间: < 2 秒

### 优化建议

- 调整 `maxHistory` 减少内存占用
- 调整 `refreshInterval` 降低 CPU 使用
- 启用 `enableCache` 提升性能

---

## 📚 相关文档

- **[TUI_USER_GUIDE.md](docs/TUI_USER_GUIDE.md)** - 完整使用指南
- **[TUI_FINAL_SUMMARY.md](docs/TUI_FINAL_SUMMARY.md)** - 项目总结
- **[TUI_INTEGRATION_COMPLETE.md](docs/TUI_INTEGRATION_COMPLETE.md)** - 实现报告

---

## 🔄 更新日志

### v2.1.0 (2026-02-25)

**新增**:
- ✅ 独立可执行程序
- ✅ 欢迎界面
- ✅ 便捷编译脚本

**功能**:
- ✅ 7 个视图
- ✅ 会话持久化
- ✅ 流式聊天框架
- ✅ 完整命令系统

---

## 💡 提示

1. **首次使用**建议先查看 Settings 视图了解快捷键
2. **聊天前**务必先选择 Agent
3. **会话**会自动保存，无需手动操作
4. **退出**前建议查看是否有未保存的数据

---

**享受使用 AgentFramework TUI！** 🎉

---

**版本**: 2.1.0
**编译日期**: 2026-02-25
**许可证**: AGPL-3.0-or-later
