//go:build experimental

// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// ReAct Agent 核心类型定义 - 简化版本
// Copyright (C) 2025 Agent Framework Contributors

package react

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/planning"
)

// Model is a simplified interface for language models
type Model interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// ErrorCode constants
const (
	ErrorTypeExecution = "execution_error"
	ErrorTypeTimeout   = "timeout_error"
	ErrorTypeValidation = "validation_error"
	ErrorTypeTool      = "tool_error"
)

// ReActActionType 定义 ReAct 动作类型
type ReActActionType string

const (
	ActionTypeThink   ReActActionType = "think"
	ActionTypeTool    ReActActionType = "tool"
	ActionTypeSearch  ReActActionType = "search"
	ActionTypeReflect ReActActionType = "reflect"
	ActionTypeFinish  ReActActionType = "finish"
)

// String 返回动作类型的字符串表示
func (t ReActActionType) String() string {
	return string(t)
}

// IsValid 检查动作类型是否有效
func (t ReActActionType) IsValid() bool {
	switch t {
	case ActionTypeThink, ActionTypeTool, ActionTypeSearch, ActionTypeReflect, ActionTypeFinish:
		return true
	default:
		return false
	}
}

// Thought 表示 ReAct 循环中的思考步骤
type Thought struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	Content            string                 `json:"content"`
	Reasoning          string                 `json:"reasoning,omitempty"`
	Confidence         float64                `json:"confidence,omitempty"`
	Context            map[string]interface{} `json:"context,omitempty"`
	AssociatedContexts []string               `json:"associated_contexts,omitempty"`
}

// NewThought 创建新的思考对象
func NewThought(content string, reasoning string, confidence float64) *Thought {
	thought := &Thought{
		ID:                 uuid.New().String(),
		Timestamp:          time.Now().UTC(),
		Content:            content,
		Reasoning:          reasoning,
		Confidence:         confidence,
		Context:            make(map[string]interface{}),
		AssociatedContexts: make([]string, 0),
	}
	return thought
}

// Clone 克隆思考对象
func (t *Thought) Clone() *Thought {
	cloned := &Thought{
		ID:                 t.ID,
		Timestamp:          t.Timestamp,
		Content:            t.Content,
		Reasoning:          t.Reasoning,
		Confidence:         t.Confidence,
		Context:            make(map[string]interface{}),
		AssociatedContexts: make([]string, len(t.AssociatedContexts)),
	}

	for k, v := range t.Context {
		cloned.Context[k] = v
	}
	copy(cloned.AssociatedContexts, t.AssociatedContexts)

	return cloned
}

// Action 表示 ReAct 循环中的动作步骤
type Action struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	Type               ReActActionType        `json:"type"`
	Name               string                 `json:"name"`
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
	Description        string                 `json:"description,omitempty"`
	ExpectedOutcome    string                 `json:"expected_outcome,omitempty"`
	Dependencies       []string               `json:"dependencies,omitempty"`
	AssociatedContexts []string               `json:"associated_contexts,omitempty"`
}

// NewAction 创建新的动作对象
func NewAction(actionType ReActActionType, name string, parameters map[string]interface{}, description ...string) (*Action, error) {
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return &Action{
		ID:                 uuid.New().String(),
		Timestamp:          time.Now().UTC(),
		Type:               actionType,
		Name:               name,
		Parameters:         parameters,
		Description:        desc,
		Dependencies:       make([]string, 0),
		AssociatedContexts: make([]string, 0),
	}, nil
}

// Observation 表示 ReAct 循环中的观察步骤
type Observation struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	ActionID           string                 `json:"action_id"`
	Success            bool                   `json:"success"`
	Result             interface{}            `json:"result,omitempty"`
	Error              string                 `json:"error,omitempty"`
	ExecutionTime      time.Duration          `json:"execution_time,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	AssociatedContexts []string               `json:"associated_contexts,omitempty"`
}

// NewObservation 创建新的观察对象
func NewObservation(actionID string, success bool, result interface{}, errorMsg string) *Observation {
	obs := &Observation{
		ID:                 uuid.New().String(),
		Timestamp:          time.Now().UTC(),
		ActionID:           actionID,
		Success:            success,
		Result:             result,
		Error:              errorMsg,
		Metadata:           make(map[string]interface{}),
		AssociatedContexts: make([]string, 0),
	}
	return obs
}

// ResultSummary 返回结果摘要
func (o *Observation) ResultSummary() string {
	if o.Success {
		if resultStr, ok := o.Result.(string); ok {
			if len(resultStr) > 100 {
				return resultStr[:100] + "..."
			}
			return resultStr
		}
		return fmt.Sprintf("%v", o.Result)
	}

	if o.Error != "" {
		if len(o.Error) > 100 {
			return "Error: " + o.Error[:100] + "..."
		}
		return "Error: " + o.Error
	}

	return "No result available"
}

// Plan 表示 ReAct 循环中的规划步骤
type Plan struct {
	ID                 string                 `json:"id"`
	Timestamp          time.Time              `json:"timestamp"`
	Goal               string                 `json:"goal"`
	Steps              []*planning.Step      `json:"steps"`
	EstimatedDuration  time.Duration          `json:"estimated_duration"`
	Priority           int                    `json:"priority"`
	Status             planning.PlanStatus    `json:"status"`
	Dependencies       map[string][]string    `json:"dependencies"`
	AssociatedContexts []string               `json:"associated_contexts,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NewPlan 创建新的规划对象
func NewPlan(goal string) *Plan {
	plan := &Plan{
		ID:                 uuid.New().String(),
		Timestamp:          time.Now().UTC(),
		Goal:               goal,
		Steps:              make([]*planning.Step, 0),
		Priority:           0,
		Status:             planning.PlanStatusDraft,
		Dependencies:       make(map[string][]string),
		AssociatedContexts: make([]string, 0),
		Metadata:           make(map[string]interface{}),
	}
	return plan
}

// Clone 克隆规划对象
func (p *Plan) Clone() *Plan {
	cloned := &Plan{
		ID:                 p.ID,
		Timestamp:          p.Timestamp,
		Goal:               p.Goal,
		EstimatedDuration:  p.EstimatedDuration,
		Priority:           p.Priority,
		Status:             p.Status,
		AssociatedContexts: make([]string, len(p.AssociatedContexts)),
	}

	cloned.Steps = make([]*planning.Step, len(p.Steps))
	for i, step := range p.Steps {
		if step != nil {
			cloned.Steps[i] = &planning.Step{
				ID:                step.ID,
				Name:              step.Name,
				Description:       step.Description,
				Order:             step.Order,
				Status:            step.Status,
				Priority:          step.Priority,
				EstimatedDuration: step.EstimatedDuration,
				Dependencies:      make([]string, len(step.Dependencies)),
				CreatedAt:         step.CreatedAt,
				UpdatedAt:         step.UpdatedAt,
			}
			copy(cloned.Steps[i].Dependencies, step.Dependencies)
		}
	}

	cloned.Dependencies = make(map[string][]string)
	for key, deps := range p.Dependencies {
		cloned.Dependencies[key] = make([]string, len(deps))
		copy(cloned.Dependencies[key], deps)
	}

	copy(cloned.AssociatedContexts, p.AssociatedContexts)

	return cloned
}

// ReActStatus 定义 ReAct 循环状态
type ReActStatus string

const (
	ReActStatusThinking ReActStatus = "thinking"
	ReActStatusActing   ReActStatus = "acting"
	ReActStatusObserving ReActStatus = "observing"
	ReActStatusReflecting ReActStatus = "reflecting"
	ReActStatusCompleted  ReActStatus = "completed"
	ReActStatusFailed     ReActStatus = "failed"
)

// String 返回状态的字符串表示
func (s ReActStatus) String() string {
	return string(s)
}

// IsValid 检查状态是否有效
func (s ReActStatus) IsValid() bool {
	switch s {
	case ReActStatusThinking, ReActStatusActing, ReActStatusObserving,
	     ReActStatusReflecting, ReActStatusCompleted, ReActStatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal 检查状态是否为终止状态
func (s ReActStatus) IsTerminal() bool {
	return s == ReActStatusCompleted || s == ReActStatusFailed
}

// ReActState 表示 ReAct 循环的状态
type ReActState struct {
	SessionID          string                 `json:"session_id"`
	AgentID            string                 `json:"agent_id"`
	Query              string                 `json:"query"`
	CurrentPlan        *Plan                  `json:"current_plan,omitempty"`
	Thoughts           []*Thought             `json:"thoughts,omitempty"`
	Actions            []*Action              `json:"actions,omitempty"`
	Observations       []*Observation         `json:"observations,omitempty"`
	Memory             interface{}            `json:"memory,omitempty"`
	Status             ReActStatus            `json:"status"`
	IterationCount     int                    `json:"iteration_count"`
	MaxIterations      int                    `json:"max_iterations"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	AssociatedContexts []string               `json:"associated_contexts,omitempty"`
}

// NewReActState 创建新的 ReAct 状态对象
func NewReActState(sessionID string, agentID string, query string, maxIterations int) *ReActState {
	state := &ReActState{
		SessionID:          sessionID,
		AgentID:            agentID,
		Query:              query,
		Status:             ReActStatusThinking,
		IterationCount:     0,
		MaxIterations:      maxIterations,
		StartTime:          time.Now().UTC(),
		Thoughts:           make([]*Thought, 0),
		Actions:            make([]*Action, 0),
		Observations:       make([]*Observation, 0),
		Metadata:           make(map[string]interface{}),
		AssociatedContexts: make([]string, 0),
	}
	return state
}

// AddThought 添加思考到状态
func (rs *ReActState) AddThought(thought *Thought) {
	rs.Thoughts = append(rs.Thoughts, thought)
}

// AddAction 添加动作到状态
func (rs *ReActState) AddAction(action *Action) {
	rs.Actions = append(rs.Actions, action)
}

// AddObservation 添加观察到状态
func (rs *ReActState) AddObservation(observation *Observation) {
	rs.Observations = append(rs.Observations, observation)
}

// ToJSON 将 ReAct 状态对象序列化为 JSON
func (rs *ReActState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(rs, "", "  ")
}

// FromJSON 从 JSON 反序列化 ReAct 状态对象
func (rs *ReActState) FromJSON(data []byte) error {
	return json.Unmarshal(data, rs)
}

// Clone 克隆 ReAct 状态对象
func (rs *ReActState) Clone() *ReActState {
	cloned := &ReActState{
		SessionID:          rs.SessionID,
		AgentID:            rs.AgentID,
		Query:              rs.Query,
		CurrentPlan:        rs.CurrentPlan,
		Memory:             rs.Memory,
		Status:             rs.Status,
		IterationCount:     rs.IterationCount,
		MaxIterations:      rs.MaxIterations,
		StartTime:          rs.StartTime,
		EndTime:            rs.EndTime,
		Metadata:           make(map[string]interface{}),
		AssociatedContexts: make([]string, len(rs.AssociatedContexts)),
	}

	// 深拷贝思考步骤
	cloned.Thoughts = make([]*Thought, len(rs.Thoughts))
	for i, thought := range rs.Thoughts {
		cloned.Thoughts[i] = thought.Clone()
	}

	// 深拷贝动作步骤
	cloned.Actions = make([]*Action, len(rs.Actions))
	for i, action := range rs.Actions {
		if action != nil {
			cloned.Actions[i] = &Action{
				ID:                 action.ID,
				Timestamp:          action.Timestamp,
				Type:               action.Type,
				Name:               action.Name,
				Parameters:         make(map[string]interface{}),
				Description:        action.Description,
				ExpectedOutcome:    action.ExpectedOutcome,
				Dependencies:       make([]string, len(action.Dependencies)),
				AssociatedContexts: make([]string, len(action.AssociatedContexts)),
			}

			for k, v := range action.Parameters {
				cloned.Actions[i].Parameters[k] = v
			}

			copy(cloned.Actions[i].Dependencies, action.Dependencies)
			copy(cloned.Actions[i].AssociatedContexts, action.AssociatedContexts)
		}
	}

	// 深拷贝观察结果
	cloned.Observations = make([]*Observation, len(rs.Observations))
	for i, observation := range rs.Observations {
		if observation != nil {
			cloned.Observations[i] = &Observation{
				ID:                 observation.ID,
				Timestamp:          observation.Timestamp,
				ActionID:           observation.ActionID,
				Success:            observation.Success,
				Result:             observation.Result,
				Error:              observation.Error,
				ExecutionTime:      observation.ExecutionTime,
				Metadata:           make(map[string]interface{}),
				AssociatedContexts: make([]string, len(observation.AssociatedContexts)),
			}

			for k, v := range observation.Metadata {
				cloned.Observations[i].Metadata[k] = v
			}

			copy(cloned.Observations[i].AssociatedContexts, observation.AssociatedContexts)
		}
	}

	// 深拷贝元数据
	for key, value := range rs.Metadata {
		cloned.Metadata[key] = value
	}

	// 拷贝关联上下文列表
	copy(cloned.AssociatedContexts, rs.AssociatedContexts)

	return cloned
}

// ReActConfig 表示 ReAct Agent 的配置
type ReActConfig struct {
	// AgentID Agent唯一标识
	AgentID string `json:"agent_id"`
	// AgentName Agent名称
	AgentName string `json:"agent_name"`
	// ModelName 模型名称
	ModelName string `json:"model_name"`
	// MaxIterations 最大迭代次数
	MaxIterations int `json:"max_iterations"`
	// MaxTokens 最大token数
	MaxTokens int `json:"max_tokens"`
	// Temperature 温度参数
	Temperature float64 `json:"temperature"`
	// MaxExecutionTime 最大执行时间
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	// EnableThoughtChain 是否启用思考链
	EnableThoughtChain bool `json:"enable_thought_chain"`
	// EnablePlanning 是否启用规划
	EnablePlanning bool `json:"enable_planning"`
	// EnableMemoryIntegration 是否启用内存集成
	EnableMemoryIntegration bool `json:"enable_memory_integration"`
	// EnableReflection 是否启用反思
	EnableReflection bool `json:"enable_reflection"`
	// EnableToolExecution 是否启用工具执行
	EnableToolExecution bool `json:"enable_tool_execution"`
	// ConfidenceThreshold 置信度阈值
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	// EnableParallelExecution 是否启用并行执行
	EnableParallelExecution bool `json:"enable_parallel_execution"`
	// MaxParallelActions 最大并行动作数
	MaxParallelActions int `json:"max_parallel_actions"`
	// ToolTimeout 工具执行超时时间
	ToolTimeout time.Duration `json:"tool_timeout"`
	// EnableMetrics 是否启用指标收集
	EnableMetrics bool `json:"enable_metrics"`
	// LogLevel 日志级别
	LogLevel string `json:"log_level"`

	// 上下文和日志
	Logger *zap.Logger `json:"-"`
}

// DefaultReActConfig 返回默认的 ReAct 配置
func DefaultReActConfig() *ReActConfig {
	return &ReActConfig{
		AgentID:                  "default_react_agent",
		AgentName:                "ReActAgent",
		ModelName:                "gpt-4",
		MaxIterations:            10,
		MaxTokens:                2000,
		Temperature:              0.7,
		MaxExecutionTime:        5 * time.Minute,
		EnableThoughtChain:       true,
		EnablePlanning:           true,
		EnableMemoryIntegration:  true,
		EnableReflection:        true,
		EnableToolExecution:      true,
		ConfidenceThreshold:      0.7,
		EnableParallelExecution:  false,
		MaxParallelActions:      3,
		ToolTimeout:              30 * time.Second,
		EnableMetrics:            true,
		LogLevel:                 "info",
	}
}

// Validate 验证配置有效性
// 【推荐】验证配置参数
func (rc *ReActConfig) Validate() error {
	if rc.MaxIterations <= 0 {
		return errors.NewValidationError("max_iterations must be positive", nil)
	}

	if rc.MaxExecutionTime <= 0 {
		return errors.NewValidationError("max_execution_time must be positive", nil)
	}

	if rc.ConfidenceThreshold < 0 || rc.ConfidenceThreshold > 1 {
		return errors.NewValidationError("confidence_threshold must be between 0 and 1", nil)
	}

	if rc.MaxParallelActions <= 0 {
		return errors.NewValidationError("max_parallel_actions must be positive", nil)
	}

	if rc.ToolTimeout <= 0 {
		return errors.NewValidationError("tool_timeout must be positive", nil)
	}

	return nil
}

// Clone 克隆配置
// 【推荐】返回配置的深拷贝
func (rc *ReActConfig) Clone() *ReActConfig {
	if rc == nil {
		return nil
	}

	clone := *rc
	return &clone
}

// NewReActConfig 创建新的 ReAct 配置
// 【必须】提供配置的构造函数
func NewReActConfig() *ReActConfig {
	return DefaultReActConfig()
}
