# 常见问题解答 (FAQ)

## 概述

本文档回答代码执行模块使用过程中的常见问题。

---

## 目录

1. [安装和配置](#安装和配置)
2. [代码执行](#代码执行)
3. [Yaegi 解释器](#yaegi-解释器)
4. [容器执行](#容器执行)
5. [代码分析](#代码分析)
6. [性能问题](#性能问题)
7. [错误处理](#错误处理)
8. [最佳实践](#最佳实践)

---

## 安装和配置

### Q1: 如何安装代码执行模块？

**A:** 代码执行模块是 Agent Framework 的一部分，通过以下方式安装：

```bash
go get github.com/your-org/agent-framework/agent/aiosandbox/code_exec
```

然后在代码中导入：

```go
import "github.com/your-org/agent-framework/agent/aiosandbox/code_exec"
```

---

### Q2: 需要安装哪些依赖？

**A:** 根据使用的功能，需要不同的依赖：

**基础功能**（必需）：
- Go 1.21+
- Yaegi（自动安装）

**容器执行**（可选）：
- Docker 20.10+
- Docker 服务运行中

**语言支持**（可选）：
- Python 3.8+（执行 Python 代码）
- Node.js 18+（执行 JavaScript 代码）
- Go 1.21+（执行 Go 代码）
- Bash（执行 Shell 脚本）

---

### Q3: 如何配置代码执行模块？

**A:** 有三种配置方式：

**方式 1: 程序化配置**
```go
config := code_exec.CodeExecutorConfig{
    Timeout:            60000,
    MemoryLimit:        512,
    CPULimit:           2,
    SupportedLanguages: []string{"python", "go"},
    ExecutionMode:      "auto",
}
module, err := code_exec.NewCodeExecutorModule(config)
```

**方式 2: YAML 配置文件**
```go
module, err := code_exec.NewCodeExecutorModuleFromFile("config.yaml")
```

**方式 3: 完整配置**
```go
fullConfig := code_exec.DefaultFullConfig()
fullConfig.Executor.Timeout = 60000
module, err := code_exec.NewCodeExecutorModuleWithFullConfig(&fullConfig)
```

详见 [配置指南](./CONFIGURATION_GUIDE.md)

---

### Q4: 配置文件放在哪里？

**A:** 配置文件可以放在任何位置，通过路径指定：

```go
// 相对路径
module, err := NewCodeExecutorModuleFromFile("config.yaml")

// 绝对路径
module, err := NewCodeExecutorModuleFromFile("/etc/app/config.yaml")

// 用户目录
module, err := NewCodeExecutorModuleFromFile("~/.config/app/config.yaml")
```

推荐位置：
- 开发环境: `./config_dev.yaml`
- 生产环境: `/etc/app/config_prod.yaml`
- 用户配置: `~/.config/app/config.yaml`

---

## 代码执行

### Q5: 支持哪些编程语言？

**A:** 目前支持以下语言：

- **Python** (python, py)
- **JavaScript** (javascript, js)
- **Go** (go)
- **Bash** (bash, sh)

配置示例：
```yaml
executor:
  supported_languages:
    - python
    - javascript
    - go
    - bash
```

---

### Q6: 如何执行代码？

**A:** 使用 `ExecuteCode` 方法：

```go
ctx := context.Background()
result, err := module.ExecuteCode(ctx, "python", `
print("Hello, World!")
print(2 + 2)
`)

if err != nil {
    log.Fatal(err)
}

fmt.Println("输出:", result.Output)
fmt.Println("执行时间:", result.ExecutionTime, "ms")
```

---

### Q7: 代码执行超时怎么办？

**A:** 检查并调整超时配置：

```yaml
executor:
  timeout: 60000  # 60 秒
```

或在代码中设置：
```go
config.Timeout = 120000  // 120 秒
```

如果代码确实需要长时间运行：
1. 增加超时时间
2. 优化代码性能
3. 考虑异步执行

---

### Q8: 如何限制代码的资源使用？

**A:** 通过配置限制 CPU 和内存：

```yaml
executor:
  memory_limit: 512   # 512 MB
  cpu_limit: 2        # 2 个 CPU 核心
  timeout: 60000      # 60 秒
```

容器模式下的额外限制：
```yaml
container:
  cpu_limit: "0.5"      # 0.5 个 CPU 核心
  memory_limit: "512m"  # 512 MB
  timeout: 30s          # 30 秒
```

---

### Q9: 执行模式有什么区别？

**A:** 三种执行模式的对比：

| 特性 | local | container | auto |
|------|-------|-----------|------|
| 速度 | 最快 | 较慢 | 自动选择 |
| 安全性 | 一般 | 最高 | 自动选择 |
| 隔离性 | 无 | 完全隔离 | 自动选择 |
| 依赖 | 无 | 需要 Docker | 无 |
| 适用场景 | 开发环境 | 生产环境 | 推荐使用 |

**推荐**：使用 `auto` 模式，系统会自动选择最佳方式。

---

## Yaegi 解释器

### Q10: 什么是 Yaegi？

**A:** Yaegi 是一个 Go 语言解释器，用于快速执行 Go 代码，无需编译。

**优势**：
- 启动速度快 428 倍（相比 go run）
- 无需 Go 编译器
- 支持大部分标准库
- 内存占用更小

**使用**：
```go
config.ExecutionMode = "local"  // local 模式自动使用 Yaegi
```

---

### Q11: Yaegi 支持哪些 Go 特性？

**A:** Yaegi 支持大部分 Go 特性：

**支持**：
- ✅ 标准库（fmt, strings, time, math 等）
- ✅ 结构体和方法
- ✅ 接口
- ✅ Goroutines
- ✅ Channels
- ✅ 闭包
- ✅ 反射（部分）

**不支持**：
- ❌ CGO
- ❌ 汇编代码
- ❌ 某些底层包（syscall 等）
- ❌ 编译器指令

---

### Q12: Yaegi 执行失败怎么办？

**A:** Yaegi 失败时会自动回退到 `go run`：

```go
config.ExecutionMode = "auto"  // 启用自动回退
```

或手动切换：
```go
// 尝试 Yaegi
result, err := module.ExecuteCode(ctx, "go", code)
if err != nil {
    // 切换到 go run
    config.ExecutionMode = "container"
    result, err = module.ExecuteCode(ctx, "go", code)
}
```

---

### Q13: 如何提高 Yaegi 性能？

**A:** 启用编译缓存：

```yaml
yaegi:
  enable_cache: true
  cache_capacity: 500  # 缓存 500 个代码片段
```

性能提升：
- 首次执行：正常速度
- 缓存命中：提升 12,600 倍

---

## 容器执行

### Q14: 如何启用容器执行？

**A:** 配置容器执行：

```yaml
executor:
  execution_mode: container

container:
  enabled: true
  cpu_limit: "0.5"
  memory_limit: "512m"
  network_mode: none
  auto_cleanup: true
```

**前提条件**：
1. 安装 Docker
2. Docker 服务运行中
3. 有 Docker 权限

---

### Q15: Docker 未安装或未运行怎么办？

**A:** 检查 Docker 状态：

```bash
# 检查 Docker 是否安装
docker --version

# 检查 Docker 是否运行
docker ps

# 启动 Docker（Linux）
sudo systemctl start docker

# 启动 Docker（macOS/Windows）
# 打开 Docker Desktop
```

如果无法使用 Docker：
```yaml
executor:
  execution_mode: local  # 使用本地模式
```

---

### Q16: 容器执行很慢怎么办？

**A:** 启用容器池：

```yaml
container:
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

性能提升：
- 减少 80% 容器启动时间
- 提高 3-5 倍吞吐量

其他优化：
1. 使用 Alpine 镜像（更小更快）
2. 预拉取镜像
3. 增加资源限制

---

### Q17: 如何自定义容器镜像？

**A:** 配置自定义镜像：

```yaml
container:
  default_images:
    python: python:3.11-alpine
    javascript: node:18-alpine
    go: golang:1.21-alpine
    bash: alpine:latest
```

或使用自己的镜像：
```yaml
container:
  default_images:
    python: myregistry.com/python:custom
```

---

### Q18: 容器无法访问网络？

**A:** 这是安全特性，默认禁用网络：

```yaml
container:
  network_mode: none  # 无网络（推荐）
```

如需网络访问：
```yaml
container:
  network_mode: bridge  # 启用网络（谨慎使用）
```

**注意**：启用网络会降低安全性。

---

## 代码分析

### Q19: 如何分析代码安全性？

**A:** 使用 `AnalyzeCode` 方法：

```go
result, err := module.AnalyzeCode(ctx, "python", code)

fmt.Println("安全:", result.Safe)
fmt.Println("评分:", result.Score)
fmt.Println("问题:", result.Issues)
fmt.Println("建议:", result.Suggestions)
```

分析内容：
- 网络操作
- 文件系统操作
- 进程操作
- 加密问题
- 数据库操作
- 代码质量

---

### Q20: 如何添加自定义规则？

**A:** 创建自定义规则文件：

```yaml
# custom_rules.yaml
rules:
  - name: "禁止使用 eval"
    language: "python"
    pattern: "eval\\("
    severity: "critical"
    message: "不要使用 eval()"
    suggestion: "使用 ast.literal_eval()"
```

配置使用：
```yaml
analyzer:
  custom_rules_file: "custom_rules.yaml"
```

详见 [自定义规则示例](../examples/custom_rules_example.go)

---

### Q21: 代码分析误报怎么办？

**A:** 调整分析配置：

```yaml
analyzer:
  strict_mode: false  # 关闭严格模式
  
  # 禁用特定检测
  enable_network_detection: false
  enable_filesystem_detection: true
```

或在代码中添加注释（未来支持）：
```python
# nosec: 这是安全的
eval(trusted_input)
```

---

### Q22: 如何提高代码质量评分？

**A:** 根据建议改进代码：

1. **修复安全问题**（最重要）
   - 移除危险操作
   - 验证用户输入
   - 使用安全的 API

2. **改进代码质量**
   - 使用描述性命名
   - 添加文档字符串
   - 遵循代码风格

3. **优化性能**
   - 避免重复计算
   - 使用高效算法
   - 减少内存使用

---

## 性能问题

### Q23: 代码执行太慢怎么办？

**A:** 性能优化建议：

1. **使用 Yaegi**（Go 代码）
   ```yaml
   executor:
     execution_mode: local
   ```

2. **启用缓存**
   ```yaml
   yaegi:
     enable_cache: true
     cache_capacity: 500
   ```

3. **启用容器池**
   ```yaml
   container:
     enable_pool: true
     pool_min_size: 5
   ```

4. **增加资源限制**
   ```yaml
   executor:
     cpu_limit: 4
     memory_limit: 1024
   ```

---

### Q24: 如何监控性能？

**A:** 使用内置监控：

```go
// 获取执行统计
stats := module.GetStats()
fmt.Printf("总执行次数: %d\n", stats.TotalExecutions)
fmt.Printf("平均时间: %v\n", stats.AverageTime)

// 获取容器状态
status := module.GetContainerStatus()
fmt.Printf("活动容器: %d\n", status.ActiveContainers)
```

或使用 MCP 工具：
```go
result := callTool("code_exec_container_status", {})
```

---

### Q25: 内存使用过高怎么办？

**A:** 优化内存使用：

1. **限制内存**
   ```yaml
   executor:
     memory_limit: 256  # 降低限制
   ```

2. **启用自动清理**
   ```yaml
   container:
     auto_cleanup: true
   ```

3. **减少缓存容量**
   ```yaml
   yaegi:
     cache_capacity: 50  # 减少缓存
   ```

4. **定期重启模块**
   ```go
   module.Close()
   module, _ = NewCodeExecutorModule(config)
   ```

---

## 错误处理

### Q26: 常见错误代码及解决方法？

**A:** 常见错误及解决：

| 错误 | 原因 | 解决方法 |
|------|------|----------|
| `timeout exceeded` | 执行超时 | 增加 timeout 配置 |
| `memory limit exceeded` | 内存不足 | 增加 memory_limit |
| `docker not available` | Docker 未运行 | 启动 Docker 或使用 local 模式 |
| `language not supported` | 语言不支持 | 添加到 supported_languages |
| `compilation failed` | 编译错误 | 检查代码语法 |
| `permission denied` | 权限不足 | 检查文件/Docker 权限 |

---

### Q27: 如何调试执行失败？

**A:** 调试步骤：

1. **检查错误信息**
   ```go
   result, err := module.ExecuteCode(ctx, lang, code)
   if err != nil {
       log.Printf("执行失败: %v", err)
   }
   if !result.Success {
       log.Printf("错误: %s", result.Error)
       log.Printf("退出码: %d", result.ExitCode)
   }
   ```

2. **启用详细日志**
   ```go
   config.Verbose = true
   ```

3. **测试简单代码**
   ```go
   result, _ := module.ExecuteCode(ctx, "python", `print("test")`)
   ```

4. **检查配置**
   ```go
   err := ValidateConfig(&config)
   if err != nil {
       log.Printf("配置错误: %v", err)
   }
   ```

---

### Q28: 代码在本地运行正常，但在模块中失败？

**A:** 可能的原因：

1. **缺少依赖**
   - 检查是否安装了所需的包
   - 容器模式下需要在镜像中安装

2. **路径问题**
   - 使用绝对路径
   - 或在代码中设置工作目录

3. **环境变量**
   - 设置必要的环境变量
   ```yaml
   container:
     env:
       - KEY=value
   ```

4. **权限问题**
   - 容器中以非 root 用户运行
   - 检查文件权限

---

## 最佳实践

### Q29: 生产环境推荐配置？

**A:** 生产环境配置示例：

```yaml
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 1
  execution_mode: container

analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  strict_mode: true

container:
  enabled: true
  cpu_limit: "0.5"
  memory_limit: "512m"
  network_mode: none
  auto_cleanup: true
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20

yaegi:
  enable_cache: true
  cache_capacity: 500
```

---

### Q30: 如何确保代码安全？

**A:** 安全最佳实践：

1. **始终先分析后执行**
   ```go
   analysis, _ := module.AnalyzeCode(ctx, lang, code)
   if !analysis.Safe {
       return errors.New("代码不安全")
   }
   result, _ := module.ExecuteCode(ctx, lang, code)
   ```

2. **使用容器隔离**
   ```yaml
   executor:
     execution_mode: container
   ```

3. **禁用网络访问**
   ```yaml
   container:
     network_mode: none
   ```

4. **限制资源使用**
   ```yaml
   executor:
     timeout: 30000
     memory_limit: 256
     cpu_limit: 1
   ```

5. **启用严格模式**
   ```yaml
   analyzer:
     strict_mode: true
   ```

---

## 获取帮助

### Q31: 在哪里获取更多帮助？

**A:** 资源列表：

- **文档**
  - [API 文档](./ENHANCED_CODE_ANALYSIS_API.md)
  - [配置指南](./CONFIGURATION_GUIDE.md)
  - [MCP 工具文档](./MCP_TOOLS_API.md)

- **示例**
  - [代码分析示例](../examples/enhanced_analysis_example.go)
  - [Yaegi 示例](../examples/yaegi_execution_example.go)
  - [容器执行示例](../examples/container_execution_example.go)

- **故障排查**
  - [错误代码说明](./ERROR_CODES.md)
  - [性能优化指南](./PERFORMANCE_GUIDE.md)
  - [安全配置指南](./SECURITY_GUIDE.md)

- **社区**
  - GitHub Issues
  - 讨论论坛
  - 技术支持

---

**版本**: 1.0  
**更新日期**: 2026-01-31  
**维护者**: Agent Framework Team

