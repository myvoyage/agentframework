// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
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

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestNewFileModule 测试创建文件模块
func TestNewFileModule(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	if module.config.RootDir == "" {
		t.Error("Root directory not set")
	}

	if module.manager == nil {
		t.Error("File manager not initialized")
	}

	if module.validator == nil {
		t.Error("Path validator not initialized")
	}
}

// TestPathValidator 测试路径验证器
func TestPathValidator(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{filepath.Join(tempDir, "blocked")},
	}

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "Valid path",
			path:      filepath.Join(tempDir, "test.txt"),
			shouldErr: false,
		},
		{
			name:      "Path traversal with ..",
			path:      filepath.Join(tempDir, "..", "etc", "passwd"),
			shouldErr: true,
		},
		{
			name:      "Blocked path",
			path:      filepath.Join(tempDir, "blocked", "file.txt"),
			shouldErr: true,
		},
		{
			name:      "Path outside root",
			path:      "/etc/passwd",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.path)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for path %s, got nil", tt.path)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for path %s: %v", tt.path, err)
			}
		})
	}
}

// TestFileRead 测试文件读取
func TestFileRead(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 读取文件
	result, err := module.readFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Read operation failed: %v", result["error"])
	}

	if result["content"].(string) != testContent {
		t.Errorf("Expected content %s, got %s", testContent, result["content"])
	}
}

// TestFileWrite 测试文件写入
func TestFileWrite(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 写入文件
	testContent := "Test content"
	result, err := module.writeFile("test.txt", testContent, false)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Write operation failed: %v", result["error"])
	}

	// 验证文件内容
	content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %s, got %s", testContent, string(content))
	}
}

// TestFileAppend 测试文件追加
func TestFileAppend(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 写入初始内容
	initialContent := "Hello"
	_, err = module.writeFile("test.txt", initialContent, false)
	if err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	// 追加内容
	appendContent := " World"
	result, err := module.writeFile("test.txt", appendContent, true)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Append operation failed: %v", result["error"])
	}

	// 验证文件内容
	content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read appended file: %v", err)
	}

	expected := initialContent + appendContent
	if string(content) != expected {
		t.Errorf("Expected content %s, got %s", expected, string(content))
	}
}

// TestFileDelete 测试文件删除
func TestFileDelete(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 删除文件
	result, err := module.deleteFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Delete operation failed: %v", result["error"])
	}

	// 验证文件已删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File still exists after deletion")
	}
}

// TestFileList 测试文件列表
func TestFileList(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件和目录
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "subdir", "file3.txt"), []byte("test"), 0644)

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 非递归列表
	result, err := module.listFiles(".", false)
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("List operation failed: %v", result["error"])
	}

	files := result["files"].([]map[string]any)
	dirs := result["directories"].([]map[string]any)

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	if len(dirs) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(dirs))
	}

	// 递归列表
	result, err = module.listFiles(".", true)
	if err != nil {
		t.Fatalf("Failed to list files recursively: %v", err)
	}

	totalCount := result["total_count"].(int)
	if totalCount != 4 { // 2 files + 1 dir + 1 file in subdir
		t.Errorf("Expected 4 total items, got %d", totalCount)
	}
}

// TestFileSearch 测试文件搜索
func TestFileSearch(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	os.WriteFile(filepath.Join(tempDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "other.log"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "subdir", "test3.txt"), []byte("test"), 0644)

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 搜索 .txt 文件（非递归）
	result, err := module.searchFiles(".", "*.txt", false)
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Search operation failed: %v", result["error"])
	}

	files := result["files"].([]map[string]any)
	if len(files) != 2 {
		t.Errorf("Expected 2 .txt files, got %d", len(files))
	}

	// 搜索 .txt 文件（递归）
	result, err = module.searchFiles(".", "*.txt", true)
	if err != nil {
		t.Fatalf("Failed to search files recursively: %v", err)
	}

	files = result["files"].([]map[string]any)
	if len(files) != 3 {
		t.Errorf("Expected 3 .txt files recursively, got %d", len(files))
	}
}

// TestFileInfo 测试获取文件信息
func TestFileInfo(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 获取文件信息
	result, err := module.getFileInfo("test.txt")
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Get info operation failed: %v", result["error"])
	}

	if result["name"].(string) != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", result["name"])
	}

	if result["size"].(int64) != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), result["size"])
	}

	if result["is_dir"].(bool) {
		t.Error("File should not be a directory")
	}
}

// TestFileSizeLimit 测试文件大小限制
func TestFileSizeLimit(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 1, // 1MB limit
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 尝试写入超过限制的文件
	largeContent := strings.Repeat("a", 2*1024*1024) // 2MB
	result, err := module.writeFile("large.txt", largeContent, false)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow writing file exceeding size limit")
	}

	if !strings.Contains(result["error"].(string), "exceeds limit") {
		t.Errorf("Expected size limit error, got: %v", result["error"])
	}
}

// TestWritePermission 测试写入权限
func TestWritePermission(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  false, // 禁用写入
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 尝试写入文件
	result, err := module.writeFile("test.txt", "content", false)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow writing when write permission is disabled")
	}

	// 尝试创建文件
	result, err = module.createFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow creating when write permission is disabled")
	}
}

// TestDeletePermission 测试删除权限
func TestDeletePermission(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: false, // 禁用删除
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 尝试删除文件
	result, err := module.deleteFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow deleting when delete permission is disabled")
	}
}

// TestFileExtensionWhitelist 测试文件扩展名白名单
func TestFileExtensionWhitelist(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:         tempDir,
		MaxFileSize:     10,
		AllowWrite:      true,
		AllowDelete:     true,
		AllowedFileExts: []string{".txt", ".log"},
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 允许的扩展名
	result, err := module.writeFile("test.txt", "content", false)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Should allow writing .txt file: %v", result["error"])
	}

	// 不允许的扩展名
	result, err = module.writeFile("test.exe", "content", false)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow writing .exe file")
	}
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 并发写入多个文件
	var wg sync.WaitGroup
	numFiles := 10

	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			filename := filepath.Join("file", string(rune('0'+index)), ".txt")
			content := "Content " + string(rune('0'+index))

			result, err := module.writeFile(filename, content, false)
			if err != nil {
				t.Errorf("Failed to write file %s: %v", filename, err)
				return
			}

			if !result["success"].(bool) {
				t.Errorf("Write operation failed for %s: %v", filename, result["error"])
			}
		}(i)
	}

	wg.Wait()

	// 验证统计信息
	stats := module.GetStats()
	if stats["total_operations"] != int64(numFiles) {
		t.Errorf("Expected %d operations, got %d", numFiles, stats["total_operations"])
	}
}

// TestMCPTools 测试 MCP 工具
func TestMCPTools(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	ctx := context.Background()

	// 获取工具列表
	tools, err := module.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	expectedTools := []string{
		"file_read",
		"file_write",
		"file_create",
		"file_delete",
		"file_list",
		"file_info",
		"file_search",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	// 验证每个工具都有正确的信息
	for i, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			t.Errorf("Failed to get info for tool %d: %v", i, err)
			continue
		}

		if info.Name != expectedTools[i] {
			t.Errorf("Expected tool name %s, got %s", expectedTools[i], info.Name)
		}

		if info.Desc == "" {
			t.Errorf("Tool %s has empty description", info.Name)
		}
	}
}

// TestPathTraversalSecurity 测试路径遍历安全性
func TestPathTraversalSecurity(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 尝试路径遍历攻击
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"test/../../etc/passwd",
		"./../../etc/passwd",
	}

	for _, path := range maliciousPaths {
		t.Run("Path: "+path, func(t *testing.T) {
			result, err := module.readFile(path)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if result["success"].(bool) {
				t.Errorf("Should not allow reading path: %s", path)
			}

			if !strings.Contains(result["error"].(string), "traversal") &&
				!strings.Contains(result["error"].(string), "outside root") {
				t.Errorf("Expected path traversal error, got: %v", result["error"])
			}
		})
	}
}

// TestStatistics 测试统计功能
func TestStatistics(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: false,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 执行一些操作
	module.writeFile("test1.txt", "content", false)
	module.writeFile("test2.txt", "content", false)
	module.readFile("test1.txt")
	module.deleteFile("test1.txt") // 应该被阻止

	// 检查统计信息
	stats := module.GetStats()

	if stats["total_operations"] != 4 {
		t.Errorf("Expected 4 total operations, got %d", stats["total_operations"])
	}

	if stats["success_count"] != 3 {
		t.Errorf("Expected 3 successful operations, got %d", stats["success_count"])
	}

	if stats["blocked_count"] != 1 {
		t.Errorf("Expected 1 blocked operation, got %d", stats["blocked_count"])
	}
}

// TestFileReaderChunked 测试分块读取
func TestFileReaderChunked(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 10, // 小块大小用于测试
	}

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test file with more than 10 bytes of content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 分块读取
	var chunks [][]byte
	err := reader.ReadChunked(testFile, func(chunk []byte) error {
		// 复制 chunk 因为 buffer 会被重用
		chunkCopy := make([]byte, len(chunk))
		copy(chunkCopy, chunk)
		chunks = append(chunks, chunkCopy)
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to read chunked: %v", err)
	}

	// 验证分块数量
	if len(chunks) < 2 {
		t.Errorf("Expected at least 2 chunks, got %d", len(chunks))
	}

	// 重组内容
	var reconstructed strings.Builder
	for _, chunk := range chunks {
		reconstructed.Write(chunk)
	}

	if reconstructed.String() != testContent {
		t.Errorf("Reconstructed content doesn't match original")
	}
}

// TestFileNavigatorGetInfo 测试获取文件信息
func TestFileNavigatorGetInfo(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 获取文件信息
	info, err := navigator.GetInfo(testFile)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Name != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", info.Name)
	}

	if info.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), info.Size)
	}

	if info.IsDir {
		t.Error("File should not be a directory")
	}

	// 测试目录
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dirInfo, err := navigator.GetInfo(testDir)
	if err != nil {
		t.Fatalf("Failed to get directory info: %v", err)
	}

	if !dirInfo.IsDir {
		t.Error("Directory should be marked as directory")
	}
}

// TestFileWriterCreate 测试创建空文件
func TestFileWriterCreate(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	writer := &FileWriter{
		validator:   validator,
		maxFileSize: 10 * 1024 * 1024,
	}

	// 创建空文件
	testFile := filepath.Join(tempDir, "empty.txt")
	err := writer.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// 验证文件存在且为空
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("File not created: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("Expected empty file, got size %d", info.Size())
	}
}

// TestPathValidatorGetSafePath 测试获取安全路径
func TestPathValidatorGetSafePath(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "Relative path",
			input:     "test.txt",
			shouldErr: false,
		},
		{
			name:      "Absolute path in root",
			input:     filepath.Join(tempDir, "test.txt"),
			shouldErr: false,
		},
		{
			name:      "Path with ..",
			input:     "../test.txt",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safePath, err := validator.GetSafePath(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for path %s, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for path %s: %v", tt.input, err)
			}
			if !tt.shouldErr && safePath == "" {
				t.Error("Expected non-empty safe path")
			}
		})
	}
}

// TestFileNavigatorListNonDirectory 测试列出非目录路径
func TestFileNavigatorListNonDirectory(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 尝试列出文件（应该失败）
	_, err := navigator.List(testFile, false)
	if err == nil {
		t.Error("Expected error when listing a file, got nil")
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Expected 'not a directory' error, got: %v", err)
	}
}

// TestFileNavigatorSearchInvalidPattern 测试无效搜索模式
func TestFileNavigatorSearchInvalidPattern(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 搜索时使用无效模式（应该被忽略）
	files, err := navigator.Search(tempDir, "[invalid", false)
	if err != nil {
		t.Fatalf("Search should not fail with invalid pattern: %v", err)
	}

	// 应该返回空列表
	if len(files) != 0 {
		t.Errorf("Expected 0 files with invalid pattern, got %d", len(files))
	}
}

// TestFileReaderReadDirectory 测试读取目录
func TestFileReaderReadDirectory(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 1024,
	}

	// 创建测试目录
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// 尝试读取目录（应该失败）
	_, err := reader.Read(testDir)
	if err == nil {
		t.Error("Expected error when reading a directory, got nil")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Expected 'directory' error, got: %v", err)
	}
}

// TestFileReaderReadNonExistent 测试读取不存在的文件
func TestFileReaderReadNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 1024,
	}

	// 尝试读取不存在的文件
	_, err := reader.Read(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error when reading non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestFileWriterAppendExceedingLimit 测试追加超过限制
func TestFileWriterAppendExceedingLimit(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	writer := &FileWriter{
		validator:   validator,
		maxFileSize: 100, // 100 bytes limit
	}

	// 创建初始文件
	testFile := filepath.Join(tempDir, "test.txt")
	initialContent := strings.Repeat("a", 60)
	if err := writer.Write(testFile, []byte(initialContent), false); err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	// 尝试追加超过限制的内容
	appendContent := strings.Repeat("b", 50)
	err := writer.Write(testFile, []byte(appendContent), true)
	if err == nil {
		t.Error("Expected error when appending exceeds limit, got nil")
	}

	if !strings.Contains(err.Error(), "exceed limit") {
		t.Errorf("Expected 'exceed limit' error, got: %v", err)
	}
}

// TestFileNavigatorDeleteNonExistent 测试删除不存在的文件
func TestFileNavigatorDeleteNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 删除不存在的文件（应该成功，因为 RemoveAll 不会报错）
	err := navigator.Delete(filepath.Join(tempDir, "nonexistent.txt"))
	if err != nil {
		t.Errorf("Delete non-existent file should not error: %v", err)
	}
}

// TestModuleWithInvalidRootDir 测试无效根目录
func TestModuleWithInvalidRootDir(t *testing.T) {
	config := FileConfig{
		RootDir:     "\x00invalid", // 无效路径
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	_, err := NewFileModule(config)
	if err == nil {
		t.Error("Expected error with invalid root directory, got nil")
	}
}

// TestFileNavigatorListInvalidPath 测试列出无效路径
func TestFileNavigatorListInvalidPath(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 尝试列出不存在的路径
	_, err := navigator.List(filepath.Join(tempDir, "nonexistent"), false)
	if err == nil {
		t.Error("Expected error when listing non-existent path, got nil")
	}
}

// TestFileNavigatorSearchInvalidPath 测试搜索无效路径
func TestFileNavigatorSearchInvalidPath(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 尝试在不存在的路径中搜索（应该返回错误或空结果）
	files, err := navigator.Search(filepath.Join(tempDir, "nonexistent"), "*.txt", false)

	// filepath.Walk 会在路径不存在时返回错误
	if err != nil {
		// 这是预期的行为
		return
	}

	// 如果没有错误，应该返回空列表
	if len(files) != 0 {
		t.Errorf("Expected 0 files for non-existent path, got %d", len(files))
	}
}

// TestFileReaderChunkedInvalidPath 测试分块读取无效路径
func TestFileReaderChunkedInvalidPath(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 1024,
	}

	// 尝试分块读取不存在的文件
	err := reader.ReadChunked(filepath.Join(tempDir, "nonexistent.txt"), func(chunk []byte) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error when reading non-existent file, got nil")
	}
}

// TestFileWriterCreateNestedDirectory 测试创建嵌套目录中的文件
func TestFileWriterCreateNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	writer := &FileWriter{
		validator:   validator,
		maxFileSize: 10 * 1024 * 1024,
	}

	// 创建嵌套目录中的文件
	nestedFile := filepath.Join(tempDir, "a", "b", "c", "test.txt")
	err := writer.Write(nestedFile, []byte("test content"), false)
	if err != nil {
		t.Fatalf("Failed to create file in nested directory: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(nestedFile); err != nil {
		t.Errorf("File not created in nested directory: %v", err)
	}
}

// TestConcurrentReadWrite 测试并发读写
func TestConcurrentReadWrite(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 创建初始文件
	testFile := "test.txt"
	module.writeFile(testFile, "initial content", false)

	var wg sync.WaitGroup
	numOps := 20

	// 并发读写
	for i := 0; i < numOps; i++ {
		wg.Add(2)

		// 读操作
		go func() {
			defer wg.Done()
			module.readFile(testFile)
		}()

		// 写操作
		go func(index int) {
			defer wg.Done()
			content := "content " + string(rune('0'+index%10))
			module.writeFile(testFile, content, false)
		}(i)
	}

	wg.Wait()

	// 验证统计信息
	stats := module.GetStats()
	if stats["total_operations"] != int64(numOps*2+1) { // +1 for initial write
		t.Errorf("Expected %d operations, got %d", numOps*2+1, stats["total_operations"])
	}
}

// TestFileModuleIntegration 测试完整的文件模块集成
func TestFileModuleIntegration(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 1. 创建文件
	result, err := module.createFile("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 2. 写入内容
	result, err = module.writeFile("test.txt", "Hello, World!", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to write file: %v", err)
	}

	// 3. 读取内容
	result, err = module.readFile("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to read file: %v", err)
	}

	if result["content"].(string) != "Hello, World!" {
		t.Errorf("Content mismatch: got %s", result["content"])
	}

	// 4. 追加内容
	result, err = module.writeFile("test.txt", " Appended", true)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to append file: %v", err)
	}

	// 5. 再次读取
	result, err = module.readFile("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to read file after append: %v", err)
	}

	if result["content"].(string) != "Hello, World! Appended" {
		t.Errorf("Content mismatch after append: got %s", result["content"])
	}

	// 6. 获取文件信息
	result, err = module.getFileInfo("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to get file info: %v", err)
	}

	// 7. 列出文件
	result, err = module.listFiles(".", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to list files: %v", err)
	}

	files := result["files"].([]map[string]any)
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// 8. 搜索文件
	result, err = module.searchFiles(".", "*.txt", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to search files: %v", err)
	}

	searchFiles := result["files"].([]map[string]any)
	if len(searchFiles) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searchFiles))
	}

	// 9. 删除文件
	result, err = module.deleteFile("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// 10. 验证文件已删除
	result, err = module.readFile("test.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["success"].(bool) {
		t.Error("File should not exist after deletion")
	}
}

// TestFileModuleDefaultConfig 测试默认配置
func TestFileModuleDefaultConfig(t *testing.T) {
	config := FileConfig{
		// 使用默认值
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module with default config: %v", err)
	}
	defer module.Close()

	if module.config.RootDir == "" {
		t.Error("Root directory should have default value")
	}

	if module.config.MaxFileSize != 100 {
		t.Errorf("Expected default max file size 100, got %d", module.config.MaxFileSize)
	}
}

// TestFileNavigatorListRecursiveWithSubdirs 测试递归列出多层目录
func TestFileNavigatorListRecursiveWithSubdirs(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多层目录结构
	os.MkdirAll(filepath.Join(tempDir, "a", "b", "c"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "a", "file2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "a", "b", "file3.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "a", "b", "c", "file4.txt"), []byte("test"), 0644)

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 递归列出所有文件
	result, err := module.listFiles(".", true)
	if err != nil {
		t.Fatalf("Failed to list files recursively: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("List operation failed: %v", result["error"])
	}

	totalCount := result["total_count"].(int)
	// 应该有 4 个文件 + 3 个目录 = 7 个项目
	if totalCount != 7 {
		t.Errorf("Expected 7 total items, got %d", totalCount)
	}
}

// TestFileSearchRecursiveMultipleMatches 测试递归搜索多个匹配
func TestFileSearchRecursiveMultipleMatches(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个匹配的文件
	os.MkdirAll(filepath.Join(tempDir, "subdir1"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "subdir2"), 0755)
	os.WriteFile(filepath.Join(tempDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test2.log"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir1", "test3.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir2", "test4.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir2", "other.log"), []byte("test"), 0644)

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 递归搜索所有 .txt 文件
	result, err := module.searchFiles(".", "*.txt", true)
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Search operation failed: %v", result["error"])
	}

	files := result["files"].([]map[string]any)
	// 应该找到 3 个 .txt 文件（test1.txt, test3.txt, test4.txt）
	// 注意：test2.log 不匹配
	if len(files) != 3 {
		t.Errorf("Expected 3 .txt files, got %d", len(files))
	}
}

// TestFileModuleBlockedPaths 测试阻止路径
func TestFileModuleBlockedPaths(t *testing.T) {
	tempDir := t.TempDir()

	// 创建阻止的子目录
	blockedDir := filepath.Join(tempDir, "blocked")
	os.Mkdir(blockedDir, 0755)

	config := FileConfig{
		RootDir:      tempDir,
		MaxFileSize:  10,
		AllowWrite:   true,
		AllowDelete:  true,
		BlockedPaths: []string{blockedDir},
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 尝试在阻止的路径中创建文件
	result, err := module.writeFile(filepath.Join("blocked", "test.txt"), "content", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Should not allow writing to blocked path")
	}

	if !strings.Contains(result["error"].(string), "blocked") {
		t.Errorf("Expected 'blocked' error, got: %v", result["error"])
	}
}

// TestFileReaderChunkedCallbackError 测试分块读取回调错误
func TestFileReaderChunkedCallbackError(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 10,
	}

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test file with more than 10 bytes"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 分块读取，回调返回错误
	callbackErr := fmt.Errorf("callback error")
	err := reader.ReadChunked(testFile, func(chunk []byte) error {
		return callbackErr
	})

	if err == nil {
		t.Error("Expected error from callback, got nil")
	}

	if err != callbackErr {
		t.Errorf("Expected callback error, got: %v", err)
	}
}

// TestFileWriterInvalidPath 测试写入无效路径
func TestFileWriterInvalidPath(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	writer := &FileWriter{
		validator:   validator,
		maxFileSize: 10 * 1024 * 1024,
	}

	// 尝试写入根目录外的路径
	err := writer.Write("../outside.txt", []byte("content"), false)
	if err == nil {
		t.Error("Expected error when writing outside root, got nil")
	}
}

// TestPathValidatorConcurrent 测试路径验证器并发安全
func TestPathValidatorConcurrent(t *testing.T) {
	tempDir := t.TempDir()

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	var wg sync.WaitGroup
	numOps := 100

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			path := filepath.Join(tempDir, "file"+string(rune('0'+index%10))+".txt")
			validator.Validate(path)
			validator.GetSafePath(path)
		}(i)
	}

	wg.Wait()
}

// TestFileNavigatorSearchNonRecursive 测试非递归搜索
func TestFileNavigatorSearchNonRecursive(t *testing.T) {
	tempDir := t.TempDir()

	// 创建文件结构
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir", "test3.txt"), []byte("test"), 0644)

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 非递归搜索
	files, err := navigator.Search(tempDir, "*.txt", false)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// 应该只找到根目录的 2 个文件
	if len(files) != 2 {
		t.Errorf("Expected 2 files in non-recursive search, got %d", len(files))
	}
}

// TestFileModuleStatsAfterClose 测试关闭后的统计
func TestFileModuleStatsAfterClose(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}

	// 执行一些操作
	module.writeFile("test.txt", "content", false)
	module.readFile("test.txt")

	// 关闭模块
	if err := module.Close(); err != nil {
		t.Errorf("Failed to close module: %v", err)
	}

	// 获取统计信息（应该仍然可用）
	stats := module.GetStats()
	if stats["total_operations"] != 2 {
		t.Errorf("Expected 2 operations, got %d", stats["total_operations"])
	}
}

// TestFileNavigatorListEmptyDirectory 测试列出空目录
func TestFileNavigatorListEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// 创建空目录
	emptyDir := filepath.Join(tempDir, "empty")
	os.Mkdir(emptyDir, 0755)

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 列出空目录
	files, err := navigator.List(emptyDir, false)
	if err != nil {
		t.Fatalf("Failed to list empty directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

// TestFileNavigatorSearchEmptyDirectory 测试在空目录中搜索
func TestFileNavigatorSearchEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// 创建空目录
	emptyDir := filepath.Join(tempDir, "empty")
	os.Mkdir(emptyDir, 0755)

	validator := &PathValidator{
		rootDir:      tempDir,
		blockedPaths: []string{},
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	// 在空目录中搜索
	files, err := navigator.Search(emptyDir, "*.txt", false)
	if err != nil {
		t.Fatalf("Failed to search empty directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory search, got %d", len(files))
	}
}

// TestFileModuleReadAfterWrite 测试写入后立即读取
func TestFileModuleReadAfterWrite(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	testContent := "Test content for immediate read"

	// 写入文件
	writeResult, err := module.writeFile("test.txt", testContent, false)
	if err != nil || !writeResult["success"].(bool) {
		t.Fatalf("Failed to write file: %v", err)
	}

	// 立即读取
	readResult, err := module.readFile("test.txt")
	if err != nil || !readResult["success"].(bool) {
		t.Fatalf("Failed to read file: %v", err)
	}

	if readResult["content"].(string) != testContent {
		t.Errorf("Content mismatch: expected %s, got %s", testContent, readResult["content"])
	}
}

// TestFileModuleMultipleAppends 测试多次追加
func TestFileModuleMultipleAppends(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 初始写入
	module.writeFile("test.txt", "Line1\n", false)

	// 多次追加
	for i := 2; i <= 5; i++ {
		line := fmt.Sprintf("Line%d\n", i)
		result, err := module.writeFile("test.txt", line, true)
		if err != nil || !result["success"].(bool) {
			t.Fatalf("Failed to append line %d: %v", i, err)
		}
	}

	// 读取并验证
	result, err := module.readFile("test.txt")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "Line1\nLine2\nLine3\nLine4\nLine5\n"
	if result["content"].(string) != expected {
		t.Errorf("Content mismatch after multiple appends")
	}
}

// TestFileModuleListAfterOperations 测试操作后列出文件
func TestFileModuleListAfterOperations(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 创建多个文件
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		module.writeFile(filename, "content", false)
	}

	// 列出文件
	result, err := module.listFiles(".", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to list files: %v", err)
	}

	files := result["files"].([]map[string]any)
	if len(files) != 5 {
		t.Errorf("Expected 5 files, got %d", len(files))
	}

	// 删除一些文件
	module.deleteFile("file1.txt")
	module.deleteFile("file3.txt")

	// 再次列出
	result, err = module.listFiles(".", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to list files after deletion: %v", err)
	}

	files = result["files"].([]map[string]any)
	if len(files) != 3 {
		t.Errorf("Expected 3 files after deletion, got %d", len(files))
	}
}

// TestFileModuleSearchWithWildcards 测试通配符搜索
func TestFileModuleSearchWithWildcards(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 创建不同类型的文件
	module.writeFile("test1.txt", "content", false)
	module.writeFile("test2.txt", "content", false)
	module.writeFile("data1.log", "content", false)
	module.writeFile("data2.log", "content", false)
	module.writeFile("readme.md", "content", false)

	// 搜索 test*.txt
	result, err := module.searchFiles(".", "test*.txt", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to search: %v", err)
	}

	files := result["files"].([]map[string]any)
	if len(files) != 2 {
		t.Errorf("Expected 2 test*.txt files, got %d", len(files))
	}

	// 搜索 *.log
	result, err = module.searchFiles(".", "*.log", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to search: %v", err)
	}

	files = result["files"].([]map[string]any)
	if len(files) != 2 {
		t.Errorf("Expected 2 *.log files, got %d", len(files))
	}
}

// TestFileModuleGetInfoAfterModification 测试修改后获取信息
func TestFileModuleGetInfoAfterModification(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 创建文件
	module.writeFile("test.txt", "initial", false)

	// 获取初始信息
	result1, err := module.getFileInfo("test.txt")
	if err != nil || !result1["success"].(bool) {
		t.Fatalf("Failed to get initial file info: %v", err)
	}

	initialSize := result1["size"].(int64)

	// 修改文件
	module.writeFile("test.txt", "modified content", false)

	// 再次获取信息
	result2, err := module.getFileInfo("test.txt")
	if err != nil || !result2["success"].(bool) {
		t.Fatalf("Failed to get modified file info: %v", err)
	}

	modifiedSize := result2["size"].(int64)

	if modifiedSize == initialSize {
		t.Error("File size should have changed after modification")
	}
}

// TestFileModuleCreateInSubdirectory 测试在子目录中创建文件
func TestFileModuleCreateInSubdirectory(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 在子目录中创建文件（目录会自动创建）
	result, err := module.writeFile("subdir/test.txt", "content", false)
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to create file in subdirectory: %v", err)
	}

	// 验证文件存在
	readResult, err := module.readFile("subdir/test.txt")
	if err != nil || !readResult["success"].(bool) {
		t.Errorf("File not created in subdirectory: %v", err)
	}
}

// TestFileModuleDeleteDirectory 测试删除目录
func TestFileModuleDeleteDirectory(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: true,
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 创建目录和文件
	os.Mkdir(filepath.Join(tempDir, "testdir"), 0755)
	module.writeFile("testdir/file1.txt", "content", false)
	module.writeFile("testdir/file2.txt", "content", false)

	// 删除整个目录
	result, err := module.deleteFile("testdir")
	if err != nil || !result["success"].(bool) {
		t.Fatalf("Failed to delete directory: %v", err)
	}

	// 验证目录已删除
	listResult, err := module.listFiles(".", false)
	if err != nil || !listResult["success"].(bool) {
		t.Fatalf("Failed to list files: %v", err)
	}

	dirs := listResult["directories"].([]map[string]any)
	if len(dirs) != 0 {
		t.Error("Directory should have been deleted")
	}
}

// TestFileModuleStatsAccuracy 测试统计准确性
func TestFileModuleStatsAccuracy(t *testing.T) {
	tempDir := t.TempDir()

	config := FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 10,
		AllowWrite:  true,
		AllowDelete: false, // 禁用删除以测试阻止计数
	}

	module, err := NewFileModule(config)
	if err != nil {
		t.Fatalf("Failed to create file module: %v", err)
	}
	defer module.Close()

	// 执行各种操作
	module.writeFile("test1.txt", "content", false) // 成功
	module.writeFile("test2.txt", "content", false) // 成功
	module.readFile("test1.txt")                    // 成功
	module.readFile("nonexistent.txt")              // 失败
	module.deleteFile("test1.txt")                  // 被阻止
	module.deleteFile("test2.txt")                  // 被阻止

	stats := module.GetStats()

	if stats["total_operations"] != 6 {
		t.Errorf("Expected 6 total operations, got %d", stats["total_operations"])
	}

	if stats["success_count"] != 3 {
		t.Errorf("Expected 3 successful operations, got %d", stats["success_count"])
	}

	if stats["failure_count"] != 1 {
		t.Errorf("Expected 1 failed operation, got %d", stats["failure_count"])
	}

	if stats["blocked_count"] != 2 {
		t.Errorf("Expected 2 blocked operations, got %d", stats["blocked_count"])
	}
}
