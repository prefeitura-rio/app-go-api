package utils

import (
	"testing"
)

func TestExtractDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Formatted CNAE", "77.23-3/00", "7723300"},
		{"Digits only", "7723300", "7723300"},
		{"With leading zeros", "0123456", "0123456"},
		{"Empty string", "", ""},
		{"Only special chars", ".-/", ""},
		{"Mixed format", "12-34.56/78", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDigits(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractDigits(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCNAEListToDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			"Mixed formats",
			[]string{"77.23-3/00", "4789004", "12-34"},
			[]string{"7723300", "4789004", "1234"},
		},
		{
			"Empty list",
			[]string{},
			[]string{},
		},
		{
			"With empty strings",
			[]string{"123", "", "456"},
			[]string{"123", "", "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CNAEListToDigits(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("CNAEListToDigits length = %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("CNAEListToDigits[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestHasMatchingCNAE(t *testing.T) {
	tests := []struct {
		name             string
		cnpjCNAEs        []string
		opportunityCNAEs []string
		expected         bool
	}{
		{
			"Exact match (digits only)",
			[]string{"7723300", "4789004"},
			[]string{"7723300"},
			true,
		},
		{
			"Match with formatting",
			[]string{"7723300", "4789004"},
			[]string{"77.23-3/00"},
			true,
		},
		{
			"Multiple matches",
			[]string{"7723300", "4789004"},
			[]string{"77.23-3/00", "47.89-0/04"},
			true,
		},
		{
			"No match",
			[]string{"7723300", "4789004"},
			[]string{"1234567"},
			false,
		},
		{
			"Empty CNPJ list",
			[]string{},
			[]string{"7723300"},
			false,
		},
		{
			"Empty opportunity list",
			[]string{"7723300"},
			[]string{},
			false,
		},
		{
			"Both empty",
			[]string{},
			[]string{},
			false,
		},
		{
			"Match with leading zeros",
			[]string{"0123456"},
			[]string{"01.23-4/56"},
			true,
		},
		{
			"Skip empty strings",
			[]string{"", "7723300", ""},
			[]string{"", "77.23-3/00"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasMatchingCNAE(tt.cnpjCNAEs, tt.opportunityCNAEs)
			if result != tt.expected {
				t.Errorf("HasMatchingCNAE(%v, %v) = %v, want %v",
					tt.cnpjCNAEs, tt.opportunityCNAEs, result, tt.expected)
			}
		})
	}
}
