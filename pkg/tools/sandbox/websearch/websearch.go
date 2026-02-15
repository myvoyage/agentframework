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
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
)

// WebSearchModule 网络搜索模块，支持多个搜索引擎
type WebSearchModule struct {
	config      WebSearchConfig
	clients     map[string]SearchClient
	stats       *SearchStats
	mu          sync.RWMutex
	initialized bool
}

// WebSearchConfig 网络搜索配置
type WebSearchConfig struct {
	// Timeout 搜索超时时间（毫秒）
	Timeout int `json:"timeout"`
	// MaxResults 最大返回结果数
	MaxResults int `json:"max_results"`
	// EnableSearchers 启用的搜索引擎
	EnableSearchers []string `json:"enable_searchers"`
	// APIKeys 各搜索引擎的API密钥
	APIKeys map[string]string `json:"api_keys"`
	// UserAgent 自定义User-Agent
	UserAgent string `json:"user_agent"`
	// Proxy 代理设置
	Proxy ProxyConfig `json:"proxy"`
	// CacheEnabled 是否启用缓存
	CacheEnabled bool `json:"cache_enabled"`
	// CacheTTL 缓存过期时间（秒）
	CacheTTL int `json:"cache_ttl"`
	// SafeSearch 是否启用安全搜索
	SafeSearch bool `json:"safe_search"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SearchStats 搜索统计
type SearchStats struct {
	TotalSearches  int64
	SuccessCount   int64
	FailureCount   int64
	CacheHitCount  int64
	TotalResults   int64
	mu             sync.RWMutex
	searcherCounts map[string]int64
}

// SearchResult 搜索结果
type SearchResult struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Snippet  string `json:"snippet"`
	Source   string `json:"source"`
	Position int    `json:"position"`
	Score    float64 `json:"score,omitempty"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	NumResults int    `json:"num_results"`
	Offset     int    `json:"offset"`
	SafeSearch bool   `json:"safe_search"`
	Language   string `json:"language"`
	Region     string `json:"region"`
	TimeRange  string `json:"time_range"` // past_day, past_week, past_month, past_year
	SiteFilter string `json:"site_filter"`
}

// SearchClient 搜索引擎客户端接口
type SearchClient interface {
	// Name 返回搜索引擎名称
	Name() string
	// Search 执行搜索
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
	// IsAvailable 检查搜索引擎是否可用
	IsAvailable(ctx context.Context) bool
}

// NewWebSearchModule 创建网络搜索模块
func NewWebSearchModule(config WebSearchConfig) (*WebSearchModule, error) {
	if config.Timeout == 0 {
		config.Timeout = 30000 // 默认30秒
	}
	if config.MaxResults == 0 {
		config.MaxResults = 10 // 默认10条结果
	}
	if config.EnableSearchers == nil {
		config.EnableSearchers = []string{"duckduckgo"} // 默认使用 DuckDuckGo
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 3600 // 默认1小时
	}

	module := &WebSearchModule{
		config:  config,
		clients: make(map[string]SearchClient),
		stats: &SearchStats{
			searcherCounts: make(map[string]int64),
		},
	}

	// 初始化搜索引擎客户端
	if err := module.initClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize search clients: %w", err)
	}

	module.initialized = true
	return module, nil
}

// initClients 初始化搜索引擎客户端
func (m *WebSearchModule) initClients() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, searcherName := range m.config.EnableSearchers {
		var client SearchClient
		var err error

		switch searcherName {
		case "google":
			client, err = NewGoogleSearchClient(GoogleConfig{
				APIKey:   m.config.APIKeys["google"],
				CXID:     m.config.APIKeys["google_cx"],
				Timeout:  time.Duration(m.config.Timeout) * time.Millisecond,
				SafeSearch: m.config.SafeSearch,
			})
		case "bing":
			client, err = NewBingSearchClient(BingConfig{
				APIKey:     m.config.APIKeys["bing"],
				Timeout:    time.Duration(m.config.Timeout) * time.Millisecond,
				SafeSearch: m.config.SafeSearch,
			})
		case "duckduckgo":
			client, err = NewDuckDuckGoClient(DuckDuckGoConfig{
				Timeout:    time.Duration(m.config.Timeout) * time.Millisecond,
				SafeSearch: m.config.SafeSearch,
			})
		default:
			return fmt.Errorf("unsupported search engine: %s", searcherName)
		}

		if err != nil {
			return fmt.Errorf("failed to create %s client: %w", searcherName, err)
		}

		m.clients[searcherName] = client
	}

	return nil
}

// Search 执行网络搜索
func (m *WebSearchModule) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	if !m.initialized {
		return nil, fmt.Errorf("websearch module not initialized")
	}

	if opts == nil {
		opts = &SearchOptions{
			NumResults: m.config.MaxResults,
			SafeSearch: m.config.SafeSearch,
		}
	}

	// 记录统计
	m.stats.mu.Lock()
	m.stats.TotalSearches++
	m.stats.mu.Unlock()

	// 聚合所有搜索引擎的结果
	allResults := make([]SearchResult, 0)
	resultsMap := make(map[string]SearchResult) // 用于去重

	m.mu.RLock()
	availableClients := make([]SearchClient, 0, len(m.clients))
	for _, client := range m.clients {
		if client.IsAvailable(ctx) {
			availableClients = append(availableClients, client)
		}
	}
	m.mu.RUnlock()

	if len(availableClients) == 0 {
		return nil, fmt.Errorf("no available search engines")
	}

	// 并发搜索所有搜索引擎
	var wg sync.WaitGroup
	resultsChan := make(chan []SearchResult, len(availableClients))
	errorsChan := make(chan error, len(availableClients))

	for _, client := range availableClients {
		wg.Add(1)
		go func(sc SearchClient) {
			defer wg.Done()

			results, err := sc.Search(ctx, query, *opts)
			if err != nil {
				errorsChan <- err
				return
			}
			resultsChan <- results
		}(client)
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// 收集结果
	for results := range resultsChan {
		for _, result := range results {
			// 使用URL去重
			if _, exists := resultsMap[result.URL]; !exists {
				resultsMap[result.URL] = result
				allResults = append(allResults, result)
			}
		}
	}

	// 处理错误（至少有一个成功即可）
	for range errorsChan {
		// 记录错误但继续处理
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no results found from any search engine")
	}

	// 按相关性排序并限制结果数量
	allResults = m.sortAndDeduplicateResults(allResults)
	if len(allResults) > opts.NumResults {
		allResults = allResults[:opts.NumResults]
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TotalResults += int64(len(allResults))
	m.stats.mu.Unlock()

	return allResults, nil
}

// sortAndDeduplicateResults 对结果进行排序和去重
func (m *WebSearchModule) sortAndDeduplicateResults(results []SearchResult) []SearchResult {
	// 智能排序算法，考虑多个因素
	sorted := make([]SearchResult, len(results))
	copy(sorted, results)

	// 定义搜索引擎权重（可以根据实际情况调整）
	searcherWeights := map[string]float64{
		"google":    1.0,
		"bing":      0.9,
		"duckduckgo": 0.8,
	}

	// 计算每个结果的综合得分
	for i := range sorted {
		result := &sorted[i]

		// 基础得分：位置越低（越靠前）得分越高
		positionScore := 1.0 / float64(result.Position+1)

		// 搜索引擎权重
		searcherWeight := searcherWeights[result.Source]
		if searcherWeight == 0 {
			searcherWeight = 0.7 // 默认权重
		}

		// 结果分数（如果有的话）
		score := result.Score
		if score == 0 {
			score = 0.5 // 默认分数
		}

		// 综合得分
		result.Score = positionScore*0.6 + searcherWeight*0.3 + score*0.1
	}

	// 按综合得分降序排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Score < sorted[j+1].Score {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// GetTools 返回MCP工具列表
func (m *WebSearchModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	return []tool.BaseTool{
		NewWebSearchTool(m),
		NewSearchAggregateTool(m),
	}, nil
}

// GetStats 获取搜索统计信息
func (m *WebSearchModule) GetStats() SearchStatsInfo {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return SearchStatsInfo{
		TotalSearches: m.stats.TotalSearches,
		SuccessCount:  m.stats.SuccessCount,
		FailureCount:  m.stats.FailureCount,
		CacheHitCount: m.stats.CacheHitCount,
		TotalResults:  m.stats.TotalResults,
		SearcherCounts: make(map[string]int64),
	}
}

// SearchStatsInfo 搜索统计信息
type SearchStatsInfo struct {
	TotalSearches  int64            `json:"total_searches"`
	SuccessCount   int64            `json:"success_count"`
	FailureCount   int64            `json:"failure_count"`
	CacheHitCount  int64            `json:"cache_hit_count"`
	TotalResults   int64            `json:"total_results"`
	SearcherCounts map[string]int64 `json:"searcher_counts"`
}

// Close 关闭模块并释放资源
func (m *WebSearchModule) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients = make(map[string]SearchClient)
	m.initialized = false

	return nil
}

// Config 返回当前配置
func (m *WebSearchModule) Config() WebSearchConfig {
	return m.config
}

// IsAvailable 检查模块是否可用
func (m *WebSearchModule) IsAvailable(ctx context.Context) bool {
	if !m.initialized {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.IsAvailable(ctx) {
			return true
		}
	}

	return false
}

// DefaultWebSearchConfig 返回默认配置
func DefaultWebSearchConfig() WebSearchConfig {
	return WebSearchConfig{
		Timeout:         30000,
		MaxResults:      10,
		EnableSearchers: []string{"duckduckgo"},
		APIKeys:         make(map[string]string),
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Proxy: ProxyConfig{
			Enable: false,
		},
		CacheEnabled: true,
		CacheTTL:     3600,
		SafeSearch:   false,
	}
}
