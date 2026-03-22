// Agent Framework - Schedule Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// scheduleCmd represents the schedule command
var scheduleCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"scheduler", "sched"},
	Short:   "管理调度任务",
	Long: `管理定时调度任务的注册、查询和控制。
调度器支持基于 cron 表达式的周期性任务。

注意：需要在配置中启用 scheduler 功能。`,
}

// addScheduleCommands adds schedule-related commands to root command
func addScheduleCommands() {
	// ── status ───────────────────────────────────────────────────────────────
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "查看调度器状态",
		Long:  `显示调度器的当前状态和基本信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sched := app.GetHost().Scheduler()
			if sched == nil {
				return fmt.Errorf("scheduler is not configured or not enabled.\nAdd scheduler config to your host.yaml:\n  scheduler:\n    enabled: true\n    maxJobs: 100")
			}

			fmt.Println("Scheduler Status:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Status:  Active\n")
			fmt.Printf("Type:    %T\n", sched)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	scheduleCmd.AddCommand(statusCmd)

	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "显示调度器配置信息",
		Long:  `显示调度器的配置参数。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			if cfg.Scheduler == nil {
				fmt.Println("Scheduler is not configured.")
				fmt.Println("To enable it, add to your config:")
				fmt.Println("  scheduler:")
				fmt.Println("    enabled: true")
				fmt.Println("    timezone: Asia/Shanghai")
				fmt.Println("    maxJobs: 100")
				fmt.Println("    jobTimeout: 3600")
				return nil
			}

			fmt.Println("Scheduler Configuration:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Enabled:    %v\n", cfg.Scheduler.Enabled)
			fmt.Printf("Timezone:   %s\n", cfg.Scheduler.Timezone)
			fmt.Printf("Max Jobs:   %d\n", cfg.Scheduler.MaxJobs)
			fmt.Printf("Job Timeout: %d seconds\n", cfg.Scheduler.JobTimeout)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	scheduleCmd.AddCommand(infoCmd)

	// ── submit ───────────────────────────────────────────────────────────────
	submitCmd := &cobra.Command{
		Use:   "submit [job-name] [agent-id] [task]",
		Short: "提交一个定时任务",
		Long:  `向调度器提交一个基于 agent 的定时任务。`,
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			jobName := args[0]
			agentID := args[1]
			task := args[2]

			handler := func(c interface{ Done() <-chan struct{} }) error {
				a, err := app.GetHost().GetAgent(agentID)
				if err != nil {
					return fmt.Errorf("agent '%s' not found: %w", agentID, err)
				}
				_, err = a.Run(ctx, task)
				return err
			}
			_ = handler

			jobID, err := app.GetHost().ScheduleJob(ctx, jobName, func(jobCtx interface{}) error {
				a, err := app.GetHost().GetAgent(agentID)
				if err != nil {
					return fmt.Errorf("agent '%s' not found: %w", agentID, err)
				}
				_, err = a.Run(ctx, task)
				return err
			})
			if err != nil {
				return fmt.Errorf("failed to schedule job: %w", err)
			}

			fmt.Printf("✓ Job scheduled\n  Job ID:   %s\n  Job Name: %s\n  Agent:    %s\n", jobID, jobName, agentID)
			return nil
		},
	}
	scheduleCmd.AddCommand(submitCmd)

	// ── heartbeat ────────────────────────────────────────────────────────────
	heartbeatCmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "发送心跳信号",
		Long:  `手动向心跳服务发送一个心跳信号（用于测试）。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			hb := app.GetHost().Heartbeat()
			if hb == nil {
				return fmt.Errorf("heartbeat service is not configured.\nAdd heartbeat config to your host.yaml:\n  heartbeat:\n    enabled: true\n    interval: 30\n    timeout: 10")
			}

			if err := app.GetHost().SendHeartbeat(ctx); err != nil {
				return fmt.Errorf("failed to send heartbeat: %w", err)
			}

			fmt.Println("✓ Heartbeat sent successfully")
			return nil
		},
	}
	scheduleCmd.AddCommand(heartbeatCmd)

	// ── heartbeat-info ────────────────────────────────────────────────────────
	hbInfoCmd := &cobra.Command{
		Use:   "heartbeat-info",
		Short: "显示心跳服务配置",
		Long:  `显示心跳服务的配置参数。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			if cfg.Heartbeat == nil {
				fmt.Println("Heartbeat service is not configured.")
				return nil
			}

			fmt.Println("Heartbeat Configuration:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Enabled:  %v\n", cfg.Heartbeat.Enabled)
			fmt.Printf("Interval: %d seconds\n", cfg.Heartbeat.Interval)
			fmt.Printf("Timeout:  %d seconds\n", cfg.Heartbeat.Timeout)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	scheduleCmd.AddCommand(hbInfoCmd)

	rootCmd.AddCommand(scheduleCmd)
}
