// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package cache

import (
	"context"
	"sync"
	"time"
)

// Cache defines the interface for cache implementations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// MultiLevelCache provides a multi-level caching strategy
type MultiLevelCache struct {
	l1 Cache // Level 1: In-memory cache (fast, small)
	l2 Cache // Level 2: Redis/distributed cache (medium, large)
	l3 Cache // Level 3: Persistent storage (slow, huge)

	l1TTL time.Duration
	l2TTL time.Duration
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(l1, l2, l3 Cache, l1TTL, l2TTL time.Duration) *MultiLevelCache {
	return &MultiLevelCache{
		l1:   l1,
		l2:   l2,
		l3:   l3,
		l1TTL: l1TTL,
		l2TTL: l2TTL,
	}
}

// Get retrieves a value from the cache, checking all levels
func (mc *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	// Level 1: In-memory cache
	val, err := mc.l1.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// Level 2: Redis/distributed cache
	val, err = mc.l2.Get(ctx, key)
	if err == nil {
		// Promote to L1
		mc.l1.Set(ctx, key, val, mc.l1TTL)
		return val, nil
	}

	// Level 3: Persistent storage
	val, err = mc.l3.Get(ctx, key)
	if err == nil {
		// Promote to L2 and L1
		mc.l2.Set(ctx, key, val, mc.l2TTL)
		mc.l1.Set(ctx, key, val, mc.l1TTL)
		return val, nil
	}

	return nil, err
}

// Set stores a value in all cache levels
func (mc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Store in all levels
	l1TTL := minTTL(ttl, mc.l1TTL)
	l2TTL := minTTL(ttl, mc.l2TTL)

	// L1: In-memory cache
	if err := mc.l1.Set(ctx, key, value, l1TTL); err != nil {
		return err
	}

	// L2: Redis/distributed cache
	if mc.l2 != nil {
		if err := mc.l2.Set(ctx, key, value, l2TTL); err != nil {
			return err
		}
	}

	// L3: Persistent storage
	if mc.l3 != nil {
		return mc.l3.Set(ctx, key, value, ttl)
	}

	return nil
}

// Delete removes a value from all cache levels
func (mc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	var firstErr error

	if mc.l1 != nil {
		if err := mc.l1.Delete(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if mc.l2 != nil {
		if err := mc.l2.Delete(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if mc.l3 != nil {
		if err := mc.l3.Delete(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Clear clears all cache levels
func (mc *MultiLevelCache) Clear(ctx context.Context) error {
	var firstErr error

	if mc.l1 != nil {
		if err := mc.l1.Clear(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if mc.l2 != nil {
		if err := mc.l2.Clear(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if mc.l3 != nil {
		if err := mc.l3.Clear(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// InMemoryCache provides a simple in-memory cache
type InMemoryCache struct {
	items map[string]*cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, ErrCacheKeyNotFound
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(c.items, key)
		return nil, ErrCacheKeyNotFound
	}

	return item.value, nil
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear clears all items from the cache
func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	return nil
}

// cleanup periodically removes expired items
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// CacheStats provides cache statistics
type CacheStats struct {
	Hits     int64
	Misses   int64
	Sets     int64
	Deletes  int64
	Evictions int64
}

// StatsCache wraps a cache and provides statistics
type StatsCache struct {
	cache Cache
	stats CacheStats
	mu    sync.RWMutex
}

// NewStatsCache creates a new stats cache
func NewStatsCache(cache Cache) *StatsCache {
	return &StatsCache{
		cache: cache,
	}
}

// Get retrieves a value and records statistics
func (sc *StatsCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := sc.cache.Get(ctx, key)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if err == nil {
		sc.stats.Hits++
	} else {
		sc.stats.Misses++
	}

	return val, err
}

// Set stores a value and records statistics
func (sc *StatsCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	err := sc.cache.Set(ctx, key, value, ttl)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if err == nil {
		sc.stats.Sets++
	}

	return err
}

// Delete removes a value and records statistics
func (sc *StatsCache) Delete(ctx context.Context, key string) error {
	err := sc.cache.Delete(ctx, key)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if err == nil {
		sc.stats.Deletes++
	}

	return err
}

// Clear clears the cache
func (sc *StatsCache) Clear(ctx context.Context) error {
	return sc.cache.Clear(ctx)
}

// GetStats returns the current statistics
func (sc *StatsCache) GetStats() CacheStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.stats
}

// HitRate returns the cache hit rate
func (sc *StatsCache) HitRate() float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	total := sc.stats.Hits + sc.stats.Misses
	if total == 0 {
		return 0
	}

	return float64(sc.stats.Hits) / float64(total) * 100
}

// Errors
var (
	ErrCacheKeyNotFound = &CacheError{Message: "key not found in cache"}
)

// CacheError represents a cache error
type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

// Helper function to get minimum TTL
func minTTL(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// MultiLevelLRUCache provides an LRU (Least Recently Used) cache
type MultiLevelLRUCache struct {
	capacity int
	items    map[string]*lruItem
	head     *lruItem
	tail     *lruItem
	mu       sync.RWMutex
}

type lruItem struct {
	key        string
	value      interface{}
	prev, next *lruItem
}

// NewMultiLevelLRUCache creates a new LRU cache
func NewMultiLevelLRUCache(capacity int) *MultiLevelLRUCache {
	cache := &MultiLevelLRUCache{
		capacity: capacity,
		items:    make(map[string]*lruItem),
		head:     &lruItem{},
		tail:     &lruItem{},
	}

	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get retrieves a value from the LRU cache
func (c *MultiLevelLRUCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, ErrCacheKeyNotFound
	}

	// Move to front (most recently used)
	c.moveToFront(item)

	return item.value, nil
}

// Set stores a value in the LRU cache
func (c *MultiLevelLRUCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TTL is ignored for LRU cache
	_ = ttl

	if item, ok := c.items[key]; ok {
		// Update existing item
		item.value = value
		c.moveToFront(item)
		return nil
	}

	// Create new item
	item := &lruItem{
		key:   key,
		value: value,
	}

	c.items[key] = item
	c.addToFront(item)

	// Check capacity
	if len(c.items) > c.capacity {
		c.removeOldest()
	}

	return nil
}

// Delete removes a value from the LRU cache
func (c *MultiLevelLRUCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil
	}

	c.removeItem(item)
	delete(c.items, key)

	return nil
}

// Clear clears all items from the LRU cache
func (c *MultiLevelLRUCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*lruItem)
	c.head.next = c.tail
	c.tail.prev = c.head

	return nil
}

// moveToFront moves an item to the front of the list
func (c *MultiLevelLRUCache) moveToFront(item *lruItem) {
	c.removeItem(item)
	c.addToFront(item)
}

// addToFront adds an item to the front of the list
func (c *MultiLevelLRUCache) addToFront(item *lruItem) {
	item.prev = c.head
	item.next = c.head.next

	c.head.next.prev = item
	c.head.next = item
}

// removeItem removes an item from the list
func (c *MultiLevelLRUCache) removeItem(item *lruItem) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

// removeOldest removes the oldest item from the list
func (c *MultiLevelLRUCache) removeOldest() {
	item := c.tail.prev
	if item != c.head {
		c.removeItem(item)
		delete(c.items, item.key)
	}
}
