package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	_logFile     string
	_logLines    int
	_logLevel    string
	_logSource   string
	_logSince    string
	_logTailOnly bool
)

func addLogsCommands() {
	// ── logs ──────────────────────────────────────────────────────────────
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "查看和管理日志",
		Long: `查看和管理 AgentFramework 的日志文件。

支持的日志级别:
  DEBUG   - 调试信息
  INFO    - 一般信息
  WARN    - 警告信息
  ERROR   - 错误信息

日志文件位置: <data-dir>/logs/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// ── list ──────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有日志文件",
		Long:  `列出 <data-dir>/logs/ 目录下的所有日志文件。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")

			// Check if logs directory exists
			if _, err := os.Stat(logsDir); os.IsNotExist(err) {
				fmt.Println("ℹ 日志目录不存在")
				fmt.Printf("  路径: %s\n", logsDir)
				return nil
			}

			// List log files
			entries, err := os.ReadDir(logsDir)
			if err != nil {
				return fmt.Errorf("读取日志目录失败: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("ℹ 日志目录为空")
				return nil
			}

			fmt.Println("日志文件:")
			fmt.Println("────────────────────────────────────────────────────────────")

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				info, _ := entry.Info()
				sizeMB := float64(info.Size()) / (1024 * 1024)
				modTime := info.ModTime().Format("2006-01-02 15:04:05")

				fmt.Printf("  %-30s  %8.2f MB  %s\n",
					entry.Name(),
					sizeMB,
					modTime)
			}

			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("总计: %d 个日志文件\n", len(entries))

			return nil
		},
	}
	logsCmd.AddCommand(listCmd)

	// ── view ───────────────────────────────────────────────────────────────
	viewCmd := &cobra.Command{
		Use:   "view [log-file]",
		Short: "查看日志文件内容",
		Long: `查看指定的日志文件内容。如果未指定文件，查看最新的日志文件。

支持的过滤选项:
  --lines      限制显示的行数（默认: 100）
  --level      只显示指定级别的日志（DEBUG/INFO/WARN/ERROR）
  --source     只显示指定来源的日志
  --since      只显示指定时间之后的日志（如 2026-03-22）
  --tail-only  只显示尾部（从头开始过滤后只显示最后 N 行）`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")

			// Determine log file
			var logFile string
			if len(args) == 1 {
				logFile = args[0]
			} else {
				// Find latest log file
				latest, err := findLatestLogFile(logsDir)
				if err != nil {
					return fmt.Errorf("查找最新日志文件失败: %w", err)
				}
				logFile = latest
			}

			logPath := filepath.Join(logsDir, logFile)

			// Read and filter logs
			return viewLogs(logPath, _logLines, _logLevel, _logSource, _logSince, _logTailOnly)
		},
	}

	viewCmd.Flags().IntVarP(&_logLines, "lines", "n", 100, "显示的行数")
	viewCmd.Flags().StringVarP(&_logLevel, "level", "l", "", "日志级别过滤 (DEBUG/INFO/WARN/ERROR)")
	viewCmd.Flags().StringVarP(&_logSource, "source", "s", "", "日志来源过滤")
	viewCmd.Flags().StringVarP(&_logSince, "since", "t", "", "显示指定时间之后的日志 (如 2026-03-22)")
	viewCmd.Flags().BoolVar(&_logTailOnly, "tail-only", false, "只显示尾部")

	logsCmd.AddCommand(viewCmd)

	// ── follow ─────────────────────────────────────────────────────────────
	followCmd := &cobra.Command{
		Use:   "follow [log-file]",
		Short: "实时跟踪日志",
		Long: `实时跟踪日志文件的新增内容（类似 tail -f）。

支持的过滤选项:
  --level      只显示指定级别的日志
  --source     只显示指定来源的日志

快捷键:
  Ctrl+C       停止跟踪`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")

			// Determine log file
			var logFile string
			if len(args) == 1 {
				logFile = args[0]
			} else {
				// Find latest log file
				latest, err := findLatestLogFile(logsDir)
				if err != nil {
					return fmt.Errorf("查找最新日志文件失败: %w", err)
				}
				logFile = latest
			}

			logPath := filepath.Join(logsDir, logFile)

			// Follow logs
			fmt.Printf("[Follow] 实时跟踪日志: %s\n", logFile)
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println("按 Ctrl+C 停止跟踪")
			fmt.Println()

			return followLogs(logPath, _logLevel, _logSource)
		},
	}

	followCmd.Flags().StringVarP(&_logLevel, "level", "l", "", "日志级别过滤 (DEBUG/INFO/WARN/ERROR)")
	followCmd.Flags().StringVarP(&_logSource, "source", "s", "", "日志来源过滤")

	logsCmd.AddCommand(followCmd)

	// ── clear ──────────────────────────────────────────────────────────────
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "清空日志文件",
		Long:  `清空指定的日志文件或所有日志文件。此操作不可逆！`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")

			// Check if logs directory exists
			if _, err := os.Stat(logsDir); os.IsNotExist(err) {
				return fmt.Errorf("日志目录不存在: %s", logsDir)
			}

			var filesToClear []string

			if len(args) == 1 {
				// Clear specific file
				filesToClear = append(filesToClear, args[0])
			} else {
				// Clear all log files
				entries, err := os.ReadDir(logsDir)
				if err != nil {
					return fmt.Errorf("读取日志目录失败: %w", err)
				}

				for _, entry := range entries {
					if !entry.IsDir() {
						filesToClear = append(filesToClear, entry.Name())
					}
				}
			}

			if len(filesToClear) == 0 {
				fmt.Println("ℹ 没有日志文件可清空")
				return nil
			}

			// Confirm
			fmt.Printf("警告: 即将清空 %d 个日志文件,此操作不可逆!\n", len(filesToClear))
			for _, file := range filesToClear {
				fmt.Printf("  • %s\n", file)
			}
			fmt.Print("确认继续? [yes/NO]: ")

			var confirm string
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) != "yes" {
				fmt.Println("已取消")
				return nil
			}

			// Clear files
			count := 0
			for _, file := range filesToClear {
				logPath := filepath.Join(logsDir, file)
				if err := os.Truncate(logPath, 0); err != nil {
					fmt.Printf("✗ 清空失败: %s - %v\n", file, err)
					continue
				}
				count++
			}

			fmt.Printf("✓ 已清空 %d 个日志文件\n", count)
			return nil
		},
	}
	logsCmd.AddCommand(clearCmd)

	// ── delete ────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <log-file>",
		Short: "删除日志文件",
		Long:  `删除指定的日志文件。此操作不可逆！`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")
			logPath := filepath.Join(logsDir, args[0])

			// Check if file exists
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				return fmt.Errorf("日志文件不存在: %s", args[0])
			}

			// Confirm
			fmt.Printf("警告: 即将删除日志文件: %s,此操作不可逆!\n", args[0])
			fmt.Print("确认继续? [yes/NO]: ")

			var confirm string
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) != "yes" {
				fmt.Println("已取消")
				return nil
			}

			// Delete file
			if err := os.Remove(logPath); err != nil {
				return fmt.Errorf("删除失败: %w", err)
			}

			fmt.Printf("✓ 已删除日志文件: %s\n", args[0])
			return nil
		},
	}
	logsCmd.AddCommand(deleteCmd)

	// ── stats ─────────────────────────────────────────────────────────────
	statsCmd := &cobra.Command{
		Use:   "stats [log-file]",
		Short: "显示日志统计信息",
		Long:  `显示日志文件的统计信息，包括行数、各级别日志数量等。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := getDefaultDataDir()
			if err != nil {
				return fmt.Errorf("获取数据目录失败: %w", err)
			}

			logsDir := filepath.Join(dataDir, "logs")

			// Determine log file
			var logFile string
			if len(args) == 1 {
				logFile = args[0]
			} else {
				// Find latest log file
				latest, err := findLatestLogFile(logsDir)
				if err != nil {
					return fmt.Errorf("查找最新日志文件失败: %w", err)
				}
				logFile = latest
			}

			logPath := filepath.Join(logsDir, logFile)

			// Show stats
			return showLogStats(logPath)
		},
	}
	logsCmd.AddCommand(statsCmd)

	rootCmd.AddCommand(logsCmd)
}

// ── Helper Functions ───────────────────────────────────────────────────────

// findLatestLogFile finds the most recently modified log file
func findLatestLogFile(logsDir string) (string, error) {
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("日志目录不存在: %s", logsDir)
	}

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return "", fmt.Errorf("读取日志目录失败: %w", err)
	}

	var latestFile string
	var latestModTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if latestModTime.IsZero() || info.ModTime().After(latestModTime) {
			latestModTime = info.ModTime()
			latestFile = entry.Name()
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("未找到日志文件")
	}

	return latestFile, nil
}

// viewLogs displays log file content with filtering
func viewLogs(logPath string, lines int, level string, source string, since string, tailOnly bool) error {
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// Parse since time
	var sinceTime time.Time
	if since != "" {
		sinceTime, err = parseSinceTime(since)
		if err != nil {
			return fmt.Errorf("解析时间参数失败: %w", err)
		}
	}

	// Color definitions
	levelColors := map[string]*color.Color{
		"DEBUG": color.New(color.FgCyan),
		"INFO":  color.New(color.FgGreen),
		"WARN":  color.New(color.FgYellow),
		"ERROR": color.New(color.FgRed),
	}

	// Read and filter
	scanner := bufio.NewScanner(file)
	var filteredLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Level filter
		if level != "" && !strings.Contains(strings.ToUpper(line), level) {
			continue
		}

		// Source filter
		if source != "" && !strings.Contains(line, source) {
			continue
		}

		// Since filter
		if !sinceTime.IsZero() {
			if logTime, err := parseLogTime(line); err == nil {
				if logTime.Before(sinceTime) {
					continue
				}
			}
		}

		filteredLines = append(filteredLines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取日志文件失败: %w", err)
	}

	// Apply tail-only
	if tailOnly && len(filteredLines) > lines {
		filteredLines = filteredLines[len(filteredLines)-lines:]
	} else if len(filteredLines) > lines {
		filteredLines = filteredLines[len(filteredLines)-lines:]
	}

	// Display with colors
	for _, line := range filteredLines {
		// Apply level color
		if level != "" {
			if c, ok := levelColors[strings.ToUpper(level)]; ok {
				line = c.Sprint(line)
			}
		} else {
			// Auto-detect level
			for lvl, c := range levelColors {
				if strings.Contains(strings.ToUpper(line), lvl) {
					line = c.Sprint(line)
					break
				}
			}
		}

		fmt.Println(line)
	}

	return nil
}

// followLogs tails the log file and displays new content
func followLogs(logPath string, level string, source string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// Seek to end
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("定位到文件末尾失败: %w", err)
	}

	// Color definitions
	levelColors := map[string]*color.Color{
		"DEBUG": color.New(color.FgCyan),
		"INFO":  color.New(color.FgGreen),
		"WARN":  color.New(color.FgYellow),
		"ERROR": color.New(color.FgRed),
	}

	// Read new content
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}

		// Level filter
		if level != "" && !strings.Contains(strings.ToUpper(line), level) {
			continue
		}

		// Source filter
		if source != "" && !strings.Contains(line, source) {
			continue
		}

		// Apply color
		displayLine := strings.TrimSpace(line)
		if displayLine != "" {
			if level != "" {
				if c, ok := levelColors[strings.ToUpper(level)]; ok {
					displayLine = c.Sprint(displayLine)
				}
			} else {
				// Auto-detect level
				for lvl, c := range levelColors {
					if strings.Contains(strings.ToUpper(displayLine), lvl) {
						displayLine = c.Sprint(displayLine)
						break
					}
				}
			}
			fmt.Println(displayLine)
		}
	}
}

// showLogStats displays statistics about the log file
func showLogStats(logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// Count lines and levels
	scanner := bufio.NewScanner(file)
	var totalLines int
	levelCounts := map[string]int{
		"DEBUG": 0,
		"INFO":  0,
		"WARN":  0,
		"ERROR": 0,
	}

	for scanner.Scan() {
		totalLines++
		line := strings.ToUpper(scanner.Text())
		for lvl := range levelCounts {
			if strings.Contains(line, lvl) {
				levelCounts[lvl]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取日志文件失败: %w", err)
	}

	// Display stats
	fmt.Printf("日志文件: %s\n", filepath.Base(logPath))
	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("文件大小:     %8.2f MB\n", float64(info.Size())/(1024*1024))
	fmt.Printf("总行数:       %8d\n", totalLines)
	fmt.Println()
	fmt.Println("日志级别统计:")
	for _, lvl := range []string{"ERROR", "WARN", "INFO", "DEBUG"} {
		count := levelCounts[lvl]
		percentage := float64(count) / float64(totalLines) * 100
		fmt.Printf("  %-6s: %8d 行 (%6.2f%%)\n", lvl, count, percentage)
	}
	fmt.Println("────────────────────────────────────────────────────────────")

	return nil
}

// parseSinceTime parses the since time parameter
func parseSinceTime(since string) (time.Time, error) {
	// Try common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, since); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("不支持的时间格式: %s", since)
}

// parseLogTime extracts timestamp from log line
func parseLogTime(line string) (time.Time, error) {
	// Try to extract time from common log formats
	// ISO 8601: 2006-01-02T15:04:05.000Z
	if strings.Contains(line, "T") {
		// Find ISO 8601 timestamp
		idx := strings.Index(line, "2")
		if idx >= 0 && len(line) >= idx+19 {
			timestamp := line[idx : idx+19]
			if strings.Contains(timestamp, "T") {
				if t, err := time.Parse(time.RFC3339, timestamp+"Z"); err == nil {
					return t, nil
				}
			}
		}
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, line); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日志时间")
}
