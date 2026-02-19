// Package channels provides integration example with Agent system
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package examples demonstrates how to integrate the multi-channel system
// with the AgentFramework
package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/channels"
	"AgentFramework/pkg/channels/adapters"
)

// ChannelBridge bridges multi-channel messages to the Agent system
//
// SOLID - Single Responsibility Principle:
// Only responsible for bridging channels to agents
//
// SOLID - Dependency Inversion:
// Depends on abstractions (interfaces) not concrete implementations
type ChannelBridge struct {
	manager   *channels.Manager
	agentHost *agent.Host
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewChannelBridge creates a new channel bridge
func NewChannelBridge(agentHost *agent.Host) (*ChannelBridge, error) {
	// Create channel manager
	manager, err := channels.NewManager(&channels.ManagerConfig{
		EnableMetrics: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel manager: %w", err)
	}

	// Set message handler to forward to agent
	manager.SetMessageHandler(func(msg *channels.Message) error {
		return handleAgentMessage(agentHost, msg)
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &ChannelBridge{
		manager:   manager,
		agentHost: agentHost,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start starts the channel bridge
func (b *ChannelBridge) Start(configPath string) error {
	// Load configuration
	config, err := channels.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Register and start enabled channels
	for name, chConfig := range config.GetEnabledChannels() {
		if err := b.registerChannel(name, chConfig); err != nil {
			log.Printf("Warning: failed to register channel %s: %v", name, err)
			continue
		}
	}

	// Setup routing rules
	for _, rule := range config.GetRoutingRules() {
		if err := b.manager.AddRoutingRule(rule); err != nil {
			log.Printf("Warning: failed to add routing rule %s: %v", rule.ID, err)
		}
	}

	// Start manager
	if err := b.manager.Start(b.ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	log.Println("Channel bridge started successfully")
	return nil
}

// Stop stops the channel bridge
func (b *ChannelBridge) Stop() error {
	log.Println("Stopping channel bridge...")

	if err := b.manager.Stop(b.ctx); err != nil {
		return fmt.Errorf("failed to stop manager: %w", err)
	}

	b.cancel()
	log.Println("Channel bridge stopped")
	return nil
}

// registerChannel registers and initializes a channel
func (b *ChannelBridge) registerChannel(name string, config channels.ChannelConfig) error {
	// Create adapter using factory
	factory := b.manager.GetFactory()
	adapter, err := factory.CreateAdapter(name, config.Type)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// Initialize adapter
	if err := adapter.Initialize(b.ctx, config); err != nil {
		return fmt.Errorf("failed to initialize adapter: %w", err)
	}

	// Register with manager
	if err := b.manager.RegisterAdapter(adapter); err != nil {
		return fmt.Errorf("failed to register adapter: %w", err)
	}

	log.Printf("Channel %s (%s) registered successfully", name, config.Type)
	return nil
}

// SendMessage sends a message through a specific channel
func (b *ChannelBridge) SendMessage(channelID string, text string, opts channels.MessageSendOptions) (string, error) {
	msg := &channels.Message{
		ID:          channels.GenerateMessageID(),
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionOutgoing,
		Text:        text,
		Timestamp:   time.Now(),
	}

	return b.manager.SendMessage(b.ctx, channelID, msg, opts)
}

// Broadcast sends a message to all channels of a specific type
func (b *ChannelBridge) Broadcast(channelType channels.ChannelType, text string) (map[string]string, error) {
	msg := &channels.Message{
		ID:          channels.GenerateMessageID(),
		Type:        channels.MessageTypeText,
		Direction:   channels.MessageDirectionOutgoing,
		Text:        text,
		Timestamp:   time.Now(),
	}

	opts := channels.MessageSendOptions{}
	return b.manager.Broadcast(b.ctx, channelType, msg, opts)
}

// GetStats returns statistics for all channels
func (b *ChannelBridge) GetStats() (map[string]*channels.ChannelStats, error) {
	return b.manager.GetStats(b.ctx)
}

// handleAgentMessage handles incoming messages from channels and forwards to agent
func handleAgentMessage(agentHost *agent.Host, msg *channels.Message) error {
	// Convert channel message to agent message format
	// This is a simplified example - actual implementation would
	// need to handle different message types, attachments, etc.

	log.Printf("Received message from %s: %s", msg.ChannelType, msg.Text)

	// Create agent context
	ctx := context.Background()

	// Process through agent system
	// This would typically involve:
	// 1. Creating a conversation thread
	// 2. Running the agent with the message
	// 3. Getting the response
	// 4. Sending the response back through the channel

	// Example: Send to agent and get response
	response := fmt.Sprintf("Echo: %s", msg.Text)

	// In a real implementation, you would:
	// - Create a session for the user
	// - Run the agent with the message
	// - Get the agent's response
	// - Send the response back through the channel

	_ = ctx
	_ = agentHost
	_ = response

	return nil
}

// Example usage demonstrating how to integrate channels with AgentFramework
func Example_ChannelIntegration() {
	// This example shows how to integrate the multi-channel system
	// with the AgentFramework

	// 1. Create or get the agent host
	// host := agent.NewHost(...)

	// 2. Create channel bridge
	// bridge, err := NewChannelBridge(host)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// 3. Start with configuration
	// if err := bridge.Start("config/channels.yaml"); err != nil {
	//     log.Fatal(err)
	// }
	// defer bridge.Stop()

	// 4. Send messages
	// messageID, err := bridge.SendMessage("telegram", "Hello from agent!", channels.MessageSendOptions{})
	// if err != nil {
	//     log.Printf("Failed to send: %v", err)
	// }

	// 5. Broadcast to all channels of a type
	// results, err := bridge.Broadcast(channels.ChannelTypeTelegram, "Broadcast message")
	// if err != nil {
	//     log.Printf("Failed to broadcast: %v", err)
	// }

	// 6. Get statistics
	// stats, err := bridge.GetStats()
	// if err != nil {
	//     log.Printf("Failed to get stats: %v", err)
	// }
	// for channelID, stat := range stats {
	//     log.Printf("Channel %s: %d sent, %d received", channelID, stat.MessagesSent, stat.MessagesReceived)
	// }
}

// Example_AdvancedChannelIntegration demonstrates advanced usage
func Example_AdvancedChannelIntegration() {
	// This example shows advanced features like routing,
	// custom handlers, and event monitoring

	// 1. Create bridge
	// bridge, _ := NewChannelBridge(host)

	// 2. Add custom routing rules
	// bridge.manager.AddRoutingRule(&channels.RoutingRule{
	//     ID:       "admin-commands",
	//     Priority: 200,
	//     Pattern:  "^/admin",
	//     Action:   channels.RoutingActionAccept,
	//     ActionData: map[string]string{
	//         "handler": "admin",
	//     },
	//     RateLimit: 5,
	//     RateWindow: time.Minute,
	// })

	// 3. Set up event handler
	// bridge.manager.(*channels.Manager).SetEventHandler(func(event channels.Event) {
	//     switch event.Type {
	//     case channels.EventTypeConnected:
	//         log.Printf("Channel %s connected", event.ChannelID)
	//     case channels.EventTypeDisconnected:
	//         log.Printf("Channel %s disconnected", event.ChannelID)
	//     case channels.EventTypeError:
	//         log.Printf("Channel %s error: %v", event.ChannelID, event.Error)
	//     }
	// })

	// 4. Start bridge
	// bridge.Start("config/channels.yaml")
	// defer bridge.Stop()

	// Wait for shutdown signal
	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// <-sigChan
}

// Main function demonstrates a complete integration example
func Main() error {
	log.Println("Starting AgentFramework with multi-channel support...")

	// In a real application, you would get the agent host from your application
	// For this example, we'll create a minimal setup

	// 1. Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 2. Load configuration from environment or file
	configPath := "config/channels.example.yaml"
	if envPath := os.Getenv("CHANNELS_CONFIG"); envPath != "" {
		configPath = envPath
	}

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found: %s", configPath)
		log.Println("Using environment variables for configuration...")

		// Load from environment
		config, err := channels.LoadConfigFromEnv()
		if err != nil {
			return fmt.Errorf("failed to load config from env: %w", err)
		}

		// Save config for reference
		if err := channels.SaveConfig(config, "config/channels.generated.yaml", channels.ConfigFormatYAML); err != nil {
			log.Printf("Warning: failed to save generated config: %v", err)
		}

		configPath = "config/channels.generated.yaml"
	}

	// 3. Create and start channel manager
	// Note: In real usage, pass your actual agent.Host here
	// bridge, err := NewChannelBridge(nil)
	// if err != nil {
	//     return fmt.Errorf("failed to create bridge: %w", err)
	// }

	// if err := bridge.Start(configPath); err != nil {
	//     return fmt.Errorf("failed to start bridge: %w", err)
	// }
	// defer bridge.Stop()

	log.Printf("Multi-channel system started with config: %s", configPath)

	// 4. Display channel status
	// stats, err := bridge.GetStats()
	// if err != nil {
	//     log.Printf("Warning: failed to get stats: %v", err)
	// } else {
	//     for channelID, stat := range stats {
	//         log.Printf("Channel %s: %s", channelID, stat.Status)
	//     }
	// }

	// 5. Wait for shutdown signal
	log.Println("Press Ctrl+C to shutdown...")
	<-sigChan

	log.Println("Shutting down...")
	return nil
}

// Example_EnvironmentVariables shows required environment variables
func Example_EnvironmentVariables() {
	// Telegram
	// export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"

	// Discord
	// export DISCORD_BOT_TOKEN="your_discord_bot_token"

	// Slack
	// export SLACK_BOT_TOKEN="xoxb-your-slack-bot-token"
	// export SLACK_APP_TOKEN="xapp-your-slack-app-token"

	// Feishu
	// export FEISHU_APP_ID="cli_your_feishu_app_id"
	// export FEISHU_APP_SECRET="your_feishu_app_secret"

	// WeWork
	// export WEWORK_CORP_ID="your_wework_corp_id"
	// export WEWORK_CORP_SECRET="your_wework_corp_secret"
	// export WEWORK_AGENT_ID="your_wework_agent_id"
	// export WEWORK_TOKEN="your_wework_token"
	// export WEWORK_ENCODING_AES_KEY="your_wework_encoding_aes_key"

	// DingTalk
	// export DINGTALK_APP_KEY="your_dingtalk_app_key"
	// export DINGTALK_APP_SECRET="your_dingtalk_app_secret"
	// export DINGTALK_AGENT_ID="your_dingtalk_agent_id"
	// export DINGTALK_AGENT_SECRET="your_dingtalk_agent_secret"

	// QQ
	// export QQ_BOT_ENABLED="true"
	// export QQ_BOT_API_BASE="http://127.0.0.1:3000"
	// export QQ_BOT_SELF_ID="your_qq_bot_id"
}

// Example_ConfigFileStructure shows the recommended directory structure
func Example_ConfigFileStructure() {
	/*
		AgentFramework/
		├── config/
		│   ├── channels.yaml           # Main channels configuration
		│   ├── channels.example.yaml   # Example configuration
		│   ├── channels.example.json   # JSON format example
		│   └── channels.generated.yaml # Auto-generated from env vars
		├── pkg/
		│   └── channels/
		│       ├── types.go             # Core types
		│       ├── adapter.go           # Adapter interface
		│       ├── manager.go           # Channel manager
		│       ├── router.go            # Message router
		│       ├── config.go            # Configuration
		│       └── adapters/
		│           ├── common.go        # Common adapter base
		│           ├── telegram.go      # Telegram adapter
		│           ├── discord.go       # Discord adapter
		│           ├── slack.go         # Slack adapter
		│           ├── feishu.go        # Feishu adapter
		│           ├── wework.go        # WeWork adapter
		│           ├── dingtalk.go      # DingTalk adapter
		│           └── qq.go            # QQ adapter
		└── examples/
		    └── channels_integration.go # This file
	*/
}

// Example_CustomMessageHandler demonstrates custom message handling
func Example_CustomMessageHandler() {
	// Custom message handler with filtering and routing

	// type CustomHandler struct {
	//     manager *channels.Manager
	//     agent   *agent.Host
	// }

	// func (h *CustomHandler) HandleMessage(ctx context.Context, msg *channels.Message) error {
	//     // Filter messages
	//     if shouldIgnore(msg) {
	//         return nil
	//     }

	//     // Transform message
	//     transformed := transformMessage(msg)

	//     // Route to appropriate handler
	//     switch msg.Type {
	//     case channels.MessageTypeCommand:
	//         return h.handleCommand(ctx, transformed)
	//     case channels.MessageTypeText:
	//         return h.handleText(ctx, transformed)
	//     case channels.MessageTypeImage:
	//         return h.handleImage(ctx, transformed)
	//     default:
	//         return h.handleDefault(ctx, transformed)
	//     }
	// }

	// func (h *CustomHandler) handleCommand(ctx context.Context, msg *channels.Message) error {
	//     // Parse command and execute
	//     cmd := parseCommand(msg.Text)
	//     return executeCommand(cmd)
	// }
}

// Example_MultiChannelAgent shows how to create an agent that responds
// to messages from multiple channels
func Example_MultiChannelAgent() {
	// This example demonstrates creating an agent that can handle
	// messages from multiple channels simultaneously

	// type MultiChannelAgent struct {
	//     bridge      *ChannelBridge
	//     sessions    map[string]*Session  // Per-user sessions
	//     context     *agent.Context
	// }

	// func (a *MultiChannelAgent) ProcessMessage(msg *channels.Message) error {
	//     // Get or create session for user
	//     sessionKey := fmt.Sprintf("%s:%s", msg.ChannelType, msg.From.ID)
	//     session := a.getOrCreateSession(sessionKey)

	//     // Add message to session history
	//     session.History = append(session.History, agent.Message{
	//         Role:    "user",
	//         Content: msg.Text,
	//     })

	//     // Run agent
	//     response, err := a.context.Run(session.History)
	//     if err != nil {
	//         return err
	//     }

	//     // Add response to history
	//     session.History = append(session.History, agent.Message{
	//         Role:    "assistant",
	//         Content: response,
	//     })

	//     // Send response back through channel
	//     opts := channels.MessageSendOptions{
	//         ReplyTo: msg.ID,
	//     }
	//     _, err = a.bridge.SendMessage(msg.ChannelID, response, opts)
	//     return err
	// }

	_ = fmt.Sprintf("example")
}

// Example_ChannelSpecificFeatures demonstrates platform-specific features
func Example_ChannelSpecificFeatures() {
	// Telegram: Inline keyboards
	// telegramKeyboard := [][]tb.InlineButton{
	//     {
	//         {Text: "Yes", Data: "yes"},
	//         {Text: "No", Data: "no"},
	//     },
	// }

	// Discord: Embeds
	// discordEmbed := &discordgo.MessageEmbed{
	//     Title:       "Embed Title",
	//     Description: "Embed Description",
	//     Fields: []*discordgo.MessageEmbedField{
	//         {Name: "Field1", Value: "Value1"},
	//     },
	// }

	// Slack: Blocks
	// slackBlocks := []slack.Block{
	//     slack.NewSectionBlock(
	//         &slack.TextBlockObject{
	//             Type: slack.PlainTextType,
	//             Text: "Section text",
	//         },
	//         nil, nil,
	//     ),
	// }

	// Feishu: Cards
	// feishuCard := map[string]interface{}{
	//     "config": map[string]interface{}{
	//         "wide_screen_mode": true,
	//     },
	//     "elements": []interface{}{
	//         map[string]interface{}{
	//             "tag":  "div",
	//             "text": map[string]interface{}{
	//                 "tag":  "plain_text",
	//                 "content": "Card content",
	//             },
	//         },
	//     },
	// }

	_ = "example"
}
