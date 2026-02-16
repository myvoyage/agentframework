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

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskScheduler 任务调度器，负责任务的执行调度
type TaskScheduler struct {
	collection    *TaskCollection
	agents        map[string]Agent
	executor      TaskExecutor
	maxConcurrent int
	timeout       time.Duration
	mu            sync.RWMutex
	cancelChan    chan struct{}
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(ctx context.Context, task *SubTask, agent Agent) (string, error)
}

// DefaultTaskExecutor 默认任务执行器
type DefaultTaskExecutor struct{}

// Execute 执行任务
func (e *DefaultTaskExecutor) Execute(ctx context.Context, task *SubTask, agent Agent) (string, error) {
	input := task.Input
	if input == "" {
		input = task.Description
	}

	resp, err := agent.Run(ctx, input)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// TaskSchedulerConfig 任务调度器配置
type TaskSchedulerConfig struct {
	// Collection 任务集合
	Collection *TaskCollection
	// Agents 可用的Agent映射
	Agents map[string]Agent
	// Executor 任务执行器
	Executor TaskExecutor
	// MaxConcurrent 最大并发执行数
	MaxConcurrent int
	// Timeout 任务执行超时时间
	Timeout time.Duration
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(config TaskSchedulerConfig) (*TaskScheduler, error) {
	if config.Collection == nil {
		return nil, fmt.Errorf("collection is required")
	}
	if config.Agents == nil {
		return nil, fmt.Errorf("agents map is required")
	}

	executor := config.Executor
	if executor == nil {
		executor = &DefaultTaskExecutor{}
	}

	maxConcurrent := config.MaxConcurrent
	if maxConcurrent == 0 {
		maxConcurrent = 3
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	return &TaskScheduler{
		collection:    config.Collection,
		agents:        config.Agents,
		executor:      executor,
		maxConcurrent: maxConcurrent,
		timeout:       timeout,
		cancelChan:    make(chan struct{}),
	}, nil
}

// ScheduleOptions 调度选项
type ScheduleOptions struct {
	// ContinueOnError 遇到错误是否继续
	ContinueOnError bool
	// RetryFailedTasks 是否重试失败的任务
	RetryFailedTasks bool
	// ProgressCallback 进度回调
	ProgressCallback func(task *SubTask, total, completed int)
}

// Schedule 调度并执行所有任务
func (ts *TaskScheduler) Schedule(ctx context.Context, opts ScheduleOptions) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	totalTasks := ts.collection.Count()
	completed := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ts.cancelChan:
			return fmt.Errorf("task scheduling cancelled")
		default:
		}

		// 获取准备好的任务
		readyTasks := ts.collection.GetReadyTasks()
		if len(readyTasks) == 0 {
			// 检查是否所有任务都已完成
			counts := ts.collection.CountByStatus()
			if counts[SubTaskCompleted]+counts[SubTaskFailed]+counts[SubTaskSkipped] == totalTasks {
				return nil // 所有任务已完成
			}

			// 还有未完成的任务但没有准备好的，说明有死锁或循环依赖
			pending := ts.collection.ListTasksByStatus(SubTaskPending)
			if len(pending) > 0 {
				return fmt.Errorf("无法执行 %d 个任务，可能存在循环依赖或未满足的依赖条件", len(pending))
			}

			// 等待一段时间后重试
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 限制并发执行数
		runningCount := len(ts.collection.ListTasksByStatus(SubTaskRunning))
		availableSlots := ts.maxConcurrent - runningCount

		if availableSlots <= 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// 执行任务
		tasksToExecute := readyTasks
		if len(tasksToExecute) > availableSlots {
			tasksToExecute = tasksToExecute[:availableSlots]
		}

		for _, task := range tasksToExecute {
			go ts.executeTask(ctx, task, opts, &completed, totalTasks)
		}

		// 等待一段时间
		time.Sleep(50 * time.Millisecond)
	}
}

// executeTask 执行单个任务
func (ts *TaskScheduler) executeTask(ctx context.Context, task *SubTask, opts ScheduleOptions, completed *int, total int) {
	// 设置任务状态为运行中
	task.SetStatus(SubTaskRunning)

	// 设置超时
	execCtx := ctx
	if ts.timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, ts.timeout)
		defer cancel()
	}

	// 查找指定的Agent
	agent, ok := ts.agents[task.Assignee]
	if !ok {
		// 如果没有指定Agent，使用默认的Agent
		for _, a := range ts.agents {
			agent = a
			break
		}
		if agent == nil {
			task.SetStatus(SubTaskFailed)
			task.SetError(fmt.Errorf("no available agent"))
			return
		}
	}

	// 执行任务
	output, err := ts.executor.Execute(execCtx, task, agent)

	if err != nil {
		// 任务执行失败
		task.SetError(err)

		if opts.RetryFailedTasks && task.CanRetry() {
			task.IncrementRetry()
			task.SetStatus(SubTaskPending)
			// 将重新加入调度队列
			return
		}

		task.SetStatus(SubTaskFailed)

		if !opts.ContinueOnError {
			close(ts.cancelChan)
			return
		}
	} else {
		// 任务执行成功
		task.SetOutput(output)
		task.SetStatus(SubTaskCompleted)
	}

	// 更新完成计数
	*completed++

	// 调用进度回调
	if opts.ProgressCallback != nil {
		opts.ProgressCallback(task, total, *completed)
	}
}

// Cancel 取消任务调度
func (ts *TaskScheduler) Cancel() {
	close(ts.cancelChan)
}

// GetProgress 获取执行进度
func (ts *TaskScheduler) GetProgress() (total, completed, failed int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	counts := ts.collection.CountByStatus()
	total = ts.collection.Count()
	completed = counts[SubTaskCompleted]
	failed = counts[SubTaskFailed]
	return
}

// GetExecutionReport 获取执行报告
func (ts *TaskScheduler) GetExecutionReport() *ExecutionReport {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	counts := ts.collection.CountByStatus()
	tasks := ts.collection.ListTasks()

	report := &ExecutionReport{
		TotalTasks:     ts.collection.Count(),
		CompletedTasks: counts[SubTaskCompleted],
		FailedTasks:    counts[SubTaskFailed],
		PendingTasks:   counts[SubTaskPending],
		RunningTasks:   counts[SubTaskRunning],
		SkippedTasks:   counts[SubTaskSkipped],
		Tasks:          make([]TaskReportItem, 0, len(tasks)),
	}

	for _, task := range tasks {
		item := TaskReportItem{
			ID:           task.ID,
			Description:  task.Description,
			Status:       task.GetStatus().String(),
			Priority:     task.Priority.String(),
			Assignee:     task.Assignee,
			Error:        task.GetError(),
			Duration:     task.ActualDuration,
			Estimated:    task.EstimatedDuration,
			RetryCount:   task.RetryCount,
		}
		report.Tasks = append(report.Tasks, item)
	}

	return report
}

// ExecutionReport 执行报告
type ExecutionReport struct {
	TotalTasks     int                `json:"total_tasks"`
	CompletedTasks int                `json:"completed_tasks"`
	FailedTasks    int                `json:"failed_tasks"`
	PendingTasks   int                `json:"pending_tasks"`
	RunningTasks   int                `json:"running_tasks"`
	SkippedTasks   int                `json:"skipped_tasks"`
	Tasks          []TaskReportItem   `json:"tasks"`
	StartTime      time.Time          `json:"start_time"`
	EndTime        time.Time          `json:"end_time,omitempty"`
	TotalDuration  int                `json:"total_duration_seconds,omitempty"`
}

// TaskReportItem 任务报告项
type TaskReportItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Assignee    string `json:"assignee"`
	Error       string `json:"error,omitempty"`
	Duration    int    `json:"duration_seconds"`
	Estimated   int    `json:"estimated_duration"`
	RetryCount  int    `json:"retry_count"`
}

// GetFailedTasks 获取失败的任务列表
func (ts *TaskScheduler) GetFailedTasks() []*SubTask {
	return ts.collection.ListTasksByStatus(SubTaskFailed)
}

// RetryFailedTasks 重试失败的任务
func (ts *TaskScheduler) RetryFailedTasks(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	failedTasks := ts.GetFailedTasks()
	for _, task := range failedTasks {
		if task.CanRetry() {
			task.SetStatus(SubTaskPending)
		}
	}

	// 重新调度
	return ts.Schedule(ctx, ScheduleOptions{
		ContinueOnError:  true,
		RetryFailedTasks: false,
	})
}

// SetTimeout 设置任务执行超时
func (ts *TaskScheduler) SetTimeout(timeout time.Duration) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.timeout = timeout
}

// SetMaxConcurrent 设置最大并发数
func (ts *TaskScheduler) SetMaxConcurrent(max int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.maxConcurrent = max
}

// AddAgent 添加Agent
func (ts *TaskScheduler) AddAgent(name string, agent Agent) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.agents[name] = agent
}

// RemoveAgent 移除Agent
func (ts *TaskScheduler) RemoveAgent(name string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.agents, name)
}

// GetAgents 获取所有Agent
func (ts *TaskScheduler) GetAgents() map[string]Agent {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	agents := make(map[string]Agent)
	for k, v := range ts.agents {
		agents[k] = v
	}
	return agents
}

// ScheduleTask schedules a task for execution
func (ts *TaskScheduler) ScheduleTask(ctx context.Context, task *AITask) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Store the task
	if ts.collection.ScheduledTasks == nil {
		ts.collection.ScheduledTasks = make(map[string]*AITask)
	}

	ts.collection.ScheduledTasks[task.ID] = task

	return task.ID, nil
}

// UnscheduleJob unschedules a job by ID
func (ts *TaskScheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.collection.ScheduledTasks == nil {
		return fmt.Errorf("no scheduled tasks found")
	}

	if _, exists := ts.collection.ScheduledTasks[jobID]; !exists {
		return fmt.Errorf("scheduled task not found: %s", jobID)
	}

	delete(ts.collection.ScheduledTasks, jobID)
	return nil
}
