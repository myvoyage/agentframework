// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// FileOperationSkill 文件操作技能
// 提供完整的文件系统操作功能，包括读写、搜索、权限管理等
type FileOperationSkill struct {
	*AdvancedSkill
	config *FileOperationConfig
}

// FileOperationConfig 文件操作配置
type FileOperationConfig struct {
	SandboxDir        string   // 沙箱目录，限制操作范围
	AllowedPaths      []string // 允许访问的路径列表
	MaxFileSize       int64    // 最大文件大小（字节）
	AllowedExts       []string // 允许的文件扩展名
	DeniedExts        []string // 禁止的文件扩展名
	EnableSearch      bool     // 是否启用文件搜索
	EnableCompression bool     // 是否启用压缩功能
}

// NewFileOperationSkill 创建新的文件操作技能
func NewFileOperationSkill(config *FileOperationConfig) (*FileOperationSkill, error) {
	if config == nil {
		// 使用默认配置
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		config = &FileOperationConfig{
			SandboxDir:        wd,
			AllowedPaths:      []string{wd},
			MaxFileSize:       100 * 1024 * 1024, // 100MB
			AllowedExts:       []string{".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".csv", ".log"},
			DeniedExts:        []string{".exe", ".dll", ".so", ".dylib"},
			EnableSearch:      true,
			EnableCompression: true,
		}
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	skill := &FileOperationSkill{
		config: config,
	}

	// 创建高级技能
	skill.AdvancedSkill = NewAdvancedSkill(
		"file_operation",
		"Perform comprehensive file operations including read, write, search, and manage files",
		skill,
	)

	// 更新元数据
	skill.BaseSkill.SetMetadata(SkillMetadata{
		Name:        "file_operation",
		Version:     "2.0.0",
		Author:      "Agent Framework Contributors",
		Description: "Comprehensive file operations (read, write, delete, list, search, compress)",
		Category:    "file",
		Tags:        []string{"file", "io", "read", "write", "search", "compress"},
	})

	return skill, nil
}

// Validate 验证配置
func (c *FileOperationConfig) Validate() error {
	if c.SandboxDir == "" {
		return fmt.Errorf("sandbox_dir cannot be empty")
	}

	if _, err := os.Stat(c.SandboxDir); os.IsNotExist(err) {
		return fmt.Errorf("sandbox_dir does not exist: %s", c.SandboxDir)
	}

	if c.MaxFileSize <= 0 {
		return fmt.Errorf("max_file_size must be positive")
	}

	return nil
}

// Info 返回技能信息
func (s *FileOperationSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.GetName(),
		Desc: s.GetDescription(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {
				Type:     "string",
				Desc:     "Operation type: read, write, delete, list, search, info, copy, move, compress, decompress, mkdir, rmdir, chmod, exists",
				Required: true,
			},
			"path": {
				Type:     "string",
				Desc:     "File or directory path (relative to sandbox directory)",
				Required: false,
			},
			"content": {
				Type:     "string",
				Desc:     "Content to write (for write operation)",
				Required: false,
			},
			"target": {
				Type:     "string",
				Desc:     "Target path (for copy or move operation)",
				Required: false,
			},
			"pattern": {
				Type:     "string",
				Desc:     "Search pattern (for search operation), supports * and ? wildcards",
				Required: false,
			},
			"recursive": {
				Type:     "boolean",
				Desc:     "Recursive search (for search and list operations)",
				Required: false,
			},
			"overwrite": {
				Type:     "boolean",
				Desc:     "Overwrite existing file (for write operation)",
				Required: false,
			},
			"mode": {
				Type:     "string",
				Desc:     "File permissions (for chmod operation), e.g., '0755', '0644'",
				Required: false,
			},
			"max_results": {
				Type:     "integer",
				Desc:     "Maximum results to return (for search operation)",
				Required: false,
			},
		}),
	}, nil
}

// Validate 验证输入
func (s *FileOperationSkill) Validate(ctx context.Context, input string) error {
	var params struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return err
	}

	if params.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	return nil
}

// Prepare 准备执行
func (s *FileOperationSkill) Prepare(ctx context.Context, input string) (*ExecutionContext, error) {
	var params struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.Operation == "" {
		return nil, fmt.Errorf("operation is required")
	}

	execCtx := NewExecutionContext()
	execCtx.SetMetadata("input", input)
	return execCtx, nil
}

// Execute 执行文件操作
func (s *FileOperationSkill) Execute(ctx context.Context, execCtx *ExecutionContext) (interface{}, error) {
	// 从上下文获取输入
	input, ok := execCtx.GetMetadata("input")
	if !ok {
		return nil, fmt.Errorf("input not found in execution context")
	}

	inputStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("invalid input type")
	}

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	operation := params["operation"].(string)
	if operation == "" {
		return nil, fmt.Errorf("operation is required")
	}

	// 根据操作类型执行
	switch operation {
	case "read":
		return s.readFile(ctx, params, execCtx)
	case "write":
		return s.writeFile(ctx, params, execCtx)
	case "delete":
		return s.deleteFile(ctx, params, execCtx)
	case "list":
		return s.listDirectory(ctx, params, execCtx)
	case "search":
		return s.searchFiles(ctx, params, execCtx)
	case "info":
		return s.getFileInfo(ctx, params, execCtx)
	case "copy":
		return s.copyFile(ctx, params, execCtx)
	case "move":
		return s.moveFile(ctx, params, execCtx)
	case "mkdir":
		return s.createDirectory(ctx, params, execCtx)
	case "rmdir":
		return s.removeDirectory(ctx, params, execCtx)
	case "exists":
		return s.fileExists(ctx, params, execCtx)
	case "chmod":
		return s.changeMode(ctx, params, execCtx)
	case "compress":
		return s.compressFile(ctx, params, execCtx)
	case "decompress":
		return s.decompressFile(ctx, params, execCtx)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// readFile 读取文件
func (s *FileOperationSkill) readFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	// 检查文件大小
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > s.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", info.Size(), s.config.MaxFileSize)
	}

	// 读取文件
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 检查是否为文本文件
	if !isTextFile(content) {
		return nil, fmt.Errorf("file appears to be binary, use appropriate handler")
	}

	return map[string]interface{}{
		"success":  true,
		"path":     params["path"],
		"size":     len(content),
		"content":  string(content),
		"modified": info.ModTime(),
		"mode":     info.Mode().String(),
	}, nil
}

// writeFile 写入文件
func (s *FileOperationSkill) writeFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	content := params["content"].(string)
	overwrite := true
	if ow, ok := params["overwrite"].(bool); ok {
		overwrite = ow
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); err == nil && !overwrite {
		return nil, fmt.Errorf("file already exists (use overwrite=true to overwrite)")
	}

	// 写入文件
	mode := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !overwrite {
		mode |= os.O_EXCL
	}

	file, err := os.OpenFile(path, mode, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"size":    len(content),
		"action":  "written",
	}, nil
}

// deleteFile 删除文件
func (s *FileOperationSkill) deleteFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	// 检查是否为目录
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return nil, fmt.Errorf("path is a directory, use rmdir operation")
		}
	}

	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"action":  "deleted",
	}, nil
}

// listDirectory 列出目录内容
func (s *FileOperationSkill) listDirectory(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	recursive := false
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	var files []map[string]interface{}
	if recursive {
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(s.config.SandboxDir, filePath)
			if err != nil {
				return err
			}

			files = append(files, map[string]interface{}{
				"name":     info.Name(),
				"path":     relPath,
				"size":     info.Size(),
				"mode":     info.Mode().String(),
				"modified": info.ModTime(),
				"is_dir":   info.IsDir(),
			})

			return nil
		})
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, map[string]interface{}{
				"name":     entry.Name(),
				"path":     filepath.Join(params["path"].(string), entry.Name()),
				"size":     info.Size(),
				"mode":     info.Mode().String(),
				"modified": info.ModTime(),
				"is_dir":   entry.IsDir(),
			})
		}
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":   true,
		"path":      params["path"],
		"count":     len(files),
		"files":     files,
		"recursive": recursive,
	}, nil
}

// searchFiles 搜索文件
func (s *FileOperationSkill) searchFiles(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	pattern := params["pattern"].(string)
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required for search operation")
	}

	recursive := true
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	maxResults := 100
	if mr, ok := params["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	var results []map[string]interface{}
	count := 0

	err := filepath.Walk(s.config.SandboxDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 跳过隐藏文件和目录
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查模式匹配
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return nil
		}

		if matched {
			relPath, err := filepath.Rel(s.config.SandboxDir, path)
			if err != nil {
				return err
			}

			results = append(results, map[string]interface{}{
				"name":     info.Name(),
				"path":     relPath,
				"size":     info.Size(),
				"modified": info.ModTime(),
				"is_dir":   info.IsDir(),
			})
			count++

			// 达到最大结果数
			if count >= maxResults {
				return io.EOF
			}
		}

		// 如果不递归且是目录，跳过
		if !recursive && info.IsDir() && path != s.config.SandboxDir {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil && err != io.EOF {
		return nil, err
	}

	return map[string]interface{}{
		"success":   true,
		"pattern":   pattern,
		"count":     len(results),
		"results":   results,
		"recursive": recursive,
		"truncated": count >= maxResults,
	}, nil
}

// getFileInfo 获取文件信息
func (s *FileOperationSkill) getFileInfo(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"path":       params["path"],
		"name":       info.Name(),
		"size":       info.Size(),
		"mode":       info.Mode().String(),
		"mod_time":   info.ModTime(),
		"is_dir":     info.IsDir(),
		"is_symlink": false,
	}, nil
}

// copyFile 复制文件
func (s *FileOperationSkill) copyFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	srcPath, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	dstPath, err := s.resolvePath(params["target"].(string))
	if err != nil {
		return nil, err
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// 复制内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"source":  params["path"],
		"target":  params["target"],
		"action":  "copied",
	}, nil
}

// moveFile 移动文件
func (s *FileOperationSkill) moveFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	srcPath, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	dstPath, err := s.resolvePath(params["target"].(string))
	if err != nil {
		return nil, err
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"source":  params["path"],
		"target":  params["target"],
		"action":  "moved",
	}, nil
}

// createDirectory 创建目录
func (s *FileOperationSkill) createDirectory(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"action":  "created",
	}, nil
}

// removeDirectory 删除目录
func (s *FileOperationSkill) removeDirectory(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(path); err != nil {
		return nil, fmt.Errorf("failed to remove directory: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"action":  "removed",
	}, nil
}

// fileExists 检查文件是否存在
func (s *FileOperationSkill) fileExists(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
	exists := err == nil

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"exists":  exists,
	}, nil
}

// changeMode 修改文件权限
func (s *FileOperationSkill) changeMode(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	path, err := s.resolvePath(params["path"].(string))
	if err != nil {
		return nil, err
	}

	modeStr := params["mode"].(string)
	if modeStr == "" {
		return nil, fmt.Errorf("mode is required for chmod operation")
	}

	// 解析权限模式
	var mode uint64
	_, err = fmt.Sscanf(modeStr, "%o", &mode)
	if err != nil {
		return nil, fmt.Errorf("invalid mode format: %w", err)
	}

	if err := os.Chmod(path, os.FileMode(mode)); err != nil {
		return nil, fmt.Errorf("failed to change mode: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"path":    params["path"],
		"mode":    modeStr,
		"action":  "chmod",
	}, nil
}

// compressFile 压缩文件
func (s *FileOperationSkill) compressFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	// 获取参数
	sourcePath, ok := params["source"].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf("source path is required")
	}

	targetPath, ok := params["target"].(string)
	if !ok || targetPath == "" {
		// 默认在源路径后加 .zip
		targetPath = sourcePath + ".zip"
	}

	format, ok := params["format"].(string)
	if !ok {
		format = "zip" // 默认使用 zip 格式
	}

	// 解析路径
	resolvedSource, err := s.resolvePath(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	resolvedTarget, err := s.resolvePath(targetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid target path: %w", err)
	}

	// 根据格式压缩
	switch format {
	case "zip", ".zip":
		return s.compressToZip(resolvedSource, resolvedTarget)
	case "tar", "tar.gz", ".tar.gz":
		return s.compressToTarGz(resolvedSource, resolvedTarget)
	case "gzip", ".gz":
		return s.compressToGzip(resolvedSource, resolvedTarget)
	default:
		return nil, fmt.Errorf("unsupported compression format: %s", format)
	}
}

// decompressFile 解压文件
func (s *FileOperationSkill) decompressFile(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	// 获取参数
	sourcePath, ok := params["source"].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf("source path is required")
	}

	targetPath, ok := params["target"].(string)
	if !ok {
		// 如果没有指定目标路径，解压到与源文件同名的目录
		if strings.HasSuffix(sourcePath, ".zip") {
			targetPath = strings.TrimSuffix(sourcePath, ".zip")
		} else if strings.HasSuffix(sourcePath, ".tar.gz") {
			targetPath = strings.TrimSuffix(sourcePath, ".tar.gz")
		} else if strings.HasSuffix(sourcePath, ".gz") {
			targetPath = strings.TrimSuffix(sourcePath, ".gz")
		} else {
			targetPath = sourcePath + "_extracted"
		}
	}

	// 解析路径
	resolvedSource, err := s.resolvePath(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	resolvedTarget, err := s.resolvePath(targetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid target path: %w", err)
	}

	// 检测压缩格式并解压
	if strings.HasSuffix(resolvedSource, ".zip") {
		return s.decompressZip(resolvedSource, resolvedTarget)
	} else if strings.HasSuffix(resolvedSource, ".tar.gz") || strings.HasSuffix(resolvedSource, ".tgz") {
		return s.decompressTarGz(resolvedSource, resolvedTarget)
	} else if strings.HasSuffix(resolvedSource, ".gz") {
		return s.decompressGzip(resolvedSource, resolvedTarget)
	} else {
		return nil, fmt.Errorf("unsupported archive format")
	}
}

// compressToZip 压缩为 ZIP 格式
func (s *FileOperationSkill) compressToZip(source, target string) (map[string]interface{}, error) {
	// 创建目标文件
	targetFile, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	zipWriter := zip.NewWriter(targetFile)
	defer zipWriter.Close()

	// 如果是文件，直接添加
	fileInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source: %w", err)
	}

	if fileInfo.IsDir() {
		// 递归添加目录
		err = filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 创建 ZIP 文件路径（相对路径）
			relPath, err := filepath.Rel(source, filePath)
			if err != nil {
				return err
			}

			// 跳过目录本身
			if info.IsDir() {
				return nil
			}

			// 创建文件头
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// 设置正确的文件名（相对于压缩包根目录）
			header.Name = relPath

			header.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			// 打开源文件
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// 写入文件内容
			_, err = io.Copy(writer, file)
			return err
		})

		if err != nil {
			return nil, fmt.Errorf("failed to add files to zip: %w", err)
		}
	} else {
		// 添加单个文件
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return nil, err
		}

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return nil, err
		}

		file, err := os.Open(source)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return nil, fmt.Errorf("failed to write file to zip: %w", err)
		}
	}

	return map[string]interface{}{
		"source":      source,
		"target":      target,
		"size":        fileInfo.Size(),
		"compressed":  "zip",
		"status":      "success",
		"message":     fmt.Sprintf("Compressed %s to %s", source, target),
	}, nil
}

// compressToTarGz 压缩为 tar.gz 格式
func (s *FileOperationSkill) compressToTarGz(source, target string) (map[string]interface{}, error) {
	// 创建目标文件
	targetFile, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// 创建 gzip writer
	gzipWriter := gzip.NewWriter(targetFile)
	defer gzipWriter.Close()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 获取源文件信息
	fileInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source: %w", err)
	}

	if fileInfo.IsDir() {
		// 递归添加目录
		err = filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 创建相对路径
			relPath, err := filepath.Rel(source, filePath)
			if err != nil {
				return err
			}

			// 跳过目录本身
			if info.IsDir() {
				return nil
			}

			// 创建 tar 头
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}

			// 设置正确的文件名（相对于压缩包根目录）
			header.Name = relPath

			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// 写入文件内容
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		})

		if err != nil {
			return nil, fmt.Errorf("failed to add files to tar: %w", err)
		}
	} else {
		// 添加单个文件
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return nil, err
		}

		// 设置正确的文件名
		header.Name = filepath.Base(source)

		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, err
		}

		file, err := os.Open(source)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return nil, fmt.Errorf("failed to write file to tar: %w", err)
		}
	}

	return map[string]interface{}{
		"source":      source,
		"target":      target,
		"size":        fileInfo.Size(),
		"compressed":  "tar.gz",
		"status":      "success",
		"message":     fmt.Sprintf("Compressed %s to %s", source, target),
	}, nil
}

// compressToGzip 压缩为 gzip 格式
func (s *FileOperationSkill) compressToGzip(source, target string) (map[string]interface{}, error) {
	// 确保目标以 .gz 结尾
	if !strings.HasSuffix(target, ".gz") {
		target = target + ".gz"
	}

	// 打开源文件
	sourceFile, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// 获取文件信息
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat source: %w", err)
	}

	// 创建目标文件
	targetFile, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// 创建 gzip writer
	gzipWriter := gzip.NewWriter(targetFile)
	defer gzipWriter.Close()

	// 复制数据
	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}

	return map[string]interface{}{
		"source":      source,
		"target":      target,
		"size":        fileInfo.Size(),
		"compressed":  "gzip",
		"status":      "success",
		"message":     fmt.Sprintf("Compressed %s to %s", source, target),
	}, nil
}

// decompressZip 解压 ZIP 文件
func (s *FileOperationSkill) decompressZip(source, target string) (map[string]interface{}, error) {
	// 打开 ZIP 文件
	zipReader, err := zip.OpenReader(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// 创建目标目录
	if err := os.MkdirAll(target, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	extractedCount := 0
	totalSize := int64(0)

	// 解压每个文件
	for _, file := range zipReader.File {
		// 创建目标文件路径
		targetPath := filepath.Join(target, file.Name)

		// 确保目录存在
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, err
			}
			continue
		}

		// 创建文件
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return nil, err
		}

		targetFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return nil, fmt.Errorf("failed to create target file: %w", err)
		}

		// 打开压缩文件中的文件
		rc, err := file.Open()
		if err != nil {
			targetFile.Close()
			return nil, fmt.Errorf("failed to open file in zip: %w", err)
		}

		// 复制数据
		size, err := io.Copy(targetFile, rc)
		if err != nil {
			rc.Close()
			targetFile.Close()
			return nil, fmt.Errorf("failed to extract file: %w", err)
		}

		rc.Close()
		targetFile.Close()

		extractedCount++
		totalSize += size
	}

	return map[string]interface{}{
		"source":        source,
		"target":        target,
		"extracted_files": extractedCount,
		"total_size":     totalSize,
		"format":         "zip",
		"status":         "success",
		"message":        fmt.Sprintf("Extracted %d files from %s to %s", extractedCount, source, target),
	}, nil
}

// decompressTarGz 解压 tar.gz 文件
func (s *FileOperationSkill) decompressTarGz(source, target string) (map[string]interface{}, error) {
	// 打开源文件
	sourceFile, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer sourceFile.Close()

	// 创建 gzip reader
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// 创建 tar reader
	tarReader := tar.NewReader(gzipReader)

	// 创建目标目录
	if err := os.MkdirAll(target, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	extractedCount := 0
	totalSize := int64(0)

	// 解压每个文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// 创建目标路径
		targetPath := filepath.Join(target, header.Name)

		if header.Typeflag == tar.TypeDir {
			// 创建目录
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, err
			}
			continue
		}

		// 创建文件
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return nil, err
		}

		targetFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return nil, fmt.Errorf("failed to create target file: %w", err)
		}

		// 写入数据
		size, err := io.Copy(targetFile, tarReader)
		if err != nil {
			targetFile.Close()
			return nil, fmt.Errorf("failed to extract file: %w", err)
		}

		targetFile.Close()
		extractedCount++
		totalSize += size
	}

	return map[string]interface{}{
		"source":        source,
		"target":        target,
		"extracted_files": extractedCount,
		"total_size":     totalSize,
		"format":         "tar.gz",
		"status":         "success",
		"message":        fmt.Sprintf("Extracted %d files from %s to %s", extractedCount, source, target),
	}, nil
}

// decompressGzip 解压 gzip 文件
func (s *FileOperationSkill) decompressGzip(source, target string) (map[string]interface{}, error) {
	// 打开源文件
	sourceFile, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip file: %w", err)
	}
	defer sourceFile.Close()

	// 创建 gzip reader
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// 创建目标文件
	targetFile, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// 解压数据
	size, err := io.Copy(targetFile, gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return map[string]interface{}{
		"source":     source,
		"target":     target,
		"size":       size,
		"format":     "gzip",
		"status":     "success",
		"message":    fmt.Sprintf("Decompressed %s to %s", source, target),
	}, nil
}

// resolvePath 解析并验证路径
func (s *FileOperationSkill) resolvePath(inputPath string) (string, error) {
	if strings.Contains(inputPath, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", inputPath)
	}

	// 清理路径
	cleanPath := filepath.Clean(inputPath)

	// 如果是相对路径，基于沙箱目录
	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(s.config.SandboxDir, cleanPath)
	}

	// 验证路径是否在允许的范围内
	allowed := false
	for _, allowedPath := range s.config.AllowedPaths {
		if strings.HasPrefix(cleanPath, allowedPath) {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("access denied: path not in allowed list")
	}

	return cleanPath, nil
}

// isTextFile 检查文件是否为文本文件
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}

	// 检查前1000字节
	limit := 1000
	if len(content) < limit {
		limit = len(content)
	}

	content = content[:limit]

	// 检查是否包含空字节（二进制文件标志）
	for _, b := range content {
		if b == 0 {
			return false
		}
	}

	return true
}

// Cleanup 清理资源
func (s *FileOperationSkill) Cleanup(ctx context.Context, execCtx *ExecutionContext) error {
	// 清理临时文件等
	return nil
}
