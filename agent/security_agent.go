// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package agent provides security agent implementation.
package agent

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"AgentFramework/pkg/beads/security"
)

// SecurityAgent manages security policies and access control.
type SecurityAgent struct {
	securityManager *security.SecurityManager
	acl             *security.AccessControlList
	enabled         bool
	auditEnabled    bool
	mutex           sync.RWMutex
}

// NewSecurityAgent creates a new SecurityAgent instance.
func NewSecurityAgent(encryptionKey string, auditLogMaxSize int) *SecurityAgent {
	keyBytes := []byte(encryptionKey)
	securityManager := security.NewSecurityManager(keyBytes, auditLogMaxSize, true)

	return &SecurityAgent{
		securityManager: securityManager,
		acl:             security.NewAccessControlList(),
		enabled:         true,
		auditEnabled:    true,
	}
}

// Initialize initializes the security agent.
func (a *SecurityAgent) Initialize(ctx context.Context) error {
	// Log initialization event
	return a.AuditLog(ctx, security.AuditEvent{
		EventType: "security_init",
		Status:    "success",
		Details: map[string]interface{}{
			"timestamp": time.Now(),
		},
	})
}

// ValidateCommand validates a command before execution.
func (a *SecurityAgent) ValidateCommand(ctx context.Context, cmd string, params map[string]interface{}) error {
	if !a.enabled {
		return nil // Security disabled
	}

	if err := a.securityManager.ValidateCommand(ctx, cmd, params); err != nil {
		// Log validation failure
		if a.auditEnabled {
			a.AuditLog(ctx, security.AuditEvent{
				EventType: "command_validation",
				Status:    "failed",
				Details: map[string]interface{}{
					"command": cmd,
					"params":  params,
					"error":   err.Error(),
				},
			})
		}
		return err
	}

	// Log validation success
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "command_validation",
			Status:    "success",
			Details: map[string]interface{}{
				"command": cmd,
			},
		})
	}

	return nil
}

// CheckPermissions checks if a user has permission to access a resource.
func (a *SecurityAgent) CheckPermissions(ctx context.Context, userID string, resource string, action string) (bool, error) {
	if !a.enabled {
		return true, nil // Security disabled
	}

	allowed, err := a.securityManager.CheckPermissions(ctx, userID, resource, action)

	// Log permission check
	if a.auditEnabled {
		status := "success"
		if !allowed || err != nil {
			status = "failed"
		}

		a.AuditLog(ctx, security.AuditEvent{
			EventType: "permission_check",
			UserID:    userID,
			Resource:  resource,
			Action:    action,
			Status:    status,
			Details: map[string]interface{}{
				"allowed": allowed,
			},
		})
	}

	return allowed, err
}

// EncryptData encrypts data.
func (a *SecurityAgent) EncryptData(ctx context.Context, data []byte) ([]byte, error) {
	if !a.enabled {
		return data, nil // Security disabled
	}

	return a.securityManager.EncryptData(ctx, data)
}

// DecryptData decrypts data.
func (a *SecurityAgent) DecryptData(ctx context.Context, encrypted []byte) ([]byte, error) {
	if !a.enabled {
		return encrypted, nil // Security disabled
	}

	return a.securityManager.DecryptData(ctx, encrypted)
}

// EncryptToBase64 encrypts data and returns it as a base64 string.
func (a *SecurityAgent) EncryptToBase64(ctx context.Context, data []byte) (string, error) {
	if !a.enabled {
		return base64.StdEncoding.EncodeToString(data), nil // Security disabled
	}

	return a.securityManager.EncryptToBase64(ctx, data)
}

// DecryptFromBase64 decrypts a base64 string and returns the data.
func (a *SecurityAgent) DecryptFromBase64(ctx context.Context, encrypted string) ([]byte, error) {
	if !a.enabled {
		return base64.StdEncoding.DecodeString(encrypted)
	}

	return a.securityManager.DecryptFromBase64(ctx, encrypted)
}

// AuditLog logs an audit event.
func (a *SecurityAgent) AuditLog(ctx context.Context, event security.AuditEvent) error {
	if !a.auditEnabled {
		return nil
	}

	return a.securityManager.AuditLog(ctx, event)
}

// GrantPermission grants a user permission to perform an action on a resource.
func (a *SecurityAgent) GrantPermission(ctx context.Context, userID string, resource string, action string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.securityManager.GrantPermission(ctx, userID, resource, action); err != nil {
		return err
	}

	// Log permission grant
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "permission_grant",
			Status:    "success",
			Details: map[string]interface{}{
				"user_id":  userID,
				"resource": resource,
				"action":   action,
			},
		})
	}

	return nil
}

// RevokePermission revokes a user's permission.
func (a *SecurityAgent) RevokePermission(ctx context.Context, userID string, resource string, action string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.securityManager.RevokePermission(ctx, userID, resource, action); err != nil {
		return err
	}

	// Log permission revoke
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "permission_revoke",
			Status:    "success",
			Details: map[string]interface{}{
				"user_id":  userID,
				"resource": resource,
				"action":   action,
			},
		})
	}

	return nil
}

// GetAuditLog returns the audit log.
func (a *SecurityAgent) GetAuditLog(ctx context.Context, limit int) ([]security.AuditEvent, error) {
	return a.securityManager.GetAuditLog(ctx, limit)
}

// ClearAuditLog clears the audit log.
func (a *SecurityAgent) ClearAuditLog(ctx context.Context) error {
	return a.securityManager.ClearAuditLog(ctx)
}

// AddACLEntry adds an entry to the access control list.
func (a *SecurityAgent) AddACLEntry(ctx context.Context, principal string, resource string, actions []string, effect string) error {
	entry := security.ACLEntry{
		Principal: principal,
		Resource:  resource,
		Actions:   actions,
		Effect:    effect,
	}

	if err := a.acl.AddEntry(ctx, entry); err != nil {
		return err
	}

	// Log ACL entry addition
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "acl_entry_add",
			Status:    "success",
			Details: map[string]interface{}{
				"principal": principal,
				"resource":  resource,
				"actions":   actions,
				"effect":    effect,
			},
		})
	}

	return nil
}

// RemoveACLEntry removes an entry from the access control list.
func (a *SecurityAgent) RemoveACLEntry(ctx context.Context, principal string, resource string) error {
	if err := a.acl.RemoveEntry(ctx, principal, resource); err != nil {
		return err
	}

	// Log ACL entry removal
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "acl_entry_remove",
			Status:    "success",
			Details: map[string]interface{}{
				"principal": principal,
				"resource":  resource,
			},
		})
	}

	return nil
}

// CheckACL checks if an action is allowed via ACL.
func (a *SecurityAgent) CheckACL(ctx context.Context, principal string, resource string, action string) (bool, error) {
	if !a.enabled {
		return true, nil // Security disabled
	}

	return a.acl.Check(ctx, principal, resource, action)
}

// Enable enables security.
func (a *SecurityAgent) Enable(ctx context.Context) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.enabled = true

	// Log security enable
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "security_enable",
			Status:    "success",
			Details: map[string]interface{}{
				"timestamp": time.Now(),
			},
		})
	}
}

// Disable disables security.
func (a *SecurityAgent) Disable(ctx context.Context) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.enabled = false

	// Log security disable
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "security_disable",
			Status:    "success",
			Details: map[string]interface{}{
				"timestamp": time.Now(),
			},
		})
	}
}

// IsEnabled returns whether security is enabled.
func (a *SecurityAgent) IsEnabled(ctx context.Context) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.enabled
}

// EnableAudit enables audit logging.
func (a *SecurityAgent) EnableAudit(ctx context.Context) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.auditEnabled = true
}

// DisableAudit disables audit logging.
func (a *SecurityAgent) DisableAudit(ctx context.Context) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.auditEnabled = false
}

// IsAuditEnabled returns whether audit logging is enabled.
func (a *SecurityAgent) IsAuditEnabled(ctx context.Context) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.auditEnabled
}

// GetSecurityStatus returns the current security status.
func (a *SecurityAgent) GetSecurityStatus(ctx context.Context) *SecurityStatus {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return &SecurityStatus{
		Enabled:      a.enabled,
		AuditEnabled: a.auditEnabled,
		Timestamp:    time.Now(),
	}
}

// SecurityStatus contains security status information.
type SecurityStatus struct {
	Enabled      bool      `json:"enabled"`
	AuditEnabled bool      `json:"audit_enabled"`
	Timestamp    time.Time `json:"timestamp"`
}

// Close closes the security agent and cleans up resources.
func (a *SecurityAgent) Close(ctx context.Context) error {
	// Log security agent shutdown
	if a.auditEnabled {
		a.AuditLog(ctx, security.AuditEvent{
			EventType: "security_shutdown",
			Status:    "success",
			Details: map[string]interface{}{
				"timestamp": time.Now(),
			},
		})
	}

	return nil
}