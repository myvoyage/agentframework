// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
//
// Enhanced Markdown Skill Parser
// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"AgentFramework/agent/skills"
)

// EnhancedMarkdownSkillParser 增强的 Markdown 技能解析器
type EnhancedMarkdownSkillParser struct {
	validators []SkillValidator
	cache       *SkillParseCache
}

// SkillParseCache 技能解析缓存
type SkillParseCache struct {
	entries map[string]*CacheEntry
	ttl     time.Duration
	mu      sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	definition   *skills.SkillDefinition
	parsedAt    time.Time
	fileModTime time.Time
}

// NewEnhancedMarkdownSkillParser 创建增强的 Markdown 技能解析器
func NewEnhancedMarkdownSkillParser() *EnhancedMarkdownSkillParser {
	return &EnhancedMarkdownSkillParser{
		validators: []SkillValidator{
			&EligibilityValidator{
				osDetector:     NewOSDetector(),
				binaryChecker: NewBinaryChecker(),
				envChecker:    NewEnvChecker(),
			},
			&EnhancedValidator{}, // 新增增强验证器
		},
		cache: NewSkillParseCache(5 * time.Minute), // 5分钟TTL
	}
}

// NewSkillParseCache 创建技能解析缓存
func NewSkillParseCache(ttl time.Duration) *SkillParseCache {
	return &SkillParseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Parse parses a Markdown file with enhanced features
func (p *EnhancedMarkdownSkillParser) Parse(filePath string) (*skills.SkillDefinition, error) {
	// 检查缓存
	if cached, hit := p.cache.Get(filePath); hit {
		return cached, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	fileInfo, _ := os.Stat(filePath)
	def, err := p.parse(data, filePath, fileInfo.ModTime())
	if err != nil {
		return nil, err
	}

	// 更新缓存
	p.cache.Set(filePath, def, fileInfo.ModTime())

	return def, nil
}

// parse does the actual parsing with enhanced features
func (p *EnhancedMarkdownSkillParser) parse(data []byte, filePath string, modTime time.Time) (*skills.SkillDefinition, error) {
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

	// Enhanced metadata extraction
	if err := p.enhanceMetadata(&frontmatter, filePath, body); err != nil {
		return nil, err
	}

	// Enhanced trigger extraction
	if frontmatter.Triggers == nil || len(frontmatter.Triggers) == 0 {
		frontmatter.Triggers = p.extractTriggers(body)
	}

	// Enhanced prerequisite extraction
	if frontmatter.Prerequisites == nil || len(frontmatter.Prerequisites) == 0 {
		frontmatter.Prerequisites = p.extractPrerequisites(body)
	}

	// Enhanced workflow extraction
	if frontmatter.Workflow == nil || len(frontmatter.Workflow) == 0 {
		workflow, err := p.extractWorkflow(body)
		if err != nil {
			return nil, fmt.Errorf("extract workflow failed: %w", err)
		}
		frontmatter.Workflow = workflow
	}

	// Store markdown body in metadata
	if frontmatter.Metadata == nil {
		frontmatter.Metadata = make(map[string]interface{})
	}
	frontmatter.Metadata["markdown_body"] = body
	frontmatter.Metadata["enhanced_parsed"] = true
	frontmatter.Metadata["parsed_at"] = time.Now().Format(time.RFC3339)

	return &frontmatter, nil
}

// enhanceMetadata 增强元数据提取
func (p *EnhancedMarkdownSkillParser) enhanceMetadata(def *skills.SkillDefinition, filePath, body string) error {
	// Generate default ID from file path
	if def.ID == "" {
		fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		def.ID = fmt.Sprintf("com.agentframework.skill.%s", fileName)
	}

	// Generate default version
	if def.Version == "" {
		def.Version = "1.0.0"
	}

	// Generate default category
	if def.Category == "" {
		def.Category = p.guessCategory(body)
	}

	// Extract tags from body
	if tags := p.extractTags(body); len(tags) > 0 {
		if def.Metadata == nil {
			def.Metadata = make(map[string]interface{})
		}
		def.Metadata["tags"] = tags
	}

	// Extract examples from body
	if examples := p.extractExamples(body); len(examples) > 0 {
		if def.Metadata == nil {
			def.Metadata = make(map[string]interface{})
		}
		def.Metadata["examples"] = examples
	}

	// Extract configuration schema from body
	if schema := p.extractConfigSchema(body); schema != nil {
		if def.Metadata == nil {
			def.Metadata = make(map[string]interface{})
		}
		def.Metadata["parameters"] = schema
	}

	return nil
}

// guessCategory 根据内容猜测技能分类
func (p *EnhancedMarkdownSkillParser) guessCategory(body string) string {
	lowerBody := strings.ToLower(body)

	categoryHints := map[string]string{
		"file":      "file_operation",
		"http":      "network",
		"request":    "network",
		"api":       "network",
		"database":   "database",
		"data":      "data_processing",
		"process":   "data_processing",
		"code":      "code_execution",
		"execute":   "code_execution",
		"browser":   "browser",
		"web":       "browser",
		"compute":   "computation",
		"analyze":   "analysis",
		"image":      "multimedia",
		"video":      "multimedia",
		"audio":      "multimedia",
	}

	for hint, category := range categoryHints {
		if strings.Contains(lowerBody, hint) {
			return category
		}
	}

	return "general"
}

// extractTags 从 Markdown body 提取标签
func (p *EnhancedMarkdownSkillParser) extractTags(body string) []string {
	var tags []string

	// Look for "## Tags" or "## Labels" section
	re := regexp.MustCompile(`(?i)##\s*Tags|##\s*Labels`)
	lines := strings.Split(body, "\n")

	for i, line := range lines {
		if re.MatchString(line) {
			// Collect following lines until next header
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
						tags = append(tags, line)
					}
				}
				j++
			}
			break
		}
	}

	return tags
}

// extractExamples 从 Markdown body 提取示例
func (p *EnhancedMarkdownSkillParser) extractExamples(body string) []string {
	var examples []string

	// Look for code blocks with language specifiers
	re := regexp.MustCompile("```(\\w+?)\n([^`]+)```")
	matches := re.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		if len(match) > 2 {
			lang := match[1]
			code := match[2]

			// Filter out non-code blocks
			if lang != "" && !strings.Contains(strings.ToLower(lang), "mermaid") {
				examples = append(examples, fmt.Sprintf("```%s\n%s```", lang, code))
			}
		}
	}

	return examples
}

// extractConfigSchema 从 Markdown body 提取配置架构
func (p *EnhancedMarkdownSkillParser) extractConfigSchema(body string) map[string]string {
	schema := make(map[string]string)

	// Look for "## Configuration" or "## Settings" section
	re := regexp.MustCompile(`(?i)##\s*Configuration|##\s*Settings|##\s*Config`)
	lines := strings.Split(body, "\n")

	for i, line := range lines {
		if re.MatchString(line) {
			// Parse following lines for key-value pairs
			j := i + 1
			for j < len(lines) {
				line = strings.TrimSpace(lines[j])
				if strings.HasPrefix(line, "##") || line == "" {
					break
				}

				// Parse key: value or key = value format
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						if key != "" && value != "" {
							schema[key] = value
						}
					}
				} else if strings.Contains(line, "=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						if key != "" && value != "" {
							schema[key] = value
						}
					}
				}

				j++
			}
			break
		}
	}

	return schema
}

// extractWorkflow 从 Markdown body 提取工作流
func (p *EnhancedMarkdownSkillParser) extractWorkflow(body string) ([]skills.WorkflowStep, error) {
	var steps []skills.WorkflowStep

	// Look for "## Workflow" section
	re := regexp.MustCompile(`(?i)##\s*Workflow`)
	lines := strings.Split(body, "\n")

	workflowStart := -1
	for i, line := range lines {
		if re.MatchString(line) {
			workflowStart = i
			break
		}
	}

	if workflowStart == -1 {
		// No workflow section found
		return steps, nil
	}

	// Parse workflow steps
	stepRe := regexp.MustCompile(`(?i)^\s*(\d+)\.|\s*\*`)
	currentStep := 1

	for i := workflowStart + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Stop at next header
		if strings.HasPrefix(line, "##") {
			break
		}

		// Parse step definition
		matches := stepRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			stepNum := matches[1]
			// Try to convert to int
			if _, err := fmt.Sscanf(stepNum, "%d", &currentStep); err == nil {
				// Extract step details
				step := p.parseWorkflowStep(lines, i)
				if step != nil {
					steps = append(steps, *step)
				}
				currentStep++
			}
		}
	}

	return steps, nil
}

// parseWorkflowStep 解析单个工作流步骤
func (p *EnhancedMarkdownSkillParser) parseWorkflowStep(lines []string, startIdx int) *skills.WorkflowStep {
	// Extract step information from multiple lines
	step := &skills.WorkflowStep{
		ID:      fmt.Sprintf("step_%d", startIdx),
		Name:     fmt.Sprintf("Step %d", startIdx),
		Action:   "execute", // Default action
		Timeout:  30 * time.Second, // Default timeout
	}

	// Parse action
	if startIdx < len(lines) {
		line := strings.ToLower(lines[startIdx])
		if strings.Contains(line, "validate") {
			step.Action = "validate"
		} else if strings.Contains(line, "prepare") {
			step.Action = "prepare"
		} else if strings.Contains(line, "cleanup") {
			step.Action = "cleanup"
		}
	}

	// Extract description from following lines
	startIdx++
	for startIdx < len(lines) {
		line := strings.TrimSpace(lines[startIdx])
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#") {
			break
		}
		if line != "" && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") {
			if step.Description == "" {
				step.Description = line
			} else {
				step.Description += " " + line
			}
		}
		startIdx++
	}

	return step
}

// extractTriggers extracts trigger patterns from Markdown body
func (p *EnhancedMarkdownSkillParser) extractTriggers(body string) []string {
	var triggers []string

	// Look for "## When to use" or "## Trigger" sections
	re := regexp.MustCompile(`(?i)##\s*When to use|##\s*Trigger|##\s*Usage`)
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
func (p *EnhancedMarkdownSkillParser) extractPrerequisites(body string) []skills.Prerequisite {
	var prerequisites []skills.Prerequisite

	// Look for "## Prerequisites" or "## Requirements" section
	re := regexp.MustCompile(`(?i)##\s*Prerequisites|##\s*Requirements|##\s*Dependencies`)
	lines := strings.Split(body, "\n")

	for i, line := range lines {
		if re.MatchString(line) {
			// Enhanced prerequisite parsing
			j := i + 1
			currentReq := &skills.Prerequisite{
				Type:        "custom",
				Required:    true,
				AutoFix:     false,
			}

			for j < len(lines) {
				line = strings.TrimSpace(lines[j])
				if strings.HasPrefix(line, "##") || line == "" {
					break
				}

				// Parse prerequisite type
				lowerLine := strings.ToLower(line)
				if strings.Contains(lowerLine, "file:") {
					currentReq.Type = "file_exists"
				} else if strings.Contains(lowerLine, "env:") {
					currentReq.Type = "env_var"
				} else if strings.Contains(lowerLine, "command:") {
					currentReq.Type = "command"
				} else if strings.Contains(lowerLine, "dependency:") {
					currentReq.Type = "dependency"
				}

				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					line = strings.TrimLeft(line, "-* ")
					line = strings.TrimSpace(line)
					if line != "" {
						currentReq.Description = line
						prerequisites = append(prerequisites, *currentReq)
						currentReq = &skills.Prerequisite{
							Type:        "custom",
							Required:    true,
							AutoFix:     false,
						}
					}
				}
				j++
			}
			break
		}
	}

	return prerequisites
}

// Get 从缓存获取
func (c *SkillParseCache) Get(filePath string) (*skills.SkillDefinition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[filePath]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.parsedAt) > c.ttl {
		delete(c.entries, filePath)
		return nil, false
	}

	// 检查文件是否被修改
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		delete(c.entries, filePath)
		return nil, false
	}

	if fileInfo.ModTime().After(entry.fileModTime) {
		delete(c.entries, filePath)
		return nil, false
	}

	return entry.definition, true
}

// Set 设置缓存
func (c *SkillParseCache) Set(filePath string, def *skills.SkillDefinition, modTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[filePath] = &CacheEntry{
		definition:   def,
		parsedAt:    time.Now(),
		fileModTime: modTime,
	}
}

// Clear 清空缓存
func (c *SkillParseCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Size 获取缓存大小
func (c *SkillParseCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// EnhancedValidator 增强验证器
type EnhancedValidator struct{}

// Validate 验证技能定义
func (v *EnhancedValidator) Validate(def *skills.SkillDefinition) error {
	// Basic validation
	if def.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if def.Name == "" {
		return fmt.Errorf("name is required")
	}
	if def.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Enhanced validation: Check ID format
	if !isValidSkillID(def.ID) {
		return fmt.Errorf("invalid ID format: %s", def.ID)
	}

	// Enhanced validation: Check version format
	if !isValidVersion(def.Version) {
		return fmt.Errorf("invalid version format: %s", def.Version)
	}

	// Enhanced validation: Check workflow steps
	for i, step := range def.Workflow {
		if step.ID == "" {
			return fmt.Errorf("workflow step %d: ID is required", i)
		}
		if step.Action == "" {
			return fmt.Errorf("workflow step %s: action is required", step.ID)
		}
		// Validate timeout
		if step.Timeout > 5*time.Minute {
			return fmt.Errorf("workflow step %s: timeout too long (max 5 minutes)", step.ID)
		}
	}

	return nil
}

// isValidSkillID 验证技能ID格式
func isValidSkillID(id string) bool {
	// ID should be in reverse domain format (e.g., com.agentframework.skill.name)
	parts := strings.Split(id, ".")
	return len(parts) >= 3
}

// isValidVersion 验证版本格式
func isValidVersion(version string) bool {
	// Version should be in semantic versioning format (e.g., 1.0.0)
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}

	return true
}
