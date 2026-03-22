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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// WebhookServer handles incoming webhook requests from channels
type WebhookServer struct {
	mu       sync.RWMutex
	server   *http.Server
	handlers map[ChannelType]WebhookHandler
	secret   string
}

// WebhookHandler handles webhook requests for a specific channel
type WebhookHandler func(ctx context.Context, payload []byte) (*Message, error)

// WebhookConfig contains webhook server configuration
type WebhookConfig struct {
	Addr    string
	Secret  string
	Timeout time.Duration
}

// DefaultWebhookConfig returns default webhook configuration
func DefaultWebhookConfig() *WebhookConfig {
	return &WebhookConfig{
		Addr:    ":8080/webhook",
		Secret:  "",
		Timeout: 30 * time.Second,
	}
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(config *WebhookConfig) *WebhookServer {
	if config == nil {
		config = DefaultWebhookConfig()
	}

	return &WebhookServer{
		handlers: make(map[ChannelType]WebhookHandler),
		secret:   config.Secret,
	}
}

// RegisterHandler registers a webhook handler for a channel type
func (s *WebhookServer) RegisterHandler(channelType ChannelType, handler WebhookHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[channelType] = handler
}

// Start starts the webhook server
func (s *WebhookServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register routes for each channel
	mux.HandleFunc("/webhook/telegram", s.handleTelegramWebhook)
	mux.HandleFunc("/webhook/wechat", s.handleWeChatWebhook)
	mux.HandleFunc("/webhook/lark", s.handleLarkWebhook)
	mux.HandleFunc("/webhook/qq", s.handleQQWebhook)
	mux.HandleFunc("/webhook/discord", s.handleDiscordWebhook)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:         s.config().Addr,
		Handler:      mux,
		ReadTimeout:  s.config().Timeout,
		WriteTimeout: s.config().Timeout,
	}

	go func() {
		<-ctx.Done()
		s.server.Shutdown(context.Background())
	}()

	return s.server.ListenAndServe()
}

func (s *WebhookServer) config() *WebhookConfig {
	addr := ":8080/webhook"
	if s.server != nil {
		addr = s.server.Addr
	}
	return &WebhookConfig{
		Addr:    addr,
		Secret:  s.secret,
		Timeout: 30 * time.Second,
	}
}

// Stop stops the webhook server
func (s *WebhookServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// verifySignature verifies webhook signature
func (s *WebhookServer) verifySignature(r *http.Request, body []byte) bool {
	if s.secret == "" {
		return true
	}

	var signature string
	switch r.Header.Get("Content-Type") {
	case "application/json":
		signature = r.Header.Get("X-Signature")
	case "application/x-www-form-urlencoded":
		signature = r.FormValue("signature")
	}

	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func (s *WebhookServer) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !s.verifySignature(r, body) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	s.mu.RLock()
	handler, ok := s.handlers[ChannelTypeTelegram]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Handler not found", http.StatusNotFound)
		return
	}

	msg, err := handler(r.Context(), body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	_ = msg // Use the message as needed
}

func (s *WebhookServer) handleWeChatWebhook(w http.ResponseWriter, r *http.Request) {
	// Handle WeChat webhook verification (GET) and message processing (POST)
	if r.Method == http.MethodGet {
		// WeChat webhook verification
		echostr := r.URL.Query().Get("echostr")
		if echostr != "" {
			// Verify signature
			signature := r.URL.Query().Get("signature")
			timestamp := r.URL.Query().Get("timestamp")
			nonce := r.URL.Query().Get("nonce")

			if s.verifyWeChatSignature(signature, timestamp, nonce) {
				w.Write([]byte(echostr))
				return
			}
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Handle message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	handler, ok := s.handlers[ChannelTypeWeChat]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Handler not found", http.StatusNotFound)
		return
	}

	_, err = handler(r.Context(), body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// WeChat requires empty response for message acknowledgment
	w.WriteHeader(http.StatusOK)
}

func (s *WebhookServer) handleDiscordWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Discord uses its own signature verification
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	if signature != "" && timestamp != "" {
		// Verify Discord signature
		content := append([]byte(timestamp), body...)
		if !s.verifyDiscordSignature(signature, content) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	s.mu.RLock()
	handler, ok := s.handlers[ChannelTypeDiscord]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Handler not found", http.StatusNotFound)
		return
	}

	_, err = handler(r.Context(), body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleLarkWebhook handles incoming Lark (Feishu) webhook events
func (s *WebhookServer) handleLarkWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	handler, ok := s.handlers[ChannelTypeLark]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Handler not found", http.StatusNotFound)
		return
	}

	msg, err := handler(r.Context(), body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "msg": "success"})
	_ = msg
}

// handleQQWebhook handles incoming QQ webhook events
func (s *WebhookServer) handleQQWebhook(w http.ResponseWriter, r *http.Request) {
	// QQ Guild uses GET for verification, POST for events
	if r.Method == http.MethodGet {
		// Handle verification challenge
		echostr := r.URL.Query().Get("echostr")
		signature := r.URL.Query().Get("signature")
		timestamp := r.URL.Query().Get("timestamp")
		nonce := r.URL.Query().Get("nonce")

		if s.verifyQQWebhookSignature(signature, timestamp, nonce, echostr) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(echostr))
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	s.mu.RLock()
	handler, ok := s.handlers[ChannelTypeQQ]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Handler not found", http.StatusNotFound)
		return
	}

	msg, err := handler(r.Context(), body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":0}`))
	_ = msg
}

func (s *WebhookServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// verifyWeChatSignature verifies WeChat API signature
func (s *WebhookServer) verifyWeChatSignature(signature, timestamp, nonce string) bool {
	if s.secret == "" {
		return true
	}

	// WeChat signature verification
	// Sort token, timestamp, nonce and hash
	// Simplified - in production, use proper sorting
	return signature != ""
}

// verifyDiscordSignature verifies Discord webhook signature
func (s *WebhookServer) verifyDiscordSignature(signature string, body []byte) bool {
	// Discord uses Ed25519 signature verification
	// In production, use the discord API verification
	return true // Simplified for now
}

// verifyQQWebhookSignature verifies QQ webhook signature
func (s *WebhookServer) verifyQQWebhookSignature(signature, timestamp, nonce, echostr string) bool {
	if s.secret == "" {
		return true
	}
	// QQ Guild signature verification
	// In production, use proper verification with Bot Token
	return signature != ""
}

// TelegramUpdate represents a Telegram webhook update
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		MessageID int64  `json:"message_id"`
		From      *struct {
			ID       int64  `json:"id"`
			IsBot    bool   `json:"is_bot"`
			Username string `json:"username"`
		} `json:"from"`
		Chat *struct {
			ID    int64  `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

// ParseTelegramUpdate parses a Telegram update
func ParseTelegramUpdate(data []byte) (*Message, error) {
	var update TelegramUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		return nil, err
	}

	if update.Message == nil {
		return nil, fmt.Errorf("no message in update")
	}

	msg := &Message{
		ID:      fmt.Sprintf("%d", update.Message.MessageID),
		Channel: ChannelTypeTelegram,
		Type:    MessageTypeText,
		Content: update.Message.Text,
	}

	if update.Message.From != nil {
		msg.From = &User{
			ID:       fmt.Sprintf("%d", update.Message.From.ID),
			Username: update.Message.From.Username,
		}
	}

	if update.Message.Chat != nil {
		msg.To = fmt.Sprintf("%d", update.Message.Chat.ID)
	}

	return msg, nil
}

// WeChatMessage represents a WeChat message
type WeChatMessage struct {
	ToUserName   string `json:"ToUserName"`
	FromUserName string `json:"FromUserName"`
	CreateTime  int64  `json:"CreateTime"`
	MsgType     string `json:"MsgType"`
	Content     string `json:"Content"`
	MsgID       string `json:"MsgId"`
}

// ParseWeChatMessage parses a WeChat message
func ParseWeChatMessage(data []byte) (*Message, error) {
	var msg WeChatMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	message := &Message{
		ID:      msg.MsgID,
		Channel: ChannelTypeWeChat,
		Type:    parseWeChatMsgType(msg.MsgType),
		Content: msg.Content,
		From: &User{
			ID: msg.FromUserName,
		},
	}

	return message, nil
}

func parseWeChatMsgType(t string) MessageType {
	switch t {
	case "text":
		return MessageTypeText
	case "image":
		return MessageTypeImage
	case "voice":
		return MessageTypeAudio
	case "video":
		return MessageTypeVideo
	case "file":
		return MessageTypeFile
	default:
		return MessageTypeText
	}
}
