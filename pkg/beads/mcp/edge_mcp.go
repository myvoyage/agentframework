// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package mcp provides edge computing MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/beads/edge"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// EdgeMCPTools provides edge computing MCP tools.
type EdgeMCPTools struct {
	agent *agent.EdgeAgent
}

// NewEdgeMCPTools creates a new EdgeMCPTools instance.
func NewEdgeMCPTools(edgeAgent *agent.EdgeAgent) *EdgeMCPTools {
	return &EdgeMCPTools{
		agent: edgeAgent,
	}
}

// RegisterTools registers all edge computing MCP tools with the MCP server.
func (t *EdgeMCPTools) RegisterTools(s *server.MCPServer) {
	// DeployModel tool
	s.AddTool(mcp.Tool{
		Name:        "deploy_edge_model",
		Description: "Deploy a model to an edge device with optimization",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"deployment_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the deployment",
				},
				"model_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the model file",
				},
				"device_type": map[string]interface{}{
					"type":        "string",
					"description": "Target edge device type",
					"enum":        []string{"raspberry_pi", "jetson_nano", "edge_tpu", "custom"},
				},
				"config": map[string]interface{}{
					"type":        "object",
					"description": "Deployment configuration (optional)",
				},
			},
			Required: []string{"deployment_id", "model_path", "device_type"},
		},
	}, t.handleDeployModel)

	// UndeployModel tool
	s.AddTool(mcp.Tool{
		Name:        "undeploy_edge_model",
		Description: "Remove a model deployment from an edge device",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"deployment_id": map[string]interface{}{
					"type":        "string",
					"description": "Deployment identifier to remove",
				},
			},
			Required: []string{"deployment_id"},
		},
	}, t.handleUndeployModel)

	// OptimizeModel tool
	s.AddTool(mcp.Tool{
		Name:        "optimize_edge_model",
		Description: "Optimize a model for edge deployment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"model_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the model file",
				},
				"device_type": map[string]interface{}{
					"type":        "string",
					"description": "Target edge device type",
					"enum":        []string{"raspberry_pi", "jetson_nano", "edge_tpu", "custom"},
				},
			},
			Required: []string{"model_path", "device_type"},
		},
	}, t.handleOptimizeModel)

	// CompressModel tool
	s.AddTool(mcp.Tool{
		Name:        "compress_edge_model",
		Description: "Compress a model to reduce its size",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"model_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the model file",
				},
				"compression_level": map[string]interface{}{
					"type":        "number",
					"description": "Compression level (1-9, where 9 is maximum compression)",
					"minimum":     1,
					"maximum":     9,
				},
			},
			Required: []string{"model_path", "compression_level"},
		},
	}, t.handleCompressModel)

	// QuantizeModel tool
	s.AddTool(mcp.Tool{
		Name:        "quantize_edge_model",
		Description: "Quantize a model to reduce its memory footprint and improve inference speed",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"model_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the model file",
				},
				"quantization_type": map[string]interface{}{
					"type":        "string",
					"description": "Quantization type",
					"enum":        []string{"int8", "int4", "fp16", "fp8"},
				},
			},
			Required: []string{"model_path", "quantization_type"},
		},
	}, t.handleQuantizeModel)

	// GetDeployment tool
	s.AddTool(mcp.Tool{
		Name:        "get_edge_deployment",
		Description: "Get information about a specific edge deployment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"deployment_id": map[string]interface{}{
					"type":        "string",
					"description": "Deployment identifier",
				},
			},
			Required: []string{"deployment_id"},
		},
	}, t.handleGetDeployment)

	// ListDeployments tool
	s.AddTool(mcp.Tool{
		Name:        "list_edge_deployments",
		Description: "List all edge model deployments",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleListDeployments)

	// GetPerformanceMetrics tool
	s.AddTool(mcp.Tool{
		Name:        "get_edge_performance_metrics",
		Description: "Get performance metrics for a specific deployment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"deployment_id": map[string]interface{}{
					"type":        "string",
					"description": "Deployment identifier",
				},
			},
			Required: []string{"deployment_id"},
		},
	}, t.handleGetPerformanceMetrics)

	// GetAllPerformanceMetrics tool
	s.AddTool(mcp.Tool{
		Name:        "get_all_edge_performance_metrics",
		Description: "Get performance metrics for all edge deployments",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetAllPerformanceMetrics)

	// AllocateResources tool
	s.AddTool(mcp.Tool{
		Name:        "allocate_edge_resources",
		Description: "Allocate resources for edge computing",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"allocation_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the allocation",
				},
				"memory_mb": map[string]interface{}{
					"type":        "number",
					"description": "Memory to allocate in MB",
				},
				"cpu_percent": map[string]interface{}{
					"type":        "number",
					"description": "CPU percentage to allocate (0-100)",
				},
				"duration_seconds": map[string]interface{}{
					"type":        "number",
					"description": "Duration in seconds (0 for permanent allocation)",
				},
			},
			Required: []string{"allocation_id", "memory_mb", "cpu_percent"},
		},
	}, t.handleAllocateResources)

	// ReleaseResources tool
	s.AddTool(mcp.Tool{
		Name:        "release_edge_resources",
		Description: "Release allocated edge computing resources",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"allocation_id": map[string]interface{}{
					"type":        "string",
					"description": "Allocation identifier to release",
				},
			},
			Required: []string{"allocation_id"},
		},
	}, t.handleReleaseResources)

	// GetAvailableResources tool
	s.AddTool(mcp.Tool{
		Name:        "get_available_edge_resources",
		Description: "Get currently available edge computing resources",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetAvailableResources)

	// GetSystemInfo tool
	s.AddTool(mcp.Tool{
		Name:        "get_edge_system_info",
		Description: "Get system information for edge computing",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetSystemInfo)
}

func (t *EdgeMCPTools) handleDeployModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deploymentID, _ := request.Params.Arguments["deployment_id"].(string)
	modelPath, _ := request.Params.Arguments["model_path"].(string)
	deviceTypeStr, _ := request.Params.Arguments["device_type"].(string)

	deviceType, err := t.parseDeviceType(deviceTypeStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid device type: %v", err)), nil
	}

	config, _ := request.Params.Arguments["config"].(map[string]interface{})

	deployment, err := t.agent.DeployModel(ctx, deploymentID, modelPath, deviceType, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to deploy model: %v", err)), nil
	}

	deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
	return mcp.NewToolResultText(string(deploymentJSON)), nil
}

func (t *EdgeMCPTools) handleUndeployModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deploymentID, _ := request.Params.Arguments["deployment_id"].(string)

	if err := t.agent.UndeployModel(ctx, deploymentID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to undeploy model: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully undeployed model %s", deploymentID)), nil
}

func (t *EdgeMCPTools) handleOptimizeModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	modelPath, _ := request.Params.Arguments["model_path"].(string)
	deviceTypeStr, _ := request.Params.Arguments["device_type"].(string)

	deviceType, err := t.parseDeviceType(deviceTypeStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid device type: %v", err)), nil
	}

	optimizedPath, err := t.agent.OptimizeModel(ctx, modelPath, deviceType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to optimize model: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully optimized model: %s", optimizedPath)), nil
}

func (t *EdgeMCPTools) handleCompressModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	modelPath, _ := request.Params.Arguments["model_path"].(string)
	compressionLevel, _ := request.Params.Arguments["compression_level"].(float64)

	compressedPath, err := t.agent.CompressModel(ctx, modelPath, int(compressionLevel))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to compress model: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully compressed model: %s", compressedPath)), nil
}

func (t *EdgeMCPTools) handleQuantizeModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	modelPath, _ := request.Params.Arguments["model_path"].(string)
	quantizationTypeStr, _ := request.Params.Arguments["quantization_type"].(string)

	quantizationType, err := t.parseQuantizationType(quantizationTypeStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid quantization type: %v", err)), nil
	}

	quantizedPath, err := t.agent.QuantizeModel(ctx, modelPath, quantizationType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to quantize model: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully quantized model: %s", quantizedPath)), nil
}

func (t *EdgeMCPTools) handleGetDeployment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deploymentID, _ := request.Params.Arguments["deployment_id"].(string)

	deployment, err := t.agent.GetDeployment(ctx, deploymentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get deployment: %v", err)), nil
	}

	deploymentJSON, _ := json.MarshalIndent(deployment, "", "  ")
	return mcp.NewToolResultText(string(deploymentJSON)), nil
}

func (t *EdgeMCPTools) handleListDeployments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deployments, err := t.agent.ListDeployments(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list deployments: %v", err)), nil
	}

	deploymentsJSON, _ := json.MarshalIndent(map[string]interface{}{
		"deployments": deployments,
		"count":       len(deployments),
	}, "", "  ")

	return mcp.NewToolResultText(string(deploymentsJSON)), nil
}

func (t *EdgeMCPTools) handleGetPerformanceMetrics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deploymentID, _ := request.Params.Arguments["deployment_id"].(string)

	metrics, err := t.agent.GetPerformanceMetrics(ctx, deploymentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get performance metrics: %v", err)), nil
	}

	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	return mcp.NewToolResultText(string(metricsJSON)), nil
}

func (t *EdgeMCPTools) handleGetAllPerformanceMetrics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	metrics, err := t.agent.GetAllPerformanceMetrics(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get performance metrics: %v", err)), nil
	}

	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	return mcp.NewToolResultText(string(metricsJSON)), nil
}

func (t *EdgeMCPTools) handleAllocateResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	allocationID, _ := request.Params.Arguments["allocation_id"].(string)
	memoryMB, _ := request.Params.Arguments["memory_mb"].(float64)
	cpuPercent, _ := request.Params.Arguments["cpu_percent"].(float64)

	duration := 0
	if durationSeconds, ok := request.Params.Arguments["duration_seconds"].(float64); ok {
		duration = int(durationSeconds)
	}

	memory := int64(memoryMB * 1024 * 1024) // Convert MB to bytes
	cpu := cpuPercent / 100.0 // Convert percentage to fraction

	if err := t.agent.AllocateResources(ctx, allocationID, memory, cpu, time.Duration(duration)*time.Second); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to allocate resources: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully allocated resources: %s", allocationID)), nil
}

func (t *EdgeMCPTools) handleReleaseResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	allocationID, _ := request.Params.Arguments["allocation_id"].(string)

	if err := t.agent.ReleaseResources(ctx, allocationID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to release resources: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully released resources: %s", allocationID)), nil
}

func (t *EdgeMCPTools) handleGetAvailableResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	memory, cpu, err := t.agent.GetAvailableResources(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get available resources: %v", err)), nil
	}

	resourcesJSON, _ := json.MarshalIndent(map[string]interface{}{
		"available_memory_mb": memory / (1024 * 1024),
		"available_cpu_percent": cpu * 100,
	}, "", "  ")

	return mcp.NewToolResultText(string(resourcesJSON)), nil
}

func (t *EdgeMCPTools) handleGetSystemInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info, err := t.agent.GetSystemInfo(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get system info: %v", err)), nil
	}

	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	return mcp.NewToolResultText(string(infoJSON)), nil
}

func (t *EdgeMCPTools) parseDeviceType(deviceTypeStr string) (edge.DeviceType, error) {
	switch deviceTypeStr {
	case "raspberry_pi":
		return edge.DeviceRaspberryPi, nil
	case "jetson_nano":
		return edge.DeviceJetsonNano, nil
	case "edge_tpu":
		return edge.DeviceEdgeTPU, nil
	case "custom":
		return edge.DeviceCustom, nil
	default:
		return 0, fmt.Errorf("unknown device type: %s", deviceTypeStr)
	}
}

func (t *EdgeMCPTools) parseQuantizationType(quantizationTypeStr string) (edge.QuantizationType, error) {
	switch quantizationTypeStr {
	case "int8":
		return edge.QuantizationINT8, nil
	case "int4":
		return edge.QuantizationINT4, nil
	case "fp16":
		return edge.QuantizationFP16, nil
	case "fp8":
		return edge.QuantizationFP8, nil
	default:
		return 0, fmt.Errorf("unknown quantization type: %s", quantizationTypeStr)
	}
}