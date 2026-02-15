// Agent Framework - Configuration Management
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"

	"github.com/mitchellh/mapstructure"
)

// ConfigManager provides unified configuration management
type ConfigManager struct {
	configPath string
	config     map[string]interface{}
	mu         sync.RWMutex
	validators map[string]ConfigSpec
}

// ConfigSpec defines specification for a configuration item
type ConfigSpec struct {
	Name         string `json:"name"`
	Type         string `json:"type"`          // string, int, float64, bool, []string, etc.
	Required     bool   `json:"required"`
	Default      interface{} `json:"default"`
	Validate     Validator `json:"validate"`      // Custom validator function
	Description  string `json:"description"`
	EnvVar       string `json:"envVar"`       // Environment variable name
	Choices      []string `json:"choices,omitempty"` // Valid choices for selection
	Hidden       bool   `json:"hidden,omitempty"`     // Whether to hide from user
	Group        string `json:"group,omitempty"`     // Group name for organization
}

// Validator is a function that validates a configuration value
type Validator func(value interface{}) error

// Built-in validators
var (
	// String validators
	StringNotEmpty Validator = func(value interface{}) error {
		if s, ok := value.(string); !ok || s == "" {
			return fmt.Errorf("value cannot be empty")
		}
		return nil
	}

	StringMinLength = func(minLength int) Validator {
		return func(value interface{}) error {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("value must be a string")
			}
			if len(s) < minLength {
				return fmt.Errorf("value length must be at least %d", minLength)
			}
			return nil
		}
	}

	StringMaxLength = func(maxLength int) Validator {
		return func(value interface{}) error {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("value must be a string")
			}
			if len(s) > maxLength {
				return fmt.Errorf("value length must not exceed %d", maxLength)
			}
			return nil
		}
	}

	StringMatch = func(pattern string) Validator {
		return func(value interface{}) error {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("value must be a string")
			}
			matched, err := filepath.Match(pattern, s)
			if err != nil {
				return fmt.Errorf("value must match pattern %s: %v", pattern, err)
			}
			if !matched {
				return fmt.Errorf("value does not match pattern %s", pattern)
			}
			return nil
		}
	}

	// Number validators
	IntRange = func(min, max int) Validator {
		return func(value interface{}) error {
			var n int64
			var err error
			switch v := value.(type) {
			case int:
				n = int64(v)
			case int8:
				n = int64(v)
			case int16:
				n = int64(v)
			case int32:
				n = int64(v)
			case int64:
				n = v
			case uint:
				n = int64(v)
			case uint8:
				n = int64(v)
			case uint16:
				n = int64(v)
			case uint32:
				n = int64(v)
			case uint64:
				n = int64(v)
			default:
				err = fmt.Errorf("unsupported type: %T", v)
			}
			if err != nil {
				return err
			}
			if n < int64(min) || n > int64(max) {
				return fmt.Errorf("value must be between %d and %d", min, max)
			}
			return nil
		}
	}

	IntGreaterThan = func(threshold int64) Validator {
		return func(value interface{}) error {
			var n int64
			var err error
			switch v := value.(type) {
			case int:
				n = int64(v)
			case int8:
				n = int64(v)
			case int16:
				n = int64(v)
			case int32:
				n = int64(v)
			case int64:
				n = v
			case uint:
				n = int64(v)
			case uint8:
				n = int64(v)
			case uint16:
				n = int64(v)
			case uint32:
				n = int64(v)
			case uint64:
				n = int64(v)
			default:
				err = fmt.Errorf("unsupported type: %T", v)
			}
			if err != nil {
				return err
			}
			if n <= threshold {
				return fmt.Errorf("value must be greater than %d", threshold)
			}
			return nil
		}
	}

	// Slice validators
	SliceMinLength = func(minLength int) Validator {
		return func(value interface{}) error {
			rv := reflect.ValueOf(value)
			if rv.Kind() != reflect.Slice {
				return fmt.Errorf("value must be a slice")
			}
			if rv.Len() < minLength {
				return fmt.Errorf("slice length must be at least %d", minLength)
			}
			return nil
		}
	}

	SliceMaxLength = func(maxLength int) Validator {
		return func(value interface{}) error {
			rv := reflect.ValueOf(value)
			if rv.Kind() != reflect.Slice {
				return fmt.Errorf("value must be a slice")
			}
			if rv.Len() > maxLength {
				return fmt.Errorf("slice length must not exceed %d", maxLength)
			}
			return nil
		}
	}

	// Map validators
	MapKeyExists = func(keys []string) Validator {
		return func(value interface{}) error {
			rv := reflect.ValueOf(value)
			if rv.Kind() != reflect.Map {
				return fmt.Errorf("value must be a map")
			}
			m, ok := rv.Interface().(map[string]interface{})
			if !ok {
				return fmt.Errorf("value must be a map[string]interface{}")
			}
			for _, key := range keys {
				if _, exists := m[key]; !exists {
					return fmt.Errorf("required key %s is missing", key)
				}
			}
			return nil
		}
	}
)

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		config:     make(map[string]interface{}),
		validators: make(map[string]ConfigSpec),
		mu:         sync.RWMutex{},
	}
}

// Load loads configuration from file
func (m *ConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.configPath)
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, &m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save saves configuration to file
func (m *ConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get retrieves a configuration value
func (m *ConfigManager) Get(key string, defaultValue interface{}) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if value, exists := m.config[key]; exists {
		return value
	}

	return defaultValue
}

// Set sets a configuration value
func (m *ConfigManager) Set(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config[key] = value
	return nil
}

// GetGroup retrieves all configurations in a group
func (m *ConfigManager) GetGroup(group string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})

	for key, value := range m.config {
		if spec, ok := m.validators[key]; ok {
			if spec.Group == group {
				result[key] = value
			}
		}
	}

	return result
}

// Validate validates a value against its specification
func (m *ConfigManager) Validate(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	spec, exists := m.validators[key]
	if !exists {
		return fmt.Errorf("unknown config key: %s", key)
	}

	value, valueExists := m.config[key]
	if !valueExists {
		value = spec.Default
	}

	// Run custom validator if provided
	if spec.Validate != nil {
		if err := spec.Validate(value); err != nil {
			return err
		}
	}

	return nil
}

// UseEnvVar loads value from environment variable
func (m *ConfigManager) UseEnvVar(key string) (string, error) {
	m.mu.RLock()
	spec, exists := m.validators[key]
	m.mu.RUnlock()

	if !exists {
		if val := m.Get(key, ""); val != nil {
			if str, ok := val.(string); ok {
				return str, nil
			}
		}
		return "", fmt.Errorf("config value for key %s not set", key)
	}

	if spec.EnvVar == "" {
		if val := m.Get(key, ""); val != nil {
			if str, ok := val.(string); ok {
				return str, nil
			}
		}
		return "", fmt.Errorf("config value for key %s is not a string", key)
	}

	envValue := os.Getenv(spec.EnvVar)
	if envValue == "" {
		return "", fmt.Errorf("environment variable %s not set", spec.EnvVar)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	switch spec.Type {
	case "string":
		m.config[key] = envValue
		return envValue, nil
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		var n int64
		var err error
		switch spec.Type {
		case "int":
			n, err = strconv.ParseInt(envValue, 10, 64)
		case "int8":
			var val int64
			val, err = strconv.ParseInt(envValue, 10, 8)
			n = val
		case "int16":
			var val int64
			val, err = strconv.ParseInt(envValue, 10, 16)
			n = val
		case "int32":
			var val int64
			val, err = strconv.ParseInt(envValue, 10, 32)
			n = val
		case "int64":
			n, err = strconv.ParseInt(envValue, 10, 64)
		case "uint":
			var val uint64
			val, err = strconv.ParseUint(envValue, 10, 64)
			n = int64(val)
		case "uint8":
			var val uint64
			val, err = strconv.ParseUint(envValue, 10, 8)
			n = int64(val)
		case "uint16":
			var val uint64
			val, err = strconv.ParseUint(envValue, 10, 16)
			n = int64(val)
		case "uint32":
			var val uint64
			val, err = strconv.ParseUint(envValue, 10, 32)
			n = int64(val)
		case "uint64":
			var val uint64
			val, err = strconv.ParseUint(envValue, 10, 64)
			n = int64(val)
		}
		if err != nil {
			return "", fmt.Errorf("invalid int value: %v", envValue)
		}
		m.config[key] = n
		return envValue, nil
	case "bool":
		boolVal := envValue == "true" || envValue == "1" || envValue == "yes"
		m.config[key] = boolVal
		return envValue, nil
	default:
		return "", fmt.Errorf("unsupported env var type: %s", spec.Type)
	}
}

// GetValidator retrieves or creates a validator
func (m *ConfigManager) GetValidator(key string) (Validator, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	spec, exists := m.validators[key]
	if !exists {
		return nil, false
	}
	return spec.Validate, true
}

// RegisterValidator registers a new validator
func (m *ConfigManager) RegisterValidator(key string, spec ConfigSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.validators[key]; exists {
		return fmt.Errorf("validator %s already registered", key)
	}

	m.validators[key] = spec
	return nil
}

// GetAllSpecs returns all configuration specifications
func (m *ConfigManager) GetAllSpecs() map[string]ConfigSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.validators
}

// GetConfigMap returns entire configuration map
func (m *ConfigManager) GetConfigMap() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid external modification
	resultCopy := make(map[string]interface{})
	for k, v := range m.config {
		resultCopy[k] = v
	}

	return resultCopy
}

// RegisterSpec registers a configuration specification
func (m *ConfigManager) RegisterSpec(spec ConfigSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.validators[spec.Name]; exists {
		return fmt.Errorf("config spec %s already registered", spec.Name)
	}

	m.validators[spec.Name] = spec

	// Set default value if specified
	if spec.Default != nil {
		if _, exists := m.config[spec.Name]; !exists {
			m.config[spec.Name] = spec.Default
		}
	}

	return nil
}

// LoadFromMap loads configuration from a map
func (m *ConfigManager) LoadFromMap(data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use mapstructure for advanced mapping
	result := make(map[string]interface{})
	if err := mapstructure.WeakDecode(data, &result); err != nil {
		return fmt.Errorf("failed to decode config map: %w", err)
	}

	for k, v := range result {
		m.config[k] = v
	}

	return nil
}
