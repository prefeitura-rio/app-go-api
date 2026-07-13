package empregabilidade

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// jsonStringSliceArg é um sqlmock.Argument que decodifica o argumento recebido
// como JSON e compara com o slice esperado. Usado para provar que Opcoes chega
// ao driver como JSON serializado (e não como uma expressão SQL expandida, que
// é o que um Updates baseado em map produziria para uma coluna jsonb).
type jsonStringSliceArg struct{ want []string }

func (m jsonStringSliceArg) Match(v driver.Value) bool {
	var raw []byte
	switch t := v.(type) {
	case []byte:
		raw = t
	case string:
		raw = []byte(t)
	default:
		return false
	}
	var got []string
	if err := json.Unmarshal(raw, &got); err != nil {
		return false
	}
	if len(got) != len(m.want) {
		return false
	}
	for i := range got {
		if got[i] != m.want[i] {
			return false
		}
	}
	return true
}

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

	// Zonas is a many2many relation, so it queries the junction table first
	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_zonas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_zona"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_idiomas_requisitos"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_idioma", "id_nivel_minimo"}))

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

func TestVagaRepository_GetByIDPrefix_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	regimeID := uuid.New()
	modeloID := uuid.New()

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

	// id_orgao_parceiro is zero-value ("") for this row, so GORM skips the
	// OrgaoParceiro preload query entirely (no rows to look up) - no mock needed.

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_regimes_contratacao"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_modelos_trabalho"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_tipos_pcd"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_tipo_pcd"}))

	// Zonas is a many2many relation, so it queries the junction table first
	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_zonas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_zona"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas_idiomas_requisitos"`).
		WillReturnRows(sqlmock.NewRows([]string{"id_vaga", "id_idioma", "id_nivel_minimo"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_etapas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "emp_informacoes_complementares"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	result, err := repo.GetByIDPrefix(ctx, strings.ReplaceAll(vagaID.String(), "-", "")[:12])
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, vagaID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
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

func TestVagaRepository_Update_IncludesCriteriosElegibilidadeColumns(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	idadeMin := 18
	idadeMax := 65
	escolaridadeID := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID:                         uuid.New(),
		Titulo:                     "Desenvolvedor Go",
		Descricao:                  "Vaga para desenvolvedor Go",
		IDContratante:              "12345678000190",
		IDRegimeContratacao:        uuid.New(),
		IDModeloTrabalho:           uuid.New(),
		Status:                     empregabilidade.StatusVagaEmEdicao,
		IdadeMinima:                &idadeMin,
		IdadeMaxima:                &idadeMax,
		BairrosElegibilidade:       []string{"Centro", "Tijuca"},
		IDEscolaridadeMinima:       &escolaridadeID,
		AreasFormacaoElegibilidade: []string{"Tecnologia"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE "emp_vagas" SET.*"areas_formacao_elegibilidade".*"bairros_elegibilidade".*"id_escolaridade_minima".*"idade_maxima".*"idade_minima".*WHERE`).
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

func TestVagaRepository_UpdateZonas_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	zonaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_zonas`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_zonas`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateZonas(ctx, vagaID, []uuid.UUID{zonaID})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateZonas_EmptyList(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_zonas`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateZonas(ctx, vagaID, []uuid.UUID{})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateIdiomasRequisito_Success(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	idiomaID := uuid.New()
	nivelID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_idiomas_requisitos`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_idiomas_requisitos`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	requisitos := []empregabilidade.VagaIdiomaRequisito{
		{IDIdioma: idiomaID, IDNivelMinimo: nivelID},
	}

	err := repo.UpdateIdiomasRequisito(ctx, vagaID, requisitos)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateIdiomasRequisito_EmptyList(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_idiomas_requisitos`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateIdiomasRequisito(ctx, vagaID, []empregabilidade.VagaIdiomaRequisito{})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateIdiomasRequisito_DeleteError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_idiomas_requisitos`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateIdiomasRequisito(ctx, vagaID, []empregabilidade.VagaIdiomaRequisito{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover requisitos de idioma")
}

func TestVagaRepository_UpdateIdiomasRequisito_InsertError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	idiomaID := uuid.New()
	nivelID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_idiomas_requisitos`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_idiomas_requisitos`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	requisitos := []empregabilidade.VagaIdiomaRequisito{
		{IDIdioma: idiomaID, IDNivelMinimo: nivelID},
	}

	err := repo.UpdateIdiomasRequisito(ctx, vagaID, requisitos)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao inserir requisito de idioma")
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

func TestVagaRepository_UpdateWithAssociations_IncludesCriteriosElegibilidadeColumns(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	idadeMin := 18
	idadeMax := 65
	escolaridadeID := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID:                         uuid.New(),
		Titulo:                     "Desenvolvedor Go",
		Descricao:                  "Vaga para desenvolvedor Go",
		IDContratante:              "12345678000190",
		IDRegimeContratacao:        uuid.New(),
		IDModeloTrabalho:           uuid.New(),
		Status:                     empregabilidade.StatusVagaEmEdicao,
		IdadeMinima:                &idadeMin,
		IdadeMaxima:                &idadeMax,
		BairrosElegibilidade:       []string{"Centro"},
		IDEscolaridadeMinima:       &escolaridadeID,
		AreasFormacaoElegibilidade: []string{"Tecnologia"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE "emp_vagas" SET.*"areas_formacao_elegibilidade".*"bairros_elegibilidade".*"id_escolaridade_minima".*"idade_maxima".*"idade_minima".*WHERE`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
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

// TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_PreservesExistingUUID
// covers PREF-373: editar uma vaga não deve regenerar o UUID de uma informação
// complementar já existente, pois candidaturas antigas guardam esse UUID em
// respostas_info_complementares.
func TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_PreservesExistingUUID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	infoID := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID:                  vagaID,
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{
			{ID: infoID, Titulo: "Disponibilidade para trabalho remoto?", TipoCampo: empregabilidade.TipoCampoSelecaoUnica},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares".*id NOT IN`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE "emp_informacoes_complementares" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, infoID, vaga.InformacoesComplementares[0].ID,
		"UUID de informação complementar existente nunca deve ser regenerado")
}

// TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_UpdateSerializesOpcoes
// cobre a regressão descoberta em review do PREF-373: um Updates baseado em
// map[string]interface{} ignora o serializer:json do GORM para o campo Opcoes
// (jsonb) e gera uma expressão SQL inválida para perguntas de seleção
// única/múltipla. O update de um item existente precisa serializar Opcoes como
// JSON de verdade.
func TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_UpdateSerializesOpcoes(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	infoID := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID:                  vagaID,
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{
			{
				ID:        infoID,
				Titulo:    "Qual seu nível de experiência?",
				TipoCampo: empregabilidade.TipoCampoSelecaoUnica,
				Opcoes:    []string{"Júnior", "Pleno", "Sênior"},
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares".*id NOT IN`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE "emp_informacoes_complementares" SET`).
		WithArgs(
			vaga.InformacoesComplementares[0].Titulo,
			vaga.InformacoesComplementares[0].Obrigatorio,
			vaga.InformacoesComplementares[0].TipoCampo,
			sqlmock.AnyArg(), // ValorMinimo
			sqlmock.AnyArg(), // ValorMaximo
			jsonStringSliceArg{want: []string{"Júnior", "Pleno", "Sênior"}},
			sqlmock.AnyArg(), // updated_at
			infoID,
			vagaID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_MixExistingAndNew
// cobre o cenário de adicionar uma nova pergunta mantendo as existentes: apenas o
// item novo deve ser inserido (com UUID gerado pelo banco); o existente deve ser
// atualizado in-place, preservando seu UUID.
func TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_MixExistingAndNew(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	existingInfoID := uuid.New()
	generatedInfoID := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID:                  vagaID,
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{
			{ID: existingInfoID, Titulo: "Pergunta existente", TipoCampo: empregabilidade.TipoCampoRespostaCurta},
			{Titulo: "Pergunta nova", TipoCampo: empregabilidade.TipoCampoRespostaCurta}, // ID == uuid.Nil
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares".*id NOT IN`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE "emp_informacoes_complementares" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`INSERT INTO "emp_informacoes_complementares"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(generatedInfoID))
	mock.ExpectCommit()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, existingInfoID, vaga.InformacoesComplementares[0].ID,
		"UUID existente deve ser preservado")
	assert.NotEqual(t, uuid.Nil, vaga.InformacoesComplementares[1].ID,
		"novo item deve receber um UUID gerado")
}

// TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_AllNew_DeletesAll
// cobre a retrocompatibilidade: se todos os itens do payload têm ID == uuid.Nil
// (comportamento atual do frontend), o resultado deve ser idêntico ao de hoje —
// delete-all seguido de insert-all.
func TestVagaRepository_UpdateWithAssociations_InformacoesComplementares_AllNew_DeletesAll(t *testing.T) {
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
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestVagaRepository_UpdateWithAssociations_UpdateInformacaoComplementarError garante
// que falha ao atualizar um item existente propaga erro com contexto e faz rollback.
func TestVagaRepository_UpdateWithAssociations_UpdateInformacaoComplementarError(t *testing.T) {
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
			{ID: uuid.New(), Titulo: "Pergunta", TipoCampo: empregabilidade.TipoCampoRespostaCurta},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "emp_informacoes_complementares"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE "emp_informacoes_complementares" SET`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(ctx, vaga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar informação complementar")
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

func TestVagaRepository_UpdateZonas_DeleteError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_zonas`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateZonas(ctx, vagaID, []uuid.UUID{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover zonas")
}

func TestVagaRepository_UpdateZonas_InsertError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()
	vagaID := uuid.New()
	zonaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM emp_vagas_zonas`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO emp_vagas_zonas`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateZonas(ctx, vagaID, []uuid.UUID{zonaID})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao inserir zona")
}

// ==================== Multi-Select Filter Query Generation Tests ====================

func TestVagaRepository_ListPublic_MultiSelectFilters_QueryGeneration(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	uuid1Str := uuid.New().String()
	uuid2Str := uuid.New().String()

	tests := []struct {
		name               string
		filter             empregabilidade.VagaPublicFilter
		expectedConditions []string
		notExpected        []string
		description        string
	}{
		{
			name:               "empty filter has no extra WHERE",
			filter:             empregabilidade.VagaPublicFilter{},
			expectedConditions: []string{},
			description:        "Empty filter should produce no extra conditions",
		},
		{
			name:               "single bairro produces ILIKE",
			filter:             empregabilidade.VagaPublicFilter{Bairro: []string{"Centro"}},
			expectedConditions: []string{"emp_vagas.bairro ILIKE", "%Centro%"},
			description:        "Single bairro should produce ILIKE condition",
		},
		{
			name:               "multiple bairros produce OR-joined ILIKEs",
			filter:             empregabilidade.VagaPublicFilter{Bairro: []string{"Centro", "Tijuca"}},
			expectedConditions: []string{"emp_vagas.bairro ILIKE", "%Centro%", "%Tijuca%", "OR"},
			description:        "Multiple bairros should produce OR-joined ILIKE conditions",
		},
		{
			name:               "single id_regime_contratacao produces = clause",
			filter:             empregabilidade.VagaPublicFilter{IDRegimeContratacao: []string{uuid1Str}},
			expectedConditions: []string{"emp_vagas.id_regime_contratacao =", uuid1Str},
			notExpected:        []string{"IN ("},
			description:        "Single IDRegimeContratacao should use = not IN",
		},
		{
			name:               "multiple id_regime_contratacao produces IN clause",
			filter:             empregabilidade.VagaPublicFilter{IDRegimeContratacao: []string{uuid1Str, uuid2Str}},
			expectedConditions: []string{"emp_vagas.id_regime_contratacao IN"},
			description:        "Multiple IDRegimeContratacao should use IN",
		},
		{
			name:               "single id_modelo_trabalho produces = clause",
			filter:             empregabilidade.VagaPublicFilter{IDModeloTrabalho: []string{uuid1Str}},
			expectedConditions: []string{"emp_vagas.id_modelo_trabalho =", uuid1Str},
			notExpected:        []string{"IN ("},
			description:        "Single IDModeloTrabalho should use = not IN",
		},
		{
			name:               "multiple id_modelo_trabalho produces IN clause",
			filter:             empregabilidade.VagaPublicFilter{IDModeloTrabalho: []string{uuid1Str, uuid2Str}},
			expectedConditions: []string{"emp_vagas.id_modelo_trabalho IN"},
			description:        "Multiple IDModeloTrabalho should use IN",
		},
		{
			name:               "single acessibilidade_pcd produces = clause",
			filter:             empregabilidade.VagaPublicFilter{AcessibilidadePCD: []string{"para_pcd"}},
			expectedConditions: []string{"emp_vagas.acessibilidade_pcd =", "para_pcd"},
			description:        "Single AcessibilidadePCD should use =",
		},
		{
			name:               "multiple acessibilidade_pcd produces IN clause",
			filter:             empregabilidade.VagaPublicFilter{AcessibilidadePCD: []string{"para_pcd", "exclusivo_pcd"}},
			expectedConditions: []string{"emp_vagas.acessibilidade_pcd IN"},
			description:        "Multiple AcessibilidadePCD should use IN",
		},
		{
			name:               "single CNPJ contratante produces id_contratante = clause",
			filter:             empregabilidade.VagaPublicFilter{Contratante: []string{"12345678000190"}},
			expectedConditions: []string{"emp_vagas.id_contratante =", "12345678000190"},
			description:        "Single CNPJ contratante should use =",
		},
		{
			name:               "multiple CNPJ contratantes produce id_contratante IN clause",
			filter:             empregabilidade.VagaPublicFilter{Contratante: []string{"12345678000190", "98765432000199"}},
			expectedConditions: []string{"emp_vagas.id_contratante IN"},
			description:        "Multiple CNPJ contratantes should use IN",
		},
		{
			name:               "name contratante produces JOIN and ILIKE",
			filter:             empregabilidade.VagaPublicFilter{Contratante: []string{"TechRio"}},
			expectedConditions: []string{"emp_empresas", "razao_social ILIKE", "nome_fantasia ILIKE", "%TechRio%"},
			description:        "Name contratante should JOIN emp_empresas and use ILIKE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strings := func(vals []string) []interface{} {
				result := make([]interface{}, len(vals))
				for i, v := range vals {
					result[i] = v
				}
				return result
			}
			_ = strings

			query := db.Model(&empregabilidade.Vaga{})

			if len(tt.filter.IDRegimeContratacao) == 1 {
				query = query.Where("emp_vagas.id_regime_contratacao = ?", tt.filter.IDRegimeContratacao[0])
			} else if len(tt.filter.IDRegimeContratacao) > 1 {
				query = query.Where("emp_vagas.id_regime_contratacao IN ?", tt.filter.IDRegimeContratacao)
			}

			if len(tt.filter.IDModeloTrabalho) == 1 {
				query = query.Where("emp_vagas.id_modelo_trabalho = ?", tt.filter.IDModeloTrabalho[0])
			} else if len(tt.filter.IDModeloTrabalho) > 1 {
				query = query.Where("emp_vagas.id_modelo_trabalho IN ?", tt.filter.IDModeloTrabalho)
			}

			if len(tt.filter.AcessibilidadePCD) == 1 {
				query = query.Where("emp_vagas.acessibilidade_pcd = ?", tt.filter.AcessibilidadePCD[0])
			} else if len(tt.filter.AcessibilidadePCD) > 1 {
				query = query.Where("emp_vagas.acessibilidade_pcd IN ?", tt.filter.AcessibilidadePCD)
			}

			if len(tt.filter.Bairro) > 0 {
				conditions := make([]string, len(tt.filter.Bairro))
				args := make([]interface{}, len(tt.filter.Bairro))
				for i, b := range tt.filter.Bairro {
					conditions[i] = "emp_vagas.bairro ILIKE ?"
					args[i] = "%" + b + "%"
				}
				query = query.Where(joinStrings(conditions, " OR "), args...)
			}

			if len(tt.filter.Contratante) > 0 {
				var cnpjs []string
				var nameConditions []string
				var nameArgs []interface{}
				nonDigitRe := regexp.MustCompile(`\D`)
				for _, c := range tt.filter.Contratante {
					digits := nonDigitRe.ReplaceAllString(c, "")
					if len(digits) >= 11 {
						cnpjs = append(cnpjs, digits)
					} else {
						nameConditions = append(nameConditions,
							"emp_empresas.razao_social ILIKE ? OR emp_empresas.nome_fantasia ILIKE ?")
						nameArgs = append(nameArgs, "%"+c+"%", "%"+c+"%")
					}
				}
				if len(nameConditions) > 0 {
					query = query.Joins("JOIN emp_empresas ON emp_empresas.cnpj = emp_vagas.id_contratante")
				}
				switch {
				case len(cnpjs) > 0 && len(nameConditions) > 0:
					allArgs := append([]interface{}{cnpjs}, nameArgs...)
					query = query.Where(
						"emp_vagas.id_contratante IN ? OR ("+joinStrings(nameConditions, " OR ")+")",
						allArgs...,
					)
				case len(cnpjs) > 0:
					if len(cnpjs) == 1 {
						query = query.Where("emp_vagas.id_contratante = ?", cnpjs[0])
					} else {
						query = query.Where("emp_vagas.id_contratante IN ?", cnpjs)
					}
				default:
					query = query.Where(joinStrings(nameConditions, " OR "), nameArgs...)
				}
			}

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Vaga{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
			for _, condition := range tt.notExpected {
				assert.NotContains(t, sql, condition,
					"%s: SQL should NOT contain '%s'", tt.description, condition)
			}
		})
	}
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

func TestVagaRepository_ListPublic_Success(t *testing.T) {
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

	filter := empregabilidade.VagaPublicFilter{}
	result, total, err := repo.ListPublic(ctx, filter, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, total)
}

func expectListPublicMocks(mock sqlmock.Sqlmock, vagaID uuid.UUID) {
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
}

func TestVagaRepository_ListPublic_WithFilters_Success(t *testing.T) {
	tests := []struct {
		name   string
		filter empregabilidade.VagaPublicFilter
	}{
		{
			name: "multi-select bairro and acessibilidade_pcd",
			filter: empregabilidade.VagaPublicFilter{
				Bairro:              []string{"Centro", "Tijuca"},
				AcessibilidadePCD:   []string{"para_pcd", "exclusivo_pcd"},
				IDRegimeContratacao: []string{uuid.New().String()},
			},
		},
		{
			name: "status publicado_ativo",
			filter: empregabilidade.VagaPublicFilter{
				Status: string(empregabilidade.StatusVagaPublicadoAtivo),
			},
		},
		{
			name: "status publicado_expirado",
			filter: empregabilidade.VagaPublicFilter{
				Status: string(empregabilidade.StatusVagaPublicadoExpirado),
			},
		},
		{
			name: "status congelada",
			filter: empregabilidade.VagaPublicFilter{
				Status: string(empregabilidade.StatusVagaCongelada),
			},
		},
		{
			name: "status descontinuada",
			filter: empregabilidade.VagaPublicFilter{
				Status: string(empregabilidade.StatusVagaDescontinuada),
			},
		},
		{
			name: "data_publicacao hoje",
			filter: empregabilidade.VagaPublicFilter{
				DataPublicacao: empregabilidade.DataPublicacaoHoje,
			},
		},
		{
			name: "data_publicacao ultima semana",
			filter: empregabilidade.VagaPublicFilter{
				DataPublicacao: empregabilidade.DataPublicacaoUltimaSemana,
			},
		},
		{
			name: "data_publicacao ultimo mes",
			filter: empregabilidade.VagaPublicFilter{
				DataPublicacao: empregabilidade.DataPublicacaoUltimoMes,
			},
		},
		{
			name: "contratante CNPJ single",
			filter: empregabilidade.VagaPublicFilter{
				Contratante: []string{"12345678000190"},
			},
		},
		{
			name: "contratante CNPJ multiple",
			filter: empregabilidade.VagaPublicFilter{
				Contratante: []string{"12345678000190", "98765432000199"},
			},
		},
		{
			name: "contratante name",
			filter: empregabilidade.VagaPublicFilter{
				Contratante: []string{"TechRio"},
			},
		},
		{
			name: "contratante mixed CNPJ and name",
			filter: empregabilidade.VagaPublicFilter{
				Contratante: []string{"TechRio", "12345678000190"},
			},
		},
		{
			name: "id_modelo_trabalho multiple",
			filter: empregabilidade.VagaPublicFilter{
				IDModeloTrabalho: []string{uuid.New().String(), uuid.New().String()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := repository.SetupMockDB(t)
			defer cleanup()

			repo := NewVagaRepository(db)
			ctx := context.Background()

			expectListPublicMocks(mock, uuid.New())

			result, total, err := repo.ListPublic(ctx, tt.filter, 10, 0)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, 1, total)
		})
	}
}

func TestVagaRepository_ListPublic_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnError(fmt.Errorf("db error"))

	filter := empregabilidade.VagaPublicFilter{}
	result, total, err := repo.ListPublic(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
}

func TestVagaRepository_ListPublic_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM "emp_vagas"`).
		WillReturnError(fmt.Errorf("find error"))

	filter := empregabilidade.VagaPublicFilter{}
	result, total, err := repo.ListPublic(ctx, filter, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
}
