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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package sandbox

import (
	"context"
	"fmt"
	"testing"
)

// TestAIOSandboxIntegration tests the integration of all AIO Sandbox modules
func TestAIOSandboxIntegration(t *testing.T) {
	// Create AIO Sandbox with all modules enabled
	config := DefaultConfig()
	config.Visual.Enable = true
	config.Auth.Enable = true
	config.Proxy.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Test getting all tools
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	// Verify that we got tools from all modules
	toolCount := len(tools)
	expectedMinTools := 15 // At least 5 tools per module, 3 modules
	if toolCount < expectedMinTools {
		t.Errorf("Expected at least %d tools, got %d", expectedMinTools, toolCount)
	}

	// Verify tool types
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			t.Errorf("Failed to get tool info: %v", err)
			continue
		}
		toolNames[info.Name] = true
	}

	// Check that we have at least one tool from each module
	modules := []string{
		"browser",
		"code_exec",
		"shell",
		"file",
		"visual",
		"auth",
		"proxy",
	}

	for _, module := range modules {
		found := false
		for name := range toolNames {
			if containsPrefix(name, module) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected at least one tool from %s module, but none found", module)
		}
	}

	t.Logf("Successfully retrieved %d tools from AIO Sandbox", toolCount)
}

// TestAIOSandboxWithCustomConfig2 tests AIO Sandbox with custom configuration (second variant)
func TestAIOSandboxWithCustomConfig2(t *testing.T) {
	// Create custom configuration
	config := DefaultConfig()
	config.Browser.Headless = false
	config.CodeExec.Timeout = 120000 // 2 minutes
	config.Shell.EnableBlacklist = true
	config.Shell.CommandBlacklist = []string{"rm", "rmdir", "shutdown"}

	// Create AIO Sandbox with custom configuration
	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox with custom config: %v", err)
	}
	defer sandbox.Close()

	// Verify that the sandbox was created successfully
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools from sandbox with custom config: %v", err)
	}

	if len(tools) == 0 {
		t.Error("Expected tools from sandbox with custom config, got none")
	}

	t.Logf("Successfully created AIOSandbox with custom config, retrieved %d tools", len(tools))
}

// TestAIOSandboxModuleAccess tests accessing individual modules from AIO Sandbox
func TestAIOSandboxModuleAccess(t *testing.T) {
	// Create AIO Sandbox
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Test accessing each module
	modules := []struct {
		name     string
		accessor func() interface{}
	}{
		{"Browser", func() interface{} { return sandbox.Browser() }},
		{"CodeExec", func() interface{} { return sandbox.CodeExec() }},
		{"Shell", func() interface{} { return sandbox.Shell() }},
		{"File", func() interface{} { return sandbox.File() }},
		{"Visual", func() interface{} { return sandbox.Visual() }},
		{"Auth", func() interface{} { return sandbox.Auth() }},
		{"Proxy", func() interface{} { return sandbox.Proxy() }},
	}

	for _, m := range modules {
		if m.accessor() == nil {
			t.Errorf("%s module should not be nil", m.name)
		} else {
			t.Logf("Successfully accessed %s module", m.name)
		}
	}
}

// TestToolInfoConsistency tests that all tools provide consistent info
func TestToolInfoConsistency(t *testing.T) {
	// Create AIO Sandbox
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	// Check that all tools provide valid info
	for i, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			t.Errorf("Tool %d Info() failed: %v", i, err)
			continue
		}

		if info.Name == "" {
			t.Errorf("Tool %d has empty name", i)
		}

		if info.Desc == "" {
			t.Errorf("Tool %d has empty description", i)
		}

		if info.ParamsOneOf == nil {
			t.Errorf("Tool %d has nil ParamsOneOf", i)
		}
	}
}

// Helper function to check if a string has a prefix (case-insensitive)
func containsPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	// Check if the string starts with the prefix
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	// Make sure it's a proper prefix (either end of string or followed by underscore)
	return len(s) == len(prefix) || s[len(prefix)] == '_'
}

// TestEndToEnd_AuthenticationWorkflow tests authentication workflow
func TestEndToEnd_AuthenticationWorkflow(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Step 1: Generate authentication token
	authResult, err := sandbox.Auth().GenerateToken("user123", []string{"code_exec"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if !authResult["success"].(bool) {
		t.Fatalf("Token generation failed: %v", authResult["error"])
	}

	token := authResult["token"].(string)
	t.Logf("Generated token: %s", token)

	// Step 2: Verify token
	verifyResult, err := sandbox.Auth().VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if !verifyResult["valid"].(bool) {
		t.Fatalf("Token verification failed")
	}

	// Step 3: Check permission
	permResult, err := sandbox.Auth().CheckPermission([]string{"code_exec"}, "code_exec")
	if err != nil {
		t.Fatalf("Failed to check permission: %v", err)
	}

	if !permResult["has_permission"].(bool) {
		t.Fatalf("Permission check failed")
	}

	t.Log("Authentication workflow completed successfully")
}

// TestEndToEnd_MultiModuleWorkflow tests a complex workflow involving multiple modules
func TestEndToEnd_MultiModuleWorkflow(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true
	config.Proxy.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Workflow: Generate API key -> Generate JWT token -> Verify both

	// Step 1: Generate API key for automation
	apiKeyResult, err := sandbox.Auth().GenerateAPIKey("automation-bot", []string{"*"}, 30)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if !apiKeyResult["success"].(bool) {
		t.Fatalf("API key generation failed")
	}

	apiKey := apiKeyResult["api_key"].(string)
	t.Logf("Generated API key: %s", apiKey)

	// Step 2: Verify API key
	verifyResult, err := sandbox.Auth().VerifyAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}

	if !verifyResult["valid"].(bool) {
		t.Fatalf("API key verification failed")
	}

	// Step 3: Generate JWT token
	tokenResult, err := sandbox.Auth().GenerateToken("user456", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if !tokenResult["success"].(bool) {
		t.Fatalf("Token generation failed")
	}

	token := tokenResult["token"].(string)
	t.Logf("Generated token: %s", token)

	// Step 4: Verify JWT token
	tokenVerifyResult, err := sandbox.Auth().VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if !tokenVerifyResult["valid"].(bool) {
		t.Fatalf("Token verification failed")
	}

	// Step 5: Get execution statistics
	authStats := sandbox.Auth().GetStats()
	codeStats := sandbox.CodeExec().GetStats()

	t.Logf("Auth stats: %+v", authStats)
	t.Logf("Code execution stats: %+v", codeStats)

	if authStats["total_requests"] < 4 {
		t.Errorf("Expected at least 4 auth requests, got %d", authStats["total_requests"])
	}
}

// TestMultiModule_ConcurrentOperations tests concurrent operations across modules
func TestMultiModule_ConcurrentOperations(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Run concurrent operations
	const numOperations = 10
	done := make(chan bool, numOperations*2)
	errors := make(chan error, numOperations*2)

	// Concurrent token generation
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user%d", id)
			result, err := sandbox.Auth().GenerateToken(userID, []string{"read"})
			if err != nil {
				errors <- fmt.Errorf("token generation error for %s: %v", userID, err)
			} else if !result["success"].(bool) {
				errors <- fmt.Errorf("token generation failed for %s", userID)
			}
			done <- true
		}(i)
	}

	// Concurrent API key generation
	for i := 0; i < numOperations; i++ {
		go func(id int) {
			userID := fmt.Sprintf("bot%d", id)
			result, err := sandbox.Auth().GenerateAPIKey(userID, []string{"write"}, 30)
			if err != nil {
				errors <- fmt.Errorf("API key generation error for %s: %v", userID, err)
			} else if !result["success"].(bool) {
				errors <- fmt.Errorf("API key generation failed for %s", userID)
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations*2; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	t.Logf("Successfully completed %d concurrent operations", numOperations*2)
}

// TestMultiModule_ErrorHandling tests error handling across modules
func TestMultiModule_ErrorHandling(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Test 1: Invalid token verification
	t.Run("InvalidTokenVerification", func(t *testing.T) {
		result, err := sandbox.Auth().VerifyToken("invalid-token")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result["valid"].(bool) {
			t.Error("Expected invalid token to fail verification")
		}
	})

	// Test 2: Invalid API key verification
	t.Run("InvalidAPIKeyVerification", func(t *testing.T) {
		result, err := sandbox.Auth().VerifyAPIKey("invalid-key")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result["valid"].(bool) {
			t.Error("Expected invalid API key to fail verification")
		}
	})

	// Test 3: Permission check with missing permission
	t.Run("MissingPermission", func(t *testing.T) {
		result, err := sandbox.Auth().CheckPermission([]string{"read"}, "write")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result["has_permission"].(bool) {
			t.Error("Expected permission check to fail for missing permission")
		}
	})
}

// TestMultiModule_ResourceCleanup tests resource cleanup across modules
func TestMultiModule_ResourceCleanup(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true

	// Create and close multiple sandboxes
	for i := 0; i < 5; i++ {
		sandbox, err := NewAIOSandbox(&config)
		if err != nil {
			t.Fatalf("Failed to create AIOSandbox iteration %d: %v", i, err)
		}

		// Perform some operations
		_, err = sandbox.Auth().GenerateToken("user", []string{"read"})
		if err != nil {
			t.Errorf("Failed to generate token in iteration %d: %v", i, err)
		}

		_, err = sandbox.Auth().GenerateAPIKey("bot", []string{"write"}, 30)
		if err != nil {
			t.Errorf("Failed to generate API key in iteration %d: %v", i, err)
		}

		// Close sandbox
		if err := sandbox.Close(); err != nil {
			t.Errorf("Failed to close sandbox in iteration %d: %v", i, err)
		}
	}

	t.Log("Successfully created and cleaned up 5 sandbox instances")
}

// TestMultiModule_StatsAggregation tests statistics aggregation across modules
func TestMultiModule_StatsAggregation(t *testing.T) {
	config := DefaultConfig()
	config.Auth.Enable = true

	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		t.Fatalf("Failed to create AIOSandbox: %v", err)
	}
	defer sandbox.Close()

	// Perform various operations
	for i := 0; i < 5; i++ {
		sandbox.Auth().GenerateToken("user", []string{"read"})
		sandbox.Auth().GenerateAPIKey("bot", []string{"write"}, 30)
	}

	// Get stats from both modules
	authStats := sandbox.Auth().GetStats()
	codeStats := sandbox.CodeExec().GetStats()

	t.Logf("Auth stats: %+v", authStats)
	t.Logf("Code execution stats: %+v", codeStats)

	// Verify stats
	if authStats["total_requests"] < 10 {
		t.Errorf("Expected at least 10 auth requests, got %d", authStats["total_requests"])
	}

	// Code stats should be 0 since we didn't execute any code
	t.Logf("Code execution total: %d", codeStats["total_executions"])
}
