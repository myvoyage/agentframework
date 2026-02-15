# Auth Module - Example Usage

This document provides practical examples of using the Auth Module in various scenarios.

## Table of Contents

1. [Basic Setup](#basic-setup)
2. [JWT Token Workflows](#jwt-token-workflows)
3. [API Key Workflows](#api-key-workflows)
4. [Permission Management](#permission-management)
5. [Integration Examples](#integration-examples)
6. [Error Handling](#error-handling)

## Basic Setup

### Creating an Auth Module

```go
package main

import (
    "fmt"
    "log"
    "AgentFramework/agent/aiosandbox/auth"
)

func main() {
    // Create configuration
    config := auth.AuthConfig{
        Enable:    true,
        JWTSecret: "your-super-secret-key-min-32-bytes",
        JWTExpiry: 3600, // 1 hour
        JWTIssuer: "my-app",
    }

    // Initialize module
    module, err := auth.NewAuthModule(config)
    if err != nil {
        log.Fatal("Failed to create auth module:", err)
    }
    defer module.Close()

    fmt.Println("Auth module initialized successfully")
}
```

### Using Default Configuration

```go
// Minimal configuration with defaults
config := auth.AuthConfig{
    Enable: true,
    // JWTSecret will be auto-generated
    // JWTExpiry defaults to 3600 seconds
    // JWTIssuer defaults to "aio-sandbox"
}

module, _ := auth.NewAuthModule(config)
defer module.Close()
```

## JWT Token Workflows

### Example 1: User Login Flow

```go
func loginUser(module *auth.AuthModule, username, password string) (string, error) {
    // Validate credentials (your logic here)
    if !validateCredentials(username, password) {
        return "", fmt.Errorf("invalid credentials")
    }

    // Determine user permissions based on role
    permissions := getUserPermissions(username)

    // Generate JWT token
    result, err := module.GenerateToken(username, permissions)
    if err != nil {
        return "", err
    }

    if !result["success"].(bool) {
        return "", fmt.Errorf("failed to generate token")
    }

    token := result["token"].(string)
    expiresIn := result["expires_in"].(int)

    fmt.Printf("Token generated for %s, expires in %d seconds\n", username, expiresIn)
    return token, nil
}

func validateCredentials(username, password string) bool {
    // Your authentication logic
    return true
}

func getUserPermissions(username string) []string {
    // Your permission logic
    return []string{"read", "write"}
}
```

### Example 2: Protected API Endpoint

```go
func protectedHandler(module *auth.AuthModule, token string) error {
    // Verify the token
    result, err := module.VerifyToken(token)
    if err != nil {
        return fmt.Errorf("verification error: %w", err)
    }

    if !result["valid"].(bool) {
        return fmt.Errorf("invalid token: %s", result["error"])
    }

    // Extract user information
    userID := result["user_id"].(string)
    permissions := result["permissions"].([]interface{})

    fmt.Printf("Authenticated user: %s\n", userID)
    fmt.Printf("Permissions: %v\n", permissions)

    // Proceed with protected operation
    return nil
}
```

### Example 3: Token Refresh Pattern

```go
func refreshTokenIfNeeded(module *auth.AuthModule, token string) (string, error) {
    // Verify current token
    result, err := module.VerifyToken(token)
    if err != nil {
        return "", err
    }

    if !result["valid"].(bool) {
        return "", fmt.Errorf("token is invalid")
    }

    // Check if token is expiring soon (e.g., within 5 minutes)
    expiresAt := result["expires_at"].(time.Time)
    if time.Until(expiresAt) < 5*time.Minute {
        // Generate new token
        userID := result["user_id"].(string)
        permissions := result["permissions"].([]string)

        newResult, err := module.GenerateToken(userID, permissions)
        if err != nil {
            return "", err
        }

        return newResult["token"].(string), nil
    }

    // Token is still valid
    return token, nil
}
```

## API Key Workflows

### Example 4: Generate API Key for Service Account

```go
func createServiceAccount(module *auth.AuthModule, serviceName string) (string, error) {
    // Define service permissions
    permissions := []string{"read", "write", "execute"}

    // Generate API key with 90-day expiration
    result, err := module.GenerateAPIKey(serviceName, permissions, 90)
    if err != nil {
        return "", err
    }

    if !result["success"].(bool) {
        return "", fmt.Errorf("failed to generate API key")
    }

    apiKey := result["api_key"].(string)
    expiresAt := result["expires_at"].(time.Time)

    fmt.Printf("API Key created for %s\n", serviceName)
    fmt.Printf("Key: %s\n", apiKey)
    fmt.Printf("Expires: %s\n", expiresAt.Format(time.RFC3339))

    return apiKey, nil
}
```

### Example 5: API Key Authentication Middleware

```go
func apiKeyMiddleware(module *auth.AuthModule, apiKey string) error {
    // Verify API key
    result, err := module.VerifyAPIKey(apiKey)
    if err != nil {
        return fmt.Errorf("verification error: %w", err)
    }

    if !result["valid"].(bool) {
        return fmt.Errorf("invalid API key: %s", result["error"])
    }

    // Extract service information
    userID := result["user_id"].(string)
    permissions := result["permissions"].([]interface{})
    expiresAt := result["expires_at"].(time.Time)

    // Check if key is expiring soon
    if time.Until(expiresAt) < 7*24*time.Hour {
        fmt.Printf("Warning: API key for %s expires in less than 7 days\n", userID)
    }

    fmt.Printf("Authenticated service: %s\n", userID)
    fmt.Printf("Permissions: %v\n", permissions)

    return nil
}
```

### Example 6: List and Manage API Keys

```go
func manageAPIKeys(module *auth.AuthModule, userID string) {
    // List all keys for a user
    keys := module.apiKeyMgr.List(userID)

    fmt.Printf("API Keys for %s:\n", userID)
    for i, key := range keys {
        fmt.Printf("%d. Key: %s...\n", i+1, key.Key[:20])
        fmt.Printf("   Created: %s\n", key.CreatedAt.Format(time.RFC3339))
        fmt.Printf("   Expires: %s\n", key.ExpiresAt.Format(time.RFC3339))
        fmt.Printf("   Permissions: %v\n", key.Permissions)

        // Revoke expired keys
        if time.Now().After(key.ExpiresAt) {
            module.apiKeyMgr.Revoke(key.Key)
            fmt.Printf("   Status: REVOKED (expired)\n")
        } else {
            fmt.Printf("   Status: ACTIVE\n")
        }
    }
}
```

## Permission Management

### Example 7: Role-Based Access Control

```go
func setupRoles(module *auth.AuthModule) {
    // Add custom roles
    module.permChecker.AddRole("editor", []string{"read", "write", "edit"})
    module.permChecker.AddRole("moderator", []string{"read", "write", "delete", "moderate"})
    module.permChecker.AddRole("viewer", []string{"read"})

    fmt.Println("Custom roles configured")
}

func assignUserRole(module *auth.AuthModule, userID, role string) []string {
    // Get permissions for role
    permissions := module.permChecker.GetRolePermissions(role)

    if len(permissions) == 0 {
        fmt.Printf("Role '%s' not found, using default 'user' role\n", role)
        permissions = module.permChecker.GetRolePermissions("user")
    }

    fmt.Printf("Assigned role '%s' to user %s\n", role, userID)
    fmt.Printf("Permissions: %v\n", permissions)

    return permissions
}
```

### Example 8: Permission Checking in Operations

```go
func performOperation(module *auth.AuthModule, userPermissions []string, operation string) error {
    // Define required permissions for operations
    requiredPerms := map[string]string{
        "read_data":   "read",
        "write_data":  "write",
        "delete_data": "delete",
        "admin_panel": "admin",
    }

    required, exists := requiredPerms[operation]
    if !exists {
        return fmt.Errorf("unknown operation: %s", operation)
    }

    // Check permission
    result, err := module.CheckPermission(userPermissions, required)
    if err != nil {
        return err
    }

    if !result["has_permission"].(bool) {
        return fmt.Errorf("permission denied: %s required", required)
    }

    fmt.Printf("Permission granted for operation: %s\n", operation)
    return nil
}
```

### Example 9: Complex Permission Scenarios

```go
func checkComplexPermissions(module *auth.AuthModule) {
    // Scenario 1: Admin with wildcard permission
    adminPerms := []string{"*"}
    result, _ := module.CheckPermission(adminPerms, "anything")
    fmt.Printf("Admin can do anything: %v\n", result["has_permission"])

    // Scenario 2: User with specific permissions
    userPerms := []string{"read", "write"}
    result, _ = module.CheckPermission(userPerms, "delete")
    fmt.Printf("User can delete: %v\n", result["has_permission"])

    // Scenario 3: Check multiple permissions
    hasRead := module.permChecker.HasPermission(userPerms, "read")
    hasWrite := module.permChecker.HasPermission(userPerms, "write")
    hasDelete := module.permChecker.HasPermission(userPerms, "delete")

    fmt.Printf("User permissions - Read: %v, Write: %v, Delete: %v\n",
        hasRead, hasWrite, hasDelete)

    // Scenario 4: Check if user has ANY of required permissions
    hasAny := module.permChecker.HasAnyPermission(
        userPerms,
        []string{"delete", "write", "admin"},
    )
    fmt.Printf("User has any of [delete, write, admin]: %v\n", hasAny)

    // Scenario 5: Check if user has ALL required permissions
    hasAll := module.permChecker.HasAllPermissions(
        userPerms,
        []string{"read", "write"},
    )
    fmt.Printf("User has all of [read, write]: %v\n", hasAll)
}
```

## Integration Examples

### Example 10: HTTP Server Integration

```go
package main

import (
    "encoding/json"
    "net/http"
    "strings"
    "AgentFramework/agent/aiosandbox/auth"
)

var authModule *auth.AuthModule

func main() {
    // Initialize auth module
    config := auth.AuthConfig{
        Enable:    true,
        JWTSecret: "your-secret-key",
        JWTExpiry: 3600,
    }
    authModule, _ = auth.NewAuthModule(config)
    defer authModule.Close()

    // Setup routes
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/protected", authMiddleware(protectedHandler))

    http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    json.NewDecoder(r.Body).Decode(&req)

    // Validate credentials (simplified)
    if req.Username == "" || req.Password == "" {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate token
    result, _ := authModule.GenerateToken(req.Username, []string{"read", "write"})

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Verify token
        result, _ := authModule.VerifyToken(token)
        if !result["valid"].(bool) {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add user info to context and proceed
        next(w, r)
    }
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Protected resource accessed successfully"))
}
```

### Example 11: CLI Tool Integration

```go
package main

import (
    "fmt"
    "os"
    "AgentFramework/agent/aiosandbox/auth"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: auth-cli <command> [args]")
        os.Exit(1)
    }

    config := auth.AuthConfig{Enable: true}
    module, _ := auth.NewAuthModule(config)
    defer module.Close()

    command := os.Args[1]

    switch command {
    case "generate-token":
        if len(os.Args) < 3 {
            fmt.Println("Usage: auth-cli generate-token <user_id>")
            os.Exit(1)
        }
        userID := os.Args[2]
        result, _ := module.GenerateToken(userID, []string{"read", "write"})
        fmt.Printf("Token: %s\n", result["token"])

    case "verify-token":
        if len(os.Args) < 3 {
            fmt.Println("Usage: auth-cli verify-token <token>")
            os.Exit(1)
        }
        token := os.Args[2]
        result, _ := module.VerifyToken(token)
        if result["valid"].(bool) {
            fmt.Printf("Valid token for user: %s\n", result["user_id"])
        } else {
            fmt.Println("Invalid token")
        }

    case "generate-key":
        if len(os.Args) < 3 {
            fmt.Println("Usage: auth-cli generate-key <user_id>")
            os.Exit(1)
        }
        userID := os.Args[2]
        result, _ := module.GenerateAPIKey(userID, []string{"admin"}, 365)
        fmt.Printf("API Key: %s\n", result["api_key"])

    default:
        fmt.Printf("Unknown command: %s\n", command)
        os.Exit(1)
    }
}
```

## Error Handling

### Example 12: Comprehensive Error Handling

```go
func robustTokenVerification(module *auth.AuthModule, token string) {
    result, err := module.VerifyToken(token)

    // Handle system errors
    if err != nil {
        fmt.Printf("System error: %v\n", err)
        return
    }

    // Check operation success
    if !result["success"].(bool) {
        fmt.Printf("Operation failed\n")
        if errorMsg, ok := result["error"].(string); ok {
            fmt.Printf("Error: %s\n", errorMsg)
        }
        return
    }

    // Check token validity
    if !result["valid"].(bool) {
        fmt.Println("Token is invalid")
        if errorMsg, ok := result["error"].(string); ok {
            fmt.Printf("Reason: %s\n", errorMsg)
        }
        return
    }

    // Token is valid, extract information
    userID := result["user_id"].(string)
    permissions := result["permissions"].([]interface{})
    expiresAt := result["expires_at"].(time.Time)

    fmt.Printf("Valid token for user: %s\n", userID)
    fmt.Printf("Permissions: %v\n", permissions)
    fmt.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))
}
```

### Example 13: Graceful Degradation

```go
func authenticateWithFallback(module *auth.AuthModule, token, apiKey string) (string, []string, error) {
    // Try JWT token first
    if token != "" {
        result, err := module.VerifyToken(token)
        if err == nil && result["valid"].(bool) {
            userID := result["user_id"].(string)
            perms := result["permissions"].([]string)
            return userID, perms, nil
        }
        fmt.Println("JWT verification failed, trying API key...")
    }

    // Fallback to API key
    if apiKey != "" {
        result, err := module.VerifyAPIKey(apiKey)
        if err == nil && result["valid"].(bool) {
            userID := result["user_id"].(string)
            perms := result["permissions"].([]string)
            return userID, perms, nil
        }
        fmt.Println("API key verification failed")
    }

    return "", nil, fmt.Errorf("authentication failed")
}
```

## Statistics and Monitoring

### Example 14: Monitor Auth Operations

```go
func monitorAuthStats(module *auth.AuthModule) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := module.GetStats()

        totalRequests := stats["total_requests"]
        successCount := stats["success_count"]
        failureCount := stats["failure_count"]

        successRate := float64(successCount) / float64(totalRequests) * 100

        fmt.Printf("\n=== Auth Statistics ===\n")
        fmt.Printf("Total Requests: %d\n", totalRequests)
        fmt.Printf("Success: %d (%.2f%%)\n", successCount, successRate)
        fmt.Printf("Failures: %d\n", failureCount)
        fmt.Printf("Tokens Generated: %d\n", stats["tokens_generated"])
        fmt.Printf("Tokens Verified: %d\n", stats["tokens_verified"])
        fmt.Println("=====================")
    }
}
```

## Best Practices

1. **Always use HTTPS** in production for token transmission
2. **Store secrets securely** using environment variables or secret management systems
3. **Implement token refresh** to maintain user sessions
4. **Set appropriate expiration times** based on security requirements
5. **Log authentication events** for security auditing
6. **Rotate API keys regularly** to minimize security risks
7. **Use least privilege principle** when assigning permissions
8. **Validate tokens on every request** to protected resources
9. **Handle errors gracefully** and provide meaningful error messages
10. **Monitor authentication metrics** to detect anomalies

## See Also

- [README](./README.md) - Module overview and features
- [API Reference](./API_REFERENCE.md) - Detailed API documentation
- [Security Best Practices](./SECURITY.md) - Security guidelines
