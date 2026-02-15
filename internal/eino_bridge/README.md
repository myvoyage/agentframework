# eino_bridge

> 桥接模块 - 为 AgentFramework 提供跨进程通信和 MCP 协议适配能力

## 概述

`eino_bridge` 是一个架构层面的抽象层，用于解决 **PipelineEngine** 与外部组件之间的通信问题。它提供统一的接口抽象，支持多种通信协议（HTTP、TCP），并实现了类 MCP (Model Context Protocol) 的通信规范。

### 设计目标

- **解耦通信与业务逻辑**：将通信协议与核心引擎分离
- **支持分布式部署**：为微服务架构预留接口
- **协议统一标准化**：提供标准化的工具调用接口
- **灵活部署模式**：支持单机/客户端/服务器多种模式
- **安全可控**：内置 JWT 验证机制

---

## 架构设计

### 模块结构

```
internal/eino_bridge/
├── eino_bridge.go              # 核心 BridgeClient 实现 (build: eino)
├── eino_bridge_fallback.go     # 非 eino 模式的空实现 (build: !eino)
├── eino_bridge_integration.go  # 桥接集成与配置管理 (build: eino)
├── eino_rpc.go                 # RPC 接口定义
├── eino_rpc_impl.go            # Mock 实现（本地测试）
├── eino_rpc_http.go            # HTTP RPC 服务器 (build: eino)
├── eino_rpc_client_http.go     # HTTP RPC 客户端 (build: eino)
├── eino_rpc_security.go        # JWT 安全验证
└── README.md                   # 本文档
```

### 通信模式

```
                    eino_bridge
┌─────────────────────────────────────────────────────────────┐
│  ┌───────────┐     ┌───────────┐     ┌───────────┐         │
│  │ 本地模式    │     │ HTTP 模式  │     │ TCP 模式   │         │
│  │(In-Process)│     │(REST API) │     │(Raw RPC)  │         │
│  └─────┬─────┘     └─────┬─────┘     └─────┬─────┘         │
│        └────────────────────┴──────────────────┘            │
│                              │                               │
│                      ┌───────▼───────┐                       │
│                      │ BridgeClient │                       │
│                      └───────┬───────┘                       │
│                              │                               │
│                      ┌───────▼───────┘                       │
│                      │BridgeEngine  │                       │
│                      └───────┬───────┘                       │
└──────────────────────────────┼───────────────────────────────┘
                               │
                               ▼
                      ┌───────────────────┐
                      │  PipelineEngine   │
                      └───────────────────┘
```

---

## 快速开始

### 构建标签

```bash
# 启用 eino 功能
go build -tags eino

# 标准构建（使用 fallback 实现）
go build
```

### 基本使用

```go
import einobridge "AgentFramework/internal/eino_bridge"

// 1. 初始化桥接
if err := einobridge.InitBridge(); err != nil {
    log.Fatal(err)
}

// 2. 设置引擎
einobridge.SetBridgeEngine(engine)

// 3. 创建客户端
client := &einobridge.BridgeClient{}

// 4. 调用工具
data, err := client.InvokeTool(ctx, "tool_name", params, context)
```

### HTTP 服务器模式

```go
config := &einobridge.BridgeConfig{
    Protocol:   "http",
    Port:       8080,
    EnableHTTP: true,
}
einobridge.SetBridgeConfig(config)
einobridge.SetBridgeEngine(engine)

if err := einobridge.InitBridge(); err != nil {
    log.Fatal(err)
}
```

### HTTP 客户端模式

```go
client := einobridge.NewHTTPRPCClient("http://localhost:8080")

req := pe.MCPInvokeToolRequest{
    Tool:    "my_tool",
    Params:  map[string]interface{}{"arg": "value"},
    Context: map[string]interface{}{},
}
resp, err := client.InvokeTool(req)
```

---

## 配置

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `EINO_BRIDGE_PORT` | HTTP 服务端口 | `8080` |
| `EINO_BRIDGE_PROTOCOL` | 通信协议 (`http`/`tcp`/`auto`) | `auto` |
| `EINO_BRIDGE_HOST` | TCP 服务器地址 | `localhost:8080` |
| `EINO_BRIDGE_ENABLE_HTTP` | 启用 HTTP 服务 | `true` |

---

## API 接口

### BridgeClient

| 方法 | 描述 |
|------|------|
| `InvokeTool(ctx, name, params, context)` | 调用指定工具 |
| `RunPipeline(ctx, spec)` | 运行流水线 |

### HTTP REST API

#### POST `/invoke_tool`

**请求：**
```json
{
  "tool": "tool_name",
  "params": {"key": "value"},
  "context": {},
  "version": "1.0"
}
```

**响应：**
```json
{
  "success": true,
  "data": {"result": "value"},
  "error": ""
}
```

#### POST `/run_pipeline`

**请求：**
```json
{
  "pipeline": {
    "name": "my_pipeline",
    "stages": [...]
  }
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "results": [...]
  },
  "error": ""
}
```

---

## 数据结构

```go
// MCPInvokeToolRequest 请求结构
type MCPInvokeToolRequest struct {
    Tool    string                 `json:"tool"`
    Params  map[string]interface{} `json:"params"`
    Context map[string]interface{} `json:"context"`
    Version string                 `json:"version"`
}

// MCPInvokeToolResponse 响应结构
type MCPInvokeToolResponse struct {
    Success bool                   `json:"success"`
    Data    map[string]interface{} `json:"data"`
    Error   string                 `json:"error"`
}
```

---

## 设计原则

| 原则 | 实现 |
|------|------|
| **单一职责 (S)** | 每个文件专注单一功能 |
| **开闭原则 (O)** | 通过接口扩展 |
| **依赖倒置 (D)** | 依赖抽象接口 |
| **KISS** | 简洁的 API 设计 |
| **DRY** | 配置管理统一 |

---

## 路线图

- [ ] 完整的 TCP RPC 实现
- [ ] JWT 签名验证
- [ ] WebSocket 支持
- [ ] 连接池管理
- [ ] 指标与监控

---

## 注意事项

1. 使用 `-tags eino` 构建以启用完整功能
2. `BridgeEngine` 设置后不应并发修改
3. HTTP 模式会占用指定端口
