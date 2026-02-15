# 安全配置指南

## 概述

本文档提供代码执行模块的安全配置建议和最佳实践。

---

## 安全威胁模型

### 潜在威胁

1. **恶意代码执行**
   - 系统命令注入
   - 文件系统破坏
   - 网络攻击

2. **资源滥用**
   - CPU 耗尽
   - 内存耗尽
   - 磁盘填满

3. **数据泄露**
   - 敏感文件读取
   - 环境变量泄露
   - 网络数据窃取

4. **权限提升**
   - 容器逃逸
   - 权限滥用

---

## 安全配置

### 1. 使用容器隔离

**最高安全级别**

```yaml
executor:
  execution_mode: container  # 强制容器模式

container:
  enabled: true
  network_mode: none         # 禁用网络
  auto_cleanup: true         # 自动清理
```

**安全效果**:
- ✅ 完全隔离
- ✅ 无法访问主机文件
- ✅ 无法访问网络
- ✅ 资源限制

---

### 2. 启用代码分析

**执行前检查**

```go
// 先分析
analysis, _ := module.AnalyzeCode(ctx, lang, code)

// 检查安全性
if !analysis.Safe {
    return errors.New("代码不安全")
}

// 检查评分
if analysis.Score < 70 {
    return errors.New("代码质量不达标")
}

// 再执行
result, _ := module.ExecuteCode(ctx, lang, code)
```

**配置**:
```yaml
analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  strict_mode: true
```

---

### 3. 限制资源使用

**防止资源滥用**

```yaml
executor:
  timeout: 30000        # 30 秒超时
  memory_limit: 256     # 256 MB 内存
  cpu_limit: 1          # 1 个 CPU 核心

container:
  cpu_limit: "0.25"     # 0.25 个 CPU 核心
  memory_limit: "256m"  # 256 MB 内存
  timeout: 30s          # 30 秒超时
```

---

### 4. 禁用网络访问

**防止数据泄露**

```yaml
container:
  network_mode: none  # 完全禁用网络
```

**验证**:
```python
# 此代码将失败
import socket
sock = socket.socket()
sock.connect(('example.com', 80))  # 连接失败
```

---

### 5. 文件系统隔离

**限制文件访问**

```yaml
container:
  # 只挂载必要的目录
  volumes:
    - /tmp:/tmp:ro  # 只读挂载
```

**避免**:
- 挂载敏感目录（/etc, /root）
- 可写挂载
- 主机路径访问

---

### 6. 使用非 root 用户

**降低权限**

```yaml
container:
  user: "1000:1000"  # 非 root 用户
```

**Dockerfile 示例**:
```dockerfile
FROM python:3.11-alpine
RUN adduser -D appuser
USER appuser
```

---

### 7. 自定义安全规则

**团队安全策略**

```yaml
# security_rules.yaml
rules:
  - name: "禁止 eval"
    language: "python"
    pattern: "eval\\("
    severity: "critical"
    
  - name: "禁止 exec"
    language: "python"
    pattern: "exec\\("
    severity: "critical"
    
  - name: "禁止 os.system"
    language: "python"
    pattern: "os\\.system\\("
    severity: "critical"
```

```yaml
analyzer:
  custom_rules_file: "security_rules.yaml"
  strict_mode: true
```

---

## 安全级别配置

### 级别 1: 开发环境（低安全）

```yaml
executor:
  timeout: 120000
  memory_limit: 1024
  cpu_limit: 4
  execution_mode: local

analyzer:
  strict_mode: false
  enable_network_detection: true
  enable_filesystem_detection: true

container:
  enabled: false
```

**适用**: 本地开发、测试

---

### 级别 2: 测试环境（中安全）

```yaml
executor:
  timeout: 60000
  memory_limit: 512
  cpu_limit: 2
  execution_mode: auto

analyzer:
  strict_mode: false
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true

container:
  enabled: true
  network_mode: bridge
  auto_cleanup: true
```

**适用**: 集成测试、预发布

---

### 级别 3: 生产环境（高安全）

```yaml
executor:
  timeout: 30000
  memory_limit: 512
  cpu_limit: 1
  execution_mode: container

analyzer:
  strict_mode: true
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  custom_rules_file: "production_rules.yaml"

container:
  enabled: true
  network_mode: none
  auto_cleanup: true
  cpu_limit: "0.5"
  memory_limit: "512m"
```

**适用**: 生产环境

---

### 级别 4: 高安全环境（最高安全）

```yaml
executor:
  timeout: 10000
  memory_limit: 256
  cpu_limit: 1
  execution_mode: container
  supported_languages:
    - python  # 只允许 Python

analyzer:
  strict_mode: true
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  custom_rules_file: "strict_rules.yaml"

container:
  enabled: true
  network_mode: none
  auto_cleanup: true
  cpu_limit: "0.25"
  memory_limit: "256m"
  timeout: 10s
  user: "1000:1000"
```

**适用**: 金融、医疗等高安全要求场景

---

## 安全检查清单

### 部署前检查

- [ ] 启用容器隔离
- [ ] 禁用网络访问
- [ ] 限制资源使用
- [ ] 启用代码分析
- [ ] 配置自定义规则
- [ ] 使用非 root 用户
- [ ] 启用自动清理
- [ ] 设置合理超时
- [ ] 限制支持的语言
- [ ] 启用严格模式

### 运行时监控

- [ ] 监控资源使用
- [ ] 记录执行日志
- [ ] 检测异常行为
- [ ] 定期安全审计
- [ ] 更新安全规则

---

## 常见安全问题

### 问题 1: 代码执行恶意操作

**症状**: 检测到危险操作

**解决**:
```go
analysis, _ := module.AnalyzeCode(ctx, lang, code)
if !analysis.Safe {
    log.Printf("检测到危险操作: %v", analysis.Issues)
    return errors.New("代码不安全")
}
```

---

### 问题 2: 资源被耗尽

**症状**: CPU/内存使用过高

**解决**:
```yaml
executor:
  timeout: 10000
  memory_limit: 256
  cpu_limit: 1
```

---

### 问题 3: 网络数据泄露

**症状**: 代码尝试网络访问

**解决**:
```yaml
container:
  network_mode: none
```

---

### 问题 4: 文件系统破坏

**症状**: 代码尝试删除文件

**解决**:
1. 使用容器隔离
2. 只读挂载
3. 启用文件系统检测

---

## 安全最佳实践

### 1. 纵深防御

```
层次 1: 代码分析（检测）
层次 2: 容器隔离（隔离）
层次 3: 资源限制（限制）
层次 4: 网络隔离（隔离）
层次 5: 监控告警（监控）
```

### 2. 最小权限原则

- 只授予必要的权限
- 使用非 root 用户
- 限制文件访问
- 禁用不需要的功能

### 3. 定期审计

```go
// 记录所有执行
log.Printf("执行代码: 用户=%s, 语言=%s, 长度=%d", 
    user, lang, len(code))

// 分析结果
log.Printf("安全评分: %d, 问题: %v", 
    analysis.Score, analysis.Issues)
```

### 4. 及时更新

- 更新 Docker 镜像
- 更新安全规则
- 更新依赖包
- 修复已知漏洞

---

## 安全事件响应

### 1. 检测

```go
// 监控异常
if !result.Success && result.ExitCode == 137 {
    alert("内存耗尽攻击")
}

if analysis.Score < 30 {
    alert("高危代码")
}
```

### 2. 响应

```go
// 阻止执行
if !analysis.Safe {
    return errors.New("代码被阻止")
}

// 记录日志
log.Printf("安全事件: %v", analysis.Issues)

// 通知管理员
notifyAdmin(analysis)
```

### 3. 恢复

```go
// 清理资源
module.Close()

// 重新初始化
module, _ = NewCodeExecutorModule(config)
```

---

## 合规性

### GDPR 合规

- 不记录敏感数据
- 提供数据删除功能
- 加密存储日志

### SOC 2 合规

- 访问控制
- 审计日志
- 加密传输

### ISO 27001 合规

- 风险评估
- 安全策略
- 持续监控

---

**版本**: 1.0  
**更新日期**: 2026-01-31

