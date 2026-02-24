# Agent Framework - 主程序使用指南

## 🎯 快速启动

### 方式一：直接运行可执行文件

```bash
# 桌面应用 (GUI)
./AgentFramework.exe

# 终端界面 (TUI)
./tui.exe

# 命令行工具 (CLI)
./agent-cli.exe --help
```

### 方式二：使用启动脚本

```bash
# Windows
run.bat              # 启动桌面应用
run.bat --tui        # 启动 TUI
run.bat cli --help   # CLI 帮助

# Linux/macOS
./run.sh             # 启动桌面应用
./run.sh --tui       # 启动 TUI
./run.sh cli --help  # CLI 帮助
```

---

## 🖥️ 模式一：桌面应用 (Desktop GUI)

基于 Wails v2 + Vue.js 的现代化桌面应用。

**启动命令:**
```bash
./AgentFramework.exe
```

**主要功能:**
- 📊 **Dashboard** - 实时监控和统计
- 💬 **Chat** - 多 Agent 对话界面
- 🔄 **Workflow** - 工作流可视化编辑
- 🧩 **Skills** - 技能管理面板
- ⚙️ **Settings** - 系统配置
- 📝 **Logs** - 日志查看器

**特点:**
- 跨平台支持 (Windows/Linux/macOS)
- 现代化 Web 技术栈
- 实时数据更新
- 响应式设计

---

## 📟 模式二：终端界面 (TUI)

基于 Bubble Tea 的终端用户界面。

**启动命令:**
```bash
./tui.exe
```

**界面导航:**
```
┌─────────────────────────────────────────┐
│  Agent Framework TUI                    │
├─────────────────────────────────────────┤
│ [菜单]              [主内容区]          │
│ • Dashboard                              │
│ • Agents                                 │
│ • Chat                                   │
│ • Workflows                              │
│ • Skills                                 │
│ • Settings                               │
│ • Logs                                   │
├─────────────────────────────────────────┤
│ [状态栏]                                  │
└─────────────────────────────────────────┘
```

**快捷键:**
| 按键 | 功能 |
|------|------|
| `Tab` | 切换视图 |
| `Enter` | 确认选择 |
| `Esc` | 返回上级 |
| `q` / `Ctrl+C` | 退出 |
| `↑` `↓` | 上下选择 |
| `Page Up/Down` | 快速翻页 |

**特点:**
- 纯终端操作
- 键盘快捷键
- 彩色界面
- 适合 SSH 远程使用

---

## 📟 模式三：命令行工具 (CLI)

强大的命令行工具，适合脚本和自动化。

**基本命令:**
```bash
# 查看帮助
./agent-cli.exe --help

# 列出所有 agents
./agent-cli.exe agent list

# 列出所有工作流
./agent-cli.exe workflow list

# 列出所有技能
./agent-cli.exe skill list

# 与 agent 对话
./agent-cli.exe chat default "你好"

# 执行工作流
./agent-cli.exe workflow exec <workflow-id> '{"input": "data"}'

# 查看日志
./agent-cli.exe logs tail

# 查看配置
./agent-cli.exe config get
```

**命令结构:**
```
agent-cli <command> [subcommand] [arguments]

可用命令:
  agent      管理 AI Agents
  workflow   管理工作流
  skill      管理技能
  chat       与 Agent 对话
  config     配置管理
  logs       查看日志
  completion Shell 自动完成
  help       显示帮助信息

全局选项:
  -c, --config string   配置文件路径
  -m, --model string   默认模型
  -o, --output string  输出格式 (json/text)
  -v, --verbose       详细输出
```

---

## 📊 可执行文件说明

| 文件 | 大小 | 功能 | 推荐场景 |
|------|------|------|----------|
| `AgentFramework.exe` | 79MB | 桌面应用 | 日常使用、完整功能 |
| `tui.exe` | 74MB | 终端界面 | SSH 远程、无图形环境 |
| `agent-cli.exe` | 1.5MB | 命令行工具 | 脚本自动化、批处理 |
| `server_demo.exe` | 72MB | HTTP 服务 | REST API 演示 |
| `simplebot.exe` | 6.9MB | 简单机器人 | 快速测试 |

---

## 🔧 构建

### 构建所有组件
```bash
# Windows
build.bat

# Linux/macOS
./build.sh
```

### 构建单个组件
```bash
# 主程序
go build -o AgentFramework.exe .

# TUI
go build -o tui.exe ./cmd/tui

# CLI
go build -o agent-cli.exe ./cmd/cli

# 服务器演示
go build -o server_demo.exe ./cmd/server_demo

# 简单机器人
go build -o simplebot.exe ./cmd/simplebot
```

---

## 📖 使用示例

### 示例 1: 快速对话 (桌面应用)
```bash
# 启动桌面应用
./AgentFramework.exe

# 在界面中:
# 1. 选择 Agent
# 2. 输入消息
# 3. 查看回复
```

### 示例 2: SSH 远程使用 (TUI)
```bash
# SSH 登录服务器
ssh user@server

# 启动 TUI
./tui.exe

# 使用键盘导航操作
```

### 示例 3: 自动化脚本 (CLI)
```bash
#!/bin/bash
# 自动对话脚本

AGENT_ID="default"
MESSAGE="分析以下数据: ${1}"

./agent-cli.exe chat "$AGENT_ID" "$MESSAGE"
```

### 示例 4: 批处理工作流
```bash
#!/bin/bash
# 批量执行工作流

for workflow in $(./agent-cli.exe workflow list --format json | jq -r '.[].id'); do
    echo "执行工作流: $workflow"
    ./agent-cli.exe workflow exec "$workflow" '{"input": "data"}'
done
```

---

## 🛠️ 故障排除

### 问题 1: 桌面应用无法启动
**解决方案:**
```bash
# 检查前端资源
ls frontend/dist/

# 重新构建前端
cd frontend
npm install
npm run build
```

### 问题 2: TUI 显示乱码
**解决方案:**
```bash
# Windows CMD
chcp 65001

# Windows PowerShell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

### 问题 3: CLI 找不到配置
**解决方案:**
```bash
# 创建配置文件
./agent-cli.exe config init

# 指定配置路径
./agent-cli.exe -c /path/to/config.json agent list
```

---

## 📚 更多文档

- [架构设计](./docs/ARCHITECTURE.md)
- [API 文档](./docs/API.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [部署指南](./docs/DEPLOYMENT.md)

---

**版本:** 1.2.0
**许可证:** AGPL-3.0-or-later
