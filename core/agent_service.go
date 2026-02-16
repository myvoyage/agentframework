// Agent Framework - Agent Runner Service
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"AgentFramework/agent"
)

// AgentRunnerService handles agent execution operations
type AgentRunnerService struct {
	app *Application
}

// NewAgentRunnerService creates a new agent runner service
func NewAgentRunnerService(app *Application) *AgentRunnerService {
	return &AgentRunnerService{app: app}
}

// AgentType represents the type of agent to run
type AgentType string

const (
	AgentTypeChat    AgentType = "chat"
	AgentTypeReAct   AgentType = "react"
	AgentTypeWorker  AgentType = "worker"
	AgentTypeHuman   AgentType = "human"
)

// RunAgentOptions contains options for running an agent
type RunAgentOptions struct {
	AgentType    AgentType            `json:"agentType"`
	ModelName    string                `json:"modelName"`
	Name         string                `json:"name"`
	Instructions string                `json:"instructions"`
	Tools        []string              `json:"tools,omitempty"`
	Streaming    bool                  `json:"streaming"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RunAgentResponse represents the response from running an agent
type RunAgentResponse struct {
	Success bool        `json:"success"`
	Content string       `json:"content,omitempty"`
	Error   string       `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RunAgent runs an agent with the given options and input
func (s *AgentRunnerService) RunAgent(ctx context.Context, opts *RunAgentOptions, input string) (*RunAgentResponse, error) {
	// Get model factory
	modelFactory := s.app.host.GetModelFactory()
	if modelFactory == nil {
		return &RunAgentResponse{
			Success: false,
			Error:   "model factory not available",
		}, nil
	}

	// Get model
	var model agent.ChatModel
	var err error

	if opts.ModelName != "" {
		model, err = modelFactory(ctx, opts.ModelName)
	} else {
		// Use default model
		model, err = modelFactory(ctx, s.app.config.DefaultModel)
	}

	if err != nil {
		return &RunAgentResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to get model: %v", err),
		}, nil
	}

	// Create agent based on type
	var a agent.Agent
	switch opts.AgentType {
	case AgentTypeChat:
		var err error
		a, err = agent.NewChatAgent(
			ctx,
			agent.ChatAgentConfig{
				Name:         opts.Name,
				Instructions: opts.Instructions,
				Model:        model,
			},
		)
		if err != nil {
			return &RunAgentResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to create chat agent: %v", err),
			}, nil
		}
	case AgentTypeReAct:
		var err error
		a, err = agent.NewChatAgent(
			ctx,
			agent.ChatAgentConfig{
				Name:         opts.Name,
				Instructions: opts.Instructions,
				Model:        model,
			},
		)
		if err != nil {
			return &RunAgentResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to create react agent: %v", err),
			}, nil
		}
	default:
		return &RunAgentResponse{
			Success: false,
			Error:   fmt.Sprintf("unsupported agent type: %s", opts.AgentType),
		}, nil
	}

	// Add tools if specified
	if len(opts.Tools) > 0 {
		for _, _ = range opts.Tools {
			// Get tool from registry
			// This would require implementing tool registry lookup
			// For now, skip
		}
	}

	// Run agent
	response, err := a.Run(ctx, input)
	if err != nil {
		return &RunAgentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &RunAgentResponse{
		Success: true,
		Content: response.Content,
		Metadata: map[string]interface{}{
			"agent_type": string(opts.AgentType),
			"model":      opts.ModelName,
		},
	}, nil
}

// Chat runs a chat agent with simple input/output
func (s *AgentRunnerService) Chat(ctx context.Context, modelName string, input string) (string, error) {
	opts := &RunAgentOptions{
		AgentType:    AgentTypeChat,
		ModelName:    modelName,
		Name:         "CLI Chat Agent",
		Instructions:  "You are a helpful AI assistant",
		Streaming:     false,
	}

	response, err := s.RunAgent(ctx, opts, input)
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("agent failed: %s", response.Error)
	}

	return response.Content, nil
}
