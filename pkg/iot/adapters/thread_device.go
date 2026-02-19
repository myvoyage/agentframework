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
	"net"
	"net/url"
	"time"

	"AgentFramework/pkg/iot"
)

// ThreadDevice implements IoTDevice for Thread devices.
type ThreadDevice struct {
	*iot.BaseDevice
	deviceID string
	IPv6     string
	adapter  *ThreadAdapter
}

// NewThreadDevice creates a new Thread device instance.
func NewThreadDevice(info *iot.DeviceInfo, adapter *ThreadAdapter) *ThreadDevice {
	return &ThreadDevice{
		BaseDevice: iot.NewBaseDevice(info),
		deviceID:   info.ID,
		IPv6:       getStringFromInterface(info.Properties, "ipv6", ""),
		adapter:    adapter,
	}
}

// ID returns the device ID.
func (d *ThreadDevice) ID() string {
	return d.deviceID
}

// Connect connects to the Thread device.
func (d *ThreadDevice) Connect(ctx context.Context) error {
	// For Thread devices, connectivity is handled by the network
	// This method marks the device as connected
	// Call BaseDevice's internal method via reflection or just track state
	return nil
}

// Disconnect disconnects from the Thread device.
func (d *ThreadDevice) Disconnect(ctx context.Context) error {
	// Mark device as disconnected
	return nil
}

// Read reads an attribute from the device.
func (d *ThreadDevice) Read(ctx context.Context, attribute string) (interface{}, error) {
	if !d.IsConnected() {
		return nil, iot.ErrNetworkError
	}

	// Use CoAP GET to read device attribute
	url := fmt.Sprintf("coap://%s/%s", d.IPv6, attribute)
	response, err := d.coapGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("CoAP GET failed: %w", err)
	}

	return response, nil
}

// Write writes an attribute to the device.
func (d *ThreadDevice) Write(ctx context.Context, attribute string, value interface{}) error {
	if !d.IsConnected() {
		return iot.ErrNetworkError
	}

	// Use CoAP PUT to write device attribute
	url := fmt.Sprintf("coap://%s/%s", d.IPv6, attribute)
	if err := d.coapPut(ctx, url, value); err != nil {
		return fmt.Errorf("CoAP PUT failed: %w", err)
	}

	return nil
}

// Subscribe subscribes to device events.
func (d *ThreadDevice) Subscribe(ctx context.Context, events []string, handler iot.DeviceEventHandler) error {
	// Use BaseDevice's Subscribe method
	return d.BaseDevice.Subscribe(ctx, events, handler)
}

// Unsubscribe unsubscribes from device events.
func (d *ThreadDevice) Unsubscribe(ctx context.Context, events []string) error {
	// Use BaseDevice's Unsubscribe method
	return d.BaseDevice.Unsubscribe(ctx, events)
}

// GetCapabilities returns device capabilities.
func (d *ThreadDevice) GetCapabilities() []iot.DeviceCapability {
	return d.Capabilities()
}

// GetProperty returns a device property.
func (d *ThreadDevice) GetProperty(ctx context.Context, property string) (interface{}, error) {
	return d.Read(ctx, property)
}

// SetProperty sets a device property.
func (d *ThreadDevice) SetProperty(ctx context.Context, property string, value interface{}) error {
	return d.Write(ctx, property, value)
}

// GetInfo returns device information.
func (d *ThreadDevice) GetInfo() *iot.DeviceInfo {
	return &iot.DeviceInfo{
		ID:           d.ID(),
		Name:         d.Name(),
		Type:         d.Type(),
		Protocol:     d.Protocol(),
		Manufacturer: d.Manufacturer(),
		Model:        d.Model(),
		Version:      d.Version(),
		Status:       d.Status(),
		Capabilities: d.Capabilities(),
		Properties: map[string]interface{}{
			"ipv6": d.IPv6,
		},
		LastSeen: d.LastSeen(),
	}
}

// RefreshState refreshes the device state from the actual device.
func (d *ThreadDevice) RefreshState(ctx context.Context) error {
	// Read current state from device
	state, err := d.Read(ctx, "state")
	if err != nil {
		return err
	}

	// Update device config (using exported method)
	config, _ := d.GetConfig(ctx)
	if config == nil {
		config = make(map[string]interface{})
	}
	config["state"] = state
	_ = d.SetConfig(ctx, config)

	return nil
}

// coapGet performs a CoAP GET request.
func (d *ThreadDevice) coapGet(ctx context.Context, urlStr string) (interface{}, error) {
	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Resolve IPv6 address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", net.ParseIP(u.Host), 5683))
	if err != nil {
		return nil, err
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set deadline
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetWriteDeadline(deadline)
		conn.SetReadDeadline(deadline)
	}

	// Send CoAP GET request
	// TODO: Implement actual CoAP GET message format
	// For now, this is a placeholder

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response interface{}
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		return nil, err
	}

	return response, nil
}

// coapPut performs a CoAP PUT request.
func (d *ThreadDevice) coapPut(ctx context.Context, urlStr string, value interface{}) error {
	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Resolve IPv6 address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", net.ParseIP(u.Host), 5683))
	if err != nil {
		return err
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set deadline
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetWriteDeadline(deadline)
	}

	// Marshal value to JSON
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Send CoAP PUT request
	// TODO: Implement actual CoAP PUT message format
	// For now, this is a placeholder
	_, err = conn.Write(payload)
	return err
}

// handleStateUpdate handles state update from CoAP message.
func (d *ThreadDevice) handleStateUpdate(payload map[string]interface{}) {
	// Update device config
	config, _ := d.GetConfig(context.Background())
	if config == nil {
		config = make(map[string]interface{})
	}
	for key, value := range payload {
		config[key] = value
	}
	_ = d.SetConfig(context.Background(), config)
}

// handleEvent handles event from CoAP message.
func (d *ThreadDevice) handleEvent(payload map[string]interface{}) {
	// Events are handled through the Subscribe mechanism
	// This method can be used for logging or other side effects
	// Actual event delivery is handled by BaseDevice
	_ = payload // Avoid unused parameter warning
}

// Toggle toggles the device state (if supported).
func (d *ThreadDevice) Toggle(ctx context.Context) error {
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

// Close closes the device and releases resources.
func (d *ThreadDevice) Close(ctx context.Context) error {
	d.Disconnect(ctx)
	return nil
}

// Stream opens a data stream from the device (for sensors, etc.).
func (d *ThreadDevice) Stream(ctx context.Context, property string, interval time.Duration) (<-chan interface{}, error) {
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

// BatchRead reads multiple attributes at once.
func (d *ThreadDevice) BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error) {
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
func (d *ThreadDevice) BatchWrite(ctx context.Context, values map[string]interface{}) error {
	for attr, value := range values {
		if err := d.Write(ctx, attr, value); err != nil {
			return fmt.Errorf("failed to write %s: %w", attr, err)
		}
	}

	return nil
}

// GetDiagnosticInfo returns diagnostic information for the device.
func (d *ThreadDevice) GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Basic info
	info["id"] = d.ID()
	info["ipv6"] = d.IPv6
	info["status"] = d.Status()
	info["last_seen"] = d.LastSeen()

	// Connection test
	if err := d.testConnection(ctx); err != nil {
		info["connection_error"] = err.Error()
	} else {
		info["connection"] = "ok"
	}

	// Network info
	info["rssi"] = d.getRSSI(ctx)
	info["packet_count"] = d.getPacketCount(ctx)

	return info, nil
}

// testConnection tests connectivity to the device.
func (d *ThreadDevice) testConnection(ctx context.Context) error {
	// Simple ping test via CoAP
	_, err := d.coapGet(ctx, fmt.Sprintf("coap://%s/ping", d.IPv6))
	return err
}

// getRSSI gets the signal strength.
func (d *ThreadDevice) getRSSI(ctx context.Context) int {
	// TODO: Implement actual RSSI retrieval
	return -50 // Placeholder
}

// getPacketCount gets the packet count.
func (d *ThreadDevice) getPacketCount(ctx context.Context) uint64 {
	// TODO: Implement actual packet count retrieval
	return 0 // Placeholder
}

// SubscribeToChanges subscribes to device state changes.
func (d *ThreadDevice) SubscribeToChanges(ctx context.Context, handler func(changes map[string]interface{})) func() {
	// Use CoAP observe to monitor changes
	// This is a simplified implementation

	cancelChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		lastState := make(map[string]interface{})

		for {
			select {
			case <-ticker.C:
				currentState, err := d.GetConfig(ctx)
				if err != nil {
					continue
				}

				// Detect changes
				changes := make(map[string]interface{})
				for key, value := range currentState {
					if lastValue, exists := lastState[key]; !exists || lastValue != value {
						changes[key] = value
					}
				}

				if len(changes) > 0 {
					handler(changes)
				}

				lastState = currentState

			case <-cancelChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() {
		close(cancelChan)
	}
}

// UpdateFirmware updates the device firmware.
func (d *ThreadDevice) UpdateFirmware(ctx context.Context, firmwareURL string, progressCallback func(progress int)) error {
	// TODO: Implement OTA firmware update
	// This would involve:
	// 1. Download firmware
	// 2. Verify firmware signature
	// 3. Send firmware to device via CoAP
	// 4. Monitor update progress
	// 5. Verify update success

	return fmt.Errorf("firmware update not implemented")
}

// Reset resets the device to factory settings.
func (d *ThreadDevice) Reset(ctx context.Context) error {
	// TODO: Implement device reset
	return fmt.Errorf("device reset not implemented")
}

// GetLogs retrieves device logs.
func (d *ThreadDevice) GetLogs(ctx context.Context, lines int) ([]string, error) {
	// TODO: Implement log retrieval
	return nil, fmt.Errorf("log retrieval not implemented")
}

// Ping sends a ping to the device and returns the round-trip time.
func (d *ThreadDevice) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	if err := d.testConnection(ctx); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}
