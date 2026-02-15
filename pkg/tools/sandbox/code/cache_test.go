// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"strings"
	"testing"
)

// TestLRUCache_BasicOperations tests basic cache operations
func TestLRUCache_BasicOperations(t *testing.T) {
	cache := NewLRUCache(3)

	// Test Put and Get
	result1 := &ExecutionResult{Success: true, Output: "test1"}
	cache.Put("key1", result1)

	got, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected cache hit, got miss")
	}
	if got.Output != "test1" {
		t.Errorf("Expected output 'test1', got '%s'", got.Output)
	}

	// Test cache miss
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected cache miss, got hit")
	}
}

// TestLRUCache_LRUEviction tests LRU eviction policy
func TestLRUCache_LRUEviction(t *testing.T) {
	cache := NewLRUCache(3)

	// Fill cache to capacity
	cache.Put("key1", &ExecutionResult{Success: true, Output: "test1"})
	cache.Put("key2", &ExecutionResult{Success: true, Output: "test2"})
	cache.Put("key3", &ExecutionResult{Success: true, Output: "test3"})

	// Verify all keys are present
	if _, ok := cache.Get("key1"); !ok {
		t.Error("key1 should be in cache")
	}
	if _, ok := cache.Get("key2"); !ok {
		t.Error("key2 should be in cache")
	}
	if _, ok := cache.Get("key3"); !ok {
		t.Error("key3 should be in cache")
	}

	// Add one more item, should evict key1 (least recently used)
	cache.Put("key4", &ExecutionResult{Success: true, Output: "test4"})

	// key1 should be evicted
	if _, ok := cache.Get("key1"); ok {
		t.Error("key1 should be evicted")
	}

	// Other keys should still be present
	if _, ok := cache.Get("key2"); !ok {
		t.Error("key2 should still be in cache")
	}
	if _, ok := cache.Get("key3"); !ok {
		t.Error("key3 should still be in cache")
	}
	if _, ok := cache.Get("key4"); !ok {
		t.Error("key4 should be in cache")
	}

	// Check eviction count
	stats := cache.GetStats()
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
}

// TestLRUCache_AccessOrder tests that access updates LRU order
func TestLRUCache_AccessOrder(t *testing.T) {
	cache := NewLRUCache(3)

	// Fill cache
	cache.Put("key1", &ExecutionResult{Success: true, Output: "test1"})
	cache.Put("key2", &ExecutionResult{Success: true, Output: "test2"})
	cache.Put("key3", &ExecutionResult{Success: true, Output: "test3"})

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add new item, should evict key2 (now least recently used)
	cache.Put("key4", &ExecutionResult{Success: true, Output: "test4"})

	// key2 should be evicted
	if _, ok := cache.Get("key2"); ok {
		t.Error("key2 should be evicted")
	}

	// key1 should still be present (was accessed recently)
	if _, ok := cache.Get("key1"); !ok {
		t.Error("key1 should still be in cache")
	}
}

// TestLRUCache_Stats tests cache statistics
func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(3)

	// Initial stats
	stats := cache.GetStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Initial stats should be zero")
	}

	// Add items
	cache.Put("key1", &ExecutionResult{Success: true, Output: "test1"})
	cache.Put("key2", &ExecutionResult{Success: true, Output: "test2"})

	// Cache hits
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key2")

	// Cache misses
	cache.Get("key3")
	cache.Get("key4")

	stats = cache.GetStats()
	if stats.Hits != 3 {
		t.Errorf("Expected 3 hits, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}

	// Check hit rate
	expectedHitRate := 3.0 / 5.0 // 3 hits out of 5 total accesses
	if stats.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, stats.HitRate)
	}
}

// TestLRUCache_Clear tests cache clearing
func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(3)

	// Add items
	cache.Put("key1", &ExecutionResult{Success: true, Output: "test1"})
	cache.Put("key2", &ExecutionResult{Success: true, Output: "test2"})

	// Verify items exist
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// Verify items are gone
	if _, ok := cache.Get("key1"); ok {
		t.Error("key1 should not exist after clear")
	}

	// Verify stats are reset
	stats := cache.GetStats()
	if stats.Hits != 0 || stats.Misses != 1 {
		t.Errorf("Stats should be reset after clear, got hits=%d, misses=%d", stats.Hits, stats.Misses)
	}
}

// TestLRUCache_HashCode tests hash code generation
func TestLRUCache_HashCode(t *testing.T) {
	cache := NewLRUCache(10)

	code1 := "fmt.Println(\"Hello\")"
	code2 := "fmt.Println(\"World\")"
	code3 := "fmt.Println(\"Hello\")" // Same as code1

	hash1 := cache.HashCode(code1)
	hash2 := cache.HashCode(code2)
	hash3 := cache.HashCode(code3)

	// Same code should produce same hash
	if hash1 != hash3 {
		t.Error("Same code should produce same hash")
	}

	// Different code should produce different hash
	if hash1 == hash2 {
		t.Error("Different code should produce different hash")
	}

	// Hash should be hex string
	if len(hash1) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

// TestLRUCache_Capacity tests capacity limits
func TestLRUCache_Capacity(t *testing.T) {
	capacity := 5
	cache := NewLRUCache(capacity)

	if cache.Capacity() != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, cache.Capacity())
	}

	// Fill beyond capacity
	for i := 0; i < 10; i++ {
		key := cache.HashCode(string(rune(i)))
		cache.Put(key, &ExecutionResult{Success: true, Output: "test"})
	}

	// Size should not exceed capacity
	if cache.Size() > capacity {
		t.Errorf("Cache size %d exceeds capacity %d", cache.Size(), capacity)
	}
}

// TestLRUCache_DefaultCapacity tests default capacity
func TestLRUCache_DefaultCapacity(t *testing.T) {
	// Test with zero capacity
	cache1 := NewLRUCache(0)
	if cache1.Capacity() != 100 {
		t.Errorf("Expected default capacity 100, got %d", cache1.Capacity())
	}

	// Test with negative capacity
	cache2 := NewLRUCache(-1)
	if cache2.Capacity() != 100 {
		t.Errorf("Expected default capacity 100, got %d", cache2.Capacity())
	}
}

// TestYaegiInterpreter_CacheIntegration tests cache integration with yaegi
func TestYaegiInterpreter_CacheIntegration(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
		CacheCapacity: 10,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `fmt.Println("Hello, Cache!")`

	ctx := context.Background()

	// First execution - cache miss
	result1, err := yi.Run(ctx, code, "")
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}
	if !result1.Success {
		t.Errorf("First execution should succeed: %s", result1.Error)
	}
	duration1 := result1.Duration

	// Second execution - cache hit
	result2, err := yi.Run(ctx, code, "")
	if err != nil {
		t.Fatalf("Second execution failed: %v", err)
	}
	if !result2.Success {
		t.Errorf("Second execution should succeed: %s", result2.Error)
	}
	duration2 := result2.Duration

	// Cache hit should be much faster
	if duration2 >= duration1 {
		t.Logf("Warning: Cache hit (%v) not faster than miss (%v)", duration2, duration1)
	}

	// Verify output is the same
	if !strings.Contains(result2.Output, "Hello, Cache!") {
		t.Errorf("Expected cached output to contain 'Hello, Cache!', got: %s", result2.Output)
	}

	// Check cache stats
	stats := yi.GetCacheStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
	if stats.HitRate != 0.5 {
		t.Errorf("Expected hit rate 0.5, got %.2f", stats.HitRate)
	}
}

// TestYaegiInterpreter_CacheDisabled tests behavior when cache is disabled
func TestYaegiInterpreter_CacheDisabled(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   false,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `fmt.Println("No cache")`

	ctx := context.Background()

	// Execute twice
	result1, _ := yi.Run(ctx, code, "")
	result2, _ := yi.Run(ctx, code, "")

	// Both should execute normally
	if !result1.Success || !result2.Success {
		t.Error("Executions should succeed even without cache")
	}

	// Cache stats should be empty
	stats := yi.GetCacheStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Cache stats should be zero when cache is disabled")
	}
}

// TestYaegiInterpreter_CacheClear tests cache clearing
func TestYaegiInterpreter_CacheClear(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `fmt.Println("Test")`
	ctx := context.Background()

	// Execute to populate cache
	yi.Run(ctx, code, "")
	yi.Run(ctx, code, "")

	// Verify cache has data
	stats := yi.GetCacheStats()
	if stats.Hits == 0 {
		t.Error("Cache should have hits before clear")
	}

	// Clear cache
	yi.ClearCache()

	// Verify cache is empty
	stats = yi.GetCacheStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Cache stats should be zero after clear")
	}

	// Execute again - should be cache miss
	yi.Run(ctx, code, "")
	stats = yi.GetCacheStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss after clear, got %d", stats.Misses)
	}
}

// TestYaegiInterpreter_CacheFailedExecution tests that failed executions are not cached
func TestYaegiInterpreter_CacheFailedExecution(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	// Code with error
	code := `fmt.Println(undefinedVariable)`
	ctx := context.Background()

	// Execute twice
	result1, _ := yi.Run(ctx, code, "")
	result2, _ := yi.Run(ctx, code, "")

	// Both should fail
	if result1.Success || result2.Success {
		t.Error("Executions should fail")
	}

	// Check cache stats - should have 2 misses, 0 hits
	stats := yi.GetCacheStats()
	if stats.Hits != 0 {
		t.Errorf("Failed executions should not be cached, got %d hits", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
}

// TestYaegiInterpreter_CacheMultipleCode tests caching multiple different code snippets
func TestYaegiInterpreter_CacheMultipleCode(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
		CacheCapacity: 5,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	ctx := context.Background()

	codes := []string{
		`fmt.Println("Code 1")`,
		`fmt.Println("Code 2")`,
		`fmt.Println("Code 3")`,
	}

	// Execute each code twice
	for _, code := range codes {
		// First execution - cache miss
		result1, _ := yi.Run(ctx, code, "")
		if !result1.Success {
			t.Errorf("Execution failed: %s", result1.Error)
		}

		// Second execution - cache hit
		result2, _ := yi.Run(ctx, code, "")
		if !result2.Success {
			t.Errorf("Execution failed: %s", result2.Error)
		}
	}

	// Check cache stats
	stats := yi.GetCacheStats()
	if stats.Hits != 3 {
		t.Errorf("Expected 3 cache hits, got %d", stats.Hits)
	}
	if stats.Misses != 3 {
		t.Errorf("Expected 3 cache misses, got %d", stats.Misses)
	}
}

// BenchmarkLRUCache_Get benchmarks cache get operations
func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(100)
	result := &ExecutionResult{Success: true, Output: "test"}
	cache.Put("key", result)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

// BenchmarkLRUCache_Put benchmarks cache put operations
func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(100)
	result := &ExecutionResult{Success: true, Output: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := cache.HashCode(string(rune(i)))
		cache.Put(key, result)
	}
}

// BenchmarkYaegiCache_Hit benchmarks cache hit performance
func BenchmarkYaegiCache_Hit(b *testing.B) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		b.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `fmt.Println("Benchmark")`
	ctx := context.Background()

	// Warm up cache
	yi.Run(ctx, code, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		yi.Run(ctx, code, "")
	}
}

// BenchmarkYaegiCache_Miss benchmarks cache miss performance
func BenchmarkYaegiCache_Miss(b *testing.B) {
	config := YaegiConfig{
		PreloadStdlib: true,
		EnableCache:   true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		b.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each iteration uses different code to ensure cache miss
		code := `fmt.Println("` + string(rune(i)) + `")`
		yi.Run(ctx, code, "")
	}
}
