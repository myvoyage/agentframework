# AgentFramework CLI 使用指南

> **AgentFramework CLI** - 企业级 AI 代理框架命令行接口

## 📋 目录

- [快速开始](#快速开始)
- [安装](#安装)
- [基本命令](#基本命令)
- [工作流管理](#工作流管理)
- [技能系统](#技能系统)
- [文件操作](#文件操作)
- [配置管理](#配置管理)
- [Agent 交互](#agent-交互)
- [高级用法](#高级用法)

---

## 快速开始

### 安装

```bash
# 从源码构建
git clone https://github.com/myvoyage/agentframework.git
cd agentframework
make cli

# 或使用 Go 直接构建
go build -o af ./cmd/cli
```

### 基本使用

```bash
# 显示帮助信息
af --help

# 显示版本信息
af version

# 列出所有工作流
af workflow list

# 运行对话 Agent
af agent chat "你好，请介绍一下你自己"
```

---

## 安装

### 系统要求

| 组件 | 版本要求 |
|------|---------|
| **Go** | 1.24 或更高版本 |
| **OS** | Linux、macOS、Windows |

### 构建选项

```bash
# 仅构建 CLI 工具
make cli

# 构建桌面应用
make desktop

# 构建所有
make build
```

---

## 基本命令

### 全局选项

```bash
af [global options] <command> [arguments]

Global Options:
  -c, --config string   配置文件路径 (default "host.yaml")
  -m, --model string    指定模型名称
  -o, --output string   输出格式 (table/json) (default "table")
  -v, --verbose       详细输出
  -h, --help          显示帮助信息
      --version       显示版本信息
```

### 示例

```bash
# 使用自定义配置文件
af -c /path/to/config.yaml workflow list

# 指定输出格式为 JSON
af -o json workflow list

# 使用特定模型
af -m gpt-4 agent chat "你好"
```

---

## 工作流管理

### 列出工作流

```bash
af workflow list
```

输出示例：
```
Workflows:
────────────────────────────────────────────────────────────
ID: wf-001
  Name: 数据分析流程
  Description: 自动化数据处理和分析工作流
  Status: running
────────────────────────────────────────────────────────────
```

### 创建工作流

```bash
# 基本创建
af workflow create "我的工作流" "工作流描述"

# 使用 JSON 输出
af -o json workflow create "测试工作流" | jq .
```

### 执行工作流

```bash
# 简单执行
af workflow execute wf-001

# 带输入参数执行
af workflow execute wf-001 "{\"input\": \"data\"}"
```

### 删除工作流

```bash
af workflow delete wf-001
```

### 查看工作流详情

```bash
af workflow get wf-001
```

### 工作流版本管理

```bash
# 列出版本
af workflow versions wf-001

# 恢复到指定版本
af workflow restore wf-001 2
```

---

## 技能系统

### 列出技能

```bash
af skill list
```

输出示例：
```
Skills:
────────────────────────────────────────────────────────────
Name: http-request
  Description: HTTP 请求技能
  Version: 1.0.0
  Category: network
────────────────────────────────────────────────────────────
```

### 执行技能

```bash
# 简单执行
af skill execute http-request

# 带输入参数执行
af skill execute http-request "{\"url\": \"https://api.example.com\"}"
```

### 技能系统信息

```bash
af skill info
```

输出示例：
```
Skill System Information:
────────────────────────────────────────────────────────────
Initialized: true
Base Directory: .skills
Total Skills: 5
────────────────────────────────────────────────────────────
```

---

## 文件操作

### 列出文件

```bash
# 列出当前目录
af file ls

# 列出指定目录
af file ls /path/to/directory
```

### 读取文件

```bash
af file read /path/to/file.txt
```

### 写入文件

```bash
# 简单写入
af file write /path/to/file.txt "文件内容"

# 从标准输入读取
echo "内容" | af file write /path/to/file.txt
```

### 创建目录

```bash
af file mkdir /path/to/directory
```

### 删除文件

```bash
af file delete /path/to/file.txt
```

---

## 配置管理

### 查看配置

```bash
# 查看所有配置
af config get

# 查看特定配置值
af config get defaultModel
```

输出示例：
```
Configuration:
────────────────────────────────────────────────────────────
Default Model: llama3
Skill System Dir: .skills
Models: 2 configured
────────────────────────────────────────────────────────────
```

### JSON 格式输出

```bash
af -o json config get
```

---

## Agent 交互

### 对话模式

```bash
# 简单对话
af agent chat "你好，请介绍一下你自己"

# 使用指定模型
af -m gpt-4 agent chat "分析以下数据"
```

### 交互式对话

```bash
# 进入交互模式（未来功能）
af agent chat --interactive
```

---

## 高级用法

### Shell 自动补全

```bash
# Bash
source <(af completion bash)

# Zsh
af completion zsh > ~/.zshrc/_af

# Fish
af completion fish | source

# PowerShell
af completion powershell | Invoke-Expression
```

### 输出格式

所有命令都支持 `-o` 或 `--output` 选项来指定输出格式：

| 格式 | 说明 | 示例 |
|------|------|---------|
| **table** | 表格格式（默认） | `af -o table workflow list` |
| **json** | JSON 格式 | `af -o json workflow list \| jq .` |

### 配置文件

AgentFramework 使用 YAML 格式的配置文件（默认：`host.yaml`）：

```yaml
# host.yaml 示例
models:
  default:
    type: ollama
    model: llama3
    base_url: http://localhost:11434

  gpt-4:
    type: openai
    model: gpt-4-turbo
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}

skill_system_dir: .skills
```

---

## 最佳实践

### 1. 使用 Shell 别名

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
alias af='af -c ~/.config/agentframework/host.yaml'
alias afl='af workflow list'
alias afe='af workflow execute'
```

### 2. 脚本集成

```bash
#!/bin/bash
# 自动化数据处理脚本

# 检查工作流是否存在
if ! af workflow list | grep -q "数据处理"; then
  af workflow create "数据处理" "自动数据处理工作流"
fi

# 执行工作流
af workflow execute "数据处理" "{\"data\": \"$1\"}"
```

### 3. CI/CD 集成

```yaml
# GitHub Actions 示例
- name: Run Workflow
  run: |
    go build -o af ./cmd/cli
    af workflow execute test-workflow
```

---

## 故障排除

### 常见问题

**问题**: 配置文件未找到
```
Error: failed to load config: open host.yaml: no such file or directory
```
**解决**: 使用 `-c` 选项指定配置文件路径
```bash
af -c /path/to/config.yaml workflow list
```

---

**问题**: 模型未找到
```
Error: model not found: gpt-4
```
**解决**: 检查配置文件中的模型配置，或使用 `-m` 选项指定模型
```bash
af -m llama3 agent chat "测试"
```

---

## 更多资源

- 📘 [完整文档](../README_CN.md)
- 📘 [API 文档](../api/API.md)
- 🐛 [项目主页](https://agentframework.dev)
- 💬 [GitHub Issues](https://github.com/myvoyage/agentframework/issues)

---

**由 ❤️ 和 Go 语言构建**
