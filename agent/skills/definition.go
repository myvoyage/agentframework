// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SkillDefinition 技能定义（对应 TRAE 的 SKILL.md）
// 提供声明式的流程定义能力
type SkillDefinition struct {
	// 基本信息
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
	Category    string `json:"category" yaml:"category"`
	Author      string `json:"author" yaml:"author"`
	License     string `json:"license" yaml:"license"`

	// 前置条件
	Prerequisites []Prerequisite `json:"prerequisites" yaml:"prerequisites"`

	// 使用时机
	Triggers []string `json:"triggers" yaml:"triggers"`

	// 执行流程
	Workflow []WorkflowStep `json:"workflow" yaml:"workflow"`

	// 配置
	Config SkillConfig `json:"config" yaml:"config"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`

	// 加载信息
	SourceFile string    `json:"-" yaml:"-"`
	LoadedAt   time.Time `json:"-" yaml:"-"`
}

// Prerequisite 前置条件定义
type Prerequisite struct {
	Type        string            `json:"type" yaml:"type"` // "file_exists", "env_var", "dependency", "command", "custom"
	Description string            `json:"description" yaml:"description"`
	Check       string            `json:"check" yaml:"check"`             // 检查逻辑
	AutoFix     bool              `json:"auto_fix" yaml:"auto_fix"`       // 是否自动修复
	FixCommand  string            `json:"fix_command" yaml:"fix_command"` // 修复命令
	Parameters  map[string]string `json:"parameters" yaml:"parameters"`   // 参数
	Required    bool              `json:"required" yaml:"required"`       // 是否必须满足
}

// WorkflowStep 工作流步骤定义
type WorkflowStep struct {
	ID             string            `json:"id" yaml:"id"`
	Name           string            `json:"name" yaml:"name"`
	Description    string            `json:"description" yaml:"description"`
	Action         string            `json:"action" yaml:"action"` // "validate", "prepare", "execute", "cleanup", "check_exists", "generate_code"
	Parameters     map[string]string `json:"parameters" yaml:"parameters"`
	NextStep       string            `json:"next_step,omitempty" yaml:"next_step,omitempty"`
	SkipIf         string            `json:"skip_if,omitempty" yaml:"skip_if,omitempty"` // 跳过条件
	RetryOnFailure bool              `json:"retry_on_failure,omitempty" yaml:"retry_on_failure,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// DefinitionManager 技能定义管理器
type DefinitionManager struct {
	definitions map[string]*SkillDefinition
	mu          sync.RWMutex
	baseDir     string
	loaders     []DefinitionLoader
}

// DefinitionLoader 定义加载器接口
type DefinitionLoader interface {
	Load(filePath string) (*SkillDefinition, error)
	Save(filePath string, definition *SkillDefinition) error
	CanLoad(filePath string) bool
}

// NewDefinitionManager 创建定义管理器
func NewDefinitionManager(baseDir string) *DefinitionManager {
	if baseDir == "" {
		baseDir = "agent/skills/definitions"
	}

	// 确保目录存在
	os.MkdirAll(baseDir, 0755)

	manager := &DefinitionManager{
		definitions: make(map[string]*SkillDefinition),
		baseDir:     baseDir,
	}

	// 注册默认加载器
	manager.RegisterLoader(&YAMLLoader{})
	manager.RegisterLoader(&JSONLoader{})

	// 加载所有定义
	manager.LoadAll()

	return manager
}

// RegisterLoader 注册加载器
func (m *DefinitionManager) RegisterLoader(loader DefinitionLoader) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loaders = append(m.loaders, loader)
}

// Load 加载技能定义
func (m *DefinitionManager) Load(skillID string) (*SkillDefinition, error) {
	m.mu.RLock()
	if def, exists := m.definitions[skillID]; exists {
		m.mu.RUnlock()
		return def, nil
	}
	m.mu.RUnlock()

	// 尝试从文件加载
	skillDir := filepath.Join(m.baseDir, skillID)
	defFile := filepath.Join(skillDir, "SKILL.yaml")

	if _, err := os.Stat(defFile); os.IsNotExist(err) {
		defFile = filepath.Join(skillDir, "SKILL.json")
		if _, err := os.Stat(defFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("skill definition not found: %s", skillID)
		}
	}

	definition, err := m.loadFromFile(defFile)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.definitions[skillID] = definition
	m.mu.Unlock()

	return definition, nil
}

// LoadAll 加载所有技能定义
func (m *DefinitionManager) LoadAll() error {
	entries, err := os.ReadDir(m.baseDir)
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
		if _, err := m.Load(skillID); err != nil {
			// 记录错误但继续加载其他技能
			fmt.Printf("Warning: failed to load skill %s: %v\n", skillID, err)
		}
	}

	return nil
}

// Get 获取技能定义
func (m *DefinitionManager) Get(skillID string) (*SkillDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def, exists := m.definitions[skillID]
	return def, exists
}

// List 列出所有技能定义
func (m *DefinitionManager) List() []*SkillDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	defs := make([]*SkillDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		defs = append(defs, def)
	}

	return defs
}

// ListByCategory 按分类列出
func (m *DefinitionManager) ListByCategory(category string) []*SkillDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var defs []*SkillDefinition
	for _, def := range m.definitions {
		if def.Category == category {
			defs = append(defs, def)
		}
	}

	return defs
}

// Save 保存技能定义
func (m *DefinitionManager) Save(definition *SkillDefinition) error {
	if definition.ID == "" {
		return fmt.Errorf("skill ID is required")
	}

	// 创建技能目录
	skillDir := filepath.Join(m.baseDir, definition.ID)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("create skill directory failed: %w", err)
	}

	// 保存为 YAML（默认）
	defFile := filepath.Join(skillDir, "SKILL.yaml")

	// 使用 YAML 加载器保存
	yamlLoader := &YAMLLoader{}
	if err := yamlLoader.Save(defFile, definition); err != nil {
		return fmt.Errorf("save definition failed: %w", err)
	}

	definition.SourceFile = defFile
	definition.LoadedAt = time.Now()

	// 更新内存中的定义
	m.mu.Lock()
	m.definitions[definition.ID] = definition
	m.mu.Unlock()

	return nil
}

// SaveDefinition 保存技能定义（便捷方法，接受skillID和definition）
func (m *DefinitionManager) SaveDefinition(skillID string, definition interface{}) error {
	// 将通用的definition转换为SkillDefinition
	var skillDef *SkillDefinition

	switch def := definition.(type) {
	case *SkillDefinition:
		skillDef = def
		skillDef.ID = skillID
	default:
		// 尝试创建一个新的SkillDefinition
		skillDef = &SkillDefinition{
			ID:   skillID,
			Name: "Imported Skill",
		}
	}

	return m.Save(skillDef)
}

// Delete 删除技能定义
func (m *DefinitionManager) Delete(skillID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.definitions[skillID]; !exists {
		return fmt.Errorf("skill %s not found", skillID)
	}

	delete(m.definitions, skillID)

	// 删除文件
	skillDir := filepath.Join(m.baseDir, skillID)
	return os.RemoveAll(skillDir)
}

// GetCategories 获取所有分类
func (m *DefinitionManager) GetCategories() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	categories := make(map[string]bool)
	for _, def := range m.definitions {
		if def.Category != "" {
			categories[def.Category] = true
		}
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}

	return result
}

// Reload 重新加载所有定义
func (m *DefinitionManager) Reload() error {
	m.mu.Lock()
	m.definitions = make(map[string]*SkillDefinition)
	m.mu.Unlock()

	return m.LoadAll()
}

// loadFromFile 从文件加载
func (m *DefinitionManager) loadFromFile(filePath string) (*SkillDefinition, error) {
	for _, loader := range m.loaders {
		if loader.CanLoad(filePath) {
			definition, err := loader.Load(filePath)
			if err != nil {
				return nil, err
			}

			definition.SourceFile = filePath
			definition.LoadedAt = time.Now()

			return definition, nil
		}
	}

	return nil, fmt.Errorf("no suitable loader found for: %s", filePath)
}

// Validate 验证技能定义
func (d *SkillDefinition) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}
	if d.Description == "" {
		return fmt.Errorf("description is required")
	}

	// 验证工作流步骤
	for i, step := range d.Workflow {
		if step.ID == "" {
			return fmt.Errorf("workflow step %d: ID is required", i)
		}
		if step.Action == "" {
			return fmt.Errorf("workflow step %s: action is required", step.ID)
		}
	}

	return nil
}

// GetStep 根据ID获取工作流步骤
func (d *SkillDefinition) GetStep(stepID string) (*WorkflowStep, bool) {
	for _, step := range d.Workflow {
		if step.ID == stepID {
			return &step, true
		}
	}
	return nil, false
}

// GetTriggerTriggers 获取触发条件
func (d *SkillDefinition) GetTriggers() []string {
	if d.Triggers == nil {
		return []string{}
	}
	return d.Triggers
}

// ShouldTrigger 检查是否应该触发
func (d *SkillDefinition) ShouldTrigger(input string) bool {
	if len(d.Triggers) == 0 {
		return false
	}

	for _, trigger := range d.Triggers {
		if containsString(input, trigger) {
			return true
		}
	}

	return false
}

// YAMLLoader YAML加载器
type YAMLLoader struct{}

// JSONLoader JSON加载器
type JSONLoader struct{}

// CanLoad 检查是否可以加载
func (l *YAMLLoader) CanLoad(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".yaml" || ext == ".yml"
}

func (l *JSONLoader) CanLoad(filePath string) bool {
	return filepath.Ext(filePath) == ".json"
}

// Load 加载定义
func (l *YAMLLoader) Load(filePath string) (*SkillDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	var definition SkillDefinition
	if err := yamlUnmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("parse YAML failed: %w", err)
	}

	if err := definition.Validate(); err != nil {
		return nil, err
	}

	return &definition, nil
}

func (l *JSONLoader) Load(filePath string) (*SkillDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	var definition SkillDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	if err := definition.Validate(); err != nil {
		return nil, err
	}

	return &definition, nil
}

// Save 保存定义
func (l *YAMLLoader) Save(filePath string, definition *SkillDefinition) error {
	data, err := yamlMarshal(definition)
	if err != nil {
		return fmt.Errorf("marshal YAML failed: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

func (l *JSONLoader) Save(filePath string, definition *SkillDefinition) error {
	data, err := json.MarshalIndent(definition, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON failed: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

// 辅助函数

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		indexOfString(str, substr) >= 0)
}

func indexOfString(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// 简单的 YAML marshal/unmarshal（避免依赖 gopkg.in/yaml.v3）
func yamlUnmarshal(data []byte, v interface{}) error {
	// 这里简化处理，实际应该使用 yaml 库
	// 为了避免外部依赖，这里使用 JSON 作为后备
	return json.Unmarshal(data, v)
}

func yamlMarshal(v interface{}) ([]byte, error) {
	// 这里简化处理，实际应该使用 yaml 库
	// 为了避免外部依赖，这里使用 JSON 作为后备
	return json.MarshalIndent(v, "", "  ")
}
