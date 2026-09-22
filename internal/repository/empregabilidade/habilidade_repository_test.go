package empregabilidade_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	repoEmpregabilidade "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// --- TESTES DE CONSTRUÇÃO DE SQL E FILTROS ---

func TestHabilidadeRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	var areaID int64 = 5

	tests := []struct {
		name               string
		filter             empregabilidade.HabilidadeFilter
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             empregabilidade.HabilidadeFilter{},
			expectedConditions: []string{},
			description:        "Filtro vazio não deve gerar cláusulas WHERE",
		},
		{
			name: "filter by search with unaccent",
			filter: empregabilidade.HabilidadeFilter{
				Search: "Acabamento",
			},
			expectedConditions: []string{
				"lower(immutable_unaccent(emp_habilidades.nome)) LIKE lower(immutable_unaccent(",
				"%Acabamento%",
			},
			description: "Busca textual deve aplicar a função unaccent",
		},
		{
			name: "filter by area_atuacao_id",
			filter: empregabilidade.HabilidadeFilter{
				AreaAtuacaoID: areaID,
			},
			expectedConditions: []string{
				"JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id",
				"aah.id_area_atuacao =",
			},
			description: "Filtro por área deve realizar JOIN com a tabela associativa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.Search != "" {
					searchNome := "%" + tt.filter.Search + "%"
					db = db.Where("lower(immutable_unaccent(emp_habilidades.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
				}

				if tt.filter.AreaAtuacaoID > 0 {
					db = db.Joins("JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id").
						Where("aah.id_area_atuacao = ?", tt.filter.AreaAtuacaoID)
				}

				return db
			}

			query := db.Model(&empregabilidade.Habilidade{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Habilidade{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition, "%s: O SQL deve conter '%s'", tt.description, condition)
			}
		})
	}
}

// --- TESTES DOS MÉTODOS DE CRUD BÁSICOS ---

func TestHabilidadeRepository_Create_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	entity := &empregabilidade.Habilidade{
		ID:   habilidadeID,
		Nome: "Desenvolvimento Go",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_habilidades"`).
		WithArgs(entity.Nome, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(habilidadeID))
	mock.ExpectCommit()

	id, err := repo.CreateHabilidade(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, habilidadeID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_Create_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Habilidade{Nome: "Desenvolvimento Go"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_habilidades"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.CreateHabilidade(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar habilidade")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetByID_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var habilidadeID int64 = 10

	mock.ExpectQuery(`SELECT \* FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).AddRow(habilidadeID, "Desenvolvimento Go"))

	// Preload das Áreas vinculadas
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, err := repo.GetHabilidadeByID(ctx, habilidadeID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, habilidadeID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var habilidadeID int64 = 999

	mock.ExpectQuery(`SELECT \* FROM "emp_habilidades"`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetHabilidadeByID(ctx, habilidadeID)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var habilidadeID int64 = 10

	mock.ExpectQuery(`SELECT \* FROM "emp_habilidades"`).
		WillReturnError(assert.AnError)

	result, err := repo.GetHabilidadeByID(ctx, habilidadeID)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "erro ao buscar habilidade por ID")
	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_Update_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Habilidade{
		ID:        15,
		Nome:      "Go Avançado",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_habilidades"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateHabilidade(ctx, entity)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_Update_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Habilidade{
		ID:        15,
		Nome:      "Go Avançado",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_habilidades"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateHabilidade(ctx, entity)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "erro ao atualizar habilidade")
	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_Delete_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var habilidadeID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_habilidades"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteHabilidade(ctx, habilidadeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var habilidadeID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_habilidades"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.DeleteHabilidade(ctx, habilidadeID)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "erro ao excluir habilidade")
	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- TESTES DE CRUD PARA ÁREA DE ATUAÇÃO ---

func TestHabilidadeRepository_CreateAreaAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var areaID int64 = 5
	entity := &empregabilidade.AreaAtuacao{
		ID:   areaID,
		Nome: "Tecnologia da Informação",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_areas_atuacao"`).
		WithArgs(entity.Nome, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(areaID))
	mock.ExpectCommit()

	id, err := repo.CreateAreaAtuacao(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, areaID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_CreateAreaAtuacao_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.AreaAtuacao{Nome: "Tecnologia da Informação"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_areas_atuacao"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.CreateAreaAtuacao(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar Área de Atuação")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetAreaAtuacaoByID_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var areaID int64 = 10

	// Query principal da Área de Atuação
	mock.ExpectQuery(`SELECT \* FROM "emp_areas_atuacao"`).
		WithArgs(areaID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).AddRow(areaID, "Construção Civil"))

	// Preload das Habilidades vinculadas via tabela associativa
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WithArgs(areaID).
		WillReturnRows(sqlmock.NewRows([]string{"id_area_atuacao", "id_habilidade"}))

	result, err := repo.GetAreaAtuacaoByID(ctx, areaID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, areaID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetAreaAtuacaoByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var areaID int64 = 999

	mock.ExpectQuery(`SELECT \* FROM "emp_areas_atuacao"`).
		WithArgs(areaID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetAreaAtuacaoByID(ctx, areaID)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_UpdateAreaAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.AreaAtuacao{
		ID:        15,
		Nome:      "Construção Civil Atualizada",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_areas_atuacao"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateAreaAtuacao(ctx, entity)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_DeleteAreaAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var areaID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_areas_atuacao"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteAreaAtuacao(ctx, areaID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- TESTES DOS MÉTODOS DE LISTAGEM ---

func TestHabilidadeRepository_List_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.HabilidadeFilter{Search: "Acabamento"}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Acabamento em gesso"))

	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, total, err := repo.ListHabilidades(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestHabilidadeRepository_ListAreasAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.AreaAtuacaoFilter{Search: "Confecção"}

	// 1. Contagem total
	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_areas_atuacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// 2. Busca das áreas de atuação
	mock.ExpectQuery(`SELECT .* FROM "emp_areas_atuacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Confecção de Calçados").
			AddRow(2, "Confecção de Roupas"))

	// 3. Preload das Habilidades/Relações da Área de Atuação (ADICIONADO)
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, total, err := repo.ListAreasAtuacao(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- TESTES DE VÍNCULO E DESVÍNCULO de Áreas de Atuação (MANY-TO-MANY) ---

func TestHabilidadeRepository_AttachAreaAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	var areaID int64 = 5

	mock.ExpectBegin()

	// 1. O GORM atualiza o updated_at da habilidade pai
	mock.ExpectExec(`UPDATE "emp_habilidades"`).
		WithArgs(sqlmock.AnyArg(), habilidadeID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. Insere a associação na tabela pivô usando ExpectExec em vez de ExpectQuery
	mock.ExpectExec(`INSERT INTO "area_atuacao_habilidade"`).
		WithArgs(habilidadeID, areaID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.AttachAreaAtuacao(ctx, habilidadeID, areaID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_DetachAreaAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	var areaID int64 = 5

	// O GORM deleta a relação específica na tabela pivô sem apagar as entidades principais
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "area_atuacao_habilidade"`).
		WithArgs(habilidadeID, areaID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DetachAreaAtuacao(ctx, habilidadeID, areaID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ReplaceAreasAtuacao_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	areaIDs := []int64{10, 20}

	// 1. O GORM busca as áreas existentes antes de abrir a transação
	mock.ExpectQuery(`SELECT \* FROM "emp_areas_atuacao" WHERE id IN \(\$1,\$2\)`).
		WithArgs(areaIDs[0], areaIDs[1]).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(areaIDs[0], "Área 10").
			AddRow(areaIDs[1], "Área 20"))

	// 2. Transação 1: UPDATE do pai + INSERT dos novos vínculos na pivô
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_habilidades"`).
		WithArgs(sqlmock.AnyArg(), habilidadeID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// ALTERADO: Substituído ExpectQuery por ExpectExec
	mock.ExpectExec(`INSERT INTO "area_atuacao_habilidade"`).
		WithArgs(habilidadeID, areaIDs[0], habilidadeID, areaIDs[1]).
		WillReturnResult(sqlmock.NewResult(2, 2))

	mock.ExpectCommit()

	// 3. Transação 2: Limpeza dos vínculos antigos que não pertencem mais ao slice
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "area_atuacao_habilidade"`).
		WithArgs(habilidadeID, areaIDs[0], areaIDs[1]).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.ReplaceAreasAtuacao(ctx, habilidadeID, areaIDs)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- COBERTURA COMPLEMENTAR DE ListHabilidades ---

func TestHabilidadeRepository_ListHabilidades_AreaAtuacaoFilter_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.HabilidadeFilter{AreaAtuacaoID: 5}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_habilidades" JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id WHERE aah.id_area_atuacao = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(
		`SELECT .* FROM "emp_habilidades" JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id WHERE aah.id_area_atuacao = \$1`,
	).
		WithArgs(int64(5), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nome",
				"created_at",
				"updated_at",
			}).
				AddRow(
					1,
					"Pintura Predial",
					time.Now(),
					time.Now(),
				),
		)

	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, total, err := repo.ListHabilidades(ctx, filter, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListHabilidades_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.HabilidadeFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_habilidades"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListHabilidades(ctx, filter, 10, 0)

	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "erro ao contar habilidades")
	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListHabilidades_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.HabilidadeFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_habilidades"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListHabilidades(ctx, filter, 10, 0)

	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "erro ao listar habilidades")
	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListAreaAtuacaoHabilidades_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	now := time.Now()

	// 1. Consulta principal
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "area_atuacao_habilidade" ORDER BY id_area_atuacao ASC, id_habilidade ASC`,
		),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"id_habilidade",
				"id_area_atuacao",
				"created_at",
				"updated_at",
			}).
				AddRow(1, 10, 20, now, now).
				AddRow(2, 11, 20, now, now),
		)

	// 2. Preload("AreaAtuacao")
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "emp_areas_atuacao" WHERE "emp_areas_atuacao"."id" = $1`,
		),
	).
		WithArgs(int64(20)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nome",
				"created_at",
				"updated_at",
			}).
				AddRow(20, "Tecnologia da Informação", now, now),
		)

	// 3. Preload("Habilidade")
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "emp_habilidades" WHERE "emp_habilidades"."id" IN ($1,$2)`,
		),
	).
		WithArgs(int64(10), int64(11)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nome",
				"created_at",
				"updated_at",
			}).
				AddRow(10, "Go", now, now).
				AddRow(11, "Java", now, now),
		)

	result, err := repo.ListAreaAtuacaoHabilidades(ctx)

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(10), result[0].IDHabilidade)
	assert.Equal(t, int64(20), result[0].IDAreaAtuacao)

	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, int64(11), result[1].IDHabilidade)
	assert.Equal(t, int64(20), result[1].IDAreaAtuacao)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListAreaAtuacaoHabilidades_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(
		`SELECT \* FROM "area_atuacao_habilidade" ORDER BY id_area_atuacao ASC, id_habilidade ASC`,
	).
		WillReturnError(assert.AnError)

	result, err := repo.ListAreaAtuacaoHabilidades(ctx)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorContains(
		t,
		err,
		"erro ao listar áreas de atuação e habilidades",
	)
	assert.ErrorIs(t, err, assert.AnError)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListAreaAtuacaoHabilidades_AreaAtuacaoPreloadError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Consulta principal
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "area_atuacao_habilidade" ORDER BY id_area_atuacao ASC, id_habilidade ASC`,
		),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"id_habilidade",
				"id_area_atuacao",
				"created_at",
				"updated_at",
			}).
				AddRow(1, 10, 20, now, now),
		)

	// Preload("AreaAtuacao") falha
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "emp_areas_atuacao" WHERE "emp_areas_atuacao"."id" = $1`,
		),
	).
		WithArgs(int64(20)).
		WillReturnError(assert.AnError)

	result, err := repo.ListAreaAtuacaoHabilidades(ctx)

	assert.Nil(t, result)
	require.Error(t, err)

	assert.ErrorContains(
		t,
		err,
		"erro ao listar áreas de atuação e habilidades",
	)
	assert.ErrorIs(t, err, assert.AnError)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListAreaAtuacaoHabilidades_HabilidadePreloadError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Consulta principal
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "area_atuacao_habilidade" ORDER BY id_area_atuacao ASC, id_habilidade ASC`,
		),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"id_habilidade",
				"id_area_atuacao",
				"created_at",
				"updated_at",
			}).
				AddRow(1, 10, 20, now, now),
		)

	// Preload("AreaAtuacao") passa
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "emp_areas_atuacao" WHERE "emp_areas_atuacao"."id" = $1`,
		),
	).
		WithArgs(int64(20)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nome",
				"created_at",
				"updated_at",
			}).
				AddRow(20, "Tecnologia da Informação", now, now),
		)

	// Preload("Habilidade") falha
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "emp_habilidades" WHERE "emp_habilidades"."id" = $1`,
		),
	).
		WithArgs(int64(10)).
		WillReturnError(assert.AnError)

	result, err := repo.ListAreaAtuacaoHabilidades(ctx)

	assert.Nil(t, result)
	require.Error(t, err)

	assert.ErrorContains(
		t,
		err,
		"erro ao listar áreas de atuação e habilidades",
	)
	assert.ErrorIs(t, err, assert.AnError)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_AttachAreaAtuacao_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	var areaID int64 = 5

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE "emp_habilidades"`).
		WithArgs(sqlmock.AnyArg(), habilidadeID).
		WillReturnError(errors.New("erro ao atualizar habilidade"))

	mock.ExpectRollback()

	err := repo.AttachAreaAtuacao(ctx, habilidadeID, areaID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_DetachAreaAtuacao_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var habilidadeID int64 = 1
	var areaID int64 = 5

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "area_atuacao_habilidade"`).
		WithArgs(habilidadeID, areaID).
		WillReturnError(errors.New("erro ao desvincular área de atuação"))
	mock.ExpectRollback()

	err := repo.DetachAreaAtuacao(ctx, habilidadeID, areaID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_GetAreaAtuacaoByID_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var areaID int64 = 10

	mock.ExpectQuery(`SELECT \* FROM "emp_areas_atuacao"`).
		WithArgs(areaID, 1).
		WillReturnError(errors.New("erro ao buscar área de atuação"))

	result, err := repo.GetAreaAtuacaoByID(ctx, areaID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar área de atuação por ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_UpdateAreaAtuacao_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.AreaAtuacao{
		ID:        15,
		Nome:      "Construção Civil Atualizada",
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_areas_atuacao"`).
		WillReturnError(errors.New("erro ao atualizar área de atuação"))
	mock.ExpectRollback()

	err := repo.UpdateAreaAtuacao(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar área de atuação")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_DeleteAreaAtuacao_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	var areaID int64 = 20

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_areas_atuacao"`).
		WillReturnError(errors.New("erro ao excluir área de atuação"))
	mock.ExpectRollback()

	err := repo.DeleteAreaAtuacao(ctx, areaID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir uma área de atuação")
	assert.NoError(t, mock.ExpectationsWereMet())
}
