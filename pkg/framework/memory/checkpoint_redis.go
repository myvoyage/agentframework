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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCheckpointStore implements CheckpointStore using Redis.
type RedisCheckpointStore struct {
	client *redis.Client
	prefix string
}

// NewRedisCheckpointStore creates a new RedisCheckpointStore.
func NewRedisCheckpointStore(client *redis.Client, prefix string) *RedisCheckpointStore {
	if prefix == "" {
		prefix = "checkpoint:"
	}
	return &RedisCheckpointStore{
		client: client,
		prefix: prefix,
	}
}

func (s *RedisCheckpointStore) key(runID string) string {
	return s.prefix + runID
}

func (s *RedisCheckpointStore) Save(ctx context.Context, cp *Checkpoint) error {
	cp.UpdatedAt = time.Now()
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = cp.UpdatedAt
	}

	data, err := json.Marshal(cp)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.key(cp.RunID), data, 0).Err()
}

func (s *RedisCheckpointStore) Load(ctx context.Context, runID string) (*Checkpoint, error) {
	data, err := s.client.Get(ctx, s.key(runID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("checkpoint not found: %s", runID)
		}
		return nil, err
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

func (s *RedisCheckpointStore) List(ctx context.Context, options ...ListCheckpointOption) ([]*Checkpoint, error) {
	// Redis doesn't support complex queries easily
	// For simplicity, we'll load all checkpoints and filter in memory
	keys, err := s.client.Keys(ctx, s.prefix+"*").Result()
	if err != nil {
		return nil, err
	}

	var checkpoints []*Checkpoint
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var cp Checkpoint
		if err := json.Unmarshal(data, &cp); err != nil {
			continue
		}

		checkpoints = append(checkpoints, &cp)
	}

	// Apply options
	opts := &ListCheckpointOptions{}
	for _, option := range options {
		option(opts)
	}

	// Filter results
	var filtered []*Checkpoint
	for _, cp := range checkpoints {
		if opts.WorkflowName != "" && cp.WorkflowName != opts.WorkflowName {
			continue
		}
		if opts.Status != "" && cp.Status != opts.Status {
			continue
		}
		filtered = append(filtered, cp)
	}

	// Apply pagination
	if opts.Limit > 0 {
		if opts.Offset >= len(filtered) {
			return []*Checkpoint{}, nil
		}
		end := opts.Offset + opts.Limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[opts.Offset:end]
	}

	return filtered, nil
}

func (s *RedisCheckpointStore) Delete(ctx context.Context, runID string) error {
	return s.client.Del(ctx, s.key(runID)).Err()
}

func (s *RedisCheckpointStore) HealthCheck(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisCheckpointStore) Close(ctx context.Context) error {
	_ = ctx
	return s.client.Close()
}
