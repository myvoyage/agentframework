// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package templates

// Skill implementation template
const SkillImplTemplate = `package generated

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
`

// Skill test template
const SkillTestTemplate = `package generated

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
`

// MCP config template
const MCPConfigTemplate = `{
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
`