// Agent Framework - Enhanced Skill Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmdcli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"AgentFramework/agent/skills"
)

// enhancedSkillCmd represents the enhanced skill command
var enhancedSkillCmd = &cobra.Command{
	Use:   "enhanced-skill",
	Short: "管理增强技能系统",
	Long: `管理增强技能的查询、执行、安装等操作。

增强技能系统支持：
- Markdown + YAML 格式的技能定义
- 依赖检查（二进制、环境变量）
- 多种执行器（Shell、HTTP、Workflow）
- 触发器系统（命令、关键词、模式）
- 变量替换和模板`,
}

var (
	skillsDir string
)

// addEnhancedSkillCommands adds enhanced skill commands
func addEnhancedSkillCommands() {
	// Persistent flags
	enhancedSkillCmd.PersistentFlags().StringVarP(&skillsDir, "dir", "d", ".skills", "技能目录")

	// List skills
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有增强技能",
		Long:  `列出目录中所有已加载的增强技能及其详细信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listEnhancedSkills(cmd.Context())
		},
	}
	enhancedSkillCmd.AddCommand(listCmd)

	// Search skills by trigger
	searchCmd := &cobra.Command{
		Use:   "search [trigger-type] [pattern]",
		Short: "根据触发器搜索技能",
		Long:  `根据触发器类型和模式搜索匹配的技能。支持命令、关键词、模式等触发器类型。`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return searchSkillsByTrigger(cmd.Context(), args[0], args[1])
		},
	}
	enhancedSkillCmd.AddCommand(searchCmd)

	// Install skill from GitHub
	installCmd := &cobra.Command{
		Use:   "install [github-repo]",
		Short: "从 GitHub 安装技能",
		Long:  `从指定的 GitHub 仓库安装技能。例如: agentframework/skill-hello-world`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installSkillFromGitHub(cmd.Context(), args[0])
		},
	}
	enhancedSkillCmd.AddCommand(installCmd)

	// Execute skill action
	var executeVars string
	executeCmd := &cobra.Command{
		Use:   "execute [skill-id] [action-id]",
		Short: "执行技能动作",
		Long:  `执行指定技能的动作。可以通过 --vars 参数传递变量。`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeSkillAction(cmd.Context(), args[0], args[1], executeVars)
		},
	}
	executeCmd.Flags().StringVar(&executeVars, "vars", "", "变量列表，格式为 key=value,key2=value2")
	enhancedSkillCmd.AddCommand(executeCmd)

	// Check dependencies
	checkCmd := &cobra.Command{
		Use:   "check [skill-id]",
		Short: "检查技能依赖",
		Long:  `检查指定技能的依赖是否满足，并提供安装提示。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkSkillDependencies(cmd.Context(), args[0])
		},
	}
	enhancedSkillCmd.AddCommand(checkCmd)

	// Get skill detail
	getCmd := &cobra.Command{
		Use:   "get [skill-id]",
		Short: "获取技能详情",
		Long:  `获取指定增强技能的详细信息，包括触发器、动作和依赖。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getEnhancedSkillDetail(cmd.Context(), args[0])
		},
	}
	enhancedSkillCmd.AddCommand(getCmd)

	rootCmd.AddCommand(enhancedSkillCmd)
}

// listEnhancedSkills lists all enhanced skills
func listEnhancedSkills(ctx context.Context) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	// 加载技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载技能失败: %v\n", err)
	}

	// 获取技能列表
	skillList := registry.ListSkills()

	if len(skillList) == 0 {
		fmt.Println("没有找到技能")
		return nil
	}

	// 使用 tabwriter 格式化输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\t名称\t版本\t分类\t描述")
	fmt.Fprintln(w, "──\t──\t──\t──\t──")

	for _, skill := range skillList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			skill.ID,
			skill.Name,
			skill.Version,
			skill.Category,
			truncateString(skill.Description, 50))
	}

	w.Flush()
	fmt.Printf("\n总计: %d 个技能\n", len(skillList))
	return nil
}

// searchSkillsByTrigger searches skills by trigger
func searchSkillsByTrigger(ctx context.Context, triggerType, pattern string) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	// 加载技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载技能失败: %v\n", err)
	}

	// 搜索技能
	foundSkills, err := registry.FindByTrigger(triggerType, pattern)
	if err != nil {
		return fmt.Errorf("搜索失败: %w", err)
	}

	if len(foundSkills) == 0 {
		fmt.Printf("没有找到匹配 '%s' 触发器的技能\n", pattern)
		return nil
	}

	fmt.Printf("找到 %d 个匹配的技能:\n\n", len(foundSkills))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\t名称\t触发器类型\t模式\t优先级")
	fmt.Fprintln(w, "──\t──\t──\t──\t──")

	for _, skill := range foundSkills {
		for _, trigger := range skill.Triggers {
			if trigger.Type == triggerType {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
					skill.ID,
					skill.Name,
					trigger.Type,
					trigger.Pattern,
					trigger.Priority)
			}
		}
	}

	w.Flush()
	return nil
}

// installSkillFromGitHub installs a skill from GitHub
func installSkillFromGitHub(ctx context.Context, repo string) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	fmt.Printf("正在从 GitHub 安装技能: %s\n", repo)

	if err := registry.InstallFromGitHub(ctx, repo); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	fmt.Printf("✓ 技能 %s 安装成功\n", repo)

	// 尝试加载新安装的技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载新技能失败: %v\n", err)
	}

	return nil
}

// executeSkillAction executes a skill action
func executeSkillAction(ctx context.Context, skillID, actionID, varsStr string) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	// 加载技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载技能失败: %v\n", err)
	}

	// 解析变量
	vars := make(map[string]string)
	if varsStr != "" {
		pairs := strings.Split(varsStr, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				vars[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// 执行技能
	fmt.Printf("正在执行技能: %s (动作: %s)\n", skillID, actionID)
	if len(vars) > 0 {
		fmt.Printf("变量: %v\n", vars)
	}
	fmt.Println()

	output, err := registry.ExecuteSkill(ctx, skillID, actionID, vars)
	if err != nil {
		return fmt.Errorf("执行失败: %w", err)
	}

	fmt.Println("执行结果:")
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Println(output)
	fmt.Println("────────────────────────────────────────────────────────────")
	return nil
}

// checkSkillDependencies checks skill dependencies
func checkSkillDependencies(ctx context.Context, skillID string) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	// 加载技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载技能失败: %v\n", err)
	}

	// 获取技能
	skill, ok := registry.GetSkill(skillID)
	if !ok {
		return fmt.Errorf("技能不存在: %s", skillID)
	}

	fmt.Printf("检查技能: %s (%s)\n\n", skill.Name, skill.ID)

	if skill.Prerequisites == nil {
		fmt.Println("✓ 该技能没有依赖要求")
		return nil
	}

	// 检查依赖
	checker := skills.NewDependencyChecker()
	result, err := checker.Check(ctx, skill.Prerequisites)
	if err != nil {
		return fmt.Errorf("依赖检查失败: %w", err)
	}

	if result.Satisfied {
		fmt.Println("✓ 所有依赖已满足")
	} else {
		fmt.Println("✗ 依赖不满足:")
		fmt.Println()
		for _, missing := range result.Missing {
			fmt.Printf("  • 缺失: %s\n", missing)
		}
		fmt.Println()
		if len(result.InstallHints) > 0 {
			fmt.Println("安装建议:")
			for _, hint := range result.InstallHints {
				fmt.Printf("  • %s:\n", hint.Label)
				fmt.Printf("    %s\n", hint.Command)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("警告:")
		for _, warning := range result.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	return nil
}

// getEnhancedSkillDetail gets enhanced skill detail
func getEnhancedSkillDetail(ctx context.Context, skillID string) error {
	registry, err := skills.NewEnhancedSkillRegistry(skillsDir)
	if err != nil {
		return fmt.Errorf("创建注册表失败: %w", err)
	}
	defer registry.Close()

	// 加载技能
	if err := registry.LoadSkillFromDirectory(skillsDir); err != nil {
		fmt.Printf("警告: 加载技能失败: %v\n", err)
	}

	// 获取技能
	skill, ok := registry.GetSkill(skillID)
	if !ok {
		return fmt.Errorf("技能不存在: %s", skillID)
	}

	// 显示详细信息
	fmt.Println("技能详情")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Printf("ID:         %s\n", skill.ID)
	fmt.Printf("名称:       %s\n", skill.Name)
	fmt.Printf("版本:       %s\n", skill.Version)
	fmt.Printf("分类:       %s\n", skill.Category)
	fmt.Printf("描述:       %s\n", skill.Description)
	fmt.Printf("作者:       %s\n", skill.Author)
	fmt.Printf("许可证:     %s\n", skill.License)
	fmt.Printf("来源文件:   %s\n", skill.SourceFile)
	fmt.Printf("加载时间:   %s\n", skill.LoadedAt.Format("2006-01-02 15:04:05"))

	// 显示触发器
	if len(skill.Triggers) > 0 {
		fmt.Println("\n触发器:")
		for _, trigger := range skill.Triggers {
			fmt.Printf("  • 类型: %s, 模式: %s, 优先级: %d\n",
				trigger.Type, trigger.Pattern, trigger.Priority)
		}
	}

	// 显示动作
	if len(skill.Actions) > 0 {
		fmt.Println("\n动作:")
		for _, action := range skill.Actions {
			fmt.Printf("  • ID: %s\n", action.ID)
			fmt.Printf("    类型: %s\n", action.Type)
			fmt.Printf("    描述: %s\n", action.Description)
			if action.Timeout > 0 {
				fmt.Printf("    超时: %s\n", action.Timeout)
			}
		}
	}

	// 显示依赖
	if skill.Prerequisites != nil {
		fmt.Println("\n前置条件:")

		if len(skill.Prerequisites.Bins) > 0 {
			fmt.Println("  二进制依赖:")
			for _, bin := range skill.Prerequisites.Bins {
				fmt.Printf("    • %s", bin.Name)
				if bin.Version != "" {
					fmt.Printf(" (版本: %s)", bin.Version)
				}
				fmt.Println()

				if len(bin.Install) > 0 {
					fmt.Println("      安装方法:")
					for pm, cmd := range bin.Install {
						fmt.Printf("        - %s: %s\n", pm, cmd)
					}
				}
			}
		}

		if len(skill.Prerequisites.Env) > 0 {
			fmt.Println("  环境变量:")
			for _, env := range skill.Prerequisites.Env {
				opt := ""
				if env.Optional {
					opt = " (可选)"
				}
				fmt.Printf("    • %s%s: %s\n", env.Name, opt, env.Description)
			}
		}
	}

	// 显示配置
	fmt.Println("\n配置:")
	fmt.Printf("  最大输出大小: %d bytes\n", skill.Config.MaxOutputSize)
	if skill.Config.MaxExecutionTime > 0 {
		fmt.Printf("  最大执行时间: %s\n", skill.Config.MaxExecutionTime)
	}
	fmt.Printf("  启用缓存: %v\n", skill.Config.EnableCache)
	if skill.Config.EnableCache {
		fmt.Printf("  缓存过期时间: %s\n", skill.Config.CacheTTL)
	}
	fmt.Printf("  始终加载: %v\n", skill.Always)

	fmt.Println("════════════════════════════════════════════════════════════════")
	return nil
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
