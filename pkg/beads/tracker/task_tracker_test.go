package tracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"AgentFramework/pkg/beads"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskTracker_Lifecycle tests the tracker start/stop lifecycle
func TestTaskTracker_Lifecycle(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)
	require.NotNil(t, tracker)

	// Start tracker
	err := tracker.Start(ctx)
	require.NoError(t, err)

	// Stop tracker
	err = tracker.Stop(ctx)
	require.NoError(t, err)
}

// TestTaskTracker_CreateTask tests creating a task
func TestTaskTracker_CreateTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create a task
	task := &beads.Task{
		Type:        beads.TaskTypeTask,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      beads.StatusOpen,
		Assignee:    "test-user",
		Tags:        []string{"test", "example"},
		Metadata:    map[string]string{"key": "value"},
	}

	taskID, err := tracker.CreateTask(ctx, task)
	require.NoError(t, err)
	require.NotEmpty(t, taskID)

	// Verify task was created
	retrievedTask, err := tracker.GetTask(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, taskID, retrievedTask.ID)
	assert.Equal(t, beads.TaskTypeTask, retrievedTask.Type)
	assert.Equal(t, "Test Task", retrievedTask.Title)
	assert.Equal(t, "Test Description", retrievedTask.Description)
	assert.Equal(t, beads.StatusOpen, retrievedTask.Status)
	assert.Equal(t, "test-user", retrievedTask.Assignee)
	assert.ElementsMatch(t, []string{"test", "example"}, retrievedTask.Tags)
}

// TestTaskTracker_UpdateTask tests updating a task
func TestTaskTracker_UpdateTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create a task
	task := &beads.Task{
		Type:     beads.TaskTypeTask,
		Title:    "Original Title",
		Status:   beads.StatusOpen,
		Assignee: "user1",
	}

	taskID, err := tracker.CreateTask(ctx, task)
	require.NoError(t, err)

	// Update task
	newTitle := "Updated Title"
	newStatus := beads.StatusInProgress
	newAssignee := "user2"

	err = tracker.UpdateTask(ctx, taskID, beads.TaskUpdate{
		Title:    &newTitle,
		Status:   &newStatus,
		Assignee: &newAssignee,
	})
	require.NoError(t, err)

	// Verify updates
	retrievedTask, err := tracker.GetTask(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedTask.Title)
	assert.Equal(t, beads.StatusInProgress, retrievedTask.Status)
	assert.Equal(t, "user2", retrievedTask.Assignee)
}

// TestTaskTracker_CloseTask tests closing a task
func TestTaskTracker_CloseTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create a task
	task := &beads.Task{
		Type:     beads.TaskTypeTask,
		Title:    "Test Task",
		Status:   beads.StatusOpen,
		Assignee: "test-user",
	}

	taskID, err := tracker.CreateTask(ctx, task)
	require.NoError(t, err)

	// Close task
	err = tracker.CloseTask(ctx, taskID, beads.StatusCompleted)
	require.NoError(t, err)

	// Verify task was closed
	retrievedTask, err := tracker.GetTask(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, beads.StatusCompleted, retrievedTask.Status)
}

// TestTaskTracker_GetReady tests getting ready tasks
func TestTaskTracker_GetReady(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create task 1 (no dependencies - should be ready)
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, err := tracker.CreateTask(ctx, task1)
	require.NoError(t, err)

	// Create task 2 (no dependencies - should be ready)
	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, err := tracker.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Create task 3 (blocked by task 1 - should not be ready)
	task3 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 3",
		Status: beads.StatusOpen,
	}
	task3ID, err := tracker.CreateTask(ctx, task3)
	require.NoError(t, err)

	err = tracker.AddDependency(ctx, task1ID, task3ID, beads.DependencyTypeBlocks)
	require.NoError(t, err)

	// Get ready tasks
	readyTasks, err := tracker.GetReady(ctx)
	require.NoError(t, err)

	// Should have 2 ready tasks (task1 and task2)
	readyTaskIDs := make(map[string]bool)
	for _, task := range readyTasks {
		readyTaskIDs[task.ID] = true
	}

	assert.True(t, readyTaskIDs[task1ID], "Task 1 should be ready")
	assert.True(t, readyTaskIDs[task2ID], "Task 2 should be ready")
	assert.False(t, readyTaskIDs[task3ID], "Task 3 should not be ready")
}

// TestTaskTracker_AddDependency tests adding dependencies
func TestTaskTracker_AddDependency(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create two tasks
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, err := tracker.CreateTask(ctx, task1)
	require.NoError(t, err)

	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, err := tracker.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Add dependency
	err = tracker.AddDependency(ctx, task1ID, task2ID, beads.DependencyTypeBlocks)
	require.NoError(t, err)

	// Verify dependency was added
	deps, err := tracker.GetDependencies(ctx, task1ID)
	require.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, task1ID, deps[0].FromTaskID)
	assert.Equal(t, task2ID, deps[0].ToTaskID)
	assert.Equal(t, beads.DependencyTypeBlocks, deps[0].Type)
}

// TestTaskTracker_CycleDetection tests cycle detection
func TestTaskTracker_CycleDetection(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create two tasks
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, err := tracker.CreateTask(ctx, task1)
	require.NoError(t, err)

	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, err := tracker.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Add dependency task1 -> task2
	err = tracker.AddDependency(ctx, task1ID, task2ID, beads.DependencyTypeBlocks)
	require.NoError(t, err)

	// Try to add dependency task2 -> task1 (should fail - cycle!)
	err = tracker.AddDependency(ctx, task2ID, task1ID, beads.DependencyTypeBlocks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

// TestTaskTracker_QueryOperations tests various query operations
func TestTaskTracker_QueryOperations(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create tasks with different properties
	task1 := &beads.Task{
		Type:     beads.TaskTypeTask,
		Title:    "Task 1",
		Status:   beads.StatusOpen,
		Assignee: "user1",
		Tags:     []string{"frontend", "urgent"},
	}
	_, _ = tracker.CreateTask(ctx, task1)

	task2 := &beads.Task{
		Type:     beads.TaskTypeBug,
		Title:    "Task 2",
		Status:   beads.StatusInProgress,
		Assignee: "user2",
		Tags:     []string{"backend", "urgent"},
	}
	task2ID, _ := tracker.CreateTask(ctx, task2)

	task3 := &beads.Task{
		Type:     beads.TaskTypeFeature,
		Title:    "Task 3",
		Status:   beads.StatusOpen,
		Assignee: "user1",
		Tags:     []string{"frontend"},
	}
	_, _ = tracker.CreateTask(ctx, task3)

	// Test GetByStatus
	openTasks, err := tracker.GetByStatus(ctx, beads.StatusOpen)
	require.NoError(t, err)
	assert.Len(t, openTasks, 2)

	// Test GetByAssignee
	user1Tasks, err := tracker.GetByAssignee(ctx, "user1")
	require.NoError(t, err)
	assert.Len(t, user1Tasks, 2)

	// Test GetByTags (AND - should return only task2)
	urgentTasks, err := tracker.GetByTags(ctx, []string{"urgent", "backend"}, beads.LogicalOpAND)
	require.NoError(t, err)
	assert.Len(t, urgentTasks, 1)
	assert.Equal(t, task2ID, urgentTasks[0].ID)

	// Test GetByTags (OR - should return task1 and task2)
	urgentOrBackend, err := tracker.GetByTags(ctx, []string{"frontend", "backend"}, beads.LogicalOpOR)
	require.NoError(t, err)
	assert.Len(t, urgentOrBackend, 3)

	// Test GetByTags (single tag)
	frontendTasks, err := tracker.GetByTags(ctx, []string{"frontend"}, beads.LogicalOpAND)
	require.NoError(t, err)
	assert.Len(t, frontendTasks, 2)
}

// TestTaskTracker_RemoveDependency tests removing dependencies
func TestTaskTracker_RemoveDependency(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create two tasks
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, err := tracker.CreateTask(ctx, task1)
	require.NoError(t, err)

	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, err := tracker.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Add dependency
	err = tracker.AddDependency(ctx, task1ID, task2ID, beads.DependencyTypeBlocks)
	require.NoError(t, err)

	// Verify dependency exists
	deps, _ := tracker.GetDependencies(ctx, task1ID)
	assert.Len(t, deps, 1)

	// Remove dependency
	err = tracker.RemoveDependency(ctx, task1ID, task2ID)
	require.NoError(t, err)

	// Verify dependency was removed
	deps, _ = tracker.GetDependencies(ctx, task1ID)
	assert.Len(t, deps, 0)
}

// TestTaskTracker_HistoricalTaskLoading tests loading historical tasks from JSONL
func TestTaskTracker_HistoricalTaskLoading(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a tracker and add a task
	config1 := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  true,
	}

	tracker1 := NewTaskTracker(config1)

	err := tracker1.Start(ctx)
	require.NoError(t, err)

	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Historical Task",
		Status: beads.StatusOpen,
	}
	taskID, err := tracker1.CreateTask(ctx, task)
	require.NoError(t, err)

	err = tracker1.Stop(ctx)
	require.NoError(t, err)

	// Delete SQLite database to simulate a fresh start
	// Check if file exists first
	if _, err := os.Stat(config1.DBPath); err == nil {
		os.Remove(config1.DBPath)
	}

	// Create a new tracker with the same storage path
	// It should load historical tasks from JSONL
	config2 := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  true,
	}

	tracker2 := NewTaskTracker(config2)

	err = tracker2.Start(ctx)
	require.NoError(t, err)
	defer tracker2.Stop(ctx)

	// Verify historical task was loaded
	retrievedTask, err := tracker2.GetTask(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "Historical Task", retrievedTask.Title)
}

// TestTaskTracker_AutoCreateDirectories tests automatic directory creation
func TestTaskTracker_AutoCreateDirectories(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Use a non-existent subdirectory
	storagePath := filepath.Join(tmpDir, "nested", "dirs", "beads")

	config := &beads.Config{
		StoragePath: storagePath,
		DBPath:      filepath.Join(storagePath, "test.db"),
		JSONLPath:   filepath.Join(storagePath, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Verify directories were created
	info, err := os.Stat(storagePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify tracker works
	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Test Task",
		Status: beads.StatusOpen,
	}
	_, err = tracker.CreateTask(ctx, task)
	require.NoError(t, err)
}

// TestTaskTracker_ConcurrentAccess tests concurrent access to the tracker
func TestTaskTracker_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "test.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := NewTaskTracker(config)

	err := tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create multiple tasks concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			task := &beads.Task{
				Type:  beads.TaskTypeTask,
				Title: fmt.Sprintf("Concurrent Task %d", index),
				Status: beads.StatusOpen,
			}
			_, err := tracker.CreateTask(ctx, task)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tasks were created
	tasks, err := tracker.GetByStatus(ctx, beads.StatusOpen)
	require.NoError(t, err)
	assert.Len(t, tasks, 10)
}
