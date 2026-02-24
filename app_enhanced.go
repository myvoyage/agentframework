// Agent Framework - Enhanced Main Application Entry Point with Security and Performance
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/agent"
	"AgentFramework/core"
)

// EnhancedApp represents the enhanced desktop application with security and performance
type EnhancedApp struct {
	core *core.EnhancedApplication
	ctx  context.Context
}

// NewEnhancedApp creates a new enhanced desktop application
func NewEnhancedApp() *EnhancedApp {
	ctx := context.Background()

	// 创建增强的默认配置
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills",
		// Note: JWTSecret, JWTAlgorithm, Audience, AdminUserID, RedisEnabled
		// should be configured through the security layer, not HostConfig
	}

	// 创建模型工厂
	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	// 创建增强的核心应用
	enhancedCore, err := core.NewEnhancedApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create enhanced core application: %w", err))
	}

	// 初始化应用
	if err := enhancedCore.Initialize(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize enhanced application: %w", err))
	}

	return &EnhancedApp{
		core: enhancedCore,
		ctx:  ctx,
	}
}

// startup is called when the app starts
func (a *EnhancedApp) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize workflow manager
	a.core.GetWorkflowManager().Init(ctx)

	// Initialize file explorer
	a.core.GetFileExplorer().Init(ctx)

	log.Println("✅ Enhanced application started successfully")
	log.Println("📊 Security features: JWT validation, RBAC, Input validation")
	log.Println("⚡ Performance features: Object pools, Lock-free structures, Multi-level cache")
}

// shutdown handles graceful shutdown
func (a *EnhancedApp) shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down enhanced application...")

	// Cleanup resources
	if err := a.core.Cleanup(ctx); err != nil {
		log.Printf("Warning: Cleanup encountered errors: %v", err)
	}

	log.Println("✅ Shutdown complete")
	return nil
}

// GetCore returns the enhanced core application
func (a *EnhancedApp) GetCore() *core.EnhancedApplication {
	return a.core
}

// GetContext returns the application context
func (a *EnhancedApp) GetContext() context.Context {
	return a.ctx
}

// ===== Backward Compatibility =====

// GetHost returns the agent host (for backward compatibility)
func (a *EnhancedApp) GetHost() *agent.Host {
	return a.core.GetHost()
}

// GetSkillLibrary returns the skill library
func (a *EnhancedApp) GetSkillLibrary() agent.SkillLibrary {
	return a.core.GetSkillLibrary()
}

// GetSkillSystem returns the skill system
func (a *EnhancedApp) GetSkillSystem() *agent.SkillSystem {
	return a.core.GetSkillSystem()
}

// GetFileExplorer returns the file explorer
func (a *EnhancedApp) GetFileExplorer() *agent.FileExplorer {
	return a.core.GetFileExplorer()
}

// GetWorkflowManager returns the workflow manager
func (a *EnhancedApp) GetWorkflowManager() *agent.WorkflowManager {
	return a.core.GetWorkflowManager()
}

// getConfig returns the host configuration
func (a *EnhancedApp) getConfig() *agent.HostConfig {
	return a.core.GetConfig()
}

// getSkillSystemBaseDir returns the base directory for the skill system
func (a *EnhancedApp) getSkillSystemBaseDir() string {
	cfg := a.getConfig()
	if cfg != nil && cfg.SkillSystemDir != "" {
		return cfg.SkillSystemDir
	}
	return ".skills"
}

// ===== Enhanced Features =====

// ValidateInput validates and sanitizes user input
func (a *EnhancedApp) ValidateInput(input string) (string, error) {
	return a.core.ValidateAndSanitizeInput(input)
}

// ValidateJWT validates a JWT token
func (a *EnhancedApp) ValidateJWT(token string) (string, error) {
	return a.core.ValidateJWT(token)
}

// CheckPermission checks if a user has permission
func (a *EnhancedApp) CheckPermission(userID, resource, action string) bool {
	return a.core.CheckPermission(userID, resource, action)
}

// RequirePermission checks permission and returns error if denied
func (a *EnhancedApp) RequirePermission(userID, resource, action string) error {
	return a.core.RequirePermission(userID, resource, action)
}

// GetFromCache retrieves from cache
func (a *EnhancedApp) GetFromCache(key string) (interface{}, error) {
	return a.core.GetFromCache(key)
}

// SetInCache stores in cache
func (a *EnhancedApp) SetInCache(key string, value interface{}, ttl int) error {
	return a.core.SetInCache(key, value, time.Duration(ttl)*time.Second)
}

// GetMetrics returns performance metrics
func (a *EnhancedApp) GetMetrics() map[string]interface{} {
	metrics := a.core.GetMetrics()

	return map[string]interface{}{
		"requestCount":      metrics.GetRequestCount(),
		"errorCount":        metrics.GetErrorCount(),
		"totalLatency":      metrics.GetTotalLatency(),
		"averageLatency":   metrics.GetAverageLatency(),
		"minLatency":       metrics.GetMinLatency(),
		"maxLatency":       metrics.GetMaxLatency(),
	}
}
