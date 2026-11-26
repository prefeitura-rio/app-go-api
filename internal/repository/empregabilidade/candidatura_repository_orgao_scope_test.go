package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// DEV-517: a listagem de candidaturas de um editor de empregabilidade fica
// restrita às vagas dos órgãos parceiros do escopo dele (OrgaoParceiroIDs),
// como a listagem de vagas. nil não restringe (admin); lista vazia não devolve
// nada.

const subconsultaOrgaoParceiro = `id_vaga IN \(SELECT id FROM emp_vagas WHERE id_orgao_parceiro IN \(\$1,\$2\)\)`

func TestCandidaturaRepository_List_OrgaoParceiroIDs_RestringeAoEscopo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	filter := empregabilidade.CandidaturaFilter{OrgaoParceiroIDs: []string{"1", "2600"}}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_candidaturas" WHERE `+subconsultaOrgaoParceiro).
		WithArgs("1", "2600").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "emp_candidaturas" WHERE `+subconsultaOrgaoParceiro).
		WithArgs("1", "2600", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	entities, total, err := repo.List(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, entities)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaRepository_List_OrgaoParceiroIDsVazio_NaoDevolveNada(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	filter := empregabilidade.CandidaturaFilter{OrgaoParceiroIDs: []string{}}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_candidaturas" WHERE 1 = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "emp_candidaturas" WHERE 1 = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, total, err := repo.List(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaRepository_List_OrgaoParceiroIDsNil_NaoRestringe(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_candidaturas" WHERE "emp_candidaturas"."deleted_at" IS NULL$`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "emp_candidaturas" WHERE "emp_candidaturas"."deleted_at" IS NULL ORDER BY`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, _, err := repo.List(context.Background(), empregabilidade.CandidaturaFilter{}, 10, 0)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaRepository_CountByStatus_OrgaoParceiroIDs_RestringeAoEscopo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	filter := empregabilidade.CandidaturaFilter{OrgaoParceiroIDs: []string{"1", "2600"}}

	mock.ExpectQuery(`SELECT status, COUNT\(\*\) as count FROM "emp_candidaturas" WHERE `+subconsultaOrgaoParceiro).
		WithArgs("1", "2600").
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).AddRow("candidatura_enviada", 3))

	counts, err := repo.CountByStatus(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(3), counts[empregabilidade.StatusCandidaturaEnviada])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaRepository_CountByStatus_OrgaoParceiroIDsVazio_NaoContaNada(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)

	mock.ExpectQuery(`SELECT status, COUNT\(\*\) as count FROM "emp_candidaturas" WHERE 1 = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}))

	_, err := repo.CountByStatus(context.Background(), empregabilidade.CandidaturaFilter{OrgaoParceiroIDs: []string{}})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
