# 最佳实践

> **AgentFramework 开发最佳实践指南**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [架构设计](#架构设计)
- [编码规范](#编码规范)
- [性能优化](#性能优化)
- [安全实践](#安全实践)
- [测试策略](#测试策略)
- [监控调试](#监控调试)
- [部署运维](#部署运维)

---

## 架构设计

### 1. 项目结构

**推荐的项目结构**:

```
my-agent-app/
├── main.go                    # 应用入口
├── config/                    # 配置文件
│   ├── config.yaml
│   └── host.yaml
├── agents/                   # 自定义 Agent
│   ├── my_agent.go
│   └── tools.go
├── workflows/                 # 工作流定义
│   └── main_workflow.go
├── skills/                   # 自定义 Skill
│   └── my_skill.go
├── pkg/                       # 私有包
│   ├── util/
│   └── model/
└── tests/                     # 测试文件
    ├── agents_test.go
    └── workflows_test.go
```

### 2. 组件职责分离

**遵循单一职责原则**:

```go
// ❌ 错误示例：一个类承担多个职责
type BadAgent struct {
    // 既处理对话
    // 又处理数据库
    // 还处理文件操作
}

// ✅ 正确示例：职责分离
type ChatAgent struct {
    // 只处理对话
}

type DataStore struct {
    // 只处理数据存储
}

type FileService struct {
    // 只处理文件操作
}
```

### 3. 依赖注入

**使用接口和依赖注入**:

```go
// ✅ 正确：依赖接口
type Agent struct {
    model  ChatModel    // 依赖接口，不依赖具体实现
    store  Storage     // 依赖接口
}

// ✅ 正确：通过构造函数注入
func NewAgent(model ChatModel, store Storage) *Agent {
    return &Agent{
        model: model,
        store: store,
    }
}

// ❌ 错误：硬编码依赖
func NewAgent() *Agent {
    model := NewOpenAIModel()     // 硬编码具体实现
    store := NewSQLStore()        // 硬编码具体实现
    return &Agent{
        model: model,
        store: store,
    }
}
```

### 4. 配置管理

**使用配置文件而非硬编码**:

```go
// ❌ 错误：硬编码配置
func NewAgent() *Agent {
    return &Agent{
        apiKey: "sk-1234567890",     // 硬编码
        baseURL: "https://api.example.com",
        timeout: 30,
    }
}

// ✅ 正确：从配置文件读取
func NewAgent(cfg *AgentConfig) *Agent {
    return &Agent{
        apiKey: cfg.APIKey,         // 从配置读取
        baseURL: cfg.BaseURL,
        timeout: cfg.Timeout,
    }
}
```

---

## 编码规范

### 1. 错误处理

**使用清晰的错误信息**:

```go
// ❌ 错误示例
return errors.New("error")

// ✅ 正确示例
return fmt.Errorf("failed to create agent: %w", err)
return fmt.Errorf("model %s not found", modelName)
return fmt.Errorf("invalid parameter: %s must be positive", paramName)
```

**使用错误包装**:

```go
// ✅ 正确：包装错误以保留上下文
func (s *Service) DoSomething() error {
    result, err := s.client.CallAPI()
    if err != nil {
        return fmt.Errorf("failed to call API: %w", err)
    }
    // ... 处理结果
}
```

### 2. 日志记录

**使用结构化日志**:

```go
import "log/slog"

// ✅ 正确：使用结构化日志
func (s *Service) DoSomething(ctx context.Context) {
    logger := slog.FromContext(ctx)

    logger.Info("starting operation",
        "operation", "create_agent",
        "agentName", "my-agent",
    )

    result, err := s.doSomethingInternal()
    if err != nil {
        logger.Error("operation failed",
            "operation", "create_agent",
            "error", err,
        )
        return err
    }

    logger.Info("operation completed",
        "operation", "create_agent",
        "duration", time.Since(start),
    )
}
```

### 3. 上下文管理

**正确使用 Context**:

```go
// ❌ 错误：忽略 Context
func (s *Service) DoSomething() error {
    // 没有使用 ctx，无法取消或设置超时
    return s.longRunningOperation()
}

// ✅ 正确：传递 Context
func (s *Service) DoSomething(ctx context.Context) error {
    // 可以设置超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 可以检查取消
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return s.longRunningOperationWithContext(ctx)
    }
}
```

### 4. 并发安全

**使用适当的同步机制**:

```go
// ✅ 正确：使用 sync.RWMutex 保护共享状态
type SafeStore struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (s *SafeStore) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    val, ok := s.cache[key]
    return val, ok
}

func (s *SafeStore) Set(key string, val interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.cache[key] = val
}
```

### 5. 资源管理

**使用 defer 确保资源释放**:

```go
// ✅ 正确：使用 defer
func (s *Service) ProcessFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()    // 确保文件被关闭

    // 处理文件
    return processFile(file)
}
```

---

## 性能优化

### 1. 连接池

**使用连接池减少创建开销**:

```go
// ✅ 正确：使用连接池
type PoolManager struct {
    httpPool *HTTPPool
    dbPool  *DBPool
}

func (p *PoolManager) DoRequest(ctx context.Context) error {
    // 从池中获取连接
    conn, err := p.httpPool.Get(ctx)
    if err != nil {
        return err
    }
    defer p.httpPool.Put(conn)

    // 使用连接执行请求
    return conn.Do(ctx)
}
```

### 2. 缓存策略

**使用缓存减少重复计算**:

```go
// ✅ 正确：使用缓存
type CacheManager struct {
    cache *skills.SkillsCache
}

func (c *CacheManager) GetSkill(ctx context.Context, name string) (Skill, error) {
    // 尝试从缓存获取
    if skill, ok := c.cache.Get(name); ok {
        return skill, nil
    }

    // 缓存未命中，加载数据库
    skill, err := c.loadSkillFromDB(ctx, name)
    if err != nil {
        return nil, err
    }

    // 存入缓存
    c.cache.Set(name, skill)
    return skill, nil
}
```

### 3. 批量处理

**使用批量操作减少网络往返**:

```go
// ❌ 错误：逐个处理
func (s *Service) ProcessItems(items []string) error {
    for _, item := range items {
        if err := s.processItem(item); err != nil {
            return err
        }
    }
}

// ✅ 正确：批量处理
func (s *Service) ProcessItems(items []string) error {
    return s.client.BatchProcess(items)
}
```

### 4. 流式处理

**使用流式处理处理大量数据**:

```go
// ✅ 正确：使用流式 API
func (s *Service) ProcessLargeFile(ctx context.Context, path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if err := s.processLine(ctx, line); err != nil {
            return err
        }
    }
}
```

### 5. 并发处理

**合理使用并发提高性能**:

```go
// ✅ 正确：使用 WaitGroup 并发处理
func (s *Service) ProcessItemsConcurrent(items []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(itm string) {
            defer wg.Done()
            if err := s.processItem(itm); err != nil {
                errChan <- err
            }
        }(item)
    }

    go func() {
        wg.Wait()
        close(errChan)
    }()

    for err := range errChan {
        if err != nil {
            return err
        }
    }
}
```

---

## 安全实践

### 1. 敏感信息保护

**不记录敏感信息**:

```go
// ❌ 错误：记录敏感信息
log.Printf("connecting to %s with token %s", url, apiKey)

// ✅ 正确：过滤敏感信息
log.Printf("connecting to %s", sanitizeURL(url))
```

### 2. 输入验证

**验证所有用户输入**:

```go
// ✅ 正确：验证输入参数
func (s *Service) ProcessInput(ctx context.Context, input string) error {
    // 验证非空
    if strings.TrimSpace(input) == "" {
        return errors.New("input cannot be empty")
    }

    // 验证长度
    if len(input) > maxInputLength {
        return fmt.Errorf("input too long: max %d chars", maxInputLength)
    }

    // 验证格式
    if !isValidFormat(input) {
        return errors.New("invalid input format")
    }

    return s.processValidatedInput(ctx, input)
}
```

### 3. 最小权限

**沙箱代码执行使用最小权限**:

```yaml
# ✅ 正确：限制代码执行权限
code_execution:
  allowed_operations:
    - "read"
    - "write"
    # 不允许网络、进程等危险操作

  resource_limits:
    memory: "512m"
    cpu: "1.0"
    timeout: 30
```

### 4. 路径遍历

**防止路径遍历攻击**:

```go
// ✅ 正确：验证路径
func (s *FileService) SafeOpen(path string) (*os.File, error) {
    // 清理路径
    cleanPath := filepath.Clean(path)

    // 验证在允许目录下
    if !isAllowedPath(cleanPath) {
        return nil, errors.New("path not allowed")
    }

    // 验证不存在路径遍历
    if strings.Contains(cleanPath, "..") {
        return nil, errors.New("path traversal detected")
    }

    return os.Open(cleanPath)
}
```

### 5. API 密钥管理

**使用环境变量存储密钥**:

```go
// ❌ 错误：硬编码密钥
apiKey := "sk-1234567890abcdef"

// ✅ 正确：从环境变量读取
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    return errors.New("OPENAI_API_KEY not set")
}
```

---

## 测试策略

### 1. 单元测试

**测试单个函数/方法**:

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("expected 5, got %d", result)
    }
}
```

### 2. 表驱动测试

**使用表格驱动多个测试用例**:

```go
func TestValidateInput(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "valid", false},
        {"empty input", "", true},
        {"too long input", strings.Repeat("a", 1001), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateInput(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 3. 集成测试

**测试组件协作**:

```go
func TestAgentWorkflow(t *testing.T) {
    // 设置测试环境
    ctx := context.Background()
    cfg := testConfig()
    host, err := NewHost(ctx, cfg, nil, nil)
    if err != nil {
        t.Fatal(err)
    }

    // 测试工作流
    workflow, err := host.GetWorkflow("test")
    if err != nil {
        t.Fatal(err)
    }

    result, err := workflow.Run(ctx, "test input")
    if err != nil {
        t.Errorf("workflow.Run() error = %v", err)
    }

    if result.Content != "expected output" {
        t.Errorf("expected 'expected output', got '%s'", result.Content)
    }
}
```

### 4. 基准测试

**测试性能指标**:

```go
func BenchmarkAgentRun(b *testing.B) {
    ctx := context.Background()
    agent := newTestAgent()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := agent.Run(ctx, "test input")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## 监控调试

### 1. 结构化日志

**使用结构化日志便于分析**:

```go
import "log/slog"

func (s *Service) DoSomething(ctx context.Context) error {
    logger := slog.FromContext(ctx)

    logger.Info("operation started",
        "operation", "do_something",
        "trace_id", generateTraceID(),
    )

    result, err := s.doSomethingInternal(ctx)
    if err != nil {
        logger.Error("operation failed",
            "operation", "do_something",
            "error", err,
            "trace_id", ctx.Value("trace_id"),
        )
        return err
    }

    logger.Info("operation completed",
        "operation", "do_something",
        "trace_id", ctx.Value("trace_id"),
    )
    return nil
}
```

### 2. 性能指标

**记录关键性能指标**:

```go
func (s *Service) DoSomething(ctx context.Context) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)

        // 记录耗时
        s.metrics.RecordDuration("operation_duration", duration)

        // 记录慢查询
        if duration > s.slowThreshold {
            slog.Warn("slow operation",
                "operation", "do_something",
                "duration", duration,
            )
        }
    }()

    return s.doSomethingInternal(ctx)
}
```

### 3. 分布式追踪

**使用 OpenTelemetry 追踪请求**:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (s *Service) DoSomething(ctx context.Context) error {
    tracer := otel.Tracer("service")

    ctx, span := tracer.Start(ctx, "DoSomething")
    defer span.End()

    // 添加属性
    span.SetAttributes(
        attribute.String("service.name", "my-service"),
        attribute.String("service.version", "1.0.0"),
    )

    return s.doSomethingInternal(ctx)
}
```

---

## 部署运维

### 1. 健康检查

**实现健康检查接口**:

```go
func (s *Service) HealthCheck(ctx context.Context) error {
    // 检查数据库连接
    if err := s.db.Ping(ctx); err != nil {
        return fmt.Errorf("database unhealthy: %w", err)
    }

    // 检查外部服务
    if err := s.externalService.Ping(ctx); err != nil {
        return fmt.Errorf("external service unhealthy: %w", err)
    }

    return nil
}
```

### 2. 优雅关闭

**实现优雅关闭机制**:

```go
func (s *Service) Shutdown(ctx context.Context) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.shuttingDown {
        return errors.New("already shutting down")
    }

    s.shuttingDown = true

    // 停止接受新请求
    s.server.Shutdown(ctx)

    // 等待现有请求完成
    s.wg.Wait()

    // 关闭资源
    return s.closeResources()
}
```

### 3. 配置热更新

**支持配置热更新**:

```go
func (h *Host) WatchConfig(ctx context.Context) error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    defer watcher.Close()

    // 监控配置文件
    watcher.Add(h.configPath)
    for {
        select {
        case <-ctx.Done():
            return nil
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                // 重新加载配置
                if err := h.reloadConfig(); err != nil {
                    log.Printf("reload config failed: %v", err)
                }
            }
        }
    }
}
```

---

## 相关文档

- 📘 [快速开始](../quickstart/QUICKSTART.md) - 5 分钟上手
- 📘 [架构概览](../architecture/ARCHITECTURE_OVERVIEW.md) - 系统架构
- 📘 [配置指南](../configuration/CONFIGURATION.md) - 详细配置
- 📘 [故障排查](../operation/TROUBLESHOOTING.md) - 问题排查

---

**Made with ❤️ by AgentFramework Team**
