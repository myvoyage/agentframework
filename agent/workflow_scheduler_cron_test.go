// Agent Framework - Workflow Scheduler Cron Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"testing"
	"time"
)

// TestCalculateNextRun 测试cron表达式解析和下次运行时间计算
func TestCalculateNextRun(t *testing.T) {
	testCases := []struct {
		name       string
		cronExpr   string
		shouldFail bool
	}{
		{
			name:       "每分钟运行",
			cronExpr:   "* * * * *",
			shouldFail: false,
		},
		{
			name:       "每小时运行",
			cronExpr:   "0 * * * *",
			shouldFail: false,
		},
		{
			name:       "每天午夜运行",
			cronExpr:   "0 0 * * *",
			shouldFail: false,
		},
		{
			name:       "每周一早上9点运行",
			cronExpr:   "0 9 * * 1",
			shouldFail: false,
		},
		{
			name:       "每月1号运行",
			cronExpr:   "0 0 1 * *",
			shouldFail: false,
		},
		{
			name:       "每5分钟运行",
			cronExpr:   "*/5 * * * *",
			shouldFail: false,
		},
		{
			name:       "工作日早上9点运行",
			cronExpr:   "0 9 * * 1-5",
			shouldFail: false,
		},
		{
			name:       "使用月份简写",
			cronExpr:   "0 0 1 JAN,MAR,MAY *",
			shouldFail: false,
		},
		{
			name:       "使用星期简写",
			cronExpr:   "0 0 * * MON,WED,FRI",
			shouldFail: false,
		},
		{
			name:       "使用范围和步长",
			cronExpr:   "0-59/5 * * * *",
			shouldFail: false,
		},
		{
			name:       "每小时30分运行",
			cronExpr:   "30 * * * *",
			shouldFail: false,
		},
		{
			name:       "每天12点和18点运行",
			cronExpr:   "0 12,18 * * *",
			shouldFail: false,
		},
		{
			name:       "无效的cron表达式 - 分钟超出范围",
			cronExpr:   "60 * * * *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 小时超出范围",
			cronExpr:   "0 25 * * *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 日期超出范围",
			cronExpr:   "0 0 32 * *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 月份超出范围",
			cronExpr:   "0 0 1 13 *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 星期超出范围",
			cronExpr:   "0 0 * * 8",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 字段过少",
			cronExpr:   "0 0 * *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 字段过多",
			cronExpr:   "0 0 * * * * *",
			shouldFail: true,
		},
		{
			name:       "无效的cron表达式 - 空表达式",
			cronExpr:   "",
			shouldFail: true,
		},
		{
			name:       "有效的带秒的6字段表达式",
			cronExpr:   "0 * * * * *",
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextRun, err := calculateNextRun(tc.cronExpr)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected error for cron expression '%s', but got none", tc.cronExpr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for cron expression '%s', but got: %v", tc.cronExpr, err)
					return
				}

				if nextRun.IsZero() {
					t.Error("Expected non-zero next run time")
					return
				}

				// 验证下次运行时间在未来
				if nextRun.Before(time.Now()) {
					t.Error("Expected next run time to be in the future")
				}

				t.Logf("Cron expression '%s' will run next at: %v", tc.cronExpr, nextRun)
			}
		})
	}
}

// TestCronExpressionComplexPatterns 测试复杂模式
func TestCronExpressionComplexPatterns(t *testing.T) {
	testCases := []struct {
		name        string
		cronExpr    string
		description string
	}{
		{
			name:        "每个工作日的每10分钟",
			cronExpr:    "*/10 9-17 * * 1-5",
			description: "每10分钟运行，仅限工作日9:00-17:59",
		},
		{
			name:        "每季度运行",
			cronExpr:    "0 0 1 1,4,7,10 *",
			description: "每季度第一天午夜运行",
		},
		{
			name:        "每小时0分和30分运行",
			cronExpr:    "0,30 * * * *",
			description: "每小时半点和整点运行",
		},
		{
			name:        "每天特定时间",
			cronExpr:    "0 9,12,18 * * *",
			description: "每天9:00, 12:00, 18:00运行",
		},
		{
			name:        "每周特定几天",
			cronExpr:    "0 10 * * MON,WED,FRI",
			description: "周一、三、五的10:00运行",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextRun, err := calculateNextRun(tc.cronExpr)
			if err != nil {
				t.Errorf("Cron expression '%s' should be valid: %v", tc.cronExpr, err)
				return
			}

			if nextRun.IsZero() {
				t.Error("Expected non-zero next run time")
				return
			}

			// 验证下次运行时间在未来
			if nextRun.Before(time.Now()) {
				t.Errorf("Expected next run time to be in the future, got: %v", nextRun)
			}

			t.Logf("Cron: %s (%s) -> Next: %v", tc.cronExpr, tc.description, nextRun)
		})
	}
}

// TestCronExpressionPerformance 测试解析性能
func TestCronExpressionPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cronExpr := "*/5 * * * *"
	iterations := 1000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := calculateNextRun(cronExpr)
		if err != nil {
			t.Errorf("Failed to calculate next run: %v", err)
			return
		}
	}
	duration := time.Since(start)

	avgDuration := duration / time.Duration(iterations)
	t.Logf("Average parsing time: %v (total: %v for %d iterations)", avgDuration, duration, iterations)

	// 性能要求：每次解析应该在10ms以内
	if avgDuration > 10*time.Millisecond {
		t.Errorf("Parsing performance is too slow: %v (expected < 10ms)", avgDuration)
	}
}
