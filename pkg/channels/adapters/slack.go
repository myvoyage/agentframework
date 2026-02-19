// Package adapters provides Slack channel adapter
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

	"github.com/slack-go/slack"
	"go.opentelemetry.io/otel/trace"

	"AgentFramework/pkg/channels"
)

// SlackAdapter implements ChannelAdapter for Slack
//
// SOLID - Single Responsibility Principle (SRP):
// Only responsible for Slack-specific communication logic
type SlackAdapter struct {
	*CommonAdapter
	client    *slack.Client
	socket    *slack.SocketModeClient
	config    channels.ChannelConfig
	ctx       context.Context
	cancel    context.CancelFunc
	useSocket bool
}

// NewSlackAdapter creates a new Slack adapter
func NewSlackAdapter(channelID string) *SlackAdapter {
	common := NewCommonAdapter(channelID, channels.ChannelTypeSlack)
	return &SlackAdapter{
		CommonAdapter: common,
	}
}

// Initialize initializes the Slack adapter with configuration
func (a *SlackAdapter) Initialize(ctx context.Context, config channels.ChannelConfig) error {
	ctx, span := a.tracer.Start(ctx, "SlackAdapter.Initialize")
	defer span.End()

	if err := ValidateConfig(config); err != nil {
		span.RecordError(err)
		return err
	}

	a.mu.Lock()
	a.config = config
	a.mu.Unlock()

	// Create Slack client
	a.client = slack.New(config.Token, slack.OptionAppLevelToken(config.AppToken))

	// Set capabilities
	a.SetCapability(channels.CapabilityThreads, true)
	a.SetCapability(channels.CapabilityEdits, true)
	a.SetCapability(channels.CapabilityReactions, true)
	a.SetCapability(channels.CapabilityTypingIndicator, true)
	a.SetCapability(channels.CapabilityWebhooks, true)
	a.SetCapability(channels.CapabilityRichText, true)

	// Determine connection mode
	a.useSocket = config.AppToken != ""

	return nil
}

// Connect establishes connection to Slack
func (a *SlackAdapter) Connect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "SlackAdapter.Connect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.client == nil {
		err := fmt.Errorf("adapter not initialized")
		span.RecordError(err)
		return err
	}

	a.ctx, a.cancel = context.WithCancel(ctx)

	if a.useSocket {
		return a.connectSocket(ctx)
	}

	// For webhook mode, just verify connection
	authTest, err := a.client.AuthTest()
	if err != nil {
		span.RecordError(err)
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("slack auth test failed: %w", err)
	}

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, map[string]interface{}{
		"team_id":     authTest.TeamID,
		"bot_user_id": authTest.UserID,
	}, nil)

	return nil
}

// connectSocket establishes Socket Mode connection
func (a *SlackAdapter) connectSocket(ctx context.Context) error {
	var err error
	a.socket = slack.New(a.client, slack.OptionDebug(false))

	// Set up handlers
	a.setupSocketHandlers()

	// Start socket mode
	if err := a.socket.Run(); err != nil {
		a.SetStatus(context.Background(), channels.ChannelStatusError, err.Error())
		return fmt.Errorf("failed to start socket mode: %w", err)
	}

	a.SetStatus(context.Background(), channels.ChannelStatusConnected, "")
	a.EmitEvent(ctx, channels.EventTypeConnected, nil, nil)

	return nil
}

// Disconnect gracefully closes the Slack connection
func (a *SlackAdapter) Disconnect(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "SlackAdapter.Disconnect")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.socket != nil {
		// Socket mode client will be disconnected when context is cancelled
	}

	if a.cancel != nil {
		a.cancel()
	}

	a.SetStatus(context.Background(), channels.ChannelStatusDisconnected, "")
	a.EmitEvent(ctx, channels.EventTypeDisconnected, nil, nil)

	return nil
}

// SendMessage sends a message to Slack
func (a *SlackAdapter) SendMessage(ctx context.Context, msg *channels.Message, opts channels.MessageSendOptions) (string, error) {
	ctx, span := a.tracer.Start(ctx, "SlackAdapter.SendMessage",
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

	channelID := a.extractChannelID(msg)
	if channelID == "" {
		err := fmt.Errorf("no valid channel ID found")
		span.RecordError(err)
		return "", err
	}

	// Build message options
	channelOptions := []slack.MsgOption{
		slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAsUser(true),
	}

	// Add attachments if present
	if len(msg.Attachments) > 0 {
		slackAttachments := make([]slack.Attachment, len(msg.Attachments))
		for i, att := range msg.Attachments {
			slackAttachments[i] = slack.Attachment{
				Title:     att.Filename,
				TitleLink: att.URL,
			}
		}
		channelOptions = append(channelOptions, slack.MsgOptionAttachments(slackAttachments...))
	}

	// Handle thread
	if msg.ThreadID != "" {
		channelOptions = append(channelOptions, slack.MsgOptionTS(msg.ThreadID))
	}

	// Handle reply
	if opts.ReplyTo != "" {
		channelOptions = append(channelOptions, slack.MsgOptionTS(opts.ReplyTo))
	}

	// Handle parse mode
	if opts.ParseMode == "markdown" {
		// Slack uses mrkdwn by default
	} else if opts.DisableWebPagePreview {
		channelOptions = append(channelOptions, slack.MsgOptionUnfurl(false))
	}

	// Send message
	_, timestamp, err := a.client.PostMessage(channelID, channelOptions...)
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
		"timestamp":  timestamp,
		"channel_id": channelID,
	}, nil)

	return timestamp, nil
}

// EditMessage edits an existing message
func (a *SlackAdapter) EditMessage(ctx context.Context, messageID string, msg *channels.Message) error {
	ctx, span := a.tracer.Start(ctx, "SlackAdapter.EditMessage")
	defer span.End()

	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	channelID := a.extractChannelID(msg)
	if channelID == "" {
		return fmt.Errorf("no valid channel ID found")
	}

	options := []slack.MsgOption{
		slack.MsgOptionText(msg.Text, false),
	}

	_, _, _, err := a.client.UpdateMessage(channelID, messageID, options...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (a *SlackAdapter) DeleteMessage(ctx context.Context, messageID string) error {
	if !a.IsConnected() {
		return channels.ErrNotConnected
	}

	// Need channel ID - limitation of unified interface
	return fmt.Errorf("deletion requires channel ID (not available in unified interface)")
}

// UploadFile uploads a file and returns an attachment
func (a *SlackAdapter) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*channels.Attachment, error) {
	if !a.IsConnected() {
		return nil, channels.ErrNotConnected
	}

	// Get channel ID from context or config
	channelID := a.extractChannelIDFromConfig()
	if channelID == "" {
		return nil, fmt.Errorf("no channel ID configured for upload")
	}

	uploadParams := slack.FileUploadParameters{
		Filename: filename,
		Reader:   content,
		Filetype: mimeType,
	}

	file, err := a.client.UploadFile(uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &channels.Attachment{
		ID:       file.ID,
		Filename: file.Name,
		URL:      file.URLPrivate,
		Size:     int64(file.Size),
		MimeType: file.Mimetype,
	}, nil
}

// setupSocketHandlers sets up Socket Mode event handlers
func (a *SlackAdapter) setupSocketHandlers() {
	a.socket.OnEvent(func(event slack.Event, client *slack.SocketModeClient) {
		switch ev := event.Data.(type) {
		case *slack.MessageEvent:
			unifiedMsg := a.convertSocketMessageToUnified(ev)
			_ = a.HandleMessage(a.ctx, unifiedMsg)

		case *slack.EventsAPIEvent:
			if ev.Type == "message" {
				if msg, ok := ev.InnerEvent.Data.(*slack.MessageEvent); ok {
					unifiedMsg := a.convertSocketMessageToUnified(msg)
					_ = a.HandleMessage(a.ctx, unifiedMsg)
				}
			}
		}
	})
}

// convertSocketMessageToUnified converts Slack Socket Mode message to unified format
func (a *SlackAdapter) convertSocketMessageToUnified(msg *slack.MessageEvent) *channels.Message {
	user := &channels.User{
		ID:            msg.User,
		ChannelUserID: msg.User,
		DisplayName:   msg.Username,
		IsBot:         msg.BotID != "" || msg.SubType == "bot_message",
		ChannelType:   channels.ChannelTypeSlack,
	}

	unifiedMsg := &channels.Message{
		ID:          msg.Timestamp,
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionIncoming,
		Text:        msg.Text,
		ChannelID:   a.channelID,
		ChannelType: channels.ChannelTypeSlack,
		ChatID:      msg.Channel,
		From:        user,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]string),
	}

	// Handle thread
	if msg.ThreadTimestamp != "" {
		unifiedMsg.ThreadID = msg.ThreadTimestamp
	}

	// Handle reply
	if msg.ParentUserId != "" {
		unifiedMsg.ReplyToID = msg.ThreadTimestamp
	}

	// Handle edits
	if msg.Edited != nil {
		unifiedMsg.Edited = true
	}

	// Handle attachments
	if len(msg.Attachments) > 0 {
		unifiedMsg.Attachments = make([]channels.Attachment, len(msg.Attachments))
		for i, att := range msg.Attachments {
			unifiedMsg.Attachments[i] = channels.Attachment{
				ID:       att.ID,
				Filename: att.Title,
				URL:      att.URL,
				Title:    att.Title,
			}

			// Determine type
			if att.Mimetype != "" {
				unifiedMsg.Attachments[i].MimeType = att.Mimetype
			}
		}
	}

	// Handle files
	if len(msg.Files) > 0 {
		for _, f := range msg.Files {
			unifiedMsg.Attachments = append(unifiedMsg.Attachments, channels.Attachment{
				ID:       f.ID,
				Filename: f.Name,
				URL:      f.URLPrivate,
				Size:     f.Size,
				MimeType: f.Mimetype,
			})
		}
	}

	return unifiedMsg
}

// extractChannelID extracts channel ID from message metadata
func (a *SlackAdapter) extractChannelID(msg *channels.Message) string {
	// Try metadata first
	if channelID, ok := msg.Metadata["channel_id"]; ok {
		return channelID
	}

	// Try ChatID field
	return msg.ChatID
}

// extractChannelIDFromConfig extracts default channel ID from config
func (a *SlackAdapter) extractChannelIDFromConfig() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.config.PlatformConfig != nil {
		if channelID, ok := a.config.PlatformConfig["channel_id"].(string); ok {
			return channelID
		}
	}

	return ""
}

// GetClient returns the underlying Slack client
func (a *SlackAdapter) GetClient() *slack.Client {
	return a.client
}
