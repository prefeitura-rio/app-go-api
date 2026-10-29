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

func TestMaskCNPJForLog(t *testing.T) {
	tests := []struct {
		name string
		cnpj string
		want string
	}{
		{"digits only", "11222333000181", "11**********81"},
		{"formatted", "11.222.333/0001-81", "11**********81"},
		{"short", "1234", "***"},
		{"empty", "", "***"},
		{"no digits", "abc", "***"},
		{"exactly 5 digits", "12345", "12*45"},
		{"non-ascii digits stay whole runes", "١٢٣٤٥٦", "١٢**٥٦"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskCNPJForLog(tt.cnpj); got != tt.want {
				t.Errorf("MaskCNPJForLog(%q) = %q, want %q", tt.cnpj, got, tt.want)
			}
		})
	}
}
