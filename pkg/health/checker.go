// Agent Framework - Health Check Package
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package health

import (
	"context"
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check is a function that performs a health check
type Check func(ctx context.Context) error

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Duration  int64     `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// Checker manages health checks
type Checker struct {
	checks map[string]Check
	mutex  sync.RWMutex
}

// New creates a new health checker
func New() *Checker {
	return &Checker{
		checks: make(map[string]Check),
	}
}

// Register registers a health check
func (c *Checker) Register(name string, check Check) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.checks[name] = check
}

// Unregister removes a health check
func (c *Checker) Unregister(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.checks, name)
}

// Check executes all registered health checks
func (c *Checker) Check(ctx context.Context) map[string]CheckResult {
	c.mutex.RLock()
	checks := make(map[string]Check)
	for name, check := range c.checks {
		checks[name] = check
	}
	c.mutex.RUnlock()

	results := make(map[string]CheckResult)

	for name, check := range checks {
		result := c.runCheck(ctx, name, check)
		results[name] = result
	}

	return results
}

func (c *Checker) runCheck(ctx context.Context, name string, check Check) CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      name,
		Timestamp: start,
	}

	err := check(ctx)
	result.Duration = time.Since(start).Milliseconds()

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = err.Error()
	} else {
		result.Status = StatusHealthy
	}

	return result
}

// OverallStatus determines the overall health status from check results
func (c *Checker) OverallStatus(results map[string]CheckResult) Status {
	unhealthyCount := 0

	for _, result := range results {
		if result.Status == StatusUnhealthy {
			unhealthyCount++
		}
	}

	if unhealthyCount == 0 {
		return StatusHealthy
	}

	if unhealthyCount == len(results) {
		return StatusUnhealthy
	}

	return StatusDegraded
}

// GetStatus returns a summary of all health checks
func (c *Checker) GetStatus(ctx context.Context) map[string]interface{} {
	results := c.Check(ctx)
	overallStatus := c.OverallStatus(results)

	return map[string]interface{}{
		"status":  overallStatus,
		"checks":  results,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}
