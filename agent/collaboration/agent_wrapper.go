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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// DefaultAgentWrapper wraps the standard agent.Agent interface
type DefaultAgentWrapper struct {
	agent        Agent
	capabilities []string
	model        string
	metadata     map[string]interface{}
}

// Agent is the base agent interface from the agent package
type Agent interface {
	Name() string
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
}

// StreamableAgent is an agent that supports streaming
type StreamableAgent interface {
	Agent
	Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

// NewDefaultAgentWrapper creates a new agent wrapper
func NewDefaultAgentWrapper(agent Agent, capabilities []string, model string) *DefaultAgentWrapper {
	return &DefaultAgentWrapper{
		agent:        agent,
		capabilities: capabilities,
		model:        model,
		metadata:     make(map[string]interface{}),
	}
}

// Name returns the agent name
func (w *DefaultAgentWrapper) Name() string {
	return w.agent.Name()
}

// Type returns the agent type
func (w *DefaultAgentWrapper) Type() string {
	// Determine type based on the wrapped agent
	switch w.agent.(type) {
	case StreamableAgent:
		return "streamable"
	default:
		return "standard"
	}
}

// Run executes the agent
func (w *DefaultAgentWrapper) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return w.agent.Run(ctx, input, opts...)
}

// Stream executes the agent with streaming
func (w *DefaultAgentWrapper) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if streamable, ok := w.agent.(StreamableAgent); ok {
		return streamable.Stream(ctx, input, opts...)
	}
	return nil, nil
}

// GetCapabilities returns the agent's capabilities
func (w *DefaultAgentWrapper) GetCapabilities() []string {
	return w.capabilities
}

// GetModel returns the model name
func (w *DefaultAgentWrapper) GetModel() string {
	return w.model
}

// SetMetadata sets metadata for the agent
func (w *DefaultAgentWrapper) SetMetadata(key string, value interface{}) {
	w.metadata[key] = value
}

// GetMetadata gets metadata for the agent
func (w *DefaultAgentWrapper) GetMetadata(key string) (interface{}, bool) {
	val, ok := w.metadata[key]
	return val, ok
}

// AddCapability adds a capability to the agent
func (w *DefaultAgentWrapper) AddCapability(capability string) {
	w.capabilities = append(w.capabilities, capability)
}

// RemoveCapability removes a capability from the agent
func (w *DefaultAgentWrapper) RemoveCapability(capability string) {
	for i, cap := range w.capabilities {
		if cap == capability {
			w.capabilities = append(w.capabilities[:i], w.capabilities[i+1:]...)
			break
		}
	}
}

// HasCapability checks if the agent has a specific capability
func (w *DefaultAgentWrapper) HasCapability(capability string) bool {
	for _, cap := range w.capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetCapabilitiesCount returns the number of capabilities
func (w *DefaultAgentWrapper) GetCapabilitiesCount() int {
	return len(w.capabilities)
}

// Clone creates a clone of the wrapper
func (w *DefaultAgentWrapper) Clone() *DefaultAgentWrapper {
	// Clone capabilities
	capabilities := make([]string, len(w.capabilities))
	copy(capabilities, w.capabilities)

	// Clone metadata
	metadata := make(map[string]interface{})
	for k, v := range w.metadata {
		metadata[k] = v
	}

	return &DefaultAgentWrapper{
		agent:        w.agent,
		capabilities: capabilities,
		model:        w.model,
		metadata:     metadata,
	}
}
