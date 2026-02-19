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
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

// MessageRouter routes IoT messages based on rules.
type MessageRouter struct {
	routes        []MessageRoute
	eventBus      *EventBus
	manager       *IoTDeviceManager
	routeHandlers map[string]DestinationHandler
	mutex         sync.RWMutex
}

// DestinationHandler handles message delivery to a destination.
type DestinationHandler func(ctx context.Context, destination RouteDestination, msg Message) error

// NewMessageRouter creates a new message router.
func NewMessageRouter(manager *IoTDeviceManager) *MessageRouter {
	router := &MessageRouter{
		routes:        make([]MessageRoute, 0),
		eventBus:      NewEventBus(),
		manager:       manager,
		routeHandlers: make(map[string]DestinationHandler),
	}

	// Register default destination handlers
	router.RegisterDestinationHandler("device", router.handleDeviceDestination)
	router.RegisterDestinationHandler("agent", router.handleAgentDestination)
	router.RegisterDestinationHandler("webhook", router.handleWebhookDestination)
	router.RegisterDestinationHandler("mqtt", router.handleMQTTDestination)

	return router
}

// AddRoute adds a message routing rule.
func (r *MessageRouter) AddRoute(route MessageRoute) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Validate route
	if err := r.validateRoute(route); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}

	r.routes = append(r.routes, route)
	return nil
}

// RemoveRoute removes a message routing rule.
func (r *MessageRouter) RemoveRoute(routeID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, route := range r.routes {
		if route.ID == routeID {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("route %s not found", routeID)
}

// UpdateRoute updates a message routing rule.
func (r *MessageRouter) UpdateRoute(routeID string, route MessageRoute) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Validate route
	if err := r.validateRoute(route); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}

	for i := 0; i < len(r.routes); i++ {
		if r.routes[i].ID == routeID {
			r.routes[i] = route
			return nil
		}
	}

	return fmt.Errorf("route %s not found", routeID)
}

// ListRoutes lists all routing rules.
func (r *MessageRouter) ListRoutes() []MessageRoute {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	routes := make([]MessageRoute, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// RouteMessage routes a message based on configured rules.
func (r *MessageRouter) RouteMessage(ctx context.Context, msg Message) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Sort routes by priority (descending)
	sortedRoutes := make([]MessageRoute, len(r.routes))
	copy(sortedRoutes, r.routes)
	sort.Slice(sortedRoutes, func(i, j int) bool {
		return sortedRoutes[i].Priority > sortedRoutes[j].Priority
	})

	// Try to match and route message
	var lastErr error
	for _, route := range sortedRoutes {
		if !route.Enabled {
			continue
		}

		// Check source match
		if !r.matchSource(route.Source, msg) {
			continue
		}

		// Apply filters
		if !r.matchFilters(route.Filters, msg) {
			continue
		}

		// Transform message
		transformedMsg, err := r.applyTransform(route.Transform, msg)
		if err != nil {
			lastErr = err
			r.eventBus.Publish(Event{
				Type:      EventError,
				Source:    "message_router",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"route_id": route.ID,
					"error":    err.Error(),
				},
			})
			continue
		}

		// Send to destination
		if err := r.sendToDestination(ctx, route.Destination, transformedMsg); err != nil {
			lastErr = err
			r.eventBus.Publish(Event{
				Type:      EventError,
				Source:    "message_router",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"route_id": route.ID,
					"error":    err.Error(),
				},
			})
			continue
		}

		// Message routed successfully
		r.eventBus.Publish(Event{
			Type:      "message_routed",
			Source:    "message_router",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"route_id":    route.ID,
				"message_id":  msg.ID,
				"destination": route.Destination.Type,
			},
		})

		// If this is the last matching route, return success
		if !route.Destination.ContinueRouting {
			return nil
		}
	}

	if lastErr != nil {
		return lastErr
	}

	// No routes matched
	return fmt.Errorf("no matching route found for message %s", msg.ID)
}

// RegisterDestinationHandler registers a custom destination handler.
func (r *MessageRouter) RegisterDestinationHandler(destinationType string, handler DestinationHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.routeHandlers[destinationType] = handler
}

// validateRoute validates a route configuration.
func (r *MessageRouter) validateRoute(route MessageRoute) error {
	if route.ID == "" {
		return fmt.Errorf("route ID cannot be empty")
	}

	if route.Name == "" {
		return fmt.Errorf("route name cannot be empty")
	}

	// Validate destination type exists
	if _, exists := r.routeHandlers[route.Destination.Type]; !exists {
		return fmt.Errorf("unknown destination type: %s", route.Destination.Type)
	}

	return nil
}

// matchSource checks if message source matches route source.
func (r *MessageRouter) matchSource(source RouteSource, msg Message) bool {
	if source.Protocol != "" && source.Protocol != msg.Protocol {
		return false
	}

	if source.DeviceID != "" && source.DeviceID != msg.DeviceID {
		return false
	}

	// For device type matching, we need to look up the device
	if source.DeviceType != "" {
		device, err := r.manager.GetDevice(msg.DeviceID)
		if err != nil {
			return false
		}
		if device.Type() != source.DeviceType {
			return false
		}
	}

	return true
}

// matchFilters checks if message matches all filters.
func (r *MessageRouter) matchFilters(filters []MessageFilter, msg Message) bool {
	for _, filter := range filters {
		value, exists := msg.Data[filter.Attribute]
		if !exists {
			return false
		}

		if !r.matchFilter(filter, value) {
			return false
		}
	}
	return true
}

// matchFilter checks if a value matches a filter.
func (r *MessageRouter) matchFilter(filter MessageFilter, value interface{}) bool {
	switch filter.Operator {
	case "==":
		return reflect.DeepEqual(value, filter.Value)
	case "!=":
		return !reflect.DeepEqual(value, filter.Value)
	case ">":
		return r.compare(value, filter.Value, ">")
	case ">=":
		return r.compare(value, filter.Value, ">=")
	case "<":
		return r.compare(value, filter.Value, "<")
	case "<=":
		return r.compare(value, filter.Value, "<=")
	case "contains":
		return r.contains(value, filter.Value)
	case "not_contains":
		return !r.contains(value, filter.Value)
	default:
		return false
	}
}

// compare compares two values.
func (r *MessageRouter) compare(a, b interface{}, operator string) bool {
	aFloat, ok1 := toFloat64(a)
	bFloat, ok2 := toFloat64(b)

	if !ok1 || !ok2 {
		return false
	}

	switch operator {
	case ">":
		return aFloat > bFloat
	case ">=":
		return aFloat >= bFloat
	case "<":
		return aFloat < bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// contains checks if a value contains another value.
func (r *MessageRouter) contains(a, b interface{}) bool {
	aStr, ok1 := a.(string)
	bStr, ok2 := b.(string)

	if !ok1 || !ok2 {
		return false
	}

	// Simple string contains check
	return len(aStr) > 0 && len(bStr) > 0 &&
		(aStr == bStr || len(aStr) >= len(bStr) && aStr[:len(bStr)] == bStr)
}

// toFloat64 converts a value to float64.
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// applyTransform applies a transformation to a message.
func (r *MessageRouter) applyTransform(transform MessageTransform, msg Message) (Message, error) {
	if transform.Type == "" {
		return msg, nil
	}

	switch transform.Type {
	case "map":
		return r.applyMapTransform(msg, transform.Parameters)
	case "scale":
		return r.applyScaleTransform(msg, transform.Parameters)
	case "format":
		return r.applyFormatTransform(msg, transform.Parameters)
	default:
		return msg, fmt.Errorf("unknown transform type: %s", transform.Type)
	}
}

// applyMapTransform applies attribute mapping.
func (r *MessageRouter) applyMapTransform(msg Message, params map[string]interface{}) (Message, error) {
	result := Message{
		ID:        msg.ID,
		DeviceID:  msg.DeviceID,
		Protocol:  msg.Protocol,
		Timestamp: msg.Timestamp,
		Data:      make(map[string]interface{}),
	}

	// Map attributes
	if mapping, ok := params["mapping"].(map[string]interface{}); ok {
		for newKey, oldKey := range mapping {
			if oldKeyStr, ok := oldKey.(string); ok {
				if value, exists := msg.Data[oldKeyStr]; exists {
					result.Data[newKey] = value
				}
			}
		}
	} else {
		// No mapping, copy all data
		for k, v := range msg.Data {
			result.Data[k] = v
		}
	}

	return result, nil
}

// applyScaleTransform applies value scaling.
func (r *MessageRouter) applyScaleTransform(msg Message, params map[string]interface{}) (Message, error) {
	result := msg

	// Get scale parameters
	attribute, ok := params["attribute"].(string)
	if !ok {
		return msg, fmt.Errorf("scale transform requires 'attribute' parameter")
	}

	multiply, ok := params["multiply"].(float64)
	if !ok {
		return msg, fmt.Errorf("scale transform requires 'multiply' parameter")
	}

	// Apply scaling
	if value, exists := msg.Data[attribute]; exists {
		if floatVal, ok := toFloat64(value); ok {
			result.Data[attribute] = floatVal * multiply
		}
	}

	return result, nil
}

// applyFormatTransform applies value formatting.
func (r *MessageRouter) applyFormatTransform(msg Message, params map[string]interface{}) (Message, error) {
	result := msg

	// Get format parameters
	attribute, ok := params["attribute"].(string)
	if !ok {
		return msg, fmt.Errorf("format transform requires 'attribute' parameter")
	}

	format, ok := params["format"].(string)
	if !ok {
		return msg, fmt.Errorf("format transform requires 'format' parameter")
	}

	// Apply formatting
	if value, exists := msg.Data[attribute]; exists {
		result.Data[attribute] = fmt.Sprintf(format, value)
	}

	return result, nil
}

// sendToDestination sends message to destination.
func (r *MessageRouter) sendToDestination(ctx context.Context, destination RouteDestination, msg Message) error {
	handler, exists := r.routeHandlers[destination.Type]
	if !exists {
		return fmt.Errorf("no handler for destination type: %s", destination.Type)
	}

	return handler(ctx, destination, msg)
}

// Default destination handlers

// handleDeviceDestination handles device destinations.
func (r *MessageRouter) handleDeviceDestination(ctx context.Context, destination RouteDestination, msg Message) error {
	deviceID, ok := destination.Target.(string)
	if !ok {
		return fmt.Errorf("device destination requires string target (device_id)")
	}

	device, err := r.manager.GetDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Extract action and value from message metadata
	action, _ := msg.Data["action"].(string)
	value := msg.Data["value"]

	if action == "" {
		// Default to writing the first attribute
		for attr, val := range msg.Data {
			if err := device.Write(ctx, attr, val); err != nil {
				return err
			}
			break // Only write first attribute
		}
		return nil
	}

	return device.Write(ctx, action, value)
}

// handleAgentDestination handles agent destinations.
func (r *MessageRouter) handleAgentDestination(ctx context.Context, destination RouteDestination, msg Message) error {
	// Publish to event bus for agents to consume
	r.eventBus.Publish(Event{
		Type:      "agent_message",
		Source:    "message_router",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"target": destination.Target,
			"message": msg,
		},
	})

	return nil
}

// handleWebhookDestination handles webhook destinations.
func (r *MessageRouter) handleWebhookDestination(ctx context.Context, destination RouteDestination, msg Message) error {
	// In production, implement actual HTTP webhook call
	// For now, just publish an event
	r.eventBus.Publish(Event{
		Type:      "webhook_call",
		Source:    "message_router",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"url":     destination.Target,
			"message": msg,
		},
	})

	return nil
}

// handleMQTTDestination handles MQTT destinations.
func (r *MessageRouter) handleMQTTDestination(ctx context.Context, destination RouteDestination, msg Message) error {
	// In production, implement actual MQTT publish
	// For now, just publish an event
	r.eventBus.Publish(Event{
		Type:      "mqtt_publish",
		Source:    "message_router",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"topic":   destination.Target,
			"message": msg,
		},
	})

	return nil
}

// MessageRoute represents a message routing rule.
type MessageRoute struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Priority    int                 `json:"priority"`
	Source      RouteSource         `json:"source"`
	Filters     []MessageFilter     `json:"filters"`
	Transform   MessageTransform    `json:"transform,omitempty"`
	Destination RouteDestination    `json:"destination"`
	Enabled     bool                `json:"enabled"`
	Metadata    map[string]string    `json:"metadata,omitempty"`
}

// RouteSource defines the source of messages for a route.
type RouteSource struct {
	Protocol   ProtocolType `json:"protocol,omitempty"`
	DeviceID   string       `json:"device_id,omitempty"`
	DeviceType DeviceType   `json:"device_type,omitempty"`
}

// MessageFilter defines a message filter.
type MessageFilter struct {
	Attribute string                 `json:"attribute"`
	Operator  string                 `json:"operator"`
	Value     interface{}            `json:"value"`
	Type      string                 `json:"type,omitempty"`
}

// MessageTransform defines a message transformation.
type MessageTransform struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// RouteDestination defines where to route messages.
type RouteDestination struct {
	Type            string                 `json:"type"`
	Target          interface{}            `json:"target"`
	ContinueRouting bool                   `json:"continue_routing"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
