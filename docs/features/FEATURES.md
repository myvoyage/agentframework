# AgentFramework 功能特性总览

> **全面、细致的功能特性说明**
> **版本**: v2.0.0
> **最后更新**: 2026-02-15

---

## 📋 目录

- [核心特性](#核心特性)
- [Agent 系统](#agent-系统)
- [工作流引擎](#工作流引擎)
- [技能系统](#技能系统)
- [MCP 协议](#mcp-协议)
- [存储系统](#存储系统)
- [监控遥测](#监控遥测)
- [安全沙箱](#安全沙箱)
- [模型管理](#模型管理)
- [协作系统](#协作系统)
- [前端界面](#前端界面)
- [CLI 工具](#cli-工具)

---

## 核心特性

### 🎯 设计理念

AgentFramework 采用**模块化、可扩展**的设计理念，构建企业级 AI 代理框架：

| 特性 | 说明 |
|------|------|
| **高性能** | Go 语言原生并发，支持 1000+ 并发任务 |
| **高可用** | 检查点恢复、错误重试、健康检查 |
| **高扩展** | 插件化架构、技能系统、MCP 协议 |
| **高安全** | 多层沙箱、权限控制、审计日志 |
| **易使用** | YAML 配置、CLI 工具、桌面应用 |

### 📊 项目规模

| 指标 | 数值 |
|------|------|
| Go 源文件 | **430 个** |
| 代码行数 | **65,000+ 行** |
| Agent 类型 | **12+ 种** |
| 工作流类型 | **6 种** |
| Worker 角色 | **7 种** |
| MCP 工具 | **44 个** |
| 测试文件 | **111 个** |

---

## Agent 系统

### Agent 类型概览

AgentFramework 提供 **12+ 种 Agent 类型**，覆盖从简单对话到复杂任务的各类场景：

```
┌────────────────────────────────────────────────────────────┐
│                    Agent 类型体系                          │
├────────────────────────────────────────────────────────────┤
│  基础层                                                     │
│  ├── ChatAgent      基础对话代理                           │
│  └── ReActAgent     推理-行动代理                          │
├────────────────────────────────────────────────────────────┤
│  专业层 (WorkerAgent)                                       │
│  ├── DeveloperAgent  开发者代理                            │
│  ├── BrowserAgent    浏览器代理                            │
│  ├── DocumentAgent   文档代理                              │
│  ├── MultiModalAgent 多模态代理                            │
│  ├── ResearcherAgent 研究者代理                            │
│  ├── WriterAgent     写作代理                              │
│  └── ReviewerAgent   审核代理                              │
├────────────────────────────────────────────────────────────┤
│  特殊层                                                     │
│  ├── SWEAgent        软件工程代理                          │
│  ├── DataAnalysisAgent 数据分析代理                        │
│  ├── EdgeAgent       边缘计算代理                          │
│  ├── RealTimeAgent   实时数据代理                          │
│  ├── SecurityAgent   安全代理                              │
│  ├── HumanNode       人工介入节点                          │
│  ├── SkillAgent      技能包装代理                          │
│  └── WorkflowAgent   工作流包装代理                        │
└────────────────────────────────────────────────────────────┘
```

### 1. ChatAgent（基础对话代理）

**核心能力**：
- 流式响应支持，实现实时对话体验
- 智能记忆管理，自动压缩长对话
- 工具调用能力，扩展 Agent 能力边界
- 状态机管理，支持生命周期钩子

**配置示例**：
```yaml
agents:
  - name: "chat"
    type: "chat"
    model: "ollama-llama3"
    instructions: "你是一个有用的AI助手"
    tools:
      - "http_request"
      - "file_operation"
```

### 2. ReActAgent（推理-行动代理）

**核心能力**：
- 基于 Eino 框架实现
- 支持 ReAct 推理循环
- 最大迭代次数控制
- 自动工具选择和调用

### 3. WorkerAgent（专业工作代理）

支持 **7 种专业角色**：

| 角色 | 能力标签 |
|------|----------|
| **Developer** | `write_code`, `execute_code`, `debug_code` |
| **Browser** | `search_web`, `browse_url`, `extract_content` |
| **Document** | `create_document`, `edit_document`, `read_document` |
| **MultiModal** | `process_image`, `generate_image`, `process_audio` |
| **Researcher** | `search_information`, `analyze_data`, `cite_sources` |
| **Writer** | `write_content`, `edit_content`, `summarize_content` |
| **Reviewer** | `review_content`, `provide_feedback`, `check_quality` |

---

## 工作流引擎

### 工作流类型

支持 **6 种工作流类型**：

| 类型 | 复杂度 | 适用场景 |
|------|--------|----------|
| **Sequential** | ⭐ | 简单任务链 |
| **Parallel** | ⭐⭐ | 独立任务并行 |
| **DAG** | ⭐⭐⭐⭐ | 复杂依赖关系 |
| **Routing** | ⭐⭐⭐ | 条件分支 |
| **Planning** | ⭐⭐⭐⭐ | 自动规划 |
| **Graph** | ⭐⭐⭐⭐⭐ | 任意拓扑结构 |

### DAG 工作流特性

- 拓扑排序执行
- 并发控制（默认 CPU 核心数，最大 8）
- 优先级调度
- 重试机制
- 检查点持久化

---

## 技能系统

### 技能类型

| 类型 | 说明 |
|------|------|
| **HTTP 技能** | 执行 HTTP 请求 |
| **文件技能** | 安全的文件系统操作 |
| **代码技能** | 多语言代码执行 |
| **数据技能** | 数据转换和处理 |
| **Markdown 技能** | 零代码定义技能 |

### SkillAgent 增强

- **SkillAgentWithValidation**：输入验证
- **SkillAgentWithRetry**：重试机制
- **SkillAgentWithCache**：结果缓存

---

## MCP 协议

### 支持的操作

| 操作 | 方法 |
|------|------|
| 初始化 | `Initialize(ctx)` |
| 获取工具列表 | `ListTools()` |
| 调用工具 | `CallTool(ctx, name, arguments)` |
| 断开连接 | `Disconnect(ctx)` |

### 工具加载器

- **HTTPToolLoader**：HTTP URL 加载
- **FileToolLoader**：JSON/YAML 文件加载
- **PluginToolLoader**：Go 插件
- **MCPToolLoader**：MCP 服务器

---

## 存储系统

### ThreadStore（会话存储）

| 实现 | 特性 |
|------|------|
| **MemoryThreadStore** | 内存存储，支持 TTL |
| **FileThreadStore** | 文件系统存储 |
| **RedisThreadStore** | Redis 存储 |
| **SQLThreadStore** | SQL 数据库存储 |

### CheckpointStore（检查点存储）

- MemoryCheckpointStore
- FileCheckpointStore
- RedisCheckpointStore
- SQLCheckpointStore

---

## 监控遥测

### 指标类型

- **Counter**：累计计数器
- **Gauge**：可增减仪表
- **Timer**：计时器
- **Histogram**：直方图

### OpenTelemetry 集成

自动追踪 Agent 执行，支持分布式追踪。

### 内存监控

- 堆内存分配
- GC 周期和暂停时间
- 优化建议系统

---

## 安全沙箱

### 资源配额

```go
type ResourceQuota struct {
    MaxFileSize     int64  // 单文件最大大小
    MaxTotalSize    int64  // 总大小限制
    MaxFileCount    int    // 文件数量限制
    MaxCPUSeconds   int    // CPU 时间限制
    MaxMemoryBytes  int64  // 内存限制
}
```

### SecurityAgent

- 命令验证
- 权限检查
- 数据加密
- 审计日志
- ACL 管理

---

## 模型管理

### 支持的模型

| 类型 | 说明 |
|------|------|
| **OpenAI** | GPT-4, GPT-3.5 等 |
| **Ollama** | 本地模型服务 |
| **LM Studio** | 本地模型工作室 |
| **vLLM** | vLLM 推理框架 |

---

## 协作系统

### 协作模式

| 模式 | 说明 |
|------|------|
| **Single** | 单代理执行 |
| **Parallel** | 并行执行 |
| **Sequential** | 顺序执行 |
| **Consensus** | 共识决策 |

---

## 前端界面

### 技术栈

- **Vue.js 3** + **TypeScript**
- **Element Plus** UI 组件库
- **AntV G6** 工作流图编辑
- **Pinia** 状态管理

### 核心功能

- 工作流可视化编辑器
- 技能管理中心
- 配置管理界面
- 文件浏览器

---

## CLI 工具

### 主要命令

| 命令 | 说明 |
|------|------|
| `workflow list` | 列出所有工作流 |
| `workflow execute` | 执行工作流 |
| `skill list` | 列出所有技能 |
| `agent chat` | 运行对话代理 |

---

## 相关文档

- 📘 [架构概览](../architecture/ARCHITECTURE_OVERVIEW.md)
- 📘 [API 文档](../api/API.md)
- 📘 [配置指南](../configuration/CONFIGURATION.md)

---

**Made with ❤️ by AgentFramework Team**
