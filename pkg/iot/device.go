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
	"time"
)

// IoTDevice defines unified interface for IoT devices across all protocols.
type IoTDevice interface {
	// Identity methods
	ID() string                  // Unique device identifier
	Name() string                // Device name
	Type() DeviceType           // Device type (sensor, actuator, etc.)
	Protocol() ProtocolType     // Protocol type (zigbee, zwave, thread)

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Status() DeviceStatus

	// Device information
	Manufacturer() string
	Model() string
	Version() string
	Capabilities() []DeviceCapability

	// Data interaction
	// Read reads a device attribute (e.g., "temperature", "humidity", "state")
	Read(ctx context.Context, attribute string) (interface{}, error)

	// Write writes a value to a device attribute
	Write(ctx context.Context, attribute string, value interface{}) error

	// Subscribe subscribes to device events
	// events can include: "state_changed", "attribute_changed", "error", etc.
	Subscribe(ctx context.Context, events []string, handler DeviceEventHandler) error

	// Unsubscribe unsubscribes from device events
	Unsubscribe(ctx context.Context, events []string) error

	// Configuration management
	GetConfig(ctx context.Context) (map[string]interface{}, error)
	SetConfig(ctx context.Context, config map[string]interface{}) error

	// Metadata
	LastSeen() time.Time      // Last communication timestamp
	SignalStrength() int      // RSSI/LQI signal strength (0-100)
	BatteryLevel() uint8      // Battery level (0-100, 255=external power)
}

// DeviceEventHandler handles device events.
type DeviceEventHandler func(ctx context.Context, event DeviceEvent)

// DeviceEvent represents a device event.
type DeviceEvent struct {
	Type      EventType              `json:"type"`
	DeviceID  string                 `json:"device_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Error     error                  `json:"error,omitempty"`
}

// BaseDevice provides base implementation for IoTDevice.
type BaseDevice struct {
	info           *DeviceInfo
	connected      bool
	lastSeen       time.Time
	signalStrength int
	batteryLevel   uint8
	eventHandlers  map[string]DeviceEventHandler
}

// NewBaseDevice creates a new base device.
func NewBaseDevice(info *DeviceInfo) *BaseDevice {
	return &BaseDevice{
		info:          info,
		connected:     false,
		lastSeen:      time.Now(),
		signalStrength: 0,
		batteryLevel:  255,
		eventHandlers: make(map[string]DeviceEventHandler),
	}
}

// ID returns the device ID.
func (d *BaseDevice) ID() string {
	return d.info.ID
}

// Name returns the device name.
func (d *BaseDevice) Name() string {
	return d.info.Name
}

// Type returns the device type.
func (d *BaseDevice) Type() DeviceType {
	return d.info.Type
}

// Protocol returns the protocol type.
func (d *BaseDevice) Protocol() ProtocolType {
	return d.info.Protocol
}

// IsConnected returns connection status.
func (d *BaseDevice) IsConnected() bool {
	return d.connected
}

// Status returns the device status.
func (d *BaseDevice) Status() DeviceStatus {
	return d.info.Status
}

// Manufacturer returns the manufacturer.
func (d *BaseDevice) Manufacturer() string {
	return d.info.Manufacturer
}

// Model returns the model.
func (d *BaseDevice) Model() string {
	return d.info.Model
}

// Version returns the version.
func (d *BaseDevice) Version() string {
	return d.info.Version
}

// Capabilities returns device capabilities.
func (d *BaseDevice) Capabilities() []DeviceCapability {
	return d.info.Capabilities
}

// LastSeen returns last seen timestamp.
func (d *BaseDevice) LastSeen() time.Time {
	return d.lastSeen
}

// SignalStrength returns signal strength.
func (d *BaseDevice) SignalStrength() int {
	return d.signalStrength
}

// BatteryLevel returns battery level.
func (d *BaseDevice) BatteryLevel() uint8 {
	return d.batteryLevel
}

// GetConfig returns device configuration.
func (d *BaseDevice) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	return d.info.Properties, nil
}

// SetConfig sets device configuration.
func (d *BaseDevice) SetConfig(ctx context.Context, config map[string]interface{}) error {
	for k, v := range config {
		d.info.Properties[k] = v
	}
	return nil
}

// Subscribe subscribes to device events.
func (d *BaseDevice) Subscribe(ctx context.Context, events []string, handler DeviceEventHandler) error {
	for _, event := range events {
		d.eventHandlers[event] = handler
	}
	return nil
}

// Unsubscribe unsubscribes from device events.
func (d *BaseDevice) Unsubscribe(ctx context.Context, events []string) error {
	for _, event := range events {
		delete(d.eventHandlers, event)
	}
	return nil
}

// emitEvent emits a device event to all registered handlers.
func (d *BaseDevice) emitEvent(ctx context.Context, eventType EventType, data map[string]interface{}, err error) {
	event := DeviceEvent{
		Type:      eventType,
		DeviceID:  d.info.ID,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}

	// Call all registered handlers
	for _, handler := range d.eventHandlers {
		handler(ctx, event)
	}
}

// updateLastSeen updates the last seen timestamp.
func (d *BaseDevice) updateLastSeen() {
	d.lastSeen = time.Now()
}

// setSignalStrength sets the signal strength.
func (d *BaseDevice) setSignalStrength(strength int) {
	d.signalStrength = strength
}

// setBatteryLevel sets the battery level.
func (d *BaseDevice) setBatteryLevel(level uint8) {
	d.batteryLevel = level
}

// setConnected sets the connection status.
func (d *BaseDevice) setConnected(connected bool) {
	d.connected = connected
	if connected {
		d.info.Status = DeviceStatusOnline
	} else {
		d.info.Status = DeviceStatusOffline
	}
}

// ===== Optional Device Interfaces =====

// BatchReader provides batch read capability.
type BatchReader interface {
	BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error)
}

// BatchWriter provides batch write capability.
type BatchWriter interface {
	BatchWrite(ctx context.Context, values map[string]interface{}) error
}

// Toggleable provides toggle capability.
type Toggleable interface {
	Toggle(ctx context.Context) error
}

// Pingable provides ping capability.
type Pingable interface {
	Ping(ctx context.Context) (time.Duration, error)
}

// Diagnosticable provides diagnostic capability.
type Diagnosticable interface {
	GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error)
}

// GetProtocolFromDeviceID extracts protocol from device ID.
func GetProtocolFromDeviceID(deviceID string) ProtocolType {
	// Parse device ID to extract protocol
	// Format: "<protocol>-<rest>"
	// e.g., "zigbee-node-001", "zwave-node-002", "thread-00:11:22:33:44:55"
	if len(deviceID) < 8 {
		return ProtocolUnknown
	}

	prefix := deviceID[:7]
	switch prefix {
	case "zigbee-":
		return ProtocolZigbee
	case "zwave-":
		return ProtocolZWave
	case "thread-":
		return ProtocolThread
	case "nearlin":
		return ProtocolNearLink
	default:
		return ProtocolUnknown
	}
}
