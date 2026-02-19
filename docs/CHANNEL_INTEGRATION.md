# AgentFramework 多渠道系统集成指南

## 概述

本文档说明如何将 `pkg/channels` 多渠道系统集成到现有的 AgentFramework 中。

## 架构设计

### 系统层次

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│                      (app.go, desktop)                      │
├─────────────────────────────────────────────────────────────┤
│                   Core Application Layer                     │
│                  (core/application.go)                      │
│  ┌───────────────┬───────────────┬─────────────────────┐  │
│  │   Skills      │   Workflows   │     Channels        │  │
│  │   System      │    Manager    │      Module         │  │
│  └───────────────┴───────────────┴─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                      Agent Host Layer                        │
│                     (agent/host.go)                         │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
User Message (Telegram/Discord/Slack/...)
         ↓
Channel Adapter → Unified Message
         ↓
Channel Manager → Routing Rules
         ↓
Message Handler → Agent Processing
         ↓
Agent Response → Channel Adapter
         ↓
User Reply (Telegram/Discord/Slack/...)
```

## 集成步骤

### 步骤 1: 扩展 core.Application

在 `core/application.go` 中添加多渠道支持：

```go
package core

import (
    "context"
    "AgentFramework/agent"
    "AgentFramework/pkg/channels"
    "AgentFramework/pkg/channels/adapters"
)

type Application struct {
    // ... 现有字段 ...

    // 新增：多渠道管理器
    channelManager *channels.Manager
    channelConfig  *channels.Config
}

// NewApplication 创建应用时初始化多渠道系统
func NewApplication(ctx context.Context, opts ...Option) (*Application, error) {
    app := &Application{
        // ... 现有初始化 ...
    }

    // 初始化多渠道管理器
    if err := app.initChannels(ctx); err != nil {
        return nil, err
    }

    return app, nil
}

// initChannels 初始化多渠道系统
func (a *Application) initChannels(ctx context.Context) error {
    // 创建通道管理器
    manager, err := channels.NewManager(&channels.ManagerConfig{
        EnableMetrics: a.enableMetrics,
        EnableTracing: a.enableTracing,
    })
    if err != nil {
        return err
    }

    a.channelManager = manager

    // 设置消息处理器 - 转发到 Agent 系统
    manager.SetMessageHandler(a.handleChannelMessage)

    // 设置事件处理器
    manager.(*channels.Manager).SetEventHandler(a.handleChannelEvent)

    // 加载配置
    configPath := "config/channels.yaml"
    if config, err := channels.LoadConfig(configPath); err == nil {
        a.channelConfig = config
        a.registerEnabledChannels(ctx)
    }

    return nil
}

// handleChannelMessage 处理来自渠道的消息
func (a *Application) handleChannelMessage(msg *channels.Message) error {
    // 创建用户会话
    userID := getUserID(msg)
    session := a.getOrCreateSession(userID, msg.ChannelID)

    // 将渠道消息转换为 Agent 消息格式
    agentMsg := convertToAgentMessage(msg)

    // 通过 Agent 处理
    response, err := a.host.ProcessMessage(a.ctx, session, agentMsg)
    if err != nil {
        return err
    }

    // 发送响应回渠道
    if response != "" {
        opts := channels.MessageSendOptions{
            ReplyTo: msg.ID,
        }
        _, err := a.channelManager.SendMessage(a.ctx, msg.ChannelID, &channels.Message{
            Type: channels.MessageTypeText,
            Text: response,
        }, opts)
        return err
    }

    return nil
}

// handleChannelEvent 处理渠道事件
func (a *Application) handleChannelEvent(event channels.Event) {
    switch event.Type {
    case channels.EventTypeConnected:
        a.logger.Infof("Channel connected: %s (%s)", event.ChannelID, event.ChannelType)
    case channels.EventTypeDisconnected:
        a.logger.Warnf("Channel disconnected: %s (%s)", event.ChannelID, event.ChannelType)
    case channels.EventTypeError:
        a.logger.Errorf("Channel error: %s: %v", event.ChannelID, event.Error)
    }
}

// StartChannels 启动所有已配置的渠道
func (a *Application) StartChannels(ctx context.Context) error {
    if a.channelManager == nil {
        return nil // 未启用多渠道
    }
    return a.channelManager.Start(ctx)
}

// StopChannels 停止所有渠道
func (a *Application) StopChannels(ctx context.Context) error {
    if a.channelManager == nil {
        return nil
    }
    return a.channelManager.Stop(ctx)
}

// GetChannelStats 获取渠道统计信息
func (a *Application) GetChannelStats(ctx context.Context) (map[string]*channels.ChannelStats, error) {
    if a.channelManager == nil {
        return nil, nil
    }
    return a.channelManager.GetStats(ctx)
}

// SendChannelMessage 发送消息到指定渠道
func (a *Application) SendChannelMessage(ctx context.Context, channelID, text string) error {
    if a.channelManager == nil {
        return fmt.Errorf("channels not enabled")
    }

    msg := &channels.Message{
        Type:    channels.MessageTypeText,
        Text:    text,
    }

    _, err := a.channelManager.SendMessage(ctx, channelID, msg, channels.MessageSendOptions{})
    return err
}

// BroadcastChannelMessage 广播消息到所有渠道
func (a *Application) BroadcastChannelMessage(ctx context.Context, channelType channels.ChannelType, text string) error {
    if a.channelManager == nil {
        return fmt.Errorf("channels not enabled")
    }

    msg := &channels.Message{
        Type:    channels.MessageTypeText,
        Text:    text,
    }

    _, err := a.channelManager.Broadcast(ctx, channelType, msg, channels.MessageSendOptions{})
    return err
}
```

### 步骤 2: 创建 API 接口

创建 `channels_api.go` 提供外部 API：

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "AgentFramework/pkg/channels"
)

// GetChannels 获取所有渠道信息
func (a *App) GetChannels(w http.ResponseWriter, r *http.Request) {
    stats, err := a.core.GetChannelStats(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

// GetChannel 获取单个渠道信息
func (a *App) GetChannel(w http.ResponseWriter, r *http.Request) {
    channelID := r.URL.Query().Get("id")
    if channelID == "" {
        http.Error(w, "channel id required", http.StatusBadRequest)
        return
    }

    stats, _ := a.core.GetChannelStats(r.Context())
    if stats == nil {
        http.Error(w, "channels not enabled", http.StatusNotFound)
        return
    }

    stat, ok := stats[channelID]
    if !ok {
        http.Error(w, "channel not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stat)
}

// SendChannelMessage 发送消息到渠道
func (a *App) SendChannelMessage(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ChannelID string `json:"channel_id"`
        Text      string `json:"text"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := a.core.SendChannelMessage(r.Context(), req.ChannelID, req.Text); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

// BroadcastChannelMessage 广播消息
func (a *App) BroadcastChannelMessage(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ChannelType string `json:"channel_type"`
        Text        string `json:"text"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ct := channels.ChannelType(req.ChannelType)
    if err := a.core.BroadcastChannelMessage(r.Context(), ct, req.Text); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "broadcast"})
}
```

### 步骤 3: 更新路由

在 `app.go` 中添加新的路由：

```go
func (a *App) setupRoutes() {
    // ... 现有路由 ...

    // 多渠道 API 路由
    a.api.HandleFunc("/channels", a.GetChannels).Methods("GET")
    a.api.HandleFunc("/channels/send", a.SendChannelMessage).Methods("POST")
    a.api.HandleFunc("/channels/broadcast", a.BroadcastChannelMessage).Methods("POST")
}
```

## 使用示例

### 示例 1: 启用多渠道支持

```go
package main

import (
    "context"
    "log"
    "AgentFramework/core"
)

func main() {
    // 创建应用（会自动初始化多渠道系统）
    app, err := core.NewApplication(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // 启动渠道
    if err := app.StartChannels(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer app.StopChannels(context.Background())

    // 启动应用
    app.Run()
}
```

### 示例 2: 处理渠道消息

```go
// 处理来自 Telegram 的消息
func (a *Application) handleChannelMessage(msg *channels.Message) error {
    switch msg.ChannelType {
    case channels.ChannelTypeTelegram:
        return a.handleTelegramMessage(msg)
    case channels.ChannelTypeDiscord:
        return a.handleDiscordMessage(msg)
    case channels.ChannelTypeQQ:
        return a.handleQQMessage(msg)
    default:
        return a.handleGenericMessage(msg)
    }
}

func (a *Application) handleTelegramMessage(msg *channels.Message) error {
    // Telegram 特定处理
    if msg.Type == channels.MessageTypeCommand {
        return a.handleCommand(msg)
    }

    // 通用消息处理
    return a.processWithAgent(msg)
}
```

### 示例 3: 发送通知到所有渠道

```go
// 发送系统通知到所有渠道
func (a *Application) SendSystemNotification(message string) error {
    ctx := context.Background()

    // 获取所有启用的渠道类型
    types := []channels.ChannelType{
        channels.ChannelTypeTelegram,
        channels.ChannelTypeDiscord,
        channels.ChannelTypeSlack,
        channels.ChannelTypeFeishu,
        channels.ChannelTypeWeWork,
        channels.ChannelTypeDingTalk,
        channels.ChannelTypeQQ,
    }

    // 广播到所有类型
    for _, ct := range types {
        if err := a.BroadcastChannelMessage(ctx, ct, message); err != nil {
            log.Printf("Failed to broadcast to %s: %v", ct, err)
        }
    }

    return nil
}
```

## 配置管理

### 渠道配置文件

创建 `config/channels.yaml`：

```yaml
version: "1.0"

global:
  enable_metrics: true
  enable_tracing: true
  log_level: info

channels:
  telegram:
    type: telegram
    enabled: true
    token: "${TELEGRAM_BOT_TOKEN}"

  qq:
    type: qq
    enabled: true

routes:
  - id: "accept-all"
    priority: 100
    action: accept
```

### 环境变量

```bash
# .env 文件
TELEGRAM_BOT_TOKEN=your_token
DISCORD_BOT_TOKEN=your_token
QQ_BOT_ENABLED=true
```

## 会话管理

### 用户会话映射

```go
type SessionManager struct {
    // channelID:userID -> Session
    sessions map[string]*Session
    mu       sync.RWMutex
}

func (sm *SessionManager) GetSession(msg *channels.Message) *Session {
    key := fmt.Sprintf("%s:%s", msg.ChannelID, msg.From.ID)

    sm.mu.Lock()
    defer sm.mu.Unlock()

    if sess, exists := sm.sessions[key]; exists {
        return sess
    }

    // 创建新会话
    sess := &Session{
        ID:          generateSessionID(),
        ChannelID:   msg.ChannelID,
        ChannelType: msg.ChannelType,
        UserID:      msg.From.ID,
        CreatedAt:   time.Now(),
    }

    sm.sessions[key] = sess
    return sess
}
```

## 测试

### 集成测试

```go
func TestApplicationWithChannels(t *testing.T) {
    // 创建测试应用
    app, err := core.NewApplication(context.Background())
    require.NoError(t, err)

    // 启动渠道
    err = app.StartChannels(context.Background())
    require.NoError(t, err)
    defer app.StopChannels(context.Background())

    // 获取统计
    stats, err := app.GetChannelStats(context.Background())
    require.NoError(t, err)
    assert.NotNil(t, stats)

    // 发送测试消息
    err = app.SendChannelMessage(context.Background(), "test", "Hello")
    require.NoError(t, err)
}
```

## 故障排查

### 常见问题

**Q: 渠道连接失败**
A: 检查网络连接和 API Token 配置是否正确

**Q: 消息没有响应**
A: 检查消息处理器是否正确设置，查看日志中的错误信息

**Q: 内存使用过高**
A: 启用消息会话清理和压缩功能

## 最佳实践

1. **使用环境变量**：敏感信息（API Token）使用环境变量配置
2. **启用限流**：防止超过平台 API 限制
3. **监控指标**：使用内置的统计功能监控渠道健康状态
4. **优雅关闭**：应用关闭时正确停止所有渠道连接
5. **错误处理**：为每个渠道操作添加适当的错误处理

---

**相关文档:**
- [多渠道系统使用指南](../pkg/channels/README.md)
- [API 文档](./api.md)
- [配置示例](../config/channels.example.yaml)
