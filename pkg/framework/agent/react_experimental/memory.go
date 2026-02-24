//go:build experimental

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
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent"
	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/errors"
)

// MemoryIntegration 内存集成器接口
// 【必须】定义ReAct内存集成的标准接口
type MemoryIntegration interface {
	// StoreThought 存储思考步骤
	// 【必须】将思考步骤持久化到内存系统
	StoreThought(ctx context.Context, thought *Thought, state *ReActState) error
	// StoreAction 存储动作步骤
	// 【必须】将动作步骤持久化到内存系统
	StoreAction(ctx context.Context, action *Action, state *ReActState) error
	// StoreObservation 存储观察结果
	// 【必须】将观察结果持久化到内存系统
	StoreObservation(ctx context.Context, observation *Observation, state *ReActState) error
	// RetrieveRelevant 检索相关内容
	// 【必须】基于查询检索相关的历史信息
	RetrieveRelevant(ctx context.Context, query string, limit int, state *ReActState) ([]memory.MemoryItem, error)
	// GetSessionHistory 获取会话历史
	// 【必须】获取当前会话的完整历史记录
	GetSessionHistory(ctx context.Context, sessionID string, limit int) ([]memory.MemoryItem, error)
	// BuildContext 构建上下文
	// 【必须】基于当前状态构建决策上下文
	BuildContext(ctx context.Context, state *ReActState) (*ReActContext, error)
	// ClearSession 清理会话数据
	// 【必须】清理指定会话的内存数据
	ClearSession(ctx context.Context, sessionID string) error
	// Validate 验证集成器配置
	// 【必须】验证集成器自身的配置有效性
	Validate() error
	// Name 返回集成器名称
	// 【必须】提供集成器标识
	Name() string
}

// BaseMemoryIntegration 内存集成器基类
// 【推荐】提供基础实现减少重复代码
type BaseMemoryIntegration struct {
	name   string
	logger *zap.Logger
	config *ReActConfig
}

// NewBaseMemoryIntegration 创建基础内存集成器
// 【必须】提供构造函数确保必要字段初始化
func NewBaseMemoryIntegration(name string, logger *zap.Logger, config *ReActConfig) *BaseMemoryIntegration {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = NewReActConfig()
	}

	return &BaseMemoryIntegration{
		name:   name,
		logger: logger.With(zap.String("memory_integration", name)),
		config: config,
	}
}

// Name 返回集成器名称
// 【必须】实现 MemoryIntegration 接口的 Name 方法
func (bmi *BaseMemoryIntegration) Name() string {
	return bmi.name
}

// Validate 验证集成器配置
// 【必须】实现 MemoryIntegration 接口的 Validate 方法
func (bmi *BaseMemoryIntegration) Validate() error {
	if bmi.name == "" {
		return errors.NewValidationError("integration name cannot be empty", nil)
	}

	if bmi.config == nil {
		return errors.NewValidationError("config cannot be nil", nil)
	}

	return bmi.config.Validate()
}

// DefaultMemoryIntegration 默认内存集成器实现
// 【必须】提供开箱即用的内存集成实现
type DefaultMemoryIntegration struct {
	BaseMemoryIntegration
	// MemoryManager 内存管理器
	MemoryManager memory.Manager
	// ContextBuilder 上下文构建器
	ContextBuilder *ContextBuilder
	// RetentionPolicy 保留策略
	RetentionPolicy RetentionPolicy
}

// RetentionPolicy 保留策略配置
// 【推荐】定义内存数据的保留规则
type RetentionPolicy struct {
	// MaxSessionItems 每会话最大条目数
	MaxSessionItems int `json:"max_session_items"`
	// MaxTotalItems 总最大条目数
	MaxTotalItems int `json:"max_total_items"`
	// ItemTTL 条目生存时间
	ItemTTL time.Duration `json:"item_ttl"`
	// EnableAutoCleanup 是否启用自动清理
	EnableAutoCleanup bool `json:"enable_auto_cleanup"`
}

// NewDefaultMemoryIntegration 创建默认内存集成器
// 【必须】提供构造函数确保必要配置
func NewDefaultMemoryIntegration(logger *zap.Logger, config *ReActConfig, memoryManager memory.Manager) *DefaultMemoryIntegration {
	integration := &DefaultMemoryIntegration{
		BaseMemoryIntegration: *NewBaseMemoryIntegration("default_memory_integration", logger, config),
		MemoryManager:         memoryManager,
		ContextBuilder:        NewContextBuilder(logger, config),
		RetentionPolicy: RetentionPolicy{
			MaxSessionItems:    1000,
			MaxTotalItems:      10000,
			ItemTTL:            24 * time.Hour,
			EnableAutoCleanup:  true,
		},
	}

	// 【必须】设置合理的默认值
	if integration.MemoryManager == nil {
		integration.logger.Warn("memory manager is nil, memory integration will be limited")
	}

	if integration.RetentionPolicy.MaxSessionItems <= 0 {
		integration.RetentionPolicy.MaxSessionItems = 1000
	}

	if integration.RetentionPolicy.MaxTotalItems <= 0 {
		integration.RetentionPolicy.MaxTotalItems = 10000
	}

	if integration.RetentionPolicy.ItemTTL <= 0 {
		integration.RetentionPolicy.ItemTTL = 24 * time.Hour
	}

	return integration
}

// StoreThought 存储思考步骤
// 【必须】实现 MemoryIntegration 接口的 StoreThought 方法
func (dmi *DefaultMemoryIntegration) StoreThought(ctx context.Context, thought *Thought, state *ReActState) error {
	if thought == nil {
		return errors.NewValidationError("thought cannot be nil", nil)
	}

	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录存储开始
	dmi.logger.Debug("storing thought in memory",
		zap.String("thought_id", thought.ID),
		zap.String("session_id", state.SessionID),
		zap.Float64("confidence", thought.Confidence),
		zap.Int("iteration", state.IterationCount),
	)

	if dmi.MemoryManager == nil {
		return errors.NewValidationError("memory manager is not available", nil)
	}

	// 创建内存条目
	item, err := dmi.createThoughtMemoryItem(thought, state)
	if err != nil {
		return errors.WrapError(err, "failed to create thought memory item", map[string]interface{}{
			"thought_id": thought.ID,
		})
	}

	// 存储到内存系统
	if err := dmi.MemoryManager.Store(ctx, item); err != nil {
		dmi.logger.Error("failed to store thought in memory",
			zap.Error(err),
			zap.String("thought_id", thought.ID),
		)
		return errors.WrapError(err, "memory storage failed", map[string]interface{}{
			"thought_id": thought.ID,
			"session_id": state.SessionID,
		})
	}

	// 更新上下文关联
	if dmi.config.Tracker != nil {
		if err := dmi.config.Tracker.AssociateContext(ctx, state.SessionID, "thought", thought.ID); err != nil {
			dmi.logger.Warn("failed to associate thought context",
				zap.Error(err),
				zap.String("thought_id", thought.ID),
				zap.String("session_id", state.SessionID),
			)
		}
	}

	// 【必须】记录存储成功
	dmi.logger.Info("thought stored in memory successfully",
		zap.String("thought_id", thought.ID),
		zap.String("session_id", state.SessionID),
		zap.String("memory_item_id", item.ID),
	)

	// 检查并执行自动清理
	if dmi.RetentionPolicy.EnableAutoCleanup {
		go dmi.performAutoCleanup(ctx, state.SessionID)
	}

	return nil
}

// StoreAction 存储动作步骤
// 【必须】实现 MemoryIntegration 接口的 StoreAction 方法
func (dmi *DefaultMemoryIntegration) StoreAction(ctx context.Context, action *Action, state *ReActState) error {
	if action == nil {
		return errors.NewValidationError("action cannot be nil", nil)
	}

	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录存储开始
	dmi.logger.Debug("storing action in memory",
		zap.String("action_id", action.ID),
		zap.String("session_id", state.SessionID),
		zap.String("action_type", action.Type.String()),
		zap.String("action_name", action.Name),
	)

	if dmi.MemoryManager == nil {
		return errors.NewValidationError("memory manager is not available", nil)
	}

	// 创建内存条目
	item, err := dmi.createActionMemoryItem(action, state)
	if err != nil {
		return errors.WrapError(err, "failed to create action memory item", map[string]interface{}{
			"action_id": action.ID,
		})
	}

	// 存储到内存系统
	if err := dmi.MemoryManager.Store(ctx, item); err != nil {
		dmi.logger.Error("failed to store action in memory",
			zap.Error(err),
			zap.String("action_id", action.ID),
		)
		return errors.WrapError(err, "memory storage failed", map[string]interface{}{
			"action_id":  action.ID,
			"session_id": state.SessionID,
		})
	}

	// 更新上下文关联
	if dmi.config.Tracker != nil {
		if err := dmi.config.Tracker.AssociateContext(ctx, state.SessionID, "action", action.ID); err != nil {
			dmi.logger.Warn("failed to associate action context",
				zap.Error(err),
				zap.String("action_id", action.ID),
				zap.String("session_id", state.SessionID),
			)
		}
	}

	// 【必须】记录存储成功
	dmi.logger.Info("action stored in memory successfully",
		zap.String("action_id", action.ID),
		zap.String("session_id", state.SessionID),
		zap.String("memory_item_id", item.ID),
	)

	return nil
}

// StoreObservation 存储观察结果
// 【必须】实现 MemoryIntegration 接口的 StoreObservation 方法
func (dmi *DefaultMemoryIntegration) StoreObservation(ctx context.Context, observation *Observation, state *ReActState) error {
	if observation == nil {
		return errors.NewValidationError("observation cannot be nil", nil)
	}

	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录存储开始
	dmi.logger.Debug("storing observation in memory",
		zap.String("observation_id", observation.ID),
		zap.String("session_id", state.SessionID),
		zap.String("action_id", observation.ActionID),
		zap.Bool("success", observation.Success),
	)

	if dmi.MemoryManager == nil {
		return errors.NewValidationError("memory manager is not available", nil)
	}

	// 创建内存条目
	item, err := dmi.createObservationMemoryItem(observation, state)
	if err != nil {
		return errors.WrapError(err, "failed to create observation memory item", map[string]interface{}{
			"observation_id": observation.ID,
		})
	}

	// 存储到内存系统
	if err := dmi.MemoryManager.Store(ctx, item); err != nil {
		dmi.logger.Error("failed to store observation in memory",
			zap.Error(err),
			zap.String("observation_id", observation.ID),
		)
		return errors.WrapError(err, "memory storage failed", map[string]interface{}{
			"observation_id": observation.ID,
			"session_id":     state.SessionID,
		})
	}

	// 更新上下文关联
	if dmi.config.Tracker != nil {
		if err := dmi.config.Tracker.AssociateContext(ctx, state.SessionID, "observation", observation.ID); err != nil {
			dmi.logger.Warn("failed to associate observation context",
				zap.Error(err),
				zap.String("observation_id", observation.ID),
				zap.String("session_id", state.SessionID),
			)
		}
	}

	// 【必须】记录存储成功
	dmi.logger.Info("observation stored in memory successfully",
		zap.String("observation_id", observation.ID),
		zap.String("session_id", state.SessionID),
		zap.String("memory_item_id", item.ID),
		zap.Bool("success", observation.Success),
	)

	return nil
}

// RetrieveRelevant 检索相关内容
// 【必须】实现 MemoryIntegration 接口的 RetrieveRelevant 方法
func (dmi *DefaultMemoryIntegration) RetrieveRelevant(ctx context.Context, query string, limit int, state *ReActState) ([]memory.MemoryItem, error) {
	if query == "" {
		return nil, errors.NewValidationError("query cannot be empty", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录检索开始
	dmi.logger.Debug("retrieving relevant memory items",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.String("session_id", state.SessionID),
	)

	if dmi.MemoryManager == nil {
		return nil, errors.NewValidationError("memory manager is not available", nil)
	}

	// 设置默认限制
	if limit <= 0 {
		limit = 10
	}

	// 构建检索查询
	retrievalQuery := dmi.buildRetrievalQuery(query, state)

	// 执行检索
	items, err := dmi.MemoryManager.Retrieve(ctx, retrievalQuery, limit)
	if err != nil {
		dmi.logger.Error("failed to retrieve relevant items",
			zap.Error(err),
			zap.String("query", query),
		)
		return nil, errors.WrapError(err, "memory retrieval failed", map[string]interface{}{
			"query":       query,
			"session_id":  state.SessionID,
			"limit":       limit,
		})
	}

	// 过滤和排序结果
	filteredItems := dmi.filterAndSortItems(items, query, state)

	// 【必须】记录检索成功
	dmi.logger.Info("retrieved relevant memory items",
		zap.String("query", query),
		zap.Int("requested_limit", limit),
		zap.Int("returned_count", len(filteredItems)),
		zap.String("session_id", state.SessionID),
	)

	return filteredItems, nil
}

// GetSessionHistory 获取会话历史
// 【必须】实现 MemoryIntegration 接口的 GetSessionHistory 方法
func (dmi *DefaultMemoryIntegration) GetSessionHistory(ctx context.Context, sessionID string, limit int) ([]memory.MemoryItem, error) {
	if sessionID == "" {
		return nil, errors.NewValidationError("session ID cannot be empty", nil)
	}

	// 【必须】记录检索开始
	dmi.logger.Debug("retrieving session history",
		zap.String("session_id", sessionID),
		zap.Int("limit", limit),
	)

	if dmi.MemoryManager == nil {
		return nil, errors.NewValidationError("memory manager is not available", nil)
	}

	// 设置默认限制
	if limit <= 0 {
		limit = 100
	}

	// 构建会话查询
	sessionQuery := memory.Query{
		Tags:      map[string]string{"session_id": sessionID},
		Limit:     limit,
		OrderBy:   "timestamp",
		OrderDesc: true, // 最新的在前
	}

	// 执行查询
	items, err := dmi.MemoryManager.Retrieve(ctx, sessionQuery, limit)
	if err != nil {
		dmi.logger.Error("failed to retrieve session history",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		return nil, errors.WrapError(err, "session history retrieval failed", map[string]interface{}{
			"session_id": sessionID,
			"limit":      limit,
		})
	}

	// 【必须】记录检索成功
	dmi.logger.Info("retrieved session history",
		zap.String("session_id", sessionID),
		zap.Int("limit", limit),
		zap.Int("returned_count", len(items)),
	)

	return items, nil
}

// BuildContext 构建上下文
// 【必须】实现 MemoryIntegration 接口的 BuildContext 方法
func (dmi *DefaultMemoryIntegration) BuildContext(ctx context.Context, state *ReActState) (*ReActContext, error) {
	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录构建开始
	dmi.logger.Debug("building React context",
		zap.String("session_id", state.SessionID),
		zap.Int("iteration", state.IterationCount),
	)

	// 委托给上下文构建器
	reactContext, err := dmi.ContextBuilder.Build(ctx, state)
	if err != nil {
		dmi.logger.Error("failed to build React context",
			zap.Error(err),
			zap.String("session_id", state.SessionID),
		)
		return nil, errors.WrapError(err, "context building failed", map[string]interface{}{
			"session_id": state.SessionID,
		})
	}

	// 【必须】记录构建成功
	dmi.logger.Info("React context built successfully",
		zap.String("session_id", state.SessionID),
		zap.Int("thoughts_count", len(reactContext.RecentThoughts)),
		zap.Int("actions_count", len(reactContext.RecentActions)),
		zap.Int("observations_count", len(reactContext.RecentObservations)),
		zap.Int("relevant_memories_count", len(reactContext.RelevantMemories)),
	)

	return reactContext, nil
}

// ClearSession 清理会话数据
// 【必须】实现 MemoryIntegration 接口的 ClearSession 方法
func (dmi *DefaultMemoryIntegration) ClearSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.NewValidationError("session ID cannot be empty", nil)
	}

	// 【必须】记录清理开始
	dmi.logger.Debug("clearing session data from memory",
		zap.String("session_id", sessionID),
	)

	if dmi.MemoryManager == nil {
		return errors.NewValidationError("memory manager is not available", nil)
	}

	// 获取会话的所有内存条目
	sessionItems, err := dmi.GetSessionHistory(ctx, sessionID, dmi.RetentionPolicy.MaxSessionItems)
	if err != nil {
		dmi.logger.Error("failed to get session items for cleanup",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		return errors.WrapError(err, "failed to retrieve session items", map[string]interface{}{
			"session_id": sessionID,
		})
	}

	// 删除所有会话条目
	deletedCount := 0
	for _, item := range sessionItems {
		if err := dmi.MemoryManager.Delete(ctx, item.ID); err != nil {
			dmi.logger.Warn("failed to delete session item",
				zap.Error(err),
				zap.String("item_id", item.ID),
				zap.String("session_id", sessionID),
			)
		} else {
			deletedCount++
		}
	}

	// 清理上下文关联
	if dmi.config.Tracker != nil {
		// 【必须】这里应该清理tracker中的上下文关联
		// 由于tracker接口限制，这里只记录日志
		dmi.logger.Debug("cleaning up tracker associations for session",
			zap.String("session_id", sessionID),
		)
	}

	// 【必须】记录清理成功
	dmi.logger.Info("session data cleared from memory",
		zap.String("session_id", sessionID),
		zap.Int("items_deleted", deletedCount),
		zap.Int("total_items_found", len(sessionItems)),
	)

	return nil
}

// createThoughtMemoryItem 创建思考内存条目
// 【推荐】将思考对象转换为内存条目
func (dmi *DefaultMemoryIntegration) createThoughtMemoryItem(thought *Thought, state *ReActState) (*memory.MemoryItem, error) {
	content := fmt.Sprintf("Thought: %s", thought.Content)
	if thought.Reasoning != "" {
		content += fmt.Sprintf("\nReasoning: %s", thought.Reasoning)
	}

	item := &memory.MemoryItem{
		ID:        uuid.New().String(),
		Content:   content,
		Type:      memory.MemoryTypeShortTerm,
		Timestamp: thought.Timestamp,
		Tags: map[string]string{
			"type":       "thought",
			"session_id": state.SessionID,
			"iteration":  fmt.Sprintf("%d", state.IterationCount),
			"confidence": fmt.Sprintf("%.2f", thought.Confidence),
		},
		Metadata: map[string]interface{}{
			"thought_id":     thought.ID,
			"content":        thought.Content,
			"reasoning":      thought.Reasoning,
			"confidence":     thought.Confidence,
			"associated_contexts": thought.AssociatedContexts,
			"timestamp_unix": thought.Timestamp.Unix(),
		},
		ExpiresAt: time.Now().Add(dmi.RetentionPolicy.ItemTTL),
	}

	return item, nil
}

// createActionMemoryItem 创建动作内存条目
// 【推荐】将动作对象转换为内存条目
func (dmi *DefaultMemoryIntegration) createActionMemoryItem(action *Action, state *ReActState) (*memory.MemoryItem, error) {
	content := fmt.Sprintf("Action: %s (%s)", action.Name, action.Type.String())
	if action.Description != "" {
		content += fmt.Sprintf("\nDescription: %s", action.Description)
	}
	if len(action.Parameters) > 0 {
		content += fmt.Sprintf("\nParameters: %v", action.Parameters)
	}

	item := &memory.MemoryItem{
		ID:        uuid.New().String(),
		Content:   content,
		Type:      memory.MemoryTypeShortTerm,
		Timestamp: action.Timestamp,
		Tags: map[string]string{
			"type":       "action",
			"session_id": state.SessionID,
			"iteration":  fmt.Sprintf("%d", state.IterationCount),
			"action_type": action.Type.String(),
		},
		Metadata: map[string]interface{}{
			"action_id":           action.ID,
			"name":                action.Name,
			"type":                action.Type.String(),
			"description":         action.Description,
			"parameters":          action.Parameters,
			"associated_contexts": action.AssociatedContexts,
			"timestamp_unix":      action.Timestamp.Unix(),
		},
		ExpiresAt: time.Now().Add(dmi.RetentionPolicy.ItemTTL),
	}

	return item, nil
}

// createObservationMemoryItem 创建观察内存条目
// 【推荐】将观察对象转换为内存条目
func (dmi *DefaultMemoryIntegration) createObservationMemoryItem(observation *Observation, state *ReActState) (*memory.MemoryItem, error) {
	content := fmt.Sprintf("Observation: Action %s - %s", observation.ActionID, observation.ResultSummary())
	if !observation.Success {
		content += fmt.Sprintf(" (Failed: %s)", observation.Error)
	}

	item := &memory.MemoryItem{
		ID:        uuid.New().String(),
		Content:   content,
		Type:      memory.MemoryTypeShortTerm,
		Timestamp: observation.Timestamp,
		Tags: map[string]string{
			"type":       "observation",
			"session_id": state.SessionID,
			"iteration":  fmt.Sprintf("%d", state.IterationCount),
			"success":    fmt.Sprintf("%t", observation.Success),
		},
		Metadata: map[string]interface{}{
			"observation_id":      observation.ID,
			"action_id":           observation.ActionID,
			"success":              observation.Success,
			"error":                observation.Error,
			"execution_time_ms":    observation.ExecutionTime.Milliseconds(),
			"result_data":          observation.Result,
			"associated_contexts": observation.AssociatedContexts,
			"timestamp_unix":       observation.Timestamp.Unix(),
		},
		ExpiresAt: time.Now().Add(dmi.RetentionPolicy.ItemTTL),
	}

	return item, nil
}

// buildRetrievalQuery 构建检索查询
// 【推荐】基于查询和状态构建内存检索查询
func (dmi *DefaultMemoryIntegration) buildRetrievalQuery(query string, state *ReActState) memory.Query {
	// 基础查询
	retrievalQuery := memory.Query{
		Text:  query,
		Limit: 10,
		Tags:  make(map[string]string),
	}

	// 添加会话过滤
	retrievalQuery.Tags["session_id"] = state.SessionID

	// 添加迭代范围过滤（最近几次迭代）
	if state.IterationCount > 0 {
		startIteration := state.IterationCount - 5 // 最近5次迭代
		if startIteration < 0 {
			startIteration = 0
		}
		// 注意：这里假设memory.Query支持复杂过滤，实际可能需要不同的实现
	}

	// 设置排序
	retrievalQuery.OrderBy = "timestamp"
	retrievalQuery.OrderDesc = true

	return retrievalQuery
}

// filterAndSortItems 过滤和排序结果
// 【推荐】对检索结果进行后处理
func (dmi *DefaultMemoryIntegration) filterAndSortItems(items []memory.MemoryItem, query string, state *ReActState) []memory.MemoryItem {
	if len(items) == 0 {
		return items
	}

	// 过滤掉过期的条目
	filtered := make([]memory.MemoryItem, 0)
	now := time.Now()
	for _, item := range items {
		if item.ExpiresAt.IsZero() || item.ExpiresAt.After(now) {
			filtered = append(filtered, item)
		}
	}

	// 基于相关性进行排序（简化实现）
	// 实际应用中可以使用更复杂的相似度算法
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if dmi.calculateRelevanceScore(filtered[i], query) < dmi.calculateRelevanceScore(filtered[j], query) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	return filtered
}

// calculateRelevanceScore 计算相关性评分
// 【推荐】计算内存条目与查询的相关性
func (dmi *DefaultMemoryIntegration) calculateRelevanceScore(item memory.MemoryItem, query string) float64 {
	// 简化的相关性计算
	// 实际应用中应该使用TF-IDF、向量相似度等算法
	score := 0.0

	// 基于文本内容匹配
	content := strings.ToLower(item.Content)
	queryLower := strings.ToLower(query)

	// 精确匹配
	if strings.Contains(content, queryLower) {
		score += 1.0
	}

	// 单词级别的匹配
	queryWords := strings.Fields(queryLower)
	for _, word := range queryWords {
		if strings.Contains(content, word) {
			score += 0.5
		}
	}

	// 基于时间戳的新鲜度加分
	age := time.Since(item.Timestamp)
	if age < time.Hour {
		score += 0.3
	} else if age < 24*time.Hour {
		score += 0.1
	}

	// 基于类型的权重
	switch item.Tags["type"] {
	case "thought":
		score *= 1.2 // 思考步骤更重要
	case "observation":
		if item.Tags["success"] == "false" {
			score *= 1.5 // 失败观察更重要
		}
	}

	return score
}

// performAutoCleanup 执行自动清理
// 【推荐】后台清理过期或过多的内存条目
func (dmi *DefaultMemoryIntegration) performAutoCleanup(ctx context.Context, sessionID string) {
	// 避免阻塞主流程，使用goroutine执行
	go func() {
		// 检查会话条目数量
		items, err := dmi.GetSessionHistory(ctx, sessionID, dmi.RetentionPolicy.MaxSessionItems+1)
		if err != nil {
			dmi.logger.Warn("failed to check session item count for cleanup", zap.Error(err))
			return
		}

		// 如果超过限制，删除最旧的条目
		if len(items) > dmi.RetentionPolicy.MaxSessionItems {
			itemsToDelete := len(items) - dmi.RetentionPolicy.MaxSessionItems
			for i := 0; i < itemsToDelete; i++ {
				if err := dmi.MemoryManager.Delete(ctx, items[i].ID); err != nil {
					dmi.logger.Warn("failed to delete old session item",
					zap.Error(err),
					zap.String("item_id", items[i].ID),
				)
				} else {
					dmi.logger.Debug("deleted old session item during auto-cleanup",
						zap.String("item_id", items[i].ID),
						zap.String("session_id", sessionID),
					)
				}
			}
		}
	}()
}

// Validate 验证默认内存集成器配置
// 【必须】实现 MemoryIntegration 接口的 Validate 方法
func (dmi *DefaultMemoryIntegration) Validate() error {
	if err := dmi.BaseMemoryIntegration.Validate(); err != nil {
		return errors.WrapError(err, "base integration validation failed", nil)
	}

	if dmi.RetentionPolicy.MaxSessionItems <= 0 {
		return errors.NewValidationError(
			"max session items must be positive",
			map[string]interface{}{"max_session_items": dmi.RetentionPolicy.MaxSessionItems},
		)
	}

	if dmi.RetentionPolicy.MaxTotalItems <= 0 {
		return errors.NewValidationError(
			"max total items must be positive",
			map[string]interface{}{"max_total_items": dmi.RetentionPolicy.MaxTotalItems},
		)
	}

	if dmi.RetentionPolicy.ItemTTL <= 0 {
		return errors.NewValidationError(
			"item TTL must be positive",
			map[string]interface{}{"item_ttl": dmi.RetentionPolicy.ItemTTL},
		)
	}

	return nil
}

// ContextBuilder 上下文构建器
// 【必须】负责构建ReAct决策上下文
type ContextBuilder struct {
	// logger 日志记录器
	logger *zap.Logger
	// config 配置对象
	config *ReActConfig
}

// NewContextBuilder 创建上下文构建器
// 【必须】提供构造函数确保必要初始化
func NewContextBuilder(logger *zap.Logger, config *ReActConfig) *ContextBuilder {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = NewReActConfig()
	}

	return &ContextBuilder{
		logger: logger.Named("context_builder"),
		config: config,
	}
}

// ReActContext ReAct决策上下文
// 【推荐】封装ReAct Agent决策所需的完整上下文
type ReActContext struct {
	// SessionID 会话ID
	SessionID string `json:"session_id"`
	// Query 原始查询
	Query string `json:"query"`
	// CurrentIteration 当前迭代次数
	CurrentIteration int `json:"current_iteration"`
	// MaxIterations 最大迭代次数
	MaxIterations int `json:"max_iterations"`
	// RecentThoughts 最近的思考
	RecentThoughts []*Thought `json:"recent_thoughts"`
	// RecentActions 最近的动极
	RecentActions []*Action `json:"recent_actions"`
	// RecentObservations 最近的观察
	RecentObservations []*Observation `json:"recent_observations"`
	// RelevantMemories 相关记忆
	RelevantMemories []memory.MemoryItem `json:"relevant_memories"`
	// AvailableTools 可用工具
	AvailableTools []tool.BaseTool `json:"available_tools"`
	// ContextSummary 上下文摘要
	ContextSummary string `json:"context_summary"`
	// Timestamp 构建时间
	Timestamp time.Time `json:"timestamp"`
}

// Build 构建ReAct上下文
// 【必须】基于当前状态构建完整的决策上下文
func (cb *ContextBuilder) Build(ctx context.Context, state *ReActState) (*ReActContext, error) {
	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	reactContext := &ReActContext{
		SessionID:          state.SessionID,
		Query:              state.Query,
		CurrentIteration:   state.IterationCount,
		MaxIterations:      state.MaxIterations,
		RecentThoughts:     cb.getRecentThoughts(state),
		RecentActions:      cb.getRecentActions(state),
		RecentObservations: cb.getRecentObservations(state),
		RelevantMemories:   make([]memory.MemoryItem, 0),
		AvailableTools:     make([]tool.BaseTool, 0),
		Timestamp:          time.Now().UTC(),
	}

	// 检索相关记忆
	if state.Memory != nil {
		memories, err := state.Memory.RetrieveRelevant(ctx, state.Query, 5, state)
		if err != nil {
			cb.logger.Warn("failed to retrieve relevant memories", zap.Error(err))
		} else {
			reactContext.RelevantMemories = memories
		}
	}

	// 获取可用工具
	// 【必须】这里应该从工具注册表获取实际工具列表
	// reactContext.AvailableTools = ...

	// 生成上下文摘要
	reactContext.ContextSummary = cb.generateContextSummary(reactContext)

	return reactContext, nil
}

// getRecentThoughts 获取最近的思考
// 【推荐】从状态中提取最近的思考步骤
func (cb *ContextBuilder) getRecentThoughts(state *ReActState) []*Thought {
	if state.Thoughts == nil {
		return []*Thought{}
	}

	// 返回最近3个思考步骤
	count := 3
	if len(state.Thoughts) < count {
		count = len(state.Thoughts)
	}

	start := len(state.Thoughts) - count
	return state.Thoughts[start:]
}

// getRecentActions 获取最近的动极
// 【推荐】从状态中提取最近的动极步骤
func (cb *ContextBuilder) getRecentActions(state *ReActState) []*Action {
	if state.Actions == nil {
		return []*Action{}
	}

	// 返回最近3个动极步骤
	count := 3
	if len(state.Actions) < count {
		count = len(state.Actions)
	}

	start := len(state.Actions) - count
	return state.Actions[start:]
}

// getRecentObservations 获取最近的观察
// 【推荐】从状态中提取最近的观察结果
func (cb *ContextBuilder) getRecentObservations(state *ReActState) []*Observation {
	if state.Observations == nil {
		return []*Observation{}
	}

	// 返回最近3个观察结果
	count := 3
	if len(state.Observations) < count {
		count = len(state.Observations)
	}

	start := len(state.Observations) - count
	return state.Observations[start:]
}

// generateContextSummary 生成上下文摘要
// 【推荐】基于上下文信息生成简洁的摘要
func (cb *ContextBuilder) generateContextSummary(context *ReActContext) string {
	summary := fmt.Sprintf("Session %s, Iteration %d/%d", 
		context.SessionID, context.CurrentIteration, context.MaxIterations)

	if len(context.RecentThoughts) > 0 {
		summary += fmt.Sprintf(". Last thought: %s", 
			truncateString(context.RecentThoughts[0].Content, 50))
	}

	if len(context.RecentActions) > 0 {
		summary += fmt.Sprintf(". Last action: %s", context.RecentActions[0].Name)
	}

	if len(context.RecentObservations) > 0 {
		status := "success"
		if !context.RecentObservations[0].Success {
			status = "failed"
		}
		summary += fmt.Sprintf(". Last observation: %s", status)
	}

	if len(context.RelevantMemories) > 0 {
		summary += fmt.Sprintf(". %d relevant memories found", len(context.RelevantMemories))
	}

	return summary
}

// truncateString 截断字符串
// 【推荐】辅助函数，截断过长的字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MemoryManagerWrapper 内存管理器包装器
// 【必须】适配现有的memory.Manager接口
type MemoryManagerWrapper struct {
	// Integration 内存集成器
	Integration MemoryIntegration
	// logger 日志记录器
	logger *zap.Logger
}

// NewMemoryManagerWrapper 创建内存管理器包装器
// 【必须】提供适配器构造函数
func NewMemoryManagerWrapper(integration MemoryIntegration, logger *zap.Logger) *MemoryManagerWrapper {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &MemoryManagerWrapper{
		Integration: integration,
		logger:      logger.Named("memory_manager_wrapper"),
	}
}

// Store 存储内存条目（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Store方法
func (mmw *MemoryManagerWrapper) Store(ctx context.Context, item *memory.MemoryItem) error {
	// 这里需要根据item的类型分发到不同的存储方法
	// 这是一个简化的实现
	return mmw.Integration.StoreThought(ctx, &Thought{
		ID:        item.Metadata["thought_id"].(string),
		Content:   item.Metadata["content"].(string),
		Timestamp: item.Timestamp,
	}, &ReActState{SessionID: item.Tags["session_id"]})
}

// Retrieve 检索内存条目（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Retrieve方法
func (mmw *MemoryManagerWrapper) Retrieve(ctx context.Context, query memory.Query, limit int) ([]memory.MemoryItem, error) {
	// 转换为ReAct检索
	state := &ReActState{SessionID: query.Tags["session_id"]}
	return mmw.Integration.RetrieveRelevant(ctx, query.Text, limit, state)
}

// Delete 删除内存条目（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Delete方法
func (mmw *MemoryManagerWrapper) Delete(ctx context.Context, id string) error {
	// 简化实现：实际的删除逻辑应该在Integration中实现
	mmw.logger.Debug("memory item deletion requested", zap.String("item_id", id))
	return nil
}

// Update 更新内存条目（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Update方法
func (mmw *MemoryManagerWrapper) Update(ctx context.Context, item *memory.MemoryItem) error {
	// 简化实现：先删除再存储
	if err := mmw.Delete(ctx, item.ID); err != nil {
		return err
	}
	return mmw.Store(ctx, item)
}

// Clear 清空所有内存（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Clear方法
func (mmw *MemoryManagerWrapper) Clear(ctx context.Context) error {
	// 这个方法需要具体的session_id才能工作，所以这里记录警告
	mmw.logger.Warn("Clear called without session_id, operation ignored")
	return nil
}

// Search 搜索内存条目（实现memory.Manager接口）
// 【必须】适配memory.Manager接口的Search方法
func (mmw *MemoryManagerWrapper) Search(ctx context.Context, pattern string, options *memory.SearchOptions) ([]memory.MemoryItem, error) {
	// 转换为检索操作
	state := &ReActState{SessionID: options.Tags["session_id"]}
	return mmw.Integration.RetrieveRelevant(ctx, pattern, options.Limit, state)
}