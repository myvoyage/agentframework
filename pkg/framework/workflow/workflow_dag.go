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

package workflow

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// NodeConfig defines configuration for a DAG node.
type NodeConfig struct {
	Workflow   Workflow          // The workflow to execute
	Priority   int               // Node priority (higher numbers mean higher priority)
	MaxRetries int               // Maximum number of retries on failure
	RetryDelay time.Duration     // Delay between retries
	Timeout    time.Duration     // Execution timeout
	Conditions []string          // Conditions for node execution
	Resources  map[string]int    // Resource requirements (e.g., CPU, memory)
	Metadata   map[string]string // Additional metadata
}

// DAGWorkflow represents a directed acyclic graph workflow.
type DAGWorkflow struct {
	name          string
	nodes         map[string]NodeConfig // node ID to config mapping
	edges         map[string][]string   // from -> [to1, to2]
	store         CheckpointStore
	maxConcurrent int                  // Maximum number of concurrent nodes
	resourcePool  map[string]int       // Resource pool for node execution
	nodeStats     map[string]NodeStats // Node execution statistics
	muStats       sync.RWMutex         // Mutex for protecting nodeStats
}

func NewDAGWorkflow(name string) *DAGWorkflow {
	return &DAGWorkflow{
		name:          name,
		nodes:         make(map[string]NodeConfig),
		edges:         make(map[string][]string),
		store:         NewMemoryCheckpointStore(), // Default to memory store
		maxConcurrent: 0,                          // 0 means no limit
		resourcePool:  make(map[string]int),
		nodeStats:     make(map[string]NodeStats),
	}
}

// AddNode adds a node to the DAG with default configuration
func (w *DAGWorkflow) AddNode(id string, wf interface{}) {
	// Check if it's already a Workflow
	if workflow, ok := wf.(Workflow); ok {
		w.AddNodeWithConfig(id, NodeConfig{
			Workflow:   workflow,
			Priority:   0,
			MaxRetries: 0,
			RetryDelay: 1 * time.Second,
			Timeout:    5 * time.Minute,
			Conditions: []string{},
			Resources:  map[string]int{},
			Metadata:   map[string]string{},
		})
		return
	}

	// Check if it's an Agent, create a simple adapter
	if agent, ok := wf.(Agent); ok {
		// Create a simple workflow that runs the agent
		agentWorkflow := &simpleAgentWorkflow{
			name:  id,
			agent: agent,
		}
		w.AddNodeWithConfig(id, NodeConfig{
			Workflow:   agentWorkflow,
			Priority:   0,
			MaxRetries: 0,
			RetryDelay: 1 * time.Second,
			Timeout:    5 * time.Minute,
			Conditions: []string{},
			Resources:  map[string]int{},
			Metadata:   map[string]string{},
		})
		return
	}
}

// AddNodeWithConfig adds a node to the DAG with custom configuration
func (w *DAGWorkflow) AddNodeWithConfig(id string, cfg NodeConfig) {
	// Set default values if not provided
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 1 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.Conditions == nil {
		cfg.Conditions = []string{}
	}
	if cfg.Resources == nil {
		cfg.Resources = map[string]int{}
	}
	if cfg.Metadata == nil {
		cfg.Metadata = map[string]string{}
	}
	w.nodes[id] = cfg
}

func (w *DAGWorkflow) AddEdge(from, to string) {
	w.edges[from] = append(w.edges[from], to)
}

// SetMaxConcurrent sets the maximum number of concurrent nodes
func (w *DAGWorkflow) SetMaxConcurrent(max int) {
	if max > 0 {
		w.maxConcurrent = max
	}
}

func (w *DAGWorkflow) Name() string {
	return w.name
}

func (w *DAGWorkflow) WithCheckpointStore(store CheckpointStore) Workflow {
	w.store = store
	return w
}

// WorkflowState represents the state of a workflow execution
type WorkflowState struct {
	NodeStates map[string]string `json:"node_states"` // nodeID -> output content
}

// NodeStats represents execution statistics for a node
type NodeStats struct {
	ExecutionCount int           `json:"execution_count"`
	TotalDuration  time.Duration `json:"total_duration"`
	AvgDuration    time.Duration `json:"avg_duration"`
	SuccessCount   int           `json:"success_count"`
	ErrorCount     int           `json:"error_count"`
	LastExecution  time.Time     `json:"last_execution"`
}

// Use NodeExecutionResult from workflow.go

func (w *DAGWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Create a new run ID for this execution
	runID := uuid.NewString()

	// Create initial checkpoint
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: w.name,
		Status:       StatusRunning,
		Input:        input,
		Progress:     0.0,
		State:        map[string]interface{}{"node_states": make(map[string]string)},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save initial checkpoint
	if err := w.store.SaveCheckpoint(ctx, cp); err != nil {
		return nil, err
	}

	// Run the workflow
	resp, state, err := w.RunResumable(ctx, input, nil, opts...)

	// Update checkpoint with final status
	cp.UpdatedAt = time.Now()
	if err != nil {
		if err == ErrSuspended {
			cp.Status = StatusSuspended
			cp.Progress = float64(len(state.NodeStates)) / float64(len(w.nodes))
		} else {
			cp.Status = StatusFailed
			cp.Error = err.Error()
		}
	} else {
		cp.Status = StatusCompleted
		cp.Progress = 1.0
		cp.Output = resp.Content
	}

	cp.State = state
	_ = w.store.Save(ctx, cp)

	return resp, err
}

func (w *DAGWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	// Load checkpoint
	cp, err := w.store.Load(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Check if workflow is already completed
	if cp.Status == StatusCompleted {
		return &schema.Message{Content: cp.Output}, nil
	}

	// Update checkpoint status to running
	cp.Status = StatusRunning
	cp.UpdatedAt = time.Now()
	if err := w.store.Save(ctx, cp); err != nil {
		return nil, err
	}

	// Use the input from checkpoint if no new input is provided
	if input == "" {
		input = cp.Input
	}

	// Run the workflow from the saved state
	resp, state, err := w.RunResumable(ctx, input, cp.State, opts...)

	// Update checkpoint with final status
	cp.UpdatedAt = time.Now()
	if err != nil {
		if err == ErrSuspended {
			cp.Status = StatusSuspended
			cp.Progress = float64(len(state.NodeStates)) / float64(len(w.nodes))
		} else {
			cp.Status = StatusFailed
			cp.Error = err.Error()
		}
	} else {
		cp.Status = StatusCompleted
		cp.Progress = 1.0
		cp.Output = resp.Content
	}

	cp.State = state
	_ = w.store.Save(ctx, cp)

	return resp, err
}

func (w *DAGWorkflow) RunResumable(ctx context.Context, input string, resumeState *WorkflowState, opts ...model.Option) (*schema.Message, *WorkflowState, error) {
	// Initialize state
	results := make(map[string]string)
	if resumeState != nil && resumeState.NodeStates != nil {
		for k, v := range resumeState.NodeStates {
			results[k] = v
		}
	}
	results["__input__"] = input

	// Calculate total nodes and completed count early
	totalNodes := len(w.nodes)
	completed := make(map[string]bool)
	for id := range results {
		if id != "__input__" {
			completed[id] = true
		}
	}
	completedCount := len(completed)

	// 1. Calculate in-degrees
	inDegree := make(map[string]int)
	for id := range w.nodes {
		inDegree[id] = 0
	}
	for _, tos := range w.edges {
		for _, to := range tos {
			inDegree[to]++
		}
	}

	// Mark completed nodes in in-degree map
	for id := range completed {
		inDegree[id] = -1 // Mark as completed
	}

	// 2. Calculate initial ready nodes with priority
	// Pre-allocate slice size for readyNodes
	readyNodes := make([]string, 0, len(w.nodes))
	for id, degree := range inDegree {
		if degree == 0 && !completed[id] {
			readyNodes = append(readyNodes, id)
		}
	}

	// Sort ready nodes by priority (higher priority first)
	sort.Slice(readyNodes, func(i, j int) bool {
		return w.nodes[readyNodes[i]].Priority > w.nodes[readyNodes[j]].Priority
	})

	// 3. Setup concurrent execution with improved resource management
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure cleanup

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a channel to control maximum concurrent execution
	var concurrencyCh chan struct{}
	if w.maxConcurrent > 0 {
		concurrencyCh = make(chan struct{}, w.maxConcurrent)
	} else {
		// Default to reasonable maximum concurrent nodes (number of CPUs or 8, whichever is smaller)
		maxConcurrent := runtime.NumCPU()
		if maxConcurrent > 8 {
			maxConcurrent = 8
		}
		concurrencyCh = make(chan struct{}, maxConcurrent)
	}

	// Use a sync.Map for better performance in concurrent scenarios
	doneMap := sync.Map{}
	for id := range w.nodes {
		if completed[id] {
			doneMap.Store(id, struct{}{})
		}
	}

	// Channels for results - use exact size based on remaining nodes
	remainingNodes := totalNodes - completedCount
	resultCh := make(chan NodeExecutionResult, remainingNodes)

	// Helper to safely update state
	updateState := func(id, content string) {
		mu.Lock()
		defer mu.Unlock()
		results[id] = content
	}

	// Helper to get node dependencies
	getDependencies := func(nodeID string) []string {
		// Pre-allocate slice size for deps
		deps := make([]string, 0, len(w.edges))
		for from, tos := range w.edges {
			for _, to := range tos {
				if to == nodeID {
					deps = append(deps, from)
					break // No need to check other edges for this from node
				}
			}
		}
		return deps
	}

	// Helper to check if all dependencies are done
	areDependenciesDone := func(nodeID string) bool {
		deps := getDependencies(nodeID)
		for _, dep := range deps {
			if _, ok := doneMap.Load(dep); !ok {
				return false
			}
		}
		return true
	}

	// Process ready nodes with improved concurrency control
	processNode := func(nodeID string) {
		defer wg.Done()

		// Wait for concurrency slot if needed
		select {
		case concurrencyCh <- struct{}{}:
		case <-ctx.Done():
			return
		}
		defer func() {
			<-concurrencyCh
		}()

		// Wait for dependencies to complete
		for !areDependenciesDone(nodeID) {
			select {
			case <-time.After(100 * time.Millisecond): // Short poll interval
			case <-ctx.Done():
				return
			}
		}

		// Check context before execution
		if ctx.Err() != nil {
			return
		}

		// Gather inputs efficiently
		var nodeInput string
		deps := getDependencies(nodeID)
		if len(deps) == 0 {
			nodeInput = results["__input__"]
		} else {
			// Pre-allocate strings.Builder capacity based on average node output size
			// This is a heuristic, but better than default capacity
			var sb strings.Builder
			mu.Lock()
			for i, dep := range deps {
				if val, ok := results[dep]; ok {
					if i > 0 {
						sb.WriteString("\n\n")
					}
					sb.WriteString(val)
				}
			}
			nodeInput = sb.String()
			mu.Unlock()
			// Clear the builder for reuse
			sb.Reset()
		}

		// Get node config
		nodeCfg := w.nodes[nodeID]

		// Execute with improved retry logic
		var resp *schema.Message
		var err error
		maxRetries := nodeCfg.MaxRetries
		retryDelay := nodeCfg.RetryDelay
		timeout := nodeCfg.Timeout
		if timeout == 0 {
			timeout = 5 * time.Minute // Default timeout
		}

		// Jitter for retry delays to avoid thundering herd
		jitter := func(d time.Duration) time.Duration {
			return time.Duration(rand.Int63n(int64(d/2)) + int64(d/2))
		}

		// Start execution timer
		startTime := time.Now()

		for attempt := 0; attempt <= maxRetries; attempt++ {
			// Create a timeout context for this node
			nodeCtx, nodeCancel := context.WithTimeout(ctx, timeout)

			// Emit Node Start
			if cb := GetWorkflowCallbacks(nodeCtx); cb != nil {
				cb.OnNodeStart(nodeCtx, nodeID, nodeInput)
			}

			// Execute the node
			resp, err = nodeCfg.Workflow.Run(nodeCtx, nodeInput, opts...)
			nodeCancel()

			if err == nil {
				// Success, break retry loop
				break
			}

			// Check if we should retry
			if attempt >= maxRetries {
				// Final attempt failed, return error
				duration := time.Since(startTime)
				// Update node stats for failed execution
				w.muStats.Lock()
				stats := w.nodeStats[nodeID]
				stats.ExecutionCount++
				stats.TotalDuration += duration
				stats.AvgDuration = stats.TotalDuration / time.Duration(stats.ExecutionCount)
				stats.ErrorCount++
				stats.LastExecution = startTime
				w.nodeStats[nodeID] = stats
				w.muStats.Unlock()

				resultCh <- NodeExecutionResult{
					NodeID:     nodeID,
					Status:     NodeStatusFailed,
					Input:      nodeInput,
					Output:     "",
					StartTime:  startTime,
					EndTime:    time.Now(),
					Error:      fmt.Sprintf("node %s failed after %d attempts: %v", nodeID, maxRetries+1, err),
					RetryCount: maxRetries + 1,
				}
				return
			}

			// Check if context was cancelled during retry delay
			select {
			case <-time.After(jitter(retryDelay)):
				// Wait with jitter before retry
			case <-ctx.Done():
				return
			}
		}

		// Check if we got a response
		if resp == nil {
			duration := time.Since(startTime)
			// Update node stats for failed execution
			w.muStats.Lock()
			stats := w.nodeStats[nodeID]
			stats.ExecutionCount++
			stats.TotalDuration += duration
			stats.AvgDuration = stats.TotalDuration / time.Duration(stats.ExecutionCount)
			stats.ErrorCount++
			stats.LastExecution = startTime
			w.nodeStats[nodeID] = stats
			w.muStats.Unlock()

			resultCh <- NodeExecutionResult{
				NodeID:     nodeID,
				Status:     NodeStatusFailed,
				Input:      nodeInput,
				Output:     "",
				StartTime:  startTime,
				EndTime:    time.Now(),
				Error:      fmt.Sprintf("node %s returned nil response", nodeID),
				RetryCount: 0,
			}
			return
		}

		// Calculate execution duration
		duration := time.Since(startTime)

		// Update node stats for successful execution
		w.muStats.Lock()
		stats := w.nodeStats[nodeID]
		stats.ExecutionCount++
		stats.TotalDuration += duration
		stats.AvgDuration = stats.TotalDuration / time.Duration(stats.ExecutionCount)
		stats.SuccessCount++
		stats.LastExecution = startTime
		w.nodeStats[nodeID] = stats
		w.muStats.Unlock()

		// Emit Node End
		if cb := GetWorkflowCallbacks(ctx); cb != nil {
			cb.OnNodeEnd(ctx, nodeID, resp.Content)
		}

		// Update state and signal completion
		updateState(nodeID, resp.Content)
		doneMap.Store(nodeID, struct{}{})
		resultCh <- NodeExecutionResult{
			NodeID:     nodeID,
			Status:     NodeStatusCompleted,
			Input:      nodeInput,
			Output:     resp.Content,
			StartTime:  startTime,
			EndTime:    time.Now(),
			Error:      "",
			RetryCount: 0,
		}
	}

	// Start processing ready nodes
	for _, nodeID := range readyNodes {
		wg.Add(1)
		go processNode(nodeID)
	}

	// Pre-allocate processed map size
	processed := make(map[string]bool, totalNodes)
	for id := range completed {
		processed[id] = true
	}

	// Use a separate goroutine to handle new ready nodes
	// Use exact size for newReadyNodesCh based on remaining nodes
	newReadyNodesCh := make(chan string, remainingNodes)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case nodeID := <-newReadyNodesCh:
				if !processed[nodeID] && inDegree[nodeID] == 0 {
					processed[nodeID] = true
					wg.Add(1)
					go processNode(nodeID)
				}
			}
		}
	}()

	// Main event loop with improved error handling
	for completedCount < totalNodes {
		select {
		case result := <-resultCh:
			if result.Error != "" {
				// Return partial state for future resume
				finalState := &WorkflowState{NodeStates: make(map[string]string)}
				mu.Lock()
				for k, v := range results {
					if k != "__input__" {
						finalState.NodeStates[k] = v
					}
				}
				mu.Unlock()
				return nil, finalState, errors.New(result.Error)
			}

			// Update in-degree for dependent nodes and check for new ready nodes
			for _, to := range w.edges[result.NodeID] {
				mu.Lock()
				inDegree[to]--
				currentDegree := inDegree[to]
				mu.Unlock()

				if currentDegree == 0 {
					// Node is now ready
					select {
					case newReadyNodesCh <- to:
					case <-ctx.Done():
						return nil, nil, ctx.Err()
					}
				}
			}
			completedCount++
		case <-ctx.Done():
			// Context cancelled
			finalState := &WorkflowState{NodeStates: make(map[string]string)}
			mu.Lock()
			for k, v := range results {
				if k != "__input__" {
					finalState.NodeStates[k] = v
				}
			}
			mu.Unlock()
			return nil, finalState, ctx.Err()
		}
	}

	// Calculate final output (leaves)
	outCounts := make(map[string]int)
	for id := range w.nodes {
		outCounts[id] = 0
	}
	for from, tos := range w.edges {
		outCounts[from] = len(tos)
	}

	var finalOutputs []string
	mu.Lock()
	for id, count := range outCounts {
		if count == 0 {
			if val, ok := results[id]; ok {
				finalOutputs = append(finalOutputs, val)
			}
		}
	}
	mu.Unlock()

	// Create final state
	finalState := &WorkflowState{NodeStates: make(map[string]string)}
	mu.Lock()
	for k, v := range results {
		if k != "__input__" {
			finalState.NodeStates[k] = v
		}
	}
	mu.Unlock()

	// Return final result
	return &schema.Message{
		Role:    schema.Assistant,
		Content: strings.Join(finalOutputs, "\n\n---\n\n"),
	}, finalState, nil
}

// GetID returns the workflow ID
func (w *DAGWorkflow) GetID() string {
	return w.name
}

// GetName returns the workflow name
func (w *DAGWorkflow) GetName() string {
	return w.name
}

// GetType returns the workflow type
func (w *DAGWorkflow) GetType() string {
	return "dag"
}
