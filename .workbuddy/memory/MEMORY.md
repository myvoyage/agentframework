# OpenClaw AgentFramework Memory

## 项目概述

OpenClaw AgentFramework 是一个高性能企业级 AI Agent 框架，使用 Go 语言开发。

## 核心架构

### 分层模块化架构

1. **Gateway 网关层** - 请求入口、WebSocket、HTTP API
2. **Agent 智能体层** - 核心运行时、任务编排
3. **Skills 技能层** - 可扩展的工具/技能模块
4. **Channels 渠道层** - 多渠道接入（Telegram、企业微信等）

### 核心模块

- `agent/` - Agent 核心运行时
- `gateway/` - Gateway 网关服务
- `pkg/workspace/` - Workspace 配置系统
- `pkg/framework/` - Agent 框架组件

## CLI 命令

```bash
# 构建
go build -o build/afcli.exe ./cmd/afcli/

# Gateway 命令
af gateway --port 18640        # 启动网关（默认端口）
af gateway --verbose           # 调试日志
af gateway --force             # 终止占用端口的进程
af --dev gateway               # dev 模式
```

## 配置系统 (pkg/workspace/)

参考 OpenClaw 架构文章实现的 Workspace 配置系统：

### 核心文件
- `config.go` - 配置解析（SOUL.md / AGENTS.md / CAPABILITIES.md）
- `prompt.go` - 系统提示词动态组合
- `memory.go` - 长期记忆存储与检索
- `sandbox.go` - Docker 沙箱隔离
- `channel.go` - 多渠道适配器

### Workspace 文件
- `SOUL.md` - Agent 灵魂配置（身份、个性、行为准则）
- `AGENTS.md` - Agent 定义列表
- `CAPABILITIES.md` - 可用能力列表
- `MEMORY.md` - 长期记忆

### 提示词组成顺序
1. Header - Agent 身份
2. Soul - 个性和价值
3. Capabilities - 可用能力
4. Skills - 当前激活的技能
5. Memory - 相关记忆
6. Guidelines - 行为准则

## 沙箱隔离

支持三种隔离级别：
- `none` - 无隔离
- `process` - 进程级隔离
- `docker` - Docker 容器隔离

## 多渠道支持

支持的渠道类型：
- `telegram` - Telegram Bot
- `wechat` - 企业微信
- `lark` - 飞书 (Feishu)
- `qq` - QQ / QQ频道
- `discord` - Discord
- `slack` - Slack
- `webchat` - 网页聊天
- `cli` - 命令行

### 飞书 (Lark) 适配器
- 文件: `pkg/workspace/lark.go`
- 支持: 文本/图片/语音/视频/文件消息、交互式卡片
- Webhook 端口: `:8089/lark`（默认）
- API: 飞书开放平台 IM API
- **重构特性**:
  - 完整支持 im.message.receive_v1 事件
  - AES-256 加密解密
  - URL 验证 (challenge)
  - 事件去重 (EventCache)
  - Token 自动刷新
  - 支持飞书卡片消息 (interactive)

### QQ 适配器
- 文件: `pkg/workspace/qq.go`
- 支持: QQ频道/私信
- Webhook 端口: `:8088/qq`
- API: QQ 机器人 API

## 项目约定

### 代码风格
- 遵循 Go 语言惯例
- 使用 `gopkg.in/yaml.v3` 进行 YAML 解析
- 模块化设计，职责单一

### 测试
- 使用 `go test ./...` 运行测试
- 单元测试覆盖核心模块

## OpenClaw 原版架构深度笔记（2026-03-22 研读）

### 四层架构（官方）
1. **控制平面（Control Plane）** - Token 认证、设备配对（类蓝牙流程）
2. **网关层（Gateway）** - 单进程 Node.js，WebSocket + HTTP 混合，零开销内部路由
3. **代理运行时（Agent Runtime）** - 上下文组装 + ReAct 循环 + 记忆刷写
4. **节点层（Nodes）** - 物理设备抽象（macOS/iOS/Android）、摄像头/屏幕/GPS/Canvas

### Gateway WebSocket 协议帧（三种）
```json
// connect 帧 - 必须是首帧
{"type":"connect","id":"<str>","params":{"auth":{"token?"},"deviceIdentity":{}}}

// req/res 请求响应
{"type":"req","id":"<str>","method":"<str>","params":{},"idempotencyKey":"<str>"}
{"type":"res","id":"<str>","ok":true,"payload":{}}

// event 服务端推送
{"type":"event","event":"<str>","payload":{},"seq":1,"stateVersion":1}
```
- 副作用方法（send/agent）必须带 `idempotencyKey` 防重放
- 非 JSON 或首帧不是 connect → 立即关闭连接
- 认证：Token 认证（可选）+ 设备配对（必须）+ 签名 challenge

### Lane Queue（核心可靠性模式）
- 每个会话分配独立队列，**默认串行**，显式并行
- SessionKey 格式：`workspace:channel:userId`
  - 同用户不同渠道 → 独立会话队列（完全隔离）
  - cron/subagent 可使用独立 lane 并行运行
- 消除竞态条件，内置背压

### Agent 上下文组装（5 个来源）
1. 系统提示词（身份定义）
2. Workspace 文件（SOUL.md、USER.md）
3. 记忆文件（MEMORY.md 长期 + 每日 memory/YYYY-MM-DD.md）
4. 会话历史（当前对话上下文窗口）
5. 工具执行结果（前序操作返回值）

### Workspace 配置文件（与我们实现对应）
- `AGENTS.md` - 核心指令（基线规则，能做什么/不能做什么）
- `SOUL.md` - 人格与语气指导
- `TOOLS.md` / `CAPABILITIES.md` - 工具使用说明
- `MEMORY.md` - 长期记忆
- `memory/YYYY-MM-DD.md` - 每日记忆
- Skills 文件夹（按需注入，渐进式披露）

### Skills 系统（Markdown 驱动）
- 每个 Skill = 文件夹 + `SKILL.md`
- Frontmatter：名称/描述/参数/权限
- 正文：自然语言操作指南 + 可引用脚本
- 启动时只加载名称+简短描述（~97字符），激活时注入完整内容
- 支持热重载（文件变更后下次对话生效）
- Agent 可自动创建/编辑 SKILL.md（自写技能）

### 我们 Go 实现 vs 原版对比
| 特性 | 原版（Node.js） | 我们（Go） |
|------|---------------|-----------|
| 网关 | Node.js 单进程 | Go goroutine |
| Agent 运行时 | Pi Agent 框架 | 自实现 |
| Skills | Markdown 文件夹 | 同样 Markdown |
| SessionKey | workspace:channel:userId | 待实现 |
| Lane Queue | 内置 | 待实现 |
| 记忆 | SQLite + 向量嵌入 | 文件 + 检索 |

### 待实现的关键机制
1. **Lane Queue** - agent/ 目录下需实现 sessionKey 串行队列
2. **ReAct 循环** - agent/execution.go 工具调用循环
3. **Skills 渐进式披露** - 启动时只扫描名称，按需加载全文
4. **runId 可观测性** - 每次 agent 运行分配唯一 runId 用于追踪

## 开发记录

### 2026-03-21
- 完成 Gateway 网关实现
- 将 Gateway 从独立应用改为 af CLI 子命令
- 默认端口改为 18640
- 实现 Workspace 配置系统（SOUL.md / AGENTS.md / CAPABILITIES.md）
- 实现系统提示词动态组合
- 实现 Docker 沙箱隔离
- 实现 Memory 检索系统
- 实现 Channel 多渠道适配器
