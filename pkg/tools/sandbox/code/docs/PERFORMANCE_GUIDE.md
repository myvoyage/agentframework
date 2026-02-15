# 性能优化指南

## 概述

本文档提供代码执行模块的性能优化建议和最佳实践。

---

## 性能指标

### 关键指标

- **执行时间**: 代码执行的总时间
- **启动时间**: 环境初始化时间
- **内存使用**: 峰值内存占用
- **吞吐量**: 每秒处理的请求数
- **缓存命中率**: 缓存有效性

### 性能目标

| 场景 | 目标 |
|------|------|
| 简单代码执行 | < 100ms |
| 复杂代码执行 | < 1s |
| 容器启动 | < 2s |
| 缓存命中 | < 10ms |
| 吞吐量 | > 100 req/s |

---

## 优化策略

### 1. 使用 Yaegi 加速 Go 代码

**性能提升**: 428x

```yaml
executor:
  execution_mode: local  # 使用 Yaegi
```

**对比**:
- go run: ~2000ms
- Yaegi: ~4.7ms

**适用场景**:
- 简单 Go 代码
- 不使用 CGO
- 标准库为主

---

### 2. 启用编译缓存

**性能提升**: 12,600x（缓存命中时）

```yaml
yaegi:
  enable_cache: true
  cache_capacity: 500
```

**效果**:
- 首次执行: 正常速度
- 缓存命中: 几乎瞬时

**建议**:
- 开发环境: cache_capacity = 100
- 生产环境: cache_capacity = 500-1000

---

### 3. 启用容器池

**性能提升**: 80% 启动时间减少

```yaml
container:
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

**效果**:
- 无池: ~2000ms
- 有池: ~400ms

**配置建议**:
```yaml
# 低负载
pool_min_size: 2
pool_max_size: 10

# 中负载
pool_min_size: 5
pool_max_size: 20

# 高负载
pool_min_size: 10
pool_max_size: 50
```

---

### 4. 选择合适的执行模式

| 模式 | 速度 | 安全性 | 适用场景 |
|------|------|--------|----------|
| local | 最快 | 一般 | 开发环境 |
| container | 较慢 | 最高 | 生产环境 |
| auto | 自动 | 平衡 | 推荐 |

```yaml
# 开发环境
executor:
  execution_mode: local

# 生产环境
executor:
  execution_mode: container

# 推荐
executor:
  execution_mode: auto
```

---

### 5. 优化资源限制

**平衡性能和安全**:

```yaml
# 高性能配置
executor:
  timeout: 60000
  memory_limit: 1024
  cpu_limit: 4

container:
  cpu_limit: "2.0"
  memory_limit: "1g"

# 平衡配置（推荐）
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 2

container:
  cpu_limit: "0.5"
  memory_limit: "512m"

# 高安全配置
executor:
  timeout: 10000
  memory_limit: 256
  cpu_limit: 1

container:
  cpu_limit: "0.25"
  memory_limit: "256m"
```

---

### 6. 使用轻量级镜像

```yaml
container:
  default_images:
    python: python:3.11-alpine      # 50 MB
    javascript: node:18-alpine      # 180 MB
    go: golang:1.21-alpine          # 300 MB
    bash: alpine:latest             # 7 MB
```

**对比**:
- 标准镜像: python:3.11 (900 MB)
- Alpine 镜像: python:3.11-alpine (50 MB)
- 启动速度提升: 3-5x

---

### 7. 预加载常用包

```yaml
yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
    - strings
    - time
    - math
    - encoding/json
```

**效果**: 减少首次执行时间

---

### 8. 并发执行优化

```go
// 使用 goroutine 并发执行
var wg sync.WaitGroup
results := make(chan *ExecutionResult, len(codes))

for _, code := range codes {
    wg.Add(1)
    go func(c string) {
        defer wg.Done()
        result, _ := module.ExecuteCode(ctx, lang, c)
        results <- result
    }(code)
}

wg.Wait()
close(results)
```

**注意**: 控制并发数，避免资源耗尽

---

## 性能监控

### 1. 内置监控

```go
// 获取统计信息
stats := module.GetStats()
fmt.Printf("总执行次数: %d\n", stats.TotalExecutions)
fmt.Printf("成功次数: %d\n", stats.SuccessCount)
fmt.Printf("失败次数: %d\n", stats.FailureCount)
fmt.Printf("平均时间: %v\n", stats.AverageTime)
```

### 2. 容器监控

```go
// 获取容器状态
status := module.GetContainerStatus()
fmt.Printf("活动容器: %d\n", status.ActiveContainers)
fmt.Printf("池大小: %d\n", status.PoolSize)
fmt.Printf("复用次数: %d\n", status.ReuseCount)
```

### 3. 性能分析

```go
import "runtime/pprof"

// CPU 分析
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// 内存分析
f, _ = os.Create("mem.prof")
pprof.WriteHeapProfile(f)
```

---

## 性能测试

### 基准测试

```go
func BenchmarkExecuteCode(b *testing.B) {
    module, _ := NewCodeExecutorModule(config)
    defer module.Close()
    
    code := `print("test")`
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        module.ExecuteCode(ctx, "python", code)
    }
}
```

### 压力测试

```go
func StressTest() {
    module, _ := NewCodeExecutorModule(config)
    defer module.Close()
    
    concurrency := 100
    requests := 1000
    
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < requests/concurrency; j++ {
                module.ExecuteCode(ctx, "python", code)
            }
        }()
    }
    
    wg.Wait()
    elapsed := time.Since(start)
    
    fmt.Printf("总请求: %d\n", requests)
    fmt.Printf("总时间: %v\n", elapsed)
    fmt.Printf("QPS: %.2f\n", float64(requests)/elapsed.Seconds())
}
```

---

## 常见性能问题

### 问题 1: 执行速度慢

**症状**: 代码执行时间过长

**诊断**:
```go
start := time.Now()
result, _ := module.ExecuteCode(ctx, lang, code)
elapsed := time.Since(start)
fmt.Printf("执行时间: %v\n", elapsed)
```

**解决**:
1. 使用 Yaegi（Go 代码）
2. 启用缓存
3. 启用容器池
4. 增加资源限制

---

### 问题 2: 内存使用高

**症状**: 内存占用持续增长

**诊断**:
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("内存使用: %d MB\n", m.Alloc/1024/1024)
```

**解决**:
1. 减少缓存容量
2. 启用自动清理
3. 定期重启模块
4. 优化代码

---

### 问题 3: 容器启动慢

**症状**: 容器模式下执行很慢

**解决**:
1. 启用容器池
2. 使用 Alpine 镜像
3. 预拉取镜像
4. 增加资源

---

### 问题 4: 吞吐量低

**症状**: 无法处理高并发

**解决**:
1. 增加容器池大小
2. 使用本地模式
3. 优化代码
4. 水平扩展

---

## 最佳实践

### 1. 开发环境配置

```yaml
executor:
  timeout: 120000
  memory_limit: 1024
  cpu_limit: 4
  execution_mode: local

yaegi:
  enable_cache: true
  cache_capacity: 100

container:
  enabled: false
```

### 2. 生产环境配置

```yaml
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 2
  execution_mode: container

yaegi:
  enable_cache: true
  cache_capacity: 500

container:
  enabled: true
  enable_pool: true
  pool_min_size: 10
  pool_max_size: 50
```

### 3. 高性能配置

```yaml
executor:
  timeout: 60000
  memory_limit: 2048
  cpu_limit: 8
  execution_mode: local

yaegi:
  enable_cache: true
  cache_capacity: 1000

container:
  enabled: true
  enable_pool: true
  pool_min_size: 20
  pool_max_size: 100
```

---

## 性能优化检查清单

- [ ] 使用 Yaegi 执行 Go 代码
- [ ] 启用编译缓存
- [ ] 启用容器池
- [ ] 使用 Alpine 镜像
- [ ] 预加载常用包
- [ ] 合理设置资源限制
- [ ] 实施性能监控
- [ ] 定期性能测试
- [ ] 优化并发处理
- [ ] 启用自动清理

---

**版本**: 1.0  
**更新日期**: 2026-01-31

