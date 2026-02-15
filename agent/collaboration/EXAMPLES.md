# Agent 协作系统使用示例

本文档展示如何使用 AgentFramework 的 Agent 协作系统和智能路由功能。

## 目录

1. [基础概念](#基础概念)
2. [创建 Agent 团队](#创建-agent-团队)
3. [任务分配](#任务分配)
4. [智能路由](#智能路由)
5. [共识机制](#共识机制)
6. [任务编排](#任务编排)
7. [完整示例](#完整示例)

---

## 基础概念

### AgentTeam

AgentTeam 是管理多个 Agent 协作的核心组件，提供以下功能：

- **成员管理**：添加、移除、查询团队成员
- **任务分配**：将任务分配给最合适的 Agent
- **智能路由**：基于多种策略选择最佳 Agent
- **性能监控**：追踪团队成员的性能指标

### IntelligentRouter

IntelligentRouter 提供多种路由策略：

- **RoundRobin**：轮询分配
- **LeastLoaded**：选择负载最低的 Agent
- **FastestResponse**：选择响应最快的 Agent
- **CostOptimized**：选择成本最低的 Agent
- **Intelligent**：综合多因素评分选择
- **CapabilityMatch**：基于能力匹配
- **PriorityBased**：基于优先级选择

---

## 创建 Agent 团队

### 步骤 1：创建团队配置

```go
import "github.com/your-org/AgentFramework/agent/collaboration"

config := collaboration.TeamConfig{
    Name:          "development-team",
    Description:   "Team for software development tasks",
    MaxConcurrent: 10,
    RouterConfig: collaboration.RouterConfig{
        DefaultStrategy: collaboration.StrategyIntelligent,
        EnableCaching:   true,
        CacheTTL:       5 * time.Minute,
        ScoringWeights: collaboration.ScoringWeights{
            Latency:  0.30,
            Success:  0.30,
            Load:     0.20,
            Cost:     0.10,
            Quality:  0.10,
        },
    },
}

team := collaboration.NewAgentTeam(config)
```

### 步骤 2：包装现有 Agent

```go
// 假设你已经有现有的 Agent
developerAgent := host.GetAgent("developer")
testerAgent := host.GetAgent("tester")
analystAgent := host.GetAgent("analyst")

// 包装为 AgentWrapper
developerWrapper := collaboration.NewDefaultAgentWrapper(
    developerAgent,
    []string{"coding", "review", "debugging"},
    "gpt-4",
)

testerWrapper := collaboration.NewDefaultAgentWrapper(
    testerAgent,
    []string{"testing", "qa"},
    "gpt-4",
)

analystWrapper := collaboration.NewDefaultAgentWrapper(
    analystAgent,
    []string{"analysis", "documentation"},
    "gpt-3.5-turbo",
)
```

### 步骤 3：添加成员到团队

```go
// 添加开发者（优先级 5）
team.AddMember(developerWrapper, "developer", []string{"coding", "review"}, 5)

// 添加测试人员（优先级 3）
team.AddMember(testerWrapper, "tester", []string{"testing"}, 3)

// 添加分析师（优先级 4）
team.AddMember(analystWrapper, "analyst", []string{"analysis"}, 4)
```

### 步骤 4：启动团队

```go
ctx := context.Background()
if err := team.Start(ctx); err != nil {
    log.Fatalf("Failed to start team: %v", err)
}
defer team.Stop()
```

---

## 任务分配

### 基础任务分配

```go
task := &collaboration.CollaborativeTask{
    ID:     "task-001",
    Type:   "code-review",
    Input:  "Please review this Go code for bugs and best practices",
    Priority: 5,
    Timeout: 30 * time.Second,
    RequiredCapabilities: []string{"review"},
}

result, err := team.AssignTask(ctx, task)
if err != nil {
    log.Fatalf("Task assignment failed: %v", err)
}

fmt.Printf("Task completed by %s\n", result.AgentName)
fmt.Printf("Output: %s\n", result.Output)
fmt.Printf("Duration: %v\n", result.Duration)
```

### 指定路由策略

```go
// 使用最少负载策略
result, err := team.AssignTaskWithStrategy(ctx, task, collaboration.StrategyLeastLoaded)
```

### 广播任务给所有匹配的 Agent

```go
results, err := team.BroadcastTask(ctx, task)
for i, result := range results {
    fmt.Printf("Result %d from %s: %s\n", i+1, result.AgentName, result.Output)
}
```

---

## 智能路由

### 自定义路由权重

```go
config := collaboration.TeamConfig{
    Name:        "custom-team",
    RouterConfig: collaboration.RouterConfig{
        DefaultStrategy: collaboration.StrategyIntelligent,
        ScoringWeights: collaboration.ScoringWeights{
            Latency:  0.40, // 重视响应速度
            Success:  0.30,
            Load:     0.10,
            Cost:     0.15, // 重视成本
            Quality:  0.05,
        },
    },
}
```

### 获取路由统计

```go
stats := team.GetRouterStats()
fmt.Printf("Total selections: %d\n", stats.TotalSelections)
fmt.Printf("Cache hit rate: %.2f%%\n",
    float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
fmt.Printf("Average selection time: %v\n", stats.AverageSelectionTime)

for strategy, count := range stats.StrategyUsage {
    fmt.Printf("  %s: %d\n", strategy, count)
}
```

---

## 共识机制

### 创建共识管理器

```go
consensusMgr := collaboration.NewConsensusManager(
    collaboration.ConsensusMajority,
    team,
    30*time.Second,
)
```

### 达成共识

```go
task := &collaboration.CollaborativeTask{
    ID:     "consensus-task",
    Type:   "decision",
    Input:  "Should we use microservices architecture?",
    RequiredCapabilities: []string{"architecture"},
    Timeout: 20 * time.Second,
}

result, err := consensusMgr.ReachConsensus(ctx, task)
if err != nil {
    log.Fatalf("Consensus failed: %v", err)
}

fmt.Printf("Consensus result: %s\n", result.Output)
fmt.Printf("Confidence: %.2f\n", result.Confidence)
fmt.Printf("Agreement: %.2f\n", result.Agreement)
fmt.Printf("Votes: %d/%d\n", result.WinningVotes, result.TotalVotes)
```

---

## 任务编排

### 创建顺序工作流

```go
orchestrator := collaboration.NewOrchestrator(team)

steps := []collaboration.OrchestrationStep{
    {
        ID:   "analyze",
        Name: "Analyze requirements",
        Type: collaboration.OrchestrationSequential,
        Task: &collaboration.CollaborativeTask{
            ID:    "task-analyze",
            Type:  "analysis",
            Input: "Analyze these requirements",
            RequiredCapabilities: []string{"analysis"},
        },
        AgentName: "analyst",
        Timeout:   10 * time.Second,
    },
    {
        ID:   "implement",
        Name: "Implement feature",
        Type: collaboration.OrchestrationSequential,
        Task: &collaboration.CollaborativeTask{
            ID:    "task-implement",
            Type:  "coding",
            Input: "Implement the feature",
            RequiredCapabilities: []string{"coding"},
        },
        Dependencies: []string{"analyze"}, // Wait for analyze to complete
        AgentName:   "developer",
        Timeout:     30 * time.Second,
    },
    {
        ID:   "test",
        Name: "Test implementation",
        Type: collaboration.OrchestrationSequential,
        Task: &collaboration.CollaborativeTask{
            ID:    "task-test",
            Type:  "testing",
            Input: "Test the implementation",
            RequiredCapabilities: []string{"testing"},
        },
        Dependencies: []string{"implement"}, // Wait for implement to complete
        AgentName:   "tester",
        Timeout:     15 * time.Second,
    },
}

workflow, err := orchestrator.CreateWorkflow(
    "dev-workflow",
    "Development Workflow",
    collaboration.OrchestrationSequential,
    steps,
)
```

### 创建并行工作流

```go
parallelSteps := []collaboration.OrchestrationStep{
    {
        ID:   "unit-test",
        Name: "Run unit tests",
        Type: collaboration.OrchestrationParallel,
        Task: &collaborative.CollaborativeTask{
            ID:    "task-unit-test",
            Type:  "testing",
            Input: "Run unit tests",
        },
    },
    {
        ID:   "integration-test",
        Name: "Run integration tests",
        Type: collaboration.OrchestrationParallel,
        Task: &collaboration.CollaborativeTask{
            ID:    "task-integration-test",
            Type:  "testing",
            Input: "Run integration tests",
        },
    },
    {
        ID:   "security-scan",
        Name: "Run security scan",
        Type: collaboration.OrchestrationParallel,
        Task: &collaboration.CollaborativeTask{
            ID:    "task-security",
            Type:  "security",
            Input: "Run security scan",
        },
    },
}

workflow, err := orchestrator.CreateWorkflow(
    "ci-workflow",
    "CI Pipeline",
    collaboration.OrchestrationParallel,
    parallelSteps,
)
```

### 执行工作流

```go
result, err := orchestrator.Execute(ctx, "dev-workflow")
if err != nil {
    log.Fatalf("Workflow execution failed: %v", err)
}

fmt.Printf("Workflow completed: %v\n", result.Success)
fmt.Printf("Duration: %v\n", result.Duration)

for _, stepResult := range result.StepResults {
    fmt.Printf("  Step %s: %v (%v)\n",
        stepResult.StepName,
        stepResult.Success,
        stepResult.Duration)
}
```

---

## 完整示例

### 软件开发团队自动化

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/your-org/AgentFramework/agent"
    "github.com/your-org/AgentFramework/agent/collaboration"
)

func main() {
    ctx := context.Background()

    // 1. 创建 Host 并获取现有 Agent
    host := createTestHost()

    developerAgent, _ := host.GetAgent("developer")
    testerAgent, _ := host.GetAgent("tester")
    reviewerAgent, _ := host.GetAgent("reviewer")

    // 2. 包装 Agent
    developer := collaboration.NewDefaultAgentWrapper(
        developerAgent,
        []string{"coding", "debugging"},
        "gpt-4",
    )

    tester := collaboration.NewDefaultAgentWrapper(
        testerAgent,
        []string{"testing", "qa"},
        "gpt-3.5-turbo",
    )

    reviewer := collaboration.NewDefaultAgentWrapper(
        reviewerAgent,
        []string{"review", "analysis"},
        "gpt-4",
    )

    // 3. 创建团队
    team := collaboration.NewAgentTeam(collaboration.TeamConfig{
        Name:          "dev-team",
        Description:   "Software development team",
        MaxConcurrent: 5,
        RouterConfig: collaboration.RouterConfig{
            DefaultStrategy: collaboration.StrategyIntelligent,
            EnableCaching:   true,
            CacheTTL:       10 * time.Minute,
        },
    })

    // 4. 添加成员
    team.AddMember(developer, "developer", []string{"coding", "debugging"}, 5)
    team.AddMember(tester, "tester", []string{"testing", "qa"}, 3)
    team.AddMember(reviewer, "reviewer", []string{"review", "analysis"}, 4)

    // 5. 启动团队
    if err := team.Start(ctx); err != nil {
        log.Fatalf("Failed to start team: %v", err)
    }
    defer team.Stop()

    // 6. 创建工作流
    orchestrator := collaboration.NewOrchestrator(team)

    steps := []collaboration.OrchestrationStep{
        {
            ID:   "implement",
            Name: "Implement feature",
            Type: collaboration.OrchestrationSequential,
            Task: &collaboration.CollaborativeTask{
                ID:     "task-1",
                Type:   "coding",
                Input:  "Implement a REST API endpoint for user authentication",
                Priority: 7,
                Timeout: 60 * time.Second,
                RequiredCapabilities: []string{"coding"},
            },
            AgentName: "developer",
            Timeout:   60 * time.Second,
        },
        {
            ID:   "review",
            Name: "Code review",
            Type: collaboration.OrchestrationSequential,
            Task: &collaboration.CollaborativeTask{
                ID:     "task-2",
                Type:   "review",
                Input:  "Review the authentication implementation",
                Priority: 6,
                Timeout: 30 * time.Second,
                RequiredCapabilities: []string{"review"},
            },
            Dependencies: []string{"implement"},
            AgentName:   "reviewer",
            Timeout:     30 * time.Second,
        },
        {
            ID:   "test",
            Name: "Testing",
            Type: collaboration.OrchestrationSequential,
            Task: &collaboration.CollaborativeTask{
                ID:     "task-3",
                Type:   "testing",
                Input:  "Test the authentication endpoint",
                Priority: 6,
                Timeout: 45 * time.Second,
                RequiredCapabilities: []string{"testing"},
            },
            Dependencies: []string{"review"},
            AgentName:   "tester",
            Timeout:     45 * time.Second,
        },
    }

    workflow, err := orchestrator.CreateWorkflow(
        "auth-workflow",
        "Authentication Feature Workflow",
        collaboration.OrchestrationSequential,
        steps,
    )
    if err != nil {
        log.Fatalf("Failed to create workflow: %v", err)
    }

    // 7. 执行工作流
    fmt.Println("Executing workflow...")
    result, err := orchestrator.Execute(ctx, workflow.ID)
    if err != nil {
        log.Fatalf("Workflow execution failed: %v", err)
    }

    // 8. 输出结果
    fmt.Println("\n=== Workflow Results ===")
    fmt.Printf("Success: %v\n", result.Success)
    fmt.Printf("Duration: %v\n", result.Duration)
    fmt.Println("\nStep Results:")

    for i, stepResult := range result.StepResults {
        fmt.Printf("\n%d. %s\n", i+1, stepResult.StepName)
        fmt.Printf("   Success: %v\n", stepResult.Success)
        fmt.Printf("   Duration: %v\n", stepResult.Duration)
        if stepResult.Error != nil {
            fmt.Printf("   Error: %v\n", stepResult.Error)
        }
    }

    // 9. 输出团队性能统计
    stats := team.GetPerformanceStats()
    fmt.Println("\n=== Team Performance ===")
    fmt.Printf("Total Members: %d\n", stats.TotalMembers)
    fmt.Printf("Idle: %d, Busy: %d, Overloaded: %d\n",
        stats.IdleMembers, stats.BusyMembers, stats.OverloadedMembers)

    for name, perf := range stats.MemberStats {
        fmt.Printf("\n%s:\n", name)
        fmt.Printf("  Total Tasks: %d\n", perf.TotalTasks)
        fmt.Printf("  Success Rate: %.2f%%\n", perf.SuccessRate*100)
        fmt.Printf("  Avg Duration: %v\n", perf.AvgDuration)
    }
}

func createTestHost() *agent.Host {
    // 创建测试 Host
    // 实际实现中，你会从配置文件加载
    cfg := &agent.HostConfig{
        Name: "test-app",
        Models: []agent.ModelDefinition{
            {Name: "gpt-4"},
            {Name: "gpt-3.5-turbo"},
        },
        Agents: []agent.AgentDefinition{
            {Name: "developer", Kind: "chat", Model: "gpt-4"},
            {Name: "tester", Kind: "chat", Model: "gpt-3.5-turbo"},
            {Name: "reviewer", Kind: "chat", Model: "gpt-4"},
        },
    }

    host, err := agent.NewHost(ctx, cfg, nil, nil)
    if err != nil {
        log.Fatalf("Failed to create host: %v", err)
    }

    return host
}
```

---

## 最佳实践

### 1. 合理设置优先级

```go
// 高优先级任务（关键功能）
highPriorityTask := &collaboration.CollaborativeTask{
    Priority: 8, // 8-9 为高优先级
    // ...
}

// 中等优先级任务（常规功能）
normalTask := &collaborative.CollaborativeTask{
    Priority: 5, // 4-6 为中等优先级
    // ...
}

// 低优先级任务（优化类）
lowPriorityTask := &collaborative.CollaborativeTask{
    Priority: 2, // 1-3 为低优先级
    // ...
}
```

### 2. 使用超时保护

```go
task := &collaborative.CollaborativeTask{
    Timeout: 30 * time.Second, // 始终设置超时
    // ...
}
```

### 3. 合理定义能力

```go
// 细粒度的能力定义
capabilities := []string{
    "coding",
    "review",
    "debugging",
    "testing",
    "documentation",
}
```

### 4. 监控性能

```go
// 定期检查性能统计
stats := team.GetPerformanceStats()
for name, perf := range stats.MemberStats {
    if perf.ErrorRate > 0.1 { // 错误率超过 10%
        log.Printf("Warning: %s has high error rate: %.2f%%",
            name, perf.ErrorRate*100)
    }
}
```

---

## API 参考

详细的 API 文档请参考：
- [AgentTeam API](./agent_team.go)
- [IntelligentRouter API](./router.go)
- [ConsensusManager API](./consensus.go)
- [Orchestrator API](./orchestration.go)
