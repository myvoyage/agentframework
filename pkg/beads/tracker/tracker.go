package tracker

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/store"
)

// TaskTrackerImpl implements the beads.TaskTracker interface
// It is the central coordinator managing task lifecycle, dependencies, and storage
type TaskTrackerImpl struct {
	config         *beads.Config
	sqliteStore    beads.SQLiteStore
	jsonlStore     beads.JSONLStore
	eventProcessor beads.EventProcessor
	depResolver    beads.DependencyResolver
	syncDaemon     beads.SyncDaemon
	started        bool
	mu             sync.RWMutex
	// Path to .beads directory
	beadsPath string

	// Context store integration
	contextStore   interface{} // Optional context store (使用 interface{} 避免循环依赖)
	contextEnabled bool                // Whether context operations are enabled
}

// NewTaskTracker creates a new TaskTracker instance with the given configuration
func NewTaskTracker(config *beads.Config) beads.TaskTracker {
	return &TaskTrackerImpl{
		config:    config,
		beadsPath: config.StoragePath,
	}
}

// Start initializes the tracker, creates storage, loads historical tasks, and starts sync daemon
func (tt *TaskTrackerImpl) Start(ctx context.Context) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if tt.started {
		return fmt.Errorf("tracker already started")
	}

	// Create .beads directory if it doesn't exist
	if err := os.MkdirAll(tt.beadsPath, 0755); err != nil {
		return fmt.Errorf("failed to create .beads directory: %w", err)
	}

	// Initialize SQLite store
	sqliteStore, err := store.NewSQLiteStore(filepath.Join(tt.beadsPath, "tasks.db"))
	if err != nil {
		return fmt.Errorf("failed to initialize SQLite store: %w", err)
	}
	tt.sqliteStore = sqliteStore

	// Initialize JSONL store
	jsonlStore, err := store.NewJSONLStore(filepath.Join(tt.beadsPath, "events"))
	if err != nil {
		return fmt.Errorf("failed to initialize JSONL store: %w", err)
	}
	tt.jsonlStore = jsonlStore

	// Initialize event processor
	tt.eventProcessor = NewEventProcessor(tt.sqliteStore, tt.jsonlStore)

	// Initialize dependency resolver
	tt.depResolver = NewDependencyResolver(tt.sqliteStore)

	// Load historical tasks from JSONL
	if err := tt.loadHistoricalTasks(ctx); err != nil {
		return fmt.Errorf("failed to load historical tasks: %w", err)
	}

	// Initialize and start sync daemon (if git is enabled)
	if tt.config.GitEnabled {
		// Note: SyncDaemon implementation would go here
		// tt.syncDaemon = NewSyncDaemon(...)
		// go tt.syncDaemon.Start(ctx)
	}

	// Start context store if configured
	if tt.contextEnabled && tt.contextStore != nil {
		// 检查 contextStore 是否实现了 Start 方法
		if store, ok := tt.contextStore.(interface{ Start(ctx context.Context) error }); ok {
			if err := store.Start(ctx); err != nil {
				// Log error but don't fail - context store is optional
				tt.contextEnabled = false
			}
		}
	}

	tt.started = true
	return nil
}

// Stop gracefully shuts down the tracker
func (tt *TaskTrackerImpl) Stop(ctx context.Context) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	// Stop sync daemon if running
	if tt.syncDaemon != nil {
		if err := tt.syncDaemon.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop sync daemon: %w", err)
		}
	}

	// Close storage
	if tt.sqliteStore != nil {
		if err := tt.sqliteStore.Close(); err != nil {
			return fmt.Errorf("failed to close SQLite store: %w", err)
		}
	}

	if tt.jsonlStore != nil {
		if err := tt.jsonlStore.Close(); err != nil {
			return fmt.Errorf("failed to close JSONL store: %w", err)
		}
	}

	tt.started = false
	return nil
}

// CreateTask creates a new task with dual-layer storage
func (tt *TaskTrackerImpl) CreateTask(ctx context.Context, task *beads.Task) (string, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return "", fmt.Errorf("tracker not started")
	}

	// Validate task
	if err := tt.validateTask(task); err != nil {
		return "", fmt.Errorf("task validation failed: %w", err)
	}

	// Generate SHA-256 task ID
	task.ID = tt.generateTaskID()

	// Set timestamps
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Set default status if not provided
	if task.Status == "" {
		task.Status = beads.StatusOpen
	}

	// Write to SQLite
	if err := tt.sqliteStore.WriteTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to write task to SQLite: %w", err)
	}

	// Create and append event to JSONL
	event := &beads.Event{
		Type:      beads.EventTaskCreated,
		TaskID:    task.ID,
		Timestamp: now.Unix(),
		Data:      tt.taskToEventData(task),
	}

	if err := tt.jsonlStore.AppendEvent(ctx, event); err != nil {
		return "", fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return task.ID, nil
}

// UpdateTask updates an existing task with dual-layer storage
func (tt *TaskTrackerImpl) UpdateTask(ctx context.Context, taskID string, updates beads.TaskUpdate) error {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	// Validate task ID
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Read existing task
	existingTask, err := tt.sqliteStore.ReadTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to read existing task: %w", err)
	}

	// Apply updates
	if err := tt.applyTaskUpdates(existingTask, updates); err != nil {
		return fmt.Errorf("failed to apply updates: %w", err)
	}

	// Update timestamp
	existingTask.UpdatedAt = time.Now()

	// Write to SQLite
	if err := tt.sqliteStore.WriteTask(ctx, existingTask); err != nil {
		return fmt.Errorf("failed to write updated task to SQLite: %w", err)
	}

	// Create and append event to JSONL
	event := &beads.Event{
		Type:      beads.EventTaskUpdated,
		TaskID:    taskID,
		Timestamp: time.Now().Unix(),
		Data:      tt.taskToEventData(existingTask),
	}

	if err := tt.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// GetTask retrieves a task by ID
func (tt *TaskTrackerImpl) GetTask(ctx context.Context, taskID string) (*beads.Task, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	return tt.sqliteStore.ReadTask(ctx, taskID)
}

// CloseTask marks a task as closed with the specified status
func (tt *TaskTrackerImpl) CloseTask(ctx context.Context, taskID string, status beads.TaskStatus) error {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	// Validate status
	if status != beads.StatusCompleted && status != beads.StatusCancelled {
		return fmt.Errorf("invalid close status: %s (must be 'completed' or 'cancelled')", status)
	}

	// Read existing task
	task, err := tt.sqliteStore.ReadTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to read task: %w", err)
	}

	// Update status
	task.Status = status
	task.UpdatedAt = time.Now()

	// Write to SQLite
	if err := tt.sqliteStore.WriteTask(ctx, task); err != nil {
		return fmt.Errorf("failed to write closed task to SQLite: %w", err)
	}

	// Create and append event to JSONL
	event := &beads.Event{
		Type:      beads.EventTaskClosed,
		TaskID:    taskID,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"status": string(status),
		},
	}

	if err := tt.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// GetReady returns all tasks that are ready to execute (not blocked by dependencies)
func (tt *TaskTrackerImpl) GetReady(ctx context.Context) ([]*beads.Task, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	// Query all open and in_progress tasks
	for _, status := range []beads.TaskStatus{beads.StatusOpen, beads.StatusInProgress} {
		query := beads.Query{
			Status: &status,
		}
		tasks, err := tt.sqliteStore.QueryTasks(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query tasks with status %s: %w", status, err)
		}

		// Filter to only ready tasks
		readyTasks := make([]*beads.Task, 0)
		for _, task := range tasks {
			ready, err := tt.depResolver.ComputeReadyState(ctx, task.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to compute ready state for %s: %w", task.ID, err)
			}

			if ready {
				readyTasks = append(readyTasks, task)
			}
		}

		// Add more ready tasks from other status
		if len(readyTasks) > 0 {
			return readyTasks, nil
		}
	}

	return []*beads.Task{}, nil
}

// GetByStatus returns all tasks with the specified status
func (tt *TaskTrackerImpl) GetByStatus(ctx context.Context, status beads.TaskStatus) ([]*beads.Task, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	query := beads.Query{
		Status: &status,
	}

	return tt.sqliteStore.QueryTasks(ctx, query)
}

// GetByAssignee returns all tasks assigned to the specified user
func (tt *TaskTrackerImpl) GetByAssignee(ctx context.Context, assignee string) ([]*beads.Task, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	query := beads.Query{
		Assignee: &assignee,
	}

	return tt.sqliteStore.QueryTasks(ctx, query)
}

// GetByTags returns all tasks matching the specified tags with AND or OR logic
func (tt *TaskTrackerImpl) GetByTags(ctx context.Context, tags []string, op beads.LogicalOp) ([]*beads.Task, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if len(tags) == 0 {
		return []*beads.Task{}, nil
	}

	query := beads.Query{
		Tags:  tags,
		TagOp: op,
	}

	return tt.sqliteStore.QueryTasks(ctx, query)
}

// AddDependency adds a dependency relationship between two tasks with cycle validation
func (tt *TaskTrackerImpl) AddDependency(ctx context.Context, fromID, toID string, depType beads.DependencyType) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	// Validate task IDs
	if fromID == "" || toID == "" {
		return fmt.Errorf("both from_task_id and to_task_id are required")
	}

	if fromID == toID {
		return fmt.Errorf("cannot add dependency from task to itself")
	}

	// Validate dependency type
	validTypes := map[beads.DependencyType]bool{
		beads.DependencyTypeBlocks:         true,
		beads.DependencyTypeParentChild:    true,
		beads.DependencyTypeRelated:        true,
		beads.DependencyTypeDiscoveredFrom: true,
	}
	if !validTypes[depType] {
		return fmt.Errorf("invalid dependency type: %s", depType)
	}

	// Validate both tasks exist
	if _, err := tt.sqliteStore.ReadTask(ctx, fromID); err != nil {
		return fmt.Errorf("from_task not found: %w", err)
	}
	if _, err := tt.sqliteStore.ReadTask(ctx, toID); err != nil {
		return fmt.Errorf("to_task not found: %w", err)
	}

	// Validate no cycles (for blocks and parent-child)
	if err := tt.depResolver.ValidateNoCycles(ctx, fromID, toID, depType); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Create dependency
	dep := &beads.Dependency{
		FromTaskID: fromID,
		ToTaskID:   toID,
		Type:       depType,
		CreatedAt:  time.Now(),
	}

	// Write to SQLite
	if err := tt.sqliteStore.WriteDependency(ctx, dep); err != nil {
		return fmt.Errorf("failed to write dependency to SQLite: %w", err)
	}

	// Create and append event to JSONL
	event := &beads.Event{
		Type:       beads.EventDependencyAdded,
		FromTaskID: fromID,
		ToTaskID:   toID,
		Timestamp:  time.Now().Unix(),
		Data: map[string]interface{}{
			"dep_type": string(depType),
		},
	}

	if err := tt.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// RemoveDependency removes a dependency relationship between two tasks
func (tt *TaskTrackerImpl) RemoveDependency(ctx context.Context, fromID, toID string) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	if fromID == "" || toID == "" {
		return fmt.Errorf("both from_task_id and to_task_id are required")
	}

	// Delete from SQLite
	if err := tt.sqliteStore.DeleteDependency(ctx, fromID, toID); err != nil {
		return fmt.Errorf("failed to delete dependency from SQLite: %w", err)
	}

	// Create and append event to JSONL
	event := &beads.Event{
		Type:       beads.EventDependencyRemoved,
		FromTaskID: fromID,
		ToTaskID:   toID,
		Timestamp:  time.Now().Unix(),
		Data:       map[string]interface{}{},
	}

	if err := tt.jsonlStore.AppendEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to append event to JSONL: %w", err)
	}

	return nil
}

// GetDependencies returns all dependencies for a task (both incoming and outgoing)
func (tt *TaskTrackerImpl) GetDependencies(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	// Get both incoming and outgoing dependencies
	incoming, err := tt.sqliteStore.ReadDependencies(ctx, taskID, beads.DirectionIncoming)
	if err != nil {
		return nil, fmt.Errorf("failed to read incoming dependencies: %w", err)
	}

	outgoing, err := tt.sqliteStore.ReadDependencies(ctx, taskID, beads.DirectionOutgoing)
	if err != nil {
		return nil, fmt.Errorf("failed to read outgoing dependencies: %w", err)
	}

	// Combine both
	allDeps := append(incoming, outgoing...)
	return allDeps, nil
}

// GetDependents returns all tasks that depend on the given task
func (tt *TaskTrackerImpl) GetDependents(ctx context.Context, taskID string) ([]*beads.Dependency, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	return tt.sqliteStore.ReadDependencies(ctx, taskID, beads.DirectionIncoming)
}

// Sync triggers a manual synchronization from JSONL to SQLite
func (tt *TaskTrackerImpl) Sync(ctx context.Context) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	// Get latest sync timestamp
	latestTimestamp, err := tt.jsonlStore.GetLatestTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest timestamp: %w", err)
	}

	// Read new events since last sync
	newEvents, err := tt.jsonlStore.ReadEvents(ctx, latestTimestamp)
	if err != nil {
		return fmt.Errorf("failed to read new events: %w", err)
	}

	// Replay events to SQLite
	if err := tt.eventProcessor.ReplayEvents(ctx, newEvents); err != nil {
		return fmt.Errorf("failed to replay events: %w", err)
	}

	return nil
}

// Private helper methods

// loadHistoricalTasks loads tasks from JSONL and rebuilds SQLite state
func (tt *TaskTrackerImpl) loadHistoricalTasks(ctx context.Context) error {
	// Read all events from JSONL
	allEvents, err := tt.jsonlStore.ReadAllEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to read historical events: %w", err)
	}

	if len(allEvents) == 0 {
		return nil // No history to load
	}

	// Rebuild SQLite from events
	if err := tt.sqliteStore.RebuildFromEvents(ctx, allEvents); err != nil {
		return fmt.Errorf("failed to rebuild SQLite from events: %w", err)
	}

	return nil
}

// validateTask validates task fields before creation
func (tt *TaskTrackerImpl) validateTask(task *beads.Task) error {
	if task.Type == "" {
		return fmt.Errorf("task type is required")
	}
	if task.Title == "" {
		return fmt.Errorf("task title is required")
	}

	// Validate task type
	validTypes := map[beads.TaskType]bool{
		beads.TaskTypeEpic:       true,
		beads.TaskTypeTask:       true,
		beads.TaskTypeBug:        true,
		beads.TaskTypeFeature:    true,
		beads.TaskTypeResearch:   true,
		beads.TaskTypeCheckpoint: true,
	}
	if !validTypes[task.Type] {
		return fmt.Errorf("invalid task type: %s", task.Type)
	}

	return nil
}

// applyTaskUpdates applies updates to a task
func (tt *TaskTrackerImpl) applyTaskUpdates(task *beads.Task, updates beads.TaskUpdate) error {
	if updates.Title != nil {
		task.Title = *updates.Title
	}
	if updates.Description != nil {
		task.Description = *updates.Description
	}
	if updates.Status != nil {
		task.Status = *updates.Status
	}
	if updates.Assignee != nil {
		task.Assignee = *updates.Assignee
	}
	if updates.Tags != nil {
		task.Tags = *updates.Tags
	}
	if updates.Metadata != nil {
		task.Metadata = *updates.Metadata
	}

	return nil
}

// taskToEventData converts a Task to event data map
func (tt *TaskTrackerImpl) taskToEventData(task *beads.Task) map[string]interface{} {
	data := map[string]interface{}{
		"type":  string(task.Type),
		"title": task.Title,
		"status": string(task.Status),
	}

	if task.Description != "" {
		data["description"] = task.Description
	}
	if task.Assignee != "" {
		data["assignee"] = task.Assignee
	}
	if len(task.Tags) > 0 {
		tags := make([]interface{}, len(task.Tags))
		for i, tag := range task.Tags {
			tags[i] = tag
		}
		data["tags"] = tags
	}
	if len(task.Metadata) > 0 {
		data["metadata"] = task.Metadata
	}

	return data
}

// generateTaskID generates a unique SHA-256 hash based task ID
func (tt *TaskTrackerImpl) generateTaskID() string {
	// Get current timestamp with nanosecond precision
	timestamp := time.Now().UnixNano()

	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to less secure but still unique method
		randomBytes = []byte{
			byte(timestamp >> 56),
			byte(timestamp >> 48),
			byte(timestamp >> 40),
			byte(timestamp >> 32),
			byte(timestamp >> 24),
			byte(timestamp >> 16),
			byte(timestamp >> 8),
			byte(timestamp),
		}
	}

	// Combine timestamp and random bytes
	data := fmt.Sprintf("%d:%x", timestamp, randomBytes)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Return hex encoded hash (first 32 chars = 256 bits = 64 hex chars, use first 16 for 128-bit ID)
	return hex.EncodeToString(hash[:])[:16]
}

// ExtractTaskMemories 从任务相关上下文中提取记忆
func (tt *TaskTrackerImpl) ExtractTaskMemories(ctx context.Context, taskID string) (interface{}, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return nil, fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 ExtractMemories 方法
	type extractMemories interface {
		ExtractMemories(ctx context.Context, contextID string) (interface{}, error)
	}

	// 首先获取任务的上下文
	type getTaskContexts interface {
		GetTaskContexts(ctx context.Context, taskID string) ([]interface{}, error)
	}

	store, ok := tt.contextStore.(getTaskContexts)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetTaskContexts")
	}

	contexts, err := store.GetTaskContexts(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task contexts: %w", err)
	}

	if len(contexts) == 0 {
		return nil, nil // 没有上下文，返回 nil
	}

	// 提取第一个上下文的记忆（可以扩展为提取所有上下文的记忆）
	firstContext := contexts[0]

	// 尝试从上下文中提取记忆
	type contextWithID interface {
		GetID() string
	}

	ctxWithID, ok := firstContext.(contextWithID)
	if !ok {
		return nil, fmt.Errorf("context does not have GetID method")
	}

	extractor, ok := tt.contextStore.(extractMemories)
	if !ok {
		return nil, fmt.Errorf("context store does not support ExtractMemories")
	}

	return extractor.ExtractMemories(ctx, ctxWithID.GetID())
}

// GetTaskMemories 获取任务的记忆
func (tt *TaskTrackerImpl) GetTaskMemories(ctx context.Context, taskID string, memoryTypes []interface{}) (interface{}, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return nil, fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 GetMemories 方法
	type getMemories interface {
		GetMemories(ctx context.Context, contextID string, memoryTypes []interface{}) (interface{}, error)
	}

	// 首先获取任务的上下文
	type getTaskContexts interface {
		GetTaskContexts(ctx context.Context, taskID string) ([]interface{}, error)
	}

	store, ok := tt.contextStore.(getTaskContexts)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetTaskContexts")
	}

	contexts, err := store.GetTaskContexts(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task contexts: %w", err)
	}

	if len(contexts) == 0 {
		return nil, nil // 没有上下文，返回 nil
	}

	// 获取第一个上下文的记忆（可以扩展为获取所有上下文的记忆）
	firstContext := contexts[0]

	type contextWithID interface {
		GetID() string
	}

	ctxWithID, ok := firstContext.(contextWithID)
	if !ok {
		return nil, fmt.Errorf("context does not have GetID method")
	}

	memoriesGetter, ok := tt.contextStore.(getMemories)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetMemories")
	}

	return memoriesGetter.GetMemories(ctx, ctxWithID.GetID(), memoryTypes)
}

// GetTaskContextWithLayer 获取任务的指定层级上下文
func (tt *TaskTrackerImpl) GetTaskContextWithLayer(ctx context.Context, taskID string, layer interface{}) (interface{}, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return nil, fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 GetLayer 方法
	type getLayer interface {
		GetLayer(ctx context.Context, contextID string, layer interface{}) (interface{}, error)
	}

	// 首先获取任务的上下文
	type getTaskContexts interface {
		GetTaskContexts(ctx context.Context, taskID string) ([]interface{}, error)
	}

	store, ok := tt.contextStore.(getTaskContexts)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetTaskContexts")
	}

	contexts, err := store.GetTaskContexts(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task contexts: %w", err)
	}

	if len(contexts) == 0 {
		return nil, nil // 没有上下文，返回 nil
	}

	firstContext := contexts[0]

	type contextWithID interface {
		GetID() string
	}

	ctxWithID, ok := firstContext.(contextWithID)
	if !ok {
		return nil, fmt.Errorf("context does not have GetID method")
	}

	layerGetter, ok := tt.contextStore.(getLayer)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetLayer")
	}

	return layerGetter.GetLayer(ctx, ctxWithID.GetID(), layer)
}

// GenerateTaskContextLayers 为任务的上下文生成缺失的层级
func (tt *TaskTrackerImpl) GenerateTaskContextLayers(ctx context.Context, taskID string) error {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 GenerateLayers 方法
	type generateLayers interface {
		GenerateLayers(ctx context.Context, contextID string) error
	}

	// 首先获取任务的上下文
	type getTaskContexts interface {
		GetTaskContexts(ctx context.Context, taskID string) ([]interface{}, error)
	}

	store, ok := tt.contextStore.(getTaskContexts)
	if !ok {
		return fmt.Errorf("context store does not support GetTaskContexts")
	}

	contexts, err := store.GetTaskContexts(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task contexts: %w", err)
	}

	if len(contexts) == 0 {
		return nil // 没有上下文，不需要生成层级
	}

	firstContext := contexts[0]

	type contextWithID interface {
		GetID() string
	}

	ctxWithID, ok := firstContext.(contextWithID)
	if !ok {
		return fmt.Errorf("context does not have GetID method")
	}

	layersGenerator, ok := tt.contextStore.(generateLayers)
	if !ok {
		return fmt.Errorf("context store does not support GenerateLayers")
	}

	return layersGenerator.GenerateLayers(ctx, ctxWithID.GetID())
}

// QueryTasksWithFullContext 查询任务及其完整上下文
func (tt *TaskTrackerImpl) QueryTasksWithFullContext(ctx context.Context, query beads.Query) (interface{}, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return nil, fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 QueryTasksWithContext 方法
	type queryTasksWithContext interface {
		QueryTasksWithContext(ctx context.Context, query beads.Query, filter interface{}) (interface{}, error)
	}

	store, ok := tt.contextStore.(queryTasksWithContext)
	if !ok {
		return nil, fmt.Errorf("context store does not support QueryTasksWithContext")
	}

	return store.QueryTasksWithContext(ctx, query, nil)
}

// ===== Context Operations =====

// CreateTaskWithContext creates a task and associates it with a context
func (tt *TaskTrackerImpl) CreateTaskWithContext(
	ctx context.Context,
	task *beads.Task,
	ctxt interface{}, // 使用 interface{} 避免循环依赖
) (string, string, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return "", "", fmt.Errorf("tracker not started")
	}

	// Create the task first
	taskID, err := tt.CreateTask(ctx, task)
	if err != nil {
		return "", "", fmt.Errorf("create task failed: %w", err)
	}

	// If context store is enabled, create the context
	if tt.contextEnabled && tt.contextStore != nil {
		// 使用反射或类型断言调用 CreateContext 方法
		type createContext interface {
			CreateContext(ctx context.Context, taskID string, ctxt interface{}) (string, error)
		}

		store, ok := tt.contextStore.(createContext)
		if !ok {
			return taskID, "", nil
		}

		contextID, err := store.CreateContext(ctx, taskID, ctxt)
		if err != nil {
			// Log error but don't fail - task was created successfully
			return taskID, "", fmt.Errorf("create context failed (task created): %w", err)
		}

		// Update task metadata with context reference
		if task.Metadata == nil {
			task.Metadata = make(map[string]string)
		}
		task.Metadata["openviking_context_id"] = contextID

		// Update task with context reference
		update := beads.TaskUpdate{
			Metadata: &task.Metadata,
		}
		if err := tt.UpdateTask(ctx, taskID, update); err != nil {
			// Non-fatal error - context exists but task metadata not updated
			return taskID, contextID, fmt.Errorf("update task metadata failed (context created): %w", err)
		}

		return taskID, contextID, nil
	}

	return taskID, "", nil
}

// GetTaskContexts retrieves all contexts associated with a task
func (tt *TaskTrackerImpl) GetTaskContexts(ctx context.Context, taskID string) (interface{}, error) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return nil, fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return nil, fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 GetTaskContexts 方法
	type getContexts interface {
		GetTaskContexts(ctx context.Context, taskID string) (interface{}, error)
	}

	store, ok := tt.contextStore.(getContexts)
	if !ok {
		return nil, fmt.Errorf("context store does not support GetTaskContexts")
	}

	return store.GetTaskContexts(ctx, taskID)
}

// AssociateContext associates an existing context with a task
func (tt *TaskTrackerImpl) AssociateContext(ctx context.Context, taskID, contextID string) error {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return fmt.Errorf("context store not enabled")
	}

	// 使用类型断言调用 AssociateContext 方法
	type associateCtx interface {
		AssociateContext(ctx context.Context, taskID, contextID string) error
	}

	store, ok := tt.contextStore.(associateCtx)
	if !ok {
		return fmt.Errorf("context store does not support AssociateContext")
	}

	// Associate through context store
	if err := store.AssociateContext(ctx, taskID, contextID); err != nil {
		return fmt.Errorf("associate context failed: %w", err)
	}

	// Update task metadata
	task, err := tt.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata["openviking_context_id"] = contextID

	update := beads.TaskUpdate{
		Metadata: &task.Metadata,
	}

	return tt.UpdateTask(ctx, taskID, update)
}

// DissociateContext removes the association between a task and a context
func (tt *TaskTrackerImpl) DissociateContext(ctx context.Context, taskID, contextID string) error {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if !tt.started {
		return fmt.Errorf("tracker not started")
	}

	if !tt.contextEnabled || tt.contextStore == nil {
		return fmt.Errorf("context store not enabled")
	}

	// TODO: 实现解除关联的逻辑
	return fmt.Errorf("DissociateContext not yet implemented")
}

// SetContextStore sets the context store for this tracker
func (tt *TaskTrackerImpl) SetContextStore(ctx context.Context, store interface{}) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	// Stop existing context store if any
	if tt.contextStore != nil {
		type stopper interface {
			Stop(ctx context.Context) error
		}
		if stopper, ok := tt.contextStore.(stopper); ok {
			if err := stopper.Stop(ctx); err != nil {
				return fmt.Errorf("stop existing context store failed: %w", err)
			}
		}
	}

	// Set new context store
	tt.contextStore = store
	tt.contextEnabled = (store != nil)

	// Start new context store if provided
	if tt.contextStore != nil && tt.started {
		type starter interface {
			Start(ctx context.Context) error
		}
		if starter, ok := tt.contextStore.(starter); ok {
			if err := starter.Start(ctx); err != nil {
				tt.contextEnabled = false
				return fmt.Errorf("start context store failed: %w", err)
			}
		}
	}

	return nil
}

// IsContextEnabled returns whether context operations are enabled
func (tt *TaskTrackerImpl) IsContextEnabled() bool {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.contextEnabled
}

// GetContextStore returns the current context store
func (tt *TaskTrackerImpl) GetContextStore() interface{} {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.contextStore
}

// EnableContext 启用上下文功能
func (tt *TaskTrackerImpl) EnableContext(ctx context.Context) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if tt.contextEnabled {
		return fmt.Errorf("context already enabled")
	}

	tt.contextEnabled = true
	return nil
}

// DisableContext 禁用上下文功能
func (tt *TaskTrackerImpl) DisableContext(ctx context.Context) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if !tt.contextEnabled {
		return fmt.Errorf("context not enabled")
	}

	// Stop context store if running
	if tt.contextStore != nil {
		type stopper interface {
			Stop(ctx context.Context) error
		}
		if stopper, ok := tt.contextStore.(stopper); ok {
			_ = stopper.Stop(ctx)
		}
	}

	tt.contextEnabled = false
	tt.contextStore = nil
	return nil
}
