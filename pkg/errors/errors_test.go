// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeNotFound, "test error")
	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Empty(t, err.Cause)
	assert.Nil(t, err.Details)
}

func TestNewf(t *testing.T) {
	err := Newf(ErrCodeInvalidInput, "invalid value: %d", 42)
	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeInvalidInput, err.Code)
	assert.Equal(t, "invalid value: 42", err.Message)
	assert.Empty(t, err.Cause)
	assert.Nil(t, err.Details)
}

func TestWrap(t *testing.T) {
	cause := New(ErrCodeInternal, "internal error")
	err := Wrap(cause, ErrCodeAgentExecution, "agent execution failed")
	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeAgentExecution, err.Code)
	assert.Equal(t, "agent execution failed", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.Nil(t, err.Details)
}

func TestWrapf(t *testing.T) {
	cause := New(ErrCodeInternal, "internal error")
	err := Wrapf(cause, ErrCodeAgentExecution, "agent execution failed for task %d", 100)
	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeAgentExecution, err.Code)
	assert.Equal(t, "agent execution failed for task 100", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.Nil(t, err.Details)
}

func TestWithDetails(t *testing.T) {
	err := New(ErrCodeNotFound, "test error").WithDetails(map[string]interface{}{
		"key": "value",
		"count": 42,
	})

	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.NotNil(t, err.Details)
	assert.Len(t, err.Details, 2)
	assert.Equal(t, "value", err.Details["key"])
	assert.Equal(t, 42, err.Details["count"])
}

func TestWithDetail(t *testing.T) {
	err := New(ErrCodeNotFound, "test error").WithDetail("key", "value").WithDetail("count", 42)

	assert.IsType(t, &AgentError{}, err)
	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.NotNil(t, err.Details)
	assert.Len(t, err.Details, 2)
	assert.Equal(t, "value", err.Details["key"])
	assert.Equal(t, 42, err.Details["count"])
}

func TestIsAgentError(t *testing.T) {
	err1 := New(ErrCodeNotFound, "test error")
	err2 := &AgentError{}

	assert.True(t, IsAgentError(err1))
	assert.True(t, IsAgentError(err2))
	assert.False(t, IsAgentError(nil))
	assert.False(t, IsAgentError(assert.AnError))
}

func TestGetAgentError(t *testing.T) {
	err := New(ErrCodeNotFound, "test error")
	agentErr, ok := GetAgentError(err)

	assert.True(t, ok)
	assert.IsType(t, &AgentError{}, agentErr)
	assert.Equal(t, ErrCodeNotFound, agentErr.Code)
	assert.Equal(t, "test error", agentErr.Message)
}

func TestIsErrorCode(t *testing.T) {
	err := New(ErrCodeNotFound, "test error")

	assert.True(t, IsErrorCode(err, ErrCodeNotFound))
	assert.False(t, IsErrorCode(err, ErrCodeInvalidInput))
	assert.False(t, IsErrorCode(nil, ErrCodeNotFound))
	assert.False(t, IsErrorCode(assert.AnError, ErrCodeNotFound))
}