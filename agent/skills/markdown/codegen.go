// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"AgentFramework/agent/skills"
)

// GeneratedCode contains generated code and metadata
type GeneratedCode struct {
	SkillImpl     string            // Go skill implementation code
	TestCode      string            // Test code
	MCPConfig     string            // MCP tool configuration (JSON)
	Dependencies  []string         // Required Go modules
	FileInfo      map[string]string // File paths and contents
}

// SkillCodeGenerator generates Go code from SkillDefinition
type SkillCodeGenerator struct {
	templates *template.Template
}

// NewSkillCodeGenerator creates a new SkillCodeGenerator
func NewSkillCodeGenerator() *SkillCodeGenerator {
	return &SkillCodeGenerator{
		templates: loadTemplates(),
	}
}

// Generate generates all code for a skill definition
func (g *SkillCodeGenerator) Generate(def *skills.SkillDefinition) (*GeneratedCode, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "skill_impl", def); err != nil {
		return nil, fmt.Errorf("execute skill template failed: %w", err)
	}

	var testBuf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&testBuf, "skill_test", def); err != nil {
		return nil, fmt.Errorf("execute test template failed: %w", err)
	}

	var mcpBuf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&mcpBuf, "mcp_config", def); err != nil {
		return nil, fmt.Errorf("execute MCP config template failed: %w", err)
	}

	return &GeneratedCode{
		SkillImpl:     buf.String(),
		TestCode:      testBuf.String(),
		MCPConfig:     mcpBuf.String(),
		Dependencies:  []string{"AgentFramework/agent"},
		FileInfo: map[string]string{
			"skill_impl.go": fmt.Sprintf("generated/%s_impl.go", def.ID),
			"skill_test.go": fmt.Sprintf("generated/%s_test.go", def.ID),
			"mcp_config.json": fmt.Sprintf("generated/%s_mcp.json", def.ID),
		},
	}, nil
}

// GenerateSkillImplementation generates only the skill implementation code
func (g *SkillCodeGenerator) GenerateSkillImplementation(def *skills.SkillDefinition) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "skill_impl", def); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateMCPTools generates MCP tool configuration
func (g *SkillCodeGenerator) GenerateMCPTools(def *skills.SkillDefinition) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "mcp_config", def); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateTests generates test code
func (g *SkillCodeGenerator) GenerateTests(def *skills.SkillDefinition) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "skill_test", def); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// loadTemplates loads all code generation templates
func loadTemplates() *template.Template {
	tmpl := template.New("codegen").Funcs(template.FuncMap{
		"toCamelCase": toCamelCase,
		"contains":    contains,
	})

	// Skill implementation template
	tmpl.Parse(`package generated

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino/schema"
    "AgentFramework/agent/skills"
)

// Generated from {{.SourceFile}}
// DO NOT EDIT - This file is auto-generated

type {{.ID | toCamelCase}}Skill struct {
    skills.BaseSkill
    config *{{.ID | toCamelCase}}Config
}

type {{.ID | toCamelCase}}Config struct {
    // Generated from YAML config
}

func New{{.ID | toCamelCase}}Skill(config *{{.ID | toCamelCase}}Config) *{{.ID | toCamelCase}}Skill {
    if config == nil {
        config = &{{.ID | toCamelCase}}Config{}
    }

    return &{{.ID | toCamelCase}}Skill{
        BaseSkill: skills.BaseSkill{
            metadata: skills.SkillMetadata{
                ID:          "{{.ID}}",
                Name:        "{{.Name}}",
                Description: "{{.Description}}",
                Category:    "{{.Category}}",
                Tags:        []string{{printf "%#v" .Tags}},
                Version:     "{{.Version}}",
                Author:      "{{.Author}}",
            },
        },
        config: config,
    }
}

func (s *{{.ID | toCamelCase}}Skill) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: s.metadata.ID,
        Desc: s.metadata.Description,
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            {{- if .InputSchema.Properties}}
            {{- range $key, $value := .InputSchema.Properties}}
            "{{$key}}": &schema.ParameterInfo{
                Type:     schema.{{$value.Type}},
                Desc:     "{{$value.Description}}",
                Required: {{contains $key $.InputSchema.Required}},
            },
            {{- end}}
            {{- end}}
        }),
    }, nil
}

func (s *{{.ID | toCamelCase}}Skill) Invoke(ctx context.Context, input string) (string, error) {
    // Generated from AI instructions in markdown body
    // Implement skill logic here
    return "", fmt.Errorf("not implemented yet")
}

func (s *{{.ID | toCamelCase}}Skill) IsEnabled(ctx context.Context) bool {
    return !s.metadata.Disabled
}

func (s *{{.ID | toCamelCase}}Skill) SetEnabled(enabled bool) {
    s.metadata.Disabled = !enabled
}

func (s *{{.ID | toCamelCase}}Skill) GetMetadata(ctx context.Context) skills.SkillMetadata {
    return s.metadata
}

func (s *{{.ID | toCamelCase}}Skill) SetMetadata(metadata skills.SkillMetadata) {
    s.metadata = metadata
}
`)

	// Test template
	tmpl.Parse(`package generated

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "AgentFramework/agent/skills"
)

func Test{{.ID | toCamelCase}}Skill_Creation(t *testing.T) {
    config := &{{.ID | toCamelCase}}Config{}
    skill := New{{.ID | toCamelCase}}Skill(config)
    assert.NotNil(t, skill)
    assert.Equal(t, "{{.ID}}", skill.GetMetadata(context.Background()).ID)
    assert.Equal(t, "{{.Name}}", skill.GetMetadata(context.Background()).Name)
    assert.True(t, skill.IsEnabled(context.Background()))
}

func Test{{.ID | toCamelCase}}Skill_Info(t *testing.T) {
    skill := New{{.ID | toCamelCase}}Skill(nil)
    info, err := skill.Info(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, "{{.ID}}", info.Name)
    assert.Equal(t, "{{.Description}}", info.Desc)
}

func Test{{.ID | toCamelCase}}Skill_Invoke(t *testing.T) {
    skill := New{{.ID | toCamelCase}}Skill(nil)
    result, err := skill.Invoke(context.Background(), "")
    assert.Error(t, err)
    assert.Equal(t, "", result)
}
`)

	// MCP config template
	tmpl.Parse(`{
    "name": "{{.ID}}",
    "description": "{{.Description}}",
    "icon": "🔧",
    "category": "{{.Category}}",
    "tags": {{printf "%#v" .Tags}},
    "version": "{{.Version}}",
    "author": "{{.Author}}",
    "params": {
        {{- if .InputSchema.Properties}}
        {{- range $key, $value := .InputSchema.Properties}}
        "{{$key}}": {
            "type": "{{$value.Type}}",
            "description": "{{$value.Description}}",
            "required": {{contains $key $.InputSchema.Required}}
        },
        {{- end}}
        {{- end}}
    }
}
`)

	return tmpl
}

// toCamelCase converts a string to CamelCase
func toCamelCase(s string) string {
	var result strings.Builder
	nextUpper := false

	for i, c := range s {
		if i == 0 {
			result.WriteRune(unicode.ToUpper(c))
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			nextUpper = true
			continue
		}

		if nextUpper {
			result.WriteRune(unicode.ToUpper(c))
			nextUpper = false
		} else {
			result.WriteRune(unicode.ToLower(c))
		}
	}

	return result.String()
}

// contains checks if a string is in a slice
func contains(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
