// Agent Framework - Comprehensive Usage Example
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"AgentFramework/pkg/tools/sandbox/code"
)

// CodeReviewResult 代码审查结果
type CodeReviewResult struct {
	Safe            bool
	Score           int
	Issues          []string
	Suggestions     []string
	ExecutionResult *code.ExecutionResult
	FormattedCode   string
}

// CodeExecutionPipeline 代码执行流水线
type CodeExecutionPipeline struct {
	module *code.CodeExecutorModule
}

// NewCodeExecutionPipeline 创建新的执行流水线
func NewCodeExecutionPipeline(configFile string) (*CodeExecutionPipeline, error) {
	module, err := code.NewCodeExecutorModuleFromFile(configFile)
	if err != nil {
		return nil, err
	}

	return &CodeExecutionPipeline{
		module: module,
	}, nil
}

// Close 关闭流水线
func (p *CodeExecutionPipeline) Close() {
	p.module.Close()
}

// ReviewAndExecute 审查并执行代码
func (p *CodeExecutionPipeline) ReviewAndExecute(ctx context.Context, language, code string) (*CodeReviewResult, error) {
	result := &CodeReviewResult{}

	// 步骤 1: 分析代码
	fmt.Println("步骤 1: 分析代码安全性和质量...")
	analysis, err := p.module.AnalyzeCode(ctx, language, code)
	if err != nil {
		return nil, fmt.Errorf("分析失败: %w", err)
	}

	result.Safe = analysis.Safe
	result.Score = analysis.Score
	// 将 SecurityIssue 转换为 string
	issues := make([]string, len(analysis.Issues))
	for i, issue := range analysis.Issues {
		issues[i] = issue.Description
	}
	result.Issues = issues
	result.Suggestions = analysis.Suggestions

	fmt.Printf("  - 安全: %v\n", analysis.Safe)
	fmt.Printf("  - 评分: %d/100\n", analysis.Score)
	fmt.Printf("  - 问题: %d 个\n", len(analysis.Issues))

	// 步骤 2: 格式化代码
	fmt.Println("\n步骤 2: 格式化代码...")
	formatted, err := p.module.FormatCode(ctx, language, code)
	if err != nil {
		fmt.Printf("  - 格式化失败: %v\n", err)
		result.FormattedCode = code
	} else {
		result.FormattedCode = formatted
		fmt.Printf("  - 格式化完成\n")
	}

	// 步骤 3: 决定是否执行
	if !analysis.Safe {
		fmt.Println("\n步骤 3: 代码不安全，跳过执行")
		return result, nil
	}

	// 步骤 4: 执行代码
	fmt.Println("\n步骤 3: 执行代码...")
	execResult, err := p.module.ExecuteCode(ctx, language, code, "", 0)
	if err != nil {
		return nil, fmt.Errorf("执行失败: %w", err)
	}

	result.ExecutionResult = execResult
	fmt.Printf("  - 执行成功: %v\n", execResult.Success)
	fmt.Printf("  - 执行时间: %d ms\n", execResult.Duration.Milliseconds())
	fmt.Printf("  - 内存使用: %d MB\n", execResult.MemoryMB)

	return result, nil
}

// 示例 1: 完整的代码审查和执行流程
func completeWorkflowExample() {
	fmt.Println("=== 示例 1: 完整的代码审查和执行流程 ===\n")

	// 创建配置文件
	configYAML := `
executor:
  timeout: 60000
  memory_limit: 512
  cpu_limit: 2
  supported_languages:
    - python
    - javascript
    - go
  execution_mode: auto

analyzer:
  enable_network_detection: true
  enable_filesystem_detection: true
  enable_process_detection: true
  enable_crypto_detection: true
  enable_database_detection: true
  enable_quality_check: true
  strict_mode: false

yaegi:
  preload_stdlib: true
  preload_packages:
    - fmt
    - strings
    - time
  enable_cache: true
  cache_capacity: 100

container:
  enabled: false
`

	tmpConfig := "temp_config.yaml"
	err := os.WriteFile(tmpConfig, []byte(configYAML), 0644)
	if err != nil {
		log.Fatalf("创建配置失败: %v", err)
	}
	defer os.Remove(tmpConfig)

	// 创建流水线
	pipeline, err := NewCodeExecutionPipeline(tmpConfig)
	if err != nil {
		log.Fatalf("创建流水线失败: %v", err)
	}
	defer pipeline.Close()

	// 测试代码
	pythonCode := `
def calculate_fibonacci(n):
    """计算斐波那契数列"""
    if n <= 1:
        return n
    return calculate_fibonacci(n-1) + calculate_fibonacci(n-2)

# 计算前 10 个斐波那契数
for i in range(10):
    print(f"F({i}) = {calculate_fibonacci(i)}")
`

	ctx := context.Background()
	result, err := pipeline.ReviewAndExecute(ctx, "python", pythonCode)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}

	// 打印结果
	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("审查结果:")
	fmt.Printf("  安全: %v\n", result.Safe)
	fmt.Printf("  评分: %d/100\n", result.Score)

	if len(result.Issues) > 0 {
		fmt.Println("\n  问题:")
		for i, issue := range result.Issues {
			fmt.Printf("    %d. %s\n", i+1, issue)
		}
	}

	if result.ExecutionResult != nil {
		fmt.Println("\n执行结果:")
		fmt.Printf("  成功: %v\n", result.ExecutionResult.Success)
		fmt.Printf("  输出:\n%s\n", result.ExecutionResult.Output)
	}

	fmt.Println()
}

// 示例 2: 多语言代码执行
func multiLanguageWorkflowExample() {
	fmt.Println("=== 示例 2: 多语言代码执行 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "go"},
		ExecutionMode:      "auto",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	ctx := context.Background()

	// Python: 数据处理
	fmt.Println("1. Python - 数据处理:")
	pythonCode := `
data = [1, 2, 3, 4, 5]
squared = [x**2 for x in data]
print(f"原始数据: {data}")
print(f"平方后: {squared}")
print(f"总和: {sum(squared)}")
`
	result, _ := module.ExecuteCode(ctx, "python", pythonCode, "", 0)
	fmt.Printf("%s\n", result.Output)

	// JavaScript: Web 逻辑
	fmt.Println("2. JavaScript - Web 逻辑:")
	jsCode := `
const users = [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 },
    { name: 'Carol', age: 35 }
];

const adults = users.filter(u => u.age >= 30);
console.log('成年用户:', adults.map(u => u.name).join(', '));
`
	result, _ = module.ExecuteCode(ctx, "javascript", jsCode, "", 0)
	fmt.Printf("%s\n", result.Output)

	// Go: 系统编程
	fmt.Println("3. Go - 系统编程:")
	goCode := `
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	
	// 模拟一些工作
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	
	elapsed := time.Since(start)
	fmt.Printf("计算完成: sum = %d\n", sum)
	fmt.Printf("耗时: %v\n", elapsed)
}
`
	result, _ = module.ExecuteCode(ctx, "go", goCode, "", 0)
	fmt.Printf("%s\n", result.Output)

	fmt.Println()
}

// 示例 3: 安全代码执行（容器模式）
func secureExecutionExample() {
	fmt.Println("=== 示例 3: 安全代码执行（容器模式） ===\n")

	fullConfig := code.DefaultFullConfig()
	fullConfig.Executor.ExecutionMode = "container"
	fullConfig.Executor.SupportedLanguages = []string{"python"}
	fullConfig.Container.Enabled = true
	fullConfig.Container.NetworkMode = "none"
	fullConfig.Container.CPULimit = "0.5"
	fullConfig.Container.MemoryLimit = "256m"
	fullConfig.Analyzer.StrictMode = true

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 不安全的代码
	unsafeCode := `
import requests

# 尝试网络访问
try:
    response = requests.get('http://example.com')
    print("网络访问成功")
except Exception as e:
    print(f"网络访问失败: {type(e).__name__}")

# 尝试文件访问
try:
    with open('/etc/passwd', 'r') as f:
        print("文件访问成功")
except Exception as e:
    print(f"文件访问失败: {type(e).__name__}")
`

	ctx := context.Background()

	// 先分析
	fmt.Println("分析代码...")
	analysis, _ := module.AnalyzeCode(ctx, "python", unsafeCode)
	fmt.Printf("  安全: %v\n", analysis.Safe)
	fmt.Printf("  问题: %d 个\n", len(analysis.Issues))

	// 在容器中执行
	fmt.Println("\n在隔离容器中执行...")
	result, _ := module.ExecuteCode(ctx, "python", unsafeCode, "", 0)
	fmt.Printf("  执行成功: %v\n", result.Success)
	fmt.Printf("  输出:\n%s\n", result.Output)

	fmt.Println("✓ 容器隔离防止了不安全的操作")
	fmt.Println()
}

// 示例 4: 性能优化（Yaegi + 缓存）
func performanceOptimizationExample() {
	fmt.Println("=== 示例 4: 性能优化（Yaegi + 缓存） ===\n")

	fullConfig := code.DefaultFullConfig()
	fullConfig.Executor.ExecutionMode = "local"
	fullConfig.Executor.SupportedLanguages = []string{"go"}
	fullConfig.Yaegi.EnableCache = true
	fullConfig.Yaegi.CacheCapacity = 100

	module, err := code.NewCodeExecutorModuleWithFullConfig(&fullConfig)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	goCode := `
package main

import "fmt"

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func main() {
	result := factorial(10)
	fmt.Printf("10! = %d\n", result)
}
`

	ctx := context.Background()

	// 测试多次执行
	fmt.Println("执行 10 次，测试缓存效果...")
	var times []time.Duration

	for i := 0; i < 10; i++ {
		start := time.Now()
		module.ExecuteCode(ctx, "go", goCode, "", 0)
		elapsed := time.Since(start)
		times = append(times, elapsed)

		if i == 0 {
			fmt.Printf("  第 1 次（编译）: %v\n", elapsed)
		} else if i == 1 {
			fmt.Printf("  第 2 次（缓存）: %v\n", elapsed)
		}
	}

	// 计算平均时间
	var total time.Duration
	for _, t := range times[1:] { // 跳过第一次
		total += t
	}
	avg := total / time.Duration(len(times)-1)

	fmt.Printf("\n平均时间（缓存）: %v\n", avg)
	fmt.Printf("性能提升: %.2fx\n", float64(times[0])/float64(avg))
	fmt.Println()
}

// 示例 5: 实时代码评审系统
func codeReviewSystemExample() {
	fmt.Println("=== 示例 5: 实时代码评审系统 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 模拟代码提交
	submissions := []struct {
		author string
		code   string
	}{
		{
			author: "Alice",
			code: `
def add(a, b):
    """Add two numbers"""
    return a + b

result = add(2, 3)
print(f"Result: {result}")
`,
		},
		{
			author: "Bob",
			code: `
import os
password = "admin123"
os.system("rm -rf /")
`,
		},
		{
			author: "Carol",
			code: `
def process_data(data):
    result = []
    for item in data:
        if item > 0:
            result.append(item * 2)
    return result

data = [1, -2, 3, -4, 5]
print(process_data(data))
`,
		},
	}

	ctx := context.Background()

	fmt.Println("代码审查结果:")
	fmt.Println("=" + string(make([]byte, 60)))

	for i, sub := range submissions {
		fmt.Printf("\n提交 %d - 作者: %s\n", i+1, sub.author)
		fmt.Println(string(make([]byte, 60)))

		// 分析代码
		analysis, _ := module.AnalyzeCode(ctx, "python", sub.code)

		fmt.Printf("评分: %d/100\n", analysis.Score)
		fmt.Printf("安全: %v\n", analysis.Safe)

		if len(analysis.Issues) > 0 {
			fmt.Println("\n问题:")
			for _, issue := range analysis.Issues {
				fmt.Printf("  ❌ %s\n", issue)
			}
		}

		if len(analysis.Suggestions) > 0 {
			fmt.Println("\n建议:")
			for _, suggestion := range analysis.Suggestions {
				fmt.Printf("  💡 %s\n", suggestion)
			}
		}

		// 决定
		if analysis.Safe && analysis.Score >= 70 {
			fmt.Println("\n✅ 批准合并")
		} else if analysis.Safe && analysis.Score >= 50 {
			fmt.Println("\n⚠️  需要改进后合并")
		} else {
			fmt.Println("\n❌ 拒绝合并")
		}
	}

	fmt.Println()
}

// 示例 6: 代码执行监控和统计
func monitoringExample() {
	fmt.Println("=== 示例 6: 代码执行监控和统计 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	ctx := context.Background()

	// 执行多个任务
	tasks := []struct {
		language string
		code     string
	}{
		{"python", `print("Task 1")`},
		{"python", `print("Task 2")`},
		{"go", `package main; import "fmt"; func main() { fmt.Println("Task 3") }`},
		{"python", `print("Task 4")`},
		{"go", `package main; import "fmt"; func main() { fmt.Println("Task 5") }`},
	}

	fmt.Println("执行任务...")
	var totalTime time.Duration
	successCount := 0
	failureCount := 0

	for i, task := range tasks {
		start := time.Now()
		result, err := module.ExecuteCode(ctx, task.language, task.code, "", 0)
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil || !result.Success {
			failureCount++
			fmt.Printf("  任务 %d [%s]: ❌ 失败 (%v)\n", i+1, task.language, elapsed)
		} else {
			successCount++
			fmt.Printf("  任务 %d [%s]: ✅ 成功 (%v)\n", i+1, task.language, elapsed)
		}
	}

	// 统计信息
	fmt.Println("\n执行统计:")
	fmt.Printf("  总任务数: %d\n", len(tasks))
	fmt.Printf("  成功: %d\n", successCount)
	fmt.Printf("  失败: %d\n", failureCount)
	fmt.Printf("  成功率: %.1f%%\n", float64(successCount)/float64(len(tasks))*100)
	fmt.Printf("  总耗时: %v\n", totalTime)
	fmt.Printf("  平均耗时: %v\n", totalTime/time.Duration(len(tasks)))

	fmt.Println()
}

// 示例 7: JSON API 集成
func jsonAPIExample() {
	fmt.Println("=== 示例 7: JSON API 集成 ===\n")

	config := code.CodeExecutorConfig{
		Timeout:            60000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
		ExecutionMode:      "local",
	}

	module, err := code.NewCodeExecutorModule(config)
	if err != nil {
		log.Fatalf("创建模块失败: %v", err)
	}
	defer module.Close()

	// 模拟 API 请求
	type APIRequest struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}

	type APIResponse struct {
		Success       bool     `json:"success"`
		Output        string   `json:"output,omitempty"`
		Error         string   `json:"error,omitempty"`
		ExecutionTime int64    `json:"execution_time_ms"`
		Safe          bool     `json:"safe"`
		Score         int      `json:"score"`
		Issues        []string `json:"issues,omitempty"`
	}

	request := APIRequest{
		Language: "python",
		Code:     `print("Hello from API!")`,
	}

	ctx := context.Background()

	// 分析
	analysis, _ := module.AnalyzeCode(ctx, request.Language, request.Code)

	// 执行
	var response APIResponse
	if analysis.Safe {
		result, err := module.ExecuteCode(ctx, request.Language, request.Code, "", 0)
		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = result.Success
			response.Output = result.Output
			response.Error = result.Error
			response.ExecutionTime = result.Duration.Milliseconds()
		}
	} else {
		response.Success = false
		response.Error = "Code is not safe to execute"
	}

	response.Safe = analysis.Safe
	response.Score = analysis.Score
	// 将 SecurityIssue 转换为字符串
	issues := make([]string, len(analysis.Issues))
	for i, issue := range analysis.Issues {
		issues[i] = issue.Description
	}
	response.Issues = issues

	// 序列化为 JSON
	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println("API 响应:")
	fmt.Println(string(jsonData))

	fmt.Println()
}

func main() {
	fmt.Println("代码执行模块 - 综合使用示例")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// 运行所有示例
	completeWorkflowExample()
	multiLanguageWorkflowExample()
	secureExecutionExample()
	performanceOptimizationExample()
	codeReviewSystemExample()
	monitoringExample()
	jsonAPIExample()

	fmt.Println("所有示例运行完成！")
	fmt.Println("\n代码执行模块特性总结:")
	fmt.Println("  ✓ 多语言支持（Python, JavaScript, Go, Bash）")
	fmt.Println("  ✓ 增强的代码分析（安全、质量、性能）")
	fmt.Println("  ✓ Yaegi 解释器（428x 性能提升）")
	fmt.Println("  ✓ 容器隔离（最高安全性）")
	fmt.Println("  ✓ 自定义规则（团队规范）")
	fmt.Println("  ✓ 完整的监控和统计")
	fmt.Println("  ✓ 灵活的配置系统")
}
