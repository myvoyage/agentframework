// Agent Framework - Workflow Scheduler Demo
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/agent"
	"AgentFramework/agent/scheduler"
)

func main() {
	fmt.Println("=== AgentFramework 工作流调度器集成演示 ===")

	// 1. 创建工作流管理器
	fmt.Println("\n1. 创建工作流管理器...")
	skillLibrary := agent.NewSimpleSkillLibrary()
	modelFactory := agent.NewDefaultModelFactory()
	workflowManager := agent.NewWorkflowManager(skillLibrary, modelFactory)

	ctx := context.Background()
	workflowManager.Init(ctx)

	// 2. 创建调度器
	fmt.Println("2. 创建任务调度器...")
	cfg := scheduler.DefaultSchedulerConfig()
	taskScheduler := scheduler.NewTaskScheduler(cfg)

	if err := taskScheduler.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer taskScheduler.Stop(ctx)

	// 3. 设置 AI 代理
	fmt.Println("3. 设置 AI 代理...")
	taskScheduler.SetAgent(&demoAgent{})

	// 4. 创建工作流调度器管理器
	fmt.Println("4. 创建工作流调度器管理器...")
	wfScheduler := agent.NewWorkflowSchedulerManager(workflowManager, taskScheduler)

	// 5. 创建一个简单的工作流
	fmt.Println("5. 创建示例工作流...")
	workflowID := "demo_workflow"

	// 定义工作流
	workflowDef := `{
		"name": "Demo Workflow",
		"description": "A simple demo workflow",
		"type": "sequential",
		"nodes": [
			{
				"id": "node1",
				"type": "skill",
				"skill": "http_request",
				"config": {
					"url": "https://httpbin.org/get",
					"method": "GET"
				}
			},
			{
				"id": "node2",
				"type": "skill",
				"skill": "data_processing",
				"config": {
					"operation": "parse_json"
				}
			}
		]
	}`

	// 注册工作流
	_, err := workflowManager.RegisterWorkflow(ctx, workflowID, workflowDef, "v1.0")
	if err != nil {
		log.Printf("Warning: Failed to register workflow (demo continues): %v", err)
	}

	// 6. 安排工作流定时执行
	fmt.Println("\n6. 安排工作流定时执行...")

	// 每天早上9点执行
	task, err := wfScheduler.ScheduleWorkflow(
		ctx,
		workflowID,
		"0 9 * * *", // 每天早上9点
		map[string]string{
			"source":   "scheduler",
			"trigger":  "cron",
			"priority": "normal",
		},
	)
	if err != nil {
		log.Printf("Warning: Failed to schedule workflow (demo continues): %v", err)
	} else {
		fmt.Printf("✅ 工作流已安排:\n")
		fmt.Printf("   ID: %s\n", task.ID)
		fmt.Printf("   工作流: %s\n", task.WorkflowID)
		fmt.Printf("   Cron: %s\n", task.CronExpr)
		fmt.Printf("   下次运行: %s\n", task.NextRun.Format("2006-01-02 15:04:05"))
		fmt.Printf("   状态: %s\n", task.Status)
	}

	// 7. 立即执行工作流测试
	fmt.Println("\n7. 立即执行工作流测试...")
	result, err := wfScheduler.ExecuteWorkflowNow(ctx, workflowID, "Test input from scheduler demo")
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else {
		fmt.Printf("✅ 工作流执行完成:\n")
		fmt.Printf("   状态: %s\n", result.Status)
		fmt.Printf("   执行ID: %s\n", result.ExecutionID)
		fmt.Printf("   耗时: %dms\n", result.Duration)
		if result.Output != "" {
			fmt.Printf("   输出: %s\n", truncateString(result.Output, 100))
		}
	}

	// 8. 获取工作流任务状态
	fmt.Println("\n8. 获取工作流任务状态...")
	status, err := wfScheduler.GetWorkflowTaskStatus(workflowID)
	if err != nil {
		log.Printf("获取状态失败: %v", err)
	} else {
		fmt.Printf("工作流任务状态:\n")
		for key, value := range status {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	// 9. 列出所有工作流任务
	fmt.Println("\n9. 列出所有工作流任务...")
	tasks := wfScheduler.ListWorkflowTasks()
	fmt.Printf("共有 %d 个工作流任务:\n", len(tasks))
	for i, task := range tasks {
		if i >= 3 { // 只显示前3个
			fmt.Printf("   ... 还有 %d 个任务\n", len(tasks)-3)
			break
		}
		fmt.Printf("   [%d] %s - %s (%s)\n", i+1, task.Name, task.WorkflowID, task.Status)
	}

	// 10. 获取统计信息
	fmt.Println("\n10. 获取统计信息...")
	stats := wfScheduler.GetStatistics()
	fmt.Printf("调度器统计:\n")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// 11. 获取调度器统计
	fmt.Println("\n11. 获取调度器统计...")
	schedStats, err := taskScheduler.GetStats(ctx)
	if err != nil {
		log.Printf("获取调度器统计失败: %v", err)
	} else {
		fmt.Printf("任务调度器统计:\n")
		fmt.Printf("   总任务数: %d\n", schedStats.TotalJobs)
		fmt.Printf("   活跃任务: %d\n", schedStats.ActiveJobs)
		fmt.Printf("   已完成: %d\n", schedStats.CompletedJobs)
		fmt.Printf("   已失败: %d\n", schedStats.FailedJobs)
		fmt.Printf("   总运行次数: %d\n", schedStats.TotalRuns)
	}

	// 12. 等待一会儿观察调度器运行
	fmt.Println("\n12. 观察调度器运行...")
	fmt.Println("调度器将在后台持续运行...")
	fmt.Println("提示: 按 Ctrl+C 退出演示")

	// 等待用户中断
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		fmt.Println("\n演示超时，自动退出...")
	case <-ctx.Done():
		fmt.Println("\n用户中断，退出演示...")
	}

	// 清理
	fmt.Println("\n13. 清理资源...")
	if err := wfScheduler.Cleanup(ctx); err != nil {
		log.Printf("清理失败: %v", err)
	}

	fmt.Println("\n演示完成！")
}

// demoAgent 演示用 AI 代理
type demoAgent struct{}

func (d *demoAgent) Run(ctx context.Context, input string, opts ...interface{}) (string, error) {
	return fmt.Sprintf("[AI 响应] 处理请求: %s", input), nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
