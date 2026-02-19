// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
//
// Smart Cache Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestLRUCache_Basic(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}
}

func TestLRUCache_Overwrite(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Set initial value
	cache.Set("key1", "value1")
	val, _ := cache.Get("key1")
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}

	// Overwrite with new value
	cache.Set("key1", "value2")
	val, _ = cache.Get("key1")
	if val != "value2" {
		t.Fatalf("expected value2, got %v", val)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3, 5*time.Minute)

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.Size())
	}

	// Add one more (should evict key1 as it was accessed first)
	cache.Set("key4", "value4")

	if cache.Size() != 3 {
		t.Fatalf("expected size 3 after eviction, got %d", cache.Size())
	}

	// key1 should be evicted
	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("expected key1 to be evicted")
	}

	// Others should still exist
	for _, key := range []string{"key2", "key3", "key4"} {
		if _, ok := cache.Get(key); !ok {
			t.Fatalf("expected %s to exist", key)
		}
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Add items
	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Access key1 (move to front)
	cache.Get("key1")

	// Add one more (should evict key2 as it's the least recently used after key1)
	cache.Set("key11", 11)

	// key2 should be evicted (least recently used, key1 was accessed)
	_, ok := cache.Get("key2")
	if ok {
		t.Fatal("expected key2 to be evicted")
	}

	// key1 should still exist (was accessed recently)
	_, ok = cache.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}

	// key11 should exist
	_, ok = cache.Get("key11")
	if !ok {
		t.Fatal("expected key11 to exist")
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(10, 100*time.Millisecond)

	// Set with short TTL
	cache.Set("key1", "value1", 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("expected key1 to be expired")
	}

	// Cache should report miss
	if cache.Misses() == 0 {
		t.Fatal("expected at least one miss")
	}
}

func TestLRUCache_HitRate(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Set some values
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Generate some hits and misses
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Get(key) // hit
	}
	cache.Get("key_invalid") // miss

	// Check metrics
	expectedHits := int64(5)
	expectedMisses := int64(1)

	if cache.Hits() != expectedHits {
		t.Fatalf("expected %d hits, got %d", expectedHits, cache.Hits())
	}

	if cache.Misses() != expectedMisses {
		t.Fatalf("expected %d misses, got %d", expectedMisses, cache.Misses())
	}

	expectedRate := float64(expectedHits) / float64(expectedHits+expectedMisses)
	actualRate := cache.HitRate()
	if actualRate != expectedRate {
		t.Fatalf("expected hit rate %.2f, got %.2f", expectedRate, actualRate)
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Set value
	cache.Set("key1", "value1")
	if cache.Size() != 1 {
		t.Fatalf("expected size 1, got %d", cache.Size())
	}

	// Delete value
	cache.Delete("key1")
	if cache.Size() != 0 {
		t.Fatalf("expected size 0 after delete, got %d", cache.Size())
	}

	// Should not exist
	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("expected key1 to not exist after delete")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Set multiple values
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	if cache.Size() != 5 {
		t.Fatalf("expected size 5, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", cache.Size())
	}

	// Metrics should be reset
	if cache.Hits() != 0 || cache.Misses() != 0 {
		t.Fatal("expected metrics to be reset")
	}
}

func TestLRUCache_Keys(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Minute)

	// Set multiple values
	expectedKeys := []string{"key1", "key2", "key3"}
	for _, key := range expectedKeys {
		cache.Set(key, key)
	}

	// Get all keys
	keys := cache.Keys()
	if len(keys) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(keys))
	}
}

func TestLRUCache_Capacity(t *testing.T) {
	cache := NewLRUCache(5, 5*time.Minute)

	if cache.Capacity() != 5 {
		t.Fatalf("expected capacity 5, got %d", cache.Capacity())
	}
}

func TestShardedLRUCache_Basic(t *testing.T) {
	cache := NewShardedLRUCache(10, 4, 5*time.Minute)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}
}

func TestShardedLRUCache_Distribution(t *testing.T) {
	cache := NewShardedLRUCache(100, 4, 5*time.Minute)

	// Add multiple items
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Check total size
	if cache.Size() != 100 {
		t.Fatalf("expected size 100, got %d", cache.Size())
	}

	// Check capacity
	if cache.Capacity() != 400 { // 100 * 4 shards
		t.Fatalf("expected capacity 400, got %d", cache.Capacity())
	}
}

func TestShardedLRUCache_HitRate(t *testing.T) {
	cache := NewShardedLRUCache(100, 4, 5*time.Minute)

	// Add items
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Generate hits
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Get(key)
	}

	// Generate miss
	cache.Get("key_invalid")

	// Check hit rate
	if cache.Hits() != 50 {
		t.Fatalf("expected 50 hits, got %d", cache.Hits())
	}

	if cache.Misses() != 1 {
		t.Fatalf("expected 1 miss, got %d", cache.Misses())
	}

	expectedRate := 50.0 / 51.0
	actualRate := cache.HitRate()
	if actualRate != expectedRate {
		t.Fatalf("expected hit rate %.2f, got %.2f", expectedRate, actualRate)
	}
}

func TestShardedLRUCache_Clear(t *testing.T) {
	cache := NewShardedLRUCache(10, 4, 5*time.Minute)

	// Set values
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", cache.Size())
	}
}

func TestSmartCache_Basic(t *testing.T) {
	config := &CacheConfig{
		Capacity:    10,
		TTL:         5 * time.Minute,
		ShardCount:  0,
		CleanupInt:  1 * time.Minute,
	}
	cache := NewSmartCache(config)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}
}

func TestSmartCache_WithShards(t *testing.T) {
	config := &CacheConfig{
		Capacity:    10,
		TTL:         5 * time.Minute,
		ShardCount:  4,
		CleanupInt:  1 * time.Minute,
	}
	cache := NewSmartCache(config)

	// Test Set and Get
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	// Check size
	if cache.Size() != 20 {
		t.Fatalf("expected size 20, got %d", cache.Size())
	}
}

func TestSmartCache_Delete(t *testing.T) {
	config := &CacheConfig{
		Capacity:    10,
		TTL:         5 * time.Minute,
		ShardCount:  0,
		CleanupInt:  1 * time.Minute,
	}
	cache := NewSmartCache(config)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("expected key1 to not exist after delete")
	}
}

func TestSmartCache_Clear(t *testing.T) {
	config := &CacheConfig{
		Capacity:    10,
		TTL:         5 * time.Minute,
		ShardCount: 0,
		CleanupInt:  1 * time.Minute,
	}
	cache := NewSmartCache(config)

	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", cache.Size())
	}
}

func TestSmartCache_Stop(t *testing.T) {
	config := &CacheConfig{
		Capacity:    10,
		TTL:         5 * time.Minute,
		ShardCount:  0,
		CleanupInt: 100 * time.Millisecond,
	}
	cache := NewSmartCache(config)

	// Cache should be running
	cache.Set("key1", "value1")

	// Stop cache
	cache.Stop()

	// Wait a bit for cleanup to finish
	time.Sleep(200 * time.Millisecond)
}

// Benchmark tests
func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)
	for i := 0; i < b.N; i++ {
		key := "key"
		cache.Set(key, "value")
		cache.Get(key)
	}
}

func BenchmarkLRUCache_ParallelGet(b *testing.B) {
	cache := NewLRUCache(10000, 5*time.Minute)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := "key"
			cache.Set(key, "value")
			cache.Get(key)
		}
	})
}

func BenchmarkShardedLRUCache_Get(b *testing.B) {
	cache := NewShardedLRUCache(10000, 4, 5*time.Minute)
	for i := 0; i < b.N; i++ {
		key := "key"
		cache.Set(key, "value")
		cache.Get(key)
	}
}

func BenchmarkShardedLRUCache_ParallelGet(b *testing.B) {
	cache := NewShardedLRUCache(10000, 4, 5*time.Minute)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := "key"
			cache.Set(key, "value")
			cache.Get(key)
		}
	})
}
