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

// SPDX-License-Identifier: AGPL-3.0-or-later

package collaboration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OrchestrationType defines the orchestration type
type OrchestrationType int

const (
	OrchestrationSequential OrchestrationType = iota
	OrchestrationParallel
	OrchestrationConditional
	OrchestrationLoop
)

// Orchestrator orchestrates complex multi-agent workflows
type Orchestrator struct {
	team      *AgentTeam
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	workflows map[string]*OrchestrationWorkflow
}

// OrchestrationWorkflow defines a workflow
type OrchestrationWorkflow struct {
	ID        string
	Name      string
	Type      OrchestrationType
	Steps     []OrchestrationStep
	Variables map[string]interface{}
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrchestrationStep defines a step in the workflow
type OrchestrationStep struct {
	ID           string
	Name         string
	Type         OrchestrationType
	Task         *CollaborativeTask
	AgentName    string   // Empty for auto-selection
	Condition    string   // Conditional expression
	LoopCount    int      // For loop steps
	Dependencies []string // Step IDs this step depends on
	Timeout      time.Duration
	RetryCount   int
}

// OrchestrationResult represents the result of an orchestration
type OrchestrationResult struct {
	WorkflowID  string
	Success     bool
	Output      string
	StepResults []OrchestrationStepResult
	Variables   map[string]interface{}
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Error       error
}

// OrchestrationStepResult represents the result of a workflow step
type OrchestrationStepResult struct {
	StepID    string
	StepName  string
	Success   bool
	Output    string
	Error     error
	Duration  time.Duration
	Retries   int
	StartTime time.Time
	EndTime   time.Time
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(team *AgentTeam) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Orchestrator{
		team:      team,
		ctx:       ctx,
		cancel:    cancel,
		workflows: make(map[string]*OrchestrationWorkflow),
	}
}

// CreateWorkflow creates a new workflow
func (o *Orchestrator) CreateWorkflow(id, name string, workflowType OrchestrationType, steps []OrchestrationStep) (*OrchestrationWorkflow, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.workflows[id]; exists {
		return nil, fmt.Errorf("workflow %s already exists", id)
	}

	workflow := &OrchestrationWorkflow{
		ID:        id,
		Name:      name,
		Type:      workflowType,
		Steps:     steps,
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	o.workflows[id] = workflow

	return workflow, nil
}

// GetWorkflow gets a workflow by ID
func (o *Orchestrator) GetWorkflow(id string) (*OrchestrationWorkflow, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflow, exists := o.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return workflow, nil
}

// DeleteWorkflow deletes a workflow
func (o *Orchestrator) DeleteWorkflow(id string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.workflows[id]; !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	delete(o.workflows, id)

	return nil
}

// Execute executes a workflow
func (o *Orchestrator) Execute(ctx context.Context, workflowID string) (*OrchestrationResult, error) {
	workflow, err := o.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Execute based on workflow type
	var result *OrchestrationResult
	switch workflow.Type {
	case OrchestrationSequential:
		result, err = o.executeSequential(ctx, workflow)
	case OrchestrationParallel:
		result, err = o.executeParallel(ctx, workflow)
	case OrchestrationConditional:
		result, err = o.executeConditional(ctx, workflow)
	case OrchestrationLoop:
		result, err = o.executeLoop(ctx, workflow)
	default:
		return nil, fmt.Errorf("unknown workflow type: %d", workflow.Type)
	}

	if err != nil {
		return nil, err
	}

	result.StartTime = startTime
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	return result, nil
}

// executeSequential executes steps sequentially
func (o *Orchestrator) executeSequential(ctx context.Context, workflow *OrchestrationWorkflow) (*OrchestrationResult, error) {
	result := &OrchestrationResult{
		WorkflowID:  workflow.ID,
		StepResults: make([]OrchestrationStepResult, 0, len(workflow.Steps)),
		Variables:   make(map[string]interface{}),
	}

	for _, step := range workflow.Steps {
		stepResult, err := o.executeStep(ctx, workflow, step)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		result.StepResults = append(result.StepResults, *stepResult)

		if !stepResult.Success {
			result.Success = false
			result.Error = stepResult.Error
			return result, nil
		}

		// Update variables with step output
		result.Variables[step.ID] = stepResult.Output
	}

	result.Success = true
	result.Output = result.StepResults[len(result.StepResults)-1].Output

	return result, nil
}

// executeParallel executes steps in parallel
func (o *Orchestrator) executeParallel(ctx context.Context, workflow *OrchestrationWorkflow) (*OrchestrationResult, error) {
	result := &OrchestrationResult{
		WorkflowID:  workflow.ID,
		StepResults: make([]OrchestrationStepResult, len(workflow.Steps)),
		Variables:   make(map[string]interface{}),
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Steps))
	stepResults := make(chan OrchestrationStepResult, len(workflow.Steps))

	// Execute all steps in parallel
	for i, step := range workflow.Steps {
		wg.Add(1)
		go func(idx int, s OrchestrationStep) {
			defer wg.Done()

			stepResult, err := o.executeStep(ctx, workflow, s)
			if err != nil {
				errChan <- err
				return
			}

			stepResults <- *stepResult
		}(i, step)
	}

	// Wait for all steps to complete
	go func() {
		wg.Wait()
		close(errChan)
		close(stepResults)
	}()

	// Collect results
	success := true
	var firstError error

	for err := range errChan {
		if err != nil {
			success = false
			if firstError == nil {
				firstError = err
			}
		}
	}

	i := 0
	for stepResult := range stepResults {
		result.StepResults[i] = stepResult
		result.Variables[workflow.Steps[i].ID] = stepResult.Output
		i++
	}

	result.Success = success
	result.Error = firstError

	if success && len(result.StepResults) > 0 {
		// Combine outputs from all steps
		combinedOutput := ""
		for _, sr := range result.StepResults {
			if combinedOutput != "" {
				combinedOutput += "\n\n"
			}
			combinedOutput += sr.Output
		}
		result.Output = combinedOutput
	}

	return result, nil
}

// executeConditional executes steps with conditions
func (o *Orchestrator) executeConditional(ctx context.Context, workflow *OrchestrationWorkflow) (*OrchestrationResult, error) {
	result := &OrchestrationResult{
		WorkflowID:  workflow.ID,
		StepResults: make([]OrchestrationStepResult, 0),
		Variables:   make(map[string]interface{}),
	}

	for _, step := range workflow.Steps {
		// Check condition
		if step.Condition != "" {
			conditionMet, err := o.evaluateCondition(step.Condition, workflow.Variables)
			if err != nil {
				result.Success = false
				result.Error = err
				return result, err
			}

			if !conditionMet {
				// Skip this step
				continue
			}
		}

		stepResult, err := o.executeStep(ctx, workflow, step)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		result.StepResults = append(result.StepResults, *stepResult)

		if !stepResult.Success {
			result.Success = false
			result.Error = stepResult.Error
			return result, nil
		}

		// Update variables with step output
		result.Variables[step.ID] = stepResult.Output
	}

	result.Success = true
	if len(result.StepResults) > 0 {
		result.Output = result.StepResults[len(result.StepResults)-1].Output
	}

	return result, nil
}

// executeLoop executes steps in a loop
func (o *Orchestrator) executeLoop(ctx context.Context, workflow *OrchestrationWorkflow) (*OrchestrationResult, error) {
	result := &OrchestrationResult{
		WorkflowID:  workflow.ID,
		StepResults: make([]OrchestrationStepResult, 0),
		Variables:   make(map[string]interface{}),
	}

	for _, step := range workflow.Steps {
		loopCount := step.LoopCount
		if loopCount <= 0 {
			loopCount = 1
		}

		for i := 0; i < loopCount; i++ {
			// Create a modified step with iteration info
			modifiedStep := step
			modifiedStep.Task.Context = make(map[string]interface{})
			for k, v := range step.Task.Context {
				modifiedStep.Task.Context[k] = v
			}
			modifiedStep.Task.Context["iteration"] = i

			stepResult, err := o.executeStep(ctx, workflow, modifiedStep)
			if err != nil {
				result.Success = false
				result.Error = err
				return result, err
			}

			result.StepResults = append(result.StepResults, *stepResult)

			if !stepResult.Success {
				result.Success = false
				result.Error = stepResult.Error
				return result, nil
			}

			// Update variables with step output
			variableKey := fmt.Sprintf("%s_%d", step.ID, i)
			result.Variables[variableKey] = stepResult.Output
		}
	}

	result.Success = true
	if len(result.StepResults) > 0 {
		result.Output = result.StepResults[len(result.StepResults)-1].Output
	}

	return result, nil
}

// executeStep executes a single step
func (o *Orchestrator) executeStep(ctx context.Context, workflow *OrchestrationWorkflow, step OrchestrationStep) (*OrchestrationStepResult, error) {
	startTime := time.Now()

	// Apply timeout if specified
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Get member
	var member *TeamMember
	var err error

	if step.AgentName != "" {
		member, err = o.team.GetMember(step.AgentName)
		if err != nil {
			return nil, fmt.Errorf("failed to get member %s: %w", step.AgentName, err)
		}
	}

	// Execute task with retries
	var taskResult *TaskResult
	maxRetries := step.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if member != nil {
			taskResult, err = o.team.scheduler.Submit(stepCtx, step.Task, member)
		} else {
			taskResult, err = o.team.AssignTask(stepCtx, step.Task)
		}

		if err == nil {
			break
		}

		if attempt < maxRetries-1 {
			// Wait before retry
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	stepResult := &OrchestrationStepResult{
		StepID:    step.ID,
		StepName:  step.Name,
		Success:   err == nil && taskResult != nil && taskResult.Success,
		Output:    "",
		Error:     err,
		Duration:  duration,
		Retries:   maxRetries - 1,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if taskResult != nil {
		stepResult.Output = taskResult.Output
		if taskResult.Error != nil {
			stepResult.Error = taskResult.Error
		}
	}

	return stepResult, nil
}

// evaluateCondition evaluates a condition expression
func (o *Orchestrator) evaluateCondition(condition string, variables map[string]interface{}) (bool, error) {
	// This is a simplified implementation
	// In production, you'd want to use a proper expression evaluator

	// For now, just check if a variable exists and is truthy
	val, exists := variables[condition]
	if !exists {
		return false, nil
	}

	// Check if value is truthy
	switch v := val.(type) {
	case bool:
		return v, nil
	case int, int64, float64:
		return true, nil // Non-zero numbers are truthy
	case string:
		return v != "", nil
	default:
		return val != nil, nil
	}
}

// ListWorkflows lists all workflows
func (o *Orchestrator) ListWorkflows() []*OrchestrationWorkflow {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflows := make([]*OrchestrationWorkflow, 0, len(o.workflows))
	for _, workflow := range o.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows
}

// Shutdown shuts down the orchestrator
func (o *Orchestrator) Shutdown() {
	o.cancel()
}
