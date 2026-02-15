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

package auth

import (
	"testing"
	"time"
)

func TestNewAuthModule(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	if module == nil {
		t.Fatal("Auth module is nil")
	}

	if module.jwtManager == nil {
		t.Fatal("JWT manager is nil")
	}

	if module.apiKeyMgr == nil {
		t.Fatal("API key manager is nil")
	}

	if module.permChecker == nil {
		t.Fatal("Permission checker is nil")
	}
}

func TestJWTManager_Generate(t *testing.T) {
	manager := &JWTManager{
		secretKey: []byte("test-secret"),
		issuer:    "test-issuer",
		expiry:    time.Hour,
	}

	token, err := manager.Generate("user123", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}
}

func TestJWTManager_Verify(t *testing.T) {
	manager := &JWTManager{
		secretKey: []byte("test-secret"),
		issuer:    "test-issuer",
		expiry:    time.Hour,
	}

	// Generate a token
	token, err := manager.Generate("user123", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify the token
	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", claims.UserID)
	}

	if len(claims.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
	}
}

func TestJWTManager_VerifyInvalidToken(t *testing.T) {
	manager := &JWTManager{
		secretKey: []byte("test-secret"),
		issuer:    "test-issuer",
		expiry:    time.Hour,
	}

	// Try to verify an invalid token
	_, err := manager.Verify("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}

func TestAPIKeyManager_Generate(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	apiKey, err := manager.Generate("user123", []string{"read", "write"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if apiKey.Key == "" {
		t.Fatal("Generated API key is empty")
	}

	if apiKey.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", apiKey.UserID)
	}

	if len(apiKey.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(apiKey.Permissions))
	}
}

func TestAPIKeyManager_Verify(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Generate an API key
	apiKey, err := manager.Generate("user123", []string{"read", "write"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Verify the API key
	verified, err := manager.Verify(apiKey.Key)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}

	if verified.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", verified.UserID)
	}
}

func TestAPIKeyManager_VerifyInvalidKey(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Try to verify an invalid key
	_, err := manager.Verify("invalid-key")
	if err == nil {
		t.Fatal("Expected error for invalid key, got nil")
	}
}

func TestAPIKeyManager_VerifyExpiredKey(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Create an expired API key
	expiredKey := &APIKey{
		Key:         "expired-key",
		UserID:      "user123",
		Permissions: []string{"read"},
		CreatedAt:   time.Now().Add(-2 * 24 * time.Hour),
		ExpiresAt:   time.Now().Add(-24 * time.Hour),
	}
	manager.keys[expiredKey.Key] = expiredKey

	// Try to verify the expired key
	_, err := manager.Verify(expiredKey.Key)
	if err == nil {
		t.Fatal("Expected error for expired key, got nil")
	}
}

func TestAPIKeyManager_Revoke(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Generate an API key
	apiKey, err := manager.Generate("user123", []string{"read", "write"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Revoke the API key
	err = manager.Revoke(apiKey.Key)
	if err != nil {
		t.Fatalf("Failed to revoke API key: %v", err)
	}

	// Try to verify the revoked key
	_, err = manager.Verify(apiKey.Key)
	if err == nil {
		t.Fatal("Expected error for revoked key, got nil")
	}
}

func TestAPIKeyManager_List(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Generate multiple API keys
	manager.Generate("user123", []string{"read"}, 365)
	manager.Generate("user123", []string{"write"}, 365)
	manager.Generate("user456", []string{"admin"}, 365)

	// List all keys
	allKeys := manager.List("")
	if len(allKeys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(allKeys))
	}

	// List keys for specific user
	user123Keys := manager.List("user123")
	if len(user123Keys) != 2 {
		t.Errorf("Expected 2 keys for user123, got %d", len(user123Keys))
	}
}

func TestPermissionChecker_HasPermission(t *testing.T) {
	checker := &PermissionChecker{
		roles: make(map[string][]string),
	}

	tests := []struct {
		name        string
		permissions []string
		required    string
		expected    bool
	}{
		{
			name:        "Exact match",
			permissions: []string{"read", "write"},
			required:    "read",
			expected:    true,
		},
		{
			name:        "No match",
			permissions: []string{"read", "write"},
			required:    "delete",
			expected:    false,
		},
		{
			name:        "Wildcard permission",
			permissions: []string{"*"},
			required:    "anything",
			expected:    true,
		},
		{
			name:        "Empty permissions",
			permissions: []string{},
			required:    "read",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.HasPermission(tt.permissions, tt.required)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPermissionChecker_HasAnyPermission(t *testing.T) {
	checker := &PermissionChecker{
		roles: make(map[string][]string),
	}

	permissions := []string{"read", "write"}
	required := []string{"delete", "write", "admin"}

	result := checker.HasAnyPermission(permissions, required)
	if !result {
		t.Error("Expected true, got false")
	}
}

func TestPermissionChecker_HasAllPermissions(t *testing.T) {
	checker := &PermissionChecker{
		roles: make(map[string][]string),
	}

	tests := []struct {
		name        string
		permissions []string
		required    []string
		expected    bool
	}{
		{
			name:        "Has all permissions",
			permissions: []string{"read", "write", "delete"},
			required:    []string{"read", "write"},
			expected:    true,
		},
		{
			name:        "Missing one permission",
			permissions: []string{"read", "write"},
			required:    []string{"read", "write", "delete"},
			expected:    false,
		},
		{
			name:        "Wildcard covers all",
			permissions: []string{"*"},
			required:    []string{"read", "write", "delete"},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.HasAllPermissions(tt.permissions, tt.required)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPermissionChecker_AddRole(t *testing.T) {
	checker := &PermissionChecker{
		roles: make(map[string][]string),
	}

	checker.AddRole("editor", []string{"read", "write"})

	permissions := checker.GetRolePermissions("editor")
	if len(permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(permissions))
	}
}

func TestAuthModule_GenerateToken(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	result, err := module.generateToken("user123", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if result["token"].(string) == "" {
		t.Error("Expected token to be non-empty")
	}
}

func TestAuthModule_VerifyToken(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Generate a token
	genResult, err := module.generateToken("user123", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token := genResult["token"].(string)

	// Verify the token
	verifyResult, err := module.verifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if !verifyResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if !verifyResult["valid"].(bool) {
		t.Error("Expected valid to be true")
	}
}

func TestAuthModule_GenerateAPIKey(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	result, err := module.generateAPIKey("user123", []string{"read", "write"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if result["api_key"].(string) == "" {
		t.Error("Expected API key to be non-empty")
	}
}

func TestAuthModule_VerifyAPIKey(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Generate an API key
	genResult, err := module.generateAPIKey("user123", []string{"read", "write"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	apiKey := genResult["api_key"].(string)

	// Verify the API key
	verifyResult, err := module.verifyAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}

	if !verifyResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if !verifyResult["valid"].(bool) {
		t.Error("Expected valid to be true")
	}
}

func TestAuthModule_CheckPermission(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	result, err := module.checkPermission([]string{"read", "write"}, "read")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if !result["has_permission"].(bool) {
		t.Error("Expected has_permission to be true")
	}
}

func TestAuthModule_GetStats(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Perform some operations
	module.generateToken("user123", []string{"read"})
	module.generateAPIKey("user123", []string{"read"}, 365)

	stats := module.GetStats()

	if stats["total_requests"] < 2 {
		t.Errorf("Expected at least 2 total requests, got %d", stats["total_requests"])
	}

	if stats["tokens_generated"] < 1 {
		t.Errorf("Expected at least 1 token generated, got %d", stats["tokens_generated"])
	}
}

func TestAuthModule_Close(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Generate some API keys
	module.generateAPIKey("user123", []string{"read"}, 365)

	// Close the module
	err = module.Close()
	if err != nil {
		t.Fatalf("Failed to close module: %v", err)
	}

	// Verify keys are cleared
	keys := module.apiKeyMgr.List("")
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys after close, got %d", len(keys))
	}
}

func TestAuthModule_DefaultConfig(t *testing.T) {
	// Test with empty config to verify defaults
	config := AuthConfig{
		Enable: true,
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	if module.config.JWTSecret == "" {
		t.Error("Expected default JWT secret to be generated")
	}

	if module.config.JWTExpiry != 3600 {
		t.Errorf("Expected default expiry 3600, got %d", module.config.JWTExpiry)
	}

	if module.config.JWTIssuer != "aio-sandbox" {
		t.Errorf("Expected default issuer 'aio-sandbox', got '%s'", module.config.JWTIssuer)
	}
}

func TestAuthModule_GetToolsDisabled(t *testing.T) {
	config := AuthConfig{
		Enable: false,
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	tools, err := module.GetTools(nil)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 0 {
		t.Errorf("Expected 0 tools when disabled, got %d", len(tools))
	}
}

func TestAuthModule_GetToolsEnabled(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	tools, err := module.GetTools(nil)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 7 {
		t.Errorf("Expected 7 tools, got %d", len(tools))
	}
}

func TestAuthModule_VerifyInvalidToken(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Verify an invalid token
	result, err := module.verifyToken("invalid-token")
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Expected success to be false for invalid token")
	}

	if result["valid"].(bool) {
		t.Error("Expected valid to be false for invalid token")
	}
}

func TestAuthModule_VerifyInvalidAPIKey(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Verify an invalid API key
	result, err := module.verifyAPIKey("invalid-key")
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Expected success to be false for invalid key")
	}

	if result["valid"].(bool) {
		t.Error("Expected valid to be false for invalid key")
	}
}

func TestAuthModule_CheckPermissionNoMatch(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	result, err := module.checkPermission([]string{"read", "write"}, "delete")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if result["has_permission"].(bool) {
		t.Error("Expected has_permission to be false")
	}
}

func TestAuthModule_ExportedMethods(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Test exported GenerateToken
	tokenResult, err := module.GenerateToken("user123", []string{"read"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if !tokenResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Test exported VerifyToken
	token := tokenResult["token"].(string)
	verifyResult, err := module.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}
	if !verifyResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Test exported GenerateAPIKey
	apiKeyResult, err := module.GenerateAPIKey("user123", []string{"read"}, 365)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}
	if !apiKeyResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Test exported VerifyAPIKey
	apiKey := apiKeyResult["api_key"].(string)
	verifyAPIResult, err := module.VerifyAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}
	if !verifyAPIResult["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Test exported CheckPermission
	permResult, err := module.CheckPermission([]string{"read"}, "read")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}
	if !permResult["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

func TestAPIKeyManager_DefaultExpiry(t *testing.T) {
	manager := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// Generate with 0 expiry days (should default to 365)
	apiKey, err := manager.Generate("user123", []string{"read"}, 0)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	expectedExpiry := time.Now().AddDate(0, 0, 365)
	if apiKey.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) || apiKey.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Error("Expected default expiry to be 365 days")
	}
}

func TestJWTManager_VerifyWrongSigningMethod(t *testing.T) {
	manager := &JWTManager{
		secretKey: []byte("test-secret"),
		issuer:    "test-issuer",
		expiry:    time.Hour,
	}

	// Create a token with a different signing method (this would need to be manually crafted)
	// For now, we'll just test with a malformed token
	_, err := manager.Verify("eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjM0NTY3ODkwIn0.")
	if err == nil {
		t.Error("Expected error for token with wrong signing method")
	}
}
