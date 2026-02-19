// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package fileops provides file operation tools for the AgentFramework
// This file implements the file read tool

package fileops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/schema"
)

// ReadTool reads the contents of a file
type ReadTool struct {
	baseDir string // Base directory for file operations (sandbox restriction)
}

// NewReadTool creates a new file read tool
func NewReadTool(baseDir string) *ReadTool {
	if baseDir == "" {
		baseDir = "." // Default to current directory
	}
	return &ReadTool{baseDir: baseDir}
}

// Read reads the contents of a file
func (t *ReadTool) Read(ctx context.Context, filePath string) (string, error) {
	// Resolve the full path
	fullPath := filepath.Join(t.baseDir, filePath)

	// Security check: ensure the path is within the base directory
	relPath, err := filepath.Rel(t.baseDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	// Prevent directory traversal attacks
	if relPath == ".." || len(relPath) >= 3 && relPath[0:3] == "../" {
		return "", fmt.Errorf("access denied: path outside base directory")
	}

	// Read the file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// Info returns the tool information for Eino integration
func (t *ReadTool) Info() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "file_read",
		Desc: "Reads the contents of a file from the file system",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"file_path": {
				Desc:     "Path to the file to read",
				Type:     "string",
				Required: true,
			},
		}),
	}
}
