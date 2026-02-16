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

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import "fmt"

// SkillErrorCode defines error codes for skill operations
type SkillErrorCode string

// Predefined skill error codes
const (
	ErrCodeSkillNotFound       SkillErrorCode = "ERR_SKILL_NOT_FOUND"
	ErrCodeSkillAlreadyExists  SkillErrorCode = "ERR_SKILL_ALREADY_EXISTS"
	ErrCodeSkillDisabled       SkillErrorCode = "ERR_SKILL_DISABLED"
	ErrCodeSkillExecution      SkillErrorCode = "ERR_SKILL_EXECUTION_FAILED"
	ErrCodeSkillTimeout        SkillErrorCode = "ERR_SKILL_TIMEOUT"
	ErrCodeSkillInvalidConfig  SkillErrorCode = "ERR_SKILL_INVALID_CONFIG"
	ErrCodeSkillNotRegistered  SkillErrorCode = "ERR_SKILL_NOT_REGISTERED"
	ErrCodeSkillImportFailed  SkillErrorCode = "ERR_SKILL_IMPORT_FAILED"
	ErrCodeSkillExportFailed  SkillErrorCode = "ERR_SKILL_EXPORT_FAILED"
	ErrCodePipelineNotFound    SkillErrorCode = "ERR_PIPELINE_NOT_FOUND"
	ErrCodePipelineExecution  SkillErrorCode = "ERR_PIPELINE_EXECUTION_FAILED"
)

// SkillError is a structured error for skill operations
type SkillError struct {
	Code    SkillErrorCode
	Message string
	Detail  error
	SkillID string // Optional: the skill that caused the error
}

func (e *SkillError) Error() string {
	if e == nil {
		return "<nil skill error>"
	}
	base := fmt.Sprintf("skill error [%s]: %s", e.Code, e.Message)
	if e.SkillID != "" {
		base += fmt.Sprintf(" (skill: %s)", e.SkillID)
	}
	if e.Detail != nil {
		base += fmt.Sprintf(": %v", e.Detail)
	}
	return base
}

func (e *SkillError) Unwrap() error {
	return e.Detail
}

// NewSkillError creates a new SkillError
func NewSkillError(code SkillErrorCode, message string, detail error) *SkillError {
	return &SkillError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// NewSkillErrorWithID creates a new SkillError with skill ID
func NewSkillErrorWithID(code SkillErrorCode, message string, skillID string, detail error) *SkillError {
	return &SkillError{
		Code:    code,
		Message: message,
		Detail:  detail,
		SkillID: skillID,
	}
}

// IsSkillNotFound checks if error is a skill not found error
func IsSkillNotFound(err error) bool {
	if se, ok := err.(*SkillError); ok {
		return se.Code == ErrCodeSkillNotFound
	}
	return false
}

// IsSkillDisabled checks if error is a skill disabled error
func IsSkillDisabled(err error) bool {
	if se, ok := err.(*SkillError); ok {
		return se.Code == ErrCodeSkillDisabled
	}
	return false
}

// IsSkillExecutionError checks if error is a skill execution error
func IsSkillExecutionError(err error) bool {
	if se, ok := err.(*SkillError); ok {
		return se.Code == ErrCodeSkillExecution || se.Code == ErrCodeSkillTimeout
	}
	return false
}
