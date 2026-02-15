// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.context.Validate()
			if (result.Valid == tt.wantErr) {
				t.Errorf("Validate() = %v, wantErr %v", result, tt.wantErr)
			}
		})
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
