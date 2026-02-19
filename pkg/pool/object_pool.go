// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package pool

import (
	"sync"
	"time"
)

// MessagePool manages a pool of Message objects to reduce GC pressure
type MessagePool struct {
	pool sync.Pool
	stats *PoolStats
}

// PoolStats tracks pool statistics
type PoolStats struct {
	Allocated int64
	Pooled     int64
	Reused     int64
	mu         sync.RWMutex
}

// NewMessagePool creates a new message pool
func NewMessagePool() *MessagePool {
	return &MessagePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Message{
					Content:    make([]byte, 0, 1024), // 1KB default capacity
					Metadata:   make(map[string]string),
					Timestamp:  time.Now(),
				}
			},
		},
		stats: &PoolStats{},
	}
}

// Get retrieves a message from the pool or creates a new one
func (p *MessagePool) Get() *Message {
	msg := p.pool.Get().(*Message)
	p.stats.mu.Lock()
	p.stats.Reused++
	p.stats.mu.Unlock()
	return msg
}

// Put returns a message to the pool for reuse
func (p *MessagePool) Put(msg *Message) {
	if msg == nil {
		return
	}

	// Reset the message for reuse
	msg.Reset()

	p.stats.mu.Lock()
	p.stats.Pooled++
	p.stats.mu.Unlock()

	p.pool.Put(msg)
}

// Stats returns the current pool statistics
func (p *MessagePool) Stats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return PoolStats{
		Allocated: p.stats.Allocated,
		Pooled:     p.stats.Pooled,
		Reused:     p.stats.Reused,
	}
}

// EventPool manages a pool of Event objects
type EventPool struct {
	pool sync.Pool
	stats *PoolStats
}

// NewEventPool creates a new event pool
func NewEventPool() *EventPool {
	return &EventPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Event{
					Type:      "",
					Timestamp: time.Now(),
					Data:      make(map[string]interface{}),
					Source:    "",
				}
			},
		},
		stats: &PoolStats{},
	}
}

// Get retrieves an event from the pool
func (p *EventPool) Get() *Event {
	event := p.pool.Get().(*Event)
	p.stats.mu.Lock()
	p.stats.Reused++
	p.stats.mu.Unlock()
	return event
}

// Put returns an event to the pool
func (p *EventPool) Put(event *Event) {
	if event == nil {
		return
	}

	event.Reset()

	p.stats.mu.Lock()
	p.stats.Pooled++
	p.stats.mu.Unlock()

	p.pool.Put(event)
}

// Stats returns the current pool statistics
func (p *EventPool) Stats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return PoolStats{
		Allocated: p.stats.Allocated,
		Pooled:     p.stats.Pooled,
		Reused:     p.stats.Reused,
	}
}

// ContextPool manages a pool of Context objects
type ContextPool struct {
	pool sync.Pool
	stats *PoolStats
}

// NewContextPool creates a new context pool
func NewContextPool() *ContextPool {
	return &ContextPool{
		pool: sync.Pool{
			New: func() interface{} {
				ctx := make(Context)
				ctx.data = make(map[interface{}]interface{})
				return &ctx
			},
		},
		stats: &PoolStats{},
	}
}

// Get retrieves a context from the pool
func (p *ContextPool) Get() *Context {
	ctx := p.pool.Get().(*Context)
	p.stats.mu.Lock()
	p.stats.Reused++
	p.stats.mu.Unlock()
	return ctx
}

// Put returns a context to the pool
func (p *ContextPool) Put(ctx *Context) {
	if ctx == nil {
		return
	}

	ctx.Reset()

	p.stats.mu.Lock()
	p.stats.Pooled++
	p.stats.mu.Unlock()

	p.pool.Put(ctx)
}

// Stats returns the current pool statistics
func (p *ContextPool) Stats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return PoolStats{
		Allocated: p.stats.Allocated,
		Pooled:     p.stats.Pooled,
		Reused:     p.stats.Reused,
	}
}

// BufferPool manages a pool of byte buffers
type BufferPool struct {
	pool sync.Pool
	size int
	stats *PoolStats
}

// NewBufferPool creates a new buffer pool with the specified size
func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
		stats: &PoolStats{},
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() []byte {
	buf := p.pool.Get().([]byte)
	p.stats.mu.Lock()
	p.stats.Reused++
	p.stats.mu.Unlock()
	return buf[:p.size] // Reset length to full size
}

// Put returns a buffer to the pool
func (p *BufferPool) Put(buf []byte) {
	if buf == nil || cap(buf) < p.size {
		return
	}

	p.stats.mu.Lock()
	p.stats.Pooled++
	p.stats.mu.Unlock()

	p.pool.Put(buf[:p.size])
}

// Stats returns the current pool statistics
func (p *BufferPool) Stats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return PoolStats{
		Allocated: p.stats.Allocated,
		Pooled:     p.stats.Pooled,
		Reused:     p.stats.Reused,
	}
}

// Global pools for convenient access
var (
	// DefaultMessagePool is the default message pool
	DefaultMessagePool = NewMessagePool()

	// DefaultEventPool is the default event pool
	DefaultEventPool = NewEventPool()

	// DefaultContextPool is the default context pool
	DefaultContextPool = NewContextPool()

	// SmallBufferPool is a pool for small buffers (1KB)
	SmallBufferPool = NewBufferPool(1024)

	// MediumBufferPool is a pool for medium buffers (4KB)
	MediumBufferPool = NewBufferPool(4096)

	// LargeBufferPool is a pool for large buffers (32KB)
	LargeBufferPool = NewBufferPool(32768)
)
