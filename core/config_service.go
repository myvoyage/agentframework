// Agent Framework - Config Service
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// ConfigService handles configuration operations
type ConfigService struct {
	app *Application
}

// NewConfigService creates a new config service
func NewConfigService(app *Application) *ConfigService {
	return &ConfigService{app: app}
}

// GetConfig returns the current configuration
func (s *ConfigService) GetConfig(ctx context.Context) (*agent.HostConfig, error) {
	return s.app.host.Config(), nil
}

// GetConfigValue returns a specific configuration value by key path
func (s *ConfigService) GetConfigValue(ctx context.Context, key string) (interface{}, error) {
	config := s.app.host.Config()
	if config == nil {
		return nil, fmt.Errorf("configuration not available")
	}

	// Support nested key paths like "models.gpt4.model"
	// For now, return basic config info
	switch key {
	case "defaultModel":
		return config.DefaultModel, nil
	case "skillSystemDir":
		return config.SkillSystemDir, nil
	default:
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
}

// SetConfigValue updates a configuration value (requires restart)
func (s *ConfigService) SetConfigValue(ctx context.Context, key string, value interface{}) error {
	// This would require persistent storage and restart
	return fmt.Errorf("config update requires restarting the application")
}

// PrintConfig prints the current configuration
func (s *ConfigService) PrintConfig(ctx context.Context, outputFormat string) error {
	config, err := s.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", config)
	case "yaml", "":
		// YAML-like output
		fmt.Println("Configuration:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("Default Model: %s\n", config.DefaultModel)
		fmt.Printf("Skill System Dir: %s\n", config.SkillSystemDir)
		fmt.Printf("Models: %d configured\n", len(config.Models))
		fmt.Println("────────────────────────────────────────────────────────────")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
