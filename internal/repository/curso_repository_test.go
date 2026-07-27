package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCursoRepository_applyFilters(t *testing.T) {
	db, _, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	tests := []struct {
		name               string
		filter             map[string]interface{}
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             map[string]interface{}{},
			expectedConditions: []string{`SELECT * FROM "cursos"`},
			description:        "Empty filter should generate base SELECT without WHERE",
		},
		{
			name: "status NOT filter",
			filter: map[string]interface{}{
				"status NOT": "archived",
			},
			expectedConditions: []string{"status !=", "archived"},
			description:        "status NOT filter should generate != condition",
		},
		{
			name: "title ILIKE filter",
			filter: map[string]interface{}{
				"title ILIKE": "%golang%",
			},
			expectedConditions: []string{"titulo ILIKE", "%golang%"},
			description:        "title ILIKE filter should generate ILIKE condition",
		},
		{
			name: "is_visible NOT_FALSE filter",
			filter: map[string]interface{}{
				"is_visible NOT_FALSE": true,
			},
			expectedConditions: []string{"is_visible IS DISTINCT FROM false"},
			description:        "is_visible NOT_FALSE filter should exclude only explicit false",
		},
		{
			name: "categoria_id filter",
			filter: map[string]interface{}{
				"categoria_id": 5,
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM cursos_categorias WHERE categoria_id"},
			description:        "categoria_id filter should generate subquery",
		},
		{
			name: "acessibilidade_id filter",
			filter: map[string]interface{}{
				"acessibilidade_id": 3,
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM cursos_acessibilidades WHERE acessibilidade_id"},
			description:        "acessibilidade_id filter should generate subquery",
		},
		{
			name: "neighborhood_zone filter",
			filter: map[string]interface{}{
				"neighborhood_zone": "Zona Sul",
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM location_classes WHERE neighborhood_zone"},
			description:        "neighborhood_zone filter should generate subquery",
		},
		{
			name: "default exact match filter",
			filter: map[string]interface{}{
				"instituicao_id": 10,
			},
			expectedConditions: []string{"instituicao_id ="},
			description:        "Unknown filter key should use exact match",
		},
		{
			name: "multiple filters",
			filter: map[string]interface{}{
				"status NOT":     "archived",
				"title ILIKE":    "%python%",
				"categoria_id":   2,
				"instituicao_id": 7,
			},
			expectedConditions: []string{
				"status !=",
				"titulo ILIKE",
				"categoria_id",
				"instituicao_id",
			},
			description: "Multiple filters should all be present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with a base query
			query := db.Model(&models.Curso{})

			// Apply filters
			filteredQuery := repo.applyFilters(query, tt.filter)

			// Get the SQL statement
			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			// Check that all expected conditions are present
			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestCursoRepository_applyFilters_AllFilterTypes(t *testing.T) {
	db, _, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	// Test that all special filter keys are handled correctly
	specialFilters := map[string]string{
		"status NOT":           "status !=",
		"is_visible NOT_FALSE": "is_visible IS DISTINCT FROM false",
		"title ILIKE":          "titulo ILIKE",
		"categoria_id":         "curso_id FROM cursos_categorias WHERE categoria_id",
		"acessibilidade_id":    "curso_id FROM cursos_acessibilidades WHERE acessibilidade_id",
		"neighborhood_zone":    "curso_id FROM location_classes WHERE neighborhood_zone",
	}

	for filterKey, expectedSQLFragment := range specialFilters {
		t.Run(filterKey, func(t *testing.T) {
			filter := map[string]interface{}{
				filterKey: "test_value",
			}

			query := db.Model(&models.Curso{})
			filteredQuery := repo.applyFilters(query, filter)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			assert.Contains(t, sql, expectedSQLFragment,
				"Filter %s should generate SQL containing %s", filterKey, expectedSQLFragment)
		})
	}
}

func TestCursoRepository_applyFilters_DefaultBehavior(t *testing.T) {
	db, _, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	// Test that unknown filter keys use default exact match behavior
	unknownFilters := []string{
		"random_field",
		"another_field",
		"custom_column",
	}

	for _, filterKey := range unknownFilters {
		t.Run(filterKey, func(t *testing.T) {
			filter := map[string]interface{}{
				filterKey: "test_value",
			}

			query := db.Model(&models.Curso{})
			filteredQuery := repo.applyFilters(query, filter)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			// Should use exact match with "="
			expectedSQL := filterKey + " ="
			assert.Contains(t, sql, expectedSQL,
				"Unknown filter key %s should use exact match (=)", filterKey)
		})
	}
}

func TestCursoRepository_ValidateForEnrollment_AutoApproveLogic(t *testing.T) {
	// Test the autoApprove logic without DB - it's a pure function
	tests := []struct {
		name                   string
		autoApproveEnrollments *bool
		expectedAutoApprove    bool
		description            string
	}{
		{
			name:                   "nil pointer should return false",
			autoApproveEnrollments: nil,
			expectedAutoApprove:    false,
			description:            "When AutoApproveEnrollments is nil, should return false",
		},
		{
			name: "false pointer should return false",
			autoApproveEnrollments: func() *bool {
				b := false
				return &b
			}(),
			expectedAutoApprove: false,
			description:         "When AutoApproveEnrollments is false, should return false",
		},
		{
			name: "true pointer should return true",
			autoApproveEnrollments: func() *bool {
				b := true
				return &b
			}(),
			expectedAutoApprove: true,
			description:         "When AutoApproveEnrollments is true, should return true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from ValidateForEnrollment
			autoApproveValue := tt.autoApproveEnrollments != nil && *tt.autoApproveEnrollments

			assert.Equal(t, tt.expectedAutoApprove, autoApproveValue, tt.description)
		})
	}
}

func TestCursoRepository_ValidateForEnrollment_NilHandling(t *testing.T) {
	// Test that nil pointer dereference is handled correctly
	t.Run("nil pointer safe check", func(t *testing.T) {
		var autoApproveEnrollments *bool = nil

		// This is the exact logic from the repository
		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.False(t, autoApproveValue,
			"Nil pointer should be handled safely and return false")
	})

	t.Run("non-nil false pointer", func(t *testing.T) {
		b := false
		autoApproveEnrollments := &b

		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.False(t, autoApproveValue,
			"False boolean pointer should return false")
	})

	t.Run("non-nil true pointer", func(t *testing.T) {
		b := true
		autoApproveEnrollments := &b

		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.True(t, autoApproveValue,
			"True boolean pointer should return true")
	})
}

func TestCursoRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		instituicaoID := 1
		curso := &models.Curso{
			Titulo:        "Curso de Golang",
			Status:        "active",
			InstituicaoID: &instituicaoID,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, curso)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		curso := &models.Curso{
			Titulo: "Curso de Python",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "cursos"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, curso)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "erro ao criar curso")
	})
}

func TestCursoRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		// Create new mock with MatchExpectationsInOrder(false) for this test
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		cursoID := 1
		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(cursoID, "Curso de Golang", "active")

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		// Expect preloads - order doesn't matter with MatchExpectationsInOrder(false)
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "instituicoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "custom_fields"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		curso, err := repo2.GetByID(ctx, cursoID)
		assert.NoError(t, err)
		assert.NotNil(t, curso)
		assert.Equal(t, cursoID, curso.ID)
		assert.Equal(t, "Curso de Golang", curso.Titulo)
	})

	t.Run("get by id not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		curso, err := repo.GetByID(ctx, 999)
		assert.NoError(t, err)
		assert.Nil(t, curso)
	})

	t.Run("get by id database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(assert.AnError)

		curso, err := repo.GetByID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, curso)
		assert.Contains(t, err.Error(), "erro ao buscar curso por ID")
	})

	t.Run("get by id with preload failures", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		cursoID := 1
		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(cursoID, "Curso Test", "active")

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		// Simulate preload failures for each association
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnError(assert.AnError)
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "instituicoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "custom_fields"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		curso, err := repo2.GetByID(ctx, cursoID)
		// The error from preload should propagate
		assert.Error(t, err)
		assert.Nil(t, curso)
	})

	t.Run("get by id with empty preloads", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		cursoID := 1
		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(cursoID, "Curso Test", "active")

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		// All preloads return empty results
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "instituicoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "custom_fields"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		curso, err := repo2.GetByID(ctx, cursoID)
		assert.NoError(t, err)
		assert.NotNil(t, curso)
		assert.Equal(t, cursoID, curso.ID)
	})
}

func TestCursoRepository_Update(t *testing.T) {
	// Note: Update() is a complex transaction with many nested operations.
	// Testing with sqlmock would require extensive expectations for all GORM preloads,
	// associations, and sub-operations. These are better covered by integration tests.
	// Here we test that the transaction framework is called correctly.

	t.Run("update requires transaction", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID:     1,
			Titulo: "Curso Test",
		}

		// Expect transaction begin
		mock.ExpectBegin()
		// First operation in transaction will fail
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(assert.AnError)
		// Expect rollback
		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		// Verify transaction was used
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCursoRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "cursos"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, 1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "cursos"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir curso")
	})
}

func TestCursoRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("list success with filters", func(t *testing.T) {
		// Create new mock with MatchExpectationsInOrder(false)
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		filter := map[string]interface{}{
			"status": "active",
		}

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(1, "Curso 1", "active").
			AddRow(2, "Curso 2", "active")
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		// Expect preloads - order doesn't matter
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		cursos, total, err := repo2.List(ctx, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, cursos, 2)
	})

	t.Run("list without pagination when limit <= 0 returns all rows", func(t *testing.T) {
		// A non-positive limit is used by availability sorting to fetch the full
		// result set (no SQL LIMIT/OFFSET) before ordering and slicing in memory.
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(3, "Curso 3", "published").
			AddRow(2, "Curso 2", "published").
			AddRow(1, "Curso 1", "published")
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		cursos, total, err := repo2.List(ctx, nil, -1, 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, cursos, 3)
		assert.NoError(t, mock2.ExpectationsWereMet())
	})

	t.Run("list count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnError(assert.AnError)

		cursos, total, err := repo.List(ctx, nil, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, cursos)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao contar cursos")
	})

	t.Run("list query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(assert.AnError)

		cursos, total, err := repo.List(ctx, nil, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, cursos)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar cursos")
	})

	t.Run("list with complex filters", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		filter := map[string]interface{}{
			"status NOT":        "archived",
			"title ILIKE":       "%golang%",
			"categoria_id":      1,
			"acessibilidade_id": 2,
			"neighborhood_zone": "Zona Sul",
			"instituicao_id":    3,
		}

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(1, "Curso Golang", "active")
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		cursos, total, err := repo2.List(ctx, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, cursos, 1)
	})

	t.Run("list with pagination boundary", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"})
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Request offset beyond total
		cursos, total, err := repo2.List(ctx, nil, 10, 150)
		assert.NoError(t, err)
		assert.Equal(t, 100, total)
		assert.Len(t, cursos, 0)
	})

	t.Run("list with empty result", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"})
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "categoria_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_acessibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"curso_id", "acessibilidade_id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		cursos, total, err := repo2.List(ctx, nil, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Len(t, cursos, 0)
	})

	t.Run("list preload error", func(t *testing.T) {
		db2, mock2, cleanup2 := SetupMockDB(t)
		defer cleanup2()
		repo2 := NewCursoRepository(db2)
		mock2.MatchExpectationsInOrder(false)

		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(1, "Curso Test", "active")
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(rows)

		// Simulate preload error
		mock2.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos_categorias"`)).
			WillReturnError(assert.AnError)

		cursos, total, err := repo2.List(ctx, nil, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, cursos)
		assert.Equal(t, 0, total)
	})
}

func TestCursoRepository_CreateCustomFields(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("create custom fields success", func(t *testing.T) {
		customFields := []models.CustomField{
			{CursoID: 1, Title: "Pergunta 1"},
			{CursoID: 1, Title: "Pergunta 2"},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "custom_fields"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.CreateCustomFields(ctx, customFields)
		assert.NoError(t, err)
		// display_order must follow the array position so reads return a stable order
		assert.Equal(t, 1, customFields[0].DisplayOrder)
		assert.Equal(t, 2, customFields[1].DisplayOrder)
	})

	t.Run("create custom fields empty", func(t *testing.T) {
		err := repo.CreateCustomFields(ctx, []models.CustomField{})
		assert.NoError(t, err)
	})

	t.Run("create custom fields error", func(t *testing.T) {
		customFields := []models.CustomField{
			{CursoID: 1, Title: "Pergunta 1"},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "custom_fields"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.CreateCustomFields(ctx, customFields)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar custom fields")
	})
}

func TestCursoRepository_CreateRemoteClass(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("create remote class success", func(t *testing.T) {
		remoteClass := &models.RemoteClass{
			CursoID: 1,
			Schedules: []models.RemoteSchedule{
				{Vacancies: 10},
			},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		// Expect schedule insert
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.CreateRemoteClass(ctx, remoteClass)
		assert.NoError(t, err)
		assert.Equal(t, 1, remoteClass.Schedules[0].DisplayOrder)
	})

	t.Run("create remote class nil", func(t *testing.T) {
		err := repo.CreateRemoteClass(ctx, nil)
		assert.NoError(t, err)
	})

	t.Run("create remote class error", func(t *testing.T) {
		remoteClass := &models.RemoteClass{
			CursoID: 1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_classes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.CreateRemoteClass(ctx, remoteClass)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar remote class")
	})
}

func TestCursoRepository_CreateLocationClasses(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("create location classes success", func(t *testing.T) {
		locationClasses := []models.LocationClass{
			{
				CursoID: 1,
				Address: "Rua A",
				Schedules: []models.CourseSchedule{
					{Vacancies: 20},
					{Vacancies: 15},
				},
			},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		// Expect schedule insert with 2 schedules
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.CreateLocationClasses(ctx, locationClasses)
		assert.NoError(t, err)
		assert.Equal(t, 1, locationClasses[0].Schedules[0].DisplayOrder)
		assert.Equal(t, 2, locationClasses[0].Schedules[1].DisplayOrder)
	})

	t.Run("create location classes empty", func(t *testing.T) {
		err := repo.CreateLocationClasses(ctx, []models.LocationClass{})
		assert.NoError(t, err)
	})

	t.Run("create location classes error", func(t *testing.T) {
		locationClasses := []models.LocationClass{
			{CursoID: 1},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location_classes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.CreateLocationClasses(ctx, locationClasses)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar location classes")
	})
}

func TestCursoRepository_CountEnrollmentsByScheduleID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("count success", func(t *testing.T) {
		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		count, err := repo.CountEnrollmentsByScheduleID(ctx, scheduleID)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("count error", func(t *testing.T) {
		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnError(assert.AnError)

		count, err := repo.CountEnrollmentsByScheduleID(ctx, scheduleID)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.Contains(t, err.Error(), "erro ao contar inscrições por schedule")
	})
}

func TestCursoRepository_CountEnrollmentsByScheduleIDs(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("count multiple schedules success", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		scheduleIDs := []uuid.UUID{id1, id2}

		rows := sqlmock.NewRows([]string{"schedule_id", "count"}).
			AddRow(id1, 3).
			AddRow(id2, 7)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT schedule_id, COUNT(*) as count FROM "inscricoes"`)).
			WillReturnRows(rows)

		countMap, err := repo.CountEnrollmentsByScheduleIDs(ctx, scheduleIDs)
		assert.NoError(t, err)
		assert.Len(t, countMap, 2)
		assert.Equal(t, int64(3), countMap[id1])
		assert.Equal(t, int64(7), countMap[id2])
	})

	t.Run("count empty schedule list", func(t *testing.T) {
		countMap, err := repo.CountEnrollmentsByScheduleIDs(ctx, []uuid.UUID{})
		assert.NoError(t, err)
		assert.Empty(t, countMap)
	})

	t.Run("count error", func(t *testing.T) {
		scheduleIDs := []uuid.UUID{uuid.New()}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT schedule_id, COUNT(*) as count FROM "inscricoes"`)).
			WillReturnError(assert.AnError)

		countMap, err := repo.CountEnrollmentsByScheduleIDs(ctx, scheduleIDs)
		assert.Error(t, err)
		assert.Nil(t, countMap)
		assert.Contains(t, err.Error(), "erro ao contar inscrições por schedules")
	})
}

func TestCursoRepository_ValidateForEnrollment(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("validate success with auto approve", func(t *testing.T) {
		autoApprove := true
		rows := sqlmock.NewRows([]string{"status", "enrollment_start_date", "enrollment_end_date", "auto_approve_enrollments"}).
			AddRow("active", nil, nil, &autoApprove)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT status, enrollment_start_date, enrollment_end_date, auto_approve_enrollments FROM "cursos"`)).
			WillReturnRows(rows)

		status, startDate, endDate, autoApproveVal, err := repo.ValidateForEnrollment(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, "active", status)
		assert.Nil(t, startDate)
		assert.Nil(t, endDate)
		assert.True(t, autoApproveVal)
	})

	t.Run("validate not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT status, enrollment_start_date, enrollment_end_date, auto_approve_enrollments FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"status"}))

		status, startDate, endDate, autoApproveVal, err := repo.ValidateForEnrollment(ctx, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "curso não encontrado")
		assert.Empty(t, status)
		assert.Nil(t, startDate)
		assert.Nil(t, endDate)
		assert.False(t, autoApproveVal)
	})

	t.Run("validate database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT status, enrollment_start_date, enrollment_end_date, auto_approve_enrollments FROM "cursos"`)).
			WillReturnError(assert.AnError)

		status, startDate, endDate, autoApproveVal, err := repo.ValidateForEnrollment(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao validar curso")
		assert.Empty(t, status)
		assert.Nil(t, startDate)
		assert.Nil(t, endDate)
		assert.False(t, autoApproveVal)
	})
}

func TestCursoRepository_GetCourseScheduleByID(t *testing.T) {
	t.Run("get schedule found", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		mock.MatchExpectationsInOrder(false)
		ctx := context.Background()

		scheduleID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "vacancies"}).
			AddRow(scheduleID, 10)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		schedule, err := repo.GetCourseScheduleByID(ctx, scheduleID)
		assert.NoError(t, err)
		assert.NotNil(t, schedule)
		assert.Equal(t, scheduleID, schedule.ID)
	})

	t.Run("get schedule not found", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		schedule, err := repo.GetCourseScheduleByID(ctx, scheduleID)
		assert.NoError(t, err)
		assert.Nil(t, schedule)
	})

	t.Run("get schedule database error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnError(assert.AnError)

		schedule, err := repo.GetCourseScheduleByID(ctx, scheduleID)
		assert.Error(t, err)
		assert.Nil(t, schedule)
		assert.Contains(t, err.Error(), "erro ao buscar course schedule")
	})
}

func TestCursoRepository_GetRemoteScheduleByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)
	ctx := context.Background()

	t.Run("get remote schedule found", func(t *testing.T) {
		scheduleID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "vacancies"}).
			AddRow(scheduleID, 15)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(rows)

		schedule, err := repo.GetRemoteScheduleByID(ctx, scheduleID)
		assert.NoError(t, err)
		assert.NotNil(t, schedule)
		assert.Equal(t, scheduleID, schedule.ID)
	})

	t.Run("get remote schedule not found", func(t *testing.T) {
		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		schedule, err := repo.GetRemoteScheduleByID(ctx, scheduleID)
		assert.NoError(t, err)
		assert.Nil(t, schedule)
	})

	t.Run("get remote schedule database error", func(t *testing.T) {
		scheduleID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnError(assert.AnError)

		schedule, err := repo.GetRemoteScheduleByID(ctx, scheduleID)
		assert.Error(t, err)
		assert.Nil(t, schedule)
		assert.Contains(t, err.Error(), "erro ao buscar remote schedule")
	})
}

// Simplified integration-style tests for Update and its helper methods
// These tests cover the complex transaction logic without overly brittle mocking

func TestCursoRepository_Update_HelperIntegration(t *testing.T) {
	t.Run("update customFields delete error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID: 1,
			CustomFields: []models.CustomField{
				{Title: "Q1"},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// updateCustomFieldsWithTx - delete fails
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar custom fields")
	})

	t.Run("update customFields insert error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID: 1,
			CustomFields: []models.CustomField{
				{Title: "Q1"},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// updateCustomFieldsWithTx - delete succeeds, insert fails
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "custom_fields"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar custom fields")
	})

	t.Run("update remoteClass create new", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID: 1,
			RemoteClass: &models.RemoteClass{
				CursoID: 1,
				Schedules: []models.RemoteSchedule{
					{Vacancies: 10},
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateRemoteClassWithTx - no existing, create new
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		// Pass locationClasses
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Final UPDATE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update remoteClass delete when nil", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID:          1,
			RemoteClass: nil,
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateRemoteClassWithTx - nil means delete existing
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Pass locationClasses
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Final UPDATE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update remoteClass update existing with schedule update", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		existingRemoteID := uuid.New()
		existingScheduleID := uuid.New()
		curso := &models.Curso{
			ID: 1,
			RemoteClass: &models.RemoteClass{
				ID:      existingRemoteID,
				CursoID: 1,
				Schedules: []models.RemoteSchedule{
					{ID: existingScheduleID, Vacancies: 20},
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateRemoteClassWithTx - update existing
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(existingRemoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(existingScheduleID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_schedules"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Pass locationClasses
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Final UPDATE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update remoteClass update schedule error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		existingScheduleID := uuid.New()
		curso := &models.Curso{
			ID: 1,
			RemoteClass: &models.RemoteClass{
				CursoID: 1,
				Schedules: []models.RemoteSchedule{
					{ID: existingScheduleID, Vacancies: 20},
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateRemoteClassWithTx - schedule update fails
		existingRemoteID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(existingRemoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(existingScheduleID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_schedules"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar remote class")
	})

	t.Run("update remoteClass batch create schedules error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		curso := &models.Curso{
			ID: 1,
			RemoteClass: &models.RemoteClass{
				CursoID: 1,
				Schedules: []models.RemoteSchedule{
					{Vacancies: 10}, // New schedule
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateRemoteClassWithTx - batch create fails
		existingRemoteID := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(existingRemoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_schedules"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar remote class")
	})

	t.Run("update locationClasses create new", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()
		mock.MatchExpectationsInOrder(false)

		curso := &models.Curso{
			ID: 1,
			LocationClasses: []models.LocationClass{
				{
					CursoID: 1,
					Address: "Rua A",
					Schedules: []models.CourseSchedule{
						{Vacancies: 15},
					},
				},
			},
		}

		mock.ExpectBegin()

		// Various helpers will query - unordered
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		// RemoteClass nil => delete existing remote class if any
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateLocationClassesWithTx - create new
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		// Final UPDATE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update locationClasses update existing with schedule update", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()
		mock.MatchExpectationsInOrder(false)

		locationID := uuid.New()
		scheduleID := uuid.New()
		curso := &models.Curso{
			ID: 1,
			LocationClasses: []models.LocationClass{
				{
					ID:      locationID,
					CursoID: 1,
					Address: "Rua B",
					Schedules: []models.CourseSchedule{
						{ID: scheduleID, Vacancies: 25},
					},
				},
			},
		}

		mock.ExpectBegin()

		// Various helpers - unordered
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		// RemoteClass nil => delete existing remote class if any
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateLocationClassesWithTx - update existing
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(locationID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "location_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(scheduleID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "course_schedules"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Final UPDATE
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update locationClasses schedule update error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		locationID := uuid.New()
		scheduleID := uuid.New()
		curso := &models.Curso{
			ID: 1,
			LocationClasses: []models.LocationClass{
				{
					ID:      locationID,
					CursoID: 1,
					Schedules: []models.CourseSchedule{
						{ID: scheduleID, Vacancies: 25},
					},
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// RemoteClass nil => delete existing remote class if any
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateLocationClassesWithTx - schedule update fails
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(locationID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "location_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(scheduleID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "course_schedules"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar location classes")
	})

	t.Run("update locationClasses batch create schedules error", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		locationID := uuid.New()
		curso := &models.Curso{
			ID: 1,
			LocationClasses: []models.LocationClass{
				{
					ID:      locationID,
					CursoID: 1,
					Schedules: []models.CourseSchedule{
						{Vacancies: 10}, // New schedule
					},
				},
			},
		}

		mock.ExpectBegin()

		// Pass categorias
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass acessibilidades
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Pass customFields
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// RemoteClass nil => delete existing remote class if any
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// updateLocationClassesWithTx - batch create fails
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(locationID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "location_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "course_schedules"`)).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := repo.Update(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar location classes")
	})

	t.Run("update remoteClass batch deletes obsolete schedules", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		existingRemoteID := uuid.New()
		keepScheduleID := uuid.New()
		obsoleteScheduleID1 := uuid.New()
		obsoleteScheduleID2 := uuid.New()
		curso := &models.Curso{
			ID: 1,
			RemoteClass: &models.RemoteClass{
				ID:      existingRemoteID,
				CursoID: 1,
				Schedules: []models.RemoteSchedule{
					{ID: keepScheduleID, Vacancies: 20},
				},
			},
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(existingRemoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(keepScheduleID).
				AddRow(obsoleteScheduleID1).
				AddRow(obsoleteScheduleID2))
		// Single batch DELETE for obsolete schedules (not N individual deletes)
		mock.ExpectExec(`DELETE FROM "remote_schedules" WHERE id IN \((\$\d+,)*\$\d+\)`).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_schedules"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})

	t.Run("update locationClasses batch deletes obsolete locations and schedules", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)
		ctx := context.Background()

		keepLocationID := uuid.New()
		obsoleteLocationID1 := uuid.New()
		obsoleteLocationID2 := uuid.New()
		keepScheduleID := uuid.New()
		obsoleteScheduleID1 := uuid.New()
		obsoleteScheduleID2 := uuid.New()
		curso := &models.Curso{
			ID: 1,
			LocationClasses: []models.LocationClass{
				{
					ID:      keepLocationID,
					CursoID: 1,
					Address: "Rua Keep",
					Schedules: []models.CourseSchedule{
						{ID: keepScheduleID, Vacancies: 25},
					},
				},
			},
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		// RemoteClass nil => delete existing remote class if any
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(keepLocationID).
				AddRow(obsoleteLocationID1).
				AddRow(obsoleteLocationID2))
		// Single batch DELETE for obsolete locations
		mock.ExpectExec(`DELETE FROM "location_classes" WHERE id IN \((\$\d+,)*\$\d+\)`).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "location_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(keepScheduleID).
				AddRow(obsoleteScheduleID1).
				AddRow(obsoleteScheduleID2))
		// Single batch DELETE for obsolete schedules
		mock.ExpectExec(`DELETE FROM "course_schedules" WHERE id IN \((\$\d+,)*\$\d+\)`).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "course_schedules"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.Update(ctx, curso)
		assert.NoError(t, err)
	})
}
