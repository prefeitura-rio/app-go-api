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

func TestExtractDigits_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Arabic-Indic numerals", "١٢٣٤٥", "١٢٣٤٥"},    // unicode.IsDigit accepts these
		{"Mixed unicode", "123٤٥٦", "123٤٥٦"},         // All unicode digits
		{"Chinese numerals", "一二三四", ""},               // Not digits
		{"Emoji with numbers", "🔢123🔢", "123"},        // Extracts ASCII digits
		{"Full-width digits", "１２３４５", "１２３４５"},     // unicode.IsDigit accepts these
		{"Subscript digits", "₁₂₃₄₅", ""},              // Not considered digits
		{"Superscript digits", "¹²³⁴⁵", ""},            // Not considered digits
		{"Roman numerals", "ⅠⅡⅢⅣⅤ", ""},               // Not digits
		{"Letters and numbers", "abc123def", "123"},   // Only digits extracted
		{"Special chars", "!@#$%123^&*", "123"},       // Only digits extracted
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

func TestExtractDigits_Performance(t *testing.T) {
	// Test with very long string
	longInput := ""
	for i := 0; i < 10000; i++ {
		longInput += "12-34.56/78 "
	}

	result := ExtractDigits(longInput)
	expectedLen := 10000 * 8 // 8 digits per iteration
	if len(result) != expectedLen {
		t.Errorf("ExtractDigits length = %d, want %d", len(result), expectedLen)
	}
}

func TestCNAEListToDigits_NilInput(t *testing.T) {
	// Even though nil slice, function should handle gracefully
	result := CNAEListToDigits(nil)
	if result == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}
}

func TestHasMatchingCNAE_LargeDataset(t *testing.T) {
	// Test with large datasets to ensure O(1) lookup works
	cnpjCNAEs := make([]string, 1000)
	opportunityCNAEs := make([]string, 1000)

	for i := 0; i < 1000; i++ {
		cnpjCNAEs[i] = string(rune('0' + (i % 10)))
		opportunityCNAEs[i] = string(rune('0' + ((i + 500) % 10)))
	}

	// Should find matches quickly
	result := HasMatchingCNAE(cnpjCNAEs, opportunityCNAEs)
	if !result {
		t.Error("Expected match in large dataset")
	}
}

func TestHasMatchingCNAE_OnlyEmptyStrings(t *testing.T) {
	// All empty strings should not match
	cnpjCNAEs := []string{"", "", ""}
	opportunityCNAEs := []string{"", "", ""}

	result := HasMatchingCNAE(cnpjCNAEs, opportunityCNAEs)
	if result {
		t.Error("Empty strings should not match")
	}
}
