# 增强代码分析 API 文档

## 概述

增强代码分析模块提供了全面的代码安全和质量检查功能，支持多种编程语言，能够检测危险操作、代码质量问题和最佳实践违规。

## 核心功能

### 1. 危险操作检测
- **网络操作**: HTTP/HTTPS 请求、Socket 连接、DNS 查询
- **文件系统操作**: 文件读写、删除、权限修改、敏感目录访问
- **进程管理**: 进程创建、信号发送、IPC 通信
- **加密操作**: 弱加密算法、硬编码密钥、不安全随机数
- **数据库操作**: SQL 注入、连接泄漏

### 2. 代码质量检查
- **命名规范**: 变量、函数、类命名检查
- **代码风格**: 缩进、行长度、空行检查
- **最佳实践**: 错误处理、资源清理、并发安全
- **性能问题**: 循环优化、类型转换、内存泄漏

### 3. 自定义规则
- 支持 YAML 格式的自定义规则
- 灵活的模式匹配
- 可配置的严重级别

---

## API 参考

### CodeAnalyzer

#### 创建分析器

```go
// 创建默认分析器
analyzer := NewCodeAnalyzer()

// 加载自定义规则
err := analyzer.LoadCustomRules("custom_rules.yaml")
if err != nil {
    log.Fatal(err)
}
```

#### Analyze 方法

分析代码并返回详细的分析结果。

**签名**:
```go
func (ca *CodeAnalyzer) Analyze(language, code string) *AnalysisResult
```

**参数**:
- `language` (string): 编程语言 ("python", "javascript", "go", "bash")
- `code` (string): 要分析的代码

**返回值**: `*AnalysisResult`

**示例**:
```go
code := `
import requests
response = requests.get('http://example.com')
print(response.text)
`

result := analyzer.Analyze("python", code)

if !result.Safe {
    fmt.Println("代码不安全!")
    for _, issue := range result.Issues {
        fmt.Printf("- %s\n", issue)
    }
}

// 检查网络操作
for _, netOp := range result.NetworkOps {
    fmt.Printf("网络操作: %s (%s)\n", netOp.Type, netOp.Target)
}

// 检查质量问题
for _, issue := range result.QualityIssues {
    fmt.Printf("[%s] %s (行 %d)\n", issue.Severity, issue.Message, issue.Line)
}

// 代码评分
fmt.Printf("代码质量评分: %.1f/100\n", result.Score)
```

---

### AnalysisResult 结构

```go
type AnalysisResult struct {
    // 基本信息
    Safe              bool              // 代码是否安全
    Issues            []string          // 发现的问题列表
    Complexity        int               // 代码复杂度
    LineCount         int               // 代码行数
    CharCount         int               // 字符数
    HasDangerousOps   bool              // 是否包含危险操作
    Suggestions       []string          // 改进建议
    
    // 操作检测
    NetworkOps        []NetworkOperation      // 网络操作
    FileSystemOps     []FileSystemOperation   // 文件系统操作
    ProcessOps        []ProcessOperation      // 进程操作
    CryptoIssues      []CryptoIssue          // 加密问题
    DatabaseOps       []DatabaseOperation     // 数据库操作
    
    // 质量检查
    QualityIssues     []QualityIssue    // 质量问题
    Score             float64           // 代码质量评分 (0-100)
}
```

---

### 操作检测结构

#### NetworkOperation

```go
type NetworkOperation struct {
    Type        string // "http", "https", "socket", "dns"
    Target      string // 目标地址
    Line        int    // 行号
    Column      int    // 列号
    Context     string // 代码上下文
    Severity    string // "high", "medium", "low"
    Description string // 描述
}
```

**示例**:
```go
for _, op := range result.NetworkOps {
    fmt.Printf("发现网络操作:\n")
    fmt.Printf("  类型: %s\n", op.Type)
    fmt.Printf("  目标: %s\n", op.Target)
    fmt.Printf("  位置: 行 %d, 列 %d\n", op.Line, op.Column)
    fmt.Printf("  严重性: %s\n", op.Severity)
}
```

#### FileSystemOperation

```go
type FileSystemOperation struct {
    Type        string // "read", "write", "delete", "chmod", "chown"
    Path        string // 文件路径
    Line        int
    Column      int
    Context     string
    Severity    string
    Description string
    IsSensitive bool   // 是否访问敏感目录
}
```

#### ProcessOperation

```go
type ProcessOperation struct {
    Type        string // "exec", "spawn", "kill", "signal"
    Command     string // 执行的命令
    Line        int
    Column      int
    Context     string
    Severity    string
    Description string
}
```

#### CryptoIssue

```go
type CryptoIssue struct {
    Type        string // "weak_algorithm", "hardcoded_key", "insecure_random"
    Algorithm   string // 算法名称
    Line        int
    Column      int
    Context     string
    Severity    string
    Description string
    Suggestion  string // 修复建议
}
```

#### DatabaseOperation

```go
type DatabaseOperation struct {
    Type           string // "query", "execute", "connection"
    Query          string // SQL 查询
    Line           int
    Column         int
    Context        string
    Severity       string
    Description    string
    HasInjection   bool   // 是否存在 SQL 注入风险
}
```

---

### QualityIssue 结构

```go
type QualityIssue struct {
    Category    string // "naming", "style", "best_practice", "performance"
    Severity    string // "high", "medium", "low", "info"
    Message     string // 问题描述
    Line        int    // 行号
    Column      int    // 列号
    Context     string // 代码上下文
    Suggestion  string // 修复建议
    Code        string // 问题代码片段
}
```

**示例**:
```go
for _, issue := range result.QualityIssues {
    fmt.Printf("[%s] %s\n", issue.Severity, issue.Message)
    fmt.Printf("  位置: 行 %d, 列 %d\n", issue.Line, issue.Column)
    fmt.Printf("  代码: %s\n", issue.Code)
    if issue.Suggestion != "" {
        fmt.Printf("  建议: %s\n", issue.Suggestion)
    }
}
```

---

## 自定义规则

### 规则格式

自定义规则使用 YAML 格式定义：

```yaml
rules:
  - name: "禁止使用 eval"
    language: "python"
    pattern: "eval\\("
    severity: "high"
    message: "不要使用 eval()，存在安全风险"
    suggestion: "使用 ast.literal_eval() 或其他安全的替代方案"
    
  - name: "禁止硬编码密码"
    language: "all"
    pattern: "password\\s*=\\s*['\"]\\w+"
    severity: "high"
    message: "不要在代码中硬编码密码"
    suggestion: "使用环境变量或配置文件存储密码"
```

### 加载自定义规则

```go
analyzer := NewCodeAnalyzer()

// 从文件加载
err := analyzer.LoadCustomRules("custom_rules.yaml")
if err != nil {
    log.Fatal(err)
}

// 清除自定义规则
analyzer.ClearCustomRules()
```

---

## 配置选项

通过 `AnalyzerConfig` 配置分析器行为：

```go
config := AnalyzerConfig{
    EnableNetworkDetection:    true,  // 启用网络操作检测
    EnableFileSystemDetection: true,  // 启用文件系统操作检测
    EnableProcessDetection:    true,  // 启用进程操作检测
    EnableCryptoDetection:     true,  // 启用加密问题检测
    EnableDatabaseDetection:   true,  // 启用数据库操作检测
    EnableQualityCheck:        true,  // 启用代码质量检查
    CustomRulesFile:           "custom_rules.yaml",
    StrictMode:                false, // 严格模式
}
```

**严格模式**: 启用后，任何检测到的问题都会将代码标记为不安全。

---

## 代码质量评分

评分算法基于以下因素：

- **基础分**: 100 分
- **扣分项**:
  - 高严重性问题: -10 分/个
  - 中严重性问题: -5 分/个
  - 低严重性问题: -2 分/个
  - 信息级问题: -1 分/个

**评分等级**:
- 90-100: 优秀
- 80-89: 良好
- 70-79: 中等
- 60-69: 需要改进
- <60: 较差

---

## 支持的语言

- **Python**: 完整支持
- **JavaScript**: 完整支持
- **Go**: 完整支持
- **Bash**: 基础支持

---

## 最佳实践

### 1. 在执行前分析

```go
// 先分析代码
result := analyzer.Analyze(language, code)

if !result.Safe {
    return fmt.Errorf("代码不安全: %v", result.Issues)
}

// 安全后再执行
execResult, err := executor.Run(ctx, code, "")
```

### 2. 检查特定操作

```go
result := analyzer.Analyze("python", code)

// 只关心网络操作
if len(result.NetworkOps) > 0 {
    fmt.Println("警告: 代码包含网络操作")
    for _, op := range result.NetworkOps {
        fmt.Printf("- %s: %s\n", op.Type, op.Target)
    }
}
```

### 3. 使用质量评分

```go
result := analyzer.Analyze(language, code)

if result.Score < 70 {
    fmt.Println("代码质量较低，建议改进:")
    for _, issue := range result.QualityIssues {
        if issue.Severity == "high" || issue.Severity == "medium" {
            fmt.Printf("- %s\n", issue.Message)
        }
    }
}
```

### 4. 自定义规则

```go
// 为特定项目添加自定义规则
analyzer := NewCodeAnalyzer()
analyzer.LoadCustomRules("project_rules.yaml")

result := analyzer.Analyze(language, code)
```

---

## 性能考虑

- **缓存**: 分析结果不会自动缓存，如需缓存请在应用层实现
- **并发**: `CodeAnalyzer` 是线程安全的，可以并发使用
- **大文件**: 对于大文件，分析时间与代码行数成正比

**性能优化建议**:
```go
// 对于大量代码分析，使用 goroutine
results := make(chan *AnalysisResult, len(codes))

for _, code := range codes {
    go func(c string) {
        results <- analyzer.Analyze(language, c)
    }(code)
}
```

---

## 错误处理

分析器不会返回错误，但会在结果中标记问题：

```go
result := analyzer.Analyze(language, code)

// 检查是否安全
if !result.Safe {
    // 处理不安全的代码
    log.Printf("发现 %d 个问题", len(result.Issues))
}

// 检查是否有高严重性问题
hasHighSeverity := false
for _, issue := range result.QualityIssues {
    if issue.Severity == "high" {
        hasHighSeverity = true
        break
    }
}
```

---

## 示例：完整的代码审查流程

```go
package main

import (
    "fmt"
    "github.com/yourorg/agent/aiosandbox/code_exec"
)

func reviewCode(language, code string) error {
    analyzer := code_exec.NewCodeAnalyzer()
    
    // 加载项目规则
    if err := analyzer.LoadCustomRules("project_rules.yaml"); err != nil {
        return fmt.Errorf("加载规则失败: %w", err)
    }
    
    // 分析代码
    result := analyzer.Analyze(language, code)
    
    // 1. 检查安全性
    if !result.Safe {
        fmt.Println("❌ 代码不安全")
        for _, issue := range result.Issues {
            fmt.Printf("  - %s\n", issue)
        }
        return fmt.Errorf("代码包含安全问题")
    }
    
    // 2. 检查质量评分
    fmt.Printf("📊 代码质量评分: %.1f/100\n", result.Score)
    
    if result.Score < 80 {
        fmt.Println("⚠️  代码质量需要改进:")
        for _, issue := range result.QualityIssues {
            if issue.Severity == "high" || issue.Severity == "medium" {
                fmt.Printf("  [%s] %s (行 %d)\n", 
                    issue.Severity, issue.Message, issue.Line)
            }
        }
    }
    
    // 3. 报告操作
    if len(result.NetworkOps) > 0 {
        fmt.Printf("🌐 网络操作: %d 个\n", len(result.NetworkOps))
    }
    if len(result.FileSystemOps) > 0 {
        fmt.Printf("📁 文件操作: %d 个\n", len(result.FileSystemOps))
    }
    if len(result.ProcessOps) > 0 {
        fmt.Printf("⚙️  进程操作: %d 个\n", len(result.ProcessOps))
    }
    
    // 4. 提供建议
    if len(result.Suggestions) > 0 {
        fmt.Println("\n💡 改进建议:")
        for _, suggestion := range result.Suggestions {
            fmt.Printf("  - %s\n", suggestion)
        }
    }
    
    fmt.Println("✅ 代码审查完成")
    return nil
}
```

---

## 参考资料

- [自定义规则指南](./CUSTOM_RULES_GUIDE.md)
- [代码质量最佳实践](./CODE_QUALITY_BEST_PRACTICES.md)
- [安全编码指南](./SECURE_CODING_GUIDE.md)

---

**版本**: 1.0  
**更新日期**: 2026-01-31
