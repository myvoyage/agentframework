---
id: "com.agentframework.skill.comprehensive_demo"
name: "comprehensive_demo"
version: "1.0.0"
category: "example"
tags:
  - demo
  - comprehensive
  - tutorial
  - examples
description: "全面演示所有执行器功能的综合示例"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins: []
  env: []

triggers:
  - type: "command"
    pattern: "/demo"
    priority: 10

actions:
  # 1. Template 执行器示例 - 条件语句
  - id: "template_conditionals"
    type: "template"
    description: "演示模板条件语句"
    config:
      template: |
        用户信息：
        姓名：{{.Name}}
        状态：{{if eq .Status "active"}}激活{{else}}未激活{{end}}
        级别：{{if eq .Level "admin"}}管理员{{else if eq .Level "user"}}普通用户{{else}}访客{{end}}
        {{if gt .Score 80}}
        ✅ 优秀表现！
        {{else if gt .Score 60}}
        ✓ 及格
        {{else}}
        ❌ 需要改进
        {{end}}
    timeout: "5s"

  # 2. Template 执行器示例 - 循环
  - id: "template_loop"
    type: "template"
    description: "演示模板循环"
    config:
      template: |
        待办事项列表：
        {{range .Items}}
        - [{{.Index}}] {{.Value}}
        {{end}}

        统计：共 {{.Count}} 项
    timeout: "5s"

  # 3. File 执行器示例 - 写入配置文件
  - id: "file_write_config"
    type: "file"
    description: "写入配置文件"
    config:
      operation: "write"
      path: "./output/{{.Filename}}.json"
      content: '{"name":"{{.Name}}","version":"{{.Version}}","enabled":{{.Enabled}}}'
      create_dirs: true
      overwrite: true
    timeout: "10s"

  # 4. File 执行器示例 - 读取文件
  - id: "file_read"
    type: "file"
    description: "读取文件内容"
    config:
      operation: "read"
      path: "./output/{{.Filename}}.txt"
    timeout: "5s"

  # 5. File 执行器示例 - 列出目录
  - id: "file_list"
    type: "file"
    description: "列出目录内容"
    config:
      operation: "list"
      path: "./output"
      recursive: false
      show_hidden: false
    timeout: "10s"

  # 6. JSON 执行器示例 - 解析和格式化
  - id: "json_format"
    type: "json"
    description: "格式化 JSON"
    config:
      operation: "format"
      input: '{"name":"{{.Name}}","age":{{.Age}},"active":{{.Active}}}'
    timeout: "5s"

  # 7. JSON 执行器示例 - 提取字段
  - id: "json_extract"
    type: "json"
    description: "提取 JSON 字段"
    config:
      operation: "extract"
      input: '{"user":{"name":"Alice","age":30},"status":"active"}'
      field: "user.name"
    timeout: "5s"

  # 8. JSON 执行器示例 - JSONPath 查询
  - id: "json_query"
    type: "json"
    description: "JSONPath 查询数组元素"
    config:
      operation: "query"
      input: '{"users":[{"name":"Alice"},{"name":"Bob"},{"name":"Charlie"}]}'
      path: "users[0].name"
    timeout: "5s"

  # 9. JSON 执行器示例 - 过滤数组
  - id: "json_filter"
    type: "json"
    description: "过滤 JSON 数组"
    config:
      operation: "filter"
      input: '[{"name":"Alice","age":30},{"name":"Bob","age":25},{"name":"Charlie","age":35}]'
      condition: "age>30"
    timeout: "5s"

  # 10. JSON 执行器示例 - 深度合并
  - id: "json_deep_merge"
    type: "json"
    description: "深度合并 JSON 对象"
    config:
      operation: "deep_merge"
      base: '{"config":{"debug":false,"level":"info"}}'
      merge: '{"config":{"level":"verbose","timeout":30}}'
    timeout: "5s"

  # 11. JSON 执行器示例 - 数组排序
  - id: "json_sort"
    type: "json"
    description: "对 JSON 数组排序"
    config:
      operation: "sort"
      input: '["Charlie","Alice","Bob"]'
      order: "asc"
    timeout: "5s"

  # 12. 综合工作流 - 生成报告
  - id: "workflow_report"
    type: "workflow"
    description: "生成综合报告"
    config:
      steps:
        # 步骤1: 使用模板生成报告头部
        - id: "header"
          type: "template"
          config:
            template: |
              ============================================
              项目报告
              ============================================
              项目名称：{{.Project}}
              生成时间：{{.Timestamp}}
              --------------------------------------------
        # 步骤2: 使用 JSON 处理数据
        - id: "process_data"
          type: "json"
          config:
            operation: "format"
            input: '{"metrics":{"requests":{{.Requests}},"errors":{{.Errors}}},"uptime":"{{.Uptime}}"}'
        # 步骤3: 使用模板格式化最终输出
        - id: "format_output"
          type: "template"
          config:
            template: |
              {{.header}}

              数据统计：
              {{.process_data}}

              --------------------------------------------
              报告结束
    timeout: "30s"

  # 13. 文件操作工作流 - 备份配置
  - id: "workflow_backup"
    type: "workflow"
    description: "配置文件备份工作流"
    config:
      steps:
        # 步骤1: 检查源文件是否存在
        - id: "check_source"
          type: "file"
          config:
            operation: "exists"
            path: "./config/{{.ConfigFile}}"
        # 步骤2: 读取配置文件
        - id: "read_config"
          type: "file"
          config:
            operation: "read"
            path: "./config/{{.ConfigFile}}"
        # 步骤3: 验证 JSON 格式
        - id: "validate_json"
          type: "json"
          config:
            operation: "parse"
            input: "{{.read_config}}"
        # 步骤4: 写入备份文件
        - id: "write_backup"
          type: "file"
          config:
            operation: "write"
            path: "./backup/{{.ConfigFile}}.{{.Date}}.bak"
            content: "{{.read_config}}"
            create_dirs: true
    timeout: "30s"

  # 14. 数据处理工作流
  - id: "workflow_data_processing"
    type: "workflow"
    description: "数据处理管道"
    config:
      steps:
        # 步骤1: 读取原始数据
        - id: "read_data"
          type: "file"
          config:
            operation: "read"
            path: "./data/{{.DataFile}}"
        # 步骤2: 解析 JSON
        - id: "parse_json"
          type: "json"
          config:
            operation: "parse"
            input: "{{.read_data}}"
        # 步骤3: 过滤数据
        - id: "filter_data"
          type: "json"
          config:
            operation: "filter"
            input: "{{.parse_json}}"
            condition: "{{.FilterCondition}}"
        # 步骤4: 排序结果
        - id: "sort_data"
          type: "json"
          config:
            operation: "sort"
            input: "{{.filter_data}}"
            order: "asc"
        # 步骤5: 保存处理后的数据
        - id: "save_result"
          type: "file"
          config:
            operation: "write"
            path: "./output/{{.OutputFile}}"
            content: "{{.sort_data}}"
            create_dirs: true
    timeout: "60s"

config:
  max_output_size: 10485760
  max_execution_time: "120s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# 综合执行器演示

这个技能全面演示了 Agent Framework 所有执行器的功能。

## 执行器概览

### 1. Template 执行器
强大的模板引擎，支持变量替换、条件语句和循环。

**功能：**
- 变量替换：`{{.Variable}}`
- 条件语句：if/else/eq/ne/gt/lt/and/or/not
- 循环：range with index
- 嵌套模板

**使用示例：**
```bash
# 条件语句
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo template_conditionals --vars "Name=Alice,Status=active,Level=admin,Score=85"

# 循环
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo template_loop --vars "Items=task1,task2,task3,Count=3"
```

### 2. File 执行器
完整的文件操作功能，包含安全检查。

**操作类型：**
- `read` - 读取文件
- `write` - 写入文件
- `append` - 追加内容
- `delete` - 删除文件/目录
- `list` - 列出目录
- `exists` - 检查文件存在
- `mkdir` - 创建目录
- `copy` - 复制文件
- `move` - 移动文件
- `chmod` - 更改权限

**安全特性：**
- 路径遍历检测
- 文件大小限制
- 自动备份
- 权限验证

**使用示例：**
```bash
# 写入配置
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo file_write_config --vars "Filename=config,Name=MyApp,Version=1.0.0,Enabled=true"

# 读取文件
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo file_read --vars "Filename=config"

# 列出目录
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo file_list --vars ""
```

### 3. JSON 执行器
全面的 JSON 处理能力。

**操作类型：**
- `parse` - 解析和格式化
- `extract` - 提取字段
- `query` - JSONPath 查询
- `filter` - 过滤数组
- `map` - 映射转换
- `reduce` - 归约操作
- `sort` - 排序
- `unique` - 去重
- `merge` - 浅合并
- `deep_merge` - 深度合并
- `keys` / `values` - 键值提取
- `invert` - 键值反转
- `flatten` - 扁平化
- `uppercase` / `lowercase` - 大小写转换
- `validate` - 验证
- `select` - 字段选择
- `compare` - 比较
- `count` - 计数

**JSONPath 支持：**
- 点表示法：`user.name`
- 数组索引：`users[0]`
- 通配符：`*`
- 组合路径：`users[0].name`

**使用示例：**
```bash
# 格式化 JSON
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_format --vars "Name=Alice,Age=30,Active=true"

# 提取字段
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_extract --vars ""

# JSONPath 查询
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_query --vars ""

# 过滤数组
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_filter --vars ""

# 深度合并
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_deep_merge --vars ""

# 排序
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_sort --vars ""
```

### 4. Workflow 执行器
组合多个执行器创建复杂的工作流。

**特性：**
- 多步骤编排
- 步骤间数据传递
- 错误处理
- 超时控制

**使用示例：**
```bash
# 生成报告
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_report --vars "Project=MyProject,Timestamp=2024-01-15,Requests=1000,Errors=5,Uptime=99.9%"

# 备份配置
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_backup --vars "ConfigFile=config.json,Date=20240115"

# 数据处理管道
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_data_processing --vars "DataFile=input.json,FilterCondition=age>18,OutputFile=output.json"
```

### 5. Shell 执行器
执行 Shell 命令（带安全检查）。

### 6. HTTP 执行器
发送 HTTP 请求。

## 实际应用场景

### 场景1：配置文件管理

```bash
# 1. 生成配置
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo file_write_config --vars "Filename=production,Name=API,Version=2.0,Enabled=true"

# 2. 验证 JSON
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo json_format --vars ""

# 3. 备份配置
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_backup --vars "ConfigFile=production.json,Date=$(date +%Y%m%d)"
```

### 场景2：数据处理管道

```bash
# 完整的数据处理流程
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_data_processing \
  --vars "DataFile=users.json,FilterCondition=status=active,OutputFile=active_users.json"
```

### 场景3：报告生成

```bash
# 生成项目报告
agentframework enhanced-skill execute com.agentframework.skill.comprehensive_demo workflow_report \
  --vars "Project=E-Commerce,Timestamp=2024-01-15 14:30,Requests=50000,Errors=12,Uptime=99.95%"
```

## 高级技巧

### 1. 嵌套模板

```yaml
template: |
  外层循环：
  {{range .Categories}}
  分类：{{.Value}}
  {{range .Items}}
    - 项目：{{.Value}}
  {{end}}
  {{end}}
```

### 2. 复杂条件

```yaml
template: |
  {{if and (eq .Status "active") (gt .Score 80)}}
  高级活跃用户
  {{else if or (eq .Status "inactive") (lt .Score 60)}}
  需要关注
  {{else}}
  正常用户
  {{end}}
```

### 3. JSONPath 深度查询

```yaml
path: "data.users[0].profile.settings.theme"
```

### 4. 工作流错误处理

工作流会在任何步骤失败时停止并返回错误。使用 `timeout` 设置每个步骤的超时。

## 最佳实践

### 1. 文件操作
- 始终使用相对路径
- 设置合理的文件大小限制
- 定期创建备份
- 使用 `exists` 检查文件存在

### 2. JSON 处理
- 先验证 JSON 格式
- 使用 `query` 而非 `extract` 处理复杂路径
- 使用 `deep_merge` 而非 `merge` 保留嵌套结构

### 3. 模板
- 保持模板简洁
- 使用条件语句减少分支
- 使用循环处理重复内容

### 4. 工作流
- 每个步骤职责单一
- 合理设置超时
- 步骤间使用变量传递数据

## 性能建议

1. **文件操作**：大文件使用流式处理
2. **JSON 处理**：避免深度嵌套
3. **模板**：预编译常用模板
4. **工作流**：并行执行独立步骤

## 故障排除

### 文件权限错误
```bash
# 检查文件权限
ls -la ./output

# 修改权限
chmod 755 ./output
```

### JSON 解析错误
```bash
# 验证 JSON 格式
cat file.json | jq .
```

### 工作流超时
```bash
# 增加超时时间
# 在 config 中设置更长的 timeout
```

## 相关文档

- [增强技能定义参考](../../agent/skills/README.md)
- [技能开发指南](../../docs/SKILL_DEVELOPMENT.md)
- [CLI 技能使用指南](../../docs/CLI_SKILLS.md)
