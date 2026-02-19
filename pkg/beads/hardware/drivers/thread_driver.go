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
)

// ThreadDriver implements HardwareController for Thread devices.
type ThreadDriver struct {
	adapter      *iot.ProtocolAdapter
	interfaceName string
	isRunning    bool
}

// NewThreadDriver creates a new Thread driver.
func NewThreadDriver() *ThreadDriver {
	return &ThreadDriver{}
}

// Connect connects to Thread hardware (via Thread border router).
func (d *ThreadDriver) Connect(ctx context.Context, config interface{}) error {
	cfg, ok := config.(*ThreadConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected ThreadConfig")
	}

	d.interfaceName = cfg.Interface

	// The actual connection is handled by the ThreadAdapter
	// This driver just stores the configuration
	return nil
}

// Disconnect disconnects from the Thread hardware.
func (d *ThreadDriver) Disconnect(ctx context.Context) error {
	d.interfaceName = ""
	d.isRunning = false
	return nil
}

// SendCommand sends a command to the Thread device.
func (d *ThreadDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	switch cmd {
	case "discover_devices":
		// Discover devices on the Thread network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.DiscoverDevices(ctx, 0)

	case "permit_join":
		// Enable/disable permit join mode
		enable, ok := params["enable"].(bool)
		if !ok {
			return nil, fmt.Errorf("enable parameter required")
		}
		_ = enable // Will be used in actual implementation
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		return adapter.StartPairing(ctx, 60*time.Second)

	case "get_devices":
		// Get all devices
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.ListDevices(ctx)

	case "get_network_info":
		// Get Thread network information
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.GetNetworkInfo(ctx)

	case "reset_network":
		// Reset Thread network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return nil, adapter.ResetNetwork(ctx)

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

// ReceiveData receives data from the Thread device.
func (d *ThreadDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	// Thread is event-driven, data comes via CoAP
	// This method is not typically used for Thread
	return nil, fmt.Errorf("Thread uses event-driven communication, use SubscribeEvents instead")
}

// GetStatus retrieves the driver status.
func (d *ThreadDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"type":       "thread",
		"interface":  d.interfaceName,
		"is_running": d.isRunning,
		"connected":  d.adapter != nil,
	}

	return status, nil
}

// SubscribeEvents subscribes to Thread device events.
func (d *ThreadDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	if d.adapter == nil {
		return fmt.Errorf("adapter not set")
	}

	// Subscribe to IoT events through the adapter
	// The adapter will need to support event subscription
	// For now, this is a placeholder for future implementation
	_ = handler // TODO: Implement actual event subscription through adapter

	return nil
}

// UnsubscribeEvents unsubscribes from Thread device events.
func (d *ThreadDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	// Implementation depends on adapter
	return nil
}

// SetAdapter sets the IoT protocol adapter.
func (d *ThreadDriver) SetAdapter(adapter iot.ProtocolAdapter) {
	d.adapter = &adapter
}

// Start starts the Thread driver.
func (d *ThreadDriver) Start(ctx context.Context) error {
	d.isRunning = true
	return nil
}

// Stop stops the Thread driver.
func (d *ThreadDriver) Stop(ctx context.Context) error {
	d.isRunning = false
	return nil
}

// ThreadConfig contains configuration for Thread hardware.
type ThreadConfig struct {
	Interface         string            `json:"interface"`
	NetworkName       string            `json:"network_name"`
	PanID             uint16            `json:"pan_id"`
	Channel           uint8             `json:"channel"`
	MeshLocalPrefix   string            `json:"mesh_local_prefix"`
	OnMeshPrefix      string            `json:"on_mesh_prefix"`
	BorderRouterAddr  string            `json:"border_router_addr"`
	Metadata          map[string]string `json:"metadata"`
}
