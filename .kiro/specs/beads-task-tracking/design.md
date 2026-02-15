# Design Document: Beads Task Tracking Integration

## Overview

The Beads Task Tracking system integrates a Git-native distributed task tracking engine into AgentFramework. The design follows a dual-layer architecture pattern where SQLite provides fast local queries while JSONL files enable Git-based version control and distributed collaboration.

The system solves the "50 First Dates" problem for AI agents by providing persistent memory across sessions. Tasks are stored as immutable events in append-only JSONL logs, enabling conflict-free merging and complete audit trails.

**Key Design Principles:**
- **Event Sourcing**: All task changes are immutable events
- **Dual-Layer Storage**: SQLite for queries, JSONL for persistence
- **Hash-Based IDs**: SHA-256 prevents distributed conflicts
- **Query-Driven**: Agents query for ready tasks, not parse markdown
- **Git-Native**: JSONL format enables line-based merging

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Framework                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                  │
│  │ChatAgent │  │ReActAgent│  │WorkerAgent│                  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘                  │
│       │             │              │                         │
│       └─────────────┴──────────────┘                         │
│                     │                                        │
│              ┌──────▼──────┐                                │
│              │ MCP Interface│                                │
│              └──────┬──────┘                                │
│                     │                                        │
│       ┌─────────────▼─────────────┐                         │
│       │    Task Tracker Core      │                         │
│       │  ┌──────────────────────┐ │                         │
│       │  │  Query Engine        │ │                         │
│       │  │  Dependency Resolver │ │                         │
│       │  │  Event Processor     │ │                         │
│       │  └──────────────────────┘ │                         │
│       └─────────┬───────┬─────────┘                         │
│                 │       │                                    │
│        ┌────────▼──┐ ┌──▼────────┐                         │
│        │ SQLite    │ │ Sync      │                         │
│        │ Store     │ │ Daemon    │                         │
│        └────────┬──┘ └──┬────────┘                         │
│                 │       │                                    │
│                 └───┬───┘                                    │
│                     │                                        │
│              ┌──────▼──────┐                                │
│              │ JSONL Store │                                │
│              │  (.beads/)  │                                │
│              └──────┬──────┘                                │
│                     │                                        │
│              ┌──────▼──────┐                                │
│              │ Git Repository│                              │
│              └─────────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

```
Agent Request → MCP Interface → Task Tracker Core
                                      ↓
                              Query Engine (SQLite)
                                      ↓
                              Event Processor
                                      ↓
                         ┌────────────┴────────────┐
                         ↓                         ↓
                   SQLite Store              JSONL Store
                         ↓                         ↓
                   Sync Daemon ←──────────────────┘
                         ↓
                   Git Repository
```

## Components and Interfaces

### 1. Task Tracker Core

**Responsibility**: Central coordinator managing task lifecycle, dependencies, and storage.

**Interface**:
```go
type TaskTracker interface {
    // Task Operations
    CreateTask(ctx context.Context, task *Task) (string, error)
    UpdateTask(ctx context.Context, taskID string, updates TaskUpdate) error
    GetTask(ctx context.Context, taskID string) (*Task, error)
    CloseTask(ctx context.Context, taskID string, status TaskStatus) error
    
    // Query Operations
    GetReady(ctx context.Context) ([]*Task, error)
    GetByStatus(ctx context.Context, status TaskStatus) ([]*Task, error)
    GetByAssignee(ctx context.Context, assignee string) ([]*Task, error)
    GetByTags(ctx context.Context, tags []string, op LogicalOp) ([]*Task, error)
    
    // Dependency Operations
    AddDependency(ctx context.Context, fromID, toID string, depType DependencyType) error
    RemoveDependency(ctx context.Context, fromID, toID string) error
    GetDependencies(ctx context.Context, taskID string) ([]*Dependency, error)
    GetDependents(ctx context.Context, taskID string) ([]*Dependency, error)
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Sync(ctx context.Context) error
}
```

**Implementation Details**:
- Validates all inputs before processing
- Delegates storage to SQLite and JSONL stores
- Coordinates synchronization between layers
- Publishes events to message bus for agent coordination

### 2. SQLite Store

**Responsibility**: Fast local query engine with indexed access.

**Schema**:
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    assignee TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    metadata TEXT  -- JSON blob for extensibility
);

CREATE INDEX idx_status ON tasks(status);
CREATE INDEX idx_assignee ON tasks(assignee);
CREATE INDEX idx_created_at ON tasks(created_at);

CREATE TABLE task_tags (
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tag ON task_tags(tag);

CREATE TABLE dependencies (
    from_task_id TEXT NOT NULL,
    to_task_id TEXT NOT NULL,
    dep_type TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (from_task_id, to_task_id),
    FOREIGN KEY (from_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (to_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_from_task ON dependencies(from_task_id);
CREATE INDEX idx_to_task ON dependencies(to_task_id);
```

**Interface**:
```go
type SQLiteStore interface {
    WriteTask(ctx context.Context, task *Task) error
    ReadTask(ctx context.Context, taskID string) (*Task, error)
    QueryTasks(ctx context.Context, query Query) ([]*Task, error)
    WriteDependency(ctx context.Context, dep *Dependency) error
    ReadDependencies(ctx context.Context, taskID string, direction Direction) ([]*Dependency, error)
    RebuildFromEvents(ctx context.Context, events []*Event) error
}
```

### 3. JSONL Store

**Responsibility**: Git-tracked append-only event log.

**File Structure**:
```
.beads/
├── tasks/
│   ├── 2024-01.jsonl
│   ├── 2024-02.jsonl
│   └── 2024-03.jsonl
└── config.json
```

**Event Format**:
```json
{"event_type":"task_created","task_id":"a3f2c1...","timestamp":1704067200,"data":{"type":"task","title":"Implement parser","status":"open"}}
{"event_type":"task_updated","task_id":"a3f2c1...","timestamp":1704070800,"data":{"status":"in_progress"}}
{"event_type":"dependency_added","from_task_id":"a3f2c1...","to_task_id":"b4e3d2...","timestamp":1704074400,"data":{"dep_type":"blocks"}}
```

**Interface**:
```go
type JSONLStore interface {
    AppendEvent(ctx context.Context, event *Event) error
    ReadEvents(ctx context.Context, since time.Time) ([]*Event, error)
    ReadAllEvents(ctx context.Context) ([]*Event, error)
    GetLatestTimestamp(ctx context.Context) (time.Time, error)
}
```

**Implementation Details**:
- Events partitioned by month for manageable file sizes
- Each line is a complete JSON object (JSONL format)
- Append-only semantics enable conflict-free merging
- Events include monotonic timestamps for ordering

### 4. Sync Daemon

**Responsibility**: Bidirectional synchronization between SQLite and JSONL.

**Interface**:
```go
type SyncDaemon interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    TriggerSync(ctx context.Context) error
    GetStatus() SyncStatus
}
```

**Synchronization Strategy**:
1. **SQLite → JSONL**: Immediate append on every write
2. **JSONL → SQLite**: Periodic replay (configurable interval)
3. **Git Pull Detection**: Watch .beads/ directory for external changes
4. **Conflict Resolution**: Last-write-wins based on timestamp

**Implementation**:
```go
// Sync loop pseudocode
func (d *SyncDaemon) syncLoop(ctx context.Context) {
    ticker := time.NewTicker(d.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            d.syncJSONLToSQLite(ctx)
        case <-d.triggerChan:
            d.syncJSONLToSQLite(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (d *SyncDaemon) syncJSONLToSQLite(ctx context.Context) error {
    lastSync := d.getLastSyncTime()
    events, err := d.jsonlStore.ReadEvents(ctx, lastSync)
    if err != nil {
        return err
    }
    
    return d.sqliteStore.RebuildFromEvents(ctx, events)
}
```

### 5. Dependency Resolver

**Responsibility**: Compute task ready state and validate dependency graphs.

**Interface**:
```go
type DependencyResolver interface {
    ComputeReadyState(ctx context.Context, taskID string) (bool, error)
    ValidateNoCycles(ctx context.Context, fromID, toID string, depType DependencyType) error
    GetBlockingTasks(ctx context.Context, taskID string) ([]*Task, error)
    GetDependencyChain(ctx context.Context, taskID string) ([]*Task, error)
}
```

**Algorithm for Ready State**:
```go
func (r *DependencyResolver) ComputeReadyState(ctx context.Context, taskID string) (bool, error) {
    // Get all "blocks" dependencies
    deps, err := r.store.ReadDependencies(ctx, taskID, Incoming)
    if err != nil {
        return false, err
    }
    
    blockingDeps := filter(deps, func(d *Dependency) bool {
        return d.Type == DependencyTypeBlocks
    })
    
    // Check if all blocking tasks are completed
    for _, dep := range blockingDeps {
        task, err := r.store.ReadTask(ctx, dep.FromTaskID)
        if err != nil {
            return false, err
        }
        
        if task.Status != StatusCompleted {
            return false, nil
        }
    }
    
    return true, nil
}
```

**Cycle Detection**:
```go
func (r *DependencyResolver) ValidateNoCycles(ctx context.Context, fromID, toID string, depType DependencyType) error {
    // Only validate for "blocks" and "parent-child" dependencies
    if depType != DependencyTypeBlocks && depType != DependencyTypeParentChild {
        return nil
    }
    
    // DFS to detect cycles
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    
    return r.dfs(ctx, toID, fromID, visited, recStack)
}

func (r *DependencyResolver) dfs(ctx context.Context, current, target string, visited, recStack map[string]bool) error {
    visited[current] = true
    recStack[current] = true
    
    if current == target {
        return ErrCycleDetected
    }
    
    deps, err := r.store.ReadDependencies(ctx, current, Outgoing)
    if err != nil {
        return err
    }
    
    for _, dep := range deps {
        if !visited[dep.ToTaskID] {
            if err := r.dfs(ctx, dep.ToTaskID, target, visited, recStack); err != nil {
                return err
            }
        } else if recStack[dep.ToTaskID] {
            return ErrCycleDetected
        }
    }
    
    recStack[current] = false
    return nil
}
```

### 6. MCP Interface

**Responsibility**: Expose task operations as MCP tools for agent consumption.

**Tools**:
```go
type MCPInterface struct {
    tracker TaskTracker
}

// MCP Tool Definitions
var MCPTools = []MCPTool{
    {
        Name: "create_task",
        Description: "Create a new task with specified type, title, and metadata",
        InputSchema: CreateTaskSchema,
    },
    {
        Name: "update_task",
        Description: "Update an existing task's fields",
        InputSchema: UpdateTaskSchema,
    },
    {
        Name: "close_task",
        Description: "Close a task with specified status (completed/cancelled)",
        InputSchema: CloseTaskSchema,
    },
    {
        Name: "get_ready_tasks",
        Description: "Get all tasks that are ready to execute (no blocking dependencies)",
        InputSchema: GetReadyTasksSchema,
    },
    {
        Name: "show_task",
        Description: "Show detailed information about a specific task",
        InputSchema: ShowTaskSchema,
    },
    {
        Name: "add_dependency",
        Description: "Add a dependency relationship between two tasks",
        InputSchema: AddDependencySchema,
    },
    {
        Name: "list_tasks",
        Description: "List tasks filtered by status, assignee, or tags",
        InputSchema: ListTasksSchema,
    },
}
```

**Example Tool Implementation**:
```go
func (m *MCPInterface) CreateTask(ctx context.Context, input CreateTaskInput) (*MCPResponse, error) {
    task := &Task{
        Type:        input.Type,
        Title:       input.Title,
        Description: input.Description,
        Status:      StatusOpen,
        Assignee:    input.Assignee,
        Tags:        input.Tags,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    taskID, err := m.tracker.CreateTask(ctx, task)
    if err != nil {
        return &MCPResponse{
            Success: false,
            Error: &MCPError{
                Code:    "CREATE_FAILED",
                Message: err.Error(),
            },
        }, nil
    }
    
    return &MCPResponse{
        Success: true,
        Data: map[string]interface{}{
            "task_id": taskID,
            "task":    task,
        },
    }, nil
}
```

### 7. Event Processor

**Responsibility**: Process events and apply them to storage layers.

**Interface**:
```go
type EventProcessor interface {
    ProcessEvent(ctx context.Context, event *Event) error
    ReplayEvents(ctx context.Context, events []*Event) error
}
```

**Event Types**:
```go
type EventType string

const (
    EventTaskCreated      EventType = "task_created"
    EventTaskUpdated      EventType = "task_updated"
    EventTaskClosed       EventType = "task_closed"
    EventDependencyAdded  EventType = "dependency_added"
    EventDependencyRemoved EventType = "dependency_removed"
)
```

**Event Processing Logic**:
```go
func (p *EventProcessor) ProcessEvent(ctx context.Context, event *Event) error {
    switch event.Type {
    case EventTaskCreated:
        return p.handleTaskCreated(ctx, event)
    case EventTaskUpdated:
        return p.handleTaskUpdated(ctx, event)
    case EventTaskClosed:
        return p.handleTaskClosed(ctx, event)
    case EventDependencyAdded:
        return p.handleDependencyAdded(ctx, event)
    case EventDependencyRemoved:
        return p.handleDependencyRemoved(ctx, event)
    default:
        return fmt.Errorf("unknown event type: %s", event.Type)
    }
}
```

## Data Models

### Task Model

```go
type Task struct {
    ID          string            `json:"id"`
    Type        TaskType          `json:"type"`
    Title       string            `json:"title"`
    Description string            `json:"description"`
    Status      TaskStatus        `json:"status"`
    Assignee    string            `json:"assignee,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type TaskType string

const (
    TaskTypeEpic       TaskType = "epic"
    TaskTypeTask       TaskType = "task"
    TaskTypeBug        TaskType = "bug"
    TaskTypeFeature    TaskType = "feature"
    TaskTypeResearch   TaskType = "research"
    TaskTypeCheckpoint TaskType = "checkpoint"
)

type TaskStatus string

const (
    StatusOpen       TaskStatus = "open"
    StatusInProgress TaskStatus = "in_progress"
    StatusBlocked    TaskStatus = "blocked"
    StatusCompleted  TaskStatus = "completed"
    StatusCancelled  TaskStatus = "cancelled"
)
```

### Dependency Model

```go
type Dependency struct {
    FromTaskID string          `json:"from_task_id"`
    ToTaskID   string          `json:"to_task_id"`
    Type       DependencyType  `json:"type"`
    CreatedAt  time.Time       `json:"created_at"`
}

type DependencyType string

const (
    DependencyTypeBlocks        DependencyType = "blocks"
    DependencyTypeParentChild   DependencyType = "parent-child"
    DependencyTypeRelated       DependencyType = "related"
    DependencyTypeDiscoveredFrom DependencyType = "discovered-from"
)
```

### Event Model

```go
type Event struct {
    Type      EventType              `json:"event_type"`
    TaskID    string                 `json:"task_id,omitempty"`
    FromTaskID string                `json:"from_task_id,omitempty"`
    ToTaskID  string                 `json:"to_task_id,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}
```

### Configuration Model

```go
type Config struct {
    StoragePath  string        `json:"storage_path"`
    GitEnabled   bool          `json:"git_enabled"`
    SyncInterval time.Duration `json:"sync_interval"`
    MaxTasks     int           `json:"max_tasks"`
    DBPath       string        `json:"db_path"`
    JSONLPath    string        `json:"jsonl_path"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Dual-Layer Storage Consistency

*For any* task created or updated, the task data in SQLite_Store and JSONL_Store SHALL be identical when queried immediately after the operation.

**Validates: Requirements 1.1**

### Property 2: JSONL Synchronization to SQLite

*For any* external modification to JSONL_Store (simulating git pull), after triggering synchronization, SQLite_Store SHALL reflect all changes from the JSONL events.

**Validates: Requirements 1.3**

### Property 3: Query Performance at Scale

*For any* dataset containing up to 10,000 tasks, query operations (GetByID, GetByStatus, GetByAssignee) SHALL complete within 100ms.

**Validates: Requirements 1.4**

### Property 4: Append-Only JSONL Semantics

*For any* sequence of task operations, the JSONL_Store SHALL only append new lines and never modify or delete existing lines, ensuring conflict-free merging.

**Validates: Requirements 1.5**

### Property 5: Task ID Uniqueness and Format

*For any* task created, the Task_ID SHALL be a 64-character hexadecimal string (SHA-256 hash) and SHALL be unique across all tasks even under concurrent creation.

**Validates: Requirements 2.1, 7.2**

### Property 6: Task Serialization Completeness

*For any* task stored in either SQLite_Store or JSONL_Store, all required fields (id, type, title, description, status, created_at, updated_at, assignee, tags) SHALL be present and retrievable.

**Validates: Requirements 2.3**

### Property 7: Event Sourcing History Preservation

*For any* task that undergoes multiple updates, the JSONL_Store SHALL contain all update events in chronological order, enabling complete history reconstruction.

**Validates: Requirements 2.5**

### Property 8: Dependency Cycle Prevention

*For any* attempt to add a "blocks" or "parent-child" dependency, if the addition would create a cycle in the Dependency_Graph, the operation SHALL be rejected with an error.

**Validates: Requirements 3.2**

### Property 9: Ready State Computation Correctness

*For any* task with "blocks" dependencies, the task SHALL be marked as ready if and only if all blocking tasks have status "completed".

**Validates: Requirements 3.3**

### Property 10: Hierarchical Task Structure Integrity

*For any* set of tasks connected by "parent-child" dependencies, the structure SHALL form a valid tree (no cycles, single parent per child).

**Validates: Requirements 3.4**

### Property 11: Dependency Query Completeness

*For any* task with dependencies, querying GetDependencies() SHALL return all dependencies with their types and the current status of each dependent task.

**Validates: Requirements 3.5**

### Property 12: GetReady Query Correctness

*For any* state of the task system, GetReady() SHALL return exactly the set of tasks that have status "open" or "in_progress" and have no incomplete blocking dependencies.

**Validates: Requirements 4.1**

### Property 13: GetReady JSON Format Validity

*For any* result from GetReady(), the output SHALL be valid JSON containing an array of task objects with all required metadata fields.

**Validates: Requirements 4.2**

### Property 14: Query Method Correctness

*For any* task in the system, GetByID(task.ID) SHALL return that task, GetByStatus(task.Status) SHALL include that task, GetByAssignee(task.Assignee) SHALL include that task, and GetByTags(task.Tags) SHALL include that task when using OR logic.

**Validates: Requirements 4.3**

### Property 15: Tag Query Logical Operations

*For any* set of tasks with various tags, querying with AND operation SHALL return only tasks having all specified tags, and querying with OR operation SHALL return tasks having any of the specified tags.

**Validates: Requirements 4.4**

### Property 16: Graph Traversal Correctness

*For any* task in a dependency graph, GetDependencies() SHALL return all tasks that this task depends on, and GetDependents() SHALL return all tasks that depend on this task.

**Validates: Requirements 4.5**

### Property 17: Sync Latency Guarantee

*For any* modification to JSONL_Store, the Sync_Daemon SHALL update SQLite_Store to reflect the changes within 1 second of the modification.

**Validates: Requirements 5.2**

### Property 18: Sync Retry with Exponential Backoff

*For any* synchronization failure, the Sync_Daemon SHALL retry the operation with exponentially increasing delays (e.g., 1s, 2s, 4s, 8s) until success or maximum retries reached.

**Validates: Requirements 5.4**

### Property 19: Conflict Resolution Last-Write-Wins

*For any* conflicting updates to the same task (same task_id, different data), the Sync_Daemon SHALL resolve the conflict by keeping the update with the latest timestamp.

**Validates: Requirements 5.5**

### Property 20: MCP Input Validation

*For any* call to create_task MCP tool without required fields (type or title), the tool SHALL return an error response with code "VALIDATION_ERROR" and SHALL NOT create a task.

**Validates: Requirements 6.2**

### Property 21: MCP Dependency Validation

*For any* call to add_dependency MCP tool with non-existent task IDs or invalid dependency type, the tool SHALL return an error response and SHALL NOT create the dependency.

**Validates: Requirements 6.4**

### Property 22: MCP Error Response Structure

*For any* MCP tool operation that fails, the response SHALL contain a "success: false" field, an "error" object with "code" and "message" fields.

**Validates: Requirements 6.5**

### Property 23: JSONL Single-Line Format

*For any* event written to JSONL_Store, the event SHALL occupy exactly one line (no embedded newlines) to enable line-based Git merging.

**Validates: Requirements 7.4**

### Property 24: Git Merge Event Replay

*For any* Git merge that combines JSONL events from different branches, replaying all events in timestamp order SHALL produce a consistent SQLite_Store state.

**Validates: Requirements 7.3**

### Property 25: Agent Session Persistence

*For any* task created in one agent session, when a new agent session starts, the task SHALL be available through GetTask() or query methods.

**Validates: Requirements 8.2**

### Property 26: Workflow Hook Integration

*For any* workflow step that triggers task creation hooks, a corresponding task SHALL be created with metadata linking it to the workflow step.

**Validates: Requirements 8.3**

### Property 27: Checkpoint Task Creation

*For any* checkpoint created when checkpoint task creation is enabled, a task of type "checkpoint" SHALL be created with metadata referencing the checkpoint.

**Validates: Requirements 8.4**

### Property 28: Message Bus Event Publishing

*For any* task operation (create, update, close), a corresponding event SHALL be published to the Message_Bus with task details.

**Validates: Requirements 8.5**

### Property 29: Bulk Ready State Performance

*For any* dataset containing 10,000 tasks with various dependency structures, computing ready state for all tasks SHALL complete within 500ms.

**Validates: Requirements 10.2**

### Property 30: Memory Usage Monitoring

*For any* operation that causes memory usage to exceed 500MB, the Task_Tracker SHALL log a warning message containing optimization suggestions.

**Validates: Requirements 10.5**

### Property 31: SQLite Recovery from JSONL

*For any* corrupted or deleted SQLite_Store, calling Sync() or initialization SHALL rebuild the SQLite database from JSONL_Store events, restoring all tasks and dependencies.

**Validates: Requirements 11.1**

### Property 32: JSONL Corruption Graceful Degradation

*For any* corrupted JSONL_Store, the Task_Tracker SHALL log errors, skip corrupted events, and continue operating with SQLite_Store data.

**Validates: Requirements 11.2**

### Property 33: Transaction Rollback on Failure

*For any* multi-step operation (e.g., create task + add dependencies) that fails partway through, the Task_Tracker SHALL rollback all changes, leaving the system in its pre-operation state.

**Validates: Requirements 11.4**

### Property 34: SQLite-Only Mode Operation

*For any* configuration with git_enabled set to false, all task operations SHALL succeed using only SQLite_Store, and no JSONL files SHALL be created or modified.

**Validates: Requirements 12.3**

### Property 35: Historical Task Loading on Initialization

*For any* existing JSONL_Store with historical events, initializing a new Task_Tracker instance SHALL load all tasks and dependencies from the JSONL events into SQLite_Store.

**Validates: Requirements 12.5**

## Error Handling

### Error Categories

1. **Validation Errors**: Invalid input data (missing required fields, invalid types)
2. **Constraint Errors**: Business rule violations (cycles in dependencies, duplicate IDs)
3. **Storage Errors**: Database or file system failures
4. **Synchronization Errors**: Conflicts or failures during sync operations
5. **System Errors**: Resource exhaustion, initialization failures

### Error Response Format

All errors follow a consistent structure:

```go
type Error struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    Retryable  bool                   `json:"retryable"`
}
```

### Error Handling Strategies

**Validation Errors**:
- Return immediately with descriptive error
- Do not modify any state
- Provide specific field-level error details

**Storage Errors**:
- Retry with exponential backoff for transient failures
- Fallback to alternative storage layer if available
- Log errors for monitoring and debugging

**Synchronization Errors**:
- Queue failed sync operations for retry
- Use last-write-wins for conflict resolution
- Provide manual merge tools for complex conflicts

**System Errors**:
- Attempt graceful degradation (e.g., SQLite-only mode)
- Provide clear error messages with remediation steps
- Log detailed context for troubleshooting

### Recovery Mechanisms

1. **SQLite Corruption**: Rebuild from JSONL events
2. **JSONL Corruption**: Continue with SQLite, log errors
3. **Sync Failures**: Retry with exponential backoff
4. **Git Conflicts**: Provide merge tool or last-write-wins
5. **Resource Exhaustion**: Log warnings, suggest optimization

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit tests and property-based tests to ensure comprehensive coverage:

- **Unit Tests**: Validate specific examples, edge cases, and error conditions
- **Property Tests**: Verify universal properties across all inputs using randomized testing

Both approaches are complementary and necessary. Unit tests catch concrete bugs in specific scenarios, while property tests verify general correctness across a wide input space.

### Property-Based Testing Configuration

**Library Selection**: Use `gopter` (Go property testing library) for property-based tests.

**Test Configuration**:
- Minimum 100 iterations per property test (due to randomization)
- Each property test references its design document property
- Tag format: `// Feature: beads-task-tracking, Property {number}: {property_text}`

**Example Property Test Structure**:

```go
func TestProperty1_DualLayerStorageConsistency(t *testing.T) {
    // Feature: beads-task-tracking, Property 1: Dual-Layer Storage Consistency
    
    properties := gopter.NewProperties(nil)
    properties.Property("task data identical in both stores", prop.ForAll(
        func(task *Task) bool {
            tracker := setupTestTracker(t)
            defer tracker.Stop(context.Background())
            
            taskID, err := tracker.CreateTask(context.Background(), task)
            if err != nil {
                return false
            }
            
            sqliteTask, _ := tracker.sqliteStore.ReadTask(context.Background(), taskID)
            jsonlEvents, _ := tracker.jsonlStore.ReadEvents(context.Background(), time.Time{})
            jsonlTask := reconstructTaskFromEvents(jsonlEvents, taskID)
            
            return tasksEqual(sqliteTask, jsonlTask)
        },
        genTask(),
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Unit Testing Focus Areas

Unit tests should focus on:

1. **Specific Examples**: Concrete scenarios demonstrating correct behavior
2. **Edge Cases**: Empty inputs, boundary conditions, special characters
3. **Error Conditions**: Invalid inputs, constraint violations, system failures
4. **Integration Points**: Interactions between components (MCP ↔ Tracker, Tracker ↔ Stores)

**Example Unit Test**:

```go
func TestCreateTask_MissingRequiredFields(t *testing.T) {
    tracker := setupTestTracker(t)
    defer tracker.Stop(context.Background())
    
    tests := []struct {
        name    string
        task    *Task
        wantErr string
    }{
        {
            name:    "missing type",
            task:    &Task{Title: "Test"},
            wantErr: "type is required",
        },
        {
            name:    "missing title",
            task:    &Task{Type: TaskTypeTask},
            wantErr: "title is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := tracker.CreateTask(context.Background(), tt.task)
            if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
                t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
            }
        })
    }
}
```

### Test Coverage Goals

- **Unit Test Coverage**: 80%+ of code paths
- **Property Test Coverage**: All correctness properties from design document
- **Integration Test Coverage**: All MCP tools and agent integration points
- **Performance Test Coverage**: All performance requirements (query latency, bulk operations)

### Testing Tools and Frameworks

- **Unit Testing**: Go standard `testing` package
- **Property Testing**: `gopter` library
- **Mocking**: `gomock` for interface mocking
- **Assertions**: `testify/assert` for readable assertions
- **Test Fixtures**: Shared test data generators for tasks, dependencies, events

### Continuous Testing

- Run unit tests on every commit
- Run property tests (with reduced iterations) on every commit
- Run full property test suite (100+ iterations) nightly
- Run performance tests weekly or on performance-critical changes
- Monitor test execution time and flakiness

## Integration with AgentFramework

### Agent Integration Points

**ChatAgent Integration**:
```go
// ChatAgent can query tasks during conversation
func (a *ChatAgent) handleTaskQuery(ctx context.Context, query string) (string, error) {
    tasks, err := a.taskTracker.GetReady(ctx)
    if err != nil {
        return "", err
    }
    
    return formatTasksForChat(tasks), nil
}
```

**ReActAgent Integration**:
```go
// ReActAgent uses tasks as action context
func (a *ReActAgent) selectAction(ctx context.Context) (*Action, error) {
    readyTasks, err := a.taskTracker.GetReady(ctx)
    if err != nil {
        return nil, err
    }
    
    // Use tasks to inform action selection
    return a.reasonAboutTasks(readyTasks)
}
```

**WorkerAgent Integration**:
```go
// WorkerAgent executes tasks from the tracker
func (a *WorkerAgent) executeNextTask(ctx context.Context) error {
    tasks, err := a.taskTracker.GetReady(ctx)
    if err != nil {
        return err
    }
    
    if len(tasks) == 0 {
        return ErrNoTasksReady
    }
    
    task := tasks[0]
    return a.executeTask(ctx, task)
}
```

### Workflow System Integration

```go
// Workflow creates tasks for each step
func (w *Workflow) createTasksFromSteps(ctx context.Context, tracker TaskTracker) error {
    var prevTaskID string
    
    for _, step := range w.Steps {
        task := &Task{
            Type:        TaskTypeTask,
            Title:       step.Name,
            Description: step.Description,
            Status:      StatusOpen,
        }
        
        taskID, err := tracker.CreateTask(ctx, task)
        if err != nil {
            return err
        }
        
        // Create dependency on previous step
        if prevTaskID != "" {
            err = tracker.AddDependency(ctx, prevTaskID, taskID, DependencyTypeBlocks)
            if err != nil {
                return err
            }
        }
        
        prevTaskID = taskID
    }
    
    return nil
}
```

### Checkpoint System Integration

```go
// Checkpoint system creates checkpoint tasks
func (c *CheckpointManager) createCheckpoint(ctx context.Context, state interface{}) error {
    // Save checkpoint data
    checkpointID, err := c.saveCheckpoint(state)
    if err != nil {
        return err
    }
    
    // Optionally create task
    if c.config.CreateTasks {
        task := &Task{
            Type:        TaskTypeCheckpoint,
            Title:       fmt.Sprintf("Checkpoint %s", checkpointID),
            Description: "System checkpoint created",
            Status:      StatusCompleted,
            Metadata: map[string]string{
                "checkpoint_id": checkpointID,
            },
        }
        
        _, err = c.taskTracker.CreateTask(ctx, task)
        return err
    }
    
    return nil
}
```

### Message Bus Integration

```go
// Task tracker publishes events to message bus
func (t *TaskTracker) publishEvent(ctx context.Context, event *Event) error {
    message := &Message{
        Type:      "task_event",
        Payload:   event,
        Timestamp: time.Now(),
    }
    
    return t.messageBus.Publish(ctx, "tasks", message)
}

// Agents subscribe to task events
func (a *Agent) subscribeToTaskEvents(ctx context.Context) error {
    return a.messageBus.Subscribe(ctx, "tasks", func(msg *Message) error {
        event := msg.Payload.(*Event)
        return a.handleTaskEvent(ctx, event)
    })
}
```

## Deployment Considerations

### Directory Structure

```
project-root/
├── .beads/
│   ├── config.json
│   ├── tasks/
│   │   ├── 2024-01.jsonl
│   │   ├── 2024-02.jsonl
│   │   └── 2024-03.jsonl
│   └── beads.db (SQLite)
├── .git/
└── ... (other project files)
```

### Configuration File Format

```json
{
  "storage_path": ".beads",
  "git_enabled": true,
  "sync_interval": "5s",
  "max_tasks": 50000,
  "db_path": ".beads/beads.db",
  "jsonl_path": ".beads/tasks"
}
```

### Initialization Sequence

1. Load configuration from `.beads/config.json` or use defaults
2. Create directories if they don't exist
3. Initialize SQLite database with schema
4. Load all JSONL events and replay into SQLite
5. Start sync daemon
6. Register MCP tools
7. Publish initialization event to message bus

### Performance Tuning

**SQLite Optimizations**:
- Use WAL mode for better concurrency
- Create indexes on frequently queried columns
- Use prepared statements for repeated queries
- Batch writes in transactions

**JSONL Optimizations**:
- Partition by month to keep file sizes manageable
- Use buffered writes to reduce I/O
- Compress old JSONL files (optional)

**Memory Optimizations**:
- Stream large query results instead of loading all into memory
- Use connection pooling for SQLite
- Implement LRU cache for frequently accessed tasks

### Monitoring and Observability

**Metrics to Track**:
- Task creation rate
- Query latency (p50, p95, p99)
- Sync lag (time between JSONL write and SQLite update)
- Error rates by category
- Memory usage
- Database size

**Logging**:
- Structured logging with context (task_id, operation, timestamp)
- Log levels: DEBUG, INFO, WARN, ERROR
- Separate logs for sync operations
- Audit log for all task modifications

### Security Considerations

**Access Control**:
- Task tracker operates within agent's security context
- No built-in authentication (relies on agent framework)
- File system permissions protect .beads/ directory

**Data Validation**:
- Validate all inputs before processing
- Sanitize task titles and descriptions
- Limit task size to prevent DoS
- Validate dependency graph complexity

**Git Security**:
- Use signed commits for audit trail
- Validate JSONL format before replay
- Detect and reject malicious events
- Implement event schema versioning

## Future Enhancements

### Potential Extensions

1. **Task Templates**: Predefined task structures for common workflows
2. **Task Scheduling**: Time-based task activation
3. **Task Priorities**: Priority-based task ordering
4. **Task Notifications**: Alert agents when tasks become ready
5. **Task Analytics**: Metrics on task completion rates, cycle times
6. **Multi-Repository Support**: Track tasks across multiple Git repositories
7. **Task Search**: Full-text search across task titles and descriptions
8. **Task Attachments**: Link files or artifacts to tasks
9. **Task Comments**: Agent collaboration through task comments
10. **Task History Visualization**: Timeline view of task evolution

### Scalability Improvements

1. **Distributed SQLite**: Shard tasks across multiple databases
2. **Event Streaming**: Use Kafka or similar for event distribution
3. **Read Replicas**: Multiple SQLite instances for read scaling
4. **Caching Layer**: Redis cache for hot tasks
5. **Async Operations**: Background processing for expensive operations

### Integration Opportunities

1. **GitHub Issues**: Sync with GitHub issue tracker
2. **Jira Integration**: Bidirectional sync with Jira
3. **Slack Notifications**: Post task updates to Slack
4. **Prometheus Metrics**: Export metrics for monitoring
5. **OpenTelemetry**: Distributed tracing for task operations
