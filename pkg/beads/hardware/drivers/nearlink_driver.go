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

// NearLinkDriver implements HardwareController for NearLink devices.
type NearLinkDriver struct {
	adapter   *iot.ProtocolAdapter
	isRunning bool
	networkMode string // SLM or SLE
}

// NewNearLinkDriver creates a new NearLink driver.
func NewNearLinkDriver() *NearLinkDriver {
	return &NearLinkDriver{}
}

// Connect connects to NearLink hardware (via NearLink controller).
func (d *NearLinkDriver) Connect(ctx context.Context, config interface{}) error {
	cfg, ok := config.(*NearLinkConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected NearLinkConfig")
	}

	d.networkMode = cfg.NetworkMode

	// The actual connection is handled by the NearLinkAdapter
	// This driver just stores the configuration
	return nil
}

// Disconnect disconnects from the NearLink hardware.
func (d *NearLinkDriver) Disconnect(ctx context.Context) error {
	d.networkMode = ""
	d.isRunning = false
	return nil
}

// SendCommand sends a command to the NearLink device.
func (d *NearLinkDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	switch cmd {
	case "discover_devices":
		// Discover devices on the NearLink network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.DiscoverDevices(ctx, 0)

	case "start_pairing":
		// Start device pairing
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		return adapter.StartPairing(ctx, 60*time.Second)

	case "cancel_pairing":
		// Cancel device pairing
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
		// Get NearLink network information
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.GetNetworkInfo(ctx)

	case "reset_network":
		// Reset NearLink network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return nil, adapter.ResetNetwork(ctx)

	case "read_attribute":
		// Read attribute from a device
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		attribute, ok := params["attribute"].(string)
		if !ok {
			return nil, fmt.Errorf("attribute parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return device.Read(ctx, attribute)

	case "write_attribute":
		// Write attribute to a device
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		attribute, ok := params["attribute"].(string)
		if !ok {
			return nil, fmt.Errorf("attribute parameter required")
		}
		value, ok := params["value"]
		if !ok {
			return nil, fmt.Errorf("value parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return nil, device.Write(ctx, attribute, value)

	case "send_command":
		// Send NearLink-specific command to a device
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		command, ok := params["command"].(string)
		if !ok {
			return nil, fmt.Errorf("command parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if nearlinkDev, ok := device.(*iotadapters.NearLinkDevice); ok {
			return nearlinkDev.SendNearLinkCommand(ctx, command, params)
		}
		return nil, fmt.Errorf("device is not a NearLink device")

	case "get_device_info":
		// Get NearLink-specific device information
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if nearlinkDev, ok := device.(*iotadapters.NearLinkDevice); ok {
			return nearlinkDev.GetNearLinkInfo(ctx)
		}
		return nil, fmt.Errorf("device is not a NearLink device")

	case "ping_device":
		// Ping a NearLink device
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		rtt, err := device.(*iotadapters.NearLinkDevice).Ping(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"rtt_ms": rtt.Milliseconds(),
		}, nil

	case "get_diagnostic_info":
		// Get diagnostic information for a device
		deviceID, ok := params["device_id"].(string)
		if !ok {
			return nil, fmt.Errorf("device_id parameter required")
		}
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		device, err := adapter.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return device.(*iotadapters.NearLinkDevice).GetDiagnosticInfo(ctx)

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

// ReceiveData receives data from the NearLink device.
func (d *NearLinkDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	// NearLink is event-driven, data comes via UDP multicast
	// This method is not typically used for NearLink
	return nil, fmt.Errorf("NearLink uses event-driven communication, use SubscribeEvents instead")
}

// GetStatus retrieves the driver status.
func (d *NearLinkDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"type":         "nearlink",
		"network_mode": d.networkMode,
		"is_running":   d.isRunning,
		"connected":    d.adapter != nil,
	}

	return status, nil
}

// SubscribeEvents subscribes to NearLink device events.
func (d *NearLinkDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	if d.adapter == nil {
		return fmt.Errorf("adapter not set")
	}

	// Subscribe to IoT events through the adapter
	_ = handler // TODO: Implement actual event subscription through adapter

	return nil
}

// UnsubscribeEvents unsubscribes from NearLink device events.
func (d *NearLinkDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	// Implementation depends on adapter
	return nil
}

// SetAdapter sets the IoT protocol adapter.
func (d *NearLinkDriver) SetAdapter(adapter iot.ProtocolAdapter) {
	d.adapter = &adapter
}

// Start starts the NearLink driver.
func (d *NearLinkDriver) Start(ctx context.Context) error {
	d.isRunning = true
	return nil
}

// Stop stops the NearLink driver.
func (d *NearLinkDriver) Stop(ctx context.Context) error {
	d.isRunning = false
	return nil
}

// NearLinkConfig contains configuration for NearLink hardware.
type NearLinkConfig struct {
	NetworkMode   string            `json:"network_mode"`   // SLM or SLE
	MulticastAddr string            `json:"multicast_addr"` // Multicast address
	Channel       uint8             `json:"channel"`        // 2.4GHz or 5.1GHz
	MeshID        uint64            `json:"mesh_id"`        // Mesh network ID
	Metadata      map[string]string `json:"metadata"`
}
