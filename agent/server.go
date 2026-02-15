// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	callbacksHelper "github.com/cloudwego/eino/utils/callbacks"
	"github.com/google/uuid"
)

// AgentRuntimeServer exposes the Host via HTTP.
type AgentRuntimeServer struct {
	host *Host
	mux  *http.ServeMux
	addr string
	// Checkpoint store for persistent workflows
	checkpointStore CheckpointStore
}

// RunRequest represents the standard request body for running an agent or workflow.
type RunRequest struct {
	Input       string                 `json:"input"`
	ThreadID    string                 `json:"thread_id,omitempty"`    // For stateful agents
	ModelConfig map[string]interface{} `json:"model_config,omitempty"` // Override model params like temperature
}

// RunResponse represents the standard response.
type RunResponse struct {
	Content  string `json:"content"`
	ThreadID string `json:"thread_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// NewAgentRuntimeServer creates a new server instance with configurable checkpoint path.
func NewAgentRuntimeServer(host *Host, addr string, checkpointPath string) *AgentRuntimeServer {
	// Initialize CheckpointStore (File based for now)
	var cpStore CheckpointStore
	var err error

	if checkpointPath != "" {
		cpStore, err = NewFileCheckpointStore(checkpointPath)
	} else {
		// Default path
		cpStore, err = NewFileCheckpointStore("./checkpoints")
	}

	if err != nil {
		log.Printf("Failed to create checkpoint store: %v, using memory store", err)
		// Fallback to memory store if file store creation fails
		cpStore = NewMemoryCheckpointStore()
	}

	s := &AgentRuntimeServer{
		host:            host,
		mux:             http.NewServeMux(),
		addr:            addr,
		checkpointStore: cpStore,
	}
	s.registerRoutes()
	return s
}

func (s *AgentRuntimeServer) registerRoutes() {
	// Agent endpoints
	s.mux.HandleFunc("POST /v1/agents/{name}/run", s.handleAgentRun)
	s.mux.HandleFunc("POST /v1/agents/{name}/stream", s.handleAgentStream)

	// Workflow endpoints
	s.mux.HandleFunc("POST /v1/workflows/{name}/run", s.handleWorkflowRun)
	s.mux.HandleFunc("POST /v1/workflows/{name}/resume", s.handleWorkflowResume)
	s.mux.HandleFunc("GET /v1/workflows/{name}", s.handleGetWorkflow)

	// Run History endpoints
	s.mux.HandleFunc("GET /v1/runs", s.handleListRuns)
	s.mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)

	// List endpoints
	s.mux.HandleFunc("GET /v1/agents", s.handleListAgents)

	// Health check
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func (s *AgentRuntimeServer) handleAgentStream(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	ag, ok := s.host.Agent(name)
	if !ok {
		http.Error(w, fmt.Sprintf("Agent '%s' not found", name), http.StatusNotFound)
		return
	}

	// Check if agent supports streaming
	streamable, ok := ag.(StreamableAgent)
	if !ok {
		http.Error(w, fmt.Sprintf("Agent '%s' does not support streaming", name), http.StatusBadRequest)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	// Create a channel to multiplex stream events and callback events
	eventCh := make(chan interface{}, 100)
	var wg sync.WaitGroup

	// Setup Callbacks to capture tool events
	handler := callbacksHelper.NewHandlerHelper().Tool(&callbacksHelper.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			// Try to capture arguments.
			// tool.CallbackInput fields might vary. We use Sprintf to capture whatever is available.
			args := ""
			if input != nil {
				args = fmt.Sprintf("%+v", input)
			}
			eventCh <- map[string]string{
				"type":  "tool_start",
				"name":  info.Name,
				"input": args,
			}
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			out := ""
			if output != nil {
				out = output.Response
			}
			eventCh <- map[string]string{
				"type":   "tool_end",
				"name":   info.Name,
				"output": out,
			}
			return ctx
		},
	}).Handler()

	// Inject callbacks into context
	ctx = callbacks.InitCallbacks(ctx, &callbacks.RunInfo{}, handler)

	// Parse model options
	var opts []model.Option
	if temp, ok := req.ModelConfig["temperature"].(float64); ok {
		opts = append(opts, model.WithTemperature(float32(temp)))
	}

	sr, err := streamable.Stream(ctx, req.Input, opts...)
	if err != nil {
		// Can't change status code now as headers might be written (implicit 200)
		// but standard is flush after write, so maybe we can send error event
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		return
	}
	// Do not defer sr.Close() here, do it in the goroutine or after loop

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sr.Close()
		for {
			chunk, err := sr.Recv()
			if err != nil {
				if err != io.EOF {
					eventCh <- err
				}
				// EOF means stream done
				return
			}
			eventCh <- chunk
		}
	}()

	// Close eventCh when stream is done
	go func() {
		wg.Wait()
		close(eventCh)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		// Should not happen in standard http server
		return
	}

	for event := range eventCh {
		switch e := event.(type) {
		case error:
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", e.Error())
			return
		case *schema.Message:
			// Send chunk content
			if e.Content != "" {
				payload := map[string]interface{}{
					"type":       "content_delta",
					"content":    e.Content,
					"agent_name": name,
					"thread_id":  req.ThreadID,
				}
				data, _ := json.Marshal(payload)
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
			}
		case map[string]string:
			// Tool event
			data, _ := json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	}

	// Send done event with final status
	donePayload, _ := json.Marshal(map[string]string{
		"type":   "status",
		"status": "completed",
	})
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", string(donePayload))
	flusher.Flush()
}

func (s *AgentRuntimeServer) Start() error {
	log.Printf("Agent Runtime Server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, s)
}

func (s *AgentRuntimeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS Middleware
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *AgentRuntimeServer) handleAgentRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	// Check if agent exists
	ag, ok := s.host.Agent(name)
	if !ok {
		http.Error(w, fmt.Sprintf("Agent '%s' not found", name), http.StatusNotFound)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Parse model options
	var opts []model.Option
	if temp, ok := req.ModelConfig["temperature"].(float64); ok {
		opts = append(opts, model.WithTemperature(float32(temp)))
	}

	// Run logic
	var resp *schema.Message
	var err error
	var threadID string

	// Try to use AgentService if available to support stateful threads
	svc := s.host.Service()
	if svc != nil {
		// Attempt to cast to ThreadAwareAgent
		if ta, ok := ag.(ThreadAwareAgent); ok {
			var thread *Thread
			thread, resp, err = svc.Send(ctx, ta, req.ThreadID, req.Input, opts...)
			if thread != nil {
				threadID = thread.ID
			}
		} else {
			// Stateless run via service (or just direct run)
			// Service.Send requires ThreadAwareAgent. If not thread aware, run directly.
			resp, err = ag.Run(ctx, req.Input, opts...)
		}
	} else {
		resp, err = ag.Run(ctx, req.Input, opts...)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RunResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(RunResponse{
		Content:  resp.Content,
		ThreadID: threadID,
	})
}

func (s *AgentRuntimeServer) handleWorkflowRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	wf, ok := s.host.Workflow(name)
	if !ok {
		http.Error(w, fmt.Sprintf("Workflow '%s' not found", name), http.StatusNotFound)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Enable SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	// Generate RunID
	runID := uuid.New().String()

	// Send RunID as first event
	startPayload, _ := json.Marshal(map[string]string{
		"type":   "run_start",
		"run_id": runID,
	})
	fmt.Fprintf(w, "data: %s\n\n", string(startPayload))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	var opts []model.Option
	if temp, ok := req.ModelConfig["temperature"].(float64); ok {
		opts = append(opts, model.WithTemperature(float32(temp)))
	}

	// Create channel for events
	eventCh := make(chan interface{}, 100)

	// Inject workflow callbacks
	handler := &workflowEventHandler{ch: eventCh}
	ctx = WithWorkflowCallbacks(ctx, handler)

	// Since Workflow.Run is blocking, we run it in goroutine
	go func() {
		defer close(eventCh)

		// Use RunResumable if it's a DAGWorkflow (or supports it)
		// Currently only DAGWorkflow supports RunResumable.
		// We should probably promote RunResumable to Workflow interface or type assert.
		// For now, type assert.
		if dag, ok := wf.(*DAGWorkflow); ok {
			// Run with initial state (nil)
			resp, state, err := dag.RunResumable(ctx, req.Input, nil, opts...)

			// Save Checkpoint
			// Determine status
			status := StatusCompleted
			if err == ErrSuspended {
				status = StatusSuspended
			} else if err != nil {
				status = StatusFailed
			}

			// If suspended, we MUST save state to allow resume.
			// If completed/failed, we save for history.
			if s.checkpointStore != nil {
				cp := &Checkpoint{
					RunID:        runID,
					WorkflowName: name,
					Status:       status,
					State:        state,      // This is the state returned by RunResumable
					CreatedAt:    time.Now(), // Ideally start time
				}
				if err != nil && err != ErrSuspended {
					cp.Error = err.Error()
				}
				if saveErr := s.checkpointStore.Save(context.Background(), cp); saveErr != nil {
					log.Printf("Failed to save checkpoint for run %s: %v", runID, saveErr)
				}
			}

			if err != nil {
				if err == ErrSuspended {
					// Send suspended event
					eventCh <- map[string]string{
						"type":   "suspended",
						"info":   "Waiting for user input",
						"run_id": runID,
					}
					return
				}
				eventCh <- err
				return
			}
			eventCh <- resp
		} else {
			// Regular Run for non-resumable workflows
			resp, err := wf.Run(ctx, req.Input, opts...)
			if err != nil {
				eventCh <- err
				return
			}
			eventCh <- resp
		}
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	for event := range eventCh {
		switch e := event.(type) {
		case error:
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", e.Error())
			return
		case *schema.Message:
			// Final response
			payload := map[string]interface{}{
				"type":    "content_delta",
				"content": e.Content,
			}
			data, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		case map[string]string:
			// Node events or suspended
			data, _ := json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	}

	// Done
	donePayload, _ := json.Marshal(map[string]string{
		"type":   "status",
		"status": "completed",
	})
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", string(donePayload))
	flusher.Flush()
}

type workflowEventHandler struct {
	ch chan interface{}
}

func (h *workflowEventHandler) OnNodeStart(ctx context.Context, nodeID string, input string) {
	h.ch <- map[string]string{
		"type": "node_start",
		"node": nodeID,
	}
}

func (h *workflowEventHandler) OnNodeEnd(ctx context.Context, nodeID string, output string) {
	h.ch <- map[string]string{
		"type": "node_end",
		"node": nodeID,
	}
}

func (h *workflowEventHandler) OnWorkflowStart(ctx context.Context, workflowID string, input string) {
	h.ch <- map[string]string{
		"type":     "workflow_start",
		"workflow": workflowID,
	}
}

func (h *workflowEventHandler) OnWorkflowEnd(ctx context.Context, workflowID string, output string, status WorkflowExecutionStatus) {
	h.ch <- map[string]string{
		"type":     "workflow_end",
		"workflow": workflowID,
		"status":   string(status),
	}
}

func (s *AgentRuntimeServer) handleListAgents(w http.ResponseWriter, r *http.Request) {
	type AgentInfo struct {
		Name string `json:"name"`
		Type string `json:"type"` // "agent" or "workflow"
	}
	var list []AgentInfo

	for _, name := range s.host.ListAgents() {
		list = append(list, AgentInfo{Name: name, Type: "agent"})
	}
	for _, name := range s.host.ListWorkflows() {
		list = append(list, AgentInfo{Name: name, Type: "workflow"})
	}

	json.NewEncoder(w).Encode(list)
}

func (s *AgentRuntimeServer) handleWorkflowResume(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	wf, ok := s.host.Workflow(name)
	if !ok {
		http.Error(w, fmt.Sprintf("Workflow '%s' not found", name), http.StatusNotFound)
		return
	}

	var req struct {
		Input string `json:"input"`
		RunID string `json:"run_id"` // Add RunID support
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load Checkpoint if RunID is provided
	var resumeState *WorkflowState
	if req.RunID != "" && s.checkpointStore != nil {
		cp, err := s.checkpointStore.Load(r.Context(), req.RunID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Checkpoint not found for run %s: %v", req.RunID, err), http.StatusNotFound)
			return
		}
		if cp.WorkflowName != name {
			http.Error(w, fmt.Sprintf("Checkpoint run %s belongs to workflow %s, not %s", req.RunID, cp.WorkflowName, name), http.StatusBadRequest)
			return
		}
		resumeState = cp.State
		log.Printf("Resuming workflow %s run %s from checkpoint", name, req.RunID)
	}

	// Enable SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	// Inject Resume Input
	ctx = context.WithValue(ctx, "resume_input", req.Input)

	// Inject workflow callbacks
	eventCh := make(chan interface{}, 100)
	handler := &workflowEventHandler{ch: eventCh}
	ctx = WithWorkflowCallbacks(ctx, handler)

	go func() {
		defer close(eventCh)

		// Use RunResumable with loaded state
		if dag, ok := wf.(*DAGWorkflow); ok {
			resp, state, err := dag.RunResumable(ctx, req.Input, resumeState, nil...)

			// Save new checkpoint
			runID := req.RunID
			if runID == "" {
				runID = uuid.New().String() // Should not happen if resuming, but just in case
			}

			status := StatusCompleted
			if err == ErrSuspended {
				status = StatusSuspended
			} else if err != nil {
				status = StatusFailed
			}

			if s.checkpointStore != nil {
				cp := &Checkpoint{
					RunID:        runID,
					WorkflowName: name,
					Status:       status,
					State:        state,
					UpdatedAt:    time.Now(),
				}
				if err != nil && err != ErrSuspended {
					cp.Error = err.Error()
				}
				// Preserve created at if we loaded it?
				// For simplicity, just save. File store handles timestamps roughly.
				_ = s.checkpointStore.Save(context.Background(), cp)
			}

			if err != nil {
				eventCh <- err
				return
			}
			eventCh <- resp
		} else {
			// Fallback for non-resumable
			resp, err := wf.Run(ctx, req.Input, nil...)
			if err != nil {
				eventCh <- err
				return
			}
			eventCh <- resp
		}
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	for event := range eventCh {
		switch e := event.(type) {
		case error:
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", e.Error())
			return
		case *schema.Message:
			payload := map[string]interface{}{
				"type":    "content_delta",
				"content": e.Content,
			}
			data, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		case map[string]string:
			data, _ := json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	}

	donePayload, _ := json.Marshal(map[string]string{
		"type":   "status",
		"status": "completed",
	})
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", string(donePayload))
	flusher.Flush()
}

func (s *AgentRuntimeServer) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if s.checkpointStore == nil {
		http.Error(w, "checkpoint store not initialized", http.StatusNotImplemented)
		return
	}

	runs, err := s.checkpointStore.List(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list runs: %v", err), http.StatusInternalServerError)
		return
	}

	// Sort by CreatedAt desc
	// For simplicity, we assume store returns in some order or we rely on frontend.
	// But let's just return JSON.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *AgentRuntimeServer) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.checkpointStore == nil {
		http.Error(w, "checkpoint store not initialized", http.StatusNotImplemented)
		return
	}

	run, err := s.checkpointStore.Load(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load run %s: %v", id, err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (s *AgentRuntimeServer) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	// We need to inspect the HostConfig or internal structure to get the graph.
	// Host doesn't expose internal map easily for graph structure inspection unless we access config.
	// But we have s.host.cfg available inside Host if we add a getter, or we can look up s.host.
	// Host struct fields are private.
	// Let's add a method to Host to GetWorkflowGraph.

	graph, err := s.host.GetWorkflowGraph(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(graph)
}
