// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MCPOptimizedClient 优化的 MCP 客户端，支持连接池和批处理
type MCPOptimizedClient struct {
	clients    []*MCPClient          // 客户端连接池
	current    int                   // 当前使用的客户端索引
	mu         sync.RWMutex          // 读写锁
	requestQueue chan *MCPRequest    // 请求队列
	batchSize   int                  // 批处理大小
	batchTimeout time.Duration       // 批处理超时时间
}

// NewMCPOptimizedClient 创建一个新的优化 MCP 客户端
func NewMCPOptimizedClient(address string, poolSize int) *MCPOptimizedClient {
	if poolSize <= 0 {
		poolSize = 3 // 默认连接池大小
	}

	client := &MCPOptimizedClient{
		clients:      make([]*MCPClient, poolSize),
		current:      0,
		requestQueue: make(chan *MCPRequest, 100),
		batchSize:    10,            // 默认批处理大小
		batchTimeout: 50 * time.Millisecond, // 默认批处理超时
	}

	// 初始化客户端连接池
	for i := 0; i < poolSize; i++ {
		client.clients[i] = NewMCPClient(address)
	}

	// 启动批处理 goroutine
	go client.batchProcessor()

	return client
}

// GetClient 获取一个可用的客户端（轮询方式）
func (c *MCPOptimizedClient) GetClient() *MCPClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	client := c.clients[c.current]
	c.current = (c.current + 1) % len(c.clients)
	return client
}

// BatchCall 批量调用多个工具
func (c *MCPOptimizedClient) BatchCall(ctx context.Context, requests []MCPRequest) ([]MCPResponse, error) {
	responses := make([]MCPResponse, len(requests))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 使用连接池并发执行请求
	for i, req := range requests {
		wg.Add(1)
		go func(index int, request MCPRequest) {
			defer wg.Add(-1)

			client := c.GetClient()
			resp, err := c.sendRequest(ctx, client, request)
			if err != nil {
				resp = &MCPResponse{
					JSONRPC: request.JSONRPC,
					ID:      request.ID,
					Error: &MCPError{
						Code:    -32603,
						Message: err.Error(),
					},
				}
			}

			mu.Lock()
			responses[index] = *resp
			mu.Unlock()
		}(i, req)
	}

	wg.Wait()
	return responses, nil
}

// sendRequest 发送单个请求
func (c *MCPOptimizedClient) sendRequest(ctx context.Context, client *MCPClient, req MCPRequest) (*MCPResponse, error) {
	// 这里复用现有的 MCP 客户端的请求发送逻辑
	// 为了简化，这里提供一个基本实现
	_, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 创建一个模拟的响应（实际应用中应该通过网络发送请求）
	return &MCPResponse{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  json.RawMessage(fmt.Sprintf("{\"response\": \"processed request %v\"}", req.ID)),
	}, nil
}

// batchProcessor 批处理请求
func (c *MCPOptimizedClient) batchProcessor() {
	batch := make([]*MCPRequest, 0, c.batchSize)
	ticker := time.NewTicker(c.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case req := <-c.requestQueue:
			batch = append(batch, req)
			if len(batch) >= c.batchSize {
				// 达到批处理大小，执行批处理
				c.processBatch(batch)
				batch = make([]*MCPRequest, 0, c.batchSize)
			}
		case <-ticker.C:
			// 超时，执行批处理
			if len(batch) > 0 {
				c.processBatch(batch)
				batch = make([]*MCPRequest, 0, c.batchSize)
			}
		}
	}
}

// processBatch 处理一批请求
func (c *MCPOptimizedClient) processBatch(batch []*MCPRequest) {
	if len(batch) == 0 {
		return
	}

	// 将批处理请求转换为标准请求格式
	requests := make([]MCPRequest, len(batch))
	for i, req := range batch {
		requests[i] = *req
	}

	// 使用批处理调用
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.BatchCall(ctx, requests)
	if err != nil {
		fmt.Printf("批处理失败: %v\n", err)
	}
}

// MCPCompressedRequest 压缩的 MCP 请求
type MCPCompressedRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Compressed bool    `json:"compressed"`
}

// MCPCompressedResponse 压缩的 MCP 响应
type MCPCompressedResponse struct {
	JSONRPC  string          `json:"jsonrpc"`
	ID       interface{}     `json:"id"`
	Result   json.RawMessage `json:"result,omitempty"`
	Error    *MCPError       `json:"error,omitempty"`
	Compressed bool          `json:"compressed"`
}

// MCPConnectionMetrics MCP 连接指标
type MCPConnectionMetrics struct {
	TotalRequests    int64         // 总请求数
	SuccessRequests  int64         // 成功请求数
	FailedRequests   int64         // 失败请求数
	AverageLatency   time.Duration // 平均延迟
	LastRequestTime  time.Time     // 最后请求时间
	Throughput       float64       // 吞吐量（请求/秒）
}

// MCPConnectionStats MCP 连接统计
type MCPConnectionStats struct {
	mu       sync.RWMutex
	metrics  map[string]*MCPConnectionMetrics
	startTime time.Time
}

// NewMCPConnectionStats 创建新的 MCP 连接统计
func NewMCPConnectionStats() *MCPConnectionStats {
	return &MCPConnectionStats{
		metrics:   make(map[string]*MCPConnectionMetrics),
		startTime: time.Now(),
	}
}

// RecordRequest 记录请求
func (s *MCPConnectionStats) RecordRequest(clientID string, success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.metrics[clientID]; !exists {
		s.metrics[clientID] = &MCPConnectionMetrics{
			LastRequestTime: time.Now(),
		}
	}

	metrics := s.metrics[clientID]
	metrics.TotalRequests++
	metrics.LastRequestTime = time.Now()

	if success {
		metrics.SuccessRequests++
	} else {
		metrics.FailedRequests++
	}

	// 更新平均延迟
	totalDuration := metrics.AverageLatency * time.Duration(metrics.TotalRequests-1)
	metrics.AverageLatency = (totalDuration + latency) / time.Duration(metrics.TotalRequests)

	// 计算吞吐量
	elapsed := time.Since(s.startTime).Seconds()
	if elapsed > 0 {
		metrics.Throughput = float64(metrics.TotalRequests) / elapsed
	}
}

// GetMetrics 获取指标
func (s *MCPConnectionStats) GetMetrics(clientID string) (*MCPConnectionMetrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, exists := s.metrics[clientID]
	if !exists {
		return nil, false
	}

	// 返回副本以避免外部修改
	copy := *metrics
	return &copy, true
}

// GetAllMetrics 获取所有指标
func (s *MCPConnectionStats) GetAllMetrics() map[string]*MCPConnectionMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copyMetrics := make(map[string]*MCPConnectionMetrics)
	for id, metrics := range s.metrics {
		metricCopy := *metrics
		copyMetrics[id] = &metricCopy
	}

	return copyMetrics
}

// Reset 重置统计
func (s *MCPConnectionStats) Reset(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clientID == "" {
		// 重置所有统计
		s.metrics = make(map[string]*MCPConnectionMetrics)
		s.startTime = time.Now()
	} else {
		// 重置指定客户端的统计
		delete(s.metrics, clientID)
	}
}
