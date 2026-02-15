// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
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

package collaboration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageType represents the type of message
type MessageType int

const (
	MessageTypeTask MessageType = iota
	MessageTypeResult
	MessageTypeError
	MessageTypeQuery
	MessageTypeNotification
	MessageTypeBroadcast
)

func (mt MessageType) String() string {
	switch mt {
	case MessageTypeTask:
		return "task"
	case MessageTypeResult:
		return "result"
	case MessageTypeError:
		return "error"
	case MessageTypeQuery:
		return "query"
	case MessageTypeNotification:
		return "notification"
	case MessageTypeBroadcast:
		return "broadcast"
	default:
		return "unknown"
	}
}

// Message represents a message between agents
type Message struct {
	ID        string
	Type      MessageType
	From      string
	To        string // Empty string means broadcast
	Content   string
	Timestamp time.Time
	Metadata  map[string]interface{}
	ReplyChan chan *Message
}

// MessageBus handles message passing between agents
type MessageBus struct {
	subscribers map[string][]chan *Message
	mu          sync.RWMutex
	ctx         context.Context
	running     bool
	messageLog  []*Message
	maxLogSize  int
	logMu       sync.Mutex
}

// MessageBusConfig configures the message bus
type MessageBusConfig struct {
	MaxLogSize int
	BufferSize int
}

// NewMessageBus creates a new message bus
func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscribers: make(map[string][]chan *Message),
		running:     false,
		messageLog:  make([]*Message, 0, 100),
		maxLogSize:  1000,
	}
}

// Start starts the message bus
func (mb *MessageBus) Start(ctx context.Context) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.running {
		return fmt.Errorf("message bus already running")
	}

	mb.ctx = ctx
	mb.running = true

	return nil
}

// Stop stops the message bus
func (mb *MessageBus) Stop() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if !mb.running {
		return
	}

	mb.running = false

	// Close all subscriber channels
	for _, chans := range mb.subscribers {
		for _, ch := range chans {
			close(ch)
		}
	}

	mb.subscribers = make(map[string][]chan *Message)
}

// Subscribe subscribes a channel to messages for a specific agent
func (mb *MessageBus) Subscribe(agentName string, ch chan *Message) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if !mb.running {
		return fmt.Errorf("message bus not running")
	}

	mb.subscribers[agentName] = append(mb.subscribers[agentName], ch)

	return nil
}

// Unsubscribe unsubscribes a channel from messages for a specific agent
func (mb *MessageBus) Unsubscribe(agentName string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	delete(mb.subscribers, agentName)
}

// Publish publishes a message to a specific agent
func (mb *MessageBus) Publish(msg *Message) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if !mb.running {
		return fmt.Errorf("message bus not running")
	}

	// Log the message
	mb.logMessage(msg)

	// If To is empty, broadcast to all subscribers
	if msg.To == "" {
		return mb.broadcast(msg)
	}

	// Send to specific subscriber
	subscribers, ok := mb.subscribers[msg.To]
	if !ok || len(subscribers) == 0 {
		return fmt.Errorf("no subscribers for %s", msg.To)
	}

	for _, ch := range subscribers {
		select {
		case ch <- msg:
		case <-mb.ctx.Done():
			return mb.ctx.Err()
		default:
			// Channel is full, log warning
			continue
		}
	}

	return nil
}

// broadcast broadcasts a message to all subscribers
func (mb *MessageBus) broadcast(msg *Message) error {
	for agentName, subscribers := range mb.subscribers {
		// Skip the sender
		if agentName == msg.From {
			continue
		}

		for _, ch := range subscribers {
			select {
			case ch <- msg:
			case <-mb.ctx.Done():
				return mb.ctx.Err()
			default:
				// Channel is full, skip
				continue
			}
		}
	}

	return nil
}

// Request sends a request and waits for a response
func (mb *MessageBus) Request(ctx context.Context, msg *Message, timeout time.Duration) (*Message, error) {
	if msg.ReplyChan == nil {
		msg.ReplyChan = make(chan *Message, 1)
	}

	if err := mb.Publish(msg); err != nil {
		return nil, err
	}

	select {
	case resp := <-msg.ReplyChan:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("request timeout")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// logMessage logs a message
func (mb *MessageBus) logMessage(msg *Message) {
	mb.logMu.Lock()
	defer mb.logMu.Unlock()

	mb.messageLog = append(mb.messageLog, msg)

	// Trim log if it exceeds max size
	if len(mb.messageLog) > mb.maxLogSize {
		mb.messageLog = mb.messageLog[1:]
	}
}

// GetMessages returns the message log
func (mb *MessageBus) GetMessages() []*Message {
	mb.logMu.Lock()
	defer mb.logMu.Unlock()

	messages := make([]*Message, len(mb.messageLog))
	copy(messages, mb.messageLog)
	return messages
}

// GetMessagesForAgent returns messages for a specific agent
func (mb *MessageBus) GetMessagesForAgent(agentName string) []*Message {
	mb.logMu.Lock()
	defer mb.logMu.Unlock()

	var messages []*Message
	for _, msg := range mb.messageLog {
		if msg.From == agentName || msg.To == agentName || msg.To == "" {
			messages = append(messages, msg)
		}
	}

	return messages
}

// GetStats returns message bus statistics
func (mb *MessageBus) GetStats() *MessageBusStats {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.logMu.Lock()
	defer mb.logMu.Unlock()

	stats := &MessageBusStats{
		TotalSubscribers: len(mb.subscribers),
		TotalMessages:    len(mb.messageLog),
		SubscriberCounts: make(map[string]int),
	}

	for agentName, subscribers := range mb.subscribers {
		stats.SubscriberCounts[agentName] = len(subscribers)
	}

	return stats
}

// MessageBusStats represents message bus statistics
type MessageBusStats struct {
	TotalSubscribers int
	TotalMessages    int
	SubscriberCounts map[string]int
}
