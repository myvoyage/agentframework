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
	"bytes"
	"strings"
	"text/template"
)

// PromptComposer builds system prompts from workspace configuration
type PromptComposer struct {
	config *Config
}

// NewPromptComposer creates a new prompt composer
func NewPromptComposer(config *Config) *PromptComposer {
	return &PromptComposer{config: config}
}

// PromptContext contains all context for prompt composition
type PromptContext struct {
	AgentID      string            // Current agent ID
	Skills       []string          // Active skills
	Memory       string            // Retrieved memory
	SessionID    string            // Current session ID
	UserContext  map[string]string  // Additional user context
}

// BuildSystemPrompt builds the complete system prompt
func (p *PromptComposer) BuildSystemPrompt(ctx *PromptContext) (string, error) {
	var buf bytes.Buffer

	// 1. Header with agent identity
	buf.WriteString(p.buildHeader(ctx))

	// 2. Soul - personality and guidelines
	buf.WriteString(p.buildSoul())

	// 3. Capabilities - what the agent can do
	buf.WriteString(p.buildCapabilities())

	// 4. Skills - specific skill instructions
	buf.WriteString(p.buildSkills(ctx))

	// 5. Memory - relevant long-term memories
	buf.WriteString(p.buildMemory(ctx))

	// 6. Guidelines - behavioral rules
	buf.WriteString(p.buildGuidelines())

	return buf.String(), nil
}

// buildHeader creates the header with agent identity
func (p *PromptComposer) buildHeader(ctx *PromptContext) string {
	var buf bytes.Buffer
	soul := p.config.SOUL

	buf.WriteString("# ")
	buf.WriteString(soul.Name)
	buf.WriteString("\n\n")

	if soul.Motto != "" {
		buf.WriteString("> ")
		buf.WriteString(soul.Motto)
		buf.WriteString("\n\n")
	}

	buf.WriteString("---\n\n")
	return buf.String()
}

// buildSoul builds the personality section
func (p *PromptComposer) buildSoul() string {
	var buf bytes.Buffer
	soul := p.config.SOUL

	buf.WriteString("## 身份与个性\n\n")

	if soul.Personality != "" {
		buf.WriteString("**个性特征**: ")
		buf.WriteString(soul.Personality)
		buf.WriteString("\n\n")
	}

	if len(soul.Values) > 0 {
		buf.WriteString("**核心价值**:\n")
		for _, v := range soul.Values {
			buf.WriteString("- ")
			buf.WriteString(v)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// buildCapabilities builds the capabilities section
func (p *PromptComposer) buildCapabilities() string {
	var buf bytes.Buffer
	caps := p.config.CAPABILITIES

	buf.WriteString("## 可用能力\n\n")

	if len(caps.Tools) > 0 {
		buf.WriteString("**工具**:\n")
		for _, tool := range caps.Tools {
			buf.WriteString("- ")
			buf.WriteString(tool)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	if len(caps.Skills) > 0 {
		buf.WriteString("**技能**:\n")
		for _, skill := range caps.Skills {
			buf.WriteString("- ")
			buf.WriteString(skill)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// buildSkills builds skill-specific instructions
func (p *PromptComposer) buildSkills(ctx *PromptContext) string {
	if len(ctx.Skills) == 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("## 当前激活的技能\n\n")

	for _, skill := range ctx.Skills {
		buf.WriteString("### ")
		buf.WriteString(skill)
		buf.WriteString("\n")
		buf.WriteString("请按照 ")
		buf.WriteString(skill)
		buf.WriteString(" 技能的定义和指令执行任务。\n\n")
	}

	return buf.String()
}

// buildMemory builds the memory retrieval section
func (p *PromptComposer) buildMemory(ctx *PromptContext) string {
	if ctx.Memory == "" {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("## 相关记忆\n\n")
	buf.WriteString("以下是从长期记忆中检索到的相关内容：\n\n")
	buf.WriteString(ctx.Memory)
	buf.WriteString("\n\n")

	return buf.String()
}

// buildGuidelines builds behavioral guidelines
func (p *PromptComposer) buildGuidelines() string {
	var buf bytes.Buffer
	soul := p.config.SOUL

	buf.WriteString("## 行为准则\n\n")

	if len(soul.Guidelines) > 0 {
		for _, g := range soul.Guidelines {
			buf.WriteString("- ")
			buf.WriteString(g)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	buf.WriteString("请遵循以上准则，为用户提供最好的帮助。\n")
	return buf.String()
}

// BuildAgentPrompt builds a prompt for a specific agent
func (p *PromptComposer) BuildAgentPrompt(agentID string, ctx *PromptContext) (string, error) {
	agent := p.config.GetAgent(agentID)
	if agent == nil {
		return "", nil
	}

	var buf bytes.Buffer

	// Start with base prompt
	base, err := p.BuildSystemPrompt(ctx)
	if err != nil {
		return "", err
	}
	buf.WriteString(base)

	// Add agent-specific configuration
	buf.WriteString("## 当前 Agent 配置\n\n")
	buf.WriteString("**名称**: ")
	buf.WriteString(agent.Name)
	buf.WriteString("\n\n")

	buf.WriteString("**描述**: ")
	buf.WriteString(agent.Description)
	buf.WriteString("\n\n")

	if agent.Model != "" {
		buf.WriteString("**推荐模型**: ")
		buf.WriteString(agent.Model)
		buf.WriteString("\n\n")
	}

	if agent.Prompt != "" {
		buf.WriteString("**自定义指令**:\n")
		buf.WriteString(agent.Prompt)
		buf.WriteString("\n\n")
	}

	return buf.String(), nil
}

// PromptTemplate is a template for system prompts
const PromptTemplate = `{{.Header}}

{{.Soul}}

{{.Capabilities}}

{{if .Skills}}
{{.Skills}}
{{end}}

{{if .Memory}}
{{.Memory}}
{{end}}

{{.Guidelines}}
`

// BuildFromTemplate builds prompt using Go templates
func (p *PromptComposer) BuildFromTemplate(ctx *PromptContext) (string, error) {
	tmpl, err := template.New("system_prompt").Parse(PromptTemplate)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"Header":      p.buildHeader(ctx),
		"Soul":        p.buildSoul(),
		"Capabilities": p.buildCapabilities(),
		"Skills":      p.buildSkills(ctx),
		"Memory":      p.buildMemory(ctx),
		"Guidelines":  p.buildGuidelines(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// BuildContextualPrompt builds a contextual prompt with session info
func (p *PromptComposer) BuildContextualPrompt(agentID string, skills []string, memories []string, userPrefs map[string]string) (string, error) {
	ctx := &PromptContext{
		AgentID:     agentID,
		Skills:      skills,
		SessionID:   "",
		UserContext: userPrefs,
	}

	if len(memories) > 0 {
		ctx.Memory = strings.Join(memories, "\n\n")
	}

	return p.BuildAgentPrompt(agentID, ctx)
}
