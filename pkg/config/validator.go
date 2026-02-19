// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package config

import (
	"fmt"
	"reflect"
	"strings"
)

// Validator defines the interface for configuration validation
type Validator interface {
	Validate() error
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation failed for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors aggregates multiple validation errors
type ValidationErrors struct {
	Errors []*ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}

	msg := fmt.Sprintf("%d validation error(s):", len(e.Errors))
	for i, err := range e.Errors {
		msg += fmt.Sprintf("\n  %d. %s", i+1, err.Error())
	}
	return msg
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string, value interface{}) {
	e.Errors = append(e.Errors, &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are any validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ToError returns the error or nil if no errors
func (e *ValidationErrors) ToError() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}

// Rule defines a validation rule
type Rule struct {
	Name      string
	Validator func(interface{}) error
	Message   string
}

// Common rules
var (
	// Required rule - value must not be empty
	Required = Rule{
		Name: "required",
		Validator: func(v interface{}) error {
			if v == nil {
				return fmt.Errorf("value is required")
			}

			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.String:
				if rv.String() == "" {
					return fmt.Errorf("string cannot be empty")
				}
			case reflect.Slice, reflect.Map, reflect.Array:
				if rv.Len() == 0 {
					return fmt.Errorf("collection cannot be empty")
				}
			case reflect.Ptr, reflect.Interface:
				if rv.IsNil() {
					return fmt.Errorf("value cannot be nil")
				}
			}
			return nil
		},
		Message: "is required",
	}

	// NonNegative rule - value must be >= 0
	NonNegative = Rule{
		Name: "non_negative",
		Validator: func(v interface{}) error {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if rv.Int() < 0 {
					return fmt.Errorf("value must be non-negative")
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				// Unsigned types are always non-negative
			case reflect.Float32, reflect.Float64:
				if rv.Float() < 0 {
					return fmt.Errorf("value must be non-negative")
				}
			default:
				return fmt.Errorf("unsupported type for non-negative validation")
			}
			return nil
		},
		Message: "must be non-negative",
	}

	// Positive rule - value must be > 0
	Positive = Rule{
		Name: "positive",
		Validator: func(v interface{}) error {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if rv.Int() <= 0 {
					return fmt.Errorf("value must be positive")
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if rv.Uint() == 0 {
					return fmt.Errorf("value must be positive")
				}
			case reflect.Float32, reflect.Float64:
				if rv.Float() <= 0 {
					return fmt.Errorf("value must be positive")
				}
			default:
				return fmt.Errorf("unsupported type for positive validation")
			}
			return nil
		},
		Message: "must be positive",
	}

	// Email rule - validates email format
	Email = Rule{
		Name: "email",
		Validator: func(v interface{}) error {
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("value must be a string")
			}

			// Basic email validation
			if !strings.Contains(str, "@") {
				return fmt.Errorf("invalid email format")
			}

			parts := strings.Split(str, "@")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid email format")
			}

			if !strings.Contains(parts[1], ".") {
				return fmt.Errorf("invalid email format")
			}

			return nil
		},
		Message: "must be a valid email address",
	}

	// URL rule - validates URL format
	URL = Rule{
		Name: "url",
		Validator: func(v interface{}) error {
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("value must be a string")
			}

			if str == "" {
				return fmt.Errorf("URL cannot be empty")
			}

			if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
				return fmt.Errorf("URL must start with http:// or https://")
			}

			return nil
		},
		Message: "must be a valid URL",
	}
)

// Min creates a minimum value rule
func Min(min int) Rule {
	return Rule{
		Name: fmt.Sprintf("min_%d", min),
		Validator: func(v interface{}) error {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if rv.Int() < int64(min) {
					return fmt.Errorf("value must be at least %d", min)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if rv.Uint() < uint64(min) {
					return fmt.Errorf("value must be at least %d", min)
				}
			case reflect.Float32, reflect.Float64:
				if rv.Float() < float64(min) {
					return fmt.Errorf("value must be at least %d", min)
				}
			case reflect.String:
				if len(rv.String()) < min {
					return fmt.Errorf("length must be at least %d", min)
				}
			case reflect.Slice, reflect.Map, reflect.Array:
				if rv.Len() < min {
					return fmt.Errorf("length must be at least %d", min)
				}
			default:
				return fmt.Errorf("unsupported type for min validation")
			}
			return nil
		},
		Message: fmt.Sprintf("must be at least %d", min),
	}
}

// Max creates a maximum value rule
func Max(max int) Rule {
	return Rule{
		Name: fmt.Sprintf("max_%d", max),
		Validator: func(v interface{}) error {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if rv.Int() > int64(max) {
					return fmt.Errorf("value must be at most %d", max)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if rv.Uint() > uint64(max) {
					return fmt.Errorf("value must be at most %d", max)
				}
			case reflect.Float32, reflect.Float64:
				if rv.Float() > float64(max) {
					return fmt.Errorf("value must be at most %d", max)
				}
			case reflect.String:
				if len(rv.String()) > max {
					return fmt.Errorf("length must be at most %d", max)
				}
			case reflect.Slice, reflect.Map, reflect.Array:
				if rv.Len() > max {
					return fmt.Errorf("length must be at most %d", max)
				}
			default:
				return fmt.Errorf("unsupported type for max validation")
			}
			return nil
		},
		Message: fmt.Sprintf("must be at most %d", max),
	}
}

// OneOf creates a rule that checks if value is one of the allowed values
func OneOf(allowed ...interface{}) Rule {
	return Rule{
		Name: "one_of",
		Validator: func(v interface{}) error {
			for _, a := range allowed {
				if v == a {
					return nil
				}
			}
			return fmt.Errorf("value must be one of %v", allowed)
		},
		Message: fmt.Sprintf("must be one of %v", allowed),
	}
}

// Validate validates a configuration struct
func Validate(config interface{}) error {
	errors := &ValidationErrors{}

	rv := reflect.ValueOf(config)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Get validation rules from tag
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Parse rules
		rules := parseRules(tag)

		// Apply rules
		for _, rule := range rules {
			if err := rule.Validator(field.Interface()); err != nil {
				errors.Add(fieldType.Name, err.Error(), field.Interface())
			}
		}
	}

	return errors.ToError()
}

// parseRules parses validation rules from a tag string
func parseRules(tag string) []Rule {
	var rules []Rule

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch part {
		case "required":
			rules = append(rules, Required)
		case "non_negative":
			rules = append(rules, NonNegative)
		case "positive":
			rules = append(rules, Positive)
		case "email":
			rules = append(rules, Email)
		case "url":
			rules = append(rules, URL)
		default:
			// Parse min/max rules
			if strings.HasPrefix(part, "min=") {
				min := parseIntSuffix(part, 4)
				rules = append(rules, Min(min))
			} else if strings.HasPrefix(part, "max=") {
				max := parseIntSuffix(part, 4)
				rules = append(rules, Max(max))
			}
		}
	}

	return rules
}

// parseIntSuffix parses integer suffix from a string
func parseIntSuffix(s string, prefixLen int) int {
	numStr := s[prefixLen:]
	var num int
	fmt.Sscanf(numStr, "%d", &num)
	return num
}

// ValidateStruct validates a struct using reflection
func ValidateStruct(config interface{}) error {
	return Validate(config)
}

// MustValidate validates a configuration and panics if invalid
func MustValidate(config Validator) {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("configuration validation failed: %v", err))
	}
}

// Validatable provides a mixin for configuration validation
type Validatable struct {
	errors *ValidationErrors
}

// AddError adds a validation error
func (v *Validatable) AddError(field, message string, value interface{}) {
	if v.errors == nil {
		v.errors = &ValidationErrors{}
	}
	v.errors.Add(field, message, value)
}

// HasErrors returns true if there are any validation errors
func (v *Validatable) HasErrors() bool {
	return v.errors != nil && v.errors.HasErrors()
}

// Error returns the validation error or nil
func (v *Validatable) Error() error {
	if v.errors == nil {
		return nil
	}
	return v.errors.ToError()
}

// RequireString validates that a string field is not empty
func RequireString(field, value string) error {
	if value == "" {
		return &ValidationError{
			Field:   field,
			Message: "cannot be empty",
			Value:   value,
		}
	}
	return nil
}

// RequireInt validates that an int field is positive
func RequireInt(field string, value int) error {
	if value <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be positive",
			Value:   value,
		}
	}
	return nil
}

// RequireDuration validates that a duration is positive
func RequireDuration(field string, value int) error {
	if value <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "duration must be positive",
			Value:   value,
		}
	}
	return nil
}
