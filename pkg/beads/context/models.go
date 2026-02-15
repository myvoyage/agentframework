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

package context

import (
	"encoding/json"
	"time"
)

// CreateContextRequest 创建上下文请求
type CreateContextRequest struct {
	Type      ContextType       `json:"type"`                // 上下文类型
	Title     string            `json:"title"`               // 上下文标题
	Workspace string            `json:"workspace,omitempty"` // 工作区路径
	Content   []byte            `json:"content,omitempty"`   // 上下文内容
	Metadata  map[string]string `json:"metadata,omitempty"`  // 扩展元数据
	ParentRef string            `json:"parent_ref,omitempty"` // 父引用（如任务 ID）
}

// CreateContextResponse 创建上下文响应
type CreateContextResponse struct {
	ID        string    `json:"id"`        // 创建的上下文 ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// UpdateContextRequest 更新上下文请求
type UpdateContextRequest struct {
	Title     *string            `json:"title,omitempty"`
	Workspace *string            `json:"workspace,omitempty"`
	Content   *[]byte            `json:"content,omitempty"`
	Metadata  *map[string]string `json:"metadata,omitempty"`
}

// GetContextRequest 获取上下文请求
type GetContextRequest struct {
	ID          string `json:"id"`           // 上下文 ID
	IncludeContent bool `json:"include_content,omitempty"` // 是否包含内容
}

// DeleteContextRequest 删除上下文请求
type DeleteContextRequest struct {
	ID     string `json:"id"`     // 上下文 ID
	Force  bool   `json:"force,omitempty"` // 强制删除（即使有关联任务）
}

// ListContextsRequest 列出上下文请求
type ListContextsRequest struct {
	TaskID    string            `json:"task_id,omitempty"`    // 按任务 ID 过滤
	Type      *ContextType      `json:"type,omitempty"`       // 按类型过滤
	Workspace string            `json:"workspace,omitempty"`  // 按工作区过滤
	Metadata  map[string]string `json:"metadata,omitempty"`   // 按元数据过滤
	Limit     int               `json:"limit,omitempty"`      // 返回数量限制
	Offset    int               `json:"offset,omitempty"`     // 偏移量
}

// ListContextsResponse 列出上下文响应
type ListContextsResponse struct {
	Contexts   []*Context `json:"contexts"`    // 上下文列表
	TotalCount int        `json:"total_count"` // 总数
	Offset     int        `json:"offset"`      // 当前偏移量
	Limit      int        `json:"limit"`       // 当前限制
}

// AssociateContextRequest 关联上下文请求
type AssociateContextRequest struct {
	TaskID    string `json:"task_id"`    // 任务 ID
	ContextID string `json:"context_id"` // 上下文 ID
}

// DissociateContextRequest 解除关联请求
type DissociateContextRequest struct {
	TaskID    string `json:"task_id"`    // 任务 ID
	ContextID string `json:"context_id"` // 上下文 ID
}

// QueryTasksWithContextRequest 联合查询请求
type QueryTasksWithContextRequest struct {
	TaskQuery     interface{}    `json:"task_query,omitempty"`     // 任务查询条件（使用接口避免循环依赖）
	ContextFilter ContextFilter `json:"context_filter,omitempty"` // 上下文过滤条件
	Limit         int            `json:"limit,omitempty"`         // 返回数量限制
	Offset        int            `json:"offset,omitempty"`        // 偏移量
}

// QueryTasksWithContextResponse 联合查询响应
type QueryTasksWithContextResponse struct {
	Results    []*TaskWithContext `json:"results"`     // 查询结果
	TotalCount int               `json:"total_count"` // 总数
	Offset     int               `json:"offset"`      // 当前偏移量
	Limit      int               `json:"limit"`       // 当前限制
}

// ContextAssociation 上下文关联关系
type ContextAssociation struct {
	ContextID   string    `json:"context_id"`   // 上下文 ID
	TaskID      string    `json:"task_id"`      // 任务 ID
	AssociatedAt time.Time `json:"associated_at"` // 关联时间
	AssociationType string `json:"association_type,omitempty"` // 关联类型
}

// ContextSnapshot 上下文快照
type ContextSnapshot struct {
	Context      *Context           `json:"context"`       // 上下文内容
	Associations []*ContextAssociation `json:"associations,omitempty"` // 关联关系
	Version      int                `json:"version"`       // 版本号
	TakenAt      time.Time          `json:"taken_at"`      // 快照时间
}

// ContextDiff 上下文差异
type ContextDiff struct {
	Before      *Context    `json:"before,omitempty"` // 修改前
	After       *Context    `json:"after,omitempty"`  // 修改后
	ChangedFields []string  `json:"changed_fields"`   // 变更字段
	ChangedAt    time.Time  `json:"changed_at"`       // 变更时间
}

// ContextIndex 上下文索引（用于快速查询）
type ContextIndex struct {
	ContextID    string            `json:"context_id"`
	Type         ContextType       `json:"type"`
	Workspace    string            `json:"workspace,omitempty"`
	MetadataKeys []string          `json:"metadata_keys,omitempty"` // 元数据键
	TaskIDs      []string          `json:"task_ids,omitempty"`       // 关联任务 ID 列表
	ContentSize  int               `json:"content_size,omitempty"`  // 内容大小
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ContextValidationResult 上下文验证结果
type ContextValidationResult struct {
	Valid   bool     `json:"valid"`              // 是否有效
	Errors  []string `json:"errors,omitempty"`   // 错误列表
	Warnings []string `json:"warnings,omitempty"` // 警告列表
}

// Validate 验证上下文
func (c *Context) Validate() *ContextValidationResult {
	result := &ContextValidationResult{Valid: true}

	// 验证必填字段
	if c.ID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "context ID is required")
	}

	if c.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "context title is required")
	}

	if c.Type == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "context type is required")
	} else {
		// 验证类型是否有效
		validTypes := map[ContextType]bool{
			ContextTypeProject:  true,
			ContextTypeFile:     true,
			ContextTypeCodebase: true,
			ContextTypeCustom:   true,
			ContextTypeMemory:   true,
			ContextTypeResource: true,
			ContextTypeSkill:    true,
		}
		if !validTypes[c.Type] {
			result.Valid = false
			result.Errors = append(result.Errors, "invalid context type: "+string(c.Type))
		}
	}

	// 验证时间戳
	if c.CreatedAt.IsZero() {
		result.Warnings = append(result.Warnings, "created_at is not set")
	}

	if c.UpdatedAt.Before(c.CreatedAt) {
		result.Valid = false
		result.Errors = append(result.Errors, "updated_at cannot be before created_at")
	}

	return result
}

// Clone 创建上下文的深拷贝
func (c *Context) Clone() *Context {
	if c == nil {
		return nil
	}

	clone := &Context{
		ID:        c.ID,
		Type:      c.Type,
		Title:     c.Title,
		Workspace: c.Workspace,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Version:   c.Version,
		ParentID:  c.ParentID,
		TaskRefs:  append([]string{}, c.TaskRefs...),
	}

	if c.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range c.Metadata {
			clone.Metadata[k] = v
		}
	}

	// 深拷贝 Layers
	if c.Layers.L0 != nil {
		clone.Layers.L0 = &LayerSummary{
			Content:     c.Layers.L0.Content,
			Tokens:      c.Layers.L0.Tokens,
			GeneratedAt: c.Layers.L0.GeneratedAt,
			Method:      c.Layers.L0.Method,
		}
	}
	if c.Layers.L1 != nil {
		clone.Layers.L1 = &LayerOverview{
			Content:     c.Layers.L1.Content,
			Tokens:      c.Layers.L1.Tokens,
			Sections:    append([]string{}, c.Layers.L1.Sections...),
			KeyPoints:   append([]string{}, c.Layers.L1.KeyPoints...),
			GeneratedAt: c.Layers.L1.GeneratedAt,
			Method:      c.Layers.L1.Method,
		}
	}
	if c.Layers.L2 != nil {
		clone.Layers.L2 = &LayerDetails{
			Content:     c.Layers.L2.Content,
			Tokens:      c.Layers.L2.Tokens,
			Format:      c.Layers.L2.Format,
			Source:      c.Layers.L2.Source,
			GeneratedAt: c.Layers.L2.GeneratedAt,
		}
		if c.Layers.L2.Metadata != nil {
			clone.Layers.L2.Metadata = make(map[string]string)
			for k, v := range c.Layers.L2.Metadata {
				clone.Layers.L2.Metadata[k] = v
			}
		}
	}

	// 深拷贝 Memories
	if c.Memories != nil {
		clone.Memories = &MemoryCollection{}
		if c.Memories.Profiles != nil {
			clone.Memories.Profiles = append([]*ProfileMemory{}, c.Memories.Profiles...)
		}
		if c.Memories.Preferences != nil {
			clone.Memories.Preferences = append([]*PreferenceMemory{}, c.Memories.Preferences...)
		}
		if c.Memories.Entities != nil {
			clone.Memories.Entities = append([]*EntityMemory{}, c.Memories.Entities...)
		}
		if c.Memories.Events != nil {
			clone.Memories.Events = append([]*EventMemory{}, c.Memories.Events...)
		}
		if c.Memories.Cases != nil {
			clone.Memories.Cases = append([]*CaseMemory{}, c.Memories.Cases...)
		}
		if c.Memories.Patterns != nil {
			clone.Memories.Patterns = append([]*PatternMemory{}, c.Memories.Patterns...)
		}
	}

	return clone
}

// ToJSON 转换为 JSON
func (c *Context) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// FromJSON 从 JSON 创建上下文
func ContextFromJSON(data []byte) (*Context, error) {
	var c Context
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetMetadataValue 获取元数据值
func (c *Context) GetMetadataValue(key string) (string, bool) {
	if c.Metadata == nil {
		return "", false
	}
	val, ok := c.Metadata[key]
	return val, ok
}

// SetMetadataValue 设置元数据值
func (c *Context) SetMetadataValue(key, value string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
}

// DeleteMetadataValue 删除元数据值
func (c *Context) DeleteMetadataValue(key string) {
	if c.Metadata != nil {
		delete(c.Metadata, key)
	}
}

// HasTask 检查上下文是否关联到指定任务
func (c *Context) HasTask(taskID string) bool {
	if c.Metadata == nil {
		return false
	}
	_, ok := c.Metadata["task_"+taskID]
	return ok
}

// AddTaskRef 添加任务引用到元数据
func (c *Context) AddTaskRef(taskID string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata["task_"+taskID] = time.Now().Format(time.RFC3339)
}

// RemoveTaskRef 从元数据中移除任务引用
func (c *Context) RemoveTaskRef(taskID string) {
	if c.Metadata != nil {
		delete(c.Metadata, "task_"+taskID)
	}
}

// GetContentSize 获取内容大小
func (c *Context) GetContentSize() int {
	if c.Layers.L2 != nil {
		return len(c.Layers.L2.Content)
	}
	return 0
}

// IsEmpty 检查上下文是否为空（无内容）
func (c *Context) IsEmpty() bool {
	if c.Layers.L2 != nil {
		return len(c.Layers.L2.Content) == 0
	}
	return true
}

// Touch 更新 UpdatedAt 时间戳
func (c *Context) Touch() {
	c.UpdatedAt = time.Now()
}

// String 返回上下文的字符串表示
func (c *Context) String() string {
	return "Context{" +
		"ID=" + c.ID +
		", Type=" + string(c.Type) +
		", Title=" + c.Title +
		", Workspace=" + c.Workspace +
		"}"
}
