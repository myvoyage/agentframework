# MCP 工具 API 文档

## 概述

代码执行模块提供了一套 MCP (Model Context Protocol) 工具，允许 AI 模型通过标准化接口执行代码、分析代码和管理执行环境。

本文档详细说明所有可用的 MCP 工具及其使用方法。

---

## 工具列表

| 工具名称 | 功能 | 状态 |
|---------|------|------|
| `code_exec_run` | 执行代码 | ✅ 可用 |
| `code_exec_analyze` | 分析代码 | ✅ 增强版 |
| `code_exec_format` | 格式化代码 | ✅ 可用 |
| `code_exec_supported_languages` | 查询支持的语言 | ✅ 可用 |
| `code_exec_set_mode` | 设置执行模式 | ✅ 新增 |
| `code_exec_container_status` | 查询容器状态 | ✅ 新增 |

---

## 1. code_exec_run

### 功能描述

执行指定语言的代码并返回执行结果。

### 参数

```json
{
  "language": "string",  // 必需：编程语言
  "code": "string"       // 必需：要执行的代码
}
```

#### language
- **类型**: string
- **必需**: 是
- **可选值**: "python", "javascript", "js", "go", "bash", "sh"
- **说明**: 代码的编程语言

#### code
- **类型**: string
- **必需**: 是
- **说明**: 要执行的代码内容

### 返回值

```json
{
  "success": true,
  "output": "执行输出",
  "error": "",
  "execution_time_ms": 123,
  "memory_used_mb": 45,
  "exit_code": 0
}
```

#### 返回字段说明

- `success`: 执行是否成功
- `output`: 标准输出内容
- `error`: 错误信息（如果有）
- `execution_time_ms`: 执行时间（毫秒）
- `memory_used_mb`: 内存使用（MB）
- `exit_code`: 退出码（0 表示成功）

### 使用示例

#### Python 代码执行

```json
{
  "language": "python",
  "code": "print('Hello, World!')\nprint(2 + 2)"
}
```

**返回**:
```json
{
  "success": true,
  "output": "Hello, World!\n4\n",
  "error": "",
  "execution_time_ms": 45,
  "memory_used_mb": 12,
  "exit_code": 0
}
```

#### JavaScript 代码执行

```json
{
  "language": "javascript",
  "code": "console.log('Hello, World!');\nconsole.log(2 + 2);"
}
```

**返回**:
```json
{
  "success": true,
  "output": "Hello, World!\n4\n",
  "error": "",
  "execution_time_ms": 38,
  "memory_used_mb": 15,
  "exit_code": 0
}
```

#### Go 代码执行

```json
{
  "language": "go",
  "code": "package main\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n    fmt.Println(2 + 2)\n}"
}
```

**返回**:
```json
{
  "success": true,
  "output": "Hello, World!\n4\n",
  "error": "",
  "execution_time_ms": 156,
  "memory_used_mb": 8,
  "exit_code": 0
}
```

### 执行模式

代码执行支持三种模式：

1. **local** - 本地执行
   - 最快速度
   - 适合开发环境
   - Go 代码使用 Yaegi 解释器（428x 性能提升）

2. **container** - 容器执行
   - 最高安全性
   - 完全隔离
   - 适合生产环境

3. **auto** - 自动选择（推荐）
   - 根据代码特征自动选择
   - 平衡性能和安全

### 错误处理

#### 语法错误

```json
{
  "language": "python",
  "code": "print('Hello"
}
```

**返回**:
```json
{
  "success": false,
  "output": "",
  "error": "SyntaxError: unterminated string literal",
  "execution_time_ms": 12,
  "memory_used_mb": 0,
  "exit_code": 1
}
```

#### 运行时错误

```json
{
  "language": "python",
  "code": "x = 1 / 0"
}
```

**返回**:
```json
{
  "success": false,
  "output": "",
  "error": "ZeroDivisionError: division by zero",
  "execution_time_ms": 23,
  "memory_used_mb": 8,
  "exit_code": 1
}
```

#### 超时错误

```json
{
  "language": "python",
  "code": "import time\ntime.sleep(100)"
}
```

**返回**:
```json
{
  "success": false,
  "output": "",
  "error": "execution timeout after 60000ms",
  "execution_time_ms": 60000,
  "memory_used_mb": 10,
  "exit_code": -1
}
```

---

## 2. code_exec_analyze

### 功能描述

分析代码的安全性、质量和潜在问题。增强版支持多种检测类型和代码质量评分。

### 参数

```json
{
  "language": "string",  // 必需：编程语言
  "code": "string"       // 必需：要分析的代码
}
```

### 返回值

```json
{
  "success": true,
  "safe": false,
  "issues": ["issue1", "issue2"],
  "dangerous_operations": ["op1", "op2"],
  "network_ops": [...],
  "filesystem_ops": [...],
  "process_ops": [...],
  "crypto_issues": [...],
  "database_ops": [...],
  "quality_issues": [...],
  "score": 75,
  "suggestions": ["suggestion1", "suggestion2"]
}
```

#### 返回字段说明

- `success`: 分析是否成功
- `safe`: 代码是否安全
- `issues`: 发现的问题列表
- `dangerous_operations`: 危险操作列表
- `network_ops`: 网络操作详情
- `filesystem_ops`: 文件系统操作详情
- `process_ops`: 进程操作详情
- `crypto_issues`: 加密问题详情
- `database_ops`: 数据库操作详情
- `quality_issues`: 代码质量问题
- `score`: 代码质量评分（0-100）
- `suggestions`: 改进建议

### 使用示例

#### 分析网络操作

```json
{
  "language": "python",
  "code": "import requests\nresponse = requests.get('http://example.com')\nprint(response.text)"
}
```

**返回**:
```json
{
  "success": true,
  "safe": false,
  "issues": ["Network operation detected: HTTP request"],
  "dangerous_operations": ["requests.get"],
  "network_ops": [
    {
      "type": "http_request",
      "line": 2,
      "column": 11,
      "code": "response = requests.get('http://example.com')",
      "severity": "medium",
      "message": "HTTP request to external URL",
      "suggestion": "Validate URL and use HTTPS"
    }
  ],
  "filesystem_ops": [],
  "process_ops": [],
  "crypto_issues": [],
  "database_ops": [],
  "quality_issues": [],
  "score": 65,
  "suggestions": [
    "Use HTTPS instead of HTTP",
    "Add error handling for network requests"
  ]
}
```

#### 分析文件系统操作

```json
{
  "language": "python",
  "code": "import os\nos.remove('/etc/passwd')"
}
```

**返回**:
```json
{
  "success": true,
  "safe": false,
  "issues": ["File system operation detected: file deletion"],
  "dangerous_operations": ["os.remove"],
  "network_ops": [],
  "filesystem_ops": [
    {
      "type": "file_delete",
      "line": 2,
      "column": 0,
      "code": "os.remove('/etc/passwd')",
      "severity": "critical",
      "message": "Attempting to delete sensitive system file",
      "suggestion": "Never delete system files"
    }
  ],
  "process_ops": [],
  "crypto_issues": [],
  "database_ops": [],
  "quality_issues": [],
  "score": 20,
  "suggestions": [
    "Remove file deletion operations",
    "Use safe file paths"
  ]
}
```

#### 分析代码质量

```json
{
  "language": "python",
  "code": "def f(x):\n  if x>0:\n   return x*2\n  else:\n   return x*3"
}
```

**返回**:
```json
{
  "success": true,
  "safe": true,
  "issues": [],
  "dangerous_operations": [],
  "network_ops": [],
  "filesystem_ops": [],
  "process_ops": [],
  "crypto_issues": [],
  "database_ops": [],
  "quality_issues": [
    {
      "type": "naming",
      "line": 1,
      "column": 4,
      "code": "def f(x):",
      "severity": "low",
      "message": "Function name 'f' is not descriptive",
      "suggestion": "Use descriptive function names"
    },
    {
      "type": "style",
      "line": 2,
      "column": 2,
      "code": "  if x>0:",
      "severity": "low",
      "message": "Missing spaces around operator",
      "suggestion": "Use 'x > 0' instead of 'x>0'"
    },
    {
      "type": "style",
      "line": 3,
      "column": 3,
      "code": "   return x*2",
      "severity": "low",
      "message": "Inconsistent indentation",
      "suggestion": "Use consistent indentation (2 or 4 spaces)"
    }
  ],
  "score": 70,
  "suggestions": [
    "Use descriptive names",
    "Fix indentation",
    "Add spaces around operators"
  ]
}
```

### 检测类型

#### 1. 网络操作检测

检测以下操作：
- HTTP/HTTPS 请求
- Socket 连接
- DNS 查询
- WebSocket 连接

#### 2. 文件系统操作检测

检测以下操作：
- 文件读取/写入/删除
- 目录操作
- 权限修改
- 敏感路径访问

#### 3. 进程操作检测

检测以下操作：
- 进程创建
- 命令执行
- 信号发送
- IPC 通信

#### 4. 加密问题检测

检测以下问题：
- 弱加密算法
- 硬编码密钥
- 不安全的随机数
- 明文密码

#### 5. 数据库操作检测

检测以下操作：
- SQL 查询
- SQL 注入风险
- 连接泄漏
- 不安全的查询

#### 6. 代码质量检查

检查以下方面：
- 命名规范
- 代码风格
- 最佳实践
- 性能问题

### 评分系统

代码质量评分（0-100）：

- **90-100**: 优秀 - 无问题或仅有轻微问题
- **70-89**: 良好 - 有一些小问题
- **50-69**: 一般 - 有明显问题需要修复
- **30-49**: 较差 - 有严重问题
- **0-29**: 很差 - 有关键问题

评分计算：
```
基础分: 100
每个 critical 问题: -20 分
每个 high 问题: -10 分
每个 medium 问题: -5 分
每个 low 问题: -2 分
最低分: 0
```

---

## 3. code_exec_format

### 功能描述

格式化代码，使其符合标准代码风格。

### 参数

```json
{
  "language": "string",  // 必需：编程语言
  "code": "string"       // 必需：要格式化的代码
}
```

### 返回值

```json
{
  "success": true,
  "formatted_code": "格式化后的代码",
  "changes": ["change1", "change2"]
}
```

### 使用示例

#### Python 代码格式化

```json
{
  "language": "python",
  "code": "def f(x):\n  if x>0:\n   return x*2\n  else:\n   return x*3"
}
```

**返回**:
```json
{
  "success": true,
  "formatted_code": "def f(x):\n    if x > 0:\n        return x * 2\n    else:\n        return x * 3\n",
  "changes": [
    "Fixed indentation",
    "Added spaces around operators"
  ]
}
```

#### Go 代码格式化

```json
{
  "language": "go",
  "code": "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"Hello\")}"
}
```

**返回**:
```json
{
  "success": true,
  "formatted_code": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n",
  "changes": [
    "Added blank lines",
    "Fixed function formatting"
  ]
}
```

---

## 4. code_exec_supported_languages

### 功能描述

查询当前支持的编程语言列表。

### 参数

无参数。

### 返回值

```json
{
  "success": true,
  "languages": ["python", "javascript", "go", "bash"]
}
```

### 使用示例

```json
{}
```

**返回**:
```json
{
  "success": true,
  "languages": ["python", "javascript", "go", "bash"]
}
```

---

## 5. code_exec_set_mode

### 功能描述

设置代码执行模式（本地/容器/自动）。

### 参数

```json
{
  "mode": "string"  // 必需：执行模式
}
```

#### mode
- **类型**: string
- **必需**: 是
- **可选值**: "local", "container", "auto"
- **说明**: 
  - `local`: 本地执行（快速）
  - `container`: 容器执行（安全）
  - `auto`: 自动选择（推荐）

### 返回值

```json
{
  "success": true,
  "mode": "container",
  "message": "Execution mode set to: container"
}
```

### 使用示例

#### 设置为容器模式

```json
{
  "mode": "container"
}
```

**返回**:
```json
{
  "success": true,
  "mode": "container",
  "message": "Execution mode set to: container"
}
```

#### 设置为自动模式

```json
{
  "mode": "auto"
}
```

**返回**:
```json
{
  "success": true,
  "mode": "auto",
  "message": "Execution mode set to: auto"
}
```

#### 无效模式

```json
{
  "mode": "invalid"
}
```

**返回**:
```json
{
  "success": false,
  "error": "invalid mode: invalid (must be 'local', 'container', or 'auto')"
}
```

### 使用场景

1. **开发环境**: 使用 `local` 模式获得最快速度
2. **生产环境**: 使用 `container` 模式获得最高安全性
3. **混合环境**: 使用 `auto` 模式自动选择

---

## 6. code_exec_container_status

### 功能描述

查询 Docker 容器执行器的状态和统计信息。

### 参数

无参数。

### 返回值

```json
{
  "success": true,
  "enabled": true,
  "stats": {
    "total_executions": 150,
    "success_count": 145,
    "failure_count": 5,
    "total_duration_ms": 45000,
    "active_containers": 3
  },
  "pool_stats": {
    "size": 5,
    "available": 2,
    "in_use": 3,
    "created": 10,
    "reused": 140
  },
  "containers": [
    {
      "id": "abc123",
      "language": "python",
      "status": "running",
      "created_at": "2026-01-31 10:30:00",
      "exit_code": 0
    }
  ]
}
```

#### 返回字段说明

**stats** - 执行统计:
- `total_executions`: 总执行次数
- `success_count`: 成功次数
- `failure_count`: 失败次数
- `total_duration_ms`: 总执行时间（毫秒）
- `active_containers`: 活动容器数

**pool_stats** - 容器池统计（如果启用）:
- `size`: 池大小
- `available`: 可用容器数
- `in_use`: 使用中容器数
- `created`: 创建的容器总数
- `reused`: 复用次数

**containers** - 活动容器列表:
- `id`: 容器 ID
- `language`: 语言
- `status`: 状态
- `created_at`: 创建时间
- `exit_code`: 退出码

### 使用示例

#### 容器已启用

```json
{}
```

**返回**:
```json
{
  "success": true,
  "enabled": true,
  "stats": {
    "total_executions": 150,
    "success_count": 145,
    "failure_count": 5,
    "total_duration_ms": 45000,
    "active_containers": 3
  },
  "pool_stats": {
    "size": 5,
    "available": 2,
    "in_use": 3,
    "created": 10,
    "reused": 140
  },
  "containers": [
    {
      "id": "abc123",
      "language": "python",
      "status": "running",
      "created_at": "2026-01-31 10:30:00",
      "exit_code": 0
    },
    {
      "id": "def456",
      "language": "javascript",
      "status": "running",
      "created_at": "2026-01-31 10:31:00",
      "exit_code": 0
    }
  ]
}
```

#### 容器未启用

```json
{}
```

**返回**:
```json
{
  "success": true,
  "enabled": false,
  "message": "Container executor not initialized"
}
```

### 使用场景

1. **监控**: 定期查询容器状态
2. **调试**: 检查容器是否正常运行
3. **优化**: 根据统计信息调整配置
4. **故障排查**: 查看失败次数和活动容器

---

## 工具集成示例

### 完整工作流

```python
# 1. 查询支持的语言
languages = call_tool("code_exec_supported_languages", {})
print(f"支持的语言: {languages['languages']}")

# 2. 设置执行模式
mode_result = call_tool("code_exec_set_mode", {
    "mode": "auto"
})
print(f"执行模式: {mode_result['mode']}")

# 3. 分析代码
code = """
import requests
response = requests.get('http://example.com')
print(response.text)
"""

analysis = call_tool("code_exec_analyze", {
    "language": "python",
    "code": code
})

if not analysis['safe']:
    print(f"代码不安全: {analysis['issues']}")
    print(f"建议: {analysis['suggestions']}")
else:
    # 4. 执行代码
    result = call_tool("code_exec_run", {
        "language": "python",
        "code": code
    })
    
    if result['success']:
        print(f"输出: {result['output']}")
        print(f"执行时间: {result['execution_time_ms']}ms")
    else:
        print(f"执行失败: {result['error']}")

# 5. 查询容器状态
status = call_tool("code_exec_container_status", {})
if status['enabled']:
    print(f"容器统计: {status['stats']}")
```

### 代码审查工作流

```python
def review_code(language, code):
    # 1. 分析代码
    analysis = call_tool("code_exec_analyze", {
        "language": language,
        "code": code
    })
    
    # 2. 格式化代码
    formatted = call_tool("code_exec_format", {
        "language": language,
        "code": code
    })
    
    # 3. 生成报告
    report = {
        "safe": analysis['safe'],
        "score": analysis['score'],
        "issues": analysis['issues'],
        "suggestions": analysis['suggestions'],
        "formatted_code": formatted['formatted_code']
    }
    
    return report
```

### 安全执行工作流

```python
def safe_execute(language, code):
    # 1. 设置容器模式
    call_tool("code_exec_set_mode", {"mode": "container"})
    
    # 2. 分析代码
    analysis = call_tool("code_exec_analyze", {
        "language": language,
        "code": code
    })
    
    # 3. 检查安全性
    if not analysis['safe']:
        return {
            "success": False,
            "error": "Code is not safe to execute",
            "issues": analysis['issues']
        }
    
    # 4. 执行代码
    result = call_tool("code_exec_run", {
        "language": language,
        "code": code
    })
    
    return result
```

---

## 最佳实践

### 1. 始终先分析后执行

```python
# ✅ 好的做法
analysis = analyze_code(code)
if analysis['safe']:
    result = execute_code(code)

# ❌ 不好的做法
result = execute_code(code)  # 未经分析直接执行
```

### 2. 根据环境选择模式

```python
# 开发环境
set_mode("local")  # 快速迭代

# 生产环境
set_mode("container")  # 安全隔离

# 混合环境
set_mode("auto")  # 自动选择
```

### 3. 监控容器状态

```python
# 定期检查
status = get_container_status()
if status['stats']['failure_count'] > 10:
    alert("容器执行失败率过高")
```

### 4. 处理错误

```python
try:
    result = execute_code(code)
    if not result['success']:
        log_error(result['error'])
        retry_with_different_mode()
except Exception as e:
    log_exception(e)
    fallback_to_safe_mode()
```

---

## 性能优化

### 1. 使用 Yaegi 加速 Go 代码

```python
# Go 代码在 local 模式下使用 Yaegi
# 性能提升 428 倍
set_mode("local")
result = execute_code("go", go_code)
```

### 2. 启用容器池

```yaml
container:
  enable_pool: true
  pool_min_size: 5
  pool_max_size: 20
```

性能提升：
- 减少 80% 容器启动时间
- 提高 3-5 倍吞吐量

### 3. 启用编译缓存

```yaml
yaegi:
  enable_cache: true
  cache_capacity: 500
```

性能提升：
- 缓存命中时提升 12,600 倍
- 减少 CPU 使用

---

## 故障排查

### 执行失败

**问题**: 代码执行失败

**检查**:
1. 查看错误信息
2. 检查语法是否正确
3. 验证执行模式是否合适
4. 查看容器状态

```python
result = execute_code(code)
if not result['success']:
    print(f"错误: {result['error']}")
    print(f"退出码: {result['exit_code']}")
```

### 容器不可用

**问题**: 容器执行器未启用

**检查**:
1. Docker 是否安装
2. Docker 服务是否运行
3. 配置是否正确

```python
status = get_container_status()
if not status['enabled']:
    print(f"容器未启用: {status['message']}")
    # 切换到本地模式
    set_mode("local")
```

### 性能问题

**问题**: 执行速度慢

**优化**:
1. 启用 Yaegi（Go 代码）
2. 启用容器池
3. 启用编译缓存
4. 使用 auto 模式

```python
# 查看执行统计
status = get_container_status()
avg_time = status['stats']['total_duration_ms'] / status['stats']['total_executions']
print(f"平均执行时间: {avg_time}ms")
```

---

## 安全注意事项

### 1. 始终分析代码

```python
# 执行前先分析
analysis = analyze_code(code)
if not analysis['safe']:
    reject_execution()
```

### 2. 使用容器隔离

```python
# 生产环境使用容器
set_mode("container")
```

### 3. 限制资源使用

```yaml
executor:
  timeout: 30000      # 30 秒超时
  memory_limit: 512   # 512 MB 内存
  cpu_limit: 1        # 1 个 CPU 核心
```

### 4. 禁用网络访问

```yaml
container:
  network_mode: none  # 无网络访问
```

---

## 参考资料

- [增强代码分析 API](./ENHANCED_CODE_ANALYSIS_API.md)
- [Yaegi 集成 API](./YAEGI_INTEGRATION_API.md)
- [容器执行 API](./CONTAINER_EXECUTION_API.md)
- [配置指南](./CONFIGURATION_GUIDE.md)

---

**版本**: 1.0  
**更新日期**: 2026-01-31  
**作者**: Agent Framework Team

