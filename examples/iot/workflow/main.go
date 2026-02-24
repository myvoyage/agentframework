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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/iot"
	"AgentFramework/pkg/iot"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== IoT工作流自动化示例 ===")
	fmt.Println()

	// 1. 创建适配器管理器
	adapterMgr := iot.NewAdapterManager()

	// 2. 创建并初始化适配器
	fmt.Println("1. 初始化适配器...")

	zigbeeAdapter := adapters.NewZigbeeAdapter()
	_ = zigbeeAdapter.Initialize(ctx, iot.ProtocolConfig{
		Type: iot.ProtocolZigbee,
		Hardware: iot.HardwareConfig{
			Type:    "websocket",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"broker_url": "ws://localhost:8000",
		},
	})
	_ = adapterMgr.RegisterAdapter(zigbeeAdapter)
	fmt.Println("  ✓ Zigbee适配器已注册")

	nearlinkAdapter := adapters.NewNearLinkAdapter()
	_ = nearlinkAdapter.Initialize(ctx, iot.ProtocolConfig{
		Type: iot.ProtocolNearLink,
		Metadata: map[string]string{
			"network_mode":   "SLM",
			"multicast_addr": "224.0.0.1:1888",
		},
	})
	_ = adapterMgr.RegisterAdapter(nearlinkAdapter)
	fmt.Println("  ✓ NearLink适配器已注册")
	fmt.Println()

	// 3. 创建工作流引擎
	fmt.Println("2. 创建工作流引擎...")
	engine := iot.NewWorkflowEngine(adapterMgr)
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("启动工作流引擎失败: %v", err)
	}
	fmt.Println("  ✓ 工作流引擎已启动")
	fmt.Println()

	// 4. 示例1：自动化规则 - 传感器触发灯光
	fmt.Println("3. 创建自动化规则...")

	motionLightRule := &iot.AutomationRule{
		ID:      "motion-light-rule",
		Name:    "人体感应自动开灯",
		Enabled: true,
		Triggers: []iot.Trigger{
			{
				Type:  iot.TriggerTypeEvent,
				Event: "motion_detected",
			},
		},
		Conditions: []iot.Condition{
			{
				Type:      iot.ConditionTypeDeviceState,
				DeviceID:  "zigbee-sensor-motion-001",
				Attribute: "motion",
				Value:     "true",
			},
			{
				Type:  iot.ConditionTypeDeviceState,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "state",
				Value:     "off",
			},
		},
		Actions: []iot.Action{
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "state",
				Value:     "on",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "brightness",
				Value:     80,
			},
		},
	}

	if err := engine.RegisterRule(motionLightRule); err != nil {
		log.Printf("注册规则失败: %v", err)
	} else {
		fmt.Println("  ✓ 规则已注册: 人体感应自动开灯")
	}

	// 5. 示例2：定时任务 - 自动关闭灯光
	autoOffRule := &iot.AutomationRule{
		ID:      "auto-off-rule",
		Name:    "定时关灯",
		Enabled: true,
		Triggers: []iot.Trigger{
			{
				Type:  iot.TriggerTypeSchedule,
				Event: "timer",
			},
		},
		Conditions: []iot.Condition{
			{
				Type:  iot.ConditionTypeTime,
				Value: "23:00",
			},
		},
		Actions: []iot.Action{
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "state",
				Value:     "off",
			},
		},
	}

	if err := engine.RegisterRule(autoOffRule); err != nil {
		log.Printf("注册规则失败: %v", err)
	} else {
		fmt.Println("  ✓ 规则已注册: 定时关灯 (23:00)")
	}
	fmt.Println()

	// 6. 示例3：场景 - 晚间模式
	fmt.Println("4. 创建场景...")

	eveningMode := &iot.Scenario{
		ID:          "evening-mode",
		Name:        "晚间模式",
		Description: "调整灯光为舒适的晚间氛围",
		Icon:        "🌙",
		Actions: []iot.Action{
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "state",
				Value:     "on",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "brightness",
				Value:     30,
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "color",
				Value:     "#FF9900",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-rgb-002",
				Attribute: "state",
				Value:     "on",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-rgb-002",
				Attribute: "color",
				Value:     "#FF6600",
			},
		},
	}

	if err := engine.RegisterScenario(eveningMode); err != nil {
		log.Printf("注册场景失败: %v", err)
	} else {
		fmt.Println("  ✓ 场景已注册: 晚间模式")
	}

	// 7. 示例4：场景 - 离家模式
	awayMode := &iot.Scenario{
		ID:          "away-mode",
		Name:        "离家模式",
		Description: "关闭所有设备，启用安防",
		Icon:        "🏠",
		Actions: []iot.Action{
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-bulb-001",
				Attribute: "state",
				Value:     "off",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-rgb-002",
				Attribute: "state",
				Value:     "off",
			},
			{
				Type:      iot.ActionTypeNotification,
				Title:     "离家模式已激活",
				Message:   "所有设备已关闭，安防系统已启用",
			},
		},
	}

	if err := engine.RegisterScenario(awayMode); err != nil {
		log.Printf("注册场景失败: %v", err)
	} else {
		fmt.Println("  ✓ 场景已注册: 离家模式")
	}
	fmt.Println()

	// 8. 示例5：工作流 - 温度监控
	fmt.Println("5. 创建工作流...")

	tempMonitorWorkflow := &iot.Workflow{
		ID:          "temp-monitor-workflow",
		Name:        "温度监控工作流",
		Description: "监控温度，超过阈值时发送通知",
		Enabled:     true,
		Triggers: []iot.Trigger{
			{
				Type:     iot.TriggerTypeSchedule,
				Interval: 300, // 每5分钟
			},
		},
		Conditions: []iot.Condition{
			{
				Type:      iot.ConditionTypeDeviceState,
				DeviceID:  "zigbee-sensor-temp-001",
				Attribute: "temperature",
				Value:     "30", // 超过30度
			},
		},
		Actions: []iot.Action{
			{
				Type:      iot.ActionTypeNotification,
				Title:     "温度警报",
				Message:   "温度超过30度！当前温度: {{temperature}}°C",
			},
			{
				Type:      iot.ActionTypeDeviceControl,
				DeviceID:  "zigbee-ac-001",
				Attribute: "state",
				Value:     "on",
			},
		},
	}

	if err := engine.RegisterWorkflow(tempMonitorWorkflow); err != nil {
		log.Printf("注册工作流失败: %v", err)
	} else {
		fmt.Println("  ✓ 工作流已注册: 温度监控")
	}
	fmt.Println()

	// 9. 执行场景示例
	fmt.Println("6. 执行场景示例...")

	if err := engine.ExecuteScenario(ctx, "evening-mode"); err != nil {
		log.Printf("执行场景失败: %v", err)
	} else {
		fmt.Println("  ✓ 场景执行成功: 晚间模式")
	}
	fmt.Println()

	// 10. 查询所有配置
	fmt.Println("7. 查询配置...")

	rules := engine.ListRules()
	fmt.Printf("  规则数量: %d\n", len(rules))
	for _, rule := range rules {
		enabled := "✓"
		if !rule.Enabled {
			enabled = "✗"
		}
		fmt.Printf("    %s [%s] %s\n", enabled, rule.ID, rule.Name)
	}

	scenarios := engine.ListScenarios()
	fmt.Printf("  场景数量: %d\n", len(scenarios))
	for _, scenario := range scenarios {
		fmt.Printf("    - [%s] %s %s\n", scenario.ID, scenario.Icon, scenario.Name)
	}

	workflows := engine.ListWorkflows()
	fmt.Printf("  工作流数量: %d\n", len(workflows))
	for _, workflow := range workflows {
		enabled := "✓"
		if !workflow.Enabled {
			enabled = "✗"
		}
		fmt.Printf("    %s [%s] %s\n", enabled, workflow.ID, workflow.Name)
	}
	fmt.Println()

	// 11. 导出配置
	fmt.Println("8. 导出配置...")

	config := map[string]interface{}{
		"rules":     rules,
		"scenarios": scenarios,
		"workflows": workflows,
	}

	configJSON, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("配置导出:\n%s\n", string(configJSON))
	fmt.Println()

	// 12. 保持运行一段时间
	fmt.Println("9. 工作流引擎运行中...")
	fmt.Println("   按Ctrl+C停止")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("[%s] 工作流引擎运行正常\n", time.Now().Format("15:04:05"))
		case <-ctx.Done():
			fmt.Println()
			fmt.Println("停止工作流引擎...")
			if err := engine.Stop(ctx); err != nil {
				log.Printf("停止引擎失败: %v", err)
			}
			fmt.Println("✓ 工作流引擎已停止")
			return
		}
	}
}
