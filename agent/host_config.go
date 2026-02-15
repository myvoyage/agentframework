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

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type ThreadStoreSpec struct {
	Type        string `yaml:"type"`                  // memory|file|redis|sql
	Dir         string `yaml:"dir,omitempty"`         // for file
	RedisAddr   string `yaml:"redisAddr,omitempty"`   // for redis
	RedisPrefix string `yaml:"redisPrefix,omitempty"` // for redis
	DriverName  string `yaml:"driver,omitempty"`      // for sql
	DSN         string `yaml:"dsn,omitempty"`         // for sql
	Table       string `yaml:"table,omitempty"`       // for sql
	// Memory-specific configuration
	MaxMessages    int `yaml:"maxMessages,omitempty"`    // Maximum number of messages to keep per thread
	MaxMessageSize int `yaml:"maxMessageSize,omitempty"` // Maximum size of a single message in bytes
	TTL            int `yaml:"ttl,omitempty"`            // Time-to-live for threads in seconds
}

type AgentSpec struct {
	Name         string   `yaml:"name"`
	Kind         string   `yaml:"kind"`                   // chat|react
	Model        string   `yaml:"model"`                  // model key
	Instructions string   `yaml:"instructions,omitempty"` // system prompt
	Tools        []string `yaml:"tools,omitempty"`        // tool names
	Middlewares  []string `yaml:"middlewares,omitempty"`  // middleware keys
	// Memory management options
	MaxMessages    int     `yaml:"maxMessages,omitempty"`    // Maximum number of messages to keep in the thread history
	MaxMessageSize int     `yaml:"maxMessageSize,omitempty"` // Maximum size of a single message in bytes
	TrimRatio      float64 `yaml:"trimRatio,omitempty"`      // Ratio of messages to keep when trimming (0.0-1.0)
	EnableTrimming bool    `yaml:"enableTrimming,omitempty"` // Enable intelligent message trimming
}

// NodeSpec represents a node in a DAG workflow
type NodeSpec struct {
	ID        string `yaml:"id"`
	Kind      string `yaml:"kind,omitempty"`  // agent|inline
	AgentName string `yaml:"agent,omitempty"` // Reference to defined agent (for kind=agent)
	// Inline agent configuration (for kind=inline)
	InlineName         string   `yaml:"inlineName,omitempty"`         // Name of the inline agent
	InlineKind         string   `yaml:"inlineKind,omitempty"`         // chat|react
	InlineModel        string   `yaml:"inlineModel,omitempty"`        // model key
	InlineInstructions string   `yaml:"inlineInstructions,omitempty"` // system prompt
	InlineTools        []string `yaml:"inlineTools,omitempty"`        // tool names
	InlineMiddlewares  []string `yaml:"inlineMiddlewares,omitempty"`  // middleware keys
}

type WorkflowSpec struct {
	Name       string            `yaml:"name"`
	Kind       string            `yaml:"kind"`                 // sequential|aggregating_parallel|routing|dag
	Model      string            `yaml:"model,omitempty"`      // for routing
	Steps      []string          `yaml:"steps,omitempty"`      // for sequential
	Agents     []string          `yaml:"agents,omitempty"`     // for parallel / aggregating
	Aggregator string            `yaml:"aggregator,omitempty"` // aggregating workflow
	Routes     map[string]string `yaml:"routes,omitempty"`     // for routing: key -> workflow name
	Nodes      []NodeSpec        `yaml:"nodes,omitempty"`      // for DAG
	Edges      map[string]string `yaml:"edges,omitempty"`      // for DAG: from -> to (simple) or from -> [to1, to2]
	// Extended Edges support: from -> [to...]
	EdgesList map[string][]string `yaml:"edgesList,omitempty"` // explicit list support
}

// ModelCacheSpec represents model cache configuration
type ModelCacheSpec struct {
	Enabled         bool `yaml:"enabled,omitempty"`         // Whether model caching is enabled
	MaxSize         int  `yaml:"maxSize,omitempty"`         // Maximum number of models to cache
	TTL             int `yaml:"ttl,omitempty"`             // Time-to-live for cached models in seconds
	CleanupInterval int `yaml:"cleanupInterval,omitempty"` // Cleanup interval in seconds
}

// MemoryMonitorSpec represents memory monitor configuration
type MemoryMonitorSpec struct {
	Enabled        bool `yaml:"enabled,omitempty"`        // Whether memory monitoring is enabled
	Interval       int `yaml:"interval,omitempty"`       // Interval between measurements in seconds
	HistorySize    int `yaml:"historySize,omitempty"`    // Number of historical measurements to keep
	AlertThreshold int `yaml:"alertThreshold,omitempty"` // Memory usage threshold for alerts in MB
	AlertInterval  int `yaml:"alertInterval,omitempty"`  // Minimum interval between alerts in seconds
}

// MemoryManagementSpec represents memory management configuration
type MemoryManagementSpec struct {
	ModelCache    ModelCacheSpec    `yaml:"modelCache,omitempty"`    // Model cache configuration
	MemoryMonitor MemoryMonitorSpec `yaml:"memoryMonitor,omitempty"` // Memory monitor configuration
	Checkpoint    struct {
		MaxCheckpoints  int `yaml:"maxCheckpoints,omitempty"`  // Maximum number of checkpoints to keep
		TTL             int `yaml:"ttl,omitempty"`             // Time-to-live for checkpoints in seconds
		CleanupInterval int `yaml:"cleanupInterval,omitempty"` // Cleanup interval in seconds
	} `yaml:"checkpoint,omitempty"` // Checkpoint store configuration
}

// ContextStoreSpec represents context storage configuration
type ContextStoreSpec struct {
	Enabled   bool                `yaml:"enabled,omitempty"`   // Whether context storage is enabled
	Type      string              `yaml:"type,omitempty"`      // Context store type: openviking|memory|none
	OpenViking OpenVikingStoreSpec `yaml:"openviking,omitempty"` // OpenViking specific configuration
}

// OpenVikingStoreSpec represents OpenViking context store configuration
type OpenVikingStoreSpec struct {
	Endpoint   string `yaml:"endpoint,omitempty"`   // OpenViking server endpoint
	APIKey     string `yaml:"apiKey,omitempty"`     // API key for authentication
	Workspace  string `yaml:"workspace,omitempty"`  // Default workspace path
	Timeout    int `yaml:"timeout,omitempty"`    // Request timeout in seconds
	MaxRetries int `yaml:"maxRetries,omitempty"` // Maximum number of retries
	AutoSync   bool `yaml:"autoSync,omitempty"`   // Enable automatic context synchronization
}

// TokenCompressionSpec represents token compression configuration
type TokenCompressionSpec struct {
	Enabled             bool   `yaml:"enabled,omitempty"`             // Whether token compression is enabled
	Strategy            string `yaml:"strategy,omitempty"`            // Compression strategy: truncate|summarize|semantic|hybrid
	TargetTokens        int    `yaml:"targetTokens,omitempty"`        // Target token count after compression
	MinTokens           int    `yaml:"minTokens,omitempty"`           // Minimum token count to consider compression
	MaxTokens           int    `yaml:"maxTokens,omitempty"`           // Maximum token count to trigger compression
	PreserveSystemMessages bool `yaml:"preserveSystemMessages,omitempty"` // Whether to preserve system messages
	SummaryModelName    string `yaml:"summaryModelName,omitempty"`    // Model to use for summarization
	SummaryMaxTokens    int `yaml:"summaryMaxTokens,omitempty"`    // Maximum tokens for summaries
	Temperature         float64 `yaml:"temperature,omitempty"`         // Temperature for summary generation
	CheckInterval       int `yaml:"checkInterval,omitempty"`       // Check interval in seconds
	// Memory-based configuration
	SessionTokenLimit   int `yaml:"sessionTokenLimit,omitempty"`   // Token limit for session memory
	DailyTokenLimit     int `yaml:"dailyTokenLimit,omitempty"`     // Token limit for daily memory
	LongTermTokenLimit  int `yaml:"longTermTokenLimit,omitempty"`  // Token limit for long-term memory
}

// AsyncTaskSpec represents async task configuration
type AsyncTaskSpec struct {
	Enabled     bool `yaml:"enabled,omitempty"`     // Whether async task execution is enabled
	MaxTasks    int `yaml:"maxTasks,omitempty"`    // Maximum concurrent tasks
	TaskTimeout int `yaml:"taskTimeout,omitempty"` // Default task timeout in seconds
}

// SchedulerSpec represents scheduler configuration
type SchedulerSpec struct {
	Enabled   bool  `yaml:"enabled,omitempty"`   // Whether scheduler is enabled
	Timezone  string `yaml:"timezone,omitempty"`  // Timezone for cron jobs
	MaxJobs   int `yaml:"maxJobs,omitempty"`   // Maximum concurrent jobs
	JobTimeout int `yaml:"jobTimeout,omitempty"` // Job timeout in seconds
}

// HeartbeatSpec represents heartbeat service configuration
type HeartbeatSpec struct {
	Enabled  bool `yaml:"enabled,omitempty"`  // Whether heartbeat is enabled
	Interval int `yaml:"interval,omitempty"` // Heartbeat interval in seconds
	Timeout  int `yaml:"timeout,omitempty"`  // Heartbeat timeout in seconds
}

type HostConfig struct {
	Name           string                 `yaml:"name,omitempty"`           // application name
	Version        string                 `yaml:"version,omitempty"`        // application version
	DefaultModel   string                 `yaml:"defaultModel"`             // default model key
	Models         map[string]ModelConfig `yaml:"models,omitempty"`         // model configurations
	ThreadStore    ThreadStoreSpec        `yaml:"threadStore"`              // thread store config
	Memory         MemoryManagementSpec   `yaml:"memory,omitempty"`         // memory management config
	Agents         []AgentSpec            `yaml:"agents"`                   // agent declarations
	Workflows      []WorkflowSpec         `yaml:"workflows,omitempty"`      // workflow declarations
	SkillSystemDir string                 `yaml:"skillSystemDir,omitempty"` // skill system directory
	ContextStore   ContextStoreSpec       `yaml:"contextStore,omitempty"`  // context storage configuration
	Messaging      *MessagingConfig       `yaml:"messaging,omitempty"`      // messaging configuration
	Scheduler      *SchedulerSpec         `yaml:"scheduler,omitempty"`      // scheduler configuration
	Heartbeat      *HeartbeatSpec         `yaml:"heartbeat,omitempty"`      // heartbeat configuration
	AsyncTask      *AsyncTaskSpec         `yaml:"asyncTask,omitempty"`      // async task configuration
	TokenCompression *TokenCompressionSpec `yaml:"tokenCompression,omitempty"` // token compression configuration
	Extensions     map[string]interface{} `yaml:"-"`                        // extensions for plugins (not serialized)
}

func LoadHostConfig(r io.Reader) (*HostConfig, error) {
	// Read all content to apply env expansion
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Expand environment variables: ${VAR} or $VAR
	expanded := os.ExpandEnv(string(content))

	var cfg HostConfig
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadHostConfigFile(path string) (*HostConfig, error) {
	// os.Open is replaced by os.ReadFile to support expansion easily
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadHostConfig(f)
}

// SaveHostConfig saves the host configuration to a writer
func SaveHostConfig(w io.Writer, cfg *HostConfig) error {
	// Marshal the configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write the YAML data to the writer
	_, err = w.Write(data)
	return err
}

// SaveHostConfigFile saves the host configuration to a file
func SaveHostConfigFile(path string, cfg *HostConfig) error {
	// Create the file for writing
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Save the configuration to the file
	return SaveHostConfig(f, cfg)
}