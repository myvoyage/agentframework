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

// ProtocolAdapter defines interface for IoT protocol adapters.
type ProtocolAdapter interface {
	// Protocol identification
	Type() ProtocolType
	Version() string

	// Lifecycle management
	// Initialize initializes the adapter with configuration
	Initialize(ctx context.Context, config ProtocolConfig) error

	// Start starts the adapter
	Start(ctx context.Context) error

	// Stop stops the adapter
	Stop(ctx context.Context) error

	// IsRunning returns whether the adapter is running
	IsRunning() bool

	// Device management
	// DiscoverDevices discovers available devices on the network
	DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*DeviceInfo, error)

	// StartPairing starts device pairing/inclusion mode
	StartPairing(ctx context.Context, timeout time.Duration) (*PairingResult, error)

	// CancelPairing cancels ongoing pairing
	CancelPairing(ctx context.Context) error

	// RemoveDevice removes a device from the network
	RemoveDevice(ctx context.Context, deviceID string) error

	// Device interaction
	// GetDevice retrieves a device by ID
	GetDevice(ctx context.Context, deviceID string) (IoTDevice, error)

	// ListDevices lists all devices managed by this adapter
	ListDevices(ctx context.Context) ([]IoTDevice, error)

	// Network management
	// GetNetworkInfo retrieves network information
	GetNetworkInfo(ctx context.Context) (*NetworkInfo, error)

	// ResetNetwork resets the network (clears all devices)
	ResetNetwork(ctx context.Context) error

	// Event handling
	// SetEventHandler sets the global event handler for this adapter
	SetEventHandler(handler ProtocolEventHandler)
}

// ProtocolEventHandler handles protocol-level events.
type ProtocolEventHandler func(event ProtocolEvent)

// ProtocolEvent represents a protocol-level event.
type ProtocolEvent struct {
	Type      EventType              `json:"type"`
	Protocol  ProtocolType           `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Error     error                  `json:"error,omitempty"`
}

// BaseAdapter provides base implementation for ProtocolAdapter.
type BaseAdapter struct {
	protocolType   ProtocolType
	version        string
	config         ProtocolConfig
	running        bool
	eventHandler   ProtocolEventHandler
	devices        map[string]IoTDevice
}

// NewBaseAdapter creates a new base adapter.
func NewBaseAdapter(protocolType ProtocolType, version string) *BaseAdapter {
	return &BaseAdapter{
		protocolType: protocolType,
		version:      version,
		running:      false,
		devices:      make(map[string]IoTDevice),
	}
}

// Type returns the protocol type.
func (a *BaseAdapter) Type() ProtocolType {
	return a.protocolType
}

// Version returns the protocol version.
func (a *BaseAdapter) Version() string {
	return a.version
}

// Initialize initializes the adapter.
func (a *BaseAdapter) Initialize(ctx context.Context, config ProtocolConfig) error {
	a.config = config
	return nil
}

// Start starts the adapter.
func (a *BaseAdapter) Start(ctx context.Context) error {
	a.running = true
	return nil
}

// Stop stops the adapter.
func (a *BaseAdapter) Stop(ctx context.Context) error {
	a.running = false
	return nil
}

// IsRunning returns whether the adapter is running.
func (a *BaseAdapter) IsRunning() bool {
	return a.running
}

// GetDevice retrieves a device by ID.
func (a *BaseAdapter) GetDevice(ctx context.Context, deviceID string) (IoTDevice, error) {
	device, exists := a.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}
	return device, nil
}

// ListDevices lists all devices.
func (a *BaseAdapter) ListDevices(ctx context.Context) ([]IoTDevice, error) {
	devices := make([]IoTDevice, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// SetEventHandler sets the event handler.
func (a *BaseAdapter) SetEventHandler(handler ProtocolEventHandler) {
	a.eventHandler = handler
}

// emitEvent emits a protocol event.
func (a *BaseAdapter) emitEvent(eventType EventType, deviceID string, data map[string]interface{}, err error) {
	if a.eventHandler == nil {
		return
	}

	event := ProtocolEvent{
		Type:      eventType,
		Protocol:  a.protocolType,
		Timestamp: time.Now(),
		DeviceID:  deviceID,
		Data:      data,
		Error:     err,
	}

	a.eventHandler(event)
}

// addDevice adds a device to the adapter.
func (a *BaseAdapter) addDevice(device IoTDevice) error {
	deviceID := device.ID()
	if _, exists := a.devices[deviceID]; exists {
		return ErrDeviceAlreadyExists
	}
	a.devices[deviceID] = device
	return nil
}

// removeDevice removes a device from the adapter.
func (a *BaseAdapter) removeDevice(deviceID string) error {
	if _, exists := a.devices[deviceID]; !exists {
		return ErrDeviceNotFound
	}
	delete(a.devices, deviceID)
	return nil
}

// Errors
var (
	ErrDeviceNotFound       = &ProtocolError{Code: "DEVICE_NOT_FOUND", Message: "device not found"}
	ErrDeviceAlreadyExists  = &ProtocolError{Code: "DEVICE_EXISTS", Message: "device already exists"}
	ErrAdapterNotRunning    = &ProtocolError{Code: "ADAPTER_NOT_RUNNING", Message: "adapter is not running"}
	ErrPairingTimeout       = &ProtocolError{Code: "PAIRING_TIMEOUT", Message: "device pairing timeout"}
	ErrInvalidConfiguration = &ProtocolError{Code: "INVALID_CONFIG", Message: "invalid configuration"}
	ErrNetworkError         = &ProtocolError{Code: "NETWORK_ERROR", Message: "network communication error"}
)

// ProtocolError represents a protocol error.
type ProtocolError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface.
func (e *ProtocolError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Unwrap implements the error unwrapping interface.
func (e *ProtocolError) Unwrap() error {
	return e.Err
}
