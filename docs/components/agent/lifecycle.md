# Agent 生命周期

> **AgentFramework Agent 生命周期管理文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [生命周期概览](#生命周期概览)
- [创建阶段](#创建阶段)
- [初始化阶段](#初始化阶段)
- [运行阶段](#运行阶段)
- [暂停恢复](#暂停恢复)
- [销毁阶段](#销毁阶段)
- [状态管理](#状态管理)
- [最佳实践](#最佳实践)

---

## 生命周期概览

### 状态转换图

```
┌──────────────────────────────────────────────────────────────┐
│                 Agent 生命周期状态图                      │
└──────────────────────────────────────────────────────────────┘

      ┌─────────┐
      │  Created │  (已创建）
      └─────────┘
            │
            ▼
      ┌─────────┐
      │Initialized│  (已初始化）
      └─────────┘
            │
            ▼
      ┌─────────┐
      │  Running │◄─────┐
      └─────────┘       │
            │            │
            ▼            │
      ┌─────────┐       │
      │ Paused  │───────┘
      └─────────┘       │
            │            │
            ▼            ▼
      ┌─────────┐     ┌─────────┐
      │Completed │     │  Failed  │
      └─────────┘     └─────────┘
                           │
                           ▼
                     ┌─────────┐
                     │Destroyed │  (已销毁）
                     └─────────┘
```

### 状态说明

| 状态 | 值 | 说明 | 可执行操作 |
|------|-----|------|----------|
| **Created** | 0 | Agent 对象已创建，但未初始化 | 初始化 |
| **Initialized** | 1 | Agent 已初始化，可以配置 | 启动、配置 |
| **Running** | 2 | Agent 正在运行 | 执行任务、暂停 |
| **Paused** | 3 | Agent 已暂停 | 恢复、销毁 |
| **Completed** | 4 | Agent 正常完成任务 | 获取结果、销毁 |
| **Failed** | 5 | Agent 执行失败 | 获取错误、重试、销毁 |
| **Destroyed** | 6 | Agent 已销毁 | 无 |

---

## 创建阶段

### 创建方式

#### 1. 通过配置创建

**YAML 配置**:

```yaml
agents:
  - name: "chat"
    type: "chat"
    model: "ollama-llama3"
    instructions: "你是一个有用的AI助手"
```

**自动加载**:

```go
// Host 自动加载配置中的 Agent
host, err := agent.NewHost(ctx, cfg, modelFactory, nil)
if err != nil {
    log.Fatal(err)
}

// 获取已创建的 Agent
chatAgent, err := host.GetAgent("chat")
```

#### 2. 通过代码创建

```go
// 创建 ChatAgent
chatAgent := agent.NewChatAgent(
    agent.WithName("chat"),
    agent.WithInstructions("你是一个有用的AI助手"),
    agent.WithModel("ollama-llama3"),
)

// 创建 ReActAgent
reactAgent := agent.NewReActAgent(
    agent.WithName("react"),
    agent.WithInstructions("你是一个专业的研究助理"),
    agent.WithModel("ollama-llama3"),
    agent.WithMaxIterations(10),
)
```

#### 3. 通过工厂创建

```go
// 使用 Host 工厂方法
host.RegisterAgentFactory("custom", func(ctx context.Context, cfg *AgentConfig) (Agent, error) {
    return &CustomAgent{
        BaseAgent: agent.NewBaseAgent(cfg),
        // 自定义初始化
    }, nil
})

// 创建自定义 Agent
customAgent, err := host.CreateAgent("custom", cfg)
```

### 创建参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| **Name** | string | ✅ | - | Agent 唯一名称 |
| **Type** | string | ✅ | "chat" | Agent 类型 |
| **Instructions** | string | ❌ | "" | 系统提示词 |
| **Model** | string | ❌ | default_model | 使用的模型 |
| **Temperature** | float64 | ❌ | 0.7 | 生成温度 |
| **MaxTokens** | int | ❌ | 2000 | 最大 Token 数 |
| **Tools** | []string | ❌ | [] | 工具列表 |
| **HITL** | bool | ❌ | false | 启用 HITL |

### 创建示例

```go
// 完整创建示例
chatAgent := agent.NewChatAgent(
    // 基本配置（必填）
    agent.WithName("chat"),
    agent.WithType("chat"),
    agent.WithModel("ollama-llama3"),

    // 可选配置
    agent.WithInstructions("你是一个有用的AI助手"),
    agent.WithTemperature(0.7),
    agent.WithMaxTokens(2000),
    agent.WithTools("http_request", "file_operation"),
    agent.WithHITL(true),

    // 高级配置
    agent.WithMiddleware(loggingMiddleware),
    agent.WithCacheConfig(cacheConfig),
    agent.WithMemoryConfig(memoryConfig),
)
```

---

## 初始化阶段

### 初始化流程

```
┌──────────────────────────────────────────────────────────────┐
│               Agent 初始化流程                        │
└──────────────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 1. 验证配置  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 2. 初始化状态  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 3. 加载工具    │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 4. 初始化内存  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 5. 注册中间件 │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 6. 准备就绪    │
              └─────────────┘
```

### 初始化代码

```go
// BaseAgent 初始化
func (a *BaseAgent) Initialize(ctx context.Context) error {
    // 1. 验证配置
    if err := a.validateConfig(); err != nil {
        return fmt.Errorf("config validation failed: %w", err)
    }

    // 2. 初始化状态
    a.state = StateInitialized
    a.startTime = time.Now()

    // 3. 加载工具
    if err := a.loadTools(ctx); err != nil {
        return fmt.Errorf("failed to load tools: %w", err)
    }

    // 4. 初始化内存
    if err := a.memory.Initialize(ctx); err != nil {
        return fmt.Errorf("memory initialization failed: %w", err)
    }

    // 5. 注册中间件
    if err := a.registerMiddlewares(); err != nil {
        return fmt.Errorf("middleware registration failed: %w", err)
    }

    // 6. 准备就绪
    a.ready = true

    log.Info("agent initialized",
        "name", a.name,
        "type", a.Type(),
        "duration", time.Since(a.startTime),
    )

    return nil
}
```

### 初始化检查点

| 检查点 | 说明 | 失理方式 |
|--------|------|----------|
| **配置验证** | 确保参数合法 | 返回详细错误信息 |
| **工具加载** | 验证工具可用性 | 检查工具注册表 |
| **内存分配** | 确保资源充足 | 监控内存使用 |
| **中间件注册** | 验证中间件链 | 测试中间件功能 |

---

## 运行阶段

### 执行模式

#### 同步执行

```go
// 简单同步执行
response, err := agent.Run(ctx, "请介绍一下你自己")
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content)
```

**特点**:
- ✅ 简单直观
- ✅ 易于调试
- ❌ 阻塞直到完成
- 适合: 简单任务、快速请求

#### 流式执行

```go
// 流式执行
streamReader, err := agent.Stream(ctx, "请写一首长诗")
if err != nil {
    log.Fatal(err)
}

// 逐块处理
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
}
```

**特点**:
- ✅ 低延迟响应
- ✅ 更好的用户体验
- ✅ 内存优化
- 适合: 长文本生成、实时响应

### 执行流程

```
┌──────────────────────────────────────────────────────────────┐
│                 Agent 执行流程                            │
└──────────────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 1. 前置处理  │
              │ (Middleware) │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 2. 输入验证  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 3. 记忆检索  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 4. 模型调用  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 5. 工具执行  │
              │ (如果需要)   │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 6. 后置处理  │
              │ (Middleware) │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 7. 记忆更新  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 8. 返回结果  │
              └─────────────┘
```

### 执行选项

```go
// 温度控制
response, err := agent.Run(
    ctx,
    "请生成一个故事",
    model.WithTemperature(0.8), // 0.0-2.0，越高越随机
)

// Token 限制
response, err := agent.Run(
    ctx,
    "请写一篇文章",
    model.WithMaxTokens(4000),
)

// 超时控制
timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
response, err := agent.Run(timeoutCtx, "请分析这个文件")

// 停止词控制
response, err := agent.Run(
    ctx,
    "请生成关键词",
    model.WithStopSequences(["\n\n\n", "END"]),
)
```

---

## 暂停恢复

### 暂停机制

```go
// 暂停 Agent
err := agent.Pause(ctx)
if err != nil {
    log.Printf("暂停失败: %v\n", err)
}

// 检查暂停状态
if agent.GetState() == agent.StatePaused {
    log.Println("Agent 已暂停")
}
```

### 恢复机制

```go
// 恢复 Agent
err := agent.Resume(ctx)
if err != nil {
    log.Printf("恢复失败: %v\n", err)
}

// 检查恢复状态
if agent.GetState() == agent.StateRunning {
    log.Println("Agent 已恢复运行")
}
```

### 暂停恢复流程

```
┌──────────────────────────────────────────────────────────────┐
│               暂停恢复流程                            │
└──────────────────────────────────────────────────────────────┘

┌─────────┐          ┌─────────┐
│ Running │          │  Pause  │
└─────────┘          └─────────┘
    │                    ▲
    │            ┌─────────────┐
    │            │ 保存状态    │
    │            │ 检查点保存  │
    │            └─────────────┘
    │                    │
    ▼                    ▼
┌─────────┐          ┌─────────┐
│ Paused │◄───────►│Running  │
└─────────┘          └─────────┘
    │                    │
    ▼                    ▼
┌─────────────┐          │
│ 恢复执行    │          │
│ 加载状态    │          │
│ 检查点恢复  │          │
└─────────────┘          │
    │                    │
    └────────────────────┘
```

### 使用场景

| 场景 | 说明 | 实现方式 |
|------|------|----------|
| **资源管理** | 高负载时暂停低优先级任务 | `agent.Pause(ctx)` |
| **维护更新** | 系统维护时暂停所有 Agent | `host.PauseAllAgents()` |
| **错误恢复** | 遇到错误时暂停，人工介入 | HITL + Pause |
| **成本控制** | 预算超限时暂停执行 | 超时检查 → Pause |

---

## 销毁阶段

### 销毁流程

```
┌──────────────────────────────────────────────────────────────┐
│                Agent 销毁流程                           │
└──────────────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 1. 停止接受  │
              │ 新任务        │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 2. 等待进行  │
              │ 中任务完成  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 3. 保存状态  │
              │ (可选)       │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 4. 释放资源  │
              └─────────────┘
                        │
                        ▼
              ┌─────────────┐
              │ 5. 取消注册  │
              └─────────────┘
```

### 销毁方法

#### 1. 优雅关闭

```go
// 优雅关闭 Agent
err := agent.Shutdown(ctx)
if err != nil {
    log.Printf("关闭失败: %v\n", err)
}

// 等待关闭完成
<-agent.Done()
log.Println("Agent 已优雅关闭")
```

#### 2. 强制销毁

```go
// 强制销毁（立即）
err := agent.Destroy(ctx)
if err != nil {
    log.Printf("销毁失败: %v\n", err)
}

// 不等待进行中任务完成
log.Println("Agent 已强制销毁")
```

#### 3. 超时销毁

```go
// 设置超时自动销毁
timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

go func() {
    <-timeoutCtx.Done()
    log.Println("超时，开始销毁 Agent")
    agent.Shutdown(context.Background())
}()

// 正常执行
response, err := agent.Run(timeoutCtx, "长任务...")
```

### 资源清理

```go
// BaseAgent 清理实现
func (a *BaseAgent) cleanup(ctx context.Context) error {
    var errs []error

    // 1. 关闭内存
    if err := a.memory.Close(); err != nil {
        errs = append(errs, fmt.Errorf("memory close failed: %w", err))
    }

    // 2. 释放缓存
    if err := a.cache.Close(); err != nil {
        errs = append(errs, fmt.Errorf("cache close failed: %w", err))
    }

    // 3. 取消订阅
    if err := a.eventBus.UnsubscribeAll(a.id); err != nil {
        errs = append(errs, fmt.Errorf("unsubscribe failed: %w", err))
    }

    // 4. 释放工具
    if err := a.tools.Close(); err != nil {
        errs = append(errs, fmt.Errorf("tools close failed: %w", err))
    }

    // 5. 标记销毁
    a.state = StateDestroyed

    // 返回所有错误
    if len(errs) > 0 {
        return fmt.Errorf("cleanup errors: %v", errs)
    }
    return nil
}
```

---

## 状态管理

### 状态查询

```go
// 获取当前状态
state := agent.GetState()

// 状态判断
switch state {
case agent.StateCreated:
    log.Println("Agent 已创建")
case agent.StateInitialized:
    log.Println("Agent 已初始化")
case agent.StateRunning:
    log.Println("Agent 正在运行")
case agent.StatePaused:
    log.Println("Agent 已暂停")
case agent.StateCompleted:
    log.Println("Agent 已完成")
case agent.StateFailed:
    log.Println("Agent 执行失败")
case agent.StateDestroyed:
    log.Println("Agent 已销毁")
}
```

### 状态监听

```go
// 监听状态变化
agent.SubscribeStateChanges(func(oldState, newState AgentState) {
    log.Infof("状态变化: %s -> %s", oldState, newState)

    // 根据状态变化采取行动
    switch newState {
    case agent.StateFailed:
        // 失理失败
        notifyFailure(agent)
    case agent.StateCompleted:
        // 处理完成
        saveResult(agent)
    }
})
```

### 状态转换验证

```go
// 验证状态转换是否合法
func (a *BaseAgent) validateStateTransition(newState AgentState) error {
    allowedTransitions := map[AgentState][]AgentState{
        StateCreated:    {StateInitialized, StateDestroyed},
        StateInitialized: {StateRunning, StateDestroyed},
        StateRunning:     {StatePaused, StateCompleted, StateFailed},
        StatePaused:     {StateRunning, StateDestroyed},
        StateCompleted:   {StateDestroyed},
        StateFailed:      {StateDestroyed},
        StateDestroyed:   {}, // 没有可转换的状态
    }

    allowed, ok := allowedTransitions[a.state]
    if !ok {
        return fmt.Errorf("invalid current state: %v", a.state)
    }

    for _, state := range allowed {
        if state == newState {
            return nil // 转换合法
        }
    }

    return fmt.Errorf("invalid state transition: %s -> %s", a.state, newState)
}
```

---

## 最佳实践

### 1. 生命周期管理

```go
// ✅ 正确：使用 defer 确保清理
func processWithAgent(ctx context.Context) error {
    agent, err := createAgent()
    if err != nil {
        return err
    }
    defer agent.Shutdown(context.Background())

    // 使用 agent
    return agent.Run(ctx, "task")
}

// ❌ 错误：忘记清理
func processWithAgent(ctx context.Context) error {
    agent, err := createAgent()
    if err != nil {
        return err
    }
    // 忘略 defer，可能导致资源泄露
    return agent.Run(ctx, "task")
}
```

### 2. 错误处理

```go
// ✅ 正确：处理失败状态
response, err := agent.Run(ctx, "task")
if err != nil {
    if agent.GetState() == agent.StateFailed {
        // 记录失败信息
        logFailure(agent, err)
        // 尝试恢复
        return retryOrFail(agent)
    }
    return err
}

// ❌ 错误：忽略状态
response, err := agent.Run(ctx, "task")
if err != nil {
    // 没有检查失败状态，可能丢失重要信息
    return err
}
```

### 3. 并发安全

```go
// ✅ 正确：保护状态访问
func (a *BaseAgent) GetState() AgentState {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.state
}

func (a *BaseAgent) SetState(state AgentState) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.state = state
}

// ❌ 错误：无保护访问
func (a *BaseAgent) GetState() AgentState {
    return a.state // 并发不安全
}
```

### 4. 资源管理

```go
// ✅ 正确：及时释放资源
func useAgent(ctx context.Context) error {
    agent, err := createAgent()
    if err != nil {
        return err
    }

    // 使用 defer 确保释放
    defer func() {
        if err := agent.Shutdown(ctx); err != nil {
            log.Printf("shutdown failed: %v\n", err)
        }
    }()

    return agent.Run(ctx, "task")
}

// ❌ 错误：延迟释放
func useAgent(ctx context.Context) error {
    agent, err := createAgent()
    if err != nil {
        return err
    }

    response, err := agent.Run(ctx, "task")
    // 错误处理中可能跳过释放
    if err != nil {
        return err
    }

    return agent.Shutdown(ctx)
}
```

### 5. 状态监听

```go
// ✅ 正确：监听关键状态变化
agent.SubscribeStateChanges(func(oldState, newState AgentState) {
    // 只关心关键状态
    if newState == agent.StateFailed {
        notifyAdmin(agent)
    }
})

// ❌ 错误：监听所有状态变化
agent.SubscribeStateChanges(func(oldState, newState AgentState) {
    // 记录所有状态变化，可能产生大量日志
    log.Infof("state change: %s -> %s", oldState, newState)
})
```

---

## 相关文档

- 📘 [Agent 概览](overview.md) - Agent 系统概览
- 📘 [Agent 类型](types.md) - 各种 Agent 类型
- 📘 [Agent API 参考](api.md) - 完整 API 文档
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南
- 📘 [配置指南](../../configuration/CONFIGURATION.md) - 配置说明

---

**Made with ❤️ by AgentFramework Team**
