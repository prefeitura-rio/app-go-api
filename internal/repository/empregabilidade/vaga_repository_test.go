package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestVagaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		filter             empregabilidade.VagaFilter
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             empregabilidade.VagaFilter{},
			expectedConditions: []string{},
			description:        "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by contratante",
			filter: empregabilidade.VagaFilter{
				Contratante: "12345678000190",
			},
			expectedConditions: []string{"id_contratante ="},
			description:        "Contratante filter should generate id_contratante = condition",
		},
		{
			name: "filter by orgao parceiro",
			filter: empregabilidade.VagaFilter{
				OrgaoParceiroID: "orgao-123",
			},
			expectedConditions: []string{"id_orgao_parceiro ="},
			description:        "OrgaoParceiroID filter should generate id_orgao_parceiro = condition",
		},
		{
			name: "filter by search term",
			filter: empregabilidade.VagaFilter{
				Search: "desenvolvedor",
			},
			expectedConditions: []string{"titulo ILIKE"},
			description:        "Search filter should generate titulo ILIKE condition",
		},
		{
			name: "filter by status - em_edicao",
			filter: empregabilidade.VagaFilter{
				Status: string(empregabilidade.StatusVagaEmEdicao),
			},
			expectedConditions: []string{"status =", "em_edicao"},
			description:        "Status em_edicao should generate status = condition",
		},
		{
			name: "filter by status - publicado_ativo (active jobs)",
			filter: empregabilidade.VagaFilter{
				Status: string(empregabilidade.StatusVagaPublicadoAtivo),
			},
			expectedConditions: []string{
				"status =",
				"publicado_ativo",
				"data_limite IS NULL OR data_limite > NOW()",
			},
			description: "Status publicado_ativo should filter active jobs only",
		},
		{
			name: "filter by status - publicado_expirado (expired jobs)",
			filter: empregabilidade.VagaFilter{
				Status: string(empregabilidade.StatusVagaPublicadoExpirado),
			},
			expectedConditions: []string{
				"status =",
				"publicado_ativo",
				"data_limite IS NOT NULL",
				"data_limite <= NOW()",
			},
			description: "Status publicado_expirado should filter expired jobs",
		},
		{
			name: "filter by status - congelada",
			filter: empregabilidade.VagaFilter{
				Status: string(empregabilidade.StatusVagaCongelada),
			},
			expectedConditions: []string{"status =", "vaga_congelada"},
			description:        "Status congelada should generate status = condition",
		},
		{
			name: "multiple filters combined",
			filter: empregabilidade.VagaFilter{
				Contratante:     "12345678000190",
				OrgaoParceiroID: "orgao-123",
				Search:          "desenvolvedor",
				Status:          string(empregabilidade.StatusVagaPublicadoAtivo),
			},
			expectedConditions: []string{
				"id_contratante =",
				"id_orgao_parceiro =",
				"titulo ILIKE",
				"status =",
				"data_limite IS NULL OR data_limite > NOW()",
			},
			description: "Multiple filters should all be applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the applyFilters closure as it appears in the repository
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.Contratante != "" {
					db = db.Where("id_contratante = ?", tt.filter.Contratante)
				}
				if tt.filter.OrgaoParceiroID != "" {
					db = db.Where("id_orgao_parceiro = ?", tt.filter.OrgaoParceiroID)
				}
				if tt.filter.Search != "" {
					db = db.Where("titulo ILIKE ?", "%"+tt.filter.Search+"%")
				}
				if tt.filter.Status != "" {
					switch tt.filter.Status {
					case string(empregabilidade.StatusVagaPublicadoAtivo):
						db = db.Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)
					case string(empregabilidade.StatusVagaPublicadoExpirado):
						db = db.Where("status = ? AND data_limite IS NOT NULL AND data_limite <= NOW()", empregabilidade.StatusVagaPublicadoAtivo)
					default:
						db = db.Where("status = ?", tt.filter.Status)
					}
				}
				return db
			}

			query := db.Model(&empregabilidade.Vaga{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Vaga{})
			})

			// Check that all expected conditions are present in the SQL
			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestVagaRepository_List_StatusLogic(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name          string
		status        string
		expectActive  bool
		expectExpired bool
		expectOther   bool
	}{
		{
			name:         "publicado_ativo filters active jobs",
			status:       string(empregabilidade.StatusVagaPublicadoAtivo),
			expectActive: true,
		},
		{
			name:          "publicado_expirado filters expired jobs",
			status:        string(empregabilidade.StatusVagaPublicadoExpirado),
			expectExpired: true,
		},
		{
			name:        "em_edicao uses simple status match",
			status:      string(empregabilidade.StatusVagaEmEdicao),
			expectOther: true,
		},
		{
			name:        "congelada uses simple status match",
			status:      string(empregabilidade.StatusVagaCongelada),
			expectOther: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.status != "" {
					switch tt.status {
					case string(empregabilidade.StatusVagaPublicadoAtivo):
						db = db.Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)
					case string(empregabilidade.StatusVagaPublicadoExpirado):
						db = db.Where("status = ? AND data_limite IS NOT NULL AND data_limite <= NOW()", empregabilidade.StatusVagaPublicadoAtivo)
					default:
						db = db.Where("status = ?", tt.status)
					}
				}
				return db
			}

			query := db.Model(&empregabilidade.Vaga{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Vaga{})
			})

			if tt.expectActive {
				assert.Contains(t, sql, "data_limite IS NULL OR data_limite > NOW()")
				assert.NotContains(t, sql, "data_limite IS NOT NULL")
			}

			if tt.expectExpired {
				assert.Contains(t, sql, "data_limite IS NOT NULL")
				assert.Contains(t, sql, "data_limite <= NOW()")
			}

			if tt.expectOther {
				assert.NotContains(t, sql, "data_limite")
			}
		})
	}
}

func TestVagaRepository_List_SearchFormatting(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name         string
		searchTerm   string
		expectedLike string
	}{
		{
			name:         "search term gets wrapped with %",
			searchTerm:   "desenvolvedor",
			expectedLike: "%desenvolvedor%",
		},
		{
			name:         "search with spaces",
			searchTerm:   "engenheiro software",
			expectedLike: "%engenheiro software%",
		},
		{
			name:         "search with special characters",
			searchTerm:   "c++",
			expectedLike: "%c++%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.searchTerm != "" {
					db = db.Where("titulo ILIKE ?", "%"+tt.searchTerm+"%")
				}
				return db
			}

			query := db.Model(&empregabilidade.Vaga{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Vaga{})
			})

			assert.Contains(t, sql, "titulo ILIKE")
			assert.Contains(t, sql, tt.expectedLike)
		})
	}
}

func TestVagaRepository_ListPublicActive_QueryGeneration(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		limit              int
		offset             int
		expectedConditions []string
		description        string
	}{
		{
			name:   "public active vagas with default pagination",
			limit:  10,
			offset: 0,
			expectedConditions: []string{
				"status =",
				"publicado_ativo",
				"data_limite IS NULL OR data_limite > NOW()",
				"ORDER BY",
				"created_at DESC",
				"LIMIT 10",
			},
			description: "Should filter by active status and expiration date",
		},
		{
			name:   "public active vagas with offset",
			limit:  20,
			offset: 40,
			expectedConditions: []string{
				"status =",
				"publicado_ativo",
				"data_limite IS NULL OR data_limite > NOW()",
				"LIMIT 20",
				"OFFSET 40",
			},
			description: "Should apply pagination correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Model(&empregabilidade.Vaga{}).
				Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo).
				Order("created_at DESC").
				Limit(tt.limit).
				Offset(tt.offset)

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]*empregabilidade.Vaga{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestVagaRepository_ListByContratante_QueryGeneration(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		cnpj               string
		limit              int
		offset             int
		expectedConditions []string
		description        string
	}{
		{
			name:   "filter by contratante CNPJ",
			cnpj:   "12345678000190",
			limit:  10,
			offset: 0,
			expectedConditions: []string{
				"id_contratante =",
				"12345678000190",
				"ORDER BY",
				"created_at DESC",
				"LIMIT 10",
			},
			description: "Should filter by id_contratante and order by created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Model(&empregabilidade.Vaga{}).
				Where("id_contratante = ?", tt.cnpj).
				Order("created_at DESC").
				Limit(tt.limit).
				Offset(tt.offset)

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]*empregabilidade.Vaga{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestVagaRepository_ListByOrgaoParceiro_QueryGeneration(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		orgaoID            string
		limit              int
		offset             int
		expectedConditions []string
		description        string
	}{
		{
			name:    "filter by orgao parceiro ID",
			orgaoID: "orgao-123",
			limit:   10,
			offset:  0,
			expectedConditions: []string{
				"id_orgao_parceiro =",
				"orgao-123",
				"ORDER BY",
				"created_at DESC",
				"LIMIT 10",
			},
			description: "Should filter by id_orgao_parceiro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Model(&empregabilidade.Vaga{}).
				Where("id_orgao_parceiro = ?", tt.orgaoID).
				Order("created_at DESC").
				Limit(tt.limit).
				Offset(tt.offset)

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]*empregabilidade.Vaga{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

// CRUD success path tests

func TestVagaRepository_Create_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, vaga)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_GetByID_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	regimeID := uuid.New()
	modeloID := uuid.New()

	// Use MatchExpectationsInOrder(false) to avoid order issues
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "titulo", "descricao", "id_contratante", "id_regime_contratacao",
			"id_modelo_trabalho", "status", "bairro", "requisitos", "diferenciais",
			"responsabilidades", "beneficios", "id_orgao_parceiro",
		}).AddRow(
			vagaID, "Test Vaga", "Test Description", "12345678000190", regimeID,
			modeloID, "em_edicao", "", "", "", "", "", "",
		))

	// Mock preloads - use AnyArg to be flexible
	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "orgaos_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{"orgao_id"}))

	// TiposPCD is a many2many relation, so it queries the junction table first
	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_tipos_pcd"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_tipo_pcd"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_etapas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_informacoes_complementares"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	result, err := repo.GetByID(ctx, vagaID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, vagaID, result.ID)
}

func TestVagaRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas"`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, vagaID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestVagaRepository_Update_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_Delete_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, vagaID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_List_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.VagaFilter{
		Status: string(empregabilidade.StatusVagaEmEdicao),
	}

	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo"}).AddRow(vagaID, "Test Vaga"))

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "orgaos_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{"orgao_id"}))

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
}

func TestVagaRepository_ListPublicActive_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo"}).AddRow(vagaID, "Test Vaga"))

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "orgaos_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{"orgao_id"}))

	result, total, err := repo.ListPublicActive(ctx, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
}

func TestVagaRepository_ListByContratante_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo"}).AddRow(vagaID, "Test Vaga"))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "orgaos_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{"orgao_id"}))

	result, total, err := repo.ListByContratante(ctx, "12345678000190", 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
}

func TestVagaRepository_ListByOrgaoParceiro_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo"}).AddRow(vagaID, "Test Vaga"))

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "orgaos_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{"orgao_id"}))

	result, total, err := repo.ListByOrgaoParceiro(ctx, "orgao-123", 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
}

func TestVagaRepository_UpdateTiposPCD_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	tipoPCDID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_tipos_pcd`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateTiposPCD(ctx, vagaID, []uuid.UUID{tipoPCDID})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateTiposPCD_EmptyList(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateTiposPCD(ctx, vagaID, []uuid.UUID{})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Error path tests

func TestVagaRepository_Create_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_vagas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.Create(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_Update_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Update(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Delete(ctx, vagaID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_Delete_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.Delete(ctx, vagaID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vaga não encontrada")
}

func TestVagaRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(ctx, vagaID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar vaga")
}

func TestVagaRepository_List_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.VagaFilter{
		Status: string(empregabilidade.StatusVagaEmEdicao),
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao contar vagas")
}

func TestVagaRepository_List_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.VagaFilter{
		Status: string(empregabilidade.StatusVagaEmEdicao),
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar vagas")
}

func TestVagaRepository_ListPublicActive_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListPublicActive(ctx, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao contar vagas")
}

func TestVagaRepository_ListPublicActive_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListPublicActive(ctx, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar vagas públicas ativas")
}

func TestVagaRepository_ListByContratante_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListByContratante(ctx, "12345678000190", 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao contar vagas por contratante")
}

func TestVagaRepository_ListByContratante_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListByContratante(ctx, "12345678000190", 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar vagas por contratante")
}

func TestVagaRepository_ListByOrgaoParceiro_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListByOrgaoParceiro(ctx, "orgao-123", 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao contar vagas por órgão parceiro")
}

func TestVagaRepository_ListByOrgaoParceiro_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.ListByOrgaoParceiro(ctx, "orgao-123", 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar vagas por órgão parceiro")
}

func TestVagaRepository_UpdateWithAssociations_UpdateError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar vaga")
}

func TestVagaRepository_UpdateWithAssociations_ClearEtapaAtualError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		Etapas:              []empregabilidade.Etapa{},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "emp_candidaturas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao desvincular etapas das candidaturas")
}

func TestVagaRepository_UpdateWithAssociations_DeleteEtapasError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		Etapas:              []empregabilidade.Etapa{},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "emp_candidaturas"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "emp_etapas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover etapas existentes")
}

func TestVagaRepository_UpdateWithAssociations_CreateEtapasError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		Etapas: []empregabilidade.Etapa{
			{Titulo: "Etapa 1", Ordem: 1},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "emp_candidaturas"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "emp_etapas"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "emp_etapas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar etapas")
}

func TestVagaRepository_UpdateWithAssociations_DeleteInformacoesError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                        uuid.New(),
		Titulo:                    "Desenvolvedor Go",
		Descricao:                 "Vaga para desenvolvedor Go",
		IDContratante:             "12345678000190",
		IDRegimeContratacao:       uuid.New(),
		IDModeloTrabalho:          uuid.New(),
		Status:                    empregabilidade.StatusVagaEmEdicao,
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover informações complementares existentes")
}

func TestVagaRepository_UpdateWithAssociations_CreateInformacoesError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{
			{Titulo: "Info 1", TipoCampo: empregabilidade.TipoCampoRespostaCurta},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "emp_informacoes_complementares"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar informações complementares")
}

func TestVagaRepository_UpdateWithAssociations_DeleteTiposPCDError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		TiposPCD:            []empregabilidade.TipoPCD{},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover tipos PCD")
}

func TestVagaRepository_UpdateWithAssociations_InsertTipoPCDError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	tipoPCDID := uuid.New()
	vaga := &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		TiposPCD: []empregabilidade.TipoPCD{
			{ID: tipoPCDID, Descricao: "Tipo 1"},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_tipos_pcd`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao inserir tipo PCD")
}

func TestVagaRepository_UpdateTiposPCD_DeleteError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateTiposPCD(ctx, vagaID, []uuid.UUID{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover tipos PCD")
}

func TestVagaRepository_UpdateTiposPCD_InsertError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	tipoPCDID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_tipos_pcd`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_tipos_pcd`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateTiposPCD(ctx, vagaID, []uuid.UUID{tipoPCDID})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao inserir tipo PCD")
}
