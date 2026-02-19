// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package security provides security policies and access control.
package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// SecurityPolicy defines the interface for security policies.
type SecurityPolicy interface {
	// ValidateCommand validates a command before execution.
	ValidateCommand(ctx context.Context, cmd string, params map[string]interface{}) error

	// CheckPermissions checks if a user has permission to access a resource.
	CheckPermissions(ctx context.Context, userID string, resource string, action string) (bool, error)

	// EncryptData encrypts data.
	EncryptData(ctx context.Context, data []byte) ([]byte, error)

	// DecryptData decrypts data.
	DecryptData(ctx context.Context, encrypted []byte) ([]byte, error)

	// AuditLog logs an audit event.
	AuditLog(ctx context.Context, event AuditEvent) error
}

// AuditEvent represents an audit event.
type AuditEvent struct {
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
}

// SecurityManager implements security policies.
type SecurityManager struct {
	permissions     map[string]map[string][]string // userID -> resource -> actions
	auditLog        []AuditEvent
	auditLogMaxSize int
	encryptionKey   []byte
	mutex           sync.RWMutex
	enabledMetrics  bool
}

// NewSecurityManager creates a new SecurityManager instance.
func NewSecurityManager(encryptionKey []byte, auditLogMaxSize int, metricsEnabled bool) *SecurityManager {
	// Derive a proper key from the provided key
	key := sha256.Sum256(encryptionKey)

	return &SecurityManager{
		permissions:     make(map[string]map[string][]string),
		auditLog:        make([]AuditEvent, 0, auditLogMaxSize),
		auditLogMaxSize: auditLogMaxSize,
		encryptionKey:   key[:],
		enabledMetrics:  metricsEnabled,
	}
}

// ValidateCommand validates a command before execution.
func (m *SecurityManager) ValidateCommand(ctx context.Context, cmd string, params map[string]interface{}) error {
	// Check if command is allowed
	if !isCommandAllowed(cmd) {
		return fmt.Errorf("command not allowed: %s", cmd)
	}

	// Check parameter safety
	if err := validateParameters(params); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	return nil
}

// CheckPermissions checks if a user has permission to access a resource.
func (m *SecurityManager) CheckPermissions(ctx context.Context, userID string, resource string, action string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	userPerms, exists := m.permissions[userID]
	if !exists {
		return false, fmt.Errorf("user not found: %s", userID)
	}

	resourceActions, exists := userPerms[resource]
	if !exists {
		return false, fmt.Errorf("resource not found for user: %s", resource)
	}

	for _, allowedAction := range resourceActions {
		if allowedAction == action || allowedAction == "*" {
			return true, nil
		}
	}

	return false, nil
}

// EncryptData encrypts data using AES-GCM.
func (m *SecurityManager) EncryptData(ctx context.Context, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData decrypts data using AES-GCM.
func (m *SecurityManager) DecryptData(ctx context.Context, encrypted []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// AuditLog logs an audit event.
func (m *SecurityManager) AuditLog(ctx context.Context, event AuditEvent) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to audit log
	m.auditLog = append(m.auditLog, event)

	// Trim audit log if it exceeds max size
	if len(m.auditLog) > m.auditLogMaxSize {
		m.auditLog = m.auditLog[1:]
	}

	return nil
}

// GrantPermission grants a user permission to perform an action on a resource.
func (m *SecurityManager) GrantPermission(ctx context.Context, userID string, resource string, action string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.permissions[userID] == nil {
		m.permissions[userID] = make(map[string][]string)
	}

	if m.permissions[userID][resource] == nil {
		m.permissions[userID][resource] = make([]string, 0)
	}

	m.permissions[userID][resource] = append(m.permissions[userID][resource], action)

	return nil
}

// RevokePermission revokes a user's permission.
func (m *SecurityManager) RevokePermission(ctx context.Context, userID string, resource string, action string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	userPerms, exists := m.permissions[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	resourceActions, exists := userPerms[resource]
	if !exists {
		return fmt.Errorf("resource not found for user: %s", resource)
	}

	for i, a := range resourceActions {
		if a == action {
			m.permissions[userID][resource] = append(resourceActions[:i], resourceActions[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("action not found: %s", action)
}

// GetAuditLog returns the audit log.
func (m *SecurityManager) GetAuditLog(ctx context.Context, limit int) ([]AuditEvent, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if limit <= 0 || limit > len(m.auditLog) {
		limit = len(m.auditLog)
	}

	start := len(m.auditLog) - limit
	return m.auditLog[start:], nil
}

// ClearAuditLog clears the audit log.
func (m *SecurityManager) ClearAuditLog(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.auditLog = make([]AuditEvent, 0, m.auditLogMaxSize)
	return nil
}

// isCommandAllowed checks if a command is allowed.
func isCommandAllowed(cmd string) bool {
	// Define dangerous commands
	dangerousCommands := map[string]bool{
		"rm":        true,
		"delete":    true,
		"format":    true,
		"shutdown":  true,
		"reboot":    true,
		"kill":      true,
		"su":        true,
		"sudo":      true,
		"chmod":     true,
		"chown":     true,
		"mv":        true,
		"cp":        true,
		"cat":       true,
		"vi":        true,
		"vim":       true,
		"nano":      true,
		"wget":      true,
		"curl":      true,
		"nc":        true,
		"telnet":    true,
		"ssh":       true,
		"ftp":       true,
		"passwd":    true,
		"useradd":   true,
		"userdel":   true,
		"groupadd":  true,
		"groupdel":  true,
		"iptables":  true,
		"firewall":  true,
		"systemctl": true,
		"service":   true,
	}

	// Check if command is dangerous
	if dangerous, exists := dangerousCommands[cmd]; exists && dangerous {
		return false
	}

	return true
}

// validateParameters validates command parameters for safety.
func validateParameters(params map[string]interface{}) error {
	for key, value := range params {
		// Check for dangerous patterns
		if str, ok := value.(string); ok {
			if containsDangerousPattern(str) {
				return fmt.Errorf("parameter %s contains dangerous pattern", key)
			}
		}
	}

	return nil
}

// containsDangerousPattern checks if a string contains dangerous patterns.
func containsDangerousPattern(s string) bool {
	dangerousPatterns := []string{
		";",
		"|",
		"&",
		"$(",
		"`",
		"$((",
		"${",
		"../",
		"..\\",
		"/etc/",
		"/root/",
		"~/.ssh",
	}

	for _, pattern := range dangerousPatterns {
		if contains(s, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOf(s, substr) >= 0)
}

// indexOf returns the index of a substring in a string.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// EncryptToBase64 encrypts data and returns it as a base64 string.
func (m *SecurityManager) EncryptToBase64(ctx context.Context, data []byte) (string, error) {
	encrypted, err := m.EncryptData(ctx, data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptFromBase64 decrypts a base64 string and returns the data.
func (m *SecurityManager) DecryptFromBase64(ctx context.Context, encrypted string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return m.DecryptData(ctx, data)
}

// AccessControlList represents an access control list.
type AccessControlList struct {
	entries []ACLEntry
	mutex   sync.RWMutex
}

// ACLEntry represents an access control list entry.
type ACLEntry struct {
	Principal string   `json:"principal"`
	Resource  string   `json:"resource"`
	Actions   []string `json:"actions"`
	Effect    string   `json:"effect"` // "allow" or "deny"
}

// NewAccessControlList creates a new AccessControlList instance.
func NewAccessControlList() *AccessControlList {
	return &AccessControlList{
		entries: make([]ACLEntry, 0),
	}
}

// AddEntry adds an entry to the ACL.
func (acl *AccessControlList) AddEntry(ctx context.Context, entry ACLEntry) error {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()

	acl.entries = append(acl.entries, entry)
	return nil
}

// RemoveEntry removes an entry from the ACL.
func (acl *AccessControlList) RemoveEntry(ctx context.Context, principal string, resource string) error {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()

	for i, entry := range acl.entries {
		if entry.Principal == principal && entry.Resource == resource {
			acl.entries = append(acl.entries[:i], acl.entries[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("ACL entry not found")
}

// Check checks if an action is allowed.
func (acl *AccessControlList) Check(ctx context.Context, principal string, resource string, action string) (bool, error) {
	acl.mutex.RLock()
	defer acl.mutex.RUnlock()

	for _, entry := range acl.entries {
		if entry.Principal == principal && entry.Resource == resource {
			for _, allowedAction := range entry.Actions {
				if allowedAction == action || allowedAction == "*" {
					return entry.Effect == "allow", nil
				}
			}
		}
	}

	return false, nil // Deny by default
}