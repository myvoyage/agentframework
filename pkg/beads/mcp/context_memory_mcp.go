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

// MemoryMCP 记忆管理 MCP 工具
// 提供 Agent 与记忆系统交互的 MCP 接口
type MemoryMCP struct {
	tracker beads.TaskTracker
}

// NewMemoryMCP 创建新的记忆 MCP 工具
func NewMemoryMCP(tracker beads.TaskTracker) *MemoryMCP {
	return &MemoryMCP{
		tracker: tracker,
	}
}

// ===== Input/Output Structures =====

// AddProfileMemoryInput 添加用户画像记忆的输入
type AddProfileMemoryInput struct {
	ContextID string            `json:"context_id"`
	Name      string            `json:"name"`
	Role      string            `json:"role,omitempty"`
	Traits    map[string]string `json:"traits,omitempty"`
	Goals     []string          `json:"goals,omitempty"`
}

// AddProfileMemoryOutput 添加用户画像记忆的输出
type AddProfileMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddPreferenceMemoryInput 添加偏好记忆的输入
type AddPreferenceMemoryInput struct {
	ContextID  string  `json:"context_id"`
	Category   string  `json:"category"`
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence,omitempty"`
}

// AddPreferenceMemoryOutput 添加偏好记忆的输出
type AddPreferenceMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddEntityMemoryInput 添加实体记忆的输入
type AddEntityMemoryInput struct {
	ContextID  string                 `json:"context_id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Attributes map[string]string      `json:"attributes,omitempty"`
}

// AddEntityMemoryOutput 添加实体记忆的输出
type AddEntityMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddEventMemoryInput 添加事件记忆的输入
type AddEventMemoryInput struct {
	ContextID   string `json:"context_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// AddEventMemoryOutput 添加事件记忆的输出
type AddEventMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddCaseMemoryInput 添加案例记忆的输入
type AddCaseMemoryInput struct {
	ContextID string   `json:"context_id"`
	Domain    string   `json:"domain"`
	Problem   string   `json:"problem"`
	Solution  string   `json:"solution"`
	Lessons   []string `json:"lessons,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// AddCaseMemoryOutput 添加案例记忆的输出
type AddCaseMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddPatternMemoryInput 添加模式记忆的输入
type AddPatternMemoryInput struct {
	ContextID string  `json:"context_id"`
	Category  string  `json:"category"`
	Pattern   string  `json:"pattern"`
	Confidence float64 `json:"confidence,omitempty"`
}

// AddPatternMemoryOutput 添加模式记忆的输出
type AddPatternMemoryOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GetMemoryByIDInput 获取单个记忆的输入
type GetMemoryByIDInput struct {
	ContextID string `json:"context_id"`
	MemoryID  string `json:"memory_id"`
	Type      string `json:"type"` // profile/preference/entity/event/case/pattern
}

// GetMemoryByIDOutput 获取单个记忆的输出
type GetMemoryByIDOutput struct {
	Success bool        `json:"success"`
	Memory  interface{} `json:"memory,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UpdateMemoryInput 更新记忆的输入
type UpdateMemoryInput struct {
	ContextID string                 `json:"context_id"`
	MemoryID  string                 `json:"memory_id"`
	Type      string                 `json:"type"` // profile/preference/entity/event/case/pattern
	Updates   map[string]interface{} `json:"updates"`
}

// UpdateMemoryOutput 更新记忆的输出
type UpdateMemoryOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DeleteMemoryInput 删除记忆的输入
type DeleteMemoryInput struct {
	ContextID string `json:"context_id"`
	MemoryID  string `json:"memory_id"`
	Type      string `json:"type"` // profile/preference/entity/event/case/pattern
}

// DeleteMemoryOutput 删除记忆的输出
type DeleteMemoryOutput struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// GetMemoryStatsInput 获取记忆统计的输入
type GetMemoryStatsInput struct {
	ContextID string `json:"context_id,omitempty"` // 可选，不提供则返回全局统计
}

// GetMemoryStatsOutput 获取记忆统计的输出
type GetMemoryStatsOutput struct {
	Success bool                      `json:"success"`
	Stats   *beadscontext.MemoryStats      `json:"stats,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// SearchMemoriesInput 搜索记忆的输入
type SearchMemoriesInput struct {
	ContextID   string `json:"context_id"`
	Query       string `json:"query"`
	Type        string `json:"type,omitempty"`  // 可选，按类型过滤
	MaxResults  int    `json:"max_results,omitempty"`
}

// SearchMemoriesOutput 搜索记忆的输出
type SearchMemoriesOutput struct {
	Success bool                        `json:"success"`
	Results []interface{}               `json:"results,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// MergeMemoriesInput 合并记忆的输入
type MergeMemoriesInput struct {
	SourceContextID string `json:"source_context_id"`
	TargetContextID string `json:"target_context_id"`
}

// MergeMemoriesOutput 合并记忆的输出
type MergeMemoriesOutput struct {
	Success        bool   `json:"success"`
	AddedCount     int    `json:"added_count,omitempty"`
	MergedCount    int    `json:"merged_count,omitempty"`
	Error          string `json:"error,omitempty"`
}

// ===== MCP Tool Implementations =====

// AddProfileMemory MCP 工具：添加用户画像记忆
func (mm *MemoryMCP) AddProfileMemory(
	ctx context.Context,
	input *AddProfileMemoryInput,
) (*AddProfileMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddProfileMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 创建记忆
	memory := &beadscontext.ProfileMemory{
		ID:        generateMemoryID("profile", input.Name),
		Name:      input.Name,
		Role:      input.Role,
		Traits:    input.Traits,
		Goals:     input.Goals,
		UpdatedAt: time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddProfileMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Profiles = append(ctxt.Memories.Profiles, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddProfileMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddProfileMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// AddPreferenceMemory MCP 工具：添加偏好记忆
func (mm *MemoryMCP) AddPreferenceMemory(
	ctx context.Context,
	input *AddPreferenceMemoryInput,
) (*AddPreferenceMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddPreferenceMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 设置默认置信度
	if input.Confidence == 0 {
		input.Confidence = 0.7
	}

	// 创建记忆
	memory := &beadscontext.PreferenceMemory{
		ID:        generateMemoryID("preference", input.Category+":"+input.Key),
		Category:  input.Category,
		Key:       input.Key,
		Value:     input.Value,
		Confidence: input.Confidence,
		UpdatedAt: time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddPreferenceMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Preferences = append(ctxt.Memories.Preferences, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddPreferenceMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddPreferenceMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// AddEntityMemory MCP 工具：添加实体记忆
func (mm *MemoryMCP) AddEntityMemory(
	ctx context.Context,
	input *AddEntityMemoryInput,
) (*AddEntityMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddEntityMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 创建记忆
	memory := &beadscontext.EntityMemory{
		ID:         generateMemoryID("entity", input.Type+":"+input.Name),
		Type:       input.Type,
		Name:       input.Name,
		Attributes: input.Attributes,
		Relations:  []beadscontext.EntityRelation{},
		FirstSeen:  time.Now(),
		LastSeen:   time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddEntityMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Entities = append(ctxt.Memories.Entities, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddEntityMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddEntityMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// AddEventMemory MCP 工具：添加事件记忆
func (mm *MemoryMCP) AddEventMemory(
	ctx context.Context,
	input *AddEventMemoryInput,
) (*AddEventMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddEventMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 创建记忆
	memory := &beadscontext.EventMemory{
		ID:          generateMemoryID("event", input.Type+":"+input.Title),
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		OccurredAt:  time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddEventMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Events = append(ctxt.Memories.Events, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddEventMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddEventMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// AddCaseMemory MCP 工具：添加案例记忆
func (mm *MemoryMCP) AddCaseMemory(
	ctx context.Context,
	input *AddCaseMemoryInput,
) (*AddCaseMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddCaseMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 创建记忆
	memory := &beadscontext.CaseMemory{
		ID:           generateMemoryID("case", input.Problem),
		Domain:       input.Domain,
		Problem:      input.Problem,
		Solution:     input.Solution,
		Lessons:      input.Lessons,
		Tags:         input.Tags,
		AppliedCount: 0,
		CreatedAt:    time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddCaseMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Cases = append(ctxt.Memories.Cases, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddCaseMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddCaseMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// AddPatternMemory MCP 工具：添加模式记忆
func (mm *MemoryMCP) AddPatternMemory(
	ctx context.Context,
	input *AddPatternMemoryInput,
) (*AddPatternMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &AddPatternMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 设置默认置信度
	if input.Confidence == 0 {
		input.Confidence = 0.7
	}

	// 创建记忆
	memory := &beadscontext.PatternMemory{
		ID:         generateMemoryID("pattern", input.Category+":"+input.Pattern),
		Category:   input.Category,
		Pattern:    input.Pattern,
		Frequency:  1,
		Confidence: input.Confidence,
		LastSeen:   time.Now(),
	}

	// 添加到上下文
	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &AddPatternMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		ctxt.Memories = &beadscontext.MemoryCollection{}
	}
	ctxt.Memories.Patterns = append(ctxt.Memories.Patterns, memory)

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &AddPatternMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &AddPatternMemoryOutput{
		Success: true,
		ID:      memory.ID,
	}, nil
}

// GetMemoryByID MCP 工具：获取单个记忆
func (mm *MemoryMCP) GetMemoryByID(
	ctx context.Context,
	input *GetMemoryByIDInput,
) (*GetMemoryByIDOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &GetMemoryByIDOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &GetMemoryByIDOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		return &GetMemoryByIDOutput{
			Success: false,
			Error:   "no memories found",
		}, nil
	}

	// 根据类型查找
	var memory interface{}
	switch input.Type {
	case "profile":
		for _, m := range ctxt.Memories.Profiles {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	case "preference":
		for _, m := range ctxt.Memories.Preferences {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	case "entity":
		for _, m := range ctxt.Memories.Entities {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	case "event":
		for _, m := range ctxt.Memories.Events {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	case "case":
		for _, m := range ctxt.Memories.Cases {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	case "pattern":
		for _, m := range ctxt.Memories.Patterns {
			if m.ID == input.MemoryID {
				memory = m
				break
			}
		}
	}

	if memory == nil {
		return &GetMemoryByIDOutput{
			Success: false,
			Error:   "memory not found",
		}, nil
	}

	return &GetMemoryByIDOutput{
		Success: true,
		Memory:  memory,
	}, nil
}

// DeleteMemory MCP 工具：删除记忆
func (mm *MemoryMCP) DeleteMemory(
	ctx context.Context,
	input *DeleteMemoryInput,
) (*DeleteMemoryOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &DeleteMemoryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	ctxt, err := store.GetContext(ctx, input.ContextID)
	if err != nil {
		return &DeleteMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("get context failed: %v", err),
		}, nil
	}

	if ctxt.Memories == nil {
		return &DeleteMemoryOutput{
			Success: false,
			Error:   "no memories found",
		}, nil
	}

	// 根据类型删除
	switch input.Type {
	case "profile":
		for i, m := range ctxt.Memories.Profiles {
			if m.ID == input.MemoryID {
				ctxt.Memories.Profiles = append(ctxt.Memories.Profiles[:i], ctxt.Memories.Profiles[i+1:]...)
				break
			}
		}
	case "preference":
		for i, m := range ctxt.Memories.Preferences {
			if m.ID == input.MemoryID {
				ctxt.Memories.Preferences = append(ctxt.Memories.Preferences[:i], ctxt.Memories.Preferences[i+1:]...)
				break
			}
		}
	case "entity":
		for i, m := range ctxt.Memories.Entities {
			if m.ID == input.MemoryID {
				ctxt.Memories.Entities = append(ctxt.Memories.Entities[:i], ctxt.Memories.Entities[i+1:]...)
				break
			}
		}
	case "event":
		for i, m := range ctxt.Memories.Events {
			if m.ID == input.MemoryID {
				ctxt.Memories.Events = append(ctxt.Memories.Events[:i], ctxt.Memories.Events[i+1:]...)
				break
			}
		}
	case "case":
		for i, m := range ctxt.Memories.Cases {
			if m.ID == input.MemoryID {
				ctxt.Memories.Cases = append(ctxt.Memories.Cases[:i], ctxt.Memories.Cases[i+1:]...)
				break
			}
		}
	case "pattern":
		for i, m := range ctxt.Memories.Patterns {
			if m.ID == input.MemoryID {
				ctxt.Memories.Patterns = append(ctxt.Memories.Patterns[:i], ctxt.Memories.Patterns[i+1:]...)
				break
			}
		}
	default:
		return &DeleteMemoryOutput{
			Success: false,
			Error:   "invalid memory type",
		}, nil
	}

	// 更新上下文
	if err := store.UpdateMemories(ctx, input.ContextID, ctxt.Memories); err != nil {
		return &DeleteMemoryOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	return &DeleteMemoryOutput{
		Success: true,
	}, nil
}

// GetMemoryStats MCP 工具：获取记忆统计
func (mm *MemoryMCP) GetMemoryStats(
	ctx context.Context,
	input *GetMemoryStatsInput,
) (*GetMemoryStatsOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &GetMemoryStatsOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 如果指定了上下文 ID，获取该上下文的记忆统计
	if input.ContextID != "" {
		ctxt, err := store.GetContext(ctx, input.ContextID)
		if err != nil {
			return &GetMemoryStatsOutput{
				Success: false,
				Error:   fmt.Sprintf("get context failed: %v", err),
			}, nil
		}

		stats := &beadscontext.MemoryStats{
			TotalMemories: int64(ctxt.Memories.GetMemoryCount()),
			ByType:        make(map[beadscontext.MemoryType]int64),
		}

		if len(ctxt.Memories.Profiles) > 0 {
			stats.ByType[beadscontext.MemoryTypeProfile] = int64(len(ctxt.Memories.Profiles))
		}
		if len(ctxt.Memories.Preferences) > 0 {
			stats.ByType[beadscontext.MemoryTypePreference] = int64(len(ctxt.Memories.Preferences))
		}
		if len(ctxt.Memories.Entities) > 0 {
			stats.ByType[beadscontext.MemoryTypeEntity] = int64(len(ctxt.Memories.Entities))
		}
		if len(ctxt.Memories.Events) > 0 {
			stats.ByType[beadscontext.MemoryTypeEvent] = int64(len(ctxt.Memories.Events))
		}
		if len(ctxt.Memories.Cases) > 0 {
			stats.ByType[beadscontext.MemoryTypeCase] = int64(len(ctxt.Memories.Cases))
		}
		if len(ctxt.Memories.Patterns) > 0 {
			stats.ByType[beadscontext.MemoryTypePattern] = int64(len(ctxt.Memories.Patterns))
		}

		return &GetMemoryStatsOutput{
			Success: true,
			Stats:   stats,
		}, nil
	}

	// 否则，尝试获取全局统计
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
				if err == nil {
					return &GetMemoryStatsOutput{
						Success: true,
						Stats:   &stats.MemoryStats,
					}, nil
				}
			}
		}
	}

	return &GetMemoryStatsOutput{
		Success: false,
		Error:   "stats not available",
	}, nil
}

// MergeMemories MCP 工具：合并两个上下文的记忆
func (mm *MemoryMCP) MergeMemories(
	ctx context.Context,
	input *MergeMemoriesInput,
) (*MergeMemoriesOutput, error) {
	store, err := mm.getContextStore(ctx)
	if err != nil {
		return &MergeMemoriesOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取源上下文
	sourceCtxt, err := store.GetContext(ctx, input.SourceContextID)
	if err != nil {
		return &MergeMemoriesOutput{
			Success: false,
			Error:   fmt.Sprintf("get source context failed: %v", err),
		}, nil
	}

	// 获取目标上下文
	targetCtxt, err := store.GetContext(ctx, input.TargetContextID)
	if err != nil {
		return &MergeMemoriesOutput{
			Success: false,
			Error:   fmt.Sprintf("get target context failed: %v", err),
		}, nil
	}

	if sourceCtxt.Memories == nil {
		sourceCtxt.Memories = &beadscontext.MemoryCollection{}
	}
	if targetCtxt.Memories == nil {
		targetCtxt.Memories = &beadscontext.MemoryCollection{}
	}

	// 记录原始数量
	originalCount := targetCtxt.Memories.GetMemoryCount()

	// 合并记忆
	targetCtxt.Memories.Profiles = append(targetCtxt.Memories.Profiles, sourceCtxt.Memories.Profiles...)
	targetCtxt.Memories.Preferences = append(targetCtxt.Memories.Preferences, sourceCtxt.Memories.Preferences...)
	targetCtxt.Memories.Entities = append(targetCtxt.Memories.Entities, sourceCtxt.Memories.Entities...)
	targetCtxt.Memories.Events = append(targetCtxt.Memories.Events, sourceCtxt.Memories.Events...)
	targetCtxt.Memories.Cases = append(targetCtxt.Memories.Cases, sourceCtxt.Memories.Cases...)
	targetCtxt.Memories.Patterns = append(targetCtxt.Memories.Patterns, sourceCtxt.Memories.Patterns...)

	// 去重
	deduplicated, err := store.DeduplicateMemories(ctx, input.TargetContextID)
	if err != nil {
		return &MergeMemoriesOutput{
			Success: false,
			Error:   fmt.Sprintf("deduplicate failed: %v", err),
		}, nil
	}

	// 更新目标上下文
	if err := store.UpdateMemories(ctx, input.TargetContextID, deduplicated); err != nil {
		return &MergeMemoriesOutput{
			Success: false,
			Error:   fmt.Sprintf("update memories failed: %v", err),
		}, nil
	}

	finalCount := deduplicated.GetMemoryCount()
	addedCount := sourceCtxt.Memories.GetMemoryCount()
	mergedCount := originalCount + addedCount - finalCount

	return &MergeMemoriesOutput{
		Success:     true,
		AddedCount:  addedCount,
		MergedCount: mergedCount,
	}, nil
}

// ===== 辅助方法 =====

// getContextStore 获取上下文存储
func (mm *MemoryMCP) getContextStore(_ context.Context) (beadscontext.ContextStore, error) {
	type memoryTracker interface {
		IsContextEnabled() bool
		GetContextStore() beadscontext.ContextStore
	}

	tracker, ok := mm.tracker.(memoryTracker)
	if !ok {
		return nil, fmt.Errorf("memory operations not supported")
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

// generateMemoryID 生成记忆 ID
func generateMemoryID(memoryType, content string) string {
	// 简化实现：使用类型和内容的哈希
	return fmt.Sprintf("%s-%x", memoryType, simpleHash(content))
}

// simpleHash 简单哈希函数
func simpleHash(s string) uint32 {
	h := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(s); i++ {
		h *= prime32
		h ^= uint32(s[i])
	}
	return h
}
