package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTurno_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		turno    Turno
		expected bool
	}{
		{"valid_manha", TurnoManha, true},
		{"valid_tarde", TurnoTarde, true},
		{"valid_noite", TurnoNoite, true},
		{"valid_integral", TurnoIntegral, true},
		{"valid_livre", TurnoLivre, true},
		{"invalid_empty", "", false},
		{"invalid_lowercase", "manha", false},
		{"invalid_mixed_case", "Manha", false},
		{"invalid_value", "VESPERTINO", false},
		{"invalid_special_chars", "MANHÃ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.turno.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
