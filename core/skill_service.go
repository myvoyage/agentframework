// Agent Framework - Skill Service
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
	"AgentFramework/agent/skills"
)

// SkillService handles skill operations
type SkillService struct {
	app *Application
}

// NewSkillService creates a new skill service
func NewSkillService(app *Application) *SkillService {
	return &SkillService{app: app}
}

// GetSkills returns all skills
func (s *SkillService) GetSkills(ctx context.Context) (map[string]agent.SkillMetadata, error) {
	skills := s.app.skillLibrary.GetAllSkills(ctx)
	result := make(map[string]agent.SkillMetadata)

	for name, skill := range skills {
		metadata := skill.GetMetadata(ctx)
		result[name] = metadata
	}

	return result, nil
}

// GetSkill returns a skill by name
func (s *SkillService) GetSkill(ctx context.Context, name string) (agent.SkillMetadata, error) {
	skill, found := s.app.skillLibrary.GetSkill(ctx, name)
	if !found {
		return agent.SkillMetadata{}, fmt.Errorf("skill not found: %s", name)
	}

	return skill.GetMetadata(ctx), nil
}

// DeleteSkill deletes a skill
func (s *SkillService) DeleteSkill(ctx context.Context, name string) error {
	return s.app.skillLibrary.UnregisterSkill(ctx, name)
}

// ExecuteSkillInput represents input for executing a skill
type ExecuteSkillInput struct {
	SkillName  string                 `json:"skillName"`
	Input      string                 `json:"input"`
	Workspace  string                 `json:"workspace"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ExecuteSkillOutput represents output from executing a skill
type ExecuteSkillOutput struct {
	Success bool                   `json:"success"`
	Result  interface{}             `json:"result,omitempty"`
	Error   string                   `json:"error,omitempty"`
	Stats   map[string]interface{}   `json:"stats,omitempty"`
}

// ExecuteSkill executes a skill by name
func (s *SkillService) ExecuteSkill(ctx context.Context, input *ExecuteSkillInput) (*ExecuteSkillOutput, error) {
	if s.app.skillSystem == nil {
		return &ExecuteSkillOutput{
			Success: false,
			Error:   "skill system not initialized",
		}, nil
	}

	// Create execution context
	execCtx := skills.NewExecutionContext()
	if input.Workspace != "" {
		execCtx.Workspace = input.Workspace
	} else {
		execCtx.Workspace = "/workspace"
	}

	// Add parameters to context
	if len(input.Parameters) > 0 {
		for key, value := range input.Parameters {
			execCtx.SetMetadata(key, value)
		}
	}

	// Execute skill
	result, err := s.app.skillSystem.ExecuteSkill(ctx, input.SkillName, input.Input, execCtx)
	if err != nil {
		return &ExecuteSkillOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ExecuteSkillOutput{
		Success: true,
		Result:  result,
		Stats: map[string]interface{}{
			"skill_name": input.SkillName,
			"workspace":  execCtx.Workspace,
		},
	}, nil
}

// ListSkillsTable prints skills in table format
func (s *SkillService) ListSkillsTable(ctx context.Context, outputFormat string) error {
	skills, err := s.GetSkills(ctx)
	if err != nil {
		return fmt.Errorf("failed to get skills: %w", err)
	}

	if len(skills) == 0 {
		fmt.Println("No skills found")
		return nil
	}

	// Print in requested format
	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", skills)
	case "table", "":
		fmt.Println("Skills:")
		fmt.Println("────────────────────────────────────────────────────────────")
		for name, skill := range skills {
			fmt.Printf("Name: %s\n", name)
			fmt.Printf("  Description: %s\n", skill.Description)
			fmt.Printf("  Version: %s\n", skill.Version)
			fmt.Printf("  Category: %s\n", skill.Category)
			fmt.Println("────────────────────────────────────────────────────────────")
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

// SkillSystemInfo contains information about skill system
type SkillSystemInfo struct {
	Initialized bool   `json:"initialized"`
	BaseDir     string `json:"baseDir"`
	TotalSkills int    `json:"totalSkills"`
}

// GetSkillSystemInfo returns basic information about skill system
func (s *SkillService) GetSkillSystemInfo(ctx context.Context) (*SkillSystemInfo, error) {
	if s.app.skillSystem == nil {
		return &SkillSystemInfo{Initialized: false}, nil
	}

	entries := s.app.skillSystem.Registry().ListAll()

	return &SkillSystemInfo{
		Initialized: true,
		BaseDir:     s.app.config.SkillSystemDir,
		TotalSkills: len(entries),
	}, nil
}
