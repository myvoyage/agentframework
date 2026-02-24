// Agent Framework - API Server
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"AgentFramework/core"
)

// Server 统一的 API 服务器，支持桌面应用和 CLI
type Server struct {
	ctx        context.Context
	core       *core.Application
	router     *mux.Router
	httpServer *http.Server
	upgrader   websocket.Upgrader

	// WebSocket 连接管理
	connections map[string]*websocket.Conn
	connMutex   sync.RWMutex

	// 配置
	config ServerConfig
}

// ServerConfig API 服务器配置
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	EnableCORS      bool
	EnableAuth      bool
	AuthSecret      string
}

// DefaultServerConfig 返回默认配置
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:            "localhost",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		EnableCORS:      true,
		EnableAuth:      false,
	}
}

// NewServer 创建新的 API 服务器
func NewServer(ctx context.Context, coreApp *core.Application, config ServerConfig) *Server {
	router := mux.NewRouter()

	server := &Server{
		ctx:         ctx,
		core:        coreApp,
		router:      router,
		connections: make(map[string]*websocket.Conn),
		config:      config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境需要验证 Origin
			},
		},
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册所有 API 路由
func (s *Server) registerRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	// CORS 中间件
	if s.config.EnableCORS {
		s.router.Use(s.corsMiddleware)
	}

	// 健康检查
	api.HandleFunc("/health", s.healthHandler).Methods("GET")

	// ========== 工作流 API ==========
	api.HandleFunc("/workflows", s.listWorkflowsHandler).Methods("GET")
	api.HandleFunc("/workflows", s.createWorkflowHandler).Methods("POST")
	api.HandleFunc("/workflows/{id}", s.getWorkflowHandler).Methods("GET")
	api.HandleFunc("/workflows/{id}", s.updateWorkflowHandler).Methods("PUT")
	api.HandleFunc("/workflows/{id}", s.deleteWorkflowHandler).Methods("DELETE")
	api.HandleFunc("/workflows/{id}/execute", s.executeWorkflowHandler).Methods("POST")
	api.HandleFunc("/workflows/{id}/versions", s.getWorkflowVersionsHandler).Methods("GET")

	// ========== 技能 API ==========
	api.HandleFunc("/skills", s.listSkillsHandler).Methods("GET")
	api.HandleFunc("/skills", s.registerSkillHandler).Methods("POST")
	api.HandleFunc("/skills/{id}", s.getSkillHandler).Methods("GET")
	api.HandleFunc("/skills/{id}", s.updateSkillHandler).Methods("PUT")
	api.HandleFunc("/skills/{id}", s.deleteSkillHandler).Methods("DELETE")
	api.HandleFunc("/skills/{id}/enable", s.enableSkillHandler).Methods("POST")
	api.HandleFunc("/skills/{id}/disable", s.disableSkillHandler).Methods("POST")

	// ========== 配置 API ==========
	api.HandleFunc("/config", s.getConfigHandler).Methods("GET")
	api.HandleFunc("/config", s.updateConfigHandler).Methods("PUT")

	// ========== 文件系统 API ==========
	api.HandleFunc("/files/list", s.listFilesHandler).Methods("GET")
	api.HandleFunc("/files/read", s.readFileHandler).Methods("GET")
	api.HandleFunc("/files/write", s.writeFileHandler).Methods("POST")
	api.HandleFunc("/files/delete", s.deleteFileHandler).Methods("DELETE")
	api.HandleFunc("/files/copy", s.copyFileHandler).Methods("POST")
	api.HandleFunc("/files/move", s.moveFileHandler).Methods("POST")
	api.HandleFunc("/files/create-dir", s.createDirectoryHandler).Methods("POST")

	// ========== Agent API ==========
	api.HandleFunc("/agents", s.listAgentsHandler).Methods("GET")
	api.HandleFunc("/agents/{id}/chat", s.chatHandler).Methods("POST")

	// WebSocket 端点
	s.router.HandleFunc("/ws", s.websocketHandler)
}

// ========== 生命周期管理 ==========

// Start 启动 API 服务器
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	errChan := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// 等待一小段时间确保服务器启动成功
	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	// 关闭所有 WebSocket 连接
	s.connMutex.Lock()
	for connID, conn := range s.connections {
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
		delete(s.connections, connID)
	}
	s.connMutex.Unlock()

	// 关闭 HTTP 服务器
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	return nil
}

// GetCore 返回核心应用实例
func (s *Server) GetCore() *core.Application {
	return s.core
}

// GetRouter 返回 HTTP 路由器（用于测试）
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// ========== 中间件 ==========

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录请求日志
		fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		// 记录响应时间
		fmt.Printf("[%s] %s %s - %dms\n", time.Now().Format("2006-01-02 15:04:05"),
			r.Method, r.URL.Path, time.Since(start).Milliseconds())
	})
}

// recoveryMiddleware 恢复中间件
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic recovered: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ========== 响应辅助函数 ==========

// respondJSON 发送 JSON 响应
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// respondError 发送错误响应
func (s *Server) respondError(w http.ResponseWriter, status int, message string, err error) {
	errorData := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
		"time":    time.Now().Unix(),
	}

	if err != nil {
		errorData["details"] = err.Error()
	}

	s.respondJSON(w, status, errorData)
}

// respondSuccess 发送成功响应
func (s *Server) respondSuccess(w http.ResponseWriter, data interface{}) error {
	return s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
		"time":    time.Now().Unix(),
	})
}

// ========== 处理器辅助函数 ==========

// parseJSONBody 解析请求体
func (s *Server) parseJSONBody(r *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

// getQueryParam 获取查询参数
func (s *Server) getQueryParam(r *http.Request, key string, defaultValue string) string {
	values := r.URL.Query()[key]
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

// getIntQueryParam 获取整数查询参数
func (s *Server) getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	strValue := s.getQueryParam(r, key, "")
	if strValue == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(strValue, "%d", &value); err == nil {
		return value
	}
	return defaultValue
}

// parseJSONBodyFromRequest 从请求中解析 JSON，返回 map[string]interface{}
func (s *Server) parseJSONBodyFromRequest(r *http.Request) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := s.parseJSONBody(r, &data); err != nil {
		return nil, fmt.Errorf("failed to parse request body: %w", err)
	}
	return data, nil
}

// ========== WebSocket 管理 ==========

// websocketHandler 处理 WebSocket 连接
func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// 生成连接 ID
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())

	// 注册连接
	s.connMutex.Lock()
	s.connections[connID] = conn
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, connID)
		s.connMutex.Unlock()
	}()

	// 发送欢迎消息
	welcomeMsg := map[string]interface{}{
		"type":      "connected",
		"connId":    connID,
		"timestamp": time.Now().Unix(),
	}
	msgBytes, _ := json.Marshal(welcomeMsg)
	conn.WriteMessage(websocket.TextMessage, msgBytes)

	// 消息循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 处理接收到的消息
		s.handleWebSocketMessage(connID, messageType, message)
	}
}

// handleWebSocketMessage 处理 WebSocket 消息
func (s *Server) handleWebSocketMessage(connID string, messageType int, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		fmt.Printf("Failed to parse WebSocket message: %v\n", err)
		return
	}

	// 根据消息类型处理
	switch msg["type"] {
	case "ping":
		s.sendToConnection(connID, map[string]interface{}{
			"type": "pong",
			"time": time.Now().Unix(),
		})
	case "subscribe":
		// 处理订阅请求
		s.handleSubscribe(connID, msg)
	default:
		fmt.Printf("Unknown message type: %v\n", msg["type"])
	}
}

// sendToConnection 发送消息到指定连接
func (s *Server) sendToConnection(connID string, data interface{}) error {
	s.connMutex.RLock()
	conn, exists := s.connections[connID]
	s.connMutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// Broadcast 广播消息到所有连接
func (s *Server) Broadcast(data interface{}) error {
	msgBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for _, conn := range s.connections {
		conn.WriteMessage(websocket.TextMessage, msgBytes)
	}

	return nil
}

// handleSubscribe 处理订阅请求
func (s *Server) handleSubscribe(connID string, msg map[string]interface{}) {
	channels, ok := msg["channels"].([]interface{})
	if !ok {
		return
	}

	// 发送订阅确认
	s.sendToConnection(connID, map[string]interface{}{
		"type":     "subscribed",
		"channels": channels,
		"time":     time.Now().Unix(),
	})
}
