# 示例项目

本目录包含 Agent Framework 的示例配置和演示项目。

## 📁 目录结构

```
examples/
├── README.md                          # 本文件
├── react_agent_workflow.json          # ReAct Agent 工作流配置 (JSON)
└── react_agent_workflow.yaml          # ReAct Agent 工作流配置 (YAML)
```

---

## 🎯 工作流示例

### ReAct Agent 工作流

[react_agent_workflow.yaml](react_agent_workflow.yaml) - 演示如何使用 ReAct Agent 构建复杂的问题解决工作流。

**工作流结构**：
- **Input Analyzer** - 分析输入问题，识别关键需求
- **ReAct Problem Solver** - 使用 ReAct 推理循环解决问题
- **Solution Validator** - 验证解决方案的正确性和完整性
- **Output Formatter** - 格式化最终输出

**运行方式**：
```bash
# 使用 YAML 配置
go run cmd/workforce_demo/main.go -config examples/react_agent_workflow.yaml

# 使用 JSON 配置
go run cmd/workforce_demo/main.go -config examples/react_agent_workflow.json
```

**配置说明**：
- **类型**: DAG 工作流
- **模型**: GPT-4
- **最大迭代次数**: 20
- **内存管理**: 启用自动裁剪（保留最近 15 条消息）

---

## 🚀 命令行示例

### 基础演示
```bash
# 运行基本演示
go run cmd/demo/main.go

# 运行工作流演示
go run cmd/workforce_demo/main.go
```

### MCP 演示
```bash
# 运行 MCP 工具演示
go run cmd/mcp_demo/main.go

# 运行服务器演示
go run cmd/server_demo/main.go
```

### CLI 工具
```bash
# Agent CLI
go run cmd/agent-cli/main.go --help

# Pipeline CLI
go run cmd/pipeline_cli/main.go --help
```

### 支持演示
```bash
# 运行支持功能演示
go run cmd/support_demo/main.go

# 运行导师演示
go run cmd/tutor_demo/main.go
```

---

## 📝 配置格式

### YAML 配置
```yaml
type: dag
name: problem_solving_workflow
metadata:
  description: A DAG workflow using ReAct agent
  version: 1.0.0

nodes:
  agent_name:
    type: agent
    name: Agent Display Name
    config:
      kind: react
      model: gpt-4
      instructions: |
        Your instructions here
      max_iterations: 20

edges:
  - from: node1
    to: node2
```

### JSON 配置
```json
{
  "type": "dag",
  "name": "problem_solving_workflow",
  "metadata": {
    "description": "A DAG workflow using ReAct agent",
    "version": "1.0.0"
  },
  "nodes": {
    "agent_name": {
      "type": "agent",
      "name": "Agent Display Name",
      "config": {
        "kind": "react",
        "model": "gpt-4",
        "instructions": "Your instructions here",
        "max_iterations": 20
      }
    }
  },
  "edges": [
    {"from": "node1", "to": "node2"}
  ]
}
```

---

## 🔧 自定义配置

### 创建自己的工作流

1. 复制示例配置：
```bash
cp examples/react_agent_workflow.yaml examples/my_workflow.yaml
```

2. 编辑配置文件：
```yaml
type: dag
name: my_custom_workflow
nodes:
  my_agent:
    type: agent
    name: My Custom Agent
    config:
      kind: chat
      model: gpt-4
      instructions: |
        You are a helpful assistant.
```

3. 运行自定义工作流：
```bash
go run cmd/workforce_demo/main.go -config examples/my_workflow.yaml
```

---

## 📚 相关文档

- [工作流引擎](../doc/getting-started/QUICKSTART.md) - 快速入门指南
- [DAG 工作流](../doc/ARCHITECTURE_UNIFIED.md) - 架构文档
- [Agent 配置](../doc/configuration/CONFIGURATION.md) - 配置说明

---

## 💡 提示

- YAML 格式更易读，推荐用于手动编辑
- JSON 格式更适合程序生成和解析
- 使用 `-config` 参数指定配置文件路径
- 查看各命令的 `--help` 了解更多选项

---

**最后更新**: 2025-02-03
