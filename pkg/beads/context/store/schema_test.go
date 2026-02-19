// Agent Framework - Context Store Schema Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package store

import (
	"strings"
	"testing"
)

// TestSchema 测试 Schema 常量
func TestSchema(t *testing.T) {
	if len(Schema) == 0 {
		t.Error("expected Schema to have elements")
	}

	// Check that it contains expected tables
	schemaSQL := GetSchemaSQL()
	if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS contexts") {
		t.Error("expected contexts table in schema")
	}
	if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS context_layers") {
		t.Error("expected context_layers table in schema")
	}
	if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS memories") {
		t.Error("expected memories table in schema")
	}
	if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS vfs_files") {
		t.Error("expected vfs_files table in schema")
	}
}

// TestIndexes 测试 Indexes 常量
func TestIndexes(t *testing.T) {
	if len(Indexes) == 0 {
		t.Error("expected Indexes to have elements")
	}

	// Check that it contains expected indexes
	indexesSQL := GetIndexesSQL()
	if !strings.Contains(indexesSQL, "CREATE INDEX") {
		t.Error("expected CREATE INDEX in indexes")
	}
	if !strings.Contains(indexesSQL, "idx_contexts_type") {
		t.Error("expected idx_contexts_type index")
	}
}

// TestGetSchemaSQL 测试获取 Schema SQL
func TestGetSchemaSQL(t *testing.T) {
	sql := GetSchemaSQL()
	if sql == "" {
		t.Error("expected non-empty SQL")
	}

	// Should contain CREATE TABLE statements
	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("expected CREATE TABLE in SQL")
	}

	// Should contain multiple tables
	tableCount := strings.Count(sql, "CREATE TABLE IF NOT EXISTS")
	if tableCount < 5 {
		t.Errorf("expected at least 5 tables, got %d", tableCount)
	}
}

// TestGetIndexesSQL 测试获取索引 SQL
func TestGetIndexesSQL(t *testing.T) {
	sql := GetIndexesSQL()
	if sql == "" {
		t.Error("expected non-empty SQL")
	}

	// Should contain CREATE INDEX statements
	if !strings.Contains(sql, "CREATE INDEX") {
		t.Error("expected CREATE INDEX in SQL")
	}

	// Should contain multiple indexes
	indexCount := strings.Count(sql, "CREATE INDEX IF NOT EXISTS")
	if indexCount < 10 {
		t.Errorf("expected at least 10 indexes, got %d", indexCount)
	}
}

// TestGetFullSchemaSQL 测试获取完整 Schema SQL
func TestGetFullSchemaSQL(t *testing.T) {
	sql := GetFullSchemaSQL()
	if sql == "" {
		t.Error("expected non-empty SQL")
	}

	// Should contain both tables and indexes
	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("expected CREATE TABLE in full schema")
	}
	if !strings.Contains(sql, "CREATE INDEX") {
		t.Error("expected CREATE INDEX in full schema")
	}

	// Tables should come before indexes
	tablePos := strings.Index(sql, "CREATE TABLE")
	indexPos := strings.Index(sql, "CREATE INDEX")
	if tablePos == -1 || indexPos == -1 {
		t.Error("could not find table or index position")
	} else if tablePos > indexPos {
		t.Error("expected tables before indexes in full schema")
	}
}

// TestValidateSchema 测试验证 Schema
func TestValidateSchema(t *testing.T) {
	err := ValidateSchema()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTableExistsSQL 测试生成表存在检查 SQL
func TestTableExistsSQL(t *testing.T) {
	tableName := "contexts"
	sql := TableExistsSQL(tableName)

	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !strings.Contains(sql, tableName) {
		t.Errorf("expected table name '%s' in SQL", tableName)
	}
	if !strings.Contains(sql, "sqlite_master") {
		t.Error("expected sqlite_master in SQL")
	}
	if !strings.Contains(sql, "type='table'") {
		t.Error("expected type='table' in SQL")
	}
}

// TestColumnExistsSQL 测试生成列存在检查 SQL
func TestColumnExistsSQL(t *testing.T) {
	tableName := "contexts"
	columnName := "id"
	sql := ColumnExistsSQL(tableName, columnName)

	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !strings.Contains(sql, tableName) {
		t.Errorf("expected table name '%s' in SQL", tableName)
	}
	if !strings.Contains(sql, "PRAGMA") {
		t.Error("expected PRAGMA in SQL")
	}
	if !strings.Contains(sql, "table_info") {
		t.Error("expected table_info in SQL")
	}
}

// TestAddColumnSQL 测试生成添加列 SQL
func TestAddColumnSQL(t *testing.T) {
	tableName := "contexts"
	columnName := "new_column"
	columnType := "TEXT"
	sql := AddColumnSQL(tableName, columnName, columnType)

	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !strings.Contains(sql, "ALTER TABLE") {
		t.Error("expected ALTER TABLE in SQL")
	}
	if !strings.Contains(sql, tableName) {
		t.Errorf("expected table name '%s' in SQL", tableName)
	}
	if !strings.Contains(sql, "ADD COLUMN") {
		t.Error("expected ADD COLUMN in SQL")
	}
	if !strings.Contains(sql, columnName) {
		t.Errorf("expected column name '%s' in SQL", columnName)
	}
	if !strings.Contains(sql, columnType) {
		t.Errorf("expected column type '%s' in SQL", columnType)
	}
}

// TestSchema_Completeness 测试 Schema 完整性
func TestSchema_Completeness(t *testing.T) {
	// Verify all expected tables exist
	expectedTables := []string{
		"contexts",
		"context_layers",
		"context_index",
		"memories",
		"memory_dedup",
		"memory_tiers",
		"task_contexts",
		"context_events",
		"vfs_files",
		"vfs_content",
	}

	schemaSQL := GetSchemaSQL()
	for _, table := range expectedTables {
		if !strings.Contains(schemaSQL, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Errorf("expected table '%s' in schema", table)
		}
	}
}

// TestIndexes_Completeness 测试索引完整性
func TestIndexes_Completeness(t *testing.T) {
	// Verify all expected indexes exist
	expectedIndexes := []string{
		"idx_contexts_type",
		"idx_contexts_workspace",
		"idx_contexts_parent",
		"idx_context_layers_context",
		"idx_memories_context",
		"idx_memories_type",
		"idx_vfs_files_path",
		"idx_vfs_content_uri",
	}

	indexesSQL := GetIndexesSQL()
	for _, index := range expectedIndexes {
		if !strings.Contains(indexesSQL, index) {
			t.Errorf("expected index '%s' in indexes", index)
		}
	}
}
