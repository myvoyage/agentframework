// Agent Framework - Execution Loop
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Execution Loop implements the model → tool → model cycle:
//   1. Model generates response or tool call
//   2. If tool call, execute and collect result
//   3. Continue until model produces final response
//
// Streaming: Events are emitted through event channel for observability
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// ExecutionConfig contains configuration for the execution loop
type ExecutionConfig struct {
	// MaxIterations limits the number of model-tool iterations
	MaxIterations int

	// Timeout sets the maximum execution time
	Timeout time.Duration

	// ToolTimeout sets the maximum time for a single tool execution
	ToolTimeout time.Duration

	// StreamOutput enables streaming output
	StreamOutput bool

	// EnableParallelTools enables parallel tool execution
	EnableParallelTools bool

	// EventChan is a channel for emitting execution events
	EventChan chan *ExecutionEvent
}

// ExecutionEvent represents an event during execution
type ExecutionEvent struct {
	// RunID is the unique identifier for this execution run
	RunID string

	// Type is the event type
	Type string // "started", "thinking", "tool_call", "tool_result", "completed", "error"

	// Content is the event content (message, tool name, error message, etc.)
	Content string

	// Timestamp is when the event occurred
	Timestamp time.Time

	// Metadata contains additional event data
	Metadata map[string]interface{}
}

// DefaultExecutionConfig returns the default execution configuration
func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		MaxIterations:   10,
		Timeout:         5 * time.Minute,
		ToolTimeout:     30 * time.Second,
		StreamOutput:    false,
		EnableParallelTools: false,
	}
}

// ExecutionResult contains the result of an execution
type ExecutionResult struct {
	// RunID is the unique identifier for this execution run
	RunID string

	// Response is the final response from the model
	Response *schema.Message

	// ToolCalls are the tool calls made during execution
	ToolCalls []*ToolCall

	// Iterations is the number of iterations made
	Iterations int

	// Duration is the total execution time
	Duration time.Duration

	// Error is the error if execution failed
	Error error

	// Events are all events emitted during execution
	Events []*ExecutionEvent
}

// ToolCall represents a tool call made during execution
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
	Result    string
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

// ExecutionLoop manages the model → tool → model execution cycle
type ExecutionLoop struct {
	model   model.ChatModel
	tools   map[string]Tool
	sandbox interface{} // SandboxManager - TODO: implement
	cfg     *ExecutionConfig
}

// NewExecutionLoop creates a new execution loop
func NewExecutionLoop(model model.ChatModel, tools map[string]Tool, sandbox interface{}, cfg *ExecutionConfig) *ExecutionLoop {
	if cfg == nil {
		cfg = DefaultExecutionConfig()
	}

	return &ExecutionLoop{
		model:   model,
		tools:   tools,
		sandbox: sandbox,
		cfg:     cfg,
	}
}

// Execute runs the execution loop until completion
func (el *ExecutionLoop) Execute(ctx context.Context, execCtx *Context) (*ExecutionResult, error) {
	runID := uuid.New().String()
	start := time.Now()
	result := &ExecutionResult{
		RunID:     runID,
		ToolCalls: make([]*ToolCall, 0),
		Events:    make([]*ExecutionEvent, 0),
	}

	// Emit started event
	el.emitEvent(&ExecutionEvent{
		RunID:     runID,
		Type:       "started",
		Content:    "Starting agent execution",
		Timestamp:  time.Now(),
	})

	// Apply timeout if configured
	if el.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, el.cfg.Timeout)
		defer cancel()
	}

	// Build initial messages
	messages := execCtx.GetAllMessages()

	// Add skill instructions if available
	if execCtx.HasSkills() {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: fmt.Sprintf("Available skills: %s", strings.Join(execCtx.Skills, ", ")),
		})
	}

	// Add memory results if available
	if execCtx.HasMemory() {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: execCtx.FormatMemoryResults(),
		})
	}

	// Execution loop
	for result.Iterations < el.cfg.MaxIterations {
		result.Iterations++

		// Emit thinking event
		el.emitEvent(&ExecutionEvent{
			RunID:     runID,
			Type:       "thinking",
			Content:    fmt.Sprintf("Iteration %d: thinking", result.Iterations),
			Timestamp:  time.Now(),
			Metadata:   map[string]interface{}{"iteration": result.Iterations},
		})

		// Generate response from model
		resp, err := el.model.Generate(ctx, messages)
		if err != nil {
			result.Error = fmt.Errorf("model generation failed: %w", err)

			el.emitEvent(&ExecutionEvent{
				RunID:     runID,
				Type:       "error",
				Content:    result.Error.Error(),
				Timestamp:  time.Now(),
			})

			return result, result.Error
		}

		// Check if response contains tool calls
		toolCalls := el.extractToolCalls(resp)
		if len(toolCalls) == 0 {
			// No tool calls, this is the final response
			result.Response = resp
			result.Duration = time.Since(start)

			el.emitEvent(&ExecutionEvent{
				RunID:     runID,
				Type:       "completed",
				Content:    "Agent execution completed successfully",
				Timestamp:  time.Now(),
				Metadata:   map[string]interface{}{
					"iterations": result.Iterations,
					"duration_ms": result.Duration.Milliseconds(),
				},
			})

			return result, nil
		}

		// Execute tool calls
		for _, tc := range toolCalls {
			tc.Timestamp = time.Now()

			// Emit tool call event
			el.emitEvent(&ExecutionEvent{
				RunID:     runID,
				Type:       "tool_call",
				Content:    fmt.Sprintf("Calling tool: %s", tc.Name),
				Timestamp:  time.Now(),
				Metadata:   map[string]interface{}{
					"tool_name": tc.Name,
					"tool_args": tc.Arguments,
				},
			})

			toolResult, err := el.executeTool(ctx, execCtx.Session, tc)
			tc.Result = toolResult
			tc.Duration = time.Since(tc.Timestamp)

			// Emit tool result event
			eventType := "tool_result"
			eventContent := fmt.Sprintf("Tool %s completed: %s", tc.Name, toolResult)
			if err != nil {
				eventType = "error"
				eventContent = fmt.Sprintf("Tool %s failed: %v", tc.Name, err)
			}

			el.emitEvent(&ExecutionEvent{
				RunID:     runID,
				Type:       eventType,
				Content:    eventContent,
				Timestamp:  time.Now(),
				Metadata: map[string]interface{}{
					"tool_name": tc.Name,
					"tool_result": toolResult,
					"error": err,
				},
			})

			if err != nil {
				tc.Error = err
				result.ToolCalls = append(result.ToolCalls, tc)

				// Add error as tool result
				messages = append(messages, &schema.Message{
					Role:    schema.Tool,
					Content: fmt.Sprintf("Error: %v", err),
				}, resp)

				// Continue to next iteration
				continue
			}

			result.ToolCalls = append(result.ToolCalls, tc)

			// Add tool result to messages
			messages = append(messages, &schema.Message{
				Role:    schema.Tool,
				Content: toolResult,
			}, resp)
		}

		// Check for context length limit
		if len(messages) > 100 {
			// TODO: Implement message summarization
			// For now, just truncate old messages
			messages = messages[len(messages)-50:]
		}
	}

	// Max iterations reached
	result.Response = &schema.Message{
		Role:    schema.Assistant,
		Content: "Maximum iterations reached. The task may be too complex.",
	}
	result.Duration = time.Since(start)
	return result, nil
}

// executeTool executes a single tool call
func (el *ExecutionLoop) executeTool(ctx context.Context, session *Session, tc *ToolCall) (string, error) {
	// Find the tool
	tool, ok := el.tools[tc.Name]
	if !ok {
		return "", fmt.Errorf("tool '%s' not found", tc.Name)
	}

	// Check session permissions
	// Non-main sessions should use sandbox for tool execution
	if session != nil && !session.HasFullPermissions() {
		// Log sandboxed execution
		// TODO: Integrate with SandboxManager for actual sandboxed execution
		// For now, we'll execute directly but could add sandboxing later
	}

	// Execute tool directly with timeout
	toolCtx, cancel := context.WithTimeout(ctx, el.cfg.ToolTimeout)
	defer cancel()

	return tool.Invoke(toolCtx, tc.Arguments)
}

// emitEvent emits an execution event to the event channel
func (el *ExecutionLoop) emitEvent(event *ExecutionEvent) {
	if el.cfg.EventChan != nil {
		select {
		case el.cfg.EventChan <- event:
		default:
			// Event channel full, drop event
		}
	}
}

// extractToolCalls extracts tool calls from a model response
func (el *ExecutionLoop) extractToolCalls(msg *schema.Message) []*ToolCall {
	if msg.ToolCalls == nil || len(msg.ToolCalls) == 0 {
		return nil
	}

	toolCalls := make([]*ToolCall, 0, len(msg.ToolCalls))
	for _, tc := range msg.ToolCalls {
		toolCalls = append(toolCalls, &ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return toolCalls
}

// ExecuteSimple is a convenience method for simple execution without context assembly
func (el *ExecutionLoop) ExecuteSimple(ctx context.Context, systemPrompt, input string) (*ExecutionResult, error) {
	execCtx := &Context{
		SystemPrompt: systemPrompt,
		Messages: []*schema.Message{
			{
				Role:    schema.User,
				Content: input,
			},
		},
	}

	return el.Execute(ctx, execCtx)
}

// Tool defines the interface for executable tools
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Invoke executes the tool with the given arguments
	Invoke(ctx context.Context, args string) (string, error)
}

// ToolRegistry manages available tools
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (tr *ToolRegistry) Register(tool Tool) {
	tr.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (tr *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := tr.tools[name]
	return tool, ok
}

// List returns all registered tools
func (tr *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ToMap returns tools as a map
func (tr *ToolRegistry) ToMap() map[string]Tool {
	return tr.tools
}
