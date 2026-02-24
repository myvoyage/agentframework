// Package channels provides a unified multi-channel messaging system
// supporting various platforms like Telegram, Discord, Slack, Feishu, and WeWork.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels

import (
	"fmt"
	"time"
)

// ChannelType represents the type of messaging channel
type ChannelType string

const (
	// ChannelTypeTelegram represents Telegram messaging platform
	ChannelTypeTelegram ChannelType = "telegram"
	// ChannelTypeDiscord represents Discord messaging platform
	ChannelTypeDiscord ChannelType = "discord"
	// ChannelTypeSlack represents Slack messaging platform
	ChannelTypeSlack ChannelType = "slack"
	// ChannelTypeFeishu represents Feishu (Lark) messaging platform
	ChannelTypeFeishu ChannelType = "feishu"
	// ChannelTypeWeWork represents WeCom (WeWork) messaging platform
	ChannelTypeWeWork ChannelType = "wework"
	// ChannelTypeDingTalk represents DingTalk messaging platform
	ChannelTypeDingTalk ChannelType = "dingtalk"
	// ChannelTypeQQ represents QQ messaging platform
	ChannelTypeQQ ChannelType = "qq"
)

// MessageType represents the type of message
type MessageType string

const (
	// MessageTypeText represents a text message
	MessageTypeText MessageType = "text"
	// MessageTypeImage represents an image message
	MessageTypeImage MessageType = "image"
	// MessageTypeAudio represents an audio message
	MessageTypeAudio MessageType = "audio"
	// MessageTypeVideo represents a video message
	MessageTypeVideo MessageType = "video"
	// MessageTypeFile represents a file attachment
	MessageTypeFile MessageType = "file"
	// MessageTypeSticker represents a sticker/emoji message
	MessageTypeSticker MessageType = "sticker"
	// MessageTypeCommand represents a command message (e.g., /help)
	MessageTypeCommand MessageType = "command"
	// MessageTypeSystem represents a system message
	MessageTypeSystem MessageType = "system"
	// MessageTypeUnknown represents an unknown message type
	MessageTypeUnknown MessageType = "unknown"
)

// MessageDirection represents the direction of a message
type MessageDirection string

const (
	// MessageDirectionIncoming represents messages received from users
	MessageDirectionIncoming MessageDirection = "incoming"
	// MessageDirectionOutgoing represents messages sent to users
	MessageDirectionOutgoing MessageDirection = "outgoing"
)

// ChannelStatus represents the status of a channel connection
type ChannelStatus string

const (
	// ChannelStatusDisconnected indicates the channel is disconnected
	ChannelStatusDisconnected ChannelStatus = "disconnected"
	// ChannelStatusConnecting indicates the channel is connecting
	ChannelStatusConnecting ChannelStatus = "connecting"
	// ChannelStatusConnected indicates the channel is connected
	ChannelStatusConnected ChannelStatus = "connected"
	// ChannelStatusReconnecting indicates the channel is reconnecting
	ChannelStatusReconnecting ChannelStatus = "reconnecting"
	// ChannelStatusError indicates the channel has an error
	ChannelStatusError ChannelStatus = "error"
)

// User represents a user from any channel
type User struct {
	ID            string            `json:"id"`
	Username      string            `json:"username,omitempty"`
	DisplayName   string            `json:"display_name"`
	Locale        string            `json:"locale,omitempty"`
	Timezone      string            `json:"timezone,omitempty"`
	IsBot         bool              `json:"is_bot"`
	IsAdmin       bool              `json:"is_admin"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ChannelType   ChannelType       `json:"channel_type"`
	ChannelUserID string            `json:"channel_user_id"` // Platform-specific user ID
}

// Mention represents a user mention in a message
type Mention struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Character rune   `json:"character,omitempty"` // @ for most platforms
}

// Attachment represents a file attachment
type Attachment struct {
	ID           string            `json:"id"`
	Filename     string            `json:"filename"`
	URL          string            `json:"url"`
	Size         int64             `json:"size,omitempty"`
	MimeType     string            `json:"mime_type,omitempty"`
	ThumbnailURL string            `json:"thumbnail_url,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Message represents a unified message format across all channels
type Message struct {
	// Core fields
	ID        string           `json:"id"`
	Type      MessageType      `json:"type"`
	Direction MessageDirection `json:"direction"`

	// Content
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Mentions    []Mention    `json:"mentions,omitempty"`
	ReplyToID   string       `json:"reply_to_id,omitempty"`
	Edited      bool         `json:"edited,omitempty"`

	// Source information
	ChannelID   string      `json:"channel_id"`
	ChannelType ChannelType `json:"channel_type"`
	ChatID      string      `json:"chat_id"`             // Group/channel ID for group chats
	ThreadID    string      `json:"thread_id,omitempty"` // For thread-based messages

	// User information
	From *User   `json:"from,omitempty"`
	To   []*User `json:"to,omitempty"` // For DMs or specific mentions

	// Timestamps
	Timestamp time.Time  `json:"timestamp"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`

	// Platform-specific data
	Metadata map[string]string `json:"metadata,omitempty"`

	// Internal processing
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	TraceID     string     `json:"trace_id,omitempty"` // For distributed tracing
}

// MessageSendOptions represents options when sending a message
type MessageSendOptions struct {
	// ReplyTo specifies the message ID to reply to
	ReplyTo string `json:"reply_to,omitempty"`

	// ParseMode specifies how to parse message formatting (Markdown, HTML, etc.)
	ParseMode string `json:"parse_mode,omitempty"`

	// DisableWebPagePreview disables link previews
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`

	// DisableNotification sends the message silently
	DisableNotification bool `json:"disable_notification,omitempty"`

	// Metadata for platform-specific options
	Metadata map[string]string `json:"metadata,omitempty"`

	// Timeout for sending the message
	Timeout time.Duration `json:"timeout,omitempty"`
}

// ChannelConfig represents configuration for a channel
type ChannelConfig struct {
	// Type is the channel type
	Type ChannelType `json:"type"`

	// Enabled indicates if this channel is enabled
	Enabled bool `json:"enabled"`

	// Name is a friendly name for this channel instance
	Name string `json:"name,omitempty"`

	// Credentials for authentication
	// Token for Telegram/Discord bots
	// BotToken + AppToken for Slack
	// AppID + AppSecret for Feishu/WeWork
	Token      string `json:"token,omitempty"`
	AppToken   string `json:"app_token,omitempty"`
	AppID      string `json:"app_id,omitempty"`
	AppSecret  string `json:"app_secret,omitempty"`
	EncryptKey string `json:"encrypt_key,omitempty"` // For WeWork callback verification

	// Webhook/Callback configuration
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// Connection settings
	UseWebhook      bool          `json:"use_webhook,omitempty"`
	PollingInterval time.Duration `json:"polling_interval,omitempty"`

	// Rate limiting
	RateLimit  int `json:"rate_limit,omitempty"` // Messages per second
	BurstLimit int `json:"burst_limit,omitempty"`

	// Feature flags
	SupportsThreads bool          `json:"supports_threads,omitempty"`
	SupportsEdits   bool          `json:"supports_edits,omitempty"`
	MaxMessageSize  int           `json:"max_message_size,omitempty"`
	AllowedTypes    []MessageType `json:"allowed_types,omitempty"`

	// Platform-specific configuration
	PlatformConfig map[string]interface{} `json:"platform_config,omitempty"`
}

// Validate validates the channel configuration
func (c *ChannelConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("channel type is required")
	}
	if c.Name == "" {
		return fmt.Errorf("channel name is required")
	}
	// Type-specific validation
	switch c.Type {
	case ChannelTypeTelegram, ChannelTypeDiscord:
		if c.Token == "" {
			return fmt.Errorf("token is required for %s", c.Type)
		}
	case ChannelTypeSlack:
		if c.Token == "" || c.AppToken == "" {
			return fmt.Errorf("token and app_token are required for Slack")
		}
	case ChannelTypeFeishu, ChannelTypeWeWork:
		if c.AppID == "" || c.AppSecret == "" {
			return fmt.Errorf("app_id and app_secret are required for %s", c.Type)
		}
	}
	return nil
}

// ChannelStats represents statistics for a channel
type ChannelStats struct {
	ChannelID   string        `json:"channel_id"`
	ChannelType ChannelType   `json:"channel_type"`
	Status      ChannelStatus `json:"status"`

	// Connection stats
	ConnectedAt    *time.Time    `json:"connected_at,omitempty"`
	DisconnectedAt *time.Time    `json:"disconnected_at,omitempty"`
	ReconnectCount int           `json:"reconnect_count"`
	Uptime         time.Duration `json:"uptime"`

	// Message stats
	MessagesReceived int64 `json:"messages_received"`
	MessagesSent     int64 `json:"messages_sent"`
	MessagesFailed   int64 `json:"messages_failed"`
	BytesReceived    int64 `json:"bytes_received"`
	BytesSent        int64 `json:"bytes_sent"`

	// Error stats
	LastError  string `json:"last_error,omitempty"`
	ErrorCount int    `json:"error_count"`

	// User stats
	UniqueUsers int `json:"unique_users"`
	ActiveUsers int `json:"active_users"`
}

// RoutingRule represents a rule for routing messages
type RoutingRule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"` // Higher priority = evaluated first

	// Matching conditions
	ChannelType []ChannelType `json:"channel_type,omitempty"`
	ChannelID   []string      `json:"channel_id,omitempty"`
	ChatID      []string      `json:"chat_id,omitempty"`
	UserID      []string      `json:"user_id,omitempty"`
	MessageType []MessageType `json:"message_type,omitempty"`

	// Pattern matching (regex on message text)
	Pattern string `json:"pattern,omitempty"`

	// Custom predicate function name
	Predicate string `json:"predicate,omitempty"`

	// Action when matched
	Action     RoutingAction     `json:"action"`
	ActionData map[string]string `json:"action_data,omitempty"`

	// Rate limiting for this rule
	RateLimit  int           `json:"rate_limit,omitempty"`
	RateWindow time.Duration `json:"rate_window,omitempty"`

	// Metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RoutingAction represents the action to take when a routing rule matches
type RoutingAction string

const (
	// RoutingActionAccept accepts and processes the message
	RoutingActionAccept RoutingAction = "accept"
	// RoutingActionReject rejects the message without processing
	RoutingActionReject RoutingAction = "reject"
	// RoutingActionForward forwards the message to another channel
	RoutingActionForward RoutingAction = "forward"
	// RoutingActionTransform transforms the message before processing
	RoutingActionTransform RoutingAction = "transform"
	// RoutingActionDelay delays processing of the message
	RoutingActionDelay RoutingAction = "delay"
)

// Event represents a channel event (connection status, errors, etc.)
type Event struct {
	Type        EventType              `json:"type"`
	ChannelID   string                 `json:"channel_id"`
	ChannelType ChannelType            `json:"channel_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       error                  `json:"error,omitempty"`
}

// EventType represents the type of channel event
type EventType string

const (
	// EventTypeConnected is emitted when a channel connects
	EventTypeConnected EventType = "connected"
	// EventTypeDisconnected is emitted when a channel disconnects
	EventTypeDisconnected EventType = "disconnected"
	// EventTypeReconnecting is emitted when a channel is reconnecting
	EventTypeReconnecting EventType = "reconnecting"
	// EventTypeError is emitted when a channel error occurs
	EventTypeError EventType = "error"
	// EventTypeMessageReceived is emitted when a message is received
	EventTypeMessageReceived EventType = "message_received"
	// EventTypeMessageSent is emitted when a message is sent
	EventTypeMessageSent EventType = "message_sent"
	// EventTypeMessageFailed is emitted when sending a message fails
	EventTypeMessageFailed EventType = "message_failed"
)

// EventHandler handles channel events
type EventHandler func(event Event)

// MessageHandler handles incoming messages
type MessageHandler func(msg *Message) error
