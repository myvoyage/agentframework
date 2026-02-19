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

type JWTMiddleware struct {
	Audience string
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
		if _, err := ValidateJWT(token, m.Audience); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// propagate context with a placeholder subject if needed
		ctx := context.WithValue(r.Context(), "jwt_subject", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
