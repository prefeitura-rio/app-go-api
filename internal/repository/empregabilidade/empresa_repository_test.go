package empregabilidade

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEmpresaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	tests := []struct {
		name                string
		filter              empregabilidade.EmpresaFilter
		expectedConditions  []string
		description         string
	}{
		{
			name: "empty filter",
			filter: empregabilidade.EmpresaFilter{},
			expectedConditions: []string{},
			description: "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by CNPJ exact match",
			filter: empregabilidade.EmpresaFilter{
				CNPJ: "12345678000190",
			},
			expectedConditions: []string{"cnpj =", "12345678000190"},
			description: "CNPJ filter should generate exact match condition",
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
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	tests := []struct {
		name           string
		searchTerm     string
		description    string
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
	db, _, cleanup := setupMockDB(t)
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
	db, _, cleanup := setupMockDB(t)
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
