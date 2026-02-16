// Agent Framework - Workflow Commands
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

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "管理工作流",
	Long:  `管理工作流的创建、执行、查询和删除等操作。支持顺序、并行、DAG、路由和规划等多种工作流类型。`,
}

// addWorkflowCommands adds workflow-related commands to root command
func addWorkflowCommands() {
	// List workflows
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有工作流",
		Long:  `列出系统中所有已创建的工作流及其基本信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)
			return svc.ListWorkflowsTable(ctx, outputFormat)
		},
	}
	workflowCmd.AddCommand(listCmd)

	// Get workflow
	getCmd := &cobra.Command{
		Use:   "get [workflow-id]",
		Short: "获取工作流详情",
		Long:  `获取指定工作流的详细信息，包括定义、状态、版本等。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)
			wf, err := svc.GetWorkflow(ctx, args[0])
			if err != nil {
				return err
			}
			return printWorkflowInfo(wf)
		},
	}
	workflowCmd.AddCommand(getCmd)

	// Create workflow
	createCmd := &cobra.Command{
		Use:   "create [name] [description]",
		Short: "创建工作流",
		Long:  `创建一个新的工作流。需要指定工作流名称和描述信息。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)

			name := args[0]
			description := ""
			if len(args) > 1 {
				description = args[1]
			}

			id, err := svc.CreateWorkflow(ctx, name, description)
			if err != nil {
				return fmt.Errorf("failed to create workflow: %w", err)
			}

			fmt.Printf("Workflow created successfully: %s\n", id)
			return nil
		},
	}
	workflowCmd.AddCommand(createCmd)

	// Delete workflow
	deleteCmd := &cobra.Command{
		Use:   "delete [workflow-id]",
		Short: "删除工作流",
		Long:  `删除指定的工作流。此操作不可撤销。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)

			if err := svc.DeleteWorkflow(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete workflow: %w", err)
			}

			fmt.Printf("Workflow deleted successfully: %s\n", args[0])
			return nil
		},
	}
	workflowCmd.AddCommand(deleteCmd)

	// Execute workflow
	executeCmd := &cobra.Command{
		Use:   "execute [workflow-id] [input]",
		Short: "执行工作流",
		Long:  `执行指定的工作流。可以提供输入参数。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)

			input := ""
			if len(args) > 1 {
				input = args[1]
			}

			result, err := svc.ExecuteWorkflow(ctx, args[0], input)
			if err != nil {
				return fmt.Errorf("failed to execute workflow: %w", err)
			}

			fmt.Printf("Workflow execution result:\n%s\n", result)
			return nil
		},
	}
	workflowCmd.AddCommand(executeCmd)

	// List versions
	versionsCmd := &cobra.Command{
		Use:   "versions [workflow-id]",
		Short: "列出工作流版本",
		Long:  `列出指定工作流的所有历史版本。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)

			versions, err := svc.GetWorkflowVersions(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get workflow versions: %w", err)
			}

			fmt.Printf("Versions for workflow %s:\n", args[0])
			for _, v := range versions {
				fmt.Printf("  Version %d: %s\n", v.Version, v.Description)
			}
			return nil
		},
	}
	workflowCmd.AddCommand(versionsCmd)

	rootCmd.AddCommand(workflowCmd)
}

// printWorkflowInfo prints workflow information in a formatted way
func printWorkflowInfo(wf *agent.WorkflowInfo) error {
	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", wf)
	case "table", "":
		fmt.Println("Workflow Information:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("ID: %s\n", wf.ID)
		fmt.Printf("Name: %s\n", wf.Name)
		fmt.Printf("Description: %s\n", wf.Description)
		fmt.Println("────────────────────────────────────────────────────────────")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return nil
}
