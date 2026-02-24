// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package drivers

import (
	"context"
	"fmt"
	"time"

	"AgentFramework/pkg/beads/hardware"
	"AgentFramework/pkg/iot"
	iotadapters "AgentFramework/pkg/iot"
)

// ZWaveDriver implements HardwareController for Z-Wave devices.
type ZWaveDriver struct {
	adapter   *iot.ProtocolAdapter
	wsURL     string
	isRunning bool
}

// NewZWaveDriver creates a new Z-Wave driver.
func NewZWaveDriver() *ZWaveDriver {
	return &ZWaveDriver{}
}

// Connect connects to Z-Wave hardware (via Z-Wave JS).
func (d *ZWaveDriver) Connect(ctx context.Context, config interface{}) error {
	cfg, ok := config.(*ZWaveConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected ZWaveConfig")
	}

	d.wsURL = cfg.WSURL

	// The actual connection is handled by the ZWaveAdapter
	// This driver just stores the configuration
	return nil
}

// Disconnect disconnects from the Z-Wave hardware.
func (d *ZWaveDriver) Disconnect(ctx context.Context) error {
	d.wsURL = ""
	d.isRunning = false
	return nil
}

// SendCommand sends a command to the Z-Wave device.
func (d *ZWaveDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	switch cmd {
	case "discover_devices":
		// Discover devices on the Z-Wave network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.DiscoverDevices(ctx, 0)

	case "start_inclusion":
		// Start inclusion mode
		enable := params["enable"].(bool)
		_ = enable // Will be used in actual implementation
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		return adapter.StartPairing(ctx, 60*time.Second)

	case "stop_inclusion":
		// Stop inclusion mode
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return nil, adapter.CancelPairing(ctx)

	case "get_devices":
		// Get all devices
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.ListDevices(ctx)

	case "get_network_info":
		// Get network information
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.GetNetworkInfo(ctx)

	case "heal_network":
		// Heal Z-Wave network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return nil, adapter.ResetNetwork(ctx)

	case "get_node_info":
		// Get node information
		nodeIDFloat, ok := params["node_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("node_id parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		deviceID := fmt.Sprintf("zwave-node-%d", int(nodeIDFloat))
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if zwaveDev, ok := device.(*iotadapters.ZWaveDevice); ok {
			return zwaveDev.GetNodeInfo(ctx)
		}
		return nil, fmt.Errorf("device is not a Z-Wave device")

	case "set_value":
		// Set value on a device
		nodeIDFloat, ok := params["node_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("node_id parameter required")
		}
		commandClass, ok := params["command_class"].(string)
		if !ok {
			return nil, fmt.Errorf("command_class parameter required")
		}
		value, ok := params["value"]
		if !ok {
			return nil, fmt.Errorf("value parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		deviceID := fmt.Sprintf("zwave-node-%d", int(nodeIDFloat))
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if zwaveDev, ok := device.(*iotadapters.ZWaveDevice); ok {
			return nil, zwaveDev.SetValue(ctx, commandClass, value)
		}
		return nil, fmt.Errorf("device is not a Z-Wave device")

	case "get_value":
		// Get value from a device
		nodeIDFloat, ok := params["node_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("node_id parameter required")
		}
		commandClass, ok := params["command_class"].(string)
		if !ok {
			return nil, fmt.Errorf("command_class parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		deviceID := fmt.Sprintf("zwave-node-%d", int(nodeIDFloat))
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if zwaveDev, ok := device.(*iotadapters.ZWaveDevice); ok {
			return zwaveDev.GetValue(ctx, commandClass)
		}
		return nil, fmt.Errorf("device is not a Z-Wave device")

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

// ReceiveData receives data from the Z-Wave device.
func (d *ZWaveDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	// Z-Wave is event-driven, data comes via WebSocket
	// This method is not typically used for Z-Wave
	return nil, fmt.Errorf("Z-Wave uses event-driven communication, use SubscribeEvents instead")
}

// GetStatus retrieves the driver status.
func (d *ZWaveDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"type":       "zwave",
		"ws_url":     d.wsURL,
		"is_running": d.isRunning,
		"connected":  d.adapter != nil,
	}

	return status, nil
}

// SubscribeEvents subscribes to Z-Wave device events.
func (d *ZWaveDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	if d.adapter == nil {
		return fmt.Errorf("adapter not set")
	}

	// Subscribe to IoT events through the adapter
	// The adapter will need to support event subscription
	// For now, this is a placeholder for future implementation
	_ = handler // TODO: Implement actual event subscription through adapter

	return nil
}

// UnsubscribeEvents unsubscribes from Z-Wave device events.
func (d *ZWaveDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	// Implementation depends on adapter
	return nil
}

// SetAdapter sets the IoT protocol adapter.
func (d *ZWaveDriver) SetAdapter(adapter iot.ProtocolAdapter) {
	d.adapter = &adapter
}

// Start starts the Z-Wave driver.
func (d *ZWaveDriver) Start(ctx context.Context) error {
	d.isRunning = true
	return nil
}

// Stop stops the Z-Wave driver.
func (d *ZWaveDriver) Stop(ctx context.Context) error {
	d.isRunning = false
	return nil
}

// ZWaveConfig contains configuration for Z-Wave hardware.
type ZWaveConfig struct {
	WSURL          string            `json:"ws_url"`
	NetworkKey     string            `json:"network_key"`
	HomeID         uint32            `json:"home_id"`
	Metadata       map[string]string `json:"metadata"`
}
