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
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// initThreadStore initializes the thread store based on configuration
func (h *Host) initThreadStore(ctx context.Context) error {
	spec := h.cfg.ThreadStore
	switch spec.Type {
	case "", "memory":
		h.threadStore = NewMemoryThreadStore()
	case "file":
		dir := spec.Dir
		if dir == "" {
			dir = "./threads"
		}
		fs, err := NewFileThreadStore(dir)
		if err != nil {
			return err
		}
		h.threadStore = fs
	case "redis":
		addr := spec.RedisAddr
		if addr == "" {
			return errors.New("redis thread store requires redisAddr")
		}
		client := newRedisClient(addr)
		h.threadStore = NewRedisThreadStore(client, spec.RedisPrefix)
	case "sql":
		if spec.DriverName == "" || spec.DSN == "" {
			return errors.New("sql thread store requires driver and dsn")
		}
		db, err := sql.Open(spec.DriverName, spec.DSN)
		if err != nil {
			return err
		}
		h.threadStore = NewSQLThreadStore(db, spec.Table)
	default:
		return fmt.Errorf("unknown thread store type %q", spec.Type)
	}
	return nil
}

// newRedisClient creates a new Redis client
func newRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
