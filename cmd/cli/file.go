// Agent Framework - File Commands
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
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "文件系统操作",
	Long:  `执行文件系统操作，包括列出、读取、写入、删除文件和目录。`,
}

// addFileCommands adds file-related commands to root command
func addFileCommands() {
	// List files
	listCmd := &cobra.Command{
		Use:   "ls [path]",
		Short: "列出文件",
		Long:  `列出指定目录下的文件和子目录。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewFileService(app)

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			return svc.ListFilesTable(ctx, path, outputFormat)
		},
	}
	fileCmd.AddCommand(listCmd)

	// Read file
	readCmd := &cobra.Command{
		Use:   "read [path]",
		Short: "读取文件",
		Long:  `读取指定文件的内容并输出到标准输出。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewFileService(app)

			content, err := svc.ReadFile(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			fmt.Println(content)
			return nil
		},
	}
	fileCmd.AddCommand(readCmd)

	// Write file
	writeCmd := &cobra.Command{
		Use:   "write [path] [content]",
		Short: "写入文件",
		Long:  `将内容写入指定文件。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewFileService(app)

			content := ""
			if len(args) > 1 {
				content = args[1]
			}

			if err := svc.WriteFile(ctx, args[0], content); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("File written successfully: %s\n", args[0])
			return nil
		},
	}
	fileCmd.AddCommand(writeCmd)

	// Delete file
	deleteCmd := &cobra.Command{
		Use:   "delete [path]",
		Short: "删除文件",
		Long:  `删除指定的文件或目录。此操作不可撤销。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewFileService(app)

			if err := svc.DeleteFile(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete file: %w", err)
			}

			fmt.Printf("File deleted successfully: %s\n", args[0])
			return nil
		},
	}
	fileCmd.AddCommand(deleteCmd)

	// Create directory
	mkdirCmd := &cobra.Command{
		Use:   "mkdir [path]",
		Short: "创建目录",
		Long:  `创建指定路径的目录。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewFileService(app)

			if err := svc.CreateDirectory(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			fmt.Printf("Directory created successfully: %s\n", args[0])
			return nil
		},
	}
	fileCmd.AddCommand(mkdirCmd)

	rootCmd.AddCommand(fileCmd)
}
