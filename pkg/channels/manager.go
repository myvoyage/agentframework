// Package channels provides channel manager for multi-channel orchestration
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Manager manages multiple channel adapters
//
// SOLID - Single Responsibility Principle (SRP):
// Responsible only for channel lifecycle and orchestration
//
// SOLID - Open/Closed Principle (OCP):
// New channel types can be added without modifying the manager
type Manager struct {
	adapters    map[string]ChannelAdapter
	factory     AdapterFactory
	router      *Router
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	tracer      trace.Tracer
	eventBuffer chan Event
}

// ManagerConfig represents configuration for the channel manager
type ManagerConfig struct {
	// Tracer for OpenTelemetry integration
	Tracer trace.Tracer

	// EventBufferSize is the buffer size for events
	EventBufferSize int

	// RouterConfig is the configuration for the message router
	RouterConfig *RouterConfig
}

// NewManager creates a new channel manager
func NewManager(config *ManagerConfig) (*Manager, error) {
	if config == nil {
		config = &ManagerConfig{}
	}

	// Set default tracer
	tracer := config.Tracer
	if tracer == nil {
		tracer = otel.Tracer("channels-manager")
	}

	// Create router
	router, err := NewRouter(config.RouterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	m := &Manager{
		adapters:    make(map[string]ChannelAdapter),
		factory:     NewAdapterFactory(),
		router:      router,
		tracer:      tracer,
		eventBuffer: make(chan Event, config.EventBufferSize),
	}

	return m, nil
}

// Start starts the channel manager and all registered adapters
func (m *Manager) Start(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "Manager.Start")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("manager already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	// Start all adapters
	for id, adapter := range m.adapters {
		if err := adapter.Connect(m.ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to connect adapter")
			return fmt.Errorf("failed to connect adapter %s: %w", id, err)
		}
	}

	// Start event processor
	go m.processEvents(ctx)

	m.running = true
	return nil
}

// Stop stops the channel manager and all adapters
func (m *Manager) Stop(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "Manager.Stop")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	// Cancel context
	if m.cancel != nil {
		m.cancel()
	}

	// Disconnect all adapters
	var lastErr error
	for id, adapter := range m.adapters {
		if err := adapter.Disconnect(ctx); err != nil {
			span.RecordError(err)
			lastErr = fmt.Errorf("failed to disconnect adapter %s: %w", id, err)
		}
	}

	m.running = false
	close(m.eventBuffer)

	return lastErr
}

// RegisterAdapter registers a new channel adapter
func (m *Manager) RegisterAdapter(adapter ChannelAdapter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := adapter.GetChannelID()
	if _, exists := m.adapters[id]; exists {
		return fmt.Errorf("adapter %s already registered", id)
	}

	m.adapters[id] = adapter

	// Set up event handler
	adapter.SetEventHandler(m.handleEvent)

	// Auto-connect if manager is running
	if m.running {
		if err := adapter.Connect(m.ctx); err != nil {
			delete(m.adapters, id)
			return fmt.Errorf("failed to connect adapter: %w", err)
		}
	}

	return nil
}

// UnregisterAdapter unregisters a channel adapter
func (m *Manager) UnregisterAdapter(channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	adapter, exists := m.adapters[channelID]
	if !exists {
		return fmt.Errorf("adapter %s not found", channelID)
	}

	// Disconnect the adapter
	ctx := context.Background()
	if err := adapter.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect adapter: %w", err)
	}

	delete(m.adapters, channelID)
	return nil
}

// GetAdapter returns an adapter by channel ID
func (m *Manager) GetAdapter(channelID string) (ChannelAdapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[channelID]
	if !exists {
		return nil, fmt.Errorf("adapter %s not found", channelID)
	}

	return adapter, nil
}

// GetAdaptersByType returns all adapters of a specific type
func (m *Manager) GetAdaptersByType(channelType ChannelType) []ChannelAdapter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ChannelAdapter, 0)
	for _, adapter := range m.adapters {
		if adapter.GetChannelType() == channelType {
			result = append(result, adapter)
		}
	}

	return result
}

// ListAdapters returns all registered adapter IDs
func (m *Manager) ListAdapters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.adapters))
	for id := range m.adapters {
		ids = append(ids, id)
	}

	return ids
}

// SendMessage sends a message through a specific channel
func (m *Manager) SendMessage(ctx context.Context, channelID string, msg *Message, opts MessageSendOptions) (string, error) {
	ctx, span := m.tracer.Start(ctx, "Manager.SendMessage",
		trace.WithAttributes(
			attribute.String("channel_id", channelID),
			attribute.String("message_type", string(msg.Type)),
		),
	)
	defer span.End()

	adapter, err := m.GetAdapter(channelID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "adapter not found")
		return "", err
	}

	if !adapter.IsConnected() {
		err := ErrNotConnected
		span.RecordError(err)
		span.SetStatus(codes.Error, "adapter not connected")
		return "", err
	}

	// Set message direction
	msg.Direction = MessageDirectionOutgoing

	messageID, err := adapter.SendMessage(ctx, msg, opts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "send failed")
		return "", err
	}

	span.SetAttributes(attribute.String("message_id", messageID))
	return messageID, nil
}

// Broadcast sends a message to all channels of a specific type
func (m *Manager) Broadcast(ctx context.Context, channelType ChannelType, msg *Message, opts MessageSendOptions) (map[string]string, error) {
	ctx, span := m.tracer.Start(ctx, "Manager.Broadcast",
		trace.WithAttributes(
			attribute.String("channel_type", string(channelType)),
		),
	)
	defer span.End()

	adapters := m.GetAdaptersByType(channelType)
	if len(adapters) == 0 {
		return nil, fmt.Errorf("no adapters found for type %s", channelType)
	}

	results := make(map[string]string)
	var lastErr error

	for _, adapter := range adapters {
		if !adapter.IsConnected() {
			continue
		}

		msgID, err := adapter.SendMessage(ctx, msg, opts)
		if err != nil {
			lastErr = err
			continue
		}

		results[adapter.GetChannelID()] = msgID
	}

	if len(results) == 0 && lastErr != nil {
		span.RecordError(lastErr)
		span.SetStatus(codes.Error, "all sends failed")
		return nil, lastErr
	}

	return results, lastErr
}

// RouteMessage routes an incoming message through the routing system
func (m *Manager) RouteMessage(ctx context.Context, msg *Message) error {
	ctx, span := m.tracer.Start(ctx, "Manager.RouteMessage",
		trace.WithAttributes(
			attribute.String("channel_id", msg.ChannelID),
			attribute.String("message_type", string(msg.Type)),
		),
	)
	defer span.End()

	return m.router.Route(ctx, msg, m)
}

// SetMessageHandler sets the message handler for all adapters
func (m *Manager) SetMessageHandler(handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, adapter := range m.adapters {
		adapter.SetMessageHandler(handler)
	}
}

// GetStats returns statistics for all adapters
func (m *Manager) GetStats(ctx context.Context) (map[string]*ChannelStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]*ChannelStats)
	for id, adapter := range m.adapters {
		adapterStats, err := adapter.GetStats(ctx)
		if err != nil {
			continue
		}
		stats[id] = adapterStats
	}

	return stats, nil
}

// GetRouter returns the message router
func (m *Manager) GetRouter() *Router {
	return m.router
}

// AddRoutingRule adds a routing rule
func (m *Manager) AddRoutingRule(rule *RoutingRule) error {
	return m.router.AddRule(rule)
}

// RemoveRoutingRule removes a routing rule
func (m *Manager) RemoveRoutingRule(ruleID string) {
	m.router.RemoveRule(ruleID)
}

// processEvents processes events from all adapters
func (m *Manager) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-m.eventBuffer:
			if !ok {
				return
			}
			// Events can be logged or forwarded to monitoring
			_ = event
		}
	}
}

// handleEvent handles events from adapters
func (m *Manager) handleEvent(event Event) {
	select {
	case m.eventBuffer <- event:
	case <-time.After(5 * time.Second):
		// Event buffer full, drop event
	}
}

// IsRunning returns whether the manager is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetFactory returns the adapter factory
func (m *Manager) GetFactory() AdapterFactory {
	return m.factory
}

// HealthCheck checks the health of all adapters
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]error)
	for id, adapter := range m.adapters {
		_, err := adapter.GetStatus(ctx)
		results[id] = err
	}

	return results
}
