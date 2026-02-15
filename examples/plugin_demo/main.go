// Agent Framework - Complete Plugin Demo
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"AgentFramework/agent"
	"AgentFramework/agent/config"
	"AgentFramework/agent/errors"
	"AgentFramework/agent/pool"
)

func main() {
	fmt.Println("=== Agent Framework Plugin System Demo ===\n")

	// 初始化插件仓库管理器
	repoPath := filepath.Join("..", "repository", "plugins")
	manager, err := agent.NewPluginRepositoryManager([]agent.PluginRepository{
		{Name: "local", URL: "file://" + repoPath, Type: "local", Enabled: true},
	}, repoPath)

	if err != nil {
		log.Fatalf("Failed to initialize plugin repository: %v\n", err)
	}
	defer manager.Stop()

	// 初始化插件生命周期管理器
	lifecycleCfg := agent.DefaultPluginHealthConfig()
	lifecycleMgr := agent.NewPluginLifecycleManager(lifecycleCfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册示例插件
	if err := registerExamplePlugins(manager, lifecycleMgr); err != nil {
		log.Fatalf("Failed to register plugins: %v\n", err)
	}

	// 演示插件功能
	demoPluginUsage(manager, lifecycleMgr)

	// 等待用户输入
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()

	log.Println("Plugin System Demo completed!")
}

// registerExamplePlugins registers example plugins for demonstration
func registerExamplePlugins(mgr *agent.PluginRepositoryManager, lifecycleMgr *agent.PluginLifecycleManager) error {
	// 示例1：健康检查插件
	healthPlugin := &ExamplePlugin{
		name:         "health_check",
		version:      "1.0.0",
		description:  "Performs periodic health checks on plugins",
		author:       "Agent Framework",
		license:      "AGPL-3.0-or-later",
		capabilities: []string{"health_check", "metrics"},
		maxMemoryMB:  128,
		maxCPUPercent: 10,
		requiredNetwork: false,
		requiredFS:       false,
		sandboxEnabled:   true,
	}

	if err := mgr.AddRepository(agent.PluginRepository{
		Name:    "health_check",
		URL:     "file://local",
		Type:    "local",
		Enabled: true,
	}); err != nil {
		return err
	}

	// 注册到生命周期管理器
	if err := lifecycleMgr.Register(healthPlugin); err != nil {
		return err
	}

	// 示例2：配置管理器
	configPlugin := &ExamplePlugin{
		name:         "config_manager",
		version:      "1.0.0",
		description:  "Provides unified configuration management",
		author:       "Agent Framework",
		license:      "AGPL-3.0-or-later",
		capabilities: []string{"config", "get", "set", "validate"},
		maxMemoryMB:   64,
		maxCPUPercent: 5,
		requiredNetwork: false,
		requiredFS:       true,
		sandboxEnabled:   false,
	}

	if err := mgr.AddRepository(agent.PluginRepository{
		Name:    "config_manager",
		URL:     "file://local",
		Type:    "local",
		Enabled: true,
	}); err != nil {
		return err
	}

	if err := lifecycleMgr.Register(configPlugin); err != nil {
		return err
	}

	return nil
}

// demoPluginUsage demonstrates plugin system features
func demoPluginUsage(mgr *agent.PluginRepositoryManager, lifecycleMgr *agent.PluginLifecycleManager) {
	fmt.Println("\n--- Plugin Repository Demo ---")

	// 列出所有插件
	repos, err := mgr.ListRepositories()
	if err != nil {
		log.Printf("Failed to list repositories: %v\n", err)
		return
	}

	for _, repo := range repos {
		fmt.Printf("Repository: %s\n", repo.Name)
		fmt.Printf("  URL: %s\n", repo.URL)
		fmt.Printf("  Type: %s\n", repo.Type)
		fmt.Printf("  Enabled: %v\n", repo.Enabled)
	}

	fmt.Println("\n--- Plugin Lifecycle Demo ---")

	// 测试生命周期管理
	stats := lifecycleMgr.GetStats()
	for name, stat := range stats {
		fmt.Printf("Plugin: %s\n", name)
		fmt.Printf("  State: %v\n", stat["state"])
		fmt.Printf("  Health: %v\n", stat["health"])
	}

	// 测试配置系统
	fmt.Println("\n--- Configuration System Demo ---")

	cfg := config.NewConfigManager("config.json")

	// 设置配置值
	if err := cfg.Set("app.name", "Agent Framework Plugin Demo"); err != nil {
		log.Printf("Failed to set config: %v\n", err)
	}

	// 获取配置值
	if val := cfg.Get("app.name", ""); val != "" {
		fmt.Printf("app.name = %v\n", val)
	}

	// 验证配置
	if err := cfg.Validate("app.name"); err != nil {
		log.Printf("Validation failed: %v\n", err)
	}

	// 测试对象池
	fmt.Println("\n--- Object Pool Demo ---")
	p := pool.GetGlobalPool()

	// 复用对象
	testObj := map[string]interface{}{"key": "value"}
	if err := p.Put("test", testObj); err != nil {
		log.Printf("Failed to put object: %v\n", err)
	}

	// 获取对象
	if retrieved, ok := p.Get("test"); ok && retrieved != nil {
		fmt.Printf("Retrieved: %v\n", retrieved)
	}

	// 清理对象池
	p.Cleanup("test")

	fmt.Println("Plugin System Demo completed successfully!")
}

// ExamplePlugin 示例插件实现
type ExamplePlugin struct {
	name            string
	version         string
	description     string
	author          string
	license         string
	capabilities    []string
	maxMemoryMB     int
	maxCPUPercent   int
	requiredNetwork bool
	requiredFS      bool
	sandboxEnabled   bool
	enabled         bool
}

// Plugin interface implementation
func (p *ExamplePlugin) Name() string {
	return p.name
}

func (p *ExamplePlugin) Version() string {
	return p.version
}

func (p *ExamplePlugin) Description() string {
	return p.description
}

func (p *ExamplePlugin) Initialize(ctx context.Context, host interface{}) error {
	fmt.Printf("[Plugin] %s v%s initialized\n", p.name, p.version)
	return nil
}

func (p *ExamplePlugin) Shutdown(ctx context.Context) error {
	fmt.Printf("[Plugin] %s v%s shutdown\n", p.name, p.version)
	return nil
}

func (p *ExamplePlugin) IsEnabled() bool {
	return p.enabled
}

func (p *ExamplePlugin) Enable() error {
	p.enabled = true
	fmt.Printf("[Plugin] %s enabled\n", p.name)
	return nil
}

func (p *ExamplePlugin) Disable() error {
	p.enabled = false
	fmt.Printf("[Plugin] %s disabled\n", p.name)
	return nil
}
