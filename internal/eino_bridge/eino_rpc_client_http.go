//go:build eino
// +build eino

package einobridge

import (
	"bytes"
	"encoding/json"
	pe "AgentFramework/internal/pipelineengine"
	"net/http"
	"time"
)

type HTTPRPCClient struct {
	BaseURL string
	Client  *http.Client
}

func NewHTTPRPCClient(url string) *HTTPRPCClient {
	return &HTTPRPCClient{BaseURL: url, Client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPRPCClient) InvokeTool(req pe.MCPInvokeToolRequest) (pe.MCPInvokeToolResponse, error) {
	payload, _ := json.Marshal(req)
	resp, err := c.Client.Post(c.BaseURL+"/invoke_tool", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return pe.MCPInvokeToolResponse{Success: false, Error: err.Error()}, err
	}
	defer resp.Body.Close()
	var out pe.MCPInvokeToolResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out, nil
}

func (c *HTTPRPCClient) RunPipeline(p *pe.PipelineSpec) (*pe.ExecutionContext, error) {
	payload := struct {
		Pipeline *pe.PipelineSpec `json:"pipeline"`
	}{Pipeline: p}
	data, _ := json.Marshal(payload)
	resp, err := c.Client.Post(c.BaseURL+"/run_pipeline", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r struct {
		Success bool
		Data    *pe.ExecutionContext
		Error   string
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Success {
		return nil, fmt.Errorf(r.Error)
	}
	return r.Data, nil
}
