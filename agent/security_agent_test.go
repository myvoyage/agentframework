// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package agent provides security agent implementation.
package agent

import (
	"context"
	"testing"
	"time"

	"AgentFramework/pkg/beads/security"
)

// TestSecurityAgent 测试安全代理
func TestSecurityAgent(t *testing.T) {
	ctx := context.Background()
	encryptionKey := "test-encryption-key-32-bytes-long!"
	agent := NewSecurityAgent(encryptionKey, 100)

	t.Run("Initialize", func(t *testing.T) {
		if err := agent.Initialize(ctx); err != nil {
			t.Errorf("Failed to initialize agent: %v", err)
		}
	})

	t.Run("ValidateCommand", func(t *testing.T) {
		// 安全命令
		err := agent.ValidateCommand(ctx, "ls", nil)
		if err != nil {
			t.Errorf("Expected no error for safe command: %v", err)
		}

		// 危险命令
		err = agent.ValidateCommand(ctx, "rm", nil)
		if err == nil {
			t.Error("Expected error for dangerous command")
		}
	})

	t.Run("CheckPermissions", func(t *testing.T) {
		// 先授予权限
		err := agent.GrantPermission(ctx, "user1", "resource1", "read")
		if err != nil {
			t.Errorf("Failed to grant permission: %v", err)
		}

		// 检查权限
		allowed, err := agent.CheckPermissions(ctx, "user1", "resource1", "read")
		if err != nil {
			t.Errorf("Failed to check permissions: %v", err)
		}
		if !allowed {
			t.Error("Expected permission to be granted")
		}

		// 检查未授权的权限
		allowed, err = agent.CheckPermissions(ctx, "user1", "resource1", "write")
		if err != nil {
			t.Errorf("Failed to check permissions: %v", err)
		}
		if allowed {
			t.Error("Expected permission to be denied")
		}
	})

	t.Run("EncryptDecryptData", func(t *testing.T) {
		originalData := []byte("sensitive information")

		// 加密
		encrypted, err := agent.EncryptData(ctx, originalData)
		if err != nil {
			t.Errorf("Failed to encrypt: %v", err)
		}

		// 验证加密后的数据不同
		if string(encrypted) == string(originalData) {
			t.Error("Encrypted data should be different from original")
		}

		// 解密
		decrypted, err := agent.DecryptData(ctx, encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt: %v", err)
		}

		// 验证解密后的数据相同
		if string(decrypted) != string(originalData) {
			t.Errorf("Decrypted data doesn't match. Got %s, want %s", string(decrypted), string(originalData))
		}
	})

	t.Run("EncryptDecryptBase64", func(t *testing.T) {
		originalData := []byte("test data")

		// 加密为Base64
		encrypted, err := agent.EncryptToBase64(ctx, originalData)
		if err != nil {
			t.Errorf("Failed to encrypt to base64: %v", err)
		}

		// 解密从Base64
		decrypted, err := agent.DecryptFromBase64(ctx, encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt from base64: %v", err)
		}

		// 验证
		if string(decrypted) != string(originalData) {
			t.Errorf("Decrypted data doesn't match. Got %s, want %s", string(decrypted), string(originalData))
		}
	})

	t.Run("AuditLog", func(t *testing.T) {
		event := security.AuditEvent{
			EventType: "test_event",
			UserID:    "user1",
			Resource:  "resource1",
			Action:    "read",
			Status:    "success",
			Timestamp: time.Now(),
		}

		if err := agent.AuditLog(ctx, event); err != nil {
			t.Errorf("Failed to log audit event: %v", err)
		}

		// 获取审计日志
		log, err := agent.GetAuditLog(ctx, 10)
		if err != nil {
			t.Errorf("Failed to get audit log: %v", err)
		}

		if len(log) == 0 {
			t.Error("Expected audit log to contain entries")
		}
	})

	t.Run("RevokePermission", func(t *testing.T) {
		// 先授予权限
		agent.GrantPermission(ctx, "user2", "resource2", "write")

		// 撤销权限
		err := agent.RevokePermission(ctx, "user2", "resource2", "write")
		if err != nil {
			t.Errorf("Failed to revoke permission: %v", err)
		}

		// 验证权限已撤销
		allowed, _ := agent.CheckPermissions(ctx, "user2", "resource2", "write")
		if allowed {
			t.Error("Expected permission to be revoked")
		}
	})

	t.Run("AddACLEntry", func(t *testing.T) {
		err := agent.AddACLEntry(ctx, "user3", "resource3", []string{"read", "write"}, "allow")
		if err != nil {
			t.Errorf("Failed to add ACL entry: %v", err)
		}

		// 检查ACL
		allowed, err := agent.CheckACL(ctx, "user3", "resource3", "read")
		if err != nil {
			t.Errorf("Failed to check ACL: %v", err)
		}
		if !allowed {
			t.Error("Expected ACL to allow access")
		}
	})

	t.Run("RemoveACLEntry", func(t *testing.T) {
		// 先添加条目
		agent.AddACLEntry(ctx, "user4", "resource4", []string{"read"}, "allow")

		// 移除条目
		err := agent.RemoveACLEntry(ctx, "user4", "resource4")
		if err != nil {
			t.Errorf("Failed to remove ACL entry: %v", err)
		}

		// 验证默认拒绝
		allowed, _ := agent.CheckACL(ctx, "user4", "resource4", "read")
		if allowed {
			t.Error("Expected ACL to deny after removal")
		}
	})

	t.Run("EnableDisable", func(t *testing.T) {
		// 禁用安全
		agent.Disable(ctx)

		// 危险命令应该通过（安全已禁用）
		err := agent.ValidateCommand(ctx, "rm", nil)
		if err != nil {
			t.Error("Expected no error when security is disabled")
		}

		// 启用安全
		agent.Enable(ctx)

		// 危险命令应该失败（安全已启用）
		err = agent.ValidateCommand(ctx, "rm", nil)
		if err == nil {
			t.Error("Expected error when security is enabled")
		}
	})

	t.Run("GetSecurityStatus", func(t *testing.T) {
		status := agent.GetSecurityStatus(ctx)

		if !status.Enabled {
			t.Error("Expected security to be enabled")
		}

		if !status.AuditEnabled {
			t.Error("Expected audit to be enabled")
		}
	})

	t.Run("ClearAuditLog", func(t *testing.T) {
		// 添加一些审计条目
		for i := 0; i < 5; i++ {
			agent.AuditLog(ctx, security.AuditEvent{
				EventType: "test_event",
				Timestamp: time.Now(),
			})
		}

		// 清除审计日志
		err := agent.ClearAuditLog(ctx)
		if err != nil {
			t.Errorf("Failed to clear audit log: %v", err)
		}

		// 验证已清除
		log, _ := agent.GetAuditLog(ctx, 10)
		if len(log) != 0 {
			t.Error("Expected audit log to be empty")
		}
	})

	t.Run("EnableDisableAudit", func(t *testing.T) {
		// 禁用审计
		agent.DisableAudit(ctx)

		if agent.IsAuditEnabled(ctx) {
			t.Error("Expected audit to be disabled")
		}

		// 启用审计
		agent.EnableAudit(ctx)

		if !agent.IsAuditEnabled(ctx) {
			t.Error("Expected audit to be enabled")
		}
	})

	t.Run("Close", func(t *testing.T) {
		if err := agent.Close(ctx); err != nil {
			t.Errorf("Failed to close agent: %v", err)
		}
	})
}

// BenchmarkSecurityAgent 性能测试
func BenchmarkSecurityAgent(b *testing.B) {
	ctx := context.Background()
	agent := NewSecurityAgent("test-encryption-key-32-bytes-long!", 1000)
	agent.Initialize(ctx)

	data := []byte("test data for encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := agent.EncryptData(ctx, data)
		agent.DecryptData(ctx, encrypted)
	}
}