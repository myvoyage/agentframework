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

package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ChannelType defines the type of communication channel
type ChannelType string

const (
	ChannelTypeTelegram   ChannelType = "telegram"
	ChannelTypeWeChat     ChannelType = "wechat"     // 企业微信
	ChannelTypeLark      ChannelType = "lark"      // 飞书
	ChannelTypeQQ         ChannelType = "qq"        // QQ
	ChannelTypeDiscord    ChannelType = "discord"
	ChannelTypeSlack      ChannelType = "slack"
	ChannelTypeWhatsApp   ChannelType = "whatsapp"
	ChannelTypeWebChat    ChannelType = "webchat"
	ChannelTypeCLI        ChannelType = "cli"
)

// ChannelConfig contains channel-specific configuration
type ChannelConfig struct {
	Type      ChannelType `yaml:"type"`
	Enabled   bool        `yaml:"enabled"`
	Name     string      `yaml:"name"`
	Token    string      `yaml:"token"`      // API token / Bot token
	Secret   string      `yaml:"secret"`    // App Secret
	Webhook  string      `yaml:"webhook"`   // Webhook URL for receiving messages
	Endpoint string      `yaml:"endpoint"`  // Custom endpoint
}

// Channel represents a communication channel adapter
type Channel interface {
	// Type returns the channel type
	Type() ChannelType

	// Name returns the channel name
	Name() string

	// Start starts the channel listener
	Start(ctx context.Context) error

	// Stop stops the channel listener
	Stop(ctx context.Context) error

	// Send sends a message to a user or channel
	Send(ctx context.Context, to string, message *Message) error

	// OnMessage registers a message handler
	OnMessage(handler MessageHandler)
}

// MessageHandler is called when a message is received
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents a chat message
type Message struct {
	ID        string            `json:"id"`
	Channel   ChannelType       `json:"channel"`
	From      *User             `json:"from"`
	To        string            `json:"to"`
	Content   string            `json:"content"`
	Type      MessageType       `json:"type"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp int64            `json:"timestamp"`

	// Extended fields for multi-channel support
	ChatID    string   `json:"chat_id,omitempty"`    // Group/channel ID
	ChatType  string   `json:"chat_type,omitempty"`  // p2p or group
	ImageKey  string   `json:"image_key,omitempty"`  // Image key for Lark
	Title     string   `json:"title,omitempty"`      // Card title
	Actions   []Action `json:"actions,omitempty"`     // Interactive actions
}

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeFile     MessageType = "file"
	MessageTypeCommand  MessageType = "command"
	MessageTypeAction   MessageType = "action"
	MessageTypeCard     MessageType = "card"    // Interactive card
	MessageTypeRichText MessageType = "post"   // Rich text (Lark post)
)

// Action represents an interactive action
type Action struct {
	Label string `json:"label"`
	URL   string `json:"url,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// User represents a chat user
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// ChannelManager manages multiple channel adapters
type ChannelManager struct {
	mu       sync.RWMutex
	channels map[ChannelType]Channel
	config   *ChannelManagerConfig
}

// ChannelManagerConfig contains channel manager configuration
type ChannelManagerConfig struct {
	Workspace   string
	ChannelsDir string
}

// NewChannelManager creates a new channel manager
func NewChannelManager(workspace string) *ChannelManager {
	return &ChannelManager{
		channels: make(map[ChannelType]Channel),
		config: &ChannelManagerConfig{
			Workspace:   workspace,
			ChannelsDir: filepath.Join(workspace, ".channels"),
		},
	}
}

// LoadChannels loads channel configurations
func (m *ChannelManager) LoadChannels() error {
	// Ensure channels directory exists
	if err := os.MkdirAll(m.config.ChannelsDir, 0755); err != nil {
		return err
	}

	// Load channel configs from files
	files, err := os.ReadDir(m.config.ChannelsDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		path := filepath.Join(m.config.ChannelsDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg ChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		if cfg.Enabled {
			channel, err := m.createChannel(&cfg)
			if err != nil {
				continue
			}
			m.channels[cfg.Type] = channel
		}
	}

	return nil
}

// createChannel creates a channel adapter
func (m *ChannelManager) createChannel(cfg *ChannelConfig) (Channel, error) {
	switch cfg.Type {
	case ChannelTypeTelegram:
		return NewTelegramChannel(cfg), nil
	case ChannelTypeWeChat:
		return NewWeChatChannel(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", cfg.Type)
	}
}

// RegisterChannel registers a channel adapter
func (m *ChannelManager) RegisterChannel(channel Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[channel.Type()] = channel
}

// UnregisterChannel unregisters a channel
func (m *ChannelManager) UnregisterChannel(channelType ChannelType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.channels, channelType)
}

// GetChannel returns a channel by type
func (m *ChannelManager) GetChannel(channelType ChannelType) Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.channels[channelType]
}

// ListChannels returns all registered channels
func (m *ChannelManager) ListChannels() []Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]Channel, 0, len(m.channels))
	for _, ch := range m.channels {
		channels = append(channels, ch)
	}
	return channels
}

// StartAll starts all registered channels
func (m *ChannelManager) StartAll(ctx context.Context) error {
	channels := m.ListChannels()

	for _, ch := range channels {
		if err := ch.Start(ctx); err != nil {
			return fmt.Errorf("failed to start %s: %w", ch.Type(), err)
		}
	}

	return nil
}

// StopAll stops all channels
func (m *ChannelManager) StopAll(ctx context.Context) error {
	for _, ch := range m.ListChannels() {
		if err := ch.Stop(ctx); err != nil {
			// Log but continue
			continue
		}
	}
	return nil
}

// SaveChannelConfig saves a channel configuration
func (m *ChannelManager) SaveChannelConfig(cfg *ChannelConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.json", cfg.Type)
	path := filepath.Join(m.config.ChannelsDir, filename)
	return os.WriteFile(path, data, 0644)
}

// TelegramChannel is a Telegram bot adapter
type TelegramChannel struct {
	config *ChannelConfig
	bot    string
}

// NewTelegramChannel creates a Telegram channel adapter
func NewTelegramChannel(cfg *ChannelConfig) *TelegramChannel {
	return &TelegramChannel{
		config: cfg,
		bot:    cfg.Token,
	}
}

// Type returns channel type
func (c *TelegramChannel) Type() ChannelType {
	return ChannelTypeTelegram
}

// Name returns channel name
func (c *TelegramChannel) Name() string {
	return c.config.Name
}

// Start starts the Telegram bot
func (c *TelegramChannel) Start(ctx context.Context) error {
	// TODO: Implement Telegram bot webhook/polling
	return nil
}

// Stop stops the Telegram bot
func (c *TelegramChannel) Stop(ctx context.Context) error {
	return nil
}

// Send sends a message via Telegram
func (c *TelegramChannel) Send(ctx context.Context, to string, message *Message) error {
	// TODO: Implement Telegram API call
	return nil
}

// OnMessage registers a message handler
func (c *TelegramChannel) OnMessage(handler MessageHandler) {
	// TODO: Implement message handler registration
}

// WeChatChannel is an enterprise WeChat adapter
type WeChatChannel struct {
	config *ChannelConfig
	corpID string
	agentID string
}

// NewWeChatChannel creates a WeChat channel adapter
func NewWeChatChannel(cfg *ChannelConfig) *WeChatChannel {
	return &WeChatChannel{
		config: cfg,
	}
}

// Type returns channel type
func (c *WeChatChannel) Type() ChannelType {
	return ChannelTypeWeChat
}

// Name returns channel name
func (c *WeChatChannel) Name() string {
	return c.config.Name
}

// Start starts the WeChat webhook listener
func (c *WeChatChannel) Start(ctx context.Context) error {
	// TODO: Implement WeChat webhook server
	return nil
}

// Stop stops the WeChat webhook listener
func (c *WeChatChannel) Stop(ctx context.Context) error {
	return nil
}

// Send sends a message via WeChat
func (c *WeChatChannel) Send(ctx context.Context, to string, message *Message) error {
	// TODO: Implement WeChat API call
	return nil
}

// OnMessage registers a message handler
func (c *WeChatChannel) OnMessage(handler MessageHandler) {
	// TODO: Implement message handler registration
}
