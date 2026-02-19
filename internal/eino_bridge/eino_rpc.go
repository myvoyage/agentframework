// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package einobridge

import (
	"context"
	"encoding/json"
)

// MCPInvokeToolRequest represents a request to invoke a tool via the MCP/RPC bridge
type MCPInvokeToolRequest struct {
	Tool    string                 `json:"tool"`
	Params  map[string]interface{} `json:"params"`
	Context map[string]interface{} `json:"context"`
	Version string                 `json:"version"`
}

// MCPInvokeToolResponse represents the result of a tool invocation via the bridge
type MCPInvokeToolResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error"`
}

// Marshal / Unmarshal helpers for JSON transport
func (r *MCPInvokeToolResponse) MarshalJSON() ([]byte, error) {
	type Alias MCPInvokeToolResponse
	return json.Marshal((*Alias)(r))
}

func (r *MCPInvokeToolRequest) MarshalJSON() ([]byte, error) {
	type Alias MCPInvokeToolRequest
	return json.Marshal((*Alias)(r))
}

// Simple RPC client interface for Eino bridge (MVP draft)
type EinoRPCClient interface {
	InvokeTool(ctx context.Context, req MCPInvokeToolRequest) (MCPInvokeToolResponse, error)
}

// Simple RPC server interface for Eino bridge (MVP draft)
type EinoRPCServer interface {
	HandleInvokeTool(ctx context.Context, req MCPInvokeToolRequest) (MCPInvokeToolResponse, error)
}
