// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProgressiveLoader 渐进式加载器
// 实现按需加载和渐进式披露机制，减少 Token 消耗
type ProgressiveLoader struct {
	basePath    string
	cache       map[string]*SkillDefinition
	metadata    map[string]*SkillMetadata // 元数据快速缓存
	mu          sync.RWMutex
	strategy    LoadStrategy
	preloadList []string // 预加载列表
}

// LoadStrategy 加载策略
type LoadStrategy int

const (
	// LoadOnDemand 按需加载（默认，推荐）
	LoadOnDemand LoadStrategy = iota
	// LoadEager 预加载（适合常用技能）
	LoadEager
	// LoadLazy 懒加载（节省内存）
	LoadLazy
)

// SkillMetadata 技能元数据（轻量级）
type SkillMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	SourceFile  string    `json:"source_file"`
	LoadedAt    time.Time `json:"loaded_at"`
	FileSize    int64     `json:"file_size"`
}

// NewProgressiveLoader 创建渐进式加载器
func NewProgressiveLoader(basePath string) *ProgressiveLoader {
	if basePath == "" {
		basePath = "agent/skills/definitions"
	}

	// 确保目录存在
	os.MkdirAll(basePath, 0755)

	loader := &ProgressiveLoader{
		basePath: basePath,
		cache:    make(map[string]*SkillDefinition),
		metadata: make(map[string]*SkillMetadata),
		strategy: LoadOnDemand,
	}

	// 预加载元数据
	loader.loadAllMetadata()

	return loader
}

// LoadSkill 加载完整的技能定义（按需）
func (l *ProgressiveLoader) LoadSkill(skillID string) (*SkillDefinition, error) {
	l.mu.RLock()
	if def, exists := l.cache[skillID]; exists {
		l.mu.RUnlock()
		return def, nil
	}
	l.mu.RUnlock()

	// 按需加载完整定义
	l.mu.Lock()
	defer l.mu.Unlock()

	// 双重检查
	if def, exists := l.cache[skillID]; exists {
		return def, nil
	}

	// 从文件加载完整定义
	def, err := l.loadSkillFromFile(skillID)
	if err != nil {
		return nil, err
	}

	l.cache[skillID] = def
	return def, nil
}

// LoadMetadata 只加载元数据（快速，不加载完整定义）
// 这是渐进式披露的核心：只加载基本信息，不加载详细的工作流
func (l *ProgressiveLoader) LoadMetadata(skillID string) (*SkillMetadata, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 优先从元数据缓存获取
	if meta, exists := l.metadata[skillID]; exists {
		return meta, nil
	}

	return nil, fmt.Errorf("skill metadata not found: %s", skillID)
}

// ListSkills 列出所有技能（只返回元数据，不加载完整定义）
// 这是渐进式披露的关键：快速浏览所有技能而不消耗大量 Token
func (l *ProgressiveLoader) ListSkills() ([]*SkillMetadata, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	metas := make([]*SkillMetadata, 0, len(l.metadata))
	for _, meta := range l.metadata {
		metas = append(metas, meta)
	}

	// 按分类和名称排序
	sortMetadata(metas)

	return metas, nil
}

// ListByCategory 按分类列出技能（只返回元数据）
func (l *ProgressiveLoader) ListByCategory(category string) ([]*SkillMetadata, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var metas []*SkillMetadata
	for _, meta := range l.metadata {
		if meta.Category == category {
			metas = append(metas, meta)
		}
	}

	return metas, nil
}

// Search 搜索技能（基于元数据）
func (l *ProgressiveLoader) Search(keyword string) ([]*SkillMetadata, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keyword = strings.ToLower(keyword)
	var results []*SkillMetadata

	for _, meta := range l.metadata {
		if strings.Contains(strings.ToLower(meta.Name), keyword) ||
			strings.Contains(strings.ToLower(meta.Description), keyword) ||
			containsAny(meta.Tags, keyword) {
			results = append(results, meta)
		}
	}

	return results, nil
}

// Warmup 预热（预加载常用技能）
// 对于高频使用的技能，可以提前加载完整定义
func (l *ProgressiveLoader) Warmup(skillIDs []string) error {
	if len(skillIDs) == 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error
	for _, id := range skillIDs {
		if _, err := l.loadSkillFromFile(id); err != nil {
			errs = append(errs, fmt.Errorf("warmup %s failed: %w", id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("warmup completed with %d errors: %v", len(errs), errs)
	}

	return nil
}

// GetStrategy 获取加载策略
func (l *ProgressiveLoader) GetStrategy() LoadStrategy {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.strategy
}

// SetStrategy 设置加载策略
func (l *ProgressiveLoader) SetStrategy(strategy LoadStrategy) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.strategy = strategy
}

// Preload 预加载列表管理
func (l *ProgressiveLoader) AddToPreload(skillIDs []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.preloadList = append(l.preloadList, skillIDs...)
}

// GetPreloadList 获取预加载列表
func (l *ProgressiveLoader) GetPreloadList() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.preloadList
}

// Reload 重新加载所有元数据
func (l *ProgressiveLoader) Reload() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 清空缓存
	l.cache = make(map[string]*SkillDefinition)
	l.metadata = make(map[string]*SkillMetadata)

	// 重新加载
	return l.loadAllMetadata()
}

// ClearCache 清除完整定义缓存（保留元数据）
func (l *ProgressiveLoader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache = make(map[string]*SkillDefinition)
}

// ClearAll 清除所有缓存（包括元数据）
func (l *ProgressiveLoader) ClearAll() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache = make(map[string]*SkillDefinition)
	l.metadata = make(map[string]*SkillMetadata)
}

// GetStats 获取加载统计
func (l *ProgressiveLoader) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := map[string]interface{}{
		"total_skills":       len(l.metadata),
		"cached_definitions": len(l.cache),
		"strategy":           l.strategy.String(),
		"preload_count":      len(l.preloadList),
		"cache_hit_rate":     l.calculateCacheHitRate(),
	}

	return stats
}

// loadAllMetadata 加载所有技能的元数据
func (l *ProgressiveLoader) loadAllMetadata() error {
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillID := entry.Name()
		if err := l.loadMetadataOnly(skillID); err != nil {
			// 记录错误但继续加载其他技能
			fmt.Printf("Warning: failed to load metadata for %s: %v\n", skillID, err)
		}
	}

	return nil
}

// loadMetadataOnly 只加载元数据（快速扫描）
// 核心优化：只读取文件的前几行，不解析完整内容
func (l *ProgressiveLoader) loadMetadataOnly(skillID string) error {
	skillDir := filepath.Join(l.basePath, skillID)

	// 查找 SKILL.yaml 或 SKILL.json
	defFile := filepath.Join(skillDir, "SKILL.yaml")
	if _, err := os.Stat(defFile); os.IsNotExist(err) {
		defFile = filepath.Join(skillDir, "SKILL.json")
		if _, err := os.Stat(defFile); os.IsNotExist(err) {
			return fmt.Errorf("skill definition file not found")
		}
	}

	// 获取文件信息
	info, err := os.Stat(defFile)
	if err != nil {
		return err
	}

	// 读取文件内容（完整读取以便解析）
	data, err := os.ReadFile(defFile)
	if err != nil {
		return err
	}

	// 解析元数据（不需要完整解析）
	var metadata SkillMetadata
	ext := filepath.Ext(defFile)

	if ext == ".yaml" || ext == ".yml" {
		if err := parseYAMLMetadata(data, &metadata); err != nil {
			return err
		}
	} else {
		if err := parseJSONMetadata(data, &metadata); err != nil {
			return err
		}
	}

	// 补充信息
	metadata.ID = skillID
	metadata.SourceFile = defFile
	metadata.LoadedAt = time.Now()
	metadata.FileSize = info.Size()

	l.metadata[skillID] = &metadata

	return nil
}

// loadSkillFromFile 从文件加载完整定义
func (l *ProgressiveLoader) loadSkillFromFile(skillID string) (*SkillDefinition, error) {
	skillDir := filepath.Join(l.basePath, skillID)

	// 查找 SKILL.yaml 或 SKILL.json
	defFile := filepath.Join(skillDir, "SKILL.yaml")
	if _, err := os.Stat(defFile); os.IsNotExist(err) {
		defFile = filepath.Join(skillDir, "SKILL.json")
		if _, err := os.Stat(defFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("skill definition file not found: %s", skillID)
		}
	}

	data, err := os.ReadFile(defFile)
	if err != nil {
		return nil, err
	}

	var definition SkillDefinition
	ext := filepath.Ext(defFile)

	if ext == ".yaml" || ext == ".yml" {
		if err := yamlUnmarshal(data, &definition); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &definition); err != nil {
			return nil, err
		}
	}

	definition.ID = skillID
	definition.SourceFile = defFile
	definition.LoadedAt = time.Now()

	return &definition, nil
}

// calculateCacheHitRate 计算缓存命中率
func (l *ProgressiveLoader) calculateCacheHitRate() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.metadata) == 0 {
		return 0
	}

	return float64(len(l.cache)) / float64(len(l.metadata)) * 100
}

// sortMetadata 排序元数据
func sortMetadata(metas []*SkillMetadata) {
	sort.Slice(metas, func(i, j int) bool {
		if metas[i].Category != metas[j].Category {
			return metas[i].Category < metas[j].Category
		}
		return metas[i].Name < metas[j].Name
	})
}

// containsAny 检查是否包含任意字符串
func containsAny(slice []string, target string) bool {
	target = strings.ToLower(target)
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), target) {
			return true
		}
	}
	return false
}

// String 返回策略的字符串表示
func (s LoadStrategy) String() string {
	switch s {
	case LoadOnDemand:
		return "on_demand"
	case LoadEager:
		return "eager"
	case LoadLazy:
		return "lazy"
	default:
		return "unknown"
	}
}

// parseYAMLMetadata 解析 YAML 元数据
func parseYAMLMetadata(data []byte, metadata *SkillMetadata) error {
	// 简化实现：只解析基本的 YAML 结构
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "name:") {
			metadata.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "description:") {
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		} else if strings.HasPrefix(line, "category:") {
			metadata.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		} else if strings.HasPrefix(line, "version:") {
			metadata.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		} else if strings.HasPrefix(line, "author:") {
			metadata.Author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
		} else if strings.HasPrefix(line, "tags:") {
			// 简化处理：提取标签
			if strings.Contains(line, "[") {
				// YAML 数组格式
				start := strings.Index(line, "[")
				end := strings.Index(line, "]")
				if start > 0 && end > start {
					tagsStr := strings.TrimSpace(line[start+1 : end])
					tags := strings.Split(tagsStr, ",")
					for i, tag := range tags {
						tags[i] = strings.TrimSpace(strings.Trim(tag, `"'`))
					}
					metadata.Tags = tags
				}
			}
		}
	}

	return nil
}

// parseJSONMetadata 解析 JSON 元数据
func parseJSONMetadata(data []byte, metadata *SkillMetadata) error {
	return json.Unmarshal(data, metadata)
}
