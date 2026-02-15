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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ThreadStore interface {
	Create(ctx context.Context) (*Thread, error)
	Get(ctx context.Context, id string) (*Thread, error)
	Save(ctx context.Context, thread *Thread) error
	HealthCheck(ctx context.Context) error
	Close(ctx context.Context) error
}

// ThreadOptions defines options for thread behavior
type ThreadOptions struct {
	// MaxMessages defines the maximum number of messages to keep in the thread
	MaxMessages int
	// MaxMessageSize defines the maximum size of a single message in bytes
	MaxMessageSize int
	// TTL defines the time-to-live for a thread
	TTL time.Duration
}

// ThreadMetadata stores metadata for threads
type ThreadMetadata struct {
	Thread    *Thread
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemoryThreadStore struct {
	mu      sync.RWMutex
	threads map[string]*ThreadMetadata
	// Options for thread behavior
	Options ThreadOptions
	// Stop channel for the cleanup goroutine
	stopCleanup chan struct{}
}

// NewMemoryThreadStore creates a new MemoryThreadStore
func NewMemoryThreadStore() *MemoryThreadStore {
	store := &MemoryThreadStore{
		threads: make(map[string]*ThreadMetadata),
		Options: ThreadOptions{
			MaxMessages:    100,
			MaxMessageSize: 1024 * 1024, // 1MB
			TTL:            24 * time.Hour,
		},
		stopCleanup: make(chan struct{}),
	}

	// Start the cleanup goroutine
	go store.cleanupLoop()

	return store
}

// NewMemoryThreadStoreWithOptions creates a new MemoryThreadStore with custom options
func NewMemoryThreadStoreWithOptions(opts ThreadOptions) *MemoryThreadStore {
	store := &MemoryThreadStore{
		threads:     make(map[string]*ThreadMetadata),
		Options:     opts,
		stopCleanup: make(chan struct{}),
	}

	// Start the cleanup goroutine
	go store.cleanupLoop()

	return store
}

// cleanupLoop periodically cleans up expired threads
func (s *MemoryThreadStore) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredThreads()
		case <-s.stopCleanup:
			return
		}
	}
}

// cleanupExpiredThreads removes threads that have expired based on TTL
func (s *MemoryThreadStore) cleanupExpiredThreads() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, meta := range s.threads {
		if now.Sub(meta.UpdatedAt) > s.Options.TTL {
			delete(s.threads, id)
		}
	}
}

// Create creates a new thread
func (s *MemoryThreadStore) Create(ctx context.Context) (*Thread, error) {
	_ = ctx

	id := uuid.NewString()
	now := time.Now()

	t := &Thread{
		ID:       id,
		Messages: []*schema.Message{},
	}

	meta := &ThreadMetadata{
		Thread:    t,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.threads[id] = meta
	s.mu.Unlock()

	return t, nil
}

// Get retrieves a thread by ID, updating its last accessed time
func (s *MemoryThreadStore) Get(ctx context.Context, id string) (*Thread, error) {
	_ = ctx

	s.mu.Lock()
	meta, ok := s.threads[id]
	if ok {
		// Update the last accessed time
		meta.UpdatedAt = time.Now()
		s.threads[id] = meta
	}
	s.mu.Unlock()

	if !ok {
		return nil, nil
	}
	return meta.Thread, nil
}

// Save saves a thread, applying thread options
func (s *MemoryThreadStore) Save(ctx context.Context, thread *Thread) error {
	_ = ctx

	if thread == nil || thread.ID == "" {
		return nil
	}

	// Apply thread options
	s.applyThreadOptions(thread)

	now := time.Now()

	s.mu.Lock()
	meta, ok := s.threads[thread.ID]
	if !ok {
		// Create new metadata if it doesn't exist
		meta = &ThreadMetadata{
			Thread:    thread,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		// Update existing metadata
		meta.Thread = thread
		meta.UpdatedAt = now
	}
	s.threads[thread.ID] = meta
	s.mu.Unlock()

	return nil
}

// applyThreadOptions applies thread options to the given thread
func (s *MemoryThreadStore) applyThreadOptions(thread *Thread) {
	// Limit message history
	if s.Options.MaxMessages > 0 && len(thread.Messages) > s.Options.MaxMessages {
		// Keep only the most recent messages
		thread.Messages = thread.Messages[len(thread.Messages)-s.Options.MaxMessages:]
	}

	// Limit message size
	if s.Options.MaxMessageSize > 0 {
		for i, msg := range thread.Messages {
			if len(msg.Content) > s.Options.MaxMessageSize {
				// Truncate message content
				thread.Messages[i].Content = msg.Content[:s.Options.MaxMessageSize]
			}
		}
	}
}

func (s *MemoryThreadStore) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

func (s *MemoryThreadStore) Close(ctx context.Context) error {
	_ = ctx

	// Stop the cleanup goroutine
	close(s.stopCleanup)

	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.threads)
	return nil
}

type FileThreadStore struct {
	dir string
	mu  sync.RWMutex
}

func NewFileThreadStore(dir string) (*FileThreadStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileThreadStore{dir: dir}, nil
}

func (s *FileThreadStore) Create(ctx context.Context) (*Thread, error) {
	_ = ctx

	id := uuid.NewString()
	t := &Thread{
		ID:       id,
		Messages: []*schema.Message{},
	}

	if err := s.Save(context.Background(), t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *FileThreadStore) Get(ctx context.Context, id string) (*Thread, error) {
	_ = ctx

	path := filepath.Join(s.dir, id+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var t Thread
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *FileThreadStore) Save(ctx context.Context, thread *Thread) error {
	_ = ctx

	if thread == nil || thread.ID == "" {
		return nil
	}

	data, err := json.Marshal(thread)
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, thread.ID+".json")

	s.mu.Lock()
	defer s.mu.Unlock()

	return os.WriteFile(path, data, 0o644)
}

func (s *FileThreadStore) HealthCheck(ctx context.Context) error {
	_ = ctx
	// Check if directory is writable
	tmpFile := filepath.Join(s.dir, ".healthcheck")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a temporary file
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	f.Close()

	// Delete the temporary file
	return os.Remove(tmpFile)
}

func (s *FileThreadStore) Close(ctx context.Context) error {
	_ = ctx
	// No resources to close for FileThreadStore
	return nil
}

type RedisThreadStore struct {
	client *redis.Client
	prefix string
}

func NewRedisThreadStore(client *redis.Client, prefix string) *RedisThreadStore {
	if prefix == "" {
		prefix = "thread:"
	}
	return &RedisThreadStore{
		client: client,
		prefix: prefix,
	}
}

func (s *RedisThreadStore) key(id string) string {
	return s.prefix + id
}

func (s *RedisThreadStore) Create(ctx context.Context) (*Thread, error) {
	id := uuid.NewString()
	t := &Thread{
		ID:       id,
		Messages: []*schema.Message{},
	}
	if err := s.Save(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *RedisThreadStore) Get(ctx context.Context, id string) (*Thread, error) {
	data, err := s.client.Get(ctx, s.key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var t Thread
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *RedisThreadStore) Save(ctx context.Context, thread *Thread) error {
	if thread == nil || thread.ID == "" {
		return nil
	}

	data, err := json.Marshal(thread)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.key(thread.ID), data, 0).Err()
}

func (s *RedisThreadStore) HealthCheck(ctx context.Context) error {
	// Ping Redis to check connection
	return s.client.Ping(ctx).Err()
}

func (s *RedisThreadStore) Close(ctx context.Context) error {
	_ = ctx
	return s.client.Close()
}

type SQLThreadStore struct {
	db    *sql.DB
	table string
}

func NewSQLThreadStore(db *sql.DB, table string) *SQLThreadStore {
	if table == "" {
		table = "threads"
	}

	store := &SQLThreadStore{
		db:    db,
		table: table,
	}

	return store
}

// Init creates the table if it doesn't exist
func (s *SQLThreadStore) Init(ctx context.Context) error {
	// Create table if it doesn't exist
	// This supports SQLite, PostgreSQL, MySQL, and other databases that use similar syntax
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR(255) PRIMARY KEY,
		data JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, s.table)

	// Use a generic TEXT type for all databases for compatibility
	query = fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR(255) PRIMARY KEY,
		data TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, s.table)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *SQLThreadStore) Create(ctx context.Context) (*Thread, error) {
	id := uuid.NewString()
	t := &Thread{
		ID:       id,
		Messages: []*schema.Message{},
	}
	if err := s.Save(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *SQLThreadStore) Get(ctx context.Context, id string) (*Thread, error) {
	row := s.db.QueryRowContext(ctx, "SELECT data FROM "+s.table+" WHERE id = ?", id)

	var data []byte
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var t Thread
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *SQLThreadStore) Save(ctx context.Context, thread *Thread) error {
	if thread == nil || thread.ID == "" {
		return nil
	}

	data, err := json.Marshal(thread)
	if err != nil {
		return err
	}

	// Try to update first
	updateQuery := fmt.Sprintf(`
	UPDATE %s SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, s.table)

	result, err := s.db.ExecContext(ctx, updateQuery, data, thread.ID)
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
		INSERT INTO %s (id, data, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)`, s.table)

		_, err = s.db.ExecContext(ctx, insertQuery, thread.ID, data)
	}
	return err
}

func (s *SQLThreadStore) HealthCheck(ctx context.Context) error {
	// Ping the database to check connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

func (s *SQLThreadStore) Close(ctx context.Context) error {
	_ = ctx
	return s.db.Close()
}
