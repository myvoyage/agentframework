# 技能系统 (Skills System)

> **最后更新**: 2025-02-15
> **维护者**: AgentFramework Team
> **版本**: v2.0.0

---

## 📋 目录

- [系统概述](#系统概述)
- [核心特性](#核心特性)
- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [内置技能](#内置技能)
- [核心组件](#核心组件)
  - [技能接口](#技能接口-skill)
  - [技能注册表](#技能注册表-skillregistry)
  - [技能加载器](#技能加载器-skillloader)
  - [技能池](#技能池-skillspool)
  - [技能缓存](#技能缓存-skillscache)
  - [增强执行器](#增强执行器-enhancedexecutor)
  - [技能导入器](#技能导入器-skillimporter)
- [自定义技能](#自定义技能)
- [最佳实践](#最佳实践)
- [监控和调试](#监控和调试)
- [故障排除](#故障排除)
- [示例项目](#示例项目)
- [相关文档](#相关文档)

---

## 系统概述

技能系统是 AgentFramework 的**核心扩展机制**，允许代理通过可插拔的技能模块执行各种操作。每个技能是一个独立的、可复用的功能单元，遵循标准的接口定义。

### 设计理念

- **简单至上**: 零编程门槛，支持 Markdown + YAML Frontmatter 定义
- **类型安全**: 完整的类型定义和接口约束
- **高性能**: 内置连接池、缓存和并发优化
- **可观测性**: 内置监控和性能指标
- **安全性**: 沙箱环境执行，权限控制和资源限制

### 架构位置

```
agent/
├── skills/
│   ├── base_skill.go          # 基础技能实现
│   ├── registry.go             # 技能注册表 (核心)
│   ├── loader.go               # 技能加载器
│   ├── pool.go                # 技能池管理
│   ├── cache.go               # 技能缓存
│   ├── enhanced_executor.go    # 增强执行器
│   ├── skill_importer.go      # 技能导入器
│   ├── types.go               # 类型定义
│   ├── examples.go            # 使用示例
│   ├── templates.go           # 技能模板
│   ├── markdown/              # Markdown 技能支持
│   │   └── markdown_skill.go
│   ├── http_request.go        # HTTP 请求技能
│   ├── file_operation.go      # 文件操作技能
│   ├── code_execution.go       # 代码执行技能
│   ├── data_processing.go     # 数据处理技能
│   ├── skills_test.go         # 单元测试
│   └── skills_benchmark_test.go  # 性能测试
```

---

## 核心特性

### 🔥 热插拔

运行时动态加载和卸载技能，无需重启应用。

```go
// 注册新技能
registry.Register(myCustomSkill)

// 卸载技能
registry.Unregister("my_custom_skill")

// 启用/禁用技能
registry.EnableSkill("my_custom_skill")
registry.DisableSkill("my_custom_skill")
```

### ⚡ 连接池

内置技能实例池，提高并发性能。

```go
pool := skills.NewSkillsPool(config{
    MaxSize:   10,    // 最大池大小
    MinIdle:   2,     // 最小空闲数
    MaxIdle:   5,     // 最大空闲数
    Timeout:   30 * time.Second,
})
```

### 🚀 缓存机制

智能缓存减少重复计算，支持多种淘汰策略。

```go
cache := skills.NewSkillsCache(config{
    MaxSize:    1000,  // 最大缓存条目
    TTL:         5 * time.Minute,
    EvictPolicy: "lru", // LRU 淘汰策略
})
```

### 🔐 类型安全

完整的类型定义和接口约束，编译时类型检查。

```go
type Skill interface {
    Info(ctx context.Context) (*schema.ToolInfo, error)
    Invoke(ctx context.Context, input string) (string, error)
    IsEnabled(ctx context.Context) bool
    SetEnabled(enabled bool)
    GetMetadata(ctx context.Context) SkillMetadata
    SetMetadata(metadata SkillMetadata)
}
```

### 📊 可观测性

内置监控和性能指标。

```go
// 获取技能统计
stats := pool.GetStats("my_skill")
fmt.Printf("调用次数: %d\n", stats.CallCount)
fmt.Printf("平均耗时: %v\n", stats.AvgDuration)
fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate*100)
```

---

## 架构设计

### 技能生命周期

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Loader    │ -> │  Registry   │ -> │    Pool     │
│  加载技能    │    │  注册技能    │    │  管理实例    │
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────┐
│                    Execution                     │
│  ┌─────────────┐    ┌─────────────┐    ┌──────────┐ │
│  │   Cache     │ -> │  Executor   │ -> │  Result  │ │
│  │  缓存结果    │    │  执行技能    │    │  返回结果  │ │
│  └─────────────┘    └─────────────┘    └──────────┘ │
└─────────────────────────────────────────────────────┘
```

### 核心组件交互

```go
// 1. 加载技能
loader := skills.NewSkillLoader()
loadedSkills, err := loader.LoadFromDirectory("./skills")

// 2. 注册到注册表
registry := skills.NewSkillRegistry(config)
for _, skill := range loadedSkills {
    if err := registry.Register(skill); err != nil {
        log.Printf("注册技能失败: %v", err)
    }
}

// 3. 创建技能池
pool := skills.NewSkillsPool(config)

// 4. 执行技能
skill, _ := registry.GetByID("my_skill")
result, err := pool.Execute(ctx, skill, input)
```

---

## 快速开始

### 创建简单技能

```go
package myskills

import (
    "context"
    "github.com/cloudwego/eino/schema"
    "myvoyage/agentframework/agent/skills"
)

// MyCustomSkill 自定义技能
type MyCustomSkill struct {
    *skills.BaseSkill
    config map[string]interface{}
}

// NewMyCustomSkill 创建技能实例
func NewMyCustomSkill() *MyCustomSkill {
    return &MyCustomSkill{
        BaseSkill: skills.NewBaseSkill(skills.SkillMetadata{
            Name:        "my_custom_skill",
            Version:      "1.0.0",
            Description:  "我的自定义技能",
            Author:       "Your Name",
            Category:      "custom",
        }),
        config: make(map[string]interface{}),
    }
}

// Invoke 执行技能
func (s *MyCustomSkill) Invoke(ctx context.Context, input string) (string, error) {
    // 解析输入
    var params map[string]interface{}
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }

    // 执行操作
    result := s.doSomething(params)

    // 返回结果
    return json.Marshal(result)
}

// Info 返回技能信息
func (s *MyCustomSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_custom_skill",
        Desc: "技能描述",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "param1": {
                Type:     schema.TypeString,
                Desc:     "参数1说明",
                Required: true,
            },
        }),
    }, nil
}

func (s *MyCustomSkill) doSomething(params map[string]interface{}) interface{} {
    // 实现你的逻辑
    return map[string]interface{}{
        "success": true,
        "data":    "result",
    }
}
```

### 使用技能

```go
package main

import (
    "context"
    "log"
    "myvoyage/agentframework/agent/skills"
    "myvoyage/agentframework/myskills"
)

func main() {
    ctx := context.Background()

    // 创建注册表
    registry := skills.NewSkillRegistry(nil)

    // 注册自定义技能
    mySkill := myskills.NewMyCustomSkill()
    if err := registry.Register(mySkill); err != nil {
        log.Fatal(err)
    }

    // 使用技能
    skill, err := registry.GetByID("my_custom_skill")
    if err != nil {
        log.Fatal(err)
    }

    result, err := skill.Invoke(ctx, `{"param1": "value"}`)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("结果: %s", result)
}
```

---

## 内置技能

### 1. HTTP 请求技能

**文件**: `http_request.go`

发送 HTTP/HTTPS 请求，支持各种方法和自定义头。

```go
skill := http_request.NewHTTPSkill()
result, err := skill.Invoke(ctx, `{
    "url": "https://api.example.com/data",
    "method": "GET",
    "headers": {
        "Authorization": "Bearer token"
    }
}`)
```

**支持的操作**:
- ✅ GET、POST、PUT、DELETE、PATCH
- ✅ 自定义请求头
- ✅ 请求体
- ✅ 超时设置
- ✅ TLS 配置

### 2. 文件操作技能

**文件**: `file_operation.go`

安全的文件系统访问和操作。

```go
skill := file_operation.NewFileOperationSkill()
result, err := skill.Invoke(ctx, `{
    "action": "read",
    "path": "/path/to/file.txt"
}`)
```

**支持的操作**:
- ✅ 读取文件
- ✅ 写入文件
- ✅ 列出目录
- ✅ 删除文件
- ✅ 文件信息查询
- ✅ 权限检查

**安全特性**:
- 🔒 路径验证
- 🔒 权限控制
- 🔒 沙箱隔离

### 3. 代码执行技能

**文件**: `code_execution.go`

安全的代码执行环境。

```go
skill := code_execution.NewCodeExecutionSkill()
result, err := skill.Invoke(ctx, `{
    "language": "python",
    "code": "print('Hello, World!')"
}`)
```

**支持的语言**:
- ✅ Python (沙箱执行)
- ✅ JavaScript (Node.js 沙箱)
- ✅ Go (yaegi 解释器，428倍加速)
- ✅ Bash (容器隔离)

**安全特性**:
- 🔒 容器隔离
- 🔒 资源限制
- 🔒 超时控制
- 🔒 增强的代码分析

### 4. 数据处理技能

**文件**: `data_processing.go`

常见数据处理和转换操作。

```go
skill := data_processing.NewDataProcessingSkill()
result, err := skill.Invoke(ctx, `{
    "operation": "json_parse",
    "data": "{\"key\": \"value\"}"
}`)
```

**支持的操作**:
- ✅ JSON 解析/序列化
- ✅ Base64 编码/解码
- ✅ 数据格式转换
- ✅ 正则表达式匹配

---

## 核心组件

### 技能接口 (Skill)

所有技能必须实现的核心接口。

```go
type Skill interface {
    // Info 返回技能的元信息
    Info(ctx context.Context) (*schema.ToolInfo, error)

    // Invoke 执行技能，处理输入并返回输出
    Invoke(ctx context.Context, input string) (string, error)

    // IsEnabled 检查技能是否启用
    IsEnabled(ctx context.Context) bool

    // SetEnabled 设置技能是否启用
    SetEnabled(enabled bool)

    // GetMetadata 获取技能的元数据
    GetMetadata(ctx context.Context) SkillMetadata

    // SetMetadata 设置技能的元数据
    SetMetadata(metadata SkillMetadata)
}
```

### 技能注册表 (SkillRegistry)

**文件**: [registry.go](../agent/skills/registry.go)

所有已注册技能的中央管理器。

**核心功能**:

```go
// 创建注册表
registry := skills.NewSkillRegistry(&skills.RegistryConfig{
    BaseDir:     ".skills/registry",
    AutoSave:     true,
    EnableIndex:   true,
})

// 注册技能（自动去重）
err := registry.Register(skillEntry)

// 更新技能
err := registry.Update(skillEntry)

// 注销技能
err := registry.Unregister("skill_id")

// 查询技能
skills, err := registry.Find(&skills.SkillQuery{
    Name:          "搜索名称",
    Category:       "分类",
    Tags:          []string{"标签1", "标签2"},
    UsedCountMin:   10,
})

// 按分类列出
skills := registry.ListByCategory("category_name")

// 按标签列出
skills := registry.ListByTag("tag_name")

// 导出
registry.ExportToJSON(writer)
registry.ExportToMarkdown(writer)
```

**数据结构**:

```go
type SkillEntry struct {
    // 基本信息
    ID          string
    Name        string
    Description string
    Category    string
    Tags        []string
    Version     string
    Enabled     bool

    // 统计信息
    CreatedAt  time.Time
    UpdatedAt  time.Time
    UsedCount  int64
    LastUsed   time.Time
    LastUsedBy string

    // 参数定义
    InputSchema  *Schema
    OutputSchema *Schema

    // 元数据
    Config   map[string]interface{}
    Metadata map[string]interface{}
}
```

### 技能加载器 (SkillLoader)

**文件**: [loader.go](../agent/skills/loader.go)

从配置或目录加载技能。

```go
loader := skills.NewSkillLoader()

// 从目录加载
skills, err := loader.LoadFromDirectory("./skills")

// 从配置文件加载
skills, err := loader.LoadFromConfig("./skills.yaml")

// 验证技能
err := loader.Validate(skill)
```

**配置文件格式** (YAML):

```yaml
skills:
  - name: http_request
    enabled: true
    config:
      timeout: 30s
      max_retries: 3

  - name: file_operation
    enabled: true
    config:
      allowed_paths:
        - /tmp
        - /home/user/work
```

### 技能池 (SkillsPool)

**文件**: [pool.go](../agent/skills/pool.go)

实例池管理，提高并发性能。

```go
pool := skills.NewSkillsPool(config{
    MaxSize:   10,     // 最大池大小
    MinIdle:   2,      // 最小空闲数
    MaxIdle:   5,      // 最大空闲数
    Timeout:   30 * time.Second,
})

// 获取技能实例
skill, err := pool.Get("http_request")
defer pool.Put(skill)

// 执行
result, err := skill.Invoke(ctx, input)
```

### 技能缓存 (SkillsCache)

**文件**: [cache.go](../agent/skills/cache.go)

智能缓存减少重复计算。

```go
cache := skills.NewSkillsCache(config{
    MaxSize:    1000,   // 最大缓存条目
    TTL:         5 * time.Minute,
    EvictPolicy: "lru", // LRU 淘汰策略
})

// 带缓存的执行
result, err := cache.GetOrCompute("cache_key", func() (interface{}, error) {
    return skill.Invoke(ctx, input)
})
```

### 增强执行器 (EnhancedExecutor)

**文件**: [enhanced_executor.go](../agent/skills/enhanced_executor.go)

提供高级执行功能。

```go
executor := skills.NewEnhancedExecutor(config{
    MaxConcurrency: 10,
    Timeout:        30 * time.Second,
    RetryPolicy:    skills.RetryPolicy{
        MaxRetries: 3,
        BackoffBase: time.Second,
    },
})

// 执行技能
result, err := executor.Execute(ctx, skill, input)

// 批量执行
results := executor.ExecuteBatch(ctx, []SkillExecution{
    {Skill: skill1, Input: input1},
    {Skill: skill2, Input: input2},
})
```

### 技能导入器 (SkillImporter)

**文件**: [skill_importer.go](../agent/skills/skill_importer.go)

从外部源导入技能。

```go
importer := skills.NewSkillImporter()

// 从 Git 仓库导入
skills, err := importer.FromGit("https://github.com/user/skills.git")

// 从本地目录导入
skills, err := importer.FromDirectory("./custom_skills")

// 验证导入的技能
for _, skill := range skills {
    if err := importer.Validate(skill); err != nil {
        log.Printf("技能 %s 验证失败: %v", skill.Info().Name, err)
    }
}
```

---

## 自定义技能

### Markdown 技能 (零编程)

使用 Markdown + YAML Frontmatter 定义技能，无需编程。

**文件**: `my_skill.md`

```markdown
---
id: "my_custom_skill"
name: "我的自定义技能"
version: "1.0.0"
category: "custom"
tags: ["工具", "实用"]
description: "这是一个自定义技能示例"
enabled: true
---

# 技能说明

这个技能演示如何使用 Markdown 定义技能。

## 参数

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| input | string | 是 | 输入文本 |

## 使用示例

```bash
# 调用技能
echo '{"input": "hello"} | my_custom_skill
```

## 实现代码

```javascript
function execute(params) {
    const { input } = params;
    return {
        output: input.toUpperCase()
    };
}
```
```

### Go 技能 (高性能)

使用 Go 实现技能，获得最佳性能。

```go
package myskills

import (
    "context"
    "encoding/json"
    "github.com/cloudwego/eino/schema"
    "myvoyage/agentframework/agent/skills"
)

type MyGoSkill struct {
    *skills.BaseSkill
}

func NewMyGoSkill() *MyGoSkill {
    return &MyGoSkill{
        BaseSkill: skills.NewBaseSkill(skills.SkillMetadata{
            Name:       "my_go_skill",
            Version:     "1.0.0",
            Description: "高性能 Go 技能",
            Author:      "Your Name",
            Category:    "performance",
        }),
    }
}

func (s *MyGoSkill) Invoke(ctx context.Context, input string) (string, error) {
    var params map[string]interface{}
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }

    result := s.execute(params)
    return json.Marshal(result)
}

func (s *MyGoSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_go_skill",
        Desc: "高性能 Go 技能",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "input": {
                Type:     schema.TypeString,
                Desc:     "输入参数",
                Required: true,
            },
        }),
    }, nil
}

func (s *MyGoSkill) execute(params map[string]interface{}) interface{} {
    // 实现你的逻辑
    return map[string]interface{}{
        "success": true,
        "output":   params["input"],
    }
}
```

---

## 最佳实践

### 1. 错误处理

```go
func (s *MySkill) Invoke(ctx context.Context, input string) (string, error) {
    // 验证输入
    if err := s.validateInput(input); err != nil {
        return "", fmt.Errorf("输入验证失败: %w", err)
    }

    // 执行操作
    result, err := s.execute(input)
    if err != nil {
        return "", fmt.Errorf("执行失败: %w", err)
    }

    return result, nil
}
```

### 2. 超时控制

```go
func (s *MySkill) Invoke(ctx context.Context, input string) (string, error) {
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 使用带超时的上下文
    return s.executeWithContext(ctx, input)
}
```

### 3. 资源清理

```go
func (s *MySkill) Close() error {
    // 清理资源
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}
```

### 4. 日志记录

```go
import "log/slog"

func (s *MySkill) Invoke(ctx context.Context, input string) (string, error) {
    logger := slog.FromContext(ctx)
    logger.Info("执行技能", "skill", s.Info().Name)

    // ... 执行逻辑

    return result, nil
}
```

---

## 监控和调试

### 性能指标

```go
// 获取技能统计
stats := pool.GetStats("my_skill")
fmt.Printf("调用次数: %d\n", stats.CallCount)
fmt.Printf("平均耗时: %v\n", stats.AvgDuration)
fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate*100)
fmt.Printf("最后使用: %v\n", stats.LastUsed)
```

### OpenTelemetry 集成

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (s *MySkill) Invoke(ctx context.Context, input string) (string, error) {
    ctx, span := otel.Tracer("skills").Start(ctx, "MySkill.Invoke")
    defer span.End()

    // 执行逻辑
    result, err := s.execute(ctx, input)

    return result, err
}
```

---

## 故障排除

### 技能未找到

```go
skill, err := registry.GetByID("my_skill")
if err != nil {
    // 检查技能是否已注册
    if errors.Is(err, skills.ErrSkillNotFound) {
        log.Println("技能未注册")
    }
}
```

### 超时问题

```go
// 增加超时时间
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
defer cancel()
```

### 内存问题

```go
// 限制缓存大小
cache := skills.NewSkillsCache(config{
    MaxSize: 100, // 减小缓存
})
```

---

## 示例项目

- 📘 基础技能使用：[agent/skills/examples.go](../agent/skills/examples.go)
- 🧪 技能测试：[agent/skills/skills_test.go](../agent/skills/skills_test.go)
- ⚡ 性能测试：[agent/skills/skills_benchmark_test.go](../agent/skills/skills_benchmark_test.go)

---

## 相关文档

- 📘 [技能 API 快速参考](../../docs/api/skills/SKILLS_API_QUICK_REFERENCE.md)
- 📘 [技能系统完整文档](../../docs/components/SKILLS.md)
- 📘 [技能增强指南](../../docs/guides/tutorials/CUSTOM_SKILLS.md)
- 📘 [最佳实践指南](../../docs/guides/best-practices/SKILLS.md)
- 📘 [故障排查指南](../../docs/operation/TROUBLESHOOTING.md#skills)

---

## 贡献指南

欢迎贡献新的技能！请参考：

- 📘 [贡献指南](../../docs/development/CONTRIBUTING.md)
- 📘 [技能模板](../../templates/skill_template.go)
- 📘 [代码规范](../../docs/development/CODING_STANDARDS.md)

---

**Made with ❤️ by AgentFramework Team**
