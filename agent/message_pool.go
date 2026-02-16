// Agent Framework - Message Pool Implementation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

// MessagePool manages a pool of messages for efficient memory usage
type MessagePool struct {
	messages map[string]*schema.Message
	mu       sync.RWMutex
}

// NewMessagePool creates a new message pool
func NewMessagePool() *MessagePool {
	return &MessagePool{
		messages: make(map[string]*schema.Message),
	}
}

// Get retrieves a message from the pool
func (p *MessagePool) Get(id string) (*schema.Message, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	msg, ok := p.messages[id]
	return msg, ok
}

// Put stores a message in the pool
func (p *MessagePool) Put(id string, msg *schema.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages[id] = msg
}

// Remove removes a message from the pool
func (p *MessagePool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.messages, id)
}

// Clear removes all messages from the pool
func (p *MessagePool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make(map[string]*schema.Message)
}

// Size returns the number of messages in the pool
func (p *MessagePool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.messages)
}
