package tracker

import (
	"context"
	"fmt"
	"time"

	"AgentFramework/pkg/beads"
)

// EventProcessorImpl implements the beads.EventProcessor interface
// It processes events and applies them to both SQLite and JSONL storage layers
type EventProcessorImpl struct {
	sqliteStore beads.SQLiteStore
	jsonlStore  beads.JSONLStore
}

// NewEventProcessor creates a new EventProcessor instance
func NewEventProcessor(sqliteStore beads.SQLiteStore, jsonlStore beads.JSONLStore) beads.EventProcessor {
	return &EventProcessorImpl{
		sqliteStore: sqliteStore,
		jsonlStore:  jsonlStore,
	}
}

// ProcessEvent validates and processes a single event, applying it to both storage layers
func (ep *EventProcessorImpl) ProcessEvent(ctx context.Context, event *beads.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Validate event type
	if !isValidEventType(event.Type) {
		return fmt.Errorf("invalid event type: %s", event.Type)
	}

	// Validate required fields based on event type
	if err := ep.validateEventFields(event); err != nil {
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Set timestamp if not set
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Process event based on type
	switch event.Type {
	case beads.EventTaskCreated:
		return ep.handleTaskCreated(ctx, event)

	case beads.EventTaskUpdated:
		return ep.handleTaskUpdated(ctx, event)

	case beads.EventTaskClosed:
		return ep.handleTaskClosed(ctx, event)

	case beads.EventDependencyAdded:
		return ep.handleDependencyAdded(ctx, event)

	case beads.EventDependencyRemoved:
		return ep.handleDependencyRemoved(ctx, event)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// ReplayEvents processes multiple events in chronological order
func (ep *EventProcessorImpl) ReplayEvents(ctx context.Context, events []*beads.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Sort events by timestamp to ensure correct order
	sortedEvents := make([]*beads.Event, len(events))
	copy(sortedEvents, events)
	sortEventsByTimestamp(sortedEvents)

	// Process each event in order
	for i, event := range sortedEvents {
		if err := ep.ProcessEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to process event %d (type: %s, task_id: %s): %w",
				i, event.Type, event.TaskID, err)
		}
	}

	return nil
}

// handleTaskCreated processes a task_created event
func (ep *EventProcessorImpl) handleTaskCreated(ctx context.Context, event *beads.Event) error {
	// Reconstruct task from event data
	task, err := ep.extractTaskFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to extract task from event: %w", err)
	}

	// Write to SQLite
	if err := ep.sqliteStore.WriteTask(ctx, task); err != nil {
		return fmt.Errorf("failed to write task to SQLite: %w", err)
	}

	// Append to JSONL
	if err := ep.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// handleTaskUpdated processes a task_updated event
func (ep *EventProcessorImpl) handleTaskUpdated(ctx context.Context, event *beads.Event) error {
	// Reconstruct task from event data
	task, err := ep.extractTaskFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to extract task from event: %w", err)
	}

	// Verify task exists in SQLite
	existingTask, err := ep.sqliteStore.ReadTask(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to read existing task: %w", err)
	}

	// Update fields that are present in the event
	// Preserve created_at from existing task
	task.CreatedAt = existingTask.CreatedAt

	// Write updated task to SQLite
	if err := ep.sqliteStore.WriteTask(ctx, task); err != nil {
		return fmt.Errorf("failed to write updated task to SQLite: %w", err)
	}

	// Append to JSONL
	if err := ep.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// handleTaskClosed processes a task_closed event
func (ep *EventProcessorImpl) handleTaskClosed(ctx context.Context, event *beads.Event) error {
	// Extract status from event data
	status, ok := event.Data["status"].(string)
	if !ok {
		return fmt.Errorf("event data missing 'status' field")
	}

	// Verify task exists
	existingTask, err := ep.sqliteStore.ReadTask(ctx, event.TaskID)
	if err != nil {
		return fmt.Errorf("failed to read existing task: %w", err)
	}

	// Update task status
	existingTask.Status = beads.TaskStatus(status)
	existingTask.UpdatedAt = time.Unix(event.Timestamp, 0)

	// Write to SQLite
	if err := ep.sqliteStore.WriteTask(ctx, existingTask); err != nil {
		return fmt.Errorf("failed to write closed task to SQLite: %w", err)
	}

	// Append to JSONL
	if err := ep.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// handleDependencyAdded processes a dependency_added event
func (ep *EventProcessorImpl) handleDependencyAdded(ctx context.Context, event *beads.Event) error {
	// Extract dependency type from event data
	depTypeStr, ok := event.Data["dep_type"].(string)
	if !ok {
		return fmt.Errorf("event data missing 'dep_type' field")
	}

	// Verify both tasks exist
	if _, err := ep.sqliteStore.ReadTask(ctx, event.FromTaskID); err != nil {
		return fmt.Errorf("failed to read from_task: %w", err)
	}
	if _, err := ep.sqliteStore.ReadTask(ctx, event.ToTaskID); err != nil {
		return fmt.Errorf("failed to read to_task: %w", err)
	}

	// Create dependency
	dep := &beads.Dependency{
		FromTaskID: event.FromTaskID,
		ToTaskID:   event.ToTaskID,
		Type:       beads.DependencyType(depTypeStr),
		CreatedAt:  time.Unix(event.Timestamp, 0),
	}

	// Write to SQLite
	if err := ep.sqliteStore.WriteDependency(ctx, dep); err != nil {
		return fmt.Errorf("failed to write dependency to SQLite: %w", err)
	}

	// Append to JSONL
	if err := ep.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// handleDependencyRemoved processes a dependency_removed event
func (ep *EventProcessorImpl) handleDependencyRemoved(ctx context.Context, event *beads.Event) error {
	// Delete from SQLite
	if err := ep.sqliteStore.DeleteDependency(ctx, event.FromTaskID, event.ToTaskID); err != nil {
		return fmt.Errorf("failed to delete dependency from SQLite: %w", err)
	}

	// Append to JSONL
	if err := ep.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// validateEventFields validates required fields for an event based on its type
func (ep *EventProcessorImpl) validateEventFields(event *beads.Event) error {
	switch event.Type {
	case beads.EventTaskCreated, beads.EventTaskUpdated, beads.EventTaskClosed:
		if event.TaskID == "" {
			return fmt.Errorf("task_id is required for event type: %s", event.Type)
		}
		if event.Data == nil {
			return fmt.Errorf("event data is required for event type: %s", event.Type)
		}

	case beads.EventDependencyAdded, beads.EventDependencyRemoved:
		if event.FromTaskID == "" {
			return fmt.Errorf("from_task_id is required for event type: %s", event.Type)
		}
		if event.ToTaskID == "" {
			return fmt.Errorf("to_task_id is required for event type: %s", event.Type)
		}
		if event.Data == nil {
			return fmt.Errorf("event data is required for event type: %s", event.Type)
		}
	}

	return nil
}

// extractTaskFromEvent reconstructs a Task from event data
func (ep *EventProcessorImpl) extractTaskFromEvent(event *beads.Event) (*beads.Task, error) {
	task := &beads.Task{
		ID:        event.TaskID,
		CreatedAt: time.Unix(event.Timestamp, 0),
		UpdatedAt: time.Unix(event.Timestamp, 0),
		Metadata:  make(map[string]string),
	}

	// Extract fields from event data
	if typeVal, ok := event.Data["type"].(string); ok {
		task.Type = beads.TaskType(typeVal)
	}
	if title, ok := event.Data["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := event.Data["description"].(string); ok {
		task.Description = desc
	}
	if status, ok := event.Data["status"].(string); ok {
		task.Status = beads.TaskStatus(status)
	}
	if assignee, ok := event.Data["assignee"].(string); ok {
		task.Assignee = assignee
	}

	// Extract tags
	if tags, ok := event.Data["tags"].([]interface{}); ok {
		task.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	// Extract metadata
	if metadata, ok := event.Data["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			if vStr, ok := v.(string); ok {
				task.Metadata[k] = vStr
			}
		}
	}

	// Validate required fields
	if task.Type == "" {
		return nil, fmt.Errorf("task type is required")
	}
	if task.Title == "" {
		return nil, fmt.Errorf("task title is required")
	}
	if task.Status == "" {
		task.Status = beads.StatusOpen
	}

	return task, nil
}

// isValidEventType checks if an event type is valid
func isValidEventType(eventType beads.EventType) bool {
	switch eventType {
	case beads.EventTaskCreated,
		beads.EventTaskUpdated,
		beads.EventTaskClosed,
		beads.EventDependencyAdded,
		beads.EventDependencyRemoved:
		return true
	default:
		return false
	}
}

// sortEventsByTimestamp sorts events by timestamp in ascending order
func sortEventsByTimestamp(events []*beads.Event) {
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp > events[j].Timestamp {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}
