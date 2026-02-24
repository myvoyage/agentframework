# AgentFramework 集成部署与问题解决指南

**创建时间**: 2025-02-19
**目的**: 指导部署集成后的系统并解决已知问题

---

## 📋 部署前检查清单

### 环境要求

- ✅ Go 1.25+
- ✅ Node.js 18+
- ✅ npm 或 yarn
- ✅ Git

### 依赖安装

#### 后端依赖

```bash
cd AgentFramework

# 使用国内镜像加速（推荐）
export GOPROXY=https://goproxy.cn,direct

# 安装依赖
go mod download

# 验证依赖
go mod verify
```

#### 前端依赖

```bash
cd frontend

# 安装依赖
npm install

# 或使用 yarn
yarn install
```

---

## 🔧 已知问题与解决方案

### 问题 1: 导入循环 (P0 - 阻塞性)

#### 症状

```
import cycle not allowed:
AgentFramework/pkg/iot ←→ AgentFramework/pkg/iot/adapters
```

#### 影响

- 阻止 `core` 包编译
- 不影响其他包（如 `pkg/auth`, `pkg/validation`, `pkg/rbac`）

#### 解决方案

**方案 A: 重构包结构（推荐）**

```bash
# 1. 分析依赖关系
go list -f '{{.ImportPath}}: {{join .Deps "\n"}}' ./pkg/iot/...

# 2. 创建新的包结构
mkdir -p pkg/iot/core
mkdir -p pkg/iot/adapters

# 3. 移动共享代码到 pkg/iot/core
# 4. 让 adapters 依赖 core，而不是相反
```

**方案 B: 使用接口（临时方案）**

```go
// pkg/iot/types.go
package iot

type Adapter interface {
    Connect() error
    Disconnect() error
    Send(data []byte) error
}

// pkg/iot/adapters/mqtt.go
package adapters

import "AgentFramework/pkg/iot"

type MQTTAdapter struct {
    // 依赖 iot.Adapter 接口，而不是具体实现
}
```

**方案 C: 合并包（最快）**

```bash
# 将 adapters 合并到 iot 包中
mv pkg/iot/adapters/*.go pkg/iot/
rm -rf pkg/iot/adapters

# 更新所有导入
find . -name "*.go" -exec sed -i 's|AgentFramework/pkg/iot/adapters|AgentFramework/pkg/iot|g' {} \;
```

#### 推荐行动

1. **立即**: 使用方案 C 快速解决（5分钟）
2. **短期**: 使用方案 A 重构（2小时）
3. **长期**: 重新设计架构（未来迭代）

---

### 问题 2: 前端 TypeScript 配置 (P1 - 重要)

#### 症状

```
error TS2583: Cannot find name 'Map'.
error TS2583: Cannot find name 'Set'.
error TS2705: An async function or method in ES5 requires the 'Promise' constructor
```

#### 解决方案

**更新 `frontend/tsconfig.json`**:

```json
{
  "compilerOptions": {
    "target": "ESNext",  // 或 "ES2020"
    "lib": [
      "ESNext",
      "DOM",
      "DOM.Iterable"  // 添加这一行
    ],
    // ... 其他配置
  }
}
```

**验证**:

```bash
cd frontend
npm run build
```

---

### 问题 3: 测试依赖缺失 (P2 - 一般)

#### 症状

```
error TS2307: Cannot find module '@vue/test-utils'
error TS2307: Cannot find module 'vitest'
```

#### 解决方案

```bash
cd frontend

# 安装测试依赖
npm install --save-dev @vue/test-utils vitest

# 或移除测试文件（如果不需要）
rm src/views/*.test.ts
```

---

## 🚀 部署步骤

### 1. 准备阶段

```bash
# 1.1 拉取最新代码
cd AgentFramework
git pull origin main

# 1.2 检查分支
git branch
git status

# 1.3 安装依赖
export GOPROXY=https://goproxy.cn,direct
go mod download
cd frontend && npm install && cd ..
```

### 2. 构建阶段

#### 后端构建

```bash
# 2.1 解决导入循环（使用方案 C）
mv pkg/iot/adapters/*.go pkg/iot/
rm -rf pkg/iot/adapters
find . -name "*.go" -type f -exec sed -i 's|AgentFramework/pkg/iot/adapters|AgentFramework/pkg/iot|g' {} \;

# 2.2 验证编译
go build -v ./...

# 2.3 构建二进制文件
go build -o bin/agent-framework ./cmd/agent-cli

# 2.4 验证二进制
./bin/agent-framework --version
./bin/agent-framework --help
```

#### 前端构建

```bash
cd frontend

# 2.5 更新 TypeScript 配置
# 手动编辑 tsconfig.json 添加 "DOM.Iterable" 到 lib

# 2.6 构建前端
npm run build

# 2.7 验证构建产物
ls -lh dist/
```

### 3. 运行阶段

#### 开发环境

```bash
# 后端（终端 1）
cd AgentFramework
go run cmd/agent-cli/main.go serve --config host.yaml --enhanced

# 前端（终端 2）
cd frontend
npm run dev
```

#### 生产环境

```bash
# 3.1 构建生产版本
cd frontend
npm run build
cd ..

# 3.2 启动后端服务
./bin/agent-framework serve \
  --config /etc/agentframework/config.yaml \
  --enhanced \
  --port 8080

# 3.3 配置反向代理（Nginx）
cat > /etc/nginx/sites-available/agent-framework <<EOF
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /var/www/agent-framework/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    # WebSocket 支持
    location /ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# 3.4 启用站点
ln -s /etc/nginx/sites-available/agent-framework /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

---

## 🧪 测试验证

### 单元测试

```bash
# 后端测试
go test ./pkg/auth/...
go test ./pkg/validation/...
go test ./pkg/rbac/...
go test ./pkg/cache/...
go test ./pkg/pool/...
go test ./pkg/lockfree/...

# 前端测试
cd frontend
npm run test
```

### 集成测试

```bash
# 1. 启动增强服务
go run cmd/agent-cli/main.go serve --config host.yaml --enhanced &

# 2. 测试安全功能
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"token":"test-jwt-token"}'

# 3. 测试性能监控
curl http://localhost:8080/api/performance/metrics

# 4. 测试缓存操作
curl -X POST http://localhost:8080/api/cache/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"data","ttl":3600}'
```

### CLI 命令测试

```bash
# 测试新增的 CLI 命令
./bin/agent-framework validate --config host.yaml
./bin/agent-framework metrics --config host.yaml
./bin/agent-framework security --config host.yaml

# 测试缓存命令
./bin/agent-framework cache get --key "test"
./bin/agent-framework cache set --key "test" --value '{"data":"value"}'
./bin/agent-framework cache clear --all
```

---

## 📊 性能监控

### 启用监控

```bash
# 在生产环境启动监控
./bin/agent-framework serve \
  --config /etc/agentframework/config.yaml \
  --enhanced \
  --monitoring-enabled \
  --metrics-port 9090
```

### 访问监控端点

```bash
# Prometheus 指标
curl http://localhost:9090/metrics

# 健康检查
curl http://localhost:9090/health

# 性能统计
curl http://localhost:9090/stats
```

---

## 🔍 故障排查

### 日志查看

```bash
# 后端日志
tail -f /var/log/agentframework/app.log

# 前端日志（浏览器控制台）
# 打开开发者工具 -> Console

# 系统日志
journalctl -u agent-framework -f
```

### 常见问题

#### 1. JWT 验证失败

**症状**: `Invalid JWT token`

**解决**:
```bash
# 检查 JWT_SECRET 配置
grep JWT_SECRET /etc/agentframework/config.yaml

# 确保前后端使用相同的密钥
```

#### 2. 权限拒绝

**症状**: `Access denied`

**解决**:
```bash
# 检查用户角色
curl http://localhost:8080/api/users/me

# 分配管理员权限
./bin/agent-framework rbac assign --user admin --role admin
```

#### 3. 性能下降

**症状**: 响应时间变长

**解决**:
```bash
# 1. 查看性能指标
curl http://localhost:8080/api/performance/metrics

# 2. 清空缓存
./bin/agent-framework cache clear --all

# 3. 重启服务
systemctl restart agent-framework
```

---

## 📈 优化建议

### 短期优化（1周内）

1. **解决导入循环问题**
   - 优先级: P0
   - 预计时间: 2-4小时
   - 影响: 解除编译阻塞

2. **完善 TypeScript 配置**
   - 优先级: P1
   - 预计时间: 10分钟
   - 影响: 前端编译通过

3. **添加基础测试**
   - 优先级: P1
   - 预计时间: 1-2天
   - 影响: 保证代码质量

### 中期优化（1个月内）

1. **实现 UI 页面**
   - 安全设置页面
   - 性能监控页面
   - 诊断工具页面

2. **完善 API**
   - RESTful API 标准化
   - API 文档（Swagger）
   - API 版本控制

3. **监控完善**
   - Prometheus 集成
   - Grafana 仪表板
   - 告警规则配置

### 长期优化（3个月内）

1. **架构重构**
   - 微服务拆分
   - 消息队列集成
   - 服务网格部署

2. **性能优化**
   - 数据库优化
   - 缓存策略优化
   - CDN 部署

3. **安全加固**
   - OAuth 2.0 集成
   - 多因素认证
   - 审计日志完善

---

## 📞 支持与反馈

- **GitHub Issues**: https://github.com/myvoyage/agentframework/issues
- **GitHub Discussions**: https://github.com/myvoyage/agentframework/discussions
- **文档**: https://github.com/myvoyage/agentframework/docs

---

## ✅ 部署检查清单

### 部署前

- [ ] 环境依赖已安装
- [ ] 代码已拉取最新
- [ ] 配置文件已准备
- [ ] 依赖已安装

### 部署中

- [ ] 导入循环已解决
- [ ] TypeScript 配置已更新
- [ ] 后端编译成功
- [ ] 前端编译成功

### 部署后

- [ ] 服务正常启动
- [ ] 日志无错误
- [ ] API 可访问
- [ ] UI 可访问
- [ ] 监控正常工作

### 测试

- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] CLI 命令正常
- [ ] 性能指标正常
- [ ] 安全功能正常

---

**文档版本**: 1.0
**最后更新**: 2025-02-19
**维护者**: AgentFramework Team

---

*祝部署顺利！* 🚀✨
