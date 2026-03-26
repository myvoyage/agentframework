// Router - Request routing for gateway to runtime
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RouteType defines the type of routing
type RouteType string

const (
	RouteTypeRoundRobin RouteType = "round_robin"
	RouteTypeLeastConn   RouteType = "least_conn"
	RouteTypeRandom      RouteType = "random"
	RouteTypeSticky      RouteType = "sticky"
)

// RouteConfig holds routing configuration
type RouteConfig struct {
	Type       RouteType
	MaxRetries int
	Timeout    time.Duration
}

// DefaultRouteConfig returns default routing configuration
func DefaultRouteConfig() *RouteConfig {
	return &RouteConfig{
		Type:       RouteTypeLeastConn,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}
}

// Router handles routing of requests to agent instances
type Router struct {
	config    *RouteConfig
	runtime   *RuntimeManager
	stickyMap map[string]string // session ID -> instance ID
	mu        sync.RWMutex
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		config:    DefaultRouteConfig(),
		stickyMap: make(map[string]string),
	}
}

// SetRuntime sets the runtime manager
func (r *Router) SetRuntime(rt *RuntimeManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runtime = rt
}

// SetConfig sets the route configuration
func (r *Router) SetConfig(cfg *RouteConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cfg != nil {
		r.config = cfg
	}
}

// RouteRequest routes a request to an appropriate instance
func (r *Router) RouteRequest(ctx context.Context, req *RouteRequest) (*RouteResponse, error) {
	if r.runtime == nil {
		return nil, fmt.Errorf("runtime manager not set")
	}
	
	// Get or create instance
	var inst *AgentInstance
	var err error
	
	switch r.config.Type {
	case RouteTypeSticky:
		inst, err = r.routeSticky(ctx, req)
	case RouteTypeRoundRobin:
		inst, err = r.routeRoundRobin(ctx, req)
	case RouteTypeLeastConn:
		inst, err = r.routeLeastConn(ctx, req)
	case RouteTypeRandom:
		fallthrough
	default:
		inst, err = r.routeLeastConn(ctx, req)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Execute request
	return r.executeRequest(ctx, inst, req)
}

// routeSticky routes to the same instance for a session
func (r *Router) routeSticky(ctx context.Context, req *RouteRequest) (*AgentInstance, error) {
	r.mu.Lock()
	
	// Check if session has a mapped instance
	if req.SessionID != "" {
		if instID, ok := r.stickyMap[req.SessionID]; ok {
			r.mu.Unlock()
			inst, ok := r.runtime.GetInstance(instID)
			if ok && (inst.Status == StatusRunning || inst.Status == StatusIdle) {
				return inst, nil
			}
		}
	}
	r.mu.Unlock()
	
	// Create or get new instance
	inst, err := r.runtime.GetOrCreateInstance(ctx, req.AgentName, req.AgentType)
	if err != nil {
		return nil, err
	}
	
	// Map session to instance
	r.mu.Lock()
	if req.SessionID != "" {
		r.stickyMap[req.SessionID] = inst.ID
	}
	r.mu.Unlock()
	
	return inst, nil
}

// routeRoundRobin routes to instances in round-robin fashion
func (r *Router) routeRoundRobin(ctx context.Context, req *RouteRequest) (*AgentInstance, error) {
	instances := r.runtime.ListInstancesByAgent(req.AgentName)
	
	if len(instances) == 0 {
		return r.runtime.GetOrCreateInstance(ctx, req.AgentName, req.AgentType)
	}
	
	// Simple round-robin (can be improved with atomic counter)
	// For now, just pick first running instance
	for _, inst := range instances {
		if inst.Status == StatusRunning || inst.Status == StatusIdle {
			return inst, nil
		}
	}
	
	return r.runtime.GetOrCreateInstance(ctx, req.AgentName, req.AgentType)
}

// routeLeastConn routes to instance with least connections
func (r *Router) routeLeastConn(ctx context.Context, req *RouteRequest) (*AgentInstance, error) {
	instances := r.runtime.ListInstancesByAgent(req.AgentName)
	
	if len(instances) == 0 {
		return r.runtime.GetOrCreateInstance(ctx, req.AgentName, req.AgentType)
	}
	
	// Find instance with least active requests
	var selected *AgentInstance
	minReq := int64(-1)
	
	for _, inst := range instances {
		if inst.Status != StatusRunning && inst.Status != StatusIdle {
			continue
		}
		if minReq == -1 || inst.RequestCount < minReq {
			selected = inst
			minReq = inst.RequestCount
		}
	}
	
	if selected != nil {
		return selected, nil
	}
	
	return r.runtime.GetOrCreateInstance(ctx, req.AgentName, req.AgentType)
}

// executeRequest executes a request on an instance
func (r *Router) executeRequest(ctx context.Context, inst *AgentInstance, req *RouteRequest) (*RouteResponse, error) {
	// Add timeout if configured
	if r.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
	}
	
	requestID := uuid.New().String()
	startTime := time.Now()
	
	log.Printf("[Router] Routing request %s to instance %s (agent: %s)", 
		requestID, inst.ID, inst.AgentName)
	
	// Execute with retries
	for retry := 0; retry <= r.config.MaxRetries; retry++ {
		if retry > 0 {
			log.Printf("[Router] Retry %d for request %s", retry, requestID)
		}

		// Execute agent (placeholder - depends on your Agent interface)
		// result, err := inst.Agent.Run(ctx, req.Input)
		// if err == nil {
		//     return &RouteResponse{
		//         RequestID: requestID,
		//         InstanceID: inst.ID,
		//         Content: result.Content,
		//         Duration: time.Since(startTime),
		//         Retries: retry,
		//     }, nil
		// }

		// Simulate execution for now
		time.Sleep(100 * time.Millisecond)
		break
	}
	
	// Placeholder response
	return &RouteResponse{
		RequestID:   requestID,
		InstanceID:  inst.ID,
		Content:     "Response from " + inst.AgentName,
		Duration:    time.Since(startTime),
		Retries:     0,
	}, nil
}

// ClearSession clears sticky session mapping
func (r *Router) ClearSession(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stickyMap, sessionID)
}

// RouteRequest represents a request to be routed
type RouteRequest struct {
	SessionID string
	AgentName string
	AgentType string
	Input     string
	Headers   http.Header
	Metadata  map[string]interface{}
}

// RouteResponse represents a routed response
type RouteResponse struct {
	RequestID   string
	InstanceID  string
	Content     string
	Error       string
	Duration    time.Duration
	Retries     int
	Metadata    map[string]interface{}
}

// RouterStats holds routing statistics
type RouterStats struct {
	TotalRequests    int64
	SuccessfulRoutes int64
	FailedRoutes     int64
	AvgLatency       time.Duration
	ActiveSessions   int
}
