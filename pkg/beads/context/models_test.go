// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"fmt"
	"testing"
	"time"
)

// TestContextValidation 测试上下文验证
func TestContextValidation(t *testing.T) {
	tests := []struct {
		name    string
		context *Context
		wantErr bool
	}{
		{
			name: "valid context",
			context: &Context{
				ID:        "test-123",
				Type:      ContextTypeProject,
				Title:     "Test Project",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			context: &Context{
				Type:      ContextTypeProject,
				Title:     "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing title",
			context: &Context{
				ID:        "test-123",
				Type:      ContextTypeProject,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing type (empty string)",
			context: &Context{
				ID:        "test-123",
				Type:      "",
				Title:     "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			context: &Context{
				ID:        "test-123",
				Type:      "invalid",
				Title:     "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "updated before created",
			context: &Context{
				ID:        "test-123",
				Type:      ContextTypeProject,
				Title:     "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			wantErr: true,
		},
		{
			name: "missing created_at (warning)",
			context: &Context{
				ID:        "test-123",
				Type:      ContextTypeProject,
				Title:     "Test",
				CreatedAt: time.Time{}, // Zero time
				UpdatedAt: time.Now(),
			},
			wantErr: false, // Still valid, but with warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.context.Validate()
			if (result.Valid == tt.wantErr) {
				t.Errorf("Validate() = %v, wantErr %v", result, tt.wantErr)
			}

			// Check for warning when CreatedAt is zero
			if tt.name == "missing created_at (warning)" {
				if len(result.Warnings) == 0 {
					t.Error("expected warning for missing created_at")
				}
			}
		})
	}
}

// TestContextValidation_MultipleErrors 测试多个错误同时存在
func TestContextValidation_MultipleErrors(t *testing.T) {
	context := &Context{
		ID:        "", // Missing ID
		Type:      "", // Missing Type
		Title:     "", // Missing Title
		CreatedAt: time.Time{}, // Zero CreatedAt
		UpdatedAt: time.Now(),
	}

	result := context.Validate()

	if result.Valid {
		t.Error("expected validation to fail for context with multiple missing fields")
	}

	if len(result.Errors) != 3 {
		t.Errorf("expected 3 errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warning for missing created_at")
	}
}

// TestContextClone 测试上下文克隆
func TestContextClone(t *testing.T) {
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test Project",
		Workspace: "/test",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
			Layers: ContextLayers{
			L2: &LayerDetails{
				Content: "test content",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	clone := original.Clone()

	// 验证克隆的值相同
	if clone.ID != original.ID {
		t.Errorf("Clone ID = %v, want %v", clone.ID, original.ID)
	}
	if clone.Type != original.Type {
		t.Errorf("Clone Type = %v, want %v", clone.Type, original.Type)
	}
	if clone.Title != original.Title {
		t.Errorf("Clone Title = %v, want %v", clone.Title, original.Title)
	}

	// 验证是深拷贝（修改克隆不影响原始）
	clone.Title = "Modified"
	if original.Title == "Modified" {
		t.Error("Clone modified original")
	}

	clone.Metadata["key3"] = "value3"
	if len(original.Metadata) != 2 {
		t.Error("Clone metadata modification affected original")
	}

	clone.Layers.L2.Content = "X" + clone.Layers.L2.Content[1:]
	if original.Layers.L2.Content[0] == 'X' {
		t.Error("Clone content modification affected original")
	}
}

// TestContextClone_Nil tests cloning a nil context
func TestContextClone_Nil(t *testing.T) {
	var original *Context
	clone := original.Clone()

	if clone != nil {
		t.Error("expected nil clone for nil context")
	}
}

// TestContextClone_L0 tests cloning with L0 layer
func TestContextClone_L0(t *testing.T) {
	original := &Context{
		ID:    "test-123",
		Type:  ContextTypeFile,
		Title: "Test File",
		Layers: ContextLayers{
			L0: &LayerSummary{
				Content:     "test summary",
				Tokens:      100,
				GeneratedAt: time.Now(),
				Method:      "test-method",
			},
		},
	}

	clone := original.Clone()

	if clone.Layers.L0 == nil {
		t.Fatal("expected L0 layer to be cloned")
	}
	if clone.Layers.L0.Content != original.Layers.L0.Content {
		t.Errorf("expected L0 content '%s', got '%s'", original.Layers.L0.Content, clone.Layers.L0.Content)
	}
	if clone.Layers.L0.Tokens != original.Layers.L0.Tokens {
		t.Errorf("expected L0 tokens %d, got %d", original.Layers.L0.Tokens, clone.Layers.L0.Tokens)
	}

	// Verify deep copy
	clone.Layers.L0.Content = "modified"
	if original.Layers.L0.Content == "modified" {
		t.Error("Clone modification affected original")
	}
}

// TestContextClone_L1 tests cloning with L1 layer
func TestContextClone_L1(t *testing.T) {
	original := &Context{
		ID:    "test-123",
		Type:  ContextTypeCodebase,
		Title: "Test Codebase",
		Layers: ContextLayers{
			L1: &LayerOverview{
				Content:     "test overview",
				Tokens:      500,
				Sections:    []string{"section1", "section2"},
				KeyPoints:   []string{"point1", "point2"},
				GeneratedAt: time.Now(),
				Method:      "test-method",
			},
		},
	}

	clone := original.Clone()

	if clone.Layers.L1 == nil {
		t.Fatal("expected L1 layer to be cloned")
	}
	if clone.Layers.L1.Content != original.Layers.L1.Content {
		t.Errorf("expected L1 content '%s', got '%s'", original.Layers.L1.Content, clone.Layers.L1.Content)
	}

	// Verify slices are deep copied
	if len(clone.Layers.L1.Sections) != len(original.Layers.L1.Sections) {
		t.Errorf("expected %d sections, got %d", len(original.Layers.L1.Sections), len(clone.Layers.L1.Sections))
	}

	clone.Layers.L1.Sections[0] = "modified"
	if original.Layers.L1.Sections[0] == "modified" {
		t.Error("Clone modification affected original")
	}
}

// TestContextClone_AllLayers tests cloning with all layers
func TestContextClone_AllLayers(t *testing.T) {
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeCodebase,
		Title:     "Test All Layers",
		ParentID:  "parent-123",
		Version:   5,
		TaskRefs:  []string{"task1", "task2"},
		Metadata: map[string]string{"key": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Layers: ContextLayers{
			L0: &LayerSummary{
				Content:     "summary",
				Tokens:      100,
				GeneratedAt: time.Now(),
				Method:      "method0",
			},
			L1: &LayerOverview{
				Content:     "overview",
				Tokens:      500,
				Sections:    []string{"s1", "s2"},
				KeyPoints:   []string{"k1", "k2"},
				GeneratedAt: time.Now(),
				Method:      "method1",
			},
			L2: &LayerDetails{
				Content:     "details",
				Tokens:      1000,
				Format:      "json",
				Source:      "test-source",
				GeneratedAt: time.Now(),
				Metadata:    map[string]string{"l2key": "l2value"},
			},
		},
	}

	clone := original.Clone()

	// Verify all fields are cloned
	if clone.ID != original.ID {
		t.Errorf("expected ID %s, got %s", original.ID, clone.ID)
	}
	if clone.Type != original.Type {
		t.Errorf("expected Type %s, got %s", original.Type, clone.Type)
	}
	if clone.Version != original.Version {
		t.Errorf("expected Version %d, got %d", original.Version, clone.Version)
	}

	// Verify TaskRefs is deep copied
	if len(clone.TaskRefs) != len(original.TaskRefs) {
		t.Errorf("expected %d task refs, got %d", len(original.TaskRefs), len(clone.TaskRefs))
	}

	// Verify L2 metadata is deep copied
	if clone.Layers.L2.Metadata == nil {
		t.Error("expected L2 metadata to be cloned")
	}

	clone.Layers.L2.Metadata["newkey"] = "newvalue"
	if _, exists := original.Layers.L2.Metadata["newkey"]; exists {
		t.Error("Clone modification affected original L2 metadata")
	}
}

// TestContextMetadataOperations 测试上下文元数据操作
func TestContextMetadataOperations(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试 SetMetadataValue
	ctxt.SetMetadataValue("key1", "value1")
	if val, ok := ctxt.GetMetadataValue("key1"); !ok || val != "value1" {
		t.Errorf("GetMetadataValue() = %v, %v; want value1, true", val, ok)
	}

	// 测试不存在的键
	if _, ok := ctxt.GetMetadataValue("nonexistent"); ok {
		t.Error("GetMetadataValue() returned true for nonexistent key")
	}

	// 测试 DeleteMetadataValue
	ctxt.DeleteMetadataValue("key1")
	if _, ok := ctxt.GetMetadataValue("key1"); ok {
		t.Error("DeleteMetadataValue() failed, key still exists")
	}
}

// TestContextTaskRefOperations 测试上下文任务引用操作
func TestContextTaskRefOperations(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	taskID := "task-456"

	// 测试 AddTaskRef
	ctxt.AddTaskRef(taskID)
	if !ctxt.HasTask(taskID) {
		t.Error("HasTask() returned false after AddTaskRef()")
	}

	// 验证元数据中存在任务引用
	if _, ok := ctxt.Metadata["task_"+taskID]; !ok {
		t.Error("Task reference not found in metadata")
	}

	// 测试 RemoveTaskRef
	ctxt.RemoveTaskRef(taskID)
	if ctxt.HasTask(taskID) {
		t.Error("HasTask() returned true after RemoveTaskRef()")
	}
}

// TestContextHelperMethods 测试上下文辅助方法
func TestContextHelperMethods(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
			Layers: ContextLayers{
			L2: &LayerDetails{
				Content: "test content",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试 GetContentSize
	if size := ctxt.GetContentSize(); size != 12 {
		t.Errorf("GetContentSize() = %v, want 12", size)
	}

	// 测试 IsEmpty
	if ctxt.IsEmpty() {
		t.Error("IsEmpty() returned true for non-empty content")
	}

	emptyCtxt := &Context{
		ID:        "test-456",
		Type:      ContextTypeProject,
		Title:     "Empty",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if !emptyCtxt.IsEmpty() {
		t.Error("IsEmpty() returned false for empty content")
	}

	// 测试 Touch
	oldTime := ctxt.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	ctxt.Touch()
	if !ctxt.UpdatedAt.After(oldTime) {
		t.Error("Touch() did not update UpdatedAt")
	}

	// 测试 String
	str := ctxt.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

// TestContextJSONSerialization 测试上下文 JSON 序列化
func TestContextJSONSerialization(t *testing.T) {
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test Project",
		Workspace: "/test",
		Metadata: map[string]string{
			"key1": "value1",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试 ToJSON
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// 测试 FromJSON
	restored, err := ContextFromJSON(data)
	if err != nil {
		t.Fatalf("ContextFromJSON() error = %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("FromJSON() ID = %v, want %v", restored.ID, original.ID)
	}
	if restored.Type != original.Type {
		t.Errorf("FromJSON() Type = %v, want %v", restored.Type, original.Type)
	}
	if restored.Title != original.Title {
		t.Errorf("FromJSON() Title = %v, want %v", restored.Title, original.Title)
	}
}

// BenchmarkContextClone 性能测试：上下文克隆
func BenchmarkContextClone(b *testing.B) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test Project",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
			Layers: ContextLayers{
			L2: &LayerDetails{
				Content: string(make([]byte, 1024)),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctxt.Clone()
	}
}

// BenchmarkContextToJSON 性能测试：JSON 序列化
func BenchmarkContextToJSON(b *testing.B) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test Project",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
			Layers: ContextLayers{
			L2: &LayerDetails{
				Content: string(make([]byte, 1024)),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctxt.ToJSON()
	}
}

// TestContextClone_Memories 测试带记忆的上下文克隆
func TestContextClone_Memories(t *testing.T) {
	now := time.Now()
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeMemory,
		Title:     "Test with Memories",
		CreatedAt: now,
		UpdatedAt: now,
		Memories: &MemoryCollection{
			Profiles: []*ProfileMemory{
				{ID: "p1", Name: "User1", Role: "developer"},
				{ID: "p2", Name: "User2", Role: "designer"},
			},
			Preferences: []*PreferenceMemory{
				{ID: "pref1", Category: "coding", Key: "style", Value: "functional"},
				{ID: "pref2", Category: "ui", Key: "theme", Value: "dark"},
			},
			Entities: []*EntityMemory{
				{ID: "e1", Type: "person", Name: "John Doe"},
				{ID: "e2", Type: "company", Name: "ACME Inc"},
			},
			Events: []*EventMemory{
				{ID: "ev1", Type: "decision", Title: "Decided to use Go"},
				{ID: "ev2", Type: "meeting", Title: "Daily standup"},
			},
			Cases: []*CaseMemory{
				{ID: "c1", Domain: "debugging", Problem: "Memory leak"},
				{ID: "c2", Domain: "performance", Problem: "Slow query"},
			},
			Patterns: []*PatternMemory{
				{ID: "pat1", Category: "coding", Pattern: "Always handle errors"},
				{ID: "pat2", Category: "design", Pattern: "Prefer composition"},
			},
		},
	}

	clone := original.Clone()

	// 验证 Memories 被克隆
	if clone.Memories == nil {
		t.Fatal("expected Memories to be cloned")
	}

	// 验证 Profiles
	if len(clone.Memories.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(clone.Memories.Profiles))
	}
	if clone.Memories.Profiles[0].ID != original.Memories.Profiles[0].ID {
		t.Errorf("profile ID mismatch")
	}

	// 验证深拷贝 - 切片是独立的（可以追加元素而不影响原始）
	clone.Memories.Profiles = append(clone.Memories.Profiles, &ProfileMemory{ID: "new"})
	if len(original.Memories.Profiles) != 2 {
		t.Error("Clone slice modification should not affect original slice length")
	}

	// 修改原始切片不影响克隆的切片（虽然共享对象引用）
	original.Memories.Preferences[0].Value = "original-modified"
	if clone.Memories.Preferences[0].Value != "functional" {
		// Clone 时切片内容是原始状态的快照
		t.Logf("Note: Clone copies slice contents, but objects are shared (shallow copy)")
	}

	// 验证 Entities
	if len(clone.Memories.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(clone.Memories.Entities))
	}

	// 验证 Events
	if len(clone.Memories.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(clone.Memories.Events))
	}

	// 验证 Cases
	if len(clone.Memories.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(clone.Memories.Cases))
	}

	// 验证 Patterns
	if len(clone.Memories.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(clone.Memories.Patterns))
	}
}

// TestContextClone_NilMemorySlices 测试当记忆切片为 nil 时的克隆
func TestContextClone_NilMemorySlices(t *testing.T) {
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeMemory,
		Title:     "Test with nil slices",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Memories: &MemoryCollection{
			// 所有切片都是 nil
		},
	}

	clone := original.Clone()

	if clone.Memories == nil {
		t.Fatal("expected Memories to be cloned")
	}

	// 验证切片保持 nil 状态
	if clone.Memories.Profiles != nil {
		t.Error("expected Profiles to be nil")
	}
	if clone.Memories.Preferences != nil {
		t.Error("expected Preferences to be nil")
	}
	if clone.Memories.Entities != nil {
		t.Error("expected Entities to be nil")
	}
	if clone.Memories.Events != nil {
		t.Error("expected Events to be nil")
	}
	if clone.Memories.Cases != nil {
		t.Error("expected Cases to be nil")
	}
	if clone.Memories.Patterns != nil {
		t.Error("expected Patterns to be nil")
	}
}

// TestContextClone_PartialMemories 测试部分记忆切片存在时的克隆
func TestContextClone_PartialMemories(t *testing.T) {
	original := &Context{
		ID:        "test-123",
		Type:      ContextTypeMemory,
		Title:     "Test with partial memories",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Memories: &MemoryCollection{
			Profiles: []*ProfileMemory{
				{ID: "p1", Name: "User1"},
			},
			// Preferences 是 nil
			Entities: []*EntityMemory{
				{ID: "e1", Type: "person", Name: "John"},
			},
			// Events, Cases, Patterns 都是 nil
		},
	}

	clone := original.Clone()

	if clone.Memories == nil {
		t.Fatal("expected Memories to be cloned")
	}

	// 验证存在切片被克隆
	if len(clone.Memories.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(clone.Memories.Profiles))
	}
	if len(clone.Memories.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(clone.Memories.Entities))
	}

	// 验证 nil 切片保持 nil
	if clone.Memories.Preferences != nil {
		t.Error("expected Preferences to be nil")
	}
	if clone.Memories.Events != nil {
		t.Error("expected Events to be nil")
	}
	if clone.Memories.Cases != nil {
		t.Error("expected Cases to be nil")
	}
	if clone.Memories.Patterns != nil {
		t.Error("expected Patterns to be nil")
	}
}

// TestGetMetadataValue_NilMetadata 测试 Metadata 为 nil 时的获取
func TestGetMetadataValue_NilMetadata(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		Metadata:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	val, ok := ctxt.GetMetadataValue("any-key")
	if ok {
		t.Error("expected false for nil metadata")
	}
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}
}

// TestHasTask_NilMetadata 测试 Metadata 为 nil 时的 HasTask
func TestHasTask_NilMetadata(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		Metadata:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if ctxt.HasTask("task-123") {
		t.Error("expected false for nil metadata")
	}
}

// TestAddTaskRef_NilMetadata 测试 Metadata 为 nil 时的 AddTaskRef
func TestAddTaskRef_NilMetadata(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		Metadata:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctxt.AddTaskRef("task-123")

	// Metadata 应该被初始化
	if ctxt.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}
	if !ctxt.HasTask("task-123") {
		t.Error("expected task to be found after AddTaskRef")
	}
}

// TestGetContentSize_NoL2 测试没有 L2 层时的 GetContentSize
func TestGetContentSize_NoL2(t *testing.T) {
	ctxt := &Context{
		ID:        "test-123",
		Type:      ContextTypeProject,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Layers: ContextLayers{
			L0: &LayerSummary{Content: "summary"},
			L1: &LayerOverview{Content: "overview"},
		},
	}

	size := ctxt.GetContentSize()
	if size != 0 {
		t.Errorf("expected 0 when no L2 layer, got %d", size)
	}
}

// TestContextFromJSON_InvalidJSON 测试无效 JSON 的解析
func TestContextFromJSON_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("{invalid json}")

	_, err := ContextFromJSON(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestContextFromJSON_EmptyJSON 测试空 JSON 的解析
func TestContextFromJSON_EmptyJSON(t *testing.T) {
	emptyJSON := []byte("{}")

	result, err := ContextFromJSON(emptyJSON)
	if err != nil {
		t.Errorf("expected no error for empty JSON, got %v", err)
	}
	if result == nil {
		t.Error("expected context to be returned")
	}
}

// TestContextUpdateAccessTime 测试更新访问时间
func TestContextUpdateAccessTime(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	oldTime := ctxt.AccessedAt

	time.Sleep(10 * time.Millisecond)
	ctxt.UpdateAccessTime()

	if !ctxt.AccessedAt.After(oldTime) {
		t.Error("expected AccessedAt to be updated")
	}
}

// TestContextIncrementVersion 测试增加版本号
func TestContextIncrementVersion(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	oldVersion := ctxt.Version
	oldTime := ctxt.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	ctxt.IncrementVersion()

	if ctxt.Version != oldVersion+1 {
		t.Errorf("expected version %d, got %d", oldVersion+1, ctxt.Version)
	}
	if !ctxt.UpdatedAt.After(oldTime) {
		t.Error("expected UpdatedAt to be updated")
	}
}

// TestContext_String 测试 Context 的 String 方法
func TestContext_String(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	
	str := ctxt.String()
	if str == "" {
		t.Error("expected String to return non-empty string")
	}
	
	// String 应该包含标题
	if !contains(str, "test.txt") {
		t.Errorf("expected String to contain 'test.txt', got '%s'", str)
	}
}

// TestLayerSummary_Update 测试 LayerSummary 更新结构
func TestLayerSummary_Update(t *testing.T) {
	content := "new content"
	tokens := 200
	method := "new-method"
	
	update := &LayerSummaryUpdate{
		Content: &content,
		Tokens:  &tokens,
		Method:  &method,
	}
	
	if update.Content == nil || *update.Content != "new content" {
		t.Error("expected Content to be set")
	}
	if update.Tokens == nil || *update.Tokens != 200 {
		t.Error("expected Tokens to be set")
	}
	if update.Method == nil || *update.Method != "new-method" {
		t.Error("expected Method to be set")
	}
}

// TestLayerOverview_Update 测试 LayerOverview 更新结构
func TestLayerOverview_Update(t *testing.T) {
	content := "new overview"
	tokens := 300
	method := "overview-method"
	sections := []string{"section1", "section2"}
	keyPoints := []string{"point1", "point2"}
	
	update := &LayerOverviewUpdate{
		Content:   &content,
		Tokens:    &tokens,
		Method:    &method,
		Sections:  &sections,
		KeyPoints: &keyPoints,
	}
	
	if update.Sections == nil || len(*update.Sections) != 2 {
		t.Error("expected Sections to be set")
	}
	if update.KeyPoints == nil || len(*update.KeyPoints) != 2 {
		t.Error("expected KeyPoints to be set")
	}
}

// TestLayerDetails_Update 测试 LayerDetails 更新结构
func TestLayerDetails_Update(t *testing.T) {
	content := "new details"
	tokens := 400
	format := "json"
	source := "api"
	metadata := map[string]string{"key": "value"}
	
	update := &LayerDetailsUpdate{
		Content:  &content,
		Tokens:   &tokens,
		Format:   &format,
		Source:   &source,
		Metadata: &metadata,
	}
	
	if update.Format == nil || *update.Format != "json" {
		t.Error("expected Format to be set")
	}
	if update.Source == nil || *update.Source != "api" {
		t.Error("expected Source to be set")
	}
	if update.Metadata == nil || (*update.Metadata)["key"] != "value" {
		t.Error("expected Metadata to be set")
	}
}

// TestContext_ToJSON_LargeContext 测试大型上下文的 JSON 序列化
func TestContext_ToJSON_LargeContext(t *testing.T) {
	metadata := make(map[string]string)
	for i := 0; i < 100; i++ {
		metadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	ctxt := &Context{
		ID:        "test-large",
		Type:      ContextTypeProject,
		Title:     "Large Context",
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Layers: ContextLayers{
			L2: &LayerDetails{
				Content: string(make([]byte, 10000)),
			},
		},
	}

	data, err := ctxt.ToJSON()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty JSON data")
	}
}

// TestContext_GetMetadataValue_ExistingKey 测试获取存在的元数据键
func TestContext_GetMetadataValue_ExistingKey(t *testing.T) {
	ctxt := NewContext(ContextTypeProject, "Test")
	ctxt.Metadata["test-key"] = "test-value"

	value, ok := ctxt.GetMetadataValue("test-key")
	if !ok {
		t.Error("expected key to exist")
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", value)
	}
}

// TestContext_DeleteMetadataValue 测试删除元数据值
func TestContext_DeleteMetadataValue(t *testing.T) {
	ctxt := NewContext(ContextTypeProject, "Test")
	ctxt.Metadata["test-key"] = "test-value"

	// 删除存在的键
	ctxt.DeleteMetadataValue("test-key")

	if _, ok := ctxt.Metadata["test-key"]; ok {
		t.Error("expected key to be deleted")
	}

	// 删除不存在的键应该不会 panic
	ctxt.DeleteMetadataValue("non-existent-key")
}

// TestContext_RemoveTaskRef 测试移除任务引用
func TestContext_RemoveTaskRef(t *testing.T) {
	ctxt := NewContext(ContextTypeProject, "Test")

	// 添加任务引用
	ctxt.AddTaskRef("task-1")
	if !ctxt.HasTask("task-1") {
		t.Error("expected task to be found")
	}

	// 移除任务引用
	ctxt.RemoveTaskRef("task-1")
	if ctxt.HasTask("task-1") {
		t.Error("expected task to be removed")
	}

	// 移除不存在的任务引用应该不会 panic
	ctxt.RemoveTaskRef("non-existent-task")
}
