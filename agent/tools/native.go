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

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NewFileReadTool creates a tool that reads files from a sandbox directory.
func NewFileReadTool(sandboxDir string) (tool.BaseTool, error) {
	return &fileReadTool{sandboxDir: sandboxDir}, nil
}

type fileReadTool struct {
	sandboxDir string
}

func (t *fileReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "read_file",
		Desc: "Read the content of a file. The path must be within the sandbox.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "Relative path to file",
			},
		}),
	}, nil
}

func (t *fileReadTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Security check: ensure path is inside sandbox
	absSandbox, err := filepath.Abs(t.sandboxDir)
	if err != nil {
		return "", fmt.Errorf("invalid sandbox path: %w", err)
	}
	targetPath := filepath.Join(absSandbox, args.Path)
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	if !strings.HasPrefix(absTarget, absSandbox) {
		return "", fmt.Errorf("access denied: path is outside sandbox")
	}

	content, err := os.ReadFile(absTarget)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// NewWebSearchTool creates a real web search tool that uses search APIs.
// It requires GOOGLE_SEARCH_API_KEY or BING_SEARCH_API_KEY environment variable.
func NewWebSearchTool() (tool.BaseTool, error) {
	apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
	if apiKey != "" {
		searchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")
		return &googleSearchTool{
			apiKey:     apiKey,
			engineID:   searchEngineID,
			cache:      make(map[string]*cachedResult),
			cacheMutex: &sync.RWMutex{},
		}, nil
	}

	bingKey := os.Getenv("BING_SEARCH_API_KEY")
	if bingKey != "" {
		return &bingSearchTool{
			apiKey:     bingKey,
			cache:      make(map[string]*cachedResult),
			cacheMutex: &sync.RWMutex{},
		}, nil
	}

	// Fall back to mock if no API key is provided
	return &webSearchTool{}, nil
}

type cachedResult struct {
	data      string
	timestamp time.Time
}

type googleSearchTool struct {
	apiKey     string
	engineID   string
	cache      map[string]*cachedResult
	cacheMutex *sync.RWMutex
	httpClient *http.Client
}

func (t *googleSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search_google",
		Desc: "Search the web using Google Custom Search API.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "The search query",
			},
		}),
	}, nil
}

func (t *googleSearchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Check cache first
	if cached := t.getCachedResult(args.Query); cached != "" {
		return cached, nil
	}

	// Call Google Custom Search API
	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	url := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=5",
		t.apiKey, t.engineID, url.QueryEscape(args.Query))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Title       string `json:"title"`
			Link        string `json:"link"`
			Snippet     string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Format results
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search Results for '%s':\n\n", args.Query))
	for i, item := range result.Items {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		output.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
		output.WriteString(fmt.Sprintf("   %s\n\n", item.Link))
	}

	// Cache result
	t.setCachedResult(args.Query, output.String(), 10*time.Minute)

	return output.String(), nil
}

func (t *googleSearchTool) getCachedResult(query string) string {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()

	if cached, ok := t.cache[query]; ok && time.Since(cached.timestamp) < 10*time.Minute {
		return cached.data
	}
	return ""
}

func (t *googleSearchTool) setCachedResult(query, data string, ttl time.Duration) {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()

	t.cache[query] = &cachedResult{
		data:      data,
		timestamp: time.Now(),
	}
}

type bingSearchTool struct {
	apiKey     string
	cache      map[string]*cachedResult
	cacheMutex *sync.RWMutex
	httpClient *http.Client
}

func (t *bingSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search_bing",
		Desc: "Search the web using Bing Search API.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "The search query",
			},
		}),
	}, nil
}

func (t *bingSearchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Check cache first
	if cached := t.getCachedResult(args.Query); cached != "" {
		return cached, nil
	}

	// Call Bing Search API
	if t.httpClient == nil {
		t.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	url := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s&count=5",
		url.QueryEscape(args.Query))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var result struct {
		WebPages struct {
			Value []struct {
				Name       string   `json:"name"`
				URL        string   `json:"url"`
				Snippet    string   `json:"snippet"`
				DisplayUrl string   `json:"displayUrl"`
			} `json:"value"`
		} `json:"webPages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Format results
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search Results for '%s':\n\n", args.Query))
	for i, item := range result.WebPages.Value {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Name))
		output.WriteString(fmt.Sprintf("   %s\n", item.Snippet))
		output.WriteString(fmt.Sprintf("   %s\n\n", item.DisplayUrl))
	}

	// Cache result
	t.setCachedResult(args.Query, output.String(), 10*time.Minute)

	return output.String(), nil
}

func (t *bingSearchTool) getCachedResult(query string) string {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()

	if cached, ok := t.cache[query]; ok && time.Since(cached.timestamp) < 10*time.Minute {
		return cached.data
	}
	return ""
}

func (t *bingSearchTool) setCachedResult(query, data string, ttl time.Duration) {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()

	t.cache[query] = &cachedResult{
		data:      data,
		timestamp: time.Now(),
	}
}

// webSearchTool is a fallback mock implementation when no API key is provided
type webSearchTool struct{}

func (t *webSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search_mock",
		Desc: "Mock web search tool (no API key configured). Set GOOGLE_SEARCH_API_KEY or BING_SEARCH_API_KEY to use real search.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "The search query",
			},
		}),
	}, nil
}

func (t *webSearchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Mock Search Results for '%s':\n1. Go 1.22 released in Feb 2024.\n2. Agent Frameworks are popular.\n3. Weather is sunny.\n\nNote: Configure GOOGLE_SEARCH_API_KEY or BING_SEARCH_API_KEY for real search results.", args.Query), nil
}
