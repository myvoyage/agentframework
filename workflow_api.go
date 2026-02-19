// Agent Framework - Workflow Management API
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// CreateWorkflow creates a new workflow
func (a *App) CreateWorkflow(name string, description string, definition ...string) (string, error) {
	return a.core.GetWorkflowManager().CreateWorkflow(a.ctx, name, description, definition...)
}

// GetWorkflows returns all workflows
func (a *App) GetWorkflows() ([]*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflows(a.ctx)
}

// GetWorkflow returns a workflow by ID
func (a *App) GetWorkflow(id string) (*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflow(a.ctx, id)
}

// UpdateWorkflow updates a workflow
func (a *App) UpdateWorkflow(id string, name string, description string, definition string) error {
	return a.core.GetWorkflowManager().UpdateWorkflow(a.ctx, id, name, description, definition)
}

// DeleteWorkflow deletes a workflow
func (a *App) DeleteWorkflow(id string) error {
	return a.core.GetWorkflowManager().DeleteWorkflow(a.ctx, id)
}

// ExecuteWorkflow executes a workflow
func (a *App) ExecuteWorkflow(id string, input string) (string, error) {
	return a.core.GetWorkflowManager().ExecuteWorkflow(a.ctx, id, input)
}

// GetWorkflowVersions returns all versions of a workflow
func (a *App) GetWorkflowVersions(workflowID string) ([]*agent.WorkflowVersion, error) {
	return a.core.GetWorkflowManager().GetWorkflowVersions(a.ctx, workflowID)
}

// GetWorkflowVersion returns a specific version of a workflow
func (a *App) GetWorkflowVersion(workflowID string, version int) (*agent.WorkflowVersion, error) {
	return a.core.GetWorkflowManager().GetWorkflowVersion(a.ctx, workflowID, version)
}

// RestoreWorkflowVersion restores a workflow to a specific version
func (a *App) RestoreWorkflowVersion(workflowID string, version int) error {
	return a.core.GetWorkflowManager().RestoreWorkflowVersion(a.ctx, workflowID, version)
}

// GetWorkflowExecutionResult gets the execution result of a workflow
func (a *App) GetWorkflowExecutionResult(executionID string) (*agent.WorkflowExecutionResult, error) {
	return a.core.GetWorkflowManager().GetWorkflowExecutionResult(a.ctx, executionID)
}

// GetWorkflowExecutionResults gets all execution results for a workflow
func (a *App) GetWorkflowExecutionResults(workflowID string) ([]*agent.WorkflowExecutionResult, error) {
	return a.core.GetWorkflowManager().GetWorkflowExecutionResults(a.ctx, workflowID)
}
