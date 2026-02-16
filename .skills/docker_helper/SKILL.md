---
id: "com.agentframework.skill.docker_helper"
name: "docker_helper"
version: "1.0.0"
category: "development"
tags:
  - docker
  - containers
  - devops
description: "Docker 容器管理助手"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "docker"
      version: ">=20.0.0"
      install:
        apt: "curl -fsSL https://get.docker.com | sh"
        brew: "brew install --cask docker"
        apk: "apk add docker"
  env: []

triggers:
  - type: "command"
    pattern: "/docker"
    priority: 10
  - type: "keyword"
    pattern: "docker helper"
    priority: 5

actions:
  - id: "list_containers"
    type: "shell"
    description: "列出所有容器"
    config:
      command: "docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'"
    timeout: "10s"

  - id: "list_images"
    type: "shell"
    description: "列出所有镜像"
    config:
      command: "docker images --format 'table {{.Repository}}\t{{.Tag}}\t{{.Size}}'"
    timeout: "10s"

  - id: "container_stats"
    type: "shell"
    description: "显示容器统计"
    config:
      command: "docker stats --no-stream --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'"
    timeout: "15s"

  - id: "inspect_container"
    type: "shell"
    description: "检查容器详情"
    config:
      command: "docker inspect {{.Container}} | jq '.[0] | {Name: .Name, State: .State.Status, Image: .Config.Image}'"
    timeout: "10s"

  - id: "view_logs"
    type: "shell"
    description: "查看容器日志"
    config:
      command: "docker logs --tail {{.Tail}} {{.Container}}"
    timeout: "10s"

  - id: "start_container"
    type: "shell"
    description: "启动容器"
    config:
      command: "docker start {{.Container}}"
    timeout: "30s"

  - id: "stop_container"
    type: "shell"
    description: "停止容器"
    config:
      command: "docker stop {{.Container}}"
    timeout: "30s"

  - id: "restart_container"
    type: "shell"
    description: "重启容器"
    config:
      command: "docker restart {{.Container}}"
    timeout: "60s"

  - id: "remove_container"
    type: "shell"
    description: "删除容器"
    config:
      command: "docker rm -f {{.Container}}"
    timeout: "30s"

  - id: "remove_image"
    type: "shell"
    description: "删除镜像"
    config:
      command: "docker rmi {{.Image}}"
    timeout: "60s"

  - id: "cleanup_system"
    type: "shell"
    description: "清理未使用的资源"
    config:
      command: "docker system prune -f"
    timeout: "60s"

  - id: "network_info"
    type: "shell"
    description: "显示网络信息"
    config:
      command: "docker network ls"
    timeout: "10s"

  - id: "volume_info"
    type: "shell"
    description: "显示卷信息"
    config:
      command: "docker volume ls"
    timeout: "10s"

  - id: "exec_command"
    type: "shell"
    description: "在容器中执行命令"
    config:
      command: "docker exec -it {{.Container}} {{.Command}}"
    timeout: "60s"

config:
  max_output_size: 5242880
  max_execution_time: "120s"
  enable_cache: true
  cache_ttl: "2m"

always: false

---

# Docker 助手技能

简化 Docker 容器管理的实用工具。

## 功能

- **容器管理**: 列出、启动、停止、删除容器
- **镜像管理**: 列出、删除镜像
- **监控**: 容器资源使用统计
- **日志**: 查看容器日志
- **清理**: 清理未使用的资源
- **网络**: 管理网络配置
- **卷**: 管理数据卷

## 使用示例

### 容器管理

```bash
# 列出所有容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper list_containers --vars ""

# 启动容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper start_container --vars "Container=myapp"

# 停止容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper stop_container --vars "Container=myapp"

# 重启容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper restart_container --vars "Container=myapp"

# 删除容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper remove_container --vars "Container=myapp"
```

### 镜像管理

```bash
# 列出所有镜像
agentframework enhanced-skill execute com.agentframework.skill.docker_helper list_images --vars ""

# 删除镜像
agentframework enhanced-skill execute com.agentframework.skill.docker_helper remove_image --vars "Image=nginx:latest"
```

### 监控

```bash
# 查看容器统计
agentframework enhanced-skill execute com.agentframework.skill.docker_helper container_stats --vars ""

# 检查容器详情
agentframework enhanced-skill execute com.agentframework.skill.docker_helper inspect_container --vars "Container=myapp"
```

### 日志

```bash
# 查看最近 100 行日志
agentframework enhanced-skill execute com.agentframework.skill.docker_helper view_logs --vars "Container=myapp,Tail=100"

# 查看所有日志
agentframework enhanced-skill execute com.agentframework.skill.docker_helper view_logs --vars "Container=myapp,Tail=all"
```

### 清理

```bash
# 清理未使用的资源
agentframework enhanced-skill execute com.agentframework.skill.docker_helper cleanup_system --vars ""
```

### 执行命令

```bash
# 在容器中执行命令
agentframework enhanced-skill execute com.agentframework.skill.docker_helper exec_command --vars "Container=myapp,Command=bash"

# 运行特定命令
agentframework enhanced-skill execute com.agentframework.skill.docker_helper exec_command --vars "Container=myapp,Command=ls -la"
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| Container | 容器名称或 ID | `myapp`, `abc123` |
| Image | 镜像名称 | `nginx:latest` |
| Tail | 日志行数 | `100`, `all` |
| Command | 要执行的命令 | `bash`, `ls -la` |

## 工作流示例

### 1. 部署新应用

```bash
# 1. 拉取镜像
docker pull nginx:latest

# 2. 运行容器
docker run -d --name myapp -p 80:80 nginx:latest

# 3. 查看状态
agentframework enhanced-skill execute com.agentframework.skill.docker_helper list_containers --vars ""

# 4. 查看日志
agentframework enhanced-skill execute com.agentframework.skill.docker_helper view_logs --vars "Container=myapp,Tail=50"
```

### 2. 维护现有应用

```bash
# 1. 查看统计
agentframework enhanced-skill execute com.agentframework.skill.docker_helper container_stats --vars ""

# 2. 重启容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper restart_container --vars "Container=myapp"

# 3. 查看日志
agentframework enhanced-skill execute com.agentframework.skill.docker_helper view_logs --vars "Container=myapp,Tail=100"
```

### 3. 清理资源

```bash
# 1. 停止所有不需要的容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper stop_container --vars "Container=old-app"

# 2. 删除容器
agentframework enhanced-skill execute com.agentframework.skill.docker_helper remove_container --vars "Container=old-app"

# 3. 清理系统
agentframework enhanced-skill execute com.agentframework.skill.docker_helper cleanup_system --vars ""
```

## 最佳实践

### 1. 定期清理

```bash
# 每周清理一次
agentframework enhanced-skill execute com.agentframework.skill.docker_helper cleanup_system --vars ""
```

### 2. 监控资源

```bash
# 定期检查容器状态
watch -n 5 'agentframework enhanced-skill execute com.agentframework.skill.docker_helper container_stats --vars ""'
```

### 3. 日志管理

```bash
# 定期检查日志大小
docker logs myapp | wc -l

# 限制日志大小
docker update --log-opt max-size=10m myapp
```

## 故障排除

### 容器无法启动

```bash
# 查看容器日志
agentframework enhanced-skill execute com.agentframework.skill.docker_helper view_logs --vars "Container=myapp,Tail=100"

# 检查容器状态
docker ps -a | grep myapp

# 检查容器详情
agentframework enhanced-skill execute com.agentframework.skill.docker_helper inspect_container --vars "Container=myapp"
```

### 资源不足

```bash
# 查看 Docker 磁盘使用
docker system df

# 清理未使用的资源
agentframework enhanced-skill execute com.agentframework.skill.docker_helper cleanup_system --vars ""
```

### 网络问题

```bash
# 查看网络配置
agentframework enhanced-skill execute com.agentframework.skill.docker_helper network_info --vars ""

# 检查容器网络
docker network inspect bridge
```

## 安全建议

### 1. 使用非 root 用户

```bash
# 将用户添加到 docker 组
sudo usermod -aG docker $USER
```

### 2. 限制资源

```bash
# 限制内存和 CPU
docker run -m 512m --cpus=1.0 myapp
```

### 3. 使用官方镜像

```bash
# 优先使用官方镜像
docker pull nginx:latest
docker pull python:3.9-slim
```

### 4. 定期更新

```bash
# 更新基础镜像
docker pull nginx:latest
```

## 相关命令

### Docker Compose

虽然这个技能专注于单个容器，但推荐使用 Docker Compose 管理多容器应用：

```yaml
version: '3'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: secret
```

```bash
docker-compose up -d
docker-compose ps
docker-compose logs
```

## 高级用法

### 1. 容器健康检查

```bash
# 检查容器健康状态
docker inspect --format='{{.State.Health.Status}}' myapp
```

### 2. 资源限制

```bash
# 设置内存限制
docker update -m 1g myapp

# 设置 CPU 限制
docker update --cpus=2.0 myapp
```

### 3. 网络配置

```bash
# 创建自定义网络
docker network create mynet

# 连接到网络
docker network connect mynet myapp
```

## 性能优化

### 1. 使用多阶段构建

```dockerfile
# 构建阶段
FROM golang:1.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o app

# 运行阶段
FROM alpine:latest
COPY --from=builder /app/app /app
CMD ["/app"]
```

### 2. 清理不需要的文件

```bash
# 定期清理
docker system prune -a --volumes
```

### 3. 使用 .dockerignore

```
.git
node_modules
*.log
```

## 相关资源

- [Docker 官方文档](https://docs.docker.com/)
- [Docker 最佳实践](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
