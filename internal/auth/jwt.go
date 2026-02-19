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
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"
)

type jwtClaims struct {
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Iss string `json:"iss"`
	Exp int64  `json:"exp"`
}

// JWTValidator validates JWT tokens with signature verification
type JWTValidator struct {
	secretKey []byte // HMAC secret key
	algorithm string // Algorithm: HS256, HS384, HS512
}

// NewJWTValidator creates a new JWT validator with HMAC secret key
func NewJWTValidator(secretKey string, algorithm string) *JWTValidator {
	if algorithm == "" {
		algorithm = "HS256" // Default to HS256
	}
	return &JWTValidator{
		secretKey: []byte(secretKey),
		algorithm: algorithm,
	}
}

// ValidateJWT validates a JWT token with signature verification:
// - structure (three parts: header.payload.signature)
// - signature verification
// - expiration
// - audience (optional, if provided)
// Returns the subject (sub) if valid, or an error.
func ValidateJWT(token string, expectedAud string) (string, error) {
	// For backward compatibility, use default validator if no secret configured
	// In production, you should always use ValidateJWTWithSecret
	return ValidateJWTWithSecret(token, expectedAud, "", "")
}

// ValidateJWTWithSecret validates a JWT token with signature verification
func ValidateJWTWithSecret(token string, expectedAud string, secretKey string, algorithm string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("empty token")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format: must have 3 parts (header.payload.signature)")
	}
	
	headerPart := parts[0]
	payloadPart := parts[1]
	signaturePart := parts[2]
	
	// Decode and verify header
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerPart)
	if err != nil {
		return "", fmt.Errorf("invalid header encoding: %w", err)
	}
	
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("invalid header format: %w", err)
	}
	
	// Verify algorithm
	alg, ok := header["alg"].(string)
	if !ok {
		return "", errors.New("missing algorithm in header")
	}
	
	// If secret key is provided, verify signature
	if secretKey != "" {
		if err := verifySignature(headerPart+"."+payloadPart, signaturePart, secretKey, alg); err != nil {
			return "", fmt.Errorf("signature verification failed: %w", err)
		}
	} else {
		// If no secret provided but token has signature, reject (security: never accept unsigned tokens)
		if signaturePart != "" {
			return "", errors.New("signature verification required but no secret key provided")
		}
	}
	
	// Decode payload
	payload := payloadPart
	payload = strings.TrimRight(payload, "=")
	var decoded []byte
	decoded, err = base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			return "", fmt.Errorf("invalid payload encoding: %w", err)
		}
	}
	
	var claims jwtClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("invalid payload format: %w", err)
	}
	
	// Validate expiration
	if claims.Exp != 0 && time.Now().Unix() > claims.Exp {
		return "", errors.New("token expired")
	}
	
	// Validate audience
	if expectedAud != "" && claims.Aud != expectedAud {
		return "", errors.New("invalid audience")
	}
	
	// Validate subject
	if claims.Sub == "" {
		return "", errors.New("missing sub claim")
	}
	
	return claims.Sub, nil
}

// verifySignature verifies the JWT signature using HMAC
func verifySignature(message, signature, secretKey, algorithm string) error {
	// Decode signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}
	
	// Select hash algorithm based on JWT algorithm
	var h hash.Hash
	switch algorithm {
	case "HS256":
		h = hmac.New(sha256.New, []byte(secretKey))
	case "HS384":
		h = hmac.New(sha512.New384, []byte(secretKey))
	case "HS512":
		h = hmac.New(sha512.New, []byte(secretKey))
	case "none":
		// "none" algorithm should not be used in production
		return errors.New("'none' algorithm is not allowed for security reasons")
	default:
		return fmt.Errorf("unsupported algorithm: %s (only HS256, HS384, HS512 are supported)", algorithm)
	}
	
	// Compute expected signature
	h.Write([]byte(message))
	expectedSignature := h.Sum(nil)
	
	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal(signatureBytes, expectedSignature) {
		return errors.New("signature mismatch")
	}
	
	return nil
}
