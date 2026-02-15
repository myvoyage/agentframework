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

package store

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/context"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore SQLite 上下文存储实现
type SQLiteStore struct {
	db        *sql.DB
	dbPath    string
	mu        sync.RWMutex
	started   bool
	ctx       stdcontext.Context // 使用标准库的 context
	cancel    stdcontext.CancelFunc
}

// SQLiteStoreConfig SQLite 存储配置
type SQLiteStoreConfig struct {
	DBPath string // 数据库文件路径
}

// NewSQLiteStore 创建新的 SQLite 存储
func NewSQLiteStore(config *SQLiteStoreConfig) (*SQLiteStore, error) {
	if config == nil {
		config = &SQLiteStoreConfig{}
	}

	// 默认数据库路径
	if config.DBPath == "" {
		config.DBPath = "./contexts.db"
	}

	// 确保目录存在
	dir := filepath.Dir(config.DBPath)
	if dir != "" && dir != "." {
		// 这里可以创建目录，简化处理
	}

	return &SQLiteStore{
		dbPath:  config.DBPath,
		started: false,
	}, nil
}

// Start 启动 SQLite 存储
func (s *SQLiteStore) Start(ctx stdcontext.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("store already started")
	}

	// 创建上下文
	s.ctx, s.cancel = stdcontext.WithCancel(ctx)

	// 打开数据库连接
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 启用 WAL 模式以提高并发性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return fmt.Errorf("enable WAL mode: %w", err)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		db.Close()
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	s.db = db
	s.started = true

	// 初始化数据库 Schema
	if err := s.initSchema(); err != nil {
		s.stop()
		return fmt.Errorf("init schema: %w", err)
	}

	return nil
}

// Stop 停止 SQLite 存储
func (s *SQLiteStore) Stop(ctx stdcontext.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stop()
}

// stop 内部停止方法（不加锁）
func (s *SQLiteStore) stop() error {
	if !s.started {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}
	}

	s.started = false
	return nil
}

// initSchema 初始化数据库 Schema
func (s *SQLiteStore) initSchema() error {
	// 创建表
	for _, tableSQL := range Schema {
		if _, err := s.db.Exec(tableSQL); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	// 创建索引
	for _, indexSQL := range Indexes {
		if _, err := s.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	return nil
}

// ===== 基础 CRUD 操作 =====

// CreateContext 创建新的上下文
func (s *SQLiteStore) CreateContext(ctx stdcontext.Context, ctxt *context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return "", fmt.Errorf("store not started")
	}

	// 生成 ID（如果还没有）
	if ctxt.ID == "" {
		ctxt.ID = ctxt.GenerateID()
	}

	now := time.Now().Unix()

	// 插入上下文
	query := `INSERT INTO contexts (id, type, title, workspace, uri, parent_id, version, metadata, created_at, updated_at, accessed_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	metadataJSON, _ := json.Marshal(ctxt.Metadata)

	_, err := s.db.ExecContext(ctx, query,
		ctxt.ID,
		string(ctxt.Type),
		ctxt.Title,
		ctxt.Workspace,
		ctxt.URI,
		ctxt.ParentID,
		ctxt.Version,
		string(metadataJSON),
		ctxt.CreatedAt.Unix(),
		now,
		now,
	)
	if err != nil {
		return "", fmt.Errorf("insert context: %w", err)
	}

	// 插入层级数据
	if err := s.insertLayers(ctx, ctxt); err != nil {
		// 回滚
		s.db.ExecContext(ctx, "DELETE FROM contexts WHERE id = ?", ctxt.ID)
		return "", fmt.Errorf("insert layers: %w", err)
	}

	// 插入索引
	if err := s.updateContextIndex(ctx, ctxt); err != nil {
		// 非致命错误，只记录
		fmt.Printf("Warning: failed to update context index: %v\n", err)
	}

	return ctxt.ID, nil
}

// insertLayers 插入层级数据
func (s *SQLiteStore) insertLayers(ctx stdcontext.Context, ctxt *context.Context) error {
	now := time.Now().Unix()

	// L0 层
	if ctxt.Layers.L0 != nil {
		query := `INSERT INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		metadataJSON, _ := json.Marshal(map[string]interface{}{
			"method":      ctxt.Layers.L0.Method,
			"generated_at": ctxt.Layers.L0.GeneratedAt.Unix(),
		})
		_, err := s.db.ExecContext(ctx, query,
			ctxt.ID, "l0",
			ctxt.Layers.L0.Content,
			ctxt.Layers.L0.Tokens,
			string(metadataJSON),
			ctxt.Version,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("insert L0 layer: %w", err)
		}
	}

	// L1 层
	if ctxt.Layers.L1 != nil {
		query := `INSERT INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		metadataJSON, _ := json.Marshal(map[string]interface{}{
			"method":      ctxt.Layers.L1.Method,
			"generated_at": ctxt.Layers.L1.GeneratedAt.Unix(),
		})
		_, err := s.db.ExecContext(ctx, query,
			ctxt.ID, "l1",
			ctxt.Layers.L1.Content,
			ctxt.Layers.L1.Tokens,
			string(metadataJSON),
			ctxt.Version,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("insert L1 layer: %w", err)
		}
	}

	// L2 层
	if ctxt.Layers.L2 != nil {
		query := `INSERT INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		metadataJSON, _ := json.Marshal(map[string]interface{}{
			"format":      ctxt.Layers.L2.Format,
			"source":      ctxt.Layers.L2.Source,
			"generated_at": ctxt.Layers.L2.GeneratedAt.Unix(),
		})
		_, err := s.db.ExecContext(ctx, query,
			ctxt.ID, "l2",
			ctxt.Layers.L2.Content,
			ctxt.Layers.L2.Tokens,
			string(metadataJSON),
			ctxt.Version,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("insert L2 layer: %w", err)
		}
	}

	return nil
}

// updateContextIndex 更新上下文索引
func (s *SQLiteStore) updateContextIndex(ctx stdcontext.Context, ctxt *context.Context) error {
	l0Tokens := 0
	if ctxt.Layers.L0 != nil {
		l0Tokens = ctxt.Layers.L0.Tokens
	}

	l1Tokens := 0
	if ctxt.Layers.L1 != nil {
		l1Tokens = ctxt.Layers.L1.Tokens
	}

	l2Tokens := 0
	if ctxt.Layers.L2 != nil {
		l2Tokens = ctxt.Layers.L2.Tokens
	}

	memoryCount := 0
	if ctxt.Memories != nil {
		memoryCount = ctxt.Memories.GetMemoryCount()
	}

	query := `INSERT OR REPLACE INTO context_index
			  (context_id, type, workspace, title, l0_tokens, l1_tokens, l2_tokens, total_tokens, memory_count, access_count)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`

	_, err := s.db.ExecContext(ctx, query,
		ctxt.ID,
		string(ctxt.Type),
		ctxt.Workspace,
		ctxt.Title,
		l0Tokens,
		l1Tokens,
		l2Tokens,
		l0Tokens+l1Tokens+l2Tokens,
		memoryCount,
	)

	return err
}

// GetContext 根据 ID 获取上下文
func (s *SQLiteStore) GetContext(ctx stdcontext.Context, contextID string) (*context.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	query := `SELECT id, type, title, workspace, uri, parent_id, version, metadata, created_at, updated_at, accessed_at
			  FROM contexts WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, contextID)

	var ctxt context.Context
	var metadataJSON []byte
	var created, updated, accessed int64

	err := row.Scan(
		&ctxt.ID,
		&ctxt.Type,
		&ctxt.Title,
		&ctxt.Workspace,
		&ctxt.URI,
		&ctxt.ParentID,
		&ctxt.Version,
		&metadataJSON,
		&created,
		&updated,
		&accessed,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("context not found: %s", contextID)
	}
	if err != nil {
		return nil, fmt.Errorf("scan context: %w", err)
	}

	// 解析元数据
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &ctxt.Metadata)
	} else {
		ctxt.Metadata = make(map[string]string)
	}

	// 设置时间戳
	ctxt.CreatedAt = time.Unix(created, 0)
	ctxt.UpdatedAt = time.Unix(updated, 0)
	ctxt.AccessedAt = time.Unix(accessed, 0)

	// 加载层级
	if err := s.loadLayers(ctx, &ctxt); err != nil {
		return nil, fmt.Errorf("load layers: %w", err)
	}

	// 加载记忆
	if err := s.loadMemories(ctx, &ctxt); err != nil {
		return nil, fmt.Errorf("load memories: %w", err)
	}

	// 更新访问时间
	go s.updateAccessTime(contextID)

	return &ctxt, nil
}

// loadLayers 加载层级数据
func (s *SQLiteStore) loadLayers(ctx stdcontext.Context, ctxt *context.Context) error {
	query := `SELECT layer_type, content, tokens, metadata FROM context_layers WHERE context_id = ?`

	rows, err := s.db.QueryContext(ctx, query, ctxt.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var layerType, content string
		var tokens int
		var metadataJSON []byte

		if err := rows.Scan(&layerType, &content, &tokens, &metadataJSON); err != nil {
			return err
		}

		var metadata map[string]interface{}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &metadata)
		}

		switch layerType {
		case "l0":
			ctxt.Layers.L0 = &context.LayerSummary{
				Content:     content,
				Tokens:      tokens,
				GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
				Method:      metadata["method"].(string),
			}
		case "l1":
			ctxt.Layers.L1 = &context.LayerOverview{
				Content:     content,
				Tokens:      tokens,
				GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
				Method:      metadata["method"].(string),
			}
		case "l2":
			ctxt.Layers.L2 = &context.LayerDetails{
				Content:     content,
				Tokens:      tokens,
				Format:      metadata["format"].(string),
				Source:      metadata["source"].(string),
				GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
			}
		}
	}

	return rows.Err()
}

// loadMemories 加载记忆数据
func (s *SQLiteStore) loadMemories(ctx stdcontext.Context, ctxt *context.Context) error {
	query := `SELECT memory_type, id, content FROM memories WHERE context_id = ?`

	rows, err := s.db.QueryContext(ctx, query, ctxt.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	ctxt.Memories = &context.MemoryCollection{}

	for rows.Next() {
		var memoryType, memoryID, contentJSON string

		if err := rows.Scan(&memoryType, &memoryID, &contentJSON); err != nil {
			return err
		}

		// 根据类型解析记忆
		switch context.MemoryType(memoryType) {
		case context.MemoryTypeProfile:
			var mem context.ProfileMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Profiles = append(ctxt.Memories.Profiles, &mem)
			}
		case context.MemoryTypePreference:
			var mem context.PreferenceMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Preferences = append(ctxt.Memories.Preferences, &mem)
			}
		case context.MemoryTypeEntity:
			var mem context.EntityMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Entities = append(ctxt.Memories.Entities, &mem)
			}
		case context.MemoryTypeEvent:
			var mem context.EventMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Events = append(ctxt.Memories.Events, &mem)
			}
		case context.MemoryTypeCase:
			var mem context.CaseMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Cases = append(ctxt.Memories.Cases, &mem)
			}
		case context.MemoryTypePattern:
			var mem context.PatternMemory
			if err := json.Unmarshal([]byte(contentJSON), &mem); err == nil {
				ctxt.Memories.Patterns = append(ctxt.Memories.Patterns, &mem)
			}
		}
	}

	return rows.Err()
}

// UpdateContext 更新上下文
func (s *SQLiteStore) UpdateContext(ctx stdcontext.Context, contextID string, updates context.ContextUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	now := time.Now().Unix()

	// 构建更新查询
	query := "UPDATE contexts SET updated_at = ?"
	args := []interface{}{now}

	if updates.Title != nil {
		query += ", title = ?"
		args = append(args, *updates.Title)
	}
	if updates.Workspace != nil {
		query += ", workspace = ?"
		args = append(args, *updates.Workspace)
	}
	if updates.URI != nil {
		query += ", uri = ?"
		args = append(args, *updates.URI)
	}
	if updates.Metadata != nil {
		metadataJSON, _ := json.Marshal(*updates.Metadata)
		query += ", metadata = ?"
		args = append(args, string(metadataJSON))
	}

	query += " WHERE id = ?"
	args = append(args, contextID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update context: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("context not found: %s", contextID)
	}

	// 更新层级（如果提供）
	if updates.Layers != nil {
		if err := s.updateLayers(ctx, contextID, updates.Layers); err != nil {
			return fmt.Errorf("update layers: %w", err)
		}
	}

	// 更新记忆（如果提供）
	if updates.Memories != nil {
		if err := s.updateMemories(ctx, contextID, updates.Memories); err != nil {
			return fmt.Errorf("update memories: %w", err)
		}
	}

	// 重新获取完整上下文以更新索引
	ctxt, err := s.GetContext(ctx, contextID)
	if err == nil {
		s.updateContextIndex(ctx, ctxt)
	}

	return nil
}

// updateLayers 更新层级
func (s *SQLiteStore) updateLayers(ctx stdcontext.Context, contextID string, updates *context.ContextLayersUpdate) error {
	now := time.Now().Unix()

	if updates.L0 != nil {
		query := `INSERT OR REPLACE INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

		metadata := map[string]interface{}{"method": "manual"}
		if updates.L0.Method != nil {
			metadata["method"] = *updates.L0.Method
		}
		metadataJSON, _ := json.Marshal(metadata)

		content := ""
		tokens := 0
		if updates.L0.Content != nil {
			content = *updates.L0.Content
		}
		if updates.L0.Tokens != nil {
			tokens = *updates.L0.Tokens
		}

		_, err := s.db.ExecContext(ctx, query, contextID, "l0", content, tokens, string(metadataJSON), now, now)
		if err != nil {
			return err
		}
	}

	if updates.L1 != nil {
		query := `INSERT OR REPLACE INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

		metadata := map[string]interface{}{"method": "manual"}
		if updates.L1.Method != nil {
			metadata["method"] = *updates.L1.Method
		}
		metadataJSON, _ := json.Marshal(metadata)

		content := ""
		tokens := 0
		if updates.L1.Content != nil {
			content = *updates.L1.Content
		}
		if updates.L1.Tokens != nil {
			tokens = *updates.L1.Tokens
		}

		_, err := s.db.ExecContext(ctx, query, contextID, "l1", content, tokens, string(metadataJSON), now, now)
		if err != nil {
			return err
		}
	}

	if updates.L2 != nil {
		query := `INSERT OR REPLACE INTO context_layers (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

		metadata := map[string]interface{}{"source": "manual"}
		if updates.L2.Format != nil {
			metadata["format"] = *updates.L2.Format
		}
		if updates.L2.Source != nil {
			metadata["source"] = *updates.L2.Source
		}
		metadataJSON, _ := json.Marshal(metadata)

		content := ""
		tokens := 0
		if updates.L2.Content != nil {
			content = *updates.L2.Content
		}
		if updates.L2.Tokens != nil {
			tokens = *updates.L2.Tokens
		}

		_, err := s.db.ExecContext(ctx, query, contextID, "l2", content, tokens, string(metadataJSON), now, now)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateMemories 更新记忆
func (s *SQLiteStore) updateMemories(ctx stdcontext.Context, contextID string, memories *context.MemoryCollection) error {
	// 先删除现有记忆
	if _, err := s.db.ExecContext(ctx, "DELETE FROM memories WHERE context_id = ?", contextID); err != nil {
		return err
	}

	now := time.Now().Unix()

	// 插入新记忆
	if memories.Profiles != nil {
		for _, mem := range memories.Profiles {
			contentJSON, _ := json.Marshal(mem)
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "profile", mem.Name, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	if memories.Preferences != nil {
		for _, mem := range memories.Preferences {
			contentJSON, _ := json.Marshal(mem)
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "preference", mem.Key, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	if memories.Entities != nil {
		for _, mem := range memories.Entities {
			contentJSON, _ := json.Marshal(mem)
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "entity", mem.Name, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	if memories.Events != nil {
		for _, mem := range memories.Events {
			contentJSON, _ := json.Marshal(mem)
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "event", mem.Title, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	if memories.Cases != nil {
		for _, mem := range memories.Cases {
			contentJSON, _ := json.Marshal(mem)
			title := mem.Problem[:50]
			if len(title) < len(mem.Problem) {
				title += "..."
			}
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "case", title, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	if memories.Patterns != nil {
		for _, mem := range memories.Patterns {
			contentJSON, _ := json.Marshal(mem)
			title := mem.Pattern[:50]
			if len(title) < len(mem.Pattern) {
				title += "..."
			}
			_, err := s.db.ExecContext(ctx,
				"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mem.ID, contextID, "pattern", title, string(contentJSON), now, now)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteContext 删除上下文
func (s *SQLiteStore) DeleteContext(ctx stdcontext.Context, contextID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	// 删除上下文（级联删除会自动删除相关的层级、记忆等）
	result, err := s.db.ExecContext(ctx, "DELETE FROM contexts WHERE id = ?", contextID)
	if err != nil {
		return fmt.Errorf("delete context: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("context not found: %s", contextID)
	}

	// 删除索引
	s.db.ExecContext(ctx, "DELETE FROM context_index WHERE context_id = ?", contextID)

	return nil
}

// ===== 三层上下文操作 =====

// GetLayer 获取指定层级的内容
func (s *SQLiteStore) GetLayer(ctx stdcontext.Context, contextID string, layer context.LayerType) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	layerType := string(layer)
	if layer == context.LayerAuto {
		// 自动选择：优先 L1，然后 L0，最后 L2
		layerType = s.selectBestLayer(ctx, contextID)
	}

	query := `SELECT content, tokens, metadata FROM context_layers WHERE context_id = ? AND layer_type = ?`

	row := s.db.QueryRowContext(ctx, query, contextID, layerType)

	var content string
	var tokens int
	var metadataJSON []byte

	err := row.Scan(&content, &tokens, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("layer %s not found", layerType)
	}
	if err != nil {
		return nil, fmt.Errorf("scan layer: %w", err)
	}

	var metadata map[string]interface{}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &metadata)
	}

	switch context.LayerType(layerType) {
	case context.LayerTypeL0:
		return &context.LayerSummary{
			Content:     content,
			Tokens:      tokens,
			GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
			Method:      metadata["method"].(string),
		}, nil
	case context.LayerTypeL1:
		return &context.LayerOverview{
			Content:     content,
			Tokens:      tokens,
			GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
			Method:      metadata["method"].(string),
		}, nil
	case context.LayerTypeL2:
		return &context.LayerDetails{
			Content:     content,
			Tokens:      tokens,
			Format:      metadata["format"].(string),
			Source:      metadata["source"].(string),
			GeneratedAt: time.Unix(int64(metadata["generated_at"].(float64)), 0),
		}, nil
	default:
		return content, nil
	}
}

// selectBestLayer 选择最佳层级
func (s *SQLiteStore) selectBestLayer(ctx stdcontext.Context, contextID string) string {
	// 优先级：L1 > L0 > L2
	layers := []string{"l1", "l0", "l2"}
	for _, layer := range layers {
		var count int
		err := s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM context_layers WHERE context_id = ? AND layer_type = ? AND content IS NOT NULL AND content != ''",
			contextID, layer).Scan(&count)
		if err == nil && count > 0 {
			return layer
		}
	}
	return "l0" // 默认返回 L0
}

// GenerateLayers 生成缺失的层级（占位实现，需要 LayerGenerator）
func (s *SQLiteStore) GenerateLayers(ctx stdcontext.Context, contextID string) error {
	// TODO: 实现层级生成逻辑
	// 这需要集成 LayerGenerator
	return fmt.Errorf("layer generation not yet implemented")
}

// RegenerateLayer 重新生成指定层级
func (s *SQLiteStore) RegenerateLayer(ctx stdcontext.Context, contextID string, layer context.LayerType) error {
	// TODO: 实现层级重新生成逻辑
	return fmt.Errorf("layer regeneration not yet implemented")
}

// SetLayer 设置指定层级的内容
func (s *SQLiteStore) SetLayer(ctx stdcontext.Context, contextID string, layer context.LayerType, content interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	now := time.Now().Unix()
	layerType := string(layer)

	var contentStr string
	var tokens int
	var metadataJSON []byte

	switch c := content.(type) {
	case *context.LayerSummary:
		contentStr = c.Content
		tokens = c.Tokens
		metadata, _ := json.Marshal(map[string]interface{}{
			"method":      c.Method,
			"generated_at": c.GeneratedAt.Unix(),
		})
		metadataJSON = metadata
	case *context.LayerOverview:
		contentStr = c.Content
		tokens = c.Tokens
		metadata, _ := json.Marshal(map[string]interface{}{
			"method":      c.Method,
			"generated_at": c.GeneratedAt.Unix(),
		})
		metadataJSON = metadata
	case *context.LayerDetails:
		contentStr = c.Content
		tokens = c.Tokens
		metadata, _ := json.Marshal(map[string]interface{}{
			"format":      c.Format,
			"source":      c.Source,
			"generated_at": c.GeneratedAt.Unix(),
		})
		metadataJSON = metadata
	case string:
		contentStr = c
		tokens = len(contentStr) / 4 // 简单估算
	}

	query := `INSERT OR REPLACE INTO context_layers
			  (context_id, layer_type, content, tokens, metadata, version, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, contextID, layerType, contentStr, tokens, string(metadataJSON), now, now)
	if err != nil {
		return fmt.Errorf("set layer: %w", err)
	}

	return nil
}

// ===== 任务关联操作 =====

// AssociateContext 将上下文与任务关联
func (s *SQLiteStore) AssociateContext(ctx stdcontext.Context, taskID, contextID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	now := time.Now().Unix()

	query := `INSERT OR IGNORE INTO task_contexts (task_id, context_id, association_type, associated_at)
			  VALUES (?, ?, 'manual', ?)`

	_, err := s.db.ExecContext(ctx, query, taskID, contextID, now)
	if err != nil {
		return fmt.Errorf("associate context: %w", err)
	}

	return nil
}

// DissociateContext 解除任务与上下文的关联
func (s *SQLiteStore) DissociateContext(ctx stdcontext.Context, taskID, contextID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	query := `DELETE FROM task_contexts WHERE task_id = ? AND context_id = ?`

	_, err := s.db.ExecContext(ctx, query, taskID, contextID)
	if err != nil {
		return fmt.Errorf("dissociate context: %w", err)
	}

	return nil
}

// GetTaskContexts 获取任务的所有上下文
func (s *SQLiteStore) GetTaskContexts(ctx stdcontext.Context, taskID string) ([]*context.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	query := `SELECT context_id FROM task_contexts WHERE task_id = ?`

	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("query task contexts: %w", err)
	}
	defer rows.Close()

	var contexts []*context.Context
	for rows.Next() {
		var contextID string
		if err := rows.Scan(&contextID); err != nil {
			return nil, err
		}

		ctxt, err := s.GetContext(ctx, contextID)
		if err != nil {
			// 跳过无法加载的上下文
			continue
		}

		contexts = append(contexts, ctxt)
	}

	return contexts, rows.Err()
}

// GetContextTasks 获取使用指定上下文的所有任务
func (s *SQLiteStore) GetContextTasks(ctx stdcontext.Context, contextID string) ([]*beads.Task, error) {
	// 这个方法需要访问 TaskTracker 的 SQLite 存储
	// 由于模块分离，这里返回空列表或需要外部注入 TaskTracker
	return []*beads.Task{}, nil
}

// ===== 联合查询 =====

// QueryTasksWithContext 联合查询任务和上下文
func (s *SQLiteStore) QueryTasksWithContext(ctx stdcontext.Context, query beads.Query, filter context.ContextFilter) ([]*context.TaskWithContext, error) {
	// 这是一个复杂的联合查询，需要与 TaskTracker 配合
	// 这里提供基础实现
	return []*context.TaskWithContext{}, nil
}

// QueryContextsByTasks 根据任务列表查询上下文
func (s *SQLiteStore) QueryContextsByTasks(ctx stdcontext.Context, taskIDs []string) (map[string][]*context.Context, error) {
	result := make(map[string][]*context.Context)

	for _, taskID := range taskIDs {
		contexts, err := s.GetTaskContexts(ctx, taskID)
		if err != nil {
			continue
		}
		result[taskID] = contexts
	}

	return result, nil
}

// ===== 批量操作 =====

// BatchCreate 批量创建上下文
func (s *SQLiteStore) BatchCreate(ctx stdcontext.Context, contexts []*context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	ids := make([]string, 0, len(contexts))

	// 使用事务
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, ctxt := range contexts {
		if ctxt.ID == "" {
			ctxt.ID = ctxt.GenerateID()
		}

		now := time.Now().Unix()
		metadataJSON, _ := json.Marshal(ctxt.Metadata)

		query := `INSERT INTO contexts (id, type, title, workspace, uri, parent_id, version, metadata, created_at, updated_at, accessed_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := tx.ExecContext(ctx, query,
			ctxt.ID, string(ctxt.Type), ctxt.Title, ctxt.Workspace, ctxt.URI,
			ctxt.ParentID, ctxt.Version, string(metadataJSON),
			ctxt.CreatedAt.Unix(), now, now,
		)
		if err != nil {
			return nil, fmt.Errorf("insert context %s: %w", ctxt.ID, err)
		}

		ids = append(ids, ctxt.ID)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return ids, nil
}

// BatchGet 批量获取上下文
func (s *SQLiteStore) BatchGet(ctx stdcontext.Context, contextIDs []string) (map[string]*context.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	result := make(map[string]*context.Context)

	for _, contextID := range contextIDs {
		ctxt, err := s.GetContext(ctx, contextID)
		if err == nil {
			result[contextID] = ctxt
		}
	}

	return result, nil
}

// BatchUpdate 批量更新上下文
func (s *SQLiteStore) BatchUpdate(ctx stdcontext.Context, updates map[string]context.ContextUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for contextID, update := range updates {
		if err := s.UpdateContext(ctx, contextID, update); err != nil {
			return fmt.Errorf("update context %s: %w", contextID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// BatchDelete 批量删除上下文
func (s *SQLiteStore) BatchDelete(ctx stdcontext.Context, contextIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, contextID := range contextIDs {
		if _, err := tx.ExecContext(ctx, "DELETE FROM contexts WHERE id = ?", contextID); err != nil {
			return fmt.Errorf("delete context %s: %w", contextID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ===== 统计与健康 =====

// GetStats 获取统计信息
func (s *SQLiteStore) GetStats(ctx stdcontext.Context) (*context.ContextStoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	stats := &context.ContextStoreStats{
		ByType: make(map[context.ContextType]int64),
	}

	// 总上下文数
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM contexts").Scan(&stats.TotalContexts)
	if err != nil {
		return nil, err
	}

	// 按类型统计
	rows, err := s.db.QueryContext(ctx, "SELECT type, COUNT(*) FROM contexts GROUP BY type")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ctxtType string
			var count int64
			if rows.Scan(&ctxtType, &count) == nil {
				stats.ByType[context.ContextType(ctxtType)] = count
			}
		}
	}

	// 存储大小（简化处理）
	stats.StorageSize = 0

	return stats, nil
}

// HealthCheck 健康检查
func (s *SQLiteStore) HealthCheck(ctx stdcontext.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	return s.db.PingContext(ctx)
}

// ===== 生命周期 =====

// Sync 同步存储（SQLite 不需要，占位实现）
func (s *SQLiteStore) Sync(ctx stdcontext.Context) error {
	return nil
}

// ===== VFS 操作（基础实现） =====

// ReadFile 从 VFS 读取文件
func (s *SQLiteStore) ReadFile(ctx stdcontext.Context, uri string) ([]byte, error) {
	// TODO: 实现 VFS 读取
	return nil, fmt.Errorf("VFS read not yet implemented")
}

// WriteFile 向 VFS 写入文件
func (s *SQLiteStore) WriteFile(ctx stdcontext.Context, uri string, data []byte) error {
	// TODO: 实现 VFS 写入
	return fmt.Errorf("VFS write not yet implemented")
}

// ListFiles 列出目录中的文件
func (s *SQLiteStore) ListFiles(ctx stdcontext.Context, uri string) ([]*context.VFSFileInfo, error) {
	// TODO: 实现 VFS 列表
	return nil, fmt.Errorf("VFS list not yet implemented")
}

// SearchFiles 搜索文件
func (s *SQLiteStore) SearchFiles(ctx stdcontext.Context, query string, opts ...context.SearchOption) ([]*context.VFSSearchResult, error) {
	// TODO: 实现 VFS 搜索
	return nil, fmt.Errorf("VFS search not yet implemented")
}

// DeleteFile 从 VFS 删除文件
func (s *SQLiteStore) DeleteFile(ctx stdcontext.Context, uri string) error {
	// TODO: 实现 VFS 删除
	return fmt.Errorf("VFS delete not yet implemented")
}

// ===== 记忆操作 =====

// ExtractMemories 从上下文中提取记忆
func (s *SQLiteStore) ExtractMemories(ctx stdcontext.Context, contextID string) (*context.MemoryCollection, error) {
	// TODO: 实现记忆提取（需要 MemoryExtractor）
	return nil, fmt.Errorf("memory extraction not yet implemented")
}

// GetMemories 获取上下文的记忆
func (s *SQLiteStore) GetMemories(ctx stdcontext.Context, contextID string, memoryTypes []context.MemoryType) (*context.MemoryCollection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	ctxt, err := s.GetContext(ctx, contextID)
	if err != nil {
		return nil, err
	}

	if ctxt.Memories == nil {
		return &context.MemoryCollection{}, nil
	}

	// 如果没有指定类型，返回全部
	if len(memoryTypes) == 0 {
		return ctxt.Memories, nil
	}

	// 过滤指定类型
	result := &context.MemoryCollection{}
	for _, memType := range memoryTypes {
		switch memType {
		case context.MemoryTypeProfile:
			result.Profiles = ctxt.Memories.Profiles
		case context.MemoryTypePreference:
			result.Preferences = ctxt.Memories.Preferences
		case context.MemoryTypeEntity:
			result.Entities = ctxt.Memories.Entities
		case context.MemoryTypeEvent:
			result.Events = ctxt.Memories.Events
		case context.MemoryTypeCase:
			result.Cases = ctxt.Memories.Cases
		case context.MemoryTypePattern:
			result.Patterns = ctxt.Memories.Patterns
		}
	}

	return result, nil
}

// UpdateMemories 更新上下文的记忆
func (s *SQLiteStore) UpdateMemories(ctx stdcontext.Context, contextID string, memories *context.MemoryCollection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	// 先删除现有记忆
	if _, err := s.db.ExecContext(ctx, "DELETE FROM memories WHERE context_id = ?", contextID); err != nil {
		return err
	}

	now := time.Now().Unix()

	// 插入新记忆
	for _, mem := range memories.Profiles {
		contentJSON, _ := json.Marshal(mem)
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO memories (id, context_id, memory_type, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			mem.ID, contextID, "profile", mem.Name, string(contentJSON), now, now)
		if err != nil {
			return err
		}
	}

	// 其他类型的记忆类似处理...

	return nil
}

// DeduplicateMemories 去重记忆
func (s *SQLiteStore) DeduplicateMemories(ctx stdcontext.Context, contextID string) (*context.MemoryCollection, error) {
	// TODO: 实现记忆去重（需要 MemoryDeduplicator）
	return nil, fmt.Errorf("memory deduplication not yet implemented")
}

// ===== 辅助方法 =====

// updateAccessTime 更新访问时间（异步）
func (s *SQLiteStore) updateAccessTime(contextID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	now := time.Now().Unix()
	s.db.Exec("UPDATE contexts SET accessed_at = ? WHERE id = ?", now, contextID)

	// 更新访问计数
	s.db.Exec("UPDATE context_index SET access_count = access_count + 1 WHERE context_id = ?", contextID)
}

// ===== 记忆分层操作 =====

// SetMemoryTier 设置记忆的分层
func (s *SQLiteStore) SetMemoryTier(ctx stdcontext.Context, memoryID string, tier context.MemoryTier, expiresAt time.Time, importance float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	now := time.Now()

	// 使用 UPSERT 语法
	query := `
		INSERT INTO memory_tiers (memory_id, tier, created_at, expires_at, access_count, last_accessed, importance_score)
		VALUES (?, ?, ?, ?, 0, ?, ?)
		ON CONFLICT(memory_id) DO UPDATE SET
			tier = excluded.tier,
			expires_at = excluded.expires_at,
			importance_score = excluded.importance_score,
			last_accessed = excluded.last_accessed
	`

	_, err := s.db.ExecContext(ctx, query, memoryID, string(tier), now.Unix(), now.Unix(), now.Unix(), importance)
	return err
}

// GetMemoryTier 获取记忆的分层
func (s *SQLiteStore) GetMemoryTier(ctx stdcontext.Context, memoryID string) (context.MemoryTier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return "", fmt.Errorf("store not started")
	}

	var tier string
	err := s.db.QueryRowContext(ctx, "SELECT tier FROM memory_tiers WHERE memory_id = ?", memoryID).Scan(&tier)
	if err != nil {
		if err == sql.ErrNoRows {
			return context.MemoryTierSession, nil // 默认为会话记忆
		}
		return "", err
	}

	return context.MemoryTier(tier), nil
}

// GetMemoriesByTier 获取指定分层的记忆列表
func (s *SQLiteStore) GetMemoriesByTier(ctx stdcontext.Context, tier context.MemoryTier, limit int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	query := "SELECT memory_id FROM memory_tiers WHERE tier = ?"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, string(tier))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memoryIDs []string
	for rows.Next() {
		var memoryID string
		if err := rows.Scan(&memoryID); err != nil {
			continue
		}
		memoryIDs = append(memoryIDs, memoryID)
	}

	return memoryIDs, nil
}

// CleanupExpiredTiers 清理过期的分层记忆
func (s *SQLiteStore) CleanupExpiredTiers(ctx stdcontext.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return 0, fmt.Errorf("store not started")
	}

	now := time.Now().Unix()

	// 获取过期记忆的ID（先查询以便统计）
	rows, err := s.db.QueryContext(ctx, "SELECT memory_id FROM memory_tiers WHERE expires_at IS NOT NULL AND expires_at < ?", now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var expiredIDs []string
	for rows.Next() {
		var memoryID string
		if err := rows.Scan(&memoryID); err != nil {
			continue
		}
		expiredIDs = append(expiredIDs, memoryID)
	}

	// 删除过期记录
	result, err := s.db.ExecContext(ctx, "DELETE FROM memory_tiers WHERE expires_at IS NOT NULL AND expires_at < ?", now)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// UpdateMemoryAccess 更新记忆访问信息
func (s *SQLiteStore) UpdateMemoryAccess(ctx stdcontext.Context, memoryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	now := time.Now().Unix()

	// 增加访问计数并更新最后访问时间
	_, err := s.db.ExecContext(ctx,
		"UPDATE memory_tiers SET access_count = access_count + 1, last_accessed = ? WHERE memory_id = ?",
		now, memoryID)

	return err
}

// GetTierStatistics 获取分层的统计信息
func (s *SQLiteStore) GetTierStatistics(ctx stdcontext.Context) (map[context.MemoryTier]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	stats := make(map[context.MemoryTier]int)

	// 统计各分层的记忆数量
	rows, err := s.db.QueryContext(ctx, "SELECT tier, COUNT(*) FROM memory_tiers GROUP BY tier")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tier string
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			continue
		}
		stats[context.MemoryTier(tier)] = count
	}

	return stats, nil
}

// PromoteMemories 提升记忆到更高分层
func (s *SQLiteStore) PromoteMemories(ctx stdcontext.Context, memoryIDs []string, fromTier, toTier context.MemoryTier) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 更新分层的过期时间
	var expiresAt int64
	now := time.Now()

	switch toTier {
	case context.MemoryTierSession:
		expiresAt = now.Add(24 * time.Hour).Unix()
	case context.MemoryTierDaily:
		expiresAt = now.Add(7 * 24 * time.Hour).Unix()
	case context.MemoryTierLongTerm:
		expiresAt = now.Add(365 * 24 * time.Hour).Unix()
	}

	// 批量更新
	stmt, err := tx.Prepare("UPDATE memory_tiers SET tier = ?, expires_at = ? WHERE memory_id = ? AND tier = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, memoryID := range memoryIDs {
		if _, err := stmt.Exec(string(toTier), expiresAt, memoryID, string(fromTier)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetExpiringMemories 获取即将过期的记忆
func (s *SQLiteStore) GetExpiringMemories(ctx stdcontext.Context, within time.Duration, tier context.MemoryTier) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	cutoff := time.Now().Add(within).Unix()

	query := "SELECT memory_id FROM memory_tiers WHERE tier = ? AND expires_at IS NOT NULL AND expires_at <= ?"
	rows, err := s.db.QueryContext(ctx, query, string(tier), cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memoryIDs []string
	for rows.Next() {
		var memoryID string
		if err := rows.Scan(&memoryID); err != nil {
			continue
		}
		memoryIDs = append(memoryIDs, memoryID)
	}

	return memoryIDs, nil
}
