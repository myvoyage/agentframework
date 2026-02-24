# Agent Framework - 使用指南

Agent Framework 提供三种用户界面模式，满足不同的使用场景。

## 🖥️ 界面模式

### 1. 桌面应用 (Desktop GUI)

基于 Wails v2 的桌面应用，提供完整的图形用户界面。

**启动方式:**
```bash
# 方式 1: 直接运行
./AgentFramework.exe        # Windows
./AgentFramework            # Linux/macOS

# 方式 2: 使用启动脚本
./run.sh                    # Linux/macOS
./run.bat                   # Windows
```

**功能特性:**
- 🎨 现代化 Web 界面
- 📊 实时监控仪表板
- 💬 聊天交互界面
- 🔄 工作流可视化编辑器
- ⚙️ 配置管理面板
- 📝 日志查看器

---

### 2. 终端界面 (TUI)

基于 Bubble Tea 的终端用户界面，提供丰富的命令行交互体验。

**启动方式:**
```bash
# 方式 1: 直接运行
./tui.exe                   # Windows
./tui                       # Linux/macOS

# 方式 2: 使用启动脚本
./run.sh --tui              # Linux/macOS
./run.bat --tui             # Windows
```

**功能特性:**
- 📟 全终端操作体验
- 🎯 键盘快捷键导航
- 📋 多视图切换 (Dashboard/Agents/Chat/Workflows/Skills/Settings/Logs)
- 🎨 彩色界面输出
- 📊 实时状态更新

**快捷键:**
- `Ctrl+C` 或 `q` - 退出
- `Tab` - 切换视图
- `Enter` - 选择/确认
- `Esc` - 返回上级

---

### 3. 命令行界面 (CLI)

强大的命令行工具，适合脚本自动化和批处理。

**启动方式:**
```bash
# 方式 1: 直接运行
./agent-cli.exe [command]   # Windows
./agent-cli [command]        # Linux/macOS

# 方式 2: 通过启动脚本
./run.sh cli [command]       # Linux/macOS
./run.bat cli [command]      # Windows
```

**常用命令:**

```bash
# 查看帮助
./agent-cli --help

# 列出所有 agents
./agent-cli agent list

# 与 agent 对话
./agent-cli chat default "你好，请介绍一下你自己"

# 列出工作流
./agent-cli workflow list

# 列出技能
./agent-cli skill list

# 执行工作流
./agent-cli workflow exec <workflow-id> '{"input": "data"}'

# 查看配置
./agent-cli config get

# 查看日志
./agent-cli logs tail
```

---

## 🔨 构建指南

### 快速构建（所有组件）

```bash
# Linux/macOS
./build.sh

# Windows
build.bat
```

### 选择性构建

```bash
# 只构建主程序
./build.sh --main

# 只构建 CLI
./build.sh --cli

# 只构建 TUI
./build.sh --tui

# 只构建服务器演示
./build.sh --server

# 只构建简单机器人
./build.sh --simplebot

# 清理构建产物
./build.sh --clean
```

### 手动构建

```bash
# 主程序 (桌面应用)
go build -o AgentFramework.exe .

# CLI 工具
go build -o agent-cli.exe ./cmd/cli

# TUI
go build -o tui.exe ./cmd/tui

# 服务器演示
go build -o server_demo.exe ./cmd/server_demo

# 简单机器人
go build -o simplebot.exe ./cmd/simplebot
```

---

## 📁 项目结构

```
AgentFramework/
├── main.go                 # 主入口（支持 UI/CLI/TUI 模式）
├── app.go                  # 桌面应用实现
├── run.sh / run.bat        # 启动脚本
├── build.sh / build.bat    # 构建脚本
│
├── cmd/
│   ├── cli/               # CLI 工具
│   ├── tui/               # TUI (Terminal UI)
│   ├── simplebot/         # 简单机器人示例
│   └── server_demo/       # HTTP 服务器演示
│
├── frontend/              # 桌面应用前端 (Vue.js)
│   ├── src/
│   └── dist/              # 构建产物
│
├── agent/                # Agent 核心接口
├── api/                  # REST API 服务
├── core/                 # 应用核心
└── pkg/                  # 功能包
    ├── beads/           # MCP 和硬件驱动
    ├── cache/           # 缓存系统
    ├── channels/        # 多渠道消息
    ├── errors/          # 错误处理
    ├── framework/       # 框架层
    ├── iot/            # IoT 设备支持
    └── ...
```

---

## 🚀 快速开始

### 1. 首次构建

```bash
# 克隆仓库
git clone https://github.com/your-repo/AgentFramework.git
cd AgentFramework

# 构建所有组件
./build.sh
```

### 2. 启动应用

**桌面应用（推荐新用户）:**
```bash
./run.sh
```

**终端界面:**
```bash
./run.sh --tui
```

**命令行界面:**
```bash
./run.sh cli agent list
```

---

## 📖 更多文档

- [架构设计文档](./docs/ARCHITECTURE.md)
- [API 文档](./docs/API.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [插件开发](./docs/PLUGIN_DEVELOPMENT.md)
- [部署指南](./docs/DEPLOYMENT.md)

---

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

---

**许可证:** AGPL-3.0-or-later
