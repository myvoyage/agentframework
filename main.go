// Agent Framework - Main Entry Point
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
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"AgentFramework/cmd/cli"
	cmdtui "AgentFramework/cmd/tui"
	"AgentFramework/agent"
)

//go:embed all:frontend/dist
var assets embed.FS

// RunMode determines the application run mode
type RunMode int

const (
	// ModeDesktop runs the desktop GUI application
	ModeDesktop RunMode = iota
	// ModeCLI runs the command-line interface
	ModeCLI
	// ModeTUI runs the terminal user interface
	ModeTUI
)

func main() {
	// Determine run mode based on command-line arguments
	mode := detectRunMode()

	// Print startup banner
	printBanner(mode)

	switch mode {
	case ModeCLI:
		// Run CLI mode
		if err := runCLIMode(); err != nil {
			fmt.Fprintf(os.Stderr, "CLI error: %v\n", err)
			os.Exit(1)
		}
	case ModeTUI:
		// Run TUI mode
		if err := runTUIMode(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
	case ModeDesktop:
		// Run desktop mode (Wails UI)
		if err := runDesktopMode(); err != nil {
			fmt.Fprintf(os.Stderr, "Desktop error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown run mode")
		os.Exit(1)
	}
}

// printBanner prints the startup banner with mode information
func printBanner(mode RunMode) {
	// Only print banner for CLI and TUI modes
	// Desktop mode has its own UI
	if mode == ModeDesktop {
		return
	}

	modeName := ""
	switch mode {
	case ModeCLI:
		modeName = "CLI"
	case ModeTUI:
		modeName = "TUI (Terminal UI)"
	default:
		modeName = "Unknown"
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        AgentFramework - Enterprise AI Agent Framework     ║")
	fmt.Println("║                      Version 2.0.0                        ║")
	fmt.Printf("║           Running Mode: %-30s        ║\n", modeName)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// detectRunMode determines whether to run in CLI, TUI, or desktop mode
// 支持的启动方式:
//   无参数             -> 默认 Wails UI (桌面GUI)
//   -tui, --tui       -> TUI (终端用户界面)
//   -cli, --cli       -> CLI (命令行模式)
//   任何其他参数      -> 自动检测为 CLI 模式
func detectRunMode() RunMode {
	// Check for CLI-specific flags or subcommands
	args := os.Args[1:]

	// No arguments - default to desktop mode (Wails UI)
	if len(args) == 0 {
		return ModeDesktop
	}

	// Check for explicit mode flags
	for _, arg := range args {
		switch arg {
		case "-tui", "--tui":
			return ModeTUI
		case "-cli", "--cli":
			return ModeCLI
		}
	}

	// Check for CLI-specific flags or subcommands
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "help", "-v", "--version", "version",
			"completion", "init",
			// Core commands
			"workflow", "skill", "config", "file", "agent",
			// Advanced commands
			"task", "schedule", "scheduler", "sched",
			"channel", "plugin", "monitor",
			"token", "host",
			// Legacy
			"enhanced-skill",
			// Global flags
			"-c", "--config", "-m", "--model", "-o", "--output":
			return ModeCLI
		}
	}

	// Check if first argument is a flag (starts with -)
	if len(args) > 0 && args[0][0] == '-' {
		return ModeCLI
	}

	// Default to desktop mode
	return ModeDesktop
}

// runCLIMode runs the application in CLI mode
func runCLIMode() error {
	// Run CLI mode
	return cli.Execute()
}

// runTUIMode runs the application in TUI (Terminal UI) mode
func runTUIMode() error {
	// Create default config
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

	// Run TUI using the tui package
	return cmdtui.Run(defaultHostConfig, modelFactory)
}

// runDesktopMode runs the application in desktop mode
func runDesktopMode() error {
	// Create application instance
	app := NewAppWithConfig(true, 8080)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "agentframework-desktop",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to run desktop app: %w", err)
	}

	return nil
}
