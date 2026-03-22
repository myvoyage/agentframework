// Gateway WebSocket Handler - OpenClaw Protocol over WebSocket
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketHandler manages WebSocket connections and the control plane
type WebSocketHandler struct {
	svc       *Service
	config    *Config
	conns     map[string]*wsConnection
	connMutex sync.RWMutex
	broadcast chan *broadcastMsg
}

type wsConnection struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Service  *Service
	StopChan chan struct{}
	Seq      int64
	mu       sync.Mutex
}

type broadcastMsg struct {
	connID string
	data   []byte
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(svc *Service, cfg *Config) *WebSocketHandler {
	h := &WebSocketHandler{
		svc:       svc,
		config:    cfg,
		conns:     make(map[string]*wsConnection),
		broadcast: make(chan *broadcastMsg, 100),
	}
	go h.broadcastLoop()
	return h
}

// ServeHTTP handles WebSocket upgrade requests on root path
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Gateway] WebSocket upgrade failed: %v", err)
		return
	}

	wsConn := &wsConnection{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Service:  h.svc,
		StopChan: make(chan struct{}),
	}

	h.connMutex.Lock()
	h.conns["pending"] = wsConn
	h.connMutex.Unlock()

	go h.writePump(wsConn)
	go h.readPump(wsConn)
}

func (h *WebSocketHandler) broadcastLoop() {
	for msg := range h.broadcast {
		h.connMutex.RLock()
		for id, conn := range h.conns {
			if msg.connID == "" || msg.connID == id {
				select {
				case conn.Send <- msg.data:
				default:
				}
			}
		}
		h.connMutex.RUnlock()
	}
}

// ============================================================
// Read Pump
// ============================================================

func (h *WebSocketHandler) readPump(wsConn *wsConnection) {
	defer func() {
		h.removeConnection(wsConn)
		wsConn.Conn.Close()
	}()

	wsConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wsConn.Conn.SetPongHandler(func(string) error {
		wsConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, rawMsg, err := wsConn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Gateway] WS read error: %v", err)
			}
			break
		}
		if err := h.handleMessage(wsConn, rawMsg); err != nil {
			log.Printf("[Gateway] Message handle error: %v", err)
		}
	}
}

func (h *WebSocketHandler) handleMessage(wsConn *wsConnection, rawMsg []byte) error {
	var frame Frame
	if err := json.Unmarshal(rawMsg, &frame); err != nil {
		return h.sendError(wsConn, "", ErrCodeInvalidReq, "invalid JSON frame")
	}

	if frame.Method != "connect" && h.getConnectionID(wsConn) == "" {
		return h.sendError(wsConn, frame.ID, ErrCodeUnauthorized, "must connect first")
	}

	switch frame.Method {
	case "connect":
		return h.handleConnect(wsConn, &frame)
	case "health":
		return h.handleHealth(wsConn, &frame)
	case "status":
		return h.handleStatus(wsConn, &frame)
	case "system-presence":
		return h.handleSystemPresence(wsConn, &frame)
	case "agent":
		return h.handleAgentWS(wsConn, &frame)
	case "send":
		return h.handleSendWS(wsConn, &frame)
	case "node.list":
		return h.handleNodeListWS(wsConn, &frame)
	default:
		return h.sendError(wsConn, frame.ID, ErrCodeInvalidReq, fmt.Sprintf("unknown method: %s", frame.Method))
	}
}

// ============================================================
// Handlers
// ============================================================

func (h *WebSocketHandler) handleConnect(wsConn *wsConnection, frame *Frame) error {
	params := &ConnectParams{}
	if frame.Params != nil {
		data, _ := json.Marshal(frame.Params)
		json.Unmarshal(data, params)
	}

	hello, err := h.svc.HandleConnect(context.Background(), params)
	if err != nil {
		return h.sendError(wsConn, frame.ID, ErrCodeUnauthorized, err.Error())
	}

	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	wsConn.ID = connID
	h.connMutex.Lock()
	delete(h.conns, "pending")
	h.conns[connID] = wsConn
	h.connMutex.Unlock()

	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: hello}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) handleHealth(wsConn *wsConnection, frame *Frame) error {
	health := h.svc.HandleHealth(context.Background())
	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: health}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) handleStatus(wsConn *wsConnection, frame *Frame) error {
	status := h.svc.HandleStatus()
	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: status}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) handleSystemPresence(wsConn *wsConnection, frame *Frame) error {
	presence := h.svc.HandleSystemPresence()
	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: presence}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) handleAgentWS(wsConn *wsConnection, frame *Frame) error {
	params := &AgentRunParams{}
	if frame.Params != nil {
		data, _ := json.Marshal(frame.Params)
		json.Unmarshal(data, params)
	}

	result, eventCh, err := h.svc.HandleAgent(context.Background(), params)
	if err != nil {
		return h.sendError(wsConn, frame.ID, ErrCodeInternal, err.Error())
	}

	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: result}
	if err := h.sendFrame(wsConn, &res); err != nil {
		return err
	}

	go func() {
		for evtFrame := range eventCh {
			wsConn.mu.Lock()
			wsConn.Seq++
			seq := wsConn.Seq
			wsConn.mu.Unlock()

			// Inject runID into the agent event payload
			if ae, ok := evtFrame.Payload.(*AgentEvent); ok {
				ae.RunID = result.RunID
				evtFrame.Seq = seq
			}
			h.sendFrame(wsConn, evtFrame)
		}
		final := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: &AgentRunResult{RunID: result.RunID, Status: "ok"}}
		h.sendFrame(wsConn, &final)
	}()

	return nil
}

func (h *WebSocketHandler) handleSendWS(wsConn *wsConnection, frame *Frame) error {
	params := &SendParams{}
	if frame.Params != nil {
		data, _ := json.Marshal(frame.Params)
		json.Unmarshal(data, params)
	}

	result, err := h.svc.HandleSend(context.Background(), params)
	if err != nil {
		return h.sendError(wsConn, frame.ID, ErrCodeInternal, err.Error())
	}
	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: result}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) handleNodeListWS(wsConn *wsConnection, frame *Frame) error {
	result := h.svc.HandleNodeList()
	res := Frame{Type: FrameTypeRes, ID: frame.ID, OK: true, Payload: result}
	return h.sendFrame(wsConn, &res)
}

// ============================================================
// Write Pump
// ============================================================

func (h *WebSocketHandler) writePump(wsConn *wsConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		wsConn.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-wsConn.Send:
			wsConn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				wsConn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := wsConn.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			wsConn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := wsConn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-wsConn.StopChan:
			return
		}
	}
}

// ============================================================
// Helpers
// ============================================================

func (h *WebSocketHandler) sendFrame(wsConn *wsConnection, frame *Frame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	select {
	case wsConn.Send <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

func (h *WebSocketHandler) sendError(wsConn *wsConnection, id, code, msg string) error {
	res := Frame{
		Type:  FrameTypeRes,
		ID:    id,
		OK:    false,
		Error: &FrameError{Code: code, Message: msg},
	}
	return h.sendFrame(wsConn, &res)
}

func (h *WebSocketHandler) getConnectionID(wsConn *wsConnection) string {
	h.connMutex.RLock()
	defer h.connMutex.RUnlock()
	for id, c := range h.conns {
		if c == wsConn {
			return id
		}
	}
	return ""
}

func (h *WebSocketHandler) removeConnection(wsConn *wsConnection) {
	h.connMutex.Lock()
	defer h.connMutex.Unlock()
	for id, c := range h.conns {
		if c == wsConn {
			delete(h.conns, id)
			wsConn.Service.RemoveConnection(id)
			break
		}
	}
	close(wsConn.StopChan)
}

// BroadcastPresence broadcasts presence update
func (h *WebSocketHandler) BroadcastPresence() {
	presence := h.svc.HandleSystemPresence()
	frame := Frame{Type: FrameTypeEvent, Event: "presence", Payload: presence}
	data, _ := json.Marshal(frame)
	h.broadcast <- &broadcastMsg{data: data}
}
