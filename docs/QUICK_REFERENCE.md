# AgentFramework 集成快速参考

**创建时间**: 2025-02-19
**用途**: 快速查找关键文件、命令和配置

---

## 📁 关键文件索引

### 后端核心文件

| 文件 | 行数 | 功能 | 状态 |
|------|------|------|------|
| [core/enhanced_application.go](../core/enhanced_application.go) | 449 | 增强核心应用 | ✅ |
| [app_enhanced.go](../app_enhanced.go) | 210 | 增强桌面应用 | ✅ |
| [cmd/agent-cli/main_enhanced.go](../cmd/agent-cli/main_enhanced.go) | 520+ | 增强 CLI | ✅ |
| [pkg/validation/input_validator.go](../pkg/validation/input_validator.go) | 新建 | 输入验证器 | ✅ |
| [pkg/validation/middleware.go](../pkg/validation/middleware.go) | 修复 | HTTP 中间件 | ✅ |

### 前端核心文件

| 文件 | 行数 | 功能 | 状态 |
|------|------|------|------|
| [frontend/src/stores/securityStore.ts](../frontend/src/stores/securityStore.ts) | 269 | 安全状态管理 | ✅ |
| [frontend/src/stores/performanceStore.ts](../frontend/src/stores/performanceStore.ts) | 250+ | 性能状态管理 | ✅ |
| [frontend/src/stores/appStore.ts](../frontend/src/stores/appStore.ts) | 110 | 应用状态管理 | ✅ |
| [frontend/src/main.ts](../frontend/src/main.ts) | 更新 | 前端入口 | ✅ |

### 文档文件

| 文件 | 类型 | 用途 | 状态 |
|------|------|------|------|
| [docs/MAIN_INTEGRATION_COMPLETE.md](MAIN_INTEGRATION_COMPLETE.md) | 集成文档 | 完整集成说明 | ✅ |
| [docs/INTEGRATION_TEST_REPORT.md](INTEGRATION_TEST_REPORT.md) | 测试报告 | 测试结果汇总 | ✅ |
| [docs/DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) | 部署指南 | 部署和故障排查 | ✅ |
| [docs/FINAL_INTEGRATION_SUMMARY.md](FINAL_INTEGRATION_SUMMARY.md) | 项目总结 | 项目总结和评价 | ✅ |
| [docs/QUICK_REFERENCE.md](QUICK_REFERENCE.md) | 本文档 | 快速参考 | ✅ |

### 工具脚本

| 文件 | 平台 | 功能 | 状态 |
|------|------|------|------|
| [scripts/fix-import-cycle.sh](../scripts/fix-import-cycle.sh) | Linux/Mac | 修复导入循环 | ✅ |
| [scripts/fix-import-cycle.bat](../scripts/fix-import-cycle.bat) | Windows | 修复导入循环 | ✅ |

---

## 🚀 快速启动

### 开发环境

```bash
# 1. 修复已知问题
# Windows
scripts\fix-import-cycle.bat

# Linux/Mac
chmod +x scripts/fix-import-cycle.sh && ./scripts/fix-import-cycle.sh

# 2. 更新前端配置
# 编辑 frontend/tsconfig.json，添加 "DOM.Iterable" 到 lib

# 3. 启动后端
go run cmd/agent-cli/main.go serve --config host.yaml --enhanced

# 4. 启动前端
cd frontend && npm run dev
```

### 生产环境

```bash
# 1. 构建
cd frontend && npm run build && cd ..
go build -o bin/agent-framework ./cmd/agent-cli

# 2. 运行
./bin/agent-framework serve --config /etc/agentframework/config.yaml --enhanced
```

---

## 🔧 CLI 命令速查

### 服务命令

```bash
# 启动增强服务
agent-framework serve --config host.yaml --enhanced

# 启动并指定端口
agent-framework serve --config host.yaml --port 8080

# 启用监控
agent-framework serve --config host.yaml --monitoring-enabled
```

### 验证命令

```bash
# 验证配置
agent-framework validate --config host.yaml

# 验证安全配置
agent-framework security --config host.yaml --check-security

# 验证性能配置
agent-framework validate --config host.yaml --check-performance
```

### 指标命令

```bash
# 显示性能指标
agent-framework metrics --config host.yaml

# 显示实时指标
agent-framework metrics --config host.yaml --watch

# 导出指标
agent-framework metrics --config host.yaml --export metrics.json
```

### 安全命令

```bash
# 运行安全诊断
agent-framework security --config host.yaml

# 检查权限
agent-framework security check-permission --user admin --resource agent --action execute

# 分配角色
agent-framework security assign-role --user john --role operator
```

### 缓存命令

```bash
# 获取缓存
agent-framework cache get --key "user:123"

# 设置缓存
agent-framework cache set --key "user:123" --value '{"name": "John"}' --ttl 3600

# 删除缓存
agent-framework cache delete --key "user:123"

# 清空缓存
agent-framework cache clear --all

# 显示缓存统计
agent-framework cache stats
```

---

## 📡 API 端点速查

### 认证 API

```bash
# 登录
POST /api/auth/login
Body: {"token": "jwt-token"}

# 登出
POST /api/auth/logout

# 获取当前用户
GET /api/auth/me
```

### 安全 API

```bash
# 检查权限
GET /api/security/permission?resource=agent&action=execute
Headers: Authorization: Bearer <token>

# 验证输入
POST /api/security/validate
Body: {"input": "user input", "maxLength": 1000}

# 获取安全配置
GET /api/security/config
```

### 性能 API

```bash
# 获取性能指标
GET /api/performance/metrics

# 获取性能报告
GET /api/performance/report

# 获取性能建议
GET /api/performance/recommendations

# 导出指标
GET /api/performance/export?format=json
```

### 缓存 API

```bash
# 获取缓存
GET /api/cache/{key}

# 设置缓存
PUT /api/cache/{key}
Body: {"value": "...", "ttl": 3600}

# 删除缓存
DELETE /api/cache/{key}

# 清空缓存
POST /api/cache/clear
```

### 诊断 API

```bash
# 健康检查
GET /api/health

# 系统状态
GET /api/diagnostics/status

# 运行诊断
POST /api/diagnostics/run
Body: {"type": "security|performance|full"}
```

---

## 🎨 前端使用速查

### 安全功能

```typescript
import { useSecurityStore } from '@/stores/securityStore'

const securityStore = useSecurityStore()

// 登录
await securityStore.login(token, {
  id: 'user-123',
  name: 'John Doe',
  roles: ['user', 'viewer']
})

// 检查权限
if (securityStore.hasPermission('agent', 'execute')) {
  // 允许执行
}

// 验证输入
const { valid, error } = securityStore.validateInput(userInput)
if (!valid) {
  console.error(error)
}
```

### 性能监控

```typescript
import { usePerformanceStore } from '@/stores/performanceStore'

const performanceStore = usePerformanceStore()

// 启动监控
performanceStore.startMonitoring(5000) // 每5秒更新

// 获取报告
const report = performanceStore.getPerformanceReport()
console.log('状态:', report.summary.status)
console.log('建议:', report.recommendations)

// 导出数据
performanceStore.exportMetrics()
```

---

## 🔧 配置速查

### 后端配置 (host.yaml)

```yaml
# JWT 配置
jwt_secret: "your-secret-key-change-in-production"
jwt_algorithm: "HS256"
audience: "agent-framework"

# RBAC 配置
admin_user_id: "admin"

# 缓存配置
redis_enabled: true
redis_url: "localhost:6379"

# 监控配置
monitoring_enabled: true
metrics_port: 9090
```

### 前端配置 (.env)

```bash
# API 地址
VITE_API_BASE_URL=http://localhost:8080

# WebSocket 地址
VITE_WS_BASE_URL=ws://localhost:8080

# 监控配置
VITE_MONITORING_ENABLED=true
VITE_MONITORING_INTERVAL=5000
```

---

## 🧪 测试速查

### 单元测试

```bash
# 后端测试
go test ./pkg/auth/...
go test ./pkg/validation/...
go test ./pkg/rbac/...
go test ./pkg/cache/...
go test ./pkg/pool/...

# 运行所有测试
go test ./...

# 带覆盖率
go test -cover ./...
```

### 集成测试

```bash
# 启动测试服务
go test -tags=integration ./...

# 端到端测试
cd frontend
npm run test:e2e
```

### 性能测试

```bash
# 基准测试
go test -bench=. -benchmem ./...

# 压力测试
go test -bench=BenchmarkHighConcurrency ./...
```

---

## 🐛 故障排查速查

### 常见错误

#### 1. 编译错误 - 导入循环

```
import cycle not allowed
```

**解决**: 运行修复脚本
```bash
scripts/fix-import-cycle.bat  # Windows
./scripts/fix-import-cycle.sh # Linux/Mac
```

#### 2. TypeScript 错误 - 找不到 Map/Set

```
error TS2583: Cannot find name 'Map'
```

**解决**: 更新 `tsconfig.json`
```json
{
  "compilerOptions": {
    "lib": ["ESNext", "DOM", "DOM.Iterable"]
  }
}
```

#### 3. JWT 验证失败

```
Invalid JWT token
```

**解决**: 检查 JWT_SECRET 配置
```bash
grep JWT_SECRET config.yaml
```

#### 4. 权限拒绝

```
Access denied
```

**解决**: 分配正确角色
```bash
agent-framework security assign-role --user <user> --role admin
```

#### 5. 性能下降

**解决**:
1. 查看指标: `agent-framework metrics`
2. 清空缓存: `agent-framework cache clear --all`
3. 重启服务

---

## 📊 监控指标

### 关键指标

| 指标 | 说明 | 健康阈值 |
|------|------|----------|
| `request_count` | 请求总数 | - |
| `error_count` | 错误总数 | < 1% |
| `average_latency` | 平均延迟 | < 10ms |
| `cache_hit_rate` | 缓存命中率 | > 80% |
| `pool_reuse_rate` | 对象池复用率 | > 70% |
| `throughput` | 吞吐量 | > 1000 req/s |

### 监控命令

```bash
# 实时监控
watch -n 5 'agent-framework metrics --config host.yaml'

# 导出监控数据
agent-framework metrics --config host.yaml --export metrics.json

# Prometheus 指标
curl http://localhost:9090/metrics
```

---

## 🔗 相关链接

- **GitHub**: https://github.com/myvoyage/agentframework
- **文档**: [docs/](.)
- **Issues**: https://github.com/myvoyage/agentframework/issues
- **Discussions**: https://github.com/myvoyage/agentframework/discussions

---

## 📚 文档导航

1. **新用户入门**: [MAIN_INTEGRATION_COMPLETE.md](MAIN_INTEGRATION_COMPLETE.md)
2. **测试验证**: [INTEGRATION_TEST_REPORT.md](INTEGRATION_TEST_REPORT.md)
3. **部署生产**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
4. **项目总结**: [FINAL_INTEGRATION_SUMMARY.md](FINAL_INTEGRATION_SUMMARY.md)
5. **快速参考**: [QUICK_REFERENCE.md](QUICK_REFERENCE.md) (本文档)

---

**快速参考版本**: 1.0
**最后更新**: 2025-02-19
**维护者**: AgentFramework Team

---

*快速查询，快速解决问题！* ⚡🎯
