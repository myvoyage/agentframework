# 代码执行模块优化路线图

## 概述

本文档提供了代码执行模块的详细优化路线图，包括短期、中期和长期优化计划。

---

## 第一阶段：短期优化 (1-2 周)

### 1.1 Yaegi 编译缓存实现

**目标**: 减少重复代码的执行时间 50-80%

**实施步骤**:

#### 步骤 1: 设计缓存结构
```go
// cache.go
package code_exec

import (
    "container/list"
    "crypto/sha256"
    "encoding/hex"
    "sync"
)

// CacheEntry 缓存条目
type CacheEntry struct {
    Code       string
    Result     *ExecutionResult
    AccessTime time.Time
    HitCount   int64
}

// LRUCache LRU 缓存
type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    list     *list.List
    mu       sync.RWMutex
    stats    CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
    Hits       int64
    Misses     int64
    Evictions  int64
    TotalSize  int64
}

// NewLRUCache 创建 LRU 缓存
func NewLRUCache(capacity int) *LRUCache {
    return &LRUCache{
        capacity: capacity,
        cache:    make(map[string]*list.Element),
        list:     list.New(),
    }
}

// Get 获取缓存
func (c *LRUCache) Get(key string) (*ExecutionResult, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if elem, ok := c.cache[key]; ok {
        c.list.MoveToFront(elem)
        entry := elem.Value.(*CacheEntry)
        entry.AccessTime = time.Now()
        entry.HitCount++
        c.stats.Hits++
        return entry.Result, true
    }
    
    c.stats.Misses++
    return nil, false
}

// Put 添加缓存
func (c *LRUCache) Put(key string, result *ExecutionResult) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if elem, ok := c.cache[key]; ok {
        c.list.MoveToFront(elem)
        entry := elem.Value.(*CacheEntry)
        entry.Result = result
        entry.AccessTime = time.Now()
        return
    }
    
    if c.list.Len() >= c.capacity {
        // 淘汰最久未使用的条目
        oldest := c.list.Back()
        if oldest != nil {
            c.list.Remove(oldest)
            entry := oldest.Value.(*CacheEntry)
            delete(c.cache, c.hashCode(entry.Code))
            c.stats.Evictions++
        }
    }
    
    entry := &CacheEntry{
        Result:     result,
        AccessTime: time.Now(),
        HitCount:   0,
    }
    
    elem := c.list.PushFront(entry)
    c.cache[key] = elem
    c.stats.TotalSize++
}

// hashCode 计算代码哈希
func (c *LRUCache) hashCode(code string) string {
    hash := sha256.Sum256([]byte(code))
    return hex.EncodeToString(hash[:])
}

// GetStats 获取缓存统计
func (c *LRUCache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.stats
}
```


#### 步骤 2: 集成到 YaegiInterpreter
```go
// yaegi_interpreter.go (修改)

type YaegiInterpreter struct {
    interp      *interp.Interpreter
    mu          sync.RWMutex
    initialized bool
    config      YaegiConfig
    cache       *LRUCache  // 新增缓存
}

func NewYaegiInterpreter(config YaegiConfig) (*YaegiInterpreter, error) {
    yi := &YaegiInterpreter{
        config: config,
        cache:  NewLRUCache(100),  // 缓存 100 个条目
    }
    
    if err := yi.initialize(); err != nil {
        return nil, fmt.Errorf("failed to initialize yaegi interpreter: %w", err)
    }
    
    return yi, nil
}

func (yi *YaegiInterpreter) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
    // 检查缓存
    if yi.config.EnableCache {
        cacheKey := yi.cache.hashCode(code)
        if cached, ok := yi.cache.Get(cacheKey); ok {
            // 返回缓存结果（复制一份避免修改）
            result := *cached
            result.Duration = 0  // 缓存命中，执行时间为 0
            return &result, nil
        }
    }
    
    // 执行代码
    result, err := yi.executeCode(ctx, code, input)
    if err != nil {
        return result, err
    }
    
    // 缓存成功的结果
    if yi.config.EnableCache && result.Success {
        cacheKey := yi.cache.hashCode(code)
        yi.cache.Put(cacheKey, result)
    }
    
    return result, nil
}

// GetCacheStats 获取缓存统计
func (yi *YaegiInterpreter) GetCacheStats() CacheStats {
    if yi.cache == nil {
        return CacheStats{}
    }
    return yi.cache.GetStats()
}
```

#### 步骤 3: 添加测试
```go
// cache_test.go
func TestLRUCache(t *testing.T) {
    cache := NewLRUCache(3)
    
    // 测试基本功能
    result1 := &ExecutionResult{Success: true, Output: "test1"}
    cache.Put("key1", result1)
    
    got, ok := cache.Get("key1")
    if !ok || got.Output != "test1" {
        t.Error("Cache get failed")
    }
    
    // 测试 LRU 淘汰
    cache.Put("key2", &ExecutionResult{Success: true, Output: "test2"})
    cache.Put("key3", &ExecutionResult{Success: true, Output: "test3"})
    cache.Put("key4", &ExecutionResult{Success: true, Output: "test4"})
    
    // key1 应该被淘汰
    _, ok = cache.Get("key1")
    if ok {
        t.Error("key1 should be evicted")
    }
    
    // 检查统计
    stats := cache.GetStats()
    if stats.Evictions != 1 {
        t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
    }
}

func BenchmarkCacheHit(b *testing.B) {
    cache := NewLRUCache(100)
    result := &ExecutionResult{Success: true, Output: "test"}
    cache.Put("key", result)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get("key")
    }
}
```

**预期收益**:
- 缓存命中率: 60-80%
- 性能提升: 50-80% (重复代码)
- 内存开销: ~10-20MB (100 条目)


---

### 1.2 容器池实现

**目标**: 减少容器启动开销 80%

**实施步骤**:

#### 步骤 1: 设计容器池结构
```go
// container_pool.go
package code_exec

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// ContainerPool 容器池
type ContainerPool struct {
    pools       map[string]*LanguagePool  // 每种语言一个池
    config      PoolConfig
    executor    *ContainerExecutor
    mu          sync.RWMutex
    stopChan    chan struct{}
    wg          sync.WaitGroup
}

// PoolConfig 容器池配置
type PoolConfig struct {
    MinSize         int           // 最小容器数
    MaxSize         int           // 最大容器数
    IdleTimeout     time.Duration // 空闲超时
    HealthCheckInterval time.Duration // 健康检查间隔
}

// LanguagePool 语言容器池
type LanguagePool struct {
    language    string
    containers  chan *PooledContainer
    active      map[string]*PooledContainer
    mu          sync.RWMutex
    stats       PoolStats
}

// PooledContainer 池化容器
type PooledContainer struct {
    ID          string
    Language    string
    CreatedAt   time.Time
    LastUsedAt  time.Time
    UseCount    int64
    Healthy     bool
}

// PoolStats 池统计
type PoolStats struct {
    TotalCreated  int64
    TotalDestroyed int64
    CurrentSize   int
    ActiveCount   int
    IdleCount     int
    ReuseCount    int64
}

// NewContainerPool 创建容器池
func NewContainerPool(executor *ContainerExecutor, config PoolConfig) *ContainerPool {
    pool := &ContainerPool{
        pools:    make(map[string]*LanguagePool),
        config:   config,
        executor: executor,
        stopChan: make(chan struct{}),
    }
    
    // 为每种语言创建池
    for _, lang := range []string{"python", "javascript", "go", "bash"} {
        pool.pools[lang] = &LanguagePool{
            language:   lang,
            containers: make(chan *PooledContainer, config.MaxSize),
            active:     make(map[string]*PooledContainer),
        }
    }
    
    // 启动维护协程
    pool.wg.Add(1)
    go pool.maintain()
    
    return pool
}

// Acquire 获取容器
func (p *ContainerPool) Acquire(ctx context.Context, language string) (*PooledContainer, error) {
    langPool, ok := p.pools[language]
    if !ok {
        return nil, fmt.Errorf("unsupported language: %s", language)
    }
    
    // 尝试从池中获取
    select {
    case container := <-langPool.containers:
        // 检查容器健康状态
        if container.Healthy {
            container.LastUsedAt = time.Now()
            container.UseCount++
            
            langPool.mu.Lock()
            langPool.active[container.ID] = container
            langPool.stats.ReuseCount++
            langPool.mu.Unlock()
            
            return container, nil
        }
        // 容器不健康，销毁并创建新的
        p.destroyContainer(container)
    default:
        // 池为空，创建新容器
    }
    
    // 创建新容器
    return p.createContainer(ctx, language)
}

// Release 释放容器
func (p *ContainerPool) Release(container *PooledContainer) error {
    langPool, ok := p.pools[container.Language]
    if !ok {
        return fmt.Errorf("unknown language: %s", container.Language)
    }
    
    langPool.mu.Lock()
    delete(langPool.active, container.ID)
    langPool.mu.Unlock()
    
    // 检查容器是否健康
    if !container.Healthy {
        return p.destroyContainer(container)
    }
    
    // 检查是否超过最大空闲时间
    if time.Since(container.LastUsedAt) > p.config.IdleTimeout {
        return p.destroyContainer(container)
    }
    
    // 放回池中
    select {
    case langPool.containers <- container:
        return nil
    default:
        // 池已满，销毁容器
        return p.destroyContainer(container)
    }
}

// createContainer 创建容器
func (p *ContainerPool) createContainer(ctx context.Context, language string) (*PooledContainer, error) {
    // 使用 ContainerExecutor 创建容器
    containerID, err := p.executor.createContainer(ctx, p.executor.getImage(language), language, "")
    if err != nil {
        return nil, err
    }
    
    container := &PooledContainer{
        ID:         containerID,
        Language:   language,
        CreatedAt:  time.Now(),
        LastUsedAt: time.Now(),
        UseCount:   0,
        Healthy:    true,
    }
    
    langPool := p.pools[language]
    langPool.mu.Lock()
    langPool.stats.TotalCreated++
    langPool.active[containerID] = container
    langPool.mu.Unlock()
    
    return container, nil
}

// destroyContainer 销毁容器
func (p *ContainerPool) destroyContainer(container *PooledContainer) error {
    err := p.executor.removeContainer(context.Background(), container.ID)
    
    langPool := p.pools[container.Language]
    langPool.mu.Lock()
    langPool.stats.TotalDestroyed++
    delete(langPool.active, container.ID)
    langPool.mu.Unlock()
    
    return err
}

// maintain 维护协程
func (p *ContainerPool) maintain() {
    defer p.wg.Done()
    
    ticker := time.NewTicker(p.config.HealthCheckInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            p.healthCheck()
            p.warmUp()
        case <-p.stopChan:
            return
        }
    }
}

// healthCheck 健康检查
func (p *ContainerPool) healthCheck() {
    for _, langPool := range p.pools {
        langPool.mu.RLock()
        containers := make([]*PooledContainer, 0, len(langPool.active))
        for _, container := range langPool.active {
            containers = append(containers, container)
        }
        langPool.mu.RUnlock()
        
        for _, container := range containers {
            // 检查容器健康状态
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            _, err := p.executor.client.ContainerInspect(ctx, container.ID)
            cancel()
            
            if err != nil {
                container.Healthy = false
            }
        }
    }
}

// warmUp 预热容器
func (p *ContainerPool) warmUp() {
    for lang, langPool := range p.pools {
        langPool.mu.RLock()
        currentSize := len(langPool.containers) + len(langPool.active)
        langPool.mu.RUnlock()
        
        // 如果容器数量少于最小值，创建新容器
        if currentSize < p.config.MinSize {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            container, err := p.createContainer(ctx, lang)
            cancel()
            
            if err == nil {
                langPool.containers <- container
            }
        }
    }
}

// Close 关闭容器池
func (p *ContainerPool) Close() error {
    close(p.stopChan)
    p.wg.Wait()
    
    // 销毁所有容器
    for _, langPool := range p.pools {
        close(langPool.containers)
        for container := range langPool.containers {
            p.destroyContainer(container)
        }
        
        langPool.mu.Lock()
        for _, container := range langPool.active {
            p.destroyContainer(container)
        }
        langPool.mu.Unlock()
    }
    
    return nil
}

// GetStats 获取统计信息
func (p *ContainerPool) GetStats() map[string]PoolStats {
    stats := make(map[string]PoolStats)
    
    for lang, langPool := range p.pools {
        langPool.mu.RLock()
        langPool.stats.CurrentSize = len(langPool.containers) + len(langPool.active)
        langPool.stats.ActiveCount = len(langPool.active)
        langPool.stats.IdleCount = len(langPool.containers)
        stats[lang] = langPool.stats
        langPool.mu.RUnlock()
    }
    
    return stats
}
```


#### 步骤 2: 集成到 ContainerExecutor
```go
// container_executor.go (修改)

type ContainerExecutor struct {
    client      *client.Client
    config      ContainerConfig
    mu          sync.RWMutex
    containers  map[string]*ContainerInfo
    imageCache  map[string]bool
    initialized bool
    pool        *ContainerPool  // 新增容器池
}

func NewContainerExecutor(config ContainerConfig) (*ContainerExecutor, error) {
    // ... 现有代码 ...
    
    ce := &ContainerExecutor{
        client:      cli,
        config:      config,
        containers:  make(map[string]*ContainerInfo),
        imageCache:  make(map[string]bool),
        initialized: true,
    }
    
    // 创建容器池
    if config.EnablePool {
        poolConfig := PoolConfig{
            MinSize:             2,
            MaxSize:             10,
            IdleTimeout:         5 * time.Minute,
            HealthCheckInterval: 30 * time.Second,
        }
        ce.pool = NewContainerPool(ce, poolConfig)
    }
    
    return ce, nil
}

func (ce *ContainerExecutor) Execute(ctx context.Context, language, code string) (*ExecutionResult, error) {
    if !ce.IsEnabled() {
        return nil, fmt.Errorf("container executor not enabled")
    }
    
    startTime := time.Now()
    
    // 如果启用了容器池，使用池化容器
    if ce.pool != nil {
        return ce.executeWithPool(ctx, language, code, startTime)
    }
    
    // 否则使用原有的一次性容器逻辑
    return ce.executeWithoutPool(ctx, language, code, startTime)
}

func (ce *ContainerExecutor) executeWithPool(ctx context.Context, language, code string, startTime time.Time) (*ExecutionResult, error) {
    // 从池中获取容器
    container, err := ce.pool.Acquire(ctx, language)
    if err != nil {
        return &ExecutionResult{
            Success:  false,
            Error:    fmt.Sprintf("failed to acquire container: %v", err),
            Duration: time.Since(startTime),
        }, nil
    }
    defer ce.pool.Release(container)
    
    // 创建临时文件
    tmpFile, err := ce.createTempFile(language, code)
    if err != nil {
        return &ExecutionResult{
            Success:  false,
            Error:    fmt.Sprintf("failed to create temp file: %v", err),
            Duration: time.Since(startTime),
        }, nil
    }
    defer os.Remove(tmpFile)
    
    // 在容器中执行代码
    result, err := ce.executeInContainer(ctx, container.ID, tmpFile)
    if err != nil {
        container.Healthy = false  // 标记容器不健康
        return &ExecutionResult{
            Success:  false,
            Error:    fmt.Sprintf("execution failed: %v", err),
            Duration: time.Since(startTime),
        }, nil
    }
    
    result.Duration = time.Since(startTime)
    return result, nil
}
```

#### 步骤 3: 添加测试
```go
// container_pool_test.go
func TestContainerPool(t *testing.T) {
    config := ContainerConfig{
        Enabled: true,
        EnablePool: true,
    }
    
    executor, err := NewContainerExecutor(config)
    if err != nil {
        t.Skip("Docker not available")
    }
    defer executor.Close()
    
    // 测试容器获取和释放
    ctx := context.Background()
    container1, err := executor.pool.Acquire(ctx, "python")
    if err != nil {
        t.Fatalf("Failed to acquire container: %v", err)
    }
    
    // 释放容器
    err = executor.pool.Release(container1)
    if err != nil {
        t.Errorf("Failed to release container: %v", err)
    }
    
    // 再次获取，应该复用同一个容器
    container2, err := executor.pool.Acquire(ctx, "python")
    if err != nil {
        t.Fatalf("Failed to acquire container: %v", err)
    }
    
    if container1.ID != container2.ID {
        t.Error("Container should be reused")
    }
    
    // 检查统计
    stats := executor.pool.GetStats()
    pythonStats := stats["python"]
    if pythonStats.ReuseCount != 1 {
        t.Errorf("Expected 1 reuse, got %d", pythonStats.ReuseCount)
    }
}

func BenchmarkContainerPool(b *testing.B) {
    config := ContainerConfig{
        Enabled: true,
        EnablePool: true,
    }
    
    executor, err := NewContainerExecutor(config)
    if err != nil {
        b.Skip("Docker not available")
    }
    defer executor.Close()
    
    ctx := context.Background()
    code := "print('Hello, World!')"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := executor.Execute(ctx, "python", code)
        if err != nil {
            b.Errorf("Execution failed: %v", err)
        }
    }
}
```

**预期收益**:
- 容器启动开销: 减少 80%
- 平均执行时间: 从 1.5s 降至 0.3s
- 容器复用率: 70-90%
- 内存开销: ~50-100MB (池化容器)


---

### 1.3 代码分析器性能优化

**目标**: 提升代码分析性能 20-30%

**实施步骤**:

#### 步骤 1: 性能分析
```bash
# 运行性能分析
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
go tool pprof mem.prof
```

#### 步骤 2: 优化正则表达式
```go
// code_analyzer.go (优化)

// 使用更高效的字符串匹配
func (a *CodeAnalyzer) Analyze(language, code string) *AnalysisResult {
    // 预先分割行，避免重复分割
    lines := strings.Split(code, "\n")
    lineCount := len(lines)
    
    result := &AnalysisResult{
        Safe:          true,
        Issues:        make([]SecurityIssue, 0, 10),
        Language:      language,
        LineCount:     lineCount,
        CharCount:     len(code),
        NetworkOps:    make([]NetworkOperation, 0, 5),
        FileSystemOps: make([]FileSystemOperation, 0, 5),
        ProcessOps:    make([]ProcessOperation, 0, 5),
        CryptoIssues:  make([]CryptoIssue, 0, 5),
        DatabaseOps:   make([]DatabaseOperation, 0, 5),
        QualityIssues: make([]QualityIssue, 0, 10),
    }
    
    // 并行检测（对于大文件）
    if lineCount > 100 {
        return a.analyzeParallel(language, code, lines)
    }
    
    // 串行检测（对于小文件）
    return a.analyzeSerial(language, code, lines)
}

// analyzeParallel 并行分析
func (a *CodeAnalyzer) analyzeParallel(language, code string, lines []string) *AnalysisResult {
    var wg sync.WaitGroup
    resultChan := make(chan *AnalysisResult, 6)
    
    // 并行执行各种检测
    wg.Add(6)
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.detectNetworkOps(language, lines, result)
        resultChan <- result
    }()
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.detectFileSystemOps(language, lines, result)
        resultChan <- result
    }()
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.detectProcessOps(language, lines, result)
        resultChan <- result
    }()
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.detectCryptoIssues(language, lines, result)
        resultChan <- result
    }()
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.detectDatabaseOps(language, lines, result)
        resultChan <- result
    }()
    
    go func() {
        defer wg.Done()
        result := &AnalysisResult{}
        a.checkQualityIssues(language, lines, result)
        resultChan <- result
    }()
    
    // 等待所有检测完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // 合并结果
    finalResult := &AnalysisResult{
        Safe:          true,
        Language:      language,
        LineCount:     len(lines),
        CharCount:     len(strings.Join(lines, "\n")),
        Issues:        make([]SecurityIssue, 0),
        NetworkOps:    make([]NetworkOperation, 0),
        FileSystemOps: make([]FileSystemOperation, 0),
        ProcessOps:    make([]ProcessOperation, 0),
        CryptoIssues:  make([]CryptoIssue, 0),
        DatabaseOps:   make([]DatabaseOperation, 0),
        QualityIssues: make([]QualityIssue, 0),
    }
    
    for result := range resultChan {
        finalResult.Issues = append(finalResult.Issues, result.Issues...)
        finalResult.NetworkOps = append(finalResult.NetworkOps, result.NetworkOps...)
        finalResult.FileSystemOps = append(finalResult.FileSystemOps, result.FileSystemOps...)
        finalResult.ProcessOps = append(finalResult.ProcessOps, result.ProcessOps...)
        finalResult.CryptoIssues = append(finalResult.CryptoIssues, result.CryptoIssues...)
        finalResult.DatabaseOps = append(finalResult.DatabaseOps, result.DatabaseOps...)
        finalResult.QualityIssues = append(finalResult.QualityIssues, result.QualityIssues...)
        
        if !result.Safe {
            finalResult.Safe = false
        }
    }
    
    // 计算复杂度和评分
    finalResult.Complexity = a.calculateComplexity(strings.Join(lines, "\n"))
    finalResult.Score = a.calculateQualityScore(finalResult)
    finalResult.Suggestions = a.generateSuggestions(finalResult)
    
    return finalResult
}
```

#### 步骤 3: 添加性能基准测试
```go
// analyzer_benchmark_test.go
func BenchmarkAnalyzeSmallFile(b *testing.B) {
    analyzer := NewCodeAnalyzer()
    code := `
import requests
response = requests.get('http://example.com')
print(response.text)
`
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        analyzer.Analyze("python", code)
    }
}

func BenchmarkAnalyzeLargeFile(b *testing.B) {
    analyzer := NewCodeAnalyzer()
    
    // 生成大文件（1000 行）
    var sb strings.Builder
    for i := 0; i < 1000; i++ {
        sb.WriteString(fmt.Sprintf("x%d = %d\n", i, i))
    }
    code := sb.String()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        analyzer.Analyze("python", code)
    }
}

func BenchmarkParallelAnalysis(b *testing.B) {
    analyzer := NewCodeAnalyzer()
    code := generateLargeCode(1000)
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            analyzer.Analyze("python", code)
        }
    })
}
```

**预期收益**:
- 小文件 (<100 行): 10-20% 性能提升
- 大文件 (>100 行): 30-50% 性能提升
- 并发分析: 2-3x 吞吐量提升

---

## 第二阶段：中期优化 (1-2 月)

### 2.1 智能执行模式选择

**目标**: 根据代码特征自动选择最佳执行模式

**实施方案**:
```go
// execution_mode_selector.go
package code_exec

type ExecutionModeSelector struct {
    history map[string]*ExecutionHistory
    mu      sync.RWMutex
}

type ExecutionHistory struct {
    CodeHash     string
    YaegiTime    time.Duration
    GoRunTime    time.Duration
    ContainerTime time.Duration
    BestMode     ExecutionMode
    SampleCount  int
}

func (s *ExecutionModeSelector) SelectMode(code string, language string) ExecutionMode {
    // 基于代码特征选择模式
    features := s.extractFeatures(code, language)
    
    // 检查历史数据
    if history, ok := s.getHistory(code); ok {
        return history.BestMode
    }
    
    // 基于特征预测
    if features.HasComplexImports {
        return ModeGoRun
    }
    
    if features.IsSimple {
        return ModeYaegi
    }
    
    if features.RequiresIsolation {
        return ModeContainer
    }
    
    return ModeAuto
}

func (s *ExecutionModeSelector) extractFeatures(code string, language string) CodeFeatures {
    return CodeFeatures{
        LineCount:          len(strings.Split(code, "\n")),
        HasComplexImports:  strings.Contains(code, "import"),
        IsSimple:           len(code) < 1000,
        RequiresIsolation:  strings.Contains(code, "network") || strings.Contains(code, "file"),
    }
}
```


### 2.2 实时资源监控

**目标**: 提供实时资源使用监控和告警

**实施方案**:
```go
// monitoring.go
package code_exec

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    executionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "code_exec_duration_seconds",
            Help: "Code execution duration in seconds",
        },
        []string{"language", "mode"},
    )
    
    executionTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "code_exec_total",
            Help: "Total number of code executions",
        },
        []string{"language", "mode", "status"},
    )
    
    activeExecutions = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "code_exec_active",
            Help: "Number of active code executions",
        },
        []string{"language"},
    )
    
    memoryUsage = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "code_exec_memory_bytes",
            Help: "Memory usage in bytes",
        },
        []string{"language"},
    )
)

// 在执行代码时记录指标
func (m *CodeExecutorModule) runCodeWithMetrics(language, code, input string, timeout int) (map[string]any, error) {
    startTime := time.Now()
    activeExecutions.WithLabelValues(language).Inc()
    defer activeExecutions.WithLabelValues(language).Dec()
    
    result, err := m.runCode(language, code, input, timeout)
    
    duration := time.Since(startTime).Seconds()
    executionDuration.WithLabelValues(language, m.config.ExecutionMode).Observe(duration)
    
    status := "success"
    if err != nil || !result["success"].(bool) {
        status = "failure"
    }
    executionTotal.WithLabelValues(language, m.config.ExecutionMode, status).Inc()
    
    return result, err
}
```

### 2.3 增强错误诊断

**目标**: 提供更详细的错误信息和修复建议

**实施方案**:
```go
// error_diagnostics.go
package code_exec

type ErrorDiagnostics struct {
    ErrorType    string
    ErrorMessage string
    Location     ErrorLocation
    Suggestion   string
    RelatedDocs  []string
}

type ErrorLocation struct {
    Line   int
    Column int
    Code   string
}

func (yi *YaegiInterpreter) enhanceError(err error, code string) *ErrorDiagnostics {
    // 解析错误信息
    errMsg := err.Error()
    
    // 提取位置信息
    location := extractLocation(errMsg, code)
    
    // 生成修复建议
    suggestion := generateSuggestion(errMsg)
    
    // 查找相关文档
    docs := findRelatedDocs(errMsg)
    
    return &ErrorDiagnostics{
        ErrorType:    classifyError(errMsg),
        ErrorMessage: errMsg,
        Location:     location,
        Suggestion:   suggestion,
        RelatedDocs:  docs,
    }
}
```

---

## 第三阶段：长期优化 (3-6 月)

### 3.1 分布式执行支持

**目标**: 支持 Kubernetes 集群执行，提升并发能力 10x+

**架构设计**:
```
                    Load Balancer
                          |
        +----------------+----------------+
        |                |                |
   Executor 1       Executor 2       Executor 3
        |                |                |
   +----+----+      +----+----+      +----+----+
   |         |      |         |      |         |
 Pod 1    Pod 2   Pod 3    Pod 4   Pod 5    Pod 6
```

**实施步骤**:
1. 设计分布式任务调度器
2. 实现 Kubernetes Job 执行器
3. 添加负载均衡和故障转移
4. 实现跨节点资源管理

### 3.2 AI 辅助代码分析

**目标**: 使用 AI 模型提升代码分析准确性

**实施方案**:
1. 集成代码语义分析模型
2. 实现智能安全检测
3. 提供 AI 驱动的修复建议
4. 持续学习和优化

### 3.3 多版本语言支持

**目标**: 支持多版本 Python/Node/Go

**实施方案**:
```go
type LanguageVersion struct {
    Language string
    Version  string
    Image    string
}

var supportedVersions = []LanguageVersion{
    {Language: "python", Version: "3.9", Image: "python:3.9-alpine"},
    {Language: "python", Version: "3.10", Image: "python:3.10-alpine"},
    {Language: "python", Version: "3.11", Image: "python:3.11-alpine"},
    {Language: "node", Version: "16", Image: "node:16-alpine"},
    {Language: "node", Version: "18", Image: "node:18-alpine"},
    {Language: "node", Version: "20", Image: "node:20-alpine"},
}
```

---

## 实施时间表

### 第 1-2 周
- ✅ Yaegi 编译缓存实现
- ✅ 容器池实现
- ✅ 代码分析器性能优化

### 第 3-4 周
- 智能执行模式选择
- 实时资源监控
- 增强错误诊断

### 第 2-3 月
- 代码分析规则优化
- 性能持续优化
- 文档和示例完善

### 第 4-6 月
- 分布式执行支持研究
- AI 辅助代码分析探索
- 多版本语言支持

---

## 成功指标

### 性能指标
- Yaegi 缓存命中率: ≥ 60%
- 容器复用率: ≥ 70%
- 平均执行时间: 减少 50%
- 并发能力: 提升 3-5x

### 质量指标
- 代码覆盖率: 保持 ≥ 80%
- 误报率: 减少 30%
- 错误诊断准确率: ≥ 90%

### 运维指标
- 系统可用性: ≥ 99.9%
- 平均故障恢复时间: < 5 分钟
- 资源利用率: 70-80%

---

## 风险管理

### 技术风险
- **缓存一致性**: 确保缓存失效机制正确
- **容器池稳定性**: 完善健康检查和故障恢复
- **并发安全**: 严格的并发测试

### 运维风险
- **资源消耗**: 监控资源使用，及时调整
- **性能回归**: 持续性能测试，防止回归
- **兼容性**: 确保向后兼容

---

## 总结

本优化路线图提供了清晰的优化方向和实施计划：

**短期优化** (1-2 周):
- 重点: 性能提升
- 目标: 50-80% 性能提升
- 风险: 低

**中期优化** (1-2 月):
- 重点: 功能增强
- 目标: 提升用户体验
- 风险: 中

**长期优化** (3-6 月):
- 重点: 架构升级
- 目标: 10x+ 能力提升
- 风险: 高

建议按照优先级逐步实施，每个阶段完成后进行充分测试和验证，确保系统稳定性和性能提升。

