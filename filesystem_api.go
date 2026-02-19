// Agent Framework - File System API
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"

	"AgentFramework/agent"
)

// ListFiles lists files in a directory
func (a *App) ListFiles(path string) ([]*agent.FileInfo, error) {
	return a.core.GetFileExplorer().ListFiles(a.ctx, path)
}

// CreateFile creates a new file
func (a *App) CreateFile(path string, content string) error {
	return a.core.GetFileExplorer().CreateFile(a.ctx, path, content)
}

// ReadFile reads a file's content
func (a *App) ReadFile(path string) (string, error) {
	return a.core.GetFileExplorer().ReadFile(a.ctx, path)
}

// WriteFile writes to a file
func (a *App) WriteFile(path string, content string) error {
	return a.core.GetFileExplorer().WriteFile(a.ctx, path, content)
}

// DeleteFile deletes a file
func (a *App) DeleteFile(path string) error {
	return a.core.GetFileExplorer().DeleteFile(a.ctx, path)
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(path string) error {
	return a.core.GetFileExplorer().CreateDirectory(a.ctx, path)
}

// DeleteDirectory deletes a directory
func (a *App) DeleteDirectory(path string) error {
	return a.core.GetFileExplorer().DeleteDirectory(a.ctx, path)
}

// MoveFile moves a file or directory
func (a *App) MoveFile(src string, dst string) error {
	return a.core.GetFileExplorer().MoveFile(a.ctx, src, dst)
}

// CopyFile copies a file or directory
func (a *App) CopyFile(src string, dst string) error {
	return a.core.GetFileExplorer().CopyFile(a.ctx, src, dst)
}

// GetFileInfo returns information about a file or directory
func (a *App) GetFileInfo(path string) (*agent.FileInfo, error) {
	return a.core.GetFileExplorer().GetFileInfo(a.ctx, path)
}

// UploadFile uploads a file to the specified path
func (a *App) UploadFile(path string, content []byte) error {
	return a.core.GetFileExplorer().UploadFile(a.ctx, path, content)
}

// DownloadFile downloads a file from the specified path
func (a *App) DownloadFile(path string) ([]byte, error) {
	return a.core.GetFileExplorer().DownloadFile(a.ctx, path)
}
