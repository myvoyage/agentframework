// Package adapters provides Feishu (Lark) channel adapter
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// Feishu API endpoints
const (
	feishuAPIBase    = "https://open.feishu.cn/open-apis"
	feishuAuthURL    = feishuAPIBase + "/auth/v3/tenant_access_token/internal"
	feishuMessageURL = feishuAPIBase + "/im/v1/messages"
	feishuUploadURL  = feishuAPIBase + "/im/v1/files"
	feishuBotInfoURL = feishuAPIBase + "/bot/v3/info"
)

// FeishuAdapter implements ChannelAdapter for Feishu (Lark)
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for Feishu-specific communication logic
type FeishuAdapter struct {
	*CommonAdapter
	client      *http.Client
	appID       string
	appSecret   string
	tenantToken string
	tokenExpiry time.Time
	config      channels.ChannelConfig
	ctx         context.Context
	cancel      context.CancelFunc
}

// FeishuMessage represents Feishu message format
type FeishuMessage struct {
	ReceiveID     string `json:"receive_id,omitempty"`
	MsgType       string `json:"msg_type"`
	Content       string `json:"content"` // JSON string
	ReceiveIDType string `json:"receive_id_type,omitempty"`
	UUID          string `json:"uuid,omitempty"`
}

// FeishuEvent represents Feishu event callback
type FeishuEvent struct {
	Token     string          `json:"token"`
	Timestamp string          `json:"timestamp"`
	Event     FeishuEventData `json:"event"`
	Type      string          `json:"type"`
}

// FeishuEventData represents event data
type FeishuEventData struct {
	AppID     string `json:"app_id"`
	TenantKey string `json:"tenant_key"`
	Type      string `json:"type"`
}

// FeishuAuthResponse represents auth token response
type FeishuAuthResponse struct {
	Code              int    `json:"code"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

// NewFeishuAdapter creates a new Feishu adapter
func NewFeishuAdapter(channelID string) *FeishuAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeFeishu)
	return &FeishuAdapter{
		CommonAdapter: common,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize initializes the Feishu adapter with configuration
func (a *FeishuAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "FeishuAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.appID = config.AppID
	a.appSecret = config.AppSecret
	a.mu.Unlock()

	// Set capabilities
	a.SetCapability(channels.CapabilityThreads, true)
	a.SetCapability(channels.CapabilityRichText, true)
	a.SetCapability(channels.CapabilityWebhooks, true)

	// Get initial tenant access token
	if err := a.refreshToken(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get initial token: %w", err)
	}

	return nil
}

// Connect establishes connection to Feishu
func (a *FeishuAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "FeishuAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.appID == "" {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	// Verify connection by getting bot info
	if err := a.getBotInfo(ctx); err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	// Start token refresh goroutine
	go a.tokenRefreshLoop(ctx)

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"app_id": a.appID,
	}, nil)

	return nil
}

// Disconnect gracefully closes the Feishu connection
func (a *FeishuAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "FeishuAdapter.Disconnect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	a.SetStatus(context.Background(), channels.ChannelStatusDisconnected, "")
	a.EmitEvent(ctx, channels.EventTypeDisconnected, nil, nil)

	return nil
}

// SendMessage sends a message to Feishu
func (a *FeishuAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "FeishuAdapter.SendMessage",
		trace.WithAttributes(
			attribute.String("message.id", msg.ID),
			attribute.String("message.type", string(msg.Type)),
		),
	)
	defer span.End()

	if !a.IsConnected() {
		err := channels.ErrNotConnected
		span.RecordError(err)
		return "", err
	}

	// Ensure we have valid token
	if err := a.ensureToken(ctx); err != nil {
		span.RecordError(err)
		return "", err
	}

	receiveID := a.extractReceiveID(msg)
	if receiveID == "" {
		err := fmt.Errorf("no valid receive ID found")
		span.RecordError(err)
		return "", err
	}

	// Build Feishu message
	feishuMsg, err := a.buildFeishuMessage(msg, opts)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	feishuMsg.ReceiveID = receiveID
	feishuMsg.ReceiveIDType = "user_id"

	// Marshal message
	body, err := json.Marshal(feishuMsg)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", feishuMessageURL, nil)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+a.tenantToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		span.RecordError(err)
		a.RecordMessageSent(ctx, false, 0)
		a.EmitEvent(ctx, channels.EventTypeMessageFailed, map[string]interface{}{
			"message_id": msg.ID,
			"error":      err.Error(),
		}, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("feishu API returned status %d", resp.StatusCode)
		span.RecordError(err)
		return "", err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return "", err
	}

	a.RecordMessageSent(ctx, true, len(msg.Text))
	a.EmitEvent(ctx, channels.EventTypeMessageSent, map[string]interface{}{
		"message_id": result.Data.MessageID,
		"receive_id": receiveID,
	}, nil)

	return result.Data.MessageID, nil
}

// EditMessage edits an existing message
func (a *FeishuAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "FeishuAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// Feishu requires updating messages via API
	// URL: PATCH /im/v1/messages/{message_id}
	// Implementation similar to SendMessage but with different endpoint
	return fmt.Errorf("message editing not yet implemented for Feishu")
}

// DeleteMessage deletes a message
func (a *FeishuAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// Feishu allows deleting messages via API
	// URL: DELETE /im/v1/messages/{message_id}
	return fmt.Errorf("message deletion not yet implemented for Feishu")
}

// UploadFile uploads a file and returns an attachment
func (a *FeishuAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// Feishu requires multipart upload
	return nil, fmt.Errorf("file upload not yet implemented for Feishu")
}

// HandleWebhook handles webhook callbacks from Feishu
func (a *FeishuAdapter) HandleWebhook(ctx context.Context, event FeishuEvent) error {
	switch event.Type {
	case "im.message.receive_v1":
		return a.handleMessageReceive(ctx, event)
	default:
		return nil // Ignore other event types
	}
}

// refreshToken refreshes the tenant access token
func (a *FeishuAdapter) refreshToken(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	authReq := struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{
		AppID:     a.appID,
		AppSecret: a.appSecret,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", feishuAuthURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var authResp FeishuAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	if authResp.Code != 0 {
		return fmt.Errorf("feishu auth failed with code %d", authResp.Code)
	}

	a.tenantToken = authResp.TenantAccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(authResp.Expire) * time.Second)

	return nil
}

// ensureToken ensures we have a valid token
func (a *FeishuAdapter) ensureToken(ctx context.Context) error {
	a.mu.RLock()
	valid := time.Now().Before(a.tokenExpiry.Add(-5 * time.Minute))
	a.mu.RUnlock()

	if !valid {
		return a.refreshToken(ctx)
	}

	return nil
}

// tokenRefreshLoop periodically refreshes the token
func (a *FeishuAdapter) tokenRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = a.refreshToken(ctx)
		}
	}
}

// getBotInfo verifies the bot connection
func (a *FeishuAdapter) getBotInfo(ctx context.Context) error {
	if err := a.ensureToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feishuBotInfoURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.tenantToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bot info request failed with status %d", resp.StatusCode)
	}

	return nil
}

// buildFeishuMessage builds a Feishu message from unified message
func (a *FeishuAdapter) buildFeishuMessage(msg *channels.Message, opts channels.MessageSendOptions) (*FeishuMessage, error) {
	content := make(map[string]interface{})

	switch msg.Type {
	case channels.MessageTypeText:
		content["text"] = msg.Text
		return &FeishuMessage{
			MsgType: "text",
			Content: mustMarshalJSON(content),
		}, nil

	case channels.MessageTypeImage:
		if len(msg.Attachments) > 0 {
			content["image_key"] = msg.Attachments[0].URL
		}
		return &FeishuMessage{
			MsgType: "image",
			Content: mustMarshalJSON(content),
		}, nil

	default:
		// Default to text with message content
		content["text"] = msg.Text
		return &FeishuMessage{
			MsgType: "text",
			Content: mustMarshalJSON(content),
		}, nil
	}
}

// extractReceiveID extracts receive ID from message
func (a *FeishuAdapter) extractReceiveID(msg *channels.Message) string {
	// Try metadata first
	if receiveID, ok := msg.Metadata["receive_id"]; ok {
		return receiveID
	}

	// Try From field
	if msg.From != nil {
		return msg.From.ChannelUserID
	}

	// Try ChatID
	return msg.ChatID
}

// handleMessageReceive handles message receive event
func (a *FeishuAdapter) handleMessageReceive(ctx context.Context, event FeishuEvent) error {
	// Parse the event and convert to unified message format
	// This would require parsing the event.Event field which contains the actual message
	return nil
}

// mustMarshalJSON marshals to JSON, panics on error
func mustMarshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
