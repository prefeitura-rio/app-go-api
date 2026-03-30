package models

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOportunidadeMEI_TableName(t *testing.T) {
	oportunidade := &OportunidadeMEI{}
	assert.Equal(t, "oportunidades_mei", oportunidade.TableName())
}

func TestStatusOportunidadeMEI_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   StatusOportunidadeMEI
		expected bool
	}{
		{"valid_draft", StatusOportunidadeDraft, true},
		{"valid_active", StatusOportunidadeActive, true},
		{"valid_expired", StatusOportunidadeExpired, true},
		{"invalid_empty", "", false},
		{"invalid_value", "pending", false},
		{"invalid_uppercase", "DRAFT", false},
		{"invalid_mixed_case", "Active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormaPagamento_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		forma    FormaPagamento
		expected bool
	}{
		{"valid_cheque", FormaPagamentoCheque, true},
		{"valid_dinheiro", FormaPagamentoDinheiro, true},
		{"valid_cartao", FormaPagamentoCartao, true},
		{"valid_pix", FormaPagamentoPix, true},
		{"valid_transferencia", FormaPagamentoTransferencia, true},
		{"valid_empty", "", true},
		{"invalid_lowercase", "pix", false},
		{"invalid_value", "BOLETO", false},
		{"invalid_mixed_case", "Pix", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.forma.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOportunidadeMEI_Validate(t *testing.T) {
	dataExpiracao := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		oportunidade *OportunidadeMEI
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_draft_minimal_fields",
			oportunidade: &OportunidadeMEI{
				Status: StatusOportunidadeDraft,
			},
			expectError: false,
		},
		{
			name: "valid_draft_with_partial_fields",
			oportunidade: &OportunidadeMEI{
				Titulo: "Teste",
				Status: StatusOportunidadeDraft,
			},
			expectError: false,
		},
		{
			name: "valid_active_complete",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Manutenção de Jardins",
				DescricaoServico: "Serviço de poda e jardinagem",
				OrgaoID:          "SECONSERVA",
				CNAEIDs:          pq.StringArray{"8130300", "8121400"},
				Logradouro:       "Praça XV de Novembro",
				Numero:           "S/N",
				Bairro:           "Centro",
				Cidade:           "Rio de Janeiro",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: false,
		},
		{
			name: "invalid_status",
			oportunidade: &OportunidadeMEI{
				Status: "INVALID_STATUS",
			},
			expectError: true,
			errorMsg:    "status inválido",
		},
		{
			name: "active_missing_titulo",
			oportunidade: &OportunidadeMEI{
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "active_empty_titulo",
			oportunidade: &OportunidadeMEI{
				Titulo:           "   ",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "active_missing_descricao",
			oportunidade: &OportunidadeMEI{
				Titulo:        "Título",
				OrgaoID:       "ORGAO",
				CNAEIDs:       pq.StringArray{"8130300"},
				Logradouro:    "Rua",
				Numero:        "123",
				Bairro:        "Bairro",
				Cidade:        "Cidade",
				Estado:        "RJ",
				DataExpiracao: &dataExpiracao,
				Status:        StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "descrição do serviço é obrigatória",
		},
		{
			name: "active_missing_orgao",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "órgão demandante é obrigatório",
		},
		{
			name: "active_missing_cnaes",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "pelo menos um CNAE é obrigatório",
		},
		{
			name: "active_missing_logradouro",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "logradouro é obrigatório",
		},
		{
			name: "active_missing_numero",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "número é obrigatório",
		},
		{
			name: "active_missing_bairro",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "bairro é obrigatório",
		},
		{
			name: "active_missing_cidade",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "cidade é obrigatória",
		},
		{
			name: "active_missing_estado",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "estado é obrigatório",
		},
		{
			name: "active_missing_data_expiracao",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				Status:           StatusOportunidadeActive,
			},
			expectError: true,
			errorMsg:    "data de expiração é obrigatória",
		},
		{
			name: "draft_with_invalid_forma_pagamento",
			oportunidade: &OportunidadeMEI{
				Status:         StatusOportunidadeDraft,
				FormaPagamento: stringPtr("INVALID"),
			},
			expectError: true,
			errorMsg:    "forma de pagamento inválida",
		},
		{
			name: "active_with_valid_forma_pagamento",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
				FormaPagamento:   stringPtr("PIX"),
			},
			expectError: false,
		},
		{
			name: "active_with_invalid_forma_pagamento",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				Status:           StatusOportunidadeActive,
				FormaPagamento:   stringPtr("BOLETO"),
			},
			expectError: true,
			errorMsg:    "forma de pagamento inválida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.oportunidade.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOportunidadeMEI_ValidateForPublish(t *testing.T) {
	dataExpiracao := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		oportunidade *OportunidadeMEI
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_complete_oportunidade",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
			},
			expectError: false,
		},
		{
			name: "missing_all_required_fields",
			oportunidade: &OportunidadeMEI{},
			expectError: true,
			errorMsg:    "título é obrigatório",
		},
		{
			name: "with_valid_forma_pagamento",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				FormaPagamento:   stringPtr("PIX"),
			},
			expectError: false,
		},
		{
			name: "with_empty_forma_pagamento",
			oportunidade: &OportunidadeMEI{
				Titulo:           "Título",
				DescricaoServico: "Descrição",
				OrgaoID:          "ORGAO",
				CNAEIDs:          pq.StringArray{"8130300"},
				Logradouro:       "Rua",
				Numero:           "123",
				Bairro:           "Bairro",
				Cidade:           "Cidade",
				Estado:           "RJ",
				DataExpiracao:    &dataExpiracao,
				FormaPagamento:   stringPtr(""),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.oportunidade.ValidateForPublish()
			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOportunidadeMEI_UpdateStatusBasedOnExpiration(t *testing.T) {
	tests := []struct {
		name           string
		oportunidade   *OportunidadeMEI
		expectedStatus StatusOportunidadeMEI
	}{
		{
			name: "nil_data_expiracao_no_change",
			oportunidade: &OportunidadeMEI{
				Status:        StatusOportunidadeActive,
				DataExpiracao: nil,
			},
			expectedStatus: StatusOportunidadeActive,
		},
		{
			name: "expired_date_active_to_expired",
			oportunidade: &OportunidadeMEI{
				Status:        StatusOportunidadeActive,
				DataExpiracao: timePtr(time.Now().Add(-24 * time.Hour)),
			},
			expectedStatus: StatusOportunidadeExpired,
		},
		{
			name: "expired_date_draft_unchanged",
			oportunidade: &OportunidadeMEI{
				Status:        StatusOportunidadeDraft,
				DataExpiracao: timePtr(time.Now().Add(-24 * time.Hour)),
			},
			expectedStatus: StatusOportunidadeDraft,
		},
		{
			name: "future_date_active_unchanged",
			oportunidade: &OportunidadeMEI{
				Status:        StatusOportunidadeActive,
				DataExpiracao: timePtr(time.Now().Add(24 * time.Hour)),
			},
			expectedStatus: StatusOportunidadeActive,
		},
		{
			name: "future_date_expired_to_active",
			oportunidade: &OportunidadeMEI{
				Status:        StatusOportunidadeExpired,
				DataExpiracao: timePtr(time.Now().Add(24 * time.Hour)),
			},
			expectedStatus: StatusOportunidadeActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.oportunidade.UpdateStatusBasedOnExpiration()
			assert.Equal(t, tt.expectedStatus, tt.oportunidade.Status)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
