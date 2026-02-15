# Requirements Document: Beads Task Tracking Integration

## Introduction

This document specifies the requirements for integrating a Git-native distributed task tracking engine (inspired by Beads) into the AgentFramework system. The system will provide AI agents with persistent memory across sessions, complex dependency graph management, and distributed collaboration capabilities using a dual-layer architecture (SQLite + JSONL).

## Glossary

- **Task_Tracker**: The core system that manages task lifecycle, dependencies, and persistence
- **Task**: A unit of work with metadata, status, dependencies, and Git-tracked history
- **Dependency_Graph**: A directed acyclic graph (DAG) representing relationships between tasks
- **JSONL_Store**: Git-tracked append-only log storing task events in JSON Lines format
- **SQLite_Store**: Local database providing fast query capabilities for task data
- **Sync_Daemon**: Background process synchronizing data between SQLite and JSONL stores
- **MCP_Interface**: Model Context Protocol tools exposing task operations to agents
- **Task_ID**: SHA-256 hash-based unique identifier preventing distributed conflicts
- **Ready_State**: Computed state indicating a task has no blocking dependencies
- **Agent_Memory**: Persistent task context available across agent sessions
- **Dependency_Type**: One of: blocks, parent-child, related, discovered-from

## Requirements

### Requirement 1: Dual-Layer Storage Architecture

**User Story:** As a system architect, I want a dual-layer storage system, so that agents have both fast local queries and Git-tracked persistence.

#### Acceptance Criteria

1. THE Task_Tracker SHALL maintain task data in both SQLite_Store and JSONL_Store simultaneously
2. WHEN a task is created, THE Task_Tracker SHALL write to SQLite_Store immediately and append to JSONL_Store
3. WHEN the JSONL_Store is modified externally (git pull), THE Sync_Daemon SHALL update SQLite_Store to reflect changes
4. WHEN SQLite_Store is queried, THE Task_Tracker SHALL return results within 100ms for datasets up to 10,000 tasks
5. THE JSONL_Store SHALL use append-only semantics to enable conflict-free merging across Git branches

### Requirement 2: Task Entity Model

**User Story:** As an agent, I want to create and manage tasks with rich metadata, so that I can track work across sessions.

#### Acceptance Criteria

1. WHEN a task is created, THE Task_Tracker SHALL generate a Task_ID using SHA-256 hash of (timestamp + random_bytes)
2. THE Task_Tracker SHALL support task types: epic, task, bug, feature, research, checkpoint
3. WHEN storing a task, THE Task_Tracker SHALL include fields: id, type, title, description, status, created_at, updated_at, assignee, tags
4. THE Task_Tracker SHALL support task statuses: open, in_progress, blocked, completed, cancelled
5. WHEN a task is updated, THE Task_Tracker SHALL append a new event to JSONL_Store preserving full history

### Requirement 3: Dependency Graph Management

**User Story:** As an agent, I want to define dependencies between tasks, so that I can model complex workflows.

#### Acceptance Criteria

1. THE Task_Tracker SHALL support four Dependency_Types: blocks, parent-child, related, discovered-from
2. WHEN a dependency is added, THE Task_Tracker SHALL validate that it does not create cycles in the Dependency_Graph
3. WHEN computing Ready_State, THE Task_Tracker SHALL mark a task as ready only if all "blocks" dependencies are completed
4. THE Task_Tracker SHALL allow parent-child relationships to form hierarchical task structures
5. WHEN a task is queried, THE Task_Tracker SHALL return all its dependencies with their types and current statuses

### Requirement 4: Query Interface for Agents

**User Story:** As an agent, I want to query tasks by various criteria, so that I can find relevant work to execute.

#### Acceptance Criteria

1. THE Task_Tracker SHALL provide a GetReady() method returning all tasks in Ready_State
2. WHEN GetReady() is called, THE Task_Tracker SHALL return tasks in JSON format with full metadata
3. THE Task_Tracker SHALL provide query methods: GetByID(), GetByStatus(), GetByAssignee(), GetByTags()
4. WHEN querying by tags, THE Task_Tracker SHALL support AND/OR logical operations
5. THE Task_Tracker SHALL provide GetDependencies() and GetDependents() methods for graph traversal

### Requirement 5: Synchronization Mechanism

**User Story:** As a system operator, I want automatic synchronization between storage layers, so that data remains consistent.

#### Acceptance Criteria

1. THE Sync_Daemon SHALL run as a background goroutine monitoring JSONL_Store for changes
2. WHEN JSONL_Store is modified, THE Sync_Daemon SHALL replay events into SQLite_Store within 1 second
3. WHEN SQLite_Store is modified, THE Sync_Daemon SHALL append corresponding events to JSONL_Store immediately
4. IF synchronization fails, THEN THE Sync_Daemon SHALL log errors and retry with exponential backoff
5. THE Sync_Daemon SHALL detect and resolve conflicts using last-write-wins strategy based on timestamps

### Requirement 6: MCP Tool Interface

**User Story:** As an agent, I want MCP tools for task operations, so that I can interact with the task tracker through standard protocols.

#### Acceptance Criteria

1. THE MCP_Interface SHALL provide tools: create_task, update_task, close_task, get_ready_tasks, show_task, add_dependency
2. WHEN create_task is called, THE MCP_Interface SHALL validate required fields (type, title) and return the Task_ID
3. WHEN get_ready_tasks is called, THE MCP_Interface SHALL return JSON array of tasks in Ready_State
4. WHEN add_dependency is called, THE MCP_Interface SHALL validate both task IDs exist and the Dependency_Type is valid
5. THE MCP_Interface SHALL return structured error responses with error codes and human-readable messages

### Requirement 7: Git Integration and Distributed Collaboration

**User Story:** As a team of agents, we want to collaborate on tasks across different branches, so that we can work in parallel without conflicts.

#### Acceptance Criteria

1. THE JSONL_Store SHALL store task events in files under .beads/ directory within the Git repository
2. WHEN two agents create tasks simultaneously, THE Task_Tracker SHALL use hash-based Task_IDs to prevent collisions
3. WHEN Git branches are merged, THE Task_Tracker SHALL replay all JSONL events to rebuild SQLite_Store
4. THE JSONL_Store SHALL use one line per event to enable line-based Git merging
5. WHEN a Git conflict occurs in JSONL files, THE Task_Tracker SHALL provide a merge tool to resolve conflicts

### Requirement 8: Agent Framework Integration

**User Story:** As a framework developer, I want task tracking integrated with existing agent types, so that agents can use tasks naturally.

#### Acceptance Criteria

1. THE Task_Tracker SHALL integrate with ChatAgent, ReActAgent, and WorkerAgent through a common interface
2. WHEN an agent starts a session, THE Task_Tracker SHALL load Agent_Memory from previous sessions
3. THE Task_Tracker SHALL provide hooks for workflow system to create tasks from workflow steps
4. WHEN a checkpoint is created, THE Task_Tracker SHALL optionally create a checkpoint task
5. THE Task_Tracker SHALL publish task events to the existing Message_Bus for multi-agent coordination

### Requirement 9: Cross-Platform Compatibility

**User Story:** As a system operator, I want the task tracker to work on all major platforms, so that agents can run anywhere.

#### Acceptance Criteria

1. THE Task_Tracker SHALL use Go's filepath package for cross-platform path handling
2. THE Task_Tracker SHALL use SQLite with CGO or pure-Go implementation for Windows/Linux/macOS compatibility
3. WHEN initializing on Windows, THE Task_Tracker SHALL handle Windows-specific file locking semantics
4. THE Task_Tracker SHALL use Git commands through exec.Command with platform-appropriate shell detection
5. THE Task_Tracker SHALL provide configuration for custom Git executable paths

### Requirement 10: Performance and Scalability

**User Story:** As a system operator, I want the task tracker to handle large task sets efficiently, so that it scales with agent workloads.

#### Acceptance Criteria

1. THE Task_Tracker SHALL support at least 10,000 tasks with query response times under 100ms
2. WHEN computing Ready_State for all tasks, THE Task_Tracker SHALL complete within 500ms for 10,000 tasks
3. THE Task_Tracker SHALL use database indexes on: task_id, status, assignee, created_at
4. THE Sync_Daemon SHALL batch JSONL writes to reduce I/O overhead
5. WHEN memory usage exceeds 500MB, THE Task_Tracker SHALL log a warning and suggest optimization

### Requirement 11: Error Handling and Recovery

**User Story:** As a system operator, I want robust error handling, so that the task tracker recovers from failures gracefully.

#### Acceptance Criteria

1. IF SQLite_Store is corrupted, THEN THE Task_Tracker SHALL rebuild it from JSONL_Store
2. IF JSONL_Store is corrupted, THEN THE Task_Tracker SHALL log errors and continue with SQLite_Store data
3. WHEN a task operation fails, THE Task_Tracker SHALL return structured errors with context
4. THE Task_Tracker SHALL implement transaction rollback for multi-step operations
5. WHEN initialization fails, THE Task_Tracker SHALL provide clear error messages with remediation steps

### Requirement 12: Configuration and Initialization

**User Story:** As a developer, I want simple configuration and initialization, so that I can integrate the task tracker easily.

#### Acceptance Criteria

1. THE Task_Tracker SHALL initialize with a single NewTaskTracker() constructor accepting a config struct
2. THE Task_Tracker SHALL support configuration options: storage_path, git_enabled, sync_interval, max_tasks
3. WHEN git_enabled is false, THE Task_Tracker SHALL operate in SQLite-only mode for testing
4. THE Task_Tracker SHALL auto-create necessary directories (.beads/, .beads/tasks/) on initialization
5. WHEN initializing in an existing repository, THE Task_Tracker SHALL load all historical tasks from JSONL_Store
