// Agent Framework - Session Management
// Based on OpenClaw Architecture: https://www.cnblogs.com/tangshiye/p/19642495
//
// Session Types:
//   - main: User's private chat (highest permissions)
//   - dm: Direct message from others (sandboxed by default)
//   - group: Group chat (sandboxed by default)
//
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// SessionType represents the type of conversation session
type SessionType string

const (
	SessionTypeMain  SessionType = "main"  // User's own private chat
	SessionTypeDM    SessionType = "dm"    // Direct message from others
	SessionTypeGroup SessionType = "group" // Group chat
)

// Session represents a conversation session
type Session struct {
	ID         string      `json:"id"`
	Type       SessionType `json:"type"`
	Channel    string      `json:"channel"`     // Platform: telegram, lark, etc.
	ChannelID  string      `json:"channel_id"`  // Channel/Chat ID on the platform
	UserID     string      `json:"user_id"`    // User ID
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Messages   []*schema.Message `json:"messages"`
	Metadata   map[string]string `json:"metadata"`
	IsMain     bool        `json:"is_main"`     // Whether this is a main session
}

// NewSession creates a new session with the given type
func NewSession(sessionType SessionType, channel, channelID, userID string) *Session {
	return &Session{
		ID:        GenerateSessionID(sessionType, channel, channelID, userID),
		Type:      sessionType,
		Channel:   channel,
		ChannelID: channelID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  make([]*schema.Message, 0),
		Metadata:  make(map[string]string),
		IsMain:    sessionType == SessionTypeMain,
	}
}

// GenerateSessionID generates a unique session ID based on type and source
// Format: {type}:{channel}:{id} or {type}:{channel}:{channelID}:{userID}
func GenerateSessionID(sessionType SessionType, channel, channelID, userID string) string {
	switch sessionType {
	case SessionTypeMain:
		// Main session uses only channel and user
		return fmt.Sprintf("%s:%s:%s", sessionType, channel, userID)
	case SessionTypeDM:
		// DM session includes channel ID
		return fmt.Sprintf("%s:%s:%s:%s", sessionType, channel, channelID, userID)
	case SessionTypeGroup:
		// Group session uses channel and group ID
		return fmt.Sprintf("%s:%s:%s", sessionType, channel, channelID)
	default:
		return fmt.Sprintf("%s:%s:%s", sessionType, channel, channelID)
	}
}

// ParseSessionID parses a session ID into its components
func ParseSessionID(sessionID string) (SessionType, string, string, string) {
	parts := strings.Split(sessionID, ":")
	if len(parts) < 2 {
		return SessionTypeMain, "", "", ""
	}

	sessionType := SessionType(parts[0])
	channel := parts[1]

	switch sessionType {
	case SessionTypeMain:
		if len(parts) >= 3 {
			return sessionType, channel, "", parts[2]
		}
	case SessionTypeDM:
		if len(parts) >= 4 {
			return sessionType, channel, parts[2], parts[3]
		}
	case SessionTypeGroup:
		if len(parts) >= 3 {
			return sessionType, channel, parts[2], ""
		}
	}

	return sessionType, channel, "", ""
}

// IsSandboxed returns whether this session should be sandboxed
func (s *Session) IsSandboxed() bool {
	// Main sessions (user's own chat) are not sandboxed
	// DM and Group sessions are sandboxed by default
	return s.Type != SessionTypeMain
}

// HasFullPermissions returns whether this session has full permissions
func (s *Session) HasFullPermissions() bool {
	// Only main sessions have full permissions
	return s.Type == SessionTypeMain
}

// AddMessage adds a message to the session
func (s *Session) AddMessage(msg *schema.Message) {
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
}

// GetRecentMessages returns the most recent messages
func (s *Session) GetRecentMessages(count int) []*schema.Message {
	if count <= 0 {
		return s.Messages
	}
	if count >= len(s.Messages) {
		return s.Messages
	}
	return s.Messages[len(s.Messages)-count:]
}

// SessionManager manages conversation sessions
type SessionManager struct {
	sessions map[string]*Session
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// GetOrCreate gets an existing session or creates a new one
func (sm *SessionManager) GetOrCreate(sessionType SessionType, channel, channelID, userID string) *Session {
	sessionID := GenerateSessionID(sessionType, channel, channelID, userID)

	if session, exists := sm.sessions[sessionID]; exists {
		return session
	}

	session := NewSession(sessionType, channel, channelID, userID)
	sm.sessions[sessionID] = session
	return session
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(sessionID string) (*Session, bool) {
	session, ok := sm.sessions[sessionID]
	return session, ok
}

// List returns all sessions
func (sm *SessionManager) List() []*Session {
	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Delete removes a session
func (sm *SessionManager) Delete(sessionID string) {
	delete(sm.sessions, sessionID)
}

// SessionResolver resolves the session type based on message source
type SessionResolver struct {
	// OwnerID is the ID of the main session owner (user's own ID)
	OwnerID string
}

// NewSessionResolver creates a new session resolver
func NewSessionResolver(ownerID string) *SessionResolver {
	return &SessionResolver{
		OwnerID: ownerID,
	}
}

// Resolve determines the session type based on the message context
func (sr *SessionResolver) Resolve(ctx context.Context, channel, senderID, chatID string) (SessionType, string, string, string) {
	// If the sender is the owner, this is a main session
	if senderID == sr.OwnerID {
		return SessionTypeMain, channel, chatID, senderID
	}

	// If chat ID is empty or single-user, it's a DM
	if chatID == "" || isSingleUserChat(chatID) {
		return SessionTypeDM, channel, chatID, senderID
	}

	// Otherwise, it's a group chat
	return SessionTypeGroup, channel, chatID, senderID
}

// isSingleUserChat determines if a chat ID represents a single user conversation
func isSingleUserChat(chatID string) bool {
	// This can be extended with platform-specific logic
	// For now, assume empty or very short IDs are single-user
	return len(chatID) < 10
}

// ResolveFromMessage resolves session from a message
// Note: This is a simplified version - full implementation would need
// access to channel-specific message structures
func (sr *SessionResolver) ResolveFromMessage(msg *schema.Message) (SessionType, string, string, string) {
	if msg == nil {
		return SessionTypeDM, "", "", ""
	}
	// Default to DM for message-based resolution
	return SessionTypeDM, "", "", ""
}

// SessionStore defines the interface for session persistence
type SessionStore interface {
	Save(ctx context.Context, session *Session) error
	Load(ctx context.Context, sessionID string) (*Session, error)
	List(ctx context.Context) ([]*Session, error)
	Delete(ctx context.Context, sessionID string) error
}

// InMemorySessionStore implements SessionStore in memory
type InMemorySessionStore struct {
	sessions map[string]*Session
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

// Save saves a session
func (s *InMemorySessionStore) Save(ctx context.Context, session *Session) error {
	s.sessions[session.ID] = session
	return nil
}

// Load loads a session
func (s *InMemorySessionStore) Load(ctx context.Context, sessionID string) (*Session, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	return session, nil
}

// List lists all sessions
func (s *InMemorySessionStore) List(ctx context.Context) ([]*Session, error) {
	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// Delete deletes a session
func (s *InMemorySessionStore) Delete(ctx context.Context, sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}
