package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// SQL Injection and XSS tests for DTOs
func TestCurso_Validate_SecurityPatterns(t *testing.T) {
	sqlInjectionPatterns := []string{
		"'; DROP TABLE cursos; --",
		"1' OR '1'='1",
		"admin'--",
		"' UNION SELECT NULL--",
		"1; DELETE FROM users WHERE '1'='1",
	}

	for _, pattern := range sqlInjectionPatterns {
		curso := &Curso{
			Titulo:     pattern,
			Modalidade: ModalidadePresencial,
			Status:     StatusCursoOpened,
		}
		// Should not error - sanitization happens at DB layer
		assert.NoError(t, curso.Validate(), "SQL pattern should be accepted: %s", pattern)
	}

	xssPatterns := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert(1)>",
		"javascript:alert(1)",
		"<iframe src='javascript:alert(1)'>",
		"<body onload=alert(1)>",
	}

	for _, pattern := range xssPatterns {
		curso := &Curso{
			Titulo:     pattern,
			Modalidade: ModalidadePresencial,
			Status:     StatusCursoOpened,
		}
		// Should not error - sanitization happens at presentation layer
		assert.NoError(t, curso.Validate(), "XSS pattern should be accepted: %s", pattern)
	}
}

func TestOportunidadeMEI_Validate_SecurityPatterns(t *testing.T) {
	sqlPattern := "'; DROP TABLE oportunidades_mei; --"

	oportunidade := &OportunidadeMEI{
		Status: StatusOportunidadeDraft,
		Titulo: sqlPattern,
	}
	// Draft allows partial data
	assert.NoError(t, oportunidade.Validate())
}

func TestPropostaMEI_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cnpj        string
		expectError bool
	}{
		{"cnpj_with_dots_and_slashes", "12.345.678/0001-90", false},
		{"cnpj_only_numbers", "12345678000190", false},
		{"cnpj_with_spaces", "12 345 678 0001 90", false},
		{"cnpj_very_long", strings.Repeat("1", 100), false},
		{"cnpj_special_chars", "CNPJ@#$%", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposta := &PropostaMEI{
				OportunidadeMEIID: 1,
				MEIEmpresaID:      tt.cnpj,
			}
			err := proposta.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmprego_Validate_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		titulo      string
		expectError bool
	}{
		{"titulo_single_char", "A", false},
		{"titulo_very_long", strings.Repeat("A", 20000), false},
		{"titulo_unicode", "Título com ação, coração, programação", false},
		{"titulo_emoji", "Desenvolvedor 🚀", false},
		{"titulo_only_numbers", "12345", false},
		{"titulo_mixed_languages", "Developer 開発者 Développeur", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emprego := &Emprego{
				Titulo:          tt.titulo,
				TipoContratacao: TipoContratacaoCLT,
				Status:          StatusEmpregoAberto,
			}
			err := emprego.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOportunidadeMEI_ValidateForPublish_CNAEValidation(t *testing.T) {
	tests := []struct {
		name        string
		cnaes       []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "single_cnae",
			cnaes:       []string{"8130300"},
			expectError: false,
		},
		{
			name:        "multiple_cnaes",
			cnaes:       []string{"8130300", "8121400", "8129000"},
			expectError: false,
		},
		{
			name:        "empty_cnae_array",
			cnaes:       []string{},
			expectError: true,
			errorMsg:    "pelo menos um CNAE é obrigatório",
		},
		{
			name:        "cnae_with_formatting",
			cnaes:       []string{"81.30-3/00", "81.21-4/00"},
			expectError: false,
		},
		{
			name:        "very_long_cnae",
			cnaes:       []string{strings.Repeat("1", 100)},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oportunidade := &OportunidadeMEI{
				Titulo:           "Test",
				DescricaoServico: "Test",
				OrgaoID:          "TEST",
				CNAEIDs:          tt.cnaes,
				Logradouro:       "Rua",
				Numero:           "1",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    timePointer(time.Now().Add(24 * time.Hour)),
			}

			err := oportunidade.ValidateForPublish()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOportunidadeMEI_Validate_AddressFields(t *testing.T) {
	baseOportunidade := func() *OportunidadeMEI {
		return &OportunidadeMEI{
			Titulo:           "Test",
			DescricaoServico: "Test",
			OrgaoID:          "TEST",
			CNAEIDs:          []string{"8130300"},
			Logradouro:       "Av. Atlântica",
			Numero:           "1234",
			Bairro:           "Copacabana",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			DataExpiracao:    timePointer(time.Now().Add(24 * time.Hour)),
			Status:           StatusOportunidadeActive,
		}
	}

	tests := []struct {
		name        string
		modify      func(*OportunidadeMEI)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_complete_address",
			modify:      func(o *OportunidadeMEI) {},
			expectError: false,
		},
		{
			name: "numero_s_n",
			modify: func(o *OportunidadeMEI) {
				o.Numero = "S/N"
			},
			expectError: false,
		},
		{
			name: "numero_sn",
			modify: func(o *OportunidadeMEI) {
				o.Numero = "SN"
			},
			expectError: false,
		},
		{
			name: "complemento_empty",
			modify: func(o *OportunidadeMEI) {
				o.Complemento = ""
			},
			expectError: false,
		},
		{
			name: "complemento_with_special_chars",
			modify: func(o *OportunidadeMEI) {
				o.Complemento = "Apto 101 - Bloco A"
			},
			expectError: false,
		},
		{
			name: "estado_valid_codes",
			modify: func(o *OportunidadeMEI) {
				o.Estado = "SP"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oportunidade := baseOportunidade()
			tt.modify(oportunidade)

			err := oportunidade.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCurso_Validate_URLFields(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"http_url", "http://example.com"},
		{"https_url", "https://example.com"},
		{"url_with_path", "https://example.com/path/to/resource"},
		{"url_with_query", "https://example.com?param=value"},
		{"url_with_fragment", "https://example.com#section"},
		{"url_with_port", "https://example.com:8080"},
		{"relative_url", "/path/to/resource"},
		{"data_url", "data:image/png;base64,iVBORw0KG"},
		{"invalid_url", "not a url"},
		{"empty_url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			curso := &Curso{
				Titulo:            "Curso",
				Modalidade:        ModalidadePresencial,
				Status:            StatusCursoOpened,
				FormacaoLink:      tt.url,
				LinkInscricao:     tt.url,
				InstitutionalLogo: tt.url,
				CoverImage:        tt.url,
			}
			// URL validation should happen at application layer, not model validation
			assert.NoError(t, curso.Validate())
		})
	}
}

func TestPropostaMEI_Validate_NumericBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		valor       *float64
		expectError bool
	}{
		{"zero_value", float64Ptr(0), false},
		{"small_positive", float64Ptr(0.01), false},
		{"large_value", float64Ptr(99999999.99), false},
		{"negative_value", float64Ptr(-0.01), true},
		{"very_negative", float64Ptr(-99999999.99), true},
		{"nil_value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposta := &PropostaMEI{
				OportunidadeMEIID: 1,
				MEIEmpresaID:      "12345678000190",
				ValorProposta:     tt.valor,
			}

			err := proposta.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions
func float64Ptr(f float64) *float64 {
	return &f
}

func timePointer(t time.Time) *time.Time {
	return &t
}
