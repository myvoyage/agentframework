// Gateway HTTP Server - OpenAI-compatible endpoints
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTPServer handles HTTP endpoints (OpenAI compatible + gateway controls)
type HTTPServer struct {
	svc    *Service
	config *Config
	mux    *http.ServeMux
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(svc *Service, cfg *Config) *HTTPServer {
	h := &HTTPServer{svc: svc, config: cfg}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

func (h *HTTPServer) registerRoutes() {
	// OpenAI-compatible endpoints
	h.mux.HandleFunc("POST /v1/chat/completions", h.handleChatCompletions)
	h.mux.HandleFunc("POST /v1/responses", h.handleResponses)
	h.mux.HandleFunc("POST /tools/invoke", h.handleToolsInvoke)

	// Gateway control endpoints
	h.mux.HandleFunc("GET /health", h.handleHealthHTTP)
	h.mux.HandleFunc("GET /status", h.handleStatusHTTP)

	// CORS preflight
	h.mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		h.setCORS(w)
		w.WriteHeader(http.StatusOK)
	})
}

func (h *HTTPServer) setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, openclaw-auth")
}

// Handler returns the HTTP handler
func (h *HTTPServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.setCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.mux.ServeHTTP(w, r)
	})
}

// ============================================================
// OpenAI Compatible Endpoints
// ============================================================

func (h *HTTPServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var req ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	// Use default model if not specified
	if req.Model == "" {
		req.Model = "default"
	}

	ctx := r.Context()
	resp, err := h.svc.HandleChatCompletion(ctx, &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	elapsed := time.Since(start)
	if h.config.Gateway.Verbose {
		log.Printf("[Gateway] /v1/chat/completions completed in %v", elapsed)
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPServer) handleResponses(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var req ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	if req.Model == "" {
		req.Model = "default"
	}

	resp, err := h.svc.HandleResponses(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPServer) handleToolsInvoke(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	toolName, _ := req["tool"].(string)
	method, _ := req["method"].(string)
	params, _ := req["params"].(map[string]interface{})

	result, err := h.svc.HandleToolsInvoke(r.Context(), toolName, method, params)
	if err != nil {
		fe, ok := err.(*FrameError)
		if ok {
			h.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"error": map[string]string{
					"code":    fe.Code,
					"message": fe.Message,
				},
			})
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// ============================================================
// Control Endpoints
// ============================================================

func (h *HTTPServer) handleHealthHTTP(w http.ResponseWriter, r *http.Request) {
	health := h.svc.HandleHealth(r.Context())
	h.writeJSON(w, http.StatusOK, health)
}

func (h *HTTPServer) handleStatusHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.svc.HandleStatus()
	h.writeJSON(w, http.StatusOK, status)
}

// ============================================================
// Helpers
// ============================================================

func (h *HTTPServer) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *HTTPServer) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"type":   "invalid_request_error",
		},
	})
}

// OpenAIStreamResponse represents a streamed chunk
type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		Delta        Delta  `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

type Delta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}

// StreamChatCompletion streams a chat completion response
func (h *HTTPServer) StreamChatCompletion(w http.ResponseWriter, r *http.Request, req *ChatCompletionRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	agentName := req.Agent
	if agentName == "" && h.svc != nil {
		// Get first agent
		_ = agentName
	}

	// Generate streaming chunks
	id := uuid.New().String()
	chunk := OpenAIStreamResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   req.Model,
	}
	chunk.Choices = append(chunk.Choices, struct {
		Index        int    `json:"index"`
		Delta        Delta  `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{Index: 0, Delta: Delta{Role: "assistant"}})

	data, _ := json.Marshal(chunk)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()

	// Content chunks would be streamed from agent
	// For now, send done
	doneData, _ := json.Marshal(map[string]interface{}{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []map[string]interface{}{
			{"index": 0, "delta": map[string]string{}, "finish_reason": "stop"},
		},
	})
	fmt.Fprintf(w, "data: %s\n\n", string(doneData))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// CheckAuth checks Authorization header
func CheckAuth(r *http.Request, cfg *Config) bool {
	if cfg.Auth.Token == "" && cfg.Auth.Password == "" {
		return true
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == cfg.Auth.Token {
			return true
		}
	}

	// Also check openclaw-auth header
	openclawAuth := r.Header.Get("openclaw-auth")
	if openclawAuth == cfg.Auth.Token {
		return true
	}

	return false
}
