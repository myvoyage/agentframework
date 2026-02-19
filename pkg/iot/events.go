// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package iot

import (
	"context"
	"sync"
	"time"
)

// Event represents a generic IoT event.
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventHandler handles events.
type EventHandler func(ctx context.Context, event Event)

// EventBus manages event subscriptions and publishing.
type EventBus struct {
	subscribers map[string][]EventHandler
	mutex       sync.RWMutex
	eventChan   chan Event
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	bus := &EventBus{
		subscribers: make(map[string][]EventHandler),
		eventChan:   make(chan Event, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start event processing loop
	bus.wg.Add(1)
	go bus.processEvents()

	return bus
}

// Subscribe subscribes to events of a specific type.
// If eventType is "*", subscribes to all events.
func (bus *EventBus) Subscribe(eventType string, handler EventHandler) func() {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)

	// Return unsubscribe function
	return func() {
		bus.Unsubscribe(eventType, handler)
	}
}

// Unsubscribe unsubscribes a handler from an event type.
func (bus *EventBus) Unsubscribe(eventType string, handler EventHandler) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers := bus.subscribers[eventType]
	for i, h := range handlers {
		if &h == &handler {
			// Remove handler from slice
			bus.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish publishes an event to all subscribers.
func (bus *EventBus) Publish(event Event) {
	select {
	case bus.eventChan <- event:
	default:
		// Event channel full, drop event
		// In production, you might want to log this or implement backpressure
	}
}

// PublishAsync publishes an event asynchronously.
func (bus *EventBus) PublishAsync(event Event) {
	go bus.Publish(event)
}

// processEvents processes events from the event channel.
func (bus *EventBus) processEvents() {
	defer bus.wg.Done()

	for {
		select {
		case <-bus.ctx.Done():
			// Drain remaining events
			for len(bus.eventChan) > 0 {
				event := <-bus.eventChan
				bus.dispatchEvent(event)
			}
			return

		case event := <-bus.eventChan:
			bus.dispatchEvent(event)
		}
	}
}

// dispatchEvent dispatches an event to all relevant subscribers.
func (bus *EventBus) dispatchEvent(event Event) {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	// Get handlers for this specific event type
	handlers, exists := bus.subscribers[string(event.Type)]
	if !exists {
		// No specific handlers, check for wildcard handlers
		handlers = bus.subscribers["*"]
		if len(handlers) == 0 {
			return
		}
	}

	// Add wildcard handlers if any
	if wildcardHandlers, ok := bus.subscribers["*"]; ok && string(event.Type) != "*" {
		handlers = append(handlers, wildcardHandlers...)
	}

	// Call all handlers
	for _, handler := range handlers {
		// Call handler in goroutine to avoid blocking
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and continue
					// In production, you might want to log this
				}
			}()
			h(bus.ctx, event)
		}(handler)
	}
}

// Close closes the event bus.
func (bus *EventBus) Close() error {
	bus.cancel()
	bus.wg.Wait()
	close(bus.eventChan)
	return nil
}

// EventAggregator aggregates events over time.
type EventAggregator struct {
	bus         *EventBus
	aggregators map[string]*aggregatorConfig
	mutex       sync.RWMutex
}

type aggregatorConfig struct {
	window    time.Duration
	threshold int
	events    []Event
	handler   EventHandler
	lastFlush time.Time
}

// NewEventAggregator creates a new event aggregator.
func NewEventAggregator(bus *EventBus) *EventAggregator {
	agg := &EventAggregator{
		bus:         bus,
		aggregators: make(map[string]*aggregatorConfig),
	}

	return agg
}

// AddAggregator adds an event aggregator.
// eventType: the type of event to aggregate
// window: the time window to aggregate events within
// threshold: the minimum number of events to trigger the handler
// handler: the handler to call when threshold is reached
func (agg *EventAggregator) AddAggregator(
	eventType string,
	window time.Duration,
	threshold int,
	handler EventHandler,
) {
	agg.mutex.Lock()
	defer agg.mutex.Unlock()

	agg.aggregators[eventType] = &aggregatorConfig{
		window:    window,
		threshold: threshold,
		events:    make([]Event, 0, threshold),
		handler:   handler,
		lastFlush: time.Now(),
	}

	// Subscribe to events
	agg.bus.Subscribe(eventType, func(ctx context.Context, event Event) {
		agg.handleEvent(eventType, event)
	})
}

// handleEvent handles an event for aggregation.
func (agg *EventAggregator) handleEvent(eventType string, event Event) {
	agg.mutex.Lock()
	defer agg.mutex.Unlock()

	config, exists := agg.aggregators[eventType]
	if !exists {
		return
	}

	// Add event to buffer
	config.events = append(config.events, event)

	// Check if threshold reached
	if len(config.events) >= config.threshold {
		agg.flush(eventType, config)
		return
	}

	// Check if window expired
	if time.Since(config.lastFlush) > config.window {
		agg.flush(eventType, config)
	}
}

// flush flushes aggregated events.
func (agg *EventAggregator) flush(eventType string, config *aggregatorConfig) {
	if len(config.events) == 0 {
		return
	}

	// Create aggregated event
	aggregatedEvent := Event{
		ID:        generateEventID(),
		Type:      EventType(eventType + "_aggregated"),
		Source:    "event_aggregator",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"count":  len(config.events),
			"events": config.events,
		},
	}

	// Call handler
	go config.handler(context.Background(), aggregatedEvent)

	// Clear buffer
	config.events = make([]Event, 0, config.threshold)
	config.lastFlush = time.Now()
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	return time.Now().Format("20060102150405.000000000")
}

// EventFilter filters events based on criteria.
type EventFilter struct {
	EventType    EventType
	Source       string
	DataFilters  map[string]interface{}
	TimeRange    struct {
		Start time.Time
		End   time.Time
	}
}

// Match checks if an event matches the filter criteria.
func (f *EventFilter) Match(event Event) bool {
	// Check event type
	if f.EventType != "" && event.Type != f.EventType {
		return false
	}

	// Check source
	if f.Source != "" && event.Source != f.Source {
		return false
	}

	// Check data filters
	for key, value := range f.DataFilters {
		eventValue, exists := event.Data[key]
		if !exists || eventValue != value {
			return false
		}
	}

	// Check time range
	if !f.TimeRange.Start.IsZero() && event.Timestamp.Before(f.TimeRange.Start) {
		return false
	}
	if !f.TimeRange.End.IsZero() && event.Timestamp.After(f.TimeRange.End) {
		return false
	}

	return true
}

// EventTransformer transforms events.
type EventTransformer interface {
	Transform(event Event) (Event, error)
}

// ChainTransformer chains multiple transformers.
type ChainTransformer struct {
	transformers []EventTransformer
}

// NewChainTransformer creates a new chain transformer.
func NewChainTransformer(transformers ...EventTransformer) *ChainTransformer {
	return &ChainTransformer{
		transformers: transformers,
	}
}

// Transform transforms an event through all transformers in the chain.
func (c *ChainTransformer) Transform(event Event) (Event, error) {
	var err error
	for _, transformer := range c.transformers {
		event, err = transformer.Transform(event)
		if err != nil {
			return Event{}, err
		}
	}
	return event, nil
}
