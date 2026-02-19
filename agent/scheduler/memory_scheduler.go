// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Agent Framework Memory Scheduler Implementation
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MemoryScheduler struct {
	config   *SchedulerConfig
	jobs     map[string]*Job
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	startTime     time.Time
	totalRuns     int64
	totalFailures int64
	runningJobs   int64
	jobQueue chan *jobTask
}

type jobTask struct {
	job *Job
	ctx context.Context
}

func NewMemoryScheduler(config *SchedulerConfig) *MemoryScheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}
	return &MemoryScheduler{
		config:   config,
		jobs:     make(map[string]*Job),
		jobQueue: make(chan *jobTask, config.MaxConcurrentJobs),
	}
}

func (s *MemoryScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return fmt.Errorf("already started")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.startTime = time.Now()
	s.wg.Add(1)
	go s.scheduleLoop()
	for i := 0; i < s.config.MaxConcurrentJobs; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	s.config.Logger.Info("scheduler started")
	return nil
}

func (s *MemoryScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	s.cancel()
	s.running = false
	s.wg.Wait()
	s.config.Logger.Info("scheduler stopped")
	return nil
}

func (s *MemoryScheduler) ScheduleJob(ctx context.Context, job *Job) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	s.jobs[job.ID] = job
	job.CreatedAt = time.Now()
	job.Status = JobStatusPending
	job.Enabled = true
	job.NextRunAt = time.Now().Add(1 * time.Minute)
	return job.ID, nil
}

func (s *MemoryScheduler) ScheduleCron(ctx context.Context, cronExpr string, handler JobHandler) (string, error) {
	return s.ScheduleInterval(ctx, 1*time.Minute, handler)
}

func (s *MemoryScheduler) ScheduleInterval(ctx context.Context, interval time.Duration, handler JobHandler) (string, error) {
	job := &Job{Handler: handler, Schedule: JobSchedule{Type: ScheduleTypeInterval, Interval: interval}}
	return s.ScheduleJob(ctx, job)
}

func (s *MemoryScheduler) ScheduleOnce(ctx context.Context, delay time.Duration, handler JobHandler) (string, error) {
	job := &Job{Handler: handler, Schedule: JobSchedule{Type: ScheduleTypeOnce, Delay: delay}}
	return s.ScheduleJob(ctx, job)
}

func (s *MemoryScheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, jobID)
	return nil
}

func (s *MemoryScheduler) GetJob(ctx context.Context, jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	jobCopy := *job
	return &jobCopy, nil
}

func (s *MemoryScheduler) ListJobs(ctx context.Context) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}
	return jobs, nil
}

func (s *MemoryScheduler) GetStats(ctx context.Context) (*SchedulerStats, error) {
	return &SchedulerStats{TotalJobs: int64(len(s.jobs))}, nil
}

func (s *MemoryScheduler) PauseJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[jobID]; ok {
		job.Status = JobStatusPaused
		job.Enabled = false
	}
	return nil
}

func (s *MemoryScheduler) ResumeJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[jobID]; ok {
		job.Status = JobStatusPending
		job.Enabled = true
	}
	return nil
}

func (s *MemoryScheduler) IsJobRunning(jobID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if job, ok := s.jobs[jobID]; ok {
		return job.Status == JobStatusRunning
	}
	return false
}

func (s *MemoryScheduler) scheduleLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkJobs()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MemoryScheduler) checkJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, job := range s.jobs {
		if job.Enabled && job.NextRunAt.Before(now) {
			select {
			case s.jobQueue <- &jobTask{job: job, ctx: s.ctx}:
				job.Status = JobStatusRunning
			default:
			}
			job.NextRunAt = now.Add(1 * time.Minute)
		}
	}
}

func (s *MemoryScheduler) worker() {
	defer s.wg.Done()
	for {
		select {
		case task := <-s.jobQueue:
			s.executeJob(task)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MemoryScheduler) executeJob(task *jobTask) {
	job := task.job
	if err := job.Handler(task.ctx); err != nil {
		s.config.Logger.Error("job failed", "id", job.ID, "error", err)
		job.Status = JobStatusFailed
	} else {
		job.Status = JobStatusPending
	}
}

func (s *MemoryScheduler) calculateNextRun(job *Job) (time.Time, error) {
	return time.Now().Add(1 * time.Minute), nil
}
