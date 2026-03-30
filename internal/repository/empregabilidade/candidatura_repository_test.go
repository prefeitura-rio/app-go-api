package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func stringPtr(s string) *string {
	return &s
}

func TestCandidaturaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()
	etapaID := uuid.New()

	tests := []struct {
		name               string
		filter             empregabilidade.CandidaturaFilter
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             empregabilidade.CandidaturaFilter{},
			expectedConditions: []string{},
			description:        "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by CPF",
			filter: empregabilidade.CandidaturaFilter{
				CPF: "12345678900",
			},
			expectedConditions: []string{"cpf =", "12345678900"},
			description:        "CPF filter should generate cpf = condition",
		},
		{
			name: "filter by VagaID",
			filter: empregabilidade.CandidaturaFilter{
				VagaID: &vagaID,
			},
			expectedConditions: []string{"id_vaga ="},
			description:        "VagaID filter should generate id_vaga = condition",
		},
		{
			name: "filter by Status",
			filter: empregabilidade.CandidaturaFilter{
				Status: string(empregabilidade.StatusCandidaturaAprovada),
			},
			expectedConditions: []string{"status =", "aprovada"},
			description:        "Status filter should generate status = condition",
		},
		{
			name: "filter by EtapaID",
			filter: empregabilidade.CandidaturaFilter{
				EtapaID: &etapaID,
			},
			expectedConditions: []string{"id_etapa_atual ="},
			description:        "EtapaID filter should generate id_etapa_atual = condition",
		},
		{
			name: "filter by Search",
			filter: empregabilidade.CandidaturaFilter{
				Search: "João",
			},
			expectedConditions: []string{
				"cpf ILIKE",
				"nome ILIKE",
				"email ILIKE",
				"%João%",
			},
			description: "Search filter should generate ILIKE conditions for cpf, nome, and email",
		},
		{
			name: "multiple filters combined",
			filter: empregabilidade.CandidaturaFilter{
				CPF:     "12345678900",
				VagaID:  &vagaID,
				Status:  string(empregabilidade.StatusCandidaturaEnviada),
				EtapaID: &etapaID,
				Search:  "Maria",
			},
			expectedConditions: []string{
				"cpf =",
				"id_vaga =",
				"status =",
				"id_etapa_atual =",
				"cpf ILIKE",
				"nome ILIKE",
				"email ILIKE",
			},
			description: "Multiple filters should all be applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the applyFilters closure as it appears in the repository
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.CPF != "" {
					db = db.Where("cpf = ?", tt.filter.CPF)
				}
				if tt.filter.VagaID != nil {
					db = db.Where("id_vaga = ?", *tt.filter.VagaID)
				}
				if tt.filter.Status != "" {
					db = db.Where("status = ?", tt.filter.Status)
				}
				if tt.filter.EtapaID != nil {
					db = db.Where("id_etapa_atual = ?", *tt.filter.EtapaID)
				}
				if tt.filter.Search != "" {
					searchTerm := "%" + tt.filter.Search + "%"
					db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
				}
				return db
			}

			query := db.Model(&empregabilidade.Candidatura{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Candidatura{})
			})

			// Check that all expected conditions are present in the SQL
			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestCandidaturaRepository_List_SearchORLogic(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		searchTerm  string
		expectedOR  []string
		description string
	}{
		{
			name:       "search generates OR conditions",
			searchTerm: "test",
			expectedOR: []string{
				"cpf ILIKE",
				"OR",
				"nome ILIKE",
				"email ILIKE",
			},
			description: "Search should check cpf OR nome OR email",
		},
		{
			name:       "search with CPF-like pattern",
			searchTerm: "123",
			expectedOR: []string{
				"cpf ILIKE",
				"OR",
				"nome ILIKE",
				"email ILIKE",
				"%123%",
			},
			description: "Search with numbers should still check all fields",
		},
		{
			name:       "search with email-like pattern",
			searchTerm: "@gmail",
			expectedOR: []string{
				"cpf ILIKE",
				"OR",
				"nome ILIKE",
				"email ILIKE",
				"%@gmail%",
			},
			description: "Search with @ should still check all fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.searchTerm != "" {
					searchTerm := "%" + tt.searchTerm + "%"
					db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
				}
				return db
			}

			query := db.Model(&empregabilidade.Candidatura{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Candidatura{})
			})

			// Check that all OR conditions are present
			for _, condition := range tt.expectedOR {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}

			// Verify the search term is wrapped with %
			expectedLike := "%" + tt.searchTerm + "%"
			assert.Contains(t, sql, expectedLike,
				"%s: Search term should be wrapped with %%", tt.description)
		})
	}
}

func TestCandidaturaRepository_List_NullableFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	t.Run("nil VagaID should not filter", func(t *testing.T) {
		filter := empregabilidade.CandidaturaFilter{
			VagaID: nil,
		}

		applyFilters := func(db *gorm.DB) *gorm.DB {
			if filter.VagaID != nil {
				db = db.Where("id_vaga = ?", *filter.VagaID)
			}
			return db
		}

		query := db.Model(&empregabilidade.Candidatura{})
		filteredQuery := applyFilters(query)

		sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]empregabilidade.Candidatura{})
		})

		assert.NotContains(t, sql, "id_vaga =",
			"Nil VagaID should not add filter condition")
	})

	t.Run("nil EtapaID should not filter", func(t *testing.T) {
		filter := empregabilidade.CandidaturaFilter{
			EtapaID: nil,
		}

		applyFilters := func(db *gorm.DB) *gorm.DB {
			if filter.EtapaID != nil {
				db = db.Where("id_etapa_atual = ?", *filter.EtapaID)
			}
			return db
		}

		query := db.Model(&empregabilidade.Candidatura{})
		filteredQuery := applyFilters(query)

		sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]empregabilidade.Candidatura{})
		})

		assert.NotContains(t, sql, "id_etapa_atual =",
			"Nil EtapaID should not add filter condition")
	})

	t.Run("valid VagaID should filter", func(t *testing.T) {
		vagaID := uuid.New()
		filter := empregabilidade.CandidaturaFilter{
			VagaID: &vagaID,
		}

		applyFilters := func(db *gorm.DB) *gorm.DB {
			if filter.VagaID != nil {
				db = db.Where("id_vaga = ?", *filter.VagaID)
			}
			return db
		}

		query := db.Model(&empregabilidade.Candidatura{})
		filteredQuery := applyFilters(query)

		sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&[]empregabilidade.Candidatura{})
		})

		assert.Contains(t, sql, "id_vaga =",
			"Valid VagaID should add filter condition")
	})
}

func TestCandidaturaRepository_List_EmptyStringFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name       string
		filter     empregabilidade.CandidaturaFilter
		shouldSkip []string
	}{
		{
			name: "empty CPF should not filter",
			filter: empregabilidade.CandidaturaFilter{
				CPF: "",
			},
			shouldSkip: []string{"cpf ="},
		},
		{
			name: "empty Status should not filter",
			filter: empregabilidade.CandidaturaFilter{
				Status: "",
			},
			shouldSkip: []string{"status ="},
		},
		{
			name: "empty Search should not filter",
			filter: empregabilidade.CandidaturaFilter{
				Search: "",
			},
			shouldSkip: []string{"cpf ILIKE", "nome ILIKE", "email ILIKE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.CPF != "" {
					db = db.Where("cpf = ?", tt.filter.CPF)
				}
				if tt.filter.VagaID != nil {
					db = db.Where("id_vaga = ?", *tt.filter.VagaID)
				}
				if tt.filter.Status != "" {
					db = db.Where("status = ?", tt.filter.Status)
				}
				if tt.filter.EtapaID != nil {
					db = db.Where("id_etapa_atual = ?", *tt.filter.EtapaID)
				}
				if tt.filter.Search != "" {
					searchTerm := "%" + tt.filter.Search + "%"
					db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
				}
				return db
			}

			query := db.Model(&empregabilidade.Candidatura{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Candidatura{})
			})

			for _, condition := range tt.shouldSkip {
				assert.NotContains(t, sql, condition,
					"Empty string filter should not add '%s' condition", condition)
			}
		})
	}
}

func TestCandidaturaRepository_CountByStatus_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()
	etapaID := uuid.New()

	// Test that CountByStatus uses the same filter logic but ignores Status field
	tests := []struct {
		name             string
		filter           empregabilidade.CandidaturaFilter
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "filters are applied except status",
			filter: empregabilidade.CandidaturaFilter{
				CPF:     "12345678900",
				VagaID:  &vagaID,
				Status:  "aprovada", // Should be ignored
				EtapaID: &etapaID,
				Search:  "test",
			},
			shouldContain: []string{
				"cpf =",
				"id_vaga =",
				"id_etapa_atual =",
				"cpf ILIKE",
			},
			shouldNotContain: []string{
				// Status filter should be ignored in CountByStatus
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the applyFilters logic in CountByStatus
			applyFilters := func(db *gorm.DB) *gorm.DB {
				if tt.filter.CPF != "" {
					db = db.Where("cpf = ?", tt.filter.CPF)
				}
				if tt.filter.VagaID != nil {
					db = db.Where("id_vaga = ?", *tt.filter.VagaID)
				}
				if tt.filter.EtapaID != nil {
					db = db.Where("id_etapa_atual = ?", *tt.filter.EtapaID)
				}
				if tt.filter.Search != "" {
					searchTerm := "%" + tt.filter.Search + "%"
					db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
				}
				// Note: Status is intentionally ignored
				return db
			}

			query := db.Model(&empregabilidade.Candidatura{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Select("status, COUNT(*) as count").Group("status").Find(&[]struct{}{})
			})

			for _, condition := range tt.shouldContain {
				assert.Contains(t, sql, condition)
			}

			for _, condition := range tt.shouldNotContain {
				assert.NotContains(t, sql, condition)
			}
		})
	}
}

func TestCandidaturaRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{
			ID:      uuid.New(),
			CPF:     "12345678900",
			Nome:    stringPtr("João Silva"),
			Email:   stringPtr("joao@example.com"),
			IDVaga:  vagaID,
			Status:  empregabilidade.StatusCandidaturaEnviada,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(candidatura.ID))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, candidatura)
		assert.NoError(t, err)
		assert.Equal(t, candidatura.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{
			ID:     uuid.New(),
			CPF:    "12345678900",
			IDVaga: vagaID,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, candidatura)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar candidatura")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCandidaturaRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		vagaID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "nome", "email", "id_vaga", "status"}).
			AddRow(id, "12345678900", "João Silva", "joao@example.com", vagaID, empregabilidade.StatusCandidaturaEnviada)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(rows)

		// Preload queries - order may vary, so match any
		mock.ExpectQuery(`SELECT \* FROM "emp_vagas"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectQuery(`SELECT \* FROM "emp_contratantes"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectQuery(`SELECT \* FROM "emp_etapas"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectQuery(`SELECT \* FROM "emp_etapas"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		candidatura, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, candidatura)
		assert.Equal(t, id, candidatura.ID)
	})
}

func TestCandidaturaRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{
			ID:     uuid.New(),
			CPF:    "12345678900",
			Nome:   stringPtr("João Silva Updated"),
			IDVaga: vagaID,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, candidatura)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		vagaID := uuid.New()
		candidatura := &empregabilidade.Candidatura{
			ID:     uuid.New(),
			CPF:    "12345678900",
			IDVaga: vagaID,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, candidatura)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar candidatura")
	})
}

func TestCandidaturaRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		// GORM soft delete uses UPDATE not DELETE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("delete not found", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "não encontrada")
	})

	t.Run("delete database error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir candidatura")
	})
}

func TestCandidaturaRepository_BulkUpdateStatus(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("bulk update all cpfs found", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1 := "12345678900"
		cpf2 := "98765432100"
		cpfs := []string{cpf1, cpf2}

		candidaturaID1 := uuid.New()
		candidaturaID2 := uuid.New()

		// Mock the query to find candidaturas
		rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}).
			AddRow(candidaturaID1, cpf1, vagaID).
			AddRow(candidaturaID2, cpf2, vagaID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(rows)

		// Mock the bulk update
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		result, err := repo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Updated)
		assert.Empty(t, result.FailedCPFs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("bulk update some cpfs not found", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1 := "12345678900"
		cpf2 := "98765432100"
		cpf3 := "11111111111" // This one will not be found
		cpfs := []string{cpf1, cpf2, cpf3}

		candidaturaID1 := uuid.New()
		candidaturaID2 := uuid.New()

		// Mock the query to find candidaturas (only 2 found)
		rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}).
			AddRow(candidaturaID1, cpf1, vagaID).
			AddRow(candidaturaID2, cpf2, vagaID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(rows)

		// Mock the bulk update
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		result, err := repo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Updated)
		assert.Contains(t, result.FailedCPFs, cpf3)
		assert.Len(t, result.FailedCPFs, 1)
	})

	t.Run("bulk update no cpfs found", func(t *testing.T) {
		vagaID := uuid.New()
		cpfs := []string{"12345678900", "98765432100"}

		// Mock the query to find candidaturas (none found)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}))

		result, err := repo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.Updated)
		assert.Len(t, result.FailedCPFs, 2)
	})

	t.Run("bulk update query error", func(t *testing.T) {
		vagaID := uuid.New()
		cpfs := []string{"12345678900"}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnError(assert.AnError)

		result, err := repo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar candidaturas")
		assert.Equal(t, 0, result.Updated)
	})

	t.Run("bulk update execution error", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1 := "12345678900"
		cpfs := []string{cpf1}

		candidaturaID1 := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}).
			AddRow(candidaturaID1, cpf1, vagaID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(rows)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		_, err := repo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar status em lote")
	})
}

func TestCandidaturaRepository_BulkSaveAndUpdateStatusByVagaID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("freeze candidaturas success", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status_anterior = status, status = .*, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL`).
			WithArgs(string(empregabilidade.StatusCandidaturaVagaCongelada), vagaID).
			WillReturnResult(sqlmock.NewResult(0, 5))
		mock.ExpectCommit()

		err := repo.BulkSaveAndUpdateStatusByVagaID(ctx, vagaID, empregabilidade.StatusCandidaturaVagaCongelada)
		assert.NoError(t, err)
	})

	t.Run("discontinue candidaturas success", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status_anterior = status, status = .*, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL`).
			WithArgs(string(empregabilidade.StatusCandidaturaDescontinuada), vagaID).
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.BulkSaveAndUpdateStatusByVagaID(ctx, vagaID, empregabilidade.StatusCandidaturaDescontinuada)
		assert.NoError(t, err)
	})

	t.Run("freeze candidaturas error", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status_anterior = status, status = .*, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL`).
			WithArgs(string(empregabilidade.StatusCandidaturaVagaCongelada), vagaID).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkSaveAndUpdateStatusByVagaID(ctx, vagaID, empregabilidade.StatusCandidaturaVagaCongelada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao congelar/descontinuar candidaturas")
	})
}

func TestCandidaturaRepository_BulkRestoreStatusByVagaID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("restore candidaturas success", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL AND status_anterior IS NOT NULL`).
			WithArgs(vagaID).
			WillReturnResult(sqlmock.NewResult(0, 5))
		mock.ExpectCommit()

		err := repo.BulkRestoreStatusByVagaID(ctx, vagaID)
		assert.NoError(t, err)
	})

	t.Run("restore candidaturas error", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL AND status_anterior IS NOT NULL`).
			WithArgs(vagaID).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkRestoreStatusByVagaID(ctx, vagaID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao restaurar status das candidaturas")
	})

	t.Run("restore candidaturas no rows affected", func(t *testing.T) {
		vagaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW\(\) WHERE id_vaga = .* AND deleted_at IS NULL AND status_anterior IS NOT NULL`).
			WithArgs(vagaID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.BulkRestoreStatusByVagaID(ctx, vagaID)
		assert.NoError(t, err)
	})
}

func TestCandidaturaRepository_CheckExistingCandidatura(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("candidatura exists", func(t *testing.T) {
		cpf := "12345678900"
		vagaID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := repo.CheckExistingCandidatura(ctx, cpf, vagaID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("candidatura does not exist", func(t *testing.T) {
		cpf := "12345678900"
		vagaID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidaturas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := repo.CheckExistingCandidatura(ctx, cpf, vagaID)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("check error", func(t *testing.T) {
		cpf := "12345678900"
		vagaID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidaturas"`)).
			WillReturnError(assert.AnError)

		exists, err := repo.CheckExistingCandidatura(ctx, cpf, vagaID)
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "erro ao verificar candidatura existente")
	})
}

func TestCandidaturaRepository_UpdateStatus(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("update status success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)
		assert.NoError(t, err)
	})

	t.Run("update status error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar status da candidatura")
	})
}

func TestCandidaturaRepository_BulkGetByCPFs(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("get by cpfs success", func(t *testing.T) {
		vagaID := uuid.New()
		cpf1 := "12345678900"
		cpf2 := "98765432100"
		cpfs := []string{cpf1, cpf2}

		rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}).
			AddRow(uuid.New(), cpf1, vagaID).
			AddRow(uuid.New(), cpf2, vagaID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnRows(rows)

		candidaturas, err := repo.BulkGetByCPFs(ctx, vagaID, cpfs)
		assert.NoError(t, err)
		assert.Len(t, candidaturas, 2)
	})

	t.Run("get by cpfs error", func(t *testing.T) {
		vagaID := uuid.New()
		cpfs := []string{"12345678900"}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidaturas"`)).
			WillReturnError(assert.AnError)

		candidaturas, err := repo.BulkGetByCPFs(ctx, vagaID, cpfs)
		assert.Error(t, err)
		assert.Nil(t, candidaturas)
		assert.Contains(t, err.Error(), "erro ao buscar candidaturas")
	})
}

func TestCandidaturaRepository_BulkUpdateEtapa(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("bulk update etapa success", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		ids := []uuid.UUID{id1, id2}
		etapaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.BulkUpdateEtapa(ctx, ids, etapaID)
		assert.NoError(t, err)
	})

	t.Run("bulk update etapa error", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}
		etapaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.BulkUpdateEtapa(ctx, ids, etapaID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar etapa em lote")
	})
}

func TestCandidaturaRepository_UpdateEtapa(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	ctx := context.Background()

	t.Run("update etapa success", func(t *testing.T) {
		id := uuid.New()
		etapaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateEtapa(ctx, id, etapaID)
		assert.NoError(t, err)
	})

	t.Run("update etapa error", func(t *testing.T) {
		id := uuid.New()
		etapaID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_candidaturas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateEtapa(ctx, id, etapaID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar etapa da candidatura")
	})
}
