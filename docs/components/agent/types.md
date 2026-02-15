# Agent 类型详解

> **AgentFramework Agent 类型完整文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [类型概览](#类型概览)
- [ChatAgent](#chatagent)
- [ReActAgent](#reactagent)
- [HumanAgent](#humanagent)
- [WorkerAgent](#workeragent)
- [EdgeAgent](#edgeagent)
- [类型选择指南](#类型选择指南)

---

## 类型概览

### 类型对比表

| 类型 | 复杂度 | 推理能力 | 工具使用 | HITL | 典型场景 |
|------|---------|-----------|---------|------|---------|
| **ChatAgent** | ⭐ | ❌ | ✅ | ✅ | 客服机器人、简单对话 |
| **ReActAgent** | ⭐⭐⭐ | ✅ | ✅ | ✅ | 复杂任务、推理循环 |
| **HumanAgent** | ⭐⭐ | ❌ | ❌ | ✅✅ | 人工审核、HITL 节点 |
| **WorkerAgent** | ⭐⭐⭐⭐ | ✅ | ✅✅ | ✅ | 专业任务、7 种角色 |
| **EdgeAgent** | ⭐⭐⭐ | ✅ | ✅ | ❌ | 边缘计算、资源受限 |

### 能力矩阵

```
┌──────────────────────────────────────────────────────────────┐
│                    Agent 能力矩阵                      │
├──────────────────────────────────────────────────────────────┤
│ 能力          │ Chat │ ReAct │ Human │ Worker │ Edge │
├──────────────────────────────────────────────────────────────┤
│ 基础对话      │  ✅  │  ✅  │  ✅  │  ✅  │  ✅  │
│ 推理循环      │  ❌  │  ✅  │  ❌  │  ✅  │  ✅  │
│ 工具调用       │  ✅  │  ✅  │  ❌  │  ✅✅ │  ✅  │
│ 流式响应      │  ✅  │  ✅  │  ❌  │  ✅  │  ✅  │
│ HITL 支持      │  ✅  │  ✅  │  ✅✅ │  ✅  │  ❌  │
│ 专业角色      │  ❌  │  ❌  │  ❌  │  ✅  │  ❌  │
│ 记忆压缩      │  ✅  │  ✅  │  ❌  │  ✅  │  ✅  │
│ 模型切换      │  ✅  │  ✅  │  ❌  │  ✅  │  ✅  │
└──────────────────────────────────────────────────────────────┘
```

---

## ChatAgent

### 类型说明

**文件**: [agent/chat_agent.go](../../agent/chat_agent.go)

ChatAgent 是最简单的 Agent 类型，专注于基础的对话交互。适合不需要复杂推理的场景。

### 核心特性

| 特性 | 说明 |
|------|------|
| **简单直观** | API 简洁，易于理解 |
| **低延迟** | 无额外推理开销 |
| **高可靠** | 行为可预测，易于调试 |
| **工具支持** | 可以使用工具扩展能力 |

### 配置示例

**YAML 配置**:

```yaml
agents:
  - name: "chat"
    type: "chat"
    model: "ollama-llama3"
    instructions: "你是一个有用的AI助手"
    temperature: 0.7
    max_tokens: 2000

    # 工具配置
    tools:
      - "http_request"
      - "file_operation"

    # HITL 配置
    hitl:
      enabled: true
      approval_mode: "manual"
```

**代码创建**:

```go
chatAgent := agent.NewChatAgent(
    agent.WithName("chat"),
    agent.WithInstructions("你是一个有用的AI助手"),
    agent.WithModel("ollama-llama3"),
    agent.WithTools("http_request", "file_operation"),
    agent.WithHITL(true),
)
```

### 使用场景

| 场景 | 说明 | 配置建议 |
|------|------|----------|
| **客服机器人** | 回答常见问题 | 添加知识库工具 |
| **聊天助手** | 日常对话 | 降低 temperature |
| **简单任务** | 文本处理、翻译 | 添加数据处理工具 |
| **API 服务** | 对话式 API | 关闭 HITL |

### 最佳实践

```go
// ✅ 正确：简单场景使用 ChatAgent
chatAgent := agent.NewChatAgent(
    agent.WithInstructions("简洁回答用户问题"),
    agent.WithTemperature(0.3), // 降低随机性
)

// ❌ 错误：复杂任务使用 ChatAgent
// 应该使用 ReActAgent 或 WorkerAgent
```

### 性能指标

| 指标 | 数值 |
|------|------|
| **响应延迟** | 500ms - 2s |
| **内存占用** | 50MB - 100MB |
| **并发能力** | 1000+ 并发 |
| **Token 效率** | 高（无额外推理） |

---

## ReActAgent

### 类型说明

**文件**: [agent/react_agent.go](../../agent/react_agent.go)

ReActAgent 实现 **ReAct (Reasoning + Acting)** 范式，通过推理-行动循环完成复杂任务。

### 核心特性

| 特性 | 说明 |
|------|------|
| **推理循环** | Thought → Action → Observation 循环 |
| **自我纠错** | 可以从错误中学习和恢复 |
| **任务分解** | 自动分解复杂任务为子任务 |
| **工具编排** | 智能选择和组合工具 |

### ReAct 循环

```
┌────────────────────────────────────────────────────────┐
│                 ReAct 循环流程                     │
└────────────────────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────┐
          │   用户输入   │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  Thought:     │
          │  我应该做什么？ │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  Action:      │
          │  调用工具 X    │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ Observation:  │
          │ 工具返回结果   │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ 完成了吗？   │
          └─────────────┘
         │              │
         No             Yes
         │              │
         ▼              ▼
    ┌─────────────┐  ┌─────────────┐
    │ 继续循环      │  │ 返回结果      │
    └─────────────┘  └─────────────┘
```

### 配置示例

**YAML 配置**:

```yaml
agents:
  - name: "react"
    type: "react"
    model: "ollama-llama3"
    instructions: "你是一个专业的研究助理"

    # 循环配置
    max_iterations: 10       # 最大循环次数
    max_time: 300            # 最大执行时间（秒）
    verbose: true            # 详细日志

    # 工具配置
    tools:
      - "web_search"
      - "data_analysis"
      - "file_operation"

    # 推理配置
    reasoning:
      show_thoughts: true     # 显示推理过程
      require_plan: true     # 要求制定计划
```

**代码创建**:

```go
reactAgent := agent.NewReActAgent(
    agent.WithName("react"),
    agent.WithInstructions("你是一个专业的研究助理"),
    agent.WithModel("ollama-llama3"),
    agent.WithMaxIterations(10),
    agent.WithMaxTime(300*time.Second),
    agent.WithVerbose(true),
    agent.WithTools(
        "web_search",
        "data_analysis",
        "file_operation",
    ),
)
```

### 使用场景

| 场景 | 说明 | 推荐工具 |
|------|------|----------|
| **数据分析** | 多步骤数据处理 | data_processing, analysis |
| **研究任务** | 信息收集和分析 | web_search, file_operation |
| **代码生成** | 多步骤编程任务 | code_execution, file_operation |
| **自动化流程** | 多步骤工作流 | 多个工具组合 |

### 最佳实践

```go
// ✅ 正确：为复杂任务设置合理限制
reactAgent := agent.NewReActAgent(
    agent.WithMaxIterations(10),    // 防止无限循环
    agent.WithMaxTime(300*time.Second), // 防止超时
    agent.WithVerbose(true),         // 便于调试
)

// ✅ 正确：提供详细的指令
reactAgent := agent.NewReActAgent(
    agent.WithInstructions(`
        你是一个专业的研究助理。
        工作流程：
        1. 理解用户需求
        2. 制定详细计划
        3. 逐步执行计划
        4. 验证结果
    `),
)
```

### 性能指标

| 指标 | 数值 |
|------|------|
| **平均循环次数** | 3 - 7 次 |
| **任务成功率** | 85% - 95% |
| **响应延迟** | 5s - 30s |
| **内存占用** | 100MB - 200MB |

---

## HumanAgent

### 类型说明

**文件**: [agent/hitl.go](../../agent/hitl.go)

HumanAgent 实现 **Human-in-the-Loop (HITL)** 模式，用于需要人工审核和介入的场景。

### 核心特性

| 特性 | 说明 |
|------|------|
| **人工审核** | 关键决策需要人工批准 |
| **通知机制** | 多渠道通知（Slack、Email 等） |
| **超时处理** | 可配置的等待超时 |
| **审核历史** | 完整的审核记录 |

### HITL 流程

```
┌────────────────────────────────────────────────────────┐
│                 HITL 工作流程                     │
└────────────────────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ Agent 请求   │
          │ 需要审核      │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ 发送通知     │
          │ (多渠道)      │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ 等待人工响应  │
          │ (超时配置)     │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ 人工决策      │
          │ (批准/拒绝)     │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │ 记录决策      │
          │ 继续/终止      │
          └─────────────┘
```

### 配置示例

**YAML 配置**:

```yaml
agents:
  - name: "human_agent"
    type: "human"
    model: "ollama-llama3"

    # HITL 配置
    hitl:
      enabled: true
      approval_mode: "manual"  # manual, auto

      # 超时配置
      timeout: 3600           # 1小时（秒）
      reminder_interval: 600  # 10分钟提醒一次

      # 通知配置
      notification_channels:
        - type: "slack"
          channel: "#agent-approval"
          priority: "high"

        - type: "email"
          recipients: ["admin@example.com"]
          priority: "high"

      # 审批规则
      approval_rules:
        - rule: "data_deletion"
          requires_approval: true
          approvers: ["admin", "manager"]

        - rule: "api_call"
          requires_approval: false
          condition: "cost < 100"
```

**代码创建**:

```go
humanAgent := agent.NewHumanAgent(
    agent.WithName("human_agent"),
    agent.WithModel("ollama-llama3"),
    agent.WithHITLConfig(&agent.HITLConfig{
        Enabled:         true,
        ApprovalMode:    "manual",
        Timeout:         3600 * time.Second,
        ReminderInterval: 600 * time.Second,
        NotificationChannels: []agent.NotificationChannel{
            {
                Type:     "slack",
                Channel:  "#agent-approval",
                Priority:  "high",
            },
        },
    },
)
```

### 使用场景

| 场景 | 说明 | 审批规则 |
|------|------|----------|
| **数据删除** | 危险操作，需要审核 | 必须批准 |
| **API 调用** | 成本敏感操作 | 有条件批准 |
| **敏感内容** | 安全策略检查 | 必须批准 |
| **财务决策** | 金额相关操作 | 必须批准 |

### 最佳实践

```go
// ✅ 正确：设置合理的超时
humanAgent := agent.NewHumanAgent(
    agent.WithTimeout(3600*time.Second), // 1小时
    agent.WithReminderInterval(600*time.Second), // 10分钟提醒
)

// ✅ 正确：多渠道通知
humanAgent := agent.NewHumanAgent(
    agent.WithNotificationChannels(
        agent.NotificationChannel{
            Type:     "slack",
            Channel:  "#approval",
            Priority:  "high",
        },
        agent.NotificationChannel{
            Type:      "email",
            Recipients: []string{"admin@example.com"},
            Priority:   "high",
        },
    ),
)
```

---

## WorkerAgent

### 类型说明

**文件**: [agent/swe_agent.go](../../agent/swe_agent.go)

WorkerAgent 是专业工作代理，内置 7 种专业角色，每种角色有特定的技能和工具集。

### 专业角色

| 角色 | 说明 | 专业工具 |
|------|------|----------|
| **Developer** | 开发者代理 | 代码生成、审查、调试 |
| **Browser** | 浏览器代理 | 网页导航、表单填充、数据提取 |
| **Document** | 文档代理 | 文档处理、格式转换、摘要生成 |
| **MultiModal** | 多模态代理 | 图像分析、图表生成 |
| **Researcher** | 研究者代理 | 网络搜索、数据收集、分析 |
| **Writer** | 写作代理 | 内容生成、编辑、格式化 |
| **Reviewer** | 审核员代理 | 代码审查、质量检查 |

### 配置示例

**YAML 配置**:

```yaml
agents:
  # 开发者代理
  - name: "developer"
    type: "worker"
    role: "developer"
    model: "ollama-llama3"
    instructions: "你是一个专业的软件开发者"

    # 专业技能
    skills:
      - "code_generation"
      - "code_review"
      - "debugging"

    # 专业工具
    tools:
      - "code_execution"
      - "file_operation"
      - "git_operations"

  # 浏览器代理
  - name: "browser"
    type: "worker"
    role: "browser"
    model: "ollama-llama3"
    instructions: "你是一个专业的网页自动化专家"

    skills:
      - "web_navigation"
      - "form_filling"
      - "data_extraction"

    tools:
      - "browser_automation"
      - "http_request"
```

**代码创建**:

```go
// 开发者代理
developerAgent := agent.NewWorkerAgent(
    agent.WithName("developer"),
    agent.WithRole("developer"),
    agent.WithInstructions("你是一个专业的软件开发者"),
    agent.WithModel("ollama-llama3"),
    agent.WithSkills(
        "code_generation",
        "code_review",
        "debugging",
    ),
    agent.WithTools(
        "code_execution",
        "file_operation",
        "git_operations",
    ),
)

// 浏览器代理
browserAgent := agent.NewWorkerAgent(
    agent.WithName("browser"),
    agent.WithRole("browser"),
    agent.WithInstructions("你是一个专业的网页自动化专家"),
    agent.WithModel("ollama-llama3"),
    agent.WithSkills(
        "web_navigation",
        "form_filling",
        "data_extraction",
    ),
    agent.WithTools(
        "browser_automation",
        "http_request",
    ),
)
```

### 角色详解

#### Developer（开发者）

**能力**: 代码生成、审查、调试、重构

**推荐工具**:
- `code_execution` - 代码执行
- `file_operation` - 文件操作
- `git_operations` - Git 操作

**使用场景**:
- 自动化编码任务
- 代码审查和重构
- Bug 修复和调试

#### Browser（浏览器）

**能力**: 网页导航、表单填充、数据提取

**推荐工具**:
- `browser_automation` - 浏览器自动化
- `http_request` - HTTP 请求

**使用场景**:
- 网页数据抓取
- 自动化测试
- 表单自动填充

#### Document（文档）

**能力**: 文档处理、格式转换、摘要生成

**推荐工具**:
- `file_operation` - 文件操作
- `data_processing` - 数据处理

**使用场景**:
- 文档格式转换
- 自动化摘要生成
- 文档质量检查

#### MultiModal（多模态）

**能力**: 图像分析、图表生成、多模态理解

**推荐工具**:
- `image_analysis` - 图像分析
- `chart_generation` - 图表生成

**使用场景**:
- 图像理解和分析
- 数据可视化
- 多模态内容生成

#### Researcher（研究者）

**能力**: 网络搜索、数据收集、分析研究

**推荐工具**:
- `web_search` - 网络搜索
- `data_collection` - 数据收集
- `analysis` - 数据分析

**使用场景**:
- 信息收集和验证
- 市场研究
- 竞争分析

#### Writer（写作）

**能力**: 内容生成、编辑、格式化

**推荐工具**:
- `content_generation` - 内容生成
- `formatting` - 格式化工具

**使用场景**:
- 自动化写作
- 内容编辑和优化
- 文档格式化

#### Reviewer（审核员）

**能力**: 代码审查、质量检查、合规验证

**推荐工具**:
- `code_review` - 代码审查
- `quality_check` - 质量检查

**使用场景**:
- 自动化代码审查
- 质量保证
- 合规性检查

### 最佳实践

```go
// ✅ 正确：选择合适的角色
developerAgent := agent.NewWorkerAgent(
    agent.WithRole("developer"), // 明确角色
    agent.WithSkills(
        "code_generation",
        "code_review",
        "debugging",
    ),
)

// ✅ 正确：提供专业的指令
developerAgent := agent.NewWorkerAgent(
    agent.WithInstructions(`
        你是一个专业的软件开发者，专长于：
        - 编写高质量、可维护的代码
        - 遵循 SOLID 原则和最佳实践
        - 进行全面的代码审查
        - 系统化地调试和修复问题
    `),
)
```

---

## EdgeAgent

### 类型说明

**文件**: [agent/edge_agent.go](../../agent/edge_agent.go)

EdgeAgent 是专为边缘计算和资源受限环境设计的轻量级 Agent。

### 核心特性

| 特性 | 说明 |
|------|------|
| **低资源占用** | 内存占用 < 50MB |
| **离线工作** | 支持本地模型，无需网络 |
| **快速启动** | 冷启动 < 1s |
| **省电模式** | 智能电源管理 |

### 配置示例

**YAML 配置**:

```yaml
agents:
  - name: "edge"
    type: "edge"
    model: "ollama-llama3"
    instructions: "你是一个边缘计算助手"

    # 资源限制
    resource_limits:
      memory: "50m"          # 50MB
      cpu: "0.5"             # 50% CPU
      battery_saver: true    # 省电模式

    # 缓存策略
    cache:
      enabled: true
      type: "disk"            # memory, disk
      size: 100               # 100MB

    # 离线支持
    offline_mode:
      enabled: true
      local_model: true
      fallback_to_online: false
```

**代码创建**:

```go
edgeAgent := agent.NewEdgeAgent(
    agent.WithName("edge"),
    agent.WithInstructions("你是一个边缘计算助手"),
    agent.WithModel("ollama-llama3"),
    agent.WithResourceLimits(&agent.ResourceLimits{
        Memory:        50 * 1024 * 1024, // 50MB
        CPU:           0.5,                // 50%
        BatterySaver:  true,
    }),
    agent.WithCacheConfig(&agent.CacheConfig{
        Enabled: true,
        Type:     "disk",
        Size:     100 * 1024 * 1024, // 100MB
    }),
    agent.WithOfflineMode(true),
)
```

### 使用场景

| 场景 | 说明 | 推荐配置 |
|------|------|----------|
| **IoT 设备** | 资源受限的嵌入式设备 | 内存限制、省电模式 |
| **离线环境** | 无网络连接的环境 | 本地模型、离线模式 |
| **移动设备** | 电池供电的设备 | 省电模式、低资源占用 |

---

## 类型选择指南

### 决策流程图

```
┌────────────────────────────────────────────────────────┐
│              Agent 类型选择决策树                  │
└────────────────────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────────┐
          │ 需要人工审核？ │
          └─────────────────┘
         │               │
         Yes              No
         │               │
         ▼               ▼
  ┌────────────┐  ┌─────────────────┐
  │ HumanAgent │  │ 需要复杂推理？ │
  └────────────┘  └─────────────────┘
                         │               │
                        Yes              No
                        │               │
                        ▼               ▼
              ┌─────────────────┐  ┌─────────────────┐
              │ ReActAgent      │  │ 是简单对话？ │
              └─────────────────┘  └─────────────────┘
                     │               │               │
                     ▼               Yes            No
                     ▼               ▼               ▼
           ┌─────────────────┐  ┌────────────┐  ┌─────────────┐
           │ WorkerAgent    │  │ ChatAgent  │  │ EdgeAgent   │
           │ (专业角色)      │  │           │  │             │
           └─────────────────┘  └────────────┘  └─────────────┘
```

### 场景映射表

| 使用场景 | 推荐类型 | 理由 |
|---------|---------|------|
| 客服机器人 | ChatAgent | 简单对话，低延迟 |
| 数据分析 | ReActAgent | 需要推理循环 |
| 代码生成 | WorkerAgent (Developer) | 专业角色，专门工具 |
| 网页自动化 | WorkerAgent (Browser) | 专业角色，浏览器工具 |
| 内容审查 | HumanAgent | 需要人工审核 |
| IoT 设备 | EdgeAgent | 资源受限 |
| 自动化测试 | WorkerAgent (Developer) | 专业角色，代码执行 |
| 市场研究 | WorkerAgent (Researcher) | 专业角色，搜索工具 |

---

## 相关文档

- 📘 [Agent 概览](overview.md) - Agent 系统概览
- 📘 [Agent 生命周期](lifecycle.md) - 生命周期管理
- 📘 [Agent API 参考](api.md) - 完整 API 文档
- 📘 [配置指南](../configuration/CONFIGURATION.md) - 详细配置说明
- 📘 [最佳实践](../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
