// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package beads

import (
	"context"
	"time"
)

// TaskTracker is the central coordinator managing task lifecycle, dependencies, and storage
type TaskTracker interface {
	// Task Operations
	CreateTask(ctx context.Context, task *Task) (string, error)
	UpdateTask(ctx context.Context, taskID string, updates TaskUpdate) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	CloseTask(ctx context.Context, taskID string, status TaskStatus) error

	// Query Operations
	GetReady(ctx context.Context) ([]*Task, error)
	GetByStatus(ctx context.Context, status TaskStatus) ([]*Task, error)
	GetByAssignee(ctx context.Context, assignee string) ([]*Task, error)
	GetByTags(ctx context.Context, tags []string, op LogicalOp) ([]*Task, error)

	// Dependency Operations
	AddDependency(ctx context.Context, fromID, toID string, depType DependencyType) error
	RemoveDependency(ctx context.Context, fromID, toID string) error
	GetDependencies(ctx context.Context, taskID string) ([]*Dependency, error)
	GetDependents(ctx context.Context, taskID string) ([]*Dependency, error)

	// Context Operations
	// Context is a forward declaration for the context.Context type from pkg/beads/context
	// CreateTaskWithContext creates a task and associates it with a context
	// Returns task ID and context ID
	CreateTaskWithContext(ctx context.Context, task *Task, ctxt interface{}) (string, string, error)
	// GetTaskContexts retrieves all contexts associated with a task
	GetTaskContexts(ctx context.Context, taskID string) (interface{}, error)
	// AssociateContext associates an existing context with a task
	AssociateContext(ctx context.Context, taskID, contextID string) error
	// DissociateContext removes the association between a task and a context
	DissociateContext(ctx context.Context, taskID, contextID string) error

	// Enhanced Context Operations (支持三层模型和记忆)
	// IsContextEnabled 检查上下文功能是否启用
	IsContextEnabled() bool
	// EnableContext 启用上下文功能
	EnableContext(ctx context.Context) error
	// DisableContext 禁用上下文功能
	DisableContext(ctx context.Context) error
	// GetContextStore 获取上下文存储（如果支持）
	GetContextStore() interface{}

	// 三层上下文操作
	// GetTaskContextWithLayer 获取任务的指定层级上下文
	GetTaskContextWithLayer(ctx context.Context, taskID string, layer interface{}) (interface{}, error)
	// GenerateTaskContextLayers 为任务的上下文生成缺失的层级
	GenerateTaskContextLayers(ctx context.Context, taskID string) error

	// 记忆操作
	// ExtractTaskMemories 从任务相关上下文中提取记忆
	ExtractTaskMemories(ctx context.Context, taskID string) (interface{}, error)
	// GetTaskMemories 获取任务的记忆
	GetTaskMemories(ctx context.Context, taskID string, memoryTypes []interface{}) (interface{}, error)

	// 联合查询
	// QueryTasksWithFullContext 查询任务及其完整上下文
	QueryTasksWithFullContext(ctx context.Context, query Query) (interface{}, error)

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Sync(ctx context.Context) error
}

// SQLiteStore provides fast local query engine with indexed access
type SQLiteStore interface {
	WriteTask(ctx context.Context, task *Task) error
	ReadTask(ctx context.Context, taskID string) (*Task, error)
	QueryTasks(ctx context.Context, query Query) ([]*Task, error)
	WriteDependency(ctx context.Context, dep *Dependency) error
	DeleteDependency(ctx context.Context, fromID, toID string) error
	ReadDependencies(ctx context.Context, taskID string, direction Direction) ([]*Dependency, error)
	RebuildFromEvents(ctx context.Context, events []*Event) error
	Close() error
}

// JSONLStore provides Git-tracked append-only event log
type JSONLStore interface {
	AppendEvent(ctx context.Context, event *Event) error
	ReadEvents(ctx context.Context, since time.Time) ([]*Event, error)
	ReadAllEvents(ctx context.Context) ([]*Event, error)
	GetLatestTimestamp(ctx context.Context) (time.Time, error)
	ForceFlush() error // For testing purposes
	Close() error
}

// SyncDaemon provides bidirectional synchronization between SQLite and JSONL
type SyncDaemon interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	TriggerSync(ctx context.Context) error
	GetStatus() SyncStatus
}

// DependencyResolver computes task ready state and validates dependency graphs
type DependencyResolver interface {
	ComputeReadyState(ctx context.Context, taskID string) (bool, error)
	ValidateNoCycles(ctx context.Context, fromID, toID string, depType DependencyType) error
	GetBlockingTasks(ctx context.Context, taskID string) ([]*Task, error)
	GetDependencyChain(ctx context.Context, taskID string) ([]*Task, error)
}

// EventProcessor processes events and applies them to storage layers
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *Event) error
	ReplayEvents(ctx context.Context, events []*Event) error
}
