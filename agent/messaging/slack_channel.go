// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package messaging

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/slack-go/slack"
)

// SlackChannel Slack 通道实现
// 使用 Slack Web API 进行消息发送和接收
type SlackChannel struct {
	name       string
	client     *slack.Client
	config     SlackConfig
	handlers   []ChannelHandler
	mu         sync.RWMutex
	running    bool
	stats      ChannelStats
	ctx        context.Context
	cancel     context.CancelFunc
	rtm        *slack.RTM
}

// SlackConfig Slack 配置
type SlackConfig struct {
	Token            string `json:"token"`                        // Bot Token
	AppLevelToken    string `json:"app_level_token,omitempty"`    // App-Level Token（可选）
	SigningSecret    string `json:"signing_secret,omitempty"`     // Signing Secret（用于 Webhook）
	MessageTypes     []string `json:"message_types,omitempty"`    // 要监听的消息类型
}

// NewSlackChannel 创建新的 Slack 通道
func NewSlackChannel(name string, config SlackConfig) (*SlackChannel, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("slack token is required")
	}

	// 创建 Slack 客户端
	client := slack.New(config.Token)

	return &SlackChannel{
		name:     name,
		client:   client,
		config:   config,
		handlers: make([]ChannelHandler, 0),
		stats:    ChannelStats{},
	}, nil
}

// Start 启动 Slack 通道
func (ch *SlackChannel) Start(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.running {
		return ErrChannelAlreadyStarted{Name: ch.name}
	}

	ch.ctx, ch.cancel = context.WithCancel(ctx)

	// 创建 RTM 连接
	rtm := ch.client.NewRTM()
	ch.rtm = rtm

	// 启动 RTM
	go ch.manageRTM()

	ch.running = true
	return nil
}

// Stop 停止 Slack 通道
func (ch *SlackChannel) Stop(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.running {
		return nil
	}

	// 断开 RTM
	if ch.rtm != nil {
		ch.rtm.Disconnect()
	}

	// 取消上下文
	if ch.cancel != nil {
		ch.cancel()
	}

	ch.running = false
	return nil
}

// IsRunning 检查通道是否正在运行
func (ch *SlackChannel) IsRunning() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.running
}

// Name 返回通道名称
func (ch *SlackChannel) Name() string {
	return ch.name
}

// Type 返回通道类型
func (ch *SlackChannel) Type() ChannelType {
	return ChannelTypeSlack
}

// Publish 发布消息到 Slack
func (ch *SlackChannel) Publish(ctx context.Context, msg *ChannelMessage) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	startTime := time.Now()

	// 确定目标频道
	channelID, ok := msg.Metadata["channel_id"].(string)
	if !ok {
		channelID = msg.Target
	}

	if channelID == "" {
		ch.stats.RecordMessage(false, time.Since(startTime))
		return ErrPublishFailed{
			Channel: ch.name,
			Reason:  "target channel not specified",
		}
	}

	// 发送消息
	_, _, err := ch.client.PostMessage(
		channelID,
		slack.MsgOptionText(msg.Content, false),
		slack.MsgOptionAsUser(true),
	)

	if err != nil {
		ch.stats.RecordMessage(false, time.Since(startTime))
		return ErrPublishFailed{
			Channel: ch.name,
			Reason:  err.Error(),
		}
	}

	ch.stats.RecordMessage(true, time.Since(startTime))
	return nil
}

// Subscribe 订阅 Slack 消息
func (ch *SlackChannel) Subscribe(ctx context.Context, handler ChannelHandler) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.handlers = append(ch.handlers, handler)
	return nil
}

// Unsubscribe 取消订阅
func (ch *SlackChannel) Unsubscribe(ctx context.Context, handler ChannelHandler) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	for i, h := range ch.handlers {
		if &h == &handler {
			ch.handlers = append(ch.handlers[:i], ch.handlers[i+1:]...)
			break
		}
	}

	return nil
}

// HealthCheck 健康检查
func (ch *SlackChannel) HealthCheck(ctx context.Context) error {
	if !ch.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// 测试 API 连接
	_, err := ch.client.AuthTest()
	if err != nil {
		return fmt.Errorf("slack auth test failed: %w", err)
	}

	return nil
}

// GetStats 获取统计信息
func (ch *SlackChannel) GetStats() *ChannelStats {
	snapshot := ch.stats.GetSnapshot()
	return &snapshot
}

// GetClient 获取 Slack 客户端实例（用于高级用法）
func (ch *SlackChannel) GetClient() *slack.Client {
	return ch.client
}

// manageRTM 管理 RTM 连接和消息
func (ch *SlackChannel) manageRTM() {
	for {
		select {
		case <-ch.ctx.Done():
			return
		case msg, ok := <-ch.rtm.IncomingEvents:
			if !ok {
				return
			}

			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				ch.handleMessage(ev)
			case *slack.RTMError:
				// 处理 RTM 错误
			case *slack.ConnectionErrorEvent:
				// 处理连接错误
			}
		}
	}
}

// handleMessage 处理 Slack 消息
func (ch *SlackChannel) handleMessage(ev *slack.MessageEvent) {
	// 忽略 Bot 发送的消息
	if ev.BotID != "" || ev.SubType != "" {
		return
	}

	// 获取用户信息
	user, err := ch.client.GetUserInfo(ev.User)
	if err != nil {
		return
	}

	// 转换时间戳（Slack 的 Timestamp 是字符串类型）
	var timestamp time.Time
	if ev.Timestamp != "" {
		if ts, err := strconv.ParseInt(ev.Timestamp, 10, 64); err == nil {
			timestamp = time.Unix(ts, 0)
		} else {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	// 转换为 ChannelMessage
	channelMsg := &ChannelMessage{
		ID:        generateMessageID(),
		Channel:   ch.name,
		Type:      ChannelTypeSlack,
		Content:   ev.Text,
		Metadata:  make(map[string]interface{}),
		Timestamp: timestamp,
		Source:    user.Name,
		Target:    ev.Channel,
		ReplyTo:   "",
		RawData:   ev,
	}

	// 添加额外元数据
	channelMsg.Metadata["channel_id"] = ev.Channel
	channelMsg.Metadata["user_id"] = ev.User
	channelMsg.Metadata["timestamp"] = ev.Timestamp
	channelMsg.Metadata["thread_ts"] = ev.ThreadTimestamp

	// 调用所有处理器
	for _, handler := range ch.handlers {
		if err := handler(ch.ctx, channelMsg); err != nil {
			ch.stats.RecordMessage(false, 0)
		} else {
			ch.stats.RecordMessage(true, 0)
		}
	}
}
