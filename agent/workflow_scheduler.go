// Agent Framework - Workflow Scheduler Integration
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WorkflowSchedulerManager 管理工作流和调度器的集成
type WorkflowSchedulerManager struct {
	workflowManager *WorkflowManager
	scheduler       *TaskScheduler
	workflowTasks   map[string]*WorkflowTask // workflowID -> task
	taskWorkflows   map[string]string        // taskID -> workflowID
	mu              sync.RWMutex
}

// WorkflowTask 工作流任务
type WorkflowTask struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	WorkflowID      string            `json:"workflow_id"`
	WorkflowVersion int               `json:"workflow_version"`
	CronExpr        string            `json:"cron_expr"`
	Enabled         bool              `json:"enabled"`
	InputTemplate   map[string]string `json:"input_template"`  // 输入模板
	NextRun         time.Time         `json:"next_run"`
	LastRun         *time.Time        `json:"last_run,omitempty"`
	LastExecutionID string            `json:"last_execution_id,omitempty"`
	Status          string            `json:"status"` // pending, running, completed, failed
	RunCount        int64             `json:"run_count"`
	FailureCount    int64             `json:"failure_count"`
	LastStatus      string            `json:"last_status"`  // success, failure
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// WorkflowTaskResult 工作流任务执行结果
type WorkflowTaskResult struct {
	TaskID          string    `json:"task_id"`
	WorkflowID      string    `json:"workflow_id"`
	ExecutionID     string    `json:"execution_id"`
	Status          string    `json:"status"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        int64     `json:"duration_ms"`
	Input           string    `json:"input"`
	Output          string    `json:"output,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// NewWorkflowSchedulerManager 创建工作流调度器管理器
func NewWorkflowSchedulerManager(wm *WorkflowManager, scheduler *TaskScheduler) *WorkflowSchedulerManager {
	return &WorkflowSchedulerManager{
		workflowManager: wm,
		scheduler:       scheduler,
		workflowTasks:   make(map[string]*WorkflowTask),
		taskWorkflows:   make(map[string]string),
	}
}

// ScheduleWorkflow 安排工作流定时执行
func (m *WorkflowSchedulerManager) ScheduleWorkflow(ctx context.Context, workflowID string, cronExpr string, inputTemplate map[string]string) (*WorkflowTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证工作流存在
	m.workflowManager.mu.RLock()
	_, exists := m.workflowManager.workflows[workflowID]
	m.workflowManager.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// 计算下次运行时间
	nextRun, err := calculateNextRun(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// 创建工作流任务
	task := &WorkflowTask{
		ID:            generateTaskID(),
		Name:          fmt.Sprintf("Workflow_%s", workflowID),
		WorkflowID:    workflowID,
		CronExpr:      cronExpr,
		Enabled:       true,
		InputTemplate: inputTemplate,
		NextRun:       nextRun,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 创建调度任务
	schedTask := AITask(
		task.Name,
		cronExpr,
		fmt.Sprintf("Execute workflow: %s", workflowID),
	)

	// 注册到调度器
	jobID, err := m.scheduler.ScheduleTask(ctx, schedTask)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule task: %w", err)
	}

	task.ID = jobID
	m.workflowTasks[workflowID] = task
	m.taskWorkflows[jobID] = workflowID

	return task, nil
}

// UnscheduleWorkflow 取消工作流定时执行
func (m *WorkflowSchedulerManager) UnscheduleWorkflow(ctx context.Context, workflowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.workflowTasks[workflowID]
	if !exists {
		return fmt.Errorf("workflow not scheduled: %s", workflowID)
	}

	// 从调度器中取消
	if err := m.scheduler.UnscheduleJob(ctx, task.ID); err != nil {
		return fmt.Errorf("failed to unschedule task: %w", err)
	}

	delete(m.workflowTasks, workflowID)
	delete(m.taskWorkflows, task.ID)

	return nil
}

// GetWorkflowTask 获取工作流任务
func (m *WorkflowSchedulerManager) GetWorkflowTask(workflowID string) (*WorkflowTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.workflowTasks[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow task not found: %s", workflowID)
	}

	return task, nil
}

// ListWorkflowTasks 列出所有工作流任务
func (m *WorkflowSchedulerManager) ListWorkflowTasks() []*WorkflowTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*WorkflowTask, 0, len(m.workflowTasks))
	for _, task := range m.workflowTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// ExecuteWorkflowNow 立即执行工作流
func (m *WorkflowSchedulerManager) ExecuteWorkflowNow(ctx context.Context, workflowID string, input string) (*WorkflowTaskResult, error) {
	m.mu.Lock()
	task, exists := m.workflowTasks[workflowID]
	m.mu.Unlock()

	result := &WorkflowTaskResult{
		TaskID:     task.ID,
		WorkflowID: workflowID,
		Status:     "running",
		StartTime:  time.Now(),
		Input:      input,
	}

	// 执行工作流
	workflow, err := m.workflowManager.LoadWorkflow(ctx, workflowID)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.Duration = time.Since(result.StartTime).Milliseconds()
		return result, nil
	}

	output, err := workflow.Run(ctx, input)
	result.EndTime = time.Now()
	result.Duration = time.Since(result.StartTime).Milliseconds()

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()

		// 更新任务统计
		m.mu.Lock()
		if task != nil {
			task.FailureCount++
			task.LastStatus = "failure"
			task.LastExecutionID = generateExecutionID()
		}
		m.mu.Unlock()

		return result, nil
	}

	result.Status = "completed"
	result.ExecutionID = generateExecutionID()
	result.Output = output.Content

	// 更新任务统计
	now := time.Now()
	m.mu.Lock()
	if task != nil {
		task.RunCount++
		task.LastStatus = "success"
		task.LastExecutionID = result.ExecutionID
		task.LastRun = &now
		task.UpdatedAt = now
	}
	m.mu.Unlock()

	return result, nil
}

// EnableWorkflowTask 启用工作流任务
func (m *WorkflowSchedulerManager) EnableWorkflowTask(ctx context.Context, workflowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.workflowTasks[workflowID]
	if !exists {
		return fmt.Errorf("workflow task not found: %s", workflowID)
	}

	task.Enabled = true
	task.Status = "pending"
	task.UpdatedAt = time.Now()

	return nil
}

// DisableWorkflowTask 禁用工作流任务
func (m *WorkflowSchedulerManager) DisableWorkflowTask(ctx context.Context, workflowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.workflowTasks[workflowID]
	if !exists {
		return fmt.Errorf("workflow task not found: %s", workflowID)
	}

	task.Enabled = false
	task.Status = "disabled"
	task.UpdatedAt = time.Now()

	return nil
}

// UpdateWorkflowTask 更新工作流任务
func (m *WorkflowSchedulerManager) UpdateWorkflowTask(ctx context.Context, workflowID string, cronExpr string, inputTemplate map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.workflowTasks[workflowID]
	if !exists {
		return fmt.Errorf("workflow task not found: %s", workflowID)
	}

	// 重新计算下次运行时间
	nextRun, err := calculateNextRun(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// 取消旧任务
	if err := m.scheduler.UnscheduleJob(ctx, task.ID); err != nil {
		return fmt.Errorf("failed to unschedule old task: %w", err)
	}

	// 更新任务
	task.CronExpr = cronExpr
	task.InputTemplate = inputTemplate
	task.NextRun = nextRun
	task.UpdatedAt = time.Now()

	// 创建新调度任务
	schedTask := AITask(
		task.Name,
		cronExpr,
		fmt.Sprintf("Execute workflow: %s", workflowID),
	)

	jobID, err := m.scheduler.ScheduleTask(ctx, schedTask)
	if err != nil {
		return fmt.Errorf("failed to schedule new task: %w", err)
	}

	// 更新映射
	delete(m.taskWorkflows, task.ID)
	task.ID = jobID
	m.taskWorkflows[jobID] = workflowID

	return nil
}

// GetScheduledWorkflows 获取所有已安排的工作流
func (m *WorkflowSchedulerManager) GetScheduledWorkflows() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workflows := make([]string, 0, len(m.workflowTasks))
	for workflowID := range m.workflowTasks {
		workflows = append(workflows, workflowID)
	}

	return workflows
}

// GetWorkflowTaskStatus 获取工作流任务状态
func (m *WorkflowSchedulerManager) GetWorkflowTaskStatus(workflowID string) (map[string]interface{}, error) {
	task, err := m.GetWorkflowTask(workflowID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task_id":           task.ID,
		"workflow_id":       task.WorkflowID,
		"name":              task.Name,
		"enabled":           task.Enabled,
		"status":            task.Status,
		"cron_expression":   task.CronExpr,
		"next_run":          task.NextRun,
		"last_run":          task.LastRun,
		"last_execution_id": task.LastExecutionID,
		"last_status":       task.LastStatus,
		"run_count":         task.RunCount,
		"failure_count":     task.FailureCount,
		"success_rate":      float64(task.RunCount-task.FailureCount) / float64(task.RunCount),
		"created_at":        task.CreatedAt,
		"updated_at":        task.UpdatedAt,
	}, nil
}

// ProcessScheduledTask 处理调度任务（由调度器调用）
func (m *WorkflowSchedulerManager) ProcessScheduledTask(ctx context.Context, taskID string) (*WorkflowTaskResult, error) {
	m.mu.RLock()
	workflowID, exists := m.taskWorkflows[taskID]
	task := m.workflowTasks[workflowID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// 准备输入
	input := ""
	if len(task.InputTemplate) > 0 {
		// 从模板生成输入
		input = generateInputFromTemplate(task.InputTemplate)
	} else {
		input = fmt.Sprintf("Scheduled execution of workflow: %s", workflowID)
	}

	// 执行工作流
	return m.ExecuteWorkflowNow(ctx, workflowID, input)
}

// GetStatistics 获取统计信息
func (m *WorkflowSchedulerManager) GetStatistics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalTasks := len(m.workflowTasks)
	enabledTasks := 0
	totalRuns := int64(0)
	totalFailures := int64(0)

	for _, task := range m.workflowTasks {
		if task.Enabled {
			enabledTasks++
		}
		totalRuns += task.RunCount
		totalFailures += task.FailureCount
	}

	successRate := 0.0
	if totalRuns > 0 {
		successRate = float64(totalRuns-totalFailures) / float64(totalRuns)
	}

	return map[string]interface{}{
		"total_tasks":    totalTasks,
		"enabled_tasks":  enabledTasks,
		"disabled_tasks": totalTasks - enabledTasks,
		"total_runs":     totalRuns,
		"total_failures": totalFailures,
		"success_rate":   successRate,
	}
}

// Cleanup 清理资源
func (m *WorkflowSchedulerManager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 取消所有工作流任务
	for workflowID := range m.workflowTasks {
		if err := m.UnscheduleWorkflow(ctx, workflowID); err != nil {
			// 记录错误但继续清理
			continue
		}
	}

	return nil
}

// ==================== 辅助函数 ====================

// generateTaskID 生成任务 ID
func generateTaskID() string {
	return fmt.Sprintf("task_%s", uuid.New().String()[:8])
}

// generateExecutionID 生成执行 ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%s", uuid.New().String()[:8])
}

// calculateNextRun 计算下次运行时间
func calculateNextRun(cronExpr string) (time.Time, error) {
	// 这里使用 cron 库解析并计算下次运行时间
	// 简化实现，返回 1 小时后
	return time.Now().Add(time.Hour), nil
}

// generateInputFromTemplate 从模板生成输入
func generateInputFromTemplate(template map[string]string) string {
	input := ""
	for key, value := range template {
		if input != "" {
			input += ", "
		}
		input += fmt.Sprintf("%s: %s", key, value)
	}
	return input
}
