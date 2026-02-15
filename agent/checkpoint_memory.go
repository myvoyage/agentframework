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
	"sort"
	"sync"
	"time"
)

// MemoryCheckpointStoreConfig contains configuration options for the MemoryCheckpointStore
type MemoryCheckpointStoreConfig struct {
	MaxCheckpoints  int           // Maximum number of checkpoints to keep
	TTL             time.Duration // Time-to-live for checkpoints
	CleanupInterval time.Duration // Interval for cleaning up expired checkpoints
}

// MemoryCheckpointStore for testing or default
type MemoryCheckpointStore struct {
	data        map[string]*Checkpoint
	mu          sync.RWMutex
	config      MemoryCheckpointStoreConfig
	stopCleanup chan struct{}
}

// DefaultMemoryCheckpointStoreConfig returns the default configuration for MemoryCheckpointStore
func DefaultMemoryCheckpointStoreConfig() MemoryCheckpointStoreConfig {
	return MemoryCheckpointStoreConfig{
		MaxCheckpoints:  1000,
		TTL:             24 * time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
}

// NewMemoryCheckpointStore creates a new MemoryCheckpointStore with default configuration
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return NewMemoryCheckpointStoreWithConfig(DefaultMemoryCheckpointStoreConfig())
}

// NewMemoryCheckpointStoreWithConfig creates a new MemoryCheckpointStore with custom configuration
func NewMemoryCheckpointStoreWithConfig(config MemoryCheckpointStoreConfig) *MemoryCheckpointStore {
	// Set default values if not provided
	if config.MaxCheckpoints <= 0 {
		config.MaxCheckpoints = 1000
	}
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 10 * time.Minute
	}

	store := &MemoryCheckpointStore{
		data:        make(map[string]*Checkpoint),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start periodic cleanup goroutine
	go store.periodicCleanup()

	return store
}

func (s *MemoryCheckpointStore) Save(ctx context.Context, cp *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy to simulate persistence
	// Create a copy to avoid external modification
	cpCopy := *cp
	cpCopy.UpdatedAt = time.Now()
	if cpCopy.CreatedAt.IsZero() {
		cpCopy.CreatedAt = cpCopy.UpdatedAt
	}

	// Add the new checkpoint
	isNewCheckpoint := s.data[cpCopy.RunID] == nil
	s.data[cpCopy.RunID] = &cpCopy

	// If this is a new checkpoint and we're over the limit, remove the oldest one
	if isNewCheckpoint && len(s.data) > s.config.MaxCheckpoints {
		// Find the oldest checkpoint
		var oldestID string
		var oldestTime time.Time
		for id, cp := range s.data {
			if oldestID == "" || cp.UpdatedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = cp.UpdatedAt
			}
		}
		if oldestID != "" {
			delete(s.data, oldestID)
		}
	}

	return nil
}

func (s *MemoryCheckpointStore) Load(ctx context.Context, runID string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp, ok := s.data[runID]
	if !ok {
		return nil, fmt.Errorf("checkpoint not found: %s", runID)
	}
	// Return a copy to avoid external modification
	cpCopy := *cp
	return &cpCopy, nil
}

func (s *MemoryCheckpointStore) List(ctx context.Context, options ...ListCheckpointOption) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	opts := &ListCheckpointOptions{}
	for _, option := range options {
		option(opts)
	}

	var list []*Checkpoint
	for _, cp := range s.data {
		// Apply filters
		if opts.WorkflowName != "" && cp.WorkflowName != opts.WorkflowName {
			continue
		}
		if opts.Status != "" && cp.Status != opts.Status {
			continue
		}
		// Add a copy to avoid external modification
		cpCopy := *cp
		list = append(list, &cpCopy)
	}

	// Apply pagination
	if opts.Limit > 0 {
		if opts.Offset >= len(list) {
			return []*Checkpoint{}, nil
		}
		end := opts.Offset + opts.Limit
		if end > len(list) {
			end = len(list)
		}
		list = list[opts.Offset:end]
	}

	return list, nil
}

func (s *MemoryCheckpointStore) Delete(ctx context.Context, runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, runID)
	return nil
}

func (s *MemoryCheckpointStore) HealthCheck(ctx context.Context) error {
	// Memory store is always healthy
	return nil
}

// periodicCleanup removes expired and old checkpoints at regular intervals
func (s *MemoryCheckpointStore) periodicCleanup() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes expired and old checkpoints
func (s *MemoryCheckpointStore) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Remove expired checkpoints first
	for id, cp := range s.data {
		if now.Sub(cp.UpdatedAt) > s.config.TTL {
			delete(s.data, id)
		}
	}

	// If still over the limit, remove oldest checkpoints
	if len(s.data) > s.config.MaxCheckpoints {
		// Sort checkpoints by UpdatedAt
		type checkpointEntry struct {
			id string
			cp *Checkpoint
		}
		var entries []checkpointEntry
		for id, cp := range s.data {
			entries = append(entries, checkpointEntry{id: id, cp: cp})
		}

		// Sort by UpdatedAt, oldest first
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].cp.UpdatedAt.Before(entries[j].cp.UpdatedAt)
		})

		// Remove oldest checkpoints until we're under the limit
		removeCount := len(s.data) - s.config.MaxCheckpoints
		for i := 0; i < removeCount; i++ {
			delete(s.data, entries[i].id)
		}
	}
}

func (s *MemoryCheckpointStore) Close(ctx context.Context) error {
	// Stop the cleanup goroutine
	close(s.stopCleanup)

	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.data)
	return nil
}
