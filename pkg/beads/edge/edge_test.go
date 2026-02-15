// Package edge provides edge computing optimization capabilities.
package edge

import (
	"context"
	"testing"
	"time"
)

// TestModelOptimizer 测试模型优化器
func TestModelOptimizer(t *testing.T) {
	ctx := context.Background()
	optimizer := NewModelOptimizer(DeviceRaspberryPi, QuantizationINT8)

	t.Run("OptimizeModel", func(t *testing.T) {
		modelPath := "/path/to/model.tflite"
		targetDevice := DeviceRaspberryPi

		optimizedPath, err := optimizer.OptimizeModel(ctx, modelPath, targetDevice)
		if err != nil {
			t.Errorf("Failed to optimize model: %v", err)
		}

		if optimizedPath == "" {
			t.Error("Optimized path should not be empty")
		}

		// 验证缓存
		optimizedPath2, err := optimizer.OptimizeModel(ctx, modelPath, targetDevice)
		if err != nil {
			t.Errorf("Failed to optimize model (cached): %v", err)
		}

		if optimizedPath != optimizedPath2 {
			t.Error("Cached optimized path should match original")
		}
	})

	t.Run("CompressModel", func(t *testing.T) {
		modelPath := "/path/to/model.tflite"
		compressionLevel := 5

		compressedPath, err := optimizer.CompressModel(ctx, modelPath, compressionLevel)
		if err != nil {
			t.Errorf("Failed to compress model: %v", err)
		}

		if compressedPath == "" {
			t.Error("Compressed path should not be empty")
		}

		// 测试无效的压缩级别
		_, err = optimizer.CompressModel(ctx, modelPath, 15)
		if err == nil {
			t.Error("Expected error for invalid compression level")
		}
	})

	t.Run("QuantizeModel", func(t *testing.T) {
		modelPath := "/path/to/model.tflite"
		quantizationType := QuantizationINT8

		quantizedPath, err := optimizer.QuantizeModel(ctx, modelPath, quantizationType)
		if err != nil {
			t.Errorf("Failed to quantize model: %v", err)
		}

		if quantizedPath == "" {
			t.Error("Quantized path should not be empty")
		}
	})

	t.Run("GetPerformanceMetrics", func(t *testing.T) {
		// 先优化一个模型以生成指标
		modelPath := "/path/to/model.tflite"
		optimizer.OptimizeModel(ctx, modelPath, DeviceJetsonNano)

		metrics, err := optimizer.GetPerformanceMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get performance metrics: %v", err)
		}

		if metrics.DeviceType != DeviceJetsonNano {
			t.Errorf("Expected device type %v, got %v", DeviceJetsonNano, metrics.DeviceType)
		}

		if metrics.QuantizationType != QuantizationINT8 {
			t.Errorf("Expected quantization type %v, got %v", QuantizationINT8, metrics.QuantizationType)
		}
	})
}

// TestResourceManager 测试资源管理器
func TestResourceManager(t *testing.T) {
	ctx := context.Background()
	maxMemory := int64(1024 * 1024 * 1024) // 1GB
	maxCPU := 100.0

	manager := NewResourceManager(maxMemory, maxCPU)

	t.Run("AllocateResources", func(t *testing.T) {
		allocationID := "allocation1"
		memory := int64(512 * 1024 * 1024) // 512MB
		cpu := 50.0
		duration := 1 * time.Hour

		err := manager.Allocate(ctx, allocationID, memory, cpu, duration)
		if err != nil {
			t.Errorf("Failed to allocate resources: %v", err)
		}

		// 验证分配
		allocation, err := manager.GetAllocation(ctx, allocationID)
		if err != nil {
			t.Errorf("Failed to get allocation: %v", err)
		}

		if allocation.Memory != memory {
			t.Errorf("Expected memory %d, got %d", memory, allocation.Memory)
		}

		if allocation.CPU != cpu {
			t.Errorf("Expected CPU %f, got %f", cpu, allocation.CPU)
		}
	})

	t.Run("AllocateDuplicate", func(t *testing.T) {
		allocationID := "allocation1"

		err := manager.Allocate(ctx, allocationID, 100, 10.0, 0)
		if err == nil {
			t.Error("Expected error for duplicate allocation")
		}
	})

	t.Run("ReleaseResources", func(t *testing.T) {
		allocationID := "allocation2"
		memory := int64(256 * 1024 * 1024)
		cpu := 25.0

		// 分配资源
		err := manager.Allocate(ctx, allocationID, memory, cpu, 0)
		if err != nil {
			t.Errorf("Failed to allocate resources: %v", err)
		}

		// 释放资源
		err = manager.Release(ctx, allocationID)
		if err != nil {
			t.Errorf("Failed to release resources: %v", err)
		}

		// 验证已释放
		_, err = manager.GetAllocation(ctx, allocationID)
		if err == nil {
			t.Error("Expected error for released allocation")
		}
	})

	t.Run("AllocateExceedsLimit", func(t *testing.T) {
		allocationID := "allocation3"

		// 尝试分配超过限制的资源
		err := manager.Allocate(ctx, allocationID, 2*1024*1024*1024, 150.0, 0)
		if err == nil {
			t.Error("Expected error for allocation exceeding limits")
		}
	})

	t.Run("GetAvailableResources", func(t *testing.T) {
		// 先分配一些资源
		manager.Allocate(ctx, "allocation4", 100*1024*1024, 10.0, 0)

		memory, cpu := manager.GetAvailableResources(ctx)

		expectedMemory := maxMemory - 100*1024*1024
		if memory != expectedMemory {
			t.Errorf("Expected available memory %d, got %d", expectedMemory, memory)
		}

		expectedCPU := maxCPU - 10.0
		if cpu != expectedCPU {
			t.Errorf("Expected available CPU %f, got %f", expectedCPU, cpu)
		}
	})
}

// TestDeploymentManager 测试部署管理器
func TestDeploymentManager(t *testing.T) {
	ctx := context.Background()
	optimizer := NewModelOptimizer(DeviceRaspberryPi, QuantizationINT8)
	resources := NewResourceManager(1024*1024*1024, 100.0)
	manager := NewDeploymentManager(optimizer, resources)

	t.Run("DeployModel", func(t *testing.T) {
		deploymentID := "deployment1"
		modelPath := "/path/to/model.tflite"
		deviceType := DeviceRaspberryPi
		config := map[string]interface{}{
			"batch_size":  32,
			"input_shape": []int{224, 224, 3},
		}

		deployment, err := manager.Deploy(ctx, deploymentID, modelPath, deviceType, config)
		if err != nil {
			t.Errorf("Failed to deploy model: %v", err)
		}

		if deployment.ID != deploymentID {
			t.Errorf("Expected deployment ID %s, got %s", deploymentID, deployment.ID)
		}

		if deployment.Status != "deployed" {
			t.Errorf("Expected status 'deployed', got %s", deployment.Status)
		}

		if deployment.DeviceType != deviceType {
			t.Errorf("Expected device type %v, got %v", deviceType, deployment.DeviceType)
		}
	})

	t.Run("DeployDuplicate", func(t *testing.T) {
		deploymentID := "deployment1"
		modelPath := "/path/to/model2.tflite"

		_, err := manager.Deploy(ctx, deploymentID, modelPath, DeviceJetsonNano, nil)
		if err == nil {
			t.Error("Expected error for duplicate deployment")
		}
	})

	t.Run("GetDeployment", func(t *testing.T) {
		deploymentID := "deployment1"

		deployment, err := manager.GetDeployment(ctx, deploymentID)
		if err != nil {
			t.Errorf("Failed to get deployment: %v", err)
		}

		if deployment.ID != deploymentID {
			t.Errorf("Expected deployment ID %s, got %s", deploymentID, deployment.ID)
		}
	})

	t.Run("ListDeployments", func(t *testing.T) {
		// 再添加一个部署
		manager.Deploy(ctx, "deployment2", "/path/to/model2.tflite", DeviceEdgeTPU, nil)

		deployments := manager.ListDeployments(ctx)

		if len(deployments) < 2 {
			t.Errorf("Expected at least 2 deployments, got %d", len(deployments))
		}
	})

	t.Run("UpdateDeployment", func(t *testing.T) {
		deploymentID := "deployment1"
		newConfig := map[string]interface{}{
			"batch_size":  64,
			"input_shape": []int{299, 299, 3},
		}

		deployment, err := manager.UpdateDeployment(ctx, deploymentID, newConfig)
		if err != nil {
			t.Errorf("Failed to update deployment: %v", err)
		}

		if deployment.Configuration["batch_size"] != 64.0 {
			t.Errorf("Expected batch_size to be updated to 64, got %v", deployment.Configuration["batch_size"])
		}
	})

	t.Run("UndeployModel", func(t *testing.T) {
		deploymentID := "deployment1"

		err := manager.Undeploy(ctx, deploymentID)
		if err != nil {
			t.Errorf("Failed to undeploy model: %v", err)
		}

		// 验证已删除
		_, err = manager.GetDeployment(ctx, deploymentID)
		if err == nil {
			t.Error("Expected error for undeployed deployment")
		}
	})
}

// BenchmarkModelOptimizer 性能测试
func BenchmarkModelOptimizer(b *testing.B) {
	ctx := context.Background()
	optimizer := NewModelOptimizer(DeviceRaspberryPi, QuantizationINT8)
	modelPath := "/path/to/model.tflite"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.OptimizeModel(ctx, modelPath, DeviceRaspberryPi)
	}
}