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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ChatModel interface {
	Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error)
	Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

type Agent interface {
	Name() string
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
}

type StreamableAgent interface {
	Agent
	Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

type Thread struct {
	ID       string
	Messages []*schema.Message
}

type ChatAgentConfig struct {
	Name         string
	Instructions string
	Model        ChatModel
	Tools        []tool.BaseTool
	MemoryOpts   MemoryOptions // Memory management options
}

type ChatAgent struct {
	name          string
	instructions  string
	model         ChatModel
	thread        *Thread
	tools         map[string]tool.InvokableTool
	memoryManager *MemoryManager
	stateMachine  *StateMachine
	messagePool    *MessagePool
}

func NewChatAgent(ctx context.Context, cfg ChatAgentConfig) (*ChatAgent, error) {
	boundModel := cfg.Model
	toolMap := make(map[string]tool.InvokableTool)

	if len(cfg.Tools) > 0 {
		var infos []*schema.ToolInfo
		for _, t := range cfg.Tools {
			info, err := t.Info(ctx)
			if err != nil {
				return nil, err
			}
			if info == nil {
				continue
			}
			infos = append(infos, info)

			if inv, ok := t.(tool.InvokableTool); ok {
				toolMap[info.Name] = inv
			}
		}

		if len(infos) > 0 {
			if m, ok := cfg.Model.(model.ToolCallingChatModel); ok {
				tm, err := m.WithTools(infos)
				if err != nil {
					return nil, err
				}
				boundModel = tm
			}
		}
	}

	// Create memory manager with provided or default options
	memoryManager := NewMemoryManager(cfg.MemoryOpts)

	// Create state machine
	stateMachine := NewStateMachineWithDefaults()

	return &ChatAgent{
		name:          cfg.Name,
		instructions:  cfg.Instructions,
		model:         boundModel,
		thread:        &Thread{ID: cfg.Name},
		tools:         toolMap,
		memoryManager: memoryManager,
		stateMachine:  stateMachine,
		messagePool:    NewMessagePool(),
	}, nil
}

func (a *ChatAgent) Name() string {
	return a.name
}

func (a *ChatAgent) UseThread(thread *Thread) {
	a.thread = thread
}

// SetMemoryOptions updates the memory management options
func (a *ChatAgent) SetMemoryOptions(opts MemoryOptions) {
	a.memoryManager.SetOptions(opts)
}

// GetMemoryOptions returns the current memory management options
func (a *ChatAgent) GetMemoryOptions() MemoryOptions {
	return a.memoryManager.GetOptions()
}

// ClearHistory clears the message history
func (a *ChatAgent) ClearHistory() {
	a.thread.Messages = a.memoryManager.ClearHistory()
}

// GetHistory returns the conversation history
func (a *ChatAgent) GetHistory() []*ThreadMessage {
	messages := a.thread.Messages
	result := make([]*ThreadMessage, 0, len(messages))
	for _, msg := range messages {
		result = append(result, &ThreadMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}
	return result
}

// ThreadMessage represents a message in the conversation thread
type ThreadMessage struct {
	Role    string
	Content string
}

func (a *ChatAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// 如果处于最终状态，先重置到 IDLE（支持多轮对话）
	if a.stateMachine.IsTerminal() {
		if err := a.stateMachine.Transition(ctx, StateIdle, "Resetting for new conversation turn", nil); err != nil {
			// 如果转换失败，尝试强制重置
			_ = a.stateMachine.Reset()
		}
	}

	// 转换到 RUNNING 状态
	if err := a.stateMachine.Transition(ctx, StateRunning, "Starting Run execution", map[string]any{
		"input_length": len(input),
		"tool_count":   len(a.tools),
	}); err != nil {
		return nil, fmt.Errorf("state transition failed: %w", err)
	}

	// 确保在函数结束时处理状态转换
	defer func() {
		if a.stateMachine.Current() == StateRunning {
			// 如果仍然在运行状态，转换为 FINISHED
			_ = a.stateMachine.Transition(context.Background(), StateFinished, "Run execution completed", nil)
		}
	}()

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	// Process user message (check and trim if necessary)
	userMsg = a.memoryManager.ProcessMessage(userMsg)

	messages := a.buildMessages(userMsg)

	resp, err := a.model.Generate(ctx, messages, opts...)
	if err != nil {
		_ = a.stateMachine.Transition(ctx, StateError, "Model generation failed", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}

	if len(resp.ToolCalls) > 0 && len(a.tools) > 0 {
		toolMsgs, err := a.runTools(ctx, resp)
		if err != nil {
			_ = a.stateMachine.Transition(ctx, StateError, "Tool execution failed", map[string]any{
				"error": err.Error(),
			})
			return nil, err
		}
		messages = append(messages, resp)
		messages = append(messages, toolMsgs...)

		resp, err = a.model.Generate(ctx, messages, opts...)
		if err != nil {
			_ = a.stateMachine.Transition(ctx, StateError, "Model generation after tools failed", map[string]any{
				"error": err.Error(),
			})
			return nil, err
		}
	}

	// Process response (check and trim if necessary)
	resp = a.memoryManager.ProcessMessage(resp)

	a.thread.Messages = append(a.thread.Messages, userMsg, resp)

	// Apply intelligent message trimming
	a.thread.Messages = a.memoryManager.LimitHistory(a.thread.Messages)

	// 转换到 FINISHED 状态
	_ = a.stateMachine.Transition(ctx, StateFinished, "Run execution completed successfully", map[string]any{
		"response_length": len(resp.Content),
		"tool_calls_count": len(resp.ToolCalls),
	})

	return resp, nil
}

func (a *ChatAgent) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	// Process user message (check and trim if necessary)
	userMsg = a.memoryManager.ProcessMessage(userMsg)

	msgs := a.buildMessages(userMsg)
	return a.model.Stream(ctx, msgs, opts...)
}

func (a *ChatAgent) runTools(ctx context.Context, toolCallMsg *schema.Message) ([]*schema.Message, error) {
	var toolMessages []*schema.Message

	for _, call := range toolCallMsg.ToolCalls {
		name := call.Function.Name
		t, ok := a.tools[name]
		if !ok {
			continue
		}

		output, err := t.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			return nil, err
		}

		toolMsg := &schema.Message{
			Role:       schema.Tool,
			Content:    output,
			ToolCallID: call.ID,
		}

		// Process tool message (check and trim if necessary)
		toolMsg = a.memoryManager.ProcessMessage(toolMsg)

		toolMessages = append(toolMessages, toolMsg)
	}

	return toolMessages, nil
}

func (a *ChatAgent) buildMessages(latest *schema.Message) []*schema.Message {
	var msgs []*schema.Message
	if a.instructions != "" {
		msgs = append(msgs, &schema.Message{
			Role:    schema.System,
			Content: a.instructions,
		})
	}
	if a.thread != nil && len(a.thread.Messages) > 0 {
		msgs = append(msgs, a.thread.Messages...)
	}
	msgs = append(msgs, latest)
	return msgs
}

// ==================== 状态机相关方法 ====================

// State 返回当前状态
func (a *ChatAgent) State() AgentState {
	return a.stateMachine.Current()
}

// StateHistory 返回状态历史记录
func (a *ChatAgent) StateHistory() []StateTransition {
	return a.stateMachine.History()
}

// IsTerminal 检查是否处于最终状态
func (a *ChatAgent) IsTerminal() bool {
	return a.stateMachine.IsTerminal()
}

// IsActive 检查是否处于活动状态
func (a *ChatAgent) IsActive() bool {
	return a.stateMachine.IsActive()
}

// AddStateHook 添加状态钩子
func (a *ChatAgent) AddStateHook(state AgentState, hook StateHook) {
	a.stateMachine.AddHook(state, hook)
}

// OnStateTransition 注册状态转换回调
func (a *ChatAgent) OnStateTransition(callback func(transition StateTransition)) {
	a.stateMachine.OnTransition(callback)
}

// ResetStateMachine 重置状态机
func (a *ChatAgent) ResetStateMachine() error {
	return a.stateMachine.Reset()
}
