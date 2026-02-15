// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Key        string // 缓存键
	Result     *ExecutionResult
	AccessTime time.Time
	HitCount   int64
}

// LRUCache LRU 缓存
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
	stats    CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	TotalSize int64
	HitRate   float64
}

// NewLRUCache 创建 LRU 缓存
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 100 // 默认容量
	}

	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get 获取缓存
func (c *LRUCache) Get(key string) (*ExecutionResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry)
		entry.AccessTime = time.Now()
		entry.HitCount++
		c.stats.Hits++

		// 更新命中率
		c.updateHitRate()

		// 返回结果的副本，避免修改缓存
		result := *entry.Result
		return &result, true
	}

	c.stats.Misses++
	c.updateHitRate()
	return nil, false
}

// Put 添加缓存
func (c *LRUCache) Put(key string, result *ExecutionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新
	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry)
		entry.Result = result
		entry.AccessTime = time.Now()
		return
	}

	// 如果达到容量，淘汰最久未使用的条目
	if c.list.Len() >= c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			entry := oldest.Value.(*CacheEntry)
			delete(c.cache, entry.Key)
			c.stats.Evictions++
		}
	}

	// 添加新条目
	entry := &CacheEntry{
		Key:        key,
		Result:     result,
		AccessTime: time.Now(),
		HitCount:   0,
	}

	elem := c.list.PushFront(entry)
	c.cache[key] = elem
	c.stats.TotalSize++
}

// hashCode 计算代码哈希
func (c *LRUCache) hashCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

// HashCode 公开的哈希方法
func (c *LRUCache) HashCode(code string) string {
	return c.hashCode(code)
}

// GetStats 获取缓存统计
func (c *LRUCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// updateHitRate 更新命中率
func (c *LRUCache) updateHitRate() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	}
}

// Clear 清空缓存
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.list = list.New()
	c.stats = CacheStats{}
}

// Size 获取当前缓存大小
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// Capacity 获取缓存容量
func (c *LRUCache) Capacity() int {
	return c.capacity
}
