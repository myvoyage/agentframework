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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"sync"
	"time"
)

// Event represents a message in the event bus.
type Event struct {
	Topic    string
	Payload  interface{}
	Priority int // Event priority (higher = more important)
}

// EventHandler is a function that processes an event.
type EventHandler func(event Event) error

// Subscription represents a subscription to an event topic
type Subscription struct {
	id    int
	topic string
}

// EventBusOption defines options for configuring EventBus
type EventBusOption func(*MemoryEventBus)

// WithAsyncHandler enables asynchronous event handling for all subscribers
func WithAsyncHandler() EventBusOption {
	return func(bus *MemoryEventBus) {
		bus.async = true
	}
}

// WithBatchSize sets the batch size for event processing
func WithBatchSize(size int) EventBusOption {
	return func(bus *MemoryEventBus) {
		bus.batchSize = size
	}
}

// WithQueueSize sets the maximum queue size for asynchronous event processing
func WithQueueSize(size int) EventBusOption {
	return func(bus *MemoryEventBus) {
		bus.queueSize = size
	}
}

// WithMonitor enables event bus monitoring
func WithMonitor() EventBusOption {
	return func(bus *MemoryEventBus) {
		bus.monitor = true
	}
}

// EventBus provides a mechanism for decoupled communication between components.
// It supports simple publish-subscribe pattern.
type EventBus interface {
	Subscribe(topic string, handler EventHandler) Subscription
	SubscribeAsync(topic string, handler EventHandler) Subscription
	Publish(topic string, payload interface{}) map[int]error
	Unsubscribe(subscription Subscription)
	GetStats() EventBusStats
	Close() error
}

// EventBusStats represents event bus statistics
type EventBusStats struct {
	EventCount   int64 `json:"event_count"`   // Total number of events published
	HandlerCount int64 `json:"handler_count"` // Total number of events handled
	ErrorCount   int64 `json:"error_count"`   // Total number of errors occurred
	BatchSize    int   `json:"batch_size"`    // Current batch size
	QueueSize    int   `json:"queue_size"`    // Current queue size
	IsMonitoring bool  `json:"is_monitoring"` // Whether monitoring is enabled
}

// subscriptionEntry represents an entry in the subscriber list
type subscriptionEntry struct {
	id      int
	handler EventHandler
	isAsync bool
}

// MemoryEventBus is an in-memory implementation of EventBus.
type MemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]subscriptionEntry
	nextID      int
	async       bool // default event handling mode

	// Batch processing options
	batchSize int           // Number of events to process in a batch
	batchChan chan Event    // Channel for batch processing
	stopBatch chan struct{} // Stop signal for batch processing
	queueSize int           // Maximum queue size for asynchronous processing

	// Monitoring
	monitor      bool         // Enable monitoring
	eventCount   int64        // Total number of events published
	handlerCount int64        // Total number of events handled
	errorCount   int64        // Total number of errors occurred
	muStats      sync.RWMutex // Mutex for protecting stats
}

// NewMemoryEventBus creates a new instance of MemoryEventBus.
func NewMemoryEventBus(opts ...EventBusOption) *MemoryEventBus {
	bus := &MemoryEventBus{
		subscribers: make(map[string][]subscriptionEntry),
		nextID:      1,
		batchSize:   1,    // Default batch size is 1 (no batching)
		queueSize:   1000, // Default queue size is 1000
		monitor:     false,
		stopBatch:   make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(bus)
	}

	// Initialize batch channel if batch processing is enabled
	if bus.batchSize > 1 {
		bus.batchChan = make(chan Event, bus.queueSize)
		// Start batch processor goroutine
		go bus.batchProcessor()
	}

	return bus
}

// Subscribe registers a handler for a specific topic and returns a subscription that can be used to unsubscribe.
// The handler will be executed according to the default event handling mode (sync/async).
func (b *MemoryEventBus) Subscribe(topic string, handler EventHandler) Subscription {
	return b.subscribe(topic, handler, b.async)
}

// SubscribeAsync registers a handler for a specific topic that will be executed asynchronously.
func (b *MemoryEventBus) SubscribeAsync(topic string, handler EventHandler) Subscription {
	return b.subscribe(topic, handler, true)
}

// subscribe is an internal method that handles subscription logic
func (b *MemoryEventBus) subscribe(topic string, handler EventHandler, isAsync bool) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get next ID
	id := b.nextID
	b.nextID++

	entry := subscriptionEntry{
		id:      id,
		handler: handler,
		isAsync: isAsync,
	}
	b.subscribers[topic] = append(b.subscribers[topic], entry)

	return Subscription{
		id:    id,
		topic: topic,
	}
}

// Publish sends an event to all subscribers of a topic.
// Handlers are executed either synchronously or asynchronously based on their subscription configuration.
// For synchronous handlers, returns errors immediately. For asynchronous handlers, returns nil for their entries.
func (b *MemoryEventBus) Publish(topic string, payload interface{}) map[int]error {
	// Update monitoring stats
	if b.monitor {
		b.muStats.Lock()
		b.eventCount++
		b.muStats.Unlock()
	}

	event := Event{
		Topic:   topic,
		Payload: payload,
	}

	// If batch processing is enabled, send event to batch channel
	if b.batchSize > 1 {
		select {
		case b.batchChan <- event:
			// Event added to batch channel
			return nil
		default:
			// Queue is full, fallback to immediate processing
			return b.processEvent(event)
		}
	}

	// Process event immediately
	return b.processEvent(event)
}

// processEvent processes a single event
func (b *MemoryEventBus) processEvent(event Event) map[int]error {
	// Get a copy of the subscribers list while holding the read lock briefly
	b.mu.RLock()
	// Create a copy of the entries slice to reduce lock holding time
	subscribers, ok := b.subscribers[event.Topic]
	if !ok {
		b.mu.RUnlock()
		return nil
	}
	entriesCopy := make([]subscriptionEntry, len(subscribers))
	copy(entriesCopy, subscribers)
	b.mu.RUnlock()

	// If no subscribers, return early
	if len(entriesCopy) == 0 {
		return nil
	}

	// Pre-allocate errors map with expected size
	errors := make(map[int]error, len(entriesCopy))

	// Separate sync and async handlers for better performance
	var syncEntries []subscriptionEntry
	var asyncEntries []subscriptionEntry

	for _, entry := range entriesCopy {
		if entry.isAsync {
			asyncEntries = append(asyncEntries, entry)
		} else {
			syncEntries = append(syncEntries, entry)
		}
	}

	// Process synchronous handlers first (in order)
	for _, entry := range syncEntries {
		if err := entry.handler(event); err != nil {
			errors[entry.id] = err
			// Update error count if monitoring is enabled
			if b.monitor {
				b.muStats.Lock()
				b.errorCount++
				b.muStats.Unlock()
			}
		}
		// Update handler count if monitoring is enabled
		if b.monitor {
			b.muStats.Lock()
			b.handlerCount++
			b.muStats.Unlock()
		}
	}

	// Process asynchronous handlers in parallel
	for _, entry := range asyncEntries {
		go func(id int, handler EventHandler, event Event) {
			if err := handler(event); err != nil {
				// Update error count if monitoring is enabled
				if b.monitor {
					b.muStats.Lock()
					b.errorCount++
					b.muStats.Unlock()
				}
				// For async handlers, errors are currently ignored
				// Could be enhanced with error callback or logging
			}
			// Update handler count if monitoring is enabled
			if b.monitor {
				b.muStats.Lock()
				b.handlerCount++
				b.muStats.Unlock()
			}
		}(entry.id, entry.handler, event)
	}

	return errors
}

// batchProcessor processes events in batches
func (b *MemoryEventBus) batchProcessor() {
	// Use a priority queue for events
	// This allows higher priority events to be processed first
	var batch []Event
	batchTimer := time.NewTimer(100 * time.Millisecond)
	defer batchTimer.Stop()

	resetTimer := func() {
		if !batchTimer.Stop() {
			select {
			case <-batchTimer.C:
			default:
			}
		}
		batchTimer.Reset(100 * time.Millisecond)
	}

	for {
		select {
		case event := <-b.batchChan:
			// Add event to batch
			batch = append(batch, event)

			// If batch is full, process it immediately
			if len(batch) >= b.batchSize {
				b.processBatch(batch)
				batch = nil
				resetTimer()
			}
		case <-batchTimer.C:
			// Process batch after timeout even if it's not full
			if len(batch) > 0 {
				b.processBatch(batch)
				batch = nil
			}
			// Reset the timer for next batch
			resetTimer()
		case <-b.stopBatch:
			// Process remaining events before exiting
			if len(batch) > 0 {
				b.processBatch(batch)
			}
			return
		}
	}
}

// processBatch processes a batch of events
func (b *MemoryEventBus) processBatch(events []Event) {
	// Sort events by priority (higher first)
	b.sortEventsByPriority(events)

	// Process all events in the batch
	for _, event := range events {
		b.processEvent(event)
	}
}

// sortEventsByPriority sorts events by priority in descending order
func (b *MemoryEventBus) sortEventsByPriority(events []Event) {
	// Simple bubble sort for small batches
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Priority < events[j].Priority {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

// GetStats returns the current event bus statistics
func (b *MemoryEventBus) GetStats() EventBusStats {
	b.muStats.Lock()
	defer b.muStats.Unlock()

	return EventBusStats{
		EventCount:   b.eventCount,
		HandlerCount: b.handlerCount,
		ErrorCount:   b.errorCount,
		BatchSize:    b.batchSize,
		QueueSize:    b.queueSize,
		IsMonitoring: b.monitor,
	}
}

// Unsubscribe removes a subscription from the event bus.
func (b *MemoryEventBus) Unsubscribe(subscription Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entries, ok := b.subscribers[subscription.topic]
	if !ok {
		return
	}

	// Filter out the subscription to remove
	var newEntries []subscriptionEntry
	for _, entry := range entries {
		if entry.id != subscription.id {
			newEntries = append(newEntries, entry)
		}
	}

	// Update the subscribers list
	if len(newEntries) > 0 {
		b.subscribers[subscription.topic] = newEntries
	} else {
		// If no subscribers left, remove the topic entry
		delete(b.subscribers, subscription.topic)
	}
}

// Close closes the event bus and releases resources
func (b *MemoryEventBus) Close() error {
	if b.batchSize > 1 {
		close(b.stopBatch)
		close(b.batchChan)
	}
	return nil
}
