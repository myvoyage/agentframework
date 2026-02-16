# 快速开始

> **5 分钟上手 AgentFramework**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [前置要求](#前置要求)
- [安装方式](#安装方式)
  - [方式一：从源码构建](#方式一从源码构建)
  - [方式二：使用预编译二进制](#方式二使用预编译二进制)
- [创建第一个应用](#创建第一个应用)
- [核心概念](#核心概念)
- [下一步](#下一步)

---

## 前置要求

### 环境要求

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| **Go** | 1.24 或更高 | 用于运行应用 |
| **Node.js** | 18+ 或 20+ | 用于构建桌面应用（可选） |
| **Git** | 任意版本 | 用于克隆代码库 |

### 可选组件

| 组件 | 用途 |
|------|------|
| **Docker** | 容器化代码执行 |
| **Redis** | 缓存和消息存储 |
| **Ollama** | 本地模型服务 |
| **Wails** | 桌面应用开发（可选） |

---

## 安装方式

### 方式一：从源码构建

#### 1. 克隆仓库

```bash
git clone https://github.com/myvoyage/agentframework.git
cd agentframework
```

#### 2. 下载依赖

```bash
go mod download
```

#### 3. 构建应用

```bash
# 构建 CLI 应用
go build -o bin/agentframework ./cmd/app

# 或者构建桌面应用（需要 Wails）
wails build
```

#### 4. 运行应用

```bash
# 运行 CLI 应用
./bin/agentframework

# 或运行桌面应用
./build/bin/AgentFramework.exe  # Windows
./build/bin/AgentFramework  # Linux/macOS
```

### 方式二：使用预编译二进制

#### 下载最新版本

访问 [Releases 页面](https://github.com/myvoyage/agentframework/releases) 下载适合您系统的二进制文件。

#### 解压并运行

```bash
# 解压
tar -xzf agentframework-v1.0.0-linux-amd64.tar.gz
cd agentframework

# 运行
./agentframework
```

---

## 创建第一个应用

### 步骤 1：创建项目目录

```bash
mkdir my-first-agent
cd my-first-agent
```

### 步骤 2：初始化 Go 模块

```bash
go mod init my-first-agent
```

### 步骤 3：创建主程序

创建 `main.go` 文件：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/schema"

    "agentframework/agent"
)

func main() {
    ctx := context.Background()

    // 步骤 1：创建模型工厂
    modelFactory := agent.NewDefaultModelFactory(agent.DefaultModelFactoryConfig{
        Models: map[string]agent.ModelConfig{
            "ollama-llama3": {
                Type:    "ollama",
                Model:   "llama3",
                BaseURL: "http://localhost:11434",
                Enabled: true,
            },
        },
    })

    // 步骤 2：创建 Host 配置
    cfg := &agent.HostConfig{
        Name:         "my-first-app",
        DefaultModel: "ollama-llama3",
        Models: map[string]agent.ModelConfig{
            "ollama-llama3": {
                Type:    "ollama",
                Model:   "llama3",
                BaseURL: "http://localhost:11434",
                Enabled: true,
            },
        },
    }

    // 步骤 3：创建 Host
    host, err := agent.NewHost(ctx, cfg, modelFactory, nil)
    if err != nil {
        log.Fatal("创建 Host 失败:", err)
    }

    // 步骤 4：获取或创建 Agent
    chatAgent, err := host.GetAgent("chat")
    if err != nil {
        // 创建简单的聊天代理
        chatAgent = agent.NewChatAgent(
            agent.WithName("chat"),
            agent.WithInstructions("你是一个有用的AI助手，请用中文回答问题"),
            agent.WithModelName("ollama-llama3"),
        )
    }

    // 步骤 5：运行 Agent
    response, err := chatAgent.(interface {
        Run(context.Context, string) (*schema.Message, error)
    }).Run(ctx, "你好，请介绍一下你自己")
    if err != nil {
        log.Fatal("运行 Agent 失败:", err)
    }

    // 步骤 6：输出结果
    fmt.Println("=== AI 回复 ===")
    fmt.Println(response.Content)
}
```

### 步骤 4：安装依赖

```bash
go get agentframework/agent
```

### 步骤 5：运行应用

```bash
go run main.go
```

### 预期输出

```
=== AI 回复 ===
你好！我是一个AI助手，可以帮你解答问题、提供建议和执行各种任务。请问我任何问题！
```

---

## 核心概念

### Host（主机）

Host 是整个框架的核心容器，负责管理所有组件的生命周期。

```go
host, err := agent.NewHost(ctx, cfg, modelFactory, toolRegistry)
```

**主要职责**:
- 🎯 管理 Agent 实例
- 🔄 管理 Workflow 实例
- 🛠️ 管理 Skill 库
- 📊 管理监控和日志

### Agent（代理）

Agent 是自主执行单元，使用 LLM 和工具完成任务。

**支持的类型**:
- **ChatAgent**: 基础对话代理
- **ReActAgent**: 推理-行动代理
- **WorkerAgent**: 专业工作代理
- **CollaborationAgent**: 协作代理

```go
agent := agent.NewChatAgent(
    agent.WithName("my-agent"),
    agent.WithInstructions("你是一个有用的AI助手"),
    agent.WithModelName("ollama-llama3"),
)
```

### ModelFactory（模型工厂）

ModelFactory 负责创建和管理 AI 模型实列。

**支持的模型**:
- ✅ OpenAI (GPT-4, GPT-3.5)
- ✅ Ollama (本地模型)
- ✅ LM Studio (本地模型工作室)

```go
factory := agent.NewDefaultModelFactory(config)
```

### Skill（技能）

Skill 是可插拔的功能单元，扩展 Agent 的能力。

**内置技能**:
- 🌐 HTTP 请求
- 📄 文件操作
- 💻 代码执行
- 🔄 数据处理

---

## 下一步

### 学习资源

- 📘 [架构概览](ARCHITECTURE_OVERVIEW.md) - 了解系统架构
- 📘 [配置指南](CONFIGURATION.md) - 详细配置说明
- 📘 [最佳实践](BEST_PRACTICES.md) - 开发建议
- 📘 [示例项目](../examples/) - 完整示例

### 进阶主题

- 🎯 [创建工作流](../guides/tutorials/CREATING_WORKFLOW.md)
- 🤝 [多代理协作](../guides/tutorials/CREATING_TEAM.md)
- 🔧 [自定义技能](../guides/tutorials/CUSTOM_SKILLS.md)
- 📊 [监控和调试](../guides/best-practices/MONITORING.md)

### 社区资源

- 💬 [GitHub Discussions](https://github.com/myvoyage/agentframework/discussions)
- 📖 [文档网站](https://docs.agentframework.dev)
- 🐛 [问题反馈](https://github.com/myvoyage/agentframework/issues)

---

**遇到问题？** 请查看 [故障排查指南](../operation/TROUBLESHOOTING.md) 或 [提交 Issue](https://github.com/myvoyage/agentframework/issues/new)
