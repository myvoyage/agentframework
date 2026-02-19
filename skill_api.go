// Agent Framework - Skill Management API
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// GetSkills returns all skills
func (a *App) GetSkills() (map[string]agent.SkillMetadata, error) {
	skills := a.core.GetSkillLibrary().GetAllSkills(a.ctx)
	result := make(map[string]agent.SkillMetadata)

	for name, skill := range skills {
		metadata := skill.GetMetadata(a.ctx)
		result[name] = metadata
	}

	return result, nil
}

// GetSkill returns a skill by name
func (a *App) GetSkill(name string) (agent.SkillMetadata, error) {
	skill, found := a.core.GetSkillLibrary().GetSkill(a.ctx, name)
	if !found {
		return agent.SkillMetadata{}, fmt.Errorf("skill not found: %s", name)
	}

	return skill.GetMetadata(a.ctx), nil
}

// DeleteSkill deletes a skill
func (a *App) DeleteSkill(name string) error {
	return a.core.GetSkillLibrary().UnregisterSkill(a.ctx, name)
}
