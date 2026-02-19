# AgentFramework IoT 快速入门指南

欢迎使用 AgentFramework IoT 协议支持！本指南将帮助您快速开始使用 IoT 设备管理和自动化功能。

## 目录

1. [安装](#安装)
2. [快速开始](#快速开始)
3. [基础概念](#基础概念)
4. [协议选择](#协议选择)
5. [常见用例](#常见用例)
6. [故障排查](#故障排查)
7. [进阶主题](#进阶主题)
8. [资源](#资源)

## 安装

### 前置要求

- Go 1.21 或更高版本
- 适当的硬件支持（USB适配器、网关等）

### 安装步骤

```bash
# 克隆仓库
git clone https://github.com/myvoyage/agentframework.git
cd agentframework

# 下载依赖
go mod download

# 构建项目
go build ./...
```

## 快速开始

### 5分钟入门

#### 1. 创建第一个IoT程序

创建文件 `main.go`:

```go
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

	// 1. 创建适配器
	adapter := adapters.NewZigbeeAdapter()

	// 2. 初始化
	config := iot.ProtocolConfig{
		Type: iot.ProtocolZigbee,
		Hardware: iot.HardwareConfig{
			Type:    "websocket",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"broker_url": "ws://localhost:8000/mqtt",
		},
	}

	err := adapter.Initialize(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 启动适配器
	err = adapter.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer adapter.Stop(ctx)

	// 4. 发现设备
	fmt.Println("发现设备...")
	devices, err := adapter.DiscoverDevices(ctx, 10*time.Second)
	if err != nil {
		log.Printf("发现设备失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个设备\n", len(devices))

	// 5. 控制设备
	for _, device := range devices {
		fmt.Printf("设备: %s (%s)\n", device.Name, device.ID)

		// 获取设备实例
		dev, err := adapter.GetDevice(ctx, device.ID)
		if err != nil {
			continue
		}

		// 读取状态
		state, _ := dev.Read(ctx, "state")
		fmt.Printf("  状态: %v\n", state)

		// 如果是灯泡，尝试控制
		if device.Type == iot.DeviceTypeActuator {
			// 打开设备
			_ = dev.Write(ctx, "state", "on")
			fmt.Println("  ✓ 设备已打开")
		}
	}
}
```

#### 2. 运行程序

```bash
go run main.go
```

## 基础概念

### IoT设备

所有IoT设备都实现了 `IoTDevice` 接口，提供统一的操作方法：

```go
type IoTDevice interface {
	// 获取设备信息
	ID() string
	Name() string
	Type() DeviceType
	Protocol() ProtocolType

	// 连接管理
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	// 读写属性
	Read(ctx context.Context, attribute string) (interface{}, error)
	Write(ctx context.Context, attribute string, value interface{}) error

	// 事件订阅
	Subscribe(ctx context.Context, events []string, handler DeviceEventHandler) error
}
```

### 协议适配器

每个协议（Zigbee、Thread、Z-Wave、NearLink）都有一个适配器实现 `ProtocolAdapter` 接口。

**主要功能：**
- 设备发现
- 设备配对
- 设备管理
- 网络管理

## 协议选择

### 选择指南

| 使用场景 | 推荐协议 | 原因 |
|---------|---------|------|
| 智能家居基础 | Zigbee | 成熟生态，低功耗 |
| 需要互联网集成 | Thread | 原生IPv6支持 |
| 长距离覆盖 | Z-Wave | Sub-GHz频段，穿透力强 |
| 实时音视频 | NearLink | 超低延迟，高带宽 |
| 电池传感器 | NearLink (SLE) | 极低功耗 |
| 工业控制 | NearLink (SLM) | 低延迟，高可靠 |

### 对比表

| 特性 | Zigbee | Thread | Z-Wave | NearLink |
|------|--------|--------|--------|----------|
| 功耗 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 延迟 | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 距离 | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 带宽 | ⭐⭐ | ⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| 生态 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

## 常见用例

### 用例1：智能灯光控制

```go
// 获取灯泡设备
bulb, _ := adapter.GetDevice(ctx, "zigbee-bulb-001")

// 打开灯光
bulb.Write(ctx, "state", "on")

// 设置亮度
bulb.Write(ctx, "brightness", 80)

// 设置颜色（RGB灯）
bulb.Write(ctx, "color", "#FF0000")
```

### 用例2：传感器监控

```go
// 获取温度传感器
sensor, _ := adapter.GetDevice(ctx, "zigbee-sensor-temp-001")

// 读取温度
temp, _ := sensor.Read(ctx, "temperature")
fmt.Printf("温度: %v°C\n", temp)

// 订阅温度变化
sensor.Subscribe(ctx, []string{"attribute_changed"},
	func(event iot.DeviceEvent) {
		newTemp := event.Data["temperature"]
		fmt.Printf("温度变化: %v°C\n", newTemp)
	},
)
```

### 用例3：自动化场景

```go
// 创建"回家模式"场景
scenario := &iot.Scenario{
	ID:   "home-mode",
	Name: "回家模式",
	Actions: []iot.Action{
		{Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-bulb-001", Attribute: "state", Value: "on"},
		{Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-ac-001", Attribute: "state", Value: "on"},
		{Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-ac-001", Attribute: "temperature", Value: 24},
	},
}

// 注册场景
engine.RegisterScenario(scenario)

// 执行场景
engine.ExecuteScenario(ctx, "home-mode")
```

### 用例4：自动化规则

```go
// 人体感应自动开灯
rule := &iot.AutomationRule{
	ID:      "motion-light",
	Name:    "人体感应自动开灯",
	Enabled: true,
	Triggers: []iot.Trigger{
		{Type: iot.TriggerTypeEvent, Event: "motion_detected"},
	},
	Actions: []iot.Action{
		{Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-bulb-001", Attribute: "state", Value: "on"},
		{Type: iot.ActionTypeDelay, Value: 30 * time.Second},
		{Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-bulb-001", Attribute: "state", Value: "off"},
	},
}

engine.RegisterRule(rule)
```

## 故障排查

### 问题1：设备未发现

**症状：** `DiscoverDevices` 返回空列表

**可能原因：**
1. 适配器未启动
2. 设备未处于配对模式
3. 网络配置错误

**解决方案：**
```go
// 检查适配器状态
if !adapter.IsRunning() {
    log.Fatal("适配器未启动")
}

// 检查网络配置
info, _ := adapter.GetNetworkInfo(ctx)
fmt.Printf("网络状态: %s\n", info.Status)

// 确认设备处于配对模式
// 按照设备说明书操作，通常需要长按按钮
```

### 问题2：设备控制失败

**症状：** `Write` 返回错误

**可能原因：**
1. 设备离线
2. 属性名称错误
3. 值类型不匹配

**解决方案：**
```go
// 检查设备状态
device, _ := adapter.GetDevice(ctx, deviceID)
if !device.IsConnected() {
    log.Fatal("设备未连接")
}

// 读取设备信息
info := device.GetInfo()
fmt.Printf("设备状态: %s\n", info.Status)

// 检查属性是否支持
capabilities := device.GetCapabilities()
fmt.Printf("设备能力: %v\n", capabilities)
```

### 问题3：事件未触发

**症状：** 订阅的事件处理器未被调用

**可能原因：**
1. 事件名称错误
2. 订阅后设备未发送事件
3. 处理器函数有错误

**解决方案：**
```go
// 订阅多个事件以测试
device.Subscribe(ctx, []string{"state_changed", "attribute_changed", "error"},
	func(event iot.DeviceEvent) {
		log.Printf("收到事件: %s, 数据: %v, 错误: %v",
			event.Type, event.Data, event.Error)
	},
)

// 添加错误处理
device.Subscribe(ctx, []string{"*"},
	func(event iot.DeviceEvent) {
		if event.Error != nil {
			log.Printf("事件错误: %v", event.Error)
		}
	},
)
```

## 进阶主题

### 1. 批量操作

使用批量操作提高效率：

```go
// 批量读取（需要设备支持BatchReader接口）
if batchDevice, ok := device.(iot.BatchReader); ok {
    attributes := []string{"temp", "humidity", "pressure"}
    values, _ := batchDevice.BatchRead(ctx, attributes)
    fmt.Printf("批量读取结果: %v\n", values)
}

// 批量写入
if batchDevice, ok := device.(iot.BatchWriter); ok {
    values := map[string]interface{}{
        "brightness": 80,
        "color":      "#FF0000",
    }
    _ = batchDevice.BatchWrite(ctx, values)
}
```

### 2. 设备诊断

使用诊断功能检查设备健康：

```go
// Ping设备（需要设备支持Pingable接口）
if pingable, ok := device.(iot.Pingable); ok {
    rtt, _ := pingable.Ping(ctx)
    fmt.Printf("设备延迟: %v\n", rtt)
}

// 获取诊断信息（需要设备支持Diagnosticable接口）
if diagnosticable, ok := device.(iot.Diagnosticable); ok {
    info, _ := diagnosticable.GetDiagnosticInfo(ctx)
    fmt.Printf("诊断信息: %v\n", info)
}
```

### 3. 工作流编排

创建复杂的自动化逻辑：

```go
// 工作流示例：温度控制
workflow := &iot.Workflow{
	ID:      "temp-control",
	Name:    "温度控制工作流",
	Enabled: true,
	Triggers: []iot.Trigger{
		{Type: iot.TriggerTypeSchedule, Interval: 300}, // 每5分钟
	},
	Conditions: []iot.Condition{
		{
			Type:      iot.ConditionTypeDeviceState,
			DeviceID:  "sensor-temp-001",
			Attribute: "temperature",
			Value:     "28", // 超过28度
		},
		{
			Type:  iot.ConditionTypeTime,
			Value: "09:00-18:00", // 仅在白天
		},
	},
	Actions: []iot.Action{
		{
			Type:      iot.ActionTypeNotification,
			Title:     "温度过高",
			Message:   "开启空调降温",
		},
		{
			Type:      iot.ActionTypeDeviceControl,
			DeviceID:  "ac-001",
			Attribute: "state",
			Value:     "on",
		},
		{
			Type:      iot.ActionTypeDeviceControl,
			DeviceID:  "ac-001",
			Attribute: "temperature",
			Value:     24,
		},
	},
}

engine.RegisterWorkflow(workflow)
```

### 4. 多协议管理

统一管理多个协议：

```go
// 创建适配器管理器
mgr := iot.NewAdapterManager()

// 注册多个协议适配器
zigbeeAdapter := adapters.NewZigbeeAdapter()
threadAdapter := adapters.NewThreadAdapter()
nearlinkAdapter := adapters.NewNearLinkAdapter()

mgr.RegisterAdapter(zigbeeAdapter)
mgr.RegisterAdapter(threadAdapter)
mgr.RegisterAdapter(nearlinkAdapter)

// 列出所有设备
allDevices, _ := mgr.ListDevices(ctx)
fmt.Printf("总设备数: %d\n", len(allDevices))

// 跨协议访问设备
device, _ := mgr.GetDevice(ctx, "zigbee-bulb-001")
device.Write(ctx, "state", "on")

device, _ = mgr.GetDevice(ctx, "thread-sensor-001")
temp, _ := device.Read(ctx, "temperature")
```

## 最佳实践

### 1. 错误处理

始终处理错误并提供有意义的反馈：

```go
// ✅ 好的做法
device, err := adapter.GetDevice(ctx, deviceID)
if err != nil {
    log.Printf("获取设备失败: %s, 错误: %v", deviceID, err)
    return
}

state, err := device.Read(ctx, "state")
if err != nil {
    log.Printf("读取状态失败: %v", err)
    // 尝试重试或使用默认值
    state = "unknown"
}

// ❌ 不好的做法
device, _ := adapter.GetDevice(ctx, deviceID)
state, _ := device.Read(ctx, "state")
```

### 2. 资源清理

正确清理资源：

```go
// ✅ 使用defer确保清理
adapter.Start(ctx)
defer adapter.Stop(ctx)

// 或者使用上下文取消
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

devices, err := adapter.DiscoverDevices(ctx, 10*time.Second)
```

### 3. 日志记录

记录关键操作：

```go
// ✅ 记录设备操作
log.Printf("设备 %s: 写入属性 %s = %v", device.ID(), attribute, value)

// ✅ 记录状态变化
log.Printf("设备 %s: 状态 %s -> %s", device.ID(), oldState, newState)

// ✅ 记录错误
log.Printf("设备 %s: 错误 %v", device.ID(), err)
```

### 4. 性能优化

使用批量操作和缓存：

```go
// ✅ 批量读取
attributes := []string{"temp", "humidity", "pressure"}
values, _ := batchDevice.BatchRead(ctx, attributes)

// ✅ 缓存设备列表
devices, _ := adapter.ListDevices(ctx)
deviceCache := make(map[string]iot.IoTDevice)
for _, device := range devices {
    deviceCache[device.ID()] = device
}

// ✅ 使用事件而非轮询
device.Subscribe(ctx, []string{"state_changed"}, handler)
```

## 资源

### 文档

- [完整文档](README.md) - 所有IoT功能的总览
- [Zigbee适配器](ZIGBEE_ADAPTER.md) - Zigbee协议详解
- [Thread适配器](THREAD_ADAPTER.md) - Thread协议详解
- [Z-Wave适配器](ZWAVE_ADAPTER.md) - Z-Wave协议详解
- [NearLink适配器](NEARLINK_ADAPTER.md) - NearLink协议详解
- [MCP工具](IOT_MCP_TOOLS.md) - MCP工具集成
- [工作流自动化](IOT_WORKFLOW.md) - 工作流引擎详解

### 示例代码

- [Zigbee示例](../../examples/iot/zigbee_example.go)
- [Thread示例](../../examples/iot/thread_example.go)
- [Z-Wave示例](../../examples/iot/zwave_example.go)
- [NearLink示例](../../examples/iot/nearlink_example.go)
- [工作流示例](../../examples/iot/workflow_example.go)

### 测试

```bash
# 运行所有IoT测试
go test ./pkg/iot/... -v

# 运行特定协议测试
go test ./pkg/iot/adapters -run TestZigbee -v
go test ./pkg/iot/adapters -run TestThread -v
go test ./pkg/iot/adapters -run TestZWave -v
go test ./pkg/iot/adapters -run TestNearLink -v

# 运行性能测试
go test ./pkg/iot -bench=. -benchmem
```

### 获取帮助

- GitHub Issues: [报告问题](https://github.com/myvoyage/agentframework/issues)
- 讨论区: [GitHub Discussions](https://github.com/myvoyage/agentframework/discussions)
- 邮件: support@example.com

## 下一步

现在您已经了解了基础，可以：

1. 📖 阅读[完整文档](README.md)深入了解
2. 💻 运行[示例代码](../../examples/iot/)学习更多
3. 🔧 查看[API参考](IOT_MCP_TOOLS.md)了解所有可用功能
4. 🚀 开始构建自己的IoT应用！

祝您使用愉快！🎉

---

**最后更新**: 2026-02-19
**版本**: v1.0.0
**作者**: AgentFramework Team
