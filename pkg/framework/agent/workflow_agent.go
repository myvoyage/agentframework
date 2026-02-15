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

package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// WorkflowAgent wraps a Workflow as an Agent
// This allows workflows to be used in contexts that expect an Agent
type WorkflowAgent struct {
	name     string
	workflow Workflow
	thread   *Thread
}

// NewWorkflowAgent creates a new WorkflowAgent that wraps the given workflow
func NewWorkflowAgent(workflow Workflow) (*WorkflowAgent, error) {
	if workflow == nil {
		return nil, fmt.Errorf("workflow cannot be nil")
	}

	return &WorkflowAgent{
		name:     workflow.Name(),
		workflow: workflow,
		thread:   &Thread{ID: workflow.Name()},
	}, nil
}

// NewWorkflowAgentWithName creates a new WorkflowAgent with a custom name
func NewWorkflowAgentWithName(name string, workflow Workflow) (*WorkflowAgent, error) {
	if workflow == nil {
		return nil, fmt.Errorf("workflow cannot be nil")
	}

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	return &WorkflowAgent{
		name:     name,
		workflow: workflow,
		thread:   &Thread{ID: name},
	}, nil
}

// Name returns the agent's name
func (a *WorkflowAgent) Name() string {
	return a.name
}

// Run executes the workflow with the given input
// The input is passed directly to the workflow's Run method
func (a *WorkflowAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Create user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	// Add to thread history
	a.thread.Messages = append(a.thread.Messages, userMsg)

	// Execute the workflow
	result, err := a.workflow.Run(ctx, input, opts...)
	if err != nil {
		// Create error message
		errorMsg := &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("Error executing workflow: %v", err),
		}
		a.thread.Messages = append(a.thread.Messages, errorMsg)
		return errorMsg, err
	}

	// Add result to thread history
	a.thread.Messages = append(a.thread.Messages, result)

	return result, nil
}

// UseThread sets the thread for this agent
func (a *WorkflowAgent) UseThread(thread *Thread) {
	a.thread = thread
}

// GetThread returns the current thread
func (a *WorkflowAgent) GetThread() *Thread {
	return a.thread
}

// GetWorkflow returns the underlying workflow
func (a *WorkflowAgent) GetWorkflow() Workflow {
	return a.workflow
}

// WorkflowAgentWithRetry wraps a WorkflowAgent with retry logic
type WorkflowAgentWithRetry struct {
	*WorkflowAgent
	maxRetries int
	retryDelay int // in milliseconds
}

// NewWorkflowAgentWithRetry creates a WorkflowAgent with retry logic
func NewWorkflowAgentWithRetry(workflow Workflow, maxRetries int) (*WorkflowAgentWithRetry, error) {
	agent, err := NewWorkflowAgent(workflow)
	if err != nil {
		return nil, err
	}

	if maxRetries < 0 {
		return nil, fmt.Errorf("maxRetries must be non-negative")
	}

	return &WorkflowAgentWithRetry{
		WorkflowAgent: agent,
		maxRetries:    maxRetries,
		retryDelay:    1000, // 1 second default
	}, nil
}

// Run executes the workflow with retry logic
func (a *WorkflowAgentWithRetry) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	var lastErr error

	for i := 0; i <= a.maxRetries; i++ {
		result, err := a.WorkflowAgent.Run(ctx, input, opts...)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry if it's the last attempt
		if i < a.maxRetries {
			// Wait before retrying
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-workflowTimeAfter(a.retryDelay):
				// Continue to next retry
			}
		}
	}

	return nil, fmt.Errorf("workflow execution failed after %d retries: %w", a.maxRetries, lastErr)
}

// WorkflowAgentWithCache wraps a WorkflowAgent with caching
type WorkflowAgentWithCache struct {
	*WorkflowAgent
	cache map[string]*schema.Message
}

// NewWorkflowAgentWithCache creates a WorkflowAgent with caching
func NewWorkflowAgentWithCache(workflow Workflow) (*WorkflowAgentWithCache, error) {
	agent, err := NewWorkflowAgent(workflow)
	if err != nil {
		return nil, err
	}

	return &WorkflowAgentWithCache{
		WorkflowAgent: agent,
		cache:         make(map[string]*schema.Message),
	}, nil
}

// Run executes the workflow with caching
func (a *WorkflowAgentWithCache) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Check cache
	if cached, ok := a.cache[input]; ok {
		return cached, nil
	}

	// Execute workflow
	result, err := a.WorkflowAgent.Run(ctx, input, opts...)
	if err != nil {
		return nil, err
	}

	// Store in cache
	a.cache[input] = result

	return result, nil
}

// ClearCache clears the cache
func (a *WorkflowAgentWithCache) ClearCache() {
	a.cache = make(map[string]*schema.Message)
}

// WorkflowAgentBuilder provides a fluent interface for building WorkflowAgents
type WorkflowAgentBuilder struct {
	workflow   Workflow
	name       string
	maxRetries int
	useCache   bool
	err        error
}

// NewWorkflowAgentBuilder creates a new WorkflowAgentBuilder
func NewWorkflowAgentBuilder(workflow Workflow) *WorkflowAgentBuilder {
	return &WorkflowAgentBuilder{
		workflow:   workflow,
		maxRetries: -1, // -1 means no retry
	}
}

// WithName sets a custom name for the agent
func (b *WorkflowAgentBuilder) WithName(name string) *WorkflowAgentBuilder {
	b.name = name
	return b
}

// WithRetry adds retry logic
func (b *WorkflowAgentBuilder) WithRetry(maxRetries int) *WorkflowAgentBuilder {
	b.maxRetries = maxRetries
	return b
}

// WithCache enables caching
func (b *WorkflowAgentBuilder) WithCache() *WorkflowAgentBuilder {
	b.useCache = true
	return b
}

// Build creates the WorkflowAgent with the configured options
func (b *WorkflowAgentBuilder) Build() (Agent, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.workflow == nil {
		return nil, fmt.Errorf("workflow cannot be nil")
	}

	// Create base agent
	var baseAgent *WorkflowAgent
	var err error

	if b.name != "" {
		baseAgent, err = NewWorkflowAgentWithName(b.name, b.workflow)
	} else {
		baseAgent, err = NewWorkflowAgent(b.workflow)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with features
	var agent Agent = baseAgent

	// Add retry
	if b.maxRetries >= 0 {
		retryAgent, err := NewWorkflowAgentWithRetry(b.workflow, b.maxRetries)
		if err != nil {
			return nil, err
		}
		agent = retryAgent
	}

	// Add cache
	if b.useCache {
		cacheAgent, err := NewWorkflowAgentWithCache(b.workflow)
		if err != nil {
			return nil, err
		}
		agent = cacheAgent
	}

	return agent, nil
}

// Helper function to create a time.After channel for workflow agent
func workflowTimeAfter(ms int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		// Simple sleep implementation
		for i := 0; i < ms; i++ {
			select {
			case <-ch:
				return
			default:
				// Continue
			}
		}
		close(ch)
	}()
	return ch
}
