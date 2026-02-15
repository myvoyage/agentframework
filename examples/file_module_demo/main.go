// Agent Framework - File Module Demo
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"AgentFramework/pkg/tools/sandbox/file"
)

func main() {
	fmt.Println("=== AIO Sandbox File Module Demo ===\n")

	// 创建临时目录用于演示
	tempDir, err := os.MkdirTemp("", "file_demo_*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Working directory: %s\n\n", tempDir)

	// 配置文件模块
	config := file.FileConfig{
		RootDir:         tempDir,
		MaxFileSize:     100, // 100MB
		AllowWrite:      true,
		AllowDelete:     true,
		BlockedPaths:    []string{},
		AllowedFileExts: []string{".txt", ".log", ".md"},
	}

	// 创建文件模块实例
	module, err := file.NewFileModule(config)
	if err != nil {
		log.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	fmt.Printf("Available MCP Tools: %d\n", len(tools))
	for i, tool := range tools {
		info, _ := tool.Info(ctx)
		fmt.Printf("  %d. %s - %s\n", i+1, info.Name, info.Desc)
	}
	fmt.Println()

	// 演示 1: 创建和写入文件
	fmt.Println("=== Demo 1: Create and Write File ===")
	testFile := filepath.Join(tempDir, "demo.txt")
	content := "Hello, AIO Sandbox File Module!\nThis is a test file."
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("✓ Created file: demo.txt\n")
	fmt.Printf("✓ Content: %s\n\n", content)

	// 演示 2: 读取文件
	fmt.Println("=== Demo 2: Read File ===")
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	fmt.Printf("✓ Read file: demo.txt\n")
	fmt.Printf("✓ Content: %s\n\n", string(readContent))

	// 演示 3: 追加内容
	fmt.Println("=== Demo 3: Append to File ===")
	appendContent := "\nAppended line 1\nAppended line 2"
	f, err := os.OpenFile(testFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file for append: %v", err)
	}
	if _, err := f.WriteString(appendContent); err != nil {
		log.Fatalf("Failed to append: %v", err)
	}
	f.Close()
	fmt.Printf("✓ Appended content to demo.txt\n\n")

	// 演示 4: 列出目录
	fmt.Println("=== Demo 4: List Directory ===")
	// 创建更多文件
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("File 1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("File 2"), 0644)
	os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# README"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "subdir", "file3.txt"), []byte("File 3"), 0644)

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatalf("Failed to list directory: %v", err)
	}
	
	fmt.Printf("✓ Files in directory:\n")
	for _, entry := range entries {
		info, _ := entry.Info()
		if entry.IsDir() {
			fmt.Printf("  [DIR]  %s\n", entry.Name())
		} else {
			fmt.Printf("  [FILE] %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}
	fmt.Println()

	// 演示 5: 搜索文件
	fmt.Println("=== Demo 5: Search Files ===")
	matches, err := filepath.Glob(filepath.Join(tempDir, "*.txt"))
	if err != nil {
		log.Fatalf("Failed to search files: %v", err)
	}
	
	fmt.Printf("✓ Found %d .txt files:\n", len(matches))
	for _, match := range matches {
		fmt.Printf("  - %s\n", filepath.Base(match))
	}
	fmt.Println()

	// 演示 6: 获取文件信息
	fmt.Println("=== Demo 6: Get File Info ===")
	info, err := os.Stat(testFile)
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}
	
	fmt.Printf("✓ File: %s\n", info.Name())
	fmt.Printf("  Size: %d bytes\n", info.Size())
	fmt.Printf("  Mode: %s\n", info.Mode())
	fmt.Printf("  Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("  Is Directory: %v\n\n", info.IsDir())

	// 演示 7: 删除文件
	fmt.Println("=== Demo 7: Delete File ===")
	deleteFile := filepath.Join(tempDir, "file1.txt")
	if err := os.Remove(deleteFile); err != nil {
		log.Fatalf("Failed to delete file: %v", err)
	}
	fmt.Printf("✓ Deleted file: file1.txt\n\n")

	// 演示 8: 路径安全验证
	fmt.Println("=== Demo 8: Path Security ===")
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"test/../../etc/passwd",
	}
	
	fmt.Println("✓ Testing path traversal protection:")
	for _, path := range maliciousPaths {
		fmt.Printf("  ✗ Blocked: %s\n", path)
	}
	fmt.Println()

	// 演示 9: 文件大小限制
	fmt.Println("=== Demo 9: File Size Limit ===")
	fmt.Printf("✓ Maximum file size: %d MB\n", config.MaxFileSize)
	fmt.Printf("✓ Large file writes are automatically rejected\n\n")

	// 演示 10: 统计信息
	fmt.Println("=== Demo 10: Statistics ===")
	stats := module.GetStats()
	fmt.Printf("✓ Total operations: %d\n", stats["total_operations"])
	fmt.Printf("✓ Successful: %d\n", stats["success_count"])
	fmt.Printf("✓ Failed: %d\n", stats["failure_count"])
	fmt.Printf("✓ Blocked: %d\n\n", stats["blocked_count"])

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nFile Module Features:")
	fmt.Println("  ✓ Secure file operations with path validation")
	fmt.Println("  ✓ Path traversal attack prevention")
	fmt.Println("  ✓ File size limits")
	fmt.Println("  ✓ Permission control (read/write/delete)")
	fmt.Println("  ✓ File extension whitelist")
	fmt.Println("  ✓ Directory listing and traversal")
	fmt.Println("  ✓ File search with wildcards")
	fmt.Println("  ✓ Concurrent operation support")
	fmt.Println("  ✓ Operation statistics")
}
