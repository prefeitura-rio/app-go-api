package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatoAula_NormalizeBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    FormatoAula
		expected FormatoAula
	}{
		{"empty_stays_empty", "", ""},
		{"uppercase_stays_same", "GRAVADO", "GRAVADO"},
		{"lowercase_to_uppercase", "gravado", "GRAVADO"},
		{"mixed_case_to_uppercase", "Gravado", "GRAVADO"},
		{"with_spaces_trimmed_and_uppercase", "  ao_vivo  ", "AO_VIVO"},
		{"only_spaces_becomes_empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Normalize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCurso_Validate(t *testing.T) {
	tests := []struct {
		name        string
		curso       *Curso
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_curso_minimal",
			curso: &Curso{
				Titulo:     "Curso de Go",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoDraft,
			},
			expectError: false,
		},
		{
			name: "valid_curso_with_formato_aula",
			curso: &Curso{
				Titulo:      "Curso de Go",
				Modalidade:  ModalidadeRemoto,
				Status:      StatusCursoOpened,
				FormatoAula: FormatoAulaGravado,
			},
			expectError: false,
		},
		{
			name: "missing_titulo",
			curso: &Curso{
				Titulo:     "",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoDraft,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "titulo_only_spaces",
			curso: &Curso{
				Titulo:     "   ",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoDraft,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "invalid_modalidade",
			curso: &Curso{
				Titulo:     "Curso de Go",
				Modalidade: "INVALID_MODE",
				Status:     StatusCursoDraft,
			},
			expectError: true,
			errorMsg:    "modalidade inválida",
		},
		{
			name: "invalid_status",
			curso: &Curso{
				Titulo:     "Curso de Go",
				Modalidade: ModalidadePresencial,
				Status:     "INVALID_STATUS",
			},
			expectError: true,
			errorMsg:    "status inválido",
		},
		{
			name: "invalid_formato_aula",
			curso: &Curso{
				Titulo:      "Curso de Go",
				Modalidade:  ModalidadePresencial,
				Status:      StatusCursoDraft,
				FormatoAula: "INVALID_FORMAT",
			},
			expectError: true,
			errorMsg:    "formato de aula inválido",
		},
		{
			name: "valid_empty_formato_aula",
			curso: &Curso{
				Titulo:      "Curso de Go",
				Modalidade:  ModalidadePresencial,
				Status:      StatusCursoDraft,
				FormatoAula: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.curso.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModalidade_EdgeCases(t *testing.T) {
	t.Run("case_sensitive_validation", func(t *testing.T) {
		// Test that validation is case-sensitive
		invalidCases := []Modalidade{
			"presencial", // lowercase
			"remoto",     // lowercase
			"PreSencial", // mixed case
		}

		for _, modalidade := range invalidCases {
			assert.False(t, modalidade.IsValid(), "Expected %s to be invalid", modalidade)
		}
	})

	t.Run("normalize_handles_all_legacy_values", func(t *testing.T) {
		// Test all legacy mappings
		assert.Equal(t, ModalidadePresencial, Modalidade("PRESENCIAL").Normalize())
		assert.Equal(t, ModalidadeRemoto, Modalidade("ONLINE").Normalize())
		assert.Equal(t, ModalidadeSemipresencial, Modalidade("HIBRIDO").Normalize())
	})

	t.Run("normalize_preserves_new_values", func(t *testing.T) {
		assert.Equal(t, ModalidadePresencial, ModalidadePresencial.Normalize())
		assert.Equal(t, ModalidadeRemoto, ModalidadeRemoto.Normalize())
		assert.Equal(t, ModalidadeSemipresencial, ModalidadeSemipresencial.Normalize())
	})
}

func TestStatusCurso_EdgeCases(t *testing.T) {
	t.Run("case_sensitive_validation", func(t *testing.T) {
		// Test that validation is case-sensitive
		invalidCases := []StatusCurso{
			"Draft",  // mixed case
			"DRAFT",  // uppercase (not in valid list)
			"Opened", // mixed case
		}

		for _, status := range invalidCases {
			assert.False(t, status.IsValid(), "Expected %s to be invalid", status)
		}
	})

	t.Run("normalize_handles_all_legacy_values", func(t *testing.T) {
		// Test all legacy mappings
		assert.Equal(t, StatusCursoDraft, StatusCurso("CRIADO").Normalize())
		assert.Equal(t, StatusCursoPublished, StatusCurso("ABERTO").Normalize())
		assert.Equal(t, StatusCursoClosed, StatusCurso("ENCERRADO").Normalize())
	})

	t.Run("normalize_case_insensitive", func(t *testing.T) {
		// Normalize should handle different cases
		assert.Equal(t, StatusCursoDraft, StatusCurso("criado").Normalize())
		assert.Equal(t, StatusCursoPublished, StatusCurso("aberto").Normalize())
		assert.Equal(t, StatusCursoClosed, StatusCurso("encerrado").Normalize())
	})
}

func TestCourseManagementType_EdgeCases(t *testing.T) {
	t.Run("all_valid_types", func(t *testing.T) {
		validTypes := []CourseManagementType{
			CourseManagementOwnOrg,
			CourseManagementExternalManagedByOrg,
			CourseManagementExternalManagedByPartner,
			"",
		}

		for _, cmt := range validTypes {
			assert.True(t, cmt.IsValid(), "Expected %s to be valid", cmt)
		}
	})

	t.Run("case_sensitive", func(t *testing.T) {
		// Test case sensitivity - lowercase should be invalid
		assert.False(t, CourseManagementType("own_org").IsValid())
		// OWN_ORG is valid (matches the constant)
		assert.True(t, CourseManagementType("OWN_ORG").IsValid())
		// Invalid values
		assert.False(t, CourseManagementType("INVALID").IsValid())
		assert.False(t, CourseManagementType("external").IsValid())
	})
}

func TestCurso_ValidateWithAllFields(t *testing.T) {
	t.Run("valid_complete_curso", func(t *testing.T) {
		curso := &Curso{
			Titulo:               "Curso Completo de Go",
			Descricao:            "Descrição detalhada do curso",
			Organization:         "Prefeitura do Rio",
			Modalidade:           ModalidadePresencial,
			Theme:                "Programação",
			Workload:             "40 horas",
			TargetAudience:       "Desenvolvedores",
			InstitutionalLogo:    "https://example.com/logo.png",
			CoverImage:           "https://example.com/cover.jpg",
			Status:               StatusCursoOpened,
			FormatoAula:          FormatoAulaGravado,
			NumeroVagas:          30,
			CargaHoraria:         40,
			HasCertificate:       true,
			Facilitator:          "João Silva",
			Objectives:           "Aprender Go",
			CourseManagementType: CourseManagementOwnOrg,
			CertificacaoOferecida: true,
		}

		err := curso.Validate()
		assert.NoError(t, err)
	})
}

func TestModalidade_AllValidValues(t *testing.T) {
	validValues := []Modalidade{
		ModalidadePresencial,
		ModalidadeSemipresencial,
		ModalidadeRemoto,
		ModalidadePresencialLegacy,
		ModalidadeOnline,
		ModalidadeHibrido,
		ModalidadeLivreFormacaoOnline,
	}

	for _, modalidade := range validValues {
		t.Run(string(modalidade), func(t *testing.T) {
			assert.True(t, modalidade.IsValid(), "Expected %s to be valid", modalidade)
		})
	}
}

func TestStatusCurso_AllValidValues(t *testing.T) {
	validValues := []StatusCurso{
		StatusCursoDraft,
		StatusCursoOpened,
		StatusCursoClosed,
		StatusCursoCanceled,
		StatusCursoCriado,
		StatusCursoAberto,
		StatusCursoEncerrado,
	}

	for _, status := range validValues {
		t.Run(string(status), func(t *testing.T) {
			assert.True(t, status.IsValid(), "Expected %s to be valid", status)
		})
	}
}
