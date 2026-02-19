// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package adapters

import (
	"context"
	"testing"
	"time"

	"AgentFramework/pkg/iot"
	"github.com/stretchr/testify/assert"
)

// TestZWaveAdapterType tests the adapter type.
func TestZWaveAdapterType(t *testing.T) {
	adapter := NewZWaveAdapter()
	assert.Equal(t, iot.ProtocolZWave, adapter.Type())
}

// TestZWaveAdapterVersion tests the adapter version.
func TestZWaveAdapterVersion(t *testing.T) {
	adapter := NewZWaveAdapter()
	assert.Equal(t, "1.0.0", adapter.Version())
}

// TestZWaveAdapterInitialize tests adapter initialization.
func TestZWaveAdapterInitialize(t *testing.T) {
	adapter := NewZWaveAdapter()

	config := iot.ProtocolConfig{
		Type: iot.ProtocolZWave,
		Hardware: iot.HardwareConfig{
			Type:    "websocket",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"ws_url": "ws://localhost:3000",
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, adapter.jsClient)
}

// TestZWaveDevice tests Z-Wave device operations.
func TestZWaveDevice(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "zwave-node-2",
		Name:     "Test Z-Wave Device",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolZWave,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"node_id": uint8(2),
		},
		LastSeen: time.Now(),
	}

	adapter := NewZWaveAdapter()
	device := NewZWaveDevice(deviceInfo, adapter)

	// Test device properties
	assert.Equal(t, "zwave-node-2", device.ID())
	assert.Equal(t, "Test Z-Wave Device", device.Name())
	assert.Equal(t, iot.DeviceTypeSensor, device.Type())
	assert.Equal(t, iot.ProtocolZWave, device.Protocol())
	assert.Equal(t, uint8(2), device.nodeID)
}

// TestMapAttributeToCommandClass tests attribute mapping.
func TestMapAttributeToCommandClass(t *testing.T) {
	tests := []struct {
		attribute     string
		expectedClass string
	}{
		{"state", "0x20"},
		{"brightness", "0x26"},
		{"color", "0x33"},
		{"temperature", "0x31"},
		{"humidity", "0x31"},
		{"battery_level", "0x80"},
		{"location", "0x84"},
		{"manufacturer", "0x72"},
		{"version", "0x86"},
	}

	for _, tt := range tests {
		t.Run(tt.attribute, func(t *testing.T) {
			result := mapAttributeToCommandClass(tt.attribute)
			assert.Equal(t, tt.expectedClass, result)
		})
	}
}

// TestHelperFunctions tests helper functions.
func TestHelperFunctions(t *testing.T) {
	// Test getUint8FromMessage
	message := map[string]interface{}{
		"nodeId": float64(5),
	}
	assert.Equal(t, uint8(5), getUint8FromMessage(message, "nodeId"))

	// Test with string
	message2 := map[string]interface{}{
		"nodeId": "10",
	}
	assert.Equal(t, uint8(10), getUint8FromMessage(message2, "nodeId"))

	// Test getUint8FromProps
	props := map[string]interface{}{
		"node_id": uint8(3),
	}
	assert.Equal(t, uint8(3), getUint8FromProps(props, "node_id"))

	// Test with int
	props2 := map[string]interface{}{
		"node_id": 7,
	}
	assert.Equal(t, uint8(7), getUint8FromProps(props2, "node_id"))
}

// TestGetZWaveDeviceType tests device type determination.
func TestGetZWaveDeviceType(t *testing.T) {
	// Test sensor device
	sensorNode := map[string]interface{}{
		"basic": float64(0x0031), // Sensor multilevel
	}
	assert.Equal(t, iot.DeviceTypeSensor, getZWaveDeviceType(sensorNode))

	// Test actuator device
	actuatorNode := map[string]interface{}{
		"basic": float64(0x0020), // Switch
	}
	assert.Equal(t, iot.DeviceTypeActuator, getZWaveDeviceType(actuatorNode))
}

// TestGetNodeName tests node name generation.
func TestGetNodeName(t *testing.T) {
	// Test with name
	nodeWithName := map[string]interface{}{
		"nodeId": float64(5),
		"name":   "Living Room Switch",
	}
	assert.Equal(t, "Living Room Switch", getNodeName(nodeWithName))

	// Test without name
	nodeWithoutName := map[string]interface{}{
		"nodeId": float64(10),
	}
	assert.Equal(t, "Z-Wave Node 10", getNodeName(nodeWithoutName))
}
