// Agent Framework - Workflow Service
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// WorkflowService handles workflow operations
type WorkflowService struct {
	app *Application
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(app *Application) *WorkflowService {
	return &WorkflowService{app: app}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, name string, description string, definition ...string) (string, error) {
	return s.app.workflowManager.CreateWorkflow(ctx, name, description, definition...)
}

// GetWorkflows returns all workflows
func (s *WorkflowService) GetWorkflows(ctx context.Context) ([]*agent.WorkflowInfo, error) {
	return s.app.workflowManager.GetWorkflows(ctx)
}

// GetWorkflow returns a workflow by ID
func (s *WorkflowService) GetWorkflow(ctx context.Context, id string) (*agent.WorkflowInfo, error) {
	return s.app.workflowManager.GetWorkflow(ctx, id)
}

// UpdateWorkflow updates a workflow
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id string, name string, description string, definition string) error {
	return s.app.workflowManager.UpdateWorkflow(ctx, id, name, description, definition)
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id string) error {
	return s.app.workflowManager.DeleteWorkflow(ctx, id)
}

// ExecuteWorkflow executes a workflow
func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, id string, input string) (string, error) {
	return s.app.workflowManager.ExecuteWorkflow(ctx, id, input)
}

// GetWorkflowVersions returns all versions of a workflow
func (s *WorkflowService) GetWorkflowVersions(ctx context.Context, workflowID string) ([]*agent.WorkflowVersion, error) {
	return s.app.workflowManager.GetWorkflowVersions(ctx, workflowID)
}

// GetWorkflowVersion returns a specific version of a workflow
func (s *WorkflowService) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*agent.WorkflowVersion, error) {
	return s.app.workflowManager.GetWorkflowVersion(ctx, workflowID, version)
}

// RestoreWorkflowVersion restores a workflow to a specific version
func (s *WorkflowService) RestoreWorkflowVersion(ctx context.Context, workflowID string, version int) error {
	return s.app.workflowManager.RestoreWorkflowVersion(ctx, workflowID, version)
}

// GetWorkflowExecutionResult gets execution result of a workflow
func (s *WorkflowService) GetWorkflowExecutionResult(ctx context.Context, executionID string) (*agent.WorkflowExecutionResult, error) {
	return s.app.workflowManager.GetWorkflowExecutionResult(ctx, executionID)
}

// GetWorkflowExecutionResults gets all execution results for a workflow
func (s *WorkflowService) GetWorkflowExecutionResults(ctx context.Context, workflowID string) ([]*agent.WorkflowExecutionResult, error) {
	return s.app.workflowManager.GetWorkflowExecutionResults(ctx, workflowID)
}

// ListWorkflowsTable prints workflows in table format
func (s *WorkflowService) ListWorkflowsTable(ctx context.Context, outputFormat string) error {
	workflows, err := s.GetWorkflows(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workflows: %w", err)
	}

	if len(workflows) == 0 {
		fmt.Println("No workflows found")
		return nil
	}

	// Print in requested format
	switch outputFormat {
	case "json":
		// JSON output
		fmt.Printf("%+v\n", workflows)
	case "table", "":
		// Table output (default)
		fmt.Println("Workflows:")
		fmt.Println("────────────────────────────────────────────────────────────")
		for _, wf := range workflows {
			fmt.Printf("ID: %s\n", wf.ID)
			fmt.Printf("  Name: %s\n", wf.Name)
			fmt.Printf("  Description: %s\n", wf.Description)
			fmt.Println("────────────────────────────────────────────────────────────")
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
