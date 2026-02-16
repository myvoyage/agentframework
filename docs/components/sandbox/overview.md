# Sandbox 系统概览

> **AgentFramework Sandbox 组件文档**
> **版本**: v1.0.0
> **最后更新**: 2025-02-15

---

## 📋 目录

- [系统简介](#系统简介)
- [支持的沙箱](#支持的沙箱)
- [核心组件](#核心组件)
- [安全机制](#安全机制)
- [配置方式](#配置方式)
- [快速开始](#快速开始)

---

## 系统简介

Sandbox 系统是 AgentFramework 的企业级安全执行环境，提供多层沙箱保护，确保代码执行的安全性。

### 核心价值

| 价值 | 说明 |
|------|------|
| **🔒 企业级安全** | 多层隔离和权限控制 |
| **🛡️ 容器隔离** | 基于 Docker 的容器级隔离 |
| **🔍 资源限制** | CPU、内存、网络精确限制 |
| **📊 代码分析** | 增强的静态和动态分析 |
| **🎯 监控审计** | 完整的操作日志和审计追踪 |

### 技术特点

- ✅ **多种沙箱**: 支持代码执行、浏览器自动化、文件操作、身份认证、代理服务、Shell 执行
- ✅ **动态隔离**: 每个沙箱独立运行，互不影响
- ✅ **资源控制**: 精确的 CPU、内存、网络限制
- ✅ **安全策略**: 白名单、路径验证、权限分级

---

## 支持的沙箱

### 1. 代码执行沙箱

**说明**: 支持 4 种编程语言的安全执行

| 语言 | 说明 | 安全特性 |
|------|------|--------|
| **Go** | 容器隔离、yaegi 解释器（428 倍速） |
| **Python** | 容器隔离、限制标准库 |
| **JavaScript** | 容器隔离、Node.js 沙箱 |
| **Bash** | 容器隔离、命令白名单 |

**安全特性**:
- 🔒 **容器隔离**: Docker 容器执行环境
- 🔒 **资源限制**: CPU、内存、执行时间限制
- 🔒 **代码分析**: 增强的静态和动态代码分析
- 🔒 **网络控制**: 默认禁止网络访问
- 🔒 **文件系统隔离**: 独立文件系统访问
- 🔒 **进程隔离**: 容器内无法创建新进程

---

### 2. 浏览器自动化沙箱

**说明**: 基于 ChromeDP 的浏览器操作沙箱

**MCP 工具** (9 个):
- browser_navigate
- browser_click
- browser_fill_form
- browser_type
- browser_hover
- browser_screenshot
- browser_take_screenshot
- browser_execute_script
- browser_get_page_content
- browser_go_back
- browser_go_forward
- browser_close
- browser_get_console_messages
- browser_get_network_requests

**安全特性**:
- 🔒 **DOM 操作隔离**: 只能操作页面 DOM，不能访问其他页面
- 🔒 **文件操作限制**: 限制可访问的文件路径
- 🔒 **输入验证**: 表单输入严格验证
- 🔒 **下载控制**: 文件下载需要用户确认

---

### 3. 文件操作沙箱

**说明**: 安全的文件系统访问

**MCP 工具** (7 个):
- file_read
- file_write
- file_delete
- file_list
- file_mkdir
- file_exists
- file_search

**安全特性**:
- 🔒 **路径验证**: 验证文件路径，防止路径遍历
- 🔒 **白名单**: 只允许访问白名单目录
- 🔒 **权限控制**: 读写权限分离控制
- 🔒 **文件大小限制**: 限制单文件最大 10MB

---

### 4. 身份认证沙箱

**说明**: JWT 和 API Key 管理

**MCP 工具** (7 个):
- auth_jwt_create
- auth_jwt_verify
- auth_jwt_refresh
- auth_api_key_create
- auth_api_key_list
- auth_api_key_delete
- auth_api_key_update

**安全特性**:
- 🔒 **加密存储**: API 密钥加密存储
- 🔒 **JWT 签名**: 支持 JWT token 签名验证
- 🔒 **密钥轮换**: 定期自动轮换 API 密钥
- 🔒 **权限控制**: 基于角色的密钥访问权限

---

### 5. 代理服务沙箱

**说明**: HTTP/HTTPS/SOCKS5 代理支持

**MCP 工具** (6 个):
- proxy_http_create
- proxy_http_delete
- proxy_socks_create
- proxy_socks_delete
- proxy_connect

**安全特性**:
- 🔒 **连接池**: 高效的连接池管理
- 🔒 **超时控制**: 连接和读写超时控制
- 🔒 **流量统计**: 实时监控代理流量使用

---

### 6. Shell 执行沙箱

**说明**: 安全的命令行执行环境

**MCP 工具** (2 个):
- shell_execute
- shell_interactive

**安全特性**:
- 🔒 **沙箱隔离**: 独立的 shell 执行环境
- 🔒 **命令白名单**: 只允许执行白名单命令
- 🔒 **参数验证**: 严格验证命令参数
- 🔒 **输出限制**: 限制输出大小和格式

---

## 核心组件

### SandboxManager

**职责**: 统一管理所有沙箱

```go
type SandboxManager struct {
    codeExecutor    *CodeExecutor
    browser       *BrowserAutomation
    fileOps      *FileOperations
    authService    *AuthService
    proxyService  *ProxyService
    shellService  *ShellService

    mu    sync.RWMutex
}
```

**功能**:
- ✅ 沙箱注册和发现
- ✅ 统一的生命周期管理
- ✅ 资源限制和监控
- ✅ 安全策略执行

---

## 安全机制

### 多层防护

```
┌──────────────────────────────────────────────┐
│            Application Layer                │
│  ┌───────────────┬  ┌──────────────┐ │
│  │ Desktop App │  │ CLI Tools   │  │ HTTP API  │ │
│  └───────────────┘  └──────────────┘  └──────────────┘ │
└──────────────────────────────────────────────┘
└─────────────────────────────────────────────┘
                           ▼
┌──────────────────────────────────────────────────┐
│              Capability Layer               │
│  ┌──────────────┬  ┌──────────────┬  ┌──────────────┐ │
│  │  Tool       │  │ Middleware  │  │ Event Bus  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘ │
└──────────────────────────────────────────────┘
                           ▼
┌──────────────────────────────────────────────┐
│            Infrastructure Layer            │
│  ┌──────────────┬  ┌──────────────┬  ┌──────────────┬  ┌──────────────┐ │
│  │  Checkpoint  │  │ Sandbox     │  Observability│
└──────────────────────────────────────────────┘
                           ▼
└──────────────────────────────────────────────┘
```

### 安全策略

| 策略 | 说明 |
|------|------|--------|
| **默认拒绝** | 所有操作默认拒绝，除非明确允许 |
| **白名单** | 文件操作、网络访问、命令执行 |
| **资源限制** | CPU、内存、磁盘严格限制 |
| **审计日志** | 所有操作完整记录和可追溯 |

---

## 配置方式

### YAML 配置

**文件**: `sandbox.yaml`

```yaml
sandbox:
  enabled: true
  backend: "docker"    # docker, native

  # 代码执行
  code:
    enabled: true
    languages: ["go", "python", "javascript", "bash"]
    timeout: 60
    memory_limit: "512m"
    cpu_limit: "1.0"
    allowed_operations: ["read", "write"]
    network_access: false
    resource_limits:
      max_files: 100
      max_file_size: 10485760  # 10MB

  # 浏览器自动化
  browser:
    enabled: true
    headless: true
    download_dir: "/tmp/sandbox/downloads"
    allowed_domains: ["example.com"]

  # 文件操作
  file:
    enabled: true
    allowed_paths: ["/tmp", "/home/user/sandbox"]
    max_file_size: 10485760  # 10MB

  # 身份认证
  auth:
    enabled: true
    jwt_secret: "your-jwt-secret"
    api_keys:
      encryption_key: "your-encryption-key"

  # 代理服务
  proxy:
    enabled: false
```

---

## 快速开始

### 创建沙箱配置

```go
package main

import (
    "context"
    "fmt"
    "log"

    "agentframework/sandbox"
)

func main() {
    ctx := context.Background()

    // 创建沙箱管理器
    manager := sandbox.NewSandboxManager(&sandbox.SandboxConfig{
        Code: &sandbox.CodeConfig{
            Enabled:      true,
            Languages:    []string{"go"},
        },
    })

    // 获取代码执行器
    executor := manager.GetExecutor("code")

    // 执行代码
    result, err := executor.Execute(ctx, &sandbox.ExecutionRequest{
        Language: "go",
        Code:     `package main

        import "fmt"

        func main() {
            fmt.Println("Hello from sandbox!")
        }`,
    })

    if err != nil {
        log.Fatal(err)
    }
}
```

---

## 相关文档

- 📘 [Agent 概览](../agent/overview.md) - Agent 系统概览
- 📘 [Workflow 概览](../workflow/overview.md) - Workflow 系统概览
- 📘 [配置指南](../../configuration/CONFIGURATION.md) - 详细配置说明
- 📘 [最佳实践](../../guides/best-practices/BEST_PRACTICES.md) - 开发指南

---

**Made with ❤️ by AgentFramework Team**
