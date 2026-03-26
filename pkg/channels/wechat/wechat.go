// Package wechat provides WeChat channel implementation for AgentFramework
// Based on OpenClaw WeChat integration patterns
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package wechat

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// ==================== Types and Constants ====================

// ChannelType represents the type of WeChat channel
type ChannelType string

const (
	// ChannelTypeWecom 企业微信应用
	ChannelTypeWecom ChannelType = "wecom"
	// ChannelTypeClawBot 微信官方 ClawBot 插件
	ChannelTypeClawBot ChannelType = "clawbot"
	// ChannelTypeMP 微信公众号
	ChannelTypeMP ChannelType = "mp"
	// ChannelTypeMiniProgram 微信小程序
	ChannelTypeMiniProgram ChannelType = "miniprogram"
)

// Message types
const (
	MsgTypeText       = "text"
	MsgTypeImage      = "image"
	MsgTypeVoice      = "voice"
	MsgTypeVideo      = "video"
	MsgTypeFile       = "file"
	MsgTypeLocation   = "location"
	MsgTypeLink       = "link"
	MsgTypeEvent      = "event"
	MsgTypeAttachment = "attachment"
)

// Event types
const (
	EventSubscribe   = "subscribe"
	EventUnsubscribe = "unsubscribe"
	EventClick       = "CLICK"
	EventView        = "VIEW"
	EventScan        = "SCAN"
)

// ==================== Configuration ====================

// Config contains WeChat channel configuration
type Config struct {
	// Channel type
	Type ChannelType `yaml:"type" json:"type"`

	// Common fields
	Token    string `yaml:"token" json:"token"`       // Token for signature verification
	Secret   string `yaml:"secret" json:"secret"`     // App secret
	Endpoint string `yaml:"endpoint" json:"endpoint"` // Custom endpoint

	// Enterprise WeChat (企业微信)
	CorpID    string `yaml:"corp_id" json:"corp_id"`
	AgentID   string `yaml:"agent_id" json:"agent_id"`
	CorpSecret string `yaml:"corp_secret" json:"corp_secret"`

	// Public Account (公众号)
	AppID     string `yaml:"app_id" json:"app_id"`
	AppSecret string `yaml:"app_secret" json:"app_secret"`
	EncodingAESKey string `yaml:"encoding_aes_key" json:"encoding_aes_key"`

	// ClawBot (官方插件)
	ClawBotURL   string `yaml:"clawbot_url" json:"clawbot_url"`
	ClawBotToken string `yaml:"clawbot_token" json:"clawbot_token"`

	// Webhook server
	Port int    `yaml:"port" json:"port"`
	Path string `yaml:"path" json:"path"`

	// Mini Program
	MiniProgramAppID     string `yaml:"miniprogram_app_id" json:"miniprogram_app_id"`
	MiniProgramAppSecret string `yaml:"miniprogram_app_secret" json:"miniprogram_app_secret"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Type: ChannelTypeWecom,
		Port: 8080,
		Path: "/wechat/callback",
	}
}

// ==================== Message Types ====================

// Message represents a unified WeChat message
type Message struct {
	MsgID      string            `json:"msg_id"`
	MsgType    string            `json:"msg_type"`
	Content    string            `json:"content"`
	FromUser   string            `json:"from_user"`
	ToUser     string            `json:"to_user"`
	CreateTime int64             `json:"create_time"`
	ChatID     string            `json:"chat_id,omitempty"`     // Group chat ID
	ChatType   string            `json:"chat_type,omitempty"`   // single/group
	Metadata   map[string]string `json:"metadata,omitempty"`
	
	// Rich content
	MediaID    string            `json:"media_id,omitempty"`
	Format     string            `json:"format,omitempty"`
	MsgData    []byte            `json:"msg_data,omitempty"`
}

// Reply represents a reply message
type Reply struct {
	ToUser   string `json:"touser"`
	MsgType  string `json:"msgtype"`
	Content  string `json:"content,omitempty"`
	MediaID  string `json:"media_id,omitempty"`
}

// ==================== WecomMessage (企业微信) ====================

// WecomMessage represents enterprise WeChat message format
type WecomMessage struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgID        int64  `xml:"MsgId"`
	AgentID      int    `xml:"AgentID"`
	
	// Image message
	PicURL   string `xml:"PicUrl,omitempty"`
	MediaID  string `xml:"MediaId,omitempty"`
	
	// Location message
	LocationX float64 `xml:"Location_X,omitempty"`
	LocationY float64 `xml:"Location_Y,omitempty"`
	Label     string  `xml:"Label,omitempty"`
	
	// Event message
	Event string `xml:"Event,omitempty"`
	EventKey string `xml:"EventKey,omitempty"`
}

// WecomReply represents enterprise WeChat reply
type WecomReply struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content,omitempty"`
}

// ==================== MPMessage (公众号) ====================

// MPMessage represents WeChat public account message
type MPMessage struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgID        int64  `xml:"MsgId"`
	
	// Image
	PicURL  string `xml:"PicUrl,omitempty"`
	MediaID string `xml:"MediaId,omitempty"`
	
	// Event
	Event    string `xml:"Event,omitempty"`
	EventKey string `xml:"EventKey,omitempty"`
}

// ==================== ClawBotMessage ====================

// ClawBotMessage represents ClawBot plugin message
type ClawBotMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Content     string                 `json:"content"`
	FromUser    string                 `json:"from_user"`
	ToUser      string                 `json:"to_user"`
	ChatID      string                 `json:"chat_id"`
	ChatType    string                 `json:"chat_type"`
	Timestamp   int64                  `json:"timestamp"`
	Attachments []ClawBotAttachment    `json:"attachments,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// ClawBotAttachment represents an attachment in ClawBot message
type ClawBotAttachment struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	MediaID  string `json:"media_id,omitempty"`
	Filename string `json:"filename,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// ClawBotResponse represents response to ClawBot
type ClawBotResponse struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	ToUser  string `json:"to_user,omitempty"`
	ChatID  string `json:"chat_id,omitempty"`
}

// ==================== MessageHandler ====================

// MessageHandler handles incoming WeChat messages
type MessageHandler func(ctx context.Context, msg *Message) (*Reply, error)

// ==================== Client ====================

// Client is the main WeChat channel client
type Client struct {
	config     *Config
	httpClient *http.Client
	
	// Access token management
	accessToken     string
	tokenExpireTime time.Time
	tokenMu         sync.RWMutex
	
	// Message handler
	handler MessageHandler
	
	// Webhook server
	server *http.Server
	
	// State
	mu      sync.RWMutex
	running bool
}

// NewClient creates a new WeChat channel client
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ==================== Channel Interface ====================

// Type returns the channel type
func (c *Client) Type() ChannelType {
	return c.config.Type
}

// Name returns the channel name
func (c *Client) Name() string {
	switch c.config.Type {
	case ChannelTypeWecom:
		return fmt.Sprintf("企业微信 (%s)", c.config.AgentID)
	case ChannelTypeClawBot:
		return "微信 ClawBot"
	case ChannelTypeMP:
		return fmt.Sprintf("公众号 (%s)", c.config.AppID)
	case ChannelTypeMiniProgram:
		return fmt.Sprintf("小程序 (%s)", c.config.MiniProgramAppID)
	default:
		return "微信"
	}
}

// Start starts the WeChat channel
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.running {
		return nil
	}
	
	// Get initial access token
	if c.config.Type != ChannelTypeClawBot {
		if err := c.refreshAccessToken(ctx); err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
		
		// Start token refresh goroutine
		go c.tokenRefresher(ctx)
	}
	
	// Start webhook server
	if c.config.Type != ChannelTypeClawBot {
		if err := c.startWebhookServer(ctx); err != nil {
			return fmt.Errorf("failed to start webhook server: %w", err)
		}
	}
	
	c.running = true
	return nil
}

// Stop stops the WeChat channel
func (c *Client) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return nil
	}
	
	if c.server != nil {
		if err := c.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}
	
	c.running = false
	return nil
}

// OnMessage registers a message handler
func (c *Client) OnMessage(handler MessageHandler) {
	c.handler = handler
}

// Send sends a message to WeChat
func (c *Client) Send(ctx context.Context, reply *Reply) error {
	switch c.config.Type {
	case ChannelTypeWecom:
		return c.sendWecomMessage(ctx, reply)
	case ChannelTypeClawBot:
		return c.sendClawBotMessage(ctx, reply)
	case ChannelTypeMP:
		return c.sendMPMessage(ctx, reply)
	default:
		return fmt.Errorf("unsupported channel type: %s", c.config.Type)
	}
}

// ==================== Webhook Server ====================

func (c *Client) startWebhookServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(c.config.Path, c.handleCallback)
	
	addr := fmt.Sprintf(":%d", c.config.Port)
	c.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error
		}
	}()
	
	return nil
}

// handleCallback handles WeChat callback
func (c *Client) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify signature
	if !c.verifySignature(r) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	
	// Handle verification request
	if r.Method == http.MethodGet {
		echostr := r.URL.Query().Get("echostr")
		w.Write([]byte(echostr))
		return
	}
	
	// Handle message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	
	msg, err := c.parseMessage(body)
	if err != nil {
		http.Error(w, "Failed to parse message", http.StatusBadRequest)
		return
	}
	
	// Call handler
	if c.handler != nil {
		reply, err := c.handler(r.Context(), msg)
		if err != nil {
			http.Error(w, "Handler error", http.StatusInternalServerError)
			return
		}
		
		// Send reply
		if reply != nil {
			c.sendReply(w, reply, msg)
			return
		}
	}
	
	// Acknowledge message
	w.WriteHeader(http.StatusOK)
}

// ==================== Message Parsing ====================

func (c *Client) parseMessage(body []byte) (*Message, error) {
	switch c.config.Type {
	case ChannelTypeWecom, ChannelTypeMP:
		return c.parseXMLMessage(body)
	case ChannelTypeClawBot:
		return c.parseClawBotMessage(body)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", c.config.Type)
	}
}

func (c *Client) parseXMLMessage(body []byte) (*Message, error) {
	var xmlMsg WecomMessage
	if err := xml.Unmarshal(body, &xmlMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}
	
	msg := &Message{
		MsgID:      fmt.Sprintf("%d", xmlMsg.MsgID),
		MsgType:    xmlMsg.MsgType,
		Content:    xmlMsg.Content,
		FromUser:   xmlMsg.FromUserName,
		ToUser:     xmlMsg.ToUserName,
		CreateTime: xmlMsg.CreateTime,
	}
	
	// Handle media messages
	if xmlMsg.MediaID != "" {
		msg.MediaID = xmlMsg.MediaID
	}
	
	// Handle events
	if xmlMsg.MsgType == MsgTypeEvent {
		msg.Metadata = map[string]string{
			"event":     xmlMsg.Event,
			"event_key": xmlMsg.EventKey,
		}
	}
	
	return msg, nil
}

func (c *Client) parseClawBotMessage(body []byte) (*Message, error) {
	var clawMsg ClawBotMessage
	if err := json.Unmarshal(body, &clawMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	msg := &Message{
		MsgID:      clawMsg.ID,
		MsgType:    clawMsg.Type,
		Content:    clawMsg.Content,
		FromUser:  clawMsg.FromUser,
		ToUser:    clawMsg.ToUser,
		ChatID:    clawMsg.ChatID,
		ChatType:  clawMsg.ChatType,
		CreateTime: clawMsg.Timestamp,
	}
	
	// Handle attachments
	if len(clawMsg.Attachments) > 0 {
		msg.Metadata = map[string]string{
			"attachments": fmt.Sprintf("%d", len(clawMsg.Attachments)),
		}
	}
	
	return msg, nil
}

// ==================== Signature Verification ====================

func (c *Client) verifySignature(r *http.Request) bool {
	signature := r.URL.Query().Get("signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	
	if signature == "" {
		return false
	}
	
	// Sort token, timestamp, nonce
	params := []string{c.config.Token, timestamp, nonce}
	sort.Strings(params)
	
	// Concatenate and hash
	combined := strings.Join(params, "")
	h := sha256.New()
	h.Write([]byte(combined))
	expectedSig := hex.EncodeToString(h.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// ==================== Send Methods ====================

func (c *Client) sendWecomMessage(ctx context.Context, reply *Reply) error {
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	
	// Build request
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	
	req := map[string]interface{}{
		"touser":  reply.ToUser,
		"msgtype": reply.MsgType,
		"agentid": c.config.AgentID,
	}
	
	switch reply.MsgType {
	case MsgTypeText:
		req["text"] = map[string]string{"content": reply.Content}
	case MsgTypeImage:
		req["image"] = map[string]string{"media_id": reply.MediaID}
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	
	resp, err := c.httpClient.Post(apiURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	
	if result.ErrCode != 0 {
		return fmt.Errorf("API error: %d - %s", result.ErrCode, result.ErrMsg)
	}
	
	return nil
}

func (c *Client) sendClawBotMessage(ctx context.Context, reply *Reply) error {
	resp := ClawBotResponse{
		Type:    reply.MsgType,
		Content: reply.Content,
		ToUser:  reply.ToUser,
	}
	
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	
	apiURL := c.config.ClawBotURL
	if apiURL == "" {
		apiURL = "http://localhost:6174/api/message/send"
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ClawBotToken)
	
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", httpResp.Status)
	}
	
	return nil
}

func (c *Client) sendMPMessage(ctx context.Context, reply *Reply) error {
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", token)
	
	req := map[string]interface{}{
		"touser":  reply.ToUser,
		"msgtype": reply.MsgType,
	}
	
	switch reply.MsgType {
	case MsgTypeText:
		req["text"] = map[string]string{"content": reply.Content}
	case MsgTypeImage:
		req["image"] = map[string]string{"media_id": reply.MediaID}
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	
	resp, err := c.httpClient.Post(apiURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	
	if result.ErrCode != 0 {
		return fmt.Errorf("API error: %d - %s", result.ErrCode, result.ErrMsg)
	}
	
	return nil
}

func (c *Client) sendReply(w http.ResponseWriter, reply *Reply, original *Message) {
	if c.config.Type == ChannelTypeWecom || c.config.Type == ChannelTypeMP {
		// Reply with XML
		xmlReply := WecomReply{
			ToUserName:   original.FromUser,
			FromUserName: original.ToUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      reply.MsgType,
			Content:      reply.Content,
		}
		
		data, err := xml.Marshal(xmlReply)
		if err != nil {
			http.Error(w, "Failed to marshal reply", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/xml")
		w.Write(data)
	} else {
		// Reply with JSON (ClawBot)
		jsonReply, _ := json.Marshal(reply)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonReply)
	}
}

// ==================== Token Management ====================

func (c *Client) refreshAccessToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	
	var apiURL string
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	
	switch c.config.Type {
	case ChannelTypeWecom:
		// Get access token for enterprise WeChat
		apiURL = fmt.Sprintf(
			"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
			c.config.CorpID,
			c.config.CorpSecret,
		)
	case ChannelTypeMP:
		// Get access token for public account
		apiURL = fmt.Sprintf(
			"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
			c.config.AppID,
			c.config.AppSecret,
		)
	default:
		return nil
	}
	
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	
	if result.ErrCode != 0 {
		return fmt.Errorf("failed to get access token: %d - %s", result.ErrCode, result.ErrMsg)
	}
	
	c.accessToken = result.AccessToken
	c.tokenExpireTime = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	
	return nil
}

func (c *Client) getAccessToken() (string, error) {
	c.tokenMu.RLock()
	
	if c.accessToken != "" && time.Now().Before(c.tokenExpireTime) {
		token := c.accessToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()
	
	// Need to refresh
	if err := c.refreshAccessToken(context.Background()); err != nil {
		return "", err
	}
	
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken, nil
}

func (c *Client) tokenRefresher(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshAccessToken(ctx)
		}
	}
}

// ==================== Helper Functions ====================

// ParseWebhookMessage parses a raw webhook message
func ParseWebhookMessage(channelType ChannelType, data []byte) (*Message, error) {
	switch channelType {
	case ChannelTypeWecom, ChannelTypeMP:
		var xmlMsg WecomMessage
		if err := xml.Unmarshal(data, &xmlMsg); err != nil {
			return nil, err
		}
		return &Message{
			MsgID:      fmt.Sprintf("%d", xmlMsg.MsgID),
			MsgType:    xmlMsg.MsgType,
			Content:    xmlMsg.Content,
			FromUser:   xmlMsg.FromUserName,
			ToUser:     xmlMsg.ToUserName,
			CreateTime: xmlMsg.CreateTime,
		}, nil
	case ChannelTypeClawBot:
		var clawMsg ClawBotMessage
		if err := json.Unmarshal(data, &clawMsg); err != nil {
			return nil, err
		}
		return &Message{
			MsgID:      clawMsg.ID,
			MsgType:    clawMsg.Type,
			Content:    clawMsg.Content,
			FromUser:  clawMsg.FromUser,
			ToUser:    clawMsg.ToUser,
			ChatID:    clawMsg.ChatID,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	switch config.Type {
	case ChannelTypeWecom:
		if config.CorpID == "" {
			return fmt.Errorf("corp_id is required for enterprise WeChat")
		}
		if config.CorpSecret == "" {
			return fmt.Errorf("corp_secret is required for enterprise WeChat")
		}
	case ChannelTypeMP:
		if config.AppID == "" {
			return fmt.Errorf("app_id is required for public account")
		}
		if config.AppSecret == "" {
			return fmt.Errorf("app_secret is required for public account")
		}
	case ChannelTypeClawBot:
		// ClawBot doesn't require additional validation
	case ChannelTypeMiniProgram:
		if config.MiniProgramAppID == "" {
			return fmt.Errorf("miniprogram_app_id is required for mini program")
		}
	default:
		return fmt.Errorf("unsupported channel type: %s", config.Type)
	}
	
	return nil
}

// BuildWebhookURL builds the webhook URL for the channel
func BuildWebhookURL(baseURL string, config *Config) string {
	u, _ := url.Parse(baseURL)
	u.Path = config.Path
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}
