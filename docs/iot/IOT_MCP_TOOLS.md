# IoT MCP工具集成文档

## 概述

AgentFramework 已成功集成 **IoT协议MCP工具**，提供完整的Model Context Protocol工具集，支持通过MCP协议控制和管理IoT设备。

## 架构设计

```
MCP Server
    ↓
IoTMCPTools (IoT工具集)
    ↓
HardwareAgent → IoTDeviceManager
    ↓
Protocol Adapters (Zigbee, Z-Wave, Thread, NearLink)
    ↓
IoT Devices
```

## 已实现的MCP工具

### 设备发现和配对 (3个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `discover_iot_devices` | 发现IoT设备 | protocol, timeout_seconds |
| `start_iot_pairing` | 启动设备配对 | protocol, timeout_seconds |
| `cancel_iot_pairing` | 取消配对 | protocol |

**示例：**
```json
{
  "name": "discover_iot_devices",
  "arguments": {
    "protocol": "zigbee",
    "timeout_seconds": 10
  }
}
```

### 设备管理 (3个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `list_iot_devices` | 列出所有设备 | protocol (可选) |
| `get_iot_device_info` | 获取设备详细信息 | device_id |
| `remove_iot_device` | 移除/解绑设备 | device_id |

**示例：**
```json
{
  "name": "list_iot_devices",
  "arguments": {
    "protocol": "zwave"
  }
}
```

### 设备控制 (6个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `read_iot_attribute` | 读取设备属性 | device_id, attribute |
| `write_iot_attribute` | 写入设备属性 | device_id, attribute, value |
| `batch_read_iot_attributes` | 批量读取属性 | device_id, attributes[] |
| `batch_write_iot_attributes` | 批量写入属性 | device_id, values{} |
| `set_iot_on_off` | 开关控制 | device_id, state (on/off/toggle) |
| `set_iot_level` | 级别控制 | device_id, level (0-100) |

**示例：**
```json
{
  "name": "set_iot_on_off",
  "arguments": {
    "device_id": "zigbee-bulb-001",
    "state": "on"
  }
}
```

```json
{
  "name": "batch_read_iot_attributes",
  "arguments": {
    "device_id": "zwave-sensor-002",
    "attributes": ["temperature", "humidity", "battery"]
  }
}
```

### 设备快捷操作 (2个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `set_iot_color` | 颜色控制 (RGB灯) | device_id, color (hex) |

**示例：**
```json
{
  "name": "set_iot_color",
  "arguments": {
    "device_id": "zigbee-rgb-003",
    "color": "#FF0000"
  }
}
```

### 网络管理 (2个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `get_iot_network_info` | 获取网络信息 | protocol |
| `reset_iot_network` | 重置网络 | protocol |

**示例：**
```json
{
  "name": "get_iot_network_info",
  "arguments": {
    "protocol": "thread"
  }
}
```

### 设备诊断 (2个工具)

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `ping_iot_device` | Ping设备 | device_id |
| `get_iot_device_diagnostics` | 获取诊断信息 | device_id |

**示例：**
```json
{
  "name": "ping_iot_device",
  "arguments": {
    "device_id": "nearlink-sensor-004"
  }
}
```

### 协议特定工具 (3个工具)

| 工具名 | 功能 | 协议 |
|--------|------|------|
| `zwave_heal_network` | Z-Wave网络愈合 | Z-Wave |
| `thread_get_mesh_topology` | Thread Mesh拓扑 | Thread |
| `nearlink_get_mode` | NearLink模式查询 | NearLink |

**示例：**
```json
{
  "name": "zwave_heal_network",
  "arguments": {}
}
```

## 工具使用场景

### 场景1：智能家居控制

```json
// 1. 发现设备
{"name": "discover_iot_devices", "arguments": {"protocol": "zigbee"}}

// 2. 打开灯光
{"name": "set_iot_on_off", "arguments": {"device_id": "zigbee-bulb-001", "state": "on"}}

// 3. 调整亮度
{"name": "set_iot_level", "arguments": {"device_id": "zigbee-bulb-001", "level": 80}}

// 4. 设置颜色
{"name": "set_iot_color", "arguments": {"device_id": "zigbee-rgb-002", "color": "#00FF00"}}
```

### 场景2：传感器监控

```json
// 1. 批量读取传感器数据
{
  "name": "batch_read_iot_attributes",
  "arguments": {
    "device_id": "zwave-sensor-003",
    "attributes": ["temperature", "humidity", "motion"]
  }
}

// 2. Ping设备检查连接
{
  "name": "ping_iot_device",
  "arguments": {"device_id": "zwave-sensor-003"}
}

// 3. 获取诊断信息
{
  "name": "get_iot_device_diagnostics",
  "arguments": {"device_id": "zwave-sensor-003"}
}
```

### 场景3：设备配对

```json
// 1. 启动配对模式
{
  "name": "start_iot_pairing",
  "arguments": {
    "protocol": "nearlink",
    "timeout_seconds": 60
  }
}

// 2. 等待配对结果... (自动)

// 3. 列出所有设备
{
  "name": "list_iot_devices",
  "arguments": {"protocol": "nearlink"}
}
```

### 场景4：网络管理

```json
// 1. 获取网络信息
{
  "name": "get_iot_network_info",
  "arguments": {"protocol": "thread"}
}

// 2. Z-Wave网络愈合
{
  "name": "zwave_heal_network",
  "arguments": {}
}

// 3. 获取Thread Mesh拓扑
{
  "name": "thread_get_mesh_topology",
  "arguments": {}
}
```

## MCP工具Schema定义

### discover_iot_devices

```json
{
  "name": "discover_iot_devices",
  "description": "Discover IoT devices on the specified protocol network",
  "inputSchema": {
    "type": "object",
    "properties": {
      "protocol": {
        "type": "string",
        "description": "IoT protocol to use for discovery",
        "enum": ["zigbee", "zwave", "thread", "nearlink"]
      },
      "timeout_seconds": {
        "type": "number",
        "description": "Discovery timeout in seconds (default: 10)"
      }
    },
    "required": ["protocol"]
  }
}
```

### set_iot_on_off

```json
{
  "name": "set_iot_on_off",
  "description": "Turn an IoT device on or off",
  "inputSchema": {
    "type": "object",
    "properties": {
      "device_id": {
        "type": "string",
        "description": "IoT device identifier"
      },
      "state": {
        "type": "string",
        "description": "Device state",
        "enum": ["on", "off", "toggle"]
      }
    },
    "required": ["device_id", "state"]
  }
}
```

### batch_read_iot_attributes

```json
{
  "name": "batch_read_iot_attributes",
  "description": "Read multiple attributes from an IoT device at once",
  "inputSchema": {
    "type": "object",
    "properties": {
      "device_id": {
        "type": "string",
        "description": "IoT device identifier"
      },
      "attributes": {
        "type": "array",
        "description": "List of attribute names to read",
        "items": {
          "type": "string"
        }
      }
    },
    "required": ["device_id", "attributes"]
  }
}
```

## 接口扩展

为了支持MCP工具，IoT设备接口扩展了以下可选接口：

### BatchReader

```go
type BatchReader interface {
    BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error)
}
```

**实现：**
- [ZWaveDevice.BatchRead()](pkg/iot/adapters/zwave_device.go:244)
- [ThreadDevice.BatchRead()](pkg/iot/adapters/thread_device.go:322)
- [NearLinkDevice.BatchRead()](pkg/iot/adapters/nearlink_device.go:223)

### BatchWriter

```go
type BatchWriter interface {
    BatchWrite(ctx context.Context, values map[string]interface{}) error
}
```

**实现：**
- [ZWaveDevice.BatchWrite()](pkg/iot/adapters/zwave_device.go:259)
- [ThreadDevice.BatchWrite()](pkg/iot/adapters/thread_device.go:337)
- [NearLinkDevice.BatchWrite()](pkg/iot/adapters/nearlink_device.go:238)

### Toggleable

```go
type Toggleable interface {
    Toggle(ctx context.Context) error
}
```

**实现：**
- [ThreadDevice.Toggle()](pkg/iot/adapters/thread_device.go:257)
- [NearLinkDevice.Toggle()](pkg/iot/adapters/nearlink_device.go:187)

### Pingable

```go
type Pingable interface {
    Ping(ctx context.Context) (time.Duration, error)
}
```

**实现：**
- [ZWaveDevice.Ping()](pkg/iot/adapters/zwave_device.go:185)
- [ThreadDevice.Ping()](pkg/iot/adapters/thread_device.go:464)
- [NearLinkDevice.Ping()](pkg/iot/adapters/nearlink_device.go:160)

### Diagnosticable

```go
type Diagnosticable interface {
    GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error)
}
```

**实现：**
- [ZWaveDevice.GetDiagnosticInfo()](pkg/iot/adapters/zwave_device.go:198)
- [ThreadDevice.GetDiagnosticInfo()](pkg/iot/adapters/thread_device.go:348)
- [NearLinkDevice.GetDiagnosticInfo()](pkg/iot/adapters/nearlink_device.go:167)

## 文件结构

```
pkg/beads/mcp/
├── hardware_mcp.go        # 原有硬件MCP工具
├── iot_mcp.go             # 新增IoT MCP工具 ✅
└── ...

pkg/iot/
├── device.go              # 扩展设备接口 ✅
├── types.go               # 添加ProtocolNearLink ✅
└── ...
```

## 集成步骤

### 1. 注册MCP工具

```go
import (
    "AgentFramework/pkg/beads/mcp"
    "github.com/mark3labs/mcp-go/server"
)

// 创建MCP服务器
mcpServer := server.NewMCPServer()

// 创建硬件代理
hardwareAgent := agent.NewHardwareAgent()

// 注册IoT MCP工具
iotTools := mcp.NewIoTMCPTools(hardwareAgent)
iotTools.RegisterTools(mcpServer)
```

### 2. 使用MCP工具

通过MCP客户端调用工具：

```python
# Python示例
result = mcp_client.call_tool("discover_iot_devices", {
    "protocol": "zigbee",
    "timeout_seconds": 10
})
```

```javascript
// JavaScript示例
const result = await mcpClient.callTool("set_iot_on_off", {
    device_id: "zigbee-bulb-001",
    state: "on"
});
```

## 协议支持矩阵

| 功能 | Zigbee | Z-Wave | Thread | NearLink |
|------|--------|--------|--------|----------|
| 设备发现 | ✅ | ✅ | ✅ | ✅ |
| 设备配对 | ✅ | ✅ | ✅ | ✅ |
| 属性读写 | ✅ | ✅ | ✅ | ✅ |
| 批量操作 | ✅ | ✅ | ✅ | ✅ |
| 开关控制 | ✅ | ✅ | ✅ | ✅ |
| 级别控制 | ✅ | ✅ | ✅ | ✅ |
| 颜色控制 | ✅ | ✅ | ⚠️ | ⚠️ |
| Ping测试 | ✅ | ✅ | ✅ | ✅ |
| 诊断信息 | ✅ | ✅ | ✅ | ✅ |
| 网络信息 | ✅ | ✅ | ✅ | ✅ |
| 网络愈合 | - | ✅ | - | - |
| Mesh拓扑 | - | - | ✅ | - |

*✅ 完全支持，⚠️ 部分支持，- 不适用*

## 错误处理

MCP工具返回的错误格式：

```json
{
  "error": {
    "code": "DEVICE_NOT_FOUND",
    "message": "Device zigbee-unknown not found"
  }
}
```

常见错误码：
- `DEVICE_NOT_FOUND` - 设备不存在
- `ADAPTER_NOT_RUNNING` - 适配器未运行
- `INVALID_PARAMETER` - 参数无效
- `OPERATION_TIMEOUT` - 操作超时
- `NETWORK_ERROR` - 网络错误

## 性能考虑

### 批量操作优化

使用批量操作减少网络往返：

```json
// ❌ 不推荐：3次网络往返
{"name": "read_iot_attribute", "arguments": {"device_id": "sensor-001", "attribute": "temp"}}
{"name": "read_iot_attribute", "arguments": {"device_id": "sensor-001", "attribute": "humidity"}}
{"name": "read_iot_attribute", "arguments": {"device_id": "sensor-001", "attribute": "pressure"}}

// ✅ 推荐：1次网络往返
{
  "name": "batch_read_iot_attributes",
  "arguments": {
    "device_id": "sensor-001",
    "attributes": ["temp", "humidity", "pressure"]
  }
}
```

### 超时设置

合理设置超时时间：

- 设备发现：10-30秒
- 设备配对：60秒
- 属性读取：5秒
- Ping测试：3秒

## 安全考虑

1. **设备ID验证**：所有工具都验证device_id格式
2. **协议白名单**：仅支持已注册的协议类型
3. **参数验证**：所有输入参数经过类型和范围验证
4. **权限控制**：建议实现基于设备的访问控制

## 测试

```bash
# 运行MCP工具测试
go test ./pkg/beads/mcp -run TestIoTMCPTools -v

# 启动MCP服务器
go run cmd/mcp-server/main.go
```

## 未来增强

- ✅ 基础MCP工具集
- 🔄 设备分组和场景管理
- 🔄 定时任务和自动化
- 🔄 数据历史和趋势分析
- 🔄 云服务同步
- 🔄 多租户支持

---

**状态**: ✅ 完成
**版本**: v1.0.0
**最后更新**: 2026-02-19
**作者**: AgentFramework Team
