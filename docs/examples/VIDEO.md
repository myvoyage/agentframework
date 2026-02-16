# AgentFramework 视频教程

> **AgentFramework 桌面应用视频教程**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [教程概览](#教程概览)
- [环境准备](#环境准备)
- [创建桌面应用](#创建桌面应用)
- [测试应用](#测试应用)
- [构建和运行](#构建和运行)
- [调试技巧](#调试技巧)

---

## 教程概览

本教程将指导您完成创建一个功能完整的 AgentFramework 桌面应用，从环境准备到最终测试。

### 预备知识

- ✅ **Go 语言基础** - 变量、接口、结构
- ✅ **Wails 框架** - 项目结构、组件通信
- ✅ **AgentFramework 核心概念** - Agent、Workflow、Skills、Collaboration
- ✅ **配置管理** - YAML 配置、环境变量

---

## 环境准备

### 系统要求

| 组件 | 版本要求 |
|------|----------|--------|
| **Go** | 1.24 或更高版本 |
| **Node.js** | 18.x 或更高版本 |
| **Wails** | v2.11.0 |
| **Ollama** | （可选）本地 LLM 服务 |

### 安装依赖

```bash
# 1. 克隆项目
git clone https://github.com/myvoyage/agentframework.git
cd agentframework

# 2. 安装 Wails CLI
npm install -g @wails/cli

# 3. 安装前端依赖
npm install

# 4. 验证安装
wails version
```

---

## 创建桌面应用

### 项目结构

```
agent-framework/
├── frontend/          # Wails 前端项目
│   └── wailsjs/      # 主入口
├── app.go          # Wails Go 绑定文件
├── build/           # 构建输出
├── config/          # 配置文件
├── agents/         # Agent 定义
└── main.go          # 应用入口
```

### 前端主程序

**文件**: `frontend/wailsjs/app.go`

```go
package main

import (
    "context"
    "embed"
    "agentframework/frontend/wailsjs"

    "github.com/wails-app"
)

func main() {
    frontend.Start()
}
```

### 后端通信

**Wails Bridge** 提供前后端通信机制

```go
// 前端调用后端
frontend.InvokeAgent("customer_service", "查询订单", callback)

// 后端处理
func HandleCallback(result string) {
    fmt.Println("收到结果:", result)
}
```

---

## 测试应用

### 单元测试

**文件**: `tests/unit/customer_service_test.go`

```go
package agent_test

import (
    "context"
    "testing"
    "agentframework/agent"
)

func TestCustomerService(t *testing.T) {
    // 创建测试上下文
    ctx := context.Background()

    // 创建模拟 Agent
    agent := &mockCustomerServiceAgent{}

    // 测试查询功能
    response, err := agent.Query(ctx, "SELECT * FROM orders WHERE id = ?", nil)
    if err != nil {
        t.Fatal(err)
    }

    // 验证响应
    expected := "ORDER ID: 123, STATUS: processing"
    if response.Content != expected {
        t.Errorf("unexpected response: %s", response.Content)
    }
}
```

---

## 构建和运行

### 开发模式

```bash
# 启动开发模式
wails dev

# 构建应用
wails build
```

### 生产构建

```bash
# 构建生产版本
wails build
```

---

## 调试技巧

### 日志调试

```go
import "log"

func main() {
    // 设置日志级别
    log.SetLevel("debug")

    // 输出调试信息
    log.Debug("Starting application...")

    // 记得所有 Agent 调用
    log.SetReportCaller(true)
}
```

### 错误处理

```go
import (
    "errors"
    "fmt"
    "os"
)

func main() {
    // 优雅关闭
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic: %v\n", r)
        }
    }()

    // 错误包装
    if err := run(); err != nil {
        log.Errorf("Application failed: %w", err)
    }
}
```

---

## 相关文档

- 📘 [快速开始](../quickstart/QUICKSTART.md) - 5 分钟上手
- 📘 [配置指南](../configuration/CONFIGURATION.md) - 详细配置
- 📘 [Agent 概览](../components/agent/overview.md) - Agent 系统
- 📘 [Workflow 概览](../components/workflow/overview.md) - Workflow 系统
- 📘 [Skills 概览](../components/skills/overview.md) - Skills 系统

---

**Made with ❤️ by AgentFramework Team**
