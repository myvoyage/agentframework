// Agent Framework - Lark/Feishu Channel Commands
// Based on OpenClaw Lark Integration Patterns
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"AgentFramework/pkg/channels/lark"
)

// Lark channel flags
var (
	_larkDomain        string
	_larkAppID         string
	_larkAppSecret     string
	_larkMode          string
	_larkPort          int
	_larkEncryptKey    string
	_larkVerifyToken   string
	_larkDMPolicy      string
	_larkGroupPolicy   string
)

func init() {
	larkCmd.PersistentFlags().StringVar(&_larkDomain, "domain", "feishu", "飞书域: feishu(国内版) 或 lark(国际版)")
	larkCmd.PersistentFlags().StringVar(&_larkAppID, "app-id", "", "飞书应用 AppID (cli_xxx)")
	larkCmd.PersistentFlags().StringVar(&_larkAppSecret, "app-secret", "", "飞书应用 AppSecret")
	larkCmd.PersistentFlags().StringVar(&_larkMode, "mode", "websocket", "连接模式: websocket(推荐) 或 webhook")
	larkCmd.PersistentFlags().IntVar(&_larkPort, "port", 8089, "Webhook 端口 (webhook 模式)")
	larkCmd.PersistentFlags().StringVar(&_larkEncryptKey, "encrypt-key", "", "消息加密 Key")
	larkCmd.PersistentFlags().StringVar(&_larkVerifyToken, "verify-token", "", "验证 Token")
	larkCmd.PersistentFlags().StringVar(&_larkDMPolicy, "dm-policy", "pairing", "私信策略: pairing/allowlist/open/disabled")
	larkCmd.PersistentFlags().StringVar(&_larkGroupPolicy, "group-policy", "open", "群聊策略: open/allowlist/disabled")
}

// larkCmd represents the lark command
var larkCmd = &cobra.Command{
	Use:   "lark",
	Short: "飞书/Lark 渠道管理",
	Long: `管理飞书/Lark 渠道配置和连接。

基于官方 OpenClaw 飞书插件 (@larksuite/openclaw-lark) 实现模式。

连接模式:
  websocket - WebSocket 长连接 (推荐，无需公网 URL)
  webhook   - HTTP Webhook (需要公网 URL)

会话策略 (参考官方插件):
  私信策略: pairing(配对模式)/allowlist(白名单)/open(开放)/disabled(禁用)
  群聊策略: open(开放)/allowlist(白名单)/disabled(禁用)

示例:
  # 安装官方插件
  af lark install

  # 配置飞书应用
  af lark config --app-id cli_xxx --app-secret xxx

  # 启动 WebSocket 模式
  af lark serve --mode websocket

  # 启动 Webhook 模式
  af lark serve --mode webhook --port 8089
`,
}

// larkInstallCmd installs the official Lark OpenClaw plugin
var larkInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装飞书官方 OpenClaw 插件",
	Long: `安装字节跳动官方飞书 OpenClaw 插件。

官方插件: @larksuite/openclaw-lark-tools, @larksuite/openclaw-lark

安装命令:
  npm install -g @larksuite/openclaw-lark-tools
  npx @larksuite/openclaw-lark install

特性:
  - WebSocket 长连接，无需公网 URL
  - 内网穿透支持
  - 流式卡片回复
  - 完整的飞书 API 支持
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  飞书官方 OpenClaw 插件安装")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("执行以下命令安装官方插件:")
		fmt.Println()
		fmt.Println("  # 方式一: 使用 npm")
		fmt.Println("  npm install -g @larksuite/openclaw-lark-tools")
		fmt.Println()
		fmt.Println("  # 方式二: 使用 npx")
		fmt.Println("  npx @larksuite/openclaw-lark install")
		fmt.Println()
		fmt.Println("官方文档: https://docs.openclaw.ai/zh-CN/channels/feishu")
		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Println("安装后配置:")
		fmt.Println("  1. 在飞书开放平台创建企业自建应用")
		fmt.Println("  2. 获取 AppID 和 AppSecret")
		fmt.Println("  3. 配置事件订阅 (WebSocket 模式无需配置 URL)")
		fmt.Println("  4. 运行: af lark config --app-id cli_xxx --app-secret xxx")
		fmt.Println("────────────────────────────────────────────────────────────")
		return nil
	},
}

// larkConfigCmd configures Lark channel
var larkConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "配置飞书渠道",
	Long: `配置飞书/Lark 渠道参数。

配置项:
  --app-id       飞书应用 AppID (cli_xxx)
  --app-secret   飞书应用 AppSecret
  --mode         连接模式: websocket(推荐) 或 webhook
  --domain       飞书域: feishu(国内版) 或 lark(国际版)
  --dm-policy    私信策略
  --group-policy 群聊策略

示例:
  af lark config --app-id cli_xxx --app-secret xxx
  af lark config --mode websocket --domain feishu
  af lark config --dm-policy pairing --group-policy open
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  飞书渠道配置")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()

		cfg := loadLarkConfig()

		fmt.Printf("  域:           %s\n", cfg.Domain)
		fmt.Printf("  AppID:        %s\n", maskSecret(cfg.AppID))
		fmt.Printf("  连接模式:     %s\n", cfg.ConnectionMode)
		fmt.Printf("  私信策略:     %s\n", cfg.DMPolicy)
		fmt.Printf("  群聊策略:     %s\n", cfg.GroupPolicy)
		if cfg.Port > 0 {
			fmt.Printf("  Webhook端口:  %d\n", cfg.Port)
		}

		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Println("配置已保存，使用 'af lark serve' 启动服务")
		return nil
	},
}

// larkServeCmd starts Lark channel server
var larkServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动飞书渠道服务",
	Long: `启动飞书/Lark 渠道服务。

连接模式:
  websocket - WebSocket 长连接，无需公网 URL
  webhook   - HTTP Webhook，需要公网 URL

示例:
  af lark serve --mode websocket
  af lark serve --mode webhook --port 8089
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  飞书渠道服务启动")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()

		cfg := loadLarkConfig()

		if cfg.AppID == "" || cfg.AppSecret == "" {
			return fmt.Errorf("请先配置 AppID 和 AppSecret: af lark config --app-id xxx --app-secret xxx")
		}

		channel := lark.NewChannel(cfg)

		// Set up message handler
		channel.OnMessage(func(ctx context.Context, msg *lark.Message) {
			fmt.Printf("[Lark] 收到消息: %s -> %s\n", msg.From.Name, msg.Content)

			// Simple echo response
			content, _ := msg.Content.(string)
			response := fmt.Sprintf("收到: %s", content)
			channel.SendText(ctx, msg.ChatID, response)
		})

		fmt.Printf("  域:           %s\n", cfg.Domain)
		fmt.Printf("  连接模式:     %s\n", cfg.ConnectionMode)
		fmt.Printf("  AppID:        %s\n", maskSecret(cfg.AppID))
		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Println("按 Ctrl+C 停止服务")
		fmt.Println()

		ctx := context.Background()
		if err := channel.Start(ctx); err != nil {
			return fmt.Errorf("启动失败: %w", err)
		}

		return nil
	},
}

// larkStatusCmd shows Lark channel status
var larkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看飞书渠道状态",
	Long: `查看飞书/Lark 渠道连接状态和统计信息。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  飞书渠道状态")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()

		cfg := loadLarkConfig()

		fmt.Println("配置信息:")
		fmt.Printf("  域:           %s\n", cfg.Domain)
		fmt.Printf("  连接模式:     %s\n", cfg.ConnectionMode)
		fmt.Printf("  AppID:        %s\n", maskSecret(cfg.AppID))
		fmt.Printf("  私信策略:     %s\n", cfg.DMPolicy)
		fmt.Printf("  群聊策略:     %s\n", cfg.GroupPolicy)
		fmt.Println()

		// Check configuration status
		if cfg.AppID == "" || cfg.AppSecret == "" {
			fmt.Println("状态: ❌ 未配置")
			fmt.Println("  请先运行: af lark config --app-id xxx --app-secret xxx")
		} else {
			fmt.Println("状态: ✅ 已配置")
			fmt.Println("  使用 'af lark serve' 启动服务")
		}

		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		return nil
	},
}

// larkTestCmd tests Lark channel connection
var larkTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试飞书渠道连接",
	Long: `测试飞书/Lark 渠道 API 连接。

测试内容:
  - 获取 Access Token
  - 获取 Bot 信息
  - 验证权限配置
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  飞书渠道连接测试")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()

		cfg := loadLarkConfig()

		if cfg.AppID == "" || cfg.AppSecret == "" {
			return fmt.Errorf("请先配置 AppID 和 AppSecret")
		}

		fmt.Println("测试 1: 获取 Access Token...")
		channel := lark.NewChannel(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := channel.Start(ctx); err != nil {
			fmt.Printf("  ❌ 失败: %v\n", err)
			return err
		}
		fmt.Println("  ✅ Access Token 获取成功")

		fmt.Println()
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Println("测试完成！所有检查通过。")
		return nil
	},
}

// larkCardCmd tests interactive card
var larkCardCmd = &cobra.Command{
	Use:   "card [to]",
	Short: "发送测试卡片消息",
	Long: `发送测试交互卡片消息。

示例:
  af lark card ou_xxx       # 发送给用户
  af lark card oc_xxx       # 发送到群聊
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		to := args[0]

		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println("  发送测试卡片")
		fmt.Println("════════════════════════════════════════════════════════════")
		fmt.Println()

		cfg := loadLarkConfig()

		if cfg.AppID == "" || cfg.AppSecret == "" {
			return fmt.Errorf("请先配置 AppID 和 AppSecret")
		}

		channel := lark.NewChannel(cfg)
		ctx := context.Background()

		// Start channel to get token
		if err := channel.Start(ctx); err != nil {
			return fmt.Errorf("启动失败: %w", err)
		}

		// Create test card
		card := lark.NewCard("测试卡片", "这是一个测试消息卡片").
			AddMarkdown("**测试内容**\n\n这是一个 Markdown 段落").
			AddButton("点击按钮", "test_value")

		if err := channel.SendCard(ctx, to, card); err != nil {
			return fmt.Errorf("发送失败: %w", err)
		}

		fmt.Printf("✅ 卡片已发送到: %s\n", to)
		return nil
	},
}

// loadLarkConfig loads Lark configuration from flags
func loadLarkConfig() *lark.Config {
	cfg := lark.DefaultConfig()

	// Override with flags
	if _larkDomain != "" {
		cfg.Domain = _larkDomain
	}
	if _larkAppID != "" {
		cfg.AppID = _larkAppID
	}
	if _larkAppSecret != "" {
		cfg.AppSecret = _larkAppSecret
	}
	if _larkMode != "" {
		cfg.ConnectionMode = _larkMode
	}
	if _larkPort > 0 {
		cfg.Port = _larkPort
	}
	if _larkEncryptKey != "" {
		cfg.EncryptKey = _larkEncryptKey
	}
	if _larkVerifyToken != "" {
		cfg.VerifyToken = _larkVerifyToken
	}
	if _larkDMPolicy != "" {
		cfg.DMPolicy = _larkDMPolicy
	}
	if _larkGroupPolicy != "" {
		cfg.GroupPolicy = _larkGroupPolicy
	}

	return cfg
}

func init() {
	larkCmd.AddCommand(larkInstallCmd)
	larkCmd.AddCommand(larkConfigCmd)
	larkCmd.AddCommand(larkServeCmd)
	larkCmd.AddCommand(larkStatusCmd)
	larkCmd.AddCommand(larkTestCmd)
	larkCmd.AddCommand(larkCardCmd)

	rootCmd.AddCommand(larkCmd)
}
