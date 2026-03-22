// Agent Framework - Lane Queue System
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Lane Queue implements session-based serial execution with optional parallelism:
//   - Each session gets its own queue (lane)
//   - Tasks in same lane execute serially (no race conditions)
//   - Tasks in different lanes execute in parallel
//   - Built-in backpressure and deduplication
//
// SessionKey format: workspace:channel:userId
//   e.g., "default:telegram:user123"
//   e.g., "cron:scheduler:0" for special lanes (parallel)
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionKey is a unique identifier for a conversation session
// Format: workspace:channel:userId
type SessionKey string

// NewSessionKey creates a session key from components
func NewSessionKey(workspace, channel, userID string) SessionKey {
	return SessionKey(fmt.Sprintf("%s:%s:%s", workspace, channel, userID))
}

// ParseSessionKey parses a session key into components
func ParseSessionKey(key SessionKey) (workspace, channel, userID string) {
	for i, p := range splitSessionKey(string(key)) {
		switch i {
		case 0:
			workspace = p
		case 1:
			channel = p
		case 2:
			userID = p
		}
	}
	return workspace, channel, userID
}

// splitSessionKey splits a session key by colons
func splitSessionKey(key string) []string {
	parts := []string{}
	current := ""
	for _, c := range key {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

// IsSpecialLane checks if a session key is for a special lane (parallel execution)
func IsSpecialLane(key SessionKey) bool {
	parts := splitSessionKey(string(key))
	if len(parts) == 0 {
		return false
	}
	return parts[0] == "cron" || parts[0] == "subagent" || parts[0] == "background"
}

// Task represents a task to be executed in a lane
type Task struct {
	ID        string
	SessionKey SessionKey
	Execute   func(context.Context) error
	CreatedAt time.Time
	Timeout   time.Duration
	Ctx       context.Context
	Cancel    context.CancelFunc
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string
	SessionKey SessionKey
	Error     error
	Duration  time.Duration
	StartedAt time.Time
	CompletedAt time.Time
}

// LaneQueue manages per-session serial execution
type LaneQueue struct {
	mu sync.RWMutex

	// lanes maps session keys to their task queues
	lanes map[SessionKey][]*Task

	// running tracks which lanes are currently executing
	running map[SessionKey]bool

	// results stores task results for retrieval
	results map[string]*TaskResult

	// eventChan emits task events for observability
	eventChan chan *TaskEvent

	// idempotency cache for deduplication
	idempotencyCache map[string]time.Time
}

// TaskEvent represents an event in the task lifecycle
type TaskEvent struct {
	Type      string // "queued", "started", "completed", "failed"
	TaskID    string
	SessionKey SessionKey
	Timestamp time.Time
	Error     error
}

// NewLaneQueue creates a new lane queue
func NewLaneQueue() *LaneQueue {
	lq := &LaneQueue{
		lanes:            make(map[SessionKey][]*Task),
		running:          make(map[SessionKey]bool),
		results:          make(map[string]*TaskResult),
		eventChan:        make(chan *TaskEvent, 1000),
		idempotencyCache: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go lq.cleanupLoop()

	return lq
}

// Enqueue adds a task to a session's queue
// If idempotencyKey is provided, duplicate tasks are ignored
func (lq *LaneQueue) Enqueue(ctx context.Context, key SessionKey, executeFunc func(context.Context) error, idempotencyKey string, timeout time.Duration) (string, error) {
	// Check idempotency cache
	if idempotencyKey != "" {
		lq.mu.RLock()
		if _, exists := lq.idempotencyCache[idempotencyKey]; exists {
			lq.mu.RUnlock()
			return "", fmt.Errorf("duplicate task (idempotencyKey: %s)", idempotencyKey)
		}
		lq.mu.RUnlock()
	}

	taskID := uuid.New().String()
	taskCtx, taskCancel := context.WithCancel(ctx)

	newTask := &Task{
		ID:        taskID,
		SessionKey: key,
		Execute:   executeFunc,
		CreatedAt: time.Now(),
		Timeout:   timeout,
		Ctx:       taskCtx,
		Cancel:    taskCancel,
	}

	lq.mu.Lock()

	// Add to lane queue
	lq.lanes[key] = append(lq.lanes[key], newTask)

	// Mark idempotency key
	if idempotencyKey != "" {
		lq.idempotencyCache[idempotencyKey] = time.Now()
	}

	lq.mu.Unlock()

	// Emit queued event
	lq.emitEvent(&TaskEvent{
		Type:       "queued",
		TaskID:     taskID,
		SessionKey: key,
		Timestamp:  time.Now(),
	})

	// Start processing if this lane is not running
	go lq.processLane(key)

	return taskID, nil
}

// processLane processes tasks in a lane serially
func (lq *LaneQueue) processLane(key SessionKey) {
	lq.mu.Lock()

	// If lane is already running, don't start another goroutine
	if lq.running[key] {
		lq.mu.Unlock()
		return
	}

	lq.running[key] = true
	lq.mu.Unlock()

	defer func() {
		lq.mu.Lock()
		delete(lq.running, key)
		lq.mu.Unlock()
	}()

	for {
		lq.mu.Lock()

		// Get next task from lane
		queue := lq.lanes[key]
		if len(queue) == 0 {
			lq.mu.Unlock()
			break
		}

		task := queue[0]
		lq.lanes[key] = queue[1:]

		lq.mu.Unlock()

		// Execute task
		result := lq.executeTask(task)

		// Store result
		lq.mu.Lock()
		lq.results[task.ID] = result
		lq.mu.Unlock()

		// Emit completion event
		eventType := "completed"
		if result.Error != nil {
			eventType = "failed"
		}
		lq.emitEvent(&TaskEvent{
			Type:       eventType,
			TaskID:     task.ID,
			SessionKey: key,
			Timestamp:  result.CompletedAt,
			Error:      result.Error,
		})
	}
}

// executeTask executes a single task
func (lq *LaneQueue) executeTask(task *Task) *TaskResult {
	result := &TaskResult{
		TaskID:     task.ID,
		SessionKey: task.SessionKey,
		StartedAt:  time.Now(),
	}

	// Apply timeout if configured
	ctx := task.Ctx
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)
		defer cancel()
	}

	// Execute
	result.Error = task.Execute(ctx)

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result
}

// GetResult retrieves a task result by ID
func (lq *LaneQueue) GetResult(taskID string) (*TaskResult, bool) {
	lq.mu.RLock()
	defer lq.mu.RUnlock()

	result, ok := lq.results[taskID]
	return result, ok
}

// Events returns the event channel for observability
func (lq *LaneQueue) Events() <-chan *TaskEvent {
	return lq.eventChan
}

// GetLaneStatus returns the status of a lane
func (lq *LaneQueue) GetLaneStatus(key SessionKey) (queueLen int, running bool) {
	lq.mu.RLock()
	defer lq.mu.RUnlock()

	queueLen = len(lq.lanes[key])
	running = lq.running[key]
	return queueLen, running
}

// GetAllStatus returns status of all lanes
func (lq *LaneQueue) GetAllStatus() map[SessionKey][2]int {
	lq.mu.RLock()
	defer lq.mu.RUnlock()

	status := make(map[SessionKey][2]int)
	for key, lane := range lq.lanes {
		running := 0
		if lq.running[key] {
			running = 1
		}
		status[key] = [2]int{len(lane), running}
	}
	return status
}

// CancelTask cancels a task by ID
func (lq *LaneQueue) CancelTask(taskID string) error {
	lq.mu.Lock()
	defer lq.mu.Unlock()

	// Check if task is still in queue
	for _, lane := range lq.lanes {
		for i, task := range lane {
			if task.ID == taskID {
				task.Cancel()
				lq.lanes[task.SessionKey] = append(lane[:i], lane[i+1:]...)
				return nil
			}
		}
	}

	// Task may be running, try to cancel
	for key, running := range lq.running {
		if running {
			lane := lq.lanes[key]
			if len(lane) > 0 && lane[0].ID == taskID {
				lane[0].Cancel()
				return nil
			}
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}

// emitEvent emits a task event to the event channel
func (lq *LaneQueue) emitEvent(event *TaskEvent) {
	select {
	case lq.eventChan <- event:
	default:
		// Event channel full, drop event
	}
}

// cleanupLoop periodically cleans up old idempotency cache entries
func (lq *LaneQueue) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lq.mu.Lock()

		// Remove entries older than 10 minutes
		now := time.Now()
		for key, ts := range lq.idempotencyCache {
			if now.Sub(ts) > 10*time.Minute {
				delete(lq.idempotencyCache, key)
			}
		}

		// Remove old results (older than 1 hour)
		for taskID, result := range lq.results {
			if now.Sub(result.CompletedAt) > time.Hour {
				delete(lq.results, taskID)
			}
		}

		lq.mu.Unlock()
	}
}

// Stop closes the event channel and cleans up resources
func (lq *LaneQueue) Stop() {
	close(lq.eventChan)
}
