# AgentFramework 代码重构指南

**文档版本**: 1.0.0
**创建日期**: 2025-02-19
**适用范围**: 全系统代码重构

---

## 📋 重构原则

### 1. SOLID 原则应用

#### 单一职责原则 (SRP)

**原则**: 一个类或模块应该只有一个改变的理由。

**示例**:
```go
// ❌ 错误：承担过多职责
type AgentManager struct {
    agents    map[string]*Agent
    config    Config
    db        Database
    cache     Cache
    metrics   Metrics
    logger    Logger
}

// ✅ 正确：职责分离
type AgentRegistry struct {
    agents map[string]*Agent
}

type AgentConfig struct {
    config Config
}

type AgentStorage struct {
    db Database
}

type AgentCache struct {
    cache Cache
}
```

#### 开闭原则 (OCP)

**原则**: 对扩展开放，对修改关闭。

**示例**:
```go
// ✅ 使用接口扩展
type Agent interface {
    Run(ctx context.Context, input string) (*Message, error)
}

// 新功能通过新接口添加，不修改现有代码
type StreamAgent interface {
    Agent
    Stream(ctx context.Context, input string) (<-chan *Message, error)
}
```

#### 里氏替换原则 (LSP)

**原则**: 子类型必须能够替换父类型。

**示例**:
```go
// ✅ 确保实现遵循接口契约
type BasicAgent struct{}

func (a *BasicAgent) Run(ctx context.Context, input string) (*Message, error) {
    // 总是返回有效的 Message 或错误
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    // ...
}

// ✅ 任何使用 Agent 的地方都可以用 BasicAgent 替换
```

#### 接口隔离原则 (ISP)

**原则**: 客户端不应该依赖它不需要的接口。

**示例**:
```go
// ❌ 错误：大而全的接口
type Agent interface {
    Run(ctx context.Context, input string) (*Message, error)
    Stop() error
    GetStatus() Status
    GetMetrics() Metrics
    SetConfig(config Config) error
    GetConfig() Config
    Save(path string) error
    Load(path string) error
    // ... 太多方法
}

// ✅ 正确：拆分为小接口
type Runnable interface {
    Run(ctx context.Context, input string) (*Message, error)
}

type Stoppable interface {
    Stop() error
}

type StatusProvider interface {
    GetStatus() Status
}

type Configurable interface {
    SetConfig(config Config) error
    GetConfig() Config
}

// Agent 组合需要的小接口
type Agent interface {
    Runnable
    Stoppable
    StatusProvider
    Configurable
}
```

#### 依赖倒置原则 (DIP)

**原则**: 高层模块不应该依赖低层模块，两者都应该依赖抽象。

**示例**:
```go
// ✅ 依赖抽象而非具体实现
type WorkflowManager struct {
    modelFactory ModelFactory  // 抽象
    skillLibrary SkillLibrary  // 抽象
    storage      Storage       // 抽象
}

// ❌ 避免依赖具体实现
type WorkflowManager struct {
    openAIModel *OpenAIModel      // 具体实现
    redisCache  *RedisCache       // 具体实现
    postgresDB  *PostgresStorage  // 具体实现
}
```

---

## 🔧 重构模式

### 1. 提取方法 (Extract Method)

**场景**: 一个函数过长（>100行）或包含重复逻辑。

**重构前**:
```go
func (a *Agent) Process(input string) (*Result, error) {
    // 50行验证逻辑
    // 50行处理逻辑
    // 50行清理逻辑
}
```

**重构后**:
```go
func (a *Agent) Process(input string) (*Result, error) {
    if err := a.validateInput(input); err != nil {
        return nil, err
    }

    result, err := a.processInput(input)
    if err != nil {
        return nil, err
    }

    a.cleanup()
    return result, nil
}

func (a *Agent) validateInput(input string) error {
    // 验证逻辑
}

func (a *Agent) processInput(input string) (*Result, error) {
    // 处理逻辑
}

func (a *Agent) cleanup() {
    // 清理逻辑
}
```

### 2. 提取类 (Extract Class)

**场景**: 一个类承担过多职责。

**重构前**:
```go
// agent.go: 3623行（过于庞大）
type Agent struct {
    // 执行相关
    model     Model
    tools     Tools
    // 配置相关
    config    Config
    validator Validator
    // 监控相关
    metrics   Metrics
    tracer    Tracer
    // 状态相关
    state     State
    history   History
    // ... 太多职责
}
```

**重构后**:
```
agent/
├── agent.go           // 核心Agent定义
├── agent_executor.go  // 执行逻辑
├── agent_config.go    // 配置管理
├── agent_monitor.go   // 监控和指标
├── agent_state.go     // 状态管理
└── agent_history.go   // 历史记录
```

### 3. 引入参数对象 (Introduce Parameter Object)

**场景**: 函数参数过多（>5个）。

**重构前**:
```go
func CreateAgent(
    name string,
    model string,
    temperature float64,
    maxTokens int,
    tools []Tool,
    memory int,
) (*Agent, error) {
    // ...
}
```

**重构后**:
```go
type AgentConfig struct {
    Name        string
    Model       string
    Temperature float64
    MaxTokens   int
    Tools       []Tool
    Memory      int
}

func CreateAgent(config *AgentConfig) (*Agent, error) {
    // ...
}
```

### 4. 组合方法 (Compose Method)

**场景**: 多个函数执行相似的步骤序列。

**重构前**:
```go
func (a *Agent) Process1() error {
    if err := a.validate(); err != nil {
        return err
    }
    if err := a.prepare(); err != nil {
        return err
    }
    if err := a.execute(); err != nil {
        return err
    }
    return a.cleanup()
}

func (a *Agent) Process2() error {
    if err := a.validate(); err != nil {
        return err
    }
    if err := a.prepare(); err != nil {
        return err
    }
    if err := a.executeAlternative(); err != nil {
        return err
    }
    return a.cleanup()
}
```

**重构后**:
```go
func (a *Agent) Process1() error {
    return a.processWith(a.execute)
}

func (a *Agent) Process2() error {
    return a.processWith(a.executeAlternative)
}

func (a *Agent) processWith(execute func() error) error {
    steps := []func() error{
        a.validate,
        a.prepare,
        execute,
        a.cleanup,
    }

    for _, step := range steps {
        if err := step(); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 📏 代码规范

### 1. 命名规范

#### 文件命名
```
✅ agent.go
✅ agent_registry.go
✅ http_middleware.go

❌ Agent.go (不要大写)
❌ agentRegistry.go (不要驼峰)
❌ agent-registry.go (不要连字符)
```

#### 函数命名
```go
// ✅ 动词开头，描述性
func GetAgent(id string) (*Agent, error) {}
func ValidateConfig(config *Config) error {}
func ProcessInput(ctx context.Context, input string) (*Result, error) {}

// ❌ 避免使用
func agent(id string) (*Agent, error) {}         // 太短
func DoSomething(input string) error {}          // 不明确
func Handle() error {}                           // 太模糊
```

#### 变量命名
```go
// ✅ 清晰描述性
var agentRegistry map[string]*Agent
var maxRetryCount int
var connectionTimeout time.Duration

// ❌ 避免使用
var a map[string]*Agent          // 太短
var cnt int                      // 缩写不明确
var tmp1, tmp2 interface{}       // 临时变量命名
```

### 2. 函数规范

#### 函数长度
```go
// ✅ 推荐：单个函数 < 100行
func (a *Agent) Run(ctx context.Context, input string) (*Message, error) {
    // 清晰、简洁
}

// ❌ 避免：单个函数 > 100行
func (a *Agent) Run(ctx context.Context, input string) (*Message, error) {
    // 太长，难以理解和维护
    // 应该拆分为多个子函数
}
```

#### 参数数量
```go
// ✅ 推荐：参数 <= 5个
func Process(ctx context.Context, input string, opts Options) (*Result, error) {}

// ✅ 使用配置对象处理多个参数
func Process(config *ProcessConfig) (*Result, error) {}

// ❌ 避免：参数 > 5个
func Process(ctx context.Context, input string, model string, temp float64, maxTokens int, tools []Tool) (*Result, error) {}
```

### 3. 错误处理规范

#### 统一错误处理
```go
// ✅ 使用统一的错误处理器
errHandler := errors.NewHandler("agent", logger)
if err != nil {
    return errHandler.Handle("operation", err)
}

// ✅ 使用上下文包装错误
if err := validate(input); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// ❌ 避免直接忽略错误
func process(input string) {
    validate(input)  // 忽略错误
}
```

#### 错误传播
```go
// ✅ 正确的错误传播
func (a *Agent) Run(ctx context.Context, input string) (*Message, error) {
    if err := a.validate(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    // ...
}

// ❌ 避免吞掉错误
func (a *Agent) Run(ctx context.Context, input string) (*Message, error) {
    a.validate(input)  // 忽略错误
    // ...
}
```

### 4. 注释规范

#### 包注释
```go
// Package agent provides AI agent implementations including
// ReAct agents, skill agents, and chat agents.
//
// Basic usage:
//
//   agent := agent.NewReActAgent(
//       agent.WithModel("gpt-4"),
//       agent.WithSystemPrompt("You are a helpful assistant"),
//   )
//   result, err := agent.Run(ctx, "Hello")
package agent
```

#### 函数注释
```go
// NewReActAgent creates a new ReAct agent with default memory settings.
//
// Parameters:
//   - name: Unique identifier for the agent
//   - model: The language model to use
//   - opts: Optional configuration functions
//
// Example:
//
//   agent := NewReActAgent("my-agent", "gpt-4", WithMaxIterations(10))
//
// Returns a configured ReActAgent ready for use.
func NewReActAgent(name, model string, opts ...AgentOption) *ReActAgent {
    // ...
}
```

#### 复杂逻辑注释
```go
// ✅ 注释解释"为什么"而不是"做什么"
func (a *Agent) trimHistory() {
    // Keep last 10 messages to stay within context window
    // while maintaining conversation continuity
    if len(a.history) > 10 {
        a.history = a.history[len(a.history)-10:]
    }
}

// ❌ 避免无意义的注释
func (a *Agent) trimHistory() {
    // Trim history
    if len(a.history) > 10 {
        a.history = a.history[len(a.history)-10:]
    }
}
```

---

## 🎯 重构检查清单

### 重构前检查

- [ ] 是否有单元测试？
- [ ] 测试覆盖率是否足够？
- [ ] 是否理解现有代码逻辑？
- [ ] 是否评估了重构风险？
- [ ] 是否有回滚计划？

### 重构后检查

- [ ] 所有测试是否通过？
- [ ] 代码行数是否减少？
- [ ] 重复代码是否消除？
- [ ] 命名是否清晰？
- [ ] 注释是否充分？
- [ ] 性能是否保持或提升？

---

## 📊 重构效果评估

### 代码质量指标

| 指标 | 重构前 | 目标 | 测量方法 |
|------|--------|------|----------|
| 代码行数 | 基准 | -20% | wc -l |
| 圈复杂度 | 基准 | <10 | gocyclo |
| 重复率 | ~15% | <5% | dupl |
| 测试覆盖 | 68% | >75% | go test -cover |
| 函数长度 | 基准 | <100行 | 统计分析 |

### 性能指标

| 指标 | 重构前 | 目标 | 测量方法 |
|------|--------|------|----------|
| 执行时间 | 基准 | ≤基准 | benchmark |
| 内存使用 | 基准 | ≤基准 | pprof |
| GC 暂停 | 基准 | ≤基准 | pprof |
| 吞吐量 | 基准 | ≥基准 | benchmark |

---

## 🚀 重构流程

### 1. 准备阶段
```bash
# 1. 创建重构分支
git checkout -b refactor/phase3

# 2. 运行现有测试
go test ./...

# 3. 代码质量分析
gocyclo -over 15 .
dupl -threshold 50 -t ./...
```

### 2. 执行阶段
```bash
# 1. 小步重构
# 2. 每步提交
git add .
git commit -m "refactor: 提取统一错误处理"

# 3. 运行测试
go test ./...

# 4. 代码审查
git push origin refactor/phase3
```

### 3. 验证阶段
```bash
# 1. 性能测试
go test -bench=. -benchmem

# 2. 覆盖率测试
go test ./... -cover

# 3. 集成测试
go test -tags=integration ./...
```

---

## 📚 参考资料

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Clean Code](https://www.amazon.com/Clean-Code-Handbook-Software-Craftsmanship/dp/0132350882)
- [Refactoring](https://www.amazon.com/Refactoring-Improving-Existing-Addison-Wesley-Signature/dp/0201485672)

---

**文档维护**: 请在重构后更新本文档
**最后更新**: 2025-02-19
