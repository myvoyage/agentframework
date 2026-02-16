// Agent Framework - File Service
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// FileService handles file system operations
type FileService struct {
	app *Application
}

// NewFileService creates a new file service
func NewFileService(app *Application) *FileService {
	return &FileService{app: app}
}

// ListFiles lists files in a directory
func (s *FileService) ListFiles(ctx context.Context, path string) ([]*agent.FileInfo, error) {
	return s.app.fileExplorer.ListFiles(ctx, path)
}

// CreateFile creates a new file
func (s *FileService) CreateFile(ctx context.Context, path string, content string) error {
	return s.app.fileExplorer.CreateFile(ctx, path, content)
}

// ReadFile reads a file's content
func (s *FileService) ReadFile(ctx context.Context, path string) (string, error) {
	return s.app.fileExplorer.ReadFile(ctx, path)
}

// WriteFile writes to a file
func (s *FileService) WriteFile(ctx context.Context, path string, content string) error {
	return s.app.fileExplorer.WriteFile(ctx, path, content)
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(ctx context.Context, path string) error {
	return s.app.fileExplorer.DeleteFile(ctx, path)
}

// CreateDirectory creates a new directory
func (s *FileService) CreateDirectory(ctx context.Context, path string) error {
	return s.app.fileExplorer.CreateDirectory(ctx, path)
}

// DeleteDirectory deletes a directory
func (s *FileService) DeleteDirectory(ctx context.Context, path string) error {
	return s.app.fileExplorer.DeleteDirectory(ctx, path)
}

// MoveFile moves a file or directory
func (s *FileService) MoveFile(ctx context.Context, src string, dst string) error {
	return s.app.fileExplorer.MoveFile(ctx, src, dst)
}

// CopyFile copies a file or directory
func (s *FileService) CopyFile(ctx context.Context, src string, dst string) error {
	return s.app.fileExplorer.CopyFile(ctx, src, dst)
}

// GetFileInfo returns information about a file or directory
func (s *FileService) GetFileInfo(ctx context.Context, path string) (*agent.FileInfo, error) {
	return s.app.fileExplorer.GetFileInfo(ctx, path)
}

// UploadFile uploads a file to specified path
func (s *FileService) UploadFile(ctx context.Context, path string, content []byte) error {
	return s.app.fileExplorer.UploadFile(ctx, path, content)
}

// DownloadFile downloads a file from specified path
func (s *FileService) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	return s.app.fileExplorer.DownloadFile(ctx, path)
}

// ListFilesTable prints files in table format
func (s *FileService) ListFilesTable(ctx context.Context, path string, outputFormat string) error {
	files, err := s.ListFiles(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No files found in %s\n", path)
		return nil
	}

	// Print in requested format
	switch outputFormat {
	case "json":
		fmt.Printf("%+v\n", files)
	case "table", "":
		fmt.Printf("Files in %s:\n", path)
		fmt.Println("────────────────────────────────────────────────────────────")
		for _, file := range files {
			fmt.Printf("Name: %s\n", file.Name)
			fmt.Printf("  Type: %s\n", file.Type)
			fmt.Printf("  Size: %d\n", file.Size)
			fmt.Printf("  Modified: %s\n", file.Modified)
			fmt.Println("────────────────────────────────────────────────────────────")
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
