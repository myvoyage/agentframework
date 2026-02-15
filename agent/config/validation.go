// Agent Framework - Enhanced Configuration Validation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"AgentFramework/agent/errors"
)

// ValidationRuleType defines the type of validation rule
type ValidationRuleType string

const (
	// RuleTypeRequired checks if a value is present
	RuleTypeRequired ValidationRuleType = "required"
	// RuleTypeMin checks minimum value (for numbers)
	RuleTypeMin ValidationRuleType = "min"
	// RuleTypeMax checks maximum value (for numbers)
	RuleTypeMax ValidationRuleType = "max"
	// RuleTypeMinLength checks minimum length (for strings, arrays)
	RuleTypeMinLength ValidationRuleType = "min_length"
	// RuleTypeMaxLength checks maximum length (for strings, arrays)
	RuleTypeMaxLength ValidationRuleType = "max_length"
	// RuleTypePattern checks regex pattern (for strings)
	RuleTypePattern ValidationRuleType = "pattern"
	// RuleTypeEmail validates email format
	RuleTypeEmail ValidationRuleType = "email"
	// RuleTypeURL validates URL format
	RuleTypeURL ValidationRuleType = "url"
	// RuleTypeEnum checks if value is in enum
	RuleTypeEnum ValidationRuleType = "enum"
	// RuleTypeOneOf checks if value is one of allowed values
	RuleTypeOneOf ValidationRuleType = "oneof"
	// RuleTypeRange checks numeric range
	RuleTypeRange ValidationRuleType = "range"
	// RuleTypeCustom uses custom validator function
	RuleTypeCustom ValidationRuleType = "custom"
)

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Type     ValidationRuleType `json:"type"`
	Value    interface{}        `json:"value,omitempty"`
	Message  string             `json:"message,omitempty"`
	Validator CustomValidatorFunc `json:"-"` // For custom type
}

// CustomValidatorFunc is a function that validates a value
type CustomValidatorFunc func(key string, value interface{}) error

// ValidationError represents a validation error with context
type ValidationError struct {
	Key       string        `json:"key"`
	Rule      ValidationRule `json:"rule"`
	Value      interface{}   `json:"value"`
	Message    string        `json:"message"`
	Timestamp  time.Time     `json:"timestamp"`
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Key, e.Message)
	}
	return fmt.Sprintf("validation failed for %s", e.Key)
}

// ValidationResult contains validation results
type ValidationResult struct {
	IsValid  bool              `json:"is_valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidatorChain chains multiple validators together
type ValidatorChain struct {
	validators []ConfigValidatorFunc
	all        bool // If true, all validators must pass; if false, any validator can pass
}

// NewValidatorChain creates a new validator chain
func NewValidatorChain(all bool, validators ...ConfigValidatorFunc) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		results := make([]error, 0, len(validators))
		for _, validator := range validators {
			if err := validator(key, value); err != nil {
				results = append(results, err)
			}
		}

		// If all validators must pass
		if all && len(results) > 0 {
			return errors.Newf(errors.ErrCodeInvalidInput,
				"all validators failed: %v", results)
		}

		// If any validator can pass
		if !all && len(results) == len(validators) {
			return errors.Newf(errors.ErrCodeInvalidInput,
				"all validators failed: %v", results)
		}

		return nil
	}
}

// Required creates a required validator
func Required() ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		if value == nil {
			return &ValidationError{
				Key:      key,
				Rule:     ValidationRule{Type: RuleTypeRequired},
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Min creates a minimum value validator
func Min(min interface{}) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		num, ok := toFloat64(value)
		minVal, ok2 := toFloat64(min)
		if !ok || !ok2 {
			return nil // Skip if not numeric
		}

		if num < minVal {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type:  RuleTypeMin,
					Value: min,
				},
				Value:     value,
				Message:   fmt.Sprintf("value must be at least %v", min),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Max creates a maximum value validator
func Max(max interface{}) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		num, ok := toFloat64(value)
		maxVal, ok2 := toFloat64(max)
		if !ok || !ok2 {
			return nil // Skip if not numeric
		}

		if num > maxVal {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type:  RuleTypeMax,
					Value: max,
				},
				Value:     value,
				Message:   fmt.Sprintf("value must be at most %v", max),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// MinLength creates a minimum length validator
func MinLength(min int) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		length := 0
		switch v := value.(type) {
		case string:
			length = len(v)
		case []interface{}:
			length = len(v)
		case []string:
			length = len(v)
		default:
			return nil // Skip if not applicable
		}

		if length < min {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type:  RuleTypeMinLength,
					Value: min,
				},
				Value:     value,
				Message:   fmt.Sprintf("length must be at least %d", min),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// MaxLength creates a maximum length validator
func MaxLength(max int) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		length := 0
		switch v := value.(type) {
		case string:
			length = len(v)
		case []interface{}:
			length = len(v)
		case []string:
			length = len(v)
		default:
			return nil // Skip if not applicable
		}

		if length > max {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type:  RuleTypeMaxLength,
					Value: max,
				},
				Value:     value,
				Message:   fmt.Sprintf("length must be at most %d", max),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Pattern creates a regex pattern validator
func Pattern(pattern string) ConfigValidatorFunc {
	regex := regexp.MustCompile(pattern)
	return func(key string, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return nil // Skip if not string
		}

		if !regex.MatchString(str) {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type: RuleTypePattern,
					Value: pattern,
				},
				Value:     value,
				Message:   fmt.Sprintf("value does not match pattern %s", pattern),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Email creates an email validator
func Email() ConfigValidatorFunc {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return func(key string, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return nil // Skip if not string
		}

		if !pattern.MatchString(str) {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{Type: RuleTypeEmail},
				Value:     value,
				Message:   "invalid email format",
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// URL creates a URL validator
func URL() ConfigValidatorFunc {
	pattern := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(:\d+)?(/.*)?$`)
	return func(key string, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return nil // Skip if not string
		}

		if !pattern.MatchString(str) {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{Type: RuleTypeURL},
				Value:     value,
				Message:   "invalid URL format",
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// OneOf creates an enum validator
func OneOf(allowedValues ...interface{}) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		for _, allowed := range allowedValues {
			if value == allowed {
				return nil
			}
		}

		return &ValidationError{
			Key:   key,
			Rule:  ValidationRule{
				Type:  RuleTypeOneOf,
				Value: allowedValues,
			},
			Value:     value,
			Message:   fmt.Sprintf("value must be one of %v", allowedValues),
			Timestamp: time.Now(),
		}
	}
}

// Range creates a numeric range validator
func Range(min, max float64) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		num, ok := toFloat64(value)
		if !ok {
			return nil // Skip if not numeric
		}

		if num < min || num > max {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type: RuleTypeRange,
					Value: []interface{}{min, max},
				},
				Value:     value,
				Message:   fmt.Sprintf("value must be between %.2f and %.2f", min, max),
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Custom creates a custom validator
func Custom(validator CustomValidatorFunc, message string) ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		if err := validator(key, value); err != nil {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type: RuleTypeCustom,
				},
				Value:     value,
				Message:   message,
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// StringNotEmptyValidator creates a non-empty string validator
func StringNotEmptyValidator() ConfigValidatorFunc {
	return func(key string, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return nil // Skip if not string
		}

		if strings.TrimSpace(str) == "" {
			return &ValidationError{
				Key:   key,
				Rule:  ValidationRule{
					Type: RuleTypeCustom,
				},
				Value:     value,
				Message:   "string cannot be empty",
				Timestamp: time.Now(),
			}
		}
		return nil
	}
}

// Compose creates a composed validator from multiple rules
func Compose(all bool, rules ...ValidationRule) ConfigValidatorFunc {
	validators := make([]ConfigValidatorFunc, 0, len(rules))
	for _, rule := range rules {
		validators = append(validators, createValidatorFromRule(rule))
	}
	return NewValidatorChain(all, validators...)
}

// createValidatorFromRule creates a validator function from a rule
func createValidatorFromRule(rule ValidationRule) ConfigValidatorFunc {
	switch rule.Type {
	case RuleTypeRequired:
		return Required()
	case RuleTypeMin:
		return Min(rule.Value)
	case RuleTypeMax:
		return Max(rule.Value)
	case RuleTypeMinLength:
		if min, ok := rule.Value.(int); ok {
			return MinLength(min)
		}
	case RuleTypeMaxLength:
		if max, ok := rule.Value.(int); ok {
			return MaxLength(max)
		}
	case RuleTypePattern:
		if pattern, ok := rule.Value.(string); ok {
			return Pattern(pattern)
		}
	case RuleTypeEmail:
		return Email()
	case RuleTypeURL:
		return URL()
	case RuleTypeOneOf:
		if values, ok := rule.Value.([]interface{}); ok {
			return OneOf(values...)
		}
	case RuleTypeCustom:
		if rule.Validator != nil {
			return Custom(rule.Validator, rule.Message)
		}
	default:
		// Return a no-op validator for unknown types
		return func(key string, value interface{}) error {
			return nil
		}
	}
	return func(key string, value interface{}) error {
		return nil
	}
}

// toFloat64 converts a value to float64 if possible
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// ValidateConfig validates a configuration value against multiple rules
func ValidateConfig(key string, value interface{}, rules ...ValidationRule) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]ValidationError, 0),
	}

	for _, rule := range rules {
		validator := createValidatorFromRule(rule)
		if err := validator(key, value); err != nil {
			result.IsValid = false
			if validationErr, ok := err.(*ValidationError); ok {
				result.Errors = append(result.Errors, *validationErr)
			} else {
				result.Errors = append(result.Errors, ValidationError{
					Key:   key,
					Rule:  rule,
					Value:  value,
					Message: err.Error(),
				})
			}
		}
	}

	return result
}

// ValidateConfigMap validates a map of configuration values
func ValidateConfigMap(config map[string]interface{}, rules map[string][]ValidationRule) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]ValidationError, 0),
	}

	for key, value := range config {
		keyRules, exists := rules[key]
		if !exists {
			continue
		}

		for _, rule := range keyRules {
			validator := createValidatorFromRule(rule)
			if err := validator(key, value); err != nil {
				result.IsValid = false
				if validationErr, ok := err.(*ValidationError); ok {
					result.Errors = append(result.Errors, *validationErr)
				} else {
					result.Errors = append(result.Errors, ValidationError{
						Key:   key,
						Rule:  rule,
						Value:  value,
						Message: err.Error(),
					})
				}
			}
		}
	}

	return result
}
