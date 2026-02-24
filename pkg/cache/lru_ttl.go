// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
//
// Smart Cache with LRU and TTL support
// SPDX-License-Identifier: AGPL-3.0-or-later

package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache Least Recently Used cache
type SimpleLRUCache struct {
	capacity int
	items    map[string]*list.Element
	lruList   *list.List
	ttl       time.Duration
	mu        sync.RWMutex

	// Metrics
	hits   int64
	misses int64
}

// cacheEntry 缓存条目
type cacheEntry struct {
	key       string
	value     interface{}
	expiredAt time.Time
	accessAt  time.Time
}

// NewLRUCache 创建新的 LRU 缓存
func NewSimpleLRUCache(capacity int, ttl time.Duration) *SimpleLRUCache {
	if capacity <= 0 {
		capacity = 1000 // 默认容量
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute // 默认TTL 5分钟
	}

	return &SimpleLRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lruList:   list.New(),
		ttl:       ttl,
	}
}

// Set 设置缓存值
func (c *SimpleLRUCache) Set(key string, value interface{}, ttl ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Determine TTL
	expiration := time.Now().Add(c.ttl) // 默认TTL
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = time.Now().Add(ttl[0])
	}

	// Check if key already exists
	if elem, exists := c.items[key]; exists {
		// Update existing entry
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiredAt = expiration
		entry.accessAt = time.Now()

		// Move to front (most recently used)
		c.lruList.MoveToFront(elem)
		return
	}

	// Check capacity
	if c.lruList.Len() >= c.capacity {
		c.evict()
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiredAt: expiration,
		accessAt:  time.Now(),
	}
	elem := c.lruList.PushFront(entry)
	c.items[key] = elem
}

// Get 获取缓存值
func (c *SimpleLRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	entry := elem.Value.(*cacheEntry)

	// Check expiration
	if time.Now().After(entry.expiredAt) {
		// Entry expired
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	// Update access time
	entry.accessAt = time.Now()

	// Move to front (most recently used)
	c.lruList.MoveToFront(elem)

	c.hits++
	return entry.value, true
}

// Delete 删除缓存值
func (c *SimpleLRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
	}
}

// Clear 清空缓存
func (c *SimpleLRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList.Init()
	c.hits = 0
	c.misses = 0
}

// Size 获取缓存大小
func (c *SimpleLRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Capacity 获取缓存容量
func (c *SimpleLRUCache) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.capacity
}

// Hits 获取缓存命中次数
func (c *SimpleLRUCache) Hits() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.hits
}

// Misses 获取缓存未命中次数
func (c *SimpleLRUCache) Misses() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.misses
}

// HitRate 获取缓存命中率
func (c *SimpleLRUCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total)
}

// Keys 获取所有键
func (c *SimpleLRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// evict 淘汰最少使用的条目
func (c *SimpleLRUCache) evict() {
	elem := c.lruList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement 移除元素
func (c *SimpleLRUCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
	c.lruList.Remove(elem)
}

// cleanupExpired 清理过期条目
func (c *SimpleLRUCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for elem := c.lruList.Back(); elem != nil; {
		prev := elem.Prev()
		entry := elem.Value.(*cacheEntry)

		if now.After(entry.expiredAt) {
			c.removeElement(elem)
		}

		elem = prev
	}
}

// ShardedLRUCache 分片 LRU 缓存
type ShardedLRUCache struct {
	shards    []*SimpleLRUCache
	shardMask uint32
	mu        sync.RWMutex
}

// NewShardedLRUCache 创建新的分片 LRU 缓存
func NewShardedLRUCache(capacityPerShard int, shardCount uint32, ttl time.Duration) *ShardedLRUCache {
	if shardCount == 0 {
		shardCount = 16 // 默认16个分片
	}
	if !isPowerOfTwo(shardCount) {
		shardCount = roundToPowerOfTwo(shardCount)
	}

	shards := make([]*SimpleLRUCache, shardCount)
	for i := range shards {
		shards[i] = NewSimpleLRUCache(capacityPerShard, ttl)
	}

	return &ShardedLRUCache{
		shards:    shards,
		shardMask: shardCount - 1,
	}
}

// Set 设置缓存值
func (c *ShardedLRUCache) Set(key string, value interface{}, ttl ...time.Duration) {
	shard := c.getShard(key)
	shard.Set(key, value, ttl...)
}

// Get 获取缓存值
func (c *ShardedLRUCache) Get(key string) (interface{}, bool) {
	shard := c.getShard(key)
	return shard.Get(key)
}

// Delete 删除缓存值
func (c *ShardedLRUCache) Delete(key string) {
	shard := c.getShard(key)
	shard.Delete(key)
}

// Clear 清空缓存
func (c *ShardedLRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, shard := range c.shards {
		shard.Clear()
	}
}

// Size 获取缓存大小
func (c *ShardedLRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	for _, shard := range c.shards {
		total += shard.Size()
	}
	return total
}

// Capacity 获取缓存总容量
func (c *ShardedLRUCache) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	for _, shard := range c.shards {
		total += shard.Capacity()
	}
	return total
}

// Hits 获取缓存命中次数
func (c *ShardedLRUCache) Hits() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := int64(0)
	for _, shard := range c.shards {
		total += shard.Hits()
	}
	return total
}

// Misses 获取缓存未命中次数
func (c *ShardedLRUCache) Misses() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := int64(0)
	for _, shard := range c.shards {
		total += shard.Misses()
	}
	return total
}

// HitRate 获取缓存命中率
func (c *ShardedLRUCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalHits := int64(0)
	totalMisses := int64(0)
	for _, shard := range c.shards {
		totalHits += shard.Hits()
		totalMisses += shard.Misses()
	}

	total := totalHits + totalMisses
	if total == 0 {
		return 0.0
	}
	return float64(totalHits) / float64(total)
}

// Keys 获取所有键
func (c *ShardedLRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for _, shard := range c.shards {
		keys = append(keys, shard.Keys()...)
	}
	return keys
}

// cleanupExpired 清理过期条目
func (c *ShardedLRUCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, shard := range c.shards {
		shard.cleanupExpired()
	}
}

// getShard 获取分片
func (c *ShardedLRUCache) getShard(key string) *SimpleLRUCache {
	hash := fnv32(key)
	return c.shards[hash&c.shardMask]
}

// isPowerOfTwo 检查是否为2的幂
func isPowerOfTwo(n uint32) bool {
	return n != 0 && (n&(n-1)) == 0
}

// roundToPowerOfTwo 舍入到2的幂
func roundToPowerOfTwo(n uint32) uint32 {
	if isPowerOfTwo(n) {
		return n
	}

	var power uint32 = 1
	for power < n {
		power <<= 1
	}
	return power
}

// fnv32 FNV-1a 32位哈希算法
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}
	return hash
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Capacity    int           // 缓存容量
	TTL         time.Duration // TTL（生存时间）
	ShardCount  uint32        // 分片数量（0=不分片）
	CleanupInt  time.Duration // 清理间隔
}

// SmartCache 智能缓存（支持多种策略）
type SmartCache struct {
	lru      *SimpleLRUCache
	sharded  *ShardedLRUCache
	config   *CacheConfig
	stopChan chan bool
	wg        sync.WaitGroup
}

// NewSmartCache 创建新的智能缓存
func NewSmartCache(config *CacheConfig) *SmartCache {
	if config == nil {
		config = &CacheConfig{
			Capacity:   1000,
			TTL:        5 * time.Minute,
			ShardCount: 0,
			CleanupInt: 1 * time.Minute,
		}
	}

	cache := &SmartCache{
		config:   config,
		stopChan: make(chan bool),
	}

	// 创建合适的缓存类型
	if config.ShardCount > 0 {
		cache.sharded = NewShardedLRUCache(
			config.Capacity,
			config.ShardCount,
			config.TTL,
		)
	} else {
		cache.lru = NewSimpleLRUCache(
			config.Capacity,
			config.TTL,
		)
	}

	// 启动清理协程
	if config.CleanupInt > 0 {
		cache.wg.Add(1)
		go cache.cleanupLoop()
	}

	return cache
}

// Set 设置缓存值
func (c *SmartCache) Set(key string, value interface{}, ttl ...time.Duration) {
	if c.sharded != nil {
		c.sharded.Set(key, value, ttl...)
	} else {
		c.lru.Set(key, value, ttl...)
	}
}

// Get 获取缓存值
func (c *SmartCache) Get(key string) (interface{}, bool) {
	if c.sharded != nil {
		return c.sharded.Get(key)
	}
	return c.lru.Get(key)
}

// Delete 删除缓存值
func (c *SmartCache) Delete(key string) {
	if c.sharded != nil {
		c.sharded.Delete(key)
	} else {
		c.lru.Delete(key)
	}
}

// Clear 清空缓存
func (c *SmartCache) Clear() {
	if c.sharded != nil {
		c.sharded.Clear()
	} else {
		c.lru.Clear()
	}
}

// Size 获取缓存大小
func (c *SmartCache) Size() int {
	if c.sharded != nil {
		return c.sharded.Size()
	}
	return c.lru.Size()
}

// Capacity 获取缓存容量
func (c *SmartCache) Capacity() int {
	if c.sharded != nil {
		return c.sharded.Capacity()
	}
	return c.lru.Capacity()
}

// Hits 获取缓存命中次数
func (c *SmartCache) Hits() int64 {
	if c.sharded != nil {
		return c.sharded.Hits()
	}
	return c.lru.Hits()
}

// Misses 获取缓存未命中次数
func (c *SmartCache) Misses() int64 {
	if c.sharded != nil {
		return c.sharded.Misses()
	}
	return c.lru.Misses()
}

// HitRate 获取缓存命中率
func (c *SmartCache) HitRate() float64 {
	if c.sharded != nil {
		return c.sharded.HitRate()
	}
	return c.lru.HitRate()
}

// Keys 获取所有键
func (c *SmartCache) Keys() []string {
	if c.sharded != nil {
		return c.sharded.Keys()
	}
	return c.lru.Keys()
}

// Stop 停止缓存
func (c *SmartCache) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

// cleanupLoop 清理循环
func (c *SmartCache) cleanupLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CleanupInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.sharded != nil {
				c.sharded.cleanupExpired()
			} else {
				c.lru.cleanupExpired()
			}
		case <-c.stopChan:
			return
		}
	}
}
