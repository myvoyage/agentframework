// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package agent

import (
	"context"
	"os"
	"testing"
)

func TestMemoryThreadStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryThreadStore()

	// Test Create
	thread, err := store.Create(ctx)
	if err != nil {
		t.Errorf("Failed to create thread: %v", err)
	}
	if thread == nil || thread.ID == "" {
		t.Errorf("Expected valid thread, got %v", thread)
	}

	// Test Get
	retrievedThread, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Errorf("Failed to get thread: %v", err)
	}
	if retrievedThread.ID != thread.ID {
		t.Errorf("Expected thread ID '%s', got '%s'", thread.ID, retrievedThread.ID)
	}

	// Test HealthCheck
	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test Close
	if err := store.Close(ctx); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestFileThreadStore(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "thread_store_test")
	if err != nil {
		t.Errorf("Failed to create temp directory: %v", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Create FileThreadStore
	store, err := NewFileThreadStore(tempDir)
	if err != nil {
		t.Errorf("Failed to create FileThreadStore: %v", err)
		return
	}

	// Test Create
	thread, err := store.Create(ctx)
	if err != nil {
		t.Errorf("Failed to create thread: %v", err)
		return
	}

	// Test Get
	retrievedThread, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Errorf("Failed to get thread: %v", err)
		return
	}
	if retrievedThread.ID != thread.ID {
		t.Errorf("Expected thread ID '%s', got '%s'", thread.ID, retrievedThread.ID)
	}

	// Test HealthCheck
	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test Close
	if err := store.Close(ctx); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
