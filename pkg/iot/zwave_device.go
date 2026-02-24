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
	"time"

)

// ZWaveDevice implements IoTDevice for Z-Wave devices.
type ZWaveDevice struct {
	*BaseDevice
	info    *DeviceInfo
	nodeID  uint8
	adapter *ZWaveAdapter
}

// NewZWaveDevice creates a new Z-Wave device instance.
func NewZWaveDevice(info *DeviceInfo, adapter *ZWaveAdapter) *ZWaveDevice {
	return &ZWaveDevice{
		BaseDevice: NewBaseDevice(info),
		info:      info,
		nodeID:    getUint8FromProps(info.Properties, "node_id"),
		adapter:   adapter,
	}
}

// ID returns the device ID.
func (d *ZWaveDevice) ID() string {
	return d.info.ID
}

// Connect connects to the Z-Wave device.
func (d *ZWaveDevice) Connect(ctx context.Context) error {
	// Z-Wave devices are always connected through the controller
	return nil
}

// Disconnect disconnects from the Z-Wave device.
func (d *ZWaveDevice) Disconnect(ctx context.Context) error {
	// Z-Wave devices are always connected
	return nil
}

// Read reads an attribute from the device.
func (d *ZWaveDevice) Read(ctx context.Context, attribute string) (interface{}, error) {
	if !d.IsConnected() {
		return nil, ErrNetworkError
	}

	// Map attribute to command class
	commandClass := mapAttributeToCommandClass(attribute)
	if commandClass == "" {
		return nil, fmt.Errorf("unsupported attribute: %s", attribute)
	}

	// Get value from Z-Wave JS
	value, err := d.adapter.jsClient.GetValue(ctx, d.nodeID, commandClass)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	return value, nil
}

// Write writes an attribute to the device.
func (d *ZWaveDevice) Write(ctx context.Context, attribute string, value interface{}) error {
	if !d.IsConnected() {
		return ErrNetworkError
	}

	// Map attribute to command class
	commandClass := mapAttributeToCommandClass(attribute)
	if commandClass == "" {
		return fmt.Errorf("unsupported attribute: %s", attribute)
	}

	// Set value through Z-Wave JS
	if err := d.adapter.jsClient.SetValue(ctx, d.nodeID, commandClass, value); err != nil {
		return fmt.Errorf("failed to write value: %w", err)
	}

	return nil
}

// Subscribe subscribes to device events.
func (d *ZWaveDevice) Subscribe(ctx context.Context, events []string, handler DeviceEventHandler) error {
	// Use BaseDevice's Subscribe method
	return d.BaseDevice.Subscribe(ctx, events, handler)
}

// Unsubscribe unsubscribes from device events.
func (d *ZWaveDevice) Unsubscribe(ctx context.Context, events []string) error {
	// Use BaseDevice's Unsubscribe method
	return d.BaseDevice.Unsubscribe(ctx, events)
}

// GetCapabilities returns device capabilities.
func (d *ZWaveDevice) GetCapabilities() []DeviceCapability {
	return d.Capabilities()
}

// GetProperty returns a device property.
func (d *ZWaveDevice) GetProperty(ctx context.Context, property string) (interface{}, error) {
	return d.Read(ctx, property)
}

// SetProperty sets a device property.
func (d *ZWaveDevice) SetProperty(ctx context.Context, property string, value interface{}) error {
	return d.Write(ctx, property, value)
}

// GetInfo returns device information.
func (d *ZWaveDevice) GetInfo() *DeviceInfo {
	return d.info
}

// RefreshState refreshes the device state from the actual device.
func (d *ZWaveDevice) RefreshState(ctx context.Context) error {
	// Get current state from Z-Wave JS
	nodeInfo, err := d.adapter.jsClient.GetNodeInfo(ctx, d.nodeID)
	if err != nil {
		return err
	}

	// Update device config
	config, _ := d.GetConfig(ctx)
	if config == nil {
		config = make(map[string]interface{})
	}

	// Update properties from node info
	if values, ok := nodeInfo["values"].([]map[string]interface{}); ok {
		for _, v := range values {
			if propertyName, ok := v["propertyName"].(string); ok {
				if value, ok := v["value"]; ok {
					config[propertyName] = value
				}
			}
		}
	}

	_ = d.SetConfig(ctx, config)
	return nil
}

// TurnOn turns on the device.
func (d *ZWaveDevice) TurnOn(ctx context.Context) error {
	return d.Write(ctx, "state", "on")
}

// TurnOff turns off the device.
func (d *ZWaveDevice) TurnOff(ctx context.Context) error {
	return d.Write(ctx, "state", "off")
}

// SetBrightness sets the brightness level (0-100).
func (d *ZWaveDevice) SetBrightness(ctx context.Context, level uint8) error {
	return d.Write(ctx, "brightness", level)
}

// SetValue sets a generic value on the device.
func (d *ZWaveDevice) SetValue(ctx context.Context, commandClass string, value interface{}) error {
	return d.adapter.jsClient.SetValue(ctx, d.nodeID, commandClass, value)
}

// GetValue gets a generic value from the device.
func (d *ZWaveDevice) GetValue(ctx context.Context, commandClass string) (interface{}, error) {
	return d.adapter.jsClient.GetValue(ctx, d.nodeID, commandClass)
}

// GetNodeInfo retrieves detailed node information.
func (d *ZWaveDevice) GetNodeInfo(ctx context.Context) (map[string]interface{}, error) {
	return d.adapter.jsClient.GetNodeInfo(ctx, d.nodeID)
}

// HealNode heals the node with the network.
func (d *ZWaveDevice) HealNode(ctx context.Context) error {
	// Node healing is handled by network healing
	return nil
}

// Ping sends a ping to the device.
func (d *ZWaveDevice) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	// Try to get node info
	_, err := d.adapter.jsClient.GetNodeInfo(ctx, d.nodeID)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// GetDiagnosticInfo returns diagnostic information.
func (d *ZWaveDevice) GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Basic info
	info["id"] = d.ID()
	info["node_id"] = d.nodeID
	info["status"] = d.Status()
	info["last_seen"] = d.LastSeen()

	// Node information
	nodeInfo, err := d.GetNodeInfo(ctx)
	if err == nil {
		info["node_info"] = nodeInfo

		// Extract signal strength if available
		if values, ok := nodeInfo["values"].([]map[string]interface{}); ok {
			for _, v := range values {
				if propertyName, ok := v["propertyName"].(string); ok {
					if propertyName == "rssi" {
						if value, ok := v["value"]; ok {
							info["rssi"] = value
						}
					} else if propertyName == "batteryLevel" {
						if value, ok := v["value"]; ok {
							info["battery_level"] = value
						}
					}
				}
			}
		}
	}

	// Connection test
	rtt, err := d.Ping(ctx)
	if err == nil {
		info["ping_ms"] = rtt.Milliseconds()
		info["connection"] = "ok"
	} else {
		info["connection"] = "failed"
		info["error"] = err.Error()
	}

	return info, nil
}

// BatchRead reads multiple attributes at once.
func (d *ZWaveDevice) BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, attr := range attributes {
		value, err := d.Read(ctx, attr)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", attr, err)
		}
		results[attr] = value
	}

	return results, nil
}

// BatchWrite writes multiple attributes at once.
func (d *ZWaveDevice) BatchWrite(ctx context.Context, values map[string]interface{}) error {
	for attr, value := range values {
		if err := d.Write(ctx, attr, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", attr, err)
		}
	}

	return nil
}

// Close closes the device and releases resources.
func (d *ZWaveDevice) Close(ctx context.Context) error {
	d.Disconnect(ctx)
	return nil
}

// mapAttributeToCommandClass maps attribute names to Z-Wave command classes.
func mapAttributeToCommandClass(attribute string) string {
	// Map common attributes to Z-Wave command classes
	switch attribute {
	case "state", "power":
		return "0x20" // Basic command class
	case "brightness":
		return "0x26" // Multilevel switch command class
	case "color":
		return "0x33" // Color command class
	case "temperature":
		return "0x31" // Sensor multilevel command class
	case "humidity":
		return "0x31" // Sensor multilevel command class
	case "battery_level":
		return "0x80" // Battery command class
	case "location":
		return "0x84" // Wake Up command class
	case "manufacturer":
		return "0x72" // Manufacturer specific command class
	case "version":
		return "0x86" // Version command class
	default:
		// Try to derive from attribute name
		if len(attribute) >= 2 {
			return attribute[:2]
		}
		return ""
	}
}

// executeCommand executes a command on the device.
func (d *ZWaveDevice) executeCommand(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	switch command {
	case "turn_on":
		return nil, d.TurnOn(ctx)
	case "turn_off":
		return nil, d.TurnOff(ctx)
	case "set_brightness":
		if level, ok := params["level"].(float64); ok {
			return nil, d.SetBrightness(ctx, uint8(level))
		}
		return nil, fmt.Errorf("invalid level parameter")
	case "get_value":
		if commandClass, ok := params["command_class"].(string); ok {
			return d.GetValue(ctx, commandClass)
		}
		return nil, fmt.Errorf("command_class parameter required")
	case "set_value":
		if commandClass, ok := params["command_class"].(string); ok {
			if value, ok := params["value"]; ok {
				return nil, d.SetValue(ctx, commandClass, value)
			}
		}
		return nil, fmt.Errorf("command_class and value parameters required")
	case "heal":
		return nil, d.HealNode(ctx)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// IsConnected returns whether the device is connected.
func (d *ZWaveDevice) IsConnected() bool {
	// Z-Wave devices are considered always connected through the controller
	return true
}

// GetConfig returns the device configuration.
func (d *ZWaveDevice) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	return d.BaseDevice.GetConfig(ctx)
}

// SetConfig sets the device configuration.
func (d *ZWaveDevice) SetConfig(ctx context.Context, config map[string]interface{}) error {
	return d.BaseDevice.SetConfig(ctx, config)
}
