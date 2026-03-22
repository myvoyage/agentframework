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
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config represents the workspace configuration
type Config struct {
	mu sync.RWMutex

	// Workspace root directory
	Root string

	// Core configuration files
	SOUL         *SOULConfig         // Agent personality
	AGENTS       []*AgentDefinition  // Agent definitions
	CAPABILITIES *CapabilitiesConfig  // Available capabilities
}

// SOULConfig defines the agent's core personality and behavior
type SOULConfig struct {
	Name        string   `yaml:"name"`        // Agent name
	Personality string   `yaml:"personality"` // Core personality traits
	Values      []string `yaml:"values"`      // Core values
	Guidelines  []string `yaml:"guidelines"`  // Behavioral guidelines
	Motto       string   `yaml:"motto"`       // Agent motto/slogan
	Language    string   `yaml:"language"`    // Preferred language (default: "zh-CN")
}

// AgentDefinition defines a single agent configuration
type AgentDefinition struct {
	ID          string   `yaml:"id"`           // Unique agent ID
	Name        string   `yaml:"name"`         // Agent display name
	Description string   `yaml:"description"`  // Agent description
	Skills      []string `yaml:"skills"`       // Required skills
	Model       string   `yaml:"model"`        // Preferred model
	Prompt      string   `yaml:"prompt"`       // Additional system prompt
	Enabled     bool     `yaml:"enabled"`      // Whether agent is enabled
}

// CapabilitiesConfig defines available capabilities
type CapabilitiesConfig struct {
	Tools   []string `yaml:"tools"`    // Available tools
	Skills  []string `yaml:"skills"`   // Available skills
	Plugins []string `yaml:"plugins"`  // Available plugins
	Channels []string `yaml:"channels"` // Available channels
}

// Default workspace configuration files
const (
	FileSOUL         = "SOUL.md"
	FileAGENTS       = "AGENTS.md"
	FileCAPABILITIES = "CAPABILITIES.md"
	FileMEMORY       = "MEMORY.md"  // Long-term memory
	FileNOTES        = "NOTES.md"   // Daily notes
)

// New creates a new workspace configuration from the given root directory
func New(root string) (*Config, error) {
	cfg := &Config{Root: root}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Load loads all workspace configuration files
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Load SOUL.md
	soul, err := c.loadSOUL()
	if err != nil {
		errs = append(errs, err)
	}
	c.SOUL = soul

	// Load AGENTS.md
	agents, err := c.loadAGENTS()
	if err != nil {
		errs = append(errs, err)
	}
	c.AGENTS = agents

	// Load CAPABILITIES.md
	caps, err := c.loadCAPABILITIES()
	if err != nil {
		errs = append(errs, err)
	}
	c.CAPABILITIES = caps

	if len(errs) > 0 {
		// Return first error but continue with defaults
		return errs[0]
	}
	return nil
}

// loadSOUL loads the SOUL.md configuration file
func (c *Config) loadSOUL() (*SOULConfig, error) {
	path := filepath.Join(c.Root, FileSOUL)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSOUL(), nil
		}
		return nil, err
	}
	return ParseSOUL(data)
}

// loadAGENTS loads the AGENTS.md configuration file
func (c *Config) loadAGENTS() ([]*AgentDefinition, error) {
	path := filepath.Join(c.Root, FileAGENTS)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return ParseAGENTS(data)
}

// loadCAPABILITIES loads the CAPABILITIES.md configuration file
func (c *Config) loadCAPABILITIES() (*CapabilitiesConfig, error) {
	path := filepath.Join(c.Root, FileCAPABILITIES)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultCapabilities(), nil
		}
		return nil, err
	}
	return ParseCAPABILITIES(data)
}

// ParseSOUL parses SOUL.md content
func ParseSOUL(data []byte) (*SOULConfig, error) {
	content := strings.TrimSpace(string(data))

	cfg := &SOULConfig{
		Name:     "OpenClaw Agent",
		Language: "zh-CN",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse YAML frontmatter-like content
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "name":
				cfg.Name = value
			case "personality":
				cfg.Personality = value
			case "motto":
				cfg.Motto = strings.Trim(value, `"'`) // Remove quotes
			case "language":
				cfg.Language = value
			}
		}
	}

	// Extract personality and guidelines from content
	cfg.Guidelines = extractGuidelines(content)
	cfg.Values = extractValues(content)

	return cfg, nil
}

// ParseAGENTS parses AGENTS.md YAML content
func ParseAGENTS(data []byte) ([]*AgentDefinition, error) {
	// Try YAML format first with wrapper struct
	type agentsWrapper struct {
		Agents []*AgentDefinition `yaml:"agents"`
	}

	var wrapper agentsWrapper
	if err := yaml.Unmarshal(data, &wrapper); err == nil && len(wrapper.Agents) > 0 {
		return wrapper.Agents, nil
	}

	// Try direct slice format
	var agents []*AgentDefinition
	if err := yaml.Unmarshal(data, &agents); err == nil && len(agents) > 0 {
		return agents, nil
	}

	// Fallback: parse markdown list format
	agents = parseAgentsMarkdown(data)
	return agents, nil
}

// parseAgentsMarkdown parses agents from markdown list format
func parseAgentsMarkdown(data []byte) []*AgentDefinition {
	var agents []*AgentDefinition
	lines := strings.Split(string(data), "\n")
	var current *AgentDefinition

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			// New agent section
			if current != nil {
				agents = append(agents, current)
			}
			name := strings.TrimPrefix(line, "## ")
			current = &AgentDefinition{
				ID:      slugify(name),
				Name:    name,
				Enabled: true,
			}
		} else if strings.HasPrefix(line, "- ") && current != nil {
			// Agent property
			content := strings.TrimPrefix(line, "- ")
			if strings.HasPrefix(content, "Skills:") {
				skills := strings.TrimPrefix(content, "Skills:")
				current.Skills = strings.Split(skills, ",")
				for i := range current.Skills {
					current.Skills[i] = strings.TrimSpace(current.Skills[i])
				}
			} else if strings.HasPrefix(content, "Model:") {
				current.Model = strings.TrimPrefix(content, "Model:")
			} else {
				current.Description += " " + content
			}
		}
	}

	if current != nil {
		agents = append(agents, current)
	}

	return agents
}

// ParseCAPABILITIES parses CAPABILITIES.md YAML content
func ParseCAPABILITIES(data []byte) (*CapabilitiesConfig, error) {
	var caps CapabilitiesConfig
	if err := yaml.Unmarshal(data, &caps); err != nil {
		return DefaultCapabilities(), nil
	}
	return &caps, nil
}

// DefaultSOUL returns the default SOUL configuration
func DefaultSOUL() *SOULConfig {
	return &SOULConfig{
		Name:        "OpenClaw",
		Personality: "helpful, intelligent, and trustworthy",
		Values: []string{
			"User privacy first",
			"Transparent and explainable",
			"Reliable and consistent",
		},
		Guidelines: []string{
			"Always respect user privacy and data security",
			"Provide clear and honest responses",
			"Admit when you don't know something",
		},
		Motto:    "Your intelligent assistant",
		Language: "zh-CN",
	}
}

// DefaultCapabilities returns the default capabilities configuration
func DefaultCapabilities() *CapabilitiesConfig {
	return &CapabilitiesConfig{
		Tools:   []string{"file", "search", "code", "web"},
		Skills:  []string{},
		Plugins: []string{},
	}
}

// GetAgent returns an agent definition by ID
func (c *Config) GetAgent(id string) *AgentDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, agent := range c.AGENTS {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

// GetEnabledAgents returns all enabled agent definitions
func (c *Config) GetEnabledAgents() []*AgentDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var enabled []*AgentDefinition
	for _, agent := range c.AGENTS {
		if agent.Enabled {
			enabled = append(enabled, agent)
		}
	}
	return enabled
}

// Reload reloads the workspace configuration
func (c *Config) Reload() error {
	return c.Load()
}

// Helper functions

func extractGuidelines(content string) []string {
	var guidelines []string
	inGuidelines := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "guideline") {
			inGuidelines = true
			continue
		}
		if inGuidelines && strings.HasPrefix(line, "- ") {
			guidelines = append(guidelines, strings.TrimPrefix(line, "- "))
		} else if inGuidelines && line == "" {
			continue
		} else if inGuidelines {
			inGuidelines = false
		}
	}

	return guidelines
}

func extractValues(content string) []string {
	var values []string
	inValues := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "value") {
			inValues = true
			continue
		}
		if inValues && strings.HasPrefix(line, "- ") {
			values = append(values, strings.TrimPrefix(line, "- "))
		} else if inValues && line == "" {
			continue
		} else if inValues {
			inValues = false
		}
	}

	return values
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}
