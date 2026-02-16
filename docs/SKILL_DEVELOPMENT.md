# 技能开发指南

本指南将帮助您创建自己的增强技能，无需编写任何代码。

## 快速开始

### 1. 创建技能目录

```bash
mkdir -p .skills/my_skill
cd .skills/my_skill
```

### 2. 创建 SKILL.md 文件

复制模板并修改：

```bash
cp ../template/SKILL.md ./SKILL.md
```

### 3. 编辑技能定义

使用您喜欢的编辑器打开 `SKILL.md` 并编辑 YAML frontmatter。

### 4. 测试技能

```bash
# 列出技能
agentframework enhanced-skill list

# 获取技能详情
agentframework enhanced-skill get com.example.skill.my-skill

# 检查依赖
agentframework enhanced-skill check com.example.skill.my-skill

# 执行动作
agentframework enhanced-skill execute com.example.skill.my-skill my_action --vars "Input=Hello"
```

## 技能定义详解

### 基本信息

```yaml
---
id: "com.example.skill.my-skill"          # 唯一标识符（必填）
name: "my_skill"                          # 技能名称（必填）
version: "1.0.0"                          # 版本号（必填）
category: "utility"                       # 分类（必填）
tags:                                     # 标签（可选）
  - example
  - template
description: "技能描述"                   # 描述（必填）
author: "Your Name"                       # 作者（可选）
license: "MIT"                            # 许可证（可选）
enabled: true                             # 是否启用（可选）
```

### 命名规范

| 项目 | 规范 | 示例 |
|------|------|------|
| ID | 反向域名格式 | `com.example.skill.my-skill` |
| Name | 小写，下划线 | `my_skill`, `git_helper` |
| Category | 小写，标准分类 | `utility`, `development`, `system` |
| Action ID | 动词_名词 | `list_files`, `create_branch` |

### 分类列表

- `utility`: 通用工具
- `development`: 开发工具
- `system`: 系统工具
- `productivity`: 生产力工具
- `automation`: 自动化工具
- `monitoring`: 监控工具

## 依赖定义

### 二进制依赖

```yaml
prerequisites:
  bins:
    - name: "python3"
      version: ">=3.8"
      install:
        brew: "brew install python@3"
        apt: "sudo apt install python3"
        apk: "apk add python3"
```

### 环境变量依赖

```yaml
prerequisites:
  env:
    - name: "API_KEY"
      optional: false
      description: "API 密钥"
```

## 触发器定义

### 触发器类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `command` | 命令触发 | `/git` |
| `keyword` | 关键词触发 | `git helper` |
| `pattern` | 正则表达式 | `^/\w+$` |
| `schedule` | 定时任务 | `0 0 * * *` |
| `event` | 事件触发 | `file.changed` |

### 触发器示例

```yaml
triggers:
  - type: "command"
    pattern: "/myskill"
    priority: 10

  - type: "keyword"
    pattern: "my skill"
    priority: 5
```

## 动作定义

### Shell 执行器

```yaml
actions:
  - id: "list_files"
    type: "shell"
    description: "列出文件"
    config:
      command: "ls -la {{.Path}}"
    timeout: "30s"
```

### HTTP 执行器

```yaml
actions:
  - id: "fetch_data"
    type: "http"
    description: "获取数据"
    config:
      url: "https://api.example.com/{{.Endpoint}}"
      method: "GET"
      headers:
        Authorization: "Bearer {{.Token}}"
    timeout: "30s"
```

### Workflow 执行器

```yaml
actions:
  - id: "deploy"
    type: "workflow"
    description: "部署流程"
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

## 变量替换

### 基本语法

使用 `{{.VariableName}}` 语法引用变量：

```yaml
config:
  command: "echo {{.Message}}"
```

### 传递变量

```bash
agentframework enhanced-skill execute skill.id action.id --vars "Message=Hello,Count=5"
```

### 变量格式

| 格式 | 说明 | 示例 |
|------|------|------|
| `{{.Name}}` | 简单变量 | `{{.Path}}` |
| `{{.Name}}` | 带空格需要引号 | `"{{.Message}}"` |

## 配置选项

```yaml
config:
  max_output_size: 10485760      # 最大输出（字节）
  max_execution_time: "60s"      # 最大执行时间
  enable_cache: true             # 启用缓存
  cache_ttl: "5m"                # 缓存过期时间
```

## 最佳实践

### 1. 错误处理

```yaml
actions:
  - id: "safe_action"
    type: "shell"
    config:
      command: "command || echo 'Command failed'"
    timeout: "30s"
```

### 2. 参数验证

```yaml
actions:
  - id: "validated_action"
    type: "shell"
    config:
      command: |
        if [ -z "{{.Input}}" ]; then
          echo "Error: Input is required"
          exit 1
        fi
        process "{{.Input}}"
```

### 3. 安全考虑

```yaml
actions:
  - id: "safe_download"
    type: "shell"
    config:
      # 验证 URL
      command: |
        URL="{{.URL}}"
        if [[ ! "$URL" =~ ^https:// ]]; then
          echo "Error: Only HTTPS URLs allowed"
          exit 1
        fi
        curl "$URL" -o output.txt
```

### 4. 输出格式化

```yaml
actions:
  - id: "formatted_output"
    type: "shell"
    config:
      command: |
        echo "=== Results ==="
        command | jq '.'
        echo "==============="
```

## 调试技巧

### 1. 查看详细输出

```bash
# 使用 verbose 模式
agentframework -v enhanced-skill execute skill.id action.id --vars "Param=Value"
```

### 2. 测试命令

```yaml
# 添加调试动作
- id: "debug"
  type: "shell"
  description: "调试信息"
  config:
    command: |
      echo "Debug Info:"
      echo "Path: {{.Path}}"
      echo "Count: {{.Count}}"
```

### 3. 分步测试

```bash
# 1. 检查依赖
agentframework enhanced-skill check skill.id

# 2. 查看详情
agentframework enhanced-skill get skill.id

# 3. 执行动作
agentframework enhanced-skill execute skill.id action.id --vars "Param=Value"
```

## 常见问题

### Q: 技能未加载？

**A**: 检查以下内容：
1. SKILL.md 文件格式是否正确
2. YAML frontmatter 语法是否正确
3. 依赖是否满足
4. 使用 `-v` 标志查看详细日志

### Q: 变量替换不工作？

**A**: 确保：
1. 变量名格式正确：`{{.VariableName}}`
2. 执行时传递了变量：`--vars "Var=Value"`
3. 变量名区分大小写

### Q: 命令执行超时？

**A**: 增加 timeout 值：
```yaml
timeout: "120s"  # 增加到 2 分钟
```

### Q: 如何处理多行命令？

**A**: 使用 YAML 的多行字符串：
```yaml
config:
  command: |
    line 1
    line 2
    line 3
```

## 示例技能

查看 `.skills/` 目录中的示例：

- `github_pr`: GitHub PR 管理
- `network_diagnostics`: 网络诊断工具
- `file_manager`: 文件管理工具
- `git_helper`: Git 助手
- `template`: 开发模板

## 高级主题

### 条件执行

```yaml
actions:
  - id: "conditional"
    type: "shell"
    config:
      command: |
        if [ "{{.Condition}}" = "true" ]; then
          action_one
        else
          action_two
        fi
```

### 循环处理

```yaml
actions:
  - id: "loop"
    type: "shell"
    config:
      command: |
        for i in $(seq 1 {{.Count}}); do
          process_item "$i"
        done
```

### 并行执行

```yaml
actions:
  - id: "parallel"
    type: "shell"
    config:
      command: |
        command1 &
        command2 &
        command3 &
        wait
```

## 发布技能

### 1. 准备发布

```bash
# 验证技能
agentframework enhanced-skill check your.skill.id

# 测试所有动作
agentframework enhanced-skill execute your.skill.id action1
agentframework enhanced-skill execute your.skill.id action2
```

### 2. 创建 GitHub 仓库

```bash
# 创建仓库
mkdir your-skill-repo
cp -r .skills/your_skill/* your-skill-repo/
cd your-skill-repo

git init
git add SKILL.md
git commit -m "Initial commit"
gh repo create your-skill --public
git push -u origin main
```

### 3. 共享技能

用户可以安装您的技能：

```bash
agentframework enhanced-skill install yourusername/your-skill-repo
```

## 相关资源

- [增强技能系统概述](CLI_SKILLS.md)
- [CLI 使用指南](CLI_USAGE.md)
- [重构总结](REFACTORING_SUMMARY.md)

## 获取帮助

- GitHub Issues: https://github.com/your-repo/issues
- 讨论: https://github.com/your-repo/discussions
- 文档: https://your-docs-site.com

---

**祝你创建出强大的技能！** 🚀
