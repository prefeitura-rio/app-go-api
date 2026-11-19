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

// O banco segue a regra das vagas: o recrutador só vê o currículo disponível
// que tem o mínimo para se candidatar. Roda o SQL de verdade contra as
// situações atuais do seed.
func TestBancoCurriculosIntegration_SomenteCurriculoMinimo(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	situacao := func(descricao string) *uuid.UUID {
		var s empregabilidade.SituacaoAtual
		require.NoError(t, tx.Where("descricao = ?", descricao).First(&s).Error, descricao)
		return &s.ID
	}
	var outra empregabilidade.SituacaoAtual
	require.NoError(t, tx.Where("descricao NOT IN ?", empregabilidade.SituacoesProcurandoEmprego).First(&outra).Error)
	desempregado := situacao("Desempregado(a)")

	const (
		cpfCompleto          = "99999999811"
		cpfSemTermos         = "99999999812"
		cpfSemTempo          = "99999999813"
		cpfDesempregadoTempo = "99999999814"
		cpfSemSituacao       = "99999999815"
		cpfOcultou           = "99999999816"
	)
	termos := func(cpf string) {
		require.NoError(t, tx.Create(&empregabilidade.TermosUso{CPF: cpf, UserConsent: true}).Error)
	}
	situacaoDe := func(cpf string, id *uuid.UUID, tempo string) {
		require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpf, IDSituacao: id, TempoProcurandoEmprego: tempo}))
	}

	situacaoDe(cpfCompleto, &outra.ID, "")
	termos(cpfCompleto)
	situacaoDe(cpfSemTermos, &outra.ID, "")
	situacaoDe(cpfSemTempo, desempregado, "")
	termos(cpfSemTempo)
	situacaoDe(cpfDesempregadoTempo, desempregado, "Mais de 1 ano")
	termos(cpfDesempregadoTempo)
	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfSemSituacao, []*empregabilidade.CurriculoExperiencia{{Cargo: "Auxiliar", Empresa: "Empresa"}}))
	termos(cpfSemSituacao)
	situacaoDe(cpfOcultou, &outra.ID, "")
	termos(cpfOcultou)
	require.NoError(t, tx.Exec("UPDATE emp_curriculos SET disponivel_banco = FALSE WHERE cpf = ?", cpfOcultou).Error)

	items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "999999998", SomenteDisponiveis: true}, 1, 20)
	require.NoError(t, err)

	cpfs := make([]string, 0, len(items))
	for _, item := range items {
		cpfs = append(cpfs, item.CPF)
	}
	assert.ElementsMatch(t, []string{cpfCompleto, cpfDesempregadoTempo}, cpfs)
	assert.Equal(t, int64(2), total)
}
