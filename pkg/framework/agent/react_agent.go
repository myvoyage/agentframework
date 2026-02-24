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

	"github.com/cloudwego/eino/components/model"
	flowagent "github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type ReActAgent struct {
	name          string
	inner         *react.Agent
	thread        *Thread
	baseOptions   []flowagent.AgentOption
	memoryManager MemoryManager
}

// NewReActAgent creates a new ReActAgent with memory management support
func NewReActAgent(name string, inner *react.Agent, opts ...flowagent.AgentOption) *ReActAgent {
	return &ReActAgent{
		name:          name,
		inner:         inner,
		baseOptions:   opts,
		memoryManager: NewMemoryManager(MemoryOptions{}),
	}
}

// NewReActAgentWithMemory creates a new ReActAgent with custom memory options
func NewReActAgentWithMemory(name string, inner *react.Agent, memoryOpts MemoryOptions, opts ...flowagent.AgentOption) *ReActAgent {
	return &ReActAgent{
		name:          name,
		inner:         inner,
		baseOptions:   opts,
		memoryManager: NewMemoryManager(memoryOpts),
	}
}

func (r *ReActAgent) Name() string {
	return r.name
}

func (r *ReActAgent) UseThread(thread *Thread) {
	r.thread = thread
}

// SetMemoryOptions updates the memory management options
func (r *ReActAgent) SetMemoryOptions(opts MemoryOptions) {
	r.memoryManager.SetOptions(opts)
}

// GetMemoryOptions returns the current memory management options
func (r *ReActAgent) GetMemoryOptions() MemoryOptions {
	return r.memoryManager.GetOptions()
}

// ClearHistory clears the message history
func (r *ReActAgent) ClearHistory() {
	if r.thread != nil {
		r.thread.Messages = r.memoryManager.ClearHistory()
	}
}

func (r *ReActAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	msg := schema.UserMessage(input)

	// Build messages including history if enabled
	var messages []*schema.Message
	if r.memoryManager.GetOptions().EnableTrimming && r.thread != nil && len(r.thread.Messages) > 0 {
		messages = append(messages, r.thread.Messages...)
	}
	messages = append(messages, msg)

	var agentOpts []flowagent.AgentOption
	if len(r.baseOptions) > 0 {
		agentOpts = append(agentOpts, r.baseOptions...)
	}
	if len(opts) > 0 {
		agentOpts = append(agentOpts, react.WithChatModelOptions(opts...))
	}

	resp, err := r.inner.Generate(ctx, messages, agentOpts...)
	if err != nil {
		return nil, err
	}

	if r.thread != nil && r.memoryManager.GetOptions().EnableTrimming {
		r.thread.Messages = append(r.thread.Messages, msg, resp)
		// Apply memory management
		r.thread.Messages = r.memoryManager.LimitHistory(r.thread.Messages)
	}

	return resp, nil
}

func (r *ReActAgent) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg := schema.UserMessage(input)

	// Build messages including history if enabled
	var messages []*schema.Message
	if r.memoryManager.GetOptions().EnableTrimming && r.thread != nil && len(r.thread.Messages) > 0 {
		messages = append(messages, r.thread.Messages...)
	}
	messages = append(messages, msg)

	var agentOpts []flowagent.AgentOption
	if len(r.baseOptions) > 0 {
		agentOpts = append(agentOpts, r.baseOptions...)
	}
	if len(opts) > 0 {
		agentOpts = append(agentOpts, react.WithChatModelOptions(opts...))
	}

	// ReAct Agent's Stream method returns the stream of the *final* response generation,
	// or potentially intermediate steps depending on configuration.
	// We need to ensure the inner agent supports Stream.
	return r.inner.Stream(ctx, messages, agentOpts...)
}
