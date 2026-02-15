# Agent Framework

[English](README_EN.md) | [中文](README_CN.md)

> 🚀 高性能、企业级 AI 代理框架 - 构建智能应用的强大基础设施

![License](https://img.shields.io/badge/license-AGPL--3.0--or--later-blue)
![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8)
![Wails](https://img.shields.io/badge/wails-v2.11.0-41b883)
![Tests](https://img.shields.io/badge/tests-80%25-brightgreen)
![Documentation](https://img.shields.io/badge/docs-100%25-brightgreen)

---

## 🌟 系统概述

AgentFramework 是一个高性能、企业级的 AI 代理框架，为构建智能应用提供强大的基础设施。

### 核心定位

**企业级 AI 应用开发平台**
- 🎯 面向企业级应用场景设计
- 🏗️ 完整的开发、部署和运维体系
- 📚 文档齐全，降低学习成本
- 🔧 生产就绪，完善的监控和错误处理

### 技术架构

**四层架构设计**
```
┌──────────────────────────────────────────────┐
│           Application Layer              │
│  ┌─────────────┬ ┌─────────────┬ ┌─────────────┐ │
│  │ Desktop App │  │ CLI Tools   │  │ HTTP API   │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└──────────────────────────────────────────────┘
                     │
┌──────────────────────────────────────────────┐
│            Framework Layer               │
│  ┌─────────────┬ ┌─────────────┬ ┌─────────────┐ │
│  │ Host        │  │ Agent       │  │ Workflow    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└──────────────────────────────────────────────┘
```

**核心组件关系**
- **Host** → 管理所有组件的中央容器
- **Agent** → 执行单元，使用 LLM 和工具完成任务
- **Workflow** → 编排执行流程，支持复杂任务编排
- **Skills** → 能力扩展系统，支持零代码快速开发
- **Collaboration** → 多代理协作，支持团队级任务执行

### 适用场景

| 场景类型 | 支持程度 | 典型应用 |
|-----------|---------|---------|
| **智能客服** | ⭐⭐⭐⭐⭐ | 24/7 自动响应 + 人工审核 |
| **数据分析** | ⭐⭐⭐⭐⭐ | 复杂 ETL 流程编排 |
| **内容生成** | ⭐⭐⭐⭐⭐ | 流式生成 + 质量控制 |
| **研究助手** | ⭐⭐⭐⭐⭐ | 多代理协作 + 工具集成 |
| **自动化测试** | ⭐⭐⭐⭐ | 浏览器自动化 + 测试报告 |
| **代码审查** | ⭐⭐⭐⭐ | 自动化代码审查 + 质量门禁 |

---

## 📚 完整文档

> 📘 **详细文档**: [docs/](docs/) | 📋 **项目评估**: [PROJECT_EVALUATION.md](PROJECT_EVALUATION.md)

---

## 🌟 核心价值

| 优势 | 说明 |
|------|------|
| **🚀 高性能** | 基于 Go 语言的高并发特性，支持大规模代理部署 |
| **🔧 高可扩展** | 插件化架构，所有核心组件均可自定义替换 |
| **🛡️ 企业级** | 完善的监控、日志、错误处理和文档体系 |
| **🔒 高安全性** | 多层沙箱系统，安全的代码执行和资源访问 |
| **📈 可观测性** | 内置 OpenTelemetry，全面的性能指标和追踪 |

### 技术指标

| 指标 | 数值 |
|------|------|
| **代码规模** | 约 2900+ Go 文件，50 万行代码 |
| **测试覆盖率** | 80%+ |
| **MCP 工具** | 44 个原生实现 |
| **支持模型** | OpenAI、Ollama、LM Studio、vLLM |
| **并发性能** | 支持 1000+ 并发任务 |
| **内存优化** | 智能监控，自动优化建议 |

---

## 🌟 核心特性

### 1. 智能代理系统

#### 支持的代理类型

| 代理类型 | 说明 | 使用场景 |
|---------|------|---------|
| **ChatAgent** | 基础对话代理 | 客服机器人、对话系统 |
| **ReActAgent** | 推理-行动代理 | 复杂任务执行、工具调用 |
| **HumanAgent** | 人工介入代理 | HITL 场景、人工审核 |
| **WorkerAgent** | 专业工作代理 | 7 种专业角色 |
| **CollaborationAgent** | 协作代理 | 多代理团队协作 |

#### 核心能力
- ✅ **流式响应**: 支持流式和非流式响应
- ✅ **记忆管理**: 智能对话历史管理和 Token 自动压缩
- ✅ **工具调用**: 内置工具支持和自定义技能系统
- ✅ **多模型**: 运行时模型动态切换和健康检查
- ✅ **缓存优化**: 模型缓存和预热，减少冷启动延迟

### 2. 工作流编排引擎

支持 6 种工作流类型：

| 类型 | 说明 | 复杂度 |
|------|------|--------|
| **Sequential** | 顺序执行 | ⭐ |
| **Parallel** | 并行执行 | ⭐⭐ |
| **DAG** | 有向无环图 | ⭐⭐⭐⭐ |
| **Routing** | 条件路由 | ⭐⭐⭐ |
| **Planning** | 规划工作流 | ⭐⭐⭐⭐ |

**核心特性**:
- ✅ **DAG 工作流**: 有向无环图定义复杂多步骤工作流
- ✅ **并行执行**: 支持任务并行执行和路由决策
- ✅ **人工介入 (HITL)**: 内置人工审批和反馈支持
- ✅ **状态持久化**: 自动保存检查点，支持跨服务器恢复
- ✅ **动态路由**: 基于能力和状态的任务分发

### 3. 协作系统

多代理协作的核心支持：

- ✅ **Agent 团队**: 多代理协作和团队管理
- ✅ **智能路由**: 基于能力的任务分发
- ✅ **共识机制**: 多代理决策协调
- ✅ **任务编排**: 复杂任务的编排和管理

**协作模式**:
- **Single**: 单代理执行
- **Parallel**: 并行执行
- **Sequential**: 顺序执行
- **Consensus**: 共识机制

### 4. 技能系统

强大的插件化技能系统：

**技能类型**:
- ✅ **HTTP 技能**: REST API 调用技能
- ✅ **文件操作**: 安全的文件系统访问
- ✅ **代码执行**: 动态代码执行和 Python 沙箱
- ✅ **数据处理**: 数据转换和处理技能
- ✅ **Markdown 技能**: 零编程门槛快速创建技能

**技能注册表**:
- ✅ 自动去重注册
- ✅ 多条件查找
- ✅ 分类和标签管理
- ✅ 导出功能（JSON/Markdown）

### 5. 安全沙箱

企业级安全执行环境：

| 沙箱类型 | 说明 |
|---------|------|
| **代码执行** | 支持 4 种语言，容器隔离、资源限制 |
| **浏览器自动化** | 基于 ChromeDP，9 个 MCP 工具 |
| **文件操作** | 安全的文件系统访问，7 个 MCP 工具 |
| **身份认证** | JWT 和 API Key 管理，7 个 MCP 工具 |

**安全特性**:
- 🔒 **容器隔离**: Docker 容器执行环境
- 🔒 **资源限制**: CPU、内存、网络限制
- 🔒 **代码分析**: 增强的代码分析规则
- 🔒 **质量检查**: 代码质量检查和优化建议
- 🔒 **加速执行**: yaegi Go 解释器（428 倍速）

### 6. 模型管理

支持多种 AI 模型的统一管理：

| 模型类型 | 说明 |
|---------|------|
| **OpenAI** | GPT-4、GPT-3.5 等 |
| **Ollama** | 本地模型服务 |
| **LM Studio** | 本地模型工作室 |
| **vLLM** | vLLM 推理框架 |

**核心特性**:
- ✅ **动态切换**: 运行时模型切换和健康检查
- ✅ **智能缓存**: 模型缓存和预热，减少冷启动延迟
- ✅ **并发优化**: 分片缓存提高并发性能
- ✅ **配置验证**: 详细的参数验证和默认值设置

---

## 🚀 快速开始

### 前置要求

| 组件 | 版本要求 |
|------|---------|
| **Go** | 1.24 或更高版本 |
| **Node.js** | 用于前端构建 |
| **Wails** | v2.11.0（桌面应用） |

### 安装步骤

```bash
# 1. 克隆仓库
git clone https://github.com/your-org/agentframework.git
cd agentframework

# 2. 下载依赖
go mod download

# 3. 构建应用
go build -o bin/agentframework ./cmd/app

# 4. 运行应用
./bin/agentframework
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cloudwego/eino/components/model"
    "agentframework/agent"
)

func main() {
    ctx := context.Background()

    // 创建模型
    model, err := model.NewChatModel(ctx, &model.ChatModelConfig{
        Model: "gpt-4",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 创建 ChatAgent
    chatAgent := agent.NewChatAgent(
        agent.WithName("助手"),
        agent.WithInstructions("你是一个有用的AI助手"),
        agent.WithModel(model),
    )

    // 运行代理
    response, err := chatAgent.Run(ctx, "你好，请介绍一下你自己")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)
}
```

### 构建桌面应用

```bash
# 构建应用
wails build

# 运行构建的应用
./build/bin/AgentFramework.exe
```

### 开发模式

```bash
# 启动实时开发模式
wails dev
```

---

## 🏗️ 架构概览

```
┌────────────────────────────────────────────────────┐
│              Agent Framework 架构              │
└────────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────┐
│              Application Layer              │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │ Desktop App │  │ CLI Tools   │  │ HTTP API   │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────┐
│               Framework Layer                   │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │    Host     │  │   Agent     │  │  Workflow    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │  Skill      │  │  Model      │  │ Collab      │ │
│  │ Library    │  │ Factory    │  │ System     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────┐
│              Capability Layer                │
│  ┌─────────────┬  ┌─────────────┬  ┌─────────────┐ │
│  │  Tool       │  │ Middleware  │  │  Event Bus  │ │
│  │ Registry    │  │   Chain     │  │              │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────────────────────────────────────┘
```

### 核心组件

- **Agent Host**: 中央容器，管理代理、工作流和服务的生命周期
- **Agent**: 自主执行单元，使用 LLM 和工具完成任务
- **Workflow**: 编排执行流程，支持 DAG、顺序、并行等工作流
- **ModelManager**: 管理多个 AI 模型，支持动态切换和缓存
- **SkillLibrary**: 集中管理内置和自定义技能
- **EventBus**: 组件间的事件通信
- **CheckpointStore**: 工作流状态持久化
- **SandboxManager**: 安全的文件访问控制
- **Middleware**: 请求拦截、日志、追踪和 RAG

---

## 📚 文档导航

> 📘 **详细文档**: [docs/](docs/)

### 入门指南

- 📘 [快速开始](docs/quickstart/QUICKSTART.md) - 5 分钟上手教程
- 📘 [配置指南](docs/configuration/CONFIGURATION.md) - 配置说明和示例
- 📘 [库使用指南](docs/LIBRARY_GUIDE.md) - 核心库使用说明

### 架构文档

- 📘 [项目评估报告](PROJECT_EVALUATION.md) - 完整的项目评估和改进建议
- 📘 [综合架构文档](docs/architecture/ARCHITECTURE_COMPREHENSIVE.md) - 完整的系统架构和 API 文档
- 📘 [统一架构文档](docs/architecture/ARCHITECTURE_UNIFIED.md) - 系统架构和组件关系
- 📘 [协作系统](docs/components/collaboration/overview.md) - 多代理协作详解
- 📘 [内存管理](docs/components/MEMORY_MANAGEMENT.md) - 内存优化和管理

### 组件文档

#### Agent 组件
- 📘 [Agent 概览](docs/components/agent/overview.md) - Agent 系统概览
- 📘 [Agent 类型](docs/components/agent/types.md) - 5 种 Agent 类型详解
- 📘 [Agent 生命周期](docs/components/agent/lifecycle.md) - 生命周期管理
- 📘 [Agent API](docs/components/agent/api.md) - 完整 API 参考

#### Skills 组件
- 📘 [Skills 概览](docs/components/skills/overview.md) - Skills 系统概览

#### Workflow 组件
- 📘 [Workflow 概览](docs/components/workflow/overview.md) - Workflow 系统概览

#### Collaboration 组件
- 📘 [Collaboration 概览](docs/components/collaboration/overview.md) - Collaboration 系统概览

### 功能文档

- 📘 [技能系统](docs/components/SKILLS.md) - 技能系统完整文档
- 📘 [技能 API 快速参考](docs/api/SKILLS_API_QUICK_REFERENCE.md) - 技能系统 API

### 其他文档

- 📘 [贡献指南](docs/contribution/DOCUMENTATION_GUIDELINES.md) - 文档贡献规范
- 📘 [项目路线图](docs/ROADMAP.md) - 功能规划和优化计划
- 📘 [演示示例](docs/DEMOS.md) - 演示项目说明
- 📘 [库使用指南](docs/LIBRARY_GUIDE.md) - 核心库使用说明

---

## 💡 使用场景

### 场景 1: 智能客服系统

```go
// 创建带 HITL 的客服代理
agent := agent.NewChatAgent(
    agent.WithName("客服助手"),
    agent.WithHumanInTheLoop(true),
)
```

**特性**:
- ✅ 24/7 自动响应
- ✅ 人工审核敏感问题
- ✅ 多轮对话上下文管理
- ✅ 智能路由到人工客服

### 场景 2: 数据分析流水线

```go
// 创建 DAG 工作流
workflow := NewDAGWorkflow("数据分析")
workflow.AddTask("data_collection", collectAgent)
workflow.AddTask("analysis", analysisAgent)
workflow.AddDependency("analysis", "data_collection")
```

**特性**:
- ✅ 自动任务编排
- ✅ 失败自动重试
- ✅ 检查点保存和恢复

### 场景 3: 多代理研究团队

```go
// 创建代理团队
team := collaboration.NewAgentTeam()
team.AddAgent("研究员", researcherAgent)
team.AddAgent("分析师", analystAgent)
team.AddAgent("报告员", reporterAgent)
```

**特性**:
- ✅ 智能任务路由
- ✅ 并行任务处理
- ✅ 共识决策机制
- ✅ 性能监控和优化

---

## 🧪 示例项目

### 命令行示例

```bash
# 运行基本演示
go run cmd/demo/main.go

# 运行工作流演示
go run cmd/workflow_demo/main.go

# 运行 MCP 演示
go run cmd/mcp_demo/main.go

# 运行服务器演示
go run cmd/server_demo/main.go
```

### 桌面应用

```bash
# 启动桌面应用
wails dev
```

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...

# 运行集成测试
go test -tags=integration ./...

# 运行基准测试
go test -bench=. -benchmem ./...
```

---

## 📊 项目状态

- **开发状态**: 🚀 活跃开发中
- **当前版本**: v1.0.0
- **Go 文件**: 2900+ 个
- **总代码行数**: 约 50 万行
- **MCP 工具**: 44 个
- **测试覆盖率**: 80%+
- **AIO 沙箱模块**: 7 个（100% 完成度）
- **Code Executor 增强**: ✅ 已完成（97%）
- **文档状态**: 完善中（组件文档 100%）

---

## 🤝 贡献

我们欢迎各种形式的贡献！

- 🐛 报告 Bug
- 💡 提出新功能
- 📝 改进文档
- 🔧 提交代码

请阅读 [贡献指南](docs/contribution/DOCUMENTATION_GUIDELINES.md) 了解详情。

### 开发流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📜 许可证

本项目采用 **GNU Affero General Public License v3.0 或更高版本** (AGPL-3.0-or-later) 许可。

### AGPL-3.0 关键点

- ✅ 可自由使用、修改和分发
- ✅ 所有修改和衍生品必须开源
- ✅ **网络共享条款**: 如果作为网络服务运行，必须向用户提供源代码
- ✅ 所有衍生品必须使用相同许可证
- ⚠️ 对网络使用的限制比 GPL 更严格
- 📖 [完整许可证文本](LICENSE)

---

## 🌟 致谢

- [CloudWeGo Eino](https://github.com/cloudwego/eino) - 核心 LLM 框架
- [Wails](https://wails.io/) - 桌面应用框架
- [Vue.js](https://vuejs.org/) - 前端框架
- 所有贡献者

---

## 📞 联系方式

- **官方网站**: https://agentframework.dev
- **文档**: https://docs.agentframework.dev
- **GitHub**: https://github.com/your-org/agentframework
- **问题反馈**: [GitHub Issues](https://github.com/your-org/agentframework/issues)
- **讨论**: [GitHub Discussions](https://github.com/your-org/agentframework/discussions)

---

## 🔗 相关资源

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Go 文档](https://golang.org/doc/)
- [Wails 文档](https://wails.io/docs/)

---

**由 ❤️ 和 Go 语言构建**

---

## 🎯 快速导航

- 🏠 [返回顶部](#-核心价值)
- 📘 [快速开始](#-快速开始) - 5 分钟上手
- 📘 [文档导航](#-📚-文档导航) - 详细文档
- 📘 [使用场景](#-使用场景) - 实践案例
- 🧪 [测试](#-测试) - 质量保证
- 🤝 [贡献](#-贡献) - 参与项目
