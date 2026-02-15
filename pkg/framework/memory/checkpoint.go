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

package memory

import (
	"context"
	"time"
)

// CheckpointStatus represents the status of a workflow run.
type CheckpointStatus string

const (
	StatusRunning   CheckpointStatus = "running"
	StatusSuspended CheckpointStatus = "suspended"
	StatusCompleted CheckpointStatus = "completed"
	StatusFailed    CheckpointStatus = "failed"
)

// Checkpoint represents a snapshot of a workflow execution.
type Checkpoint struct {
	RunID        string            `json:"run_id"`
	WorkflowName string            `json:"workflow_name"`
	Status       CheckpointStatus  `json:"status"`
	State        *WorkflowState    `json:"state"`
	Input        string            `json:"input"`
	Output       string            `json:"output,omitempty"`
	Progress     float64           `json:"progress"` // 0.0 to 1.0
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// CheckpointStore defines the interface for persisting checkpoints.
type CheckpointStore interface {
	Save(ctx context.Context, cp *Checkpoint) error
	Load(ctx context.Context, runID string) (*Checkpoint, error)
	List(ctx context.Context, options ...ListCheckpointOption) ([]*Checkpoint, error)
	Delete(ctx context.Context, runID string) error
	HealthCheck(ctx context.Context) error
	Close(ctx context.Context) error
}

// ListCheckpointOption defines options for listing checkpoints.
type ListCheckpointOption func(*ListCheckpointOptions)

// ListCheckpointOptions contains options for listing checkpoints.
type ListCheckpointOptions struct {
	WorkflowName string
	Status       CheckpointStatus
	Limit        int
	Offset       int
	SortBy       string
	SortDesc     bool
}

// WithWorkflowName filters checkpoints by workflow name.
func WithWorkflowName(name string) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.WorkflowName = name
	}
}

// WithStatus filters checkpoints by status.
func WithStatus(status CheckpointStatus) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.Status = status
	}
}

// WithLimit sets the maximum number of checkpoints to return.
func WithLimit(limit int) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.Limit = limit
	}
}

// WithOffset sets the offset for pagination.
func WithOffset(offset int) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.Offset = offset
	}
}

// WithSortBy sets the field to sort by.
func WithSortBy(field string) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.SortBy = field
	}
}

// WithSortDesc sets the sort order to descending.
func WithSortDesc(desc bool) ListCheckpointOption {
	return func(opts *ListCheckpointOptions) {
		opts.SortDesc = desc
	}
}
