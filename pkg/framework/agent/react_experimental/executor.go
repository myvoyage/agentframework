//go:build experimental

// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package react

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent"
	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/errors"
)

// ActionExecutor 动作执行器接口
// 【必须】定义动作执行的标准接口
type ActionExecutor interface {
	// Execute 执行动作
	// 【必须】接收上下文和动作对象，返回执行结果和可能的错误
	Execute(ctx context.Context, action *Action, state *ReActState) (*Observation, error)
	// CanExecute 检查是否能执行指定类型的动作
	// 【必须】验证执行器是否支持特定动作类型
	CanExecute(actionType ReActActionType) bool
	// Validate 验证执行器配置
	// 【必须】验证执行器自身的配置有效性
	Validate() error
	// Name 返回执行器名称
	// 【必须】提供执行器标识
	Name() string
}

// BaseActionExecutor 动作执行器基类
// 【推荐】提供基础实现减少重复代码
type BaseActionExecutor struct {
	name   string
	logger *zap.Logger
	config *ReActConfig
}

// NewBaseActionExecutor 创建基础动作执行器
// 【必须】提供构造函数确保必要字段初始化
func NewBaseActionExecutor(name string, logger *zap.Logger, config *ReActConfig) *BaseActionExecutor {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = NewReActConfig()
	}

	return &BaseActionExecutor{
		name:   name,
		logger: logger.With(zap.String("executor", name)),
		config: config,
	}
}

// Name 返回执行器名称
// 【必须】实现 ActionExecutor 接口的 Name 方法
func (bae *BaseActionExecutor) Name() string {
	return bae.name
}

// Validate 验证执行器配置
// 【必须】实现 ActionExecutor 接口的 Validate 方法
func (bae *BaseActionExecutor) Validate() error {
	if bae.name == "" {
		return errors.NewValidationError("executor name cannot be empty", nil)
	}

	if bae.config == nil {
		return errors.NewValidationError("config cannot be nil", nil)
	}

	return bae.config.Validate()
}

// CanExecute 基础实现，子类可以重写
// 【推荐】默认不支持任何动作类型，子类必须显式实现
func (bae *BaseActionExecutor) CanExecute(actionType ReActActionType) bool {
	return false
}

// ToolActionExecutor 工具动作执行器
// 【必须】专门执行工具调用的动作执行器
type ToolActionExecutor struct {
	BaseActionExecutor
	// ToolRegistry 工具注册表
	ToolRegistry map[string]tool.BaseTool
	// ExecutionTimeout 执行超时时间
	ExecutionTimeout time.Duration
	// MaxConcurrentExecutions 最大并发执行数
	MaxConcurrentExecutions int
	// executionSemaphore 执行信号量
	executionSemaphore chan struct{}
	// executionMutex 执行互斥锁
	executionMutex sync.RWMutex
}

// NewToolActionExecutor 创建工具动作执行器
// 【必须】提供构造函数确保必要配置
func NewToolActionExecutor(logger *zap.Logger, config *ReActConfig, toolRegistry map[string]tool.BaseTool) *ToolActionExecutor {
	executor := &ToolActionExecutor{
		BaseActionExecutor:      *NewBaseActionExecutor("tool_action_executor", logger, config),
		ToolRegistry:            toolRegistry,
		ExecutionTimeout:        config.ToolTimeout,
		MaxConcurrentExecutions: 5, // 默认最大并发数
		executionSemaphore:      make(chan struct{}, 5),
	}

	// 【必须】设置合理的默认值
	if executor.ExecutionTimeout <= 0 {
		executor.ExecutionTimeout = 30 * time.Second
	}

	if executor.MaxConcurrentExecutions <= 0 {
		executor.MaxConcurrentExecutions = 5
	}

	// 重新创建合适大小的信号量
	executor.executionSemaphore = make(chan struct{}, executor.MaxConcurrentExecutions)

	if executor.ToolRegistry == nil {
		executor.logger.Warn("tool registry is nil, tool execution will be limited")
	}

	return executor
}

// CanExecute 检查是否能执行工具动作
// 【必须】实现 ActionExecutor 接口的 CanExecute 方法
func (tae *ToolActionExecutor) CanExecute(actionType ReActActionType) bool {
	return actionType == ActionTypeTool
}

// Execute 执行工具动作
// 【必须】实现 ActionExecutor 接口的 Execute 方法
func (tae *ToolActionExecutor) Execute(ctx context.Context, action *Action, state *ReActState) (*Observation, error) {
	if action == nil {
		return nil, errors.NewValidationError("action cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】验证动作类型
	if !tae.CanExecute(action.Type) {
		return nil, errors.NewValidationError(
			"executor cannot handle action type",
			map[string]interface{}{
				"action_type": action.Type,
				"executor":    tae.Name(),
			},
		)
	}

	// 【必须】记录执行开始
	tae.logger.Debug("starting tool action execution",
		zap.String("action_id", action.ID),
		zap.String("tool_name", action.Name),
		zap.Any("parameters", action.Parameters),
		zap.String("session_id", state.SessionID),
	)

	// 获取执行许可（限制并发）
	select {
	case tae.executionSemaphore <- struct{}{}:
		defer func() { <-tae.executionSemaphore }()
	case <-ctx.Done():
		return nil, errors.New(ErrorTypeExecution, "execution cancelled due to context cancellation")
	case <-time.After(tae.ExecutionTimeout):
		return nil, errors.New(ErrorTypeTimeout, "execution timeout waiting for semaphore")
	}

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, tae.ExecutionTimeout)
	defer cancel()

	var observation *Observation
	var executionErr error

	// 使用互斥锁保护执行过程
	tae.executionMutex.Lock()
	defer tae.executionMutex.Unlock()

	// 查找工具
	tool, err := tae.findTool(execCtx, action.Name)
	if err != nil {
		observation = NewObservation(action.ID, false, nil, err)
		executionErr = err
	} else {
		// 验证工具
		if err := tool.Validate(); err != nil {
			observation = NewObservation(action.ID, false, nil,
				errors.WrapError(err, "tool validation failed", map[string]interface{}{"tool_name": action.Name}))
			executionErr = err
		} else {
			// 准备工具参数
			params, err := tae.prepareToolParameters(execCtx, tool, action.Parameters)
			if err != nil {
				observation = NewObservation(action.ID, false, nil, err)
				executionErr = err
			} else {
				// 执行工具
				result, err := tool.Execute(execCtx, params)
				if err != nil {
					observation = NewObservation(action.ID, false, nil, err)
					executionErr = err
				} else {
					// 创建成功观察
					observation = NewObservation(action.ID, true, result, nil)

					// 记录成功日志
					tae.logger.Info("tool action executed successfully",
						zap.String("action_id", action.ID),
						zap.String("tool_name", action.Name),
						zap.Duration("execution_time", time.Since(startTime)),
					)
				}
			}
		}
	}

	// 处理异常情况
	if observation == nil {
		observation = NewObservation(action.ID, false, nil,
			errors.NewInternalError("unexpected error during tool execution", nil))
	}

	// 更新动作上下文关联
	for _, contextID := range action.AssociatedContexts {
		if tae.config.Tracker != nil {
			if err := tae.config.Tracker.AssociateContext(ctx, contextID, "action", action.ID); err != nil {
				tae.logger.Warn("failed to associate action context",
					zap.Error(err),
					zap.String("action_id", action.ID),
					zap.String("context_id", contextID),
				)
			}
		}
	}

// 更新观察结果上下文关联
for _, contextID := range observation.AssociatedContexts {
	if tae.config.Tracker != nil {
		if err := tae.config.Tracker.AssociateContext(ctx, contextID, "observation", observation.ID); err != nil {
			tae.logger.Warn("failed to associate observation context",
				zap.Error(err),
				zap.String("observation_id", observation.ID),
				zap.String("context_id", contextID),
			)
		}
	}
}

return observation, executionErr
}

// findTool 查找指定的工具
// 【推荐】从工具注册表中查找工具
func (tae *ToolActionExecutor) findTool(ctx context.Context, toolName string) (tool.BaseTool, error) {
	if tae.ToolRegistry == nil {
		return nil, errors.NewValidationError("tool registry is nil", nil)
	}

	// 【必须】从注册表获取工具
	tool, err := tae.ToolRegistry.GetTool(toolName)
	if err != nil {
		return nil, errors.WrapError(err, "failed to get tool from registry", map[string]interface{}{
			"tool_name": toolName,
		})
	}

	if tool == nil {
		return nil, errors.NewValidationError("tool not found in registry", map[string]interface{}{
			"tool_name": toolName,
		})
	}

	return tool, nil
}

// prepareToolParameters 准备工具参数
// 【推荐】将动作参数转换为工具可接受的格式
func (tae *ToolActionExecutor) prepareToolParameters(ctx context.Context, tool tool.BaseTool, actionParams map[string]interface{}) (map[string]interface{}, error) {
	if actionParams == nil {
		return make(map[string]interface{}), nil
	}

	// 【必须】这里应该根据工具的具体要求转换参数格式
	// 暂时直接返回原始参数，实际实现可能需要复杂的转换逻辑
	
	// 验证必需参数
	requiredParams := tae.getRequiredParameters(tool)
	for paramName := range requiredParams {
		if _, exists := actionParams[paramName]; !exists {
			return nil, errors.NewValidationError(
				fmt.Sprintf("required parameter '%s' missing", paramName),
				map[string]interface{}{"parameter": paramName, "tool_name": tool.Name()},
			)
		}
	}

	// 过滤和验证参数类型
	filteredParams := make(map[string]interface{})
	for key, value := range actionParams {
		if tae.isValidParameter(key, value) {
			filteredParams[key] = value
		} else {
			tae.logger.Warn("invalid parameter value, skipping",
				zap.String("parameter", key),
				zap.Any("value", value),
			)
		}
	}

	return filteredParams, nil
}

// getRequiredParameters 获取工具的必需参数
// 【推荐】提取工具的必需参数信息
func (tae *ToolActionExecutor) getRequiredParameters(tool tool.BaseTool) map[string]bool {
	// 【必须】这里应该根据工具的具体实现获取必需参数
	// 暂时返回空map作为示例
	return make(map[string]bool)
}

// isValidParameter 验证参数有效性
// 【推荐】验证单个参数的有效性
func (tae *ToolActionExecutor) isValidParameter(name string, value interface{}) bool {
	// 基本的有效性检查
	if name == "" {
		return false
	}

	// 检查nil值
	if value == nil {
		return false
	}

	// 根据参数名进行特定验证
	switch name {
	case "path", "file_path", "directory":
		// 路径参数应该是字符串且非空
		if str, ok := value.(string); ok {
			return str != ""
		}
		return false
	case "content", "data", "text":
		// 内容参数应该是字符串
		_, ok := value.(string)
		return ok
	case "recursive", "overwrite", "enabled":
		// 布尔参数
		_, ok := value.(bool)
		return ok
	case "max_results", "depth", "limit":
		// 数值参数
		if num, ok := value.(int); ok {
			return num >= 0
		}
		if num, ok := value.(float64); ok {
			return num >= 0
		}
		return false
	}

	return true
}

// SearchActionExecutor 搜索动作执行器
// 【必须】执行搜索类动作的专用执行器
type SearchActionExecutor struct {
	BaseActionExecutor
	// SearchProviders 搜索提供者列表
	SearchProviders []SearchProvider
	// MaxResults 最大搜索结果数
	MaxResults int
}

// SearchProvider 搜索提供者接口
// 【必须】定义搜索服务的标准接口
type SearchProvider interface {
	// Search 执行搜索
	// 【必须】接收搜索查询和参数，返回搜索结果
	Search(ctx context.Context, query string, params map[string]interface{}) ([]SearchResult, error)
	// Name 返回提供者名称
	// 【必须】提供提供者标识
	Name() string
	// Validate 验证提供者配置
	// 【必须】验证提供者自身配置
	Validate() error
}

// SearchResult 搜索结果
// 【推荐】封装搜索结果的结构
type SearchResult struct {
	// Title 结果标题
	Title string `json:"title"`
	// URL 结果链接
	URL string `json:"url"`
	// Snippet 结果摘要
	Snippet string `json:"snippet"`
	// RelevanceScore 相关性评分
	RelevanceScore float64 `json:"relevance_score"`
	// Source 数据源
	Source string `json:"source"`
	// Metadata 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewSearchActionExecutor 创建搜索动作执行器
// 【必须】提供构造函数确保必要配置
func NewSearchActionExecutor(logger *zap.Logger, config *ReActConfig, searchProviders []SearchProvider) *SearchActionExecutor {
	executor := &SearchActionExecutor{
		BaseActionExecutor: *NewBaseActionExecutor("search_action_executor", logger, config),
		SearchProviders:    searchProviders,
		MaxResults:         10,
	}

	// 【必须】设置合理的默认值
	if executor.MaxResults <= 0 {
		executor.MaxResults = 10
	}

	if len(executor.SearchProviders) == 0 {
		executor.logger.Warn("no search providers configured, search functionality will be limited")
	}

	return executor
}

// CanExecute 检查是否能执行搜索动作
// 【必须】实现 ActionExecutor 接口的 CanExecute 方法
func (sae *SearchActionExecutor) CanExecute(actionType ReActActionType) bool {
	return actionType == ActionTypeSearch
}

// Execute 执行搜索动作
// 【必须】实现 ActionExecutor 接口的 Execute 方法
func (sae *SearchActionExecutor) Execute(ctx context.Context, action *Action, state *ReActState) (*Observation, error) {
	if action == nil {
		return nil, errors.NewValidationError("action cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】验证动作类型
	if !sae.CanExecute(action.Type) {
		return nil, errors.NewValidationError(
			"executor cannot handle action type",
			map[string]interface{}{
				"action_type": action.Type,
				"executor":    sae.Name(),
			},
		)
	}

	// 【必须】记录执行开始
	sae.logger.Debug("starting search action execution",
		zap.String("action_id", action.ID),
		zap.String("search_query", action.Name),
		zap.Any("parameters", action.Parameters),
		zap.String("session_id", state.SessionID),
	)

	// 提取搜索查询和参数
	query := action.Name
	searchParams := action.Parameters

	// 限制最大结果数
	if maxResultsParam, exists := searchParams["max_results"]; exists {
		if maxResults, ok := maxResultsParam.(int); ok && maxResults > 0 {
			if maxResults < sae.MaxResults {
				sae.MaxResults = maxResults
			}
		}
	}

	// 执行搜索
	results, err := sae.performSearch(ctx, query, searchParams)
	if err != nil {
		observation := NewObservation(action.ID, false, nil, err)
		return observation, nil
	}

	// 限制结果数量
	if len(results) > sae.MaxResults {
		results = results[:sae.MaxResults]
	}

	// 创建成功观察
	observation := NewObservation(action.ID, true, results, nil)

	// 记录成功日志
	sae.logger.Info("search action executed successfully",
		zap.String("action_id", action.ID),
		zap.String("search_query", query),
		zap.Int("results_count", len(results)),
	)

	return observation, nil
}

// performSearch 执行搜索逻辑
// 【推荐】封装搜索的具体实现
func (sae *SearchActionExecutor) performSearch(ctx context.Context, query string, params map[string]interface{}) ([]SearchResult, error) {
	if query == "" {
		return nil, errors.NewValidationError("search query cannot be empty", nil)
	}

	var allResults []SearchResult
	var lastError error

	// 尝试所有搜索提供者
	for _, provider := range sae.SearchProviders {
		if provider == nil {
			continue
		}

		select {
		case <-ctx.Done():
			return nil, errors.New(ErrorTypeExecution, "search cancelled")
		default:
			results, err := provider.Search(ctx, query, params)
			if err != nil {
				sae.logger.Warn("search provider failed",
					zap.Error(err),
					zap.String("provider_name", provider.Name()),
					zap.String("query", query),
				)
				lastError = err
				continue
			}

		if len(results) > 0 {
			allResults = append(allResults, results...)
		}
	}
}

// 如果没有任何提供者成功，返回错误
if len(allResults) == 0 && lastError != nil {
	return nil, errors.WrapError(lastError, "all search providers failed", map[string]interface{}{
		"query": query,
	})
}

// 按相关性评分排序
if len(allResults) > 1 {
	for i := 0; i < len(allResults)-1; i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[i].RelevanceScore < allResults[j].RelevanceScore {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}
}

return allResults, nil
}

// ReflectActionExecutor 反思动作执行器
// 【必须】执行反思类动作的专用执行器
type ReflectActionExecutor struct {
	BaseActionExecutor
	// ReflectionDepth 反思深度
	ReflectionDepth int
	// EnableCritique 是否启用批评性反思
	EnableCritique bool
}

// NewReflectActionExecutor 创建反思动作执行器
// 【必须】提供构造函数确保必要配置
func NewReflectActionExecutor(logger *zap.Logger, config *ReActConfig) *ReflectActionExecutor {
	executor := &ReflectActionExecutor{
		BaseActionExecutor: *NewBaseActionExecutor("reflect_action_executor", logger, config),
		ReflectionDepth:    2,
		EnableCritique:     true,
	}

	// 【必须】设置合理的默认值
	if executor.ReflectionDepth <= 0 {
		executor.ReflectionDepth = 2
	}

	return executor
}

// CanExecute 检查是否能执行反思动作
// 【必须】实现 ActionExecutor 接口的 CanExecute 方法
func (rae *ReflectActionExecutor) CanExecute(actionType ReActActionType) bool {
	return actionType == ActionTypeReflect
}

// Execute 执行反思动作
// 【必须】实现 ActionExecutor 接口的 Execute 方法
func (rae *ReflectActionExecutor) Execute(ctx context.Context, action *Action, state *ReActState) (*Observation, error) {
	if action == nil {
		return nil, errors.NewValidationError("action cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】验证动作类型
	if !rae.CanExecute(action.Type) {
		return nil, errors.NewValidationError(
			"executor cannot handle action type",
			map[string]interface{}{
				"action_type": action.Type,
				"executor":    rae.Name(),
			},
		)
	}

	// 【必须】记录执行开始
	rae.logger.Debug("starting reflect action execution",
		zap.String("action_id", action.ID),
		zap.String("reflection_focus", action.Name),
		zap.Any("parameters", action.Parameters),
		zap.String("session_id", state.SessionID),
	)

	// 执行反思分析
	reflection, err := rae.performReflection(ctx, action, state)
	if err != nil {
		observation := NewObservation(action.ID, false, nil, err)
		return observation, nil
	}

	// 创建成功观察
	observation := NewObservation(action.ID, true, reflection, nil)

	// 记录成功日志
	rae.logger.Info("reflect action executed successfully",
		zap.String("action_id", action.ID),
		zap.String("reflection_focus", action.Name),
		zap.Int("insights_count", len(reflection.Insights)),
	)

	return observation, nil
}

// Reflection 反思结果
// 【推荐】封装反思分析的结极
type Reflection struct {
	// Focus 反思焦点
	Focus string `json:"focus"`
	// Insights 洞察列表
	Insights []Insight `json:"insights"`
	// Improvements 改进建议
	Improvements []Improvement `json:"improvements"`
	// OverallAssessment 整体评估
	OverallAssessment string `json:"overall_assessment"`
	// ConfidenceLevel 置信度等级
	ConfidenceLevel float64 `json:"confidence_level"`
	// GeneratedAt 生成时间
	GeneratedAt time.Time `json:"generated_at"`
}

// Insight 洞察
// 【推荐】封装单个洞察
type Insight struct {
	// Description 洞察描述
	Description string `json:"description"`
	// Category 类别
	Category string `json:"category"`
	// Importance 重要性评分
	Importance float64 `json:"importance"`
	// Evidence 支撑证据
	Evidence []string `json:"evidence"`
}

// Improvement 改进建议
// 【推荐】封装改进建议
type Improvement struct {
	// Area 改进领域
	Area string `json:"area"`
	// Suggestion 具体建议
	Suggestion string `json:"suggestion"`
	// ExpectedImpact 预期影响
	ExpectedImpact string `json:"expected_impact"`
	// Effort 所需努力
	Effort string `json:"effort"`
}

// performReflection 执行反思分析
// 【推荐】封装反思的具体实现
func (rae *ReflectActionExecutor) performReflection(ctx context.Context, action *Action, state *ReActState) (*Reflection, error) {
	reflection := &Reflection{
		Focus:          action.Name,
		Insights:       make([]Insight, 0),
		Improvements:   make([]Improvement, 0),
		GeneratedAt:    time.Now().UTC(),
		ConfidenceLevel: 0.8,
	}

	// 分析思考过程
	thoughtInsights := rae.analyzeThoughts(state.Thoughts)
	reflection.Insights = append(reflection.Insights, thoughtInsights...)

	// 分析动作执行
	actionInsights := rae.analyzeActions(state.Actions, state.Observations)
	reflection.Insights = append(reflection.Insights, actionInsights...)

	// 分析目标达成情况
	goalInsights := rae.analyzeGoalProgress(action.Name, state)
	reflection.Insights = append(reflection.Insights, goalInsights...)

	// 生成改进建议
	improvements := rae.generateImprovements(reflection.Insights, state)
	reflection.Improvements = improvements

	// 生成整体评估
	reflection.OverallAssessment = rae.generateOverallAssessment(reflection)

	// 调整置信度
	reflection.ConfidenceLevel = rae.calculateConfidence(reflection, state)

	return reflection, nil
}

// analyzeThoughts 分析思考过程
// 【推荐】分析思考步骤的质量和效果
func (rae *ReflectActionExecutor) analyzeThoughts(thoughts []*Thought) []Insight {
	insights := make([]Insight, 0)

	if len(thoughts) == 0 {
		insights = append(insights, Insight{
			Description: "No thoughts recorded yet",
			Category:    "process",
			Importance:  0.5,
		})
		return insights
	}

	// 分析思考深度
	avgConfidence := 0.0
	totalThoughts := len(thoughts)
	for _, thought := range thoughts {
		avgConfidence += thought.Confidence
	}
	avgConfidence /= float64(totalThoughts)

	if avgConfidence < 0.5 {
		insights = append(insights, Insight{
			Description: fmt.Sprintf("Low average confidence (%.2f) suggests uncertainty in reasoning", avgConfidence),
			Category:    "reasoning_quality",
			Importance:  0.8,
			Evidence:    []string{fmt.Sprintf("Average confidence across %d thoughts", totalThoughts)},
		})
	}

	// 分析思考连贯性
	if totalThoughts > 1 {
		consistencyScore := rae.calculateThoughtConsistency(thoughts)
		if consistencyScore < 0.6 {
			insights = append(insights, Insight{
				Description: "Inconsistent reasoning pattern detected across thoughts",
				Category:    "reasoning_consistency",
				Importance:  0.7,
				Evidence:    []string{"Multiple thoughts show divergent reasoning paths"},
			})
		}
	}

	return insights
}

// analyzeActions 分析动作执行
// 【推荐】分析动作执行的效果和质量
func (rae *ReflectActionExecutor) analyzeActions(actions []*Action, observations []*Observation) []Insight {
	insights := make([]Insight, 0)

	if len(actions) == 0 {
		return insights
	}

	// 计算成功率
	successCount := 0
	for _, obs := range observations {
		if obs != nil && obs.Success {
			successCount++
		}
	}

	successRate := 0.0
	if len(observations) > 0 {
		successRate = float64(successCount) / float64(len(observations))
	}

	if successRate < 0.7 {
		insights = append(insights, Insight{
			Description: fmt.Sprintf("Low action success rate (%.1f%%) indicates execution challenges", successRate*100),
			Category:    "execution_effectiveness",
			Importance:  0.9,
			Evidence:    []string{fmt.Sprintf("%d successful out of %d attempts", successCount, len(observations))},
		})
	}

	return insights
}

// analyzeGoalProgress 分析目标进展
// 【推荐】分析相对于目标的进展情况
func (rae *ReflectActionExecutor) analyzeGoalProgress(focus string, state *ReActState) []Insight {
	insights := make([]Insight, 0)

	// 基于迭代次数分析进展
	if state.IterationCount > state.MaxIterations/2 {
		insights = append(insights, Insight{
			Description: "Approaching iteration limit, consider wrapping up or adjusting approach",
			Category:    "progress_management",
			Importance:  0.8,
			Evidence:    []string{fmt.Sprintf("Iteration %d of %d", state.IterationCount, state.MaxIterations)},
		})
	}

	return insights
}

// generateImprovements 生成改进建议
// 【推荐】基于洞察生成具体的改进建议
func (rae *ReflectActionExecutor) generateImprovements(insights []Insight, state *ReActState) []Improvement {
	improvements := make([]Improvement, 0)

	for _, insight := range insights {
		switch insight.Category {
		case "reasoning_quality":
			improvements = append(improvements, Improvement{
				Area:            "reasoning",
				Suggestion:      "Gather more information before forming conclusions",
				ExpectedImpact:  "Higher confidence in decisions",
				Effort:          "medium",
			})
		case "execution_effectiveness":
			improvements = append(improvements, Improvement{
				Area:            "execution",
				Suggestion:      "Add validation steps and error handling for actions",
				ExpectedImpact:  "Improved success rate",
				Effort:          "high",
			})
		case "progress_management":
			improvements = append(improvements, Improvement{
				Area:            "planning",
				Suggestion:      "Prioritize remaining high-impact actions",
				ExpectedImpact:  "More efficient use of remaining iterations",
				Effort:          "low",
			})
		}
	}

	return improvements
}

// generateOverallAssessment 生成整体评估
// 【推荐】基于反思结果生成整体评估
func (rae *ReflectActionExecutor) generateOverallAssessment(reflection *Reflection) string {
	positiveCount := 0
	concernCount := 0

	for _, insight := range reflection.Insights {
		if insight.Importance >= 0.7 {
			if strings.Contains(strings.ToLower(insight.Description), "low") || strings.Contains(strings.ToLower(insight.Description), "challenge") {
				concernCount++
			} else {
				positiveCount++
			}
		}
	}

	if concernCount > positiveCount {
		return "Significant concerns identified that require attention"
	} else if concernCount > 0 {
		return "Mixed results with some areas needing improvement"
	} else {
		return "Generally positive progress with minor optimization opportunities"
	}
}

// calculateConfidence 计算置信度
// 【推荐】基于反思质量计算置信度
func (rae *ReflectActionExecutor) calculateConfidence(reflection *Reflection, state *ReActState) float64 {
	baseConfidence := 0.7

	// 根据洞察数量调整
	insightBonus := float64(len(reflection.Insights)) * 0.02
	baseConfidence += insightBonus

	// 根据改进建议数量调整
	improvementBonus := float64(len(reflection.Improvements)) * 0.01
	baseConfidence += improvementBonus

	// 根据整体评估调整
	if strings.Contains(strings.ToLower(reflection.OverallAssessment), "significant") {
		baseConfidence -= 0.2
	} else if strings.Contains(strings.ToLower(reflection.OverallAssessment), "generally positive") {
		baseConfidence += 0.1
	}

	// 确保在合理范围内
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	} else if baseConfidence < 0.0 {
		baseConfidence = 0.0
	}

	return baseConfidence
}

// calculateThoughtConsistency 计算思考一致性
// 【推荐】计算思考步骤之间的一致性得分
func (rae *ReflectActionExecutor) calculateThoughtConsistency(thoughts []*Thought) float64 {
	if len(thoughts) < 2 {
		return 1.0
	}

	// 简单的实现：基于置信度的方差
	meanConfidence := 0.0
	for _, thought := range thoughts {
		meanConfidence += thought.Confidence
	}
	meanConfidence /= float64(len(thoughts))

	variance := 0.0
	for _, thought := range thoughts {
		diff := thought.Confidence - meanConfidence
		variance += diff * diff
	}
	variance /= float64(len(thoughts))

	// 方差越小，一致性越高
	consistency := 1.0 / (1.0 + variance)
	return consistency
}

// StandardActionExecutorManager 标准动作执行器管理器实现
// 【必须】管理多个动作执行器的执行
type StandardActionExecutorManager struct {
	// executors 执行器集合
	executors []ActionExecutor
	// logger 日志记录器
	logger *zap.Logger
}

// NewActionExecutorManager 创建动作执行器管理器
// 【必须】提供构造函数确保必要初始化
func NewActionExecutorManager(logger *zap.Logger) *StandardActionExecutorManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &StandardActionExecutorManager{
		executors: make([]ActionExecutor, 0),
		logger:    logger.Named("action_executor_manager"),
	}
}

// AddExecutor 添加动作执行器
// 【必须】提供执行器管理功能
func (aem *ActionExecutorManager) AddExecutor(executor ActionExecutor) error {
	if executor == nil {
		return errors.NewValidationError("executor cannot be nil", nil)
	}

	if err := executor.Validate(); err != nil {
		return errors.WrapError(err, "executor validation failed", nil)
	}

aem.executors = append(aem.executors, executor)
	aem.logger.Debug("added action executor",
		zap.String("executor_name", executor.Name()),
		zap.Int("executors_count", len(aem.executors)),
	)

	return nil
}

// ExecuteAction 执行动作
// 【必须】选择合适的执行器执行动作
func (aem *ActionExecutorManager) ExecuteAction(ctx context.Context, action *Action, state *ReActState) (*Observation, error) {
	if action == nil {
		return nil, errors.NewValidationError("action cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录执行开始
	aem.logger.Debug("starting action execution via manager",
		zap.String("action_id", action.ID),
		zap.String("action_type", action.Type.String()),
		zap.String("action_name", action.Name),
		zap.String("session_id", state.SessionID),
	)

	// 查找合适的执行器
	executor, err := aem.findExecutor(action.Type)
	if err != nil {
		observation := NewObservation(action.ID, false, nil, err)
		return observation, nil
	}

	// 执行动作
	observation, err := executor.Execute(ctx, action, state)
	if err != nil {
		aem.logger.Error("action execution failed",
			zap.Error(err),
			zap.String("action_id", action.ID),
			zap.String("executor_name", executor.Name()),
		)
		// 【必须】执行失败仍返回观察结果
		if observation == nil {
			observation = NewObservation(action.ID, false, nil, err)
		}
		return observation, nil
	}

	// 【必须】记录执行成功
	aem.logger.Info("action execution completed via manager",
		zap.String("action_id", action.ID),
		zap.String("action_type", action.Type.String()),
		zap.String("executor_name", executor.Name()),
		zap.Bool("success", observation.Success),
	)

	return observation, nil
}

// findExecutor 查找合适的执行器
// 【推荐】根据动作类型查找匹配的执行器
func (aem *ActionExecutorManager) findExecutor(actionType ReActActionType) (ActionExecutor, error) {
	for _, executor := range aem.executors {
		if executor.CanExecute(actionType) {
			return executor, nil
		}
	}

	return nil, errors.NewValidationError(
		"no suitable executor found for action type",
		map[string]interface{}{"action_type": actionType},
	)
}

// GetExecutor 根据名称获取执行器
// 【推荐】按名称查找特定执行器
func (aem *ActionExecutorManager) GetExecutor(name string) ActionExecutor {
	for _, executor := range aem.executors {
		if executor.Name() == name {
			return executor
		}
	}
	return nil
}

// ExecutorCount 返回执行器数量
// 【推荐】提供管理器状态查询
func (aem *ActionExecutorManager) ExecutorCount() int {
	return len(aem.executors)
}

// Validate 验证执行器管理器配置
// 【必须】验证管理器的完整性
func (aem *ActionExecutorManager) Validate() error {
	if aem.executors == nil {
		return errors.NewValidationError("executors slice cannot be nil", nil)
	}

	for i, executor := range aem.executors {
		if executor == nil {
			return errors.NewValidationError(
				fmt.Sprintf("executor at index %d cannot be nil", i),
				map[string]interface{}{"index": i},
			)
		}

		if err := executor.Validate(); err != nil {
			return errors.WrapError(err, fmt.Sprintf("executor at index %d validation failed", i), nil)
		}
	}

	return nil
}