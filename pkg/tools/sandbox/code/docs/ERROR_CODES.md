# 错误代码说明

## 概述

本文档列出代码执行模块的所有错误代码、原因和解决方法。

---

## 错误代码分类

错误代码格式：`EXEC-XXXX`

- `EXEC-1XXX`: 配置错误
- `EXEC-2XXX`: 执行错误
- `EXEC-3XXX`: 分析错误
- `EXEC-4XXX`: 容器错误
- `EXEC-5XXX`: Yaegi 错误
- `EXEC-6XXX`: 资源错误

---

## 配置错误 (EXEC-1XXX)

### EXEC-1001: 配置文件不存在

**错误信息**: `configuration file not found: {path}`

**原因**: 指定的配置文件不存在

**解决方法**:
1. 检查文件路径是否正确
2. 确认文件存在
3. 使用绝对路径

```go
// 错误
module, err := NewCodeExecutorModuleFromFile("config.yaml")

// 正确
module, err := NewCodeExecutorModuleFromFile("./config.yaml")
```

---

### EXEC-1002: 配置文件格式错误

**错误信息**: `invalid configuration format: {error}`

**原因**: YAML 格式不正确

**解决方法**:
1. 检查 YAML 语法
2. 验证缩进（使用空格，不用 Tab）
3. 使用 YAML 验证工具

```yaml
# 错误（缩进不一致）
executor:
  timeout: 60000
 memory_limit: 512

# 正确
executor:
  timeout: 60000
  memory_limit: 512
```

---

### EXEC-1003: 配置验证失败

**错误信息**: `configuration validation failed: {reason}`

**原因**: 配置值不合法

**解决方法**:
```yaml
# 常见问题
executor:
  timeout: -1        # 错误：必须 > 0
  memory_limit: 0    # 错误：必须 > 0
  cpu_limit: 0       # 错误：必须 > 0
  execution_mode: invalid  # 错误：必须是 local/container/auto
```

---

### EXEC-1004: 不支持的语言

**错误信息**: `language not supported: {language}`

**原因**: 语言未在 supported_languages 中配置

**解决方法**:
```yaml
executor:
  supported_languages:
    - python
    - javascript
    - go
    - bash
```

---

## 执行错误 (EXEC-2XXX)

### EXEC-2001: 执行超时

**错误信息**: `execution timeout after {timeout}ms`

**原因**: 代码执行时间超过配置的超时时间

**解决方法**:
1. 增加超时时间
   ```yaml
   executor:
     timeout: 120000  # 120 秒
   ```

2. 优化代码性能
3. 检查是否有死循环

---

### EXEC-2002: 编译失败

**错误信息**: `compilation failed: {error}`

**原因**: 代码语法错误

**解决方法**:
1. 检查语法错误
2. 验证导入的包
3. 确认语言版本兼容

```python
# 错误
print("Hello  # 缺少引号

# 正确
print("Hello")
```

---

### EXEC-2003: 运行时错误

**错误信息**: `runtime error: {error}`

**原因**: 代码执行时发生错误

**常见原因**:
- 除以零
- 空指针引用
- 数组越界
- 类型错误

**解决方法**:
```python
# 错误
x = 1 / 0

# 正确
x = 1 / 2 if y != 0 else 0
```

---

### EXEC-2004: 退出码非零

**错误信息**: `process exited with code {code}`

**原因**: 程序异常退出

**常见退出码**:
- `1`: 一般错误
- `2`: 误用 shell 命令
- `126`: 命令无法执行
- `127`: 命令未找到
- `130`: 被 Ctrl+C 中断
- `137`: 被 SIGKILL 终止（通常是内存不足）

---

### EXEC-2005: 权限被拒绝

**错误信息**: `permission denied: {operation}`

**原因**: 没有执行权限

**解决方法**:
1. 检查文件权限
2. 容器模式下检查用户权限
3. 避免访问受保护的资源

---

## 分析错误 (EXEC-3XXX)

### EXEC-3001: 分析失败

**错误信息**: `code analysis failed: {error}`

**原因**: 代码分析过程出错

**解决方法**:
1. 检查代码是否完整
2. 验证语言是否正确
3. 查看详细错误信息

---

### EXEC-3002: 自定义规则加载失败

**错误信息**: `failed to load custom rules: {error}`

**原因**: 自定义规则文件格式错误

**解决方法**:
```yaml
# 正确的规则格式
rules:
  - name: "规则名称"
    language: "python"
    pattern: "正则表达式"
    severity: "critical"  # critical/high/medium/low
    message: "错误信息"
    suggestion: "修复建议"
```

---

### EXEC-3003: 规则验证失败

**错误信息**: `rule validation failed: {rule_name}`

**原因**: 规则配置不完整或不正确

**必需字段**:
- name
- language
- pattern
- severity
- message

---

## 容器错误 (EXEC-4XXX)

### EXEC-4001: Docker 不可用

**错误信息**: `docker not available: {error}`

**原因**: Docker 未安装或未运行

**解决方法**:
```bash
# 检查 Docker
docker --version
docker ps

# 启动 Docker
sudo systemctl start docker  # Linux
# 或打开 Docker Desktop (macOS/Windows)
```

或使用本地模式:
```yaml
executor:
  execution_mode: local
```

---

### EXEC-4002: 镜像拉取失败

**错误信息**: `failed to pull image: {image}`

**原因**: 无法下载 Docker 镜像

**解决方法**:
1. 检查网络连接
2. 验证镜像名称
3. 手动拉取镜像
   ```bash
   docker pull python:3.11-alpine
   ```

4. 使用镜像加速器

---

### EXEC-4003: 容器创建失败

**错误信息**: `failed to create container: {error}`

**原因**: 容器配置错误或资源不足

**解决方法**:
1. 检查资源限制是否合理
2. 确认镜像存在
3. 查看 Docker 日志
   ```bash
   docker logs {container_id}
   ```

---

### EXEC-4004: 容器启动失败

**错误信息**: `failed to start container: {error}`

**原因**: 容器无法启动

**常见原因**:
- 端口冲突
- 资源不足
- 配置错误

**解决方法**:
```yaml
# 检查配置
container:
  cpu_limit: "0.5"      # 不要太低
  memory_limit: "256m"  # 不要太低
```

---

### EXEC-4005: 容器执行超时

**错误信息**: `container execution timeout after {timeout}`

**原因**: 容器内代码执行超时

**解决方法**:
```yaml
container:
  timeout: 60s  # 增加超时时间
```

---

## Yaegi 错误 (EXEC-5XXX)

### EXEC-5001: Yaegi 初始化失败

**错误信息**: `failed to initialize yaegi: {error}`

**原因**: Yaegi 解释器初始化失败

**解决方法**:
1. 检查 Go 版本（需要 1.21+）
2. 重新安装依赖
3. 使用 go run 模式

---

### EXEC-5002: 包导入失败

**错误信息**: `failed to import package: {package}`

**原因**: Yaegi 不支持该包

**解决方法**:
1. 检查包是否在支持列表中
2. 使用 go run 模式
3. 配置自动回退
   ```yaml
   executor:
     execution_mode: auto
   ```

---

### EXEC-5003: 编译缓存错误

**错误信息**: `cache error: {error}`

**原因**: 缓存操作失败

**解决方法**:
```yaml
# 禁用缓存
yaegi:
  enable_cache: false

# 或减少缓存容量
yaegi:
  cache_capacity: 50
```

---

## 资源错误 (EXEC-6XXX)

### EXEC-6001: 内存不足

**错误信息**: `out of memory`

**原因**: 代码使用内存超过限制

**解决方法**:
1. 增加内存限制
   ```yaml
   executor:
     memory_limit: 1024  # 1 GB
   ```

2. 优化代码内存使用
3. 检查内存泄漏

---

### EXEC-6002: CPU 限制

**错误信息**: `CPU limit exceeded`

**原因**: CPU 使用超过限制

**解决方法**:
```yaml
executor:
  cpu_limit: 4  # 增加 CPU 限制
```

---

### EXEC-6003: 磁盘空间不足

**错误信息**: `no space left on device`

**原因**: 磁盘空间不足

**解决方法**:
1. 清理临时文件
2. 清理 Docker 镜像
   ```bash
   docker system prune -a
   ```
3. 增加磁盘空间

---

### EXEC-6004: 文件描述符耗尽

**错误信息**: `too many open files`

**原因**: 打开的文件过多

**解决方法**:
1. 增加文件描述符限制
   ```bash
   ulimit -n 4096
   ```

2. 确保代码正确关闭文件
3. 启用自动清理

---

## 错误处理最佳实践

### 1. 捕获和记录错误

```go
result, err := module.ExecuteCode(ctx, lang, code)
if err != nil {
    log.Printf("执行失败: %v", err)
    // 记录详细信息
    log.Printf("语言: %s", lang)
    log.Printf("代码长度: %d", len(code))
    return err
}

if !result.Success {
    log.Printf("执行错误: %s", result.Error)
    log.Printf("退出码: %d", result.ExitCode)
}
```

### 2. 优雅降级

```go
// 尝试容器模式
config.ExecutionMode = "container"
result, err := module.ExecuteCode(ctx, lang, code)

if err != nil {
    // 降级到本地模式
    log.Println("容器模式失败，切换到本地模式")
    config.ExecutionMode = "local"
    result, err = module.ExecuteCode(ctx, lang, code)
}
```

### 3. 重试机制

```go
maxRetries := 3
var result *ExecutionResult
var err error

for i := 0; i < maxRetries; i++ {
    result, err = module.ExecuteCode(ctx, lang, code)
    if err == nil && result.Success {
        break
    }
    
    log.Printf("重试 %d/%d", i+1, maxRetries)
    time.Sleep(time.Second * time.Duration(i+1))
}
```

### 4. 错误分类处理

```go
result, err := module.ExecuteCode(ctx, lang, code)

switch {
case err != nil:
    // 系统错误
    handleSystemError(err)
case !result.Success && result.ExitCode == 137:
    // 内存不足
    handleMemoryError(result)
case !result.Success && strings.Contains(result.Error, "timeout"):
    // 超时
    handleTimeoutError(result)
default:
    // 其他错误
    handleOtherError(result)
}
```

---

## 调试技巧

### 1. 启用详细日志

```go
config.Verbose = true
config.LogLevel = "debug"
```

### 2. 测试简单代码

```go
// 测试基础功能
result, _ := module.ExecuteCode(ctx, "python", `print("test")`)
```

### 3. 检查配置

```go
err := ValidateConfig(&config)
if err != nil {
    log.Printf("配置错误: %v", err)
}
```

### 4. 查看容器日志

```bash
# 列出容器
docker ps -a

# 查看日志
docker logs {container_id}

# 进入容器
docker exec -it {container_id} /bin/sh
```

---

## 获取帮助

如果问题仍未解决：

1. 查看 [FAQ](./FAQ.md)
2. 查看 [性能优化指南](./PERFORMANCE_GUIDE.md)
3. 查看 [安全配置指南](./SECURITY_GUIDE.md)
4. 提交 GitHub Issue
5. 联系技术支持

---

**版本**: 1.0  
**更新日期**: 2026-01-31  
**维护者**: Agent Framework Team

