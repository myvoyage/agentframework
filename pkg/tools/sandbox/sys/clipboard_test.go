// Agent Framework - Clipboard Module Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"os/exec"
	"testing"
)

// TestClipboardModuleCreation 测试剪贴板模块创建
func TestClipboardModuleCreation(t *testing.T) {
	config := ClipboardConfig{
		MaxHistorySize:  10,
		EnableHistory:   false,
		EnableMonitor:   false,
		AllowedFormats:  []string{"text"},
		MaxTextSize:     1024,
		EnableEncryption: false,
	}

	module, err := NewClipboardModule(config)
	if err != nil {
		t.Fatalf("Failed to create clipboard module: %v", err)
	}

	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.config.MaxHistorySize != 10 {
		t.Errorf("Expected MaxHistorySize to be 10, got %d", module.config.MaxHistorySize)
	}
}

// TestClipboardConfigDefaults 测试配置默认值
func TestClipboardConfigDefaults(t *testing.T) {
	config := ClipboardConfig{}

	module, err := NewClipboardModule(config)
	if err != nil {
		t.Fatalf("Failed to create clipboard module: %v", err)
	}

	if module.config.MaxHistorySize != 100 {
		t.Errorf("Expected default MaxHistorySize to be 100, got %d", module.config.MaxHistorySize)
	}

	if module.config.MaxTextSize != 1024*1024 {
		t.Errorf("Expected default MaxTextSize to be 1MB, got %d", module.config.MaxTextSize)
	}

	if len(module.config.AllowedFormats) == 0 || module.config.AllowedFormats[0] != "text" {
		t.Errorf("Expected default AllowedFormats to contain 'text'")
	}
}

// commandExists 检查命令是否存在
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
