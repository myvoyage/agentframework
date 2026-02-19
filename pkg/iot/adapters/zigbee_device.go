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
	"sync"
	"time"

	"AgentFramework/pkg/iot"
)

// ZigbeeDevice represents a Zigbee IoT device.
type ZigbeeDevice struct {
	*iot.BaseDevice
	ID           string
	FriendlyName string
	Type         string
	Definition   string
	adapter      *ZigbeeAdapter
	state        map[string]interface{}
	stateMutex   sync.RWMutex
}

// NewZigbeeDevice creates a new Zigbee device.
func NewZigbeeDevice(id, friendlyName, deviceType string, adapter *ZigbeeAdapter) *ZigbeeDevice {
	info := &iot.DeviceInfo{
		ID:       id,
		Name:     friendlyName,
		Type:     determineDeviceType(deviceType),
		Protocol: iot.ProtocolZigbee,
		Status:   iot.DeviceStatusOnline,
		Capabilities: determineCapabilities(deviceType),
		Properties: make(map[string]interface{}),
		LastSeen: time.Now(),
	}

	return &ZigbeeDevice{
		BaseDevice:  iot.NewBaseDevice(info),
		ID:           id,
		FriendlyName: friendlyName,
		Type:         deviceType,
		adapter:      adapter,
		state:        make(map[string]interface{}),
	}
}

// Connect connects to the device (virtual connection for Zigbee).
func (d *ZigbeeDevice) Connect(ctx context.Context) error {
	d.BaseDevice.SetConnected(true)
	d.emitEvent(iot.EventDeviceStatusChanged, map[string]interface{}{
		"status": "connected",
	}, nil)
	return nil
}

// Disconnect disconnects from the device.
func (d *ZigbeeDevice) Disconnect(ctx context.Context) error {
	d.BaseDevice.SetConnected(false)
	d.emitEvent(iot.EventDeviceStatusChanged, map[string]interface{}{
		"status": "disconnected",
	}, nil)
	return nil
}

// Read reads a device attribute.
func (d *ZigbeeDevice) Read(ctx context.Context, attribute string) (interface{}, error) {
	d.stateMutex.RLock()
	defer d.stateMutex.RUnlock()

	// Check if attribute exists in state
	if value, exists := d.state[attribute]; exists {
		return value, nil
	}

	// Request state from device
	topic := d.adapter.mqttClient.GetDeviceGetTopic(d.ID)
	payload := map[string]string{attribute: ""}

	if err := d.adapter.mqttClient.Publish(ctx, topic, payload); err != nil {
		return nil, fmt.Errorf("failed to read attribute %s: %w", attribute, err)
	}

	// Wait for response (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Return cached value if available
		if value, exists := d.state[attribute]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("attribute %s not found", attribute)
	}
}

// Write writes a value to a device attribute.
func (d *ZigbeeDevice) Write(ctx context.Context, attribute string, value interface{}) error {
	// Prepare state update
	state := map[string]interface{}{
		attribute: value,
	}

	// For common Zigbee attributes, use the proper format
	switch attribute {
	case "state":
		// On/Off
		state = map[string]interface{}{"state": value}
	case "brightness":
		// Brightness (0-255)
		state = map[string]interface{}{"brightness": value}
	case "color":
		// Color (RGB or XY)
		state = map[string]interface{}{"color": value}
	case "color_temp":
		// Color temperature
		state = map[string]interface{}{"color_temp": value}
	}

	// Send to device
	if err := d.adapter.mqttClient.SetDeviceState(ctx, d.ID, state); err != nil {
		return fmt.Errorf("failed to write attribute %s: %w", attribute, err)
	}

	// Update local state
	d.stateMutex.Lock()
	d.state[attribute] = value
	d.stateMutex.Unlock()

	// Emit event
	d.emitEvent(iot.EventDataReceived, map[string]interface{}{
		"attribute": attribute,
		"value":     value,
	}, nil)

	return nil
}

// Subscribe subscribes to device events.
func (d *ZigbeeDevice) Subscribe(ctx context.Context, events []string, handler iot.DeviceEventHandler) error {
	return d.BaseDevice.Subscribe(ctx, events, handler)
}

// Unsubscribe unsubscribes from device events.
func (d *ZigbeeDevice) Unsubscribe(ctx context.Context, events []string) error {
	return d.BaseDevice.Unsubscribe(ctx, events)
}

// GetConfig retrieves device configuration.
func (d *ZigbeeDevice) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	config, err := d.BaseDevice.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Add Zigbee-specific configuration
	config["definition"] = d.Definition
	config["zigbee_type"] = d.Type

	return config, nil
}

// SetConfig sets device configuration.
func (d *ZigbeeDevice) SetConfig(ctx context.Context, config map[string]interface{}) error {
	// Handle friendly name change
	if name, ok := config["friendly_name"].(string); ok {
		topic := fmt.Sprintf("%s/bridge/config/rename", d.adapter.mqttClient.GetDeviceTopic(""))
		payload := map[string]string{
			"from": d.ID,
			"to":   name,
		}

		ctx := context.Background()
		if err := d.adapter.mqttClient.Publish(ctx, topic, payload); err != nil {
			return fmt.Errorf("failed to rename device: %w", err)
		}

		d.FriendlyName = name
		d.BaseDevice.info.Name = name
	}

	// Update base config
	return d.BaseDevice.SetConfig(ctx, config)
}

// updateState updates the device state.
func (d *ZigbeeDevice) updateState(state map[string]interface{}) {
	d.stateMutex.Lock()
	defer d.stateMutex.Unlock()

	for key, value := range state {
		d.state[key] = value
	}

	// Update last seen
	d.updateLastSeen()
}

// getState retrieves the current device state.
func (d *ZigbeeDevice) getState() map[string]interface{} {
	d.stateMutex.RLock()
	defer d.stateMutex.RUnlock()

	// Return copy of state
	state := make(map[string]interface{})
	for k, v := range d.state {
		state[k] = v
	}

	return state
}

// GetStateValue retrieves a specific state value.
func (d *ZigbeeDevice) GetStateValue(key string) (interface{}, bool) {
	d.stateMutex.RLock()
	defer d.stateMutex.RUnlock()

	value, exists := d.state[key]
	return value, exists
}

// emitEvent emits a device event.
func (d *ZigbeeDevice) emitEvent(eventType iot.EventType, data map[string]interface{}, err error) {
	event := iot.DeviceEvent{
		Type:      eventType,
		DeviceID:  d.ID,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}

	// Call base device event handlers
	d.BaseDevice.EmitEvent(context.Background(), event)
}

// EmitEvent is the public method to emit events.
func (d *ZigbeeDevice) EmitEvent(ctx context.Context, event iot.DeviceEvent) {
	// This will be called by the adapter when events occur
	d.BaseDevice.EmitEvent(ctx, event)
}

// determineDeviceType determines device type from Zigbee device type.
func determineDeviceType(zigbeeType string) iot.DeviceType {
	switch zigbeeType {
	case "EndDevice":
		return iot.DeviceTypeSensor
	case "Router":
		return iot.DeviceTypeActuator
	case "Coordinator":
		return iot.DeviceTypeGateway
	default:
		return iot.DeviceTypeUnknown
	}
}

// determineCapabilities determines device capabilities from device type.
func determineCapabilities(deviceType string) []iot.DeviceCapability {
	capabilities := []iot.DeviceCapability{}

	// Common capabilities for most Zigbee devices
	capabilities = append(capabilities, iot.CapabilityOnOff)

	// Add capabilities based on device type/definition
	switch deviceType {
	case "Router", "EndDevice":
		// Most devices support level control
		capabilities = append(capabilities, iot.CapabilityLevelControl)

		// Check for light-specific capabilities
		if isLightDevice(deviceType) {
			capabilities = append(capabilities,
				iot.CapabilityColorControl,
				iot.CapabilityColorTemp,
			)
		}

		// Check for sensor capabilities
		if isSensorDevice(deviceType) {
			capabilities = append(capabilities, iot.CapabilitySensor)
		}
	}

	return capabilities
}

// isLightDevice checks if device is a light.
func isLightDevice(deviceType string) bool {
	lightTypes := []string{
		"Router", "EndDevice",
		"Bulb", "Light", "CCT light",
		"RGB light", "Dimmable light",
	}

	for _, t := range lightTypes {
		if deviceType == t {
			return true
		}
	}

	return false
}

// isSensorDevice checks if device is a sensor.
func isSensorDevice(deviceType string) bool {
	sensorTypes := []string{
		"Sensor", "Temperature",
		"Humidity", "Motion",
		"Contact", "Switch",
	}

	for _, t := range sensorTypes {
		if deviceType == t {
			return true
		}
	}

	return false
}

// Common device operations

// TurnOn turns the device on.
func (d *ZigbeeDevice) TurnOn(ctx context.Context) error {
	return d.Write(ctx, "state", "ON")
}

// TurnOff turns the device off.
func (d *ZigbeeDevice) TurnOff(ctx context.Context) error {
	return d.Write(ctx, "state", "OFF")
}

// SetBrightness sets the brightness (0-255).
func (d *ZigbeeDevice) SetBrightness(ctx context.Context, brightness uint8) error {
	return d.Write(ctx, "brightness", brightness)
}

// SetColor sets the color (RGB).
func (d *ZigbeeDevice) SetColor(ctx context.Context, r, g, b uint8) error {
	color := map[string]interface{}{
		"r": r,
		"g": g,
		"b": b,
	}
	return d.Write(ctx, "color", color)
}

// SetColorTemp sets the color temperature.
func (d *ZigbeeDevice) SetColorTemp(ctx context.Context, temp uint16) error {
	return d.Write(ctx, "color_temp", temp)
}
