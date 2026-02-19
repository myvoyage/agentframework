// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package edge provides edge computing optimization capabilities.
package edge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EdgeOptimizer defines the interface for edge optimization operations.
type EdgeOptimizer interface {
	// OptimizeModel optimizes a model for edge deployment.
	OptimizeModel(ctx context.Context, modelPath string, targetDevice DeviceType) (string, error)

	// CompressModel compresses a model.
	CompressModel(ctx context.Context, modelPath string, compressionLevel int) (string, error)

	// QuantizeModel quantizes a model.
	QuantizeModel(ctx context.Context, modelPath string, quantizationType QuantizationType) (string, error)

	// GetPerformanceMetrics returns performance metrics.
	GetPerformanceMetrics(ctx context.Context) (PerformanceMetrics, error)
}

// DeviceType represents edge device types.
type DeviceType int

const (
	DeviceRaspberryPi DeviceType = iota
	DeviceJetsonNano
	DeviceEdgeTPU
	DeviceCustom
)

// QuantizationType represents model quantization types.
type QuantizationType int

const (
	QuantizationINT8 QuantizationType = iota
	QuantizationINT4
	QuantizationFP16
	QuantizationFP8
)

// PerformanceMetrics contains performance metrics for edge deployments.
type PerformanceMetrics struct {
	ModelLoadTime     time.Duration     `json:"model_load_time"`
	InferenceTime     time.Duration     `json:"inference_time"`
	MemoryUsage       int64             `json:"memory_usage"`
	CPUUsage          float64           `json:"cpu_usage"`
	PowerUsage        float64           `json:"power_usage"`
	Temperature       float64           `json:"temperature"`
	Throughput        float64           `json:"throughput"`
	Accuracy          float64           `json:"accuracy"`
	DeviceType        DeviceType        `json:"device_type"`
	QuantizationType  QuantizationType  `json:"quantization_type"`
	MeasurementTime   time.Time         `json:"measurement_time"`
}

// ModelOptimizer implements edge model optimization.
type ModelOptimizer struct {
	deviceType       DeviceType
	quantizationType QuantizationType
	metrics          PerformanceMetrics
	metricsMutex     sync.RWMutex
	cache            map[string]string
	cacheMutex       sync.RWMutex
}

// NewModelOptimizer creates a new ModelOptimizer instance.
func NewModelOptimizer(deviceType DeviceType, quantizationType QuantizationType) *ModelOptimizer {
	return &ModelOptimizer{
		deviceType:       deviceType,
		quantizationType: quantizationType,
		cache:            make(map[string]string),
	}
}

// OptimizeModel optimizes a model for edge deployment.
func (o *ModelOptimizer) OptimizeModel(ctx context.Context, modelPath string, targetDevice DeviceType) (string, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s:%d", modelPath, targetDevice)
	o.cacheMutex.RLock()
	if cachedPath, exists := o.cache[cacheKey]; exists {
		o.cacheMutex.RUnlock()
		return cachedPath, nil
	}
	o.cacheMutex.RUnlock()

	startTime := time.Now()

	// Simulate optimization (in real implementation, this would use actual optimization tools)
	optimizedPath := fmt.Sprintf("%s_optimized_%d", modelPath, targetDevice)

	// Update metrics
	o.metricsMutex.Lock()
	o.metrics.ModelLoadTime = time.Since(startTime)
	o.metrics.DeviceType = targetDevice
	o.metrics.MeasurementTime = time.Now()
	o.metricsMutex.Unlock()

	// Cache result
	o.cacheMutex.Lock()
	o.cache[cacheKey] = optimizedPath
	o.cacheMutex.Unlock()

	return optimizedPath, nil
}

// CompressModel compresses a model.
func (o *ModelOptimizer) CompressModel(ctx context.Context, modelPath string, compressionLevel int) (string, error) {
	if compressionLevel < 1 || compressionLevel > 9 {
		return "", errors.New("compression level must be between 1 and 9")
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:compressed:%d", modelPath, compressionLevel)
	o.cacheMutex.RLock()
	if cachedPath, exists := o.cache[cacheKey]; exists {
		o.cacheMutex.RUnlock()
		return cachedPath, nil
	}
	o.cacheMutex.RUnlock()

	startTime := time.Now()

	// Simulate compression (in real implementation, this would use actual compression tools)
	compressedPath := fmt.Sprintf("%s_compressed_%d", modelPath, compressionLevel)

	// Update metrics
	o.metricsMutex.Lock()
	o.metrics.ModelLoadTime = time.Since(startTime)
	o.metrics.MeasurementTime = time.Now()
	o.metricsMutex.Unlock()

	// Cache result
	o.cacheMutex.Lock()
	o.cache[cacheKey] = compressedPath
	o.cacheMutex.Unlock()

	return compressedPath, nil
}

// QuantizeModel quantizes a model.
func (o *ModelOptimizer) QuantizeModel(ctx context.Context, modelPath string, quantizationType QuantizationType) (string, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s:quantized:%d", modelPath, quantizationType)
	o.cacheMutex.RLock()
	if cachedPath, exists := o.cache[cacheKey]; exists {
		o.cacheMutex.RUnlock()
		return cachedPath, nil
	}
	o.cacheMutex.RUnlock()

	startTime := time.Now()

	// Simulate quantization (in real implementation, this would use actual quantization tools)
	quantizedPath := fmt.Sprintf("%s_quantized_%d", modelPath, quantizationType)

	// Update metrics
	o.metricsMutex.Lock()
	o.metrics.QuantizationType = quantizationType
	o.metrics.ModelLoadTime = time.Since(startTime)
	o.metrics.MeasurementTime = time.Now()
	o.metricsMutex.Unlock()

	// Cache result
	o.cacheMutex.Lock()
	o.cache[cacheKey] = quantizedPath
	o.cacheMutex.Unlock()

	return quantizedPath, nil
}

// GetPerformanceMetrics returns current performance metrics.
func (o *ModelOptimizer) GetPerformanceMetrics(ctx context.Context) (PerformanceMetrics, error) {
	o.metricsMutex.RLock()
	defer o.metricsMutex.RUnlock()

	return o.metrics, nil
}

// ResourceManager manages resources for edge computing.
type ResourceManager struct {
	maxMemory     int64
	maxCPU        float64
	currentMemory int64
	currentCPU    float64
	allocations   map[string]ResourceAllocation
	mutex         sync.RWMutex
}

// ResourceAllocation represents a resource allocation.
type ResourceAllocation struct {
	ID         string        `json:"id"`
	Memory     int64         `json:"memory"`
	CPU        float64       `json:"cpu"`
	Duration   time.Duration `json:"duration"`
	CreateTime time.Time     `json:"create_time"`
}

// NewResourceManager creates a new ResourceManager instance.
func NewResourceManager(maxMemory int64, maxCPU float64) *ResourceManager {
	return &ResourceManager{
		maxMemory:   maxMemory,
		maxCPU:      maxCPU,
		allocations: make(map[string]ResourceAllocation),
	}
}

// Allocate allocates resources.
func (m *ResourceManager) Allocate(ctx context.Context, id string, memory int64, cpu float64, duration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if allocation exists
	if _, exists := m.allocations[id]; exists {
		return fmt.Errorf("allocation %s already exists", id)
	}

	// Check availability
	if m.currentMemory+memory > m.maxMemory {
		return fmt.Errorf("insufficient memory: requested %d, available %d", memory, m.maxMemory-m.currentMemory)
	}

	if m.currentCPU+cpu > m.maxCPU {
		return fmt.Errorf("insufficient CPU: requested %.2f, available %.2f", cpu, m.maxCPU-m.currentCPU)
	}

	// Allocate resources
	m.currentMemory += memory
	m.currentCPU += cpu

	allocation := ResourceAllocation{
		ID:         id,
		Memory:     memory,
		CPU:        cpu,
		Duration:   duration,
		CreateTime: time.Now(),
	}

	m.allocations[id] = allocation

	// Set up auto-release if duration is specified
	if duration > 0 {
		go func() {
			<-time.After(duration)
			m.Release(ctx, id)
		}()
	}

	return nil
}

// Release releases allocated resources.
func (m *ResourceManager) Release(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	allocation, exists := m.allocations[id]
	if !exists {
		return fmt.Errorf("allocation %s not found", id)
	}

	// Release resources
	m.currentMemory -= allocation.Memory
	m.currentCPU -= allocation.CPU

	delete(m.allocations, id)

	return nil
}

// GetAllocation returns an allocation by ID.
func (m *ResourceManager) GetAllocation(ctx context.Context, id string) (ResourceAllocation, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	allocation, exists := m.allocations[id]
	if !exists {
		return ResourceAllocation{}, fmt.Errorf("allocation %s not found", id)
	}

	return allocation, nil
}

// ListAllocations returns all allocations.
func (m *ResourceManager) ListAllocations(ctx context.Context) []ResourceAllocation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	allocations := make([]ResourceAllocation, 0, len(m.allocations))
	for _, allocation := range m.allocations {
		allocations = append(allocations, allocation)
	}

	return allocations
}

// GetAvailableResources returns available resources.
func (m *ResourceManager) GetAvailableResources(ctx context.Context) (memory int64, cpu float64) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.maxMemory - m.currentMemory, m.maxCPU - m.currentCPU
}

// EdgeDeployment manages edge deployments.
type EdgeDeployment struct {
	ID           string                 `json:"id"`
	ModelPath    string                 `json:"model_path"`
	DeviceType   DeviceType             `json:"device_type"`
	Status       string                 `json:"status"`
	Metrics      PerformanceMetrics      `json:"metrics"`
	Configuration map[string]interface{} `json:"configuration"`
	CreateTime   time.Time              `json:"create_time"`
	UpdateTime   time.Time              `json:"update_time"`
}

// DeploymentManager manages edge deployments.
type DeploymentManager struct {
	deployments  map[string]*EdgeDeployment
	optimizer    *ModelOptimizer
	resources    *ResourceManager
	mutex        sync.RWMutex
}

// NewDeploymentManager creates a new DeploymentManager instance.
func NewDeploymentManager(optimizer *ModelOptimizer, resources *ResourceManager) *DeploymentManager {
	return &DeploymentManager{
		deployments: make(map[string]*EdgeDeployment),
		optimizer:   optimizer,
		resources:   resources,
	}
}

// Deploy deploys a model to an edge device.
func (m *DeploymentManager) Deploy(ctx context.Context, id string, modelPath string, deviceType DeviceType, config map[string]interface{}) (*EdgeDeployment, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if deployment exists
	if _, exists := m.deployments[id]; exists {
		return nil, fmt.Errorf("deployment %s already exists", id)
	}

	// Optimize model for target device
	optimizedPath, err := m.optimizer.OptimizeModel(ctx, modelPath, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize model: %w", err)
	}

	// Create deployment
	deployment := &EdgeDeployment{
		ID:           id,
		ModelPath:    optimizedPath,
		DeviceType:   deviceType,
		Status:       "deployed",
		Configuration: config,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}

	// Get initial metrics
	metrics, err := m.optimizer.GetPerformanceMetrics(ctx)
	if err == nil {
		deployment.Metrics = metrics
	}

	m.deployments[id] = deployment

	return deployment, nil
}

// Undeploy removes a deployment.
func (m *DeploymentManager) Undeploy(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deployment, exists := m.deployments[id]
	if !exists {
		return fmt.Errorf("deployment %s not found", id)
	}

	// Update status
	deployment.Status = "undeployed"
	deployment.UpdateTime = time.Now()

	// Release resources
	m.resources.Release(ctx, id)

	delete(m.deployments, id)

	return nil
}

// GetDeployment returns a deployment by ID.
func (m *DeploymentManager) GetDeployment(ctx context.Context, id string) (*EdgeDeployment, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	deployment, exists := m.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found", id)
	}

	return deployment, nil
}

// ListDeployments returns all deployments.
func (m *DeploymentManager) ListDeployments(ctx context.Context) []*EdgeDeployment {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	deployments := make([]*EdgeDeployment, 0, len(m.deployments))
	for _, deployment := range m.deployments {
		deployments = append(deployments, deployment)
	}

	return deployments
}

// UpdateDeployment updates a deployment.
func (m *DeploymentManager) UpdateDeployment(ctx context.Context, id string, config map[string]interface{}) (*EdgeDeployment, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deployment, exists := m.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found", id)
	}

	// Update configuration
	for key, value := range config {
		deployment.Configuration[key] = value
	}

	deployment.UpdateTime = time.Now()

	return deployment, nil
}