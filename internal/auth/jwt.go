package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type jwtClaims struct {
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Iss string `json:"iss"`
	Exp int64  `json:"exp"`
}

// ValidateJWT validates a JWT token payload with minimal checks:
// - structure (three parts)
// - expiration
// - audience (optional, if provided)
// Returns the subject (sub) if valid, or an error.
func ValidateJWT(token string, expectedAud string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("empty token")
	}
	parts := strings.Split(token, ".")
	if len(parts) < 3 {
		return "", errors.New("invalid JWT format")
	}
	payload := parts[1]
	// decode payload (URL-safe, without padding)
	payload = strings.TrimRight(payload, "=")
	var decoded []byte
	var err error
	decoded, err = base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			return "", err
		}
	}
	var claims jwtClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", err
	}
	if claims.Exp != 0 && time.Now().Unix() > claims.Exp {
		return "", errors.New("token expired")
	}
	if expectedAud != "" && claims.Aud != expectedAud {
		return "", errors.New("invalid audience")
	}
	if claims.Sub == "" {
		return "", errors.New("missing sub claim")
	}
	return claims.Sub, nil
}
