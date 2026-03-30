package utils

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// createToken is a helper to create a fake JWT token for testing
func createToken(preferredUsername string) string {
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

func TestExtractCPFFromToken(t *testing.T) {

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
		{
			"Token with standard base64 encoding (with padding)",
			"Bearer " + createTokenWithPadding("11122233344"),
			"11122233344",
			false,
		},
		{
			"Token with whitespace",
			"  Bearer   " + createToken("12345678900") + "  ",
			"12345678900",
			false,
		},
		{
			"Token with only whitespace after Bearer",
			"Bearer    ",
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

// createTokenWithPadding creates a JWT with standard base64 encoding (with padding)
func createTokenWithPadding(preferredUsername string) string {
	header := base64.StdEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := map[string]interface{}{
		"preferred_username": preferredUsername,
		"sub":                "1234567890",
		"name":               "Test User",
	}
	claimsJSON, _ := json.Marshal(claims)
	payload := base64.StdEncoding.EncodeToString(claimsJSON)

	signature := base64.StdEncoding.EncodeToString([]byte("fake-signature"))

	return header + "." + payload + "." + signature
}

func TestExtractCPFFromToken_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		claims      map[string]interface{}
		shouldError bool
	}{
		{
			"Empty preferred_username",
			map[string]interface{}{"preferred_username": "", "sub": "123"},
			true,
		},
		{
			"Null preferred_username",
			map[string]interface{}{"sub": "123"},
			true,
		},
		{
			"CPF with special characters",
			map[string]interface{}{"preferred_username": "123.456.789-00", "sub": "123"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := createToken("")
			if !tt.shouldError {
				claimsJSON, _ := json.Marshal(tt.claims)
				payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
				header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
				signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
				token = header + "." + payload + "." + signature
			}

			_, err := ExtractCPFFromToken("Bearer " + token)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
