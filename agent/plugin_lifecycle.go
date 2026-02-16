// Agent Framework - Plugin Lifecycle Management
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"AgentFramework/agent/errors"
)

// PluginHealth defines the health status of a plugin
type PluginHealth string

const (
	PluginHealthHealthy   PluginHealth = "healthy"
	PluginHealthDegraded  PluginHealth = "degraded"
	PluginHealthUnhealthy PluginHealth = "unhealthy"
	PluginHealthUnknown   PluginHealth = "unknown"
)

// PluginState represents the current state of a plugin
type PluginState string

const (
	PluginStateStopped    PluginState = "stopped"
	PluginStateRunning    PluginState = "running"
	PluginStateRestarting PluginState = "restarting"
	PluginStateStopping   PluginState = "stopping"
	PluginStateError      PluginState = "error"
	PluginStateUnknown    PluginState = "unknown"
)

// PluginMetrics tracks execution metrics for a plugin
type PluginMetrics struct {
	TotalExecutions    int64     // Total number of executions
	SuccessExecutions  int64     // Number of successful executions
	FailureExecutions  int64     // Number of failed executions
	AverageDuration   int64     // Average execution duration in milliseconds
	LastError        string     // Last error message
	LastErrorTime    time.Time     // Time of last error
}

// PluginHealthConfig defines health check configuration
type PluginHealthConfig struct {
	Enabled               bool          `json:"enabled"`               // Enable health checking
	CheckInterval       time.Duration `json:"check_interval"`      // Health check interval
	FailureThreshold    int           `json:"failure_threshold"`     // Max failures before unhealthy
	ResponseThreshold  int           `json:"response_threshold"`    // Max response time before unhealthy
	MaxRetryCount       int           `json:"max_retry_count"`       // Max retry attempts before restart
	RetryDelay          time.Duration `json:"retry_delay"`          // Delay between retries
}

// DefaultPluginHealthConfig returns default health configuration
func DefaultPluginHealthConfig() PluginHealthConfig {
	return PluginHealthConfig{
		Enabled:          true,
		CheckInterval:    30 * time.Second,
		FailureThreshold: 5,
		ResponseThreshold: 5000, // 5 seconds
		MaxRetryCount:    3,
		RetryDelay:       5 * time.Second,
	}
}

// PluginLifecycleManager manages plugin lifecycle including health checks and auto-restart
type PluginLifecycleManager struct {
	plugins    map[string]*PluginEntry
	mu         sync.RWMutex
	healthCfg  PluginHealthConfig
	cancel     context.CancelFunc
}

// PluginEntry represents a plugin with lifecycle information
type PluginEntry struct {
	plugin     Plugin
	state      PluginState
	metrics    PluginMetrics
	health     PluginHealth
	lastCheck   time.Time
	mu         sync.RWMutex
}

// NewPluginLifecycleManager creates a new plugin lifecycle manager
func NewPluginLifecycleManager(config PluginHealthConfig) *PluginLifecycleManager {
	return &PluginLifecycleManager{
		plugins:    make(map[string]*PluginEntry),
		healthCfg:  config,
	}
}

// Register registers a plugin for lifecycle management
func (m *PluginLifecycleManager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &PluginEntry{
		plugin:    plugin,
		state:      PluginStateStopped,
		metrics:     PluginMetrics{},
		health:      PluginHealthHealthy,
		lastCheck:   time.Now(),
	}

	m.plugins[plugin.Name()] = entry

	// Start health checker if enabled
	if m.healthCfg.Enabled {
		m.startHealthChecker()
	}

	return nil
}

// Unregister removes a plugin from lifecycle management
func (m *PluginLifecycleManager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.plugins[name]
	if !exists {
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	// Stop plugin if running
	if entry.state == PluginStateRunning {
		if err := entry.plugin.Shutdown(context.Background()); err != nil {
			return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to shutdown plugin %s", name)
		}
	}

	delete(m.plugins, name)
	return nil
}

// GetState retrieves the current state of a plugin
func (m *PluginLifecycleManager) GetState(name string) (PluginState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.plugins[name]
	if !exists {
		return PluginStateUnknown, errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	return entry.state, nil
}

// GetMetrics retrieves the metrics of a plugin
func (m *PluginLifecycleManager) GetMetrics(name string) (PluginMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.plugins[name]
	if !exists {
		return PluginMetrics{}, errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	return entry.metrics, nil
}

// RecordExecution records a plugin execution for metrics tracking
func (m *PluginLifecycleManager) RecordExecution(name string, success bool, duration int64, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.plugins[name]
	if !exists {
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	entry.metrics.TotalExecutions++
	if success {
		entry.metrics.SuccessExecutions++
	} else {
		entry.metrics.FailureExecutions++
		entry.metrics.LastError = errMsg
		entry.metrics.LastErrorTime = time.Now()
	}

	entry.metrics.AverageDuration = (entry.metrics.AverageDuration*(entry.metrics.TotalExecutions-1) + duration) / entry.metrics.TotalExecutions
	return nil
}

// GetHealth retrieves the current health status of a plugin
func (m *PluginLifecycleManager) GetHealth(name string) (PluginHealth, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.plugins[name]
	if !exists {
		return PluginHealthUnknown, errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	return entry.health, nil
}

// startHealthChecker starts the background health checker
func (m *PluginLifecycleManager) startHealthChecker() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	ticker := time.NewTicker(m.healthCfg.CheckInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.checkAllPlugins(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// checkAllPlugins performs health checks on all registered plugins
func (m *PluginLifecycleManager) checkAllPlugins(ctx context.Context) {
	m.mu.RLock()

	for name, entry := range m.plugins {
		// Perform health check
		healthy := m.checkPluginHealth(ctx, entry)

		// Update health status
		oldHealth := entry.health
		entry.mu.Lock()
		entry.health = healthy
		entry.lastCheck = time.Now()
		entry.mu.Unlock()

		// Log health changes
		if oldHealth != healthy {
			fmt.Printf("[PluginLifecycle] Plugin %s health changed: %s -> %s\n", name, oldHealth, healthy)
		}

		// Auto-restart unhealthy plugins
		if healthy == PluginHealthUnhealthy && entry.state == PluginStateError {
			fmt.Printf("[PluginLifecycle] Attempting to restart unhealthy plugin %s\n", name)
			m.restartPlugin(ctx, name)
		}
	}

	m.mu.RUnlock()
}

// checkPluginHealth performs a single plugin health check
func (m *PluginLifecycleManager) checkPluginHealth(ctx context.Context, entry *PluginEntry) PluginHealth {
	// Check failure threshold
	if entry.metrics.FailureExecutions > int64(m.healthCfg.FailureThreshold) {
		return PluginHealthUnhealthy
	}

	// Check response time threshold
	if entry.metrics.AverageDuration > int64(m.healthCfg.ResponseThreshold) {
		return PluginHealthDegraded
	}

	// Check if plugin is enabled
	if !entry.plugin.IsEnabled() {
		return PluginHealthUnhealthy
	}

	return PluginHealthHealthy
}

// restartPlugin attempts to restart a plugin
func (m *PluginLifecycleManager) restartPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	// Update state to restarting
	entry.state = PluginStateRestarting
	m.mu.Unlock()

	// Shutdown plugin
	if err := entry.plugin.Shutdown(ctx); err != nil {
		return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to shutdown plugin %s", name)
	}

	// Initialize plugin
	if err := entry.plugin.Initialize(ctx, nil); err != nil {
		entry.mu.Lock()
		entry.state = PluginStateError
		entry.metrics.LastError = err.Error()
		entry.metrics.LastErrorTime = time.Now()
		entry.mu.Unlock()
		return errors.Wrapf(err, errors.ErrCodeInitFailed, "failed to initialize plugin %s", name)
	}

	// Update state to running
	entry.mu.Lock()
	entry.state = PluginStateRunning
	entry.mu.Unlock()

	fmt.Printf("[PluginLifecycle] Successfully restarted plugin %s\n", name)
	return nil
}

// Stop stops a plugin gracefully
func (m *PluginLifecycleManager) Stop(name string) error {
	m.mu.Lock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	if entry.state != PluginStateRunning {
		m.mu.Unlock()
		return nil // Already stopped
	}

	// Update state
	entry.state = PluginStateStopping
	m.mu.Unlock()

	// Shutdown plugin
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := entry.plugin.Shutdown(ctx); err != nil {
		entry.mu.Lock()
		entry.state = PluginStateError
		entry.mu.Unlock()
		return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to shutdown plugin %s", name)
	}

	entry.mu.Lock()
	entry.state = PluginStateStopped
	entry.mu.Unlock()
	return nil
}

// StopAll stops all registered plugins
func (m *PluginLifecycleManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var wg sync.WaitGroup
	for name, entry := range m.plugins {
		if entry.state == PluginStateRunning {
			wg.Add(1)
			go func(n string, e *PluginEntry) {
				defer wg.Done()
				if err := e.plugin.Shutdown(context.Background()); err != nil {
					fmt.Printf("[PluginLifecycle] Error stopping plugin %s: %v\n", n, err)
				}
			}(name, entry)
		}
	}

	wg.Wait()
	return nil
}

// GetStats returns lifecycle statistics for all plugins
func (m *PluginLifecycleManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, entry := range m.plugins {
		stats[name] = map[string]interface{}{
			"state":        entry.state,
			"health":       entry.health,
			"total":        entry.metrics.TotalExecutions,
			"success":      entry.metrics.SuccessExecutions,
			"failure":      entry.metrics.FailureExecutions,
			"avg_duration": entry.metrics.AverageDuration,
			"last_error":   entry.metrics.LastError,
		}
	}

	return stats
}

// Global plugin lifecycle manager instance
var globalPluginLifecycleManager *PluginLifecycleManager
var globalPluginLifecycleOnce sync.Once

// InitGlobalPluginLifecycleManager initializes the global plugin lifecycle manager
func InitGlobalPluginLifecycleManager(config PluginHealthConfig) {
	globalPluginLifecycleOnce.Do(func() {
		globalPluginLifecycleManager = NewPluginLifecycleManager(config)
	})
}

// GetGlobalPluginLifecycleManager returns the global plugin lifecycle manager
func GetGlobalPluginLifecycleManager() *PluginLifecycleManager {
	if globalPluginLifecycleManager == nil {
		InitGlobalPluginLifecycleManager(DefaultPluginHealthConfig())
	}
	return globalPluginLifecycleManager
}
