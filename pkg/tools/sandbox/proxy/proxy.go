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

package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/net/proxy"
)

// ProxyModule 代理支持模块
type ProxyModule struct {
	config  ProxyConfig
	manager *ProxyManager
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enable              bool   `json:"enable"`
	Type                string `json:"type"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	PoolSize            int    `json:"pool_size"`
	HealthCheckInterval int    `json:"health_check_interval"` // seconds
	HealthCheckURL      string `json:"health_check_url"`
	Strategy            string `json:"strategy"` // round_robin, random, least_used
}

// Proxy 代理信息
type Proxy struct {
	URL         string    `json:"url"`
	Type        string    `json:"type"` // http, https, socks5
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Healthy     bool      `json:"healthy"`
	LastCheck   time.Time `json:"last_check"`
	FailCount   int       `json:"fail_count"`
	SuccessRate float64   `json:"success_rate"`
	UseCount    int       `json:"use_count"`
}

// ProxyManager 代理管理器
type ProxyManager struct {
	proxies       []*Proxy
	mu            sync.RWMutex
	healthChecker *HealthChecker
	strategy      LoadBalanceStrategy
	stopChan      chan struct{}
}

// LoadBalanceStrategy 负载均衡策略接口
type LoadBalanceStrategy interface {
	Select(proxies []*Proxy) (*Proxy, error)
}

// RoundRobinStrategy 轮询策略
type RoundRobinStrategy struct {
	current int
	mu      sync.Mutex
}

// RandomStrategy 随机策略
type RandomStrategy struct{}

// LeastUsedStrategy 最少使用策略
type LeastUsedStrategy struct{}

// HealthChecker 健康检查器
type HealthChecker struct {
	manager  *ProxyManager
	interval time.Duration
	testURL  string
	stopChan chan struct{}
}

// NewProxyModule 创建代理模块实例
func NewProxyModule(config ProxyConfig) (*ProxyModule, error) {
	// 设置默认值
	if config.PoolSize <= 0 {
		config.PoolSize = 5
	}
	if config.HealthCheckInterval <= 0 {
		config.HealthCheckInterval = 60
	}
	if config.HealthCheckURL == "" {
		config.HealthCheckURL = "https://www.google.com"
	}
	if config.Strategy == "" {
		config.Strategy = "round_robin"
	}

	manager := &ProxyManager{
		proxies:  make([]*Proxy, 0),
		stopChan: make(chan struct{}),
	}

	// 设置负载均衡策略
	switch config.Strategy {
	case "round_robin":
		manager.strategy = &RoundRobinStrategy{}
	case "random":
		manager.strategy = &RandomStrategy{}
	case "least_used":
		manager.strategy = &LeastUsedStrategy{}
	default:
		manager.strategy = &RoundRobinStrategy{}
	}

	// 创建健康检查器
	manager.healthChecker = &HealthChecker{
		manager:  manager,
		interval: time.Duration(config.HealthCheckInterval) * time.Second,
		testURL:  config.HealthCheckURL,
		stopChan: make(chan struct{}),
	}

	// 如果配置了代理，添加到池中
	if config.Host != "" && config.Port > 0 {
		proxyURL := fmt.Sprintf("%s://%s:%d", config.Type, config.Host, config.Port)
		p := &Proxy{
			URL:       proxyURL,
			Type:      config.Type,
			Username:  config.Username,
			Password:  config.Password,
			Healthy:   true,
			LastCheck: time.Now(),
		}
		manager.proxies = append(manager.proxies, p)
	}

	// 启动健康检查
	if config.Enable {
		go manager.healthChecker.Start()
	}

	return &ProxyModule{
		config:  config,
		manager: manager,
	}, nil
}

// LoadBalanceStrategy 实现

// Select 轮询策略选择代理
func (s *RoundRobinStrategy) Select(proxies []*Proxy) (*Proxy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	healthy := filterHealthy(proxies)
	if len(healthy) == 0 {
		return nil, errors.New("no healthy proxy available")
	}

	proxy := healthy[s.current%len(healthy)]
	s.current++
	proxy.UseCount++
	return proxy, nil
}

// Select 随机策略选择代理
func (s *RandomStrategy) Select(proxies []*Proxy) (*Proxy, error) {
	healthy := filterHealthy(proxies)
	if len(healthy) == 0 {
		return nil, errors.New("no healthy proxy available")
	}

	proxy := healthy[rand.Intn(len(healthy))]
	proxy.UseCount++
	return proxy, nil
}

// Select 最少使用策略选择代理
func (s *LeastUsedStrategy) Select(proxies []*Proxy) (*Proxy, error) {
	healthy := filterHealthy(proxies)
	if len(healthy) == 0 {
		return nil, errors.New("no healthy proxy available")
	}

	// 选择使用次数最少的代理
	minUsed := healthy[0]
	for _, p := range healthy {
		if p.UseCount < minUsed.UseCount {
			minUsed = p
		}
	}
	minUsed.UseCount++
	return minUsed, nil
}

// filterHealthy 过滤健康的代理
func filterHealthy(proxies []*Proxy) []*Proxy {
	healthy := make([]*Proxy, 0)
	for _, p := range proxies {
		if p.Healthy {
			healthy = append(healthy, p)
		}
	}
	return healthy
}

// ProxyManager 方法

// AddProxy 添加代理
func (m *ProxyManager) AddProxy(proxyURL, proxyType, username, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证代理 URL
	if _, err := url.Parse(proxyURL); err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	// 检查是否已存在
	for _, p := range m.proxies {
		if p.URL == proxyURL {
			return errors.New("proxy already exists")
		}
	}

	p := &Proxy{
		URL:       proxyURL,
		Type:      proxyType,
		Username:  username,
		Password:  password,
		Healthy:   true,
		LastCheck: time.Now(),
	}

	m.proxies = append(m.proxies, p)
	return nil
}

// RemoveProxy 移除代理
func (m *ProxyManager) RemoveProxy(proxyURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.proxies {
		if p.URL == proxyURL {
			m.proxies = append(m.proxies[:i], m.proxies[i+1:]...)
			return nil
		}
	}

	return errors.New("proxy not found")
}

// GetProxy 获取可用代理
func (m *ProxyManager) GetProxy() (*Proxy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.proxies) == 0 {
		return nil, errors.New("no proxy available")
	}

	return m.strategy.Select(m.proxies)
}

// ListProxies 列出所有代理
func (m *ProxyManager) ListProxies() []*Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Proxy, len(m.proxies))
	copy(result, m.proxies)
	return result
}

// MarkFailed 标记代理失败
func (m *ProxyManager) MarkFailed(proxyURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.proxies {
		if p.URL == proxyURL {
			p.FailCount++
			p.SuccessRate = calculateSuccessRate(p)

			// 失败次数超过阈值，标记为不健康
			if p.FailCount > 3 {
				p.Healthy = false
			}
			break
		}
	}
}

// MarkSuccess 标记代理成功
func (m *ProxyManager) MarkSuccess(proxyURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.proxies {
		if p.URL == proxyURL {
			p.SuccessRate = calculateSuccessRate(p)
			break
		}
	}
}

// SetStrategy 设置负载均衡策略
func (m *ProxyManager) SetStrategy(strategy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch strategy {
	case "round_robin":
		m.strategy = &RoundRobinStrategy{}
	case "random":
		m.strategy = &RandomStrategy{}
	case "least_used":
		m.strategy = &LeastUsedStrategy{}
	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}

	return nil
}

// calculateSuccessRate 计算成功率
func calculateSuccessRate(p *Proxy) float64 {
	total := p.UseCount
	if total == 0 {
		return 1.0
	}
	success := total - p.FailCount
	return float64(success) / float64(total)
}

// HealthChecker 方法

// Start 启动健康检查
func (h *HealthChecker) Start() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkAll()
		case <-h.stopChan:
			return
		}
	}
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopChan)
}

// checkAll 检查所有代理
func (h *HealthChecker) checkAll() {
	h.manager.mu.RLock()
	proxies := make([]*Proxy, len(h.manager.proxies))
	copy(proxies, h.manager.proxies)
	h.manager.mu.RUnlock()

	for _, p := range proxies {
		go h.checkOne(p)
	}
}

// checkOne 检查单个代理
func (h *HealthChecker) checkOne(p *Proxy) {
	healthy := h.testProxy(p)

	h.manager.mu.Lock()
	defer h.manager.mu.Unlock()

	p.LastCheck = time.Now()
	if healthy {
		p.Healthy = true
		// 如果之前失败，重置失败计数
		if p.FailCount > 0 {
			p.FailCount = 0
		}
	} else {
		p.FailCount++
		if p.FailCount > 3 {
			p.Healthy = false
		}
	}
}

// testProxy 测试代理连接
func (h *HealthChecker) testProxy(p *Proxy) bool {
	client, err := createHTTPClient(p, 10*time.Second)
	if err != nil {
		return false
	}

	resp, err := client.Get(h.testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// createHTTPClient 创建 HTTP 客户端
func createHTTPClient(p *Proxy, timeout time.Duration) (*http.Client, error) {
	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	// 设置认证
	if p.Username != "" && p.Password != "" {
		proxyURL.User = url.UserPassword(p.Username, p.Password)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// CreateSOCKS5Dialer 创建 SOCKS5 拨号器
func CreateSOCKS5Dialer(p *Proxy) (proxy.Dialer, error) {
	var auth *proxy.Auth
	if p.Username != "" && p.Password != "" {
		auth = &proxy.Auth{
			User:     p.Username,
			Password: p.Password,
		}
	}

	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	return proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
}

// GetTools 返回代理模块的 MCP 工具列表
func (m *ProxyModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	// 如果代理模块未启用，返回空列表
	if !m.config.Enable {
		return []tool.BaseTool{}, nil
	}

	tools := []tool.BaseTool{
		&proxyAddTool{module: m},
		&proxyRemoveTool{module: m},
		&proxyListTool{module: m},
		&proxyGetTool{module: m},
		&proxyHealthCheckTool{module: m},
		&proxySetStrategyTool{module: m},
	}

	return tools, nil
}

// proxy_add 工具
type proxyAddTool struct {
	module *ProxyModule
}

func (t *proxyAddTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "proxy_add",
		Desc: "Add a proxy to the proxy pool",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     "string",
				Desc:     "Proxy URL (e.g., http://proxy.example.com:8080)",
				Required: true,
			},
			"type": {
				Type:     "string",
				Desc:     "Proxy type: http, https, socks5",
				Required: true,
			},
			"username": {
				Type: "string",
				Desc: "Proxy username (optional)",
			},
			"password": {
				Type: "string",
				Desc: "Proxy password (optional)",
			},
		}),
	}, nil
}

func (t *proxyAddTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		URL      string `json:"url"`
		Type     string `json:"type"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.URL == "" || args.Type == "" {
		return "", errors.New("url and type are required")
	}

	err := t.module.manager.AddProxy(args.URL, args.Type, args.Username, args.Password)
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"success": true,
		"message": "Proxy added successfully",
		"proxy": map[string]any{
			"url":  args.URL,
			"type": args.Type,
		},
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// proxy_remove 工具
type proxyRemoveTool struct {
	module *ProxyModule
}

func (t *proxyRemoveTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "proxy_remove",
		Desc: "Remove a proxy from the proxy pool",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     "string",
				Desc:     "Proxy URL to remove",
				Required: true,
			},
		}),
	}, nil
}

func (t *proxyRemoveTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.URL == "" {
		return "", errors.New("url is required")
	}

	err := t.module.manager.RemoveProxy(args.URL)
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"success": true,
		"message": "Proxy removed successfully",
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// proxy_list 工具
type proxyListTool struct {
	module *ProxyModule
}

func (t *proxyListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "proxy_list",
		Desc:        "List all proxies in the pool",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *proxyListTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	proxies := t.module.manager.ListProxies()

	result := map[string]any{
		"success": true,
		"count":   len(proxies),
		"proxies": proxies,
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// proxy_get 工具
type proxyGetTool struct {
	module *ProxyModule
}

func (t *proxyGetTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "proxy_get",
		Desc:        "Get an available proxy from the pool based on load balancing strategy",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *proxyGetTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	proxy, err := t.module.manager.GetProxy()
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"success": true,
		"proxy":   proxy,
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// proxy_health_check 工具
type proxyHealthCheckTool struct {
	module *ProxyModule
}

func (t *proxyHealthCheckTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "proxy_health_check",
		Desc: "Perform health check on all proxies or specific proxy",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type: "string",
				Desc: "Specific proxy URL to check (optional, checks all if not provided)",
			},
		}),
	}, nil
}

func (t *proxyHealthCheckTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	json.Unmarshal([]byte(input), &args)

	if args.URL != "" {
		// 检查特定代理
		t.module.manager.mu.RLock()
		var targetProxy *Proxy
		for _, p := range t.module.manager.proxies {
			if p.URL == args.URL {
				targetProxy = p
				break
			}
		}
		t.module.manager.mu.RUnlock()

		if targetProxy == nil {
			return "", errors.New("proxy not found")
		}

		t.module.manager.healthChecker.checkOne(targetProxy)

		result := map[string]any{
			"success": true,
			"proxy":   targetProxy,
		}

		output, _ := json.Marshal(result)
		return string(output), nil
	}

	// 检查所有代理
	t.module.manager.healthChecker.checkAll()

	// 等待检查完成
	time.Sleep(2 * time.Second)

	proxies := t.module.manager.ListProxies()

	result := map[string]any{
		"success": true,
		"count":   len(proxies),
		"proxies": proxies,
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// proxy_set_strategy 工具
type proxySetStrategyTool struct {
	module *ProxyModule
}

func (t *proxySetStrategyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "proxy_set_strategy",
		Desc: "Set the load balancing strategy for proxy selection",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"strategy": {
				Type:     "string",
				Desc:     "Load balancing strategy: round_robin, random, least_used",
				Required: true,
			},
		}),
	}, nil
}

func (t *proxySetStrategyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Strategy string `json:"strategy"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Strategy == "" {
		return "", errors.New("strategy is required")
	}

	err := t.module.manager.SetStrategy(args.Strategy)
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"success":  true,
		"message":  "Strategy set successfully",
		"strategy": args.Strategy,
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// Close 关闭代理模块，释放资源
func (m *ProxyModule) Close() error {
	if m.manager != nil && m.manager.healthChecker != nil {
		m.manager.healthChecker.Stop()
	}
	if m.manager != nil {
		close(m.manager.stopChan)
	}
	return nil
}

// CreateHTTPClient 创建使用代理的 HTTP 客户端（导出方法）
func (m *ProxyModule) CreateHTTPClient(timeout time.Duration) (*http.Client, error) {
	proxy, err := m.manager.GetProxy()
	if err != nil {
		return nil, err
	}

	return createHTTPClient(proxy, timeout)
}

// CreateSOCKS5Client 创建使用 SOCKS5 代理的客户端（导出方法）
func (m *ProxyModule) CreateSOCKS5Client() (proxy.Dialer, error) {
	p, err := m.manager.GetProxy()
	if err != nil {
		return nil, err
	}

	if p.Type != "socks5" {
		return nil, errors.New("proxy is not SOCKS5 type")
	}

	return CreateSOCKS5Dialer(p)
}
