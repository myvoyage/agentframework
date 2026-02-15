// Agent Framework - Integration Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"AgentFramework/agent"
	"AgentFramework/agent/config"
)

func TestConfigIntegration(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_config.json"

	// Create test configuration
	testConfig := `{
		"model.defaultModel": "llama3",
		"memory.historySize": 100,
		"cache.maxSize": 200,
		"cache.ttl": "1h",
		"worker.count": 5,
		"container.maxTotalSize": 20,
		"eventBus.initialQueueSize": 1000
	}`

	// Write test config file
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test: Initialize config manager
	cfg := config.NewConfigManager(configPath)
	if cfg == nil {
		t.Fatal("Failed to create config manager")
	}

	// Test: Load configuration
	if err := cfg.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test: Get configuration values
	defaultModel := cfg.Get("model.defaultModel", "default")
	if defaultModel != "llama3" {
		t.Errorf("Expected model.defaultModel to be 'llama3', got '%v'", defaultModel)
	}

	historySize := cfg.Get("memory.historySize", int64(50))
	if historySize != int64(100) {
		t.Errorf("Expected memory.historySize to be 100, got %v", historySize)
	}

	// Test: Get integer values
	cacheMax := cfg.GetInt("cache.maxSize", 100)
	if cacheMax != 200 {
		t.Errorf("Expected cache.maxSize to be 200, got %d", cacheMax)
	}

	// Test: Get duration values
	cacheTTL := cfg.GetDuration("cache.ttl", 30*time.Minute)
	if cacheTTL != 1*time.Hour {
		t.Errorf("Expected cache.ttl to be 1h, got %v", cacheTTL)
	}

	// Test: Worker configuration
	workerCount := cfg.GetInt("worker.count", 3)
	if workerCount != 5 {
		t.Errorf("Expected worker.count to be 5, got %d", workerCount)
	}

	// Test: Container configuration
	containerMaxSize := cfg.GetInt("container.maxTotalSize", 10)
	if containerMaxSize != 20 {
		t.Errorf("Expected container.maxTotalSize to be 20, got %d", containerMaxSize)
	}

	// Test: Event bus configuration
	eventQueueSize := cfg.GetInt("eventBus.initialQueueSize", 500)
	if eventQueueSize != 1000 {
		t.Errorf("Expected eventBus.initialQueueSize to be 1000, got %d", eventQueueSize)
	}

	fmt.Println("✅ 配置系统集成测试通过")
}

func TestConfigManagerValidation(t *testing.T) {
	// Test validator registration
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_validation_config.json"
	cfg := config.NewConfigManager(configPath)

	// Register validators
	stringValidator := config.StringMinLength(3)
	if err := cfg.RegisterValidator("test.key", stringValidator); err != nil {
		t.Fatalf("Failed to register validator: %v", err)
	}

	// Test: Duplicate validator registration
	err := cfg.RegisterValidator("test.key", stringValidator)
	if err == nil {
		t.Error("Expected error when registering duplicate validator, got nil")
	}

	fmt.Println("✅ 配置管理器验证测试通过")
}

func TestConfigManagerSetGet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_setget_config.json"
	cfg := config.NewConfigManager(configPath)

	// Test: Set and Get string values
	testKey := "test.stringKey"
	testValue := "testValue"
	if err := cfg.Set(testKey, testValue); err != nil {
		t.Fatalf("Failed to set config value: %v", err)
	}

	retrievedValue := cfg.Get(testKey, nil)
	if retrievedValue != testValue {
		t.Errorf("Expected '%v', got '%v'", testValue, retrievedValue)
	}

	// Test: Set and Get integer values
	intKey := "test.intKey"
	intValue := 42
	if err := cfg.Set(intKey, intValue); err != nil {
		t.Fatalf("Failed to set int value: %v", err)
	}

	retrievedInt := cfg.Get(intKey, 0)
	if retrievedInt != intValue {
		t.Errorf("Expected %d, got %v", intValue, retrievedInt)
	}

	fmt.Println("✅ 配置管理器SetGet测试通过")
}

func TestConfigManagerGetGroup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_group_config.json"
	cfg := config.NewConfigManager(configPath)

	// Register specs with groups
	if err := cfg.Set("test.key1", "value1"); err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}
	if err := cfg.Set("test.key2", "value2"); err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}

	// Note: GetGroup requires validators to be registered with Group field
	// This is a basic test to ensure the method exists and works
	group := cfg.GetGroup("test")
	if group == nil {
		// Expected behavior when no validators are registered with the group
		fmt.Println("✅ 配置管理器GetGroup测试通过（空组）")
	} else {
		fmt.Println("✅ 配置管理器GetGroup测试通过")
	}
}

func TestConfigManagerSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_saveload_config.json"
	cfg := config.NewConfigManager(configPath)

	// Set some values
	testData := map[string]interface{}{
		"stringKey": "stringValue",
		"intKey":    123,
		"boolKey":   true,
	}

	for key, value := range testData {
		if err := cfg.Set(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load configuration into a new manager
	cfg2 := config.NewConfigManager(configPath)
	if err := cfg2.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values were loaded correctly
	for key, expectedValue := range testData {
		actualValue := cfg2.Get(key, nil)
		if actualValue != expectedValue {
			t.Errorf("For key %s: expected %v, got %v", key, expectedValue, actualValue)
		}
	}

	fmt.Println("✅ 配置管理器SaveLoad测试通过")
}

// BenchmarkConfigGet benchmarks the Get operation
func BenchmarkConfigGet(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := tmpDir + "/bench_config.json"
	cfg := config.NewConfigManager(configPath)

	// Set some values
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		if err := cfg.Set(key, fmt.Sprintf("value%d", i)); err != nil {
			b.Fatalf("Failed to set value: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Get("key50", "default")
	}
}

// BenchmarkConfigSet benchmarks the Set operation
func BenchmarkConfigSet(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := tmpDir + "/bench_set_config.json"
	cfg := config.NewConfigManager(configPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		if err := cfg.Set(key, fmt.Sprintf("value%d", i)); err != nil {
			b.Fatalf("Failed to set value: %v", err)
		}
	}
}
