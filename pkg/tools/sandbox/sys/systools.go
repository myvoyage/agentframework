// Agent Framework - System Tools Registry
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

// SystemTools 系统工具集合
type SystemTools struct {
	processManager  *ProcessManagerModule
	networkTools   *NetworkToolsModule
	clipboard      *ClipboardModule
	notification   *NotificationModule
}

// SystemToolsConfig 系统工具配置
type SystemToolsConfig struct {
	ProcessManager  ProcessConfig    `json:"process_manager"`
	NetworkTools   NetworkConfig    `json:"network_tools"`
	Clipboard       ClipboardConfig  `json:"clipboard"`
	Notification    NotificationConfig `json:"notification"`
}

// DefaultSystemToolsConfig 返回默认配置
func DefaultSystemToolsConfig() *SystemToolsConfig {
	return &SystemToolsConfig{
		ProcessManager: ProcessConfig{
			MaxProcesses:      1000,
			EnableAutoCleanup:  true,
			CleanupInterval:    300,
			EnableMonitoring:   false,
		},
		NetworkTools: NetworkConfig{
			Timeout:           30000,
			MaxRedirects:      10,
			UserAgent:         "AgentFramework/1.0",
			EnableDNSLookup:   true,
			EnablePortScan:    true,
			MaxPortScanPorts: 1024,
		},
		Clipboard: ClipboardConfig{
			MaxHistorySize:   100,
			EnableHistory:    false,
			EnableMonitor:    false,
			AllowedFormats:   []string{"text"},
			MaxTextSize:      1024 * 1024,
			EnableEncryption: false,
		},
		Notification: NotificationConfig{
			MaxQueueSize:      100,
			EnableSound:       true,
			DefaultSound:      "default",
			DefaultIcon:       "information",
			EnablePersistence: false,
			EnableGrouping:    false,
			EnableActions:     false,
		},
	}
}

// NewSystemTools 创建系统工具集合
func NewSystemTools(config *SystemToolsConfig) (*SystemTools, error) {
	if config == nil {
		config = DefaultSystemToolsConfig()
	}

	// 创建进程管理模块
	processManager, err := NewProcessManagerModule(config.ProcessManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create process manager: %w", err)
	}

	// 创建网络工具模块
	networkTools, err := NewNetworkToolsModule(config.NetworkTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create network tools: %w", err)
	}

	// 创建剪贴板模块
	clipboard, err := NewClipboardModule(config.Clipboard)
	if err != nil {
		return nil, fmt.Errorf("failed to create clipboard: %w", err)
	}

	// 创建通知模块
	notification, err := NewNotificationModule(config.Notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &SystemTools{
		processManager: processManager,
		networkTools:   networkTools,
		clipboard:      clipboard,
		notification:   notification,
	}, nil
}

// GetAllTools 获取所有系统工具
func (st *SystemTools) GetAllTools(ctx context.Context) ([]tool.BaseTool, error) {
	var allTools []tool.BaseTool

	// 进程管理工具
	processTools, err := st.processManager.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get process manager tools: %w", err)
	}
	allTools = append(allTools, processTools...)

	// 网络工具
	networkTools, err := st.networkTools.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network tools: %w", err)
	}
	allTools = append(allTools, networkTools...)

	// 剪贴板工具
	clipboardTools, err := st.clipboard.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get clipboard tools: %w", err)
	}
	allTools = append(allTools, clipboardTools...)

	// 通知工具
	notificationTools, err := st.notification.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification tools: %w", err)
	}
	allTools = append(allTools, notificationTools...)

	return allTools, nil
}

// GetProcessManager 获取进程管理模块
func (st *SystemTools) GetProcessManager() *ProcessManagerModule {
	return st.processManager
}

// GetNetworkTools 获取网络工具模块
func (st *SystemTools) GetNetworkTools() *NetworkToolsModule {
	return st.networkTools
}

// GetClipboard 获取剪贴板模块
func (st *SystemTools) GetClipboard() *ClipboardModule {
	return st.clipboard
}

// GetNotification 获取通知模块
func (st *SystemTools) GetNotification() *NotificationModule {
	return st.notification
}

// Close 关闭所有系统工具
func (st *SystemTools) Close() error {
	var errs []error

	if err := st.processManager.Close(); err != nil {
		errs = append(errs, fmt.Errorf("process manager close error: %w", err))
	}

	if err := st.networkTools.Close(); err != nil {
		errs = append(errs, fmt.Errorf("network tools close error: %w", err))
	}

	if err := st.clipboard.Close(); err != nil {
		errs = append(errs, fmt.Errorf("clipboard close error: %w", err))
	}

	if err := st.notification.Close(); err != nil {
		errs = append(errs, fmt.Errorf("notification close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing system tools: %v", errs)
	}

	return nil
}

// GetAllStats 获取所有模块的统计信息
func (st *SystemTools) GetAllStats() map[string]map[string]int64 {
	stats := make(map[string]map[string]int64)

	stats["process_manager"] = st.processManager.GetStats()
	stats["network_tools"] = st.networkTools.GetStats()
	stats["clipboard"] = st.clipboard.GetStats()
	stats["notification"] = st.notification.GetStats()

	return stats
}
