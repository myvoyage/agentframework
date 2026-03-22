// Agent Framework - Skill Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// skillCmd represents the skill command
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "管理技能",
	Long:  `管理技能的注册、启用、禁用、查询和执行等操作。技能是 Agent 的可插拔功能单元。`,
}

// addSkillCommands adds skill-related commands to root command
func addSkillCommands() {
	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有技能",
		Long:  `列出系统中所有已注册的技能及其状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			skills := app.GetSkillLibrary().GetAllSkills(ctx)

			if outputFormat == "json" {
				type skillJSON struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Version     string `json:"version"`
					Enabled     bool   `json:"enabled"`
				}
				list := make([]skillJSON, 0, len(skills))
				for name, skill := range skills {
					metadata := skill.GetMetadata(ctx)
					list = append(list, skillJSON{
						ID:          name,
						Name:        metadata.Name,
						Description: metadata.Description,
						Version:     metadata.Version,
						Enabled:     skill.IsEnabled(ctx),
					})
				}
				b, _ := json.MarshalIndent(list, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(skills) == 0 {
				fmt.Println("No skills available")
				return nil
			}

			fmt.Println("Available Skills:")
			fmt.Println("────────────────────────────────────────────────────────────")
			enabledCount := 0
			for name, skill := range skills {
				metadata := skill.GetMetadata(ctx)
				status := "✓"
				if !skill.IsEnabled(ctx) {
					status = "○"
				} else {
					enabledCount++
				}
				fmt.Printf("%s %-30s v%-8s\n", status, metadata.Name, metadata.Version)
				fmt.Printf("  ID: %-25s  %s\n", name, metadata.Description)
				fmt.Println("  ─────────────────────────────────────────────────────")
			}
			fmt.Printf("Total: %d skill(s), %d enabled\n", len(skills), enabledCount)
			return nil
		},
	}
	skillCmd.AddCommand(listCmd)

	// ── enable ───────────────────────────────────────────────────────────────
	enableCmd := &cobra.Command{
		Use:   "enable [skill-id]",
		Short: "启用技能",
		Long:  `启用指定的技能，使其可以被 Agent 使用。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			for _, skillID := range args {
				if err := app.GetSkillLibrary().EnableSkill(ctx, skillID); err != nil {
					return fmt.Errorf("failed to enable skill '%s': %w", skillID, err)
				}
				fmt.Printf("✓ Skill '%s' enabled\n", skillID)
			}
			return nil
		},
	}
	skillCmd.AddCommand(enableCmd)

	// ── disable ──────────────────────────────────────────────────────────────
	disableCmd := &cobra.Command{
		Use:   "disable [skill-id]",
		Short: "禁用技能",
		Long:  `禁用指定的技能，使其无法被 Agent 使用。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			for _, skillID := range args {
				if err := app.GetSkillLibrary().DisableSkill(ctx, skillID); err != nil {
					return fmt.Errorf("failed to disable skill '%s': %w", skillID, err)
				}
				fmt.Printf("○ Skill '%s' disabled\n", skillID)
			}
			return nil
		},
	}
	skillCmd.AddCommand(disableCmd)

	// ── toggle ───────────────────────────────────────────────────────────────
	toggleCmd := &cobra.Command{
		Use:   "toggle [skill-id]",
		Short: "切换技能启用/禁用状态",
		Long:  `切换指定技能的启用/禁用状态。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			skillID := args[0]

			skill, found := app.GetSkillLibrary().GetSkill(ctx, skillID)
			if !found {
				return fmt.Errorf("skill '%s' not found", skillID)
			}

			if skill.IsEnabled(ctx) {
				if err := app.GetSkillLibrary().DisableSkill(ctx, skillID); err != nil {
					return fmt.Errorf("failed to disable skill: %w", err)
				}
				fmt.Printf("○ Skill '%s' disabled\n", skillID)
			} else {
				if err := app.GetSkillLibrary().EnableSkill(ctx, skillID); err != nil {
					return fmt.Errorf("failed to enable skill: %w", err)
				}
				fmt.Printf("✓ Skill '%s' enabled\n", skillID)
			}
			return nil
		},
	}
	skillCmd.AddCommand(toggleCmd)

	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info [skill-id]",
		Short: "获取技能详情",
		Long:  `获取指定技能的详细信息，包括名称、描述、版本和当前状态。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			skill, found := app.GetSkillLibrary().GetSkill(ctx, args[0])
			if !found {
				return fmt.Errorf("skill '%s' not found", args[0])
			}

			metadata := skill.GetMetadata(ctx)

			if outputFormat == "json" {
				info := map[string]interface{}{
					"id":          args[0],
					"name":        metadata.Name,
					"description": metadata.Description,
					"version":     metadata.Version,
					"enabled":     skill.IsEnabled(ctx),
				}
				b, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Skill Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("ID:          %s\n", args[0])
			fmt.Printf("Name:        %s\n", metadata.Name)
			fmt.Printf("Description: %s\n", metadata.Description)
			fmt.Printf("Version:     %s\n", metadata.Version)
			status := "Enabled ✓"
			if !skill.IsEnabled(ctx) {
				status = "Disabled ○"
			}
			fmt.Printf("Status:      %s\n", status)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	skillCmd.AddCommand(infoCmd)

	// ── run ──────────────────────────────────────────────────────────────────
	runCmd := &cobra.Command{
		Use:   "run [skill-id] [input-json]",
		Short: "直接执行技能",
		Long:  `直接执行指定的技能，不需要通过 Agent。输入为 JSON 格式或字符串。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			skill, found := app.GetSkillLibrary().GetSkill(ctx, args[0])
			if !found {
				return fmt.Errorf("skill '%s' not found", args[0])
			}

			if !skill.IsEnabled(ctx) {
				return fmt.Errorf("skill '%s' is disabled. Enable it first with: af skill enable %s", args[0], args[0])
			}

			input := ""
			if len(args) > 1 {
				input = args[1]
			}

			result, err := skill.Invoke(ctx, input)
			if err != nil {
				return fmt.Errorf("skill execution failed: %w", err)
			}

			if outputFormat == "json" {
				out := map[string]string{"skill_id": args[0], "result": result}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Printf("Skill Result:\n%s\n", result)
			return nil
		},
	}
	skillCmd.AddCommand(runCmd)

	// ── delete ───────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:     "delete [skill-id]",
		Aliases: []string{"rm", "unregister"},
		Short:   "注销/删除技能",
		Long:    `从系统中注销指定的技能。`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			for _, skillID := range args {
				if err := app.GetSkillLibrary().UnregisterSkill(ctx, skillID); err != nil {
					return fmt.Errorf("failed to delete skill '%s': %w", skillID, err)
				}
				fmt.Printf("✓ Skill '%s' deleted\n", skillID)
			}
			return nil
		},
	}
	skillCmd.AddCommand(deleteCmd)

	rootCmd.AddCommand(skillCmd)
}
