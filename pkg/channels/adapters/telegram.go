// Package adapters provides Telegram channel adapter
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
	"fmt"
	"io"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
	tb "gopkg.in/telebot.v3"

	"AgentFramework/pkg/channels"
)

// TelegramAdapter implements ChannelAdapter for Telegram
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for Telegram-specific communication logic
type TelegramAdapter struct {
	*CommonAdapter
	bot    *tb.Bot
	config channels.ChannelConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTelegramAdapter creates a new Telegram adapter
func NewTelegramAdapter(channelID string) *TelegramAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeTelegram)

	return &TelegramAdapter{
		CommonAdapter: common,
	}
}

// Initialize initializes the Telegram adapter with configuration
func (a *TelegramAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "TelegramAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.mu.Unlock()

	// Create Telegram bot
	poller := &tb.LongPoller{Timeout: 10 * time.Second}
	if config.WebhookURL != "" {
		// Webhook mode will be configured in Connect
		poller = nil
	}

	settings := tb.Settings{
		Token:  config.Token,
		Poller: poller,
		// Use offline mode for receiving updates without sending
		Offline: true,
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	a.bot = bot

	// Set capabilities
	a.SetCapability(channels.CapabilityEdits, true)
	a.SetCapability(channels.CapabilityPolling, true)
	a.SetCapability(channels.CapabilityInlineKeyboard, true)

	if config.WebhookURL != "" {
		a.SetCapability(channels.CapabilityWebhooks, true)
	}

	return nil
}

// Connect establishes connection to Telegram
func (a *TelegramAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "TelegramAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.bot == nil {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	// Set up message handlers
	a.setupHandlers()

	// Start bot in background
	go func() {
		if err := a.bot.Start(); err != nil {
			a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		}
	}()

	// Verify bot is working
	bot, err := a.bot.Me()
	if err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	_ = bot // Bot info retrieved successfully

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"bot_username": bot.Username,
	}, nil)

	return nil
}

// Disconnect gracefully closes the Telegram connection
func (a *TelegramAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "TelegramAdapter.Disconnect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.bot != nil {
		a.bot.Stop()
	}

	if a.cancel != nil {
		a.cancel()
	}

	a.SetStatus(context.Background(), channels.ChannelStatusDisconnected, "")
	a.EmitEvent(ctx, channels.EventTypeDisconnected, nil, nil)

	return nil
}

// SendMessage sends a message to Telegram
func (a *TelegramAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "TelegramAdapter.SendMessage",
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

	// Determine target
	chatID, err := a.extractChatID(msg)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	// Send based on message type
	var sentMsg *tb.Message
	switch msg.Type {
	case channels.MessageTypeText:
		sentMsg, err = a.sendText(ctx, chatID, msg, opts)
	case channels.MessageTypeImage:
		sentMsg, err = a.sendPhoto(ctx, chatID, msg, opts)
	case channels.MessageTypeAudio:
		sentMsg, err = a.sendAudio(ctx, chatID, msg, opts)
	case channels.MessageTypeVideo:
		sentMsg, err = a.sendVideo(ctx, chatID, msg, opts)
	case channels.MessageTypeFile:
		sentMsg, err = a.sendDocument(ctx, chatID, msg, opts)
	default:
		err = fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	if err != nil {
		span.RecordError(err)
		a.RecordMessageSent(ctx, false, 0)
		a.EmitEvent(ctx, channels.EventTypeMessageFailed, map[string]interface{}{
			"message_id": msg.ID,
			"error":      err.Error(),
		}, err)
		return "", err
	}

	a.RecordMessageSent(ctx, true, len(msg.Text))
	a.EmitEvent(ctx, channels.EventTypeMessageSent, map[string]interface{}{
		"message_id": sentMsg.ID,
		"chat_id":    chatID,
	}, nil)

	return strconv.Itoa(sentMsg.ID), nil
}

// EditMessage edits an existing message
func (a *TelegramAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "TelegramAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// Parse message ID
	msgID, err := strconv.Atoi(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	// Get editable message
	editable := &tb.Message{
		ID: msgID,
		// ChatID would be needed but not available in this interface
	}

	// Edit text
	_, err = a.bot.Edit(editable, msg.Text)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (a *TelegramAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	msgID, err := strconv.Atoi(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	deletable := &tb.Message{
		ID: msgID,
	}

	err = a.bot.Delete(deletable)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// UploadFile uploads a file and returns an attachment
func (a *TelegramAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// Telegram doesn't have a separate upload API
	// Files are uploaded when sending messages
	return &channels.Attachment{
		ID:       GenerateMessageID(),
		Filename: filename,
		MimeType: mimeType,
	}, nil
}

// setupHandlers sets up message handlers for the bot
func (a *TelegramAdapter) setupHandlers() {
	// Handle text messages
	a.bot.Handle(tb.OnText, func(c tb.Context) error {
		msg := c.Message()
		if msg == nil {
			return nil
		}

		// Convert to unified message format
		unifiedMsg := a.convertToUnifiedMessage(msg)

		// Handle the message
		return a.HandleMessage(a.ctx, unifiedMsg)
	})

	// Handle photos
	a.bot.Handle(tb.OnPhoto, func(c tb.Context) error {
		msg := c.Message()
		if msg == nil {
			return nil
		}

		unifiedMsg := a.convertToUnifiedMessage(msg)
		unifiedMsg.Type = channels.MessageTypeImage

		return a.HandleMessage(a.ctx, unifiedMsg)
	})

	// Handle commands
	a.bot.Handle(tb.OnCommand, func(c tb.Context) error {
		msg := c.Message()
		if msg == nil {
			return nil
		}

		unifiedMsg := a.convertToUnifiedMessage(msg)
		unifiedMsg.Type = channels.MessageTypeCommand

		return a.HandleMessage(a.ctx, unifiedMsg)
	})
}

// convertToUnifiedMessage converts Telegram message to unified format
func (a *TelegramAdapter) convertToUnifiedMessage(msg *tb.Message) *channels.Message {
	now := time.Now()

	user := &channels.User{
		ID:            strconv.FormatInt(msg.Sender.ID, 10),
		ChannelUserID: strconv.FormatInt(msg.Sender.ID, 10),
		Username:      msg.Sender.Username,
		DisplayName:   msg.Sender.FirstName + " " + msg.Sender.LastName,
		IsBot:         msg.Sender.IsBot,
		ChannelType:   channels.ChannelTypeTelegram,
	}

	if user.DisplayName == " " {
		user.DisplayName = msg.Sender.Username
	}

	unifiedMsg := &channels.Message{
		ID:          strconv.Itoa(msg.ID),
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionIncoming,
		Text:        msg.Text,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeTelegram,
		ChatID:      strconv.FormatInt(msg.Chat.ID, 10),
		From:        user,
		Timestamp:   now,
		Metadata:    make(map[string]string),
	}

	// Handle reply
	if msg.ReplyTo != nil {
		unifiedMsg.ReplyToID = strconv.Itoa(msg.ReplyTo.ID)
	}

	// Handle edit
	if msg.EditDate != nil {
		unifiedMsg.Edited = true
		editedAt := msg.EditDate.Time
		unifiedMsg.EditedAt = &editedAt
	}

	return unifiedMsg
}

// extractChatID extracts chat ID from message metadata
func (a *TelegramAdapter) extractChatID(msg *channels.Message) (int64, error) {
	// Try metadata first
	if chatIDStr, ok := msg.Metadata["chat_id"]; ok {
		return strconv.ParseInt(chatIDStr, 10, 64)
	}

	// Try ChatID field
	if msg.ChatID != "" {
		return strconv.ParseInt(msg.ChatID, 10, 64)
	}

	// Try From field
	if msg.To != nil && len(msg.To) > 0 {
		return strconv.ParseInt(msg.To[0].ChannelUserID, 10, 64)
	}

	return 0, fmt.Errorf("no valid chat ID found in message")
}

// sendText sends a text message
func (a *TelegramAdapter) sendText(ctx context.Context, chatID int64, msg *channels.Message, opts channels.MessageSendOptions) (*tb.Message, error) {
	options := a.buildSendOptions(opts)

	return a.bot.Send(&tb.Chat{ID: chatID}, msg.Text, options...)
}

// sendPhoto sends a photo
func (a *TelegramAdapter) sendPhoto(ctx context.Context, chatID int64, msg *channels.Message, opts channels.MessageSendOptions) (*tb.Message, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	attachment := msg.Attachments[0]
	options := a.buildSendOptions(opts)

	// Add caption
	if msg.Text != "" {
		options = append(options, tb.Caption(msg.Text))
	}

	return a.bot.Send(&tb.Chat{ID: chatID}, &tb.Photo{
		File:    tb.FromURL(attachment.URL),
		Caption: msg.Text,
	}, options...)
}

// sendAudio sends an audio file
func (a *TelegramAdapter) sendAudio(ctx context.Context, chatID int64, msg *channels.Message, opts channels.MessageSendOptions) (*tb.Message, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	attachment := msg.Attachments[0]
	options := a.buildSendOptions(opts)

	return a.bot.Send(&tb.Chat{ID: chatID}, &tb.Audio{
		File:    tb.FromURL(attachment.URL),
		Caption: msg.Text,
	}, options...)
}

// sendVideo sends a video
func (a *TelegramAdapter) sendVideo(ctx context.Context, chatID int64, msg *channels.Message, opts channels.MessageSendOptions) (*tb.Message, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	attachment := msg.Attachments[0]
	options := a.buildSendOptions(opts)

	return a.bot.Send(&tb.Chat{ID: chatID}, &tb.Video{
		File:    tb.FromURL(attachment.URL),
		Caption: msg.Text,
	}, options...)
}

// sendDocument sends a document
func (a *TelegramAdapter) sendDocument(ctx context.Context, chatID int64, msg *channels.Message, opts channels.MessageSendOptions) (*tb.Message, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	attachment := msg.Attachments[0]
	options := a.buildSendOptions(opts)

	return a.bot.Send(&tb.Chat{ID: chatID}, &tb.Document{
		File:    tb.FromURL(attachment.URL),
		Caption: msg.Text,
	}, options...)
}

// buildSendOptions builds send options from MessageSendOptions
func (a *TelegramAdapter) buildSendOptions(opts channels.MessageSendOptions) []interface{} {
	var options []interface{}

	if opts.ParseMode != "" {
		switch opts.ParseMode {
		case "markdown":
			options = append(options, tb.ModeMarkdown)
		case "html":
			options = append(options, tb.ModeHTML)
		}
	} else {
		// Default to Markdown
		options = append(options, tb.ModeMarkdown)
	}

	if opts.DisableWebPagePreview {
		options = append(options, tb.NoPreview)
	}

	if opts.DisableNotification {
		options = append(options, tb.Silent)
	}

	// Handle reply
	if opts.ReplyTo != "" {
		msgID, err := strconv.Atoi(opts.ReplyTo)
		if err == nil {
			options = append(options, tb.ReplyTo(msgID))
		}
	}

	return options
}

// GetBot returns the underlying Telegram bot instance
func (a *TelegramAdapter) GetBot() *tb.Bot {
	return a.bot
}
