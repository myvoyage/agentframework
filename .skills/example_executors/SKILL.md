---
id: "com.agentframework.skill.example_executors"
name: "example_executors"
version: "1.0.0"
category: "example"
tags:
  - examples
  - executors
  - tutorial
description: "各种执行器的使用示例"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins: []
  env: []

triggers:
  - type: "command"
    pattern: "/example"
    priority: 10

actions:
  - id: "template_render"
    type: "template"
    description: "模板渲染示例"
    config:
      template: "Hello {{.Name}},\n\nWelcome to {{.App}}!\n\nYour role: {{.Role}}\n\nStatus: {{.Status}}"
    timeout: "5s"

  - id: "file_write"
    type: "file"
    description: "写入文件示例"
    config:
      operation: "write"
      path: "./output/{{.Filename}}.txt"
      content: "{{.Content}}"
    timeout: "10s"

  - id: "file_read"
    type: "file"
    description: "读取文件示例"
    config:
      operation: "read"
      path: "./output/{{.Filename}}.txt"
    timeout: "5s"

  - id: "json_parse"
    type: "json"
    description: "JSON 解析示例"
    config:
      operation: "format"
      input: '{"name":"{{.Name}}","age":{{.Age}},"active":{{.Active}}}'
    timeout: "5s"

  - id: "json_extract"
    type: "json"
    description: "JSON 字段提取示例"
    config:
      operation: "extract"
      input: '{"user":{"name":"Alice","age":30},"status":"active"}'
      field: "user.name"
    timeout: "5s"

  - id: "workflow_example"
    type: "workflow"
    description: "工作流示例"
    config:
      steps:
        - id: "step1"
          type: "template"
          config:
            template: "Step 1: Processing {{.Item}}..."
        - id: "step2"
          type: "file"
          config:
            operation: "write"
            path: "./output/workflow.txt"
            content: "{{.Item}} processed at {{.Timestamp}}"
        - id: "step3"
          type: "template"
          config:
            template: "Workflow completed for {{.Item}}"
    timeout: "30s"

config:
  max_output_size: 1048576
  max_execution_time: "60s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# 执行器示例技能

演示各种执行器类型的实用技能。

## 功能

- **Template**: 模板渲染和变量替换
- **File**: 文件读写操作
- **JSON**: JSON 解析和处理
- **Workflow**: 多步骤工作流

## 使用示例

### 1. 模板渲染

```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors template_render --vars "Name=Alice,App=AgentFramework,Role=Admin,Status=Active"
```

输出：
```
Hello Alice,

Welcome to AgentFramework!

Your role: Admin

Status: Active
```

### 2. 文件操作

#### 写入文件
```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors file_write --vars "Filename=test,Content=Hello World"
```

#### 读取文件
```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors file_read --vars "Filename=test"
```

### 3. JSON 处理

#### 格式化 JSON
```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors json_parse --vars "Name=Bob,Age=25,Active=true"
```

输出：
```json
{
  "name": "Bob",
  "age": 25,
  "active": true
}
```

#### 提取字段
```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors json_extract --vars ""
```

输出：
```
Alice
```

### 4. 工作流

```bash
agentframework enhanced-skill execute com.agentframework.skill.example_executors workflow_example --vars "Item=data-processing,Timestamp=$(date +%s)"
```

## 执行器说明

### Template 执行器

**用途**: 渲染模板并替换变量

**配置**:
- `template`: 模板内容（支持 `{{.Var}}` 语法）

**示例**:
```yaml
config:
  template: "Hello {{.Name}}, your balance is ${{.Balance}}"
```

### File 执行器

**用途**: 文件操作（读、写、追加、删除、列表、存在性检查）

**操作类型**:
- `read`: 读取文件
- `write`: 写入文件
- `append`: 追加到文件
- `delete`: 删除文件
- `list`: 列出目录
- `exists`: 检查文件是否存在

**配置**:
- `operation`: 操作类型
- `path`: 文件路径
- `content`: 文件内容（write/append 操作）

**示例**:
```yaml
config:
  operation: "write"
  path: "./output.txt"
  content: "Line 1\nLine 2"
```

### JSON 执行器

**用途**: JSON 解析、格式化、字段提取

**操作类型**:
- `parse`: 解析并格式化 JSON
- `extract`: 提取指定字段
- `format`: 格式化 JSON
- `merge`: 合并 JSON 对象
- `query`: 路径查询

**配置**:
- `operation`: 操作类型
- `input`: JSON 输入
- `field`: 要提取的字段
- `path`: 查询路径

**示例**:
```yaml
config:
  operation: "extract"
  input: '{"user":{"name":"Alice"}}'
  field: "user.name"
```

### Workflow 执行器

**用途**: 组合多个执行器步骤

**配置**:
- `steps`: 步骤数组

**示例**:
```yaml
config:
  steps:
    - id: "prepare"
      type: "template"
      config:
        template: "Preparing..."
    - id: "process"
      type: "shell"
      config:
        command: "echo 'Processing'"
```

## 高级用法

### 1. 组合执行器

```yaml
actions:
  - id: "generate_config"
    type: "workflow"
    config:
      steps:
        - id: "template"
          type: "template"
          config:
            template: '{"db_host":"{{.Host}}","db_port":{{.Port}}}'
        - id: "write"
          type: "file"
          config:
            operation: "write"
            path: "./config.json"
            content: "{{.template}}"
        - id: "verify"
          type: "json"
          config:
            operation: "parse"
            input: "{{.write}}"
```

### 2. 条件处理

```yaml
actions:
  - id: "conditional_template"
    type: "template"
    config:
      template: |
        {{if eq .Status "active"}}Status: Active{{end}}
        {{if eq .Status "inactive"}}Status: Inactive{{end}}
```

### 3. 数据转换

```yaml
actions:
  - id: "transform_data"
    type: "json"
    config:
      operation: "merge"
      base: '{"name":"default"}'
      merge: '{"name":"{{.Name}}","age":{{.Age}}}'
```

## 最佳实践

### 1. 错误处理

```yaml
actions:
  - id: "safe_file_read"
    type: "shell"
    config:
      command: "cat {{.File}} 2>/dev/null || echo 'File not found'"
```

### 2. 路径安全

```yaml
actions:
  - id: "safe_write"
    type: "file"
    config:
      operation: "write"
      path: "./output/{{.Filename}}"
      # 使用相对路径，避免覆盖重要文件
```

### 3. 数据验证

```yaml
actions:
  - id: "validate_json"
    type: "json"
    config:
      operation: "parse"
      input: "{{.Data}}"
    # 如果 JSON 无效，将返回错误
```

## 相关文档

- [增强技能定义参考](../../agent/skills/README.md)
- [技能开发指南](../../docs/SKILL_DEVELOPMENT.md)
- [CLI 技能使用指南](../../docs/CLI_SKILLS.md)
