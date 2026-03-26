// Agent Framework - Context Assembly
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Context Assembly builds the AI execution context from:
//   1. Workspace config (AGENTS.md / SOUL.md / TOOLS.md)
//   2. Session history
//   3. Memory search results
//   4. Active skills
//   5. Tool definitions
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

	"github.com/cloudwego/eino/schema"

	"AgentFramework/pkg/workspace"
)

// ContextConfig contains configuration for context assembly
type ContextConfig struct {
	// Workspace is the path to the workspace directory
	Workspace string

	// IncludeHistory includes session history in context
	IncludeHistory bool

	// HistoryLimit limits the number of recent messages to include
	HistoryLimit int

	// IncludeMemory includes memory search results
	IncludeMemory bool

	// MemoryLimit limits the number of memory results
	MemoryLimit int

	// IncludeSkills includes active skill instructions
	IncludeSkills bool

	// IncludeCapabilities includes capability definitions
	IncludeCapabilities bool

	// IncludeGuidelines includes behavioral guidelines
	IncludeGuidelines bool
}

// DefaultContextConfig returns the default context configuration
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		IncludeHistory:      true,
		HistoryLimit:        20,
		IncludeMemory:       true,
		MemoryLimit:         5,
		IncludeSkills:       true,
		IncludeCapabilities: true,
		IncludeGuidelines:   true,
	}
}

// Context represents the assembled execution context for an agent
type Context struct {
	// SystemPrompt is the system prompt for the agent
	SystemPrompt string

	// Messages is the conversation history
	Messages []*schema.Message

	// Session is the current session
	Session *Session

	// Skills are the active skills for this context
	Skills []string

	// MemoryResults are relevant memories retrieved by RAG search
	MemoryResults []MemoryResult

	// Tools are available tools
	Tools []string

	// Capabilities are agent capabilities
	Capabilities []string

	// WorkspaceConfig is the workspace configuration
	WorkspaceConfig *workspace.Config
}

// MemoryResult is defined in rag_memory.go

// ContextAssembler assembles the execution context for an agent
type ContextAssembler struct {
	wsPath   string
	wsCfg    *workspace.Config
	memory   *MemoryManager
	searcher MemorySearcher // optional; preferred over memory.GetSearcher() if set
	cfg      *ContextConfig
}

// NewContextAssembler creates a new context assembler
func NewContextAssembler(wsPath string, memory *MemoryManager, cfg *ContextConfig) (*ContextAssembler, error) {
	if cfg == nil {
		cfg = DefaultContextConfig()
	}

	// Load workspace configuration
	wsCfg := &workspace.Config{Root: wsPath}

	// Try to load workspace files
	if wsPath != "" {
		soulPath := filepath.Join(wsPath, "SOUL.md")
		if data, err := os.ReadFile(soulPath); err == nil {
			if soul, err := workspace.ParseSOUL(data); err == nil {
				wsCfg.SOUL = soul
			}
		}

		agentsPath := filepath.Join(wsPath, "AGENTS.md")
		if data, err := os.ReadFile(agentsPath); err == nil {
			if agents, err := workspace.ParseAGENTS(data); err == nil && len(agents) > 0 {
				wsCfg.AGENTS = agents
			}
		}

		capsPath := filepath.Join(wsPath, "CAPABILITIES.md")
		if data, err := os.ReadFile(capsPath); err == nil {
			if caps, err := workspace.ParseCAPABILITIES(data); err == nil {
				wsCfg.CAPABILITIES = caps
			}
		}
	}

	// Use defaults if not loaded
	if wsCfg.SOUL == nil {
		wsCfg.SOUL = workspace.DefaultSOUL()
	}
	if wsCfg.CAPABILITIES == nil {
		wsCfg.CAPABILITIES = workspace.DefaultCapabilities()
	}

	return &ContextAssembler{
		wsPath:  wsPath,
		wsCfg:   wsCfg,
		memory:  memory,
		cfg:     cfg,
	}, nil
}

// WithSearcher attaches a MemorySearcher to the ContextAssembler.
// This takes precedence over any searcher registered on the MemoryManager.
// Use this to inject a VectorMemory or SimpleMemory at startup.
func (ca *ContextAssembler) WithSearcher(s MemorySearcher) *ContextAssembler {
	ca.searcher = s
	return ca
}

// Assemble builds the execution context for the given session and input
func (ca *ContextAssembler) Assemble(ctx context.Context, session *Session, input string) (*Context, error) {
	execCtx := &Context{
		Session:         session,
		WorkspaceConfig: ca.wsCfg,
		Messages: []*schema.Message{
			{
				Role:    schema.User,
				Content: input,
			},
		},
	}

	// 1. Build system prompt from workspace config
	if err := ca.assembleSystemPrompt(ctx, execCtx); err != nil {
		return nil, fmt.Errorf("failed to assemble system prompt: %w", err)
	}

	// 2. Add session history
	if ca.cfg.IncludeHistory {
		if err := ca.assembleHistory(ctx, execCtx); err != nil {
			return nil, fmt.Errorf("failed to assemble history: %w", err)
		}
	}

	// 3. Add memory results
	if ca.cfg.IncludeMemory {
		if err := ca.assembleMemory(ctx, execCtx, input); err != nil {
			return nil, fmt.Errorf("failed to assemble memory: %w", err)
		}
	}

	// 4. Add skills
	if ca.cfg.IncludeSkills {
		ca.assembleSkills(ctx, execCtx)
	}

	// 5. Add capabilities
	if ca.cfg.IncludeCapabilities {
		ca.assembleCapabilities(ctx, execCtx)
	}

	return execCtx, nil
}

// assembleSystemPrompt builds the system prompt from workspace config
func (ca *ContextAssembler) assembleSystemPrompt(ctx context.Context, execCtx *Context) error {
	soul := ca.wsCfg.SOUL
	agents := ca.wsCfg.AGENTS

	var agentName, agentDesc string
	if len(agents) > 0 {
		agentName = agents[0].Name
		if agents[0].Description != "" {
			agentDesc = agents[0].Description
		}
	}

	// Build system prompt
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s\n\n", agentName))
	if agentDesc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", agentDesc))
	}

	// Soul / Personality
	if soul != nil {
		if soul.Personality != "" {
			sb.WriteString(fmt.Sprintf("## 个性特点\n%s\n\n", soul.Personality))
		}
		if soul.Motto != "" {
			sb.WriteString(fmt.Sprintf("## 理念\n%s\n\n", soul.Motto))
		}
		if len(soul.Values) > 0 {
			sb.WriteString("## 核心价值观\n")
			for _, v := range soul.Values {
				sb.WriteString(fmt.Sprintf("- %s\n", v))
			}
			sb.WriteString("\n")
		}
	}

	// Session context
	if execCtx.Session != nil {
		sessionInfo := fmt.Sprintf("## 当前会话\n- 类型: %s\n", execCtx.Session.Type)
		if execCtx.Session.Channel != "" {
			sessionInfo += fmt.Sprintf("- 渠道: %s\n", execCtx.Session.Channel)
		}
		sb.WriteString(sessionInfo)
		sb.WriteString("\n")
	}

	// Guidelines
	if ca.cfg.IncludeGuidelines {
		sb.WriteString("## 行为准则\n")
		for _, g := range ca.getGuidelines() {
			sb.WriteString(fmt.Sprintf("- %s\n", g))
		}
		sb.WriteString("\n")
	}

	execCtx.SystemPrompt = sb.String()
	return nil
}

// assembleHistory adds session history to the context
func (ca *ContextAssembler) assembleHistory(ctx context.Context, execCtx *Context) error {
	if execCtx.Session == nil {
		return nil
	}

	history := execCtx.Session.GetRecentMessages(ca.cfg.HistoryLimit)
	for _, msg := range history {
		role := schema.User
		content := msg.Content
		if msg.Role == "assistant" {
			role = schema.Assistant
		}

		execCtx.Messages = append(execCtx.Messages, &schema.Message{
			Role:    role,
			Content: content,
		})
	}

	return nil
}

// assembleMemory adds relevant memories to the context via RAG search.
// Resolution order for the search backend:
//  1. ca.searcher (set via WithSearcher)
//  2. ca.memory.GetSearcher() (registered on the MemoryManager)
//  3. skip silently
//
// Results are filtered by a minimum relevance score of 0.05, sorted by
// descending relevance, and stored in execCtx.MemoryResults.  They are also
// appended to execCtx.SystemPrompt as a Markdown section so the model can
// reference them.
func (ca *ContextAssembler) assembleMemory(ctx context.Context, execCtx *Context, input string) error {
	// Resolve the search backend.
	var searcher MemorySearcher
	if ca.searcher != nil {
		searcher = ca.searcher
	} else if ca.memory != nil {
		if s, ok := ca.memory.GetSearcher(); ok {
			searcher = s
		}
	}
	if searcher == nil {
		return nil // no memory backend configured
	}

	limit := ca.cfg.MemoryLimit
	if limit <= 0 {
		limit = 5
	}

	results, err := searcher.Search(ctx, input, limit)
	if err != nil {
		// Non-fatal: log and continue without memory context.
		return nil
	}

	// Filter results below the relevance threshold.
	const minScore = 0.05
	filtered := make([]MemoryResult, 0, len(results))
	for _, r := range results {
		if r.Score >= minScore {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	execCtx.MemoryResults = filtered

	// Append memory section to the system prompt so the model can reference it.
	if execCtx.SystemPrompt != "" && len(filtered) > 0 {
		var sb strings.Builder
		sb.WriteString("\n\n## 相关历史记忆\n\n")
		sb.WriteString("> 以下是与当前问题相关的历史记忆，供参考:\n\n")
		for i, r := range filtered {
			sb.WriteString(fmt.Sprintf("**[%d]** 相关性: %.2f  | 来源: `%s`\n\n", i+1, r.Score, r.Source))
			sb.WriteString(r.Content)
			sb.WriteString("\n\n---\n\n")
		}
		execCtx.SystemPrompt += sb.String()
	}

	return nil
}

// assembleSkills adds active skills to the context
func (ca *ContextAssembler) assembleSkills(ctx context.Context, execCtx *Context) {
	caps := ca.wsCfg.CAPABILITIES

	if caps != nil {
		for _, skill := range caps.Skills {
			execCtx.Skills = append(execCtx.Skills, skill)
		}
	}

	// Also include from session metadata
	if execCtx.Session != nil {
		if skills, ok := execCtx.Session.Metadata["active_skills"]; ok {
			for _, skill := range strings.Split(skills, ",") {
				skill = strings.TrimSpace(skill)
				if skill != "" {
					execCtx.Skills = append(execCtx.Skills, skill)
				}
			}
		}
	}
}

// assembleCapabilities adds capability descriptions to the context
func (ca *ContextAssembler) assembleCapabilities(ctx context.Context, execCtx *Context) {
	caps := ca.wsCfg.CAPABILITIES

	if caps != nil {
		for _, tool := range caps.Tools {
			execCtx.Capabilities = append(execCtx.Capabilities, tool)
		}
		for _, channel := range caps.Channels {
			execCtx.Capabilities = append(execCtx.Capabilities, "channel:"+channel)
		}
	}
}

// getGuidelines returns behavioral guidelines for the agent
func (ca *ContextAssembler) getGuidelines() []string {
	if ca.wsCfg.SOUL != nil && len(ca.wsCfg.SOUL.Guidelines) > 0 {
		return ca.wsCfg.SOUL.Guidelines
	}
	return []string{
		"Be helpful, harmless, and honest",
		"Respect user privacy and data security",
		"Provide accurate and relevant information",
		"Acknowledge uncertainty when appropriate",
		"Focus on being useful rather than verbose",
	}
}

// AssembleSimple creates a simple context with just the system prompt and input
// This is a convenience method for basic agent execution
func (ca *ContextAssembler) AssembleSimple(ctx context.Context, input string) (*Context, error) {
	return ca.Assemble(ctx, nil, input)
}

// AddMessages adds messages to the context
func (ec *Context) AddMessages(messages []*schema.Message) {
	ec.Messages = append(ec.Messages, messages...)
}

// GetSystemMessage returns the system prompt as a schema message
func (ec *Context) GetSystemMessage() *schema.Message {
	return &schema.Message{
		Role:    schema.System,
		Content: ec.SystemPrompt,
	}
}

// GetAllMessages returns all messages including system prompt
func (ec *Context) GetAllMessages() []*schema.Message {
	messages := []*schema.Message{ec.GetSystemMessage()}
	messages = append(messages, ec.Messages...)
	return messages
}

// HasMemory returns whether there are memory results
func (ec *Context) HasMemory() bool {
	return len(ec.MemoryResults) > 0
}

// HasSkills returns whether there are active skills
func (ec *Context) HasSkills() bool {
	return len(ec.Skills) > 0
}

// FormatMemoryResults formats memory results for inclusion in the system prompt.
// Results are ordered by descending relevance score and formatted as a Markdown section.
func (ec *Context) FormatMemoryResults() string {
	if !ec.HasMemory() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## Relevant Memories\n")

	for i, result := range ec.MemoryResults {
		sb.WriteString(fmt.Sprintf("\n[%d] %s (relevance: %.2f)\n%s\n",
			i+1, result.Source, result.Score, result.Content))
	}

	return sb.String()
}

// FormatSkills formats skills for inclusion in prompt
func (ec *Context) FormatSkills() string {
	if !ec.HasSkills() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## Active Skills\n")

	for _, skill := range ec.Skills {
		sb.WriteString(fmt.Sprintf("- %s\n", skill))
	}

	return sb.String()
}

// FormatCapabilities formats capabilities for inclusion in prompt
func (ec *Context) FormatCapabilities() string {
	if len(ec.Capabilities) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## Capabilities\n")

	for _, cap := range ec.Capabilities {
		sb.WriteString(fmt.Sprintf("- %s\n", cap))
	}

	return sb.String()
}

// LoadWorkspace loads workspace configuration from path
func LoadWorkspace(wsPath string) (*workspace.Config, error) {
	cfg := &workspace.Config{Root: wsPath}

	if wsPath == "" {
		return cfg, nil
	}

	// Load SOUL.md
	soulPath := filepath.Join(wsPath, "SOUL.md")
	if data, err := os.ReadFile(soulPath); err == nil {
		if soul, err := workspace.ParseSOUL(data); err == nil {
			cfg.SOUL = soul
		}
	}

	// Load AGENTS.md
	agentsPath := filepath.Join(wsPath, "AGENTS.md")
	if data, err := os.ReadFile(agentsPath); err == nil {
		if agents, err := workspace.ParseAGENTS(data); err == nil {
			cfg.AGENTS = agents
		}
	}

	// Load CAPABILITIES.md
	capsPath := filepath.Join(wsPath, "CAPABILITIES.md")
	if data, err := os.ReadFile(capsPath); err == nil {
		if caps, err := workspace.ParseCAPABILITIES(data); err == nil {
			cfg.CAPABILITIES = caps
		}
	}

	// Use defaults if not loaded
	if cfg.SOUL == nil {
		cfg.SOUL = workspace.DefaultSOUL()
	}
	if cfg.CAPABILITIES == nil {
		cfg.CAPABILITIES = workspace.DefaultCapabilities()
	}

	return cfg, nil
}
