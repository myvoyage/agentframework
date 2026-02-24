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
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// SkillAgent wraps a Skill as an Agent
// This allows skills to be used in workflows and other contexts that expect an Agent
type SkillAgent struct {
	name   string
	skill  *Skill
	thread *Thread
}

// NewSkillAgent creates a new SkillAgent that wraps the given skill
func NewSkillAgent(skill *Skill) (*SkillAgent, error) {
	if skill == nil {
		return nil, fmt.Errorf("skill cannot be nil")
	}

	// Get skill info to use as agent name
	ctx := context.Background()
	info, err := skill.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill info: %w", err)
	}

	return &SkillAgent{
		name:   info.Name,
		skill:  skill,
		thread: &Thread{ID: info.Name},
	}, nil
}

// NewSkillAgentWithName creates a new SkillAgent with a custom name
func NewSkillAgentWithName(name string, skill *Skill) (*SkillAgent, error) {
	if skill == nil {
		return nil, fmt.Errorf("skill cannot be nil")
	}

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	return &SkillAgent{
		name:   name,
		skill:  skill,
		thread: &Thread{ID: name},
	}, nil
}

// Name returns the agent's name
func (a *SkillAgent) Name() string {
	return a.name
}

// Run executes the skill with the given input
// The input is passed directly to the skill's Invoke method
// The result is wrapped in a schema.Message
func (a *SkillAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Check if skill is enabled
	if !a.skill.IsEnabled(ctx) {
		return nil, fmt.Errorf("skill %s is disabled", a.name)
	}

	// Create user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	// Add to thread history
	a.thread.Messages = append(a.thread.Messages, userMsg)

	// Invoke the skill
	result, err := a.skill.Invoke(ctx, input)
	if err != nil {
		// Create error message
		errorMsg := &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("Error executing skill: %v", err),
		}
		a.thread.Messages = append(a.thread.Messages, errorMsg)
		return errorMsg, err
	}

	// Convert result map to JSON string for Content
	resultJSON, err := json.Marshal(result)
	if err != nil {
		// Create error message
		errorMsg := &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("Error serializing result: %v", err),
		}
		a.thread.Messages = append(a.thread.Messages, errorMsg)
		return errorMsg, err
	}

	// Create response message
	responseMsg := &schema.Message{
		Role:    schema.Assistant,
		Content: string(resultJSON),
	}

	// Add to thread history
	a.thread.Messages = append(a.thread.Messages, responseMsg)

	return responseMsg, nil
}

// UseThread sets the thread for this agent
func (a *SkillAgent) UseThread(thread *Thread) {
	a.thread = thread
}

// GetThread returns the current thread
func (a *SkillAgent) GetThread() *Thread {
	return a.thread
}

// GetSkill returns the underlying skill
func (a *SkillAgent) GetSkill() *Skill {
	return a.skill
}

// GetSkillInfo returns the skill's info as schema.ToolInfo
func (a *SkillAgent) GetSkillInfo(ctx context.Context) (*schema.ToolInfo, error) {
	info, err := a.skill.Info(ctx)
	if err != nil {
		return nil, err
	}

	// Convert SkillInfo to schema.ToolInfo
	// Note: We can't set the unexported params field directly
	// The calling code should handle parameter validation separately
	toolInfo := &schema.ToolInfo{
		Name: info.Name,
		Desc: info.Description,
	}

	return toolInfo, nil
}

// GetSkillMetadata returns the skill's metadata
func (a *SkillAgent) GetSkillMetadata(ctx context.Context) SkillMetadata {
	// Skill has Metadata field directly, convert to SkillMetadata
	return SkillMetadata{
		ID:          a.skill.ID,
		Name:        a.skill.Name,
		Description: a.skill.Description,
		Version:     a.skill.Version,
		// Convert Metadata map to appropriate fields if needed
		// For now, use the Metadata map to populate fields if available
	}
}

// SkillAgentWithValidation wraps a SkillAgent with input validation
type SkillAgentWithValidation struct {
	*SkillAgent
	validator InputValidator
}

// InputValidator validates input before passing to the skill
type InputValidator func(ctx context.Context, input string) error

// NewSkillAgentWithValidation creates a SkillAgent with input validation
func NewSkillAgentWithValidation(skill *Skill, validator InputValidator) (*SkillAgentWithValidation, error) {
	agent, err := NewSkillAgent(skill)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return nil, fmt.Errorf("validator cannot be nil")
	}

	return &SkillAgentWithValidation{
		SkillAgent: agent,
		validator:  validator,
	}, nil
}

// Run executes the skill with validation
func (a *SkillAgentWithValidation) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Validate input
	if err := a.validator(ctx, input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Execute skill
	return a.SkillAgent.Run(ctx, input, opts...)
}

// SkillAgentWithRetry wraps a SkillAgent with retry logic
type SkillAgentWithRetry struct {
	*SkillAgent
	maxRetries int
	retryDelay int // in milliseconds
}

// NewSkillAgentWithRetry creates a SkillAgent with retry logic
func NewSkillAgentWithRetry(skill *Skill, maxRetries int) (*SkillAgentWithRetry, error) {
	agent, err := NewSkillAgent(skill)
	if err != nil {
		return nil, err
	}

	if maxRetries < 0 {
		return nil, fmt.Errorf("maxRetries must be non-negative")
	}

	return &SkillAgentWithRetry{
		SkillAgent: agent,
		maxRetries: maxRetries,
		retryDelay: 1000, // 1 second default
	}, nil
}

// Run executes the skill with retry logic
func (a *SkillAgentWithRetry) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	var lastErr error

	for i := 0; i <= a.maxRetries; i++ {
		result, err := a.SkillAgent.Run(ctx, input, opts...)
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
			case <-time.After(time.Duration(a.retryDelay) * time.Millisecond):
				// Continue to next retry
			}
		}
	}

	return nil, fmt.Errorf("skill execution failed after %d retries: %w", a.maxRetries, lastErr)
}

// SkillAgentWithCache wraps a SkillAgent with caching
type SkillAgentWithCache struct {
	*SkillAgent
	cache map[string]*schema.Message
}

// NewSkillAgentWithCache creates a SkillAgent with caching
func NewSkillAgentWithCache(skill *Skill) (*SkillAgentWithCache, error) {
	agent, err := NewSkillAgent(skill)
	if err != nil {
		return nil, err
	}

	return &SkillAgentWithCache{
		SkillAgent: agent,
		cache:      make(map[string]*schema.Message),
	}, nil
}

// Run executes the skill with caching
func (a *SkillAgentWithCache) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Check cache
	if cached, ok := a.cache[input]; ok {
		return cached, nil
	}

	// Execute skill
	result, err := a.SkillAgent.Run(ctx, input, opts...)
	if err != nil {
		return nil, err
	}

	// Store in cache
	a.cache[input] = result

	return result, nil
}

// ClearCache clears the cache
func (a *SkillAgentWithCache) ClearCache() {
	a.cache = make(map[string]*schema.Message)
}

// SkillAgentBuilder provides a fluent interface for building SkillAgents
type SkillAgentBuilder struct {
	skill      *Skill
	name       string
	validator  InputValidator
	maxRetries int
	useCache   bool
	err        error
}

// NewSkillAgentBuilder creates a new SkillAgentBuilder
func NewSkillAgentBuilder(skill *Skill) *SkillAgentBuilder {
	return &SkillAgentBuilder{
		skill:      skill,
		maxRetries: -1, // -1 means no retry
	}
}

// WithName sets a custom name for the agent
func (b *SkillAgentBuilder) WithName(name string) *SkillAgentBuilder {
	b.name = name
	return b
}

// WithValidation adds input validation
func (b *SkillAgentBuilder) WithValidation(validator InputValidator) *SkillAgentBuilder {
	b.validator = validator
	return b
}

// WithRetry adds retry logic
func (b *SkillAgentBuilder) WithRetry(maxRetries int) *SkillAgentBuilder {
	b.maxRetries = maxRetries
	return b
}

// WithCache enables caching
func (b *SkillAgentBuilder) WithCache() *SkillAgentBuilder {
	b.useCache = true
	return b
}

// Build creates the SkillAgent with the configured options
func (b *SkillAgentBuilder) Build() (Agent, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.skill == nil {
		return nil, fmt.Errorf("skill cannot be nil")
	}

	// Create base agent
	var baseAgent *SkillAgent
	var err error

	if b.name != "" {
		baseAgent, err = NewSkillAgentWithName(b.name, b.skill)
	} else {
		baseAgent, err = NewSkillAgent(b.skill)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with features
	var agent Agent = baseAgent

	// Add validation
	if b.validator != nil {
		validatedAgent, err := NewSkillAgentWithValidation(b.skill, b.validator)
		if err != nil {
			return nil, err
		}
		agent = validatedAgent
	}

	// Add retry
	if b.maxRetries >= 0 {
		retryAgent, err := NewSkillAgentWithRetry(b.skill, b.maxRetries)
		if err != nil {
			return nil, err
		}
		agent = retryAgent
	}

	// Add cache
	if b.useCache {
		cacheAgent, err := NewSkillAgentWithCache(b.skill)
		if err != nil {
			return nil, err
		}
		agent = cacheAgent
	}

	return agent, nil
}

// JSONInputValidator validates that input is valid JSON
func JSONInputValidator(ctx context.Context, input string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(input), &js); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}
	return nil
}

// NonEmptyInputValidator validates that input is not empty
func NonEmptyInputValidator(ctx context.Context, input string) error {
	if input == "" {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

// MaxLengthInputValidator creates a validator that checks input length
func MaxLengthInputValidator(maxLength int) InputValidator {
	return func(ctx context.Context, input string) error {
		if len(input) > maxLength {
			return fmt.Errorf("input length %d exceeds maximum %d", len(input), maxLength)
		}
		return nil
	}
}
