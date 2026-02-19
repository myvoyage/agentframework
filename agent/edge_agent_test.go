// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package agent provides edge computing agent implementation.
package agent

import (
	"context"
	"testing"
	
	"AgentFramework/pkg/beads/edge"
)

// TestEdgeAgent 测试边缘计算代理
func TestEdgeAgent(t *testing.T) {
	ctx := context.Background()
	maxMemory := int64(1024 * 1024 * 1024) // 1GB
	maxCPU := 100.0
	agent := NewEdgeAgent(maxMemory, maxCPU, edge.DeviceRaspberryPi, edge.QuantizationINT8)

	t.Run("Initialize", func(t *testing.T) {
		if err := agent.Initialize(ctx); err != nil {
			t.Errorf("Failed to initialize agent: %v", err)
		}
	})

	t.Run("DeployModel", func(t *testing.T) {
		deploymentID := "test_deployment"
		modelPath := "/path/to/model.tflite"
		deviceType := edge.DeviceRaspberryPi
		config := map[string]interface{}{
			"batch_size": 32,
		}

		deployment, err := agent.DeployModel(ctx, deploymentID, modelPath, deviceType, config)
		if err != nil {
			t.Errorf("Failed to deploy model: %v", err)
		}

		if deployment.ID != deploymentID {
			t.Errorf("Expected deployment ID %s, got %s", deploymentID, deployment.ID)
		}

		if deployment.DeviceType != deviceType {
			t.Errorf("Expected device type %v, got %v", deviceType, deployment.DeviceType)
		}
	})

	t.Run("GetDeployment", func(t *testing.T) {
		deployment, err := agent.GetDeployment(ctx, "test_deployment")
		if err != nil {
			t.Errorf("Failed to get deployment: %v", err)
		}

		if deployment.ID != "test_deployment" {
			t.Errorf("Expected deployment ID test_deployment, got %s", deployment.ID)
		}
	})

	t.Run("ListDeployments", func(t *testing.T) {
		// 添加另一个部署
		agent.DeployModel(ctx, "test_deployment2", "/path/to/model2.tflite", edge.DeviceJetsonNano, nil)

		deployments, err := agent.ListDeployments(ctx)
		if err != nil {
			t.Errorf("Failed to list deployments: %v", err)
		}

		if len(deployments) < 2 {
			t.Errorf("Expected at least 2 deployments, got %d", len(deployments))
		}
	})

	t.Run("UpdateDeployment", func(t *testing.T) {
		newConfig := map[string]interface{}{
			"batch_size": 64,
		}

		deployment, err := agent.UpdateDeployment(ctx, "test_deployment", newConfig)
		if err != nil {
			t.Errorf("Failed to update deployment: %v", err)
		}

		if deployment.Configuration["batch_size"] != 64.0 {
			t.Errorf("Expected batch_size to be 64, got %v", deployment.Configuration["batch_size"])
		}
	})

	t.Run("GetPerformanceMetrics", func(t *testing.T) {
		metrics, err := agent.GetPerformanceMetrics(ctx, "test_deployment")
		if err != nil {
			t.Errorf("Failed to get performance metrics: %v", err)
		}

		if metrics.DeviceType != edge.DeviceRaspberryPi {
			t.Errorf("Expected device type %v, got %v", edge.DeviceRaspberryPi, metrics.DeviceType)
		}
	})

	t.Run("UndeployModel", func(t *testing.T) {
		err := agent.UndeployModel(ctx, "test_deployment")
		if err != nil {
			t.Errorf("Failed to undeploy model: %v", err)
		}

		// 验证已删除
		_, err = agent.GetDeployment(ctx, "test_deployment")
		if err == nil {
			t.Error("Expected error for undeployed deployment")
		}
	})

	t.Run("Close", func(t *testing.T) {
		if err := agent.Close(ctx); err != nil {
			t.Errorf("Failed to close agent: %v", err)
		}
	})
}

// BenchmarkEdgeAgent 性能测试
func BenchmarkEdgeAgent(b *testing.B) {
	ctx := context.Background()
	agent := NewEdgeAgent(1024*1024*1024, 100.0, edge.DeviceRaspberryPi, edge.QuantizationINT8)
	agent.Initialize(ctx)

	modelPath := "/path/to/model.tflite"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.OptimizeModel(ctx, modelPath, edge.DeviceRaspberryPi)
	}
}