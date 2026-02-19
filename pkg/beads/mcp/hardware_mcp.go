// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package mcp provides hardware control MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/beads/hardware"
	"AgentFramework/pkg/beads/hardware/drivers"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HardwareMCPTools provides hardware control MCP tools.
type HardwareMCPTools struct {
	agent *agent.HardwareAgent
}

// NewHardwareMCPTools creates a new HardwareMCPTools instance.
func NewHardwareMCPTools(hardwareAgent *agent.HardwareAgent) *HardwareMCPTools {
	return &HardwareMCPTools{
		agent: hardwareAgent,
	}
}

// RegisterTools registers all hardware MCP tools with the MCP server.
func (t *HardwareMCPTools) RegisterTools(s *server.MCPServer) {
	// ConnectDevice tool
	s.AddTool(mcp.Tool{
		Name:        "connect_hardware_device",
		Description: "Connect to a hardware device using the specified driver and configuration",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the device",
				},
				"driver_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of driver to use (e.g., serial, modbus)",
					"enum":        []string{"serial", "modbus"},
				},
				"config": map[string]interface{}{
					"type":        "object",
					"description": "Device configuration (format depends on driver type)",
				},
			},
			Required: []string{"device_id", "driver_type", "config"},
		},
	}, t.handleConnectDevice)

	// DisconnectDevice tool
	s.AddTool(mcp.Tool{
		Name:        "disconnect_hardware_device",
		Description: "Disconnect from a hardware device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the device to disconnect",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleDisconnectDevice)

	// SendCommand tool
	s.AddTool(mcp.Tool{
		Name:        "send_hardware_command",
		Description: "Send a command to a connected hardware device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the device",
				},
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Command to send to the device",
				},
				"params": map[string]interface{}{
					"type":        "object",
					"description": "Command parameters (format depends on command)",
				},
			},
			Required: []string{"device_id", "command"},
		},
	}, t.handleSendCommand)

	// ReceiveData tool
	s.AddTool(mcp.Tool{
		Name:        "receive_hardware_data",
		Description: "Receive data from a connected hardware device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the device",
				},
				"timeout_ms": map[string]interface{}{
					"type":        "number",
					"description": "Timeout in milliseconds (default: 5000)",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleReceiveData)

	// GetDeviceStatus tool
	s.AddTool(mcp.Tool{
		Name:        "get_hardware_device_status",
		Description: "Get the current status of a connected hardware device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the device",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleGetDeviceStatus)

	// ListDevices tool
	s.AddTool(mcp.Tool{
		Name:        "list_hardware_devices",
		Description: "List all connected hardware devices",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleListDevices)

	// ListDrivers tool
	s.AddTool(mcp.Tool{
		Name:        "list_hardware_drivers",
		Description: "List all available hardware drivers",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleListDrivers)

	// ExecuteCommandSequence tool
	s.AddTool(mcp.Tool{
		Name:        "execute_hardware_command_sequence",
		Description: "Execute a sequence of hardware commands with optional delays",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"sequence": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"device_id": map[string]interface{}{
								"type": "string",
							},
							"command": map[string]interface{}{
								"type": "string",
							},
							"params": map[string]interface{}{
								"type": "object",
							},
							"delay_ms": map[string]interface{}{
								"type": "number",
							},
						},
						"required": []string{"device_id", "command"},
					},
					"description": "Sequence of commands to execute",
				},
			},
			Required: []string{"sequence"},
		},
	}, t.handleExecuteCommandSequence)

	// ===== CAN Bus Tools =====

	// CANSendFrame tool
	s.AddTool(mcp.Tool{
		Name:        "can_send_frame",
		Description: "Send a CAN message frame to the bus",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "CAN device identifier",
				},
				"id": map[string]interface{}{
					"type":        "number",
					"description": "CAN frame ID (11-bit standard or 29-bit extended)",
				},
				"data": map[string]interface{}{
					"type":        "array",
					"description": "Frame data bytes (0-8 for standard CAN, 0-64 for CAN FD)",
					"items": map[string]interface{}{
						"type": "number",
					},
				},
				"is_extended": map[string]interface{}{
					"type":        "boolean",
					"description": "Use extended 29-bit frame ID",
				},
				"is_remote": map[string]interface{}{
					"type":        "boolean",
					"description": "Remote transmission request frame",
				},
			},
			Required: []string{"device_id", "id"},
		},
	}, t.handleCANSendFrame)

	// CANReceiveFrame tool
	s.AddTool(mcp.Tool{
		Name:        "can_receive_frame",
		Description: "Receive a CAN message frame from the bus",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "CAN device identifier",
				},
				"timeout_ms": map[string]interface{}{
					"type":        "number",
					"description": "Receive timeout in milliseconds (default: 5000)",
				},
			},
			Required: []string{"device_id"},
		},
	}, t.handleCANReceiveFrame)

	// CANSetFilter tool
	s.AddTool(mcp.Tool{
		Name:        "can_set_filter",
		Description: "Set a CAN message filter for reception",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "CAN device identifier",
				},
				"filter_id": map[string]interface{}{
					"type":        "number",
					"description": "Filter ID",
				},
				"mask": map[string]interface{}{
					"type":        "number",
					"description": "Filter mask",
				},
			},
			Required: []string{"device_id", "filter_id", "mask"},
		},
	}, t.handleCANSetFilter)

	// ===== GPIO Tools =====

	// GPIOReadPin tool
	s.AddTool(mcp.Tool{
		Name:        "gpio_read_pin",
		Description: "Read the value of a GPIO pin",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "GPIO device identifier",
				},
				"pin": map[string]interface{}{
					"type":        "number",
					"description": "GPIO pin number",
				},
			},
			Required: []string{"device_id", "pin"},
		},
	}, t.handleGPIOReadPin)

	// GPIOWritePin tool
	s.AddTool(mcp.Tool{
		Name:        "gpio_write_pin",
		Description: "Write a value to a GPIO pin",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "GPIO device identifier",
				},
				"pin": map[string]interface{}{
					"type":        "number",
					"description": "GPIO pin number",
				},
				"value": map[string]interface{}{
					"type":        "number",
					"description": "Pin value (0 or 1)",
				},
			},
			Required: []string{"device_id", "pin", "value"},
		},
	}, t.handleGPIOWritePin)

	// GPIOSetupPin tool
	s.AddTool(mcp.Tool{
		Name:        "gpio_setup_pin",
		Description: "Configure a GPIO pin direction and options",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "GPIO device identifier",
				},
				"pin": map[string]interface{}{
					"type":        "number",
					"description": "GPIO pin number",
				},
				"direction": map[string]interface{}{
					"type":        "string",
					"description": "Pin direction (in, out, high, low)",
					"enum":        []string{"in", "out", "high", "low"},
				},
				"active_low": map[string]interface{}{
					"type":        "boolean",
					"description": "Active low configuration",
				},
				"pull": map[string]interface{}{
					"type":        "string",
					"description": "Pull resistor (up, down, off)",
					"enum":        []string{"up", "down", "off"},
				},
				"edge": map[string]interface{}{
					"type":        "string",
					"description": "Edge detection (none, rising, falling, both)",
					"enum":        []string{"none", "rising", "falling", "both"},
				},
			},
			Required: []string{"device_id", "pin", "direction"},
		},
	}, t.handleGPIOSetupPin)

	// GPIOSetPWM tool
	s.AddTool(mcp.Tool{
		Name:        "gpio_set_pwm",
		Description: "Configure PWM on a GPIO pin",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "GPIO device identifier",
				},
				"pin": map[string]interface{}{
					"type":        "number",
					"description": "GPIO pin number",
				},
				"period": map[string]interface{}{
					"type":        "number",
					"description": "PWM period in nanoseconds",
				},
				"duty_cycle": map[string]interface{}{
					"type":        "number",
					"description": "PWM duty cycle in nanoseconds",
				},
			},
			Required: []string{"device_id", "pin", "period", "duty_cycle"},
		},
	}, t.handleGPIOSetPWM)
}

func (t *HardwareMCPTools) handleConnectDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	driverType, _ := request.Params.Arguments["driver_type"].(string)
	config, _ := request.Params.Arguments["config"].(map[string]interface{})

	// Parse configuration based on driver type
	var configData interface{}
	var err error

	switch driverType {
	case "serial":
		configJSON, _ := json.Marshal(config)
		serialConfig := &hardware.SerialDeviceConfig{}
		err = json.Unmarshal(configJSON, serialConfig)
		if err == nil {
			configData = serialConfig
		}
	case "modbus":
		configJSON, _ := json.Marshal(config)
		modbusConfig := &hardware.ModbusDeviceConfig{}
		err = json.Unmarshal(configJSON, modbusConfig)
		if err == nil {
			configData = modbusConfig
		}
	case "can":
		// Parse CAN configuration
		canConfig := &drivers.CANDeviceConfig{}
		if iface, ok := config["interface"].(string); ok {
			canConfig.Interface = iface
		} else {
			canConfig.Interface = "can0"
		}
		if baudRate, ok := config["baud_rate"].(float64); ok {
			canConfig.BaudRate = int(baudRate)
		} else {
			canConfig.BaudRate = 500000
		}
		if timeout, ok := config["timeout"].(float64); ok {
			canConfig.Timeout = int(timeout)
		}
		if enableFD, ok := config["enable_fd"].(bool); ok {
			canConfig.EnableFD = enableFD
		}
		configData = canConfig
	case "gpio":
		// Parse GPIO configuration
		gpioConfig := &drivers.GPIODeviceConfig{}
		if chip, ok := config["chip"].(string); ok {
			gpioConfig.Chip = chip
		} else {
			gpioConfig.Chip = "gpiochip0"
		}
		if pinCount, ok := config["pin_count"].(float64); ok {
			gpioConfig.PinCount = int(pinCount)
		} else {
			gpioConfig.PinCount = 28
		}
		if platform, ok := config["platform"].(string); ok {
			gpioConfig.Platform = platform
		}
		configData = gpioConfig
	default:
		configData = config
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid configuration: %v", err)), nil
	}

	if err := t.agent.ConnectDevice(ctx, deviceID, driverType, configData); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to device: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully connected to device %s", deviceID)), nil
}

func (t *HardwareMCPTools) handleDisconnectDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)

	if err := t.agent.DisconnectDevice(ctx, deviceID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to disconnect from device: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully disconnected from device %s", deviceID)), nil
}

func (t *HardwareMCPTools) handleSendCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	command, _ := request.Params.Arguments["command"].(string)
	params, _ := request.Params.Arguments["params"].(map[string]interface{})

	result, err := t.agent.SendCommand(ctx, deviceID, command, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send command: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *HardwareMCPTools) handleReceiveData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)

	timeout := 5000 * time.Millisecond
	if timeoutMs, ok := request.Params.Arguments["timeout_ms"].(float64); ok {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	data, err := t.agent.ReceiveData(ctx, deviceID, timeout)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to receive data: %v", err)), nil
	}

	dataJSON, _ := json.MarshalIndent(data, "", "  ")
	return mcp.NewToolResultText(string(dataJSON)), nil
}

func (t *HardwareMCPTools) handleGetDeviceStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)

	status, err := t.agent.GetDeviceStatus(ctx, deviceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device status: %v", err)), nil
	}

	statusJSON, _ := json.MarshalIndent(status, "", "  ")
	return mcp.NewToolResultText(string(statusJSON)), nil
}

func (t *HardwareMCPTools) handleListDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	devices, err := t.agent.ListDevices(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list devices: %v", err)), nil
	}

	devicesJSON, _ := json.MarshalIndent(devices, "", "  ")
	return mcp.NewToolResultText(string(devicesJSON)), nil
}

func (t *HardwareMCPTools) handleListDrivers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drivers := t.agent.ListDrivers(ctx)

	driversJSON, _ := json.MarshalIndent(map[string]interface{}{
		"drivers": drivers,
	}, "", "  ")

	return mcp.NewToolResultText(string(driversJSON)), nil
}

func (t *HardwareMCPTools) handleExecuteCommandSequence(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sequenceInterface, _ := request.Params.Arguments["sequence"].(interface{})
	sequenceArray, _ := sequenceInterface.([]interface{})

	commandSequence := &agent.CommandSequence{
		Commands: make([]agent.Command, 0),
	}

	for _, cmdInterface := range sequenceArray {
		cmdMap, _ := cmdInterface.(map[string]interface{})

		cmd := agent.Command{}
		if deviceID, ok := cmdMap["device_id"].(string); ok {
			cmd.DeviceID = deviceID
		}
		if command, ok := cmdMap["command"].(string); ok {
			cmd.Command = command
		}
		if params, ok := cmdMap["params"].(map[string]interface{}); ok {
			cmd.Params = params
		}
		if delayMs, ok := cmdMap["delay_ms"].(float64); ok {
			cmd.Delay = time.Duration(delayMs) * time.Millisecond
		}

		commandSequence.Commands = append(commandSequence.Commands, cmd)
	}

	result, err := t.agent.ExecuteCommand(ctx, commandSequence)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to execute command sequence: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// ===== CAN Bus Handlers =====

func (t *HardwareMCPTools) handleCANSendFrame(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	id, _ := request.Params.Arguments["id"].(float64)

	var data []interface{}
	if dataInterface, ok := request.Params.Arguments["data"]; ok {
		data, _ = dataInterface.([]interface{})
	}

	isExtended, _ := request.Params.Arguments["is_extended"].(bool)
	isRemote, _ := request.Params.Arguments["is_remote"].(bool)

	params := map[string]interface{}{
		"id":          uint32(id),
		"data":        data,
		"is_extended": isExtended,
		"is_remote":   isRemote,
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "send_frame", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send CAN frame: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *HardwareMCPTools) handleCANReceiveFrame(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)

	timeout := 5000 * time.Millisecond
	if timeoutMs, ok := request.Params.Arguments["timeout_ms"].(float64); ok {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	data, err := t.agent.ReceiveData(ctx, deviceID, timeout)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to receive CAN frame: %v", err)), nil
	}

	dataJSON, _ := json.MarshalIndent(data, "", "  ")
	return mcp.NewToolResultText(string(dataJSON)), nil
}

func (t *HardwareMCPTools) handleCANSetFilter(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	filterID, _ := request.Params.Arguments["filter_id"].(float64)
	mask, _ := request.Params.Arguments["mask"].(float64)

	params := map[string]interface{}{
		"id":   uint32(filterID),
		"mask": uint32(mask),
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "set_filter", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set CAN filter: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// ===== GPIO Handlers =====

func (t *HardwareMCPTools) handleGPIOReadPin(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	pin, _ := request.Params.Arguments["pin"].(float64)

	params := map[string]interface{}{
		"pin": int(pin),
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "read_pin", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read GPIO pin: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *HardwareMCPTools) handleGPIOWritePin(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	pin, _ := request.Params.Arguments["pin"].(float64)
	value, _ := request.Params.Arguments["value"].(float64)

	params := map[string]interface{}{
		"pin":   int(pin),
		"value": int(value),
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "write_pin", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write GPIO pin: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *HardwareMCPTools) handleGPIOSetupPin(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	pin, _ := request.Params.Arguments["pin"].(float64)
	direction, _ := request.Params.Arguments["direction"].(string)

	params := map[string]interface{}{
		"pin":       int(pin),
		"direction": direction,
	}

	if activeLow, ok := request.Params.Arguments["active_low"].(bool); ok {
		params["active_low"] = activeLow
	}

	if pull, ok := request.Params.Arguments["pull"].(string); ok {
		params["pull"] = pull
	}

	if edge, ok := request.Params.Arguments["edge"].(string); ok {
		params["edge"] = edge
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "setup_pin", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to setup GPIO pin: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *HardwareMCPTools) handleGPIOSetPWM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, _ := request.Params.Arguments["device_id"].(string)
	pin, _ := request.Params.Arguments["pin"].(float64)
	period, _ := request.Params.Arguments["period"].(float64)
	dutyCycle, _ := request.Params.Arguments["duty_cycle"].(float64)

	params := map[string]interface{}{
		"pin":        int(pin),
		"period":     int(period),
		"duty_cycle": int(dutyCycle),
	}

	result, err := t.agent.SendCommand(ctx, deviceID, "set_pwm", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set GPIO PWM: %v", err)), nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}