// Agent Framework - Skill Importer
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

package skills

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"
)

// SkillImporter handles importing skills from various sources
type SkillImporter struct {
	registry   *SkillRegistry
	defManager *DefinitionManager
	baseDir    string
}

// NewSkillImporter creates a new skill importer
func NewSkillImporter(registry *SkillRegistry, defManager *DefinitionManager, baseDir string) *SkillImporter {
	if baseDir == "" {
		baseDir = ".skills"
	}
	return &SkillImporter{
		registry:   registry,
		defManager: defManager,
		baseDir:    baseDir,
	}
}

// ImportSource represents the source of a skill import
type ImportSource struct {
	Type      string // "file", "url", "paste", "git"
	Data      []byte
	URL       string
	Content   string
	AuthToken string
	Username  string
	Password  string
	Branch    string
	Path      string
}

// ImportOptions contains options for importing a skill
type ImportOptions struct {
	SkillID    string
	Overwrite  bool
	AutoEnable bool
	Validate   bool
	Workspace  string
}

// ImportResult contains the result of a skill import
type ImportResult struct {
	Success    bool
	SkillID    string
	SkillName  string
	Message    string
	Warnings   []string
	Definition string // Changed from *SkillDefinition to string
}

// ImportYAMLMetadata represents the YAML frontmatter metadata from SKILL.md
type ImportYAMLMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
	Category    string   `yaml:"category"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
}

// ImportFromURL imports a skill from a URL
func (si *SkillImporter) ImportFromURL(ctx context.Context, url string, options ImportOptions) (*ImportResult, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if provided
	if options.Workspace != "" {
		req.Header.Set("Authorization", "Bearer "+options.Workspace)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Determine the type and import
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/zip") ||
		strings.HasSuffix(url, ".zip") {
		return si.ImportFromArchive(ctx, data, "zip", options)
	} else if strings.Contains(contentType, "application/x-gzip") ||
		strings.Contains(contentType, "application/gzip") ||
		strings.HasSuffix(url, ".tar.gz") ||
		strings.HasSuffix(url, ".tgz") {
		return si.ImportFromArchive(ctx, data, "tar.gz", options)
	} else {
		// Assume it's raw content
		return si.ImportFromContent(ctx, string(data), options)
	}
}

// ImportFromArchive imports a skill from a ZIP or TAR.GZ archive
func (si *SkillImporter) ImportFromArchive(ctx context.Context, data []byte, format string, options ImportOptions) (*ImportResult, error) {
	var files map[string][]byte
	var err error

	if format == "zip" {
		files, err = si.extractZip(data)
	} else if format == "tar.gz" {
		files, err = si.extractTarGz(data)
	} else {
		return nil, fmt.Errorf("unsupported archive format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Look for SKILL.md
	skillContent, ok := files["SKILL.md"]
	if !ok {
		// Try case-insensitive search
		for name, content := range files {
			if strings.ToLower(strings.TrimSuffix(name, ".md")) == "skill" {
				skillContent = content
				break
			}
		}
		if len(skillContent) == 0 {
			return nil, fmt.Errorf("SKILL.md not found in archive")
		}
	}

	// Parse the skill file
	result, err := si.ImportFromContent(ctx, string(skillContent), options)
	if err != nil {
		return nil, err
	}

	// Save additional files
	warnings := []string{}
	skillDir := filepath.Join(si.baseDir, "imported", result.SkillID)

	for name, content := range files {
		if name == "SKILL.md" || name == "skill.md" {
			continue
		}

		filePath := filepath.Join(skillDir, name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to create directory for %s: %v", name, err))
			continue
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to save %s: %v", name, err))
		}
	}

	result.Warnings = warnings
	return result, nil
}

// ImportFromContent imports a skill from raw content (SKILL.md format)
func (si *SkillImporter) ImportFromContent(ctx context.Context, content string, options ImportOptions) (*ImportResult, error) {
	// Parse YAML frontmatter and content
	metadata, body, err := si.parseSkillContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill content: %w", err)
	}

	// Validate metadata
	if options.Validate {
		if err := si.validateMetadata(metadata); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Generate skill ID if not provided
	skillID := options.SkillID
	if skillID == "" {
		skillID = si.generateSkillID(metadata.Name)
	}

	// Check if skill already exists
	if !options.Overwrite {
		if _, exists := si.registry.GetByID(skillID); exists {
			return nil, fmt.Errorf("skill already exists: %s (use overwrite to replace)", skillID)
		}
	}

	// Create skill entry
	entry := &SkillEntry{
		ID:          skillID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Tags:        metadata.Tags,
		Version:     metadata.Version,
		Enabled:     options.AutoEnable,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsedCount:   0,
	}

	// Register the skill
	if err := si.registry.Register(entry); err != nil {
		return nil, fmt.Errorf("failed to register skill: %w", err)
	}

	// Create definition
	// Convert ImportYAMLMetadata to map for Metadata field
	metadataMap := map[string]interface{}{
		"name":        metadata.Name,
		"description": metadata.Description,
		"version":     metadata.Version,
		"category":    metadata.Category,
		"author":      metadata.Author,
		"tags":        metadata.Tags,
	}

	definition := &SkillDefinition{
		ID:          skillID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Version:     metadata.Version,
		Category:    metadata.Category,
		Author:      metadata.Author,
		Metadata:    metadataMap,
		Workflow:    []WorkflowStep{},
		Config:      SkillConfig{},
	}

	// Save definition
	if err := si.defManager.Save(definition); err != nil {
		return nil, fmt.Errorf("failed to save definition: %w", err)
	}

	// Save the body content to a file
	skillDir := filepath.Join(si.baseDir, "definitions", skillID)
	contentFile := filepath.Join(skillDir, "content.md")
	if err := os.MkdirAll(skillDir, 0755); err == nil {
		_ = os.WriteFile(contentFile, []byte(body), 0644)
	}

	// 保存定义到文件后，将定义转换为字符串格式返回
	definitionBytes, err := yaml.Marshal(definition)
	if err != nil {
		definitionBytes = []byte(fmt.Sprintf("# %s\n\n%s", metadata.Name, body))
	}

	return &ImportResult{
		Success:    true,
		SkillID:    skillID,
		SkillName:  metadata.Name,
		Message:    "Skill imported successfully",
		Definition: string(definitionBytes),
	}, nil
}

// ImportFromGit imports a skill from a Git repository
func (si *SkillImporter) ImportFromGit(ctx context.Context, repoURL, branch, path string, options ImportOptions) (*ImportResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skill-import-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 克隆仓库
	var repo *git.Repository
	if branch != "" {
		// 克隆特定分支
		repo, err = git.PlainCloneContext(ctx, tempDir, false, &git.CloneOptions{
			URL:           repoURL,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
			SingleBranch:  true,
			Depth:         1,
		})
	} else {
		// 克隆默认分支
		repo, err = git.PlainCloneContext(ctx, tempDir, false, &git.CloneOptions{
			URL:   repoURL,
			Depth: 1,
		})
	}

	if err != nil {
		// 如果 go-git 失败，尝试使用系统 git 命令
		return si.importFromGitSystem(ctx, repoURL, branch, path, options, tempDir)
	}

	// 获取仓库引用
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo head: %w", err)
	}

	// 获取提交对象
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// 获取文件树
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// 查找 SKILL.md 文件
	skillPath := filepath.Join(path, "SKILL.md")
	file, err := tree.File(skillPath)
	if err != nil {
		// 尝试查找 skill.md（小写）
		skillPath = filepath.Join(path, "skill.md")
		file, err = tree.File(skillPath)
		if err != nil {
			// 尝试查找其他可能的文件名
			for _, name := range []string{"SKILL.MD", "Skill.md", "skill.MD"} {
				file, err = tree.File(filepath.Join(path, name))
				if err == nil {
					skillPath = filepath.Join(path, name)
					break
				}
			}
			if err != nil {
				return nil, fmt.Errorf("SKILL.md file not found in repository at path %s", path)
			}
		}
	}

	// 读取文件内容
	contents, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}

	// 解析技能内容
	metadata, definition, err := si.parseSkillContent(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill content: %w", err)
	}

	// 生成技能 ID
	skillID := fmt.Sprintf("git-%s-%s", filepath.Base(repoURL), metadata.Name)
	if options.SkillID != "" {
		skillID = options.SkillID
	}

	// 创建技能条目
	skillEntry := &SkillEntry{
		ID:          skillID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Version:     metadata.Version,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: map[string]interface{}{
			"source":    "git",
			"repo_url":  repoURL,
			"branch":    branch,
			"path":      path,
			"definition": definition,
		},
	}

	// 添加到注册表
	if err := si.registry.Register(skillEntry); err != nil {
		return nil, fmt.Errorf("failed to register skill: %w", err)
	}

	return &ImportResult{
		Success:    true,
		SkillID:    skillID,
		SkillName:  metadata.Name,
		Message:    "Skill imported from Git successfully",
		Definition: definition,
	}, nil
}

// importFromGitSystem 使用系统 git 命令导入（备用方案）
func (si *SkillImporter) importFromGitSystem(ctx context.Context, repoURL, branch, path string, options ImportOptions, tempDir string) (*ImportResult, error) {
	// 构建 git clone 命令
	args := []string{"clone", "--depth", "1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, tempDir)

	// 执行 git clone
	cmd := exec.CommandContext(ctx, "git", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// 读取 SKILL.md 文件
	skillPath := filepath.Join(tempDir, path, "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		// 尝试其他可能的文件名
		for _, name := range []string{"skill.md", "SKILL.MD", "Skill.md", "skill.MD"} {
			skillPath = filepath.Join(tempDir, path, name)
			content, err = os.ReadFile(skillPath)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("SKILL.md file not found in repository")
		}
	}

	// 解析技能内容
	metadata, definition, err := si.parseSkillContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill content: %w", err)
	}

	// 生成技能 ID
	skillID := fmt.Sprintf("git-%s-%s", filepath.Base(repoURL), metadata.Name)
	if options.SkillID != "" {
		skillID = options.SkillID
	}

	// 创建技能条目
	skillEntry := &SkillEntry{
		ID:          skillID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Version:     metadata.Version,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: map[string]interface{}{
			"source":    "git",
			"repo_url":  repoURL,
			"branch":    branch,
			"path":      path,
			"definition": definition,
		},
	}

	// 添加到注册表
	if err := si.registry.Register(skillEntry); err != nil {
		return nil, fmt.Errorf("failed to register skill: %w", err)
	}

	return &ImportResult{
		Success:    true,
		SkillID:    skillID,
		SkillName:  metadata.Name,
		Message:    "Skill imported from Git successfully",
		Definition: definition,
	}, nil
}

// parseSkillContent parses SKILL.md content with YAML frontmatter
func (si *SkillImporter) parseSkillContent(content string) (*ImportYAMLMetadata, string, error) {
	lines := strings.Split(content, "\n")

	// Check for YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return nil, "", fmt.Errorf("missing YAML frontmatter delimiter")
	}

	// Find the end of frontmatter
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return nil, "", fmt.Errorf("missing YAML frontmatter end delimiter")
	}

	// Parse YAML
	frontmatter := strings.Join(lines[1:endIndex], "\n")
	var metadata ImportYAMLMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract body content
	bodyContent := strings.Join(lines[endIndex+1:], "\n")

	return &metadata, bodyContent, nil
}

// validateMetadata validates the skill metadata
func (si *SkillImporter) validateMetadata(metadata *ImportYAMLMetadata) error {
	if metadata.Name == "" {
		return fmt.Errorf("name is required")
	}
	if metadata.Description == "" {
		return fmt.Errorf("description is required")
	}
	if metadata.Version == "" {
		return fmt.Errorf("version is required")
	}
	if metadata.Category == "" {
		metadata.Category = "custom"
	}
	if metadata.Tags == nil {
		metadata.Tags = []string{}
	}
	return nil
}

// generateSkillID generates a unique skill ID from the skill name
func (si *SkillImporter) generateSkillID(name string) string {
	// Convert to lowercase and replace spaces with underscores
	id := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Remove special characters
	id = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, id)

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s", id, timestamp)
}

// extractZip extracts files from a ZIP archive
func (si *SkillImporter) extractZip(data []byte) (map[string][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		// Normalize path separator
		name := strings.ReplaceAll(file.Name, "\\", "/")
		files[name] = content
	}

	return files, nil
}

// extractTarGz extracts files from a TAR.GZ archive
func (si *SkillImporter) extractTarGz(data []byte) (map[string][]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	files := make(map[string][]byte)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		content, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, err
		}

		// Normalize path separator
		name := strings.ReplaceAll(header.Name, "\\", "/")
		files[name] = content
	}

	return files, nil
}

// ValidateSkillFile validates a skill file without importing it
func (si *SkillImporter) ValidateSkillFile(content string) (*ImportYAMLMetadata, error) {
	metadata, _, err := si.parseSkillContent(content)
	if err != nil {
		return nil, err
	}

	if err := si.validateMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
