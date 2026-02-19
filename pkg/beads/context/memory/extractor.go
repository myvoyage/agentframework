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
	"fmt"
	"regexp"
	"strings"
	"time"

	"AgentFramework/pkg/beads/context"
)

// Extractor 记忆提取器实现
type Extractor struct {
	profileExtractor    *ProfileExtractor
	preferenceExtractor *PreferenceExtractor
	entityExtractor     *EntityExtractor
	eventExtractor      *EventExtractor
	caseExtractor       *CaseExtractor
	patternExtractor    *PatternExtractor
}

// NewExtractor 创建新的记忆提取器
func NewExtractor() *Extractor {
	return &Extractor{
		profileExtractor:    NewProfileExtractor(),
		preferenceExtractor: NewPreferenceExtractor(),
		entityExtractor:     NewEntityExtractor(),
		eventExtractor:      NewEventExtractor(),
		caseExtractor:       NewCaseExtractor(),
		patternExtractor:    NewPatternExtractor(),
	}
}

// ExtractFromContext 从上下文中提取记忆
func (e *Extractor) ExtractFromContext(ctx context.Context, ctxt *context.Context) (*context.MemoryCollection, error) {
	collection := &context.MemoryCollection{}

	// 从 L2 内容中提取
	if ctxt.Layers.L2 != nil {
		text := ctxt.Layers.L2.Content
		format := ctxt.Layers.L2.Format

		memories, err := e.ExtractFromText(ctx, text, format)
		if err == nil {
			collection = memories
		}
	}

	// 合并现有记忆
	if ctxt.Memories != nil {
		collection = e.mergeMemories(collection, ctxt.Memories)
	}

	return collection, nil
}

// ExtractFromConversation 从对话中提取记忆
func (e *Extractor) ExtractFromConversation(ctx context.Context, messages []context.Message) (*context.MemoryCollection, error) {
	// 合并所有消息内容
	var allText strings.Builder
	for _, msg := range messages {
		allText.WriteString(msg.Role)
		allText.WriteString(": ")
		allText.WriteString(msg.Content)
		allText.WriteString("\n\n")
	}

	// 从合并文本中提取
	return e.ExtractFromText(ctx, allText.String(), "conversation")
}

// ExtractFromText 从文本中提取记忆
func (e *Extractor) ExtractFromText(ctx context.Context, text string, contentType string) (*context.MemoryCollection, error) {
	collection := &context.MemoryCollection{}

	// 提取各种类型的记忆
	profiles := e.profileExtractor.Extract(text)
	collection.Profiles = append(collection.Profiles, profiles...)

	preferences := e.preferenceExtractor.Extract(text)
	collection.Preferences = append(collection.Preferences, preferences...)

	entities := e.entityExtractor.Extract(text)
	collection.Entities = append(collection.Entities, entities...)

	events := e.eventExtractor.Extract(text)
	collection.Events = append(collection.Events, events...)

	cases := e.caseExtractor.Extract(text)
	collection.Cases = append(collection.Cases, cases...)

	patterns := e.patternExtractor.Extract(text)
	collection.Patterns = append(collection.Patterns, patterns...)

	return collection, nil
}

// Deduplicate 去重记忆
func (e *Extractor) Deduplicate(ctx context.Context, memories *context.MemoryCollection) (*context.MemoryCollection, error) {
	deduplicator := NewDeduplicator()
	return deduplicator.Deduplicate(ctx, memories)
}

// Merge 合并记忆
func (e *Extractor) Merge(ctx context.Context, existing *context.MemoryCollection, new *context.MemoryCollection) (*context.MemoryCollection, error) {
	return e.mergeMemories(existing, new), nil
}

// mergeMemories 内部合并方法
func (e *Extractor) mergeMemories(existing *context.MemoryCollection, new *context.MemoryCollection) *context.MemoryCollection {
	result := &context.MemoryCollection{}

	// 合并 Profile
	result.Profiles = append(existing.Profiles, new.Profiles...)
	result.Profiles = deduplicateSlice(result.Profiles)

	// 合并 Preference
	result.Preferences = append(existing.Preferences, new.Preferences...)
	result.Preferences = deduplicateSlice(result.Preferences)

	// 合并 Entity
	result.Entities = append(existing.Entities, new.Entities...)
	result.Entities = deduplicateSlice(result.Entities)

	// 合并 Event
	result.Events = append(existing.Events, new.Events...)
	result.Events = deduplicateSlice(result.Events)

	// 合并 Case
	result.Cases = append(existing.Cases, new.Cases...)
	result.Cases = deduplicateSlice(result.Cases)

	// 合并 Pattern
	result.Patterns = append(existing.Patterns, new.Patterns...)
	result.Patterns = deduplicateSlice(result.Patterns)

	return result
}

// deduplicateSlice 对切片进行简单去重（独立函数，使用泛型）
func deduplicateSlice[T interface{ GetID() string }](items []T) []T {
	seen := make(map[string]bool)
	result := make([]T, 0)

	for _, item := range items {
		id := item.GetID()
		if !seen[id] {
			seen[id] = true
			result = append(result, item)
		}
	}

	return result
}

// ===== Profile 提取器 =====

// ProfileExtractor Profile 记忆提取器
type ProfileExtractor struct{}

// NewProfileExtractor 创建新的 Profile 提取器
func NewProfileExtractor() *ProfileExtractor {
	return &ProfileExtractor{}
}

// Extract 提取用户画像
func (e *ProfileExtractor) Extract(text string) []*context.ProfileMemory {
	var profiles []*context.ProfileMemory

	// 检测用户介绍模式
	patterns := []string{
		`我是(.{1,50})`,
		`我的名字是(.{1,30})`,
		`我叫(.{1,30})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) > 1 {
				name := strings.TrimSpace(match[1])
				if len(name) > 0 && len(name) < 50 {
					profiles = append(profiles, &context.ProfileMemory{
						ID:   e.generateID(name),
						Name: name,
						Role: e.detectRole(text),
						Traits: map[string]string{
							"source": "extracted",
						},
						UpdatedAt: time.Now(),
					})
				}
			}
		}
	}

	return profiles
}

// detectRole 检测角色
func (e *ProfileExtractor) detectRole(text string) string {
	roleKeywords := map[string]string{
		"开发者":     "developer",
		"程序员":     "programmer",
		"工程师":     "engineer",
		"设计师":     "designer",
		"产品经理":   "product manager",
		"学生":       "student",
		"老师":       "teacher",
	}

	textLower := strings.ToLower(text)
	for keyword, role := range roleKeywords {
		if strings.Contains(textLower, keyword) {
			return role
		}
	}

	return "user"
}

// generateID 生成记忆 ID
func (e *ProfileExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("profile:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法

// ===== Preference 提取器 =====

// PreferenceExtractor Preference 记忆提取器
type PreferenceExtractor struct{}

// NewPreferenceExtractor 创建新的 Preference 提取器
func NewPreferenceExtractor() *PreferenceExtractor {
	return &PreferenceExtractor{}
}

// Extract 提取用户偏好
func (e *PreferenceExtractor) Extract(text string) []*context.PreferenceMemory {
	var preferences []*context.PreferenceMemory

	// 检测偏好模式
	patterns := map[string]string{
		"我喜欢":     "like",
		"我偏好":     "preference",
		"我倾向":     "preference",
		"我喜欢用":    "tool",
		"我习惯":     "habit",
	}

	for keyword, category := range patterns {
		if strings.Contains(text, keyword) {
			// 提取偏好内容
			re := regexp.MustCompile(keyword + `(.{1,100})`)
			matches := re.FindAllStringSubmatch(text, -1)

			for _, match := range matches {
				if len(match) > 1 {
					value := strings.TrimSpace(match[1])
					if len(value) > 2 && len(value) < 100 {
						// 进一步提取键值对
						parts := strings.SplitN(value, "是", 2)
						key := value
						if len(parts) == 2 {
							key = strings.TrimSpace(parts[0])
							value = strings.TrimSpace(parts[1])
						}

						preferences = append(preferences, &context.PreferenceMemory{
							ID:        e.generateID(key + ":" + value),
							Category:  category,
							Key:       key,
							Value:     value,
							Confidence: 0.7,
							UpdatedAt: time.Now(),
						})
					}
				}
			}
		}
	}

	return preferences
}

// generateID 生成记忆 ID
func (e *PreferenceExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("preference:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法

// ===== Entity 提取器 =====

// EntityExtractor Entity 记忆提取器
type EntityExtractor struct{}

// NewEntityExtractor 创建新的 Entity 提取器
func NewEntityExtractor() *EntityExtractor {
	return &EntityExtractor{}
}

// Extract 提取实体
func (e *EntityExtractor) Extract(text string) []*context.EntityMemory {
	var entities []*context.EntityMemory

	// 检测人名（中文）
	reChineseName := regexp.MustCompile("[\\p{Han}]{2,4}")
	chineseNames := reChineseName.FindAllString(text, -1)

	// 去重
	seen := make(map[string]bool)
	for _, name := range chineseNames {
		if !seen[name] {
			seen[name] = true
			entities = append(entities, &context.EntityMemory{
				ID:     e.generateID(name),
				Type:   "person",
				Name:   name,
				Attributes: map[string]string{
					"source": "extracted",
				},
				FirstSeen: time.Now(),
				LastSeen:  time.Now(),
			})
		}
	}

	// 检测技术术语
	techTerms := e.extractTechTerms(text)
	for _, term := range techTerms {
		entities = append(entities, &context.EntityMemory{
			ID:     e.generateID(term),
			Type:   "technology",
			Name:   term,
			Attributes: map[string]string{
				"category": "tech",
			},
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		})
	}

	return entities
}

// extractTechTerms 提取技术术语
func (e *EntityExtractor) extractTechTerms(text string) []string {
	// 常见技术术语列表
	techTerms := []string{
		"Go", "Python", "JavaScript", "TypeScript", "Java",
		"React", "Vue", "Angular", "Node.js", "Docker",
		"Kubernetes", "Git", "Linux", "MySQL", "PostgreSQL",
		"MongoDB", "Redis", "API", "REST", "GraphQL",
		"AI", "机器学习", "深度学习", "神经网络",
	}

	var found []string
	textLower := strings.ToLower(text)

	for _, term := range techTerms {
		if strings.Contains(textLower, strings.ToLower(term)) {
			found = append(found, term)
		}
	}

	return found
}

// generateID 生成记忆 ID
func (e *EntityExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("entity:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法

// ===== Event 提取器 =====

// EventExtractor Event 记忆提取器
type EventExtractor struct{}

// NewEventExtractor 创建新的 Event 提取器
func NewEventExtractor() *EventExtractor {
	return &EventExtractor{}
}

// Extract 提取事件
func (e *EventExtractor) Extract(text string) []*context.EventMemory {
	var events []*context.EventMemory

	// 检测事件关键词
	eventKeywords := map[string]string{
		"决定":   "decision",
		"完成":   "milestone",
		"开始":   "start",
		"遇到":   "problem",
		"解决":   "solution",
		"发现":   "discovery",
		"学习":   "learning",
		"部署":   "deployment",
	}

	for keyword, eventType := range eventKeywords {
		if strings.Contains(text, keyword) {
			// 提取事件描述
			re := regexp.MustCompile(`.{0,200}` + keyword + `.{0,200}`)
			matches := re.FindAllString(text, -1)

			for _, match := range matches {
				match = strings.TrimSpace(match)
				if len(match) > 10 {
					events = append(events, &context.EventMemory{
						ID:          e.generateID(match),
						Type:        eventType,
						Title:       e.extractTitle(keyword, match),
						Description: match,
						OccurredAt:  time.Now(),
					})
				}
			}
		}
	}

	return events
}

// extractTitle 提取事件标题
func (e *EventExtractor) extractTitle(keyword, text string) string {
	// 简单提取：取前50个字符
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

// generateID 生成记忆 ID
func (e *EventExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("event:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法

// ===== Case 提取器 =====

// CaseExtractor Case 记忆提取器
type CaseExtractor struct{}

// NewCaseExtractor 创建新的 Case 提取器
func NewCaseExtractor() *CaseExtractor {
	return &CaseExtractor{}
}

// Extract 提取案例
func (e *CaseExtractor) Extract(text string) []*context.CaseMemory {
	var cases []*context.CaseMemory

	// 检测问题-解决模式
	problemPatterns := []string{
		`问题.{0,100}[:：].{0,500}[解解]决.{0,500}`,
		`错误.{0,100}[:：].{0,500}[修修]复.{0,500}`,
		`bug.{0,100}[:：].{0,500}[修修]复.{0,500}`,
	}

	for _, pattern := range problemPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)

		for _, match := range matches {
			match = strings.TrimSpace(match)
			if len(match) > 50 {
				// 分离问题和解决方案
				parts := e.splitProblemSolution(match)
				if parts != nil {
					cases = append(cases, &context.CaseMemory{
						ID:       e.generateID(match),
						Domain:   e.detectDomain(text),
						Problem:  parts[0],
						Solution: parts[1],
						Lessons:  e.extractLessons(match),
						Tags:     e.extractTags(match),
						CreatedAt: time.Now(),
					})
				}
			}
		}
	}

	return cases
}

// splitProblemSolution 分离问题和解决方案
func (e *CaseExtractor) splitProblemSolution(text string) []string {
	// 尝试常见的分隔符
	separators := []string{"解决", "修复", "fix", "solution", "：", ":"}

	for _, sep := range separators {
		if strings.Contains(text, sep) {
			parts := strings.SplitN(text, sep, 2)
			if len(parts) == 2 {
				return []string{
					strings.TrimSpace(parts[0]),
					strings.TrimSpace(parts[1]),
				}
			}
		}
	}

	return nil
}

// detectDomain 检测领域
func (e *CaseExtractor) detectDomain(text string) string {
	domainKeywords := map[string]string{
		"编程":   "programming",
		"调试":   "debugging",
		"设计":   "design",
		"部署":   "deployment",
		"bug":    "programming",
		"error":  "programming",
		"debug":  "debugging",
		"design": "design",
	}

	textLower := strings.ToLower(text)
	for keyword, domain := range domainKeywords {
		if strings.Contains(textLower, keyword) {
			return domain
		}
	}

	return "general"
}

// extractLessons 提取经验教训
func (e *CaseExtractor) extractLessons(text string) []string {
	var lessons []string

	lessonKeywords := []string{"教训", "学到的", "lesson", "learned"}

	for _, keyword := range lessonKeywords {
		if strings.Contains(text, keyword) {
			// 提取教训内容
			re := regexp.MustCompile(keyword + `.{0,200}`)
			matches := re.FindAllString(text, -1)

			for _, match := range matches {
				match = strings.TrimSpace(match)
				if len(match) > 10 {
					lessons = append(lessons, match)
				}
			}
		}
	}

	return lessons
}

// extractTags 提取标签
func (e *CaseExtractor) extractTags(text string) []string {
	// 简单实现：提取关键词
	tags := []string{"solved", "case"}

	if strings.Contains(strings.ToLower(text), "bug") {
		tags = append(tags, "bug")
	}
	if strings.Contains(strings.ToLower(text), "error") {
		tags = append(tags, "error")
	}

	return tags
}

// generateID 生成记忆 ID
func (e *CaseExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("case:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法

// ===== Pattern 提取器 =====

// PatternExtractor Pattern 记忆提取器
type PatternExtractor struct{}

// NewPatternExtractor 创建新的 Pattern 提取器
func NewPatternExtractor() *PatternExtractor {
	return &PatternExtractor{}
}

// Extract 提取模式
func (e *PatternExtractor) Extract(text string) []*context.PatternMemory {
	var patterns []*context.PatternMemory

	// 检测重复模式
	repeatedPatterns := e.detectRepeatedPatterns(text)
	for _, pattern := range repeatedPatterns {
		patterns = append(patterns, &context.PatternMemory{
			ID:        e.generateID(pattern),
			Category:  "repetition",
			Pattern:   pattern,
			Frequency: e.countPattern(text, pattern),
			LastSeen:  time.Now(),
			Confidence: 0.6,
		})
	}

	// 检测代码模式
	codePatterns := e.detectCodePatterns(text)
	patterns = append(patterns, codePatterns...)

	return patterns
}

// detectRepeatedPatterns 检测重复模式
func (e *PatternExtractor) detectRepeatedPatterns(text string) []string {
	var patterns []string

	// 简单实现：检测重复的句子
	sentences := strings.Split(text, "。")
	counts := make(map[string]int)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 && len(sentence) < 100 {
			counts[sentence]++
		}
	}

	// 选择出现多次的句子
	for sentence, count := range counts {
		if count >= 2 {
			patterns = append(patterns, fmt.Sprintf("重复模式(%d次): %s", count, sentence))
		}
	}

	return patterns
}

// detectCodePatterns 检测代码模式
func (e *PatternExtractor) detectCodePatterns(text string) []*context.PatternMemory {
	var patterns []*context.PatternMemory

	// 检测常见代码模式
	codePatterns := []struct {
		pattern  string
		category string
	}{
		{"func ", "function"},
		{"if ", "condition"},
		{"for ", "loop"},
		{"return ", "return"},
		{"class ", "class"},
	}

	for _, cp := range codePatterns {
		count := strings.Count(text, cp.pattern)
		if count > 0 {
			patterns = append(patterns, &context.PatternMemory{
				ID:        e.generateID(cp.pattern),
				Category:  cp.category,
				Pattern:   cp.pattern,
				Frequency: count,
				LastSeen:  time.Now(),
				Confidence: 0.8,
			})
		}
	}

	return patterns
}

// countPattern 计算模式出现次数
func (e *PatternExtractor) countPattern(text, pattern string) int {
	return strings.Count(text, pattern)
}

// generateID 生成记忆 ID
func (e *PatternExtractor) generateID(content string) string {
	hash := sha256.Sum256([]byte("pattern:" + content))
	return hex.EncodeToString(hash[:])[:16]
}

// GetID 实现 GetID 方法
