package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// DEV-518: a listagem de bloqueios de candidatura de um editor de
// empregabilidade fica restrita às vagas dos órgãos parceiros do escopo dele,
// como a listagem de vagas. nil não restringe (admin); lista vazia não devolve
// nada.

const bloqueioSubconsultaOrgao = `id_vaga IN (SELECT id FROM emp_vagas WHERE id_orgao_parceiro IN ($1,$2))`

func TestCandidaturaBloqueioRepository_List_OrgaoParceiroIDs_RestringeAoEscopo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios" WHERE `+bloqueioSubconsultaOrgao)).
		WithArgs("1", "2600").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios" WHERE `+bloqueioSubconsultaOrgao)).
		WithArgs("1", "2600", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	entities, total, err := repo.List(context.Background(), map[string]interface{}{}, []string{"1", "2600"}, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, entities)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_OrgaoParceiroIDs_ComFiltros(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	vagaID := uuid.New()
	filter := map[string]interface{}{"cpf": "12345678900", "id_vaga": vagaID}

	// O escopo soma-se aos filtros da query: um id_vaga de outro órgão não
	// escapa da restrição.
	where := `WHERE cpf = $1 AND id_vaga = $2 AND id_vaga IN (SELECT id FROM emp_vagas WHERE id_orgao_parceiro IN ($3))`
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios" `+where)).
		WithArgs("12345678900", vagaID, "1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios" `+where)).
		WithArgs("12345678900", vagaID, "1", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, _, err := repo.List(context.Background(), filter, []string{"1"}, 10, 0)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_OrgaoParceiroIDsVazio_NaoDevolveNada(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios" WHERE 1 = 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios" WHERE 1 = 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	entities, total, err := repo.List(context.Background(), map[string]interface{}{}, []string{}, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, entities)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_OrgaoParceiroIDsNil_NaoRestringe(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)

	mock.ExpectQuery(`^SELECT count\(\*\) FROM "emp_candidatura_bloqueios"$`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`^SELECT \* FROM "emp_candidatura_bloqueios" ORDER BY`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	entities, total, err := repo.List(context.Background(), map[string]interface{}{}, nil, 10, 0)

	require.NoError(t, err)
	assert.Len(t, entities, 1)
	assert.Equal(t, 1, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}
