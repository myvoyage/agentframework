// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package react

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent"
	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/errors"
)

// AgentFactory ReAct Agent工厂接口
// 【必须】定义Agent工厂的标准接口，支持创建和配置不同类型的Agent
type AgentFactory interface {
	// CreateReActAgent 创建ReAct Agent
	// 【必须】基于配置创建完整的ReAct Agent实例
	CreateReActAgent(ctx context.Context, cfg *ReActConfig, opts ...AgentOption) (ReActAgent, error)
	// CreateReActAgentWithProfile 基于配置文件创建Agent
	// 【必须】从配置文件创建预配置的Agent
	CreateReActAgentWithProfile(ctx context.Context, profileName string, customOpts ...AgentOption) (ReActAgent, error)
	// RegisterProfile 注册Agent配置文件
	// 【必须】注册预定义的Agent配置模板
	RegisterProfile(profileName string, config *ReActConfig) error
	// GetProfile 获取Agent配置文件
	// 【必须】获取已注册的配置模板
	GetProfile(profileName string) (*ReActConfig, error)
	// ListProfiles 列出所有配置文件
	// 【推荐】提供配置管理能力
	ListProfiles() []string
	// ValidateConfig 验证Agent配置
	// 【必须】验证配置的有效性和完整性
	ValidateConfig(cfg *ReActConfig) error
	// Name 返回工厂名称
	// 【必须】提供工厂标识
	Name() string
}

// BaseAgentFactory Agent工厂基类
// 【推荐】提供基础实现减少重复代码
type BaseAgentFactory struct {
	name   string
	logger *zap.Logger
	// profiles 配置模板
	profiles map[string]*ReActConfig
	// profilesMutex 配置模板访问互斥锁
	profilesMutex sync.RWMutex
	// defaultConfig 默认配置
	defaultConfig *ReActConfig
}

// NewBaseAgentFactory 创建基础Agent工厂
// 【必须】提供构造函数确保必要初始化
func NewBaseAgentFactory(name string, logger *zap.Logger) *BaseAgentFactory {
	if logger == nil {
		logger = zap.NewNop()
	}

	factory := &BaseAgentFactory{
		name:      name,
		logger:    logger.With(zap.String("factory", name)),
		profiles:  make(map[string]*ReActConfig),
		defaultConfig: NewReActConfig(),
	}

	// 注册默认配置模板
	factory.registerDefaultProfiles()

	return factory
}

// Name 返回工厂名称
// 【必须】实现 AgentFactory 接口的 Name 方法
func (baf *BaseAgentFactory) Name() string {
	return baf.name
}

// registerDefaultProfiles 注册默认配置模板
// 【推荐】注册内置的Agent配置模板
func (baf *BaseAgentFactory) registerDefaultProfiles() {
	// 轻量级Agent配置
	lightweightConfig := NewReActConfig()
	lightweightConfig.MaxIterations = 3
	lightweightConfig.EnableDetailedThinking = false
	lightweightConfig.EnableParallelExecution = false
	lightweightConfig.ToolTimeout = 15 * time.Second
	baf.profiles["lightweight"] = lightweightConfig

	// 研究型Agent配置
	researchConfig := NewReActConfig()
	researchConfig.MaxIterations = 10
	researchConfig.EnableDetailedThinking = true
	researchConfig.EnableParallelExecution = true
	researchConfig.ToolTimeout = 60 * time.Second
	researchConfig.ThoughtConfig.MinConfidence = 0.3
	researchConfig.PlannerConfig.MaxRetries = 5
	baf.profiles["research"] = researchConfig

	// 生产级Agent配置
	productionConfig := NewReActConfig()
	productionConfig.MaxIterations = 8
	productionConfig.EnableDetailedThinking = true
	productionConfig.EnableParallelExecution = true
	productionConfig.ToolTimeout = 45 * time.Second
	productionConfig.ThoughtConfig.MinConfidence = 0.5
	productionConfig.PlannerConfig.MaxRetries = 3
	productionConfig.ExecutorConfig.MaxConcurrentExecutions = 10
	baf.profiles["production"] = productionConfig

	// 调试模式Agent配置
	debugConfig := NewReActConfig()
	debugConfig.MaxIterations = 20
	debugConfig.EnableDetailedThinking = true
	debugConfig.EnableParallelExecution = false
	debugConfig.EnableMetrics = true
	debugConfig.EnableTracing = true
	debugConfig.LogLevel = "debug"
	baf.profiles["debug"] = debugConfig

	baf.logger.Debug("registered default agent profiles",
		zap.Strings("profiles", baf.ListProfiles()),
	)
}

// ValidateConfig 验证Agent配置
// 【必须】实现 AgentFactory 接口的 ValidateConfig 方法
func (baf *BaseAgentFactory) ValidateConfig(cfg *ReActConfig) error {
	if cfg == nil {
		return errors.NewValidationError("config cannot be nil", nil)
	}

	// 验证配置的各个组件
	if err := cfg.Validate(); err != nil {
		return errors.WrapError(err, "config validation failed", nil)
	}

	// 额外的工厂级验证
	if cfg.MaxIterations > 100 {
		return errors.NewValidationError(
			"max iterations too high, may cause infinite loops",
			map[string]interface{}{"max_iterations": cfg.MaxIterations},
		)
	}

	if cfg.ToolTimeout > 10*time.Minute {
		return errors.NewValidationError(
			"tool timeout too long, may block execution",
			map[string]interface{}{"tool_timeout": cfg.ToolTimeout},
		)
	}

	return nil
}

// RegisterProfile 注册Agent配置文件
// 【必须】实现 AgentFactory 接口的 RegisterProfile 方法
func (baf *BaseAgentFactory) RegisterProfile(profileName string, config *ReActConfig) error {
	if profileName == "" {
		return errors.NewValidationError("profile name cannot be empty", nil)
	}

	if config == nil {
		return errors.NewValidationError("config cannot be nil", nil)
	}

	// 验证配置
	if err := baf.ValidateConfig(config); err != nil {
		return errors.WrapError(err, "invalid profile configuration", map[string]interface{}{
			"profile_name": profileName,
		})
	}

	baf.profilesMutex.Lock()
	defer baf.profilesMutex.Unlock()

	baf.profiles[profileName] = config

	baf.logger.Info("registered agent profile",
		zap.String("profile_name", profileName),
		zap.Int("max_iterations", config.MaxIterations),
		zap.Bool("detailed_thinking", config.EnableDetailedThinking),
	)

	return nil
}

// GetProfile 获取Agent配置文件
// 【必须】实现 AgentFactory 接口的 GetProfile 方法
func (baf *BaseAgentFactory) GetProfile(profileName string) (*ReActConfig, error) {
	if profileName == "" {
		return nil, errors.NewValidationError("profile name cannot be empty", nil)
	}

	baf.profilesMutex.RLock()
	defer baf.profilesMutex.RUnlock()

	config, exists := baf.profiles[profileName]
	if !exists {
		return nil, errors.NewValidationError(
			"profile not found",
			map[string]interface{}{"profile_name": profileName},
		)
	}

	// 返回配置的副本以避免外部修改
	configCopy := *config
	return &configCopy, nil
}

// ListProfiles 列出所有配置文件
// 【推荐】实现 AgentFactory 接口的 ListProfiles 方法
func (baf *BaseAgentFactory) ListProfiles() []string {
	baf.profilesMutex.RLock()
	defer baf.profilesMutex.RUnlock()

	profiles := make([]string, 0, len(baf.profiles))
	for profileName := range baf.profiles {
		profiles = append(profiles, profileName)
	}

	return profiles
}

// DefaultReActAgentFactory 默认ReAct Agent工厂实现
// 【必须】提供开箱即用的Agent工厂实现
type DefaultReActAgentFactory struct {
	BaseAgentFactory
	// LLMFactory LLM工厂
	LLMFactory llm.Factory
	// ToolRegistry 工具注册表
	ToolRegistry registry.ToolRegistry
	// MemoryManager 内存管理器
	MemoryManager memory.Manager
	// configLoader 配置加载器
	configLoader *config.ConfigLoader
	// agents 创建的Agent实例
	agents map[string]ReActAgent
	// agentsMutex Agent实例访问互斥锁
	agentsMutex sync.RWMutex
}

// AgentOption Agent配置选项
// 【推荐】用于在创建Agent时动态配置参数
type AgentOption func(*agentCreationOptions)

// agentCreationOptions Agent创建选项
// 【推荐】封装Agent创建时的可选参数
type agentCreationOptions struct {
	// SessionID 自定义会话ID
	SessionID string
	// Logger 自定义日志记录器
	Logger *zap.Logger
	// ModelClient 自定义模型客户端
	ModelClient llm.Client
	// Planner 自定义规划器
	Planner PlanGenerator
	// Executor 自定义执行器
	Executor ActionExecutor
	// Memory 自定义内存集成器
	Memory MemoryIntegration
	// MetricsEnabled 是否启用指标收集
	MetricsEnabled bool
	// TracingEnabled 是否启用链路追踪
	TracingEnabled bool
}

// WithSessionID 设置自定义会话ID
// 【推荐】Agent配置选项：设置会话ID
func WithSessionID(sessionID string) AgentOption {
	return func(o *agentCreationOptions) {
		o.SessionID = sessionID
	}
}

// WithLogger 设置自定义日志记录器
// 【推荐】Agent配置选项：设置日志记录器
func WithLogger(logger *zap.Logger) AgentOption {
	return func(o *agentCreationOptions) {
		o.Logger = logger
	}
}

// WithModelClient 设置自定义模型客户端
// 【推荐】Agent配置选项：设置模型客户端
func WithModelClient(client llm.Client) AgentOption {
	return func(o *agentCreationOptions) {
		o.ModelClient = client
	}
}

// WithPlanner 设置自定义规划器
// 【推荐】Agent配置选项：设置规划器
func WithPlanner(planner PlanGenerator) AgentOption {
	return func(o *agentCreationOptions) {
		o.Planner = planner
	}
}

// WithExecutor 设置自定义执行器
// 【推荐】Agent配置选项：设置执行器
func WithExecutor(executor ActionExecutor) AgentOption {
	return func(o *agentCreationOptions) {
		o.Executor = executor
	}
}

// WithMemory 设置自定义内存集成器
// 【推荐】Agent配置选项：设置内存集成器
func WithMemory(memory MemoryIntegration) AgentOption {
	return func(o *agentCreationOptions) {
		o.Memory = memory
	}
}

// WithMetrics 启用指标收集
// 【推荐】Agent配置选项：启用指标收集
func WithMetrics(enabled bool) AgentOption {
	return func(o *agentCreationOptions) {
		o.MetricsEnabled = enabled
	}
}

// WithTracing 启用链路追踪
// 【推荐】Agent配置选项：启用链路追踪
func WithTracing(enabled bool) AgentOption {
	return func(o *agentCreationOptions) {
		o.TracingEnabled = enabled
	}
}

// NewDefaultReActAgentFactory 创建默认ReAct Agent工厂
// 【必须】提供构造函数确保必要依赖
func NewDefaultReActAgentFactory(logger *zap.Logger, llmFactory llm.Factory, toolRegistry registry.ToolRegistry, memoryManager memory.Manager) *DefaultReActAgentFactory {
	if logger == nil {
		logger = zap.NewNop()
	}

	factory := &DefaultReActAgentFactory{
		BaseAgentFactory: *NewBaseAgentFactory("default_react_agent_factory", logger),
		LLMFactory:       llmFactory,
		ToolRegistry:     toolRegistry,
		MemoryManager:    memoryManager,
		configLoader:     config.NewConfigLoader(logger),
		agents:          make(map[string]ReActAgent),
	}

	// 【必须】验证必要依赖
	if factory.LLMFactory == nil {
		logger.Warn("LLM factory is nil, Agent creation will be limited")
	}

	if factory.ToolRegistry == nil {
		logger.Warn("tool registry is nil, Agent will have no tools")
	}

	if factory.MemoryManager == nil {
		logger.Warn("memory manager is nil, Agent memory integration will be limited")
	}

	return factory
}

// CreateReActAgent 创建ReAct Agent
// 【必须】实现 AgentFactory 接口的 CreateReActAgent 方法
func (draf *DefaultReActAgentFactory) CreateReActAgent(ctx context.Context, cfg *ReActConfig, opts ...AgentOption) (ReActAgent, error) {
	if cfg == nil {
		return nil, errors.NewValidationError("config cannot be nil", nil)
	}

	// 验证配置
	if err := draf.ValidateConfig(cfg); err != nil {
		return nil, errors.WrapError(err, "invalid agent configuration", nil)
	}

	// 处理创建选项
	options := &agentCreationOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 【必须】记录创建开始
	draf.logger.Info("creating ReAct agent",
		zap.String("session_id", options.SessionID),
		zap.Int("max_iterations", cfg.MaxIterations),
		zap.Bool("detailed_thinking", cfg.EnableDetailedThinking),
		zap.String("log_level", cfg.LogLevel),
	)

	try:
	// 创建Agent基本信息
	agentID := options.SessionID
	if agentID == "" {
		agentID = uuid.New().String()
	}

	// 创建日志记录器
	agentLogger := options.Logger
	if agentLogger == nil {
		agentLogger = draf.logger.With(
			zap.String("agent_id", agentID),
			zap.String("agent_type", "react"),
		)
	}

	// 设置日志级别
	if cfg.LogLevel != "" {
		// 【必须】这里应该设置zap日志级别
		// 简化实现，实际项目中需要更完整的日志级别控制
	}

	// 创建模型客户端
	modelClient, err := draf.createModelClient(ctx, cfg, options)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create model client", nil)
	}

	// 创建规划器
	planner, err := draf.createPlanner(ctx, cfg, modelClient, agentLogger)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create planner", nil)
	}

	// 如果提供了自定义规划器，使用自定义的
	if options.Planner != nil {
		planner = options.Planner
	}

	// 创建执行器管理器
	executorManager, err := draf.createExecutorManager(ctx, cfg, agentLogger)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create executor manager", nil)
	}

	// 如果提供了自定义执行器，添加到管理器
	if options.Executor != nil {
		if err := executorManager.AddExecutor(options.Executor); err != nil {
			draf.logger.Warn("failed to add custom executor", zap.Error(err))
		}
	}

	// 创建内存集成器
	memoryIntegration, err := draf.createMemoryIntegration(ctx, cfg, agentLogger)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create memory integration", nil)
	}

	// 如果提供了自定义内存集成器，使用自定义的
	if options.Memory != nil {
		memoryIntegration = options.Memory
	}

	// 创建事件发射器
	eventEmitter := agent.NewEventEmitter(agentLogger)

	// 创建ReAct Agent实例
	agentCfg := &ReActConfig{
		MaxIterations:          cfg.MaxIterations,
		MaxPlanningTime:        cfg.MaxPlanningTime,
		MaxExecutionTime:       cfg.MaxExecutionTime,
		MinConfidenceThreshold: cfg.MinConfidenceThreshold,
		EnableDetailedThinking: cfg.EnableDetailedThinking,
		EnableParallelExecution: cfg.EnableParallelExecution,
		EnableMetrics:          cfg.EnableMetrics || options.MetricsEnabled,
		EnableTracing:          cfg.EnableTracing || options.TracingEnabled,
		LogLevel:               cfg.LogLevel,
		Tracker:                cfg.Tracker,
		ThoughtConfig:          cfg.ThoughtConfig,
		PlannerConfig:          cfg.PlannerConfig,
		ExecutorConfig:         cfg.ExecutorConfig,
		MemoryConfig:           cfg.MemoryConfig,
	}

	agent := &DefaultReActAgent{
		BaseReActAgent: BaseReActAgent{
			id:            agentID,
			config:        agentCfg,
			logger:        agentLogger,
			eventEmitter:  eventEmitter,
			stateManager:  NewReActStateManager(agentLogger, agentCfg),
			metrics:       NewReActMetrics(),
			creationTime:  time.Now().UTC(),
			capabilities:  []Capability{CapabilityReasoning, CapabilityToolUse, CapabilityPlanning, CapabilityLearning},
		},
		llmClient:        modelClient,
		planner:          planner,
		executorManager:  executorManager,
		memoryIntegration: memoryIntegration,
	}

	// 验证Agent配置
	if err := agent.Validate(); err != nil {
		return nil, errors.WrapError(err, "agent validation failed", nil)
	}

	// 注册Agent实例
	draf.agentsMutex.Lock()
	draf.agents[agentID] = agent
	draf.agentsMutex.Unlock()

	// 启动指标收集（如果启用）
	if agentCfg.EnableMetrics {
		go agent.startMetricsCollection(ctx)
	}

	// 【必须】记录创建成功
	draf.logger.Info("ReAct agent created successfully",
		zap.String("agent_id", agentID),
		zap.String("session_id", agent.GetSessionID()),
		zap.Int("capabilities_count", len(agent.capabilities)),
		zap.Bool("metrics_enabled", agentCfg.EnableMetrics),
		zap.Bool("tracing_enabled", agentCfg.EnableTracing),
	)

	return agent, nil

catch:
	// 处理创建过程中的异常
	draf.logger.Error("failed to create ReAct agent", zap.Error(err))
	return nil, errors.WrapError(err, "agent creation failed", nil)
}

// CreateReActAgentWithProfile 基于配置文件创建Agent
// 【必须】实现 AgentFactory 接口的 CreateReActAgentWithProfile 方法
func (draf *DefaultReActAgentFactory) CreateReActAgentWithProfile(ctx context.Context, profileName string, customOpts ...AgentOption) (ReActAgent, error) {
	if profileName == "" {
		return nil, errors.NewValidationError("profile name cannot be empty", nil)
	}

	// 获取配置模板
	baseConfig, err := draf.GetProfile(profileName)
	if err != nil {
		return nil, errors.WrapError(err, "failed to get profile", map[string]interface{}{
			"profile_name": profileName,
		})
	}

	// 合并自定义选项
	finalOpts := make([]AgentOption, 0, len(customOpts))
	finalOpts = append(finalOpts, customOpts...)

	// 创建Agent
	agent, err := draf.CreateReActAgent(ctx, baseConfig, finalOpts...)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create agent from profile", map[string]interface{}{
			"profile_name": profileName,
		})
	}

	// 【必须】记录基于配置文件的创建成功
	draf.logger.Info("created ReAct agent from profile",
		zap.String("profile_name", profileName),
		zap.String("agent_id", agent.ID()),
	)

	return agent, nil
}

// createModelClient 创建模型客户端
// 【推荐】根据配置创建合适的LLM客户端
func (draf *DefaultReActAgentFactory) createModelClient(ctx context.Context, cfg *ReActConfig, options *agentCreationOptions) (llm.Client, error) {
	// 如果提供了自定义客户端，使用自定义的
	if options.ModelClient != nil {
		return options.ModelClient, nil
	}

	if draf.LLMFactory == nil {
		return nil, errors.NewValidationError("LLM factory not available", nil)
	}

	// 根据配置选择合适的模型
	modelName := cfg.ModelName
	if modelName == "" {
		modelName = "gpt-4" // 默认模型
	}

	// 创建模型配置
	modelConfig := &llm.Config{
		Model:       modelName,
		APIKey:      cfg.ModelAPIKey,
		BaseURL:     cfg.ModelBaseURL,
		MaxTokens:   cfg.ModelMaxTokens,
		Temperature: cfg.ModelTemperature,
		Timeout:     cfg.ModelTimeout,
	}

	// 创建模型客户端
	client, err := draf.LLMFactory.CreateClient(ctx, modelConfig)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create LLM client", map[string]interface{}{
			"model_name": modelName,
		})
	}

	return client, nil
}

// createPlanner 创建规划器
// 【推荐】根据配置创建合适的规划器
func (draf *DefaultReActAgentFactory) createPlanner(ctx context.Context, cfg *ReActConfig, modelClient llm.Client, logger *zap.Logger) (PlanGenerator, error) {
	// 创建规划器配置
	plannerConfig := &PlanGeneratorConfig{
		MaxRetries:      cfg.PlannerConfig.MaxRetries,
		RetryDelay:      cfg.PlannerConfig.RetryDelay,
		MaxPlanningTime: cfg.MaxPlanningTime,
		EnableFallback:  cfg.PlannerConfig.EnableFallback,
		PromptTemplate:  cfg.PlannerConfig.PromptTemplate,
		ModelClient:     modelClient,
		Logger:          logger,
	}

	// 创建LLM规划生成器
	planner := NewLLMPlanGenerator(plannerConfig)

	// 验证规划器
	if err := planner.Validate(); err != nil {
		return nil, errors.WrapError(err, "planner validation failed", nil)
	}

	return planner, nil
}

// createExecutorManager 创建执行器管理器
// 【推荐】创建包含各种执行器的管理器
func (draf *DefaultReActAgentFactory) createExecutorManager(ctx context.Context, cfg *ReActConfig, logger *zap.Logger) (*ActionExecutorManager, error) {
	executorManager := NewActionExecutorManager(logger)

	// 创建工具执行器
	if draf.ToolRegistry != nil {
		toolExecutor := NewToolActionExecutor(logger, &cfg.ExecutorConfig, draf.ToolRegistry)
		if err := executorManager.AddExecutor(toolExecutor); err != nil {
			logger.Warn("failed to add tool executor", zap.Error(err))
		}
	}

	// 创建搜索执行器（如果有搜索提供者）
	// searchExecutor := NewSearchActionExecutor(logger, &cfg.ExecutorConfig, searchProviders)
	// if err := executorManager.AddExecutor(searchExecutor); err != nil {
	//     logger.Warn("failed to add search executor", zap.Error(err))
	// }

	// 创建反思执行器
	reflectExecutor := NewReflectActionExecutor(logger, cfg)
	if err := executorManager.AddExecutor(reflectExecutor); err != nil {
		logger.Warn("failed to add reflect executor", zap.Error(err))
	}

	// 验证执行器管理器
	if err := executorManager.Validate(); err != nil {
		return nil, errors.WrapError(err, "executor manager validation failed", nil)
	}

	return executorManager, nil
}

// createMemoryIntegration 创建内存集成器
// 【推荐】根据配置创建合适的内存集成器
func (draf *DefaultReActAgentFactory) createMemoryIntegration(ctx context.Context, cfg *ReActConfig, logger *zap.Logger) (MemoryIntegration, error) {
	// 创建内存集成器
	memoryIntegration := NewDefaultMemoryIntegration(logger, cfg, draf.MemoryManager)

	// 验证内存集成器
	if err := memoryIntegration.Validate(); err != nil {
		return nil, errors.WrapError(err, "memory integration validation failed", nil)
	}

	return memoryIntegration, nil
}

// GetAgent 根据ID获取Agent实例
// 【推荐】Agent实例管理方法
func (draf *DefaultReActAgentFactory) GetAgent(agentID string) ReActAgent {
	draf.agentsMutex.RLock()
	defer draf.agentsMutex.RUnlock()

	return draf.agents[agentID]
}

// RemoveAgent 移除Agent实例
// 【推荐】Agent实例管理方法
func (draf *DefaultReActAgentFactory) RemoveAgent(agentID string) {
	draf.agentsMutex.Lock()
	defer draf.agentsMutex.Unlock()

	delete(draf.agents, agentID)
	draf.logger.Debug("removed agent instance", zap.String("agent_id", agentID))
}

// AgentCount 返回管理的Agent数量
// 【推荐】工厂状态查询方法
func (draf *DefaultReActAgentFactory) AgentCount() int {
	draf.agentsMutex.RLock()
	defer draf.agentsMutex.RUnlock()

	return len(draf.agents)
}

// Close 关闭工厂并清理资源
// 【必须】优雅关闭工厂
func (draf *DefaultReActAgentFactory) Close() error {
	draf.logger.Info("closing ReAct agent factory")

	// 停止所有Agent
	draf.agentsMutex.Lock()
	for agentID, agent := range draf.agents {
		if closer, ok := agent.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				draf.logger.Warn("failed to close agent",
					zap.Error(err),
					zap.String("agent_id", agentID),
				)
			}
		}
	}
	// 清空Agent映射
	draf.agents = make(map[string]ReActAgent)
	draf.agentsMutex.Unlock()

	draf.logger.Info("ReAct agent factory closed")
	return nil
}

// GlobalReActAgentFactory 全局ReAct Agent工厂
// 【推荐】提供全局访问点，保持向后兼容
type GlobalReActAgentFactory struct {
	// factory 实际的工厂实例
	factory AgentFactory
	// factoryMutex 工厂访问互斥锁
	factoryMutex sync.RWMutex
}

// globalReActAgentFactory 全局工厂实例
var globalReActAgentFactory *GlobalReActAgentFactory

// init 初始化全局工厂
func init() {
	// 注意：在实际使用中，应该通过依赖注入来设置这些依赖
	// 这里只是提供一个基础的初始化
	logger := zap.NewNop()
	globalReActAgentFactory = &GlobalReActAgentFactory{
		factory: NewDefaultReActAgentFactory(logger, nil, nil, nil),
	}
}

// SetGlobalFactory 设置全局工厂
// 【推荐】允许外部设置全局工厂实例
func SetGlobalFactory(factory AgentFactory) {
	if factory == nil {
		panic("factory cannot be nil")
	}

	globalReActAgentFactory.factoryMutex.Lock()
	globalReActAgentFactory.factory = factory
	globalReActAgentFactory.factoryMutex.Unlock()
}

// GetGlobalFactory 获取全局工厂
// 【推荐】获取全局工厂实例
func GetGlobalFactory() AgentFactory {
	globalReActAgentFactory.factoryMutex.RLock()
	defer globalReActAgentFactory.factoryMutex.RUnlock()

	return globalReActAgentFactory.factory
}

// CreateReActAgent 全局便捷函数
// 【推荐】提供便捷的全局函数创建Agent
func CreateReActAgent(ctx context.Context, cfg *ReActConfig, opts ...AgentOption) (ReActAgent, error) {
	factory := GetGlobalFactory()
	if factory == nil {
		return nil, errors.NewInternalError("global factory not initialized", nil)
	}

	return factory.CreateReActAgent(ctx, cfg, opts...)
}

// CreateReActAgentWithProfile 全局便捷函数
// 【推荐】提供便捷的全局函数基于配置文件创建Agent
func CreateReActAgentWithProfile(ctx context.Context, profileName string, opts ...AgentOption) (ReActAgent, error) {
	factory := GetGlobalFactory()
	if factory == nil {
		return nil, errors.NewInternalError("global factory not initialized", nil)
	}

	return factory.CreateReActAgentWithProfile(ctx, profileName, opts...)
}

// RegisterProfile 全局便捷函数
// 【推荐】提供便捷的全局函数注册配置模板
func RegisterProfile(profileName string, config *ReActConfig) error {
	factory := GetGlobalFactory()
	if factory == nil {
		return errors.NewInternalError("global factory not initialized", nil)
	}

	return factory.RegisterProfile(profileName, config)
}

// ListProfiles 全局便捷函数
// 【推荐】提供便捷的全局函数列出配置模板
func ListProfiles() []string {
	factory := GetGlobalFactory()
	if factory == nil {
		return []string{}
	}

	return factory.ListProfiles()
}

// ValidateConfig 全局便捷函数
// 【推荐】提供便捷的全局函数验证配置
func ValidateConfig(cfg *ReActConfig) error {
	factory := GetGlobalFactory()
	if factory == nil {
		return errors.NewInternalError("global factory not initialized", nil)
	}

	return factory.ValidateConfig(cfg)
}

// QuickStartExample 快速开始示例
// 【推荐】提供使用示例和帮助函数
func QuickStartExample() {
	// 示例：创建一个研究型Agent
	/*
	ctx := context.Background()
	
	// 方式1：使用配置文件
	agent, err := CreateReActAgentWithProfile(ctx, "research")
	if err != nil {
		log.Fatal(err)
	}
	
	// 方式2：自定义配置
	cfg := NewReActConfig()
	cfg.MaxIterations = 15
	cfg.EnableDetailedThinking = true
	cfg.EnableParallelExecution = true
	
	agent, err := CreateReActAgent(ctx, cfg, WithSessionID("my-session"))
	if err != nil {
		log.Fatal(err)
	}
	
	// 使用Agent...
	state := agent.Start("Analyze the latest AI trends")
	// ... 交互逻辑
	*/
	
	fmt.Println("ReAct Agent Framework Quick Start Example")
	fmt.Println("Available profiles:", ListProfiles())
}