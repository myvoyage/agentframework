// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"

	"AgentFramework/pkg/tools/sandbox/auth"
)

func main() {
	fmt.Println("=== Auth Module Demo ===\n")

	// 创建 Auth Module 配置
	config := auth.AuthConfig{
		Enable:    true,
		JWTSecret: "my-secret-key-for-demo",
		JWTExpiry: 3600, // 1 hour
		JWTIssuer: "auth-demo",
	}

	// 创建 Auth Module 实例
	authModule, err := auth.NewAuthModule(config)
	if err != nil {
		log.Fatalf("Failed to create auth module: %v", err)
	}
	defer authModule.Close()

	fmt.Println("✓ Auth Module created successfully\n")

	// 获取 MCP 工具
	tools, err := authModule.GetTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	fmt.Printf("✓ Available tools: %d\n", len(tools))
	for i, tool := range tools {
		info, _ := tool.Info(context.Background())
		fmt.Printf("  %d. %s - %s\n", i+1, info.Name, info.Desc)
	}
	fmt.Println()

	// 演示 1: 生成 JWT 令牌
	fmt.Println("--- Demo 1: Generate JWT Token ---")
	tokenResult, err := authModule.GenerateToken("user123", []string{"read", "write", "delete"})
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", tokenResult["success"])
		token := tokenResult["token"].(string)
		fmt.Printf("Token: %s...\n", token[:50])
		fmt.Printf("User ID: %s\n", tokenResult["user_id"])
		fmt.Printf("Permissions: %v\n", tokenResult["permissions"])
		fmt.Printf("Expires in: %v seconds\n\n", tokenResult["expires_in"])

		// 演示 2: 验证 JWT 令牌
		fmt.Println("--- Demo 2: Verify JWT Token ---")
		verifyResult, err := authModule.VerifyToken(token)
		if err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Success: %v\n", verifyResult["success"])
			fmt.Printf("Valid: %v\n", verifyResult["valid"])
			fmt.Printf("User ID: %s\n", verifyResult["user_id"])
			fmt.Printf("Permissions: %v\n\n", verifyResult["permissions"])
		}
	}

	// 演示 3: 生成 API Key
	fmt.Println("--- Demo 3: Generate API Key ---")
	apiKeyResult, err := authModule.GenerateAPIKey("user456", []string{"read", "write"}, 365)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", apiKeyResult["success"])
		apiKey := apiKeyResult["api_key"].(string)
		fmt.Printf("API Key: %s...\n", apiKey[:30])
		fmt.Printf("User ID: %s\n", apiKeyResult["user_id"])
		fmt.Printf("Permissions: %v\n", apiKeyResult["permissions"])
		fmt.Printf("Created at: %v\n", apiKeyResult["created_at"])
		fmt.Printf("Expires at: %v\n\n", apiKeyResult["expires_at"])

		// 演示 4: 验证 API Key
		fmt.Println("--- Demo 4: Verify API Key ---")
		verifyAPIKeyResult, err := authModule.VerifyAPIKey(apiKey)
		if err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Success: %v\n", verifyAPIKeyResult["success"])
			fmt.Printf("Valid: %v\n", verifyAPIKeyResult["valid"])
			fmt.Printf("User ID: %s\n", verifyAPIKeyResult["user_id"])
			fmt.Printf("Permissions: %v\n\n", verifyAPIKeyResult["permissions"])
		}
	}

	// 演示 5: 检查权限
	fmt.Println("--- Demo 5: Check Permission ---")
	permResult, err := authModule.CheckPermission([]string{"read", "write"}, "write")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", permResult["success"])
		fmt.Printf("Has Permission: %v\n", permResult["has_permission"])
		fmt.Printf("User Permissions: %v\n", permResult["permissions"])
		fmt.Printf("Required: %s\n\n", permResult["required"])
	}

	// 演示 6: 检查不存在的权限
	fmt.Println("--- Demo 6: Check Missing Permission ---")
	permResult2, err := authModule.CheckPermission([]string{"read", "write"}, "admin")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", permResult2["success"])
		fmt.Printf("Has Permission: %v\n", permResult2["has_permission"])
		fmt.Printf("User Permissions: %v\n", permResult2["permissions"])
		fmt.Printf("Required: %s\n\n", permResult2["required"])
	}

	// 演示 7: 使用通配符权限
	fmt.Println("--- Demo 7: Wildcard Permission ---")
	permResult3, err := authModule.CheckPermission([]string{"*"}, "anything")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", permResult3["success"])
		fmt.Printf("Has Permission: %v\n", permResult3["has_permission"])
		fmt.Printf("User Permissions: %v\n", permResult3["permissions"])
		fmt.Printf("Required: %s\n\n", permResult3["required"])
	}

	// 显示统计信息
	fmt.Println("--- Statistics ---")
	stats := authModule.GetStats()
	fmt.Printf("Total Requests: %d\n", stats["total_requests"])
	fmt.Printf("Success Count: %d\n", stats["success_count"])
	fmt.Printf("Failure Count: %d\n", stats["failure_count"])
	fmt.Printf("Tokens Generated: %d\n", stats["tokens_generated"])
	fmt.Printf("Tokens Verified: %d\n", stats["tokens_verified"])

	fmt.Println("\n=== Demo Complete ===")
}
