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
)

// EventBus 事件总线接口（解耦循环依赖）
type EventBus interface {
	Start() error
	Stop()
	Publish(ctx context.Context, event *Event) error
	Subscribe(topic string, handler EventHandler) string
	Unsubscribe(subID string)
}

// Event 事件结构
type Event struct {
	Topic   string
	Payload interface{}
}

// EventHandler 事件处理器函数类型
type EventHandler func(ctx context.Context, event *Event) error

// ChannelManager 通道管理器
// 负责管理多个消息通道，提供统一的发布订阅接口
type ChannelManager struct {
	channels map[string]Channel
	eventBus EventBus
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	logger   Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, fields)
}

// ChannelManagerConfig 通道管理器配置
type ChannelManagerConfig struct {
	Logger        Logger
	EventBus      EventBus
	EnableMetrics bool
}

// NewChannelManager 创建新的通道管理器
func NewChannelManager(config ChannelManagerConfig) (*ChannelManager, error) {
	// 设置默认值
	if config.Logger == nil {
		config.Logger = &DefaultLogger{}
	}

	// EventBus 必须从外部提供，避免循环依赖
	if config.EventBus == nil {
		return nil, fmt.Errorf("EventBus is required, please provide one")
	}

	return &ChannelManager{
		channels: make(map[string]Channel),
		eventBus: config.EventBus,
		logger:   config.Logger,
	}, nil
}

// Start 启动通道管理器
func (cm *ChannelManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.running {
		return fmt.Errorf("channel manager already started")
	}

	// 启动 EventBus
	if err := cm.eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	cm.ctx, cm.cancel = context.WithCancel(ctx)
	cm.running = true

	// 启动所有已注册的通道
	for name, ch := range cm.channels {
		if err := ch.Start(ctx); err != nil {
			cm.logger.Error(fmt.Sprintf("failed to start channel %s", name), "error", err)
		} else {
			cm.logger.Info(fmt.Sprintf("started channel %s", name))
		}
	}

	cm.logger.Info("channel manager started")
	return nil
}

// Stop 停止通道管理器
func (cm *ChannelManager) Stop(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return nil
	}

	// 停止所有通道
	for name, ch := range cm.channels {
		if err := ch.Stop(ctx); err != nil {
			cm.logger.Error(fmt.Sprintf("failed to stop channel %s", name), "error", err)
		}
	}

	// 停止 EventBus
	cm.eventBus.Stop()

	// 取消上下文
	if cm.cancel != nil {
		cm.cancel()
	}

	cm.running = false
	cm.logger.Info("channel manager stopped")
	return nil
}

// RegisterChannel 注册通道
func (cm *ChannelManager) RegisterChannel(ch Channel) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	name := ch.Name()
	if _, exists := cm.channels[name]; exists {
		return fmt.Errorf("channel %s already registered", name)
	}

	cm.channels[name] = ch
	cm.logger.Info(fmt.Sprintf("registered channel %s (type: %s)", name, ch.Type()))

	// 如果管理器已经运行，自动启动通道
	if cm.running {
		if err := ch.Start(cm.ctx); err != nil {
			delete(cm.channels, name)
			return fmt.Errorf("failed to start channel %s: %w", name, err)
		}
	}

	return nil
}

// UnregisterChannel 注销通道
func (cm *ChannelManager) UnregisterChannel(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ch, exists := cm.channels[name]
	if !exists {
		return fmt.Errorf("channel %s not found", name)
	}

	// 停止通道
	if err := ch.Stop(cm.ctx); err != nil {
		cm.logger.Error(fmt.Sprintf("failed to stop channel %s", name), "error", err)
	}

	delete(cm.channels, name)
	cm.logger.Info(fmt.Sprintf("unregistered channel %s", name))
	return nil
}

// GetChannel 获取通道
func (cm *ChannelManager) GetChannel(name string) (Channel, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ch, exists := cm.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %s not found", name)
	}

	return ch, nil
}

// ListChannels 列出所有通道
func (cm *ChannelManager) ListChannels() []Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]Channel, 0, len(cm.channels))
	for _, ch := range cm.channels {
		channels = append(channels, ch)
	}

	return channels
}

// GetChannelsByType 按类型获取通道
func (cm *ChannelManager) GetChannelsByType(channelType ChannelType) []Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]Channel, 0)
	for _, ch := range cm.channels {
		if ch.Type() == channelType {
			channels = append(channels, ch)
		}
	}

	return channels
}

// Broadcast 广播消息到所有通道
func (cm *ChannelManager) Broadcast(ctx context.Context, msg *ChannelMessage) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.running {
		return fmt.Errorf("channel manager not running")
	}

	var lastErr error
	for _, ch := range cm.channels {
		// 只发送到目标通道（如果指定了）
		if msg.Channel != "" && msg.Channel != ch.Name() {
			continue
		}

		// 复制消息以避免修改原始消息
		msgCopy := *msg
		msgCopy.Channel = ch.Name()
		msgCopy.Type = ch.Type()

		if err := ch.Publish(ctx, &msgCopy); err != nil {
			cm.logger.Error(fmt.Sprintf("failed to publish to channel %s", ch.Name()), "error", err)
			lastErr = err
		}
	}

	return lastErr
}

// Publish 发布消息到指定通道
func (cm *ChannelManager) Publish(ctx context.Context, channelName string, msg *ChannelMessage) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.running {
		return fmt.Errorf("channel manager not running")
	}

	ch, exists := cm.channels[channelName]
	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	// 设置通道信息
	msg.Channel = channelName
	msg.Type = ch.Type()

	return ch.Publish(ctx, msg)
}

// Subscribe 订阅通道消息
func (cm *ChannelManager) Subscribe(ctx context.Context, channelName string, handler ChannelHandler) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.running {
		return fmt.Errorf("channel manager not running")
	}

	ch, exists := cm.channels[channelName]
	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	return ch.Subscribe(ctx, handler)
}

// SubscribeAll 订阅所有通道的消息
func (cm *ChannelManager) SubscribeAll(ctx context.Context, handler ChannelHandler) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.running {
		return fmt.Errorf("channel manager not running")
	}

	var lastErr error
	for _, ch := range cm.channels {
		if err := ch.Subscribe(ctx, handler); err != nil {
			cm.logger.Error(fmt.Sprintf("failed to subscribe to channel %s", ch.Name()), "error", err)
			lastErr = err
		}
	}

	return lastErr
}

// HealthCheck 检查所有通道的健康状态
func (cm *ChannelManager) HealthCheck(ctx context.Context) map[string]error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	results := make(map[string]error)
	for name, ch := range cm.channels {
		results[name] = ch.HealthCheck(ctx)
	}

	return results
}

// GetAllStats 获取所有通道的统计信息
func (cm *ChannelManager) GetAllStats() map[string]*ChannelStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make(map[string]*ChannelStats)
	for name, ch := range cm.channels {
		stats[name] = ch.GetStats()
	}

	return stats
}

// GetRunningChannels 获取正在运行的通道列表
func (cm *ChannelManager) GetRunningChannels() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	running := make([]string, 0)
	for name, ch := range cm.channels {
		if ch.IsRunning() {
			running = append(running, name)
		}
	}

	return running
}

// IsRunning 检查管理器是否正在运行
func (cm *ChannelManager) IsRunning() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.running
}

// GetEventBus 获取 EventBus（用于高级用法）
func (cm *ChannelManager) GetEventBus() EventBus {
	return cm.eventBus
}

// ===== 辅助方法 =====

// publishEvent 发布事件到 EventBus
func (cm *ChannelManager) publishEvent(ctx context.Context, topic string, payload interface{}) {
	event := &Event{
		Topic:   topic,
		Payload: payload,
	}

	_ = cm.eventBus.Publish(ctx, event)
}
