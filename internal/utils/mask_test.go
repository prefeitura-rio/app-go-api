package utils

import "testing"

func TestMaskCPFForLog(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want string
	}{
		{"full cpf", "12345678901", "123******01"},
		{"short", "12", "***"},
		{"empty", "", "***"},
		{"exactly 5", "12345", "123******45"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskCPFForLog(tt.cpf); got != tt.want {
				t.Errorf("MaskCPFForLog(%q) = %q, want %q", tt.cpf, got, tt.want)
			}
		})
	}
}
