// Package channels provides configuration management for multi-channel system
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigFormat represents the configuration file format
type ConfigFormat string

const (
	// ConfigFormatJSON represents JSON format
	ConfigFormatJSON ConfigFormat = "json"
	// ConfigFormatYAML represents YAML format
	ConfigFormatYAML ConfigFormat = "yaml"
)

// Config represents the complete channel configuration
type Config struct {
	// Version is the config version
	Version string `json:"version" yaml:"version"`

	// Global settings
	Global GlobalConfig `json:"global" yaml:"global"`

	// Channels configuration
	Channels map[string]ChannelConfig `json:"channels" yaml:"channels"`

	// Routing rules
	Routes []*RoutingRule `json:"routes,omitempty" yaml:"routes,omitempty"`

	// File modification time for hot reload
	modTime time.Time
	mu      sync.RWMutex
}

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
	// Default timeout for operations
	DefaultTimeout time.Duration `json:"default_timeout" yaml:"default_timeout"`

	// Enable metrics collection
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`

	// Enable OpenTelemetry tracing
	EnableTracing bool `json:"enable_tracing" yaml:"enable_tracing"`

	// Log level
	LogLevel string `json:"log_level" yaml:"log_level"`

	// Max message size (in bytes)
	MaxMessageSize int `json:"max_message_size" yaml:"max_message_size"`

	// Default rate limiting
	DefaultRateLimit int `json:"default_rate_limit" yaml:"default_rate_limit"`
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Detect format from extension
	var format ConfigFormat
	if len(path) > 5 {
		switch path[len(path)-5:] {
		case ".json":
			format = ConfigFormatJSON
		case ".yaml", ".yml":
			format = ConfigFormatYAML
		default:
			format = ConfigFormatYAML // Default to YAML
		}
	}

	config, err := ParseConfig(data, format)
	if err != nil {
		return nil, err
	}

	// Get file modification time for hot reload
	info, err := os.Stat(path)
	if err == nil {
		config.mu.Lock()
		config.modTime = info.ModTime()
		config.mu.Unlock()
	}

	return config, nil
}

// ParseConfig parses configuration from data
func ParseConfig(data []byte, format ConfigFormat) (*Config, error) {
	config := &Config{
		Channels: make(map[string]ChannelConfig),
	}

	var err error
	switch format {
	case ConfigFormatJSON:
		err = json.Unmarshal(data, config)
	case ConfigFormatYAML:
		err = yaml.Unmarshal(data, config)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, path string, format ConfigFormat) error {
	config.mu.Lock()
	defer config.mu.Unlock()

	var data []byte
	var err error

	switch format {
	case ConfigFormatJSON:
		data, err = json.MarshalIndent(config, "", "  ")
	case ConfigFormatYAML:
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate global config
	if c.Global.DefaultTimeout == 0 {
		c.Global.DefaultTimeout = 30 * time.Second
	}

	if c.Global.LogLevel == "" {
		c.Global.LogLevel = "info"
	}

	// Validate each channel
	for name, ch := range c.Channels {
		if err := ch.Validate(); err != nil {
			return fmt.Errorf("channel %s: %w", name, err)
		}
	}

	// Validate routing rules
	for i, rule := range c.Routes {
		if rule.ID == "" {
			return fmt.Errorf("route %d: missing ID", i)
		}
		if rule.Priority == 0 {
			rule.Priority = 100 // Default priority
		}
	}

	return nil
}

// GetChannelConfig returns configuration for a specific channel
func (c *Config) GetChannelConfig(name string) (ChannelConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ch, ok := c.Channels[name]
	return ch, ok
}

// SetChannelConfig sets configuration for a specific channel
func (c *Config) SetChannelConfig(name string, config ChannelConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Channels == nil {
		c.Channels = make(map[string]ChannelConfig)
	}

	c.Channels[name] = config
}

// RemoveChannel removes a channel configuration
func (c *Config) RemoveChannel(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Channels, name)
}

// ListChannels returns all channel names
func (c *Config) ListChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.Channels))
	for name := range c.Channels {
		names = append(names, name)
	}

	return names
}

// GetEnabledChannels returns enabled channel configurations
func (c *Config) GetEnabledChannels() map[string]ChannelConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	enabled := make(map[string]ChannelConfig)
	for name, ch := range c.Channels {
		if ch.Enabled {
			enabled[name] = ch
		}
	}

	return enabled
}

// AddRoutingRule adds a routing rule
func (c *Config) AddRoutingRule(rule *RoutingRule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Routes == nil {
		c.Routes = make([]*RoutingRule, 0)
	}

	c.Routes = append(c.Routes, rule)
}

// RemoveRoutingRule removes a routing rule by ID
func (c *Config) RemoveRoutingRule(ruleID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Routes == nil {
		return
	}

	for i, rule := range c.Routes {
		if rule.ID == ruleID {
			c.Routes = append(c.Routes[:i], c.Routes[i+1:]...)
			return
		}
	}
}

// GetRoutingRules returns all routing rules
func (c *Config) GetRoutingRules() []*RoutingRule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Routes == nil {
		return make([]*RoutingRule, 0)
	}

	rules := make([]*RoutingRule, len(c.Routes))
	copy(rules, c.Routes)
	return rules
}

// HotReload checks if the config file has been modified and reloads if needed
func (c *Config) HotReload(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	c.mu.RLock()
	modTime := c.modTime
	c.mu.RUnlock()

	if info.ModTime().After(modTime) {
		// File has been modified, reload
		newConfig, err := LoadConfig(path)
		if err != nil {
			return false, err
		}

		// Update current config
		c.mu.Lock()
		c.Version = newConfig.Version
		c.Global = newConfig.Global
		c.Channels = newConfig.Channels
		c.Routes = newConfig.Routes
		c.modTime = info.ModTime()
		c.mu.Unlock()

		return true, nil
	}

	return false, nil
}

// ConfigWatcher watches for configuration changes
type ConfigWatcher struct {
	config   *Config
	path     string
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	onChange func(*Config) error
	mu       sync.Mutex
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(config *Config, path string, interval time.Duration) *ConfigWatcher {
	return &ConfigWatcher{
		config:   config,
		path:     path,
		interval: interval,
		onChange: func(*Config) error { return nil },
	}
}

// Start starts the configuration watcher
func (w *ConfigWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ctx, w.cancel = context.WithCancel(ctx)

	go w.watch()

	return nil
}

// Stop stops the configuration watcher
func (w *ConfigWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cancel != nil {
		w.cancel()
	}
}

// SetOnChange sets the callback for configuration changes
func (w *ConfigWatcher) SetOnChange(callback func(*Config) error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.onChange = callback
}

// watch watches for configuration changes
func (w *ConfigWatcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			reloaded, err := w.config.HotReload(w.path)
			if err != nil {
				// Log error but continue watching
				continue
			}

			if reloaded {
				// Call onChange callback
				_ = w.onChange(w.config)
			}
		}
	}
}

// LoadConfigFromEnv loads configuration from environment variables
// Useful for containerized deployments
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		Version: "1.0",
		Global: GlobalConfig{
			DefaultTimeout: 30 * time.Second,
			LogLevel:       "info",
		},
		Channels: make(map[string]ChannelConfig),
	}

	// Load Telegram config if token is provided
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		config.Channels["telegram"] = ChannelConfig{
			Type:    ChannelTypeTelegram,
			Enabled: true,
			Token:   token,
		}
	}

	// Load Discord config if token is provided
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		config.Channels["discord"] = ChannelConfig{
			Type:    ChannelTypeDiscord,
			Enabled: true,
			Token:   token,
		}
	}

	// Load Slack config if tokens are provided
	if token := os.Getenv("SLACK_BOT_TOKEN"); token != "" {
		config.Channels["slack"] = ChannelConfig{
			Type:     ChannelTypeSlack,
			Enabled:  true,
			Token:    token,
			AppToken: os.Getenv("SLACK_APP_TOKEN"),
		}
	}

	// Load Feishu config if credentials are provided
	if appID := os.Getenv("FEISHU_APP_ID"); appID != "" {
		config.Channels["feishu"] = ChannelConfig{
			Type:      ChannelTypeFeishu,
			Enabled:   true,
			AppID:     appID,
			AppSecret: os.Getenv("FEISHU_APP_SECRET"),
		}
	}

	// Load WeWork config if credentials are provided
	if corpID := os.Getenv("WEWORK_CORP_ID"); corpID != "" {
		config.Channels["wework"] = ChannelConfig{
			Type:    ChannelTypeWeWork,
			Enabled: true,
			Token:   os.Getenv("WEWORK_CORP_SECRET"),
			PlatformConfig: map[string]interface{}{
				"corp_id":          corpID,
				"agent_id":         os.Getenv("WEWORK_AGENT_ID"),
				"encoding_aes_key": os.Getenv("WEWORK_ENCODING_AES_KEY"),
				"token":            os.Getenv("WEWORK_TOKEN"),
			},
		}
	}

	// Load DingTalk config if credentials are provided
	if appKey := os.Getenv("DINGTALK_APP_KEY"); appKey != "" {
		config.Channels["dingtalk"] = ChannelConfig{
			Type:    ChannelTypeDingTalk,
			Enabled: true,
			Token:   os.Getenv("DINGTALK_APP_SECRET"),
			PlatformConfig: map[string]interface{}{
				"app_key":          appKey,
				"agent_id":         os.Getenv("DINGTALK_AGENT_ID"),
				"agent_secret":     os.Getenv("DINGTALK_AGENT_SECRET"),
				"encoding_aes_key": os.Getenv("DINGTALK_ENCODING_AES_KEY"),
			},
		}
	}

	// Load QQ config (works with default settings)
	// QQ bot can be configured via environment variables or use defaults (localhost:3000)
	config.Channels["qq"] = ChannelConfig{
		Type:    ChannelTypeQQ,
		Enabled: os.Getenv("QQ_BOT_ENABLED") == "true",
		PlatformConfig: map[string]interface{}{
			"api_base": os.Getenv("QQ_BOT_API_BASE"),
			"self_id":  os.Getenv("QQ_BOT_SELF_ID"),
		},
	}

	return config, nil
}
