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

	"AgentFramework/agent/collaboration"
)

// InternalChannel 内部通道适配器
// 将现有的 MessageBus 适配为 Channel 接口
type InternalChannel struct {
	name       string
	messageBus *collaboration.MessageBus
	eventBus   EventBus
	handlers   map[string][]ChannelHandler
	mu         sync.RWMutex
	running    bool
	stats      ChannelStats
	config     ChannelConfig
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewInternalChannel 创建新的内部通道
func NewInternalChannel(name string, config ChannelConfig) (*InternalChannel, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Type == "" {
		config.Type = ChannelTypeInternal
	}
	if config.BufferSize == 0 {
		config.BufferSize = 100
	}

	// 创建 MessageBus
	messageBus := collaboration.NewMessageBus()

	// EventBus 需要从外部提供（通过 SetEventBus）
	return &InternalChannel{
		name:       name,
		messageBus: messageBus,
		eventBus:   nil, // 将通过 SetEventBus 设置
		handlers:   make(map[string][]ChannelHandler),
		config:     config,
		stats:      ChannelStats{},
	}, nil
}

// Start 启动内部通道
func (ch *InternalChannel) Start(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.running {
		return ErrChannelAlreadyStarted{Name: ch.name}
	}

	// 启动 MessageBus
	if err := ch.messageBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start message bus: %w", err)
	}

	// 启动 EventBus
	if err := ch.eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// 创建上下文
	ch.ctx, ch.cancel = context.WithCancel(ctx)
	ch.running = true

	return nil
}

// Stop 停止内部通道
func (ch *InternalChannel) Stop(ctx context.Context) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.running {
		return nil
	}

	// 取消上下文
	if ch.cancel != nil {
		ch.cancel()
	}

	// 停止 MessageBus
	ch.messageBus.Stop(ctx)

	// 停止 EventBus
	ch.eventBus.Stop()

	ch.running = false
	return nil
}

// IsRunning 检查通道是否正在运行
func (ch *InternalChannel) IsRunning() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.running
}

// Name 返回通道名称
func (ch *InternalChannel) Name() string {
	return ch.name
}

// Type 返回通道类型
func (ch *InternalChannel) Type() ChannelType {
	return ChannelTypeInternal
}

// Publish 发布消息到通道
func (ch *InternalChannel) Publish(ctx context.Context, msg *ChannelMessage) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	startTime := time.Now()

	// 转换为内部 Message 格式
	internalMsg := &collaboration.Message{
		ID:        msg.ID,
		Type:      mapMessageType(msg.Type),
		From:      msg.Source,
		To:        msg.Target,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
		Metadata:  msg.Metadata,
		ReplyChan: msg.ReplyChan,
	}

	// 发布到 MessageBus
	if err := ch.messageBus.Send(ctx, internalMsg); err != nil {
		ch.stats.RecordMessage(false, time.Since(startTime))
		return ErrPublishFailed{
			Channel: ch.name,
			Reason:  err.Error(),
		}
	}

	// 发布到 EventBus（用于事件分发）
	event := &Event{
		Topic:   "message." + msg.Channel,
		Payload: msg,
	}

	if ch.eventBus != nil {
		_ = ch.eventBus.Publish(ctx, event)
	}

	ch.stats.RecordMessage(true, time.Since(startTime))
	return nil
}

// Subscribe 订阅通道消息
func (ch *InternalChannel) Subscribe(ctx context.Context, handler ChannelHandler) error {
	if !ch.IsRunning() {
		return ErrChannelNotStarted{Name: ch.name}
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 添加到处理器列表
	ch.handlers[ctx] = append(ch.handlers[ctx], handler)

	// 在 MessageBus 上订阅
	subChan := make(chan *collaboration.Message, 10)
	if err := ch.messageBus.Subscribe(ctx, "all", subChan); err != nil {
		return fmt.Errorf("failed to subscribe to message bus: %w", err)
	}

	// 启动监听协程
	go ch.listen(ctx, subChan, handler)

	return nil
}

// Unsubscribe 取消订阅
func (ch *InternalChannel) Unsubscribe(ctx context.Context, handler ChannelHandler) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 从处理器列表中移除
	for key, handlers := range ch.handlers {
		for i, h := range handlers {
			if &h == &handler {
				ch.handlers[key] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}

	// MessageBus 的订阅会在 context 取消时自动清理
	return nil
}

// listen 监听消息并调用处理器
func (ch *InternalChannel) listen(ctx context.Context, subChan <-chan *collaboration.Message, handler ChannelHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ch.ctx.Done():
			return
		case msg, ok := <-subChan:
			if !ok {
				return
			}

			// 转换为 ChannelMessage
			channelMsg := &ChannelMessage{
				ID:        msg.ID,
				Channel:   ch.name,
				Type:      mapMessageTypeToChannel(msg.Type),
				Content:   msg.Content,
				Metadata:  msg.Metadata,
				Timestamp: msg.Timestamp,
				Source:    msg.From,
				Target:    msg.To,
				ReplyTo:   "",
				ReplyChan: msg.ReplyChan,
			}

			// 调用处理器
			if err := handler(ctx, channelMsg); err != nil {
				// 记录错误但继续运行
				ch.stats.RecordMessage(false, 0)
			} else {
				ch.stats.RecordMessage(true, 0)
			}
		}
	}
}

// HealthCheck 健康检查
func (ch *InternalChannel) HealthCheck(ctx context.Context) error {
	if !ch.IsRunning() {
		return fmt.Errorf("channel not running")
	}

	// 检查 MessageBus 状态
	if ch.messageBus == nil {
		return fmt.Errorf("message bus not initialized")
	}

	// 检查 EventBus 状态
	if ch.eventBus == nil {
		return fmt.Errorf("event bus not initialized")
	}

	return nil
}

// GetStats 获取统计信息
func (ch *InternalChannel) GetStats() *ChannelStats {
	return ch.stats.GetSnapshot()
}

// GetMessageBus 获取底层的 MessageBus（用于高级用法）
func (ch *InternalChannel) GetMessageBus() *collaboration.MessageBus {
	return ch.messageBus
}

// SetEventBus 设置 EventBus（用于事件分发）
func (ch *InternalChannel) SetEventBus(eventBus EventBus) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.eventBus = eventBus
}

// GetEventBus 获取底层的 EventBus（用于高级用法）
func (ch *InternalChannel) GetEventBus() EventBus {
	return ch.eventBus
}

// ===== 辅助函数 =====

// mapMessageType 将 ChannelType 映射到 MessageType
func mapMessageType(ct ChannelType) collaboration.MessageType {
	switch ct {
	case ChannelTypeInternal:
		return collaboration.MessageTypeNotification
	case ChannelTypeTelegram, ChannelTypeSlack, ChannelTypeDiscord:
		return collaboration.MessageTypeBroadcast
	default:
		return collaboration.MessageTypeNotification
	}
}

// mapMessageTypeToChannel 将 MessageType 映射到 ChannelType
func mapMessageTypeToChannel(mt collaboration.MessageType) ChannelType {
	switch mt {
	case collaboration.MessageTypeTask:
		return ChannelTypeInternal
	case collaboration.MessageTypeResult:
		return ChannelTypeInternal
	case collaboration.MessageTypeError:
		return ChannelTypeInternal
	case collaboration.MessageTypeQuery:
		return ChannelTypeInternal
	case collaboration.MessageTypeNotification:
		return ChannelTypeInternal
	case collaboration.MessageTypeBroadcast:
		return ChannelTypeInternal
	default:
		return ChannelTypeInternal
	}
}
