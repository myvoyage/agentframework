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
	"testing"
	"time"
)

func TestTaskTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		taskType TaskType
		expected string
	}{
		{"Epic", TaskTypeEpic, "epic"},
		{"Task", TaskTypeTask, "task"},
		{"Bug", TaskTypeBug, "bug"},
		{"Feature", TaskTypeFeature, "feature"},
		{"Research", TaskTypeResearch, "research"},
		{"Checkpoint", TaskTypeCheckpoint, "checkpoint"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.taskType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.taskType))
			}
		})
	}
}

func TestTaskStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected string
	}{
		{"Open", StatusOpen, "open"},
		{"InProgress", StatusInProgress, "in_progress"},
		{"Blocked", StatusBlocked, "blocked"},
		{"Completed", StatusCompleted, "completed"},
		{"Cancelled", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestDependencyTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		depType  DependencyType
		expected string
	}{
		{"Blocks", DependencyTypeBlocks, "blocks"},
		{"ParentChild", DependencyTypeParentChild, "parent-child"},
		{"Related", DependencyTypeRelated, "related"},
		{"DiscoveredFrom", DependencyTypeDiscoveredFrom, "discovered-from"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.depType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.depType))
			}
		})
	}
}

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  string
	}{
		{"TaskCreated", EventTaskCreated, "task_created"},
		{"TaskUpdated", EventTaskUpdated, "task_updated"},
		{"TaskClosed", EventTaskClosed, "task_closed"},
		{"DependencyAdded", EventDependencyAdded, "dependency_added"},
		{"DependencyRemoved", EventDependencyRemoved, "dependency_removed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.eventType))
			}
		})
	}
}

func TestTaskModel(t *testing.T) {
	now := time.Now()
	task := &Task{
		ID:          "test-id",
		Type:        TaskTypeTask,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      StatusOpen,
		Assignee:    "test-user",
		Tags:        []string{"tag1", "tag2"},
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: map[string]string{
			"key1": "value1",
		},
	}

	if task.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", task.ID)
	}
	if task.Type != TaskTypeTask {
		t.Errorf("expected Type 'task', got %s", task.Type)
	}
	if task.Title != "Test Task" {
		t.Errorf("expected Title 'Test Task', got %s", task.Title)
	}
	if task.Status != StatusOpen {
		t.Errorf("expected Status 'open', got %s", task.Status)
	}
	if len(task.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(task.Tags))
	}
	if task.Metadata["key1"] != "value1" {
		t.Errorf("expected metadata key1='value1', got %s", task.Metadata["key1"])
	}
}

func TestDependencyModel(t *testing.T) {
	now := time.Now()
	dep := &Dependency{
		FromTaskID: "task1",
		ToTaskID:   "task2",
		Type:       DependencyTypeBlocks,
		CreatedAt:  now,
	}

	if dep.FromTaskID != "task1" {
		t.Errorf("expected FromTaskID 'task1', got %s", dep.FromTaskID)
	}
	if dep.ToTaskID != "task2" {
		t.Errorf("expected ToTaskID 'task2', got %s", dep.ToTaskID)
	}
	if dep.Type != DependencyTypeBlocks {
		t.Errorf("expected Type 'blocks', got %s", dep.Type)
	}
}

func TestEventModel(t *testing.T) {
	now := time.Now().Unix()
	event := &Event{
		Type:      EventTaskCreated,
		TaskID:    "test-task",
		Timestamp: now,
		Data: map[string]interface{}{
			"title": "Test Task",
			"type":  "task",
		},
	}

	if event.Type != EventTaskCreated {
		t.Errorf("expected Type 'task_created', got %s", event.Type)
	}
	if event.TaskID != "test-task" {
		t.Errorf("expected TaskID 'test-task', got %s", event.TaskID)
	}
	if event.Data["title"] != "Test Task" {
		t.Errorf("expected Data title 'Test Task', got %v", event.Data["title"])
	}
}

func TestConfigModel(t *testing.T) {
	config := &Config{
		StoragePath:  ".beads",
		GitEnabled:   true,
		SyncInterval: 5 * time.Second,
		MaxTasks:     10000,
		DBPath:       ".beads/beads.db",
		JSONLPath:    ".beads/tasks",
	}

	if config.StoragePath != ".beads" {
		t.Errorf("expected StoragePath '.beads', got %s", config.StoragePath)
	}
	if !config.GitEnabled {
		t.Error("expected GitEnabled to be true")
	}
	if config.SyncInterval != 5*time.Second {
		t.Errorf("expected SyncInterval 5s, got %v", config.SyncInterval)
	}
	if config.MaxTasks != 10000 {
		t.Errorf("expected MaxTasks 10000, got %d", config.MaxTasks)
	}
}

func TestErrorModel(t *testing.T) {
	now := time.Now()
	err := &Error{
		Code:      "TEST_ERROR",
		Message:   "Test error message",
		Timestamp: now,
		Retryable: true,
		Details: map[string]interface{}{
			"field": "value",
		},
	}

	if err.Code != "TEST_ERROR" {
		t.Errorf("expected Code 'TEST_ERROR', got %s", err.Code)
	}
	if err.Message != "Test error message" {
		t.Errorf("expected Message 'Test error message', got %s", err.Message)
	}
	if !err.Retryable {
		t.Error("expected Retryable to be true")
	}
	if err.Error() != "Test error message" {
		t.Errorf("expected Error() to return 'Test error message', got %s", err.Error())
	}
}

func TestLogicalOpConstants(t *testing.T) {
	if LogicalOpAND != "AND" {
		t.Errorf("expected LogicalOpAND 'AND', got %s", LogicalOpAND)
	}
	if LogicalOpOR != "OR" {
		t.Errorf("expected LogicalOpOR 'OR', got %s", LogicalOpOR)
	}
}

func TestDirectionConstants(t *testing.T) {
	if DirectionIncoming != "incoming" {
		t.Errorf("expected DirectionIncoming 'incoming', got %s", DirectionIncoming)
	}
	if DirectionOutgoing != "outgoing" {
		t.Errorf("expected DirectionOutgoing 'outgoing', got %s", DirectionOutgoing)
	}
}
