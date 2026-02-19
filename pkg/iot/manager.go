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
	"sync"
	"time"
)

// IoTDeviceManager manages IoT devices across multiple protocols.
type IoTDeviceManager struct {
	adapters map[ProtocolType]ProtocolAdapter
	devices  map[string]IoTDevice
	eventBus *EventBus
	registry *DeviceRegistry
	mutex    sync.RWMutex
}

// NewIoTDeviceManager creates a new IoT device manager.
func NewIoTDeviceManager() *IoTDeviceManager {
	return &IoTDeviceManager{
		adapters: make(map[ProtocolType]ProtocolAdapter),
		devices:  make(map[string]IoTDevice),
		eventBus: NewEventBus(),
		registry: NewDeviceRegistry(),
	}
}

// RegisterAdapter registers a protocol adapter.
func (m *IoTDeviceManager) RegisterAdapter(adapter ProtocolAdapter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	protocolType := adapter.Type()
	if _, exists := m.adapters[protocolType]; exists {
		return fmt.Errorf("adapter for %s already registered", protocolType)
	}

	m.adapters[protocolType] = adapter

	// Set adapter event handler
	adapter.SetEventHandler(func(event ProtocolEvent) {
		m.handleProtocolEvent(event)
	})

	return nil
}

// UnregisterAdapter unregisters a protocol adapter.
func (m *IoTDeviceManager) UnregisterAdapter(protocolType ProtocolType) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	adapter, exists := m.adapters[protocolType]
	if !exists {
		return ErrAdapterNotFound
	}

	// Stop adapter if running
	if adapter.IsRunning() {
		adapter.Stop(context.Background())
	}

	delete(m.adapters, protocolType)
	return nil
}

// ListAdapters lists all registered adapters.
func (m *IoTDeviceManager) ListAdapters() []ProtocolType {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	adapters := make([]ProtocolType, 0, len(m.adapters))
	for protocolType := range m.adapters {
		adapters = append(adapters, protocolType)
	}
	return adapters
}

// GetAdapter retrieves an adapter by protocol type.
func (m *IoTDeviceManager) GetAdapter(protocolType ProtocolType) (ProtocolAdapter, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	adapter, exists := m.adapters[protocolType]
	if !exists {
		return nil, ErrAdapterNotFound
	}
	return adapter, nil
}

// StartAdapter starts an adapter.
func (m *IoTDeviceManager) StartAdapter(ctx context.Context, protocolType ProtocolType) error {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocolType]
	m.mutex.RUnlock()

	if !exists {
		return ErrAdapterNotFound
	}

	return adapter.Start(ctx)
}

// StopAdapter stops an adapter.
func (m *IoTDeviceManager) StopAdapter(ctx context.Context, protocolType ProtocolType) error {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocolType]
	m.mutex.RUnlock()

	if !exists {
		return ErrAdapterNotFound
	}

	return adapter.Stop(ctx)
}

// DiscoverDevices discovers devices across all protocols or a specific protocol.
func (m *IoTDeviceManager) DiscoverDevices(
	ctx context.Context,
	protocol ProtocolType,
	timeout time.Duration,
) ([]*DeviceInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// If protocol specified, discover only from that protocol
	if protocol != "" {
		adapter, exists := m.adapters[protocol]
		if !exists {
			return nil, ErrAdapterNotFound
		}
		return adapter.DiscoverDevices(ctx, timeout)
	}

	// Discover from all adapters
	var allDevices []*DeviceInfo
	var wg sync.WaitGroup
	results := make(chan []*DeviceInfo, len(m.adapters))
	errors := make(chan error, len(m.adapters))

	for protocolType, adapter := range m.adapters {
		wg.Add(1)
		go func(pt ProtocolType, a ProtocolAdapter) {
			defer wg.Done()

			devices, err := a.DiscoverDevices(ctx, timeout)
			if err != nil {
				m.eventBus.Publish(Event{
					Type:      EventError,
					Source:    string(pt),
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"error": err.Error(),
					},
				})
				errors <- err
				return
			}
			results <- devices
		}(protocolType, adapter)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	for devices := range results {
		allDevices = append(allDevices, devices...)
	}

	// Check for errors (optional, depending on requirements)
	select {
	case err := <-errors:
		// At least one adapter failed
		// You might want to handle this differently
		_ = err
	default:
	}

	return allDevices, nil
}

// StartPairing starts device pairing for a specific protocol.
func (m *IoTDeviceManager) StartPairing(
	ctx context.Context,
	protocol ProtocolType,
	timeout time.Duration,
) (*PairingResult, error) {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocol]
	m.mutex.RUnlock()

	if !exists {
		return nil, ErrAdapterNotFound
	}

	if !adapter.IsRunning() {
		return nil, ErrAdapterNotRunning
	}

	return adapter.StartPairing(ctx, timeout)
}

// CancelPairing cancels ongoing device pairing.
func (m *IoTDeviceManager) CancelPairing(ctx context.Context, protocol ProtocolType) error {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocol]
	m.mutex.RLock()

	if !exists {
		return ErrAdapterNotFound
	}

	return adapter.CancelPairing(ctx)
}

// GetDevice retrieves a device by ID.
func (m *IoTDeviceManager) GetDevice(deviceID string) (IoTDevice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	device, exists := m.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}
	return device, nil
}

// ListDevices lists all devices across all protocols.
func (m *IoTDeviceManager) ListDevices(ctx context.Context) ([]IoTDevice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	devices := make([]IoTDevice, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// ListDevicesByProtocol lists devices for a specific protocol.
func (m *IoTDeviceManager) ListDevicesByProtocol(
	ctx context.Context,
	protocol ProtocolType,
) ([]IoTDevice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	devices := make([]IoTDevice, 0)
	for _, device := range m.devices {
		if device.Protocol() == protocol {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

// RemoveDevice removes a device.
func (m *IoTDeviceManager) RemoveDevice(ctx context.Context, deviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	device, exists := m.devices[deviceID]
	if !exists {
		return ErrDeviceNotFound
	}

	protocol := device.Protocol()
	adapter, exists := m.adapters[protocol]
	if !exists {
		return ErrAdapterNotFound
	}

	// Remove from adapter
	if err := adapter.RemoveDevice(ctx, deviceID); err != nil {
		return err
	}

	// Remove from manager
	delete(m.devices, deviceID)

	// Remove from registry
	m.registry.Unregister(deviceID)

	// Publish event
	m.eventBus.Publish(Event{
		Type:      EventDeviceLeft,
		Source:    string(protocol),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"device_id": deviceID,
		},
	})

	return nil
}

// GetNetworkInfo retrieves network information for a protocol.
func (m *IoTDeviceManager) GetNetworkInfo(
	ctx context.Context,
	protocol ProtocolType,
) (*NetworkInfo, error) {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocol]
	m.mutex.RUnlock()

	if !exists {
		return nil, ErrAdapterNotFound
	}

	return adapter.GetNetworkInfo(ctx)
}

// ResetNetwork resets a protocol network (clears all devices).
func (m *IoTDeviceManager) ResetNetwork(
	ctx context.Context,
	protocol ProtocolType,
) error {
	m.mutex.RLock()
	adapter, exists := m.adapters[protocol]
	m.mutex.RUnlock()

	if !exists {
		return ErrAdapterNotFound
	}

	return adapter.ResetNetwork(ctx)
}

// SubscribeToEvents subscribes to IoT events.
func (m *IoTDeviceManager) SubscribeToEvents(
	eventType string,
	handler EventHandler,
) func() {
	return m.eventBus.Subscribe(eventType, handler)
}

// handleProtocolEvent handles protocol adapter events.
func (m *IoTDeviceManager) handleProtocolEvent(event ProtocolEvent) {
	// Publish to internal event bus
	m.eventBus.Publish(Event{
		Type:      event.Type,
		Source:    string(event.Protocol),
		Timestamp: event.Timestamp,
		Data:      event.Data,
	})

	// Update device status if needed
	if event.DeviceID != "" {
		m.mutex.RLock()
		_, exists := m.devices[event.DeviceID]
		m.mutex.RUnlock()

		if exists {
			switch event.Type {
			case EventDeviceStatusChanged:
				// Update device status in registry
				if status, ok := event.Data["status"].(DeviceStatus); ok {
					m.registry.UpdateStatus(event.DeviceID, status)
				}
			case EventDataReceived:
				// Update last seen
				m.registry.UpdateLastSeen(event.DeviceID, time.Now())
			}
		}
	}
}

// addDevice adds a device to the manager.
func (m *IoTDeviceManager) addDevice(device IoTDevice) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deviceID := device.ID()
	if _, exists := m.devices[deviceID]; exists {
		return ErrDeviceAlreadyExists
	}

	m.devices[deviceID] = device

	// Add to registry
	info := &DeviceInfo{
		ID:           device.ID(),
		Name:         device.Name(),
		Type:         device.Type(),
		Protocol:     device.Protocol(),
		Manufacturer: device.Manufacturer(),
		Model:        device.Model(),
		Version:      device.Version(),
		Status:       device.Status(),
		Capabilities: device.Capabilities(),
		LastSeen:     device.LastSeen(),
	}

	m.registry.Register(info)

	return nil
}

// Close closes the device manager and all adapters.
func (m *IoTDeviceManager) Close(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errs []error

	// Stop all adapters
	for protocolType, adapter := range m.adapters {
		if adapter.IsRunning() {
			if err := adapter.Stop(ctx); err != nil {
				errs = append(errs, fmt.Errorf("failed to stop %s adapter: %w", protocolType, err))
			}
		}
	}

	// Close event bus
	if err := m.eventBus.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close event bus: %w", err))
	}

	// Close registry
	if err := m.registry.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close registry: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during cleanup", len(errs))
	}

	return nil
}

// Errors
var (
	ErrAdapterNotFound = &ManagerError{Code: "ADAPTER_NOT_FOUND", Message: "protocol adapter not found"}
)

// ManagerError represents a device manager error.
type ManagerError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface.
func (e *ManagerError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Unwrap implements the error unwrapping interface.
func (e *ManagerError) Unwrap() error {
	return e.Err
}
