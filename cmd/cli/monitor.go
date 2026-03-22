// Agent Framework - Monitor Commands (System Monitoring)
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:     "monitor",
	Aliases: []string{"mon", "metrics"},
	Short:   "系统监控与指标",
	Long: `查看 AgentFramework 的系统监控信息、指标和告警。
监控系统由 MonitorManager 驱动，支持内存、CPU、自定义指标等多种监控类型。

示例:
  af monitor status              # 查看监控系统状态
  af monitor list                # 列出所有监控器
  af monitor metrics             # 查看所有当前指标
  af monitor metrics memory      # 查看指定监控器的指标
  af monitor alerts              # 查看活跃告警规则
  af monitor start               # 启动所有监控器
  af monitor stop                # 停止所有监控器`,
}

// addMonitorCommands adds monitoring-related commands to root command
func addMonitorCommands() {
	// ── status ───────────────────────────────────────────────────────────────
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "查看监控系统整体状态",
		Long:  `显示 MonitorManager 的整体状态，包括监控器数量、是否运行等。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			monMgr := app.GetHost().MonitorManager()
			monitors := monMgr.GetMonitors()

			running := 0
			for _, m := range monitors {
				if m.IsRunning() {
					running++
				}
			}

			fmt.Println("Monitor Manager Status:")
			fmt.Println("────────────────────────────────────────────────────────────")
			mgStatus := "Active"
			if !monMgr.IsRunning() {
				mgStatus = "Stopped"
			}
			fmt.Printf("Manager Status:   %s\n", mgStatus)
			fmt.Printf("Total Monitors:   %d\n", len(monitors))
			fmt.Printf("Running:          %d\n", running)
			fmt.Printf("Stopped:          %d\n", len(monitors)-running)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	monitorCmd.AddCommand(statusCmd)

	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有监控器",
		Long:  `列出当前注册的所有监控器及其状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			monitors := app.GetHost().MonitorManager().GetMonitors()

			if outputFormat == "json" {
				type monJSON struct {
					Name    string `json:"name"`
					Running bool   `json:"running"`
				}
				list := make([]monJSON, 0, len(monitors))
				for name, m := range monitors {
					list = append(list, monJSON{Name: name, Running: m.IsRunning()})
				}
				b, _ := json.MarshalIndent(list, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(monitors) == 0 {
				fmt.Println("No monitors registered")
				return nil
			}

			fmt.Println("Monitors:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for name, m := range monitors {
				status := "stopped"
				if m.IsRunning() {
					status = "running"
				}
				fmt.Printf("  %-20s  %s\n", name, status)
			}
			fmt.Printf("Total: %d monitor(s)\n", len(monitors))
			return nil
		},
	}
	monitorCmd.AddCommand(listCmd)

	// ── metrics ──────────────────────────────────────────────────────────────
	metricsCmd := &cobra.Command{
		Use:   "metrics [monitor-name]",
		Short: "查看监控指标",
		Long:  `显示当前监控指标。不指定监控器名称则显示所有指标。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			monMgr := app.GetHost().MonitorManager()

			// If specific monitor requested
			if len(args) == 1 {
				monitors := monMgr.GetMonitors()
				m, ok := monitors[args[0]]
				if !ok {
					return fmt.Errorf("monitor '%s' not found", args[0])
				}
				metrics := m.GetMetrics()
				if outputFormat == "json" {
					b, _ := json.MarshalIndent(metrics, "", "  ")
					fmt.Println(string(b))
					return nil
				}
				fmt.Printf("Monitor: %s\n", args[0])
				fmt.Println("────────────────────────────────────────────────────────────")
				if len(metrics) == 0 {
					fmt.Println("  (no metrics available)")
				}
				for _, metric := range metrics {
					fmt.Printf("  %-25s  type=%-10s  value=%v\n",
						metric.Name, metric.Type, metric.Value)
				}
				return nil
			}

			// All monitors
			allMetrics := monMgr.GetMetrics()

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(allMetrics, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(allMetrics) == 0 {
				fmt.Println("No metrics available")
				return nil
			}

			fmt.Println("All Metrics:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, metric := range allMetrics {
				fmt.Printf("  %-25s  type=%-10s  value=%v\n",
					metric.Name, metric.Type, metric.Value)
			}
			fmt.Printf("Total: %d metric(s)\n", len(allMetrics))
			return nil
		},
	}
	monitorCmd.AddCommand(metricsCmd)

	// ── stats ────────────────────────────────────────────────────────────────
	statsCmd := &cobra.Command{
		Use:   "stats [monitor-name]",
		Short: "查看监控器统计数据",
		Long:  `获取指定监控器的原始统计数据（监控器自有格式）。不指定则显示所有。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			monitors := app.GetHost().MonitorManager().GetMonitors()

			targetName := ""
			if len(args) == 1 {
				targetName = args[0]
			}

			found := false
			for name, m := range monitors {
				if targetName != "" && name != targetName {
					continue
				}
				found = true
				stats := m.GetStats()
				fmt.Printf("\nMonitor: %s\n", name)
				fmt.Println("────────────────────────────────────────────────────────────")
				if outputFormat == "json" {
					b, _ := json.MarshalIndent(stats, "", "  ")
					fmt.Println(string(b))
				} else {
					fmt.Printf("%v\n", stats)
				}
			}

			if targetName != "" && !found {
				return fmt.Errorf("monitor '%s' not found", targetName)
			}
			return nil
		},
	}
	monitorCmd.AddCommand(statsCmd)

	// ── start ────────────────────────────────────────────────────────────────
	startCmd := &cobra.Command{
		Use:   "start [monitor-name]",
		Short: "启动监控器",
		Long:  `启动指定的监控器（不指定则启动全部）。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			monMgr := app.GetHost().MonitorManager()

			if len(args) == 0 {
				// Start all
				monMgr.Start()
				fmt.Println("✓ All monitors started")
				return nil
			}

			monitors := monMgr.GetMonitors()
			m, ok := monitors[args[0]]
			if !ok {
				return fmt.Errorf("monitor '%s' not found", args[0])
			}

			if m.IsRunning() {
				fmt.Printf("Monitor '%s' is already running\n", args[0])
			} else {
				m.Start()
				fmt.Printf("✓ Monitor '%s' started\n", args[0])
			}
			return nil
		},
	}
	monitorCmd.AddCommand(startCmd)

	// ── stop ─────────────────────────────────────────────────────────────────
	stopCmd := &cobra.Command{
		Use:   "stop [monitor-name]",
		Short: "停止监控器",
		Long:  `停止指定的监控器（不指定则停止全部）。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			monMgr := app.GetHost().MonitorManager()

			if len(args) == 0 {
				// Stop all
				monMgr.Stop()
				fmt.Println("✓ All monitors stopped")
				return nil
			}

			monitors := monMgr.GetMonitors()
			m, ok := monitors[args[0]]
			if !ok {
				return fmt.Errorf("monitor '%s' not found", args[0])
			}

			if !m.IsRunning() {
				fmt.Printf("Monitor '%s' is already stopped\n", args[0])
			} else {
				m.Stop()
				fmt.Printf("✓ Monitor '%s' stopped\n", args[0])
			}
			return nil
		},
	}
	monitorCmd.AddCommand(stopCmd)

	// ── alerts ───────────────────────────────────────────────────────────────
	alertsCmd := &cobra.Command{
		Use:   "alerts [monitor-name]",
		Short: "查看告警规则",
		Long:  `显示指定监控器（或全部）的告警规则列表。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			monitors := app.GetHost().MonitorManager().GetMonitors()

			targetName := ""
			if len(args) == 1 {
				targetName = args[0]
			}

			totalRules := 0
			fmt.Println("Alert Rules:")
			fmt.Println("────────────────────────────────────────────────────────────")

			for name, m := range monitors {
				if targetName != "" && name != targetName {
					continue
				}
				rules := m.GetAlertRules()
				if len(rules) == 0 {
					continue
				}
				fmt.Printf("Monitor: %s\n", name)
				for _, rule := range rules {
					status := "disabled"
					if rule.Enabled {
						status = "enabled"
					}
					fmt.Printf("  %-20s  severity=%-8s  metric=%-20s  status=%s\n",
						rule.Name, rule.Severity, rule.MetricName, status)
					totalRules++
				}
			}

			if targetName != "" {
				_, ok := monitors[targetName]
				if !ok {
					return fmt.Errorf("monitor '%s' not found", targetName)
				}
			}

			if totalRules == 0 {
				fmt.Println("No alert rules configured")
			} else {
				fmt.Printf("Total: %d rule(s)\n", totalRules)
			}
			return nil
		},
	}
	monitorCmd.AddCommand(alertsCmd)

	rootCmd.AddCommand(monitorCmd)
}
