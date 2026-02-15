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
	"context"
	"time"

	"AgentFramework/pkg/beads"
)

// ContextStore 定义上下文存储接口，提供统一的上下文管理能力
// 支持三层模型（L0摘要、L1概览、L2详情）、VFS、记忆提取等功能
type ContextStore interface {
	// ===== 基础 CRUD 操作 =====

	// CreateContext 创建新的上下文
	CreateContext(ctx context.Context, ctxt *Context) (string, error)

	// GetContext 根据 ID 获取上下文
	GetContext(ctx context.Context, contextID string) (*Context, error)

	// UpdateContext 更新上下文
	UpdateContext(ctx context.Context, contextID string, updates ContextUpdate) error

	// DeleteContext 删除上下文
	DeleteContext(ctx context.Context, contextID string) error

	// ===== 三层上下文操作 =====

	// GetLayer 获取指定层级的内容
	GetLayer(ctx context.Context, contextID string, layer LayerType) (interface{}, error)

	// GenerateLayers 为上下文生成缺失的层级
	GenerateLayers(ctx context.Context, contextID string) error

	// RegenerateLayer 重新生成指定层级
	RegenerateLayer(ctx context.Context, contextID string, layer LayerType) error

	// SetLayer 设置指定层级的内容
	SetLayer(ctx context.Context, contextID string, layer LayerType, content interface{}) error

	// ===== VFS 操作 =====

	// ReadFile 从 VFS 读取文件
	ReadFile(ctx context.Context, uri string) ([]byte, error)

	// WriteFile 向 VFS 写入文件
	WriteFile(ctx context.Context, uri string, data []byte) error

	// ListFiles 列出目录中的文件
	ListFiles(ctx context.Context, uri string) ([]*VFSFileInfo, error)

	// SearchFiles 搜索文件
	SearchFiles(ctx context.Context, query string, opts ...SearchOption) ([]*VFSSearchResult, error)

	// DeleteFile 从 VFS 删除文件
	DeleteFile(ctx context.Context, uri string) error

	// ===== 记忆操作 =====

	// ExtractMemories 从上下文中提取记忆
	ExtractMemories(ctx context.Context, contextID string) (*MemoryCollection, error)

	// GetMemories 获取上下文的记忆
	GetMemories(ctx context.Context, contextID string, memoryTypes []MemoryType) (*MemoryCollection, error)

	// UpdateMemories 更新上下文的记忆
	UpdateMemories(ctx context.Context, contextID string, memories *MemoryCollection) error

	// DeduplicateMemories 去重记忆
	DeduplicateMemories(ctx context.Context, contextID string) (*MemoryCollection, error)

	// ===== 任务关联操作 =====

	// AssociateContext 将上下文与任务关联
	AssociateContext(ctx context.Context, taskID, contextID string) error

	// DissociateContext 解除任务与上下文的关联
	DissociateContext(ctx context.Context, taskID, contextID string) error

	// GetTaskContexts 获取任务的所有上下文
	GetTaskContexts(ctx context.Context, taskID string) ([]*Context, error)

	// GetContextTasks 获取使用指定上下文的所有任务
	GetContextTasks(ctx context.Context, contextID string) ([]*beads.Task, error)

	// ===== 联合查询 =====

	// QueryTasksWithContext 联合查询任务和上下文
	QueryTasksWithContext(ctx context.Context, query beads.Query, filter ContextFilter) ([]*TaskWithContext, error)

	// QueryContextsByTasks 根据任务列表查询上下文
	QueryContextsByTasks(ctx context.Context, taskIDs []string) (map[string][]*Context, error)

	// ===== 批量操作 =====

	// BatchCreate 批量创建上下文
	BatchCreate(ctx context.Context, contexts []*Context) ([]string, error)

	// BatchGet 批量获取上下文
	BatchGet(ctx context.Context, contextIDs []string) (map[string]*Context, error)

	// BatchUpdate 批量更新上下文
	BatchUpdate(ctx context.Context, updates map[string]ContextUpdate) error

	// BatchDelete 批量删除上下文
	BatchDelete(ctx context.Context, contextIDs []string) error

	// ===== 统计与健康 =====

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*ContextStoreStats, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// ===== 生命周期 =====

	// Start 启动上下文存储服务
	Start(ctx context.Context) error

	// Stop 停止上下文存储服务
	Stop(ctx context.Context) error

	// Sync 同步存储
	Sync(ctx context.Context) error
}

// ContextFilter 上下文查询过滤器
type ContextFilter struct {
	ContextID *string          // 按 ID 过滤
	Type      *ContextType     // 按类型过滤
	Workspace *string          // 按工作区过滤
	Metadata  map[string]string // 按元数据过滤
	CreatedAfter  *time.Time  // 创建时间起始
	CreatedBefore *time.Time  // 创建时间结束
}

// TaskWithContext 任务和上下文的组合结构
// 用于联合查询结果
type TaskWithContext struct {
	Task     *beads.Task  // 任务信息
	Contexts []*Context   // 关联的上下文列表
}

// ContextStoreConfig 上下文存储配置
type ContextStoreConfig struct {
	Type     string                 // 存储类型：openviking, memory, none
	Enabled  bool                   // 是否启用
	Config   map[string]interface{} // 类型特定配置
}

// ContextEvent 上下文相关事件
type ContextEvent struct {
	Type      ContextEventType `json:"type"`       // 事件类型
	ContextID string           `json:"context_id"` // 上下文 ID
	TaskID    string           `json:"task_id"`    // 关联任务 ID
	Timestamp time.Time        `json:"timestamp"`  // 时间戳
	Data      map[string]interface{} `json:"data"` // 事件数据
}

// ContextEventType 上下文事件类型
type ContextEventType string

const (
	// ContextEventCreated 上下文创建事件
	ContextEventCreated ContextEventType = "context.created"
	// ContextEventUpdated 上下文更新事件
	ContextEventUpdated ContextEventType = "context.updated"
	// ContextEventDeleted 上下文删除事件
	ContextEventDeleted ContextEventType = "context.deleted"
	// ContextEventAssociated 上下文关联事件
	ContextEventAssociated ContextEventType = "context.associated"
	// ContextEventDissociated 上下文解除关联事件
	ContextEventDissociated ContextEventType = "context.dissociated"
)

// ContextStoreStats 上下文存储统计信息
type ContextStoreStats struct {
	TotalContexts  int64         // 总上下文数
	TotalTasks     int64         // 关联任务数
	ByType         map[ContextType]int64 // 按类型统计
	StorageSize    int64         // 存储大小（字节）
	CacheHitRate   float64       // 缓存命中率
	AvgAccessTime  time.Duration // 平均访问时间
}

// Stater 统计接口
type Stater interface {
	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*ContextStoreStats, error)
}

// ===== VFS (虚拟文件系统) 接口 =====

// VFS 虚拟文件系统接口
// 提供类似文件系统的操作，使用 viking:// URI scheme
type VFS interface {
	// URI 操作

	// ParseURI 解析 URI
	ParseURI(uri string) (*VFSPath, error)

	// BuildURI 构建 URI
	BuildURI(scheme, path string, opts ...URIOption) (string, error)

	// 文件操作

	// ReadFile 读取文件内容
	ReadFile(ctx context.Context, uri string, layer LayerType) ([]byte, error)

	// WriteFile 写入文件内容
	WriteFile(ctx context.Context, uri string, data []byte) error

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, uri string) error

	// ListFiles 列出目录中的文件
	ListFiles(ctx context.Context, uri string) ([]*VFSFileInfo, error)

	// 目录操作

	// Mkdir 创建目录
	Mkdir(ctx context.Context, uri string) error

	// MkdirAll 递归创建目录
	MkdirAll(ctx context.Context, uri string) error

	// 查询操作

	// Glob 模式匹配文件
	Glob(ctx context.Context, pattern string) ([]string, error)

	// Search 搜索文件
	Search(ctx context.Context, query string, opts ...SearchOption) ([]*VFSSearchResult, error)

	// 移动和重命名

	// Move 移动文件或目录
	Move(ctx context.Context, oldPath, newPath string) error

	// Rename 重命名文件或目录
	Rename(ctx context.Context, oldPath, newPath string) error
}

// VFSPath 虚拟文件系统路径
type VFSPath struct {
	Scheme    string            `json:"scheme"`     // viking
	Workspace string            `json:"workspace"`  // 工作区
	Path      string            `json:"path"`       // 路径
	Layer     LayerType         `json:"layer"`      // 层级
	Query     map[string]string `json:"query"`      // 查询参数
}

// VFSFileInfo 文件信息
type VFSFileInfo struct {
	URI     string      `json:"uri"`               // 完整 URI
	Name    string      `json:"name"`              // 文件名
	Size    int64       `json:"size"`              // 文件大小
	Type    string      `json:"type"`              // file/dir/symlink
	Mode    interface{} `json:"mode"`              // 文件权限
	ModTime time.Time   `json:"mod_time"`          // 修改时间
	Layers  LayerAvailability `json:"layers"`       // 可用层级
}

// LayerAvailability 层级可用性
type LayerAvailability struct {
	L0 bool `json:"l0"` // 是否有 L0 层
	L1 bool `json:"l1"` // 是否有 L1 层
	L2 bool `json:"l2"` // 是否有 L2 层
}

// VFSSearchResult 搜索结果
type VFSSearchResult struct {
	URI     string    `json:"uri"`               // 匹配的 URI
	Score   float64   `json:"score"`             // 相关性得分
	Snippet string    `json:"snippet"`           // 匹配片段
	Layer   LayerType `json:"layer"`             // 匹配的层级
}

// URIOption URI 构建选项
type URIOption func(*VFSPath)

// SearchOption 搜索选项
type SearchOption func(*SearchOptions)

// SearchOptions 搜索选项配置
type SearchOptions struct {
	Layer       LayerType   // 搜索的层级
	MaxResults  int         // 最大结果数
	MinScore    float64     // 最小相关性得分
	ContextType ContextType // 上下文类型过滤
	Workspace   string      // 工作区过滤
}

// ===== 记忆提取接口 =====

// MemoryExtractor 记忆提取器接口
// 从文本、对话、上下文中提取和去重记忆
type MemoryExtractor interface {
	// ExtractFromContext 从上下文中提取记忆
	ExtractFromContext(ctx context.Context, ctxt *Context) (*MemoryCollection, error)

	// ExtractFromConversation 从对话中提取记忆
	ExtractFromConversation(ctx context.Context, messages []Message) (*MemoryCollection, error)

	// ExtractFromText 从文本中提取记忆
	ExtractFromText(ctx context.Context, text string, contentType string) (*MemoryCollection, error)

	// Deduplicate 去重记忆
	Deduplicate(ctx context.Context, memories *MemoryCollection) (*MemoryCollection, error)

	// Merge 合并记忆
	Merge(ctx context.Context, existing *MemoryCollection, new *MemoryCollection) (*MemoryCollection, error)
}

// Message 对话消息
type Message struct {
	Role      string            `json:"role"`        // user/assistant/system
	Content   string            `json:"content"`     // 消息内容
	Timestamp int64             `json:"timestamp"`   // 时间戳
	Metadata  map[string]string `json:"metadata"`    // 元数据
}

// MemoryDeduplicator 记忆去重器接口
type MemoryDeduplicator interface {
	// Deduplicate 去重记忆
	Deduplicate(ctx context.Context, memories *MemoryCollection) (*MemoryCollection, error)

	// FindSimilar 查找相似记忆
	FindSimilar(ctx context.Context, memory interface{}) ([]interface{}, error)

	// ShouldMerge 判断是否应该合并记忆
	ShouldMerge(ctx context.Context, existing, new interface{}) (bool, error)
}

// MemoryDeduplication 记忆去重决策类型
type MemoryDeduplication string

const (
	// DedupDecisionCreate 新记忆，直接创建
	DedupDecisionCreate MemoryDeduplication = "create"
	// DedupDecisionUpdate 更新现有记忆
	DedupDecisionUpdate MemoryDeduplication = "update"
	// DedupDecisionMerge 合并多条记忆
	DedupDecisionMerge MemoryDeduplication = "merge"
	// DedupDecisionSkip 完全重复，跳过
	DedupDecisionSkip MemoryDeduplication = "skip"
)

// ===== 层级生成器接口 =====

// LayerGenerator 层级生成器接口
// 为上下文生成三层内容（L0摘要、L1概览、L2详情）
type LayerGenerator interface {
	// GenerateL0 生成 L0 摘要层
	GenerateL0(ctx context.Context, content string) (*LayerSummary, error)

	// GenerateL1 生成 L1 概览层
	GenerateL1(ctx context.Context, content string) (*LayerOverview, error)

	// GenerateL2 生成 L2 详情层
	GenerateL2(ctx context.Context, content string, format string) (*LayerDetails, error)

	// GenerateAll 生成所有层级
	GenerateAll(ctx context.Context, content string, format string) (*ContextLayers, error)
}

// TokenCounter Token 计数器接口
type TokenCounter interface {
	// CountTokens 计算 Token 数量
	CountTokens(text string) int

	// CountTokensApprox 近似计算 Token 数量
	CountTokensApprox(text string) int
}

// ===== 上下文协调器接口 =====

// Coordinator 上下文协调器接口
// 管理三层上下文的同步和协调
type Coordinator interface {
	// 启动/停止

	// Start 启动协调器
	Start(ctx context.Context) error

	// Stop 停止协调器
	Stop(ctx context.Context) error

	// 同步操作

	// TriggerSync 手动触发同步
	TriggerSync(ctx context.Context) error

	// SyncTasksToContexts 同步任务到上下文
	SyncTasksToContexts(ctx context.Context) error

	// SyncContextsToTasks 同步上下文到任务
	SyncContextsToTasks(ctx context.Context) error

	// 层级生成

	// GenerateMissingLayers 生成缺失的层级
	GenerateMissingLayers(ctx context.Context, contextID string) error

	// RegenerateAllLayers 重新生成所有层级
	RegenerateAllLayers(ctx context.Context, contextID string) error

	// 记忆管理

	// ExtractAndMergeMemories 提取并合并记忆
	ExtractAndMergeMemories(ctx context.Context, contextID string) error

	// DeduplicateAllMemories 去重所有记忆
	DeduplicateAllMemories(ctx context.Context) error

	// VFS 管理

	// RegisterVFS 注册 VFS
	RegisterVFS(vfs VFS) error

	// UnregisterVFS 注销 VFS
	UnregisterVFS(scheme string) error

	// GetVFS 获取 VFS
	GetVFS(scheme string) (VFS, error)

	// 统计信息

	// GetStats 获取协调器统计信息
	GetStats(ctx context.Context) (*CoordinatorStats, error)
}

// CoordinatorStats 协调器统计信息
type CoordinatorStats struct {
	SyncCount      int64                `json:"sync_count"`       // 同步次数
	LastSyncTime   time.Time            `json:"last_sync_time"`   // 最后同步时间
	LayerGenStats  map[LayerType]int64  `json:"layer_gen_stats"`  // 层级生成统计
	MemoryStats    MemoryStats          `json:"memory_stats"`     // 记忆统计
	VFSStats       map[string]int64     `json:"vfs_stats"`        // VFS 统计
}

// MemoryStats 记忆统计
type MemoryStats struct {
	TotalMemories  int64                       `json:"total_memories"`  // 总记忆数
	ByType         map[MemoryType]int64        `json:"by_type"`         // 按类型统计
	DeduplicationRate float64                  `json:"deduplication_rate"` // 去重率
}

// ===== 上下文事件接口 =====

// EventPublisher 事件发布器接口
type EventPublisher interface {
	// Publish 发布事件
	Publish(ctx context.Context, event *ContextEvent) error

	// Subscribe 订阅事件
	Subscribe(ctx context.Context, eventType ContextEventType, handler EventHandler) error

	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, eventType ContextEventType) error
}

// EventHandler 事件处理器函数类型
type EventHandler func(ctx context.Context, event *ContextEvent) error

// EventStore 事件存储接口
type EventStore interface {
	// Append 添加事件
	Append(ctx context.Context, event *ContextEvent) error

	// Replay 回放事件
	Replay(ctx context.Context, contextID string, fromTime time.Time) ([]*ContextEvent, error)

	// GetLatest 获取最新事件
	GetLatest(ctx context.Context, contextID string, limit int) ([]*ContextEvent, error)

	// Compact 压缩事件
	Compact(ctx context.Context, beforeTime time.Time) error
}

// ===== 搜索选项辅助函数 =====

// WithSearchLayer 设置搜索层级
func WithSearchLayer(layer LayerType) SearchOption {
	return func(opts *SearchOptions) {
		opts.Layer = layer
	}
}

// WithMaxResults 设置最大结果数
func WithMaxResults(maxResults int) SearchOption {
	return func(opts *SearchOptions) {
		opts.MaxResults = maxResults
	}
}

// WithMinScore 设置最小相关性分数
func WithMinScore(minScore float64) SearchOption {
	return func(opts *SearchOptions) {
		opts.MinScore = minScore
	}
}

// WithContextType 设置上下文类型过滤
func WithContextType(contextType ContextType) SearchOption {
	return func(opts *SearchOptions) {
		opts.ContextType = contextType
	}
}

// WithWorkspace 设置工作区过滤
func WithWorkspace(workspace string) SearchOption {
	return func(opts *SearchOptions) {
		opts.Workspace = workspace
	}
}

// ===== 记忆压缩器接口 =====

// MemoryCompressor 记忆压缩器接口
// 提供记忆压缩、精华提取和重要性评分功能
type MemoryCompressor interface {
	// CompressMemories 压缩记忆集合到指定层级
	// 将大量记忆压缩为少量精华记忆，用于从短期转移到长期存储
	CompressMemories(ctx context.Context, memories *MemoryCollection, tier MemoryTier) (*MemoryCollection, error)

	// SummarizeByType 按类型压缩记忆
	// 对特定类型的记忆进行摘要压缩
	SummarizeByType(ctx context.Context, memories *MemoryCollection, memoryType MemoryType) (interface{}, error)

	// ExtractEssentials 提取最重要的记忆
	// 根据重要性评分提取指定数量的精华记忆
	ExtractEssentials(ctx context.Context, memories *MemoryCollection, maxCount int) (*MemoryCollection, error)

	// CalculateImportance 计算记忆重要性
	// 返回 0-1 之间的重要性分数
	CalculateImportance(ctx context.Context, memory interface{}) (float64, error)

	// MergeMemories 合并记忆
	// 将增量记忆合并到基础记忆中，去重并更新
	MergeMemories(ctx context.Context, base, delta *MemoryCollection) (*MemoryCollection, error)
}

// MemoryCompressionConfig 记忆压缩配置
type MemoryCompressionConfig struct {
	// 保留时长配置
	SessionRetentionDuration  time.Duration `json:"session_retention_duration"`  // 会话记忆保留时长
	DailyRetentionDuration    time.Duration `json:"daily_retention_duration"`    // 每日记忆保留时长
	LongTermRetentionDuration time.Duration `json:"longterm_retention_duration"` // 长期记忆保留时长

	// 数量限制配置
	MaxSessionMemories int `json:"max_session_memories"` // 会话记忆最大数量
	MaxDailyMemories   int `json:"max_daily_memories"`   // 每日记忆最大数量
	MaxLongTermMemories int `json:"max_longterm_memories"` // 长期记忆最大数量

	// 压缩触发配置
	CompressionThreshold float64 `json:"compression_threshold"` // 压缩阈值 (0-1)，当记忆重要性低于此值时可能被压缩
	CompressionInterval  time.Duration `json:"compression_interval"`  // 定期压缩间隔

	// LLM 配置
	ModelName   string  `json:"model_name"`    // 用于压缩的 LLM 模型名称
	MaxTokens   int     `json:"max_tokens"`    // 最大 Token 数
	Temperature float64 `json:"temperature"`   // 温度参数

	// 压缩策略配置
	PreserveSystemMemories bool `json:"preserve_system_memories"` // 是否保留系统记忆
	EnableAsyncCompression bool `json:"enable_async_compression"` // 是否启用异步压缩
}

// DefaultMemoryCompressionConfig 返回默认的压缩配置
func DefaultMemoryCompressionConfig() *MemoryCompressionConfig {
	return &MemoryCompressionConfig{
		SessionRetentionDuration:  24 * time.Hour,      // 会话记忆保留 24 小时
		DailyRetentionDuration:    7 * 24 * time.Hour,  // 每日记忆保留 7 天
		LongTermRetentionDuration: 365 * 24 * time.Hour, // 长期记忆保留 1 年

		MaxSessionMemories:  100,  // 会话记忆最多 100 条
		MaxDailyMemories:    500,  // 每日记忆最多 500 条
		MaxLongTermMemories: 1000, // 长期记忆最多 1000 条

		CompressionThreshold: 0.3,            // 重要性低于 0.3 的记忆可能被压缩
		CompressionInterval:  1 * time.Hour,  // 每小时检查一次是否需要压缩

		ModelName:   "gpt-4",  // 默认使用 GPT-4
		MaxTokens:   2000,     // 摘要最多 2000 tokens
		Temperature: 0.3,      // 较低温度以保持一致性

		PreserveSystemMemories: true, // 保留系统记忆
		EnableAsyncCompression: true, // 启用异步压缩
	}
}
