# Workflow 系统概览

> **AgentFramework Workflow 组件文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [工作流类型](#工作流类型)
- [核心概念](#核心概念)
- [系统架构](#系统架构)
- [快速开始](#快速开始)
- [高级特性](#高级特性)

---

## 系统简介

Workflow 是 AgentFramework 的工作流编排引擎，支持定义复杂的任务流程和依赖关系。

### 核心职责

| 职责 | 说明 |
|------|------|
| **流程编排** | 定义和管理任务执行流程 |
| **依赖管理** | 处理任务间的依赖关系 |
| **并行执行** | 支持并行和串行执行 |
| **状态持久化** | 保存和恢复执行状态 |
| **错误处理** | 统一的错误处理和重试 |
| **HITL 支持** | 人工介入和审批节点 |

### 技术特点

- ✅ **DAG 支持**: 有向无环图支持复杂依赖
- ✅ **多种模式**: 顺序、并行、路由等 6 种类型
- ✅ **可视化**: 可视化工作流定义
- ✅ **检查点**: 自动保存和恢复
- ✅ **智能调度**: 基于能力的任务分发
- ✅ **事件驱动**: 基于事件的任务触发

---

## 工作流类型

### 类型概览

| 类型 | 说明 | 复杂度 | 典型场景 |
|------|------|--------|---------|
| **Sequential** | 顺序执行 | ⭐ | 简单任务链 |
| **Parallel** | 并行执行 | ⭐⭐ | 独立任务 |
| **DAG** | 有向无环图 | ⭐⭐⭐⭐ | 复杂依赖 |
| **Routing** | 条件路由 | ⭐⭐⭐ | 动态分支 |
| **Planning** | 规划工作流 | ⭐⭐⭐⭐ | 自动规划 |
| **Graph** | 通用图 | ⭐⭐⭐⭐⭐ | 任意拓扑 |

### 类型对比

```
┌──────────────────────────────────────────────────────────────┐
│              Workflow 类型对比矩阵                       │
├──────────────────────────────────────────────────────────────┤
│ 类型          │依赖│并行│条件│复杂│典型场景    │
├──────────────────────────────────────────────────────────────┤
│ Sequential    │ ✅  │ ❌  │ ❌  │ ⭐  │简单任务链  │
│ Parallel      │ ❌  │ ✅  │ ❌  │ ⭐⭐ │独立任务    │
│ DAG          │ ✅  │ ✅  │ ❌  │ ⭐⭐⭐⭐ │复杂依赖    │
│ Routing      │ ✅  │ ❌  │ ✅  │ ⭐⭐⭐ │动态分支    │
│ Planning     │ ✅  │ ✅  │ ❌  │ ⭐⭐⭐⭐ │自动规划    │
│ Graph        │ ✅  │ ✅  │ ✅  │ ⭐⭐⭐⭐⭐ │任意拓扑    │
└──────────────────────────────────────────────────────────────┘
```

### 类型详解

#### 1. Sequential（顺序）

**说明**: 按顺序执行任务

**配置**:

```yaml
workflows:
  - name: "data_pipeline"
    type: "sequential"
    nodes:
      - name: "collect"
        agent: "worker"
        tools: ["web_search"]
      - name: "process"
        agent: "worker"
        tools: ["data_processing"]
      - name: "save"
        agent: "worker"
        tools: ["file_operation"]
```

**执行流程**:
```
collect → process → save
```

---

#### 2. Parallel（并行）

**说明**: 并行执行多个任务

**配置**:

```yaml
workflows:
  - name: "parallel_tasks"
    type: "parallel"
    nodes:
      - name: "task1"
        agent: "worker"
      - name: "task2"
        agent: "worker"
      - name: "task3"
        agent: "worker"
```

**执行流程**:
```
┌─────────┐     ┌─────────┐     ┌─────────┐
│  task1  │     │  task2  │     │  task3  │
└─────────┘     └─────────┘     └─────────┘
     │                │                │
     └────────────────┴────────────────┘
                        │
                        ▼
                  ┌─────────────┐
                  │   汇聚结果  │
                  └─────────────┘
```

---

#### 3. DAG（有向无环图）

**说明**: 支持复杂依赖关系

**配置**:

```yaml
workflows:
  - name: "complex_workflow"
    type: "dag"
    start_node: "start"
    nodes:
      - name: "start"
        agent: "chat"
      - name: "process_a"
        agent: "worker"
        depends_on: ["start"]
      - name: "process_b"
        agent: "worker"
        depends_on: ["start"]
      - name: "merge"
        agent: "analyst"
        depends_on: ["process_a", "process_b"]
```

**执行流程**:
```
    start
      │
      ▼
┌─────────┴─────────┐
│                   │
│process_a         process_b│
│                   │
└─────────┬─────────┘
          │
          ▼
       merge
```

---

#### 4. Routing（路由）

**说明**: 基于条件动态路由

**配置**:

```yaml
workflows:
  - name: "dynamic_routing"
    type: "routing"
    nodes:
      - name: "router"
        agent: "chat"
        routes:
          - condition: "{{input.type}} == 'api'"
            target: "api_handler"
          - condition: "{{input.type}} == 'file'"
            target: "file_handler"
          - default: true
            target: "default_handler"
      - name: "api_handler"
        agent: "worker"
      - name: "file_handler"
        agent: "worker"
      - name: "default_handler"
        agent: "chat"
```

**执行流程**:
```
    router
      │
      ├─(type=api)→ api_handler
      │
      ├─(type=file)→ file_handler
      │
      └─(default) → default_handler
```

---

#### 5. Planning（规划）

**说明**: 自动任务规划和分解

**配置**:

```yaml
workflows:
  - name: "auto_planning"
    type: "planning"
    nodes:
      - name: "planner"
        agent: "worker"
        role: "researcher"
      - name: "executor"
        agent: "worker"
        role: "developer"
      - name: "validator"
        agent: "worker"
        role: "reviewer"
```

**执行流程**:
1. Planner 分析任务并制定计划
2. Executor 执行计划中的步骤
3. Validator 验证结果

---

#### 6. Graph（通用图）

**说明**: 支持任意拓扑结构

**配置**:

```yaml
workflows:
  - name: "complex_graph"
    type: "graph"
    nodes:
      - name: "node_a"
        agent: "worker"
      - name: "node_b"
        agent: "worker"
      - name: "node_c"
        agent: "worker"
    edges:
      - from: "node_a"
        to: "node_b"
        condition: "{{result.success}} == true"
      - from: "node_a"
        to: "node_c"
        condition: "{{result.success}} == false"
      - from: "node_b"
        to: "node_c"
```

**执行流程**:
```
    node_a
      │
      ├─(success)→ node_b ─┐
      │                       │
      └─(failure)→ node_c ←─────┘
```

---

## 核心概念

### Node（节点）

**定义**: 工作流的基本执行单元

**节点类型**:

| 类型 | 说明 | 配置 |
|------|------|------|
| **Agent Node** | 使用 Agent 执行 | `agent: "agent_name"` |
| **Function Node** | 调用函数 | `function: "function_name"` |
| **Sub-workflow Node** | 嵌套工作流 | `workflow: "workflow_name"` |
| **HITL Node** | 人工介入 | `hitl: true` |

**节点配置**:

```yaml
nodes:
  - name: "data_collection"        # 节点名称
    agent: "worker"                   # 使用的 Agent
    role: "researcher"                # Agent 角色
    tools: ["web_search"]              # 使用的工具

    # 输入配置
    input:
      data_source: "{{input.url}}"
      query: "{{input.query}}"

    # 输出配置
    output:
      result: "{{output.data}}"

    # 重试配置
    retry:
      max_retries: 3
      backoff: "exponential"
      base_delay: 1
      max_delay: 60

    # 超时配置
    timeout: 300
```

### Edge（边）

**定义**: 节点间的连接关系

**边类型**:

| 类型 | 说明 | 配置 |
|------|------|------|
| **Simple Edge** | 简单连接 | `from: "a", to: "b"` |
| **Conditional Edge** | 条件连接 | `condition: "{{result}} > 0"` |
| **Parallel Edge** | 并行分支 | 多个 `from` 指向同一个 `to` |

**边配置**:

```yaml
edges:
  # 简单连接
  - from: "collect"
    to: "process"

  # 条件连接
  - from: "decision"
    to: "branch_a"
    condition: "{{input.type}} == 'a'"

  # 多条件连接
  - from: "decision"
    to: "branch_b"
    condition: |
      and:
        - "{{input.type}} == 'b'"
        - "{{input.priority}} > 5"

  # 默认连接
  - from: "decision"
    to: "default_branch"
    default: true
```

### Variable（变量）

**作用**: 节点间传递数据

**变量类型**:

| 类型 | 说明 | 示例 |
|------|------|------|
| **输入变量** | 工作流输入 | `{{input.param}}` |
| **节点输出** | 节点输出 | `{{node_name.output}}` |
| **环境变量** | 预定义变量 | `{{env.var_name}}` |
| **系统变量** | 系统提供变量 | `{{system.timestamp}}` |

**变量使用**:

```yaml
nodes:
  - name: "collect"
    output:
      data: "{{nodes.collect.output}}"

  - name: "process"
    input:
      source_data: "{{nodes.collect.output.data}}"
```

---

## 系统架构

### 组件架构

```
┌──────────────────────────────────────────────────────────────┐
│                Workflow 系统架构                        │
└──────────────────────────────────────────────────────────────┘

┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│ Workflow    │          │   Node      │          │   Edge      │
│ Engine      │────────▶│ Manager     │────────▶│ Manager     │
└─────────────┘          └─────────────┘          └─────────────┘
        │                        │                        │
        └────────────┬────────────┴────────────┘
                     │
                     ▼
          ┌──────────────────────────┐
          │    Workflow Runtime   │
          │  (执行时）            │
          └──────────────────────────┘
                     │
         ┌───────────┼───────────┬───────────┐
         ▼            ▼            ▼           ▼
┌─────────────┐┌─────────────┐┌─────────────┐┌─────────────┐
│ Scheduler  ││ Checkpoint ││  Executor  ││ Event Bus  │
└─────────────┘└─────────────┘└─────────────┘└─────────────┘
```

### 核心组件

#### 1. Workflow Engine

**功能**: 工作流执行引擎

**职责**:
- 解析工作流定义
- 管理执行状态
- 协调节点执行
- 处理错误和重试

#### 2. Node Manager

**功能**: 节点管理器

**职责**:
- 创建和初始化节点
- 管理节点状态
- 执行节点逻辑
- 处理节点输出

#### 3. Edge Manager

**功能**: 边管理器

**职责**:
- 管理节点连接
- 验证边条件
- 控制执行流向
- 处理循环检测

#### 4. Scheduler

**功能**: 任务调度器

**职责**:
- 决定执行顺序
- 管理并行执行
- 控制并发数
- 优化资源使用

---

## 快速开始

### 创建第一个工作流

#### YAML 配置

**文件**: `workflows/data_pipeline.yaml`

```yaml
name: "data_pipeline"
type: "sequential"
description: "数据处理流水线"

nodes:
  - name: "collect"
    agent: "worker"
    role: "researcher"
    tools: ["web_search"]
    retry:
      max_retries: 3

  - name: "analyze"
    agent: "worker"
    role: "analyst"
    tools: ["data_processing"]
    timeout: 300

  - name: "report"
    agent: "worker"
    role: "reporter"
    tools: ["report_generation"]

edges:
  - from: "collect"
    to: "analyze"
  - from: "analyze"
    to: "report"
```

#### 代码创建

```go
// 创建工作流
workflow, err := host.CreateWorkflow(&workflow.Config{
    Name: "data_pipeline",
    Type:  workflow.TypeSequential,
    Nodes: []workflow.Node{
        {
            Name:  "collect",
            Agent: "worker",
            Role:  "researcher",
            Tools: []string{"web_search"},
        },
        {
            Name:  "analyze",
            Agent: "worker",
            Role:  "analyst",
            Tools: []string{"data_processing"},
        },
        {
            Name:  "report",
            Agent: "worker",
            Role:  "reporter",
            Tools: []string{"report_generation"},
        },
    },
    Edges: []workflow.Edge{
        {From: "collect", To: "analyze"},
        {From: "analyze", To: "report"},
    },
})
if err != nil {
    log.Fatal(err)
}

// 执行工作流
result, err := workflow.Run(ctx, `{"url": "https://example.com/data"}`)
```

---

## 高级特性

### 1. 检查点（Checkpoint）

**功能**: 保存和恢复执行状态

**配置**:

```yaml
workflows:
  - name: "checkpoint_workflow"
    checkpoint:
      enabled: true
      backend: "sqlite"         # sqlite, redis, memory
      interval: 60              # 每60秒保存一次
      save_on_nodes:             # 保存这些节点的状态
        - "critical_step"
        - "data_processing"
```

**使用**:
```go
// 工作流会自动保存状态
result, err := workflow.Run(ctx, input)

// 如果失败，可以从检查点恢复
if err != nil {
    recoveredResult, recoverErr := workflow.Resume(ctx)
    if recoverErr != nil {
        log.Fatal(recoverErr)
    }
    result = recoveredResult
}
```

### 2. HITL（人工介入）

**功能**: 关键决策需要人工批准

**配置**:

```yaml
workflows:
  - name: "hitl_workflow"
    nodes:
      - name: "critical_decision"
        agent: "worker"
        hitl:
          enabled: true
          approval_mode: "manual"    # manual, auto
          timeout: 3600             # 1小时
          notification_channels:
            - type: "slack"
              channel: "#approvals"
            - type: "email"
              recipients: ["admin@example.com"]
```

### 3. 错误处理和重试

**配置**:

```yaml
workflows:
  - name: "retry_workflow"
    retry:
      max_retries: 3
      backoff: "exponential"    # linear, exponential
      base_delay: 1             # 秒
      max_delay: 60              # 秒
      retry_on:
        - "network_error"
        - "timeout"
```

### 4. 并发控制

**配置**:

```yaml
workflows:
  - name: "concurrent_workflow"
    concurrency:
      max_parallel: 5           # 最多5个节点并行
      queue_size: 100            # 队列大小
      priority: "fifo"          # fifo, priority
```

---

## 相关文档

- 📘 [Workflow 类型详解](types.md) - 各种工作流类型
- 📘 [Workflow API 参考](api.md) - 完整 API 文档
- 📘 [agent/flow.go](../../agent/flow.go) - 源代码
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
