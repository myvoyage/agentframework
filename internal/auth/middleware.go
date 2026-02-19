// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const jwtSubjectKey contextKey = "jwt_subject"

// GetJWTSubject retrieves the JWT subject from context
func GetJWTSubject(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(jwtSubjectKey).(string)
	return subject, ok
}

type JWTMiddleware struct {
	Audience  string
	SecretKey string // HMAC secret key for signature verification
	Algorithm string // JWT algorithm (HS256, HS384, HS512)
}

// Middleware wraps an http.Handler and enforces JWT Authorization header validation.
func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		subject, err := ValidateJWTWithSecret(token, m.Audience, m.SecretKey, m.Algorithm)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		// propagate context with subject
		ctx := context.WithValue(r.Context(), jwtSubjectKey, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
