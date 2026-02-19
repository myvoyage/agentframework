// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/iot"
	"AgentFramework/pkg/iot/adapters"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== NearLink (星闪) 协议示例程序 ===")
	fmt.Println()

	// 1. 创建NearLink适配器
	fmt.Println("1. 创建NearLink适配器...")
	adapter := adapters.NewNearLinkAdapter()

	// 2. 配置适配器
	fmt.Println("2. 配置适配器...")
	config := iot.ProtocolConfig{
		Type: iot.ProtocolNearLink,
		Hardware: iot.HardwareConfig{
			Type:    "udp",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"network_mode":    "SLM", // Spark Link Mesh mode
			"multicast_addr":  "224.0.0.1:1888",
			"channel":         "0",  // 2.4GHz
			"mesh_id":         "12345678",
		},
	}

	err := adapter.Initialize(ctx, config)
	if err != nil {
		log.Fatalf("初始化适配器失败: %v", err)
	}
	fmt.Println("   ✓ 适配器初始化成功")

	// 3. 启动适配器
	fmt.Println("3. 启动适配器...")
	err = adapter.Start(ctx)
	if err != nil {
		log.Fatalf("启动适配器失败: %v", err)
	}
	fmt.Println("   ✓ 适配器已启动")
	fmt.Println()

	// 4. 获取网络信息
	fmt.Println("4. 获取NearLink网络信息...")
	networkInfo, err := adapter.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("   获取网络信息失败: %v", err)
	} else {
		fmt.Printf("   ✓ Mesh ID: 0x%X\n", networkInfo.PanID)
		fmt.Printf("   ✓ 频段: %s\n", networkInfo.Properties["frequency"])
		fmt.Printf("   ✓ 网络模式: %s\n", networkInfo.Properties["network_mode"])
		fmt.Printf("   ✓ 设备数量: %d\n", networkInfo.DeviceCount)
		fmt.Printf("   ✓ 状态: %s\n", networkInfo.Status)
	}
	fmt.Println()

	// 5. 发现设备
	fmt.Println("5. 发现NearLink设备...")
	devices, err := adapter.DiscoverDevices(ctx, 10*time.Second)
	if err != nil {
		log.Printf("   发现设备失败: %v", err)
	} else {
		fmt.Printf("   ✓ 发现 %d 个设备:\n", len(devices))
		for i, device := range devices {
			fmt.Printf("     %d. %s (%s)\n", i+1, device.Name, device.ID)
			fmt.Printf("        类型: %s\n", device.Type)
			fmt.Printf("        状态: %s\n", device.Status)
		}
	}
	fmt.Println()

	// 6. 配对新设备（示例）
	fmt.Println("6. 配对新设备...")
	fmt.Println("   提示: 请将NearLink设备设置为配对模式...")

	pairingResult, err := adapter.StartPairing(ctx, 60*time.Second)
	if err != nil {
		log.Printf("   配对失败: %v", err)
	} else if pairingResult.Success {
		fmt.Printf("   ✓ 设备配对成功!\n")
		fmt.Printf("     设备ID: %s\n", pairingResult.Device.ID)
		fmt.Printf("     设备名称: %s\n", pairingResult.Device.Name)
		fmt.Printf("     MAC地址: %s\n", pairingResult.Device.Properties["mac_address"])
	} else {
		fmt.Printf("   ✗ 配对失败: %s\n", pairingResult.Error)
	}
	fmt.Println()

	// 7. 列出所有设备
	fmt.Println("7. 列出所有已配对设备...")
	allDevices, err := adapter.ListDevices(ctx)
	if err != nil {
		log.Printf("   获取设备列表失败: %v", err)
	} else {
		fmt.Printf("   ✓ 共有 %d 个设备:\n", len(allDevices))
		for _, device := range allDevices {
			fmt.Printf("     - %s\n", device.ID())
		}
	}
	fmt.Println()

	// 8. 控制设备（示例）
	if len(allDevices) > 0 {
		device := allDevices[0]
		fmt.Printf("8. 控制设备: %s\n", device.ID())

		// 读取设备属性
		state, err := device.Read(ctx, "state")
		if err != nil {
			log.Printf("   读取状态失败: %v", err)
		} else {
			fmt.Printf("   当前状态: %v\n", state)
		}

		// 如果是NearLink设备，使用特殊命令
		if nearlinkDev, ok := device.(*adapters.NearLinkDevice); ok {
			// 获取NearLink特定信息
			nearlinkInfo, err := nearlinkDev.GetNearLinkInfo(ctx)
			if err == nil {
				fmt.Printf("   RSSI: %v dBm\n", nearlinkInfo["rssi"])
				fmt.Printf("   电池电量: %v%%\n", nearlinkInfo["battery_level"])
			}

			// 发送NearLink命令
			result, err := nearlinkDev.SendNearLinkCommand(ctx, "get_config", nil)
			if err == nil {
				fmt.Printf("   命令结果: %v\n", result)
			}

			// Ping测试
			rtt, err := nearlinkDev.Ping(ctx)
			if err == nil {
				fmt.Printf("   延迟: %v ms\n", rtt.Milliseconds())
			}

			// 获取诊断信息
			diag, err := nearlinkDev.GetDiagnosticInfo(ctx)
			if err == nil {
				fmt.Printf("   诊断信息:\n")
				fmt.Printf("     连接状态: %v\n", diag["connection"])
				if ping, ok := diag["ping_ms"]; ok {
					fmt.Printf("     Ping: %v ms\n", ping)
				}
			}
		}

		// 批量读取
		attributes := []string{"state", "level", "battery"}
		batchResult, err := device.(*adapters.NearLinkDevice).BatchRead(ctx, attributes)
		if err == nil {
			fmt.Printf("   批量读取结果:\n")
			for attr, value := range batchResult {
				fmt.Printf("     %s: %v\n", attr, value)
			}
		}
	}
	fmt.Println()

	// 9. 订阅设备事件
	if len(allDevices) > 0 {
		device := allDevices[0]
		fmt.Printf("9. 订阅设备事件: %s\n", device.ID())

		err := device.Subscribe(ctx, []string{"state_changed", "data_received"}, func(event iot.DeviceEvent) {
			log.Printf("收到事件: Type=%s, Data=%v", event.Type, event.Data)
		})
		if err != nil {
			log.Printf("   订阅事件失败: %v", err)
		} else {
			fmt.Println("   ✓ 事件订阅成功")

			// 等待一段时间接收事件
			fmt.Println("   等待事件...")
			time.Sleep(5 * time.Second)
		}
	}
	fmt.Println()

	// 10. 数据流示例（用于传感器）
	if len(allDevices) > 0 {
		device := allDevices[0].(*adapters.NearLinkDevice)
		fmt.Println("10. 创建数据流 (读取传感器数据)...")

		stream, err := device.Stream(ctx, "temperature", 1*time.Second)
		if err != nil {
			log.Printf("   创建数据流失败: %v", err)
		} else {
			fmt.Println("   ✓ 数据流已创建")
			fmt.Println("   接收数据:")
			timeout := time.After(5 * time.Second)
			for {
				select {
				case data, ok := <-stream:
					if !ok {
						fmt.Println("   数据流已关闭")
						goto streamEnd
					}
					fmt.Printf("     温度: %v°C\n", data)
				case <-timeout:
					fmt.Println("   数据流超时")
					goto streamEnd
				}
			}
		streamEnd:
		}
	}
	fmt.Println()

	// 11. 清理资源
	fmt.Println("11. 清理资源...")
	err = adapter.Stop(ctx)
	if err != nil {
		log.Printf("   停止适配器失败: %v", err)
	} else {
		fmt.Println("   ✓ 适配器已停止")
	}

	fmt.Println()
	fmt.Println("=== NearLink示例程序结束 ===")
}
