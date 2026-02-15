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

// AIOSandboxConfig 是 AIO Sandbox 的整体配置结构体，包含所有功能模块的配置
// 通过该配置结构体，可以自定义 AIO Sandbox 的各个模块的行为
//
// 示例用法：
//
//	config := aiosandbox.AIOSandboxConfig{
//	    Browser: aiosandbox.BrowserConfig{
//	        Headless: true,
//	        Timeout:  30000,
//	    },
//	    // 其他模块配置...
//	}
type AIOSandboxConfig struct {
	Browser   BrowserConfig      // 浏览器模块配置
	CodeExec  CodeExecutorConfig // 代码执行模块配置
	Shell     ShellConfig        // Shell 命令模块配置
	File      FileConfig         // 文件操作模块配置
	Visual    VisualConfig       // 可视化接管模块配置
	Auth      AuthConfig         // 鉴权机制模块配置
	Proxy     ProxyConfig        // 代理支持模块配置
	WebSearch WebSearchConfig    // 网络搜索模块配置
}

// BrowserConfig 浏览器模块配置
type BrowserConfig struct {
	// Headless 是否使用无头模式
	Headless bool `json:"headless"`
	// Timeout 操作超时时间（毫秒）
	Timeout int `json:"timeout"`
	// UserAgent 用户代理字符串
	UserAgent string `json:"user_agent"`
	// Viewport 视口大小
	Viewport Viewport `json:"viewport"`
}

// Viewport 浏览器视口配置
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CodeExecutorConfig 代码执行模块配置
type CodeExecutorConfig struct {
	// Timeout 代码执行超时时间（毫秒）
	Timeout int `json:"timeout"`
	// MemoryLimit 内存限制（MB）
	MemoryLimit int `json:"memory_limit"`
	// CPULimit CPU核心数限制
	CPULimit int `json:"cpu_limit"`
	// SupportedLanguages 支持的编程语言列表
	SupportedLanguages []string `json:"supported_languages"`
}

// ShellConfig Shell模块配置
type ShellConfig struct {
	// Timeout 命令执行超时时间（毫秒）
	Timeout int `json:"timeout"`
	// MemoryLimit 内存限制（MB）
	MemoryLimit int `json:"memory_limit"`
	// CPULimit CPU核心数限制
	CPULimit int `json:"cpu_limit"`
	// CommandWhitelist 允许执行的命令白名单
	CommandWhitelist []string `json:"command_whitelist"`
	// EnableBlacklist 是否启用命令黑名单
	EnableBlacklist bool `json:"enable_blacklist"`
	// CommandBlacklist 禁止执行的命令黑名单
	CommandBlacklist []string `json:"command_blacklist"`
}

// FileConfig 文件操作模块配置
type FileConfig struct {
	// RootDir 文件操作根目录
	RootDir string `json:"root_dir"`
	// MaxFileSize 最大文件大小（MB）
	MaxFileSize int `json:"max_file_size"`
	// AllowWrite 是否允许写入操作
	AllowWrite bool `json:"allow_write"`
	// AllowDelete 是否允许删除操作
	AllowDelete bool `json:"allow_delete"`
}

// VisualConfig 可视化模块配置
type VisualConfig struct {
	// Enable 是否启用可视化模块
	Enable bool `json:"enable"`
	// Port 可视化服务端口
	Port int `json:"port"`
	// Host 可视化服务主机
	Host string `json:"host"`
}

// AuthConfig 鉴权模块配置
type AuthConfig struct {
	// Enable 是否启用鉴权
	Enable bool `json:"enable"`
	// APIKey API密钥
	APIKey string `json:"api_key"`
	// OAuth2 OAuth2.0配置
	OAuth2 OAuth2Config `json:"oauth2"`
}

// OAuth2Config OAuth2.0配置
type OAuth2Config struct {
	// ClientID 客户端ID
	ClientID string `json:"client_id"`
	// ClientSecret 客户端密钥
	ClientSecret string `json:"client_secret"`
	// RedirectURL 重定向URL
	RedirectURL string `json:"redirect_url"`
	// Scopes 权限范围
	Scopes []string `json:"scopes"`
}

// ProxyConfig 代理模块配置
type ProxyConfig struct {
	// Enable 是否启用代理
	Enable bool `json:"enable"`
	// Type 代理类型（http, https, socks5）
	Type string `json:"type"`
	// Host 代理主机
	Host string `json:"host"`
	// Port 代理端口
	Port int `json:"port"`
	// Username 代理用户名
	Username string `json:"username"`
	// Password 代理密码
	Password string `json:"password"`
	// ProxyPool 代理池配置
	ProxyPool ProxyPoolConfig `json:"proxy_pool"`
}

// ProxyPoolConfig 代理池配置
type ProxyPoolConfig struct {
	// Enable 是否启用代理池
	Enable bool `json:"enable"`
	// Proxies 代理列表
	Proxies []string `json:"proxies"`
	// HealthCheckInterval 健康检查间隔（秒）
	HealthCheckInterval int `json:"health_check_interval"`
}

// WebSearchConfig 网络搜索模块配置
type WebSearchConfig struct {
	// Enable 是否启用网络搜索模块
	Enable bool `json:"enable"`
	// Timeout 搜索超时时间（毫秒）
	Timeout int `json:"timeout"`
	// MaxResults 最大返回结果数
	MaxResults int `json:"max_results"`
	// EnableSearchers 启用的搜索引擎列表
	// 可选: "google", "bing", "duckduckgo"
	EnableSearchers []string `json:"enable_searchers"`
	// APIKeys 各搜索引擎的API密钥
	// 对于Google: 需要提供 "google" 和 "google_cx" (自定义搜索引擎ID)
	// 对于Bing: 需要提供 "bing"
	// DuckDuckGo 不需要API密钥
	APIKeys map[string]string `json:"api_keys"`
	// UserAgent 自定义User-Agent
	UserAgent string `json:"user_agent"`
	// CacheEnabled 是否启用缓存
	CacheEnabled bool `json:"cache_enabled"`
	// CacheTTL 缓存过期时间（秒）
	CacheTTL int `json:"cache_ttl"`
	// SafeSearch 是否启用安全搜索
	SafeSearch bool `json:"safe_search"`
	// DefaultLanguage 默认搜索语言
	DefaultLanguage string `json:"default_language"`
	// DefaultRegion 默认搜索区域
	DefaultRegion string `json:"default_region"`
}

// DefaultConfig 返回默认的AIO Sandbox配置
//
// 默认配置包含了各模块的合理默认值，适合大多数使用场景
// 浏览器：无头模式，30秒超时，1920x1080视口
// 代码执行：60秒超时，512MB内存，2核CPU，支持go/python/javascript/bash
// Shell：30秒超时，512MB内存，1核CPU，启用黑名单
// 文件：当前目录，100MB最大文件大小，允许写入和删除
// 可视化：禁用
// 鉴权：禁用
// 代理：禁用
// 网络搜索：启用DuckDuckGo，30秒超时，10条结果
//
// 示例用法：
// config := aiosandbox.DefaultConfig()
// // 自定义部分配置
// config.Browser.Headless = false
// config.CodeExec.Timeout = 120000
func DefaultConfig() AIOSandboxConfig {
	return AIOSandboxConfig{
		Browser: BrowserConfig{
			Headless:  true,
			Timeout:   30000,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Viewport: Viewport{
				Width:  1920,
				Height: 1080,
			},
		},
		CodeExec: CodeExecutorConfig{
			Timeout:            60000,
			MemoryLimit:        512,
			CPULimit:           2,
			SupportedLanguages: []string{"go", "python", "javascript", "bash"},
		},
		Shell: ShellConfig{
			Timeout:          30000,
			MemoryLimit:      512,
			CPULimit:         1,
			CommandWhitelist: []string{},
			EnableBlacklist:  true,
			CommandBlacklist: []string{"rm", "rmdir", "shutdown", "reboot"},
		},
		File: FileConfig{
			RootDir:     ".",
			MaxFileSize: 100,
			AllowWrite:  true,
			AllowDelete: true,
		},
		Visual: VisualConfig{
			Enable: false,
			Port:   8080,
			Host:   "localhost",
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
		WebSearch: WebSearchConfig{
			Enable:          true,
			Timeout:         30000,
			MaxResults:      10,
			EnableSearchers: []string{"duckduckgo"},
			APIKeys:         make(map[string]string),
			UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			CacheEnabled:    true,
			CacheTTL:        3600,
			SafeSearch:      false,
			DefaultLanguage: "zh-CN",
			DefaultRegion:   "CN",
		},
	}
}
