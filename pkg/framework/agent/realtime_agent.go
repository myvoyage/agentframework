// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package agent provides real-time data processing agent implementation.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	beadscontext "AgentFramework/pkg/beads/context"
	"AgentFramework/pkg/beads/stream"
)

// RealTimeAgent manages real-time data processing and event handling.
type RealTimeAgent struct {
	pipelines      map[string]*stream.DataPipeline
	realTimeCtx    *beadscontext.RealTimeContext
	subscribers    map[string][]EventHandler
	eventBus       chan Event
	mutex          sync.RWMutex
	bufferSize     int
	maxWorkers     int
	enabledMetrics bool
	ctx            context.Context    // Context for managing goroutine lifecycle
	cancel         context.CancelFunc // Cancel function to stop all goroutines
	started        bool               // Track if agent has been started
}

// Event represents a real-time event.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"source"`
}

// EventHandler handles real-time events.
type EventHandler func(ctx context.Context, event Event)

// NewRealTimeAgent creates a new RealTimeAgent instance.
func NewRealTimeAgent(bufferSize, maxWorkers int, metricsEnabled bool) *RealTimeAgent {
	return &RealTimeAgent{
		pipelines:      make(map[string]*stream.DataPipeline),
		realTimeCtx:    beadscontext.NewRealTimeContext(10000, 5*time.Minute),
		subscribers:    make(map[string][]EventHandler),
		eventBus:       make(chan Event, bufferSize),
		bufferSize:     bufferSize,
		maxWorkers:     maxWorkers,
		enabledMetrics: metricsEnabled,
	}
}

// Initialize initializes the real-time agent.
func (a *RealTimeAgent) Initialize(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.started {
		return fmt.Errorf("real-time agent already initialized")
	}

	// Create a cancellable context for managing goroutine lifecycle
	a.ctx, a.cancel = context.WithCancel(ctx)
	a.started = true

	// Start event bus processor with agent's context
	go a.processEventBus(a.ctx)

	return nil
}

// CreatePipeline creates a new data processing pipeline.
func (a *RealTimeAgent) CreatePipeline(ctx context.Context, pipelineID string, processors []stream.DataProcessor, opts ...stream.PipelineOption) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.pipelines[pipelineID]; exists {
		return fmt.Errorf("pipeline %s already exists", pipelineID)
	}

	// Apply default options
	defaultOpts := []stream.PipelineOption{
		stream.WithWorkers(a.maxWorkers),
		stream.WithBufferSize(a.bufferSize),
	}

	if a.enabledMetrics {
		defaultOpts = append(defaultOpts, stream.WithErrorHandler(func(err error) {
			// Log error or send to monitoring
			fmt.Printf("Pipeline %s error: %v\n", pipelineID, err)
		}))
	}

	opts = append(defaultOpts, opts...)

	pipeline := stream.NewDataPipeline(processors, opts...)
	if err := pipeline.Start(ctx); err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	a.pipelines[pipelineID] = pipeline

	// Start consuming pipeline output with agent's context
	// Use agent context to ensure goroutine can be cancelled when agent closes
	agentCtx := a.ctx
	if agentCtx == nil {
		// Fallback to provided context if agent not initialized
		agentCtx = ctx
	}
	go a.consumePipelineOutput(agentCtx, pipelineID, pipeline)

	return nil
}

// ProcessData processes data through a pipeline.
func (a *RealTimeAgent) ProcessData(ctx context.Context, pipelineID string, data interface{}) error {
	a.mutex.RLock()
	pipeline, exists := a.pipelines[pipelineID]
	a.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pipeline %s not found", pipelineID)
	}

	return pipeline.Process(data)
}

// SubscribeEvents subscribes to events of a specific type.
func (a *RealTimeAgent) SubscribeEvents(ctx context.Context, eventType string, handler EventHandler) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.subscribers[eventType] = append(a.subscribers[eventType], handler)

	return nil
}

// UnsubscribeEvents unsubscribes from events.
func (a *RealTimeAgent) UnsubscribeEvents(ctx context.Context, eventType string, handler EventHandler) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	handlers, exists := a.subscribers[eventType]
	if !exists {
		return fmt.Errorf("no subscribers for event type %s", eventType)
	}

	// Remove handler
	for i, h := range handlers {
		// Compare function pointers (simple implementation)
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			a.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type %s", eventType)
}

// PublishEvent publishes an event to the event bus.
func (a *RealTimeAgent) PublishEvent(ctx context.Context, event Event) error {
	// Use agent context if available for cancellation
	agentCtx := a.ctx
	if agentCtx == nil {
		agentCtx = ctx
	}

	select {
	case a.eventBus <- event:
		return nil
	case <-agentCtx.Done():
		return fmt.Errorf("event bus closed: %w", agentCtx.Err())
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("event bus timeout")
	}
}

// processEventBus processes events from the event bus.
func (a *RealTimeAgent) processEventBus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-a.eventBus:
			a.notifySubscribers(ctx, event)
		}
	}
}

// notifySubscribers notifies subscribers of an event.
func (a *RealTimeAgent) notifySubscribers(ctx context.Context, event Event) {
	a.mutex.RLock()
	handlers, exists := a.subscribers[event.Type]
	// Create a copy of handlers to avoid holding lock during execution
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	a.mutex.RUnlock()

	if !exists || len(handlersCopy) == 0 {
		return
	}

	// Use agent context if available, otherwise use provided context
	agentCtx := a.ctx
	if agentCtx == nil {
		agentCtx = ctx
	}

	for _, handler := range handlersCopy {
		// Check if context is cancelled before starting goroutine
		select {
		case <-agentCtx.Done():
			return
		default:
		}

		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Event handler panic: %v\n", r)
				}
			}()
			// Use agent context to allow cancellation
			h(agentCtx, event)
		}(handler)
	}
}

// consumePipelineOutput consumes output from a pipeline.
func (a *RealTimeAgent) consumePipelineOutput(ctx context.Context, pipelineID string, pipeline *stream.DataPipeline) {
	output := pipeline.Output()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-output:
			if !ok {
				return
			}

			// Store in real-time context
			key := fmt.Sprintf("pipeline:%s:output:%d", pipelineID, time.Now().UnixNano())
			if err := a.realTimeCtx.Set(ctx, key, data); err != nil {
				fmt.Printf("Failed to store pipeline output: %v\n", err)
			}

			// Publish event
			event := Event{
				Type:      "pipeline_output",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"pipeline_id": pipelineID,
					"data":        data,
				},
				Source: pipelineID,
			}

			if err := a.PublishEvent(ctx, event); err != nil {
				fmt.Printf("Failed to publish pipeline output event: %v\n", err)
			}
		}
	}
}

// GetPipelineMetrics returns metrics for a pipeline.
func (a *RealTimeAgent) GetPipelineMetrics(ctx context.Context, pipelineID string) (*stream.PipelineMetrics, error) {
	a.mutex.RLock()
	pipeline, exists := a.pipelines[pipelineID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", pipelineID)
	}

	metrics := pipeline.GetMetrics()
	return &metrics, nil
}

// GetAllPipelineMetrics returns metrics for all pipelines.
func (a *RealTimeAgent) GetAllPipelineMetrics(ctx context.Context) (map[string]*stream.PipelineMetrics, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	metrics := make(map[string]*stream.PipelineMetrics)

	for pipelineID, pipeline := range a.pipelines {
		m := pipeline.GetMetrics()
		metrics[pipelineID] = &m
	}

	return metrics, nil
}

// GetRealTimeStats returns statistics about the real-time context.
func (a *RealTimeAgent) GetRealTimeStats(ctx context.Context) (*beadscontext.RealTimeStats, error) {
	return a.realTimeCtx.GetStats(ctx), nil
}

// QueryRealTimeData queries the real-time context.
func (a *RealTimeAgent) QueryRealTimeData(ctx context.Context, query *beadscontext.Query) ([]*beadscontext.QueryResult, error) {
	return a.realTimeCtx.Query(ctx, *query)
}

// SearchRealTimeData searches the real-time context.
func (a *RealTimeAgent) SearchRealTimeData(ctx context.Context, searchTerm string, limit int) ([]*beadscontext.SearchResult, error) {
	return a.realTimeCtx.Search(ctx, searchTerm, limit)
}

// DeletePipeline removes a pipeline.
func (a *RealTimeAgent) DeletePipeline(ctx context.Context, pipelineID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	pipeline, exists := a.pipelines[pipelineID]
	if !exists {
		return fmt.Errorf("pipeline %s not found", pipelineID)
	}

	pipeline.Stop()
	delete(a.pipelines, pipelineID)

	return nil
}

// ListPipelines lists all pipeline IDs.
func (a *RealTimeAgent) ListPipelines(ctx context.Context) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	pipelineIDs := make([]string, 0, len(a.pipelines))
	for id := range a.pipelines {
		pipelineIDs = append(pipelineIDs, id)
	}

	return pipelineIDs, nil
}

// ClearRealTimeData clears all data from the real-time context.
func (a *RealTimeAgent) ClearRealTimeData(ctx context.Context) error {
	return a.realTimeCtx.Clear(ctx)
}

// Close closes the real-time agent and cleans up resources.
func (a *RealTimeAgent) Close(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.started {
		return nil // Already closed or never started
	}

	// Cancel all goroutines by cancelling the context
	if a.cancel != nil {
		a.cancel()
	}

	// Stop all pipelines
	for _, pipeline := range a.pipelines {
		pipeline.Stop()
	}

	a.pipelines = make(map[string]*stream.DataPipeline)

	// Close real-time context
	if err := a.realTimeCtx.Close(); err != nil {
		return fmt.Errorf("failed to close real-time context: %w", err)
	}

	// Clear subscribers
	a.subscribers = make(map[string][]EventHandler)

	// Close event bus channel safely
	select {
	case <-a.eventBus:
		// Channel already closed or empty
	default:
		close(a.eventBus)
	}

	a.started = false

	return nil
}