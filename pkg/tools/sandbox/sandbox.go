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

	"github.com/cloudwego/eino/components/tool"

	"AgentFramework/pkg/tools/sandbox/auth"
	"AgentFramework/pkg/tools/sandbox/browser"
	"AgentFramework/pkg/tools/sandbox/code"
	"AgentFramework/pkg/tools/sandbox/file"
	"AgentFramework/pkg/tools/sandbox/proxy"
	"AgentFramework/pkg/tools/sandbox/shell"
	"AgentFramework/pkg/tools/sandbox/visual"
	"AgentFramework/pkg/tools/sandbox/websearch"
)

// AIOSandbox 是 AIO Sandbox 的核心组件，集成了所有功能模块
type AIOSandbox struct {
	browser   *browser.BrowserModule
	codeExec  *code.CodeExecutorModule
	shell     *shell.ShellModule
	file      *file.FileModule
	visual    *visual.VisualModule
	auth      *auth.AuthModule
	proxy     *proxy.ProxyModule
	webSearch *websearch.WebSearchModule
}

// NewAIOSandbox 创建一个新的 AIO Sandbox 实例
// config 是 AIO Sandbox 的配置，如果为 nil 则使用默认配置
func NewAIOSandbox(config *AIOSandboxConfig) (*AIOSandbox, error) {
	// 如果没有提供配置，使用默认配置
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}

	// 初始化各个模块
	browser, err := browser.NewBrowserModule(browser.BrowserConfig{
		Headless:  config.Browser.Headless,
		Timeout:   config.Browser.Timeout,
		UserAgent: config.Browser.UserAgent,
		Viewport: browser.Viewport{
			Width:  int64(config.Browser.Viewport.Width),
			Height: int64(config.Browser.Viewport.Height),
		},
	})
	if err != nil {
		return nil, err
	}

	codeExec, err := code.NewCodeExecutorModule(code.CodeExecutorConfig{
		Timeout:            config.CodeExec.Timeout,
		MemoryLimit:        config.CodeExec.MemoryLimit,
		CPULimit:           config.CodeExec.CPULimit,
		SupportedLanguages: config.CodeExec.SupportedLanguages,
	})
	if err != nil {
		return nil, err
	}

	shell, err := shell.NewShellModule(shell.ShellConfig{
		Timeout:          config.Shell.Timeout,
		MemoryLimit:      config.Shell.MemoryLimit,
		CPULimit:         config.Shell.CPULimit,
		CommandWhitelist: config.Shell.CommandWhitelist,
		EnableBlacklist:  config.Shell.EnableBlacklist,
		CommandBlacklist: config.Shell.CommandBlacklist,
	})
	if err != nil {
		return nil, err
	}

	file, err := file.NewFileModule(file.FileConfig{
		RootDir:     config.File.RootDir,
		MaxFileSize: int64(config.File.MaxFileSize),
		AllowWrite:  config.File.AllowWrite,
		AllowDelete: config.File.AllowDelete,
	})
	if err != nil {
		return nil, err
	}

	visual, err := visual.NewVisualModule(visual.VisualConfig{
		Enable: config.Visual.Enable,
		Port:   config.Visual.Port,
		Host:   config.Visual.Host,
	})
	if err != nil {
		return nil, err
	}

	auth, err := auth.NewAuthModule(auth.AuthConfig{
		Enable:    config.Auth.Enable,
		JWTSecret: config.Auth.APIKey,
		JWTExpiry: 3600,
		JWTIssuer: "aio-sandbox",
		OAuth2: auth.OAuth2Config{
			ClientID:     config.Auth.OAuth2.ClientID,
			ClientSecret: config.Auth.OAuth2.ClientSecret,
			RedirectURL:  config.Auth.OAuth2.RedirectURL,
			Scopes:       config.Auth.OAuth2.Scopes,
		},
	})
	if err != nil {
		return nil, err
	}

	proxy, err := proxy.NewProxyModule(proxy.ProxyConfig{
		Enable:              config.Proxy.Enable,
		Type:                config.Proxy.Type,
		Host:                config.Proxy.Host,
		Port:                config.Proxy.Port,
		Username:            config.Proxy.Username,
		Password:            config.Proxy.Password,
		PoolSize:            5,
		HealthCheckInterval: config.Proxy.ProxyPool.HealthCheckInterval,
		HealthCheckURL:      "https://www.google.com",
		Strategy:            "round_robin",
	})
	if err != nil {
		return nil, err
	}

	// 初始化 WebSearch 模块
	webSearchConfig := websearch.WebSearchConfig{
		Timeout:         config.WebSearch.Timeout,
		MaxResults:      config.WebSearch.MaxResults,
		EnableSearchers: config.WebSearch.EnableSearchers,
		APIKeys:         config.WebSearch.APIKeys,
		UserAgent:       config.WebSearch.UserAgent,
		CacheEnabled:    config.WebSearch.CacheEnabled,
		CacheTTL:        config.WebSearch.CacheTTL,
		SafeSearch:      config.WebSearch.SafeSearch,
	}

	webSearch, err := websearch.NewWebSearchModule(webSearchConfig)
	if err != nil {
		return nil, err
	}

	return &AIOSandbox{
		browser:   browser,
		codeExec:  codeExec,
		shell:     shell,
		file:      file,
		visual:    visual,
		auth:      auth,
		proxy:     proxy,
		webSearch: webSearch,
	}, nil
}

// GetTools 返回所有模块的 MCP 工具列表
func (s *AIOSandbox) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{}

	// 获取各个模块的工具
	browserTools, err := s.browser.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, browserTools...)

	codeExecTools, err := s.codeExec.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, codeExecTools...)

	shellTools, err := s.shell.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, shellTools...)

	fileTools, err := s.file.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, fileTools...)

	visualTools, err := s.visual.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, visualTools...)

	authTools, err := s.auth.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, authTools...)

	proxyTools, err := s.proxy.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, proxyTools...)

	// 获取 WebSearch 工具
	webSearchTools, err := s.webSearch.GetTools(ctx)
	if err != nil {
		return nil, err
	}
	tools = append(tools, webSearchTools...)

	return tools, nil
}

// Close 关闭 AIO Sandbox，释放资源
func (s *AIOSandbox) Close() error {
	// 关闭各个模块
	if err := s.browser.Close(); err != nil {
		return err
	}
	if err := s.codeExec.Close(); err != nil {
		return err
	}
	if err := s.shell.Close(); err != nil {
		return err
	}
	if err := s.file.Close(); err != nil {
		return err
	}
	if err := s.visual.Close(); err != nil {
		return err
	}
	if err := s.auth.Close(); err != nil {
		return err
	}
	if err := s.proxy.Close(); err != nil {
		return err
	}
	if err := s.webSearch.Close(); err != nil {
		return err
	}
	return nil
}

// Browser 返回浏览器模块
func (s *AIOSandbox) Browser() *browser.BrowserModule {
	return s.browser
}

// CodeExec 返回代码执行模块
func (s *AIOSandbox) CodeExec() *code.CodeExecutorModule {
	return s.codeExec
}

// Shell 返回 Shell 模块
func (s *AIOSandbox) Shell() *shell.ShellModule {
	return s.shell
}

// File 返回文件模块
func (s *AIOSandbox) File() *file.FileModule {
	return s.file
}

// Visual 返回可视化模块
func (s *AIOSandbox) Visual() *visual.VisualModule {
	return s.visual
}

// Auth 返回鉴权模块
func (s *AIOSandbox) Auth() *auth.AuthModule {
	return s.auth
}

// Proxy 返回代理模块
func (s *AIOSandbox) Proxy() *proxy.ProxyModule {
	return s.proxy
}

// WebSearch 返回网络搜索模块
func (s *AIOSandbox) WebSearch() *websearch.WebSearchModule {
	return s.webSearch
}
