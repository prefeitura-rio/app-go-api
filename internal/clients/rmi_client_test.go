package clients

import (
	"errors"
	"testing"
)

func TestNormalizeCNPJ(t *testing.T) {
	t.Run("Valid CNPJ without formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12345678000195")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Valid CNPJ with standard formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12.345.678/0001-95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Valid CNPJ with partial formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12345678/0001-95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("CNPJ with spaces", func(t *testing.T) {
		result, err := NormalizeCNPJ("12 345 678 0001 95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Invalid CNPJ - too short", func(t *testing.T) {
		_, err := NormalizeCNPJ("1234567800019")

		if err == nil {
			t.Error("Expected error for CNPJ with 13 digits")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - too long", func(t *testing.T) {
		_, err := NormalizeCNPJ("123456780001950")

		if err == nil {
			t.Error("Expected error for CNPJ with 15 digits")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - empty string", func(t *testing.T) {
		_, err := NormalizeCNPJ("")

		if err == nil {
			t.Error("Expected error for empty CNPJ")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - only letters", func(t *testing.T) {
		_, err := NormalizeCNPJ("abcdefghijklmn")

		if err == nil {
			t.Error("Expected error for CNPJ with only letters")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - mixed letters and numbers", func(t *testing.T) {
		_, err := NormalizeCNPJ("12.345.678/ABCD-95")

		if err == nil {
			t.Error("Expected error for CNPJ with letters")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})
}
