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

package agent

import (
	"github.com/cloudwego/eino/schema"
)

// MemoryOptions contains memory management settings for agents
type MemoryOptions struct {
	MaxMessages    int     // Maximum number of messages to keep in history
	MaxMessageSize int     // Maximum size of a single message in bytes
	TrimRatio      float64 // Ratio of messages to keep when trimming (0.0-1.0)
	EnableTrimming bool    // Enable intelligent message trimming
}

// MemoryManager provides centralized memory management for agents.
// It manages message history trimming and optionally holds a MemorySearcher
// for RAG-based long-term memory retrieval.
type MemoryManager struct {
	opts     MemoryOptions
	searcher MemorySearcher // optional RAG/keyword search backend
}

// NewMemoryManager creates a new MemoryManager with the given options
func NewMemoryManager(opts MemoryOptions) *MemoryManager {
	// Set default values if not provided
	if opts.MaxMessages <= 0 {
		opts.MaxMessages = 100
	}
	if opts.TrimRatio <= 0 || opts.TrimRatio > 1 {
		opts.TrimRatio = 0.7
	}
	if !opts.EnableTrimming {
		opts.EnableTrimming = true
	}

	return &MemoryManager{opts: opts}
}

// DefaultMemoryManager returns a MemoryManager with default settings
func DefaultMemoryManager() *MemoryManager {
	return NewMemoryManager(MemoryOptions{
		MaxMessages:    100,
		MaxMessageSize: 0, // No size limit by default
		TrimRatio:      0.7,
		EnableTrimming: true,
	})
}

// SetSearcher attaches a MemorySearcher implementation (e.g. VectorMemory or
// SimpleMemory) to the manager so that ContextAssembler can perform RAG lookups.
func (m *MemoryManager) SetSearcher(s MemorySearcher) {
	m.searcher = s
}

// GetSearcher returns the current MemorySearcher and whether it is set.
func (m *MemoryManager) GetSearcher() (MemorySearcher, bool) {
	if m.searcher == nil {
		return nil, false
	}
	return m.searcher, true
}

// SetOptions updates the memory management options
func (m *MemoryManager) SetOptions(opts MemoryOptions) {
	m.opts = opts
}

// GetOptions returns the current memory management options
func (m *MemoryManager) GetOptions() MemoryOptions {
	return m.opts
}

// LimitHistory limits the message history with intelligent trimming
func (m *MemoryManager) LimitHistory(messages []*schema.Message) []*schema.Message {
	if !m.opts.EnableTrimming {
		return messages // No trimming enabled
	}

	maxMessages := m.opts.MaxMessages
	if maxMessages <= 0 {
		return messages // No message limit
	}

	// If we're already within the limit, no action needed
	if len(messages) <= maxMessages {
		return messages
	}

	// Create a new slice to hold trimmed messages
	trimmed := make([]*schema.Message, 0, maxMessages)

	// Always preserve system messages if present
	for _, msg := range messages {
		if msg.Role == schema.System {
			trimmed = append(trimmed, msg)
		}
	}

	// Calculate remaining slots for non-system messages
	remaining := maxMessages - len(trimmed)
	if remaining <= 0 {
		// If no slots left, keep only system messages
		return trimmed
	}

	// Apply trim ratio if specified, otherwise use remaining slots
	trimRatio := m.opts.TrimRatio
	if trimRatio <= 0 || trimRatio > 1 {
		trimRatio = 0.7 // Default trim ratio
	}

	// Calculate how many recent messages to keep
	recentCount := int(float64(len(messages)) * trimRatio)
	if recentCount > remaining {
		recentCount = remaining
	}

	// Add the most recent non-system messages
	startIndex := len(messages) - recentCount
	if startIndex < 0 {
		startIndex = 0
	}

	// Skip system messages since we already added them
	for i := startIndex; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role != schema.System {
			trimmed = append(trimmed, msg)
			if len(trimmed) >= maxMessages {
				break
			}
		}
	}

	return trimmed
}

// CheckMessageSize checks if a message exceeds the maximum size limit
func (m *MemoryManager) CheckMessageSize(msg *schema.Message) bool {
	if m.opts.MaxMessageSize <= 0 {
		return true // No size limit
	}

	// Calculate message size
	msgSize := len(msg.Content)
	for _, tc := range msg.ToolCalls {
		msgSize += len(tc.Function.Name) + len(tc.Function.Arguments)
	}

	return msgSize <= m.opts.MaxMessageSize
}

// TrimMessage trims a large message to fit within the size limit
func (m *MemoryManager) TrimMessage(msg *schema.Message) *schema.Message {
	if m.opts.MaxMessageSize <= 0 {
		return msg // No size limit
	}

	maxSize := m.opts.MaxMessageSize
	if len(msg.Content) <= maxSize {
		return msg // Already within limit
	}

	// Create a trimmed version of the message
	trimmed := &schema.Message{
		Role:       msg.Role,
		Content:    msg.Content[:maxSize-3] + "...", // Truncate with ellipsis
		ToolCalls:  msg.ToolCalls,
		ToolCallID: msg.ToolCallID,
	}

	return trimmed
}

// ProcessMessage checks and trims a message if necessary
func (m *MemoryManager) ProcessMessage(msg *schema.Message) *schema.Message {
	if !m.CheckMessageSize(msg) {
		return m.TrimMessage(msg)
	}
	return msg
}

// ClearHistory returns an empty message slice
func (m *MemoryManager) ClearHistory() []*schema.Message {
	return []*schema.Message{}
}
