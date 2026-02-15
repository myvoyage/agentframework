// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"fmt"

	"AgentFramework/agent/skills"
)

// MarkdownSkillManager manages all markdown skill operations
type MarkdownSkillManager struct {
	discoverer *MarkdownSkillDiscoverer
	parser     *MarkdownSkillParser
	generator  *SkillCodeGenerator
	validator  *EligibilityValidator
	registry   *skills.SkillRegistry
}

// MarkdownSkillConfig configuration for Markdown skill manager
type MarkdownSkillConfig struct {
	SkillDirs []string // Priority-based skill directories
	AutoLoad  bool     // Auto load on initialization
	AutoGen   bool     // Auto generate code for discovered skills
}

// NewMarkdownSkillManager creates a new MarkdownSkillManager
func NewMarkdownSkillManager(config *MarkdownSkillConfig, registry *skills.SkillRegistry) (*MarkdownSkillManager, error) {
	if config == nil {
		config = &MarkdownSkillConfig{
			SkillDirs: []string{
				"agent/skills/bundled",    // Priority 0 (highest)
				"~/.agentframework/skills", // Priority 1
				"./.skills",              // Priority 2 (lowest)
			},
			AutoLoad: true,
			AutoGen:  false,
		}
	}

	discoverer, err := NewMarkdownSkillDiscoverer(config.SkillDirs)
	if err != nil {
		return nil, fmt.Errorf("create discoverer failed: %w", err)
	}

	manager := &MarkdownSkillManager{
		discoverer: discoverer,
		parser:     NewMarkdownSkillParser(),
		generator:  NewSkillCodeGenerator(),
		validator:  NewEligibilityValidator(),
		registry:   registry,
	}

	if config.AutoLoad {
		if err := manager.LoadAll(); err != nil {
			return nil, fmt.Errorf("auto load skills failed: %w", err)
		}
	}

	return manager, nil
}

// LoadAll loads all skills from configured directories
func (m *MarkdownSkillManager) LoadAll() error {
	definitions, err := m.discoverer.Discover()
	if err != nil {
		return fmt.Errorf("discover skills failed: %w", err)
	}

	var errors []error
	for _, def := range definitions {
		if err := m.validator.Validate(def); err != nil {
			errors = append(errors, fmt.Errorf("validate skill %s failed: %w", def.ID, err))
			continue
		}

		if err := m.registerToRegistry(def); err != nil {
			errors = append(errors, fmt.Errorf("register skill %s failed: %w", def.ID, err))
		}

		if err := m.generateCodeIfNeeded(def); err != nil {
			errors = append(errors, fmt.Errorf("generate code for %s failed: %w", def.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("load failed with errors: %v", errors)
	}

	return nil
}

// LoadFromPath loads a single skill from file path
func (m *MarkdownSkillManager) LoadFromPath(path string) error {
	def, err := m.parser.Parse(path)
	if err != nil {
		return fmt.Errorf("parse skill failed: %w", err)
	}

	if err := m.validator.Validate(def); err != nil {
		return fmt.Errorf("validate skill failed: %w", err)
	}

	if err := m.registerToRegistry(def); err != nil {
		return fmt.Errorf("register skill failed: %w", err)
	}

	if err := m.generateCodeIfNeeded(def); err != nil {
		return fmt.Errorf("generate code failed: %w", err)
	}

	return nil
}

// GenerateCode generates code for a specific skill
func (m *MarkdownSkillManager) GenerateCode(skillID string) (*GeneratedCode, error) {
	// Find the skill in the discoverer (not just registry)
	definitions, err := m.discoverer.Discover()
	if err != nil {
		return nil, fmt.Errorf("discover skills failed: %w", err)
	}

	for _, def := range definitions {
		if def.ID == skillID {
			return m.generator.Generate(def)
		}
	}

	return nil, fmt.Errorf("skill %s not found", skillID)
}

// RegisterToRegistry registers all skills to the skill registry
func (m *MarkdownSkillManager) RegisterToRegistry() error {
	definitions, err := m.discoverer.Discover()
	if err != nil {
		return err
	}

	var errors []error
	for _, def := range definitions {
		if err := m.registerToRegistry(def); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("registration errors: %v", errors)
	}

	return nil
}

// registerToRegistry registers a single skill definition to the registry
func (m *MarkdownSkillManager) registerToRegistry(def *skills.SkillDefinition) error {
	entry := &skills.SkillEntry{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		Category:    def.Category,
		Tags:        []string{}, // Tags not defined on SkillDefinition
		Version:     def.Version,
		Enabled:     true, // Disabled not defined on SkillDefinition
		Config:      map[string]interface{}{}, // Convert SkillConfig to map
		Metadata:    def.Metadata,
	}

	return m.registry.Register(entry)
}

// generateCodeIfNeeded generates code for skills if AutoGen is configured
func (m *MarkdownSkillManager) generateCodeIfNeeded(def *skills.SkillDefinition) error {
	// Currently disabled by default
	return nil
}

// GetAll returns all discovered skill definitions
func (m *MarkdownSkillManager) GetAll() []*skills.SkillDefinition {
	definitions, err := m.discoverer.Discover()
	if err != nil {
		return nil
	}
	return definitions
}

// GetByID returns a skill by ID
func (m *MarkdownSkillManager) GetByID(id string) (*skills.SkillDefinition, bool) {
	definitions, err := m.discoverer.Discover()
	if err != nil {
		return nil, false
	}

	for _, def := range definitions {
		if def.ID == id {
			return def, true
		}
	}

	return nil, false
}

// GetByCategory returns skills by category
func (m *MarkdownSkillManager) GetByCategory(category string) []*skills.SkillDefinition {
	var result []*skills.SkillDefinition

	definitions, err := m.discoverer.Discover()
	if err != nil {
		return result
	}

	for _, def := range definitions {
		if def.Category == category {
			result = append(result, def)
		}
	}

	return result
}

// GetByTag returns skills with specific tag
func (m *MarkdownSkillManager) GetByTag(tag string) []*skills.SkillDefinition {
	var result []*skills.SkillDefinition

	definitions, err := m.discoverer.Discover()
	if err != nil {
		return result
	}

	// Tags are not defined on SkillDefinition, so we check metadata instead
	for _, def := range definitions {
		if metadataTags, ok := def.Metadata["tags"]; ok {
			if tagsSlice, isSlice := metadataTags.([]interface{}); isSlice {
				for _, t := range tagsSlice {
					if tagStr, isStr := t.(string); isStr && tagStr == tag {
						result = append(result, def)
						break
					}
				}
			}
		}
	}

	return result
}

// Watch starts watching for skill changes
func (m *MarkdownSkillManager) Watch(callback func(*skills.SkillDefinition, string)) error {
	return m.discoverer.Watch(func(def *skills.SkillDefinition, filePath string) {
		if def != nil {
			if err := m.registerToRegistry(def); err != nil {
				fmt.Printf("Failed to update skill %s: %v\n", def.ID, err)
			}
		} else {
			// TODO: Handle skill deletion
		}
		callback(def, filePath)
	})
}
