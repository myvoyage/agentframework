// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// Checkpoint Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"testing"
	"time"
)

func TestCheckpointStatusConstants(t *testing.T) {
	statuses := []CheckpointStatus{
		StatusRunning,
		StatusSuspended,
		StatusCompleted,
		StatusFailed,
	}

	for _, status := range statuses {
		if status == "" {
			t.Errorf("CheckpointStatus should not be empty")
		}
	}
}

func TestCheckpoint(t *testing.T) {
	now := time.Now()
	cp := &Checkpoint{
		RunID:        "run-1",
		WorkflowName: "test-workflow",
		Status:       StatusRunning,
		State: &WorkflowState{
			ID:         "state-1",
			WorkflowID: "workflow-1",
			Status:     "running",
		},
		Input:    "test input",
		Output:   "test output",
		Progress: 0.5,
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if cp.RunID != "run-1" {
		t.Errorf("expected RunID 'run-1', got '%s'", cp.RunID)
	}
	if cp.WorkflowName != "test-workflow" {
		t.Errorf("expected WorkflowName 'test-workflow', got '%s'", cp.WorkflowName)
	}
	if cp.Status != StatusRunning {
		t.Errorf("expected Status StatusRunning, got '%v'", cp.Status)
	}
	if cp.Progress != 0.5 {
		t.Errorf("expected Progress 0.5, got %v", cp.Progress)
	}
	if len(cp.Metadata) != 2 {
		t.Errorf("expected 2 metadata items, got %d", len(cp.Metadata))
	}
}

func TestCheckpointWithError(t *testing.T) {
	now := time.Now()
	cp := &Checkpoint{
		RunID:        "run-2",
		WorkflowName: "failed-workflow",
		Status:       StatusFailed,
		Error:        "something went wrong",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if cp.Status != StatusFailed {
		t.Errorf("expected Status StatusFailed, got '%v'", cp.Status)
	}
	if cp.Error != "something went wrong" {
		t.Errorf("expected Error 'something went wrong', got '%s'", cp.Error)
	}
}

func TestWithWorkflowName(t *testing.T) {
	opts := &ListCheckpointOptions{}
	WithWorkflowName("test-workflow")(opts)

	if opts.WorkflowName != "test-workflow" {
		t.Errorf("expected WorkflowName 'test-workflow', got '%s'", opts.WorkflowName)
	}
}

func TestWithStatus(t *testing.T) {
	opts := &ListCheckpointOptions{}
	WithStatus(StatusCompleted)(opts)

	if opts.Status != StatusCompleted {
		t.Errorf("expected Status StatusCompleted, got '%v'", opts.Status)
	}
}

func TestWithLimit(t *testing.T) {
	opts := &ListCheckpointOptions{}
	WithLimit(100)(opts)

	if opts.Limit != 100 {
		t.Errorf("expected Limit 100, got %d", opts.Limit)
	}
}

func TestWithOffset(t *testing.T) {
	opts := &ListCheckpointOptions{}
	WithOffset(50)(opts)

	if opts.Offset != 50 {
		t.Errorf("expected Offset 50, got %d", opts.Offset)
	}
}

func TestMultipleCheckpointOptions(t *testing.T) {
	opts := &ListCheckpointOptions{}
	WithWorkflowName("test-workflow")(opts)
	WithStatus(StatusRunning)(opts)
	WithLimit(50)(opts)
	WithOffset(10)(opts)

	if opts.WorkflowName != "test-workflow" {
		t.Errorf("expected WorkflowName 'test-workflow', got '%s'", opts.WorkflowName)
	}
	if opts.Status != StatusRunning {
		t.Errorf("expected Status StatusRunning, got '%v'", opts.Status)
	}
	if opts.Limit != 50 {
		t.Errorf("expected Limit 50, got %d", opts.Limit)
	}
	if opts.Offset != 10 {
		t.Errorf("expected Offset 10, got %d", opts.Offset)
	}
}

func TestCheckpointWithEmptyOutput(t *testing.T) {
	now := time.Now()
	cp := &Checkpoint{
		RunID:        "run-3",
		WorkflowName: "running-workflow",
		Status:       StatusRunning,
		Input:        "test input",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if cp.Output != "" {
		t.Errorf("expected empty Output, got '%s'", cp.Output)
	}
}

func TestCheckpointWithSuspendedStatus(t *testing.T) {
	now := time.Now()
	cp := &Checkpoint{
		RunID:        "run-4",
		WorkflowName: "suspended-workflow",
		Status:       StatusSuspended,
		Progress:     0.75,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if cp.Status != StatusSuspended {
		t.Errorf("expected Status StatusSuspended, got '%v'", cp.Status)
	}
	if cp.Progress != 0.75 {
		t.Errorf("expected Progress 0.75, got %v", cp.Progress)
	}
}
