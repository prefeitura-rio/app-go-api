package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr é um helper para criar ponteiros de inteiros inline
func ptr(i int) *int { return &i }

// newVagaServiceForQtd monta o service com mocks prontos e uma empresa válida no CNPJ fornecido
func newVagaServiceForQtd(cnpj string) (*services.VagaService, *MockVagaRepoForService, *MockEmpresaRepoForService) {
	vagaRepo := NewMockVagaRepoForService()
	empresaRepo := NewMockEmpresaRepoForService()
	empresaRepo.empresas[cnpj] = &empregabilidade.Empresa{CNPJ: cnpj, RazaoSocial: "Empresa Teste"}
	svc := services.NewVagaServiceWithInterfaces(vagaRepo, empresaRepo, nil)
	return svc, vagaRepo, empresaRepo
}

// ─────────────────────────────────────────────────────────────────────────────
// Create — campo quantidade_estimada_contratacoes
// ─────────────────────────────────────────────────────────────────────────────

func TestVagaService_Create_QuantidadeEstimada_Nil(t *testing.T) {
	svc, _, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:        "Repositor",
		IDContratante: "12.345.678/0001-90",
		// campo ausente = nil
	}
	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	assert.Nil(t, vaga.QuantidadeEstimadaContratacoes)
}

func TestVagaService_Create_QuantidadeEstimada_Positivo(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:                         "Repositor",
		IDContratante:                  "12.345.678/0001-90",
		QuantidadeEstimadaContratacoes: ptr(300),
	}
	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	stored := vagaRepo.vagas[id]
	require.NotNil(t, stored.QuantidadeEstimadaContratacoes)
	assert.Equal(t, 300, *stored.QuantidadeEstimadaContratacoes)
}

func TestVagaService_Create_QuantidadeEstimada_Um(t *testing.T) {
	// valor mínimo positivo válido = 1
	svc, _, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:                         "Operador de Caixa",
		IDContratante:                  "12.345.678/0001-90",
		QuantidadeEstimadaContratacoes: ptr(1),
	}
	_, err := svc.Create(context.Background(), vaga)
	assert.NoError(t, err)
}

func TestVagaService_Create_QuantidadeEstimada_Zero_Rejeitado(t *testing.T) {
	svc, _, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:                         "Operador de Caixa",
		IDContratante:                  "12.345.678/0001-90",
		QuantidadeEstimadaContratacoes: ptr(0),
	}
	_, err := svc.Create(context.Background(), vaga)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantidade_estimada_contratacoes deve ser um número inteiro positivo")
}

func TestVagaService_Create_QuantidadeEstimada_Negativo_Rejeitado(t *testing.T) {
	svc, _, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:                         "Deposista",
		IDContratante:                  "12.345.678/0001-90",
		QuantidadeEstimadaContratacoes: ptr(-10),
	}
	_, err := svc.Create(context.Background(), vaga)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantidade_estimada_contratacoes deve ser um número inteiro positivo")
}

// ─────────────────────────────────────────────────────────────────────────────
// Update — campo quantidade_estimada_contratacoes
// ─────────────────────────────────────────────────────────────────────────────

func TestVagaService_Update_QuantidadeEstimada_Positivo(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")

	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:     vagaID,
		Titulo: "Repositor",
		Status: empregabilidade.StatusVagaEmEdicao,
	}

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                             vagaID,
		Titulo:                         "Repositor",
		QuantidadeEstimadaContratacoes: ptr(250),
	})
	require.NoError(t, err)
	stored := vagaRepo.vagas[vagaID]
	require.NotNil(t, stored.QuantidadeEstimadaContratacoes)
	assert.Equal(t, 250, *stored.QuantidadeEstimadaContratacoes)
}

func TestVagaService_Update_QuantidadeEstimada_ParaNil(t *testing.T) {
	// permite limpar o campo passando nil (omitempty no JSON faz isso naturalmente)
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")

	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:                             vagaID,
		Titulo:                         "Repositor",
		Status:                         empregabilidade.StatusVagaEmEdicao,
		QuantidadeEstimadaContratacoes: ptr(300),
	}

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                             vagaID,
		Titulo:                         "Repositor",
		QuantidadeEstimadaContratacoes: nil,
	})
	require.NoError(t, err)
	assert.Nil(t, vagaRepo.vagas[vagaID].QuantidadeEstimadaContratacoes)
}

func TestVagaService_Update_QuantidadeEstimada_Zero_Rejeitado(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")

	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:     vagaID,
		Titulo: "Repositor",
		Status: empregabilidade.StatusVagaEmEdicao,
	}

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                             vagaID,
		Titulo:                         "Repositor",
		QuantidadeEstimadaContratacoes: ptr(0),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantidade_estimada_contratacoes deve ser um número inteiro positivo")
}

func TestVagaService_Update_QuantidadeEstimada_Negativo_Rejeitado(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")

	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:     vagaID,
		Titulo: "Deposista",
		Status: empregabilidade.StatusVagaEmEdicao,
	}

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                             vagaID,
		Titulo:                         "Deposista",
		QuantidadeEstimadaContratacoes: ptr(-1),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantidade_estimada_contratacoes deve ser um número inteiro positivo")
}

// validação ocorre ANTES de checar a existência da vaga no repo
func TestVagaService_Update_QuantidadeEstimada_ValidacaoAntesDoGetByID(t *testing.T) {
	svc, _, _ := newVagaServiceForQtd("12.345.678/0001-90")

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                             uuid.New(), // não existe no repo
		Titulo:                         "Qualquer",
		QuantidadeEstimadaContratacoes: ptr(-5),
	})
	require.Error(t, err)
	// deve falhar na validação, não em "vaga não encontrada"
	assert.Contains(t, err.Error(), "quantidade_estimada_contratacoes deve ser um número inteiro positivo")
	assert.NotContains(t, err.Error(), "vaga não encontrada")
}

// ─────────────────────────────────────────────────────────────────────────────
// Limites e valores extremos
// ─────────────────────────────────────────────────────────────────────────────

func TestVagaService_Create_QuantidadeEstimada_ValorGrande(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd("12.345.678/0001-90")
	vaga := &empregabilidade.Vaga{
		Titulo:                         "Repositor",
		IDContratante:                  "12.345.678/0001-90",
		QuantidadeEstimadaContratacoes: ptr(10000),
	}
	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.Equal(t, 10000, *vagaRepo.vagas[id].QuantidadeEstimadaContratacoes)
}
