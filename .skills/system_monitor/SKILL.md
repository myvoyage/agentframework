---
id: "com.agentframework.skill.system_monitor"
name: "system_monitor"
version: "1.0.0"
category: "system"
tags:
  - monitoring
  - system
  - performance
description: "系统资源监控工具"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "df"
      version: "any"
      install:
        apt: "df ships with coreutils"
        brew: "df ships with macOS"
        apk: "apk add coreutils"
  env: []

triggers:
  - type: "command"
    pattern: "/monitor"
    priority: 10
  - type: "keyword"
    pattern: "system monitor"
    priority: 5

actions:
  - id: "disk_usage"
    type: "shell"
    description: "显示磁盘使用情况"
    config:
      command: "df -h"
    timeout: "5s"

  - id: "memory_info"
    type: "shell"
    description: "显示内存使用情况"
    config:
      command: "free -h 2>/dev/null || vm_stat | head -5"
    timeout: "5s"

  - id: "cpu_info"
    type: "shell"
    description: "显示 CPU 信息"
    config:
      command: "uptime"
    timeout: "5s"

  - id: "load_average"
    type: "shell"
    description: "显示系统负载"
    config:
      command: "uptime"
    timeout: "5s"

  - id: "process_count"
    type: "shell"
    description: "统计进程数量"
    config:
      command: "ps aux 2>/dev/null | wc -l || ps -al | wc -l"
    timeout: "10s"

  - id: "quick_health"
    type: "shell"
    description: "快速健康检查"
    config:
      command: "echo '=== System Health ===' && df -h | head -5 && uptime"
    timeout: "10s"

config:
  max_output_size: 1048576
  max_execution_time: "30s"
  enable_cache: true
  cache_ttl: "1m"

always: false

---

# 系统监控器技能

实时系统资源监控工具，帮助了解系统健康状况。

## 功能

- **磁盘监控**: 查看磁盘使用情况
- **内存监控**: 监控内存使用率
- **CPU 信息**: 查看 CPU 核心和负载
- **进程统计**: 系统进程数量
- **健康检查**: 快速系统健康报告

## 使用示例

### 快速健康检查

```bash
agentframework enhanced-skill execute com.agentframework.skill.system_monitor quick_health --vars ""
```

### 磁盘监控

```bash
agentframework enhanced-skill execute com.agentframework.skill.system_monitor disk_usage --vars ""
```

### 内存监控

```bash
agentframework enhanced-skill execute com.agentframework.skill.system_monitor memory_info --vars ""
```

### CPU 监控

```bash
agentframework enhanced-skill execute com.agentframework.skill.system_monitor cpu_info --vars ""
```

## 参数说明

所有命令都不需要参数，直接执行即可。

## 故障排除

### 命令未找到

确保安装了必要的系统工具：
- Linux: `coreutils`, `procps`
- macOS: 大多数命令已内置
- WSL: 使用 Linux 版本的工具

### 权限问题

某些命令可能需要普通权限即可运行。
