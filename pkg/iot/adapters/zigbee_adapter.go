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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/iot"
)

// ZigbeeAdapter implements ProtocolAdapter for Zigbee via Zigbee2MQTT.
type ZigbeeAdapter struct {
	*iot.BaseAdapter
	mqttClient   *ZigbeeMQTTClient
	devices      map[string]*ZigbeeDevice
	storage      iot.DeviceRegistry
	discoveryCtx context.Context
	cancelFunc   context.CancelFunc
	mutex        sync.RWMutex
}

// NewZigbeeAdapter creates a new Zigbee adapter.
func NewZigbeeAdapter() *ZigbeeAdapter {
	ctx, cancel := context.WithCancel(context.Background())

	return &ZigbeeAdapter{
		BaseAdapter:  iot.NewBaseAdapter(iot.ProtocolZigbee, "3.0"),
		devices:     make(map[string]*ZigbeeDevice),
		storage:     iot.NewDeviceRegistry(),
		discoveryCtx: ctx,
		cancelFunc:   cancel,
	}
}

// Initialize initializes the Zigbee adapter.
func (a *ZigbeeAdapter) Initialize(ctx context.Context, config iot.ProtocolConfig) error {
	if err := a.BaseAdapter.Initialize(ctx, config); err != nil {
		return err
	}

	// Create MQTT client
	brokerURL, ok := config.Hardware.Metadata["broker_url"].(string)
	if !ok {
		brokerURL = "tcp://localhost:1883" // Default
	}

	topicPrefix := config.Hardware.Metadata["topic_prefix"]
	if topicPrefix == "" {
		topicPrefix = "zigbee2mqtt"
	}

	a.mqttClient = NewZigbeeMQTTClient(brokerURL, topicPrefix)

	// Set message handler
	a.mqttClient.SetMessageHandler(a.handleMQTTMessage)

	return nil
}

// Start starts the Zigbee adapter.
func (a *ZigbeeAdapter) Start(ctx context.Context) error {
	if err := a.BaseAdapter.Start(ctx); err != nil {
		return err
	}

	clientID := fmt.Sprintf("zigbee_adapter_%d", time.Now().Unix())
	if err := a.mqttClient.Connect(ctx, clientID); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	// Subscribe to all device topics
	if err := a.mqttClient.SubscribeToAllDevices(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to device topics: %w", err)
	}

	// Load existing devices
	if err := a.loadExistingDevices(ctx); err != nil {
		// Non-fatal error, log and continue
		a.emitEvent(iot.EventError, "", map[string]interface{}{
			"warning": "failed to load existing devices",
			"error":   err.Error(),
		}, nil)
	}

	return nil
}

// Stop stops the Zigbee adapter.
func (a *ZigbeeAdapter) Stop(ctx context.Context) error {
	if !a.BaseAdapter.IsRunning() {
		return nil
	}

	a.cancelFunc()

	if err := a.mqttClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MQTT broker: %w", err)
	}

	return a.BaseAdapter.Stop(ctx)
}

// DiscoverDevices discovers Zigbee devices.
func (a *ZigbeeAdapter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*iot.DeviceInfo, error) {
	devices, err := a.mqttClient.GetDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	deviceInfos := make([]*iot.DeviceInfo, 0, len(devices))
	for _, device := range devices {
		deviceInfos = append(deviceInfos, a.convertToDeviceInfo(device))
	}

	return deviceInfos, nil
}

// StartPairing starts device pairing (permit join).
func (a *ZigbeeAdapter) StartPairing(ctx context.Context, timeout time.Duration) (*iot.PairingResult, error) {
	// Enable permit join
	if err := a.mqttClient.PermitJoin(ctx, true); err != nil {
		return &iot.PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Create a channel to capture the joined device
	deviceChan := make(chan *ZigbeeDevice, 1)

	// Set up temporary handler for device join
	originalHandler := a.mqttClient.handler
	a.mqttClient.SetMessageHandler(func(topic string, payload []byte) {
		// Call original handler
		if originalHandler != nil {
			originalHandler(topic, payload)
		}

		// Check if this is a new device
		var deviceUpdate struct {
			FriendlyName string                 `json:"friendly_name"`
			Type         string                 `json:"type"`
		}

		if err := json.Unmarshal(payload, &deviceUpdate); err == nil {
			// Extract device ID from topic
			deviceID := extractDeviceIDFromTopic(topic)
			if deviceID != "" {
				// Check if this is a new device
				a.mutex.Lock()
				_, exists := a.devices[deviceID]
				a.mutex.Unlock()

				if !exists {
					// New device joined!
					zigbeeDevice := &ZigbeeDevice{
						ID:           deviceID,
						FriendlyName: deviceUpdate.FriendlyName,
						Type:         deviceUpdate.Type,
						adapter:      a,
						state:        make(map[string]interface{}),
					}

					select {
					case deviceChan <- zigbeeDevice:
					case <-time.After(timeout):
					}
				}
			}
		}
	})

	// Wait for device to join or timeout
	select {
	case device := <-deviceChan:
		// Disable permit join after device joins
		a.mqttClient.PermitJoin(ctx, false)

		// Restore original handler
		a.mqttClient.SetMessageHandler(originalHandler)

		return &iot.PairingResult{
			Success: true,
			Device:  a.convertToDeviceInfo(device),
		}, nil

	case <-time.After(timeout):
		// Timeout
		a.mqttClient.PermitJoin(ctx, false)
		a.mqttClient.SetMessageHandler(originalHandler)

		return &iot.PairingResult{
			Success: false,
			Error:   "pairing timeout",
		}, nil

	case <-ctx.Done():
		a.mqttClient.PermitJoin(ctx, false)
		a.mqttClient.SetMessageHandler(originalHandler)

		return nil, ctx.Err()
	}
}

// CancelPairing cancels ongoing pairing.
func (a *ZigbeeAdapter) CancelPairing(ctx context.Context) error {
	return a.mqttClient.PermitJoin(ctx, false)
}

// RemoveDevice removes a device from the network.
func (a *ZigbeeAdapter) RemoveDevice(ctx context.Context, deviceID string) error {
	// Send remove command to Zigbee2MQTT
	topic := fmt.Sprintf("%s/bridge/config/remove", a.mqttClient.GetDeviceTopic(""))
	payload := map[string]string{
		"device": deviceID,
		"force":  "true",
	}

	if err := a.mqttClient.Publish(ctx, topic, payload); err != nil {
		return err
	}

	// Remove from local cache
	a.mutex.Lock()
	delete(a.devices, deviceID)
	a.mutex.Unlock()

	return a.BaseAdapter.removeDevice(deviceID)
}

// GetDevice retrieves a device by ID.
func (a *ZigbeeAdapter) GetDevice(ctx context.Context, deviceID string) (iot.IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	device, exists := a.devices[deviceID]
	if !exists {
		return nil, iot.ErrDeviceNotFound
	}

	return device, nil
}

// ListDevices lists all Zigbee devices.
func (a *ZigbeeAdapter) ListDevices(ctx context.Context) ([]iot.IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	devices := make([]iot.IoTDevice, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// GetNetworkInfo retrieves network information.
func (a *ZigbeeAdapter) GetNetworkInfo(ctx context.Context) (*iot.NetworkInfo, error) {
	// Get network info from Zigbee2MQTT bridge
	topic := fmt.Sprintf("%s/bridge/config", a.mqttClient.GetDeviceTopic(""))

	// In production, you would subscribe to the response topic and request the info
	// For now, return basic info
	return &iot.NetworkInfo{
		PanID:       0, // Zigbee2MQTT manages this
		Channel:     11,
		DeviceCount: len(a.devices),
		Status:      "online",
		PermitJoin:  false,
		Properties:  make(map[string]string),
	}, nil
}

// ResetNetwork resets the Zigbee network.
func (a *ZigbeeAdapter) ResetNetwork(ctx context.Context) error {
	// This is a dangerous operation - in production, require explicit confirmation
	return fmt.Errorf("network reset must be performed manually through Zigbee2MQTT")
}

// handleMQTTMessage handles incoming MQTT messages from Zigbee2MQTT.
func (a *ZigbeeAdapter) handleMQTTMessage(topic string, payload []byte) {
	deviceID := extractDeviceIDFromTopic(topic)
	if deviceID == "" {
		return
	}

	a.mutex.RLock()
	device, exists := a.devices[deviceID]
	a.mutex.RUnlock()

	if !exists {
		// Device doesn't exist, might be a new device
		return
	}

	// Parse state update
	var state map[string]interface{}
	if err := json.Unmarshal(payload, &state); err != nil {
		return
	}

	// Update device state
	device.updateState(state)

	// Emit event
	a.emitEvent(iot.EventDataReceived, deviceID, state, nil)
}

// loadExistingDevices loads devices that are already paired.
func (a *ZigbeeAdapter) loadExistingDevices(ctx context.Context) error {
	devices, err := a.mqttClient.GetDevices(ctx)
	if err != nil {
		return err
	}

	for _, device := range devices {
		zigbeeDevice := &ZigbeeDevice{
			ID:           device.ID,
			FriendlyName: device.FriendlyName,
			Type:         device.Type,
			Definition:   device.Definition,
			adapter:      a,
			state:        make(map[string]interface{}),
		}

		a.mutex.Lock()
		a.devices[device.ID] = zigbeeDevice
		a.mutex.Unlock()

		// Register in base adapter
		a.BaseAdapter.addDevice(zigbeeDevice)
	}

	return nil
}

// convertToDeviceInfo converts a Zigbee device to DeviceInfo.
func (a *ZigbeeAdapter) convertToDeviceInfo(device *ZigbeeDevice) *iot.DeviceInfo {
	var deviceType iot.DeviceType
	switch device.Type {
	case "EndDevice":
		deviceType = iot.DeviceTypeSensor
	case "Router":
		deviceType = iot.DeviceTypeController
	case "Coordinator":
		deviceType = iot.DeviceTypeGateway
	default:
		deviceType = iot.DeviceTypeUnknown
	}

	info := &iot.DeviceInfo{
		ID:       device.ID,
		Name:     device.FriendlyName,
		Type:     deviceType,
		Protocol: iot.ProtocolZigbee,
		Status:   iot.DeviceStatusOnline,
		Properties: map[string]interface{}{
			"definition": device.Definition,
			"power_source": device.PowerSource,
		},
		LastSeen: time.Now(),
	}

	return info
}

// convertToDeviceInfo converts ZigbeeMQTTDevice to DeviceInfo.
func (a *ZigbeeAdapter) convertToDeviceInfoFromMQTT(mqttDevice *ZigbeeDevice) *iot.DeviceInfo {
	info := &iot.DeviceInfo{
		ID:       mqttDevice.ID,
		Name:     mqttDevice.FriendlyName,
		Protocol: iot.ProtocolZigbee,
		Status:   iot.DeviceStatusOnline,
		LastSeen: mqttDevice.LastSeen,
	}

	// Determine device type
	if mqttDevice.PowerSource == "Battery" {
		info.Type = iot.DeviceTypeSensor
	} else {
		info.Type = iot.DeviceTypeActuator
	}

	info.Properties = map[string]interface{}{
		"type":        mqttDevice.Type,
		"definition":  mqttDevice.Definition,
		"power_source": mqttDevice.PowerSource,
		"supported":   mqttDevice.Supported,
	}

	return info
}

// extractDeviceIDFromTopic extracts device ID from MQTT topic.
func extractDeviceIDFromTopic(topic string) string {
	// Topic format: zigbee2mqtt/<device_id>/...
	// We need to extract the device ID
	parts := splitTopic(topic)
	if len(parts) >= 2 && parts[0] == "zigbee2mqtt" {
		return parts[1]
	}
	return ""
}

// splitTopic splits MQTT topic into parts.
func splitTopic(topic string) []string {
	var parts []string
	current := ""
	for _, ch := range topic {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
