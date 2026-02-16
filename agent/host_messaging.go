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

package agent

import (
	"context"
	"fmt"
	"os"

	"AgentFramework/agent/messaging"
)

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, fields ...interface{}) {
	fmt.Fprintf(os.Stdout, "[INFO] %s %v\n", msg, fields)
}

func (l *defaultLogger) Error(msg string, fields ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s %v\n", msg, fields)
}

func (l *defaultLogger) Debug(msg string, fields ...interface{}) {
	fmt.Fprintf(os.Stdout, "[DEBUG] %s %v\n", msg, fields)
}

func (l *defaultLogger) Warn(msg string, fields ...interface{}) {
	fmt.Fprintf(os.Stdout, "[WARN] %s %v\n", msg, fields)
}

// WithMessaging 配置消息通道
func WithMessaging(enabled bool) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Messaging == nil {
			cfg.Messaging = &MessagingConfig{}
		}
		cfg.Messaging.Enabled = enabled
		return nil
	}
}

// WithMessagingChannel 添加消息通道配置
func WithMessagingChannel(name string, channelType string, config map[string]interface{}) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Messaging == nil {
			cfg.Messaging = &MessagingConfig{}
		}
		if cfg.Messaging.Channels == nil {
			cfg.Messaging.Channels = make(map[string]ChannelConfigSpec)
		}
		cfg.Messaging.Channels[name] = ChannelConfigSpec{
			Type:    channelType,
			Enabled: true,
			Config:  config,
		}
		return nil
	}
}

// WithTelegramChannel 添加 Telegram 通道
func WithTelegramChannel(name, token string) HostOption {
	return WithMessagingChannel(name, "telegram", map[string]interface{}{
		"token": token,
	})
}

// WithSlackChannel 添加 Slack 通道
func WithSlackChannel(name, token string) HostOption {
	return WithMessagingChannel(name, "slack", map[string]interface{}{
		"token": token,
	})
}

// WithDefaultChannel 设置默认通道
func WithDefaultChannel(name string) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Messaging == nil {
			cfg.Messaging = &MessagingConfig{}
		}
		cfg.Messaging.DefaultChannel = name
		return nil
	}
}

// ===== Shutdown 和清理 =====

// Shutdown 优雅关闭 Host 及其所有组件
func (h *Host) Shutdown() error {
	var lastErr error

	// 停止通道管理器
	if h.channelMgr != nil {
		ctx := context.Background()
		if err := h.channelMgr.Stop(ctx); err != nil {
			lastErr = fmt.Errorf("failed to stop channel manager: %w", err)
		}
	}

	// 停止监控管理器
	if h.monitorMgr != nil {
		h.monitorMgr.Stop()
	}

	// 清理插件管理器
	if h.pluginMgr != nil {
		// PluginManager may not have Shutdown method, skip for now
		_ = h.pluginMgr
	}

	// 清理线程存储
	if h.threadStore != nil {
		ctx := context.Background()
		if err := h.threadStore.Close(ctx); err != nil {
			if lastErr == nil {
				lastErr = fmt.Errorf("failed to close thread store: %w", err)
			}
		}
	}

	return lastErr
}

// ===== 消息通道辅助方法 =====

// PublishMessage 发布消息到指定通道
func (h *Host) PublishMessage(channelName string, content string) error {
	if h.channelMgr == nil {
		return fmt.Errorf("channel manager not configured")
	}

	ctx := context.Background()
	msg := messaging.NewChannelMessage(channelName, "", content) // Type will be set by channel

	return h.channelMgr.Publish(ctx, channelName, msg)
}

// BroadcastMessage 广播消息到所有通道
func (h *Host) BroadcastMessage(content string) error {
	if h.channelMgr == nil {
		return fmt.Errorf("channel manager not configured")
	}

	ctx := context.Background()
	msg := messaging.NewChannelMessage("", "", content) // Channel and Type will be set by each channel

	return h.channelMgr.Broadcast(ctx, msg)
}

// SubscribeToChannel 订阅通道消息
func (h *Host) SubscribeToChannel(channelName string, handler messaging.ChannelHandler) error {
	if h.channelMgr == nil {
		return fmt.Errorf("channel manager not configured")
	}

	ctx := context.Background()
	return h.channelMgr.Subscribe(ctx, channelName, handler)
}

// GetChannelStats 获取通道统计信息
func (h *Host) GetChannelStats() map[string]*messaging.ChannelStats {
	if h.channelMgr == nil {
		return nil
	}

	return h.channelMgr.GetAllStats()
}
