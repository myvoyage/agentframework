package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"AgentFramework/pkg/beads"
)

// TaskTrackerMCP implements the MCP interface for beads task tracking
type TaskTrackerMCP struct {
	tracker beads.TaskTracker
}

// NewTaskTrackerMCP creates a new TaskTrackerMCP instance
func NewTaskTrackerMCP(tracker beads.TaskTracker) *TaskTrackerMCP {
	return &TaskTrackerMCP{
		tracker: tracker,
	}
}

// MCP tool schema definitions

// CreateTaskInput represents the input for create_task tool
type CreateTaskInput struct {
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status,omitempty"`
	Assignee    string            `json:"assignee,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateTaskOutput represents the output for create_task tool
type CreateTaskOutput struct {
	Success  bool   `json:"success"`
	TaskID   string `json:"task_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// UpdateTaskInput represents the input for update_task tool
type UpdateTaskInput struct {
	TaskID      string            `json:"task_id"`
	Title       *string           `json:"title,omitempty"`
	Description *string           `json:"description,omitempty"`
	Status      *string           `json:"status,omitempty"`
	Assignee    *string           `json:"assignee,omitempty"`
	Tags        *[]string         `json:"tags,omitempty"`
	Metadata    *map[string]string `json:"metadata,omitempty"`
}

// UpdateTaskOutput represents the output for update_task tool
type UpdateTaskOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CloseTaskInput represents the input for close_task tool
type CloseTaskInput struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// CloseTaskOutput represents the output for close_task tool
type CloseTaskOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// GetReadyTasksOutput represents the output for get_ready_tasks tool
type GetReadyTasksOutput struct {
	Success bool              `json:"success"`
	Tasks   []*beads.Task     `json:"tasks,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// ShowTaskInput represents the input for show_task tool
type ShowTaskInput struct {
	TaskID string `json:"task_id"`
}

// ShowTaskOutput represents the output for show_task tool
type ShowTaskOutput struct {
	Success bool          `json:"success"`
	Task    *beads.Task   `json:"task,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// AddDependencyInput represents the input for add_dependency tool
type AddDependencyInput struct {
	FromTaskID string `json:"from_task_id"`
	ToTaskID   string `json:"to_task_id"`
	DepType    string `json:"dep_type,omitempty"`
}

// AddDependencyOutput represents the output for add_dependency tool
type AddDependencyOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ListTasksInput represents the input for list_tasks tool
type ListTasksInput struct {
	Status   string   `json:"status,omitempty"`
	Assignee string   `json:"assignee,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	TagOp    string   `json:"tag_op,omitempty"`
}

// ListTasksOutput represents the output for list_tasks tool
type ListTasksOutput struct {
	Success bool              `json:"success"`
	Tasks   []*beads.Task     `json:"tasks,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// MCP tool handler implementations

// CreateTask handles the create_task tool request
func (m *TaskTrackerMCP) CreateTask(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input CreateTaskInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	task := &beads.Task{
		Type:        beads.TaskType(input.Type),
		Title:       input.Title,
		Description: input.Description,
		Status:      beads.TaskStatus(input.Status),
		Assignee:    input.Assignee,
		Tags:        input.Tags,
		Metadata:    input.Metadata,
	}

	taskID, err := m.tracker.CreateTask(ctx, task)
	if err != nil {
		return json.Marshal(CreateTaskOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(CreateTaskOutput{
		Success: true,
		TaskID:  taskID,
	})
}

// UpdateTask handles the update_task tool request
func (m *TaskTrackerMCP) UpdateTask(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input UpdateTaskInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	var status *beads.TaskStatus
	if input.Status != nil {
		s := beads.TaskStatus(*input.Status)
		status = &s
	}

	var tags *[]string
	if input.Tags != nil {
		t := *input.Tags
		tags = &t
	}

	var metadata *map[string]string
	if input.Metadata != nil {
		md := *input.Metadata
		metadata = &md
	}

	err := m.tracker.UpdateTask(ctx, input.TaskID, beads.TaskUpdate{
		Title:       input.Title,
		Description: input.Description,
		Status:      status,
		Assignee:    input.Assignee,
		Tags:        tags,
		Metadata:    metadata,
	})

	if err != nil {
		return json.Marshal(UpdateTaskOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(UpdateTaskOutput{
		Success: true,
	})
}

// CloseTask handles the close_task tool request
func (m *TaskTrackerMCP) CloseTask(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input CloseTaskInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	err := m.tracker.CloseTask(ctx, input.TaskID, beads.TaskStatus(input.Status))
	if err != nil {
		return json.Marshal(CloseTaskOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(CloseTaskOutput{
		Success: true,
	})
}

// GetReadyTasks handles the get_ready_tasks tool request
func (m *TaskTrackerMCP) GetReadyTasks(ctx context.Context, params json.RawMessage) ([]byte, error) {
	tasks, err := m.tracker.GetReady(ctx)
	if err != nil {
		return json.Marshal(GetReadyTasksOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(GetReadyTasksOutput{
		Success: true,
		Tasks:   tasks,
	})
}

// ShowTask handles the show_task tool request
func (m *TaskTrackerMCP) ShowTask(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input ShowTaskInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	task, err := m.tracker.GetTask(ctx, input.TaskID)
	if err != nil {
		return json.Marshal(ShowTaskOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(ShowTaskOutput{
		Success: true,
		Task:    task,
	})
}

// AddDependency handles the add_dependency tool request
func (m *TaskTrackerMCP) AddDependency(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input AddDependencyInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	err := m.tracker.AddDependency(ctx, input.FromTaskID, input.ToTaskID, beads.DependencyType(input.DepType))
	if err != nil {
		return json.Marshal(AddDependencyOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(AddDependencyOutput{
		Success: true,
	})
}

// ListTasks handles the list_tasks tool request
func (m *TaskTrackerMCP) ListTasks(ctx context.Context, params json.RawMessage) ([]byte, error) {
	var input ListTasksInput
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Determine which query method to use
	if input.Status != "" && input.Assignee != "" {
		// Both status and assignee specified - get by status then filter
		tasks, err := m.tracker.GetByStatus(ctx, beads.TaskStatus(input.Status))
		if err != nil {
			return json.Marshal(ListTasksOutput{
				Success: false,
				Error:   err.Error(),
			})
		}

		// Filter by assignee
		filtered := make([]*beads.Task, 0)
		for _, task := range tasks {
			if task.Assignee == input.Assignee {
				filtered = append(filtered, task)
			}
		}

		return json.Marshal(ListTasksOutput{
			Success: true,
			Tasks:   filtered,
		})
	}

	if input.Status != "" {
		// Query by status
		return m.GetByStatus(ctx, params)
	}

	if input.Assignee != "" {
		// Query by assignee
		return m.GetByAssignee(ctx, params)
	}

	// Query by tags
	tagOp := beads.LogicalOpAND
	if input.TagOp == "OR" {
		tagOp = beads.LogicalOpOR
	}

	tasks, err := m.tracker.GetByTags(ctx, input.Tags, tagOp)
	if err != nil {
		return json.Marshal(ListTasksOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(ListTasksOutput{
		Success: true,
		Tasks:   tasks,
	})
}

// GetByStatus handles the get_by_status tool request
func (m *TaskTrackerMCP) GetByStatus(ctx context.Context, params json.RawMessage) ([]byte, error) {
	type Input struct {
		Status string `json:"status"`
	}

	var input Input
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	tasks, err := m.tracker.GetByStatus(ctx, beads.TaskStatus(input.Status))
	if err != nil {
		return json.Marshal(ListTasksOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(ListTasksOutput{
		Success: true,
		Tasks:   tasks,
	})
}

// GetByAssignee handles the get_by_assignee tool request
func (m *TaskTrackerMCP) GetByAssignee(ctx context.Context, params json.RawMessage) ([]byte, error) {
	type Input struct {
		Assignee string `json:"assignee"`
	}

	var input Input
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	tasks, err := m.tracker.GetByAssignee(ctx, input.Assignee)
	if err != nil {
		return json.Marshal(ListTasksOutput{
			Success: false,
			Error:   err.Error(),
		})
	}

	return json.Marshal(ListTasksOutput{
		Success: true,
		Tasks:   tasks,
	})
}
