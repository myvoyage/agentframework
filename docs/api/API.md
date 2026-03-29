# API 参考

> AgentFramework Gateway API  
> 默认端口：`18640`

---

## 目录

- [HTTP API](#http-api)
- [WebSocket API](#websocket-api)
- [错误码](#错误码)

---

## HTTP API

所有 HTTP 接口支持 CORS，认证 Token 通过 `Authorization: Bearer <token>` 或自定义 Header `openclaw-auth: <token>` 传递（可选，默认无认证）。

### 健康检查

```http
GET /health
```

响应示例：

```json
{
  "status": "ok",
  "version": "2.0.0",
  "uptimeMs": 12345,
  "timestamp": 1711699200,
  "checks": {
    "model": { "status": "ok" },
    "channels": { "status": "ok" }
  },
  "channels": [
    { "name": "lark", "type": "lark", "connected": true, "healthy": true }
  ],
  "agents": [
    { "name": "default_chat", "model": "default", "running": false }
  ],
  "memoryUsageMB": 48.5,
  "cpuPercent": 1.2
}
```

### 状态查询

```http
GET /status
```

返回 Gateway 当前状态摘要。

---

### OpenAI 兼容接口

Gateway 实现了 OpenAI Chat Completions 格式，可直接用 OpenAI SDK 接入。

#### Chat Completions

```http
POST /v1/chat/completions
Content-Type: application/json
```

请求体（OpenAI 格式）：

```json
{
  "model": "default",
  "messages": [
    { "role": "user", "content": "你好" }
  ],
  "stream": false,
  "temperature": 0.7,
  "max_tokens": 1024
}
```

响应（非流式）：

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1711699200,
  "model": "default",
  "choices": [
    {
      "index": 0,
      "message": { "role": "assistant", "content": "你好！有什么可以帮你的？" },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  }
}
```

响应（流式，`"stream": true`）：SSE 格式，每行 `data: {...}`，以 `data: [DONE]` 结束。

#### Responses

```http
POST /v1/responses
Content-Type: application/json
```

类似 Chat Completions，使用 `input` 字段替代 `messages`：

```json
{
  "model": "default",
  "input": "你好"
}
```

---

### 工具调用

```http
POST /tools/invoke
Content-Type: application/json
```

请求体：

```json
{
  "tool": "web_search",
  "input": "{ \"query\": \"AgentFramework\" }"
}
```

响应：

```json
{
  "result": "搜索结果...",
  "elapsed_ms": 1234
}
```

---

## WebSocket API

### 连接地址

```
ws://localhost:18640/
```

### 协议概述

全双工 JSON 帧协议，三种帧类型：

| 帧类型 | 方向 | 说明 |
|--------|------|------|
| `req` | 客户端 → 服务端 | 请求 |
| `res` | 服务端 → 客户端 | 响应（对应某个 req） |
| `event` | 服务端 → 客户端 | 服务端主动推送 |

**重要规则**：
- 第一帧**必须**是 `connect` 方法的 `req` 帧
- 非 JSON 或首帧不是 `connect` → 服务端立即关闭连接
- 有副作用的方法（如 `agent.run`）**必须**带 `idempotencyKey` 防重放

### 帧结构

**请求帧（req）**：

```json
{
  "type": "req",
  "id": "req-001",
  "method": "connect",
  "params": { ... },
  "idempotencyKey": "uuid-xxx"
}
```

**响应帧（res）**：

```json
{
  "type": "res",
  "id": "req-001",
  "ok": true,
  "payload": { ... }
}
```

错误时：

```json
{
  "type": "res",
  "id": "req-001",
  "ok": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "missing required field: agent",
    "retryable": false
  }
}
```

**事件帧（event）**：

```json
{
  "type": "event",
  "event": "agent.token",
  "payload": { "token": "你好" },
  "seq": 1,
  "stateVersion": 1
}
```

---

### 方法列表

#### `connect` — 握手（首帧必须）

请求：

```json
{
  "type": "req",
  "id": "req-001",
  "method": "connect",
  "params": {
    "minProtocol": 1,
    "maxProtocol": 1,
    "client": {
      "id": "client-unique-id",
      "displayName": "My Client",
      "version": "1.0.0",
      "platform": "web"
    },
    "caps": {
      "streaming": true
    },
    "auth": {
      "token": "optional-auth-token"
    }
  }
}
```

成功响应 payload：

```json
{
  "protocolVersion": 1,
  "gatewayVersion": "2.0.0",
  "uptimeMs": 12345,
  "stateVersion": 1,
  "health": { "status": "ok" },
  "policy": {
    "maxPayload": 1048576,
    "maxBufferedBytes": 10485760,
    "tickIntervalMs": 30000
  }
}
```

---

#### `agent.run` — 执行 Agent

```json
{
  "type": "req",
  "id": "req-002",
  "method": "agent.run",
  "params": {
    "agent": "default_chat",
    "input": "帮我写一首诗",
    "sessionKey": "default:websocket:user123",
    "stream": true
  },
  "idempotencyKey": "uuid-xxx"
}
```

**流式模式**下，服务端在最终 `res` 之前会推送多个事件：

```json
// 开始推理
{"type":"event","event":"agent.started","payload":{"runId":"run-xxx"}}

// 流式 token（流式模式）
{"type":"event","event":"agent.token","payload":{"token":"春风"}}
{"type":"event","event":"agent.token","payload":{"token":"送暖"}}

// 工具调用（如有）
{"type":"event","event":"agent.tool_call","payload":{"tool":"web_search","input":"{...}"}}
{"type":"event","event":"agent.tool_result","payload":{"tool":"web_search","result":"..."}}

// 完成
{"type":"event","event":"agent.completed","payload":{"runId":"run-xxx","elapsed_ms":1234}}
```

最终 `res` 帧：

```json
{
  "type": "res",
  "id": "req-002",
  "ok": true,
  "payload": {
    "runId": "run-xxx",
    "content": "春风送暖入屠苏...",
    "role": "assistant",
    "elapsed_ms": 1234
  }
}
```

---

#### `agent.status` — 查询 Agent 状态

```json
{
  "type": "req",
  "id": "req-003",
  "method": "agent.status",
  "params": { "agent": "default_chat" }
}
```

响应 payload：

```json
{
  "name": "default_chat",
  "model": "default",
  "running": false,
  "tools": ["web_search"]
}
```

---

#### `ping` — 心跳

```json
{"type":"req","id":"req-004","method":"ping","params":{}}
```

响应：

```json
{"type":"res","id":"req-004","ok":true,"payload":{"pong":true,"ts":1711699200}}
```

---

## 错误码

| 错误码 | 说明 | 是否可重试 |
|--------|------|-----------|
| `INVALID_REQUEST` | 参数错误或 Schema 验证失败 | 否 |
| `UNAUTHORIZED` | 认证失败 | 否 |
| `NOT_FOUND` | 资源不存在（如 Agent 名称错误） | 否 |
| `AGENT_TIMEOUT` | Agent 执行超时（默认 5min） | 是 |
| `UNAVAILABLE` | Gateway 正在关闭或依赖不可用 | 是 |
| `INTERNAL_ERROR` | 内部错误 | 是（建议延迟后重试） |

---

## 示例：Python 客户端

```python
import asyncio
import json
import websockets
import uuid

async def chat():
    async with websockets.connect("ws://localhost:18640/") as ws:
        # 第一步：握手
        await ws.send(json.dumps({
            "type": "req",
            "id": "req-001",
            "method": "connect",
            "params": {
                "minProtocol": 1,
                "maxProtocol": 1,
                "client": {"id": "python-client", "platform": "python"},
                "caps": {"streaming": True}
            }
        }))
        resp = json.loads(await ws.recv())
        assert resp["ok"], f"连接失败: {resp['error']}"

        # 第二步：发送消息
        await ws.send(json.dumps({
            "type": "req",
            "id": "req-002",
            "method": "agent.run",
            "params": {
                "agent": "default_chat",
                "input": "你好！",
                "stream": True
            },
            "idempotencyKey": str(uuid.uuid4())
        }))

        # 第三步：接收流式响应
        while True:
            msg = json.loads(await ws.recv())
            if msg["type"] == "event" and msg["event"] == "agent.token":
                print(msg["payload"]["token"], end="", flush=True)
            elif msg["type"] == "res":
                print()  # 换行
                break

asyncio.run(chat())
```

---

## 示例：curl HTTP

```bash
# 健康检查
curl http://localhost:18640/health

# 对话
curl -X POST http://localhost:18640/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"default","messages":[{"role":"user","content":"你好"}]}'
```
