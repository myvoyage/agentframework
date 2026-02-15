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

package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Agent AI 代理接口（简化定义）
type Agent interface {
	Run(ctx context.Context, input string, opts ...interface{}) (string, error)
}

// Scheduler 调度器接口
type Scheduler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	ScheduleJob(ctx context.Context, job *Job) (string, error)
	ScheduleCron(ctx context.Context, cronExpr string, handler JobHandler) (string, error)
	ScheduleInterval(ctx context.Context, interval time.Duration, handler JobHandler) (string, error)
	ScheduleOnce(ctx context.Context, delay time.Duration, handler JobHandler) (string, error)
	UnscheduleJob(ctx context.Context, jobID string) error
	GetJob(ctx context.Context, jobID string) (*Job, error)
	ListJobs(ctx context.Context) ([]*Job, error)
	GetStats(ctx context.Context) (*SchedulerStats, error)
	PauseJob(ctx context.Context, jobID string) error
	ResumeJob(ctx context.Context, jobID string) error
	IsJobRunning(jobID string) bool
}

// Job 定时任务定义
type Job struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Handler     JobHandler
	Schedule    JobSchedule
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastRunAt   time.Time
	NextRunAt   time.Time
	RunCount    int64
	FailCount   int64
	Enabled     bool
	Status      JobStatus
}

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusPaused    JobStatus = "paused"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// JobSchedule 调度配置
type JobSchedule struct {
	Type       ScheduleType
	CronExpr   string
	Interval   time.Duration
	Delay      time.Duration
	StartTime  time.Time
	EndTime    time.Time
	Timezone   string
	MaxRuns    int
	Concurrent bool
}

// ScheduleType 调度类型
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOnce     ScheduleType = "once"
)

// JobHandler 任务处理器函数类型
type JobHandler func(ctx context.Context) error

// SchedulerStats 调度器统计信息
type SchedulerStats struct {
	TotalJobs     int64
	ActiveJobs    int64
	PausedJobs    int64
	CompletedJobs int64
	FailedJobs    int64
	TotalRuns     int64
	TotalFailures int64
	AvgDuration   time.Duration
	Uptime        time.Duration
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Enabled           bool
	Timezone          string
	MaxConcurrentJobs int
	JobTimeout        time.Duration
	Logger            Logger
	EnableRetry       bool
	MaxRetries        int
	RetryInterval     time.Duration
	EnableMetrics     bool
	EnableTracing     bool
	PersistJobs       bool
	StoragePath       string
}

// DefaultSchedulerConfig 返回默认配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		Enabled:           true,
		Timezone:          "Local",
		MaxConcurrentJobs: 10,
		JobTimeout:        30 * time.Minute,
		Logger:            &DefaultLogger{},
		EnableRetry:       true,
		MaxRetries:        3,
		RetryInterval:     1 * time.Minute,
		EnableMetrics:     true,
		EnableTracing:     false,
		PersistJobs:       false,
	}
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	l.log("[INFO]", msg, fields...)
}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	l.log("[ERROR]", msg, fields...)
}

func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	l.log("[DEBUG]", msg, fields...)
}

func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	l.log("[WARN]", msg, fields...)
}

func (l *DefaultLogger) log(prefix, msg string, fields ...interface{}) {
	result := prefix + " " + msg
	if len(fields) > 0 {
		result += fmt.Sprintf(" %v", fields)
	}
	fmt.Println(result)
}

// ==================== Task Scheduler Implementation ====================

// TaskScheduler 任务调度器实现
type TaskScheduler struct {
	config *SchedulerConfig
	jobs   map[string]*Job
	mu     sync.RWMutex
	cron   *CronScheduler
	agent  Agent // AI agent for intelligent tasks
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(cfg *SchedulerConfig) *TaskScheduler {
	if cfg == nil {
		cfg = DefaultSchedulerConfig()
	}

	return &TaskScheduler{
		config: cfg,
		jobs:   make(map[string]*Job),
		cron:   NewCronScheduler(cfg.Timezone),
	}
}

// SetAgent 设置 AI 代理
func (s *TaskScheduler) SetAgent(agent Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agent = agent
}

// Start 启动调度器
func (s *TaskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enabled {
		return fmt.Errorf("scheduler is disabled")
	}

	s.config.Logger.Info("Starting task scheduler")

	// Start cron scheduler
	if err := s.cron.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cron scheduler: %w", err)
	}

	return nil
}

// Stop 停止调度器
func (s *TaskScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Logger.Info("Stopping task scheduler")

	// Stop cron scheduler
	if err := s.cron.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop cron scheduler: %w", err)
	}

	return nil
}

// ScheduleTask 安排任务
func (s *TaskScheduler) ScheduleTask(ctx context.Context, task *Task) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate task ID if not provided
	if task.ID == "" {
		task.ID = generateTaskID()
	}

	// Create job from task
	job := &Job{
		ID:          task.ID,
		Name:        task.Name,
		Description: task.Description,
		Tags:        []string{"scheduled"},
		Handler:     s.createTaskHandler(task),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Enabled:     task.Enabled,
		Status:      JobStatusPending,
	}

	// Schedule based on task type
	switch task.Type {
	case TaskTypeStatic:
		return s.scheduleStaticTask(ctx, job, task)
	case TaskTypeAI:
		return s.scheduleAITask(ctx, job, task)
	default:
		return "", fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// scheduleStaticTask 安排静态消息任务
func (s *TaskScheduler) scheduleStaticTask(ctx context.Context, job *Job, task *Task) (string, error) {
	if task.StaticMessage == nil || *task.StaticMessage == "" {
		return "", fmt.Errorf("static message is required for static tasks")
	}

	// Store message in job metadata
	job.Metadata["message"] = *task.StaticMessage

	// Schedule with cron
	handler := s.createStaticMessageHandler(*task.StaticMessage)
	jobID, err := s.cron.Schedule(ctx, task.CronExpr, handler)
	if err != nil {
		return "", fmt.Errorf("failed to schedule static task: %w", err)
	}

	job.Schedule = JobSchedule{
		Type:     ScheduleTypeCron,
		CronExpr: task.CronExpr,
	}

	// Store job
	s.jobs[job.ID] = job

	return jobID, nil
}

// scheduleAITask 安排 AI 智能任务
func (s *TaskScheduler) scheduleAITask(ctx context.Context, job *Job, task *Task) (string, error) {
	if s.agent == nil {
		return "", fmt.Errorf("AI agent not set, cannot schedule AI tasks")
	}

	if task.AIPrompt == nil || *task.AIPrompt == "" {
		return "", fmt.Errorf("AI prompt is required for AI tasks")
	}

	// Store AI config in job metadata
	job.Metadata["prompt"] = *task.AIPrompt
	if task.AIModel != nil {
		job.Metadata["model"] = *task.AIModel
	}
	if len(task.AITools) > 0 {
		job.Metadata["tools"] = strings.Join(task.AITools, ",")
	}

	// Schedule with cron
	handler := s.createAIHandler(*task.AIPrompt)
	jobID, err := s.cron.Schedule(ctx, task.CronExpr, handler)
	if err != nil {
		return "", fmt.Errorf("failed to schedule AI task: %w", err)
	}

	job.Schedule = JobSchedule{
		Type:     ScheduleTypeCron,
		CronExpr: task.CronExpr,
	}

	// Store job
	s.jobs[job.ID] = job

	return jobID, nil
}

// UnscheduleJob 取消任务
func (s *TaskScheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Unschedule from cron
	if err := s.cron.Unschedule(ctx, jobID); err != nil {
		return fmt.Errorf("failed to unschedule job: %w", err)
	}

	// Mark as disabled
	job.Enabled = false
	job.Status = JobStatusPaused
	job.UpdatedAt = time.Now()

	return nil
}

// GetJob 获取任务
func (s *TaskScheduler) GetJob(ctx context.Context, jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs 列出所有任务
func (s *TaskScheduler) ListJobs(ctx context.Context) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetStats 获取统计信息
func (s *TaskScheduler) GetStats(ctx context.Context) (*SchedulerStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SchedulerStats{
		TotalJobs: int64(len(s.jobs)),
	}

	for _, job := range s.jobs {
		switch job.Status {
		case JobStatusRunning:
			stats.ActiveJobs++
		case JobStatusPaused:
			stats.PausedJobs++
		case JobStatusCompleted:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		}
		stats.TotalRuns += job.RunCount
		stats.TotalFailures += job.FailCount
	}

	return stats, nil
}

// createTaskHandler 创建任务处理器
func (s *TaskScheduler) createTaskHandler(task *Task) JobHandler {
	return func(ctx context.Context) error {
		s.config.Logger.Info("Executing task", "name", task.Name, "type", task.Type)

		startTime := time.Now()
		job, exists := s.jobs[task.ID]
		if !exists {
			return fmt.Errorf("job not found: %s", task.ID)
		}

		// Update job status
		job.Status = JobStatusRunning
		job.LastRunAt = startTime
		job.RunCount++

		// Execute task
		var err error
		switch task.Type {
		case TaskTypeStatic:
			err = s.executeStaticTask(ctx, task)
		case TaskTypeAI:
			err = s.executeAITask(ctx, task)
		}

		// Update job status
		duration := time.Since(startTime)
		if err != nil {
			job.Status = JobStatusFailed
			job.FailCount++
			s.config.Logger.Error("Task execution failed", "name", task.Name, "error", err, "duration", duration)
		} else {
			job.Status = JobStatusCompleted
			s.config.Logger.Info("Task execution completed", "name", task.Name, "duration", duration)
		}

		return err
	}
}

// createStaticMessageHandler 创建静态消息处理器
func (s *TaskScheduler) createStaticMessageHandler(message string) JobHandler {
	return func(ctx context.Context) error {
		s.config.Logger.Info("Static message", "message", message)
		// TODO: Send message to appropriate channel
		return nil
	}
}

// createAIHandler 创建 AI 处理器
func (s *TaskScheduler) createAIHandler(prompt string) JobHandler {
	return func(ctx context.Context) error {
		if s.agent == nil {
			return fmt.Errorf("AI agent not available")
		}

		s.config.Logger.Info("AI task prompt", "prompt", prompt)

		// Execute AI task
		response, err := s.agent.Run(ctx, prompt)
		if err != nil {
			return fmt.Errorf("AI execution failed: %w", err)
		}

		s.config.Logger.Info("AI task response", "response", response)
		// TODO: Send response to appropriate channel

		return nil
	}
}

// executeStaticTask 执行静态消息任务
func (s *TaskScheduler) executeStaticTask(ctx context.Context, task *Task) error {
	if task.StaticMessage == nil {
		return fmt.Errorf("static message is nil")
	}

	// TODO: Implement actual message sending
	return nil
}

// executeAITask 执行 AI 智能任务
func (s *TaskScheduler) executeAITask(ctx context.Context, task *Task) error {
	if s.agent == nil {
		return fmt.Errorf("AI agent not available")
	}

	if task.AIPrompt == nil {
		return fmt.Errorf("AI prompt is nil")
	}

	// Execute AI task
	_, err := s.agent.Run(ctx, *task.AIPrompt)
	return err
}

// generateTaskID 生成任务 ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
