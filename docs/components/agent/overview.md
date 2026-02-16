# Agent 系统概览

> **AgentFramework Agent 组件文档**
> **版本**: v2.0.0
> **最后更新**: 2026-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [核心接口](#核心接口)
- [Agent 类型](#agent-类型)
- [WorkerAgent 角色](#workeragent-角色)
- [设计原则](#设计原则)
- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [核心特性](#核心特性)

---

## 系统简介

Agent 是 AgentFramework 的核心执行单元，负责使用大语言模型（LLM）和工具（Tools/Skills）完成用户指定的任务。

### 项目规模

| 指标 | 数值 |
|------|------|
| Agent 类型 | **12+ 种** |
| Worker 角色 | **7 种** |
| MCP 工具 | **44 个** |
| 测试覆盖率 | **80%+** |

### 核心职责

| 职责 | 说明 |
|------|------|
| **对话管理** | 管理与用户的对话历史和上下文 |
| **模型交互** | 调用 LLM 进行推理和生成 |
| **工具调用** | 使用工具扩展能力，执行具体操作 |
| **状态管理** | 管理执行状态和生命周期 |
| **记忆管理** | 智能压缩和管理对话历史 |
| **流式响应** | 支持流式和非流式响应 |

### 技术特点

- ✅ **高性能**: 基于 Go 的高并发处理能力
- ✅ **可扩展**: 插件化架构，支持自定义 Agent
- ✅ **类型丰富**: 12+ 种内置 Agent 类型
- ✅ **工具集成**: 44+ MCP 工具原生支持
- ✅ **内存优化**: 智能内存管理和 Token 压缩
- ✅ **安全可靠**: HITL 支持、检查点恢复、错误重试

---

## 核心接口

### Agent 接口定义

**文件**: [agent/agent.go](../../agent/agent.go)

```go
type Agent interface {
    // 基本信息
    Name() string
    Type() string
    Version() string

    // 执行接口
    Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
    Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)

    // 状态管理
    GetState() AgentState
    SetState(state AgentState)

    // 配置
    GetModel() string
    SetModel(modelName string)

    // 工具管理
    GetTools() []string
    AddTool(tool ...Tool)
    RemoveTool(toolName ...string)
}
```

### 接口说明

| 方法 | 说明 | 参数 | 返回值 |
|------|------|------|--------|
| **Name()** | 获取 Agent 名称 | - | string |
| **Type()** | 获取 Agent 类型 | - | string |
| **Version()** | 获取 Agent 版本 | - | string |
| **Run()** | 同步执行 Agent | ctx, input, opts | *schema.Message, error |
| **Stream()** | 流式执行 Agent | ctx, input, opts | *schema.StreamReader, error |
| **GetState()** | 获取当前状态 | - | AgentState |
| **SetState()** | 设置状态 | state | - |
| **GetModel()** | 获取使用的模型 | - | string |
| **SetModel()** | 设置模型 | modelName | - |
| **GetTools()** | 获取工具列表 | - | []string |
| **AddTool()** | 添加工具 | tool... | - |
| **RemoveTool()** | 移除工具 | toolName... | - |

---

## Agent 类型

### 类型概览

AgentFramework 提供了 5 种内置 Agent 类型：

| 类型 | 文件 | 复杂度 | 使用场景 |
|------|------|--------|---------|
| **ChatAgent** | [chat_agent.go](../../agent/chat_agent.go) | ⭐ 简单对话、客服机器人 |
| **ReActAgent** | [react_agent.go](../../agent/react_agent.go) | ⭐⭐⭐ 推理-行动循环、复杂任务 |
| **HumanAgent** | [hitl.go](../../agent/hitl.go) | ⭐⭐ 人工审核、HITL 场景 |
| **WorkerAgent** | [swe_agent.go](../../agent/swe_agent.go) | ⭐⭐⭐⭐ 专业工作代理、7 种角色 |
| **EdgeAgent** | [edge_agent.go](../../agent/edge_agent.go) | ⭐⭐⭐ 边缘计算、资源受限环境 |

### 类型对比

```
┌───────────────────────────────────────────────────────┐
│                    Agent 类型对比                   │
├───────────────────────────────────────────────────────┤
│ 类型          │ 推理 │ 工具 │ HITL │ 复杂度 │
├───────────────────────────────────────────────────────┤
│ ChatAgent     │  ❌  │  ✅  │  ✅  │  ⭐    │
│ ReActAgent    │  ✅  │  ✅  │  ✅  │  ⭐⭐⭐ │
│ HumanAgent    │  ❌  │  ❌  │  ✅✅ │  ⭐⭐  │
│ WorkerAgent   │  ✅  │  ✅✅ │  ✅  │  ⭐⭐⭐⭐ │
│ EdgeAgent     │  ✅  │  ✅  │  ❌  │  ⭐⭐⭐  │
└───────────────────────────────────────────────────────┘
```

---

## 设计原则

### SOLID 原则应用

| 原则 | 评分 | 说明 |
|------|------|------|
| **S - 单一职责** | ⭐⭐⭐⭐⭐ | 每个 Agent 类型职责明确 |
| **O - 开闭原则** | ⭐⭐⭐⭐⭐ | 通过接口扩展，无需修改现有代码 |
| **L - 里氏替换** | ⭐⭐⭐⭐⭐ | 所有 Agent 可互相替换 |
| **I - 接口隔离** | ⭐⭐⭐⭐☆ | 接口设计良好，少数方法可优化 |
| **D - 依赖倒置** | ⭐⭐⭐⭐⭐ | 依赖抽象接口，不依赖具体实现 |

### 其他设计原则

| 原则 | 应用 | 说明 |
|------|------|------|
| **KISS** | ⭐⭐⭐⭐ | API 简洁直观，避免过度设计 |
| **DRY** | ⭐⭐⭐⭐⭐ | 公共逻辑抽取到 BaseAgent |
| **YAGNI** | ⭐⭐⭐⭐ | 只实现必要功能，无未来预留 |
| **性能优先** | ⭐⭐⭐⭐⭐ | 内置缓存、连接池、并发优化 |

---

## 架构设计

### 继承层次

```
┌───────────────────────────────────────────────────────┐
│                   Agent 继承树                    │
└───────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │   Agent (接口)  │
              └─────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │   BaseAgent    │
              │  (基础实现)     │
              └─────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌───────────┐  ┌───────────┐  ┌───────────┐
│ ChatAgent  │  │ ReActAgent │  │ HumanAgent │
└───────────┘  └───────────┘  └───────────┘
        │               │               │
        ▼               ▼               ▼
┌───────────┐  ┌───────────┐  ┌───────────┐
│ WorkerAgent │  │ EdgeAgent  │
└───────────┘  └───────────┘
```

### 核心组件关系

```
┌───────────────────────────────────────────────────────┐
│                   Agent 系统关系                    │
└───────────────────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  Host       │ │  Model      │ │  Skills     │
│  Manager    │ │  Factory    │ │  Library    │
└─────────────┘ └─────────────┘ └─────────────┘
        │               │               │
        └───────────────┼───────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │    Agent       │
              │  (具体实例)      │
              └─────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ ThreadStore │ │ Tool        │ │ Middleware  │
└─────────────┘ └─────────────┘ └─────────────┘
```

---

## 快速开始

### 创建基本 Agent

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

    // 1. 创建模型工厂
    modelFactory := agent.NewDefaultModelFactory(
        &agent.DefaultModelFactoryConfig{
            Models: map[string]agent.ModelConfig{
                "ollama-llama3": {
                    Type:    "ollama",
                    Model:   "llama3",
                    BaseURL: "http://localhost:11434",
                    Enabled: true,
                },
            },
        },
    )

    // 2. 创建配置
    cfg := &agent.HostConfig{
        Name:         "my-agent-app",
        Version:      "1.0.0",
        DefaultModel: "ollama-llama3",
    }

    // 3. 创建 Host
    host, err := agent.NewHost(ctx, cfg, modelFactory, nil)
    if err != nil {
        log.Fatal(err)
    }

    // 4. 获取或创建 Agent
    chatAgent, err := host.GetAgent("chat")
    if err != nil {
        log.Fatal(err)
    }

    // 5. 运行 Agent
    response, err := chatAgent.Run(ctx, "你好，请介绍一下你自己")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)
}
```

### 流式响应

```go
// 使用 Stream API 获取流式响应
streamReader, err := chatAgent.Stream(ctx, "请写一首关于春天的诗")
if err != nil {
    log.Fatal(err)
}

// 逐块读取响应
for {
    msg, err := streamReader.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        log.Fatal(err)
    }

    fmt.Print(msg.Content) // 逐块输出
}
```

### 添加工具

```go
// 方式1：通过配置添加
cfg := &agent.HostConfig{
    // ... 其他配置
    Tools: []string{
        "http_request",
        "file_operation",
        "code_execution",
    },
}

// 方式2：运行时添加
chatAgent.AddTool("http_request")
chatAgent.AddTool("file_operation")

// 方式3：移除工具
chatAgent.RemoveTool("file_operation")
```

---

## 核心特性

### 1. 对话管理

#### 消息压缩

AgentFramework 智能管理对话历史，自动压缩长对话：

```go
memory := &agent.MemoryConfig{
    // 监控配置
    Monitoring: true,
    Interval:  5 * time.Second,
    HistorySize: 100,

    // 告警规则
    AlertRules: []agent.AlertRule{
        {
            ID:         "heap-512mb",
            Name:       "堆内存告警",
            Severity:   "warning",
            Threshold:  536870912, // 512MB
            Operator:   ">",
            Duration:   30 * time.Second,
        },
    },

    // 缓存配置
    Cache: agent.CacheConfig{
        MaxSize:         200,
        TTL:             1 * time.Hour,
        DynamicWeights:   true,
    },
}
```

#### Token 压缩

支持多种压缩策略：

| 策略 | 说明 | 适用场景 |
|------|------|---------|
| **Truncate** | 截断旧消息 | 简单场景 |
| **Summarize** | 摘要旧消息 | 保留上下文 |
| **Hybrid** | 混合策略 | 平衡性能和质量 |

```go
tokenCompression := &agent.TokenCompressionConfig{
    Enabled: true,

    // 压缩策略
    Strategy: "hybrid", // truncate, summarize, hybrid

    // 目标配置
    TargetTokens: 4000,
    MinTokens:     500,
    MaxTokens:     8000,

    // 摘要配置
    PreserveSystemMessages: true,
    SummaryModel:          "ollama-llama3",
    SummaryMaxTokens:     500,
    Temperature:           0.3,
}
```

### 2. 工具调用

#### 工具管理

```go
// 获取所有可用工具
tools := chatAgent.GetTools()
fmt.Printf("可用工具: %v\n", tools)

// 动态添加工具
chatAgent.AddTool(
    &MyCustomTool{},  // 自定义工具
    "http_request",        // 内置工具
)

// 移除工具
chatAgent.RemoveTool("file_operation")
```

#### 工具执行流程

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  User   │ --> │  Agent  │ --> │   LLM    │ --> │  Tool    │
│  Input  │     │  Reasoning│     │ Decision │     │ Execute  │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
                                                    │
                                                    ▼
                                             ┌─────────┐
                                             │  Result  │
                                             └─────────┘
                                                    │
                                                    ▼
┌─────────┐     ┌─────────┐     ┌─────────┐
│  User   │ <-- │  Agent  │ <-- │   LLM    │
│  Output  │     │ Response │     │ Synthesis │     │          │
└─────────┘     └─────────┘     └─────────┘
```

### 3. 人工介入 (HITL)

#### HITL 配置

```yaml
agents:
  - name: "human_agent"
    type: "human"
    hitl:
      enabled: true
      approval_mode: "manual"  # manual, auto

      # 通知配置
      notification_channels:
        - "slack"
        - "email"

      # 审批节点
      approval_nodes:
        - "critical_decision"
        - "data_deletion"
```

#### HITL 流程

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│  Agent  │ --> │ Decision │ --> │  Human   │
│ Request │     │  Point   │     │  Review  │
└─────────┘     └─────────┘     └─────────┘
                                     │
                                     ▼
                              ┌──────────────┐
                              │  Approval    │
                              │  Required?  │
                              └──────────────┘
                                │           │
                               Yes          No
                                │           │
                                ▼           ▼
                        ┌───────────┐  ┌───────────┐
                        │  Human    │  │  Auto     │
                        │  Approved │  │  Proceed  │
                        └───────────┘  └───────────┘
                                │           │
                                └───────────┘
                                        │
                                        ▼
                                ┌───────────┐
                                │  Continue  │
                                │  Execution │
                                └───────────┘
```

### 4. 流式响应

#### 流式 API 优势

| 特性 | 说明 |
|------|------|
| **低延迟** | 首块快速响应，无需等待完整生成 |
| **更好的 UX** | 用户可以看到生成过程 |
| **超时控制** | 可以随时中断生成 |
| **内存优化** | 不需要保存完整响应 |

#### 流式使用示例

```go
// 创建流式请求
streamReader, err := chatAgent.Stream(
    ctx,
    "请写一个长篇故事",
    model.WithTemperature(0.8),
    model.WithMaxTokens(2000),
)
if err != nil {
    log.Fatal(err)
}

// 处理流式响应
for {
    msg, err := streamReader.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        log.Printf("接收错误: %v\n", err)
        break
    }

    // 处理每个消息块
    fmt.Print(msg.Content)

    // 可以添加超时控制
    select {
    case <-ctx.Done():
        streamReader.Close()
        return ctx.Err()
    default:
        // 继续处理
    }
}
```

---

## 相关文档

- 📘 [Agent 类型详解](types.md) - 各种 Agent 类型详细说明
- 📘 [Agent 生命周期](lifecycle.md) - 创建、运行、销毁
- 📘 [Agent API 参考](api.md) - 完整 API 文档
- 📘 [架构概览](../architecture/ARCHITECTURE_OVERVIEW.md) - 系统架构
- 📘 [快速开始](../quickstart/QUICKSTART.md) - 5 分钟上手
- 📘 [最佳实践](../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
