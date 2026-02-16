// Agent Framework - TTS Module Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"testing"
)

// TestTTSModuleCreation 测试TTS模块创建
func TestTTSModuleCreation(t *testing.T) {
	config := TTSConfig{
		Engine:        "local",
		Voice:         "default",
		Rate:          0,
		Pitch:         0,
		Volume:       1.0,
		OutputFormat: "wav",
		OutputQuality: "medium",
		EnableSSML:   false,
		TempDir:      "",
	}

	module, err := NewTTSModule(config)
	if err != nil {
		t.Fatalf("Failed to create TTS module: %v", err)
	}

	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.config.Engine != "local" {
		t.Errorf("Expected Engine to be 'local', got '%s'", module.config.Engine)
	}
}

// TestTTSConfigDefaults 测试配置默认值
func TestTTSConfigDefaults(t *testing.T) {
	config := TTSConfig{}

	module, err := NewTTSModule(config)
	if err != nil {
		t.Fatalf("Failed to create TTS module: %v", err)
	}

	if module.config.Engine != "local" {
		t.Errorf("Expected default Engine to be 'local', got '%s'", module.config.Engine)
	}

	if module.config.Rate != 0 {
		t.Errorf("Expected default Rate to be 0, got %d", module.config.Rate)
	}

	if module.config.Volume != 1.0 {
		t.Errorf("Expected default Volume to be 1.0, got %.1f", module.config.Volume)
	}

	module.Close()
}
