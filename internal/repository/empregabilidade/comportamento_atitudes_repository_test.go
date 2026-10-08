package empregabilidade_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	repoEmpregabilidade "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// --- TESTES DE CONSTRUÇÃO DE SQL E FILTROS ---

func TestComportamentoAtitudesRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		filter             empregabilidade.ComportamentoAtitudesFilter
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             empregabilidade.ComportamentoAtitudesFilter{},
			expectedConditions: []string{},
			description:        "Filtro vazio não deve gerar cláusulas WHERE",
		},
		{
			name: "filter by search with unaccent",
			filter: empregabilidade.ComportamentoAtitudesFilter{
				Search: "Proatividade",
			},
			expectedConditions: []string{
				"lower(immutable_unaccent(emp_comportamento_atitudes.nome)) LIKE lower(immutable_unaccent(",
				"%Proatividade%",
			},
			description: "Busca textual deve aplicar a função unaccent na tabela emp_comportamento_atitudes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.Search != "" {
					searchNome := "%" + tt.filter.Search + "%"
					db = db.Where("lower(immutable_unaccent(emp_comportamento_atitudes.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
				}
				return db
			}

			query := db.Model(&empregabilidade.ComportamentoAtitudes{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.ComportamentoAtitudes{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition, "%s: O SQL deve conter '%s'", tt.description, condition)
			}
		})
	}
}

// --- TESTES DOS MÉTODOS DE CRUD BÁSICOS DO CATÁLOGO MESTRE ---

func TestComportamentoAtitudesRepository_Create_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()

	var comportamentoID int64 = 1
	entity := &empregabilidade.ComportamentoAtitudes{
		ID:   comportamentoID,
		Nome: "Proatividade",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_comportamento_atitudes"`).
		WithArgs(entity.Nome, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(comportamentoID))
	mock.ExpectCommit()

	id, err := repo.CreateComportamentoAtitudes(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, comportamentoID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_Create_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ComportamentoAtitudes{Nome: "Proatividade"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_comportamento_atitudes"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.CreateComportamentoAtitudes(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar Comportamento e Atitudes")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_GetByID_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	var comportamentoID int64 = 10

	mock.ExpectQuery(`SELECT \* FROM "emp_comportamento_atitudes"`).
		WithArgs(comportamentoID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).AddRow(comportamentoID, "Liderança"))

	result, err := repo.GetComportamentoAtitudesByID(ctx, comportamentoID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, comportamentoID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	var comportamentoID int64 = 10

	mock.ExpectQuery(`SELECT \* FROM "emp_comportamento_atitudes"`).
		WithArgs(comportamentoID, 1).
		WillReturnError(assert.AnError)

	result, err := repo.GetComportamentoAtitudesByID(ctx, comportamentoID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao buscar comportamento e atitude por ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	var comportamentoID int64 = 999

	mock.ExpectQuery(`SELECT \* FROM "emp_comportamento_atitudes"`).
		WithArgs(comportamentoID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetComportamentoAtitudesByID(ctx, comportamentoID)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_Update_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ComportamentoAtitudes{
		ID:        15,
		Nome:      "Empatia Atualizada",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_comportamento_atitudes"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateComportamentoAtitudes(ctx, entity)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_Update_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ComportamentoAtitudes{
		ID:        15,
		Nome:      "Empatia Atualizada",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_comportamento_atitudes"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateComportamentoAtitudes(ctx, entity)
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao atualizar comportamento e atitude")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_Delete_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	var comportamentoID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_comportamento_atitudes"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteComportamentoAtitudes(ctx, comportamentoID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	var comportamentoID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_comportamento_atitudes"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.DeleteComportamentoAtitudes(ctx, comportamentoID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao excluir comportamento e atitude")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_List_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	filter := empregabilidade.ComportamentoAtitudesFilter{Search: "Proativid"}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_comportamento_atitudes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_comportamento_atitudes"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Proatividade"))

	result, total, err := repo.ListComportamentoAtitudes(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_List_WithoutSearch_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	filter := empregabilidade.ComportamentoAtitudesFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_comportamento_atitudes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM "emp_comportamento_atitudes"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empatia").
			AddRow(2, "Proatividade"))

	result, total, err := repo.ListComportamentoAtitudes(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "Empatia", result[0].Nome)
	assert.Equal(t, "Proatividade", result[1].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_List_CountDatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	filter := empregabilidade.ComportamentoAtitudesFilter{Search: "Proativid"}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_comportamento_atitudes"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListComportamentoAtitudes(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao contar comportamentos e atitudes")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComportamentoAtitudesRepository_List_FindDatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewComportamentoAtitudesRepository(db)
	ctx := context.Background()
	filter := empregabilidade.ComportamentoAtitudesFilter{Search: "Proativid"}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_comportamento_atitudes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_comportamento_atitudes"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListComportamentoAtitudes(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao listar comportamentos e atitudes")
	assert.NoError(t, mock.ExpectationsWereMet())
}
