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

// ThreadAdapter implements ProtocolAdapter for Thread networks.
type ThreadAdapter struct {
	*iot.BaseAdapter
	borderRouter *ThreadBorderRouter
	coapServer   *CoAPServer
	devices      map[string]*ThreadDevice
	storage      iot.DeviceRegistry
	network      *ThreadNetworkConfig
	eventHandler iot.ProtocolEventHandler
}

// NewThreadAdapter creates a new Thread protocol adapter.
func NewThreadAdapter() *ThreadAdapter {
	return &ThreadAdapter{
		BaseAdapter: iot.NewBaseAdapter(iot.ProtocolThread, "1.0.0"),
		devices:     make(map[string]*ThreadDevice),
	}
}

// Initialize initializes the Thread adapter with the given configuration.
func (a *ThreadAdapter) Initialize(ctx context.Context, config iot.ProtocolConfig) error {
	a.network = &ThreadNetworkConfig{
		NetworkName:      config.Metadata["network_name"],
		PanID:           parseUint16(config.Metadata["pan_id"], 0x1234),
		Channel:         parseUint8(config.Metadata["channel"], 15),
		MeshLocalPrefix: config.Metadata["mesh_local_prefix"],
		OnMeshPrefix:    config.Metadata["on_mesh_prefix"],
	}

	// Initialize border router
	borderRouterConfig := &BorderRouterConfig{
		Interface: config.Metadata["interface"],
		Port:      parseInt(config.Metadata["border_router_port"], 8080),
	}

	var err error
	a.borderRouter, err = NewThreadBorderRouter(borderRouterConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize border router: %w", err)
	}

	// Initialize CoAP server
	coapConfig := &CoAPServerConfig{
		Port: parseInt(config.Metadata["coap_port"], 5683),
	}

	a.coapServer, err = NewCoAPServer(coapConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize CoAP server: %w", err)
	}

	// Initialize device registry
	a.storage = *iot.NewDeviceRegistry()

	return nil
}

// Start starts the Thread adapter.
func (a *ThreadAdapter) Start(ctx context.Context) error {
	// Start border router
	if err := a.borderRouter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start border router: %w", err)
	}

	// Start CoAP server
	if err := a.coapServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start CoAP server: %w", err)
	}

	// Set up CoAP message handler
	a.coapServer.SetMessageHandler(a.handleCoAPMessage)

	// Start network discovery
	go a.discoverDevicesLoop(ctx)

	// Mark adapter as running
	a.BaseAdapter.Start(ctx)

	return nil
}

// Stop stops the Thread adapter.
func (a *ThreadAdapter) Stop(ctx context.Context) error {
	// Stop CoAP server
	if a.coapServer != nil {
		a.coapServer.Stop(ctx)
	}

	// Stop border router
	if a.borderRouter != nil {
		a.borderRouter.Stop(ctx)
	}

	// Mark adapter as stopped
	a.BaseAdapter.Stop(ctx)

	return nil
}

// DiscoverDevices discovers Thread devices on the network.
func (a *ThreadAdapter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*iot.DeviceInfo, error) {
	if !a.IsRunning() {
		return nil, iot.ErrAdapterNotRunning
	}

	// Use CoAP multicast to discover devices
	discovered, err := a.borderRouter.DiscoverDevices(ctx, timeout)
	if err != nil {
		return nil, fmt.Errorf("device discovery failed: %w", err)
	}

	devices := make([]*iot.DeviceInfo, 0, len(discovered))
	for _, deviceInfo := range discovered {
		// Create ThreadDevice for discovered device
		device := a.createThreadDevice(deviceInfo)
		a.devices[device.ID()] = device

		// Register device in storage
		a.storage.Register(deviceInfo)

		// Publish discovery event
		a.PublishEvent(iot.ProtocolEvent{
			Type:      iot.EventDeviceDiscovered,
			Protocol:  iot.ProtocolThread,
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

// StartPairing starts device commissioning (pairing) for Thread devices.
func (a *ThreadAdapter) StartPairing(ctx context.Context, timeout time.Duration) (*iot.PairingResult, error) {
	if !a.IsRunning() {
		return nil, iot.ErrAdapterNotRunning
	}

	// For Thread, commissioning involves:
	// 1. Enable commissioner role
	// 2. Generate commissioning credentials
	// 3. Wait for joiner requests
	// 4. Authorize and provision joiners

	if err := a.borderRouter.EnableCommissioning(ctx, timeout); err != nil {
		return &iot.PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Wait for device to join
	deviceChan := make(chan *iot.DeviceInfo, 1)
	errorChan := make(chan error, 1)

	go func() {
		device, err := a.waitForJoiner(ctx, timeout)
		if err != nil {
			errorChan <- err
			return
		}
		deviceChan <- device
	}()

	select {
	case device := <-deviceChan:
		// Create device instance
		threadDevice := a.createThreadDevice(device)
		a.devices[threadDevice.ID()] = threadDevice
		a.storage.Register(device)

		return &iot.PairingResult{
			Success: true,
			Device:  device,
		}, nil

	case err := <-errorChan:
		return &iot.PairingResult{
			Success: false,
			Error:   err.Error(),
		}, nil

	case <-ctx.Done():
		return &iot.PairingResult{
			Success: false,
			Error:   "pairing timeout",
		}, nil
	}
}

// CancelPairing cancels ongoing device pairing.
func (a *ThreadAdapter) CancelPairing(ctx context.Context) error {
	return a.borderRouter.DisableCommissioning(ctx)
}

// GetDevice retrieves a Thread device by ID.
func (a *ThreadAdapter) GetDevice(ctx context.Context, deviceID string) (iot.IoTDevice, error) {
	for _, device := range a.devices {
		if device.ID() == deviceID {
			return device, nil
		}
	}
	return nil, iot.ErrDeviceNotFound
}

// ListDevices lists all Thread devices.
func (a *ThreadAdapter) ListDevices(ctx context.Context) ([]iot.IoTDevice, error) {
	devices := make([]iot.IoTDevice, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// RemoveDevice removes a Thread device.
func (a *ThreadAdapter) RemoveDevice(ctx context.Context, deviceID string) error {
	for id, device := range a.devices {
		if device.ID() == deviceID {
			delete(a.devices, id)
			a.storage.Unregister(deviceID)

			a.PublishEvent(iot.ProtocolEvent{
				Type:      iot.EventDeviceLeft,
				Protocol:  iot.ProtocolThread,
				DeviceID:  deviceID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"device_id": deviceID,
				},
			})
			return nil
		}
	}
	return iot.ErrDeviceNotFound
}

// GetNetworkInfo retrieves Thread network information.
func (a *ThreadAdapter) GetNetworkInfo(ctx context.Context) (*iot.NetworkInfo, error) {
	info, err := a.borderRouter.GetNetworkInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &iot.NetworkInfo{
		PanID:        a.network.PanID,
		Channel:      a.network.Channel,
		DeviceCount:  len(a.devices),
		Status:       "running",
		Properties: map[string]string{
			"network_name":       a.network.NetworkName,
			"mesh_local_prefix":  a.network.MeshLocalPrefix,
			"on_mesh_prefix":     a.network.OnMeshPrefix,
			"border_router_addr": info.BorderRouterAddr,
			"partition_id":       fmt.Sprintf("%d", info.PartitionID),
		},
	}, nil
}

// ResetNetwork resets the Thread network (clears all devices).
func (a *ThreadAdapter) ResetNetwork(ctx context.Context) error {
	// Remove all devices
	for deviceID := range a.devices {
		a.RemoveDevice(ctx, deviceID)
	}

	// Reset border router
	if err := a.borderRouter.ResetNetwork(ctx); err != nil {
		return err
	}

	return nil
}

// SetEventHandler sets the event handler for protocol events.
func (a *ThreadAdapter) SetEventHandler(handler iot.ProtocolEventHandler) {
	a.eventHandler = handler
}

// PublishEvent publishes an event to the event handler.
func (a *ThreadAdapter) PublishEvent(event iot.ProtocolEvent) {
	if a.eventHandler != nil {
		a.eventHandler(event)
	}
}

// discoverDevicesLoop continuously discovers devices in the background.
func (a *ThreadAdapter) discoverDevicesLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.DiscoverDevices(ctx, 30*time.Second)
		case <-ctx.Done():
			return
		}
	}
}

// waitForJoiner waits for a Thread device to join the network.
func (a *ThreadAdapter) waitForJoiner(ctx context.Context, timeout time.Duration) (*iot.DeviceInfo, error) {
	joinerChan, err := a.borderRouter.WaitForJoiner(ctx, timeout)
	if err != nil {
		return nil, err
	}

	select {
	case joiner := <-joinerChan:
		return &iot.DeviceInfo{
			ID:           joiner.JoinerID,
			Name:         fmt.Sprintf("Thread-%s", joiner.JoinerID[:8]),
			Type:         iot.DeviceTypeSensor,
			Protocol:     iot.ProtocolThread,
			Manufacturer: "Unknown",
			Model:        "Thread Device",
			Version:      "1.0",
			Status:       iot.DeviceStatusOnline,
			Capabilities: []iot.DeviceCapability{
				iot.CapabilitySensor,
			},
			Properties: map[string]interface{}{
				"ipv6":          joiner.IPv6Addr,
				"joiner_rloc":   joiner.RLOC16,
				"partition_id":  joiner.PartitionID,
			},
			LastSeen: time.Now(),
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func getStringFromInterface(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

// createThreadDevice creates a ThreadDevice from device info.
func (a *ThreadAdapter) createThreadDevice(info *iot.DeviceInfo) *ThreadDevice {
	return &ThreadDevice{
		BaseDevice: iot.NewBaseDevice(info),
		deviceID:   info.ID,
		IPv6:       getStringFromInterface(info.Properties, "ipv6", ""),
		adapter:    a,
	}
}

// handleCoAPMessage handles incoming CoAP messages from Thread devices.
func (a *ThreadAdapter) handleCoAPMessage(msg *CoAPMessage) {
	// Find device by IPv6 address
	var device *ThreadDevice
	for _, d := range a.devices {
		if d.IPv6 == msg.SourceAddr {
			device = d
			break
		}
	}

	if device == nil {
		// Unknown device, might need to register it
		return
	}

	// Process CoAP message
	switch msg.Code {
	case CoAPPut:
		// Device state update
		device.handleStateUpdate(msg.Payload)
	case CoAPPost:
		// Device event
		device.handleEvent(msg.Payload)
	}

	// Publish event
	a.PublishEvent(iot.ProtocolEvent{
		Type:      iot.EventDataReceived,
		Protocol:  iot.ProtocolThread,
		DeviceID:  device.ID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"coap_code": msg.Code,
			"payload":   msg.Payload,
		},
	})
}

// Helper functions

func getString(m map[string]string, key, defaultVal string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultVal
}

func getInt(m map[string]string, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		var num int
		fmt.Sscanf(val, "%d", &num)
		return num
	}
	return defaultVal
}

func parseInt(val string, defaultVal int) int {
	if val == "" {
		return defaultVal
	}
	var num int
	fmt.Sscanf(val, "%d", &num)
	return num
}

func parseUint8(val string, defaultVal uint8) uint8 {
	if val == "" {
		return defaultVal
	}
	var num uint8
	fmt.Sscanf(val, "%d", &num)
	return num
}

func parseUint16(val string, defaultVal uint16) uint16 {
	if val == "" {
		return defaultVal
	}
	var num uint16
	fmt.Sscanf(val, "%d", &num)
	return num
}

func getUint8(m map[string]interface{}, key string, defaultVal uint8) uint8 {
	if val, ok := m[key]; ok {
		if num, ok := val.(int); ok {
			return uint8(num)
		}
		if num, ok := val.(float64); ok {
			return uint8(num)
		}
	}
	return defaultVal
}

func getUint16(m map[string]interface{}, key string, defaultVal uint16) uint16 {
	if val, ok := m[key]; ok {
		if num, ok := val.(int); ok {
			return uint16(num)
		}
		if num, ok := val.(float64); ok {
			return uint16(num)
		}
	}
	return defaultVal
}

// ThreadNetworkConfig contains Thread network configuration.
type ThreadNetworkConfig struct {
	NetworkName      string
	PanID           uint16
	Channel         uint8
	MeshLocalPrefix string
	OnMeshPrefix    string
}

// BorderRouterConfig contains border router configuration.
type BorderRouterConfig struct {
	Interface string
	Port      int
}

// CoAPServerConfig contains CoAP server configuration.
type CoAPServerConfig struct {
	Port int
}

// ThreadBorderRouter manages Thread border router operations.
type ThreadBorderRouter struct {
	config     *BorderRouterConfig
	network    *ThreadNetworkConfig
	isRunning  bool
}

// NewThreadBorderRouter creates a new Thread border router instance.
func NewThreadBorderRouter(config *BorderRouterConfig) (*ThreadBorderRouter, error) {
	return &ThreadBorderRouter{
		config: config,
	}, nil
}

// Start starts the border router.
func (br *ThreadBorderRouter) Start(ctx context.Context) error {
	// TODO: Implement actual border router startup
	// This would involve:
	// 1. Configure wpantund or similar
	// 2. Start OpenThread border router
	// 3. Set up network interface
	br.isRunning = true
	return nil
}

// Stop stops the border router.
func (br *ThreadBorderRouter) Stop(ctx context.Context) error {
	// TODO: Implement actual border router shutdown
	br.isRunning = false
	return nil
}

// DiscoverDevices discovers Thread devices using CoAP multicast.
func (br *ThreadBorderRouter) DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*iot.DeviceInfo, error) {
	// TODO: Implement actual device discovery
	// This would involve:
	// 1. Send CoAP multicast discovery message
	// 2. Wait for responses
	// 3. Parse device information

	// Placeholder implementation
	return []*iot.DeviceInfo{}, nil
}

// EnableCommissioning enables Thread commissioner role.
func (br *ThreadBorderRouter) EnableCommissioning(ctx context.Context, timeout time.Duration) error {
	// TODO: Implement actual commissioning
	return nil
}

// DisableCommissioning disables Thread commissioner role.
func (br *ThreadBorderRouter) DisableCommissioning(ctx context.Context) error {
	// TODO: Implement actual commissioning disable
	return nil
}

// WaitForJoiner waits for a Thread joiner.
func (br *ThreadBorderRouter) WaitForJoiner(ctx context.Context, timeout time.Duration) (chan *ThreadJoinerInfo, error) {
	// TODO: Implement actual joiner waiting
	joinerChan := make(chan *ThreadJoinerInfo, 1)
	return joinerChan, nil
}

// GetNetworkInfo retrieves Thread network information.
func (br *ThreadBorderRouter) GetNetworkInfo(ctx context.Context) (*ThreadNetworkInfo, error) {
	// TODO: Implement actual network info retrieval
	return &ThreadNetworkInfo{
		BorderRouterAddr: "fd00:abcd::1",
		PartitionID:      0x12345678,
	}, nil
}

// ResetNetwork resets the Thread network.
func (br *ThreadBorderRouter) ResetNetwork(ctx context.Context) error {
	// TODO: Implement actual network reset
	return nil
}

// ThreadNetworkInfo contains Thread network information.
type ThreadNetworkInfo struct {
	BorderRouterAddr string
	PartitionID      uint32
}

// ThreadJoinerInfo contains information about a Thread joiner.
type ThreadJoinerInfo struct {
	JoinerID   string
	IPv6Addr   string
	RLOC16     uint16
	PartitionID uint32
}

// CoAPServer manages CoAP server for Thread device communication.
type CoAPServer struct {
	config    *CoAPServerConfig
	isRunning bool
	handler   CoAPMessageHandler
}

// CoAPMessageHandler handles CoAP messages.
type CoAPMessageHandler func(msg *CoAPMessage)

// NewCoAPServer creates a new CoAP server instance.
func NewCoAPServer(config *CoAPServerConfig) (*CoAPServer, error) {
	return &CoAPServer{
		config: config,
	}, nil
}

// Start starts the CoAP server.
func (s *CoAPServer) Start(ctx context.Context) error {
	// TODO: Implement actual CoAP server startup
	// This would involve:
	// 1. Listen on UDP port
	// 2. Handle CoAP messages
	// 3. Route messages to handler
	s.isRunning = true
	return nil
}

// Stop stops the CoAP server.
func (s *CoAPServer) Stop(ctx context.Context) error {
	// TODO: Implement actual CoAP server shutdown
	s.isRunning = false
	return nil
}

// SetMessageHandler sets the CoAP message handler.
func (s *CoAPServer) SetMessageHandler(handler CoAPMessageHandler) {
	s.handler = handler
}

// CoAPMessage represents a CoAP message.
type CoAPMessage struct {
	Code       CoAPCode
	SourceAddr string
	Payload    map[string]interface{}
}

// CoAPCode represents CoAP message codes.
type CoAPCode string

const (
	CoAPGet    CoAPCode = "GET"
	CoAPPost   CoAPCode = "POST"
	CoAPPut    CoAPCode = "PUT"
	CoAPDelete CoAPCode = "DELETE"
)
