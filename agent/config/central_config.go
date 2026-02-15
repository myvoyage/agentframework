// Agent Framework - Central Configuration
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// CentralConfig provides centralized configuration management
type CentralConfig struct {
	mu    sync.RWMutex
	data   map[string]interface{}
	path   string
}

var globalCentralConfig *CentralConfig

// InitCentralConfig initializes central configuration
func InitCentralConfig(configPath string) error {
	globalCentralConfig = &CentralConfig{
		mu:   sync.RWMutex{},
		data:  make(map[string]interface{}),
		path: configPath,
	}

	// Load configuration
	if err := globalCentralConfig.Load(configPath); err != nil {
		return err
	}
	return nil
}

// Load loads configuration from a JSON file
func (c *CentralConfig) Load(configPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist yet, that's ok
		return nil
	}

	return json.Unmarshal(data, &c.data)
}

// Save saves configuration to a JSON file
func (c *CentralConfig) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}

// Get retrieves a configuration value
func (c *CentralConfig) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, exists := c.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

// Set sets a configuration value
func (c *CentralConfig) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	return nil
}

// GetConfig retrieves a configuration value
func (c *CentralConfig) GetConfig(key string) (interface{}, error) {
	return globalCentralConfig.Get(key)
}

// SetConfig sets a configuration value
func (c *CentralConfig) SetConfig(key string, value interface{}) error {
	err := globalCentralConfig.Set(key, value)
	if err != nil {
		return err
	}
	// Auto-save on set
	return globalCentralConfig.Save()
}

// GetInt retrieves an integer configuration value
func (c *CentralConfig) GetInt(key string, defaultValue int64) int64 {
	val, err := c.GetConfig(key)
	if err == nil && val != nil {
		if intVal, ok := val.(int64); ok {
			return intVal
		}
	}
	return defaultValue
}

// GetString retrieves a string configuration value
func (c *CentralConfig) GetString(key string, defaultValue string) string {
	val, err := c.GetConfig(key)
	if err == nil && val != nil {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetBool retrieves a boolean configuration value
func (c *CentralConfig) GetBool(key string, defaultValue bool) bool {
	val, err := c.GetConfig(key)
	if err == nil && val != nil {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// GetDuration retrieves a duration configuration value
func (c *CentralConfig) GetDuration(key string, defaultValue time.Duration) time.Duration {
	val, err := c.GetConfig(key)
	if err == nil && val != nil {
		if durVal, ok := val.(time.Duration); ok {
			return durVal
		}
	}
	return defaultValue
}

// MustGetInt is like GetInt but panics if value is missing
func (c *CentralConfig) MustGetInt(key string, defaultValue int64) int64 {
	val, err := c.GetConfig(key)
	if err != nil {
		panic(fmt.Sprintf("required config key missing: %s", key))
	}
	if val == nil {
		panic(fmt.Sprintf("required config key not set: %s", key))
	}
	intVal, ok := val.(int64)
	if !ok {
		panic(fmt.Sprintf("config value for key %s is not an integer", key))
	}
	return intVal
}

// MustGetString is like GetString but panics if value is missing
func (c *CentralConfig) MustGetString(key string, defaultValue string) string {
	val, err := c.GetConfig(key)
	if err != nil {
		panic(fmt.Sprintf("required config key missing: %s", key))
	}
	if val == nil {
		panic(fmt.Sprintf("required config key not set: %s", key))
	}
	strVal, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("config value for key %s is not a string", key))
	}
	return strVal
}

// GetGlobalCentralConfig returns the global central config instance
func GetGlobalCentralConfig() *CentralConfig {
	if globalCentralConfig == nil {
		// Initialize with default path if not initialized
		_ = InitCentralConfig("config.json")
	}
	return globalCentralConfig
}
