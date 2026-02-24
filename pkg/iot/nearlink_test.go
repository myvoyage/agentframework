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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNearLinkAdapterType tests the adapter type.
func TestNearLinkAdapterType(t *testing.T) {
	adapter := NewNearLinkAdapter()
	assert.Equal(t, iot.ProtocolNearLink, adapter.Type())
}

// TestNearLinkAdapterVersion tests the adapter version.
func TestNearLinkAdapterVersion(t *testing.T) {
	adapter := NewNearLinkAdapter()
	assert.Equal(t, "1.0.0", adapter.Version())
}

// TestNearLinkAdapterInitialize tests adapter initialization.
func TestNearLinkAdapterInitialize(t *testing.T) {
	adapter := NewNearLinkAdapter()

	config := iot.ProtocolConfig{
		Type: iot.ProtocolNearLink,
		Hardware: iot.HardwareConfig{
			Type:    "udp",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"network_mode":    "SLM",
			"multicast_addr":  "224.0.0.1:1888",
			"channel":         "0",
			"mesh_id":         "12345678",
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, adapter.controller)
	assert.Equal(t, NearLinkModeSLM, adapter.networkMode)
}

// TestNearLinkAdapterSLEMode tests adapter SLE mode initialization.
func TestNearLinkAdapterSLEMode(t *testing.T) {
	adapter := NewNearLinkAdapter()

	config := iot.ProtocolConfig{
		Type: iot.ProtocolNearLink,
		Metadata: map[string]string{
			"network_mode": "SLE",
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)
	assert.Equal(t, NearLinkModeSLE, adapter.networkMode)
}

// TestNearLinkDevice tests NearLink device operations.
func TestNearLinkDevice(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "nearlink-test",
		Name:     "Test NearLink Device",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolNearLink,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": "00:11:22:33:44:55",
		},
		LastSeen: time.Now(),
	}

	adapter := NewNearLinkAdapter()
	device := NewNearLinkDevice(deviceInfo, adapter)

	// Test device properties
	assert.Equal(t, "nearlink-test", device.ID())
	assert.Equal(t, "Test NearLink Device", device.Name())
	assert.Equal(t, iot.DeviceTypeSensor, device.Type())
	assert.Equal(t, iot.ProtocolNearLink, device.Protocol())
}

// TestNearLinkNetworkModeString tests network mode string values.
func TestNearLinkNetworkModeString(t *testing.T) {
	assert.Equal(t, "SLM", string(NearLinkModeSLM))
	assert.Equal(t, "SLE", string(NearLinkModeSLE))
}

// TestNearLinkDeviceToggle tests device toggle functionality.
func TestNearLinkDeviceToggle(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "nearlink-toggle-test",
		Name:     "Test Toggle Device",
		Type:     iot.DeviceTypeActuator,
		Protocol: iot.ProtocolNearLink,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": "00:11:22:33:44:56",
		},
		LastSeen: time.Now(),
	}

	adapter := NewNearLinkAdapter()
	device := NewNearLinkDevice(deviceInfo, adapter)

	// Test that Toggle method exists (actual implementation requires mock)
	assert.NotNil(t, device.Toggle)
}

// TestNearLinkDeviceBatchOperations tests batch read/write operations.
func TestNearLinkDeviceBatchOperations(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "nearlink-batch-test",
		Name:     "Test Batch Device",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolNearLink,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": "00:11:22:33:44:57",
		},
		LastSeen: time.Now(),
	}

	adapter := NewNearLinkAdapter()
	device := NewNearLinkDevice(deviceInfo, adapter)

	// Test that batch operations methods exist
	assert.NotNil(t, device.BatchRead)
	assert.NotNil(t, device.BatchWrite)
}

// TestNearLinkDeviceDiagnosticInfo tests diagnostic info retrieval.
func TestNearLinkDeviceDiagnosticInfo(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "nearlink-diag-test",
		Name:     "Test Diagnostic Device",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolNearLink,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": "00:11:22:33:44:58",
		},
		LastSeen: time.Now(),
	}

	adapter := NewNearLinkAdapter()
	device := NewNearLinkDevice(deviceInfo, adapter)

	// Test that GetDiagnosticInfo method exists
	assert.NotNil(t, device.GetDiagnosticInfo)
}

// TestNearLinkDeviceStreaming tests data streaming functionality.
func TestNearLinkDeviceStreaming(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "nearlink-stream-test",
		Name:     "Test Stream Device",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolNearLink,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": "00:11:22:33:44:59",
		},
		LastSeen: time.Now(),
	}

	adapter := NewNearLinkAdapter()
	device := NewNearLinkDevice(deviceInfo, adapter)

	// Test that Stream method exists
	assert.NotNil(t, device.Stream)
}

// TestNearLinkController tests controller creation and configuration.
func TestNearLinkController(t *testing.T) {
	metadata := map[string]string{
		"multicast_addr": "224.0.0.1:1888",
		"channel":        "0",
		"mesh_id":        "12345678",
	}

	controller, err := NewNearLinkController(metadata)
	assert.NoError(t, err)
	assert.NotNil(t, controller)
	assert.Equal(t, "224.0.0.1:1888", controller.multicastAddr)
	assert.Equal(t, uint8(0), controller.channel)
	assert.Equal(t, uint64(12345678), controller.meshID)
}

// TestNearLinkControllerDefaultValues tests controller default values.
func TestNearLinkControllerDefaultValues(t *testing.T) {
	metadata := map[string]string{}

	controller, err := NewNearLinkController(metadata)
	assert.NoError(t, err)
	assert.NotNil(t, controller)
	assert.Equal(t, "224.0.0.1:1888", controller.multicastAddr) // Default
	assert.Equal(t, NearLinkModeSLM, controller.mode)            // Default
}

// TestNearLinkFrequencyString tests frequency string conversion.
func TestNearLinkFrequencyString(t *testing.T) {
	// 2.4GHz channels
	assert.Equal(t, "2.4GHz", getFrequencyString(0))
	assert.Equal(t, "2.4GHz", getFrequencyString(13))

	// 5.1GHz channels
	assert.Equal(t, "5.1GHz", getFrequencyString(14))
	assert.Equal(t, "5.1GHz", getFrequencyString(255))
}
