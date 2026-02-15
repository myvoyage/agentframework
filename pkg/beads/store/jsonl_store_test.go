package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONLStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "jsonl")

	store, err := NewJSONLStore(storePath)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// Verify directory was created
	info, err := os.Stat(storePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestAppendEvent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	event := &beads.Event{
		Type:      beads.EventTaskCreated,
		TaskID:    "task-123",
		Timestamp: now.Unix(),
		Data: map[string]interface{}{
			"title": "Test Task",
			"type":  "task",
		},
	}

	err = store.AppendEvent(ctx, event)
	require.NoError(t, err)

	// Force flush to ensure event is written
	err = store.ForceFlush()
	require.NoError(t, err)

	// Verify event was written
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, event.Type, events[0].Type)
	assert.Equal(t, event.TaskID, events[0].TaskID)
	assert.Equal(t, event.Timestamp, events[0].Timestamp)
}

func TestAppendEvent_NilEvent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	err = store.AppendEvent(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestAppendEvent_BufferedWrites(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Add multiple events
	for i := 0; i < 10; i++ {
		event := &beads.Event{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-" + string(rune(i)),
			Timestamp: now.Add(time.Duration(i) * time.Second).Unix(),
			Data:      map[string]interface{}{"index": i},
		}
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	// Force flush
	err = store.ForceFlush()
	require.NoError(t, err)

	// Verify all events were written
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 10)
}

func TestReadEvents_TimeFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add events at different times
	for i := 0; i < 5; i++ {
		event := &beads.Event{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-" + string(rune(i)),
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour).Unix(),
			Data:      map[string]interface{}{"index": i},
		}
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Read events since 2 hours after base time
	since := baseTime.Add(2 * time.Hour)
	events, err := store.ReadEvents(ctx, since)
	require.NoError(t, err)

	// Should get events at 2, 3, 4 hours (3 events)
	assert.Len(t, events, 3)
	assert.Equal(t, "task-"+string(rune(2)), events[0].TaskID)
}

func TestReadEvents_EmptyStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestReadEvents_MonthBoundary(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Add events in different months
	jan := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	feb := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	mar := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	events := []*beads.Event{
		{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-jan",
			Timestamp: jan.Unix(),
			Data:      map[string]interface{}{"month": "jan"},
		},
		{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-feb",
			Timestamp: feb.Unix(),
			Data:      map[string]interface{}{"month": "feb"},
		},
		{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-mar",
			Timestamp: mar.Unix(),
			Data:      map[string]interface{}{"month": "mar"},
		},
	}

	for _, event := range events {
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Verify separate files were created
	files, err := store.getJSONLFiles()
	require.NoError(t, err)
	assert.Len(t, files, 3)

	// Verify all events can be read back
	allEvents, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, allEvents, 3)

	// Verify events are sorted by timestamp
	assert.Equal(t, "task-jan", allEvents[0].TaskID)
	assert.Equal(t, "task-feb", allEvents[1].TaskID)
	assert.Equal(t, "task-mar", allEvents[2].TaskID)
}

func TestGetLatestTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Empty store should return zero time
	timestamp, err := store.GetLatestTimestamp(ctx)
	require.NoError(t, err)
	assert.True(t, timestamp.IsZero())

	// Add events
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latestTime := baseTime.Add(5 * time.Hour)

	for i := 0; i < 6; i++ {
		event := &beads.Event{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-" + string(rune(i)),
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour).Unix(),
			Data:      map[string]interface{}{"index": i},
		}
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Get latest timestamp
	timestamp, err = store.GetLatestTimestamp(ctx)
	require.NoError(t, err)
	assert.Equal(t, latestTime.Unix(), timestamp.Unix())
}

func TestGetLatestTimestamp_MultipleMonths(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Add events across multiple months
	jan := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	feb := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	mar := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	events := []*beads.Event{
		{Type: beads.EventTaskCreated, TaskID: "task-jan", Timestamp: jan.Unix()},
		{Type: beads.EventTaskCreated, TaskID: "task-feb", Timestamp: feb.Unix()},
		{Type: beads.EventTaskCreated, TaskID: "task-mar", Timestamp: mar.Unix()},
	}

	for _, event := range events {
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Latest should be March
	timestamp, err := store.GetLatestTimestamp(ctx)
	require.NoError(t, err)
	assert.Equal(t, mar.Unix(), timestamp.Unix())
}

func TestFilePartitioning(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Add events in January and February 2024
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	feb1 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	events := []*beads.Event{
		{Type: beads.EventTaskCreated, TaskID: "task-jan-1", Timestamp: jan1.Unix(), Data: map[string]interface{}{}},
		{Type: beads.EventTaskCreated, TaskID: "task-jan-31", Timestamp: jan31.Unix(), Data: map[string]interface{}{}},
		{Type: beads.EventTaskCreated, TaskID: "task-feb-1", Timestamp: feb1.Unix(), Data: map[string]interface{}{}},
	}

	for _, event := range events {
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Verify two files were created
	janFile := filepath.Join(tmpDir, "2024-01.jsonl")
	febFile := filepath.Join(tmpDir, "2024-02.jsonl")

	_, err = os.Stat(janFile)
	assert.NoError(t, err, "January file should exist")

	_, err = os.Stat(febFile)
	assert.NoError(t, err, "February file should exist")

	// Read all events and verify partitioning
	allEvents, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, allEvents, 3, "Should have 3 total events")

	// Verify January file has 2 events
	janEvents, err := store.readEventsFromFile(janFile, time.Time{})
	require.NoError(t, err)
	assert.Len(t, janEvents, 2, "January file should have 2 events")

	// Verify February file has 1 event
	febEvents, err := store.readEventsFromFile(febFile, time.Time{})
	require.NoError(t, err)
	assert.Len(t, febEvents, 1, "February file should have 1 event")
}

func TestSingleLineFormat(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Add event with data that might contain newlines
	event := &beads.Event{
		Type:      beads.EventTaskCreated,
		TaskID:    "task-123",
		Timestamp: now.Unix(),
		Data: map[string]interface{}{
			"description": "This is a\nmulti-line\ndescription",
		},
	}

	err = store.AppendEvent(ctx, event)
	require.NoError(t, err)

	err = store.ForceFlush()
	require.NoError(t, err)

	// Read the file and verify it's a single line
	monthKey := store.getMonthKey(now)
	filePath := store.getFilePath(monthKey)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}

	// Should have exactly one newline (at the end)
	assert.Equal(t, 1, lines, "Event should be written as a single line")
}

func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Write events concurrently
	numGoroutines := 10
	eventsPerGoroutine := 10
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < eventsPerGoroutine; i++ {
				event := &beads.Event{
					Type:      beads.EventTaskCreated,
					TaskID:    "task-" + string(rune(goroutineID*100+i)),
					Timestamp: now.Add(time.Duration(i) * time.Millisecond).Unix(),
					Data:      map[string]interface{}{"g": goroutineID, "i": i},
				}
				store.AppendEvent(ctx, event)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Verify all events were written
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Equal(t, numGoroutines*eventsPerGoroutine, len(events))
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

	// Add events
	for i := 0; i < 5; i++ {
		event := &beads.Event{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-" + string(rune(i)),
			Timestamp: now.Add(time.Duration(i) * time.Second).Unix(),
			Data:      map[string]interface{}{"index": i},
		}
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	// Close should flush remaining events
	err = store.Close()
	require.NoError(t, err)

	// Create new store to read events
	store2, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store2.Close()

	events, err := store2.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 5)
}

func TestReadEvents_CorruptedLine(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Add a valid event
	event1 := &beads.Event{
		Type:      beads.EventTaskCreated,
		TaskID:    "task-1",
		Timestamp: now.Unix(),
		Data:      map[string]interface{}{"valid": true},
	}
	err = store.AppendEvent(ctx, event1)
	require.NoError(t, err)

	err = store.ForceFlush()
	require.NoError(t, err)

	// Manually append a corrupted line
	monthKey := store.getMonthKey(now)
	filePath := store.getFilePath(monthKey)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = file.WriteString("this is not valid json\n")
	require.NoError(t, err)
	file.Close()

	// Add another valid event
	event2 := &beads.Event{
		Type:      beads.EventTaskCreated,
		TaskID:    "task-2",
		Timestamp: now.Add(time.Second).Unix(),
		Data:      map[string]interface{}{"valid": true},
	}
	err = store.AppendEvent(ctx, event2)
	require.NoError(t, err)

	err = store.ForceFlush()
	require.NoError(t, err)

	// Should read valid events and skip corrupted line
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 2, "Should read 2 valid events, skipping corrupted line")
	assert.Equal(t, "task-1", events[0].TaskID)
	assert.Equal(t, "task-2", events[1].TaskID)
}

func TestEventOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewJSONLStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add events in random order
	timestamps := []int{5, 2, 8, 1, 9, 3, 7, 4, 6, 0}
	for _, ts := range timestamps {
		event := &beads.Event{
			Type:      beads.EventTaskCreated,
			TaskID:    "task-" + string(rune(ts)),
			Timestamp: baseTime.Add(time.Duration(ts) * time.Hour).Unix(),
			Data:      map[string]interface{}{"ts": ts},
		}
		err = store.AppendEvent(ctx, event)
		require.NoError(t, err)
	}

	err = store.ForceFlush()
	require.NoError(t, err)

	// Read all events - should be sorted by timestamp
	events, err := store.ReadAllEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 10)

	// Verify events are in chronological order
	for i := 0; i < len(events)-1; i++ {
		assert.True(t, events[i].Timestamp <= events[i+1].Timestamp,
			"Events should be sorted by timestamp")
	}
}
