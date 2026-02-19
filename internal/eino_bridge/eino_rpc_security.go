// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package einobridge

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// ValidateJWT is a lightweight JWT payload extractor (no signature verification in MVP).
// It returns the subject (sub) if present.
func ValidateJWT(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", errors.New("empty token")
	}
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid JWT format")
	}
	payload := parts[1]
	// Base64 decode payload (URL-safe, without padding)
	payload = strings.TrimRight(payload, "=")
	b, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		b, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			return "", err
		}
	}
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", err
	}
	if sub, ok := data["sub"].(string); ok {
		return sub, nil
	}
	// fallback to 'user' claim if present
	if user, ok := data["user"].(string); ok {
		return user, nil
	}
	return "", nil
}
