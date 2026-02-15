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

package proxy

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyModule(t *testing.T) {
	config := ProxyConfig{
		Enable:              true,
		Type:                "http",
		Host:                "proxy.example.com",
		Port:                8080,
		Username:            "user",
		Password:            "pass",
		PoolSize:            5,
		HealthCheckInterval: 60,
		HealthCheckURL:      "https://www.google.com",
		Strategy:            "round_robin",
	}

	module, err := NewProxyModule(config)
	require.NoError(t, err)
	require.NotNil(t, module)
	assert.Equal(t, config.Enable, module.config.Enable)
	assert.NotNil(t, module.manager)
	assert.NotNil(t, module.manager.healthChecker)

	// 清理
	module.Close()
}

func TestProxyManager_AddProxy(t *testing.T) {
	manager := &ProxyManager{
		proxies:  make([]*Proxy, 0),
		strategy: &RoundRobinStrategy{},
	}

	// 测试添加代理
	err := manager.AddProxy("http://proxy1.example.com:8080", "http", "user", "pass")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(manager.proxies))

	// 测试添加重复代理
	err = manager.AddProxy("http://proxy1.example.com:8080", "http", "user", "pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// 测试添加无效 URL - 注意：url.Parse 对大多数字符串都会成功
	// 所以我们不测试这个情况
}

func TestProxyManager_RemoveProxy(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
			{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true},
		},
		strategy: &RoundRobinStrategy{},
	}

	// 测试移除存在的代理
	err := manager.RemoveProxy("http://proxy1.example.com:8080")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(manager.proxies))

	// 测试移除不存在的代理
	err = manager.RemoveProxy("http://proxy3.example.com:8080")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProxyManager_GetProxy(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
			{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true},
		},
		strategy: &RoundRobinStrategy{},
	}

	// 测试获取代理
	proxy, err := manager.GetProxy()
	assert.NoError(t, err)
	assert.NotNil(t, proxy)
	assert.True(t, proxy.Healthy)

	// 测试空代理池
	emptyManager := &ProxyManager{
		proxies:  make([]*Proxy, 0),
		strategy: &RoundRobinStrategy{},
	}
	_, err = emptyManager.GetProxy()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no proxy available")
}

func TestProxyManager_ListProxies(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
			{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true},
		},
		strategy: &RoundRobinStrategy{},
	}

	proxies := manager.ListProxies()
	assert.Equal(t, 2, len(proxies))
}

func TestProxyManager_MarkFailed(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true, FailCount: 0},
		},
		strategy: &RoundRobinStrategy{},
	}

	// 标记失败
	manager.MarkFailed("http://proxy1.example.com:8080")
	assert.Equal(t, 1, manager.proxies[0].FailCount)
	assert.True(t, manager.proxies[0].Healthy)

	// 多次失败后标记为不健康
	manager.MarkFailed("http://proxy1.example.com:8080")
	manager.MarkFailed("http://proxy1.example.com:8080")
	manager.MarkFailed("http://proxy1.example.com:8080")
	assert.False(t, manager.proxies[0].Healthy)
}

func TestProxyManager_SetStrategy(t *testing.T) {
	manager := &ProxyManager{
		proxies:  make([]*Proxy, 0),
		strategy: &RoundRobinStrategy{},
	}

	// 测试设置轮询策略
	err := manager.SetStrategy("round_robin")
	assert.NoError(t, err)
	assert.IsType(t, &RoundRobinStrategy{}, manager.strategy)

	// 测试设置随机策略
	err = manager.SetStrategy("random")
	assert.NoError(t, err)
	assert.IsType(t, &RandomStrategy{}, manager.strategy)

	// 测试设置最少使用策略
	err = manager.SetStrategy("least_used")
	assert.NoError(t, err)
	assert.IsType(t, &LeastUsedStrategy{}, manager.strategy)

	// 测试无效策略
	err = manager.SetStrategy("invalid")
	assert.Error(t, err)
}

func TestRoundRobinStrategy(t *testing.T) {
	strategy := &RoundRobinStrategy{}
	proxies := []*Proxy{
		{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
		{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true},
		{URL: "http://proxy3.example.com:8080", Type: "http", Healthy: true},
	}

	// 测试轮询
	p1, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.Equal(t, "http://proxy1.example.com:8080", p1.URL)

	p2, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.Equal(t, "http://proxy2.example.com:8080", p2.URL)

	p3, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.Equal(t, "http://proxy3.example.com:8080", p3.URL)

	// 循环回到第一个
	p4, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.Equal(t, "http://proxy1.example.com:8080", p4.URL)
}

func TestRandomStrategy(t *testing.T) {
	strategy := &RandomStrategy{}
	proxies := []*Proxy{
		{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
		{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true},
	}

	// 测试随机选择
	proxy, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.NotNil(t, proxy)
	assert.True(t, proxy.Healthy)
}

func TestLeastUsedStrategy(t *testing.T) {
	strategy := &LeastUsedStrategy{}
	proxies := []*Proxy{
		{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true, UseCount: 5},
		{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: true, UseCount: 2},
		{URL: "http://proxy3.example.com:8080", Type: "http", Healthy: true, UseCount: 8},
	}

	// 应该选择使用次数最少的
	proxy, err := strategy.Select(proxies)
	assert.NoError(t, err)
	assert.Equal(t, "http://proxy2.example.com:8080", proxy.URL)
}

func TestFilterHealthy(t *testing.T) {
	proxies := []*Proxy{
		{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true},
		{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: false},
		{URL: "http://proxy3.example.com:8080", Type: "http", Healthy: true},
	}

	healthy := filterHealthy(proxies)
	assert.Equal(t, 2, len(healthy))
	assert.True(t, healthy[0].Healthy)
	assert.True(t, healthy[1].Healthy)
}

func TestCalculateSuccessRate(t *testing.T) {
	// 测试无使用记录
	p1 := &Proxy{UseCount: 0, FailCount: 0}
	rate1 := calculateSuccessRate(p1)
	assert.Equal(t, 1.0, rate1)

	// 测试部分失败
	p2 := &Proxy{UseCount: 10, FailCount: 2}
	rate2 := calculateSuccessRate(p2)
	assert.Equal(t, 0.8, rate2)

	// 测试全部失败
	p3 := &Proxy{UseCount: 5, FailCount: 5}
	rate3 := calculateSuccessRate(p3)
	assert.Equal(t, 0.0, rate3)
}

func TestProxyAddTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	tool := &proxyAddTool{module: module}

	// 测试工具信息
	info, err := tool.Info(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "proxy_add", info.Name)

	// 测试添加代理
	input := `{"url":"http://proxy.example.com:8080","type":"http","username":"user","password":"pass"}`
	output, err := tool.InvokableRun(context.Background(), input)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))

	// 测试无效输入
	invalidInput := `{"url":"","type":""}`
	_, err = tool.InvokableRun(context.Background(), invalidInput)
	assert.Error(t, err)
}

func TestProxyRemoveTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 先添加一个代理
	module.manager.AddProxy("http://proxy.example.com:8080", "http", "", "")

	tool := &proxyRemoveTool{module: module}

	// 测试移除代理
	input := `{"url":"http://proxy.example.com:8080"}`
	output, err := tool.InvokableRun(context.Background(), input)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
}

func TestProxyListTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 添加一些代理
	module.manager.AddProxy("http://proxy1.example.com:8080", "http", "", "")
	module.manager.AddProxy("http://proxy2.example.com:8080", "http", "", "")

	tool := &proxyListTool{module: module}

	// 测试列出代理
	output, err := tool.InvokableRun(context.Background(), "{}")
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, float64(2), result["count"].(float64))
}

func TestProxyGetTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 添加一个代理
	module.manager.AddProxy("http://proxy.example.com:8080", "http", "", "")

	tool := &proxyGetTool{module: module}

	// 测试获取代理
	output, err := tool.InvokableRun(context.Background(), "{}")
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.NotNil(t, result["proxy"])
}

func TestProxySetStrategyTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	tool := &proxySetStrategyTool{module: module}

	// 测试设置策略
	input := `{"strategy":"random"}`
	output, err := tool.InvokableRun(context.Background(), input)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "random", result["strategy"].(string))

	// 测试无效策略
	invalidInput := `{"strategy":"invalid"}`
	_, err = tool.InvokableRun(context.Background(), invalidInput)
	assert.Error(t, err)
}

func TestHealthChecker(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy.example.com:8080", Type: "http", Healthy: true},
		},
		strategy: &RoundRobinStrategy{},
	}

	checker := &HealthChecker{
		manager:  manager,
		interval: 1 * time.Second,
		testURL:  "https://www.google.com",
		stopChan: make(chan struct{}),
	}

	// 测试单个代理检查（注意：这会实际尝试连接，可能失败）
	checker.checkOne(manager.proxies[0])
	assert.NotNil(t, manager.proxies[0].LastCheck)
}

func TestGetTools(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	tools, err := module.GetTools(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 6, len(tools))

	// 测试禁用模块
	disabledConfig := ProxyConfig{
		Enable: false,
	}
	disabledModule, err := NewProxyModule(disabledConfig)
	require.NoError(t, err)
	defer disabledModule.Close()

	disabledTools, err := disabledModule.GetTools(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(disabledTools))
}

func TestCreateHTTPClient(t *testing.T) {
	proxy := &Proxy{
		URL:      "http://proxy.example.com:8080",
		Type:     "http",
		Username: "user",
		Password: "pass",
	}

	client, err := createHTTPClient(proxy, 10*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 10*time.Second, client.Timeout)
}

func TestCreateSOCKS5Dialer(t *testing.T) {
	proxy := &Proxy{
		URL:      "socks5://proxy.example.com:1080",
		Type:     "socks5",
		Username: "user",
		Password: "pass",
	}

	dialer, err := CreateSOCKS5Dialer(proxy)
	assert.NoError(t, err)
	assert.NotNil(t, dialer)
}

func TestProxyModule_CreateHTTPClient(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 添加一个代理
	module.manager.AddProxy("http://proxy.example.com:8080", "http", "", "")

	// 测试创建 HTTP 客户端
	client, err := module.CreateHTTPClient(10 * time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestProxyModule_CreateSOCKS5Client(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 添加一个 SOCKS5 代理
	module.manager.AddProxy("socks5://proxy.example.com:1080", "socks5", "", "")

	// 测试创建 SOCKS5 客户端
	dialer, err := module.CreateSOCKS5Client()
	assert.NoError(t, err)
	assert.NotNil(t, dialer)

	// 测试非 SOCKS5 代理
	module.manager.AddProxy("http://proxy2.example.com:8080", "http", "", "")
	module.manager.RemoveProxy("socks5://proxy.example.com:1080")

	_, err = module.CreateSOCKS5Client()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not SOCKS5 type")
}

func TestProxyHealthCheckTool(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 添加一个代理
	module.manager.AddProxy("http://proxy.example.com:8080", "http", "", "")

	tool := &proxyHealthCheckTool{module: module}

	// 测试检查所有代理
	output, err := tool.InvokableRun(context.Background(), "{}")
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))

	// 测试检查特定代理
	input := `{"url":"http://proxy.example.com:8080"}`
	output, err = tool.InvokableRun(context.Background(), input)
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))

	// 测试检查不存在的代理
	invalidInput := `{"url":"http://nonexistent.example.com:8080"}`
	_, err = tool.InvokableRun(context.Background(), invalidInput)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProxyManager_MarkSuccess(t *testing.T) {
	manager := &ProxyManager{
		proxies: []*Proxy{
			{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: true, UseCount: 10, FailCount: 2},
		},
		strategy: &RoundRobinStrategy{},
	}

	// 标记成功
	manager.MarkSuccess("http://proxy1.example.com:8080")
	assert.Equal(t, 0.8, manager.proxies[0].SuccessRate)
}

func TestHealthChecker_Stop(t *testing.T) {
	manager := &ProxyManager{
		proxies:  make([]*Proxy, 0),
		strategy: &RoundRobinStrategy{},
	}

	checker := &HealthChecker{
		manager:  manager,
		interval: 1 * time.Second,
		testURL:  "https://www.google.com",
		stopChan: make(chan struct{}),
	}

	// 启动健康检查
	go checker.Start()

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	// 停止健康检查
	checker.Stop()

	// 验证停止成功（不会panic）
	time.Sleep(100 * time.Millisecond)
}

func TestNewProxyModule_DefaultValues(t *testing.T) {
	config := ProxyConfig{
		Enable: true,
	}

	module, err := NewProxyModule(config)
	require.NoError(t, err)
	require.NotNil(t, module)

	// 验证默认值
	assert.Equal(t, 5, module.config.PoolSize)
	assert.Equal(t, 60, module.config.HealthCheckInterval)
	assert.Equal(t, "https://www.google.com", module.config.HealthCheckURL)
	assert.Equal(t, "round_robin", module.config.Strategy)

	module.Close()
}

func TestProxyModule_Close(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)

	// 测试关闭
	err = module.Close()
	assert.NoError(t, err)
}

func TestStrategyWithNoHealthyProxies(t *testing.T) {
	proxies := []*Proxy{
		{URL: "http://proxy1.example.com:8080", Type: "http", Healthy: false},
		{URL: "http://proxy2.example.com:8080", Type: "http", Healthy: false},
	}

	// 测试轮询策略
	roundRobin := &RoundRobinStrategy{}
	_, err := roundRobin.Select(proxies)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy proxy")

	// 测试随机策略
	random := &RandomStrategy{}
	_, err = random.Select(proxies)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy proxy")

	// 测试最少使用策略
	leastUsed := &LeastUsedStrategy{}
	_, err = leastUsed.Select(proxies)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy proxy")
}

func TestProxyModule_CreateSOCKS5Client_NoProxy(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 测试没有代理时创建客户端
	_, err = module.CreateSOCKS5Client()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no proxy available")
}

func TestProxyModule_CreateHTTPClient_NoProxy(t *testing.T) {
	config := ProxyConfig{
		Enable:   true,
		Strategy: "round_robin",
	}
	module, err := NewProxyModule(config)
	require.NoError(t, err)
	defer module.Close()

	// 测试没有代理时创建客户端
	_, err = module.CreateHTTPClient(10 * time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no proxy available")
}
