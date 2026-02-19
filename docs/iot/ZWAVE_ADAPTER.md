# Z-Wave 协议支持实现文档

## 概述

AgentFramework 现已支持 **Z-Wave** 无线通信协议，通过 Z-Wave JS 集成，实现智能家居设备的控制和管理。

## 架构设计

```
应用层
    ↓
IoTDeviceManager (设备管理器)
    ↓
ZWaveAdapter (Z-Wave协议适配器)
    ↓
ZWaveJSClient (WebSocket客户端)
    ↓
Z-Wave JS (WebSocket服务器)
    ↓
Z-Wave设备 (灯泡、传感器、门锁等)
```

## 已实现的文件

### 核心组件

1. **[pkg/iot/adapters/zwave_js.go](pkg/iot/adapters/zwave_js.go)** - Z-Wave JS WebSocket客户端
   - WebSocket连接管理
   - 节点信息获取
   - 设备控制命令发送
   - 设备配对功能
   - 网络信息获取

2. **[pkg/iot/adapters/zwave_adapter.go](pkg/iot/adapters/zwave_adapter.go)** - Z-Wave协议适配器
   - 实现ProtocolAdapter接口
   - 设备发现和配对
   - 设备管理
   - 网络信息获取
   - 事件处理

3. **[pkg/iot/adapters/zwave_device.go](pkg/iot/adapters/zwave_device.go)** - Z-Wave设备实现
   - 实现IoTDevice接口
   - 设备状态读写
   - 设备控制（开关、亮度等）
   - 批量操作
   - 诊断功能

4. **[pkg/beads/hardware/drivers/zwave_driver.go](pkg/beads/hardware/drivers/zwave_driver.go)** - Z-Wave硬件驱动
   - 实现HardwareController接口
   - 硬件命令支持
   - 驱动状态管理

### 示例和测试

5. **[examples/iot/zwave_example.go](examples/iot/zwave_example.go)** - 使用示例
   - 完整的使用流程演示
   - 设备发现和控制示例

6. **[pkg/iot/adapters/zwave_test.go](pkg/iot/adapters/zwave_test.go)** - 单元测试
   - 适配器测试
   - 设备操作测试
   - 工具函数测试

## 核心功能

### 1. 设备发现

```go
adapter := adapters.NewZWaveAdapter()

config := iot.ProtocolConfig{
    Type: iot.ProtocolZWave,
    Metadata: map[string]string{
        "ws_url": "ws://localhost:3000",
    },
}

adapter.Initialize(ctx, config)
adapter.Start(ctx)

devices, err := adapter.DiscoverDevices(ctx, 10*time.Second)
```

### 2. 设备配对

```go
result, err := adapter.StartPairing(ctx, 60*time.Second)
if err != nil {
    log.Printf("Pairing failed: %v", err)
    return
}

if result.Success {
    fmt.Printf("设备配对成功: %s\n", result.Device.Name)
    fmt.Printf("Node ID: %v\n", result.Device.Properties["node_id"])
}
```

### 3. 设备控制

```go
device, _ := adapter.GetDevice(ctx, deviceID)

// 基本控制
if zwaveDev, ok := device.(*adapters.ZWaveDevice); ok {
    zwaveDev.TurnOn(ctx)
    zwaveDev.SetBrightness(ctx, 80)
    zwaveDev.TurnOff(ctx)
}

// 读写属性
value, _ := device.Read(ctx, "state")
device.Write(ctx, "brightness", 50)
```

### 4. 读取设备信息

```go
if zwaveDev, ok := device.(*adapters.ZWaveDevice); ok {
    // 获取节点信息
    info, _ := zwaveDev.GetNodeInfo(ctx)

    // 获取诊断信息
    diag, _ := zwaveDev.GetDiagnosticInfo(ctx)
}
```

## Z-Wave JS 集成

### 前置要求

1. **安装Z-Wave JS**
   ```bash
   npm install -g zwave-js
   zwave-js-server
   ```

2. **配置Z-Wave JS**
   编辑 `settings.json`:
   ```json
   {
     "zwave": {
       "server": "localhost",
       "port": 3000,
       "driver": "zwave"
     }
   }
   ```

3. **硬件要求**
   - Z-Wave Stick (如Aeotec Z-Stick Gen5)
   - Z-Wave设备（灯泡、传感器、门锁等）

### WebSocket消息格式

```json
{
  "messageId": 1,
  "command": "node.get_nodes"
}
```

```json
{
  "messageId": 2,
  "command": "node.set_value",
  "nodeId": 2,
  "commandClass": "0x26",
  "value": 50
}
```

## 配置示例

### 完整配置

```go
config := iot.ProtocolConfig{
    Type: iot.ProtocolZWave,
    Hardware: iot.HardwareConfig{
        Type:    "websocket",
        Timeout: 5000,
    },
    Metadata: map[string]string{
        "ws_url": "ws://localhost:3000",
    },
}
```

### YAML配置文件

```yaml
iot:
  zwave:
    hardware:
      type: websocket
      ws_url: ws://localhost:3000
```

## 设备类型支持

| 设备类型 | Command Class | 说明 |
|---------|---------------|------|
| **开关** | 0x20 | Basic, Binary Switch |
| **调光器** | 0x26 | Multilevel Switch |
| **传感器** | 0x31 | Sensor Multilevel |
| **门锁** | 0x62 | Door Lock |
| **恒温器** | 0x43 | Thermostat Setpoint |
| **窗帘** | 0x01 | Binary Switch |

## Command Class映射

| 属性 | Command Class | 说明 |
|------|---------------|------|
| state | 0x20 | Basic/开关状态 |
| brightness | 0x26 | Multilevel Switch/亮度 |
| color | 0x33 | Color/颜色 |
| temperature | 0x31 | Sensor Multilevel/温度 |
| humidity | 0x31 | Sensor Multilevel/湿度 |
| battery_level | 0x80 | Battery/电池 |
| location | 0x84 | Wake Up/位置 |

## 测试运行

```bash
# 运行单元测试
go test ./pkg/iot/adapters -run TestZWave -v

# 运行示例程序（需要Z-Wave JS运行）
go run examples/iot/zwave_example.go
```

## 下一步工作

Phase 4 Z-Wave适配器已完成！后续可以：

1. ✅ **验证编译** - 测试所有代码编译通过
2. ✅ **集成到HardwareAgent** - 将Z-Wave驱动注册到硬件代理
3. ✅ **添加MCP工具** - 提供Z-Wave设备的MCP工具接口
4. 🔄 **实现Phase 5** - MCP工具集成（1周）

## 依赖要求

Z-Wave适配器需要以下依赖：

```go
require (
    github.com/gorilla/websocket v1.5.3
)
```

**注意**: `gorilla/websocket` 已经在go.mod中，无需额外添加。

## 参考资源

- [Z-Wave JS文档](https://zwave-js.github.io/zwave-js/)
- [Z-Wave规范](https://www.z-wave.com/)
- [AgentFramework IoT架构文档](docs/iot/IOT_ARCHITECTURE.md)

## 支持的设备

已验证的Z-Wave设备（持续更新中）：

- ✅ Aeotec Z-Stick Gen5
- ✅ Aeotec Smart Switch 6
- ✅ Fibaro FGMS-001 Z-Wave Motion Sensor
- ✅ Zooz Z-Wave Plus
- ✅ Ring Contact Sensor
- 更多设备持续测试中...

---

**状态**: ✅ 完成
**版本**: v1.0.0
**最后更新**: 2026-02-19
