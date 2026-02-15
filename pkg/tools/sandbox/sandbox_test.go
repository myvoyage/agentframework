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
	"testing"
)

// TestNewAIOSandbox 测试创建 AIO Sandbox 实例
func TestNewAIOSandbox(t *testing.T) {
	// 使用默认配置创建实例（传递nil）
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		t.Fatalf("Failed to create AIO Sandbox: %v", err)
	}
	if sandbox == nil {
		t.Fatal("Expected sandbox to be non-nil")
	}

	// 验证模块是否正确初始化
	if sandbox.Browser() == nil {
		t.Fatal("Expected browser module to be non-nil")
	}
	if sandbox.CodeExec() == nil {
		t.Fatal("Expected code exec module to be non-nil")
	}
	if sandbox.Shell() == nil {
		t.Fatal("Expected shell module to be non-nil")
	}
	if sandbox.File() == nil {
		t.Fatal("Expected file module to be non-nil")
	}
	if sandbox.Visual() == nil {
		t.Fatal("Expected visual module to be non-nil")
	}
	if sandbox.Auth() == nil {
		t.Fatal("Expected auth module to be non-nil")
	}
	if sandbox.Proxy() == nil {
		t.Fatal("Expected proxy module to be non-nil")
	}
}

// TestAIOSandboxGetTools 测试获取 MCP 工具列表
func TestAIOSandboxGetTools(t *testing.T) {
	// 使用默认配置创建实例（传递nil）
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		t.Fatalf("Failed to create AIO Sandbox: %v", err)
	}
	if sandbox == nil {
		t.Fatal("Expected sandbox to be non-nil")
	}

	// 获取工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	// 验证工具列表不为空
	if len(tools) <= 0 {
		t.Fatal("Expected tools to be non-empty")
	}

	// 打印工具列表，方便调试
	t.Logf("Available tools: %d", len(tools))
	for _, tool := range tools {
		// 使用Info()方法获取工具信息
		info, err := tool.Info(ctx)
		if err != nil {
			t.Logf("Failed to get tool info: %v", err)
			continue
		}
		t.Logf("Tool: %s - %s", info.Name, info.Desc)
	}
}

// TestAIOSandboxWithEnabledModules 测试启用特定模块
func TestAIOSandboxWithEnabledModules(t *testing.T) {
	// 使用默认配置创建实例
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		t.Fatalf("Failed to create AIO Sandbox: %v", err)
	}
	if sandbox == nil {
		t.Fatal("Expected sandbox to be non-nil")
	}

	// 获取工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	// 验证工具列表包含工具
	if len(tools) <= 0 {
		t.Fatal("Expected tools to be non-empty")
	}
}

// TestAIOSandboxWithCustomConfig 测试使用自定义配置创建 AIO Sandbox 实例
func TestAIOSandboxWithCustomConfig(t *testing.T) {
	// 创建自定义配置
	customConfig := AIOSandboxConfig{
		Browser: BrowserConfig{
			Headless:  false,
			Timeout:   60000,
			UserAgent: "Custom User Agent",
			Viewport: Viewport{
				Width:  1280,
				Height: 800,
			},
		},
		CodeExec: CodeExecutorConfig{
			Timeout:            120000,
			MemoryLimit:        1024,
			CPULimit:           4,
			SupportedLanguages: []string{"go", "python", "javascript"},
		},
		Shell: ShellConfig{
			Timeout:          60000,
			MemoryLimit:      512,
			CPULimit:         2,
			CommandWhitelist: []string{"ls", "pwd", "echo"},
			EnableBlacklist:  true,
			CommandBlacklist: []string{"rm", "rmdir"},
		},
		File: FileConfig{
			RootDir:     "./test_files",
			MaxFileSize: 50,
			AllowWrite:  true,
			AllowDelete: false,
		},
		Visual: VisualConfig{
			Enable: true,
			Port:   8081,
			Host:   "0.0.0.0",
		},
		Auth: AuthConfig{
			Enable: false,
			APIKey: "",
			OAuth2: OAuth2Config{
				ClientID:     "",
				ClientSecret: "",
				RedirectURL:  "",
				Scopes:       []string{},
			},
		},
		Proxy: ProxyConfig{
			Enable:   true,
			Type:     "http",
			Host:     "localhost",
			Port:     3128,
			Username: "testuser",
			Password: "testpass",
			ProxyPool: ProxyPoolConfig{
				Enable:              true,
				Proxies:             []string{"http://proxy1:3128", "http://proxy2:3128"},
				HealthCheckInterval: 30,
			},
		},
	}

	// 使用自定义配置创建实例
	sandbox, err := NewAIOSandbox(&customConfig)
	if err != nil {
		t.Fatalf("Failed to create AIO Sandbox with custom config: %v", err)
	}
	if sandbox == nil {
		t.Fatal("Expected sandbox to be non-nil")
	}

	// 验证模块是否正确初始化
	if sandbox.Browser() == nil {
		t.Fatal("Expected browser module to be non-nil")
	}
	if sandbox.CodeExec() == nil {
		t.Fatal("Expected code exec module to be non-nil")
	}
	if sandbox.Shell() == nil {
		t.Fatal("Expected shell module to be non-nil")
	}
	if sandbox.File() == nil {
		t.Fatal("Expected file module to be non-nil")
	}
	if sandbox.Visual() == nil {
		t.Fatal("Expected visual module to be non-nil")
	}
	if sandbox.Auth() == nil {
		t.Fatal("Expected auth module to be non-nil")
	}
	if sandbox.Proxy() == nil {
		t.Fatal("Expected proxy module to be non-nil")
	}

	// 获取工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	// 验证工具列表不为空
	if len(tools) <= 0 {
		t.Fatal("Expected tools to be non-empty")
	}

	t.Logf("Created AIO Sandbox with custom config successfully, got %d tools", len(tools))
}
