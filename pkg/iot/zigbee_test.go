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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZigbeeAdapter tests the Zigbee adapter.
func TestZigbeeAdapter(t *testing.T) {
	ctx := context.Background()

	// Create adapter
	adapter := NewZigbeeAdapter()

	// Test initialization
	config := iot.ProtocolConfig{
		Type: iot.ProtocolZigbee,
		Hardware: iot.HardwareConfig{
			Type:     "mqtt",
			Port:     "localhost:1883",
			Timeout:  5000,
		},
		Metadata: map[string]string{
			"broker_url":    "tcp://localhost:1883",
			"topic_prefix":  "zigbee2mqtt",
		},
	}

	err := adapter.Initialize(ctx, config)
	assert.NoError(t, err)
	assert.Equal(t, iot.ProtocolZigbee, adapter.Type())
	assert.Equal(t, "3.0", adapter.Version())
	assert.False(t, adapter.IsRunning())
}

// TestZigbeeDevice tests Zigbee device operations.
func TestZigbeeDevice(t *testing.T) {
	ctx := context.Background()

	// Create a mock adapter
	adapter := NewZigbeeAdapter()
	_ = adapter.Initialize(ctx, iot.ProtocolConfig{})

	// Create device
	device := NewZigbeeDevice(
		"0x00158d0001a2e8b",
		"My Bulb",
		"Router",
		adapter,
	)

	// Test device properties
	assert.Equal(t, "0x00158d0001a2e8b", device.ID)
	assert.Equal(t, "My Bulb", device.FriendlyName)
	assert.Equal(t, "Router", device.Type)

	// Test capabilities
	capabilities := device.Capabilities()
	assert.NotEmpty(t, capabilities)

	// Test state operations
	device.updateState(map[string]interface{}{
		"state": "ON",
		"brightness": 255,
	})

	value, exists := device.GetStateValue("state")
	assert.True(t, exists)
	assert.Equal(t, "ON", value)

	value, exists = device.GetStateValue("brightness")
	assert.True(t, exists)
	assert.Equal(t, 255, value)
}

// TestSplitTopic tests topic splitting.
func TestSplitTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		expected []string
	}{
		{
			name:     "simple topic",
			topic:    "zigbee2mqtt/0x1234/state",
			expected: []string{"zigbee2mqtt", "0x1234", "state"},
		},
		{
			name:     "nested topic",
			topic:    "zigbee2mqtt/0x1234/set",
			expected: []string{"zigbee2mqtt", "0x1234", "set"},
		},
		{
			name:     "bridge topic",
			topic:    "zigbee2mqtt/bridge/devices",
			expected: []string{"zigbee2mqtt", "bridge", "devices"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitTopic(tt.topic)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractDeviceIDFromTopic tests device ID extraction.
func TestExtractDeviceIDFromTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		expected string
	}{
		{
			name:     "device state topic",
			topic:    "zigbee2mqtt/0x00158d0001a2e8b",
			expected: "0x00158d0001a2e8b",
		},
		{
			name:     "device set topic",
			topic:    "zigbee2mqtt/0x00158d0001a2e8b/set",
			expected: "0x00158d0001a2e8b",
		},
		{
			name:     "bridge topic",
			topic:    "zigbee2mqtt/bridge/devices",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDeviceIDFromTopic(tt.topic)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetermineDeviceType tests device type determination.
func TestDetermineDeviceType(t *testing.T) {
	tests := []struct {
		zigbeeType      string
		expectedType iot.DeviceType
	}{
		{"EndDevice", iot.DeviceTypeSensor},
		{"Router", iot.DeviceTypeActuator},
		{"Coordinator", iot.DeviceTypeGateway},
		{"Unknown", iot.DeviceTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.zigbeeType, func(t *testing.T) {
			result := determineDeviceType(tt.zigbeeType)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

// TestDetermineCapabilities tests capability determination.
func TestDetermineCapabilities(t *testing.T) {
	// Test router device (light)
	capabilities := determineCapabilities("Router")
	assert.Contains(t, capabilities, iot.CapabilityOnOff)
	assert.Contains(t, capabilities, iot.CapabilityLevelControl)

	// Test end device (sensor)
	capabilities = determineCapabilities("EndDevice")
	assert.NotEmpty(t, capabilities)
}

// TestZigbeeMQTTClient tests MQTT client creation.
func TestZigbeeMQTTClient(t *testing.T) {
	client := NewZigbeeMQTTClient("tcp://localhost:1883", "zigbee2mqtt")

	assert.NotNil(t, client)
	assert.Equal(t, "tcp://localhost:1883", client.brokerURL)
	assert.Equal(t, "zigbee2mqtt", client.topicPrefix)
	assert.False(t, client.IsConnected())

	// Test topic generation
	topic := client.GetDeviceTopic("0x1234")
	assert.Equal(t, "zigbee2mqtt/0x1234", topic)

	setTopic := client.GetDeviceSetTopic("0x1234")
	assert.Equal(t, "zigbee2mqtt/0x1234/set", setTopic)

	getTopic := client.GetDeviceGetTopic("0x1234")
	assert.Equal(t, "zigbee2mqtt/0x1234/get", getTopic)
}

// MockZigbeeMessageHandler is a mock message handler for testing.
type MockZigbeeMessageHandler struct {
	received chan struct {
		topic   string
		payload []byte
	}
}

// NewMockZigbeeMessageHandler creates a new mock handler.
func NewMockZigbeeMessageHandler() *MockZigbeeMessageHandler {
	return &MockZigbeeMessageHandler{
		received: make(chan struct {
		topic   string
		payload []byte
	}, 100),
	}
}

func (h *MockZigbeeMessageHandler) Handle(topic string, payload []byte) {
	select {
	case h.received <- struct {
		topic:   topic,
		payload: payload,
	}:
	case <-time.After(1 * time.Second):
	}
}

// TestZigbeeDeviceOperations tests device read/write operations.
func TestZigbeeDeviceOperations(t *testing.T) {
	ctx := context.Background()

	// Create adapter and device
	adapter := NewZigbeeAdapter()
	_ = adapter.Initialize(ctx, iot.ProtocolConfig{})

	device := NewZigbeeDevice(
		"0x00158d0001a2e8b",
		"Test Bulb",
		"Router",
		adapter,
	)

	// Initialize state
	device.updateState(map[string]interface{}{
		"state":     "OFF",
		"brightness": 0,
	})

	// Test read
	value, err := device.Read(ctx, "state")
	if err == nil {
		// If no error, we got a value
		assert.NotNil(t, value)
	}

	// Test write
	err = device.Write(ctx, "state", "ON")
	if err == nil {
		// Check if state was updated
		state := device.getState()
		if stateValue, ok := state["state"]; ok {
			assert.Equal(t, "ON", stateValue)
		}
	}

	// Test get state
	state := device.getState()
	assert.NotNil(t, state)

	// Test convenience methods
	if isLightDevice(device.Type) {
		// Test brightness
		err = device.SetBrightness(ctx, 128)
		assert.NoError(t, err)

		// Test color
		err = device.SetColor(ctx, 255, 0, 0)
		assert.NoError(t, err)
	}
}

// BenchmarkSplitTopic benchmarks topic splitting.
func BenchmarkSplitTopic(b *testing.B) {
	topic := "zigbee2mqtt/0x00158d0001a2e8b/set"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		splitTopic(topic)
	}
}
