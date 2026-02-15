// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown_test

import (
	"os"
	"path/filepath"
	"testing"

	"AgentFramework/agent/skills/markdown"
)

func TestMarkdownSkillParser_Parse(t *testing.T) {
	// Create a temporary test file
	testContent := `---
id: com.example.test
name: Test Skill
version: "1.0.0"
description: "A test skill"
category: "test"
author: "Test Author"
license: "MIT"
triggers:
  - "test"
prerequisites: []
workflow: []
config: {}
metadata:
  tags: ["test"]
  os: ["darwin", "linux", "windows"]
---

# Test Skill
This is a test skill.
`

	testFile, err := os.CreateTemp("", "test-skill-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(testFile.Name())

	if _, err := testFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := testFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test parsing the file
	parser := markdown.NewMarkdownSkillParser()
	def, err := parser.Parse(testFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify the parsed definition
	if def.ID != "com.example.test" {
		t.Errorf("Expected ID 'com.example.test', got '%s'", def.ID)
	}
	if def.Name != "Test Skill" {
		t.Errorf("Expected Name 'Test Skill', got '%s'", def.Name)
	}
	if def.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", def.Version)
	}
	if def.Description != "A test skill" {
		t.Errorf("Expected Description 'A test skill', got '%s'", def.Description)
	}
	if def.Category != "test" {
		t.Errorf("Expected Category 'test', got '%s'", def.Category)
	}
	if def.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", def.Author)
	}
	if def.License != "MIT" {
		t.Errorf("Expected License 'MIT', got '%s'", def.License)
	}
	if len(def.Triggers) != 1 || def.Triggers[0] != "test" {
		t.Errorf("Expected Triggers ['test'], got '%v'", def.Triggers)
	}
}

func TestMarkdownSkillDiscoverer_Discover(t *testing.T) {
	// Create a temporary directory for test skills
	testDir, err := os.MkdirTemp("", "test-skills-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a test skill file
	testSkillDir := filepath.Join(testDir, "example")
	if err := os.MkdirAll(testSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create test skill dir: %v", err)
	}

	testSkillFile := filepath.Join(testSkillDir, "SKILL.md")
	if err := os.WriteFile(testSkillFile, []byte(`---
id: com.example.discover
name: Discoverable Skill
version: "1.0.0"
description: "A skill for testing discovery"
category: "discovery"
author: "Test Author"
license: "MIT"
triggers:
  - "discover"
prerequisites: []
workflow: []
config: {}
metadata: {}
---

# Discoverable Skill
This skill is used for testing discovery functionality.
`), 0644); err != nil {
		t.Fatalf("Failed to write test skill file: %v", err)
	}

	// Test discovery
	discoverer, err := markdown.NewMarkdownSkillDiscoverer([]string{testDir})
	if err != nil {
		t.Fatalf("Failed to create discoverer: %v", err)
	}

	definitions, err := discoverer.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(definitions) != 1 {
		t.Errorf("Expected 1 skill to be discovered, got %d", len(definitions))
	} else {
		def := definitions[0]
		if def.ID != "com.example.discover" {
			t.Errorf("Expected skill ID 'com.example.discover', got '%s'", def.ID)
		}
		if def.Name != "Discoverable Skill" {
			t.Errorf("Expected skill Name 'Discoverable Skill', got '%s'", def.Name)
		}
	}
}

