// Gateway Server - Combined WS + HTTP server on single port
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is the main gateway server
type Server struct {
	config *Config
	svc    *Service
	ws     *WebSocketHandler
	http   *HTTPServer
	server *http.Server
	addr   string
}

// NewServer creates a new gateway server
func NewServer(cfg *Config, svc *Service) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.Port)

	ws := NewWebSocketHandler(svc, cfg)
	httpServer := NewHTTPServer(svc, cfg)

	mux := http.NewServeMux()
	mux.Handle("/", ws)
	mux.Handle("/v1/", httpServer.Handler())
	mux.Handle("/health", httpServer.Handler())
	mux.Handle("/status", httpServer.Handler())
	mux.Handle("/tools/", httpServer.Handler())

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		config: cfg,
		svc:    svc,
		ws:     ws,
		http:   httpServer,
		server: server,
		addr:   addr,
	}
}

// Start starts the gateway server
func (s *Server) Start(ctx context.Context) error {
	if s.config.Gateway.Force {
		if err := s.killExisting(); err != nil {
			log.Printf("[Gateway] Warning: could not kill existing process: %v", err)
		}
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to bind port %d: %w (use --force to kill existing)", s.config.Gateway.Port, err)
	}

	if err := s.svc.Start(ctx); err != nil {
		return err
	}

	go func() {
		log.Printf("[Gateway] Listening on %s (WS: /, HTTP: /v1/*, /health, /tools/*)", s.addr)
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[Gateway] Server error: %v", err)
		}
	}()

	s.waitForShutdown(ctx)
	return nil
}

func (s *Server) waitForShutdown(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		log.Printf("[Gateway] Shutting down...")
		s.Shutdown(ctx)
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	s.svc.Stop(ctx)
	s.server.Shutdown(shutdownCtx)
	log.Printf("[Gateway] Shutdown complete")
}

func (s *Server) reload() {
	cfg, err := LoadConfig(ConfigPath())
	if err != nil {
		log.Printf("[Gateway] Failed to reload config: %v", err)
		return
	}
	cfg.ApplyFlags(s.config.Gateway.Port, s.config.Gateway.Verbose, false, false)
	log.Printf("[Gateway] Configuration reloaded")
}

func (s *Server) killExisting() error {
	addr := fmt.Sprintf(":%d", s.config.Gateway.Port)
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return nil
	}
	conn.Close()
	return fmt.Errorf("port %d is in use", s.config.Gateway.Port)
}

// RunGateway runs the gateway as a standalone process
func RunGateway(cfg *Config) error {
	ctx := context.Background()

	if cfg == nil {
		var err error
		cfg, err = LoadConfig(ConfigPath())
		if err != nil {
			cfg = DefaultConfig()
		}
	}

	if cfg.Gateway.Dev {
		devCfg := DevConfig()
		if cfg.Gateway.Port == 0 {
			cfg.Gateway.Port = devCfg.Gateway.Port
		}
		cfg.Agents.Workspace = devCfg.Agents.Workspace
	}

	svc := NewService(cfg, nil)
	server := NewServer(cfg, svc)

	log.Printf("[Gateway] Starting OpenClaw Gateway v1.0.0")
	log.Printf("[Gateway] Port: %d", cfg.Gateway.Port)
	log.Printf("[Gateway] Auth: token=%s password=%s", maskStr(cfg.Auth.Token), maskStr(cfg.Auth.Password))
	log.Printf("[Gateway] Canvas: enabled=%v port=%d", cfg.CanvasHost.Enabled, cfg.CanvasHost.Port)
	log.Printf("[Gateway] Hot-reload: mode=%s", cfg.Reload.Mode)

	return server.Start(ctx)
}

func maskStr(s string) string {
	if len(s) == 0 {
		return "(none)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****"
}
