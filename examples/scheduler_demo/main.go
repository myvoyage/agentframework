// Agent Framework - Task Scheduler Demo
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/agent/scheduler"
)

func main() {
	// Create scheduler with default configuration
	cfg := scheduler.DefaultSchedulerConfig()
	sched := scheduler.NewTaskScheduler(cfg)

	if sched == nil {
		log.Fatal("Failed to create task scheduler")
	}

	// Set a mock AI agent for demonstration
	sched.SetAgent(&mockAgent{})

	// Start the scheduler
	ctx := context.Background()
	if err := sched.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer sched.Stop(ctx)

	fmt.Println("=== 定时任务系统演示 ===")
	fmt.Println("本演示展示如何创建和管理定时任务")

	// Example 1: Static message task - Daily reminder
	staticTask := scheduler.StaticTask(
		"每日晨会提醒",
		"0 9 * * *", // 9 AM every day
		"该开早会了！",
	)

	jobID, err := sched.ScheduleTask(ctx, staticTask)
	if err != nil {
		log.Printf("Failed to schedule static task: %v\n", err)
	} else {
		fmt.Printf("✅ 已创建静态消息任务: %s (ID: %s)\n", staticTask.Name, jobID)
		fmt.Printf("   - Cron 表达式: %s\n", staticTask.CronExpr)
		fmt.Printf("   - 下次运行: %s\n", staticTask.NextRun.Format("2006-01-02 15:04:05"))
	}

	// Example 2: AI-powered task - Weekly news summary
	aiTask := scheduler.AITask(
		"每周 AI 新闻摘要",
		"0 9 * * 1", // 9 AM every Monday
		"搜索最新 AI 新闻并生成摘要",
		scheduler.WithAIModel("deepseek-chat"),
		scheduler.WithAITools("web_search"),
	)

	jobID, err = sched.ScheduleTask(ctx, aiTask)
	if err != nil {
		log.Printf("Failed to schedule AI task: %v\n", err)
	} else {
		fmt.Printf("✅ 已创建 AI 智能任务: %s (ID: %s)\n", aiTask.Name, jobID)
		fmt.Printf("   - Cron 表达式: %s\n", aiTask.CronExpr)
		fmt.Printf("   - AI 模型: %s\n", *aiTask.AIModel)
		fmt.Printf("   - 可用工具: %v\n", aiTask.AITools)
	}

	// List all scheduled jobs
	jobs, err := sched.ListJobs(ctx)
	if err != nil {
		log.Printf("Failed to list jobs: %v\n", err)
	} else {
		fmt.Println("\n=== 当前计划任务 ===")
		for _, job := range jobs {
			status := "禁用"
			if job.Enabled {
				status = "启用"
			}
			fmt.Printf("- %s [%s] %s\n", job.ID, job.Name, status)
		}
	}

	// Get scheduler statistics
	stats, err := sched.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v\n", err)
	} else {
		fmt.Println("\n=== 调度器统计 ===")
		fmt.Printf("总任务数: %d\n", stats.TotalJobs)
		fmt.Printf("活跃任务: %d\n", stats.ActiveJobs)
		fmt.Printf("已完成: %d\n", stats.CompletedJobs)
		fmt.Printf("已失败: %d\n", stats.FailedJobs)
		fmt.Printf("总运行次数: %d\n", stats.TotalRuns)
	}

	// Example 3: Natural language parsing
	fmt.Println("\n=== 自然语言解析示例 ===")
	nlExamples := []string{
		"每天早上9点提醒我开会",
		"每小时43分发一段鸡汤激励我写代码",
		"每周一早上8点搜索AI新闻发给我摘要",
		"cron 0 9 * * * 开会",
		"cron 0 0 9 1 1 周一到周五开早会",
	}

	for _, example := range nlExamples {
		task, err := scheduler.ParseNaturalLanguage(example)
		if err != nil {
			fmt.Printf("❌ 解析失败: %s - %v\n", example, err)
		} else {
			fmt.Printf("✅ 解析成功: %s -> 任务类型: %s\n", example, task.Type)
			if task.Type == scheduler.TaskTypeStatic {
				fmt.Printf("   消息: %s\n", *task.StaticMessage)
			} else if task.Type == scheduler.TaskTypeAI {
				fmt.Printf("   AI 提示: %s\n", *task.AIPrompt)
				if task.AIModel != nil {
					fmt.Printf("   AI 模型: %s\n", *task.AIModel)
				}
				if len(task.AITools) > 0 {
					fmt.Printf("   工具: %v\n", task.AITools)
				}
			}
			fmt.Printf("   Cron: %s\n", task.CronExpr)
		}
	}

	fmt.Println("\n提示: 按 Ctrl+C 退出演示")
	fmt.Println("调度器将持续运行，直到您停止程序...")

	// Keep the program running
	select {
	case <-ctx.Done():
		fmt.Println("\n调度器已停止")
		return
	case <-time.After(10 * time.Minute):
		fmt.Println("\n演示将在 10 秒后自动退出...")
	}
}

// mockAgent is a mock AI agent for demonstration purposes
type mockAgent struct{}

func (m *mockAgent) Run(ctx context.Context, input string, opts ...interface{}) (string, error) {
	return fmt.Sprintf("[AI 模拟响应] 收到请求: %s", input), nil
}
