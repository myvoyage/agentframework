# 配置指南

> **AgentFramework 完整配置说明**
> **版本**: v2.0.0
> **最后更新**: 2026-02-15

---

## 📋 目录

- [配置概览](#配置概览)
- [Host 配置](#host-配置)
- [Agent 配置](#agent-配置)
- [Workflow 配置](#workflow-配置)
- [Model 配置](#model-配置)
- [Skill 配置](#skill-配置)
- [监控配置](#监控配置)
- [存储配置](#存储配置)
- [高级配置](#高级配置)
- [环境变量](#环境变量)

---

## 配置概览

### 配置文件层次

```
┌────────────────────────────────────────────────────┐
│              Configuration Hierarchy          │
│  ┌────────────┬  ┌────────────┬  ┌────────────┐ │
│  │  Global    │  │  Project    │  │  Runtime    │ │
│  │ Config    │  │  Config     │  │  Config     │ │
│  └────────────┘  └────────────┘  └────────────┘ │
└────────────────────────────────────────────────────┘
       │                │                │
       ▼                ▼                ▼
┌────────────────────────────────────────────────────┐
│              host.yaml (Merged)              │
└────────────────────────────────────────────────────┘
```

### 项目规模

| 指标 | 数值 |
|------|------|
| 配置项 | **100+** |
| 存储后端 | **4 种** |
| 模型支持 | **4 种** |

### 配置文件格式

AgentFramework 支持多种配置格式：

| 格式 | 文件扩展名 | 说明 |
|------|-----------|------|
| **YAML** | `.yaml`, `.yml` | 推荐格式，人类可读 |
| **JSON** | `.json` | 机器可读，适合自动化 |
| **TOML** | `.toml` | 类似 INI 格式 |
| **Environment** | `.env` | 环境变量 |

### 配置优先级

```
Environment Variables (最高)
        │
        ▼
Command Line Flags
        │
        ▼
Config Files (host.yaml, config.yaml)
        │
        ▼
Default Values (最低)
```

---

## Host 配置

### 基本配置

**文件**: `host.yaml`

```yaml
# 基本信息
name: "my-agent-app"
version: "1.0.0"
description: "我的第一个 Agent 应用"

# 默认模型
default_model: "ollama-llama3"

# 日志配置
logging:
  level: "info"          # debug, info, warn, error
  format: "json"        # json, text
  output: "stdout"       # stdout, stderr, file path
```

### HostConfig 结构

```go
type HostConfig struct {
    // 基本信息
    Name         string
    Version      string
    Description  string

    // 模型配置
    DefaultModel string
    Models       map[string]ModelConfig

    // 日志配置
    Logging      *LoggingConfig

    // 内存配置
    Memory       *MemoryConfig

    // 监控配置
    Monitoring   *MonitoringConfig

    // 异步任务
    AsyncTask   *AsyncTaskConfig

    // 调度器
    Scheduler   *SchedulerConfig

    // 心跳服务
    Heartbeat   *HeartbeatConfig

    // Token 压缩
    TokenCompression *TokenCompressionConfig

    // 消息总线
    Messaging   *MessagingConfig

    // 技能系统目录
    SkillSystemDir string
}
```

### 组件启用/禁用

```yaml
# 启用/禁用组件
components:
  checkpoint:
    enabled: true
    backend: "sqlite"    # sqlite, redis, memory

  sandbox:
    enabled: true
    backend: "docker"     # docker, native

  monitoring:
    enabled: true
    backend: "prometheus"  # prometheus, log
```

---

## Agent 配置

### 内置 Agent 配置

```yaml
agents:
  # 聊天代理
  - name: "chat"
    type: "chat"
    model: "ollama-llama3"
    instructions: "你是一个有用的AI助手"
    tools:
      - "http_request"
      - "file_operation"
    hitl:
      enabled: true
      approval_mode: "manual"  # manual, auto

  # ReAct 代理
  - name: "worker"
    type: "react"
    model: "ollama-llama3"
    instructions: "你是一个专业的工作代理"
    max_iterations: 10
    tools:
      - "code_execution"
      - "data_processing"

  # 人工代理
  - name: "human"
    type: "human"
    instructions: "人工审核代理"
    timeout: 3600  # 1小时
```

### WorkerAgent 专业角色

```yaml
agents:
  # 开发者代理
  - name: "developer"
    type: "worker"
    role: "developer"
    skills:
      - "code_generation"
      - "code_review"
      - "debugging"

  # 浏览器代理
  - name: "browser"
    type: "worker"
    role: "browser"
    skills:
      - "web_navigation"
      - "form_filling"
      - "data_extraction"

  # 文档代理
  - name: "writer"
    type: "worker"
    role: "writer"
    skills:
      - "content_generation"
      - "editing"
      - "formatting"

  # 多模态代理
  - name: "multimodal"
    type: "worker"
    role: "multimodal"
    skills:
      - "image_analysis"
      - "chart_generation"

  # 研究者代理
  - name: "researcher"
    type: "worker"
    role: "researcher"
    skills:
      - "web_search"
      - "data_collection"
      - "analysis"

  # 审核员代理
  - name: "reviewer"
    type: "worker"
    role: "reviewer"
    skills:
      - "code_review"
      - "quality_check"

  # 报告员代理
  - name: "reporter"
    type: "worker"
    role: "reporter"
    skills:
      - "report_generation"
      - "visualization"
```

---

## Workflow 配置

### 基本工作流

```yaml
workflows:
  # 顺序工作流
  - name: "data_pipeline"
    type: "sequential"
    nodes:
      - name: "collect"
        agent: "worker"
        tools: ["web_search"]
      - name: "analyze"
        agent: "researcher"
        tools: ["data_analysis"]
      - name: "report"
        agent: "reporter"
        tools: ["report_generation"]
    edges:
      - from: "collect"
        to: "analyze"
      - from: "analyze"
        to: "report"

  # 并行工作流
  - name: "parallel_processing"
    type: "parallel"
    nodes:
      - name: "task1"
        agent: "worker"
      - name: "task2"
        agent: "worker"
      - name: "task3"
        agent: "worker"

  # DAG 工作流
  - name: "complex_workflow"
    type: "dag"
    start_node: "start"
    nodes:
      - name: "start"
        agent: "chat"
      - name: "process_a"
        agent: "worker"
        depends_on: ["start"]
      - name: "process_b"
        agent: "worker"
        depends_on: ["start"]
      - name: "merge"
        agent: "analyst"
        depends_on: ["process_a", "process_b"]
```

### 工作流高级配置

```yaml
workflows:
  - name: "advanced_workflow"
    type: "dag"

    # 检查点配置
    checkpoint:
      enabled: true
      backend: "sqlite"
      interval: 60  # 每60秒保存一次

    # 重试策略
    retry:
      max_retries: 3
      backoff_base: 1  # 秒
      backoff_max: 60   # 秒

    # 超时配置
    timeout: 3600  # 整个工作流超时(秒)

    # HITL 配置
    human_in_the_loop:
      enabled: true
      approval_nodes: ["merge"]
      notification_channels: ["slack", "email"]
```

---

## Model 配置

### OpenAI 配置

```yaml
models:
  gpt-4:
    type: "openai"
    model: "gpt-4-turbo"
    api_key: "${OPENAI_API_KEY}"    # 从环境变量读取
    base_url: "https://api.openai.com/v1"
    enabled: true

    # 高级配置
    timeout: 30
    max_retries: 3
    retry_interval: 1

    # 生成参数
    temperature: 0.7
    max_tokens: 4096
    top_p: 1.0
    top_k: 0

    # 请求头
    headers:
      Custom-Header: "value"

    # 优先级
    priority: 10
```

### Ollama 配置

```yaml
models:
  llama3:
    type: "ollama"
    model: "llama3"
    base_url: "http://localhost:11434"
    enabled: true

    # 超时
    timeout: 120

    # 连接池
    pool_size: 10

    # 缓存
    cache_enabled: true
    cache_ttl: 3600  # 秒
```

### LM Studio 配置

```yaml
models:
  local-model:
    type: "lmstudio"
    model: "my-model"
    base_url: "http://localhost:1234/v1"
    api_key: "lm-studio"
    enabled: true
```

---

## Skill 配置

### 技能注册表配置

**文件**: `.skills/registry/registry.yaml`

```yaml
skills:
  # HTTP 请求技能
  - name: "http_request"
    enabled: true
    config:
      timeout: 30
      max_retries: 3
      allowed_hosts:
        - "api.example.com"
        - "*.github.com"

  # 文件操作技能
  - name: "file_operation"
    enabled: true
    config:
      allowed_paths:
        - "/tmp"
        - "/home/user/work"
      max_file_size: 10485760  # 10MB

  # 代码执行技能
  - name: "code_execution"
    enabled: true
    config:
      timeout: 60
      memory_limit: "512m"
      cpu_limit: "1.0"

  # 数据处理技能
  - name: "data_processing"
    enabled: true
    config:
      max_data_size: 1048576
```

### 自定义技能配置

```yaml
skills:
  - name: "my_custom_skill"
    version: "1.0.0"
    category: "custom"
    tags: ["api", "automation"]
    enabled: true

    # 参数定义
    input_schema:
      type: "object"
      properties:
        url:
          type: "string"
          description: "目标 URL"
        method:
          type: "string"
          enum: ["GET", "POST", "PUT", "DELETE"]
          default: "GET"
      required: ["url"]

    # 输出定义
    output_schema:
      type: "object"
      properties:
        status:
          type: "integer"
        data:
          type: "object"

    # 执行配置
    config:
      timeout: 30
      max_retries: 3
      cache_enabled: true
      cache_ttl: 300
```

---

## 高级配置

### 内存管理配置

```yaml
memory:
  # 监控配置
  monitoring:
    enabled: true
    interval: 5s          # 监控间隔
    history_size: 100     # 历史记录数

  # 告警规则
  alert_rules:
    - id: "heap-512mb"
      name: "堆内存告警"
      severity: "warning"
      threshold: 536870912   # 512MB
      operator: ">"
      duration: 30s

    - id: "heap-1gb"
      name: "堆内存严重告警"
      severity: "error"
      threshold: 1073741824  # 1GB
      operator: ">"
      duration: 15s

  # 缓存配置
  cache:
    max_size: 200
    ttl: 1h
    dynamic_weights: true

  # 工作线程配置
  worker:
    count: 5
    initial_pool_size: 10
    max_pool_size: 100
    resize_policy: dynamic    # dynamic, fixed

  # 容器配置
  container:
    max_initial_size: 5
    max_total_size: 20

  # 事件总线配置
  event_bus:
    initial_queue_size: 1000
    max_queue_size: 10000
    resize_threshold: 0.8
```

### 监控配置

```yaml
monitoring:
  # 采样率
  sampling_rate: 0.2    # 20% 采样

  # 最大样本数
  max_samples: 200

  # OpenTelemetry 配置
  opentelemetry:
    enabled: true
    endpoint: "http://localhost:4318"
    headers:
      Authorization: "Bearer token"

  # Prometheus 配置
  prometheus:
    enabled: true
    endpoint: "/metrics"
    port: 9090
```

### 异步任务配置

```yaml
async_task:
  enabled: true

  # 并发配置
  max_concurrent: 10

  # 队列配置
  queue_size: 1000

  # 超时配置
  default_timeout: 300    # 5分钟

  # 重试配置
  retry:
    max_retries: 3
    backoff_base: 1
    backoff_max: 60
```

### 调度器配置

```yaml
scheduler:
  enabled: true
  timezone: "Asia/Shanghai"

  # 并发配置
  max_concurrent_jobs: 5

  # 任务超时
  job_timeout: 3600    # 1小时

  # 日志配置
  log_level: "info"
```

### 心跳服务配置

```yaml
heartbeat:
  enabled: true

  # 心跳间隔
  interval: 30s    # 30秒

  # 超时时间
  timeout: 90s     # 90秒

  # 日志配置
  log_level: "info"
```

### Token 压缩配置

```yaml
token_compression:
  enabled: true

  # 压缩策略
  strategy: "hybrid"    # truncate, summarize, hybrid

  # 目标配置
  target_tokens: 4000
  min_tokens: 500
  max_tokens: 8000

  # 摘要配置
  preserve_system_messages: true
  summary_model: "ollama-llama3"
  summary_max_tokens: 500
  temperature: 0.3
```

### 消息总线配置

```yaml
messaging:
  enabled: true

  # 频道配置
  channels:
    - name: "slack"
      type: "slack"
      enabled: true

    - name: "telegram"
      type: "telegram"
      enabled: false

  # 启用指标
  enable_metrics: true
```

---

## 环境变量

### 通用环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `AF_LOG_LEVEL` | 日志级别 | `info` |
| `AF_LOG_FORMAT` | 日志格式 | `json` |
| `AF_CONFIG_PATH` | 配置文件路径 | `./host.yaml` |
| `AF_WORK_DIR` | 工作目录 | `./` |
| `AF_DATA_DIR` | 数据目录 | `./data` |

### OpenAI 配置

| 变量名 | 说明 |
|--------|------|
| `OPENAI_API_KEY` | OpenAI API 密钥 |
| `OPENAI_BASE_URL` | OpenAI API 地址 |
| `OPENAI_ORG_ID` | OpenAI 组织 ID |

### Ollama 配置

| 变量名 | 说明 |
|--------|------|
| `OLLAMA_BASE_URL` | Ollama 服务地址 |
| `OLLAMA_MODEL` | 默认模型名称 |

### 数据库配置

| 变量名 | 说明 |
|--------|------|
| `REDIS_URL` | Redis 连接地址 |
| `REDIS_PASSWORD` | Redis 密码 |
| `SQLITE_PATH` | SQLite 数据库路径 |

---

## 配置验证

### 验证命令

```bash
# 验证配置文件
agentframework validate host.yaml

# 验证并输出详细信息
agentframework validate --verbose host.yaml

# 验证特定组件
agentframework validate --component models host.yaml
agentframework validate --component agents host.yaml
agentframework validate --component workflows host.yaml
```

### 常见错误

| 错误 | 说明 | 解决方案 |
|------|------|---------|
| `model not found` | 模型未配置 | 检查 `models` 配置 |
| `invalid API key` | API 密钥无效 | 检查环境变量或配置文件 |
| `agent not found` | Agent 不存在 | 检查 `agents` 配置 |
| `skill not found` | 技能未注册 | 检查技能配置 |
| `timeout` | 操作超时 | 增加 timeout 配置 |

---

## 相关文档

- 📘 [快速开始](../quickstart/QUICKSTART.md) - 5 分钟上手指南
- 📘 [最佳实践](BEST_PRACTICES.md) - 配置最佳实践
- 📘 [故障排查](../operation/TROUBLESHOOTING.md) - 配置问题排查
- 📘 [生产部署](../deployment/PRODUCTION.md) - 生产环境配置

---

**Made with ❤️ by AgentFramework Team**
