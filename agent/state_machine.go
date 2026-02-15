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
	"sync"
	"time"
)

// AgentState 表示智能体的状态
type AgentState int

const (
	// StateIdle 空闲状态，等待任务
	StateIdle AgentState = iota
	// StateRunning 运行中，正在处理任务
	StateRunning
	// StateFinished 任务完成
	StateFinished
	// StateError 发生错误
	StateError
	// StatePaused 暂停状态
	StatePaused
	// StateCancelled 任务被取消
	StateCancelled
)

// String 返回状态的字符串表示
func (s AgentState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateRunning:
		return "RUNNING"
	case StateFinished:
		return "FINISHED"
	case StateError:
		return "ERROR"
	case StatePaused:
		return "PAUSED"
	case StateCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON 实现 JSON 序列化
func (s AgentState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// StateTransition 状态转换记录
type StateTransition struct {
	From      AgentState     `json:"from"`
	To        AgentState     `json:"to"`
	Timestamp time.Time      `json:"timestamp"`
	Reason    string         `json:"reason,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// StateHook 状态钩子函数类型
type StateHook func(ctx context.Context, transition StateTransition) error

// StateMachine 状态机，管理智能体的状态转换
type StateMachine struct {
	current      AgentState
	transitions  map[AgentState][]AgentState // 允许的状态转换
	hooks        map[AgentState][]StateHook  // 各状态的钩子
	history      []StateTransition           // 状态历史
	historyLimit int                          // 历史记录限制
	mu           sync.RWMutex
	onTransition []func(StateTransition)      // 转换回调
}

// StateMachineConfig 状态机配置
type StateMachineConfig struct {
	// HistoryLimit 状态历史记录最大数量，0 表示不限制
	HistoryLimit int
	// EnableHistory 是否启用状态历史记录
	EnableHistory bool
}

// NewStateMachine 创建新的状态机
func NewStateMachine(config StateMachineConfig) *StateMachine {
	sm := &StateMachine{
		current:      StateIdle,
		transitions:  make(map[AgentState][]AgentState),
		hooks:        make(map[AgentState][]StateHook),
		history:      make([]StateTransition, 0),
		historyLimit: config.HistoryLimit,
		onTransition: make([]func(StateTransition), 0),
	}

	// 设置默认的状态转换规则
	sm.setupDefaultTransitions()

	return sm
}

// setupDefaultTransitions 设置默认的状态转换规则
func (sm *StateMachine) setupDefaultTransitions() {
	// IDLE 可以转换为任何状态
	sm.transitions[StateIdle] = []AgentState{
		StateRunning, StatePaused, StateCancelled,
	}

	// RUNNING 可以转换为 FINISHED, ERROR, PAUSED, CANCELLED
	sm.transitions[StateRunning] = []AgentState{
		StateFinished, StateError, StatePaused, StateCancelled,
	}

	// PAUSED 可以转换为 RUNNING, CANCELLED
	sm.transitions[StatePaused] = []AgentState{
		StateRunning, StateCancelled,
	}

	// FINISHED, ERROR, CANCELLED 是最终状态
	sm.transitions[StateFinished] = []AgentState{}
	sm.transitions[StateError] = []AgentState{}
	sm.transitions[StateCancelled] = []AgentState{}

	// ERROR 可以重新开始（可选）
	sm.transitions[StateError] = []AgentState{
		StateIdle, StateRunning,
	}
}

// Current 返回当前状态
func (sm *StateMachine) Current() AgentState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// CanTransition 检查是否可以转换到目标状态
func (sm *StateMachine) CanTransition(to AgentState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 如果已经是目标状态，不需要转换
	if sm.current == to {
		return true
	}

	allowed, exists := sm.transitions[sm.current]
	if !exists {
		return false
	}

	for _, state := range allowed {
		if state == to {
			return true
		}
	}

	return false
}

// Transition 执行状态转换
func (sm *StateMachine) Transition(ctx context.Context, to AgentState, reason string, metadata map[string]any) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 如果已经是目标状态，直接返回
	if sm.current == to {
		return nil
	}

	// 检查转换是否允许
	if !sm.isAllowedTransition(sm.current, to) {
		return &InvalidStateTransitionError{
			From: sm.current,
			To:   to,
		}
	}

	// 创建状态转换记录
	transition := StateTransition{
		From:      sm.current,
		To:        to,
		Timestamp: time.Now(),
		Reason:    reason,
		Metadata:  metadata,
	}

	// 执行源状态的退出钩子
	if hooks, exists := sm.hooks[sm.current]; exists {
		for _, hook := range hooks {
			if err := hook(ctx, transition); err != nil {
				return fmt.Errorf("exit hook failed for state %s: %w", sm.current, err)
			}
		}
	}

	// 更新状态
	oldState := sm.current
	sm.current = to

	// 记录历史
	if sm.historyLimit == 0 || len(sm.history) < sm.historyLimit {
		sm.history = append(sm.history, transition)
	} else if sm.historyLimit > 0 {
		// 移除最旧的记录
		sm.history = sm.history[1:]
		sm.history = append(sm.history, transition)
	}

	// 执行目标状态的进入钩子
	if hooks, exists := sm.hooks[to]; exists {
		for _, hook := range hooks {
			if err := hook(ctx, transition); err != nil {
				// 钩子失败，回滚状态
				sm.current = oldState
				return fmt.Errorf("enter hook failed for state %s: %w", to, err)
			}
		}
	}

	// 触发转换回调
	for _, callback := range sm.onTransition {
		callback(transition)
	}

	return nil
}

// isAllowedTransition 检查转换是否被允许（内部方法，假设已持有锁）
func (sm *StateMachine) isAllowedTransition(from, to AgentState) bool {
	allowed, exists := sm.transitions[from]
	if !exists {
		return false
	}

	for _, state := range allowed {
		if state == to {
			return true
		}
	}

	return false
}

// AddTransition 添加状态转换规则
func (sm *StateMachine) AddTransition(from, to AgentState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.transitions[from]; !exists {
		sm.transitions[from] = make([]AgentState, 0)
	}

	sm.transitions[from] = append(sm.transitions[from], to)
}

// RemoveTransition 移除状态转换规则
func (sm *StateMachine) RemoveTransition(from, to AgentState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if allowed, exists := sm.transitions[from]; exists {
		for i, state := range allowed {
			if state == to {
				sm.transitions[from] = append(allowed[:i], allowed[i+1:]...)
				break
			}
		}
	}
}

// AddHook 添加状态钩子
func (sm *StateMachine) AddHook(state AgentState, hook StateHook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.hooks[state]; !exists {
		sm.hooks[state] = make([]StateHook, 0)
	}

	sm.hooks[state] = append(sm.hooks[state], hook)
}

// RemoveHook 移除状态钩子
func (sm *StateMachine) RemoveHook(state AgentState, hook StateHook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if hooks, exists := sm.hooks[state]; exists {
		for i, h := range hooks {
			// 使用指针比较（简单实现）
			if &h == &hook {
				sm.hooks[state] = append(hooks[:i], hooks[i+1:]...)
				break
			}
		}
	}
}

// OnTransition 注册状态转换回调
func (sm *StateMachine) OnTransition(callback func(transition StateTransition)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.onTransition = append(sm.onTransition, callback)
}

// History 返回状态历史记录
func (sm *StateMachine) History() []StateTransition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 返回副本
	history := make([]StateTransition, len(sm.history))
	copy(history, sm.history)
	return history
}

// ClearHistory 清除状态历史
func (sm *StateMachine) ClearHistory() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.history = make([]StateTransition, 0)
}

// Reset 重置状态机到初始状态
func (sm *StateMachine) Reset() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.current = StateIdle
	sm.history = make([]StateTransition, 0)
	return nil
}

// IsTerminal 检查当前状态是否为最终状态
func (sm *StateMachine) IsTerminal() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch sm.current {
	case StateFinished, StateError, StateCancelled:
		return true
	default:
		return false
	}
}

// IsActive 检查状态机是否处于活动状态（非空闲、非最终状态）
func (sm *StateMachine) IsActive() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.current == StateRunning || sm.current == StatePaused
}

// ==================== 错误类型 ====================

// InvalidStateTransitionError 无效的状态转换错误
type InvalidStateTransitionError struct {
	From AgentState
	To   AgentState
}

// Error 实现 error 接口
func (e *InvalidStateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s", e.From, e.To)
}

// ==================== 辅助函数 ====================

// NewStateMachineWithDefaults 创建带有默认配置的状态机
func NewStateMachineWithDefaults() *StateMachine {
	return NewStateMachine(StateMachineConfig{
		HistoryLimit:   100,
		EnableHistory:  true,
	})
}

// AgentStateMachine 是智能体使用的状态机接口
// 这个接口定义了智能体状态机应该提供的基本功能
type AgentStateMachine interface {
	// Current 返回当前状态
	Current() AgentState
	// CanTransition 检查是否可以转换到目标状态
	CanTransition(to AgentState) bool
	// Transition 执行状态转换
	Transition(ctx context.Context, to AgentState, reason string, metadata map[string]any) error
	// History 返回状态历史记录
	History() []StateTransition
	// IsTerminal 检查是否为最终状态
	IsTerminal() bool
	// IsActive 检查是否为活动状态
	IsActive() bool
	// AddHook 添加状态钩子
	AddHook(state AgentState, hook StateHook)
	// OnTransition 注册转换回调
	OnTransition(callback func(transition StateTransition))
	// Reset 重置状态机
	Reset() error
}

// 确保 StateMachine 实现了 AgentStateMachine 接口
var _ AgentStateMachine = (*StateMachine)(nil)

// StateMachineToSchema 将状态机状态转换为 schema.Message 的 metadata
func StateMachineToMetadata(sm AgentStateMachine) map[string]string {
	meta := make(map[string]string)

	meta["state"] = sm.Current().String()
	meta["is_terminal"] = fmt.Sprintf("%t", sm.IsTerminal())
	meta["is_active"] = fmt.Sprintf("%t", sm.IsActive())

	// 添加最近的状态转换
	if history := sm.History(); len(history) > 0 {
		last := history[len(history)-1]
		meta["last_transition_from"] = last.From.String()
		meta["last_transition_to"] = last.To.String()
		meta["last_transition_reason"] = last.Reason
		meta["last_transition_time"] = last.Timestamp.Format(time.RFC3339)
	}

	return meta
}
