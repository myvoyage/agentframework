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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// FileType represents the type of a file
type FileType string

const (
	// FileTypeFile represents a regular file
	FileTypeFile FileType = "file"
	// FileTypeDirectory represents a directory
	FileTypeDirectory FileType = "directory"
	// FileTypeLink represents a symbolic link
	FileTypeLink FileType = "link"
)

// FileInfo contains information about a file or directory
type FileInfo struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     FileType    `json:"type"`
	Size     int64       `json:"size"`
	Modified time.Time   `json:"modified"`
	Created  time.Time   `json:"created"`
	Mode     fs.FileMode `json:"mode"`
}

// FileExplorer provides file system operations
type FileExplorer struct {
	// Add any fields needed for file exploration
}

// NewFileExplorer creates a new FileExplorer
func NewFileExplorer() *FileExplorer {
	return &FileExplorer{}
}

// Init initializes the file explorer
func (fe *FileExplorer) Init(ctx context.Context) {
	// Initialize any resources needed for file exploration
}

// ListFiles lists files in a directory
func (fe *FileExplorer) ListFiles(ctx context.Context, path string) ([]*FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in directory %q: %w", path, err)
	}

	var files []*FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// Skip entries we can't get info for, but continue with other entries
			continue
		}

		fileType := FileTypeFile
		if entry.IsDir() {
			fileType = FileTypeDirectory
		}

		file := &FileInfo{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Type:     fileType,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Created:  info.ModTime(), // On Windows, this is the same as ModTime
			Mode:     info.Mode(),
		}

		files = append(files, file)
	}

	return files, nil
}

// CreateFile creates a new file
func (fe *FileExplorer) CreateFile(ctx context.Context, path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}
	return nil
}

// ReadFile reads a file's content
func (fe *FileExplorer) ReadFile(ctx context.Context, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return string(content), nil
}

// WriteFile writes to a file
func (fe *FileExplorer) WriteFile(ctx context.Context, path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write to file %q: %w", path, err)
	}
	return nil
}

// DeleteFile deletes a file
func (fe *FileExplorer) DeleteFile(ctx context.Context, path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %q: %w", path, err)
	}
	return nil
}

// CreateDirectory creates a new directory
func (fe *FileExplorer) CreateDirectory(ctx context.Context, path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", path, err)
	}
	return nil
}

// DeleteDirectory deletes a directory
func (fe *FileExplorer) DeleteDirectory(ctx context.Context, path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory %q: %w", path, err)
	}
	return nil
}

// MoveFile moves a file or directory
func (fe *FileExplorer) MoveFile(ctx context.Context, src string, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move %q to %q: %w", src, dst, err)
	}
	return nil
}

// CopyFile copies a file or directory
func (fe *FileExplorer) CopyFile(ctx context.Context, src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get file info for %q: %w", src, err)
	}

	if info.IsDir() {
		return fe.copyDirectory(ctx, src, dst)
	}

	return fe.copyRegularFile(ctx, src, dst)
}

// copyRegularFile copies a regular file
func (fe *FileExplorer) copyRegularFile(ctx context.Context, src string, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %q: %w", src, err)
	}

	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file %q: %w", dst, err)
	}
	return nil
}

// copyDirectory copies a directory recursively
func (fe *FileExplorer) copyDirectory(ctx context.Context, src string, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %q: %w", dst, err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory %q: %w", src, err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := fe.copyDirectory(ctx, srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %q to %q: %w", srcPath, dstPath, err)
			}
		} else {
			if err := fe.copyRegularFile(ctx, srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %q to %q: %w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

// GetFileInfo returns information about a file or directory
func (fe *FileExplorer) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %q: %w", path, err)
	}

	fileType := FileTypeFile
	if info.IsDir() {
		fileType = FileTypeDirectory
	}

	file := &FileInfo{
		Name:     filepath.Base(path),
		Path:     path,
		Type:     fileType,
		Size:     info.Size(),
		Modified: info.ModTime(),
		Created:  info.ModTime(), // On Windows, this is the same as ModTime
		Mode:     info.Mode(),
	}

	return file, nil
}

// UploadFile uploads a file to the specified path
func (fe *FileExplorer) UploadFile(ctx context.Context, path string, content []byte) error {
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to upload file to %q: %w", path, err)
	}
	return nil
}

// DownloadFile downloads a file from the specified path
func (fe *FileExplorer) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from %q: %w", path, err)
	}
	return content, nil
}
