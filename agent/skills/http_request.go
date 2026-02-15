// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// HTTPRequestSkill HTTP请求技能模板
// 用于发送HTTP请求，支持GET、POST、PUT、DELETE等方法
// ------------------------------
type HTTPRequestSkill struct {
	*BaseSkill
	client *http.Client // 复用HTTP客户端，提高性能
}

// NewHTTPRequestSkill 创建一个新的HTTP请求技能实例
func NewHTTPRequestSkill() Skill {
	skill := &HTTPRequestSkill{
		BaseSkill: NewBaseSkill(
			"http_request",
			"Send HTTP requests (GET, POST, PUT, DELETE)",
		),
		client: &http.Client{ // 初始化HTTP客户端，复用连接
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}

	// 更新元数据
	skill.metadata.Category = "network"
	skill.metadata.Tags = []string{"http", "network", "request"}

	return skill
}

// Info 返回HTTP请求技能的元信息
func (s *HTTPRequestSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.name,
		Desc: s.description,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"method": {
				Type:     "string",
				Desc:     "HTTP method: GET, POST, PUT, DELETE",
				Required: true,
			},
			"url": {
				Type:     "string",
				Desc:     "Target URL",
				Required: true,
			},
			"headers": {
				Type:     "object",
				Desc:     "HTTP headers (optional)",
				Required: false,
			},
			"body": {
				Type:     "object",
				Desc:     "Request body (optional)",
				Required: false,
			},
			"timeout": {
				Type:     "integer",
				Desc:     "Request timeout in seconds (optional, default: 30)",
				Required: false,
			},
		}),
	}, nil
}

// Invoke 执行HTTP请求技能
func (s *HTTPRequestSkill) Invoke(ctx context.Context, input string) (string, error) {
	// 解析输入参数
	var params struct {
		Method  string                 `json:"method"`
		URL     string                 `json:"url"`
		Headers map[string]string      `json:"headers"`
		Body    map[string]interface{} `json:"body"`
		Timeout int                    `json:"timeout"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证参数
	if params.Method == "" {
		return "", fmt.Errorf("method is required")
	}
	if params.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	// 创建带超时的上下文
	var reqCtx context.Context
	var cancel context.CancelFunc
	if params.Timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, time.Duration(params.Timeout)*time.Second)
	} else {
		reqCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
	}
	defer cancel()

	// 准备请求体
	var reqBody io.Reader
	var contentType string

	if params.Body != nil {
		bodyBytes, err := json.Marshal(params.Body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(bodyBytes))
		contentType = "application/json"
	}

	// 创建请求
	req, err := http.NewRequestWithContext(reqCtx, strings.ToUpper(params.Method), params.URL, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 添加请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 构建响应头映射
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	// 构建结果
	result := map[string]interface{}{
		"success": true,
		"method":  params.Method,
		"url":     params.URL,
		"status":  resp.StatusCode,
		"headers": respHeaders,
		"body":    string(respBody),
		"message": fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
	}

	// 转换为JSON字符串
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
