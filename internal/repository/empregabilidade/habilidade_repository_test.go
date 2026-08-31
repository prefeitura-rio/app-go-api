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
	mock.ExpectQuery(`INSERT INTO "area_atuacao"`).
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
	mock.ExpectQuery(`INSERT INTO "area_atuacao"`).
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
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao"`).
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

	mock.ExpectQuery(`SELECT \* FROM "area_atuacao"`).
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
	mock.ExpectExec(`UPDATE "area_atuacao"`).
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
	mock.ExpectExec(`DELETE FROM "area_atuacao"`).
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
	mock.ExpectQuery(`SELECT count\(\*\) FROM "area_atuacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// 2. Busca das áreas de atuação
	mock.ExpectQuery(`SELECT .* FROM "area_atuacao"`).
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

// --- TESTES DE VÍNCULO COM O CURRÍCULO ---

func TestHabilidadeRepository_AddHabilidadeAoCurriculo_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	var vinculoID int64 = 100
	var habilidadeID int64 = 1
	vinculo := &empregabilidade.CurriculoHabilidade{
		ID:           vinculoID,
		CPF:          "12345678901",
		IDHabilidade: habilidadeID,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_curriculo_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vinculoID))
	mock.ExpectCommit()

	err := repo.AddHabilidadeAoCurriculo(ctx, vinculo)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestHabilidadeRepository_ListHabilidadesPorCPF_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	cpf := "12345678901"
	var vinculoID int64 = 100
	var habilidadeID int64 = 1

	// 1. Busca dos vínculos do currículo (usando os dois mapeamentos id_habilidade e habilidade_id)
	mock.ExpectQuery(`SELECT \* FROM "emp_curriculo_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "cpf", "id_habilidade", "habilidade_id"}).
			AddRow(vinculoID, cpf, habilidadeID, habilidadeID))

	// 2. Preload da Habilidade
	mock.ExpectQuery(`SELECT \* FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(habilidadeID, "Pintura Predial"))

	// 3. Preload das Áreas vinculadas à Habilidade
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, err := repo.ListHabilidadesPorCPF(ctx, cpf)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, cpf, result[0].CPF)
	assert.NoError(t, mock.ExpectationsWereMet())
}
