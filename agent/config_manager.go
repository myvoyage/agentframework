// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigChangeHandler 定义配置变化处理函数类型
type ConfigChangeHandler func(oldConfig, newConfig *HostConfig)

// ConfigValidator 定义配置验证器函数类型
type ConfigValidator func(config *HostConfig) error

// ConfigManager provides a centralized way to access configuration settings
// It supports retrieving configuration for different components and allows for
// runtime updates, validation, hot reload, and persistence

type ConfigManager interface {
	// GetHostConfig returns the full host configuration
	GetHostConfig() *HostConfig

	// GetModelConfig retrieves configuration for a specific model
	// It returns the model configuration and a boolean indicating if the model was found
	GetModelConfig(modelName string) (ModelConfig, bool)

	// GetAllModelConfigs returns all model configurations
	GetAllModelConfigs() map[string]ModelConfig

	// GetDefaultModelConfig returns the default model configuration
	GetDefaultModelConfig() (ModelConfig, bool)

	// GetModelType retrieves the type of a specific model
	// It returns the model type and a boolean indicating if the model was found
	GetModelType(modelName string) (string, bool)

	// IsModelEnabled checks if a specific model is enabled
	// It returns true if the model is enabled, false otherwise or if the model was not found
	IsModelEnabled(modelName string) bool

	// GetModelPriority retrieves the priority of a specific model
	// It returns the model priority and a boolean indicating if the model was found
	GetModelPriority(modelName string) (int, bool)

	// GetModelOptions retrieves the options for a specific model
	// It returns the model options and a boolean indicating if the model was found
	GetModelOptions(modelName string) (map[string]any, bool)

	// GetAgentConfig retrieves configuration for a specific agent
	// It returns the agent specification and a boolean indicating if the agent was found
	GetAgentConfig(agentName string) (AgentSpec, bool)

	// GetWorkflowConfig retrieves configuration for a specific workflow
	// It returns the workflow specification and a boolean indicating if the workflow was found
	GetWorkflowConfig(workflowName string) (WorkflowSpec, bool)

	// GetModelCacheConfig returns the model cache configuration
	GetModelCacheConfig() ModelCacheSpec

	// GetMemoryMonitorConfig returns the memory monitor configuration
	GetMemoryMonitorConfig() MemoryMonitorSpec

	// GetThreadStoreConfig returns the thread store configuration
	GetThreadStoreConfig() ThreadStoreSpec

	// -----------------------------------------------
	// Configuration Hot Reload and Validation Methods
	// -----------------------------------------------

	// ReloadConfig reloads the configuration from the source
	// It returns an error if the reload fails
	ReloadConfig() error

	// EnableConfigWatch enables watching for configuration changes
	// and automatically reloads the configuration when changes are detected
	// It returns an error if watching fails
	EnableConfigWatch() error

	// DisableConfigWatch disables watching for configuration changes
	DisableConfigWatch() error

	// RegisterConfigChangeHandler registers a handler for configuration changes
	// The handler will be called when the configuration is reloaded
	RegisterConfigChangeHandler(handler ConfigChangeHandler)

	// RegisterConfigValidator registers a validator for configuration validation
	// The validator will be called before the configuration is applied
	RegisterConfigValidator(validator ConfigValidator)

	// ValidateConfig validates the current configuration
	// It returns an error if the validation fails
	ValidateConfig() error

	// UpdateConfig updates the configuration with the provided host configuration
	// It returns an error if the update fails or validation fails
	UpdateConfig(cfg *HostConfig) error

	// -----------------------------------------------
	// Configuration Persistence Methods
	// -----------------------------------------------

	// SaveConfig saves the current configuration to the original source
	// It returns an error if the save fails or validation fails
	SaveConfig() error

	// SaveConfigToFile saves the current configuration to a file
	// It returns an error if the save fails or validation fails
	SaveConfigToFile(path string) error

	// SaveConfigWithConfig saves the provided configuration to the original source
	// It returns an error if the save fails or validation fails
	SaveConfigWithConfig(cfg *HostConfig) error
}

// DefaultConfigManager is the default implementation of ConfigManager
// that wraps a HostConfig and provides fast access to its components
// It also supports configuration hot reload and validation

type DefaultConfigManager struct {
	cfg          *HostConfig           // The underlying host configuration
	sourcePath   string                // Configuration source path (if loaded from file)
	handlers     []ConfigChangeHandler // Configuration change handlers
	validators   []ConfigValidator     // Configuration validators
	watchEnabled bool                  // Whether config watching is enabled
	watchStopped chan struct{}         // Channel to stop watching
	mu           sync.RWMutex          // Read-write lock for thread safety
}

// NewConfigManager creates a new ConfigManager instance from the given HostConfig
// It provides a centralized way to access various configuration components
func NewConfigManager(cfg *HostConfig) ConfigManager {
	return &DefaultConfigManager{
		cfg:          cfg,
		handlers:     []ConfigChangeHandler{},
		validators:   []ConfigValidator{},
		watchEnabled: false,
		watchStopped: make(chan struct{}),
	}
}

// NewConfigManagerFromFile creates a new ConfigManager instance from a configuration file
// It loads the configuration from the specified file path
func NewConfigManagerFromFile(path string) (ConfigManager, error) {
	cfg, err := LoadHostConfigFile(path)
	if err != nil {
		return nil, err
	}

	return &DefaultConfigManager{
		cfg:          cfg,
		sourcePath:   path,
		handlers:     []ConfigChangeHandler{},
		validators:   []ConfigValidator{},
		watchEnabled: false,
		watchStopped: make(chan struct{}),
	}, nil
}

// ReloadConfig reloads the configuration from the source
// It returns an error if the reload fails
func (cm *DefaultConfigManager) ReloadConfig() error {
	if cm.sourcePath == "" {
		return fmt.Errorf("no configuration source path specified")
	}

	// Load new configuration
	newCfg, err := LoadHostConfigFile(cm.sourcePath)
	if err != nil {
		return err
	}

	// Validate new configuration
	if err := cm.validateConfig(newCfg); err != nil {
		return err
	}

	// Update configuration with thread safety
	cm.mu.Lock()
	oldCfg := cm.cfg
	cm.cfg = newCfg
	cm.mu.Unlock()

	// Notify all change handlers
	for _, handler := range cm.handlers {
		handler(oldCfg, newCfg)
	}

	return nil
}

// EnableConfigWatch enables watching for configuration changes
// and automatically reloads the configuration when changes are detected
// It returns an error if watching fails
func (cm *DefaultConfigManager) EnableConfigWatch() error {
	if cm.sourcePath == "" {
		return fmt.Errorf("no configuration source path specified")
	}

	if cm.watchEnabled {
		return nil // Already watching
	}

	cm.mu.Lock()
	cm.watchEnabled = true
	cm.watchStopped = make(chan struct{})
	cm.mu.Unlock()

	// Create a new file system watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cm.mu.Lock()
		cm.watchEnabled = false
		cm.mu.Unlock()
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Watch the configuration file and its directory
	// Watching the directory is necessary to catch rename/move operations
	configDir := filepath.Dir(cm.sourcePath)
	err = watcher.Add(configDir)
	if err != nil {
		watcher.Close()
		cm.mu.Lock()
		cm.watchEnabled = false
		cm.mu.Unlock()
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Start the watcher goroutine with debouncing
	go cm.watchConfigChanges(watcher)

	return nil
}

// watchConfigChanges monitors configuration file changes with debouncing
func (cm *DefaultConfigManager) watchConfigChanges(watcher *fsnotify.Watcher) {
	defer watcher.Close()

	// Debouncing: accumulate events and only reload after a quiet period
	var debounceTimer *time.Timer
	debounceDuration := 300 * time.Millisecond // Adjust as needed

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Check if the event is for our config file
			if event.Name != cm.sourcePath {
				continue
			}

			// Only process Write and Create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Reset or create debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(debounceDuration, func() {
				cm.handleConfigChange()
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Config watcher error: %v\n", err)

		case <-cm.watchStopped:
			// Stop the debounce timer if active
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return
		}
	}
}

// handleConfigChange handles configuration file changes
func (cm *DefaultConfigManager) handleConfigChange() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.watchEnabled {
		return
	}

	// Save old config for handlers
	oldConfig := cm.cfg

	// Reload the configuration
	newConfig, err := LoadHostConfigFile(cm.sourcePath)
	if err != nil {
		fmt.Printf("Failed to reload config: %v\n", err)
		return
	}

	// Validate the new configuration
	for _, validator := range cm.validators {
		if err := validator(newConfig); err != nil {
			fmt.Printf("Config validation failed after reload: %v\n", err)
			return
		}
	}

	// Update current config
	cm.cfg = newConfig

	// Notify all registered handlers
	for _, handler := range cm.handlers {
		// Call handler in a goroutine to avoid blocking
		go func(h ConfigChangeHandler) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Config change handler panic: %v\n", r)
				}
			}()
			h(oldConfig, newConfig)
		}(handler)
	}

	fmt.Printf("Configuration reloaded successfully from %s\n", cm.sourcePath)
}

// DisableConfigWatch disables watching for configuration changes
func (cm *DefaultConfigManager) DisableConfigWatch() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.watchEnabled {
		return nil // Already not watching
	}

	cm.watchEnabled = false
	close(cm.watchStopped)

	return nil
}

// RegisterConfigChangeHandler registers a handler for configuration changes
// The handler will be called when the configuration is reloaded
func (cm *DefaultConfigManager) RegisterConfigChangeHandler(handler ConfigChangeHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.handlers = append(cm.handlers, handler)
}

// RegisterConfigValidator registers a validator for configuration validation
// The validator will be called before the configuration is applied
func (cm *DefaultConfigManager) RegisterConfigValidator(validator ConfigValidator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.validators = append(cm.validators, validator)
}

// ValidateConfig validates the current configuration
// It returns an error if the validation fails
func (cm *DefaultConfigManager) ValidateConfig() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.validateConfig(cm.cfg)
}

// validateConfig validates the given configuration
// It returns an error if the validation fails
func (cm *DefaultConfigManager) validateConfig(cfg *HostConfig) error {
	// Run all registered validators
	for _, validator := range cm.validators {
		if err := validator(cfg); err != nil {
			return err
		}
	}

	// Default validation rules

	// 1. Check if default model is specified and exists
	if cfg.DefaultModel != "" {
		if _, exists := cfg.Models[cfg.DefaultModel]; !exists {
			return fmt.Errorf("default model %q not found in models configuration", cfg.DefaultModel)
		}
	}

	// 2. Validate model configurations
	if len(cfg.Models) == 0 {
		return fmt.Errorf("at least one model configuration is required")
	}

	for modelName, modelConfig := range cfg.Models {
		if modelConfig.Type == "" {
			return fmt.Errorf("model %q missing required field: type", modelName)
		}
	}

	// 3. Validate thread store configuration
	if cfg.ThreadStore.Type == "" {
		return fmt.Errorf("thread store type is required")
	}

	// Validate thread store specific configurations based on type
	switch cfg.ThreadStore.Type {
	case "file":
		if cfg.ThreadStore.Dir == "" {
			return fmt.Errorf("thread store directory is required for file type")
		}
	case "redis":
		if cfg.ThreadStore.RedisAddr == "" {
			return fmt.Errorf("redis address is required for redis type thread store")
		}
	case "sql":
		if cfg.ThreadStore.DriverName == "" {
			return fmt.Errorf("driver name is required for sql type thread store")
		}
		if cfg.ThreadStore.DSN == "" {
			return fmt.Errorf("dsn is required for sql type thread store")
		}
	}

	// 4. Validate agents configuration
	if len(cfg.Agents) == 0 {
		return fmt.Errorf("at least one agent configuration is required")
	}

	// Check for duplicate agent names
	agentNames := make(map[string]bool)
	for _, agent := range cfg.Agents {
		if agent.Name == "" {
			return fmt.Errorf("agent name is required")
		}
		if agent.Kind == "" {
			return fmt.Errorf("agent kind is required for agent %q", agent.Name)
		}
		if agentNames[agent.Name] {
			return fmt.Errorf("duplicate agent name: %q", agent.Name)
		}
		agentNames[agent.Name] = true

		// Check if agent model exists
		if agent.Model != "" {
			if _, exists := cfg.Models[agent.Model]; !exists {
				return fmt.Errorf("model %q referenced by agent %q not found", agent.Model, agent.Name)
			}
		}
	}

	// 5. Validate workflows configuration
	// Check for duplicate workflow names
	workflowNames := make(map[string]bool)
	for _, workflow := range cfg.Workflows {
		if workflow.Name == "" {
			return fmt.Errorf("workflow name is required")
		}
		if workflow.Kind == "" {
			return fmt.Errorf("workflow kind is required for workflow %q", workflow.Name)
		}
		if workflowNames[workflow.Name] {
			return fmt.Errorf("duplicate workflow name: %q", workflow.Name)
		}
		workflowNames[workflow.Name] = true

		// Check if workflow model exists
		if workflow.Model != "" {
			if _, exists := cfg.Models[workflow.Model]; !exists {
				return fmt.Errorf("model %q referenced by workflow %q not found", workflow.Model, workflow.Name)
			}
		}
	}

	// 6. Validate memory management configuration
	// Model cache validation
	if cfg.Memory.ModelCache.Enabled {
		if cfg.Memory.ModelCache.MaxSize <= 0 {
			return fmt.Errorf("model cache max size must be greater than 0")
		}
		if cfg.Memory.ModelCache.TTL < 0 {
			return fmt.Errorf("model cache TTL must be non-negative")
		}
		if cfg.Memory.ModelCache.CleanupInterval <= 0 {
			return fmt.Errorf("model cache cleanup interval must be greater than 0")
		}
	}

	// Memory monitor validation
	if cfg.Memory.MemoryMonitor.Enabled {
		if cfg.Memory.MemoryMonitor.Interval <= 0 {
			return fmt.Errorf("memory monitor interval must be greater than 0")
		}
		if cfg.Memory.MemoryMonitor.HistorySize <= 0 {
			return fmt.Errorf("memory monitor history size must be greater than 0")
		}
		if cfg.Memory.MemoryMonitor.AlertThreshold <= 0 {
			return fmt.Errorf("memory monitor alert threshold must be greater than 0")
		}
		if cfg.Memory.MemoryMonitor.AlertInterval <= 0 {
			return fmt.Errorf("memory monitor alert interval must be greater than 0")
		}
	}

	return nil
}

// UpdateConfig updates the configuration with the provided host configuration
// It returns an error if the update fails or validation fails
func (cm *DefaultConfigManager) UpdateConfig(cfg *HostConfig) error {
	// Validate new configuration
	if err := cm.validateConfig(cfg); err != nil {
		return err
	}

	// Update configuration with thread safety
	cm.mu.Lock()
	oldCfg := cm.cfg
	cm.cfg = cfg
	cm.mu.Unlock()

	// Notify all change handlers
	for _, handler := range cm.handlers {
		handler(oldCfg, cfg)
	}

	return nil
}

// GetHostConfig returns the full host configuration
// It returns a copy of the configuration to prevent external modification
func (cm *DefaultConfigManager) GetHostConfig() *HostConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a deep copy of the configuration
	// For simplicity, we'll use a manual copy since the struct is not too complex
	cfgCopy := *cm.cfg

	// Copy maps and slices to prevent external modification
	cfgCopy.Models = make(map[string]ModelConfig, len(cm.cfg.Models))
	for k, v := range cm.cfg.Models {
		cfgCopy.Models[k] = v
	}

	cfgCopy.Agents = make([]AgentSpec, len(cm.cfg.Agents))
	copy(cfgCopy.Agents, cm.cfg.Agents)

	cfgCopy.Workflows = make([]WorkflowSpec, len(cm.cfg.Workflows))
	copy(cfgCopy.Workflows, cm.cfg.Workflows)

	return &cfgCopy
}

// GetModelConfig retrieves configuration for a specific model
// It returns the model configuration and a boolean indicating if the model was found
func (cm *DefaultConfigManager) GetModelConfig(modelName string) (ModelConfig, bool) {
	cfg, ok := cm.cfg.Models[modelName]
	return cfg, ok
}

// GetAgentConfig retrieves configuration for a specific agent
// It returns the agent specification and a boolean indicating if the agent was found
func (cm *DefaultConfigManager) GetAgentConfig(agentName string) (AgentSpec, bool) {
	for _, spec := range cm.cfg.Agents {
		if spec.Name == agentName {
			return spec, true
		}
	}
	return AgentSpec{}, false
}

// GetWorkflowConfig retrieves configuration for a specific workflow
// It returns the workflow specification and a boolean indicating if the workflow was found
func (cm *DefaultConfigManager) GetWorkflowConfig(workflowName string) (WorkflowSpec, bool) {
	for _, spec := range cm.cfg.Workflows {
		if spec.Name == workflowName {
			return spec, true
		}
	}
	return WorkflowSpec{}, false
}

// GetModelCacheConfig returns the model cache configuration
// It provides access to settings like cache size, TTL, and cleanup interval
func (cm *DefaultConfigManager) GetModelCacheConfig() ModelCacheSpec {
	return cm.cfg.Memory.ModelCache
}

// GetMemoryMonitorConfig returns the memory monitor configuration
// It provides access to settings like monitoring interval, history size, and alert thresholds
func (cm *DefaultConfigManager) GetMemoryMonitorConfig() MemoryMonitorSpec {
	return cm.cfg.Memory.MemoryMonitor
}

// GetThreadStoreConfig returns the thread store configuration
// It provides access to settings like store type, connection details, and retention policies
func (cm *DefaultConfigManager) GetThreadStoreConfig() ThreadStoreSpec {
	return cm.cfg.ThreadStore
}

// GetAllModelConfigs returns all model configurations
func (cm *DefaultConfigManager) GetAllModelConfigs() map[string]ModelConfig {
	return cm.cfg.Models
}

// GetDefaultModelConfig returns the default model configuration
func (cm *DefaultConfigManager) GetDefaultModelConfig() (ModelConfig, bool) {
	if cm.cfg.DefaultModel == "" {
		return ModelConfig{}, false
	}
	return cm.GetModelConfig(cm.cfg.DefaultModel)
}

// GetModelType retrieves the type of a specific model
// It returns the model type and a boolean indicating if the model was found
func (cm *DefaultConfigManager) GetModelType(modelName string) (string, bool) {
	cfg, found := cm.GetModelConfig(modelName)
	if !found {
		return "", false
	}
	return cfg.Type, true
}

// IsModelEnabled checks if a specific model is enabled
// It returns true if the model is enabled, false otherwise or if the model was not found
func (cm *DefaultConfigManager) IsModelEnabled(modelName string) bool {
	cfg, found := cm.GetModelConfig(modelName)
	if !found {
		return false
	}
	return cfg.Enabled
}

// GetModelPriority retrieves the priority of a specific model
// It returns the model priority and a boolean indicating if the model was found
func (cm *DefaultConfigManager) GetModelPriority(modelName string) (int, bool) {
	cfg, found := cm.GetModelConfig(modelName)
	if !found {
		return 0, false
	}
	return cfg.Priority, true
}

// GetModelOptions retrieves the options for a specific model
// It returns the model options and a boolean indicating if the model was found
func (cm *DefaultConfigManager) GetModelOptions(modelName string) (map[string]any, bool) {
	cfg, found := cm.GetModelConfig(modelName)
	if !found {
		return nil, false
	}
	return cfg.Options, true
}

// SaveConfig saves the current configuration to the original source
// It returns an error if the save fails or validation fails
func (cm *DefaultConfigManager) SaveConfig() error {
	cm.mu.RLock()
	cfg := cm.cfg
	sourcePath := cm.sourcePath
	cm.mu.RUnlock()

	if sourcePath == "" {
		return fmt.Errorf("no configuration source path specified")
	}

	// Validate current configuration
	if err := cm.validateConfig(cfg); err != nil {
		return err
	}

	// Save configuration to file
	return SaveHostConfigFile(sourcePath, cfg)
}

// SaveConfigToFile saves the current configuration to a file
// It returns an error if the save fails or validation fails
func (cm *DefaultConfigManager) SaveConfigToFile(path string) error {
	cm.mu.RLock()
	cfg := cm.cfg
	cm.mu.RUnlock()

	// Validate current configuration
	if err := cm.validateConfig(cfg); err != nil {
		return err
	}

	// Save configuration to file
	return SaveHostConfigFile(path, cfg)
}

// SaveConfigWithConfig saves the provided configuration to the original source
// It returns an error if the save fails or validation fails
func (cm *DefaultConfigManager) SaveConfigWithConfig(cfg *HostConfig) error {
	cm.mu.RLock()
	sourcePath := cm.sourcePath
	cm.mu.RUnlock()

	if sourcePath == "" {
		return fmt.Errorf("no configuration source path specified")
	}

	// Validate the provided configuration
	if err := cm.validateConfig(cfg); err != nil {
		return err
	}

	// Save configuration to file
	if err := SaveHostConfigFile(sourcePath, cfg); err != nil {
		return err
	}

	// Update the current configuration in memory
	return cm.UpdateConfig(cfg)
}
