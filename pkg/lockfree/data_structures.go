// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package lockfree

import (
	"sync"
	"sync/atomic"
)

// AgentRegistry provides lock-free agent storage using sync.Map
type AgentRegistry struct {
	agents sync.Map
}

// NewAgentRegistry creates a new lock-free agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{}
}

// Register registers an agent
func (r *AgentRegistry) Register(id string, agent interface{}) {
	r.agents.Store(id, agent)
}

// Unregister removes an agent
func (r *AgentRegistry) Unregister(id string) {
	r.agents.Delete(id)
}

// Get retrieves an agent by ID
func (r *AgentRegistry) Get(id string) (interface{}, bool) {
	return r.agents.Load(id)
}

// Range iterates over all agents
func (r *AgentRegistry) Range(fn func(id string, agent interface{}) bool) {
	r.agents.Range(func(key, value interface{}) bool {
		return fn(key.(string), value)
	})
}

// Count returns the approximate number of agents
func (r *AgentRegistry) Count() int {
	count := 0
	r.agents.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Metrics provides lock-free metrics storage
type Metrics struct {
	requestCount uint64
	errorCount   uint64
	totalLatency uint64
	minLatency   uint64
	maxLatency   uint64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		minLatency: ^uint64(0), // Max uint64
	}
}

// IncrementRequestCount atomically increments the request count
func (m *Metrics) IncrementRequestCount() {
	atomic.AddUint64(&m.requestCount, 1)
}

// IncrementErrorCount atomically increments the error count
func (m *Metrics) IncrementErrorCount() {
	atomic.AddUint64(&m.errorCount, 1)
}

// AddLatency atomically adds latency to the total
func (m *Metrics) AddLatency(latency uint64) {
	atomic.AddUint64(&m.totalLatency, latency)

	// Update min latency
	for {
		old := atomic.LoadUint64(&m.minLatency)
		if latency >= old {
			break
		}
		if atomic.CompareAndSwapUint64(&m.minLatency, old, latency) {
			break
		}
	}

	// Update max latency
	for {
		old := atomic.LoadUint64(&m.maxLatency)
		if latency <= old {
			break
		}
		if atomic.CompareAndSwapUint64(&m.maxLatency, old, latency) {
			break
		}
	}
}

// GetRequestCount returns the current request count
func (m *Metrics) GetRequestCount() uint64 {
	return atomic.LoadUint64(&m.requestCount)
}

// GetErrorCount returns the current error count
func (m *Metrics) GetErrorCount() uint64 {
	return atomic.LoadUint64(&m.errorCount)
}

// GetTotalLatency returns the total latency
func (m *Metrics) GetTotalLatency() uint64 {
	return atomic.LoadUint64(&m.totalLatency)
}

// GetMinLatency returns the minimum latency
func (m *Metrics) GetMinLatency() uint64 {
	return atomic.LoadUint64(&m.minLatency)
}

// GetMaxLatency returns the maximum latency
func (m *Metrics) GetMaxLatency() uint64 {
	return atomic.LoadUint64(&m.maxLatency)
}

// GetAverageLatency calculates the average latency
func (m *Metrics) GetAverageLatency() uint64 {
	total := atomic.LoadUint64(&m.totalLatency)
	count := atomic.LoadUint64(&m.requestCount)
	if count == 0 {
		return 0
	}
	return total / count
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	atomic.StoreUint64(&m.requestCount, 0)
	atomic.StoreUint64(&m.errorCount, 0)
	atomic.StoreUint64(&m.totalLatency, 0)
	atomic.StoreUint64(&m.minLatency, ^uint64(0))
	atomic.StoreUint64(&m.maxLatency, 0)
}

// StatusFlag provides an atomic status flag
type StatusFlag struct {
	flag uint32
}

// NewStatusFlag creates a new status flag
func NewStatusFlag(initialStatus uint32) *StatusFlag {
	return &StatusFlag{
		flag: initialStatus,
	}
}

// Set atomically sets the status
func (s *StatusFlag) Set(status uint32) {
	atomic.StoreUint32(&s.flag, status)
}

// Get atomically gets the status
func (s *StatusFlag) Get() uint32 {
	return atomic.LoadUint32(&s.flag)
}

// CompareAndSwap performs a compare-and-swap operation
func (s *StatusFlag) CompareAndSwap(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&s.flag, old, new)
}

// IsRunning checks if the status is running
func (s *StatusFlag) IsRunning() bool {
	return atomic.LoadUint32(&s.flag) == 1
}

// IsStopped checks if the status is stopped
func (s *StatusFlag) IsStopped() bool {
	return atomic.LoadUint32(&s.flag) == 0
}

// ReferenceCounter provides an atomic reference counter
type ReferenceCounter struct {
	count int64
}

// NewReferenceCounter creates a new reference counter
func NewReferenceCounter(initialCount int64) *ReferenceCounter {
	return &ReferenceCounter{
		count: initialCount,
	}
}

// Increment atomically increments the counter
func (r *ReferenceCounter) Increment() {
	atomic.AddInt64(&r.count, 1)
}

// Decrement atomically decrements the counter
func (r *ReferenceCounter) Decrement() int64 {
	return atomic.AddInt64(&r.count, -1)
}

// Get returns the current count
func (r *ReferenceCounter) Get() int64 {
	return atomic.LoadInt64(&r.count)
}

// IsZero checks if the counter is zero
func (r *ReferenceCounter) IsZero() bool {
	return atomic.LoadInt64(&r.count) == 0
}

// Wait waits for the counter to reach zero
func (r *ReferenceCounter) Wait() {
	for atomic.LoadInt64(&r.count) > 0 {
		// In production, use a proper wait mechanism
		// This is simplified for demonstration
	}
}

// Shard provides a sharded map for reduced lock contention
type Shard struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

// NewShard creates a new shard
func NewShard() *Shard {
	return &Shard{
		items: make(map[string]interface{}),
	}
}

// Get retrieves a value from the shard
func (s *Shard) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.items[key]
	return val, ok
}

// Set stores a value in the shard
func (s *Shard) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
}

// Delete removes a value from the shard
func (s *Shard) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// Count returns the number of items in the shard
func (s *Shard) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// ShardedMap provides a map sharded across multiple shards
type ShardedMap struct {
	shards []*Shard
	count  uint32
}

// NewShardedMap creates a new sharded map with the specified number of shards
func NewShardedMap(shardCount uint32) *ShardedMap {
	shards := make([]*Shard, shardCount)
	for i := uint32(0); i < shardCount; i++ {
		shards[i] = NewShard()
	}
	return &ShardedMap{
		shards: shards,
		count:  shardCount,
	}
}

// getShard returns the shard for a given key
func (m *ShardedMap) getShard(key string) *Shard {
	// Simple hash function
	hash := uint32(0)
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return m.shards[hash%m.count]
}

// Get retrieves a value from the sharded map
func (m *ShardedMap) Get(key string) (interface{}, bool) {
	return m.getShard(key).Get(key)
}

// Set stores a value in the sharded map
func (m *ShardedMap) Set(key string, value interface{}) {
	m.getShard(key).Set(key, value)
}

// Delete removes a value from the sharded map
func (m *ShardedMap) Delete(key string) {
	m.getShard(key).Delete(key)
}

// Count returns the total number of items across all shards
func (m *ShardedMap) Count() int {
	total := 0
	for _, shard := range m.shards {
		total += shard.Count()
	}
	return total
}

// Range iterates over all items in all shards
func (m *ShardedMap) Range(fn func(key string, value interface{}) bool) {
	for _, shard := range m.shards {
		shard.mu.RLock()
		for key, value := range shard.items {
			if !fn(key, value) {
				shard.mu.RUnlock()
				return
			}
		}
		shard.mu.RUnlock()
	}
}
