// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Worker Pool 性能测试
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/pool"
)

func main() {
	fmt.Println("=== Worker Pool 性能测试 ===")

	// 创建工作协程池，并发度设置为 4
	pool := NewWorkerPool(4)

	// 模拟 10 个快速任务
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("quick-task-%d", i)
		task := i

		wg.Add(1)
		go func() {
			defer wg.Done()
			// 模拟快速任务执行（10-50ms）
			time.Sleep(10 + time.Duration(i*10)*time.Millisecond)

			// 提交任务到池
			err := pool.SubmitAndWait(func() error {
				fmt.Printf("Task %s 执行成功\n", taskID)
				return nil
			})
			if err != nil {
				fmt.Printf("Task %s 执行失败: %v\n", taskID, err)
			}
		}(task)
	}

	// 等待所有任务完成
	wg.Wait()

	// 获取池统计信息
	stats := pool.Metrics()
	fmt.Printf("工作协程总数: %d\n", stats.TotalWorkers)
	fmt.Printf("活跃工作协程数: %d\n", stats.ActiveWorkers)
	fmt.Printf("总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("完成任务数: %d\n", stats.CompletedTasks)
	fmt.Printf("失败任务数: %d\n", stats.FailedTasks)
	fmt.Printf("待处理任务数: %d\n", stats.PendingTasks)

	fmt.Println("\n=== Worker Pool 性能测试完成 ===")
	fmt.Printf("平均执行时间: %v\n", calculateAverageExecutionTime(stats))
}

func calculateAverageExecutionTime(stats *WorkerPoolMetrics) time.Duration {
	if stats.CompletedTasks == 0 {
		return 0
	}
	totalTime := stats.TotalExecutionTime.Milliseconds()
	avgTime := time.Duration(totalTime / int64(stats.CompletedTasks))
	return avgTime
}

// 模拟任务执行函数 - 轻量级
func simulateLightTask(taskID string) {
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("Task %s 完成\n", taskID)
}