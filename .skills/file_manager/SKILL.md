---
id: "com.agentframework.skill.file_manager"
name: "file_manager"
version: "1.0.0"
category: "utility"
tags:
  - files
  - utilities
  - productivity
description: "高效文件管理工具"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "find"
      version: "any"
      install:
        apt: "find ships with GNU findutils"
        brew: "find ships with macOS"
        apk: "apk add findutils"
    - name: "grep"
      version: "any"
      install:
        apt: "grep ships with GNU grep"
        brew: "grep ships with macOS"
        apk: "apk add grep"
  env: []

triggers:
  - type: "command"
    pattern: "/file"
    priority: 10
  - type: "keyword"
    pattern: "file manager"
    priority: 5

actions:
  - id: "find_large_files"
    type: "shell"
    description: "查找大于指定大小的文件"
    config:
      command: "find {{.Path}} -type f -size +{{.Size}} -exec ls -lh {} \\; | awk '{print $5, $9}'"
    timeout: "60s"

  - id: "find_recent_files"
    type: "shell"
    description: "查找最近修改的文件"
    config:
      command: "find {{.Path}} -type f -mtime -{{.Days}} -exec ls -lt {} \\; | head -{{.Limit}}"
    timeout: "30s"

  - id: "search_content"
    type: "shell"
    description: "在文件中搜索内容"
    config:
      command: "grep -r \"{{.Pattern}}\" {{.Path}} --include=\"{{.Include}}\" -n"
    timeout: "60s"

  - id: "count_files"
    type: "shell"
    description: "统计目录中的文件数量"
    config:
      command: "find {{.Path}} -type f | wc -l"
    timeout: "30s"

  - id: "disk_usage"
    type: "shell"
    description: "显示磁盘使用情况"
    config:
      command: "du -sh {{.Path}}/* 2>/dev/null | sort -hr | head -{{.Limit}}"
    timeout: "30s"

  - id: "clean_temp"
    type: "shell"
    description: "清理临时文件"
    config:
      command: "find {{.Path}} -type f \\( -name \"*.tmp\" -o -name \"*.log\" -o -name \"*.cache\" \\) -mtime +{{.Days}} -delete"
    timeout: "60s"

  - id: "list_duplicates"
    type: "shell"
    description: "列出重复文件（基于大小）"
    config:
      command: "find {{.Path}} -type f -not -path \"*/.*\" -exec du -h {} \\; | sort -hr | uniq -d -f1"
    timeout: "120s"

config:
  max_output_size: 10485760
  max_execution_time: "120s"
  enable_cache: true
  cache_ttl: "10m"

always: false

---

# 文件管理技能

高效的文件管理工具集，帮助快速定位、整理和清理文件。

## 功能

- **查找大文件**: 快速定位占用空间的文件
- **查找最近文件**: 查看最近修改的文件
- **内容搜索**: 在文件中搜索文本
- **文件统计**: 统计文件数量
- **磁盘使用**: 分析目录磁盘占用
- **清理临时文件**: 自动清理过期临时文件
- **查找重复**: 发现潜在的重复文件

## 使用示例

### 查找大文件

```bash
# 查找大于 100MB 的文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_large_files --vars "Path=.,Size=100M"

# 查找大于 1GB 的文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_large_files --vars "Path=/home/user,Size=1G"
```

### 查找最近文件

```bash
# 查找最近 7 天修改的文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_recent_files --vars "Path=.,Days=7,Limit=20"

# 查找最近 24 小时修改的文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_recent_files --vars "Path=.,Days=1,Limit=50"
```

### 内容搜索

```bash
# 在所有文件中搜索 "TODO"
agentframework enhanced-skill execute com.agentframework.skill.file_manager search_content --vars "Path=.,Pattern=TODO,Include=*"

# 在代码文件中搜索 "function"
agentframework enhanced-skill execute com.agentframework.skill.file_manager search_content --vars "Path=src,Pattern=function,Include=*.go"
```

### 文件统计

```bash
# 统计当前目录文件数
agentframework enhanced-skill execute com.agentframework.skill.file_manager count_files --vars "Path=."

# 统计项目目录文件数
agentframework enhanced-skill execute com.agentframework.skill.file_manager count_files --vars "Path=/path/to/project"
```

### 磁盘使用分析

```bash
# 显示占用最大的 10 个目录
agentframework enhanced-skill execute com.agentframework.skill.file_manager disk_usage --vars "Path=.,Limit=10"

# 分析用户目录
agentframework enhanced-skill execute com.agentframework.skill.file_manager disk_usage --vars "Path=/home/user,Limit=20"
```

### 清理临时文件

```bash
# 清理 30 天前的临时文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager clean_temp --vars "Path=.,Days=30"

# 清理 7 天前的日志和缓存
agentframework enhanced-skill execute com.agentframework.skill.file_manager clean_temp --vars "Path=/tmp,Days=7"
```

### 查找重复文件

```bash
# 查找可能的重复文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager list_duplicates --vars "Path=."
```

## 参数说明

### 通用参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| Path | 目标路径 | 当前目录 | `.`, `/home/user`, `src` |
| Limit | 结果限制数 | 10 | `20`, `50` |
| Days | 天数 | 7 | `1`, `30` |

### find_large_files 参数

| 参数 | 说明 | 格式 |
|------|------|------|
| Size | 文件大小阈值 | `100M`, `1G`, `500K` |

### search_content 参数

| 参数 | 说明 | 示例 |
|------|------|------|
| Pattern | 搜索模式（支持正则） | `TODO`, `^import`, `function` |
| Include | 文件匹配模式 | `*.go`, `*.txt`, `*` |

## 使用技巧

### 1. 定期清理

```bash
# 创建定期清理任务
# 每周清理 30 天前的临时文件
agentframework enhanced-skill execute com.agentframework.skill.file_manager clean_temp --vars "Path=/tmp,Days=30"
```

### 2. 磁盘空间监控

```bash
# 检查大文件占用
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_large_files --vars "Path=.,Size=500M"

# 分析目录占用
agentframework enhanced-skill execute com.agentframework.skill.file_manager disk_usage --vars "Path=.,Limit=20"
```

### 3. 代码维护

```bash
# 查找 TODO 注释
agentframework enhanced-skill execute com.agentframework.skill.file_manager search_content --vars "Path=src,Pattern=TODO,Include=*.go"

# 查找调试语句
agentframework enhanced-skill execute com.agentframework.skill.file_manager search_content --vars "Path=src,Pattern=console.log,Include=*.js"
```

## 注意事项

⚠️ **危险操作警告**

- `clean_temp` 操作会**永久删除**文件
- 执行前请确认 Path 参数正确
- 建议先使用 `find_recent_files` 验证
- 无法撤销删除操作

## 性能优化

- 对于大型目录，增加 `timeout` 值
- 使用具体的 Path 而非 `.`
- 限制搜索结果数量
- 使用 Include 参数过滤文件类型

## 故障排除

### 命令执行缓慢
1. 缩小搜索范围（Path）
2. 减少结果数量（Limit）
3. 使用更具体的文件模式（Include）

### 权限错误
1. 检查目录读取权限
2. 使用 `sudo` 运行（如果需要）
3. 避免扫描系统目录

### 未找到结果
1. 检查路径是否正确
2. 调整搜索参数（Days, Size）
3. 验证搜索模式（Pattern）
