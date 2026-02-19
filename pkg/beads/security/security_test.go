// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package security provides security policies and access control.
package security

import (
	"context"
	"testing"
	"time"
)

// TestSecurityManager 测试安全管理器
func TestSecurityManager(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes-long!")
	manager := NewSecurityManager(key, 100, true)

	ctx := context.Background()

	// 测试命令验证
	t.Run("ValidateCommand", func(t *testing.T) {
		// 安全命令
		if err := manager.ValidateCommand(ctx, "ls", nil); err != nil {
			t.Errorf("Expected no error for safe command: %v", err)
		}

		// 危险命令
		if err := manager.ValidateCommand(ctx, "rm", nil); err == nil {
			t.Error("Expected error for dangerous command")
		}
	})

	// 测试权限管理
	t.Run("GrantPermission", func(t *testing.T) {
		if err := manager.GrantPermission(ctx, "user1", "resource1", "read"); err != nil {
			t.Errorf("Failed to grant permission: %v", err)
		}

		// 检查权限
		allowed, err := manager.CheckPermissions(ctx, "user1", "resource1", "read")
		if err != nil {
			t.Errorf("Failed to check permissions: %v", err)
		}
		if !allowed {
			t.Error("Expected permission to be granted")
		}

		// 检查未授权的权限
		allowed, err = manager.CheckPermissions(ctx, "user1", "resource1", "write")
		if err != nil {
			t.Errorf("Failed to check permissions: %v", err)
		}
		if allowed {
			t.Error("Expected permission to be denied")
		}
	})

	// 测试加密解密
	t.Run("EncryptDecrypt", func(t *testing.T) {
		originalData := []byte("sensitive data")

		// 加密
		encrypted, err := manager.EncryptData(ctx, originalData)
		if err != nil {
			t.Fatalf("Failed to encrypt: %v", err)
		}

		// 验证加密后的数据与原始数据不同
		if string(encrypted) == string(originalData) {
			t.Error("Encrypted data should be different from original")
		}

		// 解密
		decrypted, err := manager.DecryptData(ctx, encrypted)
		if err != nil {
			t.Fatalf("Failed to decrypt: %v", err)
		}

		// 验证解密后的数据与原始数据相同
		if string(decrypted) != string(originalData) {
			t.Errorf("Decrypted data doesn't match original. Got %s, want %s", string(decrypted), string(originalData))
		}
	})

	// 测试Base64加密解密
	t.Run("EncryptDecryptBase64", func(t *testing.T) {
		originalData := []byte("sensitive data")

		// 加密为Base64
		encrypted, err := manager.EncryptToBase64(ctx, originalData)
		if err != nil {
			t.Fatalf("Failed to encrypt to base64: %v", err)
		}

		// 解密从Base64
		decrypted, err := manager.DecryptFromBase64(ctx, encrypted)
		if err != nil {
			t.Fatalf("Failed to decrypt from base64: %v", err)
		}

		// 验证
		if string(decrypted) != string(originalData) {
			t.Errorf("Decrypted data doesn't match original. Got %s, want %s", string(decrypted), string(originalData))
		}
	})

	// 测试审计日志
	t.Run("AuditLog", func(t *testing.T) {
		event := AuditEvent{
			EventType: "test_event",
			UserID:    "user1",
			Resource:  "resource1",
			Action:    "read",
			Status:    "success",
			Timestamp: time.Now(),
		}

		if err := manager.AuditLog(ctx, event); err != nil {
			t.Errorf("Failed to log audit event: %v", err)
		}

		// 获取审计日志
		log, err := manager.GetAuditLog(ctx, 10)
		if err != nil {
			t.Errorf("Failed to get audit log: %v", err)
		}

		if len(log) == 0 {
			t.Error("Expected audit log to contain entries")
		}

		// 验证最后的事件是我们刚添加的
		lastEvent := log[len(log)-1]
		if lastEvent.EventType != "test_event" {
			t.Errorf("Expected last event type to be 'test_event', got %s", lastEvent.EventType)
		}
	})

	// 测试权限撤销
	t.Run("RevokePermission", func(t *testing.T) {
		// 先授予权限
		if err := manager.GrantPermission(ctx, "user2", "resource2", "write"); err != nil {
			t.Errorf("Failed to grant permission: %v", err)
		}

		// 验证权限存在
		allowed, _ := manager.CheckPermissions(ctx, "user2", "resource2", "write")
		if !allowed {
			t.Error("Expected permission to exist before revocation")
		}

		// 撤销权限
		if err := manager.RevokePermission(ctx, "user2", "resource2", "write"); err != nil {
			t.Errorf("Failed to revoke permission: %v", err)
		}

		// 验证权限已撤销
		allowed, _ = manager.CheckPermissions(ctx, "user2", "resource2", "write")
		if allowed {
			t.Error("Expected permission to be revoked")
		}
	})
}

// TestAccessControlList 测试访问控制列表
func TestAccessControlList(t *testing.T) {
	ctx := context.Background()
	acl := NewAccessControlList()

	// 添加允许规则
	t.Run("AddAllowEntry", func(t *testing.T) {
		entry := ACLEntry{
			Principal: "user1",
			Resource:  "resource1",
			Actions:   []string{"read", "write"},
			Effect:    "allow",
		}

		if err := acl.AddEntry(ctx, entry); err != nil {
			t.Errorf("Failed to add ACL entry: %v", err)
		}

		// 检查允许的权限
		allowed, err := acl.Check(ctx, "user1", "resource1", "read")
		if err != nil {
			t.Errorf("Failed to check ACL: %v", err)
		}
		if !allowed {
			t.Error("Expected action to be allowed")
		}

		// 检查未授权的权限
		allowed, err = acl.Check(ctx, "user1", "resource1", "delete")
		if err != nil {
			t.Errorf("Failed to check ACL: %v", err)
		}
		if allowed {
			t.Error("Expected action to be denied")
		}
	})

	// 添加拒绝规则
	t.Run("AddDenyEntry", func(t *testing.T) {
		entry := ACLEntry{
			Principal: "user2",
			Resource:  "resource2",
			Actions:   []string{"*"},
			Effect:    "deny",
		}

		if err := acl.AddEntry(ctx, entry); err != nil {
			t.Errorf("Failed to add ACL entry: %v", err)
		}

		// 检查所有权限都被拒绝
		allowed, err := acl.Check(ctx, "user2", "resource2", "read")
		if err != nil {
			t.Errorf("Failed to check ACL: %v", err)
		}
		if allowed {
			t.Error("Expected all actions to be denied")
		}
	})

	// 移除条目
	t.Run("RemoveEntry", func(t *testing.T) {
		// 先添加一个条目
		entry := ACLEntry{
			Principal: "user3",
			Resource:  "resource3",
			Actions:   []string{"read"},
			Effect:    "allow",
		}

		if err := acl.AddEntry(ctx, entry); err != nil {
			t.Errorf("Failed to add ACL entry: %v", err)
		}

		// 移除条目
		if err := acl.RemoveEntry(ctx, "user3", "resource3"); err != nil {
			t.Errorf("Failed to remove ACL entry: %v", err)
		}

		// 验证条目已移除（默认拒绝）
		allowed, err := acl.Check(ctx, "user3", "resource3", "read")
		if err != nil {
			t.Errorf("Failed to check ACL: %v", err)
		}
		if allowed {
			t.Error("Expected action to be denied after entry removal")
		}
	})
}

// TestCommandValidation 测试命令验证
func TestCommandValidation(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantAllow bool
	}{
		{"SafeCommand", "ls", true},
		{"SafeCommand2", "echo", true},
		{"DangerousCommand", "rm", false},
		{"DangerousCommand2", "delete", false},
		{"DangerousCommand3", "format", false},
		{"DangerousCommand4", "shutdown", false},
		{"DangerousCommand5", "sudo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllow := isCommandAllowed(tt.command)
			if gotAllow != tt.wantAllow {
				t.Errorf("isCommandAllowed(%q) = %v, want %v", tt.command, gotAllow, tt.wantAllow)
			}
		})
	}
}

// BenchmarkEncryptDecrypt 性能测试
func BenchmarkEncryptDecrypt(b *testing.B) {
	key := []byte("test-encryption-key-32-bytes-long!")
	manager := NewSecurityManager(key, 100, true)
	ctx := context.Background()
	data := []byte("test data for encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := manager.EncryptData(ctx, data)
		manager.DecryptData(ctx, encrypted)
	}
}