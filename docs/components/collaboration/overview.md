# Collaboration 系统概览

> **AgentFramework Collaboration 组件文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [协作模式](#协作模式)
- [核心组件](#核心组件)
- [Agent Team](#agent-team)
- [任务调度](#任务调度)
- [消息通信](#消息通信)
- [快速开始](#快速开始)

---

## 系统简介

Collaboration 系统实现多代理协作，支持复杂的团队协作模式。

### 核心职责

| 职责 | 说明 |
|------|------|
| **团队管理** | 管理代理团队和成员 |
| **任务调度** | 智能任务分发和调度 |
| **消息路由** | 代理间消息通信 |
| **共识决策** | 多代理决策协调 |
| **状态同步** | 团队状态同步 |

### 技术特点

- ✅ **多种模式**: Single、Parallel、Sequential、Consensus
- ✅ **智能路由**: 基于能力的任务分发
- ✅ **事件驱动**: 基于消息的异步通信
- ✅ **动态扩缩**: 运行时添加/移除代理
- ✅ **容错机制**: 代理失败自动恢复
- ✅ **可观测性**: 完整的监控和日志

---

## 协作模式

### 模式概览

| 模式 | 说明 | 复杂度 | 适用场景 |
|------|------|--------|---------|
| **Single** | 单代理执行 | ⭐ | 简单任务 |
| **Parallel** | 并行执行 | ⭐⭐ | 独立任务 |
| **Sequential** | 顺序执行 | ⭐⭐ | 依赖任务 |
| **Consensus** | 共识决策 | ⭐⭐⭐⭐ | 关键决策 |

### 模式对比

```
┌──────────────────────────────────────────────────────────────┐
│              Collaboration 模式对比                       │
├──────────────────────────────────────────────────────────────┤
│ 模式        │代理│并发│共识│容错│典型场景    │
├──────────────────────────────────────────────────────────────┤
│ Single      │  1  │ ❌  │ ❌  │ ❌  │简单任务    │
│ Parallel    │  N  │ ✅  │ ❌  │ ✅  │独立任务    │
│ Sequential  │  N  │ ❌  │ ❌  │ ✅  │依赖任务    │
│ Consensus   │  N  │ ✅  │ ✅  │ ✅  │关键决策    │
└──────────────────────────────────────────────────────────────┘
```

---

## 核心组件

### 系统架构

```
┌──────────────────────────────────────────────────────────────┐
│              Collaboration 系统架构                     │
└──────────────────────────────────────────────────────────────┘

┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│   Agent    │          │  Message   │          │   Router    │
│    Team     │────────▶│    Bus     │────────▶│  (智能路由)  │
└─────────────┘          └─────────────┘          └─────────────┘
        │                        │                        │
        └────────────┬────────────┴────────────┘
                     │
                     ▼
          ┌──────────────────────────┐
          │    Collaboration      │
          │    Runtime          │
          └──────────────────────────┘
                     │
         ┌───────────┼───────────┬───────────┐
         ▼            ▼            ▼           ▼
┌─────────────┐┌─────────────┐┌─────────────┐┌─────────────┐
│ Scheduler  ││ Consensus   ││ State      ││ Monitor    │
│ (任务调度) ││ (共识机制)  ││ Manager    ││ (监控)     │
└─────────────┘└─────────────┘└─────────────┘└─────────────┘
```

### 组件说明

#### 1. Agent Team

**功能**: 代理团队管理

**职责**:
- 管理团队成员
- 定义角色和权限
- 处理成员加入/离开
- 维护团队状态

#### 2. Message Bus

**功能**: 消息通信总线

**职责**:
- 代理间消息传递
- 主题订阅/发布
- 消息广播
- 消息持久化

#### 3. Router

**功能**: 智能任务路由

**职责**:
- 分析任务需求
- 评估代理能力
- 选择最优代理
- 负载均衡

---

## Agent Team

### 团队结构

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

type TeamMember struct {
    ID       string
    Agent    Agent
    Role     string
    Skills   []string
    Status   MemberStatus
    Metadata map[string]interface{}
}
```

### 团队配置

**YAML 配置**:

```yaml
teams:
  - name: "research_team"
    description: "多代理研究团队"

    # 协作模式
    mode: "consensus"    # single, parallel, sequential, consensus

    # 成员配置
    members:
      - id: "researcher_1"
        agent: "worker"
        role: "researcher"
        skills: ["web_search", "data_collection"]

      - id: "analyst_1"
        agent: "worker"
        role: "analyst"
        skills: ["data_analysis", "report_generation"]

      - id: "reviewer_1"
        agent: "worker"
        role: "reviewer"
        skills: ["quality_check", "validation"]

    # 路由配置
    router:
      type: "capability"    # capability, round_robin, priority
      fallback: "all"       # all, any

    # 共识配置
    consensus:
      required: true
      threshold: 0.67       # 67% 同意
      timeout: 300          # 5 分钟
```

### 团队角色

| 角色 | 职责 | 推荐技能 |
|------|------|----------|
| **Researcher** | 信息收集 | web_search, data_collection |
| **Analyst** | 数据分析 | data_analysis, statistics |
| **Developer** | 代码开发 | code_generation, debugging |
| **Reviewer** | 质量审核 | quality_check, validation |
| **Reporter** | 报告生成 | report_generation, visualization |

---

## 任务调度

### 调度策略

#### 1. Capability Based（基于能力）

```go
// 根据代理能力分配任务
func (r *IntelligentRouter) RouteByCapability(task Task) (string, error) {
    // 分析任务需求
    requiredSkills := analyzeRequiredSkills(task)

    // 查找匹配的代理
    for _, member := range r.team.Members {
        if hasSkills(member, requiredSkills) {
            return member.ID, nil
        }
    }

    return "", ErrNoCapableAgent
}
```

#### 2. Round Robin（轮询）

```go
// 依次分配给每个代理
func (r *IntelligentRouter) RouteRoundRobin(task Task) (string, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    member := r.members[r.currentIndex]
    r.currentIndex = (r.currentIndex + 1) % len(r.members)

    return member.ID, nil
}
```

#### 3. Priority Based（基于优先级）

```go
// 根据代理优先级分配
func (r *IntelligentRouter) RouteByPriority(task Task) (string, error) {
    // 按优先级排序成员
    sortedMembers := sortByPriority(r.team.Members)

    // 选择最高优先级的可用成员
    for _, member := range sortedMembers {
        if member.Status == StatusIdle {
            return member.ID, nil
        }
    }

    return "", ErrAllAgentsBusy
}
```

### 调度配置

```yaml
scheduler:
  type: "capability"      # capability, round_robin, priority

  # 并发控制
  max_concurrent_jobs: 5
  queue_size: 100

  # 优先级配置
  priority:
    enabled: true
    levels:
      high:
        - role: "reviewer"
      medium:
        - role: "analyst"
      low:
        - role: "researcher"
```

---

## 消息通信

### 消息类型

| 类型 | 说明 | 使用场景 |
|------|------|----------|
| **Direct Message** | 点对点消息 | 私密通信 |
| **Broadcast** | 广播消息 | 通知所有成员 |
| **Topic Message** | 主题消息 | 特定主题订阅 |
| **Request/Response** | 请求响应模式 | 同步通信 |

### 消息格式

```go
type Message struct {
    ID        string
    Type      MessageType  // Direct, Broadcast, Topic
    From      string       // 发送者 ID
    To        string       // 接收者 ID（Direct 消息）
    Topic     string       // 主题（Topic 消息）
    Content   interface{}   // 消息内容
    Timestamp time.Time
    Metadata  map[string]interface{}
}
```

### 消息发送

```go
// 点对点消息
err := team.SendMessage(ctx, &collaboration.Message{
    Type:    collaboration.MessageTypeDirect,
    From:    "agent_1",
    To:      "agent_2",
    Content: map[string]interface{}{
        "action": "process_data",
        "data":   "...",
    },
})

// 广播消息
err := team.Broadcast(ctx, &collaboration.Message{
    Type:    collaboration.MessageTypeBroadcast,
    From:    "coordinator",
    Content: "任务开始",
})

// 主题消息
err := team.Publish(ctx, "task_updates", &collaboration.Message{
    Type:    collaboration.MessageTypeTopic,
    From:    "agent_1",
    Topic:   "task_updates",
    Content: map[string]interface{}{
        "task_id": "123",
        "status":  "completed",
    },
})
```

### 消息订阅

```go
// 订阅主题
err := team.Subscribe("task_updates", func(msg *collaboration.Message) error {
    taskID := msg.Content["task_id"].(string)
    status := msg.Content["status"].(string)

    log.Infof("任务 %s 状态: %s", taskID, status)
    return nil
})

// 订阅所有消息
err := team.Subscribe("*", func(msg *collaboration.Message) error {
    log.Infof("收到消息: %+v", msg)
    return nil
})
```

---

## 快速开始

### 创建协作团队

**YAML 配置**:

```yaml
teams:
  - name: "data_processing_team"
    description: "数据处理协作团队"
    mode: "parallel"

    members:
      - id: "collector"
        agent: "worker"
        role: "researcher"
        skills: ["web_search", "data_collection"]

      - id: "processor"
        agent: "worker"
        role: "analyst"
        skills: ["data_processing", "analysis"]

      - id: "reporter"
        agent: "worker"
        role: "reporter"
        skills: ["report_generation"]

    router:
      type: "capability"
```

**代码创建**:

```go
// 创建团队
team, err := collaboration.NewTeam(&collaboration.TeamConfig{
    Name:        "data_processing_team",
    Description: "数据处理协作团队",
    Mode:        collaboration.ModeParallel,
})

// 添加成员
team.AddMember(&collaboration.TeamMember{
    ID:     "collector",
    Agent:  researcherAgent,
    Role:   "researcher",
    Skills: []string{"web_search", "data_collection"},
})

team.AddMember(&collaboration.TeamMember{
    ID:     "processor",
    Agent:  analystAgent,
    Role:   "analyst",
    Skills: []string{"data_processing", "analysis"},
})

team.AddMember(&collaboration.TeamMember{
    ID:     "reporter",
    Agent:  reporterAgent,
    Role:   "reporter",
    Skills: []string{"report_generation"},
})

// 启动团队
if err := team.Start(ctx); err != nil {
    log.Fatal(err)
}

// 执行任务
result, err := team.ExecuteTask(ctx, &collaboration.Task{
    Type:        "data_analysis",
    Input:       "https://example.com/data",
    RequiredSkills: []string{"data_processing"},
})
```

---

## 相关文档

- 📘 [Collaboration API 参考](api.md) - 完整 API 文档
- 📘 [Agent 组件](../agent/overview.md) - Agent 系统详解
- 📘 [Workflow 组件](../workflow/overview.md) - 工作流系统
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
