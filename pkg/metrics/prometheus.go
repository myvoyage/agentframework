// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Prometheus metrics package
// SPDX-License-Identifier: AGPL-3.0-or-later

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all Prometheus metrics for the Agent Framework
type Metrics struct {
	// Agent metrics
	AgentTotalAgents      prometheus.Gauge
	AgentActiveAgents     prometheus.Gauge
	AgentCreationTime     prometheus.Summary
	AgentRunTime         prometheus.Summary
	AgentErrors           prometheus.Counter

	// Model metrics
	ModelTotalModels      prometheus.Gauge
	ModelActiveModels     prometheus.Gauge
	ModelSwitchTime      prometheus.Summary
	ModelLatency         prometheus.Summary
	ModelErrors          *prometheus.CounterVec
	ModelCacheHits       prometheus.Counter
	ModelCacheMisses     prometheus.Counter

	// Skill metrics
	SkillCalls           *prometheus.CounterVec
	SkillLatency         *prometheus.HistogramVec
	SkillErrors          *prometheus.CounterVec
	SkillCacheHits      prometheus.Counter
	SkillCacheMisses     prometheus.Counter

	// Workflow metrics
	WorkflowRuns         *prometheus.CounterVec
	WorkflowLatency      *prometheus.HistogramVec
	WorkflowErrors       *prometheus.CounterVec
	WorkflowActiveTasks  prometheus.Gauge
	TaskExecutionTime    prometheus.Summary

	// System metrics
	SystemMemoryUsage    prometheus.Gauge
	SystemCPUUsage       prometheus.Gauge
	SystemGoroutines     prometheus.Gauge
	SystemGCCount        prometheus.Gauge
	SystemGCPauseTime     prometheus.Histogram
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		// Agent metrics
		AgentTotalAgents: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_agents_total",
			Help: "Total number of agents",
		}),
		AgentActiveAgents: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_agents_active",
			Help: "Number of active agents",
		}),
		AgentCreationTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "agentframework_agent_creation_duration_seconds",
			Help: "Time spent creating agents",
		}),
		AgentRunTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "agentframework_agent_run_duration_seconds",
			Help: "Time spent running agents",
		}),
		AgentErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "agentframework_agent_errors_total",
			Help: "Total number of agent errors",
		}),

		// Model metrics
		ModelTotalModels: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_models_total",
			Help: "Total number of models",
		}),
		ModelActiveModels: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_models_active",
			Help: "Number of active models",
		}),
		ModelSwitchTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "agentframework_model_switch_duration_seconds",
			Help: "Time spent switching models",
		}),
		ModelLatency: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "agentframework_model_latency_seconds",
			Help: "Model response latency",
		}),
		ModelErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agentframework_model_errors_total",
				Help: "Total number of model errors",
			},
			[]string{"model", "error_type"},
		),
		ModelCacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "agentframework_model_cache_hits_total",
			Help: "Total number of model cache hits",
		}),
		ModelCacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "agentframework_model_cache_misses_total",
			Help: "Total number of model cache misses",
		}),

		// Skill metrics
		SkillCalls: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agentframework_skill_calls_total",
				Help: "Total number of skill calls",
			},
			[]string{"skill", "status"},
		),
		SkillLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agentframework_skill_latency_seconds",
				Help:    "Skill execution latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"skill"},
		),
		SkillErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agentframework_skill_errors_total",
				Help: "Total number of skill errors",
			},
			[]string{"skill", "error_type"},
		),
		SkillCacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "agentframework_skill_cache_hits_total",
			Help: "Total number of skill cache hits",
		}),
		SkillCacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "agentframework_skill_cache_misses_total",
			Help: "Total number of skill cache misses",
		}),

		// Workflow metrics
		WorkflowRuns: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agentframework_workflow_runs_total",
				Help: "Total number of workflow runs",
			},
			[]string{"workflow", "status"},
		),
		WorkflowLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agentframework_workflow_latency_seconds",
				Help:    "Workflow execution latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"workflow"},
		),
		WorkflowErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agentframework_workflow_errors_total",
				Help: "Total number of workflow errors",
			},
			[]string{"workflow", "error_type"},
		),
		WorkflowActiveTasks: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_workflow_active_tasks",
			Help: "Number of active workflow tasks",
		}),
		TaskExecutionTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "agentframework_task_execution_duration_seconds",
			Help: "Time spent executing tasks",
		}),

		// System metrics
		SystemMemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_system_memory_bytes",
			Help: "System memory usage in bytes",
		}),
		SystemCPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_system_cpu_usage",
			Help: "System CPU usage percentage",
		}),
		SystemGoroutines: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_system_goroutines",
			Help: "Number of running goroutines",
		}),
		SystemGCCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "agentframework_system_gc_count",
			Help: "Number of garbage collections",
		}),
		SystemGCPauseTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "agentframework_system_gc_pause_seconds",
			Help:    "Garbage collection pause time",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}),
	}
}

// Agent metric methods
func (m *Metrics) IncAgentTotal() {
	m.AgentTotalAgents.Inc()
}

func (m *Metrics) SetAgentActive(count float64) {
	m.AgentActiveAgents.Set(count)
}

func (m *Metrics) ObserveAgentCreation(duration float64) {
	m.AgentCreationTime.Observe(duration)
}

func (m *Metrics) ObserveAgentRun(duration float64) {
	m.AgentRunTime.Observe(duration)
}

func (m *Metrics) IncAgentErrors() {
	m.AgentErrors.Inc()
}

// Model metric methods
func (m *Metrics) IncModelTotal() {
	m.ModelTotalModels.Inc()
}

func (m *Metrics) SetModelActive(count float64) {
	m.ModelActiveModels.Set(count)
}

func (m *Metrics) ObserveModelSwitch(duration float64) {
	m.ModelSwitchTime.Observe(duration)
}

func (m *Metrics) ObserveModelLatency(duration float64) {
	m.ModelLatency.Observe(duration)
}

func (m *Metrics) IncModelErrors(model, errorType string) {
	m.ModelErrors.WithLabelValues(model, errorType).Inc()
}

func (m *Metrics) IncModelCacheHits() {
	m.ModelCacheHits.Inc()
}

func (m *Metrics) IncModelCacheMisses() {
	m.ModelCacheMisses.Inc()
}

// Skill metric methods
func (m *Metrics) IncSkillCalls(skill, status string) {
	m.SkillCalls.WithLabelValues(skill, status).Inc()
}

func (m *Metrics) ObserveSkillLatency(skill string, duration float64) {
	m.SkillLatency.WithLabelValues(skill).Observe(duration)
}

func (m *Metrics) IncSkillErrors(skill, errorType string) {
	m.SkillErrors.WithLabelValues(skill, errorType).Inc()
}

func (m *Metrics) IncSkillCacheHits() {
	m.SkillCacheHits.Inc()
}

func (m *Metrics) IncSkillCacheMisses() {
	m.SkillCacheMisses.Inc()
}

// Workflow metric methods
func (m *Metrics) IncWorkflowRuns(workflow, status string) {
	m.WorkflowRuns.WithLabelValues(workflow, status).Inc()
}

func (m *Metrics) ObserveWorkflowLatency(workflow string, duration float64) {
	m.WorkflowLatency.WithLabelValues(workflow).Observe(duration)
}

func (m *Metrics) IncWorkflowErrors(workflow, errorType string) {
	m.WorkflowErrors.WithLabelValues(workflow, errorType).Inc()
}

func (m *Metrics) SetWorkflowActiveTasks(count float64) {
	m.WorkflowActiveTasks.Set(count)
}

func (m *Metrics) ObserveTaskExecution(duration float64) {
	m.TaskExecutionTime.Observe(duration)
}

// System metric methods
func (m *Metrics) SetSystemMemoryUsage(bytes float64) {
	m.SystemMemoryUsage.Set(bytes)
}

func (m *Metrics) SetSystemCPUUsage(percentage float64) {
	m.SystemCPUUsage.Set(percentage)
}

func (m *Metrics) SetSystemGoroutines(count float64) {
	m.SystemGoroutines.Set(count)
}

func (m *Metrics) SetSystemGCCount(count float64) {
	m.SystemGCCount.Set(count)
}

func (m *Metrics) ObserveGCPause(duration float64) {
	m.SystemGCPauseTime.Observe(duration)
}