// Agent Framework - Skill System Integration
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

package agent

import (
	"context"
	"fmt"

	"AgentFramework/agent/skills"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SkillSystem integrates the enhanced skill system with the Host
type SkillSystem struct {
	registry    *skills.SkillRegistry
	defManager  *skills.DefinitionManager
	examples    *skills.ExampleLibrary
	loader      *skills.ProgressiveLoader
	executor    *skills.EnhancedSkillExecutor
	initialized bool
}

// NewSkillSystem creates a new skill system with default configuration
func NewSkillSystem(baseDir string) (*SkillSystem, error) {
	if baseDir == "" {
		baseDir = ".skills"
	}

	// Create registry
	registry := skills.NewSkillRegistry(&skills.RegistryConfig{
		BaseDir:  baseDir + "/registry",
		AutoSave: true,
	})

	// Create definition manager
	defManager := skills.NewDefinitionManager(baseDir + "/definitions")

	// Create example library
	examples := skills.NewExampleLibrary(baseDir + "/examples")
	if err := examples.CreateBuiltInTemplates(); err != nil {
		return nil, fmt.Errorf("failed to create built-in templates: %w", err)
	}

	// Create progressive loader
	loader := skills.NewProgressiveLoader(baseDir + "/definitions")

	// Create enhanced executor
	executor := skills.NewEnhancedSkillExecutor(&skills.ExecutorConfig{
		EnableRetry:    true,
		MaxRetries:     3,
		RetryDelay:     1000000000, // 1 second in nanoseconds
		EnableTimeout:  true,
		DefaultTimeout: 30000000000, // 30 seconds
		EnableSkip:     true,
		EnableCache:    true,
		EnableLog:      true,
		LogLevel:       "info",
	})

	executor.SetRegistry(registry)
	executor.SetExamples(examples)

	return &SkillSystem{
		registry:    registry,
		defManager:  defManager,
		examples:    examples,
		loader:      loader,
		executor:    executor,
		initialized: true,
	}, nil
}

// Registry returns the skill registry
func (ss *SkillSystem) Registry() *skills.SkillRegistry {
	return ss.registry
}

// DefinitionManager returns the definition manager
func (ss *SkillSystem) DefinitionManager() *skills.DefinitionManager {
	return ss.defManager
}

// ExampleLibrary returns the example library
func (ss *SkillSystem) ExampleLibrary() *skills.ExampleLibrary {
	return ss.examples
}

// ProgressiveLoader returns the progressive loader
func (ss *SkillSystem) ProgressiveLoader() *skills.ProgressiveLoader {
	return ss.loader
}

// Executor returns the enhanced executor
func (ss *SkillSystem) Executor() *skills.EnhancedSkillExecutor {
	return ss.executor
}

// ExecuteSkill executes a skill by name with given input
func (ss *SkillSystem) ExecuteSkill(ctx context.Context, skillName string, input string, execCtx *skills.ExecutionContext) (interface{}, error) {
	definition, err := ss.defManager.Load(skillName)
	if err != nil {
		return nil, fmt.Errorf("failed to load skill definition %s: %w", skillName, err)
	}

	ss.executor.SetDefinition(definition)
	return ss.executor.Execute(ctx, input, execCtx)
}

// ConvertSkillsToTools converts registered skills to Eino tools
func (ss *SkillSystem) ConvertSkillsToTools(ctx context.Context) ([]tool.BaseTool, error) {
	entries := ss.registry.ListAll()
	tools := make([]tool.BaseTool, 0, len(entries))

	for _, entry := range entries {
		skillTool := &SkillTool{
			skillSystem: ss,
			entry:       entry,
		}
		tools = append(tools, skillTool)
	}

	return tools, nil
}

// SkillTool wraps a skill entry as an Eino tool
type SkillTool struct {
	skillSystem *SkillSystem
	entry       *skills.SkillEntry
}

// Info returns the tool info for the skill
func (st *SkillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return st.skillSystem.registry.ExportToEinoSchema(st.entry.ID)
}

// Invoke executes the skill tool
func (st *SkillTool) Invoke(ctx context.Context, input string) (string, error) {
	// Find the skill definition by category or name
	var skillName string

	// Try to find by category first
	defs := st.skillSystem.DefinitionManager().List()
	for _, def := range defs {
		if def.Category == st.entry.Category {
			skillName = def.ID
			break
		}
	}

	// Fallback to a default skill name based on category
	if skillName == "" {
		skillName = st.entry.Category
		if skillName == "http" {
			skillName = "http_request"
		} else if skillName == "api" {
			skillName = "api_call"
		} else if skillName == "file" {
			skillName = "file_operation"
		}
	}

	// Create execution context
	execCtx := skills.NewExecutionContext()
	execCtx.Workspace = "/workspace"

	// Execute the skill
	result, err := st.skillSystem.ExecuteSkill(ctx, skillName, input, execCtx)
	if err != nil {
		return "", fmt.Errorf("skill execution failed: %w", err)
	}

	// Convert result to string
	if resultStr, ok := result.(string); ok {
		return resultStr, nil
	}

	// Try to format as JSON
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Format the result map
		output := fmt.Sprintf("Skill '%s' executed successfully:\n", st.entry.Name)
		for key, value := range resultMap {
			output += fmt.Sprintf("  %s: %v\n", key, value)
		}
		return output, nil
	}

	return fmt.Sprintf("%v", result), nil
}

// Stream is not supported for skill tools
func (st *SkillTool) Stream(ctx context.Context, input string) (*schema.StreamReader[string], error) {
	return nil, fmt.Errorf("streaming is not supported for skill tools")
}

// HostOption is a function that configures a Host
type HostOption func(*HostConfig) error

// WithSkillSystem adds a skill system to the host
func WithSkillSystem(baseDir string) HostOption {
	return func(cfg *HostConfig) error {
		cfg.SkillSystemDir = baseDir
		return nil
	}
}

// WithSkillRegistry adds a custom skill registry to the host
func WithSkillRegistry(registry *skills.SkillRegistry) HostOption {
	return func(cfg *HostConfig) error {
		// This would need to be stored in the Host
		// For now, we'll mark it in the config
		if cfg.Extensions == nil {
			cfg.Extensions = make(map[string]interface{})
		}
		cfg.Extensions["skill_registry"] = registry
		return nil
	}
}

// WithDefinitionManager adds a custom definition manager to the host
func WithDefinitionManager(defManager *skills.DefinitionManager) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Extensions == nil {
			cfg.Extensions = make(map[string]interface{})
		}
		cfg.Extensions["definition_manager"] = defManager
		return nil
	}
}

// WithExampleLibrary adds a custom example library to the host
func WithExampleLibrary(library *skills.ExampleLibrary) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Extensions == nil {
			cfg.Extensions = make(map[string]interface{})
		}
		cfg.Extensions["example_library"] = library
		return nil
	}
}

// WithSkillLoader adds a custom skill loader to the host
func WithSkillLoader(loader *skills.ProgressiveLoader) HostOption {
	return func(cfg *HostConfig) error {
		if cfg.Extensions == nil {
			cfg.Extensions = make(map[string]interface{})
		}
		cfg.Extensions["skill_loader"] = loader
		return nil
	}
}

// GetSkillSystem retrieves the skill system from a Host
func (h *Host) GetSkillSystem() (*SkillSystem, error) {
	if h.cfg.Extensions == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	ss, ok := h.cfg.Extensions["skill_system"].(*SkillSystem)
	if !ok || ss == nil {
		return nil, fmt.Errorf("skill system not found in extensions")
	}

	if !ss.initialized {
		return nil, fmt.Errorf("skill system not properly initialized")
	}

	return ss, nil
}

// InitializeSkillSystem initializes the skill system for the host
func (h *Host) InitializeSkillSystem(baseDir string) error {
	if baseDir == "" {
		baseDir = ".skills"
	}

	ss, err := NewSkillSystem(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create skill system: %w", err)
	}

	if h.cfg.Extensions == nil {
		h.cfg.Extensions = make(map[string]interface{})
	}

	h.cfg.Extensions["skill_system"] = ss
	return nil
}

// GetSkillTools returns all registered skills as Eino tools
func (h *Host) GetSkillTools(ctx context.Context) ([]tool.BaseTool, error) {
	ss, err := h.GetSkillSystem()
	if err != nil {
		return nil, err
	}

	return ss.ConvertSkillsToTools(ctx)
}

// ExecuteSkill executes a skill by name
func (h *Host) ExecuteSkill(ctx context.Context, skillName string, input string, execCtx *skills.ExecutionContext) (interface{}, error) {
	ss, err := h.GetSkillSystem()
	if err != nil {
		return nil, err
	}

	return ss.ExecuteSkill(ctx, skillName, input, execCtx)
}
