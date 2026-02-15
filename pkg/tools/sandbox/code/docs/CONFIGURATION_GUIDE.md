# 配置说明文档

## 概述

代码执行模块提供了灵活的配置系统，支持 YAML 文件配置和程序化配置。本文档详细说明所有配置选项及其用法。

---

## 配置结构

完整配置分为四个部分：

```
FullConfig
├── Executor  (执行器配置)
├── Analyzer  (分析器配置)
├── Yaegi     (Yaegi 解释器配置)
└── Container (容器配置)
```

---

## 1. 执行器配置 (Executor)

### CodeExecutorConfig

```go
type CodeExecutorConfig struct {
    Timeout            int      // 执行超时时间（毫秒）
    MemoryLimit        int      // 内存限制（MB）
    CPULimit           int      // CPU 限制（核心数）
    SupportedLanguages []string // 支持的语言列表
    ExecutionMode      string   // 执行模式
    ContainerConfig    ContainerConfig // 容器配置
}
```

### 配置选项

#### Timeout
- **类型**: int
- **单位**: 毫秒
- **默认值**: 60000 (60秒)
- **说明**: 代码执行的最大时间
- **建议值**:
  - 开发环境: 60000-120000
  - 生产环境: 30000-60000
  - 高安全环境: 10000-30000

```yaml
executor:
  timeout: 60000  # 60 秒
```

#### MemoryLimit
- **类型**: int
- **单位**: MB
- **默认值**: 512
- **说明**: 代码执行的最大内存使用
- **建议值**:
  - 轻量任务: 256-512 MB
  - 一般任务: 512-1024 MB
  - 重量任务: 1024-2048 MB

```yaml
executor:
  memory_limit: 512  # 512 MB
```

#### CPULimit
- **类型**: int
- **单位**: 核心数
- **默认值**: 2
- **说明**: 代码执行可使用的 CPU 核心数
- **建议值**:
  - 单核: 1
  - 双核: 2
  - 多核: 4-8

```yaml
executor:
  cpu_limit: 2  # 2 核
```

#### SupportedLanguages
- **类型**: []string
- **默认值**: ["go", "python", "javascript", "bash"]
- **说明**: 支持的编程语言列表
- **可选值**: "go", "python", "javascript", "js", "bash", "sh"

```yaml
executor:
  supported_languages:
    - python
    - javascript
    - go
    - bash
```

#### ExecutionMode
- **类型**: string
- **默认值**: "local"
- **说明**: 代码执行模式
- **可选值**:
  - `local`: 本地执行（快速，适合开发）
  - `container`: 容器执行（安全，适合生产）
  - `auto`: 自动选择（推荐）

```yaml
executor:
  execution_mode: auto  # 推荐使用 auto
```

### 完整示例

```yaml
executor:
  timeout: 60000
  memory_limit: 512
  cpu_limit: 2
  supported_languages:
    - python
    - javascript
    - go
    - bash
  execution_mode: local
```

---

## 2. 分析器配置 (Analyzer)

### AnalyzerConfig

```go
type AnalyzerConfig struct {
    EnableNetworkDetection    bool   // 启用网络操作检测
    EnableFileSystemDetection bool   // 启用文件系统操作检测
    EnableProcessDetection    bool   // 启用进程操作检测
    EnableCryptoDetection     bool   // 启用加密问题检测
    EnableDatabaseDetection   bool   // 启用数据库操作检测
    EnableQualityCheck        bool   // 启用代码质量检查
    CustomRulesFile           string // 自定义规则文件路径
    StrictMode                bool   // 严格模式
}
```

### 配置选项

#### 检测开关

所有检测默认启用，可根据需求关闭：

```yaml
analyzer:
  enable_network_detection: true     # 检测网络操作
  enable_filesystem_detection: true  # 检测文件系统操作
  enable_process_detection: true     # 检测进程操作
  enable_crypto_detection: true      # 检测加密问题
  enable_database_detection: true    # 检测数据库操作
  enable_quality_check: true         # 检测代码质量
```

**使用场景**:
- 全部启用: 最高安全级别
- 部分启用: 根据实际需求调整
- 全部禁用: 仅做基本检查（不推荐）

#### CustomRulesFile
- **类型**: string
- **默认值**: "" (空)
- **说明**: 自定义规则文件的路径
- **格式**: YAML

```yaml
analyzer:
  custom_rules_file: "custom_rules.yaml"
```

自定义规则文件格式：
```yaml
rules:
  - name: "禁止使用 eval"
    language: "python"
    pattern: "eval\\("
    severity: "high"
    message: "不要使用 eval()"
    suggestion: "使用 ast.literal_eval()"
```

#### StrictMode
- **类型**: bool
- **默认值**: false
- **说明**: 严格模式，任何问题都标记为不安全
- **建议**:
  - 开发环境: false
  - 生产环境: true
  - 高安全环境: true

```yaml
analyzer:
  strict_mode: false
```

### 完整示例

```yaml
analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  custom_rules_file: ""
  strict_mode: false
```

---

## 3. Yaegi 配置

### YaegiConfig

```go
type YaegiConfig struct {
    PreloadStdlib   bool     // 是否预加载标准库
    PreloadPackages []string // 预加载的包列表
    EnableCache     bool     // 是否启用编译缓存
    CacheCapacity   int      // 缓存容量
}
```

### 配置选项

#### PreloadStdlib
- **类型**: bool
- **默认值**: true
- **说明**: 是否预加载 Go 标准库
- **建议**: 始终启用（提升性能）

```yaml
yaegi:
  preload_stdlib: true
```

#### PreloadPackages
- **类型**: []string
- **默认值**: ["fmt", "strings", "time", "math"]
- **说明**: 预加载的包列表
- **常用包**:
  - fmt, strings, time, math
  - encoding/json, io, os
  - net/http, context, sync

```yaml
yaegi:
  preload_packages:
    - fmt
    - strings
    - time
    - math
    - encoding/json
```

#### EnableCache
- **类型**: bool
- **默认值**: true
- **说明**: 是否启用编译结果缓存
- **效果**: 缓存可提升 12,600 倍性能
- **建议**: 始终启用

```yaml
yaegi:
  enable_cache: true
```

#### CacheCapacity
- **类型**: int
- **默认值**: 100
- **说明**: 缓存容量（缓存的代码数量）
- **建议值**:
  - 小型应用: 50-100
  - 中型应用: 100-500
  - 大型应用: 500-1000

```yaml
yaegi:
  cache_capacity: 100
```

### 完整示例

```yaml
yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
    - strings
    - time
    - math
  enable_cache: true
  cache_capacity: 100
```

---

## 4. 容器配置 (Container)

### ContainerConfig

```go
type ContainerConfig struct {
    Enabled       bool              // 是否启用容器执行
    DefaultImages map[string]string // 默认镜像映射
    CPULimit      string            // CPU 限制
    MemoryLimit   string            // 内存限制
    NetworkMode   string            // 网络模式
    Timeout       time.Duration     // 超时时间
    AutoCleanup   bool              // 自动清理
    EnablePool    bool              // 启用容器池
    PoolMinSize   int               // 池最小大小
    PoolMaxSize   int               // 池最大大小
}
```

### 配置选项

#### Enabled
- **类型**: bool
- **默认值**: false
- **说明**: 是否启用容器执行
- **注意**: 需要 Docker 环境

```yaml
container:
  enabled: true
```

#### DefaultImages
- **类型**: map[string]string
- **默认值**: 
  ```
  python: python:3.11-alpine
  javascript: node:18-alpine
  go: golang:1.21-alpine
  bash: alpine:latest
  ```
- **说明**: 各语言的默认 Docker 镜像

```yaml
container:
  default_images:
    python: python:3.11-alpine
    javascript: node:18-alpine
    go: golang:1.21-alpine
    bash: alpine:latest
```

#### CPULimit
- **类型**: string
- **默认值**: "0.5"
- **格式**: "0.5", "1.0", "2.0"
- **说明**: CPU 核心数限制

```yaml
container:
  cpu_limit: "0.5"  # 0.5 个 CPU 核心
```

#### MemoryLimit
- **类型**: string
- **默认值**: "512m"
- **格式**: "256m", "512m", "1g", "2g"
- **说明**: 内存限制

```yaml
container:
  memory_limit: "512m"  # 512 MB
```

#### NetworkMode
- **类型**: string
- **默认值**: "none"
- **可选值**:
  - `none`: 无网络（最安全）
  - `bridge`: 桥接网络
  - `host`: 主机网络（不推荐）

```yaml
container:
  network_mode: none  # 推荐使用 none
```

#### Timeout
- **类型**: time.Duration
- **默认值**: 30s
- **格式**: "10s", "30s", "1m", "5m"
- **说明**: 容器执行超时时间

```yaml
container:
  timeout: 30s
```

#### AutoCleanup
- **类型**: bool
- **默认值**: true
- **说明**: 执行后自动清理容器
- **建议**: 始终启用

```yaml
container:
  auto_cleanup: true
```

#### EnablePool
- **类型**: bool
- **默认值**: false
- **说明**: 是否启用容器池
- **效果**: 减少 80% 启动时间
- **建议**: 生产环境启用

```yaml
container:
  enable_pool: true
```

#### PoolMinSize / PoolMaxSize
- **类型**: int
- **默认值**: 2 / 10
- **说明**: 容器池的最小/最大容器数
- **建议值**:
  - 低负载: 2-5 / 5-10
  - 中负载: 5-10 / 10-20
  - 高负载: 10-20 / 20-50

```yaml
container:
  pool_min_size: 5
  pool_max_size: 20
```

### 完整示例

```yaml
container:
  enabled: true
  default_images:
    python: python:3.11-alpine
    javascript: node:18-alpine
    go: golang:1.21-alpine
    bash: alpine:latest
  cpu_limit: "0.5"
  memory_limit: "512m"
  network_mode: none
  timeout: 30s
  auto_cleanup: true
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

---

## 配置文件示例

### 开发环境配置

```yaml
# config_dev.yaml
executor:
  timeout: 120000
  memory_limit: 1024
  cpu_limit: 2
  supported_languages:
    - python
    - javascript
    - go
    - bash
  execution_mode: local

analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  custom_rules_file: ""
  strict_mode: false

yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
    - strings
    - time
  enable_cache: true
  cache_capacity: 50

container:
  enabled: false
```

### 生产环境配置

```yaml
# config_prod.yaml
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 1
  supported_languages:
    - python
    - javascript
    - go
  execution_mode: container

analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  custom_rules_file: "production_rules.yaml"
  strict_mode: true

yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
    - strings
    - time
    - math
    - encoding/json
  enable_cache: true
  cache_capacity: 500

container:
  enabled: true
  cpu_limit: "0.5"
  memory_limit: "512m"
  network_mode: none
  timeout: 30s
  auto_cleanup: true
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

### 高安全环境配置

```yaml
# config_secure.yaml
executor:
  timeout: 10000
  memory_limit: 256
  cpu_limit: 1
  supported_languages:
    - python
  execution_mode: container

analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  custom_rules_file: "strict_rules.yaml"
  strict_mode: true

yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
  enable_cache: false

container:
  enabled: true
  cpu_limit: "0.25"
  memory_limit: "256m"
  network_mode: none
  timeout: 10s
  auto_cleanup: true
  enable_pool: false
```

---

## 使用配置

### 1. 从文件加载

```go
// 加载配置文件
module, err := NewCodeExecutorModuleFromFile("config.yaml")
if err != nil {
    log.Fatal(err)
}
defer module.Close()
```

### 2. 程序化配置

```go
// 创建配置
config := DefaultFullConfig()
config.Executor.Timeout = 60000
config.Executor.ExecutionMode = "auto"
config.Yaegi.EnableCache = true
config.Container.Enabled = true

// 创建模块
module, err := NewCodeExecutorModuleWithFullConfig(&config)
if err != nil {
    log.Fatal(err)
}
defer module.Close()
```

### 3. 合并配置

```go
// 基础配置
base := DefaultFullConfig()

// 覆盖配置
override := DefaultFullConfig()
override.Executor.Timeout = 30000
override.Container.Enabled = true

// 合并
merged := MergeConfigs(&base, &override)

module, err := NewCodeExecutorModuleWithFullConfig(merged)
```

### 4. 运行时更新

```go
// 创建新配置
newConfig := DefaultFullConfig()
newConfig.Executor.Timeout = 60000

// 应用到模块
err := ApplyConfigToModule(module, &newConfig)
if err != nil {
    log.Fatal(err)
}
```

---

## 配置验证

配置会在加载时自动验证：

```go
config := DefaultFullConfig()
config.Executor.Timeout = -1  // 无效值

err := ValidateConfig(&config)
if err != nil {
    fmt.Println("配置无效:", err)
    // 输出: 配置无效: executor timeout must be positive
}
```

### 验证规则

- Timeout > 0
- MemoryLimit > 0
- CPULimit > 0
- SupportedLanguages 非空
- ExecutionMode 为 "local", "container", "auto" 之一
- CacheCapacity >= 0
- 如果 Container.Enabled，则 Timeout > 0
- PoolMaxSize >= PoolMinSize

---

## 最佳实践

### 1. 环境分离

为不同环境使用不同配置：

```bash
# 开发
./app --config=config_dev.yaml

# 测试
./app --config=config_test.yaml

# 生产
./app --config=config_prod.yaml
```

### 2. 安全优先

生产环境建议：
- 启用容器执行
- 启用严格模式
- 禁用网络访问
- 限制资源使用

### 3. 性能优化

- 启用 Yaegi 缓存
- 启用容器池
- 预加载常用包
- 合理设置超时

### 4. 监控配置

```go
// 定期检查配置
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := module.GetStats()
        log.Printf("执行统计: %+v", stats)
    }
}()
```

---

## 故障排查

### 配置加载失败

```go
config, err := LoadConfigFromFile("config.yaml")
if err != nil {
    log.Printf("加载配置失败: %v", err)
    // 使用默认配置
    config = DefaultFullConfig()
}
```

### 配置验证失败

```go
err := ValidateConfig(&config)
if err != nil {
    log.Printf("配置验证失败: %v", err)
    // 修复配置或使用默认值
}
```

### 配置不生效

检查：
1. 配置文件路径是否正确
2. YAML 格式是否正确
3. 配置是否被后续代码覆盖

---

## 参考资料

- [配置示例文件](../config_example.yaml)
- [API 文档](./ENHANCED_CODE_ANALYSIS_API.md)
- [最佳实践指南](./BEST_PRACTICES.md)

---

**版本**: 1.0  
**更新日期**: 2026-01-31
