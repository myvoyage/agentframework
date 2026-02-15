# 容器执行 API 文档

## 概述

容器执行模块提供了基于 Docker 的安全代码执行环境，实现了完整的资源隔离、网络隔离和安全控制。支持容器池管理，可以显著提升执行性能。

## 核心特性

- **完全隔离**: 网络、文件系统、进程完全隔离
- **资源限制**: CPU、内存限制
- **容器池**: 预创建容器，80% 启动时间减少
- **自动清理**: 执行后自动清理容器
- **多语言支持**: Python、JavaScript、Go、Bash
- **健康检查**: 容器健康状态监控

---

## 快速开始

### 基本使用

```go
// 创建容器执行器
config := ContainerConfig{
    Enabled:      true,
    CPULimit:     "0.5",
    MemoryLimit:  "512m",
    NetworkMode:  "none",
    Timeout:      30 * time.Second,
    AutoCleanup:  true,
}

executor, err := NewContainerExecutor(config)
if err != nil {
    log.Fatal(err)
}
defer executor.Close()

// 执行代码
ctx := context.Background()
result, err := executor.Execute(ctx, "python", "print('Hello from container!')")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Output)
```

---

## API 参考

### ContainerConfig

容器执行器配置。

```go
type ContainerConfig struct {
    // 基本配置
    Enabled       bool              // 是否启用容器执行
    DefaultImages map[string]string // 默认镜像映射
    
    // 资源限制
    CPULimit    string        // CPU 限制 (如 "0.5", "1.0")
    MemoryLimit string        // 内存限制 (如 "512m", "1g")
    NetworkMode string        // 网络模式 ("none", "bridge", "host")
    Timeout     time.Duration // 执行超时时间
    
    // 容器管理
    AutoCleanup bool // 是否自动清理容器
    
    // 容器池配置
    EnablePool  bool // 是否启用容器池
    PoolMinSize int  // 池最小大小
    PoolMaxSize int  // 池最大大小
}
```

**默认配置**:
```go
config := DefaultContainerConfig()
// Enabled: false
// CPULimit: "0.5"
// MemoryLimit: "512m"
// NetworkMode: "none"
// Timeout: 30s
// AutoCleanup: true
// EnablePool: false
// PoolMinSize: 2
// PoolMaxSize: 10
```

**默认镜像**:
```go
DefaultImages: map[string]string{
    "python":     "python:3.11-alpine",
    "javascript": "node:18-alpine",
    "go":         "golang:1.21-alpine",
    "bash":       "alpine:latest",
}
```

---

### ContainerExecutor

#### 创建执行器

```go
// 使用默认配置
executor, err := NewContainerExecutor(DefaultContainerConfig())

// 使用自定义配置
config := ContainerConfig{
    Enabled:      true,
    CPULimit:     "1.0",
    MemoryLimit:  "1g",
    NetworkMode:  "none",
    Timeout:      60 * time.Second,
    AutoCleanup:  true,
    EnablePool:   true,
    PoolMinSize:  3,
    PoolMaxSize:  10,
}
executor, err := NewContainerExecutor(config)
```

#### Execute 方法

在容器中执行代码。

**签名**:
```go
func (ce *ContainerExecutor) Execute(ctx context.Context, language, code string) (*ExecutionResult, error)
```

**参数**:
- `ctx` (context.Context): 上下文，用于超时控制
- `language` (string): 编程语言
- `code` (string): 要执行的代码

**返回值**: `(*ExecutionResult, error)`

**示例**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

code := `
import sys
print(f"Python version: {sys.version}")
print("Hello from container!")
`

result, err := executor.Execute(ctx, "python", code)
if err != nil {
    log.Fatal(err)
}

if result.Success {
    fmt.Println(result.Output)
    fmt.Printf("执行时间: %v\n", result.Duration)
    fmt.Printf("内存使用: %d MB\n", result.MemoryMB)
}
```

#### IsEnabled 方法

检查容器执行器是否启用。

```go
if executor.IsEnabled() {
    fmt.Println("容器执行器已启用")
} else {
    fmt.Println("容器执行器未启用")
}
```

#### CheckConnection 方法

检查 Docker 连接。

```go
err := executor.CheckConnection()
if err != nil {
    fmt.Printf("Docker 连接失败: %v\n", err)
} else {
    fmt.Println("Docker 连接正常")
}
```

#### GetStats 方法

获取执行统计信息。

```go
stats := executor.GetStats()
fmt.Printf("总执行次数: %d\n", stats["total_executions"])
fmt.Printf("成功次数: %d\n", stats["success_count"])
fmt.Printf("失败次数: %d\n", stats["failure_count"])
fmt.Printf("活跃容器: %d\n", stats["active_containers"])
```

#### ListContainers 方法

列出所有容器。

```go
containers, err := executor.ListContainers()
if err != nil {
    log.Fatal(err)
}

for _, container := range containers {
    fmt.Printf("容器 ID: %s\n", container.ID)
    fmt.Printf("  状态: %s\n", container.Status)
    fmt.Printf("  镜像: %s\n", container.Image)
}
```

#### Close 方法

关闭执行器并清理资源。

```go
err := executor.Close()
if err != nil {
    log.Printf("关闭执行器失败: %v\n", err)
}
```

---

## 容器池

容器池通过预创建容器来减少启动时间，可以实现 **80%** 的性能提升。

### 启用容器池

```go
config := ContainerConfig{
    Enabled:     true,
    EnablePool:  true,
    PoolMinSize: 3,  // 最少保持 3 个容器
    PoolMaxSize: 10, // 最多 10 个容器
}

executor, err := NewContainerExecutor(config)
```

### 容器池工作原理

1. **预创建**: 启动时创建 PoolMinSize 个容器
2. **获取**: 执行时从池中获取容器
3. **复用**: 执行完成后容器返回池中
4. **扩展**: 池不足时自动创建新容器（不超过 PoolMaxSize）
5. **健康检查**: 定期检查容器健康状态

### 容器池统计

```go
if executor.pool != nil {
    stats := executor.pool.GetStats()
    fmt.Printf("池统计:\n")
    fmt.Printf("  总容器数: %d\n", stats["total_containers"])
    fmt.Printf("  可用容器: %d\n", stats["available_containers"])
    fmt.Printf("  使用中: %d\n", stats["in_use_containers"])
    fmt.Printf("  获取次数: %d\n", stats["acquire_count"])
    fmt.Printf("  释放次数: %d\n", stats["release_count"])
}
```

---

## 资源限制

### CPU 限制

```go
config := ContainerConfig{
    CPULimit: "0.5", // 限制为 0.5 个 CPU 核心
}

// 其他示例
// "1.0"  - 1 个 CPU 核心
// "2.0"  - 2 个 CPU 核心
// "0.25" - 0.25 个 CPU 核心
```

### 内存限制

```go
config := ContainerConfig{
    MemoryLimit: "512m", // 限制为 512 MB
}

// 其他示例
// "256m" - 256 MB
// "1g"   - 1 GB
// "2g"   - 2 GB
```

### 网络模式

```go
config := ContainerConfig{
    NetworkMode: "none", // 完全隔离网络
}

// 可选模式
// "none"   - 无网络访问（推荐，最安全）
// "bridge" - 桥接网络（有限网络访问）
// "host"   - 主机网络（不推荐，安全风险）
```

### 超时控制

```go
config := ContainerConfig{
    Timeout: 30 * time.Second, // 30 秒超时
}

// 也可以在执行时指定
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := executor.Execute(ctx, language, code)
```

---

## 镜像管理

### 自定义镜像

```go
config := ContainerConfig{
    DefaultImages: map[string]string{
        "python":     "python:3.11-alpine",
        "javascript": "node:18-alpine",
        "go":         "golang:1.21-alpine",
        "bash":       "alpine:latest",
        "custom":     "myregistry/custom-image:latest",
    },
}
```

### 镜像拉取

容器执行器会自动检查并拉取镜像：

```go
// 自动拉取镜像（如果不存在）
result, err := executor.Execute(ctx, "python", code)
// 首次执行会拉取镜像，后续执行使用缓存的镜像
```

---

## 安全特性

### 1. 网络隔离

```go
config := ContainerConfig{
    NetworkMode: "none", // 完全禁用网络
}
```

**效果**: 容器内代码无法访问任何网络资源。

### 2. 文件系统隔离

容器使用独立的文件系统，无法访问主机文件。

### 3. 进程隔离

容器内进程与主机进程完全隔离。

### 4. 资源限制

```go
config := ContainerConfig{
    CPULimit:    "0.5",  // 限制 CPU 使用
    MemoryLimit: "512m", // 限制内存使用
}
```

**效果**: 防止恶意代码消耗过多资源。

### 5. 非 root 用户

容器内代码以非 root 用户身份运行（如果镜像支持）。

### 6. 自动清理

```go
config := ContainerConfig{
    AutoCleanup: true, // 执行后自动删除容器
}
```

**效果**: 防止容器堆积，节省资源。

---

## 性能优化

### 1. 使用容器池

```go
config := ContainerConfig{
    EnablePool:  true,
    PoolMinSize: 5,
    PoolMaxSize: 20,
}
```

**效果**: 减少 80% 的容器启动时间。

### 2. 使用轻量级镜像

```go
config := ContainerConfig{
    DefaultImages: map[string]string{
        "python": "python:3.11-alpine", // Alpine 镜像更小更快
    },
}
```

### 3. 预拉取镜像

```bash
# 在部署前预拉取镜像
docker pull python:3.11-alpine
docker pull node:18-alpine
docker pull golang:1.21-alpine
```

### 4. 调整超时时间

```go
config := ContainerConfig{
    Timeout: 10 * time.Second, // 根据实际需求调整
}
```

---

## 监控和日志

### 获取容器日志

```go
// 执行代码
result, err := executor.Execute(ctx, language, code)

// 查看输出
fmt.Println("标准输出:", result.Output)
fmt.Println("标准错误:", result.Error)
```

### 监控资源使用

```go
result, err := executor.Execute(ctx, language, code)

fmt.Printf("资源使用:\n")
fmt.Printf("  执行时间: %v\n", result.Duration)
fmt.Printf("  内存使用: %d MB\n", result.MemoryMB)
fmt.Printf("  退出码: %d\n", result.ExitCode)
```

### 统计信息

```go
stats := executor.GetStats()

fmt.Printf("执行统计:\n")
fmt.Printf("  总次数: %d\n", stats["total_executions"])
fmt.Printf("  成功: %d\n", stats["success_count"])
fmt.Printf("  失败: %d\n", stats["failure_count"])
fmt.Printf("  活跃容器: %d\n", stats["active_containers"])
```

---

## 配置示例

### 开发环境配置

```yaml
container:
  enabled: true
  cpu_limit: "1.0"
  memory_limit: "1g"
  network_mode: bridge  # 允许网络访问（开发调试）
  timeout: 60s
  auto_cleanup: true
  enable_pool: false    # 开发环境不需要池
```

### 生产环境配置

```yaml
container:
  enabled: true
  cpu_limit: "0.5"
  memory_limit: "512m"
  network_mode: none    # 完全隔离网络
  timeout: 30s
  auto_cleanup: true
  enable_pool: true     # 启用池提升性能
  pool_min_size: 5
  pool_max_size: 20
```

### 高安全配置

```yaml
container:
  enabled: true
  cpu_limit: "0.25"     # 严格限制 CPU
  memory_limit: "256m"  # 严格限制内存
  network_mode: none    # 禁用网络
  timeout: 10s          # 短超时
  auto_cleanup: true
  enable_pool: false    # 不复用容器
```

---

## 故障排查

### 问题：Docker 连接失败

**症状**: `docker connection failed`

**解决方案**:
1. 检查 Docker 是否运行
2. 检查 Docker 权限
3. 检查 Docker socket 路径

```bash
# 检查 Docker 状态
docker ps

# 检查 Docker 版本
docker version

# 测试 Docker 连接
docker run hello-world
```

### 问题：镜像拉取失败

**症状**: `failed to pull image`

**解决方案**:
1. 检查网络连接
2. 检查镜像名称是否正确
3. 手动拉取镜像

```bash
# 手动拉取镜像
docker pull python:3.11-alpine

# 查看本地镜像
docker images
```

### 问题：容器启动超时

**症状**: 容器创建或启动时间过长

**解决方案**:
1. 使用更轻量的镜像
2. 启用容器池
3. 预拉取镜像

```go
// 使用 Alpine 镜像
config.DefaultImages["python"] = "python:3.11-alpine"

// 启用容器池
config.EnablePool = true
config.PoolMinSize = 3
```

### 问题：内存不足

**症状**: `OOMKilled` 或内存相关错误

**解决方案**:
1. 增加内存限制
2. 优化代码
3. 检查是否有内存泄漏

```go
// 增加内存限制
config.MemoryLimit = "1g"
```

### 问题：容器清理失败

**症状**: 容器堆积，占用资源

**解决方案**:
1. 确保 AutoCleanup 启用
2. 手动清理容器
3. 检查 Close 方法是否被调用

```bash
# 手动清理所有停止的容器
docker container prune -f

# 查看所有容器
docker ps -a
```

---

## 最佳实践

### 1. 始终启用自动清理

```go
config := ContainerConfig{
    AutoCleanup: true, // 防止容器堆积
}
```

### 2. 使用网络隔离

```go
config := ContainerConfig{
    NetworkMode: "none", // 最安全的选择
}
```

### 3. 设置合理的资源限制

```go
config := ContainerConfig{
    CPULimit:    "0.5",  // 防止 CPU 占用过高
    MemoryLimit: "512m", // 防止内存耗尽
}
```

### 4. 生产环境使用容器池

```go
config := ContainerConfig{
    EnablePool:  true,
    PoolMinSize: 5,
    PoolMaxSize: 20,
}
```

### 5. 设置超时时间

```go
config := ContainerConfig{
    Timeout: 30 * time.Second, // 防止长时间运行
}
```

### 6. 监控容器状态

```go
// 定期检查统计信息
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := executor.GetStats()
        log.Printf("容器统计: %+v", stats)
    }
}()
```

### 7. 优雅关闭

```go
// 使用 defer 确保资源清理
defer executor.Close()

// 或在程序退出时清理
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
    <-c
    executor.Close()
    os.Exit(0)
}()
```

---

## 示例：完整的容器执行流程

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/yourorg/agent/aiosandbox/code_exec"
)

func executeInContainer(language, code string) error {
    // 1. 创建配置
    config := code_exec.ContainerConfig{
        Enabled:      true,
        CPULimit:     "0.5",
        MemoryLimit:  "512m",
        NetworkMode:  "none",
        Timeout:      30 * time.Second,
        AutoCleanup:  true,
        EnablePool:   true,
        PoolMinSize:  3,
        PoolMaxSize:  10,
    }
    
    // 2. 创建执行器
    executor, err := code_exec.NewContainerExecutor(config)
    if err != nil {
        return fmt.Errorf("创建执行器失败: %w", err)
    }
    defer executor.Close()
    
    // 3. 检查 Docker 连接
    if err := executor.CheckConnection(); err != nil {
        return fmt.Errorf("Docker 连接失败: %w", err)
    }
    
    // 4. 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 5. 执行代码
    fmt.Println("🚀 开始执行...")
    start := time.Now()
    
    result, err := executor.Execute(ctx, language, code)
    if err != nil {
        return fmt.Errorf("执行失败: %w", err)
    }
    
    duration := time.Since(start)
    
    // 6. 处理结果
    if result.Success {
        fmt.Println("✅ 执行成功")
        fmt.Printf("\n输出:\n%s\n", result.Output)
    } else {
        fmt.Println("❌ 执行失败")
        fmt.Printf("\n错误:\n%s\n", result.Error)
    }
    
    // 7. 显示资源使用
    fmt.Printf("\n📊 资源使用:\n")
    fmt.Printf("  执行时间: %v\n", duration)
    fmt.Printf("  内存使用: %d MB\n", result.MemoryMB)
    fmt.Printf("  退出码: %d\n", result.ExitCode)
    
    // 8. 显示统计信息
    stats := executor.GetStats()
    fmt.Printf("\n📈 执行统计:\n")
    fmt.Printf("  总次数: %d\n", stats["total_executions"])
    fmt.Printf("  成功: %d\n", stats["success_count"])
    fmt.Printf("  失败: %d\n", stats["failure_count"])
    
    return nil
}

func main() {
    code := `
import sys
import platform

print(f"Python version: {sys.version}")
print(f"Platform: {platform.platform()}")
print("Hello from secure container!")

# 尝试网络访问（会失败，因为网络被隔离）
try:
    import urllib.request
    urllib.request.urlopen('http://example.com')
    print("Network access: OK")
except Exception as e:
    print(f"Network access: BLOCKED ({type(e).__name__})")
`
    
    if err := executeInContainer("python", code); err != nil {
        log.Fatal(err)
    }
}
```

---

## 参考资料

- [Docker 官方文档](https://docs.docker.com/)
- [容器安全最佳实践](./CONTAINER_SECURITY_BEST_PRACTICES.md)
- [性能优化指南](./PERFORMANCE_OPTIMIZATION.md)

---

**版本**: 1.0  
**更新日期**: 2026-01-31
