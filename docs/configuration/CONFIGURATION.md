# 配置指南

> AgentFramework 完整配置参考
> 配置文件：`host.yaml`（项目根目录）

---

## 目录

- [配置优先级](#配置优先级)
- [基础配置](#基础配置)
- [模型配置](#模型配置)
- [Agent 配置](#agent-配置)
- [工作流配置](#工作流配置)
- [渠道配置](#渠道配置)
- [线程存储](#线程存储)
- [技能系统](#技能系统)
- [调度器](#调度器)
- [Token 压缩](#token-压缩)
- [Gateway 配置](#gateway-配置)
- [环境变量注入](#环境变量注入)

---

## 配置优先级

```
环境变量 ${VAR}  >  host.yaml 文件值  >  默认值
```

`host.yaml` 放在项目根目录，启动时自动加载。完整示例见 `config/host.example.yaml`。

---

## 基础配置

```yaml
name: agent_framework    # 应用名称
version: "2.0.0"         # 版本
defaultModel: "default"  # 默认模型名（必须在 models 中定义）
```

---

## 模型配置

```yaml
models:
  # LM Studio 本地模型（推荐本地测试）
  default:
    provider: "lmstudio"
    baseurl: "http://localhost:1234/v1"
    model: "local-model"
    maxtokens: 4096
    temperature: 0.7

  # Ollama 本地模型
  ollama:
    provider: "ollama"
    baseurl: "http://localhost:11434/api"
    model: "llama3"
    maxtokens: 4096
    temperature: 0.7

  # 智谱 GLM
  glm:
    provider: "zhipu"
    baseurl: "https://open.bigmodel.cn/api/paas/v4"
    model: "glm-4-flash"
    apikey: "${ZHIPU_API_KEY}"

  # DeepSeek
  deepseek:
    provider: "deepseek"
    baseurl: "https://api.deepseek.com/v1"
    model: "deepseek-chat"
    apikey: "${DEEPSEEK_API_KEY}"

  # OpenAI
  gpt4:
    provider: "openai"
    baseurl: "https://api.openai.com/v1"
    model: "gpt-4-turbo-preview"
    apikey: "${OPENAI_API_KEY}"
    maxtokens: 8192
```

**provider 可选值**：`lmstudio` | `ollama` | `zhipu` | `deepseek` | `openai`

---

## Agent 配置

```yaml
agents:
  - name: default_chat         # Agent 名称（唯一）
    kind: chat                 # chat | react
    model: "default"           # 使用哪个模型
    instructions: |
      你是一个有帮助的 AI 助手。
      请用中文回复中文问题。
    tools:                     # 可用工具列表
      - web_search
      - file_read
    middlewares:               # 中间件
      - logging
      - recovery

  - name: code_agent
    kind: react                # ReAct 模式：支持工具调用循环
    model: "gpt4"
    instructions: "你是一个代码助手。"
    tools:
      - code_execute
      - file_write
    maxMessages: 50            # 上下文窗口最大消息数
    enableTrimming: true       # 自动裁剪过长上下文
    trimRatio: 0.8             # 裁剪到 80%
```

**kind 说明**：
- `chat` — 基础对话，无工具调用循环
- `react` — ReAct 模式，支持 `工具调用 → 结果 → 继续推理` 循环

---

## 工作流配置

```yaml
workflows:
  # 顺序执行
  - name: simple_chat
    kind: sequential
    steps:
      - default_chat

  # 并行执行 + 聚合
  - name: parallel_analysis
    kind: aggregating_parallel
    agents:
      - default_chat
      - code_agent
    aggregator: default_chat    # 汇总 Agent
```

**kind 可选值**：`sequential` | `aggregating_parallel` | `dag` | `routing`

---

## 渠道配置

### 飞书 / Lark（推荐）

WebSocket 模式无需公网 URL，本地开发最方便：

```yaml
channels:
  lark:
    enabled: true
    domain: "feishu"               # feishu(国内版) / lark(国际版)
    appId: "${LARK_APP_ID}"        # cli_xxx
    appSecret: "${LARK_APP_SECRET}"
    connectionMode: "websocket"    # websocket(推荐) / webhook
    dmPolicy: "pairing"            # 私信策略：pairing/allowlist/open/disabled
    groupPolicy: "open"            # 群聊策略：open/allowlist/disabled
    streaming: true                # 流式回复
    botName: "AI 助手"
```

Webhook 模式（需要公网 URL）：

```yaml
channels:
  lark:
    enabled: true
    domain: "feishu"
    appId: "${LARK_APP_ID}"
    appSecret: "${LARK_APP_SECRET}"
    connectionMode: "webhook"
    port: 8089
    encryptKey: "${LARK_ENCRYPT_KEY}"
    verifyToken: "${LARK_VERIFY_TOKEN}"
```

### 企业微信

```yaml
channels:
  wechat:
    enabled: true
    type: "wecom"
    corpId: "${WECHAT_CORP_ID}"
    agentId: "${WECHAT_AGENT_ID}"
    corpSecret: "${WECHAT_CORP_SECRET}"
    token: "${WECHAT_TOKEN}"
    encryptKey: "${WECHAT_ENCRYPT_KEY}"
    port: 8080
    sessionPolicy: "open"         # open/restricted/private
```

### Telegram

```yaml
channels:
  telegram:
    enabled: true
    botToken: "${TELEGRAM_BOT_TOKEN}"
    webhookUrl: "${TELEGRAM_WEBHOOK_URL}"
    port: 8443
    allowedChats: []              # 空表示不限制
```

### Slack

```yaml
channels:
  slack:
    enabled: true
    botToken: "${SLACK_BOT_TOKEN}"
    appToken: "${SLACK_APP_TOKEN}"
    signingSecret: "${SLACK_SIGNING_SECRET}"
    socketMode: true              # Socket Mode 无需公网 URL
```

### 钉钉

```yaml
channels:
  dingtalk:
    enabled: true
    clientId: "${DINGTALK_CLIENT_ID}"
    clientSecret: "${DINGTALK_CLIENT_SECRET}"
    agentId: "${DINGTALK_AGENT_ID}"
    port: 8090
    token: "${DINGTALK_TOKEN}"
    encryptKey: "${DINGTALK_ENCRYPT_KEY}"
    sessionPolicy: "open"
```

---

## 线程存储

对话历史的持久化方式：

```yaml
threadStore:
  type: "memory"          # memory | file | redis | sql
  maxMessages: 100        # 每个线程最大消息数
  maxMessageSize: 1048576 # 单条消息最大字节 (1MB)
  ttl: 86400              # 线程过期时间（秒）

  # file 模式
  dir: "./data/threads"

  # redis 模式
  # redisAddr: "localhost:6379"
  # redisPrefix: "agent:"

  # sql 模式
  # driver: "postgres"
  # dsn: "postgres://user:pass@localhost:5432/agent?sslmode=disable"
  # table: "threads"
```

---

## 技能系统

```yaml
skillSystemDir: "./skills"    # Skill 根目录（默认 ./.skills）
```

Skill 目录结构：

```
.skills/
└── my_tool/
    └── SKILL.md
```

框架还会自动扫描：
- `agent/skills/bundled/` — 内置 Skill（最高优先级）
- `~/.agentframework/skills/` — 用户级 Skill

---

## 调度器

```yaml
scheduler:
  enabled: false
  timezone: "Asia/Shanghai"
  maxJobs: 10
  jobTimeout: 300              # 单个任务超时秒数
```

---

## Token 压缩

当上下文过长时自动压缩：

```yaml
tokenCompression:
  enabled: false
  strategy: "truncate"         # truncate | summarize | semantic | hybrid
  targetTokens: 4000
  minTokens: 1000
  maxTokens: 8000
  preserveSystemMessages: true # 保留系统消息不压缩
```

---

## Gateway 配置

Gateway 启动参数通过命令行传入（不在 host.yaml 中）：

```bash
./build/afcli gateway \
  --port 18640 \       # 监听端口（默认 18640）
  --verbose \          # 详细日志
  --force              # 强制终止占用端口的进程
```

---

## 环境变量注入

配置文件中 `${VAR_NAME}` 格式会在运行时从环境变量读取：

```bash
# 设置环境变量
export LARK_APP_ID=cli_xxx
export LARK_APP_SECRET=xxx
export DEEPSEEK_API_KEY=sk-xxx

# 或者用 .env 文件（需要自行加载）
```

建议把敏感信息（API Key、Secret）全部用环境变量注入，不要硬写到配置文件。
