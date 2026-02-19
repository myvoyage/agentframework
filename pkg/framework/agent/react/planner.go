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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent"
	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/framework/planning"
	"AgentFramework/pkg/errors"
)

// PlanGenerator 规划生成器接口
// 【必须】定义规划生成的标准接口
type PlanGenerator interface {
	// Generate 生成规划
	// 【必须】基于查询和上下文生成执行规划
	Generate(ctx context.Context, query string, state *ReActState) (*Plan, error)
	// Refine 优化规划
	// 【必须】基于观察结果优化现有规划
	Refine(ctx context.Context, plan *Plan, observations []*Observation, state *ReActState) (*Plan, error)
	// Validate 验证规划生成器配置
	// 【必须】验证生成器自身配置的有效性
	Validate() error
	// Name 返回生成器名称
	// 【必须】提供生成器标识
	Name() string
}

// BasePlanGenerator 规划生成器基类
// 【推荐】提供基础实现减少重复代码
type BasePlanGenerator struct {
	name   string
	logger *zap.Logger
	config *ReActConfig
}

// NewBasePlanGenerator 创建基础规划生成器
// 【必须】提供构造函数确保必要字段初始化
func NewBasePlanGenerator(name string, logger *zap.Logger, config *ReActConfig) *BasePlanGenerator {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = NewReActConfig()
	}

	return &BasePlanGenerator{
		name:   name,
		logger: logger.With(zap.String("generator", name)),
		config: config,
	}
}

// Name 返回生成器名称
// 【必须】实现 PlanGenerator 接口的 Name 方法
func (bpg *BasePlanGenerator) Name() string {
	return bpg.name
}

// Validate 验证生成器配置
// 【必须】实现 PlanGenerator 接口的 Validate 方法
func (bpg *BasePlanGenerator) Validate() error {
	if bpg.name == "" {
		return errors.NewValidationError("generator name cannot be empty", nil)
	}

	if bpg.config == nil {
		return errors.NewValidationError("config cannot be nil", nil)
	}

	return bpg.config.Validate()
}

// LLMPlanGenerator 基于LLM的规划生成器
// 【必须】使用大语言模型生成规划
type LLMPlanGenerator struct {
	BasePlanGenerator
	// ModelClient 模型客户端
	ModelClient agent.LLMClient
	// PromptTemplate 提示词模板
	PromptTemplate string
	// MaxRetries 最大重试次数
	MaxRetries int
	// EnableFallback 是否启用降级策略
	EnableFallback bool
}

// NewLLMPlanGenerator 创建基于LLM的规划生成器
// 【必须】提供构造函数确保必要配置
func NewLLMPlanGenerator(logger *zap.Logger, config *ReActConfig, modelClient agent.LLMClient, promptTemplate string) *LLMPlanGenerator {
	generator := &LLMPlanGenerator{
		BasePlanGenerator: *NewBasePlanGenerator("llm_plan_generator", logger, config),
		ModelClient:       modelClient,
		PromptTemplate:    promptTemplate,
		MaxRetries:        3,
		EnableFallback:    true,
	}

	// 【必须】设置合理的默认值
	if generator.PromptTemplate == "" {
		generator.PromptTemplate = generator.getDefaultPromptTemplate()
	}

	if generator.MaxRetries <= 0 {
		generator.MaxRetries = 3
	}

	return generator
}

// getDefaultPromptTemplate 获取默认提示词模板
// 【推荐】提供默认的规划生成提示词
func (lpg *LLMPlanGenerator) getDefaultPromptTemplate() string {
	return `你是一个智能规划助手，需要为用户查询生成详细的执行规划。

用户查询: {{.Query}}

历史上下文:
{{range .Thoughts}}- {{.Content}} (置信度: {{.Confidence}}){{end}}

可用工具:
{{range .AvailableTools}}- {{.Name}}: {{.Description}}{{end}}

请生成一个详细的执行规划，包括以下步骤：
1. 分析用户查询的关键要素
2. 确定需要的工具和动作
3. 安排执行顺序和依赖关系
4. 预估所需时间和资源

规划要求：
- 每个步骤必须有明确的目标和预期结果
- 考虑步骤间的依赖关系
- 提供适当的错误处理策略
- 确保规划的可执行性

请以JSON格式返回规划，包含steps数组，每个step包含name、description、dependencies等字段。`
}

// Generate 生成规划
// 【必须】实现 PlanGenerator 接口的 Generate 方法
func (lpg *LLMPlanGenerator) Generate(ctx context.Context, query string, state *ReActState) (*Plan, error) {
	if query == "" {
		return nil, errors.NewValidationError("query cannot be empty", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录生成开始
	lpg.logger.Debug("starting plan generation",
		zap.String("query", query),
		zap.String("session_id", state.SessionID),
		zap.Int("available_tools", len(lpg.getAvailableTools())),
	)

	var lastError error
	
	// 重试机制
	for retry := 0; retry <= lpg.MaxRetries; retry++ {
		select {
		case <-ctx.Done():
			return nil, errors.NewContextError("plan generation cancelled", ctx.Err(), nil)
		default:
			plan, err := lpg.generateWithRetry(ctx, query, state, retry)
			if err == nil {
				// 【必须】记录生成成功
				lpg.logger.Info("plan generation completed successfully",
					zap.String("plan_id", plan.ID),
					zap.Int("steps_count", len(plan.Steps)),
					zap.Int("retries", retry),
				)
				return plan, nil
			}

			lastError = err
			lpg.logger.Warn("plan generation attempt failed",
				zap.Error(err),
				zap.Int("retry_attempt", retry),
				zap.Int("max_retries", lpg.MaxRetries),
			)

			if retry < lpg.MaxRetries {
				// 指数退避
				backoffTime := time.Duration(1<<uint(retry)) * time.Second
				lpg.logger.Debug("waiting before retry", zap.Duration("backoff", backoffTime))
				time.Sleep(backoffTime)
			}
		}
	}

	// 如果启用了降级策略，使用备用方案
	if lpg.EnableFallback {
		lpg.logger.Warn("plan generation failed, using fallback strategy",
			zap.Error(lastError),
			zap.Int("attempts", lpg.MaxRetries+1),
		)
		return lpg.generateFallbackPlan(query, state)
	}

	return nil, errors.WrapError(lastError, "plan generation failed after all retries", map[string]interface{}{
		"query":       query,
		"session_id":  state.SessionID,
		"max_retries": lpg.MaxRetries,
	})
}

// generateWithRetry 带重试的生成逻辑
// 【推荐】封装单次生成逻辑便于重试
func (lpg *LLMPlanGenerator) generateWithRetry(ctx context.Context, query string, state *ReActState, retryCount int) (*Plan, error) {
	// 准备提示词
	prompt, err := lpg.preparePrompt(query, state)
	if err != nil {
		return nil, errors.WrapError(err, "failed to prepare prompt", nil)
	}

	// 调用LLM生成规划
	response, err := lpg.ModelClient.Generate(ctx, prompt, &agent.GenerateOptions{
		MaxTokens:   lpg.config.MaxTokens,
		Temperature: lpg.config.Temperature,
		StopSequences: []string{"\n\n", "###"},
	})
	if err != nil {
		return nil, errors.WrapError(err, "LLM generation failed", nil)
	}

	if response == "" {
		return nil, errors.NewValidationError("LLM returned empty response", nil)
	}

	// 解析生成的规划
	plan, err := lpg.parseGeneratedPlan(response, query)
	if err != nil {
		return nil, errors.WrapError(err, "failed to parse generated plan", map[string]interface{}{
			"response_length": len(response),
			"retry_count":     retryCount,
		})
	}

	// 验证生成的规划
	if err := plan.Validate(); err != nil {
		return nil, errors.WrapError(err, "generated plan validation failed", nil)
	}

	return plan, nil
}

// preparePrompt 准备提示词
// 【推荐】构建发送给LLM的提示词
func (lpg *LLMPlanGenerator) preparePrompt(query string, state *ReActState) (string, error) {
	// 简单的模板替换实现
	prompt := lpg.PromptTemplate

	// 替换查询
	prompt = replacePlaceholder(prompt, "{{.Query}}", query)

	// 构建思考历史
	thoughtsText := ""
	if state.Thoughts != nil {
		for _, thought := range state.Thoughts {
			thoughtsText += fmt.Sprintf("- %s (置信度: %.2f)\n", thought.Content, thought.Confidence)
		}
	}
	prompt = replacePlaceholder(prompt, "{{range .Thoughts}}- {{.Content}} (置信度: {{.Confidence}}){{end}}", thoughtsText)

	// 构建可用工具信息
	toolsText := ""
	tools := lpg.getAvailableTools()
	for _, tool := range tools {
		toolsText += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
	}
	prompt = replacePlaceholder(prompt, "{{range .AvailableTools}}- {{.Name}}: {{.Description}}{{end}}", toolsText)

	return prompt, nil
}

// getAvailableTools 获取可用工具列表
// 【推荐】获取当前可用的工具信息
func (lpg *LLMPlanGenerator) getAvailableTools() []agent.Tool {
	// 【必须】这里应该从工具注册表获取实际工具
	// 暂时返回空数组作为示例
	return []agent.Tool{}
}

// parseGeneratedPlan 解析生成的规划
// 【推荐】从LLM响应中解析出规划对象
func (lpg *LLMPlanGenerator) parseGeneratedPlan(response string, query string) (*Plan, error) {
	// 尝试提取JSON部分
	jsonStart := -1
	jsonEnd := -1

	// 查找JSON数组的开始和结束
	for i, char := range response {
		if char == '[' && jsonStart == -1 {
			jsonStart = i
		}
		if char == ']' && jsonStart != -1 {
			jsonEnd = i + 1
			break
		}
	}

	var stepsData string
	if jsonStart != -1 && jsonEnd != -1 {
		stepsData = response[jsonStart:jsonEnd]
	} else {
		// 如果没有找到JSON数组，尝试整个响应
		stepsData = response
	}

	// 解析步骤数据
	var stepInfos []map[string]interface{}
	if err := json.Unmarshal([]byte(stepsData), &stepInfos); err != nil {
		// 如果JSON解析失败，尝试手动解析
		stepInfos = lpg.manualParseSteps(response)
	}

	// 创建规划对象
	plan, err := NewPlan(query, planning.PlanPriorityMedium)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create plan", nil)
	}

	// 转换步骤信息为规划步骤
	for i, stepInfo := range stepInfos {
		step, err := lpg.convertToPlanStep(stepInfo, i+1)
		if err != nil {
			lpg.logger.Warn("failed to convert step info, skipping",
				zap.Error(err),
				zap.Any("step_info", stepInfo),
			)
			continue
		}
		plan.Steps = append(plan.Steps, step)
	}

	if len(plan.Steps) == 0 {
		return nil, errors.NewValidationError("no valid steps parsed from response", nil)
	}

	return plan, nil
}

// manualParseSteps 手动解析步骤（备用方案）
// 【推荐】当JSON解析失败时的备用解析方案
func (lpg *LLMPlanGenerator) manualParseSteps(response string) []map[string]interface{} {
	steps := make([]map[string]interface{}, 0)
	
	// 简单的文本解析逻辑
	lines := splitLines(response)
	currentStep := make(map[string]interface{})
		hasStepContent := false
	
	for _, line := range lines {
		line = trimWhitespace(line)
		
		// 检测步骤开始
		if startsWithNumber(line) && len(line) > 2 {
			// 保存前一个步骤
			if hasStepContent {
				steps = append(steps, currentStep)
			}
			
			// 开始新步骤
			currentStep = make(map[string]interface{})
			currentStep["name"] = extractStepName(line)
			currentStep["description"] = extractStepDescription(line)
			hasStepContent = true
		} else if hasStepContent && line != "" {
			// 累积步骤描述
			if desc, exists := currentStep["description"]; exists {
				currentStep["description"] = desc.(string) + " " + line
			} else {
				currentStep["description"] = line
			}
		}
	}
	
	// 添加最后一个步骤
	if hasStepContent {
		steps = append(steps, currentStep)
	}
	
	return steps
}

// convertToPlanStep 转换为规划步骤
// 【推荐】将步骤信息转换为标准的规划步骤
func (lpg *LLMPlanGenerator) convertToPlanStep(stepInfo map[string]interface{}, order int) (*planning.Step, error) {
	if stepInfo == nil {
		return nil, errors.NewValidationError("step info cannot be nil", nil)
	}

	step := &planning.Step{
		ID:          uuid.New().String(),
		Name:        getStringFromMap(stepInfo, "name", fmt.Sprintf("Step %d", order)),
		Description: getStringFromMap(stepInfo, "description", "No description provided"),
		Order:       order,
		Status:      planning.StepStatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	// 设置依赖关系
	if dependencies, exists := stepInfo["dependencies"]; exists {
		if depArray, ok := dependencies.([]interface{}); ok {
			depStrings := make([]string, 0, len(depArray))
			for _, dep := range depArray {
				if depStr, ok := dep.(string); ok {
					depStrings = append(depStrings, depStr)
				}
			}
			step.Dependencies = depStrings
		}
	}

	// 设置预估时长
	if durationStr, exists := stepInfo["duration"]; exists {
		if duration, err := time.ParseDuration(getStringFromMap(stepInfo, "duration", "0")); err == nil {
			step.EstimatedDuration = duration
		}
	}

	// 设置优先级
	if priorityStr, exists := stepInfo["priority"]; exists {
		if priority, err := planning.ParsePlanPriority(getStringFromMap(stepInfo, "priority", "medium")); err == nil {
			step.Priority = priority
		}
	}

	return step, nil
}

// generateFallbackPlan 生成降级规划
// 【必须】当主要生成策略失败时的备用方案
func (lpg *LLMPlanGenerator) generateFallbackPlan(query string, state *ReActState) (*Plan, error) {
	lpg.logger.Info("generating fallback plan",
		zap.String("query", query),
		zap.String("session_id", state.SessionID),
	)

	plan, err := NewPlan(query, planning.PlanPriorityLow)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create fallback plan", nil)
	}

	// 创建基础步骤
	fallbackSteps := []struct{
		name        string
		description string
		duration    time.Duration
	}{
		{"analyze_query", "分析用户查询的需求和目标", 30 * time.Second},
		{"search_information", "搜索相关信息和数据", 60 * time.Second},
		{"process_results", "处理和分析搜索结果", 45 * time.Second},
		{"generate_response", "生成最终回答和建议", 30 * time.Second},
	}

	for i, stepInfo := range fallbackSteps {
		step := &planning.Step{
			ID:               uuid.New().String(),
			Name:             stepInfo.name,
			Description:      stepInfo.description,
			Order:            i + 1,
			Status:           planning.StepStatusPending,
			EstimatedDuration: stepInfo.duration,
			Priority:         planning.PlanPriorityMedium,
			CreatedAt:        time.Now().UTC(),
		}
		plan.Steps = append(plan.Steps, step)
	}

	// 设置步骤依赖关系
	if len(plan.Steps) >= 2 {
		plan.Steps[1].Dependencies = []string{plan.Steps[0].ID} // search depends on analyze
	}
	if len(plan.Steps) >= 3 {
		plan.Steps[2].Dependencies = []string{plan.Steps[1].ID} // process depends on search
	}
	if len(plan.Steps) >= 4 {
		plan.Steps[3].Dependencies = []string{plan.Steps[2].ID} // generate depends on process
	}

	return plan, nil
}

// Refine 优化规划
// 【必须】实现 PlanGenerator 接口的 Refine 方法
func (lpg *LLMPlanGenerator) Refine(ctx context.Context, plan *Plan, observations []*Observation, state *ReActState) (*Plan, error) {
	if plan == nil {
		return nil, errors.NewValidationError("plan cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录优化开始
	lpg.logger.Debug("starting plan refinement",
		zap.String("plan_id", plan.ID),
		zap.Int("observations_count", len(observations)),
		zap.Int("current_steps", len(plan.Steps)),
	)

	// 分析观察结果
	refinementAnalysis, err := lpg.analyzeObservations(ctx, observations, plan)
	if err != nil {
		lpg.logger.Warn("failed to analyze observations for refinement", zap.Error(err))
		// 【必须】分析失败不应阻止优化，返回原规划
		return plan, nil
	}

	// 检查是否需要优化
	if !refinementAnalysis.NeedsRefinement {
		lpg.logger.Debug("plan does not need refinement", zap.String("plan_id", plan.ID))
		return plan, nil
	}

	// 创建优化后的规划副本
	refinedPlan := plan.Clone()
	refinedPlan.ID = uuid.New().String() // 新的规划ID
	refinedPlan.Timestamp = time.Now().UTC()

	// 应用优化建议
	err = lpg.applyRefinements(ctx, refinedPlan, refinementAnalysis, state)
	if err != nil {
		lpg.logger.Error("failed to apply refinements", zap.Error(err))
		// 【必须】优化失败返回原规划
		return plan, nil
	}

	// 验证优化后的规划
	if err := refinedPlan.Validate(); err != nil {
		lpg.logger.Error("refined plan validation failed", zap.Error(err))
		return plan, nil
	}

	// 【必须】记录优化完成
	lpg.logger.Info("plan refinement completed",
		zap.String("original_plan_id", plan.ID),
		zap.String("refined_plan_id", refinedPlan.ID),
		zap.Int("original_steps", len(plan.Steps)),
		zap.Int("refined_steps", len(refinedPlan.Steps)),
	)

	return refinedPlan, nil
}

// RefinementAnalysis 优化分析结果
// 【推荐】封装优化分析的结果
type RefinementAnalysis struct {
	NeedsRefinement bool                    `json:"needs_refinement"`
	Issues          []string                `json:"issues"`
	Suggestions     []string                `json:"suggestions"`
	StepModifications []StepModification   `json:"step_modifications"`
	NewSteps        []NewStepSuggestion     `json:"new_steps"`
}

// StepModification 步骤修改建议
// 【推荐】封装单个步骤的修改建议
type StepModification struct {
	StepID     string                 `json:"step_id"`
	Action     string                 `json:"action"` // modify, remove, reorder
	Reason     string                 `json:"reason"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// NewStepSuggestion 新步骤建议
// 【推荐】封装新增步骤的建议
type NewStepSuggestion struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Position    int                    `json:"position"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Reason      string                 `json:"reason"`
}

// analyzeObservations 分析观察结果
// 【推荐】基于观察结果分析规划优化需求
func (lpg *LLMPlanGenerator) analyzeObservations(ctx context.Context, observations []*Observation, plan *Plan) (*RefinementAnalysis, error) {
	analysis := &RefinementAnalysis{
		NeedsRefinement: false,
		Issues:          make([]string, 0),
		Suggestions:     make([]string, 0),
		StepModifications: make([]StepModification, 0),
		NewSteps:        make([]NewStepSuggestion, 0),
	}

	// 分析失败的操作
	failedObservations := lpg.filterFailedObservations(observations)
	if len(failedObservations) > 0 {
		analysis.NeedsRefinement = true
		analysis.Issues = append(analysis.Issues, fmt.Sprintf("%d operations failed", len(failedObservations)))
	
		// 为每个失败的观察生成修改建议
		for _, obs := range failedObservations {
			modification := lpg.suggestStepModification(obs, plan)
			analysis.StepModifications = append(analysis.StepModifications, modification)
		}
	}

	// 分析步骤执行时间
	slowSteps := lpg.identifySlowSteps(plan, observations)
	if len(slowSteps) > 0 {
		analysis.Suggestions = append(analysis.Suggestions, fmt.Sprintf("optimize %d slow steps", len(slowSteps)))
	}

	// 检查步骤依赖关系
	dependencyIssues := lpg.checkDependencyIssues(plan)
	if len(dependencyIssues) > 0 {
		analysis.NeedsRefinement = true
		analysis.Issues = append(analysis.Issues, "dependency issues detected")
		analysis.StepModifications = append(analysis.StepModifications, dependencyIssues...)
	}

	// 基于历史表现建议新步骤
	historicalSuggestions := lpg.suggestNewStepsBasedOnHistory(plan, observations)
	analysis.NewSteps = append(analysis.NewSteps, historicalSuggestions...)

	if len(analysis.Issues) > 0 || len(analysis.NewSteps) > 0 {
		analysis.NeedsRefinement = true
	}

	return analysis, nil
}

// filterFailedObservations 过滤失败的观察结果
// 【推荐】提取执行失败的操作观察
func (lpg *LLMPlanGenerator) filterFailedObservations(observations []*Observation) []*Observation {
	failed := make([]*Observation, 0)
	for _, obs := range observations {
		if obs != nil && !obs.Success {
			failed = append(failed, obs)
		}
	}
	return failed
}

// suggestStepModification 建议步骤修改
// 【推荐】基于失败的观察结果建议步骤修改
func (lpg *LLMPlanGenerator) suggestStepModification(obs *Observation, plan *Plan) StepModification {
	modification := StepModification{
		StepID: obs.ActionID,
		Reason: fmt.Sprintf("action failed: %s", obs.Error),
	}

	// 根据失败类型建议不同的修改
	if containsSubstring(obs.Error, "timeout") || containsSubstring(obs.Error, "deadline") {
		modification.Action = "modify"
		modification.Parameters = map[string]interface{}{
			"increase_timeout": true,
			"add_retry_logic":  true,
		}
	} else if containsSubstring(obs.Error, "not found") || containsSubstring(obs.Error, "missing") {
		modification.Action = "modify"
		modification.Parameters = map[string]interface{}{
			"add_validation": true,
			"provide_default": true,
		}
	} else {
		modification.Action = "modify"
		modification.Parameters = map[string]interface{}{
			"add_error_handling": true,
			"improve_robustness": true,
		}
	}

	return modification
}

// identifySlowSteps 识别慢速步骤
// 【推荐】基于观察结果识别执行缓慢的步骤
func (lpg *LLMPlanGenerator) identifySlowSteps(plan *Plan, observations []*Observation) []string {
	// 简单的实现：基于预估时间和实际表现的比较
	slowSteps := make([]string, 0)
	// 这里可以实现更复杂的性能分析逻辑
	return slowSteps
}

// checkDependencyIssues 检查依赖问题
// 【推荐】检查规划中的依赖关系问题
func (lpg *LLMPlanGenerator) checkDependencyIssues(plan *Plan) []StepModification {
	issues := make([]StepModification, 0)
	
	// 检查循环依赖
	visited := make(map[string]bool)
	path := make(map[string]bool)
	
	var checkCycle func(stepID string) bool
	checkCycle = func(stepID string) bool {
		if path[stepID] {
			// 发现循环依赖
			issues = append(issues, StepModification{
				StepID: stepID,
				Action: "modify",
				Reason: "circular dependency detected",
			})
			return true
		}
		
		if visited[stepID] {
			return false
		}
		
		visited[stepID] = true
		path[stepID] = true
		
		// 检查该步骤的依赖
		for _, step := range plan.Steps {
			if step.ID == stepID {
				for _, depID := range step.Dependencies {
					if checkCycle(depID) {
						return true
					}
				}
				break
			}
		}
		
		path[stepID] = false
		return false
	}
	
	// 检查所有步骤
	for _, step := range plan.Steps {
		if !visited[step.ID] {
			if checkCycle(step.ID) {
				break
			}
		}
	}
	
	return issues
}

// suggestNewStepsBasedOnHistory 基于历史建议新步骤
// 【推荐】基于历史执行经验建议新的规划步骤
func (lpg *LLMPlanGenerator) suggestNewStepsBasedOnHistory(plan *Plan, observations []*Observation) []NewStepSuggestion {
	suggestions := make([]NewStepSuggestion, 0)
	
	// 分析常见的失败模式并建议预防措施
	failurePatterns := lpg.analyzeFailurePatterns(observations)
	
	for pattern, frequency := range failurePatterns {
		if frequency >= 2 { // 如果某个模式失败了2次以上
			suggestion := NewStepSuggestion{
				Name:        fmt.Sprintf("prevent_%s", pattern),
				Description: fmt.Sprintf("Add preventive step for %s pattern", pattern),
				Position:    0, // 在开始前插入
				Reason:      fmt.Sprintf("%s pattern failed %d times", pattern, frequency),
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions
}

// analyzeFailurePatterns 分析失败模式
// 【推荐】分析观察结果中的失败模式
func (lpg *LLMPlanGenerator) analyzeFailurePatterns(observations []*Observation) map[string]int {
	patterns := make(map[string]int)
	
	for _, obs := range observations {
		if obs != nil && !obs.Success {
			// 简单的错误模式识别
			if containsSubstring(obs.Error, "timeout") {
				patterns["timeout"]++
			} else if containsSubstring(obs.Error, "not found") {
				patterns["not_found"]++
			} else if containsSubstring(obs.Error, "permission") {
				patterns["permission_denied"]++
			} else {
				patterns["general_failure"]++
			}
		}
	}
	
	return patterns
}

// applyRefinements 应用优化建议
// 【推荐】将优化分析的结果应用到规划上
func (lpg *LLMPlanGenerator) applyRefinements(ctx context.Context, plan *Plan, analysis *RefinementAnalysis, state *ReActState) error {
	// 应用步骤修改
	for _, modification := range analysis.StepModifications {
		switch modification.Action {
		case "modify":
			lpg.modifyStep(plan, modification)
		case "remove":
			lpg.removeStep(plan, modification.StepID)
		case "reorder":
			lpg.reorderStep(plan, modification)
		}
	}

	// 添加新步骤
	for _, newStep := range analysis.NewSteps {
		lpg.addNewStep(plan, newStep)
	}

	return nil
}

// modifyStep 修改步骤
// 【推荐】修改现有步骤的属性
func (lpg *LLMPlanGenerator) modifyStep(plan *Plan, modification StepModification) {
	for i, step := range plan.Steps {
		if step.ID == modification.StepID {
			// 应用修改参数
			if params := modification.Parameters; params != nil {
				if increaseTimeout, ok := params["increase_timeout"].(bool); ok && increaseTimeout {
					step.EstimatedDuration = step.EstimatedDuration * 2
				}
				if addValidation, ok := params["add_validation"].(bool); ok && addValidation {
					step.Description += " (with enhanced validation)"
				}
			}
			plan.Steps[i] = step
			break
		}
	}
}

// removeStep 移除步骤
// 【推荐】从规划中移除指定步骤
func (lpg *LLMPlanGenerator) removeStep(plan *Plan, stepID string) {
	filteredSteps := make([]*planning.Step, 0)
	for _, step := range plan.Steps {
		if step.ID != stepID {
			filteredSteps = append(filteredSteps, step)
		}
	}
	plan.Steps = filteredSteps
}

// reorderStep 重新排序步骤
// 【推荐】调整步骤的执行顺序
func (lpg *LLMPlanGenerator) reorderStep(plan *Plan, modification StepModification) {
	// 实现步骤重新排序逻辑
	// 这里可以根据modification.Parameters中的信息调整步骤顺序
}

// addNewStep 添加新步骤
// 【推荐】向规划中添加新的步骤
func (lpg *LLMPlanGenerator) addNewStep(plan *Plan, newStep NewStepSuggestion) {
	step := &planning.Step{
		ID:          uuid.New().String(),
		Name:        newStep.Name,
		Description: newStep.Description,
		Order:       newStep.Position,
		Status:      planning.StepStatusPending,
		CreatedAt:   time.Now().UTC(),
		Dependencies: newStep.Dependencies,
	}

	// 插入到指定位置
	if newStep.Position >= len(plan.Steps) {
		plan.Steps = append(plan.Steps, step)
	} else {
		// 在指定位置插入
		plan.Steps = append(plan.Steps, nil)
		copy(plan.Steps[newStep.Position+1:], plan.Steps[newStep.Position:])
		plan.Steps[newStep.Position] = step
		
		// 调整后续步骤的序号
		for i := newStep.Position + 1; i < len(plan.Steps); i++ {
			plan.Steps[i].Order = i + 1
		}
	}
}

// Validate 验证LLM规划生成器配置
// 【必须】实现 PlanGenerator 接口的 Validate 方法
func (lpg *LLMPlanGenerator) Validate() error {
	if err := lpg.BasePlanGenerator.Validate(); err != nil {
		return errors.WrapError(err, "base generator validation failed", nil)
	}

	if lpg.ModelClient == nil {
		return errors.NewValidationError("model client cannot be nil", nil)
	}

	if lpg.MaxRetries < 0 {
		return errors.NewValidationError("max retries cannot be negative", map[string]interface{}{
			"max_retries": lpg.MaxRetries,
		})
	}

	return nil
}

// PlanManager 规划管理器
// 【必须】管理规划的生命周期和执行
type PlanManager struct {
	// generators 规划生成器集合
	generators []PlanGenerator
	// activePlans 活跃的规划
	activePlans map[string]*Plan
	// logger 日志记录器
	logger *zap.Logger
}

// NewPlanManager 创建规划管理器
// 【必须】提供构造函数确保必要初始化
func NewPlanManager(logger *zap.Logger) *PlanManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PlanManager{
		generators:   make([]PlanGenerator, 0),
		activePlans:  make(map[string]*Plan),
		logger:       logger.Named("plan_manager"),
	}
}

// AddGenerator 添加规划生成器
// 【必须】提供生成器管理功能
func (pm *PlanManager) AddGenerator(generator PlanGenerator) error {
	if generator == nil {
		return errors.NewValidationError("generator cannot be nil", nil)
	}

	if err := generator.Validate(); err != nil {
		return errors.WrapError(err, "generator validation failed", nil)
	}

	pm.generators = append(pm.generators, generator)
	pm.logger.Debug("added plan generator",
		zap.String("generator_name", generator.Name()),
		zap.Int("generators_count", len(pm.generators)),
	)

	return nil
}

// GeneratePlan 生成规划
// 【必须】使用注册的生成器生成规划
func (pm *PlanManager) GeneratePlan(ctx context.Context, query string, state *ReActState) (*Plan, error) {
	if query == "" {
		return nil, errors.NewValidationError("query cannot be empty", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录生成开始
	pm.logger.Debug("starting plan generation via manager",
		zap.String("query", query),
		zap.String("session_id", state.SessionID),
		zap.Int("generators_count", len(pm.generators)),
	)

	if len(pm.generators) == 0 {
		return nil, errors.NewValidationError("no plan generators available", nil)
	}

	var lastError error
	var bestPlan *Plan
	
	// 尝试所有生成器，选择最好的结果
	for i, generator := range pm.generators {
		select {
		case <-ctx.Done():
			return nil, errors.NewContextError("plan generation cancelled", ctx.Err(), nil)
		default:
			plan, err := generator.Generate(ctx, query, state)
			if err == nil {
				// 如果是第一个成功的生成器，直接使用
				if bestPlan == nil {
					bestPlan = plan
				} else {
					// 比较规划质量，选择更好的
					if pm.comparePlanQuality(plan, bestPlan) > 0 {
						bestPlan = plan
					}
				}
			} else {
				lastError = err
				pm.logger.Warn("plan generator failed",
					zap.Error(err),
					zap.String("generator_name", generator.Name()),
					zap.Int("generator_index", i),
				)
			}
		}
	}

	if bestPlan != nil {
		// 保存活跃规划
		pm.activePlans[bestPlan.ID] = bestPlan
		
		// 【必须】记录生成成功
		pm.logger.Info("plan generation completed via manager",
			zap.String("plan_id", bestPlan.ID),
			zap.Int("steps_count", len(bestPlan.Steps)),
			zap.Int("generators_tried", len(pm.generators)),
		)
		return bestPlan, nil
	}

	return nil, errors.WrapError(lastError, "all plan generators failed", map[string]interface{}{
		"query":            query,
		"session_id":       state.SessionID,
		"generators_tried": len(pm.generators),
	})
}

// RefinePlan 优化规划
// 【必须】优化现有规划
func (pm *PlanManager) RefinePlan(ctx context.Context, plan *Plan, observations []*Observation, state *ReActState) (*Plan, error) {
	if plan == nil {
		return nil, errors.NewValidationError("plan cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录优化开始
	pm.logger.Debug("starting plan refinement via manager",
		zap.String("plan_id", plan.ID),
		zap.Int("observations_count", len(observations)),
	)

	// 使用第一个可用的生成器进行优化
	for _, generator := range pm.generators {
		select {
		case <-ctx.Done():
			return nil, errors.NewContextError("plan refinement cancelled", ctx.Err(), nil)
		default:
			refinedPlan, err := generator.Refine(ctx, plan, observations, state)
			if err == nil {
				// 更新活跃规划
				delete(pm.activePlans, plan.ID)
				pm.activePlans[refinedPlan.ID] = refinedPlan
				
				// 【必须】记录优化成功
				pm.logger.Info("plan refinement completed via manager",
					zap.String("original_plan_id", plan.ID),
					zap.String("refined_plan_id", refinedPlan.ID),
				)
				return refinedPlan, nil
			}
		}
	}

	// 【必须】所有生成器都失败时返回原规划
	pm.logger.Warn("plan refinement failed for all generators, returning original",
		zap.String("plan_id", plan.ID),
	)
	return plan, nil
}

// GetActivePlan 获取活跃规划
// 【推荐】根据ID获取活跃规划
func (pm *PlanManager) GetActivePlan(planID string) *Plan {
	return pm.activePlans[planID]
}

// RemoveActivePlan 移除活跃规划
// 【推荐】从活跃规划中移除指定规划
func (pm *PlanManager) RemoveActivePlan(planID string) {
	delete(pm.activePlans, planID)
	pm.logger.Debug("removed active plan", zap.String("plan_id", planID))
}

// ActivePlanCount 返回活跃规划数量
// 【推荐】提供管理器状态查询
func (pm *PlanManager) ActivePlanCount() int {
	return len(pm.activePlans)
}

// comparePlanQuality 比较规划质量
// 【推荐】比较不同规划的质量，选择更好的
func (pm *PlanManager) comparePlanQuality(plan1, plan2 *Plan) int {
	// 基于步骤数量、预估时间等因素比较
	// 这里可以实现更复杂的比较逻辑
	
	if len(plan1.Steps) != len(plan2.Steps) {
		if len(plan1.Steps) > len(plan2.Steps) {
			return 1
		}
		return -1
	}
	
	// 比较预估总时间
	duration1 := plan1.EstimatedDuration
	duration2 := plan2.EstimatedDuration
	
	if duration1 != duration2 {
		if duration1 < duration2 {
			return 1 // 时间越短越好
		}
		return -1
	}
	
	return 0 // 质量相等
}

// Validate 验证规划管理器配置
// 【必须】验证管理器的完整性
func (pm *PlanManager) Validate() error {
	if pm.generators == nil {
		return errors.NewValidationError("generators slice cannot be nil", nil)
	}

	if pm.activePlans == nil {
		return errors.NewValidationError("active plans map cannot be nil", nil)
	}

	for i, generator := range pm.generators {
		if generator == nil {
			return errors.NewValidationError(
				fmt.Sprintf("generator at index %d cannot be nil", i),
				map[string]interface{}{"index": i},
			)
		}

		if err := generator.Validate(); err != nil {
			return errors.WrapError(err, fmt.Sprintf("generator at index %d validation failed", i), nil)
		}
	}

	return nil
}

// Helper functions

// replacePlaceholder 替换占位符
func replacePlaceholder(text, placeholder, replacement string) string {
	// 简单的字符串替换实现
	result := text
	for {
		idx := indexOf(result, placeholder)
		if idx == -1 {
			break
		}
		result = result[:idx] + replacement + result[idx+len(placeholder):]
	}
	return result
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// splitLines 分割行为数组
func splitLines(s string) []string {
	lines := make([]string, 0)
	current := ""
	
	for _, r := range s {
		if r == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}

// trimWhitespace 去除首尾空白字符
func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	
	// 找到第一个非空白字符
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	// 找到最后一个非空白字符
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	if start >= end {
		return ""
	}
	
	return s[start:end]
}

// startsWithNumber 检查是否以数字开头
func startsWithNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	firstChar := s[0]
	return firstChar >= '0' && firstChar <= '9'
}

// extractStepName 提取步骤名称
func extractStepName(line string) string {
	// 简单的步骤名称提取
	for i, char := range line {
		if char == ' ' || char == '.' || char == '-' {
			return line[:i]
		}
	}
	return line
}

// extractStepDescription 提取步骤描述
func extractStepDescription(line string) string {
	// 简单的步骤描述提取
	for i, char := range line {
		if char == ' ' && i > 0 {
			return trimWhitespace(line[i:])
		}
	}
	return "No description"
}

// getStringFromMap 从map中获取字符串值
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// containsSubstring 检查字符串是否包含子串
func containsSubstring(s, substr string) bool {
	return indexOf(s, substr) != -1
}