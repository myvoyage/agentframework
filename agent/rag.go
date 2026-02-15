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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// GraphlitClient interface defines the contract for interacting with Graphlit.
type GraphlitClient interface {
	Ingest(ctx context.Context, content string, metadata map[string]interface{}) error
	Query(ctx context.Context, text string) (string, error)
}

// RealGraphlitClient is a real implementation that makes HTTP calls to Graphlit API.
type RealGraphlitClient struct {
	apiKey       string
	endpoint     string
	httpClient   *http.Client
	queryTimeout time.Duration
	ingestTimeout time.Duration
}

// GraphlitIngestRequest represents the request body for ingesting content.
type GraphlitIngestRequest struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphlitIngestResponse represents the response from ingesting content.
type GraphlitIngestResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
}

// GraphlitQueryRequest represents the request body for querying content.
type GraphlitQueryRequest struct {
	Query     string `json:"query"`
	MaxResults int   `json:"maxResults,omitempty"`
}

// GraphlitQueryResponse represents the response from querying content.
type GraphlitQueryResponse struct {
	Success bool `json:"success"`
	Results []struct {
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
		Score    float64                `json:"score"`
	} `json:"results"`
	Error string `json:"error,omitempty"`
}

// NewRealGraphlitClient creates a new Graphlit client with real API integration.
// It requires GRAPHLIT_API_KEY environment variable.
// Optionally set GRAPHLIT_ENDPOINT (default: https://api.graphlit.io/v1).
func NewRealGraphlitClient() (*RealGraphlitClient, error) {
	apiKey := os.Getenv("GRAPHLIT_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GRAPHLIT_API_KEY environment variable not set")
	}

	endpoint := os.Getenv("GRAPHLIT_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.graphlit.io/v1"
	}

	return &RealGraphlitClient{
		apiKey:        apiKey,
		endpoint:      endpoint,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		queryTimeout:  10 * time.Second,
		ingestTimeout: 5 * time.Second,
	}, nil
}

// NewRealGraphlitClientWithConfig creates a new Graphlit client with custom configuration.
func NewRealGraphlitClientWithConfig(apiKey, endpoint string, timeout time.Duration) (*RealGraphlitClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	if endpoint == "" {
		endpoint = "https://api.graphlit.io/v1"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &RealGraphlitClient{
		apiKey:        apiKey,
		endpoint:      endpoint,
		httpClient:    &http.Client{Timeout: timeout},
		queryTimeout:  10 * time.Second,
		ingestTimeout: 5 * time.Second,
	}, nil
}

// Ingest stores content to Graphlit's vector database.
func (c *RealGraphlitClient) Ingest(ctx context.Context, content string, metadata map[string]interface{}) error {
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.ingestTimeout)
	defer cancel()

	// Prepare request body
	reqBody := GraphlitIngestRequest{
		Content:  content,
		Metadata: metadata,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/ingest", c.endpoint)
	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Graphlit API returned status %d", resp.StatusCode)
	}

	// Parse response
	var ingestResp GraphlitIngestResponse
	if err := json.NewDecoder(resp.Body).Decode(&ingestResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !ingestResp.Success {
		return fmt.Errorf("ingest failed: %s", ingestResp.Error)
	}

	return nil
}

// Query retrieves relevant content from Graphlit's vector database.
func (c *RealGraphlitClient) Query(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()

	// Prepare request body
	reqBody := GraphlitQueryRequest{
		Query:     text,
		MaxResults: 5,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/query", c.endpoint)
	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Graphlit API returned status %d", resp.StatusCode)
	}

	// Parse response
	var queryResp GraphlitQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !queryResp.Success {
		return "", fmt.Errorf("query failed: %s", queryResp.Error)
	}

	// Format results
	if len(queryResp.Results) == 0 {
		return "No relevant memory found.", nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d relevant result(s):\n\n", len(queryResp.Results)))

	for i, result := range queryResp.Results {
		output.WriteString(fmt.Sprintf("Result %d (relevance: %.2f):\n", i+1, result.Score))
		output.WriteString(result.Content)
		output.WriteString("\n\n")

		if len(result.Metadata) > 0 {
			output.WriteString("Metadata:\n")
			for key, value := range result.Metadata {
				output.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
			output.WriteString("\n")
		}
	}

	return output.String(), nil
}

// MockGraphlitClient is a mock implementation for demonstration and testing.
type MockGraphlitClient struct {
	memory []string
	mu     sync.RWMutex
}

func NewMockGraphlitClient() *MockGraphlitClient {
	return &MockGraphlitClient{
		memory: make([]string, 0),
	}
}

func (c *MockGraphlitClient) Ingest(ctx context.Context, content string, metadata map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memory = append(c.memory, content)
	return nil
}

func (c *MockGraphlitClient) Query(ctx context.Context, text string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Simple keyword match for mock
	var results []string
	for _, mem := range c.memory {
		if strings.Contains(mem, text) {
			results = append(results, mem)
		}
	}

	if len(results) == 0 {
		return "No relevant memory found.", nil
	}

	return strings.Join(results, "\n---\n"), nil
}

// NewGraphlitClient creates a Graphlit client, automatically choosing
// between real and mock implementation based on environment configuration.
func NewGraphlitClient() (GraphlitClient, error) {
	// Try to create real client first
	if client, err := NewRealGraphlitClient(); err == nil {
		return client, nil
	}

	// Fall back to mock client if API key not configured
	fmt.Println("Warning: GRAPHLIT_API_KEY not set, using mock Graphlit client")
	return NewMockGraphlitClient(), nil
}

// GraphlitRAGTool is a tool that allows agents to search long-term memory.
type GraphlitRAGTool struct {
	client GraphlitClient
}

func NewGraphlitRAGTool(client GraphlitClient) tool.BaseTool {
	return &GraphlitRAGTool{client: client}
}

func (t *GraphlitRAGTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_memory",
		Desc: "Search through long-term memory and knowledge base for relevant information.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "The search query to find relevant memories.",
			},
		}),
	}, nil
}

func (t *GraphlitRAGTool) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	// In a real scenario, arguments is a JSON string, we need to parse it.
	// For this simple mock, we assume arguments IS the query or simple JSON.
	// Let's do a simple clean up.
	query := arguments
	if strings.Contains(arguments, ":") {
		// Very naive JSON parsing assumption for demo
		parts := strings.SplitN(arguments, ":", 2)
		if len(parts) > 1 {
			query = strings.Trim(parts[1], " \"'}")
		}
	}

	return t.client.Query(ctx, query)
}

// RAGMiddleware creates a middleware that automatically retrieves context from Graphlit
// and appends it to the user input before the agent processes it.
func NewRAGMiddleware(client GraphlitClient) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				// 1. Retrieve relevant context
				contextInfo, err := client.Query(ctx, input)
				if err != nil {
					// Log error but proceed without context
					fmt.Printf("RAG retrieval failed: %v\n", err)
				}

				// 2. Augment input
				augmentedInput := input
				if contextInfo != "" && contextInfo != "No relevant memory found." {
					augmentedInput = fmt.Sprintf("Context from memory:\n%s\n\nUser Request:\n%s", contextInfo, input)
				}

				// 3. Execute Agent
				resp, err := next.Run(ctx, augmentedInput, opts...)

				// 4. Async Ingest (Store the interaction for future RAG)
				// In a real app, this should be non-blocking
				if err == nil && resp != nil {
					go func() {
						_ = client.Ingest(context.Background(), fmt.Sprintf("User: %s\nAssistant: %s", input, resp.Content), nil)
					}()
				}

				return resp, err
			},
		}
	}
}
