// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// File Thread Store Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileThreadStore(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewFileThreadStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileThreadStore: %v", err)
	}
	if store == nil {
		t.Fatal("expected store to be created")
	}
	if store.dir != tmpDir {
		t.Errorf("expected dir '%s', got '%s'", tmpDir, store.dir)
	}
}

func TestFileThreadStore_CreateThread(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	if thread == nil {
		t.Fatal("expected thread to be created")
	}

	if thread.ID == "" {
		t.Error("expected thread ID to be non-empty")
	}

	if len(thread.Messages) != 0 {
		t.Errorf("expected empty messages slice, got %d", len(thread.Messages))
	}

	// Verify file was created
	path := filepath.Join(tmpDir, thread.ID+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected thread file to be created")
	}
}

func TestFileThreadStore_GetThread(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	// Create thread
	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// Get the thread
	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("failed to get thread: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected thread to be retrieved")
	}

	if retrieved.ID != thread.ID {
		t.Errorf("expected thread ID %s, got %s", thread.ID, retrieved.ID)
	}
}

func TestFileThreadStore_GetThread_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	thread, err := store.Get(ctx, "non-existent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if thread != nil {
		t.Error("expected nil thread for non-existent ID")
	}
}

func TestFileThreadStore_SaveThread(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// Modify the thread
	thread.Messages = append(thread.Messages, Message{
		Role:      "user",
		Content:   "Hello world",
		Timestamp: time.Now().Unix(),
	})
	thread.Metadata = map[string]interface{}{
		"test": "value",
	}

	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Verify by getting it again
	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("failed to get saved thread: %v", err)
	}

	if len(retrieved.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(retrieved.Messages))
	}

	if retrieved.Metadata == nil || retrieved.Metadata["test"] != "value" {
		t.Error("expected metadata to be saved")
	}
}

func TestFileThreadStore_SaveThread_Nil(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	// Save nil thread - should not error
	err := store.Save(ctx, nil)
	if err != nil {
		t.Errorf("unexpected error saving nil thread: %v", err)
	}
}

func TestFileThreadStore_SaveThread_EmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	thread := &Thread{ID: ""}
	err := store.Save(ctx, thread)
	if err != nil {
		t.Errorf("unexpected error saving thread with empty ID: %v", err)
	}
}

func TestFileThreadStore_HealthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestFileThreadStore_Close(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	if err := store.Close(ctx); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestFileThreadStore_ThreadLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	// Create
	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get
	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if retrieved == nil {
		t.Fatal("thread should exist")
	}

	// Update
	thread.Messages = append(thread.Messages, Message{
		Role:      "system",
		Content:   "Welcome!",
		Timestamp: time.Now().Unix(),
	})
	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify update
	retrieved, err = store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("Get after save: %v", err)
	}
	if len(retrieved.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(retrieved.Messages))
	}

	// Close
	if err := store.Close(ctx); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestFileThreadStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a thread and save it
	store1, _ := NewFileThreadStore(tmpDir)
	ctx := context.Background()

	thread, err := store1.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}
	threadID := thread.ID

	// Close first store
	if err := store1.Close(ctx); err != nil {
		t.Errorf("failed to close first store: %v", err)
	}

	// Create a new store with the same directory
	store2, _ := NewFileThreadStore(tmpDir)
	retrieved, err := store2.Get(ctx, threadID)
	if err != nil {
		t.Fatalf("failed to get thread from new store: %v", err)
	}
	if retrieved == nil {
		t.Fatal("thread should persist across store instances")
	}
	if retrieved.ID != threadID {
		t.Errorf("expected thread ID %s, got %s", threadID, retrieved.ID)
	}
}
