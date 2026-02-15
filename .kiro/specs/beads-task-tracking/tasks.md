# Implementation Plan: Beads Task Tracking Integration

## Overview

This implementation plan breaks down the Beads task tracking integration into incremental, testable steps. The approach follows a bottom-up strategy: first establishing the storage layers, then building the core tracker logic, adding dependency management, implementing the MCP interface, and finally integrating with the AgentFramework ecosystem.

Each task builds on previous work, with property-based tests integrated throughout to catch errors early. The plan includes checkpoints to ensure stability before proceeding to more complex features.

## Tasks

- [x] 1. Set up project structure and data models
  - Create directory structure: `pkg/beads/` with subdirectories for `store`, `tracker`, `mcp`, `sync`
  - Define core data models: Task, Dependency, Event, Config structs in `pkg/beads/models.go`
  - Define interfaces: TaskTracker, SQLiteStore, JSONLStore, SyncDaemon in `pkg/beads/interfaces.go`
  - Set up Go module dependencies: SQLite driver, JSONL handling, testing libraries (gopter)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.1, 12.1, 12.2_

- [ ] 2. Implement JSONL Store
  - [x] 2.1 Create JSONLStore implementation with file partitioning by month
    - Implement AppendEvent() with buffered writes
    - Implement ReadEvents() with time-based filtering
    - Implement ReadAllEvents() for full history replay
    - Implement GetLatestTimestamp() for sync tracking
    - Handle file creation and directory management
    - _Requirements: 1.5, 7.1, 7.4_
  
  - [x] 2.2 Write property test for JSONL append-only semantics

    - **Property 4: Append-Only JSONL Semantics**
    - **Validates: Requirements 1.5**
  
  - [x] 2.3 Write property test for JSONL single-line format

    - **Property 23: JSONL Single-Line Format**
    - **Validates: Requirements 7.4**
  
  - [ ]* 2.4 Write unit tests for JSONL edge cases
    - Test empty JSONL files
    - Test month boundary transitions
    - Test concurrent writes
    - _Requirements: 7.1, 7.4_

- [ ] 3. Implement SQLite Store
  - [x] 3.1 Create SQLite schema with tables and indexes
    - Create tasks table with all required fields
    - Create task_tags table for tag relationships
    - Create dependencies table for dependency graph
    - Add indexes on task_id, status, assignee, created_at
    - Configure WAL mode for better concurrency
    - _Requirements: 1.4, 2.3, 3.1, 10.3_
  
  - [x] 3.2 Implement SQLiteStore write operations
    - Implement WriteTask() with transaction support
    - Implement WriteDependency() with foreign key validation
    - Implement batch write operations for performance
    - _Requirements: 1.1, 2.3, 3.1_
  
  - [ ] 3.3 Implement SQLiteStore query operations
    - Implement ReadTask() by ID
    - Implement QueryTasks() with flexible filtering
    - Implement ReadDependencies() with direction support (incoming/outgoing)
    - Optimize queries with prepared statements
    - _Requirements: 4.1, 4.3, 4.5_
  
  - [ ] 3.4 Implement RebuildFromEvents() for recovery
    - Parse events and apply them in chronological order
    - Handle all event types (task_created, task_updated, dependency_added, etc.)
    - Use transactions for atomic rebuild
    - _Requirements: 11.1, 12.5_
  
  - [ ]* 3.5 Write property test for query performance
    - **Property 3: Query Performance at Scale**
    - **Validates: Requirements 1.4**
  
  - [ ]* 3.6 Write property test for task serialization completeness
    - **Property 6: Task Serialization Completeness**
    - **Validates: Requirements 2.3**
  
  - [ ]* 3.7 Write unit tests for SQLite edge cases
    - Test empty database queries
    - Test invalid task IDs
    - Test transaction rollback scenarios
    - _Requirements: 4.3, 11.4_

- [ ] 4. Checkpoint - Verify storage layers work independently
  - Ensure all tests pass for JSONL and SQLite stores
  - Manually test creating tasks in each store
  - Ask the user if questions arise

- [ ] 5. Implement Event Processor
  - [ ] 5.1 Create EventProcessor with event type handling
    - Implement ProcessEvent() with switch on event type
    - Implement handlers for each event type: task_created, task_updated, task_closed, dependency_added, dependency_removed
    - Implement ReplayEvents() for batch processing
    - Add event validation before processing
    - _Requirements: 2.5, 7.3, 11.1_
  
  - [ ]* 5.2 Write property test for event sourcing history preservation
    - **Property 7: Event Sourcing History Preservation**
    - **Validates: Requirements 2.5**
  
  - [ ]* 5.3 Write unit tests for event processing
    - Test each event type handler
    - Test invalid event handling
    - Test event replay ordering
    - _Requirements: 2.5, 7.3_

- [ ] 6. Implement Dependency Resolver
  - [ ] 6.1 Create DependencyResolver with cycle detection
    - Implement ValidateNoCycles() using DFS algorithm
    - Implement ComputeReadyState() checking blocking dependencies
    - Implement GetBlockingTasks() for dependency queries
    - Implement GetDependencyChain() for full traversal
    - _Requirements: 3.2, 3.3, 3.4, 4.5_
  
  - [ ]* 6.2 Write property test for dependency cycle prevention
    - **Property 8: Dependency Cycle Prevention**
    - **Validates: Requirements 3.2**
  
  - [ ]* 6.3 Write property test for ready state computation
    - **Property 9: Ready State Computation Correctness**
    - **Validates: Requirements 3.3**
  
  - [ ]* 6.4 Write property test for hierarchical task structure
    - **Property 10: Hierarchical Task Structure Integrity**
    - **Validates: Requirements 3.4**
  
  - [ ]* 6.5 Write unit tests for dependency edge cases
    - Test empty dependency graphs
    - Test self-referential dependency attempts
    - Test complex multi-level hierarchies
    - _Requirements: 3.2, 3.3, 3.4_

- [ ] 7. Implement Task Tracker Core
  - [ ] 7.1 Create TaskTracker with initialization logic
    - Implement NewTaskTracker() constructor with config
    - Implement Start() to initialize stores and sync daemon
    - Implement Stop() for graceful shutdown
    - Auto-create .beads/ directories on initialization
    - Load historical tasks from JSONL on startup
    - _Requirements: 12.1, 12.2, 12.4, 12.5_
  
  - [ ] 7.2 Implement task CRUD operations
    - Implement CreateTask() with ID generation (SHA-256 hash)
    - Implement UpdateTask() with event creation
    - Implement GetTask() with error handling
    - Implement CloseTask() with status validation
    - Write to both SQLite and JSONL stores
    - _Requirements: 1.1, 2.1, 2.3, 2.5_
  
  - [ ] 7.3 Implement query operations
    - Implement GetReady() using dependency resolver
    - Implement GetByStatus() with SQLite queries
    - Implement GetByAssignee() with filtering
    - Implement GetByTags() with AND/OR logic support
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [ ] 7.4 Implement dependency operations
    - Implement AddDependency() with cycle validation
    - Implement RemoveDependency() with cleanup
    - Implement GetDependencies() and GetDependents()
    - _Requirements: 3.2, 3.5, 4.5_
  
  - [ ]* 7.5 Write property test for dual-layer storage consistency
    - **Property 1: Dual-Layer Storage Consistency**
    - **Validates: Requirements 1.1**
  
  - [ ]* 7.6 Write property test for task ID uniqueness
    - **Property 5: Task ID Uniqueness and Format**
    - **Validates: Requirements 2.1, 7.2**
  
  - [ ]* 7.7 Write property test for GetReady query correctness
    - **Property 12: GetReady Query Correctness**
    - **Validates: Requirements 4.1**
  
  - [ ]* 7.8 Write property test for query method correctness
    - **Property 14: Query Method Correctness**
    - **Validates: Requirements 4.3**
  
  - [ ]* 7.9 Write property test for tag query logical operations
    - **Property 15: Tag Query Logical Operations**
    - **Validates: Requirements 4.4**
  
  - [ ]* 7.10 Write property test for dependency query completeness
    - **Property 11: Dependency Query Completeness**
    - **Validates: Requirements 3.5**
  
  - [ ]* 7.11 Write property test for graph traversal correctness
    - **Property 16: Graph Traversal Correctness**
    - **Validates: Requirements 4.5**
  
  - [ ]* 7.12 Write unit tests for task tracker operations
    - Test task creation with various types
    - Test task updates and status transitions
    - Test error handling for invalid operations
    - _Requirements: 2.2, 2.4, 4.1, 4.3_

- [ ] 8. Checkpoint - Verify core tracker functionality
  - Ensure all tests pass for task tracker
  - Manually test creating tasks, adding dependencies, querying ready tasks
  - Verify both storage layers are updated correctly
  - Ask the user if questions arise

- [ ] 9. Implement Sync Daemon
  - [ ] 9.1 Create SyncDaemon with background goroutine
    - Implement Start() launching sync loop goroutine
    - Implement Stop() for graceful shutdown
    - Implement TriggerSync() for manual synchronization
    - Implement GetStatus() for monitoring
    - _Requirements: 5.1, 5.2_
  
  - [ ] 9.2 Implement JSONL to SQLite synchronization
    - Monitor JSONL directory for changes (file watcher or polling)
    - Read new events since last sync timestamp
    - Replay events into SQLite using EventProcessor
    - Update last sync timestamp
    - _Requirements: 1.3, 5.2, 7.3_
  
  - [ ] 9.3 Implement error handling and retry logic
    - Implement exponential backoff for failed syncs
    - Log sync errors with context
    - Implement conflict resolution using last-write-wins
    - _Requirements: 5.4, 5.5_
  
  - [ ]* 9.4 Write property test for JSONL synchronization
    - **Property 2: JSONL Synchronization to SQLite**
    - **Validates: Requirements 1.3**
  
  - [ ]* 9.5 Write property test for sync latency
    - **Property 17: Sync Latency Guarantee**
    - **Validates: Requirements 5.2**
  
  - [ ]* 9.6 Write property test for sync retry with exponential backoff
    - **Property 18: Sync Retry with Exponential Backoff**
    - **Validates: Requirements 5.4**
  
  - [ ]* 9.7 Write property test for conflict resolution
    - **Property 19: Conflict Resolution Last-Write-Wins**
    - **Validates: Requirements 5.5**
  
  - [ ]* 9.8 Write unit tests for sync daemon
    - Test sync loop execution
    - Test manual trigger
    - Test graceful shutdown
    - _Requirements: 5.1, 5.2, 5.4_

- [ ] 10. Implement MCP Interface
  - [ ] 10.1 Define MCP tool schemas
    - Define input schemas for all tools: create_task, update_task, close_task, get_ready_tasks, show_task, add_dependency, list_tasks
    - Define output schemas with success/error structure
    - Define error codes and messages
    - _Requirements: 6.1, 6.5_
  
  - [ ] 10.2 Implement MCP tool handlers
    - Implement CreateTask() tool with validation
    - Implement UpdateTask() tool
    - Implement CloseTask() tool
    - Implement GetReadyTasks() tool returning JSON
    - Implement ShowTask() tool
    - Implement AddDependency() tool with validation
    - Implement ListTasks() tool with filtering
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  
  - [ ] 10.3 Implement error response formatting
    - Create structured error responses with code and message
    - Add context details to errors
    - Ensure all errors follow consistent format
    - _Requirements: 6.5_
  
  - [ ]* 10.4 Write property test for MCP input validation
    - **Property 20: MCP Input Validation**
    - **Validates: Requirements 6.2**
  
  - [ ]* 10.5 Write property test for MCP dependency validation
    - **Property 21: MCP Dependency Validation**
    - **Validates: Requirements 6.4**
  
  - [ ]* 10.6 Write property test for MCP error response structure
    - **Property 22: MCP Error Response Structure**
    - **Validates: Requirements 6.5**
  
  - [ ]* 10.7 Write property test for GetReady JSON format validity
    - **Property 13: GetReady JSON Format Validity**
    - **Validates: Requirements 4.2**
  
  - [ ]* 10.8 Write unit tests for MCP tools
    - Test each tool with valid inputs
    - Test each tool with invalid inputs
    - Test error response formatting
    - _Requirements: 6.1, 6.2, 6.4, 6.5_

- [ ] 11. Checkpoint - Verify MCP interface works end-to-end
  - Ensure all tests pass for MCP interface
  - Manually test each MCP tool through the interface
  - Verify error responses are properly formatted
  - Ask the user if questions arise

- [ ] 12. Implement error handling and recovery
  - [ ] 12.1 Implement SQLite recovery from JSONL
    - Detect corrupted SQLite database on initialization
    - Trigger full rebuild from JSONL events
    - Validate rebuilt database integrity
    - _Requirements: 11.1_
  
  - [ ] 12.2 Implement JSONL corruption handling
    - Detect corrupted JSONL lines during read
    - Log errors with line numbers and content
    - Skip corrupted events and continue processing
    - _Requirements: 11.2_
  
  - [ ] 12.3 Implement transaction rollback
    - Wrap multi-step operations in transactions
    - Implement rollback on any step failure
    - Ensure atomic operations for task + dependency creation
    - _Requirements: 11.4_
  
  - [ ]* 12.4 Write property test for SQLite recovery
    - **Property 31: SQLite Recovery from JSONL**
    - **Validates: Requirements 11.1**
  
  - [ ]* 12.5 Write property test for JSONL corruption graceful degradation
    - **Property 32: JSONL Corruption Graceful Degradation**
    - **Validates: Requirements 11.2**
  
  - [ ]* 12.6 Write property test for transaction rollback
    - **Property 33: Transaction Rollback on Failure**
    - **Validates: Requirements 11.4**
  
  - [ ]* 12.7 Write unit tests for error scenarios
    - Test initialization with corrupted SQLite
    - Test reading corrupted JSONL files
    - Test transaction rollback scenarios
    - _Requirements: 11.1, 11.2, 11.4, 11.5_

- [ ] 13. Implement Git integration features
  - [ ] 13.1 Implement Git merge event replay
    - Detect Git merge by checking for new JSONL events
    - Replay all events in timestamp order
    - Rebuild SQLite from merged event history
    - _Requirements: 7.3_
  
  - [ ] 13.2 Implement SQLite-only mode for testing
    - Add git_enabled configuration flag
    - Skip JSONL operations when git_enabled is false
    - Ensure all operations work with SQLite only
    - _Requirements: 12.3_
  
  - [ ]* 13.3 Write property test for Git merge event replay
    - **Property 24: Git Merge Event Replay**
    - **Validates: Requirements 7.3**
  
  - [ ]* 13.4 Write property test for SQLite-only mode
    - **Property 34: SQLite-Only Mode Operation**
    - **Validates: Requirements 12.3**
  
  - [ ]* 13.5 Write unit tests for Git integration
    - Test merge detection
    - Test SQLite-only mode configuration
    - Test event replay after merge
    - _Requirements: 7.3, 12.3_

- [ ] 14. Implement AgentFramework integration
  - [ ] 14.1 Create agent integration helpers
    - Create helper functions for ChatAgent, ReActAgent, WorkerAgent
    - Implement task query methods for agents
    - Implement task execution helpers
    - _Requirements: 8.1, 8.2_
  
  - [ ] 14.2 Implement workflow system hooks
    - Create hook for workflow step to task conversion
    - Implement automatic dependency creation for sequential steps
    - Add metadata linking tasks to workflow steps
    - _Requirements: 8.3_
  
  - [ ] 14.3 Implement checkpoint system integration
    - Add configuration for checkpoint task creation
    - Create checkpoint tasks with metadata
    - Link checkpoint tasks to checkpoint IDs
    - _Requirements: 8.4_
  
  - [ ] 14.4 Implement message bus event publishing
    - Publish task events to message bus on all operations
    - Include full task details in events
    - Add event types for task lifecycle
    - _Requirements: 8.5_
  
  - [ ]* 14.5 Write property test for agent session persistence
    - **Property 25: Agent Session Persistence**
    - **Validates: Requirements 8.2**
  
  - [ ]* 14.6 Write property test for workflow hook integration
    - **Property 26: Workflow Hook Integration**
    - **Validates: Requirements 8.3**
  
  - [ ]* 14.7 Write property test for checkpoint task creation
    - **Property 27: Checkpoint Task Creation**
    - **Validates: Requirements 8.4**
  
  - [ ]* 14.8 Write property test for message bus event publishing
    - **Property 28: Message Bus Event Publishing**
    - **Validates: Requirements 8.5**
  
  - [ ]* 14.9 Write integration tests for agent framework
    - Test ChatAgent task queries
    - Test ReActAgent task execution
    - Test WorkerAgent task processing
    - Test workflow to task conversion
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 15. Implement cross-platform compatibility
  - [ ] 15.1 Add cross-platform path handling
    - Use filepath.Join() for all path operations
    - Test on Windows, Linux, macOS
    - Handle platform-specific path separators
    - _Requirements: 9.1, 9.2_
  
  - [ ] 15.2 Add platform-specific file locking
    - Implement Windows file locking using appropriate syscalls
    - Implement Unix file locking using flock
    - Test concurrent access on each platform
    - _Requirements: 9.3_
  
  - [ ] 15.3 Add Git command execution with platform detection
    - Detect Git executable path on each platform
    - Support custom Git paths via configuration
    - Handle platform-specific shell differences
    - _Requirements: 9.4, 9.5_
  
  - [ ]* 15.4 Write unit tests for cross-platform features
    - Test path handling on different platforms
    - Test file locking on Windows and Unix
    - Test Git command execution
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 16. Implement performance optimizations and monitoring
  - [ ] 16.1 Add performance optimizations
    - Implement connection pooling for SQLite
    - Add LRU cache for frequently accessed tasks
    - Batch JSONL writes to reduce I/O
    - Use prepared statements for repeated queries
    - _Requirements: 10.3, 10.4_
  
  - [ ] 16.2 Add memory usage monitoring
    - Track memory usage during operations
    - Log warnings when memory exceeds 500MB
    - Provide optimization suggestions in warnings
    - _Requirements: 10.5_
  
  - [ ]* 16.3 Write property test for bulk ready state performance
    - **Property 29: Bulk Ready State Performance**
    - **Validates: Requirements 10.2**
  
  - [ ]* 16.4 Write property test for memory usage monitoring
    - **Property 30: Memory Usage Monitoring**
    - **Validates: Requirements 10.5**
  
  - [ ]* 16.5 Write performance tests
    - Test query performance with 10,000 tasks
    - Test bulk ready state computation
    - Test memory usage under load
    - _Requirements: 1.4, 10.2, 10.5_

- [ ] 17. Implement configuration and initialization
  - [ ] 17.1 Create configuration file handling
    - Implement config loading from .beads/config.json
    - Implement config validation
    - Provide sensible defaults for all options
    - Support environment variable overrides
    - _Requirements: 12.1, 12.2_
  
  - [ ] 17.2 Implement initialization error handling
    - Provide clear error messages for initialization failures
    - Include remediation steps in error messages
    - Validate configuration before initialization
    - _Requirements: 11.5_
  
  - [ ]* 17.3 Write property test for historical task loading
    - **Property 35: Historical Task Loading on Initialization**
    - **Validates: Requirements 12.5**
  
  - [ ]* 17.4 Write unit tests for configuration
    - Test config loading from file
    - Test config validation
    - Test default values
    - Test initialization error messages
    - _Requirements: 12.1, 12.2, 11.5_

- [ ] 18. Final checkpoint - Integration testing and documentation
  - [ ] 18.1 Run full integration test suite
    - Test end-to-end workflows: create task → add dependencies → query ready → close task
    - Test multi-agent scenarios with message bus
    - Test Git merge scenarios
    - Test recovery scenarios (corrupted SQLite, corrupted JSONL)
    - _Requirements: All_
  
  - [ ]* 18.2 Run full property test suite with 100+ iterations
    - Execute all 35 property tests with full iteration count
    - Verify all properties pass consistently
    - _Requirements: All_
  
  - [ ] 18.3 Create usage examples and documentation
    - Write example code for basic usage
    - Write example code for agent integration
    - Write example code for workflow integration
    - Document configuration options
    - Document MCP tool usage
  
  - [ ] 18.4 Final verification
    - Ensure all tests pass
    - Verify cross-platform compatibility
    - Verify performance requirements are met
    - Ask the user if questions arise

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each property test references a specific correctness property from the design document
- Checkpoints ensure incremental validation and provide opportunities for user feedback
- All tasks reference specific requirements for traceability
- Property tests use gopter library with minimum 100 iterations
- Unit tests focus on edge cases and error conditions
- Integration tests verify end-to-end workflows and agent interactions
