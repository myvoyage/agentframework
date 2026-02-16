// Agent Framework - Notification Module Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"os/exec"
	"runtime"
	"testing"
)

// TestNotificationModuleCreation 测试通知模块创建
func TestNotificationModuleCreation(t *testing.T) {
	config := NotificationConfig{
		MaxQueueSize:      10,
		EnableSound:       false,
		DefaultIcon:       "",
		EnablePersistence: false,
		PersistencePath:    "",
		AllowedCategories: []string{"info", "warning", "error"},
		EnableGrouping:    false,
		EnableActions:     false,
	}

	module, err := NewNotificationModule(config)
	if err != nil {
		t.Fatalf("Failed to create notification module: %v", err)
	}

	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.config.MaxQueueSize != 10 {
		t.Errorf("Expected MaxQueueSize to be 10, got %d", module.config.MaxQueueSize)
	}
}

// TestNotificationConfigDefaults 测试配置默认值
func TestNotificationConfigDefaults(t *testing.T) {
	config := NotificationConfig{}

	module, err := NewNotificationModule(config)
	if err != nil {
		t.Fatalf("Failed to create notification module: %v", err)
	}

	if module.config.MaxQueueSize != 100 {
		t.Errorf("Expected default MaxQueueSize to be 100, got %d", module.config.MaxQueueSize)
	}

	if len(module.config.AllowedCategories) == 0 {
		t.Error("Expected default AllowedCategories to be non-empty")
	}
}

// isPlatformSupported 检查平台是否支持通知
func isPlatformSupported() bool {
	if runtime.GOOS == "windows" {
		return true
	}

	if runtime.GOOS == "darwin" {
		return true
	}

	if runtime.GOOS == "linux" {
		// 检查notify-send是否可用
		_, err := exec.LookPath("notify-send")
		return err == nil
	}

	return false
}
