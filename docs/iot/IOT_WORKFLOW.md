# IoT工作流自动化文档

## 概述

AgentFramework 提供了完整的 **IoT工作流自动化引擎**，支持设备自动化规则、场景管理和复杂工作流编排。

## 架构设计

```
应用层 (用户交互)
    ↓
场景管理器 (Scenario Manager)
    ↓
工作流引擎 (Workflow Engine)
    ├── 事件总线 (Event Bus)
    ├── 任务调度器 (Task Scheduler)
    └── 规则引擎 (Rule Engine)
    ↓
适配器管理器 (Adapter Manager)
    ↓
协议适配器 (Zigbee, Z-Wave, Thread, NearLink)
    ↓
IoT设备
```

## 核心组件

### 1. 工作流引擎 (WorkflowEngine)

工作流引擎是自动化系统的核心，负责：
- 管理工作流、规则和场景
- 处理事件和触发器
- 执行自动化动作
- 调度定时任务

**初始化：**
```go
adapterMgr := iot.NewAdapterManager()
engine := iot.NewWorkflowEngine(adapterMgr)
engine.Start(ctx)
```

### 2. 事件总线 (EventBus)

事件总线实现发布-订阅模式，用于：
- 设备事件通知
- 跨组件通信
- 触发自动化规则

**发布事件：**
```go
eventBus.Publish(iot.Event{
    Type:    "motion_detected",
    Source:  "zigbee-sensor-001",
    Payload: map[string]interface{}{
        "motion": true,
    },
})
```

**订阅事件：**
```go
eventBus.Subscribe("motion_detected", func(event iot.Event) {
    fmt.Printf("Motion detected: %+v\n", event)
})
```

### 3. 任务调度器 (TaskScheduler)

任务调度器管理定时任务：
- Cron表达式支持
- 间隔调度
- 一次性任务

### 4. 适配器管理器 (AdapterManager)

适配器管理器统一管理所有协议适配器：
- 注册和获取适配器
- 跨协议设备访问
- 统一设备列表

## 自动化规则

### 规则结构

```go
rule := &iot.AutomationRule{
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
    },
    Actions: []iot.Action{
        {
            Type:      iot.ActionTypeDeviceControl,
            DeviceID:  "zigbee-bulb-001",
            Attribute: "state",
            Value:     "on",
        },
    },
}
```

### 触发器类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `TriggerTypeEvent` | 事件触发 | 传感器检测到运动 |
| `TriggerTypeSchedule` | 定时触发 | 每天早上8点 |
| `TriggerTypeManual` | 手动触发 | 用户手动执行 |

### 条件类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `ConditionTypeDeviceState` | 设备状态 | 灯泡关闭状态 |
| `ConditionTypeTime` | 时间条件 | 23:00 |
| `ConditionTypeEvent` | 事件匹配 | 特定事件发生 |
| `ConditionTypeExpression` | 表达式 | 复杂逻辑判断 |

### 动作类型

| 类型 | 说明 | 参数 |
|------|------|------|
| `ActionTypeDeviceControl` | 设备控制 | device_id, attribute, value |
| `ActionTypeDelay` | 延时 | duration |
| `ActionTypeNotification` | 发送通知 | title, message |
| `ActionTypeScenario` | 执行场景 | scenario_id |
| `ActionTypeHTTPRequest` | HTTP请求 | url, method |
| `ActionTypeSetVariable` | 设置变量 | name, value |

## 场景管理

### 场景结构

场景是一组预设的设备动作，可以一键执行：

```go
scenario := &iot.Scenario{
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
    },
}
```

### 注册和执行场景

```go
// 注册场景
engine.RegisterScenario(scenario)

// 执行场景
err := engine.ExecuteScenario(ctx, "evening-mode")
```

## 工作流

### 工作流结构

工作流支持更复杂的自动化逻辑：

```go
workflow := &iot.Workflow{
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
            Value:     "30",
        },
    },
    Actions: []iot.Action{
        {
            Type:      iot.ActionTypeNotification,
            Title:     "温度警报",
            Message:   "温度超过30度！",
        },
        {
            Type:      iot.ActionTypeDeviceControl,
            DeviceID:  "zigbee-ac-001",
            Attribute: "state",
            Value:     "on",
        },
    },
}
```

## 使用示例

### 示例1：人体感应自动开灯

**场景：** 检测到人体移动时，自动打开灯光并设置亮度为80%

```go
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
            Type:      iot.ConditionTypeDeviceState,
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

engine.RegisterRule(motionLightRule)
```

### 示例2：定时关灯

**场景：** 每天23:00自动关闭所有灯光

```go
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

engine.RegisterRule(autoOffRule)
```

### 示例3：温度监控

**场景：** 温度超过30度时，发送通知并自动打开空调

```go
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
            Value:     "30",
        },
    },
    Actions: []iot.Action{
        {
            Type:      iot.ActionTypeNotification,
            Title:     "温度警报",
            Message:   "温度超过30度！",
        },
        {
            Type:      iot.ActionTypeDeviceControl,
            DeviceID:  "zigbee-ac-001",
            Attribute: "state",
            Value:     "on",
        },
    },
}

engine.RegisterWorkflow(tempMonitorWorkflow)
```

### 示例4：场景联动

**场景：** 执行"晚间模式"场景，触发多个设备控制

```go
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
    },
}

engine.RegisterScenario(eveningMode)
engine.ExecuteScenario(ctx, "evening-mode")
```

## 配置文件格式

### JSON格式

```json
{
  "rules": [
    {
      "id": "motion-light-rule",
      "name": "人体感应自动开灯",
      "enabled": true,
      "triggers": [
        {
          "type": "event",
          "event": "motion_detected"
        }
      ],
      "conditions": [
        {
          "type": "device_state",
          "device_id": "zigbee-sensor-motion-001",
          "attribute": "motion",
          "value": "true"
        }
      ],
      "actions": [
        {
          "type": "device_control",
          "device_id": "zigbee-bulb-001",
          "attribute": "state",
          "value": "on"
        }
      ]
    }
  ],
  "scenarios": [
    {
      "id": "evening-mode",
      "name": "晚间模式",
      "description": "调整灯光为舒适的晚间氛围",
      "icon": "🌙",
      "actions": [
        {
          "type": "device_control",
          "device_id": "zigbee-bulb-001",
          "attribute": "brightness",
          "value": 30
        }
      ]
    }
  ]
}
```

## 高级功能

### 1. 条件表达式

支持复杂的逻辑条件：

```go
{
    Type: iot.ConditionTypeExpression,
    Expression: "device.state == 'on' AND time >= '18:00' AND time <= '23:00'",
}
```

### 2. 动作参数化

动作支持动态参数：

```go
{
    Type: iot.ActionTypeNotification,
    Title: "温度警报",
    Message: "当前温度: {{temperature}}°C，湿度: {{humidity}}%",
}
```

### 3. 场景嵌套

场景可以调用其他场景：

```go
{
    Type: iot.ActionTypeScenario,
    Value: "evening-mode",
}
```

### 4. 延时动作

支持在动作之间添加延时：

```go
{
    Type: iot.ActionTypeDelay,
    Value: 5 * time.Second,
},
{
    Type: iot.ActionTypeDeviceControl,
    DeviceID: "zigbee-bulb-001",
    Attribute: "state",
    Value: "off",
}
```

## 最佳实践

### 1. 规则设计

- ✅ **简单明确**：每个规则只做一件事
- ✅ **单一职责**：避免规则过于复杂
- ✅ **命名清晰**：使用描述性名称
- ✅ **测试验证**：创建后立即测试

### 2. 场景组织

- ✅ **按场景分类**：家庭、安防、节能等
- ✅ **图标标识**：使用直观的图标
- ✅ **描述完整**：说明场景用途
- ✅ **适度数量**：避免场景过多难以管理

### 3. 性能优化

- ✅ **批量操作**：使用批量读写减少通信
- ✅ **条件前置**：在规则层过滤不必要的执行
- ✅ **异步执行**：耗时操作使用异步
- ✅ **事件节流**：高频事件添加节流

### 4. 错误处理

- ✅ **日志记录**：记录规则执行日志
- ✅ **失败重试**：网络错误自动重试
- ✅ **告警通知**：关键规则失败发送通知
- ✅ **降级策略**：部分失败时的备选方案

## 故障排查

### 规则未触发

1. 检查规则是否启用：`rule.Enabled`
2. 检查触发器配置：事件类型、时间等
3. 检查条件是否满足
4. 查看引擎日志

### 场景执行失败

1. 检查设备是否在线
2. 检查设备ID是否正确
3. 检查属性和值是否有效
4. 查看错误消息

### 性能问题

1. 减少不必要的轮询频率
2. 使用事件触发替代定时检查
3. 优化规则条件顺序
4. 启用动作批处理

## 文件结构

```
pkg/iot/
├── workflow_engine.go        # 工作流引擎 ✅
│   ├── WorkflowEngine
│   ├── EventBus
│   ├── TaskScheduler
│   └── AdapterManager
├── ...
examples/iot/
└── workflow_example.go       # 工作流示例 ✅
docs/iot/
└── IOT_WORKFLOW.md           # 工作流文档 ✅
```

## 与Agent集成

### HardwareAgent集成

```go
import (
    "AgentFramework/agent"
    "AgentFramework/pkg/iot"
)

// 创建HardwareAgent
hardwareAgent := agent.NewHardwareAgent()

// 创建IoT工作流引擎
adapterMgr := iot.NewAdapterManager()
engine := iot.NewWorkflowEngine(adapterMgr)

// 将IoT设备注册到HardwareAgent
devices, _ := adapterMgr.ListDevices(ctx)
for _, device := range devices {
    // 注册设备...
}
```

### RealTimeAgent集成

```go
// 使用事件总线实现实时响应
eventBus.Subscribe("device_state_changed", func(event iot.Event) {
    // 触发RealTimeAgent处理
    realtimeAgent.ProcessEvent(event)
})
```

## API参考

### WorkflowEngine

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动引擎 |
| `Stop(ctx)` | 停止引擎 |
| `RegisterWorkflow(workflow)` | 注册工作流 |
| `EnableWorkflow(ctx, id)` | 启用工作流 |
| `DisableWorkflow(id)` | 禁用工作流 |
| `RegisterScenario(scenario)` | 注册场景 |
| `ExecuteScenario(ctx, id)` | 执行场景 |
| `RegisterRule(rule)` | 注册规则 |
| `ListWorkflows()` | 列出工作流 |
| `ListScenarios()` | 列出场景 |
| `ListRules()` | 列出规则 |

### EventBus

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动事件总线 |
| `Stop(ctx)` | 停止事件总线 |
| `Subscribe(eventType, handler)` | 订阅事件 |
| `Publish(event)` | 发布事件 |

### TaskScheduler

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动调度器 |
| `Stop(ctx)` | 停止调度器 |
| `Schedule(id, schedule, task)` | 调度任务 |

## 未来增强

- ✅ 基础工作流引擎
- 🔄 可视化工作流编辑器
- 🔄 规则模板市场
- 🔄 机器学习优化
- 🔄 云端同步和备份
- 🔄 多用户协作
- 🔄 版本控制

---

**状态**: ✅ 完成
**版本**: v1.0.0
**最后更新**: 2026-02-19
**作者**: AgentFramework Team
