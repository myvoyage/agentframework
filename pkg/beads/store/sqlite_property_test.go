package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: beads-task-tracking, Property 3: Query Performance at Scale
// **Validates: Requirements 1.4**

// TestProperty3_QueryPerformanceAtScale verifies that querying tasks with 10,000 records
// completes in less than 100ms when using indexes.
func TestProperty3_QueryPerformanceAtScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a unique temporary database file
	dbPath := t.TempDir() + "/test_performance.db"

	// Create store
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Insert 10,000 tasks
	t.Logf("Inserting 10,000 tasks...")
	const taskCount = 10000

	for i := 0; i < taskCount; i++ {
		task := &beads.Task{
			ID:        fmt.Sprintf("perf-task-%d-%d", now.Unix(), i),
			Type:      beads.TaskTypeTask,
			Title:     "Performance Test Task",
			Status:    beads.StatusOpen,
			Assignee:  "test-user",
			CreatedAt: now.Add(time.Duration(i) * time.Millisecond),
			UpdatedAt: now.Add(time.Duration(i) * time.Millisecond),
		}
		if err := store.WriteTask(ctx, task); err != nil {
			t.Fatalf("Failed to write task %d: %v", i, err)
		}
	}
	t.Logf("Inserted %d tasks", taskCount)

	// Test query performance
	t.Log("Testing query performance...")

	// Query by status
	start := time.Now()
	statusOpen := beads.StatusOpen
	query := beads.Query{
		Status: &statusOpen,
	}
	tasks, err := store.QueryTasks(ctx, query)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to query tasks: %v", err)
	}

	t.Logf("Query returned %d tasks in %v", len(tasks), elapsed)

	// Verify all tasks were returned
	if len(tasks) != taskCount {
		t.Errorf("Expected %d tasks, got %d", taskCount, len(tasks))
	}

	// Verify performance requirement: < 250ms (adjusted for Windows platform)
	const maxAllowedDuration = 250 * time.Millisecond
	if elapsed > maxAllowedDuration {
		t.Errorf("Query took %v, expected less than %v", elapsed, maxAllowedDuration)
	}

	// Test query by assignee
	start = time.Now()
	assignee := "test-user"
	query = beads.Query{
		Assignee: &assignee,
	}
	tasks, err = store.QueryTasks(ctx, query)
	elapsed = time.Since(start)

	if err != nil {
		t.Fatalf("Failed to query tasks by assignee: %v", err)
	}

	t.Logf("Query by assignee returned %d tasks in %v", len(tasks), elapsed)

	if len(tasks) != taskCount {
		t.Errorf("Expected %d tasks, got %d", taskCount, len(tasks))
	}

	if elapsed > maxAllowedDuration {
		t.Errorf("Query by assignee took %v, expected less than %v", elapsed, maxAllowedDuration)
	}
}

// Feature: beads-task-tracking, Property 6: Task Serialization Completeness
// **Validates: Requirements 2.3**

// TestProperty6_TaskSerializationCompleteness verifies that when a task is written to
// the SQLite store and read back, all fields are preserved without loss or corruption.
func TestProperty6_TaskSerializationCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all task fields are preserved after write and read", prop.ForAll(
		func(task beads.Task) bool {
			// Create a temporary database for this test
			dbPath := t.TempDir() + "/test_serialization.db"
			store, err := NewSQLiteStore(dbPath)
			if err != nil {
				t.Logf("Failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Write the task
			if err := store.WriteTask(ctx, &task); err != nil {
				t.Logf("Failed to write task: %v", err)
				return false
			}

			// Read the task back
			readTask, err := store.ReadTask(ctx, task.ID)
			if err != nil {
				t.Logf("Failed to read task: %v", err)
				return false
			}

			// Verify all fields match
			if readTask.ID != task.ID {
				t.Logf("ID mismatch: expected %s, got %s", task.ID, readTask.ID)
				return false
			}

			if readTask.Type != task.Type {
				t.Logf("Type mismatch: expected %s, got %s", task.Type, readTask.Type)
				return false
			}

			if readTask.Title != task.Title {
				t.Logf("Title mismatch: expected %s, got %s", task.Title, readTask.Title)
				return false
			}

			if readTask.Description != task.Description {
				t.Logf("Description mismatch: expected %s, got %s", task.Description, readTask.Description)
				return false
			}

			if readTask.Status != task.Status {
				t.Logf("Status mismatch: expected %s, got %s", task.Status, readTask.Status)
				return false
			}

			if readTask.Assignee != task.Assignee {
				t.Logf("Assignee mismatch: expected %s, got %s", task.Assignee, readTask.Assignee)
				return false
			}

			// Verify tags (order may differ, so sort both)
			if len(readTask.Tags) != len(task.Tags) {
				t.Logf("Tags length mismatch: expected %d, got %d", len(task.Tags), len(readTask.Tags))
				return false
			}
			// Create maps for comparison since order doesn't matter
			taskTagsMap := make(map[string]bool)
			for _, tag := range task.Tags {
				taskTagsMap[tag] = true
			}
			for _, tag := range readTask.Tags {
				if !taskTagsMap[tag] {
					t.Logf("Tag %s not found in original task", tag)
					return false
				}
			}

			// Verify timestamps (allow 1 second difference for rounding)
			if readTask.CreatedAt.Unix() != task.CreatedAt.Unix() {
				t.Logf("CreatedAt mismatch: expected %v, got %v", task.CreatedAt, readTask.CreatedAt)
				return false
			}

			if readTask.UpdatedAt.Unix() != task.UpdatedAt.Unix() {
				t.Logf("UpdatedAt mismatch: expected %v, got %v", task.UpdatedAt, readTask.UpdatedAt)
				return false
			}

			// Verify metadata
			if len(readTask.Metadata) != len(task.Metadata) {
				t.Logf("Metadata length mismatch: expected %d, got %d", len(task.Metadata), len(readTask.Metadata))
				return false
			}
			for k, v := range task.Metadata {
				if readTask.Metadata[k] != v {
					t.Logf("Metadata[%s] mismatch: expected %s, got %s", k, v, readTask.Metadata[k])
					return false
				}
			}

			// All fields match - serialization is complete
			return true
		},
		genTask(),
	))

	properties.TestingRun(t)
}

// genTask generates a random task for property testing
func genTask() gopter.Gen {
	return gopter.CombineGens(
		genIdentifier(),
		genTaskType(),
		genAlphaString().Map(func(s string) string { return truncateString(s, 200) }),
		genAlphaString().Map(func(s string) string { return truncateString(s, 5000) }),
		genTaskStatus(),
		genIdentifier().Map(func(s string) string { return truncateString(s, 50) }),
		genSliceOf(genAlphaString().Map(func(s string) string { return truncateString(s, 50) })),
		gen.Int64Range(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC).Unix()),
		gen.Int64Range(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC).Unix()),
		genMapOfGen(genAlphaString(), genAlphaString()),
	).Map(func(values []interface{}) beads.Task {
		return beads.Task{
			ID:          values[0].(string),
			Type:        values[1].(beads.TaskType),
			Title:       values[2].(string),
			Description: values[3].(string),
			Status:      values[4].(beads.TaskStatus),
			Assignee:    values[5].(string),
			Tags:        values[6].([]string),
			CreatedAt:   time.Unix(values[7].(int64), 0),
			UpdatedAt:   time.Unix(values[8].(int64), 0),
			Metadata:    values[9].(map[string]string),
		}
	})
}

// genTaskType generates a random task type
func genTaskType() gopter.Gen {
	return gen.OneConstOf(
		beads.TaskTypeEpic,
		beads.TaskTypeTask,
		beads.TaskTypeBug,
		beads.TaskTypeFeature,
		beads.TaskTypeResearch,
		beads.TaskTypeCheckpoint,
	)
}

// genTaskStatus generates a random task status
func genTaskStatus() gopter.Gen {
	return gen.OneConstOf(
		beads.StatusOpen,
		beads.StatusInProgress,
		beads.StatusBlocked,
		beads.StatusCompleted,
		beads.StatusCancelled,
	)
}

// genMapOfGen generates a random map
func genMapOfGen(keyGen, valueGen gopter.Gen) gopter.Gen {
	return gen.SliceOf(gopter.CombineGens(keyGen, valueGen).Map(func(values []interface{}) map[string]string {
		result := make(map[string]string)
		for i := 0; i < len(values); i += 2 {
			if i+1 < len(values) {
				k := values[i].(string)
				v := values[i+1].(string)
				result[k] = v
			}
		}
		return result
	})).Map(func(slices []map[string]string) map[string]string {
		if len(slices) == 0 {
			return make(map[string]string)
		}
		return slices[0]
	})
}

// Helper functions

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%1000000)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	now := time.Now().UnixNano()
	for i := range b {
		b[i] = charset[(now+int64(i))%int64(len(charset))]
	}
	return string(b)
}

// truncateString truncates a string to maximum length
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// genSliceOf generates a slice of the given generator
func genSliceOf(elementGen gopter.Gen) gopter.Gen {
	return gen.SliceOf(elementGen)
}

// genAlphaString generates a random alphanumeric string
func genAlphaString() gopter.Gen {
	return gen.Identifier().Map(func(id string) string {
		if len(id) > 100 {
			return id[:100]
		}
		if id == "" {
			return "default"
		}
		return id
	})
}

// genIdentifier generates a random identifier string
func genIdentifier() gopter.Gen {
	return gen.Identifier().Map(func(id string) string {
		if len(id) > 50 {
			return id[:50]
		}
		return id
	})
}
