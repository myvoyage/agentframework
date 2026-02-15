# Auth Module - Authentication and Authorization

The Auth Module provides comprehensive authentication and authorization capabilities for the AIO Sandbox, including JWT token management, API key management, and permission checking.

## Features

- **JWT Token Management**: Generate and verify JSON Web Tokens (JWT) for secure authentication
- **API Key Management**: Create, verify, and revoke API keys with expiration support
- **Permission System**: Role-based access control (RBAC) with flexible permission checking
- **Statistics Tracking**: Monitor authentication operations and success rates
- **Thread-Safe**: All operations are safe for concurrent use

## Architecture

```
AuthModule
├── JWTManager          # JWT token generation and verification
├── APIKeyManager       # API key lifecycle management
├── PermissionChecker   # Permission validation
└── AuthStats          # Operation statistics
```

## Configuration

```go
config := auth.AuthConfig{
    Enable:    true,                    // Enable/disable auth module
    JWTSecret: "your-secret-key",       // Secret key for JWT signing
    JWTExpiry: 3600,                    // Token expiry in seconds (default: 3600)
    JWTIssuer: "aio-sandbox",           // Token issuer (default: "aio-sandbox")
}

module, err := auth.NewAuthModule(config)
if err != nil {
    log.Fatal(err)
}
defer module.Close()
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Enable` | bool | false | Enable/disable the auth module |
| `JWTSecret` | string | auto-generated | Secret key for JWT signing (32 bytes recommended) |
| `JWTExpiry` | int | 3600 | Token expiration time in seconds |
| `JWTIssuer` | string | "aio-sandbox" | JWT token issuer identifier |

## MCP Tools

The Auth Module provides 5 MCP tools for authentication and authorization:

### 1. auth_generate_token

Generate a JWT token for a user with specified permissions.

**Parameters:**
- `user_id` (string, required): User identifier
- `permissions` (array, optional): List of permission strings

**Example:**
```json
{
  "user_id": "user123",
  "permissions": ["read", "write"]
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "user123",
  "permissions": ["read", "write"],
  "expires_in": 3600,
  "message": "Token generated successfully"
}
```

### 2. auth_verify_token

Verify a JWT token and extract its claims.

**Parameters:**
- `token` (string, required): JWT token to verify

**Example:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "user_id": "user123",
  "permissions": ["read", "write"],
  "issuer": "aio-sandbox",
  "expires_at": "2025-01-30T12:00:00Z",
  "issued_at": "2025-01-29T12:00:00Z",
  "message": "Token verified successfully"
}
```

### 3. auth_generate_api_key

Generate an API key for a user with specified permissions and expiration.

**Parameters:**
- `user_id` (string, required): User identifier
- `permissions` (array, optional): List of permission strings
- `expiry_days` (integer, optional): Expiration in days (default: 365)

**Example:**
```json
{
  "user_id": "user456",
  "permissions": ["admin"],
  "expiry_days": 30
}
```

**Response:**
```json
{
  "success": true,
  "api_key": "sk_AbCdEf123456...",
  "user_id": "user456",
  "permissions": ["admin"],
  "created_at": "2025-01-29T12:00:00Z",
  "expires_at": "2025-02-28T12:00:00Z",
  "message": "API key generated successfully"
}
```

### 4. auth_verify_api_key

Verify an API key and retrieve its information.

**Parameters:**
- `api_key` (string, required): API key to verify

**Example:**
```json
{
  "api_key": "sk_AbCdEf123456..."
}
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "user_id": "user456",
  "permissions": ["admin"],
  "created_at": "2025-01-29T12:00:00Z",
  "expires_at": "2025-02-28T12:00:00Z",
  "message": "API key verified successfully"
}
```

### 5. auth_check_permission

Check if a user has a required permission.

**Parameters:**
- `permissions` (array, required): User's permissions
- `required` (string, required): Required permission to check

**Example:**
```json
{
  "permissions": ["read", "write"],
  "required": "read"
}
```

**Response:**
```json
{
  "success": true,
  "has_permission": true,
  "permissions": ["read", "write"],
  "required": "read",
  "message": "Permission check completed"
}
```

## Usage Examples

### Basic JWT Authentication

```go
// Create auth module
config := auth.AuthConfig{
    Enable:    true,
    JWTSecret: "my-secret-key",
    JWTExpiry: 3600,
}
module, _ := auth.NewAuthModule(config)
defer module.Close()

// Generate token
tokenResult, _ := module.GenerateToken("user123", []string{"read", "write"})
token := tokenResult["token"].(string)

// Verify token
verifyResult, _ := module.VerifyToken(token)
if verifyResult["valid"].(bool) {
    userID := verifyResult["user_id"].(string)
    permissions := verifyResult["permissions"].([]string)
    fmt.Printf("User %s has permissions: %v\n", userID, permissions)
}
```

### API Key Management

```go
// Generate API key
apiKeyResult, _ := module.GenerateAPIKey("user456", []string{"admin"}, 30)
apiKey := apiKeyResult["api_key"].(string)

// Verify API key
verifyResult, _ := module.VerifyAPIKey(apiKey)
if verifyResult["valid"].(bool) {
    fmt.Println("API key is valid")
}
```

### Permission Checking

```go
// Check single permission
result, _ := module.CheckPermission([]string{"read", "write"}, "read")
hasPermission := result["has_permission"].(bool)

// Using wildcard permission
result, _ = module.CheckPermission([]string{"*"}, "anything")
// Returns true for any permission
```

## Permission System

### Built-in Roles

The module comes with three predefined roles:

| Role | Permissions | Description |
|------|-------------|-------------|
| `admin` | `*` (wildcard) | Full access to all operations |
| `user` | `read`, `write` | Standard user permissions |
| `guest` | `read` | Read-only access |

### Custom Roles

You can add custom roles with specific permissions:

```go
module.permChecker.AddRole("editor", []string{"read", "write", "edit"})
permissions := module.permChecker.GetRolePermissions("editor")
```

### Permission Checking Methods

```go
// Check single permission
hasPermission := module.permChecker.HasPermission(
    []string{"read", "write"}, 
    "read",
)

// Check if user has ANY of the required permissions
hasAny := module.permChecker.HasAnyPermission(
    []string{"read", "write"},
    []string{"delete", "write", "admin"},
)

// Check if user has ALL required permissions
hasAll := module.permChecker.HasAllPermissions(
    []string{"read", "write", "delete"},
    []string{"read", "write"},
)
```

## Statistics

Monitor authentication operations:

```go
stats := module.GetStats()
fmt.Printf("Total Requests: %d\n", stats["total_requests"])
fmt.Printf("Success Count: %d\n", stats["success_count"])
fmt.Printf("Failure Count: %d\n", stats["failure_count"])
fmt.Printf("Tokens Generated: %d\n", stats["tokens_generated"])
fmt.Printf("Tokens Verified: %d\n", stats["tokens_verified"])
```

## Security Considerations

### JWT Tokens

1. **Secret Key**: Use a strong, randomly generated secret key (minimum 32 bytes)
2. **Expiration**: Set appropriate expiration times based on your security requirements
3. **HTTPS Only**: Always transmit tokens over HTTPS in production
4. **Storage**: Store tokens securely (e.g., HTTP-only cookies, secure storage)

### API Keys

1. **Prefix**: All API keys are prefixed with `sk_` for easy identification
2. **Expiration**: Set reasonable expiration periods (default: 365 days)
3. **Revocation**: Revoke compromised keys immediately
4. **Rotation**: Implement regular key rotation policies

### Permissions

1. **Least Privilege**: Grant minimum required permissions
2. **Wildcard**: Use wildcard (`*`) permission sparingly
3. **Validation**: Always validate permissions before sensitive operations
4. **Audit**: Log permission checks for security auditing

## Error Handling

All operations return structured responses with success indicators:

```go
result, err := module.VerifyToken(token)
if err != nil {
    // Handle system error
    log.Fatal(err)
}

if !result["success"].(bool) {
    // Handle operation failure
    errorMsg := result["error"].(string)
    fmt.Println("Verification failed:", errorMsg)
}
```

## Thread Safety

All components are thread-safe and can be used concurrently:

- JWT generation and verification
- API key management operations
- Permission checking
- Statistics updates

## Testing

The module includes comprehensive tests:

```bash
# Run all tests
go test ./agent/aiosandbox/auth

# Run with coverage
go test -cover ./agent/aiosandbox/auth

# Run integration tests
go test -v ./agent/aiosandbox/auth -run Integration
```

Current test coverage: **86.7%**

## Performance

- JWT generation: < 1ms
- JWT verification: < 1ms
- API key generation: < 1ms
- API key verification: < 0.1ms (in-memory lookup)
- Permission checking: < 0.01ms

## Limitations

1. **In-Memory Storage**: API keys are stored in memory and will be lost on restart
2. **No Persistence**: For production use, implement persistent storage
3. **No OAuth2**: OAuth2.0 support is planned but not yet implemented
4. **No Session Management**: Session management is planned but not yet implemented

## Future Enhancements

- [ ] Persistent storage for API keys
- [ ] OAuth2.0 authorization flows
- [ ] Session management
- [ ] Token refresh mechanism
- [ ] Token revocation list
- [ ] Multi-factor authentication (MFA)
- [ ] Rate limiting per user/key

## License

AGPL-3.0-or-later

## See Also

- [Example Usage](./example_usage.md)
- [API Reference](./API_REFERENCE.md)
- [Security Best Practices](./SECURITY.md)
