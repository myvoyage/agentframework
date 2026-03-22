// Agent Framework - File Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"AgentFramework/agent"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "文件操作",
	Long:  `提供文件浏览、读取、写入等操作功能。`,
}

// addFileCommands adds file-related commands to root command
func addFileCommands() {
	// List files
	listCmd := &cobra.Command{
		Use:   "list [path]",
		Short: "列出目录内容",
		Long:  `列出指定目录的文件和子目录。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			files, err := app.GetFileExplorer().ListFiles(ctx, path)
			if err != nil {
				return fmt.Errorf("failed to list files: %w", err)
			}

			fmt.Printf("Contents of %s:\n", path)
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, file := range files {
				if file.Type == agent.FileTypeDirectory {
					fmt.Printf("📁 %s\n", file.Name)
				} else {
					fmt.Printf("📄 %s (%d bytes)\n", file.Name, file.Size)
				}
			}
			fmt.Println("────────────────────────────────────────────────────────────")

			return nil
		},
	}
	fileCmd.AddCommand(listCmd)

	// Read file
	readCmd := &cobra.Command{
		Use:   "read [path]",
		Short: "读取文件内容",
		Long:  `读取指定文件的内容并显示。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			content, err := app.GetFileExplorer().ReadFile(ctx, args[0])
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
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			path := args[0]
			content := args[1]

			if err := app.GetFileExplorer().WriteFile(ctx, path, content); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("File written successfully: %s\n", path)
			return nil
		},
	}
	fileCmd.AddCommand(writeCmd)

	// Delete file
	deleteCmd := &cobra.Command{
		Use:   "delete [path]",
		Short: "删除文件",
		Long:  `删除指定的文件或目录。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			if err := app.GetFileExplorer().DeleteFile(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete file: %w", err)
			}

			fmt.Printf("File deleted successfully: %s\n", args[0])
			return nil
		},
	}
	fileCmd.AddCommand(deleteCmd)

	// Copy file
	copyCmd := &cobra.Command{
		Use:   "copy [src] [dst]",
		Short: "复制文件",
		Long:  `复制文件到目标位置。`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()

			if err := app.GetFileExplorer().CopyFile(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

			fmt.Printf("File copied successfully: %s -> %s\n", args[0], args[1])
			return nil
		},
	}
	fileCmd.AddCommand(copyCmd)

	rootCmd.AddCommand(fileCmd)
}
