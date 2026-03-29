# AgentFramework 文档

> 高性能企业级 AI Agent 框架 · Go 语言实现 · AGPL-3.0

---

## 快速导航

| 想做什么 | 去哪里 |
|---------|--------|
| 5 分钟跑起来 | [快速上手](quickstart/QUICKSTART.md) |
| 理解整体架构 | [架构概览](architecture/ARCHITECTURE_OVERVIEW.md) |
| 配置 Agent / 模型 / 渠道 | [配置指南](configuration/CONFIGURATION.md) |
| 查 REST / WebSocket API | [API 参考](api/API.md) |
| 开发自定义 Skill | [Skill 开发](SKILL_DEVELOPMENT.md) |
| 接入即时通讯渠道 | [渠道集成](CHANNEL_INTEGRATION.md) |
| 生产部署 / Docker | [部署指南](DEPLOYMENT_GUIDE.md) |
| TUI 终端界面 | [TUI 使用指南](TUI_USER_GUIDE.md) |

---

## 项目简介

AgentFramework 是一个用 Go 语言从头实现的企业级 AI Agent 运行时，参照 OpenClaw 架构设计。它不是 SDK 包装器——而是完整的 Agent 基础设施：

- **Gateway** — WebSocket + HTTP 混合网关，单端口双协议
- **ReAct 执行循环** — 模型 → 工具 → 模型，内置流式输出
- **Lane Queue** — 基于 `workspace:channel:userId` 的会话串行队列，消除竞态
- **Workflow DAG** — 并发安全的有向无环图工作流引擎
- **Markdown Skill** — 用 SKILL.md 文件定义技能，热重载，渐进式注入
- **多渠道** — Telegram、飞书、企业微信、QQ、Discord、Slack、WebChat、CLI

---

## 目录结构

```
docs/
├── README.md                    # 本文件：文档导航
├── architecture/
│   └── ARCHITECTURE_OVERVIEW.md # 四层架构详解
├── quickstart/
│   └── QUICKSTART.md            # 快速上手
├── configuration/
│   └── CONFIGURATION.md         # 完整配置参考
├── api/
│   └── API.md                   # REST + WebSocket API
├── guides/
│   └── best-practices/          # 最佳实践
├── components/                  # 核心组件详解
├── iot/                         # IoT 专题文档
├── examples/                    # 示例代码说明
├── SKILL_DEVELOPMENT.md         # Skill 开发指南
├── CHANNEL_INTEGRATION.md       # 渠道接入指南
├── CHANNELS_OVERVIEW.md         # 渠道概览
├── DEPLOYMENT_GUIDE.md          # 部署运维
├── TUI_USER_GUIDE.md            # TUI 终端界面
├── CLI_USAGE.md                 # CLI 命令参考
├── MULTIMODE_USAGE.md           # 多模式使用
├── QUICK_REFERENCE.md           # 速查手册
└── REFACTORING_GUIDE.md         # 重构指南（贡献者用）
```

---

## 核心概念速览

### Lane Queue

每条消息由 `SessionKey`（`workspace:channel:userId`）路由到独立队列，同一会话串行执行，不同会话并行。这是框架保证消息顺序、消除竞态的核心机制。

### ReAct 执行循环

```
用户消息
   ↓
[组装上下文] ← SOUL.md + MEMORY.md + 会话历史
   ↓
[模型推理]
   ↓
有工具调用? ──是──→ [并行/串行执行工具] ──→ [收集结果] ──→ [模型推理]
   ↓ 否
[流式输出到渠道]
```

最大迭代次数默认 10，工具超时默认 30s，整体超时默认 5min。

### Markdown Skill

在任意目录放一个 `SKILL.md` 文件即可定义 Skill：

```markdown
---
name: my_tool
description: 做某件事
parameters:
  query:
    type: string
    description: 查询内容
    required: true
---

# 使用说明

调用此工具时，应当...
```

框架启动时只扫描 `name + description`（渐进式披露），激活时才注入全文到上下文。

---

## 构建与运行

```bash
# 构建 CLI
go build -o build/afcli.exe ./cmd/afcli/

# 启动 Gateway（默认端口 18640）
./build/afcli gateway

# 指定端口 + 详细日志
./build/afcli gateway --port 8080 --verbose

# 强制终止占用端口的进程后启动
./build/afcli gateway --force

# Dev 模式
./build/afcli --dev gateway
```

---

## 许可证

AGPL-3.0-or-later — 详见 [LICENSE](../LICENSE)
