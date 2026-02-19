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

// ZigbeeDriver implements HardwareController for Zigbee devices.
type ZigbeeDriver struct {
	adapter *iot.ProtocolAdapter
	port     string
	baudRate int
}

// NewZigbeeDriver creates a new Zigbee driver.
func NewZigbeeDriver() *ZigbeeDriver {
	return &ZigbeeDriver{}
}

// Connect connects to Zigbee hardware (via Zigbee2MQTT).
func (d *ZigbeeDriver) Connect(ctx context.Context, config interface{}) error {
	cfg, ok := config.(*ZigbeeConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected ZigbeeConfig")
	}

	d.port = cfg.Port
	d.baudRate = cfg.BaudRate

	// The actual connection is handled by the ZigbeeAdapter via MQTT
	// This driver just stores the configuration
	return nil
}

// Disconnect disconnects from the Zigbee hardware.
func (d *ZigbeeDriver) Disconnect(ctx context.Context) error {
	d.port = ""
	d.baudRate = 0
	return nil
}

// SendCommand sends a command to the Zigbee device.
func (d *ZigbeeDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	switch cmd {
	case "discover_devices":
		// Discover devices on the Zigbee network
		if d.adapter == nil {
			return nil, fmt.Errorf("adapter not set")
		}
		adapter := (*d.adapter)
		return adapter.DiscoverDevices(ctx, 0)

	case "permit_join":
		// Enable/disable permit join mode
		_, ok := params["enable"].(bool)
		if !ok {
			return nil, fmt.Errorf("enable parameter required")
		}
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

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

// ReceiveData receives data from the Zigbee device.
func (d *ZigbeeDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	// Zigbee is event-driven, data comes via MQTT
	// This method is not typically used for Zigbee
	return nil, fmt.Errorf("Zigbee uses event-driven communication, use SubscribeEvents instead")
}

// GetStatus retrieves the driver status.
func (d *ZigbeeDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"type":     "zigbee",
		"port":     d.port,
		"baud_rate": d.baudRate,
		"connected": d.adapter != nil,
	}

	return status, nil
}

// SubscribeEvents subscribes to Zigbee device events.
func (d *ZigbeeDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	if d.adapter == nil {
		return fmt.Errorf("adapter not set")
	}

	// Subscribe to IoT events through the adapter
	// The adapter will need to support event subscription
	// For now, this is a placeholder for future implementation
	_ = handler // TODO: Implement actual event subscription through adapter

	return nil
}

// UnsubscribeEvents unsubscribes from Zigbee device events.
func (d *ZigbeeDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	// Implementation depends on adapter
	return nil
}

// SetAdapter sets the IoT protocol adapter.
func (d *ZigbeeDriver) SetAdapter(adapter iot.ProtocolAdapter) {
	d.adapter = &adapter
}

// ZigbeeConfig contains configuration for Zigbee hardware.
type ZigbeeConfig struct {
	Port        string            `json:"port"`
	BaudRate    int               `json:"baud_rate"`
	BrokerURL   string            `json:"broker_url"`
	TopicPrefix string            `json:"topic_prefix"`
	Metadata    map[string]string `json:"metadata"`
}
