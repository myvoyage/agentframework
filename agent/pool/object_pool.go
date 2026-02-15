// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package pool

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

// ObjectPool provides reusable object pooling to reduce memory allocations
type ObjectPool struct {
 pools map[string]interface{}
	mu   sync.RWMutex
}

// NewObjectPool creates a new object pool
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		pools: make(map[string]interface{}),
	}
}

// Get retrieves a pool by name, creating it if it doesn't exist
func (p *ObjectPool) Get(name string) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if pool, exists := p.pools[name]; exists {
		return pool
	}

	// Create new pool based on type
	var newPool interface{}
	switch name {
	case "map":
		newPool = make(map[string]interface{})
	case "slice":
		newPool = make([]interface{}, 0, 10)
	case "messages":
		newPool = make([]*schema.Message, 0, 5)
	default:
		newPool = make(map[string]interface{})
	}

	p.pools[name] = newPool
	return newPool
}

// Put returns an object to the pool for reuse
func (p *ObjectPool) Put(name string, obj interface{}) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, exists := p.pools[name]; exists {
		// For simplicity, we just discard the object
		// In production, you might want to reset/clear the object
		// Store the object back in the pool for reuse
		p.pools[name] = obj
	}
}

// Cleanup removes a pool from the pool
func (p *ObjectPool) Cleanup(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.pools, name)
}

// GetMap retrieves or creates a map pool
func (p *ObjectPool) GetMap(name string) map[string]interface{} {
	if obj := p.Get(name); obj != nil {
		return obj.(map[string]interface{})
	}
	return make(map[string]interface{})
}

// GetSlice retrieves or creates a slice pool
func (p *ObjectPool) GetSlice(name string, capacity int) []interface{} {
	if obj := p.Get(name); obj != nil {
		return obj.([]interface{})
	}
	return make([]interface{}, 0, capacity)
}

// PutMap returns a map to the pool
func (p *ObjectPool) PutMap(name string, m map[string]interface{}) {
	p.Put(name, m)
}

// PutSlice returns a slice to the pool
func (p *ObjectPool) PutSlice(name string, s []interface{}) {
	p.Put(name, s)
}

// Global object pool instance
var globalObjectPool = NewObjectPool()

// GetGlobalPool returns the global object pool
func GetGlobalPool() *ObjectPool {
	return globalObjectPool
}
