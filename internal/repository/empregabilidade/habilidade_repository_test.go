package empregabilidade_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
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

	areaID := uuid.New()

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

				if tt.filter.AreaAtuacaoID != uuid.Nil {
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

	habilidadeID := uuid.New()
	entity := &empregabilidade.Habilidade{
		ID:   habilidadeID,
		Nome: "Desenvolvimento Go",
	}

	mock.ExpectBegin()
	// Usamos regexp.QuoteMeta para evitar erros de sintaxe de regex com aspas e símbolos do SQL
	// e sqlmock.AnyArg() para lidar com timestamps e UUIDs sem falha de tipo
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
	// Sem predefinir a quantidade exata de WithArgs() para não quebrar
	// quando o ID não é informado previamente na struct.
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
	habilidadeID := uuid.New()

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
	habilidadeID := uuid.New()

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
		ID:        uuid.New(),
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
	habilidadeID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_habilidades"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteHabilidade(ctx, habilidadeID)
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
			AddRow(uuid.New(), "Acabamento em gesso"))

	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, total, err := repo.ListHabilidades(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHabilidadeRepository_ListAreas_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()
	filter := empregabilidade.AreaAtuacaoFilter{Search: "Confecção"}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "area_atuacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM "area_atuacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(uuid.New(), "Confecção de Calçados").
			AddRow(uuid.New(), "Confecção de Roupas"))

	result, total, err := repo.ListAreas(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- TESTES DE VÍNCULO COM O CURRÍCULO ---

func TestHabilidadeRepository_AddHabilidadeAoCurriculo_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := repoEmpregabilidade.NewHabilidadeRepository(db)
	ctx := context.Background()

	vinculoID := uuid.New()
	vinculo := &empregabilidade.CurriculoHabilidade{
		ID:           vinculoID,
		CPF:          "12345678901",
		IDHabilidade: uuid.New(),
	}

	mock.ExpectBegin()
	// Trocado de ExpectExec para ExpectQuery por causa do RETURNING "id" do GORM
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
	habilidadeID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "emp_curriculo_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "cpf", "id_habilidade"}).
			AddRow(uuid.New(), cpf, habilidadeID))

	// Preload Habilidade
	mock.ExpectQuery(`SELECT \* FROM "emp_habilidades"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(habilidadeID, "Pintura Predial"))

	// Preload Áreas da Habilidade
	mock.ExpectQuery(`SELECT \* FROM "area_atuacao_habilidade"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_habilidade", "id_area_atuacao"}))

	result, err := repo.ListHabilidadesPorCPF(ctx, cpf)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, cpf, result[0].CPF)
	assert.NoError(t, mock.ExpectationsWereMet())
}
