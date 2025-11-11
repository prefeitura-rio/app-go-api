package models_test

import (
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestFormaPagamentoValidation(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		shouldBeValid bool
	}{
		{"Valid CHEQUE", "CHEQUE", true},
		{"Valid DINHEIRO", "DINHEIRO", true},
		{"Valid CARTAO", "CARTAO", true},
		{"Valid PIX", "PIX", true},
		{"Valid TRANSFERENCIA", "TRANSFERENCIA", true},
		{"Empty string (optional)", "", true},
		{"Invalid value", "BITCOIN", false},
		{"Invalid lowercase", "pix", false},
		{"Invalid mixed case", "Pix", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := models.FormaPagamento(tt.value).IsValid()
			if isValid != tt.shouldBeValid {
				t.Errorf("FormaPagamento(%q).IsValid() = %v, want %v",
					tt.value, isValid, tt.shouldBeValid)
			}
		})
	}
}

func TestOportunidadeMEIFormaPagamentoField(t *testing.T) {
	// Teste com forma_pagamento como nil (opcional)
	t.Run("FormaPagamento nil", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Titulo:           "Teste",
			DescricaoServico: "Descrição teste",
			OrgaoID:          "1",
			CNAEID:           "1",
			Logradouro:       "Rua Teste",
			Numero:           "123",
			Bairro:           "Centro",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			Status:           models.StatusOportunidadeActive,
			FormaPagamento:   nil, // Campo opcional
		}

		// Com DataExpiracao para passar a validação
		now := time.Now().Add(24 * time.Hour)
		oportunidade.DataExpiracao = &now

		err := oportunidade.Validate()
		if err != nil {
			t.Errorf("Unexpected error for nil FormaPagamento: %v", err)
		}
	})

	// Teste com forma_pagamento como string vazia
	t.Run("FormaPagamento empty string", func(t *testing.T) {
		emptyStr := ""
		oportunidade := &models.OportunidadeMEI{
			Titulo:           "Teste",
			DescricaoServico: "Descrição teste",
			OrgaoID:          "1",
			CNAEID:           "1",
			Logradouro:       "Rua Teste",
			Numero:           "123",
			Bairro:           "Centro",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			Status:           models.StatusOportunidadeActive,
			FormaPagamento:   &emptyStr, // String vazia
		}

		// Com DataExpiracao para passar a validação
		now := time.Now().Add(24 * time.Hour)
		oportunidade.DataExpiracao = &now

		err := oportunidade.Validate()
		if err != nil {
			t.Errorf("Unexpected error for empty FormaPagamento: %v", err)
		}
	})

	// Teste com forma_pagamento válida
	t.Run("FormaPagamento valid PIX", func(t *testing.T) {
		pixStr := "PIX"
		oportunidade := &models.OportunidadeMEI{
			Titulo:           "Teste",
			DescricaoServico: "Descrição teste",
			OrgaoID:          "1",
			CNAEID:           "1",
			Logradouro:       "Rua Teste",
			Numero:           "123",
			Bairro:           "Centro",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			Status:           models.StatusOportunidadeActive,
			FormaPagamento:   &pixStr, // PIX válido
		}

		// Com DataExpiracao para passar a validação
		now := time.Now().Add(24 * time.Hour)
		oportunidade.DataExpiracao = &now

		err := oportunidade.Validate()
		if err != nil {
			t.Errorf("Unexpected error for valid FormaPagamento PIX: %v", err)
		}
	})

	// Teste com forma_pagamento inválida
	t.Run("FormaPagamento invalid", func(t *testing.T) {
		invalidStr := "BITCOIN"
		oportunidade := &models.OportunidadeMEI{
			Titulo:           "Teste",
			DescricaoServico: "Descrição teste",
			OrgaoID:          "1",
			CNAEID:           "1",
			Logradouro:       "Rua Teste",
			Numero:           "123",
			Bairro:           "Centro",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			Status:           models.StatusOportunidadeActive,
			FormaPagamento:   &invalidStr, // Valor inválido
		}

		// Com DataExpiracao para passar a validação
		now := time.Now().Add(24 * time.Hour)
		oportunidade.DataExpiracao = &now

		err := oportunidade.Validate()
		if err == nil {
			t.Error("Expected error for invalid FormaPagamento, got nil")
		} else if err.Error() != "forma de pagamento inválida" {
			t.Errorf("Expected 'forma de pagamento inválida', got %v", err)
		}
	})
}

func TestOportunidadeMEIDraftValidation(t *testing.T) {
	// Teste: Rascunho com campos vazios deve passar validação
	t.Run("Draft with empty fields", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Status: models.StatusOportunidadeDraft,
			// Todos os outros campos vazios
		}

		err := oportunidade.Validate()
		if err != nil {
			t.Errorf("Draft should allow empty fields, got error: %v", err)
		}
	})

	// Teste: Rascunho com apenas título deve passar validação
	t.Run("Draft with only title", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Status: models.StatusOportunidadeDraft,
			Titulo: "Oportunidade em rascunho",
			// Outros campos vazios
		}

		err := oportunidade.Validate()
		if err != nil {
			t.Errorf("Draft should allow partial data, got error: %v", err)
		}
	})

	// Teste: Rascunho com forma_pagamento inválida deve falhar
	t.Run("Draft with invalid FormaPagamento", func(t *testing.T) {
		invalidStr := "BITCOIN"
		oportunidade := &models.OportunidadeMEI{
			Status:         models.StatusOportunidadeDraft,
			Titulo:         "Teste",
			FormaPagamento: &invalidStr,
		}

		err := oportunidade.Validate()
		if err == nil {
			t.Error("Expected error for invalid FormaPagamento in draft")
		} else if err.Error() != "forma de pagamento inválida" {
			t.Errorf("Expected 'forma de pagamento inválida', got %v", err)
		}
	})

	// Teste: Active status com campos vazios deve falhar
	t.Run("Active with empty fields should fail", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Status: models.StatusOportunidadeActive,
			// Campos obrigatórios vazios
		}

		err := oportunidade.Validate()
		if err == nil {
			t.Error("Active status should require all mandatory fields")
		}
	})

	// Teste: ValidateForPublish com campos vazios deve falhar
	t.Run("ValidateForPublish with empty fields", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Status: models.StatusOportunidadeDraft,
			Titulo: "Apenas título",
		}

		err := oportunidade.ValidateForPublish()
		if err == nil {
			t.Error("ValidateForPublish should require all mandatory fields")
		}
	})

	// Teste: ValidateForPublish com todos os campos deve passar
	t.Run("ValidateForPublish with all fields", func(t *testing.T) {
		now := time.Now().Add(24 * time.Hour)
		pixStr := "PIX"
		oportunidade := &models.OportunidadeMEI{
			Status:           models.StatusOportunidadeDraft,
			Titulo:           "Teste Completo",
			DescricaoServico: "Descrição completa do serviço",
			OrgaoID:          "1",
			CNAEID:           "1",
			Logradouro:       "Rua Teste",
			Numero:           "123",
			Bairro:           "Centro",
			Cidade:           "Rio de Janeiro",
			Estado:           "RJ",
			FormaPagamento:   &pixStr,
			DataExpiracao:    &now,
		}

		err := oportunidade.ValidateForPublish()
		if err != nil {
			t.Errorf("ValidateForPublish should pass with all fields, got error: %v", err)
		}
	})
}
