package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestImportRowErrorData_MasksNameAndCPF garante que o JobError.Data de uma
// linha que falhou na importação nunca leva o nome completo nem o CPF em claro
// (DEV-149): o status do job é lido por HTTP e o dado fica persistido em jobs.errors.
func TestImportRowErrorData_MasksNameAndCPF(t *testing.T) {
	tests := []struct {
		name string
		row  EnrollmentRow
		want string
	}{
		{
			name: "CPF formatado e nome completo",
			row:  EnrollmentRow{NomeCompleto: "Maria da Silva Souza", CPF: "123.456.789-09"},
			want: "M. d. S. S. - 123******09",
		},
		{
			name: "CPF só dígitos",
			row:  EnrollmentRow{NomeCompleto: "João Pereira", CPF: "12345678909"},
			want: "J. P. - 123******09",
		},
		{
			name: "espaços extras e acento na inicial",
			row:  EnrollmentRow{NomeCompleto: "  Ângela   Lima ", CPF: " 123.456.789-09 "},
			want: "Â. L. - 123******09",
		},
		{
			name: "CPF curto demais não vaza dígitos",
			row:  EnrollmentRow{NomeCompleto: "Ana", CPF: "1234"},
			want: "A. - ***",
		},
		{
			name: "nome vazio",
			row:  EnrollmentRow{NomeCompleto: "", CPF: "12345678909"},
			want: " - 123******09",
		},
		{
			name: "CPF vazio",
			row:  EnrollmentRow{NomeCompleto: "Carlos Alberto", CPF: ""},
			want: "C. A. - ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := importRowErrorData(tt.row)
			assert.Equal(t, tt.want, got)
			assert.NotContains(t, got, "12345678909")
			assert.NotContains(t, got, "456.789")
			if tt.row.NomeCompleto != "" {
				assert.NotContains(t, got, "Silva")
				assert.NotContains(t, got, "Pereira")
				assert.NotContains(t, got, "Alberto")
			}
		})
	}
}
