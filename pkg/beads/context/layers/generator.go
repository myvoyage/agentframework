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

package layers

import (
	stdcontext "context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"AgentFramework/pkg/beads/context"
)

// LayerGenerator 层级生成器实现
type LayerGenerator struct {
	l0Generator *L0Generator
	l1Generator *L1Generator
	l2Processor *L2Processor
	tokenCounter TokenCounter
}

// NewLayerGenerator 创建新的层级生成器
func NewLayerGenerator() *LayerGenerator {
	return &LayerGenerator{
		l0Generator: NewL0Generator(),
		l1Generator: NewL1Generator(),
		l2Processor: NewL2Processor(),
		tokenCounter: NewSimpleTokenCounter(),
	}
}

// GenerateL0 生成 L0 摘要层
func (lg *LayerGenerator) GenerateL0(ctx stdcontext.Context, content string) (*context.LayerSummary, error) {
	return lg.l0Generator.Generate(ctx, content)
}

// GenerateL1 生成 L1 概览层
func (lg *LayerGenerator) GenerateL1(ctx stdcontext.Context, content string) (*context.LayerOverview, error) {
	return lg.l1Generator.Generate(ctx, content)
}

// GenerateL2 生成 L2 详情层
func (lg *LayerGenerator) GenerateL2(ctx stdcontext.Context, content string, format string) (*context.LayerDetails, error) {
	return lg.l2Processor.Process(ctx, content, format)
}

// GenerateAll 生成所有层级
func (lg *LayerGenerator) GenerateAll(ctx stdcontext.Context, content string, format string) (*context.ContextLayers, error) {
	layers := &context.ContextLayers{}

	// 生成 L0
	l0, err := lg.GenerateL0(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("generate L0: %w", err)
	}
	layers.L0 = l0

	// 生成 L1
	l1, err := lg.GenerateL1(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("generate L1: %w", err)
	}
	layers.L1 = l1

	// 处理 L2
	l2, err := lg.GenerateL2(ctx, content, format)
	if err != nil {
		return nil, fmt.Errorf("generate L2: %w", err)
	}
	layers.L2 = l2

	return layers, nil
}

// ===== L0 生成器 =====

// L0Generator L0 摘要层生成器
type L0Generator struct {
	maxTokens int
}

// NewL0Generator 创建新的 L0 生成器
func NewL0Generator() *L0Generator {
	return &L0Generator{
		maxTokens: 100,
	}
}

// Generate 生成 L0 摘要
func (g *L0Generator) Generate(ctx stdcontext.Context, content string) (*context.LayerSummary, error) {
	// 清理内容
	content = strings.TrimSpace(content)
	if content == "" {
		return &context.LayerSummary{
			Content:     "",
			Tokens:      0,
			GeneratedAt: time.Now(),
			Method:      "empty",
		}, nil
	}

	// 生成摘要
	summary := g.generateSummary(content)

	return &context.LayerSummary{
		Content:     summary,
		Tokens:      len(summary) / 4, // 简单估算
		GeneratedAt: time.Now(),
		Method:      "template",
	}, nil
}

// generateSummary 生成摘要（模板方法）
func (g *L0Generator) generateSummary(content string) string {
	// 移除多余空白
	content = strings.Join(strings.Fields(content), " ")

	// 如果内容够短，直接返回
	if len(content) <= 400 {
		return content
	}

	// 按句子分割
	sentences := g.splitSentences(content)

	// 选择前几句
	wordCount := 0
	var summary strings.Builder
	for _, sentence := range sentences {
		words := len(strings.Fields(sentence))
		if wordCount+words > 100 {
			break
		}
		summary.WriteString(sentence)
		summary.WriteString(" ")
		wordCount += words
	}

	return strings.TrimSpace(summary.String())
}

// splitSentences 分割句子
func (g *L0Generator) splitSentences(content string) []string {
	// 简单的句子分割（按句号、问号、感叹号）
	re := regexp.MustCompile(`[.!?]+\s+`)
	sentences := re.Split(content, -1)

	var result []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

// ===== L1 生成器 =====

// L1Generator L1 概览层生成器
type L1Generator struct {
	maxTokens int
}

// NewL1Generator 创建新的 L1 生成器
func NewL1Generator() *L1Generator {
	return &L1Generator{
		maxTokens: 2000,
	}
}

// Generate 生成 L1 概览
func (g *L1Generator) Generate(ctx stdcontext.Context, content string) (*context.LayerOverview, error) {
	// 清理内容
	content = strings.TrimSpace(content)
	if content == "" {
		return &context.LayerOverview{
			Content:     "",
			Tokens:      0,
			Sections:    []string{},
			KeyPoints:   []string{},
			GeneratedAt: time.Now(),
			Method:      "empty",
		}, nil
	}

	// 分析内容结构
	sections := g.extractSections(content)
	keyPoints := g.extractKeyPoints(content)

	// 生成概览内容
	overview := g.generateOverview(content, sections)

	return &context.LayerOverview{
		Content:     overview,
		Tokens:      len(overview) / 4,
		Sections:    sections,
		KeyPoints:   keyPoints,
		GeneratedAt: time.Now(),
		Method:      "template",
	}, nil
}

// extractSections 提取章节
func (g *L1Generator) extractSections(content string) []string {
	var sections []string

	// 检测 Markdown 标题
	re := regexp.MustCompile(`^#{1,6}\s+(.+)$`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if re.MatchString(line) {
			// 移除 # 符号
			section := strings.TrimLeft(line, "#")
			section = strings.TrimSpace(section)
			sections = append(sections, section)
		}
	}

	// 如果没有找到章节，尝试其他方式
	if len(sections) == 0 {
		// 按段落分割
		paragraphs := strings.Split(content, "\n\n")
		for i, p := range paragraphs {
			p = strings.TrimSpace(p)
			if p != "" && len(p) > 50 {
				sections = append(sections, fmt.Sprintf("段落 %d", i+1))
			}
		}
	}

	return sections
}

// extractKeyPoints 提取关键点
func (g *L1Generator) extractKeyPoints(content string) []string {
	var keyPoints []string

	// 检测列表项
	re := regexp.MustCompile(`^[-*+]\s+(.+)$`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				point := strings.TrimSpace(matches[1])
				if len(point) > 10 && len(point) < 200 {
					keyPoints = append(keyPoints, point)
				}
			}
		}
	}

	// 如果没有找到列表，从段落中提取
	if len(keyPoints) == 0 {
		paragraphs := strings.Split(content, "\n\n")
		for _, p := range paragraphs {
			p = strings.TrimSpace(p)
			// 选择包含重要词汇的段落
			if g.containsImportantWords(p) && len(p) < 300 {
				keyPoints = append(keyPoints, p)
			}
		}
	}

	// 限制数量
	if len(keyPoints) > 10 {
		keyPoints = keyPoints[:10]
	}

	return keyPoints
}

// containsImportantWords 检查是否包含重要词汇
func (g *L1Generator) containsImportantWords(text string) bool {
	importantWords := []string{
		"重要", "关键", "核心", "主要", "必须", "应该",
		"注意", "实现", "功能", "方法", "系统",
		"import", "important", "key", "main", "note",
	}

	textLower := strings.ToLower(text)
	for _, word := range importantWords {
		if strings.Contains(textLower, strings.ToLower(word)) {
			return true
		}
	}

	return false
}

// generateOverview 生成概览内容
func (g *L1Generator) generateOverview(content string, sections []string) string {
	var overview strings.Builder

	// 添加章节列表
	if len(sections) > 0 {
		overview.WriteString("## 章节\n\n")
		for _, section := range sections {
			overview.WriteString(fmt.Sprintf("- **%s**\n", section))
		}
		overview.WriteString("\n")
	}

	// 添加内容摘要
	overview.WriteString("## 内容摘要\n\n")

	// 截取适当长度的内容
	maxLength := 1500
	if len(content) <= maxLength {
		overview.WriteString(content)
	} else {
		overview.WriteString(content[:maxLength])
		overview.WriteString("...")
	}

	return overview.String()
}

// ===== L2 处理器 =====

// L2Processor L2 详情层处理器
type L2Processor struct {
	maxTokens int
}

// NewL2Processor 创建新的 L2 处理器
func NewL2Processor() *L2Processor {
	return &L2Processor{
		maxTokens: 100000, // 无限制
	}
}

// Process 处理 L2 详情
func (p *L2Processor) Process(ctx stdcontext.Context, content string, format string) (*context.LayerDetails, error) {
	// 清理内容
	content = strings.TrimSpace(content)

	// 自动检测格式（如果未指定）
	if format == "" {
		format = p.detectFormat(content)
	}

	return &context.LayerDetails{
		Content:     content,
		Tokens:      len(content) / 4,
		Format:      format,
		Source:      "input",
		GeneratedAt: time.Now(),
		Metadata: map[string]string{
			"size": fmt.Sprintf("%d", len(content)),
		},
	}, nil
}

// detectFormat 自动检测内容格式
func (p *L2Processor) detectFormat(content string) string {
	// 检查是否为 Markdown
	if p.isMarkdown(content) {
		return "markdown"
	}

	// 检查是否为 JSON
	if p.isJSON(content) {
		return "json"
	}

	// 检查是否为 HTML
	if p.isHTML(content) {
		return "html"
	}

	// 检查是否为代码
	if p.isCode(content) {
		return "code"
	}

	return "plain"
}

// isMarkdown 检查是否为 Markdown
func (p *L2Processor) isMarkdown(content string) bool {
	markdownIndicators := []string{
		"# ", "## ", "### ", // 标题
		"**", "__", // 粗体
		"* ", "- ", // 列表
		"`", "```", // 代码
		"[", "](", // 链接
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range markdownIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}

	return false
}

// isJSON 检查是否为 JSON
func (p *L2Processor) isJSON(content string) bool {
	content = strings.TrimSpace(content)
	return strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[")
}

// isHTML 检查是否为 HTML
func (p *L2Processor) isHTML(content string) bool {
	contentLower := strings.ToLower(content)
	return strings.Contains(contentLower, "<html") ||
		strings.Contains(contentLower, "<div") ||
		strings.Contains(contentLower, "<span")
}

// isCode 检查是否为代码
func (p *L2Processor) isCode(content string) bool {
	// 检查常见代码特征
	codeIndicators := []string{
		"function ", "class ", "def ", "func ",
		"import ", "require(", "use ",
		"if ", "for ", "while ",
		"return ", "yield ",
		"{", "}", "(", ");",
	}

	for _, indicator := range codeIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}

	return false
}

// ===== Token 计数器 =====

// TokenCounter Token 计数器接口
type TokenCounter interface {
	CountTokens(text string) int
	CountTokensApprox(text string) int
}

// SimpleTokenCounter 简单的 Token 计数器实现
type SimpleTokenCounter struct{}

// NewSimpleTokenCounter 创建新的简单计数器
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{}
}

// CountTokens 计算 Token 数量（简化版）
func (c *SimpleTokenCounter) CountTokens(text string) int {
	// 简单估算：英文约 4 字符/token，中文约 1.5 字符/token
	englishChars := 0
	chineseChars := 0

	for _, r := range text {
		if r < 128 {
			// ASCII 字符
			englishChars++
		} else {
			// 非ASCII 字符（主要是中文）
			chineseChars++
		}
	}

	englishTokens := englishChars / 4
	chineseTokens := chineseChars * 2 / 3

	return englishTokens + chineseTokens
}

// CountTokensApprox 近似计算 Token 数量
func (c *SimpleTokenCounter) CountTokensApprox(text string) int {
	// 更简单的估算：总字符数 / 3
	return len(text) / 3
}

// GenerateL2 生成 L2 详情层（别名）
func (p *L2Processor) GenerateL2(ctx stdcontext.Context, content string, format string) (*context.LayerDetails, error) {
	return p.Process(ctx, content, format)
}
