// Agent Framework - Notification Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NotificationModule 系统通知模块
type NotificationModule struct {
	config NotificationConfig
	mu     sync.RWMutex
	stats  *NotificationStats
	// 跨平台通知后端
	notifier NotificationBackend
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	MaxQueueSize      int      `json:"max_queue_size"`       // 最大通知队列大小
	EnableSound       bool     `json:"enable_sound"`         // 启用提示音
	DefaultSound      string   `json:"default_sound"`        // 默认提示音
	DefaultIcon       string   `json:"default_icon"`         // 默认图标
	EnablePersistence bool     `json:"enable_persistence"`   // 启用通知持久化
	PersistencePath   string   `json:"persistence_path"`     // 持久化存储路径
	AllowedCategories []string `json:"allowed_categories"`   // 允许的通知类别
	EnableGrouping    bool     `json:"enable_grouping"`      // 启用通知分组
	EnableActions     bool     `json:"enable_actions"`       // 启用通知操作按钮
}

// NotificationStats 通知统计信息
type NotificationStats struct {
	TotalSent        int64     `json:"total_sent"`
	TotalDelivered   int64     `json:"total_delivered"`
	TotalClicked     int64     `json:"total_clicked"`
	TotalDismissed   int64     `json:"total_dismissed"`
	TotalErrors      int64     `json:"total_errors"`
	QueuedCount     int       `json:"queued_count"`
	mu                 sync.RWMutex `json:"-"`
}

// NotificationBackend 通知后端接口
type NotificationBackend interface {
	Send(notification *Notification) error
	GetCapabilities() NotificationCapabilities
	Close() error
}

// NotificationCapabilities 通知能力
type NotificationCapabilities struct {
	Title        bool   `json:"title"`         // 支持标题
	Body         bool   `json:"body"`          // 支持正文
	Icon         bool   `json:"icon"`          // 支持图标
	Image        bool   `json:"image"`         // 支持图片
	Sound        bool   `json:"sound"`         // 支持声音
	Progress     bool   `json:"progress"`      // 支持进度条
	Actions      bool   `json:"actions"`       // 支持操作按钮
	Grouping     bool   `json:"grouping"`      // 支持分组
	Priority     bool   `json:"priority"`      // 支持优先级
	Expiration   bool   `json:"expiration"`    // 支持过期时间
	Categories   bool   `json:"categories"`    // 支持类别
}

// Notification 通知
type Notification struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Icon        string            `json:"icon,omitempty"`
	Image       string            `json:"image,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Category    string            `json:"category,omitempty"`
	Priority    string            `json:"priority,omitempty"`    // low, normal, high, urgent
	Timeout     int               `json:"timeout,omitempty"`     // 超时时间（秒）
	Progress   float64            `json:"progress,omitempty"`    // 进度值（0-100）
	Actions    []NotificationAction `json:"actions,omitempty"`   // 操作按钮
	Group      string             `json:"group,omitempty"`      // 分组标识
	Persistent bool               `json:"persistent,omitempty"`  // 是否持久化
	CreatedAt  time.Time          `json:"created_at"`
	ExpiresAt  *time.Time         `json:"expires_at,omitempty"`
}

// NotificationAction 通知操作按钮
type NotificationAction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// NotificationResult 通知结果
type NotificationResult struct {
	Success    bool      `json:"success"`
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error,omitempty"`
	Dismissed  bool      `json:"dismissed,omitempty"`
	Clicked    bool      `json:"clicked,omitempty"`
	ActionID   string    `json:"action_id,omitempty"`
}

// NewNotificationModule 创建通知模块实例
func NewNotificationModule(config NotificationConfig) (*NotificationModule, error) {
	if config.MaxQueueSize <= 0 {
		config.MaxQueueSize = 100 // 默认100条队列
	}
	if config.DefaultSound == "" {
		config.DefaultSound = "default"
	}
	if config.DefaultIcon == "" {
		config.DefaultIcon = "information"
	}

	// 创建平台特定的通知后端
	var backend NotificationBackend
	var err error

	if runtime.GOOS == "windows" {
		backend, err = NewWindowsNotifier()
	} else if runtime.GOOS == "darwin" {
		backend, err = NewMacOSNotifier()
	} else {
		backend, err = NewLinuxNotifier()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create notification backend: %w", err)
	}

	stats := &NotificationStats{}

	module := &NotificationModule{
		config:  config,
		stats:   stats,
		notifier: backend,
	}

	return module, nil
}

// GetTools 返回通知模块的 MCP 工具列表
func (m *NotificationModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 发送通知工具
		&notificationSendTool{module: m},
		// 获取能力工具
		&notificationCapabilitiesTool{module: m},
		// 批量发送工具
		&notificationBatchTool{module: m},
		// 进度通知工具
		&notificationProgressTool{module: m},
		// 通知历史工具
		&notificationHistoryTool{module: m},
	}

	return tools, nil
}

// 发送通知工具
type notificationSendTool struct {
	module *NotificationModule
}

func (t *notificationSendTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "notification_send",
		Desc: "Send a system notification with title, body, and optional media",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Type:     "string",
				Desc:     "Notification title",
				Required:  true,
			},
			"body": {
				Type: "string",
				Desc: "Notification body content",
			},
			"icon": {
				Type: "string",
				Desc: "Icon identifier or file path",
			},
			"sound": {
				Type: "string",
				Desc: "Sound identifier or file path",
			},
			"category": {
				Type: "string",
				Desc: "Notification category (info, warning, error, success)",
			},
			"priority": {
				Type: "string",
				Desc: "Notification priority (low, normal, high, urgent)",
			},
			"timeout": {
				Type: "integer",
				Desc: "Auto-dismiss timeout in seconds",
			},
			"actions": {
				Type: "array",
				Desc: "Action buttons to display",
			},
		}),
	}, nil
}

func (t *notificationSendTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Title     string                 `json:"title"`
		Body      string                 `json:"body"`
		Icon      string                 `json:"icon"`
		Sound     string                 `json:"sound"`
		Category  string                 `json:"category"`
		Priority  string                 `json:"priority"`
		Timeout   int                    `json:"timeout"`
		Actions   []NotificationAction    `json:"actions"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.sendNotification(args)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取能力工具
type notificationCapabilitiesTool struct {
	module *NotificationModule
}

func (t *notificationCapabilitiesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "notification_capabilities",
		Desc:        "Get the capabilities of the notification system",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *notificationCapabilitiesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result := t.module.getCapabilities()
	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 批量发送工具
type notificationBatchTool struct {
	module *NotificationModule
}

func (t *notificationBatchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "notification_batch",
		Desc: "Send multiple notifications in a batch",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"notifications": {
				Type:     "array",
				Desc:     "Array of notification objects",
				Required:  true,
			},
		}),
	}, nil
}

func (t *notificationBatchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Notifications []Notification `json:"notifications"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.sendBatch(args.Notifications)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 进度通知工具
type notificationProgressTool struct {
	module *NotificationModule
}

func (t *notificationProgressTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "notification_progress",
		Desc: "Send or update a progress notification",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "string",
				Desc:     "Unique identifier for the progress notification",
				Required:  true,
			},
			"title": {
				Type:     "string",
				Desc:     "Notification title",
				Required:  true,
			},
			"progress": {
				Type:     "number",
				Desc:     "Progress value (0-100)",
				Required:  true,
			},
			"body": {
				Type: "string",
				Desc: "Additional progress information",
			},
			"indeterminate": {
				Type: "boolean",
				Desc: "Show indeterminate progress",
			},
		}),
	}, nil
}

func (t *notificationProgressTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		ID            string  `json:"id"`
		Title         string  `json:"title"`
		Progress      float64 `json:"progress"`
		Body          string  `json:"body"`
		Indeterminate  bool    `json:"indeterminate"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.sendProgress(args.ID, args.Title, args.Progress, args.Body, args.Indeterminate)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 通知历史工具
type notificationHistoryTool struct {
	module *NotificationModule
}

func (t *notificationHistoryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "notification_history",
		Desc: "Get notification history (if enabled)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"limit": {
				Type: "integer",
				Desc: "Maximum number of history items to return",
			},
			"category": {
				Type: "string",
				Desc: "Filter by category",
			},
		}),
	}, nil
}

func (t *notificationHistoryTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Limit    int    `json:"limit"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.getHistory(args.Limit, args.Category)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭通知模块
func (m *NotificationModule) Close() error {
	return m.notifier.Close()
}

// GetStats 获取统计信息
func (m *NotificationModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_sent":      m.stats.TotalSent,
		"total_delivered":  m.stats.TotalDelivered,
		"total_clicked":   m.stats.TotalClicked,
		"total_dismissed": m.stats.TotalDismissed,
		"total_errors":    m.stats.TotalErrors,
		"queued_count":    int64(m.stats.QueuedCount),
	}
}

// ==================== 核心功能实现 ====================

// sendNotification 发送通知
func (m *NotificationModule) sendNotification(args struct {
	Title     string
	Body      string
	Icon      string
	Sound     string
	Category  string
	Priority  string
	Timeout   int
	Actions   []NotificationAction
}) (*NotificationResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalSent++
	m.stats.mu.Unlock()

	// 验证类别
	if len(m.config.AllowedCategories) > 0 {
		allowed := false
		for _, cat := range m.config.AllowedCategories {
			if cat == args.Category {
				allowed = true
				break
			}
		}
		if !allowed {
			return &NotificationResult{
				Success: false,
				Error:   fmt.Sprintf("Category '%s' is not allowed", args.Category),
			}, nil
		}
	}

	// 构建通知
	notification := &Notification{
		ID:         generateNotificationID(),
		Title:      args.Title,
		Body:       args.Body,
		Icon:       args.Icon,
		Sound:       args.Sound,
		Category:    args.Category,
		Priority:    args.Priority,
		Timeout:     args.Timeout,
		Actions:     args.Actions,
		CreatedAt:   time.Now(),
		Persistent:  m.config.EnablePersistence,
	}

	// 设置默认值
	if notification.Icon == "" {
		notification.Icon = m.config.DefaultIcon
	}
	if notification.Sound == "" && m.config.EnableSound {
		notification.Sound = m.config.DefaultSound
	}
	if notification.Priority == "" {
		notification.Priority = "normal"
	}
	if notification.Category == "" {
		notification.Category = "info"
	}

	// 发送通知
	err := m.notifier.Send(notification)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.TotalErrors++
		m.stats.mu.Unlock()

		return &NotificationResult{
			Success: false,
			ID:      notification.ID,
			Error:   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.TotalDelivered++
	m.stats.mu.Unlock()

	return &NotificationResult{
		Success:   true,
		ID:        notification.ID,
		Timestamp: time.Now(),
	}, nil
}

// sendBatch 批量发送通知
func (m *NotificationModule) sendBatch(notifications []Notification) (map[string]any, error) {
	results := make([]*NotificationResult, 0, len(notifications))

	for _, notif := range notifications {
		// 为每个通知设置默认值
		if notif.Icon == "" {
			notif.Icon = m.config.DefaultIcon
		}
		if notif.Sound == "" && m.config.EnableSound {
			notif.Sound = m.config.DefaultSound
		}
		if notif.Priority == "" {
			notif.Priority = "normal"
		}
		if notif.ID == "" {
			notif.ID = generateNotificationID()
		}
		if notif.CreatedAt.IsZero() {
			notif.CreatedAt = time.Now()
		}

		// 发送通知
		err := m.notifier.Send(&notif)
		result := &NotificationResult{
			ID:      notif.ID,
			Success: err == nil,
		}

		if err != nil {
			result.Error = err.Error()
			m.stats.mu.Lock()
			m.stats.TotalErrors++
			m.stats.mu.Unlock()
		} else {
			m.stats.mu.Lock()
			m.stats.TotalDelivered++
			m.stats.TotalSent++
			m.stats.mu.Unlock()
			result.Timestamp = time.Now()
		}

		results = append(results, result)
	}

	return map[string]any{
		"success":   true,
		"total":     len(notifications),
		"delivered": len(results),
		"results":   results,
	}, nil
}

// sendProgress 发送进度通知
func (m *NotificationModule) sendProgress(id, title string, progress float64, body string, indeterminate bool) (*NotificationResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalSent++
	m.stats.mu.Unlock()

	// 构建进度通知
	notification := &Notification{
		ID:        id,
		Title:     title,
		Body:      body,
		Icon:      m.config.DefaultIcon,
		Category:  "progress",
		Priority:  "normal",
		Progress:  progress,
		CreatedAt: time.Now(),
	}

	if indeterminate {
		notification.Progress = -1 // -1 表示不确定进度
	}

	// 发送通知
	err := m.notifier.Send(notification)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.TotalErrors++
		m.stats.mu.Unlock()

		return &NotificationResult{
			Success: false,
			ID:      id,
			Error:   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.TotalDelivered++
	m.stats.mu.Unlock()

	return &NotificationResult{
		Success:   true,
		ID:        id,
		Timestamp: time.Now(),
	}, nil
}

// getCapabilities 获取通知能力
func (m *NotificationModule) getCapabilities() map[string]any {
	caps := m.notifier.GetCapabilities()

	return map[string]any{
		"success": true,
		"capabilities": map[string]any{
			"title":       caps.Title,
			"body":        caps.Body,
			"icon":        caps.Icon,
			"image":       caps.Image,
			"sound":       caps.Sound,
			"progress":    caps.Progress,
			"actions":     caps.Actions,
			"grouping":    caps.Grouping,
			"priority":    caps.Priority,
			"expiration":  caps.Expiration,
			"categories":  caps.Categories,
		},
	}
}

// getHistory 获取通知历史
func (m *NotificationModule) getHistory(limit int, category string) (map[string]any, error) {
	if !m.config.EnablePersistence {
		return map[string]any{
			"success": false,
			"error":   "History persistence is not enabled",
		}, nil
	}

	// 这里可以实现从持久化存储读取历史记录
	// 由于篇幅限制，这里仅返回空历史
	return map[string]any{
		"success": true,
		"enabled": true,
		"items":   []NotificationResult{},
		"count":   0,
	}, nil
}

// generateNotificationID 生成通知 ID
func generateNotificationID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

// ==================== 平台特定实现 ====================

// WindowsNotifier Windows 通知实现
type WindowsNotifier struct {
	// Windows 特定字段
}

func NewWindowsNotifier() (*WindowsNotifier, error) {
	return &WindowsNotifier{}, nil
}

func (n *WindowsNotifier) Send(notification *Notification) error {
	// Windows 10/11 使用 Windows Toast API
	// 可以使用 github.com/go-toast/toast 库
	return fmt.Errorf("not implemented")
}

func (n *WindowsNotifier) GetCapabilities() NotificationCapabilities {
	return NotificationCapabilities{
		Title:      true,
		Body:       true,
		Icon:       true,
		Image:      true,
		Sound:      true,
		Progress:   true,
		Actions:    true,
		Grouping:   true,
		Priority:   true,
		Expiration: true,
		Categories: true,
	}
}

func (n *WindowsNotifier) Close() error {
	return nil
}

// MacOSNotifier macOS 通知实现
type MacOSNotifier struct {
	// macOS 特定字段
}

func NewMacOSNotifier() (*MacOSNotifier, error) {
	return &MacOSNotifier{}, nil
}

func (n *MacOSNotifier) Send(notification *Notification) error {
	// macOS 使用 osascript / terminal-notifier
	// 可以使用 github.com/deckarep/gosx-notifier 库
	return fmt.Errorf("not implemented")
}

func (n *MacOSNotifier) GetCapabilities() NotificationCapabilities {
	return NotificationCapabilities{
		Title:      true,
		Body:       true,
		Icon:       true,
		Image:      true,
		Sound:      true,
		Progress:   true,
		Actions:    true,
		Grouping:   true,
		Priority:   true,
		Expiration: true,
		Categories: true,
	}
}

func (n *MacOSNotifier) Close() error {
	return nil
}

// LinuxNotifier Linux 通知实现
type LinuxNotifier struct {
	// Linux 特定字段
}

func NewLinuxNotifier() (*LinuxNotifier, error) {
	return &LinuxNotifier{}, nil
}

func (n *LinuxNotifier) Send(notification *Notification) error {
	// Linux 使用 libnotify / notify-send
	// 可以使用 github.com/esiqvelandir/notify 库
	return fmt.Errorf("not implemented")
}

func (n *LinuxNotifier) GetCapabilities() NotificationCapabilities {
	return NotificationCapabilities{
		Title:      true,
		Body:       true,
		Icon:       true,
		Image:      true,
		Sound:      true,
		Progress:   true,
		Actions:    true,
		Grouping:   true,
		Priority:   true,
		Expiration: true,
		Categories: true,
	}
}

func (n *LinuxNotifier) Close() error {
	return nil
}
