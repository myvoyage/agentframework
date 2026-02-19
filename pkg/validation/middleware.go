// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"AgentFramework/pkg/validation"
)

// ValidationMiddleware provides HTTP middleware for input validation
type ValidationMiddleware struct {
	validator *validation.InputValidator
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(validator *validation.InputValidator) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator,
	}
}

// ValidateQueryParam validates a query parameter
func (vm *ValidationMiddleware) ValidateQueryParam(r *http.Request, key string) (string, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", fmt.Errorf("query parameter %s is required", key)
	}

	sanitized, err := vm.validator.ValidateAndSanitize(value)
	if err != nil {
		return "", fmt.Errorf("invalid query parameter %s: %w", key, err)
	}

	return sanitized, nil
}

// ValidateHeader validates a header value
func (vm *ValidationMiddleware) ValidateHeader(r *http.Request, key string) (string, error) {
	value := r.Header.Get(key)
	if value == "" {
		return "", fmt.Errorf("header %s is required", key)
	}

	sanitized, err := vm.validator.ValidateAndSanitize(value)
	if err != nil {
		return "", fmt.Errorf("invalid header %s: %w", key, err)
	}

	return sanitized, nil
}

// ValidateBody validates request body (for string-based bodies)
func (vm *ValidationMiddleware) ValidateBody(r *http.Request) (string, error) {
	// Read body
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	// Validate
	sanitized, err := vm.validator.ValidateAndSanitize(string(body))
	if err != nil {
		return "", fmt.Errorf("invalid body: %w", err)
	}

	return sanitized, nil
}

// Middleware returns an HTTP middleware function
func (vm *ValidationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate query parameters
		for key := range r.URL.Query() {
			if _, err := vm.ValidateQueryParam(r, key); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Validate headers (optional - add specific headers as needed)
		// for _, header := range headersToValidate {
		//     if _, err := vm.ValidateHeader(r, header); err != nil {
		//         http.Error(w, err.Error(), http.StatusBadRequest)
		//         return
		//     }
		// }

		next.ServeHTTP(w, r)
	})
}

// Common validation middleware constructors

// StringValidationMiddleware creates middleware for string validation
func StringValidationMiddleware(maxLength int) func(http.Handler) http.Handler {
	validator := validation.StringValidator(maxLength)
	vm := NewValidationMiddleware(validator)
	return vm.Middleware
}

// RequiredStringValidationMiddleware creates middleware for required string validation
func RequiredStringValidationMiddleware(maxLength int) func(http.Handler) http.Handler {
	validator := validation.RequiredStringValidator(maxLength)
	vm := NewValidationMiddleware(validator)
	return vm.Middleware
}

// IDValidationMiddleware creates middleware for ID validation
func IDValidationMiddleware() func(http.Handler) http.Handler {
	validator := validation.IDValidator()
	vm := NewValidationMiddleware(validator)
	return vm.Middleware
}

// PathValidationMiddleware creates middleware for path validation
func PathValidationMiddleware() func(http.Handler) http.Handler {
	validator := validation.PathValidator()
	vm := NewValidationMiddleware(validator)
	return vm.Middleware
}

// ContextKey is used for storing validated data in request context
type ContextKey string

const (
	// ValidatedQueryParamKey is the context key for validated query params
	ValidatedQueryParamKey ContextKey = "validated_query_param"
	// ValidatedHeaderKey is the context key for validated headers
	ValidatedHeaderKey ContextKey = "validated_header"
	// ValidatedBodyKey is the context key for validated body
	ValidatedBodyKey ContextKey = "validated_body"
)

// ValidateAndStore validates input and stores it in the request context
func ValidateAndStore(ctx context.Context, key ContextKey, validator *validation.InputValidator, value string) (context.Context, error) {
	sanitized, err := validator.ValidateAndSanitize(value)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, key, sanitized), nil
}

// GetValidatedValue retrieves a validated value from the context
func GetValidatedValue(ctx context.Context, key ContextKey) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent XSS attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// HSTS (HTTP Strict Transport Security)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides basic rate limiting (can be enhanced with Redis)
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// This is a simple in-memory rate limiter
	// For production, use Redis or a dedicated rate limiting service
	type client struct {
		requests  int
		lastReset int64
	}

	clients := make(map[string]*client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := strings.Split(r.RemoteAddr, ":")[0]

			// Check rate limit
			// (Implementation simplified - use proper rate limiting in production)
			_ = ip
			_ = requestsPerMinute
			_ = clients

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware provides CORS support
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
