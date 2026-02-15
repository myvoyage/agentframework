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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// TaskDecomposer 任务分解器，将复杂任务分解为子任务
type TaskDecomposer struct {
	model        ChatModel
	maxDepth     int
	maxTasks     int
	taskCounter  int
	availableAgents []string
}

// TaskDecomposerConfig 任务分解器配置
type TaskDecomposerConfig struct {
	// Model 用于分解任务的模型
	Model ChatModel
	// MaxDepth 最大分解深度
	MaxDepth int
	// MaxTasks 单次分解最大子任务数
	MaxTasks int
	// AvailableAgents 可用的Agent列表
	AvailableAgents []string
}

// NewTaskDecomposer 创建任务分解器
func NewTaskDecomposer(config TaskDecomposerConfig) *TaskDecomposer {
	if config.MaxDepth == 0 {
		config.MaxDepth = 3
	}
	if config.MaxTasks == 0 {
		config.MaxTasks = 10
	}

	return &TaskDecomposer{
		model:          config.Model,
		maxDepth:       config.MaxDepth,
		maxTasks:       config.MaxTasks,
		availableAgents: config.AvailableAgents,
	}
}

// DecomposeRequest 任务分解请求
type DecomposeRequest struct {
	// Task 要分解的任务描述
	Task string
	// Context 任务的上下文信息
	Context string
	// CurrentDepth 当前分解深度
	CurrentDepth int
	// Dependencies 依赖的任务ID
	Dependencies []string
	// ParentID 父任务ID
	ParentID string
}

// DecomposeResult 任务分解结果
type DecomposeResult struct {
	// SubTasks 分解后的子任务列表
	SubTasks []*SubTask
	// Reasoning 分解的推理过程
	Reasoning string
	// NeedsFurtherDecomposition 是否需要进一步分解
	NeedsFurtherDecomposition bool
	// EstimatedTotalTime 预估总时间（秒）
	EstimatedTotalTime int
}

// Decompose 分解任务为子任务
func (td *TaskDecomposer) Decompose(ctx context.Context, req DecomposeRequest) (*DecomposeResult, error) {
	// 检查深度限制
	if req.CurrentDepth >= td.maxDepth {
		return &DecomposeResult{
			SubTasks:                   []*SubTask{},
			Reasoning:                  "达到最大分解深度，停止分解",
			NeedsFurtherDecomposition:  false,
		}, nil
	}

	// 构建分解提示
	prompt := td.buildDecompositionPrompt(req)

	// 调用模型进行分解
	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个任务规划专家。你的职责是将复杂的任务分解为可执行的子任务，并分析任务之间的依赖关系。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := td.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, fmt.Errorf("任务分解失败: %w", err)
	}

	// 解析分解结果
	result, err := td.parseDecompositionResponse(resp.Content, req)
	if err != nil {
		return nil, fmt.Errorf("解析分解结果失败: %w", err)
	}

	return result, nil
}

// buildDecompositionPrompt 构建任务分解提示
func (td *TaskDecomposer) buildDecompositionPrompt(req DecomposeRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 任务分解请求\n\n"))
	sb.WriteString(fmt.Sprintf("## 任务描述\n%s\n\n", req.Task))

	if req.Context != "" {
		sb.WriteString(fmt.Sprintf("## 上下文信息\n%s\n\n", req.Context))
	}

	if req.CurrentDepth > 0 {
		sb.WriteString(fmt.Sprintf("## 当前分解深度\n%d\n\n", req.CurrentDepth))
	}

	if len(req.Dependencies) > 0 {
		sb.WriteString(fmt.Sprintf("## 依赖任务\n%s\n\n", strings.Join(req.Dependencies, ", ")))
	}

	if len(td.availableAgents) > 0 {
		sb.WriteString(fmt.Sprintf("## 可用的Agent\n%s\n\n", strings.Join(td.availableAgents, ", ")))
	}

	sb.WriteString(fmt.Sprintf("## 要求\n\n"))
	sb.WriteString(fmt.Sprintf("1. 将任务分解为 %d 个以内的子任务\n", td.maxTasks))
	sb.WriteString("2. 每个子任务应该是可独立执行的\n")
	sb.WriteString("3. 明确标识任务之间的依赖关系\n")
	sb.WriteString("4. 为每个子任务预估执行时间\n")
	sb.WriteString("5. 指定每个子任务的优先级（LOW, NORMAL, HIGH, CRITICAL）\n")
	sb.WriteString("6. 如果任务仍然复杂，标记需要进一步分解\n\n")

	sb.WriteString("## 输出格式\n\n")
	sb.WriteString("请以JSON格式返回分解结果，格式如下：\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"reasoning\": \"分解推理过程\",\n")
	sb.WriteString("  \"subtasks\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"id\": \"task-1\",\n")
	sb.WriteString("      \"description\": \"子任务描述\",\n")
	sb.WriteString("      \"details\": \"详细说明\",\n")
	sb.WriteString("      \"dependencies\": [],\n")
	sb.WriteString("      \"priority\": \"NORMAL\",\n")
	sb.WriteString("      \"assignee\": \"推荐的agent名称\",\n")
	sb.WriteString("      \"estimated_duration\": 300\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"needs_further_decomposition\": false,\n")
	sb.WriteString("  \"estimated_total_time\": 600\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	return sb.String()
}

// decompositionResponse 分解响应结构
type decompositionResponse struct {
	Reasoning                  string `json:"reasoning"`
	SubTasks                   []jsonSubTask `json:"subtasks"`
	NeedsFurtherDecomposition  bool   `json:"needs_further_decomposition"`
	EstimatedTotalTime         int    `json:"estimated_total_time"`
}

// jsonSubTask JSON格式的子任务
type jsonSubTask struct {
	ID                string `json:"id"`
	Description       string `json:"description"`
	Details           string `json:"details,omitempty"`
	Dependencies      []string `json:"dependencies,omitempty"`
	Priority          string `json:"priority,omitempty"`
	Assignee          string `json:"assignee,omitempty"`
	EstimatedDuration int    `json:"estimated_duration,omitempty"`
}

// parseDecompositionResponse 解析分解响应
func (td *TaskDecomposer) parseDecompositionResponse(content string, req DecomposeRequest) (*DecomposeResult, error) {
	// 清理响应内容
	content = strings.TrimSpace(content)

	// 移除markdown代码块标记
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// 查找JSON内容
	jsonStart := strings.Index(content, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("响应中未找到JSON内容")
	}
	jsonEnd := strings.LastIndex(content, "}")
	if jsonEnd == -1 {
		return nil, fmt.Errorf("响应中JSON未闭合")
	}
	jsonStr := content[jsonStart : jsonEnd+1]

	// 解析JSON
	var resp decompositionResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 转换为SubTask对象
	subTasks := make([]*SubTask, 0, len(resp.SubTasks))
	for _, jsonTask := range resp.SubTasks {
		task := NewSubTask(jsonTask.ID, jsonTask.Description)
		task.Details = jsonTask.Details
		task.Dependencies = jsonTask.Dependencies
		task.ParentID = req.ParentID
		task.EstimatedDuration = jsonTask.EstimatedDuration
		task.Assignee = jsonTask.Assignee

		// 解析优先级
		switch strings.ToUpper(jsonTask.Priority) {
		case "LOW":
			task.Priority = SubTaskPriorityLow
		case "HIGH":
			task.Priority = SubTaskPriorityHigh
		case "CRITICAL":
			task.Priority = SubTaskPriorityCritical
		default:
			task.Priority = SubTaskPriorityNormal
		}

		subTasks = append(subTasks, task)
	}

	return &DecomposeResult{
		SubTasks:                   subTasks,
		Reasoning:                  resp.Reasoning,
		NeedsFurtherDecomposition:  resp.NeedsFurtherDecomposition,
		EstimatedTotalTime:         resp.EstimatedTotalTime,
	}, nil
}

// AnalyzeDependencies 分析任务依赖关系，检测循环依赖
func (td *TaskDecomposer) AnalyzeDependencies(collection *TaskCollection) error {
	tasks := collection.ListTasks()

	// 构建依赖图
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.GetDependencies()
	}

	// 检测循环依赖
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for taskID := range graph {
		if !visited[taskID] {
			if td.hasCycle(taskID, graph, visited, recStack) {
				return fmt.Errorf("检测到循环依赖，涉及任务: %s", taskID)
			}
		}
	}

	return nil
}

// hasCycle 检测循环依赖（DFS）
func (td *TaskDecomposer) hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if td.hasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// GenerateTaskID 生成唯一的任务ID
func (td *TaskDecomposer) GenerateTaskID() string {
	td.taskCounter++
	return fmt.Sprintf("task-%d-%d", time.Now().Unix(), td.taskCounter)
}

// EstimateComplexity 估算任务复杂度
func (td *TaskDecomposer) EstimateComplexity(task string) (int, error) {
	// 使用模型估算任务复杂度
	prompt := fmt.Sprintf(
		"请评估以下任务的复杂度（1-10分，1最简单，10最复杂）：\n\n%s\n\n"+
			"只返回一个数字，不要其他内容。",
		task,
	)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是任务复杂度评估专家。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	ctx := context.Background()
	resp, err := td.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return 5, nil // 默认中等复杂度
	}

	// 解析复杂度分数
	content := strings.TrimSpace(resp.Content)
	var complexity int
	if _, err := fmt.Sscanf(content, "%d", &complexity); err != nil {
		return 5, nil // 默认中等复杂度
	}

	// 限制在1-10范围内
	if complexity < 1 {
		complexity = 1
	} else if complexity > 10 {
		complexity = 10
	}

	return complexity, nil
}

// ShouldDecompose 判断任务是否需要分解
func (td *TaskDecomposer) ShouldDecompose(task string, currentDepth int) (bool, error) {
	// 如果已经达到最大深度，不再分解
	if currentDepth >= td.maxDepth {
		return false, nil
	}

	// 估算任务复杂度
	complexity, err := td.EstimateComplexity(task)
	if err != nil {
		return false, err
	}

	// 复杂度大于6的任务需要分解
	return complexity > 6, nil
}
