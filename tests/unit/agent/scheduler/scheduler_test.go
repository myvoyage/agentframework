// Agent Framework - Task Scheduler Tests
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestStaticTask(t *testing.T) {
	task := StaticTask(
		"测试任务",
		"0 * * * * *", // Every second
		"测试消息",
	)

	if task == nil {
		t.Fatal("Failed to create static task")
	}

	if task.Type != TaskTypeStatic {
		t.Errorf("Expected task type %s, got %s", TaskTypeStatic, task.Type)
	}

	if task.StaticMessage == nil || *task.StaticMessage != "测试消息" {
		t.Errorf("Expected static message '测试消息', got %v", task.StaticMessage)
	}
}

func TestAITask(t *testing.T) {
	task := AITask(
		"AI 测试任务",
		"0 * * * * *",
		"搜索 AI 新闻并生成摘要",
		WithAIModel("gpt-4"),
		WithAITools("web_search", "weather"),
	)

	if task == nil {
		t.Fatal("Failed to create AI task")
	}

	if task.Type != TaskTypeAI {
		t.Errorf("Expected task type %s, got %s", TaskTypeAI, task.Type)
	}

	if task.AIPrompt == nil || *task.AIPrompt != "搜索 AI 新闻并生成摘要" {
		t.Errorf("Expected AI prompt '搜索 AI 新闻并生成摘要', got %v", task.AIPrompt)
	}

	if task.AIModel == nil || *task.AIModel != "gpt-4" {
		t.Errorf("Expected AI model 'gpt-4', got %v", task.AIModel)
	}

	expectedTools := []string{"web_search", "weather"}
	if len(task.AITools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(task.AITools))
	}
}

func TestNewCronScheduler(t *testing.T) {
	cs := NewCronScheduler("Asia/Shanghai")
	if cs == nil {
		t.Fatal("Failed to create cron scheduler")
	}

	if cs.cron == nil {
		t.Error("Cron instance not initialized")
	}

	if cs.jobs == nil {
		t.Error("Jobs map not initialized")
	}
}

func TestCronSchedulerSchedule(t *testing.T) {
	cs := NewCronScheduler("Asia/Shanghai")
	if cs == nil {
		t.Fatal("Failed to create cron scheduler")
	}

	ctx := context.Background()

	// Test starting
	if err := cs.Start(ctx); err != nil {
		t.Fatalf("Failed to start cron scheduler: %v", err)
	}
	defer cs.Stop(ctx)

	executed := false
	handler := func(ctx context.Context) error {
		executed = true
		return nil
	}

	// Schedule a task to run every second
	jobID, err := cs.Schedule(ctx, "*/1 * * * * *", handler)
	if err != nil {
		t.Fatalf("Failed to schedule job: %v", err)
	}

	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}

	// Wait for execution
	time.Sleep(2 * time.Second)

	if !executed {
		t.Error("Job was not executed")
	}

	// Test unscheduling
	if err := cs.Unschedule(ctx, jobID); err != nil {
		t.Errorf("Failed to unschedule job: %v", err)
	}
}

func TestValidateCronExpr(t *testing.T) {
	tests := []struct {
		cronExpr  string
		wantError bool
	}{
		{
			cronExpr: "0 * * * *",
			wantError: false,
		},
		{
			cronExpr: "0 9 * * *",
			wantError: false,
		},
		{
			cronExpr: "0 9 * * * 1",
			wantError: false,
		},
		{
			cronExpr: "invalid",
			wantError: true,
		},
		{
			cronExpr: "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.cronExpr, func(t *testing.T) {
			err := ValidateCronExpr(tt.cronExpr)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseChineseTime(t *testing.T) {
	tests := []struct {
		input      string
		wantHour   int
		wantMinute int
	}{
		{
			input:      "早上9点",
			wantHour:   9,
			wantMinute: 0,
		},
		{
			input:      "下午3点30",
			wantHour:   15,
			wantMinute: 30,
		},
		{
			input:      "晚上8点",
			wantHour:   20,
			wantMinute: 0,
		},
		{
			input:      "10点30",
			wantHour:   10,
			wantMinute: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			hour, minute, err := parseChineseTime(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if hour != tt.wantHour || minute != tt.wantMinute {
				t.Errorf("Got hour=%d minute=%d, want hour=%d minute=%d",
					hour, minute, tt.wantHour, tt.wantMinute)
			}
		})
	}
}

func TestParseChineseWeekday(t *testing.T) {
	tests := []struct {
		input   string
		wantInt int
	}{
		{input: "一", wantInt: 1},
		{input: "周一", wantInt: 1},
		{input: "二", wantInt: 2},
		{input: "周二", wantInt: 2},
		{input: "三", wantInt: 3},
		{input: "周三", wantInt: 3},
		{input: "四", wantInt: 4},
		{input: "周四", wantInt: 4},
		{input: "五", wantInt: 5},
		{input: "周五", wantInt: 5},
		{input: "六", wantInt: 6},
		{input: "周六", wantInt: 6},
		{input: "日", wantInt: 0},
		{input: "周日", wantInt: 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseChineseWeekday(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.wantInt {
				t.Errorf("Got %d, want %d", result, tt.wantInt)
			}
		})
	}
}

func TestGetCronDescription(t *testing.T) {
	tests := []struct {
		cronExpr string
	}{
		{cronExpr: "0 * * * * *", desc: "每分钟"},
		{cronExpr: "0 9 * * * *", desc: "每天 9点"},
		{cronExpr: "0 9 * * * 1", desc: "每周一 9点"},
		{cronExpr: "*/5 * * * *", desc: "每 5 分钟"},
	}

	for _, tt := range tests {
		t.Run(tt.cronExpr, func(t *testing.T) {
			desc := GetCronDescription(tt.cronExpr)
			if desc == "" {
				t.Error("Expected non-empty description")
			}
			t.Logf("Cron: %s -> Description: %s", tt.cronExpr, desc)
		})
	}
}

func BenchmarkTaskScheduler(b *testing.B) {
	cfg := DefaultSchedulerConfig()
	s := NewTaskScheduler(cfg)

	ctx := context.Background()

	task := StaticTask("benchmark", "0 * * * * *", "message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ScheduleTask(ctx, task)
	}
	_ = ctx
}
