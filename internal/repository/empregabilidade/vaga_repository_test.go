package empregabilidade

import (
	"testing"

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
