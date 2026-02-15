# 自定义代码分析规则指南

本指南介绍如何创建和使用自定义代码分析规则。

## 配置格式

自定义规则使用 YAML 格式配置，基本结构如下：

```yaml
custom_rules:
  - name: "规则名称"
    language: "编程语言"
    pattern: "正则表达式模式"
    category: "类别"
    severity: "严重级别"
    description: "问题描述"
    suggestion: "修复建议"
```

## 字段说明

### 必填字段

- **name**: 规则的唯一标识符（字符串）
- **language**: 目标编程语言，支持：
  - `python`
  - `javascript`
  - `go`
  - `bash`
- **pattern**: 正则表达式模式，用于匹配代码
- **description**: 问题的详细描述

### 可选字段

- **category**: 规则类别，默认为 `custom`，支持：
  - `naming` - 命名规范
  - `style` - 代码风格
  - `performance` - 性能问题
  - `best_practice` - 最佳实践
  - `security` - 安全问题
  - `custom` - 自定义类别

- **severity**: 严重级别，默认为 `info`，支持：
  - `high` - 高危
  - `medium` - 中危
  - `low` - 低危
  - `info` - 信息

- **suggestion**: 修复建议或替代方案

## 正则表达式注意事项

1. 使用 Go 的 `regexp` 包语法
2. 不支持某些 Perl 特性（如负向前瞻 `(?!...)`）
3. 特殊字符需要转义：`\b` `\s` `\(` `\)` 等
4. 推荐使用原始字符串：`'\bpattern\b'`

## 示例规则

### Python 示例

```yaml
custom_rules:
  # 检测 print 语句
  - name: "python_print_statement"
    language: "python"
    pattern: '\bprint\s*\('
    category: "best_practice"
    severity: "info"
    description: "生产代码中避免使用 print 语句"
    suggestion: "使用 logging 模块代替 print"
  
  # 检测 TODO 注释
  - name: "python_todo_comment"
    language: "python"
    pattern: '#\s*TODO'
    category: "style"
    severity: "info"
    description: "发现 TODO 注释"
    suggestion: "完成或删除 TODO 注释"
  
  # 检测全局变量
  - name: "python_global_variable"
    language: "python"
    pattern: '\bglobal\s+\w+'
    category: "best_practice"
    severity: "low"
    description: "使用了全局变量"
    suggestion: "考虑使用类或函数参数传递"
```

### JavaScript 示例

```yaml
custom_rules:
  # 检测 debugger 语句
  - name: "js_debugger_statement"
    language: "javascript"
    pattern: '\bdebugger\b'
    category: "best_practice"
    severity: "medium"
    description: "生产代码中不应包含 debugger 语句"
    suggestion: "移除 debugger 语句"
  
  # 检测 alert 使用
  - name: "js_alert_usage"
    language: "javascript"
    pattern: '\balert\s*\('
    category: "best_practice"
    severity: "low"
    description: "避免使用 alert"
    suggestion: "使用更好的用户提示方式"
  
  # 检测 document.write
  - name: "js_document_write"
    language: "javascript"
    pattern: '\bdocument\.write\s*\('
    category: "best_practice"
    severity: "medium"
    description: "避免使用 document.write"
    suggestion: "使用 DOM 操作方法"
```

### Go 示例

```yaml
custom_rules:
  # 检测 fmt.Println
  - name: "go_fmt_println"
    language: "go"
    pattern: '\bfmt\.Println\s*\('
    category: "best_practice"
    severity: "info"
    description: "生产代码中避免使用 fmt.Println"
    suggestion: "使用日志库代替 fmt.Println"
  
  # 检测 panic 使用
  - name: "go_panic_usage"
    language: "go"
    pattern: '\bpanic\s*\('
    category: "best_practice"
    severity: "medium"
    description: "谨慎使用 panic"
    suggestion: "考虑返回错误而不是 panic"
  
  # 检测 time.Sleep
  - name: "go_time_sleep"
    language: "go"
    pattern: '\btime\.Sleep\s*\('
    category: "performance"
    severity: "info"
    description: "使用了 time.Sleep"
    suggestion: "考虑使用 context 或 channel 进行同步"
```

### Bash 示例

```yaml
custom_rules:
  # 检测 sudo 使用
  - name: "bash_sudo_usage"
    language: "bash"
    pattern: '\bsudo\s+'
    category: "security"
    severity: "medium"
    description: "使用了 sudo 命令"
    suggestion: "确保 sudo 使用是必要且安全的"
  
  # 检测未引用的变量
  - name: "bash_unquoted_variable"
    language: "bash"
    pattern: '\$[A-Za-z_][A-Za-z0-9_]*(?!["\}])'
    category: "best_practice"
    severity: "low"
    description: "变量未使用引号"
    suggestion: "使用引号包裹变量，如 \"$VAR\""
```

## 使用方法

### 1. 创建规则文件

创建一个 YAML 文件（如 `my_rules.yaml`），添加自定义规则。

### 2. 加载规则

```go
analyzer := NewCodeAnalyzer()
err := analyzer.LoadCustomRules("my_rules.yaml")
if err != nil {
    log.Fatal(err)
}
```

### 3. 分析代码

```go
result := analyzer.Analyze("python", code)

// 检查质量问题
for _, issue := range result.QualityIssues {
    fmt.Printf("Line %d: %s - %s\n", 
        issue.Line, 
        issue.Rule, 
        issue.Description)
}
```

### 4. 清除规则

```go
analyzer.ClearCustomRules()
```

## 最佳实践

1. **规则命名**: 使用描述性的名称，包含语言前缀
   - 好: `python_print_statement`
   - 差: `rule1`

2. **模式编写**: 
   - 使用 `\b` 匹配单词边界
   - 使用 `\s*` 匹配可选空格
   - 测试正则表达式以确保准确性

3. **严重级别**:
   - `high`: 安全漏洞、严重错误
   - `medium`: 可能导致问题的代码
   - `low`: 代码异味、次要问题
   - `info`: 建议性改进

4. **描述和建议**:
   - 描述应清晰说明问题
   - 建议应提供可操作的解决方案

5. **测试规则**:
   - 在实际代码上测试规则
   - 检查误报和漏报
   - 根据需要调整模式

## 常见问题

### Q: 如何匹配多行代码？
A: 当前实现按行匹配，复杂的多行模式可能需要在应用层处理。

### Q: 可以使用哪些正则表达式特性？
A: 支持 Go `regexp` 包的所有特性，但不支持某些 Perl 特性（如负向前瞻）。

### Q: 如何调试正则表达式？
A: 使用在线工具（如 regex101.com）测试，选择 Go 语言模式。

### Q: 规则会影响性能吗？
A: 每个规则都会增加分析时间，建议只添加必要的规则。

### Q: 可以覆盖内置规则吗？
A: 自定义规则是附加的，不会覆盖内置规则。

## 示例文件

完整的示例文件请参考：`custom_rules_example.yaml`

## 技术支持

如有问题或建议，请查看项目文档或提交 issue。
