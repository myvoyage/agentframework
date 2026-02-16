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
// of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// SkillPipeline 定义了技能执行管道接口，用于管理多个技能的组合执行
// 支持顺序执行、并行执行、条件执行等复杂执行流程
// 技能管道可以看作是一个有向无环图(DAG)，其中节点是技能，边是技能间的数据依赖
// 技能管道执行时，会按照依赖关系执行各个技能，并处理技能间的数据传递
// 技能管道支持执行监控、错误处理、重试机制等

type SkillPipeline interface {
	// AddStep 添加一个技能执行步骤到管道中
	AddStep(step *SkillStep) SkillPipeline

	// AddSteps 批量添加技能执行步骤到管道中
	AddSteps(steps []*SkillStep) SkillPipeline

	// AddDependency 添加技能步骤间的依赖关系
	AddDependency(fromStepID, toStepID string) SkillPipeline

	// Execute 执行技能管道，返回执行结果（同步）
	Execute(ctx context.Context, input map[string]interface{}) (*PipelineResult, error)

	// ExecuteAsync 异步执行技能管道，返回结果通道
	ExecuteAsync(ctx context.Context, input map[string]interface{}) (<-chan *PipelineResult, error)

	// GetSteps 获取所有技能执行步骤
	GetSteps() []*SkillStep

	// GetDependencies 获取所有技能步骤间的依赖关系
	GetDependencies() map[string][]string

	// Validate 验证技能管道的合法性，包括依赖关系、步骤配置等
	Validate() error
}

// SkillStep 定义了技能执行步骤，包含技能名称、输入映射、输出映射、错误处理策略等

type SkillStep struct {
	ID            string                 `json:"id"`             // 步骤唯一标识
	SkillName     string                 `json:"skill_name"`     // 技能名称
	Description   string                 `json:"description"`    // 步骤描述
	InputMapping  map[string]string      `json:"input_mapping"`  // 输入映射，键为技能参数名，值为输入来源表达式
	OutputMapping map[string]string      `json:"output_mapping"` // 输出映射，键为输出变量名，值为技能输出表达式
	ErrorPolicy   ErrorPolicy            `json:"error_policy"`   // 错误处理策略
	RetryPolicy   *RetryPolicy           `json:"retry_policy"`   // 重试策略
	Timeout       int                    `json:"timeout"`        // 步骤执行超时时间(秒)
	Parallel      bool                   `json:"parallel"`       // 是否并行执行
	Condition     string                 `json:"condition"`      // 条件表达式，决定是否执行该步骤
	Dependencies  []string               `json:"dependencies"`   // 依赖的步骤ID列表
	Config        map[string]interface{} `json:"config"`         // 步骤配置
}

// ErrorPolicy 定义了技能执行错误处理策略
type ErrorPolicy string

// 错误处理策略常量
const (
	ErrorPolicyContinue ErrorPolicy = "continue"  // 忽略错误，继续执行后续步骤
	ErrorPolicyFailFast ErrorPolicy = "fail_fast" // 立即失败，终止整个管道执行
	ErrorPolicyRetry    ErrorPolicy = "retry"     // 重试失败的步骤
	ErrorPolicySkip     ErrorPolicy = "skip"      // 跳过失败的步骤，继续执行后续步骤
	ErrorPolicyFallback ErrorPolicy = "fallback"  // 使用备用步骤替代失败的步骤
)

// RetryPolicy 定义了技能执行重试策略
type RetryPolicy struct {
	MaxRetries   int  `json:"max_retries"`   // 最大重试次数
	DelaySeconds int  `json:"delay_seconds"` // 重试间隔(秒)
	Exponential  bool `json:"exponential"`   // 是否使用指数退避策略
	MaxDelay     int  `json:"max_delay"`     // 最大重试间隔(秒)
}

// PipelineResult 定义了技能管道执行结果
type PipelineResult struct {
	Success    bool                   `json:"success"`     // 管道执行是否成功
	Steps      map[string]*StepResult `json:"steps"`       // 各个步骤的执行结果
	Outputs    map[string]interface{} `json:"outputs"`     // 管道输出
	StartTime  int64                  `json:"start_time"`  // 开始时间戳
	EndTime    int64                  `json:"end_time"`    // 结束时间戳
	DurationMs int64                  `json:"duration_ms"` // 执行时长(毫秒)
}

// StepResult 定义了单个技能步骤的执行结果
type StepResult struct {
	StepID     string                 `json:"step_id"`         // 步骤ID
	SkillName  string                 `json:"skill_name"`      // 技能名称
	Success    bool                   `json:"success"`         // 执行是否成功
	Input      map[string]interface{} `json:"input"`           // 技能输入
	Output     map[string]interface{} `json:"output"`          // 技能输出
	Error      string                 `json:"error,omitempty"` // 错误信息
	StartTime  int64                  `json:"start_time"`      // 开始时间戳
	EndTime    int64                  `json:"end_time"`        // 结束时间戳
	DurationMs int64                  `json:"duration_ms"`     // 执行时长(毫秒)
	RetryCount int                    `json:"retry_count"`     // 重试次数
}

// DefaultSkillPipeline 技能管道的默认实现
type DefaultSkillPipeline struct {
	skills  SkillLibrary          // 技能库，用于获取技能实例
	steps   map[string]*SkillStep // 技能执行步骤映射，键为步骤ID
	deps    map[string][]string   // 步骤依赖关系映射，键为步骤ID，值为依赖的步骤ID列表
	revDeps map[string][]string   // 反向依赖关系映射，键为步骤ID，值为依赖该步骤的步骤ID列表
}

// NewSkillPipeline 创建一个新的技能管道实例
func NewSkillPipeline(skillLibrary SkillLibrary) SkillPipeline {
	return &DefaultSkillPipeline{
		skills:  skillLibrary,
		steps:   make(map[string]*SkillStep),
		deps:    make(map[string][]string),
		revDeps: make(map[string][]string),
	}
}

// AddStep 添加一个技能执行步骤到管道中
func (p *DefaultSkillPipeline) AddStep(step *SkillStep) SkillPipeline {
	if step == nil {
		return p
	}

	if step.ID == "" {
		step.ID = fmt.Sprintf("step_%d", len(p.steps)+1)
	}

	p.steps[step.ID] = step

	// 添加步骤的依赖关系
	for _, depID := range step.Dependencies {
		p.AddDependency(depID, step.ID)
	}

	return p
}

// AddSteps 批量添加技能执行步骤到管道中
func (p *DefaultSkillPipeline) AddSteps(steps []*SkillStep) SkillPipeline {
	for _, step := range steps {
		p.AddStep(step)
	}
	return p
}

// AddDependency 添加技能步骤间的依赖关系
func (p *DefaultSkillPipeline) AddDependency(fromStepID, toStepID string) SkillPipeline {
	// 检查步骤是否存在
	if _, exists := p.steps[fromStepID]; !exists {
		return p
	}

	if _, exists := p.steps[toStepID]; !exists {
		return p
	}

	// 添加正向依赖
	if _, exists := p.deps[fromStepID]; !exists {
		p.deps[fromStepID] = []string{}
	}
	p.deps[fromStepID] = append(p.deps[fromStepID], toStepID)

	// 添加反向依赖
	if _, exists := p.revDeps[toStepID]; !exists {
		p.revDeps[toStepID] = []string{}
	}
	p.revDeps[toStepID] = append(p.revDeps[toStepID], fromStepID)

	return p
}

// Execute 执行技能管道，返回执行结果
func (p *DefaultSkillPipeline) Execute(ctx context.Context, input map[string]interface{}) (*PipelineResult, error) {
	// 验证管道合法性
	if err := p.Validate(); err != nil {
		return nil, err
	}

	// 初始化执行上下文
	execCtx := &executionContext{
		pipeline: p,
		input:    input,
		results:  make(map[string]*StepResult),
		state:    make(map[string]interface{}),
	}

	// 执行管道
	result, err := execCtx.execute(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteAsync 异步执行技能管道，返回结果通道
func (p *DefaultSkillPipeline) ExecuteAsync(ctx context.Context, input map[string]interface{}) (<-chan *PipelineResult, error) {
	// 验证管道合法性
	if err := p.Validate(); err != nil {
		return nil, err
	}

	// 创建结果通道
	resultCh := make(chan *PipelineResult, 1)

	// 在 goroutine 中执行管道
	go func() {
		// 初始化执行上下文
		execCtx := &executionContext{
			pipeline: p,
			input:    input,
			results:  make(map[string]*StepResult),
			state:    make(map[string]interface{}),
		}

		// 执行管道
		result, err := execCtx.execute(ctx)
		if err != nil {
			// 返回错误结果
			resultCh <- &PipelineResult{
				Success:    false,
				Steps:      nil,
				Outputs:    nil,
				StartTime:  0,
				EndTime:    0,
				DurationMs: 0,
			}
		} else {
			resultCh <- result
		}
		close(resultCh)
	}()

	return resultCh, nil
}

// GetSteps 获取所有技能执行步骤
func (p *DefaultSkillPipeline) GetSteps() []*SkillStep {
	steps := make([]*SkillStep, 0, len(p.steps))
	for _, step := range p.steps {
		steps = append(steps, step)
	}
	return steps
}

// GetDependencies 获取所有技能步骤间的依赖关系
func (p *DefaultSkillPipeline) GetDependencies() map[string][]string {
	return p.deps
}

// Validate 验证技能管道的合法性，包括依赖关系、步骤配置等
func (p *DefaultSkillPipeline) Validate() error {
	// 检查是否有循环依赖
	if hasCycle, cycle := p.hasCycle(); hasCycle {
		return errors.New(fmt.Sprintf("skill pipeline has cycle dependency: %v", cycle))
	}

	// 检查每个步骤的配置
	for _, step := range p.steps {
		if step.SkillName == "" {
			return errors.New(fmt.Sprintf("step %s has no skill name", step.ID))
		}

		// 检查技能是否存在于技能库中
		if _, exists := p.skills.GetSkill(context.Background(), step.SkillName); !exists {
			return errors.New(fmt.Sprintf("skill %s not found in skill library", step.SkillName))
		}

		// 检查依赖的步骤是否存在
		for _, depID := range step.Dependencies {
			if _, exists := p.steps[depID]; !exists {
				return errors.New(fmt.Sprintf("step %s depends on non-existent step %s", step.ID, depID))
			}
		}
	}

	return nil
}

// hasCycle 检查技能管道是否存在循环依赖
func (p *DefaultSkillPipeline) hasCycle() (bool, []string) {
	// 使用深度优先搜索检测循环依赖
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cycle := []string{}

	var dfs func(stepID string) bool
	dfs = func(stepID string) bool {
		visited[stepID] = true
		recStack[stepID] = true

		// 检查当前步骤的所有依赖
		for _, depID := range p.revDeps[stepID] {
			if !visited[depID] {
				if dfs(depID) {
					cycle = append(cycle, depID)
					return true
				}
			} else if recStack[depID] {
				cycle = append(cycle, depID, stepID)
				return true
			}
		}

		recStack[stepID] = false
		return false
	}

	// 对所有未访问的步骤执行DFS
	for stepID := range p.steps {
		if !visited[stepID] {
			if dfs(stepID) {
				return true, cycle
			}
		}
	}

	return false, nil
}

// executionContext 定义了技能管道执行上下文，用于管理执行过程中的状态、结果等
type executionContext struct {
	pipeline SkillPipeline          // 技能管道
	input    map[string]interface{} // 管道输入
	results  map[string]*StepResult // 步骤执行结果映射
	state    map[string]interface{} // 执行状态，用于存储步骤间传递的数据
}

// execute 执行技能管道
func (ctx *executionContext) execute(pipelineCtx context.Context) (*PipelineResult, error) {
	// 记录开始时间
	startTime := time.Now()

	// 初始化执行状态
	for k, v := range ctx.input {
		ctx.state[k] = v
	}

	// 构建拓扑排序
	topOrder, err := ctx.buildTopologicalOrder()
	if err != nil {
		return nil, err
	}

	// 分组并行步骤
	parallelGroups := ctx.groupParallelSteps(topOrder)

	// 执行各个并行分组
	for _, group := range parallelGroups {
		if err := ctx.executeParallelGroup(pipelineCtx, group); err != nil {
			return nil, err
		}
	}

	// 记录结束时间
	endTime := time.Now()

	// 构建执行结果
	result := &PipelineResult{
		Success:    true,
		Steps:      ctx.results,
		Outputs:    ctx.state,
		StartTime:  startTime.Unix(),
		EndTime:    endTime.Unix(),
		DurationMs: endTime.Sub(startTime).Milliseconds(),
	}

	// 检查是否有执行失败的步骤
	for _, stepResult := range ctx.results {
		if !stepResult.Success {
			result.Success = false
			break
		}
	}

	return result, nil
}

// buildTopologicalOrder 构建技能步骤的拓扑排序
func (ctx *executionContext) buildTopologicalOrder() ([]string, error) {
	pipeline := ctx.pipeline.(*DefaultSkillPipeline)

	// 计算每个步骤的入度
	inDegree := make(map[string]int)
	for stepID := range pipeline.steps {
		inDegree[stepID] = len(pipeline.revDeps[stepID])
	}

	// 将入度为0的步骤加入队列
	queue := []string{}
	for stepID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, stepID)
		}
	}

	// 执行拓扑排序
	topOrder := []string{}
	for len(queue) > 0 {
		// 取出队列中的第一个步骤
		current := queue[0]
		queue = queue[1:]

		// 将当前步骤加入拓扑排序结果
		topOrder = append(topOrder, current)

		// 遍历当前步骤的所有后继步骤
		for _, nextStepID := range pipeline.deps[current] {
			// 减少后继步骤的入度
			inDegree[nextStepID]--

			// 如果后继步骤的入度为0，加入队列
			if inDegree[nextStepID] == 0 {
				queue = append(queue, nextStepID)
			}
		}
	}

	// 检查是否所有步骤都被访问到（即是否存在循环依赖）
	if len(topOrder) != len(pipeline.steps) {
		return nil, errors.New("failed to build topological order, cycle dependency may exist")
	}

	return topOrder, nil
}

// groupParallelSteps 将拓扑排序后的步骤分组，同一组的步骤可以并行执行
func (ctx *executionContext) groupParallelSteps(topOrder []string) [][]string {
	pipeline := ctx.pipeline.(*DefaultSkillPipeline)

	groups := [][]string{}
	currentGroup := []string{}
	visited := make(map[string]bool)

	for _, stepID := range topOrder {
		step := pipeline.steps[stepID]

		// 如果步骤已经被访问过，跳过
		if visited[stepID] {
			continue
		}

		// 如果步骤不是并行执行，单独作为一组
		if !step.Parallel {
			groups = append(groups, []string{stepID})
			visited[stepID] = true
			continue
		}

		// 查找所有可以并行执行的步骤
		currentGroup = append(currentGroup, stepID)
		visited[stepID] = true

		// 查找所有依赖关系为并行的步骤
		for _, nextStepID := range pipeline.deps[stepID] {
			nextStep := pipeline.steps[nextStepID]
			if nextStep.Parallel && !visited[nextStepID] {
				currentGroup = append(currentGroup, nextStepID)
				visited[nextStepID] = true
			}
		}

		groups = append(groups, currentGroup)
		currentGroup = []string{}
	}

	return groups
}

// executeParallelGroup 并行执行一组技能步骤
func (ctx *executionContext) executeParallelGroup(pipelineCtx context.Context, group []string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// 遍历组中的每个步骤
	for _, stepID := range group {
		wg.Add(1)

		go func(stepID string) {
			defer wg.Done()

			// 执行步骤
			result, err := ctx.executeStep(pipelineCtx, stepID)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			// 处理执行结果
			mu.Lock()
			ctx.results[stepID] = result

			// 更新执行状态
			if result.Output != nil {
				for k, v := range result.Output {
					ctx.state[k] = v
				}
			}
			mu.Unlock()
		}(stepID)
	}

	// 等待所有步骤执行完成
	wg.Wait()

	return firstErr
}

// executeStep 执行单个技能步骤
func (ctx *executionContext) executeStep(pipelineCtx context.Context, stepID string) (*StepResult, error) {
	pipeline := ctx.pipeline.(*DefaultSkillPipeline)
	step := pipeline.steps[stepID]
	startTime := time.Now()

	// 检查条件表达式
	if step.Condition != "" {
		// 简单的条件表达式求值：支持变量引用和基本比较操作
		shouldExecute, err := ctx.evaluateCondition(pipelineCtx, step.Condition)
		if err != nil {
			return &StepResult{
				StepID:     stepID,
				SkillName:  step.SkillName,
				Success:    false,
				Input:      nil,
				Output:     nil,
				Error:      err.Error(),
				StartTime:  startTime.Unix(),
				EndTime:    time.Now().Unix(),
				DurationMs: time.Since(startTime).Milliseconds(),
				RetryCount: 0,
			}, nil
		}

		if !shouldExecute {
			return &StepResult{
				StepID:     stepID,
				SkillName:  step.SkillName,
				Success:    true,
				Input:      nil,
				Output:     nil,
				Error:      "",
				StartTime:  startTime.Unix(),
				EndTime:    time.Now().Unix(),
				DurationMs: time.Since(startTime).Milliseconds(),
				RetryCount: 0,
			}, nil
		}
	}

	// 解析输入参数
	inputParams, err := ctx.parseInputMapping(step)
	if err != nil {
		return nil, err
	}

	// 转换输入参数为JSON字符串
	inputJSON, err := json.Marshal(inputParams)
	if err != nil {
		return nil, err
	}

	// 获取技能实例
	skill, exists := pipeline.skills.GetSkill(pipelineCtx, step.SkillName)
	if !exists {
		return nil, errors.New(fmt.Sprintf("skill %s not found", step.SkillName))
	}

	// 执行技能
	outputJSON, err := skill.Invoke(pipelineCtx, string(inputJSON))
	endTime := time.Now()
	if err != nil {
		return &StepResult{
			StepID:     stepID,
			SkillName:  step.SkillName,
			Success:    false,
			Input:      inputParams,
			Output:     nil,
			Error:      err.Error(),
			StartTime:  startTime.Unix(),
			EndTime:    endTime.Unix(),
			DurationMs: endTime.Sub(startTime).Milliseconds(),
			RetryCount: 0,
		}, nil
	}

	// 解析技能输出
	var outputData map[string]interface{}
	if err := json.Unmarshal([]byte(outputJSON), &outputData); err != nil {
		return &StepResult{
			StepID:     stepID,
			SkillName:  step.SkillName,
			Success:    false,
			Input:      inputParams,
			Output:     nil,
			Error:      fmt.Sprintf("failed to parse skill output: %v", err),
			StartTime:  startTime.Unix(),
			EndTime:    endTime.Unix(),
			DurationMs: endTime.Sub(startTime).Milliseconds(),
			RetryCount: 0,
		}, nil
	}

	// 应用输出映射
	outputVars, err := ctx.applyOutputMapping(step, outputData)
	if err != nil {
		return &StepResult{
			StepID:     stepID,
			SkillName:  step.SkillName,
			Success:    false,
			Input:      inputParams,
			Output:     nil,
			Error:      fmt.Sprintf("failed to apply output mapping: %v", err),
			StartTime:  startTime.Unix(),
			EndTime:    endTime.Unix(),
			DurationMs: endTime.Sub(startTime).Milliseconds(),
			RetryCount: 0,
		}, nil
	}

	// 构建执行结果
	result := &StepResult{
		StepID:     stepID,
		SkillName:  step.SkillName,
		Success:    true,
		Input:      inputParams,
		Output:     outputVars,
		Error:      "",
		StartTime:  startTime.Unix(),
		EndTime:    endTime.Unix(),
		DurationMs: endTime.Sub(startTime).Milliseconds(),
		RetryCount: 0,
	}

	return result, nil
}

// parseInputMapping 解析技能步骤的输入映射，生成技能输入参数
func (ctx *executionContext) parseInputMapping(step *SkillStep) (map[string]interface{}, error) {
	input := make(map[string]interface{})

	// 如果没有输入映射，直接返回空映射
	if step.InputMapping == nil {
		return input, nil
	}

	// 遍历输入映射
	for paramName, expr := range step.InputMapping {
		// 使用 JavaScript 表达式求值引擎
		val, err := ctx.evaluateExpression(context.Background(), expr)
		if err != nil {
			// 如果表达式求值失败，使用空值
			input[paramName] = nil
		} else {
			input[paramName] = val
		}
	}

	return input, nil
}

// applyOutputMapping 应用技能步骤的输出映射，生成输出变量
func (ctx *executionContext) applyOutputMapping(step *SkillStep, outputData map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	// 如果没有输出映射，直接返回技能输出
	if step.OutputMapping == nil {
		return outputData, nil
	}

	// 遍历输出映射
	for varName, expr := range step.OutputMapping {
		// 使用 JavaScript 表达式求值引擎
		val, err := ctx.evaluateExpressionWithData(context.Background(), expr, outputData)
		if err != nil {
			// 如果表达式求值失败，使用空值
			output[varName] = nil
		} else {
			output[varName] = val
		}
	}

	return output, nil
}

// evaluateExpression 求值表达式，支持 JavaScript 表达式
func (ctx *executionContext) evaluateExpression(pipelineCtx context.Context, expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}

	// 创建 JavaScript 虚拟机
	vm := goja.New()

	// 注入执行状态变量
	for k, v := range ctx.state {
		if err := vm.Set(k, v); err != nil {
			return nil, fmt.Errorf("failed to set variable %s: %w", k, err)
		}
	}

	// 求值表达式
	value, err := vm.RunString(expr)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w", err)
	}

	return value.Export(), nil
}

// evaluateExpressionWithData 求值表达式，支持 JavaScript 表达式，并注入额外数据
func (ctx *executionContext) evaluateExpressionWithData(pipelineCtx context.Context, expr string, data map[string]interface{}) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}

	// 创建 JavaScript 虚拟机
	vm := goja.New()

	// 注入执行状态变量
	for k, v := range ctx.state {
		if err := vm.Set(k, v); err != nil {
			return nil, fmt.Errorf("failed to set variable %s: %w", k, err)
		}
	}

	// 注入额外数据
	for k, v := range data {
		if err := vm.Set(k, v); err != nil {
			return nil, fmt.Errorf("failed to set variable %s: %w", k, err)
		}
	}

	// 求值表达式
	value, err := vm.RunString(expr)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w", err)
	}

	return value.Export(), nil
}

// evaluateCondition 求值条件表达式
// 支持简单的变量引用和比较操作，格式：var1 operator var2
// 其中var1和var2可以是状态变量或常量，operator可以是 ==, !=, <, <=, >, >=
func (ctx *executionContext) evaluateCondition(pipelineCtx context.Context, condition string) (bool, error) {
	// 简单的条件表达式解析
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return true, nil
	}

	// 支持的比较操作符
	operators := []string{"!=", "==", "<=", ">=", "<", ">"}
	var op string
	var left, right string

	// 查找操作符
	for _, opcandidate := range operators {
		if idx := strings.Index(condition, opcandidate); idx != -1 {
			op = opcandidate
			left = strings.TrimSpace(condition[:idx])
			right = strings.TrimSpace(condition[idx+len(opcandidate):])
			break
		}
	}

	// 如果没有找到操作符，直接检查变量是否存在且为真
	if op == "" {
		val, exists := ctx.state[condition]
		if !exists {
			return false, nil
		}

		// 简单的真值判断
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			return v != "", nil
		case int, int64, float64:
			return v != 0, nil
		case nil:
			return false, nil
		default:
			return true, nil
		}
	}

	// 求值左侧和右侧表达式
	leftVal, err := ctx.evaluateExpression(pipelineCtx, left)
	if err != nil {
		return false, err
	}

	rightVal, err := ctx.evaluateExpression(pipelineCtx, right)
	if err != nil {
		return false, err
	}

	// 执行比较操作
	switch op {
	case "==":
		return leftVal == rightVal, nil
	case "!=":
		return leftVal != rightVal, nil
	case "<":
		// 简单的数值比较
		leftNum, leftOk := toNumber(leftVal)
		rightNum, rightOk := toNumber(rightVal)
		if leftOk && rightOk {
			return leftNum < rightNum, nil
		}
	case "<=":
		leftNum, leftOk := toNumber(leftVal)
		rightNum, rightOk := toNumber(rightVal)
		if leftOk && rightOk {
			return leftNum <= rightNum, nil
		}
	case ">":
		leftNum, leftOk := toNumber(leftVal)
		rightNum, rightOk := toNumber(rightVal)
		if leftOk && rightOk {
			return leftNum > rightNum, nil
		}
	case ">=":
		leftNum, leftOk := toNumber(leftVal)
		rightNum, rightOk := toNumber(rightVal)
		if leftOk && rightOk {
			return leftNum >= rightNum, nil
		}
	}

	return false, fmt.Errorf("unsupported comparison: %v %s %v", leftVal, op, rightVal)
}


// toNumber 将值转换为数值，如果无法转换则返回false
func toNumber(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case string:
		num, err := strconv.ParseFloat(v, 64)
		return num, err == nil
	default:
		return 0, false
	}
}
