// Agent Framework - Skill System Performance Optimization
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"sync"
	"time"
)

// CacheLevel defines the caching level
type CacheLevel int

const (
	CacheLevelNone       CacheLevel = iota // No caching
	CacheLevelMetadata                     // Cache metadata only
	CacheLevelDefinition                   // Cache full definitions
	CacheLevelAll                          // Cache everything
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      interface{}
	CreatedAt time.Time
	AccessAt  time.Time
	ExpiresAt time.Time
	HitCount  int64
	Size      int64
}

// IsExpired checks if the cache entry is expired
func (e *CacheEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false // Never expires
	}
	return time.Now().After(e.ExpiresAt)
}

// CacheConfig defines cache configuration
type CacheConfig struct {
	MaxSize       int64         // Maximum cache size in bytes
	MaxEntries    int           // Maximum number of entries
	DefaultTTL    time.Duration // Default time-to-live
	CleanupPeriod time.Duration // Cleanup period for expired entries
	EnableStats   bool          // Enable cache statistics
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:       100 * 1024 * 1024, // 100MB
		MaxEntries:    1000,
		DefaultTTL:    30 * time.Minute,
		CleanupPeriod: 5 * time.Minute,
		EnableStats:   true,
	}
}

// CacheStats tracks cache statistics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	TotalSize  int64
	EntryCount int
	mu         sync.RWMutex
}

// HitRate returns the cache hit rate
func (s *CacheStats) HitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// RecordHit records a cache hit
func (s *CacheStats) RecordHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hits++
}

// RecordMiss records a cache miss
func (s *CacheStats) RecordMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Misses++
}

// RecordEviction records a cache eviction
func (s *CacheStats) RecordEviction() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Evictions++
}

// MultiLevelCache implements a multi-level cache system
type MultiLevelCache struct {
	l1Cache *MemoryCache // Fast memory cache (L1)
	l2Cache *MemoryCache // Slower memory cache (L2)
	config  *CacheConfig
	stats   *CacheStats
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(config *CacheConfig) *MultiLevelCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &MultiLevelCache{
		l1Cache: NewMemoryCache(config.MaxSize/2, config.MaxEntries/2, config.DefaultTTL/2),
		l2Cache: NewMemoryCache(config.MaxSize/2, config.MaxEntries/2, config.DefaultTTL),
		config:  config,
		stats:   &CacheStats{},
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves an item from cache (tries L1 first, then L2)
func (c *MultiLevelCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Try L1 cache first
	if data, found := c.l1Cache.Get(key); found {
		c.stats.RecordHit()
		return data, true
	}

	// Try L2 cache
	if data, found := c.l2Cache.Get(key); found {
		c.stats.RecordHit()
		// Promote to L1
		c.l1Cache.Set(key, data, c.config.DefaultTTL/2)
		return data, true
	}

	c.stats.RecordMiss()
	return nil, false
}

// Set stores an item in both L1 and L2 caches
func (c *MultiLevelCache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store in both levels
	c.l1Cache.Set(key, data, ttl/2)
	c.l2Cache.Set(key, data, ttl)
}

// Delete removes an item from all cache levels
func (c *MultiLevelCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.l1Cache.Delete(key)
	c.l2Cache.Delete(key)
}

// Clear clears all cache levels
func (c *MultiLevelCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.l1Cache.Clear()
	c.l2Cache.Clear()
}

// GetStats returns cache statistics
func (c *MultiLevelCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"hits":       c.stats.Hits,
		"misses":     c.stats.Misses,
		"evictions":  c.stats.Evictions,
		"hit_rate":   c.stats.HitRate(),
		"l1_entries": c.l1Cache.EntryCount(),
		"l2_entries": c.l2Cache.EntryCount(),
		"l1_size":    c.l1Cache.TotalSize(),
		"l2_size":    c.l2Cache.TotalSize(),
		"total_size": c.l1Cache.TotalSize() + c.l2Cache.TotalSize(),
	}
}

// cleanupLoop periodically cleans up expired entries
func (c *MultiLevelCache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.l1Cache.CleanupExpired()
			c.l2Cache.CleanupExpired()
		case <-c.ctx.Done():
			return
		}
	}
}

// Close closes the cache and stops cleanup goroutines
func (c *MultiLevelCache) Close() {
	c.cancel()
}

// MemoryCache implements a simple in-memory cache
type MemoryCache struct {
	entries     map[string]*CacheEntry
	maxSize     int64
	maxEntries  int
	currentSize int64
	mu          sync.RWMutex
	stats       *CacheStats // 添加统计字段
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(maxSize int64, maxEntries int, defaultTTL time.Duration) *MemoryCache {
	return &MemoryCache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    maxSize,
		maxEntries: maxEntries,
		stats:      &CacheStats{}, // 初始化统计对象
	}
}

// Get retrieves an item from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.entries[key]
	if !found {
		c.stats.RecordMiss()
		return nil, false
	}

	// Check expiration
	if entry.IsExpired() {
		c.stats.RecordMiss()
		return nil, false
	}

	// Update access time and hit count
	entry.AccessAt = time.Now()
	entry.HitCount++
	c.stats.RecordHit()

	return entry.Data, true
}

// Set stores an item in cache
func (c *MemoryCache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict
	c.evictIfNeeded()

	// Calculate size (rough estimate)
	size := estimateSize(data)

	// Create entry
	now := time.Now()
	entry := &CacheEntry{
		Data:      data,
		CreatedAt: now,
		AccessAt:  now,
		ExpiresAt: now.Add(ttl),
		HitCount:  0,
		Size:      size,
	}

	// Update size
	if oldEntry, exists := c.entries[key]; exists {
		c.currentSize -= oldEntry.Size
	}
	c.currentSize += size

	c.entries[key] = entry
}

// Delete removes an item from cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		c.currentSize -= entry.Size
		delete(c.entries, key)
	}
}

// Clear clears all entries
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.currentSize = 0
}

// EntryCount returns the number of entries
func (c *MemoryCache) EntryCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// TotalSize returns the total cache size
func (c *MemoryCache) TotalSize() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// GetStats returns cache statistics
func (c *MemoryCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"hits":       c.stats.Hits,
		"misses":     c.stats.Misses,
		"evictions":  c.stats.Evictions,
		"hit_rate":   c.stats.HitRate(),
		"entries":    c.EntryCount(),
		"size":       c.TotalSize(),
		"max_size":   c.maxSize,
		"max_entries": c.maxEntries,
	}
}
func (c *MemoryCache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.IsExpired() {
			c.currentSize -= entry.Size
			delete(c.entries, key)
		}
	}
}

// evictIfNeeded evicts entries if cache is full
func (c *MemoryCache) evictIfNeeded() {
	// Check entry count
	if len(c.entries) >= c.maxEntries {
		c.evictLRU(1)
		c.stats.RecordEviction() // 记录驱逐次数
	}

	// Check size
	if c.currentSize >= c.maxSize {
		c.evictLRU(5) // Evict more entries if over size
		c.stats.RecordEviction() // 记录驱逐次数
	}
}

// evictLRU evicts least recently used entries
func (c *MemoryCache) evictLRU(count int) {
	// Find LRU entries
	type lruEntry struct {
		key      string
		accessAt time.Time
		hitCount int64
	}

	lruList := make([]lruEntry, 0, len(c.entries))
	for key, entry := range c.entries {
		lruList = append(lruList, lruEntry{
			key:      key,
			accessAt: entry.AccessAt,
			hitCount: entry.HitCount,
		})
	}

	// Sort by access time (oldest first) and hit count
	for i := 0; i < len(lruList)-1; i++ {
		for j := i + 1; j < len(lruList); j++ {
			if lruList[i].accessAt.After(lruList[j].accessAt) ||
				(lruList[i].accessAt.Equal(lruList[j].accessAt) && lruList[i].hitCount > lruList[j].hitCount) {
				lruList[i], lruList[j] = lruList[j], lruList[i]
			}
		}
	}

	// Evict entries
	evicted := 0
	for _, entry := range lruList {
		if evicted >= count {
			break
		}
		if existing, exists := c.entries[entry.key]; exists {
			c.currentSize -= existing.Size
			delete(c.entries, entry.key)
			evicted++
		}
	}
}

// estimateSize estimates the size of data in bytes
func estimateSize(data interface{}) int64 {
	// Rough estimation
	switch v := data.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int8, int16, int32, uint, uint8, uint16, uint32:
		return 4
	case int64, uint64, float64:
		return 8
	case float32:
		return 4
	case bool:
		return 1
	default:
		// Default estimation for complex types
		return 100
	}
}
