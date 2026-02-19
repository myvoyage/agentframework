// Agent Framework - Application Extension with Multi-Channel Support
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// This file demonstrates how to extend the core.Application with multi-channel support
// To use this, you would typically modify core/application.go to include the channel manager

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
	"AgentFramework/pkg/channels"
)

// ApplicationWithChannels extends Application with multi-channel support
//
// Usage:
//
//	app := &core.ApplicationWithChannels{
//	    Application: coreApp,
//	}
//	app.InitChannels(ctx)
//	app.StartChannels()
type ApplicationWithChannels struct {
	*Application

	// Channel manager
	channels *ChannelManager
}

// NewApplicationWithChannels creates a new application with multi-channel support
func NewApplicationWithChannels(ctx context.Context, config *agent.HostConfig, modelFactory agent.ModelFactory, toolRegistry map[string]agent.Tool) (*ApplicationWithChannels, error) {
	// Create base application
	baseApp, err := NewApplication(ctx, config, modelFactory, toolRegistry)
	if err != nil {
		return nil, err
	}

	app := &ApplicationWithChannels{
		Application: baseApp,
	}

	return app, nil
}

// InitChannels initializes the multi-channel system
func (a *ApplicationWithChannels) InitChannels(ctx context.Context) error {
	cm, err := NewChannelManager(ctx, a.Application)
	if err != nil {
		return fmt.Errorf("failed to create channel manager: %w", err)
	}

	a.channels = cm
	return nil
}

// LoadChannelConfig loads configuration from file
func (a *ApplicationWithChannels) LoadChannelConfig(configPath string) error {
	return a.channels.LoadConfig(configPath)
}

// LoadChannelConfigFromEnv loads configuration from environment
func (a *ApplicationWithChannels) LoadChannelConfigFromEnv() error {
	return a.channels.LoadConfigFromEnv()
}

// StartChannels starts all enabled channels
func (a *ApplicationWithChannels) StartChannels() error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.Start()
}

// StopChannels stops all channels
func (a *ApplicationWithChannels) StopChannels() error {
	if a.channels == nil {
		return nil
	}

	return a.channels.Stop()
}

// SendChannelMessage sends a message to a specific channel
func (a *ApplicationWithChannels) SendChannelMessage(channelID, text string) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.SendMessage(channelID, text, channels.MessageSendOptions{})
}

// BroadcastChannelMessage broadcasts a message to all channels of a type
func (a *ApplicationWithChannels) BroadcastChannelMessage(channelType channels.ChannelType, text string) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	_, err := a.channels.Broadcast(channelType, text)
	return err
}

// GetChannelStats returns statistics for all channels
func (a *ApplicationWithChannels) GetChannelStats() (map[string]*channels.ChannelStats, error) {
	if a.channels == nil {
		return nil, nil
	}

	return a.channels.GetStats()
}

// GetChannelSessions returns all active channel sessions
func (a *ApplicationWithChannels) GetChannelSessions() []*ChannelSession {
	if a.channels == nil {
		return nil
	}

	return a.channels.GetSessions()
}

// SendChannelMessageToUser sends a message to a specific user
func (a *ApplicationWithChannels) SendChannelMessageToUser(userID, text string) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.SendMessageToUser(userID, text)
}

// BroadcastToAllChannels sends a message to all active channels
func (a *ApplicationWithChannels) BroadcastToAllChannels(text string) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.BroadcastToAll(text)
}

// GetActiveChannelIDs returns all active channel IDs
func (a *ApplicationWithChannels) GetActiveChannelIDs() []string {
	if a.channels == nil {
		return nil
	}

	return a.channels.GetActiveChannelIDs()
}

// AddChannelRoutingRule adds a routing rule
func (a *ApplicationWithChannels) AddChannelRoutingRule(rule *channels.RoutingRule) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.AddRoutingRule(rule)
}

// RemoveChannelRoutingRule removes a routing rule
func (a *ApplicationWithChannels) RemoveChannelRoutingRule(ruleID string) {
	if a.channels == nil {
		return
	}

	a.channels.RemoveRoutingRule(ruleID)
}

// GetChannelManager returns the underlying channel manager
func (a *ApplicationWithChannels) GetChannelManager() *ChannelManager {
	return a.channels
}

// ProcessChannelMessage processes a message through channels and agents
// This is useful for testing or external message injection
func (a *ApplicationWithChannels) ProcessChannelMessage(msg *channels.Message) error {
	if a.channels == nil {
		return fmt.Errorf("channels not initialized")
	}

	return a.channels.handleChannelMessage(msg)
}

// Example usage:
//
//   // Create application with channels
//   app, err := core.NewApplicationWithChannels(ctx, config, modelFactory, tools)
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   // Initialize channels
//   if err := app.InitChannels(ctx); err != nil {
//       log.Printf("Warning: channels not available: %v", err)
//   }
//
//   // Load channel configuration
//   if err := app.LoadChannelConfigFromEnv(); err != nil {
//       log.Printf("Warning: failed to load channel config: %v", err)
//   }
//
//   // Start channels
//   if err := app.StartChannels(); err != nil {
//       log.Printf("Warning: failed to start channels: %v", err)
//   }
//   defer app.StopChannels()
//
//   // Use application...
//   app.Run()
//
//   // Send messages
//   app.SendChannelMessage("telegram", "Hello from AgentFramework!")
//   app.BroadcastChannelMessage(channels.ChannelTypeTelegram, "Broadcast message")
