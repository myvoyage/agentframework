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

// SPDX-License-Identifier: AGPL-3.0-or-later

package collaboration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HostIntegration provides integration between AgentTeam and Host system
type HostIntegration struct {
	team   *AgentTeam
	host   HostInterface
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// HostInterface defines the interface for interacting with the Host system
type HostInterface interface {
	GetAgent(name string) (interface{}, error)
	ListAgents() []string
	GetModel(name string) (interface{}, error)
}

// NewHostIntegration creates a new host integration
func NewHostIntegration(team *AgentTeam, host HostInterface) *HostIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	return &HostIntegration{
		team:   team,
		host:   host,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AutoRegisterAgents automatically registers agents from the host to the team
func (hi *HostIntegration) AutoRegisterAgents(agentConfigs []AgentConfig) error {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	for _, config := range agentConfigs {
		// Get agent from host
		agent, err := hi.host.GetAgent(config.Name)
		if err != nil {
			return fmt.Errorf("failed to get agent %s: %w", config.Name, err)
		}

		// Create wrapper
		wrapper, err := hi.createWrapper(agent, config)
		if err != nil {
			return fmt.Errorf("failed to create wrapper for %s: %w", config.Name, err)
		}

		// Add to team
		hi.team.AddMember(wrapper, config.Role, config.Capabilities, config.Priority)
	}

	return nil
}

// createWrapper creates an agent wrapper from an agent and config
func (hi *HostIntegration) createWrapper(rawAgent interface{}, config AgentConfig) (*DefaultAgentWrapper, error) {
	// Determine model name
	modelName := config.Model
	if modelName == "" {
		modelName = "default"
	}

	// Convert interface{} to Agent type
	// The host returns agent.Agent interface, we need to adapt it
	agent, ok := rawAgent.(Agent)
	if !ok {
		return nil, fmt.Errorf("agent %s does not implement Agent interface", config.Name)
	}

	// Create wrapper
	wrapper := NewDefaultAgentWrapper(
		agent,
		config.Capabilities,
		modelName,
	)

	// Set metadata
	for k, v := range config.Metadata {
		wrapper.SetMetadata(k, v)
	}

	return wrapper, nil
}

// AgentConfig defines the configuration for an agent in the team
type AgentConfig struct {
	Name         string
	Role         string
	Model        string
	Capabilities []string
	Priority     int
	Metadata     map[string]interface{}
}

// Shutdown shuts down the host integration
func (hi *HostIntegration) Shutdown() {
	hi.cancel()
}

// TeamConfigBuilder helps build team configurations
type TeamConfigBuilder struct {
	config TeamConfig
}

// NewTeamConfigBuilder creates a new team config builder
func NewTeamConfigBuilder(name string) *TeamConfigBuilder {
	return &TeamConfigBuilder{
		config: TeamConfig{
			Name:          name,
			MaxConcurrent: 10,
			RouterConfig: RouterConfig{
				DefaultStrategy: StrategyIntelligent,
				EnableCaching:   true,
				CacheTTL:        5 * time.Minute,
				ScoringWeights:  DefaultScoringWeights(),
			},
		},
	}
}

// WithDescription sets the team description
func (b *TeamConfigBuilder) WithDescription(desc string) *TeamConfigBuilder {
	b.config.Description = desc
	return b
}

// WithMaxConcurrent sets the max concurrent tasks
func (b *TeamConfigBuilder) WithMaxConcurrent(max int) *TeamConfigBuilder {
	b.config.MaxConcurrent = max
	return b
}

// WithRouterStrategy sets the default routing strategy
func (b *TeamConfigBuilder) WithRouterStrategy(strategy RoutingStrategy) *TeamConfigBuilder {
	b.config.RouterConfig.DefaultStrategy = strategy
	return b
}

// WithCache enables or disables caching
func (b *TeamConfigBuilder) WithCache(enabled bool, ttl time.Duration) *TeamConfigBuilder {
	b.config.RouterConfig.EnableCaching = enabled
	b.config.RouterConfig.CacheTTL = ttl
	return b
}

// WithScoringWeights sets the scoring weights
func (b *TeamConfigBuilder) WithScoringWeights(weights ScoringWeights) *TeamConfigBuilder {
	b.config.RouterConfig.ScoringWeights = weights
	return b
}

// Build builds the team configuration
func (b *TeamConfigBuilder) Build() TeamConfig {
	return b.config
}

// QuickStart provides a quick way to set up a team with the host
func QuickStart(ctx context.Context, host HostInterface, teamName string, agentConfigs []AgentConfig) (*AgentTeam, error) {
	// Build team config
	builder := NewTeamConfigBuilder(teamName)
	config := builder.Build()

	// Create team
	team := NewAgentTeam(config)

	// Create integration
	integration := NewHostIntegration(team, host)

	// Auto-register agents
	if err := integration.AutoRegisterAgents(agentConfigs); err != nil {
		return nil, fmt.Errorf("failed to register agents: %w", err)
	}

	// Start team
	if err := team.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start team: %v", err)
	}

	return team, nil
}

// TeamManager manages multiple teams
type TeamManager struct {
	teams map[string]*AgentTeam
	mu    sync.RWMutex
}

// NewTeamManager creates a new team manager
func NewTeamManager() *TeamManager {
	return &TeamManager{
		teams: make(map[string]*AgentTeam),
	}
}

// RegisterTeam registers a team
func (tm *TeamManager) RegisterTeam(team *AgentTeam) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	name := team.Name()
	if _, exists := tm.teams[name]; exists {
		return fmt.Errorf("team %s already registered", name)
	}

	tm.teams[name] = team

	return nil
}

// UnregisterTeam unregisters a team
func (tm *TeamManager) UnregisterTeam(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.teams[name]; !exists {
		return fmt.Errorf("team %s not found", name)
	}

	delete(tm.teams, name)

	return nil
}

// GetTeam gets a team by name
func (tm *TeamManager) GetTeam(name string) (*AgentTeam, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	team, exists := tm.teams[name]
	if !exists {
		return nil, fmt.Errorf("team %s not found", name)
	}

	return team, nil
}

// ListTeams lists all team names
func (tm *TeamManager) ListTeams() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	names := make([]string, 0, len(tm.teams))
	for name := range tm.teams {
		names = append(names, name)
	}

	return names
}

// ShutdownAll shuts down all teams
func (tm *TeamManager) ShutdownAll(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var lastErr error
	for _, team := range tm.teams {
		if err := team.Stop(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
