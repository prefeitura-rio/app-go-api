package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// JWTClaims represents the claims we care about from the JWT token
type JWTClaims struct {
	PreferredUsername string `json:"preferred_username"`
	// Add other claims if needed in the future
}

// ExtractCPFFromToken decodes a JWT token and extracts the CPF from preferred_username
// No signature validation is performed (already validated by gateway)
// Expected format: "Bearer <token>" or just "<token>"
func ExtractCPFFromToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		return "", errors.New("token is empty after removing Bearer prefix")
	}

	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format: expected 3 parts")
	}

	// Decode the payload (second part)
	payload := parts[1]

	// Decode base64 using both RawURLEncoding and StdEncoding as fallback
	var decoded []byte
	var err error

	// Try RawURLEncoding first (no padding)
	decoded, err = base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		// Try with standard encoding (with padding)
		decoded, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			return "", errors.New("failed to decode JWT payload: " + err.Error())
		}
	}

	// Parse JSON
	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", errors.New("failed to parse JWT claims: " + err.Error())
	}

	// Extract CPF from preferred_username
	if claims.PreferredUsername == "" {
		return "", errors.New("preferred_username not found in JWT claims")
	}

	return claims.PreferredUsername, nil
}
