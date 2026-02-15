// Agent Framework Async Task Execution
package async

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
)

// TaskFunc 任务函数类型
type TaskFunc func(ctx context.Context, progress ProgressCallback) (interface{}, error)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress float64, message string)

// AsyncTask 异步任务接口
type AsyncTask interface {
	ID() string
	Name() string
	Status() TaskStatus
	CreatedAt() time.Time
	StartedAt() time.Time
	CompletedAt() time.Time
	Result() (interface{}, error)
	Cancel() error
	Wait(ctx context.Context) error
	Progress() float64
}

// TaskManager 任务管理器接口
type TaskManager interface {
	Submit(ctx context.Context, task TaskFunc, opts ...TaskOption) (AsyncTask, error)
	Get(taskID string) (AsyncTask, error)
	List(opts ...TaskListOption) []AsyncTask
	Cancel(taskID string) error
	WaitFor(ctx context.Context, taskID string) (interface{}, error)
	GetStats() *TaskManagerStats
}

// TaskManagerStats 任务管理器统计
type TaskManagerStats struct {
	TotalTasks     int64
	RunningTasks   int64
	CompletedTasks  int64
	FailedTasks    int64
	CanceledTasks  int64
	PendingTasks   int64
	AvgDuration    time.Duration
}

// TaskConfig 任务配置
type TaskConfig struct {
	Name        string
	Description string
	Tags        []string
	Timeout     time.Duration
	Priority    int
	Metadata    map[string]string
}

// TaskOption 任务选项
type TaskOption func(*TaskConfig)

// TaskListOption 任务列表选项
type TaskListOption func(*TaskListOptions)

// TaskListOptions 任务列表选项配置
type TaskListOptions struct {
	Status *TaskStatus
	Tags   []string
	Limit  int
	Offset int
}

// ===== 内存任务实现 =====

type memoryTask struct {
	id          string
	name        string
	description string
	tags        []string
	status      TaskStatus
	result      interface{}
	resultErr   error
	createdAt   time.Time
	startedAt   time.Time
	completedAt time.Time
	progress    float64
	progressMu  sync.Mutex
	cancel      context.CancelFunc
	cancelOnce  sync.Once
	timeout     time.Duration
	priority    int
	metadata    map[string]string
	ctx         context.Context
}

func (t *memoryTask) ID() string { return t.id }
func (t *memoryTask) Name() string { return t.name }
func (t *memoryTask) Status() TaskStatus { return t.status }
func (t *memoryTask) CreatedAt() time.Time { return t.createdAt }
func (t *memoryTask) StartedAt() time.Time { return t.startedAt }
func (t *memoryTask) CompletedAt() time.Time { return t.completedAt }
func (t *memoryTask) Result() (interface{}, error) { return t.result, t.resultErr }
func (t *memoryTask) Progress() float64 {
	t.progressMu.Lock()
	defer t.progressMu.Unlock()
	return t.progress
}

func (t *memoryTask) setProgress(progress float64, message string) {
	t.progressMu.Lock()
	defer t.progressMu.Unlock()
	if progress > 1.0 {
		progress = 1.0
	} else if progress < 0 {
		progress = 0
	}
	t.progress = progress
}

func (t *memoryTask) Cancel() error {
	t.cancelOnce.Do(func() {
		t.cancel()
		t.status = TaskStatusCanceled
	})
	return nil
}
func (t *memoryTask) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.ctx.Done():
		return nil
	}
}

// MemoryTaskManager 内存任务管理器
type MemoryTaskManager struct {
	tasks   map[string]*memoryTask
	mu      sync.RWMutex
	nextID  uint64
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewMemoryTaskManager 创建内存任务管理器
func NewMemoryTaskManager() *MemoryTaskManager {
	return &MemoryTaskManager{
		tasks: make(map[string]*memoryTask),
	}
}

// Start 启动任务管理器
func (m *MemoryTaskManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("task manager already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true
	return nil
}

// Stop 停止任务管理器
func (m *MemoryTaskManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	// 取消所有运行中的任务
	for _, task := range m.tasks {
		if task.status == TaskStatusRunning {
			task.Cancel()
		}
	}

	m.cancel()
	m.running = false
	m.wg.Wait()

	return nil
}

// Submit 提交异步任务
func (m *MemoryTaskManager) Submit(ctx context.Context, taskFn TaskFunc, opts ...TaskOption) (AsyncTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil, fmt.Errorf("task manager not running")
	}

	// 创建任务配置
	config := &TaskConfig{
		Name:     fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Timeout:  30 * time.Minute,
		Priority: 0,
	}

	for _, opt := range opts {
		opt(config)
	}

	// 创建任务上下文
	taskCtx, cancel := context.WithCancel(m.ctx)
	if config.Timeout > 0 {
		taskCtx, cancel = context.WithTimeout(taskCtx, config.Timeout)
	}

	// 创建任务
	task := &memoryTask{
		id:          m.generateID(),
		name:        config.Name,
		description: config.Description,
		tags:        config.Tags,
		status:      TaskStatusPending,
		createdAt:   time.Now(),
		cancel:      cancel,
		timeout:     config.Timeout,
		priority:    config.Priority,
		metadata:    config.Metadata,
		ctx:         taskCtx,
	}

	m.tasks[task.id] = task

	// 异步执行任务
	m.wg.Add(1)
	go m.executeTask(task, taskFn)

	return task, nil
}

// Get 获取任务
func (m *MemoryTaskManager) Get(taskID string) (AsyncTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// List 列出任务
func (m *MemoryTaskManager) List(opts ...TaskListOption) []AsyncTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	options := &TaskListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	result := make([]AsyncTask, 0)
	for _, task := range m.tasks {
		if options.Status != nil && task.status != *options.Status {
			continue
		}
		if len(options.Tags) > 0 && !hasAnyTag(task.tags, options.Tags) {
			continue
		}
		result = append(result, task)
	}

	return result
}

// Cancel 取消任务
func (m *MemoryTaskManager) Cancel(taskID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return task.Cancel()
}

// WaitFor 等待任务完成
func (m *MemoryTaskManager) WaitFor(ctx context.Context, taskID string) (interface{}, error) {
	task, err := m.Get(taskID)
	if err != nil {
		return nil, err
	}

	if err := task.Wait(ctx); err != nil {
		return nil, err
	}

	return task.Result()
}

// GetStats 获取统计信息
func (m *MemoryTaskManager) GetStats() *TaskManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &TaskManagerStats{
		TotalTasks: int64(len(m.tasks)),
	}

	for _, task := range m.tasks {
		switch task.status {
		case TaskStatusPending:
			stats.PendingTasks++
		case TaskStatusRunning:
			stats.RunningTasks++
		case TaskStatusCompleted:
			stats.CompletedTasks++
		case TaskStatusFailed:
			stats.FailedTasks++
		case TaskStatusCanceled:
			stats.CanceledTasks++
		}
	}

	return stats
}

// executeTask 执行任务
func (m *MemoryTaskManager) executeTask(task *memoryTask, taskFn TaskFunc) {
	defer m.wg.Done()

	// 检查是否已取消
	select {
	case <-task.ctx.Done():
		task.status = TaskStatusCanceled
		return
	default:
	}

	// 更新状态为运行中
	task.status = TaskStatusRunning
	task.startedAt = time.Now()

	// 执行任务
	result, err := taskFn(task.ctx, func(progress float64, message string) {
		task.setProgress(progress, message)
	})

	task.completedAt = time.Now()

	// 更新结果
	task.result = result
	task.resultErr = err

	if err != nil {
		task.status = TaskStatusFailed
	} else {
		task.status = TaskStatusCompleted
	}
}

func (m *MemoryTaskManager) generateID() string {
	m.nextID++
	return fmt.Sprintf("task-%d", m.nextID)
}

func hasAnyTag(tags []string, required []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}
	for _, req := range required {
		if tagSet[req] {
			return true
		}
	}
	return false
}

// ===== 任务选项 =====

// WithName 设置任务名称
func WithName(name string) TaskOption {
	return func(c *TaskConfig) {
		c.Name = name
	}
}

// WithDescription 设置任务描述
func WithDescription(desc string) TaskOption {
	return func(c *TaskConfig) {
		c.Description = desc
	}
}

// WithTimeout 设置任务超时
func WithTimeout(timeout time.Duration) TaskOption {
	return func(c *TaskConfig) {
		c.Timeout = timeout
	}
}

// WithTags 设置任务标签
func WithTags(tags ...string) TaskOption {
	return func(c *TaskConfig) {
		c.Tags = tags
	}
}

// WithPriority 设置任务优先级
func WithPriority(priority int) TaskOption {
	return func(c *TaskConfig) {
		c.Priority = priority
	}
}

// WithMetadata 设置任务元数据
func WithMetadata(key string, value string) TaskOption {
	return func(c *TaskConfig) {
		if c.Metadata == nil {
			c.Metadata = make(map[string]string)
		}
		c.Metadata[key] = value
	}
}

// ===== 任务列表选项 =====

// WithStatus 按状态过滤
func WithStatus(status TaskStatus) TaskListOption {
	return func(o *TaskListOptions) {
		o.Status = &status
	}
}

// WithTagsFilter 按标签过滤
func WithTagsFilter(tags ...string) TaskListOption {
	return func(o *TaskListOptions) {
		o.Tags = tags
	}
}

// WithLimit 限制结果数量
func WithLimit(limit int) TaskListOption {
	return func(o *TaskListOptions) {
		o.Limit = limit
	}
}

// WithOffset 设置偏移量
func WithOffset(offset int) TaskListOption {
	return func(o *TaskListOptions) {
		o.Offset = offset
	}
}
