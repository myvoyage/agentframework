// Agent Framework - Lane Queue Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSessionKey(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		channel   string
		userID    string
		expected  string
	}{
		{
			name:      "standard session",
			workspace: "default",
			channel:   "telegram",
			userID:    "user123",
			expected:  "default:telegram:user123",
		},
		{
			name:      "cron lane",
			workspace: "cron",
			channel:   "scheduler",
			userID:    "0",
			expected:  "cron:scheduler:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := NewSessionKey(tt.workspace, tt.channel, tt.userID)
			if string(key) != tt.expected {
				t.Errorf("NewSessionKey() = %v, want %v", key, tt.expected)
			}
		})
	}
}

func TestParseSessionKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantWS    string
		wantCh    string
		wantUser  string
	}{
		{
			name:     "standard key",
			key:      "default:telegram:user123",
			wantWS:   "default",
			wantCh:   "telegram",
			wantUser: "user123",
		},
		{
			name:     "cron key",
			key:      "cron:scheduler:0",
			wantWS:   "cron",
			wantCh:   "scheduler",
			wantUser: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, channel, userID := ParseSessionKey(SessionKey(tt.key))
			if workspace != tt.wantWS || channel != tt.wantCh || userID != tt.wantUser {
				t.Errorf("ParseSessionKey() = (%v, %v, %v), want (%v, %v, %v)",
					workspace, channel, userID, tt.wantWS, tt.wantCh, tt.wantUser)
			}
		})
	}
}

func TestIsSpecialLane(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"cron lane", "cron:scheduler:0", true},
		{"subagent lane", "subagent:task:abc123", true},
		{"background lane", "background:worker:1", true},
		{"standard lane", "default:telegram:user123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSpecialLane(SessionKey(tt.key)); got != tt.expected {
				t.Errorf("IsSpecialLane() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLaneQueue_EnqueueAndExecute(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	var executed atomic.Int32

	taskID, err := lq.Enqueue(ctx, key, func(context.Context) error {
		executed.Add(1)
		return nil
	}, "", 0)

	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 1 {
		t.Errorf("Task was not executed, executed = %v", executed.Load())
	}

	// Check result
	result, ok := lq.GetResult(taskID)
	if !ok {
		t.Fatal("GetResult() returned false")
	}

	if result.Error != nil {
		t.Errorf("Result error = %v, want nil", result.Error)
	}
}

func TestLaneQueue_SerialExecution(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	var order []int
	var mu sync.Mutex

	// Enqueue multiple tasks
	for i := 0; i < 5; i++ {
		i := i
		lq.Enqueue(ctx, key, func(context.Context) error {
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Small delay
			return nil
		}, "", 0)
	}

	// Wait for all tasks to complete
	time.Sleep(500 * time.Millisecond)

	// Check execution order
	if len(order) != 5 {
		t.Fatalf("Expected 5 executions, got %d", len(order))
	}

	for i, execOrder := range order {
		if execOrder != i {
			t.Errorf("Task executed out of order: got %v, want [0,1,2,3,4]", order)
			break
		}
	}
}

func TestLaneQueue_ParallelLanes(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()

	// Two different lanes
	key1 := SessionKey("default:telegram:user1")
	key2 := SessionKey("default:telegram:user2")

	var executed atomic.Int32
	var maxConcurrent atomic.Int32

	// Enqueue tasks for both lanes
	for range 3 {
		lq.Enqueue(ctx, key1, func(c context.Context) error {
			current := maxConcurrent.Add(1)
			defer maxConcurrent.Add(-1)

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			if current > 1 {
				t.Errorf("Detected concurrent execution in same lane: %v", current)
			}

			executed.Add(1)
			return nil
		}, "", 0)
	}

	for range 3 {
		lq.Enqueue(ctx, key2, func(c context.Context) error {
			maxConcurrent.Add(1)
			defer maxConcurrent.Add(-1)

			time.Sleep(50 * time.Millisecond)
			executed.Add(1)
			return nil
		}, "", 0)
	}

	// Wait for completion
	time.Sleep(500 * time.Millisecond)

	if executed.Load() != 6 {
		t.Errorf("Expected 6 executions, got %d", executed.Load())
	}
}

func TestLaneQueue_Idempotency(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	var executed atomic.Int32

	// First task
	_, err := lq.Enqueue(ctx, key, func(context.Context) error {
		executed.Add(1)
		return nil
	}, "same-key", 0)
	if err != nil {
		t.Fatalf("First Enqueue() error = %v", err)
	}

	// Duplicate task with same idempotency key
	_, err = lq.Enqueue(ctx, key, func(context.Context) error {
		executed.Add(1)
		return nil
	}, "same-key", 0)
	if err == nil {
		t.Error("Expected error for duplicate task, got nil")
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 1 {
		t.Errorf("Expected 1 execution with idempotency, got %d", executed.Load())
	}
}

func TestLaneQueue_CancelTask(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	var executed atomic.Int32

	// Enqueue a long-running task
	taskID, _ := lq.Enqueue(ctx, key, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			executed.Add(1)
			return nil
		}
	}, "", 0)

	// Cancel immediately
	time.Sleep(10 * time.Millisecond)
	err := lq.CancelTask(taskID)
	if err != nil {
		t.Errorf("CancelTask() error = %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 0 {
		t.Errorf("Task was executed despite cancellation, executed = %v", executed.Load())
	}

	// Check result
	result, ok := lq.GetResult(taskID)
	if !ok {
		t.Fatal("GetResult() returned false")
	}

	if result.Error == nil || result.Error != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", result.Error)
	}
}

func TestLaneQueue_GetLaneStatus(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	// Check empty lane
	queueLen, running := lq.GetLaneStatus(key)
	if queueLen != 0 || running {
		t.Errorf("Empty lane status = (%v, %v), want (0, false)", queueLen, running)
	}

	// Enqueue a long-running task
	startChan := make(chan struct{})
	doneChan := make(chan struct{})
	lq.Enqueue(ctx, key, func(c context.Context) error {
		close(startChan)
		<-doneChan
		return nil
	}, "", 0)

	// Wait for start
	<-startChan

	// Check status while running
	queueLen, running = lq.GetLaneStatus(key)
	if queueLen != 0 || !running {
		t.Errorf("Running lane status = (%v, %v), want (0, true)", queueLen, running)
	}

	// Enqueue another task
	lq.Enqueue(ctx, key, func(context.Context) error { return nil }, "", 0)

	// Check queue length
	queueLen, _ = lq.GetLaneStatus(key)
	if queueLen != 1 {
		t.Errorf("Queue length = %v, want 1", queueLen)
	}

	// Finish first task
	close(doneChan)
	time.Sleep(100 * time.Millisecond)
}

func TestLaneQueue_GetAllStatus(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key1 := SessionKey("default:telegram:user1")
	key2 := SessionKey("default:telegram:user2")

	lq.Enqueue(ctx, key1, func(context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, "", 0)

	lq.Enqueue(ctx, key2, func(context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, "", 0)

	time.Sleep(10 * time.Millisecond)

	status := lq.GetAllStatus()

	if len(status) != 2 {
		t.Errorf("GetAllStatus() returned %d lanes, want 2", len(status))
	}

	for key, st := range status {
		if st[1] != 1 { // running
			t.Errorf("Lane %v should be running, status = %v", key, st)
		}
	}
}

func TestLaneQueue_Timeout(t *testing.T) {
	lq := NewLaneQueue()
	defer lq.Stop()

	ctx := context.Background()
	key := SessionKey("default:telegram:user123")

	taskID, _ := lq.Enqueue(ctx, key, func(context.Context) error {
		time.Sleep(5 * time.Second)
		return nil
	}, "", 100*time.Millisecond)

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	result, ok := lq.GetResult(taskID)
	if !ok {
		t.Fatal("GetResult() returned false")
	}

	if result.Error == nil {
		t.Error("Expected timeout error, got nil")
	}
}
