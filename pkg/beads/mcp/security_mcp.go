// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package mcp provides security MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"AgentFramework/agent"
	"AgentFramework/pkg/beads/security"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SecurityMCPTools provides security MCP tools.
type SecurityMCPTools struct {
	agent *agent.SecurityAgent
}

// NewSecurityMCPTools creates a new SecurityMCPTools instance.
func NewSecurityMCPTools(securityAgent *agent.SecurityAgent) *SecurityMCPTools {
	return &SecurityMCPTools{
		agent: securityAgent,
	}
}

// RegisterTools registers all security MCP tools with the MCP server.
func (t *SecurityMCPTools) RegisterTools(s *server.MCPServer) {
	// ValidateCommand tool
	s.AddTool(mcp.Tool{
		Name:        "validate_security_command",
		Description: "Validate a command before execution for security risks",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Command to validate",
				},
				"params": map[string]interface{}{
					"type":        "object",
					"description": "Command parameters",
				},
			},
			Required: []string{"command"},
		},
	}, t.handleValidateCommand)

	// CheckPermissions tool
	s.AddTool(mcp.Tool{
		Name:        "check_security_permissions",
		Description: "Check if a user has permission to access a resource",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "User identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform (e.g., read, write, execute)",
				},
			},
			Required: []string{"user_id", "resource", "action"},
		},
	}, t.handleCheckPermissions)

	// EncryptData tool
	s.AddTool(mcp.Tool{
		Name:        "encrypt_security_data",
		Description: "Encrypt data using AES-GCM encryption",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Data to encrypt",
				},
				"encoding": map[string]interface{}{
					"type":        "string",
					"description": "Output encoding format (base64 or hex)",
					"enum":        []string{"base64", "hex"},
					"default":     "base64",
				},
			},
			Required: []string{"data"},
		},
	}, t.handleEncryptData)

	// DecryptData tool
	s.AddTool(mcp.Tool{
		Name:        "decrypt_security_data",
		Description: "Decrypt data that was encrypted using AES-GCM",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"encrypted_data": map[string]interface{}{
					"type":        "string",
					"description": "Encrypted data",
				},
				"encoding": map[string]interface{}{
					"type":        "string",
					"description": "Input encoding format (base64 or hex)",
					"enum":        []string{"base64", "hex"},
					"default":     "base64",
				},
			},
			Required: []string{"encrypted_data"},
		},
	}, t.handleDecryptData)

	// GrantPermission tool
	s.AddTool(mcp.Tool{
		Name:        "grant_security_permission",
		Description: "Grant a user permission to perform an action on a resource",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "User identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to grant permission for (e.g., read, write, execute, *)",
				},
			},
			Required: []string{"user_id", "resource", "action"},
		},
	}, t.handleGrantPermission)

	// RevokePermission tool
	s.AddTool(mcp.Tool{
		Name:        "revoke_security_permission",
		Description: "Revoke a user's permission to perform an action on a resource",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "User identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to revoke permission for",
				},
			},
			Required: []string{"user_id", "resource", "action"},
		},
	}, t.handleRevokePermission)

	// GetAuditLog tool
	s.AddTool(mcp.Tool{
		Name:        "get_security_audit_log",
		Description: "Get the security audit log",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of entries to return (0 = unlimited)",
				},
			},
		},
	}, t.handleGetAuditLog)

	// ClearAuditLog tool
	s.AddTool(mcp.Tool{
		Name:        "clear_security_audit_log",
		Description: "Clear the security audit log",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleClearAuditLog)

	// AddACLEntry tool
	s.AddTool(mcp.Tool{
		Name:        "add_security_acl_entry",
		Description: "Add an entry to the access control list",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"principal": map[string]interface{}{
					"type":        "string",
					"description": "Principal (user or group) identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
				"actions": map[string]interface{}{
					"type":        "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "List of actions (e.g., [\"read\", \"write\"] or [\"*\"] for all)",
				},
				"effect": map[string]interface{}{
					"type":        "string",
					"description": "Effect of the ACL entry",
					"enum":        []string{"allow", "deny"},
				},
			},
			Required: []string{"principal", "resource", "actions", "effect"},
		},
	}, t.handleAddACLEntry)

	// RemoveACLEntry tool
	s.AddTool(mcp.Tool{
		Name:        "remove_security_acl_entry",
		Description: "Remove an entry from the access control list",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"principal": map[string]interface{}{
					"type":        "string",
					"description": "Principal identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
			},
			Required: []string{"principal", "resource"},
		},
	}, t.handleRemoveACLEntry)

	// CheckACL tool
	s.AddTool(mcp.Tool{
		Name:        "check_security_acl",
		Description: "Check if an action is allowed via the access control list",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"principal": map[string]interface{}{
					"type":        "string",
					"description": "Principal identifier",
				},
				"resource": map[string]interface{}{
					"type":        "string",
					"description": "Resource identifier",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to check",
				},
			},
			Required: []string{"principal", "resource", "action"},
		},
	}, t.handleCheckACL)

	// GetSecurityStatus tool
	s.AddTool(mcp.Tool{
		Name:        "get_security_status",
		Description: "Get the current security status",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleGetSecurityStatus)

	// EnableSecurity tool
	s.AddTool(mcp.Tool{
		Name:        "enable_security",
		Description: "Enable security enforcement",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleEnableSecurity)

	// DisableSecurity tool
	s.AddTool(mcp.Tool{
		Name:        "disable_security",
		Description: "Disable security enforcement",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, t.handleDisableSecurity)
}

func (t *SecurityMCPTools) handleValidateCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command, _ := request.Params.Arguments["command"].(string)
	params, _ := request.Params.Arguments["params"].(map[string]interface{})

	if err := t.agent.ValidateCommand(ctx, command, params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Command validation failed: %v", err)), nil
	}

	return mcp.NewToolResultText("Command validated successfully"), nil
}

func (t *SecurityMCPTools) handleCheckPermissions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, _ := request.Params.Arguments["user_id"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)
	action, _ := request.Params.Arguments["action"].(string)

	allowed, err := t.agent.CheckPermissions(ctx, userID, resource, action)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Permission check failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"allowed": allowed,
		"message": fmt.Sprintf("Permission %s for user %s on resource %s: %v", action, userID, resource, allowed),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *SecurityMCPTools) handleEncryptData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, _ := request.Params.Arguments["data"].(string)
	encoding, _ := request.Params.Arguments["encoding"].(string)

	if encoding == "" {
		encoding = "base64"
	}

	encrypted, err := t.agent.EncryptData(ctx, []byte(data))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Encryption failed: %v", err)), nil
	}

	var result string
	switch encoding {
	case "base64":
		result, err = t.agent.EncryptToBase64(ctx, []byte(data))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Encryption failed: %v", err)), nil
		}
	default:
		return mcp.NewToolResultError(fmt.Sprintf("Unsupported encoding: %s", encoding)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Encrypted data (%s): %s", encoding, result)), nil
}

func (t *SecurityMCPTools) handleDecryptData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	encryptedData, _ := request.Params.Arguments["encrypted_data"].(string)
	encoding, _ := request.Params.Arguments["encoding"].(string)

	if encoding == "" {
		encoding = "base64"
	}

	var decrypted []byte
	var err error

	switch encoding {
	case "base64":
		decrypted, err = t.agent.DecryptFromBase64(ctx, encryptedData)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Decryption failed: %v", err)), nil
		}
	default:
		return mcp.NewToolResultError(fmt.Sprintf("Unsupported encoding: %s", encoding)), nil
	}

	return mcp.NewToolResultText(string(decrypted)), nil
}

func (t *SecurityMCPTools) handleGrantPermission(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, _ := request.Params.Arguments["user_id"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)
	action, _ := request.Params.Arguments["action"].(string)

	if err := t.agent.GrantPermission(ctx, userID, resource, action); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to grant permission: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully granted permission: user %s can %s resource %s", userID, action, resource)), nil
}

func (t *SecurityMCPTools) handleRevokePermission(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, _ := request.Params.Arguments["user_id"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)
	action, _ := request.Params.Arguments["action"].(string)

	if err := t.agent.RevokePermission(ctx, userID, resource, action); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to revoke permission: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully revoked permission: user %s cannot %s resource %s", userID, action, resource)), nil
}

func (t *SecurityMCPTools) handleGetAuditLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 0
	if lim, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(lim)
	}

	log, err := t.agent.GetAuditLog(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get audit log: %v", err)), nil
	}

	logJSON, _ := json.MarshalIndent(log, "", "  ")
	return mcp.NewToolResultText(string(logJSON)), nil
}

func (t *SecurityMCPTools) handleClearAuditLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := t.agent.ClearAuditLog(ctx); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to clear audit log: %v", err)), nil
	}

	return mcp.NewToolResultText("Successfully cleared audit log"), nil
}

func (t *SecurityMCPTools) handleAddACLEntry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	principal, _ := request.Params.Arguments["principal"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)

	actionsInterface, _ := request.Params.Arguments["actions"].(interface{})
	actionsArray, _ := actionsInterface.([]interface{})
	actions := make([]string, 0, len(actionsArray))

	for _, actionInterface := range actionsArray {
		if action, ok := actionInterface.(string); ok {
			actions = append(actions, action)
		}
	}

	effect, _ := request.Params.Arguments["effect"].(string)

	if err := t.agent.AddACLEntry(ctx, principal, resource, actions, effect); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add ACL entry: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully added ACL entry: %s %s %s -> %s", principal, resource, actions, effect)), nil
}

func (t *SecurityMCPTools) handleRemoveACLEntry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	principal, _ := request.Params.Arguments["principal"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)

	if err := t.agent.RemoveACLEntry(ctx, principal, resource); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove ACL entry: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully removed ACL entry: %s %s", principal, resource)), nil
}

func (t *SecurityMCPTools) handleCheckACL(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	principal, _ := request.Params.Arguments["principal"].(string)
	resource, _ := request.Params.Arguments["resource"].(string)
	action, _ := request.Params.Arguments["action"].(string)

	allowed, err := t.agent.CheckACL(ctx, principal, resource, action)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("ACL check failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"allowed": allowed,
		"message": fmt.Sprintf("ACL check for %s on %s by %s: %v", action, resource, principal, allowed),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (t *SecurityMCPTools) handleGetSecurityStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := t.agent.GetSecurityStatus(ctx)

	statusJSON, _ := json.MarshalIndent(status, "", "  ")
	return mcp.NewToolResultText(string(statusJSON)), nil
}

func (t *SecurityMCPTools) handleEnableSecurity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.agent.Enable(ctx)
	return mcp.NewToolResultText("Security enabled"), nil
}

func (t *SecurityMCPTools) handleDisableSecurity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.agent.Disable(ctx)
	return mcp.NewToolResultText("Security disabled"), nil
}