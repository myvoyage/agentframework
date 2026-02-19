// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package einobridge

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	pe "AgentFramework/internal/pipelineengine"
)

// RunHTTPBridge starts a minimal HTTP RPC bridge on the given port using the provided engine.
// This is a convenience entry for local testing of Stage 6/Stage 5 capabilities.
func RunHTTPBridge(port int, engine *pe.PipelineEngine) *http.Server {
	if engine == nil {
		return nil
	}
	r := mux.NewRouter()
	s := &httpServer{engine: engine}
	r.HandleFunc("/invoke_tool", s.handleInvokeTool).Methods("POST")
	r.HandleFunc("/run_pipeline", s.handleRunPipeline).Methods("POST")
	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: r}
	go func() {
		log.Printf("Stage6 HTTP bridge starting on port %d", port)
		_ = srv.ListenAndServe()
	}()
	return srv
}

type httpServer struct{ engine *pe.PipelineEngine }

func (s *httpServer) handleInvokeTool(w http.ResponseWriter, r *http.Request) {
	var req pe.MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := s.engine.InvokeTool(r.Context(), req.Tool, req.Params, req.Context)
	resp := pe.MCPResponse{Success: err == nil, Data: data}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *httpServer) handleRunPipeline(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Pipeline *pe.PipelineSpec `json:"pipeline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	exec, err := s.engine.RunPipeline(r.Context(), payload.Pipeline)
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
