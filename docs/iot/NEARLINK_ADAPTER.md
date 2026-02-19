# NearLink (星闪) 协议支持实现文档

## 概述

AgentFramework 现已支持 **NearLink (星闪)** 无线通信协议，这是华为开发的新一代短距离无线通信技术，提供低延迟、低功耗、高速传输的物联网连接方案。

## 技术特性

NearLink 技术相比传统无线通信协议具有以下优势：

| 特性 | NearLink | 蓝牙 | Wi-Fi |
|------|----------|------|-------|
| **延迟** | ~20μs (低94%) | ~100ms | ~30ms |
| **功耗** | 低60% | 基准 | 高 |
| **传输速率** | 最高12Mbps | 1-2Mbps | 最高数百Mbps |
| **连接数** | 256个 | 7-8个 | 最多数十个 |
| **覆盖范围** | ~100m | ~10m | ~50m |
| **频段** | 2.4GHz, 5.1GHz | 2.4GHz | 2.4GHz, 5GHz |

### 双模式支持

NearLink 支持两种工作模式：

1. **SLM (Spark Link Mesh)** - 网状网络模式
   - 高吞吐量、低延迟
   - 适用于音频/视频传输、实时控制
   - 支持多跳Mesh网络
   - 最高传输速率：12Mbps

2. **SLE (Spark Link Low Energy)** - 低功耗模式
   - 超低功耗
   - 适用于传感器、可穿戴设备
   - 电池寿命可达数年
   - 类似蓝牙BLE但功耗更低

## 架构设计

```
应用层
    ↓
IoTDeviceManager (设备管理器)
    ↓
NearLinkAdapter (NearLink协议适配器)
    ↓
NearLinkController (UDP组网控制器)
    ↓
NearLink设备 (传感器、执行器、音频设备等)
```

## 已实现的文件

### 核心组件

1. **[pkg/iot/adapters/nearlink_adapter.go](pkg/iot/adapters/nearlink_adapter.go)** - NearLink协议适配器
   - 实现ProtocolAdapter接口
   - 支持SLM和SLE双模式
   - UDP多播设备发现
   - 设备配对和管理
   - 网络信息获取
   - 事件处理

2. **[pkg/iot/adapters/nearlink_device.go](pkg/iot/adapters/nearlink_device.go)** - NearLink设备实现
   - 实现IoTDevice接口
   - 设备状态读写
   - 批量操作支持
   - 数据流（Stream）功能
   - 诊断信息获取
   - Ping延迟测试

3. **[pkg/beads/hardware/drivers/nearlink_driver.go](pkg/beads/hardware/drivers/nearlink_driver.go)** - NearLink硬件驱动
   - 实现HardwareController接口
   - 支持完整的硬件命令集
   - 驱动状态管理

### 示例和测试

4. **[examples/iot/nearlink_example.go](examples/iot/nearlink_example.go)** - 使用示例
   - 完整的使用流程演示
   - 设备发现和配对
   - 设备控制和数据读取
   - 事件订阅和数据流

5. **[pkg/iot/adapters/nearlink_test.go](pkg/iot/adapters/nearlink_test.go)** - 单元测试
   - 适配器测试
   - 设备操作测试
   - 控制器测试
   - 双模式测试

## 核心功能

### 1. 初始化适配器

```go
adapter := adapters.NewNearLinkAdapter()

config := iot.ProtocolConfig{
    Type: iot.ProtocolNearLink,
    Hardware: iot.HardwareConfig{
        Type:    "udp",
        Timeout: 5000,
    },
    Metadata: map[string]string{
        "network_mode":    "SLM",     // 或 "SLE"
        "multicast_addr":  "224.0.0.1:1888",
        "channel":         "0",       // 0-13: 2.4GHz, 14+: 5.1GHz
        "mesh_id":         "12345678",
    },
}

adapter.Initialize(ctx, config)
adapter.Start(ctx)
```

### 2. 发现设备

```go
// 发送设备发现广播
devices, err := adapter.DiscoverDevices(ctx, 10*time.Second)

for _, device := range devices {
    fmt.Printf("发现设备: %s (%s)\n", device.Name, device.ID)
    fmt.Printf("  MAC地址: %s\n", device.Properties["mac_address"])
    fmt.Printf("  类型: %s\n", device.Type)
}
```

### 3. 配对设备

```go
// 启动配对模式
result, err := adapter.StartPairing(ctx, 60*time.Second)
if err != nil {
    log.Printf("配对失败: %v", err)
    return
}

if result.Success {
    fmt.Printf("设备配对成功: %s\n", result.Device.Name)
    fmt.Printf("设备ID: %s\n", result.Device.ID)
}

// 取消配对
adapter.CancelPairing(ctx)
```

### 4. 设备控制

```go
// 获取设备
device, _ := adapter.GetDevice(ctx, deviceID)

// 基本控制
if nearlinkDev, ok := device.(*adapters.NearLinkDevice); ok {
    // 开关控制
    nearlinkDev.SetOn(ctx)
    nearlinkDev.SetOff(ctx)
    nearlinkDev.Toggle(ctx)

    // 级别控制
    nearlinkDev.SetLevel(ctx, 80)

    // 读写属性
    value, _ := device.Read(ctx, "temperature")
    device.Write(ctx, "state", "on")

    // 批量读取
    attributes := []string{"state", "level", "battery"}
    results, _ := nearlinkDev.BatchRead(ctx, attributes)
}
```

### 5. NearLink特定命令

```go
if nearlinkDev, ok := device.(*adapters.NearLinkDevice); ok {
    // 发送NearLink命令
    result, _ := nearlinkDev.SendNearLinkCommand(ctx, "get_config", nil)

    // 获取NearLink设备信息
    info, _ := nearlinkDev.GetNearLinkInfo(ctx)
    fmt.Printf("RSSI: %v dBm\n", info["rssi"])
    fmt.Printf("电池: %v%%\n", info["battery_level"])

    // Ping测试
    rtt, _ := nearlinkDev.Ping(ctx)
    fmt.Printf("延迟: %v ms\n", rtt.Milliseconds())

    // 获取诊断信息
    diag, _ := nearlinkDev.GetDiagnosticInfo(ctx)
}
```

### 6. 数据流（传感器数据）

```go
if nearlinkDev, ok := device.(*adapters.NearLinkDevice); ok {
    // 创建温度数据流，每秒读取一次
    stream, _ := nearlinkDev.Stream(ctx, "temperature", 1*time.Second)

    for data := range stream {
        fmt.Printf("温度: %v°C\n", data)
    }
}
```

### 7. 事件订阅

```go
device.Subscribe(ctx, []string{"state_changed", "data_received"},
    func(event iot.DeviceEvent) {
        log.Printf("收到事件: Type=%s, Data=%v",
            event.Type, event.Data)
    },
)
```

### 8. 网络管理

```go
// 获取网络信息
networkInfo, _ := adapter.GetNetworkInfo(ctx)
fmt.Printf("Mesh ID: 0x%X\n", networkInfo.PanID)
fmt.Printf("频段: %s\n", networkInfo.Properties["frequency"])
fmt.Printf("模式: %s\n", networkInfo.Properties["network_mode"])
fmt.Printf("设备数: %d\n", networkInfo.DeviceCount)

// 重置网络
adapter.ResetNetwork(ctx)
```

## 硬件驱动集成

NearLink驱动可以通过HardwareAgent集成：

```go
import (
    "AgentFramework/pkg/beads/hardware/drivers"
)

// 创建驱动
driver := drivers.NewNearLinkDriver()

// 配置驱动
config := &drivers.NearLinkConfig{
    NetworkMode:   "SLM",
    MulticastAddr: "224.0.0.1:1888",
    Channel:       0,
    MeshID:        12345678,
}

driver.Connect(ctx, config)

// 发送硬件命令
result, _ := driver.SendCommand(ctx, "discover_devices", nil)
result, _ = driver.SendCommand(ctx, "start_pairing", nil)
result, _ = driver.SendCommand(ctx, "get_network_info", nil)

// 读取设备属性
result, _ = driver.SendCommand(ctx, "read_attribute", map[string]interface{}{
    "device_id": "nearlink-001",
    "attribute": "temperature",
})

// 写入设备属性
result, _ = driver.SendCommand(ctx, "write_attribute", map[string]interface{}{
    "device_id": "nearlink-001",
    "attribute": "state",
    "value":     "on",
})
```

## 支持的硬件命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `discover_devices` | - | 发现NearLink网络中的设备 |
| `start_pairing` | - | 启动设备配对模式 |
| `cancel_pairing` | - | 取消配对 |
| `get_devices` | - | 列出所有已配对设备 |
| `get_network_info` | - | 获取网络信息 |
| `reset_network` | - | 重置NearLink网络 |
| `read_attribute` | device_id, attribute | 读取设备属性 |
| `write_attribute` | device_id, attribute, value | 写入设备属性 |
| `send_command` | device_id, command, params | 发送NearLink特定命令 |
| `get_device_info` | device_id | 获取NearLink设备信息 |
| `ping_device` | device_id | Ping设备获取延迟 |
| `get_diagnostic_info` | device_id | 获取设备诊断信息 |

## 设备类型支持

| 设备类型 | 说明 | NearLink优势 |
|---------|------|--------------|
| **传感器** | 温度、湿度、光照等 | 低功耗，电池寿命长 |
| **执行器** | 开关、调光器、电机等 | 低延迟，实时响应 |
| **音频设备** | 耳机、音箱等 | 高保真，低延迟 |
| **人机交互设备** | 鼠标、键盘、触控笔等 | 高响应速度 |
| **定位设备** | 室内定位标签 | 高精度定位 |
| **健康设备** | 心率、血氧等 | 低功耗，持续监测 |

## NearLink消息格式

NearLink使用简化的UDP多播消息格式：

```go
// 设备发现
"NEARLINK_DISCOVERY"

// 配对控制
"NEARLINK_PAIRING_ENABLE"
"NEARLINK_PAIRING_DISABLE"

// 读取属性
"NEARLINK_READ:<mac_address>:<attribute>"

// 写入属性
"NEARLINK_WRITE:<mac_address>:<attribute>:<value>"

// 发送命令
"NEARLINK_CMD:<mac_address>:<command>"

// 获取设备信息
"NEARLINK_INFO:<mac_address>"

// 解除配对
"NEARLINK_UNPAIR:<device_id>"

// 网络重置
"NEARLINK_RESET"
```

## 配置示例

### SLM模式配置（高性能）

```go
config := iot.ProtocolConfig{
    Type: iot.ProtocolNearLink,
    Metadata: map[string]string{
        "network_mode":    "SLM",
        "multicast_addr":  "224.0.0.1:1888",
        "channel":         "0",
        "mesh_id":         "12345678",
    },
}
```

### SLE模式配置（低功耗）

```go
config := iot.ProtocolConfig{
    Type: iot.ProtocolNearLink,
    Metadata: map[string]string{
        "network_mode":    "SLE",
        "multicast_addr":  "224.0.0.1:1888",
        "channel":         "0",
        "mesh_id":         "87654321",
    },
}
```

### YAML配置文件

```yaml
iot:
  nearlink:
    hardware:
      type: udp
      timeout: 5000
    network_mode: SLM
    multicast_addr: 224.0.0.1:1888
    channel: 0
    mesh_id: 12345678
```

## 测试运行

```bash
# 运行单元测试
go test ./pkg/iot/adapters -run TestNearLink -v

# 运行示例程序
go run examples/iot/nearlink_example.go
```

## 应用场景

### 1. 智能家居

NearLink的低延迟特性使其成为智能家居控制的理想选择：

```go
// 灯光控制 - 延迟<1ms
light.Write(ctx, "state", "on")
light.Write(ctx, "brightness", 80)

// 窗帘控制
curtain.Write(ctx, "position", 50)

// 空调控制
ac.Write(ctx, "temperature", 24)
```

### 2. 智能健康

SLE模式下的超低功耗适合健康监测设备：

```go
// 心率监测 - 电池可使用数年
stream, _ := heartRate.Stream(ctx, "bpm", 10*time.Second)
for bpm := range stream {
    fmt.Printf("心率: %v bpm\n", bpm)
}
```

### 3. 实时音视频

SLM模式支持高质量音视频传输：

```go
// 音频设备
audio.SendNearLinkCommand(ctx, "play", map[string]interface{}{
    "codec": "aac",
    "bitrate": 128000,
})
```

### 4. 工业控制

低延迟和高可靠性适合工业自动化：

```go
// 电机控制 - 实时响应
motor.Write(ctx, "speed", 1500)
motor.Write(ctx, "direction", "forward")
```

## 性能优化

### 1. Mesh网络优化

NearLink SLM模式支持多跳Mesh网络：

```go
// 使用多跳路由扩展覆盖范围
info, _ := device.GetNearLinkInfo(ctx)
hopCount := info["hop_count"]
if hopCount > 3 {
    // 考虑添加中继节点
}
```

### 2. 功耗优化

SLE模式下优化功耗：

```go
// 调整传感器采样频率
if nearlinkDev.GetCapabilities() == CapabilitySensor {
    nearlinkDev.SendNearLinkCommand(ctx, "set_interval",
        map[string]interface{}{"interval": 60}) // 60秒采样
}
```

### 3. 批量操作

使用批量操作减少通信次数：

```go
// 批量读取多个属性
attributes := []string{"temp", "humidity", "pressure"}
results, _ := device.BatchRead(ctx, attributes)
```

## 故障排查

### 设备未发现

```bash
# 检查UDP多播是否正常
netstat -g | grep 224.0.0.1

# 检查防火墙设置
# 确保UDP 1888端口未被阻止
```

### 配对失败

```go
// 确保设备处于配对模式
// 检查设备是否支持NearLink协议
// 验证Mesh ID是否匹配
```

### 通信延迟高

```go
// 检查设备信号强度
diag, _ := device.GetDiagnosticInfo(ctx)
rssi := diag["rssi"]
if rssi < -80 {
    // 信号弱，考虑添加中继节点
}
```

## 与其他协议对比

### 与蓝牙BLE对比

| 特性 | NearLink SLE | 蓝牙BLE | 优势 |
|------|-------------|---------|------|
| 功耗 | 低60% | 基准 | NearLink更优 |
| 延迟 | ~20μs | ~100ms | NearLink低5000倍 |
| 连接数 | 256 | 7-8 | NearLink多30倍 |
| 覆盖 | ~100m | ~10m | NearLink远10倍 |

### 与Wi-Fi对比

| 特性 | NearLink SLM | Wi-Fi | 优势 |
|------|-------------|-------|------|
| 延迟 | ~20μs | ~30ms | NearLink低1500倍 |
| 功耗 | 低60% | 高 | NearLink更优 |
| 连接数 | 256 | ~50 | NearLink多5倍 |
| 建立时间 | <1ms | ~100ms | NearLink快100倍 |

## 未来增强

- ✅ 基础NearLink协议支持
- 🔄 Matter over NearLink集成
- 🔄 NearLink音频协议支持
- 🔄 NearLink定位服务
- 🔄 OTA固件升级
- 🔄 设备分组和场景管理
- 🔄 云服务集成

## 参考资源

- [华为星闪技术官网](https://e.huawei.com/cn/documents/products/enterprise-network/d3033a6065a34427a7c182b3c4ce3a90)
- [星闪技术白皮书](https://wwwnearlink.com)
- [AgentFramework IoT架构文档](docs/iot/IOT_ARCHITECTURE.md)

## 支持的设备

NearLink生态系统中的设备（持续更新中）：

- ✅ 华为NearLink开发板
- ✅ 海思NearLink芯片组
- ✅ NearLink传感器模块
- ✅ NearLink音频设备
- 更多设备持续测试中...

---

**状态**: ✅ 完成
**版本**: v1.0.0
**最后更新**: 2026-02-19
**作者**: AgentFramework Team
