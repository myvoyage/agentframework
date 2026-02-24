// AgentFramework - TUI (Terminal User Interface)
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"AgentFramework/core"
	"AgentFramework/agent"
)

// Run starts the TUI application
func Run() error {
	ctx := context.Background()

	// 初始化核心应用
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
	}

	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	coreApp, err := core.NewApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		return fmt.Errorf("failed to create core application: %w", err)
	}

	if err := coreApp.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize core application: %w", err)
	}

	// 初始化工作流管理器和文件浏览器
	coreApp.GetWorkflowManager().Init(ctx)
	coreApp.GetFileExplorer().Init(ctx)

	// 创建 TUI 模型
	model := NewTUIModel(ctx, coreApp)

	// 启动 TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
