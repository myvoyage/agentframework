// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// Thread Store Options Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"testing"
	"time"
)

func TestMemoryThreadStore_ApplyThreadOptions_MaxMessages(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		MaxMessages: 3,
	})

	ctx := context.Background()
	thread, _ := store.Create(ctx)

	// Add more messages than the limit
	for i := 0; i < 10; i++ {
		thread.Messages = append(thread.Messages, Message{
			Role:      "user",
			Content:   "message",
			Timestamp: time.Now().Unix(),
		})
	}

	// Save the thread - this should apply the options
	err := store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Get the thread back
	retrieved, _ := store.Get(ctx, thread.ID)

	// Should only have 3 messages (the most recent ones)
	if len(retrieved.Messages) != 3 {
		t.Errorf("expected 3 messages after applying MaxMessages limit, got %d", len(retrieved.Messages))
	}
}

func TestMemoryThreadStore_ApplyThreadOptions_MaxMessageSize(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		MaxMessageSize: 10,
	})

	ctx := context.Background()
	thread, _ := store.Create(ctx)

	// Add messages with content longer than the limit
	thread.Messages = []Message{
		{
			Role:      "user",
			Content:   "this is a very long message that exceeds the limit",
			Timestamp: time.Now().Unix(),
		},
		{
			Role:      "assistant",
			Content:   "another long message that should be truncated",
			Timestamp: time.Now().Unix(),
		},
	}

	// Save the thread - this should apply the options
	err := store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Get the thread back
	retrieved, _ := store.Get(ctx, thread.ID)

	// Check that messages were truncated
	for i, msg := range retrieved.Messages {
		if len(msg.Content) > 10 {
			t.Errorf("message %d was not truncated, got length %d", i, len(msg.Content))
		}
	}
}

func TestMemoryThreadStore_ApplyThreadOptions_Combined(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		MaxMessages:    5,
		MaxMessageSize: 20,
	})

	ctx := context.Background()
	thread, _ := store.Create(ctx)

	// Add many long messages
	for i := 0; i < 10; i++ {
		thread.Messages = append(thread.Messages, Message{
			Role:      "user",
			Content:   "this is a long message that should be truncated to 20 characters",
			Timestamp: time.Now().Unix(),
		})
	}

	// Save the thread
	err := store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Get the thread back
	retrieved, _ := store.Get(ctx, thread.ID)

	// Check both limits were applied
	if len(retrieved.Messages) != 5 {
		t.Errorf("expected 5 messages after applying MaxMessages limit, got %d", len(retrieved.Messages))
	}

	for i, msg := range retrieved.Messages {
		if len(msg.Content) > 20 {
			t.Errorf("message %d was not truncated, got length %d", i, len(msg.Content))
		}
	}
}

func TestMemoryThreadStore_ApplyThreadOptions_NoLimits(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		MaxMessages:    0, // No limit
		MaxMessageSize: 0, // No limit
	})

	ctx := context.Background()
	thread, _ := store.Create(ctx)

	// Add many long messages
	longContent := "this is a very long message that should not be truncated since there is no limit"
	for i := 0; i < 20; i++ {
		thread.Messages = append(thread.Messages, Message{
			Role:      "user",
			Content:   longContent,
			Timestamp: time.Now().Unix(),
		})
	}

	// Save the thread
	err := store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Get the thread back
	retrieved, _ := store.Get(ctx, thread.ID)

	// Check no limits were applied
	if len(retrieved.Messages) != 20 {
		t.Errorf("expected 20 messages when no limit is set, got %d", len(retrieved.Messages))
	}

	for i, msg := range retrieved.Messages {
		if msg.Content != longContent {
			t.Errorf("message %d was modified when no limit was set", i)
		}
	}
}

func TestThreadMetadata(t *testing.T) {
	now := time.Now()
	metadata := &ThreadMetadata{
		Thread: &Thread{
			ID:        "test-id",
			Messages:  []Message{{Role: "user", Content: "test"}},
			CreatedAt: now.Unix(),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if metadata.Thread.ID != "test-id" {
		t.Errorf("expected thread ID 'test-id', got '%s'", metadata.Thread.ID)
	}
	if metadata.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if metadata.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNewMemoryThreadStoreWithOptions(t *testing.T) {
	opts := ThreadOptions{
		MaxMessages:    50,
		MaxMessageSize: 1000,
		TTL:            time.Hour,
	}

	store := NewMemoryThreadStoreWithOptions(opts)

	if store.Options.MaxMessages != 50 {
		t.Errorf("expected MaxMessages 50, got %d", store.Options.MaxMessages)
	}
	if store.Options.MaxMessageSize != 1000 {
		t.Errorf("expected MaxMessageSize 1000, got %d", store.Options.MaxMessageSize)
	}
	if store.Options.TTL != time.Hour {
		t.Errorf("expected TTL 1 hour, got %v", store.Options.TTL)
	}
	if store.threads == nil {
		t.Error("expected threads map to be initialized")
	}
}

func TestDefaultThreadOptions(t *testing.T) {
	store := NewMemoryThreadStore()

	// Check default values
	if store.Options.MaxMessages <= 0 {
		t.Error("expected default MaxMessages to be set")
	}
	if store.Options.MaxMessageSize <= 0 {
		t.Error("expected default MaxMessageSize to be set")
	}
	if store.Options.TTL <= 0 {
		t.Error("expected default TTL to be set")
	}
}
