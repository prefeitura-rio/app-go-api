package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegalEntity_GetAllCNAEs(t *testing.T) {
	tests := []struct {
		name     string
		entity   *LegalEntity
		expected []string
	}{
		{
			name: "single_fiscal",
			entity: &LegalEntity{
				CNAEFiscal: "1234567",
			},
			expected: []string{"1234567"},
		},
		{
			name: "with_secundarias",
			entity: &LegalEntity{
				CNAEFiscal:      "1234567",
				CNAESecundarias: []string{"2345678", "3456789"},
			},
			expected: []string{"1234567", "2345678", "3456789"},
		},
		{
			name: "empty_fiscal",
			entity: &LegalEntity{
				CNAEFiscal:      "",
				CNAESecundarias: []string{"2345678"},
			},
			expected: []string{"2345678"},
		},
		{
			name: "only_secundarias",
			entity: &LegalEntity{
				CNAESecundarias: []string{"2345678", "3456789"},
			},
			expected: []string{"2345678", "3456789"},
		},
		{
			name:     "all_empty",
			entity:   &LegalEntity{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entity.GetAllCNAEs()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLegalEntityFull_ToConsultaResponse(t *testing.T) {
	t.Run("valid_entity", func(t *testing.T) {
		entity := &LegalEntityFull{
			CNPJ:        "12345678000190",
			RazaoSocial: "Empresa Teste LTDA",
		}

		response := entity.ToConsultaResponse()

		assert.NotNil(t, response)
		assert.Equal(t, "12345678000190", response.CNPJ)
		assert.Equal(t, "Empresa Teste LTDA", response.RazaoSocial)
	})

	t.Run("empty_entity", func(t *testing.T) {
		entity := &LegalEntityFull{}

		response := entity.ToConsultaResponse()

		assert.NotNil(t, response)
		assert.Equal(t, "", response.CNPJ)
		assert.Equal(t, "", response.RazaoSocial)
	})
}

func TestLegalEntityDetails_GetSocioCPFs(t *testing.T) {
	tests := []struct {
		name     string
		entity   *LegalEntityDetails
		expected []string
	}{
		{
			name: "multiple_socios_and_responsavel",
			entity: &LegalEntityDetails{
				Responsavel: struct {
					CPF string `json:"cpf"`
				}{CPF: "12345678901"},
				Socios: []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "98765432100"},
					{CPFSocio: "11122233344"},
				},
			},
			expected: []string{"98765432100", "11122233344", "12345678901"},
		},
		{
			name: "single_socio",
			entity: &LegalEntityDetails{
				Socios: []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678901"},
				},
			},
			expected: []string{"12345678901"},
		},
		{
			name: "no_socios_with_responsavel",
			entity: &LegalEntityDetails{
				Responsavel: struct {
					CPF string `json:"cpf"`
				}{CPF: "12345678901"},
			},
			expected: []string{"12345678901"},
		},
		{
			name:     "nil_socios",
			entity:   &LegalEntityDetails{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entity.GetSocioCPFs()
			assert.ElementsMatch(t, tt.expected, result) // Use ElementsMatch since order may vary
		})
	}
}

func TestCitizenContactInfo_GetEmail(t *testing.T) {
	tests := []struct {
		name     string
		citizen  *CitizenContactInfo
		expected string
	}{
		{
			name: "with_valid_email",
			citizen: &CitizenContactInfo{
				Email: CitizenEmailInfo{
					Indicador: true,
					Principal: struct {
						Valor     string `json:"valor"`
						Origem    string `json:"origem"`
						Sistema   string `json:"sistema"`
						UpdatedAt string `json:"updated_at"`
					}{
						Valor: "test@example.com",
					},
				},
			},
			expected: "test@example.com",
		},
		{
			name: "indicador_false",
			citizen: &CitizenContactInfo{
				Email: CitizenEmailInfo{
					Indicador: false,
					Principal: struct {
						Valor     string `json:"valor"`
						Origem    string `json:"origem"`
						Sistema   string `json:"sistema"`
						UpdatedAt string `json:"updated_at"`
					}{
						Valor: "test@example.com",
					},
				},
			},
			expected: "",
		},
		{
			name:     "empty_citizen",
			citizen:  &CitizenContactInfo{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.citizen.GetEmail()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "valid_email",
			email:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "valid_email_with_subdomain",
			email:    "user@mail.example.com",
			expected: "user@mail.example.com",
		},
		{
			name:     "invalid_email_no_at",
			email:    "testexample.com",
			expected: "",
		},
		{
			name:     "invalid_email_no_domain",
			email:    "test@",
			expected: "",
		},
		{
			name:     "invalid_email_short_tld",
			email:    "test@example.c",
			expected: "",
		},
		{
			name:     "empty_email",
			email:    "",
			expected: "",
		},
		{
			name:     "email_with_spaces",
			email:    "test @example.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCitizenContactInfo_GetCelular(t *testing.T) {
	tests := []struct {
		name     string
		citizen  *CitizenContactInfo
		expected string
	}{
		{
			name: "with_telefone",
			citizen: &CitizenContactInfo{
				Telefone: CitizenTelefoneInfo{
					Indicador: true,
					Principal: struct {
						DDI       string `json:"ddi"`
						DDD       string `json:"ddd"`
						Valor     string `json:"valor"`
						Origem    string `json:"origem"`
						Sistema   string `json:"sistema"`
						UpdatedAt string `json:"updated_at"`
					}{
						DDI:   "55",
						DDD:   "21",
						Valor: "987654321",
					},
				},
			},
			expected: "5521987654321",
		},
		{
			name: "indicador_false",
			citizen: &CitizenContactInfo{
				Telefone: CitizenTelefoneInfo{
					Indicador: false,
				},
			},
			expected: "",
		},
		{
			name:     "empty_citizen",
			citizen:  &CitizenContactInfo{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.citizen.GetCelular()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCitizenContactInfo_GetDataNascimento(t *testing.T) {
	tests := []struct {
		name     string
		citizen  *CitizenContactInfo
		expected string
	}{
		{
			name: "with_nascimento",
			citizen: &CitizenContactInfo{
				Nascimento: &CitizenNascimento{
					Data: "1990-01-01",
				},
			},
			expected: "1990-01-01",
		},
		{
			name: "nil_nascimento",
			citizen: &CitizenContactInfo{
				Nascimento: nil,
			},
			expected: "",
		},
		{
			name:     "empty_citizen",
			citizen:  &CitizenContactInfo{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.citizen.GetDataNascimento()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCitizenContactInfo_GetRaca(t *testing.T) {
	citizen := &CitizenContactInfo{Raca: "Branca"}
	assert.Equal(t, "Branca", citizen.GetRaca())
}

func TestCitizenContactInfo_GetGenero(t *testing.T) {
	t.Run("with_genero", func(t *testing.T) {
		citizen := &CitizenContactInfo{Genero: "Masculino"}
		assert.Equal(t, "Masculino", citizen.GetGenero())
	})

	t.Run("fallback_to_sexo", func(t *testing.T) {
		citizen := &CitizenContactInfo{Genero: "", Sexo: "Feminino"}
		assert.Equal(t, "Feminino", citizen.GetGenero())
	})
}

func TestCitizenContactInfo_GetRendaFamiliar(t *testing.T) {
	citizen := &CitizenContactInfo{RendaFamiliar: "3000"}
	assert.Equal(t, "3000", citizen.GetRendaFamiliar())
}

func TestCitizenContactInfo_GetEscolaridade(t *testing.T) {
	citizen := &CitizenContactInfo{Escolaridade: "Superior Completo"}
	assert.Equal(t, "Superior Completo", citizen.GetEscolaridade())
}

func TestCitizenContactInfo_GetDeficiencia(t *testing.T) {
	citizen := &CitizenContactInfo{Deficiencia: "Visual"}
	assert.Equal(t, "Visual", citizen.GetDeficiencia())
}

func TestCitizenContactInfo_GetEndereco(t *testing.T) {
	t.Run("with_valid_endereco", func(t *testing.T) {
		endereco := &CitizenEnderecoInfo{Indicador: true}
		citizen := &CitizenContactInfo{Endereco: endereco}
		result := citizen.GetEndereco()
		assert.NotNil(t, result)
		assert.True(t, result.Indicador)
	})

	t.Run("indicador_false", func(t *testing.T) {
		endereco := &CitizenEnderecoInfo{Indicador: false}
		citizen := &CitizenContactInfo{Endereco: endereco}
		result := citizen.GetEndereco()
		assert.Nil(t, result)
	})

	t.Run("nil_endereco", func(t *testing.T) {
		citizen := &CitizenContactInfo{Endereco: nil}
		result := citizen.GetEndereco()
		assert.Nil(t, result)
	})
}
