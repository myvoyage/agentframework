# AgentFramework IoT 协议支持文档

## 概述

AgentFramework 提供完整的 **IoT协议抽象层**，支持多种无线通信协议，实现统一的设备管理和自动化控制。

## 支持的协议

| 协议 | 说明 | 状态 | 文档 |
|------|------|------|------|
| **Zigbee** | 基于IEEE 802.15.4的Mesh网络 | ✅ 完成 | [ZIGBEE_ADAPTER.md](ZIGBEE_ADAPTER.md) |
| **Thread** | 基于IP的Mesh网络 (6LoWPAN) | ✅ 完成 | [THREAD_ADAPTER.md](THREAD_ADAPTER.md) |
| **Z-Wave** | 专为智能家居设计的无线协议 | ✅ 完成 | [ZWAVE_ADAPTER.md](ZWAVE_ADAPTER.md) |
| **NearLink** | 华为星闪技术 (低延迟、低功耗) | ✅ 完成 | [NEARLINK_ADAPTER.md](NEARLINK_ADAPTER.md) |

## 协议对比

| 特性 | Zigbee | Thread | Z-Wave | NearLink |
|------|--------|--------|--------|----------|
| **工作频段** | 2.4GHz | 2.4GHz | 800/900MHz | 2.4/5.1GHz |
| **数据速率** | 250Kbps | 250Kbps | 100Kbps | 12Mbps |
| **网络拓扑** | Mesh | Mesh | Mesh | Mesh |
| **最大节点数** | 65K+ | 300+ | 232 | 256 |
| **延迟** | ~100ms | ~50ms | ~200ms | ~20μs |
| **功耗** | 低 | 低 | 中 | 极低 |
| **Mesh路由** | 是 | 是 | 是 | 是 |
| **IP支持** | 否 | 是 (IPv6) | 否 | 是 |
| **主要应用** | 智能家居 | 智能家居 | 智能家居 | 智能家居/工业/音频 |

## 快速开始

### 1. 初始化适配器

```go
import (
    "AgentFramework/pkg/iot"
    "AgentFramework/pkg/iot/adapters"
)

// 创建适配器
zigbeeAdapter := adapters.NewZigbeeAdapter()

// 初始化
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

err := zigbeeAdapter.Initialize(ctx, config)
if err != nil {
    log.Fatal(err)
}

// 启动适配器
err = zigbeeAdapter.Start(ctx)
if err != nil {
    log.Fatal(err)
}
```

### 2. 发现设备

```go
// 发现设备
devices, err := zigbeeAdapter.DiscoverDevices(ctx, 10*time.Second)
if err != nil {
    log.Printf("发现设备失败: %v", err)
    return
}

for _, device := range devices {
    fmt.Printf("发现设备: %s (%s)\n", device.Name, device.ID)
}
```

### 3. 控制设备

```go
// 获取设备
device, err := zigbeeAdapter.GetDevice(ctx, "zigbee-bulb-001")
if err != nil {
    log.Fatal(err)
}

// 打开设备
err = device.Write(ctx, "state", "on")
if err != nil {
    log.Fatal(err)
}

// 设置亮度
err = device.Write(ctx, "brightness", 80)
if err != nil {
    log.Fatal(err)
}
```

### 4. 订阅事件

```go
// 订阅设备事件
err = device.Subscribe(ctx, []string{"state_changed", "attribute_changed"},
    func(event iot.DeviceEvent) {
        log.Printf("收到事件: Type=%s, Data=%v", event.Type, event.Data)
    },
)
if err != nil {
    log.Fatal(err)
}
```

## 核心概念

### 设备抽象

所有IoT设备实现统一的 `IoTDevice` 接口：

```go
type IoTDevice interface {
    // 身份信息
    ID() string
    Name() string
    Type() DeviceType
    Protocol() ProtocolType

    // 连接管理
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    IsConnected() bool

    // 数据交互
    Read(ctx context.Context, attribute string) (interface{}, error)
    Write(ctx context.Context, attribute string, value interface{}) error

    // 事件订阅
    Subscribe(ctx context.Context, events []string, handler iot.DeviceEventHandler) error
    Unsubscribe(ctx context.Context, events []string) error
}
```

### 协议适配器

每个协议实现 `ProtocolAdapter` 接口：

```go
type ProtocolAdapter interface {
    // 适配器信息
    Type() ProtocolType
    Version() string

    // 生命周期
    Initialize(ctx context.Context, config iot.ProtocolConfig) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error

    // 设备管理
    DiscoverDevices(ctx context.Context, timeout time.Duration) ([]*DeviceInfo, error)
    StartPairing(ctx context.Context, timeout time.Duration) (*PairingResult, error)
    CancelPairing(ctx context.Context) error

    // 设备访问
    GetDevice(ctx context.Context, deviceID string) (IoTDevice, error)
    ListDevices(ctx context.Context) ([]IoTDevice, error)
    RemoveDevice(ctx context.Context, deviceID string) error

    // 网络管理
    GetNetworkInfo(ctx context.Context) (*NetworkInfo, error)
    ResetNetwork(ctx context.Context) error
}
```

### 可选设备接口

设备可以实现的可选接口：

```go
// 批量操作
type BatchReader interface {
    BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error)
}

type BatchWriter interface {
    BatchWrite(ctx context.Context, values map[string]interface{}) error
}

// 快捷操作
type Toggleable interface {
    Toggle(ctx context.Context) error
}

// 诊断
type Pingable interface {
    Ping(ctx context.Context) (time.Duration, error)
}

type Diagnosticable interface {
    GetDiagnosticInfo(ctx context.Context) (map[string]interface{}, error)
}
```

## 高级功能

### 1. MCP工具集成

通过MCP (Model Context Protocol) 工具控制IoT设备。

**文档**: [IOT_MCP_TOOLS.md](IOT_MCP_TOOLS.md)

**工具列表**:
- `discover_iot_devices` - 发现设备
- `list_iot_devices` - 列出设备
- `read_iot_attribute` - 读取属性
- `write_iot_attribute` - 写入属性
- `set_iot_on_off` - 开关控制
- `set_iot_level` - 级别控制
- `set_iot_color` - 颜色控制
- 等等...

### 2. 工作流自动化

通过工作流引擎实现设备自动化。

**文档**: [IOT_WORKFLOW.md](IOT_WORKFLOW.md)

**核心组件**:
- **WorkflowEngine** - 工作流引擎
- **EventBus** - 事件总线
- **TaskScheduler** - 任务调度器
- **AutomationRule** - 自动化规则
- **Scenario** - 场景管理

**示例**:
```go
// 创建自动化规则
rule := &iot.AutomationRule{
    ID:      "motion-light-rule",
    Name:    "人体感应自动开灯",
    Enabled: true,
    Triggers: []iot.Trigger{
        {Type: iot.TriggerTypeEvent, Event: "motion_detected"},
    },
    Actions: []iot.Action{
        {Type: iot.ActionTypeDeviceControl, DeviceID: "zigbee-bulb-001", Attribute: "state", Value: "on"},
    },
}

engine.RegisterRule(rule)
```

### 3. Agent集成

与HardwareAgent和RealTimeAgent集成。

**示例**:
```go
// 创建HardwareAgent
hardwareAgent := agent.NewHardwareAgent()

// 注册IoT驱动
hardwareAgent.RegisterDriver("zigbee", &drivers.ZigbeeDriver{})

// 通过MCP工具控制设备
// 或通过工作流引擎自动化
```

## 文档索引

### 协议文档

1. **[Zigbee适配器](ZIGBEE_ADAPTER.md)**
   - 架构设计
   - API使用
   - 设备类型支持
   - 故障排查

2. **[Thread适配器](THREAD_ADAPTER.md)**
   - Thread网络特性
   - Border Router配置
   - Mesh拓扑管理
   - IP通信支持

3. **[Z-Wave适配器](ZWAVE_ADAPTER.md)**
   - Z-Wave JS集成
   - Command Class映射
   - Inclusion/Exclusion
   - 网络愈合

4. **[NearLink适配器](NEARLINK_ADAPTER.md)**
   - 星闪技术特性
   - SLM/SLE双模式
   - 性能优势
   - 应用场景

### 功能文档

5. **[MCP工具集成](IOT_MCP_TOOLS.md)**
   - MCP工具列表
   - Schema定义
   - 使用示例
   - 接口扩展

6. **[工作流自动化](IOT_WORKFLOW.md)**
   - 工作流引擎
   - 自动化规则
   - 场景管理
   - API参考

### 进度报告

7. **[进度报告](PROGRESS_REPORT.md)**
   - 项目进展
   - 里程碑
   - 统计数据

## 示例代码

### Zigbee示例

```bash
go run examples/iot/zigbee_example.go
```

### Thread示例

```bash
go run examples/iot/thread_example.go
```

### Z-Wave示例

```bash
go run examples/iot/zwave_example.go
```

### NearLink示例

```bash
go run examples/iot/nearlink_example.go
```

### 工作流示例

```bash
go run examples/iot/workflow_example.go
```

## 测试

### 单元测试

```bash
# Zigbee适配器测试
go test ./pkg/iot/adapters -run TestZigbee -v

# Thread适配器测试
go test ./pkg/iot/adapters -run TestThread -v

# Z-Wave适配器测试
go test ./pkg/iot/adapters -run TestZWave -v

# NearLink适配器测试
go test ./pkg/iot/adapters -run TestNearLink -v
```

### 集成测试

```bash
# 工作流引擎测试
go test ./pkg/iot -run TestWorkflowEngine -v

# MCP工具测试
go test ./pkg/beads/mcp -run TestIoTMCPTools -v
```

## 性能考虑

### Zigbee

- ✅ 低功耗，适合电池供电设备
- ✅ Mesh网络，自愈能力强
- ⚠️ 数据速率较低 (250Kbps)
- ⚠️ 需要协调器 (Coordinator)

### Thread

- ✅ 基于IP，易于集成
- ✅ 低延迟，适合实时控制
- ✅ 与互联网无缝集成
- ⚠️ 需要Border Router
- ⚠️ 设备数量相对有限

### Z-Wave

- ✅ 专为智能家居优化
- ✅ 长距离传输 (sub-GHz)
- ✅ 设备互操作性好
- ⚠️ 数据速率低
- ⚠️ 需要Z-Wave JS服务器

### NearLink

- ✅ 极低延迟 (~20μs)
- ✅ 低功耗，电池寿命长
- ✅ 高数据速率 (12Mbps)
- ✅ 支持大量设备 (256个)
- ⚠️ 新技术，生态尚在发展

## 应用场景

### 智能家居

- **灯光控制**: 所有协议都支持
- **传感器监控**: Zigbee, Z-Wave, NearLink
- **门锁控制**: Z-Wave, Thread, NearLink
- **温控系统**: 所有协议

### 工业控制

- **设备监控**: NearLink, Thread
- **实时控制**: NearLink, Thread
- **数据采集**: Zigbee, NearLink
- **预测性维护**: Thread, NearLink

### 智能健康

- **可穿戴设备**: NearLink (SLE模式)
- **健康监测**: Zigbee, Thread
- **医疗设备**: NearLink

### 音视频

- **无线音频**: NearLink (SLM模式)
- **视频传输**: NearLink (SLM模式)
- **低延迟通信**: NearLink, Thread

## 最佳实践

### 1. 协议选择

根据应用场景选择合适的协议：

| 场景 | 推荐协议 |
|------|---------|
| 智能家居（灯光、传感器） | Zigbee, Z-Wave |
| IP设备和互联网集成 | Thread |
| 实时音频/视频 | NearLink (SLM) |
| 电池供电传感器 | NearLink (SLE), Zigbee |
| 工业控制 | NearLink, Thread |

### 2. 设备管理

- ✅ 使用适配器管理器统一管理多协议设备
- ✅ 定期执行设备发现以更新设备列表
- ✅ 监控设备状态和连接质量
- ✅ 实现设备降级策略

### 3. 错误处理

- ✅ 实现重试机制处理网络错误
- ✅ 记录详细的错误日志
- ✅ 提供设备离线通知
- ✅ 实现优雅降级

### 4. 性能优化

- ✅ 使用批量操作减少通信次数
- ✅ 优化事件订阅，避免过度监听
- ✅ 实现本地缓存减少网络请求
- ✅ 使用异步操作避免阻塞

## 故障排查

### 常见问题

1. **设备未发现**
   - 检查适配器是否启动
   - 验证网络配置
   - 确认设备处于配对模式

2. **连接失败**
   - 检查硬件连接
   - 验证设备ID
   - 查看错误日志

3. **控制无响应**
   - 检查设备是否在线
   - 验证属性和值是否正确
   - Ping设备测试连通性

4. **性能问题**
   - 减少轮询频率
   - 使用事件驱动模式
   - 优化批量操作

## 贡献指南

欢迎贡献代码和建议！

1. Fork项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

本项目采用Apache 2.0许可证。详见 [LICENSE](../../LICENSE) 文件。

## 联系方式

- 项目主页: [AgentFramework](https://github.com/myvoyage/agentframework)
- 问题反馈: [GitHub Issues](https://github.com/myvoyage/agentframework/issues)
- 文档: [完整文档](../../docs/)

---

**最后更新**: 2026-02-19
**版本**: v1.0.0
**作者**: AgentFramework Team
