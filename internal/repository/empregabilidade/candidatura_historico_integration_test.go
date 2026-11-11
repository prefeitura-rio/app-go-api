package empregabilidade_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Consultas de análise do histórico (DEV-532), as mesmas de
// docs/candidatura-historico.md: o teste prova que a tabela responde ao funil
// por etapa, ao tempo médio em cada etapa e aos reprovados por etapa/motivo.
const (
	sqlFunilPorEtapa = `
SELECT h.id_etapa_nova AS id_etapa, COUNT(DISTINCT h.id_candidatura) AS candidaturas
FROM emp_candidatura_historico h
WHERE h.id_vaga = ?
  AND h.id_etapa_nova IS NOT NULL
  AND h.tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa')
GROUP BY h.id_etapa_nova`

	sqlTempoMedioPorEtapa = `
WITH eventos AS (
    SELECT tipo_evento, id_etapa_nova, created_at,
           LEAD(created_at) OVER (PARTITION BY id_candidatura ORDER BY created_at, id) AS saiu_em
    FROM emp_candidatura_historico
    WHERE id_vaga = ?
      AND tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa', 'etapa_removida',
                          'candidatura_aprovada', 'candidatura_reprovada', 'candidatura_cancelada')
)
SELECT id_etapa_nova AS id_etapa,
       EXTRACT(EPOCH FROM AVG(saiu_em - created_at)) AS segundos_medios,
       COUNT(*) AS passagens
FROM eventos
WHERE tipo_evento IN ('candidatura_criada', 'mudanca_etapa', 'volta_etapa')
  AND id_etapa_nova IS NOT NULL
  AND saiu_em IS NOT NULL
GROUP BY id_etapa_nova`

	sqlReprovadosPorEtapaMotivo = `
SELECT id_etapa_nova AS id_etapa, ator, COALESCE(motivo, 'sem_motivo') AS motivo, COUNT(*) AS reprovacoes
FROM emp_candidatura_historico
WHERE id_vaga = ? AND tipo_evento = 'candidatura_reprovada'
GROUP BY id_etapa_nova, ator, COALESCE(motivo, 'sem_motivo')`
)

func eventosDaCandidatura(t *testing.T, tx *gorm.DB, id uuid.UUID) []empregabilidade.CandidaturaHistorico {
	t.Helper()
	var eventos []empregabilidade.CandidaturaHistorico
	require.NoError(t, tx.Where("id_candidatura = ?", id).Order("id").Find(&eventos).Error)
	return eventos
}

func etapaDeTeste(t *testing.T, tx *gorm.DB, vaga string, titulo string, ordem int) uuid.UUID {
	t.Helper()
	var id string
	require.NoError(t, tx.Raw(`INSERT INTO emp_etapas (id_vaga, titulo, ordem, tipo_etapa) VALUES (?, ?, ?, ?) RETURNING id`,
		vaga, titulo, ordem, "tipo_"+titulo).Scan(&id).Error)
	return uuid.MustParse(id)
}

func TestCandidaturaHistoricoIntegration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	repo := emprepo.NewCandidaturaRepository(tx)
	vagaTexto := vagaDeTesteParticipacao(t, tx)
	vaga := uuid.MustParse(vagaTexto)
	triagem := etapaDeTeste(t, tx, vagaTexto, "triagem", 1)
	entrevista := etapaDeTeste(t, tx, vagaTexto, "entrevista", 2)
	candidato := empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorCandidato})
	operador := context.Background()

	const cpfNovo = "90000000101"
	nova := &empregabilidade.Candidatura{CPF: cpfNovo, IDVaga: vaga, Status: empregabilidade.StatusCandidaturaEnviada}

	t.Run("candidatura nova nasce com candidatura_criada", func(t *testing.T) {
		_, err := repo.Create(candidato, nova)
		require.NoError(t, err)

		eventos := eventosDaCandidatura(t, tx, nova.ID)
		require.Len(t, eventos, 1)
		assert.Equal(t, empregabilidade.EventoCandidaturaCriada, eventos[0].TipoEvento)
		assert.Equal(t, empregabilidade.AtorCandidato, eventos[0].Ator)
		assert.Nil(t, eventos[0].StatusAnterior)
		assert.Equal(t, empregabilidade.StatusCandidaturaEnviada, eventos[0].StatusNovo)
		assert.Equal(t, vaga, eventos[0].IDVaga)
	})

	t.Run("avanço, volta e mesma etapa", func(t *testing.T) {
		require.NoError(t, repo.UpdateEtapa(operador, nova.ID, triagem))
		time.Sleep(20 * time.Millisecond)
		require.NoError(t, repo.UpdateEtapa(operador, nova.ID, entrevista))
		time.Sleep(20 * time.Millisecond)
		require.NoError(t, repo.UpdateEtapa(operador, nova.ID, triagem))
		require.NoError(t, repo.UpdateEtapa(operador, nova.ID, triagem)) // sem mudança

		eventos := eventosDaCandidatura(t, tx, nova.ID)
		require.Len(t, eventos, 4)
		assert.Equal(t, empregabilidade.EventoMudancaEtapa, eventos[1].TipoEvento)
		assert.Nil(t, eventos[1].IDEtapaAnterior)
		assert.Equal(t, triagem, *eventos[1].IDEtapaNova)
		assert.Equal(t, empregabilidade.EventoMudancaEtapa, eventos[2].TipoEvento)
		assert.Equal(t, triagem, *eventos[2].IDEtapaAnterior)
		assert.Equal(t, entrevista, *eventos[2].IDEtapaNova)
		assert.Equal(t, empregabilidade.EventoVoltaEtapa, eventos[3].TipoEvento)
		assert.Equal(t, entrevista, *eventos[3].IDEtapaAnterior)
		assert.Equal(t, triagem, *eventos[3].IDEtapaNova)
	})

	t.Run("reprovação automática com motivo e aprovação", func(t *testing.T) {
		automacao := empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao, Motivo: "criterio_teste"})
		require.NoError(t, repo.UpdateStatus(automacao, nova.ID, empregabilidade.StatusCandidaturaReprovada))
		require.NoError(t, repo.UpdateStatus(operador, nova.ID, empregabilidade.StatusCandidaturaAprovada))

		eventos := eventosDaCandidatura(t, tx, nova.ID)
		require.Len(t, eventos, 6)
		reprovacao := eventos[4]
		assert.Equal(t, empregabilidade.EventoCandidaturaReprovada, reprovacao.TipoEvento)
		assert.Equal(t, empregabilidade.AtorAutomacao, reprovacao.Ator)
		require.NotNil(t, reprovacao.Motivo)
		assert.Equal(t, "criterio_teste", *reprovacao.Motivo)
		assert.Equal(t, triagem, *reprovacao.IDEtapaNova)
		assert.Equal(t, empregabilidade.EventoCandidaturaAprovada, eventos[5].TipoEvento)
		assert.Equal(t, empregabilidade.StatusCandidaturaReprovada, *eventos[5].StatusAnterior)
	})

	const cpfAntigo = "90000000102"
	var antiga uuid.UUID
	var antigaTexto string

	t.Run("candidatura antiga: o primeiro evento traz o estado de origem, sem migração", func(t *testing.T) {
		require.NoError(t, tx.Raw(`INSERT INTO emp_candidaturas (cpf, id_vaga, status, id_etapa_atual) VALUES (?, ?, 'reprovada', ?) RETURNING id`,
			cpfAntigo, vaga, entrevista).Scan(&antigaTexto).Error)
		antiga = uuid.MustParse(antigaTexto)
		assert.Empty(t, eventosDaCandidatura(t, tx, antiga), "nada é migrado retroativamente")

		require.NoError(t, repo.UpdateStatus(operador, antiga, empregabilidade.StatusCandidaturaEnviada))

		eventos := eventosDaCandidatura(t, tx, antiga)
		require.Len(t, eventos, 1)
		assert.Equal(t, empregabilidade.EventoStatusAlterado, eventos[0].TipoEvento)
		assert.Equal(t, empregabilidade.StatusCandidaturaReprovada, *eventos[0].StatusAnterior)
		assert.Equal(t, empregabilidade.StatusCandidaturaEnviada, eventos[0].StatusNovo)
		assert.Equal(t, entrevista, *eventos[0].IDEtapaAnterior)
		assert.Equal(t, entrevista, *eventos[0].IDEtapaNova)
	})

	t.Run("congelar e descongelar a vaga: um evento da automação por candidatura", func(t *testing.T) {
		congelar := empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao, Motivo: empregabilidade.MotivoVagaCongelada})
		descongelar := empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao, Motivo: empregabilidade.MotivoVagaDescongelada})
		require.NoError(t, repo.BulkSaveAndUpdateStatusByVagaID(congelar, vaga, empregabilidade.StatusCandidaturaVagaCongelada))
		require.NoError(t, repo.BulkRestoreStatusByVagaID(descongelar, vaga))

		eventos := eventosDaCandidatura(t, tx, antiga)
		require.Len(t, eventos, 3)
		assert.Equal(t, empregabilidade.StatusCandidaturaVagaCongelada, eventos[1].StatusNovo)
		assert.Equal(t, empregabilidade.MotivoVagaCongelada, *eventos[1].Motivo)
		assert.Equal(t, empregabilidade.StatusCandidaturaEnviada, eventos[2].StatusNovo)
		assert.Equal(t, empregabilidade.MotivoVagaDescongelada, *eventos[2].Motivo)
		assert.Len(t, eventosDaCandidatura(t, tx, nova.ID), 8)
	})

	t.Run("status em lote e etapa em lote", func(t *testing.T) {
		resultado, err := repo.BulkUpdateStatus(operador, vaga, []string{cpfNovo, cpfAntigo, cpfNovo}, empregabilidade.StatusCandidaturaReprovada)
		require.NoError(t, err)
		assert.Equal(t, 2, resultado.Updated)
		require.NoError(t, repo.BulkUpdateEtapa(operador, []uuid.UUID{nova.ID, antiga}, entrevista))

		novos := eventosDaCandidatura(t, tx, nova.ID)
		require.Len(t, novos, 10)
		assert.Equal(t, empregabilidade.EventoCandidaturaReprovada, novos[8].TipoEvento)
		assert.Equal(t, empregabilidade.EventoMudancaEtapa, novos[9].TipoEvento)
		antigos := eventosDaCandidatura(t, tx, antiga)
		require.Len(t, antigos, 4, "a antiga já estava na entrevista: só o evento do status")
	})

	t.Run("etapa excluída da vaga deixa a candidatura sem etapa com etapa_removida", func(t *testing.T) {
		vagaRepo := emprepo.NewVagaRepository(tx)
		var atual empregabilidade.Vaga
		require.NoError(t, tx.First(&atual, "id = ?", vaga).Error)
		atual.Etapas = []empregabilidade.Etapa{{ID: triagem, Titulo: "triagem", Ordem: 1}}
		require.NoError(t, vagaRepo.UpdateWithAssociations(context.Background(), &atual))

		eventos := eventosDaCandidatura(t, tx, nova.ID)
		require.Len(t, eventos, 11)
		assert.Equal(t, empregabilidade.EventoEtapaRemovida, eventos[10].TipoEvento)
		assert.Equal(t, empregabilidade.AtorAutomacao, eventos[10].Ator)
		assert.Equal(t, entrevista, *eventos[10].IDEtapaAnterior)
		assert.Nil(t, eventos[10].IDEtapaNova)
		antigos := eventosDaCandidatura(t, tx, antiga)
		require.Len(t, antigos, 5)
		assert.Equal(t, empregabilidade.EventoEtapaRemovida, antigos[4].TipoEvento)

		var tipo *string
		require.NoError(t, tx.Raw(`SELECT tipo_etapa FROM emp_etapas WHERE id = ?`, triagem).Scan(&tipo).Error)
		require.NotNil(t, tipo, "etapa reenviada sem tipo_etapa mantém o tipo gravado")
		assert.Equal(t, "tipo_triagem", *tipo)
	})

	t.Run("cancelamento", func(t *testing.T) {
		require.NoError(t, repo.Delete(candidato, antiga))
		eventos := eventosDaCandidatura(t, tx, antiga)
		require.Len(t, eventos, 6)
		assert.Equal(t, empregabilidade.EventoCandidaturaCancelada, eventos[5].TipoEvento)
		assert.Equal(t, empregabilidade.AtorCandidato, eventos[5].Ator)
		assert.Nil(t, eventos[5].IDEtapaNova)
	})

	t.Run("edição genérica com status velho não desfaz a transição", func(t *testing.T) {
		velha := *nova
		velha.Status = empregabilidade.StatusCandidaturaEnviada
		velha.IDEtapaAtual = &triagem
		velha.Vaga, velha.EtapaAtual = nil, nil
		require.NoError(t, repo.Update(operador, &velha))

		var status string
		require.NoError(t, tx.Raw(`SELECT status FROM emp_candidaturas WHERE id = ?`, nova.ID).Scan(&status).Error)
		assert.Equal(t, string(empregabilidade.StatusCandidaturaReprovada), status)
		assert.Len(t, eventosDaCandidatura(t, tx, nova.ID), 11, "a edição genérica não gera evento")
	})

	t.Run("o histórico não pode ser editado nem apagado", func(t *testing.T) {
		for _, comando := range []string{
			`UPDATE emp_candidatura_historico SET ator = 'operador' WHERE id_vaga = ?`,
			`DELETE FROM emp_candidatura_historico WHERE id_vaga = ?`,
		} {
			require.NoError(t, tx.SavePoint("antes_da_edicao").Error)
			err := tx.Exec(comando, vaga).Error
			require.Error(t, err, comando)
			assert.Contains(t, err.Error(), "só de inclusão")
			require.NoError(t, tx.RollbackTo("antes_da_edicao").Error)
		}
		require.NoError(t, tx.SavePoint("antes_do_truncate").Error)
		assert.Error(t, tx.Exec(`TRUNCATE emp_candidatura_historico`).Error)
		require.NoError(t, tx.RollbackTo("antes_do_truncate").Error)
	})

	t.Run("nenhum dado pessoal no histórico", func(t *testing.T) {
		var linhas []string
		require.NoError(t, tx.Raw(`SELECT row_to_json(h)::text FROM emp_candidatura_historico h WHERE id_vaga = ?`, vaga).Scan(&linhas).Error)
		require.NotEmpty(t, linhas)
		for _, linha := range linhas {
			assert.NotContains(t, linha, cpfNovo)
			assert.NotContains(t, linha, cpfAntigo)
		}
	})

	t.Run("consultas de análise: funil, tempo médio e reprovados", func(t *testing.T) {
		var funil []struct {
			IDEtapa      uuid.UUID
			Candidaturas int
		}
		require.NoError(t, tx.Raw(sqlFunilPorEtapa, vaga).Scan(&funil).Error)
		porEtapa := map[uuid.UUID]int{}
		for _, linha := range funil {
			porEtapa[linha.IDEtapa] = linha.Candidaturas
		}
		assert.Equal(t, 1, porEtapa[triagem], "só a nova passou pela triagem")
		assert.Equal(t, 1, porEtapa[entrevista], "a antiga já estava na entrevista antes do histórico")

		var tempos []struct {
			IDEtapa        uuid.UUID
			SegundosMedios float64
			Passagens      int
		}
		require.NoError(t, tx.Raw(sqlTempoMedioPorEtapa, vaga).Scan(&tempos).Error)
		passagens := map[uuid.UUID]int{}
		for _, linha := range tempos {
			passagens[linha.IDEtapa] = linha.Passagens
			assert.GreaterOrEqual(t, linha.SegundosMedios, 0.0)
		}
		assert.Equal(t, 2, passagens[triagem], "triagem: ida e volta, encerradas pela entrevista e pela reprovação")
		assert.Equal(t, 2, passagens[entrevista], "entrevista: a passagem até a volta e a do lote, encerrada pela exclusão da etapa")

		var reprovados []struct {
			IDEtapa     *uuid.UUID
			Ator        string
			Motivo      string
			Reprovacoes int
		}
		require.NoError(t, tx.Raw(sqlReprovadosPorEtapaMotivo, vaga).Scan(&reprovados).Error)
		contagem := map[string]int{}
		for _, linha := range reprovados {
			contagem[linha.Ator+"/"+linha.Motivo] += linha.Reprovacoes
		}
		assert.Equal(t, 1, contagem["automacao/criterio_teste"])
		assert.Equal(t, 2, contagem["operador/sem_motivo"])
	})
}
