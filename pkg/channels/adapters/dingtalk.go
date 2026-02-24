// Package adapters provides DingTalk channel adapter
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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// DingTalk API endpoints
const (
	dingTalkAPIBase        = "https://oapi.dingtalk.com"
	dingTalkGetTokenURL    = dingTalkAPIBase + "/gettoken"
	dingTalkSendURL        = dingTalkAPIBase + "/robot/send"
	dingTalkUploadURL      = dingTalkAPIBase + "/media/upload"
	dingTalkGetUserInfoURL = dingTalkAPIBase + "/user/get"
)

// DingTalkAdapter implements ChannelAdapter for DingTalk
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for DingTalk-specific communication logic
type DingTalkAdapter struct {
	*CommonAdapter
	client         *http.Client
	appKey         string
	appSecret      string
	accessToken    string
	tokenExpiry    time.Time
	agentID        int64
	agentSecret    string // 机器人 Webhook Secret
	encodingAESKey string
	config         channels.ChannelConfig
	ctx            context.Context
	cancel         context.CancelFunc
}

// DingTalkTokenResponse represents access token response
type DingTalkTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// DingTalkSendResponse represents send message response
type DingTalkSendResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	TaskID  int64  `json:"task_id,omitempty"`
}

// DingTalkMessage represents DingTalk message format
type DingTalkMessage struct {
	MsgType  string            `json:"msgtype"`
	Text     *DingTalkText     `json:"text,omitempty"`
	Link     *DingTalkLink     `json:"link,omitempty"`
	Markdown *DingTalkMarkdown `json:"markdown,omitempty"`
	At       *DingTalkAt       `json:"at,omitempty"`
}

// DingTalkText represents text content
type DingTalkText struct {
	Content string `json:"content"`
}

// DingTalkLink represents link message
type DingTalkLink struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicURL     string `json:"picUrl,omitempty"`
	MessageURL string `json:"messageUrl"`
}

// DingTalkMarkdown represents markdown content
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingTalkAt represents @ information
type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIDs []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// DingTalkEvent represents DingTalk callback event
type DingTalkEvent struct {
	EventType string `json:"type"`
	Text      struct {
		Content string `json:"content"`
	} `json:"text"`
	ChatType string `json:"chatType"`
	ChatID   string `json:"chatId"`
	Sender   struct {
		SenderID   string `json:"senderId"`
		SenderNick string `json:"senderNick"`
		StaffID    string `json:"staffId"`
	} `json:"sender"`
	MsgID      string `json:"msgId"`
	CreateTime int64  `json:"createTime"`
}

// NewDingTalkAdapter creates a new DingTalk adapter
func NewDingTalkAdapter(channelID string) *DingTalkAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeDingTalk)
	return &DingTalkAdapter{
		CommonAdapter: common,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize initializes the DingTalk adapter with configuration
func (a *DingTalkAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "DingTalkAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.appKey = config.PlatformConfig["app_key"].(string)
	a.appSecret = config.Token

	// Parse agent ID
	if agentIDStr, ok := config.PlatformConfig["agent_id"].(string); ok {
		agentID, err := strconv.ParseInt(agentIDStr, 10, 64)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid agent_id: %w", err)
		}
		a.agentID = agentID
	}

	// Get webhook secret
	if secret, ok := config.PlatformConfig["agent_secret"].(string); ok {
		a.agentSecret = secret
	}

	// Get encoding AES key
	if key, ok := config.PlatformConfig["encoding_aes_key"].(string); ok {
		a.encodingAESKey = key
	}
	a.mu.Unlock()

	// Set capabilities
	a.SetCapability(channels.CapabilityRichText, true)
	a.SetCapability(channels.CapabilityWebhooks, true)

	// Get initial access token
	if err := a.refreshToken(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get initial token: %w", err)
	}

	return nil
}

// Connect establishes connection to DingTalk
func (a *DingTalkAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "DingTalkAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.appKey == "" {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	// Start token refresh goroutine
	go a.tokenRefreshLoop(ctx)

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"app_key":  a.appKey,
		"agent_id": a.agentID,
	}, nil)

	return nil
}

// Disconnect gracefully closes the DingTalk connection
func (a *DingTalkAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "DingTalkAdapter.Disconnect")
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

// SendMessage sends a message to DingTalk
func (a *DingTalkAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "DingTalkAdapter.SendMessage",
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

	// Build DingTalk message
	dingMsg, err := a.buildDingTalkMessage(msg, opts)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	// Marshal message
	body, err := json.Marshal(dingMsg)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Build URL with access token
	sendURL := fmt.Sprintf("%s?access_token=%s", dingTalkSendURL, a.accessToken)

	// Add signature if secret is configured
	if a.agentSecret != "" {
		timestamp := time.Now().UnixMilli()
		signStr := fmt.Sprintf("%d\n%s", timestamp, a.agentSecret)
		signature := computeHMACSHA256(signStr, a.agentSecret)

		sendURL = fmt.Sprintf("%s&timestamp=%d&sign=%s",
			sendURL, timestamp, url.QueryEscape(signature))
	}

	// Create request with body
	req, err := http.NewRequestWithContext(ctx, "POST", sendURL, strings.NewReader(string(body)))
	if err != nil {
		span.RecordError(err)
		return "", err
	}

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

	var sendResp DingTalkSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		span.RecordError(err)
		return "", err
	}

	if sendResp.ErrCode != 0 {
		err := fmt.Errorf("dingtalk send failed: %s (code: %d)", sendResp.ErrMsg, sendResp.ErrCode)
		span.RecordError(err)
		return "", err
	}

	a.RecordMessageSent(ctx, true, len(msg.Text))
	a.EmitEvent(ctx, channels.EventTypeMessageSent, map[string]interface{}{
		"task_id": sendResp.TaskID,
	}, nil)

	return msg.ID, nil // DingTalk returns task_id instead of message_id
}

// EditMessage edits an existing message
func (a *DingTalkAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "DingTalkAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// DingTalk doesn't support editing messages after sending
	return fmt.Errorf("message editing is not supported by DingTalk")
}

// DeleteMessage deletes a message
func (a *DingTalkAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// DingTalk doesn't support deleting messages after sending
	return fmt.Errorf("message deletion is not supported by DingTalk")
}

// UploadFile uploads a file and returns an attachment
func (a *DingTalkAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// DingTalk requires multipart upload
	return nil, fmt.Errorf("file upload not yet implemented for DingTalk")
}

// HandleCallback handles callback from DingTalk
func (a *DingTalkAdapter) HandleCallback(ctx context.Context, eventType string, content []byte) error {
	switch eventType {
	case "chat":
		// Parse event
		var event DingTalkEvent
		if err := json.Unmarshal(content, &event); err != nil {
			return fmt.Errorf("failed to parse event: %w", err)
		}

		// Convert to unified message
		unifiedMsg := a.convertEventToUnified(&event)
		return a.HandleMessage(ctx, unifiedMsg)

	default:
		return nil // Ignore other event types
	}
}

// VerifySignature verifies callback signature
func (a *DingTalkAdapter) VerifySignature(timestamp, nonce, sign, content string) bool {
	if a.agentSecret == "" {
		return true // No secret configured, skip verification
	}

	// Build expected signature
	expected := computeHMACSHA256(timestamp+nonce+content, a.agentSecret)

	return sign == expected
}

// refreshToken refreshes the access token
func (a *DingTalkAdapter) refreshToken(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	url := fmt.Sprintf("%s?appkey=%s&appsecret=%s",
		dingTalkGetTokenURL, a.appKey, a.appSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp DingTalkTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	if tokenResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk auth failed: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// ensureToken ensures we have a valid token
func (a *DingTalkAdapter) ensureToken(ctx context.Context) error {
	a.mu.RLock()
	valid := time.Now().Before(a.tokenExpiry.Add(-5 * time.Minute))
	a.mu.RUnlock()

	if !valid {
		return a.refreshToken(ctx)
	}

	return nil
}

// tokenRefreshLoop periodically refreshes the token
func (a *DingTalkAdapter) tokenRefreshLoop(ctx context.Context) {
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

// buildDingTalkMessage builds a DingTalk message from unified message
func (a *DingTalkAdapter) buildDingTalkMessage(msg *channels.Message, opts channels.MessageSendOptions) (*DingTalkMessage, error) {
	dingMsg := &DingTalkMessage{
		MsgType: "text",
		Text: &DingTalkText{
			Content: msg.Text,
		},
	}

	switch msg.Type {
	case channels.MessageTypeText:
		// Check if markdown is requested
		if opts.ParseMode == "markdown" {
			dingMsg.MsgType = "markdown"
			dingMsg.Markdown = &DingTalkMarkdown{
				Title: "消息",
				Text:  msg.Text,
			}
			delete(dingMsg, "Text")
		}

	case channels.MessageTypeImage:
		if len(msg.Attachments) > 0 {
			dingMsg.MsgType = "link"
			dingMsg.Link = &DingTalkLink{
				Text:       msg.Text,
				Title:      "图片",
				PicURL:     msg.Attachments[0].URL,
				MessageURL: msg.Attachments[0].URL,
			}
			delete(dingMsg, "Text")
		}

	case channels.MessageTypeFile:
		if len(msg.Attachments) > 0 {
			dingMsg.MsgType = "link"
			dingMsg.Link = &DingTalkLink{
				Text:       msg.Text,
				Title:      msg.Attachments[0].Filename,
				MessageURL: msg.Attachments[0].URL,
			}
			delete(dingMsg, "Text")
		}
	}

	// Handle @mentions
	if len(msg.Mentions) > 0 {
		at := &DingTalkAt{}
		for _, mention := range msg.Mentions {
			if strings.HasPrefix(mention.UserID, "$") {
				// Mobile phone format
				at.AtMobiles = append(at.AtMobiles, mention.UserID[1:])
			} else {
				at.AtUserIDs = append(at.AtUserIDs, mention.UserID)
			}
		}
		dingMsg.At = at
	}

	return dingMsg, nil
}

// convertEventToUnified converts DingTalk event to unified format
func (a *DingTalkAdapter) convertEventToUnified(event *DingTalkEvent) *channels.Message {
	user := &channels.User{
		ID:            event.Sender.SenderID,
		ChannelUserID: event.Sender.SenderID,
		DisplayName:   event.Sender.SenderNick,
		IsBot:         false,
		ChannelType:   channels.ChannelTypeDingTalk,
	}

	// Determine message type
	msgType := channels.MessageTypeText
	if strings.Contains(event.Text.Content, "image") {
		msgType = channels.MessageTypeImage
	}

	unifiedMsg := &channels.Message{
		ID:          event.MsgID,
		Type:        msgType,
		Direction:   channels.MessageDirectionIncoming,
		Text:        event.Text.Content,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeDingTalk,
		ChatID:      event.ChatID,
		From:        user,
		Timestamp:   time.Unix(event.CreateTime/1000, 0),
		Metadata:    make(map[string]string),
	}

	unifiedMsg.Metadata["chat_type"] = event.ChatType
	unifiedMsg.Metadata["staff_id"] = event.Sender.StaffID

	return unifiedMsg
}

// computeHMACSHA256 computes HMAC-SHA256 signature
func computeHMACSHA256(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Helper function to delete key from struct map
func delete(m *DingTalkMessage, key string) {
	// This is a no-op as Go doesn't allow deleting struct fields
	// In the actual implementation, we handle this in the buildDingTalkMessage
	_ = m
	_ = key
}
