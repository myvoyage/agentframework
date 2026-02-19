// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// File Checkpoint Store Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNewFileCheckpointStore(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

	if store == nil {
		t.Fatal("expected FileCheckpointStore to be created")
	}

	if store.dir != tmpDir {
		t.Errorf("expected directory %s, got %s", tmpDir, store.dir)
	}
}

func TestFileCheckpointStore_SaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

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
	err = store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Check if file exists
	filePath := filepath.Join(tmpDir, runID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("checkpoint file not created: %v", err)
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

func TestFileCheckpointStore_SaveUpdate(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

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
	err = store.Save(ctx, cp)
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

func TestFileCheckpointStore_LoadNonExistent(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

	ctx := context.Background()

	nonExistentID := uuid.New().String()
	_, err = store.Load(ctx, nonExistentID)
	if err == nil {
		t.Error("expected error when loading non-existent checkpoint")
	}
}

func TestFileCheckpointStore_Delete(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

	ctx := context.Background()

	// Create a checkpoint
	runID := uuid.New().String()
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
	}

	// Save checkpoint
	err = store.Save(ctx, cp)
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Check if file exists
	filePath := filepath.Join(tmpDir, runID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("checkpoint file not created: %v", err)
	}

	// Delete checkpoint
	err = store.Delete(ctx, runID)
	if err != nil {
		t.Fatalf("failed to delete checkpoint: %v", err)
	}

	// Check if file was deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("checkpoint file should have been deleted")
	}

	// Try to load deleted checkpoint
	_, err = store.Load(ctx, runID)
	if err == nil {
		t.Error("expected error when loading deleted checkpoint")
	}
}

func TestFileCheckpointStore_List(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

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

	// List checkpoints
	checkpoints, err := store.List(ctx)
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 5 {
		t.Errorf("expected 5 checkpoints, got %d", len(checkpoints))
	}
}

func TestFileCheckpointStore_ListWithOptions(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

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

func TestFileCheckpointStore_ListWithPagination(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

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

func TestFileCheckpointStore_HealthCheck(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

	ctx := context.Background()

	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestFileCheckpointStore_Close(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file-checkpoint-store")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	store, err := NewFileCheckpointStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileCheckpointStore: %v", err)
	}

	ctx := context.Background()

	if err := store.Close(ctx); err != nil {
		t.Errorf("close failed: %v", err)
	}
}
