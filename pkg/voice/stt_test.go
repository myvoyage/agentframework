// Agent Framework - STT Module Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"testing"
)

// TestSTTModuleCreation 测试STT模块创建
func TestSTTModuleCreation(t *testing.T) {
	config := STTConfig{
		Engine:           "local",
		Language:         "en-US",
		SampleRate:       16000,
		Channels:         1,
		EnableAutoDetect: false,
		EnablePunctuation: false,
		EnableTimestamp:  false,
		MaxDuration:      60,
		TempDir:          "",
	}

	module, err := NewSTTModule(config)
	if err != nil {
		t.Fatalf("Failed to create STT module: %v", err)
	}

	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.config.Engine != "local" {
		t.Errorf("Expected Engine to be 'local', got '%s'", module.config.Engine)
	}
}

// TestSTTConfigDefaults 测试配置默认值
func TestSTTConfigDefaults(t *testing.T) {
	config := STTConfig{}

	module, err := NewSTTModule(config)
	if err != nil {
		t.Fatalf("Failed to create STT module: %v", err)
	}

	if module.config.Engine != "local" {
		t.Errorf("Expected default Engine to be 'local', got '%s'", module.config.Engine)
	}

	if module.config.SampleRate != 16000 {
		t.Errorf("Expected default SampleRate to be 16000, got %d", module.config.SampleRate)
	}

	if module.config.Channels != 1 {
		t.Errorf("Expected default Channels to be 1, got %d", module.config.Channels)
	}

	module.Close()
}
