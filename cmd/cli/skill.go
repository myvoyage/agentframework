// Agent Framework - Skill Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmdcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"AgentFramework/core"
	"AgentFramework/agent"
)

// skillCmd represents the skill command
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "管理技能系统",
	Long:  `管理技能的查询、执行、导入和导出等操作。技能是扩展 Agent 功能的可复用单元。`,
}

// addSkillCommands adds skill-related commands to root command
func addSkillCommands() {
	// List skills
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有技能",
		Long:  `列出系统中所有已注册的技能及其信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewSkillService(app)
			return svc.ListSkillsTable(ctx, outputFormat)
		},
	}
	skillCmd.AddCommand(listCmd)

	// Get skill
	getCmd := &cobra.Command{
		Use:   "get [skill-name]",
		Short: "获取技能详情",
		Long:  `获取指定技能的详细信息，包括描述、版本、类别等。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewSkillService(app)

			skill, err := svc.GetSkill(ctx, args[0])
			if err != nil {
				return err
			}
			return printSkillInfo(skill)
		},
	}
	skillCmd.AddCommand(getCmd)

	// Execute skill
	executeCmd := &cobra.Command{
		Use:   "execute [skill-name] [input]",
		Short: "执行技能",
		Long:  `执行指定的技能。可以提供输入参数。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewSkillService(app)

			input := ""
			if len(args) > 1 {
				input = args[1]
			}

			result, err := svc.ExecuteSkill(ctx, &core.ExecuteSkillInput{
				SkillName: args[0],
				Input:      input,
			})
			if err != nil {
				return err
			}

			if !result.Success {
				return fmt.Errorf("skill execution failed: %s", result.Error)
			}

			fmt.Printf("Skill execution result:\n%+v\n", result.Result)
			return nil
		},
	}
	skillCmd.AddCommand(executeCmd)

	// Skill system info
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "技能系统信息",
		Long:  `显示技能系统的基本信息和统计。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewSkillService(app)

			info, err := svc.GetSkillSystemInfo(ctx)
			if err != nil {
				return err
			}

			fmt.Println("Skill System Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Initialized: %v\n", info.Initialized)
			fmt.Printf("Base Directory: %s\n", info.BaseDir)
			fmt.Printf("Total Skills: %d\n", info.TotalSkills)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	skillCmd.AddCommand(infoCmd)

	rootCmd.AddCommand(skillCmd)
}

// printSkillInfo prints skill information in a formatted way
func printSkillInfo(skill agent.SkillMetadata) error {
	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", skill)
	case "table", "":
		fmt.Println("Skill Information:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("Name: %s\n", skill.Name)
		fmt.Printf("Description: %s\n", skill.Description)
		fmt.Printf("Version: %s\n", skill.Version)
		fmt.Printf("Category: %s\n", skill.Category)
		fmt.Printf("Author: %s\n", skill.Author)
		fmt.Println("────────────────────────────────────────────────────────────")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return nil
}
