// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Worker Pool 使用示例
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/pool"
)

func main() {
	// 创建工作协程池，并发度设置为 4
	pool := NewWorkerPool(4)

	// 模拟 10 个任务
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		task := i

		wg.Add(1)
		go func() {
			defer wg.Done()
			// 模拟耗时任务（100-500ms）
			duration := time.Duration(100+time.Duration(i*50)) * time.Millisecond
			time.Sleep(duration)

			// 使用 WorkerPool 执行任务
			taskID := fmt.Sprintf("Task %d", i)
			err := pool.SubmitAndWait(func() error {
				fmt.Printf("Task %s completed\n", taskID)
				return nil
			})
			if err != nil {
				fmt.Printf("Task %s failed: %v\n", taskID, err)
			}
		}(task)
	}

	// 等待所有任务完成
	wg.Wait()

	// 获取池统计信息
	stats := pool.Metrics()
	fmt.Println("\n=== Worker Pool 性能统计 ===")
	fmt.Printf("总工作协程数: %d\n", stats.TotalWorkers)
	fmt.Printf("活跃工作协程数: %d\n", stats.ActiveWorkers)
	fmt.Printf("总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("完成任务数: %d\n", stats.CompletedTasks)
	fmt.Printf("失败任务数: %d\n", stats.FailedTasks)
	fmt.Printf("待处理任务数: %d\n", stats.PendingTasks)
	fmt.Printf("池使用命中率: %.2f%%\n", calculateHitRate(stats))
}

// 停止工作协程池
	pool.Stop()

	fmt.Println("\n=== 停止工作协程池 ===")
	fmt.Println("工作协程池已停止")
}

func calculateHitRate(stats *WorkerPoolMetrics) float64 {
	if stats.TotalTasks == 0 {
		return 0.0
	}
	return float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
}

// 模拟任务执行函数
func simulateTaskExecution(taskID string) {
	// 模拟任务执行耗时：100-500ms
	duration := time.Duration(100+time.Duration(taskID[5]%50) * time.Millisecond
	time.Sleep(duration)
}