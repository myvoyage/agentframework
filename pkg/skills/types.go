// Agent Framework - Skills Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// SkillRegistry manages skill registration and lookup
type SkillRegistry struct {
	skills map[string]*SkillEntry
	mu     sync.RWMutex
}

// NewSkillRegistry creates a new skill registry
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*SkillEntry),
	}
}

// Register registers a skill in the registry
func (r *SkillRegistry) Register(name string, skill *SkillEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[name] = skill
	return nil
}

// Get retrieves a skill from the registry
func (r *SkillRegistry) Get(name string) (*SkillEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, ok := r.skills[name]
	return skill, ok
}

// List returns all registered skills
func (r *SkillRegistry) List() []*SkillEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skills := make([]*SkillEntry, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills
}

// SkillEntry represents a skill entry in the registry
type SkillEntry struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Enabled     bool                   `json:"enabled"`
	Handler     SkillHandler           `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
	Info        *schema.ToolInfo       `json:"info"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UsedCount   int                    `json:"used_count"`
}

// SkillHandler is the function signature for skill execution
type SkillHandler func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// Skill represents the base skill interface
type Skill interface {
	Name() string
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
	Info() *schema.ToolInfo
}

// metadataVersionToString converts SkillVersion to string
func metadataVersionToString(v SkillVersion) string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
