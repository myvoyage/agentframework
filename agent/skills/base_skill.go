// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// ExecutionContext 技能执行上下文，提供执行环境和工具
type ExecutionContext struct {
	RequestID   string                 // 请求ID，用于追踪
	StartTime   time.Time              // 开始时间
	Timeout     time.Duration          // 超时时间
	Environment map[string]string      // 环境变量
	Workspace   string                 // 工作目录
	Metadata    map[string]interface{} // 元数据
	mu          sync.RWMutex           // 读写锁
}

// NewExecutionContext 创建新的执行上下文
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		StartTime:   time.Now(),
		Environment: make(map[string]string),
		Workspace:   ".",
		Metadata:    make(map[string]interface{}),
	}
}

// GetEnv 获取环境变量
func (ec *ExecutionContext) GetEnv(key string) string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.Environment[key]
}

// SetEnv 设置环境变量
func (ec *ExecutionContext) SetEnv(key, value string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.Environment[key] = value
}

// GetMetadata 获取元数据
func (ec *ExecutionContext) GetMetadata(key string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	val, ok := ec.Metadata[key]
	return val, ok
}

// SetMetadata 设置元数据
func (ec *ExecutionContext) SetMetadata(key string, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.Metadata[key] = value
}

// ExecutionResult 技能执行结果
type ExecutionResult struct {
	Success    bool                   `json:"success"`     // 是否成功
	Data       interface{}            `json:"data"`        // 结果数据
	Error      string                 `json:"error"`       // 错误信息
	Duration   time.Duration          `json:"duration"`    // 执行时长
	Bytes      int64                  `json:"bytes"`       // 处理字节数
	Metadata   map[string]interface{} `json:"metadata"`    // 额外元数据
	ExecutedAt time.Time              `json:"executed_at"` // 执行时间
}

// ToJSON 转换为JSON字符串
func (er *ExecutionResult) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(er, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(bytes), nil
}

// BaseSkill 增强的基础技能模板
type BaseSkill struct {
	name        string
	description string
	version     string
	category    string
	enabled     bool
	config      SkillConfig
	metadata    SkillMetadata
	stats       *SkillStats
	mu          sync.RWMutex
}

// SkillConfig 技能配置
type SkillConfig struct {
	MaxOutputSize    int64         // 最大输出大小（字节）
	MaxExecutionTime time.Duration // 最大执行时间
	EnableCache      bool          // 是否启用缓存
	CacheTTL         time.Duration // 缓存过期时间
	AllowedPaths     []string      // 允许访问的路径（文件操作）
	AllowedHosts     []string      // 允许访问的主机（网络操作）
	EnableRateLimit  bool          // 是否启用速率限制
	RateLimit        int           // 每分钟最大调用次数
}

// SkillStats 技能统计信息
type SkillStats struct {
	TotalCalls      int64         // 总调用次数
	SuccessCalls    int64         // 成功次数
	FailedCalls     int64         // 失败次数
	TotalDuration   time.Duration // 总执行时长
	AverageDuration time.Duration // 平均执行时长
	LastCalled      time.Time     // 最后调用时间
	LastError       string        // 最后错误信息
	LastErrorTime   time.Time     // 最后错误时间
	CacheHits       int64         // 缓存命中次数
	CacheMisses     int64         // 缓存未命中次数
	mu              sync.RWMutex  // 读写锁
}

// NewBaseSkill 创建新的基础技能
func NewBaseSkill(name, description string) *BaseSkill {
	return &BaseSkill{
		name:        name,
		description: description,
		version:     "1.0.0",
		category:    "general",
		enabled:     true,
		config: SkillConfig{
			MaxOutputSize:    10 * 1024 * 1024, // 10MB
			MaxExecutionTime: 30 * time.Second,
			EnableCache:      false,
			EnableRateLimit:  false,
			RateLimit:        60,
		},
		metadata: SkillMetadata{
			Name:        name,
			Version:     "1.0.0",
			Author:      "Agent Framework Contributors",
			Description: description,
			Category:    "general",
			Tags:        []string{"base"},
		},
		stats: &SkillStats{},
	}
}

// GetName 获取技能名称
func (s *BaseSkill) GetName() string {
	return s.name
}

// GetDescription 获取技能描述
func (s *BaseSkill) GetDescription() string {
	return s.description
}

// GetVersion 获取技能版本
func (s *BaseSkill) GetVersion() string {
	return s.version
}

// GetCategory 获取技能分类
func (s *BaseSkill) GetCategory() string {
	return s.category
}

// Info 返回技能信息（由子类实现具体参数）
func (s *BaseSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.name,
		Desc: s.description,
	}, nil
}

// Invoke 执行技能（由子类实现具体逻辑）
func (s *BaseSkill) Invoke(ctx context.Context, input string) (string, error) {
	return "", fmt.Errorf("skill %s is not implemented", s.name)
}

// IsEnabled 检查技能是否启用
func (s *BaseSkill) IsEnabled(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetEnabled 设置技能是否启用
func (s *BaseSkill) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// GetMetadata 获取技能元数据
func (s *BaseSkill) GetMetadata(ctx context.Context) SkillMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metadata
}

// SetMetadata 设置技能元数据
func (s *BaseSkill) SetMetadata(metadata SkillMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata = metadata
}

// GetConfig 获取技能配置
func (s *BaseSkill) GetConfig() SkillConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig 设置技能配置
func (s *BaseSkill) SetConfig(config SkillConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetStats 获取技能统计信息
func (s *BaseSkill) GetStats() *SkillStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// RecordCall 记录一次调用
func (s *BaseSkill) RecordCall(success bool, duration time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalCalls++
	s.stats.LastCalled = time.Now()
	s.stats.TotalDuration += duration

	if success {
		s.stats.SuccessCalls++
	} else {
		s.stats.FailedCalls++
		if err != nil {
			s.stats.LastError = err.Error()
			s.stats.LastErrorTime = time.Now()
		}
	}

	// 更新平均时长
	if s.stats.TotalCalls > 0 {
		s.stats.AverageDuration = time.Duration(int64(s.stats.TotalDuration) / s.stats.TotalCalls)
	}
}

// ValidateInput 验证输入参数
func (s *BaseSkill) ValidateInput(input string) error {
	if input == "" {
		return fmt.Errorf("input cannot be empty")
	}

	// 检查输入大小
	if int64(len(input)) > s.config.MaxOutputSize {
		return fmt.Errorf("input size exceeds maximum allowed size")
	}

	return nil
}

// CreateExecutionContext 创建执行上下文
func (s *BaseSkill) CreateExecutionContext(ctx context.Context) (*ExecutionContext, context.CancelFunc, error) {
	execCtx := NewExecutionContext()
	execCtx.RequestID = fmt.Sprintf("%s-%d", s.name, time.Now().UnixNano())
	if len(s.config.AllowedPaths) > 0 {
		execCtx.Workspace = s.config.AllowedPaths[0] // 使用第一个允许的路径
	} else {
		execCtx.Workspace = "." // 默认工作目录
	}

	// 设置超时
	if s.config.MaxExecutionTime > 0 {
		_, cancel := context.WithTimeout(ctx, s.config.MaxExecutionTime)
		return execCtx, cancel, nil
	}

	return execCtx, func() {}, nil
}

// ExecuteTemplate 执行模板方法，提供统一的执行流程
func (s *BaseSkill) ExecuteTemplate(
	ctx context.Context,
	input string,
	executeFunc func(*ExecutionContext) (interface{}, error),
) *ExecutionResult {
	startTime := time.Now()
	result := &ExecutionResult{
		ExecutedAt: startTime,
	}

	// 验证输入
	if err := s.ValidateInput(input); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		s.RecordCall(false, result.Duration, err)
		return result
	}

	// 创建执行上下文
	execCtx, cancel, err := s.CreateExecutionContext(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to create execution context: %v", err)
		result.Duration = time.Since(startTime)
		s.RecordCall(false, result.Duration, err)
		return result
	}
	defer cancel()

	// 执行技能逻辑
	data, err := executeFunc(execCtx)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		s.RecordCall(false, result.Duration, err)
		return result
	}

	// 成功
	result.Success = true
	result.Data = data
	result.Duration = time.Since(startTime)

	// 计算输出大小
	if jsonBytes, err := json.Marshal(data); err == nil {
		result.Bytes = int64(len(jsonBytes))
	}

	s.RecordCall(true, result.Duration, nil)
	return result
}

// SkillExecutor 技能执行器接口，用于扩展技能功能
type SkillExecutor interface {
	Validate(ctx context.Context, input string) error
	Prepare(ctx context.Context, input string) (*ExecutionContext, error)
	Execute(ctx context.Context, execCtx *ExecutionContext) (interface{}, error)
	Cleanup(ctx context.Context, execCtx *ExecutionContext) error
}

// AdvancedSkill 高级技能基类，提供更完整的功能
type AdvancedSkill struct {
	*BaseSkill
	executor SkillExecutor
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
}

// CacheEntry 缓存条目

// NewAdvancedSkill 创建新的高级技能
func NewAdvancedSkill(name, description string, executor SkillExecutor) *AdvancedSkill {
	return &AdvancedSkill{
		BaseSkill: NewBaseSkill(name, description),
		executor:  executor,
		cache:     make(map[string]*CacheEntry),
	}
}

// Invoke 实现技能调用
func (s *AdvancedSkill) Invoke(ctx context.Context, input string) (string, error) {
	// 检查缓存
	if s.GetConfig().EnableCache {
		if cached := s.getFromCache(input); cached != nil {
			s.GetStats().CacheHits++
			result := &ExecutionResult{
				Success:  true,
				Data:     cached.Data,
				Duration: 0,
				Metadata: map[string]interface{}{
					"cached": true,
				},
			}
			return result.ToJSON()
		}
		s.GetStats().CacheMisses++
	}

	// 执行技能
	result := s.ExecuteTemplate(ctx, input, func(execCtx *ExecutionContext) (interface{}, error) {
		// 准备阶段
		if s.executor != nil {
			if _, err := s.executor.Prepare(ctx, input); err != nil {
				return nil, fmt.Errorf("prepare failed: %w", err)
			}
		}

		// 执行阶段
		if s.executor != nil {
			return s.executor.Execute(ctx, execCtx)
		}

		return nil, fmt.Errorf("no executor configured")
	})

	// 缓存结果
	if result.Success && s.GetConfig().EnableCache {
		s.setToCache(input, result.Data)
	}

	return result.ToJSON()
}

// getFromCache 从缓存获取
func (s *AdvancedSkill) getFromCache(key string) *CacheEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.cache[key]
	if !ok {
		return nil
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		delete(s.cache, key)
		return nil
	}

	return entry
}

// setToCache 设置缓存
func (s *AdvancedSkill) setToCache(key string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(s.GetConfig().CacheTTL),
	}
}

// ClearCache 清除缓存
func (s *AdvancedSkill) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*CacheEntry)
}
