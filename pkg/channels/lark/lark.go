// Package lark provides Feishu/Lark channel implementation
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package lark implements the Feishu/Lark channel based on official OpenClaw plugin
// Reference: https://docs.openclaw.ai/zh-CN/channels/feishu
// GitHub: https://github.com/larksuite/openclaw-lark

package lark

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// =============================================================================
// Constants
// =============================================================================

// API Endpoints
const (
	// 国内版 API
	FeishuAPIBase       = "https://open.feishu.cn/open-apis"
	FeishuAuthURL       = FeishuAPIBase + "/auth/v3/tenant_access_token/internal"
	FeishuSendMessageURL = FeishuAPIBase + "/im/v1/messages"
	FeishuUploadURL     = FeishuAPIBase + "/im/v1/files"
	FeishuBotInfoURL    = FeishuAPIBase + "/bot/v3/info"
	FeishuUserInfoURL   = FeishuAPIBase + "/contact/v3/users"

	// 国际版 API (Lark)
	LarkAPIBase         = "https://open.larksuite.com/open-apis"
)

// Event Types (WebSocket)
const (
	EventMessageReceive   = "im.message.receive_v1"
	EventMessageEdit      = "im.message.message_edit_v1"
	EventP2PChatCreate    = "im.chat.member.bot_p2p_chat_create_v1"
	EventGroupAtMessage   = "im.message.group_at_msg"
)

// Message Types
const (
	MsgTypeText        = "text"
	MsgTypePost        = "post"
	MsgTypeImage       = "image"
	MsgTypeFile        = "file"
	MsgTypeAudio       = "audio"
	MsgTypeVideo       = "video"
	MsgTypeSticker     = "sticker"
	MsgTypeInteractive  = "interactive"
	MsgTypeShareCard   = "share_card"
)

// Chat Types
const (
	ChatTypeP2P   = "p2p"
	ChatTypeGroup = "group"
)

// Domain Types
const (
	DomainFeishu = "feishu" // 国内版
	DomainLark   = "lark"   // 国际版
)

// =============================================================================
// Configuration
// =============================================================================

// Config contains Lark/Feishu channel configuration
type Config struct {
	// Basic settings
	Domain    string `yaml:"domain"`     // "feishu" or "lark" (default: feishu)
	AppID     string `yaml:"app_id"`     // cli_xxx
	AppSecret string `yaml:"app_secret"` // App secret

	// Connection mode
	ConnectionMode string `yaml:"connection_mode"` // "websocket" (recommended) or "webhook"

	// Webhook settings (for webhook mode)
	WebhookURL     string `yaml:"webhook_url"`
	EncryptKey     string `yaml:"encrypt_key"`      // AES-256 key
	VerifyToken     string `yaml:"verify_token"`     // Verification token
	Port           int    `yaml:"port"`             // Webhook server port (default: 8089)

	// Bot settings
	BotName string `yaml:"bot_name"` // Bot display name

	// Session policies (参考官方 OpenClaw 插件)
	DMPolicy      string   `yaml:"dm_policy"`       // pairing/allowlist/open/disabled
	DMAllowlist   []string `yaml:"dm_allowlist"`    // Allowed user IDs for DM

	GroupPolicy    string   `yaml:"group_policy"`    // open/allowlist/disabled
	GroupAllowlist []string `yaml:"group_allowlist"` // Allowed group IDs

	// Streaming settings
	Streaming       bool `yaml:"streaming"`        // Enable streaming replies (default: true)
	TextChunkLimit  int  `yaml:"text_chunk_limit"` // Max text chunk size (default: 2000)
	MediaMaxMB      int  `yaml:"media_max_mb"`     // Max media file size (default: 30)

	// Performance tuning
	TypingIndicator   bool `yaml:"typing_indicator"`   // Show typing indicator (default: true)
	ResolveSenderNames bool `yaml:"resolve_sender_names"` // Resolve sender names (default: true)

	// Rate limiting
	RateLimitEnabled  bool `yaml:"rate_limit_enabled"`  // Enable rate limiting
	RateLimitPerSec   int  `yaml:"rate_limit_per_sec"`  // Requests per second
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Domain:            DomainFeishu,
		ConnectionMode:    "websocket",
		Port:              8089,
		DMPolicy:          "pairing",
		GroupPolicy:       "open",
		Streaming:         true,
		TextChunkLimit:    2000,
		MediaMaxMB:        30,
		TypingIndicator:   true,
		ResolveSenderNames: true,
		RateLimitEnabled:  true,
		RateLimitPerSec:   10,
	}
}

// =============================================================================
// Channel Implementation
// =============================================================================

// Channel implements the Lark/Feishu channel
type Channel struct {
	config     *Config
	client     *http.Client
	token      string
	tokenExp   time.Time
	tokenMu    sync.RWMutex
	handler    MessageHandler
	server     *http.Server
	eventCache *EventCache
	limiter    *rate.Limiter
	ctx        context.Context
	cancel     context.CancelFunc

	// Metrics
	metricsMu     sync.RWMutex
	messagesSent int64
	messagesRecv int64
	errors       int64
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *Message)

// NewChannel creates a new Lark/Feishu channel
func NewChannel(cfg *Config) *Channel {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.Port == 0 {
		cfg.Port = 8089
	}
	if cfg.Domain == "" {
		cfg.Domain = DomainFeishu
	}

	var limiter *rate.Limiter
	if cfg.RateLimitEnabled && cfg.RateLimitPerSec > 0 {
		limiter = rate.NewLimiter(rate.Limit(cfg.RateLimitPerSec), cfg.RateLimitPerSec*2)
	}

	return &Channel{
		config:     cfg,
		client:     &http.Client{Timeout: 30 * time.Second},
		eventCache: NewEventCache(10000),
		limiter:    limiter,
	}
}

// Type returns channel type
func (c *Channel) Type() string {
	return "lark"
}

// Name returns channel name
func (c *Channel) Name() string {
	if c.config.BotName != "" {
		return c.config.BotName
	}
	return "Lark Bot"
}

// =============================================================================
// Connection Management
// =============================================================================

// Start starts the Lark channel
func (c *Channel) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Get initial access token
	if err := c.getAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Start token refresh goroutine
	go c.tokenRefresher(c.ctx)

	// Start connection based on mode
	switch c.config.ConnectionMode {
	case "websocket":
		// WebSocket mode - 官方推荐
		// 通过飞书 SDK 建立长连接，无需公网 URL
		log.Printf("[Lark] Starting in WebSocket mode (recommended)")
		// TODO: Implement WebSocket client using official SDK
		return c.startWebSocketMode(c.ctx)

	case "webhook":
		// Webhook mode - 需要公网 URL
		log.Printf("[Lark] Starting in Webhook mode on port %d", c.config.Port)
		return c.startWebhookServer(c.ctx)

	default:
		return fmt.Errorf("unsupported connection mode: %s", c.config.ConnectionMode)
	}
}

// Stop stops the Lark channel
func (c *Channel) Stop(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// startWebSocketMode starts WebSocket connection
func (c *Channel) startWebSocketMode(ctx context.Context) error {
	// WebSocket 长连接模式
	// 优势：无需公网 URL，内网穿透，低延迟
	// 参考: https://docs.openclaw.ai/zh-CN/channels/feishu
	
	// Note: 实际实现需要使用飞书官方 Go SDK
	// 这里提供框架结构
	log.Printf("[Lark] WebSocket mode - waiting for official SDK integration")
	log.Printf("[Lark] App ID: %s", c.config.AppID)
	
	// 当前返回 nil 以支持 webhook 模式回退
	return nil
}

// startWebhookServer starts webhook HTTP server
func (c *Channel) startWebhookServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/lark", c.handleWebhook)
	mux.HandleFunc("/lark/verify", c.handleVerification)
	mux.HandleFunc("/health", c.handleHealth)

	addr := fmt.Sprintf(":%d", c.config.Port)
	c.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		c.server.Shutdown(context.Background())
	}()

	log.Printf("[Lark] Webhook server starting on %s", addr)
	return c.server.ListenAndServe()
}

// =============================================================================
// Token Management
// =============================================================================

// getAccessToken obtains a new access token from Lark
func (c *Channel) getAccessToken(ctx context.Context) error {
	authURL := FeishuAuthURL
	if c.config.Domain == DomainLark {
		authURL = LarkAPIBase + "/auth/v3/tenant_access_token/internal"
	}

	params := url.Values{}
	params.Add("app_id", c.config.AppID)
	params.Add("app_secret", c.config.AppSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("lark auth error: code=%d, msg=%s", result.Code, result.Msg)
	}

	c.tokenMu.Lock()
	c.token = result.TenantAccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Expire-60) * time.Second)
	c.tokenMu.Unlock()

	log.Printf("[Lark] Access token obtained, expires in %d seconds", result.Expire)
	return nil
}

// getToken returns current token, refreshing if needed
func (c *Channel) getToken() (string, error) {
	c.tokenMu.RLock()
	if time.Now().Before(c.tokenExp) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	if err := c.getAccessToken(context.Background()); err != nil {
		return "", err
	}

	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()
	return token, nil
}

// tokenRefresher periodically refreshes the access token
func (c *Channel) tokenRefresher(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.getAccessToken(ctx); err != nil {
				log.Printf("[Lark] Token refresh error: %v", err)
			}
		}
	}
}

// =============================================================================
// Message Handling
// =============================================================================

// OnMessage registers a message handler
func (c *Channel) OnMessage(handler MessageHandler) {
	c.handler = handler
}

// Send sends a message via Lark
func (c *Channel) Send(ctx context.Context, to string, msg *Message) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	// Rate limiting
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}

	receiveIDType := c.detectReceiveIDType(to)
	apiURL := fmt.Sprintf("%s?receive_id_type=%s", FeishuSendMessageURL, receiveIDType)

	payload := c.buildMessagePayload(to, msg)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.recordError()
		return err
	}
	defer resp.Body.Close()

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		c.recordError()
		return fmt.Errorf("lark API error: code=%d, msg=%s", result.Code, result.Msg)
	}

	c.recordMessageSent()
	return nil
}

// SendText sends a text message
func (c *Channel) SendText(ctx context.Context, to, content string) error {
	return c.Send(ctx, to, &Message{
		Type:    MsgTypeText,
		Content: content,
	})
}

// SendCard sends an interactive card message (支持流式回复)
func (c *Channel) SendCard(ctx context.Context, to string, card *Card) error {
	return c.Send(ctx, to, &Message{
		Type:    MsgTypeInteractive,
		Content: card,
	})
}

// SendStreamingCard sends a streaming card for real-time updates
// 参考官方 OpenClaw 插件的流式回复机制
func (c *Channel) SendStreamingCard(ctx context.Context, to string, initialContent string) (*StreamingCard, error) {
	card := NewStreamingCard()
	
	// Send initial card
	if err := c.SendCard(ctx, to, card); err != nil {
		return nil, err
	}

	return &StreamingCard{
		channel:    c,
		to:         to,
		card:       card,
		messageID:  card.MessageID,
	}, nil
}

// ReplyMessage replies to a message
func (c *Channel) ReplyMessage(ctx context.Context, messageID string, msg *Message) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s/%s/reply", FeishuSendMessageURL, messageID)
	payload := c.buildReplyPayload(msg)
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result APIResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		return fmt.Errorf("lark API error: %s", result.Msg)
	}
	return nil
}

// UploadFile uploads a file and returns the key
func (c *Channel) UploadFile(ctx context.Context, filename, fileType string, reader io.Reader) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	larkFileType := c.mapFileType(fileType)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := newMultipartWriter(body)
	writer.WriteField("file_type", larkFileType)
	writer.WriteField("file_name", filename)
	writer.WriteFile("file", filename, fileType, reader)

	req, _ := http.NewRequestWithContext(ctx, "POST", FeishuUploadURL, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.ContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		FileKey string `json:"file_key"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Code != 0 {
		return "", fmt.Errorf("lark upload error: %s", result.Msg)
	}
	return result.FileKey, nil
}

// =============================================================================
// Webhook Handlers
// =============================================================================

// handleWebhook handles incoming Lark webhook events
func (c *Channel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		c.handleVerification(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Check for encrypted payload
	if r.Header.Get("X-Lark-Encryption") == "true" && c.config.EncryptKey != "" {
		body, err = c.decryptPayload(body)
		if err != nil {
			log.Printf("[Lark] Decryption error: %v", err)
			http.Error(w, "Decryption Error", http.StatusBadRequest)
			return
		}
	}

	// Parse event envelope
	var envelope EventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		log.Printf("[Lark] Parse error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if envelope.Challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})
		return
	}

	// Handle different event types
	switch envelope.Header.EventType {
	case EventMessageReceive:
		c.handleMessageEvent(w, r, envelope)
	case EventGroupAtMessage:
		c.handleGroupAtMessage(w, r, envelope)
	case EventP2PChatCreate:
		c.handleP2PChatCreate(w, r, envelope)
	default:
		log.Printf("[Lark] Unhandled event type: %s", envelope.Header.EventType)
	}

	// Respond immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":0}`))
}

// handleVerification handles Lark URL verification
func (c *Channel) handleVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if c.config.VerifyToken != "" && token != c.config.VerifyToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	challenge := r.URL.Query().Get("challenge")
	if challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": challenge})
		return
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
}

// handleMessageEvent processes incoming message events
func (c *Channel) handleMessageEvent(w http.ResponseWriter, r *http.Request, envelope EventEnvelope) {
	// Check for duplicate event
	if c.eventCache.Exists(envelope.Header.EventID) {
		log.Printf("[Lark] Duplicate event: %s", envelope.Header.EventID)
		return
	}
	c.eventCache.Add(envelope.Header.EventID)

	// Parse message data
	eventData, ok := envelope.Event["message"].(map[string]interface{})
	if !ok {
		log.Printf("[Lark] Invalid message event data")
		return
	}

	msg := c.parseMessageEvent(eventData, envelope.Header)
	if msg == nil {
		return
	}

	// Apply session policy
	if !c.checkSessionPolicy(msg) {
		return
	}

	c.recordMessageRecv()

	// Handle async
	if c.handler != nil {
		go c.handler(r.Context(), msg)
	}
}

// handleGroupAtMessage handles group @mention messages
func (c *Channel) handleGroupAtMessage(w http.ResponseWriter, r *http.Request, envelope EventEnvelope) {
	// 群聊 @提及消息处理
	c.handleMessageEvent(w, r, envelope)
}

// handleP2PChatCreate handles P2P chat creation events
func (c *Channel) handleP2PChatCreate(w http.ResponseWriter, r *http.Request, envelope EventEnvelope) {
	log.Printf("[Lark] P2P chat created: %s", envelope.Header.EventID)
}

// handleHealth handles health check
func (c *Channel) handleHealth(w http.ResponseWriter, r *http.Request) {
	c.metricsMu.RLock()
	sent := c.messagesSent
	recv := c.messagesRecv
	errs := c.errors
	c.metricsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"channel":       "lark",
		"messages_sent": sent,
		"messages_recv": recv,
		"errors":        errs,
	})
}

// =============================================================================
// Session Policy (参考官方 OpenClaw 插件)
// =============================================================================

// checkSessionPolicy checks if the message should be processed based on policy
func (c *Channel) checkSessionPolicy(msg *Message) bool {
	// Group message
	if msg.ChatType == ChatTypeGroup {
		switch c.config.GroupPolicy {
		case "disabled":
			return false
		case "allowlist":
			// Check if group is in allowlist
			for _, groupID := range c.config.GroupAllowlist {
				if msg.ChatID == groupID {
					return true
				}
			}
			return false
		case "open":
			// Check @mention if required
			if c.config.BotName != "" {
				content, _ := msg.Content.(string)
				if !strings.Contains(content, "@"+c.config.BotName) {
					return false
				}
			}
			return true
		}
	}

	// Direct message
	switch c.config.DMPolicy {
	case "disabled":
		return false
	case "allowlist":
		for _, userID := range c.config.DMAllowlist {
			if msg.From.ID == userID {
				return true
			}
		}
		return false
	case "pairing":
		// 配对模式 - 支持会话绑定
		return true
	case "open":
		return true
	}

	return true
}

// =============================================================================
// Message Building
// =============================================================================

// buildMessagePayload builds message payload for sending
func (c *Channel) buildMessagePayload(to string, msg *Message) map[string]interface{} {
	payload := map[string]interface{}{
		"receive_id": to,
	}

	switch msg.Type {
	case MsgTypeText:
		payload["msg_type"] = MsgTypeText
		content, _ := msg.Content.(string)
		payload["content"] = map[string]string{
			"text": content,
		}

	case MsgTypeImage:
		payload["msg_type"] = MsgTypeImage
		payload["content"] = map[string]string{
			"image_key": msg.ImageKey,
		}

	case MsgTypeInteractive:
		card := msg.Content.(*Card)
		cardBytes, _ := json.Marshal(card)
		payload["msg_type"] = MsgTypeInteractive
		payload["content"] = string(cardBytes)

	default:
		payload["msg_type"] = MsgTypeText
		payload["content"] = map[string]string{
			"text": fmt.Sprintf("%v", msg.Content),
		}
	}

	return payload
}

// buildReplyPayload builds reply payload
func (c *Channel) buildReplyPayload(msg *Message) map[string]interface{} {
	payload := map[string]interface{}{}

	switch msg.Type {
	case MsgTypeText:
		payload["msg_type"] = MsgTypeText
		content, _ := msg.Content.(string)
		payload["content"] = map[string]string{"text": content}
	default:
		payload["msg_type"] = MsgTypeText
		content, _ := msg.Content.(string)
		payload["content"] = map[string]string{"text": content}
	}

	return payload
}

// parseMessageEvent converts Lark message event to our Message format
func (c *Channel) parseMessageEvent(eventData map[string]interface{}, header EventHeader) *Message {
	msgType, _ := eventData["msg_type"].(string)
	chatType, _ := eventData["chat_type"].(string)
	contentStr, _ := eventData["content"].(string)
	messageID, _ := eventData["message_id"].(string)
	chatID, _ := eventData["chat_id"].(string)
	senderData, _ := eventData["sender"].(map[string]interface{})

	msg := &Message{
		ID:       messageID,
		Type:     msgType,
		ChatID:   chatID,
		ChatType: chatType,
		Metadata: map[string]string{
			"event_id":   header.EventID,
			"tenant_key": header.TenantKey,
			"msg_type":   msgType,
			"chat_type":  chatType,
		},
	}

	// Parse sender
	if senderData != nil {
		senderID, _ := senderData["sender_id"].(map[string]interface{})
		if senderID != nil {
			msg.From = &User{
				ID:       getString(senderID["open_id"]),
				Username: getString(senderID["open_id"]),
				Name:     getString(senderData["sender_nickname"]),
			}
		}
	}

	// Parse content
	msg.Content = c.parseContent(msgType, contentStr)

	return msg
}

// parseContent parses message content based on type
func (c *Channel) parseContent(msgType, contentStr string) string {
	switch msgType {
	case MsgTypeText:
		var text struct {
			Text string `json:"text"`
		}
		json.Unmarshal([]byte(contentStr), &text)
		return text.Text

	case MsgTypePost:
		var post PostContent
		json.Unmarshal([]byte(contentStr), &post)
		return c.parseRichText(post)

	case MsgTypeImage:
		return "[图片]"

	case MsgTypeFile:
		var file struct {
			Key string `json:"file_key"`
		}
		json.Unmarshal([]byte(contentStr), &file)
		return "[文件: " + file.Key + "]"

	case MsgTypeAudio:
		return "[语音消息]"

	case MsgTypeVideo:
		return "[视频消息]"

	default:
		return contentStr
	}
}

// parseRichText parses Lark rich text content
func (c *Channel) parseRichText(post PostContent) string {
	var sb strings.Builder

	if post.Title != "" {
		sb.WriteString(post.Title)
		sb.WriteString("\n\n")
	}

	for _, paragraph := range post.Content {
		for _, item := range paragraph {
			switch item.Tag {
			case "text":
				sb.WriteString(item.Text)
			case "a":
				sb.WriteString(item.Text)
			case "at":
				sb.WriteString("@" + item.Text)
			default:
				sb.WriteString(item.Text)
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

// detectReceiveIDType detects receive_id type from ID format
func (c *Channel) detectReceiveIDType(id string) string {
	// 用户ID: ou_xxx
	if strings.HasPrefix(id, "ou_") {
		return "open_id"
	}
	// 群组ID: oc_xxx
	if strings.HasPrefix(id, "oc_") {
		return "chat_id"
	}
	// 邮箱
	if strings.Contains(id, "@") {
		return "email"
	}
	return "open_id"
}

// mapFileType maps file extension to Lark file type
func (c *Channel) mapFileType(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3", ".wav", ".m4a", ".ogg":
		return "audio"
	case ".mp4", ".avi", ".mov", ".wmv":
		return "video"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		return "file"
	default:
		return "file"
	}
}

// =============================================================================
// Encryption/Decryption
// =============================================================================

// decryptPayload decrypts AES-256 encrypted payload
func (c *Channel) decryptPayload(encrypted []byte) ([]byte, error) {
	if c.config.EncryptKey == "" {
		return encrypted, nil
	}

	data, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return nil, err
	}

	if len(data) < 48 {
		return nil, errors.New("encrypted data too short")
	}

	iv := data[:16]
	ciphertext := data[16 : len(data)-32]

	block, err := aes.NewCipher([]byte(c.config.EncryptKey))
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS7 padding
	padLen := int(ciphertext[len(ciphertext)-1])
	if padLen > 16 || padLen == 0 {
		return nil, errors.New("invalid padding")
	}

	return ciphertext[:len(ciphertext)-padLen], nil
}

// EncryptPayload encrypts data for Lark webhook
func (c *Channel) EncryptPayload(data []byte) (string, error) {
	if c.config.EncryptKey == "" {
		return base64.StdEncoding.EncodeToString(data), nil
	}

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(c.config.EncryptKey))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	padLen := 16 - (len(data) % 16)
	padded := append(data, bytes.Repeat([]byte{byte(padLen)}, padLen)...)

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// HMAC-SHA256
	mac := hmacSHA256(append(iv, ciphertext...), []byte(c.config.EncryptKey))

	result := append(iv, ciphertext...)
	result = append(result, mac...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// =============================================================================
// Metrics
// =============================================================================

func (c *Channel) recordMessageSent() {
	c.metricsMu.Lock()
	c.messagesSent++
	c.metricsMu.Unlock()
}

func (c *Channel) recordMessageRecv() {
	c.metricsMu.Lock()
	c.messagesRecv++
	c.metricsMu.Unlock()
}

func (c *Channel) recordError() {
	c.metricsMu.Lock()
	c.errors++
	c.metricsMu.Unlock()
}

// GetMetrics returns current metrics
func (c *Channel) GetMetrics() map[string]int64 {
	c.metricsMu.RLock()
	defer c.metricsMu.RUnlock()
	return map[string]int64{
		"messages_sent": c.messagesSent,
		"messages_recv": c.messagesRecv,
		"errors":        c.errors,
	}
}

// =============================================================================
// Data Structures
// =============================================================================

// Message represents a Lark message
type Message struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Content  interface{}       `json:"content"`
	ChatID   string            `json:"chat_id"`
	ChatType string            `json:"chat_type"`
	From     *User             `json:"from"`
	ImageKey string            `json:"image_key,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// User represents a Lark user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// EventEnvelope represents incoming Lark event envelope
type EventEnvelope struct {
	Schema   string                 `json:"schema"`
	Header   EventHeader             `json:"header"`
	Event    map[string]interface{}  `json:"event"`
	Challenge string                `json:"challenge"`
}

// EventHeader contains event metadata
type EventHeader struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	CreateTime string `json:"create_time"`
	Token     string `json:"token"`
	AppID     string `json:"app_id"`
	TenantKey string `json:"tenant_key"`
}

// APIResponse represents standard Lark API response
type APIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// PostContent represents rich text content
type PostContent struct {
	Title   string `json:"title"`
	Content [][]struct {
		Tag   string `json:"tag"`
		Text  string `json:"text"`
		Href  string `json:"href"`
	} `json:"content"`
}

// Card represents a Lark interactive card
type Card struct {
	Config   *CardConfig              `json:"config,omitempty"`
	Header   *CardHeader              `json:"header,omitempty"`
	Elements []map[string]interface{} `json:"elements"`
	MessageID string                 `json:"message_id,omitempty"`
}

// CardConfig represents card configuration
type CardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode,omitempty"`
}

// CardHeader represents card header
type CardHeader struct {
	Title    *CardText `json:"title,omitempty"`
	Template string    `json:"template,omitempty"`
}

// CardText represents text element in card
type CardText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// NewCard creates a new Lark card
func NewCard(title, content string) *Card {
	return &Card{
		Config: &CardConfig{WideScreenMode: true},
		Header: &CardHeader{
			Title:    &CardText{Tag: "plain_text", Content: title},
			Template: "blue",
		},
		Elements: []map[string]interface{}{
			{
				"tag":     "markdown",
				"content": content,
			},
		},
	}
}

// AddMarkdown adds a markdown element
func (c *Card) AddMarkdown(content string) *Card {
	c.Elements = append(c.Elements, map[string]interface{}{
		"tag":     "markdown",
		"content": content,
	})
	return c
}

// AddButton adds a button element
func (c *Card) AddButton(label, value string) *Card {
	button := map[string]interface{}{
		"tag":   "button",
		"text": map[string]string{
			"tag":     "plain_text",
			"content": label,
		},
		"value": map[string]string{
			"key": value,
		},
	}
	c.Elements = append(c.Elements, map[string]interface{}{
		"tag":     "action",
		"actions": []map[string]interface{}{button},
	})
	return c
}

// SetTemplate sets the card header template color
func (c *Card) SetTemplate(color string) *Card {
	if c.Header == nil {
		c.Header = &CardHeader{}
	}
	c.Header.Template = color
	return c
}

// StreamingCard supports real-time updates
type StreamingCard struct {
	channel   *Channel
	to        string
	card      *Card
	messageID string
}

// Update updates the streaming card content
func (s *StreamingCard) Update(ctx context.Context, content string) error {
	// Update card content
	s.card.Elements = []map[string]interface{}{
		{
			"tag":     "markdown",
			"content": content,
		},
	}

	// Send update
	token, err := s.channel.getToken()
	if err != nil {
		return err
	}

	cardBytes, _ := json.Marshal(s.card)
	payload := map[string]interface{}{
		"msg_type": MsgTypeInteractive,
		"content":  string(cardBytes),
	}
	body, _ := json.Marshal(payload)

	apiURL := fmt.Sprintf("%s/%s", FeishuSendMessageURL, s.messageID)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.channel.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// NewStreamingCard creates a new streaming card
func NewStreamingCard() *Card {
	return &Card{
		Config: &CardConfig{WideScreenMode: true},
		Header: &CardHeader{
			Title:    &CardText{Tag: "plain_text", Content: "AI 助手"},
			Template: "blue",
		},
		Elements: []map[string]interface{}{
			{
				"tag":     "markdown",
				"content": "思考中...",
			},
		},
	}
}

// =============================================================================
// Event Cache
// =============================================================================

// EventCache caches processed event IDs
type EventCache struct {
	mu      sync.RWMutex
	events  map[string]time.Time
	maxSize int
}

// NewEventCache creates a new event cache
func NewEventCache(maxSize int) *EventCache {
	return &EventCache{
		events:  make(map[string]time.Time),
		maxSize: maxSize,
	}
}

// Add adds an event ID to the cache
func (c *EventCache) Add(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events[eventID] = time.Now()

	if len(c.events) > c.maxSize {
		now := time.Now()
		for id, t := range c.events {
			if now.Sub(t) > 24*time.Hour {
				delete(c.events, id)
			}
		}
	}
}

// Exists checks if an event ID exists
func (c *EventCache) Exists(eventID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.events[eventID]
	return exists
}

// =============================================================================
// Helper Functions
// =============================================================================

func hmacSHA256(message, key []byte) []byte {
	h := sha256.New()
	h.Write(message)
	h.Write(key)
	return h.Sum(nil)
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// multipartWriter helper
type multipartWriter struct {
	Body     *bytes.Buffer
	boundary string
}

func newMultipartWriter(body *bytes.Buffer) *multipartWriter {
	return &multipartWriter{
		Body:     body,
		boundary: "----WebKitFormBoundary" + randomString(16),
	}
}

func (m *multipartWriter) WriteField(key, value string) {
	m.Body.WriteString("--" + m.boundary + "\r\n")
	m.Body.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"\r\n\r\n", key))
	m.Body.WriteString(value + "\r\n")
}

func (m *multipartWriter) WriteFile(fieldName, fileName, contentType string, reader io.Reader) {
	m.Body.WriteString("--" + m.boundary + "\r\n")
	m.Body.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n", fieldName, fileName))
	m.Body.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", contentType))
	io.Copy(m.Body, reader)
	m.Body.WriteString("\r\n")
}

func (m *multipartWriter) ContentType() string {
	return "multipart/form-data; boundary=" + m.boundary
}

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// =============================================================================
// Signature Verification
// =============================================================================

// VerifySignature verifies Lark event signature
func (c *Channel) VerifySignature(body []byte, signature string) bool {
	if c.config.EncryptKey == "" {
		return true
	}
	expected := sha256.Sum256(append(body, []byte(c.config.EncryptKey)...))
	return signature == hex.EncodeToString(expected[:])
}

// SignURL generates a signed URL for Lark
func SignURL(baseURL, appID, appSecret string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	params := url.Values{}
	params.Add("app_id", appID)
	params.Add("timestamp", timestamp)

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k][0])
		sb.WriteString("&")
	}
	sb.WriteString(appSecret)

	signature := sha256.Sum256([]byte(sb.String()))

	return fmt.Sprintf("%s?app_id=%s&timestamp=%s&signature=%s",
		baseURL, appID, timestamp, hex.EncodeToString(signature[:])), nil
}
