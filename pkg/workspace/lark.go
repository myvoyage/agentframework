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
)

// Lark API Endpoints
const (
	LarkAPIGetToken         = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	LarkAPISendMessage      = "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=%s"
	LarkAPIGetMessage       = "https://open.feishu.cn/open-apis/im/v1/messages/%s"
	LarkAPIUploadFile       = "https://open.feishu.cn/open-apis/im/v1/files"
	LarkAPIGetUserInfo      = "https://open.feishu.cn/open-apis/contact/v3/users/%s?user_id_type=open_id"
	LarkAPIReplyMessage     = "https://open.feishu.cn/open-apis/im/v1/messages/%s/reply"
)

// Lark Event Types
const (
	LarkEventMessageReceive = "im.message.receive_v1"
	LarkEventMessageEdit    = "im.message.message_edit_v1"
	LarkEventP2PChatCreate  = "im.chat.member.bot_p2p_chat_create_v1"
)

// Lark Message Types
const (
	LarkMsgTypeText       = "text"
	LarkMsgTypePost       = "post"
	LarkMsgTypeImage      = "image"
	LarkMsgTypeFile       = "file"
	LarkMsgTypeAudio      = "audio"
	LarkMsgTypeVideo      = "video"
	LarkMsgTypeSticker    = "sticker"
	LarkMsgTypeMedia      = "media"
	LarkMsgTypeShareCard  = "share_card"
	LarkMsgTypeInteractive = "interactive"
)

// Lark Chat Types
const (
	LarkChatTypeP2P = "p2p"
	LarkChatTypeGroup = "group"
)

// LarkConfig contains Lark (Feishu) channel configuration
type LarkConfig struct {
	AppID       string `yaml:"app_id"`
	AppSecret   string `yaml:"app_secret"`
	BotName     string `yaml:"bot_name"`
	EncryptKey  string `yaml:"encrypt_key"`  // AES-256 key for event encryption
	VerificationToken string `yaml:"verification_token"` // Verification token from Lark
	WebhookURL  string `yaml:"webhook_url"` // Webhook URL for this server
	Port        int    `yaml:"port"`        // Webhook server port (default: 8089)
}

// LarkChannel implements Channel interface for Lark/Feishu
// Reference: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/bot-v3/bot-overview
type LarkChannel struct {
	config     *LarkConfig
	client     *http.Client
	token      string
	tokenExp   time.Time
	tokenMu    sync.RWMutex
	handler    MessageHandler
	server     *http.Server
	eventCache *EventCache
}

// NewLarkChannel creates a new Lark channel adapter
func NewLarkChannel(cfg *LarkConfig) *LarkChannel {
	if cfg.Port == 0 {
		cfg.Port = 8089
	}
	return &LarkChannel{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		eventCache: NewEventCache(10000),
	}
}

// Type returns channel type
func (c *LarkChannel) Type() ChannelType {
	return ChannelTypeLark
}

// Name returns channel name
func (c *LarkChannel) Name() string {
	if c.config.BotName != "" {
		return c.config.BotName
	}
	return "Lark Bot"
}

// Start starts the Lark webhook server
func (c *LarkChannel) Start(ctx context.Context) error {
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

	// Start token refresh goroutine
	go c.tokenRefresher(ctx)

	// Get initial access token
	if err := c.getAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	go func() {
		<-ctx.Done()
		c.server.Shutdown(context.Background())
	}()

	log.Printf("Lark webhook server starting on %s", addr)
	return c.server.ListenAndServe()
}

// Stop stops the Lark webhook server
func (c *LarkChannel) Stop(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// Send sends a message via Lark
// Reference: https://open.feishu.cn/document/server-docs/im-v1/message-content-description/create-message-content
func (c *LarkChannel) Send(ctx context.Context, to string, message *Message) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	// Determine receive_id_type
	receiveIDType := c.detectReceiveIDType(to)

	// Build message payload based on type
	payload := c.buildMessagePayload(receiveIDType, to, message)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf(LarkAPISendMessage, receiveIDType)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result LarkAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("lark API error: code=%d, msg=%s", result.Code, result.Msg)
	}

	return nil
}

// SendText sends a text message
func (c *LarkChannel) SendText(ctx context.Context, to, content string) error {
	return c.Send(ctx, to, &Message{
		Type:    MessageTypeText,
		Content: content,
	})
}

// SendImage sends an image message (requires image_key from upload)
func (c *LarkChannel) SendImage(ctx context.Context, to, imageKey string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	receiveIDType := c.detectReceiveIDType(to)
	payload := map[string]interface{}{
		"receive_id": to,
		"msg_type":   LarkMsgTypeImage,
		"content": map[string]string{
			"image_key": imageKey,
		},
	}

	body, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf(LarkAPISendMessage, receiveIDType)
	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result LarkAPIResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Code != 0 {
		return fmt.Errorf("lark API error: %s", result.Msg)
	}
	return nil
}

// SendCard sends an interactive card message
func (c *LarkChannel) SendCard(ctx context.Context, to string, card *LarkCard) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	receiveIDType := c.detectReceiveIDType(to)
	cardContent, _ := json.Marshal(card)

	payload := map[string]interface{}{
		"receive_id": to,
		"msg_type":   LarkMsgTypeInteractive,
		"content":    string(cardContent),
	}

	body, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf(LarkAPISendMessage, receiveIDType)
	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result LarkAPIResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Code != 0 {
		return fmt.Errorf("lark API error: %s", result.Msg)
	}
	return nil
}

// ReplyMessage replies to a message
func (c *LarkChannel) ReplyMessage(ctx context.Context, messageID string, message *Message) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	payload := c.buildMessagePayload("message_id", messageID, message)
	body, _ := json.Marshal(payload)

	apiURL := fmt.Sprintf(LarkAPIReplyMessage, messageID)
	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result LarkAPIResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Code != 0 {
		return fmt.Errorf("lark API error: %s", result.Msg)
	}
	return nil
}

// UploadFile uploads a file and returns the key
func (c *LarkChannel) UploadFile(ctx context.Context, filename string, fileType string, reader io.Reader) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	// Determine file type
	larkFileType := c.mapFileType(fileType)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := &multipartWriter{Body: body}
	writer.WriteField("file_type", larkFileType)
	writer.WriteField("file_name", filename)
	writer.WriteFile("file", filename, fileType, reader)

	req, _ := http.NewRequestWithContext(ctx, "POST", LarkAPIUploadFile, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.ContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		FileKey string `json:"file_key"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Code != 0 {
		return "", fmt.Errorf("lark upload error: %s", result.Msg)
	}
	return result.FileKey, nil
}

// OnMessage registers a message handler
func (c *LarkChannel) OnMessage(handler MessageHandler) {
	c.handler = handler
}

// getAccessToken obtains a new access token from Lark
func (c *LarkChannel) getAccessToken(ctx context.Context) error {
	params := url.Values{
		"app_id":     {c.config.AppID},
		"app_secret": {c.config.AppSecret},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", LarkAPIGetToken, strings.NewReader(params.Encode()))
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
		return fmt.Errorf("lark auth error: %s", result.Msg)
	}

	c.tokenMu.Lock()
	c.token = result.TenantAccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Expire-60) * time.Second)
	c.tokenMu.Unlock()

	return nil
}

// getToken returns current token, refreshing if needed
func (c *LarkChannel) getToken() (string, error) {
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
func (c *LarkChannel) tokenRefresher(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.getAccessToken(ctx)
		}
	}
}

// handleWebhook handles incoming Lark webhook events
// Reference: https://open.feishu.cn/document/server-docs/event-subscription-guide/event-list
func (c *LarkChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Handle both GET and POST
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

	// Log raw request for debugging
	log.Printf("Lark webhook received: %s", string(body))

	// Check for encrypted payload
	encrypted := r.Header.Get("X-Lark-Encryption")
	if encrypted == "true" && c.config.EncryptKey != "" {
		body, err = c.decryptPayload(body)
		if err != nil {
			log.Printf("Failed to decrypt payload: %v", err)
			http.Error(w, "Decryption Error", http.StatusBadRequest)
			return
		}
	}

	// Parse event envelope
	var envelope LarkEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		log.Printf("Failed to parse envelope: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verify challenge for URL verification
	if envelope.Challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})
		return
	}

	// Handle different event types
	switch envelope.Header.EventType {
	case LarkEventMessageReceive:
		c.handleMessageEvent(w, r, envelope)
	case LarkEventMessageEdit:
		c.handleMessageEditEvent(w, r, envelope)
	case LarkEventP2PChatCreate:
		c.handleP2PChatCreateEvent(w, r, envelope)
	default:
		log.Printf("Unhandled event type: %s", envelope.Header.EventType)
	}

	// Respond immediately to Lark
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":0}`))
}

// handleVerification handles Lark URL verification
func (c *LarkChannel) handleVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	// Verify token if configured
	if c.config.VerificationToken != "" && token != c.config.VerificationToken {
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
func (c *LarkChannel) handleMessageEvent(w http.ResponseWriter, r *http.Request, envelope LarkEventEnvelope) {
	// Check for duplicate event
	if c.eventCache.Exists(envelope.Header.EventID) {
		log.Printf("Duplicate event: %s", envelope.Header.EventID)
		return
	}
	c.eventCache.Add(envelope.Header.EventID)

	// Extract message data
	eventData, ok := envelope.Event["message"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid message event data")
		return
	}

	msg := c.parseMessageEvent(eventData, envelope.Header)
	if msg == nil {
		return
	}

	// Check if bot was mentioned in group chats
	if msg.ChatType == LarkChatTypeGroup {
		content, _ := json.Marshal(eventData["content"])
		var textContent struct {
			Text string `json:"text"`
		}
		json.Unmarshal(content, &textContent)

		// Check if bot is mentioned
		botName := c.config.BotName
		if botName != "" && !strings.Contains(textContent.Text, "@"+botName) {
			// Bot not mentioned, skip
			return
		}
	}

	// Handle async
	if c.handler != nil {
		go c.handler(r.Context(), msg)
	}
}

// handleMessageEditEvent handles message edit events
func (c *LarkChannel) handleMessageEditEvent(w http.ResponseWriter, r *http.Request, envelope LarkEventEnvelope) {
	// Similar to message event but for edits
	log.Printf("Message edited: %s", envelope.Header.EventID)
}

// handleP2PChatCreateEvent handles P2P chat creation events
func (c *LarkChannel) handleP2PChatCreateEvent(w http.ResponseWriter, r *http.Request, envelope LarkEventEnvelope) {
	// Handle new P2P conversation
	log.Printf("P2P chat created: %s", envelope.Header.EventID)
}

// parseMessageEvent converts Lark message event to our Message format
func (c *LarkChannel) parseMessageEvent(eventData map[string]interface{}, header LarkEventHeader) *Message {
	msgType, _ := eventData["msg_type"].(string)
	chatType, _ := eventData["chat_type"].(string)
	contentStr, _ := eventData["content"].(string)
	messageID, _ := eventData["message_id"].(string)
	chatID, _ := eventData["chat_id"].(string)
	senderData, _ := eventData["sender"].(map[string]interface{})

	msg := &Message{
		ID:      messageID,
		Channel: ChannelTypeLark,
		Type:    parseLarkMsgType(msgType),
		ChatID:  chatID,
		ChatType: chatType,
		Metadata: map[string]string{
			"event_id":  header.EventID,
			"tenant_key": header.TenantKey,
			"msg_type":  msgType,
			"chat_type": chatType,
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

	// Parse content based on type
	msg.Content = c.parseContent(msgType, contentStr, eventData)

	return msg
}

// parseContent parses message content based on type
func (c *LarkChannel) parseContent(msgType, contentStr string, eventData map[string]interface{}) string {
	switch msgType {
	case LarkMsgTypeText:
		var text struct {
			Text string `json:"text"`
		}
		json.Unmarshal([]byte(contentStr), &text)
		return text.Text

	case LarkMsgTypePost:
		// Parse rich text content
		var post struct {
			Title   string `json:"title"`
			Content [][]struct {
				Tag   string `json:"tag"`
				Text  string `json:"text"`
				Href  string `json:"href"`
			} `json:"content"`
		}
		json.Unmarshal([]byte(contentStr), &post)
		return parseRichText(post)

	case LarkMsgTypeImage:
		return "[图片]"

	case LarkMsgTypeFile:
		var file struct {
			Key string `json:"file_key"`
		}
		json.Unmarshal([]byte(contentStr), &file)
		return "[文件: " + file.Key + "]"

	case LarkMsgTypeAudio:
		return "[语音消息]"

	case LarkMsgTypeVideo:
		return "[视频消息]"

	default:
		return contentStr
	}
}

// buildMessagePayload builds message payload for sending
func (c *LarkChannel) buildMessagePayload(receiveIDType, to string, message *Message) map[string]interface{} {
	payload := map[string]interface{}{
		"receive_id": to,
	}

	switch message.Type {
	case MessageTypeText:
		payload["msg_type"] = LarkMsgTypeText
		payload["content"] = map[string]string{
			"text": message.Content,
		}

	case MessageTypeImage:
		payload["msg_type"] = LarkMsgTypeImage
		payload["content"] = map[string]string{
			"image_key": message.ImageKey,
		}

	case MessageTypeCard:
		card := c.buildCardFromMessage(message)
		payload["msg_type"] = LarkMsgTypeInteractive
		payload["content"] = card

	default:
		payload["msg_type"] = LarkMsgTypeText
		payload["content"] = map[string]string{
			"text": message.Content,
		}
	}

	return payload
}

// buildCardFromMessage builds a Lark card from message
func (c *LarkChannel) buildCardFromMessage(message *Message) string {
	card := LarkCard{
		Config: &LarkCardConfig{
			WideScreenMode: true,
		},
		Header: &LarkCardHeader{
			Title: &LarkCardText{
				Tag:     "plain_text",
				Content: message.Title,
			},
			Template: "blue",
		},
		Elements: []map[string]interface{}{
			{
				"tag":  "markdown",
				"content": message.Content,
			},
		},
	}

	if len(message.Actions) > 0 {
		var actions []map[string]interface{}
		for _, action := range message.Actions {
			actions = append(actions, map[string]interface{}{
				"tag": "button",
				"text": map[string]string{
					"tag":     "plain_text",
					"content": action.Label,
				},
				"type": "primary",
			})
		}
		card.Elements = append(card.Elements, map[string]interface{}{
			"tag":     "action",
			"actions": actions,
		})
	}

	bytes, _ := json.Marshal(card)
	return string(bytes)
}

// detectReceiveIDType detects receive_id type from ID format
func (c *LarkChannel) detectReceiveIDType(id string) string {
	if strings.HasPrefix(id, "ou_") {
		return "open_id"
	} else if strings.HasPrefix(id, "oc_") {
		return "chat_id"
	} else if strings.HasPrefix(id, "oc") && len(id) > 10 {
		return "chat_id"
	} else if strings.HasPrefix(id, "ou") && len(id) > 10 {
		return "user_id"
	} else if strings.Contains(id, "@") {
		return "email"
	}
	return "open_id"
}

// mapFileType maps file extension to Lark file type
func (c *LarkChannel) mapFileType(ext string) string {
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

// decryptPayload decrypts AES-256 encrypted payload
func (c *LarkChannel) decryptPayload(encrypted []byte) ([]byte, error) {
	if c.config.EncryptKey == "" {
		return encrypted, nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return nil, err
	}

	// Parse encrypted data format: [16-byte IV][encrypted data][HMAC]
	if len(data) < 48 {
		return nil, errors.New("encrypted data too short")
	}

	iv := data[:16]
	ciphertext := data[16 : len(data)-32]
	// mac := data[len(data)-32:]

	// Decrypt
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

// EncryptPayload encrypts data for Lark webhook (if needed)
func (c *LarkChannel) EncryptPayload(data []byte) (string, error) {
	if c.config.EncryptKey == "" {
		return base64.StdEncoding.EncodeToString(data), nil
	}

	// Generate random IV
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

	// Combine: IV + ciphertext + HMAC
	result := append(iv, ciphertext...)
	result = append(result, mac...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// hmacSHA256 computes HMAC-SHA256
func hmacSHA256(message, key []byte) []byte {
	h := sha256.New()
	h.Write(message)
	h.Write(key)
	return h.Sum(nil)
}

// verifySignature verifies Lark event signature
func (c *LarkChannel) verifySignature(body []byte, signature string) bool {
	expected := sha256Sum(append(body, []byte(c.config.EncryptKey)...))
	return signature == hex.EncodeToString(expected)
}

func sha256Sum(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func (c *LarkChannel) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "channel": "lark"})
}

// =============================================================================
// Data Structures
// =============================================================================

// LarkEventEnvelope represents incoming Lark event envelope
type LarkEventEnvelope struct {
	Schema        string                 `json:"schema"`
	Header        LarkEventHeader        `json:"header"`
	Event        map[string]interface{}  `json:"event"`
	Challenge    string                 `json:"challenge"` // For URL verification
}

// LarkEventHeader contains event metadata
type LarkEventHeader struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	CreateTime string `json:"create_time"`
	Token     string `json:"token"`
	AppID     string `json:"app_id"`
	TenantKey string `json:"tenant_key"`
}

// LarkAPIResponse represents standard Lark API response
type LarkAPIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// LarkUser represents a Lark user
type LarkUser struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	TenantKey  string `json:"tenant_key"`
}

// LarkSender represents message sender
type LarkSender struct {
	SenderID       LarkSenderID `json:"sender_id"`
	SenderIDStr    string       `json:"sender_id_str"`
	SenderType     string       `json:"sender_type"`
	SenderNickname string       `json:"sender_nickname"`
}

// LarkSenderID represents sender ID
type LarkSenderID struct {
	OpenID  string `json:"open_id"`
	UserID  string `json:"user_id"`
	UnionID string `json:"union_id"`
}

// LarkCard represents a Lark interactive card
type LarkCard struct {
	Config   *LarkCardConfig    `json:"config,omitempty"`
	Header   *LarkCardHeader    `json:"header,omitempty"`
	Elements []map[string]interface{} `json:"elements"`
}

// LarkCardConfig represents card configuration
type LarkCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode,omitempty"`
}

// LarkCardHeader represents card header
type LarkCardHeader struct {
	Title    *LarkCardText `json:"title,omitempty"`
	Template string        `json:"template,omitempty"` // blue, red, yellow, green, purple, gray, orange
}

// LarkCardText represents text element in card
type LarkCardText struct {
	Tag     string `json:"tag"` // plain_text, lark_md
	Content string `json:"content"`
}

// LarkAction represents a card action
type LarkAction struct {
	Label   string
	URL     string
	Type    string
}

// =============================================================================
// Helper Functions
// =============================================================================

// NewLarkCard creates a new Lark card
func NewLarkCard(title, content string) *LarkCard {
	return &LarkCard{
		Config: &LarkCardConfig{
			WideScreenMode: true,
		},
		Header: &LarkCardHeader{
			Title: &LarkCardText{
				Tag:     "plain_text",
				Content: title,
			},
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

// AddElement adds an element to the card
func (c *LarkCard) AddMarkdown(content string) *LarkCard {
	c.Elements = append(c.Elements, map[string]interface{}{
		"tag":     "markdown",
		"content": content,
	})
	return c
}

// AddButton adds a button to the card
func (c *LarkCard) AddButton(label, value, actionType string) *LarkCard {
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

	if actionType == "" {
		actionType = "click"
	}
	button["action_type"] = actionType

	actionElement := map[string]interface{}{
		"tag":     "action",
		"actions": []map[string]interface{}{button},
	}
	c.Elements = append(c.Elements, actionElement)
	return c
}

// SetTemplate sets the card header template color
func (c *LarkCard) SetTemplate(color string) *LarkCard {
	if c.Header == nil {
		c.Header = &LarkCardHeader{}
	}
	c.Header.Template = color
	return c
}

// BuildCardResponse builds a Lark card response
func BuildCardResponse(title, content string) string {
	card := NewLarkCard(title, content)
	bytes, _ := json.Marshal(card)
	return string(bytes)
}

// parseRichText parses Lark rich text content to plain text
func parseRichText(post struct {
	Title   string `json:"title"`
	Content [][]struct {
		Tag   string `json:"tag"`
		Text  string `json:"text"`
		Href  string `json:"href"`
	} `json:"content"`
}) string {
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

// getString safely gets string from interface
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// SignLarkURL generates a signed URL for Lark
func SignLarkURL(baseURL string, appID, appSecret string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	params := url.Values{
		"app_id":    {appID},
		"timestamp": {timestamp},
	}

	// Sort and concatenate
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

	signature := sha256Sum([]byte(sb.String()))

	return fmt.Sprintf("%s?app_id=%s&timestamp=%s&signature=%s",
		baseURL, appID, timestamp, hex.EncodeToString(signature)), nil
}

// =============================================================================
// Event Cache (for deduplication)
// =============================================================================

// EventCache caches processed event IDs to prevent duplicate processing
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

	// Cleanup old entries if too large
	if len(c.events) > c.maxSize {
		now := time.Now()
		for id, t := range c.events {
			if now.Sub(t) > 24*time.Hour {
				delete(c.events, id)
			}
		}
	}
}

// Exists checks if an event ID exists in the cache
func (c *EventCache) Exists(eventID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.events[eventID]
	return exists
}

// =============================================================================
// Multipart Writer Helper
// =============================================================================

type multipartWriter struct {
	Body      *bytes.Buffer
	boundary  string
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

// =============================================================================
// Message Type Conversion
// =============================================================================

// parseLarkMsgType converts Lark message type to our MessageType
func parseLarkMsgType(t string) MessageType {
	switch t {
	case LarkMsgTypeText:
		return MessageTypeText
	case LarkMsgTypePost:
		return MessageTypeRichText
	case LarkMsgTypeImage:
		return MessageTypeImage
	case LarkMsgTypeAudio:
		return MessageTypeAudio
	case LarkMsgTypeVideo:
		return MessageTypeVideo
	case LarkMsgTypeFile:
		return MessageTypeFile
	case LarkMsgTypeSticker, LarkMsgTypeMedia:
		return MessageTypeImage
	case LarkMsgTypeInteractive:
		return MessageTypeCard
	default:
		return MessageTypeText
	}
}

// toLarkMsgType converts our MessageType to Lark message type
func toLarkMsgType(t MessageType) string {
	switch t {
	case MessageTypeText:
		return LarkMsgTypeText
	case MessageTypeRichText:
		return LarkMsgTypePost
	case MessageTypeImage:
		return LarkMsgTypeImage
	case MessageTypeAudio:
		return LarkMsgTypeAudio
	case MessageTypeVideo:
		return LarkMsgTypeVideo
	case MessageTypeFile:
		return LarkMsgTypeFile
	case MessageTypeCard:
		return LarkMsgTypeInteractive
	default:
		return LarkMsgTypeText
	}
}
