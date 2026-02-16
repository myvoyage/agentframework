// Agent Framework - Enhanced Skills Test
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"AgentFramework/agent/skills"
)

func main() {
	ctx := context.Background()

	// 创建增强的技能注册表
	fmt.Println("=== 创建增强技能注册表 ===")
	registry, err := skills.NewEnhancedSkillRegistry(".skills")
	if err != nil {
		log.Fatalf("创建注册表失败: %v", err)
	}
	defer registry.Close()

	// 从目录加载技能
	fmt.Println("\n=== 加载技能 ===")
	if err := registry.LoadSkillFromDirectory(".skills"); err != nil {
		log.Printf("加载技能失败: %v", err)
	}

	// 列出所有技能
	fmt.Println("\n=== 列出所有技能 ===")
	skillList := registry.ListSkills()
	if len(skillList) == 0 {
		fmt.Println("没有找到技能")
		return
	}

	for _, skill := range skillList {
		fmt.Printf("\n技能: %s\n", skill.Name)
		fmt.Printf("  ID: %s\n", skill.ID)
		fmt.Printf("  版本: %s\n", skill.Version)
		fmt.Printf("  分类: %s\n", skill.Category)
		fmt.Printf("  描述: %s\n", skill.Description)
		fmt.Printf("  作者: %s\n", skill.Author)

		if len(skill.Triggers) > 0 {
			fmt.Printf("  触发器:\n")
			for _, trigger := range skill.Triggers {
				fmt.Printf("    - 类型: %s, 模式: %s, 优先级: %d\n",
					trigger.Type, trigger.Pattern, trigger.Priority)
			}
		}

		if len(skill.Actions) > 0 {
			fmt.Printf("  动作:\n")
			for _, action := range skill.Actions {
				fmt.Printf("    - ID: %s, 类型: %s, 描述: %s\n",
					action.ID, action.Type, action.Description)
			}
		}

		if skill.Prerequisites != nil {
			fmt.Printf("  前置条件:\n")

			if len(skill.Prerequisites.Bins) > 0 {
				fmt.Printf("    二进制依赖:\n")
				for _, bin := range skill.Prerequisites.Bins {
					fmt.Printf("      - %s (版本: %s)\n", bin.Name, bin.Version)
					for pm, cmd := range bin.Install {
						fmt.Printf("        安装 (%s): %s\n", pm, cmd)
					}
				}
			}

			if len(skill.Prerequisites.Env) > 0 {
				fmt.Printf("    环境变量:\n")
				for _, env := range skill.Prerequisites.Env {
					opt := ""
					if env.Optional {
						opt = " (可选)"
					}
					fmt.Printf("      - %s%s: %s\n", env.Name, opt, env.Description)
				}
			}
		}
	}

	// 测试触发器查找
	fmt.Println("\n=== 测试触发器查找 ===")
	testFindByTrigger(registry)

	// 测试依赖检查
	fmt.Println("\n=== 测试依赖检查 ===")
	testDependencyCheck(ctx, registry)

	// 测试技能执行（如果有 gh 命令）
	fmt.Println("\n=== 测试技能执行 ===")
	testSkillExecution(ctx, registry)
}

func testFindByTrigger(registry *skills.EnhancedSkillRegistry) {
	// 测试命令触发器
	foundSkills, err := registry.FindByTrigger("command", "/pr")
	if err != nil {
		log.Printf("查找触发器失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个匹配 '/pr' 命令的技能\n", len(foundSkills))
	for _, skill := range foundSkills {
		fmt.Printf("  - %s (%s)\n", skill.Name, skill.ID)
	}

	// 测试关键词触发器
	foundSkills, err = registry.FindByTrigger("keyword", "github")
	if err != nil {
		log.Printf("查找触发器失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个匹配 'github' 关键词的技能\n", len(foundSkills))
	for _, skill := range foundSkills {
		fmt.Printf("  - %s (%s)\n", skill.Name, skill.ID)
	}
}

func testDependencyCheck(ctx context.Context, registry *skills.EnhancedSkillRegistry) {
	skillList := registry.ListSkills()

	for _, skill := range skillList {
		if skill.Prerequisites == nil {
			continue
		}

		fmt.Printf("\n检查 %s 的依赖:\n", skill.Name)

		// 获取依赖检查器
		checker := skills.NewDependencyChecker()

		result, err := checker.Check(ctx, skill.Prerequisites)
		if err != nil {
			log.Printf("  依赖检查失败: %v", err)
			continue
		}

		if result.Satisfied {
			fmt.Printf("  ✓ 所有依赖已满足\n")
		} else {
			fmt.Printf("  ✗ 依赖不满足:\n")
			for _, missing := range result.Missing {
				fmt.Printf("    - 缺失: %s\n", missing)
			}
			for _, hint := range result.InstallHints {
				fmt.Printf("    - 安装提示: %s\n", hint.Label)
				fmt.Printf("      命令: %s\n", hint.Command)
			}
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("  ! 警告:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("    - %s\n", warning)
			}
		}
	}
}

func testSkillExecution(ctx context.Context, registry *skills.EnhancedSkillRegistry) {
	// 检查是否有 github_pr 技能
	skill, ok := registry.GetSkill("com.agentframework.skill.github_pr")
	if !ok {
		fmt.Println("未找到 github_pr 技能，跳过执行测试")
		return
	}

	// 检查依赖是否满足
	checker := skills.NewDependencyChecker()
	result, err := checker.Check(ctx, skill.Prerequisites)
	if err != nil {
		log.Printf("依赖检查失败: %v", err)
		return
	}

	if !result.Satisfied {
		fmt.Println("依赖不满足，跳过执行测试")
		fmt.Println("请先安装必要的依赖：")
		for _, hint := range result.InstallHints {
			fmt.Printf("  - %s: %s\n", hint.Label, hint.Command)
		}
		return
	}

	// 执行 list_prs 动作
	fmt.Println("\n执行 list_prs 动作...")

	// 准备变量
	repo := os.Getenv("TEST_GITHUB_REPO")
	if repo == "" {
		repo = "cli/cli"
	}

	vars := map[string]string{
		"Repo":  repo,
		"State": "open",
		"Limit": "5",
	}

	output, err := registry.ExecuteSkill(ctx, skill.ID, "list_prs", vars)
	if err != nil {
		log.Printf("执行失败: %v", err)
		return
	}

	fmt.Println("执行结果:")
	fmt.Println(output)
}
