// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// ReAct Agent 核心接口定义
// Copyright (C) 2025 Agent Framework Contributors

package react

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"AgentFramework/pkg/errors"
)

// ReActAgent ReAct Agent接口
// 【必须】定义ReAct Agent的标准接口
type ReActAgent interface {
	// Name 返回Agent名称
	Name() string

	// Start 启动ReAct循环
	// 【必须】开始处理用户查询，返回初始状态
	Start(ctx context.Context, query string) (*ReActState, error)

	// Step 执行一个ReAct步骤
	// 【必须】执行单次Think-Act-Observe循环，返回更新后的状态
	Step(ctx context.Context, state *ReActState) (*ReActState, error)

	// Run 完整运行ReAct循环
	// 【必须】从开始到结束完整执行ReAct循环
	Run(ctx context.Context, query string) (*ReActState, error)

	// GetState 获取当前状态
	// 【必须】获取Agent的当前ReAct状态
	GetState() *ReActState

	// SetState 设置当前状态
	// 【必须】设置Agent的当前ReAct状态
	SetState(state *ReActState) error

	// GetCapabilities 获取Agent能力列表
	// 【推荐】返回Agent支持的能力
	GetCapabilities() []Capability

	// HasCapability 检查是否具有特定能力
	// 【推荐】检查Agent是否支持特定能力
	HasCapability(capability Capability) bool
}

// Capability Agent能力类型
// 【推荐】定义Agent的各种能力
type Capability string

const (
	// CapabilityReasoning 推理能力
	CapabilityReasoning Capability = "reasoning"
	// CapabilityToolUse 工具使用能力
	CapabilityToolUse Capability = "tool_use"
	// CapabilityPlanning 规划能力
	CapabilityPlanning Capability = "planning"
	// CapabilityLearning 学习能力
	CapabilityLearning Capability = "learning"
	// CapabilityReflection 反思能力
	CapabilityReflection Capability = "reflection"
	// CapabilityMultiTask 多任务处理能力
	CapabilityMultiTask Capability = "multi_task"
	// CapabilityParallelExecution 并行执行能力
	CapabilityParallelExecution Capability = "parallel_execution"
	// CapabilityMemoryUse 记忆使用能力
	CapabilityMemoryUse Capability = "memory_use"
)

// String 返回能力的字符串表示
// 【必须】为枚举类型实现 String() 方法
func (c Capability) String() string {
	return string(c)
}

// IsValid 检查能力是否有效
// 【必须】验证能力枚举值的合法性
func (c Capability) IsValid() bool {
	switch c {
	case CapabilityReasoning, CapabilityToolUse, CapabilityPlanning, CapabilityLearning,
		CapabilityReflection, CapabilityMultiTask, CapabilityParallelExecution, CapabilityMemoryUse:
		return true
	default:
		return false
	}
}

// ActionExecutorManager 动作执行器管理器接口
// 【必须】管理多种类型的动作执行器
type ActionExecutorManager interface {
	// Execute 执行动作
	// 【必须】执行给定状态下的动作，返回观察结果
	Execute(ctx context.Context, state *ReActState) (*Observation, error)

	// RegisterExecutor 注册执行器
	// 【推荐】注册特定类型的执行器
	RegisterExecutor(actionType ReActActionType, executor ActionExecutor) error

	// UnregisterExecutor 注销执行器
	// 【推荐】注销特定类型的执行器
	UnregisterExecutor(actionType ReActActionType) error

	// GetExecutor 获取执行器
	// 【推荐】获取特定类型的执行器
	GetExecutor(actionType ReActActionType) (ActionExecutor, bool)
}

// 注意：ActionExecutor 接口定义在 executor.go 中

// DefaultActionExecutorManager 默认的动作执行器管理器
// 【必须】提供默认的执行器管理实现
type DefaultActionExecutorManager struct {
	// executors 执行器映射
	executors map[ReActActionType]ActionExecutor
	// executorsMutex 执行器访问互斥锁
	executorsMutex sync.RWMutex
	// logger 日志记录器
	logger *zap.Logger
	// config 配置对象
	config *ReActConfig
}

// NewDefaultActionExecutorManager 创建默认的动作执行器管理器
// 【必须】提供构造函数
func NewDefaultActionExecutorManager(logger *zap.Logger, config *ReActConfig) *DefaultActionExecutorManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultReActConfig()
	}

	return &DefaultActionExecutorManager{
		executors: make(map[ReActActionType]ActionExecutor),
		logger:    logger.Named("executor_manager"),
		config:    config,
	}
}

// Execute 执行动作
// 【必须】实现ActionExecutorManager接口
func (dem *DefaultActionExecutorManager) Execute(ctx context.Context, state *ReActState) (*Observation, error) {
	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 获取最后一个动作
	if len(state.Actions) == 0 {
		return nil, errors.NewValidationError("no actions to execute", nil)
	}

	lastAction := state.Actions[len(state.Actions)-1]

	// 获取对应的执行器
	dem.executorsMutex.RLock()
	executor, exists := dem.executors[lastAction.Type]
	dem.executorsMutex.RUnlock()

	if !exists {
		return NewObservation(lastAction.ID, false, nil, "no executor registered for action type: "+string(lastAction.Type)), nil
	}

	// 执行动作
	startTime := time.Now()
	observation, err := executor.Execute(ctx, lastAction, state)
	executionTime := time.Since(startTime)

	if err != nil {
		dem.logger.Error("action execution failed",
			zap.String("action_type", string(lastAction.Type)),
			zap.Error(err),
		)
		return observation, err
	}

	if observation != nil {
		observation.ExecutionTime = executionTime
	}

	dem.logger.Debug("action executed",
		zap.String("action_type", string(lastAction.Type)),
		zap.Duration("execution_time", executionTime),
	)

	return observation, nil
}

// RegisterExecutor 注册执行器
// 【推荐】实现ActionExecutorManager接口
func (dem *DefaultActionExecutorManager) RegisterExecutor(actionType ReActActionType, executor ActionExecutor) error {
	if executor == nil {
		return errors.NewValidationError("executor cannot be nil", nil)
	}

	if !actionType.IsValid() {
		return errors.NewValidationError("invalid action type", nil)
	}

	dem.executorsMutex.Lock()
	defer dem.executorsMutex.Unlock()

	dem.executors[actionType] = executor

	dem.logger.Info("executor registered",
		zap.String("action_type", string(actionType)),
		zap.String("executor_name", executor.Name()),
	)

	return nil
}

// UnregisterExecutor 注销执行器
// 【推荐】实现ActionExecutorManager接口
func (dem *DefaultActionExecutorManager) UnregisterExecutor(actionType ReActActionType) error {
	if !actionType.IsValid() {
		return errors.NewValidationError("invalid action type", nil)
	}

	dem.executorsMutex.Lock()
	defer dem.executorsMutex.Unlock()

	delete(dem.executors, actionType)

	dem.logger.Info("executor unregistered",
		zap.String("action_type", string(actionType)),
	)

	return nil
}

// GetExecutor 获取执行器
// 【推荐】实现ActionExecutorManager接口
func (dem *DefaultActionExecutorManager) GetExecutor(actionType ReActActionType) (ActionExecutor, bool) {
	dem.executorsMutex.RLock()
	defer dem.executorsMutex.RUnlock()

	executor, exists := dem.executors[actionType]
	return executor, exists
}

// ReActStateManager ReAct状态管理器
// 【必须】管理ReAct状态的生命周期
type ReActStateManager struct {
	// states 状态存储
	states map[string]*ReActState
	// statesMutex 状态访问互斥锁
	statesMutex sync.RWMutex
	// logger 日志记录器
	logger *zap.Logger
	// config 配置对象
	config *ReActConfig
}

// NewReActStateManager 创建ReAct状态管理器
// 【必须】提供构造函数确保必要初始化
func NewReActStateManager(logger *zap.Logger, config *ReActConfig) *ReActStateManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultReActConfig()
	}

	return &ReActStateManager{
		states: make(map[string]*ReActState),
		logger: logger.Named("state_manager"),
		config: config,
	}
}

// GetState 获取状态
// 【推荐】根据会话ID获取状态
func (rsm *ReActStateManager) GetState(sessionID string) *ReActState {
	if sessionID == "" {
		return nil
	}

	rsm.statesMutex.RLock()
	defer rsm.statesMutex.RUnlock()

	state, exists := rsm.states[sessionID]
	if !exists {
		return nil
	}

	// 返回状态的副本
	return state.Clone()
}

// SetState 设置状态
// 【推荐】根据会话ID设置状态
func (rsm *ReActStateManager) SetState(state *ReActState) error {
	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	if state.SessionID == "" {
		return errors.NewValidationError("session ID cannot be empty", nil)
	}

	rsm.statesMutex.Lock()
	defer rsm.statesMutex.Unlock()

	// 存储状态的副本
	rsm.states[state.SessionID] = state.Clone()

	rsm.logger.Debug("state updated",
		zap.String("session_id", state.SessionID),
		zap.String("status", state.Status.String()),
		zap.Int("iteration", state.IterationCount),
	)

	return nil
}

// DeleteState 删除状态
// 【推荐】根据会话ID删除状态
func (rsm *ReActStateManager) DeleteState(sessionID string) error {
	if sessionID == "" {
		return errors.NewValidationError("session ID cannot be empty", nil)
	}

	rsm.statesMutex.Lock()
	defer rsm.statesMutex.Unlock()

	delete(rsm.states, sessionID)

	rsm.logger.Debug("state deleted",
		zap.String("session_id", sessionID),
	)

	return nil
}

// ListStates 列出所有状态
// 【推荐】返回所有活跃的会话ID
func (rsm *ReActStateManager) ListStates() []string {
	rsm.statesMutex.RLock()
	defer rsm.statesMutex.RUnlock()

	ids := make([]string, 0, len(rsm.states))
	for id := range rsm.states {
		ids = append(ids, id)
	}

	return ids
}

// ClearStates 清除所有状态
// 【推荐】清除所有状态，用于清理资源
func (rsm *ReActStateManager) ClearStates() {
	rsm.statesMutex.Lock()
	defer rsm.statesMutex.Unlock()

	rsm.states = make(map[string]*ReActState)

	rsm.logger.Info("all states cleared")
}

// ReActMetrics ReAct Agent指标收集器
// 【推荐】收集和报告ReAct循环的指标
type ReActMetrics struct {
	// totalIterations 总迭代次数
	totalIterations int64
	// successfulIterations 成功迭代次数
	successfulIterations int64
	// failedIterations 失败迭代次数
	failedIterations int64
	// totalExecutionTime 总执行时间
	totalExecutionTime time.Duration
	// averageExecutionTime 平均执行时间
	averageExecutionTime time.Duration
	// actionsExecuted 已执行的动作数
	actionsExecuted int64
	// toolsCalled 调用的工具数
	toolsCalled int64
	// errorsEncountered 遇到的错误数
	errorsEncountered int64
	// metricsMutex 指标访问互斥锁
	metricsMutex sync.RWMutex
}

// NewReActMetrics 创建新的指标收集器
// 【推荐】提供构造函数
func NewReActMetrics() *ReActMetrics {
	return &ReActMetrics{
		totalIterations:      0,
		successfulIterations: 0,
		failedIterations:     0,
		totalExecutionTime:   0,
		averageExecutionTime: 0,
		actionsExecuted:      0,
		toolsCalled:          0,
		errorsEncountered:    0,
	}
}

// RecordIteration 记录一次迭代
// 【推荐】记录一次完整的ReAct循环迭代
func (rm *ReActMetrics) RecordIteration(success bool, duration time.Duration) {
	rm.metricsMutex.Lock()
	defer rm.metricsMutex.Unlock()

	rm.totalIterations++
	rm.totalExecutionTime += duration

	if success {
		rm.successfulIterations++
	} else {
		rm.failedIterations++
	}

	if rm.totalIterations > 0 {
		rm.averageExecutionTime = rm.totalExecutionTime / time.Duration(rm.totalIterations)
	}
}

// RecordAction 记录一次动作执行
// 【推荐】记录动作执行
func (rm *ReActMetrics) RecordAction() {
	rm.metricsMutex.Lock()
	defer rm.metricsMutex.Unlock()

	rm.actionsExecuted++
}

// RecordToolCall 记录一次工具调用
// 【推荐】记录工具调用
func (rm *ReActMetrics) RecordToolCall() {
	rm.metricsMutex.Lock()
	defer rm.metricsMutex.Unlock()

	rm.toolsCalled++
}

// RecordError 记录一次错误
// 【推荐】记录错误发生
func (rm *ReActMetrics) RecordError() {
	rm.metricsMutex.Lock()
	defer rm.metricsMutex.Unlock()

	rm.errorsEncountered++
}

// GetMetrics 获取当前指标
// 【推荐】返回当前指标快照
func (rm *ReActMetrics) GetMetrics() map[string]interface{} {
	rm.metricsMutex.RLock()
	defer rm.metricsMutex.RUnlock()

	return map[string]interface{}{
		"total_iterations":        rm.totalIterations,
		"successful_iterations":  rm.successfulIterations,
		"failed_iterations":      rm.failedIterations,
		"total_execution_time":   rm.totalExecutionTime.String(),
		"average_execution_time": rm.averageExecutionTime.String(),
		"actions_executed":        rm.actionsExecuted,
		"tools_called":            rm.toolsCalled,
		"errors_encountered":      rm.errorsEncountered,
		"success_rate":           rm.calculateSuccessRate(),
	}
}

// calculateSuccessRate 计算成功率
// 【内部】计算迭代成功率
func (rm *ReActMetrics) calculateSuccessRate() float64 {
	if rm.totalIterations == 0 {
		return 0.0
	}
	return float64(rm.successfulIterations) / float64(rm.totalIterations) * 100
}

// Reset 重置指标
// 【推荐】重置所有指标
func (rm *ReActMetrics) Reset() {
	rm.metricsMutex.Lock()
	defer rm.metricsMutex.Unlock()

	rm.totalIterations = 0
	rm.successfulIterations = 0
	rm.failedIterations = 0
	rm.totalExecutionTime = 0
	rm.averageExecutionTime = 0
	rm.actionsExecuted = 0
	rm.toolsCalled = 0
	rm.errorsEncountered = 0
}

// BaseReActAgent ReAct Agent基类
// 【推荐】提供基础实现减少重复代码
type BaseReActAgent struct {
	// id Agent唯一标识
	id string
	// name Agent名称
	name string
	// config Agent配置
	config *ReActConfig
	// logger 日志记录器
	logger *zap.Logger
	// stateManager 状态管理器
	stateManager *ReActStateManager
	// metrics 指标收集器
	metrics *ReActMetrics
	// creationTime 创建时间
	creationTime time.Time
	// capabilities 能力列表
	capabilities []Capability
	// capabilitiesMutex 能力访问互斥锁
	capabilitiesMutex sync.RWMutex
}

// ID 返回Agent ID
// 【必须】实现ID方法
func (bra *BaseReActAgent) ID() string {
	return bra.id
}

// Name 返回Agent名称
func (bra *BaseReActAgent) Name() string {
	return bra.name
}

// Type 返回Agent类型
func (bra *BaseReActAgent) Type() string {
	return "react"
}

// Config 返回Agent配置
// 【推荐】返回当前配置
func (bra *BaseReActAgent) Config() *ReActConfig {
	return bra.config.Clone()
}

// Logger 返回日志记录器
// 【推荐】返回日志记录器
func (bra *BaseReActAgent) Logger() *zap.Logger {
	return bra.logger
}

// Metrics 返回指标收集器
// 【推荐】返回指标收集器
func (bra *BaseReActAgent) Metrics() *ReActMetrics {
	return bra.metrics
}

// CreationTime 返回创建时间
// 【推荐】返回Agent创建时间
func (bra *BaseReActAgent) CreationTime() time.Time {
	return bra.creationTime
}

// GetCapabilities 获取能力列表
// 【推荐】返回Agent支持的能力列表
func (bra *BaseReActAgent) GetCapabilities() []Capability {
	bra.capabilitiesMutex.RLock()
	defer bra.capabilitiesMutex.RUnlock()

	caps := make([]Capability, len(bra.capabilities))
	copy(caps, bra.capabilities)

	return caps
}

// HasCapability 检查是否具有特定能力
// 【推荐】检查Agent是否支持特定能力
func (bra *BaseReActAgent) HasCapability(capability Capability) bool {
	bra.capabilitiesMutex.RLock()
	defer bra.capabilitiesMutex.RUnlock()

	for _, cap := range bra.capabilities {
		if cap == capability {
			return true
		}
	}

	return false
}

// addCapability 添加能力
// 【推荐】添加Agent能力
func (bra *BaseReActAgent) addCapability(capabilities ...Capability) {
	bra.capabilitiesMutex.Lock()
	defer bra.capabilitiesMutex.Unlock()

	bra.capabilities = append(bra.capabilities, capabilities...)
}

// DefaultReActAgent ReAct Agent默认实现
// 【必须】提供完整的ReAct Agent实现
type DefaultReActAgent struct {
	// BaseReActAgent 基础ReAct Agent字段
	*BaseReActAgent
	// planner 规划器
	planner PlanGenerator
	// executorManager 执行器管理器
	executorManager *DefaultActionExecutorManager
	// memoryIntegration 内存集成器
	memoryIntegration MemoryIntegration
	// thoughtChain 思考链
	thoughtChain *ThoughtChain
	// currentState 当前状态
	currentState *ReActState
	// stateMutex 状态访问互斥锁
	stateMutex sync.RWMutex
}

// NewDefaultReActAgent 创建默认ReAct Agent
// 【必须】提供构造函数
func NewDefaultReActAgent(
	name string,
	planner PlanGenerator,
	executor *DefaultActionExecutorManager,
	memoryIntegration MemoryIntegration,
	thoughtChain *ThoughtChain,
	logger *zap.Logger,
	config *ReActConfig,
) (*DefaultReActAgent, error) {
	if name == "" {
		return nil, errors.NewValidationError("agent name cannot be empty", nil)
	}

	if config == nil {
		config = DefaultReActConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, errors.WrapError(err, "config validation failed", nil)
	}

	// 创建基础Agent
	baseAgent := &BaseReActAgent{
		id:            config.AgentID,
		name:          name,
		config:        config,
		logger:        logger.Named("react_agent"),
		stateManager:  NewReActStateManager(logger, config),
		metrics:       NewReActMetrics(),
		creationTime:  time.Now().UTC(),
		capabilities:  make([]Capability, 0),
	}

	// 添加默认能力
	baseAgent.addCapability(
		CapabilityReasoning,
		CapabilityToolUse,
		CapabilityReflection,
		CapabilityMemoryUse,
	)

	if config.EnablePlanning {
		baseAgent.addCapability(CapabilityPlanning)
	}

	if config.EnableParallelExecution {
		baseAgent.addCapability(CapabilityParallelExecution)
	}

	agent := &DefaultReActAgent{
		BaseReActAgent:     baseAgent,
		planner:            planner,
		executorManager:    executor,
		memoryIntegration:  memoryIntegration,
		thoughtChain:       thoughtChain,
		currentState:       nil,
	}

	agent.logger.Info("ReAct agent created",
		zap.String("name", name),
		zap.Int("max_iterations", config.MaxIterations),
		zap.Bool("enable_planning", config.EnablePlanning),
	)

	return agent, nil
}

// Start 启动ReAct循环
// 【必须】实现ReActAgent接口
func (dra *DefaultReActAgent) Start(ctx context.Context, query string) (*ReActState, error) {
	if query == "" {
		return nil, errors.NewValidationError("query cannot be empty", nil)
	}

	// 创建新的状态
	state := NewReActState(
		dra.config.AgentID+"_"+dra.id,
		dra.id,
		query,
		dra.config.MaxIterations,
	)

	dra.stateMutex.Lock()
	dra.currentState = state
	dra.stateMutex.Unlock()

	dra.logger.Info("ReAct cycle started",
		zap.String("session_id", state.SessionID),
		zap.String("query", query),
	)

	return state, nil
}

// createThought 创建思考
// 【内部】创建新的思考对象
func (dra *DefaultReActAgent) createThought(state *ReActState) *Thought {
	return NewThought(
		"Processing query: "+state.Query,
		"Analyzing the current state and determining next action",
		0.8,
	)
}

// Step 执行一个ReAct步骤
// 【必须】实现ReActAgent接口
func (dra *DefaultReActAgent) Step(ctx context.Context, state *ReActState) (*ReActState, error) {
	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 检查是否达到最大迭代次数
	if state.IterationCount >= state.MaxIterations {
		state.Status = ReActStatusFailed
		dra.logger.Warn("max iterations reached",
			zap.Int("iteration", state.IterationCount),
			zap.Int("max_iterations", state.MaxIterations),
		)
		return state, nil
	}

	// 更新状态为思考中
	state.Status = ReActStatusThinking

	// 增加迭代计数
	state.IterationCount++

	// 执行思考步骤
	if dra.thoughtChain != nil && dra.config.EnableThoughtChain {
		thought := dra.createThought(state)
		processedThought, err := dra.thoughtChain.Process(ctx, thought, state)
		if err != nil {
			dra.metrics.RecordError()
			dra.logger.Error("thought processing failed",
				zap.Error(err),
			)
		} else if processedThought != nil {
			state.AddThought(processedThought)
		}
	}

	// 执行规划步骤
	if dra.planner != nil && dra.config.EnablePlanning && state.CurrentPlan == nil {
		plan, err := dra.planner.Generate(ctx, state.Query, state)
		if err != nil {
			dra.logger.Warn("plan generation failed, continuing without plan",
				zap.Error(err),
			)
		} else if plan != nil {
			state.CurrentPlan = plan
		}
	}

	// 执行动作步骤
	state.Status = ReActStatusActing

	if dra.executorManager != nil && dra.config.EnableToolExecution {
		observation, err := dra.executorManager.Execute(ctx, state)
		if err != nil {
			dra.metrics.RecordError()
			dra.logger.Error("action execution failed",
				zap.Error(err),
			)
			state.AddObservation(NewObservation("", false, nil, err.Error()))
		} else if observation != nil {
			state.AddObservation(observation)
			dra.metrics.RecordAction()
		}
	}

	// 执行反思步骤
	if dra.config.EnableReflection {
		state.Status = ReActStatusReflecting

		// 简单的反思逻辑：检查是否应该继续
		if len(state.Observations) > 0 {
			lastObservation := state.Observations[len(state.Observations)-1]
			if !lastObservation.Success {
				// 如果上一个动作失败，记录反思
				dra.logger.Debug("reflection: last action failed, may need to retry or adjust")
			}
		}
	}

	// 检查是否完成
	if len(state.Observations) > 0 {
		lastObs := state.Observations[len(state.Observations)-1]
		// 如果动作成功且没有更多需要处理的，可以标记为完成
		if lastObs.Success && state.IterationCount >= state.MaxIterations {
			state.Status = ReActStatusCompleted
		}
	}

	// 更新当前状态
	dra.stateMutex.Lock()
	dra.currentState = state
	dra.stateMutex.Unlock()

	dra.logger.Debug("ReAct step completed",
		zap.String("status", state.Status.String()),
		zap.Int("iteration", state.IterationCount),
	)

	return state, nil
}

// Run 完整运行ReAct循环
// 【必须】实现ReActAgent接口
func (dra *DefaultReActAgent) Run(ctx context.Context, query string) (*ReActState, error) {
	// 启动循环
	state, err := dra.Start(ctx, query)
	if err != nil {
		return nil, errors.WrapError(err, "failed to start ReAct cycle", nil)
	}

	// 执行循环直到完成或失败
	for {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			state.Status = ReActStatusFailed
			return state, errors.NewValidationError("context cancelled", nil)
		default:
		}

		// 执行一步
		state, err = dra.Step(ctx, state)
		if err != nil {
			dra.metrics.RecordIteration(false, 0)
			return state, errors.WrapError(err, "step execution failed", nil)
		}

		// 检查是否完成
		if state.Status.IsTerminal() {
			break
		}

		// 额外检查最大迭代
		if state.IterationCount >= state.MaxIterations {
			state.Status = ReActStatusCompleted
			break
		}
	}

	// 记录迭代结果
	dra.metrics.RecordIteration(state.Status == ReActStatusCompleted, 0)

	dra.logger.Info("ReAct cycle completed",
		zap.String("status", state.Status.String()),
		zap.Int("total_iterations", state.IterationCount),
	)

	return state, nil
}

// GetState 获取当前状态
// 【必须】实现ReActAgent接口
func (dra *DefaultReActAgent) GetState() *ReActState {
	dra.stateMutex.RLock()
	defer dra.stateMutex.RUnlock()

	if dra.currentState == nil {
		return nil
	}

	return dra.currentState.Clone()
}

// SetState 设置当前状态
// 【必须】实现ReActAgent接口
func (dra *DefaultReActAgent) SetState(state *ReActState) error {
	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	dra.stateMutex.Lock()
	defer dra.stateMutex.Unlock()

	dra.currentState = state.Clone()

	dra.logger.Debug("state updated",
		zap.String("session_id", state.SessionID),
	)

	return nil
}
