// Agent Framework - Skill System Integration Demo
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"AgentFramework/agent"
	"AgentFramework/agent/skills"
)

const demoHostConfigYAML = `
name: skill_integration_demo
version: 1.0.0
defaultModel: default
threadStore:
  type: memory
skillSystemDir: ".skills"
agents:
  - name: skill_assistant
    kind: chat
    model: default
    instructions: "You are an AI assistant with access to various skills. Use the available skills to help users accomplish tasks."
    middlewares: ["logging"]
`

func main() {
	ctx := context.Background()

	fmt.Println("=== AgentFramework 技能系统集成演示 ===\n")

	// ============================================================
	// 步骤1：加载配置并创建 Host
	// ============================================================
	fmt.Println("步骤1: 创建 Host（集成技能系统）")

	cfg, err := agent.LoadHostConfigFile("config.yaml")
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		log.Printf("Warning: 无法加载 config.yaml: %v\n", err)
		log.Printf("使用默认配置...\n")

		// 创建一个简单的配置
		cfg = &agent.HostConfig{
			Name:           "skill_integration_demo",
			DefaultModel:   "default",
			SkillSystemDir: ".skills",
		}
	}

	// 创建 Host（会自动初始化技能系统）
	host, err := agent.NewHost(ctx, cfg, nil, nil)
	if err != nil {
		log.Fatalf("创建 Host 失败: %v\n", err)
	}

	fmt.Printf("✅ Host 创建成功\n")
	fmt.Printf("   应用名称: %s\n", cfg.Name)

	// ============================================================
	// 步骤2：获取技能系统
	// ============================================================
	fmt.Println("\n步骤2: 获取技能系统")

	skillSystem, err := host.GetSkillSystem()
	if err != nil {
		log.Fatalf("获取技能系统失败: %v\n", err)
	}

	fmt.Printf("✅ 技能系统已初始化\n")

	// ============================================================
	// 步骤3：探索已注册的技能
	// ============================================================
	fmt.Println("\n步骤3: 探索已注册的技能")

	registry := skillSystem.Registry()
	allSkills := registry.ListAll()

	fmt.Printf("📋 已注册 %d 个技能:\n", len(allSkills))
	for _, skill := range allSkills {
		fmt.Printf("   - %s (%s)\n", skill.Name, skill.Category)
	}

	// ============================================================
	// 步骤4：查看技能定义
	// ============================================================
	fmt.Println("\n步骤4: 查看技能定义")

	defManager := skillSystem.DefinitionManager()
	definitions := defManager.List()

	fmt.Printf("📋 可用 %d 个技能定义:\n", len(definitions))
	for _, def := range definitions {
		fmt.Printf("   - %s (%s) - %s\n", def.Name, def.Category, def.Description)
		fmt.Printf("     工作流步骤: %d\n", len(def.Workflow))
	}

	// ============================================================
	// 步骤5：使用渐进式加载器
	// ============================================================
	fmt.Println("\n步骤5: 使用渐进式加载器")

	loader := skillSystem.ProgressiveLoader()

	// 快速加载元数据
	metas, _ := loader.ListSkills()
	fmt.Printf("⚡ 快速加载 %d 个技能元数据\n", len(metas))

	// 按需加载完整定义
	fmt.Println("\n按需加载技能定义...")
	for _, meta := range metas {
		def, err := loader.LoadSkill(meta.ID)
		if err != nil {
			log.Printf("Warning: 加载技能 %s 失败: %v\n", meta.ID, err)
			continue
		}
		fmt.Printf("   ✅ %s - %d 个工作流步骤\n", def.Name, len(def.Workflow))
	}

	// 获取统计信息
	stats := loader.GetStats()
	fmt.Printf("\n📊 加载器统计:\n")
	fmt.Printf("   总技能数: %v\n", stats["total_skills"])
	fmt.Printf("   缓存定义: %v\n", stats["cached_definitions"])

	// ============================================================
	// 步骤6：执行技能
	// ============================================================
	fmt.Println("\n步骤6: 执行技能示例")

	// 准备执行上下文
	execCtx := skills.NewExecutionContext()
	execCtx.Workspace = "/workspace"
	execCtx.SetEnv("ENV", "demo")
	execCtx.SetMetadata("config", map[string]interface{}{
		"cache_disabled": false,
		"force_new":       false,
	})

	// 示例1：执行 HTTP 请求技能
	fmt.Println("\n--- 示例1: HTTP 请求技能 ---")
	httpInput := `{
		"method": "GET",
		"url": "https://api.example.com/users"
	}`

	result, err := host.ExecuteSkill(ctx, "http_request", httpInput, execCtx)
	if err != nil {
		log.Printf("执行 HTTP 请求技能失败: %v\n", err)
	} else {
		fmt.Printf("✅ HTTP 请求技能执行完成\n")
		fmt.Printf("   结果类型: %T\n", result)

		if resultMap, ok := result.(map[string]interface{}); ok {
			for key, value := range resultMap {
				fmt.Printf("   %s: %v\n", key, value)
			}
		}
	}

	// 示例2：执行文件操作技能
	fmt.Println("\n--- 示例2: 文件操作技能 ---")
	fileInput := `{
		"operation": "read",
		"path": "test.txt"
	}`

	result, err = host.ExecuteSkill(ctx, "file_operation", fileInput, execCtx)
	if err != nil {
		log.Printf("执行文件操作技能失败: %v\n", err)
	} else {
		fmt.Printf("✅ 文件操作技能执行完成\n")
		fmt.Printf("   结果类型: %T\n", result)
	}

	// ============================================================
	// 步骤7：将技能转换为工具
	// ============================================================
	fmt.Println("\n步骤7: 将技能转换为 Eino 工具")

	tools, err := host.GetSkillTools(ctx)
	if err != nil {
		log.Printf("获取技能工具失败: %v\n", err)
	} else {
		fmt.Printf("✅ 已将 %d 个技能转换为工具\n", len(tools))
		for i, tool := range tools {
			info, _ := tool.Info(ctx)
			fmt.Printf("   %d. %s - %s\n", i+1, info.Name, info.Desc)
		}
	}

	// ============================================================
	// 步骤8: 查看技能统计
	// ============================================================
	fmt.Println("\n步骤8: 查看技能使用统计")

	allStats := registry.GetStats()
	fmt.Printf("📊 注册表统计:\n")
	fmt.Printf("   总技能数: %v\n", allStats["total_skills"])
	fmt.Printf("   总使用次数: %v\n", allStats["total_uses"])
	fmt.Printf("   分类数: %v\n", allStats["categories"])
	fmt.Printf("   未使用技能: %v\n", allStats["never_used"])

	if mostUsed, ok := allStats["most_used"].(*skills.SkillEntry); ok && mostUsed != nil {
		fmt.Printf("   最常用技能: %s (%d次)\n", mostUsed.Name, mostUsed.UsedCount)
	}

	// ============================================================
	// 步骤9: 导出文档
	// ============================================================
	fmt.Println("\n步骤9: 导出技能文档")

	markdownFile := ".skills/registry/INTEGRATION_DEMO.md"
	file, err := os.Create(markdownFile)
	if err != nil {
		log.Printf("Warning: 创建文档文件失败: %v\n", err)
	} else {
		defer file.Close()

		if err := registry.ExportToMarkdown(file); err != nil {
			log.Printf("Warning: 导出文档失败: %v\n", err)
		} else {
			fmt.Printf("✅ 已导出技能文档: %s\n", markdownFile)
		}
	}

	// ============================================================
	// 总结
	// ============================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 技能系统集成演示完成！")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n核心功能演示：")
	fmt.Println("  ✅ Host 自动初始化技能系统")
	fmt.Println("  ✅ 通过配置文件启用技能系统")
	fmt.Println("  ✅ 访问技能系统的所有组件")
	fmt.Println("  ✅ 执行声明式工作流")
	fmt.Println("  ✅ 渐进式加载技能定义")
	fmt.Println("  ✅ 将技能转换为 Eino 工具")
	fmt.Println("  ✅ 查看使用统计和导出文档")

	fmt.Println("\n集成优势：")
	fmt.Println("  🎯 无缝集成 - 与现有 Agent 系统完美融合")
	fmt.Println("  🔄 自动初始化 - 配置即可使用")
	fmt.Println("  📝 统一接口 - 通过 Host 访问所有功能")
	fmt.Println("  ⚡ 高性能 - 渐进式加载节省 Token")
	fmt.Println("  🔧 易扩展 - 支持自定义技能和模板")

	fmt.Println("\n下一步：")
	fmt.Println("  1. 在 config.yaml 中启用技能系统")
	fmt.Println("  2. 创建自定义技能定义")
	fmt.Println("  3. 将技能集成到 Agent 工作流中")
	fmt.Println("  4. 利用模板库生成代码")

	fmt.Println("\n文档参考：")
	fmt.Println("  - SKILL_COMPARISON_ANALYSIS.md")
	fmt.Println("  - SKILL_ENHANCEMENT_GUIDE.md")
	fmt.Println("  - SKILLS_API_QUICK_REFERENCE.md")
}
