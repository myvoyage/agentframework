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

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/golang-jwt/jwt/v5"
)

// AuthModule 鉴权机制模块
type AuthModule struct {
	config      AuthConfig
	manager     *AuthManager
	jwtManager  *JWTManager
	apiKeyMgr   *APIKeyManager
	permChecker *PermissionChecker
	stats       *AuthStats
	mu          sync.RWMutex
}

// AuthConfig 鉴权配置
type AuthConfig struct {
	Enable    bool         `json:"enable"`
	JWTSecret string       `json:"jwt_secret"`
	JWTExpiry int          `json:"jwt_expiry"` // 秒
	JWTIssuer string       `json:"jwt_issuer"`
	OAuth2    OAuth2Config `json:"oauth2"`
}

// OAuth2Config OAuth2.0配置
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// AuthManager 鉴权管理器
type AuthManager struct {
	jwtManager    *JWTManager
	apiKeyMgr     *APIKeyManager
	permChecker   *PermissionChecker
	oauth2Handler *OAuth2Handler
}

// JWTManager JWT令牌管理器
type JWTManager struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
	mu        sync.RWMutex
}

// Claims JWT声明
type Claims struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// APIKeyManager API Key管理器
type APIKeyManager struct {
	keys map[string]*APIKey
	mu   sync.RWMutex
}

// APIKey API密钥
type APIKey struct {
	Key         string    `json:"key"`
	UserID      string    `json:"user_id"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PermissionChecker 权限检查器
type PermissionChecker struct {
	roles map[string][]string // role -> permissions
	mu    sync.RWMutex
}

// AuthStats 鉴权统计
type AuthStats struct {
	TotalRequests   int64
	SuccessCount    int64
	FailureCount    int64
	TokensGenerated int64
	TokensVerified  int64
	mu              sync.RWMutex
}

// NewAuthModule 创建鉴权模块实例
func NewAuthModule(config AuthConfig) (*AuthModule, error) {
	// 验证配置
	if config.JWTSecret == "" {
		config.JWTSecret = generateRandomSecret()
	}
	if config.JWTExpiry <= 0 {
		config.JWTExpiry = 3600 // 默认1小时
	}
	if config.JWTIssuer == "" {
		config.JWTIssuer = "aio-sandbox"
	}

	// 创建 JWT 管理器
	jwtManager := &JWTManager{
		secretKey: []byte(config.JWTSecret),
		issuer:    config.JWTIssuer,
		expiry:    time.Duration(config.JWTExpiry) * time.Second,
	}

	// 创建 API Key 管理器
	apiKeyMgr := &APIKeyManager{
		keys: make(map[string]*APIKey),
	}

	// 创建权限检查器
	permChecker := &PermissionChecker{
		roles: make(map[string][]string),
	}

	// 初始化默认角色
	permChecker.roles["admin"] = []string{"*"}
	permChecker.roles["user"] = []string{"read", "write"}
	permChecker.roles["guest"] = []string{"read"}

	manager := &AuthManager{
		jwtManager:    jwtManager,
		apiKeyMgr:     apiKeyMgr,
		permChecker:   permChecker,
		oauth2Handler: NewOAuth2Handler(config.OAuth2),
	}

	stats := &AuthStats{}

	return &AuthModule{
		config:      config,
		manager:     manager,
		jwtManager:  jwtManager,
		apiKeyMgr:   apiKeyMgr,
		permChecker: permChecker,
		stats:       stats,
	}, nil
}

// GetTools 返回鉴权模块的 MCP 工具列表
func (m *AuthModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	// 如果鉴权模块未启用，返回空列表
	if !m.config.Enable {
		return []tool.BaseTool{}, nil
	}

	tools := []tool.BaseTool{
		// JWT令牌生成工具
		&authGenerateTokenTool{module: m},
		// JWT令牌验证工具
		&authVerifyTokenTool{module: m},
		// API Key生成工具
		&authGenerateAPIKeyTool{module: m},
		// API Key验证工具
		&authVerifyAPIKeyTool{module: m},
		// 权限检查工具
		&authCheckPermissionTool{module: m},
		// OAuth2授权工具
		&authOAuth2AuthorizeTool{module: m},
		// OAuth2令牌交换工具
		&authOAuth2TokenTool{module: m},
	}

	return tools, nil
}

// ============================================================================
// MCP Tools Implementation
// ============================================================================

// JWT令牌生成工具
type authGenerateTokenTool struct {
	module *AuthModule
}

func (t *authGenerateTokenTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_generate_token",
		Desc: "Generate a JWT token for a user",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "User ID",
				Required: true,
			},
			"permissions": {
				Type: "array",
				Desc: "List of permissions",
			},
		}),
	}, nil
}

func (t *authGenerateTokenTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		UserID      string   `json:"user_id"`
		Permissions []string `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.generateToken(args.UserID, args.Permissions)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// JWT令牌验证工具
type authVerifyTokenTool struct {
	module *AuthModule
}

func (t *authVerifyTokenTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_verify_token",
		Desc: "Verify a JWT token",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"token": {
				Type:     "string",
				Desc:     "JWT token to verify",
				Required: true,
			},
		}),
	}, nil
}

func (t *authVerifyTokenTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.verifyToken(args.Token)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// API Key生成工具
type authGenerateAPIKeyTool struct {
	module *AuthModule
}

func (t *authGenerateAPIKeyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_generate_api_key",
		Desc: "Generate an API key for a user",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "User ID",
				Required: true,
			},
			"permissions": {
				Type: "array",
				Desc: "List of permissions",
			},
			"expiry_days": {
				Type: "integer",
				Desc: "Expiry in days (default: 365)",
			},
		}),
	}, nil
}

func (t *authGenerateAPIKeyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		UserID      string   `json:"user_id"`
		Permissions []string `json:"permissions"`
		ExpiryDays  int      `json:"expiry_days"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.generateAPIKey(args.UserID, args.Permissions, args.ExpiryDays)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// API Key验证工具
type authVerifyAPIKeyTool struct {
	module *AuthModule
}

func (t *authVerifyAPIKeyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_verify_api_key",
		Desc: "Verify an API key",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"api_key": {
				Type:     "string",
				Desc:     "API key to verify",
				Required: true,
			},
		}),
	}, nil
}

func (t *authVerifyAPIKeyTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		APIKey string `json:"api_key"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.verifyAPIKey(args.APIKey)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 权限检查工具
type authCheckPermissionTool struct {
	module *AuthModule
}

func (t *authCheckPermissionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_check_permission",
		Desc: "Check if user has required permission",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"permissions": {
				Type:     "array",
				Desc:     "User's permissions",
				Required: true,
			},
			"required": {
				Type:     "string",
				Desc:     "Required permission",
				Required: true,
			},
		}),
	}, nil
}

func (t *authCheckPermissionTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Permissions []string `json:"permissions"`
		Required    string   `json:"required"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.checkPermission(args.Permissions, args.Required)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭鉴权模块，释放资源
func (m *AuthModule) Close() error {
	// 清理资源
	m.apiKeyMgr.mu.Lock()
	m.apiKeyMgr.keys = make(map[string]*APIKey)
	m.apiKeyMgr.mu.Unlock()

	return nil
}

// GetStats 获取鉴权统计信息
func (m *AuthModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_requests":   m.stats.TotalRequests,
		"success_count":    m.stats.SuccessCount,
		"failure_count":    m.stats.FailureCount,
		"tokens_generated": m.stats.TokensGenerated,
		"tokens_verified":  m.stats.TokensVerified,
	}
}

// ============================================================================
// JWTManager Implementation
// ============================================================================

// Generate 生成JWT令牌
func (m *JWTManager) Generate(userID string, permissions []string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Verify 验证JWT令牌
func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ============================================================================
// APIKeyManager Implementation
// ============================================================================

// Generate 生成API Key
func (m *APIKeyManager) Generate(userID string, permissions []string, expiryDays int) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成随机密钥
	key := generateRandomKey()

	// 设置过期时间
	if expiryDays <= 0 {
		expiryDays = 365 // 默认1年
	}

	apiKey := &APIKey{
		Key:         key,
		UserID:      userID,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, expiryDays),
	}

	m.keys[key] = apiKey
	return apiKey, nil
}

// Verify 验证API Key
func (m *APIKeyManager) Verify(key string) (*APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	apiKey, exists := m.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	// 检查是否过期
	if time.Now().After(apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	return apiKey, nil
}

// Revoke 撤销API Key
func (m *APIKeyManager) Revoke(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.keys[key]; !exists {
		return fmt.Errorf("API key not found")
	}

	delete(m.keys, key)
	return nil
}

// List 列出所有API Key
func (m *APIKeyManager) List(userID string) []*APIKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []*APIKey
	for _, key := range m.keys {
		if userID == "" || key.UserID == userID {
			keys = append(keys, key)
		}
	}

	return keys
}

// ============================================================================
// PermissionChecker Implementation
// ============================================================================

// HasPermission 检查是否有指定权限
func (c *PermissionChecker) HasPermission(userPermissions []string, required string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, perm := range userPermissions {
		// 通配符权限
		if perm == "*" {
			return true
		}
		// 精确匹配
		if perm == required {
			return true
		}
	}

	return false
}

// HasAnyPermission 检查是否有任意一个权限
func (c *PermissionChecker) HasAnyPermission(userPermissions []string, required []string) bool {
	for _, req := range required {
		if c.HasPermission(userPermissions, req) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查是否有所有权限
func (c *PermissionChecker) HasAllPermissions(userPermissions []string, required []string) bool {
	for _, req := range required {
		if !c.HasPermission(userPermissions, req) {
			return false
		}
	}
	return true
}

// AddRole 添加角色
func (c *PermissionChecker) AddRole(role string, permissions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.roles[role] = permissions
}

// GetRolePermissions 获取角色权限
func (c *PermissionChecker) GetRolePermissions(role string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.roles[role]
}

// ============================================================================
// Auth Module Core Functions
// ============================================================================

// GenerateToken 生成JWT令牌（导出方法）
func (m *AuthModule) GenerateToken(userID string, permissions []string) (map[string]any, error) {
	return m.generateToken(userID, permissions)
}

// VerifyToken 验证JWT令牌（导出方法）
func (m *AuthModule) VerifyToken(tokenString string) (map[string]any, error) {
	return m.verifyToken(tokenString)
}

// GenerateAPIKey 生成API Key（导出方法）
func (m *AuthModule) GenerateAPIKey(userID string, permissions []string, expiryDays int) (map[string]any, error) {
	return m.generateAPIKey(userID, permissions, expiryDays)
}

// VerifyAPIKey 验证API Key（导出方法）
func (m *AuthModule) VerifyAPIKey(key string) (map[string]any, error) {
	return m.verifyAPIKey(key)
}

// CheckPermission 检查权限（导出方法）
func (m *AuthModule) CheckPermission(permissions []string, required string) (map[string]any, error) {
	return m.checkPermission(permissions, required)
}

// generateToken 生成JWT令牌
func (m *AuthModule) generateToken(userID string, permissions []string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 生成令牌
	token, err := m.jwtManager.Generate(userID, permissions)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"user_id": userID,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TokensGenerated++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":     true,
		"token":       token,
		"user_id":     userID,
		"permissions": permissions,
		"expires_in":  int(m.jwtManager.expiry.Seconds()),
		"message":     "Token generated successfully",
	}, nil
}

// verifyToken 验证JWT令牌
func (m *AuthModule) verifyToken(tokenString string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.TokensVerified++
	m.stats.mu.Unlock()

	// 验证令牌
	claims, err := m.jwtManager.Verify(tokenString)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":     true,
		"valid":       true,
		"user_id":     claims.UserID,
		"permissions": claims.Permissions,
		"issuer":      claims.Issuer,
		"expires_at":  claims.ExpiresAt.Time,
		"issued_at":   claims.IssuedAt.Time,
		"message":     "Token verified successfully",
	}, nil
}

// generateAPIKey 生成API Key
func (m *AuthModule) generateAPIKey(userID string, permissions []string, expiryDays int) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 生成API Key
	apiKey, err := m.apiKeyMgr.Generate(userID, permissions, expiryDays)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"user_id": userID,
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":     true,
		"api_key":     apiKey.Key,
		"user_id":     apiKey.UserID,
		"permissions": apiKey.Permissions,
		"created_at":  apiKey.CreatedAt,
		"expires_at":  apiKey.ExpiresAt,
		"message":     "API key generated successfully",
	}, nil
}

// verifyAPIKey 验证API Key
func (m *AuthModule) verifyAPIKey(key string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 验证API Key
	apiKey, err := m.apiKeyMgr.Verify(key)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":     true,
		"valid":       true,
		"user_id":     apiKey.UserID,
		"permissions": apiKey.Permissions,
		"created_at":  apiKey.CreatedAt,
		"expires_at":  apiKey.ExpiresAt,
		"message":     "API key verified successfully",
	}, nil
}

// checkPermission 检查权限
func (m *AuthModule) checkPermission(permissions []string, required string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 检查权限
	hasPermission := m.permChecker.HasPermission(permissions, required)

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":        true,
		"has_permission": hasPermission,
		"permissions":    permissions,
		"required":       required,
		"message":        "Permission check completed",
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateRandomSecret 生成随机密钥
func generateRandomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// generateRandomKey 生成随机API Key
func generateRandomKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "sk_" + base64.URLEncoding.EncodeToString(b)
}

// ============================================================================
// OAuth2 Implementation
// ============================================================================

// OAuth2Handler OAuth2处理器
type OAuth2Handler struct {
	config        OAuth2Config
	authCodes     map[string]*AuthorizationCode
	accessTokens  map[string]*AccessToken
	refreshTokens map[string]*RefreshToken
	mu            sync.RWMutex
}

// AuthorizationCode 授权码
type AuthorizationCode struct {
	Code        string    `json:"code"`
	ClientID    string    `json:"client_id"`
	UserID      string    `json:"user_id"`
	RedirectURI string    `json:"redirect_uri"`
	Scopes      []string  `json:"scopes"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AccessToken 访问令牌
type AccessToken struct {
	Token     string    `json:"token"`
	ClientID  string    `json:"client_id"`
	UserID    string    `json:"user_id"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RefreshToken 刷新令牌
type RefreshToken struct {
	Token     string    `json:"token"`
	ClientID  string    `json:"client_id"`
	UserID    string    `json:"user_id"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewOAuth2Handler 创建OAuth2处理器
func NewOAuth2Handler(config OAuth2Config) *OAuth2Handler {
	return &OAuth2Handler{
		config:        config,
		authCodes:     make(map[string]*AuthorizationCode),
		accessTokens:  make(map[string]*AccessToken),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

// GenerateAuthorizationCode 生成授权码
func (h *OAuth2Handler) GenerateAuthorizationCode(clientID, userID, redirectURI string, scopes []string) (*AuthorizationCode, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 验证客户端ID
	if clientID != h.config.ClientID {
		return nil, fmt.Errorf("invalid client_id")
	}

	// 验证重定向URI
	if redirectURI != h.config.RedirectURL {
		return nil, fmt.Errorf("invalid redirect_uri")
	}

	// 生成授权码
	code := generateRandomCode()
	authCode := &AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 授权码10分钟有效
	}

	h.authCodes[code] = authCode
	return authCode, nil
}

// ExchangeAuthorizationCode 交换授权码获取访问令牌
func (h *OAuth2Handler) ExchangeAuthorizationCode(code, clientID, clientSecret, redirectURI string) (*AccessToken, *RefreshToken, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 验证客户端凭证
	if clientID != h.config.ClientID || clientSecret != h.config.ClientSecret {
		return nil, nil, fmt.Errorf("invalid client credentials")
	}

	// 查找授权码
	authCode, exists := h.authCodes[code]
	if !exists {
		return nil, nil, fmt.Errorf("invalid authorization code")
	}

	// 验证授权码是否过期
	if time.Now().After(authCode.ExpiresAt) {
		delete(h.authCodes, code)
		return nil, nil, fmt.Errorf("authorization code expired")
	}

	// 验证客户端ID和重定向URI
	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		return nil, nil, fmt.Errorf("authorization code mismatch")
	}

	// 生成访问令牌
	accessToken := &AccessToken{
		Token:     generateRandomToken(),
		ClientID:  clientID,
		UserID:    authCode.UserID,
		Scopes:    authCode.Scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // 访问令牌1小时有效
	}

	// 生成刷新令牌
	refreshToken := &RefreshToken{
		Token:     generateRandomToken(),
		ClientID:  clientID,
		UserID:    authCode.UserID,
		Scopes:    authCode.Scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 30), // 刷新令牌30天有效
	}

	// 保存令牌
	h.accessTokens[accessToken.Token] = accessToken
	h.refreshTokens[refreshToken.Token] = refreshToken

	// 删除已使用的授权码
	delete(h.authCodes, code)

	return accessToken, refreshToken, nil
}

// VerifyAccessToken 验证访问令牌
func (h *OAuth2Handler) VerifyAccessToken(token string) (*AccessToken, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	accessToken, exists := h.accessTokens[token]
	if !exists {
		return nil, fmt.Errorf("invalid access token")
	}

	// 检查是否过期
	if time.Now().After(accessToken.ExpiresAt) {
		return nil, fmt.Errorf("access token expired")
	}

	return accessToken, nil
}

// RefreshAccessToken 刷新访问令牌
func (h *OAuth2Handler) RefreshAccessToken(refreshTokenStr, clientID, clientSecret string) (*AccessToken, *RefreshToken, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 验证客户端凭证
	if clientID != h.config.ClientID || clientSecret != h.config.ClientSecret {
		return nil, nil, fmt.Errorf("invalid client credentials")
	}

	// 查找刷新令牌
	refreshToken, exists := h.refreshTokens[refreshTokenStr]
	if !exists {
		return nil, nil, fmt.Errorf("invalid refresh token")
	}

	// 验证刷新令牌是否过期
	if time.Now().After(refreshToken.ExpiresAt) {
		delete(h.refreshTokens, refreshTokenStr)
		return nil, nil, fmt.Errorf("refresh token expired")
	}

	// 验证客户端ID
	if refreshToken.ClientID != clientID {
		return nil, nil, fmt.Errorf("refresh token mismatch")
	}

	// 生成新的访问令牌
	newAccessToken := &AccessToken{
		Token:     generateRandomToken(),
		ClientID:  clientID,
		UserID:    refreshToken.UserID,
		Scopes:    refreshToken.Scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// 生成新的刷新令牌
	newRefreshToken := &RefreshToken{
		Token:     generateRandomToken(),
		ClientID:  clientID,
		UserID:    refreshToken.UserID,
		Scopes:    refreshToken.Scopes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	}

	// 保存新令牌
	h.accessTokens[newAccessToken.Token] = newAccessToken
	h.refreshTokens[newRefreshToken.Token] = newRefreshToken

	// 删除旧的刷新令牌
	delete(h.refreshTokens, refreshTokenStr)

	return newAccessToken, newRefreshToken, nil
}

// RevokeToken 撤销令牌
func (h *OAuth2Handler) RevokeToken(token string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 尝试删除访问令牌
	if _, exists := h.accessTokens[token]; exists {
		delete(h.accessTokens, token)
		return nil
	}

	// 尝试删除刷新令牌
	if _, exists := h.refreshTokens[token]; exists {
		delete(h.refreshTokens, token)
		return nil
	}

	return fmt.Errorf("token not found")
}

// generateRandomCode 生成随机授权码
func generateRandomCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// generateRandomToken 生成随机令牌
func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ============================================================================
// OAuth2 MCP Tools
// ============================================================================

// OAuth2授权工具
type authOAuth2AuthorizeTool struct {
	module *AuthModule
}

func (t *authOAuth2AuthorizeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_oauth2_authorize",
		Desc: "Generate OAuth2 authorization code",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"client_id": {
				Type:     "string",
				Desc:     "Client ID",
				Required: true,
			},
			"user_id": {
				Type:     "string",
				Desc:     "User ID",
				Required: true,
			},
			"redirect_uri": {
				Type:     "string",
				Desc:     "Redirect URI",
				Required: true,
			},
			"scopes": {
				Type: "array",
				Desc: "Requested scopes",
			},
		}),
	}, nil
}

func (t *authOAuth2AuthorizeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		ClientID    string   `json:"client_id"`
		UserID      string   `json:"user_id"`
		RedirectURI string   `json:"redirect_uri"`
		Scopes      []string `json:"scopes"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.oauth2Authorize(args.ClientID, args.UserID, args.RedirectURI, args.Scopes)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// OAuth2令牌交换工具
type authOAuth2TokenTool struct {
	module *AuthModule
}

func (t *authOAuth2TokenTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "auth_oauth2_token",
		Desc: "Exchange authorization code for access token or refresh access token",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"grant_type": {
				Type:     "string",
				Desc:     "Grant type: authorization_code or refresh_token",
				Required: true,
			},
			"code": {
				Type: "string",
				Desc: "Authorization code (for authorization_code grant)",
			},
			"refresh_token": {
				Type: "string",
				Desc: "Refresh token (for refresh_token grant)",
			},
			"client_id": {
				Type:     "string",
				Desc:     "Client ID",
				Required: true,
			},
			"client_secret": {
				Type:     "string",
				Desc:     "Client secret",
				Required: true,
			},
			"redirect_uri": {
				Type: "string",
				Desc: "Redirect URI (for authorization_code grant)",
			},
		}),
	}, nil
}

func (t *authOAuth2TokenTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		RefreshToken string `json:"refresh_token"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var result map[string]any
	var err error

	switch args.GrantType {
	case "authorization_code":
		result, err = t.module.oauth2ExchangeCode(args.Code, args.ClientID, args.ClientSecret, args.RedirectURI)
	case "refresh_token":
		result, err = t.module.oauth2RefreshToken(args.RefreshToken, args.ClientID, args.ClientSecret)
	default:
		return "", fmt.Errorf("unsupported grant_type: %s", args.GrantType)
	}

	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// ============================================================================
// Auth Module OAuth2 Functions
// ============================================================================

// oauth2Authorize 生成OAuth2授权码
func (m *AuthModule) oauth2Authorize(clientID, userID, redirectURI string, scopes []string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 创建OAuth2处理器（如果不存在）
	if m.manager.oauth2Handler == nil {
		m.manager.oauth2Handler = NewOAuth2Handler(m.config.OAuth2)
	}

	// 生成授权码
	authCode, err := m.manager.oauth2Handler.GenerateAuthorizationCode(clientID, userID, redirectURI, scopes)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":      true,
		"code":         authCode.Code,
		"redirect_uri": redirectURI,
		"expires_in":   int(time.Until(authCode.ExpiresAt).Seconds()),
		"message":      "Authorization code generated successfully",
	}, nil
}

// oauth2ExchangeCode 交换授权码获取访问令牌
func (m *AuthModule) oauth2ExchangeCode(code, clientID, clientSecret, redirectURI string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 创建OAuth2处理器（如果不存在）
	if m.manager.oauth2Handler == nil {
		m.manager.oauth2Handler = NewOAuth2Handler(m.config.OAuth2)
	}

	// 交换授权码
	accessToken, refreshToken, err := m.manager.oauth2Handler.ExchangeAuthorizationCode(code, clientID, clientSecret, redirectURI)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":       true,
		"access_token":  accessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(accessToken.ExpiresAt).Seconds()),
		"refresh_token": refreshToken.Token,
		"scopes":        accessToken.Scopes,
		"message":       "Access token generated successfully",
	}, nil
}

// oauth2RefreshToken 刷新访问令牌
func (m *AuthModule) oauth2RefreshToken(refreshToken, clientID, clientSecret string) (map[string]any, error) {
	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 创建OAuth2处理器（如果不存在）
	if m.manager.oauth2Handler == nil {
		m.manager.oauth2Handler = NewOAuth2Handler(m.config.OAuth2)
	}

	// 刷新令牌
	newAccessToken, newRefreshToken, err := m.manager.oauth2Handler.RefreshAccessToken(refreshToken, clientID, clientSecret)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.mu.Unlock()

	return map[string]any{
		"success":       true,
		"access_token":  newAccessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(newAccessToken.ExpiresAt).Seconds()),
		"refresh_token": newRefreshToken.Token,
		"scopes":        newAccessToken.Scopes,
		"message":       "Access token refreshed successfully",
	}, nil
}
