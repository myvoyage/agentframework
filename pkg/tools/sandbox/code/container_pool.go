// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ContainerPool 容器池
type ContainerPool struct {
	pools    map[string]*LanguagePool // 每种语言一个池
	config   PoolConfig
	executor *ContainerExecutor
	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// PoolConfig 容器池配置
type PoolConfig struct {
	MinSize             int           // 最小容器数
	MaxSize             int           // 最大容器数
	IdleTimeout         time.Duration // 空闲超时
	HealthCheckInterval time.Duration // 健康检查间隔
}

// LanguagePool 语言容器池
type LanguagePool struct {
	language   string
	containers chan *PooledContainer
	active     map[string]*PooledContainer
	mu         sync.RWMutex
	stats      PoolStats
}

// PooledContainer 池化容器
type PooledContainer struct {
	ID         string
	Language   string
	CreatedAt  time.Time
	LastUsedAt time.Time
	UseCount   int64
	Healthy    bool
}

// PoolStats 池统计
type PoolStats struct {
	TotalCreated   int64
	TotalDestroyed int64
	CurrentSize    int
	ActiveCount    int
	IdleCount      int
	ReuseCount     int64
}

// NewContainerPool 创建容器池
func NewContainerPool(executor *ContainerExecutor, config PoolConfig) *ContainerPool {
	pool := &ContainerPool{
		pools:    make(map[string]*LanguagePool),
		config:   config,
		executor: executor,
		stopChan: make(chan struct{}),
	}

	// 为每种语言创建池
	for _, lang := range []string{"python", "javascript", "go", "bash"} {
		pool.pools[lang] = &LanguagePool{
			language:   lang,
			containers: make(chan *PooledContainer, config.MaxSize),
			active:     make(map[string]*PooledContainer),
		}
	}

	// 启动维护协程
	pool.wg.Add(1)
	go pool.maintain()

	return pool
}

// Acquire 获取容器
func (p *ContainerPool) Acquire(ctx context.Context, language string) (*PooledContainer, error) {
	langPool, ok := p.pools[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// 尝试从池中获取
	select {
	case container := <-langPool.containers:
		// 检查容器健康状态
		if container.Healthy {
			container.LastUsedAt = time.Now()
			container.UseCount++

			langPool.mu.Lock()
			langPool.active[container.ID] = container
			langPool.stats.ReuseCount++
			langPool.mu.Unlock()

			return container, nil
		}
		// 容器不健康，销毁并创建新的
		p.destroyContainer(container)
	default:
		// 池为空，创建新容器
	}

	// 创建新容器
	return p.createContainer(ctx, language)
}

// Release 释放容器
func (p *ContainerPool) Release(container *PooledContainer) error {
	langPool, ok := p.pools[container.Language]
	if !ok {
		return fmt.Errorf("unknown language: %s", container.Language)
	}

	langPool.mu.Lock()
	delete(langPool.active, container.ID)
	langPool.mu.Unlock()

	// 检查容器是否健康
	if !container.Healthy {
		return p.destroyContainer(container)
	}

	// 检查是否超过最大空闲时间
	if time.Since(container.LastUsedAt) > p.config.IdleTimeout {
		return p.destroyContainer(container)
	}

	// 放回池中
	select {
	case langPool.containers <- container:
		return nil
	default:
		// 池已满，销毁容器
		return p.destroyContainer(container)
	}
}

// createContainer 创建容器
func (p *ContainerPool) createContainer(ctx context.Context, language string) (*PooledContainer, error) {
	// 使用 ContainerExecutor 创建容器
	image := p.executor.getImage(language)
	containerID, err := p.executor.createContainer(ctx, image, language, "")
	if err != nil {
		return nil, err
	}

	container := &PooledContainer{
		ID:         containerID,
		Language:   language,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UseCount:   0,
		Healthy:    true,
	}

	langPool := p.pools[language]
	langPool.mu.Lock()
	langPool.stats.TotalCreated++
	langPool.active[containerID] = container
	langPool.mu.Unlock()

	return container, nil
}

// destroyContainer 销毁容器
func (p *ContainerPool) destroyContainer(container *PooledContainer) error {
	err := p.executor.removeContainer(context.Background(), container.ID)

	langPool := p.pools[container.Language]
	langPool.mu.Lock()
	langPool.stats.TotalDestroyed++
	delete(langPool.active, container.ID)
	langPool.mu.Unlock()

	return err
}

// maintain 维护协程
func (p *ContainerPool) maintain() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.healthCheck()
			p.warmUp()
		case <-p.stopChan:
			return
		}
	}
}

// healthCheck 健康检查
func (p *ContainerPool) healthCheck() {
	for _, langPool := range p.pools {
		langPool.mu.RLock()
		containers := make([]*PooledContainer, 0, len(langPool.active))
		for _, container := range langPool.active {
			containers = append(containers, container)
		}
		langPool.mu.RUnlock()

		for _, container := range containers {
			// 检查容器健康状态
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := p.executor.client.ContainerInspect(ctx, container.ID)
			cancel()

			if err != nil {
				container.Healthy = false
			}
		}
	}
}

// warmUp 预热容器
func (p *ContainerPool) warmUp() {
	for lang, langPool := range p.pools {
		langPool.mu.RLock()
		currentSize := len(langPool.containers) + len(langPool.active)
		langPool.mu.RUnlock()

		// 如果容器数量少于最小值，创建新容器
		if currentSize < p.config.MinSize {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			container, err := p.createContainer(ctx, lang)
			cancel()

			if err == nil {
				langPool.mu.Lock()
				delete(langPool.active, container.ID)
				langPool.mu.Unlock()
				langPool.containers <- container
			}
		}
	}
}

// Close 关闭容器池
func (p *ContainerPool) Close() error {
	close(p.stopChan)
	p.wg.Wait()

	// 销毁所有容器
	for _, langPool := range p.pools {
		close(langPool.containers)
		for container := range langPool.containers {
			p.destroyContainer(container)
		}

		langPool.mu.Lock()
		for _, container := range langPool.active {
			p.destroyContainer(container)
		}
		langPool.mu.Unlock()
	}

	return nil
}

// GetStats 获取统计信息
func (p *ContainerPool) GetStats() map[string]PoolStats {
	stats := make(map[string]PoolStats)

	for lang, langPool := range p.pools {
		langPool.mu.RLock()
		langPool.stats.CurrentSize = len(langPool.containers) + len(langPool.active)
		langPool.stats.ActiveCount = len(langPool.active)
		langPool.stats.IdleCount = len(langPool.containers)
		stats[lang] = langPool.stats
		langPool.mu.RUnlock()
	}

	return stats
}
