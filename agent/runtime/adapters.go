// Channel Adapters - Built-in channel adapters for Gateway Bridge
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"encoding/json"
	"fmt"
	"time"
)

// HTTPAdapter handles HTTP/REST requests
type HTTPAdapter struct {
	name string
}

// NewHTTPAdapter creates a new HTTP adapter
func NewHTTPAdapter() *HTTPAdapter {
	return &HTTPAdapter{name: "http"}
}

func (a *HTTPAdapter) Name() string { return a.name }

func (a *HTTPAdapter) ParseRequest(data []byte) (*GatewayRequest, error) {
	var req struct {
		SessionID string                 `json:"session_id"`
		AgentName string                 `json:"agent_name"`
		AgentType string                 `json:"agent_type"`
		Input     string                 `json:"input"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse HTTP request: %w", err)
	}

	return &GatewayRequest{
		ID:         generateID(),
		Channel:    "http",
		SessionID:  req.SessionID,
		AgentName:  req.AgentName,
		AgentType:  req.AgentType,
		Input:      req.Input,
		Metadata:   req.Metadata,
		Timestamp:  time.Now(),
	}, nil
}

func (a *HTTPAdapter) FormatResponse(resp *GatewayResponse) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"request_id": resp.RequestID,
		"content":    resp.Content,
		"error":      resp.Error,
		"duration":   resp.Duration.Milliseconds(),
		"metadata":   resp.Metadata,
	})
}

func (a *HTTPAdapter) ValidateAuth(data []byte) (string, error) {
	// Parse Authorization header or token from request
	var req struct {
		UserID  string `json:"user_id"`
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return "", fmt.Errorf("invalid auth format")
	}
	return req.UserID, nil
}

// WebSocketAdapter handles WebSocket connections
type WebSocketAdapter struct {
	name string
}

// NewWebSocketAdapter creates a new WebSocket adapter
func NewWebSocketAdapter() *WebSocketAdapter {
	return &WebSocketAdapter{name: "websocket"}
}

func (a *WebSocketAdapter) Name() string { return a.name }

func (a *WebSocketAdapter) ParseRequest(data []byte) (*GatewayRequest, error) {
	var req struct {
		SessionID string                 `json:"session_id"`
		AgentName string                 `json:"agent_name"`
		AgentType string                 `json:"agent_type"`
		Input     string                 `json:"input"`
		Type      string                 `json:"type"` // "text", "command", etc.
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket request: %w", err)
	}

	return &GatewayRequest{
		ID:         generateID(),
		Channel:    "websocket",
		SessionID:  req.SessionID,
		AgentName:  req.AgentName,
		AgentType:  req.AgentType,
		Input:      req.Input,
		Metadata:   req.Metadata,
		Timestamp:  time.Now(),
	}, nil
}

func (a *WebSocketAdapter) FormatResponse(resp *GatewayResponse) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":       "response",
		"request_id": resp.RequestID,
		"content":    resp.Content,
		"error":      resp.Error,
		"timestamp":  time.Now().Unix(),
	})
}

func (a *WebSocketAdapter) ValidateAuth(data []byte) (string, error) {
	var req struct {
		UserID string `json:"user_id"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return "", fmt.Errorf("invalid auth format")
	}
	return req.UserID, nil
}

// WeChatMiniProgramAdapter handles WeChat Mini Program requests
type WeChatMiniProgramAdapter struct {
	name string
}

// NewWeChatMiniProgramAdapter creates a new WeChat Mini Program adapter
func NewWeChatMiniProgramAdapter() *WeChatMiniProgramAdapter {
	return &WeChatMiniProgramAdapter{name: "wechat_miniprogram"}
}

func (a *WeChatMiniProgramAdapter) Name() string { return a.name }

func (a *WeChatMiniProgramAdapter) ParseRequest(data []byte) (*GatewayRequest, error) {
	var req struct {
		OpenID    string                 `json:"openid"`
		SessionID string                 `json:"session_id"`
		AgentName string                 `json:"agent_name"`
		Input     string                 `json:"input"`
		Type      string                 `json:"type"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse WeChat request: %w", err)
	}

	return &GatewayRequest{
		ID:         generateID(),
		Channel:    "wechat_miniprogram",
		UserID:     req.OpenID,
		SessionID:  req.SessionID,
		AgentName:  req.AgentName,
		AgentType:  "chat",
		Input:      req.Input,
		Metadata:   req.Metadata,
		Timestamp:  time.Now(),
	}, nil
}

func (a *WeChatMiniProgramAdapter) FormatResponse(resp *GatewayResponse) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"request_id": resp.RequestID,
		"content":    resp.Content,
		"error":      resp.Error,
		"code":       0,
		"msg":        "success",
	})
}

func (a *WeChatMiniProgramAdapter) ValidateAuth(data []byte) (string, error) {
	var req struct {
		OpenID string `json:"openid"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return "", fmt.Errorf("invalid WeChat auth")
	}
	return req.OpenID, nil
}

// TelegramAdapter handles Telegram bot requests
type TelegramAdapter struct {
	name string
}

// NewTelegramAdapter creates a new Telegram adapter
func NewTelegramAdapter() *TelegramAdapter {
	return &TelegramAdapter{name: "telegram"}
}

func (a *TelegramAdapter) Name() string { return a.name }

func (a *TelegramAdapter) ParseRequest(data []byte) (*GatewayRequest, error) {
	var update struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int    `json:"message_id"`
			From      struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID   int    `json:"id"`
				Type string `json:"type"`
			} `json:"chat"`
			Text string `json:"text"`
		} `json:"message"`
	}

	if err := json.Unmarshal(data, &update); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram request: %w", err)
	}

	return &GatewayRequest{
		ID:         fmt.Sprintf("tg_%d", update.UpdateID),
		Channel:    "telegram",
		UserID:     fmt.Sprintf("tg_%d", update.Message.From.ID),
		SessionID:  fmt.Sprintf("tg_chat_%d", update.Message.Chat.ID),
		AgentName:  "default",
		AgentType:  "chat",
		Input:      update.Message.Text,
		Metadata: map[string]interface{}{
			"telegram_update_id": update.UpdateID,
			"telegram_chat_id":  update.Message.Chat.ID,
			"telegram_user":      update.Message.From.Username,
		},
		Timestamp: time.Now(),
	}, nil
}

func (a *TelegramAdapter) FormatResponse(resp *GatewayResponse) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"method":     "sendMessage",
		"chat_id":    resp.Metadata["telegram_chat_id"],
		"text":       resp.Content,
		"parse_mode": "Markdown",
	})
}

func (a *TelegramAdapter) ValidateAuth(data []byte) (string, error) {
	// Telegram validates via webhook secret, not per-request
	var update struct {
		Message struct {
			From struct {
				ID int `json:"id"`
			} `json:"from"`
		} `json:"message"`
	}
	if err := json.Unmarshal(data, &update); err != nil {
		return "", fmt.Errorf("invalid Telegram format")
	}
	return fmt.Sprintf("tg_%d", update.Message.From.ID), nil
}

// SlackAdapter handles Slack app requests
type SlackAdapter struct {
	name string
}

// NewSlackAdapter creates a new Slack adapter
func NewSlackAdapter() *SlackAdapter {
	return &SlackAdapter{name: "slack"}
}

func (a *SlackAdapter) Name() string { return a.name }

func (a *SlackAdapter) ParseRequest(data []byte) (*GatewayRequest, error) {
	var event struct {
		Type    string `json:"type"`
		Team    string `json:"team"`
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text"`
		TS      string `json:"ts"`
		ThreadTS string `json:"thread_ts"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse Slack request: %w", err)
	}

	sessionID := event.Channel
	if event.ThreadTS != "" {
		sessionID = event.ThreadTS
	}

	return &GatewayRequest{
		ID:         fmt.Sprintf("slack_%s", event.TS),
		Channel:    "slack",
		UserID:     event.User,
		SessionID:  sessionID,
		AgentName:  "default",
		AgentType:  "chat",
		Input:      event.Text,
		Metadata: map[string]interface{}{
			"slack_team":    event.Team,
			"slack_channel": event.Channel,
			"slack_ts":      event.TS,
			"slack_thread":  event.ThreadTS,
		},
		Timestamp: time.Now(),
	}, nil
}

func (a *SlackAdapter) FormatResponse(resp *GatewayResponse) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"channel":   resp.Metadata["slack_channel"],
		"text":      resp.Content,
		"thread_ts":  resp.Metadata["slack_thread"],
	})
}

func (a *SlackAdapter) ValidateAuth(data []byte) (string, error) {
	var event struct {
		User string `json:"user"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return "", fmt.Errorf("invalid Slack format")
	}
	return event.User, nil
}

// Helper function
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// RegisterBuiltinAdapters registers all built-in adapters
func RegisterBuiltinAdapters(bridge *GatewayBridge) {
	bridge.RegisterAdapter(NewHTTPAdapter())
	bridge.RegisterAdapter(NewWebSocketAdapter())
	bridge.RegisterAdapter(NewWeChatMiniProgramAdapter())
	bridge.RegisterAdapter(NewTelegramAdapter())
	bridge.RegisterAdapter(NewSlackAdapter())
}
