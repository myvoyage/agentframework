# Skills 系统概览

> **AgentFramework Skills 组件文档**
> **版本**: v2.0.0
> **最后更新**: 2026-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [核心概念](#核心概念)
- [技能类型](#技能类型)
- [技能架构](#技能架构)
- [SkillAgent 增强](#skillagent-增强)
- [技能注册表](#技能注册表)
- [快速开始](#快速开始)
- [核心特性](#核心特性)

---

## 系统简介

Skills 系统是 AgentFramework 的能力扩展机制，允许 Agent 通过调用预定义的功能来扩展其能力。

### 技术特点

- ✅ **零代码技能**: 支持 Markdown + YAML Frontmatter 定义
- ✅ **动态加载**: 运行时动态加载技能
- ✅ **参数验证**: JSON Schema 自动验证
- ✅ **结果缓存**: 智能缓存提高性能
- ✅ **MCP 集成**: 44+ MCP 工具原生支持
- ✅ **安全沙箱**: 技能在安全沙箱中执行

### MCP 工具集成

| 类别 | 工具数量 | 功能 |
|------|----------|------|
| 浏览器自动化 | 9 | navigate, click, input, screenshot, PDF |
| 文件操作 | 7 | read, write, copy, move, delete |
| 身份认证 | 7 | JWT, API Key 管理 |
| 可视化 | 9 | 图像处理、图表生成 |
| Shell 执行 | 6 | 命令执行、进程管理 |
| 网络工具 | 6 | HTTP 请求、端口扫描 |
| 代码执行 | 4 | 多语言沙箱执行 |
| 语音 | 2 | TTS/STT |

---

## 核心概念

### Skill 定义

Skill 是一个可执行的功能单元，具有以下特征：

| 特征 | 说明 |
|------|------|
| **唯一标识** | 每个技能有唯一的名称 |
| **输入参数** | 定义输入参数的 Schema |
| **输出结果** | 定义输出结果的格式 |
| **执行逻辑** | 技能的具体实现代码 |
| **元数据** | 描述、分类、标签等信息 |

### Skill vs Tool

| 特性 | Skill | Tool |
|------|-------|------|
| **定义方式** | YAML/代码 | 代码实现 |
| **复杂度** | 可简单（Markdown） | 通常较复杂 |
| **调用方式** | 通过技能系统 | 通过工具系统 |
| **适用场景** | 业务逻辑、数据处理 | 系统操作、API 调用 |
| **参数验证** | JSON Schema | 接口定义 |

### 技能生命周期

```
┌──────────────────────────────────────────────────────────────┐
│                Skill 生命周期                           │
└──────────────────────────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  1. 定义     │
          │ (YAML/代码) │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  2. 注册     │
          │ SkillRegistry │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  3. 加载     │
          │ SkillLoader   │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  4. 验证     │
          │ 参数验证      │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  5. 执行     │
          │ 调用执行     │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  6. 缓存     │
          │ 结果缓存      │
          └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  7. 卸载     │
          │ (可选)       │
          └─────────────┘
```

---

## 技能类型

### 1. HTTP 技能

**功能**: 执行 HTTP 请求

**配置**:

```yaml
skills:
  - name: "http_request"
    type: "http"
    enabled: true

    # 配置
    config:
      timeout: 30
      max_retries: 3
      allowed_hosts:
        - "api.example.com"
        - "*.github.com"

    # 参数定义
    input_schema:
      type: "object"
      properties:
        url:
          type: "string"
          description: "请求 URL"
        method:
          type: "string"
          enum: ["GET", "POST", "PUT", "DELETE"]
          default: "GET"
        headers:
          type: "object"
          description: "请求头"
        body:
          type: "object"
          description: "请求体"
      required: ["url"]

    # 输出定义
    output_schema:
      type: "object"
      properties:
        status:
          type: "integer"
        data:
          type: "object"
```

**使用场景**:
- API 集成
- 数据获取
- Webhook 调用

---

### 2. 文件操作技能

**功能**: 安全的文件系统访问

**配置**:

```yaml
skills:
  - name: "file_operation"
    type: "file"
    enabled: true

    config:
      # 安全限制
      allowed_paths:
        - "/tmp"
        - "/home/user/work"
      max_file_size: 10485760  # 10MB
      read_only: false

    input_schema:
      type: "object"
      properties:
        operation:
          type: "string"
          enum: ["read", "write", "delete", "list"]
        path:
          type: "string"
          description: "文件路径"
        content:
          type: "string"
          description: "写入内容"
      required: ["operation", "path"]
```

**使用场景**:
- 日志读取
- 配置文件操作
- 数据文件处理

---

### 3. 代码执行技能

**功能**: 安全的代码执行

**配置**:

```yaml
skills:
  - name: "code_execution"
    type: "code"
    enabled: true

    config:
      # 超时和资源限制
      timeout: 60
      memory_limit: "512m"
      cpu_limit: "1.0"

      # 允许的操作
      allowed_operations:
        - "read"
        - "write"
        # 不允许网络、进程等危险操作

    input_schema:
      type: "object"
      properties:
        language:
          type: "string"
          enum: ["go", "python", "javascript", "bash"]
        code:
          type: "string"
          description: "执行的代码"
      required: ["language", "code"]
```

**使用场景**:
- 数据转换
- 算法执行
- 动态计算

---

### 4. 数据处理技能

**功能**: 数据转换和处理

**配置**:

```yaml
skills:
  - name: "data_processing"
    type: "data"
    enabled: true

    config:
      max_data_size: 1048576  # 1MB

    input_schema:
      type: "object"
      properties:
        operation:
          type: "string"
          enum: ["filter", "transform", "aggregate", "sort"]
        data:
          type: "array"
          description: "数据数组"
        params:
          type: "object"
          description: "操作参数"
      required: ["operation", "data"]
```

**使用场景**:
- 数据过滤
- 格式转换
- 数据聚合

---

### 5. Markdown 技能

**功能**: 零代码定义技能

**文件**: `SKILL.md`

```markdown
---
name: "weather_query"
version: "1.0.0"
category: "api"
tags: ["weather", "api"]
enabled: true

description: |
  查询天气信息的技能

input_schema: |
  type: "object"
  properties:
    city:
      type: "string"
      description: "城市名称"
    unit:
      type: "string"
      enum: ["celsius", "fahrenheit"]
      default: "celsius"
  required: ["city"]

output_schema: |
  type: "object"
  properties:
    temperature:
      type: "number"
      description: "当前温度"
    condition:
      type: "string"
      description: "天气状况"
    humidity:
      type: "number"
      description: "湿度百分比"

execution: |
  # 这里可以编写执行逻辑
  # 或者使用模板调用 API
  api_call:
    url: "https://api.weather.com/current"
    method: "GET"
    params:
      city: "{{input.city}}"
      unit: "{{input.unit}}"
```

**使用场景**:
- 快速原型开发
- 业务逻辑定义
- API 集成

---

## 技能架构

### 系统架构

```
┌──────────────────────────────────────────────────────────────┐
│                  Skills 系统架构                         │
└──────────────────────────────────────────────────────────────┘

┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│ Skill       │          │  Skill      │          │   Skill     │
│ Registry   │────────▶│  Pool       │────────▶│   Cache     │
└─────────────┘          └─────────────┘          └─────────────┘
        │                        │                        │
        └────────────┬──────────────┴────────────┘
                     │
                     ▼
          ┌──────────────────────────┐
          │     Skill Library     │
          │  (技能库）           │
          └──────────────────────────┘
                     │
         ┌───────────┼───────────┬───────────┐
         ▼            ▼            ▼           ▼
┌─────────────┐┌─────────────┐┌─────────────┐┌─────────────┐
│ HTTP Skill ││ File Skill  ││ Code Skill  ││Data Skill  │
└─────────────┘└─────────────┘└─────────────┘└─────────────┘
```

### 核心组件

#### 1. SkillRegistry

**功能**: 技能注册和管理中心

**特性**:
- ✅ 自动去重注册
- ✅ 多条件查找
- ✅ 分类和标签管理
- ✅ 导出功能（JSON/Markdown）

#### 2. SkillLoader

**功能**: 技能加载器

**支持的来源**:
- `.skills/` 目录
- 数据库存储
- 远程服务器
- 运行时动态加载

#### 3. SkillsPool

**功能**: 技能连接池

**特性**:
- ✅ 连接复用
- ✅ 并发控制
- ✅ 健康检查
- ✅ 自动重连

#### 4. SkillsCache

**功能**: 技能执行结果缓存

**策略**:
- LRU (Least Recently Used)
- TTL (Time To Live)
- 基于参数的缓存键

---

## 技能注册表

### 注册表功能

**文件**: [agent/skills/registry.go](../../agent/skills/registry.go)

```go
type SkillRegistry struct {
    skills  map[string]*SkillEntry
    mu      sync.RWMutex
    indexes map[string][]string
    baseDir string
}

// 注册技能（自动去重）
func (sr *SkillRegistry) Register(skill Skill) error

// 更新技能
func (sr *SkillRegistry) Update(skill Skill) error

// 注销技能
func (sr *SkillRegistry) Unregister(name string) error

// 查找技能（多条件）
func (sr *SkillRegistry) Find(opts ...FindOption) ([]Skill, error)

// 按分类列出
func (sr *SkillRegistry) ListByCategory(category string) ([]Skill, error)

// 按标签列出
func (sr *SkillRegistry) ListByTag(tag string) ([]Skill, error)

// 导出 JSON
func (sr *SkillRegistry) ExportToJSON() ([]byte, error)

// 导出 Markdown
func (sr *SkillRegistry) ExportToMarkdown() ([]byte, error)
```

### 注册表配置

**目录结构**:

```
.skills/
├── registry/
│   └── registry.yaml       # 注册表配置
├── builtin/               # 内置技能
│   ├── http_request/
│   ├── file_operation/
│   └── ...
├── custom/                # 自定义技能
│   ├── my_skill/
│   └── another_skill/
└── cache/                 # 技能缓存
    └── ...
```

**registry.yaml**:

```yaml
# 技能注册表配置
version: "1.0"

# 自动扫描目录
auto_discover:
  enabled: true
  directories:
    - "builtin"
    - "custom"

# 缓存配置
cache:
  enabled: true
  backend: "memory"    # memory, redis, file
  ttl: 3600             # 1小时（秒）

# 验证规则
validation:
  strict_schema: true    # 严格验证 Schema
  require_tests: false   # 是否需要测试
```

---

## 快速开始

### 创建第一个技能

#### 方式1: Markdown 技能（推荐）

**文件**: `.skills/custom/weather_query/SKILL.md`

```markdown
---
name: "weather_query"
version: "1.0.0"
category: "api"
tags: ["weather", "api"]
enabled: true

description: |
  查询天气信息的技能

input_schema: |
  type: "object"
  properties:
    city:
      type: "string"
      description: "城市名称"
    unit:
      type: "string"
      enum: ["celsius", "fahrenheit"]
      default: "celsius"
  required: ["city"]

output_schema: |
  type: "object"
  properties:
    temperature:
      type: "number"
    condition:
      type: "string"
    humidity:
      type: "number"

execution: |
  # 调用天气 API
  url: "https://api.weather.com/current"
  method: "GET"
  params:
    city: "{{input.city}}"
    unit: "{{input.unit}}"
```

#### 方式2: Go 代码技能

**文件**: `.skills/custom/calculator/skill.go`

```go
package custom

import (
    "context"
    "fmt"
)

type CalculatorSkill struct {
    *skills.BaseSkill
}

func (s *CalculatorSkill) Invoke(ctx context.Context, input string) (string, error) {
    // 解析输入
    var args struct {
        A float64 `json:"a"`
        B float64 `json:"b"`
        Op string  `json:"op"`
    }

    if err := json.Unmarshal([]byte(input), &args); err != nil {
        return "", err
    }

    // 执行计算
    var result float64
    switch args.Op {
    case "add":
        result = args.A + args.B
    case "subtract":
        result = args.A - args.B
    case "multiply":
        result = args.A * args.B
    case "divide":
        if args.B == 0 {
            return "", fmt.Errorf("division by zero")
        }
        result = args.A / args.B
    default:
        return "", fmt.Errorf("unknown operation: %s", args.Op)
    }

    return fmt.Sprintf(`{"result": %f}`, result), nil
}

func (s *CalculatorSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "calculator",
        Desc: "执行基本数学计算",
    }, nil
}
```

### 使用技能

```go
// 获取技能
skill, err := skillRegistry.Find(
    skills.WithName("weather_query"),
)
if err != nil {
    log.Fatal(err)
}

// 执行技能
input := `{"city": "Beijing", "unit": "celsius"}`
result, err := skill.Invoke(ctx, input)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

---

## 核心特性

### 1. 参数验证

**自动验证**:

```go
// 技能定义了 input_schema
input := `{"city": "Beijing"}`

// 系统自动验证
err := skill.ValidateInput(input)
if err != nil {
    // 验证失败
    return err
}

// 验证成功，可以执行
result, err := skill.Invoke(ctx, input)
```

**验证规则**:
- 类型检查（string, number, boolean 等）
- 必填字段检查
- 枚举值验证
- 格式验证（email, URI 等）
- 自定义验证规则

### 2. 结果缓存

**缓存策略**:

```go
// 启用缓存
skill := skills.NewSkill(
    skills.WithName("weather_query"),
    skills.WithCacheConfig(&skills.CacheConfig{
        Enabled:  true,
        TTL:      10 * time.Minute,
        Strategy: skills.CacheStrategyLRU,
    }),
)

// 第一次调用 - 执行技能
result1, err1 := skill.Invoke(ctx, input)

// 第二次调用（相同参数）- 返回缓存
result2, err2 := skill.Invoke(ctx, input)
// result2 来自缓存，不会重新执行
```

**缓存键生成**:
```go
// 默认缓存键（基于所有输入参数）
cacheKey := sha256.Sum([]byte(input))

// 自定义缓存键
skill.WithCacheKeyFunc(func(input string) string {
    var parsed struct {
        City string `json:"city"`
    }
    json.Unmarshal([]byte(input), &parsed)
    return fmt.Sprintf("weather:%s", parsed.City)
})
```

### 3. 安全控制

**权限配置**:

```yaml
skills:
  - name: "file_operation"
    enabled: true

    # 安全限制
    security:
      # 允许的路径
      allowed_paths:
        - "/tmp"
        - "/home/user/work"

      # 拒绝的路径
      denied_paths:
        - "/etc"
        - "/root"

      # 最大文件大小
      max_file_size: 10485760  # 10MB

      # 只读模式
      read_only: false
```

**权限验证**:
```go
// 系统自动验证权限
if err := skill.CheckPermission(ctx, "file_operation", "/etc/passwd"); err != nil {
    // 权限不足
    return fmt.Errorf("permission denied: %w", err)
}
```

### 4. 监控和日志

**执行监控**:

```go
// 记录执行开始
startTime := time.Now()
log.Infof("skill invoke: %s", skill.Name())

// 执行技能
result, err := skill.Invoke(ctx, input)

// 记录执行完成
duration := time.Since(startTime)
log.Infof("skill complete: %s, duration: %v, success: %v",
    skill.Name(), duration, err == nil)
```

**性能指标**:
- 执行次数
- 平均耗时
- 成功/失败率
- 缓存命中率

---

## 相关文档

- 📘 [内置技能](builtin.md) - 内置技能详解
- 📘 [自定义技能](custom.md) - 自定义技能开发
- 📘 [Skills API 参考](api.md) - 完整 API 文档
- 📘 [agent/skills/README.md](../../../agent/skills/README.md) - Skills 系统详细文档
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
