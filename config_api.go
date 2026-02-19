// Agent Framework - Configuration Management API
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"AgentFramework/agent"
)

// GetConfig returns the current configuration
func (a *App) GetConfig() (*agent.HostConfig, error) {
	return a.core.GetHost().Config(), nil
}

// UpdateConfig updates the configuration
func (a *App) UpdateConfig(config *agent.HostConfig) error {
	// For now, we'll just return nil since Host doesn't have an UpdateConfig method yet
	// This will be implemented in a future update
	return nil
}

// ReloadConfig reloads the configuration
func (a *App) ReloadConfig() error {
	// For now, we'll just return nil since Host doesn't have a ReloadConfig method yet
	// This will be implemented in a future update
	return nil
}
