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
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/beads"
)

// ===== 本地接口定义（避免循环导入） =====

// layerGenerator 层级生成器接口
type layerGenerator interface {
	GenerateL0(ctx context.Context, content string) (*LayerSummary, error)
	GenerateL1(ctx context.Context, content string) (*LayerOverview, error)
	GenerateAll(ctx context.Context, content string, format string) (*ContextLayers, error)
}

// memoryExtractor 记忆提取器接口
type memoryExtractor interface {
	ExtractFromContext(ctx context.Context, ctxt *Context) (*MemoryCollection, error)
	Merge(ctx context.Context, existing, newMemories *MemoryCollection) (*MemoryCollection, error)
	Deduplicate(ctx context.Context, memories *MemoryCollection) (*MemoryCollection, error)
}

// memoryCompressor 记忆压缩器接口
type memoryCompressor interface {
	CompressMemories(ctx context.Context, memories *MemoryCollection, tier MemoryTier) (*MemoryCollection, error)
	MergeMemories(ctx context.Context, base, delta *MemoryCollection) (*MemoryCollection, error)
}

// CoordinatorImpl 上下文协调器实现
// 管理三层上下文的同步和协调
type CoordinatorImpl struct {
	taskTracker  beads.TaskTracker
	contextStore ContextStore
	layerGen     interface{} // *layers.LayerGenerator (avoid circular import)
	memoryExt    interface{} // *memory.Extractor (avoid circular import)
	memoryCompressor interface{} // *memory.LLMCompressor (avoid circular import)

	// 同步配置
	syncInterval   time.Duration
	autoSync       bool

	// 记忆压缩配置
	memoryConfig   *MemoryCompressionConfig
	compressionEnabled bool

	// 状态管理
	mu         sync.RWMutex
	started    bool
	ctx        context.Context
	cancel     context.CancelFunc
	syncCount  int64
	lastSync   time.Time

	// 统计信息
	layerGenStats map[LayerType]int64
	memoryStats   MemoryStats

	// VFS 注册表
	vfsRegistry map[string]VFS
}

// NewCoordinatorImpl 创建新的协调器实现
func NewCoordinatorImpl(taskTracker beads.TaskTracker, contextStore ContextStore) *CoordinatorImpl {
	return &CoordinatorImpl{
		taskTracker:         taskTracker,
		contextStore:        contextStore,
		layerGen:            nil, // 延迟初始化
		memoryExt:           nil, // 延迟初始化
		memoryCompressor:    nil, // 延迟初始化
		syncInterval:        5 * time.Minute,
		autoSync:            true,
		memoryConfig:        DefaultMemoryCompressionConfig(),
		compressionEnabled:  false, // 默认不启用，需要显式设置压缩器
		layerGenStats:       make(map[LayerType]int64),
		memoryStats: MemoryStats{
			ByType: make(map[MemoryType]int64),
		},
		vfsRegistry: make(map[string]VFS),
	}
}

// Start 启动协调器
func (c *CoordinatorImpl) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("coordinator already started")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.started = true

	// 如果启用自动同步，启动同步循环
	if c.autoSync {
		go c.syncLoop()
	}

	return nil
}

// Stop 停止协调器
func (c *CoordinatorImpl) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.started = false
	return nil
}

// ===== 同步操作 =====

// TriggerSync 手动触发同步
func (c *CoordinatorImpl) TriggerSync(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.performSync(ctx)
}

// SyncTasksToContexts 同步任务到上下文
func (c *CoordinatorImpl) SyncTasksToContexts(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取所有任务
	tasks, err := c.getAllTasks(ctx)
	if err != nil {
		return fmt.Errorf("get tasks: %w", err)
	}

	// 为每个任务同步上下文
	for _, task := range tasks {
		if err := c.syncTaskToContext(ctx, task); err != nil {
			fmt.Printf("Warning: failed to sync task %s: %v\n", task.ID, err)
		}
	}

	return nil
}

// SyncContextsToTasks 同步上下文到任务
func (c *CoordinatorImpl) SyncContextsToTasks(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取所有上下文
	contexts, err := c.getAllContexts(ctx)
	if err != nil {
		return fmt.Errorf("get contexts: %w", err)
	}

	// 为每个上下文同步任务引用
	for _, ctxt := range contexts {
		if err := c.syncContextToTask(ctx, ctxt); err != nil {
			fmt.Printf("Warning: failed to sync context %s: %v\n", ctxt.ID, err)
		}
	}

	return nil
}

// ===== 层级生成 =====

// GenerateMissingLayers 生成缺失的层级
func (c *CoordinatorImpl) GenerateMissingLayers(ctx context.Context, contextID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取上下文
	ctxt, err := c.contextStore.GetContext(ctx, contextID)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	// 检查缺失的层级
	missingLayers := c.findMissingLayers(ctxt)

	// 如果没有缺失的层级，直接返回
	if len(missingLayers) == 0 {
		return nil
	}

	// 从 L2 生成其他层级
	if ctxt.Layers.L2 != nil {
		content := ctxt.Layers.L2.Content

		// 生成缺失的层级
		for _, layer := range missingLayers {
			var err error

			switch layer {
			case LayerTypeL0:
				layerContent, err := c.generateL0(ctx, content)
				if err == nil {
					ctxt.Layers.L0 = layerContent
					c.layerGenStats[LayerTypeL0]++
				}
			case LayerTypeL1:
				layerContent, err := c.generateL1(ctx, content)
				if err == nil {
					ctxt.Layers.L1 = layerContent
					c.layerGenStats[LayerTypeL1]++
				}
			}

			if err != nil {
				return fmt.Errorf("generate layer %s: %w", layer, err)
			}
		}

		// 更新上下文
		update := ContextUpdate{}
		if ctxt.Layers.L0 != nil {
			update.Layers = &ContextLayersUpdate{}
			update.Layers.L0 = &LayerSummaryUpdate{
				Content: func(s string) *string { return &s }(ctxt.Layers.L0.Content),
				Tokens:  func(i int) *int { return &i }(ctxt.Layers.L0.Tokens),
				Method:  func(s string) *string { return &s }(ctxt.Layers.L0.Method),
			}
		}
		if ctxt.Layers.L1 != nil {
			if update.Layers == nil {
				update.Layers = &ContextLayersUpdate{}
			}
			update.Layers.L1 = &LayerOverviewUpdate{
				Content: func(s string) *string { return &s }(ctxt.Layers.L1.Content),
				Tokens:  func(i int) *int { return &i }(ctxt.Layers.L1.Tokens),
				Method:  func(s string) *string { return &s }(ctxt.Layers.L1.Method),
			}
		}

		return c.contextStore.UpdateContext(ctx, contextID, update)
	}

	return fmt.Errorf("no L2 content to generate from")
}

// RegenerateAllLayers 重新生成所有层级
func (c *CoordinatorImpl) RegenerateAllLayers(ctx context.Context, contextID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取上下文
	ctxt, err := c.contextStore.GetContext(ctx, contextID)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	// 从 L2 内容重新生成所有层级
	if ctxt.Layers.L2 != nil {
		content := ctxt.Layers.L2.Content
		format := ctxt.Layers.L2.Format

		// 生成所有层级
		allLayers, err := c.generateAllLayers(ctx, content, format)
		if err != nil {
			return fmt.Errorf("generate all layers: %w", err)
		}

		// 更新统计
		if allLayers.L0 != nil {
			c.layerGenStats[LayerTypeL0]++
		}
		if allLayers.L1 != nil {
			c.layerGenStats[LayerTypeL1]++
		}

		// 更新上下文
		return c.contextStore.UpdateContext(ctx, contextID, ContextUpdate{
			Layers: &ContextLayersUpdate{
				L0: &LayerSummaryUpdate{
					Content: func(s string) *string { return &s }(allLayers.L0.Content),
					Tokens:  func(i int) *int { return &i }(allLayers.L0.Tokens),
					Method:  func(s string) *string { return &s }(allLayers.L0.Method),
				},
				L1: &LayerOverviewUpdate{
					Content: func(s string) *string { return &s }(allLayers.L1.Content),
					Tokens:  func(i int) *int { return &i }(allLayers.L1.Tokens),
					Method:  func(s string) *string { return &s }(allLayers.L1.Method),
				},
			},
		})
	}

	return fmt.Errorf("no L2 content to regenerate from")
}

// ===== 记忆管理 =====

// ExtractAndMergeMemories 提取并合并记忆
func (c *CoordinatorImpl) ExtractAndMergeMemories(ctx context.Context, contextID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取上下文
	ctxt, err := c.contextStore.GetContext(ctx, contextID)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	// 提取记忆
	newMemories, err := c.extractMemories(ctx, ctxt)
	if err != nil {
		return fmt.Errorf("extract memories: %w", err)
	}

	// 合并现有记忆
	var mergedMemories *MemoryCollection
	if ctxt.Memories != nil {
		mergedMemories, _ = c.mergeMemories(ctx, ctxt.Memories, newMemories)
	} else {
		mergedMemories = newMemories
	}

	// 去重
	deduplicatedMemories, err := c.deduplicateMemories(ctx, mergedMemories)
	if err != nil {
		return fmt.Errorf("deduplicate memories: %w", err)
	}

	// 更新统计
	c.updateMemoryStats(deduplicatedMemories)

	// 更新上下文
	return c.contextStore.UpdateMemories(ctx, contextID, deduplicatedMemories)
}

// DeduplicateAllMemories 去重所有记忆
func (c *CoordinatorImpl) DeduplicateAllMemories(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取所有上下文
	contexts, err := c.getAllContexts(ctx)
	if err != nil {
		return fmt.Errorf("get contexts: %w", err)
	}

	// 去重每个上下文的记忆
	for _, ctxt := range contexts {
		if ctxt.Memories != nil {
			deduplicated, err := c.deduplicateMemories(ctx, ctxt.Memories)
			if err != nil {
				fmt.Printf("Warning: failed to deduplicate memories for context %s: %v\n", ctxt.ID, err)
				continue
			}

			// 更新上下文
			if err := c.contextStore.UpdateMemories(ctx, ctxt.ID, deduplicated); err != nil {
				fmt.Printf("Warning: failed to update memories for context %s: %v\n", ctxt.ID, err)
			}
		}
	}

	return nil
}

// ===== VFS 管理 =====

// RegisterVFS 注册 VFS
func (c *CoordinatorImpl) RegisterVFS(vfsObj VFS) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取 scheme（简化处理）
	scheme := "viking"

	c.vfsRegistry[scheme] = vfsObj

	return nil
}

// UnregisterVFS 注销 VFS
func (c *CoordinatorImpl) UnregisterVFS(scheme string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.vfsRegistry, scheme)

	return nil
}

// GetVFS 获取 VFS
func (c *CoordinatorImpl) GetVFS(scheme string) (VFS, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vfsObj, ok := c.vfsRegistry[scheme]
	if !ok {
		return nil, fmt.Errorf("VFS not found: %s", scheme)
	}

	return vfsObj, nil
}

// ===== 统计信息 =====

// GetStats 获取协调器统计信息
func (c *CoordinatorImpl) GetStats(ctx context.Context) (*CoordinatorStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &CoordinatorStats{
		SyncCount:     c.syncCount,
		LastSyncTime:  c.lastSync,
		LayerGenStats: c.layerGenStats,
		MemoryStats:   c.memoryStats,
	}

	// 转换 VFS 统计
	stats.VFSStats = make(map[string]int64)
	for scheme := range c.vfsRegistry {
		stats.VFSStats[scheme] = 1
	}

	return stats, nil
}

// ===== 内部方法 =====

// syncLoop 同步循环
func (c *CoordinatorImpl) syncLoop() {
	ticker := time.NewTicker(c.syncInterval)
	defer ticker.Stop()

	// 如果启用了压缩，添加单独的压缩定时器
	var compressionTicker *time.Ticker
	if c.compressionEnabled && c.memoryConfig != nil {
		compressionTicker = time.NewTicker(c.memoryConfig.CompressionInterval)
		defer compressionTicker.Stop()
	}

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := c.performSync(ctx); err != nil {
				fmt.Printf("Sync error: %v\n", err)
			}

			// 每次同步时也清理过期记忆
			if err := c.CleanupExpiredMemories(ctx); err != nil {
				fmt.Printf("Memory cleanup error: %v\n", err)
			}

		case <-compressionTicker.C:
			// 定期压缩记忆
			if c.compressionEnabled {
				ctx := context.Background()
				if err := c.CompressAllMemories(ctx); err != nil {
					fmt.Printf("Memory compression error: %v\n", err)
				}
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// performSync 执行同步
func (c *CoordinatorImpl) performSync(ctx context.Context) error {
	// 同步任务到上下文
	if err := c.SyncTasksToContexts(ctx); err != nil {
		return fmt.Errorf("sync tasks to contexts: %w", err)
	}

	// 同步上下文到任务
	if err := c.SyncContextsToTasks(ctx); err != nil {
		return fmt.Errorf("sync contexts to tasks: %w", err)
	}

	c.syncCount++
	c.lastSync = time.Now()

	return nil
}

// getAllTasks 获取所有任务
func (c *CoordinatorImpl) getAllTasks(ctx context.Context) ([]*beads.Task, error) {
	var allTasks []*beads.Task

	// 获取各种状态的任务
	statuses := []beads.TaskStatus{
		beads.StatusOpen,
		beads.StatusInProgress,
		beads.StatusBlocked,
	}

	for _, status := range statuses {
		result, err := c.taskTracker.GetByStatus(ctx, status)
		if err != nil {
			continue
		}

		// 直接追加结果（假设返回类型正确）
		allTasks = append(allTasks, result...)
	}

	return allTasks, nil
}

// getAllContexts 获取所有上下文
func (c *CoordinatorImpl) getAllContexts(ctx context.Context) ([]*Context, error) {
	// 简化实现：返回空列表
	// 实际需要从 ContextStore 查询所有上下文
	return []*Context{}, nil
}

// syncTaskToContext 同步任务到上下文
func (c *CoordinatorImpl) syncTaskToContext(ctx context.Context, task *beads.Task) error {
	// 检查是否已有关联的上下文
	contexts, err := c.contextStore.GetTaskContexts(ctx, task.ID)
	if err != nil || len(contexts) == 0 {
		// 创建新上下文
		ctxt := NewContext(ContextTypeTask, task.Title)
		ctxt.Metadata = map[string]string{
			"task_id": task.ID,
			"status":  string(task.Status),
		}
		ctxt.URI = "viking://tasks/" + task.ID

		// 生成三层内容
		if task.Description != "" {
			layers, err := c.generateAllLayers(ctx, task.Description, "plain")
			if err == nil {
				ctxt.Layers = *layers
			}
		}

		// 创建上下文
		contextID, err := c.contextStore.CreateContext(ctx, ctxt)
		if err != nil {
			return fmt.Errorf("create context: %w", err)
		}

		// 关联任务
		return c.contextStore.AssociateContext(ctx, task.ID, contextID)
	}

	return nil
}

// syncContextToTask 同步上下文到任务
func (c *CoordinatorImpl) syncContextToTask(ctx context.Context, ctxt *Context) error {
	// 检查是否需要生成缺失的层级
	if err := c.GenerateMissingLayers(ctx, ctxt.ID); err != nil {
		fmt.Printf("Warning: failed to generate missing layers for context %s: %v\n", ctxt.ID, err)
	}

	// 检查是否需要提取记忆
	if ctxt.Memories == nil || ctxt.Memories.GetMemoryCount() == 0 {
		if err := c.ExtractAndMergeMemories(ctx, ctxt.ID); err != nil {
			fmt.Printf("Warning: failed to extract memories for context %s: %v\n", ctxt.ID, err)
		}
	}

	return nil
}

// findMissingLayers 查找缺失的层级
func (c *CoordinatorImpl) findMissingLayers(ctxt *Context) []LayerType {
	var missing []LayerType

	if ctxt.Layers.L0 == nil {
		missing = append(missing, LayerTypeL0)
	}
	if ctxt.Layers.L1 == nil {
		missing = append(missing, LayerTypeL1)
	}

	return missing
}

// updateMemoryStats 更新记忆统计
func (c *CoordinatorImpl) updateMemoryStats(memories *MemoryCollection) {
	c.memoryStats.TotalMemories += int64(memories.GetMemoryCount())

	if len(memories.Profiles) > 0 {
		c.memoryStats.ByType[MemoryTypeProfile] += int64(len(memories.Profiles))
	}
	if len(memories.Preferences) > 0 {
		c.memoryStats.ByType[MemoryTypePreference] += int64(len(memories.Preferences))
	}
	if len(memories.Entities) > 0 {
		c.memoryStats.ByType[MemoryTypeEntity] += int64(len(memories.Entities))
	}
	if len(memories.Events) > 0 {
		c.memoryStats.ByType[MemoryTypeEvent] += int64(len(memories.Events))
	}
	if len(memories.Cases) > 0 {
		c.memoryStats.ByType[MemoryTypeCase] += int64(len(memories.Cases))
	}
	if len(memories.Patterns) > 0 {
		c.memoryStats.ByType[MemoryTypePattern] += int64(len(memories.Patterns))
	}
}

// ===== 辅助方法（安全调用 layerGen 和 memoryExt） =====

// getLayerGenerator 获取层级生成器
func (c *CoordinatorImpl) getLayerGenerator() (layerGenerator, bool) {
	if c.layerGen == nil {
		return nil, false
	}
	gen, ok := c.layerGen.(layerGenerator)
	return gen, ok
}

// getMemoryExtractor 获取记忆提取器
func (c *CoordinatorImpl) getMemoryExtractor() (memoryExtractor, bool) {
	if c.memoryExt == nil {
		return nil, false
	}
	ext, ok := c.memoryExt.(memoryExtractor)
	return ext, ok
}

// generateL0 生成 L0 层级
func (c *CoordinatorImpl) generateL0(ctx context.Context, content string) (*LayerSummary, error) {
	gen, ok := c.getLayerGenerator()
	if !ok {
		// 返回一个简单的 L0 摘要
		return &LayerSummary{
			Content:     content[:min(len(content), 400)],
			Tokens:      min(len(content), 400) / 4,
			GeneratedAt: time.Now(),
			Method:      "simple",
		}, nil
	}
	return gen.GenerateL0(ctx, content)
}

// generateL1 生成 L1 层级
func (c *CoordinatorImpl) generateL1(ctx context.Context, content string) (*LayerOverview, error) {
	gen, ok := c.getLayerGenerator()
	if !ok {
		// 返回一个简单的 L1 概览
		return &LayerOverview{
			Content:     content[:min(len(content), 8000)],
			Tokens:      min(len(content), 8000) / 4,
			Sections:    []string{"Overview"},
			KeyPoints:   []string{"Key point"},
			GeneratedAt: time.Now(),
			Method:      "simple",
		}, nil
	}
	return gen.GenerateL1(ctx, content)
}

// generateAllLayers 生成所有层级
func (c *CoordinatorImpl) generateAllLayers(ctx context.Context, content string, format string) (*ContextLayers, error) {
	gen, ok := c.getLayerGenerator()
	if !ok {
		// 返回简单的层级
		l0, _ := c.generateL0(ctx, content)
		l1, _ := c.generateL1(ctx, content)
		return &ContextLayers{
			L0: l0,
			L1: l1,
			L2: &LayerDetails{
				Content:     content,
				Tokens:      len(content) / 4,
				Format:      format,
				GeneratedAt: time.Now(),
			},
		}, nil
	}
	return gen.GenerateAll(ctx, content, format)
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractMemories 提取记忆
func (c *CoordinatorImpl) extractMemories(ctx context.Context, ctxt *Context) (*MemoryCollection, error) {
	ext, ok := c.getMemoryExtractor()
	if !ok {
		return nil, fmt.Errorf("memory extractor not available")
	}
	return ext.ExtractFromContext(ctx, ctxt)
}

// mergeMemories 合并记忆
func (c *CoordinatorImpl) mergeMemories(ctx context.Context, existing, newMemories *MemoryCollection) (*MemoryCollection, error) {
	ext, ok := c.getMemoryExtractor()
	if !ok {
		return nil, fmt.Errorf("memory extractor not available")
	}
	return ext.Merge(ctx, existing, newMemories)
}

// deduplicateMemories 去重记忆
func (c *CoordinatorImpl) deduplicateMemories(ctx context.Context, memories *MemoryCollection) (*MemoryCollection, error) {
	ext, ok := c.getMemoryExtractor()
	if !ok {
		return nil, fmt.Errorf("memory extractor not available")
	}
	return ext.Deduplicate(ctx, memories)
}

// SetLayerGenerator 设置层级生成器（由外部调用）
func (c *CoordinatorImpl) SetLayerGenerator(gen interface{}) {
	c.layerGen = gen
}

// SetMemoryExtractor 设置记忆提取器（由外部调用）
func (c *CoordinatorImpl) SetMemoryExtractor(ext interface{}) {
	c.memoryExt = ext
}

// SetMemoryCompressor 设置记忆压缩器（由外部调用）
func (c *CoordinatorImpl) SetMemoryCompressor(compressor interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memoryCompressor = compressor
	c.compressionEnabled = (compressor != nil)
}

// SetMemoryConfig 设置记忆压缩配置
func (c *CoordinatorImpl) SetMemoryConfig(config *MemoryCompressionConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memoryConfig = config
}

// GetMemoryConfig 获取记忆压缩配置
func (c *CoordinatorImpl) GetMemoryConfig() *MemoryCompressionConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.memoryConfig
}

// ===== 记忆压缩管理 =====

// CompressAllMemories 压缩所有分层的记忆
// 将短期记忆压缩并转移到长期存储
func (c *CoordinatorImpl) CompressAllMemories(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.compressionEnabled || c.memoryCompressor == nil {
		return fmt.Errorf("memory compression not enabled")
	}

	// 获取所有上下文
	contexts, err := c.getAllContexts(ctx)
	if err != nil {
		return fmt.Errorf("get contexts: %w", err)
	}

	// 压缩每个上下文的记忆
	for _, ctxt := range contexts {
		if ctxt.Memories == nil || ctxt.Memories.GetMemoryCount() == 0 {
			continue
		}

		// 获取压缩器
		compressor, ok := c.memoryCompressor.(memoryCompressor)
		if !ok {
			return fmt.Errorf("invalid memory compressor type")
		}

		// 检查是否需要压缩
		config := c.memoryConfig
		if config == nil {
			config = DefaultMemoryCompressionConfig()
		}

		// 检查是否超过数量限制
		memoryCount := ctxt.Memories.GetMemoryCount()
		if memoryCount < config.MaxSessionMemories {
			continue
		}

		// 压缩记忆到长期存储
		compressed, err := compressor.CompressMemories(ctx, ctxt.Memories, MemoryTierLongTerm)
		if err != nil {
			fmt.Printf("Warning: failed to compress memories for context %s: %v\n", ctxt.ID, err)
			continue
		}

		// 更新上下文
		if err := c.contextStore.UpdateMemories(ctx, ctxt.ID, compressed); err != nil {
			fmt.Printf("Warning: failed to update compressed memories for context %s: %v\n", ctxt.ID, err)
		}
	}

	return nil
}

// CleanupExpiredMemories 清理过期的分层记忆
func (c *CoordinatorImpl) CleanupExpiredMemories(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否是支持分层操作的存储
	if tierStore, ok := c.contextStore.(interface {
		CleanupExpiredTiers(ctx context.Context) (int, error)
	}); ok {
		count, err := tierStore.CleanupExpiredTiers(ctx)
		if err != nil {
			return fmt.Errorf("cleanup expired tiers: %w", err)
		}
		fmt.Printf("Cleaned up %d expired memories\n", count)
	}

	return nil
}

// CompressContextMemories 压缩指定上下文的记忆
func (c *CoordinatorImpl) CompressContextMemories(ctx context.Context, contextID string, targetTier MemoryTier) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.compressionEnabled || c.memoryCompressor == nil {
		return fmt.Errorf("memory compression not enabled")
	}

	// 获取上下文
	ctxt, err := c.contextStore.GetContext(ctx, contextID)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	if ctxt.Memories == nil || ctxt.Memories.GetMemoryCount() == 0 {
		return fmt.Errorf("no memories to compress")
	}

	// 获取压缩器
	compressor, ok := c.memoryCompressor.(memoryCompressor)
	if !ok {
		return fmt.Errorf("invalid memory compressor type")
	}

	// 压缩记忆
	compressed, err := compressor.CompressMemories(ctx, ctxt.Memories, targetTier)
	if err != nil {
		return fmt.Errorf("compress memories: %w", err)
	}

	// 更新上下文
	return c.contextStore.UpdateMemories(ctx, contextID, compressed)
}

// PromoteMemories 提升记忆到更高分层
func (c *CoordinatorImpl) PromoteMemories(ctx context.Context, contextID string, fromTier, toTier MemoryTier) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否支持分层操作
	if tierStore, ok := c.contextStore.(interface {
		GetMemoriesByTier(ctx context.Context, tier MemoryTier, limit int) ([]string, error)
		PromoteMemories(ctx context.Context, memoryIDs []string, fromTier, toTier MemoryTier) error
	}); ok {
		// 获取源分层的记忆
		memoryIDs, err := tierStore.GetMemoriesByTier(ctx, fromTier, 0)
		if err != nil {
			return fmt.Errorf("get memories by tier: %w", err)
		}

		// 提升分层
		return tierStore.PromoteMemories(ctx, memoryIDs, fromTier, toTier)
	}

	return fmt.Errorf("context store does not support tier operations")
}

// GetMemoryStatistics 获取记忆统计信息
func (c *CoordinatorImpl) GetMemoryStatistics(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})

	// 基本统计
	stats["total_contexts"] = len(c.vfsRegistry) // 简化处理
	stats["memory_stats"] = c.memoryStats
	stats["layer_gen_stats"] = c.layerGenStats

	// 分层统计（如果支持）
	if tierStore, ok := c.contextStore.(interface {
		GetTierStatistics(ctx context.Context) (map[MemoryTier]int, error)
	}); ok {
		tierStats, err := tierStore.GetTierStatistics(ctx)
		if err == nil {
			stats["tier_stats"] = tierStats
		}
	}

	return stats, nil
}
