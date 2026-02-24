// AgentFramework - Local Configuration Store
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// 本地优先架构：SQLite 配置和对话历史存储

package local

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store 本地数据存储接口
type Store interface {
	// 配置管理
	SaveConfig(key string, value interface{}) error
	GetConfig(key string, value interface{}) error
	ListConfigs() (map[string]interface{}, error)

	// Agent 配置
	SaveAgentConfig(cfg AgentConfig) error
	GetAgentConfig(name string) (AgentConfig, error)
	ListAgentConfigs() ([]AgentConfig, error)
	DeleteAgentConfig(name string) error

	// 对话历史
	SaveConversation(conv Conversation) error
	GetConversation(id string) (Conversation, error)
	GetConversationsByAgent(agentID string, limit int) ([]Conversation, error)
	ListConversations(limit int) ([]Conversation, error)
	DeleteConversation(id string) error

	// 工作流状态
	SaveWorkflowState(state WorkflowState) error
	GetWorkflowState(id string) (WorkflowState, error)
	ListWorkflowStates() ([]WorkflowState, error)
	DeleteWorkflowState(id string) error

	// 技能缓存
	CacheSkill(skill SkillCache) error
	GetCachedSkill(name string) (SkillCache, error)
	ListCachedSkills() ([]SkillCache, error)
	DeleteCachedSkill(name string) error

	// 统计信息
	IncrementStat(name string) (int64, error)
	GetStat(name string) (int64, error)
	ResetStats() error

	// 关闭连接
	Close() error
}

// ========== 数据类型 ==========

// AgentConfig Agent 配置
type AgentConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Model       string                 `json:"model"`
	SystemPrompt string                `json:"system_prompt"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Tools       []string               `json:"tools"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Conversation 对话记录
type Conversation struct {
	ID        string          `json:"id"`
	AgentID   string          `json:"agent_id"`
	Title     string          `json:"title"`
	Messages  []Message       `json:"messages"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Message 消息
type Message struct {
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowState 工作流状态
type WorkflowState struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed"
	Definition  string                 `json:"definition"`
	Input       string                 `json:"input"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SkillCache 技能缓存
type SkillCache struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Code        string                 `json:"code"`
	Metadata    map[string]interface{} `json:"metadata"`
	CachedAt    time.Time              `json:"cached_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// ========== SQLite 实现 ==========

// SQLiteStore SQLite 存储实现
type SQLiteStore struct {
	db   *sql.DB
	mu   sync.RWMutex
	path string
}

// NewSQLiteStore 创建新的 SQLite 存储实例
func NewSQLiteStore(dataDir string) (*SQLiteStore, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 创建数据库文件路径
	dbPath := filepath.Join(dataDir, "agentframework.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接参数
	db.SetMaxOpenConns(1) // SQLite 不支持多个写入者
	db.SetMaxIdleConns(1)

	store := &SQLiteStore{
		db:   db,
		path: dbPath,
	}

	// 初始化数据库表结构
	if err := store.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return store, nil
}

// init 初始化数据库表
func (s *SQLiteStore) init() error {
	schema := `
	-- 配置表
	CREATE TABLE IF NOT EXISTS configs (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Agent 配置表
	CREATE TABLE IF NOT EXISTS agent_configs (
		name TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		model TEXT,
		system_prompt TEXT,
		temperature REAL DEFAULT 0.7,
		max_tokens INTEGER DEFAULT 2000,
		tools TEXT,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 对话表
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_conversations_agent ON conversations(agent_id);
	CREATE INDEX IF NOT EXISTS idx_conversations_updated ON conversations(updated_at DESC);

	-- 消息表
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id);

	-- 工作流状态表
	CREATE TABLE IF NOT EXISTS workflow_states (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		definition TEXT,
		input TEXT,
		output TEXT,
		error TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_workflow_status ON workflow_states(status);

	-- 技能缓存表
	CREATE TABLE IF NOT EXISTS skill_cache (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		version TEXT,
		code TEXT,
		metadata TEXT,
		cached_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_skill_name ON skill_cache(name);
	CREATE INDEX IF NOT EXISTS idx_skill_expires ON skill_cache(expires_at);

	-- 统计表
	CREATE TABLE IF NOT EXISTS stats (
		name TEXT PRIMARY KEY,
		value INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// ========== 配置管理 ==========

func (s *SQLiteStore) SaveConfig(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO configs (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		key, string(data),
	)
	return err
}

func (s *SQLiteStore) GetConfig(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data string
	err := s.db.QueryRow("SELECT value FROM configs WHERE key = ?", key).Scan(&data)
	if err == sql.ErrNoRows {
		return fmt.Errorf("config not found: %s", key)
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), value)
}

func (s *SQLiteStore) ListConfigs() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT key, value FROM configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]interface{})
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		var v interface{}
		if err := json.Unmarshal([]byte(value), &v); err == nil {
			configs[key] = v
		}
	}

	return configs, rows.Err()
}

// ========== Agent 配置 ==========

func (s *SQLiteStore) SaveAgentConfig(cfg AgentConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolsJSON, _ := json.Marshal(cfg.Tools)
	metadataJSON, _ := json.Marshal(cfg.Metadata)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO agent_configs
		(name, type, model, system_prompt, temperature, max_tokens, tools, metadata, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, cfg.Name, cfg.Type, cfg.Model, cfg.SystemPrompt,
		cfg.Temperature, cfg.MaxTokens, string(toolsJSON), string(metadataJSON))

	return err
}

func (s *SQLiteStore) GetAgentConfig(name string) (AgentConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cfg AgentConfig
	var toolsJSON, metadataJSON string

	err := s.db.QueryRow(`
		SELECT name, type, model, system_prompt, temperature, max_tokens, tools, metadata, created_at, updated_at
		FROM agent_configs WHERE name = ?
	`, name).Scan(
		&cfg.Name, &cfg.Type, &cfg.Model, &cfg.SystemPrompt,
		&cfg.Temperature, &cfg.MaxTokens, &toolsJSON, &metadataJSON,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return cfg, fmt.Errorf("agent config not found: %s", name)
	}
	if err != nil {
		return cfg, err
	}

	json.Unmarshal([]byte(toolsJSON), &cfg.Tools)
	json.Unmarshal([]byte(metadataJSON), &cfg.Metadata)

	return cfg, nil
}

func (s *SQLiteStore) ListAgentConfigs() ([]AgentConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT name, type, model, system_prompt, temperature, max_tokens, tools, metadata, created_at, updated_at
		FROM agent_configs ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []AgentConfig
	for rows.Next() {
		var cfg AgentConfig
		var toolsJSON, metadataJSON string

		if err := rows.Scan(
			&cfg.Name, &cfg.Type, &cfg.Model, &cfg.SystemPrompt,
			&cfg.Temperature, &cfg.MaxTokens, &toolsJSON, &metadataJSON,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(toolsJSON), &cfg.Tools)
		json.Unmarshal([]byte(metadataJSON), &cfg.Metadata)

		configs = append(configs, cfg)
	}

	return configs, rows.Err()
}

func (s *SQLiteStore) DeleteAgentConfig(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM agent_configs WHERE name = ?", name)
	return err
}

// ========== 对话历史 ==========

func (s *SQLiteStore) SaveConversation(conv Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除旧消息
	if _, err := tx.Exec("DELETE FROM messages WHERE conversation_id = ?", conv.ID); err != nil {
		return err
	}

	// 插入或更新对话
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO conversations (id, agent_id, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, conv.ID, conv.AgentID, conv.Title, conv.CreatedAt)
	if err != nil {
		return err
	}

	// 插入消息
	for _, msg := range conv.Messages {
		metadataJSON, _ := json.Marshal(msg.Metadata)
		_, err = tx.Exec(`
			INSERT INTO messages (conversation_id, role, content, timestamp, metadata)
			VALUES (?, ?, ?, ?, ?)
		`, conv.ID, msg.Role, msg.Content, msg.Timestamp, string(metadataJSON))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetConversation(id string) (Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var conv Conversation
	err := s.db.QueryRow(`
		SELECT id, agent_id, title, created_at, updated_at
		FROM conversations WHERE id = ?
	`, id).Scan(&conv.ID, &conv.AgentID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)

	if err == sql.ErrNoRows {
		return conv, fmt.Errorf("conversation not found: %s", id)
	}
	if err != nil {
		return conv, err
	}

	// 加载消息
	rows, err := s.db.Query(`
		SELECT role, content, timestamp, metadata FROM messages
		WHERE conversation_id = ? ORDER BY timestamp ASC
	`, id)
	if err != nil {
		return conv, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		var metadataJSON string
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.Timestamp, &metadataJSON); err != nil {
			return conv, err
		}
		json.Unmarshal([]byte(metadataJSON), &msg.Metadata)
		conv.Messages = append(conv.Messages, msg)
	}

	return conv, rows.Err()
}

func (s *SQLiteStore) GetConversationsByAgent(agentID string, limit int) ([]Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, agent_id, title, created_at, updated_at
		FROM conversations WHERE agent_id = ?
		ORDER BY updated_at DESC LIMIT ?
	`, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.AgentID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}

	return convs, rows.Err()
}

func (s *SQLiteStore) ListConversations(limit int) ([]Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, agent_id, title, created_at, updated_at
		FROM conversations ORDER BY updated_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.AgentID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}

	return convs, rows.Err()
}

func (s *SQLiteStore) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM conversations WHERE id = ?", id)
	return err
}

// ========== 工作流状态 ==========

func (s *SQLiteStore) SaveWorkflowState(state WorkflowState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadataJSON, _ := json.Marshal(state.Metadata)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO workflow_states
		(id, name, status, definition, input, output, error, started_at, completed_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, state.ID, state.Name, state.Status, state.Definition, state.Input,
		state.Output, state.Error, state.StartedAt, state.CompletedAt, string(metadataJSON))

	return err
}

func (s *SQLiteStore) GetWorkflowState(id string) (WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state WorkflowState
	var metadataJSON string

	err := s.db.QueryRow(`
		SELECT id, name, status, definition, input, output, error, started_at, completed_at, metadata
		FROM workflow_states WHERE id = ?
	`, id).Scan(
		&state.ID, &state.Name, &state.Status, &state.Definition, &state.Input,
		&state.Output, &state.Error, &state.StartedAt, &state.CompletedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return state, fmt.Errorf("workflow state not found: %s", id)
	}
	if err != nil {
		return state, err
	}

	json.Unmarshal([]byte(metadataJSON), &state.Metadata)
	return state, nil
}

func (s *SQLiteStore) ListWorkflowStates() ([]WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, name, status, definition, input, output, error, started_at, completed_at, metadata
		FROM workflow_states ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []WorkflowState
	for rows.Next() {
		var state WorkflowState
		var metadataJSON string

		if err := rows.Scan(
			&state.ID, &state.Name, &state.Status, &state.Definition, &state.Input,
			&state.Output, &state.Error, &state.StartedAt, &state.CompletedAt, &metadataJSON,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(metadataJSON), &state.Metadata)
		states = append(states, state)
	}

	return states, rows.Err()
}

func (s *SQLiteStore) DeleteWorkflowState(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM workflow_states WHERE id = ?", id)
	return err
}

// ========== 技能缓存 ==========

func (s *SQLiteStore) CacheSkill(skill SkillCache) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadataJSON, _ := json.Marshal(skill.Metadata)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO skill_cache
		(id, name, description, version, code, metadata, cached_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`, skill.ID, skill.Name, skill.Description, skill.Version, skill.Code,
		string(metadataJSON), skill.ExpiresAt)

	return err
}

func (s *SQLiteStore) GetCachedSkill(name string) (SkillCache, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var skill SkillCache
	var metadataJSON string

	err := s.db.QueryRow(`
		SELECT id, name, description, version, code, metadata, cached_at, expires_at
		FROM skill_cache WHERE name = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, name).Scan(
		&skill.ID, &skill.Name, &skill.Description, &skill.Version,
		&skill.Code, &metadataJSON, &skill.CachedAt, &skill.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return skill, fmt.Errorf("skill not found in cache: %s", name)
	}
	if err != nil {
		return skill, err
	}

	json.Unmarshal([]byte(metadataJSON), &skill.Metadata)
	return skill, nil
}

func (s *SQLiteStore) ListCachedSkills() ([]SkillCache, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, name, description, version, code, metadata, cached_at, expires_at
		FROM skill_cache WHERE expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP
		ORDER BY cached_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []SkillCache
	for rows.Next() {
		var skill SkillCache
		var metadataJSON string

		if err := rows.Scan(
			&skill.ID, &skill.Name, &skill.Description, &skill.Version,
			&skill.Code, &metadataJSON, &skill.CachedAt, &skill.ExpiresAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(metadataJSON), &skill.Metadata)
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

func (s *SQLiteStore) DeleteCachedSkill(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM skill_cache WHERE name = ?", name)
	return err
}

// ========== 统计信息 ==========

func (s *SQLiteStore) IncrementStat(name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO stats (name, value) VALUES (?, 1)
		ON CONFLICT(name) DO UPDATE SET value = value + 1, updated_at = CURRENT_TIMESTAMP
	`, name)
	if err != nil {
		return 0, err
	}

	var value int64
	err = s.db.QueryRow("SELECT value FROM stats WHERE name = ?", name).Scan(&value)
	return value, err
}

func (s *SQLiteStore) GetStat(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var value int64
	err := s.db.QueryRow("SELECT value FROM stats WHERE name = ?", name).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return value, err
}

func (s *SQLiteStore) ResetStats() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM stats")
	return err
}

// ========== 关闭连接 ==========

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ========== 默认存储实例 ==========

var (
	defaultStore *SQLiteStore
	storeOnce    sync.Once
)

// GetDefaultStore 获取默认存储实例
func GetDefaultStore() (Store, error) {
	var initErr error
	storeOnce.Do(func() {
		// 获取默认数据目录
		dataDir, err := getDefaultDataDir()
		if err != nil {
			initErr = err
			return
		}

		defaultStore, initErr = NewSQLiteStore(dataDir)
	})

	return defaultStore, initErr
}

// getDefaultDataDir 获取默认数据目录
func getDefaultDataDir() (string, error) {
	// 优先使用环境变量
	if dir := os.Getenv("AGENT_FRAMEWORK_DATA_DIR"); dir != "" {
		return dir, nil
	}

	// 使用用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".agentframework"), nil
}

// InitDefaultStore 初始化默认存储
func InitDefaultStore(dataDir string) error {
	var err error
	defaultStore, err = NewSQLiteStore(dataDir)
	return err
}
