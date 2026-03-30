package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "ISO 8601 date",
			dateStr: "2006-01-02",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "Brazilian date format",
			dateStr: "02/01/2006",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "ISO 8601 datetime with Z",
			dateStr: "2006-01-02T15:04:05Z",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "ISO 8601 datetime with timezone",
			dateStr: "2006-01-02T15:04:05-03:00",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "invalid date format",
			dateStr: "invalid",
			wantErr: false,
			wantNil: true,
		},
		{
			name:    "empty string",
			dateStr: "",
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.dateStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
			}
		})
	}
}

func TestParseDateSpecificValues(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected time.Time
	}{
		{
			name:     "ISO date",
			dateStr:  "2023-12-25",
			expected: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Brazilian date",
			dateStr:  "25/12/2023",
			expected: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.dateStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Year(), got.Year())
			assert.Equal(t, tt.expected.Month(), got.Month())
			assert.Equal(t, tt.expected.Day(), got.Day())
		})
	}
}

func TestMaskCPF(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want string
	}{
		{
			name: "full CPF without formatting",
			cpf:  "12345678901",
			want: "123******01",
		},
		{
			name: "full CPF with dots",
			cpf:  "123.456.789-01",
			want: "123******01",
		},
		{
			name: "full CPF with dots only",
			cpf:  "123.456.78901",
			want: "123******01",
		},
		{
			name: "short CPF - too short to mask properly",
			cpf:  "1234",
			want: "***",
		},
		{
			name: "very short CPF",
			cpf:  "12",
			want: "***",
		},
		{
			name: "empty CPF",
			cpf:  "",
			want: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskCPF(tt.cpf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskCPFFormat(t *testing.T) {
	// Test that the mask always shows first 3 and last 2 digits
	cpf := "12345678901"
	masked := maskCPF(cpf)

	assert.Equal(t, 11, len(masked), "Masked CPF should have 11 characters")
	assert.Equal(t, "123", masked[:3], "Should show first 3 digits")
	assert.Equal(t, "01", masked[len(masked)-2:], "Should show last 2 digits")
	assert.Contains(t, masked, "******", "Should contain 6 asterisks")
}
