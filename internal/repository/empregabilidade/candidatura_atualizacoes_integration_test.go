package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Executa de verdade a contagem das atualizações das candidaturas, que lê o
// histórico (emp_candidatura_historico) gravado pelas próprias mudanças de
// status e de etapa: só conta o que o recrutador ou a automação mudou depois do
// último evento visto, e marcar como vista não mexe em updated_at.
func TestCandidaturaAtualizacoesIntegration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCandidaturaRepository(tx)
	const cpf = "99999999701"

	var regime empregabilidade.RegimeContratacao
	require.NoError(t, tx.First(&regime).Error)
	var modelo empregabilidade.ModeloTrabalho
	require.NoError(t, tx.First(&modelo).Error)
	empresa := &empregabilidade.Empresa{CNPJ: "11222333000262", NomeFantasia: "Atualizações Integração"}
	require.NoError(t, tx.Create(empresa).Error)
	novaCandidatura := func() (*empregabilidade.Candidatura, []uuid.UUID) {
		t.Helper()
		vaga := &empregabilidade.Vaga{
			Titulo: "Vaga Integração", Descricao: "Descrição", IDContratante: empresa.CNPJ,
			IDRegimeContratacao: regime.ID, IDModeloTrabalho: modelo.ID, Status: empregabilidade.StatusVagaPublicadoAtivo,
		}
		require.NoError(t, tx.Create(vaga).Error)
		var etapas []uuid.UUID
		for ordem, titulo := range []string{"Análise", "Entrevista"} {
			etapa := &empregabilidade.Etapa{IDVaga: vaga.ID, Titulo: titulo, Ordem: ordem + 1}
			require.NoError(t, tx.Create(etapa).Error)
			etapas = append(etapas, etapa.ID)
		}
		c := &empregabilidade.Candidatura{CPF: cpf, IDVaga: vaga.ID, Status: empregabilidade.StatusCandidaturaEnviada, IDEtapaAtual: &etapas[0]}
		_, err := repo.Create(ctx, c)
		require.NoError(t, err)
		return c, etapas
	}
	contar := func() int64 {
		t.Helper()
		total, err := repo.CountAtualizacoesNaoVistas(ctx, cpf)
		require.NoError(t, err)
		return total
	}

	primeira, _ := novaCandidatura()
	segunda, etapasSegunda := novaCandidatura()
	assert.Zero(t, contar(), "a criação (candidatura_criada) não é atualização")

	// Pedir o status ou a etapa que a candidatura já tem não gera evento.
	require.NoError(t, repo.UpdateStatus(ctx, primeira.ID, empregabilidade.StatusCandidaturaEnviada))
	require.NoError(t, repo.UpdateEtapa(ctx, segunda.ID, etapasSegunda[0]))
	assert.Zero(t, contar())

	// O que o próprio candidato provoca não é novidade para ele.
	doCandidato := empregabilidade.ComOrigemEvento(ctx, empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorCandidato})
	require.NoError(t, repo.UpdateEtapa(doCandidato, segunda.ID, etapasSegunda[1]))
	assert.Zero(t, contar())

	require.NoError(t, repo.UpdateStatus(ctx, primeira.ID, empregabilidade.StatusCandidaturaAprovada))
	assert.Equal(t, int64(1), contar())

	require.NoError(t, repo.BulkSaveAndUpdateStatusByVagaID(ctx, segunda.IDVaga, empregabilidade.StatusCandidaturaVagaCongelada))
	assert.Equal(t, int64(2), contar())

	var antes empregabilidade.Candidatura
	require.NoError(t, tx.First(&antes, "id = ?", primeira.ID).Error)
	require.NoError(t, repo.MarcarAtualizacoesComoVistas(ctx, cpf))
	assert.Zero(t, contar())
	var vista empregabilidade.Candidatura
	require.NoError(t, tx.First(&vista, "id = ?", primeira.ID).Error)
	require.NotNil(t, vista.UltimoEventoVisto)
	assert.True(t, antes.UpdatedAt.Equal(vista.UpdatedAt), "marcar como vista não mexe em updated_at")

	// Reativar a vaga devolve o status anterior: evento novo, atualização nova.
	require.NoError(t, repo.BulkRestoreStatusByVagaID(ctx, segunda.IDVaga))
	assert.Equal(t, int64(1), contar())
}
