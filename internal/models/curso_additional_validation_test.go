package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatoAula_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		formato  FormatoAula
		expected bool
	}{
		{"valid_gravado", FormatoAulaGravado, true},
		{"valid_ao_vivo", FormatoAulaAoVivo, true},
		{"valid_presencial", "PRESENCIAL", true},
		{"valid_empty", "", true}, // Empty is allowed
		{"invalid_lowercase", "gravado", false},
		{"invalid_mixed_case", "Gravado", false},
		{"invalid_value", "HIBRIDO", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.formato.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatoAula_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    FormatoAula
		expected FormatoAula
	}{
		{"uppercase_gravado", "GRAVADO", "GRAVADO"},
		{"lowercase_gravado", "gravado", "GRAVADO"},
		{"mixed_case_gravado", "Gravado", "GRAVADO"},
		{"with_spaces", "  GRAVADO  ", "GRAVADO"},
		{"ao_vivo", "ao_vivo", "AO_VIVO"},
		{"empty_string", "", ""},
		{"presencial", "presencial", "PRESENCIAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Normalize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCurso_Validate_ComprehensiveValidation(t *testing.T) {
	tests := []struct {
		name        string
		curso       *Curso
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_complete_curso",
			curso: &Curso{
				Titulo:     "Curso Completo",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoOpened,
			},
			expectError: false,
		},
		{
			name: "titulo_with_only_spaces",
			curso: &Curso{
				Titulo:     "     ",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoOpened,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "titulo_with_newlines_and_tabs",
			curso: &Curso{
				Titulo:     "\n\t\r",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoOpened,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "titulo_very_long",
			curso: &Curso{
				Titulo:     string(make([]byte, 20001)),
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoOpened,
			},
			expectError: false, // DB validation will catch this
		},
		{
			name: "modalidade_legacy_values",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencialLegacy,
				Status:     StatusCursoOpened,
			},
			expectError: false,
		},
		{
			name: "modalidade_online",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadeOnline,
				Status:     StatusCursoOpened,
			},
			expectError: false,
		},
		{
			name: "modalidade_hibrido",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadeHibrido,
				Status:     StatusCursoOpened,
			},
			expectError: false,
		},
		{
			name: "modalidade_livre_formacao_online",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadeLivreFormacaoOnline,
				Status:     StatusCursoOpened,
			},
			expectError: false,
		},
		{
			name: "modalidade_invalid_case",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: "presencial",
				Status:     StatusCursoOpened,
			},
			expectError: true,
			errorMsg:    "modalidade inválida",
		},
		{
			name: "status_legacy_criado",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoCriado,
			},
			expectError: false,
		},
		{
			name: "status_legacy_aberto",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoAberto,
			},
			expectError: false,
		},
		{
			name: "status_legacy_encerrado",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoEncerrado,
			},
			expectError: false,
		},
		{
			name: "status_canceled",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoCanceled,
			},
			expectError: false,
		},
		{
			name: "status_invalid_case",
			curso: &Curso{
				Titulo:     "Curso",
				Modalidade: ModalidadePresencial,
				Status:     "DRAFT",
			},
			expectError: true,
			errorMsg:    "status inválido",
		},
		{
			name: "formato_aula_gravado",
			curso: &Curso{
				Titulo:      "Curso",
				Modalidade:  ModalidadeRemoto,
				Status:      StatusCursoOpened,
				FormatoAula: FormatoAulaGravado,
			},
			expectError: false,
		},
		{
			name: "formato_aula_ao_vivo",
			curso: &Curso{
				Titulo:      "Curso",
				Modalidade:  ModalidadeRemoto,
				Status:      StatusCursoOpened,
				FormatoAula: FormatoAulaAoVivo,
			},
			expectError: false,
		},
		{
			name: "formato_aula_presencial",
			curso: &Curso{
				Titulo:      "Curso",
				Modalidade:  ModalidadePresencial,
				Status:      StatusCursoOpened,
				FormatoAula: "PRESENCIAL",
			},
			expectError: false,
		},
		{
			name: "formato_aula_empty",
			curso: &Curso{
				Titulo:      "Curso",
				Modalidade:  ModalidadePresencial,
				Status:      StatusCursoOpened,
				FormatoAula: "",
			},
			expectError: false,
		},
		{
			name: "formato_aula_invalid",
			curso: &Curso{
				Titulo:      "Curso",
				Modalidade:  ModalidadePresencial,
				Status:      StatusCursoOpened,
				FormatoAula: "HIBRIDO",
			},
			expectError: true,
			errorMsg:    "formato de aula inválido",
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

func TestCurso_FieldValidation_EdgeCases(t *testing.T) {
	t.Run("nullable_dates", func(t *testing.T) {
		curso := &Curso{
			Titulo:              "Curso",
			Modalidade:          ModalidadePresencial,
			Status:              StatusCursoOpened,
			EnrollmentStartDate: nil,
			EnrollmentEndDate:   nil,
			DataInicio:          nil,
			DataTermino:         nil,
		}
		assert.NoError(t, curso.Validate())
	})

	t.Run("valid_dates", func(t *testing.T) {
		now := time.Now()
		future := now.Add(30 * 24 * time.Hour)

		curso := &Curso{
			Titulo:              "Curso",
			Modalidade:          ModalidadePresencial,
			Status:              StatusCursoOpened,
			EnrollmentStartDate: &now,
			EnrollmentEndDate:   &future,
		}
		assert.NoError(t, curso.Validate())
	})

	t.Run("boolean_flags", func(t *testing.T) {
		trueVal := true
		falseVal := false

		curso := &Curso{
			Titulo:                 "Curso",
			Modalidade:             ModalidadePresencial,
			Status:                 StatusCursoOpened,
			HasCertificate:         true,
			IsExternalPartner:      &trueVal,
			IsVisible:              &trueVal,
			AutoApproveEnrollments: &falseVal,
		}
		assert.NoError(t, curso.Validate())
	})

	t.Run("course_management_types", func(t *testing.T) {
		types := []CourseManagementType{
			CourseManagementOwnOrg,
			CourseManagementExternalManagedByOrg,
			CourseManagementExternalManagedByPartner,
			"",
		}

		for _, cmt := range types {
			curso := &Curso{
				Titulo:               "Curso",
				Modalidade:           ModalidadePresencial,
				Status:               StatusCursoOpened,
				CourseManagementType: cmt,
			}
			assert.NoError(t, curso.Validate())
		}
	})

	t.Run("special_characters_in_titulo", func(t *testing.T) {
		titulos := []string{
			"Curso com @!#$%",
			"Curso com números 12345",
			"Curso com àçénts",
			"Curso com ñ e ü",
			"Curso: Subtítulo",
			"Curso (Edição 2024)",
			"Curso - Módulo 1",
		}

		for _, titulo := range titulos {
			curso := &Curso{
				Titulo:     titulo,
				Modalidade: ModalidadePresencial,
				Status:     StatusCursoOpened,
			}
			assert.NoError(t, curso.Validate(), "Failed for titulo: %s", titulo)
		}
	})
}

func TestCurso_OptionalFields(t *testing.T) {
	curso := &Curso{
		Titulo:     "Curso Mínimo",
		Modalidade: ModalidadePresencial,
		Status:     StatusCursoOpened,
	}

	// Verify all optional fields can be nil/empty
	assert.Empty(t, curso.Descricao)
	assert.Nil(t, curso.EnrollmentStartDate)
	assert.Nil(t, curso.EnrollmentEndDate)
	assert.Empty(t, curso.Organization)
	assert.Empty(t, curso.Theme)
	assert.Empty(t, curso.Workload)
	assert.Empty(t, curso.TargetAudience)
	assert.Empty(t, curso.InstitutionalLogo)
	assert.Empty(t, curso.CoverImage)
	assert.Empty(t, curso.OrgaoID)
	assert.Nil(t, curso.InstituicaoID)
	assert.Empty(t, curso.LocalRealizacao)
	assert.Nil(t, curso.DataInicio)
	assert.Nil(t, curso.DataTermino)
	assert.Nil(t, curso.DataLimiteInscricoes)
	assert.Zero(t, curso.NumeroVagas)
	assert.Zero(t, curso.CargaHoraria)
	assert.Empty(t, curso.Turno)
	assert.Empty(t, curso.FormatoAula)
	assert.Empty(t, curso.FormacaoLink)
	assert.Empty(t, curso.LinkInscricao)
	assert.Empty(t, curso.ContatoDuvidas)
	assert.False(t, curso.HasCertificate)
	assert.Empty(t, curso.Facilitator)
	assert.Empty(t, curso.Objectives)
	assert.Empty(t, curso.ExpectedResults)
	assert.Empty(t, curso.ProgramContent)
	assert.Empty(t, curso.Methodology)
	assert.Empty(t, curso.ResourcesUsed)
	assert.Empty(t, curso.MaterialUsed)
	assert.Empty(t, curso.TeachingMaterial)
	assert.Empty(t, curso.PreRequisitos)
	assert.False(t, curso.CertificacaoOferecida)
	assert.Nil(t, curso.IsExternalPartner)
	assert.Empty(t, curso.ExternalPartnerName)
	assert.Empty(t, curso.ExternalPartnerURL)
	assert.Empty(t, curso.ExternalPartnerLogoURL)
	assert.Empty(t, curso.ExternalPartnerContact)
	assert.Empty(t, curso.Accessibility)
	assert.Nil(t, curso.IsVisible)
	assert.Nil(t, curso.AutoApproveEnrollments)

	assert.NoError(t, curso.Validate())
}
