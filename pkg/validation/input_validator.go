// Agent Framework - Input Validation System
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// InputValidator provides comprehensive input validation
type InputValidator struct {
	maxLength     int
	allowedChars  *regexp.Regexp
	blockPatterns []string
	sanitizeHTML  bool
}

// ValidatorConfig configures an InputValidator
type ValidatorConfig struct {
	MaxLength     int
	AllowedChars  string
	BlockPatterns []string
	SanitizeHTML  bool
}

// NewInputValidator creates a new InputValidator with the given configuration
func NewInputValidator(config ValidatorConfig) (*InputValidator, error) {
	var allowedCharsRegex *regexp.Regexp
	var err error

	if config.AllowedChars != "" {
		allowedCharsRegex, err = regexp.Compile("^" + config.AllowedChars + "$")
		if err != nil {
			return nil, fmt.Errorf("invalid allowed chars pattern: %w", err)
		}
	}

	return &InputValidator{
		maxLength:     config.MaxLength,
		allowedChars:  allowedCharsRegex,
		blockPatterns: config.BlockPatterns,
		sanitizeHTML:  config.SanitizeHTML,
	}, nil
}

// StringValidator creates a validator for string inputs
func StringValidator(maxLength int) *InputValidator {
	return RequiredStringValidation(maxLength)
}

// RequiredStringValidation creates a validator for required string inputs
func RequiredStringValidation(maxLength int) *InputValidator {
	validator, _ := NewInputValidator(ValidatorConfig{
		MaxLength: maxLength,
		BlockPatterns: []string{
			"<script",
			"javascript:",
			"onerror=",
			"onclick=",
			"onload=",
		},
		SanitizeHTML: true,
	})
	return validator
}

// IDValidator creates a validator for ID fields
func IDValidator() *InputValidator {
	validator, _ := NewInputValidator(ValidatorConfig{
		MaxLength:    100,
		AllowedChars: "[a-zA-Z0-9-_]+",
	})
	return validator
}

// PathValidator creates a validator for file paths
func PathValidator() *InputValidator {
	validator, _ := NewInputValidator(ValidatorConfig{
		MaxLength: 500,
		BlockPatterns: []string{
			"../",
			"..\\",
			"/etc/",
			"\\windows\\",
		},
	})
	return validator
}

// EmailValidator creates a validator for email addresses
func EmailValidator() *InputValidator {
	validator, _ := NewInputValidator(ValidatorConfig{
		MaxLength:    254,
		AllowedChars: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
	})
	return validator
}

// UsernameValidator creates a validator for usernames
func UsernameValidator() *InputValidator {
	validator, _ := NewInputValidator(ValidatorConfig{
		MaxLength:    50,
		AllowedChars: "[a-zA-Z0-9_-]+",
	})
	return validator
}

// ValidateAndSanitize validates and sanitizes user input
func (v *InputValidator) ValidateAndSanitize(input string) (string, error) {
	// Check max length
	if v.maxLength > 0 && len(input) > v.maxLength {
		return "", fmt.Errorf("input exceeds maximum length of %d", v.maxLength)
	}

	// Check for blocked patterns
	for _, pattern := range v.blockPatterns {
		if strings.Contains(strings.ToLower(input), strings.ToLower(pattern)) {
			return "", fmt.Errorf("input contains blocked pattern: %s", pattern)
		}
	}

	// Check allowed characters
	if v.allowedChars != nil {
		if !v.allowedChars.MatchString(input) {
			return "", fmt.Errorf("input contains invalid characters")
		}
	}

	// Sanitize HTML if enabled
	sanitized := input
	if v.sanitizeHTML {
		sanitized = v.sanitizeHTMLString(input)
	}

	return sanitized, nil
}

// sanitizeHTMLString removes or escapes dangerous HTML content
func (v *InputValidator) sanitizeHTMLString(input string) string {
	// Simple HTML sanitization - replace < and >
	sanitized := strings.ReplaceAll(input, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
	sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
	return sanitized
}

// Validate validates input without sanitization
func (v *InputValidator) Validate(input string) error {
	_, err := v.ValidateAndSanitize(input)
	return err
}
