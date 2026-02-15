// Agent Framework - System Tools Demo
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/tools/sandbox/sys"
)

func main() {
	// 创建系统工具配置
	config := sys.DefaultSystemToolsConfig()

	// 自定义配置（可选）
	config.NetworkTools.EnableDNSLookup = true
	config.NetworkTools.EnablePortScan = true
	config.ProcessManager.EnableAutoCleanup = true
	config.Notification.EnableSound = true

	// 创建系统工具集合
	systemTools, err := sys.NewSystemTools(config)
	if err != nil {
		log.Fatalf("Failed to create system tools: %v", err)
	}
	defer systemTools.Close()

	ctx := context.Background()

	fmt.Println("=== AgentFramework 系统工具演示 ===")
	fmt.Println("本演示展示如何使用系统工具模块\n")

	// 演示 1: 进程管理
	fmt.Println("--- 进程管理演示 ---")
	demoProcessManager(ctx, systemTools.GetProcessManager())

	// 演示 2: 网络工具
	fmt.Println("\n--- 网络工具演示 ---")
	demoNetworkTools(ctx, systemTools.GetNetworkTools())

	// 演示 3: 剪贴板（注释掉，因为需要平台特定实现）
	// fmt.Println("\n--- 剪贴板演示 ---")
	// demoClipboard(ctx, systemTools.GetClipboard())

	// 演示 4: 系统通知（注释掉，因为需要平台特定实现）
	// fmt.Println("\n--- 系统通知演示 ---")
	// demoNotification(ctx, systemTools.GetNotification())

	// 演示 5: 获取所有统计信息
	fmt.Println("\n--- 统计信息 ---")
	stats := systemTools.GetAllStats()
	for module, moduleStats := range stats {
		fmt.Printf("%s:\n", module)
		for key, value := range moduleStats {
			fmt.Printf("  %s: %d\n", key, value)
		}
	}

	fmt.Println("\n演示完成！")
}

// demoProcessManager 演示进程管理功能
func demoProcessManager(ctx context.Context, pm *sys.ProcessManagerModule) {
	// 获取系统信息
	fmt.Println("获取系统信息...")
	info, err := pm.getSystemInfo()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if sysInfo, ok := info["system"].(map[string]any); ok {
		fmt.Printf("操作系统: %v\n", sysInfo["os"])
		fmt.Printf("架构: %v\n", sysInfo["arch"])
		fmt.Printf("CPU 核心数: %v\n", sysInfo["cpu_count"])
	}

	// 列出进程
	fmt.Println("\n列出前 10 个进程...")
	result, err := pm.listProcesses("", 10)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if success, ok := result["success"].(bool); ok && success {
		if processes, ok := result["processes"].([]sys.ProcessInfo); ok {
			fmt.Printf("找到 %d 个进程:\n", len(processes))
			for i, p := range processes {
				if i >= 5 { // 只显示前5个
					break
				}
				fmt.Printf("  [%d] PID=%d Name=%s\n", i+1, p.PID, p.Name)
			}
		}
	}
}

// demoNetworkTools 演示网络工具功能
func demoNetworkTools(ctx context.Context, nt *sys.NetworkToolsModule) {
	// DNS 查询
	fmt.Println("DNS 查询 example.com...")
	result, err := nt.dnsLookup("example.com", "A")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if success, ok := result["success"].(bool); ok && success {
		fmt.Printf("DNS 查询成功\n")
		if records, ok := result["records"].([]string); ok && len(records) > 0 {
			fmt.Printf("  IP 地址: %s\n", records[0])
		}
	}

	// Ping 测试
	fmt.Println("\nPing 测试 cloudflare.com (1次)...")
	pingResult, err := nt.pingHost("1.1.1.1", 1)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if success, ok := pingResult["success"].(bool); ok && success {
		fmt.Printf("Ping 完成\n")
		if results, ok := pingResult["results"].([]sys.PingResult); ok && len(results) > 0 {
			r := results[0]
			if r.Success {
				fmt.Printf("  响应时间: %dms\n", r.TimeMs)
			} else {
				fmt.Printf("  请求失败: %s\n", r.Error)
			}
		}
	}

	// HTTP 请求
	fmt.Println("\nHTTP �试 https://httpbin.org/get...")
	httpResult, err := nt.makeRequest("GET", "https://httpbin.org/get", nil, "", 5000)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if success, ok := httpResult["success"].(bool); ok && success {
		fmt.Printf("HTTP 请求成功\n")
		if statusCode, ok := httpResult["status_code"].(int); ok {
			fmt.Printf("  状态码: %d\n", statusCode)
		}
		if timeMs, ok := httpResult["time_ms"].(int64); ok {
			fmt.Printf("  响应时间: %dms\n", timeMs)
		}
	}
}

// demoClipboard 演示剪贴板功能（需要平台实现）
func demoClipboard(ctx context.Context, cb *sys.ClipboardModule) {
	fmt.Println("写入剪贴板...")
	result, err := cb.writeContent("Hello from AgentFramework!", "text", "")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if success, ok := result["success"].(bool); ok && success {
		fmt.Println("写入成功")

		// 读取剪贴板
		fmt.Println("读取剪贴板...")
		readResult, err := cb.readContent("text")
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		if success, ok := readResult["success"].(bool); ok && success {
			if content, ok := readResult["content"].(string); ok {
				fmt.Printf("剪贴板内容: %s\n", content)
			}
		}
	}
}

// demoNotification 演示通知功能（需要平台实现）
func demoNotification(ctx context.Context, nf *sys.NotificationModule) {
	fmt.Println("发送系统通知...")

	// 获取通知能力
	caps := nf.getCapabilities()
	fmt.Printf("通知能力: %+v\n", caps)

	// 发送测试通知（当平台实现后可用）
	// result, err := nf.sendNotification(struct {
	// 	Title    string
	// 	Body     string
	// 	Icon     string
	// 	Sound    string
	// 	Category string
	// 	Priority string
	// 	Timeout  int
	// 	Actions  []sys.NotificationAction
	// }{
	// 	Title:    "AgentFramework",
	// 	Body:     "系统工具演示通知",
	// 	Category: "info",
	// 	Priority: "normal",
	// })
	// if err != nil {
	// 	log.Printf("Error: %v\n", err)
	// 	return
	// }
	// if result.Success {
	// 	fmt.Println("通知发送成功")
	// }
}
