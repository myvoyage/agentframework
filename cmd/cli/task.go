// Agent Framework - Task Commands (Async Task Management)
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "管理异步任务",
	Long: `管理异步任务的提交、查询、取消和等待。
异步任务允许在后台执行长时间运行的操作。

注意：需要在配置中启用 asyncTask 功能。`,
}

// addTaskCommands adds async task commands to root command
func addTaskCommands() {
	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有异步任务",
		Long:  `列出当前所有异步任务及其状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := app.GetHost().ListAsyncTasks()
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(tasks, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found")
				return nil
			}

			fmt.Println("Async Tasks:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, t := range tasks {
				fmt.Printf("  %-20s  status=%-10s  created=%v\n",
					t.ID(), t.Status(), t.CreatedAt())
			}
			fmt.Printf("Total: %d task(s)\n", len(tasks))
			return nil
		},
	}
	taskCmd.AddCommand(listCmd)

	// ── get ──────────────────────────────────────────────────────────────────
	getCmd := &cobra.Command{
		Use:   "get [task-id]",
		Short: "获取任务详情",
		Long:  `获取指定异步任务的详细状态和结果。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := app.GetHost().GetAsyncTask(args[0])
			if err != nil {
				return fmt.Errorf("failed to get task '%s': %w", args[0], err)
			}

			if outputFormat == "json" {
				info := map[string]interface{}{
					"id":         task.ID(),
					"status":     task.Status(),
					"created_at": task.CreatedAt(),
				}
				b, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Task Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("ID:         %s\n", task.ID())
			fmt.Printf("Status:     %s\n", task.Status())
			fmt.Printf("Created At: %v\n", task.CreatedAt())
			if task.Error() != nil {
				fmt.Printf("Error:      %v\n", task.Error())
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	taskCmd.AddCommand(getCmd)

	// ── cancel ───────────────────────────────────────────────────────────────
	cancelCmd := &cobra.Command{
		Use:   "cancel [task-id]",
		Short: "取消异步任务",
		Long:  `取消指定的异步任务（如果任务尚未完成）。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, taskID := range args {
				if err := app.GetHost().CancelAsyncTask(taskID); err != nil {
					return fmt.Errorf("failed to cancel task '%s': %w", taskID, err)
				}
				fmt.Printf("✓ Task '%s' cancelled\n", taskID)
			}
			return nil
		},
	}
	taskCmd.AddCommand(cancelCmd)

	// ── wait ─────────────────────────────────────────────────────────────────
	waitCmd := &cobra.Command{
		Use:   "wait [task-id]",
		Short: "等待任务完成",
		Long:  `阻塞等待指定的异步任务完成，并显示结果。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			fmt.Printf("Waiting for task '%s'...\n", args[0])

			result, err := app.GetHost().WaitAsyncTask(ctx, args[0])
			if err != nil {
				return fmt.Errorf("task '%s' failed: %w", args[0], err)
			}

			fmt.Printf("Task completed.\nResult: %v\n", result)
			return nil
		},
	}
	taskCmd.AddCommand(waitCmd)

	// ── stats ────────────────────────────────────────────────────────────────
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "查看任务统计信息",
		Long:  `显示异步任务管理器的统计信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := app.GetHost().GetTaskStats()
			if err != nil {
				return fmt.Errorf("failed to get task stats: %w", err)
			}

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(stats, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Task Manager Statistics:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Total Tasks:     %d\n", stats.TotalTasks)
			fmt.Printf("Pending:         %d\n", stats.PendingTasks)
			fmt.Printf("Running:         %d\n", stats.RunningTasks)
			fmt.Printf("Completed:       %d\n", stats.CompletedTasks)
			fmt.Printf("Failed:          %d\n", stats.FailedTasks)
			fmt.Printf("Cancelled:       %d\n", stats.CancelledTasks)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	taskCmd.AddCommand(statsCmd)

	rootCmd.AddCommand(taskCmd)
}
