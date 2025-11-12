package utils

import (
	"strings"
	"unicode"
)

// ExtractDigits removes all non-digit characters from a CNAE string
// This is used to normalize CNAEs for comparison (removes dots, hyphens, etc.)
// Example: "77.23-3/00" -> "7723300"
func ExtractDigits(cnae string) string {
	var result strings.Builder
	for _, r := range cnae {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// CNAEListToDigits converts an array of CNAEs to digits-only format
func CNAEListToDigits(cnaes []string) []string {
	result := make([]string, len(cnaes))
	for i, cnae := range cnaes {
		result[i] = ExtractDigits(cnae)
	}
	return result
}

// HasMatchingCNAE checks if any CNPJ CNAE matches any opportunity CNAE
// Both lists are compared using digits-only format (formatting removed)
// Returns true if at least one match is found
func HasMatchingCNAE(cnpjCNAEs []string, opportunityCNAEs []string) bool {
	// Normalize opportunity CNAEs to digits only
	opportunityDigits := CNAEListToDigits(opportunityCNAEs)

	// Create a map for O(1) lookup
	opportunitySet := make(map[string]bool)
	for _, cnae := range opportunityDigits {
		if cnae != "" { // Skip empty strings
			opportunitySet[cnae] = true
		}
	}

	// Check if any CNPJ CNAE matches
	for _, cnpjCNAE := range cnpjCNAEs {
		digitsOnly := ExtractDigits(cnpjCNAE)
		if digitsOnly != "" && opportunitySet[digitsOnly] {
			return true
		}
	}

	return false
}
