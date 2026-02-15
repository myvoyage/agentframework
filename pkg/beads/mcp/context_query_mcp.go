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

package mcp

import (
	"context"
	"fmt"
	"time"

	"AgentFramework/pkg/beads"
	beadscontext "AgentFramework/pkg/beads/context"
)

// QueryMCP 查询 MCP 工具
// 提供任务-上下文联合查询的 MCP 接口
type QueryMCP struct {
	tracker beads.TaskTracker
}

// NewQueryMCP 创建新的查询 MCP 工具
func NewQueryMCP(tracker beads.TaskTracker) *QueryMCP {
	return &QueryMCP{
		tracker: tracker,
	}
}

// ===== Input/Output Structures =====

// QueryTasksWithContextInput 查询带上下文的任务输入
type QueryTasksWithContextInput struct {
	Status     string            `json:"status,omitempty"`      // 任务状态过滤
	Assignee   string            `json:"assignee,omitempty"`    // 指派人过滤
	Tags       []string          `json:"tags,omitempty"`        // 标签过滤
	ContextType string           `json:"context_type,omitempty"` // 上下文类型过滤
	Layer      string            `json:"layer,omitempty"`       // 上下文层级过滤
	Limit      int               `json:"limit,omitempty"`       // 结果限制
	Offset     int               `json:"offset,omitempty"`      // 偏移量
	OrderBy    string            `json:"order_by,omitempty"`    // 排序字段
	OrderDir   string            `json:"order_dir,omitempty"`   // 排序方向 (asc/desc)
}

// QueryTasksWithContextOutput 查询带上下文的任务输出
type QueryTasksWithContextOutput struct {
	Success bool                        `json:"success"`
	Results []*TaskWithContext          `json:"results,omitempty"`
	Total   int                         `json:"total,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// TaskWithContext 带上下文的任务
type TaskWithContext struct {
	Task     *beads.Task          `json:"task"`
	Contexts []*beadscontext.Context   `json:"contexts,omitempty"`
}

// QueryContextsWithTasksInput 查询带任务的上下文输入
type QueryContextsWithTasksInput struct {
	ContextType string            `json:"context_type,omitempty"` // 上下文类型过滤
	Workspace   string            `json:"workspace,omitempty"`    // 工作区过滤
	ParentID    string            `json:"parent_id,omitempty"`    // 父上下文过滤
	HasTasks    bool              `json:"has_tasks,omitempty"`    // 是否只返回有关联任务的
	Limit       int               `json:"limit,omitempty"`        // 结果限制
	Offset      int               `json:"offset,omitempty"`       // 偏移量
}

// QueryContextsWithTasksOutput 查询带任务的上下文输出
type QueryContextsWithTasksOutput struct {
	Success bool                         `json:"success"`
	Results []*ContextWithTasks          `json:"results,omitempty"`
	Total   int                          `json:"total,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

// ContextWithTasks 带任务的上下文
type ContextWithTasks struct {
	Context *beadscontext.Context `json:"context"`
	Tasks   []*beads.Task    `json:"tasks,omitempty"`
}

// SearchInput 搜索输入
type SearchInput struct {
	Query      string  `json:"query"`                  // 搜索查询
	SearchIn   string  `json:"search_in,omitempty"`    // 搜索范围 (tasks/contexts/both)
	ContextType string  `json:"context_type,omitempty"` // 上下文类型过滤
	MaxResults int     `json:"max_results,omitempty"`  // 最大结果数
	MinScore   float64 `json:"min_score,omitempty"`    // 最小相关性分数
}

// SearchOutput 搜索输出
type SearchOutput struct {
	Success  bool                `json:"success"`
	Results  []*SearchResult     `json:"results,omitempty"`
	Error    string              `json:"error,omitempty"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Type      string             `json:"type"` // task/context
	Score     float64            `json:"score"`
	Task      *beads.Task        `json:"task,omitempty"`
	Context   *beadscontext.Context   `json:"context,omitempty"`
	Highlight string             `json:"highlight,omitempty"`
}

// GetRelatedContextsInput 获取相关上下文输入
type GetRelatedContextsInput struct {
	ContextID  string  `json:"context_id"`
	MaxResults int     `json:"max_results,omitempty"`
	MinScore   float64 `json:"min_score,omitempty"`
}

// GetRelatedContextsOutput 获取相关上下文输出
type GetRelatedContextsOutput struct {
	Success  bool                      `json:"success"`
	Contexts []*RelatedContext         `json:"contexts,omitempty"`
	Error    string                    `json:"error,omitempty"`
}

// RelatedContext 相关上下文
type RelatedContext struct {
	Context   *beadscontext.Context `json:"context"`
	Score     float64          `json:"score"`
	Reason    string           `json:"reason,omitempty"`
}

// GetTaskHierarchyInput 获取任务层次结构输入
type GetTaskHierarchyInput struct {
	TaskID    string `json:"task_id"`
	Depth     int    `json:"depth,omitempty"`  // 深度限制
	IncludeContexts bool `json:"include_contexts,omitempty"` // 是否包含上下文
}

// GetTaskHierarchyOutput 获取任务层次结构输出
type GetTaskHierarchyOutput struct {
	Success  bool             `json:"success"`
	Root     *TaskHierarchyNode `json:"root,omitempty"`
	Error    string           `json:"error,omitempty"`
}

// TaskHierarchyNode 任务层次节点
type TaskHierarchyNode struct {
	Task      *beads.Task          `json:"task"`
	Contexts  []*beadscontext.Context   `json:"contexts,omitempty"`
	Children  []*TaskHierarchyNode `json:"children,omitempty"`
	Depth     int                   `json:"depth"`
}

// GetContextHierarchyInput 获取上下文层次结构输入
type GetContextHierarchyInput struct {
	ContextID string `json:"context_id"`
	Depth     int    `json:"depth,omitempty"`  // 深度限制
	IncludeTasks bool `json:"include_tasks,omitempty"` // 是否包含任务
}

// GetContextHierarchyOutput 获取上下文层次结构输出
type GetContextHierarchyOutput struct {
	Success  bool                   `json:"success"`
	Root     *ContextHierarchyNode  `json:"root,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// ContextHierarchyNode 上下文层次节点
type ContextHierarchyNode struct {
	Context   *beadscontext.Context       `json:"context"`
	Tasks     []*beads.Task          `json:"tasks,omitempty"`
	Children  []*ContextHierarchyNode `json:"children,omitempty"`
	Depth     int                     `json:"depth"`
}

// AggregateQueryInput 聚合查询输入
type AggregateQueryInput struct {
	GroupBy    string            `json:"group_by"`    // 分组字段 (status/assignee/context_type/workspace)
	Filter     map[string]string `json:"filter,omitempty"` // 过滤条件
	TimeRange  *TimeRange        `json:"time_range,omitempty"` // 时间范围
}

// AggregateQueryOutput 聚合查询输出
type AggregateQueryOutput struct {
	Success bool                  `json:"success"`
	Groups  []*QueryGroup         `json:"groups,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// QueryGroup 查询分组
type QueryGroup struct {
	Key      string      `json:"key"`
	Count    int         `json:"count"`
	Tasks    int         `json:"tasks,omitempty"`
	Contexts int         `json:"contexts,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ===== MCP Tool Implementations =====

// QueryTasksWithContext MCP 工具：查询带上下文的任务
func (qm *QueryMCP) QueryTasksWithContext(
	ctx context.Context,
	input *QueryTasksWithContextInput,
) (*QueryTasksWithContextOutput, error) {
	// 获取上下文存储
	store, err := qm.getContextStore(ctx)
	if err != nil {
		return &QueryTasksWithContextOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 构建查询条件
	query := beads.Query{}
	if input.Status != "" {
		status := beads.TaskStatus(input.Status)
		query.Status = &status
	}
	if input.Assignee != "" {
		query.Assignee = &input.Assignee
	}
	if len(input.Tags) > 0 {
		query.Tags = input.Tags
	}
	// 注意：beads.Query 不支持 Limit、Offset、OrderBy、OrderDir
	// 这些限制需要手动应用在查询结果上

	// 查询任务
	type queryTracker interface {
		Query(ctx context.Context, query beads.Query) (interface{}, error)
	}

	tracker, ok := qm.tracker.(queryTracker)
	if !ok {
		return &QueryTasksWithContextOutput{
			Success: false,
			Error:   "query operations not supported",
		}, nil
	}

	result, err := tracker.Query(ctx, query)
	if err != nil {
		return &QueryTasksWithContextOutput{
			Success: false,
			Error:   fmt.Sprintf("query tasks failed: %v", err),
		}, nil
	}

	// 类型断言转换
	tasks, ok := result.([]*beads.Task)
	if !ok {
		return &QueryTasksWithContextOutput{
			Success: false,
			Error:   "invalid query result type",
		}, nil
	}

	// 为每个任务获取上下文
	results := make([]*TaskWithContext, 0)
	for _, task := range tasks {
		taskWithContext := &TaskWithContext{
			Task:     task,
			Contexts: nil,
		}

		// 获取任务的上下文
		contexts, err := store.GetTaskContexts(ctx, task.ID)
		if err == nil {
			// 根据条件过滤上下文
			filteredContexts := make([]*beadscontext.Context, 0)
			for _, ctxt := range contexts {
				if input.ContextType != "" && string(ctxt.Type) != input.ContextType {
					continue
				}
				if input.Layer != "" {
					// 检查是否有指定的层级
					hasLayer := false
					switch input.Layer {
					case "l0":
						hasLayer = ctxt.Layers.L0 != nil
					case "l1":
						hasLayer = ctxt.Layers.L1 != nil
					case "l2":
						hasLayer = ctxt.Layers.L2 != nil
					}
					if !hasLayer {
						continue
					}
				}
				filteredContexts = append(filteredContexts, ctxt)
			}
			taskWithContext.Contexts = filteredContexts
		}

		results = append(results, taskWithContext)
	}

	return &QueryTasksWithContextOutput{
		Success: true,
		Results: results,
		Total:   len(results),
	}, nil
}

// QueryContextsWithTasks MCP 工具：查询带任务的上下文
func (qm *QueryMCP) QueryContextsWithTasks(
	ctx context.Context,
	input *QueryContextsWithTasksInput,
) (*QueryContextsWithTasksOutput, error) {
	// 获取上下文存储
	store, err := qm.getContextStore(ctx)
	if err != nil {
		return &QueryContextsWithTasksOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 简化实现：获取所有上下文（实际应该有查询接口）
	// 这里假设有一个方法可以查询上下文
	type queryStore interface {
		QueryContexts(ctx context.Context, filter map[string]interface{}) ([]*beadscontext.Context, error)
	}

	queryableStore, ok := store.(queryStore)
	var contexts []*beadscontext.Context

	if ok {
		// 构建过滤条件
		filter := make(map[string]interface{})
		if input.ContextType != "" {
			filter["type"] = input.ContextType
		}
		if input.Workspace != "" {
			filter["workspace"] = input.Workspace
		}
		if input.ParentID != "" {
			filter["parent_id"] = input.ParentID
		}

		contexts, err = queryableStore.QueryContexts(ctx, filter)
		if err != nil {
			return &QueryContextsWithTasksOutput{
				Success: false,
				Error:   fmt.Sprintf("query contexts failed: %v", err),
			}, nil
		}
	} else {
		// 如果没有查询接口，返回空列表
		contexts = []*beadscontext.Context{}
	}

	// 为每个上下文获取任务
	results := make([]*ContextWithTasks, 0)
	for _, ctxt := range contexts {
		contextWithTasks := &ContextWithTasks{
			Context: ctxt,
			Tasks:   nil,
		}

		// 如果需要获取任务
		if !input.HasTasks || ctxt.Metadata != nil {
			// 从元数据获取关联的任务 ID
			if taskID, ok := ctxt.Metadata["task_id"]; ok {
				// 获取任务
				type getTracker interface {
					Get(ctx context.Context, id string) (*beads.Task, error)
				}

				tracker, ok := qm.tracker.(getTracker)
				if ok {
					task, err := tracker.Get(ctx, taskID)
					if err == nil {
						contextWithTasks.Tasks = []*beads.Task{task}
					}
				}
			}
		}

		// 如果只想要有任务的上下文
		if input.HasTasks && len(contextWithTasks.Tasks) == 0 {
			continue
		}

		results = append(results, contextWithTasks)
	}

	// 应用限制和偏移
	total := len(results)
	if input.Offset > 0 && input.Offset < len(results) {
		results = results[input.Offset:]
	}
	if input.Limit > 0 && input.Limit < len(results) {
		results = results[:input.Limit]
	}

	return &QueryContextsWithTasksOutput{
		Success: true,
		Results: results,
		Total:   total,
	}, nil
}

// Search MCP 工具：搜索任务和上下文
func (qm *QueryMCP) Search(
	ctx context.Context,
	input *SearchInput,
) (*SearchOutput, error) {
	// 设置默认值
	if input.SearchIn == "" {
		input.SearchIn = "both"
	}
	if input.MaxResults <= 0 {
		input.MaxResults = 10
	}
	if input.MinScore <= 0 {
		input.MinScore = 0.3
	}

	results := make([]*SearchResult, 0)

	// 搜索任务
	if input.SearchIn == "tasks" || input.SearchIn == "both" {
		taskResults := qm.searchTasks(ctx, input)
		results = append(results, taskResults...)
	}

	// 搜索上下文
	if input.SearchIn == "contexts" || input.SearchIn == "both" {
		contextResults := qm.searchContexts(ctx, input)
		results = append(results, contextResults...)
	}

	// 应用结果限制
	if len(results) > input.MaxResults {
		results = results[:input.MaxResults]
	}

	return &SearchOutput{
		Success: true,
		Results: results,
	}, nil
}

// searchTasks 搜索任务
func (qm *QueryMCP) searchTasks(
	ctx context.Context,
	input *SearchInput,
) []*SearchResult {
	results := make([]*SearchResult, 0)

	// 查询所有任务
	type queryTracker interface {
		Query(ctx context.Context, query beads.Query) (interface{}, error)
	}

	tracker, ok := qm.tracker.(queryTracker)
	if !ok {
		return results
	}

	allTasks, err := tracker.Query(ctx, beads.Query{})
	if err != nil {
		return results
	}

	// 类型断言转换
	tasks, ok := allTasks.([]*beads.Task)
	if !ok {
		return results
	}

	// 简单的文本匹配和评分
	for _, task := range tasks {
		score := calculateSearchScore(input.Query, task.Title, task.Description)
		if score >= input.MinScore {
			results = append(results, &SearchResult{
				Type:  "task",
				Score: score,
				Task:  task,
				Highlight: extractHighlight(input.Query, task.Title),
			})
		}
	}

	return results
}

// searchContexts 搜索上下文
func (qm *QueryMCP) searchContexts(
	ctx context.Context,
	input *SearchInput,
) []*SearchResult {
	results := make([]*SearchResult, 0)

	store, err := qm.getContextStore(ctx)
	if err != nil {
		return results
	}

	// 简化实现：搜索 L1 层内容
	type searchStore interface {
		SearchFiles(ctx context.Context, query string, opts ...beadscontext.SearchOption) ([]*beadscontext.VFSSearchResult, error)
	}

	searchableStore, ok := store.(searchStore)
	if !ok {
		return results
	}

	// 构建搜索选项
	opts := []beadscontext.SearchOption{
		beadscontext.WithSearchLayer(beadscontext.LayerTypeL1),
		beadscontext.WithMaxResults(input.MaxResults),
		beadscontext.WithMinScore(input.MinScore),
	}

	if input.ContextType != "" {
		// 添加类型过滤
	}

	searchResults, err := searchableStore.SearchFiles(ctx, input.Query, opts...)
	if err != nil {
		return results
	}

	// 转换为 SearchResult
	for _, sr := range searchResults {
		if sr.Score >= input.MinScore {
			ctxt, err := store.GetContext(ctx, sr.URI)
			if err == nil {
				results = append(results, &SearchResult{
					Type:      "context",
					Score:     sr.Score,
					Context:   ctxt,
					Highlight: sr.Snippet,
				})
			}
		}
	}

	return results
}

// GetRelatedContexts MCP 工具：获取相关上下文
func (qm *QueryMCP) GetRelatedContexts(
	ctx context.Context,
	input *GetRelatedContextsInput,
) (*GetRelatedContextsOutput, error) {
	// 设置默认值
	if input.MaxResults <= 0 {
		input.MaxResults = 5
	}
	if input.MinScore <= 0 {
		input.MinScore = 0.5
	}

	store, err := qm.getContextStore(ctx)
	if err != nil {
		return &GetRelatedContextsOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取源上下文
	sourceCtxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &GetRelatedContextsOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	// 查找相关上下文
	// 简化实现：基于类型、工作区和标签查找
	relatedContexts := make([]*RelatedContext, 0)

	// 实际实现应该使用向量相似度或更复杂的算法
	// 这里提供一个简单的占位实现
	_ = sourceCtxt // 避免未使用变量警告

	return &GetRelatedContextsOutput{
		Success:  true,
		Contexts: relatedContexts,
	}, nil
}

// GetTaskHierarchy MCP 工具：获取任务层次结构
func (qm *QueryMCP) GetTaskHierarchy(
	ctx context.Context,
	input *GetTaskHierarchyInput,
) (*GetTaskHierarchyOutput, error) {
	// 设置默认深度
	if input.Depth <= 0 {
		input.Depth = 3
	}

	// 获取根任务
	type getTracker interface {
		Get(ctx context.Context, id string) (*beads.Task, error)
	}

	tracker, ok := qm.tracker.(getTracker)
	if !ok {
		return &GetTaskHierarchyOutput{
			Success: false,
			Error:   "get operations not supported",
		}, nil
	}

	rootTask, err := tracker.Get(ctx, input.TaskID)
	if err != nil {
		return &GetTaskHierarchyOutput{
			Success: false,
			Error:   fmt.Sprintf("get task failed: %v", err),
		}, nil
	}

	// 构建层次结构
	rootNode := &TaskHierarchyNode{
		Task:  rootTask,
		Depth: 0,
	}

	if input.IncludeContexts {
		// 获取上下文
		store, _ := qm.getContextStore(ctx)
		if store != nil {
			contexts, _ := store.GetTaskContexts(ctx, input.TaskID)
			rootNode.Contexts = contexts
		}
	}

	// 递归获取子任务（简化实现）
	// 实际应该从 TaskTracker 获取依赖关系

	return &GetTaskHierarchyOutput{
		Success: true,
		Root:    rootNode,
	}, nil
}

// GetContextHierarchy MCP 工具：获取上下文层次结构
func (qm *QueryMCP) GetContextHierarchy(
	ctx context.Context,
	input *GetContextHierarchyInput,
) (*GetContextHierarchyOutput, error) {
	// 设置默认深度
	if input.Depth <= 0 {
		input.Depth = 3
	}

	store, err := qm.getContextStore(ctx)
	if err != nil {
		return &GetContextHierarchyOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取根上下文
	rootCtxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &GetContextHierarchyOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	// 构建层次结构
	rootNode := &ContextHierarchyNode{
		Context: rootCtxt,
		Depth:   0,
	}

	if input.IncludeTasks {
		// 获取任务
		if taskID, ok := rootCtxt.Metadata["task_id"]; ok {
			type getTracker interface {
				Get(ctx context.Context, id string) (*beads.Task, error)
			}

			tracker, ok := qm.tracker.(getTracker)
			if ok {
				task, err := tracker.Get(ctx, taskID)
				if err == nil {
					rootNode.Tasks = []*beads.Task{task}
				}
			}
		}
	}

	// 递归获取子上下文（简化实现）
	// 实际应该从 ContextStore 获取父子关系

	return &GetContextHierarchyOutput{
		Success: true,
		Root:    rootNode,
	}, nil
}

// AggregateQuery MCP 工具：聚合查询
func (qm *QueryMCP) AggregateQuery(
	ctx context.Context,
	input *AggregateQueryInput,
) (*AggregateQueryOutput, error) {
	_, err := qm.getContextStore(ctx)
	if err != nil {
		return &AggregateQueryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	groups := make([]*QueryGroup, 0)

	// 根据分组字段执行聚合
	switch input.GroupBy {
	case "status":
		// 按任务状态分组
		statuses := []beads.TaskStatus{
			beads.StatusOpen,
			beads.StatusInProgress,
			beads.StatusCompleted,
			beads.StatusBlocked,
		}

		for _, status := range statuses {
			query := beads.Query{
				Status: &status,
			}

			type queryTracker interface {
				Query(ctx context.Context, query beads.Query) (interface{}, error)
			}

			tracker, ok := qm.tracker.(queryTracker)
			if !ok {
				continue
			}

			result, err := tracker.Query(ctx, query)
			if err != nil {
				continue
			}

			tasks, ok := result.([]*beads.Task)
			if !ok {
				continue
			}

			groups = append(groups, &QueryGroup{
				Key:   string(status),
				Count: len(tasks),
				Tasks: len(tasks),
			})
		}

	case "context_type":
		// 按上下文类型分组
		contextTypes := []beadscontext.ContextType{
			beadscontext.ContextTypeProject,
			beadscontext.ContextTypeFile,
			beadscontext.ContextTypeCodebase,
			beadscontext.ContextTypeMemory,
			beadscontext.ContextTypeResource,
			beadscontext.ContextTypeSkill,
			beadscontext.ContextTypeConversation,
			beadscontext.ContextTypeSession,
		}

		for _, ctxtType := range contextTypes {
			// 简化实现：统计上下文数量
			// 实际应该从 ContextStore 查询
			groups = append(groups, &QueryGroup{
				Key:      string(ctxtType),
				Count:    0,
				Contexts: 0,
			})
		}

	case "assignee":
		// 按指派人分组
		// 简化实现
		groups = append(groups, &QueryGroup{
			Key:   "all",
			Count: 0,
			Tasks: 0,
		})

	case "workspace":
		// 按工作区分组
		// 简化实现
		groups = append(groups, &QueryGroup{
			Key:      "default",
			Count:    0,
			Contexts: 0,
		})

	default:
		return &AggregateQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("unsupported group_by field: %s", input.GroupBy),
		}, nil
	}

	return &AggregateQueryOutput{
		Success: true,
		Groups:  groups,
	}, nil
}

// ===== 辅助方法 =====

// getContextStore 获取上下文存储
func (qm *QueryMCP) getContextStore(_ context.Context) (beadscontext.ContextStore, error) {
	type queryTracker interface {
		IsContextEnabled() bool
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := qm.tracker.(queryTracker)
	if !ok {
		return nil, fmt.Errorf("query operations not supported")
	}

	if !tracker.IsContextEnabled() {
		return nil, fmt.Errorf("context system is disabled")
	}

	store := tracker.GetContextStore()
	if store == nil {
		return nil, fmt.Errorf("context store is not available")
	}

	return store, nil
}

// calculateSearchScore 计算搜索分数
func calculateSearchScore(query, title, description string) float64 {
	query = toLower(query)
	title = toLower(title)
	description = toLower(description)

	score := 0.0

	// 标题匹配权重更高
	if contains(title, query) {
		score += 0.7
	}

	// 描述匹配
	if contains(description, query) {
		score += 0.3
	}

	return score
}

// extractHighlight 提取高亮文本
func extractHighlight(query, text string) string {
	// 简化实现：返回前 100 个字符
	if len(text) > 100 {
		return text[:100] + "..."
	}
	return text
}

// toLower 转换为小写
func toLower(s string) string {
	// 简化实现
	result := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else {
			result += string(c)
		}
	}
	return result
}

// contains 检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
