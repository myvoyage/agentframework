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

package memory

import (
	stdctx "context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"AgentFramework/agent"
	beadscontext "AgentFramework/pkg/beads/context"

	"github.com/cloudwego/eino/schema"
)

// LLMCompressor 基于 LLM 的记忆压缩器
// 使用大语言模型进行智能记忆压缩、精华提取和重要性评分
type LLMCompressor struct {
	model  agent.ChatModel
	config *beadscontext.MemoryCompressionConfig
	mu     sync.RWMutex
}

// NewLLMCompressor 创建新的 LLM 压缩器
func NewLLMCompressor(model agent.ChatModel, config *beadscontext.MemoryCompressionConfig) *LLMCompressor {
	if config == nil {
		config = beadscontext.DefaultMemoryCompressionConfig()
	}
	return &LLMCompressor{
		model:  model,
		config: config,
	}
}

// CompressMemories 压缩记忆集合到指定层级
// 将大量记忆压缩为少量精华记忆
func (c *LLMCompressor) CompressMemories(ctx stdctx.Context, memories *beadscontext.MemoryCollection, tier beadscontext.MemoryTier) (*beadscontext.MemoryCollection, error) {
	if memories == nil || memories.GetMemoryCount() == 0 {
		return NewMemoryCollection(), nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 根据目标层级确定压缩策略
	var targetCount int
	switch tier {
	case beadscontext.MemoryTierSession:
		targetCount = c.config.MaxSessionMemories
	case beadscontext.MemoryTierDaily:
		targetCount = c.config.MaxDailyMemories
	case beadscontext.MemoryTierLongTerm:
		targetCount = c.config.MaxLongTermMemories
	default:
		targetCount = c.config.MaxDailyMemories
	}

	// 如果记忆数量已经低于目标数量，无需压缩
	if memories.GetMemoryCount() <= targetCount {
		return memories, nil
	}

	// 使用异步压缩以避免阻塞主流程
	if c.config.EnableAsyncCompression {
		go func() {
			// 使用新的 context 以避免原 context 被取消
			asyncCtx := stdctx.Background()
			c.compressAsync(asyncCtx, memories, tier, targetCount)
		}()
		return memories, nil
	}

	// 同步压缩
	return c.compressSync(ctx, memories, tier, targetCount)
}

// compressSync 同步压缩记忆
func (c *LLMCompressor) compressSync(ctx stdctx.Context, memories *beadscontext.MemoryCollection, tier beadscontext.MemoryTier, targetCount int) (*beadscontext.MemoryCollection, error) {
	// 提取精华记忆
	essentials, err := c.ExtractEssentials(ctx, memories, targetCount)
	if err != nil {
		return nil, fmt.Errorf("failed to extract essentials: %w", err)
	}

	return essentials, nil
}

// compressAsync 异步压缩记忆
func (c *LLMCompressor) compressAsync(ctx stdctx.Context, memories *beadscontext.MemoryCollection, tier beadscontext.MemoryTier, targetCount int) {
	essentials, err := c.ExtractEssentials(ctx, memories, targetCount)
	if err != nil {
		// 记录错误但不阻塞
		fmt.Printf("async compression failed: %v\n", err)
		return
	}

	// 在实际应用中，这里应该更新存储
	_ = essentials
}

// SummarizeByType 按类型压缩记忆
func (c *LLMCompressor) SummarizeByType(ctx stdctx.Context, memories *beadscontext.MemoryCollection, memoryType beadscontext.MemoryType) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch memoryType {
	case beadscontext.MemoryTypeProfile:
		return c.summarizeProfiles(ctx, memories.Profiles)
	case beadscontext.MemoryTypePreference:
		return c.summarizePreferences(ctx, memories.Preferences)
	case beadscontext.MemoryTypeEntity:
		return c.summarizeEntities(ctx, memories.Entities)
	case beadscontext.MemoryTypeEvent:
		return c.summarizeEvents(ctx, memories.Events)
	case beadscontext.MemoryTypeCase:
		return c.summarizeCases(ctx, memories.Cases)
	case beadscontext.MemoryTypePattern:
		return c.summarizePatterns(ctx, memories.Patterns)
	default:
		return nil, fmt.Errorf("unknown memory type: %s", memoryType)
	}
}

// ExtractEssentials 提取最重要的记忆
func (c *LLMCompressor) ExtractEssentials(ctx stdctx.Context, memories *beadscontext.MemoryCollection, maxCount int) (*beadscontext.MemoryCollection, error) {
	if memories == nil || memories.GetMemoryCount() == 0 {
		return NewMemoryCollection(), nil
	}

	result := NewMemoryCollection()

	// 为每类记忆评分并提取最重要的
	if len(memories.Profiles) > 0 {
		profiles, err := c.extractTopProfiles(ctx, memories.Profiles, maxCount/6)
		if err == nil {
			result.Profiles = profiles
		}
	}

	if len(memories.Preferences) > 0 {
		prefs, err := c.extractTopPreferences(ctx, memories.Preferences, maxCount/6)
		if err == nil {
			result.Preferences = prefs
		}
	}

	if len(memories.Entities) > 0 {
		entities, err := c.extractTopEntities(ctx, memories.Entities, maxCount/6)
		if err == nil {
			result.Entities = entities
		}
	}

	if len(memories.Events) > 0 {
		events, err := c.extractTopEvents(ctx, memories.Events, maxCount/6)
		if err == nil {
			result.Events = events
		}
	}

	if len(memories.Cases) > 0 {
		cases, err := c.extractTopCases(ctx, memories.Cases, maxCount/6)
		if err == nil {
			result.Cases = cases
		}
	}

	if len(memories.Patterns) > 0 {
		patterns, err := c.extractTopPatterns(ctx, memories.Patterns, maxCount/6)
		if err == nil {
			result.Patterns = patterns
		}
	}

	return result, nil
}

// CalculateImportance 计算记忆重要性
func (c *LLMCompressor) CalculateImportance(ctx stdctx.Context, memory interface{}) (float64, error) {
	// 构建评分 prompt
	prompt := c.buildImportancePrompt(memory)

	msgs := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are an AI assistant that scores the importance of memories on a scale of 0 to 1. " +
				"Consider factors like recency, relevance, uniqueness, and impact. " +
				"Respond ONLY with a JSON object containing a single 'score' field with a number between 0 and 1.",
		},
		{
			Role:    schema.User,
			Content: prompt,
		},
	}

	resp, err := c.model.Generate(ctx, msgs)
	if err != nil {
		// 如果 LLM 调用失败，返回默认分数
		return 0.5, nil
	}

	// 解析响应
	var result struct {
		Score float64 `json:"score"`
	}

	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 解析失败，返回默认分数
		return 0.5, nil
	}

	return result.Score, nil
}

// MergeMemories 合并记忆
func (c *LLMCompressor) MergeMemories(ctx stdctx.Context, base, delta *beadscontext.MemoryCollection) (*beadscontext.MemoryCollection, error) {
	if base == nil {
		return delta, nil
	}
	if delta == nil {
		return base, nil
	}

	result := NewMemoryCollection()

	// 合并各类记忆
	result.Profiles = c.mergeProfiles(base.Profiles, delta.Profiles)
	result.Preferences = c.mergePreferences(base.Preferences, delta.Preferences)
	result.Entities = c.mergeEntities(base.Entities, delta.Entities)
	result.Events = c.mergeEvents(base.Events, delta.Events)
	result.Cases = c.mergeCases(base.Cases, delta.Cases)
	result.Patterns = c.mergePatterns(base.Patterns, delta.Patterns)

	return result, nil
}

// ===== 评分辅助方法 =====

func (c *LLMCompressor) calculateImportanceBatch(ctx stdctx.Context, memories []interface{}) ([]float64, error) {
	scores := make([]float64, len(memories))

	// 使用简单的启发式评分，避免过多的 LLM 调用
	for i, mem := range memories {
		score, err := c.CalculateImportance(ctx, mem)
		if err != nil {
			scores[i] = 0.5 // 默认分数
		} else {
			scores[i] = score
		}
	}

	return scores, nil
}

// ===== 按类型提取精华记忆 =====

func (c *LLMCompressor) extractTopProfiles(ctx stdctx.Context, profiles []*beadscontext.ProfileMemory, count int) ([]*beadscontext.ProfileMemory, error) {
	if len(profiles) <= count {
		return profiles, nil
	}

	// 按更新时间排序，保留最新的
	return c.topProfilesByRecency(profiles, count), nil
}

func (c *LLMCompressor) extractTopPreferences(ctx stdctx.Context, prefs []*beadscontext.PreferenceMemory, count int) ([]*beadscontext.PreferenceMemory, error) {
	if len(prefs) <= count {
		return prefs, nil
	}

	// 按置信度排序
	return c.topPreferencesByConfidence(prefs, count), nil
}

func (c *LLMCompressor) extractTopEntities(ctx stdctx.Context, entities []*beadscontext.EntityMemory, count int) ([]*beadscontext.EntityMemory, error) {
	if len(entities) <= count {
		return entities, nil
	}

	// 按最近发现时间排序
	return c.topEntitiesByRecency(entities, count), nil
}

func (c *LLMCompressor) extractTopEvents(ctx stdctx.Context, events []*beadscontext.EventMemory, count int) ([]*beadscontext.EventMemory, error) {
	if len(events) <= count {
		return events, nil
	}

	// 按发生时间排序，保留最近的
	return c.topEventsByRecency(events, count), nil
}

func (c *LLMCompressor) extractTopCases(ctx stdctx.Context, cases []*beadscontext.CaseMemory, count int) ([]*beadscontext.CaseMemory, error) {
	if len(cases) <= count {
		return cases, nil
	}

	// 按应用次数排序，保留最常用的
	return c.topCasesByUsage(cases, count), nil
}

func (c *LLMCompressor) extractTopPatterns(ctx stdctx.Context, patterns []*beadscontext.PatternMemory, count int) ([]*beadscontext.PatternMemory, error) {
	if len(patterns) <= count {
		return patterns, nil
	}

	// 按频率和置信度排序
	return c.topPatternsByFrequency(patterns, count), nil
}

// ===== 按类型摘要 =====

func (c *LLMCompressor) summarizeProfiles(ctx stdctx.Context, profiles []*beadscontext.ProfileMemory) (interface{}, error) {
	if len(profiles) == 0 {
		return nil, nil
	}

	// 合并多个 profile 为一个
	merged := &beadscontext.ProfileMemory{
		Name:   profiles[0].Name,
		Role:   profiles[0].Role,
		Traits: make(map[string]string),
		Goals:  make([]string, 0),
	}

	for _, p := range profiles {
		for k, v := range p.Traits {
			merged.Traits[k] = v
		}
		merged.Goals = append(merged.Goals, p.Goals...)
	}

	return merged, nil
}

func (c *LLMCompressor) summarizePreferences(ctx stdctx.Context, prefs []*beadscontext.PreferenceMemory) (interface{}, error) {
	if len(prefs) == 0 {
		return nil, nil
	}

	// 按类别分组
	byCategory := make(map[string][]*beadscontext.PreferenceMemory)
	for _, p := range prefs {
		byCategory[p.Category] = append(byCategory[p.Category], p)
	}

	// 每个类别保留置信度最高的
	result := make([]*beadscontext.PreferenceMemory, 0)
	for _, categoryPrefs := range byCategory {
		top := categoryPrefs[0]
		for _, p := range categoryPrefs {
			if p.Confidence > top.Confidence {
				top = p
			}
		}
		result = append(result, top)
	}

	return result, nil
}

func (c *LLMCompressor) summarizeEntities(ctx stdctx.Context, entities []*beadscontext.EntityMemory) (interface{}, error) {
	if len(entities) == 0 {
		return nil, nil
	}

	// 去重：相同名称和类型的实体只保留一个
	unique := make(map[string]*beadscontext.EntityMemory)
	for _, e := range entities {
		key := fmt.Sprintf("%s:%s", e.Type, e.Name)
		if existing, ok := unique[key]; ok {
			// 合并属性和关系
			for k, v := range e.Attributes {
				existing.Attributes[k] = v
			}
			existing.Relations = append(existing.Relations, e.Relations...)
		} else {
			unique[key] = e
		}
	}

	result := make([]*beadscontext.EntityMemory, 0, len(unique))
	for _, e := range unique {
		result = append(result, e)
	}

	return result, nil
}

func (c *LLMCompressor) summarizeEvents(ctx stdctx.Context, events []*beadscontext.EventMemory) (interface{}, error) {
	if len(events) == 0 {
		return nil, nil
	}

	// 使用 LLM 生成事件摘要
	prompt := c.buildEventSummaryPrompt(events)
	msgs := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are an AI assistant that summarizes event history. " +
				"Generate a concise summary of the following events. " +
				"Focus on key decisions, milestones, and outcomes. " +
				"Respond with a JSON object containing 'title', 'description', and 'key_outcomes' fields.",
		},
		{
			Role:    schema.User,
			Content: prompt,
		},
	}

	resp, err := c.model.Generate(ctx, msgs)
	if err != nil {
		// 失败时返回原始事件
		return events, nil
	}

	var summary struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		KeyOutcomes []string `json:"key_outcomes"`
	}

	if err := json.Unmarshal([]byte(resp.Content), &summary); err != nil {
		return events, nil
	}

	// 创建摘要事件
	return &beadscontext.EventMemory{
		Type:        string(beadscontext.MemoryTypeEvent),
		Title:       summary.Title,
		Description: summary.Description,
		Outcomes:    summary.KeyOutcomes,
		OccurredAt:  events[len(events)-1].OccurredAt,
	}, nil
}

func (c *LLMCompressor) summarizeCases(ctx stdctx.Context, cases []*beadscontext.CaseMemory) (interface{}, error) {
	if len(cases) == 0 {
		return nil, nil
	}

	// 按应用次数排序，返回最常用的
	return c.topCasesByUsage(cases, min(len(cases), 10)), nil
}

func (c *LLMCompressor) summarizePatterns(ctx stdctx.Context, patterns []*beadscontext.PatternMemory) (interface{}, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	// 按频率排序
	return c.topPatternsByFrequency(patterns, min(len(patterns), 10)), nil
}

// ===== 合并方法 =====

func (c *LLMCompressor) mergeProfiles(base, delta []*beadscontext.ProfileMemory) []*beadscontext.ProfileMemory {
	result := make([]*beadscontext.ProfileMemory, 0)
	seen := make(map[string]bool)

	// 添加 base
	for _, p := range base {
		result = append(result, p)
		seen[p.ID] = true
	}

	// 添加 delta 中新的
	for _, p := range delta {
		if !seen[p.ID] {
			result = append(result, p)
		}
	}

	return result
}

func (c *LLMCompressor) mergePreferences(base, delta []*beadscontext.PreferenceMemory) []*beadscontext.PreferenceMemory {
	result := make([]*beadscontext.PreferenceMemory, 0)
	byKey := make(map[string]*beadscontext.PreferenceMemory)

	// 合并 base
	for _, p := range base {
		key := fmt.Sprintf("%s:%s", p.Category, p.Key)
		byKey[key] = p
	}

	// 更新或添加 delta
	for _, p := range delta {
		key := fmt.Sprintf("%s:%s", p.Category, p.Key)
		if existing, ok := byKey[key]; ok {
			// 更新置信度更高的
			if p.Confidence > existing.Confidence {
				byKey[key] = p
			}
		} else {
			byKey[key] = p
		}
	}

	for _, p := range byKey {
		result = append(result, p)
	}

	return result
}

func (c *LLMCompressor) mergeEntities(base, delta []*beadscontext.EntityMemory) []*beadscontext.EntityMemory {
	result := make([]*beadscontext.EntityMemory, 0)
	byKey := make(map[string]*beadscontext.EntityMemory)

	for _, e := range base {
		key := fmt.Sprintf("%s:%s", e.Type, e.Name)
		byKey[key] = e
	}

	for _, e := range delta {
		key := fmt.Sprintf("%s:%s", e.Type, e.Name)
		if existing, ok := byKey[key]; ok {
			// 合并属性和关系
			for k, v := range e.Attributes {
				existing.Attributes[k] = v
			}
			existing.Relations = append(existing.Relations, e.Relations...)
		} else {
			byKey[key] = e
		}
	}

	for _, e := range byKey {
		result = append(result, e)
	}

	return result
}

func (c *LLMCompressor) mergeEvents(base, delta []*beadscontext.EventMemory) []*beadscontext.EventMemory {
	result := append(base, delta...)
	return result
}

func (c *LLMCompressor) mergeCases(base, delta []*beadscontext.CaseMemory) []*beadscontext.CaseMemory {
	result := make([]*beadscontext.CaseMemory, 0)
	byDomain := make(map[string]*beadscontext.CaseMemory)

	for _, c := range base {
		byDomain[c.Domain] = c
	}

	for _, c := range delta {
		if existing, ok := byDomain[c.Domain]; ok {
			// 合并相同领域的案例
			existing.Lessons = append(existing.Lessons, c.Lessons...)
		} else {
			byDomain[c.Domain] = c
		}
	}

	for _, c := range byDomain {
		result = append(result, c)
	}

	return result
}

func (c *LLMCompressor) mergePatterns(base, delta []*beadscontext.PatternMemory) []*beadscontext.PatternMemory {
	result := append(base, delta...)
	return result
}

// ===== Prompt 构建方法 =====

func (c *LLMCompressor) buildImportancePrompt(memory interface{}) string {
	data, _ := json.MarshalIndent(memory, "", "  ")
	return fmt.Sprintf("Please score the importance of this memory on a scale of 0 to 1:\n\n%s", string(data))
}

func (c *LLMCompressor) buildEventSummaryPrompt(events []*beadscontext.EventMemory) string {
	var sb strings.Builder
	sb.WriteString("Please summarize these events:\n\n")
	for i, e := range events {
		sb.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, e.Title, e.Description))
	}
	return sb.String()
}

// ===== 排序辅助方法 =====

func (c *LLMCompressor) topProfilesByRecency(profiles []*beadscontext.ProfileMemory, count int) []*beadscontext.ProfileMemory {
	sorted := make([]*beadscontext.ProfileMemory, len(profiles))
	copy(sorted, profiles)

	// 简单冒泡排序，按更新时间降序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].UpdatedAt.Before(sorted[j+1].UpdatedAt) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

func (c *LLMCompressor) topPreferencesByConfidence(prefs []*beadscontext.PreferenceMemory, count int) []*beadscontext.PreferenceMemory {
	sorted := make([]*beadscontext.PreferenceMemory, len(prefs))
	copy(sorted, prefs)

	// 按置信度降序排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Confidence < sorted[j+1].Confidence {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

func (c *LLMCompressor) topEntitiesByRecency(entities []*beadscontext.EntityMemory, count int) []*beadscontext.EntityMemory {
	sorted := make([]*beadscontext.EntityMemory, len(entities))
	copy(sorted, entities)

	// 按最近发现时间降序排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].LastSeen.Before(sorted[j+1].LastSeen) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

func (c *LLMCompressor) topEventsByRecency(events []*beadscontext.EventMemory, count int) []*beadscontext.EventMemory {
	sorted := make([]*beadscontext.EventMemory, len(events))
	copy(sorted, events)

	// 按发生时间降序排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].OccurredAt.Before(sorted[j+1].OccurredAt) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

func (c *LLMCompressor) topCasesByUsage(cases []*beadscontext.CaseMemory, count int) []*beadscontext.CaseMemory {
	sorted := make([]*beadscontext.CaseMemory, len(cases))
	copy(sorted, cases)

	// 按应用次数降序排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].AppliedCount < sorted[j+1].AppliedCount {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

func (c *LLMCompressor) topPatternsByFrequency(patterns []*beadscontext.PatternMemory, count int) []*beadscontext.PatternMemory {
	sorted := make([]*beadscontext.PatternMemory, len(patterns))
	copy(sorted, patterns)

	// 按频率和置信度排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			score1 := float64(sorted[j].Frequency) * sorted[j].Confidence
			score2 := float64(sorted[j+1].Frequency) * sorted[j+1].Confidence
			if score1 < score2 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	if len(sorted) > count {
		sorted = sorted[:count]
	}

	return sorted
}

// ===== 辅助函数 =====

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewMemoryCollection 创建新的记忆集合
func NewMemoryCollection() *beadscontext.MemoryCollection {
	return &beadscontext.MemoryCollection{
		Profiles:    make([]*beadscontext.ProfileMemory, 0),
		Preferences: make([]*beadscontext.PreferenceMemory, 0),
		Entities:    make([]*beadscontext.EntityMemory, 0),
		Events:      make([]*beadscontext.EventMemory, 0),
		Cases:       make([]*beadscontext.CaseMemory, 0),
		Patterns:    make([]*beadscontext.PatternMemory, 0),
	}
}
