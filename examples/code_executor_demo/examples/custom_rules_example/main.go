// Agent Framework - Custom Rules Example
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"AgentFramework/pkg/tools/sandbox/code"
)

// 示例 1: 基础自定义规则
func basicCustomRulesExample() {
	fmt.Println("=== 示例 1: 基础自定义规则 ===\n")

	// 创建自定义规则文件
	rulesYAML := `
rules:
  - name: "禁止使用 eval"
    language: "python"
    pattern: "eval\\("
    severity: "critical"
    message: "不要使用 eval()，这是一个严重的安全风险"
    suggestion: "使用 ast.literal_eval() 或其他安全的替代方案"

  - name: "禁止使用 exec"
    language: "python"
    pattern: "exec\\("
    severity: "critical"
    message: "不要使用 exec()，这是一个严重的安全风险"
    suggestion: "重新设计代码，避免动态执行"
`

	// 写入临时文件
	tmpFile := "temp_custom_rules.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	// 创建配置
	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 测试代码
	code := `
# 使用 eval（会被检测）
result = eval("2 + 2")
print(result)

# 使用 exec（会被检测）
exec("print('Hello')")
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("代码安全: %v\n", result.Safe)
	fmt.Printf("发现问题: %d 个\n", len(result.Issues))

	if len(result.Issues) > 0 {
		fmt.Println("\n检测到的问题:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n改进建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println()
}

// 示例 2: 多语言自定义规则
func multiLanguageRulesExample() {
	fmt.Println("=== 示例 2: 多语言自定义规则 ===\n")

	rulesYAML := `
rules:
  # Python 规则
  - name: "禁止使用 pickle"
    language: "python"
    pattern: "pickle\\.(loads|load|dumps|dump)"
    severity: "high"
    message: "pickle 可能不安全，特别是处理不可信数据时"
    suggestion: "使用 json 或其他安全的序列化方式"

  # JavaScript 规则
  - name: "禁止使用 eval"
    language: "javascript"
    pattern: "eval\\("
    severity: "critical"
    message: "eval() 是一个严重的安全风险"
    suggestion: "使用 JSON.parse() 或其他安全的替代方案"

  # Go 规则
  - name: "禁止使用 unsafe 包"
    language: "go"
    pattern: "import\\s+\"unsafe\""
    severity: "high"
    message: "unsafe 包绕过了 Go 的类型安全"
    suggestion: "尽可能避免使用 unsafe 包"
`

	tmpFile := "temp_multi_lang_rules.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Executor.SupportedLanguages = []string{"python", "javascript", "go"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	ctx := context.Background()

	// 测试 Python
	fmt.Println("测试 Python 代码:")
	pythonCode := `
import pickle
data = pickle.loads(untrusted_data)
`
	result, _ := module.AnalyzeCode(ctx, "python", pythonCode)
	fmt.Printf("  安全: %v, 问题: %d\n", result.Safe, len(result.Issues))
	if len(result.Issues) > 0 {
		fmt.Printf("  问题: %s\n", result.Issues[0])
	}

	// 测试 JavaScript
	fmt.Println("\n测试 JavaScript 代码:")
	jsCode := `
const result = eval(userInput);
console.log(result);
`
	result, _ = module.AnalyzeCode(ctx, "javascript", jsCode)
	fmt.Printf("  安全: %v, 问题: %d\n", result.Safe, len(result.Issues))
	if len(result.Issues) > 0 {
		fmt.Printf("  问题: %s\n", result.Issues[0])
	}

	// 测试 Go
	fmt.Println("\n测试 Go 代码:")
	goCode := `
package main

import "unsafe"

func main() {
	// 使用 unsafe
}
`
	result, _ = module.AnalyzeCode(ctx, "go", goCode)
	fmt.Printf("  安全: %v, 问题: %d\n", result.Safe, len(result.Issues))
	if len(result.Issues) > 0 {
		fmt.Printf("  问题: %s\n", result.Issues[0])
	}

	fmt.Println()
}

// 示例 3: 严重性级别
func severityLevelsExample() {
	fmt.Println("=== 示例 3: 严重性级别 ===\n")

	rulesYAML := `
rules:
  - name: "关键问题 - SQL 注入"
    language: "python"
    pattern: "execute\\(.*%.*\\)"
    severity: "critical"
    message: "可能存在 SQL 注入漏洞"
    suggestion: "使用参数化查询"

  - name: "高危问题 - 硬编码密码"
    language: "python"
    pattern: "password\\s*=\\s*['\"]\\w+"
    severity: "high"
    message: "不要硬编码密码"
    suggestion: "使用环境变量或密钥管理服务"

  - name: "中等问题 - 使用 print 调试"
    language: "python"
    pattern: "print\\(.*debug.*\\)"
    severity: "medium"
    message: "生产代码中不应有调试 print"
    suggestion: "使用日志系统"

  - name: "低危问题 - 长行"
    language: "python"
    pattern: ".{120,}"
    severity: "low"
    message: "代码行过长"
    suggestion: "将长行拆分为多行"
`

	tmpFile := "temp_severity_rules.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	code := `
import sqlite3

# Critical: SQL 注入
cursor.execute("SELECT * FROM users WHERE id = %s" % user_id)

# High: 硬编码密码
password = "admin123"

# Medium: 调试 print
print("debug: user_id =", user_id)

# Low: 长行
very_long_line = "这是一个非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常长的行"
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Printf("代码评分: %d/100\n", result.Score)
	fmt.Printf("发现问题: %d 个\n\n", len(result.Issues))

	// 按严重性分组
	severityGroups := map[string][]string{
		"critical": {},
		"high":     {},
		"medium":   {},
		"low":      {},
	}

	for _, issue := range result.Issues {
		// 简化示例，实际需要从 issue 中提取严重性
		severityGroups["critical"] = append(severityGroups["critical"], issue.Description)
	}

	// 打印问题
	fmt.Println("问题列表:")
	for _, issue := range result.Issues {
		fmt.Printf("  - %s\n", issue.Description)
	}

	fmt.Println()
}

// 示例 4: 团队编码规范
func teamCodingStandardsExample() {
	fmt.Println("=== 示例 4: 团队编码规范 ===\n")

	rulesYAML := `
rules:
  # 命名规范
  - name: "类名必须使用 PascalCase"
    language: "python"
    pattern: "class\\s+[a-z]"
    severity: "medium"
    message: "类名应使用 PascalCase"
    suggestion: "例如: class UserManager"

  - name: "常量必须使用大写"
    language: "python"
    pattern: "^[a-z_]+\\s*=\\s*['\"].*['\"]$"
    severity: "low"
    message: "常量应使用大写字母"
    suggestion: "例如: API_KEY = 'xxx'"

  # 文档规范
  - name: "函数必须有文档字符串"
    language: "python"
    pattern: "def\\s+\\w+\\([^)]*\\):\\s*$"
    severity: "medium"
    message: "函数缺少文档字符串"
    suggestion: "添加 docstring 说明函数用途"

  # 导入规范
  - name: "禁止使用 import *"
    language: "python"
    pattern: "from\\s+\\w+\\s+import\\s+\\*"
    severity: "medium"
    message: "不要使用 import *"
    suggestion: "明确导入需要的内容"

  # 错误处理规范
  - name: "不要使用裸 except"
    language: "python"
    pattern: "except:\\s*$"
    severity: "high"
    message: "不要使用裸 except"
    suggestion: "指定具体的异常类型"
`

	tmpFile := "temp_team_standards.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	code := `
from os import *

class userManager:
    def process_data(self, data):
        try:
            result = data.process()
        except:
            pass
        return result

api_key = "secret123"
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Println("团队编码规范检查结果:")
	fmt.Printf("代码评分: %d/100\n", result.Score)
	fmt.Printf("发现问题: %d 个\n\n", len(result.Issues))

	if len(result.Issues) > 0 {
		fmt.Println("违反规范:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n改进建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println()
}

// 示例 5: 安全审计规则
func securityAuditExample() {
	fmt.Println("=== 示例 5: 安全审计规则 ===\n")

	rulesYAML := `
rules:
  # 密码安全
  - name: "弱密码哈希算法"
    language: "python"
    pattern: "hashlib\\.(md5|sha1)\\("
    severity: "critical"
    message: "MD5 和 SHA1 已不安全"
    suggestion: "使用 SHA256 或更强的算法"

  # 加密安全
  - name: "硬编码加密密钥"
    language: "python"
    pattern: "(key|secret|token)\\s*=\\s*['\"][^'\"]{8,}['\"]"
    severity: "critical"
    message: "不要硬编码密钥"
    suggestion: "使用环境变量或密钥管理服务"

  # 输入验证
  - name: "缺少输入验证"
    language: "python"
    pattern: "request\\.(GET|POST|args|form)\\[.*\\]"
    severity: "high"
    message: "直接使用用户输入可能不安全"
    suggestion: "验证和清理用户输入"

  # 文件操作
  - name: "不安全的文件操作"
    language: "python"
    pattern: "open\\(.*\\+.*user"
    severity: "high"
    message: "使用用户输入的文件路径可能不安全"
    suggestion: "验证文件路径，使用白名单"

  # 命令执行
  - name: "命令注入风险"
    language: "python"
    pattern: "os\\.(system|popen)\\(.*\\+.*user"
    severity: "critical"
    message: "可能存在命令注入漏洞"
    suggestion: "使用 subprocess 并验证输入"
`

	tmpFile := "temp_security_audit.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Analyzer.StrictMode = true // 启用严格模式
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	code := `
import hashlib
import os

# 弱哈希算法
password_hash = hashlib.md5(password.encode()).hexdigest()

# 硬编码密钥
api_secret = "sk_live_1234567890abcdef"

# 缺少输入验证
user_id = request.GET['user_id']

# 不安全的文件操作
filename = user_input + ".txt"
with open(filename, 'r') as f:
    data = f.read()

# 命令注入风险
os.system("ls " + user_path)
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Println("安全审计报告:")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Printf("\n代码安全: %v\n", result.Safe)
	fmt.Printf("安全评分: %d/100\n", result.Score)
	fmt.Printf("发现漏洞: %d 个\n\n", len(result.Issues))

	if len(result.Issues) > 0 {
		fmt.Println("🔴 安全漏洞:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n💡 修复建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println("\n" + string(make([]byte, 50)))
	if result.Safe {
		fmt.Println("✅ 通过安全审计")
	} else {
		fmt.Println("❌ 未通过安全审计，需要修复")
	}

	fmt.Println()
}

// 示例 6: 性能优化规则
func performanceRulesExample() {
	fmt.Println("=== 示例 6: 性能优化规则 ===\n")

	rulesYAML := `
rules:
  - name: "避免在循环中使用 +"
    language: "python"
    pattern: "for.*:\\s*\\n\\s*.*\\+="
    severity: "medium"
    message: "在循环中使用 += 拼接字符串效率低"
    suggestion: "使用 list 和 join()"

  - name: "避免重复计算"
    language: "python"
    pattern: "for.*len\\(.*\\):"
    severity: "low"
    message: "在循环条件中重复调用 len()"
    suggestion: "将 len() 结果保存到变量"

  - name: "使用列表推导式"
    language: "python"
    pattern: "for.*:\\s*\\n\\s*.*\\.append\\("
    severity: "low"
    message: "可以使用列表推导式"
    suggestion: "列表推导式更快更简洁"
`

	tmpFile := "temp_performance_rules.yaml"
	err := os.WriteFile(tmpFile, []byte(rulesYAML), 0644)
	if err != nil {
		log.Fatalf("创建规则文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	fullConfig := code.DefaultFullConfig()
	fullConfig.Analyzer.CustomRulesFile = tmpFile
	fullConfig.Executor.SupportedLanguages = []string{"python"}

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	code := `
# 低效的字符串拼接
result = ""
for item in items:
    result += str(item)

# 重复计算 len()
for i in range(len(data)):
    process(data[i])

# 可以用列表推导式
squares = []
for x in range(10):
    squares.append(x * x)
`

	ctx := context.Background()
	result, err := module.AnalyzeCode(ctx, "python", code)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	fmt.Println("性能优化建议:")
	fmt.Printf("代码评分: %d/100\n", result.Score)
	fmt.Printf("发现问题: %d 个\n\n", len(result.Issues))

	if len(result.Suggestions) > 0 {
		fmt.Println("优化建议:")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println()
}

func main() {
	fmt.Println("代码执行模块 - 自定义规则示例")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// 运行所有示例
	basicCustomRulesExample()
	multiLanguageRulesExample()
	severityLevelsExample()
	teamCodingStandardsExample()
	securityAuditExample()
	performanceRulesExample()

	fmt.Println("所有示例运行完成！")
	fmt.Println("\n自定义规则的优势:")
	fmt.Println("  ✓ 适应团队特定需求")
	fmt.Println("  ✓ 强制执行编码规范")
	fmt.Println("  ✓ 自动化安全审计")
	fmt.Println("  ✓ 提高代码质量")
	fmt.Println("  ✓ 支持多种语言")
}
