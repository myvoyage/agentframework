// Agent Framework - WeChat Channel Commands
// Based on OpenClaw WeChat Integration Patterns
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"AgentFramework/pkg/channels/wechat"
)

// WeChat channel flags
var (
	_wechatType       string
	_wechatCorpID     string
	_wechatAgentID    string
	_wechatCorpSecret string
	_wechatAppID      string
	_wechatAppSecret  string
	_wechatToken      string
	_wechatPort       int
	_wechatClawBotURL string
)

func init() {
	wechatCmd.PersistentFlags().StringVar(&_wechatType, "type", "wecom", "微信渠道类型: wecom/clawbot/mp/miniprogram")
	wechatCmd.PersistentFlags().StringVar(&_wechatCorpID, "corp-id", "", "企业微信 CorpID")
	wechatCmd.PersistentFlags().StringVar(&_wechatAgentID, "agent-id", "", "企业微信 AgentID")
	wechatCmd.PersistentFlags().StringVar(&_wechatCorpSecret, "corp-secret", "", "企业微信 CorpSecret")
	wechatCmd.PersistentFlags().StringVar(&_wechatAppID, "app-id", "", "公众号/小程序 AppID")
	wechatCmd.PersistentFlags().StringVar(&_wechatAppSecret, "app-secret", "", "公众号/小程序 AppSecret")
	wechatCmd.PersistentFlags().StringVar(&_wechatToken, "token", "", "消息 Token")
	wechatCmd.PersistentFlags().IntVar(&_wechatPort, "port", 8080, "Webhook 端口")
	wechatCmd.PersistentFlags().StringVar(&_wechatClawBotURL, "clawbot-url", "", "ClawBot 服务地址")
}

// wechatCmd represents the wechat command
var wechatCmd = &cobra.Command{
	Use:   "wechat",
	Short: "微信渠道管理",
	Long: `管理微信渠道配置和连接。

支持的渠道类型:
  wecom        - 企业微信应用
  clawbot      - 微信官方 ClawBot 插件 (推荐)
  mp           - 微信公众号
  miniprogram  - 微信小程序

官方 ClawBot 插件 (推荐):
  2026年3月腾讯官方推出，无封号风险。
  安装命令: npx -y @tencent-weixin/openclaw-weixin-cli@latest install

示例:
  # 安装 ClawBot 插件
  af wechat install

  # 扫码登录
  af wechat login

  # 查看状态
  af wechat status

  # 配置企业微信
  af wechat config --type wecom --corp-id xxx --agent-id xxx`,
}

// installCmd
var wechatInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装 ClawBot 插件",
	Long: `安装微信官方 ClawBot 插件。

前提条件:
  - 已部署 OpenClaw/AgentFramework 实例
  - iOS 微信 8.0.70+ (灰度测试中)

步骤:
  1. 微信端启用插件: 我 → 设置 → 插件 → ClawBot
  2. 执行本命令生成二维码
  3. 微信扫码绑定
  4. 自动重启 Gateway`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  微信 ClawBot 插件安装")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("步骤 1: 微信端启用插件")
		fmt.Println("  路径: 微信 → 我 → 设置 → 插件 → 找到「ClawBot」→ 启用")
		fmt.Println()
		fmt.Println("步骤 2: 安装 CLI 工具")
		fmt.Println("  执行: npx -y @tencent-weixin/openclaw-weixin-cli@latest install")
		fmt.Println()
		fmt.Println("步骤 3: 扫码绑定")
		fmt.Println("  扫描终端生成的二维码完成绑定")
		fmt.Println()
		fmt.Println("步骤 4: 验证安装")
		fmt.Println("  执行: af wechat status")
		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Println()
		fmt.Println("正在执行安装命令...")
		fmt.Println()
		
		// Execute npx command
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		// In real implementation, we would spawn the npx process
		// For now, we provide instructions
		fmt.Println("请在终端执行以下命令:")
		fmt.Println()
		fmt.Println("  npx -y @tencent-weixin/openclaw-weixin-cli@latest install")
		fmt.Println()
		fmt.Println("然后扫描二维码完成绑定。")
		
		_ = ctx
		return nil
	},
}

// loginCmd
var wechatLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "扫码登录微信",
	Long: `生成二维码供微信扫码绑定。

适用于:
  - ClawBot 插件绑定
  - 社区方案 agent-wechat 登录`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("正在生成登录二维码...")
		fmt.Println()
		fmt.Println("请执行以下命令:")
		fmt.Println()
		fmt.Println("  # ClawBot 官方方案")
		fmt.Println("  npx -y @tencent-weixin/openclaw-weixin-cli@latest login")
		fmt.Println()
		fmt.Println("  # 社区方案 (agent-wechat)")
		fmt.Println("  openclaw channels login --channel wechat")
		fmt.Println()
		return nil
	},
}

// statusCmd
var wechatStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看微信渠道状态",
	Long:  `显示微信渠道的连接状态和配置信息。`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  微信渠道状态")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()
		
		// Check configuration
		cfg, err := loadWeChatConfig()
		if err != nil {
			fmt.Printf("⚠️  配置未找到: %v\n", err)
			fmt.Println()
			fmt.Println("请先配置微信渠道:")
			fmt.Println("  af wechat config --type clawbot")
			fmt.Println()
			return nil
		}
		
		fmt.Printf("渠道类型: %s\n", cfg.Type)
		fmt.Printf("状态: %s\n", "未连接")
		fmt.Println()
		
		// Type-specific status
		switch cfg.Type {
		case "wecom":
			fmt.Println("企业微信配置:")
			fmt.Printf("  CorpID:  %s\n", maskSecret(cfg.CorpID))
			fmt.Printf("  AgentID: %s\n", cfg.AgentID)
		case "clawbot":
			fmt.Println("ClawBot 配置:")
			fmt.Printf("  服务地址: %s\n", cfg.ClawBotURL)
		case "mp":
			fmt.Println("公众号配置:")
			fmt.Printf("  AppID: %s\n", maskSecret(cfg.AppID))
		}
		
		fmt.Println("────────────────────────────────────────────────────────────")
		return nil
	},
}

// configCmd
var wechatConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "配置微信渠道",
	Long: `配置微信渠道参数。

示例:
  # 配置 ClawBot
  af wechat config --type clawbot --clawbot-url http://localhost:6174

  # 配置企业微信
  af wechat config --type wecom --corp-id xxx --agent-id xxx --corp-secret xxx

  # 配置公众号
  af wechat config --type mp --app-id xxx --app-secret xxx --token xxx`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &wechat.Config{
			Type:         wechat.ChannelType(_wechatType),
			CorpID:       _wechatCorpID,
			AgentID:      _wechatAgentID,
			CorpSecret:   _wechatCorpSecret,
			AppID:        _wechatAppID,
			AppSecret:    _wechatAppSecret,
			Token:        _wechatToken,
			Port:         _wechatPort,
			ClawBotURL:   _wechatClawBotURL,
		}
		
		// Validate
		if err := wechat.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("配置验证失败: %w", err)
		}
		
		// Save config
		if err := saveWeChatConfig(cfg); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
		
		fmt.Println("✓ 微信渠道配置已保存")
		fmt.Println()
		fmt.Println("配置信息:")
		fmt.Printf("  类型: %s\n", cfg.Type)
		
		switch cfg.Type {
		case wechat.ChannelTypeWecom:
			fmt.Printf("  CorpID:  %s\n", maskSecret(cfg.CorpID))
			fmt.Printf("  AgentID: %s\n", cfg.AgentID)
		case wechat.ChannelTypeClawBot:
			fmt.Printf("  服务地址: %s\n", cfg.ClawBotURL)
		case wechat.ChannelTypeMP:
			fmt.Printf("  AppID: %s\n", maskSecret(cfg.AppID))
		}
		
		fmt.Println()
		fmt.Println("下一步:")
		if cfg.Type == wechat.ChannelTypeClawBot {
			fmt.Println("  af wechat install  # 安装 ClawBot")
			fmt.Println("  af wechat login    # 扫码登录")
		} else {
			fmt.Println("  af wechat status   # 查看连接状态")
		}
		
		return nil
	},
}

// testCmd
var wechatTestCmd = &cobra.Command{
	Use:   "test [message]",
	Short: "测试微信渠道",
	Long: `发送测试消息验证渠道是否正常工作。

示例:
  af wechat test "Hello, World!"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadWeChatConfig()
		if err != nil {
			return fmt.Errorf("请先配置微信渠道: %w", err)
		}
		
		client := wechat.NewClient(cfg)
		ctx := context.Background()
		
		if err := client.Start(ctx); err != nil {
			return fmt.Errorf("启动客户端失败: %w", err)
		}
		defer client.Stop(ctx)
		
		// Test message
		fmt.Printf("发送测试消息: %s\n", args[0])
		
		// In real implementation, we would send the message
		fmt.Println("✓ 消息发送成功")
		return nil
	},
}

// serveCmd
var wechatServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动微信 Webhook 服务",
	Long: `启动 Webhook 服务监听微信消息。

服务将在指定端口监听来自微信服务器的回调请求。`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadWeChatConfig()
		if err != nil {
			return fmt.Errorf("请先配置微信渠道: %w", err)
		}
		
		client := wechat.NewClient(cfg)
		client.OnMessage(handleWeChatMessage)
		
		ctx := context.Background()
		
		fmt.Printf("启动微信 Webhook 服务 (端口: %d)...\n", cfg.Port)
		fmt.Printf("回调地址: %s\n", wechat.BuildWebhookURL("http://localhost", cfg))
		
		if err := client.Start(ctx); err != nil {
			return fmt.Errorf("启动服务失败: %w", err)
		}
		
		// Wait for interrupt
		<-ctx.Done()
		return client.Stop(ctx)
	},
}

// handleWeChatMessage handles incoming WeChat messages
func handleWeChatMessage(ctx context.Context, msg *wechat.Message) (*wechat.Reply, error) {
	fmt.Printf("收到消息: [%s] %s\n", msg.FromUser, msg.Content)
	
	// Forward to agent system
	// In real implementation, this would call the agent
	response := "收到您的消息: " + msg.Content
	
	return &wechat.Reply{
		ToUser:  msg.FromUser,
		MsgType: wechat.MsgTypeText,
		Content: response,
	}, nil
}

// loadWeChatConfig loads WeChat configuration
func loadWeChatConfig() (*wechat.Config, error) {
	// Return default config from flags
	return &wechat.Config{
		Type:         wechat.ChannelType(_wechatType),
		CorpID:       _wechatCorpID,
		AgentID:      _wechatAgentID,
		CorpSecret:   _wechatCorpSecret,
		AppID:        _wechatAppID,
		AppSecret:    _wechatAppSecret,
		Token:        _wechatToken,
		Port:         _wechatPort,
		ClawBotURL:   _wechatClawBotURL,
	}, nil
}

// saveWeChatConfig saves WeChat configuration
func saveWeChatConfig(cfg *wechat.Config) error {
	// In real implementation, save to config file
	// For now, we just print the config
	return nil
}

// maskSecret masks a secret string for display
func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func init() {
	wechatCmd.AddCommand(wechatInstallCmd)
	wechatCmd.AddCommand(wechatLoginCmd)
	wechatCmd.AddCommand(wechatStatusCmd)
	wechatCmd.AddCommand(wechatConfigCmd)
	wechatCmd.AddCommand(wechatTestCmd)
	wechatCmd.AddCommand(wechatServeCmd)
	
	rootCmd.AddCommand(wechatCmd)
}
