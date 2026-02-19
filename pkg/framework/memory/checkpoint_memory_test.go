// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// Thread Store Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewMemoryCheckpointStore(t *testing.T) {
	store := NewMemoryCheckpointStore()
	if store == nil {
		t.Fatal("expected MemoryCheckpointStore to be created")
	}

	if store.config.MaxCheckpoints != DefaultMemoryCheckpointStoreConfig().MaxCheckpoints {
		t.Errorf("expected MaxCheckpoints %d, got %d", DefaultMemoryCheckpointStoreConfig().MaxCheckpoints, store.config.MaxCheckpoints)
	}

	if store.config.TTL != DefaultMemoryCheckpointStoreConfig().TTL {
		t.Errorf("expected TTL %v, got %v", DefaultMemoryCheckpointStoreConfig().TTL, store.config.TTL)
	}

	if store.config.CleanupInterval != DefaultMemoryCheckpointStoreConfig().CleanupInterval {
		t.Errorf("expected CleanupInterval %v, got %v", DefaultMemoryCheckpointStoreConfig().CleanupInterval, store.config.CleanupInterval)
	}

	if store.data == nil {
		t.Error("expected data map to be initialized")
	}
}

func TestNewMemoryCheckpointStoreWithConfig(t *testing.T) {
	config := MemoryCheckpointStoreConfig{
		MaxCheckpoints:  500,
		TTL:             12 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	}

	store := NewMemoryCheckpointStoreWithConfig(config)
	if store == nil {
		t.Fatal("expected MemoryCheckpointStore to be created with config")
	}

	if store.config.MaxCheckpoints != config.MaxCheckpoints {
		t.Errorf("expected MaxCheckpoints %d, got %d", config.MaxCheckpoints, store.config.MaxCheckpoints)
	}

	if store.config.TTL != config.TTL {
		t.Errorf("expected TTL %v, got %v", config.TTL, store.config.TTL)
	}

	if store.config.CleanupInterval != config.CleanupInterval {
		t.Errorf("expected CleanupInterval %v, got %v", config.CleanupInterval, store.config.CleanupInterval)
	}
}

func TestMemoryCheckpointStore_SaveAndLoad(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create a checkpoint
	runID := uuid.New().String()
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
		State:        &WorkflowState{Status: "running"},
		Input:        "test-input",
		Progress:     0.5,
		Metadata:     map[string]string{"key": "value"},
	}

	// Save checkpoint
	err := store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Load checkpoint
	loaded, err := store.Load(ctx, runID)
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if loaded == nil {
		t.Fatal("expected checkpoint to be loaded")
	}

	if loaded.RunID != runID {
		t.Errorf("expected runID %s, got %s", runID, loaded.RunID)
	}

	if loaded.WorkflowName != cp.WorkflowName {
		t.Errorf("expected workflow name %s, got %s", cp.WorkflowName, loaded.WorkflowName)
	}

	if loaded.Status != cp.Status {
		t.Errorf("expected status %s, got %s", cp.Status, loaded.Status)
	}

	if loaded.State.Status != cp.State.Status {
		t.Errorf("expected state status %s, got %s", cp.State.Status, loaded.State.Status)
	}

	if loaded.Input != cp.Input {
		t.Errorf("expected input %s, got %s", cp.Input, loaded.Input)
	}

	if loaded.Progress != cp.Progress {
		t.Errorf("expected progress %.2f, got %.2f", cp.Progress, loaded.Progress)
	}

	if loaded.Metadata["key"] != cp.Metadata["key"] {
		t.Errorf("expected metadata key %s, got %s", cp.Metadata["key"], loaded.Metadata["key"])
	}
}

func TestMemoryCheckpointStore_SaveUpdate(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create a checkpoint
	runID := uuid.New().String()
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
		State:        &WorkflowState{Status: "running"},
		Input:        "test-input",
		Progress:     0.5,
	}

	// Save initial checkpoint
	err := store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Update checkpoint
	cp.Status = StatusCompleted
	cp.Progress = 1.0
	cp.Output = "test-output"
	err = store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to update checkpoint: %v", err)
	}

	// Load updated checkpoint
	loaded, err := store.Load(ctx, runID)
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if loaded.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, loaded.Status)
	}

	if loaded.Progress != 1.0 {
		t.Errorf("expected progress %.2f, got %.2f", 1.0, loaded.Progress)
	}

	if loaded.Output != "test-output" {
		t.Errorf("expected output %s, got %s", "test-output", loaded.Output)
	}
}

func TestMemoryCheckpointStore_LoadNonExistent(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	nonExistentID := uuid.New().String()
	_, err := store.Load(ctx, nonExistentID)
	if err == nil {
		t.Error("expected error when loading non-existent checkpoint")
	}
}

func TestMemoryCheckpointStore_Delete(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create a checkpoint
	runID := uuid.New().String()
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
	}

	// Save checkpoint
	err := store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Delete checkpoint
	err = store.Delete(ctx, runID)
	if err != nil {
		t.Fatalf("failed to delete checkpoint: %v", err)
	}

	// Try to load deleted checkpoint
	_, err = store.Load(ctx, runID)
	if err == nil {
		t.Error("expected error when loading deleted checkpoint")
	}
}

func TestMemoryCheckpointStore_List(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create multiple checkpoints
	for i := 0; i < 5; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: "test-workflow",
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	// List checkpoints
	checkpoints, err := store.List(ctx)
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 5 {
		t.Errorf("expected 5 checkpoints, got %d", len(checkpoints))
	}
}

func TestMemoryCheckpointStore_ListWithOptions(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create checkpoints with different workflows and statuses
	workflow1 := "workflow1"
	workflow2 := "workflow2"

	for i := 0; i < 3; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: workflow1,
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	for i := 0; i < 2; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: workflow2,
			Status:       StatusCompleted,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	// List checkpoints for workflow1
	checkpoints, err := store.List(ctx, WithWorkflowName(workflow1))
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("expected 3 checkpoints for workflow1, got %d", len(checkpoints))
	}

	for _, cp := range checkpoints {
		if cp.WorkflowName != workflow1 {
			t.Errorf("expected workflow name %s, got %s", workflow1, cp.WorkflowName)
		}
	}

	// List checkpoints with completed status
	checkpoints, err = store.List(ctx, WithStatus(StatusCompleted))
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("expected 2 completed checkpoints, got %d", len(checkpoints))
	}

	for _, cp := range checkpoints {
		if cp.Status != StatusCompleted {
			t.Errorf("expected status %s, got %s", StatusCompleted, cp.Status)
		}
	}
}

func TestMemoryCheckpointStore_ListWithPagination(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create 10 checkpoints
	for i := 0; i < 10; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: "test-workflow",
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	// List with pagination
	checkpoints, err := store.List(ctx, WithLimit(3), WithOffset(2))
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("expected 3 checkpoints, got %d", len(checkpoints))
	}
}

func TestMemoryCheckpointStore_MaxCheckpoints(t *testing.T) {
	config := MemoryCheckpointStoreConfig{
		MaxCheckpoints:  3,
		TTL:             24 * time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
	store := NewMemoryCheckpointStoreWithConfig(config)
	ctx := context.Background()

	// Create 4 checkpoints
	for i := 0; i < 4; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: "test-workflow",
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	// Should only have 3 checkpoints
	checkpoints, err := store.List(ctx)
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("expected 3 checkpoints, got %d", len(checkpoints))
	}
}

func TestMemoryCheckpointStore_TTL_Cleanup(t *testing.T) {
	const ttl = 100 * time.Millisecond
	config := MemoryCheckpointStoreConfig{
		MaxCheckpoints:  100,
		TTL:             ttl,
		CleanupInterval: 1 * time.Second,
	}
	store := NewMemoryCheckpointStoreWithConfig(config)
	ctx := context.Background()

	// Create a checkpoint
	runID := uuid.New().String()
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
	}

	err := store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Verify it exists
	if _, err := store.Load(ctx, runID); err != nil {
		t.Errorf("checkpoint should exist before TTL: %v", err)
	}

	// Let TTL expire
	time.Sleep(ttl * 2)

	// Run cleanup manually
	store.cleanupExpired()

	// Check if checkpoint was cleaned up
	if _, err := store.Load(ctx, runID); err == nil {
		t.Error("checkpoint should have been cleaned up")
	}
}

func TestMemoryCheckpointStore_HealthCheck(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestMemoryCheckpointStore_Close(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	if err := store.Close(ctx); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestWithSortBy(t *testing.T) {
	opts := &ListCheckpointOptions{}

	WithSortBy("created_at")(opts)

	if opts.SortBy != "created_at" {
		t.Errorf("expected SortBy 'created_at', got '%s'", opts.SortBy)
	}
}

func TestWithSortDesc(t *testing.T) {
	opts := &ListCheckpointOptions{}

	WithSortDesc(true)(opts)

	if !opts.SortDesc {
		t.Error("expected SortDesc to be true")
	}
}

func TestWithSortOptions(t *testing.T) {
	store := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Create 5 checkpoints
	for i := 0; i < 5; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: "test-workflow",
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
	}

	// List with sort by created_at descending
	checkpoints, err := store.List(ctx, WithSortBy("created_at"), WithSortDesc(true))
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 5 {
		t.Errorf("expected 5 checkpoints, got %d", len(checkpoints))
	}
}

func TestMemoryCheckpointStore_MaxCheckpointsLimit(t *testing.T) {
	// Create store with small MaxCheckpoints limit
	config := MemoryCheckpointStoreConfig{
		MaxCheckpoints:  3,
		TTL:             1 * time.Hour,
		CleanupInterval: 1 * time.Second,
	}
	store := NewMemoryCheckpointStoreWithConfig(config)
	ctx := context.Background()

	// Create 5 checkpoints (more than the limit)
	for i := 0; i < 5; i++ {
		runID := uuid.New().String()
		cp := &Checkpoint{
			RunID:        runID,
			WorkflowName: "test-workflow",
			Status:       StatusRunning,
		}
		err := store.Save(ctx, cp)
		if err != nil {
			t.Fatalf("failed to save checkpoint %d: %v", i, err)
		}
		// Add small delay to ensure different UpdatedAt times
		time.Sleep(5 * time.Millisecond)
	}

	// Manually trigger cleanup to enforce max limit
	store.cleanupExpired()

	// Verify we have at most MaxCheckpoints
	checkpoints, err := store.List(ctx)
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) > 3 {
		t.Errorf("expected at most 3 checkpoints after cleanup, got %d", len(checkpoints))
	}
}
