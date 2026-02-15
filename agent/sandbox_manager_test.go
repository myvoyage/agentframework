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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"AgentFramework/pkg/tools/sandbox/file"

	"github.com/stretchr/testify/assert"
)

// TestNewSandboxManager 测试创建沙箱管理器
func TestNewSandboxManager(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()

	// 创建沙箱管理器
	sandboxMgr := NewSandboxManager(tempDir)

	// 验证沙箱管理器不为空
	assert.NotNil(t, sandboxMgr)

	// 验证沙箱目录设置正确
	assert.Equal(t, tempDir, sandboxMgr.GetSandboxDir(context.Background()))
}

// TestSandboxManager_ValidatePath 测试路径验证功能
func TestSandboxManager_ValidatePath(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 测试用例
	testCases := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "空路径",
			path:     "",
			expected: tempDir,
			wantErr:  false,
		},
		{
			name:     "相对路径",
			path:     "test.txt",
			expected: filepath.Join(tempDir, "test.txt"),
			wantErr:  false,
		},
		{
			name:     "绝对路径在沙箱内",
			path:     filepath.Join(tempDir, "test.txt"),
			expected: filepath.Join(tempDir, "test.txt"),
			wantErr:  false,
		},
		{
			name:    "路径遍历攻击",
			path:    "../outside.txt",
			wantErr: true,
		},
		{
			name:    "绝对路径在沙箱外",
			path:    filepath.Join(tempDir, "..", "outside.txt"),
			wantErr: true,
		},
		{
			name:     "多级相对路径",
			path:     "subdir1/subdir2/test.txt",
			expected: filepath.Join(tempDir, "subdir1/subdir2/test.txt"),
			wantErr:  false,
		},
		{
			name:     "当前目录",
			path:     ".",
			expected: tempDir,
			wantErr:  false,
		},
		{
			name:     "父目录符号在沙箱内",
			path:     "subdir/../test.txt",
			expected: filepath.Join(tempDir, "test.txt"),
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := sandboxMgr.ValidatePath(context.Background(), tc.path)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestSandboxManager_WithFileModule 测试设置文件模块
func TestSandboxManager_WithFileModule(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 创建文件模块
	fileConfig := file.FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 1024 * 1024, // 1MB
		AllowWrite:  true,
		AllowDelete: true,
	}

	fileModule, err := file.NewFileModule(fileConfig)
	assert.NoError(t, err)

	// 设置文件模块
	updatedMgr := sandboxMgr.WithFileModule(fileModule)

	// 验证更新后的沙箱管理器不为空
	assert.NotNil(t, updatedMgr)
}

// TestSandboxManager_CreateFileTools 测试创建文件工具
func TestSandboxManager_CreateFileTools(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 创建文件模块
	fileConfig := file.FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 1024 * 1024, // 1MB
		AllowWrite:  true,
		AllowDelete: true,
	}

	fileModule, err := file.NewFileModule(fileConfig)
	assert.NoError(t, err)

	// 设置文件模块
	sandboxMgr = sandboxMgr.WithFileModule(fileModule)

	// 创建文件工具
	fileTools, err := sandboxMgr.CreateFileTools(context.Background())

	// 验证创建文件工具成功
	assert.NoError(t, err)
	assert.NotEmpty(t, fileTools)

	// 验证包含预期的文件工具
	expectedTools := []string{
		"file_read",
		"file_write",
		"file_create",
		"file_delete",
		"file_list",
		"file_info",
		"file_search",
	}

	for _, toolName := range expectedTools {
		assert.Contains(t, fileTools, toolName, "工具 %s 不存在", toolName)
	}
}

// TestSandboxManager_CreateFileTools_WithoutFileModule 测试未设置文件模块时创建文件工具
func TestSandboxManager_CreateFileTools_WithoutFileModule(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 未设置文件模块，尝试创建文件工具应该失败
	fileTools, err := sandboxMgr.CreateFileTools(context.Background())

	// 验证创建文件工具失败
	assert.Error(t, err)
	assert.Nil(t, fileTools)
	assert.Equal(t, "file module not set", err.Error())
}

// TestSandboxManager_Close 测试关闭沙箱管理器
func TestSandboxManager_Close(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 创建文件模块
	fileConfig := file.FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 1024 * 1024, // 1MB
		AllowWrite:  true,
		AllowDelete: true,
	}

	fileModule, err := file.NewFileModule(fileConfig)
	assert.NoError(t, err)

	// 设置文件模块
	sandboxMgr = sandboxMgr.WithFileModule(fileModule)

	// 关闭沙箱管理器
	err = sandboxMgr.Close()

	// 验证关闭成功
	assert.NoError(t, err)
}

// TestSandboxManager_Close_WithoutFileModule 测试关闭未设置文件模块的沙箱管理器
func TestSandboxManager_Close_WithoutFileModule(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 未设置文件模块，关闭沙箱管理器应该成功
	err := sandboxMgr.Close()

	// 验证关闭成功
	assert.NoError(t, err)
}

// TestSandboxManager_Integration 测试沙箱管理器的集成功能
func TestSandboxManager_Integration(t *testing.T) {
	t.Parallel()

	// 创建临时目录作为沙箱根目录
	tempDir := t.TempDir()
	sandboxMgr := NewSandboxManager(tempDir)

	// 创建文件模块
	fileConfig := file.FileConfig{
		RootDir:     tempDir,
		MaxFileSize: 1024 * 1024, // 1MB
		AllowWrite:  true,
		AllowDelete: true,
	}

	fileModule, err := file.NewFileModule(fileConfig)
	assert.NoError(t, err)

	// 设置文件模块
	sandboxMgr = sandboxMgr.WithFileModule(fileModule)

	// 创建文件工具
	fileTools, err := sandboxMgr.CreateFileTools(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, fileTools)

	// 验证文件工具数量
	assert.Len(t, fileTools, 7)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test_integration.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)
	defer os.Remove(testFile)

	// 验证路径验证功能
	validPath, err := sandboxMgr.ValidatePath(context.Background(), "test_integration.txt")
	assert.NoError(t, err)
	assert.Equal(t, testFile, validPath)

	// 关闭沙箱管理器
	err = sandboxMgr.Close()
	assert.NoError(t, err)
}
