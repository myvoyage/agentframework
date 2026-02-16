# API 完整参考

> **AgentFramework API 完整参考**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [Host API](#host-api)
- [Agent API](#agent-api)
- [Workflow API](#workflow-api)
- [Skills API](#skills-api)
- [Sandbox API](#sandbox-api)
- [Model API](#model-api)
- [Collaboration API](#collaboration-api)

---

## Host API

### 基本方法

```go
package host

// NewHost 创建新的 Host 实例
func NewHost(
    ctx context.Context,
    cfg *HostConfig,
    mf ModelFactory,
    tr map[string]tool.BaseTool,
    opts ...HostOption,
) (*Host, error)
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|--------|
| **ctx** | context.Context | ✅ | Go 上下文对象 |
| **cfg** | *HostConfig | ✅ | Host 配置对象 |
| **mf** | ModelFactory | ✅ | 模型工厂接口 |
| **tr** | map[string]tool.BaseTool | ✅ | 工具注册表 |
| **opts** | ...HostOption | ❌ | 可选配置项 |

**返回值**:
- **\*Host**: 新创建的 Host 实例
- **error**: 错误信息

### 配置方法

```go
// 获取 Host 配置
func (h *Host) Config() *HostConfig

// 更新 Host 配置
func (h *Host) UpdateConfig(cfg *HostConfig) error

// 设置默认模型
func (h *Host) SetDefaultModel(modelName string) error

// 获取所有 Agent
func (h *Host) GetAgents() map[string]Agent

// 获取所有 Workflow
func (h *Host) GetWorkflows() map[string]Workflow

// 获取所有 Skill
func (h *Host) GetSkills() map[string]Skill

// 获取所有 Tool
func (h *Host) GetTools() map[string]tool.BaseTool
```

---

## Agent API

### 基本接口

```go
package agent

// Agent 接口定义
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
    HasTool(toolName string) bool

    // 内存管理
    GetMemory() *Memory
    SetMemory(memory *Memory)

    // 事件监听
    Subscribe(event string, handler EventHandler) error
    Unsubscribe(event string, handler EventHandler) error
    SubscribeStateChanges(callback StateChangeCallback)

    // 生命周期
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Pause(ctx context.Context) error
    Resume(ctx context.Context) error
    Shutdown(ctx context.Context) error
    Destroy(ctx context.Context) error
}
```

---

## Workflow API

### 工作流接口

```go
package workflow

// 工作流接口定义
type Workflow interface {
    // 基本信息
    Name() string
    Type() WorkflowType
    Version() string

    // 执行接口
    Run(ctx context.Context, input string, opts ...Option) (*schema.Message, error)

    // 节点管理
    AddNode(node Node) error
    RemoveNode(nodeID string) error
    GetNode(nodeID string) (Node, bool)

    // 边管理
    AddEdge(from, to string, condition ...string) error
    RemoveEdge(from, to string) error

    // 检查点
    SetCheckpoint(enabled bool) error

    // 生命周期
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Pause(ctx context.Context) error
    Shutdown(ctx context.Context) error
    Destroy(ctx context.Context) error
}
```

---

## Skills API

### 技能接口

```go
package skills

// Skill 接口定义
type Skill interface {
    // 元信息
    Info(ctx context.Context) (*schema.ToolInfo, error)

    // 执行接口
    Invoke(ctx context.Context, input string) (string, error)

    // 状态管理
    IsEnabled(ctx context.Context) bool
    SetEnabled(enabled bool)

    // 元数据
    GetMetadata(ctx context.Context) SkillMetadata
    SetMetadata(metadata SkillMetadata)
}
```

---

## Sandbox API

### 沙箱管理接口

```go
package sandbox

// Sandbox 管理器接口
type SandboxManager interface {
    // 获取执行器
    GetExecutor(executorType string) (Executor, error)

    // 获取配置
    GetConfig() *SandboxConfig

    // 设置配置
    SetConfig(cfg *SandboxConfig) error

    // 生命周期
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

---

## Model API

### 模型管理接口

```go
package model

// 模型工厂接口
type ModelFactory interface {
    // 创建模型
    Create(ctx context.Context, modelName string, cfg ModelConfig) (ChatModel, error)

    // 获取模型
    Get(ctx context.Context, modelName string) (ChatModel, error)

    // 设置默认模型
    SetDefault(modelName string) error
}
```

---

## Collaboration API

### 协作系统接口

```go
package collaboration

// Agent 团队接口
type AgentTeam interface {
    // 成员管理
    AddMember(member TeamMember) error
    RemoveMember(memberID string) error
    GetMembers() []TeamMember

    // 任务执行
    ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error)

    // 消息通信
    SendMessage(ctx context.Context, msg *Message) error
    Broadcast(ctx context.Context, msg *Message) error
    Subscribe(topic string, handler EventHandler) error
    Unsubscribe(topic string, handler EventHandler) error

    // 生命周期
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

---

## 快速开始

### 创建 Host

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

    // 创建配置
    cfg := &host.HostConfig{
        Name:         "my-agent-app",
        Version:       "1.0.0",
        DefaultModel:  "ollama-llama3",
        Models: map[string]host.ModelConfig{
            "ollama-llama3": {
                Type:    "ollama",
                Model:   "llama3",
                BaseURL: "http://localhost:11434",
                Enabled: true,
            },
        },
    }

    // 创建模型工厂
    mf := model.NewDefaultModelFactory(cfg)

    // 创建 Host
    h, err := host.NewHost(ctx, cfg, mf, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Host created successfully")
}
```

---

## 相关文档

- 📘 [Agent 概览](agent/overview.md) - Agent 系统概览
- 📘 [Workflow 概览](workflow/overview.md) - Workflow 系统概览
- 📘 [Skills 概览](skills/overview.md) - Skills 系统概览
- 📘 [Sandbox 概览](sandbox/overview.md) - Sandbox 系统概览
- 📘 [配置指南](../../configuration/CONFIGURATION.md) - 详细配置说明
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
