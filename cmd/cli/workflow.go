// Agent Framework - Workflow Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"AgentFramework/core"
	"AgentFramework/agent"
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wf"},
	Short:   "管理工作流",
	Long:    `管理工作流的创建、执行、查询、更新和删除等操作。支持顺序、并行、DAG、路由和规划等多种工作流类型。`,
}

// addWorkflowCommands adds workflow-related commands to root command
func addWorkflowCommands() {
	// ── list ─────────────────────────────────────────────────────────────────
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

	// ── get ──────────────────────────────────────────────────────────────────
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

	// ── create ───────────────────────────────────────────────────────────────
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

			fmt.Printf("✓ Workflow created successfully\n  ID: %s\n  Name: %s\n", id, name)
			return nil
		},
	}
	workflowCmd.AddCommand(createCmd)

	// ── update ───────────────────────────────────────────────────────────────
	updateCmd := &cobra.Command{
		Use:   "update [workflow-id] [name] [description]",
		Short: "更新工作流信息",
		Long:  `更新指定工作流的名称或描述信息。`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			wfID := args[0]
			name := args[1]
			description := ""
			if len(args) > 2 {
				description = args[2]
			}

			wfManager := app.GetWorkflowManager()
			if err := wfManager.UpdateWorkflow(ctx, wfID, name, description, ""); err != nil {
				return fmt.Errorf("failed to update workflow '%s': %w", wfID, err)
			}

			fmt.Printf("✓ Workflow updated: %s\n", wfID)
			return nil
		},
	}
	workflowCmd.AddCommand(updateCmd)

	// ── delete ───────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:     "delete [workflow-id]",
		Aliases: []string{"rm", "remove"},
		Short:   "删除工作流",
		Long:    `删除指定的工作流。此操作不可撤销。`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewWorkflowService(app)

			if err := svc.DeleteWorkflow(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete workflow: %w", err)
			}

			fmt.Printf("✓ Workflow deleted: %s\n", args[0])
			return nil
		},
	}
	workflowCmd.AddCommand(deleteCmd)

	// ── execute ──────────────────────────────────────────────────────────────
	executeCmd := &cobra.Command{
		Use:     "execute [workflow-id] [input]",
		Aliases: []string{"run", "exec"},
		Short:   "执行工作流",
		Long:    `执行指定的工作流。可以提供输入参数（JSON 格式或纯文本）。`,
		Args:    cobra.MinimumNArgs(1),
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

	// ── versions ─────────────────────────────────────────────────────────────
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

			fmt.Printf("Versions for workflow '%s':\n", args[0])
			fmt.Println("────────────────────────────────────────────────────────────")
			if len(versions) == 0 {
				fmt.Println("  (no versions found)")
			}
			for _, v := range versions {
				fmt.Printf("  v%-5d  %s\n", v.Version, v.Description)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	workflowCmd.AddCommand(versionsCmd)

	// ── graph ────────────────────────────────────────────────────────────────
	graphCmd := &cobra.Command{
		Use:   "graph [workflow-id]",
		Short: "显示工作流图结构",
		Long:  `以文本或 JSON 格式显示指定工作流的有向图结构（节点和边）。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wfID := args[0]
			graph, err := app.GetHost().GetWorkflowGraph(wfID)
			if err != nil {
				return fmt.Errorf("failed to get workflow graph for '%s': %w", wfID, err)
			}

			if outputFormat == "json" {
				fmt.Printf("%v\n", graph)
				return nil
			}

			if m, ok := graph.(map[string]interface{}); ok {
				fmt.Printf("Workflow Graph: %s\n", wfID)
				fmt.Println("────────────────────────────────────────────────────────────")
				fmt.Printf("Name: %v\n", m["name"])
				fmt.Printf("Type: %v\n", m["type"])
				if nodes, ok := m["nodes"].([]map[string]interface{}); ok {
					fmt.Printf("Nodes (%d):\n", len(nodes))
					for _, n := range nodes {
						fmt.Printf("  [%v] %v (%v)\n", n["id"], n["name"], n["type"])
					}
				}
				fmt.Println("────────────────────────────────────────────────────────────")
			}
			return nil
		},
	}
	workflowCmd.AddCommand(graphCmd)

	// ── import ───────────────────────────────────────────────────────────────
	importWfCmd := &cobra.Command{
		Use:   "import [json-string-or-file]",
		Short: "从 JSON 导入工作流",
		Long: `从 JSON 字符串或文件导入工作流定义。

JSON 格式示例:
  {"name":"my-workflow","description":"desc","definition":"..."}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			wfManager := app.GetWorkflowManager()

			jsonStr := args[0]

			// 简单解析：尝试提取 name 和 description
			id, err := wfManager.ImportWorkflowFromJSON(ctx, jsonStr)
			if err != nil {
				return fmt.Errorf("failed to import workflow: %w", err)
			}

			fmt.Printf("✓ Workflow imported successfully\n  ID: %s\n", id)
			return nil
		},
	}
	workflowCmd.AddCommand(importWfCmd)

	// ── describe ─────────────────────────────────────────────────────────────
	describeWfCmd := &cobra.Command{
		Use:     "describe [workflow-id]",
		Aliases: []string{"desc"},
		Short:   "详细描述工作流",
		Long:    `显示工作流的详细配置，包括步骤、节点、边等信息。`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wfID := args[0]
			cfg := app.GetHost().Config()

			fmt.Printf("Workflow: %s\n", wfID)
			fmt.Println("────────────────────────────────────────────────────────────")

			for _, spec := range cfg.Workflows {
				if spec.Name == wfID {
					fmt.Printf("Kind:       %s\n", spec.Kind)
					if spec.Model != "" {
						fmt.Printf("Model:      %s\n", spec.Model)
					}
					if len(spec.Steps) > 0 {
						fmt.Printf("Steps:      %v\n", spec.Steps)
					}
					if len(spec.Agents) > 0 {
						fmt.Printf("Agents:     %v\n", spec.Agents)
					}
					if spec.Aggregator != "" {
						fmt.Printf("Aggregator: %s\n", spec.Aggregator)
					}
					if len(spec.Routes) > 0 {
						fmt.Println("Routes:")
						for k, v := range spec.Routes {
							fmt.Printf("  %s -> %s\n", k, v)
						}
					}
					if len(spec.Nodes) > 0 {
						fmt.Printf("Nodes (%d):\n", len(spec.Nodes))
						for _, n := range spec.Nodes {
							fmt.Printf("  [%s] kind=%s, agent=%s\n", n.ID, n.Kind, n.AgentName)
						}
					}
					if len(spec.Edges) > 0 {
						fmt.Println("Edges:")
						for from, to := range spec.Edges {
							fmt.Printf("  %s -> %s\n", from, to)
						}
					}
					fmt.Println("────────────────────────────────────────────────────────────")
					return nil
				}
			}

			// 如果配置中没有，查找运行时
			_, err := app.GetHost().GetWorkflow(wfID)
			if err != nil {
				return fmt.Errorf("workflow '%s' not found", wfID)
			}
			fmt.Println("(workflow exists in runtime but has no static spec)")
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	workflowCmd.AddCommand(describeWfCmd)

	rootCmd.AddCommand(workflowCmd)
}

// printWorkflowInfo prints workflow information in a formatted way
func printWorkflowInfo(wf *agent.WorkflowInfo) error {
	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", wf)
	case "yaml":
		fmt.Printf("id: %s\nname: %s\ndescription: %s\n", wf.ID, wf.Name, wf.Description)
	default:
		fmt.Println("Workflow Information:")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Printf("ID:          %s\n", wf.ID)
		fmt.Printf("Name:        %s\n", wf.Name)
		fmt.Printf("Description: %s\n", wf.Description)
		fmt.Println("────────────────────────────────────────────────────────────")
	}
	return nil
}
