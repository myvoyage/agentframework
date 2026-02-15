// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

// ExampleLibrary 示例模板库（对应 TRAE 的 examples.md）
type ExampleLibrary struct {
	examples map[string]*ExampleTemplate
	mu       sync.RWMutex
	baseDir  string
}

// ExampleTemplate 示例模板
type ExampleTemplate struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Category    string   `json:"category" yaml:"category"`
	Tags        []string `json:"tags" yaml:"tags"`
	Version     string   `json:"version" yaml:"version"`

	// 模板内容
	Template   string            `json:"template" yaml:"template"`     // Go template语法
	Parameters map[string]string `json:"parameters" yaml:"parameters"` // 参数说明

	// 示例代码
	ExampleCode string `json:"example_code" yaml:"example_code"`

	// 使用场景
	UseCases []string `json:"use_cases" yaml:"use_cases"`

	// 模板类型
	TemplateType string `json:"template_type" yaml:"template_type"` // "go", "yaml", "json", "markdown"

	// 元数据
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`

	// 加载信息
	SourceFile string `json:"-" yaml:"-"`
	LoadedAt   string `json:"-" yaml:"-"`
}

// NewExampleLibrary 创建示例库
func NewExampleLibrary(baseDir string) *ExampleLibrary {
	if baseDir == "" {
		baseDir = ".skills/examples"
	}

	// 确保目录存在
	os.MkdirAll(baseDir, 0755)

	library := &ExampleLibrary{
		examples: make(map[string]*ExampleTemplate),
		baseDir:  baseDir,
	}

	// 加载所有模板
	library.loadAll()

	return library
}

// Add 添加模板
func (lib *ExampleLibrary) Add(template *ExampleTemplate) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	lib.examples[template.ID] = template
	return nil
}

// Get 获取模板
func (lib *ExampleLibrary) Get(id string) (*ExampleTemplate, bool) {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	tmpl, exists := lib.examples[id]
	return tmpl, exists
}

// Render 渲染模板
func (lib *ExampleLibrary) Render(ctx context.Context, id string, data map[string]interface{}) (string, error) {
	tmpl, exists := lib.Get(id)
	if !exists {
		return "", fmt.Errorf("template %s not found", id)
	}

	// 使用 Go template 渲染
	t, err := template.New(id).Parse(tmpl.Template)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template failed: %w", err)
	}

	return buf.String(), nil
}

// ListByCategory 按分类列出
func (lib *ExampleLibrary) ListByCategory(category string) []*ExampleTemplate {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var results []*ExampleTemplate
	for _, tmpl := range lib.examples {
		if tmpl.Category == category {
			results = append(results, tmpl)
		}
	}

	return results
}

// ListByTag 按标签列出
func (lib *ExampleLibrary) ListByTag(tag string) []*ExampleTemplate {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var results []*ExampleTemplate
	for _, tmpl := range lib.examples {
		for _, t := range tmpl.Tags {
			if t == tag {
				results = append(results, tmpl)
				break
			}
		}
	}

	return results
}

// Search 搜索模板
func (lib *ExampleLibrary) Search(keyword string) []*ExampleTemplate {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var results []*ExampleTemplate
	keyword = strings.ToLower(keyword)

	for _, tmpl := range lib.examples {
		if strings.Contains(strings.ToLower(tmpl.Name), keyword) ||
			strings.Contains(strings.ToLower(tmpl.Description), keyword) {
			results = append(results, tmpl)
		}
	}

	return results
}

// loadAll 加载所有模板
func (lib *ExampleLibrary) loadAll() error {
	entries, err := os.ReadDir(lib.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 加载 YAML 或 JSON 文件
		filePath := filepath.Join(lib.baseDir, entry.Name())
		tmpl, err := lib.loadFromFile(filePath)
		if err != nil {
			fmt.Printf("Warning: failed to load template %s: %v\n", entry.Name(), err)
			continue
		}

		lib.examples[tmpl.ID] = tmpl
	}

	return nil
}

// loadFromFile 从文件加载模板
func (lib *ExampleLibrary) loadFromFile(filePath string) (*ExampleTemplate, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tmpl ExampleTemplate
	ext := filepath.Ext(filePath)

	if ext == ".yaml" || ext == ".yml" {
		if err := yamlUnmarshal(data, &tmpl); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &tmpl); err != nil {
			return nil, err
		}
	}

	tmpl.SourceFile = filePath
	return &tmpl, nil
}

// Save 保存模板到文件
func (lib *ExampleLibrary) Save(tmpl *ExampleTemplate) error {
	if tmpl.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	// 确保目录存在
	if err := os.MkdirAll(lib.baseDir, 0755); err != nil {
		return err
	}

	// 保存为 YAML
	filePath := filepath.Join(lib.baseDir, tmpl.ID+".yaml")
	tmpl.SourceFile = filePath

	data, err := yamlMarshal(tmpl)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// 更新内存中的模板
	lib.mu.Lock()
	lib.examples[tmpl.ID] = tmpl
	lib.mu.Unlock()

	return nil
}

// GetCategories 获取所有分类
func (lib *ExampleLibrary) GetCategories() []string {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	categories := make(map[string]bool)
	for _, tmpl := range lib.examples {
		if tmpl.Category != "" {
			categories[tmpl.Category] = true
		}
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}

	return result
}

// GetTags 获取所有标签
func (lib *ExampleLibrary) GetTags() []string {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	tags := make(map[string]bool)
	for _, tmpl := range lib.examples {
		for _, tag := range tmpl.Tags {
			tags[tag] = true
		}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}

	return result
}

// CreateBuiltInTemplates 创建内置模板
func (lib *ExampleLibrary) CreateBuiltInTemplates() error {
	// HTTP GET 请求模板
	if err := lib.addHTTPGetTemplate(); err != nil {
		return err
	}

	// HTTP POST 请求模板
	if err := lib.addHTTPPostTemplate(); err != nil {
		return err
	}

	// 文件读取模板
	if err := lib.addFileReadTemplate(); err != nil {
		return err
	}

	// 文件写入模板
	if err := lib.addFileWriteTemplate(); err != nil {
		return err
	}

	// 数据转换模板
	if err := lib.addDataTransformTemplate(); err != nil {
		return err
	}

	return nil
}

// addHTTPGetTemplate 添加 HTTP GET 请求模板
func (lib *ExampleLibrary) addHTTPGetTemplate() error {
	tmpl := &ExampleTemplate{
		ID:           "http_get_request",
		Name:         "HTTP GET 请求",
		Description:  "发送 HTTP GET 请求的标准模板",
		Category:     "http",
		Tags:         []string{"http", "get", "request"},
		Version:      "1.0.0",
		TemplateType: "go",
		Template: `// {{.FunctionName}} {{.Description}}
func {{.FunctionName}}(ctx context.Context, {{.Params}}) (map[string]interface{}, error) {
	// 构建URL
	url := fmt.Sprintf({{.URLPattern}}, {{.URLArgs}})

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	{{if .Headers}}
	// 添加请求头
	{{range $key, $value := .Headers}}
	req.Header.Set("{{$key}}", "{{$value}}")
	{{end}}
	{{end}}

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %d", resp.StatusCode)
	}

	// 读取响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return result, nil
}`,
		Parameters: map[string]string{
			"FunctionName": "函数名称",
			"Description":  "函数描述",
			"Params":       "函数参数",
			"URLPattern":   "URL模式，例如: https://api.example.com/users/%d",
			"URLArgs":      "URL参数",
			"Headers":      "请求头",
		},
		ExampleCode: `// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context, userID int64) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.example.com/users/%d", userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}`,
		UseCases: []string{
			"调用 REST API 获取数据",
			"从后端服务获取资源",
			"查询配置信息",
		},
	}

	return lib.Add(tmpl)
}

// addHTTPPostTemplate 添加 HTTP POST 请求模板
func (lib *ExampleLibrary) addHTTPPostTemplate() error {
	tmpl := &ExampleTemplate{
		ID:           "http_post_request",
		Name:         "HTTP POST 请求",
		Description:  "发送 HTTP POST 请求的标准模板",
		Category:     "http",
		Tags:         []string{"http", "post", "request"},
		Version:      "1.0.0",
		TemplateType: "go",
		Template: `// {{.FunctionName}} {{.Description}}
func {{.FunctionName}}(ctx context.Context, {{.Params}}) (map[string]interface{}, error) {
	// 构建URL
	url := {{.URL}}

	// 构建请求体
	body, err := json.Marshal({{.RequestBody}})
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	{{if .Headers}}
	// 添加请求头
	{{range $key, $value := .Headers}}
	req.Header.Set("{{$key}}", "{{$value}}")
	{{end}}
	{{end}}

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return result, nil
}`,
		Parameters: map[string]string{
			"FunctionName": "函数名称",
			"Description":  "函数描述",
			"Params":       "函数参数",
			"URL":          "请求URL",
			"RequestBody":  "请求体",
			"Headers":      "请求头",
		},
		ExampleCode: `// CreateUserData 创建用户数据
func CreateUserData(ctx context.Context, user *User) (map[string]interface{}, error) {
	url := "https://api.example.com/users"

	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}`,
		UseCases: []string{
			"创建资源",
			"提交数据",
			"调用 API 接口",
		},
	}

	return lib.Add(tmpl)
}

// addFileReadTemplate 添加文件读取模板
func (lib *ExampleLibrary) addFileReadTemplate() error {
	tmpl := &ExampleTemplate{
		ID:           "file_read",
		Name:         "文件读取",
		Description:  "读取文件内容的标准模板",
		Category:     "file",
		Tags:         []string{"file", "read", "io"},
		Version:      "1.0.0",
		TemplateType: "go",
		Template: `// {{.FunctionName}} {{.Description}}
func {{.FunctionName}}({{.Params}}) ([]byte, error) {
	// 验证路径
	if !filepath.IsAbs({{.PathArg}}) {
		return nil, fmt.Errorf("absolute path required")
	}

	// 检查文件是否存在
	if _, err := os.Stat({{.PathArg}}); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", {{.PathArg}})
	}

	// 读取文件
	data, err := os.ReadFile({{.PathArg}})
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	{{if .CheckText}}
	// 检查是否为文本文件
	if !isTextFile(data) {
		return nil, fmt.Errorf("file appears to be binary")
	}
	{{end}}

	return data, nil
}`,
		Parameters: map[string]string{
			"FunctionName": "函数名称",
			"Description":  "函数描述",
			"Params":       "函数参数",
			"PathArg":      "路径参数",
			"CheckText":    "是否检查文本文件",
		},
		ExampleCode: `// ReadConfigFile 读取配置文件
func ReadConfigFile(configPath string) ([]byte, error) {
	if !filepath.IsAbs(configPath) {
		return nil, fmt.Errorf("absolute path required")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}`,
		UseCases: []string{
			"读取配置文件",
			"读取日志文件",
			"读取数据文件",
		},
	}

	return lib.Add(tmpl)
}

// addFileWriteTemplate 添加文件写入模板
func (lib *ExampleLibrary) addFileWriteTemplate() error {
	tmpl := &ExampleTemplate{
		ID:           "file_write",
		Name:         "文件写入",
		Description:  "写入文件内容的标准模板",
		Category:     "file",
		Tags:         []string{"file", "write", "io"},
		Version:      "1.0.0",
		TemplateType: "go",
		Template: `// {{.FunctionName}} {{.Description}}
func {{.FunctionName}}({{.Params}}) error {
	// 确保目录存在
	dir := filepath.Dir({{.PathArg}})
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 写入文件
	if err := os.WriteFile({{.PathArg}}, []byte({{.ContentArg}}), 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}`,
		Parameters: map[string]string{
			"FunctionName": "函数名称",
			"Description":  "函数描述",
			"Params":       "函数参数",
			"PathArg":      "路径参数",
			"ContentArg":   "内容参数",
		},
		ExampleCode: `// WriteConfigFile 写入配置文件
func WriteConfigFile(configPath string, config []byte) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, config, 0644); err != nil {
		return err
	}

	return nil
}`,
		UseCases: []string{
			"保存配置文件",
			"写入日志文件",
			"保存数据文件",
		},
	}

	return lib.Add(tmpl)
}

// addDataTransformTemplate 添加数据转换模板
func (lib *ExampleLibrary) addDataTransformTemplate() error {
	tmpl := &ExampleTemplate{
		ID:           "data_transform",
		Name:         "数据转换",
		Description:  "数据格式转换的标准模板",
		Category:     "data",
		Tags:         []string{"data", "transform", "convert"},
		Version:      "1.0.0",
		TemplateType: "go",
		Template: `// {{.FunctionName}} {{.Description}}
func {{.FunctionName}}({{.Params}}) ({{.ReturnType}}, error) {
	// 转换数据
	var result {{.ReturnType}}

	{{if .UseJSON}}
	// JSON 序列化
	data, err := json.Marshal({{.InputArg}})
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	// JSON 反序列化
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	{{end}}

	return result, nil
}`,
		Parameters: map[string]string{
			"FunctionName": "函数名称",
			"Description":  "函数描述",
			"Params":       "函数参数",
			"InputArg":     "输入参数",
			"ReturnType":   "返回类型",
			"UseJSON":      "是否使用JSON",
		},
		ExampleCode: `// MapToStruct 将map转换为struct
func MapToStruct(m map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}`,
		UseCases: []string{
			"数据格式转换",
			"map与struct互转",
			"JSON序列化/反序列化",
		},
	}

	return lib.Add(tmpl)
}

// Clear 清空模板库
func (lib *ExampleLibrary) Clear() {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	lib.examples = make(map[string]*ExampleTemplate)
}

// Count 获取模板数量
func (lib *ExampleLibrary) Count() int {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	return len(lib.examples)
}

// ExportToMarkdown 导出到Markdown
func (lib *ExampleLibrary) ExportToMarkdown(w *string) error {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("# 示例模板库\n\n")
	sb.WriteString(fmt.Sprintf("**生成时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**模板总数**: %d\n\n", len(lib.examples)))

	// 按分类组织
	categories := lib.GetCategories()
	for _, category := range categories {
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.ToUpper(category)))

		templates := lib.ListByCategory(category)
		for _, tmpl := range templates {
			sb.WriteString(fmt.Sprintf("### %s\n\n", tmpl.Name))
			sb.WriteString(fmt.Sprintf("**ID**: `%s`\n\n", tmpl.ID))
			sb.WriteString(fmt.Sprintf("**描述**: %s\n\n", tmpl.Description))

			if len(tmpl.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("**标签**: %s\n\n", strings.Join(tmpl.Tags, ", ")))
			}

			if len(tmpl.UseCases) > 0 {
				sb.WriteString("**使用场景**:\n\n")
				for _, uc := range tmpl.UseCases {
					sb.WriteString(fmt.Sprintf("- %s\n", uc))
				}
				sb.WriteString("\n")
			}

			if tmpl.ExampleCode != "" {
				sb.WriteString("**示例代码**:\n\n```go\n")
				sb.WriteString(tmpl.ExampleCode)
				sb.WriteString("\n```\n\n")
			}

			sb.WriteString("---\n\n")
		}
	}

	*w = sb.String()
	return nil
}
