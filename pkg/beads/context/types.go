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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Context 三层上下文模型
// 采用 OpenViking 风格的设计，支持 L0 摘要、L1 概览、L2 详情三层结构
type Context struct {
	// 基本信息
	ID        string            `json:"id"`                  // SHA-256 基于内容生成
	Type      ContextType       `json:"type"`                // 上下文类型
	Title     string            `json:"title"`               // 标题
	Workspace string            `json:"workspace,omitempty"` // 工作区路径
	URI       string            `json:"uri,omitempty"`       // viking:// URI

	// 三层内容
	Layers ContextLayers `json:"layers"` // 三层内容

	// 元数据和记忆
	Metadata map[string]string   `json:"metadata,omitempty"` // 元数据
	Memories *MemoryCollection   `json:"memories,omitempty"` // 提取的记忆

	// 时间戳
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`  // 更新时间
	AccessedAt time.Time `json:"accessed_at"` // 访问时间

	// 版本控制
	Version  int    `json:"version"`            // 版本号
	ParentID string `json:"parent_id,omitempty"` // 父上下文ID

	// 任务引用
	TaskRefs []string `json:"task_refs,omitempty"` // 关联的任务ID列表
}

// ContextLayers 三层内容结构
type ContextLayers struct {
	L0 *LayerSummary  `json:"l0,omitempty"` // L0: 摘要层 (~100 tokens)
	L1 *LayerOverview `json:"l1,omitempty"` // L1: 概览层 (~2k tokens)
	L2 *LayerDetails  `json:"l2,omitempty"` // L2: 详情层 (完整内容)
}

// LayerSummary L0 摘要层
// 用于快速过滤和向量搜索，约 100 tokens
type LayerSummary struct {
	Content     string    `json:"content"`              // 摘要内容
	Tokens      int       `json:"tokens"`               // Token 数量
	GeneratedAt time.Time `json:"generated_at"`         // 生成时间
	Method      string    `json:"method"`               // 生成方法 (ai/template/manual)
}

// LayerOverview L1 概览层
// 用于 Rerank 精排和内容导航，约 2k tokens
type LayerOverview struct {
	Content     string   `json:"content"`       // 概览内容
	Tokens      int      `json:"tokens"`        // Token 数量
	Sections    []string `json:"sections"`      // 章节列表
	KeyPoints   []string `json:"key_points"`    // 关键点
	GeneratedAt time.Time `json:"generated_at"` // 生成时间
	Method      string   `json:"method"`        // 生成方法
}

// LayerDetails L2 详情层
// 完整内容，无 Token 限制，按需加载
type LayerDetails struct {
	Content     string                 `json:"content"`        // 完整内容
	Tokens      int                    `json:"tokens"`         // Token 数量
	Format      string                 `json:"format"`         // 内容格式 (markdown/json/plain/etc)
	Encoding    string                 `json:"encoding"`       // 编码方式
	Source      string                 `json:"source"`         // 来源 (file/memory/generated/api)
	GeneratedAt time.Time              `json:"generated_at"`   // 生成时间
	Metadata    map[string]string      `json:"metadata"`       // 额外元数据
}

// ContextType 上下文类型
type ContextType string

const (
	ContextTypeProject     ContextType = "project"     // 项目上下文
	ContextTypeFile        ContextType = "file"        // 文件上下文
	ContextTypeCodebase    ContextType = "codebase"    // 代码库上下文
	ContextTypeMemory      ContextType = "memory"      // 记忆上下文
	ContextTypeResource    ContextType = "resource"    // 资源上下文
	ContextTypeSkill       ContextType = "skill"       // 技能上下文
	ContextTypeConversation ContextType = "conversation" // 对话上下文
	ContextTypeSession     ContextType = "session"     // 会话上下文
	ContextTypeKnowledge   ContextType = "knowledge"   // 知识库上下文
	ContextTypeTask        ContextType = "task"        // 任务上下文
	ContextTypeCustom      ContextType = "custom"      // 自定义上下文
)

// LayerType 层级类型
type LayerType string

const (
	LayerAuto       LayerType = "auto" // 自动选择最合适的层
	LayerTypeL0     LayerType = "l0"   // 摘要层
	LayerTypeL1     LayerType = "l1"   // 概览层
	LayerTypeL2     LayerType = "l2"   // 详情层
)

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeProfile    MemoryType = "profile"    // 用户画像
	MemoryTypePreference MemoryType = "preference" // 用户偏好
	MemoryTypeEntity     MemoryType = "entity"     // 实体知识
	MemoryTypeEvent      MemoryType = "event"      // 事件记录
	MemoryTypeCase       MemoryType = "case"       // 案例库
	MemoryTypePattern    MemoryType = "pattern"    // 模式识别
)

// MemoryCollection 记忆集合
// 包含 6 种记忆类型的集合
type MemoryCollection struct {
	Profiles    []*ProfileMemory    `json:"profiles,omitempty"`    // 用户画像
	Preferences []*PreferenceMemory `json:"preferences,omitempty"` // 用户偏好
	Entities    []*EntityMemory     `json:"entities,omitempty"`    // 实体知识
	Events      []*EventMemory      `json:"events,omitempty"`      // 事件记录
	Cases       []*CaseMemory       `json:"cases,omitempty"`       // 案例库
	Patterns    []*PatternMemory    `json:"patterns,omitempty"`    // 模式识别
}

// ProfileMemory 用户画像记忆
// 存储用户的身份、角色、特征等信息
type ProfileMemory struct {
	ID          string            `json:"id"`    // 记忆 ID
	Name        string            `json:"name"`  // 用户名称
	Role        string            `json:"role"`  // 用户角色
	Traits      map[string]string `json:"traits"` // 用户特征
	Goals       []string          `json:"goals"`  // 用户目标
	Constraints []string          `json:"constraints"` // 约束条件
	UpdatedAt   time.Time         `json:"updated_at"` // 更新时间
}

// PreferenceMemory 用户偏好记忆
// 存储用户的偏好设置
type PreferenceMemory struct {
	ID        string    `json:"id"`         // 记忆 ID
	Category  string    `json:"category"`   // 偏好类别 (coding/writing/analysis/etc)
	Key       string    `json:"key"`        // 偏好键
	Value     string    `json:"value"`      // 偏好值
	Confidence float64  `json:"confidence"` // 置信度 (0-1)
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// EntityRelation 实体关系
type EntityRelation struct {
	Type         string    `json:"type"`          // 关系类型
	EntityID     string    `json:"entity_id"`     // 关联实体 ID
	Relation     string    `json:"relation"`      // 关系描述
	Confidence   float64   `json:"confidence"`    // 关系置信度
	FirstSeen    time.Time `json:"first_seen"`    // 首次发现时间
	LastSeen     time.Time `json:"last_seen"`     // 最近发现时间
}

// EntityMemory 实体知识记忆
// 存储人、组织、概念等实体信息
type EntityMemory struct {
	ID         string           `json:"id"`          // 记忆 ID
	Type       string           `json:"type"`        // 实体类型 (person/org/concept/etc)
	Name       string           `json:"name"`        // 实体名称
	Attributes map[string]string `json:"attributes"` // 实体属性
	Relations  []EntityRelation `json:"relations"`   // 实体关系
	FirstSeen  time.Time        `json:"first_seen"`  // 首次发现时间
	LastSeen   time.Time        `json:"last_seen"`   // 最近发现时间
}

// EventMemory 事件记录记忆
// 存储重要事件、决策、里程碑等
type EventMemory struct {
	ID          string    `json:"id"`          // 记忆 ID
	Type        string    `json:"type"`        // 事件类型 (decision/milestone/interaction/etc)
	Title       string    `json:"title"`       // 事件标题
	Description string    `json:"description"` // 事件描述
	OccurredAt  time.Time `json:"occurred_at"` // 发生时间
	Participants []string `json:"participants"` // 参与者
	Outcomes    []string  `json:"outcomes"`    // 事件结果
}

// CaseMemory 案例库记忆
// 存储问题和解决方案的案例
type CaseMemory struct {
	ID           string    `json:"id"`            // 记忆 ID
	Domain       string    `json:"domain"`        // 领域 (programming/debugging/design/etc)
	Problem      string    `json:"problem"`       // 问题描述
	Solution     string    `json:"solution"`      // 解决方案
	Lessons      []string  `json:"lessons"`       // 经验教训
	Tags         []string  `json:"tags"`          // 标签
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	AppliedCount int       `json:"applied_count"` // 应用次数
}

// PatternMemory 模式识别记忆
// 存储可复用的模式和流程
type PatternMemory struct {
	ID         string    `json:"id"`         // 记忆 ID
	Category   string    `json:"category"`   // 类别 (coding/workflow/communication)
	Pattern    string    `json:"pattern"`    // 模式描述
	Frequency  int       `json:"frequency"`  // 出现频率
	LastSeen   time.Time `json:"last_seen"`  // 最近发现时间
	Confidence float64   `json:"confidence"` // 识别置信度
}

// NewContext 创建新的上下文
func NewContext(ctxtType ContextType, title string) *Context {
	now := time.Now()
	return &Context{
		Type:       ctxtType,
		Title:      title,
		Layers:     ContextLayers{},
		Metadata:   make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
		AccessedAt: now,
		Version:    1,
		TaskRefs:   make([]string, 0),
	}
}

// GenerateID 生成上下文 ID
// 基于 SHA-256 哈希内容生成唯一 ID
func (c *Context) GenerateID() string {
	data := fmt.Sprintf("%s:%s:%s:%d", c.Type, c.Title, c.Workspace, c.CreatedAt.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// GetContent 获取指定层级的内容
func (c *Context) GetContent(layer LayerType) (string, error) {
	switch layer {
	case LayerTypeL0:
		if c.Layers.L0 != nil {
			return c.Layers.L0.Content, nil
		}
	case LayerTypeL1:
		if c.Layers.L1 != nil {
			return c.Layers.L1.Content, nil
		}
	case LayerTypeL2:
		if c.Layers.L2 != nil {
			return c.Layers.L2.Content, nil
		}
	case LayerAuto:
		// 自动选择最合适的层级
		if c.Layers.L1 != nil {
			return c.Layers.L1.Content, nil
		}
		if c.Layers.L0 != nil {
			return c.Layers.L0.Content, nil
		}
		if c.Layers.L2 != nil {
			return c.Layers.L2.Content, nil
		}
	}
	return "", fmt.Errorf("layer %s not found", layer)
}

// GetTotalTokens 获取所有层级的总 Token 数
func (c *Context) GetTotalTokens() int {
	total := 0
	if c.Layers.L0 != nil {
		total += c.Layers.L0.Tokens
	}
	if c.Layers.L1 != nil {
		total += c.Layers.L1.Tokens
	}
	if c.Layers.L2 != nil {
		total += c.Layers.L2.Tokens
	}
	return total
}

// HasLayer 检查是否有指定层级
func (c *Context) HasLayer(layer LayerType) bool {
	switch layer {
	case LayerTypeL0:
		return c.Layers.L0 != nil
	case LayerTypeL1:
		return c.Layers.L1 != nil
	case LayerTypeL2:
		return c.Layers.L2 != nil
	default:
		return false
	}
}

// GetAvailableLayers 获取可用的层级列表
func (c *Context) GetAvailableLayers() []LayerType {
	layers := make([]LayerType, 0)
	if c.Layers.L0 != nil {
		layers = append(layers, LayerTypeL0)
	}
	if c.Layers.L1 != nil {
		layers = append(layers, LayerTypeL1)
	}
	if c.Layers.L2 != nil {
		layers = append(layers, LayerTypeL2)
	}
	return layers
}

// UpdateAccessTime 更新访问时间
func (c *Context) UpdateAccessTime() {
	c.AccessedAt = time.Now()
}

// IncrementVersion 增加版本号
func (c *Context) IncrementVersion() {
	c.Version++
	c.UpdatedAt = time.Now()
}

// ===== MemoryCollection 方法 =====

// GetMemoryCount 获取记忆总数
func (mc *MemoryCollection) GetMemoryCount() int {
	if mc == nil {
		return 0
	}
	count := 0
	count += len(mc.Profiles)
	count += len(mc.Preferences)
	count += len(mc.Entities)
	count += len(mc.Events)
	count += len(mc.Cases)
	count += len(mc.Patterns)
	return count
}

// GetMemoriesByType 获取指定类型的记忆
func (mc *MemoryCollection) GetMemoriesByType(memoryType MemoryType) interface{} {
	if mc == nil {
		return nil
	}
	switch memoryType {
	case MemoryTypeProfile:
		return mc.Profiles
	case MemoryTypePreference:
		return mc.Preferences
	case MemoryTypeEntity:
		return mc.Entities
	case MemoryTypeEvent:
		return mc.Events
	case MemoryTypeCase:
		return mc.Cases
	case MemoryTypePattern:
		return mc.Patterns
	default:
		return nil
	}
}

// AddMemory 添加记忆
func (mc *MemoryCollection) AddMemory(memoryType MemoryType, memory interface{}) error {
	if mc == nil {
		return fmt.Errorf("memory collection is nil")
	}
	switch m := memory.(type) {
	case *ProfileMemory:
		mc.Profiles = append(mc.Profiles, m)
	case *PreferenceMemory:
		mc.Preferences = append(mc.Preferences, m)
	case *EntityMemory:
		mc.Entities = append(mc.Entities, m)
	case *EventMemory:
		mc.Events = append(mc.Events, m)
	case *CaseMemory:
		mc.Cases = append(mc.Cases, m)
	case *PatternMemory:
		mc.Patterns = append(mc.Patterns, m)
	default:
		return fmt.Errorf("unknown memory type: %T", memory)
	}
	return nil
}

// Clear 清空所有记忆
func (mc *MemoryCollection) Clear() {
	if mc == nil {
		return
	}
	mc.Profiles = nil
	mc.Preferences = nil
	mc.Entities = nil
	mc.Events = nil
	mc.Cases = nil
	mc.Patterns = nil
}

// IsEmpty 检查是否为空
func (mc *MemoryCollection) IsEmpty() bool {
	if mc == nil {
		return true
	}
	return mc.GetMemoryCount() == 0
}

// ===== 上下文更新类型 =====

// ContextUpdate 上下文更新结构
type ContextUpdate struct {
	Title     *string            `json:"title,omitempty"`
	Workspace *string            `json:"workspace,omitempty"`
	URI       *string            `json:"uri,omitempty"`
	Metadata  *map[string]string `json:"metadata,omitempty"`
	Layers    *ContextLayersUpdate `json:"layers,omitempty"`
	Memories  *MemoryCollection  `json:"memories,omitempty"`
}

// ContextLayersUpdate 层级更新结构
type ContextLayersUpdate struct {
	L0 *LayerSummaryUpdate  `json:"l0,omitempty"`
	L1 *LayerOverviewUpdate `json:"l1,omitempty"`
	L2 *LayerDetailsUpdate  `json:"l2,omitempty"`
}

// MemoryCollectionUpdate 记忆集合更新结构
type MemoryCollectionUpdate struct {
	Profiles    *[]*ProfileMemory    `json:"profiles,omitempty"`
	Preferences *[]*PreferenceMemory `json:"preferences,omitempty"`
	Entities    *[]*EntityMemory     `json:"entities,omitempty"`
	Events      *[]*EventMemory      `json:"events,omitempty"`
	Cases       *[]*CaseMemory       `json:"cases,omitempty"`
	Patterns    *[]*PatternMemory    `json:"patterns,omitempty"`
}

// LayerSummaryUpdate L0 更新结构
type LayerSummaryUpdate struct {
	Content *string    `json:"content,omitempty"`
	Tokens  *int       `json:"tokens,omitempty"`
	Method  *string    `json:"method,omitempty"`
}

// LayerOverviewUpdate L1 更新结构
type LayerOverviewUpdate struct {
	Content   *string   `json:"content,omitempty"`
	Tokens    *int      `json:"tokens,omitempty"`
	Sections  *[]string `json:"sections,omitempty"`
	KeyPoints *[]string `json:"key_points,omitempty"`
	Method    *string   `json:"method,omitempty"`
}

// LayerDetailsUpdate L2 更新结构
type LayerDetailsUpdate struct {
	Content *string                 `json:"content,omitempty"`
	Tokens  *int                    `json:"tokens,omitempty"`
	Format  *string                 `json:"format,omitempty"`
	Source  *string                 `json:"source,omitempty"`
	Metadata *map[string]string      `json:"metadata,omitempty"`
}

// ===== Memory 类型方法 =====

// GetID 获取 Profile 记忆 ID
func (m *ProfileMemory) GetID() string {
	return m.ID
}

// GetID 获取 Preference 记忆 ID
func (m *PreferenceMemory) GetID() string {
	return m.ID
}

// GetID 获取 Entity 记忆 ID
func (m *EntityMemory) GetID() string {
	return m.ID
}

// GetID 获取 Event 记忆 ID
func (m *EventMemory) GetID() string {
	return m.ID
}

// GetID 获取 Case 记忆 ID
func (m *CaseMemory) GetID() string {
	return m.ID
}

// GetID 获取 Pattern 记忆 ID
func (m *PatternMemory) GetID() string {
	return m.ID
}

// ===== 记忆分层类型 =====

// MemoryTier 记忆分层类型
type MemoryTier string

const (
	// MemoryTierSession 会话记忆：当前会话的临时记忆，会话结束后可能清理
	MemoryTierSession MemoryTier = "session"
	// MemoryTierDaily 每日记忆：短期记忆，保留1-7天
	MemoryTierDaily MemoryTier = "daily"
	// MemoryTierLongTerm 长期记忆：重要记忆，长期保留
	MemoryTierLongTerm MemoryTier = "longterm"
)

// TieredMemory 分层记忆元数据
type TieredMemory struct {
	Tier            MemoryTier  `json:"tier"`                // 记忆分层
	CreatedAt       time.Time   `json:"created_at"`          // 创建时间
	ExpiresAt       time.Time   `json:"expires_at,omitempty"` // 过期时间
	AccessCount     int         `json:"access_count"`         // 访问次数
	LastAccessed    time.Time   `json:"last_accessed"`        // 最后访问时间
	ImportanceScore float64     `json:"importance_score"`     // 重要性评分(0-1)
}

// MemoryWithMeta 带元数据的记忆包装
type MemoryWithMeta struct {
	Meta TieredMemory `json:"meta"`   // 分层元数据
	Data interface{}    `json:"data"`  // *ProfileMemory, *PreferenceMemory, etc.
}

// EnhancedMemoryCollection 增强的记忆集合
// 保持向后兼容，新增分层存储能力
type EnhancedMemoryCollection struct {
	// 原有的6种记忆类型（保持兼容）
	*MemoryCollection

	// 新增分层索引
	SessionIndices  []string `json:"session_indices,omitempty"`
	DailyIndices    []string `json:"daily_indices,omitempty"`
	LongTermIndices []string `json:"longterm_indices,omitempty"`

	// 分层元数据映射
	MemoryMetadata map[string]TieredMemory `json:"memory_metadata,omitempty"`
}

// GetTier 获取记忆的分层
func (emc *EnhancedMemoryCollection) GetTier(memoryID string) MemoryTier {
	if emc.MemoryMetadata == nil {
		return MemoryTierSession // 默认会话记忆
	}
	if meta, ok := emc.MemoryMetadata[memoryID]; ok {
		return meta.Tier
	}
	return MemoryTierSession
}

// SetTier 设置记忆的分层
func (emc *EnhancedMemoryCollection) SetTier(memoryID string, tier MemoryTier, expiresAt time.Time, importance float64) {
	if emc.MemoryMetadata == nil {
		emc.MemoryMetadata = make(map[string]TieredMemory)
	}
	emc.MemoryMetadata[memoryID] = TieredMemory{
		Tier:            tier,
		CreatedAt:       time.Now(),
		ExpiresAt:       expiresAt,
		AccessCount:     0,
		LastAccessed:    time.Now(),
		ImportanceScore: importance,
	}
}

// IsEmpty 检查增强记忆集合是否为空
func (emc *EnhancedMemoryCollection) IsEmpty() bool {
	if emc == nil || emc.MemoryCollection == nil {
		return true
	}
	return emc.MemoryCollection.GetMemoryCount() == 0
}

// GetMemoryCount 获取增强记忆集合的总记忆数
func (emc *EnhancedMemoryCollection) GetMemoryCount() int {
	if emc == nil || emc.MemoryCollection == nil {
		return 0
	}
	return emc.MemoryCollection.GetMemoryCount()
}

// GetSessionMemories 获取会话记忆列表
func (emc *EnhancedMemoryCollection) GetSessionMemories() *MemoryCollection {
	return emc.MemoryCollection
}

// GetDailyMemories 获取每日记忆列表
func (emc *EnhancedMemoryCollection) GetDailyMemories() *MemoryCollection {
	return emc.MemoryCollection
}

// GetLongTermMemories 获取长期记忆列表
func (emc *EnhancedMemoryCollection) GetLongTermMemories() *MemoryCollection {
	return emc.MemoryCollection
}
