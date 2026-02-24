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

// ZWaveAdapter implements ProtocolAdapter for Z-Wave networks.
type ZWaveAdapter struct {
	*iot.BaseAdapter
	jsClient   *ZWaveJSClient
	devices    map[uint8]*ZWaveDevice
	storage    iot.DeviceRegistry
	mutex      sync.RWMutex
	eventHandler iot.ProtocolEventHandler
}

// NewZWaveAdapter creates a new Z-Wave protocol adapter.
func NewZWaveAdapter() *ZWaveAdapter {
	return &ZWaveAdapter{
		BaseAdapter: iot.NewBaseAdapter(iot.ProtocolZWave, "1.0.0"),
		devices:     make(map[uint8]*ZWaveDevice),
	}
}

// Initialize initializes the Z-Wave adapter with the given configuration.
func (a *ZWaveAdapter) Initialize(ctx context.Context, config iot.ProtocolConfig) error {
	// Get Z-Wave JS URL from metadata
	wsURL := config.Metadata["ws_url"]
	if wsURL == "" {
		// Default URL
		wsURL = "ws://localhost:3000"
	}

	// Create Z-Wave JS client
	a.jsClient = NewZWaveJSClient(wsURL)

	// Initialize device registry
	a.storage = *iot.NewDeviceRegistry()

	return nil
}

// Start starts the Z-Wave adapter.
func (a *ZWaveAdapter) Start(ctx context.Context) error {
	// Connect to Z-Wave JS server
	if err := a.jsClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to Z-Wave JS: %w", err)
	}

	// Set message handler for events
	a.jsClient.SetMessageHandler(a.handleMessage)

	// Mark adapter as running
	a.BaseAdapter.Start(ctx)

	return nil
}

// Stop stops the Z-Wave adapter.
func (a *ZWaveAdapter) Stop(ctx context.Context) error {
	// Disconnect from Z-Wave JS server
	if err := a.jsClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from Z-Wave JS: %w", err)
	}

	// Mark adapter as stopped
	a.BaseAdapter.Stop(ctx)

	return nil
}

// DiscoverDevices discovers Z-Wave devices on the network.
func (a *ZWaveAdapter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*iot.DeviceInfo, error) {
	if !a.IsRunning() {
		return nil, iot.ErrAdapterNotRunning
	}

	// Get all nodes from Z-Wave JS
	nodes, err := a.jsClient.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	devices := make([]*iot.DeviceInfo, 0, len(nodes))
	for _, node := range nodes {
		// Parse node information
		nodeID, ok := node["nodeId"].(float64)
		if !ok {
			continue
		}

		deviceInfo := &iot.DeviceInfo{
			ID:       fmt.Sprintf("zwave-node-%d", int(nodeID)),
			Name:     getNodeName(node),
			Type:     getZWaveDeviceType(node),
			Protocol: iot.ProtocolZWave,
			Status:   iot.DeviceStatusOnline,
			Properties: map[string]interface{}{
				"node_id":    uint8(nodeID),
				"manufacturer": getManufacturer(node),
				"product":     getProduct(node),
				"location":    getLocation(node),
			},
			LastSeen: time.Now(),
			Capabilities: getZWaveCapabilities(node),
		}

		// Create device instance
		device := a.createZWaveDevice(deviceInfo)
		a.devices[uint8(nodeID)] = device

		// Register in storage
		a.storage.Register(deviceInfo)

		// Publish discovery event
		a.PublishEvent(iot.ProtocolEvent{
			Type:      iot.EventDeviceDiscovered,
			Protocol:  iot.ProtocolZWave,
			DeviceID:  deviceInfo.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"device": deviceInfo,
			},
		})

		devices = append(devices, deviceInfo)
	}

	return devices, nil
}

// StartPairing starts device pairing (inclusion) for Z-Wave devices.
func (a *ZWaveAdapter) StartPairing(ctx context.Context, timeout time.Duration) (*iot.PairingResult, error) {
	if !a.IsRunning() {
		return nil, iot.ErrAdapterNotRunning
	}

	// Start inclusion mode
	includeNonSecure := true
	if err := a.jsClient.StartInclusion(ctx, includeNonSecure); err != nil {
		return &iot.PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Wait for node to be added
	resultChan := make(chan *iot.PairingResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		device, err := a.waitForInclusion(ctx, timeout)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- device
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return &iot.PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	case <-ctx.Done():
		// Stop inclusion on cancel
		_ = a.jsClient.StopInclusion(ctx)
		return &iot.PairingResult{
			Success: false,
			Error:   "pairing timeout",
		}, nil
	}
}

// CancelPairing cancels ongoing device pairing.
func (a *ZWaveAdapter) CancelPairing(ctx context.Context) error {
	return a.jsClient.StopInclusion(ctx)
}

// GetDevice retrieves a Z-Wave device by ID.
func (a *ZWaveAdapter) GetDevice(ctx context.Context, deviceID string) (iot.IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Parse node ID from device ID
	var nodeID uint8
	_, err := fmt.Sscanf(deviceID, "zwave-node-%d", &nodeID)
	if err != nil {
		return nil, iot.ErrDeviceNotFound
	}

	device, exists := a.devices[nodeID]
	if !exists {
		return nil, iot.ErrDeviceNotFound
	}

	return device, nil
}

// ListDevices lists all Z-Wave devices.
func (a *ZWaveAdapter) ListDevices(ctx context.Context) ([]iot.IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	devices := make([]iot.IoTDevice, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// RemoveDevice removes a Z-Wave device.
func (a *ZWaveAdapter) RemoveDevice(ctx context.Context, deviceID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Parse node ID from device ID
	var nodeID uint8
	_, err := fmt.Sscanf(deviceID, "zwave-node-%d", &nodeID)
	if err != nil {
		return iot.ErrDeviceNotFound
	}

	_, exists := a.devices[nodeID]
	if !exists {
		return iot.ErrDeviceNotFound
	}

	// Remove from adapter
	delete(a.devices, nodeID)

	// Remove from storage
	a.storage.Unregister(deviceID)

	// Publish event
	a.PublishEvent(iot.ProtocolEvent{
		Type:      iot.EventDeviceLeft,
		Protocol:  iot.ProtocolZWave,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"device_id": deviceID,
		},
	})

	return nil
}

// GetNetworkInfo retrieves Z-Wave network information.
func (a *ZWaveAdapter) GetNetworkInfo(ctx context.Context) (*iot.NetworkInfo, error) {
	if !a.IsRunning() {
		return nil, iot.ErrAdapterNotRunning
	}

	// Get driver info from Z-Wave JS
	info, err := a.jsClient.GetDriverInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Extract network information
	status := "running"
	if running, ok := info["status"].(bool); ok && !running {
		status = "stopped"
	}

	return &iot.NetworkInfo{
		PanID:        0, // Z-Wave doesn't use PAN ID
		Channel:      0, // Z-Wave uses dynamic channels
		DeviceCount:  len(a.devices),
		Status:       status,
		Properties: map[string]string{
			"driver_version":   getStringFromInfo(info, "version"),
			"controller_type": getStringFromInfo(info, "controllerType"),
			"home_id":          fmt.Sprintf("%d", getIntFromInfo(info, "homeId")),
		},
	}, nil
}

// ResetNetwork resets the Z-Wave network.
func (a *ZWaveAdapter) ResetNetwork(ctx context.Context) error {
	// Heal the network instead of reset
	return a.jsClient.HealNetwork(ctx)
}

// SetEventHandler sets the event handler for protocol events.
func (a *ZWaveAdapter) SetEventHandler(handler iot.ProtocolEventHandler) {
	a.eventHandler = handler
}

// PublishEvent publishes an event to the event handler.
func (a *ZWaveAdapter) PublishEvent(event iot.ProtocolEvent) {
	if a.eventHandler != nil {
		a.eventHandler(event)
	}
}

// handleMessage handles incoming messages from Z-Wave JS.
func (a *ZWaveAdapter) handleMessage(message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "event":
		// Handle Z-Wave event
		a.handleZWaveEvent(message)
	case "node_added":
		// Handle node added
		a.handleNodeAdded(message)
	case "node_removed":
		// Handle node removed
		a.handleNodeRemoved(message)
	case "value_updated":
		// Handle value updated
		a.handleValueUpdated(message)
	}
}

// handleZWaveEvent handles Z-Wave events.
func (a *ZWaveAdapter) handleZWaveEvent(message map[string]interface{}) {
	eventType, ok := message["event"].(string)
	if !ok {
		return
	}

	nodeID := getUint8FromMessage(message, "nodeId")
	if nodeID == 0 {
		return
	}

	deviceID := fmt.Sprintf("zwave-node-%d", nodeID)

	// Publish event
	a.PublishEvent(iot.ProtocolEvent{
		Type:      iot.EventType("zwave_" + eventType),
		Protocol:  iot.ProtocolZWave,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data:      message,
	})
}

// handleNodeAdded handles node added event.
func (a *ZWaveAdapter) handleNodeAdded(message map[string]interface{}) {
	nodeID := getUint8FromMessage(message, "nodeId")
	if nodeID == 0 {
		return
	}

	// Refresh device list
	_, _ = a.DiscoverDevices(context.Background(), 5*time.Second)
}

// handleNodeRemoved handles node removed event.
func (a *ZWaveAdapter) handleNodeRemoved(message map[string]interface{}) {
	nodeID := getUint8FromMessage(message, "nodeId")
	if nodeID == 0 {
		return
	}

	deviceID := fmt.Sprintf("zwave-node-%d", nodeID)

	// Remove device
	delete(a.devices, nodeID)
	a.storage.Unregister(deviceID)

	// Publish event
	a.PublishEvent(iot.ProtocolEvent{
		Type:      iot.EventDeviceLeft,
		Protocol:  iot.ProtocolZWave,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data:      message,
	})
}

// handleValueUpdated handles value updated event.
func (a *ZWaveAdapter) handleValueUpdated(message map[string]interface{}) {
	nodeID := getUint8FromMessage(message, "nodeId")
	if nodeID == 0 {
		return
	}

	deviceID := fmt.Sprintf("zwave-node-%d", nodeID)

	// Publish event
	a.PublishEvent(iot.ProtocolEvent{
		Type:      iot.EventDataReceived,
		Protocol:  iot.ProtocolZWave,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data:      message,
	})
}

// waitForInclusion waits for a node to be included during inclusion mode.
func (a *ZWaveAdapter) waitForInclusion(ctx context.Context, timeout time.Duration) (*iot.PairingResult, error) {
	// Watch for node_added events
	resultChan := make(chan *iot.DeviceInfo, 1)
	errorChan := make(chan error, 1)

	handler := func(message map[string]interface{}) {
		if msgType, ok := message["type"].(string); ok {
			if msgType == "node_added" {
				nodeID := getUint8FromMessage(message, "nodeId")
				if nodeID != 0 {
					deviceInfo := &iot.DeviceInfo{
						ID:       fmt.Sprintf("zwave-node-%d", nodeID),
						Name:     getNodeName(message),
						Type:     getZWaveDeviceType(message),
						Protocol: iot.ProtocolZWave,
						Status:   iot.DeviceStatusOnline,
						Properties: map[string]interface{}{
							"node_id": nodeID,
						},
						LastSeen: time.Now(),
					}
					resultChan <- deviceInfo
				}
			}
		}
	}

	a.jsClient.SetMessageHandler(handler)
	defer a.jsClient.SetMessageHandler(a.handleMessage)

	select {
	case deviceInfo := <-resultChan:
		// Create device instance
		device := a.createZWaveDevice(deviceInfo)
		a.devices[getUint8FromProps(deviceInfo.Properties, "node_id")] = device
		a.storage.Register(deviceInfo)

		return &iot.PairingResult{
			Success: true,
			Device:  deviceInfo,
		}, nil

	case err := <-errorChan:
		return nil, err

	case <-time.After(timeout):
		return nil, iot.ErrPairingTimeout
	}
}

// createZWaveDevice creates a ZWaveDevice from device info.
func (a *ZWaveAdapter) createZWaveDevice(info *iot.DeviceInfo) *ZWaveDevice {
	return &ZWaveDevice{
		BaseDevice: iot.NewBaseDevice(info),
		info:      info,
		nodeID:    getUint8FromProps(info.Properties, "node_id"),
		adapter:   a,
	}
}

// Helper functions

func getNodeName(node map[string]interface{}) string {
	if name, ok := node["name"].(string); ok && name != "" {
		return name
	}
	nodeID := getUint8FromMessage(node, "nodeId")
	return fmt.Sprintf("Z-Wave Node %d", nodeID)
}

func getManufacturer(node map[string]interface{}) string {
	if values, ok := node["values"].([]map[string]interface{}); ok {
		for _, v := range values {
			if propertyName, ok := v["propertyName"].(string); ok {
				if propertyName == "manufacturerName" {
					if value, ok := v["value"].(string); ok {
						return value
					}
				}
			}
		}
	}
	return "Unknown"
}

func getProduct(node map[string]interface{}) string {
	if values, ok := node["values"].([]map[string]interface{}); ok {
		for _, v := range values {
			if propertyName, ok := v["propertyName"].(string); ok {
				if propertyName == "productName" {
					if value, ok := v["value"].(string); ok {
						return value
					}
				}
			}
		}
	}
	return "Unknown"
}

func getLocation(node map[string]interface{}) string {
	if location, ok := node["location"].(string); ok {
		return location
	}
	return "Unknown"
}

func getZWaveDeviceType(node map[string]interface{}) iot.DeviceType {
	// Determine device type from device classes
	deviceClass := getDeviceClass(node)

	switch {
	case isSensor(deviceClass):
		return iot.DeviceTypeSensor
	case isActuator(deviceClass):
		return iot.DeviceTypeActuator
	default:
		return iot.DeviceTypeSensor
	}
}

func getDeviceClass(node map[string]interface{}) string {
	if basic, ok := node["basic"].(float64); ok {
		return fmt.Sprintf("0x%04X", uint16(basic))
	}
	return "Unknown"
}

func isSensor(deviceClass string) bool {
	// Z-Wave sensor device classes
	sensorClasses := []string{
		"0x0001", // Door lock
		"0x0020", // Switch
		"0x0021", // Switch binary
		"0x0030", // Sensor binary
		"0x0031", // Sensor multilevel
		"0x0040", // Meter
	}

	for _, class := range sensorClasses {
		if len(deviceClass) >= 4 && deviceClass[:4] == class {
			return true
		}
	}
	return false
}

func isActuator(deviceClass string) bool {
	// Z-Wave actuator device classes
	actuatorClasses := []string{
		"0x0001", // Door lock
		"0x0020", // Switch
		"0x0021", // Switch binary
	}

	for _, class := range actuatorClasses {
		if len(deviceClass) >= 4 && deviceClass[:4] == class {
			return true
		}
	}
	return false
}

func getZWaveCapabilities(node map[string]interface{}) []iot.DeviceCapability {
	capabilities := []iot.DeviceCapability{
		iot.CapabilitySensor,
	}

	deviceClass := getDeviceClass(node)

	// Add capabilities based on device class
	if isActuator(deviceClass) {
		capabilities = append(capabilities, iot.CapabilityOnOff)
	}

	return capabilities
}

func getUint8FromMessage(message map[string]interface{}, key string) uint8 {
	if val, ok := message[key]; ok {
		switch v := val.(type) {
		case float64:
			return uint8(v)
		case int:
			return uint8(v)
		case string:
			var num uint8
			fmt.Sscanf(v, "%d", &num)
			return num
		}
	}
	return 0
}

func getUint8FromProps(props map[string]interface{}, key string) uint8 {
	if val, ok := props[key]; ok {
		switch v := val.(type) {
		case float64:
			return uint8(v)
		case int:
			return uint8(v)
		case uint8:
			return v
		}
	}
	return 0
}

func getStringFromInfo(info map[string]interface{}, key string) string {
	if val, ok := info[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromInfo(info map[string]interface{}, key string) int {
	if val, ok := info[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}
	return 0
}
