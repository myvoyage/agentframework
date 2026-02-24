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
	"net"
	"sync"
	"time"

)

// NearLinkAdapter implements ProtocolAdapter for NearLink (Spark Link) networks.
// NearLink is a next-generation wireless communication technology developed by Huawei.
type NearLinkAdapter struct {
	*BaseAdapter
	controller  *NearLinkController
	devices     map[string]*NearLinkDevice
	storage     DeviceRegistry
	mutex       sync.RWMutex
	eventHandler ProtocolEventHandler
	networkMode NearLinkNetworkMode // SLM or SLE
}

// NearLinkNetworkMode represents the NearLink network mode.
type NearLinkNetworkMode string

const (
	// NearLinkModeSLM is Spark Link Mesh mode (high throughput, low latency)
	NearLinkModeSLM NearLinkNetworkMode = "SLM"
	// NearLinkModeSLE is Spark Link Low Energy mode (ultra-low power)
	NearLinkModeSLE NearLinkNetworkMode = "SLE"
)

// NearLinkController manages NearLink network communication.
type NearLinkController struct {
	udpConn    *net.UDPConn
	multicastAddr string
	isConnected bool
	mode       NearLinkNetworkMode
	channel    uint8 // 2.4GHz or 5.1GHz
	meshID     uint64
}

// NewNearLinkAdapter creates a new NearLink protocol adapter.
func NewNearLinkAdapter() *NearLinkAdapter {
	return &NearLinkAdapter{
		BaseAdapter: NewBaseAdapter(ProtocolNearLink, "1.0.0"),
		devices:     make(map[string]*NearLinkDevice),
		networkMode: NearLinkModeSLM, // Default to SLM mode
	}
}

// Initialize initializes the NearLink adapter with the given configuration.
func (a *NearLinkAdapter) Initialize(ctx context.Context, config ProtocolConfig) error {
	// Get network mode from metadata
	if mode, ok := config.Metadata["network_mode"]; ok {
		a.networkMode = NearLinkNetworkMode(mode)
	}

	// Create NearLink controller
	controller, err := NewNearLinkController(config.Metadata)
	if err != nil {
		return fmt.Errorf("failed to create NearLink controller: %w", err)
	}
	a.controller = controller

	// Initialize device registry
	a.storage = *NewDeviceRegistry()

	return nil
}

// Start starts the NearLink adapter.
func (a *NearLinkAdapter) Start(ctx context.Context) error {
	// Connect to NearLink network
	if err := a.controller.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to NearLink network: %w", err)
	}

	// Start network discovery listener
	go a.listenForDevices(ctx)

	// Mark adapter as running
	a.BaseAdapter.Start(ctx)

	return nil
}

// Stop stops the NearLink adapter.
func (a *NearLinkAdapter) Stop(ctx context.Context) error {
	// Disconnect from NearLink network
	if err := a.controller.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from NearLink network: %w", err)
	}

	// Mark adapter as stopped
	a.BaseAdapter.Stop(ctx)

	return nil
}

// DiscoverDevices discovers NearLink devices on the network.
func (a *NearLinkAdapter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*DeviceInfo, error) {
	if !a.IsRunning() {
		return nil, ErrAdapterNotRunning
	}

	// Send discovery broadcast
	if err := a.controller.SendDiscovery(ctx); err != nil {
		return nil, fmt.Errorf("failed to send discovery: %w", err)
	}

	// Wait for device responses
	resultChan := make(chan []*DeviceInfo, 1)
	errorChan := make(chan error, 1)

	go func() {
		devices := make([]*DeviceInfo, 0)
		timeoutTimer := time.NewTimer(timeout)
		defer timeoutTimer.Stop()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeoutTimer.C:
				resultChan <- devices
				return
			case <-ticker.C:
				// Collect discovered devices
				a.mutex.RLock()
				for _, device := range a.devices {
					deviceInfo := device.GetInfo()
					devices = append(devices, deviceInfo)
				}
				a.mutex.RUnlock()
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			}
		}
	}()

	select {
	case devices := <-resultChan:
		return devices, nil
	case err := <-errorChan:
		return nil, err
	}
}

// StartPairing starts device pairing for NearLink devices.
func (a *NearLinkAdapter) StartPairing(ctx context.Context, timeout time.Duration) (*PairingResult, error) {
	if !a.IsRunning() {
		return nil, ErrAdapterNotRunning
	}

	// Enter pairing mode (enable device discovery)
	if err := a.controller.EnablePairing(ctx); err != nil {
		return &PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Wait for new device
	resultChan := make(chan *PairingResult, 1)

	go func() {
		timeoutTimer := time.NewTimer(timeout)
		defer timeoutTimer.Stop()

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		lastDeviceCount := len(a.devices)

		for {
			select {
			case <-timeoutTimer.C:
				resultChan <- &PairingResult{
					Success: false,
					Error:   "pairing timeout",
				}
				return
			case <-ticker.C:
				a.mutex.RLock()
				currentDeviceCount := len(a.devices)
				a.mutex.RUnlock()

				if currentDeviceCount > lastDeviceCount {
					// New device found
					a.mutex.RLock()
					for deviceID, device := range a.devices {
						if !isDevicePaired(deviceID) {
							resultChan <- &PairingResult{
								Success: true,
								Device:  device.GetInfo(),
							}
							a.mutex.RUnlock()
							return
						}
					}
					a.mutex.RUnlock()
				}
			case <-ctx.Done():
				resultChan <- &PairingResult{
					Success: false,
					Error:   "pairing canceled",
				}
				return
			}
		}
	}()

	result := <-resultChan

	// Disable pairing mode
	_ = a.controller.DisablePairing(ctx)

	return result, nil
}

// CancelPairing cancels ongoing device pairing.
func (a *NearLinkAdapter) CancelPairing(ctx context.Context) error {
	return a.controller.DisablePairing(ctx)
}

// GetDevice retrieves a NearLink device by ID.
func (a *NearLinkAdapter) GetDevice(ctx context.Context, deviceID string) (IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	device, exists := a.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	return device, nil
}

// ListDevices lists all NearLink devices.
func (a *NearLinkAdapter) ListDevices(ctx context.Context) ([]IoTDevice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	devices := make([]IoTDevice, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// RemoveDevice removes a NearLink device.
func (a *NearLinkAdapter) RemoveDevice(ctx context.Context, deviceID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	_, exists := a.devices[deviceID]
	if !exists {
		return ErrDeviceNotFound
	}

	// Send unpair command
	if err := a.controller.UnpairDevice(ctx, deviceID); err != nil {
		return err
	}

	// Remove from adapter
	delete(a.devices, deviceID)

	// Remove from storage
	a.storage.Unregister(deviceID)

	// Publish event
	a.PublishEvent(ProtocolEvent{
		Type:      EventDeviceLeft,
		Protocol:  ProtocolNearLink,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"device_id": deviceID,
		},
	})

	return nil
}

// GetNetworkInfo retrieves NearLink network information.
func (a *NearLinkAdapter) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	if !a.IsRunning() {
		return nil, ErrAdapterNotRunning
	}

	status, err := a.controller.GetNetworkStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &NetworkInfo{
		PanID:       uint16(a.controller.meshID),
		Channel:     a.controller.channel,
		DeviceCount: len(a.devices),
		Status:      status,
		Properties: map[string]string{
			"network_mode": string(a.networkMode),
			"frequency":    getFrequencyString(a.controller.channel),
		},
	}, nil
}

// ResetNetwork resets the NearLink network.
func (a *NearLinkAdapter) ResetNetwork(ctx context.Context) error {
	return a.controller.ResetNetwork(ctx)
}

// SetEventHandler sets the event handler for protocol events.
func (a *NearLinkAdapter) SetEventHandler(handler ProtocolEventHandler) {
	a.eventHandler = handler
}

// PublishEvent publishes an event to the event handler.
func (a *NearLinkAdapter) PublishEvent(event ProtocolEvent) {
	if a.eventHandler != nil {
		a.eventHandler(event)
	}
}

// listenForDevices listens for device discovery messages.
func (a *NearLinkAdapter) listenForDevices(ctx context.Context) {
	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, addr, err := a.controller.udpConn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			// Parse device announcement
			deviceInfo, err := parseNearLinkDeviceAnnouncement(buf[:n], addr)
			if err != nil {
				continue
			}

			// Create device
			device := a.createNearLinkDevice(deviceInfo)

			a.mutex.Lock()
			a.devices[deviceInfo.ID] = device
			a.mutex.Unlock()

			// Register in storage
			a.storage.Register(deviceInfo)

			// Publish discovery event
			a.PublishEvent(ProtocolEvent{
				Type:      EventDeviceDiscovered,
				Protocol:  ProtocolNearLink,
				DeviceID:  deviceInfo.ID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"device": deviceInfo,
				},
			})
		}
	}
}

// createNearLinkDevice creates a NearLinkDevice from device info.
func (a *NearLinkAdapter) createNearLinkDevice(info *DeviceInfo) *NearLinkDevice {
	return &NearLinkDevice{
		BaseDevice: NewBaseDevice(info),
		info:       info,
		macAddr:    getStringFromInterface(info.Properties, "mac_address", ""),
		adapter:    a,
	}
}

// NewNearLinkController creates a new NearLink controller.
func NewNearLinkController(metadata map[string]string) (*NearLinkController, error) {
	// Default multicast address for NearLink
	multicastAddr := "224.0.0.1:1888"
	if addr, ok := metadata["multicast_addr"]; ok {
		multicastAddr = addr
	}

	// Default channel (2.4GHz)
	channel := uint8(0)
	if ch, ok := metadata["channel"]; ok {
		fmt.Sscanf(ch, "%d", &channel)
	}

	// Mesh ID
	var meshID uint64
	if id, ok := metadata["mesh_id"]; ok {
		fmt.Sscanf(id, "%d", &meshID)
	}

	return &NearLinkController{
		multicastAddr: multicastAddr,
		mode:          NearLinkModeSLM,
		channel:       channel,
		meshID:        meshID,
	}, nil
}

// Connect connects to the NearLink network.
func (c *NearLinkController) Connect(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", c.multicastAddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	c.udpConn = conn
	c.isConnected = true

	return nil
}

// Disconnect disconnects from the NearLink network.
func (c *NearLinkController) Disconnect(ctx context.Context) error {
	if c.udpConn != nil {
		_ = c.udpConn.Close()
		c.udpConn = nil
	}
	c.isConnected = false
	return nil
}

// SendDiscovery sends a device discovery broadcast.
func (c *NearLinkController) SendDiscovery(ctx context.Context) error {
	if !c.isConnected {
		return fmt.Errorf("controller not connected")
	}

	// Send discovery broadcast message
	message := buildNearLinkDiscoveryMessage()
	_, err := c.udpConn.Write(message)
	return err
}

// EnablePairing enables pairing mode.
func (c *NearLinkController) EnablePairing(ctx context.Context) error {
	// Send pairing enable command
	message := buildNearLinkPairingMessage(true)
	_, err := c.udpConn.Write(message)
	return err
}

// DisablePairing disables pairing mode.
func (c *NearLinkController) DisablePairing(ctx context.Context) error {
	// Send pairing disable command
	message := buildNearLinkPairingMessage(false)
	_, err := c.udpConn.Write(message)
	return err
}

// UnpairDevice unpairs a device.
func (c *NearLinkController) UnpairDevice(ctx context.Context, deviceID string) error {
	// Send unpair command to device
	message := buildNearLinkUnpairMessage(deviceID)
	_, err := c.udpConn.Write(message)
	return err
}

// GetNetworkStatus gets the current network status.
func (c *NearLinkController) GetNetworkStatus(ctx context.Context) (string, error) {
	if c.isConnected {
		return "running", nil
	}
	return "stopped", nil
}

// ReadDeviceAttribute reads an attribute from a NearLink device.
func (c *NearLinkController) ReadDeviceAttribute(ctx context.Context, deviceMAC string, attribute string) (interface{}, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("controller not connected")
	}

	// Send read attribute command
	message := buildNearLinkReadAttributeMessage(deviceMAC, attribute)
	_, err := c.udpConn.Write(message)
	if err != nil {
		return nil, err
	}

	// In a real implementation, we would wait for the response
	// For now, return a placeholder
	return map[string]interface{}{
		"attribute": attribute,
		"value":     "placeholder",
	}, nil
}

// WriteDeviceAttribute writes an attribute to a NearLink device.
func (c *NearLinkController) WriteDeviceAttribute(ctx context.Context, deviceMAC string, attribute string, value interface{}) error {
	if !c.isConnected {
		return fmt.Errorf("controller not connected")
	}

	// Send write attribute command
	message := buildNearLinkWriteAttributeMessage(deviceMAC, attribute, value)
	_, err := c.udpConn.Write(message)
	return err
}

// SendDeviceCommand sends a command to a NearLink device.
func (c *NearLinkController) SendDeviceCommand(ctx context.Context, deviceMAC string, command string, params map[string]interface{}) (interface{}, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("controller not connected")
	}

	// Send command to device
	message := buildNearLinkCommandMessage(deviceMAC, command, params)
	_, err := c.udpConn.Write(message)
	if err != nil {
		return nil, err
	}

	// In a real implementation, we would wait for the response
	return map[string]interface{}{
		"command": command,
		"status":  "sent",
	}, nil
}

// GetDeviceInfo gets information about a NearLink device.
func (c *NearLinkController) GetDeviceInfo(ctx context.Context, deviceMAC string) (map[string]interface{}, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("controller not connected")
	}

	// Send device info request
	message := buildNearLinkInfoMessage(deviceMAC)
	_, err := c.udpConn.Write(message)
	if err != nil {
		return nil, err
	}

	// In a real implementation, we would parse the device info response
	// For now, return placeholder info
	return map[string]interface{}{
		"mac_address":    deviceMAC,
		"rssi":           -60,
		"battery_level":  85,
		"firmware_version": "1.0.0",
		"protocol_version": "1.0",
	}, nil
}

// ResetNetwork resets the NearLink network.
func (c *NearLinkController) ResetNetwork(ctx context.Context) error {
	// Send network reset command
	message := buildNearLinkResetMessage()
	_, err := c.udpConn.Write(message)
	return err
}

// Helper functions

func parseNearLinkDeviceAnnouncement(data []byte, addr *net.UDPAddr) (*DeviceInfo, error) {
	// Parse device announcement message
	// This is a simplified implementation
	deviceID := fmt.Sprintf("nearlink-%s", addr.String())

	return &DeviceInfo{
		ID:       deviceID,
		Name:     fmt.Sprintf("NearLink Device %s", addr.String()),
		Type:     DeviceTypeSensor,
		Protocol: ProtocolNearLink,
		Status:   DeviceStatusOnline,
		Properties: map[string]interface{}{
			"mac_address": addr.String(),
			"ip_address":  addr.IP.String(),
		},
		LastSeen:    time.Now(),
		Capabilities: []DeviceCapability{CapabilitySensor},
	}, nil
}

func buildNearLinkDiscoveryMessage() []byte {
	// Build NearLink discovery broadcast message
	return []byte("NEARLINK_DISCOVERY")
}

func buildNearLinkPairingMessage(enable bool) []byte {
	if enable {
		return []byte("NEARLINK_PAIRING_ENABLE")
	}
	return []byte("NEARLINK_PAIRING_DISABLE")
}

func buildNearLinkUnpairMessage(deviceID string) []byte {
	return []byte(fmt.Sprintf("NEARLINK_UNPAIR:%s", deviceID))
}

func buildNearLinkResetMessage() []byte {
	return []byte("NEARLINK_RESET")
}

func buildNearLinkReadAttributeMessage(deviceMAC, attribute string) []byte {
	return []byte(fmt.Sprintf("NEARLINK_READ:%s:%s", deviceMAC, attribute))
}

func buildNearLinkWriteAttributeMessage(deviceMAC, attribute string, value interface{}) []byte {
	return []byte(fmt.Sprintf("NEARLINK_WRITE:%s:%s:%v", deviceMAC, attribute, value))
}

func buildNearLinkCommandMessage(deviceMAC, command string, params map[string]interface{}) []byte {
	return []byte(fmt.Sprintf("NEARLINK_CMD:%s:%s", deviceMAC, command))
}

func buildNearLinkInfoMessage(deviceMAC string) []byte {
	return []byte(fmt.Sprintf("NEARLINK_INFO:%s", deviceMAC))
}

func getFrequencyString(channel uint8) string {
	if channel < 14 {
		return "2.4GHz"
	}
	return "5.1GHz"
}

func isDevicePaired(deviceID string) bool {
	// Check if device is paired (implementation specific)
	return false
}

// getStringFromInterface safely extracts a string from a map[string]interface{}
