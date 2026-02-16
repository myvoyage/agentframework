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

package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode defines the error code type
type ErrorCode string

// Error codes for various error types
const (
	// Common errors
	ErrCodeInternal               ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidInput           ErrorCode = "INVALID_INPUT"
	ErrCodeNotFound               ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized           ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden              ErrorCode = "FORBIDDEN"
	ErrCodeTimeout                ErrorCode = "TIMEOUT"
	ErrCodeRateLimit              ErrorCode = "RATE_LIMIT"
	ErrCodeConflict               ErrorCode = "CONFLICT"
	ErrCodeUnavailable            ErrorCode = "UNAVAILABLE"
	ErrCodeMemoryLimitExceeded    ErrorCode = "MEMORY_LIMIT_EXCEEDED"
	ErrCodeMemoryAllocationFailed ErrorCode = "MEMORY_ALLOCATION_FAILED"
	ErrCodeResourceExhausted      ErrorCode = "RESOURCE_EXHAUSTED"

	// Agent errors
	ErrCodeAgentNotFound     ErrorCode = "AGENT_NOT_FOUND"
	ErrCodeAgentCreation     ErrorCode = "AGENT_CREATION_FAILED"
	ErrCodeAgentExecution    ErrorCode = "AGENT_EXECUTION_FAILED"
	ErrCodeAgentTimeout      ErrorCode = "AGENT_TIMEOUT"
	ErrCodeAgentResource     ErrorCode = "AGENT_RESOURCE_ERROR"
	ErrCodeExecutionFailed   ErrorCode = "EXECUTION_FAILED"
	ErrCodeInitFailed        ErrorCode = "INITIALIZATION_FAILED"
	ErrCodeShutdownFailed    ErrorCode = "SHUTDOWN_FAILED"

	// Workflow errors
	ErrCodeWorkflowNotFound  ErrorCode = "WORKFLOW_NOT_FOUND"
	ErrCodeWorkflowCreation  ErrorCode = "WORKFLOW_CREATION_FAILED"
	ErrCodeWorkflowExecution ErrorCode = "WORKFLOW_EXECUTION_FAILED"
	ErrCodeWorkflowTimeout   ErrorCode = "WORKFLOW_TIMEOUT"
	ErrCodeWorkflowState     ErrorCode = "WORKFLOW_STATE_ERROR"
	ErrCodeWorkflowInvalid   ErrorCode = "WORKFLOW_INVALID"

	// Tool errors
	ErrCodeToolNotFound      ErrorCode = "TOOL_NOT_FOUND"
	ErrCodeToolExecution     ErrorCode = "TOOL_EXECUTION_FAILED"
	ErrCodeToolPermission    ErrorCode = "TOOL_PERMISSION_DENIED"
	ErrCodeToolInvalidParams ErrorCode = "TOOL_INVALID_PARAMS"

	// Model errors
	ErrCodeModelNotFound  ErrorCode = "MODEL_NOT_FOUND"
	ErrCodeModelCreation  ErrorCode = "MODEL_CREATION_FAILED"
	ErrCodeModelExecution ErrorCode = "MODEL_EXECUTION_FAILED"
	ErrCodeModelTimeout   ErrorCode = "MODEL_TIMEOUT"

	// Store errors
	ErrCodeStoreNotFound   ErrorCode = "STORE_NOT_FOUND"
	ErrCodeStoreOperation  ErrorCode = "STORE_OPERATION_FAILED"
	ErrCodeStoreTimeout    ErrorCode = "STORE_TIMEOUT"
	ErrCodeStoreConnection ErrorCode = "STORE_CONNECTION_ERROR"

	// Plugin errors
	ErrCodeDownloadFailed ErrorCode = "PLUGIN_DOWNLOAD_FAILED"
	ErrCodeInstallFailed  ErrorCode = "PLUGIN_INSTALL_FAILED"
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
)

// AgentError defines the standard error structure for Agent Framework
type AgentError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AgentError) Error() string {
	parts := []string{fmt.Sprintf("%s: %s", e.Code, e.Message)}
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %v", e.Cause))
	}
	if len(e.Details) > 0 {
		parts = append(parts, fmt.Sprintf("Details: %v", e.Details))
	}
	return strings.Join(parts, "; ")
}

// Unwrap implements the errors.Unwrap interface
func (e *AgentError) Unwrap() error {
	return e.Cause
}

// Is implements the errors.Is interface
func (e *AgentError) Is(target error) bool {
	if targetErr, ok := target.(*AgentError); ok {
		return e.Code == targetErr.Code
	}
	return false
}

// New creates a new AgentError with the given code and message
func New(code ErrorCode, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new AgentError with the given code and formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *AgentError {
	return &AgentError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with a new AgentError
func Wrap(err error, code ErrorCode, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with a new AgentError and formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AgentError {
	return &AgentError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// WithDetails adds details to an AgentError
func (e *AgentError) WithDetails(details map[string]interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithDetail adds a single detail to an AgentError
func (e *AgentError) WithDetail(key string, value interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// IsAgentError checks if an error is an AgentError
func IsAgentError(err error) bool {
	_, ok := err.(*AgentError)
	return ok
}

// GetAgentError extracts an AgentError from an error chain
func GetAgentError(err error) (*AgentError, bool) {
	var agentErr *AgentError
	if err != nil {
		if errors.As(err, &agentErr) {
			return agentErr, true
		}
	}
	return nil, false
}

// IsErrorCode checks if an error is an AgentError with the given code
func IsErrorCode(err error, code ErrorCode) bool {
	if agentErr, ok := GetAgentError(err); ok {
		return agentErr.Code == code
	}
	return false
}
