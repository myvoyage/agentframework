---
id: "com.agentframework.skill.network_diagnostics"
name: "network_diagnostics"
version: "1.0.0"
category: "system"
tags:
  - network
  - diagnostics
  - system
description: "网络诊断工具集"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "ping"
      version: "any"
      install:
        apt: "sudo apt install iputils-ping"
        brew: "ping ships with macOS"
        apk: "apk add iputils"
    - name: "traceroute"
      version: "any"
      install:
        apt: "sudo apt install traceroute"
        brew: "brew install traceroute"
        apk: "apk add traceroute"
    - name: "nslookup"
      version: "any"
      install:
        apt: "sudo apt install dnsutils"
        brew: "nslookup ships with macOS"
        apk: "apk add bind-tools"
  env: []

triggers:
  - type: "command"
    pattern: "/net"
    priority: 10
  - type: "keyword"
    pattern: "network"
    priority: 5

actions:
  - id: "ping_host"
    type: "shell"
    description: "Ping 主机测试连通性"
    config:
      command: "ping -c {{.Count}} {{.Host}}"
    timeout: "30s"

  - id: "trace_route"
    type: "shell"
    description: "追踪到主机的路由"
    config:
      command: "traceroute -m {{.MaxHops}} {{.Host}}"
    timeout: "60s"

  - id: "dns_lookup"
    type: "shell"
    description: "DNS 查询"
    config:
      command: "nslookup {{.Host}}"
    timeout: "10s"

  - id: "check_port"
    type: "shell"
    description: "检查端口是否开放"
    config:
      command: "nc -zv {{.Host}} {{.Port}}"
    timeout: "5s"

  - id: "network_info"
    type: "shell"
    description: "显示网络接口信息"
    config:
      command: "ip addr show || ifconfig"
    timeout: "5s"

config:
  max_output_size: 1048576
  max_execution_time: "60s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# 网络诊断技能

强大的网络诊断工具集，帮助快速定位网络问题。

## 功能

- **Ping 测试**: 测试到目标主机的连通性
- **路由追踪**: 查看到达目标主机的网络路径
- **DNS 查询**: 查询域名的 DNS 记录
- **端口检查**: 检查目标端口是否开放
- **网络信息**: 显示本机网络接口信息

## 使用示例

### Ping 测试

```bash
# Ping 4 次
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics ping_host --vars "Host=google.com,Count=4"

# Ping 10 次
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics ping_host --vars "Host=8.8.8.8,Count=10"
```

### 路由追踪

```bash
# 追踪路由（最多30跳）
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics trace_route --vars "Host=google.com,MaxHops=30"

# 追踪到本地网关
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics trace_route --vars "Host=192.168.1.1,MaxHops=10"
```

### DNS 查询

```bash
# 查询域名
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics dns_lookup --vars "Host=google.com"

# 查询特定记录
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics dns_lookup --vars "Host=github.com"
```

### 端口检查

```bash
# 检查 HTTP 端口
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics check_port --vars "Host=google.com,Port=80"

# 检查 HTTPS 端口
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics check_port --vars "Host=github.com,Port=443"
```

### 网络信息

```bash
# 显示网络接口
agentframework enhanced-skill execute com.agentframework.skill.network_diagnostics network_info --vars ""
```

## 故障排除

### Ping 失败
1. 检查网络连接是否正常
2. 确认目标主机是否在线
3. 检查防火墙设置

### DNS 查询失败
1. 检查 DNS 服务器配置
2. 尝试使用公共 DNS (8.8.8.8, 1.1.1.1)
3. 检查域名拼写是否正确

### 端口检查失败
1. 确认目标服务是否运行
2. 检查目标防火墙规则
3. 验证端口号是否正确

## 技术细节

### 支持的工具

- **ping**: ICMP echo 请求
- **traceroute**: 路由追踪
- **nslookup**: DNS 查询
- **nc (netcat)**: 端口扫描
- **ip/ifconfig**: 网络接口配置

### 平台兼容性

| 工具 | Linux | macOS | Windows |
|------|-------|-------|---------|
| ping | ✅ | ✅ | ✅ |
| traceroute | ✅ | ✅ | ✅ (tracert) |
| nslookup | ✅ | ✅ | ✅ |
| nc | ✅ | ✅ | ❌ |

### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| Host | 目标主机名或IP | 必填 |
| Count | Ping 次数 | 4 |
| MaxHops | 最大跳数 | 30 |
| Port | 端口号 | 必填 |
