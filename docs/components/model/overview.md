# Model 管理系统概览

> **AgentFramework Model 组件文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [核心功能](#核心功能)
- [模型抽象](#模型抽象)
- [工厂模式](#工厂模式)
- [快速开始](#快速开始)
- [配置方式](#配置方式)
- [高级特性](#高级特性)

---

## 系统简介

Model 管理系统是 AgentFramework 的核心组件之一，负责管理多种 AI 模型的统一接口和动态切换。

### 核心职责

| 职责 | 说明 |
|------|------|
| **模型管理** | 统一管理所有 LLM 后端 |
| **动态切换** | 运行时模型切换和健康检查 |
| **缓存优化** | 模型缓存和预热，减少冷启动延迟 |
| **配置验证** | 详细的参数验证和默认值设置 |
| **错误处理** | 详细的错误信息和调试支持 |

### 技术特点

- ✅ **高性能**: 基于 Go 的高并发处理能力
- ✅ **可扩展**: 插件化架构，易于添加新模型
- ✅ **类型安全**: 编译时类型检查，避免运行时错误
- ✅ **智能缓存**: 分片缓存提高并发性能
- ✅ **优雅降级**: 模型不可用时自动降级到备用模型

---

## 核心功能

### 模型抽象

```go
// 模型接口定义
type ChatModel interface {
    // 基本信息
    Name() string
    Type() string
    Version() string

    // 配置
    GetConfig() *ModelConfig
    SetConfig(cfg *ModelConfig) error

    // 执行接口
    Generate(ctx context.Context, request *schema.Message, opts ...model.Option) (*schema.Message, error)
    Stream(ctx context.Context, request *schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

// 模型配置
type ModelConfig struct {
    Type         string            // 模型类型 (openai, ollama, lmstudio, vllm)
    Model        string            // 模型名称
    BaseURL     string            // API 基础 URL
    APIKey      string            // API 密钥
    Enabled      bool              // 是否启用

    // 高级配置
    Timeout      time.Duration      // 请求超时
    MaxRetries   int             // 最大重试次数
    Temperature  float64         // 生成温度
    MaxTokens    int             // 最大 Token 数
    TopP        float64          // Top-P 采样
    TopK        int             // Top-K 采样
}
```

### 支持的模型

| 模型类型 | 说明 | 配置示例 |
|---------|------|---------|
| **OpenAI** | GPT-4、GPT-3.5 等 | `type: "openai", model: "gpt-4"` |
| **Ollama** | 本地开源模型 | `type: "ollama", model: "llama3"` |
| **LM Studio** | 本地模型工作室 | `type: "lmstudio", model: "my-model"` |
| **vLLM** | vLLM 推理框架 | `type: "vllm", model: "vicuna-7b"` |

---

## 工厂模式

Model 管理系统使用工厂模式创建模型实例：

```go
// 模型工厂函数签名
type ModelFactory func(ctx context.Context, modelName string) (ChatModel, error)

// 默认模型工厂实现
func NewDefaultModelFactory(cfg DefaultModelFactoryConfig) ModelFactory {
    // 预处理配置
    preprocessed := &preprocessedModelConfig{
        configs: make(map[string]ModelConfig, len(cfg.Models)),
    }

    // 返回工厂函数
    return func(ctx context.Context, modelName string) (ChatModel, error) {
        // 查找配置
        modelCfg, ok := preprocessed.configs[modelName]
        if !ok {
            return nil, fmt.Errorf("model %s not found", modelName)
        }

        // 创建模型
        return createModel(ctx, modelCfg)
    }
}
}

// 模型创建函数
func createModel(ctx context.Context, cfg ModelConfig) (ChatModel, error) {
    switch cfg.Type {
    case "openai":
        return newOpenAIModel(ctx, cfg)
    case "ollama":
        return newOllamaModel(ctx, cfg)
    case "lmstudio":
        return newLMStudioModel(ctx, cfg)
    case "vllm":
        return newVLLMModel(ctx, cfg)
    default:
        return nil, fmt.Errorf("unknown model type: %s", cfg.Type)
    }
}
```

### 工厂优势

- ✅ **统一接口**: 所有模型使用相同接口
- ✅ **配置驱动**: 通过配置文件动态管理模型
- ✅ **类型安全**: 编译时类型检查
- ✅ **易于扩展**: 添加新模型只需实现接口

---

## 快速开始

### 创建模型工厂

```go
package main

import (
    "context"
    "fmt"
    "log"

    "agentframework/model"
    "agentframework/config"
)

func main() {
    ctx := context.Background()

    // 创建模型配置
    cfg := &config.DefaultModelFactoryConfig{
        Models: map[string]config.ModelConfig{
            "gpt-4": {
                Type:    "openai",
                Model:   "gpt-4-turbo",
                BaseURL: "https://api.openai.com/v1",
                APIKey: os.Getenv("OPENAI_API_KEY"),
            },
            "ollama-llama3": {
                Type:    "ollama",
                Model:   "llama3",
                BaseURL: "http://localhost:11434",
            },
        },
    }

    // 创建模型工厂
    factory := model.NewDefaultModelFactory(cfg)

    // 获取模型
    chatModel, err := factory(ctx, "gpt-4")
    if err != nil {
        log.Fatal(err)
    }

    // 使用模型
    response, err := chatModel.Generate(ctx, &schema.Message{
        Role:    "user",
        Content: "你好，请介绍一下你自己",
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Response:", response.Content)
}
```

---

## 配置方式

### YAML 配置文件

```yaml
models:
  # OpenAI GPT-4
  gpt-4:
    type: "openai"
    model: "gpt-4-turbo"
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    enabled: true
    timeout: 30
    max_retries: 3
    temperature: 0.7
    max_tokens: 2000

  # Ollama Llama3
  ollama-llama3:
    type: "ollama"
    model: "llama3"
    base_url: "http://localhost:11434"
    enabled: true
    timeout: 120
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|---------|------|---------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | - |
| `OPENAI_BASE_URL` | OpenAI 基础 URL | `https://api.openai.com/v1` |
| `OLLAMA_BASE_URL` | Ollama 基础 URL | `http://localhost:11434` |

---

## 高级特性

### 动态模型切换

```go
// 在运行时切换模型
func (a *Agent) SetModel(modelName string) error {
    a.model = modelName
    return nil
}

// 检查模型是否可用
func (a *Agent) GetModel() string {
    return a.model
}
```

### 智能缓存

```go
// 分片缓存配置
type CacheConfig struct {
    Enabled    bool              // 是否启用
    MaxSize    int               // 最大缓存数
    TTL        time.Duration        // 缓存时间
    Strategy   string           // 缓存策略
}

// 使用缓存
cachedModel, err := model.GetCached(ctx, "gpt-4")
if err == nil {
    // 使用缓存的模型
}
```

### 配置验证

```go
// 配置验证函数
func (cfg *ModelConfig) Validate() error {
    // 验证必填字段
    if cfg.Type == "" {
        return fmt.Errorf("model type is required")
    }
    if cfg.Model == "" {
        return fmt.Errorf("model name is required")
    }
    if cfg.BaseURL == "" {
        return fmt.Errorf("base_url is required")
    }

    // 验证 URL 格式
    if _, err := url.ParseRequestURL(cfg.BaseURL); err != nil {
        return fmt.Errorf("invalid base_url: %w", err)
    }

    return nil
}
```

---

## 相关文档

- 📘 [Agent 概览](../agent/overview.md) - Agent 系统概览
- 📘 [Workflow 概览](../workflow/overview.md) - Workflow 系统概览
- 📘 [配置指南](../../configuration/CONFIGURATION.md) - 详细配置说明
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
