---
id: "com.agentframework.skill.process_manager"
name: "process_manager"
version: "1.0.0"
category: "system"
tags:
  - processes
  - system
  - monitoring
description: "系统进程管理工具"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "ps"
      version: "any"
      install:
        apt: "ps ships with procps"
        brew: "ps ships with macOS"
        apk: "apk add procps"
    - name: "kill"
      version: "any"
      install:
        apt: "kill ships with util-linux"
        brew: "kill ships with macOS"
        apk: "apk add util-linux"
    - name: "pgrep"
      version: "any"
      install:
        apt: "pgrep ships with procps"
        brew: "pgrep ships with macOS"
        apk: "apk add procps"
  env: []

triggers:
  - type: "command"
    pattern: "/proc"
    priority: 10
  - type: "keyword"
    pattern: "process"
    priority: 5

actions:
  - id: "list_all"
    type: "shell"
    description: "列出所有进程"
    config:
      command: "ps aux --sort=-%mem | head -{{.Limit}}"
    timeout: "10s"

  - id: "find_by_name"
    type: "shell"
    description: "根据名称查找进程"
    config:
      command: "ps aux | grep {{.Pattern}} | grep -v grep"
    timeout: "10s"

  - id: "find_by_pid"
    type: "shell"
    description: "根据 PID 查找进程"
    config:
      command: "ps -p {{.PID}} -o pid,ppid,cmd,%mem,%cpu"
    timeout: "5s"

  - id: "kill_process"
    type: "shell"
    description: "杀死指定进程"
    config:
      command: "kill {{.Signal}} {{.PID}}"
    timeout: "5s"

  - id: "kill_by_name"
    type: "shell"
    description: "根据名称杀死进程"
    config:
      command: "pkill {{.Signal}} {{.Pattern}}"
    timeout: "10s"

  - id: "top_processes"
    type: "shell"
    description: "显示资源占用最高的进程"
    config:
      command: "ps aux --sort=-%cpu | head -{{.Limit}}"
    timeout: "10s"

  - id: "process_tree"
    type: "shell"
    description: "显示进程树"
    config:
      command: "ps auxf"
    timeout: "10s"

  - id: "process_stats"
    type: "shell"
    description: "显示进程统计信息"
    config:
      command: "ps aux | awk '{print $11}' | sort | uniq -c | sort -rn | head -{{.Limit}}"
    timeout: "15s"

config:
  max_output_size: 1048576
  max_execution_time: "30s"
  enable_cache: true
  cache_ttl: "2m"

always: false

---

# 进程管理器技能

强大的系统进程管理工具，帮助监控和管理运行中的进程。

## 功能

- **列出进程**: 查看所有运行中的进程
- **查找进程**: 根据名称或 PID 查找进程
- **终止进程**: 安全地终止指定的进程
- **资源分析**: 查看资源占用最高的进程
- **进程树**: 显示进程层次结构
- **统计信息**: 分析进程分布

## 使用示例

### 列出所有进程

```bash
# 列出前 20 个进程（按内存排序）
agentframework enhanced-skill execute com.agentframework.skill.process_manager list_all --vars "Limit=20"

# 列出前 50 个进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager list_all --vars "Limit=50"
```

### 查找进程

```bash
# 根据名称查找
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=chrome"

# 根据名称查找
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=python"

# 根据 PID 查找
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_pid --vars "PID=1234"
```

### 终止进程

```bash
# 温和地终止进程（SIGTERM）
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=1234,Signal=-15"

# 强制终止进程（SIGKILL）
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=1234,Signal=-9"

# 根据名称终止进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_by_name --vars "Pattern=firefox,Signal=-15"
```

### 资源分析

```bash
# 查看 CPU 占用最高的进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager top_processes --vars "Limit=10"

# 查看进程统计
agentframework enhanced-skill execute com.agentframework.skill.process_manager process_stats --vars "Limit=20"
```

### 进程树

```bash
# 显示完整进程树
agentframework enhanced-skill execute com.agentframework.skill.process_manager process_tree --vars ""
```

## 参数说明

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| Limit | 结果数量限制 | 20 | `10`, `50` |
| Pattern | 进程名称模式 | - | `chrome`, `python` |
| PID | 进程 ID | - | `1234` |
| Signal | 信号类型 | `-15` | `-9`, `-15` |

## 信号类型

| 信号 | 数值 | 说明 |
|------|------|------|
| SIGTERM | 15 | 温和地终止进程（默认） |
| SIGKILL | 9 | 强制终止进程 |
| SIGHUP | 1 | 挂起进程 |
| SIGINT | 2 | 中断进程（Ctrl+C） |

## 使用场景

### 1. 清理僵尸进程

```bash
# 查找僵尸进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=defunct"

# 查看所有进程状态
agentframework enhanced-skill execute com.agentframework.skill.process_manager list_all --vars "Limit=100"
```

### 2. 终止无响应程序

```bash
# 查找无响应的程序
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=program_name"

# 强制终止
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_by_name --vars "Pattern=program_name,Signal=-9"
```

### 3. 监控资源使用

```bash
# 查看 CPU 占用
agentframework enhanced-skill execute com.agentframework.skill.process_manager top_processes --vars "Limit=10"

# 查看内存占用
agentframework enhanced-skill execute com.agentframework.skill.process_manager list_all --vars "Limit=20"
```

### 4. 进程调试

```bash
# 查看进程树
agentframework enhanced-skill execute com.agentframework.skill.process_manager process_tree --vars ""

# 查找特定进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=app_name"
```

## 最佳实践

### 1. 优先使用 SIGTERM

```bash
# 推荐做法
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=1234,Signal=-15"

# 只有在必要时才使用 SIGKILL
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=1234,Signal=-9"
```

### 2. 确认进程 ID

```bash
# 先查找进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=app_name"

# 然后终止
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=实际PID,Signal=-15"
```

### 3. 批量操作

```bash
# 查找所有匹配的进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_name --vars "Pattern=python"

# 终止所有匹配的进程
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_by_name --vars "Pattern=python,Signal=-15"
```

## ⚠️ 安全警告

### 危险操作

1. **终止系统进程**: 不要终止 PID < 100 的系统进程
2. **强制终止**: SIGKILL 不会给进程清理的机会
3. **批量终止**: 确保只终止目标进程

### 保护措施

```bash
# 查看进程详情再决定
agentframework enhanced-skill execute com.agentframework.skill.process_manager find_by_pid --vars "PID=1234"

# 然后再决定是否终止
```

## 故障排除

### 权限不足

```bash
# 错误: Operation not permitted
# 解决: 使用 sudo 或检查权限
sudo kill -9 1234
```

### 进程不存在

```bash
# 错误: No such process
# 解决: 确认 PID 是否正确
ps -p 1234
```

### 无法终止进程

```bash
# 尝试更强的信号
agentframework enhanced-skill execute com.agentframework.skill.process_manager kill_process --vars "PID=1234,Signal=-9"
```

## 平台差异

| 功能 | Linux | macOS | Windows (WSL) |
|------|-------|-------|---------------|
| ps aux | ✅ | ✅ | ✅ |
| kill | ✅ | ✅ | ✅ |
| pgrep | ✅ | ✅ | ✅ |
| pkill | ✅ | ✅ | ✅ |

## 相关工具

- `top`: 实时进程监控
- `htop`: 交互式进程监控
- `atop`: 高级进程监控
- `pstree`: 进程树显示

## 技术细节

### 进程状态

- R: Running
- S: Sleeping
- D: Waiting
- Z: Zombie
- T: Stopped

### 进程优先级

- Nice 值: -20（最高优先级）到 19（最低优先级）
- 默认值: 0
