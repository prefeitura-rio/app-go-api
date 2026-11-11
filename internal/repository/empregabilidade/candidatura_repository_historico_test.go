package empregabilidade

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// Testes das transições de candidatura (DEV-532): cada mudança grava, na mesma
// transação, exatamente um evento em emp_candidatura_historico, só com ids.

const (
	sqlTravaCandidaturas = `SELECT "id","id_vaga","cpf","status","status_anterior","id_etapa_atual" FROM "emp_candidaturas" WHERE .+ FOR UPDATE`
	sqlOrdemEtapas       = `SELECT "id","ordem" FROM "emp_etapas" WHERE id IN`
	sqlInsertHistorico   = `INSERT INTO "emp_candidatura_historico"`
)

var colunasEstadoCandidatura = []string{"id", "id_vaga", "cpf", "status", "status_anterior", "id_etapa_atual"}

// cpfSintetico é um CPF fictício: o teste garante que ele nunca chega ao histórico.
const cpfSintetico = "12345678900"

type estadoTeste struct {
	id             uuid.UUID
	vaga           uuid.UUID
	cpf            string
	status         empregabilidade.StatusCandidatura
	statusAnterior interface{}
	etapa          interface{}
}

func linhasEstado(estados ...estadoTeste) *sqlmock.Rows {
	rows := sqlmock.NewRows(colunasEstadoCandidatura)
	for _, e := range estados {
		rows.AddRow(e.id, e.vaga, e.cpf, string(e.status), e.statusAnterior, e.etapa)
	}
	return rows
}

// eventoEsperado são os valores de um evento na ordem das colunas do INSERT
// (o created_at fica de fora e casa com qualquer valor).
type eventoEsperado struct {
	candidatura    uuid.UUID
	vaga           uuid.UUID
	tipo           empregabilidade.TipoEventoCandidatura
	ator           empregabilidade.AtorEventoCandidatura
	statusAnterior interface{}
	statusNovo     empregabilidade.StatusCandidatura
	etapaAnterior  interface{}
	etapaNova      interface{}
	motivo         interface{}
}

func argsEventos(eventos ...eventoEsperado) []driver.Value {
	var args []driver.Value
	for _, e := range eventos {
		args = append(args,
			e.candidatura.String(), e.vaga.String(), string(e.tipo), string(e.ator),
			e.statusAnterior, string(e.statusNovo), e.etapaAnterior, e.etapaNova, e.motivo,
			sqlmock.AnyArg(),
		)
	}
	return args
}

func esperarEventos(mock sqlmock.Sqlmock, eventos ...eventoEsperado) {
	rows := sqlmock.NewRows([]string{"id"})
	for i := range eventos {
		rows.AddRow(int64(i + 1))
	}
	mock.ExpectQuery(regexp.QuoteMeta(sqlInsertHistorico)).
		WithArgs(argsEventos(eventos...)...).
		WillReturnRows(rows)
}

func ctxComOrigem(ator empregabilidade.AtorEventoCandidatura, motivo string) context.Context {
	return empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: ator, Motivo: motivo})
}

func TestCandidaturaRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	t.Run("candidatura nova nasce com o evento candidatura_criada do candidato", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{
			ID:     uuid.New(),
			CPF:    cpfSintetico,
			Nome:   stringPtr("Pessoa Sintética"),
			Email:  stringPtr("pessoa@example.com"),
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(candidatura.ID))
		esperarEventos(mock, eventoEsperado{
			candidatura: candidatura.ID, vaga: vagaID,
			tipo: empregabilidade.EventoCandidaturaCriada, ator: empregabilidade.AtorCandidato,
			statusAnterior: nil, statusNovo: empregabilidade.StatusCandidaturaEnviada,
		})
		mock.ExpectCommit()

		id, err := repo.Create(ctxComOrigem(empregabilidade.AtorCandidato, ""), candidatura)
		require.NoError(t, err)
		assert.Equal(t, candidatura.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sem origem no contexto o evento é do operador", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{ID: uuid.New(), CPF: cpfSintetico, IDVaga: vagaID, Status: empregabilidade.StatusCandidaturaEnviada}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(candidatura.ID))
		esperarEventos(mock, eventoEsperado{
			candidatura: candidatura.ID, vaga: vagaID,
			tipo: empregabilidade.EventoCandidaturaCriada, ator: empregabilidade.AtorOperador,
			statusNovo: empregabilidade.StatusCandidaturaEnviada,
		})
		mock.ExpectCommit()

		_, err := repo.Create(context.Background(), candidatura)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		candidatura := &empregabilidade.Candidatura{ID: uuid.New(), CPF: cpfSintetico, IDVaga: uuid.New()}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(context.Background(), candidatura)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar candidatura")
		assert.Equal(t, uuid.Nil, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("falha ao gravar o evento desfaz a candidatura", func(t *testing.T) {
		candidatura := &empregabilidade.Candidatura{ID: uuid.New(), CPF: cpfSintetico, IDVaga: uuid.New(), Status: empregabilidade.StatusCandidaturaEnviada}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(candidatura.ID))
		mock.ExpectQuery(regexp.QuoteMeta(sqlInsertHistorico)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(context.Background(), candidatura)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao registrar histórico da candidatura")
		assert.Equal(t, uuid.Nil, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_UpdateStatus(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	etapa := uuid.New()

	tests := []struct {
		name   string
		ctx    context.Context
		atual  empregabilidade.StatusCandidatura
		novo   empregabilidade.StatusCandidatura
		tipo   empregabilidade.TipoEventoCandidatura
		ator   empregabilidade.AtorEventoCandidatura
		motivo interface{}
	}{
		{"aprovação pelo operador", context.Background(), empregabilidade.StatusCandidaturaEnviada, empregabilidade.StatusCandidaturaAprovada, empregabilidade.EventoCandidaturaAprovada, empregabilidade.AtorOperador, nil},
		{"reprovação manual", ctxComOrigem(empregabilidade.AtorOperador, ""), empregabilidade.StatusCandidaturaEnviada, empregabilidade.StatusCandidaturaReprovada, empregabilidade.EventoCandidaturaReprovada, empregabilidade.AtorOperador, nil},
		{"reprovação automática com motivo", ctxComOrigem(empregabilidade.AtorAutomacao, "criterio_teste"), empregabilidade.StatusCandidaturaEnviada, empregabilidade.StatusCandidaturaReprovada, empregabilidade.EventoCandidaturaReprovada, empregabilidade.AtorAutomacao, "criterio_teste"},
		{"reabertura de reprovada", context.Background(), empregabilidade.StatusCandidaturaReprovada, empregabilidade.StatusCandidaturaEnviada, empregabilidade.EventoStatusAlterado, empregabilidade.AtorOperador, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, vaga := uuid.New(), uuid.New()

			mock.ExpectBegin()
			mock.ExpectQuery(sqlTravaCandidaturas).
				WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: vaga, cpf: cpfSintetico, status: tt.atual, etapa: etapa}))
			mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas" SET "status"=`)).
				WillReturnResult(sqlmock.NewResult(0, 1))
			esperarEventos(mock, eventoEsperado{
				candidatura: id, vaga: vaga, tipo: tt.tipo, ator: tt.ator,
				statusAnterior: string(tt.atual), statusNovo: tt.novo,
				etapaAnterior: etapa.String(), etapaNova: etapa.String(), motivo: tt.motivo,
			})
			mock.ExpectCommit()

			require.NoError(t, repo.UpdateStatus(tt.ctx, id, tt.novo))
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("status igual ao atual não gera evento", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaAprovada}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, repo.UpdateStatus(context.Background(), id, empregabilidade.StatusCandidaturaAprovada))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("candidatura inexistente não gera evento", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		require.NoError(t, repo.UpdateStatus(context.Background(), uuid.New(), empregabilidade.StatusCandidaturaAprovada))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update status error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: uuid.New(), vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateStatus(context.Background(), uuid.New(), empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar status da candidatura")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("falha ao gravar o evento desfaz a mudança de status", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(sqlInsertHistorico)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateStatus(context.Background(), id, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao registrar histórico da candidatura")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("falha ao travar a candidatura", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateStatus(context.Background(), uuid.New(), empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar candidaturas")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_UpdateEtapa(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	etapa1, etapa2 := uuid.New(), uuid.New()

	tests := []struct {
		name   string
		atual  interface{}
		nova   uuid.UUID
		ordens [][2]interface{}
		tipo   empregabilidade.TipoEventoCandidatura
	}{
		{"primeira etapa conta como mudança", nil, etapa1, [][2]interface{}{{etapa1, 1}}, empregabilidade.EventoMudancaEtapa},
		{"avanço para etapa de ordem maior", etapa1, etapa2, [][2]interface{}{{etapa1, 1}, {etapa2, 2}}, empregabilidade.EventoMudancaEtapa},
		{"volta para etapa de ordem menor", etapa2, etapa1, [][2]interface{}{{etapa1, 1}, {etapa2, 2}}, empregabilidade.EventoVoltaEtapa},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, vaga := uuid.New(), uuid.New()
			var etapaAnterior interface{}
			if tt.atual != nil {
				etapaAnterior = tt.atual.(uuid.UUID).String()
			}

			mock.ExpectBegin()
			mock.ExpectQuery(sqlTravaCandidaturas).
				WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: vaga, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: tt.atual}))
			mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas" SET "id_etapa_atual"=`)).
				WillReturnResult(sqlmock.NewResult(0, 1))
			ordens := sqlmock.NewRows([]string{"id", "ordem"})
			for _, o := range tt.ordens {
				ordens.AddRow(o[0], o[1])
			}
			mock.ExpectQuery(regexp.QuoteMeta(sqlOrdemEtapas)).WillReturnRows(ordens)
			esperarEventos(mock, eventoEsperado{
				candidatura: id, vaga: vaga, tipo: tt.tipo, ator: empregabilidade.AtorOperador,
				statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaEnviada,
				etapaAnterior: etapaAnterior, etapaNova: tt.nova.String(),
			})
			mock.ExpectCommit()

			require.NoError(t, repo.UpdateEtapa(context.Background(), id, tt.nova))
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("mesma etapa não gera evento", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa1}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, repo.UpdateEtapa(context.Background(), id, etapa1))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update etapa error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateEtapa(context.Background(), uuid.New(), etapa1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar etapa da candidatura")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("falha ao ler a ordem das etapas desfaz a mudança", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: uuid.New(), vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa1}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(sqlOrdemEtapas)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateEtapa(context.Background(), uuid.New(), etapa2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao ler ordem das etapas")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_BulkUpdateEtapa(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	t.Run("um evento por candidatura que mudou de etapa", func(t *testing.T) {
		etapa1, etapa2 := uuid.New(), uuid.New()
		vaga := uuid.New()
		id1, id2 := uuid.New(), uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(
				estadoTeste{id: id1, vaga: vaga, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa1},
				estadoTeste{id: id2, vaga: vaga, cpf: "98765432100", status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa2},
			))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas" SET "id_etapa_atual"=`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectQuery(regexp.QuoteMeta(sqlOrdemEtapas)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "ordem"}).AddRow(etapa1, 1).AddRow(etapa2, 2))
		esperarEventos(mock, eventoEsperado{
			candidatura: id1, vaga: vaga, tipo: empregabilidade.EventoMudancaEtapa, ator: empregabilidade.AtorOperador,
			statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaEnviada,
			etapaAnterior: etapa1.String(), etapaNova: etapa2.String(),
		})
		mock.ExpectCommit()

		require.NoError(t, repo.BulkUpdateEtapa(context.Background(), []uuid.UUID{id1, id2}, etapa2))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update etapa error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkUpdateEtapa(context.Background(), []uuid.UUID{uuid.New()}, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar etapa em lote")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_BulkUpdateStatus(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	t.Run("bulk update all cpfs found, evento só para quem mudou", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1, cpf2 := cpfSintetico, "98765432100"
		id1, id2 := uuid.New(), uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(
				estadoTeste{id: id1, vaga: vagaID, cpf: cpf1, status: empregabilidade.StatusCandidaturaEnviada},
				estadoTeste{id: id2, vaga: vagaID, cpf: cpf2, status: empregabilidade.StatusCandidaturaAprovada},
			))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		esperarEventos(mock, eventoEsperado{
			candidatura: id1, vaga: vagaID, tipo: empregabilidade.EventoCandidaturaAprovada, ator: empregabilidade.AtorOperador,
			statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaAprovada,
		})
		mock.ExpectCommit()

		result, err := repo.BulkUpdateStatus(context.Background(), vagaID, []string{cpf1, cpf2}, empregabilidade.StatusCandidaturaAprovada)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Updated)
		assert.Empty(t, result.FailedCPFs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update some cpfs not found", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1, cpf2, cpf3 := cpfSintetico, "98765432100", "11111111111"
		id1, id2 := uuid.New(), uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(
				estadoTeste{id: id1, vaga: vagaID, cpf: cpf1, status: empregabilidade.StatusCandidaturaEnviada},
				estadoTeste{id: id2, vaga: vagaID, cpf: cpf2, status: empregabilidade.StatusCandidaturaEnviada},
			))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		esperarEventos(mock,
			eventoEsperado{candidatura: id1, vaga: vagaID, tipo: empregabilidade.EventoCandidaturaReprovada, ator: empregabilidade.AtorOperador, statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaReprovada},
			eventoEsperado{candidatura: id2, vaga: vagaID, tipo: empregabilidade.EventoCandidaturaReprovada, ator: empregabilidade.AtorOperador, statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaReprovada},
		)
		mock.ExpectCommit()

		result, err := repo.BulkUpdateStatus(context.Background(), vagaID, []string{cpf1, cpf2, cpf3}, empregabilidade.StatusCandidaturaReprovada)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Updated)
		assert.Equal(t, []string{cpf3}, result.FailedCPFs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update no cpfs found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectCommit()

		result, err := repo.BulkUpdateStatus(context.Background(), uuid.New(), []string{cpfSintetico, "98765432100"}, empregabilidade.StatusCandidaturaAprovada)
		require.NoError(t, err)
		assert.Equal(t, 0, result.Updated)
		assert.Len(t, result.FailedCPFs, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update query error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		result, err := repo.BulkUpdateStatus(context.Background(), uuid.New(), []string{cpfSintetico}, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar candidaturas")
		assert.Equal(t, 0, result.Updated)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update execution error", func(t *testing.T) {
		vagaID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: uuid.New(), vaga: vagaID, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		_, err := repo.BulkUpdateStatus(context.Background(), vagaID, []string{cpfSintetico}, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar status em lote")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_BulkSaveAndUpdateStatusByVagaID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	sqlCongelar := `UPDATE emp_candidaturas SET status_anterior = status, status = .*, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL`

	t.Run("freeze candidaturas success: um evento da automação por candidatura", func(t *testing.T) {
		vagaID := uuid.New()
		id1, id2 := uuid.New(), uuid.New()
		etapa := uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(
				estadoTeste{id: id1, vaga: vagaID, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa},
				estadoTeste{id: id2, vaga: vagaID, cpf: "98765432100", status: empregabilidade.StatusCandidaturaAprovada},
			))
		mock.ExpectExec(sqlCongelar).
			WithArgs(string(empregabilidade.StatusCandidaturaVagaCongelada), vagaID).
			WillReturnResult(sqlmock.NewResult(0, 2))
		esperarEventos(mock,
			eventoEsperado{candidatura: id1, vaga: vagaID, tipo: empregabilidade.EventoStatusAlterado, ator: empregabilidade.AtorAutomacao, statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaVagaCongelada, etapaAnterior: etapa.String(), etapaNova: etapa.String(), motivo: empregabilidade.MotivoVagaCongelada},
			eventoEsperado{candidatura: id2, vaga: vagaID, tipo: empregabilidade.EventoStatusAlterado, ator: empregabilidade.AtorAutomacao, statusAnterior: string(empregabilidade.StatusCandidaturaAprovada), statusNovo: empregabilidade.StatusCandidaturaVagaCongelada, motivo: empregabilidade.MotivoVagaCongelada},
		)
		mock.ExpectCommit()

		ctx := ctxComOrigem(empregabilidade.AtorAutomacao, empregabilidade.MotivoVagaCongelada)
		require.NoError(t, repo.BulkSaveAndUpdateStatusByVagaID(ctx, vagaID, empregabilidade.StatusCandidaturaVagaCongelada))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("vaga sem candidaturas não grava evento", func(t *testing.T) {
		vagaID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(sqlCongelar).
			WithArgs(string(empregabilidade.StatusCandidaturaDescontinuada), vagaID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		require.NoError(t, repo.BulkSaveAndUpdateStatusByVagaID(context.Background(), vagaID, empregabilidade.StatusCandidaturaDescontinuada))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("freeze candidaturas error", func(t *testing.T) {
		vagaID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(sqlCongelar).
			WithArgs(string(empregabilidade.StatusCandidaturaVagaCongelada), vagaID).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkSaveAndUpdateStatusByVagaID(context.Background(), vagaID, empregabilidade.StatusCandidaturaVagaCongelada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao congelar/descontinuar candidaturas")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_BulkRestoreStatusByVagaID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	sqlRestaurar := `UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL AND status_anterior IS NOT NULL`

	t.Run("restore candidaturas success: evento volta ao status guardado", func(t *testing.T) {
		vagaID := uuid.New()
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: vagaID, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaVagaCongelada, statusAnterior: string(empregabilidade.StatusCandidaturaReprovada)}))
		mock.ExpectExec(sqlRestaurar).WithArgs(vagaID).WillReturnResult(sqlmock.NewResult(0, 1))
		esperarEventos(mock, eventoEsperado{
			candidatura: id, vaga: vagaID, tipo: empregabilidade.EventoCandidaturaReprovada, ator: empregabilidade.AtorAutomacao,
			statusAnterior: string(empregabilidade.StatusCandidaturaVagaCongelada), statusNovo: empregabilidade.StatusCandidaturaReprovada,
			motivo: empregabilidade.MotivoVagaDescongelada,
		})
		mock.ExpectCommit()

		ctx := ctxComOrigem(empregabilidade.AtorAutomacao, empregabilidade.MotivoVagaDescongelada)
		require.NoError(t, repo.BulkRestoreStatusByVagaID(ctx, vagaID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("restore candidaturas error", func(t *testing.T) {
		vagaID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(sqlRestaurar).WithArgs(vagaID).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkRestoreStatusByVagaID(context.Background(), vagaID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao restaurar status das candidaturas")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("restore candidaturas no rows affected", func(t *testing.T) {
		vagaID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectExec(sqlRestaurar).WithArgs(vagaID).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		require.NoError(t, repo.BulkRestoreStatusByVagaID(context.Background(), vagaID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCandidaturaRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	t.Run("cancelamento pelo candidato grava candidatura_cancelada", func(t *testing.T) {
		id, vaga, etapa := uuid.New(), uuid.New(), uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: vaga, cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada, etapa: etapa}))
		// GORM soft delete uses UPDATE not DELETE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas" SET "deleted_at"=`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		esperarEventos(mock, eventoEsperado{
			candidatura: id, vaga: vaga, tipo: empregabilidade.EventoCandidaturaCancelada, ator: empregabilidade.AtorCandidato,
			statusAnterior: string(empregabilidade.StatusCandidaturaEnviada), statusNovo: empregabilidade.StatusCandidaturaEnviada,
			etapaAnterior: etapa.String(), etapaNova: etapa.String(),
		})
		mock.ExpectCommit()

		require.NoError(t, repo.Delete(ctxComOrigem(empregabilidade.AtorCandidato, ""), id))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).WillReturnRows(sqlmock.NewRows(colunasEstadoCandidatura))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "não encontrada")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete database error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery(sqlTravaCandidaturas).
			WillReturnRows(linhasEstado(estadoTeste{id: id, vaga: uuid.New(), cpf: cpfSintetico, status: empregabilidade.StatusCandidaturaEnviada}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir candidatura")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
