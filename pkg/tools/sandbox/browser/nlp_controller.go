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
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ChatModel LLM模型接口（简化版，用于跨包使用）
type ChatModel interface {
	Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

// NLPBrowserController NLP浏览器控制器，支持自然语言控制浏览器
type NLPBrowserController struct {
	module       *BrowserModule
	model        ChatModel
	intentCache  map[string]*BrowserIntent
	mu           sync.RWMutex
}

// BrowserIntent 浏览器意图
type BrowserIntent struct {
	Action     string                 `json:"action"`     // click, type, navigate, extract, scroll, wait
	Target     string                 `json:"target"`     // 目标元素描述
	Selector   string                 `json:"selector"`   // CSS选择器（可选）
	Value      string                 `json:"value"`      // 输入值（可选）
	Options    map[string]interface{} `json:"options"`    // 额外选项
	Confidence float64                `json:"confidence"` // 置信度
	Reasoning  string                 `json:"reasoning"`  // 推理过程
}

// NLPControllerConfig NLP控制器配置
type NLPControllerConfig struct {
	Model        ChatModel
	CacheIntents bool
}

// NewNLPBrowserController 创建NLP浏览器控制器
func NewNLPBrowserController(module *BrowserModule, config NLPControllerConfig) (*NLPBrowserController, error) {
	if config.Model == nil {
		return nil, fmt.Errorf("model is required for NLP control")
	}

	return &NLPBrowserController{
		module:      module,
		model:       config.Model,
		intentCache: make(map[string]*BrowserIntent),
	}, nil
}

// ExecuteNaturalLanguage 执行自然语言命令
func (c *NLPBrowserController) ExecuteNaturalLanguage(ctx context.Context, command string) (string, error) {
	intent, err := c.ParseIntent(ctx, command)
	if err != nil {
		return "", fmt.Errorf("failed to parse intent: %w", err)
	}

	return c.ExecuteIntent(ctx, intent)
}

// ParseIntent 解析自然语言命令为浏览器意图
func (c *NLPBrowserController) ParseIntent(ctx context.Context, command string) (*BrowserIntent, error) {
	if intent, ok := c.intentCache[command]; ok {
		return intent, nil
	}

	prompt := c.buildIntentPrompt(command)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: c.getSystemPrompt(),
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := c.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	intent, err := c.parseIntentResponse(resp.Content)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.intentCache[command] = intent
	c.mu.Unlock()

	return intent, nil
}

// ExecuteIntent 执行浏览器意图
func (c *NLPBrowserController) ExecuteIntent(ctx context.Context, intent *BrowserIntent) (string, error) {
	switch intent.Action {
	case "navigate":
		return c.executeNavigate(ctx, intent)
	case "click":
		return c.executeClick(ctx, intent)
	case "type":
		return c.executeType(ctx, intent)
	case "scroll":
		return c.executeScroll(ctx, intent)
	case "extract":
		return c.executeExtract(ctx, intent)
	case "wait":
		return c.executeWait(ctx, intent)
	case "screenshot":
		return c.executeScreenshot(ctx, intent)
	default:
		return "", fmt.Errorf("unknown action: %s", intent.Action)
	}
}

// getSystemPrompt 获取系统提示词
func (c *NLPBrowserController) getSystemPrompt() string {
	return `你是一个浏览器控制专家。将自然语言命令解析为结构化的浏览器操作意图。

支持的操作：
1. navigate - 导航到URL ("打开google.com")
2. click - 点击元素 ("点击登录按钮")
3. type - 输入文本 ("在搜索框输入 'AI'")
4. scroll - 滚动页面 ("向下滚动")
5. extract - 提取内容 ("提取页面标题")
6. wait - 等待 ("等待页面加载")
7. screenshot - 截图 ("截图保存")

返回JSON格式：
{
  "action": "操作类型",
  "target": "目标描述",
  "selector": "CSS选择器（可选）",
  "value": "输入值（可选）",
  "confidence": 0.95,
  "reasoning": "推理过程"
}`
}

// buildIntentPrompt 构建意图解析提示词
func (c *NLPBrowserController) buildIntentPrompt(command string) string {
	return fmt.Sprintf(`# 浏览器控制命令解析

用户命令: %s

请分析命令并返回浏览器操作意图。`, command)
}

// parseIntentResponse 解析意图响应
func (c *NLPBrowserController) parseIntentResponse(content string) (*BrowserIntent, error) {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	jsonStart := strings.Index(content, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var intent BrowserIntent
	if err := json.Unmarshal([]byte(content[jsonStart:]), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
	}

	return &intent, nil
}

// executeNavigate 执行导航
func (c *NLPBrowserController) executeNavigate(ctx context.Context, intent *BrowserIntent) (string, error) {
	url := intent.Target
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	_, err := c.module.navigate(url, 0)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已导航到: %s", url), nil
}

// executeClick 执行点击
func (c *NLPBrowserController) executeClick(ctx context.Context, intent *BrowserIntent) (string, error) {
	selector := intent.Selector
	if selector == "" {
		selector = c.generateSimpleSelector(intent.Target)
	}

	_, err := c.module.click(selector)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已点击: %s (选择器: %s)", intent.Target, selector), nil
}

// executeType 执行输入
func (c *NLPBrowserController) executeType(ctx context.Context, intent *BrowserIntent) (string, error) {
	selector := intent.Selector
	if selector == "" {
		selector = c.generateSimpleSelector(intent.Target)
	}

	value := intent.Value
	_, err := c.module.input(selector, value)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已在 %s 输入: %s", selector, value), nil
}

// executeScroll 执行滚动
func (c *NLPBrowserController) executeScroll(ctx context.Context, intent *BrowserIntent) (string, error) {
	direction := 300 // 默认向下滚动300像素
	if intent.Target == "up" || intent.Target == "向上" {
		direction = -300
	}

	script := fmt.Sprintf("window.scrollBy(0, %d);", direction)
	_, err := c.module.executeJS(script)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已滚动页面"), nil
}

// executeExtract 执行内容提取
func (c *NLPBrowserController) executeExtract(ctx context.Context, intent *BrowserIntent) (string, error) {
	if intent.Selector != "" {
		result, err := c.module.getText(intent.Selector)
		if err != nil {
			return "", err
		}
		// 提取文本内容
		if text, ok := result["text"].(string); ok {
			return text, nil
		}
		// 转换为JSON字符串
		jsonBytes, _ := json.Marshal(result)
		return string(jsonBytes), nil
	}

	// 获取整个页面文本（使用body选择器）
	result, err := c.module.getText("body")
	if err != nil {
		return "", err
	}

	// 提取文本内容
	if text, ok := result["text"].(string); ok {
		// 限制返回长度
		if len(text) > 2000 {
			text = text[:2000] + "..."
		}
		return text, nil
	}

	// 转换为JSON字符串
	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes), nil
}

// executeWait 执行等待
func (c *NLPBrowserController) executeWait(ctx context.Context, intent *BrowserIntent) (string, error) {
	return "已完成等待", nil
}

// executeScreenshot 执行截图
func (c *NLPBrowserController) executeScreenshot(ctx context.Context, intent *BrowserIntent) (string, error) {
	_, err := c.module.screenshot("", "", false)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已截图"), nil
}

// generateSimpleSelector 生成简单选择器
func (c *NLPBrowserController) generateSimpleSelector(description string) string {
	desc := strings.ToLower(description)

	// 按钮选择器
	if strings.Contains(desc, "按钮") || strings.Contains(desc, "button") {
		if strings.Contains(desc, "提交") || strings.Contains(desc, "submit") {
			return `button[type="submit"]`
		}
		return "button"
	}

	// 输入框选择器
	if strings.Contains(desc, "输入") || strings.Contains(desc, "input") || strings.Contains(desc, "框") {
		if strings.Contains(desc, "用户") || strings.Contains(desc, "user") {
			return `input[name="username"]`
		}
		if strings.Contains(desc, "密码") || strings.Contains(desc, "password") {
			return `input[type="password"]`
		}
		if strings.Contains(desc, "搜索") || strings.Contains(desc, "search") {
			return `input[name="search"]`
		}
		return "input"
	}

	// 链接选择器
	if strings.Contains(desc, "链接") || strings.Contains(desc, "link") {
		return "a"
	}

	return "*"
}

// SmartExecute 智能执行，支持多步骤命令
func (c *NLPBrowserController) SmartExecute(ctx context.Context, command string) (string, error) {
	intent, err := c.ParseIntent(ctx, command)
	if err == nil {
		return c.ExecuteIntent(ctx, intent)
	}

	steps, err := c.decomposeCommand(ctx, command)
	if err != nil {
		return "", err
	}

	var results []string
	for i, step := range steps {
		result, err := c.ExecuteIntent(ctx, step)
		if err != nil {
			return "", fmt.Errorf("step %d failed: %w", i+1, err)
		}
		results = append(results, result)
	}

	return strings.Join(results, "\n"), nil
}

// decomposeCommand 分解复杂命令
func (c *NLPBrowserController) decomposeCommand(ctx context.Context, command string) ([]*BrowserIntent, error) {
	prompt := fmt.Sprintf(`将浏览器命令分解为步骤。

命令: %s

返回JSON：
{
  "steps": [
    {"action": "navigate", "target": "https://example.com", "reasoning": "访问网站"},
    {"action": "type", "target": "搜索框", "value": "AI", "reasoning": "输入搜索"}
  ]
}`, command)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是浏览器命令分解专家。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := c.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result struct {
		Steps []*BrowserIntent `json:"steps"`
	}

	content := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	jsonStart := strings.Index(content, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON found")
	}

	if err := json.Unmarshal([]byte(content[jsonStart:]), &result); err != nil {
		return nil, fmt.Errorf("failed to parse steps: %w", err)
	}

	return result.Steps, nil
}

// ExtractAction 从命令中提取动作
func ExtractAction(command string) string {
	actions := []string{
		"打开", "访问", "导航", "navigate",
		"点击", "单击", "click",
		"输入", "填写", "type",
		"滚动", "scroll",
		"提取", "获取", "extract",
	}

	command = strings.ToLower(command)
	for _, action := range actions {
		if strings.Contains(command, action) {
			return action
		}
	}

	return ""
}

// ExtractValue 从命令中提取输入值
func ExtractValue(command string) string {
	re := regexp.MustCompile(`["“']([^"”']+)["“']`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}

	if idx := strings.Index(command, ":"); idx != -1 && idx < len(command)-1 {
		value := strings.TrimSpace(command[idx+1:])
		if value != "" {
			return value
		}
	}

	return ""
}

// ClearCache 清除意图缓存
func (c *NLPBrowserController) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.intentCache = make(map[string]*BrowserIntent)
}

// GetCacheSize 获取缓存大小
func (c *NLPBrowserController) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.intentCache)
}
