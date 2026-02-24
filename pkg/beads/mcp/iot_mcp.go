// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package mcp provides IoT protocol MCP tools.
package mcp

import (
	"context"
	"fmt"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/iot"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// IoTMCPTools provides IoT protocol control MCP tools.
type IoTMCPTools struct {
	agent *agent.HardwareAgent
}

// NewIoTMCPTools creates a new IoTMCPTools instance.
func NewIoTMCPTools(hardwareAgent *agent.HardwareAgent) *IoTMCPTools {
	return &IoTMCPTools{
		agent: hardwareAgent,
	}
}

// RegisterTools registers all IoT MCP tools with the MCP server.
func (t *IoTMCPTools) RegisterTools(s *server.MCPServer) {
	// ===== IoT Device Discovery and Pairing =====

	// DiscoverIoTDevices tool
	s.AddTool(mcp.Tool{
		Name:        "discover_iot_devices",
		Description: "Discover IoT devices on the specified protocol network",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "IoT protocol to use for discovery",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
				"timeout_seconds": map[string]interface{}{
					"type":        "number",
					"description": "Discovery timeout in seconds (default: 10)",
				},
			},
			Required: []string{"protocol"},
		},
	}, t.handleDiscoverDevices)

	// StartPairing tool
	s.AddTool(mcp.Tool{
		Name:        "start_iot_pairing",
		Description: "Start pairing mode for IoT devices (make devices discoverable)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "IoT protocol",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
				"timeout_seconds": map[string]interface{}{
					"type":        "number",
					"description": "Pairing timeout in seconds (default: 60)",
				},
			},
			Required: []string{"protocol"},
		},
	}, t.handleStartPairing)

	// CancelPairing tool
	s.AddTool(mcp.Tool{
		Name:        "cancel_iot_pairing",
		Description: "Cancel ongoing IoT device pairing",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "IoT protocol",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
			},
			Required: []string{"protocol"},
		},
	}, t.handleCancelPairing)

	// ===== IoT Device Management =====

	// ListIoTDevices tool
	s.AddTool(mcp.Tool{
		Name:        "list_iot_devices",
		Description: "List all discovered/paired IoT devices",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "Filter by protocol (optional)",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
			},
		},
	}, t.handleListDevices)

	// GetIoTDeviceInfo tool
	s.AddTool(mcp.Tool{
		Name:        "get_iot_device_info",
		Description: "Get detailed information about a specific IoT device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleGetDeviceInfo)

	// RemoveIoTDevice tool
	s.AddTool(mcp.Tool{
		Name:        "remove_iot_device",
		Description: "Remove/unpair an IoT device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleRemoveDevice)

	// ===== IoT Device Control =====

	// ReadIoTAttribute tool
	s.AddTool(mcp.Tool{
		Name:        "read_iot_attribute",
		Description: "Read an attribute from an IoT device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"attribute": map[string]interface{}{
					"type":        "string",
					"description": "Attribute name to read (e.g., state, temperature, brightness)",
				},
			},
			Required: []string{"device_id", "attribute"},
		},
	}, t.handleReadAttribute)

	// WriteIoTAttribute tool
	s.AddTool(mcp.Tool{
		Name:        "write_iot_attribute",
		Description: "Write an attribute to an IoT device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"attribute": map[string]interface{}{
					"type":        "string",
					"description": "Attribute name to write",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Value to write",
				},
			},
			Required: []string{"device_id", "attribute", "value"},
		},
	}, t.handleWriteAttribute)

	// BatchReadIoTAttributes tool
	s.AddTool(mcp.Tool{
		Name:        "batch_read_iot_attributes",
		Description: "Read multiple attributes from an IoT device at once",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"attributes": map[string]interface{}{
					"type":        "array",
					"description": "List of attribute names to read",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			Required: []string{"device_id", "attributes"},
		},
	}, t.handleBatchReadAttributes)

	// BatchWriteIoTAttributes tool
	s.AddTool(mcp.Tool{
		Name:        "batch_write_iot_attributes",
		Description: "Write multiple attributes to an IoT device at once",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"values": map[string]interface{}{
					"type":        "object",
					"description": "Map of attribute names to values",
				},
			},
			Required: []string{"device_id", "values"},
		},
	}, t.handleBatchWriteAttributes)

	// ===== IoT Device Quick Actions =====

	// SetIoTOnOff tool
	s.AddTool(mcp.Tool{
		Name:        "set_iot_on_off",
		Description: "Turn an IoT device on or off",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"state": map[string]interface{}{
					"type":        "string",
					"description": "Device state",
					"enum":        []string{"on", "off", "toggle"},
				},
			},
			Required: []string{"device_id", "state"},
		},
	}, t.handleSetOnOff)

	// SetIoTLevel tool
	s.AddTool(mcp.Tool{
		Name:        "set_iot_level",
		Description: "Set the level of an IoT device (e.g., brightness, speed)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"level": map[string]interface{}{
					"type":        "number",
					"description": "Level value (0-100)",
					"minimum":     0,
					"maximum":     100,
				},
			},
			Required: []string{"device_id", "level"},
		},
	}, t.handleSetLevel)

	// SetIoTColor tool
	s.AddTool(mcp.Tool{
		Name:        "set_iot_color",
		Description: "Set the color of an IoT device (for RGB lights)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
				"color": map[string]interface{}{
					"type":        "string",
					"description": "Color in hex format (e.g., #FF0000 for red)",
				},
			},
			Required: []string{"device_id", "color"},
		},
	}, t.handleSetColor)

	// ===== IoT Network Management =====

	// GetIoTNetworkInfo tool
	s.AddTool(mcp.Tool{
		Name:        "get_iot_network_info",
		Description: "Get information about the IoT network",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "IoT protocol",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
			},
			Required: []string{"protocol"},
		},
	}, t.handleGetNetworkInfo)

	// ResetIoTNetwork tool
	s.AddTool(mcp.Tool{
		Name:        "reset_iot_network",
		Description: "Reset the IoT network (use with caution)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"protocol": map[string]interface{}{
					"type":        "string",
					"description": "IoT protocol",
					"enum":        []string{"zigbee", "zwave", "thread", "nearlink"},
				},
			},
			Required: []string{"protocol"},
		},
	}, t.handleResetNetwork)

	// ===== IoT Device Diagnostics =====

	// PingIoTDevice tool
	s.AddTool(mcp.Tool{
		Name:        "ping_iot_device",
		Description: "Ping an IoT device to check connectivity",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handlePingDevice)

	// GetIoTDeviceDiagnostics tool
	s.AddTool(mcp.Tool{
		Name:        "get_iot_device_diagnostics",
		Description: "Get diagnostic information for an IoT device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "IoT device identifier",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleGetDiagnostics)

	// ===== Protocol-Specific Tools =====

	// Z-Wave: Heal Network
	s.AddTool(mcp.Tool{
		Name:        "zwave_heal_network",
		Description: "Heal the Z-Wave network (optimize routes)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleZWaveHealNetwork)

	// Thread: Get Mesh Topology
	s.AddTool(mcp.Tool{
		Name:        "thread_get_mesh_topology",
		Description: "Get the Thread mesh network topology",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleThreadGetMeshTopology)

	// NearLink: Get Mode
	s.AddTool(mcp.Tool{
		Name:        "nearlink_get_mode",
		Description: "Get the current NearLink network mode (SLM or SLE)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleNearLinkGetMode)
}

// ===== Handler Implementations =====

func (t *IoTMCPTools) handleDiscoverDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol       string  `json:"protocol"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	timeout := 10 * time.Second
	if params.TimeoutSeconds > 0 {
		timeout = time.Duration(params.TimeoutSeconds) * time.Second
	}

	// Get adapter for protocol
	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	// Discover devices
	devices, err := adapter.DiscoverDevices(ctx, timeout)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Discovery failed: %v", err)), nil
	}

	// Format results
	result := map[string]interface{}{
		"protocol":     params.Protocol,
		"device_count": len(devices),
		"devices":      devices,
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleStartPairing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol       string  `json:"protocol"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	timeout := 60 * time.Second
	if params.TimeoutSeconds > 0 {
		timeout = time.Duration(params.TimeoutSeconds) * time.Second
	}

	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	result, err := adapter.StartPairing(ctx, timeout)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Pairing failed: %v", err)), nil
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleCancelPairing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol string `json:"protocol"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	err = adapter.CancelPairing(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Cancel pairing failed: %v", err)), nil
	}

	return mcp.NewToolResultText("Pairing canceled successfully"), nil
}

func (t *IoTMCPTools) handleListDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol string `json:"protocol"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	devices, err := adapter.ListDevices(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list devices: %v", err)), nil
	}

	// Convert to device info
	deviceInfos := make([]*iot.DeviceInfo, 0, len(devices))
	for _, device := range devices {
		info := &iot.DeviceInfo{
			ID:           device.ID(),
			Name:         device.Name(),
			Type:         device.Type(),
			Protocol:     device.Protocol(),
			Manufacturer: device.Manufacturer(),
			Model:        device.Model(),
			Version:      device.Version(),
			Status:       device.Status(),
			Capabilities: device.Capabilities(),
		}
		deviceInfos = append(deviceInfos, info)
	}

	result := map[string]interface{}{
		"protocol":     params.Protocol,
		"device_count": len(deviceInfos),
		"devices":      deviceInfos,
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleGetDeviceInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	info := &iot.DeviceInfo{
		ID:           device.ID(),
		Name:         device.Name(),
		Type:         device.Type(),
		Protocol:     device.Protocol(),
		Manufacturer: device.Manufacturer(),
		Model:        device.Model(),
		Version:      device.Version(),
		Status:       device.Status(),
		Capabilities: device.Capabilities(),
	}
	return mcp.NewToolResultJSON(info)
}

func (t *IoTMCPTools) handleRemoveDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	protocol := iot.GetProtocolFromDeviceID(params.DeviceID)
	adapter, err := t.getAdapter(string(protocol))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	err = adapter.RemoveDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove device: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s removed successfully", params.DeviceID)), nil
}

func (t *IoTMCPTools) handleReadAttribute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID  string `json:"device_id"`
		Attribute string `json:"attribute"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	value, err := device.Read(ctx, params.Attribute)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read attribute: %v", err)), nil
	}

	result := map[string]interface{}{
		"device_id": params.DeviceID,
		"attribute": params.Attribute,
		"value":     value,
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleWriteAttribute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID  string `json:"device_id"`
		Attribute string `json:"attribute"`
		Value     string `json:"value"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	err = device.Write(ctx, params.Attribute, params.Value)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write attribute: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Attribute %s set to %s on device %s", params.Attribute, params.Value, params.DeviceID)), nil
}

func (t *IoTMCPTools) handleBatchReadAttributes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID   string   `json:"device_id"`
		Attributes []string `json:"attributes"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	// Check if device supports batch operations
	if batchDevice, ok := device.(iot.BatchReader); ok {
		values, err := batchDevice.BatchRead(ctx, params.Attributes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Batch read failed: %v", err)), nil
		}

		result := map[string]interface{}{
			"device_id": params.DeviceID,
			"values":    values,
		}

		return mcp.NewToolResultJSON(result)
	}

	// Fallback: read individually
	values := make(map[string]interface{})
	for _, attr := range params.Attributes {
		value, err := device.Read(ctx, attr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read %s: %v", attr, err)), nil
		}
		values[attr] = value
	}

	result := map[string]interface{}{
		"device_id": params.DeviceID,
		"values":    values,
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleBatchWriteAttributes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string                 `json:"device_id"`
		Values   map[string]interface{} `json:"values"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	// Check if device supports batch operations
	if batchDevice, ok := device.(iot.BatchWriter); ok {
		err := batchDevice.BatchWrite(ctx, params.Values)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Batch write failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote %d attributes to device %s", len(params.Values), params.DeviceID)), nil
	}

	// Fallback: write individually
	for attr, value := range params.Values {
		err := device.Write(ctx, attr, value)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write %s: %v", attr, err)), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote %d attributes to device %s", len(params.Values), params.DeviceID)), nil
}

func (t *IoTMCPTools) handleSetOnOff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
		State    string `json:"state"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	// Handle toggle
	if params.State == "toggle" {
		if toggleDevice, ok := device.(iot.Toggleable); ok {
			err := toggleDevice.Toggle(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Toggle failed: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Device %s toggled", params.DeviceID)), nil
		}
		return mcp.NewToolResultError("Device does not support toggle"), nil
	}

	// Handle on/off
	err = device.Write(ctx, "state", params.State)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set state: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s turned %s", params.DeviceID, params.State)), nil
}

func (t *IoTMCPTools) handleSetLevel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string  `json:"device_id"`
		Level    float64 `json:"level"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	err = device.Write(ctx, "level", uint8(params.Level))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set level: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s level set to %d", params.DeviceID, uint8(params.Level))), nil
}

func (t *IoTMCPTools) handleSetColor(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
		Color    string `json:"color"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	err = device.Write(ctx, "color", params.Color)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set color: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s color set to %s", params.DeviceID, params.Color)), nil
}

func (t *IoTMCPTools) handleGetNetworkInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol string `json:"protocol"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	info, err := adapter.GetNetworkInfo(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get network info: %v", err)), nil
	}

	return mcp.NewToolResultJSON(info)
}

func (t *IoTMCPTools) handleResetNetwork(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Protocol string `json:"protocol"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	adapter, err := t.getAdapter(params.Protocol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get adapter: %v", err)), nil
	}

	err = adapter.ResetNetwork(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reset network: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("%s network reset successfully", params.Protocol)), nil
}

func (t *IoTMCPTools) handlePingDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	// Check if device supports ping
	if pingable, ok := device.(iot.Pingable); ok {
		rtt, err := pingable.Ping(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Ping failed: %v", err)), nil
		}

		result := map[string]interface{}{
			"device_id": params.DeviceID,
			"rtt_ms":    rtt.Milliseconds(),
			"status":    "reachable",
		}

		return mcp.NewToolResultJSON(result)
	}

	return mcp.NewToolResultError("Device does not support ping"), nil
}

func (t *IoTMCPTools) handleGetDiagnostics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		DeviceID string `json:"device_id"`
	}
	if err := unmarshalArgs(request.Params.Arguments, &params); err != nil {
		return nil, err
	}

	device, err := t.getDevice(ctx, params.DeviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device: %v", err)), nil
	}

	// Check if device supports diagnostics
	if diagnosticable, ok := device.(iot.Diagnosticable); ok {
		info, err := diagnosticable.GetDiagnosticInfo(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get diagnostics: %v", err)), nil
		}

		return mcp.NewToolResultJSON(info)
	}

	// Fallback to basic info
	result := map[string]interface{}{
		"device_id": params.DeviceID,
		"status":    device.Status(),
	}

	return mcp.NewToolResultJSON(result)
}

// ===== Protocol-Specific Handlers =====

func (t *IoTMCPTools) handleZWaveHealNetwork(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get Z-Wave adapter
	adapter, err := t.getAdapter("zwave")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Z-Wave adapter: %v", err)), nil
	}

	// Type assert to ZWaveAdapter to access HealNetwork
	if zwaveAdapter, ok := adapter.(*iot.ZWaveAdapter); ok {
		jsClient := zwaveAdapter.GetJSClient()
		if jsClient == nil {
			return mcp.NewToolResultError("Z-Wave JS client not available"), nil
		}

		err := jsClient.HealNetwork(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Network heal failed: %v", err)), nil
		}

		return mcp.NewToolResultText("Z-Wave network heal started successfully"), nil
	}

	return mcp.NewToolResultError("Invalid Z-Wave adapter"), nil
}

func (t *IoTMCPTools) handleThreadGetMeshTopology(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get Thread adapter
	adapter, err := t.getAdapter("thread")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get Thread adapter: %v", err)), nil
	}

	// Get network info
	info, err := adapter.GetNetworkInfo(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get network info: %v", err)), nil
	}

	result := map[string]interface{}{
		"network_info": info,
		"topology":     "Mesh topology information would be retrieved from the Thread border router",
	}

	return mcp.NewToolResultJSON(result)
}

func (t *IoTMCPTools) handleNearLinkGetMode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get NearLink adapter
	adapter, err := t.getAdapter("nearlink")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get NearLink adapter: %v", err)), nil
	}

	// Get network info
	info, err := adapter.GetNetworkInfo(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get network info: %v", err)), nil
	}

	mode := info.Properties["network_mode"]

	result := map[string]interface{}{
		"mode":       mode,
		"description": "SLM = Spark Link Mesh (high performance), SLE = Spark Link Low Energy (ultra-low power)",
	}

	return mcp.NewToolResultJSON(result)
}

// ===== Helper Functions =====

func (t *IoTMCPTools) getAdapter(protocol string) (iot.ProtocolAdapter, error) {
	// TODO: Get adapter from HardwareAgent's IoTDeviceManager
	// For now, return a placeholder
	return nil, fmt.Errorf("adapter not implemented yet: %s", protocol)
}

func (t *IoTMCPTools) getDevice(ctx context.Context, deviceID string) (iot.IoTDevice, error) {
	// Extract protocol from device ID
	protocol := iot.GetProtocolFromDeviceID(deviceID)

	adapter, err := t.getAdapter(string(protocol))
	if err != nil {
		return nil, err
	}

	return adapter.GetDevice(ctx, deviceID)
}
