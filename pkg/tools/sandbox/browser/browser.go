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
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// BrowserModule 浏览器交互模块
type BrowserModule struct {
	config BrowserConfig
	pool   *BrowserPool
	stats  *OperationStats
	mu     sync.RWMutex
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	Headless       bool     `json:"headless"`
	Timeout        int      `json:"timeout"` // 毫秒
	UserAgent      string   `json:"user_agent"`
	Viewport       Viewport `json:"viewport"`
	PoolSize       int      `json:"pool_size"`       // 连接池大小
	AllowedDomains []string `json:"allowed_domains"` // 域名白名单
	BlockedDomains []string `json:"blocked_domains"` // 域名黑名单
}

// Viewport 视口配置
type Viewport struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

// BrowserPool 浏览器连接池
type BrowserPool struct {
	contexts    []*BrowserContext
	available   chan *BrowserContext
	config      BrowserConfig
	mu          sync.Mutex
	maxSize     int
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

// BrowserContext 浏览器上下文
type BrowserContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	inUse  bool
	mu     sync.Mutex
}

// OperationStats 操作统计
type OperationStats struct {
	TotalOperations int64
	SuccessCount    int64
	FailureCount    int64
	BlockedCount    int64
	mu              sync.RWMutex
}

// NewBrowserModule 创建浏览器模块实例
func NewBrowserModule(config BrowserConfig) (*BrowserModule, error) {
	// 验证配置
	if config.Timeout <= 0 {
		config.Timeout = 30000 // 默认30秒
	}
	if config.Viewport.Width <= 0 {
		config.Viewport.Width = 1920
	}
	if config.Viewport.Height <= 0 {
		config.Viewport.Height = 1080
	}
	if config.PoolSize <= 0 {
		config.PoolSize = 5 // 默认5个浏览器实例
	}

	// 创建浏览器池
	pool, err := NewBrowserPool(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser pool: %w", err)
	}

	stats := &OperationStats{}

	return &BrowserModule{
		config: config,
		pool:   pool,
		stats:  stats,
	}, nil
}

// NewBrowserPool 创建浏览器连接池
func NewBrowserPool(config BrowserConfig) (*BrowserPool, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", config.Headless),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.WindowSize(int(config.Viewport.Width), int(config.Viewport.Height)),
			chromedp.UserAgent(config.UserAgent),
		)...,
	)

	pool := &BrowserPool{
		contexts:    make([]*BrowserContext, 0, config.PoolSize),
		available:   make(chan *BrowserContext, config.PoolSize),
		config:      config,
		maxSize:     config.PoolSize,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
	}

	return pool, nil
}

// Get 从池中获取浏览器上下文
func (p *BrowserPool) Get() (*BrowserContext, error) {
	// 尝试从可用队列获取
	select {
	case ctx := <-p.available:
		ctx.mu.Lock()
		ctx.inUse = true
		ctx.mu.Unlock()
		return ctx, nil
	default:
		// 如果没有可用的，尝试创建新的
		p.mu.Lock()
		defer p.mu.Unlock()

		if len(p.contexts) < p.maxSize {
			ctx, err := p.createContext()
			if err != nil {
				return nil, err
			}
			p.contexts = append(p.contexts, ctx)
			ctx.mu.Lock()
			ctx.inUse = true
			ctx.mu.Unlock()
			return ctx, nil
		}

		// 池已满，等待可用的上下文
		p.mu.Unlock()
		ctx := <-p.available
		p.mu.Lock()
		ctx.mu.Lock()
		ctx.inUse = true
		ctx.mu.Unlock()
		return ctx, nil
	}
}

// Put 将浏览器上下文放回池中
func (p *BrowserPool) Put(ctx *BrowserContext) {
	if ctx == nil {
		return
	}

	ctx.mu.Lock()
	ctx.inUse = false
	ctx.mu.Unlock()

	select {
	case p.available <- ctx:
		// 成功放回池中
	default:
		// 池已满，关闭上下文
		ctx.cancel()
	}
}

// createContext 创建新的浏览器上下文
func (p *BrowserPool) createContext() (*BrowserContext, error) {
	ctx, cancel := chromedp.NewContext(p.allocCtx)

	// 启动浏览器
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	return &BrowserContext{
		ctx:    ctx,
		cancel: cancel,
		inUse:  false,
	}, nil
}

// Close 关闭浏览器池
func (p *BrowserPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 关闭所有上下文
	for _, ctx := range p.contexts {
		ctx.cancel()
	}

	// 关闭分配器
	p.allocCancel()

	return nil
}

// GetTools 返回浏览器模块的 MCP 工具列表
func (m *BrowserModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 浏览器导航工具
		&browserNavigateTool{module: m},
		// 浏览器点击工具
		&browserClickTool{module: m},
		// 浏览器输入工具
		&browserInputTool{module: m},
		// 浏览器截图工具
		&browserScreenshotTool{module: m},
		// 浏览器PDF工具
		&browserPDFTool{module: m},
		// 浏览器获取文本工具
		&browserGetTextTool{module: m},
		// 浏览器执行JavaScript工具
		&browserExecuteJSTool{module: m},
		// 浏览器获取Cookie工具
		&browserGetCookiesTool{module: m},
		// 浏览器设置Cookie工具
		&browserSetCookiesTool{module: m},
	}

	return tools, nil
}

// Close 关闭浏览器模块，释放资源
func (m *BrowserModule) Close() error {
	if m.pool != nil {
		return m.pool.Close()
	}
	return nil
}

// GetStats 获取操作统计信息
func (m *BrowserModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_operations": m.stats.TotalOperations,
		"success_count":    m.stats.SuccessCount,
		"failure_count":    m.stats.FailureCount,
		"blocked_count":    m.stats.BlockedCount,
	}
}

// ============================================================================
// MCP Tools Implementation
// ============================================================================

// 浏览器导航工具
type browserNavigateTool struct {
	module *BrowserModule
}

func (t *browserNavigateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_navigate",
		Desc: "Navigate to a specified URL in the browser",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     "string",
				Desc:     "URL to navigate to",
				Required: true,
			},
			"timeout": {
				Type: "integer",
				Desc: "Navigation timeout in milliseconds",
			},
		}),
	}, nil
}

func (t *browserNavigateTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		URL     string `json:"url"`
		Timeout int    `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.navigate(args.URL, args.Timeout)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器点击工具
type browserClickTool struct {
	module *BrowserModule
}

func (t *browserClickTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_click",
		Desc: "Click an element on the page by CSS selector",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"selector": {
				Type:     "string",
				Desc:     "CSS selector of element to click",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserClickTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.click(args.Selector)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器输入工具
type browserInputTool struct {
	module *BrowserModule
}

func (t *browserInputTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_input",
		Desc: "Input text into an element by CSS selector",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"selector": {
				Type:     "string",
				Desc:     "CSS selector of input field",
				Required: true,
			},
			"value": {
				Type:     "string",
				Desc:     "Text to input",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserInputTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Selector string `json:"selector"`
		Value    string `json:"value"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.input(args.Selector, args.Value)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器截图工具
type browserScreenshotTool struct {
	module *BrowserModule
}

func (t *browserScreenshotTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_screenshot",
		Desc: "Take a screenshot of the current page or a specific element",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     "string",
				Desc:     "Path to save screenshot",
				Required: true,
			},
			"selector": {
				Type: "string",
				Desc: "CSS selector of element to screenshot (optional, full page if not provided)",
			},
			"full_page": {
				Type: "boolean",
				Desc: "Take full page screenshot",
			},
		}),
	}, nil
}

func (t *browserScreenshotTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path     string `json:"path"`
		Selector string `json:"selector"`
		FullPage bool   `json:"full_page"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.screenshot(args.Path, args.Selector, args.FullPage)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器PDF工具
type browserPDFTool struct {
	module *BrowserModule
}

func (t *browserPDFTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_pdf",
		Desc: "Generate a PDF of the current page",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     "string",
				Desc:     "Path to save PDF",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserPDFTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.generatePDF(args.Path)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器获取文本工具
type browserGetTextTool struct {
	module *BrowserModule
}

func (t *browserGetTextTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_get_text",
		Desc: "Get text content of an element by CSS selector",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"selector": {
				Type:     "string",
				Desc:     "CSS selector of element",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserGetTextTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.getText(args.Selector)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器执行JavaScript工具
type browserExecuteJSTool struct {
	module *BrowserModule
}

func (t *browserExecuteJSTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_execute_js",
		Desc: "Execute JavaScript code in the page context",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"script": {
				Type:     "string",
				Desc:     "JavaScript code to execute",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserExecuteJSTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Script string `json:"script"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.executeJS(args.Script)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器获取Cookie工具
type browserGetCookiesTool struct {
	module *BrowserModule
}

func (t *browserGetCookiesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "browser_get_cookies",
		Desc:        "Get all cookies from the current page",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *browserGetCookiesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.getCookies()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 浏览器设置Cookie工具
type browserSetCookiesTool struct {
	module *BrowserModule
}

func (t *browserSetCookiesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "browser_set_cookies",
		Desc: "Set cookies for the current page",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"cookies": {
				Type:     "array",
				Desc:     "Array of cookie objects with name, value, domain, path, etc.",
				Required: true,
			},
		}),
	}, nil
}

func (t *browserSetCookiesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Cookies []map[string]interface{} `json:"cookies"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.setCookies(args.Cookies)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// ============================================================================
// Core Functionality Implementation
// ============================================================================

// isURLAllowed 检查URL是否允许访问
func (m *BrowserModule) isURLAllowed(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Hostname()

	// 检查黑名单
	for _, blocked := range m.config.BlockedDomains {
		if strings.Contains(domain, blocked) {
			return fmt.Errorf("domain %s is blocked", domain)
		}
	}

	// 如果有白名单，检查是否在白名单中
	if len(m.config.AllowedDomains) > 0 {
		allowed := false
		for _, allowedDomain := range m.config.AllowedDomains {
			if strings.Contains(domain, allowedDomain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("domain %s is not in allowed list", domain)
		}
	}

	return nil
}

// navigate 导航到指定URL
func (m *BrowserModule) navigate(urlStr string, timeout int) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 检查URL是否允许
	if err := m.isURLAllowed(urlStr); err != nil {
		m.stats.mu.Lock()
		m.stats.BlockedCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"url":     urlStr,
		}, nil
	}

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
			"url":     urlStr,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	if timeout == 0 {
		timeout = m.config.Timeout
	}
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// 导航到URL
	startTime := time.Now()
	var title string
	err = chromedp.Run(ctx,
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Title(&title),
	)
	duration := time.Since(startTime)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    err.Error(),
			"url":      urlStr,
			"duration": duration.Milliseconds(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"url":      urlStr,
		"title":    title,
		"duration": duration.Milliseconds(),
		"message":  "Navigation successful",
	}, nil
}

// click 点击元素
func (m *BrowserModule) click(selector string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    fmt.Sprintf("failed to get browser context: %v", err),
			"selector": selector,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	// 点击元素
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    err.Error(),
			"selector": selector,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"selector": selector,
		"message":  "Element clicked successfully",
	}, nil
}

// input 输入文本到元素
func (m *BrowserModule) input(selector, value string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    fmt.Sprintf("failed to get browser context: %v", err),
			"selector": selector,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	// 输入文本
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Clear(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    err.Error(),
			"selector": selector,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"selector": selector,
		"value":    value,
		"message":  "Text input successful",
	}, nil
}

// screenshot 截图
func (m *BrowserModule) screenshot(path, selector string, fullPage bool) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
			"path":    path,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	var buf []byte
	var screenshotErr error

	if selector != "" {
		// 截取特定元素
		screenshotErr = chromedp.Run(ctx,
			chromedp.WaitVisible(selector, chromedp.ByQuery),
			chromedp.Screenshot(selector, &buf, chromedp.ByQuery),
		)
	} else if fullPage {
		// 全页截图
		screenshotErr = chromedp.Run(ctx,
			chromedp.FullScreenshot(&buf, 100),
		)
	} else {
		// 视口截图
		screenshotErr = chromedp.Run(ctx,
			chromedp.CaptureScreenshot(&buf),
		)
	}

	if screenshotErr != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   screenshotErr.Error(),
			"path":    path,
		}, nil
	}

	// 保存截图
	if err := os.WriteFile(path, buf, 0644); err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to save screenshot: %v", err),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":   true,
		"path":      path,
		"selector":  selector,
		"full_page": fullPage,
		"size":      len(buf),
		"message":   "Screenshot saved successfully",
	}, nil
}

// generatePDF 生成PDF
func (m *BrowserModule) generatePDF(path string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
			"path":    path,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	// 保存PDF
	if err := os.WriteFile(path, buf, 0644); err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to save PDF: %v", err),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"path":    path,
		"size":    len(buf),
		"message": "PDF generated successfully",
	}, nil
}

// getText 获取元素文本
func (m *BrowserModule) getText(selector string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    fmt.Sprintf("failed to get browser context: %v", err),
			"selector": selector,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	var text string
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success":  false,
			"error":    err.Error(),
			"selector": selector,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"selector": selector,
		"text":     text,
		"message":  "Text retrieved successfully",
	}, nil
}

// executeJS 执行JavaScript代码
func (m *BrowserModule) executeJS(script string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
			"script":  script,
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	var result interface{}
	err = chromedp.Run(ctx,
		chromedp.Evaluate(script, &result),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"script":  script,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"script":  script,
		"result":  result,
		"message": "JavaScript executed successfully",
	}, nil
}

// getCookies 获取Cookie
func (m *BrowserModule) getCookies() (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	var cookies []*network.Cookie
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// 转换Cookie为map格式
	cookieList := make([]map[string]interface{}, 0, len(cookies))
	for _, cookie := range cookies {
		cookieList = append(cookieList, map[string]interface{}{
			"name":     cookie.Name,
			"value":    cookie.Value,
			"domain":   cookie.Domain,
			"path":     cookie.Path,
			"expires":  cookie.Expires,
			"httpOnly": cookie.HTTPOnly,
			"secure":   cookie.Secure,
			"sameSite": cookie.SameSite.String(),
		})
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"cookies": cookieList,
		"count":   len(cookieList),
		"message": "Cookies retrieved successfully",
	}, nil
}

// setCookies 设置Cookie
func (m *BrowserModule) setCookies(cookies []map[string]interface{}) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 从池中获取浏览器上下文
	browserCtx, err := m.pool.Get()
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to get browser context: %v", err),
		}, nil
	}
	defer m.pool.Put(browserCtx)

	// 设置超时
	ctx, cancel := context.WithTimeout(browserCtx.ctx, time.Duration(m.config.Timeout)*time.Millisecond)
	defer cancel()

	// 转换Cookie格式
	var cookieParams []*network.CookieParam
	for _, cookie := range cookies {
		param := &network.CookieParam{
			Name:  getString(cookie, "name"),
			Value: getString(cookie, "value"),
		}

		if domain, ok := cookie["domain"].(string); ok {
			param.Domain = domain
		}
		if path, ok := cookie["path"].(string); ok {
			param.Path = path
		}
		if expires, ok := cookie["expires"].(float64); ok {
			// Convert Unix timestamp to time.Time, then to TimeSinceEpoch
			t := time.Unix(int64(expires), 0)
			expiresTime := cdp.TimeSinceEpoch(t)
			param.Expires = &expiresTime
		}
		if httpOnly, ok := cookie["httpOnly"].(bool); ok {
			param.HTTPOnly = httpOnly
		}
		if secure, ok := cookie["secure"].(bool); ok {
			param.Secure = secure
		}
		if sameSite, ok := cookie["sameSite"].(string); ok {
			switch strings.ToLower(sameSite) {
			case "strict":
				param.SameSite = network.CookieSameSiteStrict
			case "lax":
				param.SameSite = network.CookieSameSiteLax
			case "none":
				param.SameSite = network.CookieSameSiteNone
			}
		}

		cookieParams = append(cookieParams, param)
	}

	// 设置Cookie
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetCookies(cookieParams).Do(ctx)
		}),
	)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"count":   len(cookieParams),
		"message": "Cookies set successfully",
	}, nil
}

// getString 从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
