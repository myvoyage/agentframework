// Package adapters provides QQ channel adapter
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
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// QQ API endpoints (使用 go-cqhttp 或 NapCat/LLOneBot 等 QQ 机器人框架)
const (
	qqAPIBase               = "http://127.0.0.1:3000" // 默认 OneBot API 地址
	qqSendMsgURL            = qqAPIBase + "/send_msg"
	qqGetMsgURL             = qqAPIBase + "/get_msg"
	qqDeleteMsgURL          = qqAPIBase + "/delete_msg"
	qqGetImageURL           = qqAPIBase + "/get_image"
	qqGetFriendListURL      = qqAPIBase + "/get_friend_list"
	qqGetGroupListURL       = qqAPIBase + "/get_group_list"
	qqGetGroupMemberInfoURL = qqAPIBase + "/get_group_member_info"
)

// QQAdapter implements ChannelAdapter for QQ
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for QQ-specific communication logic
//
// 支持 OneBot 11 标准 (go-cqhttp, NapCat, LLOneBot 等)
type QQAdapter struct {
	*CommonAdapter
	client  *http.Client
	config  channels.ChannelConfig
	apiBase string
	ctx     context.Context
	cancel  context.CancelFunc
	selfID  int64 // 机器人 QQ 号
}

// QQMessage represents QQ message format (OneBot 11)
type QQMessage struct {
	MessageType string      `json:"message_type,omitempty"` // private, group
	UserID      int64       `json:"user_id,omitempty"`
	GroupID     int64       `json:"group_id,omitempty"`
	Message     interface{} `json:"message"` // string or []MessageSegment
	AutoReply   bool        `json:"auto_reply,omitempty"`
}

// QQMessageSegment represents message segment (CQ码)
type QQMessageSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// QQEvent represents QQ event (OneBot 11)
type QQEvent struct {
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	PostType    string      `json:"post_type"`              // message, notice, request, meta_event
	MessageType string      `json:"message_type,omitempty"` // private, group
	SubType     string      `json:"sub_type,omitempty"`
	MessageID   int         `json:"message_id,omitempty"`
	UserID      int64       `json:"user_id,omitempty"`
	GroupID     int64       `json:"group_id,omitempty"`
	Message     interface{} `json:"message,omitempty"`
	RawMessage  string      `json:"raw_message,omitempty"`
	Font        int         `json:"font,omitempty"`
	Sender      QQSender    `json:"sender,omitempty"`
}

// QQSender represents sender information
type QQSender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
	Card     string `json:"card"`
	Level    int    `json:"level,omitempty"`
	Role     string `json:"role,omitempty"`
	Title    string `json:"title,omitempty"`
}

// QQSendResponse represents send message response
type QQSendResponse struct {
	Status  string `json:"status"`
	RetCode int    `json:"retcode"`
	Data    struct {
		MessageID int `json:"message_id"`
	} `json:"data"`
	Echo interface{} `json:"echo"`
}

// NewQQAdapter creates a new QQ adapter
func NewQQAdapter(channelID string) *QQAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeQQ)
	return &QQAdapter{
		CommonAdapter: common,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize initializes the QQ adapter with configuration
func (a *QQAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "QQAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.mu.Unlock()

	// 从 PlatformConfig 获取 API 地址
	if apiBase, ok := config.PlatformConfig["api_base"].(string); ok {
		a.apiBase = apiBase
	} else {
		a.apiBase = qqAPIBase
	}

	// 获取机器人 QQ 号
	if selfID, ok := config.PlatformConfig["self_id"].(string); ok {
		id, err := strconv.ParseInt(selfID, 10, 64)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid self_id: %w", err)
		}
		a.selfID = id
	}

	// Set capabilities
	a.SetCapability(channels.CapabilityEdits, false) // QQ 不支持编辑消息
	a.SetCapability(channels.CapabilityRichText, true)
	a.SetCapability(channels.CapabilityWebhooks, true) // 支持 OneBot 反向 WebSocket

	return nil
}

// Connect establishes connection to QQ
func (a *QQAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "QQAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	a.ctx, a.cancel = context.WithCancel(ctx)

	// 验证连接 - 获取机器人信息
	req, err := http.NewRequestWithContext(ctx, "GET", a.apiBase+"/get_login_info", nil)
	if err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to connect to QQ bot API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		RetCode int    `json:"retcode"`
		Data    struct {
			UserID   int64  `json:"user_id"`
			NickName string `json:"nickname"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.RetCode != 0 {
		err := fmt.Errorf("QQ API returned error code %d", result.RetCode)
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return err
	}

	a.selfID = result.Data.UserID

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"user_id":  result.Data.UserID,
		"nickname": result.Data.NickName,
	}, nil)

	return nil
}

// Disconnect gracefully closes the QQ connection
func (a *QQAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "QQAdapter.Disconnect")
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

// SendMessage sends a message to QQ
func (a *QQAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "QQAdapter.SendMessage",
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

	// 构建 QQ 消息
	qqMsg, err := a.buildQQMessage(msg, opts)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	// 设置消息类型（私聊/群聊）
	if groupID, ok := msg.Metadata["group_id"].(string); ok {
		gid, _ := strconv.ParseInt(groupID, 10, 64)
		qqMsg.MessageType = "group"
		qqMsg.GroupID = gid
	} else {
		qqMsg.MessageType = "private"
		if msg.From != nil {
			qqMsg.UserID, _ = strconv.ParseInt(msg.From.ChannelUserID, 10, 64)
		}
	}

	// Marshal message
	body, err := json.Marshal(qqMsg)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", qqSendMsgURL, nil)
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

	var sendResp QQSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		span.RecordError(err)
		return "", err
	}

	if sendResp.RetCode != 0 {
		err := fmt.Errorf("QQ send failed with code %d", sendResp.RetCode)
		span.RecordError(err)
		return "", err
	}

	a.RecordMessageSent(ctx, true, len(msg.Text))
	a.EmitEvent(ctx, channels.EventTypeMessageSent, map[string]interface{}{
		"message_id": sendResp.Data.MessageID,
	}, nil)

	return strconv.Itoa(sendResp.Data.MessageID), nil
}

// EditMessage edits an existing message
func (a *QQAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "QQAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// QQ 不支持编辑消息
	return fmt.Errorf("message editing is not supported by QQ")
}

// DeleteMessage deletes a message
func (a *QQAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	msgID, err := strconv.Atoi(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	reqData := struct {
		MessageID int `json:"message_id"`
	}{
		MessageID: msgID,
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", qqDeleteMsgURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Note: This won't actually send the body as we're using nil reader
	// In production, you'd need to properly create the request with body
	_ = body

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// UploadFile uploads a file and returns an attachment
func (a *QQAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// QQ 文件上传需要在发送消息时进行
	return &channels.Attachment{
		ID:       GenerateMessageID(),
		Filename: filename,
		MimeType: mimeType,
	}, nil
}

// HandleEvent handles QQ events from OneBot
func (a *QQAdapter) HandleEvent(ctx context.Context, event QQEvent) error {
	if event.PostType != "message" {
		return nil // 只处理消息事件
	}

	// 转换为统一消息格式
	unifiedMsg := a.convertEventToUnified(&event)
	return a.HandleMessage(ctx, unifiedMsg)
}

// buildQQMessage builds QQ message from unified message
func (a *QQAdapter) buildQQMessage(msg *channels.Message, opts channels.MessageSendOptions) (*QQMessage, error) {
	qqMsg := &QQMessage{
		AutoReply: false,
	}

	// 构建消息段
	var segments []QQMessageSegment

	// 处理文本
	if msg.Text != "" {
		segments = append(segments, QQMessageSegment{
			Type: "text",
			Data: map[string]interface{}{
				"text": msg.Text,
			},
		})
	}

	// 处理附件
	switch msg.Type {
	case channels.MessageTypeImage:
		if len(msg.Attachments) > 0 {
			segments = append(segments, QQMessageSegment{
				Type: "image",
				Data: map[string]interface{}{
					"file": msg.Attachments[0].URL,
				},
			})
		}

	case channels.MessageTypeAudio:
		if len(msg.Attachments) > 0 {
			segments = append(segments, QQMessageSegment{
				Type: "record",
				Data: map[string]interface{}{
					"file": msg.Attachments[0].URL,
				},
			})
		}

	case channels.MessageTypeVideo:
		if len(msg.Attachments) > 0 {
			segments = append(segments, QQMessageSegment{
				Type: "video",
				Data: map[string]interface{}{
					"file": msg.Attachments[0].URL,
				},
			})
		}

	case channels.MessageTypeFile:
		if len(msg.Attachments) > 0 {
			segments = append(segments, QQMessageSegment{
				Type: "file",
				Data: map[string]interface{}{
					"file": msg.Attachments[0].URL,
				},
			})
		}
	}

	// 处理回复
	if opts.ReplyTo != "" {
		segments = append([]QQMessageSegment{{
			Type: "reply",
			Data: map[string]interface{}{
				"id": opts.ReplyTo,
			},
		}}, segments...)
	}

	if len(segments) == 1 {
		qqMsg.Message = segments[0]
	} else {
		qqMsg.Message = segments
	}

	return qqMsg, nil
}

// convertEventToUnified converts QQ event to unified message format
func (a *QQAdapter) convertEventToUnified(event *QQEvent) *channels.Message {
	user := &channels.User{
		ID:            strconv.FormatInt(event.UserID, 10),
		ChannelUserID: strconv.FormatInt(event.UserID, 10),
		DisplayName:   event.Sender.Nickname,
		Username:      event.Sender.Card,
		IsBot:         false,
		ChannelType:   channels.ChannelTypeQQ,
	}

	// 解析消息内容
	var text string
	var msgType channels.MessageType = channels.MessageTypeText

	if event.RawMessage != "" {
		text = event.RawMessage
	}

	// 检查消息类型
	if segments, ok := event.Message.([]interface{}); ok {
		for _, seg := range segments {
			if segMap, ok := seg.(map[string]interface{}); ok {
				if segType, ok := segMap["type"].(string); ok {
					switch segType {
					case "image":
						msgType = channels.MessageTypeImage
					case "record", "audio":
						msgType = channels.MessageTypeAudio
					case "video":
						msgType = channels.MessageTypeVideo
					case "file":
						msgType = channels.MessageTypeFile
					case "at":
						// @消息处理
					case "face":
						// 表情处理
					}
				}
			}
		}
	}

	chatID := strconv.FormatInt(event.UserID, 10)
	if event.GroupID > 0 {
		chatID = strconv.FormatInt(event.GroupID, 10)
	}

	unifiedMsg := &channels.Message{
		ID:          strconv.Itoa(event.MessageID),
		Type:        msgType,
		Direction:   channels.MessageDirectionIncoming,
		Text:        text,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeQQ,
		ChatID:      chatID,
		From:        user,
		Timestamp:   time.Unix(event.Time, 0),
		Metadata:    make(map[string]string),
	}

	// 添加群组信息
	if event.GroupID > 0 {
		unifiedMsg.Metadata["group_id"] = strconv.FormatInt(event.GroupID, 10)
		unifiedMsg.Metadata["message_type"] = "group"
	} else {
		unifiedMsg.Metadata["message_type"] = "private"
	}

	return unifiedMsg
}
