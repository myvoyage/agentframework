package einobridge

import (
	"context"
	"AgentFramework/internal/pipelineengine"
)

// Ensure compile-time references to pipelineengine types work when RPC bridge is wired.
// This file provides a minimal mock RPC client/server that delegates to the local engine.

type MockRPCClient struct {
	Engine *pipelineengine.PipelineEngine
}

func (c *MockRPCClient) InvokeTool(ctx context.Context, req MCPInvokeToolRequest) (MCPInvokeToolResponse, error) {
	if c.Engine == nil {
		return MCPInvokeToolResponse{Success: false, Error: "engine not initialized"}, nil
	}
	data, err := c.Engine.InvokeTool(ctx, req.Tool, req.Params, req.Context)
	if err != nil {
		return MCPInvokeToolResponse{Success: false, Error: err.Error()}, nil
	}
	return MCPInvokeToolResponse{Success: true, Data: data}, nil
}

type MockRPCServer struct {
	Engine *pipelineengine.PipelineEngine
}

func (s *MockRPCServer) HandleInvokeTool(ctx context.Context, req MCPInvokeToolRequest) (MCPInvokeToolResponse, error) {
	if s.Engine == nil {
		return MCPInvokeToolResponse{Success: false, Error: "engine not initialized"}, nil
	}
	data, err := s.Engine.InvokeTool(ctx, req.Tool, req.Params, req.Context)
	if err != nil {
		return MCPInvokeToolResponse{Success: false, Error: err.Error()}, nil
	}
	return MCPInvokeToolResponse{Success: true, Data: data}, nil
}
