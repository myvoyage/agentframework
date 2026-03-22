// Agent Framework - TUI (Terminal User Interface)
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"AgentFramework/core"
	"AgentFramework/agent"
)

// Run starts the TUI application
// This is the main entry point for TUI mode
func Run(hostCfg *agent.HostConfig, modelFactory agent.ModelFactory) error {
	ctx := context.Background()

	// Use provided config or create default
	if hostCfg == nil {
		hostCfg = &agent.HostConfig{
			Models: map[string]agent.ModelConfig{
				"default": {
					Type:  "ollama",
					Model: "llama3",
				},
			},
			SkillSystemDir: ".skills",
		}
	}

	// Use provided model factory or create default
	if modelFactory == nil {
		modelFactory = agent.NewModelFactoryWithConfig(agent.ModelConfig{
			Type:  "ollama",
			Model: "llama3",
		})
	}

	// Create core application
	coreApp, err := core.NewApplication(ctx, hostCfg, modelFactory, nil)
	if err != nil {
		return fmt.Errorf("failed to create core application: %w", err)
	}

	if err := coreApp.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize core application: %w", err)
	}

	// Initialize managers
	coreApp.GetWorkflowManager().Init(ctx)
	coreApp.GetFileExplorer().Init(ctx)

	// Create TUI model
	model := NewTUIModel(ctx, coreApp)

	// Start TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// RunWithApp starts TUI with an existing application instance
func RunWithApp(ctx context.Context, coreApp *core.Application) error {
	// Initialize managers if not already done
	coreApp.GetWorkflowManager().Init(ctx)
	coreApp.GetFileExplorer().Init(ctx)

	// Create TUI model
	model := NewTUIModel(ctx, coreApp)

	// Start TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
