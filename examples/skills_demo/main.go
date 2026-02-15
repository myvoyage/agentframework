// AgentFramework 技能系统使用示例
// 演示如何使用增强的技能系统

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"AgentFramework/agent"
	"AgentFramework/agent/skills"
	"github.com/cloudwego/eino/components/tool"
)

func main() {
	ctx := context.Background()

	// ============================================================
	// 第1步：创建技能注册表
	// ============================================================
	fmt.Println("=== 步骤1: 创建技能注册表 ===")

	registry := skills.NewSkillRegistry(&skills.RegistryConfig{
		BaseDir:  ".skills/registry",
		AutoSave: true,
	})

	// 注册HTTP GET技能
	httpGetSkill := &skills.SkillEntry{
		ID:          "http_get_user_info",
		Name:        "获取用户信息API",
		Description: "通过用户ID获取用户详细信息的HTTP GET请求",
		Category:    "http",
		Tags:        []string{"http", "get", "user", "api"},
		Version:     "1.0.0",
		InputSchema: &skills.Schema{
			Type: "object",
			Properties: map[string]*skills.PropertyInfo{
				"user_id": {
					Type:        "integer",
					Description: "用户ID",
					Required:    true,
					Minimum:     ptrFloat64(1),
				},
				"fields": {
					Type:        "array",
					Description: "需要返回的字段列表",
					Required:    false,
					// Items: &skills.PropertyInfo{
					// 	Type: "string",
					// 	Enum: []string{"id", "name", "email", "avatar", "created_at"},
					// },
				},
			},
			Required: []string{"user_id"},
		},
		OutputSchema: &skills.Schema{
			Type: "object",
			Properties: map[string]*skills.PropertyInfo{
				"id":       {Type: "integer", Description: "用户ID"},
				"name":     {Type: "string", Description: "用户名"},
				"email":    {Type: "string", Description: "邮箱地址"},
				"avatar":   {Type: "string", Description: "头像URL"},
				"created_at": {Type: "string", Description: "创建时间", Format: "date-time"},
			},
		},
		GeneratedFile: "api/user_api.go",
		GeneratedLine: 42,
	}

	// 尝试注册（会自动去重）
	if err := registry.Register(httpGetSkill); err != nil {
		if exists, ok := registry.GetByID(httpGetSkill.ID); ok {
			fmt.Printf("✅ 技能已存在，直接复用：%s (位于 %s:%d)\n",
				exists.Name, exists.GeneratedFile, exists.GeneratedLine)
		}
	} else {
		fmt.Printf("✅ 新技能注册成功：%s\n", httpGetSkill.Name)
	}

	// ============================================================
	// 第2步：创建技能定义管理器
	// ============================================================
	fmt.Println("\n=== 步骤2: 创建技能定义管理器 ===")

	defManager := skills.NewDefinitionManager("agent/skills/definitions")

	// 加载技能定义
	if definition, err := defManager.Load("http_request"); err == nil {
		fmt.Printf("✅ 加载技能定义：%s (%s)\n", definition.Name, definition.ID)
		fmt.Printf("   触发条件：%v\n", definition.Triggers)
		fmt.Printf("   工作流步骤：%d\n", len(definition.Workflow))
	}

	// 列出所有技能定义
	allDefs := defManager.List()
	fmt.Printf("\n📋 已加载 %d 个技能定义：\n", len(allDefs))
	for _, def := range allDefs {
		fmt.Printf("   - %s (%s)\n", def.Name, def.Category)
	}

	// ============================================================
	// 第3步：创建示例模板库
	// ============================================================
	fmt.Println("\n=== 步骤3: 创建示例模板库 ===")

	library := skills.NewExampleLibrary(".skills/examples")

	// 创建内置模板
	if err := library.CreateBuiltInTemplates(); err != nil {
		log.Printf("Warning: 创建内置模板失败: %v\n", err)
	}

	fmt.Printf("✅ 示例库已创建，包含 %d 个模板\n", library.Count())

	// 渲染一个示例模板
	code, err := library.Render(ctx, "http_get_request", map[string]interface{}{
		"FunctionName": "GetUserInfo",
		"Description":  "获取用户信息",
		"Params":       "ctx context.Context, userID int64",
		"URLPattern":   `"https://api.example.com/users/%d"`,
		"URLArgs":      "userID",
		"Headers": map[string]string{
			"Content-Type": "application/json",
		},
	})
	if err != nil {
		log.Printf("Error: 渲染模板失败: %v\n", err)
	} else {
		fmt.Printf("\n✅ 渲染生成的代码示例：\n%s\n", code)
	}

	// ============================================================
	// 第4步：创建渐进式加载器
	// ============================================================
	fmt.Println("\n=== 步骤4: 创建渐进式加载器 ===")

	loader := skills.NewProgressiveLoader("agent/skills/definitions")

	// 只加载元数据（快速）
	metas, _ := loader.ListSkills()
	fmt.Printf("✅ 快速加载 %d 个技能元数据（不含详细定义）\n", len(metas))

	for _, meta := range metas {
		fmt.Printf("   - %s (%s) - %s\n", meta.Name, meta.Category, meta.Description)
	}

	// 按需加载完整定义
	fmt.Println("\n按需加载完整定义...")
	if definition, err := loader.LoadSkill("http_request"); err == nil {
		fmt.Printf("✅ 加载完整定义：%s\n", definition.Name)
		fmt.Printf("   包含 %d 个工作流步骤\n", len(definition.Workflow))
	}

	// 获取统计信息
	stats := loader.GetStats()
	fmt.Printf("\n📊 加载器统计：\n")
	fmt.Printf("   总技能数：%v\n", stats["total_skills"])
	fmt.Printf("   缓存命中数：%v\n", stats["cached_definitions"])
	fmt.Printf("   加载策略：%v\n", stats["strategy"])

	// ============================================================
	// 第5步：集成到Agent系统
	// ============================================================
	fmt.Println("\n=== 步骤5: 集成到Agent系统 ===")

	// 创建带技能系统的Host
	_, err = agent.NewHost(
		ctx,
		&agent.HostConfig{},
		nil,
		make(map[string]tool.BaseTool),
		agent.WithSkillRegistry(registry),
		agent.WithDefinitionManager(defManager),
		agent.WithExampleLibrary(library),
		agent.WithSkillLoader(loader),
	)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	fmt.Println("✅ Host已创建，集成所有新组件")

	// ============================================================
	// 第6步：查询和使用技能
	// ============================================================
	fmt.Println("\n=== 步骤6: 查询和使用技能 ===")

	// 按分类查询
	apiSkills := registry.ListByCategory("http")
	fmt.Printf("\n📋 HTTP类别技能（%d个）：\n", len(apiSkills))
	for _, skill := range apiSkills {
		fmt.Printf("   - %s (%s)\n", skill.Name, skill.ID)
	}

	// 智能查询
	results, _ := registry.Find(&skills.SkillQuery{
		Category:    "http",
		Tags:        []string{"user"},
		UsedCountMin: 0,
	})
	fmt.Printf("\n🔍 搜索结果（用户相关HTTP技能，%d个）：\n", len(results))
	for _, skill := range results {
		fmt.Printf("   - %s (使用%d次)\n", skill.Name, skill.UsedCount)
	}

	// 记录使用
	registry.RecordUsage("http_get_user_info", "demo_user")
	fmt.Println("\n✅ 已记录技能使用")

	// 获取统计信息
	allStats := registry.GetStats()
	fmt.Printf("\n📊 注册表统计：\n")
	fmt.Printf("   总技能数：%v\n", allStats["total_skills"])
	fmt.Printf("   总使用次数：%v\n", allStats["total_uses"])
	fmt.Printf("   分类数：%v\n", allStats["categories"])
	fmt.Printf("   未使用技能：%v\n", allStats["never_used"])
	if mostUsed, ok := allStats["most_used"].(*skills.SkillEntry); ok && mostUsed != nil {
		fmt.Printf("   最常用技能：%s (%d次)\n", mostUsed.Name, mostUsed.UsedCount)
	}

	// ============================================================
	// 第7步：导出文档
	// ============================================================
	fmt.Println("\n=== 步骤7: 导出文档 ===")

	// 导出到Markdown
	mdFile, err := os.Create(".skills/README.md")
	if err != nil {
		log.Fatal(err)
	}
	defer mdFile.Close()

	if err := registry.ExportToMarkdown(mdFile); err != nil {
		log.Printf("Warning: 导出Markdown失败: %v\n", err)
	} else {
		fmt.Println("✅ 已导出技能文档到 .skills/README.md")
	}

	// ============================================================
	// 总结
	// ============================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 技能系统增强示例运行完成！")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n核心功能演示：")
	fmt.Println("  ✅ 技能注册表 - 去重、查询、统计")
	fmt.Println("  ✅ 技能定义管理器 - 声明式流程定义")
	fmt.Println("  ✅ 示例模板库 - 统一代码风格")
	fmt.Println("  ✅ 渐进式加载器 - 按需加载，节省Token")
	fmt.Println("\n下一步：")
	fmt.Println("  1. 查看 .skills/README.md 了解所有技能")
	fmt.Println("  2. 在 agent/skills/definitions/ 中添加自定义技能")
	fmt.Println("  3. 在 .skills/examples/ 中添加自定义模板")
	fmt.Println("  4. 集成到您的Agent应用中")
	fmt.Println("\n文档参考：")
	fmt.Println("  - SKILL_COMPARISON_ANALYSIS.md")
	fmt.Println("  - SKILL_ENHANCEMENT_GUIDE.md")
	fmt.Println("  - SKILLS_API_QUICK_REFERENCE.md")
	fmt.Println("  - SKILL_SYSTEM_SUMMARY.md")
}

// 辅助函数
func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
