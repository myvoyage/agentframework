// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package mcp provides stream processing MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AgentFramework/agent"
	"AgentFramework/pkg/beads/context"
	"AgentFramework/pkg/beads/stream"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StreamMCPTools provides stream processing MCP tools.
type StreamMCPTools struct {
	agent *agent.RealTimeAgent
}

// NewStreamMCPTools creates a new StreamMCPTools instance.
func NewStreamMCPTools(realTimeAgent *agent.RealTimeAgent) *StreamMCPTools {
	return &StreamMCPTools{
		agent: realTimeAgent,
	}
}

// RegisterTools registers all stream processing MCP tools with the MCP server.
func (t *StreamMCPTools) RegisterTools(s *server.MCPServer) {
	// CreatePipeline tool
	s.AddTool(mcp.Tool{
		Name:        "create_stream_pipeline",
		Description: "Create a new data processing pipeline with specified processors",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the pipeline",
				},
				"processors": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{"filter", "map", "batch", "debounce", "throttle"},
							},
							"config": map[string]interface{}{
								"type": "object",
							},
						},
						"required": []string{"type"},
					},
					"description": "List of processors to apply in the pipeline",
				},
				"workers": map[string]interface{}{
					"type":        "number",
					"description": "Number of worker goroutines (default: 1)",
				},
				"buffer_size": map[string]interface{}{
					"type":        "number",
					"description": "Buffer size for channels (default: 100)",
				},
			},
			Required: []string{"pipeline_id", "processors"},
		},
	}, t.handleCreatePipeline)

	// ProcessData tool
	s.AddTool(mcp.Tool{
		Name:        "process_stream_data",
		Description: "Process data through a specified pipeline",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline identifier",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "Data to process",
				},
			},
			Required: []string{"pipeline_id", "data"},
		},
	}, t.handleProcessData)

	// GetPipelineMetrics tool
	s.AddTool(mcp.Tool{
		Name:        "get_stream_pipeline_metrics",
		Description: "Get performance metrics for a specific pipeline",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline identifier",
				},
			},
			Required: []string{"pipeline_id"},
		},
	}, t.handleGetPipelineMetrics)

	// GetAllPipelineMetrics tool
	s.AddTool(mcp.Tool{
		Name:        "get_all_stream_pipeline_metrics",
		Description: "Get performance metrics for all pipelines",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetAllPipelineMetrics)

	// ListPipelines tool
	s.AddTool(mcp.Tool{
		Name:        "list_stream_pipelines",
		Description: "List all available data processing pipelines",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleListPipelines)

	// DeletePipeline tool
	s.AddTool(mcp.Tool{
		Name:        "delete_stream_pipeline",
		Description: "Delete a data processing pipeline",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline identifier",
				},
			},
			Required: []string{"pipeline_id"},
		},
	}, t.handleDeletePipeline)

	// PublishEvent tool
	s.AddTool(mcp.Tool{
		Name:        "publish_stream_event",
		Description: "Publish an event to the event bus",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"event_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of the event",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "Event data",
				},
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Event source (optional)",
				},
			},
			Required: []string{"event_type", "data"},
		},
	}, t.handlePublishEvent)

	// QueryRealTimeData tool
	s.AddTool(mcp.Tool{
		Name:        "query_realtime_data",
		Description: "Query data from the real-time context",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "Filter expression (e.g., 'value.price > 100')",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results (0 = unlimited)",
				},
			},
			Required: []string{},
		},
	}, t.handleQueryRealTimeData)

	// SearchRealTimeData tool
	s.AddTool(mcp.Tool{
		Name:        "search_realtime_data",
		Description: "Search data in the real-time context",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"search_term": map[string]interface{}{
					"type":        "string",
					"description": "Search term",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results (0 = unlimited)",
				},
			},
			Required: []string{"search_term"},
		},
	}, t.handleSearchRealTimeData)

	// GetRealTimeStats tool
	s.AddTool(mcp.Tool{
		Name:        "get_realtime_stats",
		Description: "Get statistics about the real-time context",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetRealTimeStats)

	// ClearRealTimeData tool
	s.AddTool(mcp.Tool{
		Name:        "clear_realtime_data",
		Description: "Clear all data from the real-time context",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleClearRealTimeData)
}

func (t *StreamMCPTools) handleCreatePipeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineID, _ := request.Params.Arguments["pipeline_id"].(string)
	processorsInterface, _ := request.Params.Arguments["processors"].(interface{})
	processorsArray, _ := processorsInterface.([]interface{})

	processors, err := t.parseProcessors(processorsArray)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse processors: %v", err)), nil
	}

	opts := []stream.PipelineOption{}

	if workers, ok := request.Params.Arguments["workers"].(float64); ok {
		opts = append(opts, stream.WithWorkers(int(workers)))
	}

	if bufferSize, ok := request.Params.Arguments["buffer_size"].(float64); ok {
		opts = append(opts, stream.WithBufferSize(int(bufferSize)))
	}

	if err := t.agent.CreatePipeline(ctx, pipelineID, processors, opts...); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create pipeline: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created pipeline %s", pipelineID)), nil
}

func (t *StreamMCPTools) handleProcessData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineID, _ := request.Params.Arguments["pipeline_id"].(string)
	data, _ := request.Params.Arguments["data"].(interface{})

	if err := t.agent.ProcessData(ctx, pipelineID, data); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to process data: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully processed data through pipeline %s", pipelineID)), nil
}

func (t *StreamMCPTools) handleGetPipelineMetrics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineID, _ := request.Params.Arguments["pipeline_id"].(string)

	metrics, err := t.agent.GetPipelineMetrics(ctx, pipelineID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get pipeline metrics: %v", err)), nil
	}

	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	return mcp.NewToolResultText(string(metricsJSON)), nil
}

func (t *StreamMCPTools) handleGetAllPipelineMetrics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	metrics, err := t.agent.GetAllPipelineMetrics(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get pipeline metrics: %v", err)), nil
	}

	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	return mcp.NewToolResultText(string(metricsJSON)), nil
}

func (t *StreamMCPTools) handleListPipelines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelines, err := t.agent.ListPipelines(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list pipelines: %v", err)), nil
	}

	pipelinesJSON, _ := json.MarshalIndent(map[string]interface{}{
		"pipelines": pipelines,
		"count":     len(pipelines),
	}, "", "  ")

	return mcp.NewToolResultText(string(pipelinesJSON)), nil
}

func (t *StreamMCPTools) handleDeletePipeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineID, _ := request.Params.Arguments["pipeline_id"].(string)

	if err := t.agent.DeletePipeline(ctx, pipelineID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete pipeline: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted pipeline %s", pipelineID)), nil
}

func (t *StreamMCPTools) handlePublishEvent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	eventType, _ := request.Params.Arguments["event_type"].(string)
	data, _ := request.Params.Arguments["data"].(map[string]interface{})

	source := "unknown"
	if src, ok := request.Params.Arguments["source"].(string); ok {
		source = src
	}

	event := agent.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Source:    source,
	}

	if err := t.agent.PublishEvent(ctx, event); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to publish event: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully published event %s", eventType)), nil
}

func (t *StreamMCPTools) handleQueryRealTimeData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filter, _ := request.Params.Arguments["filter"].(string)
	limit := 0
	if lim, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(lim)
	}

	query := &context.Query{
		Limit: limit,
	}

	if filter != "" {
		// Simple filter implementation (can be enhanced with expression parsing)
		query.Filter = func(data interface{}) bool {
			return true // Placeholder for actual filter implementation
		}
	}

	results, err := t.agent.QueryRealTimeData(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to query real-time data: %v", err)), nil
	}

	resultsJSON, _ := json.MarshalIndent(map[string]interface{}{
		"results": results,
		"count":   len(results),
	}, "", "  ")

	return mcp.NewToolResultText(string(resultsJSON)), nil
}

func (t *StreamMCPTools) handleSearchRealTimeData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm, _ := request.Params.Arguments["search_term"].(string)
	limit := 0
	if lim, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(lim)
	}

	results, err := t.agent.SearchRealTimeData(ctx, searchTerm, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search real-time data: %v", err)), nil
	}

	resultsJSON, _ := json.MarshalIndent(map[string]interface{}{
		"results": results,
		"count":   len(results),
	}, "", "  ")

	return mcp.NewToolResultText(string(resultsJSON)), nil
}

func (t *StreamMCPTools) handleGetRealTimeStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stats, err := t.agent.GetRealTimeStats(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get real-time stats: %v", err)), nil
	}

	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	return mcp.NewToolResultText(string(statsJSON)), nil
}

func (t *StreamMCPTools) handleClearRealTimeData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.agent.ClearRealTimeData(ctx); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to clear real-time data: %v", err)), nil
	}

	return mcp.NewToolResultText("Successfully cleared real-time data"), nil
}

func (t *StreamMCPTools) parseProcessors(processorsArray []interface{}) ([]stream.DataProcessor, error) {
	processors := make([]stream.DataProcessor, 0)

	for _, procInterface := range processorsArray {
		procMap, ok := procInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid processor configuration")
		}

		procType, ok := procMap["type"].(string)
		if !ok {
			return nil, fmt.Errorf("missing processor type")
		}

		config, _ := procMap["config"].(map[string]interface{})

		processor, err := t.createProcessor(procType, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create processor %s: %w", procType, err)
		}

		processors = append(processors, processor)
	}

	return processors, nil
}

func (t *StreamMCPTools) createProcessor(procType string, config map[string]interface{}) (stream.DataProcessor, error) {
	switch procType {
	case "filter":
		predicate := func(data interface{}) bool {
			return true // Placeholder for actual filter implementation
		}
		return stream.NewFilterProcessor(predicate), nil

	case "map":
		mapper := func(data interface{}) (interface{}, error) {
			return data, nil // Placeholder for actual map implementation
		}
		return stream.NewMapProcessor(mapper), nil

	case "batch":
		batchSize := 100
		if bs, ok := config["batch_size"].(float64); ok {
			batchSize = int(bs)
		}

		timeout := 1000 * time.Millisecond
		if to, ok := config["timeout_ms"].(float64); ok {
			timeout = time.Duration(to) * time.Millisecond
		}

		return stream.NewBatchProcessor(batchSize, timeout), nil

	case "debounce":
		duration := 500 * time.Millisecond
		if dur, ok := config["duration_ms"].(float64); ok {
			duration = time.Duration(dur) * time.Millisecond
		}

		return stream.NewDebounceProcessor(duration), nil

	case "throttle":
		interval := 1000 * time.Millisecond
		if iv, ok := config["interval_ms"].(float64); ok {
			interval = time.Duration(iv) * time.Millisecond
		}

		return stream.NewThrottleProcessor(interval), nil

	default:
		return nil, fmt.Errorf("unknown processor type: %s", procType)
	}
}