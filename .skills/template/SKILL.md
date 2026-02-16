---
id: "com.example.skill.my-skill"
name: "my_skill"
version: "1.0.0"
category: "utility"
tags:
  - example
  - template
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
        apk: "apk add command"
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

  - id: "http_action"
    type: "http"
    description: "HTTP 请求示例"
    config:
      url: "https://api.example.com/{{.Endpoint}}"
      method: "GET"
      headers:
        Authorization: "Bearer {{.Token}}"
    timeout: "30s"

config:
  max_output_size: 10485760
  max_execution_time: "60s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# 技能名称

详细的技能说明文档。

## 功能

- 功能 1
- 功能 2
- 功能 3

## 使用示例

### 基本用法

```bash
agentframework enhanced-skill execute com.example.skill.my-skill my_action --vars "Input=Hello"
```

### 高级用法

```bash
# 带参数的示例
agentframework enhanced-skill execute com.example.skill.my-skill my_action --vars "Input=World,Param=Value"
```

## 参数说明

| 参数 | 说明 | 必填 | 默认值 | 示例 |
|------|------|------|--------|------|
| Input | 输入参数 | 是 | - | `Hello` |
| Param | 其他参数 | 否 | `default` | `Value` |

## 前置条件

### 必需的依赖

1. **command-name**: 版本 >= 1.0.0
   - macOS: `brew install command`
   - Linux: `sudo apt install command`
   - Alpine: `apk add command`

### 环境变量

- `ENV_VAR`: 环境变量描述
  ```bash
  export ENV_VAR="your_value"
  ```

## 故障排除

### 问题 1: 命令未找到

**错误信息**: `command: not found`

**解决方案**:
1. 安装依赖：`brew install command`
2. 验证安装：`command --version`
3. 检查 PATH：`echo $PATH`

### 问题 2: 权限不足

**错误信息**: `Permission denied`

**解决方案**:
1. 检查文件权限
2. 使用 `sudo` 运行（如果安全）
3. 修改文件权限：`chmod +x file`

## 技术细节

### 实现原理

技能的工作原理说明...

### 性能考虑

- 注意事项 1
- 注意事项 2

### 安全考虑

- 安全事项 1
- 安全事项 2

## 相关资源

- [相关文档链接](https://example.com)
- [API 参考](https://api.example.com)
- [示例代码](https://github.com/example)

## 许可证

MIT License

## 作者

Your Name - [@yourhandle](https://twitter.com/yourhandle)

## 贡献

欢迎提交 Issue 和 Pull Request！
