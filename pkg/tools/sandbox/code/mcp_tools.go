// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ============================================================================
// code_exec_set_mode 工具 - 设置代码执行模式
// ============================================================================

// codeExecSetModeTool 设置执行模式工具
type codeExecSetModeTool struct {
	module *CodeExecutorModule
}

func (t *codeExecSetModeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "code_exec_set_mode",
		Desc: "Set code execution mode (local, container, auto)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"mode": {
				Type: "string",
				Desc: "Execution mode: 'local' (local execution), 'container' (Docker container), 'auto' (automatic selection)",
			},
		}),
	}, nil
}

func (t *codeExecSetModeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// 验证模式
	validModes := map[string]bool{
		"local":     true,
		"container": true,
		"auto":      true,
	}

	if !validModes[args.Mode] {
		return "", fmt.Errorf("invalid mode: %s (must be 'local', 'container', or 'auto')", args.Mode)
	}

	// 设置执行模式
	t.module.config.ExecutionMode = args.Mode

	result := map[string]any{
		"success": true,
		"mode":    args.Mode,
		"message": fmt.Sprintf("Execution mode set to: %s", args.Mode),
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// ============================================================================
// code_exec_container_status 工具 - 查询容器状态
// ============================================================================

// codeExecContainerStatusTool 容器状态查询工具
type codeExecContainerStatusTool struct {
	module *CodeExecutorModule
}

func (t *codeExecContainerStatusTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "code_exec_container_status",
		Desc:        "Get Docker container executor status and statistics",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *codeExecContainerStatusTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result := map[string]any{
		"success": true,
	}

	// 检查容器执行器是否启用
	if t.module.runner.containerExecutor == nil {
		result["enabled"] = false
		result["message"] = "Container executor not initialized"
	} else if !t.module.runner.containerExecutor.IsEnabled() {
		result["enabled"] = false
		result["message"] = "Container executor disabled"
	} else {
		result["enabled"] = true

		// 获取容器统计信息
		stats := t.module.runner.containerExecutor.GetStats()
		result["stats"] = map[string]any{
			"total_executions":  stats.TotalExecutions,
			"success_count":     stats.SuccessCount,
			"failure_count":     stats.FailureCount,
			"total_duration_ms": stats.TotalDuration.Milliseconds(),
			"active_containers": stats.ActiveContainers,
		}

		// 获取容器池统计信息（如果启用）
		if poolStats := t.module.runner.containerExecutor.GetPoolStats(); poolStats != nil {
			result["pool_stats"] = poolStats
		}

		// 列出活动容器
		containers := t.module.runner.containerExecutor.ListContainers()
		containerList := make([]map[string]any, 0, len(containers))
		for _, container := range containers {
			containerList = append(containerList, map[string]any{
				"id":         container.ID,
				"language":   container.Language,
				"status":     container.Status,
				"created_at": container.CreatedAt.Format("2006-01-02 15:04:05"),
				"exit_code":  container.ExitCode,
			})
		}
		result["containers"] = containerList
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}
