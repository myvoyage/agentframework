// Package channels provides tests for multi-channel system
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ChannelConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid telegram config",
			config: ChannelConfig{
				Type:  ChannelTypeTelegram,
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "invalid telegram - missing token",
			config: ChannelConfig{
				Type: ChannelTypeTelegram,
			},
			wantErr: true,
			errMsg:  "telegram token is required",
		},
		{
			name: "valid discord config",
			config: ChannelConfig{
				Type:  ChannelTypeDiscord,
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid feishu config",
			config: ChannelConfig{
				Type:      ChannelTypeFeishu,
				AppID:     "test-app-id",
				AppSecret: "test-secret",
			},
			wantErr: false,
		},
		{
			name: "invalid feishu - missing credentials",
			config: ChannelConfig{
				Type:  ChannelTypeFeishu,
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "app_id and app_secret are required",
		},
		{
			name: "valid dingtalk config",
			config: ChannelConfig{
				Type:  ChannelTypeDingTalk,
				Token: "test-secret",
				PlatformConfig: map[string]interface{}{
					"app_key": "test-key",
				},
			},
			wantErr: false,
		},
		{
			name: "valid qq config - no required fields",
			config: ChannelConfig{
				Type: ChannelTypeQQ,
			},
			wantErr: false,
		},
		{
			name:    "unsupported channel type",
			config:  ChannelConfig{Type: "unsupported"},
			wantErr: true,
			errMsg:  "unsupported channel type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChannelConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ValidateChannelConfig validates channel configuration (extracted for testing)
func ValidateChannelConfig(config ChannelConfig) error {
	// Simplified validation for testing
	switch config.Type {
	case ChannelTypeTelegram:
		if config.Token == "" {
			return assertError("telegram token is required")
		}
	case ChannelTypeFeishu:
		if config.AppID == "" || config.AppSecret == "" {
			return assertError("app_id and app_secret are required")
		}
	case ChannelTypeDingTalk:
		if config.Token == "" {
			return assertError("app_secret is required")
		}
	case ChannelTypeQQ:
		// No required fields
		return nil
	default:
		return assertError("unsupported channel type")
	}
	return nil
}

func assertError(msg string) error {
	return &ChannelError{Code: "VALIDATION_ERROR", Message: msg}
}

// TestMessageType tests message type constants
func TestMessageType(t *testing.T) {
	tests := []struct {
		MessageType
		expected string
	}{
		{MessageTypeText, "text"},
		{MessageTypeImage, "image"},
		{MessageTypeAudio, "audio"},
		{MessageTypeVideo, "video"},
		{MessageTypeFile, "file"},
		{MessageTypeCommand, "command"},
		{MessageTypeSticker, "sticker"},
		{MessageTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.MessageType))
		})
	}
}

// TestChannelType tests channel type constants
func TestChannelType(t *testing.T) {
	expectedTypes := []string{
		"telegram", "discord", "slack", "feishu",
		"wework", "dingtalk", "qq",
	}

	actualTypes := []ChannelType{
		ChannelTypeTelegram, ChannelTypeDiscord, ChannelTypeSlack,
		ChannelTypeFeishu, ChannelTypeWeWork, ChannelTypeDingTalk,
		ChannelTypeQQ,
	}

	for i, expected := range expectedTypes {
		t.Run(expected, func(t *testing.T) {
			assert.Equal(t, expected, string(actualTypes[i]))
		})
	}
}

// TestMessage tests message structure
func TestMessage(t *testing.T) {
	t.Run("create text message", func(t *testing.T) {
		msg := &Message{
			ID:          "test-id",
			Type:        MessageTypeText,
			Direction:   MessageDirectionIncoming,
			Text:        "Hello, World!",
			ChannelID:   "test-channel",
			ChannelType: ChannelTypeTelegram,
			Timestamp:   time.Now(),
		}

		assert.Equal(t, "test-id", msg.ID)
		assert.Equal(t, MessageTypeText, msg.Type)
		assert.Equal(t, "Hello, World!", msg.Text)
		assert.Equal(t, ChannelTypeTelegram, msg.ChannelType)
	})

	t.Run("create message with attachments", func(t *testing.T) {
		msg := &Message{
			ID:   "test-id",
			Type: MessageTypeImage,
			Attachments: []Attachment{
				{
					ID:       "att-1",
					Filename: "test.jpg",
					URL:      "https://example.com/test.jpg",
					Size:     1024,
					MimeType: "image/jpeg",
				},
			},
		}

		assert.Len(t, msg.Attachments, 1)
		assert.Equal(t, "att-1", msg.Attachments[0].ID)
		assert.Equal(t, "test.jpg", msg.Attachments[0].Filename)
	})
}

// TestUser tests user structure
func TestUser(t *testing.T) {
	t.Run("create user", func(t *testing.T) {
		user := &User{
			ID:            "user-123",
			ChannelUserID: "telegram-123",
			Username:      "testuser",
			DisplayName:   "Test User",
			IsBot:         false,
			ChannelType:   ChannelTypeTelegram,
		}

		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.False(t, user.IsBot)
	})
}

// TestRoutingRule tests routing rule structure
func TestRoutingRule(t *testing.T) {
	t.Run("create accept rule", func(t *testing.T) {
		rule := &RoutingRule{
			ID:       "accept-text",
			Priority: 100,
			Action:   RoutingActionAccept,
			MessageType: []MessageType{
				MessageTypeText,
			},
		}

		assert.Equal(t, "accept-text", rule.ID)
		assert.Equal(t, 100, rule.Priority)
		assert.Equal(t, RoutingActionAccept, rule.Action)
		assert.Len(t, rule.MessageType, 1)
	})

	t.Run("create rate-limited rule", func(t *testing.T) {
		rule := &RoutingRule{
			ID:         "rate-limit",
			Priority:   200,
			Action:     RoutingActionAccept,
			RateLimit:  10,
			RateWindow: 60 * time.Second,
		}

		assert.Equal(t, 10, rule.RateLimit)
		assert.Equal(t, 60*time.Second, rule.RateWindow)
	})
}

// TestChannelStatus tests channel status constants
func TestChannelStatus(t *testing.T) {
	tests := []struct {
		ChannelStatus
		expected string
	}{
		{ChannelStatusDisconnected, "disconnected"},
		{ChannelStatusConnecting, "connecting"},
		{ChannelStatusConnected, "connected"},
		{ChannelStatusReconnecting, "reconnecting"},
		{ChannelStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.ChannelStatus))
		})
	}
}

// TestChannelConfig tests channel configuration
func TestChannelConfig(t *testing.T) {
	t.Run("minimal config", func(t *testing.T) {
		config := ChannelConfig{
			Type:    ChannelTypeTelegram,
			Token:   "test-token",
			Enabled: true,
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, ChannelTypeTelegram, config.Type)
	})

	t.Run("config with platform config", func(t *testing.T) {
		config := ChannelConfig{
			Type:  ChannelTypeQQ,
			Token: "",
			PlatformConfig: map[string]interface{}{
				"api_base": "http://localhost:3000",
				"self_id":  "123456",
			},
		}

		assert.Equal(t, "http://localhost:3000", config.PlatformConfig["api_base"])
	})
}

// TestConfigValidation tests full config validation
func TestConfigValidation(t *testing.T) {
	t.Run("valid global config", func(t *testing.T) {
		global := GlobalConfig{
			DefaultTimeout: 30 * time.Second,
			EnableMetrics:  true,
			EnableTracing:  true,
			LogLevel:       "info",
			MaxMessageSize: 10 * 1024 * 1024,
		}

		assert.Equal(t, 30*time.Second, global.DefaultTimeout)
		assert.True(t, global.EnableMetrics)
	})

	t.Run("config with multiple channels", func(t *testing.T) {
		config := &Config{
			Version: "1.0",
			Global: GlobalConfig{
				DefaultTimeout: 30 * time.Second,
				LogLevel:       "info",
			},
			Channels: map[string]ChannelConfig{
				"telegram": {
					Type:    ChannelTypeTelegram,
					Enabled: true,
					Token:   "test-token",
				},
				"qq": {
					Type:    ChannelTypeQQ,
					Enabled: true,
				},
			},
		}

		assert.Len(t, config.Channels, 2)
		assert.True(t, config.Channels["telegram"].Enabled)
		assert.True(t, config.Channels["qq"].Enabled)
	})
}

// TestMessageSendOptions tests message send options
func TestMessageSendOptions(t *testing.T) {
	t.Run("basic options", func(t *testing.T) {
		opts := MessageSendOptions{
			ParseMode:             "markdown",
			DisableWebPagePreview: true,
			DisableNotification:   false,
		}

		assert.Equal(t, "markdown", opts.ParseMode)
		assert.True(t, opts.DisableWebPagePreview)
	})

	t.Run("reply options", func(t *testing.T) {
		opts := MessageSendOptions{
			ReplyTo: "original-message-id",
		}

		assert.Equal(t, "original-message-id", opts.ReplyTo)
	})
}

// TestEventType tests event type constants
func TestEventType(t *testing.T) {
	events := []struct {
		EventType
		expected string
	}{
		{EventTypeConnected, "connected"},
		{EventTypeDisconnected, "disconnected"},
		{EventTypeReconnecting, "reconnecting"},
		{EventTypeError, "error"},
		{EventTypeMessageReceived, "message_received"},
		{EventTypeMessageSent, "message_sent"},
		{EventTypeMessageFailed, "message_failed"},
	}

	for _, tt := range events {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.EventType))
		})
	}
}

// BenchmarkMessageCreation benchmarks message creation
func BenchmarkMessageCreation(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Message{
			ID:          "test-id",
			Type:        MessageTypeText,
			Direction:   MessageDirectionIncoming,
			Text:        "Hello, World!",
			ChannelID:   "test-channel",
			ChannelType: ChannelTypeTelegram,
			Timestamp:   now,
		}
	}
}

// TestContext tests context handling
func TestContext(t *testing.T) {
	t.Run("context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Simulate operation
		select {
		case <-time.After(100 * time.Millisecond):
			// Operation completed
		case <-ctx.Done():
			t.Fatal("context expired too early")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		// Should be done
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Fatal("context should be cancelled")
		}
	})
}

// TestAttachment tests attachment structure
func TestAttachment(t *testing.T) {
	t.Run("create attachment", func(t *testing.T) {
		att := Attachment{
			ID:       "att-1",
			Filename: "document.pdf",
			URL:      "https://example.com/doc.pdf",
			Size:     2048,
			MimeType: "application/pdf",
		}

		assert.Equal(t, "att-1", att.ID)
		assert.Equal(t, "document.pdf", att.Filename)
		assert.Equal(t, int64(2048), att.Size)
	})

	t.Run("attachment with metadata", func(t *testing.T) {
		att := Attachment{
			ID: "att-2",
			Metadata: map[string]string{
				"width":  "1920",
				"height": "1080",
			},
		}

		assert.Equal(t, "1920", att.Metadata["width"])
		assert.Equal(t, "1080", att.Metadata["height"])
	})
}

// TestMention tests mention structure
func TestMention(t *testing.T) {
	mention := Mention{
		UserID:    "user-123",
		Username:  "testuser",
		Character: '@',
	}

	assert.Equal(t, "user-123", mention.UserID)
	assert.Equal(t, "testuser", mention.Username)
	assert.Equal(t, '@', mention.Character)
}

// TestChannelCapability tests channel capability constants
func TestChannelCapability(t *testing.T) {
	capabilities := []struct {
		ChannelCapability
		expected string
	}{
		{CapabilityThreads, "threads"},
		{CapabilityEdits, "edits"},
		{CapabilityReactions, "reactions"},
		{CapabilityTypingIndicator, "typing_indicator"},
		{CapabilityReadReceipt, "read_receipt"},
		{CapabilityWebhooks, "webhooks"},
		{CapabilityPolling, "polling"},
		{CapabilityRichText, "rich_text"},
		{CapabilityInlineKeyboard, "inline_keyboard"},
	}

	for _, tt := range capabilities {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.ChannelCapability))
		})
	}
}

// TestChannelStats tests channel stats structure
func TestChannelStats(t *testing.T) {
	t.Run("create stats", func(t *testing.T) {
		now := time.Now()
		stats := &ChannelStats{
			ChannelID:        "test-channel",
			ChannelType:      ChannelTypeTelegram,
			Status:           ChannelStatusConnected,
			ConnectedAt:      &now,
			MessagesReceived: 100,
			MessagesSent:     50,
			BytesReceived:    10240,
			BytesSent:        5120,
		}

		assert.Equal(t, "test-channel", stats.ChannelID)
		assert.Equal(t, ChannelStatusConnected, stats.Status)
		assert.Equal(t, int64(100), stats.MessagesReceived)
		assert.Equal(t, int64(50), stats.MessagesSent)
	})
}
