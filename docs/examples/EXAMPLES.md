# AgentFramework 使用示例

> **AgentFramework 实际应用场景和代码示例**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [快速开始](#快速开始)
- [客服系统](#客服系统)
- [数据分析平台](#数据分析平台)
- [研究助手](#研究助手)
- [内容生成平台](#内容生成平台)
- [自动化测试](#自动化测试)

---

## 快速开始

### 环境准备

**前置要求**:
- Go 1.24+
- Node.js (前端构建)
- Wails v2.11.0 (桌面应用)
- Ollama 或其他 LLM 后端

### 1. 创建配置文件

**文件**: `config/host.yaml`

```yaml
name: "my-agent-app"
version: "1.0.0"

# 默认模型
default_model: "ollama-llama3"

# 模型配置
models:
  # OpenAI GPT-4
  openai:
    type: "openai"
    model: "gpt-4-turbo"
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    enabled: true
    timeout: 30
    temperature: 0.7

  # Ollama Llama3
  ollama:
    type: "ollama"
    model: "llama3"
    base_url: "http://localhost:11434"
    enabled: true
    timeout: 120

# 日志配置
logging:
  level: "info"
  format: "json"

# Agent 配置
agents:
  # 聊天代理
  chat:
    name: "customer_service"
    type: "chat"
    model: "ollama-llama3"
    instructions: "你是一个智能客服助手，帮助用户解决问题。"
    tools:
      - "file_operation"
      - "web_search"
```

### 2. 创建主程序

**文件**: `cmd/app/main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"

    "agentframework/host"
    "agentframework/config"
    "agentframework/model"
)

func main() {
    ctx := context.Background()

    // 加载配置
    cfg, err := config.LoadHostConfig("config/host.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 创建模型工厂
    mf := model.NewDefaultModelFactory(&config.DefaultModelFactoryConfig{
        Models: cfg.Models,
    })

    // 创建 Host
    host, err := host.NewHost(ctx, cfg, mf, nil)
    if err != nil {
        log.Fatal(err)
    }

    // 获取 Agent
    agent, err := host.GetAgent("customer_service")
    if err != nil {
        log.Fatal(err)
    }

    // 执行任务
    response, err := agent.Run(ctx, "帮我查询用户订单状态")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
}
```

### 3. 运行应用

```bash
# 构建应用
wails build

# 运行应用
./build/bin/AgentFramework.app
```

---

## 客服系统

### 场景描述

智能客服系统，提供 24/7 自动响应和人工审核功能。

### 配置

**文件**: `config/agents/customer_service.yaml`

```yaml
name: "customer_service"
type: "chat"
model: "ollama-llama3"
instructions: |
  你是一个智能客服助手，帮助用户解决问题。

  你的工作流程：
  1. 理解用户问题
  2. 查询相关数据
  3. 提供解决方案
  4. 如果无法解决，转人工客服

hitl:
  enabled: true
  approval_mode: "manual"
  timeout: 3600

tools:
  - file_operation
  - web_search
  - knowledge_base
```

### 代码实现

**文件**: `agents/customer_service.go`

```go
package agents

import (
    "context"
    "agentframework/agent"
    "agentframework/skills"
)

type CustomerServiceAgent struct {
    *agent.BaseAgent
}

func NewCustomerServiceAgent(cfg *CustomerServiceConfig) (*CustomerServiceAgent, error) {
    agent := &CustomerServiceAgent{
        Name: cfg.Name,
        Type: "chat",
        Instructions: cfg.Instructions,
    HITL: &agent.HITLConfig{
            Enabled:        cfg.HITL.Enabled,
            ApprovalMode: cfg.HITL.ApprovalMode,
            Timeout:      cfg.HITL.Timeout,
        },
    Tools:        cfg.Tools,
    }

    return &agent, nil
}
```

---

## 数据分析平台

### 场景描述

使用 ReActAgent 和 Workflow 构建复杂数据分析流水线。

### 配置

**文件**: `workflows/data_analysis.yaml`

```yaml
name: "data_analysis"
type: "dag"

nodes:
  - name: "collect"
    agent: "researcher"
    tools: ["web_search", "data_collection"]

  - name: "analyze"
    agent: "analyst"
    tools: ["data_processing", "statistics"]

  - name: "report"
    agent: "reporter"
    tools: ["report_generation", "visualization"]

edges:
  - from: "collect"
    to: "analyze"
  - from: "analyze"
    to: "report"
```

### 代码实现

**文件**: `workflows/data_analysis.go`

```go
package workflows

import (
    "context"
    "agentframework/workflow"
    "agentframework/agent"
)

func main() {
    ctx := context.Background()

    // 加载工作流配置
    workflow, err := host.LoadWorkflow("data_analysis")
    if err != nil {
        log.Fatal(err)
    }

    // 执行工作流
    result, err := workflow.Run(ctx, "分析昨天的销售数据")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Analysis Result: %s\n", result.Content)
}
```

---

## 研究助手

### 场景描述

多代理协作系统，使用 Researcher、Analyst 和 Reporter 三个专业角色。

### 配置

**文件**: `teams/research_team.yaml`

```yaml
name: "research_team"
description: "多代理研究团队"
mode: "parallel"

members:
  - id: "researcher_1"
    agent: "worker"
    role: "researcher"
    skills: ["web_search", "data_collection", "analysis"]

  - id: "analyst_1"
    agent: "worker"
    role: "analyst"
    skills: ["data_processing", "statistics"]

  - id: "reporter_1"
    agent: "worker"
    role: "reporter"
    skills: ["report_generation", "visualization"]
```

---

## 内容生成平台

### 场景描述

使用 Writer Agent 生成高质量内容，支持流式响应。

### 配置

**文件**: `agents/content_writer.yaml`

```yaml
name: "content_writer"
type: "worker"
role: "writer"
model: "ollama-llama3"
instructions: |
  你是一个专业的内容创作者，能够：
  1. 根据用户需求生成内容
  2. 遵循 SEO 最佳实践
  3. 生成结构化、易读的文档
  4. 支持多格式输出

temperature: 0.7
max_tokens: 2000
```

---

## 自动化测试

### 场景描述

使用 Developer 和 Reviewer Agent 构建自动化测试流程。

### 配置

**文件**: `workflows/automation_test.yaml`

```yaml
name: "automation_test"
type: "dag"

nodes:
  - name: "develop"
    agent: "developer"
    tools: ["code_generation", "code_review"]
    depends_on: []

  - name: "review"
    agent: "reviewer"
    tools: ["quality_check", "validation"]
    depends_on: ["develop"]

edges:
  - from: "develop"
    to: "review"
```

---

## 相关文档

- 📘 [Agent 概览](../components/agent/overview.md)
- 📘 [Workflow 概览](../components/workflow/overview.md)
- 📘 [Skills 概览](../components/skills/overview.md)
- 📘 [Sandbox 概览](../components/sandbox/overview.md)
- 📘 [配置指南](../../configuration/CONFIGURATION.md)

---

**Made with ❤️ by AgentFramework Team**
