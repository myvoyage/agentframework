package tracker

import (
	"context"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDependencyResolver_EmptyGraph tests dependency resolution with no dependencies
func TestDependencyResolver_EmptyGraph(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	// Create a task with no dependencies
	task := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, sqliteStore.WriteTask(ctx, task))

	// Task should be ready (no dependencies)
	ready, err := resolver.ComputeReadyState(ctx, "task-1")
	require.NoError(t, err)
	assert.True(t, ready, "Task with no dependencies should be ready")

	// No blocking tasks
	blocking, err := resolver.GetBlockingTasks(ctx, "task-1")
	require.NoError(t, err)
	assert.Empty(t, blocking, "Task with no dependencies should have no blocking tasks")
}

// TestDependencyResolver_SingleBlocking tests a single blocking dependency
func TestDependencyResolver_SingleBlocking(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create two tasks
	task1 := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusInProgress,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task2 := &beads.Task{
		ID:        "task-2",
		Type:      beads.TaskTypeTask,
		Title:     "Task 2",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sqliteStore.WriteTask(ctx, task1))
	require.NoError(t, sqliteStore.WriteTask(ctx, task2))

	// Add blocking dependency: task-1 blocks task-2
	dep := &beads.Dependency{
		FromTaskID: "task-1",
		ToTaskID:   "task-2",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}
	require.NoError(t, sqliteStore.WriteDependency(ctx, dep))

	// task-1 should be ready (no incoming dependencies)
	ready, err := resolver.ComputeReadyState(ctx, "task-1")
	require.NoError(t, err)
	assert.True(t, ready, "Task with no incoming dependencies should be ready")

	// task-2 should NOT be ready (blocked by task-1 which is not completed)
	ready, err = resolver.ComputeReadyState(ctx, "task-2")
	require.NoError(t, err)
	assert.False(t, ready, "Task blocked by in-progress task should not be ready")

	// Get blocking tasks for task-2
	blocking, err := resolver.GetBlockingTasks(ctx, "task-2")
	require.NoError(t, err)
	assert.Len(t, blocking, 1, "Should have 1 blocking task")
	assert.Equal(t, "task-1", blocking[0].ID)

	// Complete task-1
	task1.Status = beads.StatusCompleted
	task1.UpdatedAt = time.Now()
	require.NoError(t, sqliteStore.WriteTask(ctx, task1))

	// Now task-2 should be ready
	ready, err = resolver.ComputeReadyState(ctx, "task-2")
	require.NoError(t, err)
	assert.True(t, ready, "Task should be ready when blocking task is completed")

	// No blocking tasks now
	blocking, err = resolver.GetBlockingTasks(ctx, "task-2")
	require.NoError(t, err)
	assert.Empty(t, blocking)
}

// TestDependencyResolver_MultipleBlocking tests multiple blocking dependencies
func TestDependencyResolver_MultipleBlocking(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create three tasks: A, B, C where both A and B block C
	taskA := &beads.Task{
		ID:        "task-A",
		Type:      beads.TaskTypeTask,
		Title:     "Task A",
		Status:    beads.StatusInProgress,
		CreatedAt: now,
		UpdatedAt: now,
	}
	taskB := &beads.Task{
		ID:        "task-B",
		Type:      beads.TaskTypeTask,
		Title:     "Task B",
		Status:    beads.StatusCompleted,
		CreatedAt: now,
		UpdatedAt: now,
	}
	taskC := &beads.Task{
		ID:        "task-C",
		Type:      beads.TaskTypeTask,
		Title:     "Task C",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sqliteStore.WriteTask(ctx, taskA))
	require.NoError(t, sqliteStore.WriteTask(ctx, taskB))
	require.NoError(t, sqliteStore.WriteTask(ctx, taskC))

	// Add dependencies: A blocks C, B blocks C
	depAC := &beads.Dependency{
		FromTaskID: "task-A",
		ToTaskID:   "task-C",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}
	depBC := &beads.Dependency{
		FromTaskID: "task-B",
		ToTaskID:   "task-C",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}

	require.NoError(t, sqliteStore.WriteDependency(ctx, depAC))
	require.NoError(t, sqliteStore.WriteDependency(ctx, depBC))

	// Task C should NOT be ready (blocked by A which is in-progress)
	ready, err := resolver.ComputeReadyState(ctx, "task-C")
	require.NoError(t, err)
	assert.False(t, ready, "Task C should not be ready when A is in-progress")

	// Get blocking tasks - should only return A (B is completed)
	blocking, err := resolver.GetBlockingTasks(ctx, "task-C")
	require.NoError(t, err)
	assert.Len(t, blocking, 1, "Should have 1 blocking task (A is in-progress)")
	assert.Equal(t, "task-A", blocking[0].ID)

	// Complete task A
	taskA.Status = beads.StatusCompleted
	taskA.UpdatedAt = time.Now()
	require.NoError(t, sqliteStore.WriteTask(ctx, taskA))

	// Now task C should be ready (both A and B are completed)
	ready, err = resolver.ComputeReadyState(ctx, "task-C")
	require.NoError(t, err)
	assert.True(t, ready, "Task C should be ready when all blocking tasks are completed")
}

// TestDependencyResolver_SelfReference tests self-referential dependency detection
func TestDependencyResolver_SelfReference(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	// Try to validate a self-referential dependency
	err = resolver.ValidateNoCycles(ctx, "task-1", "task-1", beads.DependencyTypeBlocks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "self-referential")
}

// TestDependencyResolver_CycleDetection tests cycle detection in complex graphs
func TestDependencyResolver_CycleDetection(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create tasks: A -> B -> C
	for _, id := range []string{"task-A", "task-B", "task-C"} {
		task := &beads.Task{
			ID:        id,
			Type:      beads.TaskTypeTask,
			Title:     id,
			Status:    beads.StatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, sqliteStore.WriteTask(ctx, task))
	}

	// Add dependencies: A blocks B, B blocks C
	depAB := &beads.Dependency{
		FromTaskID: "task-A",
		ToTaskID:   "task-B",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}
	depBC := &beads.Dependency{
		FromTaskID: "task-B",
		ToTaskID:   "task-C",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}

	require.NoError(t, sqliteStore.WriteDependency(ctx, depAB))
	require.NoError(t, sqliteStore.WriteDependency(ctx, depBC))

	// Try to add C blocks A (would create cycle: A -> B -> C -> A)
	err = resolver.ValidateNoCycles(ctx, "task-C", "task-A", beads.DependencyTypeBlocks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")

	// Try to add C blocks B (would create cycle: B -> C -> B)
	err = resolver.ValidateNoCycles(ctx, "task-C", "task-B", beads.DependencyTypeBlocks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

// TestDependencyResolver_NonBlockingTypes tests that non-blocking types don't affect ready state
func TestDependencyResolver_NonBlockingTypes(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create two tasks
	task1 := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusInProgress,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task2 := &beads.Task{
		ID:        "task-2",
		Type:      beads.TaskTypeTask,
		Title:     "Task 2",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sqliteStore.WriteTask(ctx, task1))
	require.NoError(t, sqliteStore.WriteTask(ctx, task2))

	// Add "related" dependency (not blocking)
	dep := &beads.Dependency{
		FromTaskID: "task-1",
		ToTaskID:   "task-2",
		Type:       beads.DependencyTypeRelated,
		CreatedAt:  now,
	}
	require.NoError(t, sqliteStore.WriteDependency(ctx, dep))

	// task-2 should be ready (related is not blocking)
	ready, err := resolver.ComputeReadyState(ctx, "task-2")
	require.NoError(t, err)
	assert.True(t, ready, "Related dependency should not affect ready state")

	// No blocking tasks
	blocking, err := resolver.GetBlockingTasks(ctx, "task-2")
	require.NoError(t, err)
	assert.Empty(t, blocking, "Related dependency should not be considered blocking")
}

// TestDependencyResolver_HierarchicalStructure tests parent-child dependency handling
func TestDependencyResolver_HierarchicalStructure(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create hierarchical tasks: Epic -> Feature -> Task
	epic := &beads.Task{
		ID:        "epic-1",
		Type:      beads.TaskTypeEpic,
		Title:     "Epic 1",
		Status:    beads.StatusInProgress,
		CreatedAt: now,
		UpdatedAt: now,
	}
	feature := &beads.Task{
		ID:        "feature-1",
		Type:      beads.TaskTypeFeature,
		Title:     "Feature 1",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	task := &beads.Task{
		ID:        "task-1",
		Type:      beads.TaskTypeTask,
		Title:     "Task 1",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sqliteStore.WriteTask(ctx, epic))
	require.NoError(t, sqliteStore.WriteTask(ctx, feature))
	require.NoError(t, sqliteStore.WriteTask(ctx, task))

	// Add parent-child dependencies
	depEpicFeature := &beads.Dependency{
		FromTaskID: "epic-1",
		ToTaskID:   "feature-1",
		Type:       beads.DependencyTypeParentChild,
		CreatedAt:  now,
	}
	depFeatureTask := &beads.Dependency{
		FromTaskID: "feature-1",
		ToTaskID:   "task-1",
		Type:       beads.DependencyTypeParentChild,
		CreatedAt:  now,
	}

	require.NoError(t, sqliteStore.WriteDependency(ctx, depEpicFeature))
	require.NoError(t, sqliteStore.WriteDependency(ctx, depFeatureTask))

	// Task should not be ready (parent feature is not completed)
	ready, err := resolver.ComputeReadyState(ctx, "task-1")
	require.NoError(t, err)
	assert.False(t, ready, "Task should not be ready when parent is not completed")

	// Feature should not be ready (parent epic is not completed)
	ready, err = resolver.ComputeReadyState(ctx, "feature-1")
	require.NoError(t, err)
	assert.False(t, ready, "Feature should not be ready when parent is not completed")

	// Epic should be ready (no parent)
	ready, err = resolver.ComputeReadyState(ctx, "epic-1")
	require.NoError(t, err)
	assert.True(t, ready, "Epic should be ready (no parent)")

	// Get dependency chain for task
	chain, err := resolver.GetDependencyChain(ctx, "task-1")
	require.NoError(t, err)
	assert.Len(t, chain, 3, "Dependency chain should include all 3 tasks")
}

// TestDependencyResolver_GetDependencyChain tests full dependency graph traversal
func TestDependencyResolver_GetDependencyChain(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create a complex graph:
	// A -> B -> C
	//      ↓
	//      D
	taskIDs := []string{"task-A", "task-B", "task-C", "task-D"}
	for _, id := range taskIDs {
		task := &beads.Task{
			ID:        id,
			Type:      beads.TaskTypeTask,
			Title:     id,
			Status:    beads.StatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, sqliteStore.WriteTask(ctx, task))
	}

	// Add dependencies
	deps := []*beads.Dependency{
		{FromTaskID: "task-A", ToTaskID: "task-B", Type: beads.DependencyTypeBlocks, CreatedAt: now},
		{FromTaskID: "task-B", ToTaskID: "task-C", Type: beads.DependencyTypeBlocks, CreatedAt: now},
		{FromTaskID: "task-A", ToTaskID: "task-D", Type: beads.DependencyTypeBlocks, CreatedAt: now},
	}

	for _, dep := range deps {
		require.NoError(t, sqliteStore.WriteDependency(ctx, dep))
	}

	// Get dependency chain from task-C
	chain, err := resolver.GetDependencyChain(ctx, "task-C")
	require.NoError(t, err)

	// Should include A, B, C (and maybe D if it traverses both directions)
	chainIDs := make(map[string]bool)
	for _, task := range chain {
		chainIDs[task.ID] = true
	}

	assert.True(t, chainIDs["task-A"], "Chain should include task-A")
	assert.True(t, chainIDs["task-B"], "Chain should include task-B")
	assert.True(t, chainIDs["task-C"], "Chain should include task-C")
}

// TestDependencyValidator_CompletedDependentsAreNotBlocking tests that completed dependents don't block
func TestDependencyValidator_CompletedDependentsAreNotBlocking(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	sqliteStore, err := store.NewSQLiteStore(tmpDir + "/test.db")
	require.NoError(t, err)
	defer sqliteStore.Close()

	resolver := NewDependencyResolver(sqliteStore)

	now := time.Now()

	// Create tasks: A blocks B, A is completed
	taskA := &beads.Task{
		ID:        "task-A",
		Type:      beads.TaskTypeTask,
		Title:     "Task A",
		Status:    beads.StatusCompleted,
		CreatedAt: now,
		UpdatedAt: now,
	}
	taskB := &beads.Task{
		ID:        "task-B",
		Type:      beads.TaskTypeTask,
		Title:     "Task B",
		Status:    beads.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, sqliteStore.WriteTask(ctx, taskA))
	require.NoError(t, sqliteStore.WriteTask(ctx, taskB))

	// Add dependency
	dep := &beads.Dependency{
		FromTaskID: "task-A",
		ToTaskID:   "task-B",
		Type:       beads.DependencyTypeBlocks,
		CreatedAt:  now,
	}
	require.NoError(t, sqliteStore.WriteDependency(ctx, dep))

	// Task B should be ready (A is completed)
	ready, err := resolver.ComputeReadyState(ctx, "task-B")
	require.NoError(t, err)
	assert.True(t, ready, "Task should be ready when blocking task is completed")

	// No blocking tasks
	blocking, err := resolver.GetBlockingTasks(ctx, "task-B")
	require.NoError(t, err)
	assert.Empty(t, blocking, "Completed task should not be in blocking list")
}
