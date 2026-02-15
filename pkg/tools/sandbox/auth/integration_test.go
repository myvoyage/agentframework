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
	"context"
	"encoding/json"
	"testing"
)

// TestMCPTools_Integration tests all MCP tools in an integrated workflow
func TestMCPTools_Integration(t *testing.T) {
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
	defer module.Close()

	ctx := context.Background()

	// Get all tools
	tools, err := module.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 7 {
		t.Fatalf("Expected 7 tools, got %d", len(tools))
	}

	// Cast tools to their specific types
	var genTokenTool *authGenerateTokenTool
	var verifyTokenTool *authVerifyTokenTool
	var genAPIKeyTool *authGenerateAPIKeyTool
	var verifyAPIKeyTool *authVerifyAPIKeyTool
	var checkPermTool *authCheckPermissionTool

	for _, tool := range tools {
		switch t := tool.(type) {
		case *authGenerateTokenTool:
			genTokenTool = t
		case *authVerifyTokenTool:
			verifyTokenTool = t
		case *authGenerateAPIKeyTool:
			genAPIKeyTool = t
		case *authVerifyAPIKeyTool:
			verifyAPIKeyTool = t
		case *authCheckPermissionTool:
			checkPermTool = t
		}
	}

	// Test 1: Generate JWT Token
	t.Run("GenerateToken", func(t *testing.T) {
		info, err := genTokenTool.Info(ctx)
		if err != nil {
			t.Fatalf("Failed to get tool info: %v", err)
		}

		if info.Name != "auth_generate_token" {
			t.Errorf("Expected tool name 'auth_generate_token', got '%s'", info.Name)
		}

		input := map[string]any{
			"user_id":     "user123",
			"permissions": []string{"read", "write"},
		}
		inputJSON, _ := json.Marshal(input)

		output, err := genTokenTool.InvokableRun(ctx, string(inputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !result["success"].(bool) {
			t.Error("Expected success to be true")
		}

		if result["token"].(string) == "" {
			t.Error("Expected token to be non-empty")
		}
	})

	// Test 2: Verify JWT Token
	t.Run("VerifyToken", func(t *testing.T) {
		// First generate a token
		input := map[string]any{
			"user_id":     "user123",
			"permissions": []string{"read", "write"},
		}
		inputJSON, _ := json.Marshal(input)
		output, _ := genTokenTool.InvokableRun(ctx, string(inputJSON))

		var genResult map[string]any
		json.Unmarshal([]byte(output), &genResult)
		token := genResult["token"].(string)

		// Now verify it
		verifyInput := map[string]any{
			"token": token,
		}
		verifyInputJSON, _ := json.Marshal(verifyInput)

		verifyOutput, err := verifyTokenTool.InvokableRun(ctx, string(verifyInputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var verifyResult map[string]any
		if err := json.Unmarshal([]byte(verifyOutput), &verifyResult); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !verifyResult["success"].(bool) {
			t.Error("Expected success to be true")
		}

		if !verifyResult["valid"].(bool) {
			t.Error("Expected valid to be true")
		}
	})

	// Test 3: Generate API Key
	t.Run("GenerateAPIKey", func(t *testing.T) {
		input := map[string]any{
			"user_id":     "user456",
			"permissions": []string{"admin"},
			"expiry_days": 30,
		}
		inputJSON, _ := json.Marshal(input)

		output, err := genAPIKeyTool.InvokableRun(ctx, string(inputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !result["success"].(bool) {
			t.Error("Expected success to be true")
		}

		if result["api_key"].(string) == "" {
			t.Error("Expected api_key to be non-empty")
		}
	})

	// Test 4: Verify API Key
	t.Run("VerifyAPIKey", func(t *testing.T) {
		// First generate an API key
		input := map[string]any{
			"user_id":     "user456",
			"permissions": []string{"admin"},
			"expiry_days": 30,
		}
		inputJSON, _ := json.Marshal(input)
		output, _ := genAPIKeyTool.InvokableRun(ctx, string(inputJSON))

		var genResult map[string]any
		json.Unmarshal([]byte(output), &genResult)
		apiKey := genResult["api_key"].(string)

		// Now verify it
		verifyInput := map[string]any{
			"api_key": apiKey,
		}
		verifyInputJSON, _ := json.Marshal(verifyInput)

		verifyOutput, err := verifyAPIKeyTool.InvokableRun(ctx, string(verifyInputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var verifyResult map[string]any
		if err := json.Unmarshal([]byte(verifyOutput), &verifyResult); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !verifyResult["success"].(bool) {
			t.Error("Expected success to be true")
		}

		if !verifyResult["valid"].(bool) {
			t.Error("Expected valid to be true")
		}
	})

	// Test 5: Check Permission
	t.Run("CheckPermission", func(t *testing.T) {
		input := map[string]any{
			"permissions": []string{"read", "write"},
			"required":    "read",
		}
		inputJSON, _ := json.Marshal(input)

		output, err := checkPermTool.InvokableRun(ctx, string(inputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !result["success"].(bool) {
			t.Error("Expected success to be true")
		}

		if !result["has_permission"].(bool) {
			t.Error("Expected has_permission to be true")
		}
	})
}

// TestMCPTools_ErrorHandling tests error handling in MCP tools
func TestMCPTools_ErrorHandling(t *testing.T) {
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
	defer module.Close()

	ctx := context.Background()
	tools, _ := module.GetTools(ctx)

	// Cast tools
	var genTokenTool *authGenerateTokenTool
	var verifyTokenTool *authVerifyTokenTool
	var verifyAPIKeyTool *authVerifyAPIKeyTool

	for _, tool := range tools {
		switch t := tool.(type) {
		case *authGenerateTokenTool:
			genTokenTool = t
		case *authVerifyTokenTool:
			verifyTokenTool = t
		case *authVerifyAPIKeyTool:
			verifyAPIKeyTool = t
		}
	}

	// Test invalid JSON input
	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := genTokenTool.InvokableRun(ctx, "invalid-json")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	// Test verify invalid token
	t.Run("VerifyInvalidToken", func(t *testing.T) {
		input := map[string]any{
			"token": "invalid-token",
		}
		inputJSON, _ := json.Marshal(input)

		output, err := verifyTokenTool.InvokableRun(ctx, string(inputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var result map[string]any
		json.Unmarshal([]byte(output), &result)

		if result["valid"].(bool) {
			t.Error("Expected valid to be false for invalid token")
		}
	})

	// Test verify invalid API key
	t.Run("VerifyInvalidAPIKey", func(t *testing.T) {
		input := map[string]any{
			"api_key": "invalid-key",
		}
		inputJSON, _ := json.Marshal(input)

		output, err := verifyAPIKeyTool.InvokableRun(ctx, string(inputJSON))
		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		var result map[string]any
		json.Unmarshal([]byte(output), &result)

		if result["valid"].(bool) {
			t.Error("Expected valid to be false for invalid API key")
		}
	})
}

// TestMCPTools_PermissionScenarios tests various permission scenarios
func TestMCPTools_PermissionScenarios(t *testing.T) {
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
	defer module.Close()

	ctx := context.Background()
	tools, _ := module.GetTools(ctx)

	// Cast to permission check tool
	var permTool *authCheckPermissionTool
	for _, tool := range tools {
		if t, ok := tool.(*authCheckPermissionTool); ok {
			permTool = t
			break
		}
	}

	tests := []struct {
		name        string
		permissions []string
		required    string
		expected    bool
	}{
		{
			name:        "Has exact permission",
			permissions: []string{"read", "write"},
			required:    "read",
			expected:    true,
		},
		{
			name:        "Missing permission",
			permissions: []string{"read"},
			required:    "write",
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
			input := map[string]any{
				"permissions": tt.permissions,
				"required":    tt.required,
			}
			inputJSON, _ := json.Marshal(input)

			output, err := permTool.InvokableRun(ctx, string(inputJSON))
			if err != nil {
				t.Fatalf("Failed to run tool: %v", err)
			}

			var result map[string]any
			json.Unmarshal([]byte(output), &result)

			if result["has_permission"].(bool) != tt.expected {
				t.Errorf("Expected has_permission to be %v, got %v", tt.expected, result["has_permission"])
			}
		})
	}
}
