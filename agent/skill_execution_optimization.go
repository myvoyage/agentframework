// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncSkillExecutor 异步技能执行器
type AsyncSkillExecutor struct {
	executionQueue chan *SkillExecutionRequest
	workers       int
	workerPool    chan chan *SkillExecutionRequest
	results       map[string]*SkillExecutionResult
	resultMutex   sync.RWMutex
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// SkillExecutionRequest 技能执行请求
type SkillExecutionRequest struct {
	ID        string
	SkillName string
	Arguments interface{}
	Callback  func(*SkillExecutionResult)
	Timeout   time.Duration
}

// SkillExecutionResult 技能执行结果
type SkillExecutionResult struct {
	ID        string
	SkillName string
	Result    interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// NewAsyncSkillExecutor 创建一个新的异步技能执行器
func NewAsyncSkillExecutor(workers int) *AsyncSkillExecutor {
	if workers <= 0 {
		workers = 5 // 默认工作线程数
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &AsyncSkillExecutor{
		executionQueue: make(chan *SkillExecutionRequest, 1000),
		workers:       workers,
		workerPool:    make(chan chan *SkillExecutionRequest, workers),
		results:       make(map[string]*SkillExecutionResult),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 启动工作线程
	for i := 0; i < workers; i++ {
		executor.wg.Add(1)
		go executor.worker(i)
	}

	// 启动结果清理器
	go executor.resultCleaner()

	return executor
}

// worker 工作线程
func (e *AsyncSkillExecutor) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case job := <-e.executionQueue:
			if job == nil {
				return
			}
			e.executeJob(job)
		}
	}
}

// executeJob 执行任务
func (e *AsyncSkillExecutor) executeJob(job *SkillExecutionRequest) {
	startTime := time.Now()

	// 设置超时
	ctx := e.ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, job.Timeout)
		defer cancel()
	} else {
		// 默认超时 30 秒
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(e.ctx, 30*time.Second)
		defer cancel()
	}

	// 执行技能（这里简化实现）
	result := &SkillExecutionResult{
		ID:        job.ID,
		SkillName: job.SkillName,
		Timestamp: startTime,
	}

	// 模拟执行
	select {
	case <-ctx.Done():
		result.Error = fmt.Errorf("execution timeout")
	case <-time.After(10 * time.Millisecond):
		// 模拟执行成功
		result.Result = fmt.Sprintf("执行结果: %s", job.SkillName)
	}

	result.Duration = time.Since(startTime)

	// 保存结果
	e.resultMutex.Lock()
	e.results[job.ID] = result
	e.resultMutex.Unlock()

	// 调用回调函数
	if job.Callback != nil {
		job.Callback(result)
	}
}

// Submit 提交任务
func (e *AsyncSkillExecutor) Submit(skillName string, arguments interface{}, callback func(*SkillExecutionResult), timeout time.Duration) (string, error) {
	requestID := fmt.Sprintf("%s-%d", skillName, time.Now().UnixNano())

	req := &SkillExecutionRequest{
		ID:        requestID,
		SkillName: skillName,
		Arguments: arguments,
		Callback:  callback,
		Timeout:   timeout,
	}

	select {
	case e.executionQueue <- req:
		return requestID, nil
	case <-e.ctx.Done():
		return "", fmt.Errorf("executor is shutting down")
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("execution queue is full")
	}
}

// GetResult 获取结果
func (e *AsyncSkillExecutor) GetResult(requestID string) (*SkillExecutionResult, bool) {
	e.resultMutex.RLock()
	defer e.resultMutex.RUnlock()

	result, exists := e.results[requestID]
	if !exists {
		return nil, false
	}

	// 返回副本
	resultCopy := *result
	return &resultCopy, true
}

// WaitForResult 等待结果
func (e *AsyncSkillExecutor) WaitForResult(requestID string, timeout time.Duration) (*SkillExecutionResult, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if result, exists := e.GetResult(requestID); exists {
			return result, nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for result")
}

// resultCleaner 结果清理器
func (e *AsyncSkillExecutor) resultCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.resultMutex.Lock()
			// 清理超过 10 分钟的结果
			cutoff := time.Now().Add(-10 * time.Minute)
			for id, result := range e.results {
				if result.Timestamp.Before(cutoff) {
					delete(e.results, id)
				}
			}
			e.resultMutex.Unlock()
		}
	}
}

// Shutdown 关闭执行器
func (e *AsyncSkillExecutor) Shutdown(timeout time.Duration) error {
	// 停止接受新任务
	close(e.executionQueue)

	// 关闭工作线程
	close(e.workerPool)

	// 等待所有工作线程完成
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		e.cancel()
		return fmt.Errorf("shutdown timeout")
	}
}

// GetStats 获取统计信息
func (e *AsyncSkillExecutor) GetStats() map[string]interface{} {
	e.resultMutex.RLock()
	defer e.resultMutex.RUnlock()

	stats := map[string]interface{}{
		"queue_size":    len(e.executionQueue),
		"workers":       e.workers,
		"results_count":  len(e.results),
	}

	// 计算平均执行时间
	var totalDuration time.Duration
	var count int
	for _, result := range e.results {
		totalDuration += result.Duration
		count++
	}

	if count > 0 {
		stats["average_duration"] = totalDuration / time.Duration(count)
		stats["total_executions"] = count
	}

	return stats
}

// SkillCache 技能缓存
type SkillCache struct {
	cache      map[string]*CachedSkillResult
	mu         sync.RWMutex
	ttl        time.Duration
	maxSize    int
	cleanupInterval time.Duration
	ctx        context.Context
	cancel    context.CancelFunc
}

// CachedSkillResult 缓存的技能结果
type CachedSkillResult struct {
	Result     interface{}
	Error      error
	Expiration time.Time
	HitCount   int64
	LastAccess time.Time
}

// NewSkillCache 创建一个新的技能缓存
func NewSkillCache(ttl time.Duration, maxSize int) *SkillCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute // 默认 TTL 5 分钟
	}
	if maxSize <= 0 {
		maxSize = 1000 // 默认最大缓存大小
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &SkillCache{
		cache:          make(map[string]*CachedSkillResult),
		ttl:            ttl,
		maxSize:        maxSize,
		cleanupInterval: 1 * time.Minute,
		ctx:            ctx,
		cancel:         cancel,
	}

	// 启动清理器
	go cache.cleanup()

	return cache
}

// Get 获取缓存结果
func (c *SkillCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[key]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}

	// 检查是否过期
	if time.Now().After(cached.Expiration) {
		return nil, fmt.Errorf("cache expired")
	}

	// 更新访问时间
	cached.LastAccess = time.Now()
	cached.HitCount++

	return cached.Result, cached.Error
}

// Set 设置缓存结果
func (c *SkillCache) Set(key string, result interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小
	if len(c.cache) >= c.maxSize {
		c.evict()
	}

	c.cache[key] = &CachedSkillResult{
		Result:     result,
		Error:      err,
		Expiration: time.Now().Add(c.ttl),
		LastAccess: time.Now(),
		HitCount:   0,
	}
}

// Invalidate 使缓存失效
func (c *SkillCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

// Clear 清空缓存
func (c *SkillCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedSkillResult)
}

// cleanup 清理过期缓存
func (c *SkillCache) cleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, cached := range c.cache {
				if now.After(cached.Expiration) {
					delete(c.cache, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

// evict 淘汰缓存
func (c *SkillCache) evict() {
	// 使用 LRU 策略淘汰缓存
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range c.cache {
		if oldestKey == "" || cached.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.LastAccess
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// GetStats 获取缓存统计
func (c *SkillCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalHits int64
	for _, cached := range c.cache {
		totalHits += cached.HitCount
	}

	return map[string]interface{}{
		"size":       len(c.cache),
		"max_size":   c.maxSize,
		"ttl":        c.ttl.String(),
		"total_hits": totalHits,
	}
}

// Close 关闭缓存
func (c *SkillCache) Close() {
	c.cancel()
}
