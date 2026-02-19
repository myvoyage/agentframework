// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package memory_test

import (
	"testing"
	"time"

	"AgentFramework/pkg/beads/context"
)

// TestMemoryTierTypes 测试记忆分层类型
func TestMemoryTierTypes(t *testing.T) {
	tiers := []context.MemoryTier{
		context.MemoryTierSession,
		context.MemoryTierDaily,
		context.MemoryTierLongTerm,
	}

	for _, tier := range tiers {
		if tier == "" {
			t.Errorf("MemoryTier should not be empty")
		}
	}
}

// TestMemoryCompressionConfig 测试记忆压缩配置
func TestMemoryCompressionConfig(t *testing.T) {
	config := context.DefaultMemoryCompressionConfig()

	if config == nil {
		t.Fatalf("DefaultMemoryCompressionConfig should not return nil")
	}

	if config.SessionRetentionDuration != 24*time.Hour {
		t.Errorf("Expected SessionRetentionDuration to be 24h, got %v", config.SessionRetentionDuration)
	}

	if config.MaxSessionMemories != 100 {
		t.Errorf("Expected MaxSessionMemories to be 100, got %d", config.MaxSessionMemories)
	}
}

// TestMemoryCollection 测试记忆集合结构
func TestMemoryCollection(t *testing.T) {
	collection := &context.MemoryCollection{}

	if collection == nil {
		t.Fatalf("MemoryCollection should not return nil")
	}

	// 测试初始化空集合
	if collection.Profiles != nil {
		t.Errorf("New collection should have nil Profiles")
	}
	if collection.Preferences != nil {
		t.Errorf("New collection should have nil Preferences")
	}
	if collection.Entities != nil {
		t.Errorf("New collection should have nil Entities")
	}
	if collection.Events != nil {
		t.Errorf("New collection should have nil Events")
	}
	if collection.Cases != nil {
		t.Errorf("New collection should have nil Cases")
	}
	if collection.Patterns != nil {
		t.Errorf("New collection should have nil Patterns")
	}
}
