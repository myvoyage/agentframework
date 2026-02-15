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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v3"
)

// ==================== HTTP Tool Loader ====================

// HTTPToolLoader 从HTTP URL加载工具定义
type HTTPToolLoader struct {
	client *http.Client
	timeout time.Duration
}

// NewHTTPToolLoader 创建HTTP工具加载器
func NewHTTPToolLoader() *HTTPToolLoader {
	return &HTTPToolLoader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// Name 返回加载器名称
func (l *HTTPToolLoader) Name() string {
	return "http"
}

// CanLoad 检查是否可以加载
func (l *HTTPToolLoader) CanLoad(source string) bool {
	return strings.HasPrefix(source, "url:") ||
		strings.HasPrefix(source, "http:") ||
		strings.HasPrefix(source, "https:")
}

// Load 从URL加载工具
func (l *HTTPToolLoader) Load(ctx context.Context, source string) (tool.BaseTool, error) {
	// 提取URL
	url := strings.TrimPrefix(source, "url:")
	if url == source {
		url = source
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析工具定义
	return l.parseToolDefinition(body)
}

// parseToolDefinition 解析工具定义
func (l *HTTPToolLoader) parseToolDefinition(data []byte) (tool.BaseTool, error) {
	// 尝试解析为JSON格式的工具定义
	var def ToolDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse tool definition: %w", err)
	}

	return NewToolFromDefinition(&def)
}

// Validate 验证工具
func (l *HTTPToolLoader) Validate(ctx context.Context, t tool.BaseTool) error {
	info, err := t.Info(ctx)
	if err != nil {
		return err
	}

	if info.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if info.Desc == "" {
		return fmt.Errorf("tool description cannot be empty")
	}

	return nil
}

// ==================== File Tool Loader ====================

// FileToolLoader 从文件系统加载工具
type FileToolLoader struct {
	baseDir    string
	extensions []string
}

// NewFileToolLoader 创建文件工具加载器
func NewFileToolLoader(baseDir string) *FileToolLoader {
	return &FileToolLoader{
		baseDir:    baseDir,
		extensions: []string{".json", ".yaml", ".yml"},
	}
}

// Name 返回加载器名称
func (l *FileToolLoader) Name() string {
	return "file"
}

// CanLoad 检查是否可以加载
func (l *FileToolLoader) CanLoad(source string) bool {
	return strings.HasPrefix(source, "file:") ||
		strings.HasPrefix(source, "/") ||
		strings.HasPrefix(source, ".") ||
		(len(source) > 1 && source[1] == ':') // Windows路径
}

// Load 从文件加载工具
func (l *FileToolLoader) Load(ctx context.Context, source string) (tool.BaseTool, error) {
	// 提取文件路径
	path := strings.TrimPrefix(source, "file:")

	// 处理相对路径
	if !filepath.IsAbs(path) && l.baseDir != "" {
		path = filepath.Join(l.baseDir, path)
	}

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 根据文件扩展名解析
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return l.loadJSONTool(data)
	case ".yaml", ".yml":
		return l.loadYAMLTool(data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// loadJSONTool 加载JSON格式工具
func (l *FileToolLoader) loadJSONTool(data []byte) (tool.BaseTool, error) {
	var def ToolDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return NewToolFromDefinition(&def)
}

// loadYAMLTool 加载YAML格式工具
func (l *FileToolLoader) loadYAMLTool(data []byte) (tool.BaseTool, error) {
	// 解析 YAML 为 ToolDefinition
	var def ToolDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 验证必要字段
	if def.Name == "" {
		return nil, fmt.Errorf("tool name is required in YAML definition")
	}
	if def.Description == "" {
		return nil, fmt.Errorf("tool description is required in YAML definition")
	}

	// 如果有实现函数引用，需要处理
	// YAML 格式可以包含实现代码或函数引用
	if def.Function != "" {
		// 检查是否是内置函数
		if builtinImpl, ok := builtinImplementations[def.Name]; ok {
			return NewToolWithImplementation(&def, builtinImpl)
		}
		// 否则返回错误，因为 YAML 中不能直接包含代码
		return nil, fmt.Errorf("function '%s' referenced in YAML is not a built-in implementation", def.Function)
	}

	return NewToolFromDefinition(&def)
}

// builtinImplementations 内置工具实现映射
var builtinImplementations = map[string]func(ctx context.Context, input string) (string, error){
	"echo": func(ctx context.Context, input string) (string, error) {
		return input, nil
	},
	"reverse": func(ctx context.Context, input string) (string, error) {
		runes := []rune(input)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	},
	"upper": func(ctx context.Context, input string) (string, error) {
		return strings.ToUpper(input), nil
	},
	"lower": func(ctx context.Context, input string) (string, error) {
		return strings.ToLower(input), nil
	},
	"length": func(ctx context.Context, input string) (string, error) {
		return fmt.Sprintf("%d", len(input)), nil
	},
	"timestamp": func(ctx context.Context, input string) (string, error) {
		return time.Now().Format(time.RFC3339), nil
	},
}

// Validate 验证工具
func (l *FileToolLoader) Validate(ctx context.Context, t tool.BaseTool) error {
	return validateTool(ctx, t)
}

// ==================== Plugin Tool Loader ====================

// PluginToolLoader 从Go插件加载工具
type PluginToolLoader struct {
	baseDir string
}

// NewPluginToolLoader 创建插件工具加载器
func NewPluginToolLoader(baseDir string) *PluginToolLoader {
	return &PluginToolLoader{
		baseDir: baseDir,
	}
}

// Name 返回加载器名称
func (l *PluginToolLoader) Name() string {
	return "plugin"
}

// CanLoad 检查是否可以加载
func (l *PluginToolLoader) CanLoad(source string) bool {
	return strings.HasPrefix(source, "plugin:") ||
		strings.HasSuffix(source, ".so") ||
		strings.HasSuffix(source, ".dll")
}

// Load 从插件加载工具
func (l *PluginToolLoader) Load(ctx context.Context, source string) (tool.BaseTool, error) {
	// 提取插件路径
	pluginPath := strings.TrimPrefix(source, "plugin:")

	// 处理相对路径
	if !filepath.IsAbs(pluginPath) && l.baseDir != "" {
		pluginPath = filepath.Join(l.baseDir, pluginPath)
	}

	// 加载插件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// 查找符号
	newToolFunc, err := p.Lookup("NewTool")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export NewTool function: %w", err)
	}

	// 调用函数创建工具
	toolFunc, ok := newToolFunc.(func() (tool.BaseTool, error))
	if !ok {
		return nil, fmt.Errorf("NewTool has invalid signature")
	}

	return toolFunc()
}

// Validate 验证工具
func (l *PluginToolLoader) Validate(ctx context.Context, t tool.BaseTool) error {
	return validateTool(ctx, t)
}

// ==================== MCP Tool Loader ====================

// MCPToolLoader 从MCP服务器加载工具
type MCPToolLoader struct {
	timeout     time.Duration
	transport   string // "stdio" 或 "sse"
	clientCache map[string]*MCPClient
	mu          sync.RWMutex
}

// MCPClient MCP 客户端
type MCPClient struct {
	address   string
	transport string
	timeout   time.Duration
	tools     map[string]*MCPToolDefinition
	mu        sync.RWMutex
	process   *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
}

// Disconnect 断开 MCP 服务器连接
func (c *MCPClient) Disconnect(ctx context.Context) error {
	// 目前为空实现，后续可以添加实际的断开连接逻辑
	return nil
}

// MCPToolDefinition MCP 工具定义
type MCPToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema  map[string]interface{} `json:"inputSchema"`
}

// MCPRequest MCP 请求
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPResponse MCP 响应
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError MCP 错误
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewMCPToolLoader 创建MCP工具加载器
func NewMCPToolLoader() *MCPToolLoader {
	return &MCPToolLoader{
		timeout:     10 * time.Second,
		transport:   "stdio", // 默认使用 stdio
		clientCache: make(map[string]*MCPClient),
	}
}

// SetTransport 设置传输方式
func (l *MCPToolLoader) SetTransport(transport string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.transport = transport
}

// Name 返回加载器名称
func (l *MCPToolLoader) Name() string {
	return "mcp"
}

// CanLoad 检查是否可以加载
func (l *MCPToolLoader) CanLoad(source string) bool {
	return strings.HasPrefix(source, "mcp:")
}

// Load 从MCP服务器加载工具
func (l *MCPToolLoader) Load(ctx context.Context, source string) (tool.BaseTool, error) {
	// 提取MCP服务器地址
	serverAddr := strings.TrimPrefix(source, "mcp:")

	// 创建或获取客户端
	client, err := l.getOrCreateClient(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	// 初始化连接并获取工具列表
	if err := client.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// 获取所有可用工具
	tools := client.ListTools()
	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools available from MCP server")
	}

	// 创建一个代理工具，可以调用 MCP 服务器上的所有工具
	return l.createMCPProxyTool(client, tools), nil
}

// getOrCreateClient 获取或创建 MCP 客户端
func (l *MCPToolLoader) getOrCreateClient(serverAddr string) (*MCPClient, error) {
	l.mu.RLock()
	if client, exists := l.clientCache[serverAddr]; exists {
		l.mu.RUnlock()
		return client, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	// 再次检查，避免重复创建
	if client, exists := l.clientCache[serverAddr]; exists {
		return client, nil
	}

	client := &MCPClient{
		address:   serverAddr,
		transport: l.transport,
		timeout:   l.timeout,
		tools:     make(map[string]*MCPToolDefinition),
	}

	l.clientCache[serverAddr] = client
	return client, nil
}

// createMCPProxyTool 创建 MCP 代理工具
func (l *MCPToolLoader) createMCPProxyTool(client *MCPClient, tools map[string]*MCPToolDefinition) tool.BaseTool {
	// 构建工具描述
	descriptions := make([]string, 0, len(tools))
	for _, tool := range tools {
		descriptions = append(descriptions, fmt.Sprintf("- %s: %s", tool.Name, tool.Description))
	}

	return &DynamicTool{
		definition: &ToolDefinition{
			Name:        "mcp_proxy",
			Description: fmt.Sprintf("MCP工具代理，可调用以下工具:\n%s", strings.Join(descriptions, "\n")),
			Parameters: map[string]*ParameterDef{
				"tool_name": {
					Type:        "string",
					Description: "要调用的工具名称",
					Required:    true,
					Enum:        l.getToolNames(tools),
				},
				"arguments": {
					Type:        "object",
					Description: "工具参数（JSON格式）",
					Required:    true,
				},
			},
		},
		impl: func(ctx context.Context, input string) (string, error) {
			// 解析输入
			var args struct {
				ToolName  string          `json:"tool_name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err := json.Unmarshal([]byte(input), &args); err != nil {
				return "", fmt.Errorf("failed to parse MCP arguments: %w", err)
			}

			// 调用 MCP 工具
			return client.CallTool(ctx, args.ToolName, args.Arguments)
		},
	}
}

// getToolNames 获取工具名称列表
func (l *MCPToolLoader) getToolNames(tools map[string]*MCPToolDefinition) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	return names
}

// Initialize 初始化 MCP 客户端
func (c *MCPClient) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 发送初始化请求
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "AgentFramework",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP initialize error: %s", resp.Error.Message)
	}

	// 获取工具列表
	if err := c.fetchTools(ctx); err != nil {
		return err
	}

	return nil
}

// fetchTools 获取工具列表
func (c *MCPClient) fetchTools(ctx context.Context) error {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP tools/list error: %s", resp.Error.Message)
	}

	// 解析工具列表
	var result struct {
		Tools []struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			InputSchema  map[string]interface{} `json:"inputSchema"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to parse tools list: %w", err)
	}

	// 存储工具定义
	for _, tool := range result.Tools {
		c.tools[tool.Name] = &MCPToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema:  tool.InputSchema,
		}
	}

	return nil
}

// ListTools 列出所有工具
func (c *MCPClient) ListTools() map[string]*MCPToolDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make(map[string]*MCPToolDefinition, len(c.tools))
	for name, tool := range c.tools {
		tools[name] = tool
	}
	return tools
}

// CallTool 调用工具
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments json.RawMessage) (string, error) {
	c.mu.RLock()
	_, exists := c.tools[toolName]
	c.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("tool '%s' not found", toolName)
	}

	// 构建请求
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": json.RawMessage(arguments),
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("MCP tool call error: %s", resp.Error.Message)
	}

	// 解析结果
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return string(resp.Result), nil
	}

	// 提取文本内容
	var output strings.Builder
	for _, content := range result.Content {
		if content.Type == "text" {
			output.WriteString(content.Text)
		}
	}

	return output.String(), nil
}

// sendRequest 发送请求到 MCP 服务器
func (c *MCPClient) sendRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	if c.transport == "stdio" {
		return c.sendSTDIORequest(ctx, req)
	}
	return nil, fmt.Errorf("unsupported transport: %s", c.transport)
}

// sendSTDIORequest 发送 STDIO 请求
func (c *MCPClient) sendSTDIORequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	// 编码请求
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 检查是否已经有运行中的 MCP 服务器进程
	if c.process == nil {
		// 从配置中获取服务器路径
		pluginPath := "database-integration-tool"
		mcpConfigPath := filepath.Join(pluginPath, ".mcp.json")

		// 读取 MCP 配置
		var mcpConfig struct {
			MCPServers map[string]struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env"`
				Disabled bool             `json:"disabled"`
			} `json:"mcpServers"`
		}

		configData, err := os.ReadFile(mcpConfigPath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configData, &mcpConfig); err != nil {
			return nil, err
		}

		// 假设只有一个 MCP 服务器
		var serverConfig struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
			Disabled bool             `json:"disabled"`
		}
		for _, config := range mcpConfig.MCPServers {
			serverConfig = config
			break
		}

		// 替换环境变量
		envVars := make([]string, 0, len(serverConfig.Env))
		for key, value := range serverConfig.Env {
			// 替换 ${CLAUDE_PLUGIN_ROOT} 变量
			value = strings.ReplaceAll(value, "${CLAUDE_PLUGIN_ROOT}", pluginPath)
			// 替换其他环境变量
			value = os.ExpandEnv(value)
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}

		// 替换命令参数中的环境变量
		for i, arg := range serverConfig.Args {
			arg = strings.ReplaceAll(arg, "${CLAUDE_PLUGIN_ROOT}", ".")
			serverConfig.Args[i] = os.ExpandEnv(arg)
		}

		// 启动 MCP 服务器进程
		cmd := exec.Command(serverConfig.Command, serverConfig.Args...)
		cmd.Dir = pluginPath
		cmd.Env = append(os.Environ(), envVars...)
		cmd.Env = append(cmd.Env, "NO_COLOR=true") // 禁用彩色输出

		// 创建管道用于通信
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		// 保存进程和管道引用
		c.process = cmd
		c.stdin = stdin
		c.stdout = stdout
		c.stderr = stderr

		// 启动一个 goroutine 来读取错误输出
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderr.Read(buf)
				if err != nil {
					break
				}
				fmt.Fprintf(os.Stderr, "MCP Server Error: %s", buf[:n])
			}
		}()
	}

	// 发送请求
	// 先读取并丢弃服务器的启动日志
	go func() {
		buf := make([]byte, 1024)
		for {
			_, err := c.stdout.Read(buf)
			if err != nil {
				break
			}
			// 检查是否已经读取到服务器启动完成的信息
			if strings.Contains(string(buf), "Database Integration MCP Server running on stdio") {
				break
			}
		}
	}()

	// 等待服务器启动完成
	time.Sleep(2 * time.Second)

	_, err = c.stdin.Write(append(reqData, '\n'))
	if err != nil {
		return nil, err
	}

	// 读取响应
	buf := make([]byte, 1024)
	n, err := c.stdout.Read(buf)
	if err != nil {
		return nil, err
	}

	// 过滤掉颜色代码
	respData := string(buf[:n])
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	respData = re.ReplaceAllString(respData, "")

	// 解码响应
	var resp MCPResponse
	err = json.Unmarshal([]byte(respData), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// StartMCPServer 启动 MCP 服务器进程（辅助函数）
func StartMCPServer(command string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)

	// 创建管道用于通信
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// 在实际应用中，需要保存这些管道引用以便后续通信
	_ = stdin
	_ = stdout
	_ = stderr

	return cmd, nil
}

// Validate 验证工具
func (l *MCPToolLoader) Validate(ctx context.Context, t tool.BaseTool) error {
	return validateTool(ctx, t)
}

// ==================== 工具定义和创建 ====================

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Parameters  map[string]*ParameterDef    `json:"parameters"`
	Function    string                      `json:"function,omitempty"`
	Metadata    map[string]string           `json:"metadata,omitempty"`
}

// ParameterDef 参数定义
type ParameterDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// DynamicTool 动态工具
type DynamicTool struct {
	definition *ToolDefinition
	impl       func(ctx context.Context, input string) (string, error)
}

// NewToolFromDefinition 从定义创建工具
func NewToolFromDefinition(def *ToolDefinition) (tool.BaseTool, error) {
	if def.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if def.Description == "" {
		return nil, fmt.Errorf("tool description is required")
	}

	return &DynamicTool{
		definition: def,
	}, nil
}

// NewToolWithImplementation 创建带实现的功能
func NewToolWithImplementation(def *ToolDefinition, impl func(ctx context.Context, input string) (string, error)) (tool.BaseTool, error) {
	tool, err := NewToolFromDefinition(def)
	if err != nil {
		return nil, err
	}

	dt := tool.(*DynamicTool)
	dt.impl = impl
	return dt, nil
}

// Info 返回工具信息
func (t *DynamicTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	params := make(map[string]*schema.ParameterInfo)
	for name, param := range t.definition.Parameters {
		// 根据参数类型字符串映射到 DataType
		var dataType schema.DataType
		switch strings.ToLower(param.Type) {
		case "string":
			dataType = schema.String
		case "number":
			dataType = schema.Number
		case "integer":
			dataType = schema.Integer
		case "boolean":
			dataType = schema.Boolean
		case "array":
			dataType = schema.Array
		case "object":
			dataType = schema.Object
		default:
			dataType = schema.String // 默认字符串类型
		}

		params[name] = &schema.ParameterInfo{
			Type:     dataType,
			Desc:     param.Description,
			Required: param.Required,
		}
	}

	return &schema.ToolInfo{
		Name:        t.definition.Name,
		Desc:        t.definition.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

// InvokableRun 执行工具
func (t *DynamicTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	if t.impl == nil {
		return "", fmt.Errorf("tool implementation not provided")
	}
	return t.impl(ctx, input)
}

// ==================== 辅助函数 ====================

// validateTool 验证工具
func validateTool(ctx context.Context, t tool.BaseTool) error {
	info, err := t.Info(ctx)
	if err != nil {
		return err
	}

	if info.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if info.Desc == "" {
		return fmt.Errorf("tool description cannot be empty")
	}

	return nil
}

// LoadToolsFromDirectory 从目录加载所有工具
func LoadToolsFromDirectory(ctx context.Context, dir string) ([]tool.BaseTool, error) {
	loader := NewFileToolLoader(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tools []tool.BaseTool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if !loader.CanLoad(path) {
			continue
		}

		t, err := loader.Load(ctx, path)
		if err != nil {
			// 记录错误但继续加载其他工具
			continue
		}

		tools = append(tools, t)
	}

	return tools, nil
}

// LoadToolsFromConfig 从配置文件加载工具列表
func LoadToolsFromConfig(ctx context.Context, configPath string) ([]tool.BaseTool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Tools []struct {
			Source string `json:"source"`
			Name   string `json:"name,omitempty"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	registry, err := NewDynamicToolRegistry(DynamicToolRegistryConfig{
		EnableCache: true,
	})
	if err != nil {
		return nil, err
	}

	// 注册默认加载器
	_ = registry.RegisterLoader(NewHTTPToolLoader())
	_ = registry.RegisterLoader(NewFileToolLoader(""))

	var tools []tool.BaseTool
	for _, toolConfig := range config.Tools {
		t, err := registry.LoadFromSource(ctx, toolConfig.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to load tool %s: %w", toolConfig.Name, err)
		}
		tools = append(tools, t)
	}

	return tools, nil
}
