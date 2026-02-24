// Agent Framework - Multi-Channel Integration for Core Application
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/channels"
)

// ChannelManager extends the Application with multi-channel support
// This can be embedded or used as a separate manager
type ChannelManager struct {
	// Channel management
	manager       *channels.Manager
	config        *channels.Config
	sessions      map[string]*ChannelSession
	sessionsMutex sync.RWMutex

	// Integration with agent system
	agentHost       *agent.Host
	skillLibrary    agent.SkillLibrary
	workflowManager *agent.WorkflowManager

	// Context
	ctx     context.Context
	cancel  context.CancelFunc
	enabled bool
}

// ChannelSession represents a user session across channels
type ChannelSession struct {
	ID           string
	ChannelID    string
	ChannelType  channels.ChannelType
	UserID       string
	UserName     string
	MessageCount int
	LastActive   int64
	CreatedAt    int64
	UpdatedAt    int64
	// Agent session data could be stored here
	AgentSession interface{}
}

// NewChannelManager creates a new channel manager integrated with agent system
func NewChannelManager(ctx context.Context, app *Application) (*ChannelManager, error) {
	if app == nil {
		return nil, fmt.Errorf("application is required")
	}

	// Create channel manager
	manager, err := channels.NewManager(&channels.ManagerConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel manager: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	cm := &ChannelManager{
		manager:         manager,
		sessions:        make(map[string]*ChannelSession),
		sessionsMutex:   sync.RWMutex{},
		agentHost:       app.host,
		skillLibrary:    app.skillLibrary,
		workflowManager: app.workflowManager,
		ctx:             ctx,
		cancel:          cancel,
		enabled:         true,
	}

	// Set message handler
	manager.SetMessageHandler(cm.handleChannelMessage)

	return cm, nil
}

// Enabled returns whether channels are enabled
func (cm *ChannelManager) Enabled() bool {
	return cm.enabled
}

// LoadConfig loads channel configuration from file
func (cm *ChannelManager) LoadConfig(configPath string) error {
	config, err := channels.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cm.config = config
	return nil
}

// LoadConfigFromEnv loads configuration from environment variables
func (cm *ChannelManager) LoadConfigFromEnv() error {
	config, err := channels.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config from env: %w", err)
	}

	cm.config = config
	return nil
}

// Start starts the channel manager and all registered channels
func (cm *ChannelManager) Start() error {
	if !cm.enabled {
		return nil
	}

	log.Println("🤖 Starting multi-channel system...")

	// Register enabled channels
	if cm.config != nil {
		enabledChannels := cm.config.GetEnabledChannels()
		for name, chConfig := range enabledChannels {
			if err := cm.registerChannel(name, chConfig); err != nil {
				log.Printf("⚠️  Failed to register channel %s: %v", name, err)
				continue
			}
			log.Printf("✅ Channel %s (%s) registered", name, chConfig.Type)
		}

		// Setup routing rules
		for _, rule := range cm.config.GetRoutingRules() {
			if err := cm.manager.AddRoutingRule(rule); err != nil {
				log.Printf("⚠️  Failed to add routing rule %s: %v", rule.ID, err)
			}
		}
	}

	// Start manager
	if err := cm.manager.Start(cm.ctx); err != nil {
		return fmt.Errorf("failed to start channel manager: %w", err)
	}

	log.Println("✅ Multi-channel system started")
	return nil
}

// Stop stops the channel manager and all channels
func (cm *ChannelManager) Stop() error {
	if !cm.enabled {
		return nil
	}

	log.Println("🛑 Stopping multi-channel system...")

	if err := cm.manager.Stop(cm.ctx); err != nil {
		return fmt.Errorf("failed to stop channel manager: %w", err)
	}

	cm.cancel()
	log.Println("✅ Multi-channel system stopped")
	return nil
}

// SendMessage sends a message to a specific channel
func (cm *ChannelManager) SendMessage(channelID, text string, opts channels.MessageSendOptions) (string, error) {
	if !cm.enabled {
		return "", fmt.Errorf("channels not enabled")
	}

	msg := &channels.Message{
		Type: channels.MessageTypeText,
		Text: text,
	}

	return cm.manager.SendMessage(cm.ctx, channelID, msg, opts)
}

// Broadcast sends a message to all channels of a specific type
func (cm *ChannelManager) Broadcast(channelType channels.ChannelType, text string) (map[string]string, error) {
	if !cm.enabled {
		return nil, fmt.Errorf("channels not enabled")
	}

	msg := &channels.Message{
		Type: channels.MessageTypeText,
		Text: text,
	}

	return cm.manager.Broadcast(cm.ctx, channelType, msg, channels.MessageSendOptions{})
}

// GetStats returns statistics for all channels
func (cm *ChannelManager) GetStats() (map[string]*channels.ChannelStats, error) {
	if !cm.enabled {
		return nil, nil
	}

	return cm.manager.GetStats(cm.ctx)
}

// handleChannelMessage handles incoming messages from channels
func (cm *ChannelManager) handleChannelMessage(msg *channels.Message) error {
	// Log message
	log.Printf("📨 [%s] %s: %s", msg.ChannelType, msg.From.DisplayName, msg.Text)

	// Get or create session
	session := cm.getOrCreateSession(msg)

	// Process through agent system
	response, err := cm.processWithAgent(msg, session)
	if err != nil {
		log.Printf("⚠️  Error processing message: %v", err)
		return err
	}

	// Send response back
	if response != "" {
		opts := channels.MessageSendOptions{
			ReplyTo: msg.ID,
		}

		_, err = cm.manager.SendMessage(cm.ctx, msg.ChannelID, &channels.Message{
			Type: channels.MessageTypeText,
			Text: response,
		}, opts)

		if err != nil {
			log.Printf("⚠️  Failed to send response: %v", err)
			return err
		}

		log.Printf("📤 Response sent to %s", msg.ChannelType)
	}

	return nil
}

// handleChannelEvent handles events from channels
func (cm *ChannelManager) handleChannelEvent(event channels.Event) {
	switch event.Type {
	case channels.EventTypeConnected:
		log.Printf("🔗 Channel connected: %s (%s)", event.ChannelID, event.ChannelType)
	case channels.EventTypeDisconnected:
		log.Printf("🔌 Channel disconnected: %s (%s)", event.ChannelID, event.ChannelType)
	case channels.EventTypeError:
		log.Printf("❌ Channel error: %s (%s): %v", event.ChannelID, event.ChannelType, event.Error)
	case channels.EventTypeMessageSent:
		// Could track metrics here
	}
}

// registerChannel registers a single channel
func (cm *ChannelManager) registerChannel(name string, config channels.ChannelConfig) error {
	// Create adapter using factory
	factory := cm.manager.GetFactory()
	adapter, err := factory.CreateAdapter(name, config.Type)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// Initialize adapter
	if err := adapter.Initialize(cm.ctx, config); err != nil {
		return fmt.Errorf("failed to initialize adapter: %w", err)
	}

	// Register with manager
	if err := cm.manager.RegisterAdapter(adapter); err != nil {
		return fmt.Errorf("failed to register adapter: %w", err)
	}

	return nil
}

// getOrCreateSession gets or creates a user session
func (cm *ChannelManager) getOrCreateSession(msg *channels.Message) *ChannelSession {
	key := fmt.Sprintf("%s:%s", msg.ChannelID, msg.From.ID)

	cm.sessionsMutex.Lock()
	defer cm.sessionsMutex.Unlock()

	session, exists := cm.sessions[key]
	if !exists {
		session = &ChannelSession{
			ID:           generateSessionID(),
			ChannelID:    msg.ChannelID,
			ChannelType:  msg.ChannelType,
			UserID:       msg.From.ID,
			UserName:     msg.From.DisplayName,
			MessageCount: 0,
			CreatedAt:    msg.Timestamp.Unix(),
		}
		cm.sessions[key] = session
	}

	session.MessageCount++
	session.UpdatedAt = msg.Timestamp.Unix()

	return session
}

// processWithAgent processes message through agent system
func (cm *ChannelManager) processWithAgent(msg *channels.Message, session *ChannelSession) (string, error) {
	// Check if it's a command
	if len(msg.Text) > 0 && msg.Text[0] == '/' {
		return cm.handleCommand(msg, session)
	}

	// For now, return a simple echo
	// In production, this would:
	// 1. Create/get agent conversation
	// 2. Add message to history
	// 3. Run agent
	// 4. Get response

	return fmt.Sprintf("Echo: %s\n\n(Session: %s, Messages: %d)",
		msg.Text, session.ID, session.MessageCount), nil
}

// handleCommand handles bot commands
func (cm *ChannelManager) handleCommand(msg *channels.Message, session *ChannelSession) (string, error) {
	switch msg.Text {
	case "/help":
		return `Available commands:
	/help - Show this help
	/stats - Show channel statistics
	/skills - List available skills
	/version - Show version
	/echo <text> - Echo back text`, nil
	case "/stats":
		stats, _ := cm.GetStats()
		response := "📊 Channel Statistics:\n"
		for channelID, stat := range stats {
			response += fmt.Sprintf("  %s: %d sent, %d received\n",
				channelID, stat.MessagesSent, stat.MessagesReceived)
		}
		return response, nil
	case "/skills":
		if cm.skillLibrary == nil {
			return "Skill system not available", nil
		}

		skills := cm.skillLibrary.GetAllSkills(cm.ctx)
		response := fmt.Sprintf("📚 Available skills: %d\n", len(skills))
		for name := range skills {
			response += fmt.Sprintf("  - %s\n", name)
		}
		return response, nil
	case "/version":
		return "AgentFramework Multi-Channel Bot v1.0.0", nil
	default:
		// Handle /echo command
		if len(msg.Text) > 6 && msg.Text[:5] == "/echo" {
			return msg.Text[6:], nil
		}

		return fmt.Sprintf("Unknown command: %s\nType /help for available commands", msg.Text), nil
	}
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

// AddRoutingRule adds a custom routing rule
func (cm *ChannelManager) AddRoutingRule(rule *channels.RoutingRule) error {
	return cm.manager.AddRoutingRule(rule)
}

// RemoveRoutingRule removes a routing rule
func (cm *ChannelManager) RemoveRoutingRule(ruleID string) {
	cm.manager.RemoveRoutingRule(ruleID)
}

// GetRoutingRules returns all routing rules
func (cm *ChannelManager) GetRoutingRules() []*channels.RoutingRule {
	return cm.manager.GetRouter().ListRules()
}

// GetSessions returns all active sessions
func (cm *ChannelManager) GetSessions() []*ChannelSession {
	cm.sessionsMutex.RLock()
	defer cm.sessionsMutex.RUnlock()

	sessions := make([]*ChannelSession, 0, len(cm.sessions))
	for _, sess := range cm.sessions {
		sessions = append(sessions, sess)
	}

	return sessions
}

// CleanupSessions removes inactive sessions
func (cm *ChannelManager) CleanupSessions(timeout int64) int {
	cm.sessionsMutex.Lock()
	defer cm.sessionsMutex.Unlock()

	now := time.Now().Unix()
	removed := 0

	for key, sess := range cm.sessions {
		if now-sess.UpdatedAt > timeout {
			delete(cm.sessions, key)
			removed++
		}
	}

	return removed
}

// GetChannelManager returns the underlying channel manager
func (cm *ChannelManager) GetChannelManager() *channels.Manager {
	return cm.manager
}

// SendMessageToUser sends a message to a specific user across channels
func (cm *ChannelManager) SendMessageToUser(userID string, text string) error {
	if !cm.enabled {
		return fmt.Errorf("channels not enabled")
	}

	// Find all sessions for this user
	sentCount := 0
	cm.sessionsMutex.RLock()
	for _, sess := range cm.sessions {
		if sess.UserID == userID {
			msg := &channels.Message{
				Type: channels.MessageTypeText,
				Text: text,
			}
			_, err := cm.manager.SendMessage(cm.ctx, sess.ChannelID, msg, channels.MessageSendOptions{})
			if err == nil {
				sentCount++
			}
		}
	}
	cm.sessionsMutex.RUnlock()

	if sentCount == 0 {
		return fmt.Errorf("no active sessions found for user: %s", userID)
	}

	return nil
}

// BroadcastToAll sends a message to all active channels
func (cm *ChannelManager) BroadcastToAll(text string) error {
	if !cm.enabled {
		return fmt.Errorf("channels not enabled")
	}

	// Get all active channel IDs
	cm.sessionsMutex.RLock()
	channelIDs := make(map[string]bool)
	for _, sess := range cm.sessions {
		channelIDs[sess.ChannelID] = true
	}
	cm.sessionsMutex.RUnlock()

	// Send to each channel
	for channelID := range channelIDs {
		msg := &channels.Message{
			Type: channels.MessageTypeText,
			Text: text,
		}

		_, err := cm.manager.SendMessage(cm.ctx, channelID, msg, channels.MessageSendOptions{})
		if err != nil {
			log.Printf("⚠️  Failed to broadcast to %s: %v", channelID, err)
		}
	}

	return nil
}

// GetActiveChannelIDs returns all active channel IDs
func (cm *ChannelManager) GetActiveChannelIDs() []string {
	if !cm.enabled {
		return nil
	}

	cm.sessionsMutex.RLock()
	defer cm.sessionsMutex.RUnlock()

	seen := make(map[string]bool)
	for _, sess := range cm.sessions {
		seen[sess.ChannelID] = true
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	return ids
}

// ProcessMessageWithAgent processes a message through the agent system
// This is a placeholder for full agent integration
func (cm *ChannelManager) ProcessMessageWithAgent(ctx context.Context, msg *channels.Message, session *ChannelSession) (string, error) {
	// In production, this would:
	// 1. Get or create agent conversation for session
	// 2. Convert channel message to agent message format
	// 3. Add to conversation history
	// 4. Run agent with conversation
	// 5. Get agent response
	// 6. Convert back to channel message
	// 7. Update conversation history

	// For now, return simple response
	return fmt.Sprintf("Processed: %s", msg.Text), nil
}

// SendRichMessage sends a rich message with attachments
func (cm *ChannelManager) SendRichMessage(channelID, text string, attachments []channels.Attachment) (string, error) {
	if !cm.enabled {
		return "", fmt.Errorf("channels not enabled")
	}

	msg := &channels.Message{
		Type:        channels.MessageTypeText,
		Text:        text,
		Attachments: attachments,
	}

	return cm.manager.SendMessage(cm.ctx, channelID, msg, channels.MessageSendOptions{})
}
