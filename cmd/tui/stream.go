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
// Agent Framework - TUI Streaming Chat
// Copyright (C) 2025 Agent Framework Contributors
//
// 流式聊天组件 - 借鉴 Memoh �?streamChat 模式
//
// 注意：实际的数据加载�?Agent 调用实现已移�?integration.go
// 本文件保留高级流式处理功�?
package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// ========== 流式聊天会话 ==========

// StreamingChatSession 流式聊天会话
type StreamingChatSession struct {
	mu          sync.Mutex
	ctx         context.Context
	agentID     string
	sessionID   string
	buffer      strings.Builder
	done        bool
	err         error
	chunkChan   chan string
	doneChan    chan struct{}
}

// NewStreamingChatSession 创建流式聊天会话
func NewStreamingChatSession(ctx context.Context, agentID, sessionID string) *StreamingChatSession {
	return &StreamingChatSession{
		ctx:       ctx,
		agentID:   agentID,
		sessionID: sessionID,
		chunkChan: make(chan string, 10),
		doneChan:  make(chan struct{}),
	}
}

// Start 启动流式聊天
func (s *StreamingChatSession) Start(message string) tea.Cmd {
	return func() tea.Msg {
		// 调用 Agent 并处理流式响�?		// 这里需要根据实际的 Agent API 实现

		// 模拟流式响应
		go s.simulateStreamResponse(message)

		return nil
	}
}

// simulateStreamResponse 模拟流式响应（用于演示）
func (s *StreamingChatSession) simulateStreamResponse(message string) {
	defer close(s.doneChan)

	// 模拟逐字输出
	response := fmt.Sprintf("这是来自 Agent %s 的响应：\n\n", s.agentID)
	response += fmt.Sprintf("你发送的消息是：%s\n\n", message)
	response += "这是一个模拟的流式响应。在实际实现中，\n"
	response += "这里应该连接到真实的 Agent API，\n"
	response += "并逐字或逐块地接收响应内容。\n\n"
	response += "流式响应的优势：\n"
	response += "1. 更快的首字响应时间\n"
	response += "2. 更好的用户体验\n"
	response += "3. 更自然的对话感觉"

	for _, char := range response {
		select {
		case <-s.ctx.Done():
			s.err = s.ctx.Err()
			return
		default:
			s.chunkChan <- string(char)
		}
	}

	s.done = true
}

// NextChunk 获取下一个内容块
func (s *StreamingChatSession) NextChunk() (string, bool, error) {
	select {
	case chunk, ok := <-s.chunkChan:
		if !ok {
			return "", true, nil // 完成
		}
		s.buffer.WriteString(chunk)
		return chunk, false, nil
	case <-s.doneChan:
		return "", true, s.err
	case <-s.ctx.Done():
		return "", true, s.ctx.Err()
	}
}

// GetBuffer 获取当前缓冲区内�?func (s *StreamingChatSession) GetBuffer() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.String()
}

// IsDone 检查是否完�?func (s *StreamingChatSession) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
}

// ========== SSE 流式响应解析�?==========

// SSEParser Server-Sent Events 解析�?type SSEParser struct {
	scanner *bufio.Scanner
}

// NewSSEParser 创建 SSE 解析�?func NewSSEParser(reader io.Reader) *SSEParser {
	return &SSEParser{
		scanner: bufio.NewScanner(reader),
	}
}

// Next 读取下一个事�?func (p *SSEParser) Next() (string, error) {
	for p.scanner.Scan() {
		line := p.scanner.Text()

		// 跳过注释和空�?		if strings.HasPrefix(line, ":") || line == "" {
			continue
		}

		// 解析 "data: " 前缀
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			return data, nil
		}
	}

	if err := p.scanner.Err(); err != nil {
		return "", fmt.Errorf("scan error: %w", err)
	}

	return "", io.EOF
}

// ========== 流式响应助手 ==========

// StreamChunkMsg 流式内容块消�?type StreamChunkMsg struct {
	Chunk  string
	Done   bool
	Error  error
}

// StreamResponseCmd 创建流式响应命令
func StreamResponseCmd(ctx context.Context, reader io.Reader) tea.Cmd {
	return func() tea.Msg {
		parser := NewSSEParser(reader)

		for {
			data, err := parser.Next()
			if err == io.EOF {
				return StreamChunkMsg{Done: true}
			}
			if err != nil {
				return StreamChunkMsg{Error: err, Done: true}
			}

			// 解析 JSON 数据
			var chunk struct {
				Content string `json:"content"`
				Done    bool   `json:"done"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// 如果不是 JSON，直接返回文�?				return StreamChunkMsg{Chunk: data}
			}

			if chunk.Done {
				return StreamChunkMsg{Done: true}
			}

			return StreamChunkMsg{Chunk: chunk.Content}
		}
	}
}

// ========== 未来扩展：高级流式功�?==========

// 以下功能为预留扩展接口，用于未来的高级流式实现：
// - Server-Sent Events (SSE) 支持
// - 逐字输出效果
// - 实时打字机效�?// - 流式中断和恢�?