// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package shell

import (
	"context"
	"testing"
	"time"
)

func TestShellModule_BasicExecution(t *testing.T) {
	config := ShellConfig{
		Timeout:          5000,
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{},
		EnableBlacklist:  true,
		CommandBlacklist: []string{"rm", "rmdir"},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	// Test basic command execution
	result, err := module.execCommand("echo hello", 0, "")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Command execution failed: %v", result["error"])
	}

	stdout := result["stdout"].(string)
	if stdout != "hello\n" && stdout != "hello\r\n" {
		t.Errorf("Unexpected output: %s", stdout)
	}
}

func TestShellModule_CommandBlacklist(t *testing.T) {
	config := ShellConfig{
		Timeout:          5000,
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{},
		EnableBlacklist:  true,
		CommandBlacklist: []string{"rm", "rmdir"},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	// Test blocked command
	result, err := module.execCommand("rm test.txt", 0, "")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Blocked command should not succeed")
	}

	if result["error"] != "Command not allowed" {
		t.Errorf("Unexpected error: %v", result["error"])
	}
}

func TestShellModule_CommandWhitelist(t *testing.T) {
	config := ShellConfig{
		Timeout:          5000,
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{"echo", "ls"},
		EnableBlacklist:  false,
		CommandBlacklist: []string{},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	// Test allowed command
	result, err := module.execCommand("echo test", 0, "")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Allowed command should succeed: %v", result["error"])
	}

	// Test disallowed command
	result, err = module.execCommand("pwd", 0, "")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Disallowed command should not succeed")
	}
}

func TestShellModule_Timeout(t *testing.T) {
	config := ShellConfig{
		Timeout:          1000, // 1 second
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{},
		EnableBlacklist:  false,
		CommandBlacklist: []string{},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	// Test timeout (sleep for 3 seconds with 1 second timeout)
	start := time.Now()
	result, err := module.execCommand("sleep 3", 0, "")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Command should timeout")
	}

	// Should timeout around 1 second, not 3 seconds
	if duration > 2*time.Second {
		t.Errorf("Timeout took too long: %v", duration)
	}
}

func TestShellModule_GetTools(t *testing.T) {
	config := ShellConfig{
		Timeout:          5000,
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{},
		EnableBlacklist:  true,
		CommandBlacklist: []string{},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestShellModule_Stats(t *testing.T) {
	config := ShellConfig{
		Timeout:          5000,
		MemoryLimit:      512,
		CPULimit:         1,
		CommandWhitelist: []string{},
		EnableBlacklist:  true,
		CommandBlacklist: []string{"rm"},
	}

	module, err := NewShellModule(config)
	if err != nil {
		t.Fatalf("Failed to create shell module: %v", err)
	}
	defer module.Close()

	// Execute some commands
	module.execCommand("echo test1", 0, "")
	module.execCommand("echo test2", 0, "")
	module.execCommand("rm test", 0, "")             // blocked
	module.execCommand("invalid_command_xyz", 0, "") // failed

	stats := module.GetStats()

	if stats["total_executions"] != 4 {
		t.Errorf("Expected 4 total executions, got %d", stats["total_executions"])
	}

	if stats["success_count"] != 2 {
		t.Errorf("Expected 2 successful executions, got %d", stats["success_count"])
	}

	if stats["blocked_count"] != 1 {
		t.Errorf("Expected 1 blocked execution, got %d", stats["blocked_count"])
	}

	if stats["failure_count"] != 1 {
		t.Errorf("Expected 1 failed execution, got %d", stats["failure_count"])
	}
}
