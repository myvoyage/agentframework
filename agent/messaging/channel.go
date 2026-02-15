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
	"sync"
	"time"
)

// ChannelType 定义通道类型
type ChannelType string

const (
	ChannelTypeInternal ChannelType = "internal" // 内部 Agent 通信
	ChannelTypeTelegram ChannelType = "telegram" // Telegram
	ChannelTypeSlack    ChannelType = "slack"    // Slack
	ChannelTypeDiscord  ChannelType = "discord"  // Discord
	ChannelTypeWebhook  ChannelType = "webhook"  // Webhook
	ChannelTypeCustom   ChannelType = "custom"   // 自定义通道
)

// String 返回通道类型的字符串表示
func (ct ChannelType) String() string {
	return string(ct)
}

// Channel 消息通道接口
// 定义了所有消息通道必须实现的基本操作
type Channel interface {
	// 基本信息
	Name() string
	Type() ChannelType

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 消息操作
	Publish(ctx context.Context, msg *ChannelMessage) error
	Subscribe(ctx context.Context, handler ChannelHandler) error
	Unsubscribe(ctx context.Context, handler ChannelHandler) error

	// 健康检查
	HealthCheck(ctx context.Context) error

	// 统计信息
	GetStats() *ChannelStats
}

// ChannelMessage 通道消息（标准化格式）
// 所有通道都使用统一的消息格式
type ChannelMessage struct {
	ID        string                 `json:"id"`                  // 消息唯一 ID
	Channel   string                 `json:"channel"`             // 通道名称
	Type      ChannelType            `json:"type"`                // 通道类型
	Content   string                 `json:"content"`             // 消息内容
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // 元数据
	Timestamp time.Time              `json:"timestamp"`           // 时间戳

	// 来源和目标
	Source    string `json:"source"`              // 发送者 ID/用户名
	Target    string `json:"target,omitempty"`    // 接收者 ID/频道名（空表示广播）
	ReplyTo   string `json:"reply_to,omitempty"`  // 回复消息 ID

	// 内部字段
	RawData    interface{}        `json:"raw_data,omitempty"`    // 原始数据（平台特定）
	ReplyChan  chan *ChannelMessage `json:"-"`                     // 回复通道（用于请求-响应模式）
}

// ChannelHandler 通道消息处理器
// 定义了处理消息的函数签名
type ChannelHandler func(ctx context.Context, msg *ChannelMessage) error

// ChannelStats 通道统计信息
type ChannelStats struct {
	TotalMessages   int64         `json:"total_messages"`    // 总消息数
	FailedMessages  int64         `json:"failed_messages"`   // 失败消息数
	AverageLatency  time.Duration `json:"average_latency"`   // 平均延迟
	IsHealthy       bool          `json:"is_healthy"`        // 是否健康
	LastMessageTime time.Time     `json:"last_message_time"` // 最后消息时间
	mu              sync.RWMutex  `json:"-"`                  // 读写锁（不序列化）
}

// RecordMessage 记录一条消息（更新统计信息）
func (s *ChannelStats) RecordMessage(success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalMessages++
	if !success {
		s.FailedMessages++
	}

	// 更新平均延迟（简单移动平均）
	if s.AverageLatency == 0 {
		s.AverageLatency = latency
	} else {
		s.AverageLatency = (s.AverageLatency + latency) / 2
	}

	s.LastMessageTime = time.Now()
	s.IsHealthy = s.calculateHealth()
}

// calculateHealth 计算健康状态
func (s *ChannelStats) calculateHealth() bool {
	// 如果失败率超过 10%，认为不健康
	if s.TotalMessages > 10 {
		failureRate := float64(s.FailedMessages) / float64(s.TotalMessages)
		if failureRate > 0.1 {
			return false
		}
	}

	// 如果最后消息时间超过 5 分钟，认为可能不健康
	if !s.LastMessageTime.IsZero() && time.Since(s.LastMessageTime) > 5*time.Minute {
		// 但如果是刚启动，不认为是问题
		if s.TotalMessages > 0 {
			return false
		}
	}

	return true
}

// GetSnapshot 获取统计信息的快照（线程安全）
func (s *ChannelStats) GetSnapshot() ChannelStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ChannelStats{
		TotalMessages:   s.TotalMessages,
		FailedMessages:  s.FailedMessages,
		AverageLatency:  s.AverageLatency,
		IsHealthy:       s.IsHealthy,
		LastMessageTime: s.LastMessageTime,
	}
}

// ChannelConfig 通道配置
type ChannelConfig struct {
	Name       string                 `json:"name"`                  // 通道名称
	Type       ChannelType            `json:"type"`                  // 通道类型
	Enabled    bool                   `json:"enabled"`               // 是否启用
	Config     map[string]interface{} `json:"config,omitempty"`      // 平台特定配置
	BufferSize int                    `json:"buffer_size,omitempty"` // 消息缓冲区大小
}

// Validate 验证通道配置
func (c *ChannelConfig) Validate() error {
	if c.Name == "" {
		return ErrInvalidConfig{Field: "name", Reason: "name cannot be empty"}
	}

	if c.Type == "" {
		return ErrInvalidConfig{Field: "type", Reason: "type cannot be empty"}
	}

	return nil
}

// ===== 错误类型 =====

// ErrInvalidConfig 无效配置错误
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid config: " + e.Field + " - " + e.Reason
}

// ErrChannelNotStarted 通道未启动错误
type ErrChannelNotStarted struct {
	Name string
}

func (e ErrChannelNotStarted) Error() string {
	return "channel not started: " + e.Name
}

// ErrChannelAlreadyStarted 通道已启动错误
type ErrChannelAlreadyStarted struct {
	Name string
}

func (e ErrChannelAlreadyStarted) Error() string {
	return "channel already started: " + e.Name
}

// ErrPublishFailed 发布失败错误
type ErrPublishFailed struct {
	Channel string
	Reason  string
}

func (e ErrPublishFailed) Error() string {
	return "publish failed on channel " + e.Channel + ": " + e.Reason
}

// ===== 辅助函数 =====

// NewChannelMessage 创建新的通道消息
func NewChannelMessage(channel string, msgType ChannelType, content string) *ChannelMessage {
	return &ChannelMessage{
		ID:        generateMessageID(),
		Channel:   channel,
		Type:      msgType,
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// NewChannelMessageWithSource 创建带来源的通道消息
func NewChannelMessageWithSource(channel string, msgType ChannelType, content, source string) *ChannelMessage {
	msg := NewChannelMessage(channel, msgType, content)
	msg.Source = source
	return msg
}

// generateMessageID 生成消息 ID
func generateMessageID() string {
	// 使用时间戳 + 随机数生成唯一 ID
	// 在实际生产环境中可以使用 UUID
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
