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

package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// HITLStatus represents the status of a HITL request.
type HITLStatus string

const (
	HITLStatusPending   HITLStatus = "pending"
	HITLStatusApproved  HITLStatus = "approved"
	HITLStatusRejected  HITLStatus = "rejected"
	HITLStatusExpired   HITLStatus = "expired"
	HITLStatusCompleted HITLStatus = "completed"
)

// HITLRequest represents a Human-in-the-Loop request.
type HITLRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Instruction string            `json:"instruction"`
	Input       string            `json:"input"`
	Status      HITLStatus        `json:"status"`
	Response    string            `json:"response,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	ExpiresAt   string            `json:"expires_at,omitempty"`
}

// HumanNode implements a Human-in-the-Loop workflow node with enhanced features.
type HumanNode struct {
	name             string
	instruction      string
	formSchema       map[string]any    // JSON schema for form-based interaction
	supportedActions []string          // Supported actions (approve, reject, etc.)
	timeout          int64             // Timeout in seconds
	notifications    []string          // Notification methods (email, webhook, etc.)
	autoProceed      bool              // Auto-proceed if no response within timeout
	defaultAction    string            // Default action if auto-proceed
	metadata         map[string]string // Additional metadata
}

// HumanNodeOption defines options for configuring a HumanNode.
type HumanNodeOption func(*HumanNode)

// WithFormSchema sets the JSON schema for form-based interaction.
func WithFormSchema(schema map[string]any) HumanNodeOption {
	return func(node *HumanNode) {
		node.formSchema = schema
	}
}

// WithSupportedActions sets the supported actions for the HumanNode.
func WithSupportedActions(actions ...string) HumanNodeOption {
	return func(node *HumanNode) {
		node.supportedActions = actions
	}
}

// WithTimeout sets the timeout for the HITL request in seconds.
func WithTimeout(timeout int64) HumanNodeOption {
	return func(node *HumanNode) {
		node.timeout = timeout
	}
}

// WithNotifications sets the notification methods for the HITL request.
func WithNotifications(notifications ...string) HumanNodeOption {
	return func(node *HumanNode) {
		node.notifications = notifications
	}
}

// WithAutoProceed enables auto-proceed if no response is received within the timeout.
func WithAutoProceed(defaultAction string) HumanNodeOption {
	return func(node *HumanNode) {
		node.autoProceed = true
		node.defaultAction = defaultAction
	}
}

// WithMetadata sets additional metadata for the HumanNode.
func WithMetadata(metadata map[string]string) HumanNodeOption {
	return func(node *HumanNode) {
		node.metadata = metadata
	}
}

// NewHumanNode creates a new HumanNode with the given name and instruction.
func NewHumanNode(name, instruction string, options ...HumanNodeOption) *HumanNode {
	node := &HumanNode{
		name:             name,
		instruction:      instruction,
		supportedActions: []string{"approve", "reject"}, // Default supported actions
		metadata:         make(map[string]string),
	}

	// Apply options
	for _, option := range options {
		option(node)
	}

	return node
}

func (n *HumanNode) Name() string {
	return n.name
}

// HITLContextKey defines context keys for HITL operations.
type HITLContextKey string

const (
	// ResumeInputKey is the context key for resume input.
	ResumeInputKey HITLContextKey = "resume_input"
	// ActionKey is the context key for the action taken.
	ActionKey HITLContextKey = "action"
	// RequestIDKey is the context key for the HITL request ID.
	RequestIDKey HITLContextKey = "request_id"
)

// HITLState represents the state of a HITL request.
type HITLState struct {
	Request *HITLRequest   `json:"request"`
	Form    map[string]any `json:"form,omitempty"`
	Actions []string       `json:"actions,omitempty"`
}

func (n *HumanNode) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Check context for resume input and action
	if resumeInput, ok := ctx.Value(ResumeInputKey).(string); ok {
		// Resuming! The user provided this input.
		return schema.UserMessage(resumeInput), nil
	}

	// Create a simple system message with HITL instructions
	msg := &schema.Message{
		Role: schema.System,
		Content: fmt.Sprintf("HITL Request: %s\n\nInstruction: %s\n\nInput: %s\n\nPlease provide a response or take one of the following actions: %s",
			n.name, n.instruction, input, strings.Join(n.supportedActions, ", ")),
	}

	// Return the message along with ErrSuspended to indicate we need human input
	return msg, ErrSuspended
}

func (n *HumanNode) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// Stream implementation for HITL
	if resumeInput, ok := ctx.Value(ResumeInputKey).(string); ok {
		sr, pw := schema.Pipe[*schema.Message](1)
		go func() {
			// Create response message
			resp := schema.UserMessage(resumeInput)
			pw.Send(resp, nil)
			pw.Close()
		}()
		return sr, nil
	}

	// Create a simple stream response for HITL request
	sr, pw := schema.Pipe[*schema.Message](1)
	go func() {
		// Create a simple system message with HITL instructions
		msg := &schema.Message{
			Role: schema.System,
			Content: fmt.Sprintf("HITL Request: %s\n\nInstruction: %s\n\nInput: %s\n\nPlease provide a response or take one of the following actions: %s",
				n.name, n.instruction, input, strings.Join(n.supportedActions, ", ")),
		}

		pw.Send(msg, nil)
		pw.Close()
	}()

	return sr, ErrSuspended
}

func (n *HumanNode) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: n.name, Desc: n.instruction}, nil
}

// GetFormSchema returns the form schema for the HumanNode.
func (n *HumanNode) GetFormSchema() map[string]any {
	return n.formSchema
}

// GetSupportedActions returns the supported actions for the HumanNode.
func (n *HumanNode) GetSupportedActions() []string {
	return n.supportedActions
}

// GetTimeout returns the timeout for the HumanNode in seconds.
func (n *HumanNode) GetTimeout() int64 {
	return n.timeout
}

// GetNotifications returns the notification methods for the HumanNode.
func (n *HumanNode) GetNotifications() []string {
	return n.notifications
}

// IsAutoProceed returns whether auto-proceed is enabled for the HumanNode.
func (n *HumanNode) IsAutoProceed() bool {
	return n.autoProceed
}

// GetDefaultAction returns the default action for auto-proceed.
func (n *HumanNode) GetDefaultAction() string {
	return n.defaultAction
}

// GetMetadata returns the metadata for the HumanNode.
func (n *HumanNode) GetMetadata() map[string]string {
	return n.metadata
}

// WithFormSchema updates the form schema for the HumanNode.
func (n *HumanNode) WithFormSchema(schema map[string]any) *HumanNode {
	n.formSchema = schema
	return n
}

// Resume implements the Workflow interface for HumanNode
func (n *HumanNode) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	// Create a message with the resume input
	return schema.UserMessage(input), nil
}

// WithSupportedActions updates the supported actions for the HumanNode.
func (n *HumanNode) WithSupportedActions(actions ...string) *HumanNode {
	n.supportedActions = actions
	return n
}

// WithTimeout updates the timeout for the HumanNode in seconds.
func (n *HumanNode) WithTimeout(timeout int64) *HumanNode {
	n.timeout = timeout
	return n
}

// WithNotifications updates the notification methods for the HumanNode.
func (n *HumanNode) WithNotifications(notifications ...string) *HumanNode {
	n.notifications = notifications
	return n
}

// WithAutoProceed updates the auto-proceed settings for the HumanNode.
func (n *HumanNode) WithAutoProceed(defaultAction string) *HumanNode {
	n.autoProceed = true
	n.defaultAction = defaultAction
	return n
}

// WithMetadata updates the metadata for the HumanNode.
func (n *HumanNode) WithMetadata(metadata map[string]string) *HumanNode {
	n.metadata = metadata
	return n
}
