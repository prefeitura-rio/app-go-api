package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-empty string",
			input: "test",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "string with spaces",
			input: "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringPtr(tt.input)
			assert.NotNil(t, got, "stringPtr should never return nil")
			assert.Equal(t, tt.input, *got, "stringPtr should preserve the input value")
		})
	}
}

func TestStringPtrAddressUniqueness(t *testing.T) {
	// Test that each call to stringPtr returns a unique pointer
	str1 := stringPtr("test")
	str2 := stringPtr("test")

	assert.Equal(t, *str1, *str2, "Values should be equal")
	assert.NotSame(t, str1, str2, "Pointers should be different")
}
