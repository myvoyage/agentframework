# AgentFramework 多模式启动指南

## 概述

AgentFramework 现在支持三种运行模式，可以根据使用场景灵活切换：

- **UI 模式 (默认)**: 基于 Wails 的桌面 GUI 应用
- **TUI 模式**: 基于 Bubble Tea 的终端用户界面
- **CLI 模式**: 命令行界面，适合脚本和自动化

## 启动方式

### 1. UI 模式 (桌面 GUI) - 默认

```bash
# 方式1: 无参数直接启动（默认）
AgentFramework

# 方式2: 明确指定 UI 模式
AgentFramework -ui
AgentFramework --ui
```

**特点:**
- 图形化界面，鼠标操作
- 可视化 Agent、工作流、技能管理
- 适合日常使用和演示

### 2. TUI 模式 (终端界面)

```bash
# 方式1: 使用 -tui 参数
AgentFramework -tui

# 方式2: 使用 --tui 参数
AgentFramework --tui
```

**特点:**
- 终端内运行的图形界面
- 键盘操作，高效便捷
- 支持多视图切换（Dashboard、Agents、Chat、Workflows、Skills、Settings、Logs）
- 无需离开终端即可使用

**TUI 操作说明:**
- `Tab` - 切换视图
- `Ctrl+R` - 刷新数据
- `Enter` - 选择/执行
- `Q` / `Ctrl+C` - 退出

### 3. CLI 模式 (命令行)

```bash
# 方式1: 使用 -cli 参数进入 CLI 模式
AgentFramework -cli

# 方式2: 直接使用子命令（自动进入 CLI 模式）
AgentFramework agent list
AgentFramework workflow list
AgentFramework skill list
```

**CLI 命令结构:**

```bash
# Agent 管理
AgentFramework agent list                    # 列出所有 agents
AgentFramework agent chat "message"          # 与 agent 对话
AgentFramework agent run <agent-id> <task>   # 运行指定 agent

# 工作流管理
AgentFramework workflow list                 # 列出所有工作流
AgentFramework workflow get <id>             # 获取工作流详情
AgentFramework workflow create <name>        # 创建工作流
AgentFramework workflow execute <id> <input> # 执行工作流
AgentFramework workflow delete <id>          # 删除工作流

# 技能管理
AgentFramework skill list                    # 列出所有技能
AgentFramework skill info <id>               # 获取技能详情
AgentFramework skill enable <id>             # 启用技能
AgentFramework skill disable <id>            # 禁用技能
AgentFramework skill run <id> <input>        # 直接执行技能

# 配置管理
AgentFramework config get [key]              # 获取配置
AgentFramework config set <key> <value>      # 设置配置
AgentFramework config validate               # 验证配置

# 文件操作
AgentFramework file list [path]              # 列出文件
AgentFramework file read <path>              # 读取文件
AgentFramework file write <path> <content>   # 写入文件
AgentFramework file copy <src> <dst>         # 复制文件
AgentFramework file delete <path>            # 删除文件

# 其他命令
AgentFramework version                       # 显示版本信息
AgentFramework init                          # 初始化配置
AgentFramework completion <shell>            # 生成自动补全脚本
```

## 使用场景建议

### 场景1: 日常开发和测试
**推荐: TUI 模式**
```bash
AgentFramework -tui
```
- 快速切换不同功能
- 实时查看 Agent 响应
- 方便的日志查看

### 场景2: 演示和培训
**推荐: UI 模式**
```bash
AgentFramework
```
- 图形界面更直观
- 适合向非技术人员展示
- 支持拖拽等高级交互

### 场景3: 脚本和自动化
**推荐: CLI 模式**
```bash
AgentFramework agent run my-agent "process data"
AgentFramework workflow execute wf-001 "input data"
```
- 易于集成到脚本
- 支持管道和重定向
- 适合 CI/CD 流程

### 场景4: 远程服务器使用
**推荐: TUI 模式或 CLI 模式**
```bash
# TUI - 交互式操作
ssh server "AgentFramework -tui"

# CLI - 快速命令
ssh server "AgentFramework agent list"
```
- 无需图形界面
- 低带宽消耗
- SSH 友好

## 全局选项

所有模式都支持以下全局选项：

```bash
-c, --config <file>    # 指定配置文件
-m, --model <name>     # 指定模型
-o, --output <format>  # 输出格式 (table/json/yaml)
-v, --verbose          # 详细输出
--timeout <duration>   # 操作超时时间
--watch                # 监视模式
```

## 配置文件

### 默认配置位置
- Linux/Mac: `~/.agentframework/config.yaml`
- Windows: `%USERPROFILE%\.agentframework\config.yaml`

### 环境变量
```bash
export AF_MODEL="ollama/llama3"           # 默认模型
export AGENT_FRAMEWORK_DATA_DIR="/data"   # 数据目录
```

## 示例工作流

### 1. 初始化并启动 TUI
```bash
AgentFramework init
AgentFramework -tui
```

### 2. 查看 Agent 并对话
```bash
# CLI 模式
AgentFramework agent list
AgentFramework agent chat "你好，请介绍一下你自己"

# TUI 模式
AgentFramework -tui
# 在 TUI 中切换到 Agents 视图选择 Agent
# 然后切换到 Chat 视图进行对话
```

### 3. 创建并执行工作流
```bash
# CLI 模式
AgentFramework workflow create "my-workflow" "我的第一个工作流"
AgentFramework workflow execute <workflow-id> "输入数据"

# TUI 模式
AgentFramework -tui
# 在 Workflows 视图中创建和执行
```

### 4. 管理技能
```bash
# CLI 模式
AgentFramework skill list
AgentFramework skill enable http-skill
AgentFramework skill run http-skill '{"url": "https://api.example.com"}'

# TUI 模式
AgentFramework -tui
# 在 Skills 视图中管理技能
```

## 故障排查

### 问题1: 启动失败
```bash
# 查看详细错误
AgentFramework -cli --verbose
```

### 问题2: 模型不可用
```bash
# 检查配置
AgentFramework config get model

# 设置模型
AgentFramework config set model.default llama3
```

### 问题3: 技能未加载
```bash
# 列出技能
AgentFramework skill list

# 启用技能
AgentFramework skill enable <skill-id>
```

## 高级用法

### 1. 结合管道
```bash
echo "input data" | AgentFramework agent run my-agent -
```

### 2. 后台运行
```bash
# TUI 模式不适合后台运行
# 使用 CLI 模式代替
nohup AgentFramework agent run my-agent "task" > output.log 2>&1 &
```

### 3. 并行执行
```bash
AgentFramework workflow execute wf1 "input" &
AgentFramework workflow execute wf2 "input" &
wait
```

## 更多帮助

```bash
# CLI 帮助
AgentFramework --help
AgentFramework agent --help
AgentFramework workflow --help

# TUI 帮助
AgentFramework -tui
# 按 ? 查看 TUI 内置帮助
```

---

**文档版本**: 1.0
**更新日期**: 2026-02-25
**作者**: AgentFramework Team
