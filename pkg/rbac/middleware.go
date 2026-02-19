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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"AgentFramework/pkg/errors"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserRolesKey is the context key for user roles
	UserRolesKey ContextKey = "user_roles"
)

// AuthMiddleware extracts user information from the request context
// and adds it to the request context for RBAC checks
type AuthMiddleware struct {
	rbac *RBACManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(rbac *RBACManager) *AuthMiddleware {
	return &AuthMiddleware{
		rbac: rbac,
	}
}

// Authenticate extracts user ID from the request (e.g., from JWT token)
// This is a placeholder - implement based on your auth system
func (am *AuthMiddleware) Authenticate(r *http.Request) (string, error) {
	// Extract user ID from JWT token in Authorization header
	// This is a simplified example - implement proper JWT validation
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// In production, validate the JWT token and extract user ID
	// For now, return a placeholder
	// TODO: Implement proper JWT validation
	return "user-123", nil
}

// Middleware returns an HTTP middleware that performs authentication
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := am.Authenticate(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// Get user roles
		roles, err := am.rbac.GetUserRoles(userID)
		if err != nil {
			// User exists but has no roles - create empty slice
			roles = []string{}
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRolesKey, roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission creates middleware that checks for a specific permission
func (am *AuthMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "user not authenticated", http.StatusUnauthorized)
				return
			}

			if !am.rbac.CheckPermission(userID, resource, action) {
				permissionDenied(w, r, userID, resource, action)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissionWithScope creates middleware that checks for a specific permission with scope
func (am *AuthMiddleware) RequirePermissionWithScope(resource, action, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "user not authenticated", http.StatusUnauthorized)
				return
			}

			if !am.rbac.CheckPermissionWithScope(userID, resource, action, scope) {
				permissionDenied(w, r, userID, resource, action)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates middleware that checks for a specific role
func (am *AuthMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "user not authenticated", http.StatusUnauthorized)
				return
			}

			roles, err := am.rbac.GetUserRoles(userID)
			if err != nil {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, role := range roles {
				if role == roleName {
					hasRole = true
					break
				}
			}

			if !hasRole {
				permissionDenied(w, r, userID, "role", roleName)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin creates middleware that requires admin role
func (am *AuthMiddleware) RequireAdmin() func(http.Handler) http.Handler {
	return am.RequireRole("admin")
}

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserRoles retrieves the user roles from the request context
func GetUserRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(UserRolesKey).([]string); ok {
		return roles
	}
	return nil
}

// permissionDenied writes a permission denied response
func permissionDenied(w http.ResponseWriter, r *http.Request, userID, resource, action string) {
	result := &PermissionCheckResult{
		Allowed:   false,
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		DeniedReason: fmt.Sprintf("insufficient permissions"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	json.NewEncoder(w).Encode(result)
}

// AuditLog logs permission checks for audit purposes
type AuditLog struct {
	rbac *RBACManager
}

// NewAuditLog creates a new audit logger
func NewAuditLog(rbac *RBACManager) *AuditLog {
	return &AuditLog{rbac: rbac}
}

// LogPermissionCheck logs a permission check
func (al *AuditLog) LogPermissionCheck(userID, resource, action string, allowed bool) {
	result := al.rbac.CheckPermissionDetailed(userID, resource, action)

	// In production, send this to a logging system
	// For now, just format it
	fmt.Printf("AUDIT: user=%s resource=%s action=%s allowed=%v roles=%v reason=%s\n",
		userID, resource, action, allowed, result.Roles, result.DeniedReason)
}

// AuditMiddleware creates middleware that logs all permission checks
func (al *AuditLog) AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())

		// Extract resource and action from the request
		resource := extractResource(r)
		action := extractAction(r)

		// Check permission
		allowed := al.rbac.CheckPermission(userID, resource, action)

		// Log the check
		al.LogPermissionCheck(userID, resource, action, allowed)

		// If not allowed, return error
		if !allowed {
			permissionDenied(w, r, userID, resource, action)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractResource extracts the resource from the request
func extractResource(r *http.Request) string {
	// Extract from URL path
	// /api/agents/{id} -> "agent"
	// /api/workflows/{id} -> "workflow"
	path := r.URL.Path
	if len(path) > 5 && path[:5] == "/api/" {
		parts := splitPath(path[5:])
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return "unknown"
}

// extractAction extracts the action from the request method
func extractAction(r *http.Request) string {
	switch r.Method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return r.Method
	}
}

// splitPath splits a URL path into parts
func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}

	parts := []string{}
	current := ""

	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// PermissionChecker provides a helper for checking permissions in handlers
type PermissionChecker struct {
	rbac *RBACManager
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(rbac *RBACManager) *PermissionChecker {
	return &PermissionChecker{rbac: rbac}
}

// Check checks if the request context has permission
func (pc *PermissionChecker) Check(r *http.Request, resource, action string) error {
	userID := GetUserID(r.Context())
	if userID == "" {
		return errors.New("user not authenticated")
	}

	if !pc.rbac.CheckPermission(userID, resource, action) {
		return fmt.Errorf("%w: cannot %s %s", ErrAccessDenied, action, resource)
	}

	return nil
}

// CheckWithScope checks if the request context has permission with scope
func (pc *PermissionChecker) CheckWithScope(r *http.Request, resource, action, scope string) error {
	userID := GetUserID(r.Context())
	if userID == "" {
		return errors.New("user not authenticated")
	}

	if !pc.rbac.CheckPermissionWithScope(userID, resource, action, scope) {
		return fmt.Errorf("%w: cannot %s %s with scope %s", ErrAccessDenied, action, resource, scope)
	}

	return nil
}

// MustAdmin checks if the user has admin role
func (pc *PermissionChecker) MustAdmin(r *http.Request) error {
	userID := GetUserID(r.Context())
	if userID == "" {
		return errors.New("user not authenticated")
	}

	if !pc.rbac.CheckPermission(userID, "*", "*") {
		return fmt.Errorf("%w: admin access required", ErrAccessDenied)
	}

	return nil
}
