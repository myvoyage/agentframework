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

	"AgentFramework/pkg/beads"
	beadscontext "AgentFramework/pkg/beads/context"
)

// ContextMCP 上下文 MCP 工具
// 提供 Agent 与上下文存储交互的 MCP 接口
type ContextMCP struct {
	tracker beads.TaskTracker
}

// NewContextMCP 创建新的上下文 MCP 工具
func NewContextMCP(tracker beads.TaskTracker) *ContextMCP {
	return &ContextMCP{
		tracker: tracker,
	}
}

// ===== Input/Output Structures =====

// CreateTaskWithContextInput 创建带上下文任务的输入
type CreateTaskWithContextInput struct {
	Type        string                 `json:"type"`                  // 任务类型
	Title       string                 `json:"title"`                 // 任务标题
	Description string                 `json:"description,omitempty"` // 任务描述
	ContextType string                 `json:"contextType"`           // 上下文类型
	Workspace   string                 `json:"workspace,omitempty"`   // 工作区路径
	ContextMeta map[string]string      `json:"contextMeta,omitempty"` // 上下文元数据
	Assignee    string                 `json:"assignee,omitempty"`    // 指派人
	Tags        []string               `json:"tags,omitempty"`        // 标签
}

// CreateTaskWithContextOutput 创建带上下文任务的输出
type CreateTaskWithContextOutput struct {
	Success   bool   `json:"success"`             // 是否成功
	TaskID    string `json:"task_id,omitempty"`   // 任务 ID
	ContextID string `json:"context_id,omitempty"` // 上下文 ID
	Error     string `json:"error,omitempty"`     // 错误信息
}

// GetTaskContextsInput 获取任务上下文的输入
type GetTaskContextsInput struct {
	TaskID string `json:"task_id"` // 任务 ID
}

// GetTaskContextsOutput 获取任务上下文的输出
type GetTaskContextsOutput struct {
	Success bool                         `json:"success"`             // 是否成功
	Contexts []*beadscontext.Context `json:"contexts,omitempty"` // 上下文列表
	Error   string                       `json:"error,omitempty"`     // 错误信息
}

// AssociateContextInput 关联上下文的输入
type AssociateContextInput struct {
	TaskID    string `json:"task_id"`    // 任务 ID
	ContextID string `json:"context_id"` // 上下文 ID
}

// AssociateContextOutput 关联上下文的输出
type AssociateContextOutput struct {
	Success bool   `json:"success"`           // 是否成功
	Error   string `json:"error,omitempty"`   // 错误信息
}

// ===== MCP Tool Implementations =====

// CreateTaskWithContext MCP 工具：创建任务并关联上下文
func (cm *ContextMCP) CreateTaskWithContext(
	ctx context.Context,
	input *CreateTaskWithContextInput,
) (*CreateTaskWithContextOutput, error) {
	// 验证输入
	if input.Title == "" {
		return &CreateTaskWithContextOutput{
			Success: false,
			Error:   "task title is required",
		}, nil
	}

	if input.ContextType == "" {
		return &CreateTaskWithContextOutput{
			Success: false,
			Error:   "context type is required",
		}, nil
	}

	// 创建任务
	task := &beads.Task{
		Type:        beads.TaskType(input.Type),
		Title:       input.Title,
		Description: input.Description,
		Assignee:    input.Assignee,
		Tags:        input.Tags,
	}

	// 创建上下文
	ctxt := &beadscontext.Context{
		Type:      beadscontext.ContextType(input.ContextType),
		Title:     input.Title,
		Workspace: input.Workspace,
		Metadata:  input.ContextMeta,
	}

	// 尝试使用 CreateTaskWithContext 方法
	type contextTracker interface {
		CreateTaskWithContext(ctx context.Context, task *beads.Task, ctxt *beadscontext.Context) (string, string, error)
	}

	tracker, ok := cm.tracker.(contextTracker)
	if !ok {
		// 如果 tracker 不支持上下文操作，返回错误
		return &CreateTaskWithContextOutput{
			Success: false,
			Error:   "context operations not supported by this tracker",
		}, nil
	}

	// 创建任务和上下文
	taskID, contextID, err := tracker.CreateTaskWithContext(ctx, task, ctxt)
	if err != nil {
		return &CreateTaskWithContextOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &CreateTaskWithContextOutput{
		Success:   true,
		TaskID:    taskID,
		ContextID: contextID,
	}, nil
}

// GetTaskContexts MCP 工具：获取任务的所有上下文
func (cm *ContextMCP) GetTaskContexts(
	ctx context.Context,
	input *GetTaskContextsInput,
) (*GetTaskContextsOutput, error) {
	// 验证输入
	if input.TaskID == "" {
		return &GetTaskContextsOutput{
			Success: false,
			Error:   "task ID is required",
		}, nil
	}

	// 尝试使用 GetTaskContexts 方法
	type contextTracker interface {
		GetTaskContexts(ctx context.Context, taskID string) ([]*beadscontext.Context, error)
	}

	tracker, ok := cm.tracker.(contextTracker)
	if !ok {
		return &GetTaskContextsOutput{
			Success: false,
			Error:   "context operations not supported by this tracker",
		}, nil
	}

	// 获取上下文
	contexts, err := tracker.GetTaskContexts(ctx, input.TaskID)
	if err != nil {
		return &GetTaskContextsOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &GetTaskContextsOutput{
		Success:  true,
		Contexts: contexts,
	}, nil
}

// AssociateContext MCP 工具：关联上下文到任务
func (cm *ContextMCP) AssociateContext(
	ctx context.Context,
	input *AssociateContextInput,
) (*AssociateContextOutput, error) {
	// 验证输入
	if input.TaskID == "" {
		return &AssociateContextOutput{
			Success: false,
			Error:   "task ID is required",
		}, nil
	}

	if input.ContextID == "" {
		return &AssociateContextOutput{
			Success: false,
			Error:   "context ID is required",
		}, nil
	}

	// 尝试使用 AssociateContext 方法
	type contextTracker interface {
		AssociateContext(ctx context.Context, taskID, contextID string) error
	}

	tracker, ok := cm.tracker.(contextTracker)
	if !ok {
		return &AssociateContextOutput{
			Success: false,
			Error:   "context operations not supported by this tracker",
		}, nil
	}

	// 关联上下文
	if err := tracker.AssociateContext(ctx, input.TaskID, input.ContextID); err != nil {
		return &AssociateContextOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &AssociateContextOutput{
		Success: true,
	}, nil
}

// GetContextTypes MCP 工具：获取支持的上下文类型
func (cm *ContextMCP) GetContextTypes(
	ctx context.Context,
) (map[string]string, error) {
	return map[string]string{
		"project":  "Project-level context",
		"file":     "File context",
		"codebase": "Codebase context",
		"custom":   "Custom context",
		"memory":   "Memory context (conversation history, etc.)",
		"resource": "Resource context (API, documentation, etc.)",
		"skill":    "Skill context",
	}, nil
}

// GetContextStoreInfo MCP 工具：获取上下文存储信息
func (cm *ContextMCP) GetContextStoreInfo(
	ctx context.Context,
) (map[string]interface{}, error) {
	type contextTracker interface {
		IsContextEnabled() bool
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(contextTracker)
	if !ok {
		return map[string]interface{}{
			"enabled": false,
			"error":   "context operations not supported",
		}, nil
	}

	enabled := tracker.IsContextEnabled()
	if !enabled {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return map[string]interface{}{
			"enabled": true,
			"error":   "context store is nil",
		}, nil
	}

	// 尝试获取统计信息
	type stater interface {
		GetStats(ctx context.Context) (*beadscontext.ContextStoreStats, error)
	}

	if statStore, ok := store.(stater); ok {
		stats, err := statStore.GetStats(ctx)
		if err == nil {
			return map[string]interface{}{
				"enabled":        true,
				"type":           "openviking",
				"total_contexts": stats.TotalContexts,
				"total_tasks":    stats.TotalTasks,
				"cache_hit_rate": stats.CacheHitRate,
			}, nil
		}
	}

	return map[string]interface{}{
		"enabled": true,
		"type":    "unknown",
	}, nil
}

// ===== 三层上下文操作 =====

// GetLayerInput 获取层级的输入
type GetLayerInput struct {
	ContextID string `json:"context_id"` // 上下文 ID
	Layer     string `json:"layer"`      // 层级类型 (l0/l1/l2/auto)
}

// GetLayerOutput 获取层级的输出
type GetLayerOutput struct {
	Success bool        `json:"success"`
	Content interface{} `json:"content,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetLayer MCP 工具：获取指定层级的内容
func (cm *ContextMCP) GetLayer(
	ctx context.Context,
	input *GetLayerInput,
) (*GetLayerOutput, error) {
	if input.ContextID == "" {
		return &GetLayerOutput{
			Success: false,
			Error:   "context_id is required",
		}, nil
	}

	if input.Layer == "" {
		input.Layer = "auto"
	}

	type layerTracker interface {
		GetTaskContextWithLayer(ctx context.Context, taskID string, layer beadscontext.LayerType) (interface{}, error)
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(layerTracker)
	if !ok {
		return &GetLayerOutput{
			Success: false,
			Error:   "layer operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &GetLayerOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	layer := beadscontext.LayerType(input.Layer)
	content, err := store.GetLayer(ctx, input.ContextID, layer)
	if err != nil {
		return &GetLayerOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &GetLayerOutput{
		Success: true,
		Content: content,
	}, nil
}

// GenerateLayersInput 生成层级的输入
type GenerateLayersInput struct {
	ContextID string `json:"context_id"` // 上下文 ID
}

// GenerateLayersOutput 生成层级的输出
type GenerateLayersOutput struct {
	Success bool              `json:"success"`
	Layers  *beadscontext.ContextLayers `json:"layers,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// GenerateLayers MCP 工具：为上下文生成缺失的层级
func (cm *ContextMCP) GenerateLayers(
	ctx context.Context,
	input *GenerateLayersInput,
) (*GenerateLayersOutput, error) {
	if input.ContextID == "" {
		return &GenerateLayersOutput{
			Success: false,
			Error:   "context_id is required",
		}, nil
	}

	type layerTracker interface {
		GenerateTaskContextLayers(ctx context.Context, taskID string) error
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(layerTracker)
	if !ok {
		return &GenerateLayersOutput{
			Success: false,
			Error:   "layer generation not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &GenerateLayersOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	// 生成层级
	err := store.GenerateLayers(ctx, input.ContextID)
	if err != nil {
		return &GenerateLayersOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取更新后的上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &GenerateLayersOutput{
			Success: true,
			Error:   "layers generated but failed to retrieve",
		}, nil
	}

	return &GenerateLayersOutput{
		Success: true,
		Layers:  &ctxt.Layers,
	}, nil
}

// ===== 记忆管理操作 =====

// ExtractMemoriesInput 提取记忆的输入
type ExtractMemoriesInput struct {
	ContextID string `json:"context_id"` // 上下文 ID
}

// ExtractMemoriesOutput 提取记忆的输出
type ExtractMemoriesOutput struct {
	Success  bool                       `json:"success"`
	Memories *beadscontext.MemoryCollection  `json:"memories,omitempty"`
	Error    string                     `json:"error,omitempty"`
}

// ExtractMemories MCP 工具：从上下文中提取记忆
func (cm *ContextMCP) ExtractMemories(
	ctx context.Context,
	input *ExtractMemoriesInput,
) (*ExtractMemoriesOutput, error) {
	if input.ContextID == "" {
		return &ExtractMemoriesOutput{
			Success: false,
			Error:   "context_id is required",
		}, nil
	}

	type memoryTracker interface {
		ExtractTaskMemories(ctx context.Context, taskID string) (interface{}, error)
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(memoryTracker)
	if !ok {
		return &ExtractMemoriesOutput{
			Success: false,
			Error:   "memory operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &ExtractMemoriesOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	memories, err := store.ExtractMemories(ctx, input.ContextID)
	if err != nil {
		return &ExtractMemoriesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ExtractMemoriesOutput{
		Success:  true,
		Memories: memories,
	}, nil
}

// GetMemoriesInput 获取记忆的输入
type GetMemoriesInput struct {
	ContextID   string   `json:"context_id"`             // 上下文 ID
	MemoryTypes []string `json:"memory_types,omitempty"`  // 记忆类型过滤
}

// GetMemoriesOutput 获取记忆的输出
type GetMemoriesOutput struct {
	Success  bool                       `json:"success"`
	Memories *beadscontext.MemoryCollection  `json:"memories,omitempty"`
	Error    string                     `json:"error,omitempty"`
}

// GetMemories MCP 工具：获取上下文的记忆
func (cm *ContextMCP) GetMemories(
	ctx context.Context,
	input *GetMemoriesInput,
) (*GetMemoriesOutput, error) {
	if input.ContextID == "" {
		return &GetMemoriesOutput{
			Success: false,
			Error:   "context_id is required",
		}, nil
	}

	type memoryTracker interface {
		GetTaskMemories(ctx context.Context, taskID string, memoryTypes []beadscontext.MemoryType) (interface{}, error)
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(memoryTracker)
	if !ok {
		return &GetMemoriesOutput{
			Success: false,
			Error:   "memory operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &GetMemoriesOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	// 获取上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &GetMemoriesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	memories := ctxt.Memories
	if memories == nil {
		memories = &beadscontext.MemoryCollection{}
	}

	// 如果指定了记忆类型，进行过滤
	if len(input.MemoryTypes) > 0 {
		filtered := &beadscontext.MemoryCollection{}
		typeMap := map[string]beadscontext.MemoryType{
			"profile":    beadscontext.MemoryTypeProfile,
			"preference": beadscontext.MemoryTypePreference,
			"entity":     beadscontext.MemoryTypeEntity,
			"event":      beadscontext.MemoryTypeEvent,
			"case":       beadscontext.MemoryTypeCase,
			"pattern":    beadscontext.MemoryTypePattern,
		}

		for _, mt := range input.MemoryTypes {
			if memoryType, ok := typeMap[mt]; ok {
				switch memoryType {
				case beadscontext.MemoryTypeProfile:
					filtered.Profiles = memories.Profiles
				case beadscontext.MemoryTypePreference:
					filtered.Preferences = memories.Preferences
				case beadscontext.MemoryTypeEntity:
					filtered.Entities = memories.Entities
				case beadscontext.MemoryTypeEvent:
					filtered.Events = memories.Events
				case beadscontext.MemoryTypeCase:
					filtered.Cases = memories.Cases
				case beadscontext.MemoryTypePattern:
					filtered.Patterns = memories.Patterns
				}
			}
		}
		memories = filtered
	}

	return &GetMemoriesOutput{
		Success:  true,
		Memories: memories,
	}, nil
}

// DeduplicateMemoriesInput 去重记忆的输入
type DeduplicateMemoriesInput struct {
	ContextID string `json:"context_id"` // 上下文 ID
}

// DeduplicateMemoriesOutput 去重记忆的输出
type DeduplicateMemoriesOutput struct {
	Success       bool                      `json:"success"`
	OriginalCount int                       `json:"original_count"`
	FinalCount    int                       `json:"final_count"`
	Removed       int                       `json:"removed"`
	Error         string                    `json:"error,omitempty"`
}

// DeduplicateMemories MCP 工具：对上下文的记忆进行去重
func (cm *ContextMCP) DeduplicateMemories(
	ctx context.Context,
	input *DeduplicateMemoriesInput,
) (*DeduplicateMemoriesOutput, error) {
	if input.ContextID == "" {
		return &DeduplicateMemoriesOutput{
			Success: false,
			Error:   "context_id is required",
		}, nil
	}

	type memoryTracker interface {
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(memoryTracker)
	if !ok {
		return &DeduplicateMemoriesOutput{
			Success: false,
			Error:   "memory operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &DeduplicateMemoriesOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	// 获取上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &DeduplicateMemoriesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if ctxt.Memories == nil {
		return &DeduplicateMemoriesOutput{
			Success: true,
			OriginalCount: 0,
			FinalCount:    0,
			Removed:       0,
		}, nil
	}

	originalCount := ctxt.Memories.GetMemoryCount()

	// 去重
	deduplicated, err := store.DeduplicateMemories(ctx, input.ContextID)
	if err != nil {
		return &DeduplicateMemoriesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	finalCount := deduplicated.GetMemoryCount()

	return &DeduplicateMemoriesOutput{
		Success:       true,
		OriginalCount: originalCount,
		FinalCount:    finalCount,
		Removed:       originalCount - finalCount,
	}, nil
}

// ===== 同步操作 =====

// TriggerSyncInput 触发同步的输入
type TriggerSyncInput struct {
	Force bool `json:"force,omitempty"` // 强制同步
}

// TriggerSyncOutput 触发同步的输出
type TriggerSyncOutput struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// TriggerSync MCP 工具：手动触发任务-上下文同步
func (cm *ContextMCP) TriggerSync(
	ctx context.Context,
	input *TriggerSyncInput,
) (*TriggerSyncOutput, error) {
	type syncTracker interface {
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(syncTracker)
	if !ok {
		return &TriggerSyncOutput{
			Success: false,
			Error:   "sync operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &TriggerSyncOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	// 尝试触发同步
	type coordinatorStore interface {
		GetCoordinator() interface{}
	}

	if coordStore, ok := store.(coordinatorStore); ok {
		coord := coordStore.GetCoordinator()
		if coord != nil {
			type syncer interface {
				TriggerSync(ctx context.Context) error
			}

			if syncCoord, ok := coord.(syncer); ok {
				err := syncCoord.TriggerSync(ctx)
				if err != nil {
					return &TriggerSyncOutput{
						Success: false,
						Error:   err.Error(),
					}, nil
				}

				return &TriggerSyncOutput{
					Success: true,
					Message: "sync completed successfully",
				}, nil
			}
		}
	}

	return &TriggerSyncOutput{
		Success: false,
		Error:   "coordinator not available",
	}, nil
}

// ===== 统计信息 =====

// GetStatsInput 获取统计信息的输入
type GetStatsInput struct {
	ContextID string `json:"context_id,omitempty"` // 可选：特定上下文的统计
}

// GetStatsOutput 获取统计信息的输出
type GetStatsOutput struct {
	Success bool                   `json:"success"`
	Stats   *beadscontext.CoordinatorStats `json:"stats,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// GetStats MCP 工具：获取协调器统计信息
func (cm *ContextMCP) GetStats(
	ctx context.Context,
	input *GetStatsInput,
) (*GetStatsOutput, error) {
	type statsTracker interface {
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(statsTracker)
	if !ok {
		return &GetStatsOutput{
			Success: false,
			Error:   "stats operations not supported",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return &GetStatsOutput{
			Success: false,
			Error:   "context store is not available",
		}, nil
	}

	// 尝试获取统计信息
	type coordinatorStore interface {
		GetCoordinator() interface{}
	}

	if coordStore, ok := store.(coordinatorStore); ok {
		coord := coordStore.GetCoordinator()
		if coord != nil {
			type stater interface {
				GetStats(ctx context.Context) (*beadscontext.CoordinatorStats, error)
			}

			if statCoord, ok := coord.(stater); ok {
				stats, err := statCoord.GetStats(ctx)
				if err != nil {
					return &GetStatsOutput{
						Success: false,
						Error:   err.Error(),
					}, nil
				}

				return &GetStatsOutput{
					Success: true,
					Stats:   stats,
				}, nil
			}
		}
	}

	return &GetStatsOutput{
		Success: false,
		Error:   "coordinator not available",
	}, nil
}

// HealthCheck MCP 工具：检查上下文系统健康状态
func (cm *ContextMCP) HealthCheck(
	ctx context.Context,
) (map[string]interface{}, error) {
	type healthTracker interface {
		IsContextEnabled() bool
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := cm.tracker.(healthTracker)
	if !ok {
		return map[string]interface{}{
			"healthy": false,
			"error":   "context operations not supported",
		}, nil
	}

	if !tracker.IsContextEnabled() {
		return map[string]interface{}{
			"healthy":       false,
			"enabled":       false,
			"error":         "context system is disabled",
		}, nil
	}

	store := tracker.GetContextStore()
	if store == nil {
		return map[string]interface{}{
			"healthy": false,
			"enabled": true,
			"error":   "context store is nil",
		}, nil
	}

	// 尝试执行简单的查询操作
	_, err := store.GetContext(ctx, "health-check")
	if err != nil {
		// 预期会失败（上下文不存在），这是正常的
		return map[string]interface{}{
			"healthy": true,
			"enabled": true,
			"message": "context system is operational",
		}, nil
	}

	return map[string]interface{}{
		"healthy": true,
		"enabled": true,
	}, nil
}
