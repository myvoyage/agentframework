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
	"fmt"
)

// Schema 定义上下文存储的数据库 Schema
// 支持 SQLite 和其他关系型数据库
var Schema = []string{
	// ===== 上下文表 =====
	`CREATE TABLE IF NOT EXISTS contexts (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		workspace TEXT,
		uri TEXT,
		parent_id TEXT,
		version INTEGER DEFAULT 1,
		metadata TEXT,  -- JSON
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		accessed_at INTEGER NOT NULL,
		FOREIGN KEY (parent_id) REFERENCES contexts(id) ON DELETE SET NULL
	);`,

	// ===== 上下文层级表 =====
	`CREATE TABLE IF NOT EXISTS context_layers (
		context_id TEXT NOT NULL,
		layer_type TEXT NOT NULL,  -- l0/l1/l2
		content TEXT NOT NULL,
		tokens INTEGER NOT NULL,
		metadata TEXT,  -- JSON (method, generated_at, etc)
		version INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (context_id, layer_type, version),
		FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
	);`,

	// ===== 上下文索引表 (用于快速查询) =====
	`CREATE TABLE IF NOT EXISTS context_index (
		context_id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		workspace TEXT,
		title TEXT,
		l0_tokens INTEGER,
		l1_tokens INTEGER,
		l2_tokens INTEGER,
		total_tokens INTEGER,
		memory_count INTEGER DEFAULT 0,
		access_count INTEGER DEFAULT 0,
		FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
	);`,

	// ===== 记忆表 =====
	`CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		context_id TEXT NOT NULL,
		memory_type TEXT NOT NULL,  -- profile/preference/entity/event/case/pattern
		title TEXT,
		content TEXT NOT NULL,  -- JSON encoded memory
		metadata TEXT,  -- JSON
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
	);`,

	// ===== 记忆去重表 (基于内容哈希) =====
	`CREATE TABLE IF NOT EXISTS memory_dedup (
		content_hash TEXT PRIMARY KEY,
		canonical_id TEXT NOT NULL,  -- 规范化的记忆ID
		duplicate_ids TEXT,  -- JSON array of duplicate IDs
		first_seen INTEGER NOT NULL,
		last_seen INTEGER NOT NULL,
		count INTEGER DEFAULT 1
	);`,

	// ===== 记忆分层表 (支持会话/每日/长期记忆) =====
	`CREATE TABLE IF NOT EXISTS memory_tiers (
		memory_id TEXT PRIMARY KEY,
		tier TEXT NOT NULL,  -- session/daily/longterm
		created_at INTEGER NOT NULL,
		expires_at INTEGER,
		access_count INTEGER DEFAULT 0,
		last_accessed INTEGER NOT NULL,
		importance_score REAL DEFAULT 0.5,  -- 0-1
		metadata TEXT,  -- JSON
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);`,

	// ===== 任务-上下文关联表 =====
	`CREATE TABLE IF NOT EXISTS task_contexts (
		task_id TEXT NOT NULL,
		context_id TEXT NOT NULL,
		association_type TEXT NOT NULL,  -- auto/manual/inferred
		associated_at INTEGER NOT NULL,
		PRIMARY KEY (task_id, context_id),
		FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
	);`,

	// ===== 上下文事件表 (事件溯源) =====
	`CREATE TABLE IF NOT EXISTS context_events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		context_id TEXT NOT NULL,
		task_id TEXT,
		data TEXT NOT NULL,  -- JSON
		version INTEGER NOT NULL,
		signature TEXT,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
	);`,

	// ===== VFS 文件表 =====
	`CREATE TABLE IF NOT EXISTS vfs_files (
		uri TEXT PRIMARY KEY,
		scheme TEXT NOT NULL,  -- viking
		workspace TEXT,
		path TEXT NOT NULL,
		name TEXT,
		type TEXT,  -- file/dir/symlink
		size INTEGER,
		mode TEXT,
		mod_time INTEGER,
		layers TEXT,  -- JSON (layer availability)
		metadata TEXT,  -- JSON
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);`,

	// ===== VFS 内容表 =====
	`CREATE TABLE IF NOT EXISTS vfs_content (
		uri TEXT NOT NULL,
		layer TEXT NOT NULL,  -- l0/l1/l2
		content TEXT,
		tokens INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (uri, layer),
		FOREIGN KEY (uri) REFERENCES vfs_files(uri) ON DELETE CASCADE
	);`,
}

// Indexes 定义数据库索引
var Indexes = []string{
	// contexts 表索引
	`CREATE INDEX IF NOT EXISTS idx_contexts_type ON contexts(type);`,
	`CREATE INDEX IF NOT EXISTS idx_contexts_workspace ON contexts(workspace);`,
	`CREATE INDEX IF NOT EXISTS idx_contexts_parent ON contexts(parent_id);`,
	`CREATE INDEX IF NOT EXISTS idx_contexts_updated ON contexts(updated_at);`,
	`CREATE INDEX IF NOT EXISTS idx_contexts_created ON contexts(created_at);`,

	// context_layers 表索引
	`CREATE INDEX IF NOT EXISTS idx_context_layers_context ON context_layers(context_id);`,
	`CREATE INDEX IF NOT EXISTS idx_context_layers_type ON context_layers(layer_type);`,

	// context_index 表索引
	`CREATE INDEX IF NOT EXISTS idx_context_index_type ON context_index(type);`,
	`CREATE INDEX IF NOT EXISTS idx_context_index_workspace ON context_index(workspace);`,
	`CREATE INDEX IF NOT EXISTS idx_context_index_tokens ON context_index(total_tokens);`,

	// memories 表索引
	`CREATE INDEX IF NOT EXISTS idx_memories_context ON memories(context_id);`,
	`CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(memory_type);`,
	`CREATE INDEX IF NOT EXISTS idx_memories_updated ON memories(updated_at);`,
	`CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);`,

	// memory_dedup 表索引
	`CREATE INDEX IF NOT EXISTS idx_memory_dedup_hash ON memory_dedup(content_hash);`,

	// memory_tiers 表索引
	`CREATE INDEX IF NOT EXISTS idx_memory_tiers_tier ON memory_tiers(tier);`,
	`CREATE INDEX IF NOT EXISTS idx_memory_tiers_expires ON memory_tiers(expires_at);`,
	`CREATE INDEX IF NOT EXISTS idx_memory_tiers_importance ON memory_tiers(importance_score);`,
	`CREATE INDEX IF NOT EXISTS idx_memory_tiers_accessed ON memory_tiers(last_accessed);`,

	// task_contexts 表索引
	`CREATE INDEX IF NOT EXISTS idx_task_contexts_task ON task_contexts(task_id);`,
	`CREATE INDEX IF NOT EXISTS idx_task_contexts_context ON task_contexts(context_id);`,
	`CREATE INDEX IF NOT EXISTS idx_task_contexts_type ON task_contexts(association_type);`,

	// context_events 表索引
	`CREATE INDEX IF NOT EXISTS idx_context_events_context ON context_events(context_id);`,
	`CREATE INDEX IF NOT EXISTS idx_context_events_type ON context_events(event_type);`,
	`CREATE INDEX IF NOT EXISTS idx_context_events_timestamp ON context_events(timestamp);`,

	// vfs_files 表索引
	`CREATE INDEX IF NOT EXISTS idx_vfs_files_scheme ON vfs_files(scheme);`,
	`CREATE INDEX IF NOT EXISTS idx_vfs_files_workspace ON vfs_files(workspace);`,
	`CREATE INDEX IF NOT EXISTS idx_vfs_files_path ON vfs_files(path);`,
	`CREATE INDEX IF NOT EXISTS idx_vfs_files_type ON vfs_files(type);`,

	// vfs_content 表索引
	`CREATE INDEX IF NOT EXISTS idx_vfs_content_uri ON vfs_content(uri);`,
	`CREATE INDEX IF NOT EXISTS idx_vfs_content_layer ON vfs_content(layer);`,
}

// GetSchemaSQL 获取创建表的 SQL 语句
func GetSchemaSQL() string {
	sql := ""
	for _, table := range Schema {
		sql += table + "\n\n"
	}
	return sql
}

// GetIndexesSQL 获取创建索引的 SQL 语句
func GetIndexesSQL() string {
	sql := ""
	for _, index := range Indexes {
		sql += index + "\n"
	}
	return sql
}

// GetFullSchemaSQL 获取完整的 Schema SQL（包括表和索引）
func GetFullSchemaSQL() string {
	return GetSchemaSQL() + "\n" + GetIndexesSQL()
}

// ValidateSchema 验证 Schema 是否有效
func ValidateSchema() error {
	if len(Schema) == 0 {
		return fmt.Errorf("schema is empty")
	}
	return nil
}

// TableExistsSQL 生成检查表是否存在的 SQL
func TableExistsSQL(tableName string) string {
	return fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s'", tableName)
}

// ColumnExistsSQL 生成检查列是否存在的 SQL
func ColumnExistsSQL(tableName, columnName string) string {
	return fmt.Sprintf("PRAGMA table_info(%s)", tableName)
}

// AddColumnSQL 生成添加列的 SQL
func AddColumnSQL(tableName, columnName, columnType string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
}
