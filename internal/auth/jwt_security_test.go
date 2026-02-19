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
	"testing"
)

// TestJWTSecurity verifies JWT security enhancements
func TestJWTSecurity(t *testing.T) {
	secretKey := "test-secret-key-for-testing"

	tests := []struct {
		name        string
		token       string
		secretKey   string
		algorithm   string
		expectedAud string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid token with signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			secretKey:   "your-256-bit-secret",
			algorithm:   "HS256",
			expectedAud: "",
			wantErr:     false,
		},
		{
			name:        "reject token without secret key",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.",
			secretKey:   "",
			algorithm:   "",
			expectedAud: "",
			wantErr:     true,
			errContains: "requires a secret key",
		},
		{
			name:        "reject none algorithm",
			token:       "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIn0.",
			secretKey:   secretKey,
			algorithm:   "none",
			expectedAud: "",
			wantErr:     true,
			errContains: "not permitted",
		},
		{
			name:        "reject empty token",
			token:       "",
			secretKey:   secretKey,
			algorithm:   "HS256",
			expectedAud: "",
			wantErr:     true,
			errContains: "empty token",
		},
		{
			name:        "reject invalid format",
			token:       "invalid.token",
			secretKey:   secretKey,
			algorithm:   "HS256",
			expectedAud: "",
			wantErr:     true,
			errContains: "invalid JWT format",
		},
		{
			name:        "reject unsupported algorithm",
			token:       "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.",
			secretKey:   secretKey,
			algorithm:   "RS256",
			expectedAud: "",
			wantErr:     true,
			errContains: "unsupported algorithm",
		},
		{
			name:        "validate expired token",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZXhwIjoxNTE2MjM5MDIyfQ.",
			secretKey:   secretKey,
			algorithm:   "HS256",
			expectedAud: "",
			wantErr:     true,
			errContains: "token expired",
		},
		{
			name:        "validate audience",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiYXVkIjoidGVzdC1hdWRpZW5jZSJ9.",
			secretKey:   secretKey,
			algorithm:   "HS256",
			expectedAud: "different-audience",
			wantErr:     true,
			errContains: "invalid audience",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWTWithSecret(tt.token, tt.expectedAud, tt.secretKey, tt.algorithm)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateJWTWithSecret() expected error containing %q, but got no error", tt.errContains)
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateJWTWithSecret() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateJWTWithSecret() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateJWT_Deprecated verifies that the old ValidateJWT function now rejects calls
func TestValidateJWT_Deprecated(t *testing.T) {
	_, err := ValidateJWT("any-token", "any-audience")
	if err == nil {
		t.Error("ValidateJWT() should return error as it is deprecated")
	}
	if !containsString(err.Error(), "deprecated") {
		t.Errorf("ValidateJWT() error should mention deprecation, got = %v", err)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsStringHelper(s, substr)))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkJWTValidation benchmarks JWT validation performance
func BenchmarkJWTValidation(b *testing.B) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	secretKey := "your-256-bit-secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateJWTWithSecret(token, "", secretKey, "HS256")
	}
}
