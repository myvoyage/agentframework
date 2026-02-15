// Agent Framework - Runtime Utilities
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ObjectPool provides synchronized pooling of objects to reduce memory allocation overhead
type ObjectPool struct {
	// Pool for execution context maps
	executionContextPool sync.Pool
	// Pool for message slices
	messageSlicePool sync.Pool
	// Pool for byte buffers
	byteBufferPool sync.Pool
}

// NewObjectPool creates a new ObjectPool instance
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		executionContextPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
		messageSlicePool: sync.Pool{
			New: func() interface{} {
				return make([]*schema.Message, 0, 10)
			},
		},
		byteBufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024)
			},
		},
	}
}

// GetExecutionContext retrieves an execution context map from the pool
// Returns a map ready for use. Callers must return it via PutExecutionContext.
func (p *ObjectPool) GetExecutionContext() map[string]interface{} {
	return p.executionContextPool.Get().(map[string]interface{})
}

// PutExecutionContext returns an execution context map to the pool
// The map is cleared before being returned to the pool to prevent memory leaks
func (p *ObjectPool) PutExecutionContext(ctx map[string]interface{}) {
	if ctx == nil {
		return
	}
	// Clear all keys to prevent memory leaks
	for k := range ctx {
		delete(ctx, k)
	}
	p.executionContextPool.Put(ctx)
}

// GetMessageSlice retrieves a message slice from the pool
// Returns a slice ready for use. Callers must return it via PutMessageSlice.
func (p *ObjectPool) GetMessageSlice() []*schema.Message {
	return p.messageSlicePool.Get().([]*schema.Message)
}

// PutMessageSlice returns a message slice to the pool
// The slice is reset to zero length before being returned
func (p *ObjectPool) PutMessageSlice(messages []*schema.Message) {
	if messages == nil {
		return
	}
	// Reset to zero length but keep capacity
	messages = messages[:0]
	p.messageSlicePool.Put(messages)
}

// GetByteBuffer retrieves a byte buffer from the pool
// Returns a buffer ready for use. Callers must return it via PutByteBuffer.
func (p *ObjectPool) GetByteBuffer() []byte {
	return p.byteBufferPool.Get().([]byte)
}

// PutByteBuffer returns a byte buffer to the pool
// The buffer is reset to zero length before being returned
func (p *ObjectPool) PutByteBuffer(buf []byte) {
	if buf == nil {
		return
	}
	// Reset to zero length but keep capacity
	buf = buf[:0]
	p.byteBufferPool.Put(buf)
}

// Global object pool instance
var (
	globalObjectPool     *ObjectPool
	globalObjectPoolOnce sync.Once
)

// InitGlobalObjectPool initializes the global object pool
func InitGlobalObjectPool() {
	globalObjectPoolOnce.Do(func() {
		globalObjectPool = NewObjectPool()
	})
}

// GetGlobalObjectPool returns the global object pool instance
// Initializes it with default settings if not already initialized
func GetGlobalObjectPool() *ObjectPool {
	if globalObjectPool == nil {
		InitGlobalObjectPool()
	}
	return globalObjectPool
}

// Convenience functions that use the global pool

// GetExecutionContext retrieves an execution context map from the global pool
func GetExecutionContext() map[string]interface{} {
	return GetGlobalObjectPool().GetExecutionContext()
}

// PutExecutionContext returns an execution context map to the global pool
func PutExecutionContext(ctx map[string]interface{}) {
	GetGlobalObjectPool().PutExecutionContext(ctx)
}

// GetMessageSlice retrieves a message slice from the global pool
func GetMessageSlice() []*schema.Message {
	return GetGlobalObjectPool().GetMessageSlice()
}

// PutMessageSlice returns a message slice to the global pool
func PutMessageSlice(messages []*schema.Message) {
	GetGlobalObjectPool().PutMessageSlice(messages)
}

// GetByteBuffer retrieves a byte buffer from the global pool
func GetByteBuffer() []byte {
	return GetGlobalObjectPool().GetByteBuffer()
}

// PutByteBuffer returns a byte buffer to the global pool
func PutByteBuffer(buf []byte) {
	GetGlobalObjectPool().PutByteBuffer(buf)
}

// ModelPool provides pooling for ChatModel instances to reduce initialization overhead
type ModelPool struct {
	pools map[string]chan model.ChatModel
	mu    sync.RWMutex
}

// NewModelPool creates a new ModelPool instance
func NewModelPool() *ModelPool {
	return &ModelPool{
		pools: make(map[string]chan model.ChatModel),
	}
}

// Get retrieves a model from the pool for the given key
// Returns nil if the pool is empty or doesn't exist
func (p *ModelPool) Get(key string) model.ChatModel {
	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if !exists {
		return nil
	}

	select {
	case model := <-pool:
		return model
	default:
		return nil
	}
}

// Put returns a model to the pool for the given key
// If the pool is full, the model is discarded
func (p *ModelPool) Put(key string, model model.ChatModel, poolSize int) {
	p.mu.Lock()
	if _, exists := p.pools[key]; !exists {
		p.pools[key] = make(chan model.ChatModel, poolSize)
	}
	pool := p.pools[key]
	p.mu.Unlock()

	select {
	case pool <- model:
		// Model successfully returned to pool
	default:
		// Pool is full, discard the model
	}
}

// Clear removes all models from the pool for the given key
func (p *ModelPool) Clear(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if pool, exists := p.pools[key]; exists {
		close(pool)
		delete(p.pools, key)
	}
}

// ClearAll removes all pools
func (p *ModelPool) ClearAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pool := range p.pools {
		close(pool)
		delete(p.pools, key)
	}
}
