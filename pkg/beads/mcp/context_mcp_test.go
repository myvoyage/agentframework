// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package mcp

import (
	stdcontext "context"
	"testing"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/context"
)

// MockContextTracker 模拟支持上下文的 TaskTracker
type MockContextTracker struct {
	tasks           map[string]*beads.Task
	contexts        map[string][]*context.Context
	createError      error
	getContextsError error
}

func NewMockContextTracker() *MockContextTracker {
	return &MockContextTracker{
		tasks:    make(map[string]*beads.Task),
		contexts: make(map[string][]*context.Context),
	}
}

func (m *MockContextTracker) CreateTask(ctx context.Context, task *beads.Task) (string, error) {
	if m.createError != nil {
		return "", m.createError
	}
	task.ID = "mock-task-" + task.Title
	m.tasks[task.ID] = task
	return task.ID, nil
}

func (m *MockContextTracker) CreateTaskWithContext(
	ctx context.Context,
	task *beads.Task,
	ctxt interface{},
) (string, string, error) {
	taskID, err := m.CreateTask(ctx, task)
	if err != nil {
		return "", "", err
	}

	// 类型断言获取 Context
	c, ok := ctxt.(*context.Context)
	if !ok {
		// 如果不是 Context 类型，创建一个默认的 Context
		c = &context.Context{
			ID:    "mock-context-" + taskID,
			Title: task.Title,
		}
	}

	contextID := c.ID
	if contextID == "" {
		contextID = "mock-context-" + taskID
		c.ID = contextID
	}

	if m.contexts[taskID] == nil {
		m.contexts[taskID] = []*context.Context{c}
	} else {
		m.contexts[taskID] = append(m.contexts[taskID], c)
	}

	return taskID, contextID, nil
}

func (m *MockContextTracker) GetTaskContexts(ctx context.Context, taskID string) ([]*context.Context, error) {
	if m.getContextsError != nil {
		return nil, m.getContextsError
	}
	return m.contexts[taskID], nil
}

func (m *MockContextTracker) AssociateContext(ctx context.Context, taskID, contextID string) error {
	m.contexts[taskID] = append(m.contexts[taskID], &context.Context{ID: contextID})
	return nil
}

func (m *MockContextTracker) DissociateContext(ctx context.Context, taskID, contextID string) error {
	contexts := m.contexts[taskID]
	for i, ctxt := range contexts {
		if ctxt.ID == contextID {
			m.contexts[taskID] = append(contexts[:i], contexts[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockContextTracker) IsContextEnabled() bool {
	return true
}

func (m *MockContextTracker) GetContextStore() context.ContextStore {
	return nil
}

// 实现 TaskTracker 的其他方法
func (m *MockContextTracker) UpdateTask(ctx context.Context, taskID string, updates beads.TaskUpdate) error {
	if task, ok := m.tasks[taskID]; ok {
		if updates.Title != nil {
			task.Title = *updates.Title
		}
		if updates.Description != nil {
			task.Description = *updates.Description
		}
		if updates.Metadata != nil {
			task.Metadata = *updates.Metadata
		}
	}
	return nil
}

func (m *MockContextTracker) GetTask(ctx context.Context, taskID string) (*beads.Task, error) {
	return m.tasks[taskID], nil
}

func (m *MockContextTracker) CloseTask(ctx context.Context, taskID string, status beads.TaskStatus) error {
	if task, ok := m.tasks[taskID]; ok {
		task.Status = status
	}
	return nil
}

func (m *MockContextTracker) GetReady(ctx context.Context) ([]*beads.Task, error) {
	return nil, nil
}

func (m *MockContextTracker) GetByStatus(ctx context.Context, status beads.TaskStatus) ([]*beads.Task, error) {
	return nil, nil
}

func (m *MockContextTracker) GetByAssignee(ctx context.Context, assignee string) ([]*beads.Task, error) {
	return nil, nil
}

func (m *MockContextTracker) GetByTags(ctx context.Context, tags []string, op beads.LogicalOp) ([]*beads.Task, error) {
	return nil, nil
}

func (m *MockContextTracker) AddDependency(ctx context.Context, fromID, toID string, depType beads.DependencyType) error {
	return nil
}

func (m *MockContextTracker) RemoveDependency(ctx context.Context, fromID, toID string) error {
	return nil
}

func (m *MockContextTracker) GetDependencies(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	return nil, nil
}

func (m *MockContextTracker) GetDependents(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	return nil, nil
}

func (m *MockContextTracker) Start(ctx context.Context) error {
	return nil
}

func (m *MockContextTracker) Stop(ctx context.Context) error {
	return nil
}

func (m *MockContextTracker) Sync(ctx context.Context) error {
	return nil
}

// TestContextMCP_CreateTaskWithContext 测试创建带上下文的任务
func TestContextMCP_CreateTaskWithContext(t *testing.T) {
	ctx := stdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	tests := []struct {
		name    string
		input   *CreateTaskWithContextInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &CreateTaskWithContextInput{
				Type:        "task",
				Title:       "Test Task",
				Description: "Test Description",
				ContextType: "project",
				Workspace:   "/test",
				ContextMeta: map[string]string{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			input: &CreateTaskWithContextInput{
				Type:        "task",
				ContextType: "project",
			},
			wantErr: true,
		},
		{
			name: "missing context type",
			input: &CreateTaskWithContextInput{
				Type:  "task",
				Title: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := mcp.CreateTaskWithContext(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTaskWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !output.Success {
					t.Errorf("CreateTaskWithContext() Success = false, want true")
				}
				if output.TaskID == "" {
					t.Errorf("CreateTaskWithContext() TaskID is empty")
				}
				if output.ContextID == "" {
					t.Errorf("CreateTaskWithContext() ContextID is empty")
				}
			}
		})
	}
}

// TestContextMCP_GetTaskContexts 测试获取任务上下文
func TestContextMCP_GetTaskContexts(t *testing.T) {
	ctx := stdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	// 创建测试数据
	taskID := "test-task-123"
	tracker.contexts[taskID] = []*context.Context{
		{ID: "ctx-1", Type: context.ContextTypeProject, Title: "Project 1"},
		{ID: "ctx-2", Type: context.ContextTypeFile, Title: "File 1"},
	}

	tests := []struct {
		name    string
		input   *GetTaskContextsInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &GetTaskContextsInput{
				TaskID: taskID,
			},
			wantErr: false,
		},
		{
			name: "missing task ID",
			input: &GetTaskContextsInput{
				TaskID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := mcp.GetTaskContexts(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskContexts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !output.Success {
					t.Errorf("GetTaskContexts() Success = false, want true")
				}
				if len(output.Contexts) != 2 {
					t.Errorf("GetTaskContexts() returned %d contexts, want 2", len(output.Contexts))
				}
			}
		})
	}
}

// TestContextMCP_AssociateContext 测试关联上下文
func TestContextMCP_AssociateContext(t *testing.T) {
	ctx := stdstdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	tests := []struct {
		name    string
		input   *AssociateContextInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &AssociateContextInput{
				TaskID:    "task-123",
				ContextID: "ctx-456",
			},
			wantErr: false,
		},
		{
			name: "missing task ID",
			input: &AssociateContextInput{
				ContextID: "ctx-456",
			},
			wantErr: true,
		},
		{
			name: "missing context ID",
			input: &AssociateContextInput{
				TaskID: "task-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := mcp.AssociateContext(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssociateContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !output.Success {
					t.Errorf("AssociateContext() Success = false, want true")
				}
			}
		})
	}
}

// TestContextMCP_GetContextTypes 测试获取上下文类型
func TestContextMCP_GetContextTypes(t *testing.T) {
	ctx := stdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	types, err := mcp.GetContextTypes(ctx)
	if err != nil {
		t.Fatalf("GetContextTypes() error = %v", err)
	}

	expectedTypes := []string{
		"project", "file", "codebase", "custom", "memory", "resource", "skill",
	}

	if len(types) != len(expectedTypes) {
		t.Errorf("GetContextTypes() returned %d types, want %d", len(types), len(expectedTypes))
	}

	for _, expectedType := range expectedTypes {
		if _, ok := types[expectedType]; !ok {
			t.Errorf("GetContextTypes() missing type: %s", expectedType)
		}
	}
}

// TestContextMCP_GetContextStoreInfo 测试获取上下文存储信息
func TestContextMCP_GetContextStoreInfo(t *testing.T) {
	ctx := stdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	info, err := mcp.GetContextStoreInfo(ctx)
	if err != nil {
		t.Fatalf("GetContextStoreInfo() error = %v", err)
	}

	enabled, ok := info["enabled"].(bool)
	if !ok {
		t.Error("GetContextStoreInfo() missing enabled field")
	}
	if !enabled {
		t.Error("GetContextStoreInfo() enabled = false, want true")
	}
}

// BenchmarkCreateTaskWithContext 性能测试：创建带上下文的任务
func BenchmarkCreateTaskWithContext(b *testing.B) {
	ctx := stdcontext.Background()
	tracker := NewMockContextTracker()
	mcp := NewContextMCP(tracker)

	input := &CreateTaskWithContextInput{
		Type:        "task",
		Title:       "Benchmark Task",
		ContextType: "project",
		Workspace:   "/bench",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mcp.CreateTaskWithContext(ctx, input)
	}
}
