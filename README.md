# AgentFramework

<div align="center">

**🚀 高性能、企业级 Go 语言 AI Agent 框架**

[![License](https://img.shields.io/badge/License-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/myvoyage/agentframework)](https://goreportcard.com/report/github.com/myvoyage/agentframework)
[![GoDoc](https://pkg.go.dev/badge/github.com/myvoyage/agentframework)](https://pkg.go.dev/github.com/myvoyage/agentframework)
[![License Check](https://img.shields.io/badge/license%20coverage-100%25-brightgreen.svg)](.github/workflows/license-check.yml)

[English](README_EN.md) | 简体中文

</div>

---

## 📋 项目概述

AgentFramework 是一个功能完善、架构先进的企业级 AI Agent 框架，支持多渠道通信、IoT 设备管理、工作流自动化等丰富功能。

### 核心特性

- 🤖 **多种 Agent 类型** - ReAct Agent、Skill Agent、协作 Agent
- 🔄 **工作流引擎** - DAG、并行、顺序、图工作流
- 📱 **多渠道通信** - Telegram、Discord、Slack、飞书、企业微信、钉钉、QQ
- 🌐 **IoT 协议支持** - Zigbee、Thread、Z-Wave、NearLink
- 🛠️ **丰富的工具集** - 沙箱环境、MCP 集成、代码执行
- 💾 **多层记忆系统** - L1-L2 上下文、智能压缩
- 🔌 **插件化架构** - 灵活的技能系统

---

## 📊 项目统计

| 指标 | 数值 |
|------|------|
| 代码行数 | 208,770+ |
| Go 源文件 | 551 |
| 测试文件 | 147 |
| 测试覆盖率 | 65-70% |
| Go 版本 | 1.25+ |
| 许可证 | AGPL-3.0-or-later |

---

## 🎯 核心功能

### 1. Agent 系统

```go
// ReAct Agent - 推理+行动
agent := NewReActAgent(
    WithModel("gpt-4"),
    WithTools(tools),
    WithMaxIterations(10),
)

// Skill Agent - 技能封装
skillAgent := NewSkillAgent(
    WithSkills(skills),
    WithExecutor(executor),
)

// 协作 Agent - 多 Agent 团队
team := NewAgentTeam(
    WithAgents(agents),
    WithConsensus(consensus),
)
```

### 2. 工作流引擎

```go
// DAG 工作流
dag := NewDAGWorkflow()
dag.AddNode("task1", task1)
dag.AddNode("task2", task2)
dag.AddEdge("task1", "task2")
dag.Execute(ctx)

// 并行工作流
parallel := NewParallelWorkflow(
    task1, task2, task3,
)
parallel.Execute(ctx)

// 顺序工作流
sequential := NewSequentialWorkflow(
    step1, step2, step3,
)
sequential.Execute(ctx)
```

### 3. IoT 设备管理

```go
// Zigbee 设备
zigbee := NewZigbeeAdapter(mqttConfig)
zigbee.Start(ctx)
devices := zigbee.Discover()

// Thread 设备
thread := NewThreadAdapter(borderRouter)
thread.Start(ctx)

// 控制设备
device.Set(true)  // 开启
device.SetLevel(80)  // 设置亮度 80%
```

### 4. 多渠道通信

```go
// 初始化渠道管理器
manager := channels.NewManager()

// 添加 Telegram
manager.Register("telegram", &channels.TelegramAdapter{
    Token: os.Getenv("TELEGRAM_TOKEN"),
})

// 添加 Discord
manager.Register("discord", &channels.DiscordAdapter{
    Token: os.Getenv("DISCORD_TOKEN"),
})

// 发送消息
manager.SendMessage(ctx, "telegram", &channels.Message{
    Text: "Hello from AgentFramework!",
})
```

---

## 🚀 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/myvoyage/agentframework.git
cd agentframework

# 安装依赖
go mod download

# 构建
go build ./...
```

### 第一个 Agent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/myvoyage/agentframework/agent"
)

func main() {
    // 创建 ReAct Agent
    a := agent.NewReActAgent(
        agent.WithModel("gpt-4"),
        WithSystemPrompt("你是一个有帮助的AI助手"),
    )

    // 执行
    result, err := a.Execute(context.Background(), "今天天气怎么样？")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result)
}
```

### 运行示例

```bash
# ReAct Agent 示例
go run examples/react_agent_example.go

# 工作流示例
go run examples/workflow_example.go

# IoT 示例
go run examples/iot/zigbee_example.go

# 多渠道集成
go run examples/channels_integration.go
```

---

## 📚 文档

- [📖 完整文档](docs/)
- [🏗️ 架构设计](docs/ARCHITECTURE.md)
- [📝 API 参考](https://pkg.go.dev/github.com/myvoyage/agentframework)
- [💡 使用示例](examples/)
- [🐛 故障排查](docs/TROUBLESHOOTING.md)

### 核心文档

| 文档 | 说明 |
|------|------|
| [综合评估报告](docs/COMPREHENSIVE_EVALUATION_REPORT.md) | 项目全面评估 |
| [优化路线图](docs/OPTIMIZATION_ROADMAP.md) | 性能优化指南 |
| [许可证合规报告](docs/LICENSE_COMPLIANCE_REPORT.md) | AGPL-3.0 合规性 |
| [清理报告](docs/CLEANUP_REPORT.md) | 文件清理记录 |
| [项目总结](docs/PROJECT_SUMMARY.md) | 执行摘要 |

### 功能文档

- [多渠道系统](docs/CHANNELS_OVERVIEW.md)
- [渠道集成指南](docs/CHANNEL_INTEGRATION.md)
- [IoT 协议支持](docs/iot/README.md)
- [快速入门](docs/QUICK_START.md)

---

## 📜 许可证

### GNU Affero General Public License v3.0 or later

本项目采用 **GNU Affero General Public License v3.0 or later** 许可证。

#### 🔑 关键要求

- ✅ **所有修改必须开源** - 任何对源代码的修改都必须在相同许可下发布
- ✅ **网络使用需提供源代码** - 如果您通过网络提供本软件的服务，用户有权获取完整源代码
- ✅ **保留版权声明** - 必须保留原作者的版权声明和许可信息
- ✅ **使用 SPDX 标识符** - 所有源文件都包含 `SPDX-License-Identifier: AGPL-3.0-or-later`

#### 📦 源代码获取

如果您通过网络使用本软件，您有权获取完整源代码：

- **项目仓库**: https://github.com/myvoyage/agentframework
- **许可证文件**: [LICENSE](LICENSE)
- **SPDX 标识符**: AGPL-3.0-or-later
- **许可覆盖率**: 100% (551/551 文件)

#### ⚖️ AGPL-3.0 与 GPL-3.0 的区别

AGPL-3.0 在 GPL-3.0 基础上增加了**网络使用条款**：

> 如果用户通过网络与程序交互（例如 Web 服务），即使没有分发软件副本，
> 也必须向用户提供源代码。

这意味着：
- ❌ 不能将本软件用作专有的 SaaS 服务而不公开源代码
- ✅ 可以用于内部部署
- ✅ 可以用于开源项目
- ✅ 修改必须以相同许可共享

#### 📋 许可证合规性

| 项目 | 状态 | 说明 |
|------|------|------|
| LICENSE 文件 | ✅ 完整 | AGPL-3.0 完整文本 |
| 源代码覆盖 | ✅ 100% | 所有 551 个 Go 文件 |
| SPDX 标识符 | ✅ 完整 | 统一使用 AGPL-3.0-or-later |
| 版权声明 | ✅ 完整 | 所有文件包含版权信息 |
| 格式一致性 | ✅ 良好 | 标准化许可声明格式 |

详细的合规性报告请参阅：[LICENSE_COMPLIANCE_REPORT.md](docs/LICENSE_COMPLIANCE_REPORT.md)

---

## 🤝 贡献

我们欢迎所有形式的贡献！

### 贡献方式

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复

### 贡献流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 添加适当的注释
- 编写单元测试
- 确保所有测试通过 (`go test ./...`)
- 包含 AGPL-3.0 许可声明

### 许可证要求

**重要**: 所有贡献都将自动采用 AGPL-3.0-or-later 许可证。

通过提交贡献，您同意：
- 您的贡献将以 AGPL-3.0-or-later 许可
- 您有权提交这些贡献
- 版权归属于您和/或您的雇主

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行基准测试
go test -bench=. -benchmem ./...

# 运行特定包的测试
go test ./pkg/iot/...
```

---

## 📊 性能

### 工作流性能

| 类型 | 吞吐量 | 延迟 | 并发 |
|------|--------|------|------|
| 顺序 | ~500/s | <10ms | 单线程 |
| 并行 | ~1000/s | <20ms | 多线程 |
| DAG | ~1000/s | <15ms | 多线程 |

### IoT 协议性能

| 协议 | 延迟 | 带宽 | 连接数 |
|------|------|------|--------|
| Zigbee | ~100ms | 250Kbps | 65K+ |
| Thread | ~50ms | 250Kbps | 300+ |
| Z-Wave | ~200ms | 100Kbps | 232 |
| NearLink | ~20μs | 12Mbps | 256 |

---

## 🔧 配置

### 环境变量

```bash
# OpenAI API
export OPENAI_API_KEY="your-key-here"

# Ollama (本地)
export OLLAMA_BASE_URL="http://localhost:11434"

# MQTT (for Zigbee)
export MQTT_BROKER="tcp://localhost:1883"

# Telegram Bot
export TELEGRAM_BOT_TOKEN="your-token-here"

# Discord Bot
export DISCORD_BOT_TOKEN="your-token-here"
```

### 配置文件

```yaml
# config/channels.yaml
channels:
  telegram:
    token: "${TELEGRAM_BOT_TOKEN}"
    webhook_url: "https://your-domain.com/webhook/telegram"

  discord:
    token: "${DISCORD_BOT_TOKEN}"
    guild_id: "your-guild-id"

# config/iot.yaml
iot:
  zigbee:
    mqtt_broker: "tcp://localhost:1883"
    base_topic: "zigbee2mqtt"

  thread:
    border_router: "http://thread-border-router:8080"
```

---

## 📈 路线图

### v1.1 (计划中)

- [ ] Matter 协议支持
- [ ] 可视化工作流编辑器
- [ ] 移动端 SDK
- [ ] 性能优化（目标：提升 50%）

### v1.2 (规划中)

- [ ] 分布式工作流执行
- [ ] 机器学习优化
- [ ] 插件市场
- [ ] 多租户支持

---

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

特别感谢：
- CloudWeGo 团队的 Eino 框架
- Go 社区的优秀工具和库
- 所有测试用户的宝贵反馈

---

## 📞 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/myvoyage/agentframework/issues)
- **功能建议**: [GitHub Discussions](https://github.com/myvoyage/agentframework/discussions)
- **安全问题**: security@example.com

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=myvoyage/agentframework&type=Date)](https://star-history.com/#myvoyage/agentframework&Date)

---

<div align="center">

**如果这个项目对您有帮助，请给我们一个 ⭐️**

Made with ❤️ by AgentFramework Team

</div>
