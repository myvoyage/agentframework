package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/tracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskTrackerMCP(t *testing.T) {
	mcp := NewTaskTrackerMCP(&mockTaskTracker{})
	require.NotNil(t, mcp)
	assert.NotNil(t, mcp.tracker)
}

// mockTaskTracker is a minimal mock for testing
type mockTaskTracker struct{}

func (m *mockTaskTracker) CreateTask(ctx context.Context, task *beads.Task) (string, error) {
	return "mock-task-id", nil
}

func (m *mockTaskTracker) UpdateTask(ctx context.Context, taskID string, updates beads.TaskUpdate) error {
	return nil
}

func (m *mockTaskTracker) GetTask(ctx context.Context, taskID string) (*beads.Task, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockTaskTracker) CloseTask(ctx context.Context, taskID string, status beads.TaskStatus) error {
	return nil
}

func (m *mockTaskTracker) GetReady(ctx context.Context) ([]*beads.Task, error) {
	return nil, nil
}

func (m *mockTaskTracker) GetByStatus(ctx context.Context, status beads.TaskStatus) ([]*beads.Task, error) {
	return nil, nil
}

func (m *mockTaskTracker) GetByAssignee(ctx context.Context, assignee string) ([]*beads.Task, error) {
	return nil, nil
}

func (m *mockTaskTracker) GetByTags(ctx context.Context, tags []string, op beads.LogicalOp) ([]*beads.Task, error) {
	return nil, nil
}

func (m *mockTaskTracker) AddDependency(ctx context.Context, fromID, toID string, depType beads.DependencyType) error {
	return nil
}

func (m *mockTaskTracker) RemoveDependency(ctx context.Context, fromID, toID string) error {
	return nil
}

func (m *mockTaskTracker) GetDependencies(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	return nil, nil
}

func (m *mockTaskTracker) GetDependents(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	return nil, nil
}

func (m *mockTaskTracker) Start(ctx context.Context) error {
	return nil
}

func (m *mockTaskTracker) Stop(ctx context.Context) error {
	return nil
}

func (m *mockTaskTracker) Sync(ctx context.Context) error {
	return nil
}

func TestTaskTrackerMCP_CreateTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create and start tracker
	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Test create task
	input := CreateTaskInput{
		Type:  "task",
		Title: "Test Task from MCP",
		Status: "open",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.CreateTask(ctx, params)
	require.NoError(t, err)

	var output CreateTaskOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
	assert.NotEmpty(t, output.TaskID)
}

func TestTaskTrackerMCP_CreateTaskValidationError(t *testing.T) {
	ctx := context.Background()
	mcp := NewTaskTrackerMCP(nil) // No tracker, will error

	// Invalid JSON
	_, err := mcp.CreateTask(ctx, []byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestTaskTrackerMCP_ShowTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create a task first
	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Test Task",
		Status: beads.StatusOpen,
	}
	taskID, _ := tracker.CreateTask(ctx, task)

	// Show the task
	input := ShowTaskInput{
		TaskID: taskID,
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.ShowTask(ctx, params)
	require.NoError(t, err)

	var output ShowTaskOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
	assert.NotNil(t, output.Task)
	assert.Equal(t, "Test Task", output.Task.Title)
}

func TestTaskTrackerMCP_ShowTaskNotFound(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Try to show non-existent task
	input := ShowTaskInput{
		TaskID: "non-existent-task",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.ShowTask(ctx, params)
	require.NoError(t, err)

	var output ShowTaskOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.False(t, output.Success)
	assert.NotEmpty(t, output.Error)
}

func TestTaskTrackerMCP_UpdateTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create a task first
	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Original Title",
		Status: beads.StatusOpen,
	}
	taskID, _ := tracker.CreateTask(ctx, task)

	// Update the task
	newTitle := "Updated Title"
	newStatus := "in_progress"

	input := UpdateTaskInput{
		TaskID: taskID,
		Title:  &newTitle,
		Status: &newStatus,
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.UpdateTask(ctx, params)
	require.NoError(t, err)

	var output UpdateTaskOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)

	// Verify update
	updatedTask, _ := tracker.GetTask(ctx, taskID)
	assert.Equal(t, "Updated Title", updatedTask.Title)
	assert.Equal(t, beads.StatusInProgress, updatedTask.Status)
}

func TestTaskTrackerMCP_CloseTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create a task first
	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Test Task",
		Status: beads.StatusOpen,
	}
	taskID, _ := tracker.CreateTask(ctx, task)

	// Close the task
	input := CloseTaskInput{
		TaskID: taskID,
		Status: "completed",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.CloseTask(ctx, params)
	require.NoError(t, err)

	var output CloseTaskOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
}

func TestTaskTrackerMCP_AddDependency(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create two tasks
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, _ := tracker.CreateTask(ctx, task1)

	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, _ := tracker.CreateTask(ctx, task2)

	// Add dependency
	input := AddDependencyInput{
		FromTaskID: task1ID,
		ToTaskID:   task2ID,
		DepType:    "blocks",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.AddDependency(ctx, params)
	require.NoError(t, err)

	var output AddDependencyOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
}

func TestTaskTrackerMCP_AddDependencyCycle(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create two tasks
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	}
	task1ID, _ := tracker.CreateTask(ctx, task1)

	task2 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 2",
		Status: beads.StatusOpen,
	}
	task2ID, _ := tracker.CreateTask(ctx, task2)

	// Add first dependency
	_ = tracker.AddDependency(ctx, task1ID, task2ID, beads.DependencyTypeBlocks)

	// Try to add reverse dependency (should fail - cycle!)
	input := AddDependencyInput{
		FromTaskID: task2ID,
		ToTaskID:   task1ID,
		DepType:    "blocks",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.AddDependency(ctx, params)
	require.NoError(t, err)

	var output AddDependencyOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.False(t, output.Success)
	assert.Contains(t, output.Error, "cycle")
}

func TestTaskTrackerMCP_ListTasks(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create some tasks with different statuses
	_, _ = tracker.CreateTask(ctx, &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Task 1",
		Status: beads.StatusOpen,
	})

	_, _ = tracker.CreateTask(ctx, &beads.Task{
		Type:  beads.TaskTypeBug,
		Title: "Bug 1",
		Status: beads.StatusInProgress,
	})

	// List tasks by status (should return 1 task)
	input := ListTasksInput{
		Status: "in_progress",
	}

	params, err := json.Marshal(input)
	require.NoError(t, err)

	result, err := mcp.ListTasks(ctx, params)
	require.NoError(t, err)

	var output ListTasksOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
	assert.NotNil(t, output.Tasks)
	assert.Len(t, output.Tasks, 1)
}

func TestTaskTrackerMCP_GetReadyTasks(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tracker := setupTestTracker(t, tmpDir)
	mcp := NewTaskTrackerMCP(tracker)

	// Create a task (should be ready since no dependencies)
	task := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Ready Task",
		Status: beads.StatusOpen,
	}
	_, _ = tracker.CreateTask(ctx, task)

	// Get ready tasks
	result, err := mcp.GetReadyTasks(ctx, nil)
	require.NoError(t, err)

	var output GetReadyTasksOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.True(t, output.Success)
	assert.Len(t, output.Tasks, 1)
}

// Helper function to set up a test tracker
func setupTestTracker(t *testing.T, tmpDir string) beads.TaskTracker {
	t.Helper()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
	}

	tracker := tracker.NewTaskTracker(config)
	ctx := context.Background()

	err := tracker.Start(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		tracker.Stop(ctx)
	})

	return tracker
}
