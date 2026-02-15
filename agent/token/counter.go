// Agent Framework - Token Counter and Compression
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package token

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TokenCounter Token 计数器接口
type TokenCounter interface {
	// CountText 计算文本的 Token 数量
	CountText(text string) int
	// CountMessage 计算 schema.Message 的 Token 数量
	CountMessage(msg interface{}) int
	// CountMessages 计算消息列表的总 Token 数量
	CountMessages(msgs []interface{}) int
	// EstimateCount 估算 Token 数量（快速但不精确）
	EstimateCount(text string) int
}

// CompressionStrategy 压缩策略类型
type CompressionStrategy string

const (
	// StrategyTruncate 截断策略：保留最新的 N 条消息
	StrategyTruncate CompressionStrategy = "truncate"
	// StrategySummarize LLM摘要策略：使用 LLM 生成摘要
	StrategySummarize CompressionStrategy = "summarize"
	// StrategySemantic 语义压缩策略：按重要性保留消息
	StrategySemantic CompressionStrategy = "semantic"
	// StrategyHybrid 混合策略：结合多种策略
	StrategyHybrid CompressionStrategy = "hybrid"
)

// DefaultTokenCounter 默认的 Token 计数器实现
// 基于 GPT tokenizer 的简化估算
type DefaultTokenCounter struct {
	// Token 预估因子
	// 英文约 4 字符/token，中文约 2.5 字符/token
	englishFactor float64
	chineseFactor float64
}

// NewDefaultTokenCounter 创建默认 Token 计数器
func NewDefaultTokenCounter() *DefaultTokenCounter {
	return &DefaultTokenCounter{
		englishFactor: 4.0,  // 英文约 4 字符/token
		chineseFactor: 2.5,  // 中文约 2.5 字符/token
	}
}

// CountText 计算文本的 Token 数量
func (c *DefaultTokenCounter) CountText(text string) int {
	if text == "" {
		return 0
	}

	// 快速路径：空字符串
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	totalChars := utf8.RuneCountInString(text)

	// 检测中文字符数量
	chineseChars := countChineseCharacters(text)
	englishChars := totalChars - chineseChars

	// 计算预估 Token 数量
	tokens := float64(englishChars)/c.englishFactor + float64(chineseChars)/c.chineseFactor

	return int(tokens) + 1 // +1 用于缓冲
}

// EstimateCount 快速估算 Token 数量
func (c *DefaultTokenCounter) EstimateCount(text string) int {
	if text == "" {
		return 0
	}

	// 更粗略的估算：假设平均 3 字符/token
	return utf8.RuneCountInString(text)/3 + 1
}

// CountMessage 计算 schema.Message 的 Token 数量
func (c *DefaultTokenCounter) CountMessage(msg interface{}) int {
	// 使用类型断言处理不同类型的消息结构
	if msg == nil {
		return 0
	}

	// 尝试获取消息内容的反射处理
	var content string
	var role string

	// 通过 map 访问可能的结构
	switch m := msg.(type) {
	case map[string]interface{}:
		if val, ok := m["content"]; ok {
			content = fmt.Sprintf("%v", val)
		}
		if val, ok := m["role"]; ok {
			role = fmt.Sprintf("%v", val)
		}
	case map[string]string:
		content = m["content"]
		role = m["role"]
	}

	tokens := c.CountText(content)

	// 角色信息也占用少量 Token
	if role != "" {
		tokens += len(strings.Fields(role)) + 2
	}

	return tokens
}

// CountMessages 计算消息列表的总 Token 数量
func (c *DefaultTokenCounter) CountMessages(msgs []interface{}) int {
	total := 0
	for _, msg := range msgs {
		total += c.CountMessage(msg)
	}
	return total
}

// countChineseCharacters 统计中文字符数量
func countChineseCharacters(text string) int {
	// 中文字符范围：\u4e00-\u9fff (CJK统一汉字)
	// 包括扩展区：\u3400-\u4dbf, \u20000-\u2a6df, \u2a700-\u2b73f
	chineseRegex := regexp.MustCompile(`[\p{Han}]`)
	matches := chineseRegex.FindAllString(text, -1)
	return len(matches)
}

// ===== Token 估算辅助函数 =====

// EstimateMessagesTokens 估算消息列表的 Token 数量
func EstimateMessagesTokens(messages []interface{}) int {
	counter := NewDefaultTokenCounter()
	return counter.CountMessages(messages)
}

// EstimateTextTokens 估算文本的 Token 数量
func EstimateTextTokens(text string) int {
	counter := NewDefaultTokenCounter()
	return counter.CountText(text)
}

// GetTokensFromText 根据 Token 数量分割文本
// 返回不超过 targetTokens 的文本切片
func GetTokensFromText(text string, targetTokens int) []string {
	counter := NewDefaultTokenCounter()

	// 如果文本 Token 数量小于目标，返回完整文本
	if counter.CountText(text) <= targetTokens {
		return []string{text}
	}

	// 简单按字符分割策略
	// TODO: 改进为按句子或段落分割
	tokens := []string{}
	remaining := text
	targetChars := int(float64(targetTokens) * 3.5) // 估算字符数

	for utf8.RuneCountInString(remaining) > targetChars {
		if len(remaining) <= targetChars {
			tokens = append(tokens, remaining)
			break
		}
		// 在标点符号处分割
		splitPos := findSplitPoint(remaining, targetChars)
		tokens = append(tokens, remaining[:splitPos])
		remaining = strings.TrimSpace(remaining[splitPos:])
	}

	if len(remaining) > 0 {
		tokens = append(tokens, remaining)
	}

	return tokens
}

// findSplitPoint 在指定长度附近找到合适的分割点
func findSplitPoint(text string, targetLen int) int {
	// 优先在句号、问号、感叹号处分割
	sentenceEnds := []string{"。", "！", "？", ".", "!", "?"}

	for i, char := range text {
		if i > targetLen {
			break
		}
		for _, end := range sentenceEnds {
			if string(char) == end {
				// 检查下一个字符是否也是结束符
				if i+1 < len(text) && text[i+1] == ' ' {
					return i + 2
				}
				return i + 1
			}
		}
	}

	// 如果没找到句号，在逗号处分割
	commaPos := strings.IndexAny(text, ",，、;；")
	if commaPos > 0 && commaPos < targetLen {
		return commaPos + 1
	}

	// 最后在空格处分割
	spacePos := strings.LastIndex(text[:min(len(text), targetLen)], " ")
	if spacePos > 0 {
		return spacePos + 1
	}

	// 强制分割
	return min(len(text), targetLen)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
