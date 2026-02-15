# Auth Module - OAuth2 工具文档

## 概述

Auth Module 现在支持完整的 OAuth2.0 授权码流程，包括授权码生成、令牌交换、令牌刷新和令牌撤销。

## OAuth2 工具

### 1. auth_oauth2_authorize

**功能**: 生成 OAuth2 授权码

**输入参数**:
```json
{
  "client_id": "your-client-id",
  "user_id": "user123",
  "redirect_uri": "https://example.com/callback",
  "scopes": ["read", "write"]
}
```

**输出**:
```json
{
  "success": true,
  "code": "auth_code_xxx",
  "expires_in": 600
}
```

**使用场景**: 用户授权后，生成授权码返回给客户端

---

### 2. auth_oauth2_token

**功能**: 交换授权码获取访问令牌，或刷新访问令牌

**授权码交换**:
```json
{
  "grant_type": "authorization_code",
  "code": "auth_code_xxx",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "redirect_uri": "https://example.com/callback"
}
```

**令牌刷新**:
```json
{
  "grant_type": "refresh_token",
  "refresh_token": "refresh_token_xxx",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret"
}
```

**输出**:
```json
{
  "success": true,
  "access_token": "access_token_xxx",
  "refresh_token": "refresh_token_xxx",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

## OAuth2 流程示例

### 完整授权流程

```go
// 1. 用户授权，生成授权码
authResult, err := authModule.oauth2Authorize(
    "client-id",
    "user123",
    "https://example.com/callback",
    []string{"read", "write"},
)
code := authResult["code"].(string)

// 2. 客户端使用授权码交换令牌
tokenResult, err := authModule.oauth2ExchangeCode(
    code,
    "client-id",
    "client-secret",
    "https://example.com/callback",
)
accessToken := tokenResult["access_token"].(string)
refreshToken := tokenResult["refresh_token"].(string)

// 3. 使用访问令牌访问资源
verifiedToken, err := authModule.oauth2Handler.VerifyAccessToken(accessToken)

// 4. 访问令牌过期后，使用刷新令牌获取新令牌
newTokenResult, err := authModule.oauth2RefreshToken(
    refreshToken,
    "client-id",
    "client-secret",
)
```

---

## 安全特性

1. **授权码过期**: 授权码 10 分钟后自动过期
2. **访问令牌过期**: 访问令牌 1 小时后自动过期
3. **刷新令牌过期**: 刷新令牌 30 天后自动过期
4. **客户端验证**: 验证 client_id 和 client_secret
5. **重定向 URI 验证**: 防止授权码劫持
6. **令牌撤销**: 支持主动撤销令牌
7. **一次性使用**: 授权码和刷新令牌使用后自动失效

---

## 配置

在 `AuthConfig` 中配置 OAuth2:

```go
config := AuthConfig{
    Enable:    true,
    JWTSecret: "your-secret",
    JWTExpiry: 3600,
    JWTIssuer: "your-issuer",
    OAuth2: OAuth2Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "https://example.com/callback",
        Scopes:       []string{"read", "write", "admin"},
    },
}
```

---

## 测试覆盖

OAuth2 实现包含完整的测试套件:

- ✅ 授权码生成测试
- ✅ 授权码交换测试
- ✅ 授权码过期测试
- ✅ 访问令牌验证测试
- ✅ 令牌刷新测试
- ✅ 令牌撤销测试
- ✅ 客户端验证测试
- ✅ 重定向 URI 验证测试
- ✅ MCP 工具集成测试

测试覆盖率: **80.1%**

---

## 实现细节

### OAuth2Handler 结构

```go
type OAuth2Handler struct {
    config        OAuth2Config
    authCodes     map[string]*AuthorizationCode
    accessTokens  map[string]*AccessToken
    refreshTokens map[string]*RefreshToken
    mu            sync.RWMutex
}
```

### 核心方法

- `GenerateAuthorizationCode()`: 生成授权码
- `ExchangeAuthorizationCode()`: 交换授权码为令牌
- `VerifyAccessToken()`: 验证访问令牌
- `RefreshAccessToken()`: 刷新访问令牌
- `RevokeToken()`: 撤销令牌

---

## 注意事项

1. **生产环境**: 当前实现使用内存存储，生产环境建议使用 Redis 或数据库
2. **HTTPS**: OAuth2 必须在 HTTPS 环境下使用
3. **密钥管理**: 妥善保管 client_secret
4. **令牌存储**: 客户端应安全存储 refresh_token
5. **作用域验证**: 根据业务需求实现作用域权限检查

---

## 参考资料

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)
- [OAuth 2.0 Security Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)
