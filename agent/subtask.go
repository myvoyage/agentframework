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
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SubTaskStatus 子任务状态
type SubTaskStatus int

const (
	// SubTaskPending 待执行
	SubTaskPending SubTaskStatus = iota
	// SubTaskRunning 执行中
	SubTaskRunning
	// SubTaskCompleted 已完成
	SubTaskCompleted
	// SubTaskFailed 执行失败
	SubTaskFailed
	// SubTaskSkipped 已跳过
	SubTaskSkipped
)

// String 返回状态的字符串表示
func (s SubTaskStatus) String() string {
	switch s {
	case SubTaskPending:
		return "PENDING"
	case SubTaskRunning:
		return "RUNNING"
	case SubTaskCompleted:
		return "COMPLETED"
	case SubTaskFailed:
		return "FAILED"
	case SubTaskSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON 实现 JSON 序列化
func (s SubTaskStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// SubTaskPriority 子任务优先级
type SubTaskPriority int

const (
	// SubTaskPriorityLow 低优先级
	SubTaskPriorityLow SubTaskPriority = iota
	// SubTaskPriorityNormal 正常优先级
	SubTaskPriorityNormal
	// SubTaskPriorityHigh 高优先级
	SubTaskPriorityHigh
	// SubTaskPriorityCritical 紧急优先级
	SubTaskPriorityCritical
)

// String 返回优先级的字符串表示
func (p SubTaskPriority) String() string {
	switch p {
	case SubTaskPriorityLow:
		return "LOW"
	case SubTaskPriorityNormal:
		return "NORMAL"
	case SubTaskPriorityHigh:
		return "HIGH"
	case SubTaskPriorityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON 实现 JSON 序列化
func (p SubTaskPriority) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, p.String())), nil
}

// SubTask 子任务结构
type SubTask struct {
	// ID 子任务唯一标识
	ID string `json:"id"`
	// ParentID 父任务ID（如果是嵌套子任务）
	ParentID string `json:"parent_id,omitempty"`
	// Description 任务描述
	Description string `json:"description"`
	// Details 任务详细信息
	Details string `json:"details,omitempty"`
	// Dependencies 依赖的其他子任务ID列表
	Dependencies []string `json:"dependencies,omitempty"`
	// Assignee 负责执行的Agent名称
	Assignee string `json:"assignee,omitempty"`
	// Priority 任务优先级
	Priority SubTaskPriority `json:"priority"`
	// Status 任务状态
	Status SubTaskStatus `json:"status"`
	// Input 任务输入
	Input string `json:"input,omitempty"`
	// Output 任务输出
	Output string `json:"output,omitempty"`
	// Error 错误信息（如果执行失败）
	Error string `json:"error,omitempty"`
	// EstimatedDuration 预估执行时间（秒）
	EstimatedDuration int `json:"estimated_duration,omitempty"`
	// ActualDuration 实际执行时间（秒）
	ActualDuration int `json:"actual_duration,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// StartedAt 开始执行时间
	StartedAt *time.Time `json:"started_at,omitempty"`
	// CompletedAt 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Metadata 附加元数据
	Metadata map[string]any `json:"metadata,omitempty"`
	// RetryCount 重试次数
	RetryCount int `json:"retry_count"`
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries"`
	mu sync.RWMutex `json:"-"`
}

// NewSubTask 创建新的子任务
func NewSubTask(id, description string) *SubTask {
	return &SubTask{
		ID:          id,
		Description: description,
		Priority:    SubTaskPriorityNormal,
		Status:      SubTaskPending,
		CreatedAt:   time.Now(),
		MaxRetries:  3,
		Metadata:    make(map[string]any),
	}
}

// GetID 返回任务ID
func (t *SubTask) GetID() string {
	return t.ID
}

// SetStatus 设置任务状态
func (t *SubTask) SetStatus(status SubTaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = status

	now := time.Now()
	switch status {
	case SubTaskRunning:
		if t.StartedAt == nil {
			t.StartedAt = &now
		}
	case SubTaskCompleted, SubTaskFailed, SubTaskSkipped:
		if t.CompletedAt == nil {
			t.CompletedAt = &now
		}
		if t.StartedAt != nil {
			t.ActualDuration = int(now.Sub(*t.StartedAt).Seconds())
		}
	}
}

// GetStatus 获取任务状态
func (t *SubTask) GetStatus() SubTaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}

// SetOutput 设置任务输出
func (t *SubTask) SetOutput(output string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Output = output
}

// GetOutput 获取任务输出
func (t *SubTask) GetOutput() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Output
}

// SetError 设置错误信息
func (t *SubTask) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err != nil {
		t.Error = err.Error()
	} else {
		t.Error = ""
	}
}

// GetError 获取错误信息
func (t *SubTask) GetError() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Error
}

// IncrementRetry 增加重试计数
func (t *SubTask) IncrementRetry() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.RetryCount++
}

// CanRetry 检查是否可以重试
func (t *SubTask) CanRetry() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.RetryCount < t.MaxRetries
}

// AddDependency 添加依赖
func (t *SubTask) AddDependency(taskID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, dep := range t.Dependencies {
		if dep == taskID {
			return // 已存在
		}
	}
	t.Dependencies = append(t.Dependencies, taskID)
}

// RemoveDependency 移除依赖
func (t *SubTask) RemoveDependency(taskID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, dep := range t.Dependencies {
		if dep == taskID {
			t.Dependencies = append(t.Dependencies[:i], t.Dependencies[i+1:]...)
			break
		}
	}
}

// GetDependencies 获取依赖列表
func (t *SubTask) GetDependencies() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	deps := make([]string, len(t.Dependencies))
	copy(deps, t.Dependencies)
	return deps
}

// IsReadyToExecute 检查是否准备好执行（所有依赖都已完成）
func (t *SubTask) IsReadyToExecute(completedTasks map[string]bool) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Status != SubTaskPending {
		return false
	}

	for _, depID := range t.Dependencies {
		if !completedTasks[depID] {
			return false
		}
	}

	return true
}

// SetMetadata 设置元数据
func (t *SubTask) SetMetadata(key string, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata[key] = value
}

// GetMetadata 获取元数据
func (t *SubTask) GetMetadata(key string) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.Metadata == nil {
		return nil, false
	}
	val, ok := t.Metadata[key]
	return val, ok
}

// ToJSON 转换为JSON字符串
func (t *SubTask) ToJSON() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SubTaskFromJSON 从JSON字符串创建子任务
func SubTaskFromJSON(jsonStr string) (*SubTask, error) {
	var task SubTask
	if err := json.Unmarshal([]byte(jsonStr), &task); err != nil {
		return nil, err
	}
	task.Metadata = make(map[string]any)
	return &task, nil
}

// TaskCollection 任务集合，管理多个子任务
type TaskCollection struct {
	tasks  map[string]*SubTask
	rootID string
	mu     sync.RWMutex
}

// NewTaskCollection 创建任务集合
func NewTaskCollection() *TaskCollection {
	return &TaskCollection{
		tasks: make(map[string]*SubTask),
	}
}

// AddTask 添加任务
func (tc *TaskCollection) AddTask(task *SubTask) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if _, exists := tc.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	tc.tasks[task.ID] = task
	return nil
}

// GetTask 获取任务
func (tc *TaskCollection) GetTask(id string) (*SubTask, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	task, ok := tc.tasks[id]
	return task, ok
}

// RemoveTask 移除任务
func (tc *TaskCollection) RemoveTask(id string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.tasks, id)
}

// ListTasks 列出所有任务
func (tc *TaskCollection) ListTasks() []*SubTask {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tasks := make([]*SubTask, 0, len(tc.tasks))
	for _, task := range tc.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ListTasksByStatus 按状态列出任务
func (tc *TaskCollection) ListTasksByStatus(status SubTaskStatus) []*SubTask {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tasks := make([]*SubTask, 0)
	for _, task := range tc.tasks {
		if task.GetStatus() == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetReadyTasks 获取准备好执行的任务
func (tc *TaskCollection) GetReadyTasks() []*SubTask {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// 构建已完成任务集合
	completed := make(map[string]bool)
	for _, task := range tc.tasks {
		if task.GetStatus() == SubTaskCompleted {
			completed[task.ID] = true
		}
	}

	// 找出准备好执行的任务
	ready := make([]*SubTask, 0)
	for _, task := range tc.tasks {
		if task.IsReadyToExecute(completed) {
			ready = append(ready, task)
		}
	}

	return ready
}

// Count 返回任务总数
func (tc *TaskCollection) Count() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.tasks)
}

// CountByStatus 按状态统计任务数
func (tc *TaskCollection) CountByStatus() map[SubTaskStatus]int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	counts := make(map[SubTaskStatus]int)
	for _, task := range tc.tasks {
		status := task.GetStatus()
		counts[status]++
	}
	return counts
}

// Clear 清空所有任务
func (tc *TaskCollection) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tasks = make(map[string]*SubTask)
}

// ToJSON 转换为JSON字符串
func (tc *TaskCollection) ToJSON() (string, error) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tasks := make([]*SubTask, 0, len(tc.tasks))
	for _, task := range tc.tasks {
		tasks = append(tasks, task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
