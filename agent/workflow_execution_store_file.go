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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileWorkflowExecutionStore stores workflow execution results in files
type FileWorkflowExecutionStore struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileWorkflowExecutionStore creates a new file-based workflow execution store
func NewFileWorkflowExecutionStore(baseDir string) (*FileWorkflowExecutionStore, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileWorkflowExecutionStore{
		baseDir: baseDir,
	}, nil
}

// SaveExecutionResult saves a workflow execution result to a file
func (store *FileWorkflowExecutionStore) SaveExecutionResult(ctx context.Context, result *WorkflowExecutionResult) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Create directory structure: baseDir/workflowID/
	workflowDir := filepath.Join(store.baseDir, result.WorkflowID)
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	// Create file path: baseDir/workflowID/executionID.json
	filePath := filepath.Join(workflowDir, result.ExecutionID+".json")

	// Marshal result to JSON with indentation for readability
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal execution result: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write execution result: %w", err)
	}

	return nil
}

// GetExecutionResult gets a workflow execution result by execution ID from a file
func (store *FileWorkflowExecutionStore) GetExecutionResult(ctx context.Context, executionID string) (*WorkflowExecutionResult, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Search for the execution ID in all workflow directories
	entries, err := os.ReadDir(store.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow directories: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		filePath := filepath.Join(store.baseDir, entry.Name(), executionID+".json")
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read execution result: %w", err)
			}

			var result WorkflowExecutionResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal execution result: %w", err)
			}

			return &result, nil
		}
	}

	return nil, fmt.Errorf("execution result not found: %s", executionID)
}

// GetExecutionResultsByWorkflowID gets all execution results for a workflow from files
func (store *FileWorkflowExecutionStore) GetExecutionResultsByWorkflowID(ctx context.Context, workflowID string) ([]*WorkflowExecutionResult, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	workflowDir := filepath.Join(store.baseDir, workflowID)

	// Check if workflow directory exists
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		return []*WorkflowExecutionResult{}, nil
	}

	// List all files in the workflow directory
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list execution files: %w", err)
	}

	var results []*WorkflowExecutionResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(workflowDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var result WorkflowExecutionResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue // Skip files that can't be unmarshaled
		}

		results = append(results, &result)
	}

	return results, nil
}

// DeleteExecutionResult deletes a workflow execution result file
func (store *FileWorkflowExecutionStore) DeleteExecutionResult(ctx context.Context, workflowID, executionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	filePath := filepath.Join(store.baseDir, workflowID, executionID+".json")

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("execution result not found: %s", executionID)
		}
		return fmt.Errorf("failed to delete execution result: %w", err)
	}

	return nil
}

// DeleteWorkflowResults deletes all execution results for a workflow
func (store *FileWorkflowExecutionStore) DeleteWorkflowResults(ctx context.Context, workflowID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	workflowDir := filepath.Join(store.baseDir, workflowID)

	if err := os.RemoveAll(workflowDir); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete workflow results: %w", err)
	}

	return nil
}

// GetAllExecutionResults gets all execution results across all workflows
func (store *FileWorkflowExecutionStore) GetAllExecutionResults(ctx context.Context) ([]*WorkflowExecutionResult, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	entries, err := os.ReadDir(store.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow directories: %w", err)
	}

	var allResults []*WorkflowExecutionResult

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		results, err := store.GetExecutionResultsByWorkflowID(ctx, entry.Name())
		if err != nil {
			continue // Skip workflows that can't be read
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}
