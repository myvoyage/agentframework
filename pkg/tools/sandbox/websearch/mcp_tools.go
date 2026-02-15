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

package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ==================== WebSearchTool ====================

// webSearchTool 网络搜索工具
type webSearchTool struct {
	module *WebSearchModule
}

// NewWebSearchTool 创建网络搜索工具
func NewWebSearchTool(module *WebSearchModule) tool.BaseTool {
	return &webSearchTool{module: module}
}

// Info 返回工具信息
func (t *webSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "在互联网上搜索信息，支持多个搜索引擎（Google、Bing、DuckDuckGo）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索查询内容",
				Required: true,
			},
			"num_results": {
				Type:     "integer",
				Desc:     "返回结果数量，默认10条",
				Required: false,
			},
			"language": {
				Type:     "string",
				Desc:     "搜索语言，如 'zh-CN'、'en-US'，默认为系统语言",
				Required: false,
			},
			"safe_search": {
				Type:     "boolean",
				Desc:     "是否启用安全搜索（过滤成人内容），默认false",
				Required: false,
			},
			"time_range": {
				Type:     "string",
				Desc:     "时间范围过滤: 'past_day'、'past_week'、'past_month'、'past_year'",
				Required: false,
			},
			"site_filter": {
				Type:     "string",
				Desc:     "限定搜索特定网站，如 'wikipedia.org'",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (t *webSearchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results"`
		Language   string `json:"language"`
		SafeSearch bool   `json:"safe_search"`
		TimeRange  string `json:"time_range"`
		SiteFilter string `json:"site_filter"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	// 构建搜索选项
	searchOpts := &SearchOptions{
		NumResults: 10,
		SafeSearch: false,
		Language:   args.Language,
	}

	if args.NumResults > 0 {
		searchOpts.NumResults = args.NumResults
		if searchOpts.NumResults > 50 {
			searchOpts.NumResults = 50
		}
	}

	if args.SafeSearch {
		searchOpts.SafeSearch = true
	}

	if args.TimeRange != "" {
		validRanges := map[string]bool{
			"past_day":   true,
			"past_week":  true,
			"past_month": true,
			"past_year":  true,
		}
		if validRanges[args.TimeRange] {
			searchOpts.TimeRange = args.TimeRange
		}
	}

	if args.SiteFilter != "" {
		searchOpts.SiteFilter = args.SiteFilter
		args.Query = fmt.Sprintf("site:%s %s", args.SiteFilter, args.Query)
	}

	// 执行搜索
	results, err := t.module.Search(ctx, args.Query, searchOpts)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	// 格式化结果
	return t.formatResults(results, args.Query), nil
}

// formatResults 格式化搜索结果
func (t *webSearchTool) formatResults(results []SearchResult, query string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## 网络搜索结果：'%s'\n\n", query))
	sb.WriteString(fmt.Sprintf("找到 %d 条相关结果：\n\n", len(results)))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("**URL**: %s\n", result.URL))
		sb.WriteString(fmt.Sprintf("**来源**: %s | **位置**: #%d\n", result.Source, result.Position))
		if result.Snippet != "" {
			snippet := result.Snippet
			if len(snippet) > 300 {
				snippet = snippet[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("**摘要**: %s\n", snippet))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n---\n*搜索由以下引擎提供: %s*", strings.Join(t.getAvailableEngines(), ", ")))

	return sb.String()
}

// getAvailableEngines 获取可用的搜索引擎列表
func (t *webSearchTool) getAvailableEngines() []string {
	t.module.mu.RLock()
	defer t.module.mu.RUnlock()

	engines := make([]string, 0, len(t.module.clients))
	for name := range t.module.clients {
		engines = append(engines, strings.Title(name))
	}
	return engines
}

// ==================== SearchAggregateTool ====================

// searchAggregateTool 搜索聚合工具，从多个搜索引擎获取结果并合并
type searchAggregateTool struct {
	module *WebSearchModule
}

// NewSearchAggregateTool 创建搜索聚合工具
func NewSearchAggregateTool(module *WebSearchModule) tool.BaseTool {
	return &searchAggregateTool{module: module}
}

// Info 返回工具信息
func (t *searchAggregateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_aggregate",
		Desc: "从多个搜索引擎聚合搜索结果，去重并按相关性排序，提供更全面的信息覆盖",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索查询内容",
				Required: true,
			},
			"num_results": {
				Type:     "integer",
				Desc:     "每个引擎返回的结果数量，默认5条",
				Required: false,
			},
			"total_limit": {
				Type:     "integer",
				Desc:     "最终返回的总结果数上限，默认20条",
				Required: false,
			},
			"include_stats": {
				Type:     "boolean",
				Desc:     "是否包含搜索统计信息，默认false",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (t *searchAggregateTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		Query        string `json:"query"`
		NumResults   int    `json:"num_results"`
		TotalLimit   int    `json:"total_limit"`
		IncludeStats bool   `json:"include_stats"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	perEngine := 5
	if args.NumResults > 0 {
		perEngine = args.NumResults
	}

	totalLimit := 20
	if args.TotalLimit > 0 {
		totalLimit = args.TotalLimit
	}

	// 执行聚合搜索
	searchOpts := &SearchOptions{
		NumResults: perEngine,
		SafeSearch: false,
	}

	results, err := t.module.Search(ctx, args.Query, searchOpts)
	if err != nil {
		return "", fmt.Errorf("aggregate search failed: %w", err)
	}

	// 限制总结果数
	if len(results) > totalLimit {
		results = results[:totalLimit]
	}

	// 格式化结果
	output := t.formatAggregateResults(results, args.Query)

	// 添加统计信息
	if args.IncludeStats {
		output += "\n\n" + t.formatStats()
	}

	return output, nil
}

// formatAggregateResults 格式化聚合搜索结果
func (t *searchAggregateTool) formatAggregateResults(results []SearchResult, query string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## 🔍 聚合搜索结果：'%s'\n\n", query))
	sb.WriteString(fmt.Sprintf("**总计**: %d 条结果（已去重）\n\n", len(results)))

	// 按来源分组
	bySource := make(map[string][]SearchResult)
	for _, result := range results {
		bySource[result.Source] = append(bySource[result.Source], result)
	}

	// 显示结果
	for i, result := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("**URL**: %s\n", result.URL))
		sb.WriteString(fmt.Sprintf("**来源**: %s\n", strings.Title(result.Source)))

		if result.Snippet != "" {
			snippet := result.Snippet
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("**摘要**: %s\n", snippet))
		}
		sb.WriteString("\n")
	}

	// 来源统计
	sb.WriteString("\n### 📊 来源分布\n")
	for source, sourceResults := range bySource {
		sb.WriteString(fmt.Sprintf("- **%s**: %d 条\n", strings.Title(source), len(sourceResults)))
	}

	return sb.String()
}

// formatStats 格式化统计信息
func (t *searchAggregateTool) formatStats() string {
	stats := t.module.GetStats()

	var sb strings.Builder
	sb.WriteString("### 📈 搜索统计\n")
	sb.WriteString(fmt.Sprintf("- **总搜索次数**: %d\n", stats.TotalSearches))
	sb.WriteString(fmt.Sprintf("- **成功次数**: %d\n", stats.SuccessCount))
	sb.WriteString(fmt.Sprintf("- **失败次数**: %d\n", stats.FailureCount))
	sb.WriteString(fmt.Sprintf("- **缓存命中**: %d\n", stats.CacheHitCount))
	sb.WriteString(fmt.Sprintf("- **总结果数**: %d\n", stats.TotalResults))

	if stats.TotalSearches > 0 {
		successRate := float64(stats.SuccessCount) / float64(stats.TotalSearches) * 100
		sb.WriteString(fmt.Sprintf("- **成功率**: %.1f%%\n", successRate))
	}

	return sb.String()
}

// ==================== SearchURLTool ====================

// searchURLTool 搜索特定URL相关内容
type searchURLTool struct {
	module *WebSearchModule
}

// NewSearchURLTool 创建URL搜索工具
func NewSearchURLTool(module *WebSearchModule) tool.BaseTool {
	return &searchURLTool{module: module}
}

// Info 返回工具信息
func (t *searchURLTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_url",
		Desc: "在特定网站或域名内搜索内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索查询内容",
				Required: true,
			},
			"domain": {
				Type:     "string",
				Desc:     "要搜索的域名（如 'wikipedia.org'）",
				Required: true,
			},
			"num_results": {
				Type:     "integer",
				Desc:     "返回结果数量，默认10条",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (t *searchURLTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		Query      string `json:"query"`
		Domain     string `json:"domain"`
		NumResults int    `json:"num_results"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	if args.Domain == "" {
		return "", fmt.Errorf("domain parameter is required")
	}

	searchOpts := &SearchOptions{
		NumResults: 10,
		SafeSearch: false,
		SiteFilter: args.Domain,
	}

	if args.NumResults > 0 {
		searchOpts.NumResults = args.NumResults
	}

	// 添加site:前缀
	searchQuery := fmt.Sprintf("site:%s %s", args.Domain, args.Query)

	results, err := t.module.Search(ctx, searchQuery, searchOpts)
	if err != nil {
		return "", fmt.Errorf("site search failed: %w", err)
	}

	return t.formatSiteResults(results, args.Query, args.Domain), nil
}

// formatSiteResults 格式化站点搜索结果
func (t *searchURLTool) formatSiteResults(results []SearchResult, query, domain string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## 🌐 站点搜索：'%s' 在 %s\n\n", query, domain))

	if len(results) == 0 {
		sb.WriteString("未找到相关结果。\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("找到 %d 条来自 %s 的结果：\n\n", len(results), domain))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("**URL**: %s\n", result.URL))
		if result.Snippet != "" {
			snippet := result.Snippet
			if len(snippet) > 300 {
				snippet = snippet[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("**摘要**: %s\n", snippet))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
