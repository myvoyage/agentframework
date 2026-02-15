// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"AgentFramework/agent/skills"
)

// MarkdownSkillParser parses Markdown files with YAML Frontmatter (SKILL.md)
type MarkdownSkillParser struct {
	validators []SkillValidator
}

// NewMarkdownSkillParser creates a new MarkdownSkillParser with default validators
func NewMarkdownSkillParser() *MarkdownSkillParser {
	return &MarkdownSkillParser{
		validators: []SkillValidator{
			&EligibilityValidator{
				osDetector:     NewOSDetector(),
				binaryChecker: NewBinaryChecker(),
				envChecker:    NewEnvChecker(),
			},
		},
	}
}

// Parse parses a Markdown file with YAML Frontmatter
func (p *MarkdownSkillParser) Parse(filePath string) (*skills.SkillDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	return p.parse(data, filePath)
}

// ParseWithMetadata parses raw data with metadata extraction
func (p *MarkdownSkillParser) ParseWithMetadata(data []byte, filePath string) (*skills.SkillDefinition, map[string]interface{}, error) {
	def, err := p.parse(data, filePath)
	if err != nil {
		return nil, nil, err
	}

	metadata := map[string]interface{}{
		"id":          def.ID,
		"name":        def.Name,
		"description": def.Description,
		"category":    def.Category,
		"version":     def.Version,
		"author":      def.Author,
		"source_file": filePath,
	}

	return def, metadata, nil
}

// Validate validates a skill definition
func (p *MarkdownSkillParser) Validate(def *skills.SkillDefinition) error {
	for _, validator := range p.validators {
		if err := validator.Validate(def); err != nil {
			return err
		}
	}
	return nil
}

// parse does the actual parsing of Markdown with YAML Frontmatter
func (p *MarkdownSkillParser) parse(data []byte, filePath string) (*skills.SkillDefinition, error) {
	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("no YAML Frontmatter found in %s", filePath)
	}

	var frontmatter skills.SkillDefinition

	// Parse YAML Frontmatter
	if err := yaml.Unmarshal(parts[1], &frontmatter); err != nil {
		return nil, fmt.Errorf("parse YAML Frontmatter failed: %w", err)
	}

	// Parse Markdown body for AI instructions
	body := string(parts[2])
	frontmatter.SourceFile = filePath

	// Extract metadata from frontmatter
	if frontmatter.ID == "" {
		// Generate default ID from file path
		fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		frontmatter.ID = fmt.Sprintf("com.agentframework.skill.%s", fileName)
	}

	if frontmatter.Version == "" {
		frontmatter.Version = "1.0.0"
	}

	if frontmatter.Category == "" {
		frontmatter.Category = "general"
	}

	// Parse triggers from body if not specified in frontmatter
	if frontmatter.Triggers == nil || len(frontmatter.Triggers) == 0 {
		frontmatter.Triggers = p.extractTriggers(body)
	}

	// Parse prerequisites from body if not specified
	if frontmatter.Prerequisites == nil || len(frontmatter.Prerequisites) == 0 {
		frontmatter.Prerequisites = p.extractPrerequisites(body)
	}

	// Store markdown body in metadata
	if frontmatter.Metadata == nil {
		frontmatter.Metadata = make(map[string]interface{})
	}
	frontmatter.Metadata["markdown_body"] = body

	return &frontmatter, nil
}

// extractTriggers extracts trigger patterns from Markdown body
func (p *MarkdownSkillParser) extractTriggers(body string) []string {
	var triggers []string

	// Look for "## When to use" or "## Trigger" sections
	re := regexp.MustCompile(`(?i)##\s*When to use|##\s*Trigger`)
	lines := strings.Split(body, "\n")

	for i, line := range lines {
		if re.MatchString(line) {
			// Collect following lines until next header or end
			j := i + 1
			for j < len(lines) {
				line = strings.TrimSpace(lines[j])
				if strings.HasPrefix(line, "##") || line == "" {
					break
				}
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					line = strings.TrimLeft(line, "-* ")
					line = strings.TrimSpace(line)
					if line != "" {
						triggers = append(triggers, line)
					}
				}
				j++
			}
			break
		}
	}

	return triggers
}

// extractPrerequisites extracts prerequisites from Markdown body
func (p *MarkdownSkillParser) extractPrerequisites(body string) []skills.Prerequisite {
	var prerequisites []skills.Prerequisite

	// Look for "## Prerequisites" section
	re := regexp.MustCompile(`(?i)##\s*Prerequisites|##\s*Requirements`)
	lines := strings.Split(body, "\n")

	for i, line := range lines {
		if re.MatchString(line) {
			j := i + 1
			for j < len(lines) {
				line = strings.TrimSpace(lines[j])
				if strings.HasPrefix(line, "##") || line == "" {
					break
				}
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					line = strings.TrimLeft(line, "-* ")
					line = strings.TrimSpace(line)
					if line != "" {
						prerequisites = append(prerequisites, skills.Prerequisite{
							Type:        "custom",
							Description: line,
							Required:    true,
						})
					}
				}
				j++
			}
			break
		}
	}

	return prerequisites
}
