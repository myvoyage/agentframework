// Agent Framework - Channel Commands (Messaging)
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// channelCmd represents the channel command
var channelCmd = &cobra.Command{
	Use:     "channel",
	Aliases: []string{"ch", "msg"},
	Short:   "管理消息通道",
	Long: `管理 Agent 的消息通道（Slack、Telegram、钉钉、飞书等）。
消息通道允许 Agent 接收和发送来自外部平台的消息。

注意：需要在配置中启用 messaging 功能。`,
}

// addChannelCommands adds channel-related commands to root command
func addChannelCommands() {
	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有消息通道",
		Long:  `列出当前已注册的所有消息通道及其状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			channelMgr := app.GetHost().ChannelManager()
			if channelMgr == nil {
				return fmt.Errorf("channel manager is not configured.\nAdd messaging config to your host.yaml:\n  messaging:\n    enabled: true")
			}

			channels := channelMgr.ListChannels()
			if len(channels) == 0 {
				fmt.Println("No channels registered")
				return nil
			}

			fmt.Println("Message Channels:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, ch := range channels {
				fmt.Printf("  %-20s  type=%-15s  status=%s\n",
					ch.ID(), ch.Type(), ch.Status())
			}
			fmt.Printf("Total: %d channel(s)\n", len(channels))
			return nil
		},
	}
	channelCmd.AddCommand(listCmd)

	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info [channel-id]",
		Short: "获取通道详情",
		Long:  `获取指定消息通道的详细信息。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelMgr := app.GetHost().ChannelManager()
			if channelMgr == nil {
				return fmt.Errorf("channel manager is not configured")
			}

			ch, err := channelMgr.GetChannel(args[0])
			if err != nil {
				return fmt.Errorf("channel '%s' not found: %w", args[0], err)
			}

			fmt.Println("Channel Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("ID:     %s\n", ch.ID())
			fmt.Printf("Type:   %s\n", ch.Type())
			fmt.Printf("Status: %s\n", ch.Status())
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	channelCmd.AddCommand(infoCmd)

	// ── status ───────────────────────────────────────────────────────────────
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "查看消息通道管理器状态",
		Long:  `显示消息通道管理器的整体状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			channelMgr := app.GetHost().ChannelManager()
			if channelMgr == nil {
				fmt.Println("Channel Manager: Not configured")
				fmt.Println()
				fmt.Println("To enable messaging, add to your config:")
				fmt.Println("  messaging:")
				fmt.Println("    enabled: true")
				fmt.Println("    enableMetrics: true")
				return nil
			}

			channels := channelMgr.ListChannels()
			fmt.Println("Channel Manager Status:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Status:         Active\n")
			fmt.Printf("Total Channels: %d\n", len(channels))
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	channelCmd.AddCommand(statusCmd)

	// ── messaging-config ─────────────────────────────────────────────────────
	msgCfgCmd := &cobra.Command{
		Use:   "config",
		Short: "显示消息通道配置",
		Long:  `显示当前消息通道的配置参数。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			if cfg.Messaging == nil {
				fmt.Println("Messaging is not configured.")
				return nil
			}

			fmt.Println("Messaging Configuration:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Enabled:        %v\n", cfg.Messaging.Enabled)
			fmt.Printf("EnableMetrics:  %v\n", cfg.Messaging.EnableMetrics)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	channelCmd.AddCommand(msgCfgCmd)

	rootCmd.AddCommand(channelCmd)
}
