// Package adapters provides common utilities and base implementations for channel adapters
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// CommonAdapter provides common functionality for all channel adapters
//
// SOLID - DRY Principle:
// Common functionality implemented once and reused by all adapters
type CommonAdapter struct {
	config         channels.ChannelConfig
	channelID      string
	channelType    channels.ChannelType
	messageHandler channels.MessageHandler
	eventHandler   channels.EventHandler
	status         channels.ChannelStatus
	stats          *channels.ChannelStats
	capabilities   map[channels.ChannelCapability]bool
	mu             sync.RWMutex
	tracer         trace.Tracer
	connectTime    *time.Time
	reconnectCount int
}

// NewCommonAdapter creates a new common adapter instance
func NewCommonAdapter(channelID string, channelType channels.ChannelType) *CommonAdapter {
	return &CommonAdapter{
		channelID:    channelID,
		channelType:  channelType,
		status:       channels.ChannelStatusDisconnected,
		stats:        &channels.ChannelStats{},
		capabilities: make(map[channels.ChannelCapability]bool),
		tracer:       otel.Tracer(fmt.Sprintf("adapter-%s", channelType)),
	}
}

// GetChannelID returns the channel ID
func (a *CommonAdapter) GetChannelID() string {
	return a.channelID
}

// GetChannelType returns the channel type
func (a *CommonAdapter) GetChannelType() channels.ChannelType {
	return a.channelType
}

// SetMessageHandler sets the message handler
func (a *CommonAdapter) SetMessageHandler(handler channels.MessageHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messageHandler = handler
}

// SetEventHandler sets the event handler
func (a *CommonAdapter) SetEventHandler(handler channels.EventHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventHandler = handler
}

// IsConnected returns the connection status
func (a *CommonAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status == channels.ChannelStatusConnected
}

// Supports checks if a capability is supported
func (a *CommonAdapter) Supports(feature string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	cap := channels.ChannelCapability(feature)
	supported, exists := a.capabilities[cap]
	return exists && supported
}

// SetCapability sets a capability
func (a *CommonAdapter) SetCapability(capability channels.ChannelCapability, supported bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.capabilities[capability] = supported
}

// GetStatus returns the current status
func (a *CommonAdapter) GetStatus(ctx context.Context) (channels.ChannelStatus, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status, nil
}

// GetStats returns statistics about the adapter
func (a *CommonAdapter) GetStats(ctx context.Context) (*channels.ChannelStats, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update uptime
	if a.connectTime != nil && a.status == channels.ChannelStatusConnected {
		a.stats.Uptime = time.Since(*a.connectTime)
	}

	a.stats.ChannelID = a.channelID
	a.stats.ChannelType = a.channelType
	a.stats.Status = a.status
	a.stats.ReconnectCount = a.reconnectCount

	snapshot := *a.stats
	return &snapshot, nil
}

// SetStatus updates the adapter status and emits an event
func (a *CommonAdapter) SetStatus(ctx context.Context, status channels.ChannelStatus, errorMsg string) {
	a.mu.Lock()
	oldStatus := a.status
	a.status = status

	// Update stats
	now := time.Now()
	switch status {
	case channels.ChannelStatusConnected:
		if oldStatus != channels.ChannelStatusConnected {
			t := now
			a.connectTime = &t
			a.stats.ConnectedAt = &t
		}
	case channels.ChannelStatusDisconnected, channels.ChannelStatusError:
		if oldStatus == channels.ChannelStatusConnected {
			t := now
			a.stats.DisconnectedAt = &t
		}
		if errorMsg != "" {
			a.stats.LastError = errorMsg
			a.stats.ErrorCount++
		}
	case channels.ChannelStatusReconnecting:
		a.reconnectCount++
	}
	a.mu.Unlock()

	// Emit event
	if a.eventHandler != nil {
		event := channels.Event{
			Type:        channels.EventType(status),
			ChannelID:   a.channelID,
			ChannelType: a.channelType,
			Timestamp:   now,
		}

		if status == channels.ChannelStatusError {
			event.Type = channels.EventTypeError
			if errorMsg != "" {
				if event.Data == nil {
					event.Data = make(map[string]interface{})
				}
				event.Data["error"] = errorMsg
			}
		}

		go a.eventHandler(event)
	}

	// Trace status change
	if status != channels.ChannelStatusConnected {
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("channel.status", string(status)),
			attribute.String("channel.previous_status", string(oldStatus)),
		)
		if errorMsg != "" {
			span.SetStatus(codes.Error, errorMsg)
		}
	}
}

// HandleMessage processes an incoming message through the registered handler
func (a *CommonAdapter) HandleMessage(ctx context.Context, msg *channels.Message) error {
	_, span := a.tracer.Start(ctx, "CommonAdapter.HandleMessage",
		trace.WithAttributes(
			attribute.String("message.id", msg.ID),
			attribute.String("message.type", string(msg.Type)),
		),
	)
	defer span.End()

	a.mu.RLock()
	handler := a.messageHandler
	a.mu.RUnlock()

	if handler == nil {
		return nil // No handler registered, ignore message
	}

	// Set channel info if not already set
	if msg.ChannelID == "" {
		msg.ChannelID = a.channelID
	}
	if msg.ChannelType == "" {
		msg.ChannelType = a.channelType
	}

	// Record stats
	a.mu.Lock()
	a.stats.MessagesReceived++
	a.mu.Unlock()

	// Process message
	if err := handler(msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "message handler failed")

		a.mu.Lock()
		a.stats.MessagesFailed++
		a.mu.Unlock()

		return err
	}

	return nil
}

// EmitEvent emits a channel event
func (a *CommonAdapter) EmitEvent(ctx context.Context, eventType channels.EventType, data map[string]interface{}, err error) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	event := channels.Event{
		Type:        eventType,
		ChannelID:   a.channelID,
		ChannelType: a.channelType,
		Timestamp:   time.Now(),
		Data:        data,
		Error:       err,
	}

	// Emit event in a non-blocking way
	go handler(event)
}

// RecordMessageSent records a sent message in stats
func (a *CommonAdapter) RecordMessageSent(ctx context.Context, success bool, bytes int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if success {
		a.stats.MessagesSent++
		a.stats.BytesSent += int64(bytes)
	} else {
		a.stats.MessagesFailed++
	}
}

// RecordMessageReceived records a received message in stats
func (a *CommonAdapter) RecordMessageReceived(ctx context.Context, bytes int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.MessagesReceived++
	a.stats.BytesReceived += int64(bytes)
}

// ValidateConfig validates channel configuration
func ValidateConfig(config channels.ChannelConfig) error {
	if config.Type == "" {
		return fmt.Errorf("channel type is required")
	}

	// Validate based on channel type
	switch config.Type {
	case channels.ChannelTypeTelegram:
		if config.Token == "" {
			return fmt.Errorf("telegram token is required")
		}
	case channels.ChannelTypeDiscord:
		if config.Token == "" {
			return fmt.Errorf("discord bot token is required")
		}
	case channels.ChannelTypeSlack:
		if config.Token == "" {
			return fmt.Errorf("slack bot token is required")
		}
	case channels.ChannelTypeFeishu:
		if config.AppID == "" || config.AppSecret == "" {
			return fmt.Errorf("feishu app_id and app_secret are required")
		}
	case channels.ChannelTypeWeWork:
		if config.Token == "" {
			return fmt.Errorf("wework token is required")
		}
	case channels.ChannelTypeDingTalk:
		if config.Token == "" {
			return fmt.Errorf("dingtalk app_secret is required")
		}
		if config.PlatformConfig == nil {
			return fmt.Errorf("dingtalk app_key is required in platform_config")
		}
		if _, ok := config.PlatformConfig["app_key"]; !ok {
			return fmt.Errorf("dingtalk app_key is required in platform_config")
		}
	case channels.ChannelTypeQQ:
		// QQ adapter can work with default settings (localhost:3000)
		// No required configuration
	default:
		return fmt.Errorf("unsupported channel type: %s", config.Type)
	}

	return nil
}

// GenerateMessageID generates a unique message ID
func GenerateMessageID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Wait waits for a token to be available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.tryAcquire() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rl.refillRate):
			// Retry
		}
	}
}

// tryAcquire tries to acquire a token
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
