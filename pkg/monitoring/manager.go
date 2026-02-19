// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
//
// Monitoring Integration Manager
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"AgentFramework/pkg/metrics"
)

// MonitoringManager 监控管理器
type MonitoringManager struct {
	registry   *prometheus.Registry
	metrics     *metrics.Metrics
	mu          sync.RWMutex
	agentID     string
	version     string
	environment string

	// Custom collectors
	collectors []prometheus.Collector
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	AgentID     string
	Version      string
	Environment  string // "production", "staging", "development"
	Enabled      bool
	Port         int // Metrics server port
}

// NewMonitoringManager 创建监控管理器
func NewMonitoringManager(config *MonitoringConfig) (*MonitoringManager, error) {
	if config == nil {
		config = &MonitoringConfig{
			AgentID:     "agent-framework",
			Version:      "1.0.0",
			Environment:  "development",
			Enabled:      true,
			Port:         9090,
		}
	}

	registry := prometheus.NewRegistry()
	metrics := &metrics.Metrics{}

	// Register agent metrics
	metrics.AgentTotalAgents = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "agent_total_agents",
			Help: "Total number of agents",
		},
	)
	metrics.AgentActiveAgents = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "agent_active_agents",
			Help: "Number of active agents",
		},
	)
	metrics.AgentCreationTime = promauto.With(registry).NewSummary(
		prometheus.SummaryOpts{
			Name: "agent_creation_time_seconds",
			Help: "Time taken to create agents",
		},
	)
	metrics.AgentRunTime = promauto.With(registry).NewSummary(
		prometheus.SummaryOpts{
			Name: "agent_run_time_seconds",
			Help: "Time taken by agents to run",
		},
	)
	metrics.AgentErrors = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "agent_errors_total",
			Help: "Total number of agent errors",
		},
	)

	// Register model metrics
	metrics.ModelTotalModels = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "model_total_models",
			Help: "Total number of models",
		},
	)
	metrics.ModelActiveModels = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "model_active_models",
			Help: "Number of active models",
		},
	)
	metrics.ModelSwitchTime = promauto.With(registry).NewSummary(
		prometheus.SummaryOpts{
			Name: "model_switch_time_seconds",
			Help: "Time taken to switch models",
		},
	)
	metrics.ModelLatency = promauto.With(registry).NewSummary(
		prometheus.SummaryOpts{
			Name: "model_latency_seconds",
			Help: "Model request latency",
		},
	)
	metrics.ModelErrors = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "model_errors_total",
			Help: "Total number of model errors",
		},
		[]string{"model_type", "error_type"},
	)
	metrics.ModelCacheHits = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "model_cache_hits_total",
			Help: "Total number of model cache hits",
		},
	)
	metrics.ModelCacheMisses = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "model_cache_misses_total",
			Help: "Total number of model cache misses",
		},
	)

	// Register skill metrics
	metrics.SkillCalls = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "skill_calls_total",
			Help: "Total number of skill calls",
		},
		[]string{"skill_id", "category"},
	)
	metrics.SkillLatency = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "skill_latency_seconds",
			Help: "Skill execution latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"skill_id", "category"},
	)
	metrics.SkillErrors = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "skill_errors_total",
			Help: "Total number of skill errors",
		},
		[]string{"skill_id", "category", "error_type"},
	)
	metrics.SkillCacheHits = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "skill_cache_hits_total",
			Help: "Total number of skill cache hits",
		},
	)
	metrics.SkillCacheMisses = promauto.With(registry).NewCounter(
		prometheus.CounterOpts{
			Name: "skill_cache_misses_total",
			Help: "Total number of skill cache misses",
		},
	)

	// Register workflow metrics
	metrics.WorkflowRuns = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_runs_total",
			Help: "Total number of workflow runs",
		},
		[]string{"workflow_id", "status"},
	)
	metrics.WorkflowLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "workflow_latency_seconds",
			Help:    "Workflow execution latency",
			Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 300, 600, 1800, 3600},
		},
		[]string{"workflow_id", "step_id"},
	)
	metrics.WorkflowErrors = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_errors_total",
			Help: "Total number of workflow errors",
		},
		[]string{"workflow_id", "step_id", "error_type"},
	)
	metrics.WorkflowActiveTasks = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "workflow_active_tasks",
			Help: "Number of active workflow tasks",
		},
	)
	metrics.TaskExecutionTime = promauto.With(registry).NewSummary(
		prometheus.SummaryOpts{
			Name: "task_execution_time_seconds",
			Help: "Task execution time",
		},
	)

	// Register system metrics
	metrics.SystemMemoryUsage = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "system_memory_usage_bytes",
			Help: "System memory usage in bytes",
		},
	)
	metrics.SystemCPUUsage = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "System CPU usage percentage",
		},
	)
	metrics.SystemGoroutines = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "system_gouroutines",
			Help: "Number of running goroutines",
		},
	)
	metrics.SystemGCCount = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "system_gc_count",
			Help: "Number of garbage collections",
		},
	)
	metrics.SystemGCPauseTime = promauto.With(registry).NewHistogram(
		prometheus.HistogramOpts{
			Name: "system_gc_pause_seconds",
			Help: "Garbage collection pause time",
			Buckets: []float64{0.000001, 0.00001, 0.0001, 0.001, 0.01, 0.1, 1},
		},
	)

	return &MonitoringManager{
		registry:   registry,
		metrics:     metrics,
		agentID:     config.AgentID,
		version:     config.Version,
		environment: config.Environment,
		collectors: []prometheus.Collector{},
	}, nil
}

// AgentMetrics 代理指标追踪器
type AgentMetrics struct {
	manager *MonitoringManager
	agentID string
}

// NewAgentMetrics 创建代理指标追踪器
func (m *MonitoringManager) NewAgentMetrics(agentID string) *AgentMetrics {
	return &AgentMetrics{
		manager: m,
		agentID: agentID,
	}
}

// RecordAgentCreation 记录代理创建
func (a *AgentMetrics) RecordAgentCreation(duration time.Duration) {
	a.manager.metrics.AgentTotalAgents.Inc()
	a.manager.metrics.AgentActiveAgents.Inc()
	a.manager.metrics.AgentCreationTime.Observe(duration.Seconds())
}

// RecordAgentStart 记录代理启动
func (a *AgentMetrics) RecordAgentStart() {
	a.manager.metrics.AgentActiveAgents.Inc()
}

// RecordAgentStop 记录代理停止
func (a *AgentMetrics) RecordAgentStop() {
	a.manager.metrics.AgentActiveAgents.Dec()
}

// RecordAgentRun 记录代理运行
func (a *AgentMetrics) RecordAgentRun(duration time.Duration, err error) {
	a.manager.metrics.AgentRunTime.Observe(duration.Seconds())
	if err != nil {
		a.manager.metrics.AgentErrors.Inc()
	}
}

// ModelMetrics 模型指标追踪器
type ModelMetrics struct {
	manager *MonitoringManager
	modelID string
}

// NewModelMetrics 创建模型指标追踪器
func (m *MonitoringManager) NewModelMetrics(modelID string) *ModelMetrics {
	m.metrics.ModelTotalModels.Inc()
	return &ModelMetrics{
		manager: m,
		modelID: modelID,
	}
}

// RecordModelActivation 记录模型激活
func (m *ModelMetrics) RecordModelActivation() {
	m.manager.metrics.ModelActiveModels.Inc()
}

// RecordModelDeactivation 记录模型失活
func (m *ModelMetrics) RecordModelDeactivation() {
	m.manager.metrics.ModelActiveModels.Dec()
}

// RecordModelSwitch 记录模型切换
func (m *ModelMetrics) RecordModelSwitch(duration time.Duration) {
	m.manager.metrics.ModelSwitchTime.Observe(duration.Seconds())
}

// RecordModelRequest 记录模型请求
func (m *ModelMetrics) RecordModelRequest(duration time.Duration, err error) {
	m.manager.metrics.ModelLatency.Observe(duration.Seconds())
	if err != nil {
		m.manager.metrics.ModelErrors.WithLabelValues(m.modelID, err.Error()).Inc()
	}
}

// RecordCacheHit 记录缓存命中
func (m *ModelMetrics) RecordCacheHit() {
	m.manager.metrics.ModelCacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func (m *ModelMetrics) RecordCacheMiss() {
	m.manager.metrics.ModelCacheMisses.Inc()
}

// SkillMetrics 技能指标追踪器
type SkillMetrics struct {
	manager *MonitoringManager
	skillID string
	category string
}

// NewSkillMetrics 创建技能指标追踪器
func (m *MonitoringManager) NewSkillMetrics(skillID, category string) *SkillMetrics {
	return &SkillMetrics{
		manager: m,
		skillID: skillID,
		category: category,
	}
}

// RecordSkillCall 记录技能调用
func (s *SkillMetrics) RecordSkillCall() {
	s.manager.metrics.SkillCalls.WithLabelValues(s.skillID, s.category).Inc()
}

// RecordSkillExecution 记录技能执行
func (s *SkillMetrics) RecordSkillExecution(duration time.Duration, err error) {
	s.manager.metrics.SkillLatency.WithLabelValues(s.skillID, s.category).Observe(duration.Seconds())
	if err != nil {
		s.manager.metrics.SkillErrors.WithLabelValues(s.skillID, s.category, err.Error()).Inc()
	}
}

// RecordCacheHit 记录缓存命中
func (s *SkillMetrics) RecordCacheHit() {
	s.manager.metrics.SkillCacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func (s *SkillMetrics) RecordCacheMiss() {
	s.manager.metrics.SkillCacheMisses.Inc()
}

// WorkflowMetrics 工作流指标追踪器
type WorkflowMetrics struct {
	manager     *MonitoringManager
	workflowID string
}

// NewWorkflowMetrics 创建工作流指标追踪器
func (m *MonitoringManager) NewWorkflowMetrics(workflowID string) *WorkflowMetrics {
	return &WorkflowMetrics{
		manager:     m,
		workflowID: workflowID,
	}
}

// RecordWorkflowStart 记录工作流启动
func (w *WorkflowMetrics) RecordWorkflowStart() {
	w.manager.metrics.WorkflowRuns.WithLabelValues(w.workflowID, "started").Inc()
}

// RecordWorkflowComplete 记录工作流完成
func (w *WorkflowMetrics) RecordWorkflowComplete(duration time.Duration, err error) {
	status := "completed"
	if err != nil {
		status = "failed"
		w.manager.metrics.WorkflowErrors.WithLabelValues(w.workflowID, "complete", err.Error()).Inc()
	} else {
		w.manager.metrics.WorkflowRuns.WithLabelValues(w.workflowID, status).Inc()
	}
	w.manager.metrics.WorkflowLatency.WithLabelValues(w.workflowID, "complete").Observe(duration.Seconds())
}

// RecordTaskStart 记录任务启动
func (w *WorkflowMetrics) RecordTaskStart() {
	w.manager.metrics.WorkflowActiveTasks.Inc()
}

// RecordTaskComplete 记录任务完成
func (w *WorkflowMetrics) RecordTaskComplete(duration time.Duration) {
	w.manager.metrics.WorkflowActiveTasks.Dec()
	w.manager.metrics.TaskExecutionTime.Observe(duration.Seconds())
}

// SystemMetrics 系统指标追踪器
type SystemMetrics struct {
	manager       *MonitoringManager
	updateInterval time.Duration
	stopChan      chan bool
	wg            sync.WaitGroup
}

// NewSystemMetrics 创建系统指标追踪器
func (m *MonitoringManager) NewSystemMetrics(updateInterval time.Duration) *SystemMetrics {
	if updateInterval == 0 {
		updateInterval = 15 * time.Second // 默认15秒更新一次
	}

	return &SystemMetrics{
		manager:       m,
		updateInterval: updateInterval,
		stopChan:      make(chan bool),
	}
}

// Start 开始收集系统指标
func (s *SystemMetrics) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.collectLoop(ctx)
}

// Stop 停止收集系统指标
func (s *SystemMetrics) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

// collectLoop 循环收集系统指标
func (s *SystemMetrics) collectLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateMetrics()
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// updateMetrics 更新系统指标
func (s *SystemMetrics) updateMetrics() {
	var m runtime.MemStats

	// Read memory stats
	runtime.ReadMemStats(&m)

	// Memory usage
	s.manager.metrics.SystemMemoryUsage.Set(float64(m.Alloc))

	// Number of goroutines
	s.manager.metrics.SystemGoroutines.Set(float64(runtime.NumGoroutine()))

	// GC stats
	s.manager.metrics.SystemGCCount.Set(float64(m.NumGC))

	// GC pause time (sum of all GC pauses in nanoseconds, convert to seconds)
	var totalPause uint64
	for _, pause := range m.PauseNs {
		totalPause += pause
	}
	pauseTime := time.Duration(totalPause)
	s.manager.metrics.SystemGCPauseTime.Observe(pauseTime.Seconds())
}

// GetMetrics 获取指标数据
func (m *MonitoringManager) GetMetrics() *metrics.Metrics {
	return m.metrics
}

// GetRegistry 获取注册表
func (m *MonitoringManager) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RegisterCollector 注册自定义收集器
func (m *MonitoringManager) RegisterCollector(collector prometheus.Collector) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collectors = append(m.collectors, collector)
}

// GetAgentMetrics 获取代理指标追踪器
func (m *MonitoringManager) GetAgentMetrics(agentID string) *AgentMetrics {
	return m.NewAgentMetrics(agentID)
}

// GetModelMetrics 获取模型指标追踪器
func (m *MonitoringManager) GetModelMetrics(modelID string) *ModelMetrics {
	return m.NewModelMetrics(modelID)
}

// GetSkillMetrics 获取技能指标追踪器
func (m *MonitoringManager) GetSkillMetrics(skillID, category string) *SkillMetrics {
	return m.NewSkillMetrics(skillID, category)
}

// GetWorkflowMetrics 获取工作流指标追踪器
func (m *MonitoringManager) GetWorkflowMetrics(workflowID string) *WorkflowMetrics {
	return m.NewWorkflowMetrics(workflowID)
}

// GetSystemMetrics 获取系统指标追踪器
func (m *MonitoringManager) GetSystemMetrics(updateInterval time.Duration) *SystemMetrics {
	return m.NewSystemMetrics(updateInterval)
}

// mu 读写锁 - 注释：在实际实现中应该使用 sync.RWMutex
// 为简单起见，这里使用不带锁的 map，生产环境应添加适当的锁机制
