// Gateway Bridge - Connects Gateway to Agent Runtime
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"log"
	"sync"
	"time"
)

// GatewayBridge connects the gateway layer to the agent runtime
// This follows OpenClaw's architecture pattern: Gateway -> Agent Runtime
type GatewayBridge struct {
	runtime *RuntimeManager
	router  *Router

	// Channel adapters for different platforms
	adapters map[string]ChannelAdapter
	mu       sync.RWMutex

	// Request queue for async processing
	requestQueue chan *GatewayRequest
	workerCount  int

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ChannelAdapter defines the interface for channel-specific adapters
type ChannelAdapter interface {
	// Name returns the channel name (e.g., "wechat", "telegram", "slack")
	Name() string

	// ParseRequest parses a channel-specific request to a generic format
	ParseRequest(data []byte) (*GatewayRequest, error)

	// FormatResponse formats a response for the channel
	FormatResponse(resp *GatewayResponse) ([]byte, error)

	// ValidateAuth validates channel-specific authentication
	ValidateAuth(data []byte) (string, error) // returns userID
}

// GatewayRequest represents an incoming request from the gateway
type GatewayRequest struct {
	ID         string
	Channel    string // "wechat", "telegram", "slack", "http", "websocket"
	UserID     string
	SessionID  string
	AgentName  string
	AgentType  string
	Input      string
	Metadata   map[string]interface{}
	Timestamp  time.Time

	// Response channel for async handling
	ResponseChan chan *GatewayResponse
}

// GatewayResponse represents a response to send back
type GatewayResponse struct {
	RequestID string
	Content   string
	Error     string
	Metadata  map[string]interface{}
	Duration  time.Duration
}

// NewGatewayBridge creates a new gateway bridge
func NewGatewayBridge(runtime *RuntimeManager) *GatewayBridge {
	ctx, cancel := context.WithCancel(context.Background())

	bridge := &GatewayBridge{
		runtime:      runtime,
		router:       NewRouter(),
		adapters:     make(map[string]ChannelAdapter),
		requestQueue: make(chan *GatewayRequest, 1000),
		workerCount:  10,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Link router to runtime
	bridge.router.SetRuntime(runtime)

	return bridge
}

// Start starts the gateway bridge
func (b *GatewayBridge) Start(ctx context.Context) error {
	log.Printf("[GatewayBridge] Starting with %d workers", b.workerCount)

	// Start worker goroutines
	for i := 0; i < b.workerCount; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}

	// Start runtime if not started
	if err := b.runtime.Start(); err != nil {
		return err
	}

	log.Printf("[GatewayBridge] Started successfully")
	return nil
}

// Stop stops the gateway bridge
func (b *GatewayBridge) Stop() {
	log.Printf("[GatewayBridge] Stopping...")

	b.cancel()
	b.wg.Wait()

	b.runtime.Stop()

	log.Printf("[GatewayBridge] Stopped")
}

// RegisterAdapter registers a channel adapter
func (b *GatewayBridge) RegisterAdapter(adapter ChannelAdapter) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.adapters[adapter.Name()] = adapter
	log.Printf("[GatewayBridge] Registered adapter: %s", adapter.Name())
}

// HandleRequest handles an incoming request (sync)
func (b *GatewayBridge) HandleRequest(ctx context.Context, req *GatewayRequest) (*GatewayResponse, error) {
	startTime := time.Now()

	// Validate request
	if req.Channel == "" {
		req.Channel = "http"
	}

	// Create response channel
	if req.ResponseChan == nil {
		req.ResponseChan = make(chan *GatewayResponse, 1)
	}

	// Queue request
	select {
	case b.requestQueue <- req:
	default:
		// Queue full, process synchronously
		return b.processRequest(ctx, req)
	}

	// Wait for response with timeout
	select {
	case resp := <-req.ResponseChan:
		resp.Duration = time.Since(startTime)
		return resp, nil
	case <-time.After(30 * time.Second):
		return &GatewayResponse{
			RequestID: req.ID,
			Error:     "request timeout",
			Duration:  time.Since(startTime),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// HandleRequestAsync handles an incoming request (async)
func (b *GatewayBridge) HandleRequestAsync(req *GatewayRequest) {
	req.Timestamp = time.Now()
	if req.ResponseChan == nil {
		req.ResponseChan = make(chan *GatewayResponse, 1)
	}

	select {
	case b.requestQueue <- req:
	default:
		// Queue full, send error response
		req.ResponseChan <- &GatewayResponse{
			RequestID: req.ID,
			Error:     "request queue full",
		}
	}
}

// worker processes requests from the queue
func (b *GatewayBridge) worker(id int) {
	defer b.wg.Done()

	log.Printf("[GatewayBridge] Worker %d started", id)

	for {
		select {
		case <-b.ctx.Done():
			log.Printf("[GatewayBridge] Worker %d stopped", id)
			return
		case req := <-b.requestQueue:
			resp, err := b.processRequest(b.ctx, req)
			if err != nil {
				resp = &GatewayResponse{
					RequestID: req.ID,
					Error:     err.Error(),
				}
			}
			if req.ResponseChan != nil {
				req.ResponseChan <- resp
			}
		}
	}
}

// processRequest processes a single request
func (b *GatewayBridge) processRequest(ctx context.Context, req *GatewayRequest) (*GatewayResponse, error) {
	// Get adapter for channel
	adapter := b.getAdapter(req.Channel)

	// Create route request
	routeReq := &RouteRequest{
		SessionID: req.SessionID,
		AgentName: req.AgentName,
		AgentType: req.AgentType,
		Input:     req.Input,
		Metadata:  req.Metadata,
	}

	// Default agent type if not specified
	if routeReq.AgentType == "" {
		routeReq.AgentType = "chat"
	}

	// Default agent name if not specified
	if routeReq.AgentName == "" {
		routeReq.AgentName = "default"
	}

	// Route request through router
	routeResp, err := b.router.RouteRequest(ctx, routeReq)
	if err != nil {
		return &GatewayResponse{
			RequestID: req.ID,
			Error:     err.Error(),
		}, nil
	}

	// Format response if adapter exists
	var content string
	if adapter != nil && routeResp.Content != "" {
		content = routeResp.Content
	} else {
		content = routeResp.Content
	}

	return &GatewayResponse{
		RequestID: req.ID,
		Content:   content,
		Metadata:  routeResp.Metadata,
	}, nil
}

// getAdapter gets the adapter for a channel
func (b *GatewayBridge) getAdapter(channel string) ChannelAdapter {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.adapters[channel]
}

// GetStats returns bridge statistics
func (b *GatewayBridge) GetStats() *BridgeStats {
	return &BridgeStats{
		RuntimeStats:     b.runtime.GetStats(),
		QueueSize:        len(b.requestQueue),
		QueueCapacity:    cap(b.requestQueue),
		ActiveWorkers:    b.workerCount,
		RegisteredAdapters: len(b.adapters),
	}
}

// Router returns the router instance
func (b *GatewayBridge) Router() *Router {
	return b.router
}

// SetWorkerCount sets the worker count for async processing
func (b *GatewayBridge) SetWorkerCount(count int) {
	b.workerCount = count
}

// BridgeStats holds bridge statistics
type BridgeStats struct {
	RuntimeStats       *RuntimeStats
	QueueSize          int
	QueueCapacity      int
	ActiveWorkers      int
	RegisteredAdapters int
}

// HealthCheck performs a health check
func (b *GatewayBridge) HealthCheck() *BridgeHealth {
	health := &BridgeHealth{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	// Check runtime
	if b.runtime == nil {
		health.Status = "unhealthy"
		health.Issues = append(health.Issues, "runtime not initialized")
	}

	// Check queue
	if len(b.requestQueue) >= cap(b.requestQueue) {
		health.Status = "degraded"
		health.Issues = append(health.Issues, "request queue full")
	}

	// Get health summary
	if b.runtime != nil && b.runtime.healthChecker != nil {
		health.HealthSummary = b.runtime.healthChecker.GetSummary()
	}

	return health
}

// BridgeHealth holds health check results
type BridgeHealth struct {
	Status       string
	Timestamp    time.Time
	Issues       []string
	HealthSummary *HealthSummary
}
