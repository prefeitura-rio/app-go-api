package utils

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestExtractCPFFromToken(t *testing.T) {
	// Helper to create a fake JWT token
	createToken := func(preferredUsername string) string {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

		claims := map[string]interface{}{
			"preferred_username": preferredUsername,
			"sub":                "1234567890",
			"name":               "Test User",
		}
		claimsJSON, _ := json.Marshal(claims)
		payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

		signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

		return header + "." + payload + "." + signature
	}

	tests := []struct {
		name        string
		authHeader  string
		expectedCPF string
		shouldError bool
	}{
		{
			"Valid token with Bearer prefix",
			"Bearer " + createToken("12345678900"),
			"12345678900",
			false,
		},
		{
			"Valid token without Bearer prefix",
			createToken("98765432100"),
			"98765432100",
			false,
		},
		{
			"Empty auth header",
			"",
			"",
			true,
		},
		{
			"Only Bearer without token",
			"Bearer ",
			"",
			true,
		},
		{
			"Invalid JWT format (only 2 parts)",
			"header.payload",
			"",
			true,
		},
		{
			"Invalid JWT format (only 1 part)",
			"invalid-token",
			"",
			true,
		},
		{
			"Invalid base64 in payload",
			"header.@@@invalid-base64@@@.signature",
			"",
			true,
		},
		{
			"Valid token but missing preferred_username",
			"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.fake",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpf, err := ExtractCPFFromToken(tt.authHeader)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if cpf != tt.expectedCPF {
					t.Errorf("ExtractCPFFromToken() = %q, want %q", cpf, tt.expectedCPF)
				}
			}
		})
	}
}
