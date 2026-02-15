# Docker 环境配置指南

## 概述

本文档指导如何配置 Docker 环境以支持容器执行模式。

---

## 安装 Docker

### Linux

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io
sudo systemctl start docker
sudo systemctl enable docker

# CentOS/RHEL
sudo yum install docker
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
docker ps
```

### macOS

1. 下载 [Docker Desktop for Mac](https://www.docker.com/products/docker-desktop)
2. 安装并启动 Docker Desktop
3. 验证: `docker --version`

### Windows

1. 下载 [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)
2. 安装并启动 Docker Desktop
3. 验证: `docker --version`

---

## 配置 Docker

### 1. 用户权限

```bash
# 添加用户到 docker 组
sudo usermod -aG docker $USER

# 重新登录或运行
newgrp docker

# 验证
docker ps
```

### 2. Docker 守护进程配置

创建 `/etc/docker/daemon.json`:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 64000,
      "Soft": 64000
    }
  }
}
```

重启 Docker:
```bash
sudo systemctl restart docker
```

### 3. 资源限制

**Docker Desktop (macOS/Windows)**:
- 打开 Docker Desktop
- Settings → Resources
- 设置 CPU、内存、磁盘

**Linux**:
```bash
# 编辑 /etc/docker/daemon.json
{
  "default-runtime": "runc",
  "runtimes": {
    "runc": {
      "path": "runc"
    }
  }
}
```

---

## 准备镜像

### 1. 拉取默认镜像

```bash
# Python
docker pull python:3.11-alpine

# JavaScript
docker pull node:18-alpine

# Go
docker pull golang:1.21-alpine

# Bash
docker pull alpine:latest
```

### 2. 验证镜像

```bash
docker images
```

输出示例:
```
REPOSITORY   TAG          SIZE
python       3.11-alpine  50MB
node         18-alpine    180MB
golang       1.21-alpine  300MB
alpine       latest       7MB
```

### 3. 自定义镜像

创建 `Dockerfile`:

```dockerfile
FROM python:3.11-alpine

# 安装额外的包
RUN pip install --no-cache-dir \
    requests \
    numpy \
    pandas

# 创建非 root 用户
RUN adduser -D appuser
USER appuser

WORKDIR /app
```

构建镜像:
```bash
docker build -t custom-python:latest .
```

配置使用:
```yaml
container:
  default_images:
    python: custom-python:latest
```

---

## 网络配置

### 1. 禁用网络（推荐）

```yaml
container:
  network_mode: none
```

### 2. 桥接网络

```yaml
container:
  network_mode: bridge
```

### 3. 自定义网络

```bash
# 创建网络
docker network create app-network

# 配置使用
docker run --network app-network ...
```

---

## 存储配置

### 1. 卷挂载

```yaml
container:
  volumes:
    - /host/path:/container/path:ro  # 只读
    - /tmp:/tmp:rw                    # 读写
```

### 2. 临时文件系统

```yaml
container:
  tmpfs:
    - /tmp
    - /var/tmp
```

### 3. 清理策略

```bash
# 清理未使用的容器
docker container prune -f

# 清理未使用的镜像
docker image prune -a -f

# 清理所有
docker system prune -a -f
```

---

## 性能优化

### 1. 使用 Alpine 镜像

```yaml
container:
  default_images:
    python: python:3.11-alpine      # 50 MB
    javascript: node:18-alpine      # 180 MB
    go: golang:1.21-alpine          # 300 MB
```

**优势**:
- 体积小
- 启动快
- 安全性高

### 2. 镜像缓存

```bash
# 预拉取镜像
docker pull python:3.11-alpine
docker pull node:18-alpine
docker pull golang:1.21-alpine
```

### 3. 容器池

```yaml
container:
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

---

## 安全配置

### 1. 运行时安全

```yaml
container:
  # 非 root 用户
  user: "1000:1000"
  
  # 只读根文件系统
  read_only: true
  
  # 禁用特权模式
  privileged: false
  
  # 限制能力
  cap_drop:
    - ALL
  cap_add:
    - NET_BIND_SERVICE
```

### 2. 资源限制

```yaml
container:
  cpu_limit: "0.5"
  memory_limit: "512m"
  pids_limit: 100
```

### 3. 安全选项

```yaml
container:
  security_opt:
    - no-new-privileges:true
    - seccomp=unconfined
```

---

## 监控和日志

### 1. 容器日志

```bash
# 查看日志
docker logs <container_id>

# 实时日志
docker logs -f <container_id>

# 最近 100 行
docker logs --tail 100 <container_id>
```

### 2. 资源监控

```bash
# 实时监控
docker stats

# 单个容器
docker stats <container_id>
```

### 3. 事件监控

```bash
# 监控事件
docker events

# 过滤事件
docker events --filter 'type=container'
```

---

## 故障排查

### 问题 1: Docker 未运行

**症状**: `Cannot connect to the Docker daemon`

**解决**:
```bash
# 检查状态
sudo systemctl status docker

# 启动 Docker
sudo systemctl start docker

# 设置开机启动
sudo systemctl enable docker
```

### 问题 2: 权限被拒绝

**症状**: `permission denied while trying to connect`

**解决**:
```bash
# 添加用户到 docker 组
sudo usermod -aG docker $USER

# 重新登录
newgrp docker
```

### 问题 3: 镜像拉取失败

**症状**: `Error response from daemon: Get https://...`

**解决**:
1. 检查网络连接
2. 配置镜像加速器
3. 使用代理

```json
// /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://mirror.example.com"
  ]
}
```

### 问题 4: 容器启动失败

**症状**: `Error starting container`

**解决**:
```bash
# 查看日志
docker logs <container_id>

# 检查配置
docker inspect <container_id>

# 测试镜像
docker run -it python:3.11-alpine /bin/sh
```

### 问题 5: 磁盘空间不足

**症状**: `no space left on device`

**解决**:
```bash
# 查看磁盘使用
docker system df

# 清理
docker system prune -a -f

# 清理卷
docker volume prune -f
```

---

## 最佳实践

### 1. 镜像管理

```bash
# 定期更新镜像
docker pull python:3.11-alpine
docker pull node:18-alpine

# 删除旧镜像
docker image prune -a -f

# 标记镜像
docker tag python:3.11-alpine myregistry/python:3.11
```

### 2. 容器生命周期

```bash
# 自动删除
docker run --rm ...

# 设置重启策略
docker run --restart=unless-stopped ...

# 健康检查
docker run --health-cmd='curl -f http://localhost/ || exit 1' ...
```

### 3. 资源管理

```bash
# 限制资源
docker run \
  --cpus="0.5" \
  --memory="512m" \
  --pids-limit=100 \
  ...
```

### 4. 安全加固

```bash
# 最小权限
docker run \
  --user=1000:1000 \
  --read-only \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  ...
```

---

## 生产环境配置

### 完整配置示例

```yaml
# config_prod.yaml
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 2
  execution_mode: container

container:
  enabled: true
  
  # 镜像配置
  default_images:
    python: python:3.11-alpine
    javascript: node:18-alpine
    go: golang:1.21-alpine
    bash: alpine:latest
  
  # 资源限制
  cpu_limit: "0.5"
  memory_limit: "512m"
  pids_limit: 100
  timeout: 30s
  
  # 网络配置
  network_mode: none
  
  # 安全配置
  user: "1000:1000"
  read_only: true
  privileged: false
  cap_drop:
    - ALL
  security_opt:
    - no-new-privileges:true
  
  # 清理配置
  auto_cleanup: true
  
  # 容器池
  enable_pool: true
  pool_min_size: 10
  pool_max_size: 50
```

---

## 检查清单

### 安装检查

- [ ] Docker 已安装
- [ ] Docker 服务运行中
- [ ] 用户有 Docker 权限
- [ ] 镜像已拉取
- [ ] 网络配置正确

### 安全检查

- [ ] 使用非 root 用户
- [ ] 禁用网络访问
- [ ] 限制资源使用
- [ ] 启用只读文件系统
- [ ] 配置安全选项

### 性能检查

- [ ] 使用 Alpine 镜像
- [ ] 启用容器池
- [ ] 配置资源限制
- [ ] 启用自动清理
- [ ] 监控资源使用

---

## 参考资料

- [Docker 官方文档](https://docs.docker.com/)
- [Docker 安全最佳实践](https://docs.docker.com/engine/security/)
- [Alpine Linux](https://alpinelinux.org/)

---

**版本**: 1.0  
**更新日期**: 2026-01-31

