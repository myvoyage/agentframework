// Agent Framework - Enhanced Schedule Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"AgentFramework/agent/scheduler"
)

var (
	_cronExpr string
	_taskType string
	_prompt   string
	_message  string
	_model    string
	_tools    []string
	_timezone string
	_enabled  bool
)

// schedGetScheduler returns the Scheduler interface from the application host.
// Returns nil if the scheduler is not configured or the app is not initialized.
func schedGetScheduler() scheduler.Scheduler {
	if app == nil {
		return nil
	}
	raw := app.GetHost().Scheduler()
	if raw == nil {
		return nil
	}
	if sched, ok := raw.(scheduler.Scheduler); ok {
		return sched
	}
	return nil
}

// addEnhancedScheduleCommands registers all schedule sub-commands onto rootCmd.
func addEnhancedScheduleCommands() {
	scheduleCmd := &cobra.Command{
		Use:     "schedule",
		Aliases: []string{"cron", "sched"},
		Short:   "管理定时调度任务",
		Long: `管理基于 cron 表达式的定时任务。

支持的任务类型:
  - static: 静态消息任务（定时发送固定内容）
  - ai:     AI 智能任务（定时执行 AI 提示词并输出结果）

Cron 表达式格式 (标准 5 字段):
  *    *    *    *    *
  分   时   日   月   周
  ─────────────────────────────────────
  0 9 * * *       每天 09:00
  0 9 * * 1       每周一 09:00
  0 9 1 * *       每月 1 号 09:00
  */30 * * * *    每 30 分钟
  0 */6 * * *     每 6 小时

示例:
  af schedule list
  af schedule add "每日日报" "0 18 * * *" --type static --message "今日工作总结"
  af schedule add "AI摘要"  "0 9 * * *"  --type ai --prompt "生成每日新闻摘要" --model gpt-4o
  af schedule pause  <job-id>
  af schedule resume <job-id>
  af schedule run    <job-id>
  af schedule delete <job-id>
  af schedule validate "0 9 * * *"
  af schedule stats
  af schedule export jobs.json
  af schedule import jobs.json`,
	}

	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有调度任务",
		Long:  `列出调度器中所有已注册的任务，按下次运行时间升序排列。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedListJobs()
		},
	}
	scheduleCmd.AddCommand(listCmd)

	// ── add ──────────────────────────────────────────────────────────────────
	addCmd := &cobra.Command{
		Use:   "add <name> <cron-expr>",
		Short: "添加定时任务",
		Long: `添加一个新的定时任务到调度器。

任务类型:
  --type static  静态消息任务，定时发送 --message 指定的固定内容。
  --type ai      AI 智能任务，定时执行 --prompt 指定的提示词并将模型输出送往渠道。

示例:
  af schedule add "每日提醒" "0 9 * * *" --type static --message "记得喝水"
  af schedule add "AI日报"  "0 18 * * *" --type ai --prompt "生成今日工作总结" --model gpt-4o
  af schedule add "周报"    "0 9 * * 5" --type ai --prompt "生成本周工作周报" --tools web_search`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedAddJob(args[0], args[1])
		},
	}
	addCmd.Flags().StringVarP(&_taskType, "type", "t", "static", "任务类型 (static/ai)")
	addCmd.Flags().StringVar(&_prompt, "prompt", "", "AI 任务的提示词（type=ai 必须）")
	addCmd.Flags().StringVar(&_message, "message", "", "静态消息内容（type=static 必须）")
	addCmd.Flags().StringVar(&_model, "model", "default", "AI 模型名称")
	addCmd.Flags().StringSliceVar(&_tools, "tools", []string{}, "启用的工具列表（逗号分隔）")
	addCmd.Flags().StringVar(&_timezone, "timezone", "Local", "时区（如 Asia/Shanghai）")
	addCmd.Flags().BoolVar(&_enabled, "enabled", true, "创建后立即启用")
	scheduleCmd.AddCommand(addCmd)

	// ── delete ────────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <job-id>",
		Short: "删除调度任务",
		Long:  `从调度器中永久删除指定任务。操作不可逆，请先使用 pause 暂停确认后再删除。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedDeleteJob(args[0])
		},
	}
	scheduleCmd.AddCommand(deleteCmd)

	// ── pause ────────────────────────────────────────────────────────────────
	pauseCmd := &cobra.Command{
		Use:   "pause <job-id>",
		Short: "暂停调度任务",
		Long:  `暂停指定任务，任务不会被删除，可通过 resume 恢复。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedPauseJob(args[0])
		},
	}
	scheduleCmd.AddCommand(pauseCmd)

	// ── resume ────────────────────────────────────────────────────────────────
	resumeCmd := &cobra.Command{
		Use:   "resume <job-id>",
		Short: "恢复调度任务",
		Long:  `恢复已暂停的任务，任务将从下次 cron 触发时开始执行。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedResumeJob(args[0])
		},
	}
	scheduleCmd.AddCommand(resumeCmd)

	// ── run ───────────────────────────────────────────────────────────────────
	runCmd := &cobra.Command{
		Use:   "run <job-id>",
		Short: "立即触发任务",
		Long:  `立即在后台执行一次指定任务，不影响任务的正常调度周期。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedRunJobNow(args[0])
		},
	}
	scheduleCmd.AddCommand(runCmd)

	// ── show ──────────────────────────────────────────────────────────────────
	showCmd := &cobra.Command{
		Use:   "show <job-id>",
		Short: "显示任务详情",
		Long:  `显示指定任务的完整配置和运行历史统计信息。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedShowJob(args[0])
		},
	}
	scheduleCmd.AddCommand(showCmd)

	// ── validate ──────────────────────────────────────────────────────────────
	validateCmd := &cobra.Command{
		Use:   "validate <cron-expr>",
		Short: "验证 Cron 表达式",
		Long:  `验证 cron 表达式语法，并预测未来 10 次触发时间。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedValidateCron(args[0])
		},
	}
	scheduleCmd.AddCommand(validateCmd)

	// ── stats ─────────────────────────────────────────────────────────────────
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "显示调度器统计信息",
		Long:  `显示调度器的整体统计信息，包括总任务数、执行次数、失败次数、平均耗时等。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedShowStats()
		},
	}
	scheduleCmd.AddCommand(statsCmd)

	// ── export ────────────────────────────────────────────────────────────────
	exportCmd := &cobra.Command{
		Use:   "export <file>",
		Short: "导出所有任务配置",
		Long:  `将调度器中所有任务的配置序列化为 JSON 文件，可用于备份或迁移。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedExportJobs(args[0])
		},
	}
	scheduleCmd.AddCommand(exportCmd)

	// ── import ────────────────────────────────────────────────────────────────
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "从文件导入任务",
		Long: `从 JSON 文件批量导入任务到调度器。

导入文件格式示例:
[
  {
    "id": "job_001",
    "name": "每日提醒",
    "type": "static",
    "cron_expr": "0 9 * * *",
    "static_message": "记得喝水",
    "enabled": true
  },
  {
    "id": "job_002",
    "name": "AI日报",
    "type": "ai",
    "cron_expr": "0 18 * * *",
    "ai_prompt": "生成今日工作总结",
    "ai_model": "gpt-4o",
    "enabled": true
  }
]`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schedImportJobs(args[0])
		},
	}
	scheduleCmd.AddCommand(importCmd)

	rootCmd.AddCommand(scheduleCmd)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper: format job status for display
// ─────────────────────────────────────────────────────────────────────────────

func schedFormatStatus(job *scheduler.Job) string {
	switch job.Status {
	case scheduler.JobStatusRunning:
		return "running"
	case scheduler.JobStatusPaused:
		return "paused"
	case scheduler.JobStatusCompleted:
		return "completed"
	case scheduler.JobStatusFailed:
		return "failed"
	default:
		if job.Enabled {
			return "enabled"
		}
		return "disabled"
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// list
// ─────────────────────────────────────────────────────────────────────────────

func schedListJobs() error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用。请在配置文件中开启 scheduler.enabled: true 后重启服务")
	}

	ctx := rootContext()
	jobs, err := sched.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	if len(jobs) == 0 {
		fmt.Println("暂无调度任务。使用 'af schedule add' 创建第一个任务。")
		return nil
	}

	// Sort by NextRunAt ascending (disabled jobs go to the end)
	sort.Slice(jobs, func(i, j int) bool {
		ti, tj := jobs[i].NextRunAt, jobs[j].NextRunAt
		if ti.IsZero() {
			return false
		}
		if tj.IsZero() {
			return true
		}
		return ti.Before(tj)
	})

	fmt.Printf("调度任务列表  (共 %d 个)\n", len(jobs))
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  任务ID\t任务名称\t类型\tCron\t状态\t运行次数\t失败次数\t下次运行")
	fmt.Fprintln(w, "  ──────\t────────\t────\t────\t────\t────────\t────────\t────────")

	for _, job := range jobs {
		schedType := string(job.Schedule.Type)
		if schedType == "" {
			schedType = "interval"
		}

		nextRun := "—"
		if !job.NextRunAt.IsZero() && job.Enabled {
			diff := time.Until(job.NextRunAt)
			if diff > 0 {
				if diff > 24*time.Hour {
					nextRun = fmt.Sprintf("%s (%.0fd)", job.NextRunAt.Format("01-02 15:04"), diff.Hours()/24)
				} else if diff > time.Hour {
					nextRun = fmt.Sprintf("%s (%.0fh)", job.NextRunAt.Format("15:04:05"), diff.Hours())
				} else {
					nextRun = fmt.Sprintf("%s (%dm)", job.NextRunAt.Format("15:04:05"), int(diff.Minutes()))
				}
			} else {
				nextRun = job.NextRunAt.Format("2006-01-02 15:04:05")
			}
		}

		cronExpr := job.Schedule.CronExpr
		if cronExpr == "" && job.Schedule.Interval > 0 {
			cronExpr = fmt.Sprintf("每 %v", job.Schedule.Interval)
		}

		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			job.ID,
			job.Name,
			schedType,
			cronExpr,
			schedFormatStatus(job),
			job.RunCount,
			job.FailCount,
			nextRun,
		)
	}
	w.Flush()
	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// add
// ─────────────────────────────────────────────────────────────────────────────

func schedAddJob(name, cronExpr string) error {
	// Validate cron expression first
	if err := scheduler.ValidateCronExpr(cronExpr); err != nil {
		return fmt.Errorf("无效的 cron 表达式 %q: %w", cronExpr, err)
	}

	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用。请在配置文件中开启 scheduler.enabled: true 后重启服务")
	}

	ctx := rootContext()

	// Build Job depending on task type
	var job *scheduler.Job

	switch _taskType {
	case "ai":
		if _prompt == "" {
			return fmt.Errorf("--prompt 是 AI 任务的必要参数")
		}
		meta := map[string]string{
			"prompt": _prompt,
			"model":  _model,
		}
		if len(_tools) > 0 {
			meta["tools"] = strings.Join(_tools, ",")
		}
		job = &scheduler.Job{
			Name:        name,
			Description: fmt.Sprintf("AI task: %s", _prompt),
			Tags:        []string{"ai", "scheduled"},
			Enabled:     _enabled,
			Metadata:    meta,
			Schedule: scheduler.JobSchedule{
				Type:     scheduler.ScheduleTypeCron,
				CronExpr: cronExpr,
				Timezone: _timezone,
			},
			// Handler will invoke agent at runtime; here we register a no-op placeholder
			// that logs intent. Full agent integration is done via Host.ScheduleJob.
			Handler: func(ctx context.Context) error {
				fmt.Printf("[AI Task] %s | prompt: %s\n", name, _prompt)
				return nil
			},
		}

	case "static":
		if _message == "" {
			return fmt.Errorf("--message 是静态任务的必要参数")
		}
		job = &scheduler.Job{
			Name:        name,
			Description: fmt.Sprintf("Static message: %s", _message),
			Tags:        []string{"static", "scheduled"},
			Enabled:     _enabled,
			Metadata:    map[string]string{"message": _message},
			Schedule: scheduler.JobSchedule{
				Type:     scheduler.ScheduleTypeCron,
				CronExpr: cronExpr,
				Timezone: _timezone,
			},
			Handler: func(ctx context.Context) error {
				fmt.Printf("[Static Task] %s | message: %s\n", name, _message)
				return nil
			},
		}

	default:
		return fmt.Errorf("未知任务类型 %q，支持: static / ai", _taskType)
	}

	jobID, err := sched.ScheduleJob(ctx, job)
	if err != nil {
		return fmt.Errorf("添加任务失败: %w", err)
	}

	// Fetch the stored job to get NextRunAt
	stored, err := sched.GetJob(ctx, jobID)
	if err != nil {
		// Non-fatal: fallback display
		stored = job
		stored.ID = jobID
	}

	// Compute next run from cron expression
	cronSched := scheduler.NewCronScheduler(_timezone)
	nextRun, _ := cronSched.GetNextRunTime(cronExpr)

	fmt.Println()
	fmt.Println("✓ 任务添加成功")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("  任务ID:    %s\n", stored.ID)
	fmt.Printf("  名称:      %s\n", stored.Name)
	fmt.Printf("  类型:      %s\n", _taskType)
	fmt.Printf("  Cron:      %s\n", cronExpr)
	fmt.Printf("  时区:      %s\n", _timezone)
	if _taskType == "ai" {
		fmt.Printf("  提示词:    %s\n", _prompt)
		fmt.Printf("  模型:      %s\n", _model)
		if len(_tools) > 0 {
			fmt.Printf("  工具:      %s\n", strings.Join(_tools, ", "))
		}
	} else {
		fmt.Printf("  消息:      %s\n", _message)
	}
	fmt.Printf("  已启用:    %v\n", _enabled)
	if nextRun.Year() > 2000 {
		fmt.Printf("  下次运行:  %s (距今 %v)\n",
			nextRun.Format("2006-01-02 15:04:05 MST"),
			time.Until(nextRun).Round(time.Second))
	}
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n提示: 使用 'af schedule run %s' 立即触发此任务\n", stored.ID)
	fmt.Println()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// delete
// ─────────────────────────────────────────────────────────────────────────────

func schedDeleteJob(jobID string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	// Verify job exists and show summary before asking for confirmation
	job, err := sched.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在: %s", jobID)
	}

	fmt.Printf("即将删除任务:\n  ID:   %s\n  名称: %s\n  类型: %s\n", job.ID, job.Name, job.Schedule.Type)
	fmt.Printf("确认删除? [yes/NO]: ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
		fmt.Println("已取消")
		return nil
	}

	if err := sched.UnscheduleJob(ctx, jobID); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	fmt.Printf("✓ 任务已删除: %s (%s)\n", job.Name, jobID)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// pause
// ─────────────────────────────────────────────────────────────────────────────

func schedPauseJob(jobID string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	job, err := sched.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在: %s", jobID)
	}

	if job.Status == scheduler.JobStatusPaused || !job.Enabled {
		fmt.Printf("任务 %s (%s) 已经处于暂停状态\n", job.Name, jobID)
		return nil
	}

	if err := sched.PauseJob(ctx, jobID); err != nil {
		return fmt.Errorf("暂停任务失败: %w", err)
	}

	fmt.Printf("✓ 任务已暂停: %s (%s)\n", job.Name, jobID)
	fmt.Println("  使用 'af schedule resume " + jobID + "' 恢复任务")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// resume
// ─────────────────────────────────────────────────────────────────────────────

func schedResumeJob(jobID string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	job, err := sched.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在: %s", jobID)
	}

	if job.Status != scheduler.JobStatusPaused && job.Enabled {
		fmt.Printf("任务 %s (%s) 当前状态: %s，无需恢复\n", job.Name, jobID, schedFormatStatus(job))
		return nil
	}

	if err := sched.ResumeJob(ctx, jobID); err != nil {
		return fmt.Errorf("恢复任务失败: %w", err)
	}

	// Show next run time if cron expression is available
	nextRunInfo := ""
	if job.Schedule.CronExpr != "" {
		cronSched := scheduler.NewCronScheduler(job.Schedule.Timezone)
		if nextRun, err := cronSched.GetNextRunTime(job.Schedule.CronExpr); err == nil {
			nextRunInfo = fmt.Sprintf("，下次运行: %s", nextRun.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Printf("✓ 任务已恢复: %s (%s)%s\n", job.Name, jobID, nextRunInfo)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// run (immediate trigger)
// ─────────────────────────────────────────────────────────────────────────────

func schedRunJobNow(jobID string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	job, err := sched.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在: %s", jobID)
	}

	if job.Handler == nil {
		return fmt.Errorf("任务 %s 没有注册处理器，无法立即执行", jobID)
	}

	fmt.Printf("立即执行任务: %s (%s)...\n", job.Name, jobID)
	startTime := time.Now()

	// Execute the job handler in a goroutine with a timeout context
	timeout := 10 * time.Minute
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- job.Handler(runCtx)
	}()

	select {
	case execErr := <-errCh:
		duration := time.Since(startTime).Round(time.Millisecond)
		if execErr != nil {
			return fmt.Errorf("任务执行失败 (耗时 %v): %w", duration, execErr)
		}
		fmt.Printf("✓ 任务执行完成，耗时 %v\n", duration)

	case <-runCtx.Done():
		return fmt.Errorf("任务执行超时 (超过 %v)", timeout)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// show
// ─────────────────────────────────────────────────────────────────────────────

func schedShowJob(jobID string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	job, err := sched.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("任务不存在: %s", jobID)
	}

	fmt.Println()
	fmt.Printf("任务详情: %s\n", job.Name)
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("  ID:          %s\n", job.ID)
	fmt.Printf("  名称:        %s\n", job.Name)
	if job.Description != "" {
		fmt.Printf("  描述:        %s\n", job.Description)
	}
	if len(job.Tags) > 0 {
		fmt.Printf("  标签:        %s\n", strings.Join(job.Tags, ", "))
	}
	fmt.Printf("  状态:        %s\n", schedFormatStatus(job))
	fmt.Printf("  已启用:      %v\n", job.Enabled)
	fmt.Println()

	// Schedule details
	fmt.Println("  ── 调度配置 ──────────────────────────────────────────")
	fmt.Printf("  类型:        %s\n", job.Schedule.Type)
	if job.Schedule.CronExpr != "" {
		fmt.Printf("  Cron:        %s\n", job.Schedule.CronExpr)
		desc := scheduler.GetCronDescription(job.Schedule.CronExpr)
		fmt.Printf("  描述:        %s\n", desc)
	}
	if job.Schedule.Interval > 0 {
		fmt.Printf("  间隔:        %v\n", job.Schedule.Interval)
	}
	if job.Schedule.Timezone != "" {
		fmt.Printf("  时区:        %s\n", job.Schedule.Timezone)
	}
	if job.Schedule.MaxRuns > 0 {
		fmt.Printf("  最大执行次数: %d\n", job.Schedule.MaxRuns)
	}
	fmt.Println()

	// Metadata
	if len(job.Metadata) > 0 {
		fmt.Println("  ── 任务参数 ──────────────────────────────────────────")
		keys := make([]string, 0, len(job.Metadata))
		for k := range job.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %-12s %s\n", k+":", job.Metadata[k])
		}
		fmt.Println()
	}

	// Runtime stats
	fmt.Println("  ── 运行统计 ──────────────────────────────────────────")
	fmt.Printf("  执行次数:    %d\n", job.RunCount)
	fmt.Printf("  失败次数:    %d\n", job.FailCount)
	if !job.CreatedAt.IsZero() {
		fmt.Printf("  创建时间:    %s\n", job.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if !job.LastRunAt.IsZero() {
		fmt.Printf("  上次运行:    %s\n", job.LastRunAt.Format("2006-01-02 15:04:05"))
	}
	if !job.NextRunAt.IsZero() && job.Enabled {
		fmt.Printf("  下次运行:    %s (距今 %v)\n",
			job.NextRunAt.Format("2006-01-02 15:04:05"),
			time.Until(job.NextRunAt).Round(time.Second))
	}
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Println()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// validate
// ─────────────────────────────────────────────────────────────────────────────

func schedValidateCron(cronExpr string) error {
	fmt.Printf("验证 cron 表达式: %q\n\n", cronExpr)

	if err := scheduler.ValidateCronExpr(cronExpr); err != nil {
		fmt.Printf("✗ 无效: %v\n\n", err)
		fmt.Println("常见格式示例:")
		fmt.Println("  0 9 * * *      每天 09:00")
		fmt.Println("  0 9 * * 1      每周一 09:00")
		fmt.Println("  0 9 1 * *      每月 1 号 09:00")
		fmt.Println("  */30 * * * *   每 30 分钟")
		fmt.Println("  0 */6 * * *    每 6 小时整")
		return nil
	}

	fmt.Printf("✓ 有效\n\n")

	// Human-readable description
	desc := scheduler.GetCronDescription(cronExpr)
	fmt.Printf("描述: %s\n\n", desc)

	// Compute next 10 trigger times using GetNextRunTimeAfter to correctly
	// enumerate consecutive trigger instants by advancing the reference point.
	tz := "Local"
	cronSched := scheduler.NewCronScheduler(tz)

	// Obtain the very first trigger after now.
	firstRun, err := cronSched.GetNextRunTime(cronExpr)
	if err != nil {
		return fmt.Errorf("无法计算下次运行时间: %w", err)
	}

	// Build a slice of 10 consecutive trigger times.
	triggers := make([]time.Time, 0, 10)
	ref := firstRun
	for len(triggers) < 10 {
		triggers = append(triggers, ref)
		nextRef, nextErr := cronSched.GetNextRunTimeAfter(cronExpr, ref)
		if nextErr != nil {
			// Expression parsed once already; a second failure is unexpected but we stop gracefully.
			break
		}
		ref = nextRef
	}

	fmt.Println("未来 10 次触发时间:")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	for i, t := range triggers {
		diff := time.Until(t)
		var diffStr string
		switch {
		case diff > 24*time.Hour:
			diffStr = fmt.Sprintf("%.1f 天后", diff.Hours()/24)
		case diff > time.Hour:
			diffStr = fmt.Sprintf("%.1f 小时后", diff.Hours())
		case diff > time.Minute:
			diffStr = fmt.Sprintf("%d 分钟后", int(diff.Minutes()))
		default:
			diffStr = fmt.Sprintf("%d 秒后", int(diff.Seconds()))
		}
		fmt.Printf("  %2d. %s  (%s)\n", i+1, t.Format("2006-01-02 Mon 15:04:05 MST"), diffStr)
	}
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	fmt.Println()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// stats
// ─────────────────────────────────────────────────────────────────────────────

func schedShowStats() error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用。请在配置文件中开启 scheduler.enabled: true 后重启服务")
	}

	ctx := rootContext()

	stats, err := sched.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("获取统计信息失败: %w", err)
	}

	// Also fetch job list for per-type breakdown
	jobs, listErr := sched.ListJobs(ctx)

	fmt.Println()
	fmt.Println("调度器统计信息")
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  指标\t数值")
	fmt.Fprintln(w, "  ────\t────")
	fmt.Fprintf(w, "  总任务数\t%d\n", stats.TotalJobs)
	fmt.Fprintf(w, "  活跃任务\t%d\n", stats.ActiveJobs)
	fmt.Fprintf(w, "  已暂停\t%d\n", stats.PausedJobs)
	fmt.Fprintf(w, "  已完成\t%d\n", stats.CompletedJobs)
	fmt.Fprintf(w, "  已失败\t%d\n", stats.FailedJobs)
	fmt.Fprintf(w, "  总执行次数\t%d\n", stats.TotalRuns)
	fmt.Fprintf(w, "  总失败次数\t%d\n", stats.TotalFailures)
	if stats.AvgDuration > 0 {
		fmt.Fprintf(w, "  平均耗时\t%v\n", stats.AvgDuration.Round(time.Millisecond))
	}
	if stats.Uptime > 0 {
		fmt.Fprintf(w, "  运行时长\t%v\n", stats.Uptime.Round(time.Second))
	}
	w.Flush()

	// Per-type breakdown (if job list is available)
	if listErr == nil && len(jobs) > 0 {
		typeCount := make(map[string]int)
		for _, job := range jobs {
			typeCount[string(job.Schedule.Type)]++
		}

		fmt.Println()
		fmt.Println("  任务类型分布:")
		for t, count := range typeCount {
			bar := strings.Repeat("█", count)
			fmt.Printf("    %-10s %s %d\n", t, bar, count)
		}
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// export
// ─────────────────────────────────────────────────────────────────────────────

// schedJobExport is the JSON-serializable representation of a scheduled job.
type schedJobExport struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Type         string            `json:"type"`
	CronExpr     string            `json:"cron_expr,omitempty"`
	Interval     string            `json:"interval,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	StaticMsg    string            `json:"static_message,omitempty"`
	AIPrompt     string            `json:"ai_prompt,omitempty"`
	AIModel      string            `json:"ai_model,omitempty"`
	AITools      []string          `json:"ai_tools,omitempty"`
	Enabled      bool              `json:"enabled"`
	RunCount     int64             `json:"run_count"`
	FailCount    int64             `json:"fail_count"`
	CreatedAt    time.Time         `json:"created_at"`
	LastRunAt    *time.Time        `json:"last_run_at,omitempty"`
	NextRunAt    *time.Time        `json:"next_run_at,omitempty"`
}

func schedExportJobs(filePath string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	ctx := rootContext()

	jobs, err := sched.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}
	if len(jobs) == 0 {
		fmt.Println("暂无任务可导出")
		return nil
	}

	exports := make([]schedJobExport, 0, len(jobs))
	for _, job := range jobs {
		ex := schedJobExport{
			ID:          job.ID,
			Name:        job.Name,
			Description: job.Description,
			Tags:        job.Tags,
			Type:        string(job.Schedule.Type),
			CronExpr:    job.Schedule.CronExpr,
			Timezone:    job.Schedule.Timezone,
			Metadata:    job.Metadata,
			Enabled:     job.Enabled,
			RunCount:    job.RunCount,
			FailCount:   job.FailCount,
			CreatedAt:   job.CreatedAt,
		}
		if job.Schedule.Interval > 0 {
			ex.Interval = job.Schedule.Interval.String()
		}
		if !job.LastRunAt.IsZero() {
			t := job.LastRunAt
			ex.LastRunAt = &t
		}
		if !job.NextRunAt.IsZero() {
			t := job.NextRunAt
			ex.NextRunAt = &t
		}
		// Extract well-known metadata fields
		if job.Metadata != nil {
			ex.StaticMsg = job.Metadata["message"]
			ex.AIPrompt = job.Metadata["prompt"]
			ex.AIModel = job.Metadata["model"]
			if toolsStr := job.Metadata["tools"]; toolsStr != "" {
				ex.AITools = strings.Split(toolsStr, ",")
			}
		}
		exports = append(exports, ex)
	}

	// Sort by creation time
	sort.Slice(exports, func(i, j int) bool {
		return exports[i].CreatedAt.Before(exports[j].CreatedAt)
	})

	data, err := json.MarshalIndent(exports, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("✓ 已导出 %d 个任务到: %s\n", len(exports), filePath)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// import
// ─────────────────────────────────────────────────────────────────────────────

func schedImportJobs(filePath string) error {
	sched := schedGetScheduler()
	if sched == nil {
		return fmt.Errorf("调度器未启用")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var imports []schedJobExport
	if err := json.Unmarshal(data, &imports); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}
	if len(imports) == 0 {
		return fmt.Errorf("文件中没有找到任务")
	}

	ctx := rootContext()

	var (
		successCount int
		skipCount    int
		failCount    int
	)

	for _, imp := range imports {
		// Validate cron expression
		if imp.CronExpr != "" {
			if err := scheduler.ValidateCronExpr(imp.CronExpr); err != nil {
				fmt.Printf("  ✗ 跳过 %s: 无效的 cron 表达式 %q: %v\n", imp.Name, imp.CronExpr, err)
				skipCount++
				continue
			}
		}

		// Build metadata
		meta := imp.Metadata
		if meta == nil {
			meta = make(map[string]string)
		}
		if imp.StaticMsg != "" {
			meta["message"] = imp.StaticMsg
		}
		if imp.AIPrompt != "" {
			meta["prompt"] = imp.AIPrompt
		}
		if imp.AIModel != "" {
			meta["model"] = imp.AIModel
		}
		if len(imp.AITools) > 0 {
			meta["tools"] = strings.Join(imp.AITools, ",")
		}

		// Determine interval
		var interval time.Duration
		if imp.Interval != "" {
			if d, parseErr := time.ParseDuration(imp.Interval); parseErr == nil {
				interval = d
			}
		}

		job := &scheduler.Job{
			ID:          imp.ID,
			Name:        imp.Name,
			Description: imp.Description,
			Tags:        imp.Tags,
			Enabled:     imp.Enabled,
			Metadata:    meta,
			Schedule: scheduler.JobSchedule{
				Type:     scheduler.ScheduleType(imp.Type),
				CronExpr: imp.CronExpr,
				Interval: interval,
				Timezone: imp.Timezone,
			},
			Handler: func(ctx context.Context) error {
				fmt.Printf("[Imported Task] %s\n", imp.Name)
				return nil
			},
		}

		if _, err := sched.ScheduleJob(ctx, job); err != nil {
			fmt.Printf("  ✗ 导入失败 %s: %v\n", imp.Name, err)
			failCount++
		} else {
			fmt.Printf("  ✓ 导入成功: %s\n", imp.Name)
			successCount++
		}
	}

	fmt.Println()
	fmt.Printf("导入结果: 成功 %d / 跳过 %d / 失败 %d (共 %d 个)\n",
		successCount, skipCount, failCount, len(imports))
	return nil
}
