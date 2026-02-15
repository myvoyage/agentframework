// Agent Framework - Dynamic EventBus Implementation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package event

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// DynamicEventBusConfig contains configuration for dynamic event bus
type DynamicEventBusConfig struct {
	InitialQueueSize int     `json:"initial_queue_size"` // Initial queue size
	MaxQueueSize     int     `json:"max_queue_size"`     // Maximum queue size (0 = unlimited)
	ResizeThreshold  float64 `json:"resize_threshold"`  // Threshold (0.0-1.0) to trigger resize
	ResizeMultiplier float64 `json:"resize_multiplier"` // Multiplier for queue resizing (e.g., 2.0 = double)
	Monitoring       bool    `json:"monitoring"`        // Enable performance monitoring
}

// DefaultDynamicEventBusConfig returns default configuration for dynamic event bus
func DefaultDynamicEventBusConfig() DynamicEventBusConfig {
	return DynamicEventBusConfig{
		InitialQueueSize: 1000,
		MaxQueueSize:     100000,
		ResizeThreshold:  0.8,   // Resize when 80% full
		ResizeMultiplier: 2.0,  // Double the size
		Monitoring:       true,
	}
}

// DynamicEventBus is an enhanced event bus with dynamic queue resizing
type DynamicEventBus struct {
	cfg    DynamicEventBusConfig
	ctx    context.Context
	cancel context.CancelFunc

	// Dynamic queue
	queue      chan Event
	queueSize  atomic.Int32 // Current queue size
	muQueue   sync.RWMutex

	// Subscribers management
	mu          sync.RWMutex
	subscribers map[string][]subscriptionEntry
	nextID      int

	// Monitoring
	eventCount   atomic.Int64
	handlerCount atomic.Int64
	errorCount   atomic.Int64
	resizeCount atomic.Int64

	// Monitoring stats
	lastResizeTime time.Time
	avgQueueUsage  float64
	peakQueueSize  int32
}

// NewDynamicEventBus creates a new dynamic event bus
func NewDynamicEventBus(cfg DynamicEventBusConfig) *DynamicEventBus {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.InitialQueueSize <= 0 {
		cfg.InitialQueueSize = 1000
	}
	if cfg.ResizeThreshold <= 0 || cfg.ResizeThreshold > 1.0 {
		cfg.ResizeThreshold = 0.8
	}
	if cfg.ResizeMultiplier < 1.0 {
		cfg.ResizeMultiplier = 2.0
	}

	bus := &DynamicEventBus{
		cfg:         cfg,
		ctx:          ctx,
		cancel:       cancel,
		queue:        make(chan Event, cfg.InitialQueueSize),
		subscribers:  make(map[string][]subscriptionEntry),
		nextID:       1,
		lastResizeTime: time.Now(),
	}

	// Initialize queue size
	bus.queueSize.Store(int32(cfg.InitialQueueSize))
	bus.peakQueueSize = int32(cfg.InitialQueueSize)

	// Start background goroutines
	go bus.eventDispatcher()
	go bus.resizeMonitor()

	return bus
}

// Subscribe registers a handler for a specific topic
func (b *DynamicEventBus) Subscribe(topic string, handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	entry := subscriptionEntry{
		id:      id,
		handler: handler,
		isAsync: false,
	}
	b.subscribers[topic] = append(b.subscribers[topic], entry)

	return Subscription{
		id:    id,
		topic: topic,
	}
}

// SubscribeAsync registers an asynchronous handler for a specific topic
func (b *DynamicEventBus) SubscribeAsync(topic string, handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	entry := subscriptionEntry{
		id:      id,
		handler: handler,
		isAsync: true,
	}
	b.subscribers[topic] = append(b.subscribers[topic], entry)

	return Subscription{
		id:    id,
		topic: topic,
	}
}

// Publish sends an event to all subscribers of a topic
func (b *DynamicEventBus) Publish(topic string, payload interface{}) map[int]error {
	// Update event count
	b.eventCount.Add(1)

	event := Event{
		Topic:   topic,
		Payload: payload,
	}

	// Try to add to queue
	select {
	case b.queue <- event:
		// Event successfully queued
		return nil
	default:
		// Queue is full, try to resize
		if b.tryResize() {
			// Try again after resize
			select {
			case b.queue <- event:
				return nil
			default:
				// Still full, process synchronously
				return b.processEvent(event)
			}
		}
		// Cannot resize, process synchronously
		return b.processEvent(event)
	}
}

// processEvent processes a single event immediately
func (b *DynamicEventBus) processEvent(event Event) map[int]error {
	b.mu.RLock()
	subscribers, ok := b.subscribers[event.Topic]
	if !ok {
		b.mu.RUnlock()
		return nil
	}
	entriesCopy := make([]subscriptionEntry, len(subscribers))
	copy(entriesCopy, subscribers)
	b.mu.RUnlock()

	if len(entriesCopy) == 0 {
		return nil
	}

	errors := make(map[int]error)

	// Process handlers
	for _, entry := range entriesCopy {
		if entry.isAsync {
			go func(h EventHandler) {
				if err := h(event); err != nil {
					b.errorCount.Add(1)
				}
				b.handlerCount.Add(1)
			}(entry.handler)
		} else {
			if err := entry.handler(event); err != nil {
				errors[entry.id] = err
				b.errorCount.Add(1)
			}
			b.handlerCount.Add(1)
		}
	}

	return errors
}

// eventDispatcher continuously processes events from the queue
func (b *DynamicEventBus) eventDispatcher() {
	for {
		select {
		case event := <-b.queue:
			b.processEvent(event)
		case <-b.ctx.Done():
			return
		}
	}
}

// resizeMonitor monitors queue usage and triggers resizes
func (b *DynamicEventBus) resizeMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.checkAndResize()
		case <-b.ctx.Done():
			return
		}
	}
}

// checkAndResize checks if queue needs to be resized
func (b *DynamicEventBus) checkAndResize() {
	currentSize := b.queueSize.Load()
	currentLen := int32(len(b.queue))

	// Update peak queue size
	if currentLen > b.peakQueueSize {
		b.peakQueueSize = currentLen
	}

	// Calculate usage ratio
	usageRatio := float64(currentLen) / float64(currentSize)

	// Update average queue usage (simple moving average)
	b.avgQueueUsage = 0.9*b.avgQueueUsage + 0.1*usageRatio

	// Check if resize is needed (expand when usage exceeds threshold)
	if usageRatio > b.cfg.ResizeThreshold {
		newSize := int(float64(currentSize) * b.cfg.ResizeMultiplier)
		b.resizeQueue(newSize)
	}
}

// tryResize attempts to resize the queue
func (b *DynamicEventBus) tryResize() bool {
	currentSize := b.queueSize.Load()

	// Check if we can expand
	if b.cfg.MaxQueueSize > 0 && int(currentSize) >= b.cfg.MaxQueueSize {
		return false // Already at max size
	}

	newSize := int(float64(currentSize) * b.cfg.ResizeMultiplier)
	if b.cfg.MaxQueueSize > 0 && newSize > b.cfg.MaxQueueSize {
		newSize = b.cfg.MaxQueueSize
	}

	if newSize <= int(currentSize) {
		return false // No growth possible
	}

	return b.resizeQueue(newSize)
}

// resizeQueue resizes the queue to a new size
func (b *DynamicEventBus) resizeQueue(newSize int) bool {
	b.muQueue.Lock()
	defer b.muQueue.Unlock()

	currentSize := int(b.queueSize.Load())

	if newSize == currentSize {
		return false
	}

	// Create new queue
	newQueue := make(chan Event, newSize)

	// Drain old queue
	close(b.queue)
	drained := false
	for !drained {
		select {
		case event, ok := <-b.queue:
			if ok {
				// Transfer event to new queue
				select {
				case newQueue <- event:
				case <-b.ctx.Done():
					return false
				}
			} else {
				drained = true
			}
		case <-b.ctx.Done():
			return false
		}
	}

	// Replace queue
	b.queue = newQueue
	b.queueSize.Store(int32(newSize))
	b.lastResizeTime = time.Now()
	b.resizeCount.Add(1)

	return true
}

// Unsubscribe removes a subscription
func (b *DynamicEventBus) Unsubscribe(subscription Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entries, ok := b.subscribers[subscription.topic]
	if !ok {
		return
	}

	var newEntries []subscriptionEntry
	for _, entry := range entries {
		if entry.id != subscription.id {
			newEntries = append(newEntries, entry)
		}
	}

	if len(newEntries) > 0 {
		b.subscribers[subscription.topic] = newEntries
	} else {
		delete(b.subscribers, subscription.topic)
	}
}

// DynamicEventBusStats represents dynamic event bus statistics
type DynamicEventBusStats struct {
	EventCount      int64     `json:"event_count"`
	HandlerCount    int64     `json:"handler_count"`
	ErrorCount      int64     `json:"error_count"`
	ResizeCount     int64     `json:"resize_count"`
	CurrentQueueLen int       `json:"current_queue_len"`
	CurrentQueueCap int       `json:"current_queue_cap"`
	PeakQueueSize   int32     `json:"peak_queue_size"`
	AvgQueueUsage   float64   `json:"avg_queue_usage"`
	LastResizeTime  time.Time `json:"last_resize_time"`
}

// GetStats returns current statistics
func (b *DynamicEventBus) GetStats() DynamicEventBusStats {
	currentLen := len(b.queue)
	currentCap := int(b.queueSize.Load())

	return DynamicEventBusStats{
		EventCount:      b.eventCount.Load(),
		HandlerCount:    b.handlerCount.Load(),
		ErrorCount:      b.errorCount.Load(),
		ResizeCount:     b.resizeCount.Load(),
		CurrentQueueLen: currentLen,
		CurrentQueueCap: currentCap,
		PeakQueueSize:   b.peakQueueSize,
		AvgQueueUsage:   b.avgQueueUsage,
		LastResizeTime:  b.lastResizeTime,
	}
}

// Close closes the dynamic event bus
func (b *DynamicEventBus) Close() error {
	b.cancel()

	// Drain remaining events
	close(b.queue)
	for range b.queue {
		// Drain events without processing
	}

	return nil
}

// Helper function to create dynamic event bus with default config
func NewDefaultDynamicEventBus() *DynamicEventBus {
	return NewDynamicEventBus(DefaultDynamicEventBusConfig())
}
