// Agent Framework - Memory Scheduler Implementation
// Copyright (C) 2025 Agent Framework Contributors
//
// MemoryScheduler is a fully in-process Scheduler implementation backed by
// robfig/cron for cron-expression scheduling and a worker-pool for interval
// and once-off jobs.
//
// Design:
//   - Cron jobs are delegated to an embedded CronScheduler (robfig/cron).
//   - Interval / once-off jobs are managed by an internal tick loop.
//   - All job metadata (RunCount, FailCount, Status, NextRunAt, LastRunAt)
//     is maintained in a sync.RWMutex-protected map.
//   - Retry logic: if EnableRetry is true and a job fails, it will be
//     re-queued up to MaxRetries times with RetryInterval delay.
//   - GetStats returns live aggregate statistics.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// cronEntry holds the mapping between a Job ID and the cron entry ID assigned
// by the underlying CronScheduler so we can unschedule individual jobs later.
type cronEntry struct {
	cronJobID string // string ID returned by CronScheduler.Schedule
	job       *Job
}

// intervalEntry tracks a timer-based (interval or once) job.
type intervalEntry struct {
	job     *Job
	cancel  context.CancelFunc
}

// MemoryScheduler is a fully in-process Scheduler implementation.
type MemoryScheduler struct {
	config *SchedulerConfig

	// cronScheduler handles all cron-expression-based jobs.
	cronScheduler *CronScheduler

	// cronEntries maps Job.ID → cronEntry so we can look up and remove cron jobs.
	cronEntries   map[string]*cronEntry
	cronEntriesMu sync.RWMutex

	// intervalEntries maps Job.ID → intervalEntry for interval/once jobs.
	intervalEntries   map[string]*intervalEntry
	intervalEntriesMu sync.RWMutex

	// jobs is the authoritative registry of all known jobs.
	jobs   map[string]*Job
	jobsMu sync.RWMutex

	// Worker pool for executing job handlers concurrently.
	jobQueue chan *jobTask
	workerWg sync.WaitGroup

	// Lifecycle
	running   bool
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex // guards running / startTime

	// Aggregate statistics (updated atomically)
	totalRuns     int64
	totalFailures int64
}

// jobTask is the unit of work dispatched to the worker pool.
type jobTask struct {
	job *Job
	ctx context.Context
}

// NewMemoryScheduler creates a new MemoryScheduler with the provided config.
// If config is nil the default config is used.
func NewMemoryScheduler(config *SchedulerConfig) *MemoryScheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}
	return &MemoryScheduler{
		config:          config,
		cronScheduler:   NewCronScheduler(config.Timezone),
		cronEntries:     make(map[string]*cronEntry),
		intervalEntries: make(map[string]*intervalEntry),
		jobs:            make(map[string]*Job),
		jobQueue:        make(chan *jobTask, config.MaxConcurrentJobs*4),
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Lifecycle
// ──────────────────────────────────────────────────────────────────────────────

// Start starts the scheduler.  It is idempotent: calling Start on a running
// scheduler returns immediately without error.
func (s *MemoryScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // already running
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.startTime = time.Now()

	// Start the underlying cron scheduler.
	if err := s.cronScheduler.Start(s.ctx); err != nil {
		s.cancel()
		s.running = false
		return fmt.Errorf("failed to start cron scheduler: %w", err)
	}

	// Launch worker goroutines.
	concurrency := s.config.MaxConcurrentJobs
	if concurrency <= 0 {
		concurrency = 5
	}
	for i := 0; i < concurrency; i++ {
		s.workerWg.Add(1)
		go s.worker(s.ctx)
	}

	s.config.Logger.Info("memory scheduler started",
		"concurrency", concurrency,
		"timezone", s.config.Timezone)
	return nil
}

// Stop gracefully stops the scheduler.  Ongoing job handlers are given up to
// their own context deadlines to complete.  Stop is idempotent.
func (s *MemoryScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Stop the cron scheduler first to prevent new triggers.
	if err := s.cronScheduler.Stop(ctx); err != nil {
		s.config.Logger.Error("error stopping cron scheduler", "error", err)
	}

	// Cancel all interval/once job contexts.
	s.intervalEntriesMu.Lock()
	for _, ie := range s.intervalEntries {
		ie.cancel()
	}
	s.intervalEntriesMu.Unlock()

	// Signal workers to stop and wait.
	s.cancel()
	s.workerWg.Wait()

	s.running = false
	s.config.Logger.Info("memory scheduler stopped")
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Job registration
// ──────────────────────────────────────────────────────────────────────────────

// ScheduleJob registers a Job in the scheduler.  The scheduling behaviour is
// determined by job.Schedule.Type:
//
//	ScheduleTypeCron     → registered with the embedded CronScheduler
//	ScheduleTypeInterval → launched in a recurring timer goroutine
//	ScheduleTypeOnce     → launched in a one-shot timer goroutine
func (s *MemoryScheduler) ScheduleJob(ctx context.Context, job *Job) (string, error) {
	if job == nil {
		return "", fmt.Errorf("job must not be nil")
	}
	if job.Handler == nil {
		return "", fmt.Errorf("job handler must not be nil")
	}

	// Assign an ID if none provided.
	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}

	// Initialise metadata.
	now := time.Now()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now
	job.Status = JobStatusPending
	if job.Metadata == nil {
		job.Metadata = make(map[string]string)
	}

	switch job.Schedule.Type {
	case ScheduleTypeCron:
		return s.scheduleCronJob(ctx, job)
	case ScheduleTypeInterval:
		return s.scheduleIntervalJob(ctx, job)
	case ScheduleTypeOnce:
		return s.scheduleOnceJob(ctx, job)
	default:
		// Default to interval if Interval is set; otherwise plain cron if CronExpr is set.
		if job.Schedule.CronExpr != "" {
			job.Schedule.Type = ScheduleTypeCron
			return s.scheduleCronJob(ctx, job)
		}
		if job.Schedule.Interval > 0 {
			job.Schedule.Type = ScheduleTypeInterval
			return s.scheduleIntervalJob(ctx, job)
		}
		// Register without active scheduling (disabled / manual run only).
		s.storeJob(job)
		return job.ID, nil
	}
}

// scheduleCronJob registers a cron-expression job with the embedded CronScheduler.
func (s *MemoryScheduler) scheduleCronJob(ctx context.Context, job *Job) (string, error) {
	if job.Schedule.CronExpr == "" {
		return "", fmt.Errorf("cron expression is required for cron jobs")
	}

	// Validate and pre-compute NextRunAt.
	tz := job.Schedule.Timezone
	if tz == "" {
		tz = s.config.Timezone
	}
	cronSched := NewCronScheduler(tz)
	nextRun, err := cronSched.GetNextRunTime(job.Schedule.CronExpr)
	if err != nil {
		return "", fmt.Errorf("invalid cron expression %q: %w", job.Schedule.CronExpr, err)
	}
	job.NextRunAt = nextRun

	// Build the handler closure that updates job metadata and dispatches to the pool.
	handler := s.buildHandler(job)

	// Register with the underlying CronScheduler.
	cronJobID, err := s.cronScheduler.Schedule(ctx, job.Schedule.CronExpr, handler)
	if err != nil {
		return "", fmt.Errorf("failed to schedule cron job: %w", err)
	}

	// Persist the cron entry mapping.
	s.cronEntriesMu.Lock()
	s.cronEntries[job.ID] = &cronEntry{cronJobID: cronJobID, job: job}
	s.cronEntriesMu.Unlock()

	s.storeJob(job)

	s.config.Logger.Info("cron job scheduled",
		"id", job.ID, "name", job.Name,
		"expr", job.Schedule.CronExpr, "next", nextRun.Format(time.RFC3339))

	return job.ID, nil
}

// scheduleIntervalJob registers a recurring-interval job.
func (s *MemoryScheduler) scheduleIntervalJob(ctx context.Context, job *Job) (string, error) {
	interval := job.Schedule.Interval
	if interval <= 0 {
		return "", fmt.Errorf("interval must be positive")
	}

	job.NextRunAt = time.Now().Add(interval)

	jobCtx, cancel := context.WithCancel(s.ctx)
	handler := s.buildHandler(job)

	ie := &intervalEntry{job: job, cancel: cancel}

	s.intervalEntriesMu.Lock()
	s.intervalEntries[job.ID] = ie
	s.intervalEntriesMu.Unlock()

	s.storeJob(job)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if job.Enabled {
					_ = handler(jobCtx)
				}
				// Update NextRunAt after each tick.
				s.jobsMu.Lock()
				if j, ok := s.jobs[job.ID]; ok {
					j.NextRunAt = time.Now().Add(interval)
				}
				s.jobsMu.Unlock()

			case <-jobCtx.Done():
				return
			}
		}
	}()

	s.config.Logger.Info("interval job scheduled",
		"id", job.ID, "name", job.Name, "interval", interval)

	return job.ID, nil
}

// scheduleOnceJob registers a single-fire delayed job.
func (s *MemoryScheduler) scheduleOnceJob(ctx context.Context, job *Job) (string, error) {
	delay := job.Schedule.Delay
	if delay <= 0 {
		delay = 0 // immediate
	}

	job.NextRunAt = time.Now().Add(delay)

	jobCtx, cancel := context.WithCancel(s.ctx)
	handler := s.buildHandler(job)

	ie := &intervalEntry{job: job, cancel: cancel}

	s.intervalEntriesMu.Lock()
	s.intervalEntries[job.ID] = ie
	s.intervalEntriesMu.Unlock()

	s.storeJob(job)

	go func() {
		defer cancel()
		select {
		case <-time.After(delay):
			if job.Enabled {
				_ = handler(jobCtx)
			}
			// Mark completed.
			s.jobsMu.Lock()
			if j, ok := s.jobs[job.ID]; ok {
				j.Status = JobStatusCompleted
				j.NextRunAt = time.Time{} // no next run
			}
			s.jobsMu.Unlock()

		case <-jobCtx.Done():
		}
	}()

	s.config.Logger.Info("once job scheduled",
		"id", job.ID, "name", job.Name, "delay", delay)

	return job.ID, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Convenience wrappers that satisfy the Scheduler interface
// ──────────────────────────────────────────────────────────────────────────────

// ScheduleCron creates and schedules a cron job from a bare cron expression and handler.
func (s *MemoryScheduler) ScheduleCron(ctx context.Context, cronExpr string, handler JobHandler) (string, error) {
	job := &Job{
		Name:    fmt.Sprintf("cron-%d", time.Now().UnixNano()),
		Enabled: true,
		Handler: handler,
		Schedule: JobSchedule{
			Type:     ScheduleTypeCron,
			CronExpr: cronExpr,
		},
	}
	return s.ScheduleJob(ctx, job)
}

// ScheduleInterval creates and schedules a recurring interval job.
func (s *MemoryScheduler) ScheduleInterval(ctx context.Context, interval time.Duration, handler JobHandler) (string, error) {
	job := &Job{
		Name:    fmt.Sprintf("interval-%d", time.Now().UnixNano()),
		Enabled: true,
		Handler: handler,
		Schedule: JobSchedule{
			Type:     ScheduleTypeInterval,
			Interval: interval,
		},
	}
	return s.ScheduleJob(ctx, job)
}

// ScheduleOnce creates and schedules a single-fire job.
func (s *MemoryScheduler) ScheduleOnce(ctx context.Context, delay time.Duration, handler JobHandler) (string, error) {
	job := &Job{
		Name:    fmt.Sprintf("once-%d", time.Now().UnixNano()),
		Enabled: true,
		Handler: handler,
		Schedule: JobSchedule{
			Type:  ScheduleTypeOnce,
			Delay: delay,
		},
	}
	return s.ScheduleJob(ctx, job)
}

// ──────────────────────────────────────────────────────────────────────────────
// CRUD
// ──────────────────────────────────────────────────────────────────────────────

// UnscheduleJob removes a job from the scheduler.  Active cron entries are
// removed from the underlying CronScheduler; interval/once goroutines are
// cancelled.
func (s *MemoryScheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	s.jobsMu.Lock()
	_, exists := s.jobs[jobID]
	if !exists {
		s.jobsMu.Unlock()
		return fmt.Errorf("job not found: %s", jobID)
	}
	delete(s.jobs, jobID)
	s.jobsMu.Unlock()

	// Remove from cron scheduler if applicable.
	s.cronEntriesMu.Lock()
	if entry, ok := s.cronEntries[jobID]; ok {
		_ = s.cronScheduler.Unschedule(ctx, entry.cronJobID)
		delete(s.cronEntries, jobID)
	}
	s.cronEntriesMu.Unlock()

	// Cancel interval/once goroutine if applicable.
	s.intervalEntriesMu.Lock()
	if ie, ok := s.intervalEntries[jobID]; ok {
		ie.cancel()
		delete(s.intervalEntries, jobID)
	}
	s.intervalEntriesMu.Unlock()

	s.config.Logger.Info("job unscheduled", "id", jobID)
	return nil
}

// GetJob returns a snapshot copy of a job by ID.
func (s *MemoryScheduler) GetJob(ctx context.Context, jobID string) (*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	// Return a copy to prevent external mutation.
	cp := *job
	return &cp, nil
}

// ListJobs returns snapshot copies of all registered jobs.
func (s *MemoryScheduler) ListJobs(ctx context.Context) ([]*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	jobs := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		cp := *j
		jobs = append(jobs, &cp)
	}
	return jobs, nil
}

// PauseJob marks a job as paused so the handler is skipped on the next trigger.
func (s *MemoryScheduler) PauseJob(ctx context.Context, jobID string) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job.Enabled = false
	job.Status = JobStatusPaused
	job.UpdatedAt = time.Now()
	s.config.Logger.Info("job paused", "id", jobID, "name", job.Name)
	return nil
}

// ResumeJob re-enables a paused job and recomputes NextRunAt.
func (s *MemoryScheduler) ResumeJob(ctx context.Context, jobID string) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job.Enabled = true
	job.Status = JobStatusPending
	job.UpdatedAt = time.Now()

	// Recompute NextRunAt for cron jobs.
	if job.Schedule.CronExpr != "" {
		tz := job.Schedule.Timezone
		if tz == "" {
			tz = s.config.Timezone
		}
		cronSched := NewCronScheduler(tz)
		if nextRun, err := cronSched.GetNextRunTime(job.Schedule.CronExpr); err == nil {
			job.NextRunAt = nextRun
		}
	} else if job.Schedule.Interval > 0 {
		job.NextRunAt = time.Now().Add(job.Schedule.Interval)
	}

	s.config.Logger.Info("job resumed", "id", jobID, "name", job.Name)
	return nil
}

// IsJobRunning returns true if the job is currently being executed.
func (s *MemoryScheduler) IsJobRunning(jobID string) bool {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	if job, ok := s.jobs[jobID]; ok {
		return job.Status == JobStatusRunning
	}
	return false
}

// ──────────────────────────────────────────────────────────────────────────────
// Statistics
// ──────────────────────────────────────────────────────────────────────────────

// GetStats returns live aggregate statistics about the scheduler.
func (s *MemoryScheduler) GetStats(ctx context.Context) (*SchedulerStats, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	stats := &SchedulerStats{
		TotalJobs:     int64(len(s.jobs)),
		TotalRuns:     atomic.LoadInt64(&s.totalRuns),
		TotalFailures: atomic.LoadInt64(&s.totalFailures),
	}

	var totalDuration time.Duration
	var durationCount int64

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
		// Approximate average duration (we track RunCount; duration not stored per job)
		if job.RunCount > 0 {
			durationCount += job.RunCount
		}
		_ = totalDuration
	}

	s.mu.Lock()
	if !s.startTime.IsZero() {
		stats.Uptime = time.Since(s.startTime)
	}
	s.mu.Unlock()

	return stats, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

// storeJob upserts a job into the in-memory map (locked).
func (s *MemoryScheduler) storeJob(job *Job) {
	s.jobsMu.Lock()
	s.jobs[job.ID] = job
	s.jobsMu.Unlock()
}

// buildHandler wraps a job handler with retry logic, status tracking,
// RunCount/FailCount updates and NextRunAt recomputation.
func (s *MemoryScheduler) buildHandler(job *Job) JobHandler {
	return func(ctx context.Context) error {
		// Check if paused before executing.
		s.jobsMu.RLock()
		j, ok := s.jobs[job.ID]
		if !ok || !j.Enabled {
			s.jobsMu.RUnlock()
			return nil // skipped
		}
		s.jobsMu.RUnlock()

		// Mark running.
		startTime := time.Now()
		s.jobsMu.Lock()
		if j2, ok2 := s.jobs[job.ID]; ok2 {
			j2.Status = JobStatusRunning
			j2.LastRunAt = startTime
		}
		s.jobsMu.Unlock()

		// Apply job-level timeout if configured.
		execCtx := ctx
		var execCancel context.CancelFunc
		if s.config.JobTimeout > 0 {
			execCtx, execCancel = context.WithTimeout(ctx, s.config.JobTimeout)
			defer execCancel()
		}

		// Execute with optional retry.
		var execErr error
		maxAttempts := 1
		if s.config.EnableRetry {
			maxAttempts = s.config.MaxRetries + 1
		}

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			execErr = job.Handler(execCtx)
			if execErr == nil {
				break
			}
			if attempt < maxAttempts {
				s.config.Logger.Warn("job handler failed, retrying",
					"id", job.ID, "attempt", attempt, "error", execErr)
				select {
				case <-time.After(s.config.RetryInterval):
				case <-execCtx.Done():
					execErr = execCtx.Err()
					break
				}
			}
		}

		// Update statistics and status.
		duration := time.Since(startTime)
		atomic.AddInt64(&s.totalRuns, 1)

		s.jobsMu.Lock()
		if j3, ok3 := s.jobs[job.ID]; ok3 {
			j3.RunCount++
			if execErr != nil {
				j3.Status = JobStatusFailed
				j3.FailCount++
				atomic.AddInt64(&s.totalFailures, 1)
				s.config.Logger.Error("job execution failed",
					"id", job.ID, "name", j3.Name,
					"error", execErr, "duration", duration)
			} else {
				j3.Status = JobStatusPending // ready for next trigger
				s.config.Logger.Info("job execution succeeded",
					"id", job.ID, "name", j3.Name, "duration", duration)
			}

			// Recompute NextRunAt for cron jobs.
			if j3.Schedule.CronExpr != "" {
				tz := j3.Schedule.Timezone
				if tz == "" {
					tz = s.config.Timezone
				}
				cronSched := NewCronScheduler(tz)
				if nextRun, err := cronSched.GetNextRunTime(j3.Schedule.CronExpr); err == nil {
					j3.NextRunAt = nextRun
				}
			}
		}
		s.jobsMu.Unlock()

		return execErr
	}
}

// worker drains the job queue and executes job handlers concurrently.
// Each job dispatched through ScheduleJob (non-cron) arrives here.
func (s *MemoryScheduler) worker(ctx context.Context) {
	defer s.workerWg.Done()
	for {
		select {
		case task, ok := <-s.jobQueue:
			if !ok {
				return
			}
			s.executeQueuedJob(task)
		case <-ctx.Done():
			return
		}
	}
}

// executeQueuedJob runs a job from the worker-pool queue.
func (s *MemoryScheduler) executeQueuedJob(task *jobTask) {
	if task == nil || task.job == nil || task.job.Handler == nil {
		return
	}
	handler := s.buildHandler(task.job)
	if err := handler(task.ctx); err != nil {
		s.config.Logger.Error("queued job execution failed",
			"id", task.job.ID, "error", err)
	}
}
