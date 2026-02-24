// Package main provides a simple multi-channel bot example
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AgentFramework/pkg/channels"
	// "AgentFramework/pkg/channels/adapters"
)

// SimpleBot demonstrates a minimal multi-channel bot
type SimpleBot struct {
	manager *channels.Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewSimpleBot creates a new simple bot
func NewSimpleBot() (*SimpleBot, error) {
	// Create channel manager
	manager, err := channels.NewManager(&channels.ManagerConfig{})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	bot := &SimpleBot{
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Set message handler
	manager.SetMessageHandler(bot.handleMessage)

	return bot, nil
}

// Start starts the bot
func (b *SimpleBot) Start(configPath string) error {
	log.Println("🤖 Starting SimpleBot...")

	// Load configuration
	config, err := channels.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup routing rules
	b.setupRoutes(config)

	// Register and start enabled channels
	enabledChannels := config.GetEnabledChannels()
	if len(enabledChannels) == 0 {
		log.Println("⚠️  No channels enabled!")
		log.Println("💡 Set environment variables or edit config file to enable channels")
		log.Println("")
		log.Println("Example environment variables:")
		log.Println("  export TELEGRAM_BOT_TOKEN=your_token")
		log.Println("  export DISCORD_BOT_TOKEN=your_token")
		log.Println("  export QQ_BOT_ENABLED=true")
	}

	for name, chConfig := range enabledChannels {
		if err := b.registerChannel(name, chConfig); err != nil {
			log.Printf("⚠️  Failed to register channel %s: %v", name, err)
			continue
		}
		log.Printf("✅ Channel %s (%s) registered", name, chConfig.Type)
	}

	// Start manager
	if err := b.manager.Start(b.ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	log.Println("✅ SimpleBot started successfully!")
	log.Printf("📊 Listening on %d channel(s)", len(enabledChannels))

	return nil
}

// Stop stops the bot
func (b *SimpleBot) Stop() {
	log.Println("🛑 Stopping SimpleBot...")

	if err := b.manager.Stop(b.ctx); err != nil {
		log.Printf("⚠️  Error stopping manager: %v", err)
	}

	b.cancel()
	log.Println("✅ SimpleBot stopped")
}

// handleMessage handles incoming messages
func (b *SimpleBot) handleMessage(msg *channels.Message) error {
	// Log message
	log.Printf("📨 [%s] %s: %s",
		msg.ChannelType,
		msg.From.DisplayName,
		msg.Text,
	)

	// Simple echo/response logic
	response := b.generateResponse(msg)

	if response != "" {
		opts := channels.MessageSendOptions{
			ReplyTo: msg.ID,
		}

		_, err := b.manager.SendMessage(b.ctx, msg.ChannelID, &channels.Message{
			Type:    channels.MessageTypeText,
			Text:    response,
		}, opts)

		if err != nil {
			log.Printf("⚠️  Failed to send response: %v", err)
			return err
		}

		log.Printf("📤 Sent response to %s", msg.ChannelType)
	}

	return nil
}

// handleEvent handles channel events
func (b *SimpleBot) handleEvent(event channels.Event) {
	switch event.Type {
	case channels.EventTypeConnected:
		log.Printf("🔗 Channel connected: %s (%s)", event.ChannelID, event.ChannelType)
	case channels.EventTypeDisconnected:
		log.Printf("🔌 Channel disconnected: %s (%s)", event.ChannelID, event.ChannelType)
	case channels.EventTypeError:
		log.Printf("❌ Channel error: %s (%s): %v", event.ChannelID, event.ChannelType, event.Error)
	case channels.EventTypeMessageSent:
		log.Printf("✉️  Message sent to %s", event.ChannelID)
	}
}

// generateResponse generates a response to the message
func (b *SimpleBot) generateResponse(msg *channels.Message) string {
	text := msg.Text

	// Handle commands
	if len(text) > 0 && text[0] == '/' {
		return b.handleCommand(msg)
	}

	// Handle common greetings
	switch text {
	case "hello", "hi", "你好", "嗨":
		return fmt.Sprintf("Hello %s! 👋", msg.From.DisplayName)
	case "help", "帮助":
		return `Available commands:
	/help - Show this help
	/time - Show current time
	/ping - Check if bot is alive
	/stats - Show channel statistics`
	case "ping":
		return "pong! 🏓"
	default:
		// Echo back
		return fmt.Sprintf("You said: %s", text)
	}
}

// handleCommand handles bot commands
func (b *SimpleBot) handleCommand(msg *channels.Message) string {
	switch msg.Text {
	case "/time":
		return fmt.Sprintf("Current time: %s 🕐", time.Now().Format("2006-01-02 15:04:05"))
	case "/ping":
		return "pong! 🏓"
	case "/stats":
		stats, err := b.manager.GetStats(b.ctx)
		if err != nil {
			return "Failed to get statistics"
		}

		response := "📊 Statistics:\n"
		for channelID, stat := range stats {
			response += fmt.Sprintf("  %s: %d sent, %d received\n",
				channelID, stat.MessagesSent, stat.MessagesReceived)
		}
		return response
	default:
		return fmt.Sprintf("Unknown command: %s\nType /help for available commands", msg.Text)
	}
}

// setupRoutes sets up message routing rules
func (b *SimpleBot) setupRoutes(config *channels.Config) {
	// Add default routing rules
	b.manager.AddRoutingRule(&channels.RoutingRule{
		ID:       "accept-all",
		Name:     "Accept all messages",
		Priority: 100,
		Action:   channels.RoutingActionAccept,
	})

	// Add custom rules from config
	for _, rule := range config.GetRoutingRules() {
		if err := b.manager.AddRoutingRule(rule); err != nil {
			log.Printf("⚠️  Failed to add routing rule %s: %v", rule.ID, err)
		}
	}
}

// registerChannel registers a single channel
func (b *SimpleBot) registerChannel(name string, config channels.ChannelConfig) error {
	// TODO: Implement proper adapter creation
	// The CommonAdapter doesn't fully implement ChannelAdapter interface
	// This needs a proper adapter implementation for each channel type
	log.Printf("⚠️  Channel %s (%s) - adapter not fully implemented yet", name, config.Type)
	return nil
}

// showStats shows current statistics
func (b *SimpleBot) showStats() {
	stats, err := b.manager.GetStats(b.ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return
	}

	log.Println("\n📊 Channel Statistics:")
	for channelID, stat := range stats {
		log.Printf("  %s:", channelID)
		log.Printf("    Status: %s", stat.Status)
		log.Printf("    Sent: %d", stat.MessagesSent)
		log.Printf("    Received: %d", stat.MessagesReceived)
		log.Printf("    Errors: %d", stat.ErrorCount)
		log.Printf("    Uptime: %s", stat.Uptime)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 AgentFramework Multi-Channel Bot")
	log.Println("====================================")

	// Create bot
	bot, err := NewSimpleBot()
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Determine config path
	configPath := "config/channels.example.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	} else if envPath := os.Getenv("CHANNELS_CONFIG"); envPath != "" {
		configPath = envPath
	}

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("⚠️  Config file not found: %s", configPath)
		log.Println("💡 Using environment variables for configuration...")

		// Try to load from environment
		config, err := channels.LoadConfigFromEnv()
		if err != nil {
			log.Fatalf("Failed to load config from environment: %v", err)
		}

		// Save generated config for reference
		generatedPath := "config/channels.generated.yaml"
		if err := channels.SaveConfig(config, generatedPath, channels.ConfigFormatYAML); err != nil {
			log.Printf("Warning: failed to save generated config: %v", err)
		} else {
			log.Printf("💾 Generated config saved to: %s", generatedPath)
			configPath = generatedPath
		}
	}

	// Start bot
	if err := bot.Start(configPath); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
	defer bot.Stop()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Stats ticker
	statsTicker := time.NewTicker(30 * time.Second)
	defer statsTicker.Stop()

	log.Println("\n✅ Bot is running!")
	log.Println("Send a message to any connected channel to interact")
	log.Println("Press Ctrl+C to shutdown\n")

	// Main loop
	for {
		select {
		case <-sigChan:
			log.Println("\n🛑 Shutdown signal received")
			bot.showStats()
			return

		case <-statsTicker.C:
			bot.showStats()

		case <-bot.ctx.Done():
			log.Println("\n🛑 Context cancelled")
			return
		}
	}
}
