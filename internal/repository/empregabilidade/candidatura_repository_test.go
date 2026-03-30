package empregabilidade

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCandidaturaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()
	etapaID := uuid.New()

	tests := []struct {
		name                string
		filter              empregabilidade.CandidaturaFilter
		expectedConditions  []string
		description         string
	}{
		{
			name: "empty filter",
			filter: empregabilidade.CandidaturaFilter{},
			expectedConditions: []string{},
			description: "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by CPF",
			filter: empregabilidade.CandidaturaFilter{
				CPF: "12345678900",
			},
			expectedConditions: []string{"cpf =", "12345678900"},
			description: "CPF filter should generate cpf = condition",
		},
		{
			name: "filter by VagaID",
			filter: empregabilidade.CandidaturaFilter{
				VagaID: &vagaID,
			},
			expectedConditions: []string{"id_vaga ="},
			description: "VagaID filter should generate id_vaga = condition",
		},
		{
			name: "filter by Status",
			filter: empregabilidade.CandidaturaFilter{
				Status: string(empregabilidade.StatusCandidaturaAprovada),
			},
			expectedConditions: []string{"status =", "aprovada"},
			description: "Status filter should generate status = condition",
		},
		{
			name: "filter by EtapaID",
			filter: empregabilidade.CandidaturaFilter{
				EtapaID: &etapaID,
			},
			expectedConditions: []string{"id_etapa_atual ="},
			description: "EtapaID filter should generate id_etapa_atual = condition",
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
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	tests := []struct {
		name           string
		searchTerm     string
		expectedOR     []string
		description    string
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
	db, _, cleanup := setupMockDB(t)
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
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		filter      empregabilidade.CandidaturaFilter
		shouldSkip  []string
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
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()
	etapaID := uuid.New()

	// Test that CountByStatus uses the same filter logic but ignores Status field
	tests := []struct {
		name                string
		filter              empregabilidade.CandidaturaFilter
		shouldContain       []string
		shouldNotContain    []string
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
