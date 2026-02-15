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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SQLCheckpointStore implements CheckpointStore using a SQL database.
type SQLCheckpointStore struct {
	db    *sql.DB
	table string
}

// NewSQLCheckpointStore creates a new SQLCheckpointStore.
func NewSQLCheckpointStore(db *sql.DB, table string) *SQLCheckpointStore {
	if table == "" {
		table = "checkpoints"
	}
	return &SQLCheckpointStore{
		db:    db,
		table: table,
	}
}

// Init creates the table if it doesn't exist.
func (s *SQLCheckpointStore) Init(ctx context.Context) error {
	// Create table if it doesn't exist
	// Use TEXT type for JSON columns to support all databases
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		run_id VARCHAR(255) PRIMARY KEY,
		workflow_name VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		state TEXT,
		input TEXT,
		output TEXT,
		progress FLOAT DEFAULT 0.0,
		error TEXT,
		metadata TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, s.table)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *SQLCheckpointStore) Save(ctx context.Context, cp *Checkpoint) error {
	cp.UpdatedAt = time.Now()
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = cp.UpdatedAt
	}

	stateData, err := json.Marshal(cp.State)
	if err != nil {
		return err
	}

	metadataData, err := json.Marshal(cp.Metadata)
	if err != nil {
		return err
	}

	// Try to update first
	updateQuery := fmt.Sprintf(`
	UPDATE %s SET
		workflow_name = ?, status = ?, state = ?, input = ?, output = ?, 
		progress = ?, error = ?, metadata = ?, updated_at = ?
	WHERE run_id = ?`, s.table)

	result, err := s.db.ExecContext(ctx, updateQuery,
		cp.WorkflowName, string(cp.Status), stateData,
		cp.Input, cp.Output, cp.Progress, cp.Error,
		metadataData, cp.UpdatedAt, cp.RunID)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were updated, insert a new record
	if rowsAffected == 0 {
		insertQuery := fmt.Sprintf(`
		INSERT INTO %s (
			run_id, workflow_name, status, state, input, output, progress, error, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, s.table)

		_, err = s.db.ExecContext(ctx, insertQuery,
			cp.RunID, cp.WorkflowName, string(cp.Status), stateData,
			cp.Input, cp.Output, cp.Progress, cp.Error,
			metadataData, cp.CreatedAt, cp.UpdatedAt)
	}
	return err
}

func (s *SQLCheckpointStore) Load(ctx context.Context, runID string) (*Checkpoint, error) {
	query := fmt.Sprintf(`
	SELECT workflow_name, status, state, input, output, progress, error, metadata, created_at, updated_at 
	FROM %s WHERE run_id = ?`, s.table)

	row := s.db.QueryRowContext(ctx, query, runID)

	var (
		cp           Checkpoint
		stateData    []byte
		metadataData []byte
	)

	err := row.Scan(
		&cp.WorkflowName, &cp.Status, &stateData,
		&cp.Input, &cp.Output, &cp.Progress, &cp.Error,
		&metadataData, &cp.CreatedAt, &cp.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("checkpoint not found: %s", runID)
		}
		return nil, err
	}

	cp.RunID = runID

	if len(stateData) > 0 {
		var state WorkflowState
		if err := json.Unmarshal(stateData, &state); err == nil {
			cp.State = &state
		}
	}

	if len(metadataData) > 0 {
		var metadata map[string]string
		if err := json.Unmarshal(metadataData, &metadata); err == nil {
			cp.Metadata = metadata
		}
	}

	return &cp, nil
}

func (s *SQLCheckpointStore) List(ctx context.Context, options ...ListCheckpointOption) ([]*Checkpoint, error) {
	opts := &ListCheckpointOptions{}
	for _, option := range options {
		option(opts)
	}

	// Build query
	baseQuery := fmt.Sprintf(`
	SELECT run_id, workflow_name, status, state, input, output, progress, error, metadata, created_at, updated_at 
	FROM %s`, s.table)

	var filters []string
	var args []any

	if opts.WorkflowName != "" {
		filters = append(filters, "workflow_name = ?")
		args = append(args, opts.WorkflowName)
	}

	if opts.Status != "" {
		filters = append(filters, "status = ?")
		args = append(args, opts.Status)
	}

	if len(filters) > 0 {
		baseQuery += " WHERE " + strings.Join(filters, " AND ")
	}

	// Add sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		// Validate sortBy to prevent SQL injection
		allowedSortBy := map[string]bool{
			"run_id":        true,
			"workflow_name": true,
			"status":        true,
			"created_at":    true,
			"updated_at":    true,
		}
		if allowedSortBy[opts.SortBy] {
			sortBy = opts.SortBy
		}
	}

	sortOrder := "ASC"
	if opts.SortDesc {
		sortOrder = "DESC"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	if opts.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, opts.Limit)
		if opts.Offset > 0 {
			baseQuery += " OFFSET ?"
			args = append(args, opts.Offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []*Checkpoint
	for rows.Next() {
		var (
			cp           Checkpoint
			stateData    []byte
			metadataData []byte
		)

		err := rows.Scan(
			&cp.RunID, &cp.WorkflowName, &cp.Status, &stateData,
			&cp.Input, &cp.Output, &cp.Progress, &cp.Error,
			&metadataData, &cp.CreatedAt, &cp.UpdatedAt)
		if err != nil {
			continue
		}

		if len(stateData) > 0 {
			var state WorkflowState
			if err := json.Unmarshal(stateData, &state); err == nil {
				cp.State = &state
			}
		}

		if len(metadataData) > 0 {
			var metadata map[string]string
			if err := json.Unmarshal(metadataData, &metadata); err == nil {
				cp.Metadata = metadata
			}
		}

		checkpoints = append(checkpoints, &cp)
	}

	return checkpoints, nil
}

func (s *SQLCheckpointStore) Delete(ctx context.Context, runID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE run_id = ?", s.table)
	_, err := s.db.ExecContext(ctx, query, runID)
	return err
}

func (s *SQLCheckpointStore) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

func (s *SQLCheckpointStore) Close(ctx context.Context) error {
	_ = ctx
	return s.db.Close()
}
