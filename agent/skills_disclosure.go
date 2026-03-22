// Agent Framework - Skills Progressive Disclosure
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Skills Progressive Disclosure implements a two-phase loading strategy:
//   1. Startup: Load only skill names + short descriptions (~97 chars)
//   2. Activation: Inject full SKILL.md content when skill is activated
//
// This avoids:
//   - Prompt bloat (not all skills in initial context)
//   - Latency (faster startup, smaller initial prompt)
//   - Token waste (only load what's actually needed)
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// SkillSummary is a short description for progressive disclosure (~97 chars)
type SkillSummary struct {
	Name       string
	ShortDesc  string // ~97 chars
	Path       string
	ActivatedAt time.Time
}

// SkillContent contains full skill content
type SkillContent struct {
	Metadata SkillMetadata
	Content  string // Full SKILL.md content
	LoadedAt time.Time
}

// SkillsRegistry manages skills with progressive disclosure
type SkillsRegistry struct {
	mu sync.RWMutex

	// skillsDir is the directory containing skills folders
	skillsDir string

	// summaries contains short descriptions (loaded at startup)
	summaries map[string]*SkillSummary

	// cache contains full skill content (loaded on demand)
	cache map[string]*SkillContent

	// watchEnabled enables hot-reload of skills
	watchEnabled bool

	// reloadChan receives reload notifications
	reloadChan chan string
}

// NewSkillsRegistry creates a new skills registry
func NewSkillsRegistry(skillsDir string, watchEnabled bool) *SkillsRegistry {
	return &SkillsRegistry{
		skillsDir:   skillsDir,
		summaries:    make(map[string]*SkillSummary),
		cache:        make(map[string]*SkillContent),
		watchEnabled: watchEnabled,
		reloadChan:   make(chan string, 10),
	}
}

// Initialize loads skill summaries from skills directory
// This is "progressive disclosure" part - only short descriptions
func (sr *SkillsRegistry) Initialize(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Clear existing data
	sr.summaries = make(map[string]*SkillSummary)
	sr.cache = make(map[string]*SkillContent)

	// Scan skills directory
	entries, err := os.ReadDir(sr.skillsDir)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check for SKILL.md
		skillPath := filepath.Join(sr.skillsDir, entry.Name())
		skillMDPath := filepath.Join(skillPath, "SKILL.md")

		if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
			continue
		}

		// Load and parse SKILL.md
		metadata, shortDesc, err := sr.parseSkillMetadata(ctx, skillMDPath)
		if err != nil {
			// Log warning but continue
			fmt.Printf("[SkillsRegistry] Failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		// Store summary
		sr.summaries[metadata.Name] = &SkillSummary{
			Name:      metadata.Name,
			ShortDesc: shortDesc,
			Path:      skillPath,
		}
	}

	// Start watch if enabled
	if sr.watchEnabled {
		go sr.watchLoop(ctx)
	}

	return nil
}

// GetSummaries returns all skill summaries (for startup context)
func (sr *SkillsRegistry) GetSummaries() map[string]*SkillSummary {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Return a copy to avoid race conditions
	summaries := make(map[string]*SkillSummary, len(sr.summaries))
	for k, v := range sr.summaries {
		summaries[k] = v
	}
	return summaries
}

// GetSkill retrieves full skill content (lazy load)
func (sr *SkillsRegistry) GetSkill(ctx context.Context, skillName string) (*SkillContent, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Check cache
	if content, exists := sr.cache[skillName]; exists {
		return content, nil
	}

	// Check if skill exists
	summary, exists := sr.summaries[skillName]
	if !exists {
		return nil, fmt.Errorf("skill %s not found", skillName)
	}

	// Load full content
	content, err := os.ReadFile(filepath.Join(summary.Path, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to load skill %s: %w", skillName, err)
	}

	// Parse full content
	metadata, _, err := sr.parseSkillMetadata(ctx, filepath.Join(summary.Path, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill metadata: %w", err)
	}

	// Cache content
	sr.cache[skillName] = &SkillContent{
		Metadata: *metadata,
		Content:  string(content),
		LoadedAt: time.Now(),
	}

	return sr.cache[skillName], nil
}

// FindSkills searches for skills matching a query
func (sr *SkillsRegistry) FindSkills(query string) []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []string

	for name, summary := range sr.summaries {
		if strings.Contains(strings.ToLower(summary.Name), query) ||
			strings.Contains(strings.ToLower(summary.ShortDesc), query) {
			matches = append(matches, name)
		}
	}

	return matches
}

// parseSkillMetadata parses skill frontmatter and extracts metadata + short description
func (sr *SkillsRegistry) parseSkillMetadata(ctx context.Context, path string) (*SkillMetadata, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	// Parse YAML frontmatter (between --- markers)
	sections := strings.Split(string(content), "---")
	if len(sections) < 2 {
		return nil, "", fmt.Errorf("invalid SKILL.md format: missing frontmatter")
	}

	// Parse YAML
	var metadata SkillMetadata
	if err := yaml.Unmarshal([]byte(sections[1]), &metadata); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract short description (~97 chars)
	shortDesc := sr.extractShortDesc(sections[2])

	return &metadata, shortDesc, nil
}

// extractShortDesc extracts a short description from skill content
func (sr *SkillsRegistry) extractShortDesc(content string) string {
	// Remove Markdown formatting
	content = strings.ReplaceAll(content, "**", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "`", "")

	// Remove whitespace
	content = strings.TrimSpace(content)

	// Truncate to ~97 chars
	maxLen := 97
	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen] + "..."
}

// watchLoop monitors skills directory for changes
func (sr *SkillsRegistry) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	lastModTimes := make(map[string]time.Time)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sr.checkForChanges(lastModTimes)
		}
	}
}

// checkForChanges checks if any skill files have been modified
func (sr *SkillsRegistry) checkForChanges(lastModTimes map[string]time.Time) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	entries, err := os.ReadDir(sr.skillsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillMDPath := filepath.Join(sr.skillsDir, entry.Name(), "SKILL.md")
		info, err := os.Stat(skillMDPath)
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		if lastMod, exists := lastModTimes[entry.Name()]; !exists || modTime.After(lastMod) {
			// File changed, reload
			sr.reloadSkill(entry.Name())
			lastModTimes[entry.Name()] = modTime
		}
	}
}

// reloadSkill reloads a single skill
func (sr *SkillsRegistry) reloadSkill(skillName string) {
	summary, exists := sr.summaries[skillName]
	if !exists {
		return
	}

	// Clear cache
	delete(sr.cache, skillName)

	// Reload summary
	ctx := context.Background()
	_, shortDesc, err := sr.parseSkillMetadata(ctx, filepath.Join(summary.Path, "SKILL.md"))
	if err != nil {
		fmt.Printf("[SkillsRegistry] Failed to reload %s: %v\n", skillName, err)
		return
	}

	summary.ShortDesc = shortDesc

	// Notify reload
	select {
	case sr.reloadChan <- skillName:
	default:
	}
}

// ReloadChannel returns the reload notification channel
func (sr *SkillsRegistry) ReloadChannel() <-chan string {
	return sr.reloadChan
}

// Activate activates a skill by name (loads full content)
func (sr *SkillsRegistry) Activate(ctx context.Context, skillName string) (*SkillContent, error) {
	return sr.GetSkill(ctx, skillName)
}

// Deactivate removes a skill from cache
func (sr *SkillsRegistry) Deactivate(skillName string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	delete(sr.cache, skillName)
}
