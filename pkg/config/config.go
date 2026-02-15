// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Unified configuration management package
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for the Agent Framework
type Config struct {
	Agent       AgentConfig       `yaml:"agent" json:"agent"`
	Model       ModelConfig       `yaml:"model" json:"model"`
	Workflow    WorkflowConfig    `yaml:"workflow" json:"workflow"`
	Sandbox     SandboxConfig     `yaml:"sandbox" json:"sandbox"`
	Monitoring  MonitoringConfig  `yaml:"monitoring" json:"monitoring"`
	Server      ServerConfig      `yaml:"server" json:"server"`
}

// AgentConfig represents the agent configuration
type AgentConfig struct {
	DefaultName        string        `yaml:"default_name" json:"default_name"`
	DefaultInstructions string        `yaml:"default_instructions" json:"default_instructions"`
	Memory             MemoryConfig  `yaml:"memory" json:"memory"`
	MaxConcurrent      int           `yaml:"max_concurrent" json:"max_concurrent"`
}

// MemoryConfig represents the memory configuration
type MemoryConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	MaxHistorySize     int           `yaml:"max_history_size" json:"max_history_size"`
	CompressionEnabled bool          `yaml:"compression_enabled" json:"compression_enabled"`
	CompressionRatio   float64       `yaml:"compression_ratio" json:"compression_ratio"`
}

// ModelConfig represents the model configuration
type ModelConfig struct {
	DefaultModel       string        `yaml:"default_model" json:"default_model"`
	Cache              CacheConfig   `yaml:"cache" json:"cache"`
	HealthCheck        HealthConfig  `yaml:"health_check" json:"health_check"`
}

// CacheConfig represents the model cache configuration
type CacheConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	MaxSize            int           `yaml:"max_size" json:"max_size"`
	TTL                time.Duration `yaml:"ttl" json:"ttl"`
	CleanupInterval    time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// HealthConfig represents the model health check configuration
type HealthConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	Interval           time.Duration `yaml:"interval" json:"interval"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	MaxFailures        int           `yaml:"max_failures" json:"max_failures"`
}

// WorkflowConfig represents the workflow configuration
type WorkflowConfig struct {
	MaxConcurrent      int           `yaml:"max_concurrent" json:"max_concurrent"`
	CheckpointInterval time.Duration `yaml:"checkpoint_interval" json:"checkpoint_interval"`
	MaxRetryAttempts   int           `yaml:"max_retry_attempts" json:"max_retry_attempts"`
	RetryDelay         time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// SandboxConfig represents the sandbox configuration
type SandboxConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	AllowedLanguages   []string      `yaml:"allowed_languages" json:"allowed_languages"`
	ResourceQuota      ResourceQuota `yaml:"resource_quota" json:"resource_quota"`
}

// ResourceQuota represents the resource quota configuration
type ResourceQuota struct {
	MaxFileSize        int64         `yaml:"max_file_size" json:"max_file_size"`
	MaxTotalSize       int64         `yaml:"max_total_size" json:"max_total_size"`
	MaxFileCount       int           `yaml:"max_file_count" json:"max_file_count"`
	MaxCPUSeconds      int           `yaml:"max_cpu_seconds" json:"max_cpu_seconds"`
	MaxMemoryBytes     int64         `yaml:"max_memory_bytes" json:"max_memory_bytes"`
	MaxProcessCount    int           `yaml:"max_process_count" json:"max_process_count"`
}

// MonitoringConfig represents the monitoring configuration
type MonitoringConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	MetricsEnabled     bool          `yaml:"metrics_enabled" json:"metrics_enabled"`
	TracesEnabled      bool          `yaml:"traces_enabled" json:"traces_enabled"`
	LogsEnabled        bool          `yaml:"logs_enabled" json:"logs_enabled"`
	PrometheusPort     int           `yaml:"prometheus_port" json:"prometheus_port"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port               int           `yaml:"port" json:"port"`
	Host               string        `yaml:"host" json:"host"`
	EnableCORS         bool          `yaml:"enable_cors" json:"enable_cors"`
	APIVersion         string        `yaml:"api_version" json:"api_version"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = findConfigFile()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := parseConfig(data, configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// findConfigFile tries to find the configuration file in standard locations
func findConfigFile() string {
	// Check current directory
	for _, ext := range []string{".yaml", ".yml", ".json"} {
		path := "config" + ext
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check config directory
	configDir := "config"
	if _, err := os.Stat(configDir); err == nil {
		for _, ext := range []string{".yaml", ".yml", ".json"} {
			path := filepath.Join(configDir, "config"+ext)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".agentframework")
		for _, ext := range []string{".yaml", ".yml", ".json"} {
			path := filepath.Join(configPath, "config"+ext)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// Return default config file
	return "config.yaml"
}

// parseConfig parses configuration data based on file extension
func parseConfig(data []byte, configPath string, cfg *Config) error {
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// Try to parse as YAML first, then JSON
		if err := yaml.Unmarshal(data, cfg); err == nil {
			return nil
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("unsupported config format: %s", ext)
		}
	}

	return nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	var errors []string

	// Validate agent configuration
	if cfg.Agent.MaxConcurrent <= 0 {
		errors = append(errors, "agent.max_concurrent must be greater than 0")
	}

	// Validate model configuration
	if cfg.Model.DefaultModel == "" {
		errors = append(errors, "model.default_model is required")
	}

	// Validate workflow configuration
	if cfg.Workflow.MaxConcurrent <= 0 {
		errors = append(errors, "workflow.max_concurrent must be greater than 0")
	}
	if cfg.Workflow.MaxRetryAttempts < 0 {
		errors = append(errors, "workflow.max_retry_attempts cannot be negative")
	}

	// Validate sandbox configuration
	if cfg.Sandbox.Enabled {
		if cfg.Sandbox.ResourceQuota.MaxFileSize <= 0 {
			errors = append(errors, "sandbox.resource_quota.max_file_size must be greater than 0")
		}
		if cfg.Sandbox.ResourceQuota.MaxTotalSize <= 0 {
			errors = append(errors, "sandbox.resource_quota.max_total_size must be greater than 0")
		}
		if cfg.Sandbox.ResourceQuota.MaxFileCount <= 0 {
			errors = append(errors, "sandbox.resource_quota.max_file_count must be greater than 0")
		}
	}

	// Validate server configuration
	if cfg.Server.Port < 0 || cfg.Server.Port > 65535 {
		errors = append(errors, "server.port must be between 0 and 65535")
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			DefaultName:        "Assistant",
			DefaultInstructions: "You are a helpful AI assistant",
			Memory: MemoryConfig{
				Enabled:            true,
				MaxHistorySize:     10,
				CompressionEnabled: true,
				CompressionRatio:   0.8,
			},
			MaxConcurrent:      5,
		},
		Model: ModelConfig{
			DefaultModel: "gpt-3.5-turbo",
			Cache: CacheConfig{
				Enabled:            true,
				MaxSize:            1000,
				TTL:                30 * time.Minute,
				CleanupInterval:    10 * time.Minute,
			},
			HealthCheck: HealthConfig{
				Enabled:            true,
				Interval:           5 * time.Minute,
				Timeout:            30 * time.Second,
				MaxFailures:        3,
			},
		},
		Workflow: WorkflowConfig{
			MaxConcurrent:      5,
			CheckpointInterval: 30 * time.Second,
			MaxRetryAttempts:   3,
			RetryDelay:         5 * time.Second,
		},
		Sandbox: SandboxConfig{
			Enabled:            true,
			AllowedLanguages:   []string{"python", "javascript", "go", "bash"},
			ResourceQuota: ResourceQuota{
				MaxFileSize:        1024 * 1024, // 1MB
				MaxTotalSize:       10 * 1024 * 1024, // 10MB
				MaxFileCount:       100,
				MaxCPUSeconds:      30,
				MaxMemoryBytes:     256 * 1024 * 1024, // 256MB
				MaxProcessCount:    5,
			},
		},
		Monitoring: MonitoringConfig{
			Enabled:            true,
			MetricsEnabled:     true,
			TracesEnabled:      true,
			LogsEnabled:        true,
			PrometheusPort:     9090,
		},
		Server: ServerConfig{
			Port:               8080,
			Host:               "0.0.0.0",
			EnableCORS:         true,
			APIVersion:         "v1",
		},
	}
}