package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"AgentFramework/pkg/beads"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements the SQLiteStore interface using SQLite database
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store with the given database path
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Initialize the database schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the database schema with tables and indexes
func (s *SQLiteStore) initSchema() error {
	// Enable WAL mode for better concurrency
	if _, err := s.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := s.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create tasks table
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		assignee TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		metadata TEXT
	);`

	if _, err := s.db.Exec(createTasksTable); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create indexes on tasks table
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_status ON tasks(status);",
		"CREATE INDEX IF NOT EXISTS idx_assignee ON tasks(assignee);",
		"CREATE INDEX IF NOT EXISTS idx_created_at ON tasks(created_at);",
	}

	for _, idx := range indexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create task_tags table for tag relationships
	createTaskTagsTable := `
	CREATE TABLE IF NOT EXISTS task_tags (
		task_id TEXT NOT NULL,
		tag TEXT NOT NULL,
		PRIMARY KEY (task_id, tag),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(createTaskTagsTable); err != nil {
		return fmt.Errorf("failed to create task_tags table: %w", err)
	}

	// Create index on task_tags
	if _, err := s.db.Exec("CREATE INDEX IF NOT EXISTS idx_tag ON task_tags(tag);"); err != nil {
		return fmt.Errorf("failed to create tag index: %w", err)
	}

	// Create dependencies table for dependency graph
	createDependenciesTable := `
	CREATE TABLE IF NOT EXISTS dependencies (
		from_task_id TEXT NOT NULL,
		to_task_id TEXT NOT NULL,
		dep_type TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY (from_task_id, to_task_id),
		FOREIGN KEY (from_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (to_task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(createDependenciesTable); err != nil {
		return fmt.Errorf("failed to create dependencies table: %w", err)
	}

	// Create indexes on dependencies table
	depIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_from_task ON dependencies(from_task_id);",
		"CREATE INDEX IF NOT EXISTS idx_to_task ON dependencies(to_task_id);",
	}

	for _, idx := range depIndexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create dependency index: %w", err)
		}
	}

	return nil
}

// WriteTask writes a task to the database
func (s *SQLiteStore) WriteTask(ctx context.Context, task *beads.Task) error {
	// Serialize metadata to JSON
	var metadataJSON []byte
	var err error
	if task.Metadata != nil {
		metadataJSON, err = json.Marshal(task.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or replace task
	query := `
	INSERT OR REPLACE INTO tasks (id, type, title, description, status, assignee, created_at, updated_at, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		task.ID,
		task.Type,
		task.Title,
		task.Description,
		task.Status,
		task.Assignee,
		task.CreatedAt.Unix(),
		task.UpdatedAt.Unix(),
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// Delete existing tags
	_, err = tx.ExecContext(ctx, "DELETE FROM task_tags WHERE task_id = ?", task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert tags
	if len(task.Tags) > 0 {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO task_tags (task_id, tag) VALUES (?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare tag insert: %w", err)
		}
		defer stmt.Close()

		for _, tag := range task.Tags {
			_, err = stmt.ExecContext(ctx, task.ID, tag)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReadTask reads a task from the database by ID
func (s *SQLiteStore) ReadTask(ctx context.Context, taskID string) (*beads.Task, error) {
	query := `
	SELECT id, type, title, description, status, assignee, created_at, updated_at, metadata
	FROM tasks
	WHERE id = ?`

	var task beads.Task
	var metadataJSON []byte
	var createdAt, updatedAt int64

	err := s.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID,
		&task.Type,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Assignee,
		&createdAt,
		&updatedAt,
		&metadataJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	task.CreatedAt = time.Unix(createdAt, 0)
	task.UpdatedAt = time.Unix(updatedAt, 0)

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Read tags
	tagQuery := "SELECT tag FROM task_tags WHERE task_id = ?"
	rows, err := s.db.QueryContext(ctx, tagQuery, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}
	defer rows.Close()

	task.Tags = []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		task.Tags = append(task.Tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return &task, nil
}

// QueryTasks queries tasks based on the provided query parameters
func (s *SQLiteStore) QueryTasks(ctx context.Context, query beads.Query) ([]*beads.Task, error) {
	// Build the SQL query dynamically based on query parameters
	// Use a single query with JOINs instead of N+1 pattern for better performance
	sqlQuery := `SELECT DISTINCT t.id, t.type, t.title, t.description, t.status, t.assignee,
		t.created_at, t.updated_at, t.metadata
		FROM tasks t`
	args := []interface{}{}
	conditions := []string{}

	// Add tag join if needed
	if len(query.Tags) > 0 {
		sqlQuery += " LEFT JOIN task_tags tt ON t.id = tt.task_id"
	}

	// Add status condition
	if query.Status != nil {
		conditions = append(conditions, "t.status = ?")
		args = append(args, *query.Status)
	}

	// Add assignee condition
	if query.Assignee != nil {
		conditions = append(conditions, "t.assignee = ?")
		args = append(args, *query.Assignee)
	}

	// Add tag conditions
	if len(query.Tags) > 0 {
		if query.TagOp == beads.LogicalOpAND {
			// For AND: task must have all tags
			// Use GROUP BY and HAVING COUNT
			tagPlaceholders := ""
			for i, tag := range query.Tags {
				if i > 0 {
					tagPlaceholders += ", "
				}
				tagPlaceholders += "?"
				args = append(args, tag)
			}
			conditions = append(conditions, fmt.Sprintf("tt.tag IN (%s)", tagPlaceholders))
			sqlQuery += " WHERE " + joinConditions(conditions, "AND")
			sqlQuery += fmt.Sprintf(" GROUP BY t.id HAVING COUNT(DISTINCT tt.tag) = %d", len(query.Tags))
		} else {
			// For OR: task must have at least one tag
			tagPlaceholders := ""
			for i, tag := range query.Tags {
				if i > 0 {
					tagPlaceholders += ", "
				}
				tagPlaceholders += "?"
				args = append(args, tag)
			}
			conditions = append(conditions, fmt.Sprintf("tt.tag IN (%s)", tagPlaceholders))
			if len(conditions) > 0 {
				sqlQuery += " WHERE " + joinConditions(conditions, "AND")
			}
		}
	} else {
		if len(conditions) > 0 {
			sqlQuery += " WHERE " + joinConditions(conditions, "AND")
		}
	}

	// Execute query to get task details
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	// Collect tasks from main query
	taskMap := make(map[string]*beads.Task)
	for rows.Next() {
		var task beads.Task
		var metadataJSON []byte
		var createdAt, updatedAt int64

		err := rows.Scan(
			&task.ID,
			&task.Type,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Assignee,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		task.CreatedAt = time.Unix(createdAt, 0)
		task.UpdatedAt = time.Unix(updatedAt, 0)

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		task.Tags = []string{} // Will be populated below
		taskMap[task.ID] = &task
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	// If we have tasks, fetch their tags in a single query
	if len(taskMap) > 0 {
		placeholders := ""
		taskIDs := make([]string, 0, len(taskMap))
		for taskID := range taskMap {
			taskIDs = append(taskIDs, taskID)
		}
		for i := range taskIDs {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
		}

		tagQuery := fmt.Sprintf("SELECT task_id, tag FROM task_tags WHERE task_id IN (%s)", placeholders)
		args := make([]interface{}, len(taskIDs))
		for i, id := range taskIDs {
			args[i] = id
		}
		tagRows, err := s.db.QueryContext(ctx, tagQuery, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query tags: %w", err)
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var taskID, tag string
			if err := tagRows.Scan(&taskID, &tag); err != nil {
				return nil, fmt.Errorf("failed to scan tag: %w", err)
			}
			if task, exists := taskMap[taskID]; exists {
				task.Tags = append(task.Tags, tag)
			}
		}

		if err := tagRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating tags: %w", err)
		}
	}

	// Convert map to slice
	tasks := make([]*beads.Task, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// WriteDependency writes a dependency to the database with foreign key validation
func (s *SQLiteStore) WriteDependency(ctx context.Context, dep *beads.Dependency) error {
	// Begin transaction for atomic operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Validate that both tasks exist (foreign key validation)
	var fromExists, toExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", dep.FromTaskID).Scan(&fromExists)
	if err != nil {
		return fmt.Errorf("failed to check from_task existence: %w", err)
	}
	if !fromExists {
		return fmt.Errorf("from_task_id does not exist: %s", dep.FromTaskID)
	}

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", dep.ToTaskID).Scan(&toExists)
	if err != nil {
		return fmt.Errorf("failed to check to_task existence: %w", err)
	}
	if !toExists {
		return fmt.Errorf("to_task_id does not exist: %s", dep.ToTaskID)
	}

	// Insert dependency
	query := `
	INSERT OR REPLACE INTO dependencies (from_task_id, to_task_id, dep_type, created_at)
	VALUES (?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		dep.FromTaskID,
		dep.ToTaskID,
		dep.Type,
		dep.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert dependency: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReadDependencies reads dependencies for a task in the specified direction
func (s *SQLiteStore) ReadDependencies(ctx context.Context, taskID string, direction beads.Direction) ([]*beads.Dependency, error) {
	var query string
	if direction == beads.DirectionIncoming {
		// Dependencies where this task is the target (tasks that block this one)
		query = "SELECT from_task_id, to_task_id, dep_type, created_at FROM dependencies WHERE to_task_id = ?"
	} else {
		// Dependencies where this task is the source (tasks this one blocks)
		query = "SELECT from_task_id, to_task_id, dep_type, created_at FROM dependencies WHERE from_task_id = ?"
	}

	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	dependencies := []*beads.Dependency{}
	for rows.Next() {
		var dep beads.Dependency
		var createdAt int64

		err := rows.Scan(&dep.FromTaskID, &dep.ToTaskID, &dep.Type, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}

		dep.CreatedAt = time.Unix(createdAt, 0)
		dependencies = append(dependencies, &dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dependencies: %w", err)
	}

	return dependencies, nil
}

// DeleteDependency deletes a dependency between two tasks
func (s *SQLiteStore) DeleteDependency(ctx context.Context, fromID, toID string) error {
	query := "DELETE FROM dependencies WHERE from_task_id = ? AND to_task_id = ?"
	result, err := s.db.ExecContext(ctx, query, fromID, toID)
	if err != nil {
		return fmt.Errorf("failed to delete dependency: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dependency does not exist: %s -> %s", fromID, toID)
	}

	return nil
}

// RebuildFromEvents rebuilds the database from a list of events
func (s *SQLiteStore) RebuildFromEvents(ctx context.Context, events []*beads.Event) error {
	// Begin transaction for atomic rebuild
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data
	if _, err := tx.ExecContext(ctx, "DELETE FROM dependencies"); err != nil {
		return fmt.Errorf("failed to clear dependencies: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_tags"); err != nil {
		return fmt.Errorf("failed to clear task_tags: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM tasks"); err != nil {
		return fmt.Errorf("failed to clear tasks: %w", err)
	}

	// Commit the clear operation
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit clear: %w", err)
	}

	// Process events in order
	// Note: This is a simplified implementation. A full implementation would
	// need to handle event ordering, conflict resolution, and incremental updates
	for _, event := range events {
		if err := s.processEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to process event: %w", err)
		}
	}

	return nil
}

// processEvent processes a single event and applies it to the database
func (s *SQLiteStore) processEvent(ctx context.Context, event *beads.Event) error {
	switch event.Type {
	case beads.EventTaskCreated, beads.EventTaskUpdated:
		// Reconstruct task from event data
		task := &beads.Task{
			ID: event.TaskID,
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
		if tags, ok := event.Data["tags"].([]interface{}); ok {
			task.Tags = make([]string, len(tags))
			for i, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					task.Tags[i] = tagStr
				}
			}
		}
		if metadata, ok := event.Data["metadata"].(map[string]interface{}); ok {
			task.Metadata = make(map[string]string)
			for k, v := range metadata {
				if vStr, ok := v.(string); ok {
					task.Metadata[k] = vStr
				}
			}
		}

		task.CreatedAt = time.Unix(event.Timestamp, 0)
		task.UpdatedAt = time.Unix(event.Timestamp, 0)

		return s.WriteTask(ctx, task)

	case beads.EventDependencyAdded:
		dep := &beads.Dependency{
			FromTaskID: event.FromTaskID,
			ToTaskID:   event.ToTaskID,
			CreatedAt:  time.Unix(event.Timestamp, 0),
		}
		if depType, ok := event.Data["dep_type"].(string); ok {
			dep.Type = beads.DependencyType(depType)
		}
		return s.WriteDependency(ctx, dep)

	case beads.EventDependencyRemoved:
		query := "DELETE FROM dependencies WHERE from_task_id = ? AND to_task_id = ?"
		_, err := s.db.ExecContext(ctx, query, event.FromTaskID, event.ToTaskID)
		return err

	case beads.EventTaskClosed:
		// Update task status
		if status, ok := event.Data["status"].(string); ok {
			query := "UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?"
			_, err := s.db.ExecContext(ctx, query, status, event.Timestamp, event.TaskID)
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// WriteTasks writes multiple tasks to the database in a single transaction for performance
func (s *SQLiteStore) WriteTasks(ctx context.Context, tasks []*beads.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	// Begin transaction for batch operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements for reuse
	taskStmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO tasks (id, type, title, description, status, assignee, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare task statement: %w", err)
	}
	defer taskStmt.Close()

	tagDeleteStmt, err := tx.PrepareContext(ctx, "DELETE FROM task_tags WHERE task_id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare tag delete statement: %w", err)
	}
	defer tagDeleteStmt.Close()

	tagInsertStmt, err := tx.PrepareContext(ctx, "INSERT INTO task_tags (task_id, tag) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare tag insert statement: %w", err)
	}
	defer tagInsertStmt.Close()

	// Process each task
	for _, task := range tasks {
		// Serialize metadata to JSON
		var metadataJSON []byte
		if task.Metadata != nil {
			metadataJSON, err = json.Marshal(task.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for task %s: %w", task.ID, err)
			}
		}

		// Insert task
		_, err = taskStmt.ExecContext(ctx,
			task.ID,
			task.Type,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.CreatedAt.Unix(),
			task.UpdatedAt.Unix(),
			metadataJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert task %s: %w", task.ID, err)
		}

		// Delete existing tags
		_, err = tagDeleteStmt.ExecContext(ctx, task.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing tags for task %s: %w", task.ID, err)
		}

		// Insert tags
		for _, tag := range task.Tags {
			_, err = tagInsertStmt.ExecContext(ctx, task.ID, tag)
			if err != nil {
				return fmt.Errorf("failed to insert tag for task %s: %w", task.ID, err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WriteDependencies writes multiple dependencies to the database in a single transaction for performance
func (s *SQLiteStore) WriteDependencies(ctx context.Context, deps []*beads.Dependency) error {
	if len(deps) == 0 {
		return nil
	}

	// Begin transaction for batch operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for reuse
	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO dependencies (from_task_id, to_task_id, dep_type, created_at)
		VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare dependency statement: %w", err)
	}
	defer stmt.Close()

	// Validate all tasks exist first (batch validation)
	taskIDs := make(map[string]bool)
	for _, dep := range deps {
		taskIDs[dep.FromTaskID] = false
		taskIDs[dep.ToTaskID] = false
	}

	// Check existence of all referenced tasks
	for taskID := range taskIDs {
		var exists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", taskID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check task existence for %s: %w", taskID, err)
		}
		if !exists {
			return fmt.Errorf("task_id does not exist: %s", taskID)
		}
		taskIDs[taskID] = true
	}

	// Insert all dependencies
	for _, dep := range deps {
		_, err = stmt.ExecContext(ctx,
			dep.FromTaskID,
			dep.ToTaskID,
			dep.Type,
			dep.CreatedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %s -> %s: %w", dep.FromTaskID, dep.ToTaskID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Helper function to join conditions
func joinConditions(conditions []string, separator string) string {
	result := ""
	for i, cond := range conditions {
		if i > 0 {
			result += " " + separator + " "
		}
		result += cond
	}
	return result
}
