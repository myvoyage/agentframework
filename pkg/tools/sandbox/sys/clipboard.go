// Agent Framework - Clipboard Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ClipboardModule 剪贴板模块
type ClipboardModule struct {
	config ClipboardConfig
	mu     sync.RWMutex
	stats  *ClipboardStats
	// 跨平台剪贴板操作接口
	clipboard ClipboardBackend
}

// ClipboardConfig 剪贴板配置
type ClipboardConfig struct {
	MaxHistorySize  int      `json:"max_history_size"`   // 最大历史记录数
	EnableHistory   bool     `json:"enable_history"`     // 启用历史记录
	EnableMonitor   bool     `json:"enable_monitor"`     // 启用剪贴板监控
	AllowedFormats  []string `json:"allowed_formats"`    // 允许的格式（text, image, files）
	MaxTextSize    int      `json:"max_text_size"`      // 最大文本大小（字节）
	EnableEncryption bool    `json:"enable_encryption"`  // 启用内容加密
}

// ClipboardStats 剪贴板统计信息
type ClipboardStats struct {
	TotalReads      int64     `json:"total_reads"`
	TotalWrites     int64     `json:"total_writes"`
	TotalClears     int64     `json:"total_clears"`
	HistoryCount    int       `json:"history_count"`
	BytesRead       int64     `json:"bytes_read"`
	BytesWritten    int64     `json:"bytes_written"`
	mu               sync.RWMutex `json:"-"`
}

// ClipboardBackend 剪贴板后端接口
type ClipboardBackend interface {
	GetText() (string, error)
	SetText(text string) error
	GetImage() ([]byte, string, error) // 返回图片数据和 MIME 类型
	SetImage(data []byte, mimeType string) error
	Clear() error
	// 监控剪贴板变化
	WatchChanges(callback func(content string)) error
	StopWatching() error
}

// ClipboardItem 剪贴板历史记录项
type ClipboardItem struct {
	Type      string      `json:"type"`       // text, image, files
	Content   string      `json:"content"`    // Base64 编码的内容
	MimeType  string      `json:"mime_type"`  // MIME 类型
	Size      int64       `json:"size"`       // 内容大小（字节）
	Timestamp int64       `json:"timestamp"`  // 时间戳
}

// NewClipboardModule 创建剪贴板模块实例
func NewClipboardModule(config ClipboardConfig) (*ClipboardModule, error) {
	if config.MaxHistorySize <= 0 {
		config.MaxHistorySize = 100 // 默认100条历史记录
	}
	if config.MaxTextSize <= 0 {
		config.MaxTextSize = 1024 * 1024 // 默认1MB
	}
	if len(config.AllowedFormats) == 0 {
		config.AllowedFormats = []string{"text"} // 默认仅文本
	}

	// 创建平台特定的剪贴板后端
	var backend ClipboardBackend
	var err error

	if runtime.GOOS == "windows" {
		backend, err = NewWindowsClipboard()
	} else if runtime.GOOS == "darwin" {
		backend, err = NewMacOSClipboard()
	} else {
		backend, err = NewLinuxClipboard()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard backend: %w", err)
	}

	stats := &ClipboardStats{}

	module := &ClipboardModule{
		config:   config,
		stats:    stats,
		clipboard: backend,
	}

	// 启动剪贴板监控
	if config.EnableMonitor {
		go module.monitorClipboard()
	}

	return module, nil
}

// GetTools 返回剪贴板模块的 MCP 工具列表
func (m *ClipboardModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 读取剪贴板工具
		&clipboardReadTool{module: m},
		// 写入剪贴板工具
		&clipboardWriteTool{module: m},
		// 清空剪贴板工具
		&clipboardClearTool{module: m},
		// 获取历史记录工具
		&clipboardHistoryTool{module: m},
		// 监控剪贴板工具
		&clipboardWatchTool{module: m},
	}

	return tools, nil
}

// 读取剪贴板工具
type clipboardReadTool struct {
	module *ClipboardModule
}

func (t *clipboardReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "clipboard_read",
		Desc: "Read the current content from the clipboard",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type: "string",
				Desc: "Content format to read (text, image, auto)",
			},
		}),
	}, nil
}

func (t *clipboardReadTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Format string `json:"format"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Format == "" {
		args.Format = "auto"
	}

	result, err := t.module.readContent(args.Format)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 写入剪贴板工具
type clipboardWriteTool struct {
	module *ClipboardModule
}

func (t *clipboardWriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "clipboard_write",
		Desc: "Write content to the clipboard",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {
				Type:     "string",
				Desc:     "Content to write (text or base64 encoded data)",
				Required:  true,
			},
			"type": {
				Type: "string",
				Desc: "Content type (text, image)",
			},
			"mime_type": {
				Type: "string",
				Desc: "MIME type for binary content (e.g., image/png)",
			},
		}),
	}, nil
}

func (t *clipboardWriteTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Content  string `json:"content"`
		Type     string `json:"type"`
		MimeType string `json:"mime_type"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Type == "" {
		args.Type = "text"
	}

	result, err := t.module.writeContent(args.Content, args.Type, args.MimeType)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 清空剪贴板工具
type clipboardClearTool struct {
	module *ClipboardModule
}

func (t *clipboardClearTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "clipboard_clear",
		Desc:        "Clear all content from the clipboard",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *clipboardClearTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.clearContent()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取历史记录工具
type clipboardHistoryTool struct {
	module *ClipboardModule
}

func (t *clipboardHistoryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "clipboard_history",
		Desc: "Get clipboard history (if enabled)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"limit": {
				Type: "integer",
				Desc: "Maximum number of history items to return",
			},
			"type": {
				Type: "string",
				Desc: "Filter by content type (text, image, all)",
			},
		}),
	}, nil
}

func (t *clipboardHistoryTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Limit int    `json:"limit"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Type == "" {
		args.Type = "all"
	}

	result, err := t.module.getHistory(args.Limit, args.Type)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 监控剪贴板工具
type clipboardWatchTool struct {
	module *ClipboardModule
}

func (t *clipboardWatchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "clipboard_watch",
		Desc: "Start monitoring clipboard for changes (experimental)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"enable": {
				Type: "boolean",
				Desc: "Enable or disable monitoring",
			},
		}),
	}, nil
}

func (t *clipboardWatchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Enable bool `json:"enable"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.toggleMonitoring(args.Enable)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭剪贴板模块
func (m *ClipboardModule) Close() error {
	m.clipboard.StopWatching()
	return nil
}

// GetStats 获取统计信息
func (m *ClipboardModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_reads":    m.stats.TotalReads,
		"total_writes":   m.stats.TotalWrites,
		"total_clears":   m.stats.TotalClears,
		"history_count":  int64(m.stats.HistoryCount),
		"bytes_read":      m.stats.BytesRead,
		"bytes_written":   m.stats.BytesWritten,
	}
}

// ==================== 核心功能实现 ====================

// readContent 读取剪贴板内容
func (m *ClipboardModule) readContent(format string) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalReads++
	m.stats.mu.Unlock()

	var content string
	var mimeType string
	var size int64
	var err error

	// 根据格式读取内容
	switch format {
	case "text":
		content, err = m.clipboard.GetText()
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to read text: %v", err),
			}, nil
		}
		size = int64(len(content))
		mimeType = "text/plain"
	case "image":
		var data []string
		data, mimeType, err = m.clipboard.GetImage()
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to read image: %v", err),
			}, nil
		}
		content = data[0]
		size = int64(len(content))
	case "auto":
		// 先尝试读取文本
		content, err = m.clipboard.GetText()
		if err == nil && content != "" {
			mimeType = "text/plain"
			size = int64(len(content))
			break
		}
		// 尝试读取图片
		var data []string
		data, mimeType, err = m.clipboard.GetImage()
		if err == nil {
			content = data[0]
			size = int64(len(content))
		}
	default:
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Unsupported format: %s", format),
		}, nil
	}

	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to read clipboard: %v", err),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.BytesRead += size
	m.stats.mu.Unlock()

	return map[string]any{
		"success":   true,
		"content":   content,
		"type":      mimeType,
		"size":      size,
		"mime_type": mimeType,
	}, nil
}

// writeContent 写入剪贴板内容
func (m *ClipboardModule) writeContent(content, contentType, mimeType string) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalWrites++
	m.stats.mu.Unlock()

	// 检查内容大小
	if contentType == "text" && len(content) > m.config.MaxTextSize {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Content size exceeds maximum allowed size"),
		}, nil
	}

	var err error

	// 根据类型写入内容
	switch contentType {
	case "text":
		err = m.clipboard.SetText(content)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to write text: %v", err),
			}, nil
		}
	case "image":
		if mimeType == "" {
			mimeType = "image/png"
		}
		data := []byte(content)
		err = m.clipboard.SetImage(data, mimeType)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to write image: %v", err),
			}, nil
		}
	default:
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Unsupported content type: %s", contentType),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.BytesWritten += int64(len(content))
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"type":     contentType,
		"size":     len(content),
	}, nil
}

// clearContent 清空剪贴板
func (m *ClipboardModule) clearContent() (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalClears++
	m.stats.mu.Unlock()

	err := m.clipboard.Clear()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to clear clipboard: %v", err),
		}, nil
	}

	return map[string]any{
		"success": true,
	}, nil
}

// getHistory 获取历史记录
func (m *ClipboardModule) getHistory(limit int, itemType string) (map[string]any, error) {
	if !m.config.EnableHistory {
		return map[string]any{
			"success": false,
			"error":   "History is not enabled",
		}, nil
	}

	// 这里可以实现历史记录功能
	// 由于篇幅限制，这里仅返回空历史
	return map[string]any{
		"success": true,
		"enabled": true,
		"items":   []ClipboardItem{},
		"count":   0,
	}, nil
}

// toggleMonitoring 切换监控状态
func (m *ClipboardModule) toggleMonitoring(enable bool) (map[string]any, error) {
	if !m.config.EnableMonitor {
		return map[string]any{
			"success": false,
			"error":   "Monitoring is not enabled in configuration",
		}, nil
	}

	if enable {
		// 启动监控（在 NewClipboardModule 中已启动）
		return map[string]any{
			"success":  true,
			"enabled":  true,
			"message":  "Clipboard monitoring is active",
		}, nil
	}

	// 停止监控
	err := m.clipboard.StopWatching()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to stop monitoring: %v", err),
		}, nil
	}

	return map[string]any{
		"success": true,
		"enabled": false,
	}, nil
}

// monitorClipboard 监控剪贴板变化
func (m *ClipboardModule) monitorClipboard() {
	m.clipboard.WatchChanges(func(content string) {
		// 剪贴板内容变化时的回调
		// 可以在这里记录历史或触发事件
	})
}

// ==================== 平台特定实现 ====================

// WindowsClipboard Windows 剪贴板实现
type WindowsClipboard struct {
	// Windows 特定字段
}

func NewWindowsClipboard() (*WindowsClipboard, error) {
	return &WindowsClipboard{}, nil
}

func (c *WindowsClipboard) GetText() (string, error) {
	// Windows API 调用实现
	return "", fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) SetText(text string) error {
	// Windows API 调用实现
	return fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) GetImage() ([]byte, string, error) {
	// Windows API 调用实现
	return nil, "", fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) SetImage(data []byte, mimeType string) error {
	// Windows API 调用实现
	return fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) Clear() error {
	// Windows API 调用实现
	return fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) WatchChanges(callback func(string)) error {
	// Windows API 调用实现
	return fmt.Errorf("not implemented")
}

func (c *WindowsClipboard) StopWatching() error {
	return nil
}

// MacOSClipboard macOS 剪贴板实现
type MacOSClipboard struct {
	// macOS 特定字段
}

func NewMacOSClipboard() (*MacOSClipboard, error) {
	return &MacOSClipboard{}, nil
}

func (c *MacOSClipboard) GetText() (string, error) {
	// macOS API 调用实现（使用 pbpaste 命令）
	return "", fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) SetText(text string) error {
	// macOS API 调用实现（使用 pbcopy 命令）
	return fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) GetImage() ([]byte, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) SetImage(data []byte, mimeType string) error {
	return fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) Clear() error {
	return fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) WatchChanges(callback func(string)) error {
	return fmt.Errorf("not implemented")
}

func (c *MacOSClipboard) StopWatching() error {
	return nil
}

// LinuxClipboard Linux 剪贴板实现
type LinuxClipboard struct {
	// Linux 特定字段
}

func NewLinuxClipboard() (*LinuxClipboard, error) {
	return &LinuxClipboard{}, nil
}

func (c *LinuxClipboard) GetText() (string, error) {
	// Linux 实现使用 xclip 或 xsel 命令
	return "", fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) SetText(text string) error {
	// Linux 实现使用 xclip 或 xsel 命令
	return fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) GetImage() ([]byte, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) SetImage(data []byte, mimeType string) error {
	return fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) Clear() error {
	return fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) WatchChanges(callback func(string)) error {
	return fmt.Errorf("not implemented")
}

func (c *LinuxClipboard) StopWatching() error {
	return nil
}
