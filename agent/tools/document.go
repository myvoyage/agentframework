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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent"
)

// DocumentGeneratorSkill 文档生成技能，用于生成各种格式的文档
type DocumentGeneratorSkill struct {
	name        string
	description string
}

// NewDocumentGeneratorSkill 创建一个新的文档生成技能实例
func NewDocumentGeneratorSkill() agent.Skill {
	return &DocumentGeneratorSkill{
		name:        "document_generator",
		description: "Generate various formats of documents (markdown, html, etc.)",
	}
}

// Info 返回技能的元信息
func (s *DocumentGeneratorSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.name,
		Desc: s.description,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     "string",
				Desc:     "Document format: markdown, html, text",
				Required: true,
			},
			"title": {
				Type:     "string",
				Desc:     "Document title",
				Required: true,
			},
			"content": {
				Type:     "string",
				Desc:     "Document content",
				Required: true,
			},
			"metadata": {
				Type:     "object",
				Desc:     "Document metadata (optional)",
				Required: false,
			},
		}),
	}, nil
}

// Invoke 执行技能，生成文档
func (s *DocumentGeneratorSkill) Invoke(ctx context.Context, input string) (string, error) {
	// 解析输入参数
	var params struct {
		Format   string                 `json:"format"`
		Title    string                 `json:"title"`
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证参数
	if params.Format == "" {
		return "", errors.New("format is required")
	}

	if params.Title == "" {
		return "", errors.New("title is required")
	}

	if params.Content == "" {
		return "", errors.New("content is required")
	}

	// 生成文档
	var document string
	var err error

	switch strings.ToLower(params.Format) {
	case "markdown", "md":
		document, err = s.generateMarkdown(params.Title, params.Content, params.Metadata)
	case "html":
		document, err = s.generateHTML(params.Title, params.Content, params.Metadata)
	case "text":
		document, err = s.generateText(params.Title, params.Content, params.Metadata)
	default:
		return "", fmt.Errorf("unsupported format: %s", params.Format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate document: %w", err)
	}

	// 构建返回结果
	result := map[string]interface{}{
		"success": true,
		"format":  params.Format,
		"title":   params.Title,
		"content": document,
		"length":  len(document),
	}

	// 添加元数据
	if params.Metadata != nil {
		result["metadata"] = params.Metadata
	}

	// 转换为JSON字符串
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// IsEnabled 检查技能是否启用
func (s *DocumentGeneratorSkill) IsEnabled(ctx context.Context) bool {
	return true
}

// GetMetadata 获取技能的元数据
func (s *DocumentGeneratorSkill) GetMetadata(ctx context.Context) agent.SkillMetadata {
	return agent.SkillMetadata{
		Name:        s.name,
		Version:     "1.0.0",
		Author:      "Agent Framework Contributors",
		Description: s.description,
		Category:    "document",
		Tags:        []string{"document", "generate", "markdown", "html", "text"},
		License:     "AGPL-3.0-or-later",
		Repository:  "https://github.com/AgentFramework/agentframework",
		Keywords:    []string{"document", "generator", "markdown", "html"},
	}
}

// generateMarkdown 生成Markdown格式文档
func (s *DocumentGeneratorSkill) generateMarkdown(title, content string, metadata map[string]interface{}) (string, error) {
	var sb strings.Builder

	// 添加标题
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	// 添加元数据
	if metadata != nil {
		sb.WriteString("## Metadata\n\n")
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	// 添加内容
	sb.WriteString("## Content\n\n")
	sb.WriteString(content)
	sb.WriteString("\n")

	return sb.String(), nil
}

// generateHTML 生成HTML格式文档
func (s *DocumentGeneratorSkill) generateHTML(title, content string, metadata map[string]interface{}) (string, error) {
	var sb strings.Builder

	// 添加HTML头部
	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString(fmt.Sprintf("<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>%s</title>\n    <style>\n        body {\n            font-family: Arial, sans-serif;\n            margin: 0;\n            padding: 20px;\n            line-height: 1.6;\n        }\n        h1 {\n            color: #333;\n            border-bottom: 1px solid #eee;\n            padding-bottom: 10px;\n        }\n        h2 {\n            color: #555;\n        }\n        .metadata {\n            background-color: #f5f5f5;\n            padding: 15px;\n            border-radius: 5px;\n            margin-bottom: 20px;\n        }\n        .content {\n            background-color: #fff;\n            padding: 20px;\n            border-radius: 5px;\n            box-shadow: 0 2px 4px rgba(0,0,0,0.1);\n        }\n    </style>\n</head>\n<body>\n", title))

	// 添加标题
	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", title))

	// 添加元数据
	if metadata != nil {
		sb.WriteString("<div class=\"metadata\">\n<h2>Metadata</h2>\n<ul>\n")
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf("    <li><strong>%s:</strong> %v</li>\n", key, value))
		}
		sb.WriteString("</ul>\n</div>\n")
	}

	// 添加内容
	sb.WriteString("<div class=\"content\">\n")
	sb.WriteString(fmt.Sprintf("<h2>Content</h2>\n<p>%s</p>\n", content))
	sb.WriteString("</div>\n</body>\n</html>")

	return sb.String(), nil
}

// generateText 生成纯文本格式文档
func (s *DocumentGeneratorSkill) generateText(title, content string, metadata map[string]interface{}) (string, error) {
	var sb strings.Builder

	// 添加标题
	sb.WriteString(fmt.Sprintf("%s\n", title))
	sb.WriteString(strings.Repeat("=", len(title)))
	sb.WriteString("\n\n")

	// 添加元数据
	if metadata != nil {
		sb.WriteString("Metadata:\n")
		sb.WriteString("-----------\n")
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	// 添加内容
	sb.WriteString("Content:\n")
	sb.WriteString("-----------\n")
	sb.WriteString(content)
	sb.WriteString("\n")

	return sb.String(), nil
}
