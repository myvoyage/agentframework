// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"AgentFramework/pkg/errors"
)

// WorkflowVersion represents a version of a workflow
type WorkflowVersion struct {
	ID          string    `json:"id"`
	WorkflowID  string    `json:"workflow_id"`
	Version     int       `json:"version"`
	Definition  string    `json:"definition"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// WorkflowInfo contains information about a workflow
type WorkflowInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WorkflowManager manages workflows
type WorkflowManager struct {
	workflows      map[string]*WorkflowInfo
	versions       map[string][]*WorkflowVersion // workflowID -> []WorkflowVersion
	executionStore WorkflowExecutionStore
	skillLibrary   SkillLibrary
	modelFactory   ModelFactory
	mu             sync.RWMutex
}

// NewWorkflowManager creates a new WorkflowManager
func NewWorkflowManager(skillLibrary SkillLibrary, modelFactory ModelFactory) *WorkflowManager {
	return &WorkflowManager{
		workflows:      make(map[string]*WorkflowInfo),
		versions:       make(map[string][]*WorkflowVersion),
		executionStore: NewInMemoryWorkflowExecutionStore(),
		skillLibrary:   skillLibrary,
		modelFactory:   modelFactory,
	}
}

// Init initializes the workflow manager
func (wm *WorkflowManager) Init(ctx context.Context) {
	// Initialize any resources needed for workflow management
}

// SetSkillLibrary sets the skill library for workflow execution
func (wm *WorkflowManager) SetSkillLibrary(skillLibrary SkillLibrary) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.skillLibrary = skillLibrary
}

// SetModelFactory sets the model factory for workflow execution
func (wm *WorkflowManager) SetModelFactory(modelFactory ModelFactory) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.modelFactory = modelFactory
}

// CreateWorkflow creates a new workflow
func (wm *WorkflowManager) CreateWorkflow(ctx context.Context, name string, description string, definition ...string) (string, error) {
	id := uuid.New().String()
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Use default empty JSON object if no definition provided
	workflowDefinition := "{}"
	if len(definition) > 0 && definition[0] != "" {
		workflowDefinition = definition[0]
	}

	workflow := &WorkflowInfo{
		ID:          id,
		Name:        name,
		Description: description,
		Definition:  workflowDefinition,
		CreatedAt:   getCurrentTime(),
		UpdatedAt:   getCurrentTime(),
	}

	wm.workflows[id] = workflow

	// Create initial version
	initialVersion := &WorkflowVersion{
		ID:          uuid.New().String(),
		WorkflowID:  id,
		Version:     1,
		Definition:  workflowDefinition,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	wm.versions[id] = []*WorkflowVersion{initialVersion}

	return id, nil
}

// GetWorkflows returns all workflows
func (wm *WorkflowManager) GetWorkflows(ctx context.Context) ([]*WorkflowInfo, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workflows := make([]*WorkflowInfo, 0, len(wm.workflows))
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

// GetWorkflow returns a workflow by ID
func (wm *WorkflowManager) GetWorkflow(ctx context.Context, id string) (*WorkflowInfo, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workflow, exists := wm.workflows[id]
	if !exists {
		return nil, ErrWorkflowNotFound
	}

	return workflow, nil
}

// UpdateWorkflow updates a workflow
func (wm *WorkflowManager) UpdateWorkflow(ctx context.Context, id string, name string, description string, definition string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow, exists := wm.workflows[id]
	if !exists {
		return ErrWorkflowNotFound
	}

	// Save current state as new version
	currentVersions := wm.versions[id]
	newVersion := &WorkflowVersion{
		ID:          uuid.New().String(),
		WorkflowID:  id,
		Version:     len(currentVersions) + 1,
		Definition:  definition,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	wm.versions[id] = append(currentVersions, newVersion)

	// Update workflow
	workflow.Name = name
	workflow.Description = description
	workflow.Definition = definition
	workflow.UpdatedAt = getCurrentTime()

	return nil
}

// DeleteWorkflow deletes a workflow
func (wm *WorkflowManager) DeleteWorkflow(ctx context.Context, id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	_, exists := wm.workflows[id]
	if !exists {
		return ErrWorkflowNotFound
	}

	delete(wm.workflows, id)
	delete(wm.versions, id)
	return nil
}

// GetWorkflowVersions returns all versions of a workflow
func (wm *WorkflowManager) GetWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	versions, exists := wm.versions[workflowID]
	if !exists {
		return nil, ErrWorkflowNotFound
	}

	return versions, nil
}

// GetWorkflowVersion returns a specific version of a workflow
func (wm *WorkflowManager) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	versions, exists := wm.versions[workflowID]
	if !exists {
		return nil, ErrWorkflowNotFound
	}

	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, fmt.Errorf("version %d not found for workflow %s", version, workflowID)
}

// RestoreWorkflowVersion restores a workflow to a specific version
func (wm *WorkflowManager) RestoreWorkflowVersion(ctx context.Context, workflowID string, version int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return ErrWorkflowNotFound
	}

	versions, exists := wm.versions[workflowID]
	if !exists {
		return ErrWorkflowNotFound
	}

	var versionToRestore *WorkflowVersion
	for _, v := range versions {
		if v.Version == version {
			versionToRestore = v
			break
		}
	}

	if versionToRestore == nil {
		return fmt.Errorf("version %d not found for workflow %s", version, workflowID)
	}

	// Update workflow to the restored version
	workflow.Name = versionToRestore.Name
	workflow.Description = versionToRestore.Description
	workflow.Definition = versionToRestore.Definition
	workflow.UpdatedAt = getCurrentTime()

	// Create a new version with the restored state
	newVersion := &WorkflowVersion{
		ID:          uuid.New().String(),
		WorkflowID:  workflowID,
		Version:     len(versions) + 1,
		Definition:  versionToRestore.Definition,
		Name:        versionToRestore.Name,
		Description: versionToRestore.Description,
		CreatedAt:   time.Now(),
	}

	versions = append(versions, newVersion)
	wm.versions[workflowID] = versions

	return nil
}

// ExecuteWorkflow executes a workflow
func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context, id string, input string) (string, error) {
	wm.mu.RLock()
	workflowInfo, exists := wm.workflows[id]
	skillLibrary := wm.skillLibrary
	modelFactory := wm.modelFactory
	executionStore := wm.executionStore
	wm.mu.RUnlock()

	if !exists {
		return "", ErrWorkflowNotFound
	}

	// Parse workflow definition
	wfDef, err := ParseWorkflowDefinition(workflowInfo.Definition)
	if err != nil {
		return "", fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// Create workflow instance
	workflow, err := CreateWorkflowFromDefinition(wfDef, skillLibrary, modelFactory)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create execution result
	executionID := uuid.New().String()
	result := &WorkflowExecutionResult{
		WorkflowID:  id,
		ExecutionID: executionID,
		Status:      WorkflowStatusRunning,
		Input:       input,
		StartTime:   time.Now(),
		NodeResults: make([]NodeExecutionResult, 0),
	}

	// Save initial execution state
	if err := executionStore.SaveExecutionResult(ctx, result); err != nil {
		return "", fmt.Errorf("failed to save execution result: %w", err)
	}

	// Create callback handler for tracking node execution
	callbackHandler := &executionCallbackHandler{
		executionResult: result,
	}

	// Inject callback handler into context
	ctx = WithWorkflowCallbacks(ctx, callbackHandler)

	// Execute workflow
	resp, err := workflow.Run(ctx, input)
	if err != nil {
		// Update execution result with failure status
		result.Status = WorkflowStatusFailed
		result.Error = err.Error()
		result.EndTime = time.Now()
		if saveErr := executionStore.SaveExecutionResult(ctx, result); err != nil {
			return "", fmt.Errorf("failed to save execution result: %w", saveErr)
		}
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}

	// Update execution result with success status
	result.Status = WorkflowStatusCompleted
	result.Output = resp.Content
	result.EndTime = time.Now()
	if err := executionStore.SaveExecutionResult(ctx, result); err != nil {
		return "", fmt.Errorf("failed to save execution result: %w", err)
	}

	return resp.Content, nil
}

// executionCallbackHandler implements WorkflowCallbackHandler to track node execution
type executionCallbackHandler struct {
	executionResult *WorkflowExecutionResult
}

// OnWorkflowStart is called when a workflow starts executing
func (h *executionCallbackHandler) OnWorkflowStart(ctx context.Context, workflowID string, input string) {
	// Already handled in ExecuteWorkflow
}

// OnWorkflowEnd is called when a workflow finishes executing
func (h *executionCallbackHandler) OnWorkflowEnd(ctx context.Context, workflowID string, output string, status WorkflowExecutionStatus) {
	// Already handled in ExecuteWorkflow
}

// OnNodeStart is called when a node starts executing
func (h *executionCallbackHandler) OnNodeStart(ctx context.Context, nodeID string, input string) {
	// Create node execution result
	nodeResult := NodeExecutionResult{
		NodeID:     nodeID,
		Status:     NodeStatusRunning,
		Input:      input,
		StartTime:  time.Now(),
		RetryCount: 0,
	}

	// Add to execution result
	h.executionResult.NodeResults = append(h.executionResult.NodeResults, nodeResult)
}

// OnNodeEnd is called when a node finishes executing
func (h *executionCallbackHandler) OnNodeEnd(ctx context.Context, nodeID string, output string) {
	// Find the node result and update it
	for i, nodeResult := range h.executionResult.NodeResults {
		if nodeResult.NodeID == nodeID && nodeResult.Status == NodeStatusRunning {
			h.executionResult.NodeResults[i].Status = NodeStatusCompleted
			h.executionResult.NodeResults[i].Output = output
			h.executionResult.NodeResults[i].EndTime = time.Now()
			break
		}
	}
}

// GetWorkflowExecutionResult gets the execution result of a workflow
func (wm *WorkflowManager) GetWorkflowExecutionResult(ctx context.Context, executionID string) (*WorkflowExecutionResult, error) {
	return wm.executionStore.GetExecutionResult(ctx, executionID)
}

// GetWorkflowExecutionResults gets all execution results for a workflow
func (wm *WorkflowManager) GetWorkflowExecutionResults(ctx context.Context, workflowID string) ([]*WorkflowExecutionResult, error) {
	return wm.executionStore.GetExecutionResultsByWorkflowID(ctx, workflowID)
}

// getCurrentTime returns the current time in ISO format
func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

var (
	// ErrWorkflowNotFound indicates that a workflow was not found
	ErrWorkflowNotFound = errors.New(errors.ErrCodeWorkflowNotFound, "workflow not found")
)
