// Gateway Protocol - OpenClaw Gateway Protocol definitions
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"time"
)

// ============================================================
// Frame Types (OpenClaw Protocol)
// ============================================================

// FrameType defines the type of protocol frame
type FrameType string

const (
	FrameTypeReq    FrameType = "req"    // Request
	FrameTypeRes    FrameType = "res"    // Response
	FrameTypeEvent  FrameType = "event"  // Event (pushed from server)
)

// Frame represents a protocol frame
type Frame struct {
	Type    FrameType     `json:"type"`              // "req", "res", "event"
	ID      string        `json:"id,omitempty"`      // Frame ID for req/res correlation
	Method  string        `json:"method,omitempty"`  // Method name (for req)
	Params  interface{}   `json:"params,omitempty"`  // Parameters (for req)
	OK      bool          `json:"ok,omitempty"`      // Success flag (for res)
	Payload interface{}   `json:"payload,omitempty"` // Response/event payload
	Error   *FrameError   `json:"error,omitempty"`   // Error object (for res)
	Event   string        `json:"event,omitempty"`   // Event name (for event)
	Seq     int64         `json:"seq,omitempty"`     // Sequence number (for event)
	StateVersion int64    `json:"stateVersion,omitempty"` // State version (for event)
}

// FrameError represents an error in a response frame
type FrameError struct {
	Code         string      `json:"code"`
	Message      string      `json:"message"`
	Details      interface{} `json:"details,omitempty"`
	Retryable    bool        `json:"retryable,omitempty"`
	RetryAfterMs int64       `json:"retryAfterMs,omitempty"`
}

// Error implements the error interface
func (e *FrameError) Error() string {
	return e.Message
}

// Error codes
const (
	ErrCodeNotLinked     = "NOT_LINKED"      // WhatsApp not authenticated
	ErrCodeAgentTimeout  = "AGENT_TIMEOUT"   // Agent didn't respond in time
	ErrCodeInvalidReq   = "INVALID_REQUEST" // Schema/parameter validation failed
	ErrCodeUnavailable  = "UNAVAILABLE"     // Gateway shutting down or dependency unavailable
	ErrCodeUnauthorized = "UNAUTHORIZED"    // Authentication failed
	ErrCodeNotFound     = "NOT_FOUND"       // Resource not found
	ErrCodeInternal     = "INTERNAL_ERROR"  // Internal error
)

// ============================================================
// Connect (Handshake)
// ============================================================

// ConnectParams parameters for connect method
type ConnectParams struct {
	MinProtocol int          `json:"minProtocol"`
	MaxProtocol int          `json:"maxProtocol"`
	Client      ClientInfo   `json:"client"`
	Caps        ClientCaps   `json:"caps"`
	Auth        *AuthInfo    `json:"auth,omitempty"`
	Locale      string       `json:"locale,omitempty"`
	UserAgent   string       `json:"userAgent,omitempty"`
}

// ClientInfo information about the connecting client
type ClientInfo struct {
	ID             string `json:"id"`
	DisplayName    string `json:"displayName,omitempty"`
	Version        string `json:"version,omitempty"`
	Platform       string `json:"platform,omitempty"`
	DeviceFamily   string `json:"deviceFamily,omitempty"`
	ModelIdentifier string `json:"modelIdentifier,omitempty"`
	Mode           string `json:"mode,omitempty"`
	InstanceID     string `json:"instanceId,omitempty"`
}

// ClientCaps client capabilities
type ClientCaps struct {
	Streaming bool `json:"streaming,omitempty"` // Supports streaming responses
	A2A       bool `json:"a2a,omitempty"`       // Agent-to-Agent protocol
}

// AuthInfo authentication information
type AuthInfo struct {
	Token   string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
}

// ============================================================
// Hello Response (after successful connect)
// ============================================================

// HelloOK successful connect response payload
type HelloOK struct {
	ProtocolVersion int           `json:"protocolVersion"`
	GatewayVersion  string        `json:"gatewayVersion"`
	UptimeMs        int64         `json:"uptimeMs"`
	StateVersion    int64         `json:"stateVersion"`
	Presence        []Presence    `json:"presence,omitempty"`
	Health          *HealthStatus `json:"health,omitempty"`
	Policy          *Policy       `json:"policy,omitempty"`
}

// Policy gateway policy settings
type Policy struct {
	MaxPayload        int64 `json:"maxPayload"`
	MaxBufferedBytes int64 `json:"maxBufferedBytes"`
	TickIntervalMs    int64 `json:"tickIntervalMs"`
}

// Presence represents a connected client presence entry
type Presence struct {
	Host             string   `json:"host"`
	IP               string   `json:"ip,omitempty"`
	Version          string   `json:"version,omitempty"`
	Platform         string   `json:"platform,omitempty"`
	DeviceFamily     string   `json:"deviceFamily,omitempty"`
	ModelIdentifier  string   `json:"modelIdentifier,omitempty"`
	Mode             string   `json:"mode,omitempty"`
	LastInputSeconds *int64   `json:"lastInputSeconds,omitempty"`
	Ts               int64    `json:"ts"`
	Reason           string   `json:"reason,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	InstanceID       string   `json:"instanceId,omitempty"`
}

// HealthStatus gateway health status
type HealthStatus struct {
	Status       string                 `json:"status"`                  // "ok", "degraded", "error"
	Version      string                 `json:"version"`
	UptimeMs     int64                  `json:"uptimeMs"`
	Timestamp    int64                  `json:"timestamp"`
	Checks       map[string]HealthCheck `json:"checks,omitempty"`
	Channels     []ChannelStatus        `json:"channels,omitempty"`
	Agents       []AgentStatus          `json:"agents,omitempty"`
	MemoryUsageMB float64               `json:"memoryUsageMB,omitempty"`
	CPUPercent   float64                `json:"cpuPercent,omitempty"`
}

// HealthCheck individual health check
type HealthCheck struct {
	Status  string `json:"status"` // "ok", "error", "warn"
	Message string `json:"message,omitempty"`
}

// ChannelStatus status of a messaging channel
type ChannelStatus struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Connected bool   `json:"connected"`
	Healthy   bool   `json:"healthy"`
}

// AgentStatus status of an agent
type AgentStatus struct {
	Name    string `json:"name"`
	Model   string `json:"model,omitempty"`
	Running bool   `json:"running"`
}

// ============================================================
// Methods
// ============================================================

// AgentRunParams parameters for agent method
type AgentRunParams struct {
	Agent   string                 `json:"agent"` // Agent name
	Input   string                 `json:"input"`
	Thread  string                 `json:"thread,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Stream  bool                   `json:"stream,omitempty"`
	Timeout int64                  `json:"timeout,omitempty"` // milliseconds
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// AgentRunResult result of agent run
type AgentRunResult struct {
	RunID   string `json:"runId"`
	Status  string `json:"status"` // "accepted", "ok", "error"
	Summary string `json:"summary,omitempty"`
}

// SendParams parameters for send method
type SendParams struct {
	Channel string `json:"channel"`
	Target  string `json:"target,omitempty"`
	Content string `json:"content"`
}

// SendResult result of send
type SendResult struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
}

// NodeListResult result of node.list
type NodeListResult struct {
	Nodes []NodeInfo `json:"nodes"`
}

// NodeInfo information about a node
type NodeInfo struct {
	ID             string            `json:"id"`
	Host           string            `json:"host"`
	Caps           map[string]bool   `json:"caps,omitempty"`
	DeviceFamily   string            `json:"deviceFamily,omitempty"`
	ModelIdentifier string           `json:"modelIdentifier,omitempty"`
	Paired         bool              `json:"paired"`
	Connected      bool              `json:"connected"`
	Commands       []string          `json:"commands,omitempty"`
}

// ============================================================
// Events
// ============================================================

// AgentEvent represents an agent streaming event
type AgentEvent struct {
	RunID     string      `json:"runId"`
	Seq       int64       `json:"seq"`
	Type      string      `json:"type"` // "tool_call", "tool_result", "output", "error"
	Name      string      `json:"name,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	Output    interface{} `json:"output,omitempty"`
	Content   string      `json:"content,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// TickEvent periodic keepalive event
type TickEvent struct {
	Ts int64 `json:"ts"`
}

// ShutdownEvent gateway shutdown event
type ShutdownEvent struct {
	Reason          string `json:"reason"`
	RestartExpectedMs int64 `json:"restartExpectedMs,omitempty"`
}

// ============================================================
// Connection State
// ============================================================

// Connection represents a single WebSocket client connection
type Connection struct {
	ID          string
	ClientInfo  ClientInfo
	Caps        ClientCaps
	Auth        *AuthInfo
	ConnectedAt time.Time
	LastInput   time.Time
}

// ConnectionManager manages all active connections
type ConnectionManager struct {
	connections map[string]*Connection
	seqCounters map[string]int64
}
