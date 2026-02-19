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
	"testing"
)

func TestRBACManager_AddRole(t *testing.T) {
	manager := NewRBACManager()

	role := &Role{
		Name:        "test-role",
		Description: "Test role",
		Permissions: []Permission{
			{Resource: "agent", Action: "read", Scope: "*"},
		},
	}

	err := manager.AddRole(role)
	if err != nil {
		t.Fatalf("AddRole() failed: %v", err)
	}

	retrieved, err := manager.GetRole("test-role")
	if err != nil {
		t.Fatalf("GetRole() failed: %v", err)
	}

	if retrieved.Name != role.Name {
		t.Errorf("retrieved role name = %v, want %v", retrieved.Name, role.Name)
	}
}

func TestRBACManager_AssignRole(t *testing.T) {
	manager := NewRBACManager()

	// Create a role
	role := &Role{
		Name: "user",
		Permissions: []Permission{
			{Resource: "agent", Action: "read", Scope: "*"},
		},
	}
	manager.AddRole(role)

	// Assign role to user
	err := manager.AssignRole("user-123", "user")
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	// Check user roles
	roles, err := manager.GetUserRoles("user-123")
	if err != nil {
		t.Fatalf("GetUserRoles() failed: %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf(" GetUserRoles() returned %d roles, want 1", len(roles))
	}

	if roles[0] != "user" {
		t.Errorf("user role = %v, want user", roles[0])
	}
}

func TestRBACManager_CheckPermission(t *testing.T) {
	manager := NewRBACManager()

	// Initialize default roles
	err := manager.InitializeDefaultRoles()
	if err != nil {
		t.Fatalf("InitializeDefaultRoles() failed: %v", err)
	}

	// Assign user role
	err = manager.AssignRole("user-123", "user")
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	tests := []struct {
		name     string
		userID   string
		resource string
		action   string
		want     bool
	}{
		{
			name:     "user can read agents",
			userID:   "user-123",
			resource: "agent",
			action:   "read",
			want:     true,
		},
		{
			name:     "user cannot write agents",
			userID:   "user-123",
			resource: "agent",
			action:   "write",
			want:     false,
		},
		{
			name:     "user can read workflows",
			userID:   "user-123",
			resource: "workflow",
			action:   "read",
			want:     true,
		},
		{
			name:     "user can execute own workflows",
			userID:   "user-123",
			resource: "workflow",
			action:   "execute",
			want:     true,
		},
		{
			name:     "user can create workflows",
			userID:   "user-123",
			resource: "workflow",
			action:   "create",
			want:     true,
		},
		{
			name:     "non-existent user has no permissions",
			userID:   "non-existent",
			resource: "agent",
			action:   "read",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.CheckPermission(tt.userID, tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRBACManager_AdminPermissions(t *testing.T) {
	manager := NewRBACManager()
	manager.InitializeDefaultRoles()

	// Assign admin role
	err := manager.AssignRole("admin-123", "admin")
	if err != nil {
		t.Fatalf("AssignRole() failed: %v", err)
	}

	// Admin should have all permissions
	tests := []struct {
		resource string
		action   string
	}{
		{"agent", "read"},
		{"agent", "write"},
		{"agent", "delete"},
		{"workflow", "read"},
		{"workflow", "write"},
		{"system", "configure"},
	}

	for _, tt := range tests {
		t.Run(tt.resource+"/"+tt.action, func(t *testing.T) {
			if !manager.CheckPermission("admin-123", tt.resource, tt.action) {
				t.Errorf("admin should have permission to %s:%s", tt.resource, tt.action)
			}
		})
	}
}

func TestRBACManager_RevokeRole(t *testing.T) {
	manager := NewRBACManager()

	// Create and assign role
	role := &Role{
		Name: "test-role",
		Permissions: []Permission{
			{Resource: "agent", Action: "read", Scope: "*"},
		},
	}
	manager.AddRole(role)
	manager.AssignRole("user-123", "test-role")

	// Revoke role
	err := manager.RevokeRole("user-123", "test-role")
	if err != nil {
		t.Fatalf("RevokeRole() failed: %v", err)
	}

	// Check that role is removed
	roles, _ := manager.GetUserRoles("user-123")
	if len(roles) != 0 {
		t.Errorf("user still has roles after revocation: %v", roles)
	}

	// Check that permission is removed
	if manager.CheckPermission("user-123", "agent", "read") {
		t.Error("user still has permission after role revocation")
	}
}

func TestRBACManager_CheckPermissionDetailed(t *testing.T) {
	manager := NewRBACManager()
	manager.InitializeDefaultRoles()
	manager.AssignRole("user-123", "user")

	result := manager.CheckPermissionDetailed("user-123", "agent", "read")

	if !result.Allowed {
		t.Error("expected permission to be allowed")
	}

	if result.UserID != "user-123" {
		t.Errorf("userID = %v, want user-123", result.UserID)
	}

	if result.Resource != "agent" {
		t.Errorf("resource = %v, want agent", result.Resource)
	}

	if result.Action != "read" {
		t.Errorf("action = %v, want read", result.Action)
	}

	if len(result.Roles) != 1 {
		t.Errorf("roles length = %d, want 1", len(result.Roles))
	}
}

func TestPermission_Matches(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		resource string
		action   string
		want     bool
	}{
		{
			name: "exact match",
			perm: Permission{Resource: "agent", Action: "read", Scope: "*"},
			resource: "agent",
			action:   "read",
			want:     true,
		},
		{
			name: "wildcard resource",
			perm: Permission{Resource: "*", Action: "read", Scope: "*"},
			resource: "agent",
			action:   "read",
			want:     true,
		},
		{
			name: "wildcard action",
			perm: Permission{Resource: "agent", Action: "*", Scope: "*"},
			resource: "agent",
			action:   "read",
			want:     true,
		},
		{
			name: "no match",
			perm: Permission{Resource: "workflow", Action: "read", Scope: "*"},
			resource: "agent",
			action:   "read",
			want:     false,
		},
		{
			name: "partial match",
			perm: Permission{Resource: "agent", Action: "read", Scope: "*"},
			resource: "agent",
			action:   "write",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.perm.Matches(tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWildcardMatch(t *testing.T) {
	tests := []struct {
		pattern string
		str     string
		want    bool
	}{
		{"*", "anything", true},
		{"*", "", true},
		{"agent:read:*", "agent:read:*", true},
		{"agent:read:*", "agent:read:own", true},
		{"agent:*:*", "agent:read:own", true},
		{"*:*:*", "agent:read:own", true},
		{"agent:read:*", "workflow:read:*", false},
		{"agent:read:*", "agent:write:*", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" vs "+tt.str, func(t *testing.T) {
			got := WildcardMatch(tt.pattern, tt.str)
			if got != tt.want {
				t.Errorf("WildcardMatch(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.want)
			}
		})
	}
}

func TestRBACManager_Concurrent(t *testing.T) {
	manager := NewRBACManager()
	manager.InitializeDefaultRoles()

	const numUsers = 100
	const numOps = 100

	// Concurrent operations
	done := make(chan bool, numUsers)

	for i := 0; i < numUsers; i++ {
		userID := fmt.Sprintf("user-%d", i)
		manager.AssignRole(userID, "user")

		go func(uid string) {
			for j := 0; j < numOps; j++ {
				manager.CheckPermission(uid, "agent", "read")
				manager.GetUserRoles(uid)
			}
			done <- true
		}(userID)
	}

	// Wait for all goroutines
	for i := 0; i < numUsers; i++ {
		<-done
	}
}

// BenchmarkRBACCheck benchmarks permission checking
func BenchmarkRBACCheck(b *testing.B) {
	manager := NewRBACManager()
	manager.InitializeDefaultRoles()
	manager.AssignRole("user-123", "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckPermission("user-123", "agent", "read")
	}
}

// BenchmarkRBACCheckDetailed benchmarks detailed permission checking
func BenchmarkRBACCheckDetailed(b *testing.B) {
	manager := NewRBACManager()
	manager.InitializeDefaultRoles()
	manager.AssignRole("user-123", "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckPermissionDetailed("user-123", "agent", "read")
	}
}
