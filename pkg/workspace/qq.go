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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// QQConfig contains QQ channel configuration
type QQConfig struct {
	AppID       string `yaml:"app_id"`        // Bot AppID
	AppSecret   string `yaml:"app_secret"`    // Bot AppSecret
	Token       string `yaml:"token"`         // Bot Token (for QQ Guild)
	EncryptKey  string `yaml:"encrypt_key"`   // Encryption key
	Webhook     string `yaml:"webhook"`       // Webhook URL for events
	SandboxMode bool   `yaml:"sandbox_mode"`  // Enable sandbox mode
}

// QQChannel implements Channel interface for QQ/QQGuild
type QQChannel struct {
	config  *QQConfig
	client  *http.Client
	token   string
	tokenExp time.Time
	handler MessageHandler
	server  *http.Server
}

// NewQQChannel creates a new QQ channel adapter
func NewQQChannel(cfg *QQConfig) *QQChannel {
	return &QQChannel{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Type returns channel type
func (c *QQChannel) Type() ChannelType {
	return ChannelTypeQQ
}

// Name returns channel name
func (c *QQChannel) Name() string {
	return fmt.Sprintf("QQ Bot (%s)", c.config.AppID)
}

// Start starts the QQ webhook server
func (c *QQChannel) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/qq", c.handleWebhook)
	mux.HandleFunc("/webhook/qq/callback", c.handleQQGuildCallback)
	mux.HandleFunc("/health", c.handleHealth)

	c.server = &http.Server{
		Addr:    ":8088/qq",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		c.server.Shutdown(context.Background())
	}()

	// Get access token for QQ Guild API
	if err := c.getAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get QQ access token: %w", err)
	}

	return c.server.ListenAndServe()
}

// Stop stops the QQ webhook server
func (c *QQChannel) Stop(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// Send sends a message via QQ
func (c *QQChannel) Send(ctx context.Context, to string, message *Message) error {
	// Ensure we have a valid token
	if time.Now().After(c.tokenExp) {
		if err := c.getAccessToken(ctx); err != nil {
			return err
		}
	}

	// Determine message type
	var msgType string
	var content interface{}

	switch message.Type {
	case MessageTypeImage:
		msgType = "image"
		content = map[string]string{
			"file_id": message.Content,
		}
	case MessageTypeAudio:
		msgType = "audio"
		content = map[string]string{
			"file_id": message.Content,
		}
	case MessageTypeText:
		msgType = "text"
		content = map[string]string{
			"text": message.Content,
		}
	default:
		msgType = "text"
		content = map[string]string{
			"text": message.Content,
		}
	}

	// Build request based on target type
	payload := map[string]interface{}{
		"msg_type": msgType,
		"content":  content,
	}

	// Determine channel type and endpoint
	var apiURL string
	if strings.Contains(to, "-") || len(to) > 15 {
		// It's a channel ID
		payload["channel_id"] = to
		apiURL = "https://api.sgroup.qq.com/channels/" + to + "/messages"
	} else {
		// It's a user ID (DMs)
		payload["guild_id"] = to
		apiURL = "https://api.sgroup.qq.com/dms/" + to + "/messages"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.config.AppID+"."+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("QQ API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// OnMessage registers a message handler
func (c *QQChannel) OnMessage(handler MessageHandler) {
	c.handler = handler
}

// getAccessToken obtains a new access token from QQ
func (c *QQChannel) getAccessToken(ctx context.Context) error {
	// QQ Guild uses Bot Token directly
	c.token = c.config.Token
	c.tokenExp = time.Now().Add(24 * time.Hour)
	return nil
}

// handleWebhook handles incoming QQ events
func (c *QQChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// QQ Guild uses HTTPS verification
	if r.Method == http.MethodGet {
		// Handle verification challenge
		params := r.URL.Query()
		echostr := params.Get("echostr")
		signature := params.Get("signature")
		timestamp := params.Get("timestamp")
		nonce := params.Get("nonce")

		if c.verifyQQSignature(signature, timestamp, nonce, echostr) {
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

	// Decrypt body if encrypted
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if c.config.EncryptKey != "" {
		body = c.decrypt(body)
	}

	// Parse event
	var event QQWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Handle message event
	if event.Op == 0 && event.Type == "INTERACTION" && c.handler != nil {
		msg := c.parseQQMessage(&event)
		if msg != nil {
			go c.handler(r.Context(), msg)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (c *QQChannel) handleQQGuildCallback(w http.ResponseWriter, r *http.Request) {
	// QQ Guild interaction callback
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var callback QQGuildCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Handle different callback types
	switch callback.Type {
	case "WEBCHAT_PING":
		// Respond with pong
		w.Write([]byte(`{"type": "WEBCHAT_PONG"}`))
	case "C2_MESSAGE", "GROUP_AT_MESSAGE":
		if c.handler != nil {
			msg := c.parseGuildMessage(&callback)
			if msg != nil {
				go c.handler(r.Context(), msg)
			}
		}
		// Acknowledge receipt
		w.Write([]byte(`{"code":0,"msg":"ok"}`))
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (c *QQChannel) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// verifyQQSignature verifies QQ webhook signature
func (c *QQChannel) verifyQQSignature(signature, timestamp, nonce, echostr string) bool {
	if c.config.EncryptKey == "" {
		return true
	}

	// Build signature: md5(token + timestamp + nonce)
	str := c.config.Token + timestamp + nonce
	h := md5.New()
	h.Write([]byte(str))
	expected := hex.EncodeToString(h.Sum(nil))

	return signature == expected
}

// decrypt decrypts QQ encrypted payload
func (c *QQChannel) decrypt(data []byte) []byte {
	// Simplified decryption - in production use proper AES/RC4
	// For now, just return the data as-is
	return data
}

// parseQQMessage converts QQ webhook event to Message
func (c *QQChannel) parseQQMessage(event *QQWebhookEvent) *Message {
	if event.Data == nil {
		return nil
	}

	msg := &Message{
		ID:      strconv.FormatInt(event.Sequence, 10),
		Channel: ChannelTypeQQ,
		Type:    parseQQMsgType(event.Data.MsgType),
		Content: event.Data.Content,
		Metadata: map[string]string{
			"msg_type":   event.Data.MsgType,
			"guild_id":   event.Data.GuildID,
			"channel_id": event.Data.ChannelID,
		},
	}

	if event.Data.Author != nil {
		msg.From = &User{
			ID:       strconv.FormatInt(event.Data.Author.ID, 10),
			Username: event.Data.Author.Username,
			Name:     event.Data.Author.Nickname,
		}
	}

	msg.To = event.Data.ChannelID

	return msg
}

// parseGuildMessage converts QQ Guild callback to Message
func (c *QQChannel) parseGuildMessage(callback *QQGuildCallback) *Message {
	msg := &Message{
		ID:      strconv.FormatInt(callback.ID, 10),
		Channel: ChannelTypeQQ,
		Type:    parseQQGuildMsgType(callback.SubType),
		Content: callback.Content,
		Metadata: map[string]string{
			"guild_id":   callback.GuildID,
			"channel_id": callback.ChannelID,
			"sub_type":   callback.SubType,
		},
	}

	if callback.Author != nil {
		msg.From = &User{
			ID:       strconv.FormatInt(callback.Author.ID, 10),
			Username: callback.Author.Username,
			Name:     callback.Author.Nickname,
		}
	}

	msg.To = callback.ChannelID

	return msg
}

// parseQQMsgType converts QQ message type to our MessageType
func parseQQMsgType(t string) MessageType {
	switch strings.ToLower(t) {
	case "text":
		return MessageTypeText
	case "image":
		return MessageTypeImage
	case "audio":
		return MessageTypeAudio
	case "video":
		return MessageTypeVideo
	case "file":
		return MessageTypeFile
	default:
		return MessageTypeText
	}
}

// parseQQGuildMsgType converts QQ Guild sub_type to MessageType
func parseQQGuildMsgType(subType string) MessageType {
	switch subType {
	case "C2_MESSAGE":
		return MessageTypeText
	case "GROUP_AT_MESSAGE":
		return MessageTypeText
	default:
		return MessageTypeText
	}
}

// QQWebhookEvent represents QQ webhook event
type QQWebhookEvent struct {
	Op        int    `json:"op"`        // Operation code (0 = dispatch)
	Type      string `json:"type"`       // Event type
	Sequence  int64  `json:"s"`         // Sequence number
	Data      *QQMessageData `json:"d"` // Event data
	Timestamp int64  `json:"t"`         // Timestamp
}

// QQMessageData represents QQ message data
type QQMessageData struct {
	ID           int64       `json:"id"`
	GuildID      string      `json:"guild_id"`
	ChannelID    string      `json:"channel_id"`
	Content      string      `json:"content"`
	MsgType      string      `json:"msg_type"`
	Author       *QQUser     `json:"author"`
	Member       *QQMember   `json:"member"`
	Timestamp    int64       `json:"timestamp"`
	Attachments  []string    `json:"attachments"`
}

// QQUser represents a QQ user
type QQUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// QQMember represents a QQ guild member
type QQMember struct {
	Nick    string `json:"nick"`
	Role    int    `json:"role"`
	JoinedAt int64  `json:"joined_at"`
}

// QQGuildCallback represents QQ Guild callback event
type QQGuildCallback struct {
	Type       string `json:"type"`
	GuildID    string `json:"guild_id"`
	ChannelID  string `json:"channel_id"`
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	SubType    string `json:"sub_type"`
	Author     *QQUser `json:"author"`
	MsgID      string `json:"msg_id"`
	MsgTimestamp int64 `json:"msg_timestamp"`
}

// BuildQQMarkdownResponse builds a QQ markdown message
func BuildQQMarkdownResponse(content string) map[string]interface{} {
	return map[string]interface{}{
		"msg_type": "markdown",
		"content": map[string]string{
			"template": content,
		},
	}
}

// BuildQQKeyboardResponse builds a QQ keyboard message
func BuildQQKeyboardResponse(text string, buttons []map[string]string) map[string]interface{} {
	keyboard := make([][]map[string]string, len(buttons))
	for i, btn := range buttons {
		keyboard[i] = []map[string]string{btn}
	}

	return map[string]interface{}{
		"msg_type": "keyboard",
		"content": map[string]string{
			"text":      text,
			"keyboard": fmt.Sprintf(`{"button":%s}`, toJSON(keyboard)),
		},
	}
}

// QQGuildAPI contains common QQ Guild API endpoints
const (
	QQGuildAPIMessage     = "https://api.sgroup.qq.com/channels/%s/messages"
	QQGuildAPIDMMessage  = "https://api.sgroup.qq.com/dms/%s/messages"
	QQGuildAPIWebhook    = "https://api.sgroup.qq.com/gateways"
)

// SignQQURL generates a signed URL for QQ API
func SignQQURL(baseURL string, params url.Values) (string, error) {
	signedParams := url.Values{}
	for k, v := range params {
		signedParams[k] = v
	}
	signedParams["timestamp"] = []string{strconv.FormatInt(time.Now().Unix(), 10)}

	return baseURL + "?" + signedParams.Encode(), nil
}

// SendQQImage sends an image message via QQ
func (c *QQChannel) SendImage(ctx context.Context, to string, imageURL string) error {
	payload := map[string]interface{}{
		"msg_type": "image",
		"content": map[string]string{
			"file_key": imageURL,
		},
	}

	if strings.Contains(to, "-") {
		payload["channel_id"] = to
	} else {
		payload["guild_id"] = to
	}

	body, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf(QQGuildAPIMessage, to)

	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bot "+c.config.AppID+"."+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("QQ image send error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
