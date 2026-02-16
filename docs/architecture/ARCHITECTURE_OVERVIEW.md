# 架构概览

> **AgentFramework 系统架构全景**
> **版本**: v2.0.0
> **最后更新**: 2026-02-15

---

## 📋 目录

- [架构原则](#架构原则)
- [分层设计](#分层设计)
- [核心组件](#核心组件)
- [Agent 类型系统](#agent-类型系统)
- [工作流引擎](#工作流引擎)
- [技能系统](#技能系统)
- [存储系统](#存储系统)
- [监控遥测](#监控遥测)
- [安全沙箱](#安全沙箱)
- [设计模式](#设计模式)
- [数据流](#数据流)
- [扩展机制](#扩展机制)

---

## 架构原则

### SOLID 原则

| 原则 | 应用 | 说明 |
|------|------|------|
| **S** - 单一职责 | ⭐⭐⭐⭐⭐ | 每个组件职责明确，修改影响范围小 |
| **O** - 开闭原则 | ⭐⭐⭐⭐⭐ | 通过接口支持扩展，避免修改现有代码 |
| **L** - 里氏替换 | ⭐⭐⭐⭐⭐ | 接口实现可完全替换 |
| **I** - 接口隔离 | ⭐⭐⭐⭐☆ | 接口专一性好，少数"胖接口"可优化 |
| **D** - 依赖倒置 | ⭐⭐⭐⭐⭐ | 核心模块依赖抽象接口 |

### 其他原则

- **KISS**: 保持简单直观，避免过度设计
- **DRY**: 代码复用性高，公共功能抽象到 pkg 层
- **YAGNI**: 只实现当前需要的功能，没有未来预留
- **性能优先**: 内置连接池、缓存、并发优化

---

## 分层设计

### 整体架构图

```
┌────────────────────────────────────────────────────────────┐
│                     Application Layer                  │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │ Desktop App │  │  CLI Tools   │  │  HTTP API   │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                      Framework Layer                   │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │    Host     │  │   Agent     │  │  Workflow    │ │
│  │   Manager   │  │   Manager   │  │   Engine     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │  Skill      │  │  Model      │  │ Collab      │ │
│  │  Library    │  │  Factory    │  │  System     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                     Capability Layer                 │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │  Tool       │  │ Middleware  │  │  Event Bus  │ │
│  │ Registry    │  │   Chain     │  │              │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                 │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │Checkpoint   │  │  Sandbox    │  │ Observability│ │
│  │  Store      │  │  Manager    │  │              │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### 应用层 (Application Layer)

**职责**: 提供用户交互界面和 API 入口

| 组件 | 说明 | 技术栈 |
|------|------|--------|
| **Desktop App** | 桌面应用界面 | Wails + Vue 3 |
| **CLI Tools** | 命令行工具 | Go cobra |
| **HTTP API** | RESTful API | Go Echo |

### 框架层 (Framework Layer)

**职责**: 提供 AI Agent 开发核心框架

| 模块 | 核心类 | 说明 |
|------|--------|------|
| **Host** | [Host](../../agent/host.go) | 中央容器，管理所有组件 |
| **Agent** | [Agent](../../agent/agent.go) | 代理接口和实现 |
| **Workflow** | [Workflow](../../agent/workflow.go) | 工作流引擎 |
| **Skill** | [Skill](../../agent/skills/) | 技能系统 |
| **Model** | [ModelFactory](../../agent/model_factory.go) | 模型管理 |
| **Collab** | [Collaboration](../../agent/collaboration/) | 协作系统 |

### 能力层 (Capability Layer)

**职责**: 提供可插拔的能力和扩展机制

| 组件 | 说明 | 文件位置 |
|------|------|----------|
| **Tool Registry** | 工具注册和管理 | [tool_registry.go](../../agent/dynamic_tool_registry.go) |
| **Middleware Chain** | 中间件链 | [middleware.go](../../agent/middleware.go) |
| **Event Bus** | 事件总线 | [event_bus.go](../../agent/event_bus.go) |

### 基础设施层 (Infrastructure Layer)

**职责**: 提供底层服务和存储

| 组件 | 说明 | 文件位置 |
|------|------|----------|
| **Checkpoint Store** | 状态持久化 | [checkpoint.go](../../agent/checkpoint.go) |
| **Sandbox Manager** | 沙箱执行环境 | [sandbox/](../../pkg/tools/sandbox/) |
| **Observability** | 可观测性 | [monitor.go](../../agent/monitor.go) |

---

## 核心组件

### Host 系统

**文件**: [agent/host.go](../../agent/host.go)

```go
type Host struct {
    cfg             *HostConfig
    configMgr       ConfigManager
    modelFactory    ModelFactory
    threadStore     ThreadStore
    toolRegistry    map[string]tool.BaseTool
    monitorMgr      *MonitorManager
    pluginMgr       PluginManager
    channelMgr      *messaging.ChannelManager
    scheduler      interface{}
    heartbeat      interface{}
    taskManager     interface{}
    tokenCompressor interface{}
    agents         map[string]Agent
    workflows      map[string]Workflow
    middlewares    map[string]AgentMiddleware
    service        *AgentService
}
```

**职责**:
- 🎯 管理所有 Agent 生命周期
- 🔄 管理所有 Workflow 生命周期
- 🛠️ 管理所有 Skill 注册
- 📊 管理监控、日志、性能指标
- 🔧 管理配置更新

### Agent 系统

**接口定义**: [agent/agent.go](../../agent/agent.go)

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

**支持的类型**:

| 类型 | 说明 | 文件 |
|------|------|------|
| **ChatAgent** | 基础对话代理 | [chat_agent.go](../../agent/chat_agent.go) |
| **ReActAgent** | 推理-行动代理 | [react_agent.go](../../agent/react_agent.go) |
| **HumanAgent** | 人工介入代理 | [human_agent.go](../../agent/hitl.go) |
| **WorkerAgent** | 专业工作代理 | [swe_agent.go](../../agent/swe_agent.go) |
| **EdgeAgent** | 边缘代理 | [edge_agent.go](../../agent/edge_agent.go) |

### Workflow 引擎

**文件**: [agent/workflow.go](../../agent/workflow.go)

```go
type Workflow interface {
    // 基本信息
    Name() string
    Type() WorkflowType
    Version() string

    // 执行接口
    Run(ctx context.Context, input string, opts ...Option) (*schema.Message, error)

    // 节点管理
    AddNode(node Node)
    RemoveNode(nodeID string)
    GetNode(nodeID string) (Node, bool)

    // 边管理
    AddEdge(from, to string, condition ...string)
    RemoveEdge(from, to string)
}
```

**支持的工作流类型**:

| 类型 | 说明 | 使用场景 |
|------|------|---------|
| **Sequential** | 顺序执行 | 简单任务流 |
| **Parallel** | 并行执行 | 独立任务并发 |
| **DAG** | 有向无环图 | 复杂依赖关系 |
| **Routing** | 条件路由 | 动态分支选择 |
| **Planning** | 规划工作流 | 自动任务分解 |
| **Graph** | 通用图 | 任意拓扑结构 |

### Skill 系统

**核心接口**: [agent/skills/types.go](../../agent/skills/types.go)

```go
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

**核心组件**:

| 组件 | 文件 | 职责 |
|------|------|------|
| **SkillRegistry** | [registry.go](../../agent/skills/registry.go) | 技能注册表 |
| **SkillLoader** | [loader.go](../../agent/skills/loader.go) | 技能加载器 |
| **SkillsPool** | [pool.go](../../agent/skills/skills_pool.go) | 技能连接池 |
| **SkillsCache** | [skills_cache.go](../../agent/skills/skills_cache.go) | 技能缓存 |

### Collaboration 系统

**核心组件**: [agent/collaboration/](../../agent/collaboration/)

```go
type AgentTeam struct {
    name        string
    description string
    members     []*TeamMember
    bus         *MessageBus
    scheduler   TaskScheduler
    router      *IntelligentRouter
    mu          sync.RWMutex
    ctx         context.Context
    cancel      context.CancelFunc
    running     bool
}
```

**协作模式**:

| 模式 | 说明 | 使用场景 |
|------|------|---------|
| **Single** | 单代理执行 | 简单任务 |
| **Parallel** | 并行执行 | 独立任务 |
| **Sequential** | 顺序执行 | 依赖任务 |
| **Consensus** | 共识决策 | 多代理投票 |

---

## 设计模式

### 1. 主机模式 (Host Pattern)

**描述**: Host 作为中央容器，管理所有组件的生命周期

**优点**:
- ✅ 统一的生命周期管理
- ✅ 简化的依赖注入
- ✅ 方便的配置管理

**示例**: [agent/host.go](../../agent/host.go#L71)

```go
func NewHost(ctx context.Context, cfg *HostConfig, mf ModelFactory, tr map[string]tool.BaseTool, opts ...HostOption) (*Host, error) {
    // 创建组件
    host := &Host{
        cfg:          cfg,
        configMgr:    NewConfigManager(cfg),
        modelFactory: mf,
        toolRegistry: tr,
        monitorMgr:   NewMonitorManager(...),
        pluginMgr:    NewPluginManager(),
        agents:       make(map[string]Agent),
        workflows:    make(map[string]Workflow),
        middlewares: make(map[string]AgentMiddleware),
    }

    // 初始化组件
    host.initThreadStore(ctx)
    host.registerDefaultMiddlewares()
    host.buildAgents(ctx, tr)
    host.buildWorkflows(ctx)

    return host, nil
}
```

### 2. 工厂模式 (Factory Pattern)

**描述**: 使用工厂创建各种组件实例

**优点**:
- ✅ 封装创建逻辑
- ✅ 支持参数验证
- ✅ 方便的缓存管理

**示例**: [agent/model_factory.go](../../agent/model_factory.go#L226)

```go
func NewDefaultModelFactory(cfg DefaultModelFactoryConfig) ModelFactory {
    // 预处理配置
    preprocessed := &preprocessedModelConfig{
        configs: make(map[string]ModelConfig, len(cfg.Models)),
    }

    // 返回工厂函数
    return func(ctx context.Context, modelName string) (ChatModel, error) {
        // 查找配置
        modelCfg, ok := preprocessed.configs[modelName]
        if !ok {
            return nil, fmt.Errorf("model not found")
        }

        // 创建模型
        return createModel(ctx, modelCfg)
    }
}
```

### 3. 策略模式 (Strategy Pattern)

**描述**: 算法可在运行时动态切换

**优点**:
- ✅ 算法独立
- ✅ 易于扩展
- ✅ 运行时切换

**示例**: Token 压缩策略

```go
type CompressionStrategy int

const (
    CompressionStrategyTruncate CompressionStrategy = iota
    CompressionStrategySummarize
    CompressionStrategyHybrid
)

func (c *MessageCompressor) CompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
    switch c.strategy {
    case CompressionStrategyTruncate:
        return c.truncate(messages, targetTokens)
    case CompressionStrategySummarize:
        return c.summarize(ctx, messages, targetTokens)
    case CompressionStrategyHybrid:
        return c.hybrid(ctx, messages, targetTokens)
    }
}
```

### 4. 观察者模式 (Observer Pattern)

**描述**: 组件间通过事件总线通信

**优点**:
- ✅ 松耦合
- ✅ 易扩展
- ✅ 异步通信

**示例**: [agent/event_bus.go](../../agent/event_bus.go)

```go
type EventBus struct {
    subscribers map[string][]chan *Message
    mu           sync.RWMutex
}

func (eb *EventBus) Subscribe(topic string) chan *Message {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    ch := make(chan *Message, 100)
    eb.subscribers[topic] = append(eb.subscribers[topic], ch)
    return ch
}

func (eb *EventBus) Publish(topic string, msg *Message) {
    eb.mu.RLock()
    subscribers := eb.subscribers[topic]
    eb.mu.RUnlock()

    for _, ch := range subscribers {
        select {
        case ch <- msg:
        default:
            // 非阻塞
        }
    }
}
```

### 5. 责任链模式 (Chain of Responsibility)

**描述**: 请求通过中间件链处理

**优点**:
- ✅ 灵活的处理流程
- ✅ 易于添加新处理
- ✅ 职责分离

**示例**: [agent/middleware.go](../../agent/middleware.go)

```go
type AgentMiddleware func(ctx context.Context, req *Request, next HandlerFunc) (*Response, error)

func (h *Host) Use(middleware string) AgentMiddleware {
    h.middlewares[middleware] = h.createMiddleware(middleware)
}

func (h *Host) runMiddlewares(ctx context.Context, req *Request) (*Response, error) {
    // 构建中间件链
    chain := h.buildMiddlewareChain()

    // 执行链
    return chain(ctx, req, func(ctx context.Context, req *Request) (*Response, error) {
        return h.handler(ctx, req)
    })
}
```

### 6. 状态模式 (State Pattern)

**描述**: 工作流节点根据状态执行不同行为

**优点**:
- ✅ 状态明确
- ✅ 易于扩展
- ✅ 避免条件判断

**示例**: 工作流节点状态

```go
type NodeState int

const (
    StateIdle NodeState = iota
    StateRunning
    StateCompleted
    StateFailed
    StateSkipped
)

func (n *WorkflowNode) Execute(ctx context.Context) error {
    switch n.state {
    case StateIdle:
        return n.executeIdle(ctx)
    case StateRunning:
        return n.executeRunning(ctx)
    case StateCompleted:
        return nil
    case StateFailed:
        return fmt.Errorf("node failed")
    }
}
```

---

## 数据流

### 请求处理流程

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Client  │ -> │  Host    │ -> │ Agent   │ -> │ Skills  │ -> │ Tools   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
                      │                │                │                │
                      ▼                ▼                ▼
                   ┌─────────┐    ┌─────────┐    ┌─────────┐
                   │Monitor  │    │ Cache   │    │  Event  │
                   │ Logger  │    │         │    │  Bus    │
                   └─────────┘    └─────────┘    └─────────┘
```

### 工作流执行流程

```
┌───────────────────────────────────────────────────────────┐
│                   Workflow Engine                   │
│  ┌─────────────┐    ┌─────────────┐              │
│  │  DAG Parser │ -> │  Scheduler  │              │
│  └─────────────┘    └─────────────┘              │
│         │                   │                         │
│         ▼                   ▼                         │
│  ┌─────────────┐    ┌─────────────┐              │
│  │  Node Queue │ -> │  Executor   │              │
│  └─────────────┘    └─────────────┘              │
│         │                   │                         │
│         ▼                   ▼                         │
│  ┌─────────────┐    ┌─────────────┐              │
│  │Checkpoint   │ -> │  State      │              │
│  │   Store     │    │  Manager    │              │
│  └─────────────┘    └─────────────┘              │
└───────────────────────────────────────────────────────────┘
```

### 协作系统流程

```
┌───────────────────────────────────────────────────────────┐
│                   Collaboration System                 │
│  ┌─────────────┐    ┌─────────────┐              │
│  │  Agent Team │ -> │  Message Bus │              │
│  └─────────────┘    └─────────────┘              │
│         │                   │                         │
│         ▼                   ▼                         │
│  ┌─────────────┐    ┌─────────────┐              │
│  │  Router     │ -> │  Scheduler   │              │
│  └─────────────┘    └─────────────┘              │
│         │                   │                         │
│         ▼                   ▼                         │
│  ┌─────────────┐    ┌─────────────┐              │
│  │  Agent 1   │    │  Agent 2    │              │
│  └─────────────┘    └─────────────┘              │
└───────────────────────────────────────────────────────────┘
```

---

## 扩展机制

### 插件系统

**支持类型**:
- 🔌 **Skill 插件**: 扩展 Agent 能力
- 🛠️ **Tool 插件**: 提供工具支持
- 🔄 **Middleware 插件**: 拦截处理流程
- 📊 **Monitor 插件**: 监控和日志

**示例**: 注册自定义 Skill

```go
// 1. 实现接口
type MySkill struct {
    *skills.BaseSkill
}

func (s *MySkill) Invoke(ctx context.Context, input string) (string, error) {
    // 实现逻辑
    return result, nil
}

func (s *MySkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_skill",
        Desc: "My custom skill",
    }, nil
}

// 2. 注册到 Host
host.RegisterSkill("my_skill", &MySkill{})
```

### 配置扩展

**支持配置**:
- 📝 **静态配置**: YAML 文件
- 🔄 **动态配置**: 运行时更新
- 🔐 **环境变量**: 环境变量注入
- 🌐 **HTTP 配置**: 远程配置中心

**示例**: 添加自定义配置

```go
type MyConfig struct {
    FeatureEnabled bool `yaml:"feature_enabled"`
    Threshold      int  `yaml:"threshold"`
}

// 在 HostConfig 中添加
type HostConfig struct {
    // ... 其他字段
    Custom MyConfig `yaml:"custom"`
}
```

### 模型扩展

**支持新模型**:

```go
// 1. 定义模型配置
cfg := ModelConfig{
    Type:    "custom",
    Model:   "my-model",
    BaseURL: "http://localhost:8080",
    Enabled: true,
}

// 2. 创建模型工厂
factory := func(ctx context.Context, modelName string) (ChatModel, error) {
    return &MyChatModel{
        config: cfg,
    }, nil
}

// 3. 注册到 Host
host.RegisterModelFactory("custom", factory)
```

---

## 相关文档

- 📘 [Host API](../api/host.md) - Host 接口文档
- 📘 [Agent API](../api/agent.md) - Agent 接口文档
- 📘 [Workflow API](../api/workflow.md) - Workflow 接口文档
- 📘 [Skill API](../api/skills.md) - Skill 接口文档
- 📘 [组件详解](../components/COMPONENTS.md) - 组件详细说明

---

**Made with ❤️ by AgentFramework Team**
