# Agent API 参考

> **AgentFramework Agent API 完整参考**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [接口定义](#接口定义)
- [基础方法](#基础方法)
- [执行方法](#执行方法)
- [配置方法](#配置方法)
- [工具管理](#工具管理)
- [状态管理](#状态管理)
- [事件监听](#事件监听)
- [错误处理](#错误处理)

---

## 接口定义

### Agent 接口

**文件**: [agent/agent.go](../../agent/agent.go)

```go
type Agent interface {
    // ========== 基本信息 ==========
    Name() string
    Type() string
    Version() string

    // ========== 执行接口 ==========
    Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
    Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)

    // ========== 状态管理 ==========
    GetState() AgentState
    SetState(state AgentState)

    // ========== 配置管理 ==========
    GetModel() string
    SetModel(modelName string)
    GetConfig() *AgentConfig
    UpdateConfig(cfg *AgentConfig) error

    // ========== 工具管理 ==========
    GetTools() []string
    AddTool(tool ...Tool)
    RemoveTool(toolName ...string)
    HasTool(toolName string) bool

    // ========== 内存管理 ==========
    GetMemory() *Memory
    SetMemory(memory *Memory)
    ClearMemory() error

    // ========== 事件监听 ==========
    Subscribe(event string, handler EventHandler) error
    Unsubscribe(event string, handler EventHandler) error
    SubscribeStateChanges(callback StateChangeCallback)

    // ========== 生命周期 ==========
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Pause(ctx context.Context) error
    Resume(ctx context.Context) error
    Shutdown(ctx context.Context) error
    Destroy(ctx context.Context) error
}
```

---

## 基础方法

### Name()

获取 Agent 名称。

```go
func (a *Agent) Name() string
```

**返回值**: Agent 唯一名称

**示例**:

```go
name := agent.Name()
fmt.Printf("Agent 名称: %s\n", name)
// 输出: Agent 名称: chat
```

---

### Type()

获取 Agent 类型。

```go
func (a *Agent) Type() string
```

**返回值**: Agent 类型字符串

**可能的值**:
- `"chat"` - ChatAgent
- `"react"` - ReActAgent
- `"human"` - HumanAgent
- `"worker"` - WorkerAgent
- `"edge"` - EdgeAgent

**示例**:

```go
agentType := agent.Type()
fmt.Printf("Agent 类型: %s\n", agentType)
// 输出: Agent 类型: chat
```

---

### Version()

获取 Agent 版本。

```go
func (a *Agent) Version() string
```

**返回值**: 版本字符串

**示例**:

```go
version := agent.Version()
fmt.Printf("Agent 版本: %s\n", version)
// 输出: Agent 版本: 1.0.0
```

---

## 执行方法

### Run()

同步执行 Agent。

```go
func (a *Agent) Run(
    ctx context.Context,
    input string,
    opts ...model.Option,
) (*schema.Message, error)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **ctx** | context.Context | ✅ | 上下文对象 |
| **input** | string | ✅ | 用户输入 |
| **opts** | ...model.Option | ❌ | 可选参数 |

**返回值**:

| 类型 | 说明 |
|------|------|
| **\*schema.Message** | 响应消息 |
| **error** | 错误信息 |

**可选项**:

```go
// 温度控制
model.WithTemperature(0.8)

// Token 限制
model.WithMaxTokens(2000)

// Top-P 采样
model.WithTopP(0.9)

// Top-K 采样
model.WithTopK(40)

// 停止词
model.WithStopSequences([]string{"\n\n\n", "END"})

// 超时控制
model.WithTimeout(30*time.Second)
```

**示例**:

```go
// 基本执行
response, err := agent.Run(ctx, "你好")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.Content)

// 带选项执行
response, err := agent.Run(
    ctx,
    "请写一个故事",
    model.WithTemperature(0.8),
    model.WithMaxTokens(2000),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.Content)
```

---

### Stream()

流式执行 Agent。

```go
func (a *Agent) Stream(
    ctx context.Context,
    input string,
    opts ...model.Option,
) (*schema.StreamReader[*schema.Message], error)
```

**参数**: 同 `Run()`

**返回值**:

| 类型 | 说明 |
|------|------|
| **\*schema.StreamReader** | 流式读取器 |
| **error** | 错误信息 |

**示例**:

```go
// 创建流式请求
streamReader, err := agent.Stream(
    ctx,
    "请写一首长诗",
    model.WithTemperature(0.7),
)
if err != nil {
    log.Fatal(err)
}

// 逐块读取
for {
    msg, err := streamReader.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        log.Printf("接收错误: %v\n", err)
        break
    }

    // 处理消息块
    fmt.Print(msg.Content)

    // 可以提前退出
    if shouldStop(msg) {
        streamReader.Close()
        break
    }
}
```

---

## 配置方法

### GetModel()

获取当前使用的模型。

```go
func (a *Agent) GetModel() string
```

**返回值**: 模型名称字符串

**示例**:

```go
modelName := agent.GetModel()
fmt.Printf("当前模型: %s\n", modelName)
// 输出: 当前模型: ollama-llama3
```

---

### SetModel()

设置使用的模型。

```go
func (a *Agent) SetModel(modelName string)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **modelName** | string | ✅ | 模型名称 |

**示例**:

```go
// 切换到另一个模型
agent.SetModel("gpt-4")

// 验证切换
if agent.GetModel() == "gpt-4" {
    log.Println("已切换到 GPT-4")
}
```

---

### GetConfig()

获取 Agent 配置。

```go
func (a *Agent) GetConfig() *AgentConfig
```

**返回值**: Agent 配置对象

**示例**:

```go
cfg := agent.GetConfig()
fmt.Printf("配置: %+v\n", cfg)
```

---

### UpdateConfig()

更新 Agent 配置。

```go
func (a *Agent) UpdateConfig(cfg *AgentConfig) error
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **cfg** | \*AgentConfig | ✅ | 新配置 |

**返回值**: error - 错误信息

**示例**:

```go
// 获取当前配置
cfg := agent.GetConfig()

// 修改配置
cfg.Temperature = 0.9
cfg.MaxTokens = 4000

// 更新配置
err := agent.UpdateConfig(cfg)
if err != nil {
    log.Printf("配置更新失败: %v\n", err)
}
```

---

## 工具管理

### GetTools()

获取所有可用工具列表。

```go
func (a *Agent) GetTools() []string
```

**返回值**: 工具名称列表

**示例**:

```go
tools := agent.GetTools()
fmt.Printf("可用工具: %v\n", tools)
// 输出: 可用工具: [http_request file_operation code_execution]
```

---

### AddTool()

添加工具。

```go
func (a *Agent) AddTool(tool ...Tool)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **tool** | ...Tool | ✅ | 一个或多个工具对象 |

**示例**:

```go
// 添加单个工具
customTool := &MyCustomTool{}
agent.AddTool(customTool)

// 添加多个工具
agent.AddTool(
    &Tool1{},
    &Tool2{},
    &Tool3{},
)
```

---

### RemoveTool()

移除工具。

```go
func (a *Agent) RemoveTool(toolName ...string)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **toolName** | ...string | ✅ | 一个或多个工具名称 |

**示例**:

```go
// 移除单个工具
agent.RemoveTool("file_operation")

// 移除多个工具
agent.RemoveTool("http_request", "code_execution")
```

---

### HasTool()

检查是否有指定工具。

```go
func (a *Agent) HasTool(toolName string) bool
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **toolName** | string | ✅ | 工具名称 |

**返回值**: bool - 是否有该工具

**示例**:

```go
if agent.HasTool("http_request") {
    log.Println("Agent 支持 HTTP 请求")
} else {
    log.Println("Agent 不支持 HTTP 请求")
}
```

---

## 状态管理

### GetState()

获取当前状态。

```go
func (a *Agent) GetState() AgentState
```

**返回值**: AgentState 状态值

**可能的值**:
- `StateCreated` (0) - 已创建
- `StateInitialized` (1) - 已初始化
- `StateRunning` (2) - 运行中
- `StatePaused` (3) - 已暂停
- `StateCompleted` (4) - 已完成
- `StateFailed` (5) - 失败
- `StateDestroyed` (6) - 已销毁

**示例**:

```go
state := agent.GetState()
switch state {
case agent.StateRunning:
    log.Println("Agent 正在运行")
case agent.StatePaused:
    log.Println("Agent 已暂停")
case agent.StateFailed:
    log.Println("Agent 执行失败")
}
```

---

### SetState()

设置状态。

```go
func (a *Agent) SetState(state AgentState)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **state** | AgentState | ✅ | 新状态 |

**注意**: 通常不直接调用此方法，由 Agent 内部管理状态

**示例**:

```go
// 强制设置失败状态（不推荐）
agent.SetState(agent.StateFailed)
```

---

## 事件监听

### Subscribe()

订阅事件。

```go
func (a *Agent) Subscribe(event string, handler EventHandler) error
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **event** | string | ✅ | 事件名称 |
| **handler** | EventHandler | ✅ | 事件处理函数 |

**返回值**: error - 错误信息

**可用事件**:

| 事件 | 说明 | 触发时机 |
|------|------|----------|
| **"agent:start"** | Agent 启动 | Start() 调用成功 |
| **"agent:complete"** | Agent 完成 | 任务成功完成 |
| **"agent:fail"** | Agent 失败 | 任务执行失败 |
| **"agent:pause"** | Agent 暂停 | Pause() 调用成功 |
| **"agent:resume"** | Agent 恢复 | Resume() 调用成功 |
| **"agent:shutdown"** | Agent 关闭 | Shutdown() 调用成功 |
| **"tool:invoke"** | 工具调用 | 任何工具被调用 |
| **"tool:success"** | 工具成功 | 工具执行成功 |
| **"tool:fail"** | 工具失败 | 工具执行失败 |

**示例**:

```go
// 订阅完成事件
err := agent.Subscribe("agent:complete", func(event *Event) error {
    log.Infof("任务完成: %v\n", event.Data)
    return nil
})
if err != nil {
    log.Printf("订阅失败: %v\n", err)
}

// 订阅工具调用事件
err = agent.Subscribe("tool:invoke", func(event *Event) error {
    toolName := event.Data["tool_name"].(string)
    log.Infof("调用工具: %s\n", toolName)
    return nil
})
```

---

### Unsubscribe()

取消订阅事件。

```go
func (a *Agent) Unsubscribe(event string, handler EventHandler) error
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **event** | string | ✅ | 事件名称 |
| **handler** | EventHandler | ✅ | 事件处理函数 |

**返回值**: error - 错误信息

**示例**:

```go
handler := func(event *Event) error {
    log.Println("处理事件")
    return nil
}

// 订阅
agent.Subscribe("agent:complete", handler)

// 取消订阅
agent.Unsubscribe("agent:complete", handler)
```

---

### SubscribeStateChanges()

订阅状态变化。

```go
func (a *Agent) SubscribeStateChanges(callback StateChangeCallback)
```

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| **callback** | StateChangeCallback | ✅ | 状态变化回调 |

**回调签名**:

```go
type StateChangeCallback func(oldState, newState AgentState)
```

**示例**:

```go
agent.SubscribeStateChanges(func(oldState, newState AgentState) {
    log.Infof("状态变化: %s -> %s\n", oldState, newState)

    if newState == agent.StateFailed {
        log.Println("Agent 失败，通知管理员")
        notifyAdmin(agent)
    }
})
```

---

## 错误处理

### 错误类型

| 错误 | 说明 | 处理方式 |
|------|------|----------|
| **ErrAgentNotFound** | Agent 不存在 | 检查名称，使用 `host.GetAgent()` |
| **ErrAgentNotInitialized** | Agent 未初始化 | 调用 `Initialize()` |
| **ErrAgentNotRunning** | Agent 未运行 | 调用 `Start()` |
| **ErrAgentPaused** | Agent 已暂停 | 调用 `Resume()` |
| **ErrAgentDestroyed** | Agent 已销毁 | 重新创建 Agent |
| **ErrInvalidInput** | 输入无效 | 验证输入参数 |
| **ErrToolNotFound** | 工具不存在 | 检查工具名称 |
| **ErrToolInvokeFailed** | 工具调用失败 | 检查工具配置 |
| **ErrModelNotFound** | 模型不存在 | 检查模型配置 |
| **ErrTimeout** | 执行超时 | 增加超时时间 |

### 错误处理示例

```go
// ✅ 正确：详细错误处理
response, err := agent.Run(ctx, "task")
if err != nil {
    switch {
    case errors.Is(err, ErrAgentNotInitialized):
        log.Println("Agent 未初始化")
        if initErr := agent.Initialize(ctx); initErr != nil {
            return fmt.Errorf("初始化失败: %w", initErr)
        }
        return agent.Run(ctx, "task")

    case errors.Is(err, ErrAgentPaused):
        log.Println("Agent 已暂停，尝试恢复")
        if resumeErr := agent.Resume(ctx); resumeErr != nil {
            return fmt.Errorf("恢复失败: %w", resumeErr)
        }
        return agent.Run(ctx, "task")

    case errors.Is(err, ErrTimeout):
        log.Println("执行超时")
        return fmt.Errorf("任务超时: %w", err)

    default:
        return fmt.Errorf("未知错误: %w", err)
    }
}

// ❌ 错误：忽略错误
response, err := agent.Run(ctx, "task")
if err != nil {
    log.Printf("执行失败: %v\n", err)
    return err // 没有具体处理
}
```

---

## 完整示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "agentframework/agent"
)

func main() {
    ctx := context.Background()

    // 创建 Agent
    chatAgent := agent.NewChatAgent(
        agent.WithName("chat"),
        agent.WithInstructions("你是一个有用的AI助手"),
        agent.WithModel("ollama-llama3"),
    )

    // 初始化
    if err := chatAgent.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    // 执行
    response, err := chatAgent.Run(ctx, "你好")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)

    // 关闭
    if err := chatAgent.Shutdown(ctx); err != nil {
        log.Printf("关闭失败: %v\n", err)
    }
}
```

### 高级使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "agentframework/agent"
)

func main() {
    ctx := context.Background()

    // 创建 ReActAgent
    reactAgent := agent.NewReActAgent(
        agent.WithName("react"),
        agent.WithInstructions("你是一个专业的研究助理"),
        agent.WithModel("ollama-llama3"),
        agent.WithMaxIterations(10),
        agent.WithTools("web_search", "data_analysis"),
    )

    // 初始化
    if err := reactAgent.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    // 订阅事件
    reactAgent.Subscribe("agent:complete", func(event *agent.Event) error {
        log.Info("任务完成")
        return nil
    })

    reactAgent.Subscribe("agent:fail", func(event *agent.Event) error {
        log.Error("任务失败")
        return nil
    })

    // 监听状态变化
    reactAgent.SubscribeStateChanges(func(oldState, newState agent.AgentState) {
        log.Infof("状态变化: %s -> %s", oldState, newState)
    })

    // 执行（带超时）
    timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    response, err := reactAgent.Run(
        timeoutCtx,
        "请研究 AI 技术的最新发展",
    )
    if err != nil {
        log.Printf("执行失败: %v\n", err)
    } else {
        fmt.Println(response.Content)
    }

    // 优雅关闭
    if err := reactAgent.Shutdown(ctx); err != nil {
        log.Printf("关闭失败: %v\n", err)
    }
}
```

---

## 相关文档

- 📘 [Agent 概览](overview.md) - Agent 系统概览
- 📘 [Agent 类型](types.md) - 各种 Agent 类型
- 📘 [Agent 生命周期](lifecycle.md) - 生命周期管理
- 📘 [配置指南](../../configuration/CONFIGURATION.md) - 详细配置
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
