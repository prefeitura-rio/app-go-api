package clients

import (
	"errors"
	"testing"
)

// DEV-261: ValidateCNPJ confere o formato e os dígitos verificadores do CNPJ,
// numérico ou alfanumérico (Receita Federal), antes de ele virar chave primária.
func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name    string
		cnpj    string
		wantErr bool
	}{
		{"numérico válido", "11222333000181", false},
		{"numérico válido formatado", "11.222.333/0001-81", false},
		{"outro numérico válido", "11444777000161", false},
		{"alfanumérico válido (exemplo da Receita)", "12ABC34501DE35", false},
		{"alfanumérico válido formatado e minúsculo", "12.abc.345/01de-35", false},
		{"vazio", "", true},
		{"só espaços", "   ", true},
		{"curto demais", "1122233300018", true},
		{"longo demais", "112223330001810", true},
		{"primeiro dígito verificador errado", "11222333000191", true},
		{"segundo dígito verificador errado", "11222333000182", true},
		{"alfanumérico com DV errado", "12ABC34501DE36", true},
		{"letra no dígito verificador", "12ABC34501DEA5", true},
		{"caractere fora do formato", "11222333000#81", true},
		{"todos os caracteres iguais", "00000000000000", true},
		{"todos os dígitos iguais", "11111111111111", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCNPJ(tt.cnpj)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidCNPJ) {
					t.Fatalf("ValidateCNPJ(%q): expected ErrInvalidCNPJ, got %v", tt.cnpj, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateCNPJ(%q): expected no error, got %v", tt.cnpj, err)
			}
		})
	}
}
