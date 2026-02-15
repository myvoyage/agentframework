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
	"sync"
	"time"

	tb "gopkg.in/telebot.v3"
)

// TelegramChannel Telegram 通道实现
// 使用 Telegram Bot API 进行消息发送和接收
type TelegramChannel struct {
	name       string
	bot        *tb.Bot
	config     TelegramConfig
	handlers   []ChannelHandler
	mu         sync.RWMutex
	running    bool
	stats      ChannelStats
	ctx        context.Context
	cancel     context.CancelFunc
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	Token      string `json:"token"`                 // Bot Token
	WebhookURL string `json:"webhook_url,omitempty"` // Webhook URL（可选）
	Polling    bool   `json:"polling"`               // 是否使用长轮询（默认 true）
}

// NewTelegramChannel 创建新的 Telegram 通道
func NewTelegramChannel(name string, config TelegramConfig) (*TelegramChannel, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("telegram token is required")
	}

	// 创建 Telegram Bot
	bot, err := tb.NewBot(tb.Settings{
		Token:  config.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &TelegramChannel{
		name:     name,
		bot:      bot,
		config:   config,
		handlers: make([]ChannelHandler, 0),
		stats:    ChannelStats{},
	}, nil
}

// Start 启动 Telegram 通道
func (ch *TelegramChannel) Start(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.running {
		return ErrChannelAlreadyStarted{Name: ch.name}
	}

	ch.ctx, ch.cancel = context.WithCancel(ctx)

	// 设置消息处理
	ch.bot.Handle(tb.OnText, func(m tb.Context) error {
		msg := m.Message()
		if msg == nil {
			return nil
		}

		// 转换为 ChannelMessage
		channelMsg := &ChannelMessage{
			ID:        generateMessageID(),
			Channel:   ch.name,
			Type:      ChannelTypeTelegram,
			Content:   msg.Text,
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now(),
			Source:    fmt.Sprintf("%d", msg.Sender.ID),
			Target:    "",
			ReplyTo:   "",
			RawData:   msg,
		}

		// 调用所有处理器
		for _, handler := range ch.handlers {
			if err := handler(ch.ctx, channelMsg); err != nil {
				ch.stats.RecordMessage(false, 0)
			} else {
				ch.stats.RecordMessage(true, 0)
			}
		}

		return nil
	})

	// 启动 Bot
	go ch.bot.Start()

	ch.running = true
	return nil
}

// Stop 停止 Telegram 通道
func (ch *TelegramChannel) Stop(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.running {
		return nil
	}

	// 停止 Bot
	ch.bot.Stop()

	// 取消上下文
	if ch.cancel != nil {
		ch.cancel()
	}

	ch.running = false
	return nil
}

// IsRunning 检查通道是否正在运行
func (ch *TelegramChannel) IsRunning() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.running
}

// Name 返回通道名称
func (ch *TelegramChannel) Name() string {
	return ch.name
}

// Type 返回通道类型
func (ch *TelegramChannel) Type() ChannelType {
	return ChannelTypeTelegram
}

// Publish 发布消息到 Telegram
func (ch *TelegramChannel) Publish(ctx context.Context, msg *ChannelMessage) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	startTime := time.Now()

	// 确定目标聊天 ID
	chatID, ok := msg.Metadata["chat_id"].(int64)
	if !ok {
		// 尝试从 Target 字段获取
		if msg.Target != "" {
			// 这里可以添加解析逻辑
			chatID = 0 // 默认值
		}
	}

	// 发送消息
	_, err := ch.bot.Send(&tb.User{
		ID: chatID,
	}, msg.Content)

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

// Subscribe 订阅 Telegram 消息
func (ch *TelegramChannel) Subscribe(ctx context.Context, handler ChannelHandler) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.handlers = append(ch.handlers, handler)
	return nil
}

// Unsubscribe 取消订阅
func (ch *TelegramChannel) Unsubscribe(ctx context.Context, handler ChannelHandler) error {
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
func (ch *TelegramChannel) HealthCheck(ctx context.Context) error {
	if !ch.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// 可以添加 Bot API 健康检查
	// 例如调用 getMe 方法
	return nil
}

// GetStats 获取统计信息
func (ch *TelegramChannel) GetStats() *ChannelStats {
	return ch.stats.GetSnapshot()
}

// GetBot 获取 Telegram Bot 实例（用于高级用法）
func (ch *TelegramChannel) GetBot() *tb.Bot {
	return ch.bot
}
