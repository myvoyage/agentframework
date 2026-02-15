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
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ==================== Google Search ====================

// GoogleSearchClient Google搜索引擎客户端
type GoogleSearchClient struct {
	config GoogleConfig
	client *http.Client
}

// GoogleConfig Google配置
type GoogleConfig struct {
	APIKey     string
	CXID       string // 自定义搜索引擎ID
	Timeout    time.Duration
	SafeSearch bool
}

// NewGoogleSearchClient 创建Google搜索客户端
func NewGoogleSearchClient(config GoogleConfig) (*GoogleSearchClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Google API key is required")
	}
	if config.CXID == "" {
		return nil, fmt.Errorf("Google Custom Search Engine ID is required")
	}

	return &GoogleSearchClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// Name 返回搜索引擎名称
func (c *GoogleSearchClient) Name() string {
	return "google"
}

// Search 执行搜索
func (c *GoogleSearchClient) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// 构建请求URL
	endpoint := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("key", c.config.APIKey)
	params.Set("cx", c.config.CXID)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", opts.NumResults))

	if opts.Offset > 0 {
		params.Set("start", fmt.Sprintf("%d", opts.Offset+1))
	}

	if opts.SafeSearch || c.config.SafeSearch {
		params.Set("safe", "active")
	}

	if opts.Language != "" {
		params.Set("hl", opts.Language)
	}

	if opts.Region != "" {
		params.Set("gl", opts.Region)
	}

	// 构建请求
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var googleResp GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换结果
	results := make([]SearchResult, 0, len(googleResp.Items))
	for i, item := range googleResp.Items {
		results = append(results, SearchResult{
			Title:    item.Title,
			URL:      item.Link,
			Snippet:  item.Snippet,
			Source:   "google",
			Position: i + 1,
		})
	}

	return results, nil
}

// IsAvailable 检查是否可用
func (c *GoogleSearchClient) IsAvailable(ctx context.Context) bool {
	// 简单检查：如果有API密钥则认为可用
	return c.config.APIKey != "" && c.config.CXID != ""
}

// GoogleSearchResponse Google搜索响应
type GoogleSearchResponse struct {
	Items []GoogleSearchItem `json:"items"`
}

// GoogleSearchItem Google搜索结果项
type GoogleSearchItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// ==================== Bing Search ====================

// BingSearchClient Bing搜索引擎客户端
type BingSearchClient struct {
	config BingConfig
	client *http.Client
}

// BingConfig Bing配置
type BingConfig struct {
	APIKey     string
	Timeout    time.Duration
	SafeSearch bool
}

// NewBingSearchClient 创建Bing搜索客户端
func NewBingSearchClient(config BingConfig) (*BingSearchClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Bing API key is required")
	}

	return &BingSearchClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// Name 返回搜索引擎名称
func (c *BingSearchClient) Name() string {
	return "bing"
}

// Search 执行搜索
func (c *BingSearchClient) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// 构建请求URL
	endpoint := "https://api.bing.microsoft.com/v7.0/search"
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", opts.NumResults))

	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	safeSearch := "Off"
	if opts.SafeSearch || c.config.SafeSearch {
		safeSearch = "Strict"
	}
	params.Set("safeSearch", safeSearch)

	if opts.Language != "" {
		params.Set("mkt", opts.Language)
	}

	// 构建请求
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", c.config.APIKey)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var bingResp BingSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&bingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换结果
	results := make([]SearchResult, 0, len(bingResp.WebPages.Value))
	for i, item := range bingResp.WebPages.Value {
		results = append(results, SearchResult{
			Title:    item.Name,
			URL:      item.URL,
			Snippet:  item.Snippet,
			Source:   "bing",
			Position: i + 1,
		})
	}

	return results, nil
}

// IsAvailable 检查是否可用
func (c *BingSearchClient) IsAvailable(ctx context.Context) bool {
	return c.config.APIKey != ""
}

// BingSearchResponse Bing搜索响应
type BingSearchResponse struct {
	WebPages BingWebPages `json:"webPages"`
}

// BingWebPages Bing网页结果
type BingWebPages struct {
	Value []BingWebPageItem `json:"value"`
}

// BingWebPageItem Bing网页结果项
type BingWebPageItem struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// ==================== DuckDuckGo Search ====================

// DuckDuckGoClient DuckDuckGo搜索引擎客户端（HTML解析方式，无需API密钥）
type DuckDuckGoClient struct {
	config DuckDuckGoConfig
	client *http.Client
}

// DuckDuckGoConfig DuckDuckGo配置
type DuckDuckGoConfig struct {
	Timeout    time.Duration
	SafeSearch bool
}

// NewDuckDuckGoClient 创建DuckDuckGo搜索客户端
func NewDuckDuckGoClient(config DuckDuckGoConfig) (*DuckDuckGoClient, error) {
	return &DuckDuckGoClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// Name 返回搜索引擎名称
func (c *DuckDuckGoClient) Name() string {
	return "duckduckgo"
}

// Search 执行搜索
func (c *DuckDuckGoClient) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// 使用DuckDuckGo的HTML版本（无需API密钥）
	endpoint := "https://html.duckduckgo.com/html/"
	params := url.Values{}
	params.Set("q", query)

	// 构建请求
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析HTML响应
	results, err := c.parseDuckDuckGoHTML(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 限制结果数量
	if len(results) > opts.NumResults {
		results = results[:opts.NumResults]
	}

	return results, nil
}

// parseDuckDuckGoHTML 解析DuckDuckGo HTML响应
func (c *DuckDuckGoClient) parseDuckDuckGoHTML(html string) ([]SearchResult, error) {
	results := make([]SearchResult, 0)

	// 简单的HTML解析（实际生产环境建议使用专门的HTML解析库）
	// 查找结果块：class="result"
	parts := strings.Split(html, `<div class="result`)
	for i, part := range parts {
		if i == 0 {
			continue // 第一部分不是结果
		}

		// 提取标题和链接
		// 查找 <a class="result__a" href="...">
		titleStart := strings.Index(part, `class="result__a"`)
		if titleStart == -1 {
			continue
		}

		hrefStart := strings.Index(part[titleStart:], `href="`)
		if hrefStart == -1 {
			continue
		}
		hrefStart += titleStart + len(`href="`)

		hrefEnd := strings.Index(part[hrefStart:], `"`)
		if hrefEnd == -1 {
			continue
		}
		link := part[hrefStart : hrefStart+hrefEnd]

		// 提取URL（DuckDuckGo使用重定向链接，需要解析）
		actualURL := c.extractDuckDuckGoURL(link)

		// 提取标题文本
		titleStart = strings.Index(part[hrefStart+hrefEnd:], `>`)
		if titleStart == -1 {
			continue
		}
		titleStart += hrefStart + hrefEnd + len(`>`)

		titleEnd := strings.Index(part[titleStart:], `<`)
		if titleEnd == -1 {
			continue
		}
		title := part[titleStart : titleStart+titleEnd]
		title = strings.TrimSpace(title)

		// 提取摘要
		snippetStart := strings.Index(part[titleStart+titleEnd:], `class="result__snippet"`)
		if snippetStart == -1 {
			snippetStart = strings.Index(part[titleStart+titleEnd:], `<a class="result__snippet"`)
		}
		if snippetStart == -1 {
			continue
		}
		snippetStart += titleStart + titleEnd

		snippetContentStart := strings.Index(part[snippetStart:], `>`)
		if snippetContentStart == -1 {
			continue
		}
		snippetContentStart += snippetStart + len(`>`)

		snippetEnd := strings.Index(part[snippetContentStart:], `<`)
		if snippetEnd == -1 {
			continue
		}
		snippet := part[snippetContentStart : snippetContentStart+snippetEnd]
		snippet = strings.TrimSpace(snippet)
		snippet = strings.Join(strings.Fields(snippet), " ") // 压缩空白

		if title != "" && actualURL != "" {
			results = append(results, SearchResult{
				Title:    title,
				URL:      actualURL,
				Snippet:  snippet,
				Source:   "duckduckgo",
				Position: i,
			})
		}
	}

	return results, nil
}

// extractDuckDuckGoURL 从DuckDuckGo重定向链接中提取实际URL
func (c *DuckDuckGoClient) extractDuckDuckGoURL(ddgURL string) string {
	// DuckDuckGo URL格式: /l/?uddg=https://example.com
	if strings.HasPrefix(ddgURL, "/l/?uddg=") {
		encodedURL := ddgURL[len("/l/?uddg="):]
		if unescaped, err := url.QueryUnescape(encodedURL); err == nil {
			return unescaped
		}
	}
	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(ddgURL, "http://") || strings.HasPrefix(ddgURL, "https://") {
		return ddgURL
	}
	return ddgURL
}

// IsAvailable 检查是否可用
func (c *DuckDuckGoClient) IsAvailable(ctx context.Context) bool {
	// DuckDuckGo不需要API密钥，总是可用
	return true
}

// ==================== 辅助函数 ====================

// NormalizeURL 规范化URL
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	// 移除片段和某些跟踪参数
	return fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, u.Query().Encode())
}

// TruncateText 截断文本到指定长度
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}
