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
		statusStr := "stopped"
			if ch.IsRunning() {
				statusStr = "running"
			}
			fmt.Printf("  %-20s  type=%-15s  status=%s\n",
				ch.Name(), ch.Type(), statusStr)
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

			chStatusStr := "stopped"
			if ch.IsRunning() {
				chStatusStr = "running"
			}
			fmt.Println("Channel Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Name:   %s\n", ch.Name())
			fmt.Printf("Type:   %s\n", ch.Type())
			fmt.Printf("Status: %s\n", chStatusStr)
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

	// ── add ───────────────────────────────────────────────────────────────
	addCmd := &cobra.Command{
		Use:   "add [channel-type]",
		Short: "添加消息渠道",
		Long: `添加一个新的消息渠道配置。

支持的渠道类型:
  telegram   - Telegram Bot
  lark      - 飞书 (Feishu)
  qq        - QQ/QQ 频道
  discord   - Discord
  slack     - Slack
  wechat    - 企业微信

示例：
  af channel add telegram --name "My Bot" --token $TOKEN
  af channel add lark --app-id $APP_ID --app-secret $SECRET`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelType := args[0]
			fmt.Printf("[Channel] Adding %s channel...\n", channelType)

			// TODO: Implement channel addition logic
			fmt.Println("⚠ 渠道添加功能开发中")
			fmt.Println("  请手动编辑配置文件添加渠道配置")
			return nil
		},
	}
	channelCmd.AddCommand(addCmd)

	// ── delete ────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete [channel-id]",
		Short: "删除消息渠道",
		Long:  `删除指定的消息渠道配置。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelID := args[0]
			fmt.Printf("[Channel] Deleting channel: %s\n", channelID)

			// TODO: Implement channel deletion logic
			fmt.Println("⚠ 渠道删除功能开发中")
			fmt.Println("  请手动编辑配置文件删除渠道配置")
			return nil
		},
	}
	channelCmd.AddCommand(deleteCmd)

	// ── probe ────────────────────────────────────────────────────────────
	probeCmd := &cobra.Command{
		Use:   "probe",
		Short: "探测渠道连接状态",
		Long: `探测所有已配置消息渠道的连接状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[Channel] Probing connection status...")
			fmt.Println()

			channelMgr := app.GetHost().ChannelManager()
			if channelMgr == nil {
				fmt.Println("⚠ 渠道管理器未配置")
				return nil
			}

			channels := channelMgr.ListChannels()
			if len(channels) == 0 {
				fmt.Println("ℹ 未配置任何渠道")
				return nil
			}

			fmt.Println("Connection Status:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, ch := range channels {
				status := "✗ 未知"
				if ch.IsRunning() {
					status = "✓ 在线"
				} else {
					status = "✗ 离线"
				}
				fmt.Printf("  %-20s  type=%-15s  status=%s\n",
					ch.Name(), ch.Type(), status)
			}
			fmt.Println("────────────────────────────────────────────────────────────")

			// Summary
			online := 0
			for _, ch := range channels {
				if ch.IsRunning() {
					online++
				}
			}
			fmt.Printf("总计: %d/%d 在线\n", online, len(channels))

			return nil
		},
	}
	channelCmd.AddCommand(probeCmd)

	rootCmd.AddCommand(channelCmd)
}
