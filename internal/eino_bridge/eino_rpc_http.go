//go:build eino
// +build eino

package einobridge

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	pe "AgentFramework/internal/pipelineengine"
)

// Simple in-process HTTP RPC server to expose a minimal MCP-like API for stage 4.
// This is intentionally lightweight and meant for local testing of the patch chain.

type httpRPCServer struct {
	engine *pe.PipelineEngine
	srv    *http.Server
	mu     sync.Mutex
}

var globalHTTPServer *httpRPCServer

func StartHTTPBridge(port int, engine *pe.PipelineEngine) error {
	if engine == nil {
		return nil
	}
	r := mux.NewRouter()
	s := &httpRPCServer{engine: engine}
	r.HandleFunc("/invoke_tool", s.handleInvokeTool).Methods("POST")
	r.HandleFunc("/run_pipeline", s.handleRunPipeline).Methods("POST")
	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: r}
	s.srv = srv
	globalHTTPServer = s
	go func() {
		_ = srv.ListenAndServe()
	}()
	// give the server a moment to start
	time.Sleep(50 * time.Millisecond)
	return nil
}

func StopHTTPBridge() error {
	if globalHTTPServer == nil || globalHTTPServer.srv == nil {
		return nil
	}
	return globalHTTPServer.srv.Close()
}

func (s *httpRPCServer) handleInvokeTool(w http.ResponseWriter, r *http.Request) {
	var req pe.MCPInvokeToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := s.engine.InvokeTool(r.Context(), req.Tool, req.Params, req.Context)
	resp := pe.MCPInvokeToolResponse{Success: err == nil, Data: data}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *httpRPCServer) handleRunPipeline(w http.ResponseWriter, r *http.Request) {
	// Decode PipelineSpec JSON payload
	var payload struct {
		Pipeline *pe.PipelineSpec `json:"pipeline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	exec, err := s.engine.RunPipeline(ctx, payload.Pipeline)
	resp := struct {
		Success bool                 `json:"success"`
		Data    *pe.ExecutionContext `json:"data"`
		Error   string               `json:"error,omitempty"`
	}{Data: exec}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		resp.Success = true
	}
	_ = json.NewEncoder(w).Encode(resp)
}
