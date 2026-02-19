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
	"fmt"
	"time"

	"AgentFramework/pkg/iot"
)

// NearLinkDevice implements IoTDevice for NearLink devices.
type NearLinkDevice struct {
	*iot.BaseDevice
	info    *iot.DeviceInfo
	macAddr string
	adapter *NearLinkAdapter
}

// NewNearLinkDevice creates a new NearLink device instance.
func NewNearLinkDevice(info *iot.DeviceInfo, adapter *NearLinkAdapter) *NearLinkDevice {
	return &NearLinkDevice{
		BaseDevice: iot.NewBaseDevice(info),
		info:       info,
		macAddr:    "",
		adapter:    adapter,
	}
}

// ID returns the device ID.
func (d *NearLinkDevice) ID() string {
	return d.info.ID
}

// Connect connects to the NearLink device.
func (d *NearLinkDevice) Connect(ctx context.Context) error {
	// NearLink devices are always connected through the mesh network
	return nil
}

// Disconnect disconnects from the NearLink device.
func (d *NearLinkDevice) Disconnect(ctx context.Context) error {
	// NearLink devices remain connected via mesh
	return nil
}

// Read reads an attribute from the device.
func (d *NearLinkDevice) Read(ctx context.Context, attribute string) (interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	// Use NearLink controller to read device attribute
	value, err := d.adapter.controller.ReadDeviceAttribute(ctx, d.macAddr, attribute)
	if err != nil {
		return nil, fmt.Errorf("failed to read attribute: %w", err)
	}

	return value, nil
}

// Write writes an attribute to the device.
func (d *NearLinkDevice) Write(ctx context.Context, attribute string, value interface{}) error {
	if !d.IsConnected() {
		return iot.ErrNetworkError
	}

	// Use NearLink controller to write device attribute
	if err := d.adapter.controller.WriteDeviceAttribute(ctx, d.macAddr, attribute, value); err != nil {
		return fmt.Errorf("failed to write attribute: %w", err)
	}

	return nil
}

// Subscribe subscribes to device events.
func (d *NearLinkDevice) Subscribe(ctx context.Context, events []string, handler iot.DeviceEventHandler) error {
	// Use BaseDevice's Subscribe method
	return d.BaseDevice.Subscribe(ctx, events, handler)
}

// Unsubscribe unsubscribes from device events.
func (d *NearLinkDevice) Unsubscribe(ctx context.Context, events []string) error {
	// Use BaseDevice's Unsubscribe method
	return d.BaseDevice.Unsubscribe(ctx, events)
}

// GetCapabilities returns device capabilities.
func (d *NearLinkDevice) GetCapabilities() []iot.DeviceCapability {
	return d.Capabilities()
}

// GetProperty returns a device property.
func (d *NearLinkDevice) GetProperty(ctx context.Context, property string) (interface{}, error) {
	return d.Read(ctx, property)
}

// SetProperty sets a device property.
func (d *NearLinkDevice) SetProperty(ctx context.Context, property string, value interface{}) error {
	return d.Write(ctx, property, value)
}

// GetInfo returns device information.
func (d *NearLinkDevice) GetInfo() *iot.DeviceInfo {
	return d.info
}

// RefreshState refreshes the device state from the actual device.
func (d *NearLinkDevice) RefreshState(ctx context.Context) error {
	// Read current state from NearLink device
	state, err := d.Read(ctx, "state")
	if err != nil {
		return err
	}

	// Update device config
	config, _ := d.GetConfig(ctx)
	if config == nil {
		config = make(map[string]interface{})
	}
	config["state"] = state
	_ = d.SetConfig(ctx, config)

	return nil
}

// IsConnected returns whether the device is connected.
func (d *NearLinkDevice) IsConnected() bool {
	// NearLink devices are considered always connected through the mesh
	return true
}

// GetConfig returns the device configuration.
func (d *NearLinkDevice) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	return d.BaseDevice.GetConfig(ctx)
}

// SetConfig sets the device configuration.
func (d *NearLinkDevice) SetConfig(ctx context.Context, config map[string]interface{}) error {
	return d.BaseDevice.SetConfig(ctx, config)
}

// SendNearLinkCommand sends a NearLink-specific command to the device.
func (d *NearLinkDevice) SendNearLinkCommand(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	// Use NearLink controller to send command
	result, err := d.adapter.controller.SendDeviceCommand(ctx, d.macAddr, command, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	return result, nil
}

// GetNearLinkInfo retrieves NearLink-specific device information.
func (d *NearLinkDevice) GetNearLinkInfo(ctx context.Context) (map[string]interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	// Use NearLink controller to get device info
	info, err := d.adapter.controller.GetDeviceInfo(ctx, d.macAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	return info, nil
}

// BatchRead reads multiple attributes at once.
func (d *NearLinkDevice) BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error) {
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
func (d *NearLinkDevice) BatchWrite(ctx context.Context, values map[string]interface{}) error {
	for attr, value := range values {
		if err := d.Write(ctx, attr, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", attr, err)
		}
	}

	return nil
}

// Ping sends a ping to the device and returns the round-trip time.
func (d *NearLinkDevice) Ping(ctx context.Context) (time.Duration, error) {
	if !d.IsConnected() {
		return 0, iot.ErrNetworkError
	}

	start := time.Now()

	// Send ping command
	_, err := d.adapter.controller.SendDeviceCommand(ctx, d.macAddr, "ping", nil)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// GetDiagnosticInfo returns diagnostic information for the device.
func (d *NearLinkDevice) GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	diag := make(map[string]interface{})

	// Basic info
	diag["id"] = d.ID()
	diag["mac_address"] = d.macAddr
	diag["status"] = d.Status()
	diag["last_seen"] = d.LastSeen()

	// NearLink specific info
	nearlinkInfo, err := d.GetNearLinkInfo(ctx)
	if err == nil {
		diag["nearlink_info"] = nearlinkInfo

		// Extract signal strength if available
		if rssi, ok := nearlinkInfo["rssi"]; ok {
			diag["rssi"] = rssi
		}
		if battery, ok := nearlinkInfo["battery_level"]; ok {
			diag["battery_level"] = battery
		}
	}

	// Connection test
	rtt, err := d.Ping(ctx)
	if err == nil {
		diag["ping_ms"] = rtt.Milliseconds()
		diag["connection"] = "ok"
	} else {
		diag["connection"] = "failed"
		diag["error"] = err.Error()
	}

	return diag, nil
}

// Close closes the device and releases resources.
func (d *NearLinkDevice) Close(ctx context.Context) error {
	d.Disconnect(ctx)
	return nil
}

// SetOn turns on the device.
func (d *NearLinkDevice) SetOn(ctx context.Context) error {
	return d.Write(ctx, "state", "on")
}

// SetOff turns off the device.
func (d *NearLinkDevice) SetOff(ctx context.Context) error {
	return d.Write(ctx, "state", "off")
}

// SetLevel sets the device level (0-100).
func (d *NearLinkDevice) SetLevel(ctx context.Context, level uint8) error {
	return d.Write(ctx, "level", level)
}

// Toggle toggles the device state.
func (d *NearLinkDevice) Toggle(ctx context.Context) error {
	// Read current state
	state, err := d.Read(ctx, "state")
	if err != nil {
		return err
	}

	// Toggle state
	var newState string
	if stateStr, ok := state.(string); ok {
		if stateStr == "on" || stateStr == "ON" || stateStr == "true" {
			newState = "off"
		} else {
			newState = "on"
		}
	} else {
		return fmt.Errorf("invalid state type")
	}

	// Write new state
	return d.Write(ctx, "state", newState)
}

// Stream opens a data stream from the device (for sensors, etc.).
func (d *NearLinkDevice) Stream(ctx context.Context, property string, interval time.Duration) (<-chan interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	dataChan := make(chan interface{})

	go func() {
		defer close(dataChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				value, err := d.Read(ctx, property)
				if err != nil {
					return
				}

				select {
				case dataChan <- value:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return dataChan, nil
}