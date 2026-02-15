// Package agent provides edge computing agent implementation.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/beads/edge"
)

// EdgeAgent manages edge computing operations and deployments.
type EdgeAgent struct {
	optimizer         *edge.ModelOptimizer
	resources         *edge.ResourceManager
	deploymentManager *edge.DeploymentManager
	deployments       map[string]*edge.EdgeDeployment
	monitoringEnabled bool
	monitorInterval   time.Duration
	done              chan struct{}
	wg                sync.WaitGroup
	mutex             sync.RWMutex
}

// NewEdgeAgent creates a new EdgeAgent instance.
func NewEdgeAgent(maxMemory int64, maxCPU float64, deviceType edge.DeviceType, quantizationType edge.QuantizationType) *EdgeAgent {
	optimizer := edge.NewModelOptimizer(deviceType, quantizationType)
	resources := edge.NewResourceManager(maxMemory, maxCPU)
	deploymentManager := edge.NewDeploymentManager(optimizer, resources)

	return &EdgeAgent{
		optimizer:         optimizer,
		resources:         resources,
		deploymentManager: deploymentManager,
		deployments:       make(map[string]*edge.EdgeDeployment),
		monitoringEnabled: false,
		monitorInterval:   30 * time.Second,
		done:              make(chan struct{}),
	}
}

// Initialize initializes the edge agent.
func (a *EdgeAgent) Initialize(ctx context.Context) error {
	return nil
}

// DeployModel deploys a model to an edge device.
func (a *EdgeAgent) DeployModel(ctx context.Context, deploymentID string, modelPath string, deviceType edge.DeviceType, config map[string]interface{}) (*edge.EdgeDeployment, error) {
	deployment, err := a.deploymentManager.Deploy(ctx, deploymentID, modelPath, deviceType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy model: %w", err)
	}

	a.mutex.Lock()
	a.deployments[deploymentID] = deployment
	a.mutex.Unlock()

	return deployment, nil
}

// UndeployModel removes a model deployment.
func (a *EdgeAgent) UndeployModel(ctx context.Context, deploymentID string) error {
	if err := a.deploymentManager.Undeploy(ctx, deploymentID); err != nil {
		return fmt.Errorf("failed to undeploy model: %w", err)
	}

	a.mutex.Lock()
	delete(a.deployments, deploymentID)
	a.mutex.Unlock()

	return nil
}

// GetDeployment returns a deployment by ID.
func (a *EdgeAgent) GetDeployment(ctx context.Context, deploymentID string) (*edge.EdgeDeployment, error) {
	return a.deploymentManager.GetDeployment(ctx, deploymentID)
}

// ListDeployments returns all deployments.
func (a *EdgeAgent) ListDeployments(ctx context.Context) ([]*edge.EdgeDeployment, error) {
	return a.deploymentManager.ListDeployments(ctx), nil
}

// UpdateDeployment updates a deployment configuration.
func (a *EdgeAgent) UpdateDeployment(ctx context.Context, deploymentID string, config map[string]interface{}) (*edge.EdgeDeployment, error) {
	deployment, err := a.deploymentManager.UpdateDeployment(ctx, deploymentID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment: %w", err)
	}

	a.mutex.Lock()
	a.deployments[deploymentID] = deployment
	a.mutex.Unlock()

	return deployment, nil
}

// OptimizeModel optimizes a model for edge deployment.
func (a *EdgeAgent) OptimizeModel(ctx context.Context, modelPath string, targetDevice edge.DeviceType) (string, error) {
	return a.optimizer.OptimizeModel(ctx, modelPath, targetDevice)
}

// CompressModel compresses a model.
func (a *EdgeAgent) CompressModel(ctx context.Context, modelPath string, compressionLevel int) (string, error) {
	return a.optimizer.CompressModel(ctx, modelPath, compressionLevel)
}

// QuantizeModel quantizes a model.
func (a *EdgeAgent) QuantizeModel(ctx context.Context, modelPath string, quantizationType edge.QuantizationType) (string, error) {
	return a.optimizer.QuantizeModel(ctx, modelPath, quantizationType)
}

// GetPerformanceMetrics returns performance metrics for a deployment.
func (a *EdgeAgent) GetPerformanceMetrics(ctx context.Context, deploymentID string) (*edge.PerformanceMetrics, error) {
	deployment, err := a.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	return &deployment.Metrics, nil
}

// GetAllPerformanceMetrics returns performance metrics for all deployments.
func (a *EdgeAgent) GetAllPerformanceMetrics(ctx context.Context) (map[string]*edge.PerformanceMetrics, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	metrics := make(map[string]*edge.PerformanceMetrics)

	for deploymentID, deployment := range a.deployments {
		m := deployment.Metrics
		metrics[deploymentID] = &m
	}

	return metrics, nil
}

// AllocateResources allocates resources for a deployment.
func (a *EdgeAgent) AllocateResources(ctx context.Context, allocationID string, memory int64, cpu float64, duration time.Duration) error {
	return a.resources.Allocate(ctx, allocationID, memory, cpu, duration)
}

// ReleaseResources releases allocated resources.
func (a *EdgeAgent) ReleaseResources(ctx context.Context, allocationID string) error {
	return a.resources.Release(ctx, allocationID)
}

// GetAllocation returns a resource allocation by ID.
func (a *EdgeAgent) GetAllocation(ctx context.Context, allocationID string) (edge.ResourceAllocation, error) {
	return a.resources.GetAllocation(ctx, allocationID)
}

// ListAllocations returns all resource allocations.
func (a *EdgeAgent) ListAllocations(ctx context.Context) ([]edge.ResourceAllocation, error) {
	return a.resources.ListAllocations(ctx), nil
}

// GetAvailableResources returns available resources.
func (a *EdgeAgent) GetAvailableResources(ctx context.Context) (memory int64, cpu float64, err error) {
	return a.resources.GetAvailableResources(ctx)
}

// EnableMonitoring enables performance monitoring for deployments.
func (a *EdgeAgent) EnableMonitoring(ctx context.Context, interval time.Duration) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.monitoringEnabled {
		return
	}

	a.monitoringEnabled = true
	a.monitorInterval = interval

	a.wg.Add(1)
	go a.monitorDeployments(ctx)
}

// DisableMonitoring disables performance monitoring.
func (a *EdgeAgent) DisableMonitoring(ctx context.Context) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.monitoringEnabled {
		return
	}

	a.monitoringEnabled = false
	close(a.done)
	a.done = make(chan struct{})
}

// monitorDeployments monitors deployment performance.
func (a *EdgeAgent) monitorDeployments(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.done:
			return
		case <-ticker.C:
			a.updateDeploymentMetrics(ctx)
		}
	}
}

// updateDeploymentMetrics updates metrics for all deployments.
func (a *EdgeAgent) updateDeploymentMetrics(ctx context.Context) {
	a.mutex.RLock()
	deployments := make([]*edge.EdgeDeployment, 0, len(a.deployments))
	for _, deployment := range a.deployments {
		deployments = append(deployments, deployment)
	}
	a.mutex.RUnlock()

	for _, deployment := range deployments {
		// Simulate metrics update (in real implementation, this would collect actual metrics)
		metrics, err := a.optimizer.GetPerformanceMetrics(ctx)
		if err != nil {
			continue
		}

		a.mutex.Lock()
		deployment.Metrics = metrics
		deployment.UpdateTime = time.Now()
		a.mutex.Unlock()
	}
}

// GetSystemInfo returns system information for edge computing.
func (a *EdgeAgent) GetSystemInfo(ctx context.Context) (*EdgeSystemInfo, error) {
	memory, cpu, _ := a.GetAvailableResources(ctx)

	return &EdgeSystemInfo{
		TotalMemory:     a.resources.(*edge.ResourceManager).MaxMemory(),
		AvailableMemory: memory,
		TotalCPU:        a.resources.(*edge.ResourceManager).MaxCPU(),
		AvailableCPU:    cpu,
		DeploymentCount: len(a.deployments),
		MonitoringEnabled: a.monitoringEnabled,
	}, nil
}

// EdgeSystemInfo contains system information for edge computing.
type EdgeSystemInfo struct {
	TotalMemory       int64     `json:"total_memory"`
	AvailableMemory   int64     `json:"available_memory"`
	TotalCPU          float64   `json:"total_cpu"`
	AvailableCPU      float64   `json:"available_cpu"`
	DeploymentCount   int       `json:"deployment_count"`
	MonitoringEnabled bool      `json:"monitoring_enabled"`
}

// Close closes the edge agent and cleans up resources.
func (a *EdgeAgent) Close(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Disable monitoring
	if a.monitoringEnabled {
		a.monitoringEnabled = false
		close(a.done)
	}

	// Wait for monitoring goroutine to stop
	a.wg.Wait()

	// Undeploy all deployments
	for deploymentID := range a.deployments {
		if err := a.UndeployModel(ctx, deploymentID); err != nil {
			fmt.Printf("Failed to undeploy %s: %v\n", deploymentID, err)
		}
	}

	return nil
}