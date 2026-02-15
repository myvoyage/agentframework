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

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// FileModule 文件操作模块
type FileModule struct {
	config    FileConfig
	manager   *FileManager
	validator *PathValidator
	stats     *OperationStats
	mu        sync.RWMutex
}

// FileConfig 文件操作配置
type FileConfig struct {
	RootDir         string   `json:"root_dir"`
	MaxFileSize     int64    `json:"max_file_size"` // 单位：MB
	AllowWrite      bool     `json:"allow_write"`
	AllowDelete     bool     `json:"allow_delete"`
	BlockedPaths    []string `json:"blocked_paths"`
	AllowedFileExts []string `json:"allowed_file_exts"` // 文件扩展名白名单
}

// FileManager 文件管理器
type FileManager struct {
	config    FileConfig
	validator *PathValidator
	reader    *FileReader
	writer    *FileWriter
	navigator *FileNavigator
}

// PathValidator 路径验证器
type PathValidator struct {
	rootDir      string
	blockedPaths []string
	mu           sync.RWMutex
}

// FileReader 文件读取器
type FileReader struct {
	validator    *PathValidator
	maxChunkSize int64
}

// FileWriter 文件写入器
type FileWriter struct {
	validator   *PathValidator
	maxFileSize int64
}

// FileNavigator 目录导航器
type FileNavigator struct {
	validator *PathValidator
}

// OperationStats 操作统计
type OperationStats struct {
	TotalOperations int64
	SuccessCount    int64
	FailureCount    int64
	BlockedCount    int64
	mu              sync.RWMutex
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
	Mode    string    `json:"mode"`
}

// NewFileModule 创建文件操作模块实例
func NewFileModule(config FileConfig) (*FileModule, error) {
	// 验证配置
	if config.RootDir == "" {
		config.RootDir = "."
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 100 // 默认100MB
	}

	// 规范化根目录路径
	absRootDir, err := filepath.Abs(config.RootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}
	config.RootDir = absRootDir

	// 创建根目录（如果不存在）
	if err := os.MkdirAll(config.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	// 规范化阻止路径列表
	blockedPaths := make([]string, 0, len(config.BlockedPaths))
	for _, path := range config.BlockedPaths {
		absPath, err := filepath.Abs(path)
		if err == nil {
			blockedPaths = append(blockedPaths, absPath)
		}
	}
	config.BlockedPaths = blockedPaths

	// 创建路径验证器
	validator := &PathValidator{
		rootDir:      config.RootDir,
		blockedPaths: config.BlockedPaths,
	}

	// 创建文件管理器组件
	reader := &FileReader{
		validator:    validator,
		maxChunkSize: 1024 * 1024, // 1MB chunks
	}

	writer := &FileWriter{
		validator:   validator,
		maxFileSize: config.MaxFileSize * 1024 * 1024, // 转换为字节
	}

	navigator := &FileNavigator{
		validator: validator,
	}

	manager := &FileManager{
		config:    config,
		validator: validator,
		reader:    reader,
		writer:    writer,
		navigator: navigator,
	}

	stats := &OperationStats{}

	return &FileModule{
		config:    config,
		manager:   manager,
		validator: validator,
		stats:     stats,
	}, nil
}

// GetTools 返回文件操作模块的 MCP 工具列表
func (m *FileModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 文件读取工具
		&fileReadTool{module: m},
		// 文件写入工具
		&fileWriteTool{module: m},
		// 文件创建工具
		&fileCreateTool{module: m},
		// 文件删除工具
		&fileDeleteTool{module: m},
		// 文件列表工具
		&fileListTool{module: m},
		// 文件信息工具
		&fileInfoTool{module: m},
		// 文件搜索工具
		&fileSearchTool{module: m},
	}

	return tools, nil
}

// 文件读取工具
type fileReadTool struct {
	module *FileModule
}

func (t *fileReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_read",
		Desc: "Read the content of a file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "File path",
			},
		}),
	}, nil
}

func (t *fileReadTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.readFile(args.Path)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件写入工具
type fileWriteTool struct {
	module *FileModule
}

func (t *fileWriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_write",
		Desc: "Write content to a file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "File path",
			},
			"content": {
				Type: "string",
				Desc: "File content",
			},
			"append": {
				Type: "boolean",
				Desc: "Append to file",
			},
		}),
	}, nil
}

func (t *fileWriteTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Append  bool   `json:"append"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.writeFile(args.Path, args.Content, args.Append)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件创建工具
type fileCreateTool struct {
	module *FileModule
}

func (t *fileCreateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_create",
		Desc: "Create a new file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "File path",
			},
		}),
	}, nil
}

func (t *fileCreateTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.createFile(args.Path)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件删除工具
type fileDeleteTool struct {
	module *FileModule
}

func (t *fileDeleteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_delete",
		Desc: "Delete a file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "File path",
			},
		}),
	}, nil
}

func (t *fileDeleteTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.deleteFile(args.Path)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件列表工具
type fileListTool struct {
	module *FileModule
}

func (t *fileListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_list",
		Desc: "List files in a directory",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "Directory path",
			},
			"recursive": {
				Type: "boolean",
				Desc: "Recursively list subdirectories",
			},
		}),
	}, nil
}

func (t *fileListTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.listFiles(args.Path, args.Recursive)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件信息工具
type fileInfoTool struct {
	module *FileModule
}

func (t *fileInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_info",
		Desc: "Get detailed information about a file or directory",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "File or directory path",
			},
		}),
	}, nil
}

func (t *fileInfoTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.getFileInfo(args.Path)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文件搜索工具
type fileSearchTool struct {
	module *FileModule
}

func (t *fileSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_search",
		Desc: "Search for files by name pattern",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "Directory path to search in",
			},
			"pattern": {
				Type: "string",
				Desc: "File name pattern (supports wildcards)",
			},
			"recursive": {
				Type: "boolean",
				Desc: "Search recursively in subdirectories",
			},
		}),
	}, nil
}

func (t *fileSearchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Path      string `json:"path"`
		Pattern   string `json:"pattern"`
		Recursive bool   `json:"recursive"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.searchFiles(args.Path, args.Pattern, args.Recursive)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭文件操作模块，释放资源
func (m *FileModule) Close() error {
	// 记录最终统计信息
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	// 可以在这里添加日志记录
	return nil
}

// GetStats 获取操作统计信息
func (m *FileModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_operations": m.stats.TotalOperations,
		"success_count":    m.stats.SuccessCount,
		"failure_count":    m.stats.FailureCount,
		"blocked_count":    m.stats.BlockedCount,
	}
}

// ============================================================================
// PathValidator 实现
// ============================================================================

// Validate 验证路径是否安全
func (v *PathValidator) Validate(path string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 1. 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// 2. 清理路径（移除 .. 等）
	cleanPath := filepath.Clean(absPath)

	// 3. 检查路径遍历 - 确保清理后的路径在根目录内
	if !strings.HasPrefix(cleanPath, v.rootDir) {
		return fmt.Errorf("path traversal detected: path outside root directory")
	}

	// 4. 检查黑名单
	for _, blocked := range v.blockedPaths {
		if strings.HasPrefix(cleanPath, blocked) {
			return fmt.Errorf("access denied: path is blocked")
		}
	}

	return nil
}

// GetSafePath 获取安全的绝对路径
func (v *PathValidator) GetSafePath(path string) (string, error) {
	// 如果是相对路径，相对于根目录
	if !filepath.IsAbs(path) {
		path = filepath.Join(v.rootDir, path)
	}

	// 验证路径
	if err := v.Validate(path); err != nil {
		return "", err
	}

	// 返回清理后的绝对路径
	absPath, _ := filepath.Abs(path)
	return filepath.Clean(absPath), nil
}

// ============================================================================
// FileReader 实现
// ============================================================================

// Read 读取文件内容
func (r *FileReader) Read(path string) ([]byte, error) {
	// 验证路径
	safePath, err := r.validator.GetSafePath(path)
	if err != nil {
		return nil, err
	}

	// 检查文件是否存在
	info, err := os.Stat(safePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 检查是否为目录
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// 读取文件
	content, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

// ReadChunked 分块读取大文件
func (r *FileReader) ReadChunked(path string, callback func([]byte) error) error {
	// 验证路径
	safePath, err := r.validator.GetSafePath(path)
	if err != nil {
		return err
	}

	// 打开文件
	file, err := os.Open(safePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 分块读取
	buffer := make([]byte, r.maxChunkSize)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		if n > 0 {
			if err := callback(buffer[:n]); err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// ============================================================================
// FileWriter 实现
// ============================================================================

// Write 写入文件内容
func (w *FileWriter) Write(path string, content []byte, appendMode bool) error {
	// 验证路径
	safePath, err := w.validator.GetSafePath(path)
	if err != nil {
		return err
	}

	// 检查文件大小限制
	if int64(len(content)) > w.maxFileSize {
		return fmt.Errorf("file size exceeds limit: %d > %d bytes", len(content), w.maxFileSize)
	}

	// 如果是追加模式，检查现有文件大小
	if appendMode {
		if info, err := os.Stat(safePath); err == nil {
			totalSize := info.Size() + int64(len(content))
			if totalSize > w.maxFileSize {
				return fmt.Errorf("total file size would exceed limit: %d > %d bytes", totalSize, w.maxFileSize)
			}
		}
	}

	// 确保父目录存在
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	var flags int
	if appendMode {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}

	file, err := os.OpenFile(safePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Create 创建空文件
func (w *FileWriter) Create(path string) error {
	return w.Write(path, []byte{}, false)
}

// ============================================================================
// FileNavigator 实现
// ============================================================================

// List 列出目录内容
func (n *FileNavigator) List(path string, recursive bool) ([]FileInfo, error) {
	// 验证路径
	safePath, err := n.validator.GetSafePath(path)
	if err != nil {
		return nil, err
	}

	// 检查是否为目录
	info, err := os.Stat(safePath)
	if err != nil {
		return nil, fmt.Errorf("path not found: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	var files []FileInfo

	if recursive {
		// 递归遍历
		err = filepath.Walk(safePath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过错误的文件
			}

			// 跳过根目录本身
			if p == safePath {
				return nil
			}

			files = append(files, FileInfo{
				Name:    info.Name(),
				Path:    p,
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime(),
				Mode:    info.Mode().String(),
			})

			return nil
		})
	} else {
		// 只列出当前目录
		entries, err := os.ReadDir(safePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue // 跳过错误的文件
			}

			fullPath := filepath.Join(safePath, entry.Name())
			files = append(files, FileInfo{
				Name:    entry.Name(),
				Path:    fullPath,
				Size:    info.Size(),
				IsDir:   entry.IsDir(),
				ModTime: info.ModTime(),
				Mode:    info.Mode().String(),
			})
		}
	}

	return files, err
}

// GetInfo 获取文件信息
func (n *FileNavigator) GetInfo(path string) (*FileInfo, error) {
	// 验证路径
	safePath, err := n.validator.GetSafePath(path)
	if err != nil {
		return nil, err
	}

	// 获取文件信息
	info, err := os.Stat(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:    info.Name(),
		Path:    safePath,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode().String(),
	}, nil
}

// Search 搜索文件
func (n *FileNavigator) Search(path string, pattern string, recursive bool) ([]FileInfo, error) {
	// 验证路径
	safePath, err := n.validator.GetSafePath(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过错误的文件
		}

		// 跳过根目录本身
		if p == safePath {
			return nil
		}

		// 如果不是递归模式，跳过子目录
		if !recursive && filepath.Dir(p) != safePath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 匹配文件名
		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return nil // 跳过无效的模式
		}

		if matched {
			files = append(files, FileInfo{
				Name:    info.Name(),
				Path:    p,
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime(),
				Mode:    info.Mode().String(),
			})
		}

		return nil
	}

	err = filepath.Walk(safePath, walkFunc)
	return files, err
}

// Delete 删除文件或目录
func (n *FileNavigator) Delete(path string) error {
	// 验证路径
	safePath, err := n.validator.GetSafePath(path)
	if err != nil {
		return err
	}

	// 删除文件或目录
	if err := os.RemoveAll(safePath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// ============================================================================
// 文件操作模块核心功能实现
// ============================================================================

// readFile 读取文件
func (m *FileModule) readFile(path string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 读取文件
	content, err := m.manager.reader.Read(path)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	// 获取文件信息
	info, _ := m.manager.navigator.GetInfo(path)

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	result := map[string]any{
		"success": true,
		"path":    path,
		"content": string(content),
		"size":    len(content),
	}

	if info != nil {
		result["mod_time"] = info.ModTime
	}

	return result, nil
}

// writeFile 写入文件
func (m *FileModule) writeFile(path, content string, appendMode bool) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 检查是否允许写入
	if !m.config.AllowWrite {
		m.stats.mu.Lock()
		m.stats.BlockedCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   "Write operation not allowed",
			"path":    path,
		}, nil
	}

	// 检查文件扩展名白名单
	if len(m.config.AllowedFileExts) > 0 {
		ext := strings.ToLower(filepath.Ext(path))
		allowed := false
		for _, allowedExt := range m.config.AllowedFileExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			m.stats.mu.Lock()
			m.stats.BlockedCount++
			m.stats.mu.Unlock()

			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("File extension %s not allowed", ext),
				"path":    path,
			}, nil
		}
	}

	// 写入文件
	err := m.manager.writer.Write(path, []byte(content), appendMode)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"path":    path,
		"append":  appendMode,
		"size":    len(content),
		"message": "File written successfully",
	}, nil
}

// createFile 创建文件
func (m *FileModule) createFile(path string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 检查是否允许写入
	if !m.config.AllowWrite {
		m.stats.mu.Lock()
		m.stats.BlockedCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   "Write operation not allowed",
			"path":    path,
		}, nil
	}

	// 创建文件
	err := m.manager.writer.Create(path)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"path":    path,
		"message": "File created successfully",
	}, nil
}

// deleteFile 删除文件
func (m *FileModule) deleteFile(path string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 检查是否允许删除
	if !m.config.AllowDelete {
		m.stats.mu.Lock()
		m.stats.BlockedCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   "Delete operation not allowed",
			"path":    path,
		}, nil
	}

	// 删除文件
	err := m.manager.navigator.Delete(path)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success": true,
		"path":    path,
		"message": "File deleted successfully",
	}, nil
}

// listFiles 列出目录中的文件
func (m *FileModule) listFiles(path string, recursive bool) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 列出文件
	files, err := m.manager.navigator.List(path, recursive)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	// 分离文件和目录
	var fileList []map[string]any
	var dirList []map[string]any

	for _, file := range files {
		fileMap := map[string]any{
			"name":     file.Name,
			"path":     file.Path,
			"size":     file.Size,
			"mod_time": file.ModTime,
			"mode":     file.Mode,
		}

		if file.IsDir {
			dirList = append(dirList, fileMap)
		} else {
			fileList = append(fileList, fileMap)
		}
	}

	return map[string]any{
		"success":     true,
		"path":        path,
		"recursive":   recursive,
		"files":       fileList,
		"directories": dirList,
		"total_count": len(files),
	}, nil
}

// getFileInfo 获取文件信息
func (m *FileModule) getFileInfo(path string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 获取文件信息
	info, err := m.manager.navigator.GetInfo(path)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":  true,
		"name":     info.Name,
		"path":     info.Path,
		"size":     info.Size,
		"is_dir":   info.IsDir,
		"mod_time": info.ModTime,
		"mode":     info.Mode,
	}, nil
}

// searchFiles 搜索文件
func (m *FileModule) searchFiles(path, pattern string, recursive bool) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOperations++
	m.stats.mu.Unlock()

	// 搜索文件
	files, err := m.manager.navigator.Search(path, pattern, recursive)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"path":    path,
			"pattern": pattern,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	// 转换为 map 列表
	var fileList []map[string]any
	for _, file := range files {
		fileList = append(fileList, map[string]any{
			"name":     file.Name,
			"path":     file.Path,
			"size":     file.Size,
			"is_dir":   file.IsDir,
			"mod_time": file.ModTime,
			"mode":     file.Mode,
		})
	}

	return map[string]any{
		"success":   true,
		"path":      path,
		"pattern":   pattern,
		"recursive": recursive,
		"files":     fileList,
		"count":     len(files),
	}, nil
}
