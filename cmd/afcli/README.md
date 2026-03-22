# AFCLI - AgentFramework CLI 独立程序

## 📖 概述

AFCLI 是 AgentFramework CLI 的独立可执行程序，提供命令行界面进行 Agent、工作流、技能和配置管理。

**文件信息**:
- 文件名: `afcli.exe`
- 版本: 2.1.0
- 基于框架: Cobra CLI

---

## 🚀 快速开始

### 1. 编译

```bash
# Windows
build_afcli.bat

# Linux/Mac
chmod +x build_afcli.sh
./build_afcli.sh
```

### 2. 运行

```bash
# 显示帮助
afcli --help

# 列出 agents
afcli agent list

# 选择 agent
afcli agent select <id>
```

---

## 📋 完整命令列表

### Agent 管理

```bash
# 列出所有 agents
afcli agent list

# 运行 agent
afcli agent run <agent-id> <task>

# 与 agent 对话（需要先选择）
afcli agent chat "message"
```

### 工作流管理

```bash
# 列出工作流
afcli workflow list

# 获取工作流详情
afcli workflow get <workflow-id>

# 创建工作流
afcli workflow create <name> [description]

# 执行工作流
afcli workflow execute <workflow-id> [input]

# 删除工作流
afcli workflow delete <workflow-id>

# 查看工作流版本
afcli workflow versions <workflow-id>
```

### 技能管理

```bash
# 列出所有技能
afcli skill list

# 获取技能详情
afcli skill info <skill-id>

# 启用技能
afcli skill enable <skill-id>

# 禁用技能
afcli skill disable <skill-id>

# 直接执行技能
afcli skill run <skill-id> [input]
```

### 配置管理

```bash
# 查看配置
afcli config get [key]

# 设置配置
afcli config set <key> <value>

# 验证配置
afcli config validate
```

### 文件操作

```bash
# 列出文件
afcli file list [path]

# 读取文件
afcli file read <path>

# 写入文件
afcli file write <path> <content>

# 复制文件
afcli file copy <src> <dst>

# 删除文件
afcli file delete <path>
```

### 工具命令

```bash
# 显示版本信息
afcli version

# 初始化配置
afcli init

# 生成自动补全脚本
afcli completion [bash|zsh|fish|powershell]
```

---

## 🎯 使用场景

### 场景1: 快速 Agent 列表

```bash
$ afcli agent list

Available Agents:
┌────────────────────────────────────────────────────────────┐
│ ID                │ Name                    │ Type         │
├────────────────────────────────────────────────────────────┤
│ chat-agent-001    │ Conversational Agent    │ ChatAgent   │
│ workflow-agent    │ Workflow Agent          │ ReActAgent  │
└────────────────────────────────────────────────────────────┘
```

### 场景2: 工作流执行

```bash
# 列出工作流
$ afcli workflow list

# 执行工作流
$ afcli workflow execute wf-001 "input data"

Workflow execution result:
Success - Task completed in 2.5s
```

### 场景3: 技能管理

```bash
# 列出技能
$ afcli skill list

Available Skills:
○ http-skill      v1.0.0  HTTP请求技能
○ file-skill      v1.0.0  文件操作技能
✓ code-skill      v1.2.0  代码执行技能 (已启用)

# 启用技能
$ afcli skill enable data-skill

Skill 'data-skill' enabled successfully
```

### 场景4: 配置管理

```bash
# 查看所有配置
$ afcli config get

Current Configuration:
────────────────────────────────────────────────────────────
Skill System Dir: .skills

Models:
  default: ollama/llama3
────────────────────────────────────────────────────────────

# 设置模型
$ afcli config set model.default qwen2.5

Model set to: qwen2.5
Configuration saved
```

---

## 🔧 高级功能

### 输出格式

支持多种输出格式：

```bash
# 表格格式（默认）
afcli agent list

# JSON 格式
afcli agent list -o json

# YAML 格式
afcli agent list -o yaml
```

### 详细输出

```bash
# 详细输出
afcli --verbose agent list

# 静默输出
afcli --quiet agent list
```

### 监视模式

```bash
# 持续监视（每5秒刷新）
afcli --watch workflow list
```

---

## 📁 文件和配置

### 默认配置位置

- Windows: `%USERPROFILE%\.agentframework\config.yaml`
- Linux/Mac: `~/.agentframework/config.yaml`

### 数据目录

- Windows: `%USERPROFILE%\.agentframework\`
- Linux/Mac: `~/.agentframework/`

### 会话文件

- 位置: `~/.agentframework/tui/sessions/`
- 格式: JSON

---

## 🆚 故障排查

### 问题1: 找不到命令

**症状**: `'afcli' 不是内部或外部命令`

**解决方案**:
```bash
# 方式1: 使用完整路径
.\afcli.exe agent list

# 方式2: 添加到 PATH
# 将 build 目录添加到系统 PATH 环境变量

# 方式3: 使用便捷脚本
build\afcli.bat agent list
```

### 问题2: Agent 未加载

**症状**: `agent list` 返回空列表

**解决方案**:
```bash
# 检查核心应用是否正确初始化
# 查看配置文件
afcli config get

# 验证模型配置
afcli config validate
```

### 问题3: 技能未加载

**症状**: `skill list` 返回空列表

**解决方案**:
```bash
# 检查技能目录
# 默认技能目录: .skills

# 确保技能文件存在
ls -la .skills/

# 手动指定技能目录
afcli config set skillDir /path/to/skills
```

---

## 💡 优化建议

### 性能优化

1. **调整超时时间**
```bash
afcli --timeout 60s workflow execute wf-001 "input"
```

2. **使用 JSON 输出解析**
```bash
afcli -o json agent list | jq '.[] | .name'
```

3. **批量操作**
```bash
for id in $(afcli -o json agent list | jq -r '.[].id'); do
    afcli agent run $id "test"
done
```

### 脚本集成

```bash
#!/bin/bash
# 自动化脚本示例

# 选择 agent
afcli agent select chat-agent-001

# 执行多个任务
afcli agent run chat-agent-001 "task1"
afcli agent run chat-agent-001 "task2"
afcli agent run chat-agent-001 "task3"

# 查看工作流
afcli workflow list
```

---

## 📊 与集成版本对比

| 特性 | 独立版 (afcli.exe) | 集成版 (af.exe) |
|-----|---------------------|---------------|
| 命令 | `afcli agent list` | `af agent list` |
| 大小 | ~30 MB | ~80+ MB |
| 依赖 | 无 | 完整安装 |
| 更新 | 替换文件即可 | 需要重新编译 |
| 适用场景 | 脚本、自动化 | 完整功能 |

---

## 📚 相关文档

- **[MULTIMODE_USAGE.md](docs/MULTIMODE_USAGE.md)** - 多模式使用指南
- **[cmd/cli/README.md](cmd/cli/README.md)** - CLI 包说明

---

## 🔄 更新日志

### v2.1.0 (2026-02-25)

**新增**:
- ✅ 独立可执行程序
- ✅ 完整的命令支持
- ✅ 多种输出格式
- ✅ 便捷编译脚本

---

**享受使用 AgentFramework CLI！** 🚀

---

**版本**: 2.1.0
**编译日期**: 2026-02-25
**许可证**: AGPL-3.0-or-later
