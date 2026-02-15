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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FileCheckpointStore implements CheckpointStore using the local filesystem.
type FileCheckpointStore struct {
	dir string
	mu  sync.RWMutex
}

// NewFileCheckpointStore creates a new FileCheckpointStore instance.
func NewFileCheckpointStore(dir string) (*FileCheckpointStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileCheckpointStore{dir: dir}, nil
}

func (s *FileCheckpointStore) Save(ctx context.Context, cp *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp.UpdatedAt = time.Now()
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = cp.UpdatedAt
	}

	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, cp.RunID+".json")
	return os.WriteFile(path, data, 0644)
}

func (s *FileCheckpointStore) Load(ctx context.Context, runID string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dir, runID+".json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("checkpoint not found: %s", runID)
	}
	if err != nil {
		return nil, err
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

func (s *FileCheckpointStore) List(ctx context.Context, options ...ListCheckpointOption) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	opts := &ListCheckpointOptions{}
	for _, option := range options {
		option(opts)
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var checkpoints []*Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip error files
		}

		var cp Checkpoint
		if err := json.Unmarshal(data, &cp); err != nil {
			continue
		}

		// Apply filters
		if opts.WorkflowName != "" && cp.WorkflowName != opts.WorkflowName {
			continue
		}
		if opts.Status != "" && cp.Status != opts.Status {
			continue
		}

		checkpoints = append(checkpoints, &cp)
	}

	// Apply pagination and sorting (simplified for file store)
	if opts.Limit > 0 {
		if opts.Offset >= len(checkpoints) {
			return []*Checkpoint{}, nil
		}
		end := opts.Offset + opts.Limit
		if end > len(checkpoints) {
			end = len(checkpoints)
		}
		checkpoints = checkpoints[opts.Offset:end]
	}

	return checkpoints, nil
}

func (s *FileCheckpointStore) Delete(ctx context.Context, runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, runID+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Ignore if file doesn't exist
		}
		return err
	}
	return nil
}

func (s *FileCheckpointStore) HealthCheck(ctx context.Context) error {
	// Check if directory is writable
	tmpID := uuid.NewString()
	tmpCP := &Checkpoint{
		RunID:        tmpID,
		WorkflowName: "healthcheck",
		Status:       StatusRunning,
		Input:        "healthcheck",
	}

	// Try to save a temporary checkpoint
	if err := s.Save(ctx, tmpCP); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Try to load it back
	if _, err := s.Load(ctx, tmpID); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Clean up
	return s.Delete(ctx, tmpID)
}

func (s *FileCheckpointStore) Close(ctx context.Context) error {
	// No resources to close for FileCheckpointStore
	return nil
}
