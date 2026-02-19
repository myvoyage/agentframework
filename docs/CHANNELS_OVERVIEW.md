# AgentFramework 多渠道系统

## 项目概述

AgentFramework 多渠道系统是一个统一的消息接入和管理框架，支持同时连接 7 个主流消息平台。系统采用适配器模式设计，实现了统一的 API 接口，使得开发者可以轻松地将消息机器人部署到多个平台。

## 支持的平台

| 平台 | 类型 | 状态 | 特色功能 |
|------|------|------|----------|
| Telegram | 国际 | ✅ | 编辑消息、内联键盘、长轮询/Webhook |
| Discord | 国际 | ✅ | Slash 命令、嵌入消息、反应 |
| Slack | 国际 | ✅ | Socket Mode、Block Kit、线程 |
| 飞书 (Feishu) | 国内 | ✅ | 富文本、卡片消息、应用消息 |
| 企业微信 (WeCom) | 国内 | ✅ | 企业应用、回调验证、素材上传 |
| 钉钉 (DingTalk) | 国内 | ✅ | 群机器人、Markdown、@提及 |
| QQ | 国内 | ✅ | OneBot 11 标准、CQ 码、表情包 |

## 项目结构

```
AgentFramework/
├── pkg/channels/                    # 多渠道核心模块
│   ├── types.go                     # 核心类型定义 (250+ 行)
│   ├── types_test.go                # 单元测试 (400+ 行)
│   ├── adapter.go                   # 适配器接口 (300+ 行)
│   ├── manager.go                   # 通道管理器 (300+ 行)
│   ├── router.go                    # 消息路由器 (400+ 行)
│   ├── config.go                    # 配置管理 (500+ 行)
│   ├── README.md                    # 使用文档
│   └── adapters/                    # 平台适配器
│       ├── common.go                # 通用基类 (400+ 行)
│       ├── telegram.go              # Telegram 适配器 (400+ 行)
│       ├── discord.go               # Discord 适配器 (400+ 行)
│       ├── slack.go                 # Slack 适配器 (350+ 行)
│       ├── feishu.go                # 飞书适配器 (400+ 行)
│       ├── wework.go                # 企业微信适配器 (400+ 行)
│       ├── dingtalk.go              # 钉钉适配器 (350+ 行)
│       └── qq.go                    # QQ 适配器 (400+ 行)
├── cmd/simplebot/main.go            # 启动示例程序 (300+ 行)
├── examples/channels_integration.go # 集成示例 (500+ 行)
├── config/
│   ├── channels.example.yaml        # YAML 配置示例
│   └── channels.example.json        # JSON 配置示例
├── docs/
│   ├── CHANNEL_INTEGRATION.md       # 系统集成指南
│   └── API.md                       # API 文档
├── Makefile                          # 构建工具
├── .env.example                      # 环境变量示例
└── go.mod                            # Go 模块配置
```

## 核心功能

### 1. 统一消息格式

系统将所有平台的消息转换为统一格式：

```go
type Message struct {
    ID          string          // 消息唯一标识
    Type        MessageType     // 消息类型 (文本/图片/音频/视频/文件/命令)
    Direction   MessageDirection // 消息方向 (接收/发送)
    Text        string          // 文本内容
    Attachments []Attachment    // 附件列表
    Mentions    []Mention       // @提及列表
    ChannelID   string          // 渠道标识
    ChannelType ChannelType    // 渠道类型
    ChatID      string          // 聊天标识
    From        *User           // 发送者信息
    Timestamp   time.Time       // 时间戳
    Metadata    map[string]string // 平台特定数据
}
```

### 2. 适配器接口

所有平台适配器实现统一的接口：

```go
type ChannelAdapter interface {
    // 生命周期管理
    Initialize(ctx context.Context, config ChannelConfig) error
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    IsConnected() bool
    GetStatus(ctx context.Context) (ChannelStatus, error)
    GetStats(ctx context.Context) (*ChannelStats, error)

    // 消息操作
    SendMessage(ctx context.Context, msg *Message, opts MessageSendOptions) (string, error)
    EditMessage(ctx context.Context, messageID string, msg *Message) error
    DeleteMessage(ctx context.Context, messageID string) error
    UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*Attachment, error)

    // 事件处理
    SetMessageHandler(handler MessageHandler)
    SetEventHandler(handler EventHandler)

    // 能力查询
    Supports(feature string) bool
}
```

### 3. 消息路由

支持基于优先级的路由规则：

```go
type RoutingRule struct {
    ID          string              // 规则标识
    Priority    int                 // 优先级 (越高越先)
    ChannelType []ChannelType       // 匹配的渠道类型
    MessageType []MessageType      // 匹配的消息类型
    Pattern     string              // 正则表达式匹配
    Action      RoutingAction       // 动作 (接受/拒绝/转发/转换/延迟)
    RateLimit   int                 // 限流数量
    RateWindow  time.Duration       // 限流窗口
}
```

### 4. 配置管理

支持多种配置方式：

- **YAML 配置文件**: `config/channels.yaml`
- **JSON 配置文件**: `config/channels.json`
- **环境变量**: 支持 `.env` 文件
- **热重载**: 自动检测配置文件变化

## 快速开始

### 方法 1: 使用环境变量

```bash
# 设置环境变量
export TELEGRAM_BOT_TOKEN="your_token"
export QQ_BOT_ENABLED=true

# 运行
make run
```

### 方法 2: 使用配置文件

```bash
# 复制示例配置
cp config/channels.example.yaml config/channels.yaml

# 编辑配置文件，填入你的 API 凭证

# 运行
make run
```

### 方法 3: 代码集成

```go
import "AgentFramework/pkg/channels"

// 创建管理器
manager, _ := channels.NewManager(nil)

// 设置消息处理器
manager.SetMessageHandler(func(msg *channels.Message) error {
    fmt.Printf("收到消息: %s\n", msg.Text)
    return nil
})

// 加载配置
config, _ := channels.LoadConfig("config/channels.yaml")

// 注册渠道
for name, chConfig := range config.GetEnabledChannels() {
    adapter, _ := factory.CreateAdapter(name, chConfig.Type)
    adapter.Initialize(ctx, chConfig)
    manager.RegisterAdapter(adapter)
}

// 启动
manager.Start(ctx)
```

## 设计原则

系统严格遵循以下设计原则：

### SOLID 原则

- **单一职责 (SRP)**: 每个适配器只负责一个平台的通信逻辑
- **开闭原则 (OCP)**: 新平台通过实现接口添加，无需修改现有代码
- **里氏替换 (LSP)**: 所有适配器可以互换使用
- **接口隔离 (ISP)**: 接口专注于核心功能
- **依赖倒置 (DIP)**: 依赖抽象接口而非具体实现

### DRY 原则

- 通用适配器基类实现公共逻辑
- 工具函数封装常用操作

### KISS 原则

- 简洁的 API 设计
- 清晰的消息格式

## 性能特性

- **异步处理**: 使用 goroutine 处理消息和事件
- **连接池**: 复用 HTTP 连接
- **限流保护**: 防止超过 API 限制
- **统计监控**: 内置消息统计和健康检查

## 可观测性

### 指标统计

每个渠道自动收集：

- 发送/接收消息数
- 失败消息数
- 字节数统计
- 错误计数
- 连接时间
- 重连次数

### 事件监控

系统发送以下事件：

- `connected`: 渠道连接成功
- `disconnected`: 渠道断开连接
- `reconnecting`: 渠道重连中
- `error`: 渠道错误
- `message_received`: 收到消息
- `message_sent`: 发送消息
- `message_failed`: 发送失败

### OpenTelemetry 集成

内置分布式追踪支持：

```go
import "go.opentelemetry.io/otel"

manager, _ := channels.NewManager(&channels.ManagerConfig{
    EnableTracing: true,
})
```

## 安全特性

- **Token 保护**: 使用环境变量存储敏感信息
- **回调验证**: 飞书/企业微信/钉钉支持签名验证
- **限流保护**: 防止恶意请求
- **日志脱敏**: 避免记录敏感信息

## 测试

```bash
# 运行所有测试
make test

# 只测试渠道模块
make test-channels

# 生成覆盖率报告
make test-coverage
```

## 构建

```bash
# 构建简单机器人
make build-simplebot

# 构建所有程序
make build

# 跨平台构建
make release
```

## 开发工具

```bash
# 开发环境设置
make dev-setup

# 代码格式化
make fmt

# 静态分析
make vet

# 完整检查
make check
```

## 相关文档

- [多渠道系统使用指南](../pkg/channels/README.md)
- [系统集成指南](../docs/CHANNEL_INTEGRATION.md)
- [配置示例](../config/channels.example.yaml)

## 技术栈

- **语言**: Go 1.25+
- **依赖库**:
  - `github.com/bwmarrin/discordgo` - Discord API
  - `github.com/slack-go/slack` - Slack API
  - `gopkg.in/telebot.v3` - Telegram API
  - `gopkg.in/yaml.v3` - YAML 解析
  - `go.opentelemetry.io/otel` - 分布式追踪

## 许可证

本项目采用 AGPL-3.0-or-later 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

---

**项目状态**: ✅ 生产就绪

**最后更新**: 2025-02-18
