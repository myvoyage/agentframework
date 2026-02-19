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
	"time"
)

// Message represents a message in the system
type Message struct {
	ID        string
	Content   []byte
	Metadata  map[string]string
	Timestamp time.Time
	Type      string
}

// Reset clears the message for reuse
func (m *Message) Reset() {
	if m == nil {
		return
	}
	m.ID = ""
	m.Content = m.Content[:0]
	for k := range m.Metadata {
		delete(m.Metadata, k)
	}
	m.Timestamp = time.Time{}
	m.Type = ""
}

// Event represents an event in the system
type Event struct {
	Type      string
	Timestamp time.Time
	Data      map[string]interface{}
	Source    string
}

// Reset clears the event for reuse
func (e *Event) Reset() {
	if e == nil {
		return
	}
	e.Type = ""
	e.Timestamp = time.Time{}
	for k := range e.Data {
		delete(e.Data, k)
	}
	e.Source = ""
}

// Context represents a context in the system
type Context struct {
	data map[interface{}]interface{}
}

// Reset clears the context for reuse
func (c *Context) Reset() {
	if c == nil {
		return
	}
	for k := range c.data {
		delete(c.data, k)
	}
}

// SetValue sets a value in the context
func (c *Context) SetValue(key, value interface{}) {
	c.data[key] = value
}

// GetValue retrieves a value from the context
func (c *Context) GetValue(key interface{}) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// GetAll returns all key-value pairs
func (c *Context) GetAll() map[interface{}]interface{} {
	return c.data
}

// Clear removes all key-value pairs
func (c *Context) Clear() {
	c.data = make(map[interface{}]interface{})
}

// PoolMetrics aggregates metrics from all pools
type PoolMetrics struct {
	MessageAllocated int64
	MessagePooled     int64
	MessageReused     int64
	EventAllocated    int64
	EventPooled       int64
	EventReused       int64
	ContextAllocated  int64
	ContextPooled     int64
	ContextReused     int64
	ReusedRate        float64
}

// GetAllMetrics returns metrics from all global pools
func GetAllMetrics() PoolMetrics {
	msgStats := DefaultMessagePool.Stats()
	eventStats := DefaultEventPool.Stats()
	ctxStats := DefaultContextPool.Stats()

	totalReused := msgStats.Reused + eventStats.Reused + ctxStats.Reused
	totalAllocated := msgStats.Allocated + eventStats.Allocated + ctxStats.Allocated

	var rate float64
	if totalAllocated > 0 {
		rate = float64(totalReused) / float64(totalAllocated) * 100
	}

	return PoolMetrics{
		MessageAllocated: msgStats.Allocated,
		MessagePooled:     msgStats.Pooled,
		MessageReused:     msgStats.Reused,
		EventAllocated:    eventStats.Allocated,
		EventPooled:       eventStats.Pooled,
		EventReused:       eventStats.Reused,
		ContextAllocated:  ctxStats.Allocated,
		ContextPooled:     ctxStats.Pooled,
		ContextReused:     ctxStats.Reused,
		ReusedRate:        rate,
	}
}

// PooledMessage is a helper for using pooled messages
type PooledMessage struct {
	msg  *Message
	pool *MessagePool
}

// NewPooledMessage creates a new pooled message
func NewPooledMessage(pool *MessagePool) *PooledMessage {
	return &PooledMessage{
		msg:  pool.Get(),
		pool: pool,
	}
}

// Message returns the underlying message
func (pm *PooledMessage) Message() *Message {
	return pm.msg
}

// Close returns the message to the pool
func (pm *PooledMessage) Close() error {
	if pm.pool != nil && pm.msg != nil {
		pm.pool.Put(pm.msg)
		pm.msg = nil
	}
	return nil
}

// PooledEvent is a helper for using pooled events
type PooledEvent struct {
	event *Event
	pool  *EventPool
}

// NewPooledEvent creates a new pooled event
func NewPooledEvent(pool *EventPool) *PooledEvent {
	return &PooledEvent{
		event: pool.Get(),
		pool:  pool,
	}
}

// Event returns the underlying event
func (pe *PooledEvent) Event() *Event {
	return pe.event
}

// Close returns the event to the pool
func (pe *PooledEvent) Close() error {
	if pe.pool != nil && pe.event != nil {
		pe.pool.Put(pe.event)
		pe.event = nil
	}
	return nil
}

// PooledContext is a helper for using pooled contexts
type PooledContext struct {
	ctx  *Context
	pool *ContextPool
}

// NewPooledContext creates a new pooled context
func NewPooledContext(pool *ContextPool) *PooledContext {
	return &PooledContext{
		ctx:  pool.Get(),
		pool: pool,
	}
}

// Context returns the underlying context
func (pc *PooledContext) Context() *Context {
	return pc.ctx
}

// Close returns the context to the pool
func (pc *PooledContext) Close() error {
	if pc.pool != nil && pc.ctx != nil {
		pc.pool.Put(pc.ctx)
		pc.ctx = nil
	}
	return nil
}
