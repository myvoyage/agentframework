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

// MessagingConfig 消息通道配置
type MessagingConfig struct {
	Enabled        bool                        `yaml:"enabled"`                   // 是否启用消息通道
	EnableMetrics  bool                        `yaml:"enableMetrics,omitempty"`  // 是否启用指标收集
	Channels       map[string]ChannelConfigSpec `yaml:"channels,omitempty"`      // 通道配置
	DefaultChannel string                      `yaml:"defaultChannel,omitempty"`  // 默认通道
}

// ChannelConfigSpec 通道配置规范
type ChannelConfigSpec struct {
	Type       string                 `yaml:"type"`                  // 通道类型: internal, telegram, slack, discord, webhook
	Enabled    bool                   `yaml:"enabled"`               // 是否启用
	Config     map[string]interface{} `yaml:"config,omitempty"`     // 通道特定配置
	BufferSize int                    `yaml:"bufferSize,omitempty"` // 缓冲区大小
}

// TelegramConfig Telegram 特定配置
type TelegramConfig struct {
	Token      string `yaml:"token"`                  // Bot Token
	WebhookURL string `yaml:"webhook_url,omitempty"` // Webhook URL
}

// SlackConfig Slack 特定配置
type SlackConfig struct {
	Token         string `yaml:"token"`                    // Bot Token
	AppLevelToken string `yaml:"app_level_token,omitempty"` // App-Level Token
	SigningSecret string `yaml:"signing_secret,omitempty"`  // Signing Secret
}

// DiscordConfig Discord 特定配置
type DiscordConfig struct {
	Token      string `yaml:"token"` // Bot Token
	GuildID    string `yaml:"guild_id,omitempty"`    // Guild ID
	ChannelID  string `yaml:"channel_id,omitempty"`  // Default Channel ID
}

// WebhookConfig Webhook 特定配置
type WebhookConfig struct {
	URL         string            `yaml:"url"`                   // Webhook URL
	Headers     map[string]string `yaml:"headers,omitempty"`     // HTTP Headers
	Secret      string            `yaml:"secret,omitempty"`      // HMAC Secret
	TimeoutSec  int               `yaml:"timeout_sec,omitempty"` // Timeout in seconds
}
