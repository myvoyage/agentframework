// Agent Framework - Unified Configuration System
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"sync"
	"time"

	"AgentFramework/agent/errors"
)

// ConfigSource defines where configuration comes from
type ConfigSource string

const (
	ConfigSourceDefaults ConfigSource = "defaults"
	ConfigSourceFile     ConfigSource = "file"
	ConfigSourceEnv      ConfigSource = "env"
	ConfigSourceCLIAPIs ConfigSource = "cli"
	ConfigSourceRemote  ConfigSource = "remote"
)

// ConfigLayer represents a configuration layer with priority
type ConfigLayer struct {
	mu         sync.RWMutex
	Type       ConfigSource `json:"type"`
	Priority   int          `json:"priority"`
	Source     string       `json:"source"`
	Mutable    bool         `json:"mutable"`
	Timestamp  time.Time    `json:"timestamp"`
	data       map[string]interface{}
}

// ConfigChange represents a single configuration change
type ConfigChange struct {
	Timestamp time.Time              `json:"timestamp"`
	Key        string                 `json:"key"`
	OldValue   interface{}             `json:"old_value,omitempty"`
	NewValue   interface{}             `json:"new_value,omitempty"`
	Layer      ConfigSource        `json:"layer"`
	Version    int                    `json:"version"`
	Initiator  string                 `json:"initiator,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ConfigWatcher is notified when configuration changes
type ConfigWatcher func(change ConfigChange) error

// ConfigValidatorFunc validates a configuration value
type ConfigValidatorFunc func(key string, value interface{}) error

// ConfigEncryptor encrypts sensitive configuration values
type ConfigEncryptor func(key string, value interface{}) (string, error)

// ConfigDecryptor decrypts sensitive configuration values
type ConfigDecryptor func(key string, encrypted string) (interface{}, error)

// NewUnifiedConfigManager creates a new unified configuration manager
func NewUnifiedConfigManager() *UnifiedConfigManager {
	mgr := &UnifiedConfigManager{
		layers:      make([]*ConfigLayer, 0, 8),
		changeLog:    make([]ConfigChange, 0, 1000),
		validators:   make(map[string]ConfigValidatorFunc),
		version:      1,
	}

	// Add default layer with lowest priority
	defaultLayer := &ConfigLayer{
		Type:     ConfigSourceDefaults,
		Priority: 0,
		Source:    "defaults",
		Mutable:   false,
		data:      make(map[string]interface{}),
	}
	mgr.AddLayer(defaultLayer)

	return mgr
}

// UnifiedConfigManager provides unified configuration management
type UnifiedConfigManager struct {
	layers      []*ConfigLayer
	mu          sync.RWMutex
	changeLog    []ConfigChange
	validators   map[string]ConfigValidatorFunc
	version      int
	watchers    []ConfigWatcher
	encryptor    ConfigEncryptor
	decryptor    ConfigDecryptor
}

// AddLayer adds a new configuration layer
func (m *UnifiedConfigManager) AddLayer(layer *ConfigLayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize data map if nil
	if layer.data == nil {
		layer.data = make(map[string]interface{})
	}
	layer.Timestamp = time.Now()

	// Insert in priority order (highest first)
	inserted := false
	for i, existing := range m.layers {
		if layer.Priority > existing.Priority {
			// Insert before this layer
			newLayers := append([]*ConfigLayer{layer}, m.layers[i:]...)
			m.layers = append(m.layers[:i], newLayers...)
			inserted = true
			break
		}
	}


	if !inserted {
		// Lowest priority, append at end
		m.layers = append(m.layers, layer)
	}

	return nil
}

// Get retrieves a configuration value by key
// Returns the value from the highest priority layer that has it
func (m *UnifiedConfigManager) Get(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Search layers from highest to lowest priority
	for _, layer := range m.layers {
		layer.mu.RLock()
		if value, exists := layer.data[key]; exists {
			layer.mu.RUnlock()
			return value
		}
		layer.mu.RUnlock()
	}

	return nil
}

// Set sets a configuration value
func (m *UnifiedConfigManager) Set(key string, value interface{}) error {
	return m.SetWithInitiator(key, value, "system")
}

// SetWithInitiator sets a configuration value with a custom initiator label
func (m *UnifiedConfigManager) SetWithInitiator(key string, value interface{}, initiator string) error {
	// Validate value if validator registered
	if err := m.validate(key, value); err != nil {
		return errors.Wrapf(err, errors.ErrCodeInternal, "config validation failed for %s: %v", key, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find highest priority mutable layer
	var targetLayer *ConfigLayer
	for _, layer := range m.layers {
		if layer.Mutable {
			targetLayer = layer
			break
		}
	}

	if targetLayer == nil {
		return fmt.Errorf("no mutable layer found")
	}

	// Get old value
	oldValue := targetLayer.data[key]

	// Set new value
	targetLayer.data[key] = value
	targetLayer.Timestamp = time.Now()

	// Record change
	change := ConfigChange{
		Timestamp: time.Now(),
		Key:        key,
		OldValue:   oldValue,
		NewValue:   value,
		Layer:      targetLayer.Type,
		Version:    m.version,
		Initiator:  initiator,
	}

	if oldValue != nil {
		change.Description = fmt.Sprintf("Configuration key %s was updated", key)
	} else {
		change.Description = fmt.Sprintf("Configuration key %s was added", key)
	}

	m.changeLog = append(m.changeLog, change)

	// Notify watchers
	for _, watcher := range m.watchers {
		if err := watcher(change); err != nil {
			// Log error but continue with other watchers
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}

	return nil
}

// validate validates a configuration value
func (m *UnifiedConfigManager) validate(key string, value interface{}) error {
	validator, exists := m.validators[key]
	if !exists {
		return nil
	}
	return validator(key, value)
}

// Export exports current configuration to a map
func (m *UnifiedConfigManager) Export() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})
	for _, layer := range m.layers {
		layer.mu.RLock()
		for k, v := range layer.data {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}
		layer.mu.RUnlock()
	}

	return result
}

// RegisterValidator registers a configuration validator
func (m *UnifiedConfigManager) RegisterValidator(key string, validator ConfigValidatorFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validators[key] = validator
	return nil
}

// AddWatcher adds a configuration watcher
func (m *UnifiedConfigManager) AddWatcher(watcher ConfigWatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchers = append(m.watchers, watcher)
}

// SetEncryptor sets the configuration encryptor
func (m *UnifiedConfigManager) SetEncryptor(encryptor ConfigEncryptor) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.encryptor = encryptor
}

// SetDecryptor sets the configuration decryptor
func (m *UnifiedConfigManager) SetDecryptor(decryptor ConfigDecryptor) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.decryptor = decryptor
}

// GetChanges returns configuration change log
func (m *UnifiedConfigManager) GetChanges() []ConfigChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid external modification
	changes := make([]ConfigChange, len(m.changeLog))
	copy(changes, m.changeLog)
	return changes
}
