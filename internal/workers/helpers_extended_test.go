package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateAllFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantErr  bool
	}{
		{
			name:     "ISO 8601 date",
			input:    "2023-12-25",
			wantYear: 2023,
			wantErr:  false,
		},
		{
			name:     "Brazilian date",
			input:    "25/12/2023",
			wantYear: 2023,
			wantErr:  false,
		},
		{
			name:     "ISO 8601 with Z timezone",
			input:    "2023-12-25T15:04:05Z",
			wantYear: 2023,
			wantErr:  false,
		},
		{
			name:     "ISO 8601 with offset",
			input:    "2023-12-25T15:04:05-03:00",
			wantYear: 2023,
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			input:    "not-a-date",
			wantYear: 0,
			wantErr:  false, // Returns zero time, no error
		},
		{
			name:     "Empty string",
			input:    "",
			wantYear: 0,
			wantErr:  false,
		},
		{
			name:     "Partial date",
			input:    "2023-12",
			wantYear: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDate(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantYear > 0 {
					assert.Equal(t, tt.wantYear, result.Year())
					assert.False(t, result.IsZero())
				} else {
					assert.True(t, result.IsZero())
				}
			}
		})
	}
}

func TestParseDatePrecision(t *testing.T) {
	// Test that ISO date preserves the exact date
	input := "2023-06-15"
	result, err := parseDate(input)

	assert.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.June, result.Month())
	assert.Equal(t, 15, result.Day())
}

func TestParseDateBrazilianFormat(t *testing.T) {
	// Test Brazilian date format
	input := "15/06/2023"
	result, err := parseDate(input)

	assert.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.June, result.Month())
	assert.Equal(t, 15, result.Day())
}

func TestMaskCPFEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full CPF no formatting",
			input:    "12345678901",
			expected: "123******01",
		},
		{
			name:     "Full CPF with dots and dash",
			input:    "123.456.789-01",
			expected: "123******01",
		},
		{
			name:     "CPF with only dots",
			input:    "123.456.789.01",
			expected: "123******01",
		},
		{
			name:     "Very short CPF",
			input:    "123",
			expected: "***",
		},
		{
			name:     "4 digits",
			input:    "1234",
			expected: "***",
		},
		{
			name:     "5 digits - minimum for masking",
			input:    "12345",
			expected: "123******45",
		},
		{
			name:     "Empty CPF",
			input:    "",
			expected: "***",
		},
		{
			name:     "Exactly 11 digits",
			input:    "11111111111",
			expected: "111******11",
		},
		{
			name:     "With spaces",
			input:    "123 456 789 01",
			expected: "123******01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCPF(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskCPFPreservesFirstAndLast(t *testing.T) {
	cpf := "98765432109"
	masked := maskCPF(cpf)

	// Should preserve first 3 and last 2
	assert.True(t, len(masked) >= 5)
	assert.Equal(t, "987", masked[:3])
	assert.Equal(t, "09", masked[len(masked)-2:])
}

func TestMaskCPFMultipleCalls(t *testing.T) {
	cpf := "12345678901"

	// Multiple calls should return same result
	result1 := maskCPF(cpf)
	result2 := maskCPF(cpf)

	assert.Equal(t, result1, result2)
	assert.Equal(t, "123******01", result1)
}

func TestStringPtrVariations(t *testing.T) {
	// Test with different string values
	values := []string{
		"simple",
		"with spaces",
		"with\nnewline",
		"with\ttab",
		"",
		"   ",
		"UPPERCASE",
		"lowercase",
		"MiXeD",
		"special!@#$%^&*()",
		"números123",
		"açẽntõs",
	}

	for _, val := range values {
		ptr := stringPtr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	}
}

func TestStringPtrConcurrency(t *testing.T) {
	// Test that stringPtr is safe for concurrent use
	done := make(chan bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		go func(val int) {
			str := stringPtr("test")
			assert.NotNil(t, str)
			assert.Equal(t, "test", *str)
			done <- true
		}(i)
	}

	for i := 0; i < iterations; i++ {
		<-done
	}
}

func TestParseDateTimezone(t *testing.T) {
	// Test that timezone parsing works correctly
	withZ := "2023-01-15T10:30:00Z"
	withOffset := "2023-01-15T10:30:00-03:00"

	resultZ, err := parseDate(withZ)
	assert.NoError(t, err)
	assert.Equal(t, 2023, resultZ.Year())

	resultOffset, err := parseDate(withOffset)
	assert.NoError(t, err)
	assert.Equal(t, 2023, resultOffset.Year())
}

func TestParseDateReturnsZeroOnInvalid(t *testing.T) {
	invalidInputs := []string{
		"not a date",
		"2023-13-01", // Invalid month
		"2023-01-32", // Invalid day
		"abc/def/ghi",
		"12345",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			result, err := parseDate(input)
			assert.NoError(t, err) // No error, just returns zero
			assert.True(t, result.IsZero())
		})
	}
}

func TestMaskCPFDoesNotModifyOriginal(t *testing.T) {
	original := "12345678901"
	masked := maskCPF(original)

	// Original should be unchanged
	assert.Equal(t, "12345678901", original)
	assert.Equal(t, "123******01", masked)
}
