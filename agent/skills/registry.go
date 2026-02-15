// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// SkillRegistry 技能注册表
// 提供技能注册、查询、去重和导出功能
// 对应 TRAE 的 api_description.md 功能
type SkillRegistry struct {
	skills  map[string]*SkillEntry
	mu      sync.RWMutex
	indexes map[string][]string // 索引：按类型、标签等
	baseDir string              // 注册表文件存储目录
}

// SkillEntry 技能条目（对应 api_description.md）
type SkillEntry struct {
	// 基本信息
	ID          string   `json:"id"`          // 唯一标识
	Name        string   `json:"name"`        // 技能名称
	Description string   `json:"description"` // 技能描述
	Category    string   `json:"category"`    // 技能分类
	Tags        []string `json:"tags"`        // 技能标签
	Version     string   `json:"version"`     // 版本号
	Enabled     bool     `json:"enabled"`     // 是否启用

	// 执行信息
	CreatedAt  time.Time `json:"created_at"`   // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`   // 更新时间
	UsedCount  int64     `json:"used_count"`   // 使用次数
	LastUsed   time.Time `json:"last_used"`    // 最后使用时间
	LastUsedBy string    `json:"last_used_by"` // 最后使用者

	// 参数定义
	InputSchema  *Schema `json:"input_schema"`  // 输入参数Schema
	OutputSchema *Schema `json:"output_schema"` // 输出参数Schema

	// 生成位置
	GeneratedFile string `json:"generated_file"` // 生成的文件路径
	GeneratedLine int    `json:"generated_line"` // 生成的行号

	// 配置信息
	Config map[string]interface{} `json:"config"` // 配置项

	// 元数据
	Metadata map[string]interface{} `json:"metadata"` // 扩展元数据
}

// Schema 参数Schema定义
type Schema struct {
	Type       string                   `json:"type"`                 // 类型: object, array, string, number, boolean
	Properties map[string]*PropertyInfo `json:"properties,omitempty"` // 属性定义（object类型）
	Items      *PropertyInfo            `json:"items,omitempty"`      // 数组元素定义（array类型）
	Required   []string                 `json:"required,omitempty"`   // 必填字段列表
	Enum       []interface{}            `json:"enum,omitempty"`       // 枚举值
	Default    interface{}              `json:"default,omitempty"`    // 默认值
}

// PropertyInfo 属性信息
type PropertyInfo struct {
	Type        string      `json:"type"`                // 属性类型
	Description string      `json:"description"`         // 属性描述
	Required    bool        `json:"required,omitempty"`  // 是否必填
	Enum        []string    `json:"enum,omitempty"`      // 枚举值
	Default     interface{} `json:"default,omitempty"`   // 默认值
	MinLength   *int        `json:"minLength,omitempty"` // 最小长度
	MaxLength   *int        `json:"maxLength,omitempty"` // 最大长度
	Minimum     *float64    `json:"minimum,omitempty"`   // 最小值
	Maximum     *float64    `json:"maximum,omitempty"`   // 最大值
	Pattern     string      `json:"pattern,omitempty"`   // 正则表达式
	Format      string      `json:"format,omitempty"`    // 格式（如: date-time, email, uri）
	Example     interface{} `json:"example,omitempty"`   // 示例值
}

// SkillQuery 技能查询条件
type SkillQuery struct {
	ID            string     // 按ID查询
	Name          string     // 按名称查询（支持模糊匹配）
	Category      string     // 按分类查询
	Tags          []string   // 按标签查询（AND逻辑）
	CreatedAfter  *time.Time // 创建时间之后
	CreatedBefore *time.Time // 创建时间之前
	UsedCountMin  int64      // 最小使用次数
	UsedCountMax  int64      // 最大使用次数
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	BaseDir          string        // 注册表文件存储目录
	AutoSave         bool          // 是否自动保存
	AutoSaveInterval time.Duration // 自动保存间隔
	EnableIndex      bool          // 是否启用索引
}

// NewSkillRegistry 创建新的技能注册表
func NewSkillRegistry(config *RegistryConfig) *SkillRegistry {
	if config == nil {
		config = &RegistryConfig{
			BaseDir:     ".skills/registry",
			AutoSave:    true,
			EnableIndex: true,
		}
	}

	// 确保目录存在
	if config.BaseDir != "" {
		os.MkdirAll(config.BaseDir, 0755)
	}

	registry := &SkillRegistry{
		skills:  make(map[string]*SkillEntry),
		indexes: make(map[string][]string),
		baseDir: config.BaseDir,
	}

	// 尝试从文件加载
	if config.BaseDir != "" {
		registry.loadFromFile()
	}

	return registry
}

// Register 注册技能（自动去重）
// 返回错误如果技能已存在
func (r *SkillRegistry) Register(entry *SkillEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 验证条目
	if err := r.validateEntry(entry); err != nil {
		return fmt.Errorf("invalid entry: %w", err)
	}

	// 检查是否已存在
	if _, exists := r.skills[entry.ID]; exists {
		return fmt.Errorf("skill %s already registered (created at %s)",
			entry.ID, r.skills[entry.ID].CreatedAt.Format("2006-01-02 15:04:05"))
	}

	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	// 设置默认值
	if entry.Version == "" {
		entry.Version = "1.0.0"
	}
	if entry.Category == "" {
		entry.Category = "general"
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}

	r.skills[entry.ID] = entry

	// 更新索引
	if r.indexes != nil {
		r.indexes[entry.Category] = append(r.indexes[entry.Category], entry.ID)
		for _, tag := range entry.Tags {
			r.indexes[tag] = append(r.indexes[tag], entry.ID)
		}
	}

	// 自动保存
	if r.baseDir != "" {
		if err := r.saveToFile(); err != nil {
			// 保存失败，撤销注册
			delete(r.skills, entry.ID)
			return fmt.Errorf("failed to save registry: %w", err)
		}
		fmt.Printf("[SkillRegistry] 已保存技能注册表到: %s, 总技能数: %d\n", r.baseDir, len(r.skills))
	}

	return nil
}

// Update 更新技能条目
func (r *SkillRegistry) Update(entry *SkillEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 验证条目
	if err := r.validateEntry(entry); err != nil {
		return fmt.Errorf("invalid entry: %w", err)
	}

	// 检查是否存在
	if _, exists := r.skills[entry.ID]; !exists {
		return fmt.Errorf("skill %s not found", entry.ID)
	}

	entry.UpdatedAt = time.Now()

	// 保留创建时间和使用统计
	oldEntry := r.skills[entry.ID]
	entry.CreatedAt = oldEntry.CreatedAt
	entry.UsedCount = oldEntry.UsedCount
	entry.LastUsed = oldEntry.LastUsed

	r.skills[entry.ID] = entry

	// 自动保存
	if r.baseDir != "" {
		r.saveToFile()
	}

	return nil
}

// Unregister 注销技能
func (r *SkillRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("skill %s not found", id)
	}

	// 从索引中删除
	if r.indexes != nil {
		// 删除分类索引
		if ids, ok := r.indexes[entry.Category]; ok {
			r.indexes[entry.Category] = removeString(ids, id)
		}
		// 删除标签索引
		for _, tag := range entry.Tags {
			if ids, ok := r.indexes[tag]; ok {
				r.indexes[tag] = removeString(ids, id)
			}
		}
	}

	delete(r.skills, id)

	// 自动保存
	if r.baseDir != "" {
		r.saveToFile()
	}

	return nil
}

// GetByID 根据ID获取技能
func (r *SkillRegistry) GetByID(id string) (*SkillEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.skills[id]
	return entry, exists
}

// Exists 检查技能是否存在（去重检查）
func (r *SkillRegistry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.skills[id]
	return exists
}

// RecordUsage 记录技能使用
func (r *SkillRegistry) RecordUsage(id string, user string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("skill %s not found", id)
	}

	entry.UsedCount++
	entry.LastUsed = time.Now()
	entry.LastUsedBy = user

	// 自动保存
	if r.baseDir != "" {
		r.saveToFile()
	}

	return nil
}

// Find 查找技能（支持多种查询条件）
func (r *SkillRegistry) Find(query *SkillQuery) ([]*SkillEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*SkillEntry

	for _, entry := range r.skills {
		if r.matchQuery(entry, query) {
			results = append(results, entry)
		}
	}

	// 按更新时间排序（最新的在前）
	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})

	return results, nil
}

// ListByCategory 按分类列出技能
func (r *SkillRegistry) ListByCategory(category string) []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.indexes != nil {
		// 使用索引
		ids := r.indexes[category]
		entries := make([]*SkillEntry, 0, len(ids))

		for _, id := range ids {
			if entry, exists := r.skills[id]; exists {
				entries = append(entries, entry)
			}
		}

		return entries
	}

	// 不使用索引
	var entries []*SkillEntry
	for _, entry := range r.skills {
		if entry.Category == category {
			entries = append(entries, entry)
		}
	}

	return entries
}

// ListByTag 按标签列出技能
func (r *SkillRegistry) ListByTag(tag string) []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.indexes != nil {
		// 使用索引
		ids := r.indexes[tag]
		entries := make([]*SkillEntry, 0, len(ids))

		for _, id := range ids {
			if entry, exists := r.skills[id]; exists {
				entries = append(entries, entry)
			}
		}

		return entries
	}

	// 不使用索引
	var entries []*SkillEntry
	for _, entry := range r.skills {
		for _, t := range entry.Tags {
			if t == tag {
				entries = append(entries, entry)
				break
			}
		}
	}

	return entries
}

// ListAll 列出所有技能
func (r *SkillRegistry) ListAll() []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*SkillEntry, 0, len(r.skills))
	for _, entry := range r.skills {
		entries = append(entries, entry)
	}

	// 按分类和名称排序
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// GetCategories 获取所有分类
func (r *SkillRegistry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make(map[string]bool)
	for _, entry := range r.skills {
		categories[entry.Category] = true
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}

	sort.Strings(result)
	return result
}

// GetTags 获取所有标签
func (r *SkillRegistry) GetTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make(map[string]bool)
	for _, entry := range r.skills {
		for _, tag := range entry.Tags {
			tags[tag] = true
		}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}

	sort.Strings(result)
	return result
}

// GetStats 获取统计信息
func (r *SkillRegistry) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]interface{}{
		"total_skills":  len(r.skills),
		"total_uses":    int64(0),
		"categories":    len(r.GetCategories()),
		"tags":          len(r.GetTags()),
		"most_used":     (*SkillEntry)(nil),
		"recently_used": []*SkillEntry{},
		"never_used":    int64(0),
	}

	var mostUsed *SkillEntry
	var totalUses int64
	var neverUsed int64
	var recentlyUsed []*SkillEntry

	oneDayAgo := time.Now().Add(-24 * time.Hour)

	for _, entry := range r.skills {
		totalUses += entry.UsedCount

		if mostUsed == nil || entry.UsedCount > mostUsed.UsedCount {
			mostUsed = entry
		}

		if entry.UsedCount == 0 {
			neverUsed++
		}

		if !entry.LastUsed.IsZero() && entry.LastUsed.After(oneDayAgo) {
			recentlyUsed = append(recentlyUsed, entry)
		}
	}

	stats["total_uses"] = totalUses
	stats["most_used"] = mostUsed
	stats["recently_used"] = recentlyUsed
	stats["never_used"] = neverUsed

	return stats
}

// ExportToJSON 导出到JSON格式
func (r *SkillRegistry) ExportToJSON(w io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(r.skills)
}

// ExportToMarkdown 导出到Markdown格式（知识沉淀）
func (r *SkillRegistry) ExportToMarkdown(w io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fmt.Fprintf(w, "# Skill Registry\n\n")
	fmt.Fprintf(w, "**生成时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "**技能总数**: %d\n\n", len(r.skills))

	// 统计信息
	stats := r.GetStats()
	fmt.Fprintf(w, "## 统计信息\n\n")
	fmt.Fprintf(w, "- **总技能数**: %d\n", stats["total_skills"])
	fmt.Fprintf(w, "- **总使用次数**: %d\n", stats["total_uses"])
	fmt.Fprintf(w, "- **分类数**: %d\n", stats["categories"])
	fmt.Fprintf(w, "- **标签数**: %d\n", stats["tags"])
	fmt.Fprintf(w, "- **未使用技能**: %d\n\n", stats["never_used"])

	if mostUsed, ok := stats["most_used"].(*SkillEntry); ok && mostUsed != nil {
		fmt.Fprintf(w, "### 最常用技能\n\n")
		fmt.Fprintf(w, "- **%s** (%s): %d 次使用\n", mostUsed.Name, mostUsed.ID, mostUsed.UsedCount)
		fmt.Fprintf(w, "\n")
	}

	// 按分类组织
	categories := r.GetCategories()

	for _, category := range categories {
		fmt.Fprintf(w, "## %s\n\n", strings.ToUpper(category))

		entries := r.ListByCategory(category)
		for _, entry := range entries {
			fmt.Fprintf(w, "### %s\n\n", entry.Name)
			fmt.Fprintf(w, "**ID**: `%s`\n\n", entry.ID)
			fmt.Fprintf(w, "**描述**: %s\n\n", entry.Description)
			fmt.Fprintf(w, "**版本**: %s\n\n", entry.Version)

			if len(entry.Tags) > 0 {
				fmt.Fprintf(w, "**标签**: %s\n\n", strings.Join(entry.Tags, ", "))
			}

			if entry.InputSchema != nil {
				fmt.Fprintf(w, "#### 输入参数\n\n")
				r.exportSchema(w, entry.InputSchema)
			}

			if entry.OutputSchema != nil {
				fmt.Fprintf(w, "#### 输出参数\n\n")
				r.exportSchema(w, entry.OutputSchema)
			}

			if entry.GeneratedFile != "" {
				fmt.Fprintf(w, "**生成位置**: `%s:%d`\n\n", entry.GeneratedFile, entry.GeneratedLine)
			}

			if entry.UsedCount > 0 {
				fmt.Fprintf(w, "**使用次数**: %d\n", entry.UsedCount)
				fmt.Fprintf(w, "**最后使用**: %s\n", entry.LastUsed.Format("2006-01-02 15:04:05"))
				if entry.LastUsedBy != "" {
					fmt.Fprintf(w, "**最后使用者**: %s", entry.LastUsedBy)
				}
				fmt.Fprintf(w, "\n")
			}

			fmt.Fprintf(w, "---\n\n")
		}
	}

	return nil
}

// ExportToEinoSchema 导出为 Eino ToolInfo 格式
func (r *SkillRegistry) ExportToEinoSchema(id string) (*schema.ToolInfo, error) {
	entry, exists := r.GetByID(id)
	if !exists {
		return nil, fmt.Errorf("skill %s not found", id)
	}

	// 构建参数信息
	params := make(map[string]*schema.ParameterInfo)
	if entry.InputSchema != nil && entry.InputSchema.Properties != nil {
		for name, prop := range entry.InputSchema.Properties {
			params[name] = &schema.ParameterInfo{
				Type:     schema.DataType(prop.Type),
				Desc:     prop.Description,
				Required: contains(entry.InputSchema.Required, name),
			}
		}
	}

	return &schema.ToolInfo{
		Name:        entry.ID,
		Desc:        entry.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

// SaveToFile 保存到文件
func (r *SkillRegistry) SaveToFile() error {
	if r.baseDir == "" {
		return fmt.Errorf("base directory not set")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.saveToFile()
}

// LoadFromFile 从文件加载
func (r *SkillRegistry) LoadFromFile() error {
	if r.baseDir == "" {
		return fmt.Errorf("base directory not set")
	}

	return r.loadFromFile()
}

// Clear 清空注册表
func (r *SkillRegistry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.skills = make(map[string]*SkillEntry)
	if r.indexes != nil {
		r.indexes = make(map[string][]string)
	}

	if r.baseDir != "" {
		return r.saveToFile()
	}

	return nil
}

// validateEntry 验证技能条目
func (r *SkillRegistry) validateEntry(entry *SkillEntry) error {
	if entry.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if entry.Name == "" {
		return fmt.Errorf("name is required")
	}
	if entry.Description == "" {
		return fmt.Errorf("description is required")
	}

	// 验证ID格式（只允许字母、数字、下划线、中划线）
	if !isValidID(entry.ID) {
		return fmt.Errorf("invalid ID format: %s (only letters, numbers, underscore and hyphen allowed)", entry.ID)
	}

	return nil
}

// matchQuery 匹配查询条件
func (r *SkillRegistry) matchQuery(entry *SkillEntry, query *SkillQuery) bool {
	if query == nil {
		return true
	}

	// ID匹配
	if query.ID != "" && entry.ID != query.ID {
		return false
	}

	// 名称匹配（支持模糊匹配）
	if query.Name != "" && !containsFuzzy(entry.Name, query.Name) {
		return false
	}

	// 分类匹配
	if query.Category != "" && entry.Category != query.Category {
		return false
	}

	// 标签匹配（AND逻辑）
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			if !contains(entry.Tags, tag) {
				return false
			}
		}
	}

	// 创建时间范围
	if query.CreatedAfter != nil && entry.CreatedAt.Before(*query.CreatedAfter) {
		return false
	}
	if query.CreatedBefore != nil && entry.CreatedAt.After(*query.CreatedBefore) {
		return false
	}

	// 使用次数范围
	if query.UsedCountMin > 0 && entry.UsedCount < query.UsedCountMin {
		return false
	}
	if query.UsedCountMax > 0 && entry.UsedCount > query.UsedCountMax {
		return false
	}

	return true
}

// exportSchema 导出Schema到Markdown
func (r *SkillRegistry) exportSchema(w io.Writer, schema *Schema) {
	if schema == nil {
		return
	}

	fmt.Fprintf(w, "**类型**: %s\n\n", schema.Type)

	if schema.Properties != nil && len(schema.Properties) > 0 {
		fmt.Fprintf(w, "| 参数 | 类型 | 必填 | 描述 |\n")
		fmt.Fprintf(w, "|------|------|------|------|\n")

		// 按必填字段排序
		requiredFields := schema.Required
		var propNames []string
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Slice(propNames, func(i, j int) bool {
			iRequired := contains(requiredFields, propNames[i])
			jRequired := contains(requiredFields, propNames[j])
			if iRequired != jRequired {
				return iRequired
			}
			return propNames[i] < propNames[j]
		})

		for _, name := range propNames {
			prop := schema.Properties[name]
			required := ""
			if contains(requiredFields, name) {
				required = "是"
			}

			desc := prop.Description
			if desc == "" {
				desc = "-"
			}

			fmt.Fprintf(w, "| %s | %s | %s | %s |\n",
				name, prop.Type, required, desc)
		}

		fmt.Fprintf(w, "\n")
	}

	if schema.Default != nil {
		fmt.Fprintf(w, "**默认值**: `%v`\n\n", schema.Default)
	}

	if len(schema.Enum) > 0 {
		fmt.Fprintf(w, "**枚举值**: ")
		for i, val := range schema.Enum {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "`%v`", val)
		}
		fmt.Fprintf(w, "\n\n")
	}
}

// saveToFile 保存到文件
func (r *SkillRegistry) saveToFile() error {
	if r.baseDir == "" {
		return nil
	}

	// 保存JSON格式
	jsonFile := filepath.Join(r.baseDir, "registry.json")
	file, err := os.Create(jsonFile)
	if err != nil {
		return fmt.Errorf("create registry file failed: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(r.skills); err != nil {
		return fmt.Errorf("encode registry failed: %w", err)
	}

	// 保存Markdown格式
	markdownFile := filepath.Join(r.baseDir, "registry.md")
	mdFile, err := os.Create(markdownFile)
	if err != nil {
		return fmt.Errorf("create markdown file failed: %w", err)
	}
	defer mdFile.Close()

	if err := r.ExportToMarkdown(mdFile); err != nil {
		return fmt.Errorf("export markdown failed: %w", err)
	}

	return nil
}

// loadFromFile 从文件加载
func (r *SkillRegistry) loadFromFile() error {
	if r.baseDir == "" {
		return nil
	}

	jsonFile := filepath.Join(r.baseDir, "registry.json")
	file, err := os.Open(jsonFile)
	if err != nil {
		// 文件不存在是正常情况
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open registry file failed: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	skills := make(map[string]*SkillEntry)

	if err := decoder.Decode(&skills); err != nil {
		return fmt.Errorf("decode registry failed: %w", err)
	}

	// 重建索引
	for id, entry := range skills {
		r.skills[id] = entry

		if r.indexes != nil {
			r.indexes[entry.Category] = append(r.indexes[entry.Category], id)
			for _, tag := range entry.Tags {
				r.indexes[tag] = append(r.indexes[tag], id)
			}
		}
	}

	return nil
}

// isValidID 验证ID格式
func isValidID(id string) bool {
	if id == "" {
		return false
	}

	for _, c := range id {
		if !isAlphaNumeric(c) && c != '_' && c != '-' {
			return false
		}
	}

	return true
}

func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// contains 检查字符串切片是否包含某个字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsFuzzy 模糊匹配
func containsFuzzy(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// removeString 从字符串切片中移除指定字符串
func removeString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// ===== 启用/禁用功能 =====

// EnableSkill 启用技能
func (r *SkillRegistry) EnableSkill(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("skill not found: %s", id)
	}

	if entry.Enabled {
		return nil // 已经启用
	}

	entry.Enabled = true
	entry.UpdatedAt = time.Now()

	// 保存到文件
	if r.baseDir != "" {
		return r.SaveToFile()
	}

	return nil
}

// DisableSkill 禁用技能
func (r *SkillRegistry) DisableSkill(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.skills[id]
	if !exists {
		return fmt.Errorf("skill not found: %s", id)
	}

	if !entry.Enabled {
		return nil // 已经禁用
	}

	entry.Enabled = false
	entry.UpdatedAt = time.Now()

	// 保存到文件
	if r.baseDir != "" {
		return r.SaveToFile()
	}

	return nil
}

// ToggleSkill 切换技能启用状态
func (r *SkillRegistry) ToggleSkill(id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.skills[id]
	if !exists {
		return false, fmt.Errorf("skill not found: %s", id)
	}

	entry.Enabled = !entry.Enabled
	entry.UpdatedAt = time.Now()

	// 保存到文件
	if r.baseDir != "" {
		if err := r.SaveToFile(); err != nil {
			// 回滚
			entry.Enabled = !entry.Enabled
			entry.UpdatedAt = time.Now()
			return false, err
		}
	}

	return entry.Enabled, nil
}

// IsEnabled 检查技能是否启用
func (r *SkillRegistry) IsEnabled(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.skills[id]
	if !exists {
		return false, fmt.Errorf("skill not found: %s", id)
	}

	return entry.Enabled, nil
}

// ListEnabled 列出所有启用的技能
func (r *SkillRegistry) ListEnabled() []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*SkillEntry, 0, len(r.skills))
	for _, entry := range r.skills {
		if entry.Enabled {
			result = append(result, entry)
		}
	}
	return result
}

// ListDisabled 列出所有禁用的技能
func (r *SkillRegistry) ListDisabled() []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*SkillEntry, 0, len(r.skills))
	for _, entry := range r.skills {
		if !entry.Enabled {
			result = append(result, entry)
		}
	}
	return result
}
