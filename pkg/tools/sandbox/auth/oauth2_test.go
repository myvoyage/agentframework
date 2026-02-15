// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"testing"
	"time"
)

// TestOAuth2Handler_GenerateAuthorizationCode tests authorization code generation
func TestOAuth2Handler_GenerateAuthorizationCode(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read", "write"},
	}

	handler := NewOAuth2Handler(config)

	// Test successful authorization code generation
	authCode, err := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	if err != nil {
		t.Fatalf("Failed to generate authorization code: %v", err)
	}

	if authCode.Code == "" {
		t.Error("Authorization code should not be empty")
	}

	if authCode.UserID != "user123" {
		t.Errorf("UserID = %s, want user123", authCode.UserID)
	}

	if authCode.ClientID != "test-client-id" {
		t.Errorf("ClientID = %s, want test-client-id", authCode.ClientID)
	}

	// Test invalid client ID
	_, err = handler.GenerateAuthorizationCode(
		"invalid-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read"},
	)

	if err == nil {
		t.Error("Expected error for invalid client ID")
	}

	// Test invalid redirect URI
	_, err = handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://invalid.com/callback",
		[]string{"read"},
	)

	if err == nil {
		t.Error("Expected error for invalid redirect URI")
	}
}

// TestOAuth2Handler_ExchangeAuthorizationCode tests authorization code exchange
func TestOAuth2Handler_ExchangeAuthorizationCode(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read", "write"},
	}

	handler := NewOAuth2Handler(config)

	// Generate authorization code
	authCode, err := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	if err != nil {
		t.Fatalf("Failed to generate authorization code: %v", err)
	}

	// Exchange authorization code for tokens
	accessToken, refreshToken, err := handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	if err != nil {
		t.Fatalf("Failed to exchange authorization code: %v", err)
	}

	if accessToken.Token == "" {
		t.Error("Access token should not be empty")
	}

	if refreshToken.Token == "" {
		t.Error("Refresh token should not be empty")
	}

	if accessToken.UserID != "user123" {
		t.Errorf("UserID = %s, want user123", accessToken.UserID)
	}

	// Test invalid client credentials
	_, _, err = handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"wrong-secret",
		"https://example.com/callback",
	)

	if err == nil {
		t.Error("Expected error for invalid client credentials")
	}
}

// TestOAuth2Handler_ExchangeAuthorizationCode_Expired tests expired authorization code
func TestOAuth2Handler_ExchangeAuthorizationCode_Expired(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read"},
	}

	handler := NewOAuth2Handler(config)

	// Generate authorization code
	authCode, err := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read"},
	)

	if err != nil {
		t.Fatalf("Failed to generate authorization code: %v", err)
	}

	// Manually expire the authorization code
	handler.authCodes[authCode.Code].ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Try to exchange expired code
	_, _, err = handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	if err == nil {
		t.Error("Expected error for expired authorization code")
	}
}

// TestOAuth2Handler_VerifyAccessToken tests access token verification
func TestOAuth2Handler_VerifyAccessToken(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read", "write"},
	}

	handler := NewOAuth2Handler(config)

	// Generate and exchange authorization code
	authCode, _ := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	accessToken, _, _ := handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	// Verify access token
	verifiedToken, err := handler.VerifyAccessToken(accessToken.Token)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}

	if verifiedToken.UserID != "user123" {
		t.Errorf("UserID = %s, want user123", verifiedToken.UserID)
	}

	// Test invalid token
	_, err = handler.VerifyAccessToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid access token")
	}
}

// TestOAuth2Handler_RefreshAccessToken tests access token refresh
func TestOAuth2Handler_RefreshAccessToken(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read", "write"},
	}

	handler := NewOAuth2Handler(config)

	// Generate and exchange authorization code
	authCode, _ := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	oldAccessToken, oldRefreshToken, _ := handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	// Refresh access token
	newAccessToken, newRefreshToken, err := handler.RefreshAccessToken(
		oldRefreshToken.Token,
		"test-client-id",
		"test-client-secret",
	)

	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	if newAccessToken.Token == oldAccessToken.Token {
		t.Error("New access token should be different from old token")
	}

	if newRefreshToken.Token == oldRefreshToken.Token {
		t.Error("New refresh token should be different from old token")
	}

	if newAccessToken.UserID != "user123" {
		t.Errorf("UserID = %s, want user123", newAccessToken.UserID)
	}

	// Old refresh token should be invalid now
	_, _, err = handler.RefreshAccessToken(
		oldRefreshToken.Token,
		"test-client-id",
		"test-client-secret",
	)

	if err == nil {
		t.Error("Expected error for old refresh token")
	}
}

// TestOAuth2Handler_RevokeToken tests token revocation
func TestOAuth2Handler_RevokeToken(t *testing.T) {
	config := OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"read"},
	}

	handler := NewOAuth2Handler(config)

	// Generate and exchange authorization code
	authCode, _ := handler.GenerateAuthorizationCode(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read"},
	)

	accessToken, refreshToken, _ := handler.ExchangeAuthorizationCode(
		authCode.Code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	// Revoke access token
	err := handler.RevokeToken(accessToken.Token)
	if err != nil {
		t.Errorf("Failed to revoke access token: %v", err)
	}

	// Verify token is revoked
	_, err = handler.VerifyAccessToken(accessToken.Token)
	if err == nil {
		t.Error("Expected error for revoked access token")
	}

	// Revoke refresh token
	err = handler.RevokeToken(refreshToken.Token)
	if err != nil {
		t.Errorf("Failed to revoke refresh token: %v", err)
	}

	// Try to revoke non-existent token
	err = handler.RevokeToken("non-existent-token")
	if err == nil {
		t.Error("Expected error for non-existent token")
	}
}

// TestAuthModule_OAuth2Authorize tests OAuth2 authorization through AuthModule
func TestAuthModule_OAuth2Authorize(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
		OAuth2: OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "https://example.com/callback",
			Scopes:       []string{"read", "write"},
		},
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Test authorization
	result, err := module.oauth2Authorize(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	if err != nil {
		t.Fatalf("oauth2Authorize() error = %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	if result["code"] == nil || result["code"].(string) == "" {
		t.Error("Expected authorization code in result")
	}
}

// TestAuthModule_OAuth2ExchangeCode tests OAuth2 code exchange through AuthModule
func TestAuthModule_OAuth2ExchangeCode(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
		OAuth2: OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "https://example.com/callback",
			Scopes:       []string{"read", "write"},
		},
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Generate authorization code
	authResult, _ := module.oauth2Authorize(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	code := authResult["code"].(string)

	// Exchange code for tokens
	tokenResult, err := module.oauth2ExchangeCode(
		code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	if err != nil {
		t.Fatalf("oauth2ExchangeCode() error = %v", err)
	}

	if !tokenResult["success"].(bool) {
		t.Errorf("Expected success=true, got %v", tokenResult["success"])
	}

	if tokenResult["access_token"] == nil || tokenResult["access_token"].(string) == "" {
		t.Error("Expected access_token in result")
	}

	if tokenResult["refresh_token"] == nil || tokenResult["refresh_token"].(string) == "" {
		t.Error("Expected refresh_token in result")
	}
}

// TestAuthModule_OAuth2RefreshToken tests OAuth2 token refresh through AuthModule
func TestAuthModule_OAuth2RefreshToken(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
		OAuth2: OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "https://example.com/callback",
			Scopes:       []string{"read", "write"},
		},
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Generate authorization code and exchange for tokens
	authResult, _ := module.oauth2Authorize(
		"test-client-id",
		"user123",
		"https://example.com/callback",
		[]string{"read", "write"},
	)

	code := authResult["code"].(string)

	tokenResult, _ := module.oauth2ExchangeCode(
		code,
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
	)

	refreshToken := tokenResult["refresh_token"].(string)

	// Refresh access token
	refreshResult, err := module.oauth2RefreshToken(
		refreshToken,
		"test-client-id",
		"test-client-secret",
	)

	if err != nil {
		t.Fatalf("oauth2RefreshToken() error = %v", err)
	}

	if !refreshResult["success"].(bool) {
		t.Errorf("Expected success=true, got %v", refreshResult["success"])
	}

	if refreshResult["access_token"] == nil || refreshResult["access_token"].(string) == "" {
		t.Error("Expected new access_token in result")
	}

	if refreshResult["refresh_token"] == nil || refreshResult["refresh_token"].(string) == "" {
		t.Error("Expected new refresh_token in result")
	}
}

// TestAuthModule_OAuth2Tools tests OAuth2 MCP tools
func TestAuthModule_OAuth2Tools(t *testing.T) {
	config := AuthConfig{
		Enable:    true,
		JWTSecret: "test-secret",
		JWTExpiry: 3600,
		JWTIssuer: "test-issuer",
		OAuth2: OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "https://example.com/callback",
			Scopes:       []string{"read", "write"},
		},
	}

	module, err := NewAuthModule(config)
	if err != nil {
		t.Fatalf("Failed to create auth module: %v", err)
	}

	// Get tools
	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		t.Fatalf("GetTools() error = %v", err)
	}

	// Should have 7 tools now (5 original + 2 OAuth2)
	if len(tools) != 7 {
		t.Errorf("Expected 7 tools, got %d", len(tools))
	}

	// Check for OAuth2 tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		info, _ := tool.Info(ctx)
		toolNames[info.Name] = true
	}

	if !toolNames["auth_oauth2_authorize"] {
		t.Error("Expected auth_oauth2_authorize tool")
	}

	if !toolNames["auth_oauth2_token"] {
		t.Error("Expected auth_oauth2_token tool")
	}
}
