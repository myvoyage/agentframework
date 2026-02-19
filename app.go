// Agent Framework - Main Application Entry Point
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"AgentFramework/agent"
	"AgentFramework/core"
)

// App struct wraps the core application for desktop usage
type App struct {
	core *core.Application
	ctx  context.Context
}

// NewApp creates a new App application struct
// Refactored to use core.Application (DRY principle)
func NewApp() *App {
	ctx := context.Background()

	// Initialize OpenTelemetry
	_, err := InitOpenTelemetry(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize OpenTelemetry: %v\n", err)
	}

	// 创建默认的HostConfig
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills", // 启用技能系统
	}

	// 创建模型工厂
	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	// Create core application (shared between CLI and desktop)
	coreApp, err := core.NewApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create core application: %w", err))
	}

	// Initialize core application
	if err := coreApp.Initialize(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize core application: %w", err))
	}

	return &App{
		core: coreApp,
		ctx:  ctx,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化工作流管理器
	a.core.GetWorkflowManager().Init(ctx)
	// 初始化文件浏览器
	a.core.GetFileExplorer().Init(ctx)
}

// getSkillSystemBaseDir returns the base directory for the skill system
func (a *App) getSkillSystemBaseDir() string {
	if a.core.GetHost() != nil {
		cfg := a.core.GetHost().Config()
		if cfg != nil && cfg.SkillSystemDir != "" {
			return cfg.SkillSystemDir
		}
	}
	return ".skills" // 默认值
}
