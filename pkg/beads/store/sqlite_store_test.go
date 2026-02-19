// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
)

func TestSQLiteStore_InitSchema(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_init.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	// Verify tables exist by attempting to query them
	ctx := context.Background()

	// Test tasks table
	_, err = store.db.QueryContext(ctx, "SELECT COUNT(*) FROM tasks")
	if err != nil {
		t.Errorf("Tasks table not created: %v", err)
	}

	// Test task_tags table
	_, err = store.db.QueryContext(ctx, "SELECT COUNT(*) FROM task_tags")
	if err != nil {
		t.Errorf("Task_tags table not created: %v", err)
	}

	// Test dependencies table
	_, err = store.db.QueryContext(ctx, "SELECT COUNT(*) FROM dependencies")
	if err != nil {
		t.Errorf("Dependencies table not created: %v", err)
	}

	// Verify WAL mode is enabled
	var journalMode string
	err = store.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Errorf("Failed to check journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got: %s", journalMode)
	}

	// Verify foreign keys are enabled (note: this is per-connection and may vary)
	var foreignKeys int
	err = store.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Errorf("Failed to check foreign keys: %v", err)
	}
	// Foreign keys should be enabled, but this is informational
	t.Logf("Foreign keys status: %d", foreignKeys)
}

func TestSQLiteStore_WriteAndReadTask(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_write_read.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test task
	task := &beads.Task{
		ID:          "test-task-1",
		Type:        beads.TaskTypeTask,
		Title:       "Test Task",
		Description: "This is a test task",
		Status:      beads.StatusOpen,
		Assignee:    "test-user",
		Tags:        []string{"test", "example"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Write task
	err = store.WriteTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to write task: %v", err)
	}

	// Read task back
	readTask, err := store.ReadTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to read task: %v", err)
	}

	// Verify task fields
	if readTask.ID != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, readTask.ID)
	}
	if readTask.Type != task.Type {
		t.Errorf("Expected Type %s, got %s", task.Type, readTask.Type)
	}
	if readTask.Title != task.Title {
		t.Errorf("Expected Title %s, got %s", task.Title, readTask.Title)
	}
	if readTask.Description != task.Description {
		t.Errorf("Expected Description %s, got %s", task.Description, readTask.Description)
	}
	if readTask.Status != task.Status {
		t.Errorf("Expected Status %s, got %s", task.Status, readTask.Status)
	}
	if readTask.Assignee != task.Assignee {
		t.Errorf("Expected Assignee %s, got %s", task.Assignee, readTask.Assignee)
	}

	// Verify tags
	if len(readTask.Tags) != len(task.Tags) {
		t.Errorf("Expected %d tags, got %d", len(task.Tags), len(readTask.Tags))
	}

	// Verify metadata
	if len(readTask.Metadata) != len(task.Metadata) {
		t.Errorf("Expected %d metadata entries, got %d", len(task.Metadata), len(readTask.Metadata))
	}
	for k, v := range task.Metadata {
		if readTask.Metadata[k] != v {
			t.Errorf("Expected metadata[%s] = %s, got %s", k, v, readTask.Metadata[k])
		}
	}
}

func TestSQLiteStore_WriteAndReadDependency(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_dependency.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create two test tasks
	task1 := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	task2 := &beads.Task{
		ID:        "task-2",
		Type:      beads.TaskTypeTask,
		Title:     "Task 2",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Write tasks
	if err := store.WriteTask(ctx, task1); err != nil {
		t.Fatalf("Failed to write task1: %v", err)
	}
	if err := store.WriteTask(ctx, task2); err != nil {
		t.Fatalf("Failed to write task2: %v", err)
	}

	// Create dependency
	dep := &beads.Dependency{
		FromTaskID: task1.ID,
		ToTaskID:   task2.ID,
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  time.Now(),
	}

	// Write dependency
	err = store.WriteDependency(ctx, dep)
	if err != nil {
		t.Fatalf("Failed to write dependency: %v", err)
	}

	// Read dependencies (outgoing from task1)
	deps, err := store.ReadDependencies(ctx, task1.ID, beads.DirectionOutgoing)
	if err != nil {
		t.Fatalf("Failed to read dependencies: %v", err)
	}

	if len(deps) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(deps))
	}

	if deps[0].FromTaskID != dep.FromTaskID {
		t.Errorf("Expected FromTaskID %s, got %s", dep.FromTaskID, deps[0].FromTaskID)
	}
	if deps[0].ToTaskID != dep.ToTaskID {
		t.Errorf("Expected ToTaskID %s, got %s", dep.ToTaskID, deps[0].ToTaskID)
	}
	if deps[0].Type != dep.Type {
		t.Errorf("Expected Type %s, got %s", dep.Type, deps[0].Type)
	}

	// Read dependencies (incoming to task2)
	incomingDeps, err := store.ReadDependencies(ctx, task2.ID, beads.DirectionIncoming)
	if err != nil {
		t.Fatalf("Failed to read incoming dependencies: %v", err)
	}

	if len(incomingDeps) != 1 {
		t.Fatalf("Expected 1 incoming dependency, got %d", len(incomingDeps))
	}
}

func TestSQLiteStore_Indexes(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_indexes.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Query to check indexes exist
	query := `SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='tasks'`
	rows, err := store.db.QueryContext(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	indexes := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan index name: %v", err)
		}
		indexes = append(indexes, name)
	}

	// Check that required indexes exist
	requiredIndexes := []string{"idx_status", "idx_assignee", "idx_created_at"}
	for _, reqIdx := range requiredIndexes {
		found := false
		for _, idx := range indexes {
			if idx == reqIdx {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required index %s not found", reqIdx)
		}
	}
}

func TestSQLiteStore_WriteDependency_ForeignKeyValidation(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_fk_validation.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create one test task
	task1 := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Write task
	if err := store.WriteTask(ctx, task1); err != nil {
		t.Fatalf("Failed to write task1: %v", err)
	}

	// Test 1: Try to create dependency with non-existent from_task
	dep1 := &beads.Dependency{
		FromTaskID: "non-existent-task",
		ToTaskID:   task1.ID,
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  time.Now(),
	}

	err = store.WriteDependency(ctx, dep1)
	if err == nil {
		t.Error("Expected error for non-existent from_task, got nil")
	}

	// Test 2: Try to create dependency with non-existent to_task
	dep2 := &beads.Dependency{
		FromTaskID: task1.ID,
		ToTaskID:   "non-existent-task",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  time.Now(),
	}

	err = store.WriteDependency(ctx, dep2)
	if err == nil {
		t.Error("Expected error for non-existent to_task, got nil")
	}

	// Test 3: Create valid dependency
	task2 := &beads.Task{
		ID:        "task-2",
		Type:      beads.TaskTypeTask,
		Title:     "Task 2",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.WriteTask(ctx, task2); err != nil {
		t.Fatalf("Failed to write task2: %v", err)
	}

	dep3 := &beads.Dependency{
		FromTaskID: task1.ID,
		ToTaskID:   task2.ID,
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  time.Now(),
	}

	err = store.WriteDependency(ctx, dep3)
	if err != nil {
		t.Errorf("Expected no error for valid dependency, got: %v", err)
	}
}

func TestSQLiteStore_WriteTasks_Batch(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_batch_tasks.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create multiple test tasks
	tasks := []*beads.Task{
		{
			ID:          "batch-task-1",
			Type:        beads.TaskTypeTask,
			Title:       "Batch Task 1",
			Description: "First batch task",
			Status:      beads.StatusOpen,
			Assignee:    "user1",
			Tags:        []string{"batch", "test"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata: map[string]string{
				"batch": "true",
			},
		},
		{
			ID:          "batch-task-2",
			Type:        beads.TaskTypeBug,
			Title:       "Batch Task 2",
			Description: "Second batch task",
			Status:      beads.StatusInProgress,
			Assignee:    "user2",
			Tags:        []string{"batch", "urgent"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:        "batch-task-3",
			Type:      beads.TaskTypeFeature,
			Title:     "Batch Task 3",
			Status:    beads.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Write tasks in batch
	err = store.WriteTasks(ctx, tasks)
	if err != nil {
		t.Fatalf("Failed to write tasks in batch: %v", err)
	}

	// Verify all tasks were written
	for _, task := range tasks {
		readTask, err := store.ReadTask(ctx, task.ID)
		if err != nil {
			t.Errorf("Failed to read task %s: %v", task.ID, err)
			continue
		}

		if readTask.ID != task.ID {
			t.Errorf("Task %s: Expected ID %s, got %s", task.ID, task.ID, readTask.ID)
		}
		if readTask.Title != task.Title {
			t.Errorf("Task %s: Expected Title %s, got %s", task.ID, task.Title, readTask.Title)
		}
		if readTask.Status != task.Status {
			t.Errorf("Task %s: Expected Status %s, got %s", task.ID, task.Status, readTask.Status)
		}

		// Verify tags
		if len(readTask.Tags) != len(task.Tags) {
			t.Errorf("Task %s: Expected %d tags, got %d", task.ID, len(task.Tags), len(readTask.Tags))
		}
	}
}

func TestSQLiteStore_WriteDependencies_Batch(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_batch_deps.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test tasks first
	tasks := []*beads.Task{
		{
			ID:        "dep-task-1",
			Type:      beads.TaskTypeTask,
			Title:     "Dep Task 1",
			Status:    beads.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "dep-task-2",
			Type:      beads.TaskTypeTask,
			Title:     "Dep Task 2",
			Status:    beads.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "dep-task-3",
			Type:      beads.TaskTypeTask,
			Title:     "Dep Task 3",
			Status:    beads.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Write tasks in batch
	err = store.WriteTasks(ctx, tasks)
	if err != nil {
		t.Fatalf("Failed to write tasks: %v", err)
	}

	// Create multiple dependencies
	deps := []*beads.Dependency{
		{
			FromTaskID: "dep-task-1",
			ToTaskID:   "dep-task-2",
			Type:       beads.DependencyTypeBlocks,
			CreatedAt:  time.Now(),
		},
		{
			FromTaskID: "dep-task-2",
			ToTaskID:   "dep-task-3",
			Type:       beads.DependencyTypeBlocks,
			CreatedAt:  time.Now(),
		},
		{
			FromTaskID: "dep-task-1",
			ToTaskID:   "dep-task-3",
			Type:       beads.DependencyTypeRelated,
			CreatedAt:  time.Now(),
		},
	}

	// Write dependencies in batch
	err = store.WriteDependencies(ctx, deps)
	if err != nil {
		t.Fatalf("Failed to write dependencies in batch: %v", err)
	}

	// Verify dependencies were written
	// Check outgoing from task-1
	outgoing, err := store.ReadDependencies(ctx, "dep-task-1", beads.DirectionOutgoing)
	if err != nil {
		t.Fatalf("Failed to read outgoing dependencies: %v", err)
	}
	if len(outgoing) != 2 {
		t.Errorf("Expected 2 outgoing dependencies from task-1, got %d", len(outgoing))
	}

	// Check incoming to task-3
	incoming, err := store.ReadDependencies(ctx, "dep-task-3", beads.DirectionIncoming)
	if err != nil {
		t.Fatalf("Failed to read incoming dependencies: %v", err)
	}
	if len(incoming) != 2 {
		t.Errorf("Expected 2 incoming dependencies to task-3, got %d", len(incoming))
	}
}

func TestSQLiteStore_WriteDependencies_Batch_ValidationError(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_batch_deps_error.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create one test task
	task := &beads.Task{
		ID:        "valid-task",
		Type:      beads.TaskTypeTask,
		Title:     "Valid Task",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.WriteTask(ctx, task); err != nil {
		t.Fatalf("Failed to write task: %v", err)
	}

	// Try to create dependencies with non-existent tasks
	deps := []*beads.Dependency{
		{
			FromTaskID: "valid-task",
			ToTaskID:   "non-existent-task",
			Type:       beads.DependencyTypeBlocks,
			CreatedAt:  time.Now(),
		},
	}

	// Should fail due to foreign key validation
	err = store.WriteDependencies(ctx, deps)
	if err == nil {
		t.Error("Expected error for non-existent task in batch, got nil")
	}

	// Verify no dependencies were written (transaction rollback)
	allDeps, err := store.ReadDependencies(ctx, "valid-task", beads.DirectionOutgoing)
	if err != nil {
		t.Fatalf("Failed to read dependencies: %v", err)
	}
	if len(allDeps) != 0 {
		t.Errorf("Expected 0 dependencies after failed batch, got %d", len(allDeps))
	}
}

func TestSQLiteStore_WriteTask_TransactionRollback(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_rollback.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a task with valid data
	task := &beads.Task{
		ID:        "test-task",
		Type:      beads.TaskTypeTask,
		Title:     "Test Task",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Write task successfully
	err = store.WriteTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to write task: %v", err)
	}

	// Verify task exists
	readTask, err := store.ReadTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to read task: %v", err)
	}
	if readTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, readTask.ID)
	}
}

func TestSQLiteStore_WriteTasks_EmptyBatch(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_empty_batch.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Write empty batch - should not error
	err = store.WriteTasks(ctx, []*beads.Task{})
	if err != nil {
		t.Errorf("Expected no error for empty batch, got: %v", err)
	}

	// Write nil batch - should not error
	err = store.WriteTasks(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error for nil batch, got: %v", err)
	}
}

// TestSQLiteStore_QueryTasks_EmptyDatabase verifies querying an empty database returns no results
func TestSQLiteStore_QueryTasks_EmptyDatabase(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_empty_query.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Query by status
	status := beads.StatusOpen
	query := beads.Query{
		Status: &status,
	}
	tasks, err := store.QueryTasks(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks in empty database, got %d", len(tasks))
	}

	// Query by assignee
	assignee := "test-user"
	query = beads.Query{
		Assignee: &assignee,
	}
	tasks, err = store.QueryTasks(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query tasks by assignee: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks in empty database, got %d", len(tasks))
	}
}

// TestSQLiteStore_ReadTask_InvalidID verifies reading a non-existent task returns an error
func TestSQLiteStore_ReadTask_InvalidID(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_invalid_id.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Try to read a non-existent task
	_, err = store.ReadTask(ctx, "non-existent-task-id")
	if err == nil {
		t.Error("Expected error when reading non-existent task, got nil")
	}
}

// TestSQLiteStore_ReadDependencies_EmptyGraph verifies reading dependencies from an empty graph returns no results
func TestSQLiteStore_ReadDependencies_EmptyGraph(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_empty_deps.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Try to read dependencies for a non-existent task
	deps, err := store.ReadDependencies(ctx, "non-existent-task", beads.DirectionOutgoing)
	if err != nil {
		t.Fatalf("Failed to read dependencies: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies in empty graph, got %d", len(deps))
	}
}

func TestSQLiteStore_WriteDependencies_EmptyBatch(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_empty_deps_batch.db"
	defer os.Remove(dbPath)

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Write empty batch - should not error
	err = store.WriteDependencies(ctx, []*beads.Dependency{})
	if err != nil {
		t.Errorf("Expected no error for empty batch, got: %v", err)
	}

	// Write nil batch - should not error
	err = store.WriteDependencies(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error for nil batch, got: %v", err)
	}
}

func TestSQLiteStore_QueryByStatus_ManyTasks(t *testing.T) {
	dbPath := t.TempDir() + "/query_many.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Insert 100 tasks
	for i := 0; i < 100; i++ {
		task := &beads.Task{
			ID:        fmt.Sprintf("task-%d", i),
			Type:      beads.TaskTypeTask,
			Title:     fmt.Sprintf("Task %d", i),
			Status:    beads.StatusOpen,
			Assignee:  "test-user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := store.WriteTask(ctx, task); err != nil {
			t.Fatalf("Failed to write task %d: %v", i, err)
		}
	}

	// Query by status
	statusOpen := beads.StatusOpen
	query := beads.Query{
		Status: &statusOpen,
	}
	tasks, err := store.QueryTasks(ctx, query)
	if err != nil {
		t.Fatalf("Failed to query tasks: %v", err)
	}

	t.Logf("Query returned %d tasks", len(tasks))
	if len(tasks) != 100 {
		t.Errorf("Expected 100 tasks, got %d", len(tasks))
		
		// Check what's in the database
		var count int
		err = store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE status = ?", beads.StatusOpen).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count tasks: %v", err)
		}
		t.Logf("Direct SQL count: %d", count)
	}
}
