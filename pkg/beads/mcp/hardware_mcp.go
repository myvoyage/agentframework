// Package mcp provides hardware control MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/beads/hardware"

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