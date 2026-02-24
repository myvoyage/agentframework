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

// TestThreadAdapterType tests the adapter type.
func TestThreadAdapterType(t *testing.T) {
	adapter := NewThreadAdapter()
	assert.Equal(t, iot.ProtocolThread, adapter.Type())
}

// TestThreadAdapterVersion tests the adapter version.
func TestThreadAdapterVersion(t *testing.T) {
	adapter := NewThreadAdapter()
	assert.Equal(t, "1.0.0", adapter.Version())
}

// TestThreadAdapterInitialize tests adapter initialization.
func TestThreadAdapterInitialize(t *testing.T) {
	adapter := NewThreadAdapter()

	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Hardware: iot.HardwareConfig{
			Type:    "border_router",
			Timeout: 5000,
		},
		Network: iot.NetworkConfig{
			Channel: 15,
		},
		Metadata: map[string]interface{}{
			"interface":     "wpan0",
			"network_name":  "TestThread",
			"pan_id":        uint16(0x1234),
			"channel":       uint8(15),
			"mesh_local_prefix": "fd00:abcd::/64",
			"on_mesh_prefix":    "2001:db8:1234::/64",
			"coap_port":        5683,
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, adapter.borderRouter)
	assert.NotNil(t, adapter.coapServer)
	assert.NotNil(t, adapter.network)
	assert.Equal(t, "TestThread", adapter.network.NetworkName)
	assert.Equal(t, uint16(0x1234), adapter.network.PanID)
	assert.Equal(t, uint8(15), adapter.network.Channel)
}

// TestThreadAdapterInitializeInvalidConfig tests initialization with invalid config.
func TestThreadAdapterInitializeInvalidConfig(t *testing.T) {
	adapter := NewThreadAdapter()

	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Hardware: iot.HardwareConfig{
			Type: "border_router",
		},
		Metadata: map[string]interface{}{
			// Missing network configuration
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid network configuration")
}

// TestThreadAdapterStartStop tests starting and stopping the adapter.
func TestThreadAdapterStartStop(t *testing.T) {
	adapter := NewThreadAdapter()

	// Initialize adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Metadata: map[string]interface{}{
			"network": map[string]interface{}{
				"network_name": "TestThread",
				"pan_id":       uint16(0x1234),
				"channel":      uint8(15),
			},
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)

	// Start adapter
	ctx := context.Background()
	err = adapter.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning())

	// Stop adapter
	err = adapter.Stop(ctx)
	assert.NoError(t, err)
}

// TestThreadAdapterDiscoverDevices tests device discovery.
func TestThreadAdapterDiscoverDevices(t *testing.T) {
	adapter := NewThreadAdapter()

	// Initialize and start adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Metadata: map[string]interface{}{
			"network": map[string]interface{}{
				"network_name": "TestThread",
				"pan_id":       uint16(0x1234),
				"channel":      uint8(15),
			},
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)

	err = adapter.Start(context.Background())
	assert.NoError(t, err)
	defer adapter.Stop(context.Background())

	// Discover devices
	devices, err := adapter.DiscoverDevices(context.Background(), 5*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	// Note: In actual environment, devices might be discovered
	// For now, we're just testing that the method works
}

// TestThreadAdapterGetDevice tests getting a device.
func TestThreadAdapterGetDevice(t *testing.T) {
	adapter := NewThreadAdapter()

	ctx := context.Background()

	// Try to get non-existent device
	device, err := adapter.GetDevice(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.Equal(t, iot.ErrDeviceNotFound, err)
}

// TestThreadAdapterListDevices tests listing devices.
func TestThreadAdapterListDevices(t *testing.T) {
	adapter := NewThreadAdapter()

	ctx := context.Background()

	// List devices (should be empty initially)
	devices, err := adapter.ListDevices(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 0)
}

// TestThreadAdapterRemoveDevice tests removing a device.
func TestThreadAdapterRemoveDevice(t *testing.T) {
	adapter := NewThreadAdapter()

	ctx := context.Background()

	// Try to remove non-existent device
	err := adapter.RemoveDevice(ctx, "non-existent")
	assert.NoError(t, err) // Should not error even if device doesn't exist
}

// TestThreadAdapterGetNetworkInfo tests getting network information.
func TestThreadAdapterGetNetworkInfo(t *testing.T) {
	adapter := NewThreadAdapter()

	// Initialize adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Metadata: map[string]interface{}{
			"network": map[string]interface{}{
				"network_name": "TestThread",
				"pan_id":       uint16(0x1234),
				"channel":      uint8(15),
			},
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)

	err = adapter.Start(context.Background())
	assert.NoError(t, err)
	defer adapter.Stop(context.Background())

	// Get network info
	info, err := adapter.GetNetworkInfo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, iot.ProtocolThread, info.Protocol)
	assert.Equal(t, uint64(0x1234), info.PanID)
	assert.Equal(t, uint64(15), info.Channel)
}

// TestThreadAdapterResetNetwork tests resetting the network.
func TestThreadAdapterResetNetwork(t *testing.T) {
	adapter := NewThreadAdapter()

	// Initialize adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Metadata: map[string]interface{}{
			"network": map[string]interface{}{
				"network_name": "TestThread",
				"pan_id":       uint16(0x1234),
				"channel":      uint8(15),
			},
		},
	}

	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)

	err = adapter.Start(context.Background())
	assert.NoError(t, err)
	defer adapter.Stop(context.Background())

	// Reset network
	err = adapter.ResetNetwork(context.Background())
	assert.NoError(t, err)
}

// TestThreadDevice tests Thread device operations.
func TestThreadDevice(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:           "thread-device-1",
		Name:         "Test Thread Device",
		Type:         iot.DeviceTypeSensor,
		Protocol:     iot.ProtocolThread,
		Manufacturer: "Test Co",
		Model:        "TD-1",
		Version:      "1.0",
		Status:       iot.DeviceStatusOnline,
		Capabilities: []iot.DeviceCapability{
			iot.CapabilitySensor,
		},
		Properties: map[string]interface{}{
			"ipv6": "fd00:abcd::1",
		},
		LastSeen: time.Now(),
	}

	adapter := NewThreadAdapter()
	device := NewThreadDevice(deviceInfo, adapter)

	// Test device properties
	assert.Equal(t, "thread-device-1", device.ID())
	assert.Equal(t, "Test Thread Device", device.Name())
	assert.Equal(t, iot.DeviceTypeSensor, device.Type())
	assert.Equal(t, iot.ProtocolThread, device.Protocol())
	assert.Equal(t, "fd00:abcd::1", device.IPv6)

	// Test connection
	ctx := context.Background()
	err := device.Connect(ctx)
	assert.NoError(t, err)
	assert.True(t, device.IsConnected())

	err = device.Disconnect(ctx)
	assert.NoError(t, err)
	assert.False(t, device.IsConnected())
}

// TestThreadDeviceReadWrite tests device read/write operations.
func TestThreadDeviceReadWrite(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "thread-device-2",
		Name:     "Test Thread Device 2",
		Type:     iot.DeviceTypeSensor,
		Protocol: iot.ProtocolThread,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"ipv6": "fd00:abcd::2",
		},
		LastSeen: time.Now(),
	}

	adapter := NewThreadAdapter()
	device := NewThreadDevice(deviceInfo, adapter)

	ctx := context.Background()

	// Connect device first
	err := device.Connect(ctx)
	assert.NoError(t, err)

	// Test read (will fail in test environment without actual CoAP server)
	_, err = device.Read(ctx, "temperature")
	// We expect this to fail since there's no actual CoAP server
	assert.Error(t, err)
}

// TestThreadDeviceExecuteCommand tests device command execution.
func TestThreadDeviceExecuteCommand(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:       "thread-device-3",
		Name:     "Test Thread Device 3",
		Type:     iot.DeviceTypeActuator,
		Protocol: iot.ProtocolThread,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"ipv6": "fd00:abcd::3",
		},
		LastSeen: time.Now(),
	}

	adapter := NewThreadAdapter()
	device := NewThreadDevice(deviceInfo, adapter)

	ctx := context.Background()
	device.Connect(ctx)

	// Test get_value command
	_, err := device.ExecuteCommand(ctx, "get_value", map[string]interface{}{
		"attribute": "temperature",
	})
	// We expect this to fail since there's no actual CoAP server
	assert.Error(t, err)

	// Test set_value command
	err = device.ExecuteCommand(ctx, "set_value", map[string]interface{}{
		"attribute": "state",
		"value":     "on",
	})
	// We expect this to fail since there's no actual CoAP server
	assert.Error(t, err)

	// Test unknown command
	_, err = device.ExecuteCommand(ctx, "unknown_command", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

// TestThreadDeviceGetInfo tests getting device information.
func TestThreadDeviceGetInfo(t *testing.T) {
	deviceInfo := &iot.DeviceInfo{
		ID:           "thread-device-4",
		Name:         "Test Thread Device 4",
		Type:         iot.DeviceTypeSensor,
		Protocol:     iot.ProtocolThread,
		Manufacturer: "Test Co",
		Model:        "TD-4",
		Version:      "1.0",
		Status:       iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"ipv6": "fd00:abcd::4",
		},
		LastSeen: time.Now(),
	}

	adapter := NewThreadAdapter()
	device := NewThreadDevice(deviceInfo, adapter)

	info := device.GetInfo()
	assert.Equal(t, "thread-device-4", info.ID)
	assert.Equal(t, "Test Thread Device 4", info.Name)
	assert.Equal(t, iot.DeviceTypeSensor, info.Type)
	assert.Equal(t, iot.ProtocolThread, info.Protocol)
	assert.Equal(t, "Test Co", info.Manufacturer)
	assert.Equal(t, "TD-4", info.Model)
}

// TestThreadDriver tests the Thread driver.
func TestThreadDriver(t *testing.T) {
	driver := NewThreadDriver()
	assert.NotNil(t, driver)
}

// TestThreadDriverConnect tests driver connection.
func TestThreadDriverConnect(t *testing.T) {
	driver := NewThreadDriver()

	config := &ThreadConfig{
		Interface:        "wpan0",
		NetworkName:      "TestThread",
		PanID:            0x1234,
		Channel:          15,
		MeshLocalPrefix:  "fd00:abcd::/64",
		OnMeshPrefix:     "2001:db8:1234::/64",
		BorderRouterAddr: "fd00:abcd::1",
	}

	ctx := context.Background()
	err := driver.Connect(ctx, config)
	assert.NoError(t, err)
	assert.Equal(t, "wpan0", driver.interfaceName)

	err = driver.Disconnect(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", driver.interfaceName)
}

// TestThreadDriverConnectInvalidConfig tests connection with invalid config.
func TestThreadDriverConnectInvalidConfig(t *testing.T) {
	driver := NewThreadDriver()

	ctx := context.Background()
	err := driver.Connect(ctx, "invalid config type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

// TestThreadDriverGetStatus tests getting driver status.
func TestThreadDriverGetStatus(t *testing.T) {
	driver := NewThreadDriver()

	ctx := context.Background()
	status, err := driver.GetStatus(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "thread", status["type"])
}

// TestThreadDriverSendCommand tests sending commands.
func TestThreadDriverSendCommand(t *testing.T) {
	driver := NewThreadDriver()

	ctx := context.Background()

	// Test command without adapter set
	_, err := driver.SendCommand(ctx, "discover_devices", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adapter not set")

	// Test unknown command
	_, err = driver.SendCommand(ctx, "unknown_command", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

// TestHelperFunctions tests helper functions.
func TestHelperFunctions(t *testing.T) {
	// Test getString
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	assert.Equal(t, "value1", getString(m, "key1", "default"))
	assert.Equal(t, "default", getString(m, "key2", "default"))
	assert.Equal(t, "default", getString(m, "nonexistent", "default"))

	// Test getInt
	assert.Equal(t, 123, getInt(m, "key2", 0))
	assert.Equal(t, 0, getInt(m, "nonexistent", 0))
	assert.Equal(t, 456, getInt(m, "key2", 456)) // Existing value takes precedence

	// Test getUint8
	m2 := map[string]interface{}{
		"key1": int(255),
		"key2": float64(128),
	}
	assert.Equal(t, uint8(255), getUint8(m2, "key1", 0))
	assert.Equal(t, uint8(128), getUint8(m2, "key2", 0))
	assert.Equal(t, uint8(0), getUint8(m2, "nonexistent", 0))

	// Test getUint16
	m3 := map[string]interface{}{
		"key1": int(65535),
		"key2": float64(32768),
	}
	assert.Equal(t, uint16(65535), getUint16(m3, "key1", 0))
	assert.Equal(t, uint16(32768), getUint16(m3, "key2", 0))
	assert.Equal(t, uint16(0), getUint16(m3, "nonexistent", 0))
}
