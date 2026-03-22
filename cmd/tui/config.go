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
// Agent Framework - TUI Config & Session Management
// Copyright (C) 2025 Agent Framework Contributors
//
// 配置和会话管�?- 借鉴 Memoh 的持久化�?token 管理模式

package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ========== 配置管理 ==========

// TUIConfig TUI 配置
type TUIConfig struct {
	// 界面设置
	Theme           string `json:"theme"`            // 主题名称
	ShowLineNumbers bool   `json:"showLineNumbers"`  // 显示行号
	FontSize        int    `json:"fontSize"`         // 字体大小（终端不适用，保留）

	// 聊天设置
	DefaultAgentID  string `json:"defaultAgentId"`   // 默认 Agent
	StreamChat      bool   `json:"streamChat"`       // 流式聊天
	AutoScroll      bool   `json:"autoScroll"`       // 自动滚动
	MaxHistory      int    `json:"maxHistory"`       // 最大历史记录数

	// 会话设置
	SessionID       string `json:"sessionId"`        // 会话 ID
	LastAgentID     string `json:"lastAgentId"`      // 最后使用的 Agent
	AutoSaveSession bool   `json:"autoSaveSession"`  // 自动保存会话

	// 性能设置
	RefreshInterval int    `json:"refreshInterval"`  // 刷新间隔（毫秒）
	EnableCache     bool    `json:"enableCache"`     // 启用缓存
}

// DefaultTUIConfig 默认配置
func DefaultTUIConfig() *TUIConfig {
	sessionID := fmt.Sprintf("tui-%d", time.Now().UnixNano())
	return &TUIConfig{
		Theme:           "default",
		ShowLineNumbers: true,
		FontSize:        14,

		DefaultAgentID:  "",
		StreamChat:      true,
		AutoScroll:      true,
		MaxHistory:      100,

		SessionID:       sessionID,
		LastAgentID:     "",
		AutoSaveSession: true,

		RefreshInterval: 5000,
		EnableCache:     true,
	}
}

// ConfigManager 配置管理�?type ConfigManager struct {
	mu     sync.RWMutex
	config *TUIConfig
	path   string
}

// NewConfigManager 创建配置管理�?func NewConfigManager(configPath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		config: DefaultTUIConfig(),
		path:   configPath,
	}

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// 加载配置
	if err := cm.Load(); err != nil {
		// 配置文件不存在，使用默认配置并保�?		if os.IsNotExist(err) {
			if err := cm.Save(); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return cm, nil
}

// Load 加载配置
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.path)
	if err != nil {
		return err
	}

	var config TUIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	cm.config = &config
	return nil
}

// Save 保存配置
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get 获取配置
func (cm *ConfigManager) Get() *TUIConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Update 更新配置
func (cm *ConfigManager) Update(fn func(*TUIConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	fn(cm.config)

	if err := cm.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ========== 会话管理 ==========

// SessionManager 会话管理�?type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	path     string
}

// Session 会话
type Session struct {
	ID        string
	AgentID   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []ChatMessageItem
	Metadata  map[string]string
}

// NewSessionManager 创建会话管理�?func NewSessionManager(sessionsPath string) (*SessionManager, error) {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		path:     sessionsPath,
	}

	// 确保会话目录存在
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// 加载现有会话
	if err := sm.LoadAll(); err != nil {
		// 不存在会话不是错�?		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load sessions: %w", err)
		}
	}

	return sm, nil
}

// Create 创建新会�?func (sm *SessionManager) Create(agentID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	session := &Session{
		ID:        sessionID,
		AgentID:   agentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  make([]ChatMessageItem, 0, 100),
		Metadata:  make(map[string]string),
	}

	sm.sessions[sessionID] = session

	// 保存会话
	if err := sm.saveSession(session); err != nil {
		delete(sm.sessions, sessionID)
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// Get 获取会话
func (sm *SessionManager) Get(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	return session, ok
}

// Update 更新会话
func (sm *SessionManager) Update(sessionID string, fn func(*Session)) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	fn(session)
	session.UpdatedAt = time.Now()

	if err := sm.saveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// Delete 删除会话
func (sm *SessionManager) Delete(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(sm.sessions, sessionID)

	// 删除会话文件
	sessionPath := filepath.Join(sm.path, sessionID+".json")
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// List 列出所有会�?func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// LoadAll 加载所有会�?func (sm *SessionManager) LoadAll() error {
	entries, err := os.ReadDir(sm.path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionPath := filepath.Join(sm.path, entry.Name())
		data, err := os.ReadFile(sessionPath)
		if err != nil {
			continue // 跳过无法读取的会�?		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue // 跳过无法解析的会�?		}

		sm.sessions[session.ID] = &session
	}

	return nil
}

// saveSession 保存会话
func (sm *SessionManager) saveSession(session *Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionPath := filepath.Join(sm.path, session.ID+".json")
	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// ========== 路径辅助函数 ==========

// GetConfigPath 获取配置文件路径
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".agentframework", "tui", "config.json"), nil
}

// GetSessionsPath 获取会话目录路径
func GetSessionsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".agentframework", "tui", "sessions"), nil
}
