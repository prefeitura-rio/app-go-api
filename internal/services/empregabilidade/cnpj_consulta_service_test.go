package empregabilidade_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestLegalEntityFull_ToConsultaResponse(t *testing.T) {
	t.Run("Converts full entity to simplified response", func(t *testing.T) {
		nomeFantasia := "Test Company"
		porteDesc := "Pequeno Porte"
		situacaoDesc := "Ativa"

		legalEntity := &models.LegalEntityFull{
			CNPJ:         "12345678000195",
			RazaoSocial:  "Test Company LTDA",
			NomeFantasia: &nomeFantasia,
			Porte: &struct {
				ID        string `json:"id"`
				Descricao string `json:"descricao"`
			}{
				ID:        "1",
				Descricao: porteDesc,
			},
			SituacaoCadastral: &struct {
				ID        string `json:"id"`
				Descricao string `json:"descricao"`
				Data      string `json:"data"`
				Motivo    string `json:"motivo"`
			}{
				ID:        "2",
				Descricao: situacaoDesc,
			},
			Endereco: &struct {
				CEP            string `json:"cep"`
				UF             string `json:"uf"`
				Municipio      string `json:"municipio"`
				MunicipioID    string `json:"municipio_id"`
				Bairro         string `json:"bairro"`
				TipoLogradouro string `json:"tipo_logradouro"`
				Logradouro     string `json:"logradouro"`
				Numero         string `json:"numero"`
				Complemento    string `json:"complemento"`
			}{
				CEP:        "12345678",
				UF:         "RJ",
				Municipio:  "Rio de Janeiro",
				Bairro:     "Centro",
				Logradouro: "Rua Teste",
				Numero:     "123",
			},
			Contato: &struct {
				Telefone1 string `json:"telefone_1"`
				Telefone2 string `json:"telefone_2"`
				Fax       string `json:"fax"`
				Email     string `json:"email"`
			}{
				Telefone1: "21999999999",
				Email:     "test@test.com",
			},
		}

		response := legalEntity.ToConsultaResponse()

		if response.CNPJ != "12345678000195" {
			t.Errorf("Expected CNPJ '12345678000195', got '%s'", response.CNPJ)
		}

		if response.RazaoSocial != "Test Company LTDA" {
			t.Errorf("Expected RazaoSocial 'Test Company LTDA', got '%s'", response.RazaoSocial)
		}

		if response.NomeFantasia == nil || *response.NomeFantasia != "Test Company" {
			t.Error("Expected NomeFantasia to be 'Test Company'")
		}

		if response.Porte == nil || *response.Porte != "Pequeno Porte" {
			t.Error("Expected Porte to be 'Pequeno Porte'")
		}

		if response.SituacaoCadastral == nil || *response.SituacaoCadastral != "Ativa" {
			t.Error("Expected SituacaoCadastral to be 'Ativa'")
		}

		if response.Endereco == nil {
			t.Error("Expected Endereco to be set")
		} else {
			if response.Endereco.CEP != "12345678" {
				t.Errorf("Expected CEP '12345678', got '%s'", response.Endereco.CEP)
			}
			if response.Endereco.UF != "RJ" {
				t.Errorf("Expected UF 'RJ', got '%s'", response.Endereco.UF)
			}
		}

		if response.Contato == nil {
			t.Error("Expected Contato to be set")
		} else {
			if response.Contato.Telefone != "21999999999" {
				t.Errorf("Expected Telefone '21999999999', got '%s'", response.Contato.Telefone)
			}
			if response.Contato.Email != "test@test.com" {
				t.Errorf("Expected Email 'test@test.com', got '%s'", response.Contato.Email)
			}
		}
	})

	t.Run("Handles nil optional fields", func(t *testing.T) {
		legalEntity := &models.LegalEntityFull{
			CNPJ:        "12345678000195",
			RazaoSocial: "Test Company LTDA",
		}

		response := legalEntity.ToConsultaResponse()

		if response.CNPJ != "12345678000195" {
			t.Errorf("Expected CNPJ '12345678000195', got '%s'", response.CNPJ)
		}

		if response.NomeFantasia != nil {
			t.Error("Expected NomeFantasia to be nil")
		}

		if response.Porte != nil {
			t.Error("Expected Porte to be nil")
		}

		if response.Endereco != nil {
			t.Error("Expected Endereco to be nil")
		}

		if response.Contato != nil {
			t.Error("Expected Contato to be nil")
		}
	})

	t.Run("Falls back to Telefone2 when Telefone1 is empty", func(t *testing.T) {
		legalEntity := &models.LegalEntityFull{
			CNPJ:        "12345678000195",
			RazaoSocial: "Test Company LTDA",
			Contato: &struct {
				Telefone1 string `json:"telefone_1"`
				Telefone2 string `json:"telefone_2"`
				Fax       string `json:"fax"`
				Email     string `json:"email"`
			}{
				Telefone1: "",
				Telefone2: "21888888888",
				Email:     "test@test.com",
			},
		}

		response := legalEntity.ToConsultaResponse()

		if response.Contato == nil {
			t.Error("Expected Contato to be set")
		} else if response.Contato.Telefone != "21888888888" {
			t.Errorf("Expected Telefone '21888888888', got '%s'", response.Contato.Telefone)
		}
	})
}
