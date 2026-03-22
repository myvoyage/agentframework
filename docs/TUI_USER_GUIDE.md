# AgentFramework TUI 完整使用指南

## 📖 概述

AgentFramework TUI 是一个功能丰富的终端用户界面，支持 AI Agents、工作流和技能的管理。

基于 **Memoh** 的架构设计，提供：
- ✅ 流畅的键盘操作
- ✅ 实时数据刷新
- ✅ 会话持久化
- ✅ 多视图管理

---

## 🚀 快速开始

### 安装与编译

```bash
# 克隆仓库
git clone https://github.com/your-org/AgentFramework.git
cd AgentFramework

# 编译
go build -o af.exe

# 或使用测试脚本
test_tui.bat  # Windows
./test_tui.sh # Linux/Mac
```

### 启动 TUI

```bash
# 方式1: 使用主程序
./af.exe -tui

# 方式2: 直接启动
./af.exe --tui

# 方式3: 使用参数
./af.exe -tui
```

---

## ⌨️ 键盘快捷键

### 全局快捷键

| 快捷键 | 功能 |
|--------|------|
| `Tab` | 切换到下一个视图 |
| `Shift+Tab` | 切换到上一个视图 |
| `Ctrl+R` | 刷新当前数据 |
| `Enter` | 执行命令 |
| `Q` | 退出 TUI |
| `Ctrl+C` | 退出 TUI |

### 视图导航

```
Dashboard ←→ Agents ←→ Chat ←→ Workflows ←→ Skills ←→ Settings ←→ Logs
   ↑                                                                  ↓
   └──────────────────────────────────────────────────────────────────┘
```

---

## 💬 命令系统

### Agent 操作

```bash
# 列出所有 Agents
agent list

# 选择 Agent（用于聊天）
agent select <agent-id>

# 示例
agent select chat-agent-001
```

### 聊天功能

```bash
# 发送消息（需先选择 Agent）
chat <message>

# 示例
chat 你好，请介绍一下你自己
chat 帮我分析这段代码
chat 今天天气怎么样
```

### 工作流管理

```bash
# 列出工作流
workflow list
wf list  # 简写

# 执行工作流
workflow execute <workflow-id> <input>

# 示例
workflow execute data-process "input data"
```

### 技能管理

```bash
# 列出技能
skill list

# 启用技能
skill enable <skill-id>

# 禁用技能
skill disable <skill-id>

# 示例
skill enable http-skill
skill disable file-skill
```

### 会话管理

```bash
# 创建新会话
session new
session new <agent-id>

# 加载会话
session load <session-id>

# 列出会话
session list

# 删除会话
session delete <session-id>

# 导出会话
session export <session-id>

# 示例
session new chat-agent-001
session load session-1234567890
session export session-1234567890
```

---

## 🖥️ 视图详解

### 1. Dashboard（仪表板）

显示系统概览和统计信息：
- Agents 数量
- 工作流数量
- 技能数量
- 聊天消息数
- 当前状态
- 快速操作提示

### 2. Agents（Agent 管理）

列出所有可用的 Agents：
- Agent 名称和类型
- 当前选中的 Agent
- Agent ID

### 3. Chat（对话界面）

实时对话功能：
- 显示当前 Agent
- 消息历史记录
- 时间戳显示
- 流式输出指示器

### 4. Workflows（工作流管理）

管理工作流：
- 工作流列表
- 状态指示（Ready/Running/等）
- 工作流 ID 和描述

### 5. Skills（技能管理）

管理技能：
- 技能列表和版本
- 启用/禁用状态
- 技能描述

### 6. Settings（设置）

显示配置和帮助：
- 快捷键说明
- 命令参考
- 当前配置
- 系统信息

### 7. Logs（日志）

系统日志查看：
- 系统消息
- 错误提示
- 事件记录

---

## 💡 使用场景

### 场景1: 快速对话

```bash
# 1. 启动 TUI
./af.exe -tui

# 2. 列出 Agents
agent list

# 3. 选择 Agent
agent select chat-agent-001

# 4. 开始对话
chat 你好
chat 帮我写一个 Python 函数
```

### 场景2: 工作流执行

```bash
# 1. 切换到工作流视图（按 Tab）
# 2. 查看工作流
workflow list

# 3. 执行工作流
workflow execute wf-data-process "input data"

# 4. 查看结果
# （结果会显示在状态栏）
```

### 场景3: 会话管理

```bash
# 1. 创建新会话
session new chat-agent-001

# 2. 进行对话
chat 任务1
chat 任务2

# 3. 切换到其他 Agent
agent select other-agent

# 4. 创建新会话
session new other-agent

# 5. 加载之前的会话
session load <previous-session-id>
```

### 场景4: 技能管理

```bash
# 1. 查看所有技能
skill list

# 2. 禁用某个技能
skill disable old-skill

# 3. 启用新技能
skill enable new-skill

# 4. 刷新查看状态
Ctrl+R
```

---

## 📁 文件和配置

### 配置文件

**位置**: `~/.agentframework/tui/config.json`

```json
{
  "theme": "default",
  "showLineNumbers": true,
  "streamChat": true,
  "autoScroll": true,
  "maxHistory": 100,
  "sessionId": "tui-1234567890",
  "lastAgentId": "chat-agent-001",
  "autoSaveSession": true,
  "refreshInterval": 5000,
  "enableCache": true
}
```

### 会话文件

**位置**: `~/.agentframework/tui/sessions/`

每个会话一个 JSON 文件：
```
session-1234567890.json
session-1234567891.json
...
```

### 日志文件

**位置**: `~/.agentframework/tui/logs/`

---

## 🎨 自定义配置

### 修改默认 Agent

编辑配置文件：
```json
{
  "defaultAgentId": "your-preferred-agent"
}
```

### 调整流式聊天

```json
{
  "streamChat": true,
  "autoScroll": true
}
```

### 启用/禁用自动保存

```json
{
  "autoSaveSession": true
}
```

---

## 🔧 高级功能

### 1. 自动会话保存

每次对话后自动保存会话：
- 用户消息
- Agent 响应
- 时间戳
- 元数据

### 2. 会话导出

导出会话为文本格式：
```bash
session export <session-id>
```

### 3. 流式输出

支持实时流式输出：
- 逐字显示 Agent 响应
- 打字机效果
- 可中断

### 4. 状态监控

实时状态更新：
- 加载状态
- 错误提示
- 成功确认

---

## 🐛 故障排查

### 问题1: 无法看到 Agents

**症状**: Agents 视图为空

**解决方案**:
1. 按 `Ctrl+R` 刷新数据
2. 检查核心应用是否正确初始化
3. 查看错误日志

### 问题2: 聊天无响应

**症状**: 发送消息后没有回复

**解决方案**:
1. 确保已选择 Agent（`agent select <id>`）
2. 检查 Agent 状态
3. 查看 Logs 视图的错误信息

### 问题3: 会话未保存

**症状**: 关闭后数据丢失

**解决方案**:
1. 检查配置中 `autoSaveSession` 是否为 `true`
2. 确认会话目录可写
3. 手动保存：`session export <id>`

### 问题4: 键盘无响应

**症状**: 按键无效

**解决方案**:
1. 点击终端窗口确保获得焦点
2. 检查是否有其他程序占用键盘
3. 尝试 `Ctrl+C` 退出后重新启动

---

## 📊 性能优化

### 内存管理

- 默认保存最近 100 条消息
- 可通过 `maxHistory` 配置调整

### 刷新频率

- 默认 5 秒自动刷新
- 可通过 `refreshInterval` 调整

### 缓存控制

- 启用缓存提升性能
- 可通过 `enableCache` 控制

---

## 🆕 更新日志

### v2.1.0 (2026-02-25)

**新增功能**:
- ✅ 会话持久化
- ✅ 自动保存
- ✅ 会话导出
- ✅ 增强的视图渲染
- ✅ 集成层架构

**优化**:
- 🎨 更好的视觉效果
- ⚡ 更快的响应速度
- 🔧 更好的错误处理

**架构**:
- 📦 模块化设计
- 🔌 插件化组件
- 📚 完整文档

---

## 📚 相关文档

- **架构设计**: [TUI_REFACTORING_REPORT.md](TUI_REFACTORING_REPORT.md)
- **集成实现**: [TUI_INTEGRATION_COMPLETE.md](TUI_INTEGRATION_COMPLETE.md)
- **Memoh 项目**: https://github.com/memohai/Memoh

---

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](../CONTRIBUTING.md)

---

## 📄 许可证

AGPL-3.0-or-later

---

**文档版本**: 1.0
**更新日期**: 2026-02-25
**作者**: AgentFramework Team
