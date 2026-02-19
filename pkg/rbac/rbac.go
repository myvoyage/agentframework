// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package rbac

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	// ErrAccessDenied is returned when access is denied
	ErrAccessDenied = errors.New("access denied")
	// ErrRoleNotFound is returned when a role is not found
	ErrRoleNotFound = errors.New("role not found")
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidPermission is returned when permission is invalid
	ErrInvalidPermission = errors.New("invalid permission")
)

// Permission represents a specific permission
type Permission struct {
	Resource string   `json:"resource"` // Resource type (e.g., "agent", "workflow")
	Action   string   `json:"action"`   // Action (e.g., "read", "write", "execute")
	Scope    string   `json:"scope"`    // Scope (e.g., "*", "own", "specific-id")
	Conditions []Condition `json:"conditions,omitempty"` // Additional conditions
}

// Condition represents additional permission conditions
type Condition struct {
	Type  string      `json:"type"`  // Condition type (e.g., "ip", "time")
	Value interface{} `json:"value"` // Condition value
}

// String returns the string representation of a permission
func (p *Permission) String() string {
	return fmt.Sprintf("%s:%s:%s", p.Resource, p.Action, p.Scope)
}

// Matches checks if a permission matches the given resource and action
func (p *Permission) Matches(resource, action string) bool {
	return (p.Resource == "*" || p.Resource == resource) &&
		(p.Action == "*" || p.Action == action)
}

// Role represents a collection of permissions
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(resource, action, scope string) bool {
	for _, perm := range r.Permissions {
		if perm.Matches(resource, action) {
			// Check scope
			if perm.Scope == "*" || perm.Scope == scope {
				return true
			}
		}
	}
	return false
}

// RBACManager manages roles and permissions
type RBACManager struct {
	roles     map[string]*Role
	userRoles map[string][]string // userID -> role names
	mu        sync.RWMutex
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager() *RBACManager {
	return &RBACManager{
		roles:     make(map[string]*Role),
		userRoles: make(map[string][]string),
	}
}

// AddRole adds a new role
func (r *RBACManager) AddRole(role *Role) error {
	if role == nil {
		return errors.New("role cannot be nil")
	}
	if role.Name == "" {
		return errors.New("role name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.roles[role.Name] = role
	return nil
}

// GetRole retrieves a role by name
func (r *RBACManager) GetRole(name string) (*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, ok := r.roles[name]
	if !ok {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// AssignRole assigns a role to a user
func (r *RBACManager) AssignRole(userID, roleName string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if roleName == "" {
		return errors.New("role name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if role exists
	if _, ok := r.roles[roleName]; !ok {
		return ErrRoleNotFound
	}

	// Assign role
	r.userRoles[userID] = append(r.userRoles[userID], roleName)
	return nil
}

// RevokeRole revokes a role from a user
func (r *RBACManager) RevokeRole(userID, roleName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	roles, ok := r.userRoles[userID]
	if !ok {
		return ErrUserNotFound
	}

	// Find and remove the role
	newRoles := make([]string, 0, len(roles))
	found := false
	for _, rname := range roles {
		if rname != roleName {
			newRoles = append(newRoles, rname)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user %s does not have role %s", userID, roleName)
	}

	r.userRoles[userID] = newRoles
	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *RBACManager) GetUserRoles(userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles, ok := r.userRoles[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return roles, nil
}

// CheckPermission checks if a user has a specific permission
func (r *RBACManager) CheckPermission(userID, resource, action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roleNames, ok := r.userRoles[userID]
	if !ok {
		return false
	}

	for _, roleName := range roleNames {
		role := r.roles[roleName]
		if role.HasPermission(resource, action, "*") {
			return true
		}
	}

	return false
}

// CheckPermissionWithScope checks if a user has a specific permission with scope
func (r *RBACManager) CheckPermissionWithScope(userID, resource, action, scope string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roleNames, ok := r.userRoles[userID]
	if !ok {
		return false
	}

	for _, roleName := range roleNames {
		role := r.roles[roleName]
		if role.HasPermission(resource, action, scope) {
			return true
		}
	}

	return false
}

// RequirePermission is a convenience function that returns an error if permission is denied
func (r *RBACManager) RequirePermission(userID, resource, action string) error {
	if !r.CheckPermission(userID, resource, action) {
		return fmt.Errorf("%w: user %s cannot %s %s", ErrAccessDenied, userID, action, resource)
	}
	return nil
}

// GetAllRoles returns all roles
func (r *RBACManager) GetAllRoles() []*Role {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]*Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	return roles
}

// DefaultRoles returns the default system roles
func DefaultRoles() []*Role {
	return []*Role{
		{
			Name:        "admin",
			Description: "Full system access",
			Permissions: []Permission{
				{Resource: "*", Action: "*", Scope: "*"},
			},
		},
		{
			Name:        "user",
			Description: "Standard user access",
			Permissions: []Permission{
				{Resource: "agent", Action: "read", Scope: "*"},
				{Resource: "agent", Action: "execute", Scope: "own"},
				{Resource: "workflow", Action: "read", Scope: "*"},
				{Resource: "workflow", Action: "execute", Scope: "own"},
				{Resource: "workflow", Action: "create", Scope: "*"},
			},
		},
		{
			Name:        "viewer",
			Description: "Read-only access",
			Permissions: []Permission{
				{Resource: "agent", Action: "read", Scope: "*"},
				{Resource: "workflow", Action: "read", Scope: "*"},
				{Resource: "device", Action: "read", Scope: "*"},
			},
		},
		{
			Name:        "operator",
			Description: "Operations team access",
			Permissions: []Permission{
				{Resource: "*", Action: "read", Scope: "*"},
				{Resource: "system", Action: "restart", Scope: "*"},
				{Resource: "system", Action: "configure", Scope: "*"},
				{Resource: "logs", Action: "read", Scope: "*"},
			},
		},
	}
}

// InitializeDefaultRoles initializes the manager with default roles
func (r *RBACManager) InitializeDefaultRoles() error {
	for _, role := range DefaultRoles() {
		if err := r.AddRole(role); err != nil {
			return err
		}
	}
	return nil
}

// PermissionCheckResult represents the result of a permission check
type PermissionCheckResult struct {
	Allowed   bool     `json:"allowed"`
	UserID    string   `json:"user_id"`
	Resource  string   `json:"resource"`
	Action    string   `json:"action"`
	Roles     []string `json:"roles"`
	DeniedReason string `json:"denied_reason,omitempty"`
}

// CheckPermissionDetailed returns detailed permission check result
func (r *RBACManager) CheckPermissionDetailed(userID, resource, action string) *PermissionCheckResult {
	result := &PermissionCheckResult{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	roleNames, ok := r.userRoles[userID]
	if !ok {
		result.DeniedReason = "user not found"
		return result
	}

	result.Roles = roleNames

	for _, roleName := range roleNames {
		role := r.roles[roleName]
		if role.HasPermission(resource, action, "*") {
			result.Allowed = true
			return result
		}
	}

	result.DeniedReason = fmt.Sprintf("no roles grant permission to %s:%s", resource, action)
	return result
}

// WildcardMatch checks if a pattern matches a string using wildcards
func WildcardMatch(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	if pattern == str {
		return true
	}

	// Handle "resource:*" pattern
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, ":*")
		return strings.HasPrefix(str, prefix+":")
	}

	// Handle "resource:*:action" pattern
	parts := strings.Split(pattern, ":")
	strParts := strings.Split(str, ":")

	if len(parts) != 3 || len(strParts) != 3 {
		return false
	}

	// Match each part
	for i := 0; i < 3; i++ {
		if parts[i] != "*" && parts[i] != strParts[i] {
			return false
		}
	}

	return true
}
