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

package collaboration

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskScheduler defines the interface for task scheduling
type TaskScheduler interface {
	Submit(ctx context.Context, task *CollaborativeTask, member *TeamMember) (*TaskResult, error)
	SubmitBatch(ctx context.Context, tasks []*CollaborativeTask, members []*TeamMember, router *IntelligentRouter) ([]*TaskResult, error)
	GetStats() *SchedulerStats
	Shutdown(ctx context.Context) error
}

// DefaultTaskScheduler implements a default task scheduler with worker pool
type DefaultTaskScheduler struct {
	maxWorkers   int
	maxQueueSize int
	workerPool   *WorkerPool
	taskQueue    chan *ScheduledTask
	results      map[string]chan *TaskResult
	mu           sync.RWMutex
	shutdown     int32
	stats        *SchedulerStats
	ctx          context.Context
	cancel       context.CancelFunc
}

// ScheduledTask represents a task waiting to be executed
type ScheduledTask struct {
	Task       *CollaborativeTask
	Member     *TeamMember
	SubmitTime time.Time
	ResultChan chan *TaskResult
}

// SchedulerStats represents scheduler statistics
type SchedulerStats struct {
	TotalTasksSubmitted  int64
	TotalTasksCompleted  int64
	TotalTasksFailed     int64
	ActiveWorkers        int32
	QueuedTasks          int32
	AverageWaitTime      time.Duration
	AverageExecutionTime time.Duration
}

// NewDefaultTaskScheduler creates a new default task scheduler
func NewDefaultTaskScheduler(maxConcurrent int) TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &DefaultTaskScheduler{
		maxWorkers:   maxConcurrent,
		maxQueueSize: maxConcurrent * 10,
		taskQueue:    make(chan *ScheduledTask, maxConcurrent*10),
		results:      make(map[string]chan *TaskResult),
		stats:        &SchedulerStats{},
		ctx:          ctx,
		cancel:       cancel,
	}

	// Create worker pool
	s.workerPool = NewWorkerPool(maxConcurrent, maxConcurrent*2)

	// Start dispatcher
	go s.runDispatcher()

	return s
}

// Submit submits a task for execution
func (s *DefaultTaskScheduler) Submit(ctx context.Context, task *CollaborativeTask, member *TeamMember) (*TaskResult, error) {
	if atomic.LoadInt32(&s.shutdown) == 1 {
		return nil, fmt.Errorf("scheduler is shutdown")
	}

	// Create result channel
	resultChan := make(chan *TaskResult, 1)

	// Create scheduled task
	scheduledTask := &ScheduledTask{
		Task:       task,
		Member:     member,
		SubmitTime: time.Now(),
		ResultChan: resultChan,
	}

	// Store result channel
	s.mu.Lock()
	s.results[task.ID] = resultChan
	s.mu.Unlock()

	// Update stats
	atomic.AddInt64(&s.stats.TotalTasksSubmitted, 1)

	// Submit to queue
	select {
	case s.taskQueue <- scheduledTask:
		// Task queued successfully
	case <-ctx.Done():
		s.cleanupTask(task.ID)
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		s.cleanupTask(task.ID)
		return nil, fmt.Errorf("task queue timeout")
	}

	// Wait for result
	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		s.cleanupTask(task.ID)
		return nil, ctx.Err()
	case <-time.After(task.Timeout):
		s.cleanupTask(task.ID)
		return &TaskResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Errorf("task execution timeout"),
			Timestamp: time.Now(),
		}, nil
	}
}

// SubmitBatch submits multiple tasks for execution
func (s *DefaultTaskScheduler) SubmitBatch(ctx context.Context, tasks []*CollaborativeTask, members []*TeamMember, router *IntelligentRouter) ([]*TaskResult, error) {
	if atomic.LoadInt32(&s.shutdown) == 1 {
		return nil, fmt.Errorf("scheduler is shutdown")
	}

	// Create channels for results and errors
	results := make([]*TaskResult, len(tasks))
	errChan := make(chan error, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(index int, t *CollaborativeTask) {
			defer wg.Done()

			// Select member for this task
			member, err := router.SelectMember(ctx, members, t)
			if err != nil {
				errChan <- fmt.Errorf("task %s: failed to select member: %w", t.ID, err)
				return
			}

			// Create result channel
			resultChan := make(chan *TaskResult, 1)

			// Create scheduled task
			scheduledTask := &ScheduledTask{
				Task:       t,
				Member:     member,
				SubmitTime: time.Now(),
				ResultChan: resultChan,
			}

			// Store result channel
			s.mu.Lock()
			s.results[t.ID] = resultChan
			s.mu.Unlock()

			// Update stats
			atomic.AddInt64(&s.stats.TotalTasksSubmitted, 1)

			// Submit to queue
			select {
			case s.taskQueue <- scheduledTask:
				// Task queued successfully
			case <-ctx.Done():
				s.cleanupTask(t.ID)
				errChan <- ctx.Err()
				return
			case <-time.After(5 * time.Second):
				s.cleanupTask(t.ID)
				errChan <- fmt.Errorf("task %s: queue timeout", t.ID)
				return
			}

			// Wait for result
			select {
			case result := <-resultChan:
				results[index] = result
			case <-ctx.Done():
				s.cleanupTask(t.ID)
				errChan <- ctx.Err()
				return
			case <-time.After(t.Timeout):
				s.cleanupTask(t.ID)
				results[index] = &TaskResult{
					TaskID:    t.ID,
					Success:   false,
					Error:     fmt.Errorf("task execution timeout"),
					Timestamp: time.Now(),
				}
			}
		}(i, task)
	}

	// Wait for all tasks to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for errors
	var firstErr error
	for err := range errChan {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

// GetStats returns scheduler statistics
func (s *DefaultTaskScheduler) GetStats() *SchedulerStats {
	return s.stats
}

// Shutdown gracefully shuts down the scheduler
func (s *DefaultTaskScheduler) Shutdown(ctx context.Context) error {
	atomic.StoreInt32(&s.shutdown, 1)

	// Close task queue
	close(s.taskQueue)

	// Wait for active tasks to complete or timeout
	done := make(chan struct{})
	go func() {
		s.workerPool.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.cancel()
		return nil
	case <-ctx.Done():
		s.cancel()
		return ctx.Err()
	}
}

// runDispatcher runs the task dispatcher
func (s *DefaultTaskScheduler) runDispatcher() {
	for {
		select {
		case scheduledTask, ok := <-s.taskQueue:
			if !ok {
				// Queue closed, shutdown
				return
			}
			s.dispatchTask(scheduledTask)
		case <-s.ctx.Done():
			return
		}
	}
}

// dispatchTask dispatches a task to a worker
func (s *DefaultTaskScheduler) dispatchTask(scheduledTask *ScheduledTask) {
	// Submit to worker pool
	err := s.workerPool.Submit(s.ctx, &Job{
		ID:       scheduledTask.Task.ID,
		Task:     scheduledTask.Task,
		Member:   scheduledTask.Member,
		SubmitAt: scheduledTask.SubmitTime,
		Callback: func(result *TaskResult) {
			s.handleTaskResult(scheduledTask.Task.ID, result)
		},
	})

	if err != nil {
		// Failed to dispatch, return error result
		result := &TaskResult{
			TaskID:    scheduledTask.Task.ID,
			Success:   false,
			Error:     fmt.Errorf("failed to dispatch task: %w", err),
			Timestamp: time.Now(),
		}
		s.handleTaskResult(scheduledTask.Task.ID, result)
	}
}

// handleTaskResult handles a task result
func (s *DefaultTaskScheduler) handleTaskResult(taskID string, result *TaskResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get result channel
	resultChan, ok := s.results[taskID]
	if !ok {
		return // Task already cleaned up
	}

	// Send result
	select {
	case resultChan <- result:
	default:
		// Channel closed or full, skip
	}

	// Update stats
	if result.Success {
		atomic.AddInt64(&s.stats.TotalTasksCompleted, 1)
	} else {
		atomic.AddInt64(&s.stats.TotalTasksFailed, 1)
	}

	// Cleanup
	delete(s.results, taskID)
}

// cleanupTask cleans up a task
func (s *DefaultTaskScheduler) cleanupTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if resultChan, ok := s.results[taskID]; ok {
		close(resultChan)
		delete(s.results, taskID)
	}
}

// WorkerPool implements a worker pool for concurrent task execution
type WorkerPool struct {
	maxWorkers int
	maxQueue   int
	workers    []*Worker
	jobQueue   chan *Job
	activeJobs sync.WaitGroup
	shutdown   int32
	mu         sync.RWMutex
}

// Worker represents a worker in the pool
type Worker struct {
	id       int
	jobQueue <-chan *Job
	quit     chan bool
}

// Job represents a job to be executed
type Job struct {
	ID       string
	Task     *CollaborativeTask
	Member   *TeamMember
	SubmitAt time.Time
	Callback func(*TaskResult)
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers, maxQueue int) *WorkerPool {
	pool := &WorkerPool{
		maxWorkers: maxWorkers,
		maxQueue:   maxQueue,
		jobQueue:   make(chan *Job, maxQueue),
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		worker := &Worker{
			id:       i,
			jobQueue: pool.jobQueue,
			quit:     make(chan bool),
		}
		pool.workers = append(pool.workers, worker)
		pool.activeJobs.Add(1)
		go worker.Start()
	}

	return pool
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(ctx context.Context, job *Job) error {
	if atomic.LoadInt32(&wp.shutdown) == 1 {
		return fmt.Errorf("worker pool is shutdown")
	}

	select {
	case wp.jobQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("worker pool queue timeout")
	}
}

// Wait waits for all active jobs to complete
func (wp *WorkerPool) Wait() {
	wp.activeJobs.Wait()
}

// Shutdown shuts down the worker pool
func (wp *WorkerPool) Shutdown() {
	atomic.StoreInt32(&wp.shutdown, 1)
	close(wp.jobQueue)

	// Stop all workers
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	for _, worker := range wp.workers {
		worker.Stop()
	}
}

// Start starts the worker
func (w *Worker) Start() {
	for {
		select {
		case job := <-w.jobQueue:
			if job != nil {
				w.executeJob(job)
			}
		case <-w.quit:
			return
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}

// executeJob executes a job
func (w *Worker) executeJob(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			job.Callback(&TaskResult{
				TaskID:    job.ID,
				Success:   false,
				Error:     fmt.Errorf("panic: %v", r),
				Timestamp: time.Now(),
			})
		}
	}()

	// Execute the task
	startTime := time.Now()

	// Increment active tasks
	job.Member.mu.Lock()
	job.Member.ActiveTasks++
	job.Member.State = StateBusy
	job.Member.mu.Unlock()

	// Execute the agent
	result, err := job.Member.Agent.Run(context.Background(), job.Task.Input)

	duration := time.Since(startTime)

	// Decrement active tasks
	job.Member.mu.Lock()
	job.Member.ActiveTasks--
	if job.Member.ActiveTasks == 0 {
		job.Member.State = StateIdle
	}
	job.Member.mu.Unlock()

	// Prepare result
	taskResult := &TaskResult{
		TaskID:    job.ID,
		AgentName: job.Member.Agent.Name(),
		AgentRole: job.Member.Role,
		Success:   err == nil,
		Error:     err,
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if result != nil {
		taskResult.Output = result.Content
	}

	// Update wait time stats
	taskResult.Metadata["wait_time"] = startTime.Sub(job.SubmitAt)

	// Call callback
	job.Callback(taskResult)
}
