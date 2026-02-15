// 增强执行器使用示例
// 演示如何使用 EnhancedSkillExecutor 执行声明式流程

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"AgentFramework/agent/skills"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== AgentFramework 增强执行器演示 ===\n")

	// ============================================================
	// 第1步：创建增强执行器
	// ============================================================
	fmt.Println("步骤1: 创建增强执行器")

	config := &skills.ExecutorConfig{
		EnableRetry:    true,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		EnableTimeout:   true,
		DefaultTimeout:  30 * time.Second,
		EnableSkip:     true,
		EnableCache:    true,
		EnableLog:      true,
		LogLevel:       "info",
	}

	executor := skills.NewEnhancedSkillExecutor(config)

	// ============================================================
	// 第2步：设置组件
	// ============================================================
	fmt.Println("\n步骤2: 设置组件（注册表、示例库、定义）")

	// 创建注册表
	registry := skills.NewSkillRegistry(&skills.RegistryConfig{
		BaseDir:  ".skills/registry",
		AutoSave: true,
	})

	// 创建示例库
	library := skills.NewExampleLibrary(".skills/examples")
	library.CreateBuiltInTemplates()

	// 创建定义管理器
	defManager := skills.NewDefinitionManager("agent/skills/definitions")

	// 加载技能定义
	definition, err := defManager.Load("http_request")
	if err != nil {
		log.Printf("Warning: 无法加载技能定义: %v\n", err)
		definition = createSampleDefinition()
	}

	// 设置到执行器
	executor.SetRegistry(registry)
	executor.SetExamples(library)
	executor.SetDefinition(definition)

	fmt.Printf("✅ 组件设置完成\n")
	fmt.Printf("   - 注册表: %d 个技能\n", len(registry.ListAll()))
	fmt.Printf("   - 示例库: %d 个模板\n", library.Count())
	fmt.Printf("   - 定义: %s\n", definition.Name)

	// ============================================================
	// 第3步：定义输入并创建执行上下文
	// ============================================================
	fmt.Println("\n步骤3: 准备输入和上下文")

	// 用户输入（模拟）
	userInput := `{
		"method": "GET",
		"url": "https://api.example.com/users/123"
	}`

	inputJSON, _ := JSONMarshal(userInput)
	fmt.Printf("用户输入: %s\n", inputJSON)

	// 创建执行上下文
	execCtx := skills.NewExecutionContext()
	execCtx.SetEnv("ENV", "production")
	execCtx.Workspace = "/workspace"
	execCtx.SetMetadata("config", map[string]interface{}{
		"cache_disabled": false,
		"force_new":       false,
	})

	fmt.Printf("✅ 执行上下文已创建\n")

	// ============================================================
	// 第4步：使用注册表（去重检查）
	// ============================================================
	fmt.Println("\n步骤4: 注册表去重检查")

	// 生成技能ID
	skillID := fmt.Sprintf("http_request:%s", hashString(string(inputJSON)))
	fmt.Printf("技能ID: %s\n", skillID)

	// 检查是否已存在
	if exists, ok := registry.GetByID(skillID); ok {
		fmt.Printf("✅ 技能已存在，直接复用\n")
		fmt.Printf("   名称: %s\n", exists.Name)
		fmt.Printf("   位置: %s:%d\n", exists.GeneratedFile, exists.GeneratedLine)
		fmt.Printf("   使用次数: %d\n", exists.UsedCount)

		// 记录使用
		registry.RecordUsage(skillID, "demo_user")
	} else {
		fmt.Printf("✅ 新技能，将创建\n")

		// 创建新技能条目
		entry := &skills.SkillEntry{
			ID:          skillID,
			Name:        "获取用户信息API",
			Description: "通过用户ID获取用户信息的HTTP GET请求",
			Category:    "http",
			Tags:        []string{"http", "get", "user"},
			Version:     "1.0.0",
			InputSchema: &skills.Schema{
				Type: "object",
				Properties: map[string]*skills.PropertyInfo{
					"method": {
						Type:        "string",
						Description: "HTTP方法",
						Required:    true,
						Enum:        []string{"GET", "POST", "PUT", "DELETE"},
					},
					"url": {
						Type:        "string",
						Description: "请求URL",
						Required:    true,
					},
				},
				Required: []string{"method", "url"},
			},
			OutputSchema: &skills.Schema{
				Type: "object",
				Properties: map[string]*skills.PropertyInfo{
					"id":       {Type: "integer", Description: "用户ID"},
					"name":     {Type: "string", Description: "用户名"},
					"email":    {Type: "string", Description: "邮箱"},
				},
			},
			GeneratedFile: "api/user_api.go",
			GeneratedLine: 42,
		}

		registry.Register(entry)
		fmt.Printf("✅ 新技能已注册\n")
	}

	// ============================================================
	// 第5步：执行技能（使用增强执行器）
	// ============================================================
	fmt.Println("\n步骤5: 执行声明式流程")

	executor.SetDefinition(definition)

	// 执行技能
	result, err := executor.Execute(ctx, string(inputJSON), execCtx)
	if err != nil {
		log.Printf("执行失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 技能执行完成\n")
	fmt.Printf("   结果类型: %T\n", result)

	// 解析结果
	if resultMap, ok := result.(map[string]interface{}); ok {
		fmt.Printf("\n执行结果:\n")
		for key, value := range resultMap {
			fmt.Printf("  %s: %v\n", key, value)
		}

		// 如果生成了代码，显示代码
		if code, ok := resultMap["code"].(string); ok {
			fmt.Printf("\n生成的代码:\n%s\n", code)
		}
	}

	// ============================================================
	// 第6步: 使用统计信息
	// ============================================================
	fmt.Println("\n步骤6: 查看使用统计")

	// 重新加载技能信息
	updatedEntry, _ := registry.GetByID(skillID)
	if updatedEntry != nil {
		fmt.Printf("使用次数: %d\n", updatedEntry.UsedCount)
		fmt.Printf("最后使用: %s\n", updatedEntry.LastUsed.Format("2006-01-02 15:04:05"))
	}

	allStats := registry.GetStats()
	fmt.Printf("\n注册表统计:\n")
	fmt.Printf("  总技能数: %v\n", allStats["total_skills"])
	fmt.Printf("  总使用次数: %v\n", allStats["total_uses"])
	fmt.Printf("  分类数: %v\n", allStats["categories"])
	fmt.Printf("  未使用技能: %v\n", allStats["never_used"])

	if mostUsed, ok := allStats["most_used"].(*skills.SkillEntry); ok && mostUsed != nil {
		fmt.Printf("  最常用技能: %s (%d次)\n", mostUsed.Name, mostUsed.UsedCount)
	}

	// ============================================================
	// 第7步: 导出文档
	// ============================================================
	fmt.Println("\n步骤7: 导出技能文档")

	markdownFile := ".skills/registry/SKILL_USAGE.md"
	file, err := os.Create(markdownFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err := registry.ExportToMarkdown(file); err != nil {
		log.Printf("Warning: 导出文档失败: %v\n", err)
	} else {
		fmt.Printf("✅ 已导出技能使用文档: %s\n", markdownFile)
	}

	// ============================================================
	// 总结
	// ============================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 增强执行器演示完成！")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n核心功能演示：")
	fmt.Println("  ✅ 声明式流程执行")
	fmt.Println("  ✅ 自动去重检查")
	fmt.Println("  ✅ 步骤跳过逻辑")
	fmt.Println("  ✅ 失败重试机制")
	fmt.Println("  ✅ 使用统计追踪")
	fmt.Println("  ✅ 文档自动生成")

	fmt.Println("\n关键优势：")
	fmt.Println("  🎯 声明式定义 - 易于阅读和维护")
	fmt.Println("  🔄 自动去重 - 避免重复开发")
	fmt.Println("  📝 代码生成 - 统一代码风格")
	fmt.Println("  ⚡ 高性能执行 - 缓存和优化")
	fmt.Println("  📊 完整统计 - 使用追踪和分析")

	fmt.Println("\n下一步：")
	fmt.Println("  1. 查看 agent/skills/definitions/ 中的技能定义")
	fmt.Println("  2. 创建自定义技能定义文件")
	fmt.Println("  3. 集成到您的Agent应用中")
	fmt.Println("   4. 利用模板库生成代码")

	fmt.Println("\n文档参考：")
	fmt.Println("  - NEXT_STEPS.md")
	fmt.Println("  - SKILL_ENHANCEMENT_GUIDE.md")
	fmt.Println("  - SKILLS_API_QUICK_REFERENCE.md")
}

// createSampleDefinition 创建示例技能定义
func createSampleDefinition() *skills.SkillDefinition {
	return &skills.SkillDefinition{
		ID:          "http_request",
		Name:        "HTTP请求技能",
		Description: "发送HTTP请求的标准流程",
		Version:     "2.0.0",
		Category:    "http",
		Author:      "AgentFramework Team",
		License:     "AGPL-3.0-or-later",
		Triggers: []string{
			"user mentions HTTP, API, REST",
			"user provides URL",
		},
		Workflow: []skills.WorkflowStep{
			{
				ID:          "validate_input",
				Name:        "验证输入",
				Action:      "validate",
				Description: "验证必需的输入参数",
				Parameters: map[string]string{
					"required_fields": "method,url",
				},
			},
			{
				ID:          "check_registry",
				Name:        "检查注册表",
				Action:      "check_exists",
				Description: "检查是否已有相同API",
				SkipIf:      "config.skip_check or config.force_new",
				Parameters: map[string]string{
					"key": "api:{{method}}:{{url}}",
				},
			},
			{
				ID:          "send_request",
				Name:        "发送请求",
				Action:      "execute",
				Description: "发送HTTP请求",
				Timeout:     30 * time.Second,
				RetryOnFailure: true,
				MaxRetries:      3,
				Parameters: map[string]string{
					"use_template": "http_{{method | lower}}_request",
				},
			},
			{
				ID:          "cleanup",
				Name:        "清理资源",
				Action:      "cleanup",
				Description: "清理临时资源",
			},
		},
	}
}

// JSONMarshal JSON序列化
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// hashString 生成字符串哈希
func hashString(s string) string {
	// 简化实现
	hash := uint32(2166136261)
	for _, c := range s {
		hash *= 31
		hash ^= uint32(c)
	}
	return fmt.Sprintf("%x", hash)
}
