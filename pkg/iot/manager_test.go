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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBaseDevice tests the BaseDevice implementation.
func TestBaseDevice(t *testing.T) {
	info := &DeviceInfo{
		ID:           "test-device-1",
		Name:         "Test Device",
		Type:         DeviceTypeSensor,
		Protocol:     ProtocolZigbee,
		Manufacturer: "Test Co",
		Model:        "TD-1",
		Version:      "1.0",
		Status:       DeviceStatusOnline,
		Capabilities: []DeviceCapability{CapabilitySensor},
		Properties: map[string]interface{}{
			"temperature": 25.5,
		},
	}

	device := NewBaseDevice(info)

	// Test identity methods
	assert.Equal(t, "test-device-1", device.ID())
	assert.Equal(t, "Test Device", device.Name())
	assert.Equal(t, DeviceTypeSensor, device.Type())
	assert.Equal(t, ProtocolZigbee, device.Protocol())

	// Test connection status
	assert.False(t, device.IsConnected())
	assert.Equal(t, DeviceStatusOnline, device.Status())

	// Test device information
	assert.Equal(t, "Test Co", device.Manufacturer())
	assert.Equal(t, "TD-1", device.Model())
	assert.Equal(t, "1.0", device.Version())
	assert.Contains(t, device.Capabilities(), CapabilitySensor)

	// Test metadata
	assert.WithinDuration(t, time.Now(), device.LastSeen(), time.Second)
	assert.Equal(t, 0, device.SignalStrength())
	assert.Equal(t, uint8(255), device.BatteryLevel())

	// Test config
	config, err := device.GetConfig(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 25.5, config["temperature"])

	err = device.SetConfig(context.Background(), map[string]interface{}{
		"humidity": 60.0,
	})
	assert.NoError(t, err)

	config, err = device.GetConfig(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 60.0, config["humidity"])
}

// TestEventBus tests the EventBus.
func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// Test subscribe and publish
	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) {
		received <- event
	}

	unsubscribe := bus.Subscribe("test_event", handler)
	defer unsubscribe()

	event := Event{
		ID:        "test-1",
		Type:      "test_event",
		Source:    "test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
	}

	bus.Publish(event)

	select {
	case receivedEvent := <-received:
		assert.Equal(t, "test-1", receivedEvent.ID)
		assert.Equal(t, "test_event", string(receivedEvent.Type))
	case <-time.After(1 * time.Second):
		t.Fatal("did not receive event")
	}
}

// TestEventBusWildcard tests wildcard subscription.
func TestEventBusWildcard(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	receivedCount := 0
	var mutex sync.Mutex
	var wg sync.WaitGroup

	handler := func(ctx context.Context, event Event) {
		mutex.Lock()
		receivedCount++
		mutex.Unlock()
		wg.Done()
	}

	// Subscribe to all events
	unsubscribe := bus.Subscribe("*", handler)
	defer unsubscribe()

	// Set up wait group for 3 events
	wg.Add(3)

	// Publish multiple events
	bus.Publish(Event{Type: "event1", Source: "test", Timestamp: time.Now()})
	bus.Publish(Event{Type: "event2", Source: "test", Timestamp: time.Now()})
	bus.Publish(Event{Type: "event3", Source: "test", Timestamp: time.Now()})

	// Wait for all events to be processed
	wg.Wait()

	mutex.Lock()
	count := receivedCount
	mutex.Unlock()

	assert.Equal(t, 3, count)
}

// TestDeviceRegistry tests the DeviceRegistry.
func TestDeviceRegistry(t *testing.T) {
	registry := NewDeviceRegistry()
	defer registry.Close()

	// Test register
	info := &DeviceInfo{
		ID:       "device-1",
		Name:     "Device 1",
		Type:     DeviceTypeSensor,
		Protocol: ProtocolZigbee,
		Status:   DeviceStatusOnline,
		Capabilities: []DeviceCapability{
			CapabilitySensor,
			CapabilityOnOff,
		},
		Properties: map[string]interface{}{
			"temperature": 25.0,
		},
		LastSeen: time.Now(),
	}

	err := registry.Register(info)
	assert.NoError(t, err)

	// Test get
	retrieved, err := registry.Get("device-1")
	assert.NoError(t, err)
	assert.Equal(t, "Device 1", retrieved.Name)
	assert.Equal(t, ProtocolZigbee, retrieved.Protocol)

	// Test count
	assert.Equal(t, 1, registry.Count())

	// Test list
	devices := registry.List()
	assert.Len(t, devices, 1)

	// Test query
	results := registry.Query(QueryCriteria{
		Protocol: ProtocolZigbee,
	})
	assert.Len(t, results, 1)

	results = registry.Query(QueryCriteria{
		Protocol: ProtocolZWave,
	})
	assert.Len(t, results, 0)

	// Test update status
	err = registry.UpdateStatus("device-1", DeviceStatusOffline)
	assert.NoError(t, err)

	retrieved, _ = registry.Get("device-1")
	assert.Equal(t, DeviceStatusOffline, retrieved.Status)

	// Test update last seen
	newTime := time.Now()
	err = registry.UpdateLastSeen("device-1", newTime)
	assert.NoError(t, err)

	// Test unregister
	err = registry.Unregister("device-1")
	assert.NoError(t, err)
	assert.Equal(t, 0, registry.Count())
}

// TestMessageRouter tests the MessageRouter.
func TestMessageRouter(t *testing.T) {
	ctx := context.Background()
	manager := NewIoTDeviceManager()
	defer manager.Close(ctx)

	router := NewMessageRouter(manager)

	// Create a test route
	route := MessageRoute{
		ID:       "route-1",
		Name:     "Test Route",
		Priority: 10,
		Source: RouteSource{
			Protocol: ProtocolZigbee,
		},
		Filters: []MessageFilter{
			{
				Attribute: "temperature",
				Operator:  ">",
				Value:     20.0,
			},
		},
		Destination: RouteDestination{
			Type:   "agent",
			Target: "test-agent",
		},
		Enabled: true,
	}

	err := router.AddRoute(route)
	assert.NoError(t, err)

	// Test list routes
	routes := router.ListRoutes()
	assert.Len(t, routes, 1)
	assert.Equal(t, "route-1", routes[0].ID)

	// Test remove route
	err = router.RemoveRoute("route-1")
	assert.NoError(t, err)
	routes = router.ListRoutes()
	assert.Len(t, routes, 0)
}

// TestMessageRouterMatchFilters tests filter matching.
func TestMessageRouterMatchFilters(t *testing.T) {
	ctx := context.Background()
	manager := NewIoTDeviceManager()
	defer manager.Close(ctx)

	router := NewMessageRouter(manager)

	tests := []struct {
		name     string
		filter   MessageFilter
		value    interface{}
		expected bool
	}{
		{
			name: "equal match",
			filter: MessageFilter{
				Attribute: "state",
				Operator:  "==",
				Value:     "on",
			},
			value:    "on",
			expected: true,
		},
		{
			name: "equal no match",
			filter: MessageFilter{
				Attribute: "state",
				Operator:  "==",
				Value:     "off",
			},
			value:    "on",
			expected: false,
		},
		{
			name: "greater than match",
			filter: MessageFilter{
				Attribute: "temperature",
				Operator:  ">",
				Value:     20.0,
			},
			value:    25.0,
			expected: true,
		},
		{
			name: "greater than no match",
			filter: MessageFilter{
				Attribute: "temperature",
				Operator:  ">",
				Value:     30.0,
			},
			value:    25.0,
			expected: false,
		},
		{
			name: "less than match",
			filter: MessageFilter{
				Attribute: "temperature",
				Operator:  "<",
				Value:     30.0,
			},
			value:    25.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.matchFilter(tt.filter, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProtocolError tests protocol error.
func TestProtocolError(t *testing.T) {
	err := &ProtocolError{
		Code:    "TEST_ERROR",
		Message: "test error message",
	}

	assert.Contains(t, err.Error(), "TEST_ERROR")
	assert.Contains(t, err.Error(), "test error message")

	// Test with wrapped error
	wrappedErr := &ProtocolError{
		Code:    "WRAPPED_ERROR",
		Message: "wrapped message",
		Err:     err,
	}

	assert.Contains(t, wrappedErr.Error(), "WRAPPED_ERROR")
	assert.Contains(t, wrappedErr.Error(), "wrapped message")
	assert.Contains(t, wrappedErr.Error(), "TEST_ERROR")

	// Test unwrap
	unwrapped := wrappedErr.Unwrap()
	assert.Equal(t, err, unwrapped)
}

// TestIoTDeviceManager tests the IoT device manager.
func TestIoTDeviceManager(t *testing.T) {
	ctx := context.Background()
	manager := NewIoTDeviceManager()
	defer manager.Close(ctx)

	// Test no adapters initially
	adapters := manager.ListAdapters()
	assert.Empty(t, adapters)

	// Register mock adapter
	mockAdapter := &MockAdapter{
		BaseAdapter: NewBaseAdapter(ProtocolZigbee, "1.0"),
		devices:     make(map[string]IoTDevice),
	}

	err := manager.RegisterAdapter(mockAdapter)
	assert.NoError(t, err)

	adapters = manager.ListAdapters()
	assert.Len(t, adapters, 1)
	assert.Contains(t, adapters, ProtocolZigbee)

	// Test get adapter
	retrieved, err := manager.GetAdapter(ProtocolZigbee)
	assert.NoError(t, err)
	assert.Equal(t, ProtocolZigbee, retrieved.Type())

	// Test unregister
	err = manager.UnregisterAdapter(ProtocolZigbee)
	assert.NoError(t, err)

	adapters = manager.ListAdapters()
	assert.Empty(t, adapters)
}

// TestTypes tests type constants and conversions.
func TestTypes(t *testing.T) {
	// Test protocol types
	assert.Equal(t, "zigbee", string(ProtocolZigbee))
	assert.Equal(t, "zwave", string(ProtocolZWave))
	assert.Equal(t, "thread", string(ProtocolThread))

	// Test device types
	assert.Equal(t, "sensor", string(DeviceTypeSensor))
	assert.Equal(t, "actuator", string(DeviceTypeActuator))

	// Test device status
	assert.Equal(t, "online", string(DeviceStatusOnline))
	assert.Equal(t, "offline", string(DeviceStatusOffline))

	// Test device capabilities
	assert.Equal(t, "on_off", string(CapabilityOnOff))
	assert.Equal(t, "sensor", string(CapabilitySensor))

	// Test event types
	assert.Equal(t, "device_discovered", string(EventDeviceDiscovered))
	assert.Equal(t, "data_received", string(EventDataReceived))
}

// BenchmarkEventBus benchmarks event bus performance.
func BenchmarkEventBus(b *testing.B) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, event Event) {}

	bus.Subscribe("test", handler)

	event := Event{
		ID:        "test",
		Type:      "test",
		Source:    "bench",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 1},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

// MockAdapter is a mock adapter for testing.
type MockAdapter struct {
	*BaseAdapter
	devices map[string]IoTDevice
}

func (m *MockAdapter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*DeviceInfo, error) {
	return []*DeviceInfo{}, nil
}

func (m *MockAdapter) StartPairing(ctx context.Context, timeout time.Duration) (*PairingResult, error) {
	return &PairingResult{Success: false}, nil
}

func (m *MockAdapter) CancelPairing(ctx context.Context) error {
	return nil
}

func (m *MockAdapter) RemoveDevice(ctx context.Context, deviceID string) error {
	return nil
}

func (m *MockAdapter) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	return &NetworkInfo{}, nil
}

func (m *MockAdapter) ResetNetwork(ctx context.Context) error {
	return nil
}

func (m *MockAdapter) ListDevices(ctx context.Context) ([]IoTDevice, error) {
	devices := make([]IoTDevice, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices, nil
}
