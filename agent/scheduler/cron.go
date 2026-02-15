// Agent Framework - Cron Scheduler Implementation
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronScheduler Cron 调度器实现
type CronScheduler struct {
	cron    *cron.Cron
	jobs    map[int]context.CancelFunc
	mu       sync.RWMutex
	timezone string
}

// NewCronScheduler 创建 Cron 调度器
func NewCronScheduler(timezone string) *CronScheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Default to UTC if timezone is invalid
		loc = time.UTC
	}

	return &CronScheduler{
		cron: cron.New(
			cron.WithParser(cron.NewParser(cron.SecondOptional)),
			cron.WithLocation(loc),
		),
		jobs:    make(map[int]context.CancelFunc),
		timezone: timezone,
	}
}

// Start 启动 Cron 调度器
func (c *CronScheduler) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = ctx // Suppress unused warning
	c.cron.Start()
	return nil
}

// Stop 停止 Cron 调度器
func (c *CronScheduler) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel all jobs
	for _, cancel := range c.jobs {
		if cancel != nil {
			cancel()
		}
	}

	c.cron.Stop()
	return nil
}

// Schedule 添加定时任务
func (c *CronScheduler) Schedule(ctx context.Context, cronExpr string, handler JobHandler) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create cancelable context
	jobCtx, cancel := context.WithCancel(ctx)

	// Add job to cron
	entryID, err := c.cron.AddFunc(cronExpr, func() {
		if err := handler(jobCtx); err != nil {
			fmt.Printf("[CronScheduler] Job error: %v\n", err)
		}
	})
	if err != nil {
		cancel()
		return "", fmt.Errorf("failed to add cron job: %w", err)
	}

	// Convert entry ID to string for job ID
	jobID := strconv.FormatInt(int64(entryID), 10)

	// Store cancel function with integer key
	c.jobs[int(entryID)] = cancel

	return jobID, nil
}

// Unschedule 移除定时任务
func (c *CronScheduler) Unschedule(ctx context.Context, jobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := strconv.Atoi(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %s", jobID)
	}

	cancel, exists := c.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Cancel the job
	cancel()
	delete(c.jobs, id)

	return nil
}

// GetNextRunTime 获取下次运行时间
func (c *CronScheduler) GetNextRunTime(cronExpr string) (time.Time, error) {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return schedule.Next(time.Now()), nil
}

// ValidateCronExpr 验证 Cron 表达式
func ValidateCronExpr(cronExpr string) error {
	_, err := cron.ParseStandard(cronExpr)
	return err
}

// GetCronDescription 获取 Cron 表达式描述（中文）
func GetCronDescription(cronExpr string) string {
	// Parse cron expression
	parts := parseCronExpression(cronExpr)
	if len(parts) < 5 {
		return "无效的 Cron 表达式"
	}

	minute := parts[0]
	hour := parts[1]
	dayOfMonth := parts[2]
	month := parts[3]
	dayOfWeek := parts[4]

	// Build description
	desc := "每"

	// Minute
	if minute == "*" {
		desc += "分钟"
	} else {
		desc += fmt.Sprintf("%s 分钟", minute)
	}

	// Hour
	if hour == "*" {
		desc += "每小时"
	} else {
		desc += fmt.Sprintf(" %s点", hour)
	}

	// Day of month
	if dayOfMonth != "*" {
		desc += fmt.Sprintf(" %s号", dayOfMonth)
	}

	// Month
	if month != "*" {
		desc += fmt.Sprintf(" %s月", month)
	}

	// Day of week
	if dayOfWeek != "*" {
		desc += fmt.Sprintf(" %s", formatDayOfWeek(dayOfWeek))
	}

	return desc
}

// parseCronExpression 解析 Cron 表达式
func parseCronExpression(expr string) []string {
	parts := make([]string, 0, 6)
	current := ""
	for _, ch := range expr {
		if ch == ' ' {
			if len(current) > 0 {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}

// formatDayOfWeek 格式化星期（中文）
func formatDayOfWeek(day string) string {
	weekdayMap := map[string]string{
		"0": "周日",
		"1": "周一",
		"2": "周二",
		"3": "周三",
		"4": "周四",
		"5": "周五",
		"6": "周六",
		"SUN": "周日",
		"MON": "周一",
		"TUE": "周二",
		"WED": "周三",
		"THU": "周四",
		"FRI": "周五",
		"SAT": "周六",
	}

	if formatted, ok := weekdayMap[day]; ok {
		return formatted
	}
	return day
}
