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
	"errors"
	"fmt"
)

// Handler provides centralized error handling
type Handler struct {
	prefix string
	logger Logger
}

// Logger defines the logging interface
type Logger interface {
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

// NewHandler creates a new error handler
func NewHandler(prefix string, logger Logger) *Handler {
	return &Handler{
		prefix: prefix,
		logger: logger,
	}
}

// Handle wraps an error with operation context
func (h *Handler) Handle(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Log the error
	if h.logger != nil {
		h.logger.Errorf("[%s] %s: %v", h.prefix, operation, err)
	}

	// Wrap with context
	return fmt.Errorf("%s.%s: %w", h.prefix, operation, err)
}

// Handlef wraps an error with formatted context
func (h *Handler) Handlef(format string, args ...interface{}) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}

		message := fmt.Sprintf(format, args...)

		if h.logger != nil {
			h.logger.Errorf("[%s] %s: %v", h.prefix, message, err)
		}

		return fmt.Errorf("%s.%s: %w", h.prefix, message, err)
	}
}

// Wrap wraps an error with additional context
func (h *Handler) Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted message
func (h *Handler) Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

// New creates a new error with context
func (h *Handler) New(message string) error {
	if h.logger != nil {
		h.logger.Errorf("[%s] %s", h.prefix, message)
	}

	return fmt.Errorf("%s.%s", h.prefix, message)
}

// Newf creates a new error with formatted message
func (h *Handler) Newf(format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)

	if h.logger != nil {
		h.logger.Errorf("[%s] %s", h.prefix, message)
	}

	return fmt.Errorf("%s.%s", h.prefix, message)
}

// Is checks if an error is of a specific type
func (h *Handler) Is(err, target error) bool {
	return errors.Is(err, target)
}

// As checks if an error can be cast to a specific type
func (h *Handler) As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap returns the underlying error
func (h *Handler) Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Validate checks if an error is nil and returns it if not
func (h *Handler) Validate(err error) error {
	if err == nil {
		return nil
	}

	return h.Handle("validation", err)
}

// Validatef validates a condition and returns an error if false
func (h *Handler) Validatef(condition bool, format string, args ...interface{}) error {
	if condition {
		return nil
	}

	return h.Newf(format, args...)
}

// Recover recovers from a panic and converts it to an error
func (h *Handler) Recover() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return h.Handle("panic", v)
		default:
			return h.Newf("panic: %v", v)
		}
	}
	return nil
}

// RecoverWith recovers from a panic and executes a handler function
func (h *Handler) RecoverWith(handler func(interface{}) error) error {
	if r := recover(); r != nil {
		if handler != nil {
			return handler(r)
		}
		return h.Recover()
	}
	return nil
}

// Common error helpers

// NotFound creates a "not found" error
func (h *Handler) NotFound(resource, id string) error {
	return h.Newf("%s not found: %s", resource, id)
}

// AlreadyExists creates an "already exists" error
func (h *Handler) AlreadyExists(resource, id string) error {
	return h.Newf("%s already exists: %s", resource, id)
}

// InvalidInput creates an "invalid input" error
func (h *Handler) InvalidInput(field, reason string) error {
	return h.Newf("invalid input for %s: %s", field, reason)
}

// Unauthorized creates an "unauthorized" error
func (h *Handler) Unauthorized(reason string) error {
	return h.Newf("unauthorized: %s", reason)
}

// Forbidden creates a "forbidden" error
func (h *Handler) Forbidden(reason string) error {
	return h.Newf("forbidden: %s", reason)
}

// Internal creates an "internal error" error
func (h *Handler) Internal(err error) error {
	return h.Handle("internal", err)
}

// Retryable wraps an error to indicate it's retryable
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Retryable marks an error as retryable
func (h *Handler) Retryable(err error) error {
	if err == nil {
		return nil
	}

	return &RetryableError{Err: err}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Operation string
	Timeout   int
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation %s timed out after %d seconds", e.Operation, e.Timeout)
}

// Timeout creates a timeout error
func (h *Handler) Timeout(operation string, timeout int) error {
	return &TimeoutError{
		Operation: operation,
		Timeout:   timeout,
	}
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	var timeout *TimeoutError
	return errors.As(err, &timeout)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Validation creates a validation error
func (h *Handler) Validation(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	var validation *ValidationError
	return errors.As(err, &validation)
}

// Aggregate aggregates multiple errors
type AggregateError struct {
	Errors []error
}

func (e *AggregateError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	msg := fmt.Sprintf("%d errors occurred:", len(e.Errors))
	for i, err := range e.Errors {
		msg += fmt.Sprintf("\n  %d. %v", i+1, err)
	}
	return msg
}

// Aggregate aggregates multiple errors into one
func (h *Handler) Aggregate(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}

	if len(nonNil) == 0 {
		return nil
	}

	if len(nonNil) == 1 {
		return nonNil[0]
	}

	return &AggregateError{Errors: nonNil}
}

// IsAggregate checks if an error is an aggregate error
func IsAggregate(err error) bool {
	var aggregate *AggregateError
	return errors.As(err, &aggregate)
}

// FirstError returns the first non-nil error from a list
func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// Join joins multiple errors into one
func Join(errs ...error) error {
	return JoinWith("; ", errs...)
}

// JoinWith joins multiple errors with a custom separator
func JoinWith(sep string, errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}

	if len(nonNil) == 0 {
		return nil
	}

	if len(nonNil) == 1 {
		return nonNil[0]
	}

	msg := ""
	for i, err := range nonNil {
		if i > 0 {
			msg += sep
		}
		msg += err.Error()
	}

	return errors.New(msg)
}
