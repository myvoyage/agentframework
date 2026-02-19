# AgentFramework 多渠道系统使用指南

## 概述

AgentFramework 多渠道系统提供了一个统一的消息接口，支持同时接入多个消息平台：

- **Telegram** - 国际主流即时通讯平台
- **Discord** - 游戏社区和语音聊天平台
- **Slack** - 企业团队协作平台
- **飞书 (Feishu/Lark)** - 字节跳动企业协作平台
- **企业微信 (WeCom)** - 腾讯企业通讯平台
- **钉钉 (DingTalk)** - 阿里企业协作平台
- **QQ** - 腾讯即时通讯平台（支持 OneBot 11 标准）

## 快速开始

### 1. 环境变量配置

最简单的方式是通过环境变量配置渠道：

```bash
# Telegram
export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"

# Discord
export DISCORD_BOT_TOKEN="your_discord_bot_token"

# Slack
export SLACK_BOT_TOKEN="xoxb-your-slack-bot-token"
export SLACK_APP_TOKEN="xapp-your-slack-app-token"

# 飞书
export FEISHU_APP_ID="cli_your_feishu_app_id"
export FEISHU_APP_SECRET="your_feishu_app_secret"

# 企业微信
export WEWORK_CORP_ID="your_wework_corp_id"
export WEWORK_CORP_SECRET="your_wework_corp_secret"
export WEWORK_AGENT_ID="your_wework_agent_id"

# 钉钉
export DINGTALK_APP_KEY="your_dingtalk_app_key"
export DINGTALK_APP_SECRET="your_dingtalk_app_secret"

# QQ (可选)
export QQ_BOT_ENABLED="true"
export QQ_BOT_API_BASE="http://127.0.0.1:3000"
```

### 2. 配置文件

创建 `config/channels.yaml`：

```yaml
version: "1.0"

global:
  default_timeout: 30s
  enable_metrics: true
  enable_tracing: true
  log_level: info

channels:
  telegram:
    type: telegram
    enabled: true
    token: "${TELEGRAM_BOT_TOKEN}"
    rate_limit: 30

  discord:
    type: discord
    enabled: true
    token: "${DISCORD_BOT_TOKEN}"
    rate_limit: 50

routes:
  - id: "accept-all"
    name: "Accept all messages"
    priority: 100
    action: accept
```

### 3. 代码集成

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "AgentFramework/pkg/channels"
    "AgentFramework/pkg/channels/adapters"
)

func main() {
    // 创建通道管理器
    manager, _ := channels.NewManager(&channels.ManagerConfig{})

    // 设置消息处理器
    manager.SetMessageHandler(func(msg *channels.Message) error {
        log.Printf("收到消息: %s", msg.Text)
        // 处理消息...
        return nil
    })

    // 加载配置
    config, _ := channels.LoadConfig("config/channels.yaml")

    // 注册通道
    factory := manager.GetFactory()
    for name, chConfig := range config.GetEnabledChannels() {
        adapter, _ := factory.CreateAdapter(name, chConfig.Type)
        adapter.Initialize(context.Background(), chConfig)
        manager.RegisterAdapter(adapter)
    }

    // 启动
    manager.Start(context.Background())
    defer manager.Stop(context.Background())

    // 等待退出信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
}
```

## 平台特定配置

### Telegram

1. 创建机器人：与 [@BotFather](https://t.me/botfather) 对话
2. 获取 API Token
3. 配置环境变量：`TELEGRAM_BOT_TOKEN`

**特性：**
- 支持消息编辑
- 支持内联键盘
- 支持长轮询和 Webhook

### Discord

1. 创建应用：访问 [Discord Developer Portal](https://discord.com/developers/applications)
2. 创建 Bot 并获取 Token
3. 配置环境变量：`DISCORD_BOT_TOKEN`

**特性：**
- 支持 Slash 命令
- 支持嵌入消息 (Embeds)
- 支持线程和反应

### Slack

1. 创建应用：访问 [Slack API](https://api.slack.com/apps)
2. 启用 Bot Token Scopes
3. 配置环境变量：`SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`

**特性：**
- 支持 Socket Mode
- 支持 Block Kit
- 支持线程和消息编辑

### 飞书

1. 创建应用：访问 [飞书开放平台](https://open.feishu.cn/app)
2. 获取 App ID 和 App Secret
3. 配置环境变量：`FEISHU_APP_ID`, `FEISHU_APP_SECRET`

**特性：**
- 支持富文本消息
- 支持卡片消息
- 支持消息更新

### 企业微信

1. 创建应用：访问 [企业微信管理后台](https://work.weixin.qq.com/)
2. 获取 Corp ID、Corp Secret 和 Agent ID
3. 配置环境变量：`WEWORK_CORP_ID`, `WEWORK_CORP_SECRET`, `WEWORK_AGENT_ID`

**特性：**
- 支持应用消息
- 支持接收消息回调
- 支持素材上传

### 钉钉

1. 创建应用：访问 [钉钉开放平台](https://open.dingtalk.com/)
2. 获取 App Key 和 App Secret
3. 配置环境变量：`DINGTALK_APP_KEY`, `DINGTALK_APP_SECRET`

**特性：**
- 支持群机器人
- 支持 Markdown 消息
- 支持 @ 提及

### QQ (OneBot 11)

1. 安装 OneBot 实现：[go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 或 [NapCat](https://github.com/NapNeko/NapCatQQ)
2. 启动服务（默认端口 3000）
3. 配置环境变量：`QQ_BOT_ENABLED=true`

**特性：**
- 支持 CQ 码消息段
- 支持群消息和私聊
- 支持图片、语音、视频等

## 高级功能

### 消息路由

配置路由规则来控制消息处理：

```yaml
routes:
  # 接受所有文本消息
  - id: "accept-text"
    priority: 100
    message_type: [text]
    action: accept

  # 拒绝特定用户
  - id: "block-users"
    priority: 200
    user_id: ["blocked_user_1"]
    action: reject

  # 转发管理员消息
  - id: "forward-admin"
    priority: 150
    pattern: "^/admin"
    action: forward
    action_data:
      target_channel: "admin"

  # 限流规则
  - id: "rate-limit"
    priority: 180
    pattern: "^/"
    rate_limit: 10
    rate_window: 60s
    action: accept
```

### 发送消息

```go
// 发送文本消息
msg := &channels.Message{
    Type:    channels.MessageTypeText,
    Text:    "Hello, World!",
}

opts := channels.MessageSendOptions{
    ParseMode: "markdown",
}

messageID, err := manager.SendMessage(ctx, "telegram", msg, opts)

// 发送图片消息
msg := &channels.Message{
    Type: channels.MessageTypeImage,
    Attachments: []channels.Attachment{
        {
            URL: "https://example.com/image.jpg",
        },
    },
}

// 回复消息
opts := channels.MessageSendOptions{
    ReplyTo: "original_message_id",
}
```

### 广播消息

```go
// 向所有 Telegram 渠道广播
results, err := manager.Broadcast(
    ctx,
    channels.ChannelTypeTelegram,
    msg,
    channels.MessageSendOptions{},
)
```

### 统计信息

```go
stats, err := manager.GetStats(ctx)
for channelID, stat := range stats {
    log.Printf("Channel %s:", channelID)
    log.Printf("  Status: %s", stat.Status)
    log.Printf("  Sent: %d", stat.MessagesSent)
    log.Printf("  Received: %d", stat.MessagesReceived)
    log.Printf("  Errors: %d", stat.ErrorCount)
}
```

### 事件处理

```go
manager.SetEventHandler(func(event channels.Event) {
    switch event.Type {
    case channels.EventTypeConnected:
        log.Printf("Channel %s connected", event.ChannelID)
    case channels.EventTypeDisconnected:
        log.Printf("Channel %s disconnected", event.ChannelID)
    case channels.EventTypeError:
        log.Printf("Channel %s error: %v", event.ChannelID, event.Error)
    case channels.EventTypeMessageSent:
        log.Printf("Message sent to %s", event.ChannelID)
    }
})
```

## API 参考

### 核心类型

| 类型 | 说明 |
|------|------|
| `ChannelAdapter` | 通道适配器接口 |
| `Manager` | 多通道管理器 |
| `Router` | 消息路由器 |
| `Message` | 统一消息格式 |
| `ChannelConfig` | 通道配置 |
| `RoutingRule` | 路由规则 |

### 消息类型

| 类型 | 说明 |
|------|------|
| `MessageTypeText` | 文本消息 |
| `MessageTypeImage` | 图片消息 |
| `MessageTypeAudio` | 音频消息 |
| `MessageTypeVideo` | 视频消息 |
| `MessageTypeFile` | 文件消息 |
| `MessageTypeCommand` | 命令消息 |

### 渠道类型

| 类型 | 说明 |
|------|------|
| `ChannelTypeTelegram` | Telegram |
| `ChannelTypeDiscord` | Discord |
| `ChannelTypeSlack` | Slack |
| `ChannelTypeFeishu` | 飞书 |
| `ChannelTypeWeWork` | 企业微信 |
| `ChannelTypeDingTalk` | 钉钉 |
| `ChannelTypeQQ` | QQ |

## 测试

运行单元测试：

```bash
go test ./pkg/channels/...
```

运行特定测试：

```bash
go test -v ./pkg/channels/... -run TestMessageType
```

## 故障排查

### 连接问题

1. 检查网络连接
2. 验证 API Token 是否正确
3. 查看日志中的详细错误信息

### QQ 机器人连接失败

1. 确认 OneBot 实现正在运行
2. 检查 API 地址配置（默认 `http://127.0.0.1:3000`）
3. 验证端口是否开放

### 钉钉/飞书/企业微信回调问题

1. 确认服务器公网 IP 可访问
2. 验证回调 URL 配置正确
3. 检查加密密钥配置

## 性能优化

1. **使用连接池**：复用 HTTP 连接
2. **启用限流**：防止超过 API 限制
3. **异步处理**：使用 goroutine 处理消息
4. **批量操作**：合并多个消息请求

## 安全建议

1. **保护 API Token**：使用环境变量或密钥管理服务
2. **验证回调**：启用签名验证
3. **限流保护**：防止恶意请求
4. **日志脱敏**：避免记录敏感信息

## 许可证

本项目采用 AGPL-3.0-or-later 许可证。

---

**相关链接：**
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Discord Developer Portal](https://discord.com/developers/docs/intro)
- [Slack API](https://api.slack.com/)
- [飞书开放平台](https://open.feishu.cn/document)
- [企业微信 API](https://developer.work.weixin.qq.com/document/)
- [钉钉开放平台](https://open.dingtalk.com/document/)
- [OneBot 11 标准](https://github.com/botuniverse/onebot-11)
