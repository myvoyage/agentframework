# Beads Task Tracking Integration

This package implements a Git-native distributed task tracking engine for the AgentFramework system. It provides AI agents with persistent memory across sessions, complex dependency graph management, and distributed collaboration capabilities using a dual-layer architecture (SQLite + JSONL).

## Package Structure

```
pkg/beads/
├── models.go              # Core data models (Task, Dependency, Event, Config)
├── interfaces.go          # Interface definitions for all components
├── models_test.go         # Unit tests for data models
├── store/                 # Storage layer implementations
│   ├── sqlite_store.go    # SQLite store for fast queries
│   └── jsonl_store.go     # JSONL store for Git-tracked persistence
├── tracker/               # Core task tracking logic
│   ├── tracker.go         # Main task tracker implementation
│   ├── event_processor.go # Event sourcing processor
│   └── dependency_resolver.go # Dependency graph management
├── sync/                  # Synchronization daemon
│   └── daemon.go          # Bidirectional sync between stores
└── mcp/                   # Model Context Protocol interface
    └── interface.go       # MCP tools for agent interaction
```

## Core Data Models

### Task
Represents a unit of work with metadata, status, dependencies, and Git-tracked history.

**Types**: epic, task, bug, feature, research, checkpoint  
**Statuses**: open, in_progress, blocked, completed, cancelled

### Dependency
Represents relationships between tasks.

**Types**: blocks, parent-child, related, discovered-from

### Event
Immutable event in the event sourcing system.

**Types**: task_created, task_updated, task_closed, dependency_added, dependency_removed

### Config
Configuration for the task tracker.

## Key Interfaces

### TaskTracker
Central coordinator managing task lifecycle, dependencies, and storage.

### SQLiteStore
Fast local query engine with indexed access.

### JSONLStore
Git-tracked append-only event log.

### SyncDaemon
Bidirectional synchronization between SQLite and JSONL stores.

### DependencyResolver
Computes task ready state and validates dependency graphs.

### EventProcessor
Processes events and applies them to storage layers.

## Dependencies

- **github.com/mattn/go-sqlite3**: SQLite database driver
- **github.com/leanovate/gopter**: Property-based testing library

## Design Principles

1. **Event Sourcing**: All task changes are immutable events
2. **Dual-Layer Storage**: SQLite for queries, JSONL for persistence
3. **Hash-Based IDs**: SHA-256 prevents distributed conflicts
4. **Query-Driven**: Agents query for ready tasks, not parse markdown
5. **Git-Native**: JSONL format enables line-based merging

## Implementation Status

- [x] Task 1: Project structure and data models
- [ ] Task 2: JSONL Store implementation
- [ ] Task 3: SQLite Store implementation
- [ ] Task 5: Event Processor implementation
- [ ] Task 6: Dependency Resolver implementation
- [ ] Task 7: Task Tracker Core implementation
- [ ] Task 9: Sync Daemon implementation
- [ ] Task 10: MCP Interface implementation

## Testing

Run tests with:
```bash
go test ./pkg/beads/...
```

Run tests with coverage:
```bash
go test -cover ./pkg/beads/...
```

## Usage

(To be documented as implementation progresses)

## References

- Requirements: `.kiro/specs/beads-task-tracking/requirements.md`
- Design: `.kiro/specs/beads-task-tracking/design.md`
- Tasks: `.kiro/specs/beads-task-tracking/tasks.md`
