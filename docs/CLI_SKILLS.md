# 增强技能系统 CLI 使用指南

## 概述

增强技能系统是 AgentFramework 的核心功能之一，支持通过 Markdown + YAML 格式定义可复用的技能，无需编写代码。

### 特性

- 📝 **零编程定义**：使用 Markdown + YAML frontmatter 格式
- 🔍 **智能依赖检查**：自动检测二进制和环境变量依赖
- 🔧 **多执行器支持**：Shell、HTTP、Workflow
- 🎯 **触发器系统**：命令、关键词、模式匹配
- 🔄 **变量替换**：支持 `{{.Variable}}` 模板语法
- 👀 **文件监控**：自动检测技能变化并重载
- 🛡️ **安全检查**：防止危险命令执行

## CLI 命令

### 基本语法

```bash
agentframework enhanced-skill [command] [flags]
```

### 全局选项

- `-d, --dir string`：指定技能目录（默认：`.skills`）
- `-c, --config string`：指定配置文件（默认：`host.yaml`）
- `-m, --model string`：指定模型名称
- `-o, --output string`：输出格式（table/json）
- `-v, --verbose`：详细输出

### 可用命令

#### 1. list - 列出所有技能

```bash
agentframework enhanced-skill list
```

**示例输出：**

```
ID                                  名称         版本     分类           描述
──                                  ──         ──     ──           ──
com.agentframework.skill.github_pr  github_pr  1.0.0  development  管理 GitHub Pull Requests

总计: 1 个技能
```

#### 2. get - 获取技能详情

```bash
agentframework enhanced-skill get [skill-id]
```

**示例：**

```bash
agentframework enhanced-skill get com.agentframework.skill.github_pr
```

**输出包含：**
- 基本信息（ID、名称、版本、作者等）
- 触发器列表
- 动作列表
- 依赖要求
- 配置参数

#### 3. search - 根据触发器搜索技能

```bash
agentframework enhanced-skill search [trigger-type] [pattern]
```

**支持的触发器类型：**
- `command`：命令触发器（如 `/pr`）
- `keyword`：关键词触发器
- `pattern`：正则表达式模式
- `schedule`：定时任务
- `event`：事件触发

**示例：**

```bash
# 搜索命令触发器
agentframework enhanced-skill search command "/pr"

# 搜索关键词
agentframework enhanced-skill search keyword "github"
```

#### 4. check - 检查技能依赖

```bash
agentframework enhanced-skill check [skill-id]
```

**示例：**

```bash
agentframework enhanced-skill check com.agentframework.skill.github_pr
```

**输出示例：**

```
检查技能: github_pr (com.agentframework.skill.github_pr)

✓ 所有依赖已满足

警告:
  • Optional env GITHUB_TOKEN not set
```

**依赖不满足时会显示安装提示：**

```
✗ 依赖不满足:

  • 缺失: binary: gh

安装建议:
  • Install gh via brew:
    brew install gh
  • Install gh via apt:
    sudo apt install gh
```

#### 5. execute - 执行技能动作

```bash
agentframework enhanced-skill execute [skill-id] [action-id] --vars "key=value,key2=value2"
```

**示例：**

```bash
# 执行 list_prs 动作
agentframework enhanced-skill execute com.agentframework.skill.github_pr list_prs --vars "Repo=cli/cli,State=open,Limit=5"
```

#### 6. install - 从 GitHub 安装技能

```bash
agentframework enhanced-skill install [github-repo]
```

**示例：**

```bash
agentframework enhanced-skill install agentframework/skill-hello-world
```

## 技能定义格式

### 基本结构

技能定义文件位于 `.skills/[skill-name]/SKILL.md`，格式如下：

```markdown
---
id: "com.example.skill.my-skill"
name: "my_skill"
version: "1.0.0"
category: "utility"
tags:
  - example
  - utility
description: "技能描述"
author: "Your Name"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "command-name"
      version: ">=1.0.0"
      install:
        brew: "brew install command"
        apt: "sudo apt install command"
  env:
    - name: "ENV_VAR"
      optional: false
      description: "环境变量描述"

triggers:
  - type: "command"
    pattern: "/myskill"
    priority: 10
  - type: "keyword"
    pattern: "my keyword"
    priority: 5

actions:
  - id: "my_action"
    type: "shell"
    description: "动作描述"
    config:
      command: "echo {{.Input}}"
    timeout: "30s"

config:
  max_output_size: 10485760
  max_execution_time: "60s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# 技能名称

技能的详细说明文档...

## 使用示例

...
```

### 字段说明

#### 基本信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | ✓ | 唯一标识符，建议使用反向域名格式 |
| name | string | ✓ | 技能名称（小写，下划线分隔） |
| version | string | ✓ | 版本号（语义化版本） |
| category | string | ✓ | 分类（development, utility, system 等） |
| tags | []string | ✗ | 标签列表 |
| description | string | ✓ | 技能描述 |
| author | string | ✗ | 作者 |
| license | string | ✗ | 许可证 |
| enabled | bool | ✗ | 是否启用（默认 true） |

#### 依赖 (prerequisites)

| 字段 | 类型 | 说明 |
|------|------|------|
| bins | []BinaryDependency | 二进制依赖列表 |
| env | []EnvDependency | 环境变量依赖列表 |

**BinaryDependency:**
- `name`: 二进制名称
- `version`: 版本要求（如 `>=2.0.0`）
- `install`: 包管理器到安装命令的映射

**EnvDependency:**
- `name`: 环境变量名
- `optional`: 是否可选
- `description`: 描述

#### 触发器 (triggers)

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 触发器类型 |
| pattern | string | 匹配模式 |
| priority | int | 优先级（数字越大优先级越高） |

#### 动作 (actions)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 动作唯一标识符 |
| type | string | 执行器类型（shell, http, workflow） |
| description | string | 动作描述 |
| config | map | 执行器配置 |
| timeout | duration | 超时时间（如 "30s", "5m"） |

**Shell 执行器配置：**
- `command`: 要执行的 shell 命令

**HTTP 执行器配置：**
- `url`: 请求 URL
- `method`: HTTP 方法（GET, POST 等）
- `headers`: 请求头映射

#### 配置 (config)

| 字段 | 类型 | 说明 |
|------|------|------|
| max_output_size | int64 | 最大输出大小（字节） |
| max_execution_time | duration | 最大执行时间 |
| enable_cache | bool | 是否启用缓存 |
| cache_ttl | duration | 缓存过期时间 |

### 变量替换

在命令和 URL 中可以使用 `{{.VariableName}}` 语法引用变量：

```yaml
actions:
  - id: "list_prs"
    type: "shell"
    config:
      command: "gh pr list --repo {{.Repo}} --state {{.State}} --limit {{.Limit}}"
```

执行时传入变量：
```bash
agentframework enhanced-skill execute com.agentframework.skill.github_pr list_prs --vars "Repo=cli/cli,State=open,Limit=10"
```

## 最佳实践

### 1. 命名规范

- **ID**: 使用反向域名格式 `com.organization.skill.name`
- **Name**: 小写字母，下划线分隔
- **Action IDs**: 动词_名词格式，如 `list_items`, `create_resource`

### 2. 依赖管理

- 将必需的依赖设为 `optional: false`
- 为每个依赖提供多平台安装方法
- 在技能文档中说明依赖的用途

### 3. 错误处理

- 设置合理的超时时间
- 在命令中使用错误检查
- 提供清晰的错误消息

### 4. 安全考虑

- 避免使用危险命令（如 `rm -rf`）
- 对用户输入进行验证
- 使用最小权限原则

### 5. 文档

- 在 SKILL.md 的 Markdown 部分提供详细文档
- 包含使用示例
- 说明前置条件和限制
- 提供故障排除指南

## 示例技能

### GitHub PR 技能

完整的 GitHub PR 管理技能示例位于：`.skills/github_pr/SKILL.md`

功能：
- 列出 Pull Requests
- 查看 PR 详情
- 创建新的 Pull Request

依赖：
- GitHub CLI (`gh`) >= 2.0.0
- GitHub Token（可选）

## 故障排除

### 技能未加载

1. 检查 SKILL.md 文件格式是否正确
2. 验证 YAML frontmatter 语法
3. 查看依赖是否满足
4. 使用 `--verbose` 标志查看详细日志

### 依赖检查失败

1. 使用 `check` 命令查看详细依赖信息
2. 按照安装提示安装缺失的依赖
3. 设置必需的环境变量

### 执行失败

1. 确认技能 ID 和动作 ID 正确
2. 检查变量格式和值
3. 查看错误消息了解具体问题
4. 验证依赖是否正确安装

## 高级用法

### 工作流执行器

组合多个动作：

```yaml
actions:
  - id: "deploy"
    type: "workflow"
    config:
      steps:
        - id: "build"
          type: "shell"
          config:
            command: "make build"
        - id: "test"
          type: "shell"
          config:
            command: "make test"
        - id: "deploy"
          type: "shell"
          config:
            command: "make deploy"
```

### HTTP 请求

```yaml
actions:
  - id: "api_call"
    type: "http"
    config:
      url: "https://api.example.com/{{.Endpoint}}"
      method: "POST"
      headers:
        Authorization: "Bearer {{.Token}}"
        Content-Type: "application/json"
```

## 相关文档

- [增强技能系统架构](../agent/skills/README.md)
- [技能定义参考](../agent/skills/ENHANCED_DEFINITION.md)
- [执行器实现指南](../agent/skills/EXECUTORS.md)
