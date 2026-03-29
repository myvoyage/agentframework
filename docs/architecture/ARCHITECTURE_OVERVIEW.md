# AgentFramework 架构概览

> 基于 OpenClaw 架构设计的 Go 语言实现
> 版本：v2.x · 最后更新：2026-03-29

---

## 目录

- [整体架构](#整体架构)
- [Gateway 层](#gateway-层)
- [Agent 运行时](#agent-运行时)
- [技能系统](#技能系统)
- [渠道层](#渠道层)
- [数据流](#数据流)
- [关键设计决策](#关键设计决策)

---

## 整体架构

AgentFramework 采用四层架构，参照 [OpenClaw 架构设计](https://www.cnblogs.com/tangshiye/p/19642495)以 Go 语言实现：

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐    │
│  │  Desktop App  │  │  CLI (afcli) │  │  TUI (aftui)   │    │
│  │ (Wails+Vue3) │  │   cmd/cli/   │  │   cmd/tui/     │    │
│  └──────────────┘  └──────────────┘  └────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       Gateway Layer                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  gateway/server.go  (单端口 WS + HTTP 复用)           │   │
│  │  ├── WebSocketHandler  /                             │   │
│  │  └── HTTPServer        /v1/* /health /tools/*        │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Agent Runtime Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐   │
│  │    Host     │  │ Lane Queue  │  │  Execution Loop   │   │
│  │  host.go    │  │lane_queue.go│  │  execution.go     │   │
│  └─────────────┘  └─────────────┘  └───────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐   │
│  │ Workflow DAG│  │  Scheduler  │  │  Realtime Agent   │   │
│  │workflow_dag │  │ scheduler/  │  │realtime_agent.go  │   │
│  └─────────────┘  └─────────────┘  └───────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴──────────┐
                    ▼                    ▼
┌──────────────────────┐   ┌───────────────────────────────┐
│    Skills Layer       │   │         Channel Layer          │
│  skills/markdown/    │   │  messaging/channel_manager.go  │
│  ├── discoverer.go   │   │  ├── Telegram                  │
│  ├── manager.go      │   │  ├── Lark (飞书)               │
│  └── parser.go       │   │  ├── WeCom (企业微信)          │
│  skills/registry.go  │   │  ├── QQ / Discord / Slack      │
│  skills/base_skill.go│   │  └── WebChat / CLI             │
└──────────────────────┘   └───────────────────────────────┘
```

---

## Gateway 层

**文件**：`gateway/`

Gateway 是整个框架的网络入口。单个端口（默认 18640）同时承载 WebSocket 和 HTTP 流量：

```
gateway/
├── server.go     # 主服务器：监听端口、路由分发
├── service.go    # 业务逻辑：会话管理、消息派发
├── config.go     # 配置结构
├── protocol.go   # WebSocket 协议帧定义
├── http.go       # HTTP 处理器
└── websocket.go  # WebSocket 处理器
```

### WebSocket 协议帧

三种帧类型，首帧必须是 `connect`：

```json
// 连接帧（首帧必须）
{"type":"connect","id":"<str>","params":{"auth":{"token":"..."},"deviceIdentity":{}}}

// 请求/响应
{"type":"req","id":"<str>","method":"<str>","params":{},"idempotencyKey":"<str>"}
{"type":"res","id":"<str>","ok":true,"payload":{}}

// 服务端事件推送
{"type":"event","event":"<str>","payload":{},"seq":1}
```

- 副作用方法（`send`、`agent`）必须携带 `idempotencyKey` 防重放
- 非 JSON 或首帧不是 `connect` → 立即关闭连接

### HTTP 路由

| 路径 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /status` | 服务状态 |
| `GET /v1/agents` | 列出 Agent |
| `POST /v1/agents/{name}/run` | 执行 Agent |
| `GET /tools/list` | 列出可用工具 |

---

## Agent 运行时

### Host — 中央容器

**文件**：`agent/host.go`

`Host` 是整个运行时的中央容器，持有所有组件的引用：

```go
type Host struct {
    cfg             *HostConfig
    modelFactory    ModelFactory          // 模型工厂
    threadStore     ThreadStore           // 对话历史存储
    toolRegistry    map[string]tool.BaseTool
    monitorMgr      *MonitorManager
    channelMgr      *messaging.ChannelManager
    scheduler       interface{}           // *scheduler.Scheduler
    heartbeat       interface{}           // *heartbeat.HeartbeatService
    taskManager     interface{}           // *async.TaskManager
    tokenCompressor interface{}           // *token.MessageCompressor
    agents          map[string]Agent
    workflows       map[string]Workflow
    middlewares     map[string]AgentMiddleware
    service         *AgentService
}
```

### Lane Queue — 会话串行队列

**文件**：`agent/lane_queue.go`

Lane Queue 是框架保证消息顺序、消除竞态条件的核心机制：

```
SessionKey = workspace:channel:userId
  例：default:telegram:user123
  例：cron:scheduler:0  （特殊 lane，并行运行）
```

- 同一 SessionKey 的任务**串行**执行，天然避免竞态
- 不同 SessionKey 的任务**并行**执行，充分利用并发
- 支持优先级、超时、幂等键去重
- 内置背压机制

```go
// 提交任务到对应会话队列
queue.Enqueue(sessionKey, task, opts...)

// 特殊 lane（cron / subagent）可并行
queue.EnqueueParallel(task, opts...)
```

### ReAct 执行循环

**文件**：`agent/execution.go`

实现 `模型 → 工具 → 模型` 的标准 ReAct 循环：

```
┌─────────────────────────────────────────────────┐
│                  ReAct Loop                      │
│                                                  │
│  [组装上下文]                                    │
│       ↓                                          │
│  [模型推理] ←─────────────────────┐              │
│       ↓                           │              │
│  是否有工具调用?                  │              │
│  ├── 是 → [执行工具(并行/串行)]   │              │
│  │         [收集工具结果] ────────┘              │
│  └── 否 → [输出最终响应]                         │
│                                                  │
│  配置：MaxIterations=10, Timeout=5min            │
│        ToolTimeout=30s                           │
└─────────────────────────────────────────────────┘
```

每次运行分配唯一 `RunID`（UUID），通过 `EventChan` 发射可观测事件：

```
started → thinking → tool_call → tool_result → ... → completed
```

### Workflow DAG

**文件**：`agent/workflow_dag.go`

并发安全的有向无环图工作流引擎：

- 节点/边用 `sync.RWMutex` 保护，执行前快照防止并发修改
- `inDegree` 和 `processed` 用 `sync.Map` 实现无锁操作
- 支持条件路由（边上附加条件表达式）
- 支持断点续传（检查点存储）

### Scheduler

**文件**：`agent/scheduler/`

- `scheduler.go` — 主调度器，管理周期任务
- `cron.go` — Cron 表达式解析与执行

---

## 技能系统

**文件**：`agent/skills/`

### Markdown Skill 驱动

```
.skills/
└── my_tool/
    └── SKILL.md      # Skill 定义文件
```

`SKILL.md` 结构：

```markdown
---
name: my_tool
description: 简短描述（约 97 字符，启动时加载）
version: "1.0"
parameters:
  query:
    type: string
    description: 查询内容
    required: true
permissions:
  - read_files
---

# 详细使用说明

（激活时才注入到上下文——渐进式披露）
```

### 加载流程

```
启动时：扫描所有 SKILL.md → 只读取 name + description（轻量）
激活时：读取完整 SKILL.md → 注入系统提示词
热重载：文件变更后下次对话自动生效
```

### Skill 目录优先级

1. `agent/skills/bundled/` — 内置 Skill（最高优先级）
2. `~/.agentframework/skills/` — 用户级 Skill
3. `./.skills/` — 项目级 Skill

### Skill 接口

```go
type Skill interface {
    Info(ctx context.Context) (*schema.ToolInfo, error)
    Invoke(ctx context.Context, input string) (string, error)
    IsEnabled(ctx context.Context) bool
    GetMetadata(ctx context.Context) SkillMetadata
}
```

---

## 渠道层

**文件**：`agent/messaging/`

统一消息格式，适配器模式接入不同平台：

| 渠道 | 文件 | Webhook 端口 |
|------|------|-------------|
| Telegram | `pkg/workspace/telegram.go` | `:8087/telegram` |
| 飞书 (Lark) | `pkg/workspace/lark.go` | `:8089/lark` |
| 企业微信 | `pkg/workspace/wechat.go` | `:8090/wechat` |
| QQ | `pkg/workspace/qq.go` | `:8088/qq` |
| Discord | `pkg/workspace/discord.go` | — |
| Slack | `pkg/workspace/slack.go` | — |
| WebChat | `pkg/workspace/webchat.go` | — |
| CLI | `agent/messaging/internal_channel.go` | — |

### 消息路由流程

```
外部消息（Webhook）
       ↓
  ChannelManager
       ↓
  NewSessionKey(workspace, channel, userID)
       ↓
  LaneQueue.Enqueue(sessionKey, task)
       ↓
  ExecutionLoop.Run(context)
       ↓
  回复到原渠道
```

---

## 数据流

### 完整请求处理流程

```
[用户/渠道]
    │ WebSocket / Webhook / CLI
    ▼
[Gateway / ChannelManager]
    │ 解析消息，提取 sessionKey
    ▼
[LaneQueue]
    │ 排队，等待当前会话空闲
    ▼
[ExecutionLoop]
    │ 1. 组装上下文（SOUL.md + MEMORY.md + 历史）
    │ 2. 调用 ChatModel
    │ 3. 有工具调用 → 执行工具 → 循环
    │ 4. 生成最终回复
    ▼
[回复渠道]
    │ 流式 / 批量输出
    ▼
[用户]
```

### 上下文组装（5 个来源）

1. 系统提示词（Agent 身份定义）
2. Workspace 文件（`SOUL.md`、`AGENTS.md`）
3. 记忆文件（`MEMORY.md` + `memory/YYYY-MM-DD.md`）
4. 会话历史（当前上下文窗口内的对话）
5. 工具执行结果（前序步骤的返回值）

---

## 关键设计决策

### 为什么用 Lane Queue 而不是直接 goroutine？

直接为每条消息启动 goroutine 会导致同一用户的消息乱序执行（竞态）。Lane Queue 确保同一会话串行，不同会话并行，兼顾正确性和性能。

### 为什么 Skill 用 Markdown 而不是代码？

Markdown SKILL.md 文件可以被 Agent 自己读取和修改（自写技能），同时对人类友好，不需要编译。渐进式披露（只有激活的 Skill 才注入全文）控制了 token 消耗。

### 为什么 WebSocket 和 HTTP 共用一个端口？

简化部署配置，减少防火墙规则，方便在 Nginx 后代理。通过 HTTP Upgrade 头区分两种连接。

### 为什么 Workflow DAG 用 sync.Map？

DAG 的 `inDegree` 和 `processed` 字段在并发执行时被多个 goroutine 读写，`sync.Map` 比加锁的普通 map 在高读低写场景下性能更好，且天然避免了 2026-03-26 发现的竞态 bug。

---

## 相关文档

- [快速上手](../quickstart/QUICKSTART.md)
- [配置指南](../configuration/CONFIGURATION.md)
- [API 参考](../api/API.md)
- [Skill 开发指南](../SKILL_DEVELOPMENT.md)
- [渠道集成](../CHANNEL_INTEGRATION.md)
