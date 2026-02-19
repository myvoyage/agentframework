// Package adapters provides WeWork (WeCom) channel adapter
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
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// WeWork API endpoints
const (
	weworkAPIBase     = "https://qyapi.weixin.qq.com/cgi-bin"
	weworkGetTokenURL = weworkAPIBase + "/gettoken"
	weworkSendURL     = weworkAPIBase + "/message/send"
	weworkUploadURL   = weworkAPIBase + "/media/upload"
)

// WeWorkAdapter implements ChannelAdapter for WeWork (WeCom)
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for WeWork-specific communication logic
type WeWorkAdapter struct {
	*CommonAdapter
	client         *http.Client
	corpID         string
	corpSecret     string
	agentID        int
	accessToken    string
	tokenExpiry    time.Time
	encodingAESKey string
	token          string
	config         channels.ChannelConfig
	ctx            context.Context
	cancel         context.CancelFunc
}

// WeWorkTokenResponse represents access token response
type WeWorkTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// WeWorkMessage represents WeWork message format
type WeWorkMessage struct {
	ToUser  string       `json:"touser,omitempty"`
	ToParty string       `json:"toparty,omitempty"`
	ToTag   string       `json:"totag,omitempty"`
	MsgType string       `json:"msgtype"`
	AgentID int          `json:"agentid"`
	Text    *WeWorkText  `json:"text,omitempty"`
	Image   *WeWorkMedia `json:"image,omitempty"`
	Voice   *WeWorkMedia `json:"voice,omitempty"`
	Video   *WeWorkMedia `json:"video,omitempty"`
	File    *WeWorkMedia `json:"file,omitempty"`
}

// WeWorkText represents text content
type WeWorkText struct {
	Content string `json:"content"`
}

// WeWorkMedia represents media content
type WeWorkMedia struct {
	MediaID string `json:"media_id"`
}

// WeWorkXMLMessage represents XML message from callback
type WeWorkXMLMessage struct {
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string
	MsgID        int64
	AgentID      int
}

// WeWorkSendResponse represents send message response
type WeWorkSendResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	InvalidUser  string `json:"invaliduser,omitempty"`
	InvalidParty string `json:"invalidparty,omitempty"`
	InvalidTag   string `json:"invalidtag,omitempty"`
}

// NewWeWorkAdapter creates a new WeWork adapter
func NewWeWorkAdapter(channelID string) *WeWorkAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeWeWork)
	return &WeWorkAdapter{
		CommonAdapter: common,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize initializes the WeWork adapter with configuration
func (a *WeWorkAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "WeWorkAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.corpID = config.PlatformConfig["corp_id"].(string)
	a.corpSecret = config.Token
	a.encodingAESKey = config.PlatformConfig["encoding_aes_key"].(string)
	a.token = config.PlatformConfig["token"].(string)

	// Parse agent ID
	if agentIDStr, ok := config.PlatformConfig["agent_id"].(string); ok {
		agentID, err := strconv.Atoi(agentIDStr)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid agent_id: %w", err)
		}
		a.agentID = agentID
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

// Connect establishes connection to WeWork
func (a *WeWorkAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "WeWorkAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.corpID == "" {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	// Start token refresh goroutine
	go a.tokenRefreshLoop(ctx)

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"corp_id":  a.corpID,
		"agent_id": a.agentID,
	}, nil)

	return nil
}

// Disconnect gracefully closes the WeWork connection
func (a *WeWorkAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "WeWorkAdapter.Disconnect")
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

// SendMessage sends a message to WeWork
func (a *WeWorkAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "WeWorkAdapter.SendMessage",
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

	toUser := a.extractToUser(msg)
	if toUser == "" {
		err := fmt.Errorf("no valid user ID found")
		span.RecordError(err)
		return "", err
	}

	// Build WeWork message
	weworkMsg, err := a.buildWeWorkMessage(msg, toUser)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	// Marshal message
	body, err := json.Marshal(weworkMsg)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create request URL
	url := fmt.Sprintf("%s?access_token=%s", weworkSendURL, a.accessToken)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
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

	var sendResp WeWorkSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		span.RecordError(err)
		return "", err
	}

	if sendResp.ErrCode != 0 {
		err := fmt.Errorf("wework send failed: %s (code: %d)", sendResp.ErrMsg, sendResp.ErrCode)
		span.RecordError(err)
		return "", err
	}

	a.RecordMessageSent(ctx, true, len(msg.Text))
	a.EmitEvent(ctx, channels.EventTypeMessageSent, map[string]interface{}{
		"to_user":  toUser,
		"agent_id": a.agentID,
	}, nil)

	return msg.ID, nil // WeWork doesn't return message ID
}

// EditMessage edits an existing message
func (a *WeWorkAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "WeWorkAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// WeWork doesn't support editing messages after sending
	return fmt.Errorf("message editing not supported by WeWork")
}

// DeleteMessage deletes a message
func (a *WeWorkAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// WeWork doesn't support deleting messages after sending
	return fmt.Errorf("message deletion not supported by WeWork")
}

// UploadFile uploads a file and returns an attachment
func (a *WeWorkAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// WeWork requires multipart upload to get media_id
	return nil, fmt.Errorf("file upload not yet implemented for WeWork")
}

// HandleCallback handles callback from WeWork
func (a *WeWorkAdapter) HandleCallback(ctx context.Context, msgType, content string) error {
	switch msgType {
	case "text":
		// Parse XML message
		var xmlMsg WeWorkXMLMessage
		if err := xml.Unmarshal([]byte(content), &xmlMsg); err != nil {
			return fmt.Errorf("failed to parse XML: %w", err)
		}

		// Convert to unified message
		unifiedMsg := a.convertXMLToUnified(&xmlMsg)
		return a.HandleMessage(ctx, unifiedMsg)

	default:
		return nil // Ignore other message types
	}
}

// VerifySignature verifies callback signature
func (a *WeWorkAdapter) VerifySignature(timestamp, nonce, echoStr, signature string) bool {
	// Sort parameters
	params := []string{a.token, timestamp, nonce}
	// Simple sort (for production, use proper sorting)
	for i := 0; i < len(params); i++ {
		for j := i + 1; j < len(params); j++ {
			if params[i] > params[j] {
				params[i], params[j] = params[j], params[i]
			}
		}
	}

	// Concatenate
	str := strings.Join(params, "")

	// HMAC-SHA1
	h := hmac.New(sha1.New, []byte(a.encodingAESKey))
	h.Write([]byte(str))
	expectedSig := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature == expectedSig
}

// refreshToken refreshes the access token
func (a *WeWorkAdapter) refreshToken(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s",
		weworkGetTokenURL, a.corpID, a.corpSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp WeWorkTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	if tokenResp.ErrCode != 0 {
		return fmt.Errorf("wework auth failed: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// ensureToken ensures we have a valid token
func (a *WeWorkAdapter) ensureToken(ctx context.Context) error {
	a.mu.RLock()
	valid := time.Now().Before(a.tokenExpiry.Add(-5 * time.Minute))
	a.mu.RUnlock()

	if !valid {
		return a.refreshToken(ctx)
	}

	return nil
}

// tokenRefreshLoop periodically refreshes the token
func (a *WeWorkAdapter) tokenRefreshLoop(ctx context.Context) {
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

// buildWeWorkMessage builds a WeWork message from unified message
func (a *WeWorkAdapter) buildWeWorkMessage(msg *channels.Message, toUser string) (*WeWorkMessage, error) {
	weworkMsg := &WeWorkMessage{
		ToUser:  toUser,
		AgentID: a.agentID,
	}

	switch msg.Type {
	case channels.MessageTypeText:
		weworkMsg.MsgType = "text"
		weworkMsg.Text = &WeWorkText{Content: msg.Text}

	case channels.MessageTypeImage:
		weworkMsg.MsgType = "image"
		if len(msg.Attachments) > 0 {
			// Use attachment URL as media_id (needs to be uploaded first)
			weworkMsg.Image = &WeWorkMedia{MediaID: msg.Attachments[0].URL}
		}

	case channels.MessageTypeAudio:
		weworkMsg.MsgType = "voice"
		if len(msg.Attachments) > 0 {
			weworkMsg.Voice = &WeWorkMedia{MediaID: msg.Attachments[0].URL}
		}

	case channels.MessageTypeVideo:
		weworkMsg.MsgType = "video"
		if len(msg.Attachments) > 0 {
			weworkMsg.Video = &WeWorkMedia{MediaID: msg.Attachments[0].URL}
		}

	case channels.MessageTypeFile:
		weworkMsg.MsgType = "file"
		if len(msg.Attachments) > 0 {
			weworkMsg.File = &WeWorkMedia{MediaID: msg.Attachments[0].URL}
		}

	default:
		// Default to text
		weworkMsg.MsgType = "text"
		weworkMsg.Text = &WeWorkText{Content: msg.Text}
	}

	return weworkMsg, nil
}

// extractToUser extracts to_user from message
func (a *WeWorkAdapter) extractToUser(msg *channels.Message) string {
	// Try metadata first
	if toUser, ok := msg.Metadata["to_user"]; ok {
		return toUser
	}

	// Try From field (for reply)
	if msg.From != nil {
		return msg.From.ChannelUserID
	}

	// Try ChatID
	return msg.ChatID
}

// convertXMLToUnified converts WeWork XML message to unified format
func (a *WeWorkAdapter) convertXMLToUnified(xmlMsg *WeWorkXMLMessage) *channels.Message {
	user := &channels.User{
		ID:            xmlMsg.FromUserName,
		ChannelUserID: xmlMsg.FromUserName,
		DisplayName:   xmlMsg.FromUserName,
		IsBot:         false,
		ChannelType:   channels.ChannelTypeWeWork,
	}

	unifiedMsg := &channels.Message{
		ID:          strconv.FormatInt(xmlMsg.MsgID, 10),
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionIncoming,
		Text:        xmlMsg.Content,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeWeWork,
		ChatID:      xmlMsg.FromUserName,
		From:        user,
		Timestamp:   time.Unix(xmlMsg.CreateTime, 0),
		Metadata:    make(map[string]string),
	}

	return unifiedMsg
}
