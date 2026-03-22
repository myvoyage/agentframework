// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
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

package workspace

import (
	"testing"
)

func TestSOULParse(t *testing.T) {
	data := []byte(`
name: Test Agent
personality: helpful
motto: "test motto"

## Guidelines
- guideline 1
- guideline 2

## Values
- value 1
- value 2
`)

	cfg, err := ParseSOUL(data)
	if err != nil {
		t.Fatalf("ParseSOUL failed: %v", err)
	}

	if cfg.Name != "Test Agent" {
		t.Errorf("expected name 'Test Agent', got '%s'", cfg.Name)
	}

	if cfg.Motto != "test motto" {
		t.Errorf("expected motto 'test motto', got '%s'", cfg.Motto)
	}
}

func TestAGENTSParse(t *testing.T) {
	data := []byte(`agents:
  - id: test
    name: Test Agent
    description: A test agent
    skills:
      - skill1
      - skill2
    enabled: true
`)

	agents, err := ParseAGENTS(data)
	if err != nil {
		t.Fatalf("ParseAGENTS failed: %v", err)
	}

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	if agents[0].ID != "test" {
		t.Errorf("expected id 'test', got '%s'", agents[0].ID)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultSOUL()
	if cfg.Name != "OpenClaw" {
		t.Errorf("expected name 'OpenClaw', got '%s'", cfg.Name)
	}
}

func TestPromptComposer(t *testing.T) {
	cfg := &Config{
		SOUL: DefaultSOUL(),
		CAPABILITIES: DefaultCapabilities(),
	}

	composer := NewPromptComposer(cfg)
	ctx := &PromptContext{
		Skills: []string{"test-skill"},
	}

	prompt, err := composer.BuildSystemPrompt(ctx)
	if err != nil {
		t.Fatalf("BuildSystemPrompt failed: %v", err)
	}

	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
}
