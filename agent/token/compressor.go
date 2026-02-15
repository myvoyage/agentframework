// Agent Framework - Token Compression
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package token

import (
	"context"
	"fmt"
)

// TokenCompressor Token 压缩器接口
type TokenCompressor interface {
	// CompressMessages 压缩消息列表到目标 Token 数量
	CompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error)

	// CompressText 压缩文本到目标 Token 数量
	CompressText(ctx context.Context, text string, targetTokens int) (string, error)

	// SetStrategy 设置压缩策略
	SetStrategy(strategy CompressionStrategy)

	// GetStrategy 获取当前压缩策略
	GetStrategy() CompressionStrategy

	// GetStats 获取压缩统计信息
	GetStats() *CompressionStats
}

// CompressionStats 压缩统计信息
type CompressionStats struct {
	TotalCompressions int64
	TotalInputTokens  int64
	TotalOutputTokens  int64
	AverageRatio     float64
	LastCompression   *CompressionOperation
}

// CompressionOperation 单次压缩操作记录
type CompressionOperation struct {
	Strategy      CompressionStrategy
	InputTokens   int
	OutputTokens  int
	CompressionRatio float64
	DurationMs    int64
	Success       bool
	Error         string
}

// CompressConfig 压缩器配置
type CompressConfig struct {
	Strategy           CompressionStrategy
	TargetTokens       int
	MinTokens          int
	MaxTokens          int
	PreserveSystemMessages bool
	SummaryModelName   string
	SummaryMaxTokens   int
	Temperature         float64
}

// DefaultCompressConfig 返回默认压缩配置
func DefaultCompressConfig() *CompressConfig {
	return &CompressConfig{
		Strategy:             StrategyHybrid,
		TargetTokens:         4000,
		MinTokens:            100,
		MaxTokens:            16000,
		PreserveSystemMessages: true,
		SummaryModelName:     "gpt-4o-mini",
		SummaryMaxTokens:     500,
		Temperature:          0.7,
	}
}

// MessageCompressor 消息压缩器
type MessageCompressor struct {
	config     *CompressConfig
	strategy   CompressionStrategy
	counter    TokenCounter
	stats      *CompressionStats
	llmFunc    LLMCompressFunc
}

// LLMCompressFunc LLM 压缩函数类型
type LLMCompressFunc func(ctx context.Context, prompt string, maxTokens int) (string, error)

// NewMessageCompressor 创建消息压缩器
func NewMessageCompressor(config *CompressConfig, llmFunc LLMCompressFunc) *MessageCompressor {
	if config == nil {
		config = DefaultCompressConfig()
	}

	return &MessageCompressor{
		config:   config,
		strategy: config.Strategy,
		counter:  NewDefaultTokenCounter(),
		stats:    &CompressionStats{},
		llmFunc:  llmFunc,
	}
}

// CompressMessages 压缩消息列表
func (m *MessageCompressor) CompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// 计算当前 Token 总数
	currentTokens := m.counter.CountMessages(messages)

	// 如果已经小于目标，直接返回
	if currentTokens <= targetTokens {
		return messages, nil
	}

	// 根据策略进行压缩
	var compressed []interface{}
	var err error

	switch m.strategy {
	case StrategyTruncate:
		compressed, err = m.truncateMessages(messages, targetTokens)
	case StrategySummarize:
		compressed, err = m.summarizeMessages(ctx, messages, targetTokens)
	case StrategySemantic:
		compressed, err = m.semanticCompressMessages(ctx, messages, targetTokens)
	case StrategyHybrid:
		compressed, err = m.hybridCompressMessages(ctx, messages, targetTokens)
	default:
		return nil, fmt.Errorf("unsupported compression strategy: %s", m.strategy)
	}

	if err != nil {
		return nil, err
	}

	// 更新统计信息
	m.updateStats(messages, compressed, m.strategy)

	return compressed, nil
}

// CompressText 压缩文本
func (m *MessageCompressor) CompressText(ctx context.Context, text string, targetTokens int) (string, error) {
	if text == "" {
		return text, nil
	}

	currentTokens := m.counter.CountText(text)

	if currentTokens <= targetTokens {
		return text, nil
	}

	switch m.strategy {
	case StrategySummarize, StrategyHybrid:
		if m.llmFunc != nil {
			prompt := m.buildSummaryPrompt(text, targetTokens)
			summary, err := m.llmFunc(ctx, prompt, m.config.SummaryMaxTokens)
			if err != nil {
				return "", fmt.Errorf("LLM summarization failed: %w", err)
			}
			return summary, nil
		}
		fallthrough
	default:
		// 对于截断策略，简单截断
		return m.truncateText(text, targetTokens), nil
	}
}

// SetStrategy 设置压缩策略
func (m *MessageCompressor) SetStrategy(strategy CompressionStrategy) {
	m.strategy = strategy
}

// GetStrategy 获取当前策略
func (m *MessageCompressor) GetStrategy() CompressionStrategy {
	return m.strategy
}

// GetStats 获取统计信息
func (m *MessageCompressor) GetStats() *CompressionStats {
	return m.stats
}

// ===== 压缩策略实现 =====

// truncateMessages 截断策略：保留最新的 N 条消息
func (m *MessageCompressor) truncateMessages(messages []interface{}, targetTokens int) ([]interface{}, error) {
	result := make([]interface{}, 0)
	currentTokens := 0

	// 从后向前遍历，保留最新消息
	for i := len(messages) - 1; i >= 0; i-- {
		msgTokens := m.counter.CountMessage(messages[i])

		// 保留系统消息
		if m.config.PreserveSystemMessages && isSystemMessage(messages[i]) {
			result = append([]interface{}{messages[i]}, result...)
			continue
		}

		if currentTokens+msgTokens > targetTokens {
			break
		}

		currentTokens += msgTokens
		result = append([]interface{}{messages[i]}, result...)
	}

	return result, nil
}

// summarizeMessages 摘要策略：使用 LLM 生成摘要
func (m *MessageCompressor) summarizeMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
	if m.llmFunc == nil {
		return m.truncateMessages(messages, targetTokens)
	}

	// 构建摘要提示词
	prompt := m.buildMessagesSummaryPrompt(messages, targetTokens)

	summary, err := m.llmFunc(ctx, prompt, m.config.SummaryMaxTokens)
	if err != nil {
		return nil, fmt.Errorf("LLM summarization failed: %w", err)
	}

	// 创建摘要消息
	summaryMsg := map[string]interface{}{
		"role":    "system",
		"content": fmt.Sprintf("[Previous conversation summary]\n%s", summary),
	}

	return []interface{}{summaryMsg}, nil
}

// semanticCompressMessages 语义压缩策略：按重要性保留消息
func (m *MessageCompressor) semanticCompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
	// 简化实现：给消息类型打分
	// 优先级: system > assistant > user
	scores := make([]float64, len(messages))
	for i, msg := range messages {
		scores[i] = m.getMessageScore(msg)
	}

	// 按分数排序并选择高优先级消息
	return m.selectMessagesByScore(messages, scores, targetTokens), nil
}

// hybridCompressMessages 混合压缩策略
func (m *MessageCompressor) hybridCompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
	// 阶段1：先尝试语义压缩
	result, err := m.semanticCompressMessages(ctx, messages, targetTokens)
	if err == nil && m.counter.CountMessages(result) <= targetTokens {
		return result, nil
	}

	// 阶段2：如果还不够，使用 LLM 摘要旧消息
	if m.llmFunc != nil {
		halfIdx := len(messages) / 2
		oldMessages := messages[:halfIdx]
		newMessages := messages[halfIdx:]

		if len(oldMessages) > 0 {
			summaryPrompt := m.buildMessagesSummaryPrompt(oldMessages, targetTokens/2)
			summary, err := m.llmFunc(ctx, summaryPrompt, m.config.SummaryMaxTokens)
			if err == nil {
				summaryMsg := map[string]interface{}{
					"role":    "system",
					"content": fmt.Sprintf("[Earlier conversation summary]\n%s", summary),
				}
				return append([]interface{}{summaryMsg}, newMessages...), nil
			}
		}
	}

	// 阶段3：最后使用截断策略
	return m.truncateMessages(messages, targetTokens)
}

// ===== 辅助方法 =====

// truncateText 截断文本到目标 Token 数量
func (m *MessageCompressor) truncateText(text string, targetTokens int) string {
	tokens := GetTokensFromText(text, targetTokens)
	if len(tokens) > 0 {
		return tokens[0] + "..."
	}
	return text
}

// buildSummaryPrompt 构建文本摘要提示词
func (m *MessageCompressor) buildSummaryPrompt(text string, targetTokens int) string {
	return fmt.Sprintf(`请将以下内容简洁地总结，确保总结不超过 %d tokens。

内容：
%s

请保留关键信息和重要细节。`, targetTokens, text)
}

// buildMessagesSummaryPrompt 构建优化的消息摘要提示词
func (m *MessageCompressor) buildMessagesSummaryPrompt(messages []interface{}, targetTokens int) string {
	prompt := fmt.Sprintf(`请将以下对话历史总结成一个简洁、全面的摘要，不超过 %d 个 token。

## 摘要要求：
1. **准确性**：保留所有关键信息和核心观点
2. **完整性**：包含重要的决策、行动项目和问题
3. **结构**：使用要点列表或段落形式，层次清晰
4. **简洁性**：避免重复和冗余信息
5. **客观性**：保持中立，不添加个人观点

## 对话历史：
`, targetTokens)

	for i, msg := range messages {
		var role string
		var content string

		switch m := msg.(type) {
		case map[string]interface{}:
			if r, ok := m["role"]; ok {
				role = fmt.Sprintf("%v", r)
			}
			if c, ok := m["content"]; ok {
				content = fmt.Sprintf("%v", c)
			}
		case map[string]string:
			if r, ok := m["role"]; ok {
				role = r
			}
			if c, ok := m["content"]; ok {
				content = c
			}
		}

		// 优化角色显示
		roleText := role
		if role == "system" {
			roleText = "系统"
		} else if role == "user" {
			roleText = "用户"
		} else if role == "assistant" {
			roleText = "助手"
		}

		// 截断过长的内容
		truncatedContent := content
		if len(content) > 300 {
			truncatedContent = content[:300] + "..."
		}

		prompt += fmt.Sprintf("%d. [%s] %s\n", i+1, roleText, truncatedContent)
	}

	prompt += `

## 需要包含的关键信息：
- 对话的主要主题
- 关键问题和答案
- 决策和行动项目
- 重要的技术术语或概念
- 参与者的主要需求或目标

请确保摘要覆盖上述所有要点，并保持在指定的 token 限制内。`

	return prompt
}

// getMessageScore 计算消息重要性分数
func (m *MessageCompressor) getMessageScore(msg interface{}) float64 {
	score := 1.0

	switch m := msg.(type) {
	case map[string]interface{}:
		if role, ok := m["role"]; ok {
			switch role {
			case "system":
				score = 10.0 // 最高优先级
			case "assistant":
				score = 5.0
			case "user":
				score = 3.0
			}
		}
		// 检查是否有工具调用（更高优先级）
		if _, hasTool := m["tool_calls"]; hasTool {
			score *= 1.5
		}
	case map[string]string:
		if role, ok := m["role"]; ok {
			switch role {
			case "system":
				score = 10.0
			case "assistant":
				score = 5.0
			case "user":
				score = 3.0
			}
		}
	}

	return score
}

// selectMessagesByScore 根据分数选择消息
func (m *MessageCompressor) selectMessagesByScore(messages []interface{}, scores []float64, targetTokens int) []interface{} {
	type indexedMessage struct {
		index int
		score float64
		msg   interface{}
	}

	// 创建索引消息列表
	indexed := make([]indexedMessage, len(messages))
	for i, msg := range messages {
		indexed[i] = indexedMessage{index: i, score: scores[i], msg: msg}
	}

	// 按分数排序（简单选择排序）
	for i := 0; i < len(indexed)-1; i++ {
		for j := i + 1; j < len(indexed); j++ {
			if indexed[j].score > indexed[i].score {
				indexed[i], indexed[j] = indexed[j], indexed[i]
			}
		}
	}

	// 选择消息直到达到目标 Token 数量
	result := make([]interface{}, 0)
	currentTokens := 0

	for _, im := range indexed {
		msgTokens := m.counter.CountMessage(im.msg)
		if currentTokens+msgTokens > targetTokens {
			continue
		}

		currentTokens += msgTokens
		result = append(result, im.msg)
	}

	return result
}

// isSystemMessage 检查是否为系统消息
func isSystemMessage(msg interface{}) bool {
	switch m := msg.(type) {
	case map[string]interface{}:
		if role, ok := m["role"]; ok {
			return role == "system"
		}
	case map[string]string:
		if role, ok := m["role"]; ok {
			return role == "system"
		}
	}
	return false
}

// updateStats 更新压缩统计信息
func (m *MessageCompressor) updateStats(input, output []interface{}, strategy CompressionStrategy) {
	inputTokens := m.counter.CountMessages(input)
	outputTokens := m.counter.CountMessages(output)

	m.stats.TotalCompressions++
	m.stats.TotalInputTokens += int64(inputTokens)
	m.stats.TotalOutputTokens += int64(outputTokens)
	m.stats.AverageRatio = float64(m.stats.TotalOutputTokens) / float64(m.stats.TotalInputTokens)

	m.stats.LastCompression = &CompressionOperation{
		Strategy:        strategy,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		CompressionRatio: float64(outputTokens) / float64(inputTokens),
		Success:        true,
	}
}
