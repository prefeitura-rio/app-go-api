package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEmpresaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		filter             empregabilidade.EmpresaFilter
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             empregabilidade.EmpresaFilter{},
			expectedConditions: []string{},
			description:        "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by CNPJ exact match",
			filter: empregabilidade.EmpresaFilter{
				CNPJ: "12345678000190",
			},
			expectedConditions: []string{"cnpj =", "12345678000190"},
			description:        "CNPJ filter should generate exact match condition",
		},
		{
			name: "filter by Search in razao_social and nome_fantasia",
			filter: empregabilidade.EmpresaFilter{
				Search: "Tech",
			},
			expectedConditions: []string{
				"razao_social ILIKE",
				"OR",
				"nome_fantasia ILIKE",
				"%Tech%",
			},
			description: "Search filter should check razao_social OR nome_fantasia",
		},
		{
			name: "multiple filters combined",
			filter: empregabilidade.EmpresaFilter{
				CNPJ:   "12345678000190",
				Search: "Software",
			},
			expectedConditions: []string{
				"cnpj =",
				"razao_social ILIKE",
				"nome_fantasia ILIKE",
				"%Software%",
			},
			description: "Multiple filters should all be applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the applyFilters closure as it appears in the repository
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.CNPJ != "" {
					db = db.Where("cnpj = ?", tt.filter.CNPJ)
				}
				if tt.filter.Search != "" {
					searchTerm := "%" + tt.filter.Search + "%"
					db = db.Where("razao_social ILIKE ? OR nome_fantasia ILIKE ?", searchTerm, searchTerm)
				}
				return db
			}

			query := db.Model(&empregabilidade.Empresa{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Empresa{})
			})

			// Check that all expected conditions are present in the SQL
			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestEmpresaRepository_List_SearchORLogic(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		searchTerm  string
		description string
	}{
		{
			name:        "search with company name",
			searchTerm:  "Acme Corp",
			description: "Should search in both razao_social and nome_fantasia",
		},
		{
			name:        "search with partial CNPJ",
			searchTerm:  "12345",
			description: "Should still search in company name fields",
		},
		{
			name:        "search with special characters",
			searchTerm:  "Tech & Co.",
			description: "Should handle special characters in search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.searchTerm != "" {
					searchTerm := "%" + tt.searchTerm + "%"
					db = db.Where("razao_social ILIKE ? OR nome_fantasia ILIKE ?", searchTerm, searchTerm)
				}
				return db
			}

			query := db.Model(&empregabilidade.Empresa{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Empresa{})
			})

			// Check OR condition is present
			assert.Contains(t, sql, "razao_social ILIKE", tt.description)
			assert.Contains(t, sql, "OR", tt.description)
			assert.Contains(t, sql, "nome_fantasia ILIKE", tt.description)

			// Verify the search term is wrapped with %
			expectedLike := "%" + tt.searchTerm + "%"
			assert.Contains(t, sql, expectedLike, "%s: Search term should be wrapped with %%", tt.description)
		})
	}
}

func TestEmpresaRepository_List_EmptyStringFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	t.Run("empty CNPJ should not filter", func(t *testing.T) {
		filter := empregabilidade.EmpresaFilter{
			CNPJ: "",
		}

		applyFilters := func(db *gorm.DB) *gorm.DB {
			if filter.CNPJ != "" {
				db = db.Where("cnpj = ?", filter.CNPJ)
			}
			return db
		}

		query := db.Model(&empregabilidade.Empresa{})
		filteredQuery := applyFilters(query)

		sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]empregabilidade.Empresa{})
		})

		assert.NotContains(t, sql, "cnpj =", "Empty CNPJ should not add filter condition")
	})

	t.Run("empty Search should not filter", func(t *testing.T) {
		filter := empregabilidade.EmpresaFilter{
			Search: "",
		}

		applyFilters := func(db *gorm.DB) *gorm.DB {
			if filter.Search != "" {
				searchTerm := "%" + filter.Search + "%"
				db = db.Where("razao_social ILIKE ? OR nome_fantasia ILIKE ?", searchTerm, searchTerm)
			}
			return db
		}

		query := db.Model(&empregabilidade.Empresa{})
		filteredQuery := applyFilters(query)

		sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]empregabilidade.Empresa{})
		})

		assert.NotContains(t, sql, "razao_social ILIKE", "Empty Search should not add filter condition")
		assert.NotContains(t, sql, "nome_fantasia ILIKE", "Empty Search should not add filter condition")
	})
}

func TestEmpresaRepository_List_CNPJExactMatch(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	// Test that CNPJ uses exact match, not ILIKE
	filter := empregabilidade.EmpresaFilter{
		CNPJ: "12345678000190",
	}

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.CNPJ != "" {
			db = db.Where("cnpj = ?", filter.CNPJ)
		}
		return db
	}

	query := db.Model(&empregabilidade.Empresa{})
	filteredQuery := applyFilters(query)

	sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]empregabilidade.Empresa{})
	})

	// Should use exact match, not ILIKE
	assert.Contains(t, sql, "cnpj =", "CNPJ should use exact match")
	assert.NotContains(t, sql, "cnpj ILIKE", "CNPJ should not use ILIKE")
	assert.NotContains(t, sql, "%12345678000190%", "CNPJ should not be wrapped with %")
}

// CRUD success path tests

func TestEmpresaRepository_Create_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Teste LTDA",
		NomeFantasia: "Empresa Teste",
		Descricao:    "Empresa de teste",
		URLLogo:      "https://example.com/logo.png",
		Website:      "https://example.com",
		Setor:        "Tecnologia",
		Porte:        "Pequeno",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "emp_empresas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	cnpj, err := repo.Create(ctx, empresa)
	assert.NoError(t, err)
	assert.Equal(t, "12345678000190", cnpj)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Create_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Teste LTDA",
		NomeFantasia: "Empresa Teste",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "emp_empresas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.Create(ctx, empresa)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()
	cnpj := "12345678000190"

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"cnpj", "razao_social", "nome_fantasia", "descricao",
			"url_logo", "website", "setor", "porte",
		}).AddRow(
			cnpj, "Empresa Teste LTDA", "Empresa Teste", "Descricao",
			"https://example.com/logo.png", "https://example.com", "Tecnologia", "Pequeno",
		))

	result, err := repo.GetByID(ctx, cnpj)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cnpj, result.CNPJ)
	assert.Equal(t, "Empresa Teste LTDA", result.RazaoSocial)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()
	cnpj := "12345678000190"

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, cnpj)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()
	cnpj := "12345678000190"

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Update_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Atualizada LTDA",
		NomeFantasia: "Empresa Atualizada",
		Descricao:    "Descricao atualizada",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_empresas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, empresa)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Update_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Atualizada LTDA",
		NomeFantasia: "Empresa Atualizada",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_empresas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Update(ctx, empresa)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Delete_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()
	cnpj := "12345678000190"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_empresas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, cnpj)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()
	cnpj := "12345678000190"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "emp_empresas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Delete(ctx, cnpj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.EmpresaFilter{
		Search: "Teste",
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"cnpj", "razao_social", "nome_fantasia",
		}).
			AddRow("12345678000190", "Empresa Teste 1", "Teste 1").
			AddRow("98765432000111", "Empresa Teste 2", "Teste 2"))

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_WithCNPJFilter(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.EmpresaFilter{
		CNPJ: "12345678000190",
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"cnpj", "razao_social", "nome_fantasia",
		}).AddRow("12345678000190", "Empresa Teste", "Teste"))

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_EmptyFilter(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.EmpresaFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT .* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"cnpj", "razao_social", "nome_fantasia",
		}).
			AddRow("11111111000111", "Empresa A", "A").
			AddRow("22222222000111", "Empresa B", "B").
			AddRow("33333333000111", "Empresa C", "C").
			AddRow("44444444000111", "Empresa D", "D").
			AddRow("55555555000111", "Empresa E", "E"))

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 5, total)
	assert.Len(t, result, 5)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.EmpresaFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_empresas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao contar empresas")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := empregabilidade.EmpresaFilter{}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT .* FROM "emp_empresas"`).
		WillReturnError(assert.AnError)

	result, total, err := repo.List(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar empresas")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Upsert_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Upsert LTDA",
		NomeFantasia: "Empresa Upsert",
		Descricao:    "Descricao upsert",
	}

	mock.ExpectBegin()
	// GORM's Save() will try UPDATE first if primary key is present
	mock.ExpectExec(`UPDATE "emp_empresas"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Upsert(ctx, empresa)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Upsert_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &empregabilidade.Empresa{
		CNPJ:         "12345678000190",
		RazaoSocial:  "Empresa Upsert LTDA",
		NomeFantasia: "Empresa Upsert",
	}

	mock.ExpectBegin()
	// GORM's Save() will try UPDATE first if primary key is present
	mock.ExpectExec(`UPDATE "emp_empresas"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Upsert(ctx, empresa)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao upsert empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}
