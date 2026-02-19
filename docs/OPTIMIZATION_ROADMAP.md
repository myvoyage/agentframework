# AgentFramework 优化路线图

**版本**: v1.0
**日期**: 2026-02-19
**状态**: 待执行

---

## 📊 优先级矩阵

```
高影响 │ 🔴 P1: 性能优化      🔴 P1: 内存泄漏
      │ 🔴 P1: Goroutine管理  🟡 P2: 错误处理
──────┼────────────────────────────────
低影响 │ 🟡 P2: 日志系统      🟢 P3: 配置管理
      │ 🟢 P3: 文档完善       🟢 P3: 依赖注入
      └───────────────────────────────
        高投入                    低投入
```

---

## 🔴 Phase 1: 紧急优化 (Week 1-2)

### 1.1 性能瓶颈修复

#### 问题: DAG 工作流锁竞争

**位置**: [pkg/framework/workflow/workflow_dag.go:45-120](pkg/framework/workflow/workflow_dag.go#L45-L120)

**当前实现**:
```go
type WorkflowDAG struct {
    nodes map[string]*Node
    edges map[string][]string
    mu    sync.Mutex
}

func (dag *WorkflowDAG) AddNode(node *Node) {
    dag.mu.Lock()
    defer dag.mu.Unlock()
    dag.nodes[node.ID] = node
}
```

**优化方案**:
```go
type WorkflowDAG struct {
    nodes sync.Map  // map[string]*Node
    edges sync.Map  // map[string][]string
}

func (dag *WorkflowDAG) AddNode(node *Node) {
    dag.nodes.Store(node.ID, node)
}

func (dag *WorkflowDAG) GetNode(id string) (*Node, bool) {
    value, ok := dag.nodes.Load(id)
    if !ok {
        return nil, false
    }
    return value.(*Node), true
}
```

**预期效果**:
- 吞吐量提升: **50-70%**
- 延迟降低: **30-40%**
- CPU 使用率降低: **20%**

#### 问题: 实时代理内存泄漏

**位置**: [pkg/framework/agent/realtime_agent.go:89-134](pkg/framework/agent/realtime_agent.go#L89-L134)

**当前实现**:
```go
func (a *RealtimeAgent) Start(ctx context.Context) error {
    for _, event := range events {
        sub := a.eventBus.Subscribe(event)
        go func() {
            for msg := range sub {
                a.handleMessage(msg)
            }
        }()
    }
    return nil
}
```

**优化方案**:
```go
func (a *RealtimeAgent) Start(ctx context.Context) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    for _, event := range events {
        sub := a.eventBus.Subscribe(event)
        go func() {
            defer close(sub)
            for {
                select {
                case msg := <-sub:
                    a.handleMessage(msg)
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    <-ctx.Done()
    return nil
}
```

**预期效果**:
- 消除内存泄漏
- 优雅关闭支持
- 资源正确释放

#### 问题: Goroutine 泄漏

**位置**: [agent/workflow.go:234-289](agent/workflow.go#L234-L289)

**优化方案**:
```go
func (w *Workflow) Execute(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, w.timeout)
    defer cancel()

    errChan := make(chan error, len(w.tasks))

    for _, task := range w.tasks {
        go func(t *Task) {
            defer func() {
                if r := recover(); r != nil {
                    errChan <- fmt.Errorf("panic: %v", r)
                }
            }()

            if err := t.Execute(ctx); err != nil {
                errChan <- err
            } else {
                errChan <- nil
            }
        }(task)
    }

    // 收集结果
    for i := 0; i < len(w.tasks); i++ {
        if err := <-errChan; err != nil && w.failFast {
            return err  // cancel context will stop other goroutines
        }
    }

    return nil
}
```

### 1.2 内存优化

#### 使用对象池

```go
// pkg/framework/workflow/pool.go
package workflow

var (
    nodePool = sync.Pool{
        New: func() interface{} {
            return &Node{
                status: StatusPending,
                inputs: make(map[string]interface{}),
            }
        },
    }

    taskPool = sync.Pool{
        New: func() interface{} {
            return &Task{
                ctx: context.Background(),
            }
        },
    }
)

func AcquireNode() *Node {
    node := nodePool.Get().(*Node)
    node.Reset()
    return node
}

func ReleaseNode(node *Node) {
    nodePool.Put(node)
}
```

**预期效果**:
- 内存分配减少: **40-50%**
- GC 压力降低: **60%**
- 性能提升: **15-20%**

### 1.3 I/O 优化

#### 批量操作

```go
// pkg/storage/batch_writer.go
package storage

type BatchWriter struct {
    bufferSize int
    buffer     []Entry
    flushTimer *time.Timer
    mu         sync.Mutex
}

func (bw *BatchWriter) Write(entry Entry) error {
    bw.mu.Lock()
    defer bw.mu.Unlock()

    bw.buffer = append(bw.buffer, entry)

    if len(bw.buffer) >= bw.bufferSize {
        return bw.flush()
    }

    return nil
}

func (bw *BatchWriter) flush() error {
    if len(bw.buffer) == 0 {
        return nil
    }

    // 批量写入
    if err := bw.storage.WriteBatch(bw.buffer); err != nil {
        return err
    }

    bw.buffer = bw.buffer[:0]
    return nil
}
```

**预期效果**:
- I/O 操作减少: **80%**
- 吞吐量提升: **3-5x**
- 延迟降低: **50%**

---

## 🟡 Phase 2: 质量提升 (Week 3-4)

### 2.1 错误处理统一

#### 实现标准错误类型

```go
// pkg/errors/types.go
package errors

import "fmt"

// ErrorCode 错误码类型
type ErrorCode string

const (
    // 通用错误
    ErrCodeUnknown      ErrorCode = "UNKNOWN"
    ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
    ErrCodeNotFound     ErrorCode = "NOT_FOUND"
    ErrCodePermission   ErrorCode = "PERMISSION_DENIED"

    // Agent 错误
    ErrCodeAgentStart    ErrorCode = "AGENT_START_FAILED"
    ErrCodeAgentStop     ErrorCode = "AGENT_STOP_FAILED"
    ErrCodeAgentExecute  ErrorCode = "AGENT_EXECUTE_FAILED"

    // Workflow 错误
    ErrCodeWorkflowInvalid ErrorCode = "WORKFLOW_INVALID"
    ErrCodeWorkflowExecute ErrorCode = "WORKFLOW_EXECUTE_FAILED"

    // 工具错误
    ErrCodeToolNotFound ErrorCode = "TOOL_NOT_FOUND"
    ErrCodeToolExecute  ErrorCode = "TOOL_EXECUTE_FAILED"
)

// AppError 应用错误
type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Stack   []stack.Frame
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// New 创建新错误
func New(code ErrorCode, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Stack:   stack.Capture(2),
    }
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *AppError {
    if err == nil {
        return nil
    }
    return &AppError{
        Code:    code,
        Message: message,
        Cause:   err,
        Stack:   stack.Capture(2),
    }
}

// Is 判断错误类型
func Is(err error, code ErrorCode) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Code == code
    }
    return false
}

// 便捷函数
func InvalidInput(format string, args ...interface{}) *AppError {
    return New(ErrCodeInvalidInput, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...interface{}) *AppError {
    return New(ErrCodeNotFound, fmt.Sprintf(format, args...))
}

func PermissionDenied(format string, args ...interface{}) *AppError {
    return New(ErrCodePermission, fmt.Sprintf(format, args...))
}
```

#### 使用示例

```go
// agent/service.go
package agent

import (
    "github.com/myvoyage/agentframework/pkg/errors"
)

func (s *Service) ExecuteAgent(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    // 验证输入
    if req.AgentID == "" {
        return nil, errors.InvalidInput("agent_id is required")
    }

    // 查找 Agent
    agent, err := s.registry.Get(req.AgentID)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrCodeNotFound, "agent not found")
    }

    // 执行
    result, err := agent.Execute(ctx, req.Input)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrCodeAgentExecute, "execution failed")
    }

    return &ExecuteResponse{Result: result}, nil
}
```

### 2.2 结构化日志

#### 实现统一日志接口

```go
// pkg/logging/logger.go
package logging

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger 日志接口
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    With(fields ...Field) Logger
    WithErr(err error) Logger
}

// Field 日志字段
type Field = zap.Field

// 字段构造函数
var (
    String  = zap.String
    Int     = zap.Int
    Int64   = zap.Int64
    Float64 = zap.Float64
    Bool    = zap.Bool
    Any     = zap.Any
    Err     = zap.NamedError
    Duration = zap.Duration
)

// logger 实现
type logger struct {
    *zap.Logger
}

// New 创建新日志器
func New(level string, development bool) (Logger, error) {
    config := zap.Config{
        Level:            zap.NewAtomicLevelAt(getLogLevel(level)),
        Development:      development,
        Encoding:         "json",
        EncoderConfig:    zapcore.EncoderConfig{
            TimeKey:        "ts",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            FunctionKey:    zapcore.OmitKey,
            MessageKey:     "msg",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }

    zapLogger, err := config.Build()
    if err != nil {
        return nil, err
    }

    return &logger{Logger: zapLogger}, nil
}

func getLogLevel(level string) zapcore.Level {
    switch level {
    case "debug":
        return zapcore.DebugLevel
    case "info":
        return zapcore.InfoLevel
    case "warn":
        return zapcore.WarnLevel
    case "error":
        return zapcore.ErrorLevel
    default:
        return zapcore.InfoLevel
    }
}

func (l *logger) Debug(msg string, fields ...Field) {
    l.Logger.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
    l.Logger.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
    l.Logger.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
    l.Logger.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
    l.Logger.Fatal(msg, fields...)
}

func (l *logger) With(fields ...Field) Logger {
    return &logger{l.Logger.With(fields...)}
}

func (l *logger) WithErr(err error) Logger {
    return &logger{l.Logger.With(zap.NamedError("error", err))}
}

// 全局日志器
var log Logger

// Init 初始化全局日志器
func Init(level string, development bool) error {
    var err error
    log, err = New(level, development)
    return err
}

// L 获取全局日志器
func L() Logger {
    if log == nil {
        // 默认配置
        log, _ = New("info", false)
    }
    return log
}
```

#### 使用示例

```go
// agent/service.go
package agent

import (
    "github.com/myvoyage/agentframework/pkg/logging"
)

func (s *Service) ExecuteAgent(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    log := logging.L()

    log.Info("Executing agent",
        logging.String("agent_id", req.AgentID),
        logging.String("workflow", req.WorkflowName),
    )

    agent, err := s.registry.Get(req.AgentID)
    if err != nil {
        log.Error("Agent not found",
            logging.String("agent_id", req.AgentID),
            logging.Err(err),
        )
        return nil, err
    }

    start := time.Now()
    result, err := agent.Execute(ctx, req.Input)
    if err != nil {
        log.Error("Agent execution failed",
            logging.String("agent_id", req.AgentID),
            logging.Err(err),
            logging.Duration("elapsed", time.Since(start)),
        )
        return nil, err
    }

    log.Info("Agent executed successfully",
        logging.String("agent_id", req.AgentID),
        logging.Duration("elapsed", time.Since(start)),
    )

    return &ExecuteResponse{Result: result}, nil
}
```

### 2.3 测试增强

#### 补充边界测试

```go
// pkg/framework/workflow/workflow_dag_boundary_test.go
package workflow

import (
    "context"
    "testing"
    "time"
)

func TestWorkflowDAG_BoundaryConditions(t *testing.T) {
    tests := []struct {
        name        string
        setupDAG    func() *WorkflowDAG
        expectError bool
        description string
    }{
        {
            name: "Empty workflow",
            setupDAG: func() *WorkflowDAG {
                return NewWorkflowDAG()
            },
            expectError: true,
            description: "Should fail when executing empty DAG",
        },
        {
            name: "Single node",
            setupDAG: func() *WorkflowDAG {
                dag := NewWorkflowDAG()
                dag.AddNode(&Node{ID: "1", Task: &MockTask{}})
                return dag
            },
            expectError: false,
            description: "Should execute single node successfully",
        },
        {
            name: "Circular dependency",
            setupDAG: func() *WorkflowDAG {
                dag := NewWorkflowDAG()
                dag.AddNode(&Node{ID: "1", Task: &MockTask{}})
                dag.AddNode(&Node{ID: "2", Task: &MockTask{}})
                dag.AddEdge("1", "2")
                dag.AddEdge("2", "1")  // 循环
                return dag
            },
            expectError: true,
            description: "Should detect and reject circular dependencies",
        },
        {
            name: "Large workflow (10000 nodes)",
            setupDAG: func() *WorkflowDAG {
                dag := NewWorkflowDAG()
                for i := 0; i < 10000; i++ {
                    dag.AddNode(&Node{
                        ID:   fmt.Sprintf("node-%d", i),
                        Task: &MockTask{},
                    })
                }
                return dag
            },
            expectError: false,
            description: "Should handle large workflows efficiently",
        },
        {
            name: "Deep chain (1000 levels)",
            setupDAG: func() *WorkflowDAG {
                dag := NewWorkflowDAG()
                prevID := ""
                for i := 0; i < 1000; i++ {
                    id := fmt.Sprintf("node-%d", i)
                    dag.AddNode(&Node{ID: id, Task: &MockTask{}})
                    if prevID != "" {
                        dag.AddEdge(prevID, id)
                    }
                    prevID = id
                }
                return dag
            },
            expectError: false,
            description: "Should handle deep dependency chains",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dag := tt.setupDAG()

            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            err := dag.Execute(ctx)

            if tt.expectError {
                if err == nil {
                    t.Errorf("Expected error but got none")
                }
            } else {
                if err != nil {
                    t.Errorf("Unexpected error: %v", err)
                }
            }
        })
    }
}
```

#### 性能基准测试

```go
// pkg/framework/workflow/workflow_dag_bench_test.go
package workflow

import (
    "context"
    "testing"
    "time"
)

func BenchmarkWorkflowDAG_Execute_Small(b *testing.B) {
    dag := setupDAG(10)  // 10 nodes
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dag.Execute(ctx)
    }
}

func BenchmarkWorkflowDAG_Execute_Medium(b *testing.B) {
    dag := setupDAG(100)  // 100 nodes
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dag.Execute(ctx)
    }
}

func BenchmarkWorkflowDAG_Execute_Large(b *testing.B) {
    dag := setupDAG(1000)  // 1000 nodes
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dag.Execute(ctx)
    }
}

func BenchmarkWorkflowDAG_ConcurrentAccess(b *testing.B) {
    dag := setupDAG(100)
    ctx := context.Background()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            dag.Execute(ctx)
        }
    })
}

func setupDAG(size int) *WorkflowDAG {
    dag := NewWorkflowDAG()
    for i := 0; i < size; i++ {
        dag.AddNode(&Node{
            ID:   fmt.Sprintf("node-%d", i),
            Task: &MockTask{},
        })
    }
    return dag
}
```

---

## 🟢 Phase 3: 架构改进 (Month 2)

### 3.1 依赖注入框架

```go
// pkg/inject/container.go
package inject

import (
    "reflect"
    "sync"
)

// Container 依赖容器
type Container struct {
    providers map[string]interface{}
    singletons map[string]interface{}
    mutex     sync.RWMutex
}

// New 创建新容器
func New() *Container {
    return &Container{
        providers:  make(map[string]interface{}),
        singletons: make(map[string]interface{}),
    }
}

// Register 注册提供者
func (c *Container) Register(name string, provider interface{}) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // 验证提供者类型
    if err := c.validateProvider(provider); err != nil {
        return err
    }

    c.providers[name] = provider
    return nil
}

// RegisterSingleton 注册单例
func (c *Container) RegisterSingleton(name string, instance interface{}) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.singletons[name] = instance
}

// Resolve 解析依赖
func (c *Container) Resolve(name string) (interface{}, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    // 检查单例
    if instance, ok := c.singletons[name]; ok {
        return instance, nil
    }

    // 检查提供者
    provider, ok := c.providers[name]
    if !ok {
        return nil, fmt.Errorf("provider not found: %s", name)
    }

    // 调用提供者
    return c.callProvider(provider)
}

// ResolveTo 解析到目标
func (c *Container) ResolveTo(name string, target interface{}) error {
    instance, err := c.Resolve(name)
    if err != nil {
        return err
    }

    targetValue := reflect.ValueOf(target)
    if targetValue.Kind() != reflect.Ptr {
        return fmt.Errorf("target must be a pointer")
    }

    instanceValue := reflect.ValueOf(instance)
    targetValue.Elem().Set(instanceValue)

    return nil
}

func (c *Container) validateProvider(provider interface{}) error {
    providerType := reflect.TypeOf(provider)

    if providerType.Kind() != reflect.Func {
        return fmt.Errorf("provider must be a function")
    }

    // 验证返回值
    if providerType.NumOut() == 0 {
        return fmt.Errorf("provider must return at least one value")
    }

    return nil
}

func (c *Container) callProvider(provider interface{}) (interface{}, error) {
    providerValue := reflect.ValueOf(provider)
    results := providerValue.Call(nil)

    if len(results) == 1 {
        return results[0].Interface(), nil
    }

    // 处理 (result, error) 模式
    if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
        return nil, err
    }

    return results[0].Interface(), nil
}
```

#### 使用示例

```go
// agent/app.go
package agent

import (
    "github.com/myvoyage/agentframework/pkg/inject"
)

func SetupContainer() (*inject.Container, error) {
    container := inject.New()

    // 注册配置
    container.RegisterSingleton("config", &Config{
        LogLevel: "info",
    })

    // 注册日志
    container.Register("logger", func() (*logging.Logger, error) {
        config := container.Resolve("config").(*Config)
        return logging.New(config.LogLevel, false)
    })

    // 注册数据库
    container.Register("database", func() (*Database, error) {
        logger := container.Resolve("logger").(*logging.Logger)
        return NewDatabase(logger)
    })

    // 注册 Agent 服务
    container.Register("agent.service", func() (*Service, error) {
        logger := container.Resolve("logger").(*logging.Logger)
        db := container.Resolve("database").(*Database)
        return NewService(logger, db)
    })

    return container, nil
}
```

### 3.2 监控系统

#### Prometheus 指标

```go
// pkg/monitoring/metrics.go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Workflow 指标
    workflowExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentframework_workflow_executions_total",
            Help: "Total number of workflow executions",
        },
        []string{"workflow_type", "status"},
    )

    workflowDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "agentframework_workflow_duration_seconds",
            Help:    "Workflow execution duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"workflow_type"},
    )

    workflowActiveGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agentframework_workflow_active",
            Help: "Number of currently active workflow executions",
        },
        []string{"workflow_type"},
    )

    // Agent 指标
    agentExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentframework_agent_executions_total",
            Help: "Total number of agent executions",
        },
        []string{"agent_type", "status"},
    )

    agentDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "agentframework_agent_duration_seconds",
            Help:    "Agent execution duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"agent_type"},
    )

    // 工具指标
    toolExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentframework_tool_executions_total",
            Help: "Total number of tool executions",
        },
        []string{"tool_name", "status"},
    )

    toolDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "agentframework_tool_duration_seconds",
            Help:    "Tool execution duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"tool_name"},
    )

    // 系统指标
    goroutinesGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "agentframework_goroutines",
            Help: "Current number of goroutines",
        },
    )

    memoryGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "agentframework_memory_bytes",
            Help: "Current memory usage in bytes",
        },
    )
)

// Workflow 指标记录函数
func RecordWorkflowExecution(workflowType, status string) {
    workflowExecutions.WithLabelValues(workflowType, status).Inc()
}

func RecordWorkflowDuration(workflowType string, duration float64) {
    workflowDuration.WithLabelValues(workflowType).Observe(duration)
}

func IncrementWorkflowActive(workflowType string) {
    workflowActiveGauge.WithLabelValues(workflowType).Inc()
}

func DecrementWorkflowActive(workflowType string) {
    workflowActiveGauge.WithLabelValues(workflowType).Dec()
}

// Agent 指标记录函数
func RecordAgentExecution(agentType, status string) {
    agentExecutions.WithLabelValues(agentType, status).Inc()
}

func RecordAgentDuration(agentType string, duration float64) {
    agentDuration.WithLabelValues(agentType).Observe(duration)
}

// 工具指标记录函数
func RecordToolExecution(toolName, status string) {
    toolExecutions.WithLabelValues(toolName, status).Inc()
}

func RecordToolDuration(toolName string, duration float64) {
    toolDuration.WithLabelValues(toolName).Observe(duration)
}

// 系统指标记录函数
func RecordSystemMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    goroutinesGauge.Set(float64(runtime.NumGoroutine()))
    memoryGauge.Set(float64(m.Alloc))
}
```

#### 使用示例

```go
// agent/service.go
package agent

import (
    "github.com/myvoyage/agentframework/pkg/monitoring"
)

func (s *Service) ExecuteAgent(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    agentType := req.Type
    monitoring.RecordAgentExecution(agentType, "started")
    monitoring.IncrementAgentActive(agentType)
    defer monitoring.DecrementAgentActive(agentType)

    start := time.Now()

    result, err := s.agent.Execute(ctx, req.Input)

    duration := time.Since(start).Seconds()
    monitoring.RecordAgentDuration(agentType, duration)

    if err != nil {
        monitoring.RecordAgentExecution(agentType, "failed")
        return nil, err
    }

    monitoring.RecordAgentExecution(agentType, "success")
    return &ExecuteResponse{Result: result}, nil
}
```

### 3.3 健康检查

```go
// pkg/health/check.go
package health

import (
    "context"
    "sync"
    "time"
)

// Status 健康状态
type Status string

const (
    StatusHealthy   Status = "healthy"
    StatusUnhealthy Status = "unhealthy"
    StatusDegraded  Status = "degraded"
)

// Check 检查函数
type Check func(ctx context.Context) error

// CheckResult 检查结果
type CheckResult struct {
    Name      string    `json:"name"`
    Status    Status    `json:"status"`
    Message   string    `json:"message,omitempty"`
    Duration  int64     `json:"duration_ms"`
    Timestamp time.Time `json:"timestamp"`
}

// Checker 健康检查器
type Checker struct {
    checks map[string]Check
    mutex  sync.RWMutex
}

// New 创建新检查器
func New() *Checker {
    return &Checker{
        checks: make(map[string]Check),
    }
}

// Register 注册检查
func (c *Checker) Register(name string, check Check) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.checks[name] = check
}

// Unregister 取消注册
func (c *Checker) Unregister(name string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    delete(c.checks, name)
}

// Check 执行所有检查
func (c *Checker) Check(ctx context.Context) map[string]CheckResult {
    c.mutex.RLock()
    checks := make(map[string]Check)
    for name, check := range c.checks {
        checks[name] = check
    }
    c.mutex.RUnlock()

    results := make(map[string]CheckResult)

    for name, check := range checks {
        result := c.runCheck(ctx, name, check)
        results[name] = result
    }

    return results
}

func (c *Checker) runCheck(ctx context.Context, name string, check Check) CheckResult {
    start := time.Now()
    result := CheckResult{
        Name:      name,
        Timestamp: start,
    }

    err := check(ctx)
    result.Duration = time.Since(start).Milliseconds()

    if err != nil {
        result.Status = StatusUnhealthy
        result.Message = err.Error()
    } else {
        result.Status = StatusHealthy
    }

    return result
}

// 整体健康状态
func (c *Checker) OverallStatus(results map[string]CheckResult) Status {
    unhealthyCount := 0

    for _, result := range results {
        if result.Status == StatusUnhealthy {
            unhealthyCount++
        }
    }

    if unhealthyCount == 0 {
        return StatusHealthy
    }

    if unhealthyCount == len(results) {
        return StatusUnhealthy
    }

    return StatusDegraded
}
```

---

## 📅 实施时间表

### Week 1-2: 紧急优化

| 任务 | 负责人 | 预计时间 | 优先级 |
|------|--------|----------|--------|
| 修复 DAG 锁竞争 | @core-team | 3天 | 🔴 P0 |
| 修复内存泄漏 | @core-team | 2天 | 🔴 P0 |
| 优化 Goroutine 管理 | @core-team | 2天 | 🔴 P0 |
| 实现对象池 | @performance-team | 2天 | 🟡 P1 |

### Week 3-4: 质量提升

| 任务 | 负责人 | 预计时间 | 优先级 |
|------|--------|----------|--------|
| 统一错误处理 | @quality-team | 3天 | 🟡 P1 |
| 实现结构化日志 | @quality-team | 3天 | 🟡 P1 |
| 补充边界测试 | @test-team | 4天 | 🟡 P1 |
| 添加性能测试 | @performance-team | 2天 | 🟢 P2 |

### Month 2: 架构改进

| 任务 | 负责人 | 预计时间 | 优先级 |
|------|--------|----------|--------|
| 实现依赖注入 | @architecture-team | 1周 | 🟢 P2 |
| 添加监控系统 | @ops-team | 1周 | 🟢 P2 |
| 实现健康检查 | @ops-team | 3天 | 🟢 P2 |
| 性能优化 | @performance-team | 1周 | 🟢 P2 |

---

## 🎯 成功指标

### 性能指标

| 指标 | 当前 | 目标 | 测量方法 |
|------|------|------|----------|
| 工作流吞吐量 | ~100/s | 1000/s | 基准测试 |
| 平均响应时间 | - | <100ms | APM 监控 |
| P99 响应时间 | - | <500ms | APM 监控 |
| 内存使用 | - | <512MB | Prometheus |
| CPU 使用率 | - | <50% | Prometheus |

### 质量指标

| 指标 | 当前 | 目标 | 测量方法 |
|------|------|------|----------|
| 测试覆盖率 | 65-70% | 80%+ | go test |
| 代码重复率 | ~5% | <3% | 代码审查 |
| 平均修复时间 | - | <2天 | Issue 跟踪 |
| 代码审查通过率 | - | >90% | PR 统计 |

### 可靠性指标

| 指标 | 当前 | 目标 | 测量方法 |
|------|------|------|----------|
| 系统可用性 | - | 99.9% | Uptime 监控 |
| 错误率 | - | <0.1% | 日志分析 |
| 平均恢复时间 | - | <5min | 监控告警 |

---

## 📚 参考资料

### Go 最佳实践

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

### 性能优化

- [Go Performance Tips](https://github.com/dgryski/go-perfbook)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Go Benchmarking](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

### 监控和可观测性

- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Structured Logging in Go](https://go.dev/blog/slog)

---

**文档版本**: v1.0
**最后更新**: 2026-02-19
**维护者**: AgentFramework Team
