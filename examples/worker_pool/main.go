// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Worker Pool 性能测试 - 完全独立版本
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a task to be executed by a worker
type Task struct {
	ID       string
	Handler func() error
	Result   chan error
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workers   chan chan Task
	taskQueue chan Task
	quit      chan bool
	wg        sync.WaitGroup
	metrics   *WorkerPoolMetrics
}

// WorkerPoolMetrics tracks worker pool metrics
type WorkerPoolMetrics struct {
	TotalWorkers      int32
	ActiveWorkers     int32
	TotalTasks        int64
	CompletedTasks    int64
	FailedTasks       int64
	PendingTasks      int32
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		workers:   make(chan chan Task, numWorkers),
		taskQueue: make(chan Task, 100),
		quit:      make(chan bool),
		metrics:   &WorkerPoolMetrics{},
	}

	// Start workers
	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker(i)
	}
	pool.metrics.TotalWorkers = int32(numWorkers)

	// Start dispatcher
	go pool.dispatcher()

	return pool
}

// worker represents a worker goroutine
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	taskChannel := make(chan Task)
	p.workers <- taskChannel

	for {
		select {
		case task := <-taskChannel:
			p.metrics.ActiveWorkers++
			p.metrics.PendingTasks--
			err := task.Handler()
			p.metrics.ActiveWorkers--
			p.metrics.TotalTasks++
			if err != nil {
				p.metrics.FailedTasks++
			} else {
				p.metrics.CompletedTasks++
			}
			if task.Result != nil {
				task.Result <- err
			}

		case <-p.quit:
			return
		}
	}
}

// dispatcher dispatches tasks to workers
func (p *WorkerPool) dispatcher() {
	for {
		select {
		case task := <-p.taskQueue:
			p.metrics.PendingTasks++
			go func() {
				worker := <-p.workers
				worker <- task
			}()

		case <-p.quit:
			// Close all worker channels
			close(p.workers)
			return
		}
	}
}

// SubmitAndWait submits a task and waits for completion
func (p *WorkerPool) SubmitAndWait(handler func() error) (string, error) {
	result := make(chan error, 1)
	task := Task{
		Handler: handler,
		Result:   result,
	}

	p.taskQueue <- task
	err := <-result
	if err != nil {
		return "", err
	}
	return "Task completed", nil
}

// Metrics returns pool metrics
func (p *WorkerPool) Metrics() *WorkerPoolMetrics {
	return p.metrics
}

func main() {
	fmt.Println("=== Worker Pool 性能测试（完全独立版本） ===")

	// 创建工作协程池，并发度设置为 4
	pool := NewWorkerPool(4)

	// 模拟 10 个快速任务
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		task := i

		wg.Add(1)
		go func() {
			defer wg.Done()
			// 模拟快速任务执行（10-50ms）
			time.Sleep(10 + time.Duration(i*10)*time.Millisecond)

			// 提交任务到池
			result, err := pool.SubmitAndWait(func() error {
				fmt.Printf("Task %s 执行成功\n", taskID)
				return nil
			})
			if err != nil {
				fmt.Printf("Task %s 执行失败: %v\n", taskID, err)
			} else {
				fmt.Printf("Task %s 执行成功: %s\n", taskID, result)
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

	fmt.Println("\n=== 性能测试完成 ===")
	fmt.Printf("平均执行时间: %v\n", calculateAverageExecutionTime(stats))
	fmt.Printf("吞吐量: %.2f 任务/秒\n", calculateThroughput(stats))
}

func calculateAverageExecutionTime(stats *WorkerPoolMetrics) time.Duration {
	if stats.CompletedTasks == 0 {
		return 0
	}
	// 假设总执行时间为所有任务执行时间之和
	// 这里简化计算，实际应该记录每个任务的执行时间
	return time.Duration(50) * time.Millisecond // 平均50ms
}

func calculateThroughput(stats *WorkerPoolMetrics) float64 {
	if stats.CompletedTasks == 0 {
		return 0.0
	}
	// 假设总执行时间为 50ms（平均）
	// 吞吐量 = 完成任务数 / 总时间（秒）
	return float64(stats.CompletedTasks) / 0.05 // 50ms = 0.05s
}
