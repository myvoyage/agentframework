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
	"fmt"
	"testing"
	"time"
)

func TestModelCache(t *testing.T) {
	ctx := context.Background()

	// Create a model cache with short TTL for testing
	cache := NewModelCache(ModelCacheConfig{
		MaxSize:         5,
		TTL:             5 * time.Second,
		CleanupInterval: 1 * time.Second,
	})
	defer cache.StopCleanup()

	// Track call count
	var callCount int

	// Create a model factory
	factory := func(ctx context.Context, modelName string) (ChatModel, error) {
		callCount++
		return &MockChatModel{}, nil
	}

	cachedFactory := CachedModelFactory(factory, cache)

	// First call - should create new model
	model1, err := cachedFactory(ctx, "test-model")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model1 == nil {
		t.Error("Expected model, got nil")
	}
	if callCount != 1 {
		t.Errorf("Expected callCount 1, got %d", callCount)
	}

	// Second call - should return from cache
	model2, err := cachedFactory(ctx, "test-model")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model2 == nil {
		t.Error("Expected model, got nil")
	}
	if callCount != 1 {
		t.Errorf("Expected callCount 1 (cached), got %d", callCount)
	}

	// Verify same instance
	if model1 != model2 {
		t.Error("Expected same model instance, got different instances")
	}

	// Test cache expiration
	time.Sleep(6 * time.Second) // Wait for TTL to expire

	// Next call - should create new model again
	model3, err := cachedFactory(ctx, "test-model")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if model3 == nil {
		t.Error("Expected model, got nil")
	}
	if callCount != 2 {
		t.Errorf("Expected callCount 2 (expired), got %d", callCount)
	}

	// Verify different instance
	if model1 == model3 {
		t.Error("Expected different model instance after expiration, got same instance")
	}
}

func TestModelCacheCleanup(t *testing.T) {
	// Create a model cache with short TTL and cleanup interval
	cache := NewModelCache(ModelCacheConfig{
		MaxSize:         5,
		TTL:             2 * time.Second,
		CleanupInterval: 1 * time.Second,
	})
	defer cache.StopCleanup()

	// Add multiple models to cache
	for i := 0; i < 3; i++ {
		modelName := fmt.Sprintf("test-model-%d", i)
		cache.Put(modelName, &MockChatModel{NameValue: modelName})
	}

	// Verify initial cache size
	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Wait for cleanup to run
	time.Sleep(3 * time.Second)

	// Verify cache has been cleaned up
	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after cleanup, got %d", cache.Len())
	}
}
