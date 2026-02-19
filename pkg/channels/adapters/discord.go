// Package adapters provides Discord channel adapter
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
	"time"

	"github.com/bwmarrin/discordgo"
	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// DiscordAdapter implements ChannelAdapter for Discord
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for Discord-specific communication logic
type DiscordAdapter struct {
	*CommonAdapter
	session *discordgo.Session
	config  channels.ChannelConfig
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewDiscordAdapter creates a new Discord adapter
func NewDiscordAdapter(channelID string) *DiscordAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeDiscord)
	return &DiscordAdapter{
		CommonAdapter: common,
	}
}

// Initialize initializes the Discord adapter with configuration
func (a *DiscordAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "DiscordAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.mu.Unlock()

	// Create Discord session
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create discord session: %w", err)
	}

	a.session = session

	// Set capabilities
	a.SetCapability(channels.CapabilityThreads, true)
	a.SetCapability(channels.CapabilityEdits, true)
	a.SetCapability(channels.CapabilityReactions, true)
	a.SetCapability(channels.CapabilityTypingIndicator, true)
	a.SetCapability(channels.CapabilityRichText, true)

	return nil
}

// Connect establishes connection to Discord
func (a *DiscordAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "DiscordAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.session == nil {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	// Set up handlers
	a.setupHandlers()

	// Open connection
	if err := a.session.Open(); err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to open discord connection: %w", err)
	}

	// Wait for ready
	_ready := make(chan struct{})
	go func() {
		a.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			close(_ready)
		})
	}()

	select {
	case <-_ready:
		a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
		a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
			"bot_id": a.session.State.User.ID,
		}, nil)
		return nil
	case <-time.After(30 * time.Second):
		a.session.Close()
		return fmt.Errorf("timeout waiting for discord ready")
	case <-ctx.Done():
		a.session.Close()
		return ctx.Err()
	}
}

// Disconnect gracefully closes the Discord connection
func (a *DiscordAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "DiscordAdapter.Disconnect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.session != nil {
		if err := a.session.Close(); err != nil {
			span.RecordError(err)
			return err
		}
	}

	if a.cancel != nil {
		a.cancel()
	}

	a.SetStatus(context.Background(), channels.ChannelStatusDisconnected, "")
	a.EmitEvent(ctx, channels.EventTypeDisconnected, nil, nil)

	return nil
}

// SendMessage sends a message to Discord
func (a *DiscordAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "DiscordAdapter.SendMessage",
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

	// Determine channel ID
	channelID := a.extractChannelID(msg)
	if channelID == "" {
		err := fmt.Errorf("no valid channel ID found")
		span.RecordError(err)
		return "", err
	}

	var sendMsg *discordgo.MessageSend
	var err error

	switch msg.Type {
	case channels.MessageTypeText:
		sendMsg = a.buildTextMessage(msg, opts)
	case channels.MessageTypeImage:
		sendMsg, err = a.buildImageMessage(msg, opts)
	case channels.MessageTypeAudio:
		sendMsg, err = a.buildAudioMessage(msg, opts)
	case channels.MessageTypeVideo:
		sendMsg, err = a.buildVideoMessage(msg, opts)
	case channels.MessageTypeFile:
		sendMsg, err = a.buildFileMessage(msg, opts)
	default:
		err = fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	if err != nil {
		span.RecordError(err)
		return "", err
	}

	message, err := a.session.ChannelMessageSendComplex(channelID, sendMsg)
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
		"message_id": message.ID,
		"channel_id": channelID,
	}, nil)

	return message.ID, nil
}

// EditMessage edits an existing message
func (a *DiscordAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "DiscordAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	channelID := a.extractChannelID(msg)
	if channelID == "" {
		return fmt.Errorf("no valid channel ID found")
	}

	edit := &discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Content: &msg.Text,
	}

	_, err := a.session.ChannelMessageEditComplex(edit)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (a *DiscordAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// Need channel ID to delete - this is a limitation of the unified interface
	// In practice, you'd need to store the channel ID with the message
	return fmt.Errorf("deletion requires channel ID (not available in unified interface)")
}

// UploadFile uploads a file and returns an attachment
func (a *DiscordAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// Discord uploads files when sending messages
	return &channels.Attachment{
		ID:       GenerateMessageID(),
		Filename: filename,
		MimeType: mimeType,
	}, nil
}

// setupHandlers sets up Discord event handlers
func (a *DiscordAdapter) setupHandlers() {
	// Handle message create
	a.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot || m.Author.ID == s.State.User.ID {
			return // Ignore bot messages and self
		}

		unifiedMsg := a.convertToUnifiedMessage(m.Message)
		_ = a.HandleMessage(a.ctx, unifiedMsg)
	})

	// Handle message update (edit)
	a.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author == nil {
			return
		}

		unifiedMsg := a.convertToUnifiedMessage(m.Message)
		unifiedMsg.Edited = true
		_ = a.HandleMessage(a.ctx, unifiedMsg)
	})
}

// convertToUnifiedMessage converts Discord message to unified format
func (a *DiscordAdapter) convertToUnifiedMessage(msg *discordgo.Message) *channels.Message {
	user := &channels.User{
		ID:            msg.Author.ID,
		ChannelUserID: msg.Author.ID,
		Username:      msg.Author.Username,
		DisplayName:   msg.Author.GlobalName || msg.Author.Username,
		IsBot:         msg.Author.Bot,
		ChannelType:   channels.ChannelTypeDiscord,
	}

	unifiedMsg := &channels.Message{
		ID:          msg.ID,
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionIncoming,
		Text:        msg.Content,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeDiscord,
		ChatID:      msg.ChannelID,
		From:        user,
		Timestamp:   msg.Timestamp,
		Metadata:    make(map[string]string),
	}

	// Handle attachments
	if len(msg.Attachments) > 0 {
		unifiedMsg.Attachments = make([]channels.Attachment, len(msg.Attachments))
		for i, att := range msg.Attachments {
			unifiedMsg.Attachments[i] = channels.Attachment{
				ID:       att.ID,
				Filename: att.Filename,
				URL:      att.URL,
				Size:     int64(att.Size),
				MimeType: att.ContentType,
			}
		}

		// Set message type based on first attachment
		switch msg.Attachments[0].ContentType {
		case "image/png", "image/jpeg", "image/gif", "image/webp":
			unifiedMsg.Type = channels.MessageTypeImage
		case "video/mp4", "video/webm", "video/mov":
			unifiedMsg.Type = channels.MessageTypeVideo
		case "audio/mpeg", "audio/ogg", "audio/wav":
			unifiedMsg.Type = channels.MessageTypeAudio
		default:
			unifiedMsg.Type = channels.MessageTypeFile
		}
	}

	// Handle references (replies)
	if msg.Reference != nil && msg.Reference.MessageID != "" {
		unifiedMsg.ReplyToID = msg.Reference.MessageID
	}

	// Handle edits
	if msg.EditedTimestamp != nil {
		unifiedMsg.Edited = true
		unifiedMsg.EditedAt = msg.EditedTimestamp
	}

	return unifiedMsg
}

// extractChannelID extracts channel ID from message metadata
func (a *DiscordAdapter) extractChannelID(msg *channels.Message) string {
	// Try metadata first
	if channelID, ok := msg.Metadata["channel_id"]; ok {
		return channelID
	}

	// Try ChatID field
	return msg.ChatID
}

// buildTextMessage builds a Discord text message
func (a *DiscordAdapter) buildTextMessage(msg *channels.Message, opts channels.MessageSendOptions) *discordgo.MessageSend {
	send := &discordgo.MessageSend{
		Content: msg.Text,
	}

	if opts.ReplyTo != "" {
		send.Reference = &discordgo.MessageReference{
			MessageID: opts.ReplyTo,
		}
	}

	return send
}

// buildImageMessage builds a Discord image message
func (a *DiscordAdapter) buildImageMessage(msg *channels.Message, opts channels.MessageSendOptions) (*discordgo.MessageSend, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	att := msg.Attachments[0]

	send := &discordgo.MessageSend{
		Content: msg.Text,
		Files: []*discordgo.File{
			{
				Name:   att.Filename,
				URL:    att.URL,
				Reader: nil, // Would need io.Reader for actual upload
			},
		},
	}

	if opts.ReplyTo != "" {
		send.Reference = &discordgo.MessageReference{
			MessageID: opts.ReplyTo,
		}
	}

	return send, nil
}

// buildAudioMessage builds a Discord audio message
func (a *DiscordAdapter) buildAudioMessage(msg *channels.Message, opts channels.MessageSendOptions) (*discordgo.MessageSend, error) {
	return a.buildFileMessage(msg, opts) // Same as file for Discord
}

// buildVideoMessage builds a Discord video message
func (a *DiscordAdapter) buildVideoMessage(msg *channels.Message, opts channels.MessageSendOptions) (*discordgo.MessageSend, error) {
	return a.buildFileMessage(msg, opts) // Same as file for Discord
}

// buildFileMessage builds a Discord file message
func (a *DiscordAdapter) buildFileMessage(msg *channels.Message, opts channels.MessageSendOptions) (*discordgo.MessageSend, error) {
	if len(msg.Attachments) == 0 {
		return nil, fmt.Errorf("no attachment provided")
	}

	att := msg.Attachments[0]

	send := &discordgo.MessageSend{
		Content: msg.Text,
		Files: []*discordgo.File{
			{
				Name:   att.Filename,
				URL:    att.URL,
				Reader: nil, // Would need io.Reader for actual upload
			},
		},
	}

	if opts.ReplyTo != "" {
		send.Reference = &discordgo.MessageReference{
			MessageID: opts.ReplyTo,
		}
	}

	return send, nil
}

// GetSession returns the underlying Discord session
func (a *DiscordAdapter) GetSession() *discordgo.Session {
	return a.session
}
