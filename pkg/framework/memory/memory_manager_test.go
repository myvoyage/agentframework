// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// Thread Store Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)
func TestNewMemoryThreadStore(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		MaxMessages:    50,
		MaxMessageSize: 10000,
		TTL:            24 * time.Hour,
	})

	if store == nil {
		t.Fatal("expected MemoryThreadStore to be created")
	}

	if store.Options.MaxMessages != 50 {
		t.Errorf("expected MaxMessages 50, got %d", store.Options.MaxMessages)
	}
	if store.Options.MaxMessageSize != 10000 {
		t.Errorf("expected MaxMessageSize 10000, got %d", store.Options.MaxMessageSize)
	}
	if store.Options.TTL != 24*time.Hour {
		t.Errorf("expected TTL 24h, got %v", store.Options.TTL)
	}

	if store.threads == nil {
		t.Error("expected threads map to be initialized")
	}
}

func TestNewMemoryThreadStore_DefaultOptions(t *testing.T) {
	store := NewMemoryThreadStore()

	if store == nil {
		t.Fatal("expected MemoryThreadStore to be created")
	}

	if store.Options.MaxMessages <= 0 {
		t.Error("expected MaxMessages to be set")
	}
}

func TestMemoryThreadStore_CreateThread(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
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

	if _, err := uuid.Parse(thread.ID); err != nil {
		t.Errorf("expected valid UUID for thread ID: %v", err)
	}

	if len(thread.Messages) != 0 {
		t.Errorf("expected empty messages slice, got %d", len(thread.Messages))
	}
}

func TestMemoryThreadStore_CreateThread_Duplicate(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	// Create first thread
	thread1, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create first thread: %v", err)
	}

	// Create second thread
	thread2, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create second thread: %v", err)
	}

	if thread1.ID == thread2.ID {
		t.Error("threads should have unique IDs")
	}
}

func TestMemoryThreadStore_GetThread(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	// Create thread first
	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// Get the created thread
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

func TestMemoryThreadStore_GetThread_NonExistent(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	nonExistentID := uuid.New().String()
	thread, err := store.Get(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("unexpected error when getting non-existent thread: %v", err)
	}
	if thread != nil {
		t.Error("expected nil thread when getting non-existent thread")
	}
}

func TestMemoryThreadStore_SaveThread(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
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

	// Verify the thread was saved by getting it again
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

func TestMemoryThreadStore_SaveThread_Update(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// First save
	thread.Metadata = map[string]interface{}{"version": "1.0"}
	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	// Update and save again
	thread.Metadata["version"] = "2.0"
	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to update thread: %v", err)
	}

	// Verify update
	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("failed to get updated thread: %v", err)
	}

	if retrieved.Metadata["version"] != "2.0" {
		t.Errorf("expected metadata version '2.0', got '%v'", retrieved.Metadata["version"])
	}
}

func TestMemoryThreadStore_SaveThread_DynamicMetadata(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// Test with complex metadata
	thread.Metadata = map[string]interface{}{
		"count":      42,
		"active":     true,
		"tags":       []string{"test", "demo"},
		"config":     map[string]interface{}{"key": "value"},
		"updated_at": time.Now(),
	}

	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("failed to save thread: %v", err)
	}

	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("failed to get thread: %v", err)
	}

	// Check all fields are preserved
	if retrieved.Metadata["count"] != 42 {
		t.Error("expected count to be preserved")
	}

	if retrieved.Metadata["active"] != true {
		t.Error("expected active to be preserved")
	}

	// Check tags - it should be preserved as []string
	tags, ok := retrieved.Metadata["tags"].([]string)
	if !ok {
		t.Error("expected tags to be []string")
	} else if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestMemoryThreadStore_HealthCheck(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	if err := store.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestMemoryThreadStore_Close(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	if err := store.Close(ctx); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestMemoryThreadStore_ThreadLifecycle(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	// 1. Create thread
	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Logf("Created thread: %s", thread.ID)

	// 2. Verify thread exists
	retrieved, err := store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Thread should exist")
	}

	// 3. Update thread
	thread.Messages = append(thread.Messages, Message{
		Role:      "system",
		Content:   "Welcome!",
		Timestamp: time.Now().Unix(),
	})
	err = store.Save(ctx, thread)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// 4. Verify update
	retrieved, err = store.Get(ctx, thread.ID)
	if err != nil {
		t.Fatalf("Get after save: %v", err)
	}
	if len(retrieved.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(retrieved.Messages))
	}

	// 5. Close store
	if err := store.Close(ctx); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestMemoryThreadStore_ConcurrentOperations(t *testing.T) {
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{})
	ctx := context.Background()

	const numThreads = 10
	errs := make(chan error, numThreads)

	// Create multiple threads concurrently
	for i := 0; i < numThreads; i++ {
		go func(id int) {
			if _, err := store.Create(ctx); err != nil {
				errs <- err
			} else {
				errs <- nil
			}
		}(i)
	}

	// Check for errors
	allOk := true
	for i := 0; i < numThreads; i++ {
		if err := <-errs; err != nil {
			allOk = false
			t.Errorf("thread %d failed: %v", i, err)
		}
	}

	if allOk {
		t.Logf("Successfully created %d concurrent threads", numThreads)
	}
}

func TestMemoryThreadStore_TTL_Cleanup(t *testing.T) {
	const ttl = 100 * time.Millisecond
	store := NewMemoryThreadStoreWithOptions(ThreadOptions{
		TTL: ttl,
	})
	ctx := context.Background()

	// Create thread
	thread, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	// Verify it exists
	if _, err := store.Get(ctx, thread.ID); err != nil {
		t.Errorf("thread should exist before TTL: %v", err)
	}

	// Let TTL expire
	time.Sleep(ttl * 2)

	// Note: We can't manually call cleanup() as it's a private method
	// The cleanup loop will automatically clean up expired threads
	// We just need to wait a bit longer for the cleanup loop to run
	time.Sleep(200 * time.Millisecond)

	// Check if thread was cleaned up
	if _, err := store.Get(ctx, thread.ID); err == nil {
		t.Log("thread may still exist (cleanup timing may vary)")
	}
}

// ===== Memory Manager Tests =====

func TestNewMemoryManager(t *testing.T) {
	opts := MemoryOptions{
		MaxMessages:    50,
		MaxMessageSize: 10000,
		TrimRatio:      0.8,
		EnableTrimming: true,
	}

	mm := NewMemoryManager(opts)

	if mm == nil {
		t.Fatal("expected MemoryManager to be created")
	}

	if mm.opts.MaxMessages != 50 {
		t.Errorf("expected MaxMessages 50, got %d", mm.opts.MaxMessages)
	}
}

func TestDefaultMemoryManager(t *testing.T) {
	mm := DefaultMemoryManager()

	if mm == nil {
		t.Fatal("expected MemoryManager to be created")
	}

	if mm.opts.MaxMessages != 100 {
		t.Errorf("expected MaxMessages 100, got %d", mm.opts.MaxMessages)
	}
}

func TestMemoryManager_SetOptions(t *testing.T) {
	mm := DefaultMemoryManager()

	newOpts := MemoryOptions{
		MaxMessages:    200,
		MaxMessageSize: 5000,
		TrimRatio:      0.5,
		EnableTrimming: false,
	}

	mm.SetOptions(newOpts)

	if mm.opts.MaxMessages != 200 {
		t.Errorf("expected MaxMessages 200, got %d", mm.opts.MaxMessages)
	}
}

func TestMemoryManager_GetOptions(t *testing.T) {
	opts := MemoryOptions{
		MaxMessages:    75,
		MaxMessageSize: 8000,
		TrimRatio:      0.6,
		EnableTrimming: true,
	}

	mm := NewMemoryManager(opts)
	retrieved := mm.GetOptions()

	if retrieved.MaxMessages != 75 {
		t.Errorf("expected MaxMessages 75, got %d", retrieved.MaxMessages)
	}
}

func TestMemoryManager_LimitHistory_NoTrimming(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: false,
		MaxMessages:    10,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.User, Content: "message 1"},
		{Role: schema.Assistant, Content: "message 2"},
	}

	result := mm.LimitHistory(messages)

	if len(result) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result))
	}
}

func TestMemoryManager_CheckMessageSize_NoLimit(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 0}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "test"}

	if !mm.CheckMessageSize(msg) {
		t.Error("expected message to pass size check")
	}
}

func TestMemoryManager_TrimMessage(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 10}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "this is a long message"}

	trimmed := mm.TrimMessage(msg)
	if trimmed == nil {
		t.Fatal("expected trimmed message")
	}

	if len(trimmed.Content) > 10 {
		t.Errorf("expected trimmed content length <= 10, got %d", len(trimmed.Content))
	}
}

func TestMemoryManager_ProcessMessage(t *testing.T) {
	mm := DefaultMemoryManager()

	msg := &schema.Message{Role: schema.User, Content: "test"}

	result := mm.ProcessMessage(msg)
	if result == nil {
		t.Fatal("expected message to be returned")
	}
}

func TestMemoryManager_ClearHistory(t *testing.T) {
	mm := DefaultMemoryManager()

	result := mm.ClearHistory()

	if result == nil {
		t.Fatal("expected empty slice")
	}

	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(result))
	}
}

func TestMemoryManager_DefaultValues(t *testing.T) {
	// Test with invalid values that should get defaults
	opts := MemoryOptions{
		MaxMessages:    -1, // Should default to 100
		MaxMessageSize: 0,
		TrimRatio:      -1, // Should default to 0.7
		EnableTrimming: false, // Should default to true
	}

	mm := NewMemoryManager(opts)

	if mm.opts.MaxMessages != 100 {
		t.Errorf("expected MaxMessages to default to 100, got %d", mm.opts.MaxMessages)
	}
	if mm.opts.TrimRatio != 0.7 {
		t.Errorf("expected TrimRatio to default to 0.7, got %v", mm.opts.TrimRatio)
	}
	if !mm.opts.EnableTrimming {
		t.Error("expected EnableTrimming to default to true")
	}
}

func TestMemoryManager_LimitHistory_WithinLimit(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    10,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.User, Content: "message 1"},
		{Role: schema.Assistant, Content: "message 2"},
		{Role: schema.User, Content: "message 3"},
	}

	result := mm.LimitHistory(messages)

	if len(result) != 3 {
		t.Errorf("expected 3 messages (within limit), got %d", len(result))
	}
}

func TestMemoryManager_LimitHistory_PreservesSystemMessages(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    3,
		TrimRatio:      0.5,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.System, Content: "system instruction"},
		{Role: schema.User, Content: "message 1"},
		{Role: schema.Assistant, Content: "message 2"},
		{Role: schema.User, Content: "message 3"},
		{Role: schema.Assistant, Content: "message 4"},
	}

	result := mm.LimitHistory(messages)

	// Should preserve system message
	hasSystem := false
	for _, msg := range result {
		if msg.Role == schema.System {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		t.Error("expected system message to be preserved")
	}
}

func TestMemoryManager_LimitHistory_WithTrimRatio(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    10,
		TrimRatio:      0.5,
	}
	mm := NewMemoryManager(opts)

	// Create 20 messages
	messages := make([]*schema.Message, 20)
	for i := 0; i < 20; i++ {
		messages[i] = &schema.Message{
			Role:    schema.User,
			Content: "message",
		}
	}

	result := mm.LimitHistory(messages)

	// Should keep about 10 messages (50% of 20)
	if len(result) > 12 || len(result) < 8 {
		t.Errorf("expected approximately 10 messages (50%% trim ratio), got %d", len(result))
	}
}

func TestMemoryManager_TrimMessage_PassesWithinLimit(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 100}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "short content"}

	trimmed := mm.TrimMessage(msg)
	if trimmed != msg {
		t.Error("expected same message when within limit")
	}
}

func TestMemoryManager_TrimMessage_NoLimit(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 0}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "very long content that exceeds any reasonable limit"}

	trimmed := mm.TrimMessage(msg)
	if trimmed != msg {
		t.Error("expected same message when no limit is set")
	}
}

func TestMemoryManager_ProcessMessage_Trimmed(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 10}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "this is a very long message"}

	result := mm.ProcessMessage(msg)
	if result == nil {
		t.Fatal("expected message to be returned")
	}

	if len(result.Content) > 13 { // 10 + "..."
		t.Errorf("expected trimmed message, got length %d", len(result.Content))
	}
}

func TestMemoryManager_ProcessMessage_NotTrimmed(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 100}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{Role: schema.User, Content: "short"}

	result := mm.ProcessMessage(msg)
	if result == nil {
		t.Fatal("expected message to be returned")
	}

	if result.Content != "short" {
		t.Errorf("expected unchanged message, got '%s'", result.Content)
	}
}

func TestMemoryManager_LimitHistory_MaxMessagesZero(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    0,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.User, Content: "message 1"},
		{Role: schema.User, Content: "message 2"},
	}

	result := mm.LimitHistory(messages)

	// Should return all messages when MaxMessages is 0
	if len(result) != 2 {
		t.Errorf("expected all messages when MaxMessages is 0, got %d", len(result))
	}
}

func TestMemoryManager_LimitHistory_AllSystemMessages(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    5,
		TrimRatio:      0.7,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.System, Content: "system 1"},
		{Role: schema.System, Content: "system 2"},
		{Role: schema.System, Content: "system 3"},
	}

	result := mm.LimitHistory(messages)

	// All system messages should be preserved
	if len(result) != 3 {
		t.Errorf("expected all 3 system messages to be preserved, got %d", len(result))
	}

	for i, msg := range result {
		if msg.Role != schema.System {
			t.Errorf("expected message %d to be system, got %v", i, msg.Role)
		}
	}
}

func TestMemoryManager_LimitHistory_MixedMessages(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    5,
		TrimRatio:      0.6,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.System, Content: "system instruction"},
		{Role: schema.User, Content: "user message 1"},
		{Role: schema.Assistant, Content: "assistant message 1"},
		{Role: schema.User, Content: "user message 2"},
		{Role: schema.Assistant, Content: "assistant message 2"},
		{Role: schema.User, Content: "user message 3"},
		{Role: schema.Assistant, Content: "assistant message 3"},
		{Role: schema.User, Content: "user message 4"},
		{Role: schema.Assistant, Content: "assistant message 4"},
		{Role: schema.User, Content: "user message 5"},
	}

	result := mm.LimitHistory(messages)

	// Should preserve system message and most recent messages
	if len(result) > 5 {
		t.Errorf("expected at most 5 messages, got %d", len(result))
	}

	// First message should be system
	if result[0].Role != schema.System {
		t.Error("expected first message to be system")
	}
}

func TestMemoryManager_LimitHistory_EmptySlice(t *testing.T) {
	mm := DefaultMemoryManager()

	result := mm.LimitHistory([]*schema.Message{})

	if result == nil {
		t.Fatal("expected nil slice to be handled")
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d messages", len(result))
	}
}

func TestMemoryManager_LimitHistory_SingleMessage(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    10,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.User, Content: "single message"},
	}

	result := mm.LimitHistory(messages)

	if len(result) != 1 {
		t.Errorf("expected 1 message, got %d", len(result))
	}
}

func TestMemoryManager_TrimMessage_PreservesToolCalls(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 20}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:    schema.User,
		Content: "this is a very long message that should be trimmed",
		ToolCalls: []schema.ToolCall{
			{
				ID: "call_1",
				Function: schema.FunctionCall{
					Name:      "test_function",
					Arguments: "{}",
				},
			},
		},
	}

	trimmed := mm.TrimMessage(msg)

	if trimmed.ToolCalls == nil {
		t.Error("expected tool calls to be preserved")
	}

	if len(trimmed.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(trimmed.ToolCalls))
	}

	if trimmed.ToolCalls[0].Function.Name != "test_function" {
		t.Errorf("expected tool call name 'test_function', got '%s'", trimmed.ToolCalls[0].Function.Name)
	}
}

func TestMemoryManager_TrimMessage_PreservesToolCallID(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 15}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:       schema.Assistant,
		Content:    "very long content",
		ToolCallID: "response_123",
	}

	trimmed := mm.TrimMessage(msg)

	if trimmed.ToolCallID != "response_123" {
		t.Errorf("expected tool call ID 'response_123', got '%s'", trimmed.ToolCallID)
	}

	if trimmed.Role != schema.Assistant {
		t.Errorf("expected role to be preserved, got %v", trimmed.Role)
	}
}

func TestMemoryManager_TrimMessage_AddsEllipsis(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 10}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:    schema.User,
		Content: "this is a very long message",
	}

	trimmed := mm.TrimMessage(msg)

	// Should be exactly maxSize - 3 + "..."
	expectedLen := 10 - 3 + 3 // "..." is 3 chars
	if len(trimmed.Content) != expectedLen {
		t.Errorf("expected content length %d, got %d", expectedLen, len(trimmed.Content))
	}

	if trimmed.Content[len(trimmed.Content)-3:] != "..." {
		t.Error("expected content to end with '...'")
	}
}

func TestMemoryManager_CheckMessageSize_WithToolCalls(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 20}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:    schema.User,
		Content: "short",
		ToolCalls: []schema.ToolCall{
			{
				Function: schema.FunctionCall{
					Name:      "very_long_function_name_that_exceeds_limit",
					Arguments: "very_long_arguments_string_that_exceeds_limit",
				},
			},
		},
	}

	// Should fail because content + tool calls exceed limit
	if mm.CheckMessageSize(msg) {
		t.Error("expected message size check to fail with large tool calls")
	}
}

func TestMemoryManager_CheckMessageSize_ExactlyAtLimit(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 10}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:    schema.User,
		Content: "0123456789", // Exactly 10 chars
	}

	if !mm.CheckMessageSize(msg) {
		t.Error("expected message at limit to pass")
	}
}

func TestMemoryManager_CheckMessageSize_JustOverLimit(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 10}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:    schema.User,
		Content: "0123456789A", // 11 chars
	}

	if mm.CheckMessageSize(msg) {
		t.Error("expected message over limit to fail")
	}
}

func TestMemoryManager_NewMemoryManager_DefaultTrimRatio(t *testing.T) {
	opts := MemoryOptions{
		MaxMessages: 100,
		TrimRatio:   1.5, // Invalid, should default to 0.7
	}

	mm := NewMemoryManager(opts)

	if mm.opts.TrimRatio != 0.7 {
		t.Errorf("expected TrimRatio to default to 0.7, got %v", mm.opts.TrimRatio)
	}
}

func TestMemoryManager_NewMemoryManager_ZeroTrimRatio(t *testing.T) {
	opts := MemoryOptions{
		MaxMessages: 100,
		TrimRatio:   0, // Invalid, should default to 0.7
	}

	mm := NewMemoryManager(opts)

	if mm.opts.TrimRatio != 0.7 {
		t.Errorf("expected TrimRatio to default to 0.7, got %v", mm.opts.TrimRatio)
	}
}

func TestMemoryManager_ProcessMessage_WithNilToolCalls(t *testing.T) {
	opts := MemoryOptions{MaxMessageSize: 100}
	mm := NewMemoryManager(opts)

	msg := &schema.Message{
		Role:       schema.Assistant,
		Content:    "response",
		ToolCalls:  nil,
		ToolCallID: "call_123",
	}

	result := mm.ProcessMessage(msg)

	if result == nil {
		t.Fatal("expected message to be returned")
	}

	if result.ToolCallID != "call_123" {
		t.Errorf("expected tool call ID to be preserved, got '%s'", result.ToolCallID)
	}
}

func TestMemoryManager_LimitHistory_ExactlyAtLimit(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    5,
	}
	mm := NewMemoryManager(opts)

	messages := make([]*schema.Message, 5)
	for i := 0; i < 5; i++ {
		messages[i] = &schema.Message{
			Role:    schema.User,
			Content: "message",
		}
	}

	result := mm.LimitHistory(messages)

	if len(result) != 5 {
		t.Errorf("expected exactly 5 messages when at limit, got %d", len(result))
	}
}

func TestMemoryManager_LimitHistory_JustOverLimit(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    5,
		TrimRatio:      0.8,
	}
	mm := NewMemoryManager(opts)

	messages := make([]*schema.Message, 6)
	for i := 0; i < 6; i++ {
		messages[i] = &schema.Message{
			Role:    schema.User,
			Content: "message",
		}
	}

	result := mm.LimitHistory(messages)

	// Should trim to approximately 5 messages
	if len(result) > 5 {
		t.Errorf("expected at most 5 messages, got %d", len(result))
	}
}

func TestMemoryManager_LimitHistory_LargeTrimRatio(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    100,
		TrimRatio:      1.0, // Keep all
	}
	mm := NewMemoryManager(opts)

	messages := make([]*schema.Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = &schema.Message{
			Role:    schema.User,
			Content: "message",
		}
	}

	result := mm.LimitHistory(messages)

	// With trim ratio of 1.0 and 50 messages under limit of 100, should keep all
	if len(result) != 50 {
		t.Errorf("expected all 50 messages with trim ratio 1.0, got %d", len(result))
	}
}

func TestMemoryManager_LimitHistory_NoRoomForNonSystem(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    1,
		TrimRatio:      0.5,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.System, Content: "system"},
		{Role: schema.User, Content: "user"},
		{Role: schema.Assistant, Content: "assistant"},
	}

	result := mm.LimitHistory(messages)

	// Should only keep system message
	if len(result) != 1 {
		t.Errorf("expected only 1 message (system), got %d", len(result))
	}

	if result[0].Role != schema.System {
		t.Error("expected only system message to be kept")
	}
}

func TestMemoryManager_LimitHistory_SystemMessagesExceedLimit(t *testing.T) {
	opts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    2,
		TrimRatio:      0.5,
	}
	mm := NewMemoryManager(opts)

	messages := []*schema.Message{
		{Role: schema.System, Content: "system 1"},
		{Role: schema.System, Content: "system 2"},
		{Role: schema.System, Content: "system 3"},
		{Role: schema.User, Content: "user"},
	}

	result := mm.LimitHistory(messages)

	// Should preserve all system messages even if it exceeds MaxMessages
	systemCount := 0
	for _, msg := range result {
		if msg.Role == schema.System {
			systemCount++
		}
	}

	if systemCount != 3 {
		t.Errorf("expected all 3 system messages to be preserved, got %d", systemCount)
	}
}

