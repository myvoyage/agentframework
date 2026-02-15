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
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"AgentFramework/pkg/beads/context"
)

// Deduplicator 记忆去重器实现
type Deduplicator struct {
	threshold float64 // 相似度阈值
}

// NewDeduplicator 创建新的去重器
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		threshold: 0.85, // 85% 相似度
	}
}

// Deduplicate 去重记忆
func (d *Deduplicator) Deduplicate(ctx context.Context, memories *context.MemoryCollection) (*context.MemoryCollection, error) {
	if memories == nil {
		return &context.MemoryCollection{}, nil
	}

	result := &context.MemoryCollection{}

	// 去重 Profile
	result.Profiles = d.deduplicateProfiles(memories.Profiles)

	// 去重 Preference
	result.Preferences = d.deduplicatePreferences(memories.Preferences)

	// 去重 Entity
	result.Entities = d.deduplicateEntities(memories.Entities)

	// 去重 Event（不去重，事件应该保留）
	result.Events = memories.Events

	// 去重 Case
	result.Cases = d.deduplicateCases(memories.Cases)

	// 去重 Pattern
	result.Patterns = d.deduplicatePatterns(memories.Patterns)

	return result, nil
}

// FindSimilar 查找相似记忆
func (d *Deduplicator) FindSimilar(ctx context.Context, memory interface{}) ([]interface{}, error) {
	// 这是一个占位实现
	// 实际使用时需要从存储中查询所有记忆进行比较
	return []interface{}{}, nil
}

// ShouldMerge 判断是否应该合并记忆
func (d *Deduplicator) ShouldMerge(ctx context.Context, existing, new interface{}) (bool, error) {
	// 计算相似度
	similarity := d.calculateSimilarity(existing, new)
	return similarity >= d.threshold, nil
}

// calculateSimilarity 计算两个记忆的相似度
func (d *Deduplicator) calculateSimilarity(a, b interface{}) float64 {
	// 根据类型计算相似度
	switch mem := a.(type) {
	case *context.ProfileMemory:
		if other, ok := b.(*context.ProfileMemory); ok {
			return d.profileSimilarity(mem, other)
		}
	case *context.PreferenceMemory:
		if other, ok := b.(*context.PreferenceMemory); ok {
			return d.preferenceSimilarity(mem, other)
		}
	case *context.EntityMemory:
		if other, ok := b.(*context.EntityMemory); ok {
			return d.entitySimilarity(mem, other)
		}
	case *context.EventMemory:
		if other, ok := b.(*context.EventMemory); ok {
			return d.eventSimilarity(mem, other)
		}
	case *context.CaseMemory:
		if other, ok := b.(*context.CaseMemory); ok {
			return d.caseSimilarity(mem, other)
		}
	case *context.PatternMemory:
		if other, ok := b.(*context.PatternMemory); ok {
			return d.patternSimilarity(mem, other)
		}
	}

	return 0.0
}

// ===== 去重方法 =====

// deduplicateProfiles 去重 Profile
func (d *Deduplicator) deduplicateProfiles(profiles []*context.ProfileMemory) []*context.ProfileMemory {
	if len(profiles) == 0 {
		return profiles
	}

	// 使用 ID 去重
	seen := make(map[string]bool)
	result := make([]*context.ProfileMemory, 0)

	for _, profile := range profiles {
		if !seen[profile.ID] {
			seen[profile.ID] = true
			result = append(result, profile)
		} else {
			// 合并相同 ID 的 Profile
			for i, existing := range result {
				if existing.ID == profile.ID {
					result[i] = d.mergeProfiles(existing, profile)
					break
				}
			}
		}
	}

	return result
}

// mergeProfiles 合并 Profile
func (d *Deduplicator) mergeProfiles(existing, new *context.ProfileMemory) *context.ProfileMemory {
	merged := &context.ProfileMemory{
		ID:        existing.ID,
		Name:      existing.Name,
		Role:      existing.Role,
		Traits:    make(map[string]string),
		Goals:     append([]string{}, existing.Goals...),
		UpdatedAt: time.Now(),
	}

	// 合并 Traits
	for k, v := range existing.Traits {
		merged.Traits[k] = v
	}
	for k, v := range new.Traits {
		merged.Traits[k] = v
	}

	// 合并 Goals（去重）
	goalSet := make(map[string]bool)
	for _, goal := range existing.Goals {
		goalSet[goal] = true
		merged.Goals = append(merged.Goals, goal)
	}
	for _, goal := range new.Goals {
		if !goalSet[goal] {
			merged.Goals = append(merged.Goals, goal)
		}
	}

	return merged
}

// profileSimilarity 计算 Profile 相似度
func (d *Deduplicator) profileSimilarity(a, b *context.ProfileMemory) float64 {
	if a.ID != b.ID {
		return 0.0
	}

	score := 0.0

	// 名字完全相同
	if a.Name == b.Name {
		score += 0.5
	}

	// 角色相同
	if a.Role == b.Role {
		score += 0.3
	}

	// Traits 相似度
	traitScore := d.mapSimilarity(a.Traits, b.Traits)
	score += traitScore * 0.2

	return score
}

// deduplicatePreferences 去重 Preference
func (d *Deduplicator) deduplicatePreferences(prefs []*context.PreferenceMemory) []*context.PreferenceMemory {
	if len(prefs) == 0 {
		return prefs
	}

	// 使用 Category+Key 组合去重
	seen := make(map[string]*context.PreferenceMemory)
	result := make([]*context.PreferenceMemory, 0)

	for _, pref := range prefs {
		key := pref.Category + ":" + pref.Key

		if existing, ok := seen[key]; ok {
			// 合并：选择置信度更高的
			if pref.Confidence > existing.Confidence {
				seen[key] = pref
			}
		} else {
			seen[key] = pref
			result = append(result, pref)
		}
	}

	return result
}

// preferenceSimilarity 计算 Preference 相似度
func (d *Deduplicator) preferenceSimilarity(a, b *context.PreferenceMemory) float64 {
	if a.Category != b.Category || a.Key != b.Key {
		return 0.0
	}

	if a.Value == b.Value {
		return 1.0
	}

	return 0.0
}

// deduplicateEntities 去重 Entity
func (d *Deduplicator) deduplicateEntities(entities []*context.EntityMemory) []*context.EntityMemory {
	if len(entities) == 0 {
		return entities
	}

	// 使用名称+类型去重
	seen := make(map[string]*context.EntityMemory)
	result := make([]*context.EntityMemory, 0)

	for _, entity := range entities {
		key := entity.Type + ":" + entity.Name

		if existing, ok := seen[key]; ok {
			// 合并 Entity
			merged := d.mergeEntities(existing, entity)
			seen[key] = merged
			// 更新结果中的引用
			for i, e := range result {
				if e == existing {
					result[i] = merged
					break
				}
			}
		} else {
			seen[key] = entity
			result = append(result, entity)
		}
	}

	return result
}

// mergeEntities 合并 Entity
func (d *Deduplicator) mergeEntities(existing, new *context.EntityMemory) *context.EntityMemory {
	merged := &context.EntityMemory{
		ID:         existing.ID,
		Type:       existing.Type,
		Name:       existing.Name,
		Attributes: make(map[string]string),
		Relations:  append([]context.EntityRelation{}, existing.Relations...),
		FirstSeen:  existing.FirstSeen,
		LastSeen:   new.LastSeen,
	}

	// 合并 Attributes
	for k, v := range existing.Attributes {
		merged.Attributes[k] = v
	}
	for k, v := range new.Attributes {
		merged.Attributes[k] = v
	}

	// 合并 Relations（去重）
	relationSet := make(map[string]bool)
	for _, rel := range existing.Relations {
		key := rel.Type + ":" + rel.EntityID
		if !relationSet[key] {
			relationSet[key] = true
			merged.Relations = append(merged.Relations, rel)
		}
	}
	for _, rel := range new.Relations {
		key := rel.Type + ":" + rel.EntityID
		if !relationSet[key] {
			relationSet[key] = true
			merged.Relations = append(merged.Relations, rel)
		}
	}

	return merged
}

// entitySimilarity 计算 Entity 相似度
func (d *Deduplicator) entitySimilarity(a, b *context.EntityMemory) float64 {
	if a.Type != b.Type {
		return 0.0
	}

	score := 0.0

	// 名称相同
	if a.Name == b.Name {
		score += 0.7
	}

	// Attributes 相似度
	attrScore := d.mapSimilarity(a.Attributes, b.Attributes)
	score += attrScore * 0.3

	return score
}

// deduplicateCases 去重 Case
func (d *Deduplicator) deduplicateCases(cases []*context.CaseMemory) []*context.CaseMemory {
	if len(cases) == 0 {
		return cases
	}

	// 使用问题内容的哈希去重
	seen := make(map[string]bool)
	result := make([]*context.CaseMemory, 0)

	for _, caseMem := range cases {
		hash := d.hashContent(caseMem.Problem)

		if !seen[hash] {
			seen[hash] = true
			result = append(result, caseMem)
		}
	}

	return result
}

// caseSimilarity 计算 Case 相似度
func (d *Deduplicator) caseSimilarity(a, b *context.CaseMemory) float64 {
	score := 0.0

	// 问题相似度
	problemSim := d.stringSimilarity(a.Problem, b.Problem)
	score += problemSim * 0.5

	// 解决方案相似度
	solutionSim := d.stringSimilarity(a.Solution, b.Solution)
	score += solutionSim * 0.3

	// 域相同
	if a.Domain == b.Domain {
		score += 0.2
	}

	return score
}

// deduplicatePatterns 去重 Pattern
func (d *Deduplicator) deduplicatePatterns(patterns []*context.PatternMemory) []*context.PatternMemory {
	if len(patterns) == 0 {
		return patterns
	}

	// 使用 Category+Pattern 去重
	seen := make(map[string]*context.PatternMemory)
	result := make([]*context.PatternMemory, 0)

	for _, pattern := range patterns {
		key := pattern.Category + ":" + pattern.Pattern

		if existing, ok := seen[key]; ok {
			// 合并：累加频率
			existing.Frequency += pattern.Frequency
			existing.LastSeen = pattern.LastSeen
			// 取更高置信度
			if pattern.Confidence > existing.Confidence {
				existing.Confidence = pattern.Confidence
			}
		} else {
			seen[key] = pattern
			result = append(result, pattern)
		}
	}

	return result
}

// patternSimilarity 计算 Pattern 相似度
func (d *Deduplicator) patternSimilarity(a, b *context.PatternMemory) float64 {
	if a.Category != b.Category {
		return 0.0
	}

	if a.Pattern == b.Pattern {
		return 1.0
	}

	return d.stringSimilarity(a.Pattern, b.Pattern)
}

// eventSimilarity 计算 Event 相似度
func (d *Deduplicator) eventSimilarity(a, b *context.EventMemory) float64 {
	score := 0.0

	// 类型相同
	if a.Type == b.Type {
		score += 0.3
	}

	// 标题相似度
	titleSim := d.stringSimilarity(a.Title, b.Title)
	score += titleSim * 0.4

	// 描述相似度
	descSim := d.stringSimilarity(a.Description, b.Description)
	score += descSim * 0.3

	return score
}

// ===== 辅助方法 =====

// stringSimilarity 计算字符串相似度（简化版）
func (d *Deduplicator) stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// 简单的 Jaccard 相似度
	setA := d.wordSet(a)
	setB := d.wordSet(b)

	intersection := 0
	for word := range setA {
		if setB[word] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// wordSet 将字符串转换为词集合
func (d *Deduplicator) wordSet(text string) map[string]bool {
	// 简单分词（按空白字符）
	words := splitWords(text)
	set := make(map[string]bool)

	for _, word := range words {
		if len(word) > 2 { // 忽略短词
			set[word] = true
		}
	}

	return set
}

// splitWords 分词
func splitWords(text string) []string {
	// 简单实现：按空白字符分割
	words := make([]string, 0)
	currentWord := ""

	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(r)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// mapSimilarity 计算两个 map 的相似度
func (d *Deduplicator) mapSimilarity(a, b map[string]string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	matches := 0
	for k, va := range a {
		if vb, ok := b[k]; ok {
			if va == vb {
				matches++
			}
		}
	}

	return float64(matches) / float64(len(a)+len(b)-matches)
}

// hashContent 计算内容哈希
func (d *Deduplicator) hashContent(content string) string {
	if content == "" {
		return ""
	}

	// 标准化内容
	content = normalizeContent(content)

	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}

// normalizeContent 标准化内容
func normalizeContent(content string) string {
	// 转小写
	content = strings.ToLower(content)

	// 移除多余空白
	content = joinStrings(strings.Fields(content))

	return content
}

// joinStrings 连接字符串切片
func joinStrings(strs []string) string {
	var result strings.Builder
	for i, s := range strs {
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(s)
	}
	return result.String()
}

// ===== MemoryDeduplicator 接口实现 =====

// MemoryDeduplicator 记忆去重器（接口适配）
type MemoryDeduplicator struct {
	deduplicator *Deduplicator
}

// NewMemoryDeduplicator 创建新的记忆去重器
func NewMemoryDeduplicator() *MemoryDeduplicator {
	return &MemoryDeduplicator{
		deduplicator: NewDeduplicator(),
	}
}

// Deduplicate 去重记忆
func (md *MemoryDeduplicator) Deduplicate(ctx context.Context, memories *context.MemoryCollection) (*context.MemoryCollection, error) {
	return md.deduplicator.Deduplicate(ctx, memories)
}

// FindSimilar 查找相似记忆
func (md *MemoryDeduplicator) FindSimilar(ctx context.Context, memory interface{}) ([]interface{}, error) {
	return md.deduplicator.FindSimilar(ctx, memory)
}

// ShouldMerge 判断是否应该合并记忆
func (md *MemoryDeduplicator) ShouldMerge(ctx context.Context, existing, new interface{}) (bool, error) {
	return md.deduplicator.ShouldMerge(ctx, existing, new)
}
