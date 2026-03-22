// Gateway Command - OpenClaw Gateway integration for AgentFramework CLI
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"AgentFramework/agent"
	"AgentFramework/core"
	"AgentFramework/gateway"
)

var (
	_gatewayPort     int
	_gatewayVerbose  bool
	_gatewayForce   bool
	_gatewayDev     bool
	_gatewayToken   string
	_gatewayPassword string
	_gatewayConfig  string
)

func addGatewayCommands() {
	gatewayCmd := &cobra.Command{
		Use:   "gateway",
		Short: "启动 OpenClaw Gateway — WebSocket + HTTP 控制平面",
		Long: `启动 OpenClaw Gateway 服务器。

Gateway 提供单端口服务：
- WebSocket 控制平面 (/)
- OpenAI Chat Completions API (/v1/chat/completions)
- OpenAI Responses API (/v1/responses)
- Tools Invoke (/tools/invoke)

支持热重载（SIGUSR1）、多客户端并发、配置热更新。

示例：
	af gateway --port 18640
	af gateway --port 18640 --verbose
  af gateway --force
  af --dev gateway
  af --profile <name> gateway`,
		RunE: runGateway,
	}

	gatewayCmd.Flags().IntVarP(&_gatewayPort, "port", "p", 0, "Gateway 监听端口（默认: 18640，dev: 19001）")
	gatewayCmd.Flags().BoolVarP(&_gatewayVerbose, "verbose", "v", false, "启用详细/调试日志")
	gatewayCmd.Flags().BoolVar(&_gatewayForce, "force", false, "启动前终止占用端口的进程")
	gatewayCmd.Flags().BoolVar(&_gatewayDev, "dev", false, "开发模式（隔离配置/状态/工作区）")
	gatewayCmd.Flags().StringVar(&_gatewayToken, "token", "", "Gateway 认证令牌")
	gatewayCmd.Flags().StringVar(&_gatewayPassword, "password", "", "Gateway 认证密码")
	gatewayCmd.Flags().StringVarP(&_gatewayConfig, "config", "c", "", "配置文件路径")

	rootCmd.AddCommand(gatewayCmd)
}

func runGateway(cmd *cobra.Command, args []string) error {
	cfg := gateway.DefaultConfig()

	if _gatewayPort > 0 {
		cfg.Gateway.Port = _gatewayPort
	}
	cfg.Gateway.Verbose = _gatewayVerbose
	cfg.Gateway.Force = _gatewayForce
	cfg.Gateway.Dev = _gatewayDev

	if _gatewayToken != "" {
		cfg.Auth.Token = _gatewayToken
	}
	if _gatewayPassword != "" {
		cfg.Auth.Password = _gatewayPassword
	}
	if _gatewayConfig != "" {
		cfg.ConfigPath = _gatewayConfig
	}

	if cfg.Gateway.Dev {
		devCfg := gateway.DevConfig()
		if cfg.Gateway.Port == 0 {
			cfg.Gateway.Port = devCfg.Gateway.Port
		}
		cfg.Agents.Workspace = devCfg.Agents.Workspace
	}

	ctx := context.Background()

	// Try to initialize AgentFramework Host
	var host *agent.Host
	if hostCfg, err := agent.LoadHostConfigFile(_gatewayConfig); err == nil && hostCfg != nil {
		modelFactory := agent.NewModelFactoryWithConfig(hostCfg.Models["default"])
		coreApp, err := core.NewApplication(ctx, hostCfg, modelFactory, nil)
		if err == nil {
			if err := coreApp.Initialize(ctx); err == nil {
				host = coreApp.GetHost()
				log.Printf("[Gateway] 已连接 AgentFramework Host")
			}
		}
	}

	if host == nil {
		log.Printf("[Gateway] 无 AgentFramework Host（仅提供 API 网关服务）")
	}

	svc := gateway.NewService(cfg, host)
	srv := gateway.NewServer(cfg, svc)

	// Start server
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for shutdown signal or error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("[Gateway] 收到 %v，正在关闭...", sig)
		srv.Shutdown(ctx)
		return nil
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("gateway error: %w", err)
		}
		return nil
	}
}
