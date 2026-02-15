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
)

// ExampleAIOSandboxUsage 演示 AIO Sandbox 的基本用法
func ExampleAIOSandboxUsage() {
	// 创建 AIO Sandbox 实例（使用默认配置）
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		fmt.Printf("Failed to create AIO Sandbox: %v\n", err)
		return
	}
	defer sandbox.Close()

	fmt.Println("AIO Sandbox created successfully!")

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		fmt.Printf("Failed to get tools: %v\n", err)
		return
	}

	fmt.Printf("Available tools: %d\n", len(tools))
	for _, t := range tools {
		// 使用Info()方法获取工具信息
		info, err := t.Info(ctx)
		if err != nil {
			fmt.Printf("Failed to get tool info: %v\n", err)
			continue
		}
		fmt.Printf("  - %s: %s\n", info.Name, info.Desc)
	}

	fmt.Println("AIO Sandbox usage example completed!")
}

// ExampleAIOSandboxWithCustomConfig 演示使用自定义配置创建 AIO Sandbox
func ExampleAIOSandboxWithCustomConfig() {
	// 创建自定义配置
	customConfig := AIOSandboxConfig{
		Browser: BrowserConfig{
			Headless:  true,
			Timeout:   60000,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Viewport: Viewport{
				Width:  1280,
				Height: 800,
			},
		},
		CodeExec: CodeExecutorConfig{
			Timeout:            120000,
			MemoryLimit:        1024,
			CPULimit:           4,
			SupportedLanguages: []string{"go", "python", "javascript", "bash", "java"},
		},
		Shell: ShellConfig{
			Timeout:          60000,
			MemoryLimit:      1024,
			CPULimit:         2,
			CommandWhitelist: []string{"ls", "pwd", "echo"},
			EnableBlacklist:  true,
			CommandBlacklist: []string{"rm", "rmdir", "shutdown", "reboot"},
		},
		File: FileConfig{
			RootDir:     "./sandbox_files",
			MaxFileSize: 200,
			AllowWrite:  true,
			AllowDelete: true,
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
			Enable: false,
			Type:   "http",
			Host:   "localhost",
			Port:   8080,
			ProxyPool: ProxyPoolConfig{
				Enable:              false,
				Proxies:             []string{},
				HealthCheckInterval: 60,
			},
		},
	}

	// 使用自定义配置创建 AIO Sandbox 实例
	sandbox, err := NewAIOSandbox(&customConfig)
	if err != nil {
		fmt.Printf("Failed to create AIO Sandbox with custom config: %v\n", err)
		return
	}
	defer sandbox.Close()

	fmt.Println("AIO Sandbox with custom config created successfully!")

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		fmt.Printf("Failed to get tools: %v\n", err)
		return
	}

	fmt.Printf("Available tools with custom config: %d\n", len(tools))
}

// ExampleAIOSandboxWithEnabledModules 演示如何创建启用了所有模块的 AIO Sandbox
func ExampleAIOSandboxWithEnabledModules() {
	// 获取默认配置
	config := DefaultConfig()

	// 启用所有模块
	config.Visual.Enable = true
	config.Auth.Enable = true
	config.Proxy.Enable = true

	// 创建 AIO Sandbox 实例
	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		fmt.Printf("Failed to create AIO Sandbox with enabled modules: %v\n", err)
		return
	}
	defer sandbox.Close()

	fmt.Println("AIO Sandbox with all modules enabled created successfully!")

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		fmt.Printf("Failed to get tools: %v\n", err)
		return
	}

	fmt.Printf("Available tools with all modules enabled: %d\n", len(tools))
}

// ExampleAIOSandboxDirectModuleAccess 演示如何直接访问 AIO Sandbox 的各个模块
func ExampleAIOSandboxDirectModuleAccess() {
	// 创建 AIO Sandbox 实例
	sandbox, err := NewAIOSandbox(nil)
	if err != nil {
		fmt.Printf("Failed to create AIO Sandbox: %v\n", err)
		return
	}
	defer sandbox.Close()

	fmt.Println("AIO Sandbox created successfully!")

	// 直接访问各个模块
	browserModule := sandbox.Browser()
	codeExecModule := sandbox.CodeExec()
	shellModule := sandbox.Shell()
	fileModule := sandbox.File()
	visualModule := sandbox.Visual()
	authModule := sandbox.Auth()
	proxyModule := sandbox.Proxy()

	fmt.Println("Successfully accessed all modules:")
	fmt.Printf("- Browser module: %v\n", browserModule != nil)
	fmt.Printf("- Code Executor module: %v\n", codeExecModule != nil)
	fmt.Printf("- Shell module: %v\n", shellModule != nil)
	fmt.Printf("- File module: %v\n", fileModule != nil)
	fmt.Printf("- Visual module: %v\n", visualModule != nil)
	fmt.Printf("- Auth module: %v\n", authModule != nil)
	fmt.Printf("- Proxy module: %v\n", proxyModule != nil)
}

// ExampleAIOSandboxIntegrationWithAgent 演示如何将 AIO Sandbox 工具集成到 AI 代理中
func ExampleAIOSandboxIntegrationWithAgent() {
	// 创建 AIO Sandbox 实例
	config := DefaultConfig()
	config.Shell.EnableBlacklist = true
	sandbox, err := NewAIOSandbox(&config)
	if err != nil {
		fmt.Printf("Failed to create AIO Sandbox: %v\n", err)
		return
	}
	defer sandbox.Close()

	fmt.Println("AIO Sandbox created successfully!")

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := sandbox.GetTools(ctx)
	if err != nil {
		fmt.Printf("Failed to get tools: %v\n", err)
		return
	}

	fmt.Printf("Retrieved %d tools for agent integration\n", len(tools))

	// 这里演示了如何将工具集成到 AI 代理中
	// 实际使用时，您需要创建一个 AI 模型实例
	// 然后创建代理时将这些工具传递给代理
	// 例如：
	// chatAgentConfig := agent.ChatAgentConfig{
	//     Name:         "SandboxAgent",
	//     Instructions: "You are an AI agent with access to various tools.",
	//     Model:        model,
	//     Tools:        tools,
	// }
	// chatAgent, _ := agent.NewChatAgent(ctx, chatAgentConfig)

	fmt.Println("AIO Sandbox tools are ready for agent integration!")
}
