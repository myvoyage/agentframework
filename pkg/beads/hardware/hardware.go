// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package hardware provides unified hardware interface abstraction and driver management.
package hardware

import (
	"context"
	"errors"
	"sync"
	"time"
)

// HardwareController defines the interface for hardware device control.
type HardwareController interface {
	// Connect establishes a connection to the hardware device.
	Connect(ctx context.Context, config interface{}) error

	// Disconnect closes the connection to the hardware device.
	Disconnect(ctx context.Context) error

	// SendCommand sends a command to the hardware device.
	SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error)

	// ReceiveData receives data from the hardware device with timeout.
	ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error)

	// GetStatus retrieves the current status of the hardware device.
	GetStatus(ctx context.Context) (map[string]interface{}, error)

	// SubscribeEvents subscribes to hardware device events.
	SubscribeEvents(ctx context.Context, eventTypes []string, handler EventHandler) error

	// UnsubscribeEvents unsubscribes from hardware device events.
	UnsubscribeEvents(ctx context.Context, eventTypes []string) error
}

// EventHandler handles hardware device events.
type EventHandler func(ctx context.Context, event HardwareEvent)

// HardwareEvent represents a hardware device event.
type HardwareEvent struct {
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// HardwareDriverManager manages hardware drivers.
type HardwareDriverManager struct {
	drivers map[string]HardwareController
	mutex   sync.RWMutex
}

// NewHardwareDriverManager creates a new HardwareDriverManager instance.
func NewHardwareDriverManager() *HardwareDriverManager {
	return &HardwareDriverManager{
		drivers: make(map[string]HardwareController),
	}
}

// RegisterDriver registers a hardware driver.
func (m *HardwareDriverManager) RegisterDriver(name string, driver HardwareController) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.drivers[name]; exists {
		return errors.New("driver already registered")
	}

	m.drivers[name] = driver
	return nil
}

// UnregisterDriver unregisters a hardware driver.
func (m *HardwareDriverManager) UnregisterDriver(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.drivers[name]; !exists {
		return errors.New("driver not found")
	}

	delete(m.drivers, name)
	return nil
}

// GetDriver retrieves a hardware driver by name.
func (m *HardwareDriverManager) GetDriver(name string) (HardwareController, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	driver, exists := m.drivers[name]
	if !exists {
		return nil, errors.New("driver not found")
	}

	return driver, nil
}

// ListDrivers lists all registered hardware drivers.
func (m *HardwareDriverManager) ListDrivers() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var drivers []string
	for name := range m.drivers {
		drivers = append(drivers, name)
	}

	return drivers
}

// DeviceInfo contains information about a connected hardware device.
type DeviceInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}

// DeviceManager manages connected hardware devices.
type DeviceManager struct {
	devices map[string]*DeviceInfo
	mutex   sync.RWMutex
}

// NewDeviceManager creates a new DeviceManager instance.
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]*DeviceInfo),
	}
}

// AddDevice adds a device to the device manager.
func (m *DeviceManager) AddDevice(deviceInfo *DeviceInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.devices[deviceInfo.ID]; exists {
		return errors.New("device already exists")
	}

	m.devices[deviceInfo.ID] = deviceInfo
	return nil
}

// RemoveDevice removes a device from the device manager.
func (m *DeviceManager) RemoveDevice(deviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.devices[deviceID]; !exists {
		return errors.New("device not found")
	}

	delete(m.devices, deviceID)
	return nil
}

// GetDevice retrieves a device by ID.
func (m *DeviceManager) GetDevice(deviceID string) (*DeviceInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	device, exists := m.devices[deviceID]
	if !exists {
		return nil, errors.New("device not found")
	}

	return device, nil
}

// ListDevices lists all connected devices.
func (m *DeviceManager) ListDevices() []*DeviceInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var devices []*DeviceInfo
	for _, device := range m.devices {
		devices = append(devices, device)
	}

	return devices
}

// UpdateDeviceStatus updates the status of a device.
func (m *DeviceManager) UpdateDeviceStatus(deviceID string, status string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	device, exists := m.devices[deviceID]
	if !exists {
		return errors.New("device not found")
	}

	device.Status = status
	return nil
}

// DeviceDiscovery discovers available hardware devices.
type DeviceDiscovery interface {
	Discover(ctx context.Context) ([]*DeviceInfo, error)
	Start(ctx context.Context) error
	Stop() error
}

// SerialDeviceConfig contains configuration for a serial device.
type SerialDeviceConfig struct {
	Port     string `json:"port"`
	BaudRate int    `json:"baud_rate"`
	DataBits int    `json:"data_bits"`
	Parity   string `json:"parity"`
	StopBits int    `json:"stop_bits"`
	Timeout  int    `json:"timeout"`
}

// ModbusDeviceConfig contains configuration for a Modbus device.
type ModbusDeviceConfig struct {
	Address   string `json:"address"`
	Port      int    `json:"port"`
	SlaveID   int    `json:"slave_id"`
	Timeout   int    `json:"timeout"`
	Retries   int    `json:"retries"`
}

// CANDeviceConfig contains configuration for a CAN bus device.
type CANDeviceConfig struct {
	Interface string `json:"interface"`
	BaudRate  int    `json:"baud_rate"`
	Timeout   int    `json:"timeout"`
}