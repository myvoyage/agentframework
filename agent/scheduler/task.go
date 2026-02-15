// Agent Framework - Task Types and Helpers
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Task represents a scheduled task
type Task struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CronExpr    string      `json:"cron_expr"`
	Type        TaskType    `json:"type"`
	Enabled     bool        `json:"enabled"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	NextRun     time.Time   `json:"next_run"`
	LastRun     *time.Time  `json:"last_run,omitempty"`
	RunCount    int64       `json:"run_count"`

	// Type-specific fields
	StaticMessage *string       `json:"static_message,omitempty"` // For static tasks
	AIPrompt      *string       `json:"ai_prompt,omitempty"`       // For AI tasks
	AIModel       *string       `json:"ai_model,omitempty"`        // For AI tasks
	AITools       []string      `json:"ai_tools,omitempty"`        // For AI tasks
}

// TaskType defines the type of scheduled task
type TaskType string

const (
	TaskTypeStatic TaskType = "static" // Static message task
	TaskTypeAI     TaskType = "ai"      // AI-powered intelligent task
)

// StaticTask creates a static message task
func StaticTask(name, cronExpr, message string) *Task {
	return &Task{
		Name:          name,
		Description:   fmt.Sprintf("Static message: %s", message),
		CronExpr:      cronExpr,
		Type:          TaskTypeStatic,
		Enabled:       true,
		StaticMessage: &message,
	}
}

// AITask creates an AI-powered task
func AITask(name, cronExpr, prompt string, opts ...AITaskOption) *Task {
	task := &Task{
		Name:        name,
		Description: fmt.Sprintf("AI task: %s", prompt),
		CronExpr:    cronExpr,
		Type:        TaskTypeAI,
		Enabled:     true,
		AIPrompt:    &prompt,
		AIModel:     stringPtr("default"), // Default model
		AITools:     []string{},           // No tools by default
	}

	// Apply options
	for _, opt := range opts {
		opt(task)
	}

	return task
}

// AITaskOption configures an AI task
type AITaskOption func(*Task)

// WithAIModel sets the AI model for the task
func WithAIModel(model string) AITaskOption {
	return func(t *Task) {
		t.AIModel = &model
	}
}

// WithAITools enables specific tools for the AI task
func WithAITools(tools ...string) AITaskOption {
	return func(t *Task) {
		t.AITools = tools
	}
}

// ParseNaturalLanguage parses natural language task description
// Supports formats like:
// - "每天早上9点提醒我开会"
// - "每小时43分发一段鸡汤激励我写代码"
// - "每周一早上8点搜索AI新闻发给我摘要"
func ParseNaturalLanguage(input string) (*Task, error) {
	input = strings.TrimSpace(input)

	// Try common patterns
	type patternHandler struct {
		pattern string
		handler func([]string) (*Task, error)
	}

	patterns := []patternHandler{
		// Daily tasks
		{`每天(.+?\d{1,2}).+?\d{1,2})点(.+)`, parseDailyTask},
		// Hourly tasks
		{`每小时(.+?)`, parseHourlyTask},
		// Weekly tasks
		{`每周(.+?)(.{1,2}).+?\d{1,2}.+?\d{1,2})点(.+)`, parseWeeklyTask},
		// Cron expression tasks
		{`cron\s+(.+)\s+(.+)`, parseCronTask},
	}

	for _, p := range patterns {
		if re, err := regexp.Compile(p.pattern); err == nil {
			if matches := re.FindStringSubmatch(input); len(matches) > 0 {
				return p.handler(matches)
			}
		}
	}

	return nil, fmt.Errorf("unable to parse task description: %s", input)
}

// parseDailyTask parses daily recurring tasks
// Format: "每天早上9点提醒我开会"
func parseDailyTask(matches []string) (*Task, error) {
	timeStr := matches[2]
	message := strings.TrimSpace(matches[3])

	// Convert Chinese time to 24-hour format
	hour, minute, err := parseChineseTime(timeStr)
	if err != nil {
		return nil, err
	}

	// Build cron expression: "0 minute hour * * *"
	cronExpr := fmt.Sprintf("0 %d %d * * *", minute, hour)

	return StaticTask(
		fmt.Sprintf("每日任务_%s", message[:min(10, len(message))]),
		cronExpr,
		message,
	), nil
}

// parseHourlyTask parses hourly recurring tasks
// Format: "每小时43分发一段鸡汤激励我写代码"
func parseHourlyTask(matches []string) (*Task, error) {
	minuteStr := strings.TrimSpace(matches[1])
	message := strings.TrimSpace(matches[2])

	// Extract minute (e.g., "43分" -> 43)
	minute := 0
	if len(minuteStr) > 0 {
		if _, err := fmt.Sscanf(minuteStr, "%d", &minute); err != nil {
			// Try Chinese digit patterns
			minute = parseChineseMinute(minuteStr)
		}
	}

	// Build cron expression: "0 minute * * * *"
	cronExpr := fmt.Sprintf("0 %d * * * *", minute)

	return AITask(
		fmt.Sprintf("每小时任务_%s", message[:min(10, len(message))]),
		cronExpr,
		message,
		WithAITools("web_search"), // Enable web search by default
	), nil
}

// parseWeeklyTask parses weekly recurring tasks
// Format: "每周一早上8点搜索AI新闻发给我摘要"
func parseWeeklyTask(matches []string) (*Task, error) {
	weekdayStr := strings.TrimSpace(matches[2])
	timeStr := matches[3]
	message := strings.TrimSpace(matches[4])

	// Convert weekday to cron format (0-6, Sunday=0)
	weekday, err := parseChineseWeekday(weekdayStr)
	if err != nil {
		return nil, err
	}

	// Convert Chinese time to 24-hour format
	hour, minute, err := parseChineseTime(timeStr)
	if err != nil {
		return nil, err
	}

	// Build cron expression: "0 minute hour * * weekday"
	cronExpr := fmt.Sprintf("0 %d %d * * %d", minute, hour, weekday)

	return AITask(
		fmt.Sprintf("每周任务_%s", message[:min(10, len(message))]),
		cronExpr,
		message,
		WithAITools("web_search", "weather"), // Enable multiple tools
	), nil
}

// parseCronTask parses tasks with explicit cron expressions
// Format: "cron 0 9 * * * 开会"
func parseCronTask(matches []string) (*Task, error) {
	cronExpr := strings.TrimSpace(matches[1])
	message := strings.TrimSpace(matches[2])

	return StaticTask(
		fmt.Sprintf("Cron任务_%s", message[:min(10, len(message))]),
		cronExpr,
		message,
	), nil
}

// parseChineseTime converts Chinese time expressions to hour and minute
// Supports: "早上9点", "下午3点30", "晚上8点"
func parseChineseTime(timeStr string) (hour, minute int, err error) {
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.ReplaceAll(timeStr, "点", ":")

	// Handle period prefixes
	hourOffset := 0
	if strings.Contains(timeStr, "早上") || strings.Contains(timeStr, "上午") {
		timeStr = strings.ReplaceAll(timeStr, "早上", "")
		timeStr = strings.ReplaceAll(timeStr, "上午", "")
	} else if strings.Contains(timeStr, "下午") {
		timeStr = strings.ReplaceAll(timeStr, "下午", "")
		hourOffset = 12
	} else if strings.Contains(timeStr, "晚上") {
		timeStr = strings.ReplaceAll(timeStr, "晚上", "")
		hourOffset = 12
	}

	// Parse time
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &hour)
			fmt.Sscanf(parts[1], "%d", &minute)
		}
	} else {
		fmt.Sscanf(timeStr, "%d", &hour)
		minute = 0
	}

	hour += hourOffset
	return hour, minute, nil
}

// parseChineseWeekday converts Chinese weekday to cron format
func parseChineseWeekday(weekday string) (int, error) {
	weekdayMap := map[string]int{
		"日": 0, "周日": 0, "星期日": 0, "星期天": 0,
		"一": 1, "周一": 1, "星期一": 1,
		"二": 2, "周二": 2, "星期二": 2,
		"三": 3, "周三": 3, "星期三": 3,
		"四": 4, "周四": 4, "星期四": 4,
		"五": 5, "周五": 5, "星期五": 5,
		"六": 6, "周六": 6, "星期六": 6,
	}

	if wd, ok := weekdayMap[weekday]; ok {
		return wd, nil
	}
	return 0, fmt.Errorf("unknown weekday: %s", weekday)
}

// parseChineseMinute converts Chinese minute patterns
func parseChineseMinute(minuteStr string) int {
	// Common patterns: "43分" -> 43
	minute := 0
	fmt.Sscanf(minuteStr, "%d", &minute)
	return minute
}

// stringPtr returns a pointer to string
func stringPtr(s string) *string {
	return &s
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
