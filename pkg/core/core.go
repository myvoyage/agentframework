// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Core interfaces package
// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"time"
)

// Agent is the core interface for all AI agents
type Agent interface {
	// Name returns the agent's name
	Name() string

	// Run executes the agent with the given input and returns a response
	Run(ctx context.Context, input string, opts ...Option) (*Message, error)
}

// StreamableAgent extends Agent to support streaming responses
type StreamableAgent interface {
	Agent
	Stream(ctx context.Context, input string, opts ...Option) (*StreamReader[*Message], error)
}

// ChatModel is the core interface for AI language models
type ChatModel interface {
	// Generate generates a response from the model
	Generate(ctx context.Context, input []*Message, opts ...Option) (*Message, error)

	// Stream generates a streaming response from the model
	Stream(ctx context.Context, input []*Message, opts ...Option) (*StreamReader[*Message], error)
}

// Message represents a message in a conversation
type Message struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

// Option is the interface for configuring agent and model behavior
type Option interface {
	// Apply applies the option to the configuration
	Apply(cfg interface{}) error
}

// StreamReader is the interface for reading streaming responses
type StreamReader[T any] interface {
	// Next returns the next item from the stream
	Next() (T, error)

	// Close closes the stream
	Close() error

	// Err returns any error that occurred during streaming
	Err() error
}

// Tool is the interface for tools that agents can use
type Tool interface {
	// Info returns information about the tool
	Info(ctx context.Context) (*ToolInfo, error)

	// Invoke invokes the tool with the given input
	Invoke(ctx context.Context, input string) (string, error)
}

// ToolInfo contains information about a tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]*Parameter   `json:"parameters"`
	Returns     map[string]*Parameter   `json:"returns"`
}

// Parameter defines a parameter for a tool
type Parameter struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default,omitempty"`
}

// Skill is the interface for skills that agents can use
type Skill interface {
	// Info returns information about the skill
	Info(ctx context.Context) (*SkillInfo, error)

	// Invoke invokes the skill with the given input
	Invoke(ctx context.Context, input string) (string, error)

	// IsEnabled checks if the skill is enabled
	IsEnabled(ctx context.Context) bool

	// GetMetadata returns the skill's metadata
	GetMetadata(ctx context.Context) *SkillMetadata
}

// SkillInfo contains information about a skill
type SkillInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema *Schema                `json:"input_schema"`
	OutputSchema *Schema               `json:"output_schema"`
}

// SkillMetadata contains metadata about a skill
type SkillMetadata struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	Dependencies []string               `json:"dependencies"`
	License      string                 `json:"license"`
	Homepage     string                 `json:"homepage"`
	Repository   string                 `json:"repository"`
	Keywords     []string               `json:"keywords"`
	Config       map[string]interface{}  `json:"config"`
}

// Schema defines the JSON schema for a skill's input or output
type Schema struct {
	Type        string                 `json:"type"`
	Properties  map[string]*Schema      `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// Workflow is the interface for workflows
type Workflow interface {
	// Run runs the workflow
	Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// GetStatus returns the workflow status
	GetStatus(ctx context.Context) string

	// Cancel cancels the workflow
	Cancel(ctx context.Context) error
}