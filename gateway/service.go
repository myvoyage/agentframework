// Gateway Service - Core gateway service integrating with AgentFramework Host
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"AgentFramework/agent"
	"AgentFramework/agent/messaging"
	"AgentFramework/internal/auth"
)

// Service is the core gateway service
type Service struct {
	cfg       *Config
	host      *agent.Host
	startTime time.Time

	// Connection management
	connMgr *ConnectionManager
	_       sync.RWMutex // connMutex placeholder

	// Sequence counters
	stateVersion int64

	// Tick channel
	tickStop chan struct{}

	// Channel manager integration
	channelMgr *messaging.ChannelManager

	// Auth validator
	authValidator *auth.JWTValidator

	// Health checks
	healthChecks map[string]func() HealthCheck
}

// NewService creates a new gateway service
func NewService(cfg *Config, host *agent.Host) *Service {
	s := &Service{
		cfg:       cfg,
		host:      host,
		startTime: time.Now(),
		connMgr: &ConnectionManager{
			connections: make(map[string]*Connection),
			seqCounters: make(map[string]int64),
		},
		tickStop:     make(chan struct{}),
		healthChecks: make(map[string]func() HealthCheck),
	}

	// Set up auth validator if token is configured
	if cfg.Auth.Token != "" {
		s.authValidator = auth.NewJWTValidator(cfg.Auth.Token, "HS256")
	}

	// Set up channel manager reference
	if host != nil && host.ChannelManager() != nil {
		s.channelMgr = host.ChannelManager()
	}

	// Register default health checks
	s.registerHealthChecks()

	return s
}

// Start starts the gateway service
func (s *Service) Start(ctx context.Context) error {
	// Start tick goroutine
	go s.runTickLoop()

	log.Printf("[Gateway] Service started on %s:%d", s.cfg.Gateway.Host, s.cfg.Gateway.Port)
	return nil
}

// Stop stops the gateway service
func (s *Service) Stop(ctx context.Context) error {
	close(s.tickStop)

	s.connMgr.connections = make(map[string]*Connection)
	s.connMgr.seqCounters = make(map[string]int64)

	log.Printf("[Gateway] Service stopped")
	return nil
}

// ============================================================
// Connection Management
// ============================================================

// AddConnection registers a new connection
func (s *Service) AddConnection(conn *Connection) {
	s.connMgr.connections[conn.ID] = conn
	s.connMgr.seqCounters[conn.ID] = 0
}

// RemoveConnection removes a connection
func (s *Service) RemoveConnection(connID string) {
	delete(s.connMgr.connections, connID)
	delete(s.connMgr.seqCounters, connID)
}

// NextSeq returns the next sequence number
func (s *Service) NextSeq(connID string) int64 {
	return atomic.AddInt64(&s.stateVersion, 1)
}

// ============================================================
// Authentication
// ============================================================

// ValidateAuth validates authentication for a connection
func (s *Service) ValidateAuth(authInfo *AuthInfo) error {
	// No auth required if neither token nor password is configured
	if s.cfg.Auth.Token == "" && s.cfg.Auth.Password == "" {
		return nil
	}

	if authInfo == nil {
		return fmt.Errorf("authentication required")
	}

	// Token auth
	if authInfo.Token != "" && s.cfg.Auth.Token != "" {
		if authInfo.Token == s.cfg.Auth.Token {
			return nil
		}
	}

	// Password auth
	if authInfo.Password != "" && s.cfg.Auth.Password != "" {
		if authInfo.Password == s.cfg.Auth.Password {
			return nil
		}
	}

	return fmt.Errorf("invalid credentials")
}

// ============================================================
// Methods
// ============================================================

// HandleConnect processes a connect request
func (s *Service) HandleConnect(ctx context.Context, params *ConnectParams) (*HelloOK, error) {
	// Validate auth
	if err := s.ValidateAuth(params.Auth); err != nil {
		return nil, &FrameError{
			Code:    ErrCodeUnauthorized,
			Message: err.Error(),
		}
	}

	// Generate connection ID
	connID := uuid.New().String()

	// Register connection
	conn := &Connection{
		ID:          connID,
		ClientInfo:  params.Client,
		Caps:        params.Caps,
		Auth:        params.Auth,
		ConnectedAt: time.Now(),
		LastInput:   time.Now(),
	}
	s.AddConnection(conn)

	// Build presence list
	presence := s.buildPresenceList()

	// Build hello response
	hello := &HelloOK{
		ProtocolVersion: 1,
		GatewayVersion:  "1.0.0",
		UptimeMs:        time.Since(s.startTime).Milliseconds(),
		StateVersion:    atomic.LoadInt64(&s.stateVersion),
		Presence:        presence,
		Health:          s.BuildHealth(ctx),
		Policy:          s.cfg.Policy(),
	}

	return hello, nil
}

// HandleHealth processes a health request
func (s *Service) HandleHealth(ctx context.Context) *HealthStatus {
	return s.BuildHealth(ctx)
}

// HandleStatus processes a status request
func (s *Service) HandleStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":       "ok",
		"uptimeMs":      time.Since(s.startTime).Milliseconds(),
		"connections":   len(s.connMgr.connections),
		"stateVersion": atomic.LoadInt64(&s.stateVersion),
		"version":      "1.0.0",
	}
}

// HandleSystemPresence returns current presence list
func (s *Service) HandleSystemPresence() []Presence {
	return s.buildPresenceList()
}

// HandleAgent runs an agent and returns the result
func (s *Service) HandleAgent(ctx context.Context, params *AgentRunParams) (*AgentRunResult, <-chan *Frame, error) {
	runID := uuid.New().String()

	if s.host == nil {
		return nil, nil, fmt.Errorf("host not configured")
	}

	ag, ok := s.host.Agent(params.Agent)
	if !ok {
		return nil, nil, &FrameError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("agent %q not found", params.Agent),
		}
	}

	// Set up timeout
	timeout := s.cfg.AgentTimeout()
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// Create event channel
	eventCh := make(chan *Frame, 100)

	go func() {
		defer cancel()
		defer close(eventCh)

		// Run agent
		resp, err := ag.Run(ctx, params.Input)
		if err != nil {
			eventCh <- &Frame{
				Type:  FrameTypeEvent,
				Event: "agent",
				Payload: &AgentEvent{
					RunID: runID,
					Type:  "error",
					Error: err.Error(),
				},
			}
			return
		}

		// Send success result
		eventCh <- &Frame{
			Type:  FrameTypeEvent,
			Event: "agent",
			Payload: &AgentEvent{
				RunID:   runID,
				Type:    "output",
				Content: resp.Content,
				Output:  resp.Content,
			},
		}
	}()

	result := &AgentRunResult{
		RunID:  runID,
		Status: "accepted",
	}

	return result, eventCh, nil
}

// HandleSend sends a message through a channel
func (s *Service) HandleSend(ctx context.Context, params *SendParams) (*SendResult, error) {
	if s.channelMgr == nil {
		return nil, &FrameError{
			Code:    ErrCodeUnavailable,
			Message: "channel manager not available",
		}
	}

	ch, err := s.channelMgr.GetChannel(params.Channel)
	if err != nil {
		return nil, &FrameError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("channel %q not found", params.Channel),
		}
	}

	msg := &messaging.ChannelMessage{
		Content: params.Content,
		Target:  params.Target,
	}

	if err := ch.Publish(ctx, msg); err != nil {
		return nil, &FrameError{
			Code:    ErrCodeInternal,
			Message: fmt.Sprintf("send failed: %v", err),
		}
	}

	return &SendResult{
		MessageID: msg.ID,
		Status:    "sent",
	}, nil
}

// HandleNodeList lists connected nodes
func (s *Service) HandleNodeList() *NodeListResult {
	nodes := make([]NodeInfo, 0)

	for _, conn := range s.connMgr.connections {
		nodes = append(nodes, NodeInfo{
			ID:              conn.ID,
			Host:            s.cfg.Gateway.Host,
			Connected:       true,
			DeviceFamily:    conn.ClientInfo.DeviceFamily,
			ModelIdentifier: conn.ClientInfo.ModelIdentifier,
			Commands:        []string{"canvas.*", "browser.*"},
		})
	}

	return &NodeListResult{Nodes: nodes}
}

// ============================================================
// Health & Presence
// ============================================================

func (s *Service) registerHealthChecks() {
	s.healthChecks["host"] = func() HealthCheck {
		if s.host != nil {
			return HealthCheck{Status: "ok", Message: "host running"}
		}
		return HealthCheck{Status: "warn", Message: "host not configured"}
	}

	s.healthChecks["channels"] = func() HealthCheck {
		if s.channelMgr == nil {
			return HealthCheck{Status: "warn", Message: "channel manager not configured"}
		}
		return HealthCheck{Status: "ok", Message: "channel manager running"}
	}

	s.healthChecks["memory"] = func() HealthCheck {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.HeapAlloc > 500*1024*1024 {
			return HealthCheck{Status: "warn", Message: fmt.Sprintf("high memory: %.1fMB", float64(m.HeapAlloc)/1024/1024)}
		}
		return HealthCheck{Status: "ok", Message: fmt.Sprintf("memory: %.1fMB", float64(m.HeapAlloc)/1024/1024)}
	}
}

func (s *Service) BuildHealth(ctx context.Context) *HealthStatus {
	checks := make(map[string]HealthCheck)
	for name, check := range s.healthChecks {
		checks[name] = check()
	}

	var memMB float64
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memMB = float64(memStats.HeapAlloc) / 1024 / 1024

	channels := []ChannelStatus{}
	if s.channelMgr != nil {
		for _, ch := range s.channelMgr.ListChannels() {
			channels = append(channels, ChannelStatus{
				Name:      ch.Name(),
				Type:      string(ch.Type()),
				Connected: ch.IsRunning(),
				Healthy:   ch.GetStats().IsHealthy,
			})
		}
	}

	agents := []AgentStatus{}
	if s.host != nil {
		for _, name := range s.host.ListAgents() {
			agents = append(agents, AgentStatus{Name: name, Running: true})
		}
	}

	overallStatus := "ok"
	for _, c := range checks {
		if c.Status == "error" {
			overallStatus = "error"
			break
		}
		if c.Status == "warn" {
			overallStatus = "degraded"
		}
	}

	return &HealthStatus{
		Status:        overallStatus,
		Version:       "1.0.0",
		UptimeMs:      time.Since(s.startTime).Milliseconds(),
		Timestamp:     time.Now().Unix(),
		Checks:        checks,
		Channels:      channels,
		Agents:        agents,
		MemoryUsageMB: memMB,
	}
}

func (s *Service) buildPresenceList() []Presence {
	presence := make([]Presence, 0)
	for _, conn := range s.connMgr.connections {
		presence = append(presence, Presence{
			Host:             s.cfg.Gateway.Host,
			Version:          conn.ClientInfo.Version,
			Platform:         conn.ClientInfo.Platform,
			DeviceFamily:     conn.ClientInfo.DeviceFamily,
			ModelIdentifier:  conn.ClientInfo.ModelIdentifier,
			Mode:             conn.ClientInfo.Mode,
			Ts:               time.Now().Unix(),
			InstanceID:       conn.ClientInfo.InstanceID,
		})
	}
	return presence
}

// ============================================================
// Tick Loop
// ============================================================

func (s *Service) runTickLoop() {
	ticker := time.NewTicker(s.cfg.TickInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.tickStop:
			return
		case <-ticker.C:
			// Tick events broadcasted via WebSocket manager
		}
	}
}

// ============================================================
// OpenAI Compatible Methods
// ============================================================

// ChatCompletionRequest represents an OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Stream      bool           `json:"stream,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Agent       string         `json:"agent,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatCompletionResponse represents an OpenAI-compatible response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HandleChatCompletion handles OpenAI-compatible chat completions
func (s *Service) HandleChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	agentName := req.Agent
	if agentName == "" && s.host != nil && len(s.host.ListAgents()) > 0 {
		agentName = s.host.ListAgents()[0]
	}
	if agentName == "" {
		return nil, fmt.Errorf("no agent specified and no default agent available")
	}

	ag, ok := s.host.Agent(agentName)
	if !ok {
		return nil, fmt.Errorf("agent %q not found", agentName)
	}

	// Build input from messages
	input := ""
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			input += fmt.Sprintf("[System] %s\n", msg.Content)
		case "user":
			input += fmt.Sprintf("[User] %s\n", msg.Content)
		case "assistant":
			input += fmt.Sprintf("[Assistant] %s\n", msg.Content)
		}
	}

	resp, err := ag.Run(ctx, input)
	if err != nil {
		return nil, err
	}

	choice := Choice{
		Index: 0,
		Message: ChatMessage{
			Role:    "assistant",
			Content: resp.Content,
		},
		FinishReason: "stop",
	}

	return &ChatCompletionResponse{
		ID:      uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{choice},
		Usage: Usage{
			PromptTokens:     estimateTokens(input),
			CompletionTokens: estimateTokens(resp.Content),
			TotalTokens:      estimateTokens(input) + estimateTokens(resp.Content),
		},
	}, nil
}

// HandleResponses handles OpenAI Responses API
func (s *Service) HandleResponses(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return s.HandleChatCompletion(ctx, req)
}

// ToolsInvokeResult represents the result of tools.invoke
type ToolsInvokeResult struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// HandleToolsInvoke handles tools.invoke method
func (s *Service) HandleToolsInvoke(ctx context.Context, toolName, method string, params map[string]interface{}) (*ToolsInvokeResult, error) {
	return nil, &FrameError{
		Code:    ErrCodeUnavailable,
		Message: "tools not implemented in this version",
	}
}

// ============================================================
// Helper
// ============================================================

func estimateTokens(text string) int {
	return len(text) / 4
}
